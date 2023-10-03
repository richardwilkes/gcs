/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/collection/dict"
	"github.com/richardwilkes/unison"
)

// RegisterKnownFileTypes registers the known files types.
func RegisterKnownFileTypes() {
	registerNavigatorFileTypes()
	RegisterExternalFileTypes()
	RegisterGCSFileTypes()
}

func registerNavigatorFileTypes() {
	registerSpecialFileInfo(gurps.ClosedFolder, svg.ClosedFolder)
	registerSpecialFileInfo(gurps.OpenFolder, svg.OpenFolder)
	registerSpecialFileInfo(gurps.GenericFile, svg.GenericFile)
}

func registerSpecialFileInfo(key string, svg *unison.SVG) {
	gurps.FileInfo{
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
	groupWith := dict.Keys(all)
	for _, one := range unison.KnownImageFormatFormats {
		if one.CanRead() {
			registerImageFileInfo(one, groupWith)
		}
	}
}

func registerImageFileInfo(format unison.EncodedImageFormat, groupWith []string) {
	gurps.FileInfo{
		Name:       strings.ToUpper(format.Extension()[1:]) + " Image",
		UTI:        format.UTI(),
		ConformsTo: []string{"public.image"},
		Extensions: format.Extensions(),
		GroupWith:  groupWith,
		MimeTypes:  format.MimeTypes(),
		SVG:        svg.ImageFile,
		Load:       func(filePath string, _ int) (unison.Dockable, error) { return NewImageDockable(filePath) },
		IsImage:    true,
	}.Register()
}

func registerPDFFileInfo() {
	gurps.FileInfo{
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
	extensions := []string{gurps.MarkdownExt, ".markdown"}
	gurps.FileInfo{
		Name:             "Markdown Document",
		UTI:              "net.daringfireball.markdown",
		ConformsTo:       []string{"public.plain-text"},
		Extensions:       extensions,
		GroupWith:        extensions,
		MimeTypes:        []string{"text/markdown"},
		SVG:              svg.MarkdownFile,
		IsDeepSearchable: true,
		Load: func(filePath string, _ int) (unison.Dockable, error) {
			return NewMarkdownDockable(filePath, true, false)
		},
	}.Register()
}

// RegisterGCSFileTypes registers the GCS file types.
func RegisterGCSFileTypes() {
	registerExportableGCSFileInfo("GCS Sheet", gurps.SheetExt, svg.GCSSheet, NewSheetFromFile)
	registerGCSFileInfo("GCS Template", gurps.TemplatesExt, []string{gurps.TemplatesExt}, svg.GCSTemplate,
		NewTemplateFromFile)
	// TODO: Re-enable Campaign files
	// registerGCSFileInfo("GCS Campaign", gurps.CampaignExt, []string{gurps.CampaignExt}, svg.GCSCampaign,
	// 	NewCampaignFromFile)
	groupWith := []string{
		gurps.TraitsExt,
		gurps.TraitModifiersExt,
		gurps.EquipmentExt,
		gurps.EquipmentModifiersExt,
		gurps.SkillsExt,
		gurps.SpellsExt,
		gurps.NotesExt,
	}
	registerGCSFileInfo("GCS Traits", gurps.TraitsExt, groupWith, svg.GCSTraits, NewTraitTableDockableFromFile)
	registerGCSFileInfo("GCS Trait Modifiers", gurps.TraitModifiersExt, groupWith, svg.GCSTraitModifiers,
		NewTraitModifierTableDockableFromFile)
	registerGCSFileInfo("GCS Equipment", gurps.EquipmentExt, groupWith, svg.GCSEquipment,
		NewEquipmentTableDockableFromFile)
	registerGCSFileInfo("GCS Equipment Modifiers", gurps.EquipmentModifiersExt, groupWith, svg.GCSEquipmentModifiers,
		NewEquipmentModifierTableDockableFromFile)
	registerGCSFileInfo("GCS Skills", gurps.SkillsExt, groupWith, svg.GCSSkills, NewSkillTableDockableFromFile)
	registerGCSFileInfo("GCS Spells", gurps.SpellsExt, groupWith, svg.GCSSpells, NewSpellTableDockableFromFile)
	registerGCSFileInfo("GCS Notes", gurps.NotesExt, groupWith, svg.GCSNotes, NewNoteTableDockableFromFile)
}

func registerGCSFileInfo(name, ext string, groupWith []string, svg *unison.SVG, loader func(filePath string) (unison.Dockable, error)) {
	gurps.FileInfo{
		Name:             name,
		UTI:              cmdline.AppIdentifier + ext,
		ConformsTo:       []string{"public.data"},
		Extensions:       []string{ext},
		GroupWith:        groupWith,
		MimeTypes:        []string{"application/x-gcs-" + ext[1:]},
		SVG:              svg,
		Load:             func(filePath string, _ int) (unison.Dockable, error) { return loader(filePath) },
		IsGCSData:        true,
		IsDeepSearchable: true,
	}.Register()
}

func registerExportableGCSFileInfo(name, ext string, svg *unison.SVG, loader func(filePath string) (unison.Dockable, error)) {
	gurps.FileInfo{
		Name:             name,
		UTI:              cmdline.AppIdentifier + ext,
		ConformsTo:       []string{"public.data"},
		Extensions:       []string{ext},
		GroupWith:        []string{ext},
		MimeTypes:        []string{"application/x-gcs-" + ext[1:]},
		SVG:              svg,
		Load:             func(filePath string, _ int) (unison.Dockable, error) { return loader(filePath) },
		IsGCSData:        true,
		IsExportable:     true,
		IsDeepSearchable: true,
	}.Register()
}
