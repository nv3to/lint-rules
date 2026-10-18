package main_test

import "testing"

func TestExtTest(t *testing.T) {
	t.Run("external", func(t *testing.T) {
		expect(t, run(t, "exttest", "exttest/good_test.go", "exttest/lib.go"), 0, "")
	})
	t.Run("internal", func(t *testing.T) {
		expect(t, run(t, "exttest", "exttest/good_test.go", "exttest/internal_test.go", "exttest/lib.go", "exttest/main_test.go"), 1, lines(
			"exttest/internal_test.go:3: test file is in package x; declare package x_test",
			"exttest/main_test.go:1: test file is in package main; declare package main_test",
		))
	})
}
