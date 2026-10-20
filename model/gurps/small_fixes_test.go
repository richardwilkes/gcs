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
	"log/slog"
	"sync"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
)

// A value that fails to resolve fails every time it is asked for, which for an equipment weight is once per draw of
// the sheet, so the failure is reported once per edit rather than once per frame.
func TestScriptResolveFailureLoggedOnce(t *testing.T) {
	c := check.New(t)
	count := countLogs(t, slog.LevelError)
	DiscardGlobalResolveCache()
	for range 5 {
		c.Equal(fxp.Int(0), ResolveToNumber(nil, ScriptSelfProvider{}, "'not a number'"))
	}
	c.Equal(int32(1), count.Load(), "the same failure is reported once")
	c.Equal(fxp.Weight(0), ResolveToWeight(nil, ScriptSelfProvider{}, "'not a number'", fxp.Pound))
	c.Equal(int32(2), count.Load(), "a failure to resolve to another kind of value is a different failure")
	DiscardGlobalResolveCache()
	c.Equal(fxp.Int(0), ResolveToNumber(nil, ScriptSelfProvider{}, "'not a number'"))
	c.Equal(int32(3), count.Load(), "an edit discards the cache, after which a failure that persists is reported again")
}

// The record of reported failures is shared by every goroutine that resolves scripts.
func TestScriptResolveFailureLogConcurrency(t *testing.T) {
	c := check.New(t)
	count := countLogs(t, slog.LevelError)
	DiscardGlobalResolveCache()
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 20 {
				ResolveToNumber(nil, ScriptSelfProvider{}, "'not a number'")
			}
		})
	}
	wg.Wait()
	c.Equal(int32(1), count.Load(), "the failure is reported once, however many goroutines hit it")
}

// The carried weight is derived from every carried item, so it is cached until the caches are discarded, as the
// encumbrance level derived from it already was.
func TestWeightCarriedCached(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	eqp := NewEquipment(e, nil, false)
	eqp.BaseWeight = "10 lb"
	e.CarriedEquipment = append(e.CarriedEquipment, eqp)
	e.Recalculate()
	c.Equal(fxp.WeightFromInteger(10, fxp.Pound), e.WeightCarried(false))
	eqp.BaseWeight = "20 lb"
	c.Equal(fxp.WeightFromInteger(10, fxp.Pound), e.WeightCarried(false), "the weight is cached until the caches are discarded")
	e.DiscardCaches()
	c.Equal(fxp.WeightFromInteger(20, fxp.Pound), e.WeightCarried(false))
}
