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
}

func newEquipmentModifiersPanel(cmdRoot Rebuildable, owner gurps.DataOwner, modifiers *[]*gurps.EquipmentModifier) *equipmentModifiersPanel {
	p := &equipmentModifiersPanel{}
	p.init(p, owner, modifiers, NewEquipmentModifiersProvider(p, true), "equipment-modifiers-"+uuid.New().String())
	p.installNewItemHandler(cmdRoot, NewEquipmentModifierItemID, NoItemVariant)
	p.installNewItemHandler(cmdRoot, NewEquipmentContainerModifierItemID, ContainerItemVariant)
	return p
}

func (p *equipmentModifiersPanel) EquipmentModifierList() []*gurps.EquipmentModifier {
	return *p.list
}

func (p *equipmentModifiersPanel) SetEquipmentModifierList(list []*gurps.EquipmentModifier) {
	p.setList(list)
}
