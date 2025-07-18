// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/toolbox/v2/ximage"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xyaml"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/cmd/upack/packager"
)

//go:embed doc-1024.png
var docImgBytes []byte

func main() {
	early.Configure()
	ux.RegisterExternalFileTypes()
	ux.RegisterGCSFileTypes()
	const iconsPath = "pkgicons"
	xos.ExitIfErr(fs.WalkDir(os.DirFS(iconsPath), ".", func(path string, d fs.DirEntry, _ error) error {
		if !d.IsDir() && strings.HasSuffix(path, "_doc.png") {
			return os.Remove(filepath.Join(iconsPath, path))
		}
		return nil
	}))

	cfg := packager.Config{
		FullName:        "GURPS Character Sheet",
		ExecutableName:  xos.AppCmdName,
		AppIcon:         "pkgicons/app.png",
		Description:     ux.AppDescription(),
		CopyrightHolder: xos.CopyrightHolder,
		CopyrightYears:  xos.CopyrightYears(),
		Trademarks:      "GURPS is a trademark of Steve Jackson Games, used by permission. All rights reserved.",
		Mac: packager.MacOnlyOpts{
			FinderAppName:             xos.AppName,
			AppID:                     xos.AppIdentifier,
			MinimumSystemVersionAMD64: "10.15",
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
		extensions := make([]string, len(one.Extensions))
		for i, ext := range one.Extensions {
			extensions[i] = ext[1:]
		}
		data := packager.FileData{
			Name:       one.Name,
			Icon:       "pkgicons/" + extensions[0] + "_doc.png",
			Role:       "Viewer",
			Rank:       "Alternate",
			UTI:        one.UTI,
			ConformsTo: one.ConformsTo,
			Extensions: extensions,
			MimeTypes:  one.MimeTypes,
		}
		if one.IsGCSData {
			data.Role = "Editor"
			data.Rank = "Owner"
		}
		cfg.FileInfo = append(cfg.FileInfo, &data)
	}
	xos.ExitIfErr(xyaml.Save("packaging.yml", &cfg))

	// The doc icons use unison's image code to generate the icons, so we need to start it up.
	unison.Start(
		unison.StartupFinishedCallback(func() {
			w, err := unison.NewWindow("")
			xos.ExitIfErr(err)
			w.ToFront()
			unison.InvokeTask(func() {
				var docImg image.Image
				if docImg, _, err = image.Decode(bytes.NewBuffer(docImgBytes)); err != nil {
					xos.ExitIfErr(errs.Wrap(err))
				}
				for i := range gurps.KnownFileTypes {
					fi := gurps.KnownFileTypes[i]
					if !fi.IsGCSData {
						continue
					}
					var overlay image.Image
					overlay, err = svg.CreateImageFromSVG(fi.SVG, 512)
					xos.ExitIfErr(err)
					var f *os.File
					f, err = os.Create(filepath.Join(iconsPath, fi.Extensions[0][1:]+"_doc.png"))
					xos.ExitIfErr(err)
					xos.ExitIfErr(errs.Wrap(png.Encode(f, ximage.Stack(docImg, overlay))))
					xos.ExitIfErr(f.Close())
				}
				w.Dispose() // Will cause the app to quit.
			})
		}),
	)
}
