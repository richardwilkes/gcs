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

package external

import (
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/unison"
)

// RegisterFileTypes registers external file types.
func RegisterFileTypes() {
	registerPDFFileInfo()
	registerMarkdownFileInfo()
	for _, one := range unison.KnownImageFormatExtensions {
		if unison.EncodedImageFormatForPath(one).CanRead() {
			registerImageFileInfo(one)
		}
	}
}

func registerImageFileInfo(ext string) {
	library.FileInfo{
		Extension:             ext,
		ExtensionsToGroupWith: unison.KnownImageFormatExtensions,
		SVG:                   res.ImageFileSVG,
		Load:                  NewImageDockable,
		IsImage:               true,
	}.Register()
}

func registerPDFFileInfo() {
	library.FileInfo{
		Extension:             ".pdf",
		ExtensionsToGroupWith: []string{".pdf"},
		SVG:                   res.PDFFileSVG,
		Load:                  NewPDFDockable,
		IsPDF:                 true,
	}.Register()
}

func registerMarkdownFileInfo() {
	library.FileInfo{
		Extension:             ".md",
		ExtensionsToGroupWith: []string{".md"},
		SVG:                   res.MarkdownFileSVG,
		Load:                  NewMarkdownDockable,
	}.Register()
}
