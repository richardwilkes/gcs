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
	replacements *editedModifierReplacements
}

// newTraitModifiersPanel returns a panel for the modifiers of the trait being edited, which are kept pointed at it.
// data is the editor's data, whose replacements take any that a modifier joining the list carries from an older
// version, and give them up again should it leave the list (see followEditedModifiers).
func newTraitModifiersPanel(cmdRoot Rebuildable, owner gurps.DataOwner, trait *gurps.Trait, modifiers *[]*gurps.TraitModifier, data replacementsHolder) *traitModifiersPanel {
	p := &traitModifiersPanel{replacements: &editedModifierReplacements{data: data}}
	p.init(p, cmdRoot, owner, modifiers, NewTraitModifiersProvider(p, true), "trait-modifiers-"+uuid.New().String())
	p.attach = func(list []*gurps.TraitModifier) {
		attachEditedModifiers(trait, p.replacements, list,
			func(mod *gurps.TraitModifier) *map[string]string { return &mod.Replacements })
	}
	p.installNewItemHandler(cmdRoot, NewTraitModifierItemID, NoItemVariant)
	p.installNewItemHandler(cmdRoot, NewTraitContainerModifierItemID, ContainerItemVariant)
	p.installNewItemHandler(cmdRoot, NewTraitModifierChoiceItemID, ChoiceContainerItemVariant)
	return p
}

func (p *traitModifiersPanel) TraitModifierList() []*gurps.TraitModifier {
	return *p.list
}

func (p *traitModifiersPanel) SetTraitModifierList(list []*gurps.TraitModifier) {
	p.setList(list)
	followEditedModifiers(p.replacements, list)
}
