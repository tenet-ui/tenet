// Command tally checks the consistency-experiment outputs against the reference pack
// and prints Markdown tables. Run it from the repository root.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tenet-ui/tenet/internal/designsystem"
)

var conditions = []string{"plain", "tools"}

type row struct {
	Condition  string
	Brief      string
	Status     string // "ok" or "failed"
	TokenOnly  bool
	Findings   int
	Unresolved int
}

func main() {
	packDir := flag.String("pack", "packs/stripe-reference", "pack folder, relative to the repository root")
	outDir := flag.String("outputs", "experiments/consistency/outputs", "outputs folder, relative to the repository root")
	flag.Parse()
	rows, err := collect(*packDir, *outDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(render(rows))
	fmt.Println()
	for _, line := range summarize(rows) {
		fmt.Println(line)
	}
}

func collect(packDir, outDir string) ([]row, error) {
	ds, problems := designsystem.Load(packDir)
	if len(problems) > 0 {
		var lines []string
		for _, p := range problems {
			lines = append(lines, p.String())
		}
		return nil, fmt.Errorf("pack %s has problems:\n%s", packDir, strings.Join(lines, "\n"))
	}
	tokens, err := os.ReadFile(filepath.Join(packDir, designsystem.TokensPath))
	if err != nil {
		return nil, err
	}
	names, err := flattenTokens(tokens)
	if err != nil {
		return nil, err
	}

	briefs := map[string]bool{}
	for _, cond := range conditions {
		entries, err := os.ReadDir(filepath.Join(outDir, cond))
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if id, ok := briefID(e.Name()); ok {
				briefs[id] = true
			}
		}
	}

	var rows []row
	for _, cond := range conditions {
		for _, id := range sortedKeys(briefs) {
			r, err := readRow(ds.Rules, names, filepath.Join(outDir, cond), cond, id)
			if err != nil {
				return nil, err
			}
			rows = append(rows, r)
		}
	}
	return rows, nil
}

func readRow(rules []designsystem.Rule, names map[string]bool, dir, cond, id string) (row, error) {
	htmlPath := filepath.Join(dir, id+".html")
	_, htmlErr := os.Stat(htmlPath)
	_, failedErr := os.Stat(filepath.Join(dir, id+".failed"))
	hasHTML, hasFailed := htmlErr == nil, failedErr == nil
	switch {
	case hasHTML && hasFailed:
		return row{}, fmt.Errorf("%s/%s: both .html and .failed exist; remove the stale one", cond, id)
	case hasFailed:
		return row{Condition: cond, Brief: id, Status: "failed"}, nil
	case hasHTML:
		src, err := os.ReadFile(htmlPath)
		if err != nil {
			return row{}, err
		}
		text := string(src)
		findings := len(designsystem.Critique(text, rules))
		return row{
			Condition:  cond,
			Brief:      id,
			Status:     "ok",
			TokenOnly:  findings == 0 && hasVarRef(text),
			Findings:   findings,
			Unresolved: len(unresolvedNames(text, names)),
		}, nil
	default:
		return row{}, fmt.Errorf("%s/%s: no .html or .failed file", cond, id)
	}
}

func render(rows []row) string {
	var b strings.Builder
	b.WriteString("| Condition | Brief | Status | Token-only | Critique findings | Unresolved var names |\n")
	b.WriteString("|---|---|---|---|---|---|\n")
	for _, r := range rows {
		if r.Status == "failed" {
			fmt.Fprintf(&b, "| %s | %s | failed | no | – | – |\n", r.Condition, r.Brief)
			continue
		}
		fmt.Fprintf(&b, "| %s | %s | ok | %s | %d | %d |\n", r.Condition, r.Brief, yesNo(r.TokenOnly), r.Findings, r.Unresolved)
	}
	return b.String()
}

func summarize(rows []row) []string {
	var lines []string
	for _, cond := range conditions {
		var ok, failed, tokenOnly, findings int
		for _, r := range rows {
			if r.Condition != cond {
				continue
			}
			if r.Status == "failed" {
				failed++
				continue
			}
			ok++
			findings += r.Findings
			if r.TokenOnly {
				tokenOnly++
			}
		}
		mean := "n/a"
		if ok > 0 {
			mean = fmt.Sprintf("%.1f", float64(findings)/float64(ok))
		}
		lines = append(lines, fmt.Sprintf("- %s: %d token-only of %d; %d findings in total; %s findings per ok output; %d ok, %d failed",
			cond, tokenOnly, ok+failed, findings, mean, ok, failed))
	}
	return lines
}

func briefID(name string) (string, bool) {
	for _, ext := range []string{".html", ".failed"} {
		if strings.HasSuffix(name, ext) {
			return strings.TrimSuffix(name, ext), true
		}
	}
	return "", false
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
