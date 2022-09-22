/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package convert

import (
	"fmt"
	iofs "io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/ancestry"
	gsettings "github.com/richardwilkes/gcs/v5/model/gurps/settings"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/gcs/v5/model/theme"
	"github.com/richardwilkes/toolbox/collection"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/yookoala/realpath"
)

// Convert the GCS files found in the given paths to the current file format.
func Convert(paths ...string) error {
	var err error
	paths, err = fs.UniquePaths(paths...)
	if err != nil {
		return err
	}
	extSet := collection.NewSet(library.GCSExtensions()...)
	extSet.Add(library.GCSSecondaryExtensions()...)
	pathSet := collection.NewSet[string]()
	f := convertWalker(pathSet, extSet)
	for _, p := range paths {
		_ = filepath.WalkDir(p, f) //nolint:errcheck // We want to continue on even if there was an error
	}
	list := pathSet.Values()
	txt.SortStringsNaturalAscending(list)
	for _, p := range list {
		fmt.Printf(i18n.Text("Processing %s\n"), p)
		switch strings.ToLower(filepath.Ext(p)) {
		case library.TraitsExt:
			var data []*gurps.Trait
			if data, err = gurps.NewTraitsFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = gurps.SaveTraits(data, p); err != nil {
				return err
			}
		case library.TraitModifiersExt:
			var data []*gurps.TraitModifier
			if data, err = gurps.NewTraitModifiersFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = gurps.SaveTraitModifiers(data, p); err != nil {
				return err
			}
		case library.EquipmentExt:
			var data []*gurps.Equipment
			if data, err = gurps.NewEquipmentFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = gurps.SaveEquipment(data, p); err != nil {
				return err
			}
		case library.EquipmentModifiersExt:
			var data []*gurps.EquipmentModifier
			if data, err = gurps.NewEquipmentModifiersFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = gurps.SaveEquipmentModifiers(data, p); err != nil {
				return err
			}
		case library.SkillsExt:
			var data []*gurps.Skill
			if data, err = gurps.NewSkillsFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = gurps.SaveSkills(data, p); err != nil {
				return err
			}
		case library.SpellsExt:
			var data []*gurps.Spell
			if data, err = gurps.NewSpellsFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = gurps.SaveSpells(data, p); err != nil {
				return err
			}
		case library.NotesExt:
			var data []*gurps.Note
			if data, err = gurps.NewNotesFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = gurps.SaveNotes(data, p); err != nil {
				return err
			}
		case library.TemplatesExt:
			var tmpl *gurps.Template
			if tmpl, err = gurps.NewTemplateFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = tmpl.Save(p); err != nil {
				return err
			}
		case library.SheetExt:
			var entity *gurps.Entity
			if entity, err = gurps.NewEntityFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = entity.Save(p); err != nil {
				return err
			}
		case library.AncestryExt:
			var data *ancestry.Ancestry
			if data, err = ancestry.NewAncestryFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = data.Save(p); err != nil {
				return err
			}
		case library.AttributesExt, library.AttributesExtAlt1, library.AttributesExtAlt2:
			var data *gurps.AttributeDefs
			if data, err = gurps.NewAttributeDefsFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = data.Save(p); err != nil {
				return err
			}
		case library.BodyExt, library.BodyExtAlt:
			var data *gurps.Body
			if data, err = gurps.NewBodyFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = data.Save(p); err != nil {
				return err
			}
		case library.CalendarExt:
			// Currently have no version info, so nothing to update
		case library.ColorSettingsExt:
			var data *theme.Colors
			if data, err = theme.NewColorsFromFS(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = data.Save(p); err != nil {
				return err
			}
		case library.FontSettingsExt:
			var data *theme.Fonts
			if data, err = theme.NewFontsFromFS(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = data.Save(p); err != nil {
				return err
			}
		case library.GeneralSettingsExt:
			var data *gsettings.General
			if data, err = gsettings.NewGeneralFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = data.Save(p); err != nil {
				return err
			}
		case library.KeySettingsExt:
			var data *settings.KeyBindings
			if data, err = settings.NewKeyBindingsFromFS(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = data.Save(p); err != nil {
				return err
			}
		case library.NamesExt:
			// Currently have no version info, so nothing to update
		case library.PageRefSettingsExt:
			var data *settings.PageRefs
			if data, err = settings.NewPageRefsFromFS(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = data.Save(p); err != nil {
				return err
			}
		case library.SheetSettingsExt:
			var data *gurps.SheetSettings
			if data, err = gurps.NewSheetSettingsFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = data.Save(p); err != nil {
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

func convertWalker(pathSet, extSet collection.Set[string]) func(path string, d iofs.DirEntry, err error) error {
	var f func(path string, d iofs.DirEntry, err error) error
	visited := collection.NewSet[string]()
	f = func(path string, d iofs.DirEntry, err error) error {
		name := d.Name()
		if strings.HasPrefix(name, ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if err == nil {
			if d.IsDir() {
				visited.Add(path)
			} else {
				if d.Type() == iofs.ModeSymlink {
					if path, err = filepath.EvalSymlinks(path); err == nil && !visited.Contains(path) {
						_ = filepath.WalkDir(path, f) //nolint:errcheck // We want to continue on even if there was an error
					}
				} else {
					if extSet.Contains(filepath.Ext(name)) {
						if path, err = realpath.Realpath(path); err == nil {
							pathSet.Add(path)
						}
					}
				}
			}
		}
		return nil
	}
	return f
}
