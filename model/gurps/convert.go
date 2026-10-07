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
	"path/filepath"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xfilepath"
	"github.com/richardwilkes/toolbox/v2/xslices"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/yookoala/realpath"
)

// Convert the GCS files found in the given paths to the current file format.
func Convert(paths ...string) error {
	var err error
	paths, err = xfilepath.UniquePaths(paths...)
	if err != nil {
		return err
	}
	extSet := xslices.Set(GCSExtensions())
	maps.Copy(extSet, xslices.Set(GCSSecondaryExtensions()))
	pathSet := make(map[string]struct{})
	f := convertWalker(pathSet, extSet)
	for _, p := range paths {
		_ = filepath.WalkDir(p, f) //nolint:errcheck // We want to continue on even if there was an error
	}
	list := slices.SortedFunc(maps.Keys(pathSet), func(a, b string) int { return xstrings.NaturalCmp(a, b, true) })
	for _, p := range list {
		fmt.Printf(i18n.Text("Processing %s\n"), p)
		if convert := converters[strings.ToLower(filepath.Ext(p))]; convert != nil {
			if err = convert(p); err != nil {
				return err
			}
		}
	}
	if len(list) == 1 {
		fmt.Println(i18n.Text("Processed 1 file"))
	} else {
		fmt.Printf(i18n.Text("Processed %d files\n"), len(list))
	}
	return nil
}

// converters maps each GCS file extension, in lowercase, to the function that rewrites a file of that type in the
// current file format. A nil entry marks a type that carries no version information, so there is nothing to update for
// it. Only files whose extension appears in GCSExtensions or GCSSecondaryExtensions are collected for conversion, so an
// entry without a counterpart in one of those is never used.
var converters = map[string]func(p string) error{
	TraitsExt:             convertFile(NewTraitsFromFile, SaveTraits),
	TraitModifiersExt:     convertFile(NewTraitModifiersFromFile, SaveTraitModifiers),
	EquipmentExt:          convertFile(NewEquipmentFromFile, SaveEquipment),
	EquipmentModifiersExt: convertFile(NewEquipmentModifiersFromFile, SaveEquipmentModifiers),
	LootExt:               convertFile(NewLootFromFile, (*Loot).Save),
	SkillsExt:             convertFile(NewSkillsFromFile, SaveSkills),
	SpellsExt:             convertFile(NewSpellsFromFile, SaveSpells),
	NotesExt:              convertFile(NewNotesFromFile, SaveNotes),
	TemplatesExt:          convertFile(NewTemplateFromFile, (*Template).Save),
	// TODO: Re-enable Campaign files
	// CampaignExt:           convertFile(NewCampaignFromFile, (*Campaign).Save),
	SheetExt:           convertFile(NewEntityFromFile, (*Entity).Save),
	AncestryExt:        convertFile(NewAncestryFromFile, (*Ancestry).Save),
	AttributesExt:      convertFile(NewAttributeDefsFromFile, (*AttributeDefs).Save),
	AttributesExtAlt1:  convertFile(NewAttributeDefsFromFile, (*AttributeDefs).Save),
	AttributesExtAlt2:  convertFile(NewAttributeDefsFromFile, (*AttributeDefs).Save),
	BodyExt:            convertFile(NewBodyFromFile, (*Body).Save),
	BodyExtAlt:         convertFile(NewBodyFromFile, (*Body).Save),
	CalendarExt:        nil,
	ColorSettingsExt:   convertFile(colors.NewFromFS, (*colors.Colors).Save),
	FontSettingsExt:    convertFile(fonts.NewFromFS, (*fonts.Fonts).Save),
	GeneralSettingsExt: convertFile(NewGeneralSettingsFromFile, (*GeneralSettings).Save),
	KeySettingsExt:     convertFile(NewKeyBindingsFromFS, (*KeyBindings).Save),
	NamesExt:           nil,
	PageRefSettingsExt: convertFile(NewPageRefsFromFS, (*PageRefs).Save),
	SheetSettingsExt:   convertFile(NewSheetSettingsFromFile, (*SheetSettings).Save),
}

// convertFile returns a function that loads the file at the path it is given with load and writes it back out with
// save, which brings the file up to the current file format.
func convertFile[T any](load func(fs.FS, string) (T, error), save func(T, string) error) func(p string) error {
	return func(p string) error {
		data, err := load(os.DirFS(filepath.Dir(p)), filepath.Base(p))
		if err != nil {
			return err
		}
		return save(data, p)
	}
}

func convertWalker(pathSet, extSet map[string]struct{}) func(path string, d fs.DirEntry, err error) error {
	var f func(path string, d fs.DirEntry, err error) error
	visited := make(map[string]struct{})
	f = func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil //nolint:nilerr // Continue on even if there was an error
		}
		name := d.Name()
		if strings.HasPrefix(name, ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			visited[path] = struct{}{}
		} else {
			if d.Type() == fs.ModeSymlink {
				if path, err = filepath.EvalSymlinks(path); err == nil {
					if _, exists := visited[path]; !exists {
						_ = filepath.WalkDir(path, f) //nolint:errcheck // Continue on even if there was an error
					}
				}
			} else {
				// The set holds lowercase extension constants and the converters map is keyed the same way, so match
				// case-insensitively. Otherwise a file named "Character.GCS" is skipped without a word about it.
				if _, exists := extSet[strings.ToLower(filepath.Ext(name))]; exists {
					if path, err = realpath.Realpath(path); err == nil {
						pathSet[path] = struct{}{}
					}
				}
			}
		}
		return nil
	}
	return f
}
