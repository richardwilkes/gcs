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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

var _ GroupedCloser = &attributeSettingsDockable{}

type attributeSettingsDockable struct {
	undoableSettingsDockable[*gurps.AttributeDefs]
	owner EntityPanel
	// expandedRows holds the target key prefixes of the definitions whose rows show their fields; the rest show their
	// header alone. The prefixes survive the copies undo and redo install, while loading and resetting assign new ones,
	// so every row then starts out collapsed. Expanding a row is no edit: it is neither undone nor a modification.
	expandedRows map[string]bool
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
	d := &attributeSettingsDockable{owner: owner, expandedRows: make(map[string]bool)}
	d.Self = d
	d.init(d, undoableSettingsSpec[*gurps.AttributeDefs]{
		helpLink:     "md:User%20Guide/Attributes",
		clone:        (*gurps.AttributeDefs).Clone,
		buildContent: d.buildContent,
		apply:        d.apply,
		extraToolbar: d.addToToolbar,
	}, defs)
	d.targetMgr.reveal = d.revealKey
	return d
}

// attributeSettingsTabTitle returns the tab title for the owner's attribute settings, or for the defaults when the
// owner is nil.
func attributeSettingsTabTitle(owner EntityPanel) string {
	if owner == nil {
		return i18n.Text("Default Attributes")
	}
	return i18n.Text("Attributes: %s", owner.Entity().Profile.Name)
}

func (d *attributeSettingsDockable) CloseWithGroup(other unison.Paneler) bool {
	return d.owner != nil && d.owner == other
}

// String implements Rebuildable, naming the dockable as the editors of the items its thresholds grant name where they
// were opened from.
func (d *attributeSettingsDockable) String() string {
	return d.Title()
}

// Rebuild implements Rebuildable, rebuilding the content from the model.
func (d *attributeSettingsDockable) Rebuild(_ bool) {
	d.sync()
}

// addToToolbar adds the Add Attribute button and the button that expands or collapses every row after the standard
// buttons.
func (d *attributeSettingsDockable) addToToolbar(toolbar *unison.Panel) {
	toolbar.AddChild(NewToolbarSeparator())

	addButton := unison.NewSVGButton(unison.CircledAddSVG)
	addAttributeText := i18n.Text("Add Attribute")
	addButton.Tooltip = newWrappedTooltip(addAttributeText)
	addButton.ClickCallback = func() {
		p := d.addAttribute(addAttributeText)
		p.MarkForLayoutRecursivelyUpward()
		d.ValidateLayout()
		d.focusOn(p.def.KeyPrefix + attrDefIDKey)
	}
	toolbar.AddChild(addButton)

	toggleButton := unison.NewSVGButton(svg.Hierarchy)
	toggleButton.Tooltip = newWrappedTooltip(i18n.Text("Expands/collapses all attribute definitions"))
	toggleButton.ClickCallback = d.toggleAllExpanded
	toolbar.AddChild(toggleButton)
}

// addAttribute adds a new attribute definition, gives it a panel, expanded so that it can be edited at once, and
// records the undo edit, returning the new panel.
func (d *attributeSettingsDockable) addAttribute(undoName string) *attrDefSettingsPanel {
	undo := d.prepareUndo(undoName)
	def := d.addAttributeDef()
	d.expandedRows[def.KeyPrefix] = true
	p := newAttrDefSettingsPanel(d, def)
	d.content.AddChild(p)
	d.adjustDeleteButtons()
	d.finishAndPostUndo(undo)
	d.refreshThresholdFeatures(p.def.DefID)
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
	followGrantItemEditors(d)
}

// refreshThresholdFeatures shows again the features of the thresholds that a change to an attribute's ID or name, or
// its addition or deletion, which rebuild nothing else, may have altered, given the IDs the attribute was and is known
// by: those holding an attribute bonus to either, which names the attribute as the definitions being edited do, or is
// ignored while they define none with that ID (see featuresPanel.describe), and those with a row open, whose attribute
// popup lists what the definitions define.
func (d *attributeSettingsDockable) refreshThresholdFeatures(ids ...string) {
	for _, section := range keyedSections(d.content) {
		p, ok := section.(*featuresPanel)
		if ok && (p.open != "" || slices.ContainsFunc(*p.features, func(f gurps.Feature) bool {
			bonus, isBonus := f.(*gurps.AttributeBonus)
			return isBonus && slices.Contains(ids, bonus.Attribute)
		})) {
			p.rebuild("")
		}
	}
}

// adjustDeleteButtons enables the delete button of every attribute panel unless only one attribute remains, in which
// case the lone panel's button is disabled so that the last attribute can't be removed. Each panel is created with its
// button enabled, so this must be called whenever panels are added or rebuilt, not just when one is deleted.
func (d *attributeSettingsDockable) adjustDeleteButtons() {
	rows := d.rows()
	enabled := len(rows) > 1
	for _, p := range rows {
		p.deleteButton.SetEnabled(enabled)
	}
}

// rows returns the panels of the definitions, in the order shown.
func (d *attributeSettingsDockable) rows() []*attrDefSettingsPanel {
	children := d.content.Children()
	rows := make([]*attrDefSettingsPanel, 0, len(children))
	for _, child := range children {
		if p, ok := child.Self.(*attrDefSettingsPanel); ok {
			rows = append(rows, p)
		}
	}
	return rows
}

// toggleAllExpanded expands every row while all of them are collapsed, and otherwise collapses them all, first closing
// the editors open on the items the thresholds grant, as collapsing a single row does; one of them staying open leaves
// the rows as they are.
func (d *attributeSettingsDockable) toggleAllExpanded() {
	rows := d.rows()
	expand := !slices.ContainsFunc(rows, (*attrDefSettingsPanel).isExpanded)
	if !expand && !closeGrantItemEditorsWithin(d, d.content) {
		return
	}
	d.showExpanded(expand, rows...)
}

// showExpanded shows the rows expanded or collapsed as asked, those not already so, and hands the focus to the header
// of the row that held it, which filling the row again leaves nowhere. This is no edit: nothing is undone or marked
// modified.
func (d *attributeSettingsDockable) showExpanded(expand bool, rows ...*attrDefSettingsPanel) {
	var focus *rowDisclosure
	for _, p := range rows {
		if p.isExpanded() != expand && p.show(expand) {
			focus = p.header
		}
	}
	d.ValidateLayout()
	d.MarkForRedraw()
	if focus != nil {
		focus.RequestFocus()
		focus.ScrollIntoView()
	}
}

// showForOwnWindow implements ownWindowShower, packing the window with a typical row expanded: the first whose
// definition is neither a separator nor a pool, or else the first. The rows start out collapsed, and a window packed
// around their headers alone is too small to edit one in.
func (d *attributeSettingsDockable) showForOwnWindow(pack func()) {
	rows := d.rows()
	var row *attrDefSettingsPanel
	for _, p := range rows {
		if !p.def.IsSeparator() && p.def.Type != attribute.Pool && p.def.Type != attribute.PoolRef {
			row = p
			break
		}
	}
	if row == nil && len(rows) != 0 {
		row = rows[0]
	}
	if row == nil || row.isExpanded() {
		pack()
		return
	}
	row.show(true)
	pack()
	row.show(false)
}

// revealKey expands the collapsed row holding the widget with the reference key, if any, so that the undo of an edit
// made in the row can find the widget (see TargetMgr.findRevealing), and reports whether it did. Only the rows shown
// are looked at: a row being deleted is out of the content before its definition is out of the set.
func (d *attributeSettingsDockable) revealKey(refKey string) bool {
	for _, p := range d.rows() {
		if !p.isExpanded() && p.holdsKey(refKey) {
			d.showExpanded(true, p)
			return true
		}
	}
	return false
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
	var entity *gurps.Entity
	if d.owner == nil {
		// The library lists and templates, which have no entity, show the default definitions' names. A nil entity is
		// what tells them the change was to the defaults.
		gurps.GlobalSettings().Sheet.Attributes = d.model.Clone()
		gurps.SyncGlobalSheetSettings()
	} else {
		entity = d.owner.Entity()
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
	}
	notifySheetSettingsUpdated(entity, gurps.SheetSettingsAttributesPart, true)
}
