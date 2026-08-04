// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/progression"
	"github.com/richardwilkes/toolbox/v2/check"
)

// writeSheetSettingsFile writes the given JSON to a temporary file and returns its path.
func writeSheetSettingsFile(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "settings.sheet")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestNewSheetSettingsFromFileCurrentLocation verifies that a file with the settings at the top level loads.
func TestNewSheetSettingsFromFileCurrentLocation(t *testing.T) {
	c := check.New(t)
	s, err := NewSheetSettingsFromFile(nil, writeSheetSettingsFile(t,
		`{"damage_progression":"knowing_your_own_strength","default_length_units":"in",`+
			`"use_multiplicative_modifiers":true,"notes_display":"tooltip"}`))
	c.NoError(err)
	c.Equal(progression.KnowingYourOwnStrength, s.DamageProgression)
	c.Equal(fxp.Inch, s.DefaultLengthUnits)
	c.True(s.UseMultiplicativeModifiers)
	c.Equal(display.Tooltip, s.NotesDisplay)
}

// TestNewSheetSettingsFromFileOldLocation verifies that a legacy file, which nests the settings under a
// "sheet_settings" key, still loads. Embedding SheetSettings in the wrapper struct used to decode the file promoted its
// custom unmarshaler onto the wrapper, so the whole document was consumed by it and the legacy key was never seen,
// silently yielding factory defaults (which the --convert path would then write back over the original file).
func TestNewSheetSettingsFromFileOldLocation(t *testing.T) {
	c := check.New(t)
	s, err := NewSheetSettingsFromFile(nil, writeSheetSettingsFile(t,
		`{"type":"settings","version":5,"sheet_settings":{"damage_progression":"knowing_your_own_strength",`+
			`"default_length_units":"in","use_multiplicative_modifiers":true,"notes_display":"tooltip"}}`))
	c.NoError(err)
	c.Equal(progression.KnowingYourOwnStrength, s.DamageProgression)
	c.Equal(fxp.Inch, s.DefaultLengthUnits)
	c.True(s.UseMultiplicativeModifiers)
	c.Equal(display.Tooltip, s.NotesDisplay)
}

// TestNewSheetSettingsFromFileAppliesEnsureValidity verifies that a mostly empty file is filled in with the factory
// defaults for the pieces that must always be present.
func TestNewSheetSettingsFromFileAppliesEnsureValidity(t *testing.T) {
	for name, content := range map[string]string{
		"current": `{}`,
		"old":     `{"sheet_settings":{}}`,
	} {
		t.Run(name, func(t *testing.T) {
			c := check.New(t)
			s, err := NewSheetSettingsFromFile(nil, writeSheetSettingsFile(t, content))
			c.NoError(err)
			c.NotNil(s.Page)
			c.NotNil(s.BlockLayout)
			c.NotNil(s.Attributes)
			c.NotNil(s.BodyType)
		})
	}
}

// TestNewSheetSettingsFromFileLegacyFields verifies that the legacy field names handled by
// (*SheetSettings).UnmarshalJSONFrom are honored regardless of which location the settings were found in.
func TestNewSheetSettingsFromFileLegacyFields(t *testing.T) {
	const body = `{"name":"Test Body","roll":"3d"}`
	for name, content := range map[string]string{
		"current": `{"show_advantage_modifier_adj":true,"hit_locations":` + body + `}`,
		"old":     `{"sheet_settings":{"show_advantage_modifier_adj":true,"hit_locations":` + body + `}}`,
	} {
		t.Run(name, func(t *testing.T) {
			c := check.New(t)
			s, err := NewSheetSettingsFromFile(nil, writeSheetSettingsFile(t, content))
			c.NoError(err)
			c.True(s.ShowTraitModifierAdj)
			c.NotNil(s.BodyType)
			c.Equal("Test Body", s.BodyType.Name)
		})
	}
}

// TestNewSheetSettingsFromFileErrors verifies that unusable files are reported as errors rather than quietly becoming
// factory defaults.
func TestNewSheetSettingsFromFileErrors(t *testing.T) {
	c := check.New(t)
	_, err := NewSheetSettingsFromFile(nil, writeSheetSettingsFile(t, `{"damage_progression":`))
	c.HasError(err)
	_, err = NewSheetSettingsFromFile(nil, filepath.Join(t.TempDir(), "does_not_exist.sheet"))
	c.HasError(err)
}
