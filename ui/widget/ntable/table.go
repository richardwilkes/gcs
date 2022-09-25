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

package ntable

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/richardwilkes/gcs/v5/constants"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/theme"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/gcs/v5/ui/workspace"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xmath/geom"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/slices"
)

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
	HierarchyColumnIndex() int
	ExcessWidthColumnIndex() int
	ContextMenuItems() []ContextMenuItem
	OpenEditor(owner widget.Rebuildable, table *unison.Table[*Node[T]])
	CreateItem(owner widget.Rebuildable, table *unison.Table[*Node[T]], variant ItemVariant)
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
}

// NewNodeTable creates a new node table of the specified type, returning the header and table. Pass nil for 'font' if
// this should be a standalone top-level table for a dockable. Otherwise, pass in the typical font used for a cell.
func NewNodeTable[T gurps.NodeTypes](provider TableProvider[T], font unison.Font) (header *unison.TableHeader[*Node[T]], table *unison.Table[*Node[T]]) {
	table = unison.NewTable[*Node[T]](provider)
	provider.SetTable(table)
	table.HierarchyColumnIndex = provider.HierarchyColumnIndex()
	table.DividerInk = theme.HeaderColor
	layoutData := &unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
		VGrab:  true,
	}
	if font != nil {
		table.Padding.Top = 0
		table.Padding.Bottom = 0
		table.HierarchyIndent = font.LineHeight()
		table.MinimumRowHeight = font.LineHeight()
		layoutData.MinSize = unison.Size{Height: 4 + theme.PageFieldPrimaryFont.LineHeight()}
	}
	table.SetLayoutData(layoutData)

	headers := provider.Headers()
	table.ColumnSizes = make([]unison.ColumnSize, len(headers))
	for i := range table.ColumnSizes {
		_, pref, _ := headers[i].AsPanel().Sizes(unison.Size{})
		pref.Width += table.Padding.Left + table.Padding.Right
		table.ColumnSizes[i].AutoMinimum = pref.Width
		table.ColumnSizes[i].AutoMaximum = 800
		table.ColumnSizes[i].Minimum = pref.Width
		table.ColumnSizes[i].Maximum = 10000
	}
	header = unison.NewTableHeader(table, headers...)
	header.Less = flexibleLess
	header.BackgroundInk = theme.HeaderColor
	header.DividerInk = unison.DividerColor
	header.HeaderBorder = unison.NewLineBorder(theme.HeaderColor, 0, unison.Insets{Bottom: 1}, false)
	header.SetBorder(header.HeaderBorder)
	header.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
	})

	table.DoubleClickCallback = func() { table.PerformCmd(nil, constants.OpenEditorItemID) }
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
			table.SizeColumnsToFitWithExcessIn(provider.ExcessWidthColumnIndex())
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
				cm.Popup(geom.Rect[float32]{
					Point: table.PointToRoot(where),
					Size: geom.Size[float32]{
						Width:  1,
						Height: 1,
					},
				}, 0)
			}
			cm.Dispose()
		}
		return stop
	}

	return header, table
}

// InsertCmdContextMenuItem inserts a context menu item for the given command.
func InsertCmdContextMenuItem[T gurps.NodeTypes](table *unison.Table[*Node[T]], title string, cmdID int, id *int, cm unison.Menu) {
	if table.CanPerformCmd(table, cmdID) {
		useID := *id
		*id++
		cm.InsertItem(-1, cm.Factory().NewItem(unison.PopupMenuTemporaryBaseID+useID, title, unison.KeyBinding{}, nil,
			func(item unison.MenuItem) {
				table.PerformCmd(table, cmdID)
			}))
	}
}

func flexibleLess(s1, s2 string) bool {
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
func DeleteSelection[T gurps.NodeTypes](table *unison.Table[*Node[T]]) {
	if provider, ok := table.Model.(TableProvider[T]); ok && table.HasSelection() {
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
		if !workspace.CloseUUID(ids) {
			return
		}
		var undo *unison.UndoEdit[*TableUndoEditData[T]]
		mgr := unison.UndoManagerFor(table)
		if mgr != nil {
			undo = &unison.UndoEdit[*TableUndoEditData[T]]{
				ID:         unison.NextUndoID(),
				EditName:   i18n.Text("Delete Selection"),
				UndoFunc:   func(e *unison.UndoEdit[*TableUndoEditData[T]]) { e.BeforeData.Apply() },
				RedoFunc:   func(e *unison.UndoEdit[*TableUndoEditData[T]]) { e.AfterData.Apply() },
				AbsorbFunc: func(e *unison.UndoEdit[*TableUndoEditData[T]], other unison.Undoable) bool { return false },
				BeforeData: NewTableUndoEditData(table),
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
		if mgr != nil && undo != nil {
			undo.AfterData = NewTableUndoEditData(table)
			mgr.Add(undo)
		}
		if builder := unison.AncestorOrSelf[widget.Rebuildable](table); builder != nil {
			builder.Rebuild(true)
		}
	}
}

// DuplicateSelection duplicates the selected nodes in the table.
func DuplicateSelection[T gurps.NodeTypes](table *unison.Table[*Node[T]]) {
	if provider, ok := table.Model.(TableProvider[T]); ok && table.HasSelection() {
		var undo *unison.UndoEdit[*TableUndoEditData[T]]
		mgr := unison.UndoManagerFor(table)
		if mgr != nil {
			undo = &unison.UndoEdit[*TableUndoEditData[T]]{
				ID:         unison.NextUndoID(),
				EditName:   i18n.Text("Duplicate Selection"),
				UndoFunc:   func(e *unison.UndoEdit[*TableUndoEditData[T]]) { e.BeforeData.Apply() },
				RedoFunc:   func(e *unison.UndoEdit[*TableUndoEditData[T]]) { e.AfterData.Apply() },
				AbsorbFunc: func(e *unison.UndoEdit[*TableUndoEditData[T]], other unison.Undoable) bool { return false },
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
		if builder := unison.AncestorOrSelf[widget.Rebuildable](table); builder != nil {
			builder.Rebuild(true)
		}
	}
}

// CopyRowsTo copies the provided rows to the target table.
func CopyRowsTo[T gurps.NodeTypes](table *unison.Table[*Node[T]], rows []*Node[T]) {
	if table == nil {
		return
	}
	rows = slices.Clone(rows)
	for j, row := range rows {
		rows[j] = row.CloneForTarget(table, nil)
	}
	var undo *unison.UndoEdit[*TableUndoEditData[T]]
	mgr := unison.UndoManagerFor(table)
	if mgr != nil {
		undo = &unison.UndoEdit[*TableUndoEditData[T]]{
			ID:         unison.NextUndoID(),
			EditName:   fmt.Sprintf(i18n.Text("Insert %s"), gurps.AsNode(rows[0].Data()).Kind()),
			UndoFunc:   func(e *unison.UndoEdit[*TableUndoEditData[T]]) { e.BeforeData.Apply() },
			RedoFunc:   func(e *unison.UndoEdit[*TableUndoEditData[T]]) { e.AfterData.Apply() },
			AbsorbFunc: func(e *unison.UndoEdit[*TableUndoEditData[T]], other unison.Undoable) bool { return false },
			BeforeData: NewTableUndoEditData(table),
		}
	}
	table.SetRootRows(append(slices.Clone(table.RootRows()), rows...))
	selMap := make(map[uuid.UUID]bool, len(rows))
	for _, row := range rows {
		selMap[row.UUID()] = true
	}
	table.SetSelectionMap(selMap)
	table.ScrollRowCellIntoView(table.LastSelectedRowIndex(), 0)
	table.ScrollRowCellIntoView(table.FirstSelectedRowIndex(), 0)
	if mgr != nil && undo != nil {
		undo.AfterData = NewTableUndoEditData(table)
		mgr.Add(undo)
	}
	unison.Ancestor[widget.Rebuildable](table).Rebuild(true)
}
