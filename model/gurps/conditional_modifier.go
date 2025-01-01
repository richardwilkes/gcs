// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/tid"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
	"github.com/richardwilkes/unison/enums/align"
)

var _ Node[*ConditionalModifier] = &ConditionalModifier{}

// Columns that can be used with the conditional modifier method .CellData()
const (
	ConditionalModifierValueColumn = iota
	ConditionalModifierDescriptionColumn
)

// ConditionalModifier holds data for a reaction or conditional modifier.
type ConditionalModifier struct {
	TID     tid.TID
	From    string
	Amounts []fxp.Int
	Sources []string
}

// NewConditionalModifier creates a new ConditionalModifier.
func NewConditionalModifier(source, from string, amt fxp.Int) *ConditionalModifier {
	return &ConditionalModifier{
		TID:     TIDFromHashedString(kinds.ConditionalModifier, from),
		From:    from,
		Amounts: []fxp.Int{amt},
		Sources: []string{source},
	}
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
	result := txt.NaturalCmp(c.From, other.From, true)
	if result == 0 {
		result = cmp.Compare(c.Total(), other.Total())
	}
	return result
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
	hashhelper.String(h, c.From)
	hashhelper.Num64(h, len(c.Amounts))
	for _, amt := range c.Amounts {
		hashhelper.Num64(h, amt)
	}
	hashhelper.Num64(h, len(c.Sources))
	for _, src := range c.Sources {
		hashhelper.String(h, src)
	}
}

// Clone implements Node.
func (c *ConditionalModifier) Clone(_ LibraryFile, _ DataOwner, _ *ConditionalModifier, preserveID bool) *ConditionalModifier {
	clone := &ConditionalModifier{
		From:    c.From,
		Amounts: slices.Clone(c.Amounts),
		Sources: slices.Clone(c.Sources),
	}
	if preserveID {
		clone.TID = c.TID
	} else {
		clone.TID = tid.MustNewTID(kinds.ConditionalModifier)
	}
	return clone
}

// Kind returns the kind of data.
func (c *ConditionalModifier) Kind() string {
	return i18n.Text("Conditional Modifier")
}

// Container returns true if this is a container.
func (c *ConditionalModifier) Container() bool {
	return false
}

// IsOpen returns true if this node is currently open.
func (c *ConditionalModifier) IsOpen() bool {
	return false
}

// SetOpen sets the current open state for this node.
func (c *ConditionalModifier) SetOpen(_ bool) {
}

// Enabled returns true if this node is enabled.
func (c *ConditionalModifier) Enabled() bool {
	return true
}

// Parent returns the parent.
func (c *ConditionalModifier) Parent() *ConditionalModifier {
	return nil
}

// SetParent sets the parent.
func (c *ConditionalModifier) SetParent(_ *ConditionalModifier) {
}

// HasChildren returns true if this node has children.
func (c *ConditionalModifier) HasChildren() bool {
	return false
}

// NodeChildren returns the children of this node, if any.
func (c *ConditionalModifier) NodeChildren() []*ConditionalModifier {
	return nil
}

// SetChildren sets the children of this node.
func (c *ConditionalModifier) SetChildren(_ []*ConditionalModifier) {
}

func (c *ConditionalModifier) String() string {
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
	switch columnID {
	case ConditionalModifierValueColumn:
		data.Type = cell.Text
		data.Primary = c.Total().StringWithSign()
		data.Alignment = align.End
		var buffer strings.Builder
		for i, amt := range c.Amounts {
			if i != 0 {
				buffer.WriteByte('\n')
			}
			fmt.Fprintf(&buffer, "%s %s", amt.CommaWithSign(), c.Sources[i])
		}
		data.Tooltip = buffer.String()
	case ConditionalModifierDescriptionColumn:
		data.Type = cell.Text
		data.Primary = c.From
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
