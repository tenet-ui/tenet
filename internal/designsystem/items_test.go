package designsystem

import (
	"fmt"
	"testing"
)

// formatItems renders each item as "id|title|path" for comparison.
func formatItems(items []Item) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, fmt.Sprintf("%s|%s|%s", it.ID, it.Title, it.Path))
	}
	return out
}

func TestLoadItems(t *testing.T) {
	tests := []struct {
		name         string
		files        map[string]string
		wantItems    []string
		wantProblems []wantProblem
	}{
		{
			name: "two items sorted by file name",
			files: map[string]string{
				"design-system/components/card.md":   "# Card\n\nA surface.\n",
				"design-system/components/button.md": "# Button\n\nAn action.\n",
			},
			wantItems: []string{
				"button|Button|design-system/components/button.md",
				"card|Card|design-system/components/card.md",
			},
		},
		{
			name:      "missing folder has no items and no problems",
			files:     map[string]string{"design-system/keep.txt": ""},
			wantItems: nil,
		},
		{
			name:         "folder is a file",
			files:        map[string]string{"design-system/components": "x"},
			wantProblems: []wantProblem{{0, "not a directory"}},
		},
		{
			name: "non-md files and subdirectories are ignored",
			files: map[string]string{
				"design-system/components/.gitkeep":    "",
				"design-system/components/notes.txt":   "x",
				"design-system/components/nested/x.md": "# Nested\n",
				"design-system/components/button.md":   "# Button\n",
			},
			wantItems: []string{"button|Button|design-system/components/button.md"},
		},
		{
			name: "uppercase file name is reported",
			files: map[string]string{
				"design-system/components/Button.md": "# Button\n",
			},
			wantProblems: []wantProblem{{0, "file name must be lowercase kebab-case"}},
		},
		{
			name: "empty file",
			files: map[string]string{
				"design-system/components/button.md": " \n\n",
			},
			wantProblems: []wantProblem{{1, "file is empty"}},
		},
		{
			name: "first non-blank line is not a heading",
			files: map[string]string{
				"design-system/components/button.md": "\n\nButton\n# Later\n",
			},
			wantProblems: []wantProblem{{3, `first line must be a "# <title>" heading`}},
		},
		{
			name: "heading with empty title",
			files: map[string]string{
				"design-system/components/button.md": "#  \n",
			},
			wantProblems: []wantProblem{{1, `first line must be a "# <title>" heading`}},
		},
		{
			name: "UTF-8 BOM before the heading is ignored",
			files: map[string]string{
				"design-system/components/button.md": "\ufeff# Button\n",
			},
			wantItems: []string{"button|Button|design-system/components/button.md"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			writeTree(t, root, tt.files)
			items, problems := loadItems(root, ComponentsDir)
			if g, w := fmt.Sprint(formatItems(items)), fmt.Sprint(tt.wantItems); g != w {
				t.Errorf("items = %s, want %s", g, w)
			}
			checkProblems(t, problems, tt.wantProblems)
		})
	}
}

func TestLoadItemsBody(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"design-system/components/button.md": "\xEF\xBB\xBF# Button\n\nAn action.\n",
	})
	items, problems := loadItems(root, ComponentsDir)
	if len(problems) != 0 {
		t.Fatalf("unexpected problems: %v", problems)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0].Title != "Button" {
		t.Errorf("Title = %q, want %q", items[0].Title, "Button")
	}
	if want := "# Button\n\nAn action.\n"; items[0].Body != want {
		t.Errorf("Body = %q, want %q", items[0].Body, want)
	}
}
