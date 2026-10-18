# AGENTS.md — lint-rules

Instructions for AI coding agents and contributors. `CLAUDE.md` and
`GEMINI.md` are symlinks to this file. User documentation is in
[README.md](README.md).

## Layout

| Path | What |
|---|---|
| `.plzconfig` | `[PluginDefinition]` and the `[PluginConfig]` keys |
| `build_defs/lint.build_defs` | `lint()`, `lint_rule()`, `repo_lint()` and `repolint()` |
| `build_defs/repolint_test.build_defs` | `repolint_test()`; a file of its own, as it needs the go plugin |
| `tools/repo_lint.sh` | the coverage guard and `go vet` (POSIX sh) |
| `tools/repolint/` | the built-in checks: Go source with a `go.mod` that has no requirements, its tests, testdata and README |

## Rules

- **Nothing is downloaded or pinned here.** Every tool comes from a config key
  that the consumer sets. The one program built is `tools/repolint`, from its
  source, by the consumer's `go_tool`. Do not add a download, a toolchain or a
  dependency; its `go.mod` stays free of requirements.
- **An unset tool key skips that linter.** It never fails. The exception is
  `repolint()`, which a consumer calls to get the checks: it fails with a
  message when `go_tool` is unset.
- **`lint()` stays hermetic.** Only `repo_lint()` is local and unsandboxed.
- **No consumer names.** Nothing, `tools/repolint` included, may refer to a
  particular repo.
- **A new linter is a custom rule first.** Consumers add it with a `Rule`
  config line or `lint_rule()`. Add a check to `tools/repolint` only when it
  needs more than "run the tool over the files", such as parsing. A check uses
  the standard library only, takes what is specific to a repo as flags, prints
  `file:line: message`, exits 1 on findings and 2 on bad usage, and has tests
  and testdata.
- **Config keys, the `Rule` format, function signatures and target names
  (`lint`, `lint_gofmt`, `lint_pyflakes`, `lint_<name>`) are the public API,**
  and so are the check names and flags of `repolint`. Renaming one breaks every
  consumer.

## Testing a change

This plugin has no build of its own. Verify a change in a consuming repo that
calls `repolint()` and `repolint_test()` (see the README) and has `Rule` lines
for the built-in checks:

1. `plz test //... --include lint` is green.
2. `plz test //:repolint_test` is green: it runs the tests of `tools/repolint`
   against the program `repolint()` built. The tests need the named data that
   `repolint_test()` gives them, so they run no other way.
3. Plant each defect and confirm that its gate fails, then revert:
   - a misformatted `.go` file fails `lint_gofmt`;
   - an unused Python import fails `lint_pyflakes`;
   - `fmt.Printf("%d", "x")` fails `//:vet`;
   - a new package without `lint()` fails `//:vet`;
   - with `Rule = shsyntax .sh /bin/sh -n`, a broken `.sh` file fails
     `lint_shsyntax` and an unlinted one fails `//:vet`;
   - with `Rule = godoc .go //:repolint doc`, an exported function without a
     doc comment fails `lint_godoc`;
   - with `GoTool` unset, parsing the BUILD file that calls `repolint()` fails
     with a message that names `GoTool`.
