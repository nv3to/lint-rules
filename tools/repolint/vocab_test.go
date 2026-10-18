package main_test

import "testing"

func TestVocab(t *testing.T) {
	ban := []string{"vocab", "--ban", "forbidden", "--ban", "bad[[:space:]]+word"}
	files := func(files ...string) []string { return append(append([]string{}, ban...), files...) }

	t.Run("clean", func(t *testing.T) {
		expect(t, run(t, files("vocab/clean.md")...), 0, "")
	})
	t.Run("case, boundaries and every pattern", func(t *testing.T) {
		expect(t, run(t, files("vocab/clean.md", "vocab/dirty.md")...), 1, lines(
			`vocab/dirty.md:2: banned term "forbidden"`,
			`vocab/dirty.md:3: banned term "forbidden"`,
			`vocab/dirty.md:4: banned term "bad   word"`,
			`vocab/dirty.md:5: banned term "forbidden"`,
		))
	})
	t.Run("an alternation is one pattern", func(t *testing.T) {
		expect(t, run(t, "vocab", "--ban", "first|capitals", "vocab/dirty.md"), 1, lines(
			`vocab/dirty.md:1: banned term "first"`,
			`vocab/dirty.md:2: banned term "capitals"`,
		))
	})
	t.Run("flag with =", func(t *testing.T) {
		expect(t, run(t, "vocab", "--ban=forbidden", "vocab/clean.md"), 0, "")
	})
	t.Run("a pattern that spells nothing it bans", func(t *testing.T) {
		expect(t, run(t, "vocab", "--ban", "[f]orbidden", "vocab/dirty.md"), 1, lines(
			`vocab/dirty.md:2: banned term "forbidden"`,
			`vocab/dirty.md:3: banned term "forbidden"`,
			`vocab/dirty.md:4: banned term "forbidden"`,
			`vocab/dirty.md:5: banned term "forbidden"`,
		))
	})
	t.Run("no pattern is bad usage", func(t *testing.T) {
		if got := run(t, "vocab", "vocab/clean.md"); got.code != 2 {
			t.Errorf("exit %d, want 2", got.code)
		}
	})
	t.Run("an invalid pattern is bad usage", func(t *testing.T) {
		if got := run(t, "vocab", "--ban", "(", "vocab/clean.md"); got.code != 2 {
			t.Errorf("exit %d, want 2", got.code)
		}
	})
}
