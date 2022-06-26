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
	"github.com/richardwilkes/toolbox/i18n"
)

// SpellData holds the Spell data that is written to disk.
type SpellData struct {
	ContainerBase[*Spell]
	SpellEditData
}

// Kind returns the kind of data.
func (d *SpellData) Kind() string {
	return d.kind(i18n.Text("Spell"))
}

// ClearUnusedFieldsForType zeroes out the fields that are not applicable to this type (container vs not-container).
func (d *SpellData) ClearUnusedFieldsForType() {
	d.clearUnusedFields()
	if d.Container() {
		d.TechLevel = nil
		d.Difficulty = AttributeDifficulty{omit: true}
		d.College = nil
		d.PowerSource = ""
		d.Class = ""
		d.Resist = ""
		d.CastingCost = ""
		d.MaintenanceCost = ""
		d.CastingTime = ""
		d.Duration = ""
		d.RitualSkillName = ""
		d.RitualPrereqCount = 0
		d.Points = 0
		d.Prereq = nil
		d.Weapons = nil
	} else {
		d.Difficulty.omit = false
	}
}
