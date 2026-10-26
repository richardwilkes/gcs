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
	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
)

const (
	templatePanelFieldPrefix                 = "template:"
	templatePanelNameFieldRefKey             = templatePanelFieldPrefix + "name"
	templatePanelTargetMinFieldRefKey        = templatePanelFieldPrefix + "target_min"
	templatePanelTargetMaxFieldRefKey        = templatePanelFieldPrefix + "target_max"
	templatePanelPageRefFieldRefKey          = templatePanelFieldPrefix + "page_ref"
	templatePanelPageRefHighlightFieldRefKey = templatePanelFieldPrefix + "page_ref_highlight"
	templatePanelLocalNotesFieldRefKey       = templatePanelFieldPrefix + "local_notes"
	templatePanelCurrentPointsFieldRefKey    = templatePanelFieldPrefix + "current_points"
	templatePanelBodyTypeFieldRefKey         = templatePanelFieldPrefix + "body_type"
)

// createTopBlock returns the block at the top of the template that holds the values belonging to the template as a
// whole rather than to any one of its lists.
func (t *Template) createTopBlock() *unison.Panel {
	top := unison.NewPanel()
	_, layoutData := initTitledPagePanel(top, i18n.Text("Template"), 2, false, nil)
	layoutData.HGrab = true
	data := &t.template.TemplateData
	addLabeledStringPageField(top, t.targetMgr, templatePanelNameFieldRefKey, i18n.Text("Name"), NewPageLabel,
		func() string { return data.Name }, func(s string) { data.Name = s })
	t.addTargetPointsRow(top)
	t.addPageRefRow(top)
	field := addLabeledStringPageField(top, t.targetMgr, templatePanelPageRefHighlightFieldRefKey,
		i18n.Text("Page Highlight"), NewPageLabel, func() string { return data.PageRefHighlight },
		func(s string) { data.PageRefHighlight = s })
	field.Tooltip = newWrappedTooltip(i18n.Text(`A snippet of text to highlight on the page when opening the page reference; only needed if the default behavior isn't highlighting the expected area`))
	t.addBodyTypeRow(top)
	addMultiLineStringPageField(top, t.targetMgr, templatePanelLocalNotesFieldRefKey, i18n.Text("Notes"),
		func() string { return data.LocalNotes }, func(s string) { data.LocalNotes = s })
	return top
}

// addPageRefRow adds the row that edits the template's page reference. A button beside the label opens it, and is
// enabled only while the field holds a reference to open.
func (t *Template) addPageRefRow(parent *unison.Panel) {
	title := i18n.Text("Page Reference")
	data := &t.template.TemplateData
	label := NewPageLabel(title)
	wrapper, refresh := newPageRefLabel(label, fonts.PageLabelPrimary, func() string { return data.PageRef },
		t.pageRefHighlight)
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	parent.AddChild(wrapper)

	field := NewStringPageField(t.targetMgr, templatePanelPageRefFieldRefKey, title,
		func() string { return data.PageRef },
		func(s string) {
			data.PageRef = s
			refresh()
		})
	field.Tooltip = newWrappedTooltip(gurps.PageRefTooltip())
	field.Accessibility.LabeledBy = label
	parent.AddChild(field)
}

// pageRefHighlight returns the text to look for on the page the template's page reference opens: the highlight the
// template holds or, while it holds none, its name, as the page reference cell of a table resolves it.
func (t *Template) pageRefHighlight() string {
	data := &t.template.TemplateData
	if data.PageRefHighlight != "" {
		return data.PageRefHighlight
	}
	return data.Name
}

// addTargetPointsRow adds the row that edits the range of points the template is meant to cost, alongside what the
// template currently costs. Leaving an end blank leaves that end of the range open.
func (t *Template) addTargetPointsRow(parent *unison.Panel) {
	title := i18n.Text("Target Points")
	label := NewPageLabel(title)
	label.Tooltip = newWrappedTooltip(i18n.Text("The number of points the template is meant to cost. Enter the same value in both fields for an exact target, or leave one blank to leave that end of the range open."))
	parent.AddChild(label)

	row := unison.NewPanel()
	row.SetLayout(&unison.FlexLayout{
		Columns:  6,
		HSpacing: 4,
	})
	row.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	target := &t.template.TargetPoints
	addEnd := func(title, undoTitle, key string, end **fxp.Int) *OptionalDecimalField {
		field := NewOptionalDecimalField(t.targetMgr, key, undoTitle, func() *fxp.Int { return *end },
			func(value *fxp.Int) { *end = value })
		installPageFieldFontAndFocusBorders(field.Field)
		field.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Middle})
		field.Accessibility.Name = undoTitle
		row.AddChild(NewPageLabel(title))
		row.AddChild(field)
		return field
	}
	minField := addEnd(i18n.Text("Min"), i18n.Text("Target Points Minimum"), templatePanelTargetMinFieldRefKey,
		&target.Min)
	maxField := addEnd(i18n.Text("Max"), i18n.Text("Target Points Maximum"), templatePanelTargetMaxFieldRefKey,
		&target.Max)
	minField.Problem = func(value *fxp.Int) string { return targetPointsProblem(value, target.Max) }
	maxField.Problem = func(value *fxp.Int) string { return targetPointsProblem(target.Min, value) }
	minField.Validate()
	maxField.Validate()

	currentLabel := NewPageLabel(i18n.Text("Current"))
	currentLabel.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: unison.StdHSpacing * 2, Bottom: 1}))
	row.AddChild(currentLabel)
	current := NewNonEditablePageField(func(f *NonEditablePageField) {
		points := t.template.PointsRange()
		f.Tooltip = currentPointsTooltip(points, t.template.TargetPoints)
		var ink unison.Ink = unison.ThemeOnSurface
		if !targetPointsMet(points, t.template.TargetPoints) {
			ink = unison.ThemeError
		}
		if f.OnBackgroundInk != ink {
			f.OnBackgroundInk = ink
			f.Text.AdjustDecorations(func(d *unison.TextDecoration) { d.OnBackgroundInk = ink })
			f.MarkForRedraw()
		}
		f.SetTitleIfChanged(points.Comma())
	})
	current.Accessibility.LabeledBy = currentLabel
	setTargetRefKey(current, t.targetMgr, templatePanelCurrentPointsFieldRefKey)
	row.AddChild(current)
	parent.AddChild(row)
}

// targetPointsProblem returns why the given ends of a target range can't stand together, or an empty string if they
// can.
func targetPointsProblem(minimum, maximum *fxp.Int) string {
	if minimum != nil && maximum != nil && *minimum > *maximum {
		return i18n.Text("The minimum may not be more than the maximum")
	}
	return ""
}

// targetPointsMet returns true if every cost the template may end up with falls within the target. A template with no
// target always meets it.
func targetPointsMet(points, target gurps.NumericRange) bool {
	if target.Min != nil && (points.Min == nil || *points.Min < *target.Min) {
		return false
	}
	if target.Max != nil && (points.Max == nil || *points.Max > *target.Max) {
		return false
	}
	return true
}

// currentPointsTooltip returns the tooltip for the Current readout, or nil when there is nothing to say.
func currentPointsTooltip(points, target gurps.NumericRange) *unison.Panel {
	var text string
	if !points.IsSettled() {
		text = i18n.Text("The template presents choices that haven't been made yet, so what it costs depends on what is picked.")
	}
	if !targetPointsMet(points, target) {
		if text != "" {
			text += "\n\n"
		}
		text += i18n.Text("Not every outcome falls within the target of %s points.", target.Comma())
	}
	if text == "" {
		return nil
	}
	return newWrappedTooltip(text)
}

// addBodyTypeRow adds the row that decides whether applying the template replaces the body type of the sheet it is
// applied to, and with what.
func (t *Template) addBodyTypeRow(parent *unison.Panel) {
	label := NewPageLabel(i18n.Text("Body Type"))
	parent.AddChild(label)

	row := unison.NewPanel()
	row.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: 4,
	})
	button := unison.NewButton()
	button.Font = fonts.PageFieldPrimary
	button.SetTitle(t.lastBody.Name)
	button.ClickCallback = func() { ShowBodySettings(t) }
	if t.template.BodyType == nil {
		button.SetEnabled(false)
	}
	// The box is registered with the target manager, so that an undo reaches the one on the page after a rebuild.
	box := NewCheckBox(t.targetMgr, templatePanelBodyTypeFieldRefKey, "", func() check.Enum {
		return check.FromBool(t.template.BodyType != nil)
	}, func(state check.Enum) {
		if state == check.On {
			if t.lastBody == nil {
				t.lastBody = gurps.FactoryBody()
			}
			t.template.BodyType = t.lastBody
		} else {
			t.template.BodyType = nil
		}
		button.SetEnabled(state == check.On)
	})
	box.Font = fonts.PageFieldPrimary
	box.SetTitle(i18n.Text("Set to"))
	box.UndoTitle = i18n.Text("Set Body Type")
	box.Accessibility.LabeledBy = label
	row.AddChild(box)
	row.AddChild(button)
	parent.AddChild(row)
}
