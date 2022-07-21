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

package sheet

import (
	"strconv"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/measure"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// DescriptionPanel holds the contents of the description block on the sheet.
type DescriptionPanel struct {
	unison.Panel
	entity *gurps.Entity
}

// NewDescriptionPanel creates a new description panel.
func NewDescriptionPanel(entity *gurps.Entity) *DescriptionPanel {
	d := &DescriptionPanel{entity: entity}
	d.Self = d
	d.SetLayout(&unison.FlexLayout{
		Columns:  3,
		HSpacing: 4,
	})
	d.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: unison.FillAlignment,
		VAlign: unison.StartAlignment,
		HGrab:  true,
	})
	d.SetBorder(unison.NewCompoundBorder(&widget.TitledBorder{Title: i18n.Text("Description")}, unison.NewEmptyBorder(unison.Insets{
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
	genderField := widget.NewStringPageField(nil, "", title,
		func() string { return d.entity.Profile.Gender },
		func(s string) { d.entity.Profile.Gender = s })
	column.AddChild(widget.NewPageLabelWithRandomizer(title,
		i18n.Text("Randomize the gender using the current ancestry"), func() {
			d.entity.Profile.Gender = d.entity.Ancestry().RandomGender(d.entity.Profile.Gender)
			SetTextAndMarkModified(genderField.Field, d.entity.Profile.Gender)
		}))
	column.AddChild(genderField)

	title = i18n.Text("Age")
	ageField := widget.NewStringPageField(nil, "", title,
		func() string { return d.entity.Profile.Age },
		func(s string) { d.entity.Profile.Age = s })
	column.AddChild(widget.NewPageLabelWithRandomizer(title,
		i18n.Text("Randomize the age using the current ancestry"), func() {
			age, _ := strconv.Atoi(d.entity.Profile.Age) //nolint:errcheck // A default of 0 is ok here on error
			d.entity.Profile.Age = strconv.Itoa(d.entity.Ancestry().RandomAge(d.entity, d.entity.Profile.Gender, age))
			SetTextAndMarkModified(ageField.Field, d.entity.Profile.Age)
		}))
	column.AddChild(ageField)

	title = i18n.Text("Birthday")
	birthdayField := widget.NewStringPageField(nil, "", title,
		func() string { return d.entity.Profile.Birthday },
		func(s string) { d.entity.Profile.Birthday = s })
	column.AddChild(widget.NewPageLabelWithRandomizer(title,
		i18n.Text("Randomize the birthday using the current calendar"), func() {
			global := settings.Global()
			d.entity.Profile.Birthday = global.General.CalendarRef(global.LibrarySet).RandomBirthday(d.entity.Profile.Birthday)
			SetTextAndMarkModified(birthdayField.Field, d.entity.Profile.Birthday)
		}))
	column.AddChild(birthdayField)

	title = i18n.Text("Religion")
	column.AddChild(widget.NewPageLabelEnd(title))
	column.AddChild(widget.NewStringPageField(nil, "", title,
		func() string { return d.entity.Profile.Religion },
		func(s string) { d.entity.Profile.Religion = s }))

	return column
}

func (d *DescriptionPanel) createColumn2() *unison.Panel {
	column := createColumn()

	title := i18n.Text("Height")
	heightField := widget.NewHeightPageField(nil, "", title, d.entity,
		func() measure.Length { return d.entity.Profile.Height },
		func(v measure.Length) { d.entity.Profile.Height = v }, 0, measure.Length(fxp.Max), true)
	column.AddChild(widget.NewPageLabelWithRandomizer(title,
		i18n.Text("Randomize the height using the current ancestry"), func() {
			d.entity.Profile.Height = d.entity.Ancestry().RandomHeight(d.entity, d.entity.Profile.Gender, d.entity.Profile.Height)
			SetTextAndMarkModified(heightField.Field, d.entity.Profile.Height.String())
		}))
	column.AddChild(heightField)

	title = i18n.Text("Weight")
	weightField := widget.NewWeightPageField(nil, "", title, d.entity,
		func() measure.Weight { return d.entity.Profile.Weight },
		func(v measure.Weight) { d.entity.Profile.Weight = v }, 0, measure.Weight(fxp.Max), true)
	column.AddChild(widget.NewPageLabelWithRandomizer(title,
		i18n.Text("Randomize the weight using the current ancestry"), func() {
			d.entity.Profile.Weight = d.entity.Ancestry().RandomWeight(d.entity, d.entity.Profile.Gender, d.entity.Profile.Weight)
			SetTextAndMarkModified(weightField.Field, d.entity.Profile.Weight.String())
		}))
	column.AddChild(weightField)

	title = i18n.Text("Size")
	column.AddChild(widget.NewPageLabelEnd(title))
	field := widget.NewIntegerPageField(nil, "", title,
		func() int { return d.entity.Profile.AdjustedSizeModifier() },
		func(v int) { d.entity.Profile.SetAdjustedSizeModifier(v) }, -99, 99, true)
	field.HAlign = unison.StartAlignment
	column.AddChild(field)

	title = i18n.Text("TL")
	column.AddChild(widget.NewPageLabelEnd(title))
	tlField := widget.NewStringPageField(nil, "", title,
		func() string { return d.entity.Profile.TechLevel },
		func(s string) { d.entity.Profile.TechLevel = s })
	tlField.Tooltip = unison.NewTooltipWithText(gurps.TechLevelInfo)
	column.AddChild(tlField)

	return column
}

func (d *DescriptionPanel) createColumn3() *unison.Panel {
	column := createColumn()

	title := i18n.Text("Hair")
	hairField := widget.NewStringPageField(nil, "", title,
		func() string { return d.entity.Profile.Hair },
		func(s string) { d.entity.Profile.Hair = s })
	column.AddChild(widget.NewPageLabelWithRandomizer(title,
		i18n.Text("Randomize the hair using the current ancestry"), func() {
			d.entity.Profile.Hair = d.entity.Ancestry().RandomHair(d.entity.Profile.Gender, d.entity.Profile.Hair)
			SetTextAndMarkModified(hairField.Field, d.entity.Profile.Hair)
		}))
	column.AddChild(hairField)

	title = i18n.Text("Eyes")
	eyesField := widget.NewStringPageField(nil, "", title,
		func() string { return d.entity.Profile.Eyes },
		func(s string) { d.entity.Profile.Eyes = s })
	column.AddChild(widget.NewPageLabelWithRandomizer(title,
		i18n.Text("Randomize the eyes using the current ancestry"), func() {
			d.entity.Profile.Eyes = d.entity.Ancestry().RandomEyes(d.entity.Profile.Gender, d.entity.Profile.Eyes)
			SetTextAndMarkModified(eyesField.Field, d.entity.Profile.Eyes)
		}))
	column.AddChild(eyesField)

	title = i18n.Text("Skin")
	skinField := widget.NewStringPageField(nil, "", title,
		func() string { return d.entity.Profile.Skin },
		func(s string) { d.entity.Profile.Skin = s })
	column.AddChild(widget.NewPageLabelWithRandomizer(title,
		i18n.Text("Randomize the skin using the current ancestry"), func() {
			d.entity.Profile.Skin = d.entity.Ancestry().RandomSkin(d.entity.Profile.Gender, d.entity.Profile.Skin)
			SetTextAndMarkModified(skinField.Field, d.entity.Profile.Skin)
		}))
	column.AddChild(skinField)

	title = i18n.Text("Hand")
	handField := widget.NewStringPageField(nil, "", title,
		func() string { return d.entity.Profile.Handedness },
		func(s string) { d.entity.Profile.Handedness = s })
	column.AddChild(widget.NewPageLabelWithRandomizer(title,
		i18n.Text("Randomize the handedness using the current ancestry"), func() {
			d.entity.Profile.Handedness = d.entity.Ancestry().RandomHandedness(d.entity.Profile.Gender, d.entity.Profile.Handedness)
			SetTextAndMarkModified(handField.Field, d.entity.Profile.Handedness)
		}))
	column.AddChild(handField)

	return column
}
