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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/unison/enums/align"
)

// PageRefCellAlias is used an alias to request the page reference cell, if any.
const PageRefCellAlias = -10

// These constants are used to specify images to use in column headers.
const (
	HeaderCheckmark     = "checkmark"
	HeaderCoins         = "coins"
	HeaderWeight        = "weight"
	HeaderBookmark      = "bookmark"
	HeaderDatabase      = "database"
	HeaderStackedCoins  = "stacked-coins"
	HeaderStackedWeight = "stacked-weight"
	HeaderSwitch        = "switch"
)

// HeaderData holds data for creating a column header's visual representation.
type HeaderData struct {
	Title           string
	Detail          string
	Less            func(a, b string) bool
	TitleIsImageKey bool
	Primary         bool
}

// CellData holds data for creating a cell's visual representation.
type CellData struct {
	Self              any
	Primary           string
	Secondary         string
	Tooltip           string
	UnsatisfiedReason string
	TemplateInfo      string
	InlineTag         string
	Type              cell.Type
	Disabled          bool
	Dim               bool
	Checked           bool
	Alignment         align.Enum
	// ForPage is the one input in this struct: the caller sets it before invoking a node's CellData method, and it
	// is left untouched by that method. It is true when the cell is being displayed on a sheet, template or loot
	// page, as opposed to an editor, a library list, or a request made only to sort the rows. Display preferences
	// that belong to the sheet alone, such as the number of decimal places shown for equipment weights, are applied
	// only when it is set, so that the same node renders exactly elsewhere.
	ForPage bool
}

// Values used by ForSort to represent the state of a toggle or switch cell.
const (
	// checkedSortValue is what a toggle or switch that is on sorts and searches as.
	checkedSortValue = "√"
	// switchOffSortValue is what a switch that is off sorts and searches as. A switch column has three states, not two:
	// on, off, and "nothing to switch" -- and the last of those isn't a switch cell at all, since rows with nothing to
	// switch are left as an empty text cell. That is also how the cells are drawn: a checkmark for on, a dash for off,
	// and nothing at all for a row with no switch. The off state therefore needs a value of its own, or sorting by the
	// column would interleave the rows whose switch is off with the ones that have no switch, hiding the very
	// distinction the cells are drawn to show. An en dash is used, since it mirrors the dash the cell is drawn with
	// and, like the checkmark, won't be typed into the search field by accident, the way an ASCII hyphen would be.
	// Sorted ascending, this groups the rows as no switch, then off, then on.
	switchOffSortValue = "–"
)

// ForSort returns a string that can be used to sort or search against for this data.
func (c *CellData) ForSort() string {
	switch c.Type {
	case cell.Text:
		if c.Secondary != "" {
			return c.Primary + "\n" + c.Secondary
		}
		return c.Primary
	case cell.Toggle:
		if c.Checked {
			return checkedSortValue
		}
	case cell.Switch:
		if c.Checked {
			return checkedSortValue
		}
		return switchOffSortValue
	case cell.PageRef, cell.Tags, cell.Markdown:
		return c.Primary
	}
	return ""
}
