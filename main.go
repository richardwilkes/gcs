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
	"fmt"

	"github.com/richardwilkes/gcs/v5/dbg"
	"github.com/richardwilkes/gcs/v5/model/export"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/gcs/v5/setup"
	"github.com/richardwilkes/gcs/v5/setup/early"
	"github.com/richardwilkes/gcs/v5/ui"
	"github.com/richardwilkes/toolbox/atexit"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jotrotate"
	"github.com/richardwilkes/unison"
)

func main() {
	early.Configure()
	unison.AttachConsole()
	cl := cmdline.New(true)
	cl.Description = ui.AppDescription
	var textTmplPath string
	var showCopyrightDateAndExit bool
	cl.NewGeneralOption(&textTmplPath).SetName("text").SetSingle('x').SetArg("file").
		SetUsage(i18n.Text("Export sheets using the specified template file"))
	cl.NewGeneralOption(&showCopyrightDateAndExit).SetName("copyright-date")
	cl.NewGeneralOption(&dbg.VariableResolver).SetName("debug-variable-resolver")
	fileList := jotrotate.ParseAndSetup(cl)
	if showCopyrightDateAndExit {
		fmt.Print(cmdline.ResolveCopyrightYears())
		atexit.Exit(0)
	}
	setup.Setup()
	settings.Global() // Here to force early initialization
	if textTmplPath != "" {
		if len(fileList) == 0 {
			cl.FatalMsg(i18n.Text("No files to process."))
		}
		for _, one := range fileList {
			if !library.FileInfoFor(one).IsExportable {
				cl.FatalMsg(one + i18n.Text(" is not exportable."))
			}
		}
		if err := export.ToText(textTmplPath, fileList); err != nil {
			cl.FatalMsg(err.Error())
		}
	} else {
		ui.Start(fileList) // Never returns
	}
	atexit.Exit(0)
}
