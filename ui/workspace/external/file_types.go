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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/maps"
)

// RegisterFileTypes registers external file types.
func RegisterFileTypes() {
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
		SVG:        res.ImageFileSVG,
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
		SVG:        res.PDFFileSVG,
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
		SVG:        res.MarkdownFileSVG,
		Load:       NewMarkdownDockable,
	}.Register()
}
