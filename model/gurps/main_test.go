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
)

// TestMain raises the per-script execution time limit for the duration of the tests. The production default
// (PermittedScriptExecTimeDef) is intentionally small, but some CI machines are slow enough that legitimate scripts can
// exceed it, resulting in intermittent timeout failures. The tests are not exercising the timeout behavior, so use the
// largest permitted limit here instead. It must stay within the permitted range, since anything outside of it would be
// silently reset to the default by GeneralSettings.EnsureValidity, reintroducing the flakiness this guards against.
func TestMain(m *testing.M) {
	GlobalSettings().General.PermittedPerScriptExecTime = PermittedScriptExecTimeMax
	os.Exit(m.Run())
}
