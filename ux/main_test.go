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
	"os"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/unison/enums/mod"
)

// ciScriptExecTimeLimit is the per-script execution time limit, in seconds, the tests run with under CI. It matches
// the one in model/gurps/main_test.go, which explains the choice.
var ciScriptExecTimeLimit = fxp.FromInteger(30)

// TestMain raises the per-script execution time limit for the duration of the tests, exactly as
// model/gurps/main_test.go does and for the same reason: the sheets and templates these tests load resolve scripts as
// they are recalculated, and the production default is small enough that some CI runners cannot always finish even a
// trivial script within it.
//
// It also selects the platform-neutral modifier convention for the whole run, which is the one a headless session
// uses on every host: the menu command key is Control rather than macOS's Command. The actions registerActions builds
// bake mod.OSMenuCommand() into their key bindings, and they are built once per process by whichever test gets there
// first. Left to the host's convention, a plain test registering them on macOS would leave every menu shortcut bound to
// Command, and the headless tests that follow, whose sessions press Control, would never reach the menu items. Pinning
// the convention here means the bindings match whatever the order the tests run in, and the tests that never start a
// session behave the same on every host too.
func TestMain(m *testing.M) {
	limit := gurps.PermittedScriptExecTimeMax
	if os.Getenv("CI") != "" {
		limit = ciScriptExecTimeLimit
	}
	gurps.SetScriptExecTimeLimitForTesting(limit)
	mod.SetPlatformNeutral(true)
	os.Exit(m.Run())
}
