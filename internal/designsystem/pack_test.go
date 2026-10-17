package designsystem

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const (
	packRoot    = "../../packs/stripe-reference"
	packFixture = "../../testdata/critique/pack-clean.html"
)

// loadPack loads the reference pack and fails the test when it has any problem.
func loadPack(t *testing.T) DesignSystem {
	t.Helper()
	pack, problems := Load(packRoot)
	if len(problems) != 0 {
		t.Fatalf("pack has problems: %v", problems)
	}
	return pack
}

// readFixture returns the text of the reference pack's fixture page.
func readFixture(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(packFixture)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestPackLoadsClean(t *testing.T) {
	pack := loadPack(t)
	if len(pack.Templates) != 1 || len(pack.Flows) != 1 || len(pack.Components) != 3 {
		t.Errorf("got %d templates, %d flows, %d components; want 1, 1, 3",
			len(pack.Templates), len(pack.Flows), len(pack.Components))
	}
	if len(pack.Rules) != 6 {
		t.Errorf("got %d rules, want 6", len(pack.Rules))
	}
	if got := Checkable(pack.Rules); got != 4 {
		t.Errorf("Checkable = %d, want 4", got)
	}
}

func TestPackShipsNoBrandAssets(t *testing.T) {
	err := filepath.WalkDir(packRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		switch filepath.Ext(path) {
		case ".md", ".json", ".html":
		default:
			t.Errorf("pack ships a non-text file: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPackReadmeStatesUnofficial(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(packRoot, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ToLower(string(data))
	for _, want := range []string{"unofficial", "not endorsed by stripe", "docs/research/stripe-design-system.md"} {
		if !strings.Contains(text, want) {
			t.Errorf("README.md does not contain %q", want)
		}
	}
}

func TestPackFixtureHasNoFindings(t *testing.T) {
	pack := loadPack(t)
	if got := Critique(readFixture(t), pack.Rules); len(got) != 0 {
		t.Errorf("fixture has findings: %+v", got)
	}
}

func TestPackSeededViolationIsFound(t *testing.T) {
	pack := loadPack(t)
	source := strings.Replace(readFixture(t), "</style>", ".seed { color: #ff0000; }\n</style>", 1)
	got := Critique(source, pack.Rules)
	if len(got) != 1 || got[0].RuleID != "color.tokens-only" || got[0].Match != "#ff0000" {
		t.Errorf("got %+v, want one color.tokens-only finding for #ff0000", got)
	}
}

func TestPackPatternsOnSnippets(t *testing.T) {
	pack := loadPack(t)
	tests := []struct {
		name   string
		source string
		want   []string // rule ID of each expected finding, in order
	}{
		{
			name:   "unitless zero margin is not a pixel value",
			source: "margin: 0;\n",
		},
		{
			name:   "a cubic-bezier curve in a transition is allowed",
			source: "transition: color 120ms cubic-bezier(0.25, 1, 0.5, 1);\n",
		},
		{
			name:   "uppercase hex is still a raw colour",
			source: "color: #FFFFFF;\n",
			want:   []string{"color.tokens-only"},
		},
		{
			name:   "a raw colour fallback inside var() is a raw colour",
			source: "color: var(--color-ink, rgb(0, 0, 0));\n",
			want:   []string{"color.tokens-only"},
		},
		{
			name:   "a pixel value on the next line is not matched (per-line limit)",
			source: "margin:\n13px\n",
		},
		{
			name:   "ease-in-out is flagged as a built-in keyword (documented)",
			source: "transition: color 1s ease-in-out;\n",
			want:   []string{"motion.no-builtin-easing"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ids []string
			for _, f := range Critique(tt.source, pack.Rules) {
				ids = append(ids, f.RuleID)
			}
			if !reflect.DeepEqual(ids, tt.want) {
				t.Errorf("rule IDs = %v, want %v", ids, tt.want)
			}
		})
	}
}
