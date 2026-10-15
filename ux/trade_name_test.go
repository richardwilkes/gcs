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
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestTradeNames checks that a suggested transaction name lists every item while they fit, and otherwise lists as many
// as fit and counts the rest, cutting a first name down when even it alone runs too long.
func TestTradeNames(t *testing.T) {
	c := check.New(t)
	c.Equal("", tradeNames(nil))
	c.Equal("Shortsword", tradeNames([]string{"Shortsword"}))
	c.Equal("Shortsword, Dagger", tradeNames([]string{"Shortsword", "Dagger"}))

	// The expectations below are written out for the limit as it stands, so a change to it needs them worked out again.
	if maxTradeNameLength != 50 {
		t.Fatalf("the expectations assume a limit of 50 characters, not %d", maxTradeNameLength)
	}
	names := make([]string, 12)
	for i := range names {
		names[i] = fmt.Sprintf("Item Number %d", i+1)
	}
	c.True(utf8.RuneCountInString(strings.Join(names, ", ")) > maxTradeNameLength, "the names must run too long")
	c.Equal("Item Number 1, Item Number 2 plus 10 other items", tradeNames(names),
		"as many names as fit are listed and the rest counted")
	c.Equal("Item No. 1, Item No. 2 plus 1 other item",
		tradeNames([]string{"Item No. 1", "Item No. 2", "Very Fine Ornate Leather Armor"}),
		"one left over is counted in the singular")
	c.Equal(strings.Join(names[:3], ", "), tradeNames(names[:3]), "names that just fit are all listed")

	long := strings.Repeat("x", maxTradeNameLength+10)
	c.Equal(strings.Repeat("x", maxTradeNameLength-1)+"…", tradeNames([]string{long}),
		"a lone name too long to fit is cut down")
	cut := tradeNames([]string{long, "Dagger", "Rope"})
	c.Equal(strings.Repeat("x", maxTradeNameLength-len(" plus 2 other items")-1)+"… plus 2 other items", cut,
		"a first name too long to fit is cut down to leave room for the count")
	c.True(fitsTradeName(cut))
}
