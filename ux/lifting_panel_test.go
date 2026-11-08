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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestLiftingPanelNamesLiftingAdjustments verifies that the tooltips of the lifting values name the pool thresholds
// adjusting the lifting ST they are figured from, and that Sync rebuilds them as the thresholds come and go.
func TestLiftingPanelNamesLiftingAdjustments(t *testing.T) {
	c := check.New(t)
	e := tiredEntity(t)
	panel := NewLiftingPanel(e)
	c.NotEqual(0, len(panel.fields))
	for _, field := range panel.fields {
		c.Contains(tooltipText(field.Tooltip), "Adjusted by:\nTired (FP) [×0.5]")
	}

	e.Attributes.Set[gurps.FatiguePointsID].Damage = 0
	e.Recalculate()
	DeepSync(panel)
	for _, field := range panel.fields {
		c.NotContains(tooltipText(field.Tooltip), "Tired", "rested, lifting ST is not adjusted")
	}
}
