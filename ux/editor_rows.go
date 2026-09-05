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

	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/weight"
)

// rowDragEditor is implemented by an editor dockable whose rows can be reordered by dragging.
type rowDragEditor interface {
	unison.Paneler
	// reorderRows applies move inside an undo edit named title and rebuilds the content when it reports a change.
	reorderRows(title string, move func() bool)
}

// structuralEditor is an editor dockable whose model is changed in undoable steps, each of which rebuilds the content.
// It is what lets a list panel be shared between editors that keep different models.
type structuralEditor interface {
	rowDragEditor
	// targetManager returns the target manager the editor's widgets are registered with.
	targetManager() *TargetMgr
	// editStructure applies mutate inside an undo edit named title, then rebuilds the content and, when focusKey is not
	// empty, gives the keyboard focus to the widget with that reference key.
	editStructure(title string, mutate func(), focusKey string)
}

// editorRowDragData is the payload of a row drag within an editor. Drop positions are computed from the dragged row's
// siblings, so a row can only be reordered within the list that owns it. Once the drop position is known, move performs
// the reorder in the model.
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

// install records the editor and its content panel, then wires the drag callbacks and the insertion marker to the
// content panel.
func (s *rowDragState) install(editor rowDragEditor, content *unison.Panel) {
	s.editor = editor
	s.content = content
	installPanelDragDrop(content, editorRowDragKey, s.dataDragOver, s.dataDragExit, s.dataDragDrop)
	content.DrawOverCallback = s.drawOver
}

// dataDragOver tracks the pointer during a row drag: it keeps the pointer in view and works out which list and which
// insertion position within it the pointer is over, redrawing when that changes.
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

// dataDragExit clears the drop-target state and removes the insertion marker.
func (s *rowDragState) dataDragExit() {
	s.inDragOver = false
	s.dragInsert = -1
	s.dragTarget = nil
	s.editor.AsPanel().MarkForRedraw()
}

// dataDragDrop hands the drop to the editor that owns the payload, then clears the drop-target state. A payload from
// some other editor, or a drop with no insertion position, is ignored.
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

// moveEntry moves the entry at from to the insertion index to, which may be anywhere from 0 to len(*list). The
// insertion index is a position in the list as it stands before the entry is taken out of it, so a target beyond the
// entry's current position is adjusted down by one, exactly as a drop below a row is. It returns false and leaves the
// list untouched when either index is out of range or the move would leave the entry where it already is.
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

// newEditorSectionHeader returns a bold section title with any number of buttons beside it, in the order given. It asks
// to span two columns, which suits the editors' two-column grids; a single-column parent clamps that to one.
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

// configureEditorRow sets up a panel as one row of an editor list: the standard row insets, a background that
// alternates with the row's position among its siblings so that adjacent rows can be told apart, the given number of
// columns, and a layout that fills the width of the list.
func configureEditorRow(row *unison.Panel, columns int) {
	row.SetBorder(unison.NewEmptyBorder(geom.Insets{
		Top:    unison.StdVSpacing,
		Left:   unison.StdHSpacing,
		Bottom: unison.StdVSpacing,
		Right:  unison.StdHSpacing,
	}))
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
