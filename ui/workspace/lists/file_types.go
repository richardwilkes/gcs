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

package lists

import (
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/gcs/v5/ui/workspace/sheet"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/unison"
)

// RegisterFileTypes registers GCS file types.
func RegisterFileTypes() {
	registerExportableGCSFileInfo("GCS Sheet", library.SheetExt, res.GCSSheetSVG, sheet.NewSheetFromFile)
	registerGCSFileInfo("GCS Template", library.TemplatesExt, []string{library.TemplatesExt}, res.GCSTemplateSVG, sheet.NewTemplateFromFile)
	groupWith := []string{
		library.TraitsExt,
		library.TraitModifiersExt,
		library.EquipmentExt,
		library.EquipmentModifiersExt,
		library.SkillsExt,
		library.SpellsExt,
		library.NotesExt,
	}
	registerGCSFileInfo("GCS Traits", library.TraitsExt, groupWith, res.GCSTraitsSVG, NewTraitTableDockableFromFile)
	registerGCSFileInfo("GCS Trait Modifiers", library.TraitModifiersExt, groupWith, res.GCSTraitModifiersSVG, NewTraitModifierTableDockableFromFile)
	registerGCSFileInfo("GCS Equipment", library.EquipmentExt, groupWith, res.GCSEquipmentSVG, NewEquipmentTableDockableFromFile)
	registerGCSFileInfo("GCS Equipment Modifiers", library.EquipmentModifiersExt, groupWith, res.GCSEquipmentModifiersSVG, NewEquipmentModifierTableDockableFromFile)
	registerGCSFileInfo("GCS Skills", library.SkillsExt, groupWith, res.GCSSkillsSVG, NewSkillTableDockableFromFile)
	registerGCSFileInfo("GCS Spells", library.SpellsExt, groupWith, res.GCSSpellsSVG, NewSpellTableDockableFromFile)
	registerGCSFileInfo("GCS Notes", library.NotesExt, groupWith, res.GCSNotesSVG, NewNoteTableDockableFromFile)
}

func registerGCSFileInfo(name, ext string, groupWith []string, svg *unison.SVG, loader func(filePath string) (unison.Dockable, error)) {
	library.FileInfo{
		Name:       name,
		UTI:        cmdline.AppIdentifier + ext,
		ConformsTo: []string{"public.data"},
		Extensions: []string{ext},
		GroupWith:  groupWith,
		MimeTypes:  []string{"application/x-gcs-" + ext[1:]},
		SVG:        svg,
		Load:       loader,
		IsGCSData:  true,
	}.Register()
}

func registerExportableGCSFileInfo(name, ext string, svg *unison.SVG, loader func(filePath string) (unison.Dockable, error)) {
	library.FileInfo{
		Name:         name,
		UTI:          cmdline.AppIdentifier + ext,
		ConformsTo:   []string{"public.data"},
		Extensions:   []string{ext},
		GroupWith:    []string{ext},
		MimeTypes:    []string{"application/x-gcs-" + ext[1:]},
		SVG:          svg,
		Load:         loader,
		IsGCSData:    true,
		IsExportable: true,
	}.Register()
}
