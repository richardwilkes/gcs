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

// TestFilterBannerFollowsFilter verifies that the banner is put between the toolbar and the table whenever a filter is
// applied, whether by the quick filter or a saved one, and taken away again as soon as the list shows everything.
func TestFilterBannerFollowsFilter(t *testing.T) {
	c := check.New(t)
	d := newFilterTestTraitDockable(t, newNameContainsFilter("Vision", "Vision"))
	c.Equal(-1, d.IndexOfChild(d.filterBanner), "an unfiltered list has no banner")

	d.filterField.SetText("fur")
	c.True(d.table.IsFiltered(), "typing in the quick filter must filter the table")
	c.Equal(1, d.IndexOfChild(d.filterBanner), "the banner sits right after the toolbar")
	c.Equal(2, d.IndexOfChild(d.scroll), "and right before the table")

	d.filterField.SetText("")
	c.False(d.table.IsFiltered(), "emptying the quick filter must show everything")
	c.Equal(-1, d.IndexOfChild(d.filterBanner), "which takes the banner away")

	d.chooseFilter(gurps.GlobalSettings().ListFilters[gurps.ListFilterKeyForExtension(gurps.TraitsExt)][0])
	c.True(d.table.IsFiltered(), "the saved filter must filter the table")
	c.Equal(1, d.IndexOfChild(d.filterBanner), "the banner is back for a saved filter")

	d.chooseFilter(nil)
	c.False(d.table.IsFiltered(), "dropping the saved filter must show everything")
	c.Equal(-1, d.IndexOfChild(d.filterBanner), "which takes the banner away again")
}
