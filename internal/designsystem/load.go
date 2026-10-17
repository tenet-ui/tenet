package designsystem

import (
	"errors"
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

// Rule is one "## <id>" entry of rules.md.
type Rule struct {
	ID        string
	Statement string // body lines other than the Reason: line and fenced code, joined with a space
	Reason    string
	Line      int // line of the "## <id>" heading
}

// Item is one Markdown file in components/, templates/ or flows/.
type Item struct {
	ID    string // file name without ".md"
	Title string // text of the first "# " heading
	Path  string // slash-separated, relative to the project root
}

// Pack is everything the loader read from design-system/.
type Pack struct {
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
func Load(root string) (Pack, []Problem) {
	var pack Pack
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
