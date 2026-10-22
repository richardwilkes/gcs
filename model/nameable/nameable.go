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
	"maps"
	"slices"
	"strings"

	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Unset is a sentinel value used in the nameables map produced by Extract to indicate that a marker has no recorded
// replacement yet. It is bracketed with NUL bytes so it won't collide with a value a user typed or that came from file
// content. Callers that read Extract's output must treat it the same as the key being absent.
const Unset = "\x00unset\x00"

// Detail is an author's definition of the nameable an item's text refers to as "@code@", held by the item under the
// code.
type Detail struct {
	Label    string   `json:"label,omitzero"`
	Tooltip  string   `json:"tooltip,omitzero"`
	Options  []string `json:"options,omitempty"`
	FreeForm bool     `json:"free_form,omitzero"`
	NotBlank bool     `json:"not_blank,omitzero"`
}

// CloneDetails returns a copy of the nameables that shares nothing with them.
func CloneDetails(nameables map[string]Detail) map[string]Detail {
	if nameables == nil {
		return nil
	}
	clone := make(map[string]Detail, len(nameables))
	for code, detail := range nameables {
		detail.Options = slices.Clone(detail.Options)
		clone[code] = detail
	}
	return clone
}

// Describe returns a plain-language description of what the nameable with the code asks for, in the form the nameables
// editor shows. Its code, name and options are passed through em, which may wrap them for emphasis; pass an identity
// func for plain text.
func (d *Detail) Describe(code string, em func(string) string) string {
	name := d.Label
	if name == "" {
		name = code
	}
	marker := em(string(MarkerDelimiter) + code + string(MarkerDelimiter))
	name = em(name)
	switch {
	case len(d.Options) != 0 && d.FreeForm && d.NotBlank:
		return i18n.Text("%s asks for %s, suggesting %s, and can't be blank", marker, name,
			describeOptions(d.Options, em))
	case len(d.Options) != 0 && d.FreeForm:
		return i18n.Text("%s asks for %s, suggesting %s", marker, name, describeOptions(d.Options, em))
	case len(d.Options) != 0 && d.NotBlank:
		return i18n.Text("%s asks for %s, one of %s, and can't be blank", marker, name, describeOptions(d.Options, em))
	case len(d.Options) != 0:
		return i18n.Text("%s asks for %s, one of %s", marker, name, describeOptions(d.Options, em))
	case d.FreeForm && d.NotBlank:
		return i18n.Text("%s asks for %s as any text, which can't be blank", marker, name)
	case d.FreeForm:
		return i18n.Text("%s asks for %s as any text", marker, name)
	default:
		return i18n.Text("%s asks for %s, with nothing to choose from", marker, name)
	}
}

// describeOptions lists the options, naming at most four of them.
func describeOptions(options []string, em func(string) string) string {
	const most = 4
	named := make([]string, 0, most)
	for i, one := range options {
		if i == most-1 && len(options) > most {
			break
		}
		named = append(named, em(one))
	}
	if rest := len(options) - len(named); rest > 0 {
		return i18n.Text("%s or %d more", strings.Join(named, i18n.Text(", ")), rest)
	}
	if len(named) == 1 {
		return named[0]
	}
	return i18n.Text("%s or %s", strings.Join(named[:len(named)-1], i18n.Text(", ")), named[len(named)-1])
}

// Filler defines the method for filling the nameable key map.
type Filler interface {
	FillWithNameableKeys(m, existing map[string]string)
}

// Accesser defines the method for retrieving the nameable replacements.
type Accesser interface {
	NameableReplacements() map[string]string
}

// Setter defines the method for setting the nameable replacements.
type Setter interface {
	SetNameableReplacements(replacements map[string]string)
}

// Applier is implemented by types that participate in the nameable adjustments.
type Applier interface {
	Accesser
	Filler
	ApplyNameableKeys(m map[string]string)
}

// Definer defines the method for retrieving an item's own nameable definitions.
type Definer interface {
	NameableDefinitions() map[string]Detail
}

// Extract adds a key for each nameable marker found in the provided strings to nameables, allocating that map if it is
// nil, and returns it. Each key's value comes from replacements, or is Unset when replacements has no entry for it.
func Extract(nameables, replacements map[string]string, in ...string) map[string]string {
	if nameables == nil {
		nameables = make(map[string]string)
	}
	for _, src := range in {
		if !strings.ContainsRune(src, MarkerDelimiter) {
			continue
		}
		// Duplicate markers are fine, since they collapse to a single entry.
		for _, key := range compiled(src).keys {
			if v, exists := replacements[key]; exists {
				nameables[key] = v
			} else {
				nameables[key] = Unset
			}
		}
	}
	return nameables
}

// Normalize returns a replacements map with each key rewritten to its normalized marker form, or nil if there is
// nothing to normalize. Keys that fail to parse as markers are retained unchanged. Source keys are processed in sorted
// order, so two source keys that normalize to the same form always resolve the same way.
func Normalize(replacements map[string]string) map[string]string {
	if len(replacements) == 0 {
		return nil
	}
	out := make(map[string]string, len(replacements))
	for _, k := range slices.Sorted(maps.Keys(replacements)) {
		if m, ok := NewMarker(k); ok {
			out[m.Key()] = replacements[k]
		} else {
			out[k] = replacements[k]
		}
	}
	return out
}

// Uses returns how many times each nameable key is referred to by a marker in the strings.
func Uses(in ...string) map[string]int {
	uses := make(map[string]int)
	for _, src := range in {
		if strings.ContainsRune(src, MarkerDelimiter) {
			for _, key := range compiled(src).keys {
				uses[key]++
			}
		}
	}
	return uses
}

// Apply replaces nameable markers with their replacement values in a single string. An unresolved marker is rendered
// in its compact form, "`@Label@`", keeping the '@' wrapper so displayed text (sheet rows, table columns, tooltips)
// still visibly flags it as unresolved. A legacy "Label: text" marker with no example list keeps its full text instead,
// since that text often carries meaning the label alone does not, such as "Who: A deity". The parse of the string is
// cached (see compiled), so applying the same string again costs only the lookups of its markers.
func Apply(str string, replacements map[string]string) string {
	if !strings.ContainsRune(str, MarkerDelimiter) {
		return str
	}
	return compiled(str).apply(replacements)
}

// ApplyLinked is Apply for text whose markers may be linked to other keys: a marker whose key links names another key
// takes that key's replacement.
func ApplyLinked(str string, replacements, links map[string]string) string {
	if !strings.ContainsRune(str, MarkerDelimiter) {
		return str
	}
	return compiled(str).applyLinked(replacements, links)
}

// ApplyToList is Apply for a slice of strings. Returns nil if the slice is empty.
func ApplyToList(in []string, replacements map[string]string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	for i, str := range in {
		out[i] = Apply(str, replacements)
	}
	return out
}

// Reduce returns a map of the replacements which exist in nameables.
//
// Both maps must already be keyed by normalized marker key: nameables from Extract, and replacements from load-time
// normalization (see Normalize) or the output of a previous Reduce call. Reduce does not itself normalize either map's
// keys.
//
// An entry still holding the Unset sentinel (i.e. the substitutions dialog was shown but the user never chose a value
// for that marker) is dropped, so an untouched marker is never persisted as if it had been resolved.
//
// Returns nil, not an empty map, when either map is empty -- callers that assign the result directly to a stored
// Replacements field and later write into it (e.g. `x.Replacements[k] = v`) must guard for nil first.
func Reduce(nameables, replacements map[string]string) map[string]string {
	if len(nameables) == 0 || len(replacements) == 0 {
		return nil
	}
	result := make(map[string]string, min(len(nameables), len(replacements)))
	for k, v := range replacements {
		if v == Unset {
			continue
		}
		if _, found := nameables[k]; found {
			result[k] = v
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// Missing returns the nameable keys that have no replacement, in no particular order.
//
// Both maps must already be keyed by normalized marker key: nameables from Extract, and replacements from load-time
// normalization (see Normalize) or the output of a previous Reduce call. Missing does not itself normalize either
// map's keys. The returned keys are normalized, since they are drawn from nameables.
func Missing(nameables, replacements map[string]string) []string {
	if len(nameables) == 0 {
		return nil
	}
	if len(replacements) == 0 {
		return slices.Collect(maps.Keys(nameables))
	}
	missing := make([]string, 0, min(len(nameables), len(replacements)))
	for k := range nameables {
		if _, exists := replacements[k]; !exists {
			missing = append(missing, k)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return missing
}
