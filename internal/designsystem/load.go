package designsystem

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	// DirName is the design system folder, relative to the project root.
	DirName = "design-system"
	// TokensPath is the tokens file, relative to the project root.
	TokensPath = "design-system/tokens.json"
	// RulesPath is the rules file, relative to the project root.
	RulesPath = "design-system/rules.md"
	// ComponentsDir, TemplatesDir and FlowsDir each hold one Markdown file per item.
	ComponentsDir = "design-system/components"
	TemplatesDir  = "design-system/templates"
	FlowsDir      = "design-system/flows"
)

// Token is one entry of tokens.json.
type Token struct {
	Category string
	Name     string
	Value    string // string value, or the number's literal text
	Line     int
}

// Kind names what a rule can apply to. Its values are the kind part of a "kind:id" ref.
type Kind string

const (
	KindComponent Kind = "component"
	KindTemplate  Kind = "template"
	KindFlow      Kind = "flow"
	KindTokens    Kind = "tokens" // id is a token category name
)

// Ref is one entry of a rule's Applies to: line.
type Ref struct {
	Kind Kind
	ID   string
	Line int // 1-based line of the Applies to: line in rules.md
}

// Rule is one "## <id>" entry of rules.md.
type Rule struct {
	ID        string
	Statement string // body lines other than the Reason: line, Applies to: line and fenced code, joined with a space
	Reason    string
	Line      int   // line of the "## <id>" heading
	AppliesTo []Ref // refs in the order written
	Valid     bool  // false when parseRules reported any problem for this entry
}

// Item is one Markdown file in components/, templates/ or flows/.
type Item struct {
	ID    string // file name without ".md"
	Title string // text of the first "# " heading
	Path  string // slash-separated, relative to the project root
	Body  string // full file text, BOM stripped
}

// DesignSystem is everything the loader read from design-system/.
type DesignSystem struct {
	PointOfView string // empty when the file has problems
	Tokens      []Token
	Rules       []Rule
	Components  []Item
	Templates   []Item
	Flows       []Item
}

// Load reads root/design-system and returns the pack it could read and every
// problem it found. A non-empty problems slice means the pack is not valid,
// but Load still returns whatever it could parse.
func Load(root string) (DesignSystem, []Problem) {
	var pack DesignSystem
	info, err := os.Stat(filepath.Join(root, DirName))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return pack, []Problem{{File: DirName, Message: "directory not found; run tenet init"}}
	case err != nil:
		return pack, []Problem{{File: DirName, Message: "could not read: " + err.Error()}}
	case !info.IsDir():
		return pack, []Problem{{File: DirName, Message: "not a directory"}}
	}

	var problems []Problem
	text, err := ReadPointOfView(root)
	switch {
	case errors.Is(err, ErrPointOfViewMissing):
		problems = append(problems, Problem{File: PointOfViewPath, Message: "file not found"})
	case errors.Is(err, ErrPointOfViewEmpty):
		problems = append(problems, Problem{File: PointOfViewPath, Line: 1, Message: "file is empty"})
	case err != nil:
		problems = append(problems, Problem{File: PointOfViewPath, Message: "could not read: " + err.Error()})
	default:
		pack.PointOfView = text
	}

	tokens, ps := loadTokens(root)
	pack.Tokens = tokens
	problems = append(problems, ps...)

	rules, ps := loadRules(root)
	pack.Rules = rules
	problems = append(problems, ps...)

	components, ps := loadItems(root, ComponentsDir)
	pack.Components = components
	problems = append(problems, ps...)

	templates, ps := loadItems(root, TemplatesDir)
	pack.Templates = templates
	problems = append(problems, ps...)

	flows, ps := loadItems(root, FlowsDir)
	pack.Flows = flows
	problems = append(problems, ps...)

	problems = append(problems, checkRefs(pack)...)
	return pack, problems
}

// loadTokens reads and parses the required tokens file.
func loadTokens(root string) ([]Token, []Problem) {
	data, problems := readRequired(root, TokensPath)
	if problems != nil {
		return nil, problems
	}
	return parseTokens(TokensPath, data)
}

// loadRules reads and parses the required rules file.
func loadRules(root string) ([]Rule, []Problem) {
	data, problems := readRequired(root, RulesPath)
	if problems != nil {
		return nil, problems
	}
	return parseRules(RulesPath, data)
}

// checkRefs reports every rule ref that names an item or token category the pack does not have.
func checkRefs(pack DesignSystem) []Problem {
	var problems []Problem
	for _, r := range pack.Rules {
		for _, ref := range r.AppliesTo {
			if refExists(pack, ref) {
				continue
			}
			problems = append(problems, Problem{
				File:    RulesPath,
				Line:    ref.Line,
				Message: fmt.Sprintf("rule %q applies to unknown %s %q", r.ID, refLabel(ref.Kind), ref.ID),
			})
		}
	}
	return problems
}

func refExists(pack DesignSystem, ref Ref) bool {
	if ref.Kind == KindTokens {
		_, ok := pack.Category(ref.ID)
		return ok
	}
	_, ok := pack.Item(ref.Kind, ref.ID)
	return ok
}

func refLabel(kind Kind) string {
	if kind == KindTokens {
		return "token category"
	}
	return string(kind)
}

// readRequired reads a required file. A missing or unreadable file becomes one problem.
func readRequired(root, rel string) ([]byte, []Problem) {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil, []Problem{{File: rel, Message: "file not found"}}
	case err != nil:
		return nil, []Problem{{File: rel, Message: "could not read: " + err.Error()}}
	}
	return data, nil
}
