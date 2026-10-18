# lint-rules

A [Please](https://please.build) plugin that runs linters as ordinary plz
tests, so `plz test //...` fails on unformatted Go, `go vet` findings and
pyflakes findings. It also fails on any source file that no linter covers.
Any other linter is added as a [custom rule](#custom-rules), and the
[built-in checks](#built-in-checks) cover doc comments, markdown links, banned
terms, permission-bypass spellings, external test packages and imports.

- **`lint()`** goes in each BUILD file and checks that package's own files. It
  is hermetic and cached like any other test.
- **`repo_lint()`** goes once in the root BUILD file. It runs `go vet` over the
  module and the coverage guard.
- **`repolint()`** goes once in any BUILD file. It builds the built-in checks.

Every tool is supplied by the consuming repo, so the plugin pins nothing and
downloads nothing: the built-in checks are Go source that the consuming repo
compiles with its own go.

## Install

In `plugins/BUILD`:

```python
plugin_repo(
    name = "lint",
    owner = "nv3to",
    plugin = "lint-rules",
    revision = "v0.2.0",
)
```

A copy vendored at `plugins/lint-rules` loads the same way, with
`subrepo(name = "lint", path = "plugins/lint-rules", plugin = True)` in place
of `plugin_repo()`. Add `Exclude = plugins/lint-rules` to `[Plugin "lint"]`, so
that the coverage guard skips the copy's files.

In `.plzconfig`:

```ini
[Parse]
PreloadSubincludes = ///lint//build_defs:lint

[Plugin "lint"]
Target = //plugins:lint
GofmtTool = //third_party/go:toolchain|gofmt
GoTool = //third_party/go:toolchain|go
PyflakesTool = //tools:pyflakes   ; optional: a python_binary running pyflakes.api.main
```

## Use

Add this at the end of every BUILD file that owns `.go` or `.py` files:

```python
lint()                                  # this directory's *.go and *.py
lint(srcs = glob(["**/*.py"]))          # a package that owns sub-directories
```

Add this once, in the root BUILD file:

```python
repo_lint(
    name = "vet",
    packages = ["//cmd/app", "//internal/store"],   # every package that calls lint()
)
```

| Target | Fails when |
|---|---|
| `:lint_gofmt` | `gofmt -l` names a file (the diff is printed) |
| `:lint_pyflakes` | pyflakes reports anything |
| `:lint_<name>` | the custom rule `<name>` exits non-zero |
| `//:vet` (repo_lint) | `go vet` reports anything, or a non-ignored file with a linted extension outside `testdata/` is not in any listed package's `lint()` |

All of them carry the `lint` label, so `plz test //... --include lint` runs
lint alone.

## Custom rules

**For the whole repo**, add a `Rule` line to `[Plugin "lint"]`, one per linter:

```ini
; Rule = <name> <extensions> <tool> [args...]
Rule = shellcheck .sh,.bash //third_party/tools:shellcheck -x
Rule = buildifier .bzl /usr/local/bin/buildifier -mode=check
```

Every `lint()` then also creates `lint_<name>` over the files with those
extensions, running `<tool> [args...] <files...>`. It fails when the tool
exits non-zero. The coverage guard covers those extensions too, so every such
file must be in some package's `lint()`.

**For one package**, call `lint_rule()` in its BUILD file:

```python
lint_rule(
    name = "schema",                    # creates :lint_schema
    srcs = glob(["*.json"]),
    tool = "//tools:check_schema",      # build label or path
    args = ["--strict"],
    fail_on_output = True,              # also fail when the tool prints anything
)
```

Custom rules carry the `lint` label and their own name as a label, so
`plz test //... --include shellcheck` runs one alone.

## Built-in checks

`repolint` is a program of checks that need more than "run a tool over the
files". Its source is in [tools/repolint](tools/repolint) and imports only the
Go standard library. The consuming repo compiles it with its own `GoTool`
(Go 1.21 or later), so nothing is downloaded or pinned.

Build it once, in any BUILD file. In the root BUILD file the target is
`//:repolint`:

```python
repolint(name = "repolint")
```

Then name the built target as the tool of `Rule` lines, or of `lint_rule()` for
one package:

```ini
Rule = godoc .go //:repolint doc
Rule = exttest .go //:repolint exttest
Rule = vocab .go,.md //:repolint vocab --ban 'foo[b]ar' --ban 'old[[:space:]]name'
```

A check runs as `repolint CHECK [FLAG]... FILE...`. It prints one
`file:line: message` per finding and exits 1 when it found any, 2 for bad usage
or a file it cannot read or parse.

| Check | Files | Flags | Fails on |
|---|---|---|---|
| `doc` | `.go` | none | an exported identifier without a doc comment that starts with its name; a package without a package comment |
| `links` | `.md` | none | a relative link to a file, or to a `#anchor` heading, that does not exist |
| `vocab` | any text | `--ban PATTERN` (repeatable, at least one) | a line with a whole-word, case-insensitive match of a Go regular expression |
| `bypass` | any text | none | a line that spells a flag, mode or setting that turns off an agent CLI's permission checks |
| `exttest` | `.go` | none | a `_test.go` file whose package is not `<name>_test` |
| `imports` | `.go` | `--allow PREFIX`, `--allow-exact PATH`, `--deny PREFIX` (repeatable) | an import in a non-test file that is denied, or outside the standard library and not allowed |

`--allow` takes a path and every path below it; `--allow-exact` takes that path
only, for a module's root package. `--deny` wins over both and also applies to
the standard library. `imports` takes the allowed list of its package, so it is
a `lint_rule()`:

```python
lint_rule(
    name = "imports",                    # creates :lint_imports
    srcs = glob(["*.go"]),
    tool = "//:repolint",
    args = [
        "imports",
        "--allow=example.com/mod/store",
        "--allow-exact=example.com/mod",
    ],
)
```

`links` follows links into other directories, so it runs as a test that has
those files as data. [tools/repolint/README.md](tools/repolint/README.md)
describes every check, with that test.

`repolint_test()` runs the checks' own tests against the built program, and
needs the go plugin:

```python
subinclude("///lint//build_defs:repolint_test")

repolint_test(name = "repolint_test", tool = ":repolint")
```

## Config

| Key | Default | Meaning |
|---|---|---|
| `GofmtTool` | unset (skip) | gofmt to run |
| `GoTool` | unset (skip vet; `repolint()` fails) | go, used for `go vet` and to compile `repolint` |
| `PyflakesTool` | unset (skip) | pyflakes executable |
| `Rule` | none | a repo-wide custom rule (repeatable), see above |
| `Label` | `lint` | label on every lint test |
| `Exclude` | none | repo-relative prefixes the coverage guard ignores (repeatable) |

## Limits

- `repo_lint()` runs locally and outside the sandbox. `go vet` needs the whole
  module and the Go module and build caches, which come from `HOME`, `GOPATH`,
  `GOMODCACHE` and `GOCACHE`.
- `repo_lint()` re-runs when a listed package's sources change. A brand-new
  package that no edit to a listed package accompanies is caught on the next
  run after one does, or with `plz test //:vet --rerun`.
- `repolint()` builds offline, with the go it is given and no cgo. A build that
  plz does not take from its cache compiles the standard library packages the
  checks import, which takes a while.
- Unix only: the repo-wide script is POSIX `sh`.

## Licence

MIT. See [LICENSE](LICENSE).
