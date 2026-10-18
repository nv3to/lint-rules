package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var (
	linkRE       = regexp.MustCompile(`\[[^\]]*\]\(([^)\s]+)(?:\s+"[^"]*")?\)`)
	codeSpanRE   = regexp.MustCompile("`[^`]*`")
	headingRE    = regexp.MustCompile(`^#{1,6}\s+(.*?)\s*#*\s*$`)
	externalLink = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*:`)
)

// checkLinks reports relative markdown links whose file or anchor does not
// exist. Links inside fenced code blocks and code spans are ignored, and so
// are files without the .md extension.
func checkLinks(args []string) ([]finding, error) {
	files, err := parseFlags("links", args, nil)
	if err != nil {
		return nil, err
	}
	var findings []finding
	anchors := map[string]map[string]bool{}
	anchorsOf := func(path string) (map[string]bool, error) {
		if a, ok := anchors[path]; ok {
			return a, nil
		}
		a, err := headingAnchors(path)
		if err != nil {
			return nil, err
		}
		anchors[path] = a
		return a, nil
	}
	for _, path := range files {
		if !strings.HasSuffix(path, ".md") {
			continue
		}
		var fence fenceState
		var scanErr error
		err := eachLine(path, func(n int, line string) {
			if scanErr != nil || fence.toggles(line) || fence.open() {
				return
			}
			for _, m := range linkRE.FindAllStringSubmatch(codeSpanRE.ReplaceAllString(line, ""), -1) {
				target := m[1]
				if externalLink.MatchString(target) {
					continue
				}
				file, frag, _ := strings.Cut(target, "#")
				if u, err := url.PathUnescape(file); err == nil {
					file = u
				}
				if u, err := url.PathUnescape(frag); err == nil {
					frag = u
				}
				resolved := path
				if file != "" {
					resolved = filepath.Join(filepath.Dir(path), file)
					if _, err := os.Stat(resolved); err != nil {
						findings = append(findings, finding{path, n, fmt.Sprintf("link to %s: no such file", file)})
						continue
					}
				}
				if frag == "" || !strings.HasSuffix(resolved, ".md") {
					continue
				}
				a, err := anchorsOf(resolved)
				if err != nil {
					scanErr = err
					return
				}
				if !a[frag] {
					findings = append(findings, finding{path, n, fmt.Sprintf("link to %s: no heading with anchor #%s", target, frag)})
				}
			}
		})
		if err == nil {
			err = scanErr
		}
		if err != nil {
			return nil, err
		}
	}
	return findings, nil
}

// fenceState follows fenced code blocks, opened and closed by three or more
// backticks or tildes.
type fenceState struct{ marker string }

// toggles reports whether the line opens or closes a fence.
func (s *fenceState) toggles(line string) bool {
	t := strings.TrimSpace(line)
	for _, m := range []string{"```", "~~~"} {
		if !strings.HasPrefix(t, m) {
			continue
		}
		switch {
		case s.marker == "":
			s.marker = m
			return true
		case s.marker == m:
			s.marker = ""
			return true
		}
	}
	return false
}

func (s *fenceState) open() bool { return s.marker != "" }

// headingAnchors are the anchors a markdown renderer gives a file's headings:
// lower case, punctuation dropped, spaces as hyphens, repeats numbered.
func headingAnchors(path string) (map[string]bool, error) {
	out := map[string]bool{}
	count := map[string]int{}
	var fence fenceState
	err := eachLine(path, func(_ int, line string) {
		if fence.toggles(line) || fence.open() {
			return
		}
		m := headingRE.FindStringSubmatch(line)
		if m == nil {
			return
		}
		slug := slugify(m[1])
		if n := count[slug]; n > 0 {
			out[fmt.Sprintf("%s-%d", slug, n)] = true
		} else {
			out[slug] = true
		}
		count[slug]++
	})
	return out, err
}

func slugify(heading string) string {
	heading = strings.ReplaceAll(heading, "`", "")
	var b strings.Builder
	for _, r := range strings.ToLower(heading) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('-')
		}
	}
	return b.String()
}
