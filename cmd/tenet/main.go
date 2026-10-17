// Command tenet serves a project's design system to MCP clients over stdio.
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
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tenet-ui/tenet/internal/server"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "tenet:", err)
		os.Exit(1)
	}
	os.Exit(run(os.Args[1:], cwd, os.Stderr))
}

func run(args []string, cwd string, stderr io.Writer) int {
	logger := log.New(stderr, "tenet: ", 0)
	root, err := parseRoot(args, cwd)
	if err != nil {
		logger.Println(err)
		return 2
	}
	logger.Printf("%s serving %s on stdio", server.Version, root)

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
