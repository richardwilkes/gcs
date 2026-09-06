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
	"strconv"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

const (
	descriptionPanelFieldPrefix           = "description:"
	descriptionPanelAgeFieldRefKey        = descriptionPanelFieldPrefix + "age"
	descriptionPanelBirthdayFieldRefKey   = descriptionPanelFieldPrefix + "birthday"
	descriptionPanelEyesFieldRefKey       = descriptionPanelFieldPrefix + "eyes"
	descriptionPanelHairFieldRefKey       = descriptionPanelFieldPrefix + "hair"
	descriptionPanelSkinFieldRefKey       = descriptionPanelFieldPrefix + "skin"
	descriptionPanelHandednessFieldRefKey = descriptionPanelFieldPrefix + "handedness"
	descriptionPanelGenderFieldRefKey     = descriptionPanelFieldPrefix + "gender"
	descriptionPanelHeightFieldRefKey     = descriptionPanelFieldPrefix + "height"
	descriptionPanelWeightFieldRefKey     = descriptionPanelFieldPrefix + "weight"
)

// DescriptionPanel holds the contents of the description block on the sheet.
type DescriptionPanel struct {
	unison.Panel
	entity    *gurps.Entity
	targetMgr *TargetMgr
	prefix    string
}

// NewDescriptionPanel creates a new description panel.
func NewDescriptionPanel(entity *gurps.Entity, targetMgr *TargetMgr) *DescriptionPanel {
	d := &DescriptionPanel{
		entity:    entity,
		targetMgr: targetMgr,
		prefix:    descriptionPanelFieldPrefix,
	}
	_, layoutData := initTitledPagePanel(d, i18n.Text("Description"), 3, false, colors.TintDescription)
	layoutData.HGrab = true
	d.DrawCallback = d.drawSelf
	d.AddChild(d.createColumn1())
	d.AddChild(d.createColumn2())
	d.AddChild(d.createColumn3())
	return d
}

func (d *DescriptionPanel) drawSelf(gc *unison.Canvas, rect geom.Rect) {
	gc.DrawRect(rect, unison.ThemeBelowSurface.Paint(gc, rect, paintstyle.Fill))
	children := d.Children()
	if len(children) == 0 {
		return
	}
	column := children[0]
	children = column.Children()
	p := d.AsPanel()
	for i := 2; i < len(children); i += 4 {
		r := column.RectTo(children[i].FrameRect(), p)
		r.X = rect.X
		r.Width = rect.Width
		gc.DrawRect(r, unison.ThemeBanding.Paint(gc, r, paintstyle.Fill))
	}
}

func createColumn() *unison.Panel {
	p := unison.NewPanel()
	p.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: 4,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
	})
	return p
}

func (d *DescriptionPanel) createColumn1() *unison.Panel {
	column := createColumn()

	addRandomizedStringPageField(column, d.targetMgr, descriptionPanelGenderFieldRefKey, i18n.Text("Gender"),
		i18n.Text("Randomize the gender using the current ancestry"),
		func() string { return d.entity.Profile.Gender },
		func(s string) { d.entity.Profile.Gender = s },
		func() string { return d.entity.Ancestry().RandomGender(d.entity.Profile.Gender) })

	addRandomizedStringPageField(column, d.targetMgr, descriptionPanelAgeFieldRefKey, i18n.Text("Age"),
		i18n.Text("Randomize the age using the current ancestry"),
		func() string { return d.entity.Profile.Age },
		func(s string) { d.entity.Profile.Age = s },
		func() string {
			age, _ := strconv.Atoi(d.entity.Profile.Age) //nolint:errcheck // A default of 0 is ok here on error
			return strconv.Itoa(d.entity.Ancestry().RandomAge(d.entity, d.entity.Profile.Gender, age))
		})

	addRandomizedStringPageField(column, d.targetMgr, descriptionPanelBirthdayFieldRefKey, i18n.Text("Birthday"),
		i18n.Text("Randomize the birthday using the current calendar"),
		func() string { return d.entity.Profile.Birthday },
		func(s string) { d.entity.Profile.Birthday = s },
		func() string {
			global := gurps.GlobalSettings()
			return global.General.CalendarRef(global.Libraries).RandomBirthday(d.entity.Profile.Birthday)
		})

	religionField := addLabeledStringPageField(column, d.targetMgr, d.prefix+"religion", i18n.Text("Religion"),
		NewPageLabelEnd,
		func() string { return d.entity.Profile.Religion },
		func(s string) { d.entity.Profile.Religion = s })
	religionField.ClientData()[SkipDeepSync] = true

	return column
}

func (d *DescriptionPanel) createColumn2() *unison.Panel {
	column := createColumn()

	title := i18n.Text("Height")
	heightField := NewHeightPageField(d.targetMgr, descriptionPanelHeightFieldRefKey, title, d.entity,
		func() fxp.Length { return d.entity.Profile.Height },
		func(v fxp.Length) { d.entity.Profile.Height = v }, 0, fxp.Length(fxp.Max), true)
	addRandomizedPageField(column, heightField, title, i18n.Text("Randomize the height using the current ancestry"),
		func() string {
			d.entity.Profile.Height = d.entity.Ancestry().RandomHeight(d.entity, d.entity.Profile.Gender, d.entity.Profile.Height)
			return heightField.Format(d.entity.Profile.Height)
		})

	title = i18n.Text("Weight")
	weightField := NewWeightPageField(d.targetMgr, descriptionPanelWeightFieldRefKey, title, d.entity,
		func() fxp.Weight { return d.entity.Profile.Weight },
		func(v fxp.Weight) { d.entity.Profile.Weight = v }, 0, fxp.Weight(fxp.Max), true)
	addRandomizedPageField(column, weightField, title, i18n.Text("Randomize the weight using the current ancestry"),
		func() string {
			d.entity.Profile.Weight = d.entity.Ancestry().RandomWeight(d.entity, d.entity.Profile.Gender, d.entity.Profile.Weight)
			return weightField.Format(d.entity.Profile.Weight)
		})

	title = i18n.Text("Size")
	column.AddChild(NewPageLabelEnd(title))
	field := NewIntegerPageField(d.targetMgr, d.prefix+"size", title,
		func() int { return d.entity.Profile.AdjustedSizeModifier() },
		func(v int) { d.entity.Profile.SetAdjustedSizeModifier(v) }, -99, 99, true, false)
	field.HAlign = align.Start
	column.AddChild(field)

	tlField := addLabeledStringPageField(column, d.targetMgr, d.prefix+"tl", i18n.Text("TL"), NewPageLabelEnd,
		func() string { return d.entity.Profile.TechLevel },
		func(s string) { d.entity.Profile.TechLevel = s })
	tlField.Tooltip = newWrappedTooltip(gurps.TechLevelInfo())

	return column
}

func (d *DescriptionPanel) createColumn3() *unison.Panel {
	column := createColumn()

	addRandomizedStringPageField(column, d.targetMgr, descriptionPanelHairFieldRefKey, i18n.Text("Hair"),
		i18n.Text("Randomize the hair using the current ancestry"),
		func() string { return d.entity.Profile.Hair },
		func(s string) { d.entity.Profile.Hair = s },
		func() string { return d.entity.Ancestry().RandomHair(d.entity.Profile.Gender, d.entity.Profile.Hair) })

	addRandomizedStringPageField(column, d.targetMgr, descriptionPanelEyesFieldRefKey, i18n.Text("Eyes"),
		i18n.Text("Randomize the eyes using the current ancestry"),
		func() string { return d.entity.Profile.Eyes },
		func(s string) { d.entity.Profile.Eyes = s },
		func() string { return d.entity.Ancestry().RandomEyes(d.entity.Profile.Gender, d.entity.Profile.Eyes) })

	addRandomizedStringPageField(column, d.targetMgr, descriptionPanelSkinFieldRefKey, i18n.Text("Skin"),
		i18n.Text("Randomize the skin using the current ancestry"),
		func() string { return d.entity.Profile.Skin },
		func(s string) { d.entity.Profile.Skin = s },
		func() string { return d.entity.Ancestry().RandomSkin(d.entity.Profile.Gender, d.entity.Profile.Skin) })

	addRandomizedStringPageField(column, d.targetMgr, descriptionPanelHandednessFieldRefKey, i18n.Text("Hand"),
		i18n.Text("Randomize the handedness using the current ancestry"),
		func() string { return d.entity.Profile.Handedness },
		func(s string) { d.entity.Profile.Handedness = s },
		func() string {
			return d.entity.Ancestry().RandomHandedness(d.entity.Profile.Gender, d.entity.Profile.Handedness)
		})

	return column
}
