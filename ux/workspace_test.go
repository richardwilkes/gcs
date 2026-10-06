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
	"encoding/json/jsontext"
	"os"
	"path/filepath"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestDockStateEncoding verifies that a dock state survives being carried by the settings as raw JSON, and that the
// absence of one is preserved.
func TestDockStateEncoding(t *testing.T) {
	c := check.New(t)
	c.Nil(decodeDockState(nil))
	c.Nil(decodeDockState(jsontext.Value("null")))
	c.Equal(0, len(encodeDockState(nil)))

	state := &unison.DockState{
		Type:     unison.ContainerType,
		Children: []*unison.DockState{{Type: unison.DockableType, Key: NavigatorDockKey}},
	}
	decoded := decodeDockState(encodeDockState(state))
	c.NotNil(decoded)
	c.Equal(unison.ContainerType, decoded.Type)
	c.Equal([]string{NavigatorDockKey}, dockStateKeys(decoded))
}

// TestDockStateSurvivesFormatChange verifies that a settings file written when the dock states were fields the model
// decoded itself, rather than raw JSON it carries for the workspace, still loads, and that what is written back for
// such a state is byte for byte what was read.
func TestDockStateSurvivesFormatChange(t *testing.T) {
	c := check.New(t)
	state := &unison.DockState{
		Type:       unison.LayoutType,
		Horizontal: true,
		Divider:    300,
		Children: []*unison.DockState{
			{Type: unison.ContainerType, Children: []*unison.DockState{{Type: unison.DockableType, Key: NavigatorDockKey}}},
			{
				Type:         unison.ContainerType,
				CurrentIndex: 1,
				Children:     []*unison.DockState{{Type: unison.DockableType, Key: "a"}, {Type: unison.DockableType, Key: "b"}},
			},
		},
	}
	var older struct {
		TopDockState *unison.DockState `json:"top_dock_state,omitzero"`
	}
	older.TopDockState = state
	dir := t.TempDir()
	written := filepath.Join(dir, "older.json")
	c.NoError(jio.SaveToFile(written, &older))
	before, err := os.ReadFile(written)
	c.NoError(err)

	var settings gurps.Settings
	c.NoError(jio.LoadFromFile(nil, written, &settings))
	decoded := decodeDockState(settings.TopDockState)
	c.Equal(state, decoded, "the state the older format held is decoded as it was")
	c.Equal([]string{NavigatorDockKey, "a", "b"}, dockStateKeys(decoded))

	var newer struct {
		TopDockState jsontext.Value `json:"top_dock_state,omitzero"`
	}
	newer.TopDockState = settings.TopDockState
	rewritten := filepath.Join(dir, "newer.json")
	c.NoError(jio.SaveToFile(rewritten, &newer))
	after, err := os.ReadFile(rewritten)
	c.NoError(err)
	c.Equal(string(before), string(after), "the state is written back as it was read")
}

// TestResolveDockable verifies that a Dockable is resolved to its outermost implementation, since the placement code is
// handed the embedded SettingsDockable rather than the settings view itself.
func TestResolveDockable(t *testing.T) {
	c := check.New(t)
	d := &pageRefMappingsDockable{}
	d.Self = d
	c.Equal(d, resolveDockable(&d.SettingsDockable), "an inner layer resolves to the Dockable that contains it")
	c.Equal(d, resolveDockable(d), "an already-resolved Dockable is returned unchanged")
	c.Nil(resolveDockable(nil), "a nil Dockable is returned unchanged")
}

// TestResolveDockableRestoresOptionalInterfaces verifies that the optional interfaces a settings view implements are
// visible again once the Dockable has been resolved. Only the settings view implements GroupedCloser; the embedded
// SettingsDockable that gets handed to the placement code does not, so a view given a window of its own used to be
// skipped by traverseGroup and was left open when the sheet that owns it was closed.
func TestResolveDockableRestoresOptionalInterfaces(t *testing.T) {
	c := check.New(t)
	d := &sheetSettingsDockable{}
	d.Self = d
	var placed unison.Dockable = &d.SettingsDockable
	_, ok := placed.(GroupedCloser)
	c.False(ok, "the embedded SettingsDockable is not a GroupedCloser")
	_, ok = resolveDockable(placed).(GroupedCloser)
	c.True(ok, "the resolved settings view is a GroupedCloser")
}
