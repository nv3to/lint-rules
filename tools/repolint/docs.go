package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
)

// interfaceMethods are exported methods whose meaning is fixed by a standard
// interface; they need no comment of their own.
var interfaceMethods = []string{
	"Error", "String", "Unwrap", "ServeHTTP", "MarshalJSON", "UnmarshalJSON",
	"MarshalYAML", "UnmarshalYAML", "MarshalText", "UnmarshalText",
}

// checkDocs reports exported identifiers without a doc comment, comments that
// do not start with the identifier, and packages without a package comment.
// Test files, other files and main packages are skipped.
func checkDocs(args []string) ([]finding, error) {
	files, err := parseFlags("doc", args, nil)
	if err != nil {
		return nil, err
	}
	var findings []finding
	fset := token.NewFileSet()
	documented := map[string]bool{}
	first := map[string]string{}
	var dirs []string
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") || !strings.HasSuffix(path, ".go") {
			continue
		}
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		if f.Name.Name == "main" {
			continue
		}
		dir := filepath.Dir(path)
		if _, ok := first[dir]; !ok {
			first[dir] = path
			dirs = append(dirs, dir)
		}
		if f.Doc != nil {
			documented[dir] = true
		}
		findings = append(findings, fileDocs(fset, f)...)
	}
	for _, dir := range dirs {
		if !documented[dir] {
			findings = append(findings, finding{first[dir], 1, "package has no package comment"})
		}
	}
	return findings, nil
}

func fileDocs(fset *token.FileSet, f *ast.File) []finding {
	var out []finding
	report := func(pos token.Pos, kind, name string, doc *ast.CommentGroup) {
		p := fset.Position(pos)
		switch {
		case doc == nil:
			out = append(out, finding{p.Filename, p.Line, fmt.Sprintf("exported %s %s has no doc comment", kind, name)})
		case !startsWithName(doc.Text(), name):
			out = append(out, finding{p.Filename, p.Line, fmt.Sprintf("the comment on exported %s %s should start with %q", kind, name, name)})
		}
	}
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if !d.Name.IsExported() {
				continue
			}
			if d.Recv != nil {
				recv := receiverName(d.Recv)
				if !ast.IsExported(recv) || slices.Contains(interfaceMethods, d.Name.Name) {
					continue
				}
				report(d.Pos(), "method", recv+"."+d.Name.Name, d.Doc)
				continue
			}
			report(d.Pos(), "func", d.Name.Name, d.Doc)
		case *ast.GenDecl:
			grouped := d.Lparen.IsValid() && d.Doc != nil
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name.IsExported() {
						report(s.Pos(), "type", s.Name.Name, docOf(s.Doc, d))
					}
				case *ast.ValueSpec:
					if grouped {
						continue
					}
					for _, n := range s.Names {
						if n.IsExported() {
							doc := docOf(s.Doc, d)
							if doc != nil && len(s.Names) > 1 {
								break
							}
							report(n.Pos(), kindOf(d.Tok), n.Name, doc)
							break
						}
					}
				}
			}
		}
	}
	return out
}

// docOf is a spec's own comment, or its declaration's when the declaration
// holds only this spec.
func docOf(spec *ast.CommentGroup, d *ast.GenDecl) *ast.CommentGroup {
	if spec != nil {
		return spec
	}
	if !d.Lparen.IsValid() {
		return d.Doc
	}
	return nil
}

func kindOf(tok token.Token) string {
	if tok == token.CONST {
		return "const"
	}
	return "var"
}

func receiverName(fl *ast.FieldList) string {
	if fl == nil || len(fl.List) == 0 {
		return ""
	}
	t := fl.List[0].Type
	for {
		switch x := t.(type) {
		case *ast.StarExpr:
			t = x.X
		case *ast.IndexExpr:
			t = x.X
		case *ast.IndexListExpr:
			t = x.X
		case *ast.Ident:
			return x.Name
		default:
			return ""
		}
	}
}

// startsWithName accepts "Name …" and "A Name …", "An Name …", "The Name …".
func startsWithName(text, name string) bool {
	short := name
	if i := strings.LastIndex(name, "."); i >= 0 {
		short = name[i+1:]
	}
	text = strings.TrimSpace(text)
	for _, article := range []string{"", "A ", "An ", "The "} {
		rest, ok := strings.CutPrefix(text, article)
		if !ok {
			continue
		}
		if after, ok := strings.CutPrefix(rest, short); ok && (after == "" || !isIdentChar(after[0])) {
			return true
		}
	}
	return false
}

func isIdentChar(c byte) bool {
	return c == '_' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}
