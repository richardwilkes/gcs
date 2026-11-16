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
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// TestConvertWalker verifies which files the conversion walker collects: those with one of the wanted extensions, in
// any case, other than hidden files and what hidden directories hold, including what a symlink leads to.
func TestConvertWalker(t *testing.T) {
	for _, tc := range []struct {
		name string
		// root is the directory to walk and files the files to create, each relative to the test's directory.
		root  string
		files []string
		// links maps each symlink to create to its target, each relative to the test's directory.
		links map[string]string
		// want holds the names of the files that are collected, sorted.
		want []string
	}{
		// filepath.WalkDir hands the walker a nil DirEntry along with the error for a root it can't stat.
		{name: "a root that does not exist is skipped", root: "does-not-exist", want: []string{}},
		{
			name:  "extensions match regardless of case",
			files: []string{"lower.gcs", "upper.GCS", "Mixed.GcS", "template.GCT", "notes.TXT"},
			want:  []string{"Mixed.GcS", "lower.gcs", "template.GCT", "upper.GCS"},
		},
		{
			name:  "hidden files and directories are skipped",
			files: []string{".hidden.gcs", ".hidden/inside.gcs", "seen.gcs", "sub/inner.gcs"},
			want:  []string{"inner.gcs", "seen.gcs"},
		},
		{
			name:  "symlinks are followed",
			root:  "root",
			files: []string{"outside/dir/in-dir.gcs", "outside/file.gcs"},
			links: map[string]string{"root/dir-link": "outside/dir", "root/file-link": "outside/file.gcs"},
			want:  []string{"file.gcs", "in-dir.gcs"},
		},
		{
			// Following this one would never end.
			name:  "a symlink back to a directory it lies within is not followed",
			files: []string{"sub/inner.gcs"},
			links: map[string]string{"sub/up": "."},
			want:  []string{"inner.gcs"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			// Convert resolves the paths it is given before walking them, which the walker relies on to recognize a
			// directory it has already visited.
			dir, err := filepath.EvalSymlinks(t.TempDir())
			c.NoError(err)
			for _, name := range tc.files {
				p := filepath.Join(dir, name)
				c.NoError(os.MkdirAll(filepath.Dir(p), 0o750))
				c.NoError(os.WriteFile(p, []byte("{}"), 0o600))
			}
			for link, target := range tc.links {
				p := filepath.Join(dir, link)
				c.NoError(os.MkdirAll(filepath.Dir(p), 0o750))
				// Windows requires a privilege for this by default.
				if err = os.Symlink(filepath.Join(dir, target), p); err != nil {
					t.Skipf("unable to create a symlink here: %v", err)
				}
			}

			pathSet := make(map[string]struct{})
			walker := convertWalker(pathSet, map[string]struct{}{".gcs": {}, ".gct": {}})
			c.NotPanics(func() { c.NoError(filepath.WalkDir(filepath.Join(dir, tc.root), walker)) })

			collected := make([]string, 0, len(pathSet))
			for p := range pathSet {
				collected = append(collected, filepath.Base(p))
			}
			slices.Sort(collected)
			c.Equal(tc.want, collected)
		})
	}
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
// whether it is of a registered file type or a secondary one, uses an alternate extension or an upper-case one, or was
// written before such files carried a header, which it gains with its content kept, and that a file which cannot be
// loaded is reported rather than overwritten.
func TestConvertRewritesCollectedFiles(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	outdated := []struct {
		name string
		save func(string) error
	}{
		{name: "traits" + gurps.TraitsExt, save: func(p string) error { return gurps.SaveTraits(nil, p) }},
		{name: "body" + gurps.BodyExtAlt, save: gurps.FactoryBody().Save},
		{name: "UPPER" + strings.ToUpper(gurps.BodyExt), save: gurps.FactoryBody().Save},
		{name: "attributes" + gurps.AttributesExtAlt1, save: gurps.FactoryAttributeDefs().Save},
	}
	for _, one := range outdated {
		saveWithOldVersion(c, filepath.Join(dir, one.name), one.save)
	}
	binding := unison.KeyBinding{KeyCode: unison.KeyK, Modifiers: mod.Control | mod.Shift}.Key()
	headerless := []struct {
		ext, kind, content string
		// want is the member the content must still hold after conversion, as a path from the top level.
		want []string
	}{
		{
			ext: gurps.GeneralSettingsExt, kind: gurps.GeneralSettingsDataKind,
			content: `{"version":2,"default_player_name":"Converted"}`,
			want:    []string{"general", "default_player_name"},
		},
		{
			ext: gurps.SheetSettingsExt, kind: gurps.SheetSettingsDataKind,
			content: `{"damage_progression":"knowing_your_own_strength"}`,
			want:    []string{"damage_progression"},
		},
		{
			ext: gurps.KeySettingsExt, kind: gurps.KeyBindingsDataKind,
			content: `{"cut":"` + binding + `"}`,
			want:    []string{"bindings", "cut"},
		},
		{
			ext: gurps.PageRefSettingsExt, kind: gurps.PageRefsDataKind,
			content: `{"B":{"path":"/refs/basic.pdf","offset":2}}`,
			want:    []string{"refs", "B", "path"},
		},
		{
			// A compound generator with nothing to combine can't generate names, but is still converted.
			ext: gurps.NamesExt, kind: gurps.NamesDataKind,
			content: `{"type":"compound","separator":"-"}`,
			want:    []string{"separator"},
		},
		{
			ext: gurps.CalendarExt, kind: gurps.CalendarDataKind,
			content: `{"weekdays":["One","Two"],"months":[{"name":"Only","days":30}],"day_zero_weekday":0}`,
			want:    []string{"months"},
		},
	}
	for _, one := range headerless {
		c.NoError(os.WriteFile(filepath.Join(dir, "legacy"+one.ext), []byte(one.content), 0o600))
	}

	c.NoError(Convert(dir))

	for _, one := range outdated {
		c.Equal(jio.CurrentDataVersion, fileVersion(c, filepath.Join(dir, one.name)),
			"%s is rewritten in the current format", one.name)
	}
	for _, one := range headerless {
		p := filepath.Join(dir, "legacy"+one.ext)
		data, err := os.ReadFile(p)
		c.NoError(err)
		h, err := jio.ReadHeader(data)
		c.NoError(err)
		c.Equal(jio.NewHeader(one.kind), h, "the %s file is rewritten with the header", one.ext)
		var content map[string]any
		c.NoError(jio.Unmarshal(data, &content))
		var value any = content
		for _, key := range one.want {
			m, ok := value.(map[string]any)
			c.True(ok, "the %s file holds an object at %q", one.ext, key)
			value, ok = m[key]
			c.True(ok, "the %s file keeps %q", one.ext, key)
		}
	}

	// A filter list is used because a nil one saves without complaint, so ignoring the failed load would overwrite it.
	broken := []byte("not json")
	brokenPath := filepath.Join(dir, "broken"+gurps.ListFiltersExt)
	c.NoError(os.WriteFile(brokenPath, broken, 0o600))
	c.HasError(Convert(dir), "a file that fails to load is reported")
	data, err := os.ReadFile(brokenPath)
	c.NoError(err)
	c.Equal(broken, data, "a file that fails to load is not overwritten")
}

// saveWithOldVersion writes a file at the given path with save, the way the application does, and then rewinds the data
// version recorded in it by one and drops the file type, as a file written before headers existed has neither, which
// gives a conversion something to update.
func saveWithOldVersion(c check.Checker, p string, save func(string) error) {
	c.NoError(save(p))
	var data map[string]any
	c.NoError(jio.LoadFromFile(nil, p, &data))
	delete(data, "file_type")
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
