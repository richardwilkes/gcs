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

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/role"
	"github.com/richardwilkes/unison/enums/weight"
)

// rowDragEditor is implemented by an editor dockable whose rows can be reordered by dragging.
type rowDragEditor interface {
	unison.Paneler
	// reorderRows applies move inside an undo edit named title and rebuilds the content when it reports a change.
	reorderRows(title string, move func() bool)
}

// structuralEditor is an editor dockable whose model is changed in undoable steps, each of which rebuilds the content,
// letting a list panel be shared between editors that keep different models.
type structuralEditor interface {
	rowDragEditor
	targetManager() *TargetMgr
	// editStructure applies mutate inside an undo edit named title, then rebuilds the content and, when focusKey is not
	// empty, gives the keyboard focus to the widget with that reference key.
	editStructure(title string, mutate func(), focusKey string)
}

// editorRowDragData is the payload of a row drag within an editor. Drop positions are computed from the dragged row's
// siblings, so a row can only be reordered within the list that owns it.
type editorRowDragData struct {
	editor rowDragEditor
	row    *unison.Panel
	title  string
	move   func(to int) bool
}

// rowDragState holds the drop-target state of an in-progress row drag and draws the insertion marker. An editor embeds
// it and wires its methods to installPanelDragDrop and the content panel's DrawOverCallback by calling install.
type rowDragState struct {
	editor     rowDragEditor
	content    *unison.Panel
	dragTarget *unison.Panel
	dragInsert int
	inDragOver bool
}

func (s *rowDragState) install(editor rowDragEditor, content *unison.Panel) {
	s.editor = editor
	s.content = content
	installPanelDragDrop(content, editorRowDragKey, s.dataDragOver, s.dataDragExit, s.dataDragDrop)
	content.DrawOverCallback = s.drawOver
}

// dataDragOver keeps the pointer in view and works out which list, and which insertion position within it, the pointer
// is over, redrawing when that changes.
func (s *rowDragState) dataDragOver(where geom.Point, data any) bool {
	s.content.ScrollRectIntoView(geom.NewRect(where.X, where.Y-16, 1, 1))
	s.content.ScrollRectIntoView(geom.NewRect(where.X, where.Y+16, 1, 1))

	prevInDragOver := s.inDragOver
	dragInsert := s.dragInsert
	dragTarget := s.dragTarget
	s.inDragOver = false
	s.dragInsert = -1
	s.dragTarget = nil
	if dd, ok := data.(*editorRowDragData); ok && dd.editor == s.editor {
		if parent := dd.row.Parent(); parent != nil {
			where = parent.PointFromRoot(s.content.PointToRoot(where))
			for i, child := range parent.Children() {
				rect := child.FrameRect()
				if where.In(rect) {
					s.dragTarget = parent
					if rect.CenterY() <= where.Y {
						s.dragInsert = i + 1
					} else {
						s.dragInsert = i
					}
					s.inDragOver = true
					break
				}
			}
		}
	}
	if prevInDragOver != s.inDragOver || dragInsert != s.dragInsert || dragTarget != s.dragTarget {
		s.editor.AsPanel().MarkForRedraw()
	}
	return true
}

func (s *rowDragState) dataDragExit() {
	s.inDragOver = false
	s.dragInsert = -1
	s.dragTarget = nil
	s.editor.AsPanel().MarkForRedraw()
}

// dataDragDrop hands the drop to the editor that owns the payload, ignoring a payload from another editor or a drop
// with no insertion position.
func (s *rowDragState) dataDragDrop(_ geom.Point, data any) {
	if s.inDragOver && s.dragInsert != -1 {
		if dd, ok := data.(*editorRowDragData); ok && dd.editor == s.editor {
			insert := s.dragInsert
			dd.editor.reorderRows(dd.title, func() bool { return dd.move(insert) })
		}
	}
	s.dataDragExit()
}

// drawOver draws the insertion marker: a line across the target list at the current insertion position.
func (s *rowDragState) drawOver(gc *unison.Canvas, rect geom.Rect) {
	if s.inDragOver && s.dragInsert != -1 {
		children := s.dragTarget.Children()
		var y float32
		if s.dragInsert < len(children) {
			y = children[s.dragInsert].FrameRect().Y
		} else {
			y = children[len(children)-1].FrameRect().Bottom()
		}
		pt := s.content.PointFromRoot(s.dragTarget.PointToRoot(geom.Point{Y: y}))
		paint := unison.ThemeWarning.Paint(gc, rect, paintstyle.Stroke)
		paint.SetStrokeWidth(2)
		r := s.content.RectFromRoot(s.dragTarget.RectToRoot(s.dragTarget.ContentRect(false)))
		gc.DrawLine(geom.NewPoint(r.X, pt.Y), geom.NewPoint(r.Right(), pt.Y), paint)
	}
}

// moveEntry moves the entry at from to the insertion index to, which may be anywhere from 0 to len(*list) and is a
// position in the list as it stands before the entry is taken out of it, so a target beyond the entry's current
// position is adjusted down by one. It returns false and leaves the list untouched when either index is out of range or
// the move would leave the entry where it already is.
func moveEntry[T any](list *[]T, from, to int) bool {
	n := len(*list)
	if from < 0 || from >= n || to < 0 || to > n {
		return false
	}
	if to > from {
		to--
	}
	if to == from {
		return false
	}
	entry := (*list)[from]
	*list = slices.Insert(slices.Delete(*list, from, from+1), to, entry)
	return true
}

// initTitledEditorSection sets up a panel as a titled section of an editor, such as its features or prerequisites,
// spanning both columns of the editor's grid, and returns the border that draws its title. The caller then adds the
// section's rows.
func initTitledEditorSection(p unison.Paneler, title string) *TitledBorder {
	panel := p.AsPanel()
	panel.Self = p
	panel.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Fill,
		HGrab:  true,
	})
	border := &TitledBorder{
		Title: title,
		Font:  unison.LabelFont,
	}
	panel.SetBorder(unison.NewCompoundBorder(border, unison.NewEmptyBorder(geom.NewUniformInsets(2))))
	panel.DrawCallback = func(gc *unison.Canvas, rect geom.Rect) {
		gc.DrawRect(rect, unison.ThemeSurface.Paint(gc, rect, paintstyle.Fill))
	}
	return border
}

// sectionToggleKey is the reference key of the title bar of a section that can be collapsed.
const sectionToggleKey = "toggle"

// sectionToggle is the title bar of a section set up by initTitledEditorSection that can be collapsed, laid over the
// strip its border paints the title in. A click, Space or Return collapses or expands the section, as a screen reader
// can ask it to. The section adds it as its first child each time it fills itself, and decides what each state shows.
type sectionToggle struct {
	unison.Panel
	border    *TitledBorder
	collapsed bool
	changed   func()
}

// newSectionToggle lets the section be collapsed by clicking the title its border draws, and returns its title bar,
// which starts out collapsed or not as asked. changed is called after each collapse or expansion. The section's layout
// must be the FlexLayout initTitledEditorSection installs.
func newSectionToggle(section unison.Paneler, border *TitledBorder, collapsed bool, changed func()) *sectionToggle {
	t := &sectionToggle{border: border, collapsed: collapsed, changed: changed}
	t.Self = t
	t.RefKey = sectionToggleKey
	t.SetFocusable(true)
	t.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	panel := section.AsPanel()
	// The title bar takes the first row in place of the strip the border reserved, so the border leaves the strip to
	// the content and blockLayout puts the title bar over the whole of it.
	border.HeadingInContent = true
	var vSpacing float32
	if layout, ok := panel.Layout().(*unison.FlexLayout); ok {
		vSpacing = layout.VSpacing
		panel.SetLayout(&blockLayout{FlexLayout: layout, border: border, heading: t.AsPanel()})
	} else {
		errs.Log(errs.Newf("a section toggle needs a FlexLayout, not a %T", panel.Layout()))
	}
	// The row is short by the spacing that follows it, so the rows below start where the strip would have put them.
	t.SetSizer(func(_ geom.Size) (minSize, prefSize, maxSize geom.Size) {
		height := max(border.TitleHeight()-vSpacing, 0)
		return geom.NewSize(0, height), geom.NewSize(0, height), geom.NewSize(unison.DefaultMaxSize, height)
	})
	// A panel draws its border over its children, so the border draws the chevron and the focus ring.
	panel.SetBorder(&sectionToggleBorder{Border: panel.Border(), toggle: t})
	t.MouseDownCallback = func(_ geom.Point, _, _ int, _ mod.Modifiers) bool { return true }
	t.MouseUpCallback = func(where geom.Point, _ int, _ mod.Modifiers) bool {
		if where.In(t.ContentRect(true)) {
			t.toggle()
		}
		return true
	}
	t.UpdateCursorCallback = func(_ geom.Point) *unison.Cursor { return unison.PointingCursor() }
	t.KeyDownCallback = func(keyCode unison.KeyCode, mods mod.Modifiers, _ bool) bool {
		if mods&mod.NonSticky != 0 ||
			(keyCode != unison.KeySpace && keyCode != unison.KeyReturn && keyCode != unison.KeyNumPadEnter) {
			return false
		}
		t.toggle()
		return true
	}
	t.Accessibility.Role = role.DisclosureTriangle
	t.Accessibility.Name = border.Title
	addAccessibilityCallback(t, func(node *accessibility.Node) {
		node.Expandable = true
		node.Expanded = !t.collapsed
		node.Actions = node.Actions.With(accessibility.Press, accessibility.Expand, accessibility.Collapse)
	})
	t.Accessibility.ActionCallback = func(req accessibility.ActionRequest) bool {
		switch req.Action {
		case accessibility.Press:
			t.toggle()
		case accessibility.Expand, accessibility.Collapse:
			if t.collapsed == (req.Action == accessibility.Expand) {
				t.toggle()
			}
		default:
			return false
		}
		return true
	}
	return t
}

// toggle collapses or expands the section.
func (t *sectionToggle) toggle() {
	t.collapsed = !t.collapsed
	t.MarkForRedraw()
	t.changed()
}

// sectionToggleBorder draws a section's border, then the chevron and focus ring of its title bar.
type sectionToggleBorder struct {
	unison.Border
	toggle *sectionToggle
}

// Draw implements unison.Border.
func (b *sectionToggleBorder) Draw(gc *unison.Canvas, rect geom.Rect) {
	b.Border.Draw(gc, rect)
	strip := b.toggle.FrameRect()
	size := max(b.toggle.border.font().Baseline()-2, 6)
	chevron := &unison.DrawableSVG{SVG: unison.CircledChevronRightSVG, Size: geom.NewSize(size, size)}
	if !b.toggle.collapsed {
		chevron.RotationDegrees = 90
	}
	chevron.DrawInRect(gc, geom.NewRect(strip.X+unison.StdHSpacing, strip.CenterY()-size/2, size, size), nil,
		colors.OnHeader.Paint(gc, strip, paintstyle.Fill))
	if b.toggle.Focused() {
		ring := strip.Inset(geom.NewUniformInsets(0.5))
		gc.DrawRect(ring, unison.ThemeFocus.Paint(gc, ring, paintstyle.Stroke))
	}
}

// newSectionAddButton returns the add button of an editor section, with the given tooltip, which is also what a screen
// reader calls the button. Clicking it calls insert, which adds a new item and a row for it to the head of the
// section's list and reports whether it did so; when it did, the section is laid out again and marked as modified.
func newSectionAddButton(section unison.Paneler, tooltip string, insert func() bool) *unison.Button {
	button := unison.NewSVGButton(unison.CircledAddSVG)
	button.Tooltip = newWrappedTooltip(tooltip)
	button.ClickCallback = func() {
		if insert() {
			MarkRootAncestorForLayoutRecursively(section)
			MarkModified(section)
		}
	}
	return button
}

// newEditorSectionHeader returns a bold section title with the given buttons beside it. It asks to span two columns,
// which suits the editors' two-column grids; a single-column parent clamps that to one.
func newEditorSectionHeader(title, tooltip string, buttons ...*unison.Button) *unison.Panel {
	header := unison.NewPanel()
	for _, button := range buttons {
		header.AddChild(button)
	}
	header.SetLayout(&unison.FlexLayout{Columns: len(buttons) + 1, HSpacing: unison.StdHSpacing})
	header.SetLayoutData(&unison.FlexLayoutData{HSpan: 2, HAlign: align.Fill, HGrab: true})
	header.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: unison.StdVSpacing}))
	label := unison.NewLabel()
	desc := label.Font.Descriptor()
	desc.Weight = weight.Bold
	label.Font = desc.Font()
	label.SetTitle(title)
	label.Tooltip = newWrappedTooltip(tooltip)
	label.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Middle})
	header.AddChild(label)
	return header
}

// editorRowInsets returns the insets of an editor row. The scrollbar is drawn over the content, so an outermost row,
// which sits directly under the bar, gets twice the right inset to keep the bar from obscuring its content.
func editorRowInsets(outermost bool) geom.Insets {
	insets := geom.Insets{
		Top:    unison.StdVSpacing,
		Left:   unison.StdHSpacing,
		Bottom: unison.StdVSpacing,
		Right:  unison.StdHSpacing,
	}
	if outermost {
		insets.Right *= 2
	}
	return insets
}

// configureEditorRow sets up a panel as one row of an editor list, giving it the insets from editorRowInsets, a
// background that alternates with the row's position among its siblings so adjacent rows can be told apart, the given
// number of columns, and a layout that fills the width of the list.
func configureEditorRow(row *unison.Panel, columns int, outermost bool) {
	row.SetBorder(unison.NewEmptyBorder(editorRowInsets(outermost)))
	row.DrawCallback = func(gc *unison.Canvas, rect geom.Rect) {
		var ink unison.Ink
		if row.Parent().IndexOfChild(row)%2 == 1 {
			ink = unison.ThemeSurface
		} else {
			ink = unison.ThemeBelowSurface
		}
		gc.DrawRect(rect, ink.Paint(gc, rect, paintstyle.Fill))
	}
	row.SetLayout(&unison.FlexLayout{
		Columns:  columns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	row.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
}

// newEditorRowButtonColumn returns the column of stacked buttons that follows the drag handle in an editor row.
func newEditorRowButtonColumn(buttons ...*unison.Button) *unison.Panel {
	column := unison.NewPanel()
	column.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	column.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Middle})
	for _, button := range buttons {
		column.AddChild(button)
	}
	return column
}
