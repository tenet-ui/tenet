package server

import (
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tenet-ui/tenet/internal/designsystem"
)

const rulesNoteText = "Note: design-system/rules.md has problems, so some linked rules may be missing. Run tenet validate."

// deliveryText builds the text of a delivery result: the lead text, then the linked
// rules in the rules.md format (only if there are any), then the rules note (only if asked).
// This is the one place that defines the shape of a delivery result.
func deliveryText(lead string, rules []designsystem.Rule, rulesNote bool) string {
	var b strings.Builder
	b.WriteString(strings.TrimRight(lead, " \t\r\n"))
	b.WriteString("\n")
	if len(rules) > 0 {
		b.WriteString("\n## Linked rules\n")
		for _, r := range rules {
			fmt.Fprintf(&b, "\n### %s\n%s\nReason: %s\n", r.ID, r.Statement, r.Reason)
		}
	}
	if rulesNote {
		b.WriteString("\n" + rulesNoteText + "\n")
	}
	return b.String()
}

// deliveryTarget describes one kind of delivery target for messages and file lookups.
type deliveryTarget struct {
	singular string
	plural   string
	none     string // used in "has no <none> yet"
	hint     string // what to add to get one
	dir      string // item folder; empty for tokens
	kind     designsystem.Kind
}

var (
	componentKind = deliveryTarget{"component", "components", "components", fmt.Sprintf("add Markdown files under %s/", designsystem.ComponentsDir), designsystem.ComponentsDir, designsystem.KindComponent}
	templateKind  = deliveryTarget{"template", "templates", "templates", fmt.Sprintf("add Markdown files under %s/", designsystem.TemplatesDir), designsystem.TemplatesDir, designsystem.KindTemplate}
	flowKind      = deliveryTarget{"flow", "flows", "flows", fmt.Sprintf("add Markdown files under %s/", designsystem.FlowsDir), designsystem.FlowsDir, designsystem.KindFlow}
	tokenKind     = deliveryTarget{"token category", "token categories", "tokens", "add categories to design-system/tokens.json", "", designsystem.KindTokens}
)

// notFoundText names the missing id and lists what exists, or says the kind has none yet.
func notFoundText(k deliveryTarget, id string, available []string) string {
	if len(available) == 0 {
		return fmt.Sprintf("%s %q not found. The design system has no %s yet; %s.", k.singular, id, k.none, k.hint)
	}
	return fmt.Sprintf("%s %q not found. Available %s: %s.", k.singular, id, k.plural, strings.Join(available, ", "))
}

// invalidText reports that the requested item exists on disk but its file has a problem.
func invalidText(k deliveryTarget, id string, p designsystem.Problem) string {
	return fmt.Sprintf("%s %q is not valid: %s. Run tenet validate.", k.singular, id, p.String())
}

// GetTokensInput is the input of get_tokens.
type GetTokensInput struct {
	Category string `json:"category" jsonschema:"the token category, for example color"`
}

// GetItemInput is the input of get_component, get_template and get_flow.
type GetItemInput struct {
	ID string `json:"id" jsonschema:"the item id, the file name without .md"`
}

// loadPack loads the design system. When the design-system folder itself has a problem it
// returns an error result, which always points the user to tenet init.
func loadPack(root string) (designsystem.DesignSystem, []designsystem.Problem, *mcp.CallToolResult) {
	pack, problems := designsystem.Load(root)
	if p, ok := problemFor(problems, designsystem.DirName); ok {
		msg := p.String()
		if !strings.Contains(msg, "run tenet init") {
			msg += "; run tenet init"
		}
		return designsystem.DesignSystem{}, nil, toolError(msg)
	}
	return pack, problems, nil
}

// rulesNote reports whether rules.md has a problem, so the result should carry the rules note.
func rulesNote(problems []designsystem.Problem) bool {
	_, ok := problemFor(problems, designsystem.RulesPath)
	return ok
}

// deliverItem loads the pack and returns the item with its linked rules.
func deliverItem(root string, k deliveryTarget, id string) *mcp.CallToolResult {
	pack, problems, errRes := loadPack(root)
	if errRes != nil {
		return errRes
	}
	if item, ok := pack.Item(k.kind, id); ok {
		return textResult(deliveryText(item.Body, pack.Linked(k.kind, id), rulesNote(problems)))
	}
	file := k.dir + "/" + id + ".md"
	return toolError(missingText(problems, k, file, id, pack.ItemIDs(k.kind)))
}

// deliverTokens loads the pack and returns one line per token of the category, with its linked rules.
func deliverTokens(root, category string) *mcp.CallToolResult {
	pack, problems, errRes := loadPack(root)
	if errRes != nil {
		return errRes
	}
	if tokens, ok := pack.Category(category); ok {
		lines := make([]string, 0, len(tokens))
		for _, t := range tokens {
			lines = append(lines, fmt.Sprintf("%s.%s: %s", t.Category, t.Name, t.Value))
		}
		return textResult(deliveryText(strings.Join(lines, "\n"), pack.Linked(designsystem.KindTokens, category), rulesNote(problems)))
	}
	return toolError(missingText(problems, tokenKind, designsystem.TokensPath, category, pack.Categories()))
}

// missingText prefers the file's own problem over a plain not-found message.
func missingText(problems []designsystem.Problem, k deliveryTarget, file, id string, available []string) string {
	if p, ok := problemFor(problems, file); ok {
		return invalidText(k, id, p)
	}
	return notFoundText(k, id, available)
}

// problemFor returns the first problem reported for file.
func problemFor(problems []designsystem.Problem, file string) (designsystem.Problem, bool) {
	for _, p := range problems {
		if p.File == file {
			return p, true
		}
	}
	return designsystem.Problem{}, false
}

func textResult(s string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: s}}}
}

func toolError(s string) *mcp.CallToolResult {
	return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: s}}}
}
