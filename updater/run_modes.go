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

// newFinishRunMode registers the hidden -finish-update flag on flag.CommandLine as a side effect of being called,
// and returns the runmode.Mode that finishes applying a staged update. Not meant to be typed by anyone: a copy of
// GCS is started this way to finish applying an update once the copy that prepared it has exited, since replacing a
// running application from within itself is not something any of the supported systems allow. Its Start must never
// reach ux.Start's single-instance handoff protocol. Waiting for that protocol's port to be released is this mode's
// whole job. Running as its own runmode.Mode, mutually exclusive with every other mode, guarantees this outcome.
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
