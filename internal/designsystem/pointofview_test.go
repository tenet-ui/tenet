package designsystem

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func writePointOfView(t *testing.T, root string, content string) {
	t.Helper()
	dir := filepath.Join(root, "design-system")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "point-of-view.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestReadPointOfViewReturnsContents(t *testing.T) {
	root := t.TempDir()
	writePointOfView(t, root, "# Point of view\n\nWe are calm.\n")

	got, err := ReadPointOfView(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "# Point of view\n\nWe are calm.\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestReadPointOfViewMissingDesignSystemDir(t *testing.T) {
	_, err := ReadPointOfView(t.TempDir())
	if !errors.Is(err, ErrPointOfViewMissing) {
		t.Fatalf("got %v, want ErrPointOfViewMissing", err)
	}
}

func TestReadPointOfViewMissingFileInExistingDir(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "design-system"), 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := ReadPointOfView(root)
	if !errors.Is(err, ErrPointOfViewMissing) {
		t.Fatalf("got %v, want ErrPointOfViewMissing", err)
	}
}

func TestReadPointOfViewEmptyFile(t *testing.T) {
	root := t.TempDir()
	writePointOfView(t, root, "")

	_, err := ReadPointOfView(root)
	if !errors.Is(err, ErrPointOfViewEmpty) {
		t.Fatalf("got %v, want ErrPointOfViewEmpty", err)
	}
}

func TestReadPointOfViewWhitespaceOnlyFile(t *testing.T) {
	root := t.TempDir()
	writePointOfView(t, root, "  \n\t\n")

	_, err := ReadPointOfView(root)
	if !errors.Is(err, ErrPointOfViewEmpty) {
		t.Fatalf("got %v, want ErrPointOfViewEmpty", err)
	}
}

func TestReadPointOfViewDirectoryAtPath(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "design-system", "point-of-view.md"), 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := ReadPointOfView(root)
	if err == nil {
		t.Fatal("expected an error")
	}
	if errors.Is(err, ErrPointOfViewMissing) || errors.Is(err, ErrPointOfViewEmpty) {
		t.Fatalf("got %v, want a generic read error", err)
	}
}

func TestReadPointOfViewDesignSystemIsFile(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "design-system"), []byte("not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ReadPointOfView(root)
	if err == nil {
		t.Fatal("expected an error")
	}
	if errors.Is(err, ErrPointOfViewMissing) || errors.Is(err, ErrPointOfViewEmpty) {
		t.Fatalf("got %v, want a generic read error", err)
	}
}
