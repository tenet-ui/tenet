package designsystem

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestScaffoldWritesEveryFile(t *testing.T) {
	root := t.TempDir()
	if err := Scaffold(root); err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	for _, rel := range []string{
		PointOfViewPath,
		TokensPath,
		RulesPath,
		ComponentsDir + "/.gitkeep",
		TemplatesDir + "/.gitkeep",
		FlowsDir + "/.gitkeep",
	} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			t.Errorf("missing %s: %v", rel, err)
		}
	}
}

func TestScaffoldThenLoadIsValid(t *testing.T) {
	root := t.TempDir()
	if err := Scaffold(root); err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	pack, problems := Load(root)
	checkProblems(t, problems, nil)
	if pack.PointOfView == "" {
		t.Errorf("point of view is empty in scaffold")
	}
	if len(pack.Tokens) != 0 || len(pack.Rules) != 0 || len(pack.Components) != 0 ||
		len(pack.Templates) != 0 || len(pack.Flows) != 0 {
		t.Errorf("scaffold is not empty: %+v", pack)
	}
}

func TestScaffoldRefusesExistingFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, DirName)
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := Scaffold(root)
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("err = %v, want ErrAlreadyExists", err)
	}
	if data, _ := os.ReadFile(path); string(data) != "x" {
		t.Errorf("existing file was changed to %q", data)
	}
}

func TestScaffoldRefusesExistingDirectory(t *testing.T) {
	root := t.TempDir()
	keep := filepath.Join(root, DirName, "keep.md")
	if err := os.MkdirAll(filepath.Dir(keep), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keep, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Scaffold(root); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("err = %v, want ErrAlreadyExists", err)
	}
	if _, err := os.Stat(filepath.Join(root, PointOfViewPath)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("scaffold wrote into an existing folder: %v", err)
	}
	if _, err := os.Stat(keep); err != nil {
		t.Errorf("existing file was removed: %v", err)
	}
}

func TestScaffoldRefusesDanglingSymlink(t *testing.T) {
	root := t.TempDir()
	link := filepath.Join(root, DirName)
	if err := os.Symlink(filepath.Join(root, "nowhere"), link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := Scaffold(root); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("err = %v, want ErrAlreadyExists", err)
	}
	if _, err := os.Lstat(link); err != nil {
		t.Errorf("symlink was removed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "nowhere")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("scaffold wrote through the symlink: %v", err)
	}
}
