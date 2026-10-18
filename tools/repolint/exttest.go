package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"strings"
)

// checkExtTest reports every _test.go file whose package is not an external
// test package, one named <name>_test.
func checkExtTest(args []string) ([]finding, error) {
	files, err := parseFlags("exttest", args, nil)
	if err != nil {
		return nil, err
	}
	var findings []finding
	fset := token.NewFileSet()
	for _, path := range files {
		if !strings.HasSuffix(path, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, path, nil, parser.PackageClauseOnly)
		if err != nil {
			return nil, err
		}
		if name := f.Name.Name; !strings.HasSuffix(name, "_test") {
			findings = append(findings, finding{path, fset.Position(f.Name.Pos()).Line,
				fmt.Sprintf("test file is in package %s; declare package %s_test", name, name)})
		}
	}
	return findings, nil
}
