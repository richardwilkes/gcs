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
	"io/fs"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

var _ GroupedCloser = &attributeSettingsDockable{}

type attributeSettingsDockable struct {
	undoableSettingsDockable[*gurps.AttributeDefs]
	owner EntityPanel
}

// ShowAttributeSettings shows the Attribute Settings. Pass in nil to edit the defaults or a sheet to edit the sheet's.
func ShowAttributeSettings(owner EntityPanel) {
	if Activate(func(d unison.Dockable) bool {
		if s, ok := d.AsPanel().Self.(*attributeSettingsDockable); ok && owner == s.owner {
			return true
		}
		return false
	}) {
		return
	}
	var defs *gurps.AttributeDefs
	if owner != nil {
		defs = owner.Entity().SheetSettings.Attributes.Clone()
	} else {
		defs = gurps.GlobalSettings().Sheet.Attributes.Clone()
	}
	d := newAttributeSettingsDockable(owner, defs)
	d.TabTitle = attributeSettingsTabTitle(owner)
	d.TabIcon = svg.Attributes
	d.Extensions = []string{gurps.AttributesExt, gurps.AttributesExtAlt1, gurps.AttributesExtAlt2}
	d.Loader = d.load
	d.Resetter = d.reset
	d.show()
}

// newAttributeSettingsDockable returns a dockable for the owner, which is nil for the defaults, that edits the given
// attribute definitions in place, ready for its content to be built.
func newAttributeSettingsDockable(owner EntityPanel, defs *gurps.AttributeDefs) *attributeSettingsDockable {
	d := &attributeSettingsDockable{owner: owner}
	d.Self = d
	d.init(d, undoableSettingsSpec[*gurps.AttributeDefs]{
		helpLink:     "md:User%20Guide/Attributes",
		clone:        (*gurps.AttributeDefs).Clone,
		buildContent: d.buildContent,
		apply:        d.apply,
		extraToolbar: d.addToToolbar,
	}, defs)
	return d
}

// attributeSettingsTabTitle returns the tab title for the owner's attribute settings, or for the defaults when the
// owner is nil.
func attributeSettingsTabTitle(owner EntityPanel) string {
	if owner == nil {
		return i18n.Text("Default Attributes")
	}
	return fmt.Sprintf(i18n.Text("Attributes: %s"), owner.Entity().Profile.Name)
}

func (d *attributeSettingsDockable) CloseWithGroup(other unison.Paneler) bool {
	return d.owner != nil && d.owner == other
}

// addToToolbar adds the Add Attribute button after the standard buttons.
func (d *attributeSettingsDockable) addToToolbar(toolbar *unison.Panel) {
	toolbar.AddChild(NewToolbarSeparator())

	addButton := unison.NewSVGButton(unison.CircledAddSVG)
	addAttributeText := i18n.Text("Add Attribute")
	addButton.Tooltip = newWrappedTooltip(addAttributeText)
	addButton.ClickCallback = func() {
		p := d.addAttribute(addAttributeText)
		p.MarkForLayoutRecursivelyUpward()
		d.ValidateLayout()
		FocusFirstContent(d.toolbar, p.AsPanel())
		d.Window().Focus().ScrollIntoView()
	}
	toolbar.AddChild(addButton)
}

// addAttribute adds a new attribute definition, gives it a panel and records the undo edit, returning the new panel.
func (d *attributeSettingsDockable) addAttribute(undoName string) *attrDefSettingsPanel {
	undo := d.prepareUndo(undoName)
	p := newAttrDefSettingsPanel(d, d.addAttributeDef())
	d.content.AddChild(p)
	d.adjustDeleteButtons()
	d.finishAndPostUndo(undo)
	d.MarkModified(nil)
	return p
}

// addAttributeDef creates a new attribute definition with an unused ID, an order that places it last, and its own
// target key prefix, then adds it to the set. The key prefix is what makes the definition's widget reference keys
// unique within the dockable; without one, a second added attribute would build the same reference keys as the first,
// and undo and focus restoration would then resolve to the wrong attribute's widgets.
func (d *attributeSettingsDockable) addAttributeDef() *gurps.AttributeDef {
	attrDef := &gurps.AttributeDef{KeyPrefix: d.targetMgr.NextPrefix()}
	base := ""
	for {
		for v := 'a'; v <= 'z'; v++ {
			attempt := fmt.Sprintf("%s%c", base, v)
			if _, exists := d.model.Set[attempt]; !exists {
				attrDef.DefID = attempt
				break
			}
		}
		if attrDef.DefID != "" {
			break
		}
		base += "a"
	}
	for _, v := range d.model.Set {
		if attrDef.Order <= v.Order {
			attrDef.Order = v.Order + 1
		}
	}
	d.model.Set[attrDef.DefID] = attrDef
	return attrDef
}

func (d *attributeSettingsDockable) buildContent() {
	for _, def := range d.model.List(false) {
		d.content.AddChild(newAttrDefSettingsPanel(d, def))
	}
	d.adjustDeleteButtons()
}

// adjustDeleteButtons enables the delete button of every attribute panel unless only one attribute remains, in which
// case the lone panel's button is disabled so that the last attribute can't be removed. Each panel is created with its
// button enabled, so this must be called whenever panels are added or rebuilt, not just when one is deleted.
func (d *attributeSettingsDockable) adjustDeleteButtons() {
	children := d.content.Children()
	enabled := len(children) > 1
	for _, child := range children {
		if panel, ok := child.Self.(*attrDefSettingsPanel); ok {
			panel.deleteButton.SetEnabled(enabled)
		}
	}
}

func (d *attributeSettingsDockable) Entity() *gurps.Entity {
	if d.owner != nil {
		return d.owner.Entity()
	}
	return nil
}

// moveAttributeDef moves the definition to the given insertion position among the definitions, renumbering them to
// match, and reports whether anything changed. It is what dropping a dragged definition does.
func (d *attributeSettingsDockable) moveAttributeDef(def *gurps.AttributeDef, to int) bool {
	list := d.model.List(false)
	if !moveEntry(&list, slices.Index(list, def), to) {
		return false
	}
	for i, one := range list {
		one.Order = i
	}
	return true
}

func (d *attributeSettingsDockable) reset() {
	var defs *gurps.AttributeDefs
	if d.owner != nil {
		defs = gurps.GlobalSettings().Sheet.Attributes.Clone()
	} else {
		defs = gurps.FactoryAttributeDefs()
	}
	d.replaceModel(i18n.Text("Reset Attributes"), defs)
}

func (d *attributeSettingsDockable) load(fileSystem fs.FS, filePath string) error {
	defs, err := gurps.NewAttributeDefsFromFile(fileSystem, filePath)
	if err != nil {
		return err
	}
	replacements, replace, canceled := SelectAttributeDefs(defs)
	if canceled {
		return nil
	}
	if !replace {
		replacements = mergeAttributeDefs(d.model, replacements)
	}
	d.replaceModel(i18n.Text("Load Attributes"), replacements)
	return nil
}

// mergeAttributeDefs returns a copy of existing with the definitions from incoming merged into it. A definition that is
// already present keeps the position it currently has, while the new ones are appended after all of the existing ones,
// retaining the relative order they had in the file they came from. The incoming definitions must therefore be visited
// in that order rather than by ranging over the map, whose iteration order is random.
func mergeAttributeDefs(existing, incoming *gurps.AttributeDefs) *gurps.AttributeDefs {
	merged := existing.Clone()
	nextOrder := 0
	for _, def := range merged.Set {
		if def.Order > nextOrder {
			nextOrder = def.Order
		}
	}
	nextOrder++
	for _, def := range incoming.List(false) {
		if origDef, exists := merged.Set[def.ID()]; exists {
			def.Order = origDef.Order
		} else {
			def.Order = nextOrder
			nextOrder++
		}
		merged.Set[def.ID()] = def
	}
	return merged
}

func (d *attributeSettingsDockable) apply() {
	d.Window().FocusNext() // Intentionally move the focus to ensure any pending edits are flushed
	if d.owner == nil {
		gurps.GlobalSettings().Sheet.Attributes = d.model.Clone()
		gurps.SyncGlobalSheetSettings()
		return
	}
	entity := d.owner.Entity()
	entity.SheetSettings.Attributes = d.model.Clone()
	for attrID, def := range entity.SheetSettings.Attributes.Set {
		if attr, exists := entity.Attributes.Set[attrID]; exists {
			attr.Order = def.Order
		} else {
			entity.Attributes.Set[attrID] = gurps.NewAttribute(entity, attrID, def.Order)
		}
	}
	for attrID := range entity.Attributes.Set {
		if _, exists := d.model.Set[attrID]; !exists {
			delete(entity.Attributes.Set, attrID)
		}
	}
	for _, one := range AllDockables() {
		if s, ok := one.(gurps.SheetSettingsResponder); ok {
			s.SheetSettingsUpdated(entity, true)
		}
	}
}
