package main

import (
	"regexp"
	"strings"
)

// bypassTerms are the flags, modes and settings that turn an agent CLI's
// permission checks off, as the words they are made of. They are kept in
// pieces so that this file does not spell them.
var bypassTerms = [][]string{
	{"dangerously", "skip", "permissions"}, // also the "allow-" variant
	{"dangerously", "bypass", "approvals"}, // up to "-and-sandbox"
	{"bypass", "permissions"},              // a permission mode
	{"yo" + "lo"},                          // an approval mode and a flag
}

// bypassRE matches a term as a whole word, ignoring case, with '-' or '_'
// or nothing between its words, so that --flag-name, flag_name and flagName
// all match.
var bypassRE = func() *regexp.Regexp {
	alts := make([]string, len(bypassTerms))
	for i, words := range bypassTerms {
		quoted := make([]string, len(words))
		for j, w := range words {
			quoted[j] = regexp.QuoteMeta(w)
		}
		alts[i] = strings.Join(quoted, "[-_]*")
	}
	return regexp.MustCompile(`(?i)(?:^|[^\pL\pN])(?:` + strings.Join(alts, "|") + `)(?:$|[^\pL\pN])`)
}()

// checkBypass reports every line that spells a flag, mode or setting that
// disables an agent CLI's permission checks.
func checkBypass(args []string) ([]finding, error) {
	files, err := parseFlags("bypass", args, nil)
	if err != nil {
		return nil, err
	}
	var findings []finding
	for _, path := range files {
		err := eachLine(path, func(n int, text string) {
			if bypassRE.MatchString(text) {
				findings = append(findings, finding{path, n, "spells a flag or setting that disables permission checks"})
			}
		})
		if err != nil {
			return nil, err
		}
	}
	return findings, nil
}
