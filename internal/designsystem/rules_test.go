package designsystem

import (
	"fmt"
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
