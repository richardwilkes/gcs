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
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/unison"
)

var (
	reactionsColMap = map[int]int{
		0: model.ConditionalModifierValueColumn,
		1: model.ConditionalModifierDescriptionColumn,
	}
	_ TableProvider[*model.ConditionalModifier] = &reactionModProvider{}
)

type reactionModProvider struct {
	table    *unison.Table[*Node[*model.ConditionalModifier]]
	provider model.ReactionModifierListProvider
}

// NewReactionModifiersProvider creates a new table provider for reaction modifiers.
func NewReactionModifiersProvider(provider model.ReactionModifierListProvider) TableProvider[*model.ConditionalModifier] {
	return &reactionModProvider{
		provider: provider,
	}
}

func (p *reactionModProvider) RefKey() string {
	return model.BlockLayoutReactionsKey
}

func (p *reactionModProvider) AllTags() []string {
	return nil
}

func (p *reactionModProvider) SetTable(table *unison.Table[*Node[*model.ConditionalModifier]]) {
	p.table = table
}

func (p *reactionModProvider) RootRowCount() int {
	return len(p.provider.Reactions())
}

func (p *reactionModProvider) RootRows() []*Node[*model.ConditionalModifier] {
	data := p.provider.Reactions()
	rows := make([]*Node[*model.ConditionalModifier], 0, len(data))
	for _, one := range data {
		rows = append(rows, NewNode[*model.ConditionalModifier](p.table, nil, conditionalModifierColMap, one, true))
	}
	return rows
}

func (p *reactionModProvider) SetRootRows(_ []*Node[*model.ConditionalModifier]) {
}

func (p *reactionModProvider) RootData() []*model.ConditionalModifier {
	return p.provider.Reactions()
}

func (p *reactionModProvider) SetRootData(_ []*model.ConditionalModifier) {
}

func (p *reactionModProvider) Entity() *model.Entity {
	return p.provider.Entity()
}

func (p *reactionModProvider) DragKey() string {
	return "reaction_modifier"
}

func (p *reactionModProvider) DragSVG() *unison.SVG {
	return nil
}

func (p *reactionModProvider) DropShouldMoveData(_, _ *unison.Table[*Node[*model.ConditionalModifier]]) bool {
	// Not used
	return false
}

func (p *reactionModProvider) ProcessDropData(_, _ *unison.Table[*Node[*model.ConditionalModifier]]) {
}

func (p *reactionModProvider) AltDropSupport() *AltDropSupport {
	return nil
}

func (p *reactionModProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Reaction Modifier"), i18n.Text("Reaction Modifiers")
}

func (p *reactionModProvider) Headers() []unison.TableColumnHeader[*Node[*model.ConditionalModifier]] {
	var headers []unison.TableColumnHeader[*Node[*model.ConditionalModifier]]
	for i := 0; i < len(reactionsColMap); i++ {
		switch conditionalModifierColMap[i] {
		case model.ConditionalModifierValueColumn:
			headers = append(headers, NewEditorListHeader[*model.ConditionalModifier]("±", i18n.Text("Modifier"), true))
		case model.ConditionalModifierDescriptionColumn:
			headers = append(headers, NewEditorListHeader[*model.ConditionalModifier](i18n.Text("Reaction"), "", true))
		default:
			jot.Fatalf(1, "invalid reaction modifier column: %d", reactionsColMap[i])
		}
	}
	return headers
}

func (p *reactionModProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*model.ConditionalModifier]]) {
}

func (p *reactionModProvider) HierarchyColumnIndex() int {
	return -1
}

func (p *reactionModProvider) ExcessWidthColumnIndex() int {
	for k, v := range conditionalModifierColMap {
		if v == model.ConditionalModifierDescriptionColumn {
			return k
		}
	}
	return 0
}

func (p *reactionModProvider) OpenEditor(_ Rebuildable, _ *unison.Table[*Node[*model.ConditionalModifier]]) {
}

func (p *reactionModProvider) CreateItem(_ Rebuildable, _ *unison.Table[*Node[*model.ConditionalModifier]], _ ItemVariant) {
}

func (p *reactionModProvider) Serialize() ([]byte, error) {
	return nil, errs.New("not allowed")
}

func (p *reactionModProvider) Deserialize(_ []byte) error {
	return errs.New("not allowed")
}

func (p *reactionModProvider) ContextMenuItems() []ContextMenuItem {
	return nil
}
