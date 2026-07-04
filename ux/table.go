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

// ItemVariant holds the type of item variant to create.
type ItemVariant int

// Possible values for ItemVariant.
const (
	NoItemVariant ItemVariant = iota
	ContainerItemVariant
	AlternateItemVariant
)

// TableProvider defines the methods a table provider must contain.
type TableProvider[T gurps.NodeTypes] interface {
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
func NewNodeTable[T gurps.NodeTypes](provider TableProvider[T], font unison.Font) (header *unison.TableHeader[*Node[T]], table *unison.Table[*Node[T]]) {
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
	// Mirror the dragged rows into our own storage so that alternate drop handlers (which deal with a different row
	// type than the destination table) can access them, since unison only retains them internally.
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
					Point: table.PointToRoot(where),
					Size: geom.Size{
						Width:  1,
						Height: 1,
					},
				}, 0)
			}
			cm.Dispose()
		}
		return stop
	}

	table.InstallCmdHandlers(CopyToSheetItemID, func(_ any) bool { return canCopySelectionToSheet(table) },
		func(_ any) { copySelectionToSheet(table) })
	table.InstallCmdHandlers(CopyToTemplateItemID, func(_ any) bool { return canCopySelectionToTemplate(table) },
		func(_ any) { copySelectionToTemplate(table) })
	if t, ok := (any(table)).(*unison.Table[*Node[*gurps.Equipment]]); ok {
		t.InstallCmdHandlers(IncrementItemID,
			func(_ any) bool { return canAdjustQuantity(t, true) },
			func(_ any) { adjustQuantity(unison.AncestorOrSelf[Rebuildable](t), t, true) })
		t.InstallCmdHandlers(DecrementItemID,
			func(_ any) bool { return canAdjustQuantity(t, false) },
			func(_ any) { adjustQuantity(unison.AncestorOrSelf[Rebuildable](t), t, false) })
		t.InstallCmdHandlers(IncrementUsesItemID,
			func(_ any) bool { return canAdjustUses(t, 1) },
			func(_ any) { adjustUses(unison.AncestorOrSelf[Rebuildable](t), t, 1) })
		t.InstallCmdHandlers(DecrementUsesItemID,
			func(_ any) bool { return canAdjustUses(t, -1) },
			func(_ any) { adjustUses(unison.AncestorOrSelf[Rebuildable](t), t, -1) })
	}

	return header, table
}

// sizePageTableColumns sizes the columns of a fixed-width page table to fit, then, when the user can't resize columns
// and the setting is enabled, lets any page reference columns claim leftover space so they can show more than one
// reference before the rest goes to the excess column. This is run both when the table's frame changes and when it is
// synced, since a settings change won't necessarily alter the frame.
func sizePageTableColumns[T gurps.NodeTypes](table *unison.Table[*Node[T]], excessColumnID int) {
	table.SizeColumnsToFitWithExcessIn(excessColumnID)
	if table.PreventUserColumnResize && gurps.GlobalSettings().General.ExpandPageReferences &&
		expandPageRefColumns(table, excessColumnID) {
		table.SyncRowHeights()
	}
}

// expandPageRefColumns hands leftover horizontal space to any page reference columns so they can display more than a
// single reference, taking that space from the excess-width column. The excess column first keeps enough room to show
// its primary content with notes collapsed; only the space beyond that is offered to the page reference columns, and
// each grows only up to the width needed to show all of its references (capped by its AutoMaximum). Whatever isn't
// claimed stays with the excess column. Returns true if any column width was changed.
func expandPageRefColumns[T gurps.NodeTypes](table *unison.Table[*Node[T]], excessColumnID int) bool {
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
		// Determine the width needed to show every reference in this column across all rows, skipping the column
		// entirely if it isn't a page reference column.
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

func isAcceptableTypeForSheetOrTemplate(data any) bool {
	switch data.(type) {
	case *gurps.Equipment, *gurps.Note, *gurps.Skill, *gurps.Spell, *gurps.Trait:
		return true
	default:
		return false
	}
}

func canCopySelectionToSheet[T gurps.NodeTypes](table *unison.Table[*Node[T]]) bool {
	var t T
	return table.HasSelection() && len(OpenSheets(unison.Ancestor[*Sheet](table))) > 0 && isAcceptableTypeForSheetOrTemplate(t)
}

func canCopySelectionToTemplate[T gurps.NodeTypes](table *unison.Table[*Node[T]]) bool {
	var t T
	return table.HasSelection() && len(OpenTemplates(unison.Ancestor[*Template](table))) > 0 && isAcceptableTypeForSheetOrTemplate(t)
}

func libraryFileFromTable[T gurps.NodeTypes](table *unison.Table[*Node[T]]) gurps.LibraryFile {
	if d := unison.Ancestor[*TableDockable[T]](table); d != nil {
		for _, lib := range gurps.GlobalSettings().Libraries() {
			libPathOnDisk := lib.PathOnDisk + string(filepath.Separator)
			filePathOnDisk := d.BackingFilePath()
			if strings.HasPrefix(filePathOnDisk, libPathOnDisk) {
				return gurps.LibraryFile{
					Library: lib.Key(),
					Path:    filepath.ToSlash(strings.TrimPrefix(filePathOnDisk, libPathOnDisk)),
				}
			}
		}
	}
	return gurps.LibraryFile{}
}

func copySelectionToSheet[T gurps.NodeTypes](table *unison.Table[*Node[T]]) {
	if table.HasSelection() {
		if sheets := PromptForDestination(OpenSheets(unison.Ancestor[*Sheet](table))); len(sheets) > 0 {
			sel := table.SelectedRows(true)
			for _, s := range sheets {
				var targetTable *unison.Table[*Node[T]]
				var processDropData func()
				switch any(sel[0].Data()).(type) {
				case *gurps.Trait:
					targetTable = convertTable[T](s.Traits.Table)
					processDropData = func() { s.Traits.provider.ProcessDropData(nil, s.Traits.Table) }
				case *gurps.Skill:
					targetTable = convertTable[T](s.Skills.Table)
					processDropData = func() { s.Skills.provider.ProcessDropData(nil, s.Skills.Table) }
				case *gurps.Spell:
					targetTable = convertTable[T](s.Spells.Table)
					processDropData = func() { s.Spells.provider.ProcessDropData(nil, s.Spells.Table) }
				case *gurps.Equipment:
					targetTable = convertTable[T](s.CarriedEquipment.Table)
					processDropData = func() { s.CarriedEquipment.provider.ProcessDropData(nil, s.CarriedEquipment.Table) }
				case *gurps.Note:
					targetTable = convertTable[T](s.Notes.Table)
					processDropData = func() { s.Notes.provider.ProcessDropData(nil, s.Notes.Table) }
				default:
					continue
				}
				if targetTable != nil {
					// All processing must happen inside the postProcessor so it is captured by the undo edit's
					// after-state (CopyRowsTo records that after the postProcessor runs); otherwise redo would not
					// restore the resolved tech levels, nameables, or the merged points.
					CopyRowsTo(targetTable, sel, func(_ []*Node[T]) {
						processDropData()
						if isForCharacterOrLootSheet(targetTable) {
							// Only process modifiers and nameables when copying from something besides a character or
							// loot sheet; rows already on a sheet have had these resolved.
							if !isForCharacterOrLootSheet(table) {
								ProcessModifiersForSelection(targetTable)
								ProcessNameablesForSelection(targetTable)
							}
							// The copy always adds rows to a different sheet, so always merge points into identical
							// existing rows, including when copying from another sheet.
							MergeAddedSkillsAndSpells(targetTable)
						}
					}, true)
				}
			}
		}
	}
}

func copySelectionToTemplate[T gurps.NodeTypes](table *unison.Table[*Node[T]]) {
	if table.HasSelection() {
		if templates := PromptForDestination(OpenTemplates(unison.Ancestor[*Template](table))); len(templates) > 0 {
			sel := table.SelectedRows(true)
			for _, t := range templates {
				switch any(sel[0].Data()).(type) {
				case *gurps.Trait:
					CopyRowsTo(convertTable[T](t.Traits.Table), sel, nil, true)
				case *gurps.Skill:
					CopyRowsTo(convertTable[T](t.Skills.Table), sel, nil, true)
				case *gurps.Spell:
					CopyRowsTo(convertTable[T](t.Spells.Table), sel, nil, true)
				case *gurps.Equipment:
					CopyRowsTo(convertTable[T](t.Equipment.Table), sel, nil, true)
				case *gurps.Note:
					CopyRowsTo(convertTable[T](t.Notes.Table), sel, nil, true)
				}
			}
		}
	}
}

func convertTable[T gurps.NodeTypes](table any) *unison.Table[*Node[T]] {
	// This is here just to get around limitations in the way Go generics behave
	if t, ok := table.(*unison.Table[*Node[T]]); ok {
		return t
	}
	return nil
}

// InsertCmdContextMenuItem inserts a context menu item for the given command.
func InsertCmdContextMenuItem[T gurps.NodeTypes](table *unison.Table[*Node[T]], title string, cmdID int, id *int, cm unison.Menu) {
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
func OpenEditor[T gurps.NodeTypes](table *unison.Table[*Node[T]], edit func(item T)) {
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

// DeleteSelection removes the selected nodes from the table.
func DeleteSelection[T gurps.NodeTypes](table *unison.Table[*Node[T]], recordUndo bool) {
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
		var mgr *unison.UndoManager
		if recordUndo {
			if mgr = unison.UndoManagerFor(table); mgr != nil {
				undo = &unison.UndoEdit[*TableUndoEditData[T]]{
					ID:         unison.NextUndoID(),
					EditName:   i18n.Text("Delete Selection"),
					UndoFunc:   func(e *unison.UndoEdit[*TableUndoEditData[T]]) { e.BeforeData.Apply() },
					RedoFunc:   func(e *unison.UndoEdit[*TableUndoEditData[T]]) { e.AfterData.Apply() },
					AbsorbFunc: func(_ *unison.UndoEdit[*TableUndoEditData[T]], _ unison.Undoable) bool { return false },
					BeforeData: NewTableUndoEditData(table),
				}
			}
		}
		topLevelData := provider.RootData()
		for _, target := range list {
			parent := gurps.AsNode(target).Parent()
			if parent == zero {
				for i, one := range topLevelData {
					if one == target {
						topLevelData = slices.Delete(topLevelData, i, i+1)
						break
					}
				}
			} else {
				pNode := gurps.AsNode(parent)
				children := pNode.NodeChildren()
				for i, one := range children {
					if one == target {
						pNode.SetChildren(slices.Delete(children, i, i+1))
						break
					}
				}
			}
		}
		provider.SetRootData(topLevelData)
		if recordUndo && mgr != nil && undo != nil {
			undo.AfterData = NewTableUndoEditData(table)
			mgr.Add(undo)
		}
		if builder := unison.AncestorOrSelf[Rebuildable](table); builder != nil {
			builder.Rebuild(true)
		}
	}
}

// DuplicateSelection duplicates the selected nodes in the table.
func DuplicateSelection[T gurps.NodeTypes](table *unison.Table[*Node[T]]) {
	if provider, ok := any(table.Model).(TableProvider[T]); ok && HasSelectionAndNotFiltered(table) {
		var undo *unison.UndoEdit[*TableUndoEditData[T]]
		mgr := unison.UndoManagerFor(table)
		if mgr != nil {
			undo = &unison.UndoEdit[*TableUndoEditData[T]]{
				ID:         unison.NextUndoID(),
				EditName:   i18n.Text("Duplicate Selection"),
				UndoFunc:   func(e *unison.UndoEdit[*TableUndoEditData[T]]) { e.BeforeData.Apply() },
				RedoFunc:   func(e *unison.UndoEdit[*TableUndoEditData[T]]) { e.AfterData.Apply() },
				AbsorbFunc: func(_ *unison.UndoEdit[*TableUndoEditData[T]], _ unison.Undoable) bool { return false },
				BeforeData: NewTableUndoEditData(table),
			}
		}
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
			tData := gurps.AsNode(target)
			parent := tData.Parent()
			clone := tData.Clone(gurps.LibraryFile{}, gurps.EntityFromNode(tData), parent, false)
			selMap[gurps.AsNode(clone).ID()] = true
			if parent == zero {
				for i, child := range topLevelData {
					if child == target {
						topLevelData = slices.Insert(topLevelData, i+1, clone)
						needSet = true
						break
					}
				}
			} else {
				pNode := gurps.AsNode(parent)
				children := pNode.NodeChildren()
				for i, child := range children {
					if child == target {
						pNode.SetChildren(slices.Insert(children, i+1, clone))
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
		if mgr != nil && undo != nil {
			undo.AfterData = NewTableUndoEditData(table)
			mgr.Add(undo)
		}
		if builder := unison.AncestorOrSelf[Rebuildable](table); builder != nil {
			builder.Rebuild(true)
		}
	}
}

// HasSelectionAndNotFiltered returns true if the table has a selection and is not filtered.
func HasSelectionAndNotFiltered[T gurps.NodeTypes](table *unison.Table[*Node[T]]) bool {
	return !table.IsFiltered() && table.HasSelection()
}

// ClearSourceFromSelection clears the source from the selected nodes.
func ClearSourceFromSelection[T gurps.NodeTypes](table *unison.Table[*Node[T]]) {
	if HasSelectionAndNotFiltered(table) {
		var undo *unison.UndoEdit[*TableUndoEditData[T]]
		mgr := unison.UndoManagerFor(table)
		if mgr != nil {
			undo = &unison.UndoEdit[*TableUndoEditData[T]]{
				ID:         unison.NextUndoID(),
				EditName:   clearSourceAction.Title,
				UndoFunc:   func(e *unison.UndoEdit[*TableUndoEditData[T]]) { e.BeforeData.Apply() },
				RedoFunc:   func(e *unison.UndoEdit[*TableUndoEditData[T]]) { e.AfterData.Apply() },
				AbsorbFunc: func(_ *unison.UndoEdit[*TableUndoEditData[T]], _ unison.Undoable) bool { return false },
				BeforeData: NewTableUndoEditData(table),
			}
		}
		var zero T
		sel := table.SelectedRows(false)
		for _, row := range sel {
			if target := row.Data(); target != zero {
				gurps.AsNode(target).ClearSource()
			}
		}
		table.SyncToModel()
		if mgr != nil && undo != nil {
			undo.AfterData = NewTableUndoEditData(table)
			mgr.Add(undo)
		}
		if builder := unison.AncestorOrSelf[Rebuildable](table); builder != nil {
			builder.Rebuild(true)
		}
	}
}

// SyncWithSourceForSelection synchronizes the selected nodes with their source.
func SyncWithSourceForSelection[T gurps.NodeTypes](table *unison.Table[*Node[T]]) {
	if HasSelectionAndNotFiltered(table) {
		var undo *unison.UndoEdit[*TableUndoEditData[T]]
		mgr := unison.UndoManagerFor(table)
		if mgr != nil {
			undo = &unison.UndoEdit[*TableUndoEditData[T]]{
				ID:         unison.NextUndoID(),
				EditName:   syncWithSourceAction.Title,
				UndoFunc:   func(e *unison.UndoEdit[*TableUndoEditData[T]]) { e.BeforeData.Apply() },
				RedoFunc:   func(e *unison.UndoEdit[*TableUndoEditData[T]]) { e.AfterData.Apply() },
				AbsorbFunc: func(_ *unison.UndoEdit[*TableUndoEditData[T]], _ unison.Undoable) bool { return false },
				BeforeData: NewTableUndoEditData(table),
			}
		}
		var zero T
		sel := table.SelectedRows(false)
		for _, row := range sel {
			if target := row.Data(); target != zero {
				gurps.AsNode(target).SyncWithSource()
			}
		}
		table.SyncToModel()
		if mgr != nil && undo != nil {
			undo.AfterData = NewTableUndoEditData(table)
			mgr.Add(undo)
		}
		if builder := unison.AncestorOrSelf[Rebuildable](table); builder != nil {
			builder.Rebuild(true)
		}
	}
}

// CopyRowsTo copies the provided rows to the target table.
func CopyRowsTo[T gurps.NodeTypes](table *unison.Table[*Node[T]], rows []*Node[T], postProcessor func(rows []*Node[T]), recordUndo bool) {
	if table == nil || table.IsFiltered() {
		return
	}
	rows = slices.Clone(rows)
	for j, row := range rows {
		rows[j] = row.CloneForTarget(table, nil)
	}
	var undo *unison.UndoEdit[*TableUndoEditData[T]]
	var mgr *unison.UndoManager
	if recordUndo {
		if mgr = unison.UndoManagerFor(table); mgr != nil {
			undo = &unison.UndoEdit[*TableUndoEditData[T]]{
				ID:         unison.NextUndoID(),
				EditName:   fmt.Sprintf(i18n.Text("Insert %s"), gurps.AsNode(rows[0].Data()).Kind()),
				UndoFunc:   func(e *unison.UndoEdit[*TableUndoEditData[T]]) { e.BeforeData.Apply() },
				RedoFunc:   func(e *unison.UndoEdit[*TableUndoEditData[T]]) { e.AfterData.Apply() },
				AbsorbFunc: func(_ *unison.UndoEdit[*TableUndoEditData[T]], _ unison.Undoable) bool { return false },
				BeforeData: NewTableUndoEditData(table),
			}
		}
	}
	table.SetRootRows(append(slices.Clone(table.RootRows()), rows...))
	selMap := make(map[tid.TID]bool, len(rows))
	for _, row := range rows {
		selMap[row.ID()] = true
	}
	table.SetSelectionMap(selMap)
	if postProcessor != nil {
		postProcessor(rows)
	}
	table.ScrollRowCellIntoView(table.LastSelectedRowIndex(), 0)
	table.ScrollRowCellIntoView(table.FirstSelectedRowIndex(), 0)
	if recordUndo && mgr != nil && undo != nil {
		undo.AfterData = NewTableUndoEditData(table)
		mgr.Add(undo)
	}
	if builder := unison.AncestorOrSelf[Rebuildable](table); builder != nil {
		builder.Rebuild(true)
	}
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
