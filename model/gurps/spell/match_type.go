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

package spell

import (
	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/toolbox/log/jot"
)

// MatchForType applies the matcher and returns the result.
func (enum MatchType) MatchForType(matcher criteria.String, name, powerSource string, colleges []string) bool {
	switch enum {
	case AllColleges:
		return true
	case CollegeName:
		return matcher.MatchesList(colleges...)
	case PowerSource:
		return matcher.Matches(powerSource)
	case Spell:
		return matcher.Matches(name)
	default:
		jot.Warnf("unhandled spell match type: %d", enum)
		return false
	}
}
