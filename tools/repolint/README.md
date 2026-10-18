# repolint

The built-in checks of the [lint plugin](../../README.md#built-in-checks). Each
check is a command, `repolint CHECK [FLAG]... FILE...`, that uses only the Go
standard library and nothing of the repository it runs in: what is specific
arrives as flags.

A `Rule` line runs the tool over each package's files with the listed
extensions and fails the package's `lint_<name>` test on a non-zero exit. The
examples name the target `//:repolint`, which `repolint()` builds in the root
BUILD file:

```ini
; Rule = <name> <extensions> <tool> [args...]
Rule = godoc .go //:repolint doc
```

The plugin splits the line on spaces and hands the arguments to a shell, so an
argument holds no space (write `[[:space:]]`), no backslash, `;` or `#`, and
quotes a regular expression with single quotes.

| Output and exit status | |
|---|---|
| stdout | one `file:line: message` per finding, ordered by file and line |
| 0 | nothing found |
| 1 | at least one finding |
| 2 | bad usage, or a file that cannot be read or parsed (the message is on stderr) |

## doc

```ini
Rule = godoc .go //:repolint doc
```

Takes no flags. Over the `.go` files given, skipping `_test.go` files and
`package main`, it fails on:

- an exported function, method, type, variable or constant without a doc
  comment, or whose comment does not start with its name (a leading `A`, `An`
  or `The` is accepted). A declaration group's own comment covers its
  variables and constants.
- an exported method of an unexported type, and the methods of standard
  interfaces (`Error`, `String`, `Unwrap`, `ServeHTTP`, and the `Marshal` and
  `Unmarshal` methods of JSON, YAML and text), are not checked.
- a directory of given files, none of which has a package comment.

## links

Takes no flags. Over the `.md` files given, it fails on a relative link whose
file does not exist, and on a `#anchor` that names no heading of the target
`.md` file. Anchors are the headings in lower case, punctuation dropped,
spaces as hyphens, repeats numbered `-1`, `-2`. Links with a scheme, links
inside fenced code blocks (backticks or tildes) and code spans are ignored,
and percent-encoding in a link is decoded.

A `Rule` line runs in a package's sandbox, where the files of other
directories are absent, so links to them would fail. Run the check as a test
whose data holds every file the links reach:

```starlark
gentest(
    name = "lint_links",
    data = glob(["*.md"]) + ["//docs:pages"],
    labels = ["lint"],
    no_test_output = True,
    test_cmd = "$TOOLS links $DATA",
    test_tools = ["//:repolint"],
)
```

## vocab

```ini
Rule = vocab .go,.md //:repolint vocab --ban 'foo[b]ar' --ban 'old[[:space:]]name'
```

| Flag | |
|---|---|
| `--ban PATTERN` | a regular expression ([Go syntax](https://pkg.go.dev/regexp/syntax)) for a banned term; repeatable, at least one required |

Fails on each line holding a match of any pattern. Matching ignores case and
is by whole word: the characters before and after a match must not be letters
or digits. A pattern is one alternative of an alternation, so `a|b` is two
terms. The first match of a line is reported.

A character class (`foo[b]ar`) lets the `Rule` line match a term without
spelling it, so that the configuration passes its own check.

## bypass

```ini
Rule = bypass .go,.md,.yaml,.yml,.sh,.json //:repolint bypass
```

Takes no flags. Fails on each line that spells a flag, mode or setting that
turns an agent CLI's permission checks off:

- the flags of Claude Code, Codex and Gemini CLI that skip permission prompts,
  allow skipping them, or bypass approvals and the sandbox
- the permission mode that bypasses the checks, in a flag, a settings file or
  a JSON document
- the approve-everything flag and approval mode of Codex and Gemini CLI

The terms are kept as separate words in the code, so the tool's own sources
pass its check. A term matches as a whole word, ignoring case, with `-`, `_` or
nothing between its words (`--flag-name`, `flag_name`, `flagName`). The
message does not repeat the term. A short flag (`-y`) is not reported: it
means something else without its program's name.

## exttest

```ini
Rule = exttest .go //:repolint exttest
```

Takes no flags. Fails on a `_test.go` file whose package clause is not an
external test package, `package <name>_test`. The go plugin's `go_test` rule
with `external = True` compiles only the test sources, so a test in the
package under test cannot see the package.

## imports

```sh
repolint imports [--allow PREFIX]... [--allow-exact PATH]... [--deny PREFIX]... FILE...
```

| Flag | |
|---|---|
| `--allow PREFIX` | an import path prefix allowed outside the standard library; repeatable |
| `--allow-exact PATH` | an import path allowed outside the standard library without the paths below it, such as a module's root package, whose path is a prefix of every package; repeatable |
| `--deny PREFIX` | an import path prefix that is never allowed, the standard library included; repeatable |

Over the `.go` files given, skipping `_test.go` files, fails on each import
that matches a deny prefix, or that is outside the standard library (its first
path element has a dot) and matches no allow prefix or exact path. A prefix
matches whole path elements: `a/b` covers `a/b` and `a/b/c`, not `a/bc`. Deny
wins over both allows, and with neither every import outside the standard
library fails. `import "C"` is ignored.

A `Rule` line passes the same arguments to every package, but a package's
allowed imports differ, so each package lists its own:

```starlark
lint_rule(
    name = "imports",
    srcs = glob(["*.go"]),
    tool = "//:repolint",
    args = ["imports", "--allow", "example.com/mod/expr", "--deny", "os"],
)
```

## Tests

The tests run the built `repolint` over `testdata/` and over files they write
at run time. `repolint_test()` (see the
[plugin README](../../README.md#built-in-checks)) runs them in a consuming
repo with the program and the testdata files as named data, `$DATA_REPOLINT`
and `$DATA_TESTDATA`; they fail when either is missing. Every `bypass` fixture
is written at run time, from pieces, so that no stored file spells a bypass
flag.
