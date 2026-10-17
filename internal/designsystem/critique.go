package designsystem

import "strings"

// Finding is one breach of one rule in one source file.
type Finding struct {
	Line   int    // 1-based line in the source
	RuleID string // the rule that was broken
	Match  string // the text the rule's pattern matched
	Fix    string // the rule's Fix: text
}

// Checkable returns the number of valid rules that have a Forbid pattern.
func Checkable(rules []Rule) int {
	n := 0
	for _, r := range rules {
		if r.Valid && r.Forbid != nil {
			n++
		}
	}
	return n
}

// Critique matches the Forbid pattern of each checkable rule against each line of source.
// Lines are split on "\n" and a trailing "\r" is removed from each. A leading UTF-8 BOM is
// removed first. Zero-length matches are ignored. Findings are in line order, then rule
// order, then match order.
func Critique(source string, rules []Rule) []Finding {
	text := string(stripBOM([]byte(source)))
	var findings []Finding
	for i, line := range strings.Split(text, "\n") {
		line = strings.TrimSuffix(line, "\r")
		for _, r := range rules {
			if !r.Valid || r.Forbid == nil {
				continue
			}
			for _, m := range r.Forbid.FindAllString(line, -1) {
				if m == "" {
					continue
				}
				findings = append(findings, Finding{Line: i + 1, RuleID: r.ID, Match: m, Fix: r.Fix})
			}
		}
	}
	return findings
}
