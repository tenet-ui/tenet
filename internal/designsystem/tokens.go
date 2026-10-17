package designsystem

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
)

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// tokenParser walks tokens.json with a streaming decoder so every problem can
// carry the line of the token it describes.
type tokenParser struct {
	file     string
	data     []byte
	dec      *json.Decoder
	tokens   []Token
	problems []Problem
}

// parseTokens reads a tokens.json file. file is the project-relative path used in problems.
func parseTokens(file string, data []byte) ([]Token, []Problem) {
	data = stripBOM(data)
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, []Problem{{File: file, Line: 1, Message: "file is empty; write {} for no tokens"}}
	}
	p := &tokenParser{file: file, data: data, dec: json.NewDecoder(bytes.NewReader(data))}
	p.dec.UseNumber()
	p.parseDocument()
	return p.tokens, p.problems
}

func (p *tokenParser) add(line int, format string, args ...any) {
	p.problems = append(p.problems, Problem{File: p.file, Line: line, Message: fmt.Sprintf(format, args...)})
}

// line is the line of the decoder's current position, which is just after the last token read.
func (p *tokenParser) line() int {
	return lineAt(p.data, p.dec.InputOffset())
}

// next reads the next token. A decoding error becomes one problem and ok is false.
func (p *tokenParser) next() (json.Token, bool) {
	tok, err := p.dec.Token()
	if err == nil {
		return tok, true
	}
	line := lineAt(p.data, int64(len(p.data)))
	var se *json.SyntaxError
	if errors.As(err, &se) {
		line = lineAt(p.data, badByteIndex(se.Offset))
	}
	p.add(line, "invalid JSON: %s", err)
	return nil, false
}

// badByteIndex returns the index of the byte a syntax error points at. The
// decoder's Offset is always one past the bad byte, whatever that byte is.
func badByteIndex(off int64) int64 {
	return max(off-1, 0)
}

func (p *tokenParser) parseDocument() {
	tok, ok := p.next()
	if !ok {
		return
	}
	if d, isDelim := tok.(json.Delim); !isDelim || d != '{' {
		p.add(p.line(), "top level must be an object of categories")
		return
	}
	seen := map[string]bool{}
	for p.dec.More() {
		key, ok := p.next()
		if !ok {
			return
		}
		cat, _ := key.(string)
		if !p.parseCategory(cat, p.line(), seen) {
			return
		}
	}
	if tok, ok := p.next(); !ok || tok != json.Delim('}') {
		return
	}
	if _, err := p.dec.Token(); err != io.EOF {
		rest := p.data[p.dec.InputOffset():]
		trimmed := bytes.TrimLeft(rest, " \t\r\n")
		offset := p.dec.InputOffset() + int64(len(rest)-len(trimmed))
		p.add(lineAt(p.data, offset), "unexpected content after the top-level object")
	}
}

// parseCategory reads the value of one category key. Tokens are kept only when
// the category name is valid and not a duplicate; members are always checked.
// It returns false when decoding stopped.
func (p *tokenParser) parseCategory(cat string, keyLine int, seen map[string]bool) bool {
	dup := seen[cat]
	seen[cat] = true
	valid := namePattern.MatchString(cat)
	if !valid {
		p.add(keyLine, "category name %q must be lowercase letters, digits and dashes", cat)
	}
	if dup {
		p.add(keyLine, "duplicate category %q", cat)
	}
	keep := valid && !dup

	tok, ok := p.next()
	if !ok {
		return false
	}
	if d, isDelim := tok.(json.Delim); !isDelim || d != '{' {
		p.add(keyLine, "category %q must be an object of tokens", cat)
		return p.skipValue(tok)
	}
	names := map[string]bool{}
	for p.dec.More() {
		key, ok := p.next()
		if !ok {
			return false
		}
		name, _ := key.(string)
		if !p.parseToken(cat, name, p.line(), keep, names) {
			return false
		}
	}
	_, ok = p.next() // the closing brace of the category object
	return ok
}

// parseToken reads one token value and records it when keep is true and it is valid.
func (p *tokenParser) parseToken(cat, name string, line int, keep bool, names map[string]bool) bool {
	qualified := cat + "." + name
	vtok, ok := p.next()
	if !ok {
		return false
	}
	value, isScalar := scalarText(vtok)
	if !p.skipValue(vtok) {
		return false
	}
	if names[name] {
		p.add(line, "duplicate token %q", qualified)
		return true
	}
	names[name] = true
	switch {
	case !namePattern.MatchString(name):
		p.add(line, "token name %q must be lowercase letters, digits and dashes", name)
	case !isScalar:
		p.add(line, "token %q must be a string or number", qualified)
	case value == "":
		p.add(line, "token %q is empty", qualified)
	case keep:
		p.tokens = append(p.tokens, Token{Category: cat, Name: name, Value: value, Line: line})
	}
	return true
}

// skipValue consumes the rest of a value whose first token has already been read.
// Scalars need nothing more. Objects and arrays are read through their closing bracket.
func (p *tokenParser) skipValue(first json.Token) bool {
	d, isDelim := first.(json.Delim)
	if !isDelim || (d != '{' && d != '[') {
		return true
	}
	depth := 1
	for depth > 0 {
		tok, ok := p.next()
		if !ok {
			return false
		}
		if d, isDelim := tok.(json.Delim); isDelim {
			switch d {
			case '{', '[':
				depth++
			case '}', ']':
				depth--
			}
		}
	}
	return true
}

// scalarText returns the text of a string or number token.
func scalarText(tok json.Token) (string, bool) {
	switch v := tok.(type) {
	case string:
		return v, true
	case json.Number:
		return v.String(), true
	}
	return "", false
}
