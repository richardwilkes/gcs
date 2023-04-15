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
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/maps"
)

var _ TableProvider[*gurps.Spell] = &spellsProvider{}

type spellsProvider struct {
	table    *unison.Table[*Node[*gurps.Spell]]
	provider gurps.SpellListProvider
	forPage  bool
}

// NewSpellsProvider creates a new table provider for spells.
func NewSpellsProvider(provider gurps.SpellListProvider, forPage bool) TableProvider[*gurps.Spell] {
	return &spellsProvider{
		provider: provider,
		forPage:  forPage,
	}
}

func (p *spellsProvider) RefKey() string {
	return gurps.BlockLayoutSpellsKey
}

func (p *spellsProvider) AllTags() []string {
	set := make(map[string]struct{})
	gurps.Traverse(func(modifier *gurps.Spell) bool {
		for _, tag := range modifier.Tags {
			set[tag] = struct{}{}
		}
		return false
	}, false, false, p.RootData()...)
	tags := maps.Keys(set)
	txt.SortStringsNaturalAscending(tags)
	return tags
}

func (p *spellsProvider) SetTable(table *unison.Table[*Node[*gurps.Spell]]) {
	p.table = table
}

func (p *spellsProvider) RootRowCount() int {
	return len(p.provider.SpellList())
}

func (p *spellsProvider) RootRows() []*Node[*gurps.Spell] {
	data := p.provider.SpellList()
	rows := make([]*Node[*gurps.Spell], 0, len(data))
	for _, one := range data {
		rows = append(rows, NewNode[*gurps.Spell](p.table, nil, one, p.forPage))
	}
	return rows
}

func (p *spellsProvider) SetRootRows(rows []*Node[*gurps.Spell]) {
	p.provider.SetSpellList(ExtractNodeDataFromList(rows))
}

func (p *spellsProvider) RootData() []*gurps.Spell {
	return p.provider.SpellList()
}

func (p *spellsProvider) SetRootData(data []*gurps.Spell) {
	p.provider.SetSpellList(data)
}

func (p *spellsProvider) Entity() *gurps.Entity {
	return p.provider.Entity()
}

func (p *spellsProvider) DragKey() string {
	return gurps.SpellID
}

func (p *spellsProvider) DragSVG() *unison.SVG {
	return svg.GCSSpells
}

func (p *spellsProvider) DropShouldMoveData(from, to *unison.Table[*Node[*gurps.Spell]]) bool {
	return from == to
}

func (p *spellsProvider) ProcessDropData(_, to *unison.Table[*Node[*gurps.Spell]]) {
	entityProvider := unison.Ancestor[gurps.EntityProvider](to)
	if !toolbox.IsNil(entityProvider) {
		entity := entityProvider.Entity()
		if entity != nil {
			for _, row := range to.SelectedRows(true) {
				gurps.Traverse(func(spell *gurps.Spell) bool {
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

func (p *spellsProvider) Headers() []unison.TableColumnHeader[*Node[*gurps.Spell]] {
	ids := p.ColumnIDs()
	headers := make([]unison.TableColumnHeader[*Node[*gurps.Spell]], 0, len(ids))
	for _, id := range ids {
		switch id {
		case gurps.SpellDescriptionColumn, gurps.SpellDescriptionForPageColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Spell](i18n.Text("Spell"), "", p.forPage))
		case gurps.SpellResistColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Spell](i18n.Text("Resist"), i18n.Text("Resistance"), p.forPage))
		case gurps.SpellClassColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Spell](i18n.Text("Class"), "", p.forPage))
		case gurps.SpellCollegeColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Spell](i18n.Text("College"), "", p.forPage))
		case gurps.SpellCastCostColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Spell](i18n.Text("Cost"), i18n.Text("The mana cost to cast the spell"),
				p.forPage))
		case gurps.SpellMaintainCostColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Spell](i18n.Text("Maintain"), i18n.Text("The mana cost to maintain the spell"),
				p.forPage))
		case gurps.SpellCastTimeColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Spell](i18n.Text("Time"), i18n.Text("The time required to cast the spell"),
				p.forPage))
		case gurps.SpellDurationColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Spell](i18n.Text("Duration"), "", p.forPage))
		case gurps.SpellDifficultyColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Spell](i18n.Text("Diff"), i18n.Text("Difficulty"), p.forPage))
		case gurps.SpellTagsColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Spell](i18n.Text("Tags"), "", p.forPage))
		case gurps.SpellReferenceColumn:
			headers = append(headers, NewEditorPageRefHeader[*gurps.Spell](p.forPage))
		case gurps.SpellLevelColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Spell](i18n.Text("SL"), i18n.Text("Skill Level"), p.forPage))
		case gurps.SpellRelativeLevelColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Spell](i18n.Text("RSL"), i18n.Text("Relative Skill Level"), p.forPage))
		case gurps.SpellPointsColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Spell](i18n.Text("Pts"), i18n.Text("Points"), p.forPage))
		}
	}
	return headers
}

func (p *spellsProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*gurps.Spell]]) {
}

func (p *spellsProvider) ColumnIDs() []int {
	columnIDs := make([]int, 0, 11)
	if p.forPage {
		if _, ok := p.provider.(*gurps.Entity); ok {
			columnIDs = append(columnIDs,
				gurps.SpellDescriptionForPageColumn,
				gurps.SpellLevelColumn,
				gurps.SpellRelativeLevelColumn,
				gurps.SpellPointsColumn,
			)
		} else {
			columnIDs = append(columnIDs,
				gurps.SpellDescriptionForPageColumn,
				gurps.SpellDifficultyColumn,
				gurps.SpellPointsColumn,
			)
		}
	} else {
		columnIDs = append(columnIDs,
			gurps.SpellDescriptionColumn,
			gurps.SpellCollegeColumn,
			gurps.SpellResistColumn,
			gurps.SpellClassColumn,
			gurps.SpellCastCostColumn,
			gurps.SpellMaintainCostColumn,
			gurps.SpellCastTimeColumn,
			gurps.SpellDurationColumn,
			gurps.SpellDifficultyColumn,
			gurps.SpellTagsColumn,
		)
	}
	return append(columnIDs, gurps.SpellReferenceColumn)
}

func (p *spellsProvider) HierarchyColumnID() int {
	if p.forPage {
		return gurps.SpellDescriptionForPageColumn
	}
	return gurps.SpellDescriptionColumn
}

func (p *spellsProvider) ExcessWidthColumnID() int {
	return p.HierarchyColumnID()
}

func (p *spellsProvider) OpenEditor(owner Rebuildable, table *unison.Table[*Node[*gurps.Spell]]) {
	OpenEditor[*gurps.Spell](table, func(item *gurps.Spell) { EditSpell(owner, item) })
}

func (p *spellsProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*gurps.Spell]], variant ItemVariant) {
	var item *gurps.Spell
	switch variant {
	case NoItemVariant:
		item = gurps.NewSpell(p.Entity(), nil, false)
	case ContainerItemVariant:
		item = gurps.NewSpell(p.Entity(), nil, true)
	case AlternateItemVariant:
		item = gurps.NewRitualMagicSpell(p.Entity(), nil, false)
	default:
		jot.Fatal(1, "unhandled variant")
	}
	InsertItems[*gurps.Spell](owner, table, p.provider.SpellList, p.provider.SetSpellList,
		func(_ *unison.Table[*Node[*gurps.Spell]]) []*Node[*gurps.Spell] { return p.RootRows() }, item)
	EditSpell(owner, item)
}

func (p *spellsProvider) Serialize() ([]byte, error) {
	return jio.SerializeAndCompress(p.provider.SpellList())
}

func (p *spellsProvider) Deserialize(data []byte) error {
	var rows []*gurps.Spell
	if err := jio.DecompressAndDeserialize(data, &rows); err != nil {
		return err
	}
	p.provider.SetSpellList(rows)
	return nil
}

func (p *spellsProvider) ContextMenuItems() []ContextMenuItem {
	var list []ContextMenuItem
	list = append(list,
		ContextMenuItem{i18n.Text("New Spell"), NewSpellItemID},
		ContextMenuItem{i18n.Text("New Spell Container"), NewSpellContainerItemID},
		ContextMenuItem{i18n.Text("New Ritual Magic Spell"), NewRitualMagicSpellItemID},
	)
	return AppendDefaultContextMenuItems(list)
}
