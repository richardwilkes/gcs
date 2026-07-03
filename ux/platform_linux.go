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
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"image"
	"image/png"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/ximage"
	"github.com/richardwilkes/toolbox/v2/xos"
)

// See https://developer.gnome.org/documentation/guidelines/maintainer/integrating.html

//go:embed images/doc-256.png
var docIconBytes []byte

func performPlatformLateStartup() {
	exePath, err := os.Executable()
	if err != nil {
		errs.Log(err)
		return
	}
	if filepath.Base(exePath) != xos.AppCmdName {
		slog.Warn("skipping desktop integration", "name", filepath.Base(exePath), "expected", xos.AppCmdName)
		return
	}
	if err = installDesktopIcons(); err != nil {
		errs.Log(err)
	}
	if err = installDesktopFiles(exePath); err != nil {
		errs.Log(err)
	}
	if err = installMimeInfo(); err != nil {
		errs.Log(err)
	}
	if err = installExecutableIcon(exePath); err != nil {
		errs.Log(err)
	}
}

// installExecutableIcon attaches the application icon to the executable file itself so that file managers display it
// with our icon rather than the generic "executable" icon. Unlike Windows (PE resource) and macOS (bundle), a Linux ELF
// binary can't embed an icon a file manager will read; instead, GVfs-based file managers (e.g. GNOME Files) honor a
// per-file "custom-icon" metadata attribute, which is what we set here. This is re-applied on every launch, so it also
// fixes itself up if the binary is moved to a new location.
//
// This is GNOME/GVfs-specific. KDE's Dolphin has no equivalent: KIO's KFileItem resolves a custom icon only for
// .desktop files (via their Icon= entry) and directories (via a .directory file), otherwise falling back to the
// MIME-type icon, so the bare executable stays generic there. KDE still picks up the app-menu launcher icon from the
// .desktop file installed by installDesktopFiles.
func installExecutableIcon(exePath string) error {
	cmdPath, err := exec.LookPath("gio")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			slog.Warn("gio not found: skipping executable icon assignment")
			return nil
		}
		return errs.Wrap(err)
	}
	iconPath := filepath.Join(xos.HomeDir(), ".local", "share", "icons", "hicolor", "256x256", "apps",
		xos.AppIdentifier+".png")
	iconURI := (&url.URL{Scheme: "file", Path: iconPath}).String()
	if out, err := exec.Command(cmdPath, "set", exePath, "metadata::custom-icon", iconURI).CombinedOutput(); err != nil {
		return errs.NewWithCause(string(out), err)
	}
	return nil
}

func installDesktopFiles(exePath string) error {
	dir := filepath.Join(xos.HomeDir(), ".local", "share", "applications")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return errs.Wrap(err)
	}
	data := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=%s
Comment=%s
Exec="%s" %%F
Icon=%s
MimeType=%s;
Categories=Game;Utility;RolePlaying;
Keywords=gurps;character;sheet;rpg;roleplaying;utility;
Terminal=false
`, xos.AppName, AppDescription(), exePath, xos.AppIdentifier, strings.Join(gurps.RegisteredMimeTypes(), ";"))
	if err := os.WriteFile(filepath.Join(dir, xos.AppIdentifier+".desktop"), []byte(data), 0o640); err != nil {
		return errs.Wrap(err)
	}
	return nil
}

func installDesktopIcons() error {
	hicolorDir := filepath.Join(xos.HomeDir(), ".local", "share", "icons", "hicolor")
	baseDir := filepath.Join(hicolorDir, "256x256")
	dir := filepath.Join(baseDir, "apps")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return errs.Wrap(err)
	}
	if err := os.WriteFile(filepath.Join(dir, xos.AppIdentifier+".png"), appIconBytes, 0o640); err != nil {
		return errs.Wrap(err)
	}

	dir = filepath.Join(baseDir, "mimetypes")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return errs.Wrap(err)
	}
	docIcon, _, err := image.Decode(bytes.NewBuffer(docIconBytes))
	if err != nil {
		return errs.Wrap(err)
	}
	for i := range gurps.KnownFileTypes {
		fi := gurps.KnownFileTypes[i]
		if !fi.IsGCSData {
			continue
		}
		var overlay image.Image
		overlay, err = svg.CreateImageFromSVG(fi.SVG, 128)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dir, strings.ReplaceAll(fi.UTI.MimeTypes[0], "/", "-")+".png")
		if err = writePNG(targetPath, ximage.Stack(docIcon, overlay)); err != nil {
			return err
		}
	}
	var cmdPath string
	if cmdPath, err = exec.LookPath("gtk-update-icon-cache"); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			errs.Log(errs.NewWithCause("skipping icon cache update", err))
			return nil
		}
		return errs.Wrap(err)
	}
	var out []byte
	if out, err = exec.Command(cmdPath, "--force", "--ignore-theme-index", hicolorDir).CombinedOutput(); err != nil {
		return errs.NewWithCause(string(out), err)
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
	return err
}

func installMimeInfo() error {
	mimeDir := filepath.Join(xos.HomeDir(), ".local", "share", "mime")
	dir := filepath.Join(mimeDir, "packages")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return errs.Wrap(err)
	}
	var buffer bytes.Buffer
	buffer.WriteString("<?xml version='1.0' encoding='UTF-8'?>\n<mime-info xmlns='http://www.freedesktop.org/standards/shared-mime-info'>\n")
	for i := range gurps.KnownFileTypes {
		fi := gurps.KnownFileTypes[i]
		if !fi.IsGCSData {
			continue
		}
		fmt.Fprintf(&buffer, "  <mime-type type=\"%s\">\n", fi.UTI.MimeTypes[0])
		fmt.Fprintf(&buffer, "    <comment>%s</comment>\n", fi.Name)
		for _, mimeType := range fi.UTI.MimeTypes[1:] {
			fmt.Fprintf(&buffer, "    <alias type=\"%s\"/>\n", mimeType)
		}
		for _, ext := range fi.UTI.Extensions {
			fmt.Fprintf(&buffer, "    <glob pattern=\"*%s\"/>\n", ext)
		}
		buffer.WriteString("  </mime-type>\n")
	}
	buffer.WriteString("</mime-info>\n")
	if err := os.WriteFile(filepath.Join(dir, xos.AppIdentifier+".xml"), buffer.Bytes(), 0o640); err != nil {
		return errs.Wrap(err)
	}
	cmdPath, err := exec.LookPath("update-mime-database")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			errs.Log(errs.NewWithCause("skipping mime database update", err))
			return nil
		}
		return errs.Wrap(err)
	}
	var out []byte
	if out, err = exec.Command(cmdPath, mimeDir).CombinedOutput(); err != nil {
		return errs.NewWithCause(string(out), err)
	}
	return nil
}
