// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package nameable

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// A string that is in use must survive the cache's turnover, while strings seen only once are dropped.
func TestCompiledTextCacheRotation(t *testing.T) {
	c := check.New(t)
	discardCompiledTexts()
	defer discardCompiledTexts()
	hot := "@Hot@"
	hotCompiled := compiled(hot)
	for i := range 3 * maximumCompiledTexts {
		compiled(fmt.Sprintf("@Cold %d@", i))
		if i%(maximumCompiledTexts/2) == 0 {
			// Compared as pointers: a recompiled copy of the same text would satisfy a deep comparison.
			c.True(hotCompiled == compiled(hot), "the hot string must keep its compiled form across turnovers")
		}
	}
	compiledMutex.Lock()
	defer compiledMutex.Unlock()
	c.True(len(compiledTexts) <= maximumCompiledTexts, "the current generation is bounded")
	c.True(len(oldCompiledTexts) <= maximumCompiledTexts, "the retired generation is bounded")
	_, inCurrent := compiledTexts[hot]
	_, inOld := oldCompiledTexts[hot]
	c.True(inCurrent || inOld, "the hot string must still be cached")
	_, coldKept := compiledTexts["@Cold 0@"]
	_, coldOld := oldCompiledTexts["@Cold 0@"]
	c.False(coldKept || coldOld, "a string seen once two generations ago must have been dropped")
}

// A string longer than the cache accepts is still applied correctly, just without being cached.
func TestCompiledTextCacheSkipsLongStrings(t *testing.T) {
	c := check.New(t)
	discardCompiledTexts()
	defer discardCompiledTexts()
	long := strings.Repeat("x", maximumCompiledTextLength) + " @Skill@"
	c.Equal(strings.Repeat("x", maximumCompiledTextLength)+" Broadsword", Apply(long, map[string]string{"Skill": "Broadsword"}))
	compiledMutex.Lock()
	defer compiledMutex.Unlock()
	c.Equal(0, len(compiledTexts), "a long string must not be cached")
}

// The cache is shared by every goroutine that applies replacements, such as the workers that build the library
// search content, so applying from several at once, with strings they share and enough distinct ones to turn the cache
// over, must give each the same answers as applying alone.
func TestApplyConcurrency(t *testing.T) {
	c := check.New(t)
	discardCompiledTexts()
	defer discardCompiledTexts()
	replacements := map[string]string{"Skill": "Broadsword", "Attack": "Swung"}
	shared := []struct {
		in   string
		want string
	}{
		{in: "@Skill@", want: "Broadsword"},
		{in: "Targeted Attack (@Skill@ @Attack@/Neck)", want: "Targeted Attack (Broadsword Swung/Neck)"},
		{in: "Ritual Magic (@Specialty@)", want: "Ritual Magic (@Specialty@)"},
		{in: "@Who: A deity@", want: "@Who: A deity@"},
		{in: `\@Skill@`, want: "@Skill@"},
	}
	const workers = 8
	const unique = maximumCompiledTexts / 2
	var wg sync.WaitGroup
	failures := make([]string, workers)
	for w := range workers {
		wg.Go(func() {
			for i := range unique {
				input := fmt.Sprintf("Worker %d @Skill@ %d @Specialty|opt%d|*@", w, i, i)
				want := fmt.Sprintf("Worker %d Broadsword %d @Specialty@", w, i)
				if got := Apply(input, replacements); got != want {
					failures[w] = fmt.Sprintf("Apply(%q) = %q, want %q", input, got, want)
					return
				}
				for _, one := range shared {
					if got := Apply(one.in, replacements); got != one.want {
						failures[w] = fmt.Sprintf("Apply(%q) = %q, want %q", one.in, got, one.want)
						return
					}
				}
			}
		})
	}
	wg.Wait()
	for w, failure := range failures {
		c.Equal("", failure, "worker %d", w)
	}
}
