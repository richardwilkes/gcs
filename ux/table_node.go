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

package ux

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/theme"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/slices"
)

const invertColorsMarker = "invert"

var _ unison.TableRowData[*Node[*gurps.Trait]] = &Node[*gurps.Trait]{}

// Node represents a row in a table.
type Node[T gurps.NodeTypes] struct {
	table      *unison.Table[*Node[T]]
	parent     *Node[T]
	data       T
	dataAsNode gurps.Node[T]
	children   []*Node[T]
	cellCache  []*CellCache
	colMap     map[int]int
	forPage    bool
}

// NewNode creates a new node for a table.
func NewNode[T gurps.NodeTypes](table *unison.Table[*Node[T]], parent *Node[T], colMap map[int]int, data T, forPage bool) *Node[T] {
	return &Node[T]{
		table:      table,
		parent:     parent,
		data:       data,
		dataAsNode: gurps.AsNode(data),
		cellCache:  make([]*CellCache, len(colMap)),
		colMap:     colMap,
		forPage:    forPage,
	}
}

// NewNodeLike creates a new node for a table based on the characteristics of an existing node in that table.
func NewNodeLike[T gurps.NodeTypes](like *Node[T], data T) *Node[T] {
	return &Node[T]{
		table:      like.table,
		parent:     like.parent,
		data:       data,
		dataAsNode: gurps.AsNode(data),
		cellCache:  make([]*CellCache, len(like.colMap)),
		colMap:     like.colMap,
		forPage:    like.forPage,
	}
}

// CloneForTarget implements unison.TableRowData.
func (n *Node[T]) CloneForTarget(target unison.Paneler, newParent *Node[T]) *Node[T] {
	table, ok := target.(*unison.Table[*Node[T]])
	if !ok {
		jot.Fatal(1, "unable to convert to table")
	}
	if provider := unison.AncestorOrSelf[gurps.EntityProvider](target); provider != nil {
		return NewNode[T](table, newParent, n.colMap, n.dataAsNode.Clone(provider.Entity(), newParent.Data(), false),
			n.forPage)
	}
	jot.Fatal(1, "unable to locate entity provider")
	return nil // Never reaches here
}

// UUID implements unison.TableRowData.
func (n *Node[T]) UUID() uuid.UUID {
	return n.dataAsNode.UUID()
}

// Parent implements unison.TableRowData.
func (n *Node[T]) Parent() *Node[T] {
	return n.parent
}

// SetParent implements unison.TableRowData.
func (n *Node[T]) SetParent(parent *Node[T]) {
	n.dataAsNode.SetParent(parent.Data())
}

// CanHaveChildren implements unison.TableRowData.
func (n *Node[T]) CanHaveChildren() bool {
	return n.dataAsNode.Container()
}

// Children implements unison.TableRowData.
func (n *Node[T]) Children() []*Node[T] {
	if n.dataAsNode.Container() && n.children == nil {
		children := n.dataAsNode.NodeChildren()
		n.children = make([]*Node[T], len(children))
		for i, one := range children {
			n.children[i] = NewNode[T](n.table, n, n.colMap, one, n.forPage)
		}
	}
	return n.children
}

// SetChildren implements unison.TableRowData.
func (n *Node[T]) SetChildren(children []*Node[T]) {
	if n.dataAsNode.Container() {
		n.dataAsNode.SetChildren(ExtractNodeDataFromList(children))
		n.children = nil
	}
}

// CellDataForSort implements unison.TableRowData.
func (n *Node[T]) CellDataForSort(index int) string {
	if column, exists := n.colMap[index]; exists {
		var data gurps.CellData
		n.dataAsNode.CellData(column, &data)
		return data.ForSort()
	}
	return ""
}

// ColumnCell implements unison.TableRowData.
func (n *Node[T]) ColumnCell(row, col int, foreground, _ unison.Ink, _, _, _ bool) unison.Paneler {
	var cellData gurps.CellData
	if column, exists := n.colMap[col]; exists {
		n.dataAsNode.CellData(column, &cellData)
	}
	width := n.table.CellWidth(row, col)
	if n.cellCache[col].Matches(width, &cellData) {
		applyForegroundInkRecursively(n.cellCache[col].Panel.AsPanel(), foreground)
		return n.cellCache[col].Panel
	}
	cell := n.CellFromCellData(&cellData, width, foreground)
	n.cellCache[col] = &CellCache{
		Panel: cell,
		Data:  cellData,
		Width: width,
	}
	return cell
}

func applyForegroundInkRecursively(panel *unison.Panel, foreground unison.Ink) {
	if label, ok := panel.Self.(*unison.Label); ok {
		if _, exists := label.ClientData()[invertColorsMarker]; !exists {
			label.OnBackgroundInk = foreground
		}
	}
	for _, child := range panel.Children() {
		applyForegroundInkRecursively(child, foreground)
	}
}

// IsOpen implements unison.TableRowData.
func (n *Node[T]) IsOpen() bool {
	return n.dataAsNode.Container() && n.dataAsNode.Open()
}

// SetOpen implements unison.TableRowData.
func (n *Node[T]) SetOpen(open bool) {
	if n.dataAsNode.Container() && open != n.dataAsNode.Open() {
		n.dataAsNode.SetOpen(open)
		n.table.SyncToModel()
	}
}

// Data returns the underlying data object.
func (n *Node[T]) Data() T {
	if n == nil {
		var zero T
		return zero
	}
	return n.data
}

// HasTag returns true if the specified tag is present on the node. An empty tag will match all nodes.
func (n *Node[T]) HasTag(tag string) bool {
	if tag == "" {
		return true
	}
	if tagListable, ok := any(n.Data()).(interface{ TagList() []string }); ok {
		for _, one := range tagListable.TagList() {
			if strings.EqualFold(tag, one) {
				return true
			}
		}
	}
	return false
}

// PartialMatchExceptTag returns true if the specified text is present in the node's displayable columns other than the
// the tags column. An empty text will match all nodes.
func (n *Node[T]) PartialMatchExceptTag(text string) bool {
	if text == "" {
		return true
	}
	text = strings.ToLower(text)
	for _, column := range n.colMap {
		var data gurps.CellData
		n.dataAsNode.CellData(column, &data)
		if data.Type != gurps.Tags {
			if strings.Contains(strings.ToLower(data.ForSort()), text) {
				return true
			}
		}
	}
	return false
}

// Match looks for the text in the node and return true if it is present. Note that calls to this method should always
// pass in text that has already been run through strings.ToLower().
func (n *Node[T]) Match(text string) bool {
	if text != "" {
		count := len(n.colMap)
		for i := 0; i < count; i++ {
			if strings.Contains(strings.ToLower(n.CellDataForSort(i)), text) {
				return true
			}
		}
	}
	return false
}

// CellFromCellData creates a new panel for the given cell data.
func (n *Node[T]) CellFromCellData(c *gurps.CellData, width float32, foreground unison.Ink) unison.Paneler {
	switch c.Type {
	case gurps.Text, gurps.Tags:
		return n.createLabelCell(c, width, foreground)
	case gurps.Toggle:
		return n.createToggleCell(c, foreground)
	case gurps.PageRef:
		return n.createPageRefCell(c, foreground)
	default:
		return unison.NewPanel()
	}
}

func (n *Node[T]) createLabelCell(c *gurps.CellData, width float32, foreground unison.Ink) unison.Paneler {
	p := unison.NewPanel()
	p.SetLayout(&unison.FlexLayout{
		Columns: 1,
		HAlign:  c.Alignment,
	})
	n.addLabelCell(c, p, width, c.Primary, n.primaryFieldFont(), foreground, true)
	if c.Secondary != "" {
		n.addLabelCell(c, p, width, c.Secondary, n.secondaryFieldFont(), foreground, false)
	}
	tooltip := c.Tooltip
	if c.UnsatisfiedReason != "" {
		label := unison.NewLabel()
		label.Font = n.secondaryFieldFont()
		height := label.Font.LineHeight()
		label.Drawable = &unison.DrawableSVG{
			SVG:  unison.TriangleExclamationSVG(),
			Size: unison.NewSize(height, height),
		}
		label.Text = i18n.Text("Unsatisfied prerequisite(s)")
		label.HAlign = c.Alignment
		label.VAlign = unison.MiddleAlignment
		label.ClientData()[invertColorsMarker] = true
		label.OnBackgroundInk = unison.OnErrorColor
		label.SetBorder(unison.NewEmptyBorder(unison.Insets{
			Left:  4,
			Right: 4,
		}))
		label.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
			gc.DrawRect(rect, unison.ErrorColor.Paint(gc, rect, unison.Fill))
			label.DefaultDraw(gc, rect)
		}
		p.AddChild(label)
		tooltip = c.UnsatisfiedReason
	}
	if c.TemplateInfo != "" {
		label := unison.NewLabel()
		label.Font = n.secondaryFieldFont()
		height := label.Font.LineHeight()
		label.Drawable = &unison.DrawableSVG{
			SVG:  svg.GCSTemplate,
			Size: unison.NewSize(height, height),
		}
		label.Text = c.TemplateInfo
		label.HAlign = c.Alignment
		label.VAlign = unison.MiddleAlignment
		label.ClientData()[invertColorsMarker] = true
		label.OnBackgroundInk = theme.OnMarkerColor
		label.SetBorder(unison.NewEmptyBorder(unison.Insets{
			Left:  4,
			Right: 4,
		}))
		label.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
			gc.DrawRect(rect, theme.MarkerColor.Paint(gc, rect, unison.Fill))
			label.DefaultDraw(gc, rect)
		}
		p.AddChild(label)
	}
	if tooltip != "" {
		p.Tooltip = unison.NewTooltipWithText(txt.Wrap("", tooltip, 120))
	}
	return p
}

func (n *Node[T]) addLabelCell(c *gurps.CellData, parent *unison.Panel, width float32, text string, f unison.Font, foreground unison.Ink, primary bool) {
	decoration := &unison.TextDecoration{
		Font:          f,
		StrikeThrough: primary && c.Disabled,
	}
	var lines []*unison.Text
	if width > 0 {
		lines = unison.NewTextWrappedLines(text, decoration, width)
	} else {
		lines = unison.NewTextLines(text, decoration)
	}
	for _, line := range lines {
		label := unison.NewLabel()
		label.Text = line.String()
		label.Font = f
		label.StrikeThrough = primary && c.Disabled
		label.HAlign = c.Alignment
		label.OnBackgroundInk = foreground
		label.SetEnabled(!c.Dim)
		parent.AddChild(label)
	}
}

func (n *Node[T]) createToggleCell(c *gurps.CellData, foreground unison.Ink) unison.Paneler {
	check := unison.NewLabel()
	check.VAlign = unison.StartAlignment
	check.Font = n.primaryFieldFont()
	check.SetBorder(unison.NewEmptyBorder(unison.Insets{Top: 1}))
	baseline := check.Font.Baseline()
	if c.Checked {
		check.Drawable = &unison.DrawableSVG{
			SVG:  svg.Checkmark,
			Size: unison.Size{Width: baseline, Height: baseline},
		}
	}
	check.HAlign = c.Alignment
	check.OnBackgroundInk = foreground
	if c.Tooltip != "" {
		check.Tooltip = unison.NewTooltipWithText(c.Tooltip)
	}
	check.MouseDownCallback = func(where unison.Point, button, clickCount int, mod unison.Modifiers) bool {
		c.Checked = !c.Checked
		handleCheck(n.data, check, c.Checked)
		if c.Checked {
			check.Drawable = &unison.DrawableSVG{
				SVG:  svg.Checkmark,
				Size: unison.Size{Width: baseline, Height: baseline},
			}
		} else {
			check.Drawable = nil
		}
		check.MarkForLayoutAndRedraw()
		MarkModified(check)
		return true
	}
	return check
}

func handleCheck(data any, check unison.Paneler, checked bool) {
	switch item := data.(type) {
	case *gurps.Equipment:
		item.Equipped = checked
		if mgr := unison.UndoManagerFor(check); mgr != nil {
			owner := unison.AncestorOrSelf[Rebuildable](check)
			mgr.Add(&unison.UndoEdit[*equipmentAdjuster]{
				ID:       unison.NextUndoID(),
				EditName: i18n.Text("Toggle Equipped"),
				UndoFunc: func(edit *unison.UndoEdit[*equipmentAdjuster]) { edit.BeforeData.Apply() },
				RedoFunc: func(edit *unison.UndoEdit[*equipmentAdjuster]) { edit.AfterData.Apply() },
				BeforeData: &equipmentAdjuster{
					Owner:    owner,
					Target:   item,
					Equipped: !item.Equipped,
				},
				AfterData: &equipmentAdjuster{
					Owner:    owner,
					Target:   item,
					Equipped: item.Equipped,
				},
			})
		}
		if item.Entity != nil {
			item.Entity.Recalculate()
		}
	case *gurps.TraitModifier:
		item.Disabled = !checked
		if mgr := unison.UndoManagerFor(check); mgr != nil {
			owner := unison.AncestorOrSelf[Rebuildable](check)
			mgr.Add(&unison.UndoEdit[*traitModifierAdjuster]{
				ID:       unison.NextUndoID(),
				EditName: i18n.Text("Toggle Trait Modifier"),
				UndoFunc: func(edit *unison.UndoEdit[*traitModifierAdjuster]) { edit.BeforeData.Apply() },
				RedoFunc: func(edit *unison.UndoEdit[*traitModifierAdjuster]) { edit.AfterData.Apply() },
				BeforeData: &traitModifierAdjuster{
					Owner:    owner,
					Target:   item,
					Disabled: !item.Disabled,
				},
				AfterData: &traitModifierAdjuster{
					Owner:    owner,
					Target:   item,
					Disabled: item.Disabled,
				},
			})
		}
		if item.Entity != nil {
			item.Entity.Recalculate()
		}
	case *gurps.EquipmentModifier:
		item.Disabled = !checked
		if mgr := unison.UndoManagerFor(check); mgr != nil {
			owner := unison.AncestorOrSelf[Rebuildable](check)
			mgr.Add(&unison.UndoEdit[*equipmentModifierAdjuster]{
				ID:       unison.NextUndoID(),
				EditName: i18n.Text("Toggle Equipment Modifier"),
				UndoFunc: func(edit *unison.UndoEdit[*equipmentModifierAdjuster]) { edit.BeforeData.Apply() },
				RedoFunc: func(edit *unison.UndoEdit[*equipmentModifierAdjuster]) { edit.AfterData.Apply() },
				BeforeData: &equipmentModifierAdjuster{
					Owner:    owner,
					Target:   item,
					Disabled: !item.Disabled,
				},
				AfterData: &equipmentModifierAdjuster{
					Owner:    owner,
					Target:   item,
					Disabled: item.Disabled,
				},
			})
		}
		if item.Entity != nil {
			item.Entity.Recalculate()
		}
	}
}

type equipmentAdjuster struct {
	Owner    Rebuildable
	Target   *gurps.Equipment
	Equipped bool
}

func (a *equipmentAdjuster) Apply() {
	a.Target.Equipped = a.Equipped
	if a.Target.Entity != nil {
		a.Target.Entity.Recalculate()
	}
	MarkModified(a.Owner)
}

type equipmentModifierAdjuster struct {
	Owner    Rebuildable
	Target   *gurps.EquipmentModifier
	Disabled bool
}

func (a *equipmentModifierAdjuster) Apply() {
	a.Target.Disabled = a.Disabled || a.Target.Container()
	if a.Target.Entity != nil {
		a.Target.Entity.Recalculate()
	}
	MarkModified(a.Owner)
}

type traitModifierAdjuster struct {
	Owner    Rebuildable
	Target   *gurps.TraitModifier
	Disabled bool
}

func (a *traitModifierAdjuster) Apply() {
	a.Target.Disabled = a.Disabled || a.Target.Container()
	if a.Target.Entity != nil {
		a.Target.Entity.Recalculate()
	}
	MarkModified(a.Owner)
}

func convertLinksForPageRef(in string) (string, *unison.SVG) {
	lower := strings.ToLower(in)
	switch {
	case strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://"):
		return i18n.Text("link"), svg.Link
	case strings.HasPrefix(lower, "md:"):
		return i18n.Text("md"), svg.MarkdownFile
	default:
		return in, nil
	}
}

func (n *Node[T]) createPageRefCell(c *gurps.CellData, foreground unison.Ink) unison.Paneler {
	label := unison.NewLabel()
	label.VAlign = unison.StartAlignment
	label.Font = n.primaryFieldFont()
	label.OnBackgroundInk = foreground
	label.SetEnabled(!c.Dim)
	parts := strings.FieldsFunc(c.Primary, func(ch rune) bool { return ch == ',' || ch == ';' })
	switch len(parts) {
	case 0:
	case 1:
		var img *unison.SVG
		label.Text, img = convertLinksForPageRef(parts[0])
		if img != nil {
			label.Text = ""
			height := label.Font.Baseline()
			size := unison.NewSize(height, height)
			size.GrowToInteger()
			label.Drawable = &unison.DrawableSVG{
				SVG:  img,
				Size: size,
			}
			label.Tooltip = unison.NewTooltipWithText(parts[0])
		}
	default:
		label.Text, _ = convertLinksForPageRef(parts[0])
		label.Text += "+"
		label.Tooltip = unison.NewTooltipWithText(strings.Join(parts, "\n"))
	}
	if label.Text != "" || label.Drawable != nil {
		over := false
		pressed := false
		label.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
			if over {
				var fg, bg *unison.ThemeColor
				if pressed {
					fg = theme.OnLinkPressedColor
					bg = theme.LinkPressedColor
				} else {
					fg = theme.OnLinkColor
					bg = theme.LinkColor
				}
				gc.DrawRect(rect, bg.Paint(gc, rect, unison.Fill))
				save := label.OnBackgroundInk
				label.OnBackgroundInk = fg
				label.DefaultDraw(gc, rect)
				label.OnBackgroundInk = save
			} else {
				label.DefaultDraw(gc, rect)
			}
		}
		label.MouseEnterCallback = func(where unison.Point, mod unison.Modifiers) bool {
			over = true
			label.MarkForRedraw()
			return true
		}
		label.MouseExitCallback = func() bool {
			over = false
			label.MarkForRedraw()
			return true
		}
		label.MouseDownCallback = func(where unison.Point, button, clickCount int, mod unison.Modifiers) bool {
			pressed = label.ContentRect(true).ContainsPoint(where)
			label.MarkForRedraw()
			return true
		}
		label.MouseDragCallback = func(where unison.Point, button int, mod unison.Modifiers) bool {
			in := label.ContentRect(true).ContainsPoint(where)
			if pressed != in {
				pressed = in
				label.MarkForRedraw()
			}
			return true
		}
		label.MouseUpCallback = func(where unison.Point, button int, mod unison.Modifiers) bool {
			if over = label.ContentRect(true).ContainsPoint(where); over {
				list := ExtractPageReferences(c.Primary)
				if len(list) != 0 {
					unison.InvokeTaskAfter(
						func() { OpenPageReference(label.Window(), list[0], c.Secondary, nil) },
						time.Millisecond)
				}
			}
			pressed = false
			label.MarkForRedraw()
			return true
		}
	}
	return label
}

func (n *Node[T]) primaryFieldFont() unison.Font {
	if n.forPage {
		return theme.PageFieldPrimaryFont
	}
	return unison.FieldFont
}

func (n *Node[T]) secondaryFieldFont() unison.Font {
	if n.forPage {
		return theme.PageFieldSecondaryFont
	}
	return theme.FieldSecondaryFont
}

// FindRowIndexByID returns the row index of the row with the given ID in the given table.
func FindRowIndexByID[T gurps.NodeTypes](table *unison.Table[*Node[T]], id uuid.UUID) int {
	_, i := rowIndex(id, 0, table.RootRows())
	return i
}

func rowIndex[T gurps.NodeTypes](id uuid.UUID, startIndex int, rows []*Node[T]) (updatedStartIndex, result int) {
	for _, row := range rows {
		if id == row.dataAsNode.UUID() {
			return 0, startIndex
		}
		startIndex++
		if row.IsOpen() {
			if startIndex, result = rowIndex(id, startIndex, row.Children()); result != -1 {
				return 0, result
			}
		}
	}
	return startIndex, -1
}

// InsertItems into a table.
func InsertItems[T gurps.NodeTypes](owner Rebuildable, table *unison.Table[*Node[T]], topList func() []T, setTopList func([]T), rowData func(table *unison.Table[*Node[T]]) []*Node[T], items ...T) {
	if len(items) == 0 {
		return
	}
	var undo *unison.UndoEdit[*TableUndoEditData[T]]
	mgr := unison.UndoManagerFor(table)
	if mgr != nil {
		undo = &unison.UndoEdit[*TableUndoEditData[T]]{
			ID:         unison.NextUndoID(),
			EditName:   fmt.Sprintf(i18n.Text("Insert %s"), gurps.AsNode(items[0]).Kind()),
			UndoFunc:   func(e *unison.UndoEdit[*TableUndoEditData[T]]) { e.BeforeData.Apply() },
			RedoFunc:   func(e *unison.UndoEdit[*TableUndoEditData[T]]) { e.AfterData.Apply() },
			AbsorbFunc: func(e *unison.UndoEdit[*TableUndoEditData[T]], other unison.Undoable) bool { return false },
			BeforeData: NewTableUndoEditData(table),
		}
	}
	var target, zero T
	i := table.FirstSelectedRowIndex()
	if i != -1 {
		row := table.RowFromIndex(i)
		if target = row.Data(); target != zero {
			if row.CanHaveChildren() {
				// Target is container, append to end of that container
				SetParents(items, target)
				row.dataAsNode.SetChildren(append(row.dataAsNode.NodeChildren(), items...))
			} else {
				// Target isn't a container. If it has a parent, insert after the target within that parent.
				parent := row.Parent()
				if parentData := parent.Data(); parentData != zero {
					SetParents(items, parentData)
					children := parent.dataAsNode.NodeChildren()
					parent.dataAsNode.SetChildren(slices.Insert(children, slices.Index(children, target)+1, items...))
				} else {
					// Otherwise, insert after the target within the top-level list.
					SetParents(items, zero)
					list := topList()
					setTopList(slices.Insert(list, slices.Index(list, target)+1, items...))
				}
			}
		}
	}
	if target == zero {
		// There was no selection, so append to the end of the top-level list.
		SetParents(items, zero)
		setTopList(append(topList(), items...))
	}
	MarkModified(table)
	table.SetRootRows(rowData(table))
	table.ValidateScrollRoot()
	table.RequestFocus()
	selMap := make(map[uuid.UUID]bool)
	for _, item := range items {
		selMap[gurps.AsNode(item).UUID()] = true
	}
	table.SetSelectionMap(selMap)
	table.ScrollRowCellIntoView(table.LastSelectedRowIndex(), 0)
	table.ScrollRowCellIntoView(table.FirstSelectedRowIndex(), 0)
	if mgr != nil && undo != nil {
		undo.AfterData = NewTableUndoEditData(table)
		mgr.Add(undo)
	}
	owner.Rebuild(true)
}

// SetParents of each item.
func SetParents[T gurps.NodeTypes](items []T, parent T) {
	for _, item := range items {
		gurps.AsNode(item).SetParent(parent)
	}
}

// ExtractNodeDataFromList returns the underlying node data.
func ExtractNodeDataFromList[T gurps.NodeTypes](list []*Node[T]) []T {
	dataList := make([]T, 0, len(list))
	for _, child := range list {
		dataList = append(dataList, child.data)
	}
	return dataList
}
