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

func TestWeaponBulk(t *testing.T) {
	c := check.New(t)
	for i, s := range []string{
		"",
		"-1",
		"-10",
		"-11/-14",
		"-3*",
	} {
		c.Equal(s, gurps.ParseWeaponBulk(s).String(), "test %d", i)
	}

	cases := []struct {
		input    string
		expected string
	}{
		{"-", ""},
		{"–", ""},
		{"?", ""},
		{"0", ""},
	}
	for i, one := range cases {
		c.Equal(one.expected, gurps.ParseWeaponBulk(one.input).String(), "test %d", i)
	}
}
