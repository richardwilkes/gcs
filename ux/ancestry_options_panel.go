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
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// weightedStringListSpec describes a weighted string list an editor shows: its labels and undo titles, and the list
// itself, bound by pointer so that appends and removals reach the model. When importer is set, the list's header gets
// an import button that calls it.
type weightedStringListSpec struct {
	title                 string
	tooltip               string
	valueTooltip          string
	addTooltip            string
	removeTooltip         string
	removeSelectedTooltip string
	importTooltip         string
	addTitle              string
	removeTitle           string
	removeSelectedTitle   string
	dragTitle             string
	importer              func()
	list                  *[]*gurps.WeightedStringOption
}

// ancestryOptionValueTooltip returns the tooltip of the value field in each of an options block's weighted string
// lists.
func ancestryOptionValueTooltip() string {
	return i18n.Text("The text placed on the character sheet when this option is chosen")
}

// ancestryOptionsPanel edits one AncestryOptions block. The same panel serves the common options and each gender: when
// weighted is not nil, the block belongs to that gender and the panel starts with the gender's weight and name, so that
// every label in a gender row shares one label column.
type ancestryOptionsPanel struct {
	unison.Panel
	dockable *ancestryEditorDockable
	options  *gurps.AncestryOptions
	weighted *gurps.WeightedAncestryOptions
}

func newAncestryOptionsPanel(d *ancestryEditorDockable, options *gurps.AncestryOptions, weighted *gurps.WeightedAncestryOptions) *ancestryOptionsPanel {
	p := &ancestryOptionsPanel{
		dockable: d,
		options:  options,
		weighted: weighted,
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	p.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	if weighted != nil {
		p.addGenderFields()
	}
	p.addScriptFields()

	for _, spec := range weightedStringListSpecs(options) {
		list := newWeightedStringOptionsPanel(d, spec)
		list.SetLayoutData(&unison.FlexLayoutData{HSpan: 2, HAlign: align.Fill, HGrab: true})
		p.AddChild(list)
	}
	p.AddChild(newNameGeneratorsPanel(d, options))
	return p
}

// addGenderFields adds the weight and name of the gender this block belongs to.
func (p *ancestryOptionsPanel) addGenderFields() {
	d := p.dockable
	text := i18n.Text("Weight")
	p.AddChild(NewFieldLeadingLabel(text, false))
	weightField := NewIntegerField(d.targetMgr, p.weighted.KeyPrefix+"weight", text,
		func() int { return p.weighted.Weight },
		func(v int) { p.weighted.Weight = v }, 0, 9999, false, false)
	weightField.SetBaseTooltip(newWrappedTooltip(i18n.Text("The relative likelihood of this gender being chosen when a random gender is generated. Only the ratio between the weights of the genders matters.")))
	p.AddChild(weightField)

	text = i18n.Text("Name")
	p.AddChild(NewFieldLeadingLabel(text, false))
	nameField := NewStringField(d.targetMgr, p.options.KeyPrefix+"name", text,
		func() string { return p.options.Name },
		func(s string) { p.options.Name = s })
	nameField.SetMinimumTextWidthUsing(prototypeMinIDWidth)
	nameField.Tooltip = newWrappedTooltip(i18n.Text("The name of this gender, as it will appear on the character sheet when chosen"))
	p.AddChild(nameField)
}

// addScriptFields adds the height, weight and age scripts. Their labels say what they are scripts for, so that the
// gender's own Weight above is not followed by a second bare "Weight".
func (p *ancestryOptionsPanel) addScriptFields() {
	mgr := p.dockable.targetMgr
	o := p.options
	text := i18n.Text("Height Script")
	p.AddChild(NewFieldLeadingLabel(text, false))
	addScriptField(p.AsPanel(), mgr, o.KeyPrefix+"height_script", text,
		i18n.Text("A script that produces a height in inches, such as entity.randomHeightInInches($st). A gender leaves this empty to use the common options; when neither provides a script, 64 inches is used."),
		func() string { return o.HeightScript },
		func(s string) { o.HeightScript = s }, false)

	text = i18n.Text("Weight Script")
	p.AddChild(NewFieldLeadingLabel(text, false))
	addScriptField(p.AsPanel(), mgr, o.KeyPrefix+"weight_script", text,
		i18n.Text("A script that produces a weight in pounds, such as entity.randomWeightInPounds($st). A gender leaves this empty to use the common options; when neither provides a script, 140 pounds is used."),
		func() string { return o.WeightScript },
		func(s string) { o.WeightScript = s }, false)

	text = i18n.Text("Age Script")
	p.AddChild(NewFieldLeadingLabel(text, false))
	addScriptField(p.AsPanel(), mgr, o.KeyPrefix+"age_script", text,
		i18n.Text("A script that produces an age in years, such as dice.roll(\"1d102+140\"). A gender leaves this empty to use the common options; when neither provides a script, 18 is used."),
		func() string { return o.AgeScript },
		func(s string) { o.AgeScript = s }, false)
}

// weightedStringListSpecs returns the specs for the four weighted string lists of an options block, in the order they
// are shown, one below another: hair, eyes, skin and handedness.
func weightedStringListSpecs(options *gurps.AncestryOptions) []*weightedStringListSpec {
	return []*weightedStringListSpec{
		{
			title:                 i18n.Text("Hair"),
			tooltip:               i18n.Text("The possible hair descriptions. A gender with an empty list uses the common options; when both are empty, Brown is used."),
			valueTooltip:          ancestryOptionValueTooltip(),
			addTooltip:            i18n.Text("Add hair option"),
			removeTooltip:         i18n.Text("Remove hair option"),
			removeSelectedTooltip: i18n.Text("Remove the selected hair options"),
			addTitle:              i18n.Text("Add Hair Option"),
			removeTitle:           i18n.Text("Remove Hair Option"),
			removeSelectedTitle:   i18n.Text("Remove Selected Hair Options"),
			dragTitle:             i18n.Text("Hair Option Drag"),
			list:                  &options.HairOptions,
		},
		{
			title:                 i18n.Text("Eyes"),
			tooltip:               i18n.Text("The possible eye descriptions. A gender with an empty list uses the common options; when both are empty, Brown is used."),
			valueTooltip:          ancestryOptionValueTooltip(),
			addTooltip:            i18n.Text("Add eye option"),
			removeTooltip:         i18n.Text("Remove eye option"),
			removeSelectedTooltip: i18n.Text("Remove the selected eye options"),
			addTitle:              i18n.Text("Add Eye Option"),
			removeTitle:           i18n.Text("Remove Eye Option"),
			removeSelectedTitle:   i18n.Text("Remove Selected Eye Options"),
			dragTitle:             i18n.Text("Eye Option Drag"),
			list:                  &options.EyeOptions,
		},
		{
			title:                 i18n.Text("Skin"),
			tooltip:               i18n.Text("The possible skin descriptions. A gender with an empty list uses the common options; when both are empty, Brown is used."),
			valueTooltip:          ancestryOptionValueTooltip(),
			addTooltip:            i18n.Text("Add skin option"),
			removeTooltip:         i18n.Text("Remove skin option"),
			removeSelectedTooltip: i18n.Text("Remove the selected skin options"),
			addTitle:              i18n.Text("Add Skin Option"),
			removeTitle:           i18n.Text("Remove Skin Option"),
			removeSelectedTitle:   i18n.Text("Remove Selected Skin Options"),
			dragTitle:             i18n.Text("Skin Option Drag"),
			list:                  &options.SkinOptions,
		},
		{
			title:                 i18n.Text("Handedness"),
			tooltip:               i18n.Text("The possible handedness descriptions. A gender with an empty list uses the common options; when both are empty, Right is used."),
			valueTooltip:          ancestryOptionValueTooltip(),
			addTooltip:            i18n.Text("Add handedness option"),
			removeTooltip:         i18n.Text("Remove handedness option"),
			removeSelectedTooltip: i18n.Text("Remove the selected handedness options"),
			addTitle:              i18n.Text("Add Handedness Option"),
			removeTitle:           i18n.Text("Remove Handedness Option"),
			removeSelectedTitle:   i18n.Text("Remove Selected Handedness Options"),
			dragTitle:             i18n.Text("Handedness Option Drag"),
			list:                  &options.HandednessOptions,
		},
	}
}
