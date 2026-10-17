package designsystem

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// formatRules renders each rule as "id@line|statement|reason" for comparison.
func formatRules(rs []Rule) []string {
	out := make([]string, 0, len(rs))
	for _, r := range rs {
		out = append(out, fmt.Sprintf("%s@%d|%s|%s", r.ID, r.Line, r.Statement, r.Reason))
	}
	return out
}

func TestParseRules(t *testing.T) {
	tests := []struct {
		name         string
		data         string
		wantRules    []string
		wantProblems []wantProblem
	}{
		{
			name:      "zero entries is valid",
			data:      "# Rules\n\nIntro text.\n",
			wantRules: nil,
		},
		{
			name:      "empty file is valid",
			data:      "",
			wantRules: nil,
		},
		{
			name: "two valid entries",
			data: joinLines(
				"# Rules",
				"",
				"## color.tokens-only",
				"Components take colours from tokens.",
				"Reason: Raw values drift.",
				"",
				"## spacing.unit-multiples",
				"Spacing is a multiple of the unit.",
				"Reason: Off-grid spacing looks unplanned.",
			),
			wantRules: []string{
				"color.tokens-only@3|Components take colours from tokens.|Raw values drift.",
				"spacing.unit-multiples@7|Spacing is a multiple of the unit.|Off-grid spacing looks unplanned.",
			},
		},
		{
			name: "multi-line statement is joined with a space",
			data: joinLines(
				"## a.b",
				"First line.",
				"Second line.",
				"Reason: because.",
			),
			wantRules: []string{"a.b@1|First line. Second line.|because."},
		},
		{
			name:      "CRLF line endings",
			data:      "## a\r\nS.\r\nReason: r.\r\n",
			wantRules: []string{"a@1|S.|r."},
		},
		{
			name:      "UTF-8 BOM is ignored",
			data:      "\xEF\xBB\xBF## a\nS.\nReason: r\n",
			wantRules: []string{"a@1|S.|r"},
		},
		{
			name: "duplicate id reports second heading and drops it",
			data: joinLines(
				"## a",
				"S.",
				"Reason: r",
				"## a",
				"S.",
				"Reason: r",
			),
			wantRules:    []string{"a@1|S.|r"},
			wantProblems: []wantProblem{{4, `duplicate rule id "a" (first at line 1)`}},
		},
		{
			name: "invalid ids are reported and dropped",
			data: joinLines(
				"## Color",
				"S.",
				"Reason: r",
				"## a..b",
				"S.",
				"Reason: r",
				"## -x",
				"S.",
				"Reason: r",
				"## a b",
				"S.",
				"Reason: r",
				"## ",
				"S.",
				"Reason: r",
			),
			wantProblems: []wantProblem{
				{1, `rule id "Color" must be`},
				{4, `rule id "a..b" must be`},
				{7, `rule id "-x" must be`},
				{10, `rule id "a b" must be`},
				{13, `rule id "" must be`},
			},
		},
		{
			name:         "no statement",
			data:         "## a\nReason: x\n",
			wantRules:    []string{"a@1||x"},
			wantProblems: []wantProblem{{1, `rule "a" has no statement`}},
		},
		{
			name:         "no Reason line",
			data:         "## a\nStatement.\n",
			wantRules:    []string{"a@1|Statement.|"},
			wantProblems: []wantProblem{{1, `rule "a" has no Reason: line`}},
		},
		{
			name:         "empty Reason line",
			data:         "## a\nStatement.\nReason:   \n",
			wantRules:    []string{"a@1|Statement.|"},
			wantProblems: []wantProblem{{3, `rule "a" has an empty Reason: line`}},
		},
		{
			name:         "repeated Reason line keeps the first",
			data:         "## a\nS.\nReason: x\nReason: y\n",
			wantRules:    []string{"a@1|S.|x"},
			wantProblems: []wantProblem{{4, `rule "a" has more than one Reason: line`}},
		},
		{
			name: "heading inside a fenced code block is not a rule",
			data: joinLines(
				"## a",
				"S.",
				"Reason: r.",
				"",
				"```md",
				"## not-a-rule",
				"```",
			),
			wantRules: []string{"a@1|S.|r."},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, problems := parseRules(RulesPath, []byte(tt.data))
			if g, w := strings.Join(formatRules(got), "\n"), strings.Join(tt.wantRules, "\n"); g != w {
				t.Errorf("rules =\n%s\nwant\n%s", g, w)
			}
			for _, p := range problems {
				if p.File != RulesPath {
					t.Errorf("problem file = %q, want %q", p.File, RulesPath)
				}
			}
			checkProblems(t, problems, tt.wantProblems)
		})
	}
}

func TestParseRulesValid(t *testing.T) {
	tests := []struct {
		name         string
		data         string
		wantValid    bool
		wantAbsent   bool // the entry is skipped, so no rule is returned
		wantProblems []wantProblem
	}{
		{
			name:      "fully valid entry",
			data:      "## a\nStatement.\nReason: r.\n",
			wantValid: true,
		},
		{
			name:         "no Reason line",
			data:         "## a\nStatement.\n",
			wantProblems: []wantProblem{{1, `rule "a" has no Reason: line`}},
		},
		{
			name:         "empty Reason line",
			data:         "## a\nStatement.\nReason:   \n",
			wantProblems: []wantProblem{{3, `rule "a" has an empty Reason: line`}},
		},
		{
			name:         "repeated Reason line",
			data:         "## a\nS.\nReason: x\nReason: y\n",
			wantProblems: []wantProblem{{4, `rule "a" has more than one Reason: line`}},
		},
		{
			name:         "no statement",
			data:         "## a\nReason: x\n",
			wantProblems: []wantProblem{{1, `rule "a" has no statement`}},
		},
		{
			name:         "bad id is skipped",
			data:         "## Bad Id\nS.\nReason: r\n",
			wantAbsent:   true,
			wantProblems: []wantProblem{{1, `rule id "Bad Id" must be`}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, problems := parseRules(RulesPath, []byte(tt.data))
			checkProblems(t, problems, tt.wantProblems)
			if tt.wantAbsent {
				if len(rules) != 0 {
					t.Fatalf("got %d rules, want none", len(rules))
				}
				return
			}
			if len(rules) != 1 {
				t.Fatalf("got %d rules, want 1", len(rules))
			}
			if rules[0].Valid != tt.wantValid {
				t.Errorf("Valid = %t, want %t", rules[0].Valid, tt.wantValid)
			}
		})
	}
}

// formatRefs renders refs as "kind:id@line,..." for comparison.
func formatRefs(rs []Ref) string {
	out := make([]string, 0, len(rs))
	for _, r := range rs {
		out = append(out, fmt.Sprintf("%s:%s@%d", r.Kind, r.ID, r.Line))
	}
	return strings.Join(out, ",")
}

func TestParseRulesAppliesTo(t *testing.T) {
	tests := []struct {
		name         string
		data         string
		wantRules    []string // "id|refs|valid"
		wantProblems []wantProblem
	}{
		{
			name: "two refs, valid",
			data: joinLines(
				"# Rules",
				"",
				"## color.tokens-only",
				"Components take colours from tokens.",
				"Reason: Raw values drift.",
				"Applies to: tokens:color, component:button",
			),
			wantRules: []string{"color.tokens-only|tokens:color@6,component:button@6|true"},
		},
		{
			name: "every kind is a valid ref",
			data: joinLines(
				"## a.b",
				"Statement.",
				"Reason: r.",
				"Applies to: component:button, template:settings-page, flow:first-run, tokens:color",
			),
			wantRules: []string{"a.b|component:button@4,template:settings-page@4,flow:first-run@4,tokens:color@4|true"},
		},
		{
			name: "spaces around entries are accepted",
			data: joinLines(
				"## a.b",
				"Statement.",
				"Reason: r.",
				"Applies to:  tokens:color ,component:button  ",
			),
			wantRules: []string{"a.b|tokens:color@4,component:button@4|true"},
		},
		{
			name: "applies to line is not part of the statement",
			data: joinLines(
				"## a.b",
				"Statement.",
				"Reason: r.",
				"Applies to: component:button",
			),
			wantRules: []string{"a.b|component:button@4|true"},
		},
		{
			name: "unknown kind is reported and the valid ref is kept",
			data: joinLines(
				"## a.b",
				"Statement.",
				"Reason: r.",
				"Applies to: thing:x, component:button",
			),
			wantRules: []string{"a.b|component:button@4|false"},
			wantProblems: []wantProblem{
				{4, `unknown kind "thing" in Applies to:`},
			},
		},
		{
			name: "entry without a colon",
			data: joinLines(
				"## a.b",
				"Statement.",
				"Reason: r.",
				"Applies to: button",
			),
			wantRules: []string{"a.b||false"},
			wantProblems: []wantProblem{
				{4, `without a kind; write kind:id`},
			},
		},
		{
			name: "invalid item id",
			data: joinLines(
				"## a.b",
				"Statement.",
				"Reason: r.",
				"Applies to: component:Button",
			),
			wantRules: []string{"a.b||false"},
			wantProblems: []wantProblem{
				{4, `invalid id "Button" in Applies to:`},
			},
		},
		{
			name: "trailing comma is an empty entry",
			data: joinLines(
				"## a.b",
				"Statement.",
				"Reason: r.",
				"Applies to: component:button,",
			),
			wantRules: []string{"a.b|component:button@4|false"},
			wantProblems: []wantProblem{
				{4, "empty entry in Applies to:"},
			},
		},
		{
			name: "empty applies to line",
			data: joinLines(
				"## a.b",
				"Statement.",
				"Reason: r.",
				"Applies to:",
			),
			wantRules: []string{"a.b||false"},
			wantProblems: []wantProblem{
				{4, "empty Applies to: line"},
			},
		},
		{
			name: "second applies to line is reported and the first stands",
			data: joinLines(
				"## a.b",
				"Statement.",
				"Reason: r.",
				"Applies to: component:button",
				"Applies to: template:settings-page",
			),
			wantRules: []string{"a.b|component:button@4|false"},
			wantProblems: []wantProblem{
				{5, "more than one Applies to: line"},
			},
		},
		{
			name: "applies to inside a fence is ignored",
			data: joinLines(
				"## a.b",
				"Statement.",
				"Reason: r.",
				"```",
				"Applies to: component:x",
				"```",
			),
			wantRules: []string{"a.b||true"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, problems := parseRules("design-system/rules.md", []byte(tt.data))
			got := make([]string, 0, len(rules))
			for _, r := range rules {
				got = append(got, fmt.Sprintf("%s|%s|%t", r.ID, formatRefs(r.AppliesTo), r.Valid))
			}
			if !reflect.DeepEqual(got, tt.wantRules) {
				t.Errorf("rules = %v, want %v", got, tt.wantRules)
			}
			checkProblems(t, problems, tt.wantProblems)
			for _, r := range rules {
				if strings.Contains(r.Statement, "Applies to") {
					t.Errorf("statement %q contains the Applies to: line", r.Statement)
				}
			}
		})
	}
}
