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
	"encoding/json/jsontext"
	"encoding/json/v2"
	"regexp"
	"strings"
)

var collegeSepRegex = regexp.MustCompile(`(\s+or\s+)|/`)

// CollegeList holds a list of college names. This exists solely due to legacy file formats that stored this as a single
// string with ' or ' or '/' separating the colleges. We need to be able to load both types.
type CollegeList []string

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (c *CollegeList) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var list []string
	if dec.PeekKind() == '"' { // old format was a single string
		var str string
		if err := json.UnmarshalDecode(dec, &str); err != nil {
			return err
		}
		list = collegeSepRegex.Split(str, -1)
	} else if err := json.UnmarshalDecode(dec, &list); err != nil {
		return err
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
