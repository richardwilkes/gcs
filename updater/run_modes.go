// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package updater

import (
	"flag"

	"github.com/richardwilkes/gcs/v5/runmode"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xos"
)

func init() {
	runmode.Factories = append(runmode.Factories, newFinishRunMode)
}

// newFinishRunMode registers the hidden -finish-update flag on flagSet (flag.CommandLine if nil) and returns the
// runmode.Mode that runs Finish. A copy of GCS is started this way once the copy that staged the update has exited,
// since no supported system lets a running application replace itself. Being its own mode, exclusive of every other,
// keeps it from reaching the single-instance handoff ux.StartOptions starts, whose port Finish waits to see released.
func newFinishRunMode(flagSet *flag.FlagSet) runmode.Mode {
	if flagSet == nil {
		flagSet = flag.CommandLine
	}
	statePath := flagSet.String(FinishFlag, "", i18n.Text("Internal use only. Finish applying a previously prepared update, using the state in the specified `file`"))
	return runmode.Mode{
		Name:            FinishFlag,
		HiddenFlagNames: []string{FinishFlag},
		Requested:       func() bool { return *statePath != "" },
		Start: func(_ []string) {
			if err := Finish(*statePath); err != nil {
				errs.Log(err)
				xos.Exit(1)
			}
			xos.Exit(0)
		},
	}
}
