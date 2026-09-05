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
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// ancestryEditorPanel is the root of the ancestry editor's content: the ancestry's name, its common options, and a row
// for each of its genders.
type ancestryEditorPanel struct {
	unison.Panel
	dockable *ancestryEditorDockable
	// genders holds one genderOptionsPanel per gender and nothing else, so that a gender drag can use its children as
	// the drop positions.
	genders *unison.Panel
}

func newAncestryEditorPanel(d *ancestryEditorDockable) *ancestryEditorPanel {
	p := &ancestryEditorPanel{dockable: d}
	p.Self = p
	p.SetBorder(unison.NewEmptyBorder(geom.Insets{
		Top:    unison.StdVSpacing,
		Left:   unison.StdHSpacing,
		Bottom: unison.StdVSpacing,
		Right:  unison.StdHSpacing * 2,
	}))
	p.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	p.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})

	a := d.model
	text := i18n.Text("Name")
	p.AddChild(NewFieldLeadingLabel(text, false))
	field := NewStringField(d.targetMgr, a.KeyPrefix+"name", text,
		func() string { return a.Name },
		func(s string) { a.Name = s })
	field.SetMinimumTextWidthUsing(prototypeMinNameWidth)
	field.Tooltip = newWrappedTooltip(i18n.Text("The name of this ancestry. Traits select an ancestry by the base name of its file rather than by this name, so this is mostly descriptive; it is shown when a template asks about replacing a character's ancestry, and it is offered as the default file name when saving."))
	p.AddChild(field)

	p.AddChild(newEditorSectionHeader(i18n.Text("Common Options"),
		i18n.Text("Options used when the chosen gender does not provide its own")))
	common := newAncestryOptionsPanel(d, a.CommonOptions, nil)
	lineBorder := unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.NewUniformInsets(1), false)
	common.SetBorder(unison.NewCompoundBorder(lineBorder, unison.NewEmptyBorder(unison.StdInsets())))
	common.SetLayoutData(&unison.FlexLayoutData{HSpan: 2, HAlign: align.Fill, HGrab: true})
	p.AddChild(common)

	addButton := unison.NewSVGButton(unison.CircledAddSVG)
	addButton.Tooltip = newWrappedTooltip(i18n.Text("Add gender"))
	addButton.ClickCallback = p.addGender
	p.AddChild(newEditorSectionHeader(i18n.Text("Genders"),
		i18n.Text("Genders that may be chosen at random, each of which may override any of the common options"),
		addButton))
	p.genders = unison.NewPanel()
	p.genders.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.NewUniformInsets(1), false))
	p.genders.SetLayout(&unison.FlexLayout{Columns: 1})
	p.genders.SetLayoutData(&unison.FlexLayoutData{HSpan: 2, HAlign: align.Fill, HGrab: true})
	for _, gender := range a.GenderOptions {
		p.genders.AddChild(newGenderOptionsPanel(d, gender))
	}
	p.AddChild(p.genders)
	return p
}

// addGender appends a gender with no overrides of its own and moves the focus to its name.
func (p *ancestryEditorPanel) addGender() {
	d := p.dockable
	gender := &gurps.WeightedAncestryOptions{
		Weight:    1,
		Value:     &gurps.AncestryOptions{KeyPrefix: d.targetMgr.NextPrefix()},
		KeyPrefix: d.targetMgr.NextPrefix(),
	}
	d.editStructure(i18n.Text("Add Gender"), func() { d.model.GenderOptions = append(d.model.GenderOptions, gender) },
		gender.Value.KeyPrefix+"name")
}
