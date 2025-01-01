// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

type traitModifiersPanel struct {
	unison.Panel
	owner     gurps.DataOwner
	modifiers *[]*gurps.TraitModifier
	provider  TableProvider[*gurps.TraitModifier]
	table     *unison.Table[*Node[*gurps.TraitModifier]]
}

func newTraitModifiersPanel(owner gurps.DataOwner, modifiers *[]*gurps.TraitModifier) *traitModifiersPanel {
	p := &traitModifiersPanel{
		owner:     owner,
		modifiers: modifiers,
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{Columns: 1})
	p.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Fill,
		HGrab:  true,
	})
	p.SetBorder(unison.NewLineBorder(unison.ThemeAboveSurface, 0, unison.NewUniformInsets(1), false))
	p.provider = NewTraitModifiersProvider(p, true)
	p.table = newEditorTable(p.AsPanel(), p.provider)
	p.table.RefKey = "trait-modifiers-" + uuid.New().String()
	return p
}

func (p *traitModifiersPanel) DataOwner() gurps.DataOwner {
	return p.owner
}

func (p *traitModifiersPanel) TraitModifierList() []*gurps.TraitModifier {
	return *p.modifiers
}

func (p *traitModifiersPanel) SetTraitModifierList(list []*gurps.TraitModifier) {
	*p.modifiers = list
	sel := p.table.CopySelectionMap()
	p.table.SyncToModel()
	p.table.SetSelectionMap(sel)
}
