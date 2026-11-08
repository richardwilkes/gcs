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
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// The reference key suffixes, after the definition's target key prefix, of the widgets of an attribute's row that the
// dockable reaches for.
const (
	attrDefIDKey     = "id"
	attrDefHeaderKey = "header"
)

// attrDefSettingsPanel is the row of the attribute settings that edits one definition. It starts out collapsed to its
// header, which names the definition, and expands to the definition's fields when the header is clicked; see
// attributeSettingsDockable.expandedRows.
type attrDefSettingsPanel struct {
	unison.Panel
	dockable     *attributeSettingsDockable
	def          *gurps.AttributeDef
	deleteButton *unison.Button
	header       *rowDisclosure
	poolPanel    *poolSettingsPanel
}

func newAttrDefSettingsPanel(dockable *attributeSettingsDockable, def *gurps.AttributeDef) *attrDefSettingsPanel {
	p := &attrDefSettingsPanel{
		dockable: dockable,
		def:      def,
	}
	p.Self = p
	configureEditorRow(p.AsPanel(), 3, true)
	p.fill()
	return p
}

// fill adds the row's parts for its state: the drag handle, the buttons and the content, which is the header alone
// while the row is collapsed.
func (p *attrDefSettingsPanel) fill() {
	p.AddChild(NewDragHandle(editorRowDragKey, &editorRowDragData{
		editor: p.dockable,
		row:    p.AsPanel(),
		title:  i18n.Text("Attribute Definition Drag"),
		move:   func(to int) bool { return p.dockable.moveAttributeDef(p.def, to) },
	}))
	p.AddChild(p.createButtons())
	p.AddChild(p.createContent())
}

// isExpanded reports whether the row shows the definition's fields rather than its header alone.
func (p *attrDefSettingsPanel) isExpanded() bool {
	return p.dockable.expandedRows[p.def.KeyPrefix]
}

// toggle expands or collapses the row. Collapsing first closes the editors open on the items the definition's
// thresholds grant, which the collapsed row no longer shows, and leaves the row as it is if one of them stays open.
func (p *attrDefSettingsPanel) toggle() {
	expand := !p.isExpanded()
	if !expand && !closeGrantItemEditorsWithin(p.dockable, p) {
		return
	}
	p.dockable.showExpanded(expand, p)
}

// show fills the row again, expanded or not as asked, with its delete button as enabled as it was, and reports whether
// the focus was within it, which the filling leaves nowhere.
func (p *attrDefSettingsPanel) show(expanded bool) (hadFocus bool) {
	// CurrentFocus, unlike Focus, leaves the focus alone when an earlier row's filling has taken it out of the window.
	hadFocus = unison.AncestorIsOrSelf(p.Window().CurrentFocus(), p)
	canDelete := p.deleteButton.Enabled()
	if expanded {
		p.dockable.expandedRows[p.def.KeyPrefix] = true
	} else {
		delete(p.dockable.expandedRows, p.def.KeyPrefix)
	}
	p.RemoveAllChildren()
	p.fill()
	p.deleteButton.SetEnabled(canDelete)
	p.MarkForLayoutRecursivelyUpward()
	return hadFocus
}

// refreshHeader has the header name the definition as it now is.
func (p *attrDefSettingsPanel) refreshHeader() {
	p.header.setText(attrDefSummary(p.def))
}

// holdsKey reports whether the widget with the reference key belongs to the row: to the definition or to one of its
// thresholds, whose keys begin with their target key prefixes.
func (p *attrDefSettingsPanel) holdsKey(refKey string) bool {
	if p.def.KeyPrefix != "" && strings.HasPrefix(refKey, p.def.KeyPrefix) {
		return true
	}
	return slices.ContainsFunc(p.def.Thresholds, func(threshold *gurps.PoolThreshold) bool {
		return threshold.KeyPrefix != "" && strings.HasPrefix(refKey, threshold.KeyPrefix)
	})
}

// attrDefSummary returns the line naming the definition in its row's header: its ID, its name if it has one, and its
// type.
func attrDefSummary(def *gurps.AttributeDef) string {
	parts := []string{def.DefID}
	if name := strings.TrimSpace(def.Name); name != "" {
		parts = append(parts, name)
	}
	return strings.Join(append(parts, def.Type.String()), i18n.Text(" · "))
}

func (p *attrDefSettingsPanel) createButtons() *unison.Panel {
	p.deleteButton = unison.NewSVGButton(unison.TrashSVG)
	p.deleteButton.ClickCallback = p.deleteAttrDef
	p.deleteButton.Tooltip = newWrappedTooltip(i18n.Text("Remove attribute"))
	if !p.isExpanded() {
		return newEditorRowButtonColumn(p.deleteButton)
	}
	addThresholdButton := unison.NewSVGButton(unison.CircledAddSVG)
	addThresholdButton.ClickCallback = func() { p.poolPanel.addThreshold() }
	addThresholdButton.Tooltip = newWrappedTooltip(i18n.Text("Add pool threshold"))
	addThresholdButton.SetEnabled(p.def.Type == attribute.Pool || p.def.Type == attribute.PoolRef)
	return newEditorRowButtonColumn(p.deleteButton, addThresholdButton)
}

func (p *attrDefSettingsPanel) deleteAttrDef() {
	p.RemoveFromParent()
	p.dockable.adjustDeleteButtons()
	followGrantItemEditors(p.dockable)
	undo := p.dockable.prepareUndo(i18n.Text("Delete Attribute"))
	delete(p.dockable.model.Set, p.def.DefID)
	p.dockable.finishAndPostUndo(undo)
	p.dockable.refreshThresholdFeatures(p.def.DefID)
	p.dockable.MarkModified(nil)
}

func (p *attrDefSettingsPanel) createContent() *unison.Panel {
	content := unison.NewPanel()
	content.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	content.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
		HGrab:  true,
	})
	p.header = newRowDisclosure(attrDefSummary(p.def), p.isExpanded(), p.toggle)
	p.header.RefKey = p.def.KeyPrefix + attrDefHeaderKey
	p.header.SetLayoutData(&unison.FlexLayoutData{HSpan: 2, HAlign: align.Fill, HGrab: true})
	content.AddChild(p.header)
	if p.isExpanded() {
		p.addFields(content)
	} else {
		p.poolPanel = nil
	}
	return content
}

// addFields adds the definition's fields to the content, below the header.
func (p *attrDefSettingsPanel) addFields(content *unison.Panel) {
	text := i18n.Text("ID")
	content.AddChild(NewFieldLeadingLabel(text, false))
	field := NewStringField(p.dockable.targetMgr, p.def.KeyPrefix+attrDefIDKey, text,
		func() string { return p.def.DefID },
		func(s string) {
			if p.validateAttrID(s) {
				was := p.def.DefID
				delete(p.dockable.model.Set, was)
				p.def.DefID = strings.TrimSpace(strings.ToLower(s))
				p.dockable.model.Set[p.def.DefID] = p.def
				p.dockable.refreshThresholdFeatures(was, p.def.DefID)
				p.refreshHeader()
			}
		})
	field.ValidateCallback = func(field *StringField, _ *gurps.AttributeDef) func() bool {
		return func() bool { return p.validateAttrID(field.Text()) }
	}(field, p.def)
	field.SetMinimumTextWidthUsing(prototypeMinIDWidth)
	field.Tooltip = newWrappedTooltip(i18n.Text("A unique ID for the attribute"))
	content.AddChild(field)

	addLabelAndTargetedPopup(content, p.dockable.targetMgr, p.def.KeyPrefix+"type", i18n.Text("Attribute Type"), "",
		func() attribute.Type { return p.def.Type },
		func(typ attribute.Type) { p.applyAttributeType(typ) },
		attribute.Types...).withoutUndo()

	if !p.def.IsSeparator() {
		label := NewFieldLeadingLabel(i18n.Text("Placement"), false)
		content.AddChild(label)
		content.AddChild(p.createPlacementPanel(label))
	}

	const nameKey = "name"
	nameGet := func() string { return p.def.Name }
	nameSet := func(s string) {
		p.def.Name = s
		p.dockable.refreshThresholdFeatures(p.def.DefID)
		p.refreshHeader()
	}
	if p.def.IsSeparator() {
		addLabelAndTargetedStringField(content, p.dockable.targetMgr, p.def.KeyPrefix+nameKey, i18n.Text("Name"),
			i18n.Text("A title to use with the separator"), prototypeMinIDWidth, nameGet, nameSet)
	} else {
		addLabelAndTargetedStringField(content, p.dockable.targetMgr, p.def.KeyPrefix+nameKey, i18n.Text("Short Name"),
			i18n.Text("The name of this attribute, often an abbreviation"), prototypeMinIDWidth, nameGet, nameSet)
		addLabelAndTargetedStringField(content, p.dockable.targetMgr, p.def.KeyPrefix+"fullname", i18n.Text("Full Name"),
			i18n.Text("The full name of this attribute (may be omitted, in which case the Short Name will be used instead)"),
			prototypeMinNameWidth,
			func() string { return p.def.FullName },
			func(s string) { p.def.FullName = s })
		addLabelAndScriptField(content, p.dockable.targetMgr, p.def.KeyPrefix+"base", i18n.Text("Base Value"),
			i18n.Text("The base value, which may be a number or a script expression"),
			func() string { return p.def.Base },
			func(s string) { p.def.Base = s }, false)

		if p.def.Type != attribute.IntegerRef && p.def.Type != attribute.DecimalRef && p.def.Type != attribute.PoolRef {
			addLabelAndDecimalField(content, p.dockable.targetMgr, p.def.KeyPrefix+"cost", i18n.Text("Cost per Point"),
				i18n.Text("The cost per point difference from the base"), &p.def.CostPerPoint, 0, fxp.MaxBasePoints)

			text = i18n.Text("SM Reduction")
			content.AddChild(NewFieldLeadingLabel(text, false))
			numField := NewPercentageField(p.dockable.targetMgr, p.def.KeyPrefix+"sm", text,
				func() int { return p.def.CostAdjPercentPerSM.AsInteger[int]() },
				func(v int) { p.def.CostAdjPercentPerSM = fxp.FromInteger(v) },
				0, 80, false, false)
			numField.Tooltip = newWrappedTooltip(i18n.Text("The reduction in cost for each SM greater than 0"))
			content.AddChild(numField)
		}
	}

	if p.def.Type == attribute.Pool || p.def.Type == attribute.PoolRef {
		p.poolPanel = newPoolSettingsPanel(p.dockable, p.def)
		content.AddChild(p.poolPanel)
	} else {
		p.poolPanel = nil
	}
}

// createPlacementPanel builds the content that follows the "Placement" label. When the placement is Hidden, it expands
// to read "[Hidden] unless trait [trait name] is present, then [placement]", allowing an attribute to be revealed with
// an alternate placement whenever the character has the named trait. An empty trait name keeps it hidden. Each control
// is named explicitly, since the first sits apart from its label and the others follow bits of that sentence, which
// would otherwise name them.
func (p *attrDefSettingsPanel) createPlacementPanel(label *unison.Label) *unison.Panel {
	panel := unison.NewPanel()
	panel.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})

	var placements []attribute.Placement
	if p.def.Type == attribute.Pool || p.def.Type == attribute.PoolRef {
		placements = []attribute.Placement{attribute.Automatic, attribute.Hidden}
	} else {
		placements = attribute.Placements
	}
	panel.AddChild(labelControl(
		NewPopup(p.dockable.targetMgr, p.def.KeyPrefix+"placement", i18n.Text("Placement"),
			func() attribute.Placement { return p.def.Placement },
			func(placement attribute.Placement) {
				p.def.Placement = placement
				p.dockable.sync()
			},
			placements...),
		label,
	))

	columns := 1
	if p.def.Placement == attribute.Hidden {
		panel.AddChild(NewFieldTrailingLabel(i18n.Text("unless trait"), false))

		text := i18n.Text("Trait Name")
		field := NewStringField(p.dockable.targetMgr, p.def.KeyPrefix+"placement_trait", text,
			func() string { return p.def.PlacementTrait },
			func(s string) { p.def.PlacementTrait = s })
		field.Accessibility.Name = text
		field.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
		field.SetMinimumTextWidthUsing(prototypeMinNameWidth)
		field.Tooltip = newWrappedTooltip(i18n.Text("The name of a trait that, when present on the character, changes this attribute's placement to the value on the right. Leave empty to always keep the attribute hidden."))
		panel.AddChild(field)

		panel.AddChild(NewFieldTrailingLabel(i18n.Text("is present, then"), false))
		text = i18n.Text("Placement When Present")
		var whenPresentPopup *Popup[attribute.Placement]
		if p.def.Type == attribute.Pool || p.def.Type == attribute.PoolRef {
			whenPresentPopup = NewPopup(p.dockable.targetMgr, p.def.KeyPrefix+"placement_when_present", text,
				func() attribute.Placement { return attribute.Automatic },
				func(_ attribute.Placement) {},
				attribute.Automatic)
			whenPresentPopup.SetEnabled(false)
		} else {
			whenPresentPopup = NewPopup(p.dockable.targetMgr, p.def.KeyPrefix+"placement_when_present", text,
				func() attribute.Placement { return p.def.PlacementWhenPresent },
				func(placement attribute.Placement) { p.def.PlacementWhenPresent = placement },
				attribute.Automatic, attribute.Primary, attribute.Secondary)
		}
		whenPresentPopup.Accessibility.Name = text
		panel.AddChild(whenPresentPopup)
		columns = 5
	}
	panel.SetLayout(&unison.FlexLayout{
		Columns:  columns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	return panel
}

func (p *attrDefSettingsPanel) validateAttrID(attrID string) bool {
	if key := strings.TrimSpace(strings.ToLower(attrID)); key != "" {
		if key != gurps.SanitizeID(key, false, gurps.ReservedIDs...) {
			return false
		}
		if key == p.def.DefID {
			return true
		}
		_, exists := p.dockable.model.Set[key]
		return !exists
	}
	return false
}

// applyAttributeType changes the attribute's type in a single undo edit of the whole set of definitions, since the
// change may add a threshold, drop the thresholds of a type that has none, as copying the definitions would (see
// gurps.AttributeDef.Clone), or clear the fields a separator has no use for, all of which undoing it must bring back.
func (p *attrDefSettingsPanel) applyAttributeType(attrType attribute.Type) {
	p.dockable.editStructure(i18n.Text("Attribute Type"), func() {
		p.def.Type = attrType
		if p.def.Type != attribute.Pool && p.def.Type != attribute.PoolRef {
			p.def.Thresholds = nil
		} else if len(p.def.Thresholds) == 0 {
			threshold := gurps.NewPoolThreshold()
			threshold.KeyPrefix = p.dockable.targetMgr.NextPrefix()
			p.def.Thresholds = append(p.def.Thresholds, threshold)
		}
		if p.def.IsSeparator() {
			p.def.FullName = ""
			p.def.Base = ""
			p.def.CostPerPoint = 0
			p.def.CostAdjPercentPerSM = 0
		}
	}, "")
}
