package designsystem

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var itemIDPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// loadItems reads the Markdown items directly inside root/relDir. relDir is
// slash-separated and relative to root. A missing directory yields no items.
func loadItems(root, relDir string) ([]Item, []Problem) {
	abs := filepath.Join(root, filepath.FromSlash(relDir))
	info, err := os.Stat(abs)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, []Problem{{File: relDir, Message: "could not read: " + err.Error()}}
	}
	if !info.IsDir() {
		return nil, []Problem{{File: relDir, Message: "not a directory"}}
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, []Problem{{File: relDir, Message: "could not read: " + err.Error()}}
	}

	var (
		items    []Item
		problems []Problem
	)
	for _, e := range entries {
		if e.IsDir() || path.Ext(e.Name()) != ".md" {
			continue
		}
		file := relDir + "/" + e.Name()
		id := strings.TrimSuffix(e.Name(), ".md")
		if !itemIDPattern.MatchString(id) {
			problems = append(problems, Problem{File: file, Message: `file name must be lowercase kebab-case, for example "primary-button.md"`})
			continue
		}
		data, err := os.ReadFile(filepath.Join(abs, e.Name()))
		if err != nil {
			problems = append(problems, Problem{File: file, Message: "could not read: " + err.Error()})
			continue
		}
		title, line, msg := itemTitle(stripBOM(data))
		if msg != "" {
			problems = append(problems, Problem{File: file, Line: line, Message: msg})
			continue
		}
		items = append(items, Item{ID: id, Title: title, Path: file})
	}
	return items, problems
}

// itemTitle returns the title from the first non-blank line, which must be a
// level-1 heading. On failure it returns the line to report and a message.
func itemTitle(data []byte) (title string, line int, problem string) {
	if strings.TrimSpace(string(data)) == "" {
		return "", 1, `file is empty; start with a "# <title>" line`
	}
	for i, raw := range strings.Split(string(data), "\n") {
		t := strings.TrimSpace(raw)
		if t == "" {
			continue
		}
		if strings.HasPrefix(t, "# ") && strings.TrimSpace(t[2:]) != "" {
			return strings.TrimSpace(t[2:]), 0, ""
		}
		return "", i + 1, `first line must be a "# <title>" heading`
	}
	return "", 1, `file is empty; start with a "# <title>" line`
}
