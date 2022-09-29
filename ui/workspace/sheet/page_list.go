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

var (
	_ widget.Syncer = &PageList[*gurps.Trait]{}
	_ pdfHelper     = &PageList[*gurps.Trait]{}
)

// PageList holds a list for a sheet page.
type PageList[T gurps.NodeTypes] struct {
	unison.Panel
	tableHeader *unison.TableHeader[*ntable.Node[T]]
	Table       *unison.Table[*ntable.Node[T]]
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

func newPageList[T gurps.NodeTypes](owner widget.Rebuildable, provider ntable.TableProvider[T]) *PageList[T] {
	header, table := ntable.NewNodeTable[T](provider, theme.PageFieldPrimaryFont)
	table.RefKey = provider.RefKey()
	p := &PageList[T]{
		tableHeader: header,
		Table:       table,
		provider:    provider,
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{Columns: 1})
	p.SetBorder(unison.NewLineBorder(theme.HeaderColor, 0, unison.NewUniformInsets(1), false))

	p.Table.PreventUserColumnResize = true
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
			r.X -= p.Table.Padding.Left
			r.Width += p.Table.Padding.Left + p.Table.Padding.Right
			gc.DrawRect(r, theme.MarkerColor.Paint(gc, r, unison.Fill))
			save := p.tableHeader.BackgroundInk
			p.tableHeader.BackgroundInk = unison.Transparent
			p.tableHeader.DefaultDraw(gc, dirty)
			p.tableHeader.BackgroundInk = save
		} else {
			p.tableHeader.DefaultDraw(gc, dirty)
		}
	}
	p.Table.SyncToModel()
	p.AddChild(p.tableHeader)
	p.AddChild(p.Table)
	if owner != nil {
		ntable.InstallTableDropSupport(p.Table, p.provider)
		p.InstallCmdHandlers(constants.OpenEditorItemID,
			func(_ any) bool { return p.Table.HasSelection() },
			func(_ any) { p.provider.OpenEditor(owner, p.Table) })
		p.InstallCmdHandlers(unison.DeleteItemID,
			func(_ any) bool { return p.Table.HasSelection() },
			func(_ any) { ntable.DeleteSelection(p.Table) })
		p.InstallCmdHandlers(constants.DuplicateItemID,
			func(_ any) bool { return p.Table.HasSelection() },
			func(_ any) { ntable.DuplicateSelection(p.Table) })
	}
	p.installOpenPageReferenceHandlers()
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
	})
	return p
}

func (p *PageList[T]) installOpenPageReferenceHandlers() {
	p.InstallCmdHandlers(constants.OpenOnePageReferenceItemID,
		func(_ any) bool { return editors.CanOpenPageRef(p.Table) },
		func(_ any) { editors.OpenPageRef(p.Table) })
	p.InstallCmdHandlers(constants.OpenEachPageReferenceItemID,
		func(_ any) bool { return editors.CanOpenPageRef(p.Table) },
		func(_ any) { editors.OpenEachPageRef(p.Table) })
}

func (p *PageList[T]) installToggleDisabledHandler(owner widget.Rebuildable) {
	if t, ok := (any(p.Table)).(*unison.Table[*ntable.Node[*gurps.Trait]]); ok {
		p.InstallCmdHandlers(constants.ToggleStateItemID,
			func(_ any) bool { return canToggleDisabled(t) },
			func(_ any) { toggleDisabled(owner, t) })
	}
}

func (p *PageList[T]) installToggleEquippedHandler(owner widget.Rebuildable) {
	if t, ok := (any(p.Table)).(*unison.Table[*ntable.Node[*gurps.Equipment]]); ok {
		p.InstallCmdHandlers(constants.ToggleStateItemID,
			func(_ any) bool { return canToggleEquipped(t) },
			func(_ any) { toggleEquipped(owner, t) })
	}
}

func (p *PageList[T]) installIncrementPointsHandler(owner widget.Rebuildable) {
	p.InstallCmdHandlers(constants.IncrementItemID,
		func(_ any) bool { return canAdjustRawPoints(p.Table, true) },
		func(_ any) { adjustRawPoints(owner, p.Table, true) })
}

func (p *PageList[T]) installDecrementPointsHandler(owner widget.Rebuildable) {
	p.InstallCmdHandlers(constants.DecrementItemID,
		func(_ any) bool { return canAdjustRawPoints(p.Table, false) },
		func(_ any) { adjustRawPoints(owner, p.Table, false) })
}

func (p *PageList[T]) installIncrementLevelHandler(owner widget.Rebuildable) {
	if t, ok := (any(p.Table)).(*unison.Table[*ntable.Node[*gurps.Trait]]); ok {
		p.InstallCmdHandlers(constants.IncrementItemID,
			func(_ any) bool { return canAdjustTraitLevel(t, true) },
			func(_ any) { adjustTraitLevel(owner, t, true) })
	}
}

func (p *PageList[T]) installDecrementLevelHandler(owner widget.Rebuildable) {
	if t, ok := (any(p.Table)).(*unison.Table[*ntable.Node[*gurps.Trait]]); ok {
		p.InstallCmdHandlers(constants.DecrementItemID,
			func(_ any) bool { return canAdjustTraitLevel(t, false) },
			func(_ any) { adjustTraitLevel(owner, t, false) })
	}
}

func (p *PageList[T]) installIncrementQuantityHandler(owner widget.Rebuildable) {
	if t, ok := (any(p.Table)).(*unison.Table[*ntable.Node[*gurps.Equipment]]); ok {
		p.InstallCmdHandlers(constants.IncrementItemID,
			func(_ any) bool { return canAdjustQuantity(t, true) },
			func(_ any) { adjustQuantity(owner, t, true) })
	}
}

func (p *PageList[T]) installDecrementQuantityHandler(owner widget.Rebuildable) {
	if t, ok := (any(p.Table)).(*unison.Table[*ntable.Node[*gurps.Equipment]]); ok {
		p.InstallCmdHandlers(constants.DecrementItemID,
			func(_ any) bool { return canAdjustQuantity(t, false) },
			func(_ any) { adjustQuantity(owner, t, false) })
	}
}

func (p *PageList[T]) installIncrementUsesHandler(owner widget.Rebuildable) {
	if t, ok := (any(p.Table)).(*unison.Table[*ntable.Node[*gurps.Equipment]]); ok {
		p.InstallCmdHandlers(constants.IncrementUsesItemID,
			func(_ any) bool { return canAdjustUses(t, 1) },
			func(_ any) { adjustUses(owner, t, 1) })
	}
}

func (p *PageList[T]) installDecrementUsesHandler(owner widget.Rebuildable) {
	if t, ok := (any(p.Table)).(*unison.Table[*ntable.Node[*gurps.Equipment]]); ok {
		p.InstallCmdHandlers(constants.DecrementUsesItemID,
			func(_ any) bool { return canAdjustUses(t, -1) },
			func(_ any) { adjustUses(owner, t, -1) })
	}
}

func (p *PageList[T]) installIncrementSkillHandler(owner widget.Rebuildable) {
	p.InstallCmdHandlers(constants.IncrementSkillLevelItemID,
		func(_ any) bool { return canAdjustSkillLevel(p.Table, true) },
		func(_ any) { adjustSkillLevel(owner, p.Table, true) })
}

func (p *PageList[T]) installDecrementSkillHandler(owner widget.Rebuildable) {
	p.InstallCmdHandlers(constants.DecrementSkillLevelItemID,
		func(_ any) bool { return canAdjustSkillLevel(p.Table, false) },
		func(_ any) { adjustSkillLevel(owner, p.Table, false) })
}

func (p *PageList[T]) installIncrementTechLevelHandler(owner widget.Rebuildable) {
	p.InstallCmdHandlers(constants.IncrementTechLevelItemID,
		func(_ any) bool { return canAdjustTechLevel(p.Table, fxp.One) },
		func(_ any) { adjustTechLevel(owner, p.Table, fxp.One) })
}

func (p *PageList[T]) installDecrementTechLevelHandler(owner widget.Rebuildable) {
	p.InstallCmdHandlers(constants.DecrementTechLevelItemID,
		func(_ any) bool { return canAdjustTechLevel(p.Table, -fxp.One) },
		func(_ any) { adjustTechLevel(owner, p.Table, -fxp.One) })
}

func (p *PageList[T]) installConvertToContainerHandler(owner widget.Rebuildable) {
	if t, ok := (any(p.Table)).(*unison.Table[*ntable.Node[*gurps.Equipment]]); ok {
		p.InstallCmdHandlers(constants.ConvertToContainerItemID,
			func(_ any) bool { return CanConvertToContainer(t) },
			func(_ any) { ConvertToContainer(owner, t) })
	}
}

// SelectedNodes returns the set of selected nodes. If 'minimal' is true, then children of selected rows that may also
// be selected are not returned, just the topmost row that is selected in any given hierarchy.
func (p *PageList[T]) SelectedNodes(minimal bool) []*ntable.Node[T] {
	if p == nil {
		return nil
	}
	return p.Table.SelectedRows(minimal)
}

// RecordSelection collects the currently selected row UUIDs.
func (p *PageList[T]) RecordSelection() map[uuid.UUID]bool {
	if p == nil {
		return nil
	}
	return p.Table.CopySelectionMap()
}

// ApplySelection locates the rows with the given UUIDs and selects them, replacing any existing selection.
func (p *PageList[T]) ApplySelection(selection map[uuid.UUID]bool) {
	if p != nil {
		p.Table.SetSelectionMap(selection)
	}
}

// Sync the underlying data.
func (p *PageList[T]) Sync() {
	p.provider.SyncHeader(p.tableHeader.ColumnHeaders)
	selection := p.RecordSelection()
	p.Table.SyncToModel()
	p.ApplySelection(selection)
	p.Table.NeedsLayout = true
	p.NeedsLayout = true
	if parent := p.Parent(); parent != nil {
		parent.NeedsLayout = true
	}
}

// CreateItem calls CreateItem on the contained TableProvider.
func (p *PageList[T]) CreateItem(owner widget.Rebuildable, variant ntable.ItemVariant) {
	p.provider.CreateItem(owner, p.Table, variant)
}

func (p *PageList[T]) OverheadHeight() float32 {
	_, pref, _ := p.tableHeader.Sizes(unison.Size{})
	insets := p.Border().Insets()
	return insets.Height() + pref.Height
}

func (p *PageList[T]) RowHeights() []float32 {
	return p.Table.RowHeights()
}

func (p *PageList[T]) RowCount() int {
	return p.Table.LastRowIndex() + 1
}

func (p *PageList[T]) CurrentDrawRowRange() (start, endBefore int) {
	return p.Table.CurrentDrawRowRange()
}

func (p *PageList[T]) SetDrawRowRange(start, endBefore int) {
	p.Table.SetDrawRowRange(start, endBefore)
}
