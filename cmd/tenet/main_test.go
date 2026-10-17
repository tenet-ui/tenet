package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseRootDefaultsToCwd(t *testing.T) {
	cwd := t.TempDir()

	got, err := parseRoot(nil, cwd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != cwd {
		t.Fatalf("got %q, want %q", got, cwd)
	}
}

func TestParseRootAbsoluteDir(t *testing.T) {
	cwd := t.TempDir()
	dir := t.TempDir()

	got, err := parseRoot([]string{"--dir", dir}, cwd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != filepath.Clean(dir) {
		t.Fatalf("got %q, want %q", got, filepath.Clean(dir))
	}
}

func TestParseRootRelativeDir(t *testing.T) {
	cwd := t.TempDir()
	if err := os.Mkdir(filepath.Join(cwd, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := parseRoot([]string{"--dir", "sub"}, cwd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := filepath.Join(cwd, "sub"); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestParseRootMissingPath(t *testing.T) {
	cwd := t.TempDir()

	_, err := parseRoot([]string{"--dir", "does-not-exist"}, cwd)
	if err == nil {
		t.Fatal("expected an error for a missing path")
	}
}

func TestParseRootFileNotDirectory(t *testing.T) {
	cwd := t.TempDir()
	file := filepath.Join(cwd, "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := parseRoot([]string{"--dir", file}, cwd)
	if err == nil {
		t.Fatal("expected an error when --dir is a file")
	}
}

func TestParseRootUnknownFlag(t *testing.T) {
	_, err := parseRoot([]string{"--nope"}, t.TempDir())
	if err == nil {
		t.Fatal("expected an error for an unknown flag")
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"nope"}, t.TempDir(), &stdout, &stderr); code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), `unknown command "nope"`) {
		t.Errorf("stderr = %q", stderr.String())
	}
}

func TestRunValidateFixture(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{"validate", "--dir", "../../testdata/project"}, cwd, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}
	want := "design system valid: 4 tokens, 2 rules, 1 components, 1 templates, 1 flows"
	if !strings.Contains(stdout.String(), want) {
		t.Errorf("stdout = %q, want it to contain %q", stdout.String(), want)
	}
}

func TestRunValidateReportsFileAndLine(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"design-system/point-of-view.md": "# Brand\n",
		"design-system/rules.md":         "# Rules\n",
		"design-system/tokens.json":      "{\n  \"color\": {\n    \"ink\": true\n  }\n}\n",
	})
	var stdout, stderr bytes.Buffer
	code := run([]string{"validate", "--dir", root}, t.TempDir(), &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	out := stdout.String()
	if !strings.Contains(out, `design-system/tokens.json:3: token "color.ink" must be a string or number`) {
		t.Errorf("stdout = %q, want a file:line problem", out)
	}
	if !strings.Contains(out, "1 problem(s) found") {
		t.Errorf("stdout = %q, want the problem count", out)
	}
}

func TestRunValidateMissingDesignSystem(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"validate", "--dir", t.TempDir()}, t.TempDir(), &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(stdout.String(), "directory not found") {
		t.Errorf("stdout = %q", stdout.String())
	}
}

func TestRunValidateBadRoot(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"validate", "--dir", "does-not-exist"}, t.TempDir(), &stdout, &stderr)
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
}

// writeFiles creates each file under root, making parent directories as needed.
func writeFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
