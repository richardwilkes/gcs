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

package gurps

import (
	"fmt"
	iofs "io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/library"
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
	pathSet := collection.NewSet[string]()
	f := convertWalker(pathSet, extSet)
	for _, p := range paths {
		_ = filepath.WalkDir(p, f) //nolint:errcheck // We want to continue on even if there was an error
	}
	list := pathSet.Values()
	txt.SortStringsNaturalAscending(list)
	for _, p := range list {
		switch strings.ToLower(filepath.Ext(p)) {
		case library.TraitsExt:
			var data []*Trait
			if data, err = NewTraitsFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = SaveTraits(data, p); err != nil {
				return err
			}
			fmt.Println(p)
		case library.TraitModifiersExt:
			var data []*TraitModifier
			if data, err = NewTraitModifiersFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = SaveTraitModifiers(data, p); err != nil {
				return err
			}
			fmt.Println(p)
		case library.EquipmentExt:
			var data []*Equipment
			if data, err = NewEquipmentFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = SaveEquipment(data, p); err != nil {
				return err
			}
			fmt.Println(p)
		case library.EquipmentModifiersExt:
			var data []*EquipmentModifier
			if data, err = NewEquipmentModifiersFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = SaveEquipmentModifiers(data, p); err != nil {
				return err
			}
			fmt.Println(p)
		case library.SkillsExt:
			var data []*Skill
			if data, err = NewSkillsFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = SaveSkills(data, p); err != nil {
				return err
			}
			fmt.Println(p)
		case library.SpellsExt:
			var data []*Spell
			if data, err = NewSpellsFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = SaveSpells(data, p); err != nil {
				return err
			}
			fmt.Println(p)
		case library.NotesExt:
			var data []*Note
			if data, err = NewNotesFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = SaveNotes(data, p); err != nil {
				return err
			}
			fmt.Println(p)
		case library.TemplatesExt:
			var tmpl *Template
			if tmpl, err = NewTemplateFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = tmpl.Save(p); err != nil {
				return err
			}
			fmt.Println(p)
		case library.SheetExt:
			var entity *Entity
			if entity, err = NewEntityFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p)); err != nil {
				return err
			}
			if err = entity.Save(p); err != nil {
				return err
			}
			fmt.Println(p)
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
