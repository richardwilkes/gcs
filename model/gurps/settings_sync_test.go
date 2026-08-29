// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"fmt"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestScriptExecTimeLimitSyncIsConcurrencySafe verifies that scriptExecTimeLimit can be called from a background
// goroutine while the UI thread rewrites the general settings and republishes the limit, the pattern the deep search
// content cache produces. The reader must observe only published values, never the struct field being written.
func TestScriptExecTimeLimitSyncIsConcurrencySafe(t *testing.T) {
	c := check.New(t)
	prevOverride := scriptExecTimeLimitOverride.Load()
	scriptExecTimeLimitOverride.Store(0)
	defer scriptExecTimeLimitOverride.Store(prevOverride)
	general := GlobalSettings().General
	prevLimit := general.PermittedPerScriptExecTime
	defer func() {
		general.PermittedPerScriptExecTime = prevLimit
		SyncScriptExecTimeLimit()
	}()

	// The values the writer publishes all lie within the permitted range, so that the validation the settings dialog
	// applies would leave each of them alone rather than quietly collapsing it to the default, and none of them is the
	// default, so that a sync which failed to publish anything could not pass by coincidence with what the mirror
	// already held.
	published := []fxp.Int{
		PermittedScriptExecTimeMin,
		fxp.FromStringForced("0.1"),
		fxp.FromStringForced("0.2"),
		fxp.FromStringForced("0.3"),
		PermittedScriptExecTimeMax,
	}
	for _, limit := range published {
		c.Equal(limit, fxp.ResetIfOutOfRange(limit, PermittedScriptExecTimeMin, PermittedScriptExecTimeMax,
			PermittedScriptExecTimeDef), "%v must be a value the user could configure", limit)
		c.NotEqual(PermittedScriptExecTimeDef, limit, "%v must not be the default", limit)
	}
	general.PermittedPerScriptExecTime = published[0]
	SyncScriptExecTimeLimit()
	c.Equal(published[0], scriptExecTimeLimit(), "the sync must publish the value it was given")

	// The reader records what it sees rather than asserting as it goes: the checker's assertions lock the test's
	// mutex, which would establish an ordering between the two goroutines that could hide a race from the detector.
	var strays []fxp.Int
	var wg sync.WaitGroup
	start := make(chan struct{})
	wg.Go(func() {
		<-start
		for range 200 {
			if limit := scriptExecTimeLimit(); !slices.Contains(published, limit) {
				strays = append(strays, limit)
			}
		}
	})
	const writes = 200
	wg.Go(func() {
		<-start
		for i := range writes {
			general.PermittedPerScriptExecTime = published[i%len(published)]
			SyncScriptExecTimeLimit()
		}
	})
	close(start)
	wg.Wait()
	c.Equal(0, len(strays), "the reader must see only published limits, but saw %v", strays[:min(len(strays), 5)])
	c.Equal(published[(writes-1)%len(published)], scriptExecTimeLimit(),
		"the last limit published must be the one in effect")
}

// TestGlobalSheetSettingsSyncIsConcurrencySafe verifies that an entity without embedded sheet settings can be
// unmarshaled on a background goroutine while the UI thread mutates the global sheet settings in place, replaces them
// wholesale, and republishes the snapshot -- the combination the deep search content cache and the settings dockables
// produce. The unmarshal must clone only published snapshots, never the live settings being edited.
func TestGlobalSheetSettingsSyncIsConcurrencySafe(t *testing.T) {
	c := check.New(t)
	settings := GlobalSettings()
	// A clone must be captured for the restore, since the writer below mutates the live object in place.
	prevSheet := settings.Sheet.Clone(nil)
	defer func() {
		settings.Sheet = prevSheet
		SyncGlobalSheetSettings()
	}()

	data := fmt.Appendf(nil, `{"version":%d}`, jio.CurrentDataVersion)
	var wg sync.WaitGroup
	start := make(chan struct{})
	wg.Go(func() {
		<-start
		for range 100 {
			var e Entity
			e.DiscardCaches()
			c.NoError(jio.Unmarshal(data, &e))
			c.NotNil(e.SheetSettings, "the unmarshaled entity must have been given a clone of the global sheet settings")
		}
	})
	wg.Go(func() {
		<-start
		for i := range 100 {
			if i%10 == 9 {
				settings.Sheet = FactorySheetSettings()
			} else {
				settings.Sheet.ShowAllWeapons = !settings.Sheet.ShowAllWeapons
				settings.Sheet.Attributes = FactoryAttributeDefs()
			}
			SyncGlobalSheetSettings()
		}
	})
	close(start)
	wg.Wait()
}

// TestClosedStateIsConcurrencySafe verifies that the unmarshalers for the legacy file format, which record the open
// state of their container rows in the global closed-state map, can run on a background goroutine -- as the deep search
// content cache runs them -- while the UI thread refreshes and toggles closed states and saves the settings. Without
// the lock, both sides write the map and the runtime throws with "concurrent map writes".
func TestClosedStateIsConcurrencySafe(t *testing.T) {
	c := check.New(t)
	prevPath := SettingsPath
	SettingsPath = filepath.Join(t.TempDir(), "settings.json")
	t.Cleanup(func() { SettingsPath = prevPath })
	const closedKey = "n:closed-state-test-closed"
	const toggledKey = "n:closed-state-test-toggled"
	SetClosedState(closedKey, true)
	t.Cleanup(func() {
		SetClosedState(closedKey, false)
		SetClosedState(toggledKey, false)
	})

	// A legacy-format trait list: pre-TID ids and the "open" flag that later versions moved to the closed-state map.
	legacy := fstest.MapFS{"Test.adq": {Data: []byte(`{
	"version": 2,
	"rows": [
		{
			"type": "advantage_container",
			"id": "8f7a1c2e-legacy",
			"name": "Container",
			"open": true,
			"children": [{"type": "advantage", "id": "3b9d0e4f-legacy", "name": "Child"}]
		}
	]
}`)}}

	// The loops below avoid the checker, whose assertions lock the test's mutex and would establish enough ordering
	// between the two goroutines to hide the race from the detector; the results are checked once both have finished.
	var loadErr, saveErr error
	loaded, open, stayedClosed := 0, 0, true
	var wg sync.WaitGroup
	start := make(chan struct{})
	wg.Go(func() {
		<-start
		for range 200 {
			traits, err := NewTraitsFromFile(legacy, "Test.adq")
			if err != nil {
				loadErr = err
				return
			}
			loaded++
			if len(traits) == 1 && traits[0].IsOpen() {
				open++
			}
		}
	})
	wg.Go(func() {
		<-start
		settings := GlobalSettings()
		for i := range 200 {
			if !IsClosed(closedKey) {
				stayedClosed = false
			}
			SetClosedState(toggledKey, i%2 == 0)
			if i%50 == 49 {
				if err := settings.Save(); err != nil {
					saveErr = err
					return
				}
			}
		}
	})
	close(start)
	wg.Wait()
	c.NoError(loadErr)
	c.Equal(200, loaded)
	c.Equal(200, open, "a legacy container flagged open must load open")
	c.NoError(saveErr)
	c.True(stayedClosed, "a closed key must stay closed while it is only being refreshed")
	c.True(IsClosed(closedKey), "the closed key must survive the saves")
}
