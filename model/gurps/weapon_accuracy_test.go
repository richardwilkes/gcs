// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/check"
)

func TestWeaponAccuracy(t *testing.T) {
	for i, s := range []string{
		"0",
		"1",
		"1+3",
		"0+3",
		"Jet",
	} {
		check.Equal(t, s, gurps.ParseWeaponAccuracy(s).String(), "test %d", i)
	}

	cases := []struct {
		input    string
		expected string
	}{
		{"", "0"},
		{"-", "0"},
		{"?", "0"},
		{"+0", "0"},
		{"+1", "1"},
		{"+0+3", "0+3"},
		{"+1+3", "1+3"},
		{"+1+0", "1"},
		{"1+0", "1"},
		{"51,", "51"},
	}
	for i, c := range cases {
		check.Equal(t, c.expected, gurps.ParseWeaponAccuracy(c.input).String(), "test %d", i)
	}
}
