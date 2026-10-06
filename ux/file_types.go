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
	"sync"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xslices"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/imgfmt"
)

var (
	registerKnownFileTypesOnce sync.Once
	// fileTypeUIs holds what the user interface needs for each registered file type. Like the registry it parallels, it
	// is written by the registration functions and read without synchronization after that, so those must not run again
	// once anything may be reading it.
	fileTypeUIs = make(map[*gurps.FileInfo]fileTypeUI)
)

// fileLoader opens the file at filePath in a new dockable, showing the given page if the file type has pages.
type fileLoader func(filePath string, pageInfo gurps.PageInfo) (unison.Dockable, error)

// fileTypeUI holds the icon for a file type and the function that opens a file of that type, which is nil for the types
// that can't be opened.
type fileTypeUI struct {
	svg  *unison.SVG
	load fileLoader
}

// registerFileInfo adds the file type to the central registry, along with its icon and loader.
func registerFileInfo(fi *gurps.FileInfo, icon *unison.SVG, load fileLoader) {
	fileTypeUIs[fi] = fileTypeUI{svg: icon, load: load}
	fi.Register()
}

// FileTypeSVG returns the icon for the file type.
func FileTypeSVG(fi *gurps.FileInfo) *unison.SVG {
	return fileTypeUIs[fi].svg
}

// RegisterKnownFileTypes registers the known file types. Only the first call registers anything: the registry is read
// without synchronization by the deep search content cache's worker goroutines, so it must not be rewritten once
// populated, and the tests call this wherever they need the registry rather than relying on each other.
func RegisterKnownFileTypes() {
	registerKnownFileTypesOnce.Do(func() {
		registerNavigatorFileTypes()
		RegisterExternalFileTypes()
		RegisterGCSFileTypes()
	})
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
	registerFileInfo(&gurps.FileInfo{
		UTI:       dt,
		IsSpecial: true,
	}, icon, nil)
}

// RegisterExternalFileTypes registers the external file types.
func RegisterExternalFileTypes() {
	registerPDFFileInfo()
	registerMarkdownFileInfo()
	// The imgfmt enum has no SVG member, so add the SVG extensions explicitly, otherwise SVG files won't group with the
	// other image files (or even with each other).
	groupWith := slices.Sorted(maps.Keys(xslices.Set(append(imgfmt.AllReadableExtensions(), uti.SVG.Extensions...))))
	for _, one := range imgfmt.All {
		if one.CanRead() {
			registerImageFileInfo(one, groupWith)
		}
	}
	registerFileInfo(&gurps.FileInfo{
		Name:      "SVG Image",
		UTI:       uti.SVG,
		GroupWith: groupWith,
		IsImage:   true,
	}, svg.ImageFile, loadImageFile)
}

func registerImageFileInfo(format imgfmt.Enum, groupWith []string) {
	registerFileInfo(&gurps.FileInfo{
		Name:      format.String() + " Image",
		UTI:       format.UTI(),
		GroupWith: groupWith,
		IsImage:   true,
	}, svg.ImageFile, loadImageFile)
}

func loadImageFile(filePath string, _ gurps.PageInfo) (unison.Dockable, error) {
	return NewImageDockable(filePath)
}

func registerPDFFileInfo() {
	registerFileInfo(&gurps.FileInfo{
		Name:      "PDF Document",
		UTI:       uti.PDF,
		GroupWith: uti.PDF.Extensions,
		IsPDF:     true,
	}, svg.PDFFile, NewPDFDockable)
}

func registerMarkdownFileInfo() {
	registerFileInfo(&gurps.FileInfo{
		Name:             "Markdown Document",
		UTI:              uti.Markdown,
		GroupWith:        uti.Markdown.Extensions,
		IsDeepSearchable: true,
	}, svg.MarkdownFile, func(filePath string, _ gurps.PageInfo) (unison.Dockable, error) {
		return NewMarkdownDockable(filePath, true, false)
	})
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
	settingsGroupWith := []string{gurps.AncestryExt, gurps.NamesExt}
	registerGCSSettingsFileInfo("GCS Ancestry", gurps.AncestryExt, settingsGroupWith, svg.Ancestry, openAncestryFile)
	registerGCSSettingsFileInfo("GCS Name Generator", gurps.NamesExt, settingsGroupWith, svg.Naming,
		openNameGeneratorFile)
}

// registerGCSSettingsFileInfo registers a settings file type that has an editor of its own, so that such a file can be
// opened from the File menu, the recent files list, the command line or a drop onto the window. Unlike the library
// files, these are not deep searchable: they hold settings, not content.
func registerGCSSettingsFileInfo(name, ext string, groupWith []string, icon *unison.SVG, loader func(filePath string) (unison.Dockable, error)) {
	dt := uti.Register(&uti.DataType{
		UTI:        xos.AppIdentifier + ext,
		Parents:    []*uti.DataType{uti.JSON},
		MimeTypes:  []string{"application/x-gcs-" + ext[1:]},
		Extensions: []string{ext},
	})
	registerFileInfo(&gurps.FileInfo{
		Name:      name,
		UTI:       dt,
		GroupWith: groupWith,
		IsGCSData: true,
	}, icon, func(filePath string, _ gurps.PageInfo) (unison.Dockable, error) { return loader(filePath) })
}

func registerGCSFileInfo(name, ext string, groupWith []string, icon *unison.SVG, loader func(filePath string) (unison.Dockable, error)) {
	dt := uti.Register(&uti.DataType{
		UTI:        xos.AppIdentifier + ext,
		Parents:    []*uti.DataType{uti.JSON},
		MimeTypes:  []string{"application/x-gcs-" + ext[1:]},
		Extensions: []string{ext},
	})
	registerFileInfo(&gurps.FileInfo{
		Name:             name,
		UTI:              dt,
		GroupWith:        groupWith,
		IsGCSData:        true,
		IsDeepSearchable: true,
	}, icon, func(filePath string, _ gurps.PageInfo) (unison.Dockable, error) { return loader(filePath) })
}

func registerExportableGCSFileInfo(name, ext string, icon *unison.SVG, loader func(filePath string) (unison.Dockable, error)) {
	dt := uti.Register(&uti.DataType{
		UTI:        xos.AppIdentifier + ext,
		Parents:    []*uti.DataType{uti.JSON},
		MimeTypes:  []string{"application/x-gcs-" + ext[1:]},
		Extensions: []string{ext},
	})
	registerFileInfo(&gurps.FileInfo{
		Name:             name,
		UTI:              dt,
		GroupWith:        []string{ext},
		IsGCSData:        true,
		IsExportable:     true,
		IsDeepSearchable: true,
	}, icon, func(filePath string, _ gurps.PageInfo) (unison.Dockable, error) { return loader(filePath) })
}
