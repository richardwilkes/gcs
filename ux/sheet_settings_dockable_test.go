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
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
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
	entities []*gurps.Entity // one entry per notification, holding the entity whose settings changed
	updates  []bool          // one entry per notification, holding the fullRebuild flag
}

func (r *sheetSettingsRecorder) TitleIcon(_ geom.Size) unison.Drawable { return nil }
func (r *sheetSettingsRecorder) Title() string                         { return "recorder" }
func (r *sheetSettingsRecorder) Tooltip() string                       { return "" }
func (r *sheetSettingsRecorder) Modified() bool                        { return false }

func (r *sheetSettingsRecorder) SheetSettingsUpdated(entity *gurps.Entity, fullRebuild bool) {
	r.entities = append(r.entities, entity)
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
	swapForTest(t, &Workspace.DocumentDock, NewDocumentDock())
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
// four blocks that build these widgets are near-identical, so a copy-paste slip between them is the most likely error,
// and nothing less specific than this would notice one.
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
