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
	"github.com/richardwilkes/gcs/v5/runmode"
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
	// Run each registered runmode.Factory now, ahead of flag parsing, since each one registers its own flags as a
	// side effect of being called (see runmode.Factories). Which modes exist, if any beyond the ones this repo
	// itself registers, is entirely up to what got compiled into this build.
	runModes := make([]runmode.Mode, len(runmode.Factories))
	var hiddenFlags []string
	for i, newRunMode := range runmode.Factories {
		runModes[i] = newRunMode(flag.CommandLine)
		hiddenFlags = append(hiddenFlags, runModes[i].HiddenFlagNames...)
	}
	xflag.SetUsage(nil, ux.AppDescription(), i18n.Text("[file]..."), hiddenFlags...)
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

	// Prefer a settings file co-located with the executable over the one in the app data dir.
	settingsName := xos.AppCmdName + "_prefs.json"
	gurps.SettingsPath = filepath.Join(xos.AppDataDir(true), settingsName)
	if dir, err := xos.AppDir(); err == nil {
		settingsPath := filepath.Join(dir, settingsName)
		if xos.FileExists(settingsPath) {
			gurps.SettingsPath = settingsPath
		}
	}
	flag.StringVar(&gurps.SettingsPath, "settings", gurps.SettingsPath, i18n.Text("The `file` to load settings from and store them into"))

	var logCfg xslog.Config
	logCfg.AddFlags()

	xflag.AddVersionFlags()

	xflag.Parse()
	fileList := flag.Args()

	ux.PathToLog = logCfg.RotatorCfg.Path
	xslog.CaptureCrashOutput(logCfg.RotatorCfg)

	ux.RegisterKnownFileTypes()
	gurps.GlobalSettings() // Here to force early initialization

	var requestedRunMode *runmode.Mode
	var requestedRunModeNames []string
	for i := range runModes {
		if runModes[i].Requested() {
			requestedRunMode = &runModes[i]
			requestedRunModeNames = append(requestedRunModeNames, runModes[i].Name)
		}
	}
	if msg := exclusiveModeMsg(requestedRunModeNames); msg != "" {
		xos.ExitWithMsg(msg)
	}

	if requestedRunMode != nil {
		requestedRunMode.Start(fileList) // Never returns
	}
	ux.Start(fileList) // Never returns
}

// exclusiveModeMsg returns a non-empty error message if more than one of the requested run mode names was specified.
// Each run mode takes over the process and exits, so only one may be requested at a time.
func exclusiveModeMsg(requestedModeNames []string) string {
	if len(requestedModeNames) > 1 {
		return fmt.Sprintf(i18n.Text("Cannot specify more than one of -%s"), strings.Join(requestedModeNames, ", -"))
	}
	return ""
}
