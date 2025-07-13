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
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestWeaponBlock(t *testing.T) {
	c := check.New(t)
	for i, s := range []string{
		"0",
		"-1",
		"10",
		"No",
	} {
		c.Equal(s, gurps.ParseWeaponBlock(s).String(), "test %d", i)
	}

	cases := []struct {
		input    string
		expected string
	}{
		{"", "No"},
		{"-", "No"},
		{"+0", "0"},
		{"+1", "1"},
	}
	for i, one := range cases {
		c.Equal(one.expected, gurps.ParseWeaponBlock(one.input).String(), "test %d", i)
	}
}
