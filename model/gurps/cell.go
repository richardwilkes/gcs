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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison/enums/align"
)

// PageRefCellAlias is the column alias used to request the page reference cell, if any.
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
	Title  string
	Detail string
	// Name is what a screen reader calls the column when Title is an image key or an abbreviation such as "Pts". It is
	// short, like a title, since it is read ahead of the column's value in every row. Empty when Title reads as drawn.
	Name            string
	Less            func(a, b string) bool
	TitleIsImageKey bool
	Primary         bool
}

// imageHeaderData returns the header data for a column whose title is one of the header image keys above, with detail
// as its tooltip.
func imageHeaderData(imageKey, name, detail string) HeaderData {
	return HeaderData{Title: imageKey, Name: name, Detail: detail, TitleIsImageKey: true}
}

// abbreviatedHeaderData returns the header data for a column whose title is an abbreviation, using what it stands for
// as both its screen reader name and its tooltip.
func abbreviatedHeaderData(title, name string) HeaderData {
	return HeaderData{Title: title, Name: name, Detail: name}
}

func tagsHeaderData() HeaderData {
	return HeaderData{Title: i18n.Text("Tags")}
}

func pageRefHeaderData() HeaderData {
	return imageHeaderData(HeaderBookmark, i18n.Text("Page Reference"), PageRefTooltip())
}

func libSrcHeaderData() HeaderData {
	return imageHeaderData(HeaderDatabase, i18n.Text("Library Source"), LibSrcTooltip())
}

func switchHeaderData() HeaderData {
	return imageHeaderData(HeaderSwitch, i18n.Text("Switched On"), SwitchHeaderTooltip())
}

func enabledHeaderData() HeaderData {
	return imageHeaderData(HeaderCheckmark, i18n.Text("Enabled"), ModifierEnabledTooltip())
}

// CellData holds data for creating a cell's visual representation.
type CellData struct {
	Self      any
	Primary   string
	Secondary string
	Tooltip   string
	// Name is what an assistive technology calls a cell whose content has no words of its own, such as the check box
	// of a toggle or switch column.
	Name string
	// UnsatisfiedReason explains why the node's prerequisites are unmet.
	UnsatisfiedReason string
	// PrereqContradiction explains why a trait whose own prerequisites are met is nonetheless caught in a contradiction
	// among the prerequisites; see Trait.prereqStatus. At most one of it and UnsatisfiedReason is set.
	PrereqContradiction string
	TemplateInfo        string
	ChoiceInfo          string
	InlineTag           string
	Type                cell.Type
	Disabled            bool
	Dim                 bool
	Checked             bool
	Alignment           align.Enum
	// UnresolvedChoice explains which mandatory modifier choices of an item on a sheet have yet to be made.
	UnresolvedChoice string
	// ChoiceRequired is true for a mandatory modifier choice on a sheet that has yet to be made.
	ChoiceRequired bool
	// ForPage is the one input in this struct, set by the caller before invoking a node's CellData method and left
	// untouched by it. It is true when the cell is shown on a sheet, template or loot page rather than in an editor, a
	// library list, or a sort-only request. Sheet-only display preferences, such as the decimal places shown for
	// equipment weights, apply only when it is set.
	ForPage bool
}

func fillTagsCell(data *CellData, tags []string) {
	data.Type = cell.Tags
	data.Primary = CombineTags(tags)
}

// fillPointsCell fills in the cell data for a points column. A range that is still open is explained in the tooltip,
// since a cell showing two numbers where every other row shows one is otherwise a puzzle. Any tooltip the caller has
// already gathered -- the bonuses folded into a skill's or spell's cost, say -- is kept alongside it.
func fillPointsCell(data *CellData, r NumericRange) {
	data.Type = cell.Text
	data.Primary = r.String()
	data.Alignment = align.End
	if r.IsSettled() {
		return
	}
	if data.Tooltip == "" {
		data.Tooltip = PointsRangeTooltip()
	} else {
		data.Tooltip = PointsRangeTooltip() + "\n---\n" + data.Tooltip
	}
}

// fillPageRefCell fills in the cell data for a page reference column: the reference itself, plus the text to look for
// on the page -- highlight when the node has one, otherwise what fallback produces (normally the node's name). fallback
// is only called when it is needed, since resolving a node's text can be costly and this runs for every row on each
// sort and each keystroke of a search.
func fillPageRefCell(data *CellData, pageRef, highlight string, fallback func() string) {
	data.Type = cell.PageRef
	data.Primary = pageRef
	if highlight != "" {
		data.Secondary = highlight
	} else {
		data.Secondary = fallback()
	}
}

// fillLibSrcCell fills in the cell data for a library source column: how node compares to the library data it came
// from, with the details in the tooltip. Only the cell type and alignment are filled in when there is no owner to ask.
func fillLibSrcCell(data *CellData, owner DataOwner, node SrcProvider) {
	data.Type = cell.Text
	data.Alignment = align.Middle
	if xreflect.IsNil(owner) {
		return
	}
	state, _ := owner.SourceMatcher().Match(node)
	data.Primary = state.AltString()
	data.Tooltip = state.String()
	if state != srcstate.Custom {
		data.Tooltip += "\n" + node.GetSource().String()
	}
}

// Values used by ForSort to represent the state of a toggle or switch cell.
const (
	// checkedSortValue is what a toggle or switch that is on sorts and searches as.
	checkedSortValue = "√"
	// switchOffSortValue is what a switch that is off sorts and searches as. A switch column has three states, drawn
	// as a checkmark for on, a dash for off, and nothing for a row with nothing to switch, which is left as an empty
	// text cell. Off therefore needs a value of its own, or sorting would interleave the rows whose switch is off with
	// the ones that have no switch. An en dash mirrors the drawn dash and, like the checkmark, won't be typed into the
	// search field by accident, as an ASCII hyphen would. Sorted ascending, the rows group as no switch, off, then on.
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
