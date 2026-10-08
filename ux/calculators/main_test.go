// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.
package calculators

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

// TestMain sets the tests up exactly as ux/main_test.go does, and for the same reasons: it raises the per-script
// execution time limit, since the sheets these tests open resolve scripts as they are recalculated, and it pins the
// platform-neutral modifier convention a headless session uses on every host, so that the menu key bindings the
// headless tests press are the same wherever they run.
func TestMain(m *testing.M) {
	limit := gurps.PermittedScriptExecTimeMax
	if os.Getenv("CI") != "" {
		limit = ciScriptExecTimeLimit
	}
	gurps.SetScriptExecTimeLimitForTesting(limit)
	mod.SetPlatformNeutral(true)
	os.Exit(m.Run())
}
