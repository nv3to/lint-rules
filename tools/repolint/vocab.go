package main

import (
	"flag"
	"fmt"
	"regexp"
	"strings"
)

// wordBounded compiles regular-expression patterns into one that matches any
// of them as a whole word, ignoring case: neighbouring letters and digits do
// not count as a boundary. The text of the first group is the match.
func wordBounded(patterns []string) (*regexp.Regexp, error) {
	parts := make([]string, len(patterns))
	for i, p := range patterns {
		parts[i] = "(?:" + p + ")"
	}
	return regexp.Compile(`(?i)(?:^|[^\pL\pN])(` + strings.Join(parts, "|") + `)(?:$|[^\pL\pN])`)
}

// checkVocab reports every line that contains a term matching a --ban
// pattern. Patterns are regular expressions, matched without regard to case
// and as whole words.
func checkVocab(args []string) ([]finding, error) {
	var bans listFlag
	files, err := parseFlags("vocab", args, func(fs *flag.FlagSet) {
		fs.Var(&bans, "ban", "regular expression of a banned term (repeatable)")
	})
	if err != nil {
		return nil, err
	}
	if len(bans) == 0 {
		return nil, usageError("at least one --ban PATTERN is required")
	}
	re, err := wordBounded(bans)
	if err != nil {
		return nil, usageError(err.Error())
	}
	var findings []finding
	for _, path := range files {
		err := eachLine(path, func(n int, text string) {
			if m := re.FindStringSubmatch(text); m != nil {
				findings = append(findings, finding{path, n, fmt.Sprintf("banned term %q", strings.ToLower(m[1]))})
			}
		})
		if err != nil {
			return nil, err
		}
	}
	return findings, nil
}
