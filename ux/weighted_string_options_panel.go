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
	"fmt"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

// weightedStringOptionsPanel edits a weighted string list: a header with an add button (and an import button, when the
// spec provides an importer) and a button that removes the selected rows, and, when the list has any entries, a
// bordered container holding a row per entry. The container is omitted for an empty list rather than drawn empty, so
// its children are always exactly the rows, which is what drag reordering relies on.
//
// Rows are selected by clicking them, as the rows of a table are: a plain click selects the row alone, a click with
// the OS's menu command key adds the row to the selection or takes it out, and a shift-click selects every row from
// the anchor -- the row last clicked without shift -- to the clicked one. The selection is what the remove button in
// the header acts on, so that a long list, such as imported training names, can be pruned without a click per row.
type weightedStringOptionsPanel struct {
	unison.Panel
	dockable structuralEditor
	spec     *weightedStringListSpec
	rows     *unison.Panel // nil when the list is empty
	// selected holds the options whose rows are selected, and anchor is where a shift-click extends the selection
	// from. Both belong to the rows as built, and are let go of along with them when a structural edit rebuilds the
	// content.
	selected             map[*gurps.WeightedStringOption]bool
	anchor               *gurps.WeightedStringOption
	removeSelectedButton *unison.Button
}

func newWeightedStringOptionsPanel(d structuralEditor, spec *weightedStringListSpec) *weightedStringOptionsPanel {
	p := &weightedStringOptionsPanel{
		dockable: d,
		spec:     spec,
		selected: make(map[*gurps.WeightedStringOption]bool),
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{Columns: 1, VSpacing: unison.StdVSpacing})
	p.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, VAlign: align.Start, HGrab: true})
	addButton := unison.NewSVGButton(unison.CircledAddSVG)
	addButton.Tooltip = newWrappedTooltip(spec.addTooltip)
	addButton.ClickCallback = p.addOption
	buttons := []*unison.Button{addButton}
	if spec.importer != nil {
		importButton := unison.NewSVGButton(svg.DownToBracket)
		importButton.Tooltip = newWrappedTooltip(spec.importTooltip)
		importButton.ClickCallback = spec.importer
		buttons = append(buttons, importButton)
	}
	p.removeSelectedButton = unison.NewSVGButton(unison.TrashSVG)
	p.removeSelectedButton.Tooltip = newWrappedTooltip(spec.removeSelectedTooltip)
	p.removeSelectedButton.ClickCallback = p.removeSelected
	p.removeSelectedButton.SetEnabled(false)
	buttons = append(buttons, p.removeSelectedButton)
	p.AddChild(newEditorSectionHeader(spec.title, spec.tooltip, buttons...))
	if len(*spec.list) != 0 {
		p.rows = unison.NewPanel()
		p.rows.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.NewUniformInsets(1), false))
		p.rows.SetLayout(&unison.FlexLayout{Columns: 1})
		p.rows.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
		for _, option := range *spec.list {
			p.rows.AddChild(newWeightedStringOptionPanel(p, option))
		}
		p.AddChild(p.rows)
	}
	return p
}

// addOption appends an option with a weight of 1 and moves the focus to its value.
func (p *weightedStringOptionsPanel) addOption() {
	option := &gurps.WeightedStringOption{Weight: 1, KeyPrefix: p.dockable.targetManager().NextPrefix()}
	p.dockable.editStructure(p.spec.addTitle, func() { *p.spec.list = append(*p.spec.list, option) },
		option.KeyPrefix+"value")
}

// removeOption removes the option from the list. Removing the last one is allowed; what an empty list means is up to
// the model that owns it.
func (p *weightedStringOptionsPanel) removeOption(option *gurps.WeightedStringOption) {
	i := slices.Index(*p.spec.list, option)
	if i == -1 {
		return
	}
	p.dockable.editStructure(p.spec.removeTitle, func() { *p.spec.list = slices.Delete(*p.spec.list, i, i+1) }, "")
}

// selectRow applies a click on the row of the option to the selection; see the type's comment for what each kind of
// click does. A shift-click with nothing to extend from is a plain click.
func (p *weightedStringOptionsPanel) selectRow(option *gurps.WeightedStringOption, mods mod.Modifiers) {
	list := *p.spec.list
	to := slices.Index(list, option)
	if to == -1 {
		return
	}
	from := -1
	if p.anchor != nil {
		from = slices.Index(list, p.anchor)
	}
	switch {
	case mods.ShiftDown() && from != -1:
		if from > to {
			from, to = to, from
		}
		for _, one := range list[from : to+1] {
			p.selected[one] = true
		}
	case mods.OSMenuCommandDown():
		if p.selected[option] {
			delete(p.selected, option)
		} else {
			p.selected[option] = true
		}
		p.anchor = option
	default:
		clear(p.selected)
		p.selected[option] = true
		p.anchor = option
	}
	p.removeSelectedButton.SetEnabled(len(p.selected) != 0)
	if p.rows != nil {
		p.rows.MarkForRedraw()
	}
}

// isSelected reports whether the row of the option is selected.
func (p *weightedStringOptionsPanel) isSelected(option *gurps.WeightedStringOption) bool {
	return p.selected[option]
}

// selectedOptions returns the selected options in list order.
func (p *weightedStringOptionsPanel) selectedOptions() []*gurps.WeightedStringOption {
	var result []*gurps.WeightedStringOption
	for _, one := range *p.spec.list {
		if p.selected[one] {
			result = append(result, one)
		}
	}
	return result
}

// removeSelected removes every selected option from the list in a single undo edit. Nothing is posted when none is
// selected.
func (p *weightedStringOptionsPanel) removeSelected() {
	if len(p.selected) == 0 {
		return
	}
	p.dockable.editStructure(p.spec.removeSelectedTitle, func() {
		*p.spec.list = slices.DeleteFunc(*p.spec.list, func(one *gurps.WeightedStringOption) bool {
			return p.selected[one]
		})
	}, "")
}

// rowSelectionTooltip explains how rows are selected, for the parts of a row that are not a field or a button.
func rowSelectionTooltip() string {
	return fmt.Sprintf(i18n.Text("Click to select this row, shift-click to select every row from the last one clicked to this one, or %s-click to add this row to the selection or take it out. The selected rows are removed together by the button beside the list's title."),
		strings.TrimSuffix(mod.OSMenuCommand().String(), "+"))
}

// weightedStringOptionPanel is one row of a weighted string list: a drag handle, a remove button, the weight and the
// value. A click on the row -- on its handle, or anywhere that is not one of its widgets -- selects it; see
// weightedStringOptionsPanel.
type weightedStringOptionPanel struct {
	unison.Panel
	list         *weightedStringOptionsPanel
	option       *gurps.WeightedStringOption
	deleteButton *unison.Button
}

func newWeightedStringOptionPanel(list *weightedStringOptionsPanel, option *gurps.WeightedStringOption) *weightedStringOptionPanel {
	p := &weightedStringOptionPanel{
		list:   list,
		option: option,
	}
	p.Self = p
	configureEditorRow(p.AsPanel(), 4)
	d := list.dockable
	spec := list.spec
	p.Tooltip = newWrappedTooltip(rowSelectionTooltip())
	p.MouseDownCallback = func(_ geom.Point, button, _ int, mods mod.Modifiers) bool {
		if button != unison.ButtonLeft {
			return false
		}
		list.selectRow(option, mods)
		return true
	}
	// A selected row is drawn in the selection color in place of the alternating background configureEditorRow gave
	// it.
	alternating := p.DrawCallback
	p.DrawCallback = func(gc *unison.Canvas, rect geom.Rect) {
		if list.isSelected(option) {
			gc.DrawRect(rect, unison.ThemeFocus.Paint(gc, rect, paintstyle.Fill))
			return
		}
		alternating(gc, rect)
	}
	handle := NewDragHandle(editorRowDragKey, &editorRowDragData{
		editor: d,
		row:    p.AsPanel(),
		title:  spec.dragTitle,
		move:   func(to int) bool { return moveEntry(spec.list, slices.Index(*spec.list, option), to) },
	})
	// Pressing the handle selects the row as well, since the handle is the one part of a row that is always there to
	// click, whatever the row's fields hold.
	handleMouseDown := handle.MouseDownCallback
	handle.MouseDownCallback = func(where geom.Point, button, clickCount int, mods mod.Modifiers) bool {
		if button == unison.ButtonLeft {
			list.selectRow(option, mods)
		}
		return handleMouseDown(where, button, clickCount, mods)
	}
	p.AddChild(handle)

	p.deleteButton = unison.NewSVGButton(unison.TrashSVG)
	p.deleteButton.ClickCallback = func() { list.removeOption(option) }
	p.deleteButton.Tooltip = newWrappedTooltip(spec.removeTooltip)
	p.AddChild(p.deleteButton)

	mgr := d.targetManager()
	weightField := NewIntegerField(mgr, option.KeyPrefix+"weight", i18n.Text("Weight"),
		func() int { return option.Weight },
		func(v int) { option.Weight = v }, 0, 9999, false, false)
	weightField.SetBaseTooltip(newWrappedTooltip(i18n.Text("The relative likelihood of this option being chosen. Only the ratio between the weights within the list matters.")))
	p.AddChild(weightField)

	valueField := NewStringField(mgr, option.KeyPrefix+"value", i18n.Text("Value"),
		func() string { return option.Value },
		func(s string) { option.Value = s })
	valueField.SetMinimumTextWidthUsing(prototypeMinIDWidth)
	valueField.Tooltip = newWrappedTooltip(spec.valueTooltip)
	p.AddChild(valueField)
	return p
}
