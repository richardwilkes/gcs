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
	"github.com/richardwilkes/unison"
)

type traitModifierListProvider struct {
	fileListProvider[*gurps.TraitModifier]
}

func (p *traitModifierListProvider) TraitModifierList() []*gurps.TraitModifier {
	return p.rows()
}

func (p *traitModifierListProvider) SetTraitModifierList(list []*gurps.TraitModifier) {
	p.setRows(list)
}

// NewTraitModifierTableDockableFromFile loads a list of trait modifiers from a file and creates a new
// unison.Dockable for them.
func NewTraitModifierTableDockableFromFile(filePath string) (unison.Dockable, error) {
	return openDockableFromFile(filePath, gurps.NewTraitModifiersFromFile, NewTraitModifierTableDockable)
}

// NewTraitModifierTableDockable creates a new unison.Dockable for trait modifier list files.
func NewTraitModifierTableDockable(filePath string, modifiers []*gurps.TraitModifier) *TableDockable[*gurps.TraitModifier] {
	provider := &traitModifierListProvider{list: modifiers}
	return NewTableDockable(filePath, gurps.TraitModifiersExt,
		NewTraitModifiersProvider(provider, false),
		func(path string) error { return gurps.SaveTraitModifiers(provider.TraitModifierList(), path) },
		NewTraitModifierItemID, NewTraitContainerModifierItemID)
}
