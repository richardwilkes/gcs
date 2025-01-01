// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"regexp"
	"strings"

	"github.com/richardwilkes/json"
)

var collegeSepRegex = regexp.MustCompile(`(\s+or\s+)|/`)

// CollegeList holds a list of college names. This exists solely due to legacy file formats that stored this as a single
// string with ' or ' or '/' separating the colleges. We need to be able to load both types.
type CollegeList []string

// UnmarshalJSON implements json.Unmarshaler.
func (c *CollegeList) UnmarshalJSON(data []byte) error {
	var list []string
	if err := json.Unmarshal(data, &list); err != nil {
		// Try older format
		var str string
		if err = json.Unmarshal(data, &str); err != nil {
			return err
		}
		list = collegeSepRegex.Split(str, -1)
	}
	pruned := make([]string, 0, len(list))
	for _, one := range list {
		if one = strings.TrimSpace(one); one != "" {
			pruned = append(pruned, one)
		}
	}
	*c = pruned
	return nil
}
