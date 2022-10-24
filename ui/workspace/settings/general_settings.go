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

package settings

import (
	"io/fs"
	"strconv"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	gsettings "github.com/richardwilkes/gcs/v5/model/gurps/settings"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/gcs/v5/ui/workspace"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jotrotate"
	"github.com/richardwilkes/unison"
)

type generalSettingsDockable struct {
	Dockable
	nameField                     *widget.StringField
	autoFillProfileCheckbox       *widget.CheckBox
	autoAddNaturalAttacksCheckbox *widget.CheckBox
	pointsField                   *widget.DecimalField
	techLevelField                *widget.StringField
	calendarPopup                 *unison.PopupMenu[string]
	initialListScaleField         *widget.PercentageField
	initialEditorScaleField       *widget.PercentageField
	initialSheetScaleField        *widget.PercentageField
	maxAutoColWidthField          *widget.IntegerField
	exportResolutionField         *widget.IntegerField
	tooltipDelayField             *widget.DecimalField
	tooltipDismissalField         *widget.DecimalField
	scrollWheelMultiplierField    *widget.DecimalField
	externalPDFCmdlineField       *widget.StringField
}

// ShowGeneralSettings the General Settings window.
func ShowGeneralSettings() {
	ws, dc, found := workspace.Activate(func(d unison.Dockable) bool {
		_, ok := d.(*generalSettingsDockable)
		return ok
	})
	if !found && ws != nil {
		d := &generalSettingsDockable{}
		d.Self = d
		d.TabTitle = i18n.Text("General Settings")
		d.TabIcon = res.SettingsSVG
		d.Extensions = []string{library.GeneralSettingsExt}
		d.Loader = d.load
		d.Saver = d.save
		d.Resetter = d.reset
		d.Setup(ws, dc, nil, nil, d.initContent)
		d.nameField.RequestFocus()
	}
}

func (d *generalSettingsDockable) initContent(content *unison.Panel) {
	content.SetLayout(&unison.FlexLayout{
		Columns:  3,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	d.createPlayerAndDescFields(content)
	d.createCheckboxBlock(content)
	d.createInitialPointsFields(content)
	d.createTechLevelField(content)
	d.createCalendarPopup(content)
	initialListScaleTitle := i18n.Text("Initial List Scale")
	content.AddChild(widget.NewFieldLeadingLabel(initialListScaleTitle))
	d.initialListScaleField = widget.NewPercentageField(nil, "", initialListScaleTitle,
		func() int { return settings.Global().General.InitialListUIScale },
		func(v int) { settings.Global().General.InitialListUIScale = v },
		gsettings.InitialUIScaleMin, gsettings.InitialUIScaleMax, false, false)
	content.AddChild(widget.WrapWithSpan(2, d.initialListScaleField))
	initialEditorScaleTitle := i18n.Text("Initial Editor Scale")
	content.AddChild(widget.NewFieldLeadingLabel(initialEditorScaleTitle))
	d.initialEditorScaleField = widget.NewPercentageField(nil, "", initialEditorScaleTitle,
		func() int { return settings.Global().General.InitialEditorUIScale },
		func(v int) { settings.Global().General.InitialEditorUIScale = v },
		gsettings.InitialUIScaleMin, gsettings.InitialUIScaleMax, false, false)
	content.AddChild(widget.WrapWithSpan(2, d.initialEditorScaleField))
	initialSheetScaleTitle := i18n.Text("Initial Sheet Scale")
	content.AddChild(widget.NewFieldLeadingLabel(initialSheetScaleTitle))
	d.initialSheetScaleField = widget.NewPercentageField(nil, "", initialSheetScaleTitle,
		func() int { return settings.Global().General.InitialSheetUIScale },
		func(v int) { settings.Global().General.InitialSheetUIScale = v },
		gsettings.InitialUIScaleMin, gsettings.InitialUIScaleMax, false, false)
	content.AddChild(widget.WrapWithSpan(2, d.initialSheetScaleField))
	d.createCellAutoMaxWidthField(content)
	d.createImageResolutionField(content)
	d.createTooltipDelayField(content)
	d.createTooltipDismissalField(content)
	d.createScrollWheelMultiplierField(content)
	d.createPathInfoField(content, i18n.Text("Settings Path"), settings.Path())
	d.createPathInfoField(content, i18n.Text("Translations Path"), i18n.Dir)
	d.createPathInfoField(content, i18n.Text("Log Path"), jotrotate.PathToLog)
	d.createExternalPDFCmdLineField(content)
}

func (d *generalSettingsDockable) createPlayerAndDescFields(content *unison.Panel) {
	title := i18n.Text("Default Player Name")
	content.AddChild(widget.NewFieldLeadingLabel(title))
	d.nameField = widget.NewStringField(nil, "", title,
		func() string { return settings.Global().General.DefaultPlayerName },
		func(s string) { settings.Global().General.DefaultPlayerName = s })
	d.nameField.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	content.AddChild(d.nameField)
}

func (d *generalSettingsDockable) createCheckboxBlock(content *unison.Panel) {
	d.autoFillProfileCheckbox = widget.NewCheckBox(nil, "", i18n.Text("Fill in initial description"),
		func() unison.CheckState { return unison.CheckStateFromBool(settings.Global().General.AutoFillProfile) },
		func(state unison.CheckState) {
			settings.Global().General.AutoFillProfile = state == unison.OnCheckState
		})
	d.autoFillProfileCheckbox.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	content.AddChild(widget.NewFieldLeadingLabel(""))
	content.AddChild(d.autoFillProfileCheckbox)

	d.autoAddNaturalAttacksCheckbox = widget.NewCheckBox(nil, "", i18n.Text("Add natural attacks to new sheets"),
		func() unison.CheckState {
			return unison.CheckStateFromBool(settings.Global().General.AutoAddNaturalAttacks)
		},
		func(state unison.CheckState) {
			settings.Global().General.AutoAddNaturalAttacks = state == unison.OnCheckState
		})
	d.autoAddNaturalAttacksCheckbox.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	content.AddChild(widget.NewFieldLeadingLabel(""))
	content.AddChild(d.autoAddNaturalAttacksCheckbox)
}

func (d *generalSettingsDockable) createInitialPointsFields(content *unison.Panel) {
	title := i18n.Text("Initial Points")
	content.AddChild(widget.NewFieldLeadingLabel(title))
	d.pointsField = widget.NewDecimalField(nil, "", title,
		func() fxp.Int { return settings.Global().General.InitialPoints },
		func(v fxp.Int) { settings.Global().General.InitialPoints = v }, gsettings.InitialPointsMin,
		gsettings.InitialPointsMax, false, false)
	d.pointsField.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	content.AddChild(d.pointsField)
}

func (d *generalSettingsDockable) createTechLevelField(content *unison.Panel) {
	title := i18n.Text("Default Tech Level")
	content.AddChild(widget.NewFieldLeadingLabel(title))
	d.techLevelField = widget.NewStringField(nil, "", title,
		func() string { return settings.Global().General.DefaultTechLevel },
		func(s string) { settings.Global().General.DefaultTechLevel = s })
	d.techLevelField.Tooltip = unison.NewTooltipWithText(gurps.TechLevelInfo)
	d.techLevelField.SetMinimumTextWidthUsing("12^")
	d.techLevelField.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	content.AddChild(d.techLevelField)
}

func (d *generalSettingsDockable) createCalendarPopup(content *unison.Panel) {
	content.AddChild(widget.NewFieldLeadingLabel(i18n.Text("Calendar")))
	d.calendarPopup = unison.NewPopupMenu[string]()
	libraries := settings.Global().Libraries()
	for _, lib := range gsettings.AvailableCalendarRefs(libraries) {
		d.calendarPopup.AddDisabledItem(lib.Name)
		for _, one := range lib.List {
			d.calendarPopup.AddItem(one.Name)
		}
	}
	d.calendarPopup.Select(settings.Global().General.CalendarRef(libraries).Name)
	d.calendarPopup.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	d.calendarPopup.SelectionCallback = func(_ int, item string) {
		settings.Global().General.CalendarName = item
	}
	content.AddChild(d.calendarPopup)
}

func (d *generalSettingsDockable) createCellAutoMaxWidthField(content *unison.Panel) {
	title := i18n.Text("Max Auto Column Width")
	content.AddChild(widget.NewFieldLeadingLabel(title))
	d.maxAutoColWidthField = widget.NewIntegerField(nil, "", title,
		func() int { return settings.Global().General.MaximumAutoColWidth },
		func(v int) { settings.Global().General.MaximumAutoColWidth = v },
		gsettings.AutoColWidthMin, gsettings.AutoColWidthMax, false, false)
	d.maxAutoColWidthField.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	content.AddChild(d.maxAutoColWidthField)
}

func (d *generalSettingsDockable) createImageResolutionField(content *unison.Panel) {
	title := i18n.Text("Image Export Resolution")
	content.AddChild(widget.NewFieldLeadingLabel(title))
	d.exportResolutionField = widget.NewIntegerField(nil, "", title,
		func() int { return settings.Global().General.ImageResolution },
		func(v int) { settings.Global().General.ImageResolution = v },
		gsettings.ImageResolutionMin, gsettings.ImageResolutionMax, false, false)
	content.AddChild(widget.WrapWithSpan(2, d.exportResolutionField, widget.NewFieldTrailingLabel(i18n.Text("ppi"))))
}

func (d *generalSettingsDockable) createTooltipDelayField(content *unison.Panel) {
	title := i18n.Text("Tooltip Delay")
	content.AddChild(widget.NewFieldLeadingLabel(title))
	d.tooltipDelayField = widget.NewDecimalField(nil, "", title,
		func() fxp.Int { return settings.Global().General.TooltipDelay },
		func(v fxp.Int) {
			general := settings.Global().General
			general.TooltipDelay = v
			general.UpdateToolTipTiming()
		}, gsettings.TooltipDelayMin, gsettings.TooltipDelayMax, false, false)
	content.AddChild(widget.WrapWithSpan(2, d.tooltipDelayField, widget.NewFieldTrailingLabel(i18n.Text("seconds"))))
}

func (d *generalSettingsDockable) createTooltipDismissalField(content *unison.Panel) {
	title := i18n.Text("Tooltip Dismissal")
	content.AddChild(widget.NewFieldLeadingLabel(title))
	d.tooltipDismissalField = widget.NewDecimalField(nil, "", title,
		func() fxp.Int { return settings.Global().General.TooltipDismissal },
		func(v fxp.Int) {
			general := settings.Global().General
			general.TooltipDismissal = v
			general.UpdateToolTipTiming()
		}, gsettings.TooltipDismissalMin, gsettings.TooltipDismissalMax, false, false)
	content.AddChild(widget.WrapWithSpan(2, d.tooltipDismissalField, widget.NewFieldTrailingLabel(i18n.Text("seconds"))))
}

func (d *generalSettingsDockable) createScrollWheelMultiplierField(content *unison.Panel) {
	title := i18n.Text("Scroll Wheel Multiplier")
	content.AddChild(widget.NewFieldLeadingLabel(title))
	d.scrollWheelMultiplierField = widget.NewDecimalField(nil, "", title,
		func() fxp.Int { return settings.Global().General.ScrollWheelMultiplier },
		func(v fxp.Int) { settings.Global().General.ScrollWheelMultiplier = v },
		gsettings.ScrollWheelMultiplierMin, gsettings.ScrollWheelMultiplierMax, false, false)
	d.scrollWheelMultiplierField.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	content.AddChild(d.scrollWheelMultiplierField)
}

func (d *generalSettingsDockable) createPathInfoField(content *unison.Panel, title, value string) {
	content.AddChild(widget.NewFieldLeadingLabel(title))
	content.AddChild(widget.NewNonEditableField(func(field *widget.NonEditableField) {
		field.Text = value
	}))
	addButton := unison.NewSVGButton(res.CopySVG)
	addButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Copy to clipboard"))
	addButton.ClickCallback = func() {
		unison.GlobalClipboard.SetText(value)
	}
	content.AddChild(addButton)
}

func (d *generalSettingsDockable) createExternalPDFCmdLineField(content *unison.Panel) {
	title := i18n.Text("External PDF Viewer")
	content.AddChild(widget.NewFieldLeadingLabel(title))
	d.externalPDFCmdlineField = widget.NewStringField(nil, "", title,
		func() string { return settings.Global().General.ExternalPDFCmdLine },
		func(s string) { settings.Global().General.ExternalPDFCmdLine = strings.TrimSpace(s) })
	d.externalPDFCmdlineField.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	d.externalPDFCmdlineField.ValidateCallback = func() bool {
		_, err := cmdline.Parse(strings.TrimSpace(d.externalPDFCmdlineField.Text()))
		return err == nil
	}
	d.externalPDFCmdlineField.Tooltip = unison.NewTooltipWithText(i18n.Text(`The internal PDF viewer will be used if the External PDF Viewer field is empty.
Use $FILE where the full path to the PDF should be placed.
Use $PAGE where the page number should be placed.`))
	content.AddChild(d.externalPDFCmdlineField)
}

func (d *generalSettingsDockable) reset() {
	*settings.Global().General = *gsettings.NewGeneral()
	d.sync()
}

func (d *generalSettingsDockable) sync() {
	s := settings.Global().General
	d.nameField.SetText(s.DefaultPlayerName)
	widget.SetCheckBoxState(d.autoFillProfileCheckbox, s.AutoFillProfile)
	widget.SetCheckBoxState(d.autoAddNaturalAttacksCheckbox, s.AutoAddNaturalAttacks)
	d.pointsField.SetText(s.InitialPoints.String())
	d.techLevelField.SetText(s.DefaultTechLevel)
	d.calendarPopup.Select(s.CalendarRef(settings.Global().Libraries()).Name)
	widget.SetFieldValue(d.initialListScaleField.Field, d.initialListScaleField.Format(s.InitialListUIScale))
	widget.SetFieldValue(d.initialEditorScaleField.Field, d.initialEditorScaleField.Format(s.InitialEditorUIScale))
	widget.SetFieldValue(d.initialSheetScaleField.Field, d.initialSheetScaleField.Format(s.InitialSheetUIScale))
	d.maxAutoColWidthField.SetText(strconv.Itoa(s.MaximumAutoColWidth))
	d.exportResolutionField.SetText(strconv.Itoa(s.ImageResolution))
	d.tooltipDelayField.SetText(s.TooltipDelay.String())
	d.tooltipDismissalField.SetText(s.TooltipDismissal.String())
	d.scrollWheelMultiplierField.SetText(s.ScrollWheelMultiplier.String())
	d.MarkForRedraw()
}

func (d *generalSettingsDockable) load(fileSystem fs.FS, filePath string) error {
	s, err := gsettings.NewGeneralFromFile(fileSystem, filePath)
	if err != nil {
		return err
	}
	*settings.Global().General = *s
	d.sync()
	return nil
}

func (d *generalSettingsDockable) save(filePath string) error {
	return settings.Global().General.Save(filePath)
}
