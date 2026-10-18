package server

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// critiqueRoot is the testdata folder: it holds critique/page.html and project/design-system.
const critiqueRoot = "../../testdata"

// wantFixtureText is the exact result for the fixture against the bundled example.
const wantFixtureText = `2 finding(s) in critique/page.html against project.

critique/page.html:4: color.tokens-only: found "#ff0000". Replace the raw colour with a colour token from tokens.json.
critique/page.html:5: spacing.unit-multiples: found "margin: 13px". Replace the raw pixel value with a space token from tokens.json, so spacing stays a multiple of the space unit.
`

func TestCritiqueFixtureFindings(t *testing.T) {
	session := connect(t, critiqueRoot)
	res := callTool(t, session, "critique", map[string]any{"file": "critique/page.html", "system": "project"})
	if res.IsError || textOf(t, res) != wantFixtureText {
		t.Errorf("got IsError=%t\n%s\nwant\n%s", res.IsError, textOf(t, res), wantFixtureText)
	}
}

func TestCritiqueNoFindings(t *testing.T) {
	root := t.TempDir()
	writeDesignSystem(t, root, map[string]string{
		"design-system/point-of-view.md": "# Point of view\n\nText.\n",
		"design-system/tokens.json":      "{}",
		"design-system/rules.md":         "## r\nS.\nReason: R.\nForbid: x*\nFix: F.\n",
		"ui/page.html":                   "clean\n",
	})
	session := connect(t, root)
	res := callTool(t, session, "critique", map[string]any{"file": "ui/page.html", "system": "."})
	want := "No findings in ui/page.html against .. Checked 1 rule(s).\n"
	if res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want %q", res.IsError, textOf(t, res), want)
	}
}

func TestCritiqueBlankSystemIsDot(t *testing.T) {
	root := t.TempDir()
	writeDesignSystem(t, root, map[string]string{
		"design-system/point-of-view.md": "# Point of view\n\nText.\n",
		"design-system/tokens.json":      "{}",
		"design-system/rules.md":         "## r\nS.\nReason: R.\nForbid: x*\nFix: F.\n",
		"ui/page.html":                   "clean\n",
	})
	session := connect(t, root)
	dot := callTool(t, session, "critique", map[string]any{"file": "ui/page.html", "system": "."})
	blank := callTool(t, session, "critique", map[string]any{"file": "ui/page.html", "system": "  "})
	if blank.IsError || textOf(t, blank) != textOf(t, dot) {
		t.Errorf("blank system: IsError=%t %q, want %q", blank.IsError, textOf(t, blank), textOf(t, dot))
	}
}

func TestCritiqueNoCheckableRules(t *testing.T) {
	root := t.TempDir()
	writeDesignSystem(t, root, map[string]string{
		"design-system/point-of-view.md": "# Point of view\n\nText.\n",
		"design-system/tokens.json":      "{}",
		"design-system/rules.md":         "## r\nS.\nReason: R.\n",
		"ui/page.html":                   "x\n",
	})
	session := connect(t, root)
	res := callTool(t, session, "critique", map[string]any{"file": "ui/page.html", "system": "."})
	want := "No rule in . has a Forbid: line, so nothing was checked. Add one under a rule in design-system/rules.md.\n"
	if res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want %q", res.IsError, textOf(t, res), want)
	}
}

func TestCritiqueRulesProblemNote(t *testing.T) {
	root := t.TempDir()
	writeDesignSystem(t, root, map[string]string{
		"design-system/point-of-view.md": "# Point of view\n\nText.\n",
		"design-system/tokens.json":      "{}",
		"design-system/rules.md":         "## r\nS.\nReason: R.\nForbid: (\nFix: F.\n",
		"ui/page.html":                   "x\n",
	})
	session := connect(t, root)
	res := callTool(t, session, "critique", map[string]any{"file": "ui/page.html", "system": "."})
	want := "No rule in . has a Forbid: line, so nothing was checked. Add one under a rule in design-system/rules.md.\n\n" +
		"Note: design-system/rules.md has problems, so some rules may not be checked. Run tenet validate.\n"
	if res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want %q", res.IsError, textOf(t, res), want)
	}
}

func TestCritiqueCRLFAndBOM(t *testing.T) {
	root := t.TempDir()
	writeDesignSystem(t, root, map[string]string{
		"design-system/point-of-view.md": "# Point of view\n\nText.\n",
		"design-system/tokens.json":      "{}",
		"design-system/rules.md":         "## r\nS.\nReason: R.\nForbid: ^a$\nFix: F.\n",
		"ui/page.html":                   "\uFEFFa\r\nb\r\n",
	})
	session := connect(t, root)
	res := callTool(t, session, "critique", map[string]any{"file": "ui/page.html", "system": "."})
	want := "1 finding(s) in ui/page.html against ..\n\nui/page.html:1: r: found \"a\". F.\n"
	if res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want %q", res.IsError, textOf(t, res), want)
	}
}

func TestCritiqueMissingFileIsError(t *testing.T) {
	session := connect(t, critiqueRoot)
	res := callTool(t, session, "critique", map[string]any{"file": "nope.html", "system": "project"})
	abs, err := filepath.Abs(critiqueRoot)
	if err != nil {
		t.Fatal(err)
	}
	want := `file "nope.html" not found. Relative paths resolve against the server root ` + abs + `; pass an absolute path for a file outside it.`
	if !res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want error %q", res.IsError, textOf(t, res), want)
	}
}

func TestCritiqueDirectoryIsError(t *testing.T) {
	session := connect(t, critiqueRoot)
	res := callTool(t, session, "critique", map[string]any{"file": "critique", "system": "project"})
	want := `file "critique" is a directory; pass a file.`
	if !res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want error %q", res.IsError, textOf(t, res), want)
	}
}

func TestCritiqueBinaryFileIsError(t *testing.T) {
	root := t.TempDir()
	writeDesignSystem(t, root, map[string]string{
		"design-system/point-of-view.md": "# Point of view\n\nText.\n",
		"design-system/tokens.json":      "{}",
		"design-system/rules.md":         "## r\nS.\nReason: R.\nForbid: a\nFix: F.\n",
	})
	if err := os.WriteFile(filepath.Join(root, "bin.dat"), []byte("a\x00a"), 0o644); err != nil {
		t.Fatal(err)
	}
	session := connect(t, root)
	res := callTool(t, session, "critique", map[string]any{"file": "bin.dat", "system": "."})
	want := `file "bin.dat" is not a text file.`
	if !res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want error %q", res.IsError, textOf(t, res), want)
	}
}

func TestCritiqueTooLargeIsError(t *testing.T) {
	root := t.TempDir()
	writeDesignSystem(t, root, map[string]string{
		"design-system/point-of-view.md": "# Point of view\n\nText.\n",
		"design-system/tokens.json":      "{}",
		"design-system/rules.md":         "## r\nS.\nReason: R.\nForbid: a\nFix: F.\n",
	})
	big := strings.Repeat("a", 1<<20+1)
	if err := os.WriteFile(filepath.Join(root, "big.txt"), []byte(big), 0o644); err != nil {
		t.Fatal(err)
	}
	session := connect(t, root)
	res := callTool(t, session, "critique", map[string]any{"file": "big.txt", "system": "."})
	want := `file "big.txt" is larger than 1 MiB; pass a smaller source file.`
	if !res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want error %q", res.IsError, textOf(t, res), want)
	}
}

func TestCritiqueNoDesignSystemIsError(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "page.html"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	session := connect(t, root)
	res := callTool(t, session, "critique", map[string]any{"file": "page.html", "system": "."})
	want := "design-system not found under " + root + ". This server was started with --dir " + root + ", which has no design-system/ folder; start it with a --dir that does."
	if !res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want error %q", res.IsError, textOf(t, res), want)
	}
	if strings.Contains(textOf(t, res), "tenet init") {
		t.Errorf("message mentions tenet init: %q", textOf(t, res))
	}
}

func TestCritiqueAbsoluteSystemWithoutDesignSystem(t *testing.T) {
	r := t.TempDir()
	other := t.TempDir()
	if err := os.WriteFile(filepath.Join(r, "page.html"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	session := connect(t, r)
	res := callTool(t, session, "critique", map[string]any{"file": "page.html", "system": other})
	want := "design-system not found under " + other + ". This server's root is " + r + "; omit system to check against the system it serves."
	if !res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want error %q", res.IsError, textOf(t, res), want)
	}
}

func TestCritiqueFileNotFoundNamesRoot(t *testing.T) {
	root := t.TempDir()
	session := connect(t, root)
	res := callTool(t, session, "critique", map[string]any{"file": "missing.html"})
	want := `file "missing.html" not found. Relative paths resolve against the server root ` + root + `; pass an absolute path for a file outside it.`
	if !res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want error %q", res.IsError, textOf(t, res), want)
	}
}

func TestCritiqueRelativeRootIsShownAbsolute(t *testing.T) {
	abs, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	session := connect(t, "../..")
	res := callTool(t, session, "critique", map[string]any{"file": "missing.html"})
	text := textOf(t, res)
	if !res.IsError || !strings.Contains(text, "server root "+abs+";") || strings.Contains(text, "root ../..") {
		t.Errorf("got IsError=%t %q, want absolute root %q", res.IsError, text, abs)
	}
}

func TestCritiqueOmittedSystemDefaultsToRoot(t *testing.T) {
	root := t.TempDir()
	writeDesignSystem(t, root, map[string]string{
		"design-system/point-of-view.md": "# Point of view\n\nText.\n",
		"design-system/tokens.json":      "{}",
		"design-system/rules.md":         "## r\nS.\nReason: R.\nForbid: bad\nFix: F.\n",
		"ui/page.html":                   "a\nbad\n",
	})
	session := connect(t, root)
	base := callTool(t, session, "critique", map[string]any{"file": "ui/page.html", "system": "."})
	if base.IsError || !strings.Contains(textOf(t, base), "1 finding(s)") {
		t.Fatalf("baseline: IsError=%t %q", base.IsError, textOf(t, base))
	}
	for name, args := range map[string]map[string]any{
		"omitted": {"file": "ui/page.html"},
		"empty":   {"file": "ui/page.html", "system": ""},
	} {
		t.Run(name, func(t *testing.T) {
			res := callTool(t, session, "critique", args)
			if res.IsError || textOf(t, res) != textOf(t, base) {
				t.Errorf("got IsError=%t %q, want %q", res.IsError, textOf(t, res), textOf(t, base))
			}
		})
	}
}

func TestCritiqueSystemIsNotRequired(t *testing.T) {
	session := connect(t, critiqueRoot)
	list, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range list.Tools {
		if tool.Name != "critique" {
			continue
		}
		raw, err := json.Marshal(tool.InputSchema)
		if err != nil {
			t.Fatal(err)
		}
		var schema struct {
			Required []string `json:"required"`
		}
		if err := json.Unmarshal(raw, &schema); err != nil {
			t.Fatal(err)
		}
		var hasFile bool
		for _, r := range schema.Required {
			if r == "system" {
				t.Errorf("system is required: %v", schema.Required)
			}
			hasFile = hasFile || r == "file"
		}
		if !hasFile {
			t.Errorf("file is not required: %v", schema.Required)
		}
		return
	}
	t.Fatal("critique tool not listed")
}

func TestCritiqueNonRegularFileIsError(t *testing.T) {
	root := t.TempDir()
	if err := syscall.Mkfifo(filepath.Join(root, "pipe"), 0o644); err != nil {
		t.Skipf("mkfifo unavailable: %v", err)
	}
	session := connect(t, root)
	res := callTool(t, session, "critique", map[string]any{"file": "pipe", "system": "."})
	want := `file "pipe" is not a regular file; pass a file.`
	if !res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want error %q", res.IsError, textOf(t, res), want)
	}
}

func TestCritiqueEmptyArgumentsAreErrors(t *testing.T) {
	session := connect(t, critiqueRoot)
	tests := []struct {
		name string
		args map[string]any
		want string
	}{
		{"empty file", map[string]any{"file": "", "system": "project"}, "file is empty; pass the path to a UI source file."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := callTool(t, session, "critique", tt.args)
			if !res.IsError || textOf(t, res) != tt.want {
				t.Errorf("got IsError=%t %q, want error %q", res.IsError, textOf(t, res), tt.want)
			}
		})
	}
}

// TestCritiquePackFixtureNoFindings runs the reference pack against its fixture from the repository root.
// The pack lives outside testdata, so the server root is the repository root.
func TestCritiquePackFixtureNoFindings(t *testing.T) {
	session := connect(t, "../..")
	res := callTool(t, session, "critique", map[string]any{
		"file":   "testdata/critique/pack-clean.html",
		"system": "packs/stripe-reference",
	})
	want := "No findings in testdata/critique/pack-clean.html against packs/stripe-reference. Checked 4 rule(s).\n"
	if res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want %q", res.IsError, textOf(t, res), want)
	}
}
