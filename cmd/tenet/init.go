package main

import (
	"fmt"
	"io"

	"github.com/tenet-ui/tenet/internal/designsystem"
)

// runInit implements `tenet init [--dir <path>]`. Exit codes: 0 written, 1 the
// folder already exists or the write failed, 2 usage or --dir error.
func runInit(args []string, cwd string, stdout, stderr io.Writer) int {
	root, err := parseRoot(args, cwd)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "tenet init:", err)
		return 2
	}
	if err := designsystem.Scaffold(root); err != nil {
		_, _ = fmt.Fprintln(stderr, "tenet init:", err)
		return 1
	}
	_, _ = fmt.Fprintf(stdout, "created design-system/ under %s; next, write design-system/point-of-view.md\n", root)
	return 0
}
