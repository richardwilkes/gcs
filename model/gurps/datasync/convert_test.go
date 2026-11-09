// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package datasync

import (
	"bytes"
	"encoding/json/jsontext"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/gurpstest"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// TestConvertWalkerUnreadableRoot verifies that the conversion walker tolerates a root it can't stat. filepath.WalkDir
// hands the callback a nil DirEntry along with the error in that case, which previously panicked on the immediate
// d.Name() call rather than skipping the path.
func TestConvertWalkerUnreadableRoot(t *testing.T) {
	c := check.New(t)
	pathSet := make(map[string]struct{})
	extSet := map[string]struct{}{gurps.SheetExt: {}}
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	c.NotPanics(func() {
		c.NoError(filepath.WalkDir(missing, convertWalker(pathSet, extSet)), "a missing root is skipped, not reported")
	}, "a missing root must not panic")
	c.Equal(0, len(pathSet), "nothing is collected from a missing root")

	// A readable root still collects the files with matching extensions.
	dir := t.TempDir()
	wanted := filepath.Join(dir, "sheet"+gurps.SheetExt)
	c.NoError(os.WriteFile(wanted, []byte("{}"), 0o600))
	c.NoError(os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("{}"), 0o600))
	c.NoError(filepath.WalkDir(dir, convertWalker(pathSet, extSet)))
	c.Equal(1, len(pathSet), "only the file with a matching extension is collected")
}

// TestConvertWalkerIgnoresExtensionCase verifies that the conversion walker collects data files whose extension is
// upper- or mixed-case. The converters map is keyed by the lowercase extension, so a file the walker declines to
// collect here is silently skipped by both --convert and --sync, with nothing said about it.
func TestConvertWalkerIgnoresExtensionCase(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	wanted := []string{
		"lower" + gurps.SheetExt,
		"upper" + strings.ToUpper(gurps.SheetExt),
		"Mixed.GcS",
		"template" + strings.ToUpper(gurps.TemplatesExt),
	}
	for _, name := range append(slices.Clone(wanted), "notes.TXT") {
		c.NoError(os.WriteFile(filepath.Join(dir, name), []byte("{}"), 0o600))
	}

	pathSet := make(map[string]struct{})
	extSet := map[string]struct{}{gurps.SheetExt: {}, gurps.TemplatesExt: {}}
	c.NoError(filepath.WalkDir(dir, convertWalker(pathSet, extSet)))

	// The collected paths have been resolved through any symlinks, so compare the names rather than the paths.
	collected := make(map[string]struct{}, len(pathSet))
	for p := range pathSet {
		collected[filepath.Base(p)] = struct{}{}
	}
	for _, name := range wanted {
		_, exists := collected[name]
		c.True(exists, "%s is collected", name)
	}
	c.Equal(len(wanted), len(collected), "the file with an unrelated extension is not collected")
}

// TestConvertersCoverCollectedExtensions verifies that every extension the conversion walker collects has an entry in
// the converters map, so a file type GCS owns is never silently skipped because the map fell out of sync with the
// extension lists. The primary extensions are listed here because the file type registry that gurps.GCSExtensions
// reads from is populated by the ux package.
func TestConvertersCoverCollectedExtensions(t *testing.T) {
	c := check.New(t)
	primary := []string{
		gurps.TraitsExt, gurps.TraitModifiersExt, gurps.EquipmentExt, gurps.EquipmentModifiersExt, gurps.LootExt,
		gurps.SkillsExt, gurps.SpellsExt, gurps.NotesExt, gurps.TemplatesExt, gurps.SheetExt,
	}
	for _, ext := range append(primary, gurps.GCSSecondaryExtensions()...) {
		_, exists := converters[ext]
		c.True(exists, "%s has a converter entry", ext)
	}
	for ext := range converters {
		c.Equal(strings.ToLower(ext), ext, "%s is keyed in lowercase, which is what the dispatch looks up", ext)
	}
}

// TestConvertRewritesCollectedFiles verifies that Convert brings each collected file up to the current data version,
// including files that use an alternate extension and settings files written before they carried a header, which gain
// one with their content kept in a member beside it, and leaves files whose type carries no version information
// untouched.
func TestConvertRewritesCollectedFiles(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	bodyPath := filepath.Join(dir, "body"+gurps.BodyExtAlt)
	saveWithOldVersion(c, bodyPath, gurps.FactoryBody().Save)
	attrPath := filepath.Join(dir, "attributes"+gurps.AttributesExtAlt1)
	saveWithOldVersion(c, attrPath, gurps.FactoryAttributeDefs().Save)
	calendarPath := filepath.Join(dir, "calendar"+gurps.CalendarExt)
	calendarData := []byte(`{"version":1}`)
	c.NoError(os.WriteFile(calendarPath, calendarData, 0o600))
	binding := unison.KeyBinding{KeyCode: unison.KeyK, Modifiers: mod.Control | mod.Shift}.Key()
	settingsFiles := []struct {
		ext, kind, member, key, content string
		want                            any
	}{
		{
			ext: gurps.GeneralSettingsExt, kind: gurps.GeneralSettingsDataKind, member: "general",
			key: "default_player_name", content: `{"version":2,"default_player_name":"Converted"}`, want: "Converted",
		},
		{
			ext: gurps.SheetSettingsExt, kind: gurps.SheetSettingsDataKind, member: "sheet_settings",
			key: "damage_progression", content: `{"damage_progression":"knowing_your_own_strength"}`,
			want: "knowing_your_own_strength",
		},
		{
			ext: gurps.KeySettingsExt, kind: gurps.KeyBindingsDataKind, member: "bindings",
			key: "cut", content: `{"cut":"` + binding + `"}`, want: binding,
		},
		{
			ext: gurps.PageRefSettingsExt, kind: gurps.PageRefsDataKind, member: "refs",
			key: "B", content: `{"B":{"path":"/refs/basic.pdf","offset":2}}`,
			want: map[string]any{"path": "/refs/basic.pdf", "offset": float64(2)},
		},
	}
	for _, one := range settingsFiles {
		c.NoError(os.WriteFile(filepath.Join(dir, "settings"+one.ext), []byte(one.content), 0o600))
	}

	c.NoError(Convert(dir))

	c.Equal(jio.CurrentDataVersion, fileVersion(c, bodyPath), "the body file is rewritten in the current format")
	c.Equal(jio.CurrentDataVersion, fileVersion(c, attrPath), "the attributes file is rewritten in the current format")
	for _, one := range settingsFiles {
		p := filepath.Join(dir, "settings"+one.ext)
		h, _, err := jio.ReadHeader(nil, p)
		c.NoError(err)
		c.Equal(jio.NewHeader(one.kind), h, "the %s file is rewritten with the header", one.ext)
		var data map[string]any
		c.NoError(jio.LoadFromFile(nil, p, &data))
		member, ok := data[one.member].(map[string]any)
		c.True(ok, "the %s file holds its content under %q", one.ext, one.member)
		c.Equal(one.want, member[one.key], "the %s file keeps its content", one.ext)
	}
	data, err := os.ReadFile(calendarPath)
	c.NoError(err)
	c.Equal(calendarData, data, "a file type with no version information is left untouched")
}

// TestConvertFile verifies that the converter built for a list file type reloads the file and writes it back out
// with the current data version, and that a file which cannot be loaded is reported rather than overwritten.
func TestConvertFile(t *testing.T) {
	c := check.New(t)
	traitsPath := filepath.Join(t.TempDir(), "traits"+gurps.TraitsExt)
	saveWithOldVersion(c, traitsPath, func(p string) error { return gurps.SaveTraits(nil, p) })
	c.NoError(converters[gurps.TraitsExt](traitsPath))
	c.Equal(jio.CurrentDataVersion, fileVersion(c, traitsPath))

	broken := []byte("not json")
	c.NoError(os.WriteFile(traitsPath, broken, 0o600))
	c.HasError(converters[gurps.TraitsExt](traitsPath), "a file that fails to load is reported")
	data, err := os.ReadFile(traitsPath)
	c.NoError(err)
	c.Equal(broken, data, "a file that fails to load is not overwritten")
}

// saveWithOldVersion writes a file at the given path with save, the way the application does, and then rewinds the data
// version recorded in it by one, dropping the header members that version did not write, which gives a conversion
// something to update.
func saveWithOldVersion(c check.Checker, p string, save func(string) error) {
	c.NoError(save(p))
	var data map[string]any
	c.NoError(jio.LoadFromFile(nil, p, &data))
	delete(data, "type")
	delete(data, "min_version")
	data["version"] = jio.CurrentDataVersion - 1
	c.NoError(jio.SaveToFile(p, data))
	c.Equal(jio.CurrentDataVersion-1, fileVersion(c, p))
}

// fileVersion returns the data version recorded in the JSON file at the given path.
func fileVersion(c check.Checker, p string) int {
	var data struct {
		Version int `json:"version"`
	}
	c.NoError(jio.LoadFromFile(nil, p, &data))
	return data.Version
}

// gurpsTestdata is the gurps package's testdata directory, whose data files the conversion tests also use.
const gurpsTestdata = "../testdata"

// fileRoundTrip loads the file at from, saves what it loaded to the file at to and returns what it loaded.
type fileRoundTrip func(from, to string) (any, error)

// roundTripOf returns the fileRoundTrip for the kind of file that load and save read and write.
func roundTripOf[T any](load func(fs.FS, string) (T, error), save func(T, string) error) fileRoundTrip {
	return func(from, to string) (any, error) {
		data, err := load(nil, from)
		if err != nil {
			return nil, err
		}
		return data, save(data, to)
	}
}

// fileRoundTrips holds the fileRoundTrip for each kind of file, keyed by its extension.
var fileRoundTrips = map[string]fileRoundTrip{
	gurps.SheetExt:     roundTripOf(gurps.NewEntityFromFile, (*gurps.Entity).Save),
	gurps.TemplatesExt: roundTripOf(gurps.NewTemplateFromFile, (*gurps.Template).Save),
	gurps.LootExt:      roundTripOf(gurps.NewLootFromFile, (*gurps.Loot).Save),
	gurps.NotesExt:     roundTripOf(gurps.NewNotesFromFile, gurps.SaveNotes),
	gurps.EquipmentExt: roundTripOf(gurps.NewEquipmentFromFile, gurps.SaveEquipment),
}

// fallbacksIn follows the exported and embedded fields, pointers, slices, maps and interfaces reachable from v. It
// returns how many fields named Unknown of type jsontext.Value, which keep the members this release does not know, it
// reached, along with the path of each that holds any.
func fallbacksIn(v any) (reached int, nonEmpty []string) {
	type visit struct {
		ptr uintptr
		typ reflect.Type
	}
	fallbackType := reflect.TypeFor[jsontext.Value]()
	seen := make(map[visit]bool)
	var walk func(v reflect.Value, path string)
	walk = func(v reflect.Value, path string) {
		switch v.Kind() {
		case reflect.Pointer:
			key := visit{ptr: v.Pointer(), typ: v.Type()}
			if !v.IsNil() && !seen[key] {
				seen[key] = true
				walk(v.Elem(), path)
			}
		case reflect.Interface:
			if !v.IsNil() {
				walk(v.Elem(), path)
			}
		case reflect.Slice, reflect.Array:
			if v.Type().Elem().Kind() != reflect.Uint8 {
				for i := range v.Len() {
					walk(v.Index(i), fmt.Sprintf("%s[%d]", path, i))
				}
			}
		case reflect.Map:
			for iter := v.MapRange(); iter.Next(); {
				walk(iter.Value(), fmt.Sprintf("%s[%v]", path, iter.Key()))
			}
		case reflect.Struct:
			for i := range v.NumField() {
				field := v.Type().Field(i)
				if field.Name == "Unknown" && field.Type == fallbackType {
					reached++
					if v.Field(i).Len() != 0 {
						nonEmpty = append(nonEmpty, path+".Unknown")
					}
				} else if field.IsExported() || field.Anonymous {
					walk(v.Field(i), path+"."+field.Name)
				}
			}
		}
	}
	walk(reflect.ValueOf(v), "")
	return reached, nonEmpty
}

// TestConvertedFilesLeaveNoMembersUnknown verifies that this release knows every member of the files it writes and of the
// older ones in testdata: once one is loaded, no Unknown field holds any, so neither a member a marshaler adds, such
// as "calc", nor a legacy one a loader reads is carried as unknown and written back beside itself, which fails the
// save. Saving what a save wrote must also reproduce it byte for byte.
func TestConvertedFilesLeaveNoMembersUnknown(t *testing.T) {
	large := filepath.Join(t.TempDir(), "large"+gurps.SheetExt)
	size := gurpstest.DefaultLargeEntitySize.Scaled(0.1)
	if err := gurpstest.NewLargeEntity(&size).Save(large); err != nil {
		t.Fatal(err)
	}
	files := []string{large}
	entries, err := os.ReadDir(gurpsTestdata)
	if err != nil {
		t.Fatal(err)
	}
	for _, one := range entries {
		// testdata also holds the files of other tests, which aren't data files.
		if _, ok := fileRoundTrips[filepath.Ext(one.Name())]; ok {
			files = append(files, filepath.Join(gurpsTestdata, one.Name()))
		}
	}
	for _, path := range files {
		t.Run(filepath.Base(path), func(t *testing.T) {
			c := check.New(t)
			ext := filepath.Ext(path)
			from := path
			saved := make([][]byte, 2)
			for i := range saved {
				to := filepath.Join(t.TempDir(), "saved"+ext)
				loaded, loadErr := fileRoundTrips[ext](from, to)
				if loadErr != nil {
					t.Fatal(loadErr)
				}
				reached, nonEmpty := fallbacksIn(loaded)
				c.True(reached > 0, "load %d reaches no Unknown field", i+1)
				c.Equal([]string(nil), nonEmpty, "load %d keeps members as unknown", i+1)
				if saved[i], err = os.ReadFile(to); err != nil {
					t.Fatal(err)
				}
				from = to
			}
			c.True(bytes.Equal(saved[0], saved[1]), "saving what a save wrote changes it")
		})
	}
}
