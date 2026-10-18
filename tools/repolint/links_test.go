package main_test

import "testing"

func TestLinks(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		expect(t, run(t, "links", "links/ok/src.md", "links/ok/target.md", "links/ok/café.md"), 0, "")
	})
	t.Run("broken", func(t *testing.T) {
		expect(t, run(t, "links", "links/bad/bad.md"), 1, lines(
			"links/bad/bad.md:1: link to nowhere.md: no such file",
			"links/bad/bad.md:2: link to ../ok/target.md#absent: no heading with anchor #absent",
			"links/bad/bad.md:3: link to #absent: no heading with anchor #absent",
			"links/bad/bad.md:4: link to ../ok/gone.md: no such file",
			"links/bad/bad.md:8: link to missing.md: no such file",
			"links/bad/bad.md:12: link to #not-a-heading: no heading with anchor #not-a-heading",
		))
	})
	t.Run("other files are skipped", func(t *testing.T) {
		expect(t, run(t, "links", "doc/skipped/notes.txt", "doc/bad/bad.go"), 0, "")
	})
}
