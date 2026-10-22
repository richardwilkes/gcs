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
	"cmp"
	"io/fs"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/progression"
	"github.com/richardwilkes/gcs/v5/model/paper"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/drag"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/weight"
)

var _ GroupedCloser = &sheetSettingsDockable{}

// sellMoneyColumns is the number of columns in the sell money list: the item's name, its value and a button to remove
// it.
const sellMoneyColumns = 3

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
	buyPercentField           *PercentageField
	sellPercentField          *PercentageField
	sellMoneyPanel            *unison.Panel
	sellMoneyRows             [][]unison.Paneler // the widgets of each row of the sell money list, for banding
	sellMoneyDropping         bool
	orientationPopup          *unison.PopupMenu[paper.Orientation]
	paperSizeField            *unison.Field
	topMarginField            *unison.Field
	leftMarginField           *unison.Field
	bottomMarginField         *unison.Field
	rightMarginField          *unison.Field
}

// sheetOption describes one of the boolean sheet settings the dockable presents as a checkbox. The setting is reached
// through an accessor rather than a captured pointer, since reset and load replace the settings being edited wholesale.
type sheetOption struct {
	title    string
	pageRef  string // a page reference to link to after the title, if any
	tooltip  string // an explanation of the setting, if the title alone is not enough
	field    func(s *gurps.SheetSettings) *bool
	inverted bool // the setting hides what the checkbox offers to show, so the box is checked while the setting is off
	fullSync bool // the setting can change the columns a sheet shows, which only a full rebuild can pick up
}

func (o *sheetOption) checked(s *gurps.SheetSettings) bool {
	return *o.field(s) != o.inverted
}

func (o *sheetOption) apply(s *gurps.SheetSettings, checked bool) {
	*o.field(s) = checked != o.inverted
}

type sheetOptionCheckBox struct {
	box    *unison.CheckBox
	option sheetOption
}

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
			title:    i18n.Text("Show all attacks"),
			field:    func(s *gurps.SheetSettings) *bool { return &s.ShowAllWeapons },
			fullSync: true,
		},
		{
			title:    i18n.Text("Hide unused columns in the melee & ranged attack tables"),
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
		{
			title:    i18n.Text("Disable traits whose prerequisites are unsatisfied"),
			tooltip:  i18n.Text("A trait whose prerequisites are not met, or whose level exceeds its maximum, is treated as disabled: it contributes no points, features or weapons to the sheet until they are met. The trait keeps its own enabled state and comes back into play on its own once its prerequisites are satisfied. Where prerequisites contradict one another, such as two traits that each require the other's absence, no choice satisfies them all, so the traits caught in the contradiction are left enabled and flagged as such."),
			field:    func(s *gurps.SheetSettings) *bool { return &s.EnforceTraitPrereqs },
			fullSync: true,
		},
	}
}

// ShowSheetSettings shows the Sheet Settings. Pass nil to edit the defaults, or a sheet to edit that sheet's.
func ShowSheetSettings(owner EntityPanel) {
	if ActivateDockableOfType(func(s *sheetSettingsDockable) bool { return s.owner == owner }) {
		return
	}
	d := &sheetSettingsDockable{owner: owner}
	d.initSettings(d, &settingsSpec{
		title:             sheetSettingsTabTitle(owner),
		ext:               gurps.SheetSettingsExt,
		loader:            d.load,
		saver:             d.save,
		resetter:          d.reset,
		addToStartToolbar: d.addToStartToolbar,
		initContent:       d.initContent,
	})
}

// sheetSettingsTabTitle returns the tab title for the sheet settings of the given owner, or for the defaults when the
// owner is nil. The character name must be substituted into an otherwise constant string, since i18n.Text looks the
// whole string up in the translation catalog; a string built with the name embedded produces a key that can never match
// an entry and can't be extracted for translation in the first place.
func sheetSettingsTabTitle(owner EntityPanel) string {
	if owner == nil {
		return i18n.Text("Default Sheet Settings")
	}
	return i18n.Text("Sheet Settings: %s", owner.Entity().Profile.Name)
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
	d.createTrading(content)
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
	// The description sits under the popup with no label of its own, so it is named outright for the screen readers
	// whose cursor moves into a document.
	desc.Accessibility.Name = i18n.Text("Damage Progression Description")
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
			d.syncSheet(gurps.SheetSettingsGeneralPart, option.fullSync)
		})
		if option.tooltip != "" {
			box.Tooltip = newWrappedTooltip(option.tooltip)
		}
		d.options = append(d.options, sheetOptionCheckBox{box: box, option: option})
	}
	content.AddChild(panel)
}

// addCheckBox adds a checkbox with the given title to the panel, followed by a link to the page reference when one is
// given, and returns it. The click callback is handed the state the box was left in. The reference is a single label
// with the link inside it, so a screen reader reads it whole and still finds the link.
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
	wrapper.SetLayout(&unison.FlexLayout{Columns: 2, HSpacing: unison.StdHSpacing})
	wrapper.AddChild(checkbox)
	label := NewSingleLineLabel()
	label.Font = checkbox.Font
	label.LinkRefs(openPageRefLink, pageRef)
	label.SetTitle("(" + pageRef + ")")
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

// createDecimalPlaces adds the section that controls how many decimal places the sheet rounds various numbers to. These
// affect only what is shown: the values themselves are always stored, and edited, at full precision.
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
			func(s *gurps.SheetSettings) *fxp.NumberFormat { return &s.HeightFormat }, nil),
		d.addNumberFormatRow(panel, i18n.Text("Body Weight"),
			i18n.Text("How many decimal places the character's weight in the Description block on the sheet is rounded to for display"),
			func(s *gurps.SheetSettings) *fxp.NumberFormat { return &s.BodyWeightFormat }, nil),
		d.addNumberFormatRow(panel, i18n.Text("Equipment Weight"),
			i18n.Text("How many decimal places the weight columns and the carried & other equipment totals on the sheet are rounded to for display"),
			func(s *gurps.SheetSettings) *fxp.NumberFormat { return &s.EquipmentWeightFormat }, nil),
		d.addNumberFormatRow(panel, i18n.Text("Equipment Value"),
			i18n.Text("How many decimal places the value columns and the carried & other equipment totals on the sheet are rounded to for display"),
			func(s *gurps.SheetSettings) *fxp.NumberFormat { return &s.EquipmentValueFormat },
			d.rebuildSellMoney), // The sell money list shows its values in this format.
	}
	content.AddChild(panel)
}

// addNumberFormatRow adds the popup that chooses how many decimal places one of the sheet's number formats rounds to,
// with the given tooltip, and the checkbox that has it pad with zeros out to that many. Either change calls changed,
// if given, before the open sheets are told of it.
func (d *sheetSettingsDockable) addNumberFormatRow(panel *unison.Panel, title, tooltip string, format func(s *gurps.SheetSettings) *fxp.NumberFormat, changed func()) sheetNumberFormatRow {
	row := sheetNumberFormatRow{format: format}
	current := format(d.settings())
	apply := func(change func(f *fxp.NumberFormat)) {
		change(format(d.settings()))
		if changed != nil {
			changed()
		}
	}
	row.popup = d.createSettingPopup(panel, title, fxp.DecimalPlaces, current.Places,
		func(item fxp.DecimalPlace) { apply(func(f *fxp.NumberFormat) { f.Places = item }) })
	row.popup.Tooltip = newWrappedTooltip(tooltip)
	row.pad = d.addCheckBox(panel, i18n.Text("Pad with zeros"), "", current.PadWithZeros, func(checked bool) {
		apply(func(f *fxp.NumberFormat) { f.PadWithZeros = checked })
		d.syncSheet(gurps.SheetSettingsGeneralPart, false)
	})
	// Every row's checkbox shows the same title, so each is named for the value it pads.
	row.pad.Accessibility.Name = i18n.Text("Pad %s with zeros", title)
	row.pad.Tooltip = newWrappedTooltip(i18n.Text(`Show trailing zeros out to the number of decimal places chosen, e.g. "7.50" rather than "7.5" at 2 decimal places. Has no effect when "As Needed" or "0" is chosen.`))
	return row
}

// createTrading adds the section that controls buying and selling equipment: the percentages of full price offered by
// default and the money that sales pay out in.
func (d *sheetSettingsDockable) createTrading(content *unison.Panel) {
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill})
	d.createHeader(panel, i18n.Text("Trading"), 2)
	d.buyPercentField = d.createPercentField(panel, i18n.Text("Buy Percent"),
		i18n.Text("The percentage of an item's full price that buying it costs, unless another is chosen when buying"),
		func(s *gurps.SheetSettings) *int { return &s.BuyPercent })
	d.sellPercentField = d.createPercentField(panel, i18n.Text("Sell Percent"),
		i18n.Text("The percentage of an item's full price that selling it brings in, unless another is chosen when selling"),
		func(s *gurps.SheetSettings) *int { return &s.SellPercent })
	listBorder := unison.NewCompoundBorder(
		unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.NewUniformInsets(1), false),
		unison.NewEmptyBorder(geom.NewUniformInsets(unison.StdHSpacing)),
	)
	label := NewFieldLeadingLabel(i18n.Text("Sell Money"), false)
	// Level with the list's first row rather than its border.
	label.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: listBorder.Insets().Top}))
	label.SetLayoutData(&unison.FlexLayoutData{HAlign: align.End, VAlign: align.Start})
	panel.AddChild(label)
	d.sellMoneyPanel = unison.NewPanel()
	d.sellMoneyPanel.SetLayout(&unison.FlexLayout{
		Columns:  sellMoneyColumns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	d.sellMoneyPanel.SetBorder(listBorder)
	d.sellMoneyPanel.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	d.sellMoneyPanel.Tooltip = newWrappedTooltip(i18n.Text("The money that selling equipment pays out in, using as few items as possible. Drag equipment tagged \"%s\", which is what buying spends, from any equipment list here to add a copy of it.",
		gurps.MoneyTag))
	d.installSellMoneyBanding()
	d.installSellMoneyDropSupport()
	d.rebuildSellMoney()
	panel.AddChild(d.sellMoneyPanel)
	content.AddChild(panel)
}

// createPercentField adds a field for the percentage of the settings that value reaches, with the given title and
// tooltip, and returns it. The open sheets are told only of a change, since sync() hands the field the value it already
// holds.
func (d *sheetSettingsDockable) createPercentField(panel *unison.Panel, title, tooltip string, value func(s *gurps.SheetSettings) *int) *PercentageField {
	panel.AddChild(NewFieldLeadingLabel(title, false))
	field := NewPercentageField(nil, "", title, func() int { return *value(d.settings()) },
		func(v int) {
			if current := value(d.settings()); *current != v {
				*current = v
				d.syncSheet(gurps.SheetSettingsGeneralPart, false)
			}
		}, gurps.MinTradePercent, gurps.MaxTradePercent, false, false)
	field.Tooltip = newWrappedTooltip(tooltip)
	panel.AddChild(field)
	return field
}

// installSellMoneyDropSupport has the sell money list take equipment dragged from a list, adding what it can, and
// decline a drag that holds nothing it can add, so that the pointer says so.
func (d *sheetSettingsDockable) installSellMoneyDropSupport() {
	p := d.sellMoneyPanel
	p.CanAcceptDropCallback = func(di drag.Info) bool { return di.HasDataType(equipmentDragKey.UTI) }
	over := func(di drag.Info, _ geom.Point, _ mod.Modifiers) drag.Op {
		if !di.HasDataType(equipmentDragKey.UTI) || len(d.sellMoneyFrom(draggedEquipmentTableData())) == 0 {
			d.showSellMoneyDropTarget(false)
			return drag.None
		}
		d.showSellMoneyDropTarget(true)
		return drag.Copy
	}
	p.DragEnteredCallback = over
	p.DragUpdatedCallback = over
	p.DragExitedCallback = func() { d.showSellMoneyDropTarget(false) }
	p.DropCallback = func(di drag.Info, _ geom.Point, _ mod.Modifiers) bool {
		d.showSellMoneyDropTarget(false)
		return di.HasDataType(equipmentDragKey.UTI) && d.addSellMoney(draggedEquipmentTableData())
	}
	p.DrawOverCallback = func(gc *unison.Canvas, _ geom.Rect) {
		if d.sellMoneyDropping {
			r := p.ContentRect(true)
			paint := unison.ThemeWarning.Paint(gc, r, paintstyle.Fill)
			paint.SetColorFilter(unison.Alpha30Filter())
			gc.DrawRect(r, paint)
		}
	}
}

func (d *sheetSettingsDockable) showSellMoneyDropTarget(show bool) {
	if d.sellMoneyDropping != show {
		d.sellMoneyDropping = show
		flushDragFeedback(d.sellMoneyPanel)
	}
}

// draggedEquipmentTableData returns the data of the equipment table drag in progress, or nil if what is being dragged
// is something else.
func draggedEquipmentTableData() *unison.TableDragData[*Node[*gurps.Equipment]] {
	if tableData, ok := draggedTableData.(*unison.TableDragData[*Node[*gurps.Equipment]]); ok {
		return tableData
	}
	return nil
}

// sellMoneyFrom returns the equipment rows of the drag data that can be added to the sell money (see
// gurps.SheetSettings.CanAddSellMoney).
func (d *sheetSettingsDockable) sellMoneyFrom(tableData *unison.TableDragData[*Node[*gurps.Equipment]]) []*gurps.Equipment {
	if tableData == nil {
		return nil
	}
	s := d.settings()
	var list []*gurps.Equipment
	for _, row := range tableData.Rows {
		if eqp := row.Data(); s.CanAddSellMoney(eqp) {
			list = append(list, eqp)
		}
	}
	return list
}

// addSellMoney adds a copy of each equipment row of the drag data that can be sell money (see sellMoneyFrom) to it,
// returning true if any were added.
func (d *sheetSettingsDockable) addSellMoney(tableData *unison.TableDragData[*Node[*gurps.Equipment]]) bool {
	list := d.sellMoneyFrom(tableData)
	if len(list) == 0 {
		return false
	}
	s := d.settings()
	from := libraryFileFromTable(tableData.Table)
	added := false
	for _, eqp := range list {
		if s.AddSellMoney(eqp, from) {
			added = true
		}
	}
	if added {
		d.rebuildSellMoney()
		d.syncSheet(gurps.SheetSettingsGeneralPart, false)
	}
	return added
}

// installSellMoneyBanding has the sell money list draw its rows on the background a table gives its rows, banding
// every other one as a table does. The border is drawn after this, over the edge of the fill.
func (d *sheetSettingsDockable) installSellMoneyBanding() {
	p := d.sellMoneyPanel
	p.DrawCallback = func(gc *unison.Canvas, _ geom.Rect) {
		area := p.ContentRect(true)
		gc.DrawRect(area, unison.ThemeBelowSurface.Paint(gc, area, paintstyle.Fill))
		for i, row := range d.sellMoneyRows {
			if i&1 == 1 {
				band := sellMoneyRowBounds(row, area)
				gc.DrawRect(band, unison.ThemeBanding.Paint(gc, band, paintstyle.Fill))
			}
		}
	}
}

// sellMoneyRowBounds returns the band a row of the sell money list is drawn on: as wide as the list, and as tall as the
// row's widgets plus half of the spacing to the rows on either side, so that the bands of neighboring rows meet.
func sellMoneyRowBounds(row []unison.Paneler, area geom.Rect) geom.Rect {
	var bounds geom.Rect
	for i, w := range row {
		if i == 0 {
			bounds = w.AsPanel().FrameRect()
		} else {
			bounds = bounds.Union(w.AsPanel().FrameRect())
		}
	}
	bounds.X = area.X
	bounds.Width = area.Width
	bounds.Y -= unison.StdVSpacing / 2
	bounds.Height += unison.StdVSpacing
	return bounds
}

// sellMoneyEntry is a sell money item along with its value, worked out once for sorting and display.
type sellMoneyEntry struct {
	eqp   *gurps.Equipment
	value fxp.Int
}

// sortedSellMoney returns the sell money of the settings in the order the list shows it: by increasing value, those of
// equal value in the order they were added.
func (d *sheetSettingsDockable) sortedSellMoney() []sellMoneyEntry {
	money := d.settings().SellMoney
	entries := make([]sellMoneyEntry, 0, len(money))
	for _, eqp := range money {
		entries = append(entries, sellMoneyEntry{eqp: eqp, value: eqp.AdjustedValue()})
	}
	slices.SortStableFunc(entries, func(a, b sellMoneyEntry) int { return cmp.Compare(a.value, b.value) })
	return entries
}

// rebuildSellMoney fills the sell money list from the settings: a row for each item, in increasing order of value (see
// sortedSellMoney), showing its name and value, with a button that removes it, or a hint saying how to add one when
// there are none.
func (d *sheetSettingsDockable) rebuildSellMoney() {
	p := d.sellMoneyPanel
	p.RemoveAllChildren()
	d.sellMoneyRows = nil
	entries := d.sortedSellMoney()
	if len(entries) == 0 {
		hint := unison.NewLabel()
		hint.Font = fonts.FieldSecondary
		hint.OnBackgroundInk = dimmedTextColor
		hint.SetTitle(i18n.Text("Drag equipment tagged \"%s\" from any equipment list here.", gurps.MoneyTag))
		hint.SetLayoutData(&unison.FlexLayoutData{HSpan: sellMoneyColumns})
		p.AddChild(hint)
	}
	for i, entry := range entries {
		id := entry.eqp.TID
		name := entry.eqp.String()
		label := unison.NewLabel()
		label.OnBackgroundInk = unison.ThemeOnBelowSurface
		label.SetTitle(name)
		label.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, VAlign: align.Middle, HGrab: true})
		p.AddChild(label)
		value := unison.NewLabel()
		value.OnBackgroundInk = unison.ThemeOnBelowSurface
		value.SetTitle(i18n.Text("$%s", d.settings().FormatEquipmentValue(entry.value)))
		value.SetLayoutData(&unison.FlexLayoutData{HAlign: align.End, VAlign: align.Middle})
		p.AddChild(value)
		remove := unison.NewSVGButton(unison.TrashSVG)
		remove.Accessibility.Name = i18n.Text("Remove %s", name)
		remove.Tooltip = newWrappedTooltip(remove.Accessibility.Name)
		remove.ClickCallback = func() {
			if !d.settings().RemoveSellMoney(id) {
				return
			}
			// The rebuild takes every button off the page, this one included, and the focus doesn't leave a removed
			// panel of its own accord, so it is moved on when it was here.
			hadFocus := false
			if wnd := remove.Window(); wnd != nil {
				hadFocus = wnd.CurrentFocus() == remove.AsPanel()
			}
			d.rebuildSellMoney()
			d.syncSheet(gurps.SheetSettingsGeneralPart, false)
			if hadFocus {
				d.focusSellMoneyRemoveButton(i)
			}
		}
		remove.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Middle})
		p.AddChild(remove)
		d.sellMoneyRows = append(d.sellMoneyRows, []unison.Paneler{label, value, remove})
	}
	p.MarkForLayoutRecursivelyUpward()
	p.MarkForRedraw()
}

// focusSellMoneyRemoveButton gives the keyboard focus to the button that removes the sell money entry at index, or the
// last entry when there are fewer, or to the sell percent field when there are none.
func (d *sheetSettingsDockable) focusSellMoneyRemoveButton(index int) {
	var buttons []*unison.Button
	for _, child := range d.sellMoneyPanel.Children() {
		if b, ok := child.Self.(*unison.Button); ok {
			buttons = append(buttons, b)
		}
	}
	if len(buttons) == 0 {
		d.sellPercentField.RequestFocus()
		return
	}
	buttons[min(index, len(buttons)-1)].RequestFocus()
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
	label := NewFieldLeadingLabel(i18n.Text("Paper Size"), false)
	panel.AddChild(label)
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
	// The wrapper separates the field from its label, so the field is pointed at the label explicitly.
	field := labelControl(unison.NewField(), label)
	field.SetText(current)
	field.ValidateCallback = func() bool {
		_, _, valid := gurps.ParsePageSize(field.Text())
		return valid
	}
	field.ModifiedCallback = func(_, after *unison.FieldState) {
		if width, height, valid := gurps.ParsePageSize(after.Text); valid {
			set(gurps.ToPageSize(width, height))
			d.syncSheet(gurps.SheetSettingsGeneralPart, false)
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
	AddHelpToInfoPop(info, wrapTextForTooltip(i18n.Text(`Enter a standard paper size (e.g., one of %s) or a custom size (e.g., "8.5in x 11in", "210mm x 297mm")`, buffer.String())))
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
			d.syncSheet(gurps.SheetSettingsGeneralPart, false)
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
		d.syncSheet(gurps.SheetSettingsGeneralPart, false)
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
	d.syncSheet(gurps.SheetSettingsAllParts, true)
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
	SetFieldValue(d.buyPercentField.Field, d.buyPercentField.Format(s.BuyPercent))
	SetFieldValue(d.sellPercentField.Field, d.sellPercentField.Format(s.SellPercent))
	d.rebuildSellMoney()
	d.paperSizeField.SetText(s.Page.Size)
	d.orientationPopup.Select(s.Page.Orientation)
	d.topMarginField.SetText(s.Page.TopMargin.String())
	d.leftMarginField.SetText(s.Page.LeftMargin.String())
	d.bottomMarginField.SetText(s.Page.BottomMargin.String())
	d.rightMarginField.SetText(s.Page.RightMargin.String())
	d.MarkForRedraw()
}

// syncSheet reports a change to the given parts of the settings this dockable edits to the open dockables that show
// them.
func (d *sheetSettingsDockable) syncSheet(parts gurps.SheetSettingsParts, full bool) {
	var entity *gurps.Entity
	if d.owner != nil {
		entity = d.owner.Entity()
	} else {
		// The global sheet settings were just changed, so republish the snapshot background parses read.
		gurps.SyncGlobalSheetSettings()
	}
	notifySheetSettingsUpdated(entity, parts, full)
}

// notifySheetSettingsUpdated reports a change to the given parts of the entity's sheet settings, or of the defaults
// when entity is nil, to every open dockable that responds to such changes, first invalidating the display data
// derived from the settings as they were.
func notifySheetSettingsUpdated(entity *gurps.Entity, parts gurps.SheetSettingsParts, fullRebuild bool) {
	invalidateDisplayData()
	for _, one := range AllDockables() {
		if responder, ok := one.(gurps.SheetSettingsResponder); ok {
			responder.SheetSettingsUpdated(entity, parts, fullRebuild)
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
	d.syncSheet(gurps.SheetSettingsAllParts, true)
	return nil
}

func (d *sheetSettingsDockable) save(filePath string) error {
	return d.settings().Save(filePath)
}
