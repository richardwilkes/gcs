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
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellcmp"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/pathop"
	"github.com/richardwilkes/unison/enums/role"
	"github.com/richardwilkes/unison/enums/weight"
)

// A node of the tree is found by its path: prereqRootPath for the root, then the index of each child on the way down,
// separated by dots. A widget's reference key is the path of its node followed by one of the suffixes below, or by a
// colon and a name of its own.
const (
	prereqRootPath = "r"
	keyAdd         = ":add"
	keyMore        = ":more"
	keyPill        = ":pill"
	keySentence    = ":sentence"
	// keyChip follows the key of a chip's criterion, so that the chip doesn't share the key of the field within it.
	keyChip = ":chip"
	// keyFirst marks the editor of an open row; focusing it focuses the first control within it.
	keyFirst = ":first"
)

// prereqDragKey is the drag data type of a prerequisite being dragged to another place in its tree.
var prereqDragKey = unison.CreatePrivateDataType("gcs.prereq")

// prereqDropKey marks what a dragged prerequisite can be dropped on, holding the path of the node it shows.
const prereqDropKey = "prereq.drop"

// Where a dragged prerequisite goes in relation to what it is dropped on.
const (
	dropBefore = iota
	dropAfter
	dropInto
)

var (
	prereqAddSVG  = unison.MustSVGFromContentString(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 448 512"><path d="M256 80c0-17.7-14.3-32-32-32s-32 14.3-32 32v144H48c-17.7 0-32 14.3-32 32s14.3 32 32 32h144v144c0 17.7 14.3 32 32 32s32-14.3 32-32V288h144c17.7 0 32-14.3 32-32s-14.3-32-32-32H256V80z"/></svg>`)
	prereqMoreSVG = unison.MustSVGFromContentString(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 448 512"><path d="M8 256a56 56 0 1 1 112 0 56 56 0 1 1-112 0zm160 0a56 56 0 1 1 112 0 56 56 0 1 1-112 0zm216-56a56 56 0 1 1 0 112 56 56 0 1 1 0-112z"/></svg>`)
)

// prereqPanel edits a tree of prerequisites. Each one is a row that reads as a sentence until it is opened, one at a
// time, to edit it. Each list is a group, whose head says whether all or any of its children must be met and whose
// children hang from a rail in the head's color. Every change, typing included, records a snapshot of the whole tree
// with the editor's undo manager, and changes to the tree's shape rebuild the panel's content.
type prereqPanel struct {
	unison.Panel
	entity           *gurps.Entity
	root             **gurps.PrereqList
	placeholder      *gurps.PrereqList
	permittedChoices []prereq.Type
	targetMgr        *TargetMgr
	discard          *unison.UndoManager
	summary          *sentenceButton
	views            []prereqView
	target           any
	open             string
	follow           gurps.Prereq
	focus            string
	editKey          string
	editID           int64
	dropTarget       *unison.Panel
	dropWhere        int
	hash             uint64
	ownerIsSpell     bool
	muted            bool
	rebuilding       bool
}

// prereqView is what shows a node's status and sentence, or a group's name, refreshed in place as the tree changes.
type prereqView struct {
	node     gurps.Prereq
	icon     *unison.Label
	sentence *sentenceButton
	group    *unison.Panel
}

func newPrereqPanel(entity *gurps.Entity, root **gurps.PrereqList, permittedChoices []prereq.Type, ownerIsSpell bool) *prereqPanel {
	p := &prereqPanel{
		entity:           entity,
		root:             root,
		permittedChoices: permittedChoices,
		ownerIsSpell:     ownerIsSpell,
		discard:          unison.NewUndoManager(1, func(error) {}),
	}
	initTitledEditorSection(p, i18n.Text("Prerequisites"))
	p.targetMgr = NewTargetMgr(p)
	installPanelDragDrop(p.AsPanel(), prereqDragKey, p.dragOver, p.dragExit, p.drop)
	p.DrawOverCallback = p.drawDrop
	p.build()
	// The item being edited, which evaluation leaves out, can only be found once the panel is in its editor.
	unison.InvokeTask(p.refresh)
	return p
}

// UndoManager implements unison.UndoManagerProvider. A field reporting a change gets a manager that is thrown away,
// since edit records the change as a snapshot of the tree; at any other time the editor's is used.
func (p *prereqPanel) UndoManager() *unison.UndoManager {
	if p.muted {
		return p.discard
	}
	return nil
}

// tree returns the list being edited. A missing list is stood in for by an empty one, which becomes the real one when
// it is first changed, so that merely showing the panel changes nothing.
func (p *prereqPanel) tree() *gurps.PrereqList {
	if *p.root != nil {
		return *p.root
	}
	if p.placeholder == nil {
		p.placeholder = gurps.NewPrereqList()
	}
	return p.placeholder
}

// prereqSnapshot is the tree as it stood before or after an edit, and the reference key of the widget that undo or redo
// returning to it gives the focus to.
type prereqSnapshot struct {
	tree  *gurps.PrereqList
	focus string
}

func cloneTree(list *gurps.PrereqList) *gurps.PrereqList {
	if list == nil {
		return nil
	}
	return list.CloneAsPrereqList(nil)
}

// edit applies change to the tree and records it with the editor's undo manager, under the title, as snapshots of the
// tree before and after. key is the reference key of the widget making the change, which undo and redo give the focus
// to. Consecutive edits with the same non-empty key, such as each keystroke typed into a field, are recorded as one
// until the panel is next rebuilt. It returns the snapshot after, or nil if nothing changed, so that a change that
// moves the focus can have redo move it there too.
func (p *prereqPanel) edit(title, key string, change func()) *prereqSnapshot {
	before := &prereqSnapshot{tree: cloneTree(*p.root), focus: key}
	hash := gurps.Hash64(p.tree())
	*p.root = p.tree()
	change()
	if gurps.Hash64(*p.root) == hash {
		return nil
	}
	if key == "" || key != p.editKey {
		p.editID = unison.NextUndoID()
	}
	p.editKey = key
	after := &prereqSnapshot{tree: cloneTree(*p.root), focus: key}
	if parent := p.Parent(); parent != nil {
		if mgr := unison.UndoManagerFor(parent); mgr != nil {
			mgr.Add(&unison.UndoEdit[*prereqSnapshot]{
				ID:       p.editID,
				EditName: title,
				EditCost: 1,
				UndoFunc: func(e *unison.UndoEdit[*prereqSnapshot]) { p.install(e.BeforeData) },
				RedoFunc: func(e *unison.UndoEdit[*prereqSnapshot]) { p.install(e.AfterData) },
				AbsorbFunc: func(e *unison.UndoEdit[*prereqSnapshot], other unison.Undoable) bool {
					if o, ok := other.(*unison.UndoEdit[*prereqSnapshot]); ok && o.ID == e.ID {
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

// install makes a copy of the snapshot's tree the tree being edited, and gives the focus to its widget, for undo and
// redo.
func (p *prereqPanel) install(snapshot *prereqSnapshot) {
	*p.root = cloneTree(snapshot.tree)
	p.placeholder = nil
	MarkModified(p)
	p.rebuild(snapshot.focus)
}

// restructure changes the shape of the tree through the widget with the reference key from, then rebuilds. change
// returns the node that is the result, whose more button takes the focus; with none, the widget with the reference key
// fallback does. The open row stays open wherever it ends up, so change copies nodes with clone, and closes if it is
// removed.
func (p *prereqPanel) restructure(title, from, fallback string, change func() (dst gurps.Prereq)) {
	p.follow = p.node(p.open)
	var dst gurps.Prereq
	after := p.edit(title, from, func() { dst = change() })
	p.open = p.pathOf(p.follow)
	p.follow = nil
	if dst != nil {
		fallback = p.pathOf(dst) + keyMore
	}
	if after != nil {
		after.focus = fallback
	}
	p.rebuild(fallback)
}

// rebuild replaces the panel's content once the current event has been handled, since that may have come from a widget
// about to be thrown away. The scroll position is kept, and the focus goes to the widget with the reference key focus
// or, when that is empty, back to the widget that held it.
func (p *prereqPanel) rebuild(focus string) {
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
		p.build()
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
func (p *prereqPanel) focusOn(key string) bool {
	target := p.FindRefKey(key)
	if target == nil {
		return false
	}
	if !target.Focusable() {
		if target = target.FirstFocusableChild(); target == nil {
			return false
		}
	}
	target.RequestFocus()
	target.ScrollIntoView()
	return true
}

func (p *prereqPanel) build() {
	if node := p.node(p.open); node == nil || node.PrereqType() == prereq.List {
		p.open = ""
	}
	p.views = p.views[:0]
	p.summary = newSentenceButton("", nil, nil)
	p.summary.SetBorder(unison.NewCompoundBorder(
		unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.Insets{Bottom: 1}, false),
		unison.NewEmptyBorder(geom.Insets{Top: 4, Left: 4, Bottom: 6, Right: 4}),
	))
	p.summary.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	p.AddChild(p.summary)
	p.AddChild(p.group(p.tree(), prereqRootPath))
	p.refresh()
}

// Sync implements Syncer. Statuses are worked out again only when the tree has changed, since that runs its scripts.
func (p *prereqPanel) Sync() {
	if p.hash != gurps.Hash64(p.tree()) {
		p.refresh()
	}
}

// refresh updates the summary, the sentences and the status icons from the tree, in place.
func (p *prereqPanel) refresh() {
	tree := p.tree()
	p.hash = gurps.Hash64(tree)
	summary := i18n.Text("No prerequisites.")
	if len(tree.Prereqs) != 0 {
		summary = gurps.DescribePrereq(p.entity, tree, em)
	}
	p.summary.setText(summary, "")
	// Until the panel is in its editor, the item being edited, which evaluation leaves out, can't be found.
	check := p.entity != nil && p.Parent() != nil
	for _, v := range p.views {
		var suffix string
		if check {
			var status checkStatus
			var tip string
			status, tip, suffix = p.status(v.node)
			showCheckIcon(v.icon, status)
			v.icon.Tooltip = newWrappedTooltip(tip)
		}
		if v.sentence != nil {
			v.sentence.setText(gurps.DescribePrereq(p.entity, v.node, em), suffix)
		}
		if list, ok := v.node.(*gurps.PrereqList); ok && v.group != nil {
			v.group.Accessibility.Name = groupName(list)
		}
	}
}

// status returns the node's status against the sheet, the tooltip of its icon and what a screen reader hears after its
// sentence. Everything within a list that doesn't apply at the sheet's tech level is skipped.
func (p *prereqPanel) status(node gurps.Prereq) (status checkStatus, tip, suffix string) {
	list, ok := node.(*gurps.PrereqList)
	if !ok {
		list = node.ParentList()
	}
	for ; list != nil; list = list.Parent {
		if !list.AppliesAt(p.entity) {
			return checkSkipped, i18n.Text("Doesn't apply at this tech level"), i18n.Text("doesn't apply at this tech level")
		}
	}
	var met, failed bool
	var reason string
	if script, isScript := node.(*gurps.ScriptPrereq); isScript {
		gurps.SuppressScriptResolveErrorLogging(func() { met, reason, failed = script.Evaluate(p.entity, p.exclude()) })
	} else {
		var buffer xbytes.InsertBuffer
		met = node.Satisfied(p.entity, p.exclude(), &buffer, "", nil)
		reason = strings.TrimSpace(buffer.String())
	}
	switch {
	case failed:
		return checkFailed, reason, fmt.Sprintf(i18n.Text("script error: %s"), reason)
	case met:
		return checkMet, i18n.Text("Met"), i18n.Text("met")
	default:
		return checkUnmet, reason, i18n.Text("not met")
	}
}

// statusIcon adds the icon for a node's status to the parent when there is a sheet. Screen readers skip it, since the
// accessible name of the node's sentence says the status.
func (p *prereqPanel) statusIcon(parent *unison.Panel) *unison.Label {
	if p.entity == nil {
		return nil
	}
	icon := unison.NewLabel()
	icon.Accessibility.Role = role.None
	icon.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: 4}))
	put(parent, icon)
	return icon
}

func childPath(path string, i int) string {
	return path + "." + strconv.Itoa(i)
}

// node returns the prerequisite at the path, or nil if there is none.
func (p *prereqPanel) node(path string) gurps.Prereq {
	parts := strings.Split(path, ".")
	if parts[0] != prereqRootPath {
		return nil
	}
	var node gurps.Prereq = p.tree()
	for _, part := range parts[1:] {
		list, ok := node.(*gurps.PrereqList)
		i, err := strconv.Atoi(part)
		if !ok || err != nil || i < 0 || i >= len(list.Prereqs) {
			return nil
		}
		node = list.Prereqs[i]
	}
	return node
}

// locate returns the list holding the node at the path, other than the root, and the node's index within it.
func (p *prereqPanel) locate(path string) (parent *gurps.PrereqList, index int) {
	cut := strings.LastIndexByte(path, '.')
	index, err := strconv.Atoi(path[cut+1:])
	if list, ok := p.node(path[:cut]).(*gurps.PrereqList); ok && err == nil {
		return list, index
	}
	return nil, -1
}

// pathOf returns the path of the node, or an empty string if it is not in the tree.
func (p *prereqPanel) pathOf(target gurps.Prereq) string {
	if target == nil {
		return ""
	}
	var find func(node gurps.Prereq, path string) string
	find = func(node gurps.Prereq, path string) string {
		if node == target {
			return path
		}
		if list, ok := node.(*gurps.PrereqList); ok {
			for i, child := range list.Prereqs {
				if found := find(child, childPath(path, i)); found != "" {
					return found
				}
			}
		}
		return ""
	}
	return find(p.tree(), prereqRootPath)
}

// clone returns a copy of the node made for the parent. When the open row is the node or within it, it follows the
// copy.
func (p *prereqPanel) clone(node gurps.Prereq, parent *gurps.PrereqList) gurps.Prereq {
	copied := node.Clone(parent)
	var follow func(from, to gurps.Prereq)
	follow = func(from, to gurps.Prereq) {
		if from == p.follow {
			p.follow = to
			return
		}
		fromList, isList := from.(*gurps.PrereqList)
		toList, isCopy := to.(*gurps.PrereqList)
		if isList && isCopy {
			for i, child := range fromList.Prereqs {
				follow(child, toList.Prereqs[i])
			}
		}
	}
	follow(node, copied)
	return copied
}

// groupName returns the accessible name of a group.
func groupName(list *gurps.PrereqList) string {
	name := groupWord(list.All)
	if list.WhenTL.Compare != criteria.AnyNumber {
		name += fmt.Sprintf(i18n.Text(", only when TL %s"), list.WhenTL.AltString())
	}
	return name
}

func groupWord(all bool) string {
	if all {
		return i18n.Text("All of")
	}
	return i18n.Text("Any of")
}

// group returns the panel for a list: its head, over its children hanging from a rail in the head's color.
func (p *prereqPanel) group(list *gurps.PrereqList, path string) *unison.Panel {
	color := colors.Grouping2
	if list.All {
		color = colors.Grouping1
	}
	box := newPrereqColumn()
	box.Accessibility.Role = role.Group
	box.Accessibility.Name = groupName(list)
	head := unison.NewPanel()
	head.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: 2, Bottom: 2}))
	pill := compactPopup(p, path+keyPill, i18n.Text("Requirement"), []bool{true, false}, list.All, groupWord,
		func(all bool) { list.All = all })
	pill.HMargin = 10
	pill.CornerRadius = geom.NewUniformSize(100)
	pill.BackgroundInk = faint(color)
	pill.OnBackgroundInk = color
	pill.EdgeInk = unison.Transparent
	desc := pill.Font.Descriptor()
	desc.Weight = weight.Bold
	pill.Font = desc.Font()
	head.ClientData()[prereqDropKey] = path
	if path != prereqRootPath {
		put(head, NewDragHandle(prereqDragKey, &prereqDrag{panel: p, path: path}))
	}
	p.views = append(p.views, prereqView{node: list, icon: p.statusIcon(head), group: box})
	put(head, pill)
	if list.WhenTL.Compare != criteria.AnyNumber {
		p.chip(head, path+":tl", i18n.Text("When TL"), path+keyAdd, func() { list.WhenTL = criteria.Number{} },
			func(chip *unison.Panel) {
				p.numberCriteria(chip, path+":tl", i18n.Text("Tech Level"), &list.WhenTL, false, 0, fxp.Twelve, true)
			})
	}
	add := newPrereqIconButton(path+keyAdd, prereqAddSVG, i18n.Text("Add to this group"))
	add.ClickCallback = func() { showMenu(add.AsPanel(), p.addEntries(list, path)) }
	add.SetLayoutData(&unison.FlexLayoutData{HAlign: align.End, VAlign: align.Middle, HGrab: true})
	head.AddChild(add)
	if path != prereqRootPath {
		p.moreButton(head, list, path)
	}
	box.AddChild(hbox(head, unison.StdHSpacing))

	rail := newPrereqColumn()
	rail.SetBorder(unison.NewCompoundBorder(unison.NewEmptyBorder(geom.Insets{Left: 10}),
		unison.NewLineBorder(color, geom.Size{}, geom.Insets{Left: 3}, false),
		unison.NewEmptyBorder(geom.Insets{Left: 10, Bottom: 2})))
	for i, child := range list.Prereqs {
		if sub, ok := child.(*gurps.PrereqList); ok {
			rail.AddChild(p.group(sub, childPath(path, i)))
		} else {
			rail.AddChild(p.row(child, childPath(path, i)))
		}
	}
	if len(list.Prereqs) == 0 {
		text := i18n.Text("Empty group. Add a requirement or drag one here.")
		if path == prereqRootPath {
			text = i18n.Text("No prerequisites. Add one to get started.")
		}
		empty := newDashedButton(text, nil)
		empty.ClickCallback = func() { showMenu(empty.AsPanel(), p.addEntries(list, path)) }
		empty.HAlign = align.Start
		empty.VMargin = 6
		empty.CornerRadius = geom.NewUniformSize(6)
		empty.RefKey = path + ":empty"
		empty.ClientData()[prereqDropKey] = path
		empty.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
		rail.AddChild(empty)
	}
	box.AddChild(rail)
	return box
}

// row returns the panel for a prerequisite other than a list: its sentence, or while it is open its editor, beside a
// button for more actions.
func (p *prereqPanel) row(pr gurps.Prereq, path string) *unison.Panel {
	open := path == p.open
	row := unison.NewPanel()
	row.ClientData()[prereqDropKey] = path
	put(row, NewDragHandle(prereqDragKey, &prereqDrag{panel: p, path: path}))
	view := prereqView{node: pr, icon: p.statusIcon(row)}
	var main *unison.Panel
	if open {
		// No inset on the left, so that the grip and status icon line up with those of closed rows.
		row.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: 6, Bottom: 8, Right: 4}))
		main = p.editor(pr, path)
		row.AddChild(main)
		done := unison.NewButton()
		done.SetTitle(i18n.Text("Done"))
		done.RefKey = path + ":done"
		done.ClickCallback = func() { p.toggle(path) }
		row.AddChild(done)
		row.KeyDownCallback = func(keyCode unison.KeyCode, mods mod.Modifiers, _ bool) bool {
			if keyCode != unison.KeyEscape || !noModifiersDown(mods) {
				return false
			}
			p.toggle(path)
			return true
		}
	} else {
		row.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: 3, Bottom: 3, Right: 4}))
		var click func()
		if pr.PrereqType() != prereq.Unknown {
			click = func() { p.toggle(path) }
		}
		view.sentence = newSentenceButton(gurps.DescribePrereq(p.entity, pr, em), click, func() bool { return p.open == path })
		view.sentence.RefKey = path + keySentence
		if click == nil {
			// One this version of GCS doesn't understand can't be edited, so its sentence is static text.
			view.sentence.Tooltip = newWrappedTooltip(i18n.Text("This was most likely created by a newer version of GCS. Its original data will be written back out unchanged when this file is saved."))
		}
		main = view.sentence.AsPanel()
		row.AddChild(main)
		if script, ok := pr.(*gurps.ScriptPrereq); ok && strings.TrimSpace(script.Name) == "" {
			describe := newDashedButton(i18n.Text("Add a description"), func() {
				p.open = path
				p.rebuild(path + ":name")
			})
			describe.OnBackgroundInk = unison.ThemeWarning
			put(row, describe)
		}
	}
	p.views = append(p.views, view)
	p.moreButton(row, pr, path)
	hbox(row, unison.StdHSpacing)
	for _, child := range row.Children() {
		if child == main {
			child.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, VAlign: align.Middle, HGrab: true})
		} else if open {
			child.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Start})
		}
	}
	row.DrawCallback = func(gc *unison.Canvas, _ geom.Rect) {
		r := row.ContentRect(true)
		if open {
			gc.DrawRoundedRect(r, geom.NewUniformSize(8), unison.ThemeBelowSurface.Paint(gc, r, paintstyle.Fill))
			return
		}
		gc.DrawLine(geom.NewPoint(r.X, r.Bottom()-0.5), geom.NewPoint(r.Right(), r.Bottom()-0.5),
			faint(unison.ThemeSurfaceEdge).Paint(gc, r, paintstyle.Stroke))
	}
	return row
}

// toggle opens the row at the path, closing any other, or closes it if it is the open one. The focus goes to the first
// control of the row that opens, or to the sentence of the one that closes.
func (p *prereqPanel) toggle(path string) {
	if p.open == path {
		p.open = ""
		p.rebuild(path + keySentence)
	} else {
		p.open = path
		p.rebuild(path + keyFirst)
	}
}

// editor returns the controls for an open row: those that say what the prerequisite requires, flowing as a sentence
// would, then its optional criteria as chips, or the script editor for a script.
func (p *prereqPanel) editor(pr gurps.Prereq, path string) *unison.Panel {
	box := newPrereqColumn()
	box.RefKey = path + keyFirst
	fields := newPrereqFlow()
	box.AddChild(fields)
	chips := newPrereqFlow()
	key := func(name string) string { return path + ":" + name }
	switch one := pr.(type) {
	case *gurps.TraitPrereq:
		p.hasPopup(fields, key("has"), &one.Has, false)
		p.typePopup(fields, path, pr)
		p.textCriteria(fields, key("name"), i18n.Text("Name"), i18n.Text("Trait name"), &one.NameCriteria, true)
		level := &one.LevelCriteria
		// Every trait has a level of at least 0, so that is the same as having no level criteria.
		p.optional(chips, path, i18n.Text("level"),
			level.Compare != criteria.AnyNumber && (level.Compare != criteria.AtLeastNumber || level.Qualifier > 0),
			func() { *level = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.One} },
			func() { *level = criteria.Number{} },
			func(chip *unison.Panel) {
				p.numberCriteria(chip, key("level"), i18n.Text("Level"), level, false, 0, fxp.Thousand, false)
			})
		p.textChip(chips, path, i18n.Text("notes"), i18n.Text("Notes"), &one.NotesCriteria)
	case *gurps.SkillPrereq:
		p.hasPopup(fields, key("has"), &one.Has, false)
		p.typePopup(fields, path, pr)
		p.textCriteria(fields, key("name"), i18n.Text("Name"), i18n.Text("Skill name"), &one.NameCriteria, true)
		p.numberCriteria(fields, key("level"), i18n.Text("Level"), &one.LevelCriteria, true, 0, fxp.Thousand, false)
		p.textChip(chips, path, i18n.Text("specialization"), i18n.Text("Specialization"), &one.SpecializationCriteria)
		p.textChip(chips, path, i18n.Text("optional specialization"), i18n.Text("Optional Specialization"),
			&one.OptionalSpecializationCriteria)
	case *gurps.SpellPrereq:
		p.hasPopup(fields, key("has"), &one.Has, true)
		p.numberCriteria(fields, key("quantity"), i18n.Text("Quantity"), &one.QuantityCriteria, false, 0,
			fxp.FromInteger(9999), true)
		p.typePopup(fields, path, pr)
		put(fields, compactPopup(p, key("match"), i18n.Text("Spell Match"), spellcmp.Types, one.SubType, nil,
			func(t spellcmp.Type) { one.SubType = t }))
		if one.SubType.UsesStringCriteria() {
			p.textCriteria(fields, key("qualifier"), i18n.Text("Spell"), "", &one.QualifierCriteria, true)
		}
		p.powerSourceChip(chips, path, one)
	case *gurps.AttributePrereq:
		p.hasPopup(fields, key("has"), &one.Has, false)
		p.typePopup(fields, path, pr)
		flags := gurps.SizeFlag | gurps.DodgeFlag | gurps.ParryFlag | gurps.BlockFlag
		p.attributePopup(fields, key("which"), i18n.Text("Attribute"), &one.Which, flags)
		p.numberCriteria(fields, key("value"), i18n.Text("Attribute"), &one.QualifierCriteria, false, fxp.Min, fxp.Max,
			false)
		p.optional(chips, path, i18n.Text("combined with"), one.CombinedWith != "",
			func() { one.CombinedWith = gurps.AttributeIDFor(p.entity, gurps.DexterityID) },
			func() { one.CombinedWith = "" },
			func(chip *unison.Panel) {
				p.attributePopup(chip, key("combined"), i18n.Text("Combined With"), &one.CombinedWith, flags)
			})
	case *gurps.EquippedEquipmentPrereq:
		p.typePopup(fields, path, pr)
		p.textCriteria(fields, key("name"), i18n.Text("Name"), i18n.Text("Item name"), &one.NameCriteria, true)
		p.textChip(chips, path, i18n.Text("tag"), i18n.Text("Tag"), &one.TagsCriteria)
	case *gurps.ContainedQuantityPrereq:
		p.hasPopup(fields, key("has"), &one.Has, false)
		p.typePopup(fields, path, pr)
		p.numberCriteria(fields, key("quantity"), i18n.Text("Quantity"), &one.QualifierCriteria, false, 0,
			fxp.FromInteger(9999), true)
	case *gurps.ContainedWeightPrereq:
		p.hasPopup(fields, key("has"), &one.Has, false)
		p.typePopup(fields, path, pr)
		title := i18n.Text("Weight")
		comparison, _ := criteriaTitles(title)
		p.numberCompare(fields, key("weightcmp"), comparison, &one.WeightCriteria.Compare, false)
		p.addCompact(fields, NewWeightField(p.targetMgr, key("weight"), title, p.entity,
			func() fxp.Weight { return one.WeightCriteria.Qualifier },
			func(w fxp.Weight) { p.edit(title, key("weight"), func() { one.WeightCriteria.Qualifier = w }) },
			0, fxp.Weight(fxp.Max), false).Field)
	case *gurps.ScriptPrereq:
		p.typePopup(fields, path, pr)
		name := p.textField(fields, key("name"), i18n.Text("Description"),
			i18n.Text(`Describe this requirement, like "DX + Per totals at least 26"`), &one.Name)
		name.SetMinimumTextWidthUsing(name.Watermark)
		e := newScriptEditor(func() string { return one.Script },
			func(script string) { p.edit(i18n.Text("Script"), key("script"), func() { one.Script = script }) },
			p.scriptOptions(one))
		e.field.RefKey = key("script")
		e.field.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
		p.mute(e.field.Field)
		box.AddChild(e)
	default:
	}
	if len(chips.Children()) != 0 {
		box.AddChild(chips)
	}
	return box
}

// exclude returns the item being edited, which evaluation leaves out of what it looks at, or nil if it can't be found.
func (p *prereqPanel) exclude() any {
	if p.target == nil {
		panel := p.AsPanel()
		if t := FindTarget[*gurps.Trait](panel); t != nil {
			p.target = t
		} else if s := FindTarget[*gurps.Skill](panel); s != nil {
			p.target = s
		} else if sp := FindTarget[*gurps.Spell](panel); sp != nil {
			p.target = sp
		} else if e := FindTarget[*gurps.Equipment](panel); e != nil {
			p.target = e
		}
	}
	return p.target
}

// scriptOptions returns the script editor's options for a script prerequisite. Each snippet ends in an expression that
// is true when met and otherwise the reason it isn't. Evaluate, when there is a sheet, runs the script being edited
// against it.
func (p *prereqPanel) scriptOptions(pr *gurps.ScriptPrereq) scriptEditorOptions {
	opts := scriptEditorOptions{
		Title: i18n.Text("Script"),
		Inserts: []scriptMenuEntry{
			{Label: i18n.Text("The sheet"), Heading: true},
			{Label: `entity.attribute("st").current`, Text: `entity.attribute("st").current`},
			{Label: `entity.hasTrait("")`, Text: `entity.hasTrait("")`, CaretFromEnd: 2},
			{Label: `entity.traitLevel("")`, Text: `entity.traitLevel("")`, CaretFromEnd: 2},
			{Label: `entity.skillLevel("")`, Text: `entity.skillLevel("")`, CaretFromEnd: 2},
			{Label: `entity.findSkills("")`, Text: `entity.findSkills("")`, CaretFromEnd: 2},
			{Label: `entity.spells`, Text: `entity.spells`},
			{Label: `entity.equipment`, Text: `entity.equipment`},
			{Label: `entity.techLevel`, Text: `entity.techLevel`},
			{Label: i18n.Text("The item being edited"), Heading: true},
			{Label: `self.level`, Text: `self.level`},
		},
		Snippets: []scriptMenuEntry{
			{Label: i18n.Text("Replaces the script"), Heading: true},
			{Label: i18n.Text("Two attributes add up to at least a value"), Text: "const total = entity.attribute(\"dx\").current + entity.attribute(\"per\").current;\ntotal >= 26 || `DX + Per is ${total}, needs 26`"},
			{Label: i18n.Text("A trait at a level or more"), Text: "entity.traitLevel(\"Magery\") >= 2 || \"Needs Magery 2\""},
			{Label: i18n.Text("Any one of several traits"), Text: "const options = [\"Combat Reflexes\", \"Enhanced Dodge\"];\noptions.some(name => entity.hasTrait(name)) || `Needs one of: ${options.join(\", \")}`"},
			{Label: i18n.Text("The best of several skills"), Text: "const best = Math.max(...[\"Broadsword\", \"Shortsword\"].map(name => entity.skillLevel(name)));\nbest >= 14 || `Best sword skill is ${best}, needs 14`"},
		},
	}
	if p.entity != nil {
		opts.Evaluate = func(script string) (checkStatus, string) {
			one := *pr
			one.Script = script
			met, reason, failed := one.Evaluate(p.entity, p.exclude())
			switch {
			case failed:
				return checkFailed, reason
			case met:
				return checkMet, i18n.Text("Met by this sheet")
			default:
				return checkUnmet, reason
			}
		}
	}
	return opts
}

// addEntries returns the entries of a list's Add menu.
func (p *prereqPanel) addEntries(list *gurps.PrereqList, path string) []menuEntry {
	add := func(title string, created gurps.Prereq) {
		p.restructure(title, path+keyAdd, "", func() gurps.Prereq {
			list.Prereqs = append(list.Prereqs, created)
			return created
		})
		if created.PrereqType() != prereq.List {
			p.open = p.pathOf(created)
			p.rebuild(p.open + keyFirst)
		}
	}
	entries := []menuEntry{{Label: i18n.Text("Requirement")}}
	for _, t := range p.permittedChoices {
		entries = append(entries, menuEntry{Label: xstrings.FirstToUpper(prereqTypeName(t)), Act: func() {
			add(i18n.Text("Add Prerequisite"), p.createPrereqForType(t, list))
		}})
	}
	entries = append(entries, menuEntry{Label: i18n.Text("Structure")})
	for _, all := range []bool{true, false} {
		entries = append(entries, menuEntry{Label: fmt.Sprintf(i18n.Text("%s group"), groupWord(all)), Act: func() {
			add(i18n.Text("Add Group"), &gurps.PrereqList{Type: prereq.List, Parent: list, All: all})
		}})
	}
	if list.WhenTL.Compare == criteria.AnyNumber {
		entries = append(entries, menuEntry{Label: i18n.Text("Only when TL…"), Act: func() {
			if after := p.edit(i18n.Text("Add Tech Level Condition"), path+keyAdd, func() {
				list.WhenTL = criteria.Number{Compare: criteria.AtMostNumber, Qualifier: fxp.FromInteger(3)}
			}); after != nil {
				after.focus = path + ":tl" + keyChip
			}
			p.rebuild(path + ":tl" + keyChip)
		}})
	}
	return entries
}

// moreButton adds the button for the node's more menu.
func (p *prereqPanel) moreButton(parent *unison.Panel, node gurps.Prereq, path string) {
	b := newPrereqIconButton(path+keyMore, prereqMoreSVG, i18n.Text("More actions"))
	b.ClickCallback = func() { showMenu(b.AsPanel(), p.moreEntries(node, path)) }
	put(parent, b)
}

// moreEntries returns the entries of the node's more menu: Duplicate, Move up and down, which step into and out of
// groups, Wrap in group, Ungroup when that keeps what the group means, and Delete.
func (p *prereqPanel) moreEntries(node gurps.Prereq, path string) []menuEntry {
	list, i := p.locate(path)
	from := path + keyMore
	entries := []menuEntry{{Label: i18n.Text("Duplicate"), Act: func() {
		p.restructure(i18n.Text("Duplicate Prerequisite"), from, "", func() gurps.Prereq {
			dst := node.Clone(list)
			list.Prereqs = slices.Insert(list.Prereqs, i+1, dst)
			return dst
		})
	}}}
	titles := []string{i18n.Text("Move Up"), i18n.Text("Move Down")}
	for k, dir := range []int{-1, 1} {
		to, at, ok := p.moveTarget(path, dir)
		if !ok {
			continue
		}
		title := titles[k]
		entries = append(entries, menuEntry{Label: title, Act: func() {
			p.restructure(title, from, "", func() gurps.Prereq { return p.relocate(list, i, to, at) })
		}})
	}
	entries = append(entries, menuEntry{Label: i18n.Text("Wrap in Group"), Act: func() {
		p.restructure(i18n.Text("Wrap in Group"), from, "", func() gurps.Prereq {
			group := &gurps.PrereqList{Type: prereq.List, Parent: list, All: !list.All}
			dst := p.clone(node, group)
			group.Prereqs = gurps.Prereqs{dst}
			list.Prereqs[i] = group
			return dst
		})
	}})
	if g, ok := node.(*gurps.PrereqList); ok && len(g.Prereqs) != 0 && g.WhenTL.Compare == criteria.AnyNumber &&
		(g.All == list.All || len(g.Prereqs) == 1) {
		entries = append(entries, menuEntry{Label: i18n.Text("Ungroup"), Act: func() {
			p.restructure(i18n.Text("Ungroup"), from, "", func() gurps.Prereq {
				children := make(gurps.Prereqs, len(g.Prereqs))
				for j, child := range g.Prereqs {
					children[j] = p.clone(child, list)
				}
				list.Prereqs = slices.Replace(list.Prereqs, i, i+1, children...)
				return children[0]
			})
		}})
	}
	entries = append(entries, menuEntry{}, menuEntry{Label: i18n.Text("Delete"), Act: func() {
		parentPath := path[:strings.LastIndexByte(path, '.')]
		p.restructure(i18n.Text("Delete Prerequisite"), from, parentPath+keyAdd, func() gurps.Prereq {
			list.Prereqs = slices.Delete(list.Prereqs, i, i+1)
			if i < len(list.Prereqs) {
				return list.Prereqs[i]
			}
			return nil
		})
	}})
	return entries
}

// prereqDrag is the payload of a prerequisite being dragged: the panel it belongs to and its path there.
type prereqDrag struct {
	panel *prereqPanel
	path  string
}

// dragOver keeps the pointer in view and finds where a prerequisite dragged from this panel would go: before or after
// the row under the pointer, before a group when over the top of its head and into it below that, or into an empty
// group. Nothing can go into itself, and nothing can go before the root.
func (p *prereqPanel) dragOver(where geom.Point, data any) bool {
	p.ScrollRectIntoView(geom.NewRect(where.X, where.Y-16, 1, 1))
	p.ScrollRectIntoView(geom.NewRect(where.X, where.Y+16, 1, 1))
	target, at := p.dropAt(where, data)
	if target != p.dropTarget || at != p.dropWhere {
		p.dropTarget = target
		p.dropWhere = at
		p.MarkForRedraw()
	}
	return true
}

func (p *prereqPanel) dropAt(where geom.Point, data any) (target *unison.Panel, at int) {
	dd, ok := data.(*prereqDrag)
	if !ok || dd.panel != p || p.node(dd.path) == nil {
		return nil, 0
	}
	for target = p.PanelAt(where); target != nil && target != p.AsPanel(); target = target.Parent() {
		path, isTarget := target.ClientData()[prereqDropKey].(string)
		switch {
		case !isTarget:
			continue
		case path == dd.path || strings.HasPrefix(path, dd.path+"."):
			return nil, 0
		case target.RefKey == path+":empty":
			return target, dropInto
		}
		y := target.PointFromRoot(p.PointToRoot(where)).Y
		height := target.FrameRect().Height
		switch {
		case p.node(path).PrereqType() != prereq.List:
			if y < height/2 {
				return target, dropBefore
			}
			return target, dropAfter
		case path != prereqRootPath && y < height*0.3:
			return target, dropBefore
		default:
			return target, dropInto
		}
	}
	return nil, 0
}

func (p *prereqPanel) dragExit() {
	p.dropTarget = nil
	p.MarkForRedraw()
}

// drop moves the dragged prerequisite to where dragOver found it would go.
func (p *prereqPanel) drop(_ geom.Point, data any) {
	target, at := p.dropTarget, p.dropWhere
	p.dragExit()
	dd, ok := data.(*prereqDrag)
	if target == nil || !ok {
		return
	}
	path, _ := target.ClientData()[prereqDropKey].(string) //nolint:errcheck // dropAt found a string there.
	to, index := p.locate(path)
	if group, isList := p.node(path).(*gurps.PrereqList); isList && at == dropInto {
		to, index = group, len(group.Prereqs)
	} else if at == dropAfter {
		index++
	}
	list, i := p.locate(dd.path)
	node := list.Prereqs[i]
	p.restructure(i18n.Text("Move Prerequisite"), dd.path+keyMore, "", func() gurps.Prereq {
		if list != to {
			return p.relocate(list, i, to, index)
		}
		if moveEntry((*[]gurps.Prereq)(&to.Prereqs), i, index) {
			return node
		}
		return nil
	})
}

// drawDrop dims the prerequisite being dragged, and marks where it would go: a line before or after a row or group, or
// a tint over a group it would go into.
func (p *prereqPanel) drawDrop(gc *unison.Canvas, _ geom.Rect) {
	if dd, ok := panelDragData.(*prereqDrag); ok && dd.panel == p {
		if more := p.FindRefKey(dd.path + keyMore); more != nil {
			r := p.RectFromRoot(more.Parent().RectToRoot(more.Parent().ContentRect(true)))
			gc.DrawRect(r, faint(unison.ThemeSurface).Paint(gc, r, paintstyle.Fill))
		}
	}
	if p.dropTarget == nil {
		return
	}
	r := p.RectFromRoot(p.dropTarget.RectToRoot(p.dropTarget.ContentRect(true)))
	if p.dropWhere == dropInto {
		gc.DrawRoundedRect(r, geom.NewUniformSize(6), faint(unison.ThemeFocus).Paint(gc, r, paintstyle.Fill))
		return
	}
	y := r.Y
	if p.dropWhere == dropAfter {
		y = r.Bottom()
	}
	paint := unison.ThemeFocus.Paint(gc, r, paintstyle.Stroke)
	paint.SetStrokeWidth(2)
	gc.DrawLine(geom.NewPoint(r.X, y), geom.NewPoint(r.Right(), y), paint)
}

// moveTarget returns where Move up (dir -1) or Move down (dir 1) takes the node at the path: into an adjacent group,
// at its near end, past an adjacent sibling, or out of its list at either end. at is an index into to once the node is
// out of its list. ok is false at either end of the root list.
func (p *prereqPanel) moveTarget(path string, dir int) (to *gurps.PrereqList, at int, ok bool) {
	list, i := p.locate(path)
	if j := i + dir; j >= 0 && j < len(list.Prereqs) {
		if group, isList := list.Prereqs[j].(*gurps.PrereqList); isList {
			if dir < 0 {
				return group, len(group.Prereqs), true
			}
			return group, 0, true
		}
		return list, j, true
	}
	if parentPath := path[:strings.LastIndexByte(path, '.')]; parentPath != prereqRootPath {
		to, at = p.locate(parentPath)
		if dir > 0 {
			at++
		}
		return to, at, true
	}
	return nil, 0, false
}

// relocate takes the node at index i out of the list from and puts a copy of it, made for its new list, at index at
// of the list to, as that list stands once the node is out. It returns the copy.
func (p *prereqPanel) relocate(from *gurps.PrereqList, i int, to *gurps.PrereqList, at int) gurps.Prereq {
	node := from.Prereqs[i]
	from.Prereqs = slices.Delete(from.Prereqs, i, i+1)
	moved := p.clone(node, to)
	to.Prereqs = slices.Insert(to.Prereqs, at, moved)
	return moved
}

// typePopup adds the popup that switches a prerequisite to another of the permitted types. A type that isn't permitted,
// which only a file edited by hand can hold, is shown but not offered for others.
func (p *prereqPanel) typePopup(parent *unison.Panel, path string, pr gurps.Prereq) {
	current := pr.PrereqType()
	items := p.permittedChoices
	if !slices.Contains(items, current) {
		items = append(slices.Clone(items), current)
	}
	put(parent, compactPopup(p, path+":type", i18n.Text("Prerequisite Type"), items, current, prereqTypeName,
		func(t prereq.Type) {
			list, i := p.locate(path)
			if created := p.createPrereqForType(t, list); created != nil {
				if from, to := nameCriteria(pr), nameCriteria(created); from != nil && to != nil {
					*to = *from
				}
				list.Prereqs[i] = created
			}
		}))
}

func prereqTypeName(t prereq.Type) string {
	switch t {
	case prereq.Trait:
		return i18n.Text("trait")
	case prereq.Attribute:
		return i18n.Text("attribute")
	case prereq.ContainedQuantity:
		return i18n.Text("contained quantity")
	case prereq.ContainedWeight:
		return i18n.Text("contained weight")
	case prereq.EquippedEquipment:
		return i18n.Text("equipped equipment")
	case prereq.Skill:
		return i18n.Text("skill")
	case prereq.Spell:
		return i18n.Text("spell")
	case prereq.Script:
		return i18n.Text("script")
	default:
		return t.String()
	}
}

// createPrereqForType returns a new prerequisite of the type for the parent list, or nil for a type that can't be made.
func (p *prereqPanel) createPrereqForType(t prereq.Type, parent *gurps.PrereqList) gurps.Prereq {
	var one gurps.Prereq
	switch t {
	case prereq.Trait:
		one = gurps.NewTraitPrereq()
	case prereq.Attribute:
		one = gurps.NewAttributePrereq(p.entity)
	case prereq.ContainedQuantity:
		one = gurps.NewContainedQuantityPrereq()
	case prereq.ContainedWeight:
		one = gurps.NewContainedWeightPrereq(p.entity)
	case prereq.EquippedEquipment:
		one = gurps.NewEquippedEquipmentPrereq()
	case prereq.Skill:
		one = gurps.NewSkillPrereq()
	case prereq.Spell:
		sp := gurps.NewSpellPrereq()
		// Matching the owning spell's power source only makes sense for a prerequisite that belongs to a spell.
		sp.SamePowerSource = p.ownerIsSpell
		one = sp
	case prereq.Script:
		one = gurps.NewScriptPrereq()
	default:
		errs.Log(errs.New("unknown prerequisite type"), "type", t.Key())
		return nil
	}
	return one.Clone(parent)
}

// nameCriteria returns the name criteria of a prerequisite that has one, so it can survive a change of type.
func nameCriteria(pr gurps.Prereq) *criteria.Text {
	switch one := pr.(type) {
	case *gurps.TraitPrereq:
		return &one.NameCriteria
	case *gurps.SkillPrereq:
		return &one.NameCriteria
	case *gurps.EquippedEquipmentPrereq:
		return &one.NameCriteria
	default:
		return nil
	}
}

// hasPopup adds the popup that says whether the prerequisite is to be had or not.
func (p *prereqPanel) hasPopup(parent *unison.Panel, key string, has *bool, spell bool) {
	yes, no := i18n.Text("Has"), i18n.Text("Doesn't have")
	if spell {
		yes, no = i18n.Text("Knows"), i18n.Text("Doesn't know")
	}
	put(parent, compactPopup(p, key, i18n.Text("Has"), []bool{true, false}, *has, func(v bool) string {
		if v {
			return yes
		}
		return no
	}, func(v bool) { *has = v }))
}

// attributePopup adds a popup of attributes. A key that isn't one of them is shown as such, and kept until another is
// chosen.
func (p *prereqPanel) attributePopup(parent *unison.Panel, key, name string, value *string, flags gurps.AttributeFlags) {
	choices, current := gurps.AttributeChoices(p.entity, "", flags, *value)
	put(parent, compactPopup(p, key, name, choices, current,
		func(c *gurps.AttributeChoice) string { return c.Title }, func(c *gurps.AttributeChoice) { *value = c.Key }))
}

// powerSourceChip adds the optional power source criterion of a spell prerequisite. "Same as this spell's" is offered
// only to a spell's own prerequisites, or shown when a file edited by hand holds it elsewhere.
func (p *prereqPanel) powerSourceChip(chips *unison.Panel, path string, one *gurps.SpellPrereq) {
	const same = criteria.StringComparison(255)
	label := i18n.Text("power source")
	p.optional(chips, path, label, one.SamePowerSource || one.PowerSourceCriteria.Compare != criteria.AnyText,
		func() {
			one.SamePowerSource = p.ownerIsSpell
			if !p.ownerIsSpell {
				one.PowerSourceCriteria.Compare = criteria.IsText
			}
		},
		func() { one.SamePowerSource, one.PowerSourceCriteria = false, criteria.Text{} },
		func(chip *unison.Panel) {
			items := criteria.StringComparisons[1:]
			current := one.PowerSourceCriteria.Compare
			if p.ownerIsSpell || one.SamePowerSource {
				items = append([]criteria.StringComparison{same}, items...)
				if one.SamePowerSource {
					current = same
				}
			}
			title := i18n.Text("Power Source")
			comparison, _ := criteriaTitles(title)
			put(chip, compactPopup(p, path+":powercmp", comparison, items, current,
				func(c criteria.StringComparison) string {
					if c == same {
						return i18n.Text("is the same as this spell's")
					}
					return c.String()
				},
				func(c criteria.StringComparison) {
					one.SamePowerSource = c == same
					if one.SamePowerSource {
						c = criteria.AnyText
					}
					one.PowerSourceCriteria.Compare = c
				}))
			if !one.SamePowerSource {
				p.textField(chip, path+":power", title, "", &one.PowerSourceCriteria.Qualifier)
			}
		})
}

// textChip adds an optional text criterion, which is in use while its comparison isn't "is anything".
func (p *prereqPanel) textChip(chips *unison.Panel, path, label, subject string, c *criteria.Text) {
	p.optional(chips, path, label, c.Compare != criteria.AnyText,
		func() { *c = criteria.Text{Compare: criteria.IsText} },
		func() { *c = criteria.Text{} },
		func(chip *unison.Panel) { p.textCriteria(chip, path+":"+label, subject, "", c, false) })
}

// optional adds an optional criterion to chips: when on, as a chip holding the controls populate adds; otherwise as a
// dashed chip that adds it.
func (p *prereqPanel) optional(chips *unison.Panel, path, label string, on bool, add, remove func(), populate func(chip *unison.Panel)) {
	addKey := path + ":add " + label
	if on {
		p.chip(chips, path+":"+label, label, addKey, remove, populate)
		return
	}
	b := newDashedButton("+ "+label, func() {
		if after := p.edit(fmt.Sprintf(i18n.Text("Add %s"), label), addKey, add); after != nil {
			after.focus = path + ":" + label + keyChip
		}
		p.rebuild(path + ":" + label + keyChip)
	})
	b.RefKey = addKey
	put(chips, b)
}

// chip adds a rounded chip, whose reference key is key followed by keyChip, to the parent with the label, the controls
// populate adds, and a button that calls remove and then gives the focus to the widget with the reference key after.
func (p *prereqPanel) chip(parent *unison.Panel, key, label, after string, remove func(), populate func(chip *unison.Panel)) {
	chip := unison.NewPanel()
	chip.RefKey = key + keyChip
	chip.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: 2, Left: 10, Bottom: 2, Right: 3}))
	chip.DrawCallback = func(gc *unison.Canvas, _ geom.Rect) {
		r := chip.ContentRect(true)
		unison.DrawRoundedRectBase(gc, r, geom.NewUniformSize(r.Height/2), 1, unison.ThemeSurface,
			unison.ThemeSurfaceEdge)
	}
	text := unison.NewLabel()
	text.SetTitle(label)
	put(chip, text)
	populate(chip)
	title := fmt.Sprintf(i18n.Text("Remove %s"), label)
	x := newPrereqIconButton("", unison.CircledXSVG, title)
	x.ClickCallback = func() {
		if snapshot := p.edit(title, chip.RefKey, remove); snapshot != nil {
			snapshot.focus = after
		}
		p.rebuild(after)
	}
	put(chip, x)
	put(parent, hbox(chip, 5))
}

// textCriteria adds the comparison of a text criterion and, unless it is "is anything", the field for its qualifier,
// which shows the hint while empty. withAny offers "is anything", which a chip leaves to its remove button.
func (p *prereqPanel) textCriteria(parent *unison.Panel, key, subject, hint string, c *criteria.Text, withAny bool) {
	comparison, _ := criteriaTitles(subject)
	items := criteria.StringComparisons
	if !withAny {
		items = items[1:]
	}
	put(parent, compactPopup(p, key+"cmp", comparison, items, c.Compare, nil,
		func(v criteria.StringComparison) { c.Compare = v }))
	if c.Compare != criteria.AnyText {
		p.textField(parent, key, subject, hint, &c.Qualifier)
	}
}

// numberCriteria adds the comparison of a numeric criterion and, unless it is "anything", the field for its
// qualifier, which takes whole numbers when integer is true.
func (p *prereqPanel) numberCriteria(parent *unison.Panel, key, subject string, c *criteria.Number, withAny bool, minValue, maxValue fxp.Int, integer bool) {
	comparison, _ := criteriaTitles(subject)
	p.numberCompare(parent, key+"cmp", comparison, &c.Compare, withAny)
	if c.Compare == criteria.AnyNumber {
		return
	}
	set := func(v fxp.Int) { p.edit(subject, key, func() { c.Qualifier = v }) }
	if integer {
		p.addCompact(parent, NewIntegerField(p.targetMgr, key, subject, func() int { return c.Qualifier.AsInteger[int]() },
			func(v int) { set(fxp.FromInteger(v)) }, minValue.AsInteger[int](), maxValue.AsInteger[int](), false,
			false).Field)
		return
	}
	p.addCompact(parent, NewDecimalField(p.targetMgr, key, subject, func() fxp.Int { return c.Qualifier }, set, minValue,
		maxValue, false, false).Field)
}

// numberCompare adds the popup for a numeric comparison, in the short words a sentence uses.
func (p *prereqPanel) numberCompare(parent *unison.Panel, key, name string, c *criteria.NumericComparison, withAny bool) {
	items := criteria.NumericComparisons
	// "anything" is offered only where it is a choice, but is shown when a file edited by hand holds it elsewhere.
	if !withAny && *c != criteria.AnyNumber {
		items = items[1:]
	}
	put(parent, compactPopup(p, key, name, items, *c, func(v criteria.NumericComparison) string {
		if v == criteria.EqualsNumber {
			return i18n.Text("exactly")
		}
		return v.AltString()
	}, func(v criteria.NumericComparison) { *c = v }))
}

// textField adds a compact field for the text, which shows the hint while empty.
func (p *prereqPanel) textField(parent *unison.Panel, key, title, hint string, value *string) *StringField {
	field := NewStringField(p.targetMgr, key, title, func() string { return *value },
		func(s string) { p.edit(title, key, func() { *value = s }) })
	field.Watermark = hint
	field.SetMinimumTextWidthUsing("Weapon Master (Sword)")
	p.addCompact(parent, field.Field)
	return field
}

// addCompact adds the field to the parent with rounded borders, its changes recorded only by edit.
func (p *prereqPanel) addCompact(parent *unison.Panel, field *unison.Field) {
	put(parent, field)
	p.mute(field)
	unison.UninstallFocusBorders(field, field)
	unison.InstallFocusBorders(field, field, compactFieldBorder(true), compactFieldBorder(false))
	radius := geom.NewUniformSize(4)
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

// mute keeps the field from recording its own undo edits, since edit records them as snapshots of the tree.
func (p *prereqPanel) mute(field *unison.Field) {
	modified := field.ModifiedCallback
	field.ModifiedCallback = func(before, after *unison.FieldState) {
		p.muted = true
		defer func() { p.muted = false }()
		modified(before, after)
	}
}

func compactFieldBorder(focused bool) unison.Border {
	ink := unison.Ink(unison.ThemeSurfaceEdge)
	var w float32 = 1
	if focused {
		ink = unison.ThemeFocus
		w = 2
	}
	return unison.NewCompoundBorder(unison.NewLineBorder(ink, geom.NewUniformSize(4), geom.NewUniformInsets(w), false),
		unison.NewEmptyBorder(geom.Insets{Top: 3 - w, Left: 6 - w, Bottom: 3 - w, Right: 6 - w}))
}

// compactPopup returns a popup offering the items, as render shows them, with current selected. A choice is handed to
// set within an edit named name, after which the panel is rebuilt.
func compactPopup[T comparable](p *prereqPanel, key, name string, items []T, current T, render func(T) string, set func(T)) *unison.PopupMenu[T] {
	popup := unison.NewPopupMenu[T]()
	popup.RefKey = key
	popup.Accessibility.Name = name
	popup.ItemRendererCallback = render
	popup.HMargin = 6
	popup.AddItem(items...)
	// Sized to the choice showing rather than the widest, since any choice rebuilds the panel.
	popup.SetSizer(func(hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
		_, prefSize, _ = popup.DefaultSizes(hint)
		text := unison.NewText(popup.Text(), &unison.TextDecoration{Font: popup.Font})
		prefSize.Width = xmath.Ceil(text.Width() + popup.HMargin*2 + 4 + prefSize.Height*0.75)
		return prefSize, prefSize, prefSize
	})
	installPopupSelection(popup, current, func(v T) {
		p.edit(name, key, func() { set(v) })
		p.rebuild("")
	})
	return popup
}

// put adds the child to the parent, centered vertically on its line.
func put(parent *unison.Panel, child unison.Paneler) {
	if _, ok := parent.Layout().(*unison.FlowLayout); ok {
		child.AsPanel().SetLayoutData(align.Middle)
	} else {
		child.AsPanel().SetLayoutData(&unison.FlexLayoutData{VAlign: align.Middle})
	}
	parent.AddChild(child)
}

// hbox lays out the panel's children in a row that fills the width, returning the panel.
func hbox(panel *unison.Panel, spacing float32) *unison.Panel {
	panel.SetLayout(&unison.FlexLayout{Columns: len(panel.Children()), HSpacing: spacing})
	panel.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	return panel
}

func newPrereqColumn() *unison.Panel {
	column := unison.NewPanel()
	column.SetLayout(&unison.FlexLayout{Columns: 1, VSpacing: unison.StdVSpacing})
	column.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	return column
}

func newPrereqFlow() *unison.Panel {
	flow := newPrereqColumn()
	flow.SetLayout(&unison.FlowLayout{HSpacing: 6, VSpacing: 6})
	return flow
}

func newPrereqIconButton(key string, icon *unison.SVG, tooltip string) *unison.Button {
	b := unison.NewSVGButton(icon)
	b.RefKey = key
	b.Tooltip = newWrappedTooltip(tooltip)
	return b
}

// newDashedButton returns a button drawn as a dashed outline, which adds something.
func newDashedButton(title string, click func()) *unison.Button {
	b := unison.NewButton()
	b.HideBase = true
	b.OnBackgroundInk = unison.ThemeOnSurface
	b.CornerRadius = geom.NewUniformSize(100)
	b.SetTitle(title)
	b.ClickCallback = click
	draw := b.DrawCallback
	b.DrawCallback = func(gc *unison.Canvas, dirty geom.Rect) {
		draw(gc, dirty)
		r := b.ContentRect(true).Inset(geom.NewUniformInsets(0.5))
		paint := unison.ThemeSurfaceEdge.Paint(gc, r, paintstyle.Stroke)
		paint.SetPathEffect(unison.NewDashPathEffect([]float32{3, 3}, 0))
		gc.DrawRoundedRect(r, geom.NewUniformSize(min(b.CornerRadius.Width, r.Height/2)), paint)
	}
	return b
}
