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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// nameGeneratorsPanel edits the ordered list of name generators of an options block: a header with an add button and,
// when the list has any entries, a bordered container holding a row per generator. As with the weighted string lists,
// the container is omitted for an empty list so that its children are always exactly the rows.
type nameGeneratorsPanel struct {
	unison.Panel
	dockable *ancestryEditorDockable
	options  *gurps.AncestryOptions
	rows     *unison.Panel // nil when the list is empty
}

func newNameGeneratorsPanel(d *ancestryEditorDockable, options *gurps.AncestryOptions) *nameGeneratorsPanel {
	p := &nameGeneratorsPanel{
		dockable: d,
		options:  options,
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{Columns: 1, VSpacing: unison.StdVSpacing})
	p.SetLayoutData(&unison.FlexLayoutData{HSpan: 2, HAlign: align.Fill, HGrab: true})
	addButton := unison.NewSVGButton(unison.CircledAddSVG)
	addButton.Tooltip = newWrappedTooltip(i18n.Text("Add name generator"))
	addButton.ClickCallback = p.addGenerator
	p.AddChild(newEditorSectionHeader(i18n.Text("Name Generators"),
		i18n.Text("The name generators used to produce a random name. The output of each generator is joined, in order, with a space. A gender with an empty list uses the common options."),
		addButton))
	if len(options.NameGenerators) != 0 {
		p.rows = unison.NewPanel()
		p.rows.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.NewUniformInsets(1), false))
		p.rows.SetLayout(&unison.FlexLayout{Columns: 1})
		p.rows.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
		for i := range options.NameGenerators {
			p.rows.AddChild(newNameGeneratorRefPanel(p, i))
		}
		p.AddChild(p.rows)
	}
	return p
}

// addGenerator appends a generator, defaulting to the first available one, and moves the focus to its popup.
func (p *nameGeneratorsPanel) addGenerator() {
	d := p.dockable
	var name string
	if len(d.nameGeneratorChoices) != 0 {
		name = d.nameGeneratorChoices[0]
	}
	options := p.options
	d.editStructure(i18n.Text("Add Name Generator"),
		func() { options.NameGenerators = append(options.NameGenerators, name) },
		nameGeneratorKey(options, len(options.NameGenerators)))
}

// removeGenerator removes the generator at the index. Removing the last one is allowed: an empty list falls back to the
// common options.
func (p *nameGeneratorsPanel) removeGenerator(index int) {
	options := p.options
	if index < 0 || index >= len(options.NameGenerators) {
		return
	}
	p.dockable.editStructure(i18n.Text("Remove Name Generator"), func() {
		options.NameGenerators = slices.Delete(options.NameGenerators, index, index+1)
	}, "")
}

// nameGeneratorKey returns the reference key of the popup for the name generator at the index within the options
// block. Keys are index-based, which is safe because every structural edit rebuilds the panels.
func nameGeneratorKey(options *gurps.AncestryOptions, index int) string {
	return fmt.Sprintf("%snamegen:%d", options.KeyPrefix, index)
}

// nameGeneratorPopupItems returns the items to offer in a name generator popup: the available choices, preceded by
// current when it is not among them, so that a generator whose .names file is missing from every library remains
// visible and selectable rather than silently showing as something else. The result is always a fresh slice, so the
// caller's cached choices are never appended into.
func nameGeneratorPopupItems(choices []string, current string) []string {
	items := make([]string, 0, len(choices)+1)
	if !slices.Contains(choices, current) {
		items = append(items, current)
	}
	return append(items, choices...)
}

// nameGeneratorRefPanel is one row of the name generator list: a drag handle, a remove button, a popup of the
// generators to choose from, and a button that opens the chosen generator in the name generator editor.
type nameGeneratorRefPanel struct {
	unison.Panel
	list       *nameGeneratorsPanel
	index      int
	popup      *Popup[string]
	editButton *unison.Button
}

func newNameGeneratorRefPanel(list *nameGeneratorsPanel, index int) *nameGeneratorRefPanel {
	p := &nameGeneratorRefPanel{
		list:  list,
		index: index,
	}
	p.Self = p
	configureEditorRow(p.AsPanel(), 4)
	d := list.dockable
	options := list.options
	p.AddChild(NewDragHandle(editorRowDragKey, &editorRowDragData{
		editor: d,
		row:    p.AsPanel(),
		title:  i18n.Text("Name Generator Drag"),
		move:   func(to int) bool { return moveEntry(&options.NameGenerators, index, to) },
	}))

	deleteButton := unison.NewSVGButton(unison.TrashSVG)
	deleteButton.ClickCallback = func() { list.removeGenerator(index) }
	deleteButton.Tooltip = newWrappedTooltip(i18n.Text("Remove name generator"))
	p.AddChild(deleteButton)

	// The setter and getter check the index, since a popup's callbacks can outlive the entry they were built for when
	// an undo shortens the list before the panels are rebuilt.
	p.popup = NewPopup(d.targetMgr, nameGeneratorKey(options, index), i18n.Text("Name Generator"),
		func() string {
			if index < len(options.NameGenerators) {
				return options.NameGenerators[index]
			}
			return ""
		},
		func(name string) {
			if index < len(options.NameGenerators) {
				options.NameGenerators[index] = name
			}
		},
		nameGeneratorPopupItems(d.nameGeneratorChoices, options.NameGenerators[index])...)
	p.popup.Tooltip = newWrappedTooltip(i18n.Text("A name generator, identified by the base name of its .names file. A generator not found in any library is listed so it can be kept or replaced, but produces nothing."))
	p.AddChild(p.popup)

	p.editButton = unison.NewSVGButton(svg.Edit)
	p.editButton.Tooltip = newWrappedTooltip(i18n.Text("Edit this name generator"))
	p.editButton.ClickCallback = func() {
		if name, ok := p.popup.Selected(); ok {
			OpenNameGeneratorInEditor(name)
		}
	}
	p.AddChild(p.editButton)
	p.Sync()
	return p
}

// Sync implements Syncer. The edit button can only open a generator that some library holds, so it follows the popup's
// selection, which changes without the rows being rebuilt.
func (p *nameGeneratorRefPanel) Sync() {
	name, ok := p.popup.Selected()
	p.editButton.SetEnabled(ok && slices.Contains(p.list.dockable.nameGeneratorChoices, name))
}
