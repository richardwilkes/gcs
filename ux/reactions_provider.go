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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/gid"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/unison"
)

var (
	reactionsColMap = map[int]int{
		0: gurps.ConditionalModifierValueColumn,
		1: gurps.ConditionalModifierDescriptionColumn,
	}
	_ TableProvider[*gurps.ConditionalModifier] = &reactionModProvider{}
)

type reactionModProvider struct {
	table    *unison.Table[*Node[*gurps.ConditionalModifier]]
	provider gurps.ReactionModifierListProvider
}

// NewReactionModifiersProvider creates a new table provider for reaction modifiers.
func NewReactionModifiersProvider(provider gurps.ReactionModifierListProvider) TableProvider[*gurps.ConditionalModifier] {
	return &reactionModProvider{
		provider: provider,
	}
}

func (p *reactionModProvider) RefKey() string {
	return gurps.BlockLayoutReactionsKey
}

func (p *reactionModProvider) AllTags() []string {
	return nil
}

func (p *reactionModProvider) SetTable(table *unison.Table[*Node[*gurps.ConditionalModifier]]) {
	p.table = table
}

func (p *reactionModProvider) RootRowCount() int {
	return len(p.provider.Reactions())
}

func (p *reactionModProvider) RootRows() []*Node[*gurps.ConditionalModifier] {
	data := p.provider.Reactions()
	rows := make([]*Node[*gurps.ConditionalModifier], 0, len(data))
	for _, one := range data {
		rows = append(rows, NewNode[*gurps.ConditionalModifier](p.table, nil, conditionalModifierColMap, one, true))
	}
	return rows
}

func (p *reactionModProvider) SetRootRows(_ []*Node[*gurps.ConditionalModifier]) {
}

func (p *reactionModProvider) RootData() []*gurps.ConditionalModifier {
	return p.provider.Reactions()
}

func (p *reactionModProvider) SetRootData(_ []*gurps.ConditionalModifier) {
}

func (p *reactionModProvider) Entity() *gurps.Entity {
	return p.provider.Entity()
}

func (p *reactionModProvider) DragKey() string {
	return gid.ReactionModifier
}

func (p *reactionModProvider) DragSVG() *unison.SVG {
	return nil
}

func (p *reactionModProvider) DropShouldMoveData(_, _ *unison.Table[*Node[*gurps.ConditionalModifier]]) bool {
	// Not used
	return false
}

func (p *reactionModProvider) ProcessDropData(_, _ *unison.Table[*Node[*gurps.ConditionalModifier]]) {
}

func (p *reactionModProvider) AltDropSupport() *AltDropSupport {
	return nil
}

func (p *reactionModProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Reaction Modifier"), i18n.Text("Reaction Modifiers")
}

func (p *reactionModProvider) Headers() []unison.TableColumnHeader[*Node[*gurps.ConditionalModifier]] {
	var headers []unison.TableColumnHeader[*Node[*gurps.ConditionalModifier]]
	for i := 0; i < len(reactionsColMap); i++ {
		switch conditionalModifierColMap[i] {
		case gurps.ConditionalModifierValueColumn:
			headers = append(headers, NewEditorListHeader[*gurps.ConditionalModifier]("±", i18n.Text("Modifier"), true))
		case gurps.ConditionalModifierDescriptionColumn:
			headers = append(headers, NewEditorListHeader[*gurps.ConditionalModifier](i18n.Text("Reaction"), "", true))
		default:
			jot.Fatalf(1, "invalid reaction modifier column: %d", reactionsColMap[i])
		}
	}
	return headers
}

func (p *reactionModProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*gurps.ConditionalModifier]]) {
}

func (p *reactionModProvider) HierarchyColumnIndex() int {
	return -1
}

func (p *reactionModProvider) ExcessWidthColumnIndex() int {
	for k, v := range conditionalModifierColMap {
		if v == gurps.ConditionalModifierDescriptionColumn {
			return k
		}
	}
	return 0
}

func (p *reactionModProvider) OpenEditor(_ Rebuildable, _ *unison.Table[*Node[*gurps.ConditionalModifier]]) {
}

func (p *reactionModProvider) CreateItem(_ Rebuildable, _ *unison.Table[*Node[*gurps.ConditionalModifier]], _ ItemVariant) {
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
