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
	"github.com/google/uuid"
	"github.com/richardwilkes/gcs/v5/model/gurps"
)

type equipmentModifiersPanel struct {
	editorListPanel[*gurps.EquipmentModifier]
	replacements *editedModifierReplacements
}

// newEquipmentModifiersPanel returns a panel for the modifiers of the equipment being edited, which are kept pointed at
// it. data is the editor's data, whose replacements take any that a modifier joining the list carries from an older
// version, and give them up again should it leave the list (see followEditedModifiers).
func newEquipmentModifiersPanel(cmdRoot Rebuildable, owner gurps.DataOwner, equipment *gurps.Equipment, modifiers *[]*gurps.EquipmentModifier, data replacementsHolder) *equipmentModifiersPanel {
	p := &equipmentModifiersPanel{replacements: &editedModifierReplacements{data: data}}
	p.init(p, cmdRoot, owner, modifiers, NewEquipmentModifiersProvider(p, true), "equipment-modifiers-"+uuid.New().String())
	p.attach = func(list []*gurps.EquipmentModifier) {
		attachEditedModifiers(equipment, p.replacements, list,
			func(mod *gurps.EquipmentModifier) *map[string]string { return &mod.Replacements })
	}
	p.installNewItemHandler(cmdRoot, NewEquipmentModifierItemID, NoItemVariant)
	p.installNewItemHandler(cmdRoot, NewEquipmentContainerModifierItemID, ContainerItemVariant)
	p.installNewItemHandler(cmdRoot, NewEquipmentModifierChoiceItemID, ChoiceContainerItemVariant)
	return p
}

func (p *equipmentModifiersPanel) EquipmentModifierList() []*gurps.EquipmentModifier {
	return *p.list
}

func (p *equipmentModifiersPanel) SetEquipmentModifierList(list []*gurps.EquipmentModifier) {
	p.setList(list)
	followEditedModifiers(p.replacements, list)
}
