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
	"cmp"
	"fmt"
	"hash"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison/enums/align"
)

var _ = assertNode[*ConditionalModifier]

// Columns that can be used with the conditional modifier method .CellData()
const (
	ConditionalModifierValueColumn = iota
	ConditionalModifierDescriptionColumn
)

// ConditionalModifier holds data for a reaction or conditional modifier, or for a group container that holds them. A
// group container has no amounts or sources of its own; From is its name.
type ConditionalModifier struct {
	TID      tid.TID
	From     string
	Amounts  []fxp.Int
	Sources  []string
	Children []*ConditionalModifier
	parent   *ConditionalModifier
}

// NewConditionalModifier creates a new ConditionalModifier that is not filed under a group.
func NewConditionalModifier(source, from string, amt fxp.Int) *ConditionalModifier {
	return newConditionalModifier(TIDFromHashedString(kinds.ConditionalModifier, from), source, from, amt)
}

// newConditionalModifierInGroup creates a ConditionalModifier filed under the named group. The group is mixed into the
// TID, so the same situation appearing under two different groups yields two distinct rows rather than one. An empty
// group takes the NewConditionalModifier path unchanged, so the IDs the exporters write for ungrouped rows are exactly
// what they were before groups existed.
func newConditionalModifierInGroup(source, group, from string, amt fxp.Int) *ConditionalModifier {
	if group == "" {
		return NewConditionalModifier(source, from, amt)
	}
	return newConditionalModifier(TIDFromHashedString(kinds.ConditionalModifier, group, from), source, from, amt)
}

func newConditionalModifier(id tid.TID, source, from string, amt fxp.Int) *ConditionalModifier {
	return &ConditionalModifier{
		TID:     id,
		From:    from,
		Amounts: []fxp.Int{amt},
		Sources: []string{source},
	}
}

// NewConditionalModifierGroup creates the container row that holds the modifiers filed under a group. entityID is the
// ID of the entity the row is built for and namespace is the sheet block key the row belongs to. The TID is derived
// from those and the name rather than generated, because these rows are rebuilt from scratch on every recalculation
// while their disclosure state is stored globally by ID. Mixing in the entity keeps that state per sheet, as it is for
// the other derived rows (attribute separators and hit location sub-tables), so collapsing a group on one character
// does not collapse the same-named group on every other one.
func NewConditionalModifierGroup(entityID tid.TID, namespace, group string) *ConditionalModifier {
	return &ConditionalModifier{
		TID:  TIDFromHashedString(kinds.ConditionalModifierContainer, string(entityID), namespace, group),
		From: group,
	}
}

func conditionalModifierKind(isContainer bool) byte {
	if isContainer {
		return kinds.ConditionalModifierContainer
	}
	return kinds.ConditionalModifier
}

// Add another source.
func (c *ConditionalModifier) Add(source string, amt fxp.Int) {
	c.Amounts = append(c.Amounts, amt)
	c.Sources = append(c.Sources, source)
}

// Total returns the total of all amounts.
func (c *ConditionalModifier) Total() fxp.Int {
	var total fxp.Int
	for _, amt := range c.Amounts {
		total += amt
	}
	return total
}

// Compare returns -1, 0, 1 if this is less than, equal to, or greater than the other.
func (c *ConditionalModifier) Compare(other *ConditionalModifier) int {
	result := xstrings.NaturalCmp(c.From, other.From, true)
	if result == 0 {
		result = cmp.Compare(c.Total(), other.Total())
	}
	return result
}

// compareCondModRows orders the rows within one level of a reaction or conditional modifier table. It is what produces
// the final order for these tables, since their headers do not sort, which is also why it -- rather than Compare --
// reads the general "group containers when sorting" setting: Compare is the plain comparison of two rows and must not
// depend on a preference.
func compareCondModRows(a, b *ConditionalModifier) int {
	if GlobalSettings().General.GroupContainersOnSort {
		if result := containersFirst(a, b); result != 0 {
			return result
		}
	}
	if result := xstrings.NaturalCmp(a.From, b.From, true); result != 0 {
		return result
	}
	// With the setting off, a group and an ungrouped entry can carry the same name; the group goes first, whatever the
	// entry's total, so that the order stays deterministic. This has to come before the totals are compared, since a
	// group's total is always zero and would otherwise place it after a negative entry and before a positive one.
	if result := containersFirst(a, b); result != 0 {
		return result
	}
	return cmp.Compare(a.Total(), b.Total())
}

// containersFirst orders a container ahead of a non-container, and reports two rows of the same kind as equal.
func containersFirst(a, b *ConditionalModifier) int {
	if a.Container() == b.Container() {
		return 0
	}
	if a.Container() {
		return -1
	}
	return 1
}

// GroupName returns the name of the group this modifier is filed under, or an empty string if it isn't in one. A group
// container reports its own name.
func (c *ConditionalModifier) GroupName() string {
	if c.Container() {
		return c.From
	}
	if c.parent != nil {
		return c.parent.From
	}
	return ""
}

// GetSource returns the source of this data.
func (c *ConditionalModifier) GetSource() Source {
	return Source{}
}

// ClearSource clears the source of this data.
func (c *ConditionalModifier) ClearSource() {
}

// SyncWithSource synchronizes this data with the source.
func (c *ConditionalModifier) SyncWithSource() {
}

// ID returns the local ID of this data.
func (c *ConditionalModifier) ID() tid.TID {
	return c.TID
}

// Hash writes this object's contents into the hasher. Note that this only hashes the data that is considered to be
// "source" data, i.e. not expected to be modified by the user after copying from a library.
func (c *ConditionalModifier) Hash(h hash.Hash) {
	xhash.StringWithLen(h, c.From)
	xhash.Bool(h, c.Container())
	xhash.Num64(h, len(c.Amounts))
	for _, amt := range c.Amounts {
		xhash.Num64(h, amt)
	}
	hashStrings(h, c.Sources)
	xhash.Num64(h, len(c.Children))
	for _, child := range c.Children {
		child.Hash(h)
	}
}

// Clone implements Node.
func (c *ConditionalModifier) Clone(from LibraryFile, owner DataOwner, parent *ConditionalModifier, mode CloneMode) *ConditionalModifier {
	clone := &ConditionalModifier{
		From:    c.From,
		Amounts: slices.Clone(c.Amounts),
		Sources: slices.Clone(c.Sources),
		parent:  parent,
	}
	if mode == Copy {
		clone.TID = c.TID
	} else {
		clone.TID = tid.MustNewTID(conditionalModifierKind(c.Container()))
	}
	if c.Container() {
		clone.Children = make([]*ConditionalModifier, 0, len(c.Children))
		for _, child := range c.Children {
			clone.Children = append(clone.Children, child.Clone(from, owner, clone, mode))
		}
	}
	return clone
}

// Kind returns the kind of data.
func (c *ConditionalModifier) Kind() string {
	if c.Container() {
		return i18n.Text("Conditional Modifier Container")
	}
	return i18n.Text("Conditional Modifier")
}

// Container returns true if this is a container.
func (c *ConditionalModifier) Container() bool {
	return tid.IsKind(c.TID, kinds.ConditionalModifierContainer)
}

// IsOpen returns true if this node is currently open.
func (c *ConditionalModifier) IsOpen() bool {
	return IsNodeOpen(c)
}

// SetOpen sets the current open state for this node.
func (c *ConditionalModifier) SetOpen(open bool) {
	SetNodeOpen(c, open)
}

// Enabled returns true if this node is enabled.
func (c *ConditionalModifier) Enabled() bool {
	return true
}

// Parent returns the parent.
func (c *ConditionalModifier) Parent() *ConditionalModifier {
	return c.parent
}

// SetParent sets the parent.
func (c *ConditionalModifier) SetParent(parent *ConditionalModifier) {
	c.parent = parent
}

// HasChildren returns true if this node has children.
func (c *ConditionalModifier) HasChildren() bool {
	return c.Container() && len(c.Children) > 0
}

// NodeChildren returns the children of this node, if any.
func (c *ConditionalModifier) NodeChildren() []*ConditionalModifier {
	return c.Children
}

// SetChildren sets the children of this node.
func (c *ConditionalModifier) SetChildren(children []*ConditionalModifier) {
	c.Children = children
}

func (c *ConditionalModifier) String() string {
	if c.Container() {
		return c.From
	}
	return fmt.Sprintf("%s %s", c.Total().StringWithSign(), c.From)
}

// ConditionalModifiersHeaderData returns the header data information for the given conditional modifier column.
func ConditionalModifiersHeaderData(columnID int) HeaderData {
	var data HeaderData
	switch columnID {
	case ConditionalModifierValueColumn:
		data.Title = i18n.Text("±")
		data.Detail = i18n.Text("Modifier")
		data.Less = fxp.IntLessFromString
	case ConditionalModifierDescriptionColumn:
		data.Title = i18n.Text("Condition")
		data.Primary = true
	}
	return data
}

// ReactionModifiersHeaderData returns the header data information for the given reaction modifier column.
func ReactionModifiersHeaderData(columnID int) HeaderData {
	var data HeaderData
	switch columnID {
	case ConditionalModifierValueColumn:
		data.Title = i18n.Text("±")
		data.Detail = i18n.Text("Modifier")
		data.Less = fxp.IntLessFromString
	case ConditionalModifierDescriptionColumn:
		data.Title = i18n.Text("Reaction")
		data.Primary = true
	}
	return data
}

// CellData returns the cell data information for the given column.
func (c *ConditionalModifier) CellData(columnID int, data *CellData) {
	data.Self = c
	switch columnID {
	case ConditionalModifierValueColumn:
		data.Type = cell.Text
		data.Alignment = align.End
		if c.Container() {
			return // A group doesn't total up its members, since they apply in different situations.
		}
		data.Primary = c.Total().StringWithSign()
		var buffer strings.Builder
		for i, amt := range c.Amounts {
			if i != 0 {
				buffer.WriteByte('\n')
			}
			fmt.Fprintf(&buffer, "%s %s", amt.CommaWithSign(), c.Sources[i])
		}
		data.Tooltip = buffer.String()
	case ConditionalModifierDescriptionColumn:
		data.Type = cell.Markdown
		data.Primary = c.From // the group name, for a container
	case PageRefCellAlias:
		data.Type = cell.PageRef
	}
}

// DataOwner always returns nil.
func (c *ConditionalModifier) DataOwner() DataOwner {
	return nil
}

// SetDataOwner does nothing.
func (c *ConditionalModifier) SetDataOwner(_ DataOwner) {
}

// NameableReplacements returns the replacements to be used with Nameables.
func (c *ConditionalModifier) NameableReplacements() map[string]string {
	return nil
}

// FillWithNameableKeys adds any nameable keys found to the provided map.
func (c *ConditionalModifier) FillWithNameableKeys(_, _ map[string]string) {
}

// ApplyNameableKeys replaces any nameable keys found with the corresponding values in the provided map.
func (c *ConditionalModifier) ApplyNameableKeys(_ map[string]string) {
}
