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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// TestMoveEntry verifies the insertion-index semantics of moveEntry: the target is a position in the list before the
// entry is removed, so a target beyond the entry is adjusted down by one, out-of-range indexes and no-op moves are
// refused, and a refused move leaves the list alone.
func TestMoveEntry(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name     string
		from, to int
		moved    bool
		expected []string
	}{
		{name: "first to end", from: 0, to: 4, moved: true, expected: []string{"b", "c", "d", "a"}},
		{name: "last to start", from: 3, to: 0, moved: true, expected: []string{"d", "a", "b", "c"}},
		{name: "forward one", from: 1, to: 3, moved: true, expected: []string{"a", "c", "b", "d"}},
		{name: "back one", from: 2, to: 1, moved: true, expected: []string{"a", "c", "b", "d"}},
		{name: "onto itself", from: 1, to: 1, moved: false, expected: []string{"a", "b", "c", "d"}},
		{name: "just after itself", from: 1, to: 2, moved: false, expected: []string{"a", "b", "c", "d"}},
		{name: "from below range", from: -1, to: 2, moved: false, expected: []string{"a", "b", "c", "d"}},
		{name: "from above range", from: 4, to: 0, moved: false, expected: []string{"a", "b", "c", "d"}},
		{name: "to below range", from: 0, to: -1, moved: false, expected: []string{"a", "b", "c", "d"}},
		{name: "to above range", from: 0, to: 5, moved: false, expected: []string{"a", "b", "c", "d"}},
	} {
		list := []string{"a", "b", "c", "d"}
		c.Equal(one.moved, moveEntry(&list, one.from, one.to), "%s: return value", one.name)
		c.Equal(one.expected, list, "%s: resulting list", one.name)
	}
	var empty []int
	c.False(moveEntry(&empty, 0, 0), "an empty list has nothing to move")
	c.Equal(0, len(empty))
}

// dragDataForRow returns the drag payload the row's drag handle would deliver.
func dragDataForRow(t *testing.T, row *unison.Panel) *editorRowDragData {
	t.Helper()
	handles := panelsOfType[*DragHandle](row)
	if len(handles) == 0 {
		t.Fatal("the row has no drag handle")
	}
	dd, ok := handles[0].data.(*editorRowDragData)
	if !ok {
		t.Fatalf("the drag handle carries a %T", handles[0].data)
	}
	return dd
}

// beginDragOver puts an editor's drag state into the state dataDragOver leaves it in when the pointer is over the given
// insertion position within the rows container, so that a drop can be tested without aiming a pointer at laid-out
// rows; TestDragOverFindsInsertionPositionFromGeometry covers that part of dataDragOver on its own.
func beginDragOver(s *rowDragState, rows *unison.Panel, insert int) {
	s.inDragOver = true
	s.dragTarget = rows
	s.dragInsert = insert
}

// layOutTestEditor gives a test editor a frame of a workable size and lays it out, so that its rows have the positions
// and sizes that dataDragOver's hit-testing works from. The editor is in no window, so nothing else lays it out, and it
// has no layout of its own, since Setup, which would give it one, is not run for a test editor; the one given here has
// the scroll panel fill the frame, as Setup's does.
func layOutTestEditor(d unison.Paneler) {
	p := d.AsPanel()
	p.SetLayout(&unison.FlexLayout{Columns: 1})
	for _, child := range p.Children() {
		child.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, VAlign: align.Fill, HGrab: true, VGrab: true})
	}
	p.SetFrameRect(geom.NewRect(0, 0, 800, 1200))
	p.MarkForLayoutRecursively()
	p.ValidateLayout()
}

// TestDragOverFindsInsertionPositionFromGeometry verifies the part of a row drag that the other drag tests bypass
// through beginDragOver: with the editor laid out, dataDragOver finds the row under the pointer among the dragged row's
// siblings and chooses the insertion position ahead of it when the pointer is above the row's center line and after it
// otherwise; a pointer over no row, or a payload from another editor, makes no drop target; and a drop then lands at
// the position the last dataDragOver chose.
func TestDragOverFindsInsertionPositionFromGeometry(t *testing.T) {
	c := check.New(t)
	a := gurps.NewAncestry()
	a.CommonOptions.HairOptions = []*gurps.WeightedStringOption{
		{Weight: 1, Value: "Black"},
		{Weight: 1, Value: "Brown"},
		{Weight: 1, Value: "Blond"},
	}
	d := newTestAncestryEditorDockable(a)
	layOutTestEditor(d)
	list := listPanelFor(t, d, &a.CommonOptions.HairOptions)
	rows := list.rows.Children()
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, found %d", len(rows))
	}
	c.True(rows[0].FrameRect().Height > 0, "the rows have been laid out")
	c.True(rows[1].FrameRect().Y >= rows[0].FrameRect().Bottom(), "and stacked")
	dd := dragDataForRow(t, rows[2])

	// pointInRow returns the point the given fraction of the way down the row, in the coordinate space of the content
	// panel, which is the space dataDragOver takes its pointer in.
	pointInRow := func(row *unison.Panel, fraction float32) geom.Point {
		r := row.FrameRect()
		return d.content.PointFromRoot(list.rows.PointToRoot(geom.NewPoint(r.CenterX(), r.Y+r.Height*fraction)))
	}
	over := func(pt geom.Point, data any) (inDragOver bool, target *unison.Panel, insert int) {
		c.True(d.dataDragOver(pt, data), "a row drag is always accepted")
		return d.inDragOver, d.dragTarget, d.dragInsert
	}

	inDragOver, target, insert := over(pointInRow(rows[0], 0.25), dd)
	c.True(inDragOver)
	c.Equal(list.rows, target, "the target is the list that owns the dragged row")
	c.Equal(0, insert, "above the first row's center line inserts ahead of it")

	_, _, insert = over(pointInRow(rows[0], 0.75), dd)
	c.Equal(1, insert, "below the first row's center line inserts after it")

	_, _, insert = over(pointInRow(rows[1], 0.25), dd)
	c.Equal(1, insert, "above the second row's center line inserts ahead of it, which is the same position")

	_, _, insert = over(pointInRow(rows[1], 0.5), dd)
	c.Equal(2, insert, "on the center line counts as below it")

	_, _, insert = over(pointInRow(rows[2], 0.75), dd)
	c.Equal(3, insert, "below the last row's center line inserts at the end")

	// The list's border lies just above the first row, so a pointer there is over the list but over no row.
	first := rows[0].FrameRect()
	borderPoint := d.content.PointFromRoot(list.rows.PointToRoot(geom.NewPoint(first.CenterX(), first.Y-0.5)))
	inDragOver, target, insert = over(borderPoint, dd)
	c.False(inDragOver, "a pointer over no row makes no drop target")
	c.Nil(target)
	c.Equal(-1, insert)

	other := newTestAncestryEditorDockable(gurps.NewAncestry())
	foreign := &editorRowDragData{editor: other, row: rows[0], title: "Foreign", move: func(_ int) bool { return true }}
	inDragOver, _, _ = over(pointInRow(rows[0], 0.25), foreign)
	c.False(inDragOver, "another editor's payload makes no drop target")

	over(pointInRow(rows[0], 0.25), dd)
	d.dataDragDrop(pointInRow(rows[0], 0.25), dd)
	c.Equal([]string{"Blond", "Black", "Brown"}, optionValues(d.model.CommonOptions.HairOptions),
		"the drop lands at the position the pointer chose")
	c.False(d.inDragOver, "the drag state is cleared")
	c.True(d.undoMgr.CanUndo(), "and the drop is undoable")
}

// TestAncestryDragDropReordersOwningList verifies that dropping a row moves its entry within the list that owns it,
// posts a single undo edit, rebuilds the rows, and clears the drag state.
func TestAncestryDragDropReordersOwningList(t *testing.T) {
	c := check.New(t)
	a := gurps.NewAncestry()
	a.CommonOptions.HairOptions = []*gurps.WeightedStringOption{
		{Weight: 1, Value: "Black"},
		{Weight: 1, Value: "Brown"},
		{Weight: 1, Value: "Blond"},
	}
	d := newTestAncestryEditorDockable(a)
	list := listPanelFor(t, d, &a.CommonOptions.HairOptions)
	dd := dragDataForRow(t, list.rows.Children()[0])
	c.Equal("Hair Option Drag", dd.title)

	beginDragOver(&d.rowDragState, list.rows, 3)
	d.dataDragDrop(geom.Point{}, dd)
	c.Equal([]string{"Brown", "Blond", "Black"}, optionValues(d.model.CommonOptions.HairOptions))
	c.False(d.inDragOver, "the drag state is cleared")
	c.Equal(-1, d.dragInsert)
	c.Nil(d.dragTarget)
	c.True(d.Modified())
	rows := listPanelFor(t, d, &d.model.CommonOptions.HairOptions).rows.Children()
	c.Equal(3, len(rows), "the rows are rebuilt")
	c.Equal("Brown", stringFieldFor(t, d, d.model.CommonOptions.HairOptions[0].KeyPrefix+"value").Text())

	d.undoMgr.Undo()
	c.Equal([]string{"Black", "Brown", "Blond"}, optionValues(d.model.CommonOptions.HairOptions), "undo restores the order")
	c.False(d.undoMgr.CanUndo(), "the drop is a single edit")
	d.undoMgr.Redo()
	c.Equal([]string{"Brown", "Blond", "Black"}, optionValues(d.model.CommonOptions.HairOptions))
}

// TestAncestryDragDropOnGendersAndGenerators verifies that the same mechanism reorders the gender list and the name
// generator list.
func TestAncestryDragDropOnGendersAndGenerators(t *testing.T) {
	c := check.New(t)
	a := gurps.NewAncestry()
	a.CommonOptions.NameGenerators = []string{"First", "Last"}
	d := newTestAncestryEditorDockable(a, "First", "Last")

	genders := rootAncestryPanel(t, d).genders
	dd := dragDataForRow(t, genders.Children()[1])
	c.Equal("Gender Drag", dd.title)
	beginDragOver(&d.rowDragState, genders, 0)
	d.dataDragDrop(geom.Point{}, dd)
	c.Equal([]string{"Female", "Male"}, genderNames(d.model))

	generators := nameGeneratorsPanelFor(t, d, d.model.CommonOptions).rows
	dd = dragDataForRow(t, generators.Children()[0])
	c.Equal("Name Generator Drag", dd.title)
	beginDragOver(&d.rowDragState, generators, 2)
	d.dataDragDrop(geom.Point{}, dd)
	c.Equal([]string{"Last", "First"}, d.model.CommonOptions.NameGenerators)

	d.undoMgr.Undo()
	c.Equal([]string{"First", "Last"}, d.model.CommonOptions.NameGenerators)
	d.undoMgr.Undo()
	c.Equal([]string{"Male", "Female"}, genderNames(d.model))
	c.False(d.undoMgr.CanUndo())
}

// TestAncestryDragDropIgnoresNoOpsAndForeignPayloads verifies that a drop that would not move anything posts no undo
// edit, that a payload from another editor is ignored, and that both still clear the drag state.
func TestAncestryDragDropIgnoresNoOpsAndForeignPayloads(t *testing.T) {
	c := check.New(t)
	a := gurps.NewAncestry()
	a.CommonOptions.SkinOptions = []*gurps.WeightedStringOption{{Weight: 1, Value: "Pale"}, {Weight: 1, Value: "Dark"}}
	d := newTestAncestryEditorDockable(a)
	list := listPanelFor(t, d, &a.CommonOptions.SkinOptions)

	// Dropping a row just below itself leaves it where it is.
	beginDragOver(&d.rowDragState, list.rows, 1)
	d.dataDragDrop(geom.Point{}, dragDataForRow(t, list.rows.Children()[0]))
	c.Equal([]string{"Pale", "Dark"}, optionValues(d.model.CommonOptions.SkinOptions))
	c.False(d.undoMgr.CanUndo(), "a no-op drop posts nothing")
	c.False(d.inDragOver)

	// A payload belonging to another editor is not acted on, even if its move would succeed.
	other := newTestAncestryEditorDockable(gurps.NewAncestry())
	moved := false
	foreign := &editorRowDragData{
		editor: other,
		row:    list.rows.Children()[0],
		title:  "Foreign",
		move:   func(_ int) bool { moved = true; return true },
	}
	beginDragOver(&d.rowDragState, list.rows, 2)
	d.dataDragDrop(geom.Point{}, foreign)
	c.False(moved, "another editor's payload is ignored")
	c.False(d.undoMgr.CanUndo())
	c.False(d.inDragOver)

	// Without a drag in progress, a drop does nothing.
	d.dataDragDrop(geom.Point{}, dragDataForRow(t, listPanelFor(t, d, &d.model.CommonOptions.SkinOptions).rows.Children()[1]))
	c.Equal([]string{"Pale", "Dark"}, optionValues(d.model.CommonOptions.SkinOptions))
	c.False(d.undoMgr.CanUndo())
}
