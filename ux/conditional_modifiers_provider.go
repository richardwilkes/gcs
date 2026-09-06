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
	refKey     string
	dragKey    *uti.DataType
	singular   string
	plural     string
	headerData func(columnID int) gurps.HeaderData
	rows       func() []*gurps.ConditionalModifier
}

type condModProvider struct {
	table *unison.Table[*Node[*gurps.ConditionalModifier]]
	owner gurps.DataOwnerProvider
	spec  condModProviderSpec
}

// NewConditionalModifiersProvider creates a new table provider for conditional modifiers.
func NewConditionalModifiersProvider(provider gurps.ConditionalModifierListProvider) TableProvider[*gurps.ConditionalModifier] {
	return &condModProvider{
		owner: provider,
		spec: condModProviderSpec{
			refKey:     gurps.BlockConditionalModifiersKey,
			dragKey:    conditionalModifierDragKey,
			singular:   i18n.Text("Conditional Modifier"),
			plural:     i18n.Text("Conditional Modifiers"),
			headerData: gurps.ConditionalModifiersHeaderData,
			rows:       provider.ConditionalModifiers,
		},
	}
}

// NewReactionModifiersProvider creates a new table provider for reaction modifiers.
func NewReactionModifiersProvider(provider gurps.ReactionModifierListProvider) TableProvider[*gurps.ConditionalModifier] {
	return &condModProvider{
		owner: provider,
		spec: condModProviderSpec{
			refKey:     gurps.BlockReactionsKey,
			dragKey:    reactionModifierDragKey,
			singular:   i18n.Text("Reaction Modifier"),
			plural:     i18n.Text("Reaction Modifiers"),
			headerData: gurps.ReactionModifiersHeaderData,
			rows:       provider.Reactions,
		},
	}
}

func (p *condModProvider) RefKey() string {
	return p.spec.refKey
}

func (p *condModProvider) AllTags() []string {
	return nil
}

func (p *condModProvider) SetTable(table *unison.Table[*Node[*gurps.ConditionalModifier]]) {
	p.table = table
}

func (p *condModProvider) RootRowCount() int {
	return len(p.spec.rows())
}

func (p *condModProvider) RootRows() []*Node[*gurps.ConditionalModifier] {
	data := p.spec.rows()
	rows := make([]*Node[*gurps.ConditionalModifier], 0, len(data))
	for _, one := range data {
		rows = append(rows, NewNode(p.table, nil, one, true))
	}
	return rows
}

func (p *condModProvider) SetRootRows(_ []*Node[*gurps.ConditionalModifier]) {
}

func (p *condModProvider) RootData() []*gurps.ConditionalModifier {
	return p.spec.rows()
}

func (p *condModProvider) SetRootData(_ []*gurps.ConditionalModifier) {
}

func (p *condModProvider) DataOwner() gurps.DataOwner {
	return p.owner.DataOwner()
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

func (p *condModProvider) ProcessDropData(_, _ *unison.Table[*Node[*gurps.ConditionalModifier]]) {
}

func (p *condModProvider) AltDropSupport() *AltDropSupport {
	return nil
}

func (p *condModProvider) ItemNames() (singular, plural string) {
	return p.spec.singular, p.spec.plural
}

func (p *condModProvider) Headers() []unison.TableColumnHeader[*Node[*gurps.ConditionalModifier]] {
	ids := p.ColumnIDs()
	headers := make([]unison.TableColumnHeader[*Node[*gurps.ConditionalModifier]], 0, len(ids))
	for _, id := range ids {
		headers = append(headers, headerFromData[*gurps.ConditionalModifier](p.spec.headerData(id), true))
	}
	return DisableSorting(headers)
}

func (p *condModProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*gurps.ConditionalModifier]]) {
}

func (p *condModProvider) ColumnIDs() []int {
	return []int{
		gurps.ConditionalModifierValueColumn,
		gurps.ConditionalModifierDescriptionColumn,
	}
}

func (p *condModProvider) HierarchyColumnID() int {
	return -1
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
