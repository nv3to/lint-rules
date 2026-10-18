// Command repolint runs lint checks over files. Each check is a Please lint
// rule: it takes flags and files, prints one "file:line: message" per finding
// and exits 1 when it found any.
//
//	repolint doc FILE...                       exported identifiers carry doc comments that start with their name
//	repolint links FILE...                     relative markdown links and anchors resolve
//	repolint vocab --ban PATTERN... FILE...    no line contains a banned term
//	repolint bypass FILE...                    no file spells a flag or setting that disables an agent CLI's permission checks
//	repolint exttest FILE...                   every test file declares an external test package
//	repolint imports [--allow PREFIX]... [--deny PREFIX]... FILE...
//	                                           non-test Go files import only what is allowed
//
// The exit status is 0 when nothing was found, 1 when something was, and 2
// for bad usage or a file that cannot be read or parsed.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// finding is one reported problem.
type finding struct {
	file string
	line int
	msg  string
}

func (f finding) String() string {
	return fmt.Sprintf("%s:%d: %s", f.file, f.line, f.msg)
}

// usageError is a mistake in how a check was invoked.
type usageError string

func (e usageError) Error() string { return string(e) }

// check is one lint check: run takes the arguments after its name.
type check struct {
	name string
	run  func(args []string) ([]finding, error)
}

var checks = []check{
	{"doc", checkDocs},
	{"links", checkLinks},
	{"vocab", checkVocab},
	{"bypass", checkBypass},
	{"exttest", checkExtTest},
	{"imports", checkImports},
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		names := make([]string, len(checks))
		for i, c := range checks {
			names[i] = c.name
		}
		fmt.Fprintf(stderr, "usage: repolint %s [FLAG]... FILE...\n", strings.Join(names, "|"))
		return 2
	}
	i := slices.IndexFunc(checks, func(c check) bool { return c.name == args[0] })
	if i < 0 {
		fmt.Fprintf(stderr, "repolint: unknown check %q\n", args[0])
		return 2
	}
	findings, err := checks[i].run(args[1:])
	if err != nil {
		fmt.Fprintf(stderr, "repolint %s: %v\n", args[0], err)
		return 2
	}
	slices.SortStableFunc(findings, func(a, b finding) int {
		if c := strings.Compare(a.file, b.file); c != 0 {
			return c
		}
		return a.line - b.line
	})
	for _, f := range findings {
		fmt.Fprintln(stdout, f)
	}
	if len(findings) > 0 {
		return 1
	}
	return 0
}

// listFlag is a flag that may be repeated.
type listFlag []string

func (l *listFlag) String() string { return strings.Join(*l, " ") }

func (l *listFlag) Set(v string) error {
	*l = append(*l, v)
	return nil
}

// parseFlags parses args against the flags that define declares and returns
// what remains: the files. A bad flag is a usageError.
func parseFlags(name string, args []string, define func(*flag.FlagSet)) ([]string, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if define != nil {
		define(fs)
	}
	if err := fs.Parse(args); err != nil {
		return nil, usageError(err.Error())
	}
	return fs.Args(), nil
}

// eachLine calls fn with every line of the file and its 1-based number.
func eachLine(path string, fn func(n int, text string)) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for n := 1; sc.Scan(); n++ {
		fn(n, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}
