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

var _ TableProvider[*gurps.ConditionalModifier] = &reactionModProvider{}

type reactionModProvider struct {
	table    *unison.Table[*Node[*gurps.ConditionalModifier]]
	provider gurps.ReactionModifierListProvider
}

// NewReactionModifiersProvider creates a new table provider for reaction modifiers.
func NewReactionModifiersProvider(provider gurps.ReactionModifierListProvider) TableProvider[*gurps.ConditionalModifier] {
	return &reactionModProvider{provider: provider}
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
		rows = append(rows, NewNode(p.table, nil, one, true))
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

func (p *reactionModProvider) DataOwner() gurps.DataOwner {
	return p.provider.DataOwner()
}

func (p *reactionModProvider) DragKey() string {
	return "reaction_modifier"
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
	ids := p.ColumnIDs()
	headers := make([]unison.TableColumnHeader[*Node[*gurps.ConditionalModifier]], 0, len(ids))
	for _, id := range ids {
		headers = append(headers, headerFromData[*gurps.ConditionalModifier](gurps.ReactionModifiersHeaderData(id), true))
	}
	return DisableSorting(headers)
}

func (p *reactionModProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*gurps.ConditionalModifier]]) {
}

func (p *reactionModProvider) ColumnIDs() []int {
	return []int{
		gurps.ConditionalModifierValueColumn,
		gurps.ConditionalModifierDescriptionColumn,
	}
}

func (p *reactionModProvider) HierarchyColumnID() int {
	return -1
}

func (p *reactionModProvider) ExcessWidthColumnID() int {
	return gurps.ConditionalModifierDescriptionColumn
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
