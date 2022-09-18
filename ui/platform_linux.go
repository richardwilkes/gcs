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

package ui

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/ui/svglayer"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/formats/icon"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/xio/fs/paths"
)

// See https://developer.gnome.org/documentation/guidelines/maintainer/integrating.html

//go:embed doc-256.png
var docIconBytes []byte

func performPlatformStartup() {
	exePath, err := os.Executable()
	if err != nil {
		jot.Error(errs.Wrap(err))
		return
	}
	if filepath.Base(exePath) != cmdline.AppCmdName {
		jot.Warnf("skipping desktop integration since executable name '%s' is not '%s'", filepath.Base(exePath),
			cmdline.AppCmdName)
		return
	}
	if err = installIcons(); err != nil {
		jot.Error(err)
	}
	if err = installDesktopFiles(exePath); err != nil {
		jot.Error(err)
	}
	if err = installMimeInfo(); err != nil {
		jot.Error(err)
	}
}

func installDesktopFiles(exePath string) error {
	dir := filepath.Join(paths.HomeDir(), ".local", "share", "applications")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return errs.Wrap(err)
	}
	data := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=%s
Comment=%s
Exec=%s %%F
Icon=%s
MimeType=%s;
Categories=Game;Utility;RolePlaying;
Keywords=gurps;character;sheet;rpg;roleplaying;utility;
Terminal=false
`, cmdline.AppName, AppDescription, exePath, cmdline.AppIdentifier, strings.Join(library.RegisteredMimeTypes(), ";"))
	if err := os.WriteFile(filepath.Join(dir, cmdline.AppIdentifier+".desktop"), []byte(data), 0o640); err != nil {
		return errs.Wrap(err)
	}
	return nil
}

func installIcons() error {
	baseDir := filepath.Join(paths.HomeDir(), ".local", "share", "icons", "hicolor", "256x256")
	dir := filepath.Join(baseDir, "apps")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return errs.Wrap(err)
	}
	if err := os.WriteFile(filepath.Join(dir, cmdline.AppIdentifier+".png"), AppIconBytes, 0o640); err != nil {
		return errs.Wrap(err)
	}
	dir = filepath.Join(baseDir, "mimetypes")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return errs.Wrap(err)
	}
	if err := os.WriteFile(filepath.Join(dir, cmdline.AppIdentifier+".png"), AppIconBytes, 0o640); err != nil {
		return errs.Wrap(err)
	}

	docIcon, _, err := image.Decode(bytes.NewBuffer(docIconBytes))
	if err != nil {
		return errs.Wrap(err)
	}
	for i := range library.KnownFileTypes {
		if fi := &library.KnownFileTypes[i]; fi.IsGCSData {
			var overlay image.Image
			overlay, err = svglayer.CreateImageFromSVG(fi, 128)
			if err != nil {
				return err
			}
			targetPath := filepath.Join(dir, strings.ReplaceAll(fi.MimeTypes[0], "/", "-")+".png")
			if err = writePNG(targetPath, icon.Stack(docIcon, overlay)); err != nil {
				return err
			}
		}
	}
	return nil
}

func writePNG(dstPath string, img image.Image) (err error) {
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
	err = errs.Wrap(png.Encode(f, img))
	return
}

func installMimeInfo() error {
	dir := filepath.Join(paths.HomeDir(), ".local", "share", "mime", "packages")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return errs.Wrap(err)
	}
	var buffer bytes.Buffer
	buffer.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<mime-info xmlns='http://www.freedesktop.org/standards/shared-mime-info'>`)
	for i := range library.KnownFileTypes {
		if fi := &library.KnownFileTypes[i]; fi.IsGCSData {
			fmt.Fprintf(&buffer, "  <mime-type type=\"%s\">\n", fi.MimeTypes[0])
			for _, mimeType := range fi.MimeTypes[1:] {
				fmt.Fprintf(&buffer, "    <alias type=\"%s\"/>\n", mimeType)
			}
			for _, ext := range fi.Extensions {
				fmt.Fprintf(&buffer, "    <glob pattern=\"*%s\"/>\n", ext)
			}
			buffer.WriteString("  </mime-type>\n")
		}
	}
	buffer.WriteString(`</mime-info>\n`)
	if err := os.WriteFile(filepath.Join(dir, cmdline.AppIdentifier+".xml"), buffer.Bytes(), 0o640); err != nil {
		return errs.Wrap(err)
	}
	return nil
}
