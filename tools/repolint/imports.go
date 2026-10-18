package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
)

// checkImports reports imports of non-test Go files that are denied, or that
// are outside the standard library and not allowed. A prefix matches whole
// path elements: "a/b" covers "a/b" and "a/b/c" but not "a/bc". An exact path
// covers itself only, for a package whose path is a prefix of others, such as
// a module's root package. A deny prefix wins over an allow and also applies to
// the standard library; with no --allow and no --allow-exact, every import
// outside the standard library is reported.
func checkImports(args []string) ([]finding, error) {
	var allow, exact, deny listFlag
	files, err := parseFlags("imports", args, func(fs *flag.FlagSet) {
		fs.Var(&allow, "allow", "import path prefix to allow outside the standard library (repeatable)")
		fs.Var(&exact, "allow-exact", "import path to allow outside the standard library, without the paths below it (repeatable)")
		fs.Var(&deny, "deny", "import path prefix to report anywhere (repeatable)")
	})
	if err != nil {
		return nil, err
	}
	var findings []finding
	fset := token.NewFileSet()
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") || !strings.HasSuffix(path, ".go") {
			continue
		}
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return nil, err
		}
		for _, spec := range f.Imports {
			imp, err := strconv.Unquote(spec.Path.Value)
			if err != nil || imp == "C" {
				continue
			}
			line := fset.Position(spec.Path.Pos()).Line
			switch {
			case hasPathPrefix(imp, deny):
				findings = append(findings, finding{path, line, fmt.Sprintf("imports %q: denied", imp)})
			case !isStdlib(imp) && !hasPathPrefix(imp, allow) && !hasPath(imp, exact):
				findings = append(findings, finding{path, line, fmt.Sprintf("imports %q: not allowed", imp)})
			}
		}
	}
	return findings, nil
}

// isStdlib reports whether an import path belongs to the standard library:
// its first element has no dot.
func isStdlib(imp string) bool {
	first, _, _ := strings.Cut(imp, "/")
	return !strings.Contains(first, ".")
}

func hasPath(imp string, paths []string) bool {
	for _, p := range paths {
		if imp == strings.TrimSuffix(p, "/") {
			return true
		}
	}
	return false
}

func hasPathPrefix(imp string, prefixes []string) bool {
	for _, p := range prefixes {
		p = strings.TrimSuffix(p, "/")
		if imp == p || strings.HasPrefix(imp, p+"/") {
			return true
		}
	}
	return false
}
