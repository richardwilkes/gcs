// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"strings"
	"unicode/utf8"

	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xstrings"
)

// maxTradeNameLength is the most characters of item names a suggested transaction name runs to.
const maxTradeNameLength = 50

// tradeNames lists the names for a suggested transaction name, such as "Shortsword, Dagger". Once listing them all
// would run past maxTradeNameLength characters, as many as fit are listed and the rest are counted, as in "Shortsword,
// Dagger plus 3 other items"; a first name too long to fit on its own is cut down.
func tradeNames(names []string) string {
	if len(names) == 0 {
		return ""
	}
	for count := len(names); count > 0; count-- {
		if text := countOtherTradeItems(strings.Join(names[:count], i18n.Text(", ")), len(names)-count); fitsTradeName(text) {
			return text
		}
	}
	rest := len(names) - 1
	// Cutting a name down adds an ellipsis, which takes a character of the room left beside the count.
	room := max(maxTradeNameLength-utf8.RuneCountInString(countOtherTradeItems("", rest))-1, 1)
	return countOtherTradeItems(xstrings.Truncate(names[0], room, true), rest)
}

// countOtherTradeItems follows the names listed with a count of the rest, if there are any.
func countOtherTradeItems(listed string, rest int) string {
	switch rest {
	case 0:
		return listed
	case 1:
		return i18n.Text("%s plus 1 other item", listed)
	default:
		return i18n.Text("%s plus %d other items", listed, rest)
	}
}

// fitsTradeName returns true if the text is no longer than a suggested transaction name may run to.
func fitsTradeName(text string) bool {
	return utf8.RuneCountInString(text) <= maxTradeNameLength
}
