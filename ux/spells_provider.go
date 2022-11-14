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
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/maps"
)

var (
	spellListColMap = map[int]int{
		0:  model.SpellDescriptionColumn,
		1:  model.SpellCollegeColumn,
		2:  model.SpellResistColumn,
		3:  model.SpellClassColumn,
		4:  model.SpellCastCostColumn,
		5:  model.SpellMaintainCostColumn,
		6:  model.SpellCastTimeColumn,
		7:  model.SpellDurationColumn,
		8:  model.SpellDifficultyColumn,
		9:  model.SpellTagsColumn,
		10: model.SpellReferenceColumn,
	}
	entitySpellPageColMap = map[int]int{
		0: model.SpellDescriptionForPageColumn,
		1: model.SpellLevelColumn,
		2: model.SpellRelativeLevelColumn,
		3: model.SpellPointsColumn,
		4: model.SpellReferenceColumn,
	}
	spellPageColMap = map[int]int{
		0: model.SpellDescriptionForPageColumn,
		1: model.SpellDifficultyColumn,
		2: model.SpellPointsColumn,
		3: model.SpellReferenceColumn,
	}
	_ TableProvider[*model.Spell] = &spellsProvider{}
)

type spellsProvider struct {
	table    *unison.Table[*Node[*model.Spell]]
	colMap   map[int]int
	provider model.SpellListProvider
	forPage  bool
}

// NewSpellsProvider creates a new table provider for spells.
func NewSpellsProvider(provider model.SpellListProvider, forPage bool) TableProvider[*model.Spell] {
	p := &spellsProvider{
		provider: provider,
		forPage:  forPage,
	}
	if forPage {
		if _, ok := provider.(*model.Entity); ok {
			p.colMap = entitySpellPageColMap
		} else {
			p.colMap = spellPageColMap
		}
	} else {
		p.colMap = spellListColMap
	}
	return p
}

func (p *spellsProvider) RefKey() string {
	return model.BlockLayoutSpellsKey
}

func (p *spellsProvider) AllTags() []string {
	set := make(map[string]struct{})
	model.Traverse(func(modifier *model.Spell) bool {
		for _, tag := range modifier.Tags {
			set[tag] = struct{}{}
		}
		return false
	}, false, false, p.RootData()...)
	tags := maps.Keys(set)
	txt.SortStringsNaturalAscending(tags)
	return tags
}

func (p *spellsProvider) SetTable(table *unison.Table[*Node[*model.Spell]]) {
	p.table = table
}

func (p *spellsProvider) RootRowCount() int {
	return len(p.provider.SpellList())
}

func (p *spellsProvider) RootRows() []*Node[*model.Spell] {
	data := p.provider.SpellList()
	rows := make([]*Node[*model.Spell], 0, len(data))
	for _, one := range data {
		rows = append(rows, NewNode[*model.Spell](p.table, nil, p.colMap, one, p.forPage))
	}
	return rows
}

func (p *spellsProvider) SetRootRows(rows []*Node[*model.Spell]) {
	p.provider.SetSpellList(ExtractNodeDataFromList(rows))
}

func (p *spellsProvider) RootData() []*model.Spell {
	return p.provider.SpellList()
}

func (p *spellsProvider) SetRootData(data []*model.Spell) {
	p.provider.SetSpellList(data)
}

func (p *spellsProvider) Entity() *model.Entity {
	return p.provider.Entity()
}

func (p *spellsProvider) DragKey() string {
	return model.SpellID
}

func (p *spellsProvider) DragSVG() *unison.SVG {
	return svg.GCSSpells
}

func (p *spellsProvider) DropShouldMoveData(from, to *unison.Table[*Node[*model.Spell]]) bool {
	return from == to
}

func (p *spellsProvider) ProcessDropData(_, to *unison.Table[*Node[*model.Spell]]) {
	entityProvider := unison.Ancestor[model.EntityProvider](to)
	if !toolbox.IsNil(entityProvider) {
		entity := entityProvider.Entity()
		if entity != nil {
			for _, row := range to.SelectedRows(true) {
				model.Traverse(func(spell *model.Spell) bool {
					if spell.TechLevel != nil && *spell.TechLevel == "" {
						tl := entity.Profile.TechLevel
						spell.TechLevel = &tl
					}
					return false
				}, false, true, row.Data())
			}
		}
	}
}

func (p *spellsProvider) AltDropSupport() *AltDropSupport {
	return nil
}

func (p *spellsProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Spell"), i18n.Text("Spells")
}

func (p *spellsProvider) Headers() []unison.TableColumnHeader[*Node[*model.Spell]] {
	var headers []unison.TableColumnHeader[*Node[*model.Spell]]
	for i := 0; i < len(p.colMap); i++ {
		switch p.colMap[i] {
		case model.SpellDescriptionColumn, model.SpellDescriptionForPageColumn:
			headers = append(headers, NewEditorListHeader[*model.Spell](i18n.Text("Spell"), "", p.forPage))
		case model.SpellResistColumn:
			headers = append(headers, NewEditorListHeader[*model.Spell](i18n.Text("Resist"), i18n.Text("Resistance"), p.forPage))
		case model.SpellClassColumn:
			headers = append(headers, NewEditorListHeader[*model.Spell](i18n.Text("Class"), "", p.forPage))
		case model.SpellCollegeColumn:
			headers = append(headers, NewEditorListHeader[*model.Spell](i18n.Text("College"), "", p.forPage))
		case model.SpellCastCostColumn:
			headers = append(headers, NewEditorListHeader[*model.Spell](i18n.Text("Cost"), i18n.Text("The mana cost to cast the spell"),
				p.forPage))
		case model.SpellMaintainCostColumn:
			headers = append(headers, NewEditorListHeader[*model.Spell](i18n.Text("Maintain"), i18n.Text("The mana cost to maintain the spell"),
				p.forPage))
		case model.SpellCastTimeColumn:
			headers = append(headers, NewEditorListHeader[*model.Spell](i18n.Text("Time"), i18n.Text("The time required to cast the spell"),
				p.forPage))
		case model.SpellDurationColumn:
			headers = append(headers, NewEditorListHeader[*model.Spell](i18n.Text("Duration"), "", p.forPage))
		case model.SpellDifficultyColumn:
			headers = append(headers, NewEditorListHeader[*model.Spell](i18n.Text("Diff"), i18n.Text("Difficulty"), p.forPage))
		case model.SpellTagsColumn:
			headers = append(headers, NewEditorListHeader[*model.Spell](i18n.Text("Tags"), "", p.forPage))
		case model.SpellReferenceColumn:
			headers = append(headers, NewEditorPageRefHeader[*model.Spell](p.forPage))
		case model.SpellLevelColumn:
			headers = append(headers, NewEditorListHeader[*model.Spell](i18n.Text("SL"), i18n.Text("Skill Level"), p.forPage))
		case model.SpellRelativeLevelColumn:
			headers = append(headers, NewEditorListHeader[*model.Spell](i18n.Text("RSL"), i18n.Text("Relative Skill Level"), p.forPage))
		case model.SpellPointsColumn:
			headers = append(headers, NewEditorListHeader[*model.Spell](i18n.Text("Pts"), i18n.Text("Points"), p.forPage))
		default:
			jot.Fatalf(1, "invalid spell column: %d", p.colMap[i])
		}
	}
	return headers
}

func (p *spellsProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*model.Spell]]) {
}

func (p *spellsProvider) HierarchyColumnIndex() int {
	for k, v := range p.colMap {
		if v == model.SpellDescriptionColumn || v == model.SpellDescriptionForPageColumn {
			return k
		}
	}
	return 0
}

func (p *spellsProvider) ExcessWidthColumnIndex() int {
	return p.HierarchyColumnIndex()
}

func (p *spellsProvider) OpenEditor(owner Rebuildable, table *unison.Table[*Node[*model.Spell]]) {
	OpenEditor[*model.Spell](table, func(item *model.Spell) { EditSpell(owner, item) })
}

func (p *spellsProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*model.Spell]], variant ItemVariant) {
	var item *model.Spell
	switch variant {
	case NoItemVariant:
		item = model.NewSpell(p.Entity(), nil, false)
	case ContainerItemVariant:
		item = model.NewSpell(p.Entity(), nil, true)
	case AlternateItemVariant:
		item = model.NewRitualMagicSpell(p.Entity(), nil, false)
	default:
		jot.Fatal(1, "unhandled variant")
	}
	InsertItems[*model.Spell](owner, table, p.provider.SpellList, p.provider.SetSpellList,
		func(_ *unison.Table[*Node[*model.Spell]]) []*Node[*model.Spell] { return p.RootRows() }, item)
	EditSpell(owner, item)
}

func (p *spellsProvider) Serialize() ([]byte, error) {
	return jio.SerializeAndCompress(p.provider.SpellList())
}

func (p *spellsProvider) Deserialize(data []byte) error {
	var rows []*model.Spell
	if err := jio.DecompressAndDeserialize(data, &rows); err != nil {
		return err
	}
	p.provider.SetSpellList(rows)
	return nil
}

func (p *spellsProvider) ContextMenuItems() []ContextMenuItem {
	var list []ContextMenuItem
	list = append(list, SpellExtraContextMenuItems...)
	return append(list, DefaultContextMenuItems...)
}
