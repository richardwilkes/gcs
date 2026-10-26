// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

// CanOpenPageRef returns true if the current selection on the table has a page reference.
func CanOpenPageRef[T gurps.Node[T]](table *unison.Table[*Node[T]]) bool {
	for _, row := range table.SelectedRows(false) {
		var data gurps.CellData
		row.Data().CellData(gurps.PageRefCellAlias, &data)
		if len(ExtractPageReferences(data.Primary)) != 0 {
			return true
		}
	}
	return false
}

// OpenPageRef opens the first page reference on each selected item in the table, stopping if the user cancels.
func OpenPageRef[T gurps.Node[T]](table *unison.Table[*Node[T]]) {
	promptCtx := make(map[string]bool)
	for _, row := range table.SelectedRows(false) {
		var data gurps.CellData
		row.Data().CellData(gurps.PageRefCellAlias, &data)
		for _, one := range ExtractPageReferences(data.Primary) {
			if OpenPageReference(one, data.Secondary, promptCtx) {
				return
			}
			break
		}
	}
}

// OpenEachPageRef opens every page reference on each selected item in the table, stopping if the user cancels.
func OpenEachPageRef[T gurps.Node[T]](table *unison.Table[*Node[T]]) {
	promptCtx := make(map[string]bool)
	for _, row := range table.SelectedRows(false) {
		var data gurps.CellData
		row.Data().CellData(gurps.PageRefCellAlias, &data)
		for _, one := range ExtractPageReferences(data.Primary) {
			if OpenPageReference(one, data.Secondary, promptCtx) {
				return
			}
		}
	}
}

// pageRefLabel is the panel newPageRefLabel returns.
type pageRefLabel struct {
	unison.Panel
	label *unison.Label
}

// makeReadOnly disables the label but leaves the button as the reference makes it, since opening the reference
// changes nothing.
func (p *pageRefLabel) makeReadOnly() {
	p.label.SetEnabled(false)
}

// newPageRefLabel returns a panel holding the label for a page reference field and, beside it, a button in the given
// font that opens the first reference ref returns, looking for what highlight returns on the page. A blank highlight
// opens the page with nothing to look for, so highlight should fall back to the item's name, as the page reference
// cell of a table does. The button is enabled only while there is a reference to open, which refresh brings up to
// date, so call it whenever the reference changes.
func newPageRefLabel(label *unison.Label, font unison.Font, ref, highlight func() string) (wrapper *unison.Panel, refresh func()) {
	p := &pageRefLabel{label: label}
	p.Self = p
	wrapper = p.AsPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: 4,
	})
	wrapper.AddChild(label)
	button := NewSVGButtonForFont(svg.Bookmark, font, -2)
	tabStopForReading(button)
	open := i18n.Text("Open the page reference")
	button.Tooltip = newWrappedTooltip(open)
	button.Accessibility.Name = open
	button.ClickCallback = func() {
		if refs := ExtractPageReferences(ref()); len(refs) != 0 {
			OpenPageReference(refs[0], highlight(), nil)
		}
	}
	refresh = func() { button.SetEnabled(len(ExtractPageReferences(ref())) != 0) }
	refresh()
	wrapper.AddChild(button)
	return wrapper, refresh
}
