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
	"sync"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestCurrentGoroutineID verifies the two properties per-goroutine state keyed by the ID depends on: every call from
// one goroutine returns the same non-zero ID, including from deeper in its call stack, and no two goroutines alive at
// the same time share one. A zero ID would mean the runtime.Stack header could not be parsed, which is the signal that
// the fallback described on currentGoroutineID is in effect and this test should fail loudly rather than let it pass
// silently.
func TestCurrentGoroutineID(t *testing.T) {
	c := check.New(t)
	id := currentGoroutineID()
	c.NotEqual(uint64(0), id, "the calling goroutine's ID must be parsed from the runtime.Stack header")
	c.Equal(id, currentGoroutineID(), "repeated calls on one goroutine must agree")
	var nested func(depth int) uint64
	nested = func(depth int) uint64 {
		if depth == 0 {
			return currentGoroutineID()
		}
		return nested(depth - 1)
	}
	c.Equal(id, nested(50), "the ID must not depend on how deep the call stack is")

	const goroutines = 64
	ids := make([]uint64, goroutines)
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := range goroutines {
		go func() {
			defer wg.Done()
			ids[i] = currentGoroutineID()
		}()
	}
	wg.Wait()
	seen := map[uint64]bool{id: true}
	for i, got := range ids {
		c.NotEqual(uint64(0), got, "goroutine %d must have a parsed ID", i)
		c.False(seen[got], "goroutine %d must not share ID %d with another live goroutine", i, got)
		seen[got] = true
	}
}
