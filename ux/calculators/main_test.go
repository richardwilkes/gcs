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
	"testing"

	"github.com/richardwilkes/gcs/v5/ux"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/unison"
)

// TestMain hands uxtest the workspace these tests drive; see uxtest.Main for the rest of what it does.
func TestMain(m *testing.M) {
	uxtest.Main(m, uxtest.Workspace{
		Setup: func(t *testing.T, wnd *unison.Window) {
			uxtest.SwapForTest(t, &ux.Workspace, ux.Workspace) // The session replaces most of it; put all of it back.
			ux.RegisterKnownFileTypes()
			ux.RegisterWindowDragTypes(wnd)
			ux.SetupMenuBar(wnd)
			ux.InitWorkspace(wnd)
			ux.Workspace.ErrorHandler = func(msg string, err error) { t.Errorf("unexpected error: %s: %v", msg, err) }
		},
		AllDockables: ux.AllDockables,
	})
}
