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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/study"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

var lastStudyTypeUsed = study.Self

type studyPanel struct {
	unison.Panel
	entity      *gurps.Entity
	studyNeeded *study.Level
	study       *[]*gurps.Study
	total       *unison.Label
}

func newStudyPanel(entity *gurps.Entity, studyNeeded *study.Level, s *[]*gurps.Study) *studyPanel {
	p := &studyPanel{
		entity:      entity,
		studyNeeded: studyNeeded,
		study:       s,
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
			Title: i18n.Text("Study"),
			Font:  unison.LabelFont,
		},
		unison.NewEmptyBorder(unison.NewUniformInsets(2))))
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, unison.ThemeSurface.Paint(gc, rect, paintstyle.Fill))
	}

	top := unison.NewPanel()
	top.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	p.AddChild(top)

	addButton := unison.NewSVGButton(svg.CircledAdd)
	addButton.ClickCallback = func() {
		def := &gurps.Study{Type: lastStudyTypeUsed}
		*s = slices.Insert(*s, 0, def)
		p.insertStudyEntry(1, def, true)
		MarkRootAncestorForLayoutRecursively(p)
		MarkModified(p)
	}
	top.AddChild(addButton)

	topRight := unison.NewPanel()
	topRight.SetLayout(&unison.FlexLayout{Columns: 3})
	p.AddChild(topRight)
	top.AddChild(topRight)

	p.total = unison.NewLabel()
	p.updateTotal()
	topRight.AddChild(p.total)

	hoursNeededPopup := addPopup(topRight, study.Levels, studyNeeded)
	hoursNeededPopup.SelectionChangedCallback = func(popup *unison.PopupMenu[study.Level]) {
		if needed, ok := popup.Selected(); ok {
			*studyNeeded = needed
			p.updateTotal()
			MarkModified(top)
		}
	}

	trailer := unison.NewLabel()
	trailer.SetTitle(i18n.Text(") for 1 point"))
	topRight.AddChild(trailer)

	for i, one := range *s {
		p.insertStudyEntry(i+1, one, false)
	}
	return p
}

func (p *studyPanel) insertStudyEntry(index int, entry *gurps.Study, requestFocus bool) {
	panel := unison.NewPanel()

	deleteButton := unison.NewSVGButton(svg.Trash)
	deleteButton.ClickCallback = func() {
		if i := slices.IndexFunc(*p.study, func(one *gurps.Study) bool { return one == entry }); i != -1 {
			*p.study = slices.Delete(*p.study, i, i+1)
		}
		panel.RemoveFromParent()
		MarkRootAncestorForLayoutRecursively(p)
		p.updateTotal()
		MarkModified(p)
	}
	panel.AddChild(deleteButton)

	var adjustedHoursField *DecimalField

	info := NewInfoPop()
	updateLimitations(info, entry.Type)
	typePopup := addPopup(panel, study.Types, &entry.Type)
	typePopup.SelectionChangedCallback = func(popup *unison.PopupMenu[study.Type]) {
		if studyType, ok := popup.Selected(); ok {
			entry.Type = studyType
			lastStudyTypeUsed = studyType
			updateLimitations(info, studyType)
			p.updateTotal()
			MarkModified(panel)
		}
	}
	panel.AddChild(info)

	hoursField := NewDecimalField(nil, "", i18n.Text("Hours Spent"),
		func() fxp.Int { return entry.Hours },
		func(v fxp.Int) {
			entry.Hours = v
			p.updateTotal()
		},
		0, fxp.Thousand, false, false)
	panel.AddChild(hoursField)

	adjustedHoursField = NewDecimalField(nil, "", i18n.Text("Hours of Study"),
		func() fxp.Int { return entry.Hours.Mul(entry.Type.Multiplier()) },
		func(_ fxp.Int) {},
		0, fxp.Thousand, false, false)
	adjustedHoursField.SetEnabled(false)
	panel.AddChild(adjustedHoursField)

	note := i18n.Text("Note")
	notesField := NewStringField(nil, "", note, func() string { return entry.Note },
		func(s string) { entry.Note = s })
	notesField.Watermark = note
	panel.AddChild(notesField)

	panel.SetLayout(&unison.FlexLayout{
		Columns:  len(panel.Children()),
		HAlign:   align.Fill,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	p.AddChildAtIndex(panel, index)
	if requestFocus {
		hoursField.RequestFocus()
	}
}

func updateLimitations(info *unison.Label, studyType study.Type) {
	ClearInfoPop(info)
	for _, one := range studyType.Limitations() {
		AddHelpToInfoPop(info, txt.Wrap("", "● "+one, 60))
	}
}

func (p *studyPanel) updateTotal() {
	text := gurps.StudyHoursProgressText(gurps.ResolveStudyHours(*p.study), *p.studyNeeded, true) + " ("
	if text != p.total.Text.String() {
		p.total.SetTitle(text)
		p.total.MarkForLayoutAndRedraw()
	}
}
