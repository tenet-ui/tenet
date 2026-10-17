// Command tenet serves a project's design system to MCP clients over stdio, and
// checks (validate) or scaffolds (init) its design-system folder.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tenet-ui/tenet/internal/designsystem"
	"github.com/tenet-ui/tenet/internal/server"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "tenet:", err)
		os.Exit(1)
	}
	os.Exit(run(os.Args[1:], cwd, os.Stdout, os.Stderr))
}

// run dispatches on the first argument and returns the process exit code. With no
// subcommand it serves MCP over stdio.
func run(args []string, cwd string, stdout, stderr io.Writer) int {
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		switch args[0] {
		case "validate":
			return runValidate(args[1:], cwd, stdout, stderr)
		case "init":
			return runInit(args[1:], cwd, stdout, stderr)
		default:
			_, _ = fmt.Fprintf(stderr, "tenet: unknown command %q\n", args[0])
			return 2
		}
	}
	return runServe(args, cwd, stderr)
}

// runServe runs the stdio MCP server. It writes nothing to stdout.
func runServe(args []string, cwd string, stderr io.Writer) int {
	logger := log.New(stderr, "tenet: ", 0)
	root, err := parseRoot(args, cwd)
	if err != nil {
		logger.Println(err)
		return 2
	}
	pack, problems := designsystem.Load(root)
	logger.Printf("%s serving %s on stdio; design system: %s", server.Version, root, packSummary(pack))
	if len(problems) > 0 {
		logger.Printf("design system has %d problem(s); run tenet validate for details", len(problems))
		for _, p := range problems {
			logger.Println(p)
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := server.New(root).Run(ctx, &mcp.StdioTransport{}); err != nil && ctx.Err() == nil {
		logger.Println(err)
		return 1
	}
	return 0
}

// parseRoot reads --dir and returns the cleaned absolute project root.
// It defaults to cwd, and a relative --dir is resolved against cwd.
func parseRoot(args []string, cwd string) (string, error) {
	fs := flag.NewFlagSet("tenet", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	dir := fs.String("dir", "", "project root (default: current directory)")
	if err := fs.Parse(args); err != nil {
		return "", err
	}

	root := *dir
	if root == "" {
		root = cwd
	} else if !filepath.IsAbs(root) {
		root = filepath.Join(cwd, root)
	}
	root = filepath.Clean(root)

	info, err := os.Stat(root)
	if err != nil {
		return "", fmt.Errorf("--dir %s: %w", root, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("--dir %s is not a directory", root)
	}
	return root, nil
}
