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

const traitModifierRefKey = "trait_modifier"

var _ TableProvider[*gurps.TraitModifier] = &modifiersProvider[*gurps.TraitModifier]{}

// NewTraitModifiersProvider creates a new table provider for trait modifiers.
func NewTraitModifiersProvider(provider gurps.TraitModifierListProvider, forEditor bool) TableProvider[*gurps.TraitModifier] {
	return newModifiersProvider(provider, provider.TraitModifierList, provider.SetTraitModifierList,
		gurps.TraitModifierHeaderData, forEditor, modifierProviderSpec[*gurps.TraitModifier]{
			refKey:            traitModifierRefKey,
			dragKey:           traitModifierDragKey,
			dragSVG:           svg.GCSTraitModifiers,
			singular:          i18n.Text("Trait Modifier"),
			plural:            i18n.Text("Trait Modifiers"),
			enabledColumn:     gurps.TraitModifierEnabledColumn,
			descriptionColumn: gurps.TraitModifierDescriptionColumn,
			columns: []int{
				gurps.TraitModifierDescriptionColumn,
				gurps.TraitModifierCostColumn,
				gurps.TraitModifierTagsColumn,
				gurps.TraitModifierReferenceColumn,
			},
			libSrcColumn: gurps.TraitModifierLibSrcColumn,
			newItem:      gurps.NewTraitModifier,
			edit:         EditTraitModifier,
			menuActions:  []*unison.Action{newTraitModifierAction, newTraitContainerModifierAction},
		})
}
