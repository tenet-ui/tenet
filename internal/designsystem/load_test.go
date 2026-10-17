package designsystem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTree creates each file under root, making parent directories as needed.
func writeTree(t *testing.T, root string, files map[string]string) {
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

// wantProblem describes one expected problem: its line (0 for a file-level problem) and a substring of its message.
type wantProblem struct {
	Line     int
	Contains string
}

// checkProblems fails the test unless got holds exactly the wanted problems, in order.
func checkProblems(t *testing.T, got []Problem, want []wantProblem) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d problems %v, want %d", len(got), got, len(want))
	}
	for i, w := range want {
		if got[i].Line != w.Line || !strings.Contains(got[i].Message, w.Contains) {
			t.Errorf("problem %d = %v, want line %d with message containing %q", i, got[i], w.Line, w.Contains)
		}
	}
}

// problemsFor returns the problems whose File is file.
func problemsFor(ps []Problem, file string) []Problem {
	var out []Problem
	for _, p := range ps {
		if p.File == file {
			out = append(out, p)
		}
	}
	return out
}

func TestLoadDesignSystemMissing(t *testing.T) {
	_, problems := Load(t.TempDir())
	checkProblems(t, problems, []wantProblem{{0, "directory not found"}})
}

func TestLoadDesignSystemIsFile(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{"design-system": "x"})
	_, problems := Load(root)
	checkProblems(t, problems, []wantProblem{{0, "not a directory"}})
}

func TestLoadPointOfView(t *testing.T) {
	tests := []struct {
		name    string
		files   map[string]string
		wantPOV bool
		want    []wantProblem
	}{
		{
			name:    "valid",
			files:   map[string]string{"design-system/point-of-view.md": "# Brand\n"},
			wantPOV: true,
		},
		{
			name:  "file missing",
			files: map[string]string{"design-system/keep.txt": ""},
			want:  []wantProblem{{0, "file not found"}},
		},
		{
			name:  "whitespace only",
			files: map[string]string{"design-system/point-of-view.md": " \n\t\n"},
			want:  []wantProblem{{1, "file is empty"}},
		},
		{
			name:  "point of view is a directory",
			files: map[string]string{"design-system/point-of-view.md/keep.txt": "x"},
			want:  []wantProblem{{0, "could not read"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			writeTree(t, root, tt.files)
			pack, problems := Load(root)
			checkProblems(t, problemsFor(problems, PointOfViewPath), tt.want)
			if (pack.PointOfView != "") != tt.wantPOV {
				t.Errorf("PointOfView = %q, want non-empty=%v", pack.PointOfView, tt.wantPOV)
			}
		})
	}
}

func TestLoadRequiredFilesMissing(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{"design-system/keep.txt": ""})
	_, problems := Load(root)
	checkProblems(t, problems, []wantProblem{
		{0, "file not found"}, // point of view
		{0, "file not found"}, // tokens
		{0, "file not found"}, // rules
	})
}

func TestLoadCountsEveryKind(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"design-system/point-of-view.md":     "# Brand\n",
		"design-system/tokens.json":          `{"color":{"ink":"#fff"}}`,
		"design-system/rules.md":             "## a\nS.\nReason: r\n",
		"design-system/components/button.md": "# Button\n",
		"design-system/templates/page.md":    "# Page\n",
		"design-system/flows/first-run.md":   "# First run\n",
	})
	pack, problems := Load(root)
	checkProblems(t, problems, nil)
	if pack.PointOfView == "" || len(pack.Tokens) != 1 || len(pack.Rules) != 1 ||
		len(pack.Components) != 1 || len(pack.Templates) != 1 || len(pack.Flows) != 1 {
		t.Fatalf("pack = %+v, want one of each kind", pack)
	}
}

func TestLoadReportsEveryBadFileAndKeepsTheRest(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"design-system/point-of-view.md":     "# Brand\n",
		"design-system/tokens.json":          "{\n  \"color\": {\"ink\": true}\n}\n",
		"design-system/rules.md":             "## Bad\nS.\nReason: r\n## good\nS.\nReason: r\n",
		"design-system/components/Button.md": "# Button\n",
	})
	pack, problems := Load(root)
	checkProblems(t, problems, []wantProblem{
		{2, `token "color.ink" must be a string or number`},
		{1, `rule id "Bad" must be`},
		{0, "file name must be lowercase kebab-case"},
	})
	if len(pack.Rules) != 1 || pack.Rules[0].ID != "good" {
		t.Errorf("rules = %+v, want only good", pack.Rules)
	}
	if pack.PointOfView == "" {
		t.Errorf("point of view was dropped because another file had problems")
	}
}

func TestLoadFixtureIsValid(t *testing.T) {
	pack, problems := Load("../../testdata/project")
	checkProblems(t, problems, nil)
	if len(pack.Tokens) != 4 {
		t.Errorf("tokens = %d, want 4", len(pack.Tokens))
	}
	if len(pack.Rules) != 2 {
		t.Errorf("rules = %d, want 2", len(pack.Rules))
	}
	if len(pack.Components) != 1 || len(pack.Templates) != 1 || len(pack.Flows) != 1 {
		t.Errorf("items = %d components, %d templates, %d flows; want 1 each",
			len(pack.Components), len(pack.Templates), len(pack.Flows))
	}
}
