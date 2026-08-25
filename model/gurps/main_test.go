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

// ciScriptExecTimeLimit is the per-script execution time limit, in seconds, the tests run with under CI (which every
// hosted runner announces by setting CI in the environment). The runners are slow, shared machines that go test loads
// further by running several packages' test binaries at once, and legitimate scripts have been seen to exceed even
// PermittedScriptExecTimeMax on them. Nothing in the tests depends on the limit being tight, so it only needs to be
// small enough that a script which really has run away cannot stall the run for long.
var ciScriptExecTimeLimit = fxp.FromInteger(30)

// TestMain raises the per-script execution time limit for the duration of the tests. The production default
// (PermittedScriptExecTimeDef) is intentionally small and the tests are not exercising the timeout behavior, so they
// run with the largest limit users are permitted to set -- or, under CI, with ciScriptExecTimeLimit. The override
// bypasses the settings entirely, so GeneralSettings.EnsureValidity cannot silently reset it to the default.
// ux/main_test.go does the same for the ux tests, whose sheets resolve scripts as they are recalculated.
func TestMain(m *testing.M) {
	limit := PermittedScriptExecTimeMax
	if os.Getenv("CI") != "" {
		limit = ciScriptExecTimeLimit
	}
	SetScriptExecTimeLimitForTesting(limit)
	os.Exit(m.Run())
}
