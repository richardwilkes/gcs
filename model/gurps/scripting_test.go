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
	"sync"
	"testing"

	"github.com/dop251/goja"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestScriptMathExp2 verifies that Math.exp2 is exposed to scripts as a real member of the built-in Math object (rather
// than an unreachable global with a dotted name), while leaving the standard Math members intact.
func TestScriptMathExp2(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		script string
		want   string
	}{
		{script: "typeof Math.exp2", want: "function"},
		{script: "Math.exp2(0)", want: "1"},
		{script: "Math.exp2(3)", want: "8"},
		{script: "Math.exp2(10)", want: "1024"},
		{script: "typeof Math.pow", want: "function"}, // ensure the built-in Math members are still present
	} {
		v, err := runScript(0, tc.script)
		c.NoError(err, "script %q", tc.script)
		c.Equal(tc.want, v.String(), "script %q", tc.script)
	}
}

// TestScriptRandomWeightInPounds verifies that entity.randomWeightInPounds returns a numeric result without panicking
// for very low (even negative) strength values. At such ST the internal deviation used as the bound for rnd.Intn would
// otherwise go non-positive; the clamp keeps the call safe regardless of how rnd.Intn treats a non-positive bound.
func TestScriptRandomWeightInPounds(t *testing.T) {
	c := check.New(t)
	entityArg := ScriptArg{Name: "entity", Value: func(r *goja.Runtime) any { return newScriptEntity(r, nil) }}
	for _, st := range []int{-100, -20, -1, 0, 1, 10, 20} {
		v, err := runScript(0, fmt.Sprintf("entity.randomWeightInPounds(%d, 0)", st), entityArg)
		c.NoError(err, "st %d", st)
		c.NotNil(v, "st %d", st)
	}
}

// TestScriptResolutionConcurrency hammers the package-global script state (the compiled-program cache, the entity-less
// resolve cache and its discard path, and the entity-less recursion-depth counter) from several goroutines at once. Run
// under the race detector (go test -race) it guards against reintroducing unsynchronized access to that shared state.
// The goroutine count is kept below maximumAllowedResolvingDepth so concurrency alone cannot trip the depth limit.
func TestScriptResolutionConcurrency(t *testing.T) {
	DiscardGlobalResolveCache()
	const goroutines = 8
	const iterations = 200
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := range goroutines {
		go func() {
			defer wg.Done()
			for i := range iterations {
				// A shared script (a cache hit once warmed) exercises concurrent reads of the same key.
				if got := ResolveScript(nil, ScriptSelfProvider{}, "3 + 4"); got != "7" {
					t.Errorf("shared script = %q, want %q", got, "7")
				}
				// A unique script per (goroutine, iteration) exercises concurrent compile + store.
				_ = ResolveScript(nil, ScriptSelfProvider{}, fmt.Sprintf("%d + %d", g, i))
				if i%25 == 0 {
					DiscardGlobalResolveCache()
				}
			}
		}()
	}
	wg.Wait()
}
