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
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/mod"
)

var (
	_ unison.Dockable                         = &shoppingRecordDockable{}
	_ GroupedCloser                           = &shoppingRecordDockable{}
	_ SheetSourceUser                         = &shoppingRecordDockable{}
	_ unison.UndoManagerProvider              = &shoppingRecordDockable{}
	_ unison.TableRowData[*shoppingRecordRow] = &shoppingRecordRow{}
)

// shoppingRecordHintRows is how many entries the record asks for room to show, however few it holds.
const shoppingRecordHintRows = 12

// The columns of the shopping record, whose IDs are also their indexes.
const (
	shoppingRecordWhenColumn = iota
	shoppingRecordTransactionColumn
	shoppingRecordTypeColumn
	shoppingRecordItemsColumn
	shoppingRecordMoneyColumn
	shoppingRecordColumnCount
)

// shoppingRecordDockable shows a sheet's shopping record, most recent first, and lets the user remove entries from it.
// The record belongs to the sheet, so the removals are made on the sheet's undo manager, which is also the one Undo
// and Redo use while the record has the focus.
type shoppingRecordDockable struct {
	unison.Panel
	sheet *Sheet
	table *unison.Table[*shoppingRecordRow]
	title string
}

// showShoppingRecord shows the sheet's shopping record, bringing forward the one already open for it if there is one.
func showShoppingRecord(sheet *Sheet) {
	if ActivateDockableOfType(func(d *shoppingRecordDockable) bool { return d.sheet == sheet }) {
		return
	}
	d := newShoppingRecordDockable(sheet)
	PlaceInDock(d, dgroup.Editors, false)
	d.table.RequestFocus()
}

func newShoppingRecordDockable(sheet *Sheet) *shoppingRecordDockable {
	d := &shoppingRecordDockable{
		sheet: sheet,
		table: unison.NewTable(&unison.SimpleTableModel[*shoppingRecordRow]{}),
	}
	d.Self = d
	d.SetLayout(&unison.FlexLayout{Columns: 1})

	d.table.Accessibility.Name = i18n.Text("Shopping Record")
	d.table.ShowFirstColumnDivider = false
	d.table.PreventUserColumnResize = true
	headers := []unison.TableColumnHeader[*shoppingRecordRow]{
		newShoppingRecordColumnHeader(i18n.Text("When")),
		newShoppingRecordColumnHeader(i18n.Text("Transaction")),
		newShoppingRecordColumnHeader(i18n.Text("Type")),
		newShoppingRecordColumnHeader(i18n.Text("Items")),
		newShoppingRecordColumnHeader(i18n.Text("Money Exchanged")),
	}
	d.table.Columns = make([]unison.ColumnInfo, len(headers))
	for i, header := range headers {
		_, pref, _ := header.AsPanel().Sizes(geom.Size{})
		width := pref.Width + d.table.Padding.Left + d.table.Padding.Right
		d.table.Columns[i] = unison.ColumnInfo{
			ID:      i,
			Current: width,
			Minimum: width,
			Maximum: 10000,
		}
	}
	header := unison.NewTableHeader(d.table, headers...)
	header.BackgroundInk = colors.Header
	header.InteriorDividerColor = colors.Header
	header.SetBorder(header.HeaderBorder)
	announceColumnsOnFocus(d.table, header)
	d.table.KeyDownCallback = func(keyCode unison.KeyCode, mods mod.Modifiers, repeat bool) bool {
		if noModifiersDown(mods) && (keyCode == unison.KeyBackspace || keyCode == unison.KeyDelete) {
			d.table.PerformCmd(d.table, unison.DeleteItemID)
			return true
		}
		return d.table.DefaultKeyDown(keyCode, mods, repeat)
	}
	d.InstallCmdHandlers(unison.DeleteItemID,
		func(_ any) bool { return d.table.HasSelection() },
		func(_ any) { d.removeSelection() })

	scroll := unison.NewScrollPanel()
	scroll.SetColumnHeader(header)
	scroll.SetContent(d.table, behavior.Fill, behavior.Fill)
	scroll.SetLayoutData(&unison.FlexLayoutData{
		SizeHint: d.sizeHint(),
		HAlign:   align.Fill,
		VAlign:   align.Fill,
		HGrab:    true,
		VGrab:    true,
	})
	d.AddChild(scroll)

	hint := unison.NewLabel()
	hint.SetTitle(i18n.Text("Select entries and press Delete to remove them."))
	strip := unison.NewPanel()
	strip.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{},
		geom.Insets{Top: 1}, false), unison.NewEmptyBorder(unison.StdInsets())))
	strip.SetLayout(&unison.FlexLayout{Columns: 1})
	strip.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	strip.AddChild(hint)
	d.AddChild(strip)

	d.sync()
	return d
}

// sizeHint returns the room the table asks for however little it holds: enough for shoppingRecordHintRows entries of
// typical width. A window packed around the record would otherwise close up around its header when the record is
// empty or nearly so, since an empty table asks for no height at all.
func (d *shoppingRecordDockable) sizeHint() geom.Size {
	font := unison.DefaultLabelTheme.Font
	var width float32
	for _, sample := range []string{
		jio.Now().String(),
		"Sold Shortsword, Small Knife",
		i18n.Text("Bought"),
		"Shortsword ×1, Small Knife ×2",
		"$239.00 (+11 Silver Eagle, +19 Copper Penny)",
	} {
		width += unison.NewText(sample, &unison.TextDecoration{Font: font}).Width() + d.table.Padding.Left +
			d.table.Padding.Right
	}
	rowHeight := font.LineHeight() + d.table.Padding.Top + d.table.Padding.Bottom
	return geom.NewSize(xmath.Ceil(width), xmath.Ceil(rowHeight*shoppingRecordHintRows))
}

// newShoppingRecordColumnHeader returns a header for one of the record's columns. The record is always shown most
// recent first, so the columns don't sort.
func newShoppingRecordColumnHeader(title string) *unison.DefaultTableColumnHeader[*shoppingRecordRow] {
	header := unison.NewTableColumnHeader[*shoppingRecordRow](title, "", nil)
	header.Text = unison.NewSmallCapsText(title, &header.TextDecoration)
	header.SetSortState(unison.SortState{Order: -1})
	return header
}

// TitleIcon implements unison.Dockable.
func (d *shoppingRecordDockable) TitleIcon(suggestedSize geom.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  svg.ReleaseNotes,
		Size: suggestedSize,
	}
}

// Title implements unison.Dockable.
func (d *shoppingRecordDockable) Title() string {
	return i18n.Text("Shopping Record: %s", d.sheet.Entity().Profile.Name)
}

// Tooltip implements unison.Dockable.
func (d *shoppingRecordDockable) Tooltip() string {
	return ""
}

// Modified implements unison.Dockable. Removals are made to the sheet at once, so the record never holds unsaved
// changes of its own.
func (d *shoppingRecordDockable) Modified() bool {
	return false
}

// MayAttemptClose implements unison.TabCloser.
func (d *shoppingRecordDockable) MayAttemptClose() bool {
	return true
}

// AttemptClose implements unison.TabCloser.
func (d *shoppingRecordDockable) AttemptClose() bool {
	return AttemptCloseForDockable(d)
}

// CloseWithGroup implements GroupedCloser.
func (d *shoppingRecordDockable) CloseWithGroup(other unison.Paneler) bool {
	return other == d.sheet
}

// UndoManager implements unison.UndoManagerProvider.
func (d *shoppingRecordDockable) UndoManager() *unison.UndoManager {
	return unison.UndoManagerFor(d.sheet)
}

// SheetChanged implements SheetSourceUser.
func (d *shoppingRecordDockable) SheetChanged(sheet *Sheet) {
	if sheet == d.sheet {
		d.sync()
	}
}

// sync brings the rows and the title into line with the sheet. It runs on every change to the sheet, including each
// keystroke typed into it, so the table is touched only when what it shows has changed.
func (d *shoppingRecordDockable) sync() {
	entity := d.sheet.Entity()
	rows := make([]*shoppingRecordRow, len(entity.ShoppingRecord))
	for i, entry := range entity.ShoppingRecord {
		rows[i] = &shoppingRecordRow{id: entry.ID, cells: shoppingRecordCells(entry, entity.SheetSettings)}
	}
	if !slices.EqualFunc(rows, d.table.RootRows(), func(a, b *shoppingRecordRow) bool { return *a == *b }) {
		// The rows are handed to the model rather than to the table, since the table would clear the selection, and
		// the rows that remain are to stay selected.
		d.table.Model.SetRootRows(rows)
		d.table.SyncToModel()
		d.table.SizeColumnsToFit(true)
	}
	if title := d.Title(); title != d.title {
		d.title = title
		UpdateTitleForDockable(d)
	}
}

// removeSelection removes the selected entries from the sheet's shopping record as one undoable edit.
func (d *shoppingRecordDockable) removeSelection() {
	selected := d.table.SelectedRows(false)
	if len(selected) == 0 {
		return
	}
	remove := make(map[tid.TID]bool, len(selected))
	for _, row := range selected {
		remove[row.id] = true
	}
	sheet := d.sheet
	entity := sheet.Entity()
	before := gurps.CloneShoppingRecordList(entity.ShoppingRecord)
	after := slices.DeleteFunc(gurps.CloneShoppingRecordList(entity.ShoppingRecord),
		func(entry *gurps.ShoppingRecordEntry) bool { return remove[entry.ID] })
	if mgr := unison.UndoManagerFor(sheet); mgr != nil {
		name := i18n.Text("Remove Shopping Record Entries")
		if len(before)-len(after) == 1 {
			name = i18n.Text("Remove Shopping Record Entry")
		}
		mgr.Add(&unison.UndoEdit[[]*gurps.ShoppingRecordEntry]{
			ID:       unison.NextUndoID(),
			EditName: name,
			UndoFunc: func(edit *unison.UndoEdit[[]*gurps.ShoppingRecordEntry]) {
				entity.SetShoppingRecord(edit.BeforeData)
				rebuildAsModified(sheet, false)
			},
			RedoFunc: func(edit *unison.UndoEdit[[]*gurps.ShoppingRecordEntry]) {
				entity.SetShoppingRecord(edit.AfterData)
				rebuildAsModified(sheet, false)
			},
			BeforeData: before,
			AfterData:  after,
		})
	}
	entity.SetShoppingRecord(after)
	rebuildAsModified(sheet, false)
}

// shoppingRecordCells returns the text of each cell of the entry's row.
func shoppingRecordCells(entry *gurps.ShoppingRecordEntry, settings *gurps.SheetSettings) [shoppingRecordColumnCount]string {
	var cells [shoppingRecordColumnCount]string
	cells[shoppingRecordWhenColumn] = entry.When.String()
	cells[shoppingRecordTransactionColumn] = entry.Name
	switch entry.Kind {
	case gurps.ShoppingSold:
		cells[shoppingRecordTypeColumn] = i18n.Text("Sold")
	case gurps.ShoppingBought:
		cells[shoppingRecordTypeColumn] = i18n.Text("Bought")
	default:
		cells[shoppingRecordTypeColumn] = string(entry.Kind)
	}
	items := make([]string, len(entry.Items))
	for i, item := range entry.Items {
		items[i] = i18n.Text("%s ×%s", item.Name, item.Quantity.Comma())
	}
	cells[shoppingRecordItemsColumn] = strings.Join(items, ", ")
	exchanged := settings.FormatEquipmentValue(entry.Exchanged)
	if len(entry.Money) == 0 {
		cells[shoppingRecordMoneyColumn] = i18n.Text("$%s", exchanged)
	} else {
		money := make([]string, len(entry.Money))
		for i, one := range entry.Money {
			if one.Quantity < 0 {
				money[i] = i18n.Text("−%s %s", one.Quantity.Abs().Comma(), one.Name)
			} else {
				money[i] = i18n.Text("+%s %s", one.Quantity.Comma(), one.Name)
			}
		}
		cells[shoppingRecordMoneyColumn] = i18n.Text("$%s (%s)", exchanged, strings.Join(money, ", "))
	}
	return cells
}

// shoppingRecordRow is one entry of a shopping record as a table row. Rows are never nested, so the methods that would
// manage a parent or children do nothing.
type shoppingRecordRow struct {
	id    tid.TID
	cells [shoppingRecordColumnCount]string
}

func (r *shoppingRecordRow) CloneForTarget(_ unison.Paneler, _ *shoppingRecordRow) *shoppingRecordRow {
	return nil
}

func (r *shoppingRecordRow) ID() tid.TID {
	return r.id
}

func (r *shoppingRecordRow) Parent() *shoppingRecordRow {
	return nil
}

func (r *shoppingRecordRow) SetParent(_ *shoppingRecordRow) {}

func (r *shoppingRecordRow) CanHaveChildren() bool {
	return false
}

func (r *shoppingRecordRow) Children() []*shoppingRecordRow {
	return nil
}

func (r *shoppingRecordRow) SetChildren(_ []*shoppingRecordRow) {}

func (r *shoppingRecordRow) CellDataForSort(col int) string {
	if col < 0 || col >= len(r.cells) {
		return ""
	}
	return r.cells[col]
}

func (r *shoppingRecordRow) ColumnCell(_, col int, foreground, _ unison.Ink, _, _, _ bool) unison.Paneler {
	label := unison.NewLabel()
	label.OnBackgroundInk = foreground
	label.SetTitle(r.CellDataForSort(col))
	return label
}

func (r *shoppingRecordRow) IsOpen() bool {
	return false
}

func (r *shoppingRecordRow) SetOpen(_ bool) {}
