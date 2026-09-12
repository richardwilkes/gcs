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
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/unison"
)

var _ TableProvider[*gurps.ConditionalModifier] = &condModProvider{}

// condModProviderSpec captures the few things that differ between the conditional modifier and reaction modifier
// table providers. Both tables display read-only lists of gurps.ConditionalModifier rows.
type condModProviderSpec struct {
	refKey   string
	dragKey  *uti.DataType
	singular string
	plural   string
}

type condModProvider struct {
	listProvider[*gurps.ConditionalModifier]
	spec condModProviderSpec
}

// NewConditionalModifiersProvider creates a new table provider for conditional modifiers.
func NewConditionalModifiersProvider(provider gurps.ConditionalModifierListProvider) TableProvider[*gurps.ConditionalModifier] {
	return newCondModProvider(provider, provider.ConditionalModifiers, gurps.ConditionalModifiersHeaderData,
		condModProviderSpec{
			refKey:   gurps.BlockConditionalModifiersKey,
			dragKey:  conditionalModifierDragKey,
			singular: i18n.Text("Conditional Modifier"),
			plural:   i18n.Text("Conditional Modifiers"),
		})
}

// NewReactionModifiersProvider creates a new table provider for reaction modifiers.
func NewReactionModifiersProvider(provider gurps.ReactionModifierListProvider) TableProvider[*gurps.ConditionalModifier] {
	return newCondModProvider(provider, provider.Reactions, gurps.ReactionModifiersHeaderData,
		condModProviderSpec{
			refKey:   gurps.BlockReactionsKey,
			dragKey:  reactionModifierDragKey,
			singular: i18n.Text("Reaction Modifier"),
			plural:   i18n.Text("Reaction Modifiers"),
		})
}

func newCondModProvider(owner gurps.DataOwnerProvider, rows func() []*gurps.ConditionalModifier, headerData func(columnID int) gurps.HeaderData, spec condModProviderSpec) *condModProvider {
	p := &condModProvider{spec: spec}
	p.listProvider = listProvider[*gurps.ConditionalModifier]{
		dataOwner:  owner,
		list:       rows,
		setList:    func(_ []*gurps.ConditionalModifier) {}, // The rows are computed, so there is nothing to set.
		columnIDs:  p.ColumnIDs,
		headerData: headerData,
		forPage:    true,
	}
	return p
}

func (p *condModProvider) RefKey() string {
	return p.spec.refKey
}

func (p *condModProvider) DragKey() *uti.DataType {
	return p.spec.dragKey
}

func (p *condModProvider) DragSVG() *unison.SVG {
	return nil
}

func (p *condModProvider) DropShouldMoveData(_, _ *unison.Table[*Node[*gurps.ConditionalModifier]]) bool {
	// Not used
	return false
}

func (p *condModProvider) ItemNames() (singular, plural string) {
	return p.spec.singular, p.spec.plural
}

func (p *condModProvider) Headers() []unison.TableColumnHeader[*Node[*gurps.ConditionalModifier]] {
	return DisableSorting(p.listProvider.Headers())
}

func (p *condModProvider) ColumnIDs() []int {
	return []int{
		gurps.ConditionalModifierValueColumn,
		gurps.ConditionalModifierDescriptionColumn,
	}
}

func (p *condModProvider) HierarchyColumnID() int {
	return gurps.ConditionalModifierDescriptionColumn
}

func (p *condModProvider) ExcessWidthColumnID() int {
	return gurps.ConditionalModifierDescriptionColumn
}

func (p *condModProvider) OpenEditor(_ Rebuildable, _ *unison.Table[*Node[*gurps.ConditionalModifier]]) {
}

func (p *condModProvider) CreateItem(_ Rebuildable, _ *unison.Table[*Node[*gurps.ConditionalModifier]], _ ItemVariant) {
}

func (p *condModProvider) Serialize() ([]byte, error) {
	return nil, errs.New("not allowed")
}

func (p *condModProvider) Deserialize(_ []byte) error {
	return errs.New("not allowed")
}

func (p *condModProvider) ContextMenuItems() []ContextMenuItem {
	return nil
}
