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
	"github.com/richardwilkes/toolbox/v2/geom"
)

// The Description block's bands come from the rows of its first column, which follows the heading. The heading has no
// children, so mistaking it for the first column would draw no bands.
func TestDescriptionPanelBandsSkipTheHeading(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	description := NewDescriptionPanel(entity, NewTargetMgr(newPageUndoRoot()))
	_, pref, _ := description.Sizes(geom.Size{})
	description.SetFrameRect(geom.NewRect(0, 0, pref.Width, pref.Height))
	description.ValidateLayout()
	column := rowsAfterHeading(t, description)[0]
	pairs := len(column.Children()) / 2
	c.True(pairs >= 2, "the first column has at least two label and field pairs, but has %d", pairs)
	bands := description.bandRects()
	c.Equal(pairs/2, len(bands), "every other pair of the first column is banded")
	for i, band := range bands {
		label := column.Children()[2+i*4]
		c.Equal(column.RectTo(label.FrameRect(), description.AsPanel()), band,
			"band %d lies over the label and field pair it belongs to", i)
		c.True(band.Y > 0 && band.Y+band.Height <= pref.Height, "band %d lies within the block, but is %v", i, band)
	}
}
