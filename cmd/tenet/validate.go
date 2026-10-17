package main

import (
	"fmt"
	"io"

	"github.com/tenet-ui/tenet/internal/designsystem"
)

// runValidate implements `tenet validate [--dir <path>]`. Exit codes: 0 valid,
// 1 problems found, 2 usage or --dir error.
func runValidate(args []string, cwd string, stdout, stderr io.Writer) int {
	root, err := parseRoot(args, cwd)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "tenet validate:", err)
		return 2
	}
	pack, problems := designsystem.Load(root)
	if len(problems) > 0 {
		for _, p := range problems {
			_, _ = fmt.Fprintln(stdout, p)
		}
		_, _ = fmt.Fprintf(stdout, "%d problem(s) found\n", len(problems))
		return 1
	}
	_, _ = fmt.Fprintf(stdout, "design system valid: %s\n", packSummary(pack))
	return 0
}

// packSummary counts each kind of item in a pack, for output lines.
func packSummary(p designsystem.Pack) string {
	return fmt.Sprintf("%d tokens, %d rules, %d components, %d templates, %d flows",
		len(p.Tokens), len(p.Rules), len(p.Components), len(p.Templates), len(p.Flows))
}
