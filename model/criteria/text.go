// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package criteria

import (
	"hash"

	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
)

// Text holds the criteria for matching text.
type Text struct {
	TextData
}

// TextData holds the criteria for matching text that should be written to disk.
type TextData struct {
	Compare   StringComparison `json:"compare,omitempty"`
	Qualifier string           `json:"qualifier,omitempty"`
}

// ShouldOmit implements json.Omitter.
func (t Text) ShouldOmit() bool {
	return t.Compare.EnsureValid() == AnyText
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Text) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &t.TextData)
	t.Compare = t.Compare.EnsureValid()
	return err
}

// Matches performs a comparison and returns true if the data matches.
func (t Text) Matches(replacements map[string]string, value string) bool {
	return t.Compare.Matches(nameable.Apply(t.Qualifier, replacements), value)
}

// MatchesList performs a comparison and returns true if the data matches.
func (t Text) MatchesList(replacements map[string]string, value ...string) bool {
	qualifier := nameable.Apply(t.Qualifier, replacements)
	if len(value) == 0 {
		return t.Compare.Matches(qualifier, "")
	}
	matches := 0
	for _, one := range value {
		if t.Compare.Matches(qualifier, one) {
			matches++
		}
	}
	switch t.Compare {
	case AnyText, IsText, ContainsText, StartsWithText, EndsWithText:
		return matches > 0
	case IsNotText, DoesNotContainText, DoesNotStartWithText, DoesNotEndWithText:
		return matches == len(value)
	default:
		return matches > 0
	}
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
	if t.ShouldOmit() {
		hashhelper.Num8(h, uint8(255))
		return
	}
	hashhelper.String(h, t.Compare)
	hashhelper.String(h, t.Qualifier)
}
