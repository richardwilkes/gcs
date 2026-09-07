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
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestConvertWalkerUnreadableRoot verifies that the conversion walker tolerates a root it can't stat. filepath.WalkDir
// hands the callback a nil DirEntry along with the error in that case, which previously panicked on the immediate
// d.Name() call rather than skipping the path.
func TestConvertWalkerUnreadableRoot(t *testing.T) {
	c := check.New(t)
	pathSet := make(map[string]struct{})
	extSet := map[string]struct{}{SheetExt: {}}
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	c.NotPanics(func() {
		c.NoError(filepath.WalkDir(missing, convertWalker(pathSet, extSet)), "a missing root is skipped, not reported")
	}, "a missing root must not panic")
	c.Equal(0, len(pathSet), "nothing is collected from a missing root")

	// A readable root still collects the files with matching extensions.
	dir := t.TempDir()
	wanted := filepath.Join(dir, "sheet"+SheetExt)
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
		"lower" + SheetExt,
		"upper" + strings.ToUpper(SheetExt),
		"Mixed.GcS",
		"template" + strings.ToUpper(TemplatesExt),
	}
	for _, name := range append(slices.Clone(wanted), "notes.TXT") {
		c.NoError(os.WriteFile(filepath.Join(dir, name), []byte("{}"), 0o600))
	}

	pathSet := make(map[string]struct{})
	extSet := map[string]struct{}{SheetExt: {}, TemplatesExt: {}}
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
// extension lists. The primary extensions are listed here because the file type registry that GCSExtensions reads from
// is populated by the ux package.
func TestConvertersCoverCollectedExtensions(t *testing.T) {
	c := check.New(t)
	primary := []string{
		TraitsExt, TraitModifiersExt, EquipmentExt, EquipmentModifiersExt, LootExt, SkillsExt, SpellsExt, NotesExt,
		TemplatesExt, SheetExt,
	}
	for _, ext := range append(primary, GCSSecondaryExtensions()...) {
		_, exists := converters[ext]
		c.True(exists, "%s has a converter entry", ext)
	}
	for ext := range converters {
		c.Equal(strings.ToLower(ext), ext, "%s is keyed in lowercase, which is what the dispatch looks up", ext)
	}
}

// TestConvertRewritesCollectedFiles verifies that Convert brings each collected file up to the current data version,
// including files that use an alternate extension, and leaves files whose type carries no version information
// untouched.
func TestConvertRewritesCollectedFiles(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	oldVersion := jio.CurrentDataVersion - 1
	bodyPath := filepath.Join(dir, "body"+BodyExtAlt)
	c.NoError(jio.SaveToFile(bodyPath, &standaloneBodyData{Version: oldVersion, BodyData: FactoryBody().BodyData}))
	attrPath := filepath.Join(dir, "attributes"+AttributesExtAlt1)
	c.NoError(jio.SaveToFile(attrPath, &attributeDefsData{Version: oldVersion, Rows: FactoryAttributeDefs()}))
	calendarPath := filepath.Join(dir, "calendar"+CalendarExt)
	calendarData := []byte(`{"version":1}`)
	c.NoError(os.WriteFile(calendarPath, calendarData, 0o600))

	c.NoError(Convert(dir))

	c.Equal(jio.CurrentDataVersion, fileVersion(c, bodyPath), "the body file is rewritten in the current format")
	c.Equal(jio.CurrentDataVersion, fileVersion(c, attrPath), "the attributes file is rewritten in the current format")
	data, err := os.ReadFile(calendarPath)
	c.NoError(err)
	c.Equal(calendarData, data, "a file type with no version information is left untouched")
}

// TestConvertFile verifies that the converter built for a list file type reloads the file and writes it back out
// with the current data version, and that a file which cannot be loaded is reported rather than overwritten.
func TestConvertFile(t *testing.T) {
	c := check.New(t)
	traitsPath := filepath.Join(t.TempDir(), "traits"+TraitsExt)
	c.NoError(jio.SaveToFile(traitsPath, &listData[*Trait]{Version: jio.CurrentDataVersion - 1}))
	c.NoError(converters[TraitsExt](traitsPath))
	c.Equal(jio.CurrentDataVersion, fileVersion(c, traitsPath))

	broken := []byte("not json")
	c.NoError(os.WriteFile(traitsPath, broken, 0o600))
	c.HasError(converters[TraitsExt](traitsPath), "a file that fails to load is reported")
	data, err := os.ReadFile(traitsPath)
	c.NoError(err)
	c.Equal(broken, data, "a file that fails to load is not overwritten")
}

// fileVersion returns the data version recorded in the JSON file at the given path.
func fileVersion(c check.Checker, p string) int {
	var data struct {
		Version int `json:"version"`
	}
	c.NoError(jio.LoadFromFile(nil, p, &data))
	return data.Version
}
