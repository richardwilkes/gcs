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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/progression"
	"github.com/richardwilkes/gcs/v5/model/jio"
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
			c.NotNil(s.Layout)
			c.NotNil(s.Attributes)
			c.NotNil(s.BodyType)
		})
	}
}

// TestNewSheetSettingsFromFileLegacyFields verifies that the legacy field names handled by
// (*SheetSettings).UnmarshalJSONFrom are honored regardless of which location the settings were found in.
func TestNewSheetSettingsFromFileLegacyFields(t *testing.T) {
	const body = `{"name":"Test Body","roll":"3d"}`
	const fields = `"show_advantage_modifier_adj":true,"hit_locations":` + body +
		`,"block_layout":["notes","advantages skills"]`
	for name, content := range map[string]string{
		"current": `{` + fields + `}`,
		"old":     `{"sheet_settings":{` + fields + `}}`,
	} {
		t.Run(name, func(t *testing.T) {
			c := check.New(t)
			s, err := NewSheetSettingsFromFile(nil, writeSheetSettingsFile(t, content))
			c.NoError(err)
			c.True(s.ShowTraitModifierAdj)
			c.NotNil(s.BodyType)
			c.Equal("Test Body", s.BodyType.Name)
			c.NotNil(s.Layout)
			// The legacy block layout becomes the layout tree, with "advantages" mapped onto "traits" and the blocks
			// it didn't mention appended.
			c.Equal([][]string{
				{BlockNotesKey},
				{BlockTraitsKey, BlockSkillsKey},
				{BlockReactionsKey},
				{BlockConditionalModifiersKey},
				{BlockMeleeKey},
				{BlockRangedKey},
				{BlockSpellsKey},
				{BlockEquipmentKey},
				{BlockOtherEquipmentKey},
			}, s.Layout.ListBands())
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

// TestSheetSettingsNumberFormatsRoundTrip verifies that the display formats survive a save and load, that an unknown
// decimal places key falls back to As Needed, and that settings which have never expressed a display preference are
// written exactly as they were before the formats existed, so existing files and their content hashes are unaffected.
func TestSheetSettingsNumberFormatsRoundTrip(t *testing.T) {
	c := check.New(t)
	s := FactorySheetSettings()
	s.HeightFormat = fxp.NumberFormat{Places: fxp.ZeroPlaces}
	s.BodyWeightFormat = fxp.NumberFormat{Places: fxp.OnePlace, PadWithZeros: true}
	s.EquipmentWeightFormat = fxp.NumberFormat{Places: fxp.TwoPlaces, PadWithZeros: true}
	s.EquipmentValueFormat = fxp.NumberFormat{PadWithZeros: true}
	p := filepath.Join(t.TempDir(), "settings.sheet")
	c.NoError(s.Save(p))
	loaded, err := NewSheetSettingsFromFile(nil, p)
	c.NoError(err)
	c.Equal(s.HeightFormat, loaded.HeightFormat)
	c.Equal(s.BodyWeightFormat, loaded.BodyWeightFormat)
	c.Equal(s.EquipmentWeightFormat, loaded.EquipmentWeightFormat)
	c.Equal(s.EquipmentValueFormat, loaded.EquipmentValueFormat)

	loaded, err = NewSheetSettingsFromFile(nil, writeSheetSettingsFile(t,
		`{"height_format":{"decimal_places":"bogus","pad_with_zeros":true},"equipment_value_format":{"decimal_places":"2"}}`))
	c.NoError(err)
	c.Equal(fxp.NumberFormat{Places: fxp.AsNeeded, PadWithZeros: true}, loaded.HeightFormat)
	c.Equal(fxp.NumberFormat{Places: fxp.TwoPlaces}, loaded.EquipmentValueFormat)
	c.Equal(fxp.NumberFormat{}, loaded.BodyWeightFormat)

	before, err := jio.Marshal(FactorySheetSettings())
	c.NoError(err)
	c.False(strings.Contains(string(before), "_format"), "unset display formats must not be written: %s", before)
}

// TestSheetSettingsFormatHelpers verifies that the display helpers combine the sheet's units with the matching display
// format, and that the zero-value formats render exactly as the plain unit formatting does.
func TestSheetSettingsFormatHelpers(t *testing.T) {
	c := check.New(t)
	s := FactorySheetSettings()
	height := fxp.Length(fxp.FromStringForced("68.98"))
	weight := fxp.Weight(fxp.FromStringForced("159.0847"))
	value := fxp.FromStringForced("1234.5678")
	c.Equal(`5'8.98"`, s.FormatHeight(height))
	c.Equal("159.0847 lb", s.FormatBodyWeight(weight))
	c.Equal("159.0847 lb", s.FormatEquipmentWeight(weight))
	c.Equal("1,234.5678", s.FormatEquipmentValue(value))

	s.DefaultLengthUnits = fxp.Centimeter
	s.DefaultWeightUnits = fxp.Kilogram
	s.HeightFormat = fxp.NumberFormat{Places: fxp.ZeroPlaces}
	s.BodyWeightFormat = fxp.NumberFormat{Places: fxp.OnePlace, PadWithZeros: true}
	s.EquipmentWeightFormat = fxp.NumberFormat{Places: fxp.TwoPlaces}
	s.EquipmentValueFormat = fxp.NumberFormat{Places: fxp.OnePlace}
	c.Equal("172 cm", s.FormatHeight(height))
	c.Equal("79.5 kg", s.FormatBodyWeight(weight))
	c.Equal("79.54 kg", s.FormatEquipmentWeight(weight))
	c.Equal("1,234.6", s.FormatEquipmentValue(value))
	c.Equal("80.0 kg", s.FormatBodyWeight(fxp.Weight(fxp.FromInteger(160))))
}

// TestSheetSettingsEnsureValidityRepairsNumberFormats verifies that EnsureValidity replaces an out-of-range decimal
// places choice in each of the four display formats with As Needed, leaving the padding flag alone. An unknown key in a
// file is already repaired by the enum's UnmarshalText on the way in, so this has to assign the bad values directly.
func TestSheetSettingsEnsureValidityRepairsNumberFormats(t *testing.T) {
	c := check.New(t)
	s := FactorySheetSettings()
	bad := fxp.NumberFormat{Places: fxp.DecimalPlace(200), PadWithZeros: true}
	s.HeightFormat = bad
	s.BodyWeightFormat = bad
	s.EquipmentWeightFormat = bad
	s.EquipmentValueFormat = bad
	s.EnsureValidity()
	repaired := fxp.NumberFormat{Places: fxp.AsNeeded, PadWithZeros: true}
	c.Equal(repaired, s.HeightFormat)
	c.Equal(repaired, s.BodyWeightFormat)
	c.Equal(repaired, s.EquipmentWeightFormat)
	c.Equal(repaired, s.EquipmentValueFormat)
}
