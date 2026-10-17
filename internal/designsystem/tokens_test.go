package designsystem

import (
	"fmt"
	"strings"
	"testing"
)

// formatTokens renders each token as "category.name=value@line" for comparison.
func formatTokens(ts []Token) []string {
	out := make([]string, 0, len(ts))
	for _, tk := range ts {
		out = append(out, fmt.Sprintf("%s.%s=%s@%d", tk.Category, tk.Name, tk.Value, tk.Line))
	}
	return out
}

func joinLines(lines ...string) string {
	return strings.Join(lines, "\n")
}

func TestParseTokens(t *testing.T) {
	tests := []struct {
		name         string
		data         string
		wantTokens   []string
		wantProblems []wantProblem
	}{
		{
			name:       "valid two categories",
			data:       `{"color":{"ink":"#1d2433"},"space":{"unit":4,"half":1.5}}`,
			wantTokens: []string{"color.ink=#1d2433@1", "space.unit=4@1", "space.half=1.5@1"},
		},
		{
			name:       "negative and exponent numbers",
			data:       `{"space":{"neg":-1,"big":1e3}}`,
			wantTokens: []string{"space.neg=-1@1", "space.big=1e3@1"},
		},
		{
			name:         "empty file",
			data:         "",
			wantProblems: []wantProblem{{1, "file is empty"}},
		},
		{
			name:         "whitespace only",
			data:         "  \n \n",
			wantProblems: []wantProblem{{1, "file is empty"}},
		},
		{
			name:       "empty object is valid",
			data:       "{}\n",
			wantTokens: nil,
		},
		{
			name:       "utf-8 BOM is ignored",
			data:       "\ufeff{\"color\":{\"ink\":\"#fff\"}}",
			wantTokens: []string{"color.ink=#fff@1"},
		},
		{
			name:       "CRLF line endings",
			data:       "{\r\n  \"color\": {\r\n    \"ink\": \"#fff\"\r\n  }\r\n}\r\n",
			wantTokens: []string{"color.ink=#fff@3"},
		},
		{
			name:         "top level is an array",
			data:         `[]`,
			wantProblems: []wantProblem{{1, "top level must be an object"}},
		},
		{
			name:         "category is not an object",
			data:         `{"color": "red"}`,
			wantProblems: []wantProblem{{1, `category "color" must be an object`}},
		},
		{
			name:         "nested object value",
			data:         `{"color": {"brand": {"primary": "#fff"}}}`,
			wantProblems: []wantProblem{{1, `token "color.brand" must be a string or number`}},
		},
		{
			name:         "boolean value",
			data:         `{"color":{"on":true}}`,
			wantProblems: []wantProblem{{1, `token "color.on" must be a string or number`}},
		},
		{
			name:         "null value",
			data:         `{"color":{"on":null}}`,
			wantProblems: []wantProblem{{1, `token "color.on" must be a string or number`}},
		},
		{
			name:         "array value",
			data:         `{"color":{"list":[1,2]}}`,
			wantProblems: []wantProblem{{1, `token "color.list" must be a string or number`}},
		},
		{
			name:         "empty string value",
			data:         `{"color":{"ink":""}}`,
			wantProblems: []wantProblem{{1, `token "color.ink" is empty`}},
		},
		{
			name:         "invalid token name",
			data:         `{"color":{"Ink":"#fff"}}`,
			wantProblems: []wantProblem{{1, `token name "Ink" must be lowercase`}},
		},
		{
			name:         "token name with underscore",
			data:         `{"color":{"a_b":"#fff"}}`,
			wantProblems: []wantProblem{{1, `token name "a_b" must be lowercase`}},
		},
		{
			name:         "token name starting with a dash",
			data:         `{"color":{"-a":"#fff"}}`,
			wantProblems: []wantProblem{{1, `token name "-a" must be lowercase`}},
		},
		{
			name:         "invalid category name",
			data:         `{"Color":{"ink":"#fff"}}`,
			wantProblems: []wantProblem{{1, `category name "Color" must be lowercase`}},
		},
		{
			name: "duplicate category reports second key and keeps first",
			data: joinLines(
				"{",
				`  "color": {"ink": "#fff"},`,
				`  "color": {"paper": "#000"}`,
				"}",
			),
			wantTokens:   []string{"color.ink=#fff@2"},
			wantProblems: []wantProblem{{3, `duplicate category "color"`}},
		},
		{
			name: "duplicate token reports second key and keeps first",
			data: joinLines(
				"{",
				`  "color": {`,
				`    "ink": "#fff",`,
				`    "ink": "#000"`,
				"  }",
				"}",
			),
			wantTokens:   []string{"color.ink=#fff@3"},
			wantProblems: []wantProblem{{4, `duplicate token "color.ink"`}},
		},
		{
			name:         "trailing content after the object",
			data:         "{}\n{}",
			wantProblems: []wantProblem{{2, "unexpected content after the top-level object"}},
		},
		{
			name:         "truncated file",
			data:         `{"color": {"ink": "#fff"`,
			wantTokens:   []string{"color.ink=#fff@1"},
			wantProblems: []wantProblem{{1, "invalid JSON"}},
		},
		{
			name: "syntax error reports the bad byte's line",
			data: joinLines(
				"{",
				`  "color": {`,
				`    "ink" "#fff"`,
				"  }",
				"}",
			),
			wantProblems: []wantProblem{{3, "invalid JSON"}},
		},
		{
			name: "syntax error at the start of a line",
			data: joinLines(
				"{",
				`  "color": {`,
				`    "ink"`,
				`"#fff"`,
				"  }",
				"}",
			),
			wantProblems: []wantProblem{{4, "invalid JSON"}},
		},
		{
			name:         "raw newline inside a string reports its own line",
			data:         "{\"color\": {\"ink\": \"x\ny\"}}",
			wantProblems: []wantProblem{{1, "invalid JSON"}},
		},
		{
			name:         "newline after a bad literal reports its own line",
			data:         "{\"color\": {\"ink\": tru\n}}",
			wantProblems: []wantProblem{{1, "invalid JSON"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, problems := parseTokens(TokensPath, []byte(tt.data))
			if diff := strings.Join(formatTokens(got), "|"); diff != strings.Join(tt.wantTokens, "|") {
				t.Errorf("tokens = %v, want %v", formatTokens(got), tt.wantTokens)
			}
			for _, p := range problems {
				if p.File != TokensPath {
					t.Errorf("problem file = %q, want %q", p.File, TokensPath)
				}
			}
			checkProblems(t, problems, tt.wantProblems)
		})
	}
}
