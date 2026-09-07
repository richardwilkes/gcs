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
	"fmt"
	"io/fs"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/progression"
	"github.com/richardwilkes/gcs/v5/model/paper"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/weight"
)

var _ GroupedCloser = &sheetSettingsDockable{}

// EntityPanel defines methods for a panel that can hold an entity.
type EntityPanel interface {
	unison.Paneler
	Entity() *gurps.Entity
}

type sheetSettingsDockable struct {
	SettingsDockable
	owner                     EntityPanel
	damageProgressionPopup    *unison.PopupMenu[progression.Option]
	options                   []sheetOptionCheckBox
	lengthUnitsPopup          *unison.PopupMenu[fxp.LengthUnit]
	weightUnitsPopup          *unison.PopupMenu[fxp.WeightUnit]
	numberFormats             []sheetNumberFormatRow
	userDescDisplayPopup      *unison.PopupMenu[display.Option]
	modifiersDisplayPopup     *unison.PopupMenu[display.Option]
	notesDisplayPopup         *unison.PopupMenu[display.Option]
	skillLevelAdjDisplayPopup *unison.PopupMenu[display.Option]
	orientationPopup          *unison.PopupMenu[paper.Orientation]
	paperSizeField            *unison.Field
	topMarginField            *unison.Field
	leftMarginField           *unison.Field
	bottomMarginField         *unison.Field
	rightMarginField          *unison.Field
}

// sheetOption describes one of the boolean sheet settings the dockable presents as a checkbox. The setting is reached
// through an accessor rather than a captured pointer, since the settings being edited are replaced wholesale by reset
// and load.
type sheetOption struct {
	title    string
	pageRef  string // a page reference to link to after the title, if any
	field    func(s *gurps.SheetSettings) *bool
	inverted bool // the setting hides what the checkbox offers to show, so the box is checked while the setting is off
	fullSync bool // the setting can change the columns a sheet shows, which only a full rebuild can pick up
}

// checked returns whether the checkbox for the option should be checked, given the settings.
func (o *sheetOption) checked(s *gurps.SheetSettings) bool {
	return *o.field(s) != o.inverted
}

// apply stores the state of the checkbox for the option into the settings.
func (o *sheetOption) apply(s *gurps.SheetSettings, checked bool) {
	*o.field(s) = checked != o.inverted
}

// sheetOptionCheckBox pairs a checkbox with the option it presents.
type sheetOptionCheckBox struct {
	box    *unison.CheckBox
	option sheetOption
}

// sheetNumberFormatRow holds the widgets that present one of the sheet's number formats, along with the accessor for
// the format they present.
type sheetNumberFormatRow struct {
	popup  *unison.PopupMenu[fxp.DecimalPlace]
	pad    *unison.CheckBox
	format func(s *gurps.SheetSettings) *fxp.NumberFormat
}

// sheetOptions returns the boolean sheet settings the dockable presents as checkboxes, in the order they are shown.
func sheetOptions() []sheetOption {
	return []sheetOption{
		{
			title:    i18n.Text("Show library source column"),
			field:    func(s *gurps.SheetSettings) *bool { return &s.HideSourceMismatch },
			inverted: true,
			fullSync: true,
		},
		{
			title:    i18n.Text("Show page reference column"),
			field:    func(s *gurps.SheetSettings) *bool { return &s.HidePageRefColumn },
			inverted: true,
			fullSync: true,
		},
		{
			title:    i18n.Text("Show tech level (TL) column"),
			field:    func(s *gurps.SheetSettings) *bool { return &s.HideTLColumn },
			inverted: true,
			fullSync: true,
		},
		{
			title:    i18n.Text("Show legality class (LC) column"),
			field:    func(s *gurps.SheetSettings) *bool { return &s.HideLCColumn },
			inverted: true,
			fullSync: true,
		},
		{
			title: i18n.Text("Show trait modifier cost adjustments"),
			field: func(s *gurps.SheetSettings) *bool { return &s.ShowTraitModifierAdj },
		},
		{
			title: i18n.Text("Show equipment modifier cost & weight adjustments"),
			field: func(s *gurps.SheetSettings) *bool { return &s.ShowEquipmentModifierAdj },
		},
		{
			title:    i18n.Text("Show all weapons"),
			field:    func(s *gurps.SheetSettings) *bool { return &s.ShowAllWeapons },
			fullSync: true,
		},
		{
			title:    i18n.Text("Hide unused columns in the melee & ranged weapon tables"),
			field:    func(s *gurps.SheetSettings) *bool { return &s.HideUnusedWeaponColumns },
			fullSync: true,
		},
		{
			title: i18n.Text("Show spell ritual, cost & time adjustments"),
			field: func(s *gurps.SheetSettings) *bool { return &s.ShowSpellAdj },
		},
		{
			title: i18n.Text("Show the title instead of the name in the footer"),
			field: func(s *gurps.SheetSettings) *bool { return &s.UseTitleInFooter },
		},
		{
			title:   i18n.Text("Use Multiplicative Modifiers"),
			pageRef: "P102",
			field:   func(s *gurps.SheetSettings) *bool { return &s.UseMultiplicativeModifiers },
		},
		{
			title:   i18n.Text("Use Half-Stat Defaults"),
			pageRef: "PY65:30",
			field:   func(s *gurps.SheetSettings) *bool { return &s.UseHalfStatDefaults },
		},
		{
			title:   i18n.Text("Use Modifying Dice + Adds"),
			pageRef: "B269",
			field:   func(s *gurps.SheetSettings) *bool { return &s.UseModifyingDicePlusAdds },
		},
		{
			title: i18n.Text("Exclude unspent points from total"),
			field: func(s *gurps.SheetSettings) *bool { return &s.ExcludeUnspentPointsFromTotal },
		},
		{
			title: i18n.Text("Show Lifting ST-based damage"),
			field: func(s *gurps.SheetSettings) *bool { return &s.ShowLiftingSTDamage },
		},
		{
			title:   i18n.Text("Show IQ-based damage"),
			pageRef: "PY120:7",
			field:   func(s *gurps.SheetSettings) *bool { return &s.ShowIQBasedDamage },
		},
		{
			title:    i18n.Text("Hide conditional modifiers whose total is zero"),
			field:    func(s *gurps.SheetSettings) *bool { return &s.HideZeroValueConditionalMods },
			fullSync: true,
		},
	}
}

// ShowSheetSettings the Sheet Settings. Pass in nil to edit the defaults or a sheet to edit the sheet's.
func ShowSheetSettings(owner EntityPanel) {
	if Activate(func(d unison.Dockable) bool {
		if s, ok := d.AsPanel().Self.(*sheetSettingsDockable); ok && owner == s.owner {
			return true
		}
		return false
	}) {
		return
	}
	d := &sheetSettingsDockable{owner: owner}
	d.Self = d
	d.TabTitle = sheetSettingsTabTitle(owner)
	d.TabIcon = svg.Settings
	d.Extensions = []string{gurps.SheetSettingsExt}
	d.Loader = d.load
	d.Saver = d.save
	d.Resetter = d.reset
	d.Setup(d.addToStartToolbar, nil, d.initContent)
}

// sheetSettingsTabTitle returns the tab title to use for the sheet settings of the given owner, or for the defaults
// when the owner is nil. The character name must be substituted into an otherwise constant string, since i18n.Text
// looks the whole string up in the translation catalog; passing it a string built with the name embedded produces a key
// that can never match an entry and can't be extracted for translation in the first place.
func sheetSettingsTabTitle(owner EntityPanel) string {
	if owner == nil {
		return i18n.Text("Default Sheet Settings")
	}
	return fmt.Sprintf(i18n.Text("Sheet Settings: %s"), owner.Entity().Profile.Name)
}

func (d *sheetSettingsDockable) addToStartToolbar(toolbar *unison.Panel) {
	addHelpButton(toolbar, "md:User%20Guide/Sheet%20Settings")
}

func (d *sheetSettingsDockable) CloseWithGroup(other unison.Paneler) bool {
	return d.owner != nil && d.owner == other
}

func (d *sheetSettingsDockable) settings() *gurps.SheetSettings {
	if d.owner != nil {
		return d.owner.Entity().SheetSettings
	}
	return gurps.GlobalSettings().Sheet
}

func (d *sheetSettingsDockable) initContent(content *unison.Panel) {
	content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.DefaultLabelTheme.Font.LineHeight(),
	})
	d.createDamageProgression(content)
	d.createOptions(content)
	d.createUnitsOfMeasurement(content)
	d.createDecimalPlaces(content)
	d.createWhereToDisplay(content)
	d.createPageSettings(content)
}

func (d *sheetSettingsDockable) createDamageProgression(content *unison.Panel) {
	s := d.settings()
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	desc := unison.NewMarkdown(true)
	desc.SetContent(s.DamageProgression.AltString(), -1)
	d.damageProgressionPopup = d.createSettingPopup(panel, i18n.Text("Damage Progression"),
		progression.Options, s.DamageProgression,
		func(item progression.Option) {
			d.settings().DamageProgression = item
			desc.SetContent(item.AltString(), -1)
			desc.MarkForLayoutRecursivelyUpward()
			desc.MarkForRedraw()
		})
	d.damageProgressionPopup.Tooltip = newWrappedTooltip(i18n.Text("Determines the method used to calculate thrust and swing damage"))
	panel.AddChild(unison.NewPanel())
	panel.AddChild(desc)
	content.AddChild(panel)
}

func (d *sheetSettingsDockable) createOptions(content *unison.Panel) {
	s := d.settings()
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	options := sheetOptions()
	d.options = make([]sheetOptionCheckBox, 0, len(options))
	for _, option := range options {
		box := d.addCheckBox(panel, option.title, option.pageRef, option.checked(s), func(checked bool) {
			option.apply(d.settings(), checked)
			d.syncSheet(option.fullSync)
		})
		d.options = append(d.options, sheetOptionCheckBox{box: box, option: option})
	}
	content.AddChild(panel)
}

// addCheckBox adds a checkbox with the given title to the panel, followed by a link to the page reference when one is
// given, and returns it. The click callback is handed the state the box was left in.
func (d *sheetSettingsDockable) addCheckBox(panel *unison.Panel, title, pageRef string, checked bool, onClick func(checked bool)) *unison.CheckBox {
	checkbox := unison.NewCheckBox()
	checkbox.SetTitle(title)
	checkbox.State = check.FromBool(checked)
	checkbox.ClickCallback = func() { onClick(checkbox.State == check.On) }
	if pageRef == "" {
		panel.AddChild(checkbox)
		return checkbox
	}
	wrapper := unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{Columns: 4})
	wrapper.AddChild(checkbox)
	label := unison.NewLabel()
	label.Font = checkbox.Font
	label.SetTitle(" (")
	wrapper.AddChild(label)
	wrapper.AddChild(unison.NewLink(pageRef, "", pageRef, &unison.DefaultLinkTheme, func(_ unison.Paneler, _ string) {
		OpenPageReference(pageRef, "", nil)
	}))
	label = unison.NewLabel()
	label.Font = checkbox.Font
	label.SetTitle(")")
	wrapper.AddChild(label)
	panel.AddChild(wrapper)
	return checkbox
}

func (d *sheetSettingsDockable) createUnitsOfMeasurement(content *unison.Panel) {
	s := d.settings()
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill})
	d.createHeader(panel, i18n.Text("Units of Measurement"), 2)
	d.lengthUnitsPopup = d.createSettingPopup(panel, i18n.Text("Length Units"), fxp.LengthUnits,
		s.DefaultLengthUnits, func(item fxp.LengthUnit) { d.settings().DefaultLengthUnits = item })
	d.weightUnitsPopup = d.createSettingPopup(panel, i18n.Text("Weight Units"), fxp.WeightUnits,
		s.DefaultWeightUnits, func(item fxp.WeightUnit) { d.settings().DefaultWeightUnits = item })
	content.AddChild(panel)
}

// createDecimalPlaces adds the section that controls how many decimal places the sheet rounds various numbers to for
// display. These affect only what is shown: the values themselves are always stored, and edited, at full precision.
func (d *sheetSettingsDockable) createDecimalPlaces(content *unison.Panel) {
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  3,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill})
	d.createHeader(panel, i18n.Text("Decimal Places"), 3)
	d.numberFormats = []sheetNumberFormatRow{
		d.addNumberFormatRow(panel, i18n.Text("Height"),
			i18n.Text(`How many decimal places the height in the Description block on the sheet is rounded to for display. "As Needed" shows every decimal place the value has.`),
			func(s *gurps.SheetSettings) *fxp.NumberFormat { return &s.HeightFormat }),
		d.addNumberFormatRow(panel, i18n.Text("Body Weight"),
			i18n.Text("How many decimal places the character's weight in the Description block on the sheet is rounded to for display"),
			func(s *gurps.SheetSettings) *fxp.NumberFormat { return &s.BodyWeightFormat }),
		d.addNumberFormatRow(panel, i18n.Text("Equipment Weight"),
			i18n.Text("How many decimal places the weight columns and the carried & other equipment totals on the sheet are rounded to for display"),
			func(s *gurps.SheetSettings) *fxp.NumberFormat { return &s.EquipmentWeightFormat }),
		d.addNumberFormatRow(panel, i18n.Text("Equipment Value"),
			i18n.Text("How many decimal places the value columns and the carried & other equipment totals on the sheet are rounded to for display"),
			func(s *gurps.SheetSettings) *fxp.NumberFormat { return &s.EquipmentValueFormat }),
	}
	content.AddChild(panel)
}

// addNumberFormatRow adds the popup that chooses how many decimal places one of the sheet's number formats rounds to,
// with the given tooltip, and the checkbox that has it pad with zeros out to that many, and returns them along with
// the format's accessor.
func (d *sheetSettingsDockable) addNumberFormatRow(panel *unison.Panel, title, tooltip string, format func(s *gurps.SheetSettings) *fxp.NumberFormat) sheetNumberFormatRow {
	row := sheetNumberFormatRow{format: format}
	current := format(d.settings())
	row.popup = d.createSettingPopup(panel, title, fxp.DecimalPlaces, current.Places,
		func(item fxp.DecimalPlace) { format(d.settings()).Places = item })
	row.popup.Tooltip = newWrappedTooltip(tooltip)
	row.pad = d.addCheckBox(panel, i18n.Text("Pad with zeros"), "", current.PadWithZeros, func(checked bool) {
		format(d.settings()).PadWithZeros = checked
		d.syncSheet(false)
	})
	row.pad.Tooltip = newWrappedTooltip(i18n.Text(`Show trailing zeros out to the number of decimal places chosen, e.g. "7.50" rather than "7.5" at 2 decimal places. Has no effect when "As Needed" or "0" is chosen.`))
	return row
}

func (d *sheetSettingsDockable) createWhereToDisplay(content *unison.Panel) {
	s := d.settings()
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill})
	d.createHeader(panel, i18n.Text("Where to display…"), 2)
	d.userDescDisplayPopup = d.createSettingPopup(panel, i18n.Text("User Description"), display.Options,
		s.UserDescriptionDisplay, func(option display.Option) { d.settings().UserDescriptionDisplay = option })
	d.modifiersDisplayPopup = d.createSettingPopup(panel, i18n.Text("Modifiers"), display.Options,
		s.ModifiersDisplay, func(option display.Option) { d.settings().ModifiersDisplay = option })
	d.notesDisplayPopup = d.createSettingPopup(panel, i18n.Text("Notes"), display.Options, s.NotesDisplay,
		func(option display.Option) { d.settings().NotesDisplay = option })
	d.skillLevelAdjDisplayPopup = d.createSettingPopup(panel, i18n.Text("Skill Level Adjustments"), display.Options,
		s.SkillLevelAdjDisplay, func(option display.Option) { d.settings().SkillLevelAdjDisplay = option })
	content.AddChild(panel)
}

func (d *sheetSettingsDockable) createPageSettings(content *unison.Panel) {
	s := d.settings()
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  4,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill})
	d.createHeader(panel, i18n.Text("Page Settings"), 4)
	d.paperSizeField = d.createPaperSizeField(panel, s.Page.Size, func(option string) { d.settings().Page.Size = option })
	d.orientationPopup = d.createSettingPopup(panel, i18n.Text("Orientation"), paper.Orientations,
		s.Page.Orientation, func(option paper.Orientation) { d.settings().Page.Orientation = option })
	d.topMarginField = d.createPaperMarginField(panel, i18n.Text("Top Margin"), s.Page.TopMargin,
		func(value paper.Length) { d.settings().Page.TopMargin = value })
	d.bottomMarginField = d.createPaperMarginField(panel, i18n.Text("Bottom Margin"), s.Page.BottomMargin,
		func(value paper.Length) { d.settings().Page.BottomMargin = value })
	d.leftMarginField = d.createPaperMarginField(panel, i18n.Text("Left Margin"), s.Page.LeftMargin,
		func(value paper.Length) { d.settings().Page.LeftMargin = value })
	d.rightMarginField = d.createPaperMarginField(panel, i18n.Text("Right Margin"), s.Page.RightMargin,
		func(value paper.Length) { d.settings().Page.RightMargin = value })
	content.AddChild(panel)
}

func (d *sheetSettingsDockable) createPaperSizeField(panel *unison.Panel, current string, set func(value string)) *unison.Field {
	panel.AddChild(NewFieldLeadingLabel(i18n.Text("Paper Size"), false))
	wrapper := unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
	})
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	panel.AddChild(wrapper)
	field := unison.NewField()
	field.SetText(current)
	field.ValidateCallback = func() bool {
		_, _, valid := gurps.ParsePageSize(field.Text())
		return valid
	}
	field.ModifiedCallback = func(_, after *unison.FieldState) {
		if width, height, valid := gurps.ParsePageSize(after.Text); valid {
			set(gurps.ToPageSize(width, height))
			d.syncSheet(false)
		}
	}
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	wrapper.AddChild(field)
	info := NewInfoPop()
	var buffer strings.Builder
	for _, one := range gurps.StdPaperSizes {
		if buffer.Len() > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteByte('"')
		buffer.WriteString(one.Name)
		buffer.WriteByte('"')
	}
	AddHelpToInfoPop(info, wrapTextForTooltip(fmt.Sprintf(i18n.Text(`Enter a standard paper size (e.g., one of %s) or a custom size (e.g., "8.5in x 11in", "210mm x 297mm")`), buffer.String())))
	wrapper.AddChild(info)
	return field
}

func (d *sheetSettingsDockable) createPaperMarginField(panel *unison.Panel, title string, current paper.Length, set func(value paper.Length)) *unison.Field {
	panel.AddChild(NewFieldLeadingLabel(title, false))
	field := unison.NewField()
	field.SetText(current.String())
	field.ValidateCallback = func() bool {
		_, err := paper.ParseLengthFromString(field.Text())
		return err == nil
	}
	field.ModifiedCallback = func(_, after *unison.FieldState) {
		if value, err := paper.ParseLengthFromString(after.Text); err == nil {
			set(value)
			d.syncSheet(false)
		}
	}
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	panel.AddChild(field)
	return field
}

func (d *sheetSettingsDockable) createSettingPopup[T comparable](panel *unison.Panel, title string, choices []T, current T, set func(option T)) *unison.PopupMenu[T] {
	panel.AddChild(NewFieldLeadingLabel(title, false))
	popup := newPopupMenu(choices, current, func(item T) {
		set(item)
		d.syncSheet(false)
	})
	panel.AddChild(popup)
	return popup
}

func (d *sheetSettingsDockable) createHeader(panel *unison.Panel, title string, hspan int) {
	label := unison.NewLabel()
	desc := label.Font.Descriptor()
	desc.Weight = weight.Bold
	label.Font = desc.Font()
	label.SetTitle(title)
	label.SetLayoutData(&unison.FlexLayoutData{HSpan: hspan})
	panel.AddChild(label)
	sep := unison.NewSeparator()
	sep.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  hspan,
		HAlign: align.Fill,
		HGrab:  true,
	})
	panel.AddChild(sep)
}

func (d *sheetSettingsDockable) reset() {
	if d.owner != nil {
		entity := d.owner.Entity()
		entity.SheetSettings = gurps.GlobalSettings().Sheet.Clone(entity)
	} else {
		gurps.GlobalSettings().Sheet = gurps.FactorySheetSettings()
	}
	d.sync()
	d.syncSheet(true)
}

func (d *sheetSettingsDockable) sync() {
	s := d.settings()
	d.damageProgressionPopup.Select(s.DamageProgression)
	for _, one := range d.options {
		one.box.State = check.FromBool(one.option.checked(s))
	}
	d.lengthUnitsPopup.Select(s.DefaultLengthUnits)
	d.weightUnitsPopup.Select(s.DefaultWeightUnits)
	for _, row := range d.numberFormats {
		format := row.format(s)
		row.popup.Select(format.Places)
		row.pad.State = check.FromBool(format.PadWithZeros)
	}
	d.userDescDisplayPopup.Select(s.UserDescriptionDisplay)
	d.modifiersDisplayPopup.Select(s.ModifiersDisplay)
	d.notesDisplayPopup.Select(s.NotesDisplay)
	d.skillLevelAdjDisplayPopup.Select(s.SkillLevelAdjDisplay)
	d.paperSizeField.SetText(s.Page.Size)
	d.orientationPopup.Select(s.Page.Orientation)
	d.topMarginField.SetText(s.Page.TopMargin.String())
	d.leftMarginField.SetText(s.Page.LeftMargin.String())
	d.bottomMarginField.SetText(s.Page.BottomMargin.String())
	d.rightMarginField.SetText(s.Page.RightMargin.String())
	d.MarkForRedraw()
}

func (d *sheetSettingsDockable) syncSheet(full bool) {
	var entity *gurps.Entity
	if d.owner != nil {
		entity = d.owner.Entity()
	} else {
		// The global sheet settings were just changed, so republish the snapshot background parses read.
		gurps.SyncGlobalSheetSettings()
	}
	for _, one := range AllDockables() {
		if s, ok := one.(gurps.SheetSettingsResponder); ok {
			s.SheetSettingsUpdated(entity, full)
		}
	}
}

func (d *sheetSettingsDockable) load(fileSystem fs.FS, filePath string) error {
	s, err := gurps.NewSheetSettingsFromFile(fileSystem, filePath)
	if err != nil {
		return err
	}
	if d.owner != nil {
		entity := d.owner.Entity()
		entity.SheetSettings = s
		s.SetOwningEntity(entity)
	} else {
		gurps.GlobalSettings().Sheet = s
	}
	d.sync()
	d.syncSheet(true)
	return nil
}

func (d *sheetSettingsDockable) save(filePath string) error {
	return d.settings().Save(filePath)
}
