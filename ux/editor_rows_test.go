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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/study"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// The shared row setup widens the right inset for an outermost row so that the overlaid scrollbar doesn't cover its
// content.
func TestConfigureEditorRow(t *testing.T) {
	c := check.New(t)
	nested := unison.NewPanel()
	configureEditorRow(nested, 3, false)
	c.Equal(geom.Insets{
		Top:    unison.StdVSpacing,
		Left:   unison.StdHSpacing,
		Bottom: unison.StdVSpacing,
		Right:  unison.StdHSpacing,
	}, nested.Border().Insets())
	layout, ok := nested.Layout().(*unison.FlexLayout)
	c.True(ok, "the row is laid out as a grid")
	c.Equal(3, layout.Columns)
	data, ok := nested.LayoutData().(*unison.FlexLayoutData)
	c.True(ok, "the row has flex layout data")
	c.Equal(align.Fill, data.HAlign)
	c.True(data.HGrab, "the row fills the width of the list")
	c.NotNil(nested.DrawCallback, "the row draws its alternating background")

	outermost := unison.NewPanel()
	configureEditorRow(outermost, 4, true)
	c.Equal(float32(unison.StdHSpacing*2), outermost.Border().Insets().Right,
		"an outermost row keeps clear of the scrollbar")
	c.Equal(float32(unison.StdHSpacing), outermost.Border().Insets().Left)

	first := unison.NewSVGButton(unison.TrashSVG)
	second := unison.NewSVGButton(unison.CircledAddSVG)
	column := newEditorRowButtonColumn(first, second)
	children := column.Children()
	c.Equal(2, len(children))
	c.Equal(first.AsPanel(), children[0])
	c.Equal(second.AsPanel(), children[1])
	layout, ok = column.Layout().(*unison.FlexLayout)
	c.True(ok, "the buttons are stacked in a grid")
	c.Equal(1, layout.Columns)
	data, ok = column.LayoutData().(*unison.FlexLayoutData)
	c.True(ok, "the column has flex layout data")
	c.Equal(align.Middle, data.HAlign)
}

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

// beginDragOver puts an editor's drag state where dataDragOver would leave it for the given insertion position within
// the rows container, so a drop can be tested without aiming a pointer at laid-out rows;
// TestDragOverFindsInsertionPositionFromGeometry covers that part of dataDragOver on its own.
func beginDragOver(s *rowDragState, rows *unison.Panel, insert int) {
	s.inDragOver = true
	s.dragTarget = rows
	s.dragInsert = insert
}

// layOutTestEditor gives a test editor a frame of a workable size and lays it out, so its rows have the positions and
// sizes dataDragOver's hit-testing works from. A test editor is in no window and never runs Setup, so nothing else
// would lay it out or give it a layout; the one given here has the scroll panel fill the frame, as Setup's does.
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

// This covers the part of a row drag that the other drag tests bypass through beginDragOver: with the editor laid out,
// dataDragOver finds the row under the pointer among the dragged row's siblings and inserts ahead of it when the
// pointer is above the row's center line and after it otherwise.
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

	// pointInRow returns the point the given fraction of the way down the row, in the content panel's coordinate space,
	// which is the space dataDragOver takes its pointer in.
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

// The same mechanism must also reorder the gender list and the name generator list.
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

	// Another editor's payload is not acted on, even if its move would succeed.
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

// titledSection is a panel type of its own, so that a test can tell whether initTitledEditorSection made the panel its
// own Self.
type titledSection struct {
	unison.Panel
}

// sectionRoot is a ModifiableRoot that counts how often it is marked as modified.
type sectionRoot struct {
	unison.Panel
	modified int
}

func (r *sectionRoot) MarkModified(_ unison.Paneler) {
	r.modified++
}

// expectTitledEditorSection verifies that the panel carries the titled editor section scaffold, whose layout data and
// border are what make the sections line up with each other in an editor.
func expectTitledEditorSection(c check.Checker, p unison.Paneler, title string) {
	c.Helper()
	panel := p.AsPanel()
	layout, ok := panel.Layout().(*unison.FlexLayout)
	c.True(ok, "%s: a titled editor section uses a FlexLayout", title)
	c.Equal(1, layout.Columns, "%s: a titled editor section stacks its rows in a single column", title)
	c.Equal(float32(unison.StdHSpacing), layout.HSpacing, "%s: horizontal spacing", title)
	c.Equal(float32(unison.StdVSpacing), layout.VSpacing, "%s: vertical spacing", title)
	c.Equal(&unison.FlexLayoutData{HSpan: 2, HAlign: align.Fill, HGrab: true}, panel.LayoutData(),
		"%s: a titled editor section spans, fills and grabs both editor columns", title)
	expected := unison.NewCompoundBorder(&TitledBorder{Title: title, Font: unison.LabelFont},
		unison.NewEmptyBorder(geom.NewUniformInsets(2)))
	c.Equal(expected.Insets(), panel.Border().Insets(), "%s: the border is the titled border with a 2-point inset", title)
	c.NotNil(panel.DrawCallback, "%s: a titled editor section paints its own background", title)
}

// Each of the four section panels must be built on the shared scaffold, so none can drift from the others again in how
// it fills the editor.
func TestInitTitledEditorSection(t *testing.T) {
	c := check.New(t)
	section := &titledSection{}
	initTitledEditorSection(section, "Sample")
	c.True(section.Self == section, "the section becomes its own Self")
	expectTitledEditorSection(c, section, "Sample")

	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	defs := []*gurps.SkillDefault{{DefaultType: gurps.DexterityID}}
	expectTitledEditorSection(c, newDefaultsPanel(entity, &defs), "Defaults")
	bonus := gurps.NewAttributeBonus(gurps.StrengthID)
	bonus.SetOwner(owner)
	features := gurps.Features{bonus}
	expectTitledEditorSection(c, newFeaturesPanel(entity, owner, &features, false), "Features")
	root := gurps.NewPrereqList()
	expectTitledEditorSection(c, newPrereqPanel(entity, &root, prereq.TypesForNonEquipment, false), "Prerequisites")
	level := study.Standard
	studies := []*gurps.Study{{Type: study.Self}}
	expectTitledEditorSection(c, newStudyPanel(entity, &level, &studies), "Study")
}

func TestNewSectionAddButton(t *testing.T) {
	c := check.New(t)
	root := &sectionRoot{}
	root.Self = root
	section := unison.NewPanel()
	root.AddChild(section)
	inserted := 0
	added := true
	button := newSectionAddButton(section, func() bool {
		inserted++
		return added
	})
	icon, ok := button.Drawable.(*unison.DrawableSVG)
	c.True(ok, "the add button shows an icon")
	c.Equal(unison.CircledAddSVG, icon.SVG, "the add button shows the add icon")

	button.Click()
	c.Equal(1, inserted, "a click runs the insertion")
	c.Equal(1, root.modified, "an insertion that added something marks the root as modified")

	added = false
	button.Click()
	c.Equal(2, inserted, "a further click runs the insertion again")
	c.Equal(1, root.modified, "an insertion that added nothing leaves the root alone")
}
