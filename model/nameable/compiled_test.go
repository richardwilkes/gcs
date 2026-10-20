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
	"maps"
	"strings"
	"sync"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// referenceApplyToList is ApplyToList as it was before the parse was cached: it parses each string on every call. The
// cached version must produce exactly what it does.
func referenceApplyToList(in []string, replacements map[string]string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	for i, str := range in {
		if !strings.ContainsRune(str, MarkerDelimiter) {
			out[i] = str
			continue
		}
		var sb strings.Builder
		for _, part := range ExtractParts(str, MarkerDelimiter, MarkerDelimiter) {
			if part.Placeholder {
				if m, ok := NewMarker(UnescapeRunes(part.Value, MarkerDelimiter)); ok {
					if r, hasReplacement := replacements[m.Key()]; hasReplacement {
						sb.WriteString(r)
					} else {
						sb.WriteRune(MarkerDelimiter)
						if m.Legacy && len(m.Options) == 0 {
							sb.WriteString(m.Raw)
						} else {
							sb.WriteString(m.Label)
						}
						sb.WriteRune(MarkerDelimiter)
					}
				} else {
					sb.WriteRune(MarkerDelimiter)
					sb.WriteString(part.Value)
					sb.WriteRune(MarkerDelimiter)
				}
			} else {
				sb.WriteString(UnescapeRunes(part.Value, MarkerDelimiter))
			}
		}
		out[i] = sb.String()
	}
	return out
}

// referenceExtract is Extract as it was before the parse was cached.
func referenceExtract(nameables, replacements map[string]string, in ...string) map[string]string {
	if nameables == nil {
		nameables = make(map[string]string)
	}
	for _, src := range in {
		for _, part := range ExtractParts(src, MarkerDelimiter, MarkerDelimiter) {
			if part.Placeholder {
				if m, ok := NewMarker(UnescapeRunes(part.Value, MarkerDelimiter)); ok {
					if v, exists := replacements[m.Key()]; exists {
						nameables[m.Key()] = v
					} else {
						nameables[m.Key()] = Unset
					}
				}
			}
		}
	}
	return nameables
}

// compiledTestInputs are the strings the cached and reference implementations are compared on: plain text, every
// marker form, the escapes, and the malformed input the parser has to cope with.
var compiledTestInputs = []string{
	"",
	"Broadsword",
	"@Skill@",
	"Ritual Magic (@Specialty@)",
	"Targeted Attack (@Skill@ @Attack@/Neck)",
	"@Skill@@Attack@",
	"@a@b@c@",
	"@Skill: Broadsword, Shortsword, etc.@",
	"@Skill: Broadsword, Shortsword, etc@",
	"@Who: A deity@",
	"@Skill|Broadsword|Shortsword|*@",
	"@Skill|Broadsword|Shortsword|?@",
	"@Skill|tt(Pick one\\nor type your own)|Broadsword|*|?@",
	"@ Skill @",
	"@ |x@",
	"@|@",
	"@@",
	"@",
	"unclosed @Skill",
	"@Skill@ and then @",
	"trailing escape \\",
	"escaped \\@ delimiter",
	"double escaped \\\\@Skill@",
	"escaped inside @Sk\\@ill@",
	"@Label\\|with pipe@",
	"control\nbreaks @Skill\n@ markers",
	"tab\t@Skill@",
	"ünïcödé @Spéciälty@ ✓",
	"@Spécialty|ópt|*@",
	"@Skill@@",
	"@@Skill@",
	"\\@Skill@",
	"@Skill\\@",
}

// compiledTestReplacements are the replacement sets each input is applied with: none, a full set, a partial set and
// a set whose values are empty, which still count as replacements.
func compiledTestReplacements() []map[string]string {
	return []map[string]string{
		nil,
		{},
		{
			"Skill": "Broadsword", "Attack": "Swung", "Specialty": "Fire", "a": "A", "b": "B", "c": "C",
			"Spécialty": "Feu", "Who": "Nobody", "Label|with pipe": "Piped", "Sk@ill": "Escaped",
			"Skill: Broadsword, Shortsword, etc.": "Legacy", "Skill: Broadsword, Shortsword, etc": "Legacy too",
			"Who: A deity": "Thor", "Skill|Broadsword|Shortsword|*": "Options", "Skill|Broadsword|Shortsword|?": "Maybe",
			"Skill|Broadsword|Shortsword|*|?|tt(Pick one\\nor type your own)": "Tooltipped", "Spécialty|ópt|*": "Ópt",
		},
		{"Skill": "Broadsword"},
		{"Skill": "", "Specialty": ""},
	}
}

// The cached implementation must produce byte for byte what parsing on every call produced, for every input and
// every set of replacements, whether the input is being seen for the first time or comes from the cache.
func TestApplyMatchesReference(t *testing.T) {
	c := check.New(t)
	discardCompiledTexts()
	for round := range 2 {
		for _, input := range compiledTestInputs {
			for i, replacements := range compiledTestReplacements() {
				want := referenceApplyToList([]string{input}, replacements)[0]
				c.Equal(want, Apply(input, replacements), "Apply(%q) with replacements %d in round %d", input, i, round)
				c.Equal(referenceApplyToList([]string{input, input}, replacements),
					ApplyToList([]string{input, input}, replacements), "ApplyToList(%q) with replacements %d in round %d",
					input, i, round)
				c.Equal(referenceExtract(nil, replacements, input), Extract(nil, replacements, input),
					"Extract(%q) with replacements %d in round %d", input, i, round)
			}
		}
	}
	c.Equal(referenceApplyToList(compiledTestInputs, compiledTestReplacements()[2]),
		ApplyToList(compiledTestInputs, compiledTestReplacements()[2]), "ApplyToList over every input")
	c.Equal(referenceExtract(nil, compiledTestReplacements()[2], compiledTestInputs...),
		Extract(nil, compiledTestReplacements()[2], compiledTestInputs...), "Extract over every input")
	c.Nil(ApplyToList(nil, nil), "an empty list yields nil")
	c.Nil(ApplyToList([]string{}, nil), "an empty list yields nil")
	existing := map[string]string{"kept": "as is"}
	c.Equal(referenceExtract(maps.Clone(existing), nil, "@Skill@"), Extract(maps.Clone(existing), nil, "@Skill@"),
		"Extract adds to the map it is given")
}

func FuzzApplyMatchesReference(f *testing.F) {
	for _, input := range compiledTestInputs {
		f.Add(input)
	}
	f.Fuzz(func(t *testing.T, input string) {
		c := check.New(t)
		replacements := referenceExtract(nil, nil, input)
		for k := range replacements {
			replacements[k] = "<" + k + ">"
		}
		c.Equal(referenceApplyToList([]string{input}, replacements)[0], Apply(input, replacements))
		c.Equal(referenceApplyToList([]string{input}, nil)[0], Apply(input, nil))
		c.Equal(referenceExtract(nil, replacements, input), Extract(nil, replacements, input))
	})
}

// A compiled string must not be altered by applying it, however the replacements vary.
func TestCompiledTextIsNotAlteredByApplying(t *testing.T) {
	c := check.New(t)
	ct := compileText("Targeted Attack (@Skill@ @Attack@/Neck)")
	c.Equal([]string{"Skill", "Attack"}, ct.keys)
	c.Equal("Targeted Attack (Broadsword Swung/Neck)", ct.apply(map[string]string{"Skill": "Broadsword", "Attack": "Swung"}))
	c.Equal("Targeted Attack (@Skill@ @Attack@/Neck)", ct.apply(nil))
	c.Equal("Targeted Attack ( /Neck)", ct.apply(map[string]string{"Skill": "", "Attack": ""}))
	c.Equal("Targeted Attack (Broadsword @Attack@/Neck)", ct.apply(map[string]string{"Skill": "Broadsword"}))
	c.Equal([]string{"Skill", "Attack"}, ct.keys)
}

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
			c.Equal(hotCompiled, compiled(hot), "the hot string must keep its compiled form across turnovers")
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
// search content, so applying from several at once, with enough distinct strings to turn the cache over, must give
// each the same answers as applying alone.
func TestApplyConcurrency(t *testing.T) {
	c := check.New(t)
	discardCompiledTexts()
	defer discardCompiledTexts()
	replacements := compiledTestReplacements()[2]
	const workers = 8
	const unique = maximumCompiledTexts / 2
	var wg sync.WaitGroup
	failures := make([]string, workers)
	for w := range workers {
		wg.Go(func() {
			for i := range unique {
				input := fmt.Sprintf("Worker %d @Skill@ %d @Specialty|opt%d|*@", w, i, i)
				if got, want := Apply(input, replacements), referenceApplyToList([]string{input}, replacements)[0]; got != want {
					failures[w] = fmt.Sprintf("Apply(%q) = %q, want %q", input, got, want)
					return
				}
				for _, shared := range compiledTestInputs {
					if got, want := Apply(shared, replacements), referenceApplyToList([]string{shared}, replacements)[0]; got != want {
						failures[w] = fmt.Sprintf("Apply(%q) = %q, want %q", shared, got, want)
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
