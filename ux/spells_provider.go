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
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/atexit"
	"github.com/richardwilkes/toolbox/collection/dict"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
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
	tags := dict.Keys(set)
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
		rows = append(rows, NewNode(p.table, nil, one, p.forPage))
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

func (p *spellsProvider) DataOwner() gurps.DataOwner {
	return p.provider.DataOwner()
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
	if dataOwnerProvider := unison.Ancestor[gurps.DataOwnerProvider](to); !toolbox.IsNil(dataOwnerProvider) {
		if dataOwner := dataOwnerProvider.DataOwner(); !toolbox.IsNil(dataOwner) {
			if entity := dataOwner.OwningEntity(); entity != nil {
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
		headers = append(headers, headerFromData[*gurps.Spell](gurps.SpellsHeaderData(id), p.forPage))
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
			gurps.SpellPrereqCountColumn,
			gurps.SpellTagsColumn,
		)
	}
	columnIDs = append(columnIDs, gurps.SpellReferenceColumn)
	if p.forPage {
		if entity := p.DataOwner().OwningEntity(); entity == nil || !entity.SheetSettings.HideSourceMismatch {
			columnIDs = append(columnIDs, gurps.SpellLibSrcColumn)
		}
	}
	return columnIDs
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
	OpenEditor(table, func(item *gurps.Spell) { EditSpell(owner, item) })
}

func (p *spellsProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*gurps.Spell]], variant ItemVariant) {
	var item *gurps.Spell
	switch variant {
	case NoItemVariant:
		item = gurps.NewSpell(p.DataOwner(), nil, false)
	case ContainerItemVariant:
		item = gurps.NewSpell(p.DataOwner(), nil, true)
	case AlternateItemVariant:
		item = gurps.NewRitualMagicSpell(p.DataOwner(), nil, false)
	default:
		errs.Log(errs.New("unhandled variant"), "variant", int(variant))
		atexit.Exit(1)
	}
	InsertItems(owner, table, p.provider.SpellList, p.provider.SetSpellList,
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
