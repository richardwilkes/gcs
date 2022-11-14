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

	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/gcs/v5/model/ancestry"
	"github.com/richardwilkes/gcs/v5/model/library"
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
			var data []*model.Trait
			if data, err = model.NewTraitsFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = model.SaveTraits(data, p); err != nil {
				return err
			}
		case library.TraitModifiersExt:
			var data []*model.TraitModifier
			if data, err = model.NewTraitModifiersFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = model.SaveTraitModifiers(data, p); err != nil {
				return err
			}
		case library.EquipmentExt:
			var data []*model.Equipment
			if data, err = model.NewEquipmentFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = model.SaveEquipment(data, p); err != nil {
				return err
			}
		case library.EquipmentModifiersExt:
			var data []*model.EquipmentModifier
			if data, err = model.NewEquipmentModifiersFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = model.SaveEquipmentModifiers(data, p); err != nil {
				return err
			}
		case library.SkillsExt:
			var data []*model.Skill
			if data, err = model.NewSkillsFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = model.SaveSkills(data, p); err != nil {
				return err
			}
		case library.SpellsExt:
			var data []*model.Spell
			if data, err = model.NewSpellsFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = model.SaveSpells(data, p); err != nil {
				return err
			}
		case library.NotesExt:
			var data []*model.Note
			if data, err = model.NewNotesFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = model.SaveNotes(data, p); err != nil {
				return err
			}
		case library.TemplatesExt:
			var tmpl *model.Template
			if tmpl, err = model.NewTemplateFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = tmpl.Save(p); err != nil {
				return err
			}
		case library.SheetExt:
			var entity *model.Entity
			if entity, err = model.NewEntityFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
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
			var data *model.AttributeDefs
			if data, err = model.NewAttributeDefsFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = data.Save(p); err != nil {
				return err
			}
		case library.BodyExt, library.BodyExtAlt:
			var data *model.Body
			if data, err = model.NewBodyFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
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
			var data *model.GeneralSheetSettings
			if data, err = model.NewGeneralSheetSettingsFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = data.Save(p); err != nil {
				return err
			}
		case library.KeySettingsExt:
			var data *model.KeyBindings
			if data, err = model.NewKeyBindingsFromFS(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = data.Save(p); err != nil {
				return err
			}
		case library.NamesExt:
			// Currently have no version info, so nothing to update
		case library.PageRefSettingsExt:
			var data *model.PageRefs
			if data, err = model.NewPageRefsFromFS(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = data.Save(p); err != nil {
				return err
			}
		case library.SheetSettingsExt:
			var data *model.SheetSettings
			if data, err = model.NewSheetSettingsFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
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
