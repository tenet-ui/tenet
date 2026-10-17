package designsystem

import "slices"

// Category returns the tokens of the named category in file order, and whether there is at least one.
func (ds DesignSystem) Category(name string) ([]Token, bool) {
	var out []Token
	for _, t := range ds.Tokens {
		if t.Category == name {
			out = append(out, t)
		}
	}
	return out, len(out) > 0
}

// Categories returns the distinct token category names, sorted.
func (ds DesignSystem) Categories() []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range ds.Tokens {
		if !seen[t.Category] {
			seen[t.Category] = true
			out = append(out, t.Category)
		}
	}
	slices.Sort(out)
	return out
}

// itemsOf returns the items of a kind, or nil for KindTokens and unknown kinds.
func (ds DesignSystem) itemsOf(kind Kind) []Item {
	switch kind {
	case KindComponent:
		return ds.Components
	case KindTemplate:
		return ds.Templates
	case KindFlow:
		return ds.Flows
	}
	return nil
}

// Item returns the item of the given kind with the given id.
func (ds DesignSystem) Item(kind Kind, id string) (Item, bool) {
	for _, it := range ds.itemsOf(kind) {
		if it.ID == id {
			return it, true
		}
	}
	return Item{}, false
}

// ItemIDs returns the ids of the items of a kind, sorted.
func (ds DesignSystem) ItemIDs(kind Kind) []string {
	items := ds.itemsOf(kind)
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.ID)
	}
	slices.Sort(out)
	return out
}

// Linked returns the valid rules with at least one ref to kind:id, each rule once, in rules.md order.
func (ds DesignSystem) Linked(kind Kind, id string) []Rule {
	var out []Rule
	for _, r := range ds.Rules {
		if !r.Valid {
			continue
		}
		for _, ref := range r.AppliesTo {
			if ref.Kind == kind && ref.ID == id {
				out = append(out, r)
				break
			}
		}
	}
	return out
}
