package server

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tenet-ui/tenet/internal/designsystem"
)

const (
	critiqueMaxBytes   = 1 << 20
	critiqueBinaryScan = 8000
	critiqueRulesNote  = "Note: design-system/rules.md has problems, so some rules may not be checked. Run tenet validate."
)

// CritiqueInput is the input of critique.
type CritiqueInput struct {
	File   string `json:"file" jsonschema:"path to the UI source file, relative to the project root or absolute"`
	System string `json:"system" jsonschema:"folder holding the design-system/ to check against, relative to the project root; . for the project root"`
}

// critiqueTool checks one source file against the checkable rules of one design system.
func critiqueTool(root string, in CritiqueInput) *mcp.CallToolResult {
	if in.File == "" {
		return toolError("file is empty; pass the path to a UI source file.")
	}
	if in.System == "" {
		return toolError("system is empty; pass a folder, or . for the project root.")
	}
	path := in.File
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	info, err := os.Stat(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return toolError(fmt.Sprintf("file %q not found.", in.File))
	case err != nil:
		return toolError(fmt.Sprintf("could not read file %q: %v", in.File, err))
	case info.IsDir():
		return toolError(fmt.Sprintf("file %q is a directory; pass a file.", in.File))
	case !info.Mode().IsRegular():
		return toolError(fmt.Sprintf("file %q is not a regular file; pass a file.", in.File))
	case info.Size() > critiqueMaxBytes:
		return toolError(fmt.Sprintf("file %q is larger than 1 MiB; pass a smaller source file.", in.File))
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return toolError(fmt.Sprintf("could not read file %q: %v", in.File, err))
	}
	if bytes.IndexByte(data[:min(len(data), critiqueBinaryScan)], 0) >= 0 {
		return toolError(fmt.Sprintf("file %q is not a text file.", in.File))
	}

	systemRoot := in.System
	if !filepath.IsAbs(systemRoot) {
		systemRoot = filepath.Join(root, systemRoot)
	}
	pack, problems := designsystem.Load(systemRoot)
	if p, ok := problemFor(problems, designsystem.DirName); ok {
		if strings.HasPrefix(p.Message, "directory not found") {
			return toolError(fmt.Sprintf("design-system not found under %s; run tenet init there.", in.System))
		}
		msg := p.String()
		if !strings.Contains(msg, "run tenet init") {
			msg += "; run tenet init"
		}
		return toolError(msg)
	}
	findings := designsystem.Critique(string(data), pack.Rules)
	return textResult(renderCritique(in.File, in.System, findings, designsystem.Checkable(pack.Rules), rulesProblem(problems)))
}

// rulesProblem reports whether rules.md has a problem, so the result should carry the rules note.
func rulesProblem(problems []designsystem.Problem) bool {
	_, ok := problemFor(problems, designsystem.RulesPath)
	return ok
}

// renderCritique builds the text of a critique result. With findings it is a header, a blank
// line and one line per finding. Without findings it says how many rules were checked, or that
// none could be. The rules note follows a blank line when rules.md has problems.
func renderCritique(file, system string, findings []designsystem.Finding, checkable int, note bool) string {
	var lines []string
	switch {
	case checkable == 0:
		lines = []string{fmt.Sprintf("No rule in %s has a Forbid: line, so nothing was checked. Add one under a rule in design-system/rules.md.", system)}
	case len(findings) == 0:
		lines = []string{fmt.Sprintf("No findings in %s against %s. Checked %d rule(s).", file, system, checkable)}
	default:
		lines = []string{fmt.Sprintf("%d finding(s) in %s against %s.", len(findings), file, system), ""}
		for _, f := range findings {
			lines = append(lines, fmt.Sprintf("%s:%d: %s: found \"%s\". %s", file, f.Line, f.RuleID, f.Match, f.Fix))
		}
	}
	text := strings.Join(lines, "\n") + "\n"
	if note {
		text += "\n" + critiqueRulesNote + "\n"
	}
	return text
}
