// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package levelcap

// MatchesSkills returns true if the target includes skills.
func (enum Target) MatchesSkills() bool {
	return enum == Skills || enum == SkillsAndSpells
}

// MatchesSpells returns true if the target includes spells.
func (enum Target) MatchesSpells() bool {
	return enum == Spells || enum == SkillsAndSpells
}
