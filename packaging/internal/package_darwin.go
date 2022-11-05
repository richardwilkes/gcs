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

package internal

import (
	_ "embed"
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"text/template"

	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/ux"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/formats/icon"
	"github.com/richardwilkes/toolbox/formats/icon/icns"
)

//go:embed embedded/info.plist.tmpl
var plistTmpl string

func platformPackage() error {
	if err := os.RemoveAll("GCS.app"); err != nil && !errors.Is(err, os.ErrNotExist) {
		return errs.Wrap(err)
	}
	contentsDir := filepath.Join("GCS.app", "Contents")
	exeDir := filepath.Join(contentsDir, "MacOS")
	if err := os.MkdirAll(exeDir, 0o755); err != nil {
		return errs.Wrap(err)
	}
	resDir := filepath.Join(contentsDir, "Resources")
	if err := os.MkdirAll(resDir, 0o755); err != nil {
		return errs.Wrap(err)
	}
	if err := writePlist(filepath.Join(contentsDir, "Info.plist")); err != nil {
		return err
	}
	if err := writeICNS(filepath.Join(resDir, "app.icns"), appImg); err != nil {
		return err
	}
	if err := writeDocICNS(resDir, docImg); err != nil {
		return err
	}
	return nil
}

func writeICNS(dstPath string, img image.Image) (err error) {
	var f *os.File
	f, err = os.Create(dstPath)
	if err != nil {
		return errs.Wrap(err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = errs.Wrap(cerr)
		}
	}()
	err = errs.Wrap(icns.Encode(f, img))
	return
}

func writeDocICNS(dir string, base image.Image) error {
	for i := range library.KnownFileTypes {
		if fi := &library.KnownFileTypes[i]; fi.IsGCSData {
			overlay, err := ux.CreateImageFromSVG(fi, 512)
			if err != nil {
				return err
			}
			targetPath := filepath.Join(dir, fi.Extensions[0][1:]+"_doc.icns")
			if err = writeICNS(targetPath, icon.Stack(base, overlay)); err != nil {
				return err
			}
		}
	}
	return nil
}

func writePlist(targetPath string) (err error) {
	var tmpl *template.Template
	tmpl, err = template.New("plist").Parse(plistTmpl)
	if err != nil {
		return errs.Wrap(err)
	}
	var f *os.File
	if f, err = os.Create(targetPath); err != nil {
		return errs.Wrap(err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = errs.Wrap(cerr)
		}
	}()
	type fileData struct {
		Name       string
		Icon       string
		Role       string
		Rank       string
		UTI        string
		ConformsTo []string
		Extensions []string
		MimeTypes  []string
	}
	fileInfo := make([]*fileData, 0, len(library.KnownFileTypes))
	for _, fi := range library.KnownFileTypes {
		if !fi.IsSpecial {
			extensions := make([]string, len(fi.Extensions))
			for i, ext := range fi.Extensions {
				extensions[i] = ext[1:]
			}
			data := &fileData{
				Name:       fi.Name,
				Icon:       fi.Extensions[0][1:] + "_doc.icns",
				UTI:        fi.UTI,
				ConformsTo: fi.ConformsTo,
				Extensions: extensions,
				MimeTypes:  fi.MimeTypes,
			}
			if fi.IsGCSData {
				data.Role = "Editor"
				data.Rank = "Owner"
			} else {
				data.Role = "Viewer"
				data.Rank = "Alternate"
			}
			fileInfo = append(fileInfo, data)
		}
	}
	type tmplData struct {
		AppName              string
		AppCmdName           string
		SpokenName           string
		AppIdentifier        string
		AppVersion           string
		ShortVersion         string
		MinimumSystemVersion string
		Copyright            string
		CategoryUTI          string
		FileInfo             []*fileData
	}
	if err = tmpl.Execute(f, &tmplData{
		AppName:              cmdline.AppName,
		AppCmdName:           cmdline.AppCmdName,
		SpokenName:           "gurps character sheet",
		AppIdentifier:        cmdline.AppIdentifier,
		AppVersion:           cmdline.AppVersion,
		ShortVersion:         shortAppVersion(),
		MinimumSystemVersion: "10.14",
		Copyright:            fmt.Sprintf("©%s by %s", cmdline.ResolveCopyrightYears(), cmdline.CopyrightHolder),
		CategoryUTI:          "public.app-category.role-playing-games",
		FileInfo:             fileInfo,
	}); err != nil {
		err = errs.Wrap(err)
		return
	}
	return
}
