// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

var lastDefaultTypeUsed = gurps.DexterityID

type defaultsPanel struct {
	unison.Panel
	entity   *gurps.Entity
	defaults *[]*gurps.SkillDefault
}

func newDefaultsPanel(entity *gurps.Entity, defaults *[]*gurps.SkillDefault) *defaultsPanel {
	p := &defaultsPanel{
		entity:   entity,
		defaults: defaults,
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Fill,
		HGrab:  true,
	})
	p.SetBorder(unison.NewCompoundBorder(
		&TitledBorder{
			Title: i18n.Text("Defaults"),
			Font:  unison.LabelFont,
		},
		unison.NewEmptyBorder(unison.NewUniformInsets(2))))
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, unison.ThemeSurface.Paint(gc, rect, paintstyle.Fill))
	}
	addButton := unison.NewSVGButton(svg.CircledAdd)
	addButton.ClickCallback = func() {
		def := &gurps.SkillDefault{DefaultType: lastDefaultTypeUsed}
		// See the comment for the delete button as to why we don't just use slices.Insert here.
		defs := make([]*gurps.SkillDefault, len(*p.defaults)+1)
		defs[0] = def
		copy(defs[1:], *p.defaults)
		*p.defaults = defs
		p.insertDefaultsPanel(1, def)
		MarkRootAncestorForLayoutRecursively(p)
		MarkModified(p)
	}
	p.AddChild(addButton)
	for i, one := range *p.defaults {
		p.insertDefaultsPanel(i+1, one)
	}
	return p
}

func (p *defaultsPanel) insertDefaultsPanel(index int, def *gurps.SkillDefault) {
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  5,
		HAlign:   align.Fill,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})

	deleteButton := unison.NewSVGButton(svg.Trash)
	deleteButton.ClickCallback = func() {
		if i := slices.IndexFunc(*p.defaults, func(elem *gurps.SkillDefault) bool { return elem == def }); i != -1 {
			// Cannot use this here: *p.defaults = slices.Delete(*p.defaults, i, i+1)
			// because it will cause a panic elsewhere due to some code not having the owner properly updated. This
			// causes the original slice to be looked at, which now has a nil at its end, which causes problems in code
			// that expects no nils. So... we do it the more expensive and verbose way.
			defs := make([]*gurps.SkillDefault, len(*p.defaults)-1)
			copy(defs, (*p.defaults)[:i])
			copy(defs[i:], (*p.defaults)[i+1:])
			*p.defaults = defs
		}
		panel.RemoveFromParent()
		MarkRootAncestorForLayoutRecursively(p)
		MarkModified(p)
	}
	panel.AddChild(deleteButton)

	name := i18n.Text("Name")
	nameField := NewStringField(nil, "", name, func() string { return def.Name },
		func(s string) { def.Name = s })
	nameField.Watermark = name
	specialization := i18n.Text("Specialization")
	specializationField := NewStringField(nil, "", specialization,
		func() string { return def.Specialization }, func(s string) { def.Specialization = s })
	specializationField.Watermark = specialization
	modifierField := NewDecimalField(nil, "", i18n.Text("Modifier"),
		func() fxp.Int { return def.Modifier },
		func(v fxp.Int) { def.Modifier = v },
		-fxp.Thousand, fxp.Thousand, true, false)
	attrChoicePopup := addAttributeChoicePopup(panel, p.entity, "", &def.DefaultType,
		gurps.TenFlag|gurps.ParryFlag|gurps.BlockFlag|gurps.SkillFlag)
	callback := attrChoicePopup.SelectionChangedCallback
	attrChoicePopup.SelectionChangedCallback = func(popup *unison.PopupMenu[*gurps.AttributeChoice]) {
		if item, ok := popup.Selected(); ok {
			lastDefaultTypeUsed = item.Key
			callback(popup)
			adjustFieldBlank(nameField, item.Key != gurps.SkillID)
			adjustFieldBlank(specializationField, item.Key != gurps.SkillID)
		}
	}
	panel.AddChild(nameField)
	panel.AddChild(specializationField)
	panel.AddChild(modifierField)
	adjustFieldBlank(nameField, def.DefaultType != gurps.SkillID)
	adjustFieldBlank(specializationField, def.DefaultType != gurps.SkillID)

	addNumericCriteriaPanel(panel, nil, "", i18n.Text("when the Tech Level"), i18n.Text("When Tech Level"),
		&def.WhenTL, 0, fxp.Twelve, 4, true, true)

	panel.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	p.AddChildAtIndex(panel, index)
}
