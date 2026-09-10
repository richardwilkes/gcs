// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package main

import (
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

func TestExclusiveModeMsg(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name         string
		modes        []string
		wantContains []string
		wantErr      bool
	}{
		{
			name: "none specified",
		},
		{
			name:  "one mode",
			modes: []string{"convert"},
		},
		{
			name:         "two modes",
			modes:        []string{"convert", "sync"},
			wantErr:      true,
			wantContains: []string{"-convert", "-sync"},
		},
		{
			name:         "four modes",
			modes:        []string{"convert", "sync", "text", "finish-update"},
			wantErr:      true,
			wantContains: []string{"-convert", "-sync", "-text", "-finish-update"},
		},
		{
			name:         "five modes",
			modes:        []string{"convert", "sync", "text", "finish-update", "headless-api"},
			wantErr:      true,
			wantContains: []string{"-convert", "-sync", "-text", "-finish-update", "-headless-api"},
		},
	} {
		msg := exclusiveModeMsg(tc.modes)
		if tc.wantErr {
			c.NotEqual("", msg, tc.name)
			for _, want := range tc.wantContains {
				c.Contains(msg, want, tc.name)
			}
		} else {
			c.Equal("", msg, tc.name)
		}
	}
}
