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

package main

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/setup/early"
	"github.com/richardwilkes/gcs/v5/ui"
	"github.com/richardwilkes/gcs/v5/ui/svglayer"
	"github.com/richardwilkes/gcs/v5/ui/workspace/external"
	"github.com/richardwilkes/gcs/v5/ui/workspace/lists"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/atexit"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/formats/icon"
	"github.com/richardwilkes/toolbox/formats/icon/icns"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/tc-hib/winres"
	"github.com/tc-hib/winres/version"
)

//go:embed app-1024.png
var appImgBytes []byte
var app image.Image

//go:embed doc-1024.png
var docImgBytes []byte
var doc image.Image

func main() {
	early.Configure()
	jot.FatalIfErr(loadBaseImages())
	external.RegisterFileTypes()
	lists.RegisterFileTypes()
	switch runtime.GOOS {
	case toolbox.MacOS:
		jot.FatalIfErr(packageMacOS())
	case toolbox.WindowsOS:
		jot.FatalIfErr(packageWindows())
	}
	atexit.Exit(0)
}

func loadBaseImages() error {
	var err error
	if app, _, err = image.Decode(bytes.NewBuffer(appImgBytes)); err != nil {
		return errs.Wrap(err)
	}
	if doc, _, err = image.Decode(bytes.NewBuffer(docImgBytes)); err != nil {
		return errs.Wrap(err)
	}
	return nil
}

func packageMacOS() error {
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
	if err := writeICNS(filepath.Join(resDir, "app.icns"), app); err != nil {
		return err
	}
	if err := writeDocICNS(resDir, doc); err != nil {
		return err
	}
	return nil
}

func writeDocICNS(dir string, base image.Image) error {
	for i := range library.KnownFileTypes {
		if fi := &library.KnownFileTypes[i]; fi.IsGCSData {
			overlay, err := svglayer.CreateImageFromSVG(fi, 512)
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

func shortAppVersion() string {
	shortVersion := strings.TrimSuffix(cmdline.AppVersion, ".0")
	if strings.IndexByte(shortVersion, '.') == -1 {
		return cmdline.AppVersion
	}
	return shortVersion
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

func packageWindows() (err error) {
	rs := &winres.ResourceSet{}
	rs.SetManifest(winres.AppManifest{
		Description:    ui.AppDescription,
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

func addWindowsIcon(rs *winres.ResourceSet) error {
	winIcon, err := winres.NewIconFromImages([]image.Image{icon.Scale(app, 256, 256)})
	if err != nil {
		return errs.Wrap(err)
	}
	if err = rs.SetIconTranslation(winres.ID(0), 0, winIcon); err != nil {
		return errs.Wrap(err)
	}
	counter := 1
	for i := range library.KnownFileTypes {
		if fi := &library.KnownFileTypes[i]; fi.IsGCSData {
			var overlay image.Image
			if overlay, err = svglayer.CreateImageFromSVG(fi, 512); err != nil {
				return err
			}
			var extIcon *winres.Icon
			if extIcon, err = winres.NewIconFromImages([]image.Image{icon.Scale(icon.Stack(doc, overlay), 256, 256)}); err != nil {
				return errs.Wrap(err)
			}
			if err = rs.SetIconTranslation(winres.ID(counter), 0, extIcon); err != nil {
				return errs.Wrap(err)
			}
			counter++
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
	if err := vi.Set(version.LangDefault, version.FileDescription, ui.AppDescription); err != nil {
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

var plistTmpl = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleInfoDictionaryVersion</key>
	<string>6.0</string>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>CFBundleName</key>
	<string>{{.AppName}}</string>
	<key>CFBundleDisplayName</key>
	<string>{{.AppName}}</string>
	<key>CFBundleIdentifier</key>
	<string>{{.AppIdentifier}}</string>
	<key>CFBundleVersion</key>
	<string>{{.AppVersion}}</string>
	<key>CFBundleShortVersionString</key>
	<string>{{.ShortVersion}}</string>
    <key>LSMinimumSystemVersion</key>
    <string>{{.MinimumSystemVersion}}</string>
	<key>CFBundleExecutable</key>
	<string>{{.AppCmdName}}</string>
	<key>NSHumanReadableCopyright</key>
	<string>{{.Copyright}}</string>
	<key>CFBundleDevelopmentRegion</key>
	<string>en-US</string>
	<key>CFBundleIconFile</key>
	<string>app.icns</string>
	<key>CFBundleSpokenName</key>
	<string>{{.SpokenName}}</string>
    <key>LSApplicationCategoryType</key>
    <string>{{.CategoryUTI}}</string>
	<key>NSHighResolutionCapable</key>
	<true/>
	<key>NSSupportsAutomaticGraphicsSwitching</key>
	<true/>
    <key>CFBundleDocumentTypes</key>
    <array>
{{- range .FileInfo}}
        <dict>
            <key>CFBundleTypeName</key>
            <string>{{.Name}}</string>
            <key>CFBundleTypeIconFile</key>
            <string>{{.Icon}}</string>
            <key>CFBundleTypeRole</key>
            <string>{{.Role}}</string>
            <key>LSHandlerRank</key>
            <string>{{.Rank}}</string>
            <key>LSItemContentTypes</key>
            <array>
                <string>{{.UTI}}</string>
            </array>
        </dict>
{{- end}}
    </array>
	<key>UTExportedTypeDeclarations</key>
	<array>
{{- range .FileInfo}}
{{- if eq .Rank "Owner"}}
		<dict>
			<key>UTTypeIdentifier</key>
			<string>{{.UTI}}</string>
			<key>UTTypeDescription</key>
			<string>{{.Name}}</string>
			<key>UTTypeIconFile</key>
			<string>{{.Icon}}</string>
			<key>UTTypeConformsTo</key>
			<array>
{{- range .ConformsTo}}
				<string>{{.}}</string>
{{- end}}
			</array>
			<key>UTTypeTagSpecification</key>
			<dict>
				<key>public.filename-extension</key>
				<array>
{{- range .Extensions}}
					<string>{{.}}</string>
{{- end}}
				</array>
				<key>public.mime-type</key>
				<array>
{{- range .MimeTypes}}
					<string>{{.}}</string>
{{- end}}
				</array>
			</dict>
		</dict>
{{- end}}
{{- end}}
	</array>
	<key>UTImportedTypeDeclarations</key>
	<array>
{{- range .FileInfo}}
{{- if ne .Rank "Owner"}}
		<dict>
			<key>UTTypeIdentifier</key>
			<string>{{.UTI}}</string>
			<key>UTTypeDescription</key>
			<string>{{.Name}}</string>
			<key>UTTypeConformsTo</key>
			<array>
{{- range .ConformsTo}}
				<string>{{.}}</string>
{{- end}}
			</array>
			<key>UTTypeTagSpecification</key>
			<dict>
				<key>public.filename-extension</key>
				<array>
{{- range .Extensions}}
					<string>{{.}}</string>
{{- end}}
				</array>
				<key>public.mime-type</key>
				<array>
{{- range .MimeTypes}}
					<string>{{.}}</string>
{{- end}}
				</array>
			</dict>
		</dict>
{{- end}}
{{- end}}
	</array>
</dict>
</plist>
`
