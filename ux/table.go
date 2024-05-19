/*
 * Copyright ©1998-2024 by Richard A. Wilkes. All rights reserved.
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
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
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
	gurps.EntityProvider
	SetTable(table *unison.Table[*Node[T]])
	RootData() []T
	SetRootData(data []T)
	DragKey() string
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
	table = unison.NewTable[*Node[T]](provider)
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
		layoutData.MinSize = unison.Size{Height: 4 + gurps.PageFieldPrimaryFont.LineHeight()}
	}
	table.SetLayoutData(layoutData)

	ids := provider.ColumnIDs()
	headers := provider.Headers()
	table.Columns = make([]unison.ColumnInfo, len(headers))
	for i := range table.Columns {
		_, pref, _ := headers[i].AsPanel().Sizes(unison.Size{})
		pref.Width += table.Padding.Left + table.Padding.Right
		table.Columns[i].ID = ids[i]
		table.Columns[i].AutoMinimum = pref.Width
		table.Columns[i].AutoMaximum = max(float32(gurps.GlobalSettings().General.MaximumAutoColWidth), pref.Width)
		table.Columns[i].Minimum = pref.Width
		table.Columns[i].Maximum = 10000
	}
	header = unison.NewTableHeader(table, headers...)
	header.Less = flexibleLess
	header.BackgroundInk = gurps.ThemeHeader
	header.InteriorDividerColor = gurps.ThemeHeader
	header.SetBorder(header.HeaderBorder)
	header.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
	})

	table.DoubleClickCallback = func() { table.PerformCmd(nil, OpenEditorItemID) }
	table.KeyDownCallback = func(keyCode unison.KeyCode, mod unison.Modifiers, repeat bool) bool {
		if mod == 0 && (keyCode == unison.KeyBackspace || keyCode == unison.KeyDelete) {
			table.PerformCmd(table, unison.DeleteItemID)
			return true
		}
		return table.DefaultKeyDown(keyCode, mod, repeat)
	}
	singular, plural := provider.ItemNames()
	table.InstallDragSupport(provider.DragSVG(), provider.DragKey(), singular, plural)
	if font != nil {
		table.FrameChangeCallback = func() {
			table.SizeColumnsToFitWithExcessIn(provider.ExcessWidthColumnID())
		}
	}

	table.MouseDownCallback = func(where unison.Point, button, clickCount int, mod unison.Modifiers) bool {
		stop := table.DefaultMouseDown(where, button, clickCount, mod)
		if button == unison.ButtonRight && clickCount == 1 && !table.Window().InDrag() {
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
				cm.Popup(unison.Rect{
					Point: table.PointToRoot(where),
					Size: unison.Size{
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

func copySelectionToSheet[T gurps.NodeTypes](table *unison.Table[*Node[T]]) {
	if table.HasSelection() {
		if sheets := PromptForDestination(OpenSheets(unison.Ancestor[*Sheet](table))); len(sheets) > 0 {
			sel := table.SelectedRows(true)
			for _, s := range sheets {
				var targetTable *unison.Table[*Node[T]]
				var postProcessor func(rows []*Node[T])
				switch any(sel[0].Data()).(type) {
				case *gurps.Trait:
					targetTable = convertTable[T](s.Traits.Table)
					postProcessor = func(_ []*Node[T]) {
						s.Traits.provider.ProcessDropData(nil, s.Traits.Table)
					}
				case *gurps.Skill:
					targetTable = convertTable[T](s.Skills.Table)
					postProcessor = func(_ []*Node[T]) {
						s.Skills.provider.ProcessDropData(nil, s.Skills.Table)
					}
				case *gurps.Spell:
					targetTable = convertTable[T](s.Spells.Table)
					postProcessor = func(_ []*Node[T]) {
						s.Spells.provider.ProcessDropData(nil, s.Spells.Table)
					}
				case *gurps.Equipment:
					targetTable = convertTable[T](s.CarriedEquipment.Table)
					postProcessor = func(_ []*Node[T]) {
						s.CarriedEquipment.provider.ProcessDropData(nil, s.CarriedEquipment.Table)
					}
				case *gurps.Note:
					targetTable = convertTable[T](s.Notes.Table)
					postProcessor = func(_ []*Node[T]) {
						s.Notes.provider.ProcessDropData(nil, s.Notes.Table)
					}
				default:
					continue
				}
				if targetTable != nil {
					CopyRowsTo(targetTable, sel, postProcessor, true)
					ProcessModifiersForSelection(targetTable)
					ProcessNameablesForSelection(targetTable)
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
	if n1, err := fxp.FromString(s1); err == nil {
		var n2 fxp.Int
		if n2, err = fxp.FromString(s2); err == nil {
			return n1 < n2
		}
	}
	return txt.NaturalLess(s1, s2, true)
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
	if provider, ok := any(table.Model).(TableProvider[T]); ok && !table.IsFiltered() && table.HasSelection() {
		sel := table.SelectedRows(true)
		ids := make(map[uuid.UUID]bool, len(sel))
		list := make([]T, 0, len(sel))
		var zero T
		for _, row := range sel {
			unison.CollectUUIDsFromRow(row, ids)
			if target := row.Data(); target != zero {
				list = append(list, target)
			}
		}
		if !CloseUUID(ids) {
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
		needSet := false
		topLevelData := provider.RootData()
		for _, target := range list {
			parent := gurps.AsNode(target).Parent()
			if parent == zero {
				for i, one := range topLevelData {
					if one == target {
						topLevelData = slices.Delete(topLevelData, i, i+1)
						needSet = true
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
		if needSet {
			provider.SetRootData(topLevelData)
		}
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
	if provider, ok := any(table.Model).(TableProvider[T]); ok && !table.IsFiltered() && table.HasSelection() {
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
		selMap := make(map[uuid.UUID]bool, len(sel))
		for _, row := range sel {
			if target := row.Data(); target != zero {
				tData := gurps.AsNode(target)
				parent := tData.Parent()
				clone := tData.Clone(tData.OwningEntity(), parent, false)
				selMap[gurps.AsNode(clone).UUID()] = true
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
	selMap := make(map[uuid.UUID]bool, len(rows))
	for _, row := range rows {
		selMap[row.UUID()] = true
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
	unison.Ancestor[Rebuildable](table).Rebuild(true)
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
