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
	"github.com/richardwilkes/unison"
)

// RegisterFileTypes registers GCS file types.
func RegisterFileTypes() {
	registerExportableGCSFileInfo(library.SheetExt, res.GCSSheetSVG, sheet.NewSheetFromFile)
	registerGCSFileInfo(library.TemplatesExt, []string{library.TemplatesExt}, res.GCSTemplateSVG, sheet.NewTemplateFromFile)
	groupWith := []string{library.TraitsExt, library.TraitModifiersExt, library.EquipmentExt, library.EquipmentModifiersExt, library.SkillsExt, library.SpellsExt, library.NotesExt}
	registerGCSFileInfo(library.TraitsExt, groupWith, res.GCSTraitsSVG, NewTraitTableDockableFromFile)
	registerGCSFileInfo(library.TraitModifiersExt, groupWith, res.GCSTraitModifiersSVG, NewTraitModifierTableDockableFromFile)
	registerGCSFileInfo(library.EquipmentExt, groupWith, res.GCSEquipmentSVG, NewEquipmentTableDockableFromFile)
	registerGCSFileInfo(library.EquipmentModifiersExt, groupWith, res.GCSEquipmentModifiersSVG, NewEquipmentModifierTableDockableFromFile)
	registerGCSFileInfo(library.SkillsExt, groupWith, res.GCSSkillsSVG, NewSkillTableDockableFromFile)
	registerGCSFileInfo(library.SpellsExt, groupWith, res.GCSSpellsSVG, NewSpellTableDockableFromFile)
	registerGCSFileInfo(library.NotesExt, groupWith, res.GCSNotesSVG, NewNoteTableDockableFromFile)
}

func registerGCSFileInfo(ext string, groupWith []string, svg *unison.SVG, loader func(filePath string) (unison.Dockable, error)) {
	library.FileInfo{
		Extension:             ext,
		ExtensionsToGroupWith: groupWith,
		SVG:                   svg,
		Load:                  loader,
		IsGCSData:             true,
	}.Register()
}

func registerExportableGCSFileInfo(ext string, svg *unison.SVG, loader func(filePath string) (unison.Dockable, error)) {
	library.FileInfo{
		Extension:             ext,
		ExtensionsToGroupWith: []string{ext},
		SVG:                   svg,
		Load:                  loader,
		IsGCSData:             true,
		IsExportable:          true,
	}.Register()
}
