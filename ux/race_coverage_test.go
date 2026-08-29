// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

//go:build race

package ux

import "testing"

// TestRace re-runs the tests that put two or more goroutines over shared state, which is the only situation the race
// detector can ever report on. build.sh runs the full suite uninstrumented and then runs only this wrapper under the
// race detector; see model/gurps/race_coverage_test.go for the full rationale. When adding a test that involves
// concurrency, list it here as well so the race pass covers it.
func TestRace(t *testing.T) {
	t.Run("BuildContentCacheReuseAndFailureCaching", TestBuildContentCacheReuseAndFailureCaching)
	t.Run("HandoffRefusesOversizedPayload", TestHandoffRefusesOversizedPayload)
	t.Run("HandoffRoundTrip", TestHandoffRoundTrip)
	t.Run("SupersededCacheBuildHandsEntriesToItsSuccessor", TestSupersededCacheBuildHandsEntriesToItsSuccessor)
}
