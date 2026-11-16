// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package attack_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps/attack"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestRecoil(t *testing.T) {
	parse := func(s string) string { return attack.ParseRecoil(s).String() }
	same := []string{
		"",
		"1",
		"1/3",
	}
	adjusted := []parseCase{
		{"-", ""},
		{"?", ""},
		{"0", ""},
		{"-2", ""},
		{"1/-3", "1"},
	}
	checkFieldParsing(check.New(t), parse, same, adjusted)
}

// TestRecoilTooltip verifies that the note appears only when the shot and slug recoils are both present and differ.
func TestRecoilTooltip(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		recoil string
		want   string
	}{
		{recoil: "1/3", want: "First Recoil value is for shot, second is for slugs"},
		{recoil: "2/2"},
		{recoil: "2"},
		{recoil: "0/3"},
	} {
		c.Equal(tc.want, attack.ParseRecoil(tc.recoil).Tooltip(), "recoil %q", tc.recoil)
	}
}
