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
	"hash"
	"maps"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
)

var (
	_ FileBackedDockable           = &TableDockable[*gurps.Trait]{}
	_ unison.UndoManagerProvider   = &TableDockable[*gurps.Trait]{}
	_ ModifiableRoot               = &TableDockable[*gurps.Trait]{}
	_ Rebuildable                  = &TableDockable[*gurps.Trait]{}
	_ unison.TabCloser             = &TableDockable[*gurps.Trait]{}
	_ KeyedDockable                = &TableDockable[*gurps.Trait]{}
	_ gurps.Hashable               = &TableDockable[*gurps.Trait]{}
	_ listFilterObserver           = &TableDockable[*gurps.Trait]{}
	_ gurps.SheetSettingsResponder = &TableDockable[*gurps.Trait]{}
	_ readOnlyView                 = &TableDockable[*gurps.Trait]{}
	_ openStateSharer              = &TableDockable[*gurps.Trait]{}
)

// TableDockable holds the view for a file that contains a (potentially hierarchical) list of data.
type TableDockable[T gurps.Node[T]] struct {
	fileBackedPanel
	listFilterControls[T]
	undoMgr         *unison.UndoManager
	filterBanner    *filterBanner
	scroll          *unison.ScrollPanel
	tableHeader     *unison.TableHeader[*Node[T]]
	lockButton      *unison.Button
	lockButtonGroup *unison.Group
	scaleField      *PercentageField
	scale           int
	locked          bool
	// lockedUndoMgr records the scale changes made while the list is locked, and scaleAtLock is the scale field's state
	// when it was locked (see applyLocked).
	lockedUndoMgr *unison.UndoManager
	scaleAtLock   *unison.FieldState
}

// NewTableDockable creates a new TableDockable for list data files.
func NewTableDockable[T gurps.Node[T]](filePath, extension string, provider TableProvider[T], saver func(path string) error, canCreateIDs ...int) *TableDockable[T] {
	header, table := NewNodeTable(provider, nil)
	d := &TableDockable[T]{
		undoMgr:      unison.NewUndoManager(200, func(err error) { errs.Log(err) }),
		filterBanner: newFilterBanner(),
		scroll:       unison.NewScrollPanel(),
		tableHeader:  header,
		scale:        gurps.GlobalSettings().General.InitialListUIScale,
	}
	d.initListFilterControls(table, provider, nil, func(filtered bool, _ int) { d.showFilterBanner(filtered) })
	d.Self = d
	d.initFileEditor(d, filePath, extension, saver, d)
	d.SetLayout(&unison.FlexLayout{Columns: 1})

	d.table.SyncToModel()
	d.table.SizeColumnsToFit(true)
	applySavedColumnWidths(d.table, filePath)

	InstallTableDropSupport(d.table, d.provider)

	d.scroll.SetColumnHeader(d.tableHeader)
	d.scroll.SetContent(d.table, behavior.Fill, behavior.Fill)
	d.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})

	d.AddChild(d.createToolbar())
	d.AddChild(d.scroll)

	installStandardTableCmdHandlers(d, d.table, d.provider, func() Rebuildable { return d }, true)
	// Installed here rather than in NewNodeTable, since only library lists can be a source: modifiers applied from the
	// list inside a trait or equipment editor to the item being edited would be lost when the editor applies its
	// changes, which replace the item's modifiers wholesale.
	switch t := any(d.table).(type) {
	case *unison.Table[*Node[*gurps.TraitModifier]]:
		installApplyModifierHandler(d, t, traitModifierTargetKind())
		installModifierChoiceConversionHandlers[*gurps.TraitModifier, *gurps.TraitModifierEditData](d, t)
	case *unison.Table[*Node[*gurps.EquipmentModifier]]:
		installApplyModifierHandler(d, t, equipmentModifierTargetKind())
		installModifierChoiceConversionHandlers[*gurps.EquipmentModifier, *gurps.EquipmentModifierEditData](d, t)
	}
	d.InstallCmdHandlers(SaveItemID,
		func(_ any) bool { return d.Modified() },
		func(_ any) { d.save(false) })
	d.InstallCmdHandlers(SaveAsItemID, unison.AlwaysEnabled, func(_ any) { d.save(true) })
	d.InstallCmdHandlers(JumpToSearchFilterItemID,
		func(any) bool { return !d.filterField.Focused() },
		func(any) { d.filterField.RequestFocus() })
	for _, id := range canCreateIDs {
		variant := ItemVariant(-1)
		switch {
		case id > FirstNonContainerMarker && id < LastNonContainerMarker:
			variant = NoItemVariant
		case id > FirstContainerMarker && id < LastContainerMarker:
			variant = ContainerItemVariant
		case id > FirstGroupContainerMarker && id < LastGroupContainerMarker:
			variant = GroupContainerItemVariant
		case id > FirstModifierChoiceMarker && id < LastModifierChoiceMarker:
			variant = ChoiceContainerItemVariant
		case id > FirstAlternateNonContainerMarker && id < LastAlternateNonContainerMarker:
			variant = AlternateItemVariant
		}
		if variant != -1 {
			// Creating an item inserts a row, which unison.Table.ApplyHierarchicalFilter forbids while a filter is
			// applied, so the command is disabled then, and while the list is locked, as Delete, Duplicate and the move
			// commands are. Left on under a filter, the insert would clear the filter, open an editor on the new row
			// and then have the rebuild that follows re-filter the row out from under that editor.
			d.InstallCmdHandlers(id,
				func(_ any) bool { return rowsChangeable(d.table) },
				func(_ any) { d.provider.CreateItem(d, d.table, variant) })
		}
	}
	return d
}

func (d *TableDockable[T]) createToolbar() *unison.Panel {
	sizeToFitButton := unison.NewSVGButton(svg.SizeToFit)
	sizeToFitButton.Tooltip = newWrappedTooltip(i18n.Text("Sets the width of each column to fit its contents"))
	sizeToFitButton.ClickCallback = d.sizeToFit

	toolbar := NewToolbar()
	toolbar.AddChild(NewDefaultInfoPop())
	d.scaleField = AddUIScaleField(toolbar, func() int { return gurps.GlobalSettings().General.InitialListUIScale },
		func() int { return d.scale }, func(scale int) { d.scale = scale }, false, d.scroll)
	d.lockButton, d.lockButtonGroup = newLatchingButton(svg.Lock, i18n.Text("Lock against changes"), d.toggleLocked)
	setLatched(d.lockButton, d.lockButtonGroup, d.locked)
	toolbar.AddChild(d.lockButton)
	d.addListFilterControls(toolbar, sizeToFitButton)
	d.finishListFilterToolbar(toolbar, nil, nil)
	return toolbar
}

// Entity implements EntityPanel. A list file has no entity, so nil is always returned.
func (d *TableDockable[T]) Entity() *gurps.Entity {
	return nil
}

// UndoManager implements unison.UndoManagerProvider. A locked list offers one for the scale changes made while it is
// locked, which keeps the changes to its rows out of reach of Undo and Redo and their history for when it is unlocked.
func (d *TableDockable[T]) UndoManager() *unison.UndoManager {
	if d.locked {
		return d.lockedUndoMgr
	}
	return d.undoMgr
}

// markLoadedFromFile implements fileLoadable. A list opened from a file starts locked when the settings call for it,
// while a new list, which is never marked as loaded, starts unlocked.
func (d *TableDockable[T]) markLoadedFromFile() {
	d.fileBackedPanel.markLoadedFromFile()
	if gurps.GlobalSettings().General.InitialListLocked {
		d.applyLocked(true)
	}
}

// isReadOnly implements readOnlyView.
func (d *TableDockable[T]) isReadOnly() bool {
	return d.locked
}

func (d *TableDockable[T]) toggleLocked() {
	d.setLocked(!d.locked)
}

// setLocked locks or unlocks the list, once the editors opened from it have closed: when locking, so that none can
// apply changes later, and when unlocking, so that the read-only ones aren't reused for their rows. Canceling the
// close of one leaves the lock as it was, though the editors closed before it with Save have applied their changes.
func (d *TableDockable[T]) setLocked(locked bool) {
	if locked != d.locked && CloseGroup(d) {
		d.applyLocked(locked)
		return
	}
	// A click has already latched the button, so it has to show the state that stays in force.
	setLatched(d.lockButton, d.lockButtonGroup, d.locked)
}

// applyLocked puts the lock state in force. Sorting reorders the file's contents, so it is turned off while locked,
// which stops the re-sort that follows clearing a filter as well as header clicks. The sort order is kept for when it
// is turned back on. The scale isn't part of the file, so its changes stay undoable while locked, in a history of
// their own; on unlocking, their net change joins the list's history as a single edit.
func (d *TableDockable[T]) applyLocked(locked bool) {
	if locked == d.locked {
		return
	}
	d.locked = locked
	if locked {
		d.lockedUndoMgr = unison.NewUndoManager(200, func(err error) { errs.Log(err) })
		d.scaleAtLock = d.scaleField.GetFieldState()
	} else {
		d.lockedUndoMgr = nil
		if d.scaleField.parse(d.scaleAtLock.Text) != d.scale {
			recordTargetUndo(d.scaleField, nil, "", d.scaleField.undoTitle, unison.NextUndoID(), d.scaleAtLock,
				d.scaleField.GetFieldState(),
				func(self *PercentageField, state *unison.FieldState) { self.setWithoutUndo(state, true) })
		}
		d.scaleAtLock = nil
	}
	setColumnsSortable(d.tableHeader.ColumnHeaders, !locked)
	// The column headers aren't children of the table header, which draws them, so their own redraws reach nothing.
	d.tableHeader.MarkForRedraw()
	setLatched(d.lockButton, d.lockButtonGroup, locked)
}

// MarkModified implements ModifiableRoot.
func (d *TableDockable[T]) MarkModified(_ unison.Paneler) {
	invalidateDisplayData()
	UpdateTitleForDockable(d)
}

// AttemptClose implements unison.TabCloser. The column widths are preserved as the table goes away, so that the file
// can be opened again with the same ones.
func (d *TableDockable[T]) AttemptClose() bool {
	if AttemptSaveForDockable(d) {
		saveColumnWidths(d.table, d.BackingFilePath())
		return AttemptCloseForDockable(d)
	}
	return false
}

func (d *TableDockable[T]) sizeToFit() {
	d.table.SizeColumnsToFit(true)
	d.table.MarkForRedraw()
}

// Rebuild implements Rebuildable.
func (d *TableDockable[T]) Rebuild(_ bool) {
	d.rebuildKeepingView(d.scroll)
	UpdateTitleForDockable(d)
}

// SheetSettingsUpdated implements gurps.SheetSettingsResponder.
func (d *TableDockable[T]) SheetSettingsUpdated(entity *gurps.Entity, parts gurps.SheetSettingsParts,
	fullRebuild bool,
) {
	listSheetSettingsUpdated(d.table, d.Rebuild, entity, parts, fullRebuild)
}

// listSheetSettingsUpdated brings a library list's table in line with a change to the sheet settings, calling rebuild
// when its rows have to be built anew. A library list has no entity of its own and shows what the default sheet
// settings call for, such as the default attribute names, so only a change to those -- reported with a nil entity --
// concerns it, and a change to the default block layout or body type, neither of which it has any part in, does not. A
// change that can't alter the columns or the rows, reported without a full rebuild, leaves the rows in place: each cell
// derives its data afresh once the change has been reported and keeps its panel when that data is unchanged (see
// Node.ColumnCell), so only the row heights have to be settled again. A library list builds its rows anew on every
// sync (see rowReuse), so syncing it for such a change would build every cell anew, which a keystroke in the global
// sheet settings would then cost for every open list. A filter in force is applied over again, since what it keeps can
// turn on the settings, as a weight does on the default units.
func listSheetSettingsUpdated[T gurps.Node[T]](table *unison.Table[*Node[T]], rebuild func(full bool),
	entity *gurps.Entity, parts gurps.SheetSettingsParts, fullRebuild bool,
) {
	if entity != nil || parts&(gurps.SheetSettingsGeneralPart|gurps.SheetSettingsAttributesPart) == 0 {
		return
	}
	if fullRebuild || table.IsFiltered() {
		rebuild(fullRebuild)
		return
	}
	invalidateDisplayData()
	gurps.DiscardGlobalResolveCache()
	table.SyncRowHeights()
	_, pref, _ := table.Sizes(geom.Size{})
	rect := table.FrameRect()
	rect.Size = pref
	table.SetFrameRect(rect)
	table.MarkForLayoutRecursivelyUpward()
	table.MarkForRedraw()
}

// applySavedColumnWidths gives the table's columns the widths saved for the key by saveColumnWidths, syncing the table
// when that changes any of them, and reports whether there were saved widths.
func applySavedColumnWidths[T gurps.Node[T]](table *unison.Table[*Node[T]], key string) bool {
	columnSizing, ok := gurps.GlobalSettings().ColumnSizing[key]
	if !ok {
		return false
	}
	needSync := false
	for id, width := range columnSizing {
		if id != -1 {
			if i := table.ColumnIndexForID(id); i != -1 {
				if table.Columns[i].Current != width {
					table.Columns[i].Current = width
					needSync = true
				}
			}
		}
	}
	if needSync {
		table.SyncToModel()
	}
	return true
}

// saveColumnWidths saves the widths of the table's columns under the key, along with the time they were saved, which
// takes the slot of column ID -1.
func saveColumnWidths[T gurps.Node[T]](table *unison.Table[*Node[T]], key string) {
	m := make(map[int]float32, len(table.Columns))
	m[-1] = gurps.ToColumnCutoff(time.Now().Unix())
	for _, col := range table.Columns {
		m[col.ID] = col.Current
	}
	gurps.GlobalSettings().ColumnSizing[key] = m
}

// Hash writes this object's contents into the hasher.
func (d *TableDockable[T]) Hash(h hash.Hash) {
	rootData := d.provider.RootData()
	data := make([]any, 0, len(rootData))
	for _, one := range rootData {
		data = append(data, one)
	}
	gurps.HashJSON(h, data)
}

// selectRowsByID selects exactly the rows with the given IDs that the list holds, opening the containers above them,
// and scrolls the first of them into view, returning how many it found. A filter that hides any of them is cleared
// first (the quick filter's text, the chosen tags and the saved filter).
func (d *TableDockable[T]) selectRowsByID(ids map[tid.TID]bool) int {
	sel := make(map[tid.TID]bool)
	opened := false
	gurps.Traverse(func(one T) bool {
		if id := one.ID(); ids[id] {
			sel[id] = true
			for parent := one.Parent(); !xreflect.IsNil(parent); parent = parent.Parent() {
				if !parent.IsOpen() {
					// The open state is kept in the settings rather than in the list, so this doesn't modify the list.
					parent.SetOpen(true)
					openStateChanged(d.table, parent.ID())
					opened = true
				}
			}
		}
		return false
	}, false, false, d.provider.RootData()...)
	if len(sel) == 0 {
		d.table.ClearSelection()
		return 0
	}
	switch {
	case d.table.IsFiltered():
		// A filter shows every container it keeps as open, so the containers just opened change nothing it shows. Its
		// context rows count as shown.
		hidden := maps.Clone(sel)
		for i := range d.table.LastRowIndex() + 1 {
			delete(hidden, d.table.RowFromIndex(i).ID())
		}
		if len(hidden) != 0 {
			d.clearFilters() // Clearing the filter syncs the table.
		}
	case opened:
		d.table.SyncToModel()
	}
	d.table.SetSelectionMap(sel)
	if row := d.table.FirstSelectedRowIndex(); row != -1 {
		// Opening containers only invalidated the layout, so without this the scroll range would still end where the
		// table used to.
		d.table.ValidateScrollRoot()
		d.table.RequestFocus()
		d.table.ScrollRowIntoView(row)
	}
	return len(sel)
}

// showFilterBanner puts the banner between the toolbar and the table, or takes it away again, and has the dockable
// laid out afresh when that changes anything. The banner is added and removed rather than hidden, since a hidden panel
// still takes up its space in a FlexLayout.
func (d *TableDockable[T]) showFilterBanner(show bool) {
	if show == (d.filterBanner.Parent() != nil) {
		return
	}
	if show {
		d.AddChildAtIndex(d.filterBanner, d.IndexOfChild(d.scroll))
	} else {
		d.filterBanner.RemoveFromParent()
	}
	d.MarkForLayoutAndRedraw()
}
