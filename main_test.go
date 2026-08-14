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
		name             string
		textTmplPath     string
		finishUpdatePath string
		wantContains     []string
		convert          bool
		sync             bool
		wantErr          bool
	}{
		{name: "none specified"},
		{name: "convert only", convert: true},
		{name: "sync only", sync: true},
		{name: "text only", textTmplPath: "tmpl"},
		{name: "finish-update only", finishUpdatePath: "state.json"},
		{name: "convert and sync", convert: true, sync: true, wantErr: true, wantContains: []string{"--convert", "--sync"}},
		{name: "convert and text", convert: true, textTmplPath: "tmpl", wantErr: true, wantContains: []string{"--convert", "--text"}},
		{name: "sync and text", sync: true, textTmplPath: "tmpl", wantErr: true, wantContains: []string{"--sync", "--text"}},
		{name: "finish-update and convert", convert: true, finishUpdatePath: "state.json", wantErr: true, wantContains: []string{"--convert", "--finish-update"}},
		{name: "finish-update and text", textTmplPath: "tmpl", finishUpdatePath: "state.json", wantErr: true, wantContains: []string{"--text", "--finish-update"}},
		{name: "all three", convert: true, sync: true, textTmplPath: "tmpl", wantErr: true, wantContains: []string{"--convert", "--sync", "--text"}},
		{name: "all four", convert: true, sync: true, textTmplPath: "tmpl", finishUpdatePath: "state.json", wantErr: true, wantContains: []string{"--convert", "--sync", "--text", "--finish-update"}},
	} {
		msg := exclusiveModeMsg(tc.convert, tc.sync, tc.textTmplPath, tc.finishUpdatePath)
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
