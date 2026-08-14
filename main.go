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
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/richardwilkes/gcs/v5/early"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux"
	"github.com/richardwilkes/toolbox/v2/errs"
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
	captureCrashOutput(logCfg.RotatorCfg)

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

// captureCrashOutput arranges for the Go runtime's crash report -- the message and goroutine dump an unrecovered panic
// or a fatal runtime error produces -- to be written to a file beside the log.
//
// That report otherwise only goes to standard error, which is discarded for an application launched from the Finder or
// its equivalent, and the log receives nothing because the process dies before any of it reaches the logging handler.
// The result is a crash report from a user that no one can act on. The runtime writes here in addition to standard
// error rather than instead of it, so running from a terminal is unaffected.
//
// The file is appended to, so a crash is still recoverable if the application is relaunched before anyone collects it,
// and it is only ever written when the process actually dies, so it stays absent or small in normal use. It is rotated
// on the same terms as the log; see rotateCrashOutput for why that happens here rather than as it is written.
func captureCrashOutput(cfg xslog.Rotator) {
	cfg.Normalize()
	if cfg.Path == "" {
		return
	}
	// Both the file and its backups are named from this base, matching how the log's rotating writer derives its own.
	base := strings.TrimSuffix(cfg.Path, xslog.LogFileExt) + "-crash"
	path := base + xslog.LogFileExt
	if err := os.MkdirAll(filepath.Dir(path), cfg.DirMode); err != nil {
		errs.Log(errs.NewWithCause("unable to create the directory for the crash output file", err), "path", path)
		return
	}
	rotateCrashOutput(base, cfg)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, cfg.FileMode)
	if err != nil {
		errs.Log(errs.NewWithCause("unable to create the crash output file", err), "path", path)
		return
	}
	// SetCrashOutput duplicates the descriptor, so ours is no longer needed once it returns.
	defer func() {
		if err = f.Close(); err != nil {
			errs.Log(errs.NewWithCause("unable to close the crash output file", err), "path", path)
		}
	}()
	if err = debug.SetCrashOutput(f, debug.CrashOptions{}); err != nil {
		errs.Log(errs.NewWithCause("unable to direct crash output to a file", err), "path", path)
	}
}

// crashOutputPathFor returns the crash output file itself for n <= 0, or its n-th backup, using the same naming the
// log's rotating writer uses for its own backups.
func crashOutputPathFor(base string, n int) string {
	if n <= 0 {
		return base + xslog.LogFileExt
	}
	return fmt.Sprintf("%s-%d%s", base, n, xslog.LogFileExt)
}

// rotateCrashOutput retires the crash output file, and prunes its backups, once it has reached the size the log's
// rotation is configured with, so that repeated crashes cannot grow it without bound.
//
// This runs when the file is opened rather than when it is written, because there is no opportunity to run Go code at
// the point it is written: the runtime writes the crash report directly to the file descriptor while the process is
// dying. Rotating here is enough, since a single run can only ever append one report -- the process does not survive
// it -- so the file can only exceed the limit between runs, which is exactly when this is called.
func rotateCrashOutput(base string, cfg xslog.Rotator) {
	fi, err := os.Stat(crashOutputPathFor(base, 0))
	if err != nil || fi.Size() < cfg.MaxSize {
		return
	}
	// Remove the oldest retained backup along with any higher-numbered ones left behind by a previous run configured to
	// retain more. Backups are always consecutive, so the first missing slot ends the sweep.
	for n := cfg.MaxBackups; ; n++ {
		if err = os.Remove(crashOutputPathFor(base, n)); err != nil {
			if !os.IsNotExist(err) {
				errs.Log(errs.NewWithCause("unable to prune a crash output backup", err), "path",
					crashOutputPathFor(base, n))
			}
			break
		}
	}
	for i := cfg.MaxBackups; i > 0; i-- {
		if err = os.Rename(crashOutputPathFor(base, i-1), crashOutputPathFor(base, i)); err != nil &&
			!os.IsNotExist(err) {
			errs.Log(errs.NewWithCause("unable to rotate the crash output file", err), "path",
				crashOutputPathFor(base, i-1))
			return
		}
	}
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
