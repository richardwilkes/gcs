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

	"github.com/richardwilkes/gcs/v5/model/gurps/gid"
	"github.com/richardwilkes/toolbox/i18n"
)

// SkillData holds the Skill data that is written to disk.
type SkillData struct {
	ContainerBase[*Skill]
	SkillEditData
}

// Kind returns the kind of data.
func (d *SkillData) Kind() string {
	if strings.HasPrefix(d.Type, gid.Skill) {
		return d.kind(i18n.Text("Skill"))
	}
	return d.kind(i18n.Text("Technique"))
}

// ClearUnusedFieldsForType zeroes out the fields that are not applicable to this type (container vs not-container).
func (d *SkillData) ClearUnusedFieldsForType() {
	d.clearUnusedFields()
	if d.Container() {
		d.Specialization = ""
		d.TechLevel = nil
		d.Difficulty = AttributeDifficulty{omit: true}
		d.Points = 0
		d.EncumbrancePenaltyMultiplier = 0
		d.DefaultedFrom = nil
		d.Defaults = nil
		d.TechniqueDefault = nil
		d.TechniqueLimitModifier = nil
		d.Prereq = nil
		d.Weapons = nil
		d.Features = nil
	} else {
		d.Difficulty.omit = false
	}
}
