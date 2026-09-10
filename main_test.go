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
	"flag"
	"io"
	"testing"

	"github.com/richardwilkes/gcs/v5/runmode"
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

// TestRunModeFactories checks the wiring every registered runmode.Factory is expected to provide: it registers a flag
// named after the Mode it returns, that Mode is not requested until that flag is given, and it is requested once it
// is. Each factory is called with its own flag set, so this exercises exactly what main does with flag.CommandLine
// without disturbing it. Which modes take part depends on what got compiled into this build (see runmode.Factories).
func TestRunModeFactories(t *testing.T) {
	c := check.New(t)
	c.NotEqual(0, len(runmode.Factories))
	seen := make(map[string]bool, len(runmode.Factories))
	for _, factory := range runmode.Factories {
		flagSet := flag.NewFlagSet("gcs", flag.ContinueOnError)
		flagSet.SetOutput(io.Discard)
		m := factory(flagSet)
		c.NotEqual("", m.Name)
		c.False(seen[m.Name], m.Name)
		seen[m.Name] = true
		c.NotNil(m.Requested, m.Name)
		c.NotNil(m.Start, m.Name)
		f := flagSet.Lookup(m.Name)
		c.NotNil(f, m.Name)
		if f == nil {
			continue
		}
		for _, hidden := range m.HiddenFlagNames {
			c.NotNil(flagSet.Lookup(hidden), m.Name, hidden)
		}
		c.False(m.Requested(), m.Name)
		c.NoError(flagSet.Parse([]string{"-" + m.Name + "=" + requestValueFor(f)}), m.Name)
		c.True(m.Requested(), m.Name)
	}
}

// requestValueFor returns a value for f that asks for the run mode it belongs to: bool flags are simply turned on,
// while the rest -- the ones that name a file or an address -- are considered requested when they are non-empty.
func requestValueFor(f *flag.Flag) string {
	if b, ok := f.Value.(interface{ IsBoolFlag() bool }); ok && b.IsBoolFlag() {
		return "true"
	}
	return "requested"
}
