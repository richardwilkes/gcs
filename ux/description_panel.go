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
	"strconv"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
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
		prefix:    targetMgr.NextPrefix(),
	}
	d.Self = d
	d.SetLayout(&unison.FlexLayout{
		Columns:  3,
		HSpacing: 4,
	})
	d.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	d.SetBorder(unison.NewCompoundBorder(&TitledBorder{Title: i18n.Text("Description")}, unison.NewEmptyBorder(unison.Insets{
		Top:    1,
		Left:   2,
		Bottom: 1,
		Right:  2,
	})))
	d.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, unison.ContentColor.Paint(gc, rect, unison.Fill))
	}
	d.AddChild(d.createColumn1())
	d.AddChild(d.createColumn2())
	d.AddChild(d.createColumn3())
	return d
}

func createColumn() *unison.Panel {
	p := unison.NewPanel()
	p.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: 4,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
	})
	return p
}

func (d *DescriptionPanel) createColumn1() *unison.Panel {
	column := createColumn()

	title := i18n.Text("Gender")
	genderField := NewStringPageField(d.targetMgr, d.prefix+"gender", title,
		func() string { return d.entity.Profile.Gender },
		func(s string) { d.entity.Profile.Gender = s })
	column.AddChild(NewPageLabelWithRandomizer(title,
		i18n.Text("Randomize the gender using the current ancestry"), func() {
			d.entity.Profile.Gender = d.entity.Ancestry().RandomGender(d.entity.Profile.Gender)
			SetTextAndMarkModified(genderField.Field, d.entity.Profile.Gender)
		}))
	genderField.ClientData()[SkipDeepSync] = true
	column.AddChild(genderField)

	title = i18n.Text("Age")
	ageField := NewStringPageField(d.targetMgr, d.prefix+"age", title,
		func() string { return d.entity.Profile.Age },
		func(s string) { d.entity.Profile.Age = s })
	column.AddChild(NewPageLabelWithRandomizer(title,
		i18n.Text("Randomize the age using the current ancestry"), func() {
			age, _ := strconv.Atoi(d.entity.Profile.Age) //nolint:errcheck // A default of 0 is ok here on error
			d.entity.Profile.Age = strconv.Itoa(d.entity.Ancestry().RandomAge(d.entity, d.entity.Profile.Gender, age))
			SetTextAndMarkModified(ageField.Field, d.entity.Profile.Age)
		}))
	ageField.ClientData()[SkipDeepSync] = true
	column.AddChild(ageField)

	title = i18n.Text("Birthday")
	birthdayField := NewStringPageField(d.targetMgr, d.prefix+"birthday", title,
		func() string { return d.entity.Profile.Birthday },
		func(s string) { d.entity.Profile.Birthday = s })
	column.AddChild(NewPageLabelWithRandomizer(title,
		i18n.Text("Randomize the birthday using the current calendar"), func() {
			global := gurps.GlobalSettings()
			d.entity.Profile.Birthday = global.General.CalendarRef(global.LibrarySet).RandomBirthday(d.entity.Profile.Birthday)
			SetTextAndMarkModified(birthdayField.Field, d.entity.Profile.Birthday)
		}))
	birthdayField.ClientData()[SkipDeepSync] = true
	column.AddChild(birthdayField)

	title = i18n.Text("Religion")
	column.AddChild(NewPageLabelEnd(title))
	religionField := NewStringPageField(d.targetMgr, d.prefix+"religion", title,
		func() string { return d.entity.Profile.Religion },
		func(s string) { d.entity.Profile.Religion = s })
	religionField.ClientData()[SkipDeepSync] = true
	column.AddChild(religionField)

	return column
}

func (d *DescriptionPanel) createColumn2() *unison.Panel {
	column := createColumn()

	title := i18n.Text("Height")
	heightField := NewHeightPageField(d.targetMgr, d.prefix+"height", title, d.entity,
		func() gurps.Length { return d.entity.Profile.Height },
		func(v gurps.Length) { d.entity.Profile.Height = v }, 0, gurps.Length(fxp.Max), true)
	column.AddChild(NewPageLabelWithRandomizer(title,
		i18n.Text("Randomize the height using the current ancestry"), func() {
			d.entity.Profile.Height = d.entity.Ancestry().RandomHeight(d.entity, d.entity.Profile.Gender, d.entity.Profile.Height)
			SetTextAndMarkModified(heightField.Field, d.entity.Profile.Height.String())
		}))
	heightField.ClientData()[SkipDeepSync] = true
	column.AddChild(heightField)

	title = i18n.Text("Weight")
	weightField := NewWeightPageField(d.targetMgr, d.prefix+"weight", title, d.entity,
		func() gurps.Weight { return d.entity.Profile.Weight },
		func(v gurps.Weight) { d.entity.Profile.Weight = v }, 0, gurps.Weight(fxp.Max), true)
	column.AddChild(NewPageLabelWithRandomizer(title,
		i18n.Text("Randomize the weight using the current ancestry"), func() {
			d.entity.Profile.Weight = d.entity.Ancestry().RandomWeight(d.entity, d.entity.Profile.Gender, d.entity.Profile.Weight)
			SetTextAndMarkModified(weightField.Field, d.entity.Profile.Weight.String())
		}))
	weightField.ClientData()[SkipDeepSync] = true
	column.AddChild(weightField)

	title = i18n.Text("Size")
	column.AddChild(NewPageLabelEnd(title))
	field := NewIntegerPageField(d.targetMgr, d.prefix+"size", title,
		func() int { return d.entity.Profile.AdjustedSizeModifier() },
		func(v int) { d.entity.Profile.SetAdjustedSizeModifier(v) }, -99, 99, true, false)
	field.HAlign = unison.StartAlignment
	column.AddChild(field)

	title = i18n.Text("TL")
	column.AddChild(NewPageLabelEnd(title))
	tlField := NewStringPageField(d.targetMgr, d.prefix+"tl", title,
		func() string { return d.entity.Profile.TechLevel },
		func(s string) { d.entity.Profile.TechLevel = s })
	tlField.Tooltip = unison.NewTooltipWithText(techLevelInfo())
	column.AddChild(tlField)

	return column
}

func (d *DescriptionPanel) createColumn3() *unison.Panel {
	column := createColumn()

	title := i18n.Text("Hair")
	hairField := NewStringPageField(d.targetMgr, d.prefix+"hair", title,
		func() string { return d.entity.Profile.Hair },
		func(s string) { d.entity.Profile.Hair = s })
	column.AddChild(NewPageLabelWithRandomizer(title,
		i18n.Text("Randomize the hair using the current ancestry"), func() {
			d.entity.Profile.Hair = d.entity.Ancestry().RandomHair(d.entity.Profile.Gender, d.entity.Profile.Hair)
			SetTextAndMarkModified(hairField.Field, d.entity.Profile.Hair)
		}))
	hairField.ClientData()[SkipDeepSync] = true
	column.AddChild(hairField)

	title = i18n.Text("Eyes")
	eyesField := NewStringPageField(d.targetMgr, d.prefix+"eyes", title,
		func() string { return d.entity.Profile.Eyes },
		func(s string) { d.entity.Profile.Eyes = s })
	column.AddChild(NewPageLabelWithRandomizer(title,
		i18n.Text("Randomize the eyes using the current ancestry"), func() {
			d.entity.Profile.Eyes = d.entity.Ancestry().RandomEyes(d.entity.Profile.Gender, d.entity.Profile.Eyes)
			SetTextAndMarkModified(eyesField.Field, d.entity.Profile.Eyes)
		}))
	eyesField.ClientData()[SkipDeepSync] = true
	column.AddChild(eyesField)

	title = i18n.Text("Skin")
	skinField := NewStringPageField(d.targetMgr, d.prefix+"skin", title,
		func() string { return d.entity.Profile.Skin },
		func(s string) { d.entity.Profile.Skin = s })
	column.AddChild(NewPageLabelWithRandomizer(title,
		i18n.Text("Randomize the skin using the current ancestry"), func() {
			d.entity.Profile.Skin = d.entity.Ancestry().RandomSkin(d.entity.Profile.Gender, d.entity.Profile.Skin)
			SetTextAndMarkModified(skinField.Field, d.entity.Profile.Skin)
		}))
	skinField.ClientData()[SkipDeepSync] = true
	column.AddChild(skinField)

	title = i18n.Text("Hand")
	handField := NewStringPageField(d.targetMgr, d.prefix+"hand", title,
		func() string { return d.entity.Profile.Handedness },
		func(s string) { d.entity.Profile.Handedness = s })
	column.AddChild(NewPageLabelWithRandomizer(title,
		i18n.Text("Randomize the handedness using the current ancestry"), func() {
			d.entity.Profile.Handedness = d.entity.Ancestry().RandomHandedness(d.entity.Profile.Gender, d.entity.Profile.Handedness)
			SetTextAndMarkModified(handField.Field, d.entity.Profile.Handedness)
		}))
	handField.ClientData()[SkipDeepSync] = true
	column.AddChild(handField)

	return column
}
