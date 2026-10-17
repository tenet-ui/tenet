package designsystem

import (
	"reflect"
	"regexp"
	"testing"
)

// rule builds a valid checkable rule for the critique tests.
func rule(id, pattern, fix string) Rule {
	return Rule{ID: id, Valid: true, Forbid: regexp.MustCompile(pattern), Fix: fix}
}

func TestCritique(t *testing.T) {
	tests := []struct {
		name   string
		source string
		rules  []Rule
		want   []Finding
	}{
		{
			name:   "two matches on one line are two findings in match order",
			source: "a a",
			rules:  []Rule{rule("r", "a", "F.")},
			want: []Finding{
				{Line: 1, RuleID: "r", Match: "a", Fix: "F."},
				{Line: 1, RuleID: "r", Match: "a", Fix: "F."},
			},
		},
		{
			name:   "findings sort by line, then rule order",
			source: "a b\nb",
			rules:  []Rule{rule("a", "b", "Fa."), rule("b", "a", "Fb.")},
			want: []Finding{
				{Line: 1, RuleID: "a", Match: "b", Fix: "Fa."},
				{Line: 1, RuleID: "b", Match: "a", Fix: "Fb."},
				{Line: 2, RuleID: "a", Match: "b", Fix: "Fa."},
			},
		},
		{
			name:   "invalid rule is skipped",
			source: "a",
			rules:  []Rule{{ID: "x", Valid: false, Forbid: regexp.MustCompile("a"), Fix: "F."}},
			want:   nil,
		},
		{
			name:   "rule without Forbid is skipped",
			source: "a",
			rules:  []Rule{{ID: "x", Valid: true}},
			want:   nil,
		},
		{
			name:   "no rules, no findings",
			source: "a",
			rules:  nil,
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Critique(tt.source, tt.rules)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Critique() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestCritiqueIgnoresEmptyMatches(t *testing.T) {
	got := Critique("abc", []Rule{rule("r", "x*", "F.")})
	if len(got) != 0 {
		t.Errorf("Critique() = %+v, want no findings", got)
	}
}

func TestCritiqueCRLF(t *testing.T) {
	got := Critique("a\r\nb\r\n", []Rule{rule("r", "a$", "F.")})
	want := []Finding{{Line: 1, RuleID: "r", Match: "a", Fix: "F."}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Critique() = %+v, want %+v", got, want)
	}
}

func TestCritiqueStripsBOM(t *testing.T) {
	got := Critique("\uFEFFa", []Rule{rule("r", "^a", "F.")})
	want := []Finding{{Line: 1, RuleID: "r", Match: "a", Fix: "F."}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Critique() = %+v, want %+v", got, want)
	}
}

func TestCheckable(t *testing.T) {
	rules := []Rule{
		rule("a", "x", "F."),
		{ID: "b", Valid: true},
		{ID: "c", Valid: false, Forbid: regexp.MustCompile("x"), Fix: "F."},
	}
	if got := Checkable(rules); got != 1 {
		t.Errorf("Checkable() = %d, want 1", got)
	}
}
