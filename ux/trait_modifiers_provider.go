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
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/collection/dict"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
)

const traitModifierDragKey = "trait_modifier"

var _ TableProvider[*gurps.TraitModifier] = &traitModifiersProvider{}

type traitModifiersProvider struct {
	table     *unison.Table[*Node[*gurps.TraitModifier]]
	provider  gurps.TraitModifierListProvider
	forEditor bool
}

// NewTraitModifiersProvider creates a new table provider for trait modifiers.
func NewTraitModifiersProvider(provider gurps.TraitModifierListProvider, forEditor bool) TableProvider[*gurps.TraitModifier] {
	return &traitModifiersProvider{
		provider:  provider,
		forEditor: forEditor,
	}
}

func (p *traitModifiersProvider) RefKey() string {
	return traitModifierDragKey
}

func (p *traitModifiersProvider) AllTags() []string {
	set := make(map[string]struct{})
	gurps.Traverse(func(modifier *gurps.TraitModifier) bool {
		for _, tag := range modifier.Tags {
			set[tag] = struct{}{}
		}
		return false
	}, false, false, p.RootData()...)
	tags := dict.Keys(set)
	txt.SortStringsNaturalAscending(tags)
	return tags
}

func (p *traitModifiersProvider) SetTable(table *unison.Table[*Node[*gurps.TraitModifier]]) {
	p.table = table
}

func (p *traitModifiersProvider) RootRowCount() int {
	return len(p.provider.TraitModifierList())
}

func (p *traitModifiersProvider) RootRows() []*Node[*gurps.TraitModifier] {
	data := p.provider.TraitModifierList()
	rows := make([]*Node[*gurps.TraitModifier], 0, len(data))
	for _, one := range data {
		rows = append(rows, NewNode(p.table, nil, one, false))
	}
	return rows
}

func (p *traitModifiersProvider) SetRootRows(rows []*Node[*gurps.TraitModifier]) {
	p.provider.SetTraitModifierList(ExtractNodeDataFromList(rows))
}

func (p *traitModifiersProvider) RootData() []*gurps.TraitModifier {
	return p.provider.TraitModifierList()
}

func (p *traitModifiersProvider) SetRootData(data []*gurps.TraitModifier) {
	p.provider.SetTraitModifierList(data)
}

func (p *traitModifiersProvider) DataOwner() gurps.DataOwner {
	return p.provider.DataOwner()
}

func (p *traitModifiersProvider) DragKey() string {
	return traitModifierDragKey
}

func (p *traitModifiersProvider) DragSVG() *unison.SVG {
	return svg.GCSTraitModifiers
}

func (p *traitModifiersProvider) DropShouldMoveData(from, to *unison.Table[*Node[*gurps.TraitModifier]]) bool {
	return from == to
}

func (p *traitModifiersProvider) ProcessDropData(_, _ *unison.Table[*Node[*gurps.TraitModifier]]) {
}

func (p *traitModifiersProvider) AltDropSupport() *AltDropSupport {
	return nil
}

func (p *traitModifiersProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Trait Modifier"), i18n.Text("Trait Modifiers")
}

func (p *traitModifiersProvider) Headers() []unison.TableColumnHeader[*Node[*gurps.TraitModifier]] {
	ids := p.ColumnIDs()
	headers := make([]unison.TableColumnHeader[*Node[*gurps.TraitModifier]], 0, len(ids))
	for _, id := range ids {
		headers = append(headers, headerFromData[*gurps.TraitModifier](gurps.TraitModifierHeaderData(id), false))
	}
	return headers
}

func (p *traitModifiersProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*gurps.TraitModifier]]) {
}

func (p *traitModifiersProvider) ColumnIDs() []int {
	columnIDs := make([]int, 0, 5)
	if p.forEditor {
		columnIDs = append(columnIDs, gurps.TraitModifierEnabledColumn)
	}
	columnIDs = append(columnIDs,
		gurps.TraitModifierDescriptionColumn,
		gurps.TraitModifierCostColumn,
		gurps.TraitModifierTagsColumn,
		gurps.TraitModifierReferenceColumn,
	)
	if p.forEditor {
		columnIDs = append(columnIDs, gurps.TraitModifierLibSrcColumn)
	}
	return columnIDs
}

func (p *traitModifiersProvider) HierarchyColumnID() int {
	return gurps.TraitModifierDescriptionColumn
}

func (p *traitModifiersProvider) ExcessWidthColumnID() int {
	return gurps.TraitModifierDescriptionColumn
}

func (p *traitModifiersProvider) OpenEditor(owner Rebuildable, table *unison.Table[*Node[*gurps.TraitModifier]]) {
	OpenEditor(table, func(item *gurps.TraitModifier) {
		EditTraitModifier(owner, item)
	})
}

func (p *traitModifiersProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*gurps.TraitModifier]], variant ItemVariant) {
	item := gurps.NewTraitModifier(p.DataOwner(), nil, variant == ContainerItemVariant)
	InsertItems(owner, table, p.provider.TraitModifierList, p.provider.SetTraitModifierList,
		func(_ *unison.Table[*Node[*gurps.TraitModifier]]) []*Node[*gurps.TraitModifier] {
			return p.RootRows()
		}, item)
	EditTraitModifier(owner, item)
}

func (p *traitModifiersProvider) Serialize() ([]byte, error) {
	return jio.SerializeAndCompress(p.provider.TraitModifierList())
}

func (p *traitModifiersProvider) Deserialize(data []byte) error {
	var rows []*gurps.TraitModifier
	if err := jio.DecompressAndDeserialize(data, &rows); err != nil {
		return err
	}
	p.provider.SetTraitModifierList(rows)
	return nil
}

func (p *traitModifiersProvider) ContextMenuItems() []ContextMenuItem {
	var list []ContextMenuItem
	list = append(list,
		ContextMenuItem{i18n.Text("New Trait Modifier"), NewTraitModifierItemID},
		ContextMenuItem{i18n.Text("New Trait Modifier Container"), NewTraitContainerModifierItemID},
	)
	return AppendDefaultContextMenuItems(list)
}
