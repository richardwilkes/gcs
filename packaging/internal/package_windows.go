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
	"archive/zip"
	"errors"
	"fmt"
	"image"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/gcs/v5/ux"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/formats/icon"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/tc-hib/winres"
	"github.com/tc-hib/winres/version"
)

func platformPackage() (err error) {
	if len(os.Args) > 1 && os.Args[1] == "-z" {
		return compress()
	}
	rs := &winres.ResourceSet{}
	rs.SetManifest(winres.AppManifest{
		Description:    ux.AppDescription,
		Compatibility:  winres.Win7AndAbove,
		ExecutionLevel: winres.AsInvoker,
		DPIAwareness:   winres.DPIAware,
	})
	if err = addWindowsIcon(rs); err != nil {
		return err
	}
	if err = addWindowsVersion(rs); err != nil {
		return err
	}
	var f *os.File
	f, err = os.Create("rsrc_windows_amd64.syso")
	if err != nil {
		return errs.Wrap(err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = errs.Wrap(cerr)
		}
	}()
	if err = rs.WriteObject(f, winres.ArchAMD64); err != nil {
		return errs.Wrap(err)
	}
	return
}

func compress() (err error) {
	name := fmt.Sprintf("gcs-%s-windows.zip", cmdline.AppVersion)
	if err = os.Remove(name); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return errs.Wrap(err)
	}
	var f *os.File
	f, err = os.Create(name)
	if err != nil {
		return errs.Wrap(err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = errs.Wrap(cerr)
		}
	}()
	zw := zip.NewWriter(f)
	var fw io.Writer
	hdr := &zip.FileHeader{
		Name:     "gcs.exe",
		Method:   zip.Deflate,
		Modified: time.Now(),
	}
	hdr.SetMode(0o755)
	if fw, err = zw.CreateHeader(hdr); err != nil {
		err = errs.Wrap(err)
		return
	}
	var in *os.File
	if in, err = os.Open("gcs.exe"); err != nil {
		err = errs.Wrap(err)
		return
	}
	defer xio.CloseIgnoringErrors(in)
	if _, err = io.Copy(fw, in); err != nil {
		err = errs.Wrap(err)
		return
	}
	err = errs.Wrap(zw.Close())
	return
}

func addWindowsIcon(rs *winres.ResourceSet) error {
	winIcon, err := winres.NewIconFromImages([]image.Image{icon.Scale(appImg, 256, 256)})
	if err != nil {
		return errs.Wrap(err)
	}
	if err = rs.SetIconTranslation(winres.Name("APP"), 0, winIcon); err != nil {
		return errs.Wrap(err)
	}
	for i := range model.KnownFileTypes {
		if fi := &model.KnownFileTypes[i]; fi.IsGCSData {
			var overlay image.Image
			if overlay, err = model.CreateImageFromSVG(fi, 512); err != nil {
				return err
			}
			var extIcon *winres.Icon
			if extIcon, err = winres.NewIconFromImages([]image.Image{icon.Scale(icon.Stack(docImg, overlay), 256, 256)}); err != nil {
				return errs.Wrap(err)
			}
			if err = rs.SetIconTranslation(winres.Name(fi.Extensions[0][1:]), 0, extIcon); err != nil {
				return errs.Wrap(err)
			}
		}
	}
	return nil
}

func addWindowsVersion(rs *winres.ResourceSet) error {
	var vi version.Info
	vi.SetFileVersion(cmdline.AppVersion)
	vi.SetProductVersion(cmdline.AppVersion)
	if err := vi.Set(version.LangDefault, version.CompanyName, cmdline.CopyrightHolder); err != nil {
		return errs.Wrap(err)
	}
	if err := vi.Set(version.LangDefault, version.FileDescription, ux.AppDescription); err != nil {
		return errs.Wrap(err)
	}
	if err := vi.Set(version.LangDefault, version.FileVersion, shortAppVersion()); err != nil {
		return errs.Wrap(err)
	}
	if err := vi.Set(version.LangDefault, version.InternalName, cmdline.AppCmdName); err != nil {
		return errs.Wrap(err)
	}
	if err := vi.Set(version.LangDefault, version.LegalCopyright, cmdline.Copyright()); err != nil {
		return errs.Wrap(err)
	}
	if err := vi.Set(version.LangDefault, version.LegalTrademarks,
		"GURPS is a trademark of Steve Jackson Games, used by permission. All rights reserved."); err != nil {
		return errs.Wrap(err)
	}
	if err := vi.Set(version.LangDefault, version.OriginalFilename, cmdline.AppCmdName+".exe"); err != nil {
		return errs.Wrap(err)
	}
	if err := vi.Set(version.LangDefault, version.ProductName, "GURPS Character Sheet"); err != nil {
		return errs.Wrap(err)
	}
	if err := vi.Set(version.LangDefault, version.ProductVersion, shortAppVersion()); err != nil {
		return errs.Wrap(err)
	}
	rs.SetVersionInfo(vi)
	return nil
}
