package main

import (
	"os"
	"path/filepath"
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
