package designsystem

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// ErrAlreadyExists reports that design-system already exists, so Scaffold wrote nothing.
var ErrAlreadyExists = errors.New("design-system already exists")

const (
	pointOfViewScaffold = "# Point of view\n\nDescribe what this product stands for, who it is for, and what its users care about. Every rule should trace back to this.\n"
	tokensScaffold      = "{}\n"
	rulesScaffold       = "# Rules\n\nEach rule is a level-2 heading with its id, then a statement and a Reason: line.\n"
)

// scaffoldFiles maps each file Scaffold writes, relative to the project root, to its content.
var scaffoldFiles = map[string]string{
	PointOfViewPath:             pointOfViewScaffold,
	TokensPath:                  tokensScaffold,
	RulesPath:                   rulesScaffold,
	ComponentsDir + "/.gitkeep": "",
	TemplatesDir + "/.gitkeep":  "",
	FlowsDir + "/.gitkeep":      "",
}

// Scaffold writes an empty, valid design-system/ folder under root. It returns an
// error wrapping ErrAlreadyExists, and writes nothing, when the path exists in any
// form, including a dangling symlink.
func Scaffold(root string) error {
	dir := filepath.Join(root, DirName)
	if _, err := os.Lstat(dir); err == nil {
		return fmt.Errorf("%w under %s; nothing written", ErrAlreadyExists, root)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := os.Mkdir(dir, 0o755); err != nil {
		return err
	}
	if err := writeScaffold(root); err != nil {
		_ = os.RemoveAll(dir)
		return err
	}
	return nil
}

// writeScaffold creates each scaffold file with exclusive create, so it never overwrites.
func writeScaffold(root string) error {
	for rel, content := range scaffoldFiles {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			return err
		}
		f, err := os.OpenFile(abs, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return err
		}
		if _, err := f.WriteString(content); err != nil {
			_ = f.Close()
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	}
	return nil
}
