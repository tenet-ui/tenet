package server

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tenet-ui/tenet/internal/designsystem"
)

func TestDeliveryText(t *testing.T) {
	rule := designsystem.Rule{ID: "a", Statement: "S.", Reason: "R.", Valid: true}
	rule2 := designsystem.Rule{ID: "b", Statement: "S2.", Reason: "R2."}
	tests := []struct {
		name  string
		lead  string
		rules []designsystem.Rule
		note  bool
		want  string
	}{
		{
			name: "lead only",
			lead: "x",
			want: "x\n",
		},
		{
			name: "trailing newlines in lead are trimmed",
			lead: "x\n\n",
			want: "x\n",
		},
		{
			name:  "one rule",
			lead:  "x",
			rules: []designsystem.Rule{rule},
			want:  "x\n\n## Linked rules\n\n### a\nS.\nReason: R.\n",
		},
		{
			name:  "two rules keep their order",
			lead:  "x",
			rules: []designsystem.Rule{rule, rule2},
			want:  "x\n\n## Linked rules\n\n### a\nS.\nReason: R.\n\n### b\nS2.\nReason: R2.\n",
		},
		{
			name: "rules note only",
			lead: "x",
			note: true,
			want: "x\n\n" + rulesNoteText + "\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deliveryText(tt.lead, tt.rules, tt.note); got != tt.want {
				t.Errorf("deliveryText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNotFoundText(t *testing.T) {
	tests := []struct {
		name      string
		k         deliveryTarget
		available []string
		want      string
	}{
		{
			name:      "lists available components",
			k:         componentKind,
			available: []string{"button", "card"},
			want:      `component "x" not found. Available components: button, card.`,
		},
		{
			name: "no components yet",
			k:    componentKind,
			want: `component "x" not found. The design system has no components yet; add Markdown files under design-system/components/.`,
		},
		{
			name:      "token categories",
			k:         tokenKind,
			available: []string{"color", "space"},
			want:      `token category "x" not found. Available token categories: color, space.`,
		},
		{
			name: "no tokens yet",
			k:    tokenKind,
			want: `token category "x" not found. The design system has no tokens yet; add categories to design-system/tokens.json.`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := notFoundText(tt.k, "x", tt.available); got != tt.want {
				t.Errorf("notFoundText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInvalidText(t *testing.T) {
	p := designsystem.Problem{File: "design-system/components/bad.md", Line: 1, Message: `first line must be a "# <title>" heading`}
	want := `component "bad" is not valid: design-system/components/bad.md:1: first line must be a "# <title>" heading. Run tenet validate.`
	if got := invalidText(componentKind, "bad", p); got != want {
		t.Errorf("invalidText() = %q, want %q", got, want)
	}
}

// writeDesignSystem writes files under root, creating folders as needed.
func writeDesignSystem(t *testing.T, root string, files map[string]string) {
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

// callTool calls a tool on the session and returns the whole *mcp.CallToolResult, not just its text.
func callTool(t *testing.T, session *mcp.ClientSession, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("call %s: %v", name, err)
	}
	return res
}

func TestGetComponentFixtureExact(t *testing.T) {
	session := connect(t, fixtureRoot)
	res := callTool(t, session, "get_component", map[string]any{"id": "button"})
	if res.IsError {
		t.Fatalf("IsError, text: %s", textOf(t, res))
	}
	want := "# Button\n\nThe primary action on a surface. Use one per view.\n\n## Linked rules\n\n### color.tokens-only\nComponents take colours from tokens, never from raw hex values.\nReason: A raw value drifts from the palette the moment one screen is edited alone.\n"
	if got := textOf(t, res); got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestGetFlowFixtureHasNoLinkedRules(t *testing.T) {
	session := connect(t, fixtureRoot)
	res := callTool(t, session, "get_flow", map[string]any{"id": "first-run"})
	if res.IsError {
		t.Fatalf("IsError, text: %s", textOf(t, res))
	}
	if got := textOf(t, res); strings.Contains(got, "Linked rules") || !strings.HasPrefix(got, "# First run\n") {
		t.Errorf("got %q, want the flow body and no linked rules", got)
	}
}

func TestGetTokensFixtureExact(t *testing.T) {
	session := connect(t, fixtureRoot)
	res := callTool(t, session, "get_tokens", map[string]any{"category": "color"})
	if res.IsError {
		t.Fatalf("IsError, text: %s", textOf(t, res))
	}
	want := "color.ink: #1d2433\ncolor.paper: #fbfaf7\ncolor.accent: #2f6fed\n\n## Linked rules\n\n### color.tokens-only\nComponents take colours from tokens, never from raw hex values.\nReason: A raw value drifts from the palette the moment one screen is edited alone.\n"
	if got := textOf(t, res); got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestGetTokensNumberValue(t *testing.T) {
	session := connect(t, fixtureRoot)
	res := callTool(t, session, "get_tokens", map[string]any{"category": "space"})
	if res.IsError {
		t.Fatalf("IsError, text: %s", textOf(t, res))
	}
	if got := textOf(t, res); !strings.HasPrefix(got, "space.unit: 4\n") {
		t.Errorf("got %q, want it to start with space.unit: 4", got)
	}
}

func TestNotFoundListsAvailable(t *testing.T) {
	session := connect(t, fixtureRoot)
	tests := []struct {
		tool string
		args map[string]any
		want string
	}{
		{"get_component", map[string]any{"id": "card"}, `component "card" not found. Available components: button.`},
		{"get_template", map[string]any{"id": "missing"}, `template "missing" not found. Available templates: settings-page.`},
		{"get_flow", map[string]any{"id": "x"}, `flow "x" not found. Available flows: first-run.`},
		{"get_tokens", map[string]any{"category": "motion"}, `token category "motion" not found. Available token categories: color, space.`},
	}
	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			res := callTool(t, session, tt.tool, tt.args)
			if !res.IsError {
				t.Fatal("want IsError")
			}
			if got := textOf(t, res); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEmptyIDListsAvailable(t *testing.T) {
	session := connect(t, fixtureRoot)
	res := callTool(t, session, "get_component", map[string]any{"id": ""})
	if !res.IsError {
		t.Fatal("want IsError")
	}
	if got := textOf(t, res); !strings.HasPrefix(got, `component "" not found. Available components: button.`) {
		t.Errorf("got %q", got)
	}
}

func TestNoItemsOfKindSaysSo(t *testing.T) {
	root := t.TempDir()
	writeDesignSystem(t, root, map[string]string{
		"design-system/point-of-view.md": "# Point of view\n\nText.\n",
		"design-system/tokens.json":      "{}",
		"design-system/rules.md":         "# Rules\n",
	})
	session := connect(t, root)
	res := callTool(t, session, "get_component", map[string]any{"id": "button"})
	want := `component "button" not found. The design system has no components yet; add Markdown files under design-system/components/.`
	if !res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want %q", res.IsError, textOf(t, res), want)
	}
}

func TestInvalidItemReportsProblem(t *testing.T) {
	root := t.TempDir()
	writeDesignSystem(t, root, map[string]string{
		"design-system/point-of-view.md":  "# Point of view\n\nText.\n",
		"design-system/tokens.json":       "{}",
		"design-system/rules.md":          "# Rules\n",
		"design-system/components/bad.md": "Bad\n",
	})
	session := connect(t, root)
	res := callTool(t, session, "get_component", map[string]any{"id": "bad"})
	want := `component "bad" is not valid: design-system/components/bad.md:1: first line must be a "# <title>" heading. Run tenet validate.`
	if !res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want %q", res.IsError, textOf(t, res), want)
	}
}

func TestMissingDesignSystemIsToolError(t *testing.T) {
	session := connect(t, t.TempDir())
	for _, tt := range []struct {
		tool string
		args map[string]any
	}{
		{"get_component", map[string]any{"id": "button"}},
		{"get_template", map[string]any{"id": "x"}},
		{"get_flow", map[string]any{"id": "x"}},
		{"get_tokens", map[string]any{"category": "color"}},
	} {
		res := callTool(t, session, tt.tool, tt.args)
		if !res.IsError || !strings.Contains(textOf(t, res), "run tenet init") {
			t.Errorf("%s: got IsError=%t %q, want run tenet init", tt.tool, res.IsError, textOf(t, res))
		}
	}
}

func TestDesignSystemFileIsToolError(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "design-system"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	session := connect(t, root)
	for _, tt := range []struct {
		tool string
		args map[string]any
	}{
		{"get_component", map[string]any{"id": "button"}},
		{"get_template", map[string]any{"id": "x"}},
		{"get_flow", map[string]any{"id": "x"}},
		{"get_tokens", map[string]any{"category": "color"}},
	} {
		res := callTool(t, session, tt.tool, tt.args)
		got := textOf(t, res)
		if !res.IsError || !strings.Contains(got, "not a directory") || !strings.Contains(got, "run tenet init") {
			t.Errorf("%s: got IsError=%t %q, want not a directory and run tenet init", tt.tool, res.IsError, got)
		}
	}
}

func TestEditsVisibleWithoutRestart(t *testing.T) {
	root := t.TempDir()
	writeDesignSystem(t, root, map[string]string{
		"design-system/point-of-view.md":     "# Point of view\n\nText.\n",
		"design-system/tokens.json":          "{}",
		"design-system/rules.md":             "# Rules\n",
		"design-system/components/button.md": "# Button\n\nOld.\n",
	})
	session := connect(t, root)
	if got := textOf(t, callTool(t, session, "get_component", map[string]any{"id": "button"})); !strings.Contains(got, "Old.") {
		t.Fatalf("first call got %q", got)
	}
	writeDesignSystem(t, root, map[string]string{"design-system/components/button.md": "# Button\n\nNew.\n"})
	if got := textOf(t, callTool(t, session, "get_component", map[string]any{"id": "button"})); !strings.Contains(got, "New.") {
		t.Errorf("second call got %q, want the edited body", got)
	}
}

func TestRulesProblemNote(t *testing.T) {
	root := t.TempDir()
	writeDesignSystem(t, root, map[string]string{
		"design-system/point-of-view.md":     "# Point of view\n\nText.\n",
		"design-system/tokens.json":          "{}",
		"design-system/rules.md":             "## a.b\nStatement.\nApplies to: component:button\n",
		"design-system/components/button.md": "# Button\n\nBody.\n",
	})
	session := connect(t, root)
	res := callTool(t, session, "get_component", map[string]any{"id": "button"})
	want := "# Button\n\nBody.\n\n" + rulesNoteText + "\n"
	if res.IsError || textOf(t, res) != want {
		t.Errorf("got IsError=%t %q, want %q", res.IsError, textOf(t, res), want)
	}
}

func TestResultsUnder4KB(t *testing.T) {
	session := connect(t, fixtureRoot)
	for _, tt := range []struct {
		tool string
		args map[string]any
	}{
		{"get_component", map[string]any{"id": "button"}},
		{"get_template", map[string]any{"id": "settings-page"}},
		{"get_flow", map[string]any{"id": "first-run"}},
		{"get_tokens", map[string]any{"category": "color"}},
	} {
		res := callTool(t, session, tt.tool, tt.args)
		if n := len(textOf(t, res)); n >= 4096 {
			t.Errorf("%s returned %d bytes, want under 4096", tt.tool, n)
		}
	}
}

func TestDeliveryToolSchemas(t *testing.T) {
	session := connect(t, fixtureRoot)
	res, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	wantRequired := map[string]string{"get_component": "id", "get_template": "id", "get_flow": "id", "get_tokens": "category"}
	for _, tool := range res.Tools {
		field, ok := wantRequired[tool.Name]
		if !ok {
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
		if !slices.Contains(schema.Required, field) {
			t.Errorf("%s schema required = %v, want %q (schema: %s)", tool.Name, schema.Required, field, raw)
		}
	}
}
