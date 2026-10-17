package designsystem

import (
	"fmt"
	"regexp"
	"strings"
)

var ruleIDPattern = regexp.MustCompile(`^[a-z0-9]+([.-][a-z0-9]+)*$`)

// numberedLine is one body line of a rule entry with its 1-based line number.
type numberedLine struct {
	no   int
	text string
}

// ruleEntry is one "## <id>" section before its body is interpreted.
type ruleEntry struct {
	id   string
	line int
	body []numberedLine
}

// parseRules reads rules.md. A line "## <id>" outside a fenced code block starts
// a rule. Its body holds a statement and one "Reason:" line.
func parseRules(file string, data []byte) ([]Rule, []Problem) {
	var (
		entries []*ruleEntry
		current *ruleEntry
		inFence bool
	)
	for i, raw := range strings.Split(string(stripBOM(data)), "\n") {
		text := strings.TrimRight(raw, "\r")
		if strings.HasPrefix(strings.TrimSpace(text), "```") {
			inFence = !inFence
			continue
		}
		if !inFence && (text == "##" || strings.HasPrefix(text, "## ")) {
			current = &ruleEntry{id: strings.TrimSpace(strings.TrimPrefix(text, "##")), line: i + 1}
			entries = append(entries, current)
			continue
		}
		if current != nil && !inFence {
			current.body = append(current.body, numberedLine{no: i + 1, text: text})
		}
	}

	var (
		rules    []Rule
		problems []Problem
		seen     = map[string]int{}
	)
	add := func(line int, format string, args ...any) {
		problems = append(problems, Problem{File: file, Line: line, Message: fmt.Sprintf(format, args...)})
	}
	for _, e := range entries {
		if !ruleIDPattern.MatchString(e.id) {
			add(e.line, "rule id %q must be lowercase letters and digits joined by dots or dashes", e.id)
			continue
		}
		if first, dup := seen[e.id]; dup {
			add(e.line, "duplicate rule id %q (first at line %d)", e.id, first)
			continue
		}
		seen[e.id] = e.line

		var (
			statement  []string
			reason     string
			reasonLine int
			reasons    int
		)
		for _, l := range e.body {
			t := strings.TrimSpace(l.text)
			if strings.HasPrefix(t, "Reason:") {
				reasons++
				if reasons == 1 {
					reason = strings.TrimSpace(strings.TrimPrefix(t, "Reason:"))
					reasonLine = l.no
				} else {
					add(l.no, "rule %q has more than one Reason: line", e.id)
				}
				continue
			}
			if t != "" {
				statement = append(statement, t)
			}
		}
		if len(statement) == 0 {
			add(e.line, "rule %q has no statement", e.id)
		}
		if reasons == 0 {
			add(e.line, "rule %q has no Reason: line", e.id)
		} else if reason == "" {
			add(reasonLine, "rule %q has an empty Reason: line", e.id)
		}
		rules = append(rules, Rule{
			ID:        e.id,
			Statement: strings.Join(statement, " "),
			Reason:    reason,
			Line:      e.line,
		})
	}
	return rules, problems
}
