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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

const equipmentModifierRefKey = "equipment_modifier"

var _ TableProvider[*gurps.EquipmentModifier] = &modifiersProvider[*gurps.EquipmentModifier]{}

// NewEquipmentModifiersProvider creates a new table provider for equipment modifiers.
func NewEquipmentModifiersProvider(provider gurps.EquipmentModifierListProvider, forEditor bool) TableProvider[*gurps.EquipmentModifier] {
	return newModifiersProvider(provider, provider.EquipmentModifierList, provider.SetEquipmentModifierList,
		gurps.EquipmentModifierHeaderData, forEditor, modifierProviderSpec[*gurps.EquipmentModifier]{
			refKey:            equipmentModifierRefKey,
			dragKey:           equipmentModifierDragKey,
			dragSVG:           svg.GCSEquipmentModifiers,
			singular:          i18n.Text("Equipment Modifier"),
			plural:            i18n.Text("Equipment Modifiers"),
			enabledColumn:     gurps.EquipmentModifierEnabledColumn,
			descriptionColumn: gurps.EquipmentModifierDescriptionColumn,
			columns: []int{
				gurps.EquipmentModifierDescriptionColumn,
				gurps.EquipmentModifierTechLevelColumn,
				gurps.EquipmentModifierCostColumn,
				gurps.EquipmentModifierWeightColumn,
				gurps.EquipmentModifierTagsColumn,
				gurps.EquipmentModifierReferenceColumn,
			},
			libSrcColumn: gurps.EquipmentModifierLibSrcColumn,
			newItem:      gurps.NewEquipmentModifier,
			edit:         EditEquipmentModifier,
			menuActions:  []*unison.Action{newEquipmentModifierAction, newEquipmentContainerModifierAction},
		})
}
