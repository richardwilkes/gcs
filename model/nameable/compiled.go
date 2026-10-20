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
	"strings"
	"sync"
)

// compiledPart is one piece of a compiled string: literal text, or a marker identified by its key along with the text
// that stands in for it when the replacements have no entry for that key.
type compiledPart struct {
	key  string // Empty for literal text; a parsed marker's key is never empty, since its label can't be.
	text string // The literal text, or a marker's fallback text.
}

// compiledText is a string that may contain markers, parsed once so that applying any set of replacements to it is
// a matter of looking up keys. Parsing a marker runs two regular expressions and allocates a dozen times, and the
// names on a sheet are applied over and over -- every lookup of a skill by name applies the name of every skill -- so
// the parse is what gets cached. A compiled text is never modified once built, so any number of goroutines may share
// one.
type compiledText struct {
	parts []compiledPart
	keys  []string // The keys of the markers that parsed, in order, duplicates included.
}

// compileText parses the string. It must produce exactly what ApplyToList and Extract produced when they parsed on
// every call; see compiled_test.go, which checks it against copies of them.
func compileText(str string) *compiledText {
	var ct compiledText
	var literal strings.Builder
	flush := func() {
		if literal.Len() > 0 {
			ct.parts = append(ct.parts, compiledPart{text: literal.String()})
			literal.Reset()
		}
	}
	for _, part := range ExtractParts(str, MarkerDelimiter, MarkerDelimiter) {
		if !part.Placeholder {
			// Unescape any escaped marker delimiters, so a literal "\@" the user typed to keep an '@' out of marker
			// detection displays as a plain '@' rather than surfacing the escape itself.
			literal.WriteString(UnescapeRunes(part.Value, MarkerDelimiter))
			continue
		}
		m, ok := NewMarker(UnescapeRunes(part.Value, MarkerDelimiter))
		if !ok {
			// Marker parsing failed, so the original text goes back out with its delimiters.
			literal.WriteRune(MarkerDelimiter)
			literal.WriteString(part.Value)
			literal.WriteRune(MarkerDelimiter)
			continue
		}
		flush()
		// An unresolved marker is rendered in its compact form, "@Label@", keeping the '@' wrapper so displayed text
		// still visibly flags it as unresolved. A legacy "Label: text" marker with no example list keeps its full
		// text instead, since that text often carries meaning the label alone does not, such as "Who: A deity".
		var fallback strings.Builder
		fallback.WriteRune(MarkerDelimiter)
		if m.Legacy && len(m.Options) == 0 {
			fallback.WriteString(m.Raw)
		} else {
			fallback.WriteString(m.Label) // We do not escape markers here by design
		}
		fallback.WriteRune(MarkerDelimiter)
		key := m.Key()
		ct.parts = append(ct.parts, compiledPart{key: key, text: fallback.String()})
		ct.keys = append(ct.keys, key)
	}
	flush()
	return &ct
}

// apply returns the text with each marker replaced by its replacement, or by its fallback text when there is none.
func (t *compiledText) apply(replacements map[string]string) string {
	switch len(t.parts) {
	case 0:
		return ""
	case 1:
		// The common cases -- a name that is one marker, or one whose '@' turned out not to start a marker -- need
		// no building.
		p := &t.parts[0]
		if p.key == "" {
			return p.text
		}
		if r, ok := replacements[p.key]; ok {
			return r
		}
		return p.text
	}
	var sb strings.Builder
	for i := range t.parts {
		p := &t.parts[i]
		if p.key == "" {
			sb.WriteString(p.text)
		} else if r, ok := replacements[p.key]; ok {
			sb.WriteString(r)
		} else {
			sb.WriteString(p.text)
		}
	}
	return sb.String()
}

// maximumCompiledTexts is how many compiled strings one generation of the cache holds before it is retired, so the
// cache as a whole never holds more than twice this many; see compiled.
const maximumCompiledTexts = 4096

// maximumCompiledTextLength is the longest string the cache will hold. A longer one is compiled on every call, which
// keeps a pasted novel of a note from taking up most of the cache.
const maximumCompiledTextLength = 4096

// compiledTexts holds compiled strings keyed by the string, so that a name applied over and over is parsed only once.
// It is shared by every goroutine that applies replacements, so all access is guarded by compiledMutex.
//
// The cache has to be bounded, because the set of strings it sees is not: the item editors apply the text being typed
// after every keystroke, so each intermediate is a distinct key an unbounded map would hold until the process exited.
// The bound is generational, as the compiled script cache's is. Entries go into the current generation; once it is
// full, it becomes oldCompiledTexts and a fresh generation takes its place, dropping whatever the generation before it
// still held. A lookup that misses the current generation but hits the old one promotes the entry, so anything still
// in use survives the turnover and only entries left untouched across two full generations are discarded.
var (
	compiledMutex    sync.Mutex
	compiledTexts    = make(map[string]*compiledText, maximumCompiledTexts)
	oldCompiledTexts map[string]*compiledText
)

// compiled returns the compiled form of the string, from the cache when it holds one.
func compiled(str string) *compiledText {
	if len(str) > maximumCompiledTextLength {
		return compileText(str)
	}
	compiledMutex.Lock()
	if ct, ok := compiledTexts[str]; ok {
		compiledMutex.Unlock()
		return ct
	}
	if ct, ok := oldCompiledTexts[str]; ok {
		addCompiledText(str, ct)
		compiledMutex.Unlock()
		return ct
	}
	compiledMutex.Unlock()
	// Parsing happens outside the lock, so that one slow parse doesn't hold up every other goroutine. Another
	// goroutine may have parsed the same string in the meantime, in which case its result is the one kept, so that
	// all callers share a single instance.
	ct := compileText(str)
	compiledMutex.Lock()
	defer compiledMutex.Unlock()
	if existing, ok := compiledTexts[str]; ok {
		return existing
	}
	addCompiledText(str, ct)
	return ct
}

// addCompiledText adds an entry to the current generation of the cache, retiring that generation first if it is
// already full. The key is cloned, so that the cache never keeps a larger string the key was sliced from alive.
// compiledMutex must be held.
func addCompiledText(str string, ct *compiledText) {
	if len(compiledTexts) >= maximumCompiledTexts {
		oldCompiledTexts = compiledTexts
		compiledTexts = make(map[string]*compiledText, maximumCompiledTexts)
	}
	compiledTexts[strings.Clone(str)] = ct
}

// discardCompiledTexts drops every compiled string the cache holds. The cache bounds itself, so this exists only for
// tests that need a known starting point.
func discardCompiledTexts() {
	compiledMutex.Lock()
	defer compiledMutex.Unlock()
	clear(compiledTexts)
	oldCompiledTexts = nil
}
