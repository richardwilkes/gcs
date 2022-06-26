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

// EquipmentModifierData holds the EquipmentModifier data that is written to disk.
type EquipmentModifierData struct {
	ContainerBase[*EquipmentModifier]
	EquipmentModifierEditData
}

// Kind returns the kind of data.
func (d *EquipmentModifierData) Kind() string {
	return d.kind(i18n.Text("Equipment Modifier"))
}

// ClearUnusedFieldsForType zeroes out the fields that are not applicable to this type (container vs not-container).
func (d *EquipmentModifierData) ClearUnusedFieldsForType() {
	d.clearUnusedFields()
	if d.Container() {
		d.CostType = 0
		d.WeightType = 0
		d.Disabled = false
		d.TechLevel = ""
		d.CostAmount = ""
		d.WeightAmount = ""
		d.Features = nil
	}
}
