package main_test

import "testing"

func TestDoc(t *testing.T) {
	t.Run("documented", func(t *testing.T) {
		expect(t, run(t, "doc", "doc/good/doc.go", "doc/good/ok.go"), 0, "")
	})
	t.Run("undocumented", func(t *testing.T) {
		expect(t, run(t, "doc", "doc/bad/bad.go"), 1, lines(
			"doc/bad/bad.go:1: package has no package comment",
			"doc/bad/bad.go:3: exported type Bare has no doc comment",
			`doc/bad/bad.go:6: the comment on exported func Exported should start with "Exported"`,
			"doc/bad/bad.go:8: exported method Bare.Method has no doc comment",
			"doc/bad/bad.go:10: exported var Loose has no doc comment",
			"doc/bad/bad.go:16: exported const Ungrouped has no doc comment",
		))
	})
	t.Run("skipped files", func(t *testing.T) {
		expect(t, run(t, "doc", "doc/skipped/x_test.go", "doc/skipped/main.go", "doc/skipped/notes.txt"), 0, "")
	})
	t.Run("package comment from any file", func(t *testing.T) {
		expect(t, run(t, "doc", "doc/good/ok.go", "doc/good/doc.go"), 0, "")
	})
	t.Run("findings of several packages are ordered by file", func(t *testing.T) {
		expect(t, run(t, "doc", "doc/bad/bad.go", "doc/good/ok.go"), 1, lines(
			"doc/bad/bad.go:1: package has no package comment",
			"doc/bad/bad.go:3: exported type Bare has no doc comment",
			`doc/bad/bad.go:6: the comment on exported func Exported should start with "Exported"`,
			"doc/bad/bad.go:8: exported method Bare.Method has no doc comment",
			"doc/bad/bad.go:10: exported var Loose has no doc comment",
			"doc/bad/bad.go:16: exported const Ungrouped has no doc comment",
			"doc/good/ok.go:1: package has no package comment",
		))
	})
}
