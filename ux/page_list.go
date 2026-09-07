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
	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

var (
	_ Syncer     = &PageList[*gurps.Trait]{}
	_ pageHelper = &PageList[*gurps.Trait]{}
	_ sheetList  = &PageList[*gurps.Trait]{}
)

// sheetList is what a dockable that shows page lists asks of each of them without regard to the type of row a list
// holds, so that its lists can be worked on as a group -- cleared, searched, disclosed and their selections carried
// across a rebuild -- by looping over the dockable's lists() rather than by naming each one. The selection and search
// methods tolerate a nil *PageList, since a dockable's lists() includes lists it hasn't built yet while it is first
// being put together.
type sheetList interface {
	unison.Paneler
	hierarchyDiscloser
	noteDiscloser
	clearSelection()
	search(refList *[]*searchRef, text string, namesOnly bool)
	RecordSelection() map[tid.TID]bool
	ApplySelection(selection map[tid.TID]bool)
}

// itemCreator is what the "New ..." commands ask of the list they add an item to.
type itemCreator interface {
	CreateItem(Rebuildable, ItemVariant)
}

// installNewItemCmdHandlers installs on the owner the handlers for the "New ..." commands that add an item to one of
// its lists: the plain item, and -- unless the container ID is -1, in which case the item is the alternate kind (a
// technique, a ritual magic spell) -- the container as well. The list is looked up through the getter each time a
// command is invoked rather than captured here, since a list whose set of columns has to change can only do so by being
// replaced outright (a table's columns are fixed at creation -- see PageList.needReconstruction), leaving whatever
// captured it holding an orphan. Creating an item in an orphaned list still updates the model, but everything that goes
// with it is aimed at a table nobody is looking at: the undo edit can't even find the undo manager, so the insertion
// isn't undoable and the user's next undo silently takes back the edit before it, and the new row is neither selected
// nor scrolled into view in the list that is actually on screen.
func installNewItemCmdHandlers(owner Rebuildable, itemID, containerID int, creator func() itemCreator) {
	p := owner.AsPanel()
	variant := NoItemVariant
	if containerID == -1 {
		variant = AlternateItemVariant
	} else {
		p.InstallCmdHandlers(containerID, unison.AlwaysEnabled,
			func(_ any) { creator().CreateItem(owner, ContainerItemVariant) })
	}
	p.InstallCmdHandlers(itemID, unison.AlwaysEnabled, func(_ any) { creator().CreateItem(owner, variant) })
}

// listItemCreators names the lists of a page dockable that the "New ..." commands add items to, each looked up when a
// command is invoked (see installNewItemCmdHandlers for why). A dockable leaves out the lists it doesn't have, and the
// commands for those are not installed, so they stay disabled while it is active.
type listItemCreators struct {
	traits           func() itemCreator
	skills           func() itemCreator
	spells           func() itemCreator
	carriedEquipment func() itemCreator
	otherEquipment   func() itemCreator
	notes            func() itemCreator
}

// installListItemCmdHandlers installs on the owner the handlers for the "New ..." commands for each of the lists it
// has: the item and the container for every list, and the alternate kind of item -- the technique, the ritual magic
// spell -- for the skills and the spells.
func installListItemCmdHandlers(owner Rebuildable, creators listItemCreators) {
	install := func(itemID, containerID int, creator func() itemCreator) {
		if creator != nil {
			installNewItemCmdHandlers(owner, itemID, containerID, creator)
		}
	}
	install(NewTraitItemID, NewTraitContainerItemID, creators.traits)
	install(NewSkillItemID, NewSkillContainerItemID, creators.skills)
	install(NewTechniqueItemID, -1, creators.skills)
	install(NewSpellItemID, NewSpellContainerItemID, creators.spells)
	install(NewRitualMagicSpellItemID, -1, creators.spells)
	install(NewCarriedEquipmentItemID, NewCarriedEquipmentContainerItemID, creators.carriedEquipment)
	install(NewOtherEquipmentItemID, NewOtherEquipmentContainerItemID, creators.otherEquipment)
	install(NewNoteItemID, NewNoteContainerItemID, creators.notes)
}

// installTraitListCmdHandlers installs on the owner the handlers for the commands that act on its trait list as a
// whole: adding the natural attacks and organizing the traits. The list is looked up when a command is invoked, for
// the same reason as in installNewItemCmdHandlers. The entity is the one the natural attacks are made for, and is nil
// for a template, whose traits have no entity until the template is applied.
func installTraitListCmdHandlers(owner Rebuildable, provider gurps.TraitListProvider, entity *gurps.Entity, traits func() *PageList[*gurps.Trait]) {
	p := owner.AsPanel()
	p.InstallCmdHandlers(AddNaturalAttacksItemID, unison.AlwaysEnabled, func(_ any) {
		InsertItems(owner, traits().Table, provider.TraitList, provider.SetTraitList,
			func(_ *unison.Table[*Node[*gurps.Trait]]) []*Node[*gurps.Trait] { return traits().provider.RootRows() },
			gurps.NewNaturalAttacks(entity, nil))
	})
	p.InstallCmdHandlers(OrganizeTraitsItemID, unison.AlwaysEnabled, func(_ any) { organizeTraits(owner, traits().Table) })
}

// listsForKeys returns the lists for the block keys the predicate accepts, in the canonical block order (see
// gurps.AllBlockKeys), each fetched with the given function.
func listsForKeys(list func(key string) sheetList, accept func(key string) bool) []sheetList {
	lists := make([]sheetList, 0, len(gurps.AllBlockKeys))
	for _, key := range gurps.AllBlockKeys {
		if accept(key) {
			lists = append(lists, list(key))
		}
	}
	return lists
}

// syncOrRebuildList brings a page list up to date with its model, building it anew when the columns it has to show no
// longer match the ones it has -- or when it doesn't exist yet -- since a table's columns are fixed at creation. The
// result is stored back through the pointer given, so that the caller's field always names the list that is on
// screen, and is returned as well. Anything that captured the old list has to allow for it having been replaced; see
// installNewItemCmdHandlers for what goes wrong when it doesn't.
func syncOrRebuildList[T gurps.Node[T]](list **PageList[T], build func() *PageList[T]) *PageList[T] {
	if (*list).needReconstruction() {
		*list = build()
	} else {
		(*list).Sync()
	}
	return *list
}

// preserveSelections records the selection of each of the lists the function returns and returns a function that puts
// the selections back. The lists are fetched again when the selections are put back rather than held from when they
// were recorded, since a rebuild replaces any list whose columns changed (see syncOrRebuildList) and the selection
// belongs in the list that is on screen, not in the orphan it replaced. Both calls therefore have to yield the lists in
// the same order.
func preserveSelections(lists func() []sheetList) (restore func()) {
	current := lists()
	selections := make([]map[tid.TID]bool, len(current))
	for i, list := range current {
		selections[i] = list.RecordSelection()
	}
	return func() {
		for i, list := range lists() {
			if i < len(selections) {
				list.ApplySelection(selections[i])
			}
		}
	}
}

// PageList holds a list for a sheet page.
type PageList[T gurps.Node[T]] struct {
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
	InstallTintFunc(p, colors.TintTraits)
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
	installEquipmentLevelHandlers(p, owner)
	InstallTintFunc(p, colors.TintCarriedEquipment)
	return p
}

// NewOtherEquipmentPageList creates the other equipment page list.
func NewOtherEquipmentPageList(owner Rebuildable, provider gurps.ListProvider) *PageList[*gurps.Equipment] {
	p := newPageList(owner, NewEquipmentProvider(provider, false, true))
	p.installIncrementTechLevelHandler(owner)
	p.installDecrementTechLevelHandler(owner)
	p.installContainerConversionHandlers(owner)
	p.installMoveToCarriedEquipmentHandler(owner)
	installEquipmentLevelHandlers(p, owner)
	InstallTintFunc(p, colors.TintOtherEquipment)
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
	InstallTintFunc(p, colors.TintSkills)
	return p
}

// NewSpellsPageList creates the spells page list.
func NewSpellsPageList(owner Rebuildable, provider gurps.SpellListProvider) *PageList[*gurps.Spell] {
	p := newPageList(owner, NewSpellsProvider(provider, true))
	p.installIncrementPointsHandler(owner)
	p.installDecrementPointsHandler(owner)
	p.installIncrementSkillHandler(owner)
	p.installDecrementSkillHandler(owner)
	InstallTintFunc(p, colors.TintSpells)
	return p
}

// NewNotesPageList creates the notes page list.
func NewNotesPageList(owner Rebuildable, provider gurps.ListProvider) *PageList[*gurps.Note] {
	p := newPageList(owner, NewNotesProvider(provider, true))
	p.installContainerConversionHandlers(owner)
	InstallTintFunc(p, colors.TintNotes)
	return p
}

// NewConditionalModifiersPageList creates the conditional modifiers page list.
func NewConditionalModifiersPageList(entity *gurps.Entity) *PageList[*gurps.ConditionalModifier] {
	p := newPageList(nil, NewConditionalModifiersProvider(entity))
	InstallTintFunc(p, colors.TintConditions)
	return p
}

// NewReactionsPageList creates the reaction modifiers page list.
func NewReactionsPageList(entity *gurps.Entity) *PageList[*gurps.ConditionalModifier] {
	p := newPageList(nil, NewReactionModifiersProvider(entity))
	InstallTintFunc(p, colors.TintReactions)
	return p
}

// NewMeleeWeaponsPageList creates the melee weapons page list.
func NewMeleeWeaponsPageList(entity *gurps.Entity) *PageList[*gurps.Weapon] {
	p := newPageList(nil, NewWeaponsProvider(entity, true, true))
	InstallTintFunc(p, colors.TintMelee)
	return p
}

// NewRangedWeaponsPageList creates the ranged weapons page list.
func NewRangedWeaponsPageList(entity *gurps.Entity) *PageList[*gurps.Weapon] {
	p := newPageList(nil, NewWeaponsProvider(entity, false, true))
	InstallTintFunc(p, colors.TintRanged)
	return p
}

func newPageList[T gurps.Node[T]](owner Rebuildable, provider TableProvider[T]) *PageList[T] {
	header, table := NewNodeTable(provider, fonts.PageFieldPrimary)
	table.ClientData()[WorkingDirKey] = WorkingDirProvider(owner)
	editable := !xreflect.IsNil(owner)
	if editable {
		table.ClientData()[TableOwnerClientKey] = owner
	}
	table.RefKey = provider.RefKey()
	p := &PageList[T]{
		tableHeader: header,
		Table:       table,
		provider:    provider,
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{Columns: 1})
	p.SetBorder(unison.NewLineBorder(header.BackgroundInk, geom.Size{}, geom.NewUniformInsets(1), false))

	p.Table.PreventUserColumnResize = true
	p.Table.ShowLastColumnDivider = false
	p.Table.SyncToModel()
	p.AddChild(p.tableHeader)
	p.AddChild(p.Table)
	// The read-only lists -- the weapons and the conditional and reaction modifiers, which are derived from the rest
	// of the sheet -- have no owner and offer none of the editing commands.
	installStandardTableCmdHandlers(p, p.Table, p.provider, func() Rebuildable { return owner }, editable)
	if editable {
		InstallTableDropSupport(p.Table, p.provider)
	}
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
	})
	return p
}

func (p *PageList[T]) needReconstruction() bool {
	return p == nil || columnsOutOfSync(p.provider.ColumnIDs(), p.Table.Columns)
}

// undoData collects the undo edit data for the list's table, stripped of the row type so that an edit spanning lists
// of different row types can hold it (see tablesUndoData). Nothing comes back for a list that doesn't exist or whose
// data couldn't be collected.
func (p *PageList[T]) undoData() tableRestorer {
	if p == nil {
		return nil
	}
	data := NewTableUndoEditData(p.Table)
	if data == nil {
		// Returned as an untyped nil, so that the caller's nil check sees it: a nil *TableUndoEditData wrapped in the
		// interface would not be nil.
		return nil
	}
	return data
}

// syncToModel brings the list's table up to date with its model. A list that doesn't exist has nothing to bring up to
// date.
func (p *PageList[T]) syncToModel() {
	if p != nil {
		p.Table.SyncToModel()
	}
}

// columnsOutOfSync returns true if the columns a table is currently showing no longer match the column IDs its provider
// wants, which means the table has to be built anew, since a table's columns are fixed at creation.
func columnsOutOfSync(ids []int, columns []unison.ColumnInfo) bool {
	if len(ids) != len(columns) {
		return true
	}
	for i, col := range columns {
		if col.ID != ids[i] {
			return true
		}
	}
	return false
}

func (p *PageList[T]) installMoveToCarriedEquipmentHandler(owner Rebuildable) {
	if sheet, ok := owner.AsPanel().Self.(*Sheet); ok {
		var t *unison.Table[*Node[*gurps.Equipment]]
		if t, ok = any(p.Table).(*unison.Table[*Node[*gurps.Equipment]]); ok {
			p.InstallCmdHandlers(MoveToCarriedEquipmentItemID,
				func(_ any) bool { return t.HasSelection() },
				func(_ any) { moveSelectedEquipment(sheet, t, sheet.CarriedEquipment.Table) })
		}
	}
}

func (p *PageList[T]) installMoveToOtherEquipmentHandler(owner Rebuildable) {
	if sheet, ok := owner.AsPanel().Self.(*Sheet); ok {
		var t *unison.Table[*Node[*gurps.Equipment]]
		if t, ok = any(p.Table).(*unison.Table[*Node[*gurps.Equipment]]); ok {
			p.InstallCmdHandlers(MoveToOtherEquipmentItemID,
				func(_ any) bool { return t.HasSelection() },
				func(_ any) { moveSelectedEquipment(sheet, t, sheet.OtherEquipment.Table) })
		}
	}
}

// moveSelectedEquipment moves the rows selected in one of the sheet's two equipment lists to the other one. The copy
// into the destination and the deletion from the source are each told not to report the change, so that the sheet is
// rebuilt once for the move as a whole rather than once per half of it -- and so that both tables are still the ones
// on screen when the undo data is collected, since a rebuild can replace either list (see liveTable).
func moveSelectedEquipment(sheet *Sheet, from, to *unison.Table[*Node[*gurps.Equipment]]) {
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
	copyRowsTo(to, from.SelectedRows(true), nil, false, false)
	deleteSelection(from, false, false)
	undo.AfterData = NewTableDragUndoEditData(from, to)
	mgr.Add(undo)
	rebuildAsModified(sheet, true)
}

func (p *PageList[T]) installToggleDisabledHandler(owner Rebuildable) {
	if t, ok := any(p.Table).(*unison.Table[*Node[*gurps.Trait]]); ok {
		p.InstallCmdHandlers(ToggleStateItemID,
			func(_ any) bool { return canToggleDisabled(t) },
			func(_ any) { toggleDisabled(owner, t) })
	}
}

func (p *PageList[T]) installToggleEquippedHandler(owner Rebuildable) {
	if t, ok := any(p.Table).(*unison.Table[*Node[*gurps.Equipment]]); ok {
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
	if t, ok := any(p.Table).(*unison.Table[*Node[*gurps.Trait]]); ok {
		p.InstallCmdHandlers(IncrementItemID,
			func(_ any) bool { return canAdjustTraitLevel(t, true) },
			func(_ any) { adjustTraitLevel(owner, t, true) })
	}
}

func (p *PageList[T]) installDecrementLevelHandler(owner Rebuildable) {
	if t, ok := any(p.Table).(*unison.Table[*Node[*gurps.Trait]]); ok {
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

func installEquipmentLevelHandlers(p *PageList[*gurps.Equipment], owner Rebuildable) {
	p.InstallCmdHandlers(IncrementEquipmentLevelItemID,
		func(_ any) bool { return canAdjustEquipmentLevel(p.Table, fxp.One) },
		func(_ any) { adjustEquipmentLevel(owner, p.Table, fxp.One) })
	p.InstallCmdHandlers(DecrementEquipmentLevelItemID,
		func(_ any) bool { return canAdjustEquipmentLevel(p.Table, -fxp.One) },
		func(_ any) { adjustEquipmentLevel(owner, p.Table, -fxp.One) })
}

func (p *PageList[T]) installContainerConversionHandlers(owner Rebuildable) {
	InstallContainerConversionHandlers(p, owner, p.Table)
}

// SelectedNodes returns the set of selected nodes. If 'minimal' is true, then children of selected rows that may also
// be selected are not returned, just the topmost row that is selected in any given hierarchy.
func (p *PageList[T]) SelectedNodes(minimal bool) []*Node[T] {
	if p == nil {
		return nil
	}
	return p.Table.SelectedRows(minimal)
}

// clearSelection deselects every row in the list. A list that doesn't exist has nothing to deselect.
func (p *PageList[T]) clearSelection() {
	if p != nil {
		p.Table.ClearSelection()
	}
}

// RecordSelection collects the currently selected row IDs.
func (p *PageList[T]) RecordSelection() map[tid.TID]bool {
	if p == nil {
		return nil
	}
	return p.Table.CopySelectionMap()
}

// ApplySelection locates the rows with the given IDs and selects them, replacing any existing selection.
func (p *PageList[T]) ApplySelection(selection map[tid.TID]bool) {
	if p != nil {
		p.Table.SetSelectionMap(selection)
	}
}

// FirstDisclosureState implements hierarchyDiscloser.
func (p *PageList[T]) FirstDisclosureState() (open, exists bool) {
	return firstTableDisclosureState(p.Table)
}

// SetDisclosureState implements hierarchyDiscloser.
func (p *PageList[T]) SetDisclosureState(open bool) {
	setTableDisclosureState(p.Table, open)
}

// FirstNoteState implements noteDiscloser.
func (p *PageList[T]) FirstNoteState() int {
	return firstTableNoteState(p.Table)
}

// ApplyNoteState implements noteDiscloser.
func (p *PageList[T]) ApplyNoteState(closed bool) {
	applyTableNoteState(p.Table, closed)
}

// Sync the underlying data.
func (p *PageList[T]) Sync() {
	p.provider.SyncHeader(p.tableHeader.ColumnHeaders)
	selection := p.RecordSelection()
	p.Table.SyncToModel()
	p.ApplySelection(selection)
	// Re-evaluate page reference column expansion here, since a change to that setting won't necessarily alter the
	// table's frame and thus won't trigger the table's FrameChangeCallback.
	sizePageTableColumns(p.Table, p.provider.ExcessWidthColumnID())
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
	_, pref, _ := p.tableHeader.Sizes(geom.Size{})
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
