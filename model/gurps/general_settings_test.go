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
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
)

// TestMonitorPPIForDisplay verifies that deriving the monitor PPI is robust against a missing display. On some Linux
// configurations unison.PrimaryDisplay() can return nil when no monitor is enumerated; previously that caused a nil
// dereference (and, when reached from an unrecovered background goroutine such as the markdown image loader, an
// unlogged process crash). It also guards against a zero content scale, which would otherwise divide by zero.
func TestMonitorPPIForDisplay(t *testing.T) {
	c := check.New(t)

	// A nil display must not panic and must fall back to the default.
	c.Equal(108, monitorPPIForDisplay(nil))

	// A display reporting a zero content scale must not divide by zero; it falls back to the default.
	c.Equal(108, monitorPPIForDisplay(&unison.Display{PPI: 216, Scale: geom.Point{}}))

	// A display that computes a non-positive PPI falls back to the default.
	c.Equal(108, monitorPPIForDisplay(&unison.Display{PPI: 0, Scale: geom.NewPoint(2, 2)}))

	// A normal display yields its scaled PPI.
	c.Equal(108, monitorPPIForDisplay(&unison.Display{PPI: 216, Scale: geom.NewPoint(2, 2)}))
	c.Equal(216, monitorPPIForDisplay(&unison.Display{PPI: 216, Scale: geom.NewPoint(1, 1)}))
}

// TestMonitorPPIUsesSettingOverride verifies that an explicit monitor resolution setting is honored and doesn't touch
// the display at all.
func TestMonitorPPIUsesSettingOverride(t *testing.T) {
	s := &GeneralSettings{MonitorResolution: 150}
	check.New(t).Equal(150, s.MonitorPPI())
}
