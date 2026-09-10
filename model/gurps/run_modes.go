// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"flag"
	"fmt"

	"github.com/richardwilkes/gcs/v5/runmode"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xos"
)

func init() {
	runmode.Factories = append(runmode.Factories, newConvertRunMode, newSyncRunMode, newExportSheetsRunMode)
}

// newConvertRunMode registers 'convert' on flag.CommandLine as a side effect of being called, and returns the
// runmode.Mode that converts the files given on the command line to the current file format.
func newConvertRunMode(flagSet *flag.FlagSet) runmode.Mode {
	if flagSet == nil {
		flagSet = flag.CommandLine
	}
	convert := flagSet.Bool("convert", false, i18n.Text("Convert all files specified on the command line to the current data format. If a directory is specified, it will be traversed recursively and all files found will be converted. After all files have been processed, GCS will exit"))
	return runmode.Mode{
		Name:      "convert",
		Requested: func() bool { return *convert },
		Start: func(files []string) {
			if err := Convert(files...); err != nil {
				xos.ExitWithMsg(err.Error())
			}
			xos.Exit(0)
		},
	}
}

// newSyncRunMode registers 'sync' on flag.CommandLine as a side effect of being called, and returns the runmode.Mode
// that syncs the files given on the command line with their library sources.
func newSyncRunMode(flagSet *flag.FlagSet) runmode.Mode {
	if flagSet == nil {
		flagSet = flag.CommandLine
	}
	sync := flagSet.Bool("sync", false, fmt.Sprintf(i18n.Text("Syncs all character sheet (%s), template (%s), and loot (%s) files specified on the command line with their library sources. If a directory is specified, it will be traversed recursively and all files found will be converted. After all files have been processed, GCS will exit"), SheetExt, TemplatesExt, LootExt))
	return runmode.Mode{
		Name:      "sync",
		Requested: func() bool { return *sync },
		Start: func(files []string) {
			if err := SyncToLibraryData(files...); err != nil {
				xos.ExitWithMsg(err.Error())
			}
			xos.Exit(0)
		},
	}
}

// newExportSheetsRunMode registers 'text' on flag.CommandLine as a side effect of being called, and returns the
// runmode.Mode that exports the sheets given on the command line using the specified text template.
func newExportSheetsRunMode(flagSet *flag.FlagSet) runmode.Mode {
	if flagSet == nil {
		flagSet = flag.CommandLine
	}
	textTmplPath := flagSet.String("text", "", i18n.Text("Export sheets using the specified text template `file`"))
	return runmode.Mode{
		Name:      "text",
		Requested: func() bool { return *textTmplPath != "" },
		Start: func(files []string) {
			if len(files) == 0 {
				xos.ExitWithMsg(i18n.Text("No files to process."))
			}
			if err := ExportSheets(*textTmplPath, files); err != nil {
				xos.ExitWithMsg(err.Error())
			}
			xos.Exit(0)
		},
	}
}
