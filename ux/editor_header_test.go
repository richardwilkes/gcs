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
	"github.com/richardwilkes/unison"
)

// TestGroupContainersFirst verifies that a column's own comparison is handed the value a cell shows rather than the
// marker a container's sort text carries, and that containers still group ahead of everything else.
func TestGroupContainersFirst(t *testing.T) {
	c := check.New(t)

	less := groupContainersFirst(gurps.PointsLessFromString)
	c.NotNil(less, "a column with a comparison of its own keeps one")
	c.Nil(groupContainersFirst(nil), "a column with no comparison of its own is left without one")

	c.True(less("9", "10"), "ordinary rows compare by value")
	c.True(less(containerMarker+"100", "9"), "containers group ahead of everything else")
	c.False(less("9", containerMarker+"100"), "containers group ahead of everything else, from either side")
	c.True(less(containerMarker+"9", containerMarker+"10"), "containers compare by value among themselves")
	c.True(less(containerMarker+"9~20", containerMarker+"10"), "a container showing a range compares by its low end")
}

// TestPageTableColumnHeaderClickSorts verifies that clicking a column header of a sheet's list sorts the list by that
// column, ascending and then, on a second click, descending.
func TestPageTableColumnHeaderClickSorts(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	entity.Traits = nil
	for _, name := range []string{"Beta", "Alpha", "Gamma"} {
		trait := gurps.NewTrait(entity, nil, false)
		trait.Name = name
		entity.Traits = append(entity.Traits, trait)
	}
	sheet.Rebuild(true)
	list := sheet.Traits
	names := func() []string {
		rows := list.Table.RootRows()
		result := make([]string, 0, len(rows))
		for _, row := range rows {
			result = append(result, row.Data().Name)
		}
		return result
	}
	c.Equal([]string{"Beta", "Alpha", "Gamma"}, names(), "precondition: the list starts unsorted")
	header := list.tableHeader
	where := header.ColumnFrame(list.Table.ColumnIndexForID(gurps.TraitDescriptionColumn)).Center()
	click := func() {
		header.MouseDownCallback(where, unison.ButtonLeft, 1, 0)
		header.MouseUpCallback(where, unison.ButtonLeft, 0)
	}
	click()
	c.Equal([]string{"Alpha", "Beta", "Gamma"}, names(), "a click sorts the list ascending")
	click()
	c.Equal([]string{"Gamma", "Beta", "Alpha"}, names(), "and another descending")
}
