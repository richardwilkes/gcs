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
	"github.com/richardwilkes/unison/drag"
	"github.com/richardwilkes/unison/enums/align"
)

var (
	_ Syncer     = &PageList[*gurps.Trait]{}
	_ pageHelper = &PageList[*gurps.Trait]{}
	_ sheetList  = &PageList[*gurps.Trait]{}
	_ pageHelper = &pageListContinuation[*gurps.Trait]{}
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
	search(matches *[]*searchMatch, text, title string)
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
// command is invoked rather than captured here, since a list whose columns change is replaced outright (see
// PageList.needReconstruction). Creating an item in a captured, orphaned list still updates the model, but its undo
// edit can't find the undo manager, so the user's next undo silently takes back the edit before it, and the new row is
// neither selected nor scrolled into view in the list on screen.
func installNewItemCmdHandlers(owner Rebuildable, itemID, containerID int, creator func() itemCreator) {
	p := owner.AsPanel()
	variant := NoItemVariant
	if containerID == -1 {
		variant = AlternateItemVariant
	} else {
		installCreateItemCmdHandler(p, containerID, owner, ContainerItemVariant, creator)
	}
	installCreateItemCmdHandler(p, itemID, owner, variant, creator)
}

// installCreateItemCmdHandler installs on the panel the handler for a command that creates an item of the given variant
// in the list the creator returns. The command is disabled while the creator returns nil, which a dockable does for a
// list that can't take new items just now.
func installCreateItemCmdHandler(p *unison.Panel, id int, owner Rebuildable, variant ItemVariant,
	creator func() itemCreator,
) {
	p.InstallCmdHandlers(id,
		func(_ any) bool { return !xreflect.IsNil(creator()) },
		func(_ any) {
			if c := creator(); !xreflect.IsNil(c) {
				c.CreateItem(owner, variant)
			}
		})
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
// has: the item and the container for every list, the alternate kind of item -- the technique, the ritual magic spell
// -- for the skills and the spells, and the group for the equipment lists.
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
	installGroup := func(id int, creator func() itemCreator) {
		if creator != nil {
			installCreateItemCmdHandler(owner.AsPanel(), id, owner, GroupContainerItemVariant, creator)
		}
	}
	installGroup(NewCarriedEquipmentGroupItemID, creators.carriedEquipment)
	installGroup(NewOtherEquipmentGroupItemID, creators.otherEquipment)
}

// installChoiceContainerCmdHandlers installs on the owner the handlers for the "New ... Choice" commands for each of
// the lists it has whose nodes may be one. Only a template may hold template choices, so only a template installs
// these, which keeps the commands disabled, and out of the context menus, everywhere else.
func installChoiceContainerCmdHandlers(owner Rebuildable, creators listItemCreators) {
	p := owner.AsPanel()
	install := func(id int, creator func() itemCreator) {
		if creator != nil {
			installCreateItemCmdHandler(p, id, owner, ChoiceContainerItemVariant, creator)
		}
	}
	install(NewTraitChoiceContainerItemID, creators.traits)
	install(NewEquipmentChoiceContainerItemID, creators.carriedEquipment)
	install(NewSkillChoiceContainerItemID, creators.skills)
	install(NewSpellChoiceContainerItemID, creators.spells)
}

// installTraitListCmdHandlers installs on the owner the handlers for the commands that act on its trait list as a
// whole: adding the natural attacks and organizing the traits. The list is looked up when a command is invoked, for
// the same reason as in installNewItemCmdHandlers, and the commands are disabled while the lookup returns nil, which
// it does while the layout leaves the list off the page (see onPageCreator). The entity is the one the natural attacks
// are made for, and is nil for a template, whose traits have no entity until the template is applied.
func installTraitListCmdHandlers(owner Rebuildable, provider gurps.TraitListProvider, entity *gurps.Entity, traits func() *PageList[*gurps.Trait]) {
	p := owner.AsPanel()
	enabled := func(_ any) bool { return traits() != nil }
	p.InstallCmdHandlers(AddNaturalAttacksItemID, enabled, func(_ any) {
		InsertItems(owner, traits().Table, provider.TraitList, provider.SetTraitList,
			func(_ *unison.Table[*Node[*gurps.Trait]]) []*Node[*gurps.Trait] { return traits().provider.RootRows() },
			gurps.NewNaturalAttacks(entity, nil))
	})
	p.InstallCmdHandlers(OrganizeTraitsItemID, enabled, func(_ any) { organizeTraits(owner, traits().Table) })
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

// searchableListsForKeys returns the lists for the block keys accept admits, as listsForKeys does, each with the title
// the search shows beside its matches.
func searchableListsForKeys(list func(key string) sheetList, accept func(key string) bool, title func(key string) string) []searchableList {
	lists := make([]searchableList, 0, len(gurps.AllBlockKeys))
	for _, key := range gurps.AllBlockKeys {
		if accept(key) {
			lists = append(lists, searchableList{list: list(key), title: title(key)})
		}
	}
	return lists
}

// syncOrRebuildList brings a page list up to date with its model, building it anew when it doesn't exist yet or the
// columns it has to show no longer match the ones it has, since a table's columns are fixed at creation. The result is
// stored through the pointer, so the caller's field always names the list on screen, and returned. Anything that
// captured the old list has to allow for its replacement (see installNewItemCmdHandlers).
func syncOrRebuildList[T gurps.Node[T]](list **PageList[T], build func() *PageList[T]) *PageList[T] {
	if (*list).needReconstruction() {
		*list = build()
	} else {
		(*list).Sync()
	}
	return *list
}

// preserveSelections records the selection of each of the lists the function returns and returns a function that puts
// the selections back. The lists are fetched again when restoring, since a rebuild replaces any list whose columns
// changed (see syncOrRebuildList) and the selection belongs in the list on screen; both calls must therefore yield the
// lists in the same order.
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

// PageList holds a list for a sheet page. The rows that don't fit on the page the list is on go on in continuations of
// it (see continuation), each showing a view of the list's table.
type PageList[T gurps.Node[T]] struct {
	unison.Panel
	tableHeader *unison.TableHeader[*Node[T]]
	Table       *unison.Table[*Node[T]]
	provider    TableProvider[T]
	owner       Rebuildable // nil for a read-only list
	tint        *unison.ThemeColor
	// installKindHandlers installs the commands only this kind of list offers on the list and on its continuations.
	installKindHandlers kindHandlerInstaller[T]
	continuations       []*pageListContinuation[T]
	syncedAt            uint64 // The display stamp the list was last synced at; see syncedAtStamp.
}

// kindHandlerInstaller installs the commands only one kind of page list offers on a panel showing a view of the list's
// table: the list itself, or one of its continuations.
type kindHandlerInstaller[T gurps.Node[T]] func(target *unison.Panel, table *unison.Table[*Node[T]])

// NewTraitsPageList creates the traits page list.
func NewTraitsPageList(owner Rebuildable, provider gurps.ListProvider) *PageList[*gurps.Trait] {
	return newPageList(owner, NewTraitsProvider(provider, true), colors.TintTraits,
		func(target *unison.Panel, table *unison.Table[*Node[*gurps.Trait]]) {
			installToggleDisabledHandler(target, table, owner)
			installTraitLevelHandlers(target, table, owner)
			installChoiceConversionHandlers[*gurps.Trait, *gurps.TraitEditData](target, table, owner)
		})
}

// NewCarriedEquipmentPageList creates the carried equipment page list.
func NewCarriedEquipmentPageList(owner Rebuildable, provider gurps.ListProvider) *PageList[*gurps.Equipment] {
	return newPageList(owner, NewEquipmentProvider(provider, true, true), colors.TintCarriedEquipment,
		func(target *unison.Panel, table *unison.Table[*Node[*gurps.Equipment]]) {
			installToggleEquippedHandler(target, table, owner)
			installTechLevelHandlers(target, table, owner)
			InstallContainerConversionHandlers(target, owner, table)
			installEquipmentContainerConversionHandlers(target, table, owner)
			installMoveToOtherEquipmentHandler(target, table, owner)
			installSellEquipmentHandler(target, table, owner)
			installBuyEquipmentHandler(target, table)
			installBuyDragDeferral(table, owner)
			installEquipmentLevelHandlers(target, table, owner)
		})
}

// NewOtherEquipmentPageList creates the other equipment page list.
func NewOtherEquipmentPageList(owner Rebuildable, provider gurps.ListProvider) *PageList[*gurps.Equipment] {
	return newPageList(owner, NewEquipmentProvider(provider, false, true), colors.TintOtherEquipment,
		func(target *unison.Panel, table *unison.Table[*Node[*gurps.Equipment]]) {
			installTechLevelHandlers(target, table, owner)
			InstallContainerConversionHandlers(target, owner, table)
			installEquipmentContainerConversionHandlers(target, table, owner)
			installMoveToCarriedEquipmentHandler(target, table, owner)
			installBuyEquipmentHandler(target, table)
			installBuyDragDeferral(table, owner)
			installEquipmentLevelHandlers(target, table, owner)
		})
}

// installBuyDragDeferral has a view of a sheet's list decline the equipment drags the sheet's buy dockable would take
// while it is open, so that the drag goes on up to the sheet, which reroutes it to the dockable (see Sheet.keyToPanel):
// equipment dragged anywhere onto a sheet that is buying is to be bought, not carried. The sheet's own equipment is
// still taken, so rows can be moved about while buying; dropped on the dockable itself, it is bought like any other
// (see buyDockable.claimsDrag). The view must already accept drops (see InstallTableDropSupport).
func installBuyDragDeferral[T gurps.Node[T]](table *unison.Table[*Node[T]], owner Rebuildable) {
	if sheet, ok := owner.AsPanel().Self.(*Sheet); ok {
		accept := table.CanAcceptDropCallback
		table.CanAcceptDropCallback = func(di drag.Info) bool {
			if d := buyDockableFor(sheet); d != nil && d.claimsDrag(di) {
				return false
			}
			return accept(di)
		}
	}
}

// NewSkillsPageList creates the skills page list.
func NewSkillsPageList(owner Rebuildable, provider gurps.ListProvider) *PageList[*gurps.Skill] {
	return newPageList(owner, NewSkillsProvider(provider, true), colors.TintSkills,
		func(target *unison.Panel, table *unison.Table[*Node[*gurps.Skill]]) {
			installPointsHandlers(target, table, owner)
			installSkillLevelHandlers(target, table, owner)
			installTechLevelHandlers(target, table, owner)
			installChoiceConversionHandlers[*gurps.Skill, *gurps.SkillEditData](target, table, owner)
		})
}

// NewSpellsPageList creates the spells page list.
func NewSpellsPageList(owner Rebuildable, provider gurps.SpellListProvider) *PageList[*gurps.Spell] {
	return newPageList(owner, NewSpellsProvider(provider, true), colors.TintSpells,
		func(target *unison.Panel, table *unison.Table[*Node[*gurps.Spell]]) {
			installPointsHandlers(target, table, owner)
			installSkillLevelHandlers(target, table, owner)
			installChoiceConversionHandlers[*gurps.Spell, *gurps.SpellEditData](target, table, owner)
		})
}

// NewNotesPageList creates the notes page list.
func NewNotesPageList(owner Rebuildable, provider gurps.ListProvider) *PageList[*gurps.Note] {
	return newPageList(owner, NewNotesProvider(provider, true), colors.TintNotes,
		func(target *unison.Panel, table *unison.Table[*Node[*gurps.Note]]) {
			InstallContainerConversionHandlers(target, owner, table)
		})
}

// NewConditionalModifiersPageList creates the conditional modifiers page list.
func NewConditionalModifiersPageList(entity *gurps.Entity) *PageList[*gurps.ConditionalModifier] {
	return newPageList(nil, NewConditionalModifiersProvider(entity), colors.TintConditions, nil)
}

// NewReactionsPageList creates the reaction modifiers page list.
func NewReactionsPageList(entity *gurps.Entity) *PageList[*gurps.ConditionalModifier] {
	return newPageList(nil, NewReactionModifiersProvider(entity), colors.TintReactions, nil)
}

// NewMeleeWeaponsPageList creates the melee weapons page list.
func NewMeleeWeaponsPageList(entity *gurps.Entity) *PageList[*gurps.Weapon] {
	return newPageList(nil, NewWeaponsProvider(entity, true, true), colors.TintMelee, nil)
}

// NewRangedWeaponsPageList creates the ranged weapons page list.
func NewRangedWeaponsPageList(entity *gurps.Entity) *PageList[*gurps.Weapon] {
	return newPageList(nil, NewWeaponsProvider(entity, false, true), colors.TintRanged, nil)
}

// newPageList creates a page list for the provider. A nil owner makes a read-only list. The tint and the kind-specific
// command handlers go on the list and on each of its continuations alike.
func newPageList[T gurps.Node[T]](owner Rebuildable, provider TableProvider[T], tint *unison.ThemeColor, installKindHandlers kindHandlerInstaller[T]) *PageList[T] {
	header, table := NewNodeTable(provider, fonts.PageFieldPrimary)
	table.ClientData()[WorkingDirKey] = WorkingDirProvider(owner)
	if xreflect.IsNil(owner) {
		owner = nil
	} else {
		table.ClientData()[TableOwnerClientKey] = owner
	}
	table.RefKey = provider.RefKey()
	p := &PageList[T]{
		tableHeader:         header,
		Table:               table,
		provider:            provider,
		owner:               owner,
		tint:                tint,
		installKindHandlers: installKindHandlers,
	}
	p.Self = p
	p.Table.PreventUserColumnResize = true
	p.Table.ShowLastColumnDivider = false
	p.Table.SyncToModel()
	p.installOnPanel(p.AsPanel(), header, table)
	return p
}

// installOnPanel shapes the panel of the list or of one of its continuations around the given header and view of the
// list's table, and installs on the panel and the view what the list offers.
func (p *PageList[T]) installOnPanel(panel *unison.Panel, header *unison.TableHeader[*Node[T]], view *unison.Table[*Node[T]]) {
	panel.SetLayout(&unison.FlexLayout{Columns: 1})
	panel.SetBorder(unison.NewLineBorder(header.BackgroundInk, geom.Size{}, geom.NewUniformInsets(1), false))
	panel.AddChild(header)
	panel.AddChild(view)
	// The read-only lists -- the weapons and the conditional and reaction modifiers, which are derived from the rest
	// of the sheet -- have no owner and offer none of the editing commands.
	editable := p.owner != nil
	installStandardTableCmdHandlers(panel, view, p.provider, func() Rebuildable { return p.owner }, editable)
	if editable {
		InstallTableDropSupport(view, p.provider)
	}
	if p.installKindHandlers != nil {
		p.installKindHandlers(panel, view)
	}
	if p.tint != nil {
		InstallTintFunc(panel, p.tint)
	}
	panel.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
	})
}

// continuation returns the n-th continuation of the list, counting from 0, creating it and any before it that don't
// exist yet. A new one shows no rows, having an empty range, until whoever places it sets one with SetDrawRowRange, and
// records the block it shows in its ClientData under pageKey. Pagination may move a continuation from one page to
// another.
func (p *PageList[T]) continuation(n int) *pageListContinuation[T] {
	for len(p.continuations) <= n {
		p.continuations = append(p.continuations, p.newContinuation())
	}
	return p.continuations[n]
}

// continuationPanel is continuation for a caller that doesn't know the type of the list's rows.
func (p *PageList[T]) continuationPanel(n int) pagePanel {
	return p.continuation(n)
}

// newContinuation creates a continuation of the list: a panel shaped like the list, holding a new view of the list's
// table under a header of its own, and offering the same commands.
func (p *PageList[T]) newContinuation() *pageListContinuation[T] {
	view := p.Table.NewView(0, 0)
	c := &pageListContinuation[T]{
		tableHeader: newNodeTableHeader(view, p.tableHeader.ColumnHeaders...),
		Table:       view,
	}
	c.Self = c
	installNodeTableView(view, c.tableHeader, p.provider, fonts.PageFieldPrimary)
	p.installOnPanel(c.AsPanel(), c.tableHeader, view)
	return c
}

// trimContinuations removes the continuations of the list beyond the first used of them, taking each out of its parent
// and its view out of the list's table.
func (p *PageList[T]) trimContinuations(used int) {
	if used >= len(p.continuations) {
		return
	}
	for _, c := range p.continuations[used:] {
		c.RemoveFromParent()
		p.Table.RemoveView(c.Table)
	}
	clear(p.continuations[used:])
	p.continuations = p.continuations[:used]
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

// refKey returns the reference key of the list's table, which names the list whatever table shows it; see liveTable.
func (p *PageList[T]) refKey() string {
	if p == nil {
		return ""
	}
	return p.Table.RefKey
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

func installMoveToCarriedEquipmentHandler(target *unison.Panel, table *unison.Table[*Node[*gurps.Equipment]], owner Rebuildable) {
	if sheet, ok := owner.AsPanel().Self.(*Sheet); ok {
		// While the layout leaves the carried equipment list off the page, the list's table has no undo manager, so a
		// move would do nothing.
		target.InstallCmdHandlers(MoveToCarriedEquipmentItemID,
			func(_ any) bool { return table.HasSelection() && sheet.showsBlock(gurps.BlockEquipmentKey) },
			func(_ any) { moveSelectedEquipment(sheet, table, sheet.CarriedEquipment.Table) })
	}
}

func installMoveToOtherEquipmentHandler(target *unison.Panel, table *unison.Table[*Node[*gurps.Equipment]], owner Rebuildable) {
	if sheet, ok := owner.AsPanel().Self.(*Sheet); ok {
		// While the layout leaves the other equipment list off the page, its table has no undo manager, so a move would
		// do nothing.
		target.InstallCmdHandlers(MoveToOtherEquipmentItemID,
			func(_ any) bool { return table.HasSelection() && sheet.showsBlock(gurps.BlockOtherEquipmentKey) },
			func(_ any) { moveSelectedEquipment(sheet, table, sheet.OtherEquipment.Table) })
	}
}

func installSellEquipmentHandler(target *unison.Panel, table *unison.Table[*Node[*gurps.Equipment]], owner Rebuildable) {
	if sheet, ok := owner.AsPanel().Self.(*Sheet); ok {
		target.InstallCmdHandlers(SellEquipmentItemID,
			func(_ any) bool { return !table.IsFiltered() && len(selectedSaleRows(table)) != 0 },
			func(_ any) { sellSelectedEquipment(sheet, table) })
	}
}

// moveSelectedEquipment moves the rows selected in one of the sheet's two equipment lists to the other one, unless an
// editor open on one of them stays open. The copy into the destination and the deletion from the source are each told
// not to report the change, so that the sheet is rebuilt once for the move as a whole rather than once per half of it
// -- and so that both tables are still the ones on screen when the undo data is collected, since a rebuild can replace
// either list (see liveTable).
func moveSelectedEquipment(sheet *Sheet, from, to *unison.Table[*Node[*gurps.Equipment]]) {
	mgr := unison.UndoManagerFor(from)
	if mgr == nil || mgr != unison.UndoManagerFor(to) {
		return
	}
	rows := from.SelectedRows(true)
	// The editors are closed before the undo data is collected or anything is copied, since saving an editor's changes
	// as it closes alters the item and rebuilds the sheet, which can replace either table.
	list := selectionToDelete(from)
	if len(list) == 0 {
		return
	}
	from = liveTable(from)
	to = liveTable(to)
	before := NewTableDragUndoEditData(from, to)
	// Only carried equipment grants anything, so the move can take the grants of the equipment moved away, and the
	// lists they put rows in are then part of the edit; see grantRowListsOf.
	if rowsCarryGrants(rows) {
		before.Others = grantRowListsOf(sheet, "")
	}
	undo := &unison.UndoEdit[*TableDragUndoEditData[*gurps.Equipment]]{
		ID:       unison.NextUndoID(),
		EditName: i18n.Text("Move Equipment"),
		UndoFunc: func(e *unison.UndoEdit[*TableDragUndoEditData[*gurps.Equipment]]) { e.BeforeData.Apply() },
		RedoFunc: func(e *unison.UndoEdit[*TableDragUndoEditData[*gurps.Equipment]]) { e.AfterData.Apply() },
		AbsorbFunc: func(_ *unison.UndoEdit[*TableDragUndoEditData[*gurps.Equipment]], _ unison.Undoable) bool {
			return false
		},
		BeforeData: before,
	}
	copyRowsTo(to, rows, nil, false, false)
	deleteNodes(from, list, false, false)
	after := NewTableDragUndoEditData(from, to)
	if len(before.Others) != 0 {
		after.Others = grantRowListsOf(sheet, "")
	}
	undo.AfterData = after
	mgr.Add(undo)
	rebuildAsModified(sheet, true)
}

func installToggleDisabledHandler(target *unison.Panel, table *unison.Table[*Node[*gurps.Trait]], owner Rebuildable) {
	target.InstallCmdHandlers(ToggleStateItemID,
		func(_ any) bool { return canToggleDisabled(table) },
		func(_ any) { toggleDisabled(owner, table) })
}

func installToggleEquippedHandler(target *unison.Panel, table *unison.Table[*Node[*gurps.Equipment]], owner Rebuildable) {
	target.InstallCmdHandlers(ToggleStateItemID,
		func(_ any) bool { return canToggleEquipped(table) },
		func(_ any) { toggleEquipped(owner, table) })
}

func installPointsHandlers[T gurps.Node[T]](target *unison.Panel, table *unison.Table[*Node[T]], owner Rebuildable) {
	target.InstallCmdHandlers(IncrementItemID,
		func(_ any) bool { return canAdjustRawPoints(table, true) },
		func(_ any) { adjustRawPoints(owner, table, true) })
	target.InstallCmdHandlers(DecrementItemID,
		func(_ any) bool { return canAdjustRawPoints(table, false) },
		func(_ any) { adjustRawPoints(owner, table, false) })
}

func installTraitLevelHandlers(target *unison.Panel, table *unison.Table[*Node[*gurps.Trait]], owner Rebuildable) {
	target.InstallCmdHandlers(IncrementItemID,
		func(_ any) bool { return canAdjustTraitLevel(table, true) },
		func(_ any) { adjustTraitLevel(owner, table, true) })
	target.InstallCmdHandlers(DecrementItemID,
		func(_ any) bool { return canAdjustTraitLevel(table, false) },
		func(_ any) { adjustTraitLevel(owner, table, false) })
}

func installSkillLevelHandlers[T gurps.Node[T]](target *unison.Panel, table *unison.Table[*Node[T]], owner Rebuildable) {
	target.InstallCmdHandlers(IncrementSkillLevelItemID,
		func(_ any) bool { return canAdjustSkillLevel(table, true) },
		func(_ any) { adjustSkillLevel(owner, table, true) })
	target.InstallCmdHandlers(DecrementSkillLevelItemID,
		func(_ any) bool { return canAdjustSkillLevel(table, false) },
		func(_ any) { adjustSkillLevel(owner, table, false) })
}

func installTechLevelHandlers[T gurps.Node[T]](target *unison.Panel, table *unison.Table[*Node[T]], owner Rebuildable) {
	target.InstallCmdHandlers(IncrementTechLevelItemID,
		func(_ any) bool { return canAdjustTechLevel(table, fxp.One) },
		func(_ any) { adjustTechLevel(owner, table, fxp.One) })
	target.InstallCmdHandlers(DecrementTechLevelItemID,
		func(_ any) bool { return canAdjustTechLevel(table, -fxp.One) },
		func(_ any) { adjustTechLevel(owner, table, -fxp.One) })
}

func installEquipmentLevelHandlers(target *unison.Panel, table *unison.Table[*Node[*gurps.Equipment]], owner Rebuildable) {
	target.InstallCmdHandlers(IncrementEquipmentLevelItemID,
		func(_ any) bool { return canAdjustEquipmentLevel(table, fxp.One) },
		func(_ any) { adjustEquipmentLevel(owner, table, fxp.One) })
	target.InstallCmdHandlers(DecrementEquipmentLevelItemID,
		func(_ any) bool { return canAdjustEquipmentLevel(table, -fxp.One) },
		func(_ any) { adjustEquipmentLevel(owner, table, -fxp.One) })
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
	// Size the columns here, since a settings change, such as to page reference column expansion, won't necessarily
	// alter the table's frame or the heights of its rows and thus won't have the table's FrameChangeCallback size them.
	sizePageTableColumns(p.Table, p.provider.ExcessWidthColumnID())
	p.Table.NeedsLayout = true
	p.NeedsLayout = true
	if parent := p.Parent(); parent != nil {
		parent.NeedsLayout = true
	}
	p.syncedAt = currentDisplayStamp()
}

// syncedAtStamp implements syncedList.
func (p *PageList[T]) syncedAtStamp() uint64 {
	return p.syncedAt
}

// CreateItem calls CreateItem on the contained TableProvider.
func (p *PageList[T]) CreateItem(owner Rebuildable, variant ItemVariant) {
	p.provider.CreateItem(owner, p.Table, variant)
}

// OverheadHeight returns the overhead for this page list, i.e. the border and header space.
func (p *PageList[T]) OverheadHeight() float32 {
	return pageListOverheadHeight(p.AsPanel(), p.tableHeader)
}

// MinimumHeight implements pageHelper.
func (p *PageList[T]) MinimumHeight() float32 {
	return pageListMinimumHeight(p.AsPanel(), p.tableHeader, p.Table)
}

// pageListMinimumHeight returns the height the panel of a page list, or of one of its continuations, stands at however
// few rows it draws: its overhead and the minimum height its view of the table is laid out at.
func pageListMinimumHeight[T gurps.Node[T]](panel *unison.Panel, header *unison.TableHeader[*Node[T]], view *unison.Table[*Node[T]]) float32 {
	height := pageListOverheadHeight(panel, header)
	if data, ok := view.LayoutData().(*unison.FlexLayoutData); ok {
		height += data.MinSize.Height
	}
	return height
}

// pageListOverheadHeight returns the height the panel of a page list, or of one of its continuations, takes beyond the
// rows it shows: its border and its header.
func pageListOverheadHeight[T gurps.Node[T]](panel *unison.Panel, header *unison.TableHeader[*Node[T]]) float32 {
	_, pref, _ := header.Sizes(geom.Size{})
	if border := panel.Border(); border != nil {
		pref.Height += border.Insets().Height()
	}
	return pref.Height
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

// ClearDrawRowRange removes the row range set by SetDrawRowRange, so that the list shows every row again.
func (p *PageList[T]) ClearDrawRowRange() {
	p.Table.ClearDrawRowRange()
}

// pagePanel is a panel showing a range of the rows of a page list: the list itself or one of its continuations.
type pagePanel interface {
	unison.Paneler
	pageHelper
}

// pageListContinuation shows rows of a page list that go on past the page the list is on: a panel shaped like the list,
// holding a view of the list's table under a header of its own (see PageList.continuation).
type pageListContinuation[T gurps.Node[T]] struct {
	unison.Panel
	tableHeader *unison.TableHeader[*Node[T]]
	Table       *unison.Table[*Node[T]] // The view of the list's table
}

// OverheadHeight implements pageHelper.
func (c *pageListContinuation[T]) OverheadHeight() float32 {
	return pageListOverheadHeight(c.AsPanel(), c.tableHeader)
}

// MinimumHeight implements pageHelper.
func (c *pageListContinuation[T]) MinimumHeight() float32 {
	return pageListMinimumHeight(c.AsPanel(), c.tableHeader, c.Table)
}

// RowHeights implements pageHelper.
func (c *pageListContinuation[T]) RowHeights() []float32 {
	return c.Table.RowHeights()
}

// CurrentDrawRowRange implements pageHelper.
func (c *pageListContinuation[T]) CurrentDrawRowRange() (start, endBefore int) {
	return c.Table.CurrentDrawRowRange()
}

// SetDrawRowRange implements pageHelper.
func (c *pageListContinuation[T]) SetDrawRowRange(start, endBefore int) {
	c.Table.SetDrawRowRange(start, endBefore)
}

// ClearDrawRowRange implements pageHelper.
func (c *pageListContinuation[T]) ClearDrawRowRange() {
	c.Table.ClearDrawRowRange()
}
