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

package ux

import (
	"strings"

	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/maps"
)

// RegisterKnownFileTypes registers the known files types.
func RegisterKnownFileTypes() {
	registerNavigatorFileTypes()
	RegisterExternalFileTypes()
	RegisterGCSFileTypes()
}

func registerNavigatorFileTypes() {
	registerSpecialFileInfo(library.ClosedFolder, svg.ClosedFolder)
	registerSpecialFileInfo(library.OpenFolder, svg.OpenFolder)
	registerSpecialFileInfo(library.GenericFile, svg.GenericFile)
}

func registerSpecialFileInfo(key string, svg *unison.SVG) {
	library.FileInfo{
		Extensions: []string{key},
		SVG:        svg,
		IsSpecial:  true,
	}.Register()
}

// RegisterExternalFileTypes registers the external file types.
func RegisterExternalFileTypes() {
	registerPDFFileInfo()
	registerMarkdownFileInfo()
	all := make(map[string]bool)
	for _, one := range unison.KnownImageFormatFormats {
		if one.CanRead() {
			for _, ext := range one.Extensions() {
				all[ext] = true
			}
		}
	}
	groupWith := maps.Keys(all)
	for _, one := range unison.KnownImageFormatFormats {
		if one.CanRead() {
			registerImageFileInfo(one, groupWith)
		}
	}
}

func registerImageFileInfo(format unison.EncodedImageFormat, groupWith []string) {
	library.FileInfo{
		Name:       strings.ToUpper(format.Extension()[1:]) + " Image",
		UTI:        format.UTI(),
		ConformsTo: []string{"public.image"},
		Extensions: format.Extensions(),
		GroupWith:  groupWith,
		MimeTypes:  format.MimeTypes(),
		SVG:        svg.ImageFile,
		Load:       NewImageDockable,
		IsImage:    true,
	}.Register()
}

func registerPDFFileInfo() {
	library.FileInfo{
		Name:       "PDF Document",
		UTI:        "com.adobe.pdf",
		ConformsTo: []string{"public.data"},
		Extensions: []string{".pdf"},
		GroupWith:  []string{".pdf"},
		MimeTypes:  []string{"application/pdf", "application/x-pdf"},
		SVG:        svg.PDFFile,
		Load:       NewPDFDockable,
		IsPDF:      true,
	}.Register()
}

func registerMarkdownFileInfo() {
	extensions := []string{".md", ".markdown"}
	library.FileInfo{
		Name:       "Markdown Document",
		UTI:        "net.daringfireball.markdown",
		ConformsTo: []string{"public.plain-text"},
		Extensions: extensions,
		GroupWith:  extensions,
		MimeTypes:  []string{"text/markdown"},
		SVG:        svg.MarkdownFile,
		Load:       NewMarkdownDockable,
	}.Register()
}

// RegisterGCSFileTypes registers the GCS file types.
func RegisterGCSFileTypes() {
	registerExportableGCSFileInfo("GCS Sheet", library.SheetExt, svg.GCSSheet, NewSheetFromFile)
	registerGCSFileInfo("GCS Template", library.TemplatesExt, []string{library.TemplatesExt}, svg.GCSTemplate, NewTemplateFromFile)
	groupWith := []string{
		library.TraitsExt,
		library.TraitModifiersExt,
		library.EquipmentExt,
		library.EquipmentModifiersExt,
		library.SkillsExt,
		library.SpellsExt,
		library.NotesExt,
	}
	registerGCSFileInfo("GCS Traits", library.TraitsExt, groupWith, svg.GCSTraits, NewTraitTableDockableFromFile)
	registerGCSFileInfo("GCS Trait Modifiers", library.TraitModifiersExt, groupWith, svg.GCSTraitModifiers, NewTraitModifierTableDockableFromFile)
	registerGCSFileInfo("GCS Equipment", library.EquipmentExt, groupWith, svg.GCSEquipment, NewEquipmentTableDockableFromFile)
	registerGCSFileInfo("GCS Equipment Modifiers", library.EquipmentModifiersExt, groupWith, svg.GCSEquipmentModifiers, NewEquipmentModifierTableDockableFromFile)
	registerGCSFileInfo("GCS Skills", library.SkillsExt, groupWith, svg.GCSSkills, NewSkillTableDockableFromFile)
	registerGCSFileInfo("GCS Spells", library.SpellsExt, groupWith, svg.GCSSpells, NewSpellTableDockableFromFile)
	registerGCSFileInfo("GCS Notes", library.NotesExt, groupWith, svg.GCSNotes, NewNoteTableDockableFromFile)
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
