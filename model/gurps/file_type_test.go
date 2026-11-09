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
	"io/fs"
	"maps"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/uti"
)

// TestFilteredExtensionLists verifies that each of the filtered extension lists selects only the registered
// extensions carrying its flag and returns them in natural sort order, so that a type registered under several
// extensions contributes each of them and a flag on one type never leaks into another list. The model tests don't run
// the ux-layer file-type registration, so the registry is seeded here and restored afterward.
func TestFilteredExtensionLists(t *testing.T) {
	c := check.New(t)
	savedRegistry := maps.Clone(fileTypeRegistry)
	savedKnown := KnownFileTypes
	t.Cleanup(func() {
		fileTypeRegistry = savedRegistry
		KnownFileTypes = savedKnown
	})
	fileTypeRegistry = make(map[string]*FileInfo)
	KnownFileTypes = nil

	(&FileInfo{
		Name:             "GCS Data",
		UTI:              &uti.DataType{Extensions: []string{".test10", ".test2"}},
		IsGCSData:        true,
		IsDeepSearchable: true,
	}).Register()
	(&FileInfo{
		Name: "Openable Only",
		UTI:  &uti.DataType{Extensions: []string{".test1"}},
	}).Register()
	(&FileInfo{
		Name:      "Special",
		UTI:       &uti.DataType{Extensions: []string{".special"}},
		IsSpecial: true,
		IsGCSData: true,
	}).Register()

	c.Equal([]string{".test1", ".test2", ".test10"}, AcceptableExtensions())
	c.Equal([]string{".test2", ".test10"}, DeepSearchableExtensions())
	c.Equal([]string{".special", ".test2", ".test10"}, GCSExtensions())
}

// headerPrefix returns the text every saved file of the given kind starts with: the header, written first.
func headerPrefix(fileType string) string {
	return fmt.Sprintf("{\n\t\"file_type\": %q,\n\t\"version\": %d", fileType, jio.CurrentDataVersion)
}

// dataFileKind describes one kind of data file: how to save a new one and how to load it.
type dataFileKind struct {
	fileType string
	ext      string
	save     func(filePath string) error
	load     func(fileSystem fs.FS, filePath string) error
}

// loadOnly adapts a loader that returns what it loaded into one that reports only whether the load succeeded.
func loadOnly[T any](load func(fs.FS, string) (T, error)) func(fs.FS, string) error {
	return func(fileSystem fs.FS, filePath string) error {
		_, err := load(fileSystem, filePath)
		return err
	}
}

// dataFileKinds returns every kind of data file GCS writes.
func dataFileKinds(t *testing.T) []dataFileKind {
	t.Helper()
	gregorian, err := NewCalendarRefFromFS(embeddedFS, "embedded_data/Gregorian.calendar")
	if err != nil {
		t.Fatal(err)
	}
	var themeColors colors.Colors
	themeColors.CaptureCurrent()
	var themeFonts fonts.Fonts
	themeFonts.CaptureCurrent()
	return []dataFileKind{
		{CharacterDataKind, SheetExt, NewEntity().Save, loadOnly(NewEntityFromFile)},
		{TemplateDataKind, TemplatesExt, NewTemplate().Save, loadOnly(NewTemplateFromFile)},
		{LootDataKind, LootExt, NewLoot().Save, loadOnly(NewLootFromFile)},
		{TraitListDataKind, TraitsExt, saveEmptyList(SaveTraits), loadOnly(NewTraitsFromFile)},
		{
			TraitModifierListDataKind, TraitModifiersExt, saveEmptyList(SaveTraitModifiers),
			loadOnly(NewTraitModifiersFromFile),
		},
		{EquipmentListDataKind, EquipmentExt, saveEmptyList(SaveEquipment), loadOnly(NewEquipmentFromFile)},
		{
			EquipmentModifierListDataKind, EquipmentModifiersExt, saveEmptyList(SaveEquipmentModifiers),
			loadOnly(NewEquipmentModifiersFromFile),
		},
		{SkillListDataKind, SkillsExt, saveEmptyList(SaveSkills), loadOnly(NewSkillsFromFile)},
		{SpellListDataKind, SpellsExt, saveEmptyList(SaveSpells), loadOnly(NewSpellsFromFile)},
		{NoteListDataKind, NotesExt, saveEmptyList(SaveNotes), loadOnly(NewNotesFromFile)},
		{BodyDataKind, BodyExt, FactoryBody().Save, loadOnly(NewBodyFromFile)},
		{AttributesDataKind, AttributesExt, FactoryAttributeDefs().Save, loadOnly(NewAttributeDefsFromFile)},
		{AncestryDataKind, AncestryExt, NewAncestry().Save, loadOnly(NewAncestryFromFile)},
		{FiltersDataKind, ListFiltersExt, saveEmptyList(SaveListFilters), loadOnly(NewListFiltersFromFile)},
		{SkillLevelCapsDataKind, SkillLevelCapsExt, saveEmptyList(SaveSkillLevelCaps), loadOnly(NewSkillLevelCapsFromFile)},
		{SettingsDataKind, ".json", saveSettingsTo, loadOnly(loadSettingsFile)},
		{GeneralSettingsDataKind, GeneralSettingsExt, NewGeneralSettings().Save, loadOnly(NewGeneralSettingsFromFile)},
		{SheetSettingsDataKind, SheetSettingsExt, FactorySheetSettings().Save, loadOnly(NewSheetSettingsFromFile)},
		{KeyBindingsDataKind, KeySettingsExt, (&KeyBindings{}).Save, loadOnly(NewKeyBindingsFromFS)},
		{PageRefsDataKind, PageRefSettingsExt, (&PageRefs{}).Save, loadOnly(NewPageRefsFromFS)},
		{NamesDataKind, NamesExt, NewNameGenerator().Save, loadOnly(ReadNameGeneratorFromFS)},
		{CalendarDataKind, CalendarExt, gregorian.Save, loadOnly(NewCalendarRefFromFS)},
		{colors.DataKind, ColorSettingsExt, themeColors.Save, loadOnly(colors.NewFromFS)},
		{fonts.DataKind, FontSettingsExt, themeFonts.Save, loadOnly(fonts.NewFromFS)},
	}
}

// saveEmptyList adapts a list saver into one that saves an empty list to the path it is given.
func saveEmptyList[T any](save func(T, string) error) func(filePath string) error {
	return func(filePath string) error {
		var empty T
		return save(empty, filePath)
	}
}

// saveSettingsTo saves factory settings to the given path, the way the application saves them to SettingsPath.
func saveSettingsTo(filePath string) error {
	prevPath := SettingsPath
	SettingsPath = filePath
	defer func() { SettingsPath = prevPath }()
	settings := FactorySettings()
	return settings.Save()
}

// loadSettingsFile loads a settings file with the check the application applies to its own, failing rather than
// falling back to the factory defaults.
func loadSettingsFile(fileSystem fs.FS, filePath string) (*Settings, error) {
	var settings Settings
	if err := jio.LoadVersionedFileWithMinimum(fileSystem, filePath, &settings, SettingsDataKind, 0); err != nil {
		return nil, err
	}
	return &settings, nil
}

// TestDataFilesCarryTheirKind verifies that every kind of data file GCS writes starts with a header naming its kind
// and the current data version, that the file loads as that kind, and that it is refused as any other kind with a
// message naming both kinds.
func TestDataFilesCarryTheirKind(t *testing.T) {
	all := dataFileKinds(t)
	for i, kind := range all {
		t.Run(kind.fileType, func(t *testing.T) {
			c := check.New(t)
			dir := t.TempDir()
			base := "Test" + kind.ext
			p := filepath.Join(dir, base)
			c.NoError(kind.save(p))
			data, err := os.ReadFile(p)
			c.NoError(err)
			c.True(strings.HasPrefix(string(data), headerPrefix(kind.fileType)), "the header is written first: %s", data)
			c.NoError(kind.load(os.DirFS(dir), base), "the file loads as its own kind")
			other := all[(i+1)%len(all)]
			err = other.load(os.DirFS(dir), base)
			c.HasError(err, "the file is refused as %s data", other.fileType)
			c.Contains(err.Error(), `"`+kind.fileType+`"`)
			c.Contains(err.Error(), `"`+other.fileType+`"`)
		})
	}
}

// TestEmbeddedDataCarriesTheHeader verifies that the data files built into GCS start with the header, like every file
// it writes.
func TestEmbeddedDataCarriesTheHeader(t *testing.T) {
	c := check.New(t)
	kindsByExt := map[string]string{
		AncestryExt:   AncestryDataKind,
		AttributesExt: AttributesDataKind,
		BodyExt:       BodyDataKind,
		CalendarExt:   CalendarDataKind,
		NamesExt:      NamesDataKind,
	}
	count := 0
	c.NoError(fs.WalkDir(embeddedFS, "embedded_data", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		kind, ok := kindsByExt[path.Ext(p)]
		c.True(ok, "%s has a known extension", p)
		data, err := fs.ReadFile(embeddedFS, p)
		c.NoError(err)
		h, err := jio.ReadHeader(data)
		c.NoError(err, p)
		c.Equal(jio.NewHeader(kind), h, p)
		count++
		return nil
	}))
	c.True(count > 0, "the embedded data was found")
}
