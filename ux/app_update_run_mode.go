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
	"flag"

	"github.com/richardwilkes/gcs/v5/runmode"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/unison/updater"
)

func init() {
	runmode.Factories = append(runmode.Factories, newFinishRunMode)
}

// newFinishRunMode registers the hidden -finish-update flag on flagSet (flag.CommandLine if nil) and returns the
// runmode.Mode that finishes applying a prepared update. The copy of GCS that staged the update starts another this way
// as it exits, and that helper waits for it to be gone before swapping in the new version, since no supported system
// lets a running application replace itself. Being its own mode, exclusive of every other, keeps it from reaching the
// single-instance handoff StartOptions starts, whose port the helper waits to see released.
func newFinishRunMode(flagSet *flag.FlagSet) runmode.Mode {
	if flagSet == nil {
		flagSet = flag.CommandLine
	}
	statePath := flagSet.String(updater.FinishFlag, "", i18n.Text("Internal use only. Finish applying a previously prepared update, using the state in the specified `file`"))
	return runmode.Mode{
		Name:            updater.FinishFlag,
		HiddenFlagNames: []string{updater.FinishFlag},
		Requested:       func() bool { return *statePath != "" },
		Start: func(_ []string) {
			if err := appUpdater().Finish(*statePath); err != nil {
				errs.Log(err)
				xos.Exit(1)
			}
			xos.Exit(0)
		},
	}
}
