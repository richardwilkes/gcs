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
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/mod"
)

const containerMarker = "\000"

// TableOwnerClientKey is the key used to store a table's Rebuildable owner with the table. It is stored on the table
// itself, rather than looked up through the panel hierarchy, so that it can still be found after the owner has
// replaced the table with a new one.
const TableOwnerClientKey = "table-owner"

// ItemVariant holds the type of item variant to create.
type ItemVariant int

// Possible values for ItemVariant.
const (
	NoItemVariant ItemVariant = iota
	ContainerItemVariant
	AlternateItemVariant
)

// TableProvider defines the methods a table provider must contain.
type TableProvider[T gurps.Node[T]] interface {
	unison.TableModel[*Node[T]]
	gurps.DataOwnerProvider
	SetTable(table *unison.Table[*Node[T]])
	RootData() []T
	SetRootData(data []T)
	DragKey() *uti.DataType
	DragSVG() *unison.SVG
	DropShouldMoveData(from, to *unison.Table[*Node[T]]) bool
	ProcessDropData(from, to *unison.Table[*Node[T]])
	AltDropSupport() *AltDropSupport
	ItemNames() (singular, plural string)
	Headers() []unison.TableColumnHeader[*Node[T]]
	SyncHeader(headers []unison.TableColumnHeader[*Node[T]])
	ColumnIDs() []int
	HierarchyColumnID() int
	ExcessWidthColumnID() int
	ContextMenuItems() []ContextMenuItem
	OpenEditor(owner Rebuildable, table *unison.Table[*Node[T]])
	CreateItem(owner Rebuildable, table *unison.Table[*Node[T]], variant ItemVariant)
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
	RefKey() string
	AllTags() []string
}

// NewNodeTable creates a new node table of the specified type, returning the header and table. Pass nil for 'font' if
// this should be a standalone top-level table for a dockable. Otherwise, pass in the typical font used for a cell.
func NewNodeTable[T gurps.Node[T]](provider TableProvider[T], font unison.Font) (header *unison.TableHeader[*Node[T]], table *unison.Table[*Node[T]]) {
	table = unison.NewTable(provider)
	table.ShowFirstColumnDivider = false
	provider.SetTable(table)
	table.HierarchyColumnID = provider.HierarchyColumnID()
	layoutData := &unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	}
	if font != nil {
		table.Padding.Top = 0
		table.Padding.Bottom = 0
		table.HierarchyIndent = font.LineHeight()
		table.MinimumRowHeight = font.LineHeight()
		layoutData.MinSize = geom.Size{Height: 4 + fonts.PageFieldPrimary.LineHeight()}
	}
	table.SetLayoutData(layoutData)

	ids := provider.ColumnIDs()
	headers := provider.Headers()
	table.Columns = make([]unison.ColumnInfo, len(headers))
	for i := range table.Columns {
		_, pref, _ := headers[i].AsPanel().Sizes(geom.Size{})
		pref.Width += table.Padding.Left + table.Padding.Right
		table.Columns[i].ID = ids[i]
		table.Columns[i].AutoMinimum = pref.Width
		table.Columns[i].AutoMaximum = max(float32(gurps.GlobalSettings().General.MaximumAutoColWidth), pref.Width)
		table.Columns[i].Minimum = pref.Width
		table.Columns[i].Maximum = 10000
	}
	header = unison.NewTableHeader(table, headers...)
	header.Less = flexibleLess
	header.BackgroundInk = colors.Header
	header.InteriorDividerColor = colors.Header
	header.SetBorder(header.HeaderBorder)
	header.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
	})

	table.DoubleClickCallback = func() { table.PerformCmd(nil, OpenEditorItemID) }
	table.KeyDownCallback = func(keyCode unison.KeyCode, mods mod.Modifiers, repeat bool) bool {
		if noModifiersDown(mods) && (keyCode == unison.KeyBackspace || keyCode == unison.KeyDelete) {
			table.PerformCmd(table, unison.DeleteItemID)
			return true
		}
		return table.DefaultKeyDown(keyCode, mods, repeat)
	}
	singular, plural := provider.ItemNames()
	table.InstallDragSupport(provider.DragSVG(), provider.DragKey(), singular, plural)
	// Mirror the dragged rows into our own storage, since unison only retains them internally, so that alternate drop
	// handlers -- which deal with a different row type than the destination table -- can reach them.
	origMouseDrag := table.MouseDragCallback
	table.MouseDragCallback = func(where geom.Point, button int, mods mod.Modifiers) bool {
		if button == unison.ButtonLeft && table.HasSelection() && table.IsDragGesture(where) {
			draggedTableData = &unison.TableDragData[*Node[T]]{Table: table, Rows: table.SelectedRows(true)}
		}
		return origMouseDrag(where, button, mods)
	}
	if font != nil {
		table.FrameChangeCallback = func() {
			sizePageTableColumns(table, provider.ExcessWidthColumnID())
		}
	}

	table.MouseDownCallback = func(where geom.Point, button, clickCount int, mods mod.Modifiers) bool {
		stop := table.DefaultMouseDown(where, button, clickCount, mods)
		if button == unison.ButtonRight && clickCount == 1 {
			f := unison.DefaultMenuFactory()
			cm := f.NewMenu(unison.PopupMenuTemporaryBaseID|unison.ContextMenuIDFlag, "", nil)
			id := 1
			for _, one := range provider.ContextMenuItems() {
				if one.ID == -1 {
					cm.InsertSeparator(-1, true)
				} else {
					InsertCmdContextMenuItem(table, one.Title, one.ID, &id, cm)
				}
			}
			count := cm.Count()
			if count > 0 {
				count--
				if cm.ItemAtIndex(count).IsSeparator() {
					cm.RemoveItem(count)
				}
				table.FlushDrawing()
				cm.Popup(geom.Rect{
					Point:  table.PointToRoot(where),
					Width:  1,
					Height: 1,
				}, 0)
			}
			cm.Dispose()
		}
		return stop
	}

	table.InstallCmdHandlers(CopyToSheetItemID,
		func(_ any) bool { return canCopySelectionTo(table, OpenSheets(table.Ancestor[*Sheet]())) },
		func(_ any) { copySelectionTo(table, OpenSheets(table.Ancestor[*Sheet]())) })
	table.InstallCmdHandlers(CopyToTemplateItemID,
		func(_ any) bool { return canCopySelectionTo(table, OpenTemplates(table.Ancestor[*Template]())) },
		func(_ any) { copySelectionTo(table, OpenTemplates(table.Ancestor[*Template]())) })
	if t, ok := any(table).(*unison.Table[*Node[*gurps.Equipment]]); ok {
		t.InstallCmdHandlers(IncrementItemID,
			func(_ any) bool { return canAdjustQuantity(t, true) },
			func(_ any) { adjustQuantity(t.AncestorOrSelf[Rebuildable](), t, true) })
		t.InstallCmdHandlers(DecrementItemID,
			func(_ any) bool { return canAdjustQuantity(t, false) },
			func(_ any) { adjustQuantity(t.AncestorOrSelf[Rebuildable](), t, false) })
		t.InstallCmdHandlers(IncrementUsesItemID,
			func(_ any) bool { return canAdjustUses(t, 1) },
			func(_ any) { adjustUses(t.AncestorOrSelf[Rebuildable](), t, 1) })
		t.InstallCmdHandlers(DecrementUsesItemID,
			func(_ any) bool { return canAdjustUses(t, -1) },
			func(_ any) { adjustUses(t.AncestorOrSelf[Rebuildable](), t, -1) })
		t.InstallCmdHandlers(ResetUsesToMaxItemID,
			func(_ any) bool { return canResetUsesToMax(t) },
			func(_ any) { resetUsesToMax(t.AncestorOrSelf[Rebuildable](), t) })
	}

	return header, table
}

// installStandardTableCmdHandlers installs the commands every node table offers -- opening the editor, following the
// page references, and, when editable, deleting, duplicating, syncing with and clearing the source of the selection --
// on the panel that owns the table rather than on the table itself, so that they are reachable from anywhere within
// that panel: in a list dockable that means the filter field too, keeping the commands active on the selection while
// the user types in the filter. The owner is resolved through a function rather than taken up front because the
// editor's list panels are built before they are attached to the editor that owns them.
func installStandardTableCmdHandlers[T gurps.Node[T]](target unison.Paneler, table *unison.Table[*Node[T]], provider TableProvider[T], owner func() Rebuildable, editable bool) {
	panel := target.AsPanel()
	panel.InstallCmdHandlers(OpenEditorItemID,
		func(_ any) bool { return table.HasSelection() },
		func(_ any) { provider.OpenEditor(owner(), table) })
	panel.InstallCmdHandlers(OpenOnePageReferenceItemID,
		func(_ any) bool { return CanOpenPageRef(table) },
		func(_ any) { OpenPageRef(table) })
	panel.InstallCmdHandlers(OpenEachPageReferenceItemID,
		func(_ any) bool { return CanOpenPageRef(table) },
		func(_ any) { OpenEachPageRef(table) })
	if !editable {
		return
	}
	panel.InstallCmdHandlers(unison.DeleteItemID,
		func(_ any) bool { return HasSelectionAndNotFiltered(table) },
		func(_ any) { DeleteSelection(table, true) })
	panel.InstallCmdHandlers(DuplicateItemID,
		func(_ any) bool { return HasSelectionAndNotFiltered(table) },
		func(_ any) { DuplicateSelection(table) })
	panel.InstallCmdHandlers(SyncWithSourceItemID,
		func(_ any) bool { return HasSelectionAndNotFiltered(table) },
		func(_ any) { SyncWithSourceForSelection(table) })
	panel.InstallCmdHandlers(ClearSourceItemID,
		func(_ any) bool { return HasSelectionAndNotFiltered(table) },
		func(_ any) { ClearSourceFromSelection(table) })
}

// sizePageTableColumns sizes the columns of a fixed-width page table to fit, then, when the user can't resize columns
// and the setting is enabled, lets any page reference columns claim leftover space so they can show more than one
// reference. Run both when the table's frame changes and when it is synced, since a settings change won't necessarily
// alter the frame.
func sizePageTableColumns[T gurps.Node[T]](table *unison.Table[*Node[T]], excessColumnID int) {
	table.SizeColumnsToFitWithExcessIn(excessColumnID)
	if table.PreventUserColumnResize && gurps.GlobalSettings().General.ExpandPageReferences &&
		expandPageRefColumns(table, excessColumnID) {
		table.SyncRowHeights()
	}
}

// expandPageRefColumns hands leftover horizontal space to any page reference columns so they can display more than a
// single reference, taking that space from the excess-width column. The excess column keeps enough room to show its
// primary content with notes collapsed; only the space beyond that is offered, and each page reference column grows
// only up to the width needed to show all of its references (capped by its AutoMaximum). Returns true if any column
// width was changed.
func expandPageRefColumns[T gurps.Node[T]](table *unison.Table[*Node[T]], excessColumnID int) bool {
	excess := table.ColumnIndexForID(excessColumnID)
	if excess < 0 || excess >= len(table.Columns) {
		return false
	}
	lastRow := table.LastRowIndex()
	if lastRow < 0 {
		return false
	}
	// Reserve enough room for the excess column to show its primary content with notes collapsed. Computed lazily,
	// since it isn't needed unless a page reference column actually has room to grow.
	floor := float32(-1)
	excessFloor := func() float32 {
		if floor < 0 {
			floor = table.Columns[excess].Minimum
			for row := 0; row <= lastRow; row++ {
				floor = max(floor, table.RowFromIndex(row).excessColumnCollapsedWidth(excess))
			}
		}
		return floor
	}
	changed := false
	for col := range table.Columns {
		if col == excess {
			continue
		}
		// A negative width means this isn't a page reference column, so skip it.
		content := float32(-1)
		isPageRef := true
		for row := 0; row <= lastRow; row++ {
			w := table.RowFromIndex(row).pageRefColumnFullWidth(col)
			if w < 0 {
				isPageRef = false
				break
			}
			content = max(content, w)
		}
		if !isPageRef {
			continue
		}
		target := content + table.Padding.Left + table.Padding.Right
		if m := table.Columns[col].AutoMaximum; m > 0 && target > m {
			target = m
		}
		want := target - table.Columns[col].Current
		if want <= 0 {
			continue
		}
		available := table.Columns[excess].Current - excessFloor()
		if give := min(want, available); give > 0 {
			table.Columns[col].Current += give
			table.Columns[excess].Current -= give
			changed = true
		}
	}
	return changed
}

// copyDestination is what copySelectionTo asks of a sheet or template it copies rows onto: the page list for a block
// key, so the type of the rows being copied can pick the list they land in without naming the destination's fields.
// Both *Sheet and *Template satisfy it.
type copyDestination interface {
	FileBackedDockable
	list(key string) sheetList
}

// blockKeyForRow returns the key of the block that holds rows of the given type on a sheet or template, or "" when rows
// of that type can't be copied onto one. Equipment goes to the carried list; a sheet's other-equipment list only takes
// rows by drag and drop or by the move commands.
func blockKeyForRow(data any) string {
	switch data.(type) {
	case *gurps.Trait:
		return gurps.BlockTraitsKey
	case *gurps.Skill:
		return gurps.BlockSkillsKey
	case *gurps.Spell:
		return gurps.BlockSpellsKey
	case *gurps.Equipment:
		return gurps.BlockEquipmentKey
	case *gurps.Note:
		return gurps.BlockNotesKey
	default:
		return ""
	}
}

// canCopySelectionTo returns true if the table has a selection whose rows can be copied onto a sheet or template and
// there is at least one destination to copy them to.
func canCopySelectionTo[T gurps.Node[T], D copyDestination](table *unison.Table[*Node[T]], destinations []D) bool {
	var t T
	return table.HasSelection() && len(destinations) > 0 && blockKeyForRow(t) != ""
}

func libraryFileFromTable[T gurps.Node[T]](table *unison.Table[*Node[T]]) gurps.LibraryFile {
	if d := table.Ancestor[*TableDockable[T]](); d != nil {
		for _, lib := range gurps.GlobalSettings().Libraries.List() {
			libPathOnDisk := lib.Data().PathOnDisk + string(filepath.Separator)
			filePathOnDisk := d.BackingFilePath()
			if after, ok := strings.CutPrefix(filePathOnDisk, libPathOnDisk); ok {
				return gurps.LibraryFile{
					Library: lib.Key(),
					Path:    filepath.ToSlash(after),
				}
			}
		}
	}
	return gurps.LibraryFile{}
}

// copySelectionTo copies the table's selected rows onto each of the destinations the user picks from those given (see
// PromptForDestination), landing them in the destination's list for the rows' type and resolving them the way a drop
// onto a sheet would (see processCopiedRows).
func copySelectionTo[T gurps.Node[T], D copyDestination](table *unison.Table[*Node[T]], destinations []D) {
	if !table.HasSelection() {
		return
	}
	destinations = PromptForDestination(destinations)
	if len(destinations) == 0 {
		return
	}
	sel := table.SelectedRows(true)
	key := blockKeyForRow(sel[0].Data())
	for _, d := range destinations {
		// The assertion fails for a key that isn't a block key of this destination, since its list then comes back as
		// an untyped nil, and for a destination that hasn't built the list yet, whose list is a typed nil.
		target, ok := d.list(key).(*PageList[T])
		if !ok || target == nil {
			continue
		}
		// All processing must happen inside the postProcessor so it is captured by the undo edit's after-state
		// (CopyRowsTo records that after the postProcessor runs); otherwise redo would not restore the resolved tech
		// levels, nameables, or the merged points.
		CopyRowsTo(target.Table, sel, func(rows []*Node[T]) {
			target.provider.ProcessDropData(nil, target.Table)
			processCopiedRows(table, target.Table)
			clearPreconfiguredFlag(target.Table, rows)
		}, true)
	}
}

// processCopiedRows resolves the just-copied, currently-selected rows of a sheet's or template's table the same way a
// drop onto one does: prompting for the modifiers and nameables of rows that arrived from somewhere other than a
// sheet, then folding the points of rows that duplicate ones already present into those rows. Does nothing when the
// destination isn't a character sheet, loot sheet or template.
func processCopiedRows[T gurps.Node[T]](source, target *unison.Table[*Node[T]]) {
	if shouldProcessModifiersAndNameablesTo(target) {
		if shouldProcessModifiersAndNameablesFrom(source) {
			// Answering the modifier prompt rebuilds the owner, and that rebuild can replace the table underneath us:
			// only enabled modifiers count toward a row having switchable features, so toggling one can add or take
			// away the switch column, and a list can only change its columns by building a new table. An orphaned table
			// has no Rebuildable above it and reports its own rows as selected rather than the ones the user is now
			// looking at, both of which the steps below depend upon. Applying nameable substitutions rebuilds as well,
			// so look it up again afterwards too.
			ProcessModifiersForSelection(target)
			target = liveTable(target)
			ProcessNameablesForSelection(target)
			target = liveTable(target)
		}
		// The copy always adds rows to a different sheet, so merge points into identical existing rows even when
		// copying from another sheet.
		MergeAddedRows(target)
	}
}

// InsertCmdContextMenuItem inserts a context menu item for the given command.
func InsertCmdContextMenuItem[T gurps.Node[T]](table *unison.Table[*Node[T]], title string, cmdID int, id *int, cm unison.Menu) {
	if table.CanPerformCmd(table, cmdID) {
		useID := *id
		*id++
		cm.InsertItem(-1, cm.Factory().NewItem(unison.PopupMenuTemporaryBaseID+useID, title, unison.KeyBinding{}, nil,
			func(_ unison.MenuItem) {
				table.PerformCmd(table, cmdID)
			}))
	}
}

func flexibleLess(s1, s2 string) bool {
	c1 := strings.HasPrefix(s1, containerMarker)
	c2 := strings.HasPrefix(s2, containerMarker)
	if c1 != c2 {
		return c1
	}
	if c1 {
		s1 = s1[1:]
	}
	if c2 {
		s2 = s2[1:]
	}
	return xstrings.NaturalLess(strings.ReplaceAll(s1, ",", ""), strings.ReplaceAll(s2, ",", ""), true)
}

// OpenEditor opens an editor for each selected row in the table.
func OpenEditor[T gurps.Node[T]](table *unison.Table[*Node[T]], edit func(item T)) {
	var zero T
	selection := table.SelectedRows(false)
	if len(selection) > 4 {
		if unison.QuestionDialog(i18n.Text("Are you sure you want to open all of these?"),
			fmt.Sprintf(i18n.Text("%d editors will be opened."), len(selection))) != unison.ModalResponseOK {
			return
		}
	}
	for _, row := range selection {
		if data := row.Data(); data != zero {
			edit(data)
		}
	}
}

// DeleteSelection removes the selected nodes from the table and reports the change by rebuilding the table's owner.
func DeleteSelection[T gurps.Node[T]](table *unison.Table[*Node[T]], recordUndo bool) {
	deleteSelection(table, recordUndo, true)
}

// deleteSelection removes the selected nodes from the table. When report is true, the change is reported by rebuilding
// the table's owner; when it is false, the caller takes that on, for when the deletion is only one part of a larger
// edit whose parts should be reported once, together (see moveSelectedEquipment).
func deleteSelection[T gurps.Node[T]](table *unison.Table[*Node[T]], recordUndo, report bool) {
	if provider, ok := any(table.Model).(TableProvider[T]); ok && HasSelectionAndNotFiltered(table) {
		sel := table.SelectedRows(true)
		ids := make(map[tid.TID]bool, len(sel))
		list := make([]T, 0, len(sel))
		var zero T
		for _, row := range sel {
			unison.CollectIDsFromRow(row, ids)
			if target := row.Data(); target != zero {
				list = append(list, target)
			}
		}
		if !CloseID(ids) {
			return
		}
		var undo *unison.UndoEdit[*TableUndoEditData[T]]
		if recordUndo {
			undo = beginTableUndo(table, i18n.Text("Delete Selection"), nil, nil)
		}
		topLevelData := provider.RootData()
		for _, target := range list {
			parent := target.Parent()
			if parent == zero {
				for i, one := range topLevelData {
					if one == target {
						topLevelData = slices.Delete(topLevelData, i, i+1)
						break
					}
				}
			} else {
				children := parent.NodeChildren()
				for i, one := range children {
					if one == target {
						parent.SetChildren(slices.Delete(children, i, i+1))
						break
					}
				}
			}
		}
		provider.SetRootData(topLevelData)
		commitTableUndo(table, undo)
		if report {
			rebuildAsModified(table.AncestorOrSelf[Rebuildable](), true)
		}
	}
}

// DuplicateSelection duplicates the selected nodes in the table.
func DuplicateSelection[T gurps.Node[T]](table *unison.Table[*Node[T]]) {
	if provider, ok := any(table.Model).(TableProvider[T]); ok && HasSelectionAndNotFiltered(table) {
		undo := beginTableUndo(table, i18n.Text("Duplicate Selection"), nil, nil)
		var zero T
		needSet := false
		topLevelData := provider.RootData()
		sel := table.SelectedRows(true)
		selMap := make(map[tid.TID]bool, len(sel))
		for _, row := range sel {
			target := row.Data()
			if target == zero {
				continue
			}
			parent := target.Parent()
			clone := target.Clone(gurps.LibraryFile{}, gurps.EntityFromNode(target), parent, gurps.Duplicate)
			selMap[clone.ID()] = true
			if parent == zero {
				for i, child := range topLevelData {
					if child == target {
						topLevelData = slices.Insert(topLevelData, i+1, clone)
						needSet = true
						break
					}
				}
			} else {
				children := parent.NodeChildren()
				for i, child := range children {
					if child == target {
						parent.SetChildren(slices.Insert(children, i+1, clone))
						break
					}
				}
			}
		}
		if needSet {
			provider.SetRootData(topLevelData)
		}
		table.SyncToModel()
		table.SetSelectionMap(selMap)
		commitTableUndo(table, undo)
		rebuildAsModified(table.AncestorOrSelf[Rebuildable](), true)
	}
}

// HasSelectionAndNotFiltered returns true if the table has a selection and is not filtered.
func HasSelectionAndNotFiltered[T gurps.Node[T]](table *unison.Table[*Node[T]]) bool {
	return !table.IsFiltered() && table.HasSelection()
}

// ClearSourceFromSelection clears the source from the selected nodes.
func ClearSourceFromSelection[T gurps.Node[T]](table *unison.Table[*Node[T]]) {
	applyToSelectedRows(table, clearSourceAction.Title, T.ClearSource)
}

// SyncWithSourceForSelection synchronizes the selected nodes with their source.
func SyncWithSourceForSelection[T gurps.Node[T]](table *unison.Table[*Node[T]]) {
	applyToSelectedRows(table, syncWithSourceAction.Title, T.SyncWithSource)
}

// applyToSelectedRows applies the given change to the data behind each of the selected rows as a single undoable edit
// with the given title, then reports the change by rebuilding the table's owner. Nothing is done when the table has no
// selection or is showing search results (see HasSelectionAndNotFiltered).
func applyToSelectedRows[T gurps.Node[T]](table *unison.Table[*Node[T]], undoTitle string, apply func(T)) {
	if !HasSelectionAndNotFiltered(table) {
		return
	}
	undo := beginTableUndo(table, undoTitle, nil, nil)
	var zero T
	for _, row := range table.SelectedRows(false) {
		if target := row.Data(); target != zero {
			apply(target)
		}
	}
	table.SyncToModel()
	commitTableUndo(table, undo)
	rebuildAsModified(table.AncestorOrSelf[Rebuildable](), true)
}

// CopyRowsTo copies the provided rows to the target table and reports the change by rebuilding the table's owner.
func CopyRowsTo[T gurps.Node[T]](table *unison.Table[*Node[T]], rows []*Node[T], postProcessor func(rows []*Node[T]), recordUndo bool) {
	copyRowsTo(table, rows, postProcessor, recordUndo, true)
}

// copyRowsTo copies the provided rows to the target table. When report is true, the change is reported by rebuilding
// the table's owner; when it is false, the caller takes that on, for when the copy is only one part of a larger edit
// whose parts should be reported once, together (see moveSelectedEquipment).
func copyRowsTo[T gurps.Node[T]](table *unison.Table[*Node[T]], rows []*Node[T], postProcessor func(rows []*Node[T]), recordUndo, report bool) {
	if table == nil || table.IsFiltered() {
		return
	}
	rows = slices.Clone(rows)
	for j, row := range rows {
		rows[j] = row.CloneForTarget(table, nil)
	}
	var undo *unison.UndoEdit[*TableUndoEditData[T]]
	if recordUndo {
		undo = beginTableUndo(table, fmt.Sprintf(i18n.Text("Insert %s"), rows[0].Data().Kind()), nil, nil)
	}
	table.SetRootRows(append(slices.Clone(table.RootRows()), rows...))
	selMap := make(map[tid.TID]bool, len(rows))
	for _, row := range rows {
		selMap[row.ID()] = true
	}
	table.SetSelectionMap(selMap)
	if postProcessor != nil {
		postProcessor(rows)
		// The post-processing can put a prompt in front of the user, and answering one rebuilds the owner, which can
		// replace the table: only enabled modifiers count toward a row having switchable features, so toggling one can
		// add or take away the switch column, and a list can only change its columns by building a new table. That
		// leaves the table we were handed orphaned, with no Rebuildable above it, so the scroll below would aim at a
		// detached table and the closing rebuild would be skipped entirely. The selection needs no restoring, since
		// the rebuild that produced the replacement records it and puts it back by row ID.
		table = liveTable(table)
	}
	table.ScrollRowCellIntoView(table.LastSelectedRowIndex(), 0)
	table.ScrollRowCellIntoView(table.FirstSelectedRowIndex(), 0)
	commitTableUndo(table, undo)
	if report {
		rebuildAsModified(table.AncestorOrSelf[Rebuildable](), true)
	}
}

// syncTablePreservingSelection re-syncs the table's rows with its model, keeping whatever was selected.
func syncTablePreservingSelection[T unison.TableRowConstraint[T]](table *unison.Table[T]) {
	sel := table.CopySelectionMap()
	table.SyncToModel()
	table.SetSelectionMap(sel)
}

// DisableSorting disables the sorting capability in the table headers.
func DisableSorting[T unison.TableRowConstraint[T]](headers []unison.TableColumnHeader[T]) []unison.TableColumnHeader[T] {
	for _, header := range headers {
		state := header.SortState()
		state.Sortable = false
		header.SetSortState(state)
	}
	return headers
}
