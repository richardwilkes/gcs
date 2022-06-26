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

package gurps

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
)

// String constants
var (
	NoAdditionalModifiers = i18n.Text("No additional modifiers")
	IncludesModifiersFrom = i18n.Text("Includes modifiers from")
	PageRefTooltipText    = i18n.Text(`A reference to the book and page the item appears on e.g. B22 would refer to "Basic Set", page 22`)
)

func convertOldCategoriesToTags(tags, categories []string) []string {
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
