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

type traitModifiersPanel struct {
	editorListPanel[*gurps.TraitModifier]
}

func newTraitModifiersPanel(cmdRoot Rebuildable, owner gurps.DataOwner, modifiers *[]*gurps.TraitModifier) *traitModifiersPanel {
	p := &traitModifiersPanel{}
	p.init(p, owner, modifiers, NewTraitModifiersProvider(p, true), "trait-modifiers-"+uuid.New().String())
	p.installNewItemHandler(cmdRoot, NewTraitModifierItemID, NoItemVariant)
	p.installNewItemHandler(cmdRoot, NewTraitContainerModifierItemID, ContainerItemVariant)
	return p
}

func (p *traitModifiersPanel) TraitModifierList() []*gurps.TraitModifier {
	return *p.list
}

func (p *traitModifiersPanel) SetTraitModifierList(list []*gurps.TraitModifier) {
	p.setList(list)
}
