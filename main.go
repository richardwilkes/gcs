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
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/richardwilkes/gcs/v5/early"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xflag"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xslog"
	"github.com/richardwilkes/toolbox/v2/xterm"
	"github.com/richardwilkes/unison"
)

func main() {
	early.Configure()
	ux.LoadLanguageSetting()
	unison.AttachConsole()
	xflag.SetUsage(nil, ux.AppDescription(), i18n.Text("[file]..."))
	savedUsage := flag.CommandLine.Usage
	flag.CommandLine.Usage = func() {
		savedUsage()
		var w *xterm.AnsiWriter
		switch out := flag.CommandLine.Output().(type) {
		case *xterm.AnsiWriter:
			w = out
		default:
			w = xterm.NewAnsiWriter(out)
		}
		w.WriteString(i18n.Text("Translations dir: "))
		w.Blue()
		w.WriteString(i18n.Dir)
		w.Reset()
		w.WriteByte('\n')
	}

	// Look for a settings file co-located with the executable and prefer that over the one in the app data dir.
	settingsName := xos.AppCmdName + "_prefs.json"
	gurps.SettingsPath = filepath.Join(xos.AppDataDir(true), settingsName)
	if dir, err := xos.AppDir(); err == nil {
		settingsPath := filepath.Join(dir, settingsName)
		if xos.FileExists(settingsPath) {
			gurps.SettingsPath = settingsPath
		}
	}
	flag.StringVar(&gurps.SettingsPath, "settings", gurps.SettingsPath, i18n.Text("The `file` to load settings from and store them into"))

	textTmplPath := flag.String("text", "", i18n.Text("Export sheets using the specified text template `file`"))

	convertFiles := flag.Bool("convert", false, i18n.Text("Convert all files specified on the command line to the current data format. If a directory is specified, it will be traversed recursively and all files found will be converted. After all files have been processed, GCS will exit"))

	syncToLibraryData := flag.Bool("sync", false, fmt.Sprintf(i18n.Text("Syncs all character sheet (%s), template (%s), and loot (%s) files specified on the command line with their library sources. If a directory is specified, it will be traversed recursively and all files found will be converted. After all files have been processed, GCS will exit"), gurps.SheetExt, gurps.TemplatesExt, gurps.LootExt))

	var logCfg xslog.Config
	logCfg.AddFlags()

	xflag.AddVersionFlags()

	xflag.Parse()
	fileList := flag.Args()

	ux.PathToLog = logCfg.RotatorCfg.Path

	ux.RegisterKnownFileTypes()
	gurps.GlobalSettings() // Here to force early initialization

	if msg := exclusiveModeMsg(*convertFiles, *syncToLibraryData, *textTmplPath); msg != "" {
		xos.ExitWithMsg(msg)
	}

	switch {
	case *convertFiles:
		if err := gurps.Convert(fileList...); err != nil {
			xos.ExitWithMsg(err.Error())
		}
	case *syncToLibraryData:
		if err := gurps.SyncToLibraryData(fileList...); err != nil {
			xos.ExitWithMsg(err.Error())
		}
	case *textTmplPath != "":
		if len(fileList) == 0 {
			xos.ExitWithMsg(i18n.Text("No files to process."))
		}
		if err := gurps.ExportSheets(*textTmplPath, fileList); err != nil {
			xos.ExitWithMsg(err.Error())
		}
	default:
		ux.Start(fileList) // Never returns
	}
	xos.Exit(0)
}

// exclusiveModeMsg returns a non-empty error message if more than one of the mutually exclusive command-line modes
// (--convert, --sync, --text) has been specified. These modes each take over the process and exit, so only one may be
// requested at a time.
func exclusiveModeMsg(convert, sync bool, textTmplPath string) string {
	var modes []string
	if convert {
		modes = append(modes, "--convert")
	}
	if sync {
		modes = append(modes, "--sync")
	}
	if textTmplPath != "" {
		modes = append(modes, "--text")
	}
	if len(modes) > 1 {
		return fmt.Sprintf(i18n.Text("Cannot specify more than one of %s"), strings.Join(modes, ", "))
	}
	return ""
}
