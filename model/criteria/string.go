/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package criteria

import (
	"github.com/richardwilkes/json"
)

// String holds the criteria for matching a string.
type String struct {
	StringData
}

// StringData holds the criteria for matching a string that should be written to disk.
type StringData struct {
	Compare   StringCompareType `json:"compare,omitempty"`
	Qualifier string            `json:"qualifier,omitempty"`
}

// ShouldOmit implements json.Omitter.
func (s String) ShouldOmit() bool {
	return s.Compare.EnsureValid() == Any
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *String) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &s.StringData)
	s.Compare = s.Compare.EnsureValid()
	return err
}

// Matches performs a comparison and returns true if the data matches.
func (s String) Matches(value ...string) bool {
	for _, one := range value {
		if s.Compare.Matches(s.Qualifier, one) {
			return true
		}
	}
	return s.Compare.Matches(s.Qualifier, "")
}

func (s String) String() string {
	return s.Compare.Describe(s.Qualifier)
}
