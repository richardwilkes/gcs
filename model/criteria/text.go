// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package criteria

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"hash"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// Text holds the criteria for matching text.
type Text struct {
	TextData
}

// TextData holds the criteria for matching text that should be written to disk.
type TextData struct {
	Compare   StringComparison `json:"compare,omitzero"`
	Qualifier string           `json:"qualifier,omitzero"`
}

// IsZero implements json.isZero.
func (t Text) IsZero() bool {
	return t.Compare.EnsureValid() == AnyText
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (t *Text) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	err := json.UnmarshalDecode(dec, &t.TextData)
	t.Compare = t.Compare.EnsureValid()
	return err
}

// Matches performs a comparison and returns true if the data matches.
func (t Text) Matches(replacements map[string]string, value string) bool {
	return t.Compare.Matches(nameable.Apply(t.Qualifier, replacements), value)
}

// MatchesList performs a comparison and returns true if the data matches. The qualifier may hold a comma-separated
// list of qualifiers; for the positive comparison types (e.g. "is", "contains") a match against any one of them is
// sufficient, while for the negative comparison types (e.g. "is not", "does not contain") every value must fail to
// match all of them.
func (t Text) MatchesList(replacements map[string]string, value ...string) bool {
	qualifiers := splitQualifiers(nameable.Apply(t.Qualifier, replacements))
	if len(value) == 0 {
		value = []string{""}
	}
	if t.Compare.IsNotType() {
		for _, one := range value {
			for _, qualifier := range qualifiers {
				if !t.Compare.Matches(qualifier, one) {
					return false
				}
			}
		}
		return true
	}
	for _, one := range value {
		for _, qualifier := range qualifiers {
			if t.Compare.Matches(qualifier, one) {
				return true
			}
		}
	}
	return false
}

// splitQualifiers splits a qualifier on commas, trimming surrounding whitespace and dropping empty entries. If nothing
// remains, a single empty qualifier is returned so that comparisons still function.
func splitQualifiers(qualifier string) []string {
	parts := strings.Split(qualifier, ",")
	list := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			list = append(list, part)
		}
	}
	if len(list) == 0 {
		return []string{""}
	}
	return list
}

func (t Text) String(replacements map[string]string) string {
	return t.Compare.Describe(nameable.Apply(t.Qualifier, replacements))
}

// StringWithPrefix returns a string representation of this criteria with a prefix.
func (t Text) StringWithPrefix(replacements map[string]string, prefix, notPrefix string) string {
	return t.Compare.DescribeWithPrefix(prefix, notPrefix, nameable.Apply(t.Qualifier, replacements))
}

// Hash writes this object's contents into the hasher.
func (t Text) Hash(h hash.Hash) {
	if t.IsZero() {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.StringWithLen(h, t.Compare)
	xhash.StringWithLen(h, t.Qualifier)
}
