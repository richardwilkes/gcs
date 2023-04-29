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
	"io/fs"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

var _ GroupedCloser = &sheetSettingsDockable{}

// EntityPanel defines methods for a panel that can hold an entity.
type EntityPanel interface {
	unison.Paneler
	Entity() *gurps.Entity
}

type sheetSettingsDockable struct {
	SettingsDockable
	owner                              EntityPanel
	damageProgressionPopup             *unison.PopupMenu[gurps.DamageProgression]
	showTraitModifier                  *unison.CheckBox
	showEquipmentModifier              *unison.CheckBox
	showSpellAdjustments               *unison.CheckBox
	showTitleInsteadOfNameInPageFooter *unison.CheckBox
	useMultiplicativeModifiers         *unison.CheckBox
	useModifyDicePlusAdds              *unison.CheckBox
	excludeUnspentPointsFromTotal      *unison.CheckBox
	useHalfStatDefaults                *unison.CheckBox
	lengthUnitsPopup                   *unison.PopupMenu[gurps.LengthUnits]
	weightUnitsPopup                   *unison.PopupMenu[gurps.WeightUnits]
	userDescDisplayPopup               *unison.PopupMenu[gurps.DisplayOption]
	modifiersDisplayPopup              *unison.PopupMenu[gurps.DisplayOption]
	notesDisplayPopup                  *unison.PopupMenu[gurps.DisplayOption]
	skillLevelAdjDisplayPopup          *unison.PopupMenu[gurps.DisplayOption]
	paperSizePopup                     *unison.PopupMenu[gurps.PaperSize]
	orientationPopup                   *unison.PopupMenu[gurps.PaperOrientation]
	topMarginField                     *unison.Field
	leftMarginField                    *unison.Field
	bottomMarginField                  *unison.Field
	rightMarginField                   *unison.Field
	blockLayoutField                   *unison.Field
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
	if owner != nil {
		d.TabTitle = i18n.Text("Sheet Settings: " + owner.Entity().Profile.Name)
	} else {
		d.TabTitle = i18n.Text("Default Sheet Settings")
	}
	d.TabIcon = svg.Settings
	d.Extensions = []string{gurps.SheetSettingsExt}
	d.Loader = d.load
	d.Saver = d.save
	d.Resetter = d.reset
	d.Setup(d.addToStartToolbar, nil, d.initContent)
}

func (d *sheetSettingsDockable) addToStartToolbar(toolbar *unison.Panel) {
	helpButton := unison.NewSVGButton(svg.Help)
	helpButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Help"))
	helpButton.ClickCallback = func() { HandleLink(nil, "md:Help/Interface/Sheet Settings") }
	toolbar.AddChild(helpButton)
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
	d.createWhereToDisplay(content)
	d.createPageSettings(content)
	d.createBlockLayout(content)
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
	d.damageProgressionPopup = createSettingPopup(d, panel, i18n.Text("Damage Progression"),
		gurps.AllDamageProgression, s.DamageProgression,
		func(item gurps.DamageProgression) {
			d.settings().DamageProgression = item
			desc.SetContent(item.AltString(), -1)
			desc.MarkForLayoutRecursivelyUpward()
			desc.MarkForRedraw()
		})
	d.damageProgressionPopup.Tooltip = unison.NewTooltipWithText(i18n.Text("Determines the method used to calculate thrust and swing damage"))
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
	d.showTraitModifier = d.addCheckBox(panel, i18n.Text("Show trait modifier cost adjustments"),
		s.ShowTraitModifierAdj, func() {
			d.settings().ShowTraitModifierAdj = d.showTraitModifier.State == unison.OnCheckState
			d.syncSheet(false)
		})
	d.showEquipmentModifier = d.addCheckBox(panel, i18n.Text("Show equipment modifier cost & weight adjustments"),
		s.ShowEquipmentModifierAdj, func() {
			d.settings().ShowEquipmentModifierAdj = d.showEquipmentModifier.State == unison.OnCheckState
			d.syncSheet(false)
		})
	d.showSpellAdjustments = d.addCheckBox(panel, i18n.Text("Show spell ritual, cost & time adjustments"),
		s.ShowSpellAdj, func() {
			d.settings().ShowSpellAdj = d.showSpellAdjustments.State == unison.OnCheckState
			d.syncSheet(false)
		})
	d.showTitleInsteadOfNameInPageFooter = d.addCheckBox(panel,
		i18n.Text("Show the title instead of the name in the footer"), s.UseTitleInFooter, func() {
			d.settings().UseTitleInFooter = d.showTitleInsteadOfNameInPageFooter.State == unison.OnCheckState
			d.syncSheet(false)
		})
	d.useMultiplicativeModifiers = d.addCheckBoxWithLink(panel,
		i18n.Text("Use Multiplicative Modifiers"), "P102", s.UseMultiplicativeModifiers, func() {
			d.settings().UseMultiplicativeModifiers = d.useMultiplicativeModifiers.State == unison.OnCheckState
			d.syncSheet(false)
		})
	d.useHalfStatDefaults = d.addCheckBoxWithLink(panel, i18n.Text("Use Half-Stat Defaults"), "PY65:30",
		s.UseHalfStatDefaults, func() {
			d.settings().UseHalfStatDefaults = d.useHalfStatDefaults.State == unison.OnCheckState
			d.syncSheet(false)
		})
	d.useModifyDicePlusAdds = d.addCheckBoxWithLink(panel, i18n.Text("Use Modifying Dice + Adds"), "B269",
		s.UseModifyingDicePlusAdds, func() {
			d.settings().UseModifyingDicePlusAdds = d.useModifyDicePlusAdds.State == unison.OnCheckState
			d.syncSheet(false)
		})
	d.excludeUnspentPointsFromTotal = d.addCheckBox(panel, i18n.Text("Exclude unspent points from total"),
		s.ExcludeUnspentPointsFromTotal, func() {
			d.settings().ExcludeUnspentPointsFromTotal = d.excludeUnspentPointsFromTotal.State == unison.OnCheckState
			d.syncSheet(false)
		})
	content.AddChild(panel)
}

func (d *sheetSettingsDockable) addCheckBox(panel *unison.Panel, title string, checked bool, onClick func()) *unison.CheckBox {
	checkbox := unison.NewCheckBox()
	checkbox.Text = title
	checkbox.State = unison.CheckStateFromBool(checked)
	checkbox.ClickCallback = onClick
	panel.AddChild(checkbox)
	return checkbox
}

func (d *sheetSettingsDockable) addCheckBoxWithLink(panel *unison.Panel, title, ref string, checked bool, onClick func()) *unison.CheckBox {
	wrapper := unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{Columns: 4})
	checkbox := unison.NewCheckBox()
	checkbox.Text = title
	checkbox.State = unison.CheckStateFromBool(checked)
	checkbox.ClickCallback = onClick
	wrapper.AddChild(checkbox)
	label := unison.NewLabel()
	label.Font = checkbox.Font
	label.Text = " ("
	wrapper.AddChild(label)
	wrapper.AddChild(unison.NewLink(ref, "", ref, unison.DefaultLinkTheme, func(_ unison.Paneler, _ string) {
		OpenPageReference(ref, "", nil)
	}))
	label = unison.NewLabel()
	label.Font = checkbox.Font
	label.Text = ")"
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
	panel.SetLayoutData(&unison.FlexLayoutData{HAlign: unison.FillAlignment})
	d.createHeader(panel, i18n.Text("Units of Measurement"), 2)
	d.lengthUnitsPopup = createSettingPopup(d, panel, i18n.Text("Length Units"), gurps.AllLengthUnits,
		s.DefaultLengthUnits, func(item gurps.LengthUnits) { d.settings().DefaultLengthUnits = item })
	d.weightUnitsPopup = createSettingPopup(d, panel, i18n.Text("Weight Units"), gurps.AllWeightUnits,
		s.DefaultWeightUnits, func(item gurps.WeightUnits) { d.settings().DefaultWeightUnits = item })
	content.AddChild(panel)
}

func (d *sheetSettingsDockable) createWhereToDisplay(content *unison.Panel) {
	s := d.settings()
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{HAlign: unison.FillAlignment})
	d.createHeader(panel, i18n.Text("Where to display…"), 2)
	d.userDescDisplayPopup = createSettingPopup(d, panel, i18n.Text("User Description"), gurps.AllDisplayOption,
		s.UserDescriptionDisplay, func(option gurps.DisplayOption) { d.settings().UserDescriptionDisplay = option })
	d.modifiersDisplayPopup = createSettingPopup(d, panel, i18n.Text("Modifiers"), gurps.AllDisplayOption,
		s.ModifiersDisplay, func(option gurps.DisplayOption) { d.settings().ModifiersDisplay = option })
	d.notesDisplayPopup = createSettingPopup(d, panel, i18n.Text("Notes"), gurps.AllDisplayOption, s.NotesDisplay,
		func(option gurps.DisplayOption) { d.settings().NotesDisplay = option })
	d.skillLevelAdjDisplayPopup = createSettingPopup(d, panel, i18n.Text("Skill Level Adjustments"), gurps.AllDisplayOption,
		s.SkillLevelAdjDisplay, func(option gurps.DisplayOption) { d.settings().SkillLevelAdjDisplay = option })
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
	panel.SetLayoutData(&unison.FlexLayoutData{HAlign: unison.FillAlignment})
	d.createHeader(panel, i18n.Text("Page Settings"), 4)
	d.paperSizePopup = createSettingPopup(d, panel, i18n.Text("Paper Size"), gurps.AllPaperSize,
		s.Page.Size, func(option gurps.PaperSize) { d.settings().Page.Size = option })
	d.orientationPopup = createSettingPopup(d, panel, i18n.Text("Orientation"), gurps.AllPaperOrientation,
		s.Page.Orientation, func(option gurps.PaperOrientation) { d.settings().Page.Orientation = option })
	d.topMarginField = d.createPaperMarginField(panel, i18n.Text("Top Margin"), s.Page.TopMargin,
		func(value gurps.PaperLength) { d.settings().Page.TopMargin = value })
	d.bottomMarginField = d.createPaperMarginField(panel, i18n.Text("Bottom Margin"), s.Page.BottomMargin,
		func(value gurps.PaperLength) { d.settings().Page.BottomMargin = value })
	d.leftMarginField = d.createPaperMarginField(panel, i18n.Text("Left Margin"), s.Page.LeftMargin,
		func(value gurps.PaperLength) { d.settings().Page.LeftMargin = value })
	d.rightMarginField = d.createPaperMarginField(panel, i18n.Text("Right Margin"), s.Page.RightMargin,
		func(value gurps.PaperLength) { d.settings().Page.RightMargin = value })
	content.AddChild(panel)
}

func (d *sheetSettingsDockable) createBlockLayout(content *unison.Panel) {
	s := d.settings()
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{HAlign: unison.FillAlignment})
	label := unison.NewLabel()
	label.Text = i18n.Text("Block Layout")
	desc := label.Font.Descriptor()
	desc.Weight = unison.BoldFontWeight
	label.Font = desc.Font()
	panel.AddChild(label)
	d.blockLayoutField = unison.NewMultiLineField()
	lastBlockLayout := s.BlockLayout.String()
	d.blockLayoutField.SetText(lastBlockLayout)
	d.blockLayoutField.ValidateCallback = func() bool {
		_, valid := gurps.NewBlockLayoutFromString(d.blockLayoutField.Text())
		return valid
	}
	d.blockLayoutField.ModifiedCallback = func(_, after *unison.FieldState) {
		if blockLayout, valid := gurps.NewBlockLayoutFromString(after.Text); valid {
			localSettings := d.settings()
			currentBlockLayout := blockLayout.String()
			if lastBlockLayout != currentBlockLayout {
				lastBlockLayout = currentBlockLayout
				localSettings.BlockLayout = blockLayout
				d.syncSheet(true)
			}
		}
	}
	d.blockLayoutField.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	panel.AddChild(d.blockLayoutField)
	content.AddChild(panel)
}

func (d *sheetSettingsDockable) createPaperMarginField(panel *unison.Panel, title string, current gurps.PaperLength, set func(value gurps.PaperLength)) *unison.Field {
	panel.AddChild(NewFieldLeadingLabel(title))
	field := unison.NewField()
	field.SetText(current.String())
	field.ValidateCallback = func() bool {
		_, err := gurps.ParsePaperLengthFromString(field.Text())
		return err == nil
	}
	field.ModifiedCallback = func(_, after *unison.FieldState) {
		if value, err := gurps.ParsePaperLengthFromString(after.Text); err == nil {
			set(value)
			d.syncSheet(false)
		}
	}
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	panel.AddChild(field)
	return field
}

func createSettingPopup[T comparable](d *sheetSettingsDockable, panel *unison.Panel, title string, choices []T, current T, set func(option T)) *unison.PopupMenu[T] {
	panel.AddChild(NewFieldLeadingLabel(title))
	popup := unison.NewPopupMenu[T]()
	for _, one := range choices {
		popup.AddItem(one)
	}
	popup.Select(current)
	popup.SelectionChangedCallback = func(p *unison.PopupMenu[T]) {
		if item, ok := p.Selected(); ok {
			set(item)
			d.syncSheet(false)
		}
	}
	panel.AddChild(popup)
	return popup
}

func (d *sheetSettingsDockable) createHeader(panel *unison.Panel, title string, hspan int) {
	label := unison.NewLabel()
	label.Text = title
	desc := label.Font.Descriptor()
	desc.Weight = unison.BoldFontWeight
	label.Font = desc.Font()
	label.SetLayoutData(&unison.FlexLayoutData{HSpan: hspan})
	panel.AddChild(label)
	sep := unison.NewSeparator()
	sep.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  hspan,
		HAlign: unison.FillAlignment,
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
}

func (d *sheetSettingsDockable) sync() {
	s := d.settings()
	d.damageProgressionPopup.Select(s.DamageProgression)
	d.showTraitModifier.State = unison.CheckStateFromBool(s.ShowTraitModifierAdj)
	d.showEquipmentModifier.State = unison.CheckStateFromBool(s.ShowEquipmentModifierAdj)
	d.showSpellAdjustments.State = unison.CheckStateFromBool(s.ShowSpellAdj)
	d.showTitleInsteadOfNameInPageFooter.State = unison.CheckStateFromBool(s.UseTitleInFooter)
	d.useMultiplicativeModifiers.State = unison.CheckStateFromBool(s.UseMultiplicativeModifiers)
	d.useHalfStatDefaults.State = unison.CheckStateFromBool(s.UseHalfStatDefaults)
	d.useModifyDicePlusAdds.State = unison.CheckStateFromBool(s.UseModifyingDicePlusAdds)
	d.excludeUnspentPointsFromTotal.State = unison.CheckStateFromBool(s.ExcludeUnspentPointsFromTotal)
	d.lengthUnitsPopup.Select(s.DefaultLengthUnits)
	d.weightUnitsPopup.Select(s.DefaultWeightUnits)
	d.userDescDisplayPopup.Select(s.UserDescriptionDisplay)
	d.modifiersDisplayPopup.Select(s.ModifiersDisplay)
	d.notesDisplayPopup.Select(s.NotesDisplay)
	d.skillLevelAdjDisplayPopup.Select(s.SkillLevelAdjDisplay)
	d.paperSizePopup.Select(s.Page.Size)
	d.orientationPopup.Select(s.Page.Orientation)
	d.topMarginField.SetText(s.Page.TopMargin.String())
	d.leftMarginField.SetText(s.Page.LeftMargin.String())
	d.bottomMarginField.SetText(s.Page.BottomMargin.String())
	d.rightMarginField.SetText(s.Page.RightMargin.String())
	d.blockLayoutField.SetText(s.BlockLayout.String())
	d.MarkForRedraw()
}

func (d *sheetSettingsDockable) syncSheet(full bool) {
	var entity *gurps.Entity
	if d.owner != nil {
		entity = d.owner.Entity()
	}
	for _, one := range allDockables() {
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
	return nil
}

func (d *sheetSettingsDockable) save(filePath string) error {
	return d.settings().Save(filePath)
}
