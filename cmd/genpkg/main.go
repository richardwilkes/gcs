// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package main

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/richardwilkes/gcs/v5/early"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/gcs/v5/ux"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/toolbox/v2/ximage"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xyaml"
	"github.com/richardwilkes/unison/cmd/upack/packager"
)

const (
	iconsPath     = "pkgicons"
	docIconSuffix = "_doc.png"
	overlaySize   = 512
)

//go:embed doc-1024.png
var docImgBytes []byte

func main() {
	early.Configure()
	ux.RegisterExternalFileTypes()
	ux.RegisterGCSFileTypes()
	if fi, err := os.Stat(iconsPath); err != nil || !fi.IsDir() {
		xos.ExitWithMsg("must be run from the top-level directory of the repository; no " + iconsPath +
			" directory was found")
	}
	xos.ExitIfErr(removeGeneratedDocIcons(iconsPath))
	xos.ExitIfErr(xyaml.Save("packaging.yml", buildConfig()))
	xos.ExitIfErr(generateDocIcons(iconsPath))
}

// generateDocIcons creates a document icon for each of the file types GCS owns by overlaying the file type's icon onto
// the generic document image. The SVGs are rendered by canvas' CPU rasterizer, which needs neither a GPU context nor a
// window, so unison doesn't have to be started up for this.
func generateDocIcons(dir string) error {
	docImg, _, err := image.Decode(bytes.NewBuffer(docImgBytes))
	if err != nil {
		return errs.Wrap(err)
	}
	for _, fi := range gurps.KnownFileTypes {
		if !fi.IsGCSData {
			continue
		}
		var overlay image.Image
		if overlay, err = svg.CreateImageFromSVG(fi.SVG, overlaySize); err != nil {
			return err
		}
		if err = writePNG(filepath.Join(dir, docIconName(fi)), ximage.Stack(docImg, overlay)); err != nil {
			return err
		}
	}
	return nil
}

func writePNG(dstPath string, img image.Image) (err error) {
	var f *os.File
	if f, err = os.Create(dstPath); err != nil {
		return errs.Wrap(err)
	}
	defer func() {
		if closeErr := errs.Wrap(f.Close()); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	return errs.Wrap(png.Encode(f, img))
}

// removeGeneratedDocIcons removes the document icons left over from a previous run, so that only the ones this run
// generates remain. The error passed into the walk function has to be checked, since the directory entry is nil
// whenever that error isn't, which is what happens when dir itself can't be read.
func removeGeneratedDocIcons(dir string) error {
	return fs.WalkDir(os.DirFS(dir), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(p, docIconSuffix) {
			return os.Remove(filepath.Join(dir, p))
		}
		return nil
	})
}

// docIconName returns the name of the document icon file generated for the given file type.
func docIconName(fi *gurps.FileInfo) string {
	return fi.UTI.Extensions[0][1:] + docIconSuffix
}

func buildConfig() *packager.Config {
	cfg := packager.Config{
		FullName:        "GURPS Character Sheet",
		ExecutableName:  xos.AppCmdName,
		AppIcon:         iconsPath + "/app.png",
		Description:     ux.AppDescription(),
		CopyrightHolder: xos.CopyrightHolder,
		CopyrightYears:  xos.CopyrightYears(),
		Trademarks:      "GURPS is a trademark of Steve Jackson Games, used by permission. All rights reserved.",
		Mac: packager.MacOnlyOpts{
			FinderAppName:             xos.AppName,
			AppID:                     xos.AppIdentifier,
			MinimumSystemVersionAMD64: "11.0",
			MinimumSystemVersionARM64: "11.0",
			CategoryUTI:               "public.app-category.role-playing-games",
			CodeSigning: packager.MacCodeSigning{
				Identity:    "Richard Wilkes",
				Credentials: "gcs-notary",
				Options:     []string{"runtime"},
			},
		},
	}
	for _, one := range gurps.KnownFileTypes {
		if one.IsSpecial {
			continue
		}
		extensions := make([]string, len(one.UTI.Extensions))
		for i, ext := range one.UTI.Extensions {
			extensions[i] = ext[1:]
		}
		data := packager.FileData{
			Name:       one.Name,
			Role:       "Viewer",
			Rank:       "Alternate",
			UTI:        one.UTI.UTI,
			ConformsTo: extractConformsTo(one.UTI),
			Extensions: extensions,
			MimeTypes:  one.UTI.MimeTypes,
		}
		if one.IsGCSData {
			// Document icons are only generated for the file types GCS owns, and the packager only consumes the icon
			// of an entry whose Rank is Owner, so the others are deliberately left without one rather than pointing at
			// a file that will never exist.
			data.Icon = iconsPath + "/" + docIconName(one)
			data.Role = "Editor"
			data.Rank = "Owner"
		}
		cfg.FileInfo = append(cfg.FileInfo, &data)
	}
	return &cfg
}

func extractConformsTo(u *uti.DataType) []string {
	list := make([]string, len(u.Parents))
	for i, p := range u.Parents {
		list[i] = p.UTI
	}
	return list
}
