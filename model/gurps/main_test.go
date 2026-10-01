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
	"os"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
)

// ciScriptExecTimeLimit is the per-script execution time limit, in seconds, the tests run with under CI (detected by
// the CI environment variable). Its runners are slow, shared machines, further loaded by go test running several
// packages' test binaries at once, and legitimate scripts have exceeded even PermittedScriptExecTimeMax on them. The
// tests don't need a tight limit, only one that stops a runaway script from stalling the run for long.
var ciScriptExecTimeLimit = fxp.FromInteger(30)

// TestMain raises the per-script execution time limit from the intentionally small production default
// (PermittedScriptExecTimeDef) to the largest limit users may set, or to ciScriptExecTimeLimit under CI, since the
// tests don't exercise the timeout. The override bypasses the settings, so GeneralSettings.EnsureValidity can't reset
// it. ux/main_test.go does the same for the ux tests.
func TestMain(m *testing.M) {
	limit := PermittedScriptExecTimeMax
	if os.Getenv("CI") != "" {
		limit = ciScriptExecTimeLimit
	}
	SetScriptExecTimeLimitForTesting(limit)
	os.Exit(m.Run())
}
