package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var varRef = regexp.MustCompile(`var\(--([A-Za-z0-9_-]*)`)

// flattenTokens returns the CSS-style names of the leaf values in a tokens.json document.
func flattenTokens(data []byte) (map[string]bool, error) {
	var root any
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("tokens.json: %w", err)
	}
	obj, ok := root.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("tokens.json: top level must be an object")
	}
	names := map[string]bool{}
	if err := collectNames(obj, nil, names); err != nil {
		return nil, err
	}
	return names, nil
}

func collectNames(obj map[string]any, path []string, names map[string]bool) error {
	for key, value := range obj {
		p := append(append([]string{}, path...), key)
		switch v := value.(type) {
		case map[string]any:
			if err := collectNames(v, p, names); err != nil {
				return err
			}
		case []any:
			return fmt.Errorf("tokens.json: %s is an array; tokens must be objects or values", strings.Join(p, "."))
		default:
			names["--"+strings.Join(p, "-")] = true
		}
	}
	return nil
}

// unresolvedNames returns each var(--name) in source whose name is not in names, in order.
func unresolvedNames(source string, names map[string]bool) []string {
	var out []string
	for _, m := range varRef.FindAllStringSubmatch(source, -1) {
		name := "--" + m[1]
		if !names[name] {
			out = append(out, name)
		}
	}
	return out
}

// hasVarRef reports whether source contains a var(-- reference.
func hasVarRef(source string) bool {
	return strings.Contains(source, "var(--")
}
