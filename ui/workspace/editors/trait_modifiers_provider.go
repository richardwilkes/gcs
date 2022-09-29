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

package editors

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/gid"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/gcs/v5/ui/widget/ntable"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/unison"
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
	_ ntable.TableProvider[*gurps.TraitModifier] = &traitModifierProvider{}
)

type traitModifierProvider struct {
	table    *unison.Table[*ntable.Node[*gurps.TraitModifier]]
	colMap   map[int]int
	provider gurps.TraitModifierListProvider
}

// NewTraitModifiersProvider creates a new table provider for trait modifiers.
func NewTraitModifiersProvider(provider gurps.TraitModifierListProvider, forEditor bool) ntable.TableProvider[*gurps.TraitModifier] {
	p := &traitModifierProvider{
		provider: provider,
	}
	if forEditor {
		p.colMap = traitModifierInEditorColMap
	} else {
		p.colMap = traitModifierColMap
	}
	return p
}

func (p *traitModifierProvider) RefKey() string {
	return gid.TraitModifier
}

func (p *traitModifierProvider) SetTable(table *unison.Table[*ntable.Node[*gurps.TraitModifier]]) {
	p.table = table
}

func (p *traitModifierProvider) RootRowCount() int {
	return len(p.provider.TraitModifierList())
}

func (p *traitModifierProvider) RootRows() []*ntable.Node[*gurps.TraitModifier] {
	data := p.provider.TraitModifierList()
	rows := make([]*ntable.Node[*gurps.TraitModifier], 0, len(data))
	for _, one := range data {
		rows = append(rows, ntable.NewNode[*gurps.TraitModifier](p.table, nil, p.colMap, one, false))
	}
	return rows
}

func (p *traitModifierProvider) SetRootRows(rows []*ntable.Node[*gurps.TraitModifier]) {
	p.provider.SetTraitModifierList(ntable.ExtractNodeDataFromList(rows))
}

func (p *traitModifierProvider) RootData() []*gurps.TraitModifier {
	return p.provider.TraitModifierList()
}

func (p *traitModifierProvider) SetRootData(data []*gurps.TraitModifier) {
	p.provider.SetTraitModifierList(data)
}

func (p *traitModifierProvider) Entity() *gurps.Entity {
	return p.provider.Entity()
}

func (p *traitModifierProvider) DragKey() string {
	return gid.TraitModifier
}

func (p *traitModifierProvider) DragSVG() *unison.SVG {
	return res.GCSTraitModifiersSVG
}

func (p *traitModifierProvider) DropShouldMoveData(from, to *unison.Table[*ntable.Node[*gurps.TraitModifier]]) bool {
	return from == to
}

func (p *traitModifierProvider) ProcessDropData(_, _ *unison.Table[*ntable.Node[*gurps.TraitModifier]]) {
}

func (p *traitModifierProvider) AltDropSupport() *ntable.AltDropSupport {
	return nil
}

func (p *traitModifierProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Trait Modifier"), i18n.Text("Trait Modifiers")
}

func (p *traitModifierProvider) Headers() []unison.TableColumnHeader[*ntable.Node[*gurps.TraitModifier]] {
	var headers []unison.TableColumnHeader[*ntable.Node[*gurps.TraitModifier]]
	for i := 0; i < len(p.colMap); i++ {
		switch p.colMap[i] {
		case gurps.TraitModifierEnabledColumn:
			headers = append(headers, NewEnabledHeader[*gurps.TraitModifier](false))
		case gurps.TraitModifierDescriptionColumn:
			headers = append(headers, NewHeader[*gurps.TraitModifier](i18n.Text("Trait Modifier"), "", false))
		case gurps.TraitModifierCostColumn:
			headers = append(headers, NewHeader[*gurps.TraitModifier](i18n.Text("Cost Modifier"), "", false))
		case gurps.TraitModifierTagsColumn:
			headers = append(headers, NewHeader[*gurps.TraitModifier](i18n.Text("Tags"), "", false))
		case gurps.TraitModifierReferenceColumn:
			headers = append(headers, NewPageRefHeader[*gurps.TraitModifier](false))
		default:
			jot.Fatalf(1, "invalid trait modifier column: %d", p.colMap[i])
		}
	}
	return headers
}

func (p *traitModifierProvider) SyncHeader(_ []unison.TableColumnHeader[*ntable.Node[*gurps.TraitModifier]]) {
}

func (p *traitModifierProvider) HierarchyColumnIndex() int {
	for k, v := range p.colMap {
		if v == gurps.TraitModifierDescriptionColumn {
			return k
		}
	}
	return 0
}

func (p *traitModifierProvider) ExcessWidthColumnIndex() int {
	return p.HierarchyColumnIndex()
}

func (p *traitModifierProvider) OpenEditor(owner widget.Rebuildable, table *unison.Table[*ntable.Node[*gurps.TraitModifier]]) {
	ntable.OpenEditor[*gurps.TraitModifier](table, func(item *gurps.TraitModifier) {
		EditTraitModifier(owner, item)
	})
}

func (p *traitModifierProvider) CreateItem(owner widget.Rebuildable, table *unison.Table[*ntable.Node[*gurps.TraitModifier]], variant ntable.ItemVariant) {
	item := gurps.NewTraitModifier(p.Entity(), nil, variant == ntable.ContainerItemVariant)
	ntable.InsertItems[*gurps.TraitModifier](owner, table, p.provider.TraitModifierList, p.provider.SetTraitModifierList,
		func(_ *unison.Table[*ntable.Node[*gurps.TraitModifier]]) []*ntable.Node[*gurps.TraitModifier] {
			return p.RootRows()
		}, item)
	EditTraitModifier(owner, item)
}

func (p *traitModifierProvider) Serialize() ([]byte, error) {
	return jio.SerializeAndCompress(p.provider.TraitModifierList())
}

func (p *traitModifierProvider) Deserialize(data []byte) error {
	var rows []*gurps.TraitModifier
	if err := jio.DecompressAndDeserialize(data, &rows); err != nil {
		return err
	}
	p.provider.SetTraitModifierList(rows)
	return nil
}

func (p *traitModifierProvider) ContextMenuItems() []ntable.ContextMenuItem {
	var list []ntable.ContextMenuItem
	list = append(list, ntable.TraitModifierExtraContextMenuItems...)
	return append(list, ntable.DefaultContextMenuItems...)
}
