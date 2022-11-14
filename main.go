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
	"github.com/richardwilkes/gcs/v5/early"
	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/gcs/v5/model/convert"
	"github.com/richardwilkes/gcs/v5/model/export"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/ux"
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
	cl.Description = ux.AppDescription
	cl.UsageTrailer = fmt.Sprintf(i18n.Text(`   Settings path: "%s"
Translations dir: "%s"`), model.Path(), i18n.Dir)
	var textTmplPath string
	cl.NewGeneralOption(&textTmplPath).SetName("text").SetSingle('x').SetArg("file").
		SetUsage(i18n.Text("Export sheets using the specified template file"))
	var convertFiles bool
	cl.NewGeneralOption(&convertFiles).SetName("convert").SetSingle('c').
		SetUsage(i18n.Text("Converts all files specified on the command line to the current data format. If a directory is specified, it will be traversed recursively and all files found will be converted. This operation is intended to easily bring files up to the current version's data format. After all files have been processed, GCS will exit"))
	cl.NewGeneralOption(&dbg.VariableResolver).SetName("debug-variable-resolver")
	fileList := jotrotate.ParseAndSetup(cl)
	ux.RegisterKnownFileTypes()
	model.Global() // Here to force early initialization
	unison.DefaultScrollPanelTheme.MouseWheelMultiplier = func() float32 {
		return fxp.As[float32](model.Global().General.ScrollWheelMultiplier)
	}
	switch {
	case convertFiles:
		if err := convert.Convert(fileList...); err != nil {
			cl.FatalMsg(err.Error())
		}
	case textTmplPath != "":
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
	default:
		ux.Start(fileList) // Never returns
	}
	atexit.Exit(0)
}
