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
	"image/color"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/paper"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/drag"
	uncheck "github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/side"
)

var (
	_ gurps.SheetSettingsResponder = &sheetSettingsRecorder{}
	_ unison.Dockable              = &sheetSettingsRecorder{}
	_ EntityPanel                  = &entityPanelForTest{}
)

// sheetSettingsRecorder is a Dockable that records the sheet settings updates it is sent.
type sheetSettingsRecorder struct {
	unison.Panel
	entities []*gurps.Entity            // one entry per notification, holding the entity whose settings changed
	parts    []gurps.SheetSettingsParts // one entry per notification, holding the parts that changed
	updates  []bool                     // one entry per notification, holding the fullRebuild flag
}

func (r *sheetSettingsRecorder) TitleIcon(_ geom.Size) unison.Drawable { return nil }
func (r *sheetSettingsRecorder) Title() string                         { return "recorder" }
func (r *sheetSettingsRecorder) Tooltip() string                       { return "" }
func (r *sheetSettingsRecorder) Modified() bool                        { return false }

func (r *sheetSettingsRecorder) SheetSettingsUpdated(entity *gurps.Entity, parts gurps.SheetSettingsParts,
	fullRebuild bool,
) {
	r.entities = append(r.entities, entity)
	r.parts = append(r.parts, parts)
	r.updates = append(r.updates, fullRebuild)
}

func (r *sheetSettingsRecorder) sawFullRebuild() bool {
	for _, full := range r.updates {
		if full {
			return true
		}
	}
	return false
}

type entityPanelForTest struct {
	unison.Panel
	entity *gurps.Entity
}

func (p *entityPanelForTest) Entity() *gurps.Entity { return p.entity }

// newTestSheetSettingsDockable returns a dockable whose widgets have been built from the given owner's settings, along
// with a recorder docked where syncSheet will find it. The document dock is restored when the test finishes.
func newTestSheetSettingsDockable(t *testing.T, owner EntityPanel) (*sheetSettingsDockable, *sheetSettingsRecorder) {
	t.Helper()
	uxtest.SwapForTest(t, &Workspace.DocumentDock, NewDocumentDock())
	recorder := &sheetSettingsRecorder{}
	recorder.Self = recorder
	Workspace.DocumentDock.DockTo(recorder, nil, side.Left)
	d := &sheetSettingsDockable{owner: owner}
	d.Self = d
	d.initContent(unison.NewPanel())
	return d, recorder
}

// newEntityPanelWithFlaggedSettings returns an EntityPanel whose sheet settings differ from the global defaults only in
// the boolean options the dockable presents as checkboxes, so that nothing in sync() fires a change callback of its
// own.
func newEntityPanelWithFlaggedSettings() *entityPanelForTest {
	p := &entityPanelForTest{entity: gurps.NewEntity()}
	p.Self = p
	s := gurps.GlobalSettings().Sheet.Clone(p.entity)
	flipCheckboxOptions(s)
	p.entity.SheetSettings = s
	s.SetOwningEntity(p.entity)
	return p
}

// flipCheckboxOptions inverts every boolean option the sheet settings dockable exposes as a checkbox, leaving the
// popup- and field-backed settings alone.
func flipCheckboxOptions(s *gurps.SheetSettings) {
	s.HideSourceMismatch = !s.HideSourceMismatch
	s.HidePageRefColumn = !s.HidePageRefColumn
	s.HideTLColumn = !s.HideTLColumn
	s.HideLCColumn = !s.HideLCColumn
	s.ShowTraitModifierAdj = !s.ShowTraitModifierAdj
	s.ShowEquipmentModifierAdj = !s.ShowEquipmentModifierAdj
	s.ShowAllWeapons = !s.ShowAllWeapons
	s.HideUnusedWeaponColumns = !s.HideUnusedWeaponColumns
	s.ShowSpellAdj = !s.ShowSpellAdj
	s.UseTitleInFooter = !s.UseTitleInFooter
	s.ShowLiftingSTDamage = !s.ShowLiftingSTDamage
	s.ShowIQBasedDamage = !s.ShowIQBasedDamage
	s.HideZeroValueConditionalMods = !s.HideZeroValueConditionalMods
	s.EnforceTraitPrereqs = !s.EnforceTraitPrereqs
	s.UseMultiplicativeModifiers = !s.UseMultiplicativeModifiers
	s.UseHalfStatDefaults = !s.UseHalfStatDefaults
	s.UseModifyingDicePlusAdds = !s.UseModifyingDicePlusAdds
	s.ExcludeUnspentPointsFromTotal = !s.ExcludeUnspentPointsFromTotal
	// Only the padding flags are flipped for the number formats: the decimal place choices are popup-backed, so
	// changing them here would make sync() alter a popup's selection and fire its callback.
	s.HeightFormat.PadWithZeros = !s.HeightFormat.PadWithZeros
	s.BodyWeightFormat.PadWithZeros = !s.BodyWeightFormat.PadWithZeros
	s.EquipmentWeightFormat.PadWithZeros = !s.EquipmentWeightFormat.PadWithZeros
	s.EquipmentValueFormat.PadWithZeros = !s.EquipmentValueFormat.PadWithZeros
}

// TestSheetSettingsResetNotifiesSheets checks that resetting the sheet settings tells the open sheets to rebuild.
// sync() only pushes the new values into this dockable's own widgets, and assigning a CheckBox's State does not fire
// its ClickCallback, so settings differing only in the checkbox-backed options used to leave the open sheet showing the
// old columns and values until some unrelated edit happened to trigger a rebuild.
func TestSheetSettingsResetNotifiesSheets(t *testing.T) {
	c := check.New(t)
	owner := newEntityPanelWithFlaggedSettings()
	d, recorder := newTestSheetSettingsDockable(t, owner)

	d.reset()

	c.Equal(gurps.GlobalSettings().Sheet.HideTLColumn, owner.entity.SheetSettings.HideTLColumn,
		"reset must restore the default options")
	c.True(len(recorder.updates) != 0, "reset must notify the open sheets")
	c.True(recorder.sawFullRebuild(), "reset must ask for a full rebuild, since column visibility may have changed")
	c.Equal(gurps.SheetSettingsAllParts, recorder.parts[0], "reset replaces the settings wholesale, so every part changed")
}

// TestSheetSettingsLoadNotifiesSheets checks the same for importing a sheet settings file: a file differing only in the
// checkbox-backed options must still refresh the open sheets.
func TestSheetSettingsLoadNotifiesSheets(t *testing.T) {
	c := check.New(t)
	owner := newEntityPanelWithFlaggedSettings()
	d, recorder := newTestSheetSettingsDockable(t, owner)

	// The file holds the global defaults, i.e. it differs from the owner's current settings only in the checkboxes.
	dir := t.TempDir()
	name := "settings." + gurps.SheetSettingsExt
	c.NoError(gurps.GlobalSettings().Sheet.Save(filepath.Join(dir, name)))
	c.NoError(d.load(os.DirFS(dir), name))

	c.Equal(gurps.GlobalSettings().Sheet.HideTLColumn, owner.entity.SheetSettings.HideTLColumn,
		"load must apply the file's options")
	c.True(len(recorder.updates) != 0, "load must notify the open sheets")
	c.True(recorder.sawFullRebuild(), "load must ask for a full rebuild, since column visibility may have changed")
	c.Equal(gurps.SheetSettingsAllParts, recorder.parts[0], "load replaces the settings wholesale, so every part changed")
}

// TestSheetSettingsSyncDoesNotFireCheckBoxCallbacks documents why the notification above has to be explicit: sync()
// assigns each CheckBox's State directly, which unison does not treat as a click, so the callbacks that would otherwise
// push the change out to the sheets never run.
func TestSheetSettingsSyncDoesNotFireCheckBoxCallbacks(t *testing.T) {
	c := check.New(t)
	owner := newEntityPanelWithFlaggedSettings()
	d, recorder := newTestSheetSettingsDockable(t, owner)

	// The replacement differs from the owner's current settings only in the checkbox-backed options.
	replacement := gurps.GlobalSettings().Sheet.Clone(owner.entity)
	owner.entity.SheetSettings = replacement
	replacement.SetOwningEntity(owner.entity)
	d.sync()

	c.Equal(0, len(recorder.updates), "sync() alone must not be relied upon to notify the sheets")
}

type numberFormatWidgets struct {
	name   string
	popup  *unison.PopupMenu[fxp.DecimalPlace]
	pad    *unison.CheckBox
	format *fxp.NumberFormat
}

// allNumberFormatWidgets returns the widgets for each of the four display formats in the given settings. The rows are
// taken in the order createDecimalPlaces builds them, and each is paired with its format here independently of the
// accessor the row carries, so that a row pointed at the wrong format is caught.
func allNumberFormatWidgets(d *sheetSettingsDockable, s *gurps.SheetSettings) []numberFormatWidgets {
	rows := d.numberFormats
	return []numberFormatWidgets{
		{name: "height", popup: rows[0].popup, pad: rows[0].pad, format: &s.HeightFormat},
		{name: "body weight", popup: rows[1].popup, pad: rows[1].pad, format: &s.BodyWeightFormat},
		{name: "equipment weight", popup: rows[2].popup, pad: rows[2].pad, format: &s.EquipmentWeightFormat},
		{name: "equipment value", popup: rows[3].popup, pad: rows[3].pad, format: &s.EquipmentValueFormat},
	}
}

// TestSheetSettingsSyncSelectsNumberFormats checks that sync() pushes each of the four number format settings into its
// own decimal places popup and padding checkbox, so that loading a settings file or resetting to the defaults leaves
// the widgets showing what is in effect. Every format is given a different choice, so a widget reading from the wrong
// format is caught.
func TestSheetSettingsSyncSelectsNumberFormats(t *testing.T) {
	c := check.New(t)
	owner := newEntityPanelWithFlaggedSettings()
	d, _ := newTestSheetSettingsDockable(t, owner)
	widgets := allNumberFormatWidgets(d, owner.entity.SheetSettings)
	for i, w := range widgets {
		*w.format = fxp.NumberFormat{Places: fxp.DecimalPlace(i + 1), PadWithZeros: i%2 == 0}
	}
	d.sync()
	for i, w := range widgets {
		selected, ok := w.popup.Selected()
		c.True(ok, "the %s popup must have a selection", w.name)
		c.Equal(fxp.DecimalPlace(i+1), selected, "the %s popup must show its setting's decimal places", w.name)
		c.Equal(uncheck.FromBool(i%2 == 0), w.pad.State, "the %s padding checkbox must show its setting", w.name)
	}
}

// TestSheetSettingsNumberFormatWidgetsWriteTheirOwnSetting checks the other direction: that each decimal places popup
// and padding checkbox writes to its own format and to no other, and that each change notifies the open sheets. The
// four rows differ only in the accessor each is given, so a slip there is the most likely error.
func TestSheetSettingsNumberFormatWidgetsWriteTheirOwnSetting(t *testing.T) {
	c := check.New(t)
	owner := newEntityPanelWithFlaggedSettings()
	d, recorder := newTestSheetSettingsDockable(t, owner)
	widgets := allNumberFormatWidgets(d, owner.entity.SheetSettings)
	for i, w := range widgets {
		before := make([]fxp.NumberFormat, len(widgets))
		for j, other := range widgets {
			before[j] = *other.format
		}
		updates := len(recorder.updates)

		// A choice the popup does not already show, so that selecting it fires the popup's callback.
		places := fxp.DecimalPlace(i + 1)
		if places == before[i].Places {
			places++
		}
		w.popup.Select(places)
		c.Equal(places, w.format.Places, "the %s popup must write the %s format", w.name, w.name)

		w.pad.State = uncheck.FromBool(!before[i].PadWithZeros)
		w.pad.ClickCallback()
		c.Equal(!before[i].PadWithZeros, w.format.PadWithZeros, "the %s checkbox must write the %s format", w.name,
			w.name)

		for j, other := range widgets {
			if j != i {
				c.Equal(before[j], *other.format, "the %s widgets must not touch the %s format", w.name, other.name)
			}
		}
		c.Equal(updates+2, len(recorder.updates), "each %s change must notify the open sheets", w.name)
	}
}

// TestSheetSettingsDisplayPopupsWriteTheirOwnSetting checks that each popup saying where a kind of detail is shown
// writes its own setting and no other, notifying the open sheets, and that sync() shows each popup its setting. The
// popups are paired with their settings here independently of the accessor each carries, since the rows differ only
// in that accessor and a slip there is the most likely error.
func TestSheetSettingsDisplayPopupsWriteTheirOwnSetting(t *testing.T) {
	c := check.New(t)
	owner := newEntityPanelWithFlaggedSettings()
	d, recorder := newTestSheetSettingsDockable(t, owner)
	s := owner.entity.SheetSettings
	settings := func() []display.Option {
		return []display.Option{s.UserDescriptionDisplay, s.ModifiersDisplay, s.NotesDisplay, s.SkillLevelAdjDisplay}
	}
	c.Equal(len(settings()), len(d.displayPopups), "a popup for each kind of detail")
	for i, one := range d.displayPopups {
		before := settings()
		updates := len(recorder.updates)
		// A choice the popup does not already show, so that selecting it fires the popup's callback.
		choice := display.Options[(slices.Index(display.Options, before[i])+1)%len(display.Options)]
		one.popup.Select(choice)
		expected := slices.Clone(before)
		expected[i] = choice
		c.Equal(expected, settings(), "popup #%d must write its own setting and no other", i)
		c.Equal(updates+1, len(recorder.updates), "popup #%d must notify the open sheets", i)
	}
	s.UserDescriptionDisplay = display.Options[0]
	s.ModifiersDisplay = display.Options[1]
	s.NotesDisplay = display.Options[2]
	s.SkillLevelAdjDisplay = display.Options[3]
	d.sync()
	for i, one := range d.displayPopups {
		selected, ok := one.popup.Selected()
		c.True(ok, "popup #%d must have a selection", i)
		c.Equal(settings()[i], selected, "popup #%d must show its setting", i)
	}
}

// TestSheetSettingsMarginFieldsWriteTheirOwnSetting checks the same for the page margin fields: each writes its own
// margin and no other, notifying the open sheets, and sync() shows each field its margin.
func TestSheetSettingsMarginFieldsWriteTheirOwnSetting(t *testing.T) {
	c := check.New(t)
	owner := newEntityPanelWithFlaggedSettings()
	d, recorder := newTestSheetSettingsDockable(t, owner)
	page := owner.entity.SheetSettings.Page
	margins := func() []paper.Length {
		return []paper.Length{page.TopMargin, page.BottomMargin, page.LeftMargin, page.RightMargin}
	}
	c.Equal(len(margins()), len(d.marginFields), "a field for each margin")
	for i, one := range d.marginFields {
		before := margins()
		updates := len(recorder.updates)
		value := paper.Length{Length: float64(i + 1), Units: paper.Centimeter}
		one.field.SetText(value.String())
		expected := slices.Clone(before)
		expected[i] = value
		c.Equal(expected, margins(), "field #%d must write its own margin and no other", i)
		c.Equal(updates+1, len(recorder.updates), "field #%d must notify the open sheets", i)
	}
	page.TopMargin = paper.Length{Length: 5, Units: paper.Millimeter}
	page.BottomMargin = paper.Length{Length: 6, Units: paper.Millimeter}
	page.LeftMargin = paper.Length{Length: 7, Units: paper.Millimeter}
	page.RightMargin = paper.Length{Length: 8, Units: paper.Millimeter}
	d.sync()
	for i, one := range d.marginFields {
		c.Equal(margins()[i].String(), one.field.Text(), "field #%d must show its margin", i)
	}
}

// checkBoxOptionsOf returns the settings' boolean options by name: the top-level boolean settings along with the
// padding flag of each number format, which are the settings the dockable presents as checkboxes.
func checkBoxOptionsOf(s *gurps.SheetSettings) map[string]bool {
	options := make(map[string]bool)
	v := reflect.ValueOf(s.SheetSettingsData)
	for i := range v.NumField() {
		field := v.Field(i)
		name := v.Type().Field(i).Name
		switch {
		case field.Kind() == reflect.Bool:
			options[name] = field.Bool()
		case field.Type() == reflect.TypeFor[fxp.NumberFormat]():
			options[name+".PadWithZeros"] = field.FieldByName("PadWithZeros").Bool()
		}
	}
	return options
}

// TestSheetSettingsCheckBoxesEachWriteTheirOwnOption checks that the checkboxes between them cover exactly the boolean
// options the dockable is meant to present, each writing an option of its own: clicking every box once must leave the
// settings the same as flipping every one of those options directly. A box wired to another box's option would flip it
// back again, and one wired to an option that isn't meant to be a checkbox would show up as a difference too. Each
// click has to notify the open sheets as well.
func TestSheetSettingsCheckBoxesEachWriteTheirOwnOption(t *testing.T) {
	c := check.New(t)
	owner := newEntityPanelWithFlaggedSettings()
	d, recorder := newTestSheetSettingsDockable(t, owner)
	expected := owner.entity.SheetSettings.Clone(owner.entity)
	flipCheckboxOptions(expected)
	type namedBox struct {
		name string
		box  *unison.CheckBox
	}
	boxes := make([]namedBox, 0, len(d.options)+len(d.numberFormats))
	for _, one := range d.options {
		boxes = append(boxes, namedBox{name: one.option.title, box: one.box})
	}
	for i, row := range d.numberFormats {
		boxes = append(boxes, namedBox{name: "pad #" + strconv.Itoa(i), box: row.pad})
	}
	for i, one := range boxes {
		one.box.State = uncheck.FromBool(one.box.State != uncheck.On)
		one.box.ClickCallback()
		c.Equal(i+1, len(recorder.updates), "clicking %q must notify the open sheets", one.name)
		c.Equal(gurps.SheetSettingsGeneralPart, recorder.parts[i], "clicking %q changes a general setting", one.name)
	}
	c.Equal(checkBoxOptionsOf(expected), checkBoxOptionsOf(owner.entity.SheetSettings),
		"clicking every checkbox once must flip every checkbox-backed option and nothing else")
}

// TestSheetSettingsTabTitle checks that the character name is substituted into the tab title rather than built into the
// string handed to i18n.Text, which would produce a per-character lookup key no catalog entry can ever match.
func TestSheetSettingsTabTitle(t *testing.T) {
	c := check.New(t)
	i18n.SetLocalizer(func(text string) string {
		if text == "Sheet Settings: %s" {
			return "Sheet Settings for %s"
		}
		return text
	})
	t.Cleanup(func() { i18n.SetLocalizer(nil) })

	entity := gurps.NewEntity()
	entity.Profile.Name = "Bob"
	c.Equal("Sheet Settings for Bob", sheetSettingsTabTitle(&entityPanelForTest{entity: entity}),
		"the translated title must be used, with the name substituted into it")
	c.Equal("Default Sheet Settings", sheetSettingsTabTitle(nil), "a nil owner yields the defaults title")
}

// tradeTestMoney is an equipment library file, in the user library, of the money the trading tests pay out in: a Silver
// Eagle worth $20 and a Copper Penny worth $1, both tagged as money, along with a Purse, which is a container.
type tradeTestMoney struct {
	path    string
	libFile gurps.LibraryFile
	eagle   *gurps.Equipment
	penny   *gurps.Equipment
	purse   *gurps.Equipment
}

// newTradeTestMoney writes the money library file into the user library the global settings hold, which the test must
// already have replaced (see uxtest.UseTestLibraries).
func newTradeTestMoney(t *testing.T, c check.Checker) *tradeTestMoney {
	t.Helper()
	user := gurps.GlobalSettings().Libraries.User()
	name := "Money" + gurps.EquipmentExt
	m := &tradeTestMoney{
		path:    filepath.Join(user.Path(false), name),
		libFile: gurps.LibraryFile{Library: user.Key(), Path: name},
		eagle:   newTradeTestItem(nil, "Silver Eagle", "20", fxp.One, gurps.MoneyTag),
		penny:   newTradeTestItem(nil, "Copper Penny", "1", fxp.One, gurps.MoneyTag),
		purse:   gurps.NewEquipment(nil, nil, true),
	}
	m.purse.Name = "Purse"
	c.NoError(os.MkdirAll(filepath.Dir(m.path), 0o750))
	c.NoError(gurps.SaveEquipment([]*gurps.Equipment{m.eagle, m.penny, m.purse}, m.path))
	return m
}

// source returns the source of the item in the money library file.
func (m *tradeTestMoney) source(eqp *gurps.Equipment) gurps.Source {
	return gurps.Source{LibraryFile: m.libFile, TID: eqp.TID}
}

// newTradeTestItem returns a piece of equipment owned by owner with the given name, value, quantity and tags.
func newTradeTestItem(owner gurps.DataOwner, name, value string, quantity fxp.Int, tags ...string) *gurps.Equipment {
	eqp := gurps.NewEquipment(owner, nil, false)
	eqp.Name = name
	eqp.BaseValue = value
	eqp.Quantity = quantity
	eqp.Tags = tags
	return eqp
}

// equipmentRowsFor returns the table's rows holding the equipment, in the order given, failing the test if any is not
// among the table's top-level rows.
func equipmentRowsFor(t *testing.T, table *unison.Table[*Node[*gurps.Equipment]], items ...*gurps.Equipment) []*Node[*gurps.Equipment] {
	t.Helper()
	rows := make([]*Node[*gurps.Equipment], 0, len(items))
	for _, eqp := range items {
		found := false
		for _, row := range table.RootRows() {
			if row.Data() == eqp {
				rows = append(rows, row)
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("the table has no row for %s", eqp.Name)
		}
	}
	return rows
}

// dropOnSellMoney drags the rows of the equipment from the table over the dockable's sell money list and drops them
// (see dropRowsOnSellMoney), returning what the drop reported. It must run on the test goroutine, since it fails the
// test outright when the table lacks a row for an item.
func dropOnSellMoney(t *testing.T, c check.Checker, d *sheetSettingsDockable, table *unison.Table[*Node[*gurps.Equipment]], items ...*gurps.Equipment) bool {
	t.Helper()
	return dropRowsOnSellMoney(t, c, d, table, equipmentRowsFor(t, table, items...))
}

// dropRowsOnSellMoney drags the table's rows over the dockable's sell money list and drops them, checking that the list
// offers to take the drag exactly when the drop adds something, and returning what the drop reported. It may run on
// the UI goroutine of a headless workspace, since it only ever reports failures rather than ending the test.
func dropRowsOnSellMoney(t *testing.T, c check.Checker, d *sheetSettingsDockable, table *unison.Table[*Node[*gurps.Equipment]], rows []*Node[*gurps.Equipment]) bool {
	t.Helper()
	uxtest.SwapForTest(t, &draggedTableData, any(&unison.TableDragData[*Node[*gurps.Equipment]]{Table: table, Rows: rows}))
	di := &fakeDragInfo{types: []string{equipmentDragKey.UTI}}
	p := d.sellMoneyPanel
	c.True(p.CanAcceptDropCallback(di), "the sell money list must accept equipment")
	op := p.DragEnteredCallback(di, geom.Point{}, 0)
	c.Equal(op, p.DragUpdatedCallback(di, geom.Point{}, 0), "moving within offers what entering did")
	c.Equal(op == drag.Copy, d.sellMoneyDropping, "the list is highlighted only while a drag it can take is over it")
	dropped := p.DropCallback(di, geom.Point{}, 0)
	c.Equal(dropped, op == drag.Copy, "a copy is offered exactly when the drop adds something")
	c.False(d.sellMoneyDropping, "the highlight goes once the drop is made")
	return dropped
}

// sellMoneyRemoveButton returns the button that removes the named sell money item, failing the test if there is none.
func sellMoneyRemoveButton(t *testing.T, d *sheetSettingsDockable, name string) *unison.Button {
	t.Helper()
	for _, b := range uxtest.PanelsOfType[*unison.Button](d.sellMoneyPanel) {
		if b.Accessibility.Name == "Remove "+name {
			return b
		}
	}
	t.Fatalf("the sell money list has no button to remove %s", name)
	return nil
}

// newTradingSettingsTest replaces the libraries with a set holding the money library file, which it returns along with
// a library dockable showing that file, and keeps drag feedback from trying to draw.
func newTradingSettingsTest(t *testing.T, c check.Checker) (*tradeTestMoney, *TableDockable[*gurps.Equipment]) {
	t.Helper()
	RegisterKnownFileTypes()
	uxtest.UseTestLibraries(t, c)
	uxtest.SwapForTest(t, &flushDragFeedback, func(*unison.Panel) {})
	money := newTradeTestMoney(t, c)
	return money, NewEquipmentTableDockable(money.path, []*gurps.Equipment{money.eagle, money.penny, money.purse})
}

// TestSheetSettingsTradingPercentFields checks that the buy and sell percent fields show the settings, write their own
// setting when edited and tell the open sheets of it, and that sync() shows replaced settings without telling the
// sheets anything, since the change that replaced them does that.
func TestSheetSettingsTradingPercentFields(t *testing.T) {
	c := check.New(t)
	owner := newEntityPanelWithFlaggedSettings()
	d, recorder := newTestSheetSettingsDockable(t, owner)
	s := owner.entity.SheetSettings
	sellBefore := s.SellPercent
	c.Equal(s.BuyPercent, d.buyPercentField.CurrentValue(), "the buy percent field shows the setting")
	c.Equal(s.SellPercent, d.sellPercentField.CurrentValue(), "the sell percent field shows the setting")

	d.buyPercentField.SetText("75")
	c.Equal(75, s.BuyPercent, "the buy percent field writes the buy percent")
	c.Equal(sellBefore, s.SellPercent, "the buy percent field leaves the sell percent alone")
	c.Equal([]bool{false}, recorder.updates, "a new buy percent is reported, without a full rebuild")

	d.sellPercentField.SetText("40")
	c.Equal(40, s.SellPercent, "the sell percent field writes the sell percent")
	c.Equal(75, s.BuyPercent, "the sell percent field leaves the buy percent alone")
	c.Equal([]bool{false, false}, recorder.updates, "a new sell percent is reported, without a full rebuild")

	replacement := s.Clone(owner.entity)
	replacement.BuyPercent = 120
	replacement.SellPercent = 30
	owner.entity.SheetSettings = replacement
	replacement.SetOwningEntity(owner.entity)
	d.sync()
	c.Equal(120, d.buyPercentField.CurrentValue(), "sync shows the replacement's buy percent")
	c.Equal(30, d.sellPercentField.CurrentValue(), "sync shows the replacement's sell percent")
	c.Equal(120, replacement.BuyPercent, "sync leaves the buy percent as it was")
	c.Equal(30, replacement.SellPercent, "sync leaves the sell percent as it was")
	c.Equal(2, len(recorder.updates), "sync tells the sheets nothing")
}

// sellMoneyNames returns the names of the settings' sell money, in the order held.
func sellMoneyNames(s *gurps.SheetSettings) []string {
	names := make([]string, 0, len(s.SellMoney))
	for _, eqp := range s.SellMoney {
		names = append(names, eqp.Name)
	}
	return names
}

// TestSheetSettingsSellMoneyDrop checks that dropping equipment onto the sell money list adds a copy of each item that
// can be money, once, wherever it comes from, and tells the open sheets, while containers and equipment that isn't
// tagged as money are refused, with a drag of nothing but those declined outright; that each entry's button removes it;
// that sync() shows the list of replaced settings; and that an item copied into a library from another, and so
// carrying that other library's source, is added as the item dragged, with its own name and value, keeping the source
// it carries (see gurps.Reference).
func TestSheetSettingsSellMoneyDrop(t *testing.T) {
	c := check.New(t)
	money, library := newTradingSettingsTest(t, c)
	owner := newEntityPanelWithFlaggedSettings()
	d, recorder := newTestSheetSettingsDockable(t, owner)
	s := owner.entity.SheetSettings
	hint := []string{"Drag equipment tagged \"Money\" from any equipment list here."}
	c.Equal(hint, uxtest.LabelTexts(d.sellMoneyPanel), "an empty list explains how to add to it")

	other := &fakeDragInfo{types: []string{traitDragKey.UTI}}
	c.False(d.sellMoneyPanel.CanAcceptDropCallback(other), "the list takes nothing but equipment")
	c.Equal(drag.None, d.sellMoneyPanel.DragUpdatedCallback(other, geom.Point{}, 0), "nor offers to")

	bead := newTradeTestItem(nil, "Trade Bead", "2", fxp.One)
	beadPath := filepath.Join(filepath.Dir(money.path), "Beads"+gurps.EquipmentExt)
	c.NoError(gurps.SaveEquipment([]*gurps.Equipment{bead}, beadPath))
	beads := NewEquipmentTableDockable(beadPath, []*gurps.Equipment{bead})
	c.False(dropOnSellMoney(t, c, d, beads.table, bead), "equipment that isn't tagged as money is refused")
	c.Equal(0, len(s.SellMoney))

	c.True(dropOnSellMoney(t, c, d, library.table, money.eagle), "a money item from a library is added")
	c.Equal([]string{"Silver Eagle"}, sellMoneyNames(s))
	c.Equal([]bool{false}, recorder.updates, "the addition is reported, without a full rebuild")
	c.Equal([]string{"Silver Eagle", "$20"}, uxtest.LabelTexts(d.sellMoneyPanel), "the list shows the item and its value")
	if len(s.SellMoney) == 1 {
		eagle := s.SellMoney[0]
		c.True(eagle != money.eagle, "a copy is kept, not the library's item")
		c.NotEqual(money.eagle.TID, eagle.TID)
		c.Equal(money.source(money.eagle), eagle.Source, "the copy points back at the library's item")
		c.Nil(eagle.DataOwner(), "the copy has no owner")
	}

	c.False(dropOnSellMoney(t, c, d, library.table, money.eagle), "an item already listed is not added again")
	c.False(dropOnSellMoney(t, c, d, library.table, money.purse), "a container is refused")
	c.Equal(1, len(s.SellMoney), "neither changes the list")
	c.Equal(1, len(recorder.updates), "nor is reported")

	c.True(dropOnSellMoney(t, c, d, library.table, money.purse, money.penny),
		"of several rows, those that can be money are added")
	c.Equal([]string{"Silver Eagle", "Copper Penny"}, sellMoneyNames(s),
		"the money item is added after the one already there")
	c.Equal([]string{"Copper Penny", "$1", "Silver Eagle", "$20"}, uxtest.LabelTexts(d.sellMoneyPanel),
		"the list shows the items by increasing value")

	elsewhere := newTradeTestItem(nil, "Gold Crown", "100", fxp.One, gurps.MoneyTag)
	outside := NewEquipmentTableDockable(filepath.Join(t.TempDir(), "Coins"+gurps.EquipmentExt),
		[]*gurps.Equipment{elsewhere})
	c.True(dropOnSellMoney(t, c, d, outside.table, elsewhere), "equipment from outside the libraries is added too")
	c.Equal([]string{"Silver Eagle", "Copper Penny", "Gold Crown"}, sellMoneyNames(s))
	if len(s.SellMoney) == 3 {
		c.True(s.SellMoney[2].Source.IsZero(), "a copy of an item from no library, with no Source of its own, has none")
	}
	c.Equal([]string{"Copper Penny", "$1", "Silver Eagle", "$20", "Gold Crown", "$100"},
		uxtest.LabelTexts(d.sellMoneyPanel))

	sellMoneyRemoveButton(t, d, "Silver Eagle").ClickCallback()
	c.Equal([]string{"Copper Penny", "Gold Crown"}, sellMoneyNames(s), "the button removes its own entry")
	c.Equal(4, len(recorder.updates), "the removal is reported")
	c.Equal([]string{"Copper Penny", "$1", "Gold Crown", "$100"}, uxtest.LabelTexts(d.sellMoneyPanel),
		"the list no longer shows it")

	d.reset()
	c.Equal(0, len(owner.entity.SheetSettings.SellMoney), "reset restores the defaults, which have no sell money")
	c.Equal(hint, uxtest.LabelTexts(d.sellMoneyPanel), "sync shows the list of the reset settings")

	user := gurps.GlobalSettings().Libraries.User()
	localPath := filepath.Join(user.Path(false), "Local Coins"+gurps.EquipmentExt)
	local := money.eagle.Clone(money.libFile, nil, nil, gurps.Reference)
	local.Name = "Local Eagle"
	local.BaseValue = "25"
	c.Equal(money.source(money.eagle), local.Source, "the copy keeps the source it was copied from")
	c.NoError(gurps.SaveEquipment([]*gurps.Equipment{local}, localPath))
	locals := NewEquipmentTableDockable(localPath, []*gurps.Equipment{local})
	c.True(dropOnSellMoney(t, c, d, locals.table, local), "the copied item is added")
	s = owner.entity.SheetSettings
	c.Equal([]string{"Local Eagle"}, sellMoneyNames(s),
		"the item dragged is what is copied, not the item it was copied from")
	if len(s.SellMoney) == 1 {
		c.Equal("25", s.SellMoney[0].BaseValue)
		c.NotEqual(local.TID, s.SellMoney[0].TID)
		c.Equal(money.source(money.eagle), s.SellMoney[0].Source, "the copy keeps the source the item carries")
	}
	c.Equal([]string{"Local Eagle", "$25"}, uxtest.LabelTexts(d.sellMoneyPanel), "the list shows the item dragged")
}

// TestSheetSettingsSellMoneyFollowsTheValueFormat checks that the sell money list shows the values in the equipment
// value format of the settings and follows a change to that format made in the same dockable.
func TestSheetSettingsSellMoneyFollowsTheValueFormat(t *testing.T) {
	c := check.New(t)
	money, _ := newTradingSettingsTest(t, c)
	owner := newEntityPanelWithFlaggedSettings()
	d, _ := newTestSheetSettingsDockable(t, owner)
	s := owner.entity.SheetSettings
	s.EquipmentValueFormat = fxp.NumberFormat{}
	halfCrown := newTradeTestItem(nil, "Half Crown", "2.5", fxp.One, gurps.MoneyTag)
	path := filepath.Join(filepath.Dir(money.path), "Crowns"+gurps.EquipmentExt)
	c.NoError(gurps.SaveEquipment([]*gurps.Equipment{halfCrown}, path))
	crowns := NewEquipmentTableDockable(path, []*gurps.Equipment{halfCrown})
	c.True(dropOnSellMoney(t, c, d, crowns.table, halfCrown))
	c.Equal([]string{"Half Crown", "$2.5"}, uxtest.LabelTexts(d.sellMoneyPanel),
		"the value is shown with as many decimal places as it has")

	row := d.numberFormats[3]
	row.popup.Select(fxp.ZeroPlaces)
	c.Equal(fxp.ZeroPlaces, s.EquipmentValueFormat.Places, "the equipment value popup must be the fourth row")
	c.Equal([]string{"Half Crown", "$3"}, uxtest.LabelTexts(d.sellMoneyPanel),
		"the list follows a change to the decimal places")

	row.popup.Select(fxp.TwoPlaces)
	row.pad.State = uncheck.On
	row.pad.ClickCallback()
	c.Equal([]string{"Half Crown", "$2.50"}, uxtest.LabelTexts(d.sellMoneyPanel),
		"the list follows a change to the padding")
}

// TestSheetSettingsSellMoneyWithoutOwner checks that the defaults' sell money list, which no sheet owns, takes drops
// and removals too.
func TestSheetSettingsSellMoneyWithoutOwner(t *testing.T) {
	c := check.New(t)
	money, library := newTradingSettingsTest(t, c)
	global := gurps.GlobalSettings()
	t.Cleanup(gurps.SyncGlobalSheetSettings) // Runs after the swap below is undone.
	uxtest.SwapForTest(t, &global.Sheet, global.Sheet.Clone(nil))
	d, recorder := newTestSheetSettingsDockable(t, nil)

	c.True(dropOnSellMoney(t, c, d, library.table, money.penny), "a money item from a library is added")
	c.Equal([]string{"Copper Penny"}, sellMoneyNames(global.Sheet), "the defaults hold a copy of it")
	c.Equal([]*gurps.Entity{nil}, recorder.entities, "the change is reported as one to the defaults")
	c.Equal([]string{"Copper Penny", "$1"}, uxtest.LabelTexts(d.sellMoneyPanel))

	sellMoneyRemoveButton(t, d, "Copper Penny").ClickCallback()
	c.Equal(0, len(global.Sheet.SellMoney), "the button removes the entry from the defaults")
	c.Equal(2, len(recorder.updates), "the removal is reported")
}

// TestSheetSettingsSellMoneyDropFromASheet checks that equipment dragged from a sheet's carried equipment list, which
// no library holds, is added as a copy owned by no one with a quantity of one, leaving the item carried as it was.
func TestSheetSettingsSellMoneyDropFromASheet(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &flushDragFeedback, func(*unison.Panel) {})
	var bar *gurps.Equipment
	sheet := openSellTestSheet(t, screen, nil, func(entity *gurps.Entity) []*gurps.Equipment {
		bar = newTradeTestItem(entity, "Trade Bar", "50", fxp.Five, gurps.MoneyTag)
		return []*gurps.Equipment{bar}
	})
	d := openSheetSettingsFor(t, screen, sheet)
	// The row is looked up here and the test failed on this goroutine, as testing requires, rather than by
	// dropOnSellMoney on the UI goroutine, which would leave the headless screen without its event loop.
	var rows []*Node[*gurps.Equipment]
	screen.Do(func() {
		for _, row := range sheet.CarriedEquipment.Table.RootRows() {
			if row.Data() == bar {
				rows = append(rows, row)
			}
		}
	})
	if len(rows) != 1 {
		t.Fatal("the sheet's carried equipment list must have a row for the bar")
	}
	var dropped bool
	var names, texts []string
	var entry *gurps.Equipment
	screen.Do(func() {
		dropped = dropRowsOnSellMoney(t, c, d, sheet.CarriedEquipment.Table, rows)
		s := sheet.Entity().SheetSettings
		names = sellMoneyNames(s)
		if len(s.SellMoney) == 1 {
			entry = s.SellMoney[0]
		}
		texts = uxtest.LabelTexts(d.sellMoneyPanel)
	})
	c.True(dropped, "money carried on a sheet can be dropped on its own settings")
	c.Equal([]string{"Trade Bar"}, names)
	if entry != nil {
		c.True(entry != bar, "a copy is kept, not the carried item")
		c.NotEqual(bar.TID, entry.TID)
		c.Equal(fxp.One, entry.Quantity, "the copy holds one, however many are carried")
		c.Nil(entry.DataOwner(), "the copy has no owner")
		c.True(entry.Source.IsZero(), "a copy of an item from no library, with no Source of its own, has none")
	}
	c.Equal(fxp.Five, bar.Quantity, "the carried item is untouched")
	c.Equal([]string{"Trade Bar", "$50.00"}, texts, "the list shows the copy, formatted as the sheet formats values")
}

// openSheetSettingsFor opens the sheet's settings in the headless workspace and returns their dockable, failing the test
// if it can't be found.
func openSheetSettingsFor(t *testing.T, screen *unison.HeadlessScreen, sheet *Sheet) *sheetSettingsDockable {
	t.Helper()
	var d *sheetSettingsDockable
	screen.Do(func() {
		ShowSheetSettings(sheet)
		for _, one := range AllDockables() {
			if s, ok := one.AsPanel().Self.(*sheetSettingsDockable); ok && s.owner == sheet {
				d = s
			}
		}
	})
	if d == nil {
		t.Fatal("the sheet's settings must be open")
	}
	return d
}

// TestSheetSettingsSellMoneyBanding checks that the sell money list of a sheet whose settings show two decimal places
// shows the values so, by increasing value, and that it draws the rows on the background a table gives its rows, with
// every other row banded.
func TestSheetSettingsSellMoneyBanding(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	money := newTradeTestMoney(t, c)
	sheet := openSellTestSheet(t, screen, []*gurps.Equipment{money.eagle, money.penny},
		func(_ *gurps.Entity) []*gurps.Equipment { return nil })
	d := openSheetSettingsFor(t, screen, sheet)
	var texts []string
	var rows int
	var samples []color.NRGBA
	screen.Do(func() {
		p := d.sellMoneyPanel
		texts = uxtest.LabelTexts(p)
		rows = len(d.sellMoneyRows)
		d.ValidateLayout()
		area := p.ContentRect(true)
		img, err := unison.NewImageFromDrawing(int(area.Width), int(area.Height), 72, func(gc *unison.Canvas) {
			p.DrawCallback(gc, area)
		})
		if err != nil {
			t.Error(err)
			return
		}
		defer img.Dispose()
		pixels, err := img.ToNRGBA()
		if err != nil {
			t.Error(err)
			return
		}
		for _, row := range d.sellMoneyRows {
			center := listRowBounds(row, area).Center()
			samples = append(samples, pixels.NRGBAAt(int(center.X), int(center.Y)))
		}
	})
	c.Equal([]string{"Copper Penny", "$1.00", "Silver Eagle", "$20.00"}, texts,
		"the list shows the items by increasing value, formatted as the sheet formats equipment values")
	c.Equal(2, rows, "the list keeps the widgets of each row")
	nrgba := func(ink unison.ColorProvider) color.NRGBA {
		clr := ink.GetColor()
		return color.NRGBA{R: uint8(clr.Red()), G: uint8(clr.Green()), B: uint8(clr.Blue()), A: 255}
	}
	c.Equal([]color.NRGBA{nrgba(unison.ThemeBelowSurface), nrgba(unison.ThemeBanding)}, samples,
		"the first row sits on the background a table gives its rows and the second is banded")
}

// TestSheetSettingsSellMoneyRemoveKeepsTheFocus checks that removing a sell money entry with its button while the button
// holds the keyboard focus moves the focus to the button that takes its place, or to the sell percent field once no
// entry is left, rather than leaving it on the button the rebuilt list no longer holds.
func TestSheetSettingsSellMoneyRemoveKeepsTheFocus(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	money := newTradeTestMoney(t, c)
	sheet := openSellTestSheet(t, screen, []*gurps.Equipment{money.eagle, money.penny},
		func(_ *gurps.Entity) []*gurps.Equipment { return nil })
	d := openSheetSettingsFor(t, screen, sheet)
	removeButton := func(name string) *unison.Button {
		var button *unison.Button
		screen.Do(func() {
			for _, b := range uxtest.PanelsOfType[*unison.Button](d.sellMoneyPanel) {
				if b.Accessibility.Name == "Remove "+name {
					button = b
				}
			}
		})
		return button
	}
	press := func(name string) {
		t.Helper()
		button := removeButton(name)
		if button == nil {
			t.Fatalf("the sell money list has no button to remove %s", name)
		}
		screen.Do(func() {
			button.RequestFocus()
			button.Click()
		})
	}

	press("Silver Eagle")
	var focus *unison.Panel
	screen.Do(func() { focus = wnd.CurrentFocus() })
	c.Equal([]string{"Copper Penny", "$1.00"}, uxtest.LabelTexts(d.sellMoneyPanel), "the entry is removed")
	c.True(focus != nil && focus == removeButton("Copper Penny").AsPanel(),
		"the focus must move to the button that takes the removed one's place")

	press("Copper Penny")
	screen.Do(func() { focus = wnd.CurrentFocus() })
	c.True(focus == d.sellPercentField.AsPanel(), "the focus must move to the sell percent field once no entry is left")
}

// TestSheetSettingsPageRefsAreAllLinks verifies that each sheet setting citing a page shares a wrapper with a
// focusable "(ref)" label whose one link is the whole reference and is followed by Space, and that a setting without a
// page reference adds its checkbox alone. LinkPageRefs recognizes only Basic Set references, so this covers those to
// other books, such as "PY65:30".
func TestSheetSettingsPageRefsAreAllLinks(t *testing.T) {
	c := check.New(t)
	d := &sheetSettingsDockable{}
	var refs []string
	for _, option := range sheetOptions() {
		panel := unison.NewPanel()
		box := d.addCheckBox(panel, option.title, option.pageRef, 1, false, func(bool) {})
		c.Equal(option.title, box.Text.String())
		if option.pageRef == "" {
			c.Equal([]*unison.Panel{box.AsPanel()}, panel.Children(),
				"%q has no page reference, so its checkbox stands alone", option.title)
			continue
		}
		refs = append(refs, option.pageRef)
		if len(panel.Children()) != 1 {
			t.Fatalf("the checkbox of %q and its page reference must share a wrapper, but the panel holds %d children",
				option.title, len(panel.Children()))
		}
		children := panel.Children()[0].Children()
		if len(children) != 2 {
			t.Fatalf("the wrapper of %q must hold the checkbox and its page reference, but holds %d children",
				option.title, len(children))
		}
		c.Equal(box.AsPanel(), children[0])
		label, ok := children[1].Self.(*TextLabel)
		if !ok {
			t.Fatalf("the page reference of %q must be a label holding the link, not a %T", option.title,
				children[1].Self)
		}
		c.Equal("("+option.pageRef+")", label.String())
		c.True(label.Focusable(), "the link is reached by the keyboard focus, screen reader or not")
		_, pref, _ := label.Sizes(geom.Size{})
		label.SetFrameRect(geom.NewRect(0, 0, pref.Width, pref.Height))
		links := label.Links()
		if len(links) != 1 {
			t.Fatalf("the page reference of %q must hold one link, but holds %d", option.title, len(links))
		}
		c.Equal(option.pageRef, links[0].Ref, "the link of %q is the whole of its page reference", option.title)
		var followed []string
		label.LinkHandler = func(ref string) { followed = append(followed, ref) }
		c.True(label.keyDown(unison.KeySpace, 0, false))
		c.Equal([]string{option.pageRef}, followed, "the link of %q opens its page", option.title)
	}
	c.Equal([]string{"PY120:7", "P102", "PY65:30", "B269"}, refs, "the sheet settings that cite a page")
}
