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
	"cmp"
	"fmt"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/pathop"
)

// Each row is found by its path, which its panel defines. A widget's reference key is the path of its row followed by
// one of the suffixes below, or by a colon and a name of its own.
const (
	keyMore     = ":more"
	keySentence = ":sentence"
	// keyChip follows the key of a chip's criterion, so that the chip doesn't share the key of the field within it.
	keyChip = ":chip"
	// keyFirst marks the editor of an open row; focusing it focuses its first text field, or else its first control.
	keyFirst = ":first"
)

// sectionSummaryKey is the reference key of the paragraph shown in place of the rows while the panel is collapsed.
const sectionSummaryKey = "summary"

// Where a dragged row goes in relation to what it is dropped on.
const (
	dropBefore = iota
	dropAfter
	dropInto
)

const (
	pillCornerRadius = 100
	// rowEndGap keeps what is drawn behind a row short of the panel's right edge, so it doesn't run into it.
	rowEndGap = 4
	// chipPadding is the room a chip leaves above and below the controls it holds.
	chipPadding = 2
)

// rowSnapshot is a panel of sentence rows as it stood before or after an edit: its data, the open row, and the
// reference key of the widget that undo or redo returning to it gives the focus to.
type rowSnapshot[T any] struct {
	data  T
	focus string
	open  string
}

// sentenceRows is a panel of rows that each read as a sentence until they are opened, one at a time, to edit them.
// Every change, typing included, records a snapshot of the panel's data of type T with the editor's undo manager, and
// changes to the rows rebuild the panel's content. A panel embeds it and calls initRows.
type sentenceRows[T any] struct {
	unison.Panel
	targetMgr *TargetMgr
	dragKey   *uti.DataType
	// collapse is the title bar that collapses the panel to a paragraph, once initCollapse has set it up.
	collapse *sectionToggle
	// spotAt returns the panel a row dragged to a point would be dropped on, or nil where it can't go, and where it would
	// go in relation to it. While a row is dragged over the panel, dropTarget and dropWhere hold what spotAt last
	// returned.
	spotAt     func(where geom.Point, data any) (target *unison.Panel, at int)
	dropTarget *unison.Panel
	dropWhere  int
	// fill adds the panel's content. snapshot returns a copy of the data being edited, ahead of a change and after it,
	// restore makes a copy the data being edited, and dataHash returns a hash of the data being edited.
	fill       func()
	snapshot   func() T
	restore    func(T)
	dataHash   func() uint64
	open       string
	focus      string
	editKey    string
	editID     int64
	rebuilding bool
}

// initRows sets up the rows of a panel whose rows are dragged as dragKey, with the callbacks described on sentenceRows.
func (p *sentenceRows[T]) initRows(dragKey *uti.DataType, fill func(), snapshot func() T, restore func(T), dataHash func() uint64) {
	p.targetMgr = NewTargetMgr(p)
	p.dragKey = dragKey
	p.fill = fill
	p.snapshot = snapshot
	p.restore = restore
	p.dataHash = dataHash
	p.KeyDownCallback = p.keyDown
}

// initDrop lets the rows be dragged within the panel, with spotAt as described on sentenceRows. drop moves a row to
// where spotAt says it goes.
func (p *sentenceRows[T]) initDrop(spotAt func(where geom.Point, data any) (*unison.Panel, int), drop func(where geom.Point, data any)) {
	p.spotAt = spotAt
	installPanelDragDrop(p.AsPanel(), p.dragKey, p.dragOver, p.dragExit, drop)
	p.DrawOverCallback = p.drawDrop
}

// initCollapse lets the panel be collapsed to a paragraph by clicking the title its border draws, starting out collapsed
// or not as asked. The panel's fill calls addTitleBar first.
func (p *sentenceRows[T]) initCollapse(border *TitledBorder, collapsed bool) {
	p.collapse = newSectionToggle(p, border, collapsed, p.collapseChanged)
}

// addTitleBar adds the title bar and, while the panel is collapsed, the paragraph of text shown in place of the rows,
// which expands the panel again when clicked. It returns the paragraph, or nil while the panel is expanded, when the
// caller goes on to add the rows.
func (p *sentenceRows[T]) addTitleBar(text func() string) *sentenceButton {
	p.AddChild(p.collapse)
	if !p.collapse.collapsed {
		return nil
	}
	paragraph := newSentenceButton(text(), p.collapse.toggle)
	paragraph.RefKey = sectionSummaryKey
	paragraph.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	p.AddChild(paragraph)
	return paragraph
}

// rowsView is how a panel of sentence rows is shown: whether it is collapsed, and which row is open.
type rowsView struct {
	open      string
	collapsed bool
}

// view returns how the panel is shown, so that setView can show the panel that replaces it the same way.
func (p *sentenceRows[T]) view() rowsView {
	return rowsView{open: p.open, collapsed: p.collapse != nil && p.collapse.collapsed}
}

// setView shows the panel as view says, filling it again at once when that changes anything, so that what held the
// focus can be found in it before anything else happens. The open row closes if there is no longer one at its path.
func (p *sentenceRows[T]) setView(view rowsView) {
	if view == p.view() {
		return
	}
	p.open = view.open
	if p.collapse != nil {
		p.collapse.collapsed = view.collapsed
	}
	p.RemoveAllChildren()
	p.fill()
}

// collapseChanged rebuilds the panel once it has been collapsed or expanded. That is no edit, so it isn't undone and
// doesn't mark the editor modified, and the open row stays open for when the panel expands. Collapsing hands the focus
// from what it hides to the title bar.
func (p *sentenceRows[T]) collapseChanged() {
	var focus string
	if p.collapse.collapsed && p.targetMgr.CurrentFocusRef() != nil {
		focus = sectionToggleKey
	}
	p.rebuild(focus)
}

// keyDown has Escape close the open row, leaving it alone while the panel is collapsed. Escape within the panel never
// reaches the editor, where it would discard the changes.
func (p *sentenceRows[T]) keyDown(keyCode unison.KeyCode, mods mod.Modifiers, _ bool) bool {
	if keyCode != unison.KeyEscape || !noModifiersDown(mods) {
		return false
	}
	if p.open != "" && (p.collapse == nil || !p.collapse.collapsed) {
		p.toggle(p.open)
	}
	return true
}

// edit applies change to the data and records snapshots of the panel before and after it, under the title. Undo gives
// the focus to the widget with the reference key key, as redo does unless focusAfter is set, which also rebuilds the
// panel with the focus there. Consecutive edits with the same non-empty key, such as keystrokes in a field, are
// recorded as one until the panel is next rebuilt. It returns the snapshot after, or nil if nothing changed.
func (p *sentenceRows[T]) edit(title, key, focusAfter string, change func()) *rowSnapshot[T] {
	before := &rowSnapshot[T]{data: p.snapshot(), focus: key, open: p.open}
	hash := p.dataHash()
	change()
	if focusAfter != "" {
		defer p.rebuild(focusAfter)
	}
	if p.dataHash() == hash {
		return nil
	}
	if key == "" || key != p.editKey {
		p.editID = unison.NextUndoID()
	}
	p.editKey = key
	after := &rowSnapshot[T]{data: p.snapshot(), focus: cmp.Or(focusAfter, key), open: p.open}
	if parent := p.Parent(); parent != nil {
		if mgr := unison.UndoManagerFor(parent); mgr != nil {
			mgr.Add(&unison.UndoEdit[*rowSnapshot[T]]{
				ID:       p.editID,
				EditName: title,
				EditCost: 1,
				UndoFunc: func(e *unison.UndoEdit[*rowSnapshot[T]]) { p.install(e.BeforeData) },
				RedoFunc: func(e *unison.UndoEdit[*rowSnapshot[T]]) { p.install(e.AfterData) },
				AbsorbFunc: func(e *unison.UndoEdit[*rowSnapshot[T]], other unison.Undoable) bool {
					if o, ok := other.(*unison.UndoEdit[*rowSnapshot[T]]); ok && o.ID == e.ID {
						e.AfterData = o.AfterData
						return true
					}
					return false
				},
				BeforeData: before,
				AfterData:  after,
			})
		}
	}
	MarkModified(p)
	return after
}

// install returns the panel to the snapshot, for undo and redo, editing a copy of its data.
func (p *sentenceRows[T]) install(snapshot *rowSnapshot[T]) {
	p.restore(snapshot.data)
	p.open = snapshot.open
	MarkModified(p)
	p.rebuild(snapshot.focus)
}

// restructure moves, adds or removes rows through the widget with the reference key from, then rebuilds. change
// returns the reference key of the widget that takes the focus after; with none, the widget with the reference key
// fallback does. change also keeps the open row open wherever it ends up, and closes it if it is removed.
func (p *sentenceRows[T]) restructure(title, from, fallback string, change func() (focus string)) {
	var focus string
	after := p.edit(title, from, "", func() { focus = change() })
	focus = cmp.Or(focus, fallback)
	if after != nil {
		after.focus = focus
	}
	p.rebuild(focus)
}

// rebuild replaces the panel's content once the current event has been handled, since that may have come from a widget
// about to be thrown away, keeping the scroll position and giving the focus to focus or else back where it was.
func (p *sentenceRows[T]) rebuild(focus string) {
	if focus != "" {
		p.focus = focus
	}
	// Whatever asked for the rebuild ends a run of typing, so typing in the same field after it is a step of its own.
	p.editKey = ""
	if p.rebuilding {
		return
	}
	p.rebuilding = true
	unison.InvokeTask(func() {
		p.rebuilding = false
		ref := p.targetMgr.CurrentFocusRef()
		focus = p.focus
		p.focus = ""
		scroll := p.ScrollRoot()
		var h, v float32
		if scroll != nil {
			h, v = scroll.Position()
		}
		p.RemoveAllChildren()
		p.fill()
		top := p.AsPanel()
		if d := unison.Ancestor[unison.Dockable](p); d != nil {
			top = d.AsPanel()
		}
		top.MarkForLayoutRecursively()
		top.ValidateLayout()
		if scroll != nil {
			scroll.SetPosition(h, v)
		}
		switch {
		case focus != "" && (ref == nil || ref.Key != focus) && p.focusOn(focus):
		case ref != nil && ref.Key != "" && p.focusOn(ref.Key):
			if s, ok := p.FindRefKey(ref.Key).Self.(Selectable); ok && ref.Selectable {
				s.SetSelection(ref.SelStart, ref.SelEnd)
			}
		// With the widget that had the focus gone, the open row takes it, or else the panel's first control.
		case ref != nil && p.open != "" && p.focusOn(p.open+keyFirst):
		case ref != nil:
			if first := p.FirstFocusableChild(); first != nil {
				first.RequestFocus()
			}
		}
		p.MarkForRedraw()
	})
}

// focusOn gives the focus to the widget with the reference key, or to the first control within it when it takes none
// itself, reporting whether there was one.
func (p *sentenceRows[T]) focusOn(key string) bool {
	target := p.FindRefKey(key)
	if target == nil {
		return false
	}
	if !target.Focusable() {
		first := target.FirstFocusableChild()
		if strings.HasSuffix(key, keyFirst) {
			target.HasInSelfOrDescendants(func(one *unison.Panel) bool {
				if _, ok := one.Self.(Selectable); ok && one.Focusable() {
					first = one
					return true
				}
				return false
			})
		}
		if target = first; target == nil {
			return false
		}
	}
	target.RequestFocus()
	target.ScrollIntoView()
	return true
}

// toggle opens the row at the path, closing any other, or closes it if it is the open one.
func (p *sentenceRows[T]) toggle(path string) {
	if p.open == path {
		p.open = ""
		p.rebuild(path + keySentence)
	} else {
		p.open = path
		p.rebuild(path + keyFirst)
	}
}

// rowDrag is the payload of a row being dragged: the panel it belongs to and its path there.
type rowDrag struct {
	panel *unison.Panel
	path  string
}

// dragPath returns the path of the row the data of a drag carries, or "" if it isn't one of this panel's.
func (p *sentenceRows[T]) dragPath(data any) string {
	if d, ok := data.(*rowDrag); ok && d.panel == p.AsPanel() {
		return d.path
	}
	return ""
}

// grip adds a drag handle to the row at the path and lets either be dragged.
func (p *sentenceRows[T]) grip(row *unison.Panel, path string) *DragHandle {
	handle := NewDragHandle(p.dragKey, nil)
	row.AddChild(handle)
	p.dragBy(handle.AsPanel(), row, path)
	p.dragBy(row, row, path)
	return handle
}

// dragBy lets a press on target that moves far enough to be a drag start dragging the row at the path, shown as an
// image of the row. Only moving counts, not holding the button down, so that a slow click is still a click. A press
// that becomes a drag is not also a click.
func (p *sentenceRows[T]) dragBy(target, row *unison.Panel, path string) {
	var dragged bool
	var start geom.Point
	down, up := target.MouseDownCallback, target.MouseUpCallback
	target.MouseDownCallback = func(where geom.Point, button, count int, mods mod.Modifiers) bool {
		dragged = false
		start = where
		return down == nil || down(where, button, count, mods)
	}
	target.MouseDragCallback = func(where geom.Point, button int, _ mod.Modifiers) bool {
		_, drift := unison.DragGestureParameters()
		if dragged || button != unison.ButtonLeft ||
			(xmath.Abs(where.X-start.X) <= drift && xmath.Abs(where.Y-start.Y) <= drift) {
			return true
		}
		dragged = true
		startPanelDrag(row, p.dragKey, &rowDrag{panel: p.AsPanel(), path: path}, row.FrameRect().Size, geom.Point{},
			func(gc *unison.Canvas, rect geom.Rect) {
				gc.DrawRect(rect, unison.ThemeBelowSurface.Paint(gc, rect, paintstyle.Fill))
				row.Draw(gc, rect)
			}, p.MarkForRedraw)
		return true
	}
	target.MouseUpCallback = func(where geom.Point, button int, mods mod.Modifiers) bool {
		return dragged || up == nil || up(where, button, mods)
	}
}

// doneButton returns the button that closes the open row at the path.
func (p *sentenceRows[T]) doneButton(path string) *unison.Button {
	done := unison.NewButton()
	done.SetTitle(i18n.Text("Done"))
	done.RefKey = path + ":done"
	done.ClickCallback = func() { p.toggle(path) }
	return done
}

// addMoreButton adds the button for the more menu of the row at the path, which offers the entries.
func addMoreButton(parent *unison.Panel, path string, entries func() []menuEntry) {
	b := newIconButton(path+keyMore, svg.CircledVerticalEllipsis, i18n.Text("More actions"))
	b.ClickCallback = func() { showMenu(b.AsPanel(), entries()) }
	addCentered(parent, b)
}

// sentenceRow returns the panel of the row at the path: the sentence describe returns, or while the row is open the
// controls editor returns, beside a button for the entries of its more menu. A row that isn't editable, such as one
// this version of GCS doesn't understand, has a sentence that doesn't open it. lead, when set, adds what goes between
// the row's grip and its sentence, returning it and its height, to put it on the first line. trail, when set, adds
// what follows the sentence of a closed row.
func (p *sentenceRows[T]) sentenceRow(path string, describe func() string, editable bool, editor func() *unison.Panel, more func() []menuEntry, lead func(row *unison.Panel) (*unison.Panel, float32), trail func(row *unison.Panel, sentence *sentenceButton)) *unison.Panel {
	open := path == p.open
	row := unison.NewPanel()
	grip := p.grip(row, path)
	var leader *unison.Panel
	var leaderSize float32
	if lead != nil {
		leader, leaderSize = lead(row)
	}
	var main *unison.Panel
	var line float32
	if open {
		row.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: 6, Left: 4, Bottom: 8, Right: 8}))
		main = editor()
		row.AddChild(main)
		row.AddChild(p.doneButton(path))
		line = controlHeight(row)
	} else {
		row.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: 3, Left: 4, Bottom: 3, Right: 8}))
		var click func()
		if editable {
			click = func() { p.toggle(path) }
		}
		sentence := newSentenceButton(describe(), click)
		sentence.RefKey = path + keySentence
		p.dragBy(sentence.AsPanel(), row, path)
		if click == nil {
			// One this version of GCS doesn't understand can't be edited, so its sentence is static text.
			sentence.Tooltip = newWrappedTooltip(i18n.Text("This was most likely created by a newer version of GCS. Its original data will be written back out unchanged when this file is saved."))
		}
		main = sentence.AsPanel()
		row.AddChild(main)
		if trail != nil {
			trail(row, sentence)
		}
		line = sentence.lineHeight()
	}
	addMoreButton(row, path, more)
	hbox(row, unison.StdHSpacing)
	putOnLine(grip.AsPanel(), line, grip.svg.Size.Height)
	if leader != nil {
		putOnLine(leader, line, leaderSize)
	}
	for _, child := range row.Children() {
		switch {
		case child == main:
			child.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, VAlign: align.Middle, HGrab: true})
		case open:
			if _, ok := child.Self.(*unison.Button); ok {
				fitLine(child)
			}
			child.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Start})
		case child != grip.AsPanel() && child != leader:
			_, pref, _ := child.Sizes(geom.Size{})
			putOnLine(child, line, pref.Height)
		}
	}
	frameRow(row, open)
	return row
}

// frameRow draws an open row on a rounded panel below the surface, and a closed one over a faint divider.
func frameRow(row *unison.Panel, open bool) {
	row.DrawCallback = func(gc *unison.Canvas, _ geom.Rect) {
		r := row.ContentRect(true)
		r.Width -= rowEndGap
		if open {
			gc.DrawRoundedRect(r, geom.NewUniformSize(8), unison.ThemeBelowSurface.Paint(gc, r, paintstyle.Fill))
			return
		}
		gc.DrawLine(geom.NewPoint(r.X, r.Bottom()-0.5), geom.NewPoint(r.Right(), r.Bottom()-0.5),
			faintInk(unison.ThemeSurfaceEdge).Paint(gc, r, paintstyle.Stroke))
	}
}

// dragOver keeps the pointer in view and marks where the row would be dropped.
func (p *sentenceRows[T]) dragOver(where geom.Point, data any) bool {
	p.ScrollRectIntoView(geom.NewRect(where.X, where.Y-16, 1, 1))
	p.ScrollRectIntoView(geom.NewRect(where.X, where.Y+16, 1, 1))
	target, at := p.spotAt(where, data)
	if target != p.dropTarget || at != p.dropWhere {
		p.dropTarget = target
		p.dropWhere = at
		p.MarkForRedraw()
	}
	return true
}

func (p *sentenceRows[T]) dragExit() {
	p.dropTarget = nil
	p.MarkForRedraw()
}

// drawDrop dims the row being dragged, and marks where it would go (see drawDropMarker).
func (p *sentenceRows[T]) drawDrop(gc *unison.Canvas, _ geom.Rect) {
	if path := p.dragPath(panelDragData); path != "" {
		if more := p.FindRefKey(path + keyMore); more != nil {
			rect := p.RectFromRoot(more.Parent().RectToRoot(more.Parent().ContentRect(true)))
			gc.DrawRect(rect, faintInk(unison.ThemeSurface).Paint(gc, rect, paintstyle.Fill))
		}
	}
	if p.dropTarget == nil {
		return
	}
	rect := p.RectFromRoot(p.dropTarget.RectToRoot(p.dropTarget.ContentRect(true)))
	rect.Width -= rowEndGap
	drawDropMarker(gc, rect, p.dropWhere)
}

// drawDropMarker marks where a dragged row would go in relation to r, in the ink of GCS's other drop markers: a line
// before or after it, or a tint over it for a drop into it.
func drawDropMarker(gc *unison.Canvas, r geom.Rect, at int) {
	if at == dropInto {
		gc.DrawRoundedRect(r, geom.NewUniformSize(6), faintInk(unison.ThemeWarning).Paint(gc, r, paintstyle.Fill))
		return
	}
	y := r.Y
	if at == dropAfter {
		y = r.Bottom()
	}
	paint := unison.ThemeWarning.Paint(gc, r, paintstyle.Stroke)
	paint.SetStrokeWidth(2)
	gc.DrawLine(geom.NewPoint(r.X, y), geom.NewPoint(r.Right(), y), paint)
}

// optional adds an optional criterion to chips, its chip keyed from path, a colon and key: as a chip of the controls
// populate adds while on, else as a dashed button labeled label that adds it. addTitle and removeTitle are the titles
// of adding and removing it.
func (p *sentenceRows[T]) optional(chips *unison.Panel, path, key, label, addTitle, removeTitle string, on bool, add, remove func(), populate func(chip *unison.Panel)) {
	addKey := path + ":add " + key
	if on {
		p.chip(chips, path+":"+key, removeTitle, addKey, remove, populate)
		return
	}
	b := newDashedButton(label, func() { p.edit(addTitle, addKey, path+":"+key+keyChip, add) })
	b.RefKey = addKey
	fitLine(b)
	addCentered(chips, b)
}

// rowCriterion describes an optional criterion of a row: the subject its controls are named for, the button that adds
// it, and the words its chip's first popup starts each choice with or, when notPrefix is set, each "not" choice (see
// criteria.PrefixedStringComparisonChoices). addTitle and removeTitle, when set, replace the titles of adding and
// removing it that are made from the subject.
type rowCriterion struct {
	subject, add, prefix, notPrefix, addTitle, removeTitle string
}

// rowCriteria returns the optional criterion with the key its widgets' reference keys are made from.
func rowCriteria(key string) rowCriterion {
	switch key {
	case "notes":
		return rowCriterion{subject: i18n.Text("Notes"), add: i18n.Text("+ notes"), prefix: i18n.Text("and whose notes")}
	case "tag":
		return rowCriterion{
			subject: i18n.Text("Tag"), add: i18n.Text("+ tag"),
			prefix: i18n.Text("and at least one tag"), notPrefix: i18n.Text("and all tags"),
		}
	case "usage":
		return rowCriterion{subject: i18n.Text("Usage"), add: i18n.Text("+ usage"), prefix: i18n.Text("and whose usage")}
	case "specialization":
		return rowCriterion{
			subject: i18n.Text("Specialization"), add: i18n.Text("+ specialization"),
			prefix: i18n.Text("and whose specialization"),
		}
	case "optspecialization":
		return rowCriterion{
			subject: i18n.Text("Optional Specialization"), add: i18n.Text("+ optional specialization"),
			prefix: i18n.Text("and whose optional specialization"),
		}
	case "level":
		return rowCriterion{subject: i18n.Text("Level"), add: i18n.Text("+ level"), prefix: i18n.Text("and whose level")}
	case "power":
		return rowCriterion{
			subject: i18n.Text("Power Source"), add: i18n.Text("+ power source"),
			prefix: i18n.Text("and whose power source"),
		}
	case "combined":
		return rowCriterion{
			subject: i18n.Text("Combined With"), add: i18n.Text("+ combined with"),
			prefix: i18n.Text("combined with"), addTitle: i18n.Text("Add Combined Attribute"),
			removeTitle: i18n.Text("Remove Combined Attribute"),
		}
	default:
		return rowCriterion{}
	}
}

// titles returns the titles of adding and removing the criterion.
func (c *rowCriterion) titles() (add, remove string) {
	return cmp.Or(c.addTitle, fmt.Sprintf(i18n.Text("Add %s"), c.subject)),
		cmp.Or(c.removeTitle, fmt.Sprintf(i18n.Text("Remove %s"), c.subject))
}

// textChip adds the optional text criterion with the key, which is in use while its comparison isn't "is anything".
func (p *sentenceRows[T]) textChip(chips *unison.Panel, path, key string, c *criteria.Text) {
	p.optionalCriterion(chips, path, key, c.Compare != criteria.AnyText,
		func() { *c = criteria.Text{Compare: criteria.IsText} },
		func() { *c = criteria.Text{} },
		func(chip *unison.Panel) {
			one := rowCriteria(key)
			p.textCriteria(chip, path+":"+key, one.subject, "", one.prefix, cmp.Or(one.notPrefix, one.prefix), c, false)
		})
}

// optionalCriterion adds the optional criterion with the key to chips: as a chip of the controls populate adds while
// on, else as a dashed button that adds it.
func (p *sentenceRows[T]) optionalCriterion(chips *unison.Panel, path, key string, on bool, add, remove func(), populate func(chip *unison.Panel)) {
	c := rowCriteria(key)
	addTitle, removeTitle := c.titles()
	p.optional(chips, path, key, c.add, addTitle, removeTitle, on, add, remove, populate)
}

// chipPanel is the panel of an optional criterion in use, which holds its controls.
type chipPanel struct {
	unison.Panel
}

// chip adds a rounded chip keyed key+keyChip to the parent, holding the controls populate adds and a button titled
// title that removes the criterion, giving the focus to the widget with the reference key after.
func (p *sentenceRows[T]) chip(parent *unison.Panel, key, title, after string, remove func(), populate func(chip *unison.Panel)) {
	chip := &chipPanel{}
	chip.Self = chip
	chip.RefKey = key + keyChip
	chip.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: chipPadding, Left: 10, Bottom: chipPadding, Right: 3}))
	chip.DrawCallback = func(gc *unison.Canvas, _ geom.Rect) {
		rect := chip.ContentRect(true)
		unison.DrawRoundedRectBase(gc, rect, geom.NewUniformSize(rect.Height/2), 1, unison.ThemeSurface,
			unison.ThemeSurfaceEdge)
	}
	populate(chip.AsPanel())
	x := newIconButton("", unison.CircledXSVG, title)
	x.ClickCallback = func() { p.edit(title, chip.RefKey, after, remove) }
	addCentered(chip.AsPanel(), x)
	addCentered(parent, hbox(chip.AsPanel(), 5))
}

// textCriteria adds the comparison of a text criterion, each choice after the prefix or, for a "not" comparison, the
// notPrefix, and, unless it is "is anything", the field for its qualifier, which shows the hint while empty. withAny
// offers "is anything", which a chip leaves to its remove button.
func (p *sentenceRows[T]) textCriteria(parent *unison.Panel, key, subject, hint, prefix, notPrefix string, c *criteria.Text, withAny bool) {
	comparison, _ := criteriaTitles(subject)
	items := criteria.StringComparisons
	if !withAny {
		items = items[1:]
	}
	var render func(criteria.StringComparison) string
	if prefix != "" {
		choices := criteria.PrefixedStringComparisonChoices(prefix, notPrefix)
		render = func(v criteria.StringComparison) string { return choices[v] }
	}
	addCentered(parent, compactPopup(p, key+"cmp", comparison, items, c.Compare, render,
		func(v criteria.StringComparison) { c.Compare = v }))
	if c.Compare != criteria.AnyText {
		p.textField(parent, key, subject, hint, &c.Qualifier)
	}
}

// numberCriteria adds the comparison of a numeric criterion, worded as numberCompare words it, and, unless it is
// "anything", the field for its qualifier, which takes whole numbers when integer is true.
func (p *sentenceRows[T]) numberCriteria(parent *unison.Panel, key, subject, prefix string, c *criteria.Number, minValue, maxValue fxp.Int, integer bool) {
	comparison, _ := criteriaTitles(subject)
	p.numberCompare(parent, key+"cmp", comparison, prefix, &c.Compare)
	if c.Compare == criteria.AnyNumber {
		return
	}
	set := func(v fxp.Int) { p.edit(subject, key, "", func() { c.Qualifier = v }) }
	if integer {
		p.addCompact(parent, NewIntegerField(p.targetMgr, key, subject, func() int { return c.Qualifier.AsInteger[int]() },
			func(v int) { set(fxp.FromInteger(v)) }, minValue.AsInteger[int](), maxValue.AsInteger[int](), false,
			false).withoutUndo())
		return
	}
	p.addCompact(parent, NewDecimalField(p.targetMgr, key, subject, func() fxp.Int { return c.Qualifier }, set, minValue,
		maxValue, false, false).withoutUndo())
}

// numberCompare adds the popup for a numeric comparison: with a prefix, each choice after it, as in "and whose level is
// at least"; otherwise in the short words that come before a quantity, as in "at least 2 spells".
func (p *sentenceRows[T]) numberCompare(parent *unison.Panel, key, name, prefix string, c *criteria.NumericComparison) {
	items := criteria.NumericComparisons
	// "anything" is left to a chip's remove button, but is shown when a file edited by hand holds it.
	if *c != criteria.AnyNumber {
		items = items[1:]
	}
	choices := criteria.PrefixedNumericComparisonChoices(prefix)
	addCentered(parent, compactPopup(p, key, name, items, *c, func(v criteria.NumericComparison) string {
		if prefix != "" {
			return choices[v]
		}
		if v == criteria.EqualsNumber {
			return i18n.Text("exactly")
		}
		return v.AltString()
	}, func(v criteria.NumericComparison) { *c = v }))
}

// textField adds a compact field for the text, which shows the hint while empty.
func (p *sentenceRows[T]) textField(parent *unison.Panel, key, title, hint string, value *string) *StringField {
	field := NewStringField(p.targetMgr, key, title, func() string { return *value },
		func(s string) { p.edit(title, key, "", func() { *value = s }) })
	field.Watermark = hint
	field.SetMinimumTextWidthUsing(i18n.Text("Weapon Master (Sword)"))
	p.addCompact(parent, field.withoutUndo())
	return field
}

// addCompact adds the field, which leaves undo to edit, to the parent with rounded borders.
func (p *sentenceRows[T]) addCompact(parent *unison.Panel, field *unison.Field) {
	addCentered(parent, field)
	unison.UninstallFocusBorders(field, field)
	// Padded to the height of the controls around it, less the height of its text.
	pad := (controlHeight(parent) - field.Font.LineHeight()) / 2
	unison.InstallFocusBorders(field, field, compactFieldBorder(true, pad), compactFieldBorder(false, pad))
	radius := geom.NewUniformSize(compactCornerRadius)
	draw := field.DrawCallback
	field.DrawCallback = func(gc *unison.Canvas, dirty geom.Rect) {
		gc.Save()
		path := unison.NewPath()
		path.RoundedRect(field.ContentRect(true), radius)
		gc.ClipPath(path, pathop.Intersect, true)
		draw(gc, dirty)
		gc.Restore()
	}
}

// compactFieldBorder returns the border of a compact field, which leaves pad above and below its text.
func compactFieldBorder(focused bool, pad float32) unison.Border {
	ink := unison.Ink(unison.ThemeSurfaceEdge)
	var w float32 = 1
	if focused {
		ink = unison.ThemeFocus
		w = 2
	}
	return unison.NewCompoundBorder(unison.NewLineBorder(ink, geom.NewUniformSize(compactCornerRadius),
		geom.NewUniformInsets(w), false),
		unison.NewEmptyBorder(geom.Insets{Top: pad - w, Left: 6 - w, Bottom: pad - w, Right: 6 - w}))
}

// controlHeight returns the height of a control in the parent: within a chip, that of a standard button; anywhere else,
// that of a chip holding one, so that everything on a line, chips and group pills included, is as tall.
func controlHeight(parent *unison.Panel) float32 {
	theme := &unison.DefaultButtonTheme
	height := xmath.Ceil(theme.Font.LineHeight()) + 2*(theme.VMargin+1)
	if parent != nil {
		if _, inChip := parent.Self.(*chipPanel); inChip {
			return height
		}
	}
	return height + 2*chipPadding
}

// fitLine has the control take the height controlHeight gives it in its parent.
func fitLine(control unison.Paneler) {
	panel := control.AsPanel()
	sizer := panel.Sizer()
	panel.SetSizer(func(hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
		minSize, prefSize, maxSize = sizer(hint)
		height := controlHeight(panel.Parent())
		minSize.Height, prefSize.Height, maxSize.Height = height, height, height
		return minSize, prefSize, maxSize
	})
}

// compactPopup returns a popup offering the items, as render shows them, with current selected. A choice is handed to
// set within an edit of the rows named name, after which they are rebuilt.
func compactPopup[S any, T comparable](p *sentenceRows[S], key, name string, items []T, current T, render func(T) string, set func(T)) *unison.PopupMenu[T] {
	popup := unison.NewPopupMenu[T]()
	popup.RefKey = key
	popup.Accessibility.Name = name
	popup.ItemRendererCallback = render
	popup.HMargin = 6
	popup.AddItem(items...)
	// Sized to the choice showing rather than the widest, since any choice rebuilds the panel.
	popup.SetSizer(func(hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
		_, prefSize, _ = popup.DefaultSizes(hint)
		prefSize.Height = controlHeight(popup.Parent())
		text := unison.NewText(popup.Text(), &unison.TextDecoration{Font: popup.Font})
		prefSize.Width = xmath.Ceil(text.Width() + popup.HMargin*2 + 4 + prefSize.Height*0.75)
		return prefSize, prefSize, prefSize
	})
	installPopupSelection(popup, current, func(v T) {
		p.edit(name, key, "", func() { set(v) })
		p.rebuild(key)
	})
	return popup
}

// addJoiningWords adds text that joins the controls of an open row into a sentence.
func addJoiningWords(parent *unison.Panel, text string) {
	label := unison.NewLabel()
	label.SetTitle(text)
	addCentered(parent, label)
}

// addCentered adds the child to the parent, centered vertically on its line.
func addCentered(parent *unison.Panel, child unison.Paneler) {
	switch parent.Layout().(type) {
	case *unison.FlowLayout, *hangingFlow:
		child.AsPanel().SetLayoutData(align.Middle)
	default:
		child.AsPanel().SetLayoutData(&unison.FlexLayoutData{VAlign: align.Middle})
	}
	parent.AddChild(child)
}

func newColumn() *unison.Panel {
	column := unison.NewPanel()
	column.SetLayout(&unison.FlexLayout{Columns: 1, VSpacing: unison.StdVSpacing})
	column.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	return column
}

func newFlow() *unison.Panel {
	flow := newColumn()
	flow.SetLayout(&unison.FlowLayout{HSpacing: 6, VSpacing: 6})
	return flow
}

// hangingFlow is a flow layout whose lines after the first are indented by indent, which the panel's border holds.
type hangingFlow struct {
	unison.FlowLayout
	indent float32
}

// newHangingFlow returns a flow whose first line starts at its left edge and whose others are indented by indent.
func newHangingFlow(indent float32) *unison.Panel {
	flow := newColumn()
	flow.SetLayout(&hangingFlow{HSpacing: 6, VSpacing: 6, indent: indent})
	flow.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: indent}))
	return flow
}

// PerformLayout implements unison.Layout.
func (h *hangingFlow) PerformLayout(target *unison.Panel) {
	h.FlowLayout.PerformLayout(target)
	// The first line ends where a child starts back at the left.
	x := float32(-1)
	for _, child := range target.Children() {
		r := child.FrameRect()
		if r.X <= x {
			break
		}
		x = r.X
		r.X -= h.indent
		child.SetFrameRect(r)
	}
}

func newIconButton(key string, icon *unison.SVG, tooltip string) *unison.Button {
	b := unison.NewSVGButton(icon)
	b.RefKey = key
	b.Tooltip = newWrappedTooltip(tooltip)
	return b
}

// newEmptyPlaceholder returns the dashed box, keyed key, that stands in for the rows of a panel or group that has none,
// saying text and calling click when clicked. It fills its line.
func newEmptyPlaceholder(key, text string, click func()) *unison.Button {
	empty := newDashedButton(text, click)
	empty.HAlign = align.Start
	empty.VMargin = 6
	empty.CornerRadius = geom.NewUniformSize(6)
	// The focus ring insets the text by 2.5, which moves text drawn from the start; take that out of the margins.
	draw := empty.DrawCallback
	empty.DrawCallback = func(gc *unison.Canvas, dirty geom.Rect) {
		if empty.Focused() {
			empty.HMargin, empty.VMargin = empty.HMargin-2.5, empty.VMargin-2.5
			defer func() { empty.HMargin, empty.VMargin = empty.HMargin+2.5, empty.VMargin+2.5 }()
		}
		draw(gc, dirty)
	}
	empty.RefKey = key
	empty.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	return empty
}

// newDashedButton returns a button drawn as a dashed outline, which adds something. Under the pointer it is filled, and
// its outline is solid and in the focus color.
func newDashedButton(title string, click func()) *unison.Button {
	b := unison.NewButton()
	b.HideBase = true
	b.OnBackgroundInk = unison.ThemeOnSurface
	b.CornerRadius = geom.NewUniformSize(pillCornerRadius)
	b.SetTitle(title)
	b.ClickCallback = click
	var hovered bool
	hover := func(on bool) bool {
		hovered = on
		b.MarkForRedraw()
		return true
	}
	b.MouseEnterCallback = func(_ geom.Point, _ mod.Modifiers) bool { return hover(true) }
	b.MouseExitCallback = func() bool { return hover(false) }
	draw := b.DrawCallback
	b.DrawCallback = func(gc *unison.Canvas, dirty geom.Rect) {
		r := b.ContentRect(true).Inset(geom.NewUniformInsets(0.5))
		radius := geom.NewUniformSize(min(b.CornerRadius.Width, r.Height/2))
		// Half as strong as text, for a contrast of at least 3:1 with the surface and what is below it.
		var edge unison.Ink = &unison.ColorFilteredInk{OriginalInk: unison.ThemeOnSurface, ColorFilter: unison.Alpha50Filter()}
		if hovered {
			gc.DrawRoundedRect(r, radius, unison.ThemeAboveSurface.Paint(gc, r, paintstyle.Fill))
			edge = unison.ThemeFocus
		}
		draw(gc, dirty)
		paint := edge.Paint(gc, r, paintstyle.Stroke)
		if !hovered {
			paint.SetPathEffect(unison.NewDashPathEffect([]float32{3, 3}, 0))
		}
		gc.DrawRoundedRect(r, radius, paint)
	}
	return b
}
