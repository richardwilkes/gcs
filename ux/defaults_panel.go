/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ux

import (
	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/slices"
)

var lastDefaultTypeUsed = model.DexterityID

type defaultsPanel struct {
	unison.Panel
	entity   *model.Entity
	defaults *[]*model.SkillDefault
}

func newDefaultsPanel(entity *model.Entity, defaults *[]*model.SkillDefault) *defaultsPanel {
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
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	p.SetBorder(unison.NewCompoundBorder(
		&TitledBorder{
			Title: i18n.Text("Defaults"),
			Font:  unison.LabelFont,
		},
		unison.NewEmptyBorder(unison.NewUniformInsets(2))))
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, unison.ContentColor.Paint(gc, rect, unison.Fill))
	}
	addButton := unison.NewSVGButton(svg.CircledAdd)
	addButton.ClickCallback = func() {
		def := &model.SkillDefault{DefaultType: lastDefaultTypeUsed}
		*defaults = slices.Insert(*defaults, 0, def)
		p.insertDefaultsPanel(1, def)
		unison.Ancestor[*unison.DockContainer](p).MarkForLayoutRecursively()
		MarkModified(p)
	}
	p.AddChild(addButton)
	for i, one := range *defaults {
		p.insertDefaultsPanel(i+1, one)
	}
	return p
}

func (p *defaultsPanel) insertDefaultsPanel(index int, def *model.SkillDefault) {
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  5,
		HAlign:   unison.FillAlignment,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})

	deleteButton := unison.NewSVGButton(svg.Trash)
	deleteButton.ClickCallback = func() {
		if i := slices.IndexFunc(*p.defaults, func(elem *model.SkillDefault) bool { return elem == def }); i != -1 {
			*p.defaults = slices.Delete(*p.defaults, i, i+1)
		}
		panel.RemoveFromParent()
		unison.Ancestor[*unison.DockContainer](p).MarkForLayoutRecursively()
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
		model.TenFlag|model.ParryFlag|model.BlockFlag|model.SkillFlag)
	callback := attrChoicePopup.SelectionCallback
	attrChoicePopup.SelectionCallback = func(index int, item *model.AttributeChoice) {
		lastDefaultTypeUsed = item.Key
		callback(index, item)
		adjustFieldBlank(nameField, item.Key != model.SkillID)
		adjustFieldBlank(specializationField, item.Key != model.SkillID)
	}
	panel.AddChild(nameField)
	panel.AddChild(specializationField)
	panel.AddChild(modifierField)
	adjustFieldBlank(nameField, def.DefaultType != model.SkillID)
	adjustFieldBlank(specializationField, def.DefaultType != model.SkillID)

	panel.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	p.AddChildAtIndex(panel, index)
}
