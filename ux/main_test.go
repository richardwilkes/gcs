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
)

// ciScriptExecTimeLimit is the per-script execution time limit, in seconds, the tests run with under CI. It matches
// the one in model/gurps/main_test.go, which explains the choice.
var ciScriptExecTimeLimit = fxp.FromInteger(30)

// TestMain raises the per-script execution time limit for the duration of the tests, exactly as
// model/gurps/main_test.go does and for the same reason: the sheets and templates these tests load resolve scripts as
// they are recalculated, and the production default is small enough that some CI runners cannot always finish even a
// trivial script within it.
func TestMain(m *testing.M) {
	limit := gurps.PermittedScriptExecTimeMax
	if os.Getenv("CI") != "" {
		limit = ciScriptExecTimeLimit
	}
	gurps.SetScriptExecTimeLimitForTesting(limit)
	os.Exit(m.Run())
}
