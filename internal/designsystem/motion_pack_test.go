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
	motionRoot    = "../../packs/motion-reference"
	motionFixture = "../../testdata/critique/motion-clean.tsx"
)

// loadMotionPack loads the motion reference pack and fails the test when it has any problem.
func loadMotionPack(t *testing.T) DesignSystem {
	t.Helper()
	pack, problems := Load(motionRoot)
	if len(problems) != 0 {
		t.Fatalf("motion pack has problems: %v", problems)
	}
	return pack
}

func TestMotionPackLoadsClean(t *testing.T) {
	pack := loadMotionPack(t)
	if len(pack.Templates) != 1 || len(pack.Flows) != 1 || len(pack.Components) != 3 {
		t.Errorf("got %d templates, %d flows, %d components; want 1, 1, 3",
			len(pack.Templates), len(pack.Flows), len(pack.Components))
	}
	if len(pack.Rules) != 9 {
		t.Errorf("got %d rules, want 9", len(pack.Rules))
	}
	if got := Checkable(pack.Rules); got != 5 {
		t.Errorf("Checkable = %d, want 5", got)
	}
}

func TestMotionPackShipsOnlyText(t *testing.T) {
	err := filepath.WalkDir(motionRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		switch filepath.Ext(path) {
		case ".md", ".json":
		default:
			t.Errorf("pack ships a non-text file: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMotionReadmeStatesInvented(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(motionRoot, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ToLower(string(data))
	for _, want := range []string{"invented", "no outside source", "remotion", "css"} {
		if !strings.Contains(text, want) {
			t.Errorf("README.md does not contain %q", want)
		}
	}
}

// readMotionFixture returns the text of the motion pack's fixture composition.
func readMotionFixture(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(motionFixture)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestMotionFixtureHasNoFindings(t *testing.T) {
	pack := loadMotionPack(t)
	if got := Critique(readMotionFixture(t), pack.Rules); len(got) != 0 {
		t.Errorf("fixture has findings: %+v", got)
	}
}

func TestMotionSeededViolations(t *testing.T) {
	pack := loadMotionPack(t)
	tests := []struct {
		name string
		edit string
		want string
	}{
		{"raw css time", "transition: opacity 300ms var(--motion-settle);", "motion.time-from-tokens"},
		{"css keyword easing", "transition: opacity var(--motion-duration-base) linear;", "motion.no-builtin-easing"},
		{"literal bezier", "Easing.bezier(0.16, 1, 0.3, 1)", "motion.no-builtin-easing"},
		{"low damping", "spring({ config: { damping: 8 } })", "motion.springs-settle"},
		{"raw frame count", "<Sequence from={0} durationInFrames={90}>", "motion.time-from-tokens"},
		{"raw delay", "const delay = 12;", "motion.stagger-from-tokens"},
		{"raw colour", "color: #ff0000;", "color.tokens-only"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := readMotionFixture(t) + "\n" + tt.edit + "\n"
			got := Critique(source, pack.Rules)
			if len(got) != 1 || got[0].RuleID != tt.want {
				t.Errorf("got %+v, want one %s finding", got, tt.want)
			}
		})
	}
}

func TestMotionPatternsOnSnippets(t *testing.T) {
	pack := loadMotionPack(t)
	tests := []struct {
		name   string
		source string
		want   []string // rule ID of each expected finding, in order
	}{
		{"release is not ease", "const release = 1; // release\n", nil},
		{"increase is not ease", "const increase = value;\n", nil},
		{"fps from a token expression", "<Composition fps={fps} durationInFrames={frames} />\n", nil},
		{"spread pack curve", "const easing = Easing.bezier(...settle);\n", nil},
		{"custom property declaration", "--motion-settle: cubic-bezier(0.16, 1, 0.3, 1);\n", nil},
		{"ease keyword in a transition", "transition: opacity var(--motion-duration-base) ease;\n", []string{"motion.no-builtin-easing"}},
		{"Easing.quad", "Easing.quad\n", []string{"motion.no-builtin-easing"}},
		{"literal bezier", "Easing.bezier(0.16, 1, 0.3, 1)\n", []string{"motion.no-builtin-easing"}},
		{"jsx quoted transition", "style={{ transition: \"opacity 300ms linear\" }}\n", []string{"motion.no-builtin-easing", "motion.time-from-tokens"}},
		{"jsx quoted animation", "style={{ animation: \"pulse 2s ease-in-out infinite\" }}\n", []string{"motion.no-builtin-easing", "motion.time-from-tokens"}},
		{"Easing.poly", "easing: Easing.in(Easing.poly(3))\n", []string{"motion.no-builtin-easing"}},
		{"high damping", "spring({ config: { damping: 28 } })\n", nil},
		{"overshoot allowed", "overshootClamping: false\n", []string{"motion.springs-settle"}},
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

func TestMotionPackDoesNotTriggerItself(t *testing.T) {
	pack := loadMotionPack(t)
	err := filepath.WalkDir(motionRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Base(path) == "rules.md" {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if got := Critique(string(data), pack.Rules); len(got) != 0 {
			t.Errorf("%s triggers the pack's own rules: %+v", path, got)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
