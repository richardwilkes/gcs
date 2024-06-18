// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
)

// InvalidFileDataMsg returns a message indicating that the file data is invalid.
func InvalidFileDataMsg() string {
	return i18n.Text("Invalid file data.")
}

// UnexpectedFileDataMsg returns a message indicating that the file data is not what was expected.
func UnexpectedFileDataMsg() string {
	return i18n.Text("This file does not contain the expected data.")
}

func noAdditionalModifiers() string {
	return i18n.Text("No additional modifiers")
}

func includesModifiersFrom() string {
	return i18n.Text("Includes modifiers from")
}

// ConvertOldCategoriesToTags converts the old categories to tags.
func ConvertOldCategoriesToTags(tags, categories []string) []string {
	if categories == nil {
		return tags
	}
	for _, one := range categories {
		parts := strings.Split(one, "/")
		for _, part := range parts {
			if part = strings.TrimSpace(part); part != "" {
				if !txt.CaselessSliceContains(tags, part) {
					tags = append(tags, part)
				}
			}
		}
	}
	return tags
}
