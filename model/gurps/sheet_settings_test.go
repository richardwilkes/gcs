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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/progression"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
)

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

// TestSheetSettingsEnsureValidityBoundsTradePercents verifies that EnsureValidity brings the buy and sell percentages
// into the range the sheet offers for them, as a file may hold any value, while a missing or non-positive one takes the
// default, so that nothing showing them has to clamp, and record, a value nobody edited.
func TestSheetSettingsEnsureValidityBoundsTradePercents(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	data := fmt.Appendf(nil, `{
	"version": %d,
	"profile": {"name": "Plain"},
	"settings": {"sell_percent": 2000, "buy_percent": 1}
}`, jio.CurrentDataVersion)
	c.NoError(os.WriteFile(filepath.Join(dir, "plain.gcs"), data, 0o600))
	e, err := NewEntityFromFile(os.DirFS(dir), "plain.gcs")
	c.NoError(err)
	c.Equal(MaxTradePercent, e.SheetSettings.SellPercent, "a sell percent above the range is brought down to it")
	c.Equal(MinTradePercent, e.SheetSettings.BuyPercent, "a positive buy percent below the range is brought up to it")

	s := FactorySheetSettings()
	s.BuyPercent = -5
	s.SellPercent = 0
	s.EnsureValidity()
	c.Equal(100, s.BuyPercent, "a buy percent that isn't positive takes the default")
	c.Equal(50, s.SellPercent, "a sell percent that isn't positive takes the default")
	s.BuyPercent = 250
	s.SellPercent = 35
	s.EnsureValidity()
	c.Equal(250, s.BuyPercent, "a buy percent within the range is left alone")
	c.Equal(35, s.SellPercent, "a sell percent within the range is left alone")
}

// TestSheetSettingsEnsureValidityDropsWhatCanNotBeSellMoney checks that sell money entries a file can hold but that
// can't be money (see SheetSettings.CanAddSellMoney) are dropped, since were one paid out, MoneyItems would never list
// what was received, so it couldn't be spent, and that entries repeating an ID are dropped too.
func TestSheetSettingsEnsureValidityDropsWhatCanNotBeSellMoney(t *testing.T) {
	c := check.New(t)
	s := FactorySheetSettings()
	penny := denomination("Copper Penny", "1")
	s.SellMoney = SellMoneyList{
		nil,
		withTags(newContainerItem("Purse", "2"), MoneyTag),
		newEquipmentItem("Gold Crown", "100", "0"),
		penny,
	}
	s.EnsureValidity()
	c.Equal(SellMoneyList{penny}, s.SellMoney)

	twin := denomination("Copper Penny", "1")
	twin.TID = penny.TID
	s.SellMoney = SellMoneyList{penny, twin, penny}
	s.EnsureValidity()
	c.Equal(SellMoneyList{penny}, s.SellMoney, "entries repeating an ID, which the list is keyed on, are dropped")
}

// newTestSellMoney returns a coin with a modifier and a weapon, the parts of equipment that are attached to it when
// loaded, tagged as money.
func newTestSellMoney() *Equipment {
	silver := withTags(newEquipmentItem("Silver Eagle", "20", "0.01 lb"), MoneyTag)
	silver.AddModifiers(NewEquipmentModifier(nil, nil, false))
	silver.Weapons = []*Weapon{NewWeapon(silver, true)}
	return silver
}

// TestSheetSettingsSellMoneyRoundTrip checks that the sell money is written without the derived "calc" values and read
// back as copies with their IDs and sources, no data owner, their weapons and modifiers attached, and the preconfigured
// mark and disabled state cleared; and that what a file can hold but can't be money is dropped without error: null,
// containers, equipment not tagged as money, and the entries of the unreleased format that named a library item by its
// source, which decode as equipment with nothing but an ID.
func TestSheetSettingsSellMoneyRoundTrip(t *testing.T) {
	c := check.New(t)
	libFile := LibraryFile{Library: "Test Library", Path: "Coins" + EquipmentExt}
	silver := newTestSellMoney()
	s := FactorySheetSettings()
	c.True(s.AddSellMoney(silver, libFile))
	entry := s.SellMoney[0]
	data, err := jio.Marshal(s)
	c.NoError(err)
	moneyData, err := jio.Marshal(s.SellMoney)
	c.NoError(err)
	c.False(strings.Contains(string(moneyData), `"calc"`), "the derived values aren't written")
	c.True(strings.Contains(string(data), `"sell_money":`+string(moneyData)), "the settings write the list the same way")
	var loaded SheetSettings
	c.NoError(jio.Unmarshal(data, &loaded))
	c.Equal(1, len(loaded.SellMoney))
	got := loaded.SellMoney[0]
	c.Equal(entry.TID, got.TID)
	c.Equal(Source{LibraryFile: libFile, TID: silver.TID}, got.Source)
	c.Equal("Silver Eagle", got.Name)
	c.Equal(fxp.FromInteger(20), got.AdjustedValue())
	c.Nil(got.DataOwner())
	c.Nil(got.Parent())
	c.Equal(1, len(got.Weapons))
	c.True(got.Weapons[0].Owner == WeaponOwner(got), "the weapon is attached to the copy")
	c.Equal(1, len(got.Modifiers))
	c.True(entry.MergeMatches(got), "the copy read back is the one written")

	penny := denomination("Copper Penny", "1")
	penny.Preconfigured = true
	penny.Disabled = true
	s.SellMoney = SellMoneyList{
		nil, withTags(newContainerItem("Purse", "2"), MoneyTag),
		newEquipmentItem("Gold Crown", "100", "0"), penny,
	}
	data, err = jio.Marshal(s)
	c.NoError(err)
	loaded = SheetSettings{}
	c.NoError(jio.Unmarshal(data, &loaded))
	c.Equal(1, len(loaded.SellMoney), "null, containers and items not tagged as money are dropped")
	c.Equal("Copper Penny", loaded.SellMoney[0].Name)
	c.False(loaded.SellMoney[0].Preconfigured, "the preconfigured mark is cleared")
	c.False(loaded.SellMoney[0].Disabled, "the disabled state is cleared")

	old := fmt.Sprintf(`{"sell_money":[{"library":"Lib","path":"Coins.eqp","id":%q}]}`, tid.MustNewTID(kinds.Equipment))
	loaded = SheetSettings{}
	c.NoError(jio.Unmarshal([]byte(old), &loaded), "an entry of the unreleased format must not fail the load")
	c.Equal(0, len(loaded.SellMoney), "it is dropped")

	empty, err := jio.Marshal(FactorySheetSettings())
	c.NoError(err)
	c.False(strings.Contains(string(empty), `"sell_money"`), "no sell money writes nothing")
}

// TestSheetSettingsSellMoneyPathsAvoidTheClosedStateLock checks that cloning, validating, writing and reading settings
// that hold sell money never consult the closed states the global settings hold: Settings.Save does all of that while
// holding their lock, and loading the global settings does it inside GlobalSettings' initializer, which cloning
// equipment would re-enter.
func TestSheetSettingsSellMoneyPathsAvoidTheClosedStateLock(t *testing.T) {
	c := check.New(t)
	GlobalSettings()
	s := FactorySheetSettings()
	c.True(s.AddSellMoney(newTestSellMoney(), LibraryFile{}))
	closedLock.Lock()
	defer closedLock.Unlock()
	done := make(chan error, 1)
	go func() {
		done <- func() error {
			clone := s.Clone(nil)
			clone.EnsureValidity()
			data, err := jio.Marshal(clone)
			if err != nil {
				return err
			}
			var loaded SheetSettings
			if err = jio.Unmarshal(data, &loaded); err != nil {
				return err
			}
			if len(loaded.SellMoney) != 1 {
				return fmt.Errorf("expected 1 sell money entry, got %d", len(loaded.SellMoney))
			}
			return nil
		}()
	}()
	select {
	case err := <-done:
		c.NoError(err)
	case <-time.After(10 * time.Second):
		t.Fatal("the sell money paths must not take the closed-state lock")
	}
}

// TestSellMoneyCopiesLeaveNoClosedStates checks that the copies made to check a drag, to list an entry and to pay out
// (see SheetSettings.CanAddSellMoney, SheetSettings.AddSellMoney and Entity.SellMoneyItems) record no closed states,
// which a clone of an item whose notes are collapsed records for its copy (see Equipment.Clone): none of them is shown
// with notes to collapse, so the states would only pile up in the settings.
func TestSellMoneyCopiesLeaveNoClosedStates(t *testing.T) {
	c := check.New(t)
	settings := GlobalSettings()
	coin := denomination("Copper Penny", "1")
	coin.AddModifiers(NewEquipmentModifier(nil, nil, false))
	keys := []string{"N:" + string(coin.TID), "N:" + string(coin.Modifiers[0].TID)}
	for _, key := range keys {
		SetClosedState(key, true)
	}
	t.Cleanup(func() {
		for _, key := range keys {
			SetClosedState(key, false)
		}
	})
	count := func() int {
		closedLock.Lock()
		defer closedLock.Unlock()
		return len(settings.Closed)
	}
	before := count()
	s := FactorySheetSettings()
	c.True(s.CanAddSellMoney(coin))
	c.Equal(before, count(), "checking an item records nothing")
	c.True(s.AddSellMoney(coin, LibraryFile{}))
	c.Equal(before, count(), "listing an item records nothing")
	c.False(s.AddSellMoney(coin, LibraryFile{}))
	c.Equal(before, count(), "refusing an item records nothing")
	e := NewEntity()
	e.SheetSettings.SellMoney = SellMoneyList{coin}
	c.Equal(1, len(e.SellMoneyItems()))
	c.Equal(before, count(), "paying out records nothing")
}

// TestSheetSettingsCloneTakesTheEntity checks that a clone made for an entity belongs to it, and one made for none to
// none, whatever the settings cloned belonged to.
func TestSheetSettingsCloneTakesTheEntity(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	s := FactorySheetSettings()
	c.True(s.Clone(e).Entity == e, "a clone made for an entity belongs to it")
	s.Entity = e
	c.True(s.Clone(nil).Entity == nil, "a clone made for no entity belongs to none")
}
