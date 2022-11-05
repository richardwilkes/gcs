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
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/maps"
)

var (
	traitModifierColMap = map[int]int{
		0: gurps.TraitModifierDescriptionColumn,
		1: gurps.TraitModifierCostColumn,
		2: gurps.TraitModifierTagsColumn,
		3: gurps.TraitModifierReferenceColumn,
	}
	traitModifierInEditorColMap = map[int]int{
		0: gurps.TraitModifierEnabledColumn,
		1: gurps.TraitModifierDescriptionColumn,
		2: gurps.TraitModifierCostColumn,
		3: gurps.TraitModifierTagsColumn,
		4: gurps.TraitModifierReferenceColumn,
	}
	_ TableProvider[*gurps.TraitModifier] = &traitModifiersProvider{}
)

type traitModifiersProvider struct {
	table    *unison.Table[*Node[*gurps.TraitModifier]]
	colMap   map[int]int
	provider gurps.TraitModifierListProvider
}

// NewTraitModifiersProvider creates a new table provider for trait modifiers.
func NewTraitModifiersProvider(provider gurps.TraitModifierListProvider, forEditor bool) TableProvider[*gurps.TraitModifier] {
	p := &traitModifiersProvider{
		provider: provider,
	}
	if forEditor {
		p.colMap = traitModifierInEditorColMap
	} else {
		p.colMap = traitModifierColMap
	}
	return p
}

func (p *traitModifiersProvider) RefKey() string {
	return gid.TraitModifier
}

func (p *traitModifiersProvider) AllTags() []string {
	set := make(map[string]struct{})
	gurps.Traverse(func(modifier *gurps.TraitModifier) bool {
		for _, tag := range modifier.Tags {
			set[tag] = struct{}{}
		}
		return false
	}, false, false, p.RootData()...)
	tags := maps.Keys(set)
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
		rows = append(rows, NewNode[*gurps.TraitModifier](p.table, nil, p.colMap, one, false))
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

func (p *traitModifiersProvider) Entity() *gurps.Entity {
	return p.provider.Entity()
}

func (p *traitModifiersProvider) DragKey() string {
	return gid.TraitModifier
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
	var headers []unison.TableColumnHeader[*Node[*gurps.TraitModifier]]
	for i := 0; i < len(p.colMap); i++ {
		switch p.colMap[i] {
		case gurps.TraitModifierEnabledColumn:
			headers = append(headers, NewEnabledHeader[*gurps.TraitModifier](false))
		case gurps.TraitModifierDescriptionColumn:
			headers = append(headers, NewEditorListHeader[*gurps.TraitModifier](i18n.Text("Trait Modifier"), "", false))
		case gurps.TraitModifierCostColumn:
			headers = append(headers, NewEditorListHeader[*gurps.TraitModifier](i18n.Text("Cost Modifier"), "", false))
		case gurps.TraitModifierTagsColumn:
			headers = append(headers, NewEditorListHeader[*gurps.TraitModifier](i18n.Text("Tags"), "", false))
		case gurps.TraitModifierReferenceColumn:
			headers = append(headers, NewEditorPageRefHeader[*gurps.TraitModifier](false))
		default:
			jot.Fatalf(1, "invalid trait modifier column: %d", p.colMap[i])
		}
	}
	return headers
}

func (p *traitModifiersProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*gurps.TraitModifier]]) {
}

func (p *traitModifiersProvider) HierarchyColumnIndex() int {
	for k, v := range p.colMap {
		if v == gurps.TraitModifierDescriptionColumn {
			return k
		}
	}
	return 0
}

func (p *traitModifiersProvider) ExcessWidthColumnIndex() int {
	return p.HierarchyColumnIndex()
}

func (p *traitModifiersProvider) OpenEditor(owner Rebuildable, table *unison.Table[*Node[*gurps.TraitModifier]]) {
	OpenEditor[*gurps.TraitModifier](table, func(item *gurps.TraitModifier) {
		EditTraitModifier(owner, item)
	})
}

func (p *traitModifiersProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*gurps.TraitModifier]], variant ItemVariant) {
	item := gurps.NewTraitModifier(p.Entity(), nil, variant == ContainerItemVariant)
	InsertItems[*gurps.TraitModifier](owner, table, p.provider.TraitModifierList, p.provider.SetTraitModifierList,
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
	list = append(list, TraitModifierExtraContextMenuItems...)
	return append(list, DefaultContextMenuItems...)
}
