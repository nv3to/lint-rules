package main_test

import (
	"strings"
	"testing"
)

// The flags, modes and settings the check reports, built from pieces so that
// this file does not spell them.
var (
	skipFlag     = "--danger" + "ously-" + "skip-" + "permissions"
	camelFlag    = "--danger" + "ously" + "Skip" + "Permissions"
	allowFlag    = "--allow-" + "danger" + "ously-" + "skip-" + "permissions"
	codexFlag    = "--danger" + "ously-" + "bypass-" + "approvals-and-sandbox"
	approveAll   = "yo" + "lo"
	claudeBypass = "bypass" + "Permissions"
)

func TestBypass(t *testing.T) {
	t.Run("clean", func(t *testing.T) {
		expect(t, run(t, "bypass", "bypass/clean.md"), 0, "")
	})

	msg := "spells a flag or setting that disables permission checks"
	t.Run("spelled", func(t *testing.T) {
		dir := writeFiles(t, map[string]string{
			"flags.md": lines(
				"clean",
				"run claude "+skipFlag+" now",
				camelFlag+"=true",
				"`"+allowFlag+"`",
				"codex exec "+codexFlag,
				"--"+approveAll,
				"approval_mode: "+approveAll,
				"- "+strings.ToUpper(approveAll),
				"--permission-mode="+claudeBypass,
				"permission_mode: "+strings.ToLower(claudeBypass),
				`{"permissions":{"defaultMode":"`+claudeBypass+`"}}`,
				strings.ReplaceAll(skipFlag, "-", "_"),
			),
			"script.sh": "claude -p x " + skipFlag + "\n",
		})
		expect(t, runIn(t, dir, "bypass", "flags.md", "script.sh"), 1, lines(
			"flags.md:2: "+msg,
			"flags.md:3: "+msg,
			"flags.md:4: "+msg,
			"flags.md:5: "+msg,
			"flags.md:6: "+msg,
			"flags.md:7: "+msg,
			"flags.md:8: "+msg,
			"flags.md:9: "+msg,
			"flags.md:10: "+msg,
			"flags.md:11: "+msg,
			"flags.md:12: "+msg,
			"script.sh:1: "+msg,
		))
	})
	t.Run("near misses", func(t *testing.T) {
		dir := writeFiles(t, map[string]string{"near.md": lines(
			approveAll+"v5 and "+approveAll+"x",
			"my"+approveAll,
			claudeBypass+"x and x"+claudeBypass,
			"--danger"+"ously-skip",
			"bypass permissions, skip permissions",
		)})
		expect(t, runIn(t, dir, "bypass", "near.md"), 0, "")
	})
	t.Run("the message does not spell the term", func(t *testing.T) {
		dir := writeFiles(t, map[string]string{"f.md": skipFlag + "\n"})
		got := runIn(t, dir, "bypass", "f.md")
		if strings.Contains(strings.ToLower(got.stdout), "danger"+"ously") {
			t.Errorf("output spells the flag:\n%s", got.stdout)
		}
	})
}
