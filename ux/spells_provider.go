// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
)

var _ TableProvider[*gurps.Spell] = &spellsProvider{}

type spellsProvider struct {
	listProvider[*gurps.Spell]
	provider gurps.SpellListProvider
}

// NewSpellsProvider creates a new table provider for spells.
func NewSpellsProvider(provider gurps.SpellListProvider, forPage bool) TableProvider[*gurps.Spell] {
	p := &spellsProvider{provider: provider}
	p.listProvider = listProvider[*gurps.Spell]{
		dataOwner:    provider,
		list:         provider.SpellList,
		setList:      provider.SetSpellList,
		columnIDs:    p.ColumnIDs,
		headerData:   gurps.SpellsHeaderData,
		newItem:      gurps.NewSpell,
		edit:         func(owner Rebuildable, item *gurps.Spell) { EditSpell(owner, item) },
		forPage:      forPage,
		filterKey:    gurps.ListFilterKeyForExtension(gurps.SpellsExt),
		filterFields: gurps.SpellFilterFields,
	}
	return p
}

func (p *spellsProvider) RefKey() string {
	return gurps.BlockSpellsKey
}

func (p *spellsProvider) DragKey() *uti.DataType {
	return spellDragKey
}

func (p *spellsProvider) DragSVG() *unison.SVG {
	return svg.GCSSpells
}

func (p *spellsProvider) ProcessDropData(_, to *unison.Table[*Node[*gurps.Spell]]) {
	if dataOwnerProvider := to.Ancestor[gurps.DataOwnerProvider](); !xreflect.IsNil(dataOwnerProvider) {
		if dataOwner := dataOwnerProvider.DataOwner(); !xreflect.IsNil(dataOwner) {
			if entity := dataOwner.OwningEntity(); entity != nil {
				for _, row := range to.SelectedRows(true) {
					gurps.Traverse(func(spell *gurps.Spell) bool {
						resolveEmptyTechLevel(spell, entity.Profile.TechLevel)
						return false
					}, false, true, row.Data())
				}
			}
		}
	}
}

func (p *spellsProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Spell"), i18n.Text("Spells")
}

func (p *spellsProvider) ColumnIDs() []int {
	columnIDs := make([]int, 0, 12)
	if showSwitchColumn(p.forPage, p.provider, p.RootData()) {
		columnIDs = append(columnIDs, gurps.SpellSwitchColumn)
	}
	if p.forPage {
		if _, ok := p.provider.(*gurps.Entity); ok {
			columnIDs = append(
				columnIDs,
				gurps.SpellDescriptionForPageColumn,
				gurps.SpellLevelColumn,
				gurps.SpellRelativeLevelColumn,
				gurps.SpellPointsColumn,
			)
		} else {
			columnIDs = append(
				columnIDs,
				gurps.SpellDescriptionForPageColumn,
				gurps.SpellDifficultyColumn,
				gurps.SpellPointsColumn,
			)
		}
	} else {
		columnIDs = append(
			columnIDs,
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
	return p.appendReferenceColumns(columnIDs, gurps.SpellReferenceColumn, gurps.SpellLibSrcColumn)
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

// CreateItem adds the alternate variant, a ritual magic spell, to the spell and spell container the shared
// implementation creates.
func (p *spellsProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*gurps.Spell]], variant ItemVariant) {
	switch variant {
	case NoItemVariant, ContainerItemVariant:
		p.listProvider.CreateItem(owner, table, variant)
	case AlternateItemVariant:
		p.createItem(owner, table, gurps.NewRitualMagicSpell(p.DataOwner(), nil, false))
	default:
		errs.Log(errs.New("unhandled variant"), "variant", int(variant))
	}
}

func (p *spellsProvider) ContextMenuItems() []ContextMenuItem {
	return AppendDefaultContextMenuItems([]ContextMenuItem{
		contextMenuItemFor(newSpellAction),
		contextMenuItemFor(newSpellContainerAction),
		contextMenuItemFor(newRitualMagicSpellAction),
	})
}
