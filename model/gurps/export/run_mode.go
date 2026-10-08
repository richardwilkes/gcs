// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package export

import (
	"flag"

	"github.com/richardwilkes/gcs/v5/runmode"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xos"
)

func init() {
	runmode.Factories = append(runmode.Factories, newExportSheetsRunMode)
}

// newExportSheetsRunMode registers 'text' on flagSet as a side effect of being called, and returns the
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
			if err := Sheets(*textTmplPath, files); err != nil {
				xos.ExitWithMsg(err.Error())
			}
			xos.Exit(0)
		},
	}
}
