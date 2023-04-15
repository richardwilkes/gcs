/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/toolbox/log/jot"
)

// MatchForType applies the matcher and returns the result.
func (enum SpellMatchType) MatchForType(matcher StringCriteria, name, powerSource string, colleges []string) bool {
	switch enum {
	case AllCollegesSpellMatchType:
		return true
	case CollegeNameSpellMatchType:
		return matcher.MatchesList(colleges...)
	case PowerSourceSpellMatchType:
		return matcher.Matches(powerSource)
	case NameSpellMatchType:
		return matcher.Matches(name)
	default:
		jot.Warnf("unhandled spell match type: %d", enum)
		return false
	}
}
