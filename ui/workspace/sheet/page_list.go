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

package sheet

import (
	"github.com/google/uuid"
	"github.com/richardwilkes/gcs/v5/constants"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/weapon"
	"github.com/richardwilkes/gcs/v5/model/theme"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/gcs/v5/ui/widget/ntable"
	"github.com/richardwilkes/gcs/v5/ui/workspace/editors"
	"github.com/richardwilkes/unison"
)

var _ widget.Syncer = &PageList[*gurps.Trait]{}

// PageList holds a list for a sheet page.
type PageList[T gurps.NodeConstraint[T]] struct {
	unison.Panel
	tableHeader *unison.TableHeader[*ntable.Node[T]]
	table       *unison.Table[*ntable.Node[T]]
	provider    ntable.TableProvider[T]
}

// NewTraitsPageList creates the traits page list.
func NewTraitsPageList(owner widget.Rebuildable, provider gurps.ListProvider) *PageList[*gurps.Trait] {
	p := newPageList(owner, editors.NewTraitsProvider(provider, true))
	p.installToggleDisabledHandler(owner)
	p.installIncrementLevelHandler(owner)
	p.installDecrementLevelHandler(owner)
	return p
}

// NewCarriedEquipmentPageList creates the carried equipment page list.
func NewCarriedEquipmentPageList(owner widget.Rebuildable, provider gurps.ListProvider) *PageList[*gurps.Equipment] {
	p := newPageList(owner, editors.NewEquipmentProvider(provider, true, true))
	p.installToggleEquippedHandler(owner)
	p.installIncrementQuantityHandler(owner)
	p.installDecrementQuantityHandler(owner)
	p.installIncrementUsesHandler(owner)
	p.installDecrementUsesHandler(owner)
	p.installIncrementTechLevelHandler(owner)
	p.installDecrementTechLevelHandler(owner)
	p.installConvertToContainerHandler(owner)
	return p
}

// NewOtherEquipmentPageList creates the other equipment page list.
func NewOtherEquipmentPageList(owner widget.Rebuildable, provider gurps.ListProvider) *PageList[*gurps.Equipment] {
	p := newPageList(owner, editors.NewEquipmentProvider(provider, true, false))
	p.installIncrementQuantityHandler(owner)
	p.installDecrementQuantityHandler(owner)
	p.installIncrementUsesHandler(owner)
	p.installDecrementUsesHandler(owner)
	p.installIncrementTechLevelHandler(owner)
	p.installDecrementTechLevelHandler(owner)
	p.installConvertToContainerHandler(owner)
	return p
}

// NewSkillsPageList creates the skills page list.
func NewSkillsPageList(owner widget.Rebuildable, provider gurps.ListProvider) *PageList[*gurps.Skill] {
	p := newPageList(owner, editors.NewSkillsProvider(provider, true))
	p.installIncrementPointsHandler(owner)
	p.installDecrementPointsHandler(owner)
	p.installIncrementSkillHandler(owner)
	p.installDecrementSkillHandler(owner)
	p.installIncrementTechLevelHandler(owner)
	p.installDecrementTechLevelHandler(owner)
	return p
}

// NewSpellsPageList creates the spells page list.
func NewSpellsPageList(owner widget.Rebuildable, provider gurps.SpellListProvider) *PageList[*gurps.Spell] {
	p := newPageList(owner, editors.NewSpellsProvider(provider, true))
	p.installIncrementPointsHandler(owner)
	p.installDecrementPointsHandler(owner)
	p.installIncrementSkillHandler(owner)
	p.installDecrementSkillHandler(owner)
	return p
}

// NewNotesPageList creates the notes page list.
func NewNotesPageList(owner widget.Rebuildable, provider gurps.ListProvider) *PageList[*gurps.Note] {
	return newPageList(owner, editors.NewNotesProvider(provider, true))
}

// NewConditionalModifiersPageList creates the conditional modifiers page list.
func NewConditionalModifiersPageList(entity *gurps.Entity) *PageList[*gurps.ConditionalModifier] {
	return newPageList(nil, editors.NewConditionalModifiersProvider(entity))
}

// NewReactionsPageList creates the reaction modifiers page list.
func NewReactionsPageList(entity *gurps.Entity) *PageList[*gurps.ConditionalModifier] {
	return newPageList(nil, editors.NewReactionModifiersProvider(entity))
}

// NewMeleeWeaponsPageList creates the melee weapons page list.
func NewMeleeWeaponsPageList(entity *gurps.Entity) *PageList[*gurps.Weapon] {
	return newPageList(nil, editors.NewWeaponsProvider(entity, weapon.Melee, true))
}

// NewRangedWeaponsPageList creates the ranged weapons page list.
func NewRangedWeaponsPageList(entity *gurps.Entity) *PageList[*gurps.Weapon] {
	return newPageList(nil, editors.NewWeaponsProvider(entity, weapon.Ranged, true))
}

func newPageList[T gurps.NodeConstraint[T]](owner widget.Rebuildable, provider ntable.TableProvider[T]) *PageList[T] {
	header, table := ntable.NewNodeTable[T](provider, theme.PageFieldPrimaryFont)
	p := &PageList[T]{
		tableHeader: header,
		table:       table,
		provider:    provider,
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{Columns: 1})
	p.SetBorder(unison.NewLineBorder(theme.HeaderColor, 0, unison.NewUniformInsets(1), false))

	p.table.PreventUserColumnResize = true
	p.tableHeader.DrawCallback = func(gc *unison.Canvas, dirty unison.Rect) {
		sortedOn := -1
		for i, hdr := range p.tableHeader.ColumnHeaders {
			if hdr.SortState().Order == 0 {
				sortedOn = i
				break
			}
		}
		if sortedOn != -1 {
			gc.DrawRect(dirty, p.tableHeader.BackgroundInk.Paint(gc, dirty, unison.Fill))
			r := p.tableHeader.ColumnFrame(sortedOn)
			r.X -= p.table.Padding.Left
			r.Width += p.table.Padding.Left + p.table.Padding.Right
			gc.DrawRect(r, theme.MarkerColor.Paint(gc, r, unison.Fill))
			save := p.tableHeader.BackgroundInk
			p.tableHeader.BackgroundInk = unison.Transparent
			p.tableHeader.DefaultDraw(gc, dirty)
			p.tableHeader.BackgroundInk = save
		} else {
			p.tableHeader.DefaultDraw(gc, dirty)
		}
	}
	p.table.SyncToModel()
	p.AddChild(p.tableHeader)
	p.AddChild(p.table)
	if owner != nil {
		ntable.InstallTableDropSupport(p.table, p.provider)
		p.InstallCmdHandlers(constants.OpenEditorItemID,
			func(_ any) bool { return p.table.HasSelection() },
			func(_ any) { p.provider.OpenEditor(owner, p.table) })
		p.InstallCmdHandlers(unison.DeleteItemID,
			func(_ any) bool { return p.table.HasSelection() },
			func(_ any) { ntable.DeleteSelection(p.table) })
		p.InstallCmdHandlers(constants.DuplicateItemID,
			func(_ any) bool { return p.table.HasSelection() },
			func(_ any) { ntable.DuplicateSelection(p.table) })
	}
	p.installOpenPageReferenceHandlers()
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
		VGrab:  true,
	})
	return p
}

func (p *PageList[T]) installOpenPageReferenceHandlers() {
	p.InstallCmdHandlers(constants.OpenOnePageReferenceItemID,
		func(_ any) bool { return editors.CanOpenPageRef(p.table) },
		func(_ any) { editors.OpenPageRef(p.table) })
	p.InstallCmdHandlers(constants.OpenEachPageReferenceItemID,
		func(_ any) bool { return editors.CanOpenPageRef(p.table) },
		func(_ any) { editors.OpenEachPageRef(p.table) })
}

func (p *PageList[T]) installToggleDisabledHandler(owner widget.Rebuildable) {
	if t, ok := (interface{}(p.table)).(*unison.Table[*ntable.Node[*gurps.Trait]]); ok {
		p.InstallCmdHandlers(constants.ToggleStateItemID,
			func(_ any) bool { return canToggleDisabled(t) },
			func(_ any) { toggleDisabled(owner, t) })
	}
}

func (p *PageList[T]) installToggleEquippedHandler(owner widget.Rebuildable) {
	if t, ok := (interface{}(p.table)).(*unison.Table[*ntable.Node[*gurps.Equipment]]); ok {
		p.InstallCmdHandlers(constants.ToggleStateItemID,
			func(_ any) bool { return canToggleEquipped(t) },
			func(_ any) { toggleEquipped(owner, t) })
	}
}

func (p *PageList[T]) installIncrementPointsHandler(owner widget.Rebuildable) {
	p.InstallCmdHandlers(constants.IncrementItemID,
		func(_ any) bool { return canAdjustRawPoints(p.table, true) },
		func(_ any) { adjustRawPoints(owner, p.table, true) })
}

func (p *PageList[T]) installDecrementPointsHandler(owner widget.Rebuildable) {
	p.InstallCmdHandlers(constants.DecrementItemID,
		func(_ any) bool { return canAdjustRawPoints(p.table, false) },
		func(_ any) { adjustRawPoints(owner, p.table, false) })
}

func (p *PageList[T]) installIncrementLevelHandler(owner widget.Rebuildable) {
	if t, ok := (interface{}(p.table)).(*unison.Table[*ntable.Node[*gurps.Trait]]); ok {
		p.InstallCmdHandlers(constants.IncrementItemID,
			func(_ any) bool { return canAdjustTraitLevel(t, true) },
			func(_ any) { adjustTraitLevel(owner, t, true) })
	}
}

func (p *PageList[T]) installDecrementLevelHandler(owner widget.Rebuildable) {
	if t, ok := (interface{}(p.table)).(*unison.Table[*ntable.Node[*gurps.Trait]]); ok {
		p.InstallCmdHandlers(constants.DecrementItemID,
			func(_ any) bool { return canAdjustTraitLevel(t, false) },
			func(_ any) { adjustTraitLevel(owner, t, false) })
	}
}

func (p *PageList[T]) installIncrementQuantityHandler(owner widget.Rebuildable) {
	if t, ok := (interface{}(p.table)).(*unison.Table[*ntable.Node[*gurps.Equipment]]); ok {
		p.InstallCmdHandlers(constants.IncrementItemID,
			func(_ any) bool { return canAdjustQuantity(t, true) },
			func(_ any) { adjustQuantity(owner, t, true) })
	}
}

func (p *PageList[T]) installDecrementQuantityHandler(owner widget.Rebuildable) {
	if t, ok := (interface{}(p.table)).(*unison.Table[*ntable.Node[*gurps.Equipment]]); ok {
		p.InstallCmdHandlers(constants.DecrementItemID,
			func(_ any) bool { return canAdjustQuantity(t, false) },
			func(_ any) { adjustQuantity(owner, t, false) })
	}
}

func (p *PageList[T]) installIncrementUsesHandler(owner widget.Rebuildable) {
	if t, ok := (interface{}(p.table)).(*unison.Table[*ntable.Node[*gurps.Equipment]]); ok {
		p.InstallCmdHandlers(constants.IncrementUsesItemID,
			func(_ any) bool { return canAdjustUses(t, 1) },
			func(_ any) { adjustUses(owner, t, 1) })
	}
}

func (p *PageList[T]) installDecrementUsesHandler(owner widget.Rebuildable) {
	if t, ok := (interface{}(p.table)).(*unison.Table[*ntable.Node[*gurps.Equipment]]); ok {
		p.InstallCmdHandlers(constants.DecrementUsesItemID,
			func(_ any) bool { return canAdjustUses(t, -1) },
			func(_ any) { adjustUses(owner, t, -1) })
	}
}

func (p *PageList[T]) installIncrementSkillHandler(owner widget.Rebuildable) {
	p.InstallCmdHandlers(constants.IncrementSkillLevelItemID,
		func(_ any) bool { return canAdjustSkillLevel(p.table, true) },
		func(_ any) { adjustSkillLevel(owner, p.table, true) })
}

func (p *PageList[T]) installDecrementSkillHandler(owner widget.Rebuildable) {
	p.InstallCmdHandlers(constants.DecrementSkillLevelItemID,
		func(_ any) bool { return canAdjustSkillLevel(p.table, false) },
		func(_ any) { adjustSkillLevel(owner, p.table, false) })
}

func (p *PageList[T]) installIncrementTechLevelHandler(owner widget.Rebuildable) {
	p.InstallCmdHandlers(constants.IncrementTechLevelItemID,
		func(_ any) bool { return canAdjustTechLevel(p.table, fxp.One) },
		func(_ any) { adjustTechLevel(owner, p.table, fxp.One) })
}

func (p *PageList[T]) installDecrementTechLevelHandler(owner widget.Rebuildable) {
	p.InstallCmdHandlers(constants.DecrementTechLevelItemID,
		func(_ any) bool { return canAdjustTechLevel(p.table, -fxp.One) },
		func(_ any) { adjustTechLevel(owner, p.table, -fxp.One) })
}

func (p *PageList[T]) installConvertToContainerHandler(owner widget.Rebuildable) {
	if t, ok := (interface{}(p.table)).(*unison.Table[*ntable.Node[*gurps.Equipment]]); ok {
		p.InstallCmdHandlers(constants.ConvertToContainerItemID,
			func(_ any) bool { return canConvertToContainer(t) },
			func(_ any) { convertToContainer(owner, t) })
	}
}

// SelectedNodes returns the set of selected nodes. If 'minimal' is true, then children of selected rows that may also
// be selected are not returned, just the topmost row that is selected in any given hierarchy.
func (p *PageList[T]) SelectedNodes(minimal bool) []*ntable.Node[T] {
	if p == nil {
		return nil
	}
	return p.table.SelectedRows(minimal)
}

// RecordSelection collects the currently selected row UUIDs.
func (p *PageList[T]) RecordSelection() map[uuid.UUID]bool {
	if p == nil {
		return nil
	}
	return p.table.CopySelectionMap()
}

// ApplySelection locates the rows with the given UUIDs and selects them, replacing any existing selection.
func (p *PageList[T]) ApplySelection(selection map[uuid.UUID]bool) {
	if p != nil {
		p.table.SetSelectionMap(selection)
	}
}

// Sync the underlying data.
func (p *PageList[T]) Sync() {
	p.provider.SyncHeader(p.tableHeader.ColumnHeaders)
	selection := p.RecordSelection()
	p.table.SyncToModel()
	p.ApplySelection(selection)
	p.table.NeedsLayout = true
	p.NeedsLayout = true
	if parent := p.Parent(); parent != nil {
		parent.NeedsLayout = true
	}
}

// CreateItem calls CreateItem on the contained TableProvider.
func (p *PageList[T]) CreateItem(owner widget.Rebuildable, variant ntable.ItemVariant) {
	p.provider.CreateItem(owner, p.table, variant)
}
