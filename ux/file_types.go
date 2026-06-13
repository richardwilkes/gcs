// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"maps"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xslices"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/imgfmt"
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

func registerSpecialFileInfo(extension string, icon *unison.SVG) {
	dt := uti.Register(&uti.DataType{
		UTI:        "private.gcs.nav" + extension,
		Extensions: []string{extension},
	})
	fi := gurps.FileInfo{
		UTI:       dt,
		SVG:       icon,
		IsSpecial: true,
	}
	fi.Register()
}

// RegisterExternalFileTypes registers the external file types.
func RegisterExternalFileTypes() {
	registerPDFFileInfo()
	registerMarkdownFileInfo()
	groupWith := slices.Sorted(maps.Keys(xslices.Set(imgfmt.AllReadableExtensions())))
	for _, one := range imgfmt.All {
		if one.CanRead() {
			registerImageFileInfo(one, groupWith)
		}
	}
	fi := gurps.FileInfo{
		Name:      "SVG Image",
		UTI:       uti.SVG,
		GroupWith: groupWith,
		SVG:       svg.ImageFile,
		Load:      func(filePath string, _ int) (unison.Dockable, error) { return NewImageDockable(filePath) },
		IsImage:   true,
	}
	fi.Register()
}

func registerImageFileInfo(format imgfmt.Enum, groupWith []string) {
	fi := gurps.FileInfo{
		Name:      format.String() + " Image",
		UTI:       format.UTI(),
		GroupWith: groupWith,
		SVG:       svg.ImageFile,
		Load:      func(filePath string, _ int) (unison.Dockable, error) { return NewImageDockable(filePath) },
		IsImage:   true,
	}
	fi.Register()
}

func registerPDFFileInfo() {
	fi := gurps.FileInfo{
		Name:      "PDF Document",
		UTI:       uti.PDF,
		GroupWith: uti.PDF.Extensions,
		SVG:       svg.PDFFile,
		Load:      NewPDFDockable,
		IsPDF:     true,
	}
	fi.Register()
}

func registerMarkdownFileInfo() {
	fi := gurps.FileInfo{
		Name:             "Markdown Document",
		UTI:              uti.Markdown,
		GroupWith:        uti.Markdown.Extensions,
		SVG:              svg.MarkdownFile,
		IsDeepSearchable: true,
		Load: func(filePath string, _ int) (unison.Dockable, error) {
			return NewMarkdownDockable(filePath, true, false)
		},
	}
	fi.Register()
}

// RegisterGCSFileTypes registers the GCS file types.
func RegisterGCSFileTypes() {
	registerExportableGCSFileInfo("GCS Sheet", gurps.SheetExt, svg.GCSSheet, NewSheetFromFile)
	registerGCSFileInfo("GCS Template", gurps.TemplatesExt, []string{gurps.TemplatesExt}, svg.GCSTemplate,
		NewTemplateFromFile)
	registerGCSFileInfo("GCS Loot", gurps.LootExt, []string{gurps.LootExt}, svg.GCSLoot, NewLootSheetFromFile)
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

func registerGCSFileInfo(name, ext string, groupWith []string, icon *unison.SVG, loader func(filePath string) (unison.Dockable, error)) {
	dt := uti.Register(&uti.DataType{
		UTI:        xos.AppIdentifier + ext,
		Parents:    []*uti.DataType{uti.JSON},
		MimeTypes:  []string{"application/x-gcs-" + ext[1:]},
		Extensions: []string{ext},
	})
	fi := gurps.FileInfo{
		Name:             name,
		UTI:              dt,
		GroupWith:        groupWith,
		SVG:              icon,
		Load:             func(filePath string, _ int) (unison.Dockable, error) { return loader(filePath) },
		IsGCSData:        true,
		IsDeepSearchable: true,
	}
	fi.Register()
}

func registerExportableGCSFileInfo(name, ext string, icon *unison.SVG, loader func(filePath string) (unison.Dockable, error)) {
	dt := uti.Register(&uti.DataType{
		UTI:        xos.AppIdentifier + ext,
		Parents:    []*uti.DataType{uti.JSON},
		MimeTypes:  []string{"application/x-gcs-" + ext[1:]},
		Extensions: []string{ext},
	})
	fi := gurps.FileInfo{
		Name:             name,
		UTI:              dt,
		GroupWith:        []string{ext},
		SVG:              icon,
		Load:             func(filePath string, _ int) (unison.Dockable, error) { return loader(filePath) },
		IsGCSData:        true,
		IsExportable:     true,
		IsDeepSearchable: true,
	}
	fi.Register()
}
