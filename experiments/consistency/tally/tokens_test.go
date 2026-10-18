package main

import (
	"reflect"
	"sort"
	"testing"
)

func TestFlattenTokensNamesLeaves(t *testing.T) {
	data := []byte(`{"color":{"ink":"#0d1b2e"},"space":{"md":16},"motion":{"standard":"cubic-bezier(0.25, 1, 0.5, 1)"}}`)
	got, err := flattenTokens(data)
	if err != nil {
		t.Fatalf("flattenTokens: %v", err)
	}
	want := []string{"--color-ink", "--motion-standard", "--space-md"}
	if names := sortedSet(got); !reflect.DeepEqual(names, want) {
		t.Fatalf("names = %v, want %v", names, want)
	}
}

func TestFlattenTokensNestedPath(t *testing.T) {
	got, err := flattenTokens([]byte(`{"a":{"b":{"c":1}}}`))
	if err != nil {
		t.Fatalf("flattenTokens: %v", err)
	}
	if want := []string{"--a-b-c"}; !reflect.DeepEqual(sortedSet(got), want) {
		t.Fatalf("names = %v, want %v", sortedSet(got), want)
	}
}

func TestFlattenTokensRejectsArrays(t *testing.T) {
	if _, err := flattenTokens([]byte(`{"x":[1]}`)); err == nil {
		t.Fatal("expected an error for an array value")
	}
}

func TestFlattenTokensRejectsNonObjectTop(t *testing.T) {
	for _, data := range []string{`[1]`, `"text"`, `{"a":`} {
		if _, err := flattenTokens([]byte(data)); err == nil {
			t.Fatalf("expected an error for %s", data)
		}
	}
}

func TestUnresolvedNames(t *testing.T) {
	names := map[string]bool{"--color-ink": true}
	src := "color: var(--color-ink); margin: var(--space-zz); border: var(--space-zz, #fff)"
	want := []string{"--space-zz", "--space-zz"}
	if got := unresolvedNames(src, names); !reflect.DeepEqual(got, want) {
		t.Fatalf("unresolvedNames = %v, want %v", got, want)
	}
}

func TestUnresolvedNamesNoneWhenAllResolve(t *testing.T) {
	names := map[string]bool{"--color-ink": true}
	if got := unresolvedNames("color: var(--color-ink);", names); len(got) != 0 {
		t.Fatalf("unresolvedNames = %v, want none", got)
	}
}

func TestHasVarRef(t *testing.T) {
	cases := map[string]bool{
		"color: var(--color-ink);": true,
		"color: #fff;":             false,
		"var( --color-ink )":       false,
	}
	for src, want := range cases {
		if got := hasVarRef(src); got != want {
			t.Errorf("hasVarRef(%q) = %v, want %v", src, got, want)
		}
	}
}

func sortedSet(m map[string]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
