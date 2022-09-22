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

package library

import (
	"path"
	"strings"

	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
)

// Some special "extension" values.
const (
	GenericFile  = "file"
	ClosedFolder = ".folder-closed"
	OpenFolder   = ".folder-open"
)

// Known file extensions.
const (
	TraitsExt             = ".adq"
	TraitModifiersExt     = ".adm"
	EquipmentExt          = ".eqp"
	EquipmentModifiersExt = ".eqm"
	SkillsExt             = ".skl"
	SpellsExt             = ".spl"
	NotesExt              = ".not"
	TemplatesExt          = ".gct"
	SheetExt              = ".gcs"
)

// FileInfo contains some static information about a given file type.
type FileInfo struct {
	Name         string
	UTI          string
	ConformsTo   []string
	Extensions   []string
	GroupWith    []string
	MimeTypes    []string
	SVG          *unison.SVG
	Load         func(filePath string) (unison.Dockable, error)
	IsSpecial    bool
	IsGCSData    bool
	IsImage      bool
	IsPDF        bool
	IsExportable bool
}

var (
	// KnownFileTypes holds the registered file types.
	KnownFileTypes   []FileInfo
	fileTypeRegistry = make(map[string]FileInfo)
)

// Register with the central registry.
func (f FileInfo) Register() {
	for _, ext := range f.Extensions {
		fileTypeRegistry[ext] = f
	}
	KnownFileTypes = append(KnownFileTypes, f)
}

// FileInfoFor returns the FileInfo for the given file path's extension.
func FileInfoFor(filePath string) FileInfo {
	if info, ok := fileTypeRegistry[strings.ToLower(path.Ext(filePath))]; ok {
		return info
	}
	return fileTypeRegistry[GenericFile]
}

// AcceptableExtensions returns the file extensions that we should be able to open.
func AcceptableExtensions() []string {
	list := make([]string, 0, len(fileTypeRegistry))
	for k, v := range fileTypeRegistry {
		if !v.IsSpecial {
			list = append(list, k)
		}
	}
	txt.SortStringsNaturalAscending(list)
	return list
}

// GCSExtensions returns the file extensions that are owned by GCS.
func GCSExtensions() []string {
	list := make([]string, 0, len(fileTypeRegistry))
	for k, v := range fileTypeRegistry {
		if v.IsGCSData {
			list = append(list, k)
		}
	}
	txt.SortStringsNaturalAscending(list)
	return list
}

// RegisteredMimeTypes returns the mime types that we should be able to open.
func RegisteredMimeTypes() []string {
	all := make(map[string]bool)
	for _, v := range fileTypeRegistry {
		if !v.IsSpecial {
			for _, one := range v.MimeTypes {
				all[one] = true
			}
		}
	}
	list := make([]string, 0, len(all))
	for k := range all {
		list = append(list, k)
	}
	txt.SortStringsNaturalAscending(list)
	return list
}
