package designsystem

import (
	"reflect"
	"testing"
)

// sampleSystem is a hand-built pack. Rule a is valid and linked twice to button; b is invalid; c is valid and linked to color and button.
func sampleSystem() DesignSystem {
	return DesignSystem{
		Tokens: []Token{
			{Category: "space", Name: "unit", Value: "4"},
			{Category: "color", Name: "ink", Value: "#1d2433"},
			{Category: "color", Name: "accent", Value: "#2f6fed"},
		},
		Rules: []Rule{
			{ID: "a", Valid: true, AppliesTo: []Ref{{Kind: KindComponent, ID: "button"}, {Kind: KindComponent, ID: "button"}}},
			{ID: "b", Valid: false, AppliesTo: []Ref{{Kind: KindComponent, ID: "button"}}},
			{ID: "c", Valid: true, AppliesTo: []Ref{{Kind: KindTokens, ID: "color"}, {Kind: KindComponent, ID: "button"}}},
		},
		Components: []Item{{ID: "card", Title: "Card"}, {ID: "button", Title: "Button"}},
		Flows:      []Item{{ID: "first-run", Title: "First run"}},
	}
}

func ruleIDs(rs []Rule) []string {
	out := []string{}
	for _, r := range rs {
		out = append(out, r.ID)
	}
	return out
}

func TestLookups(t *testing.T) {
	ds := sampleSystem()

	t.Run("Category keeps file order and reports missing", func(t *testing.T) {
		got, ok := ds.Category("color")
		if !ok || len(got) != 2 || got[0].Name != "ink" || got[1].Name != "accent" {
			t.Errorf("Category(color) = %v, %t", got, ok)
		}
		if got, ok := ds.Category("motion"); ok || len(got) != 0 {
			t.Errorf("Category(motion) = %v, %t, want none", got, ok)
		}
	})

	t.Run("Categories sorted and unique", func(t *testing.T) {
		if got, want := ds.Categories(), []string{"color", "space"}; !reflect.DeepEqual(got, want) {
			t.Errorf("Categories() = %v, want %v", got, want)
		}
	})

	t.Run("Item finds by kind and id", func(t *testing.T) {
		if it, ok := ds.Item(KindComponent, "button"); !ok || it.Title != "Button" {
			t.Errorf("Item(component, button) = %v, %t", it, ok)
		}
		if _, ok := ds.Item(KindFlow, "button"); ok {
			t.Error("Item(flow, button) found a component")
		}
		if _, ok := ds.Item(KindTokens, "color"); ok {
			t.Error("Item(tokens, color) found something")
		}
	})

	t.Run("ItemIDs sorted per kind", func(t *testing.T) {
		if got, want := ds.ItemIDs(KindComponent), []string{"button", "card"}; !reflect.DeepEqual(got, want) {
			t.Errorf("ItemIDs(component) = %v, want %v", got, want)
		}
		if got := ds.ItemIDs(KindTokens); len(got) != 0 {
			t.Errorf("ItemIDs(tokens) = %v, want none", got)
		}
	})

	t.Run("Linked returns valid rules once, in rules order", func(t *testing.T) {
		if got, want := ruleIDs(ds.Linked(KindComponent, "button")), []string{"a", "c"}; !reflect.DeepEqual(got, want) {
			t.Errorf("Linked(component, button) = %v, want %v", got, want)
		}
		if got, want := ruleIDs(ds.Linked(KindTokens, "color")), []string{"c"}; !reflect.DeepEqual(got, want) {
			t.Errorf("Linked(tokens, color) = %v, want %v", got, want)
		}
		if got := ds.Linked(KindComponent, "card"); len(got) != 0 {
			t.Errorf("Linked(component, card) = %v, want none", ruleIDs(got))
		}
	})
}

func TestFixtureLinks(t *testing.T) {
	pack, problems := Load("../../testdata/project")
	if len(problems) != 0 {
		t.Fatalf("fixture has problems: %v", problems)
	}
	if got, want := ruleIDs(pack.Linked(KindComponent, "button")), []string{"color.tokens-only"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Linked(component, button) = %v, want %v", got, want)
	}
	if got, want := ruleIDs(pack.Linked(KindTokens, "space")), []string{"spacing.unit-multiples"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Linked(tokens, space) = %v, want %v", got, want)
	}
	if got := pack.Linked(KindFlow, "first-run"); len(got) != 0 {
		t.Errorf("Linked(flow, first-run) = %v, want none", ruleIDs(got))
	}
	if got, ok := pack.Category("color"); !ok || len(got) != 3 {
		t.Errorf("Category(color) = %d tokens (found %t), want 3", len(got), ok)
	}
}
