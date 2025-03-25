// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/fatal"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/tid"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

const noInvertColorsMarker = "no_invert"

var _ unison.TableRowData[*Node[*gurps.Trait]] = &Node[*gurps.Trait]{}

// CellCache holds data for a table row's cell to reduce the need to constantly recreate them.
type CellCache struct {
	Panel unison.Paneler
	Data  gurps.CellData
	Width float32
}

// Matches returns true if the provided width and data match the current contents.
func (c *CellCache) Matches(width float32, data *gurps.CellData) bool {
	return c != nil && c.Panel != nil && c.Width == width && c.Data == *data
}

// Node represents a row in a table.
type Node[T gurps.NodeTypes] struct {
	table      *unison.Table[*Node[T]]
	parent     *Node[T]
	data       T
	dataAsNode gurps.Node[T]
	children   []*Node[T]
	cellCache  []*CellCache
	forPage    bool
}

// NewNode creates a new node for a table.
func NewNode[T gurps.NodeTypes](table *unison.Table[*Node[T]], parent *Node[T], data T, forPage bool) *Node[T] {
	return &Node[T]{
		table:      table,
		parent:     parent,
		data:       data,
		dataAsNode: gurps.AsNode(data),
		cellCache:  make([]*CellCache, len(table.Columns)),
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
		cellCache:  make([]*CellCache, len(like.table.Columns)),
		forPage:    like.forPage,
	}
}

// CloneForTarget implements unison.TableRowData.
func (n *Node[T]) CloneForTarget(target unison.Paneler, newParent *Node[T]) *Node[T] {
	table, ok := target.(*unison.Table[*Node[T]])
	if !ok {
		fatal.IfErr(errs.New("unable to convert to table"))
	}
	var owner gurps.DataOwner
	if provider := DetermineDataOwnerProvider(target); provider != nil {
		owner = provider.DataOwner()
	}
	return NewNode(table, newParent,
		n.dataAsNode.Clone(libraryFileFromTable(n.table), owner, newParent.Data(), false), false)
}

// ID implements unison.TableRowData.
func (n *Node[T]) ID() tid.TID {
	return n.dataAsNode.ID()
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
			n.children[i] = NewNode(n.table, n, one, n.forPage)
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
	var data gurps.CellData
	n.dataAsNode.CellData(n.table.Columns[index].ID, &data)
	s := data.ForSort()
	if gurps.GlobalSettings().General.GroupContainersOnSort && n.dataAsNode.Container() {
		return containerMarker + s
	}
	return s
}

// ColumnCell implements unison.TableRowData.
func (n *Node[T]) ColumnCell(row, col int, foreground, background unison.Ink, _, _, _ bool) unison.Paneler {
	var cellData gurps.CellData
	n.dataAsNode.CellData(n.table.Columns[col].ID, &cellData)
	width := n.table.CellWidth(row, col)
	if n.cellCache[col].Matches(width, &cellData) {
		applyInkRecursively(n.cellCache[col].Panel.AsPanel(), foreground, background)
		return n.cellCache[col].Panel
	}
	c := n.CellFromCellData(&cellData, width, foreground, background)
	n.cellCache[col] = &CellCache{
		Panel: c,
		Data:  cellData,
		Width: width,
	}
	return c
}

func applyInkRecursively(panel *unison.Panel, foreground, background unison.Ink) {
	switch part := panel.Self.(type) {
	case *unison.Markdown:
		if part.OnBackgroundInk != foreground {
			part.OnBackgroundInk = foreground
			part.Rebuild()
		}
		return
	case *unison.Label:
		if part.OnBackgroundInk != foreground {
			if _, exists := part.ClientData()[noInvertColorsMarker]; !exists {
				part.OnBackgroundInk = foreground
				part.SetTitle(part.String())
			}
		}
	case *unison.Tag:
		if part.OnBackgroundInk != background || part.BackgroundInk != foreground {
			if _, exists := part.ClientData()[noInvertColorsMarker]; !exists {
				part.BackgroundInk = foreground
				part.OnBackgroundInk = background
				part.SetTitle(part.Text.String())
			}
		}
	}
	for _, child := range panel.Children() {
		applyInkRecursively(child, foreground, background)
	}
}

// IsOpen implements unison.TableRowData.
func (n *Node[T]) IsOpen() bool {
	return n.dataAsNode.IsOpen()
}

// SetOpen implements unison.TableRowData.
func (n *Node[T]) SetOpen(open bool) {
	wasOpen := n.dataAsNode.IsOpen()
	n.dataAsNode.SetOpen(open)
	if wasOpen != n.dataAsNode.IsOpen() {
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
	for i := range n.table.Columns {
		var data gurps.CellData
		n.dataAsNode.CellData(n.table.Columns[i].ID, &data)
		if data.Type != cell.Tags {
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
		for i := range n.table.Columns {
			if strings.Contains(strings.ToLower(n.CellDataForSort(i)), text) {
				return true
			}
		}
	}
	return false
}

// CellFromCellData creates a new panel for the given cell data.
func (n *Node[T]) CellFromCellData(c *gurps.CellData, width float32, foreground, background unison.Ink) unison.Paneler {
	switch c.Type {
	case cell.Text, cell.Tags:
		return n.createLabelCell(c, width, foreground, background)
	case cell.Toggle:
		return n.createToggleCell(c, foreground)
	case cell.PageRef:
		return n.createPageRefCell(c, foreground)
	case cell.Markdown:
		return n.createMarkdownCell(c, width, foreground)
	default:
		return unison.NewPanel()
	}
}

func (n *Node[T]) createMarkdownCell(c *gurps.CellData, width float32, foreground unison.Ink) unison.Paneler {
	m := unison.NewMarkdown(false)
	if wd, ok := n.table.ClientData()[WorkingDirKey]; ok {
		m.ClientData()[WorkingDirKey] = wd
	}
	if n.forPage {
		adjustMarkdownThemeForPage(m)
	}
	if i, ok := m.OnBackgroundInk.(*unison.IndirectInk); ok {
		i.Target = foreground
	}
	m.SetContent(c.Primary, width)
	return m
}

func (n *Node[T]) createLabelCell(c *gurps.CellData, width float32, foreground, background unison.Ink) unison.Paneler {
	p := unison.NewPanel()
	p.SetLayout(&unison.FlexLayout{
		Columns: 1,
		HAlign:  c.Alignment,
	})
	if c.Secondary != "" {
		outer := unison.NewPanel()
		outer.SetLayout(&unison.FlexLayout{
			Columns:  2,
			HSpacing: unison.StdHSpacing / 2,
		})
		inner := unison.NewPanel()
		inner.SetLayout(&unison.FlexLayout{
			Columns: 1,
			HAlign:  c.Alignment,
		})
		inner.SetLayoutData(&unison.FlexLayoutData{
			HAlign: align.Fill,
			HGrab:  true,
		})
		outer.AddChild(inner)
		button := unison.NewButton()
		button.HideBase = true
		button.DrawableOnlyHMargin = 0
		button.DrawableOnlyVMargin = 0
		button.Font = n.primaryFieldFont()
		baseline := button.Font.Baseline()
		size := max(baseline-2, 6)
		key := "N:" + string(n.ID())
		isClosed := gurps.IsClosed(key)
		var s *unison.SVG
		if isClosed {
			s = svg.NotesExpand
		} else {
			s = svg.NotesCollapse
		}
		button.Drawable = &unison.DrawableSVG{
			SVG:  s,
			Size: unison.NewSize(size, size).Ceil(),
		}
		button.ClickCallback = func() {
			gurps.SetClosedState(key, !gurps.IsClosed(key))
			n.table.SyncToModel()
		}
		if baseline-size > 0 {
			button.SetBorder(unison.NewEmptyBorder(unison.Insets{Top: baseline - size}))
		}
		button.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Start})
		outer.AddChild(button)
		p.AddChild(outer)
		_, prefSize, _ := button.Sizes(unison.Size{})
		n.addLabelCell(c, inner, width-((unison.StdHSpacing/2)+prefSize.Width), c.Primary, c.InlineTag,
			n.primaryFieldFont(), foreground, background, true)
		if !isClosed {
			n.addLabelCell(c, p, width, c.Secondary, "", n.secondaryFieldFont(), foreground, background, false)
		}
	} else {
		n.addLabelCell(c, p, width, c.Primary, c.InlineTag, n.primaryFieldFont(), foreground, background, true)
	}
	tooltip := c.Tooltip
	if c.UnsatisfiedReason != "" {
		tag := unison.NewTag()
		tag.BackgroundInk = unison.ThemeError
		tag.OnBackgroundInk = unison.ThemeOnError
		tag.Font = n.secondaryFieldFont()
		height := tag.Font.LineHeight() - 2
		tag.Drawable = &unison.DrawableSVG{
			SVG:  unison.TriangleExclamationSVG,
			Size: unison.NewSize(height, height),
		}
		tag.SetTitle(i18n.Text("Unsatisfied prerequisite(s)"))
		tag.ClientData()[noInvertColorsMarker] = true
		p.AddChild(tag)
		tooltip = c.UnsatisfiedReason
	}
	if c.TemplateInfo != "" {
		tag := unison.NewTag()
		tag.BackgroundInk = foreground
		tag.OnBackgroundInk = background
		tag.Font = n.secondaryFieldFont()
		height := tag.Font.LineHeight() - 2
		tag.Drawable = &unison.DrawableSVG{
			SVG:  svg.GCSTemplate,
			Size: unison.NewSize(height, height),
		}
		tag.SetTitle(c.TemplateInfo)
		tag.ClientData()[noInvertColorsMarker] = true
		p.AddChild(tag)
	}
	if tooltip != "" {
		p.Tooltip = newWrappedTooltip(tooltip)
	}
	return p
}

func (n *Node[T]) addLabelCell(c *gurps.CellData, parent *unison.Panel, width float32, text, inlineTag string, f unison.Font, foreground, background unison.Ink, primary bool) {
	decoration := &unison.TextDecoration{
		Font:          f,
		StrikeThrough: primary && c.Disabled,
	}
	var tag *unison.Tag
	if inlineTag != "" {
		tag = unison.NewTag()
		tag.BackgroundInk = foreground
		tag.OnBackgroundInk = background
		tag.Font = &unison.DynamicFont{
			Resolver: func() unison.FontDescriptor {
				desc := f.Descriptor()
				desc.Size = max(desc.Size-2, 1)
				return desc
			},
		}
		tag.SetTitle(inlineTag)
		tag.SetEnabled(!c.Dim)
	}
	var lines []*unison.Text
	if width > 0 {
		if tag != nil {
			_, size, _ := tag.Sizes(unison.Size{})
			lines = unison.NewTextWrappedLines(text, decoration, width-(size.Width+unison.StdHSpacing))
			if len(lines) > 1 {
				lines = lines[:1]
				lines = append(lines, unison.NewTextWrappedLines(strings.TrimPrefix(text, lines[0].String()),
					decoration, width)...)
			}
		} else {
			lines = unison.NewTextWrappedLines(text, decoration, width)
		}
	} else {
		lines = unison.NewTextLines(text, decoration)
	}
	for _, line := range lines {
		label := unison.NewLabel()
		label.Font = f
		label.StrikeThrough = primary && c.Disabled
		label.HAlign = c.Alignment
		label.OnBackgroundInk = foreground
		label.SetTitle(line.String())
		label.SetEnabled(!c.Dim)
		if tag != nil {
			wrapper := unison.NewPanel()
			wrapper.SetLayout(&unison.FlexLayout{
				Columns:  2,
				HSpacing: unison.StdHSpacing,
				HAlign:   align.Start,
				VAlign:   align.Middle,
			})
			wrapper.AddChild(label)
			wrapper.AddChild(tag)
			parent.AddChild(wrapper)
			tag = nil
		} else {
			parent.AddChild(label)
		}
	}
}

func (n *Node[T]) createToggleCell(c *gurps.CellData, foreground unison.Ink) unison.Paneler {
	check := unison.NewLabel()
	check.VAlign = align.Start
	font := n.primaryFieldFont()
	fd := font.Descriptor()
	fd.Size -= 2
	check.Font = fd.Font()
	check.SetBorder(unison.NewEmptyBorder(unison.Insets{Top: 1}))
	baseline := font.Baseline()
	if c.Checked {
		check.Drawable = &unison.DrawableSVG{
			SVG:  unison.CheckmarkSVG,
			Size: unison.Size{Width: baseline, Height: baseline},
		}
	}
	check.HAlign = c.Alignment
	check.OnBackgroundInk = foreground
	if c.Tooltip != "" {
		check.Tooltip = newWrappedTooltip(c.Tooltip)
	}
	check.MouseDownCallback = func(_ unison.Point, _, _ int, _ unison.Modifiers) bool {
		c.Checked = !c.Checked
		handleCheck(n.data, check, c.Checked)
		if c.Checked {
			check.Drawable = &unison.DrawableSVG{
				SVG:  unison.CheckmarkSVG,
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
		gurps.EntityFromNode(item).Recalculate()
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
		gurps.EntityFromNode(item).Recalculate()
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
		gurps.EntityFromNode(item).Recalculate()
	case *gurps.Weapon:
		item.Hide = checked
		if mgr := unison.UndoManagerFor(check); mgr != nil {
			owner := unison.AncestorOrSelf[Rebuildable](check)
			mgr.Add(&unison.UndoEdit[*weaponAdjuster]{
				ID:       unison.NextUndoID(),
				EditName: i18n.Text("Toggle Hidden"),
				UndoFunc: func(edit *unison.UndoEdit[*weaponAdjuster]) { edit.BeforeData.Apply() },
				RedoFunc: func(edit *unison.UndoEdit[*weaponAdjuster]) { edit.AfterData.Apply() },
				BeforeData: &weaponAdjuster{
					Owner:  owner,
					Target: item,
					Hide:   !item.Hide,
				},
				AfterData: &weaponAdjuster{
					Owner:  owner,
					Target: item,
					Hide:   item.Hide,
				},
			})
		}
		gurps.EntityFromNode(item).Recalculate()
	}
}

type equipmentAdjuster struct {
	Owner    Rebuildable
	Target   *gurps.Equipment
	Equipped bool
}

func (e *equipmentAdjuster) Apply() {
	e.Target.Equipped = e.Equipped
	gurps.EntityFromNode(e.Target).Recalculate()
	MarkModified(e.Owner)
}

type equipmentModifierAdjuster struct {
	Owner    Rebuildable
	Target   *gurps.EquipmentModifier
	Disabled bool
}

func (e *equipmentModifierAdjuster) Apply() {
	e.Target.Disabled = e.Disabled || e.Target.Container()
	gurps.EntityFromNode(e.Target).Recalculate()
	MarkModified(e.Owner)
}

type traitModifierAdjuster struct {
	Owner    Rebuildable
	Target   *gurps.TraitModifier
	Disabled bool
}

func (t *traitModifierAdjuster) Apply() {
	t.Target.Disabled = t.Disabled || t.Target.Container()
	gurps.EntityFromNode(t.Target).Recalculate()
	MarkModified(t.Owner)
}

type weaponAdjuster struct {
	Owner  Rebuildable
	Target *gurps.Weapon
	Hide   bool
}

func (w *weaponAdjuster) Apply() {
	w.Target.Hide = w.Hide
	gurps.EntityFromNode(w.Target).Recalculate()
	MarkModified(w.Owner)
}

func convertLinksForPageRef(in string) (string, *unison.SVG) {
	switch {
	case unison.HasURLPrefix(in):
		return i18n.Text("link"), svg.Link
	case strings.HasPrefix(in, "md:"):
		return i18n.Text("md"), svg.MarkdownFile
	default:
		return in, nil
	}
}

func (n *Node[T]) createPageRefCell(c *gurps.CellData, foreground unison.Ink) unison.Paneler {
	var title, tooltip string
	var icon *unison.DrawableSVG
	font := n.primaryFieldFont()
	parts := strings.FieldsFunc(c.Primary, func(ch rune) bool { return ch == ',' || ch == ';' })
	switch len(parts) {
	case 0:
	case 1:
		var img *unison.SVG
		title, img = convertLinksForPageRef(parts[0])
		if img != nil {
			title = ""
			height := font.Baseline()
			icon = &unison.DrawableSVG{
				SVG:  img,
				Size: unison.NewSize(height, height).Ceil(),
			}
			tooltip = parts[0]
		}
	default:
		title, _ = convertLinksForPageRef(parts[0])
		title += "+"
		tooltip = strings.Join(parts, "\n")
	}
	theme := unison.DefaultLinkTheme
	theme.OnBackgroundInk = foreground
	theme.Font = font
	link := unison.NewLink(title, tooltip, "", theme, func(_ unison.Paneler, _ string) {
		list := ExtractPageReferences(c.Primary)
		if len(list) != 0 {
			OpenPageReference(list[0], c.Secondary, nil)
		}
	})
	link.VAlign = align.Start
	if icon != nil {
		link.Drawable = icon
	}
	if tooltip != "" {
		link.Tooltip = newWrappedTooltip(tooltip)
	}
	link.SetEnabled(!c.Dim && (title != "" || icon != nil))
	return link
}

func (n *Node[T]) primaryFieldFont() unison.Font {
	if n.forPage {
		return fonts.PageFieldPrimary
	}
	return unison.FieldFont
}

func (n *Node[T]) secondaryFieldFont() unison.Font {
	if n.forPage {
		return fonts.PageFieldSecondary
	}
	return fonts.FieldSecondary
}

// FindRowIndexByID returns the row index of the row with the given ID in the given table.
func FindRowIndexByID[T gurps.NodeTypes](table *unison.Table[*Node[T]], id tid.TID) int {
	_, i := rowIndex(id, 0, table.RootRows())
	return i
}

func rowIndex[T gurps.NodeTypes](id tid.TID, startIndex int, rows []*Node[T]) (updatedStartIndex, result int) {
	for _, row := range rows {
		if id == row.dataAsNode.ID() {
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
			AbsorbFunc: func(_ *unison.UndoEdit[*TableUndoEditData[T]], _ unison.Undoable) bool { return false },
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
	selMap := make(map[tid.TID]bool)
	for _, item := range items {
		selMap[gurps.AsNode(item).ID()] = true
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
