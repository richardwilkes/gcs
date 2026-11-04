// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package uxtest

import (
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestSwapForTest verifies the swap is visible for the duration of the test that asked for it and undone once that test
// finishes.
func TestSwapForTest(t *testing.T) {
	c := check.New(t)
	value := "original"
	t.Run("swapped", func(t *testing.T) {
		SwapForTest(t, &value, "swapped")
		check.New(t).Equal("swapped", value, "the new value is in place while the test runs")
	})
	c.Equal("original", value, "the prior value is back once the test finishes")
}
