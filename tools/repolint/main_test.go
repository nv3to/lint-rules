package main_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var (
	binary   string // the repolint executable under test
	testdata string // the testdata directory, the working directory of every run
)

// TestMain finds what the tests run on in the named data of the test target:
// $DATA_REPOLINT is the executable and $DATA_TESTDATA the files of testdata/.
func TestMain(m *testing.M) {
	var err error
	if binary, err = dataFile("REPOLINT"); err == nil {
		testdata, err = dataDir("TESTDATA", "testdata")
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "repolint tests: %v; is it named data of the test target?\n", err)
		os.Exit(2)
	}
	os.Exit(m.Run())
}

// dataPaths returns the paths, relative to the working directory, that the
// named data DATA_<name> of the test target holds.
func dataPaths(name string) ([]string, error) {
	paths := strings.Fields(os.Getenv("DATA_" + name))
	if len(paths) == 0 {
		return nil, fmt.Errorf("$DATA_%s is empty", name)
	}
	return paths, nil
}

// dataFile returns the absolute path of the one file in DATA_<name>.
func dataFile(name string) (string, error) {
	paths, err := dataPaths(name)
	if err != nil {
		return "", err
	}
	if len(paths) != 1 {
		return "", fmt.Errorf("$DATA_%s holds %d files, want 1", name, len(paths))
	}
	return existing(paths[0])
}

// dataDir returns the absolute path of the directory called base that the
// files in DATA_<name> are below.
func dataDir(name, base string) (string, error) {
	paths, err := dataPaths(name)
	if err != nil {
		return "", err
	}
	marker := "/" + base + "/"
	for _, p := range paths {
		rooted := "/" + p
		if i := strings.Index(rooted, marker); i >= 0 {
			return existing(rooted[1 : i+len(marker)-1])
		}
	}
	return "", fmt.Errorf("$DATA_%s holds no file below a directory %s", name, base)
}

// existing returns the absolute form of a path that must exist.
func existing(rel string) (string, error) {
	p, err := filepath.Abs(rel)
	if err == nil {
		_, err = os.Stat(p)
	}
	return p, err
}

// result is what one run of repolint produced.
type result struct {
	code   int
	stdout string
	stderr string
}

// runIn runs repolint in dir with args.
func runIn(t *testing.T, dir string, args ...string) result {
	t.Helper()
	cmd := exec.Command(binary, args...)
	cmd.Dir = dir
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	var exit *exec.ExitError
	switch {
	case err == nil:
	case errors.As(err, &exit):
	default:
		t.Fatal(err)
	}
	return result{cmd.ProcessState.ExitCode(), out.String(), errOut.String()}
}

// run runs repolint in the testdata directory.
func run(t *testing.T, args ...string) result {
	t.Helper()
	return runIn(t, testdata, args...)
}

// lines joins lines the way repolint prints them.
func lines(ls ...string) string {
	return strings.Join(ls, "\n") + "\n"
}

// expect fails unless the run exited with code and printed exactly stdout.
func expect(t *testing.T, got result, code int, stdout string) {
	t.Helper()
	if got.code != code || got.stdout != stdout {
		t.Errorf("exit %d, want %d; stdout:\n%s\nwant:\n%s\nstderr:\n%s", got.code, code, got.stdout, stdout, got.stderr)
	}
}

// writeFiles writes name-to-body files below a new temporary directory.
func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestUsage(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"no check", nil},
		{"unknown check", []string{"nope"}},
		{"unknown flag", []string{"doc", "--nope", "x.go"}},
		{"missing file", []string{"doc", "missing.go"}},
		{"missing markdown file", []string{"links", "missing.md"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := run(t, c.args...)
			if got.code != 2 || got.stdout != "" || got.stderr == "" {
				t.Errorf("exit %d, stdout %q, stderr %q; want exit 2, no stdout and a message", got.code, got.stdout, got.stderr)
			}
		})
	}
}

func TestSyntaxErrorIsNotAFinding(t *testing.T) {
	dir := writeFiles(t, map[string]string{"broken.go": "package\n"})
	got := runIn(t, dir, "imports", "broken.go")
	if got.code != 2 || got.stdout != "" {
		t.Errorf("exit %d, stdout %q; want exit 2", got.code, got.stdout)
	}
}

func TestNoFilesNoFindings(t *testing.T) {
	for _, check := range []string{"doc", "links", "bypass", "exttest", "imports"} {
		if got := run(t, check); got.code != 0 || got.stdout != "" {
			t.Errorf("%s with no files: exit %d, stdout %q", check, got.code, got.stdout)
		}
	}
}
