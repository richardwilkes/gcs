/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
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
	"github.com/google/uuid"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wpn"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

var (
	_ Syncer     = &PageList[*gurps.Trait]{}
	_ pageHelper = &PageList[*gurps.Trait]{}
)

// PageList holds a list for a sheet page.
type PageList[T gurps.NodeTypes] struct {
	unison.Panel
	tableHeader *unison.TableHeader[*Node[T]]
	Table       *unison.Table[*Node[T]]
	provider    TableProvider[T]
}

// NewTraitsPageList creates the traits page list.
func NewTraitsPageList(owner Rebuildable, provider gurps.ListProvider) *PageList[*gurps.Trait] {
	p := newPageList(owner, NewTraitsProvider(provider, true))
	p.installToggleDisabledHandler(owner)
	p.installIncrementLevelHandler(owner)
	p.installDecrementLevelHandler(owner)
	return p
}

// NewCarriedEquipmentPageList creates the carried equipment page list.
func NewCarriedEquipmentPageList(owner Rebuildable, provider gurps.ListProvider) *PageList[*gurps.Equipment] {
	p := newPageList(owner, NewEquipmentProvider(provider, true, true))
	p.installToggleEquippedHandler(owner)
	p.installIncrementTechLevelHandler(owner)
	p.installDecrementTechLevelHandler(owner)
	p.installContainerConversionHandlers(owner)
	p.installMoveToOtherEquipmentHandler(owner)
	return p
}

// NewOtherEquipmentPageList creates the other equipment page list.
func NewOtherEquipmentPageList(owner Rebuildable, provider gurps.ListProvider) *PageList[*gurps.Equipment] {
	p := newPageList(owner, NewEquipmentProvider(provider, false, true))
	p.installIncrementTechLevelHandler(owner)
	p.installDecrementTechLevelHandler(owner)
	p.installContainerConversionHandlers(owner)
	p.installMoveToCarriedEquipmentHandler(owner)
	return p
}

// NewSkillsPageList creates the skills page list.
func NewSkillsPageList(owner Rebuildable, provider gurps.ListProvider) *PageList[*gurps.Skill] {
	p := newPageList(owner, NewSkillsProvider(provider, true))
	p.installIncrementPointsHandler(owner)
	p.installDecrementPointsHandler(owner)
	p.installIncrementSkillHandler(owner)
	p.installDecrementSkillHandler(owner)
	p.installIncrementTechLevelHandler(owner)
	p.installDecrementTechLevelHandler(owner)
	return p
}

// NewSpellsPageList creates the spells page list.
func NewSpellsPageList(owner Rebuildable, provider gurps.SpellListProvider) *PageList[*gurps.Spell] {
	p := newPageList(owner, NewSpellsProvider(provider, true))
	p.installIncrementPointsHandler(owner)
	p.installDecrementPointsHandler(owner)
	p.installIncrementSkillHandler(owner)
	p.installDecrementSkillHandler(owner)
	return p
}

// NewNotesPageList creates the notes page list.
func NewNotesPageList(owner Rebuildable, provider gurps.ListProvider) *PageList[*gurps.Note] {
	p := newPageList(owner, NewNotesProvider(provider, true))
	p.installContainerConversionHandlers(owner)
	return p
}

// NewConditionalModifiersPageList creates the conditional modifiers page list.
func NewConditionalModifiersPageList(entity *gurps.Entity) *PageList[*gurps.ConditionalModifier] {
	return newPageList(nil, NewConditionalModifiersProvider(entity))
}

// NewReactionsPageList creates the reaction modifiers page list.
func NewReactionsPageList(entity *gurps.Entity) *PageList[*gurps.ConditionalModifier] {
	return newPageList(nil, NewReactionModifiersProvider(entity))
}

// NewMeleeWeaponsPageList creates the melee weapons page list.
func NewMeleeWeaponsPageList(entity *gurps.Entity) *PageList[*gurps.Weapon] {
	return newPageList(nil, NewWeaponsProvider(entity, wpn.Melee, true))
}

// NewRangedWeaponsPageList creates the ranged weapons page list.
func NewRangedWeaponsPageList(entity *gurps.Entity) *PageList[*gurps.Weapon] {
	return newPageList(nil, NewWeaponsProvider(entity, wpn.Ranged, true))
}

func newPageList[T gurps.NodeTypes](owner Rebuildable, provider TableProvider[T]) *PageList[T] {
	header, table := NewNodeTable[T](provider, gurps.PageFieldPrimaryFont)
	table.ClientData()[WorkingDirKey] = WorkingDirProvider(owner)
	table.RefKey = provider.RefKey()
	p := &PageList[T]{
		tableHeader: header,
		Table:       table,
		provider:    provider,
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{Columns: 1})
	p.SetBorder(unison.NewLineBorder(gurps.HeaderColor, 0, unison.NewUniformInsets(1), false))

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
			gc.DrawRect(dirty, p.tableHeader.BackgroundInk.Paint(gc, dirty, paintstyle.Fill))
			r := p.tableHeader.ColumnFrame(sortedOn)
			r.X -= p.Table.Padding.Left
			r.Width += p.Table.Padding.Left + p.Table.Padding.Right
			gc.DrawRect(r, gurps.MarkerColor.Paint(gc, r, paintstyle.Fill))
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
		InstallTableDropSupport(p.Table, p.provider)
		p.InstallCmdHandlers(OpenEditorItemID,
			func(_ any) bool { return p.Table.HasSelection() },
			func(_ any) { p.provider.OpenEditor(owner, p.Table) })
		p.InstallCmdHandlers(unison.DeleteItemID,
			func(_ any) bool { return p.Table.HasSelection() },
			func(_ any) { DeleteSelection(p.Table, true) })
		p.InstallCmdHandlers(DuplicateItemID,
			func(_ any) bool { return p.Table.HasSelection() },
			func(_ any) { DuplicateSelection(p.Table) })
	}
	p.installOpenPageReferenceHandlers()
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
	})
	return p
}

func (p *PageList[T]) installMoveToCarriedEquipmentHandler(owner Rebuildable) {
	if sheet, ok := owner.AsPanel().Self.(*Sheet); ok {
		var t *unison.Table[*Node[*gurps.Equipment]]
		if t, ok = (any(p.Table)).(*unison.Table[*Node[*gurps.Equipment]]); ok {
			p.InstallCmdHandlers(MoveToCarriedEquipmentItemID,
				func(_ any) bool { return t.HasSelection() },
				func(_ any) { moveSelectedEquipment(t, sheet.CarriedEquipment.Table) })
		}
	}
}

func (p *PageList[T]) installMoveToOtherEquipmentHandler(owner Rebuildable) {
	if sheet, ok := owner.AsPanel().Self.(*Sheet); ok {
		var t *unison.Table[*Node[*gurps.Equipment]]
		if t, ok = (any(p.Table)).(*unison.Table[*Node[*gurps.Equipment]]); ok {
			p.InstallCmdHandlers(MoveToOtherEquipmentItemID,
				func(_ any) bool { return t.HasSelection() },
				func(_ any) { moveSelectedEquipment(t, sheet.OtherEquipment.Table) })
		}
	}
}

func moveSelectedEquipment(from, to *unison.Table[*Node[*gurps.Equipment]]) {
	mgr := unison.UndoManagerFor(from)
	if mgr == nil || mgr != unison.UndoManagerFor(to) {
		return
	}
	undo := &unison.UndoEdit[*TableDragUndoEditData[*gurps.Equipment]]{
		ID:       unison.NextUndoID(),
		EditName: i18n.Text("Move Equipment"),
		UndoFunc: func(e *unison.UndoEdit[*TableDragUndoEditData[*gurps.Equipment]]) { e.BeforeData.Apply() },
		RedoFunc: func(e *unison.UndoEdit[*TableDragUndoEditData[*gurps.Equipment]]) { e.AfterData.Apply() },
		AbsorbFunc: func(_ *unison.UndoEdit[*TableDragUndoEditData[*gurps.Equipment]], _ unison.Undoable) bool {
			return false
		},
		BeforeData: NewTableDragUndoEditData(from, to),
	}
	CopyRowsTo(to, from.SelectedRows(true), nil, false)
	DeleteSelection(from, false)
	undo.AfterData = NewTableDragUndoEditData(from, to)
	mgr.Add(undo)
}

func (p *PageList[T]) installOpenPageReferenceHandlers() {
	p.InstallCmdHandlers(OpenOnePageReferenceItemID,
		func(_ any) bool { return CanOpenPageRef(p.Table) },
		func(_ any) { OpenPageRef(p.Table) })
	p.InstallCmdHandlers(OpenEachPageReferenceItemID,
		func(_ any) bool { return CanOpenPageRef(p.Table) },
		func(_ any) { OpenEachPageRef(p.Table) })
}

func (p *PageList[T]) installToggleDisabledHandler(owner Rebuildable) {
	if t, ok := (any(p.Table)).(*unison.Table[*Node[*gurps.Trait]]); ok {
		p.InstallCmdHandlers(ToggleStateItemID,
			func(_ any) bool { return canToggleDisabled(t) },
			func(_ any) { toggleDisabled(owner, t) })
	}
}

func (p *PageList[T]) installToggleEquippedHandler(owner Rebuildable) {
	if t, ok := (any(p.Table)).(*unison.Table[*Node[*gurps.Equipment]]); ok {
		p.InstallCmdHandlers(ToggleStateItemID,
			func(_ any) bool { return canToggleEquipped(t) },
			func(_ any) { toggleEquipped(owner, t) })
	}
}

func (p *PageList[T]) installIncrementPointsHandler(owner Rebuildable) {
	p.InstallCmdHandlers(IncrementItemID,
		func(_ any) bool { return canAdjustRawPoints(p.Table, true) },
		func(_ any) { adjustRawPoints(owner, p.Table, true) })
}

func (p *PageList[T]) installDecrementPointsHandler(owner Rebuildable) {
	p.InstallCmdHandlers(DecrementItemID,
		func(_ any) bool { return canAdjustRawPoints(p.Table, false) },
		func(_ any) { adjustRawPoints(owner, p.Table, false) })
}

func (p *PageList[T]) installIncrementLevelHandler(owner Rebuildable) {
	if t, ok := (any(p.Table)).(*unison.Table[*Node[*gurps.Trait]]); ok {
		p.InstallCmdHandlers(IncrementItemID,
			func(_ any) bool { return canAdjustTraitLevel(t, true) },
			func(_ any) { adjustTraitLevel(owner, t, true) })
	}
}

func (p *PageList[T]) installDecrementLevelHandler(owner Rebuildable) {
	if t, ok := (any(p.Table)).(*unison.Table[*Node[*gurps.Trait]]); ok {
		p.InstallCmdHandlers(DecrementItemID,
			func(_ any) bool { return canAdjustTraitLevel(t, false) },
			func(_ any) { adjustTraitLevel(owner, t, false) })
	}
}

func (p *PageList[T]) installIncrementSkillHandler(owner Rebuildable) {
	p.InstallCmdHandlers(IncrementSkillLevelItemID,
		func(_ any) bool { return canAdjustSkillLevel(p.Table, true) },
		func(_ any) { adjustSkillLevel(owner, p.Table, true) })
}

func (p *PageList[T]) installDecrementSkillHandler(owner Rebuildable) {
	p.InstallCmdHandlers(DecrementSkillLevelItemID,
		func(_ any) bool { return canAdjustSkillLevel(p.Table, false) },
		func(_ any) { adjustSkillLevel(owner, p.Table, false) })
}

func (p *PageList[T]) installIncrementTechLevelHandler(owner Rebuildable) {
	p.InstallCmdHandlers(IncrementTechLevelItemID,
		func(_ any) bool { return canAdjustTechLevel(p.Table, fxp.One) },
		func(_ any) { adjustTechLevel(owner, p.Table, fxp.One) })
}

func (p *PageList[T]) installDecrementTechLevelHandler(owner Rebuildable) {
	p.InstallCmdHandlers(DecrementTechLevelItemID,
		func(_ any) bool { return canAdjustTechLevel(p.Table, -fxp.One) },
		func(_ any) { adjustTechLevel(owner, p.Table, -fxp.One) })
}

func (p *PageList[T]) installContainerConversionHandlers(owner Rebuildable) {
	if t, ok := (any(p.Table)).(*unison.Table[*Node[*gurps.Equipment]]); ok {
		InstallContainerConversionHandlers(p, owner, t)
		return
	}
	if t, ok := (any(p.Table)).(*unison.Table[*Node[*gurps.Note]]); ok {
		InstallContainerConversionHandlers(p, owner, t)
	}
}

// SelectedNodes returns the set of selected nodes. If 'minimal' is true, then children of selected rows that may also
// be selected are not returned, just the topmost row that is selected in any given hierarchy.
func (p *PageList[T]) SelectedNodes(minimal bool) []*Node[T] {
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
func (p *PageList[T]) CreateItem(owner Rebuildable, variant ItemVariant) {
	p.provider.CreateItem(owner, p.Table, variant)
}

// OverheadHeight returns the overhead for this page list, i.e. the border and header space.
func (p *PageList[T]) OverheadHeight() float32 {
	_, pref, _ := p.tableHeader.Sizes(unison.Size{})
	insets := p.Border().Insets()
	return insets.Height() + pref.Height
}

// RowHeights returns the heights of each row.
func (p *PageList[T]) RowHeights() []float32 {
	return p.Table.RowHeights()
}

// RowCount returns the number of rows.
func (p *PageList[T]) RowCount() int {
	return p.Table.LastRowIndex() + 1
}

// CurrentDrawRowRange returns the current row range that will be drawn.
func (p *PageList[T]) CurrentDrawRowRange() (start, endBefore int) {
	return p.Table.CurrentDrawRowRange()
}

// SetDrawRowRange sets the row range that will be drawn.
func (p *PageList[T]) SetDrawRowRange(start, endBefore int) {
	p.Table.SetDrawRowRange(start, endBefore)
}
