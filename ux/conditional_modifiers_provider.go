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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

var _ TableProvider[*gurps.ConditionalModifier] = &condModProvider{}

type condModProvider struct {
	table    *unison.Table[*Node[*gurps.ConditionalModifier]]
	provider gurps.ConditionalModifierListProvider
}

// NewConditionalModifiersProvider creates a new table provider for conditional modifiers.
func NewConditionalModifiersProvider(provider gurps.ConditionalModifierListProvider) TableProvider[*gurps.ConditionalModifier] {
	return &condModProvider{provider: provider}
}

func (p *condModProvider) RefKey() string {
	return gurps.BlockLayoutConditionalModifiersKey
}

func (p *condModProvider) AllTags() []string {
	return nil
}

func (p *condModProvider) SetTable(table *unison.Table[*Node[*gurps.ConditionalModifier]]) {
	p.table = table
}

func (p *condModProvider) RootRowCount() int {
	return len(p.provider.ConditionalModifiers())
}

func (p *condModProvider) RootRows() []*Node[*gurps.ConditionalModifier] {
	data := p.provider.ConditionalModifiers()
	rows := make([]*Node[*gurps.ConditionalModifier], 0, len(data))
	for _, one := range data {
		rows = append(rows, NewNode(p.table, nil, one, true))
	}
	return rows
}

func (p *condModProvider) SetRootRows(_ []*Node[*gurps.ConditionalModifier]) {
}

func (p *condModProvider) RootData() []*gurps.ConditionalModifier {
	return p.provider.ConditionalModifiers()
}

func (p *condModProvider) SetRootData(_ []*gurps.ConditionalModifier) {
}

func (p *condModProvider) DataOwner() gurps.DataOwner {
	return p.provider.DataOwner()
}

func (p *condModProvider) DragKey() string {
	return "conditional_modifier"
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
	return i18n.Text("Conditional Modifier"), i18n.Text("Conditional Modifiers")
}

func (p *condModProvider) Headers() []unison.TableColumnHeader[*Node[*gurps.ConditionalModifier]] {
	ids := p.ColumnIDs()
	headers := make([]unison.TableColumnHeader[*Node[*gurps.ConditionalModifier]], 0, len(ids))
	for _, id := range ids {
		headers = append(headers, headerFromData[*gurps.ConditionalModifier](gurps.ConditionalModifiersHeaderData(id), true))
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
