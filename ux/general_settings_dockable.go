/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
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
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/log/jotrotate"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xio/fs/paths"
	"github.com/richardwilkes/unison"
)

var languageSetting string

type generalSettingsDockable struct {
	SettingsDockable
	nameField                     *StringField
	autoFillProfileCheckbox       *CheckBox
	autoAddNaturalAttacksCheckbox *CheckBox
	groupContainersOnSortCheckbox *CheckBox
	pointsField                   *DecimalField
	techLevelField                *StringField
	calendarPopup                 *unison.PopupMenu[string]
	initialListScaleField         *PercentageField
	initialEditorScaleField       *PercentageField
	initialSheetScaleField        *PercentageField
	maxAutoColWidthField          *IntegerField
	monitorResolutionField        *IntegerField
	exportResolutionField         *IntegerField
	tooltipDelayField             *DecimalField
	tooltipDismissalField         *DecimalField
	scrollWheelMultiplierField    *DecimalField
	externalPDFCmdlineField       *StringField
	localeField                   *StringField
}

// ShowGeneralSettings the General Settings window.
func ShowGeneralSettings() {
	if Activate(func(d unison.Dockable) bool {
		_, ok := d.AsPanel().Self.(*generalSettingsDockable)
		return ok
	}) {
		return
	}
	d := &generalSettingsDockable{}
	d.Self = d
	d.TabTitle = i18n.Text("General Settings")
	d.TabIcon = svg.Settings
	d.Extensions = []string{gurps.GeneralSettingsExt}
	d.Loader = d.load
	d.Saver = d.save
	d.Resetter = d.reset
	d.WillCloseCallback = d.willClose
	d.Setup(d.addToStartToolbar, nil, d.initContent)
	d.nameField.RequestFocus()
}

func (d *generalSettingsDockable) addToStartToolbar(toolbar *unison.Panel) {
	helpButton := unison.NewSVGButton(svg.Help)
	helpButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Help"))
	helpButton.ClickCallback = func() { HandleLink(nil, "md:Help/Interface/General Settings") }
	toolbar.AddChild(helpButton)
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
	content.AddChild(NewFieldLeadingLabel(initialListScaleTitle))
	d.initialListScaleField = NewPercentageField(nil, "", initialListScaleTitle,
		func() int { return gurps.GlobalSettings().General.InitialListUIScale },
		func(v int) { gurps.GlobalSettings().General.InitialListUIScale = v },
		gurps.InitialUIScaleMin, gurps.InitialUIScaleMax, false, false)
	content.AddChild(WrapWithSpan(2, d.initialListScaleField))
	initialEditorScaleTitle := i18n.Text("Initial Editor Scale")
	content.AddChild(NewFieldLeadingLabel(initialEditorScaleTitle))
	d.initialEditorScaleField = NewPercentageField(nil, "", initialEditorScaleTitle,
		func() int { return gurps.GlobalSettings().General.InitialEditorUIScale },
		func(v int) { gurps.GlobalSettings().General.InitialEditorUIScale = v },
		gurps.InitialUIScaleMin, gurps.InitialUIScaleMax, false, false)
	content.AddChild(WrapWithSpan(2, d.initialEditorScaleField))
	initialSheetScaleTitle := i18n.Text("Initial Sheet Scale")
	content.AddChild(NewFieldLeadingLabel(initialSheetScaleTitle))
	d.initialSheetScaleField = NewPercentageField(nil, "", initialSheetScaleTitle,
		func() int { return gurps.GlobalSettings().General.InitialSheetUIScale },
		func(v int) { gurps.GlobalSettings().General.InitialSheetUIScale = v },
		gurps.InitialUIScaleMin, gurps.InitialUIScaleMax, false, false)
	content.AddChild(WrapWithSpan(2, d.initialSheetScaleField))
	d.createCellAutoMaxWidthField(content)
	d.createMonitorResolutionField(content)
	d.createImageResolutionField(content)
	d.createTooltipDelayField(content)
	d.createTooltipDismissalField(content)
	d.createScrollWheelMultiplierField(content)
	d.createPathInfoField(content, i18n.Text("Settings Path"), gurps.SettingsPath)
	d.createPathInfoField(content, i18n.Text("Translations Path"), i18n.Dir)
	d.createPathInfoField(content, i18n.Text("Log Path"), jotrotate.PathToLog)
	d.createExternalPDFCmdLineField(content)
	d.createLocaleField(content)
}

func (d *generalSettingsDockable) createPlayerAndDescFields(content *unison.Panel) {
	title := i18n.Text("Default Player Name")
	content.AddChild(NewFieldLeadingLabel(title))
	d.nameField = NewStringField(nil, "", title,
		func() string { return gurps.GlobalSettings().General.DefaultPlayerName },
		func(s string) { gurps.GlobalSettings().General.DefaultPlayerName = s })
	d.nameField.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	content.AddChild(d.nameField)
}

func (d *generalSettingsDockable) createCheckboxBlock(content *unison.Panel) {
	d.autoFillProfileCheckbox = NewCheckBox(nil, "", i18n.Text("Fill in initial description"),
		func() unison.CheckState {
			return unison.CheckStateFromBool(gurps.GlobalSettings().General.AutoFillProfile)
		},
		func(state unison.CheckState) {
			gurps.GlobalSettings().General.AutoFillProfile = state == unison.OnCheckState
		})
	d.autoFillProfileCheckbox.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	content.AddChild(NewFieldLeadingLabel(""))
	content.AddChild(d.autoFillProfileCheckbox)

	d.groupContainersOnSortCheckbox = NewCheckBox(nil, "", i18n.Text("Group containers when sorting"),
		func() unison.CheckState {
			return unison.CheckStateFromBool(gurps.GlobalSettings().General.GroupContainersOnSort)
		},
		func(state unison.CheckState) {
			gurps.GlobalSettings().General.GroupContainersOnSort = state == unison.OnCheckState
		})
	d.groupContainersOnSortCheckbox.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	content.AddChild(NewFieldLeadingLabel(""))
	content.AddChild(d.groupContainersOnSortCheckbox)

	d.autoAddNaturalAttacksCheckbox = NewCheckBox(nil, "", i18n.Text("Add natural attacks to new sheets"),
		func() unison.CheckState {
			return unison.CheckStateFromBool(gurps.GlobalSettings().General.AutoAddNaturalAttacks)
		},
		func(state unison.CheckState) {
			gurps.GlobalSettings().General.AutoAddNaturalAttacks = state == unison.OnCheckState
		})
	d.autoAddNaturalAttacksCheckbox.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	content.AddChild(NewFieldLeadingLabel(""))
	content.AddChild(d.autoAddNaturalAttacksCheckbox)
}

func (d *generalSettingsDockable) createInitialPointsFields(content *unison.Panel) {
	title := i18n.Text("Initial Points")
	content.AddChild(NewFieldLeadingLabel(title))
	d.pointsField = NewDecimalField(nil, "", title,
		func() fxp.Int { return gurps.GlobalSettings().General.InitialPoints },
		func(v fxp.Int) { gurps.GlobalSettings().General.InitialPoints = v }, gurps.InitialPointsMin,
		gurps.InitialPointsMax, false, false)
	d.pointsField.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	content.AddChild(d.pointsField)
}

func (d *generalSettingsDockable) createTechLevelField(content *unison.Panel) {
	title := i18n.Text("Default Tech Level")
	content.AddChild(NewFieldLeadingLabel(title))
	d.techLevelField = NewStringField(nil, "", title,
		func() string { return gurps.GlobalSettings().General.DefaultTechLevel },
		func(s string) { gurps.GlobalSettings().General.DefaultTechLevel = s })
	d.techLevelField.Tooltip = unison.NewTooltipWithText(techLevelInfo())
	d.techLevelField.SetMinimumTextWidthUsing("12^")
	d.techLevelField.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	content.AddChild(d.techLevelField)
}

func (d *generalSettingsDockable) createCalendarPopup(content *unison.Panel) {
	content.AddChild(NewFieldLeadingLabel(i18n.Text("Calendar")))
	d.calendarPopup = unison.NewPopupMenu[string]()
	libraries := gurps.GlobalSettings().Libraries()
	for _, lib := range gurps.AvailableCalendarRefs(libraries) {
		d.calendarPopup.AddDisabledItem(lib.Name)
		for _, one := range lib.List {
			d.calendarPopup.AddItem(one.Name)
		}
	}
	d.calendarPopup.Select(gurps.GlobalSettings().General.CalendarRef(libraries).Name)
	d.calendarPopup.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	d.calendarPopup.SelectionChangedCallback = func(p *unison.PopupMenu[string]) {
		if item, ok := p.Selected(); ok {
			gurps.GlobalSettings().General.CalendarName = item
		}
	}
	content.AddChild(d.calendarPopup)
}

func (d *generalSettingsDockable) createCellAutoMaxWidthField(content *unison.Panel) {
	title := i18n.Text("Max Auto Column Width")
	content.AddChild(NewFieldLeadingLabel(title))
	d.maxAutoColWidthField = NewIntegerField(nil, "", title,
		func() int { return gurps.GlobalSettings().General.MaximumAutoColWidth },
		func(v int) { gurps.GlobalSettings().General.MaximumAutoColWidth = v },
		gurps.AutoColWidthMin, gurps.AutoColWidthMax, false, false)
	d.maxAutoColWidthField.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	content.AddChild(d.maxAutoColWidthField)
}

func (d *generalSettingsDockable) createMonitorResolutionField(content *unison.Panel) {
	title := i18n.Text("Monitor Resolution")
	content.AddChild(NewFieldLeadingLabel(title))
	d.monitorResolutionField = NewNumericFieldWithException[int](nil, "", title,
		func(min, max int) []int { return []int{min, max} },
		func() int { return gurps.GlobalSettings().General.MonitorResolution },
		func(v int) { gurps.GlobalSettings().General.MonitorResolution = v },
		strconv.Itoa, strconv.Atoi, gurps.MonitorResolutionMin, gurps.MonitorResolutionMax, 0)
	content.AddChild(WrapWithSpan(2, d.monitorResolutionField,
		NewFieldTrailingLabel(i18n.Text("ppi (A value of 0 will cause the ppi reported by your monitor to be used)"))))
}

func (d *generalSettingsDockable) createImageResolutionField(content *unison.Panel) {
	title := i18n.Text("Image Export Resolution")
	content.AddChild(NewFieldLeadingLabel(title))
	d.exportResolutionField = NewIntegerField(nil, "", title,
		func() int { return gurps.GlobalSettings().General.ImageResolution },
		func(v int) { gurps.GlobalSettings().General.ImageResolution = v },
		gurps.ImageResolutionMin, gurps.ImageResolutionMax, false, false)
	content.AddChild(WrapWithSpan(2, d.exportResolutionField, NewFieldTrailingLabel(i18n.Text("ppi"))))
}

func (d *generalSettingsDockable) createTooltipDelayField(content *unison.Panel) {
	title := i18n.Text("Tooltip Delay")
	content.AddChild(NewFieldLeadingLabel(title))
	d.tooltipDelayField = NewDecimalField(nil, "", title,
		func() fxp.Int { return gurps.GlobalSettings().General.TooltipDelay },
		func(v fxp.Int) {
			general := gurps.GlobalSettings().General
			general.TooltipDelay = v
			general.UpdateToolTipTiming()
		}, gurps.TooltipDelayMin, gurps.TooltipDelayMax, false, false)
	content.AddChild(WrapWithSpan(2, d.tooltipDelayField, NewFieldTrailingLabel(i18n.Text("seconds"))))
}

func (d *generalSettingsDockable) createTooltipDismissalField(content *unison.Panel) {
	title := i18n.Text("Tooltip Dismissal")
	content.AddChild(NewFieldLeadingLabel(title))
	d.tooltipDismissalField = NewDecimalField(nil, "", title,
		func() fxp.Int { return gurps.GlobalSettings().General.TooltipDismissal },
		func(v fxp.Int) {
			general := gurps.GlobalSettings().General
			general.TooltipDismissal = v
			general.UpdateToolTipTiming()
		}, gurps.TooltipDismissalMin, gurps.TooltipDismissalMax, false, false)
	content.AddChild(WrapWithSpan(2, d.tooltipDismissalField, NewFieldTrailingLabel(i18n.Text("seconds"))))
}

func (d *generalSettingsDockable) createScrollWheelMultiplierField(content *unison.Panel) {
	title := i18n.Text("Scroll Wheel Multiplier")
	content.AddChild(NewFieldLeadingLabel(title))
	d.scrollWheelMultiplierField = NewDecimalField(nil, "", title,
		func() fxp.Int { return gurps.GlobalSettings().General.ScrollWheelMultiplier },
		func(v fxp.Int) { gurps.GlobalSettings().General.ScrollWheelMultiplier = v },
		gurps.ScrollWheelMultiplierMin, gurps.ScrollWheelMultiplierMax, false, false)
	d.scrollWheelMultiplierField.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	content.AddChild(d.scrollWheelMultiplierField)
}

func (d *generalSettingsDockable) createPathInfoField(content *unison.Panel, title, value string) {
	content.AddChild(NewFieldLeadingLabel(title))
	content.AddChild(NewNonEditableField(func(field *NonEditableField) {
		field.Text = value
	}))
	addButton := unison.NewSVGButton(svg.Copy)
	addButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Copy to clipboard"))
	addButton.ClickCallback = func() {
		unison.GlobalClipboard.SetText(value)
	}
	content.AddChild(addButton)
}

func (d *generalSettingsDockable) createExternalPDFCmdLineField(content *unison.Panel) {
	title := i18n.Text("External PDF Viewer")
	content.AddChild(NewFieldLeadingLabel(title))
	d.externalPDFCmdlineField = NewStringField(nil, "", title,
		func() string { return gurps.GlobalSettings().General.ExternalPDFCmdLine },
		func(s string) { gurps.GlobalSettings().General.ExternalPDFCmdLine = strings.TrimSpace(s) })
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
Use $PAGE where the page number should be placed.

In most cases, you'll want to surround the $FILE variable with quotes.`))
	content.AddChild(d.externalPDFCmdlineField)
}

func (d *generalSettingsDockable) createLocaleField(content *unison.Panel) {
	title := i18n.Text("Interface Locale")
	content.AddChild(NewFieldLeadingLabel(title))
	d.localeField = NewStringField(nil, "", title,
		func() string { return languageSetting },
		func(s string) { languageSetting = strings.TrimSpace(s) })
	d.localeField.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	d.localeField.Tooltip = unison.NewTooltipWithText(txt.Wrap("", i18n.Text(`The locale to use when presenting text in the user interface. This does not affect the content of data files. Leave this value blank to use the system default. Note that changes to this generally require quitting and restarting GCS to have the desired effect.`), 100))
	d.localeField.Watermark = i18n.Locale()
	content.AddChild(d.localeField)
}

func (d *generalSettingsDockable) reset() {
	*gurps.GlobalSettings().General = *gurps.NewGeneralSettings()
	languageSetting = ""
	d.sync()
}

func (d *generalSettingsDockable) sync() {
	s := gurps.GlobalSettings()
	gs := s.General
	d.nameField.SetText(gs.DefaultPlayerName)
	SetCheckBoxState(d.autoFillProfileCheckbox, gs.AutoFillProfile)
	SetCheckBoxState(d.groupContainersOnSortCheckbox, gs.GroupContainersOnSort)
	SetCheckBoxState(d.autoAddNaturalAttacksCheckbox, gs.AutoAddNaturalAttacks)
	d.pointsField.SetText(gs.InitialPoints.String())
	d.techLevelField.SetText(gs.DefaultTechLevel)
	d.calendarPopup.Select(gs.CalendarRef(s.Libraries()).Name)
	SetFieldValue(d.initialListScaleField.Field, d.initialListScaleField.Format(gs.InitialListUIScale))
	SetFieldValue(d.initialEditorScaleField.Field, d.initialEditorScaleField.Format(gs.InitialEditorUIScale))
	SetFieldValue(d.initialSheetScaleField.Field, d.initialSheetScaleField.Format(gs.InitialSheetUIScale))
	d.maxAutoColWidthField.SetText(strconv.Itoa(gs.MaximumAutoColWidth))
	d.monitorResolutionField.SetText(strconv.Itoa(gs.MonitorResolution))
	d.exportResolutionField.SetText(strconv.Itoa(gs.ImageResolution))
	d.tooltipDelayField.SetText(gs.TooltipDelay.String())
	d.tooltipDismissalField.SetText(gs.TooltipDismissal.String())
	d.scrollWheelMultiplierField.SetText(gs.ScrollWheelMultiplier.String())
	SetFieldValue(d.externalPDFCmdlineField.Field, gs.ExternalPDFCmdLine)
	SetFieldValue(d.localeField.Field, languageSetting)
	d.MarkForRedraw()
}

func (d *generalSettingsDockable) load(fileSystem fs.FS, filePath string) error {
	s, err := gurps.NewGeneralSettingsFromFile(fileSystem, filePath)
	if err != nil {
		return err
	}
	*gurps.GlobalSettings().General = *s
	d.sync()
	return nil
}

func (d *generalSettingsDockable) save(filePath string) error {
	return gurps.GlobalSettings().General.Save(filePath)
}

func (d *generalSettingsDockable) willClose() bool {
	if languageSetting == "" {
		i18n.Language = i18n.Locale()
		if err := os.Remove(languageSettingPath()); err != nil && !errors.Is(err, os.ErrNotExist) {
			jot.Error(errs.Wrap(err))
		}
	} else {
		i18n.Language = languageSetting
		if err := os.WriteFile(languageSettingPath(), []byte(languageSetting), 0o640); err != nil {
			jot.Error(errs.Wrap(err))
		}
	}
	return true
}

// LoadLanguageSetting loads the language setting from disk, if present, and applies it.
func LoadLanguageSetting() {
	if data, err := os.ReadFile(languageSettingPath()); err == nil {
		if s := strings.TrimSpace(strings.SplitN(strings.TrimSpace(string(data)), "\n", 2)[0]); len(s) > 1 && len(s) < 20 {
			i18n.Language = s
			languageSetting = s
		}
	}
}

func languageSettingPath() string {
	return filepath.Join(paths.AppDataDir(), cmdline.AppCmdName+"_language.txt")
}
