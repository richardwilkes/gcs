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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
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
	updates []bool // one entry per notification, holding the blockLayout (full rebuild) flag
}

func (r *sheetSettingsRecorder) TitleIcon(_ geom.Size) unison.Drawable { return nil }
func (r *sheetSettingsRecorder) Title() string                         { return "recorder" }
func (r *sheetSettingsRecorder) Tooltip() string                       { return "" }
func (r *sheetSettingsRecorder) Modified() bool                        { return false }

func (r *sheetSettingsRecorder) SheetSettingsUpdated(_ *gurps.Entity, blockLayout bool) {
	r.updates = append(r.updates, blockLayout)
}

func (r *sheetSettingsRecorder) sawFullRebuild() bool {
	for _, full := range r.updates {
		if full {
			return true
		}
	}
	return false
}

// entityPanelForTest is a minimal EntityPanel for a sheet settings dockable to own.
type entityPanelForTest struct {
	unison.Panel
	entity *gurps.Entity
}

func (p *entityPanelForTest) Entity() *gurps.Entity { return p.entity }

// newTestSheetSettingsDockable returns a dockable whose widgets have been built from the given owner's settings, along
// with a recorder docked where syncSheet will find it. The document dock is restored when the test finishes.
func newTestSheetSettingsDockable(t *testing.T, owner EntityPanel) (*sheetSettingsDockable, *sheetSettingsRecorder) {
	t.Helper()
	saved := Workspace.DocumentDock
	t.Cleanup(func() { Workspace.DocumentDock = saved })
	Workspace.DocumentDock = NewDocumentDock()
	recorder := &sheetSettingsRecorder{}
	recorder.Self = recorder
	Workspace.DocumentDock.DockTo(recorder, nil, side.Left)
	d := &sheetSettingsDockable{owner: owner}
	d.Self = d
	d.initContent(unison.NewPanel())
	return d, recorder
}

// newEntityPanelWithFlaggedSettings returns an EntityPanel whose sheet settings differ from the global defaults only in
// the boolean options that the dockable presents as checkboxes. Every popup and text field therefore still matches, so
// nothing in sync() will fire a change callback of its own.
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
	s.UseMultiplicativeModifiers = !s.UseMultiplicativeModifiers
	s.UseHalfStatDefaults = !s.UseHalfStatDefaults
	s.UseModifyingDicePlusAdds = !s.UseModifyingDicePlusAdds
	s.ExcludeUnspentPointsFromTotal = !s.ExcludeUnspentPointsFromTotal
}

// TestSheetSettingsResetNotifiesSheets verifies that resetting the sheet settings tells the open sheets to rebuild.
// sync() only pushes the new values into this dockable's own widgets, and assigning a CheckBox's State does not fire
// its ClickCallback, so settings that differ only in the checkbox-backed options used to leave the open sheet showing
// the old columns and values until some unrelated edit happened to trigger a rebuild.
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

// TestSheetSettingsLoadNotifiesSheets verifies the same for importing a sheet settings file: a file differing only in
// the checkbox-backed options must still refresh the open sheets.
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

// TestSheetSettingsSyncDoesNotFireCheckBoxCallbacks documents the reason the notification above has to be explicit:
// sync() assigns each CheckBox's State directly, which unison does not treat as a click, so the callbacks that would
// otherwise push the change out to the sheets never run.
func TestSheetSettingsSyncDoesNotFireCheckBoxCallbacks(t *testing.T) {
	c := check.New(t)
	owner := newEntityPanelWithFlaggedSettings()
	d, recorder := newTestSheetSettingsDockable(t, owner)

	// Swap in settings that differ only in the checkbox-backed options, then sync the widgets to them.
	replacement := gurps.GlobalSettings().Sheet.Clone(owner.entity)
	owner.entity.SheetSettings = replacement
	replacement.SetOwningEntity(owner.entity)
	d.sync()

	c.Equal(0, len(recorder.updates), "sync() alone must not be relied upon to notify the sheets")
}

// TestSheetSettingsTabTitle verifies that the character name is substituted into the tab title rather than being built
// into the string handed to i18n.Text, which would produce a per-character lookup key that no catalog entry can ever
// match.
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
