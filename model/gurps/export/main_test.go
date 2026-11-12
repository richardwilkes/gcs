// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package export

import (
	"os"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps/gurpstest"
)

// TestMain raises the per-script execution time limit for the tests; see gurpstest.RaiseScriptExecTimeLimit.
func TestMain(m *testing.M) {
	gurpstest.RaiseScriptExecTimeLimit()
	os.Exit(m.Run())
}
