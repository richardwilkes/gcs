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

package model

import (
	"github.com/richardwilkes/json"
)

// StringCriteria holds the criteria for matching a string.
type StringCriteria struct {
	StringCriteriaData
}

// StringCriteriaData holds the criteria for matching a string that should be written to disk.
type StringCriteriaData struct {
	Compare   StringCompareType `json:"compare,omitempty"`
	Qualifier string            `json:"qualifier,omitempty"`
}

// ShouldOmit implements json.Omitter.
func (s StringCriteria) ShouldOmit() bool {
	return s.Compare.EnsureValid() == AnyString
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *StringCriteria) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &s.StringCriteriaData)
	s.Compare = s.Compare.EnsureValid()
	return err
}

// Matches performs a comparison and returns true if the data matches.
func (s StringCriteria) Matches(value string) bool {
	return s.Compare.Matches(s.Qualifier, value)
}

// MatchesList performs a comparison and returns true if the data matches.
func (s StringCriteria) MatchesList(value ...string) bool {
	if len(value) == 0 {
		return s.Compare.Matches(s.Qualifier, "")
	}
	matches := 0
	for _, one := range value {
		if s.Compare.Matches(s.Qualifier, one) {
			matches++
		}
	}
	switch s.Compare {
	case AnyString, IsString, ContainsString, StartsWithString, EndsWithString:
		return matches > 0
	case IsNotString, DoesNotContainString, DoesNotStartWithString, DoesNotEndWithString:
		return matches == len(value)
	default:
		return matches > 0
	}
}

func (s StringCriteria) String() string {
	return s.Compare.Describe(s.Qualifier)
}
