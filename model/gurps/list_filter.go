// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"hash"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/filternode"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/toolbox/v2/xstrings"
)

var (
	_ FilterNode = &FilterGroup{}
	_ FilterNode = &FilterCondition{}
	_ FilterNode = &UnknownFilterNode{}
)

// ListFilter is a named, saved filter for one library list type. Root is never nil once the filter has been
// constructed or loaded.
type ListFilter struct {
	Name string       `json:"name"`
	Root *FilterGroup `json:"root"`
}

// FilterNode is a node in a filter tree: a group of other nodes, a condition on one field, or a node this version of
// GCS doesn't understand but preserves.
type FilterNode interface {
	// NodeType returns the kind of node this is.
	NodeType() filternode.Type
	// ParentGroup returns the owning group, if any.
	ParentGroup() *FilterGroup
	// SetParentGroup sets the owning group.
	SetParentGroup(parent *FilterGroup)
	// Clone creates a deep copy of this node, owned by parent.
	Clone(parent *FilterGroup) FilterNode
	// Hash writes this object's contents into the hasher.
	Hash(h hash.Hash)
}

// FilterNodes holds a list of filter nodes.
type FilterNodes []FilterNode

// FilterGroup combines the results of its children, requiring either all of them or any one of them to match. A
// group with no children matches everything, whichever way it combines. Not inverts the result; it is stored as the
// negative so that the common case, a group that has to match, is omitted from the JSON.
type FilterGroup struct {
	Parent   *FilterGroup    `json:"-"`
	Type     filternode.Type `json:"type"`
	Not      bool            `json:"not,omitzero"`
	All      bool            `json:"all"`
	Children FilterNodes     `json:"children,omitempty"`
}

// FilterCondition compares one field of a node against a criteria. Which of the criteria is consulted depends on the
// kind of the field named by Field; the others are left at their zero values and omitted from the JSON. Not inverts
// the result, and is stored as the negative for the same reason as on a FilterGroup.
type FilterCondition struct {
	Parent *FilterGroup    `json:"-"`
	Type   filternode.Type `json:"type"`
	Not    bool            `json:"not,omitzero"`
	Field  string          `json:"field"`
	Text   criteria.Text   `json:"text,omitzero"`
	Number criteria.Number `json:"number,omitzero"`
	Weight criteria.Weight `json:"weight,omitzero"`
}

// UnknownFilterNode holds a filter node whose type this version of GCS doesn't recognize. It never matches, and its
// original data is written back out unchanged so that a newer version can still use it.
type UnknownFilterNode struct {
	// Parent is the owning group.
	Parent *FilterGroup
	// Kind is the unrecognized value that was found in the "type" field.
	Kind string
	// Data is the original JSON for the node, exactly as it was read.
	Data jsontext.Value
}

// NewListFilter creates a new, empty filter with the given name.
func NewListFilter(name string) *ListFilter {
	f := &ListFilter{Name: name}
	f.EnsureValidity()
	return f
}

// Clone creates a deep copy of this filter.
func (f *ListFilter) Clone() *ListFilter {
	clone := &ListFilter{Name: f.Name}
	if f.Root != nil {
		clone.Root = f.Root.CloneAsFilterGroup(nil)
	}
	clone.EnsureValidity()
	return clone
}

// EnsureValidity trims the name, makes sure there is a root group, and walks the tree to set each node's type and
// parent.
func (f *ListFilter) EnsureValidity() {
	f.Name = strings.TrimSpace(f.Name)
	if f.Root == nil {
		f.Root = NewFilterGroup(nil)
	}
	ensureFilterNodeValidity(f.Root, nil)
}

// ensureFilterNodeValidity sets the node's parent and type, drops any nil children, and recurses.
func ensureFilterNodeValidity(node FilterNode, parent *FilterGroup) {
	node.SetParentGroup(parent)
	switch n := node.(type) {
	case *FilterGroup:
		n.Type = filternode.Group
		n.Children = slices.DeleteFunc(n.Children, func(child FilterNode) bool { return xreflect.IsNil(child) })
		for _, child := range n.Children {
			ensureFilterNodeValidity(child, n)
		}
	case *FilterCondition:
		n.Type = filternode.Condition
	}
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (f *ListFilter) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	type listFilterData ListFilter // Avoids recursion into this method.
	var data listFilterData
	if err := json.UnmarshalDecode(dec, &data); err != nil {
		return err
	}
	*f = ListFilter(data)
	f.EnsureValidity()
	return nil
}

// String returns the filter's name.
func (f *ListFilter) String() string {
	return f.Name
}

// Hash writes this object's contents into the hasher.
func (f *ListFilter) Hash(h hash.Hash) {
	if f == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.StringWithLen(h, f.Name)
	f.Root.Hash(h)
}

// NewFilterGroup creates a new, empty group that requires all of its children to match.
func NewFilterGroup(parent *FilterGroup) *FilterGroup {
	return &FilterGroup{
		Parent: parent,
		Type:   filternode.Group,
		All:    true,
	}
}

// NodeType implements FilterNode.
func (g *FilterGroup) NodeType() filternode.Type {
	return filternode.Group
}

// ParentGroup implements FilterNode.
func (g *FilterGroup) ParentGroup() *FilterGroup {
	if g == nil {
		return nil
	}
	return g.Parent
}

// SetParentGroup implements FilterNode.
func (g *FilterGroup) SetParentGroup(parent *FilterGroup) {
	g.Parent = parent
}

// Clone implements FilterNode.
func (g *FilterGroup) Clone(parent *FilterGroup) FilterNode {
	return g.CloneAsFilterGroup(parent)
}

// CloneAsFilterGroup creates a deep copy of this group, owned by parent.
func (g *FilterGroup) CloneAsFilterGroup(parent *FilterGroup) *FilterGroup {
	clone := *g
	clone.Parent = parent
	clone.Children = make(FilterNodes, len(g.Children))
	for i, child := range g.Children {
		clone.Children[i] = child.Clone(&clone)
	}
	return &clone
}

// Hash implements FilterNode.
func (g *FilterGroup) Hash(h hash.Hash) {
	if g == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, filternode.Group)
	xhash.Bool(h, g.Not)
	xhash.Bool(h, g.All)
	hashList(h, g.Children)
}

// NewFilterCondition creates a new condition on the field with the given key, with every criteria at its zero value.
// For a text, list, number or weight field that means the condition accepts anything; for a yes/no field, which has
// no criteria, it means the condition is satisfied when the value is true.
func NewFilterCondition(parent *FilterGroup, fieldKey string) *FilterCondition {
	return &FilterCondition{
		Parent: parent,
		Type:   filternode.Condition,
		Field:  fieldKey,
	}
}

// NodeType implements FilterNode.
func (c *FilterCondition) NodeType() filternode.Type {
	return filternode.Condition
}

// ParentGroup implements FilterNode.
func (c *FilterCondition) ParentGroup() *FilterGroup {
	return c.Parent
}

// SetParentGroup implements FilterNode.
func (c *FilterCondition) SetParentGroup(parent *FilterGroup) {
	c.Parent = parent
}

// Clone implements FilterNode.
func (c *FilterCondition) Clone(parent *FilterGroup) FilterNode {
	clone := *c
	clone.Parent = parent
	return &clone
}

// Hash implements FilterNode.
func (c *FilterCondition) Hash(h hash.Hash) {
	if c == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, filternode.Condition)
	xhash.Bool(h, c.Not)
	xhash.StringWithLen(h, c.Field)
	c.Text.Hash(h)
	c.Number.Hash(h)
	c.Weight.Hash(h)
}

// NewUnknownFilterNode creates a new UnknownFilterNode holding a copy of the passed-in data.
func NewUnknownFilterNode(kind string, data jsontext.Value) *UnknownFilterNode {
	return &UnknownFilterNode{
		Kind: kind,
		Data: slices.Clone(data),
	}
}

// NodeType implements FilterNode.
func (n *UnknownFilterNode) NodeType() filternode.Type {
	return filternode.Unknown
}

// ParentGroup implements FilterNode.
func (n *UnknownFilterNode) ParentGroup() *FilterGroup {
	return n.Parent
}

// SetParentGroup implements FilterNode.
func (n *UnknownFilterNode) SetParentGroup(parent *FilterGroup) {
	n.Parent = parent
}

// Clone implements FilterNode.
func (n *UnknownFilterNode) Clone(parent *FilterGroup) FilterNode {
	clone := NewUnknownFilterNode(n.Kind, n.Data)
	clone.Parent = parent
	return clone
}

// MarshalJSONTo implements json.MarshalerTo. The original data is written back out as-is.
func (n *UnknownFilterNode) MarshalJSONTo(enc *jsontext.Encoder) error {
	return enc.WriteValue(n.Data)
}

// Hash implements FilterNode.
func (n *UnknownFilterNode) Hash(h hash.Hash) {
	if n == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, filternode.Unknown)
	xhash.StringWithLen(h, n.Kind)
	xhash.BytesWithLen(h, n.Data)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (n *FilterNodes) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	list, err := unmarshalTypedList(dec, filternode.ExtractKnownType, allocFilterNode,
		func(kind string, raw jsontext.Value) FilterNode { return NewUnknownFilterNode(kind, raw) })
	if err != nil {
		return err
	}
	*n = list
	return nil
}

// allocFilterNode returns an empty FilterNode of the concrete type that represents nodeType, or nil if there is none.
func allocFilterNode(nodeType filternode.Type) FilterNode {
	switch nodeType {
	case filternode.Group:
		return &FilterGroup{}
	case filternode.Condition:
		return &FilterCondition{}
	default:
		return nil
	}
}

// ListFilterKeyForExtension returns the key the saved filters for library list files with the given extension are
// stored under, or an empty string if the extension isn't one of the library list types. The key is the extension
// without its leading dot, in lower case.
func ListFilterKeyForExtension(ext string) string {
	ext = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(ext), "."))
	if slices.Contains(ListFilterKeys(), ext) {
		return ext
	}
	return ""
}

// ListFilterKeys returns the keys of every library list type that can have saved filters, in sorted order.
func ListFilterKeys() []string {
	keys := []string{
		strings.TrimPrefix(TraitsExt, "."),
		strings.TrimPrefix(TraitModifiersExt, "."),
		strings.TrimPrefix(SkillsExt, "."),
		strings.TrimPrefix(SpellsExt, "."),
		strings.TrimPrefix(EquipmentExt, "."),
		strings.TrimPrefix(EquipmentModifiersExt, "."),
		strings.TrimPrefix(NotesExt, "."),
	}
	slices.Sort(keys)
	return keys
}

// ListFiltersFor returns the saved filters for the list type with the given key, sorted by name. The slice is a copy,
// but the filters in it are the live ones.
func (s *Settings) ListFiltersFor(key string) []*ListFilter {
	return slices.Clone(s.ListFilters[key])
}

// SetListFiltersFor replaces the saved filters for the list type with the given key. Nil entries are dropped, the
// rest are sorted by name, and the key is removed when nothing remains.
func (s *Settings) SetListFiltersFor(key string, filters []*ListFilter) {
	filters = slices.DeleteFunc(slices.Clone(filters), func(f *ListFilter) bool { return f == nil })
	if len(filters) == 0 {
		delete(s.ListFilters, key)
		return
	}
	sortListFilters(filters)
	if s.ListFilters == nil {
		s.ListFilters = make(map[string][]*ListFilter)
	}
	s.ListFilters[key] = filters
}

// AddListFilter adds a saved filter for the list type with the given key.
func (s *Settings) AddListFilter(key string, f *ListFilter) {
	s.SetListFiltersFor(key, append(s.ListFiltersFor(key), f))
}

// ResortListFilters puts the saved filters for the list type with the given key back into name order, which is
// needed after a filter has been renamed in place.
func (s *Settings) ResortListFilters(key string) {
	s.SetListFiltersFor(key, s.ListFilters[key])
}

// RemoveListFilter removes the saved filter f for the list type with the given key.
func (s *Settings) RemoveListFilter(key string, f *ListFilter) {
	filters := s.ListFiltersFor(key)
	if i := slices.Index(filters, f); i != -1 {
		s.SetListFiltersFor(key, slices.Delete(filters, i, i+1))
	}
}

// ListFilterNameInUse returns true if a saved filter for the list type with the given key already has the name,
// ignoring case and surrounding whitespace. The filter passed as except, if any, doesn't count, so that a filter being
// edited may keep its own name.
func (s *Settings) ListFilterNameInUse(key, name string, except *ListFilter) bool {
	name = strings.TrimSpace(name)
	for _, f := range s.ListFilters[key] {
		if f != nil && f != except && strings.EqualFold(strings.TrimSpace(f.Name), name) {
			return true
		}
	}
	return false
}

// sortListFilters sorts the filters by name, ignoring case.
func sortListFilters(filters []*ListFilter) {
	slices.SortStableFunc(filters, func(a, b *ListFilter) int { return xstrings.NaturalCmp(a.Name, b.Name, true) })
}

// ensureListFiltersValidity drops the saved filters that can't be used -- nil entries, filters without a name, and
// later filters that duplicate an earlier one's name -- brings the rest into line, and keeps each list sorted. Keys
// that aren't recognized are kept, since a newer version of GCS may have written them.
func (s *Settings) ensureListFiltersValidity() {
	for key, filters := range s.ListFilters {
		seen := make(map[string]bool)
		filters = slices.DeleteFunc(filters, func(f *ListFilter) bool {
			if f == nil {
				return true
			}
			f.EnsureValidity()
			lowered := strings.ToLower(f.Name)
			if lowered == "" || seen[lowered] {
				return true
			}
			seen[lowered] = true
			return false
		})
		s.SetListFiltersFor(key, filters)
	}
}
