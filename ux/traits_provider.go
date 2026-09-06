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
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
)

var _ TableProvider[*gurps.Trait] = &traitsProvider{}

type traitsProvider struct {
	listProvider[*gurps.Trait]
	provider gurps.TraitListProvider
}

// NewTraitsProvider creates a new table provider for traits.
func NewTraitsProvider(provider gurps.TraitListProvider, forPage bool) TableProvider[*gurps.Trait] {
	p := &traitsProvider{provider: provider}
	p.listProvider = listProvider[*gurps.Trait]{
		dataOwner:  provider,
		list:       provider.TraitList,
		setList:    provider.SetTraitList,
		columnIDs:  p.ColumnIDs,
		headerData: gurps.TraitsHeaderData,
		forPage:    forPage,
	}
	return p
}

func (p *traitsProvider) RefKey() string {
	return gurps.BlockTraitsKey
}

func (p *traitsProvider) DragKey() *uti.DataType {
	return traitDragKey
}

func (p *traitsProvider) DragSVG() *unison.SVG {
	return svg.GCSTraits
}

func (p *traitsProvider) AltDropSupport() *AltDropSupport {
	return &AltDropSupport{
		DragKey: traitModifierDragKey,
		Drop: func(rowIndexes []int, data any) {
			if tableDragData, ok := data.(*unison.TableDragData[*Node[*gurps.TraitModifier]]); ok {
				// Every target is resolved up front, since the rebuild below replaces this table with a new one --
				// leaving this very table an orphan whose rows are no longer the ones on screen -- so the row indexes
				// only mean something before it runs. The sync in between is harmless: attaching modifiers adds and
				// removes no rows and changes no disclosure, so it rebuilds the row cache with the same rows in the
				// same order.
				targets := make([]*gurps.Trait, 0, len(rowIndexes))
				for _, rowIndex := range rowIndexes {
					if row := p.table.RowFromIndex(rowIndex); row != nil {
						targets = append(targets, row.Data())
					}
				}
				if len(targets) == 0 {
					return
				}
				dataOwner := p.DataOwner()
				libraryFile := libraryFileFromTable(tableDragData.Table)
				// Each target has to be given its own clones. They are separate modifiers from here on -- enabled,
				// renamed and edited independently -- so sharing one set among the targets would tie them together.
				// The clones are kept grouped by target for the nameables prompt below, which would otherwise show the
				// copies of one modifier as a run of identically titled sections with nothing to say which trait each
				// belongs to.
				groups := make([]NameableGroup[*gurps.TraitModifier], 0, len(targets))
				for _, target := range targets {
					clones := make([]*gurps.TraitModifier, 0, len(tableDragData.Rows))
					for _, row := range tableDragData.Rows {
						clones = append(clones, row.Data().Clone(libraryFile, dataOwner, nil, gurps.Reference))
					}
					target.Modifiers = append(target.Modifiers, clones...)
					groups = append(groups, NameableGroup[*gurps.TraitModifier]{Label: target.String(), Rows: clones})
				}
				p.table.SyncToModel()
				if !xreflect.IsNil(dataOwner) {
					if entity := dataOwner.OwningEntity(); entity != nil {
						// Rebuilding is also what reports the drop when the rows belong to an entity (see
						// dropRebuilder), so the owner is rebuilt as modified rather than just rebuilt.
						rebuildAsModified(dropRebuilder(p.table), true)
						// That rebuild can have replaced this very list: an enabled modifier carrying a
						// switchable feature gives the rows it was dropped onto switchable features, which brings
						// the switch column into view, and a list can only change its columns by building a new
						// table. p belongs to the list that was replaced and its table field is never updated, so
						// each prompt below has to be aimed at the table that took its place -- an orphan has no
						// Rebuildable above it, so the rebuild its answer asks for would silently be skipped. The
						// lookup is made twice because answering the modifier prompt rebuilds as well, which can
						// replace the list a second time.
						//
						// The modifier prompt has to be given the rows the modifiers were dropped onto, since
						// modifiers themselves aren't something ProcessModifiers can process, and only the
						// topmost of them, since it walks each row's descendants as well.
						ProcessModifiers(liveTable(p.table), minimalNodes(targets))
						ProcessNameableGroups(liveTable(p.table), groups)
					}
				}
			}
		},
	}
}

func (p *traitsProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Trait"), i18n.Text("Traits")
}

func (p *traitsProvider) ColumnIDs() []int {
	columnIDs := make([]int, 0, 5)
	if showSwitchColumn(p.forPage, p.provider, p.RootData()) {
		columnIDs = append(columnIDs, gurps.TraitSwitchColumn)
	}
	columnIDs = append(
		columnIDs,
		gurps.TraitDescriptionColumn,
		gurps.TraitPointsColumn,
	)
	if !p.forPage {
		columnIDs = append(columnIDs, gurps.TraitTagsColumn)
	}
	var sheetSettings *gurps.SheetSettings
	if p.forPage {
		if entity := p.DataOwner().OwningEntity(); entity != nil {
			sheetSettings = entity.SheetSettings
		} else {
			sheetSettings = gurps.GlobalSettings().SheetSettings()
		}
	}
	if p.forPage {
		if sheetSettings == nil || !sheetSettings.HidePageRefColumn {
			columnIDs = append(columnIDs, gurps.TraitReferenceColumn)
		}
		if sheetSettings == nil || !sheetSettings.HideSourceMismatch {
			columnIDs = append(columnIDs, gurps.TraitLibSrcColumn)
		}
	} else {
		columnIDs = append(columnIDs, gurps.TraitReferenceColumn)
	}
	return columnIDs
}

func (p *traitsProvider) HierarchyColumnID() int {
	return gurps.TraitDescriptionColumn
}

func (p *traitsProvider) ExcessWidthColumnID() int {
	return gurps.TraitDescriptionColumn
}

func (p *traitsProvider) OpenEditor(owner Rebuildable, table *unison.Table[*Node[*gurps.Trait]]) {
	OpenEditor(table, func(item *gurps.Trait) { EditTrait(owner, item) })
}

func (p *traitsProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*gurps.Trait]], variant ItemVariant) {
	item := gurps.NewTrait(p.DataOwner(), nil, variant == ContainerItemVariant)
	p.insertItems(owner, table, item)
	EditTrait(owner, item)
}

func (p *traitsProvider) ContextMenuItems() []ContextMenuItem {
	return AppendDefaultContextMenuItems([]ContextMenuItem{
		{
			Title: i18n.Text("New Trait"),
			ID:    NewTraitItemID,
		},
		{
			Title: i18n.Text("New Trait Container"),
			ID:    NewTraitContainerItemID,
		},
		{
			Title: i18n.Text("Add Natural Attacks"),
			ID:    AddNaturalAttacksItemID,
		},
		{
			Title: organizeTraitsAction.Title,
			ID:    OrganizeTraitsItemID,
		},
	})
}
