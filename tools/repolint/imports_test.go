package main_test

import "testing"

func TestImports(t *testing.T) {
	t.Run("standard library only", func(t *testing.T) {
		expect(t, run(t, "imports", "imports/ok.go"), 1, lines(
			`imports/ok.go:7: imports "example.com/mod/pkg": not allowed`,
			`imports/ok.go:8: imports "example.com/mod/pkg/sub": not allowed`,
			`imports/ok.go:9: imports "example.com/other/lib": not allowed`,
		))
	})
	t.Run("allowed prefixes", func(t *testing.T) {
		expect(t, run(t, "imports", "--allow", "example.com/mod", "--allow=example.com/other", "imports/ok.go"), 0, "")
	})
	t.Run("an exact path allows itself only", func(t *testing.T) {
		expect(t, run(t, "imports", "--allow-exact", "example.com/mod/pkg", "--allow-exact=example.com/other/lib/", "imports/ok.go"), 1, lines(
			`imports/ok.go:8: imports "example.com/mod/pkg/sub": not allowed`,
		))
	})
	t.Run("a prefix is whole path elements", func(t *testing.T) {
		expect(t, run(t, "imports", "--allow", "example.com/mod", "imports/bad.go"), 1, lines(
			`imports/bad.go:8: imports "example.com/modern/lib": not allowed`,
			`imports/bad.go:10: imports "gopkg.in/yaml.v3": not allowed`,
		))
	})
	t.Run("deny wins and covers the standard library", func(t *testing.T) {
		expect(t, run(t, "imports", "--allow", "example.com", "--deny", "example.com/mod/internal", "--deny", "os", "--deny", "net/http",
			"imports/ok.go", "imports/bad.go"), 1, lines(
			`imports/bad.go:4: imports "net/http": denied`,
			`imports/bad.go:5: imports "os/exec": denied`,
			`imports/bad.go:7: imports "example.com/mod/internal/secret": denied`,
			`imports/bad.go:10: imports "gopkg.in/yaml.v3": not allowed`,
			`imports/ok.go:5: imports "os": denied`,
		))
	})
	t.Run("test files and other files are skipped", func(t *testing.T) {
		expect(t, run(t, "imports", "--deny", "os", "imports/x_test.go", "doc/skipped/notes.txt"), 0, "")
	})
}
