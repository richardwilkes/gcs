/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ux

import (
	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

var _ TableProvider[*model.ConditionalModifier] = &condModProvider{}

type condModProvider struct {
	table    *unison.Table[*Node[*model.ConditionalModifier]]
	provider model.ConditionalModifierListProvider
}

// NewConditionalModifiersProvider creates a new table provider for conditional modifiers.
func NewConditionalModifiersProvider(provider model.ConditionalModifierListProvider) TableProvider[*model.ConditionalModifier] {
	return &condModProvider{provider: provider}
}

func (p *condModProvider) RefKey() string {
	return model.BlockLayoutConditionalModifiersKey
}

func (p *condModProvider) AllTags() []string {
	return nil
}

func (p *condModProvider) SetTable(table *unison.Table[*Node[*model.ConditionalModifier]]) {
	p.table = table
}

func (p *condModProvider) RootRowCount() int {
	return len(p.provider.ConditionalModifiers())
}

func (p *condModProvider) RootRows() []*Node[*model.ConditionalModifier] {
	data := p.provider.ConditionalModifiers()
	rows := make([]*Node[*model.ConditionalModifier], 0, len(data))
	for _, one := range data {
		rows = append(rows, NewNode[*model.ConditionalModifier](p.table, nil, one, true))
	}
	return rows
}

func (p *condModProvider) SetRootRows(_ []*Node[*model.ConditionalModifier]) {
}

func (p *condModProvider) RootData() []*model.ConditionalModifier {
	return p.provider.ConditionalModifiers()
}

func (p *condModProvider) SetRootData(_ []*model.ConditionalModifier) {
}

func (p *condModProvider) Entity() *model.Entity {
	return p.provider.Entity()
}

func (p *condModProvider) DragKey() string {
	return "conditional_modifier"
}

func (p *condModProvider) DragSVG() *unison.SVG {
	return nil
}

func (p *condModProvider) DropShouldMoveData(_, _ *unison.Table[*Node[*model.ConditionalModifier]]) bool {
	// Not used
	return false
}

func (p *condModProvider) ProcessDropData(_, _ *unison.Table[*Node[*model.ConditionalModifier]]) {
}

func (p *condModProvider) AltDropSupport() *AltDropSupport {
	return nil
}

func (p *condModProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Conditional Modifier"), i18n.Text("Conditional Modifiers")
}

func (p *condModProvider) Headers() []unison.TableColumnHeader[*Node[*model.ConditionalModifier]] {
	return DisableSorting([]unison.TableColumnHeader[*Node[*model.ConditionalModifier]]{
		NewEditorListHeader[*model.ConditionalModifier]("±", i18n.Text("Modifier"), true),
		NewEditorListHeader[*model.ConditionalModifier](i18n.Text("Condition"), "", true),
	})
}

func (p *condModProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*model.ConditionalModifier]]) {
}

func (p *condModProvider) ColumnIDs() []int {
	return []int{
		model.ConditionalModifierValueColumn,
		model.ConditionalModifierDescriptionColumn,
	}
}

func (p *condModProvider) HierarchyColumnID() int {
	return -1
}

func (p *condModProvider) ExcessWidthColumnID() int {
	return model.ConditionalModifierDescriptionColumn
}

func (p *condModProvider) OpenEditor(_ Rebuildable, _ *unison.Table[*Node[*model.ConditionalModifier]]) {
}

func (p *condModProvider) CreateItem(_ Rebuildable, _ *unison.Table[*Node[*model.ConditionalModifier]], _ ItemVariant) {
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
