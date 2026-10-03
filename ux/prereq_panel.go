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
	"slices"
	"strconv"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellcmp"
	"github.com/richardwilkes/gcs/v5/svg"
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
	// keyFirst marks the editor of an open row; focusing it focuses its first text field, or else its first control.
	keyFirst = ":first"
)

// prereqDropKey marks what a dragged prerequisite can be dropped on, holding a prereqDropSpot: the path of the node it
// shows and which part of the node it is, one of the dropOn values.
const prereqDropKey = "prereq.drop"

type prereqDropSpot struct {
	path string
	part int
}

const (
	dropOnRow = iota
	dropOnHead
	dropOnGroup
	dropOnEmpty
)

// Where a dragged prerequisite goes in relation to what it is dropped on.
const (
	dropBefore = iota
	dropAfter
	dropInto
)

const (
	pillCornerRadius  = 100
	maxPrereqQuantity = 9999
	defaultWhenTL     = 3
	// rowEndGap keeps what is drawn behind a row short of the panel's right edge, so it doesn't run into it.
	rowEndGap = 4
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
	summary          *sentenceButton
	views            []prereqView
	target           any
	open             string
	focus            string
	editKey          string
	editID           int64
	dropTarget       *unison.Panel
	dropWhere        int
	hash             uint64
	ownerIsSpell     bool
	rebuilding       bool
	// headed has an empty root show its head, once a group type or tech level has been chosen for it. An empty list
	// isn't saved, so these choices live here until it has prerequisites.
	headed bool
}

// prereqView shows a closed row's or a group's status, refreshed in place as the tree changes.
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
	}
	initTitledEditorSection(p, i18n.Text("Prerequisites"))
	p.targetMgr = NewTargetMgr(p)
	installPanelDragDrop(p.AsPanel(), prereqDragKey, p.dragOver, p.dragExit, p.drop)
	p.DrawOverCallback = p.drawDrop
	p.KeyDownCallback = p.keyDown
	p.build()
	// The item being edited, which evaluation leaves out, can only be found once the panel is in its editor.
	unison.InvokeTask(p.refresh)
	return p
}

// keyDown has Escape close the open row. Escape within the panel never reaches the editor, where it would discard the
// changes.
func (p *prereqPanel) keyDown(keyCode unison.KeyCode, mods mod.Modifiers, _ bool) bool {
	if keyCode != unison.KeyEscape || !noModifiersDown(mods) {
		return false
	}
	if p.open != "" {
		p.toggle(p.open)
	}
	return true
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

// prereqSnapshot is the panel as it stood before or after an edit: the tree, the open row, and the reference key of the
// widget that undo or redo returning to it gives the focus to.
type prereqSnapshot struct {
	tree   *gurps.PrereqList
	focus  string
	open   string
	headed bool
}

func cloneTree(list *gurps.PrereqList) *gurps.PrereqList {
	if list == nil {
		return nil
	}
	return list.CloneAsPrereqList(nil)
}

// edit applies change to the tree and records snapshots of the panel before and after it, under the title. Undo gives
// the focus to the widget with the reference key key, as redo does unless focusAfter is set, which also rebuilds the
// panel with the focus there. Consecutive edits with the same non-empty key, such as keystrokes in a field, are
// recorded as one until the panel is next rebuilt. It returns the snapshot after, or nil if nothing changed.
func (p *prereqPanel) edit(title, key, focusAfter string, change func()) *prereqSnapshot {
	before := &prereqSnapshot{tree: cloneTree(*p.root), focus: key, open: p.open, headed: p.headed}
	hash := gurps.Hash64(p.tree())
	*p.root = p.tree()
	change()
	if focusAfter != "" {
		defer p.rebuild(focusAfter)
	}
	if gurps.Hash64(*p.root) == hash && p.headed == before.headed {
		return nil
	}
	if key == "" || key != p.editKey {
		p.editID = unison.NextUndoID()
	}
	p.editKey = key
	after := &prereqSnapshot{tree: cloneTree(*p.root), focus: cmp.Or(focusAfter, key), open: p.open, headed: p.headed}
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

// install returns the panel to the snapshot, for undo and redo, editing a copy of its tree.
func (p *prereqPanel) install(snapshot *prereqSnapshot) {
	*p.root = cloneTree(snapshot.tree)
	p.placeholder = nil
	p.headed = snapshot.headed
	p.open = snapshot.open
	MarkModified(p)
	p.rebuild(snapshot.focus)
}

// restructure changes the shape of the tree through the widget with the reference key from, then rebuilds. change
// returns the node that is the result, whose more button takes the focus; with none, the widget with the reference key
// fallback does. The open row stays open wherever it ends up, and closes if it is removed.
func (p *prereqPanel) restructure(title, from, fallback string, change func() (dst gurps.Prereq)) {
	open := p.node(p.open)
	var dst gurps.Prereq
	after := p.edit(title, from, "", func() {
		dst = change()
		p.open = p.pathOf(open)
	})
	if dst != nil {
		fallback = p.pathOf(dst) + keyMore
	}
	if after != nil {
		after.focus = fallback
	}
	p.rebuild(fallback)
}

// rebuild replaces the panel's content once the current event has been handled, since that may have come from a widget
// about to be thrown away, keeping the scroll position and giving the focus to focus or else back where it was.
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

func (p *prereqPanel) build() {
	if node := p.node(p.open); node == nil || node.PrereqType() == prereq.List || node.PrereqType() == prereq.Unknown {
		p.open = ""
	}
	p.views = p.views[:0]
	p.summary = newSentenceButton("", nil)
	p.summary.SetBorder(unison.NewCompoundBorder(
		unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.Insets{Bottom: 1}, false),
		unison.NewEmptyBorder(geom.Insets{Top: 4, Left: 4, Bottom: 6, Right: 4}),
	))
	p.summary.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	// An empty root's placeholder says what the summary would.
	if len(p.tree().Prereqs) != 0 {
		p.AddChild(p.summary)
	}
	p.AddChild(p.group(p.tree(), prereqRootPath))
	p.refresh()
}

// Sync implements Syncer. Statuses are worked out again only once the tree has gone unchanged for
// scriptEvaluationDelay, since that runs its scripts, which would otherwise run on every keystroke.
func (p *prereqPanel) Sync() {
	hash := gurps.Hash64(p.tree())
	if hash == p.hash {
		return
	}
	unison.InvokeTaskAfter(func() {
		if hash == gurps.Hash64(p.tree()) && hash != p.hash {
			p.refresh()
		}
	}, scriptEvaluationDelay)
}

// refresh updates the summary, the sentences and the status icons from the tree, in place.
func (p *prereqPanel) refresh() {
	tree := p.tree()
	p.hash = gurps.Hash64(tree)
	summary := i18n.Text("No prerequisites.")
	if text := tree.Describe(p.entity, nil, emphasize); text != "" {
		summary = fmt.Sprintf(i18n.Text("%s."), text)
	}
	p.summary.setText(summary, "")
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
			v.sentence.setText(v.node.Describe(p.entity, nil, emphasize), suffix)
		}
		if list, ok := v.node.(*gurps.PrereqList); ok && v.group != nil {
			v.group.Accessibility.Name = groupName(list)
			if suffix != "" {
				v.group.Accessibility.Name += i18n.Text(", ") + suffix
			}
		}
	}
}

// status returns the node's status against the sheet, the tooltip of its icon and what a screen reader hears after its
// sentence.
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
	if group, isList := node.(*gurps.PrereqList); isList && len(group.Prereqs) == 0 {
		return checkSkipped, i18n.Text("Empty group, always met"), i18n.Text("empty group, always met")
	}
	var buffer xbytes.InsertBuffer
	gurps.SuppressScriptResolveErrorLogging(func() {
		if !node.Satisfied(p.entity, p.exclude(), &buffer, "\n- ", nil) {
			status = checkUnmet
		}
	})
	if status == checkMet {
		return checkMet, i18n.Text("Met"), i18n.Text("met")
	}
	// One unmet item reads as a sentence; more are a list.
	reason := buffer.String()
	if strings.Count(reason, "\n") == 1 {
		tip = fmt.Sprintf(i18n.Text("Not met: %s"), strings.TrimPrefix(reason, "\n- "))
	} else {
		tip = i18n.Text("Not met:") + reason
	}
	if script, isScript := node.(*gurps.ScriptPrereq); isScript {
		if result, message := p.evaluateScript(script); result == checkFailed {
			return checkFailed, tip, fmt.Sprintf(i18n.Text("couldn't run: %s"), message)
		}
	}
	return checkUnmet, tip, i18n.Text("not met")
}

// evaluateScript runs the script against the sheet, returning its status and the reason it gives, which is the error
// when it couldn't run.
func (p *prereqPanel) evaluateScript(script *gurps.ScriptPrereq) (status checkStatus, reason string) {
	var met, failed bool
	gurps.SuppressScriptResolveErrorLogging(func() { met, reason, failed = script.Evaluate(p.entity, p.exclude()) })
	switch {
	case failed:
		return checkFailed, reason
	case met:
		return checkMet, reason
	default:
		return checkUnmet, reason
	}
}

// statusIcon adds the icon for a node's status to the parent, which shows once there is a sheet, returning nil without
// one. Screen readers skip it, since the accessible name of the node's sentence says the status.
func (p *prereqPanel) statusIcon(parent *unison.Panel) *unison.Label {
	if p.entity == nil {
		return nil
	}
	icon := unison.NewLabel()
	icon.Accessibility.Role = role.None
	icon.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: 4}))
	addCentered(parent, icon)
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
	parentPath, last, found := strings.CutLast(path, ".")
	if !found {
		return nil, -1
	}
	index, err := strconv.Atoi(last)
	if list, ok := p.node(parentPath).(*gurps.PrereqList); ok && err == nil {
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
	if path != prereqRootPath {
		box.RefKey = path + ":group"
		box.ClientData()[prereqDropKey] = prereqDropSpot{path: path, part: dropOnGroup}
	}
	head := unison.NewPanel()
	// The same insets on the sides as a row's, so that the grips and the buttons on the right line up down the panel.
	insets := geom.Insets{Top: 2, Left: 4, Bottom: 2, Right: 8}
	if path == prereqRootPath && p.entity == nil {
		// Without a status icon before it, the root's pill keeps the icon's lead from the edge.
		insets.Left += 4
	}
	head.SetBorder(unison.NewEmptyBorder(insets))
	pill := compactPopup(p, path+keyPill, i18n.Text("Requirement"), []bool{true, false}, list.All, groupWord,
		func(all bool) { list.All = all })
	pill.HMargin = 10
	pill.CornerRadius = geom.NewUniformSize(pillCornerRadius)
	pill.BackgroundInk = color
	pill.OnBackgroundInk = color.DeriveOn()
	pill.EdgeInk = unison.Transparent
	desc := pill.Font.Descriptor()
	desc.Weight = weight.Bold
	pill.Font = desc.Font()
	head.ClientData()[prereqDropKey] = prereqDropSpot{path: path, part: dropOnHead}
	if path != prereqRootPath {
		grip := p.grip(head, path)
		putOnLine(grip.AsPanel(), controlHeight(head), grip.svg.Size.Height)
	}
	var empty *unison.Button
	var emptyRoot bool
	if len(list.Prereqs) == 0 {
		emptyRoot = path == prereqRootPath && !p.headed
		text := i18n.Text("Empty group. Add a requirement or drag one here.")
		if emptyRoot {
			text = i18n.Text("No prerequisites. Add one to get started.")
		}
		empty = newDashedButton(text, nil)
		// A click opens the menu where it lands, as a right-click does; a key opens it at the placeholder.
		empty.ClickCallback = func() { showMenu(empty.AsPanel(), p.addEntries(list, path)) }
		empty.ContextMenuCallback = func(geom.Point) unison.Menu { return newEntriesMenu(p.addEntries(list, path)) }
		empty.MouseUpCallback = func(where geom.Point, _ int, _ mod.Modifiers) bool {
			empty.Pressed = false
			empty.MarkForRedraw()
			if where.In(empty.ContentRect(false)) {
				if unison.IsAccessibilityActive() {
					empty.RequestFocus()
				}
				empty.ShowContextMenu(where)
			}
			return true
		}
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
		empty.RefKey = path + ":empty"
		empty.ClientData()[prereqDropKey] = prereqDropSpot{path: path, part: dropOnEmpty}
		empty.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	}
	// An untouched empty root has nothing for its pill, status or rail to speak of, so its placeholder takes their place.
	if emptyRoot {
		head.AddChild(empty)
	} else {
		p.views = append(p.views, prereqView{node: list, icon: p.statusIcon(head), group: box})
		addCentered(head, pill)
	}
	if list.WhenTL.Compare != criteria.AnyNumber {
		p.chip(head, path+":tl", i18n.Text("Remove Tech Level Condition"), path+keyAdd,
			func() { list.WhenTL = criteria.Number{} },
			func(chip *unison.Panel) {
				p.numberCriteria(chip, path+":tl", i18n.Text("Tech Level"), i18n.Text("When TL"), &list.WhenTL, 0,
					fxp.Twelve, true)
			})
	}
	add := newPrereqIconButton(path+keyAdd, unison.CircledAddSVG, i18n.Text("Add to this group"))
	add.ClickCallback = func() { showMenu(add.AsPanel(), p.addEntries(list, path)) }
	fitLine(add)
	add.SetLayoutData(&unison.FlexLayoutData{HAlign: align.End, VAlign: align.Middle, HGrab: !emptyRoot})
	head.AddChild(add)
	if path != prereqRootPath {
		p.moreButton(head, list, path)
	} else {
		// The root keeps room for the more button it doesn't have, so that its add button lines up with the others.
		room := newPrereqIconButton("", svg.CircledVerticalEllipsis, "")
		room.Hidden = true
		addCentered(head, room)
	}
	// As tall as the add button, since centering a shorter one could put its icon on a half pixel, blurring it.
	fitLine(head.Children()[len(head.Children())-1])
	box.AddChild(hbox(head, unison.StdHSpacing))

	rail := newPrereqColumn()
	// Lighter than the pill in dark mode, so that the rail has a contrast of at least 3:1 with the surface.
	rail.SetBorder(unison.NewCompoundBorder(unison.NewEmptyBorder(geom.Insets{Left: 14}),
		unison.NewLineBorder(color.DeriveLightness(0, 0.08), geom.Size{}, geom.Insets{Left: 3}, false),
		unison.NewEmptyBorder(geom.Insets{Left: 6, Bottom: 2})))
	for i, child := range list.Prereqs {
		if sub, ok := child.(*gurps.PrereqList); ok {
			rail.AddChild(p.group(sub, childPath(path, i)))
		} else {
			rail.AddChild(p.row(child, childPath(path, i)))
		}
	}
	if emptyRoot {
		return box
	}
	if empty != nil {
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
	row.ClientData()[prereqDropKey] = prereqDropSpot{path: path, part: dropOnRow}
	grip := p.grip(row, path)
	icon := p.statusIcon(row)
	var main *unison.Panel
	var line float32
	if open {
		if icon != nil {
			// An open row shows no status, but keeps the room for it, so that its editor lines up with the sentences.
			showCheckIcon(icon, checkMet)
			icon.OnBackgroundInk = unison.Transparent
		}
		row.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: 6, Left: 4, Bottom: 8, Right: 8}))
		main = p.editor(pr, path)
		row.AddChild(main)
		done := unison.NewButton()
		done.SetTitle(i18n.Text("Done"))
		done.RefKey = path + ":done"
		done.ClickCallback = func() { p.toggle(path) }
		row.AddChild(done)
		line = controlHeight(row)
	} else {
		row.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: 3, Left: 4, Bottom: 3, Right: 8}))
		var click func()
		if pr.PrereqType() != prereq.Unknown {
			click = func() { p.toggle(path) }
		}
		sentence := newSentenceButton(pr.Describe(p.entity, nil, emphasize), click)
		sentence.RefKey = path + keySentence
		p.dragBy(sentence.AsPanel(), row, path)
		if click == nil {
			// One this version of GCS doesn't understand can't be edited, so its sentence is static text.
			sentence.Tooltip = newWrappedTooltip(i18n.Text("This was most likely created by a newer version of GCS. Its original data will be written back out unchanged when this file is saved."))
		}
		main = sentence.AsPanel()
		row.AddChild(main)
		if script, ok := pr.(*gurps.ScriptPrereq); ok && script.ResolvedName(nil) == "" {
			describe := newDashedButton(i18n.Text("Add a description"), func() {
				p.open = path
				p.rebuild(path + ":name")
			})
			describe.OnBackgroundInk = unison.ThemeAlert
			addCentered(row, describe)
		}
		p.views = append(p.views, prereqView{node: pr, icon: icon, sentence: sentence})
		line = sentence.lineHeight()
	}
	p.moreButton(row, pr, path)
	hbox(row, unison.StdHSpacing)
	putOnLine(grip.AsPanel(), line, grip.svg.Size.Height)
	if icon != nil {
		putOnLine(icon.AsPanel(), line, checkIconSize())
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
		case child != grip.AsPanel() && (icon == nil || child != icon.AsPanel()):
			_, pref, _ := child.Sizes(geom.Size{})
			putOnLine(child, line, pref.Height)
		}
	}
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
	return row
}

// grip adds a drag handle to the row, which may be a group's head, and lets either be dragged to move the node at the
// path.
func (p *prereqPanel) grip(row *unison.Panel, path string) *DragHandle {
	handle := NewDragHandle(prereqDragKey, nil)
	row.AddChild(handle)
	p.dragBy(handle.AsPanel(), row, path)
	p.dragBy(row, row, path)
	return handle
}

// dragBy lets a press on target that moves far enough to be a drag move the node at the path, shown as an image of the
// row. Only moving counts, not holding the button down, so that a slow click is still a click. A press that becomes a
// drag is not also a click.
func (p *prereqPanel) dragBy(target, row *unison.Panel, path string) {
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
		startPanelDrag(row, prereqDragKey, &prereqDrag{panel: p, path: path}, row.FrameRect().Size, geom.Point{},
			func(gc *unison.Canvas, r geom.Rect) {
				gc.DrawRect(r, unison.ThemeBelowSurface.Paint(gc, r, paintstyle.Fill))
				row.Draw(gc, r)
			}, p.MarkForRedraw)
		return true
	}
	target.MouseUpCallback = func(where geom.Point, button int, mods mod.Modifiers) bool {
		return dragged || up == nil || up(where, button, mods)
	}
}

// toggle opens the row at the path, closing any other, or closes it if it is the open one.
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
	whose := i18n.Text("whose name")
	switch one := pr.(type) {
	case *gurps.TraitPrereq:
		p.hasPopup(fields, key("has"), &one.Has, false)
		p.typePopup(fields, path, pr)
		p.textCriteria(fields, key("name"), i18n.Text("Name"), i18n.Text("Trait name"), whose, whose, &one.NameCriteria,
			true)
		level := &one.LevelCriteria
		// Every trait has a level of at least 0, so that is the same as having no level criteria.
		p.levelChip(chips, path, level,
			level.Compare != criteria.AnyNumber && (level.Compare != criteria.AtLeastNumber || level.Qualifier > 0))
		p.textChip(chips, path, "notes", &one.NotesCriteria)
	case *gurps.SkillPrereq:
		p.hasPopup(fields, key("has"), &one.Has, false)
		p.typePopup(fields, path, pr)
		p.textCriteria(fields, key("name"), i18n.Text("Name"), i18n.Text("Skill name"), whose, whose, &one.NameCriteria,
			true)
		p.textChip(chips, path, "specialization", &one.SpecializationCriteria)
		p.textChip(chips, path, "optspecialization", &one.OptionalSpecializationCriteria)
		// Unlike a trait's, "at least 0" is a real criterion here, leaving out skills with no usable level.
		p.levelChip(chips, path, &one.LevelCriteria, one.LevelCriteria.Compare != criteria.AnyNumber)
	case *gurps.SpellPrereq:
		p.hasPopup(fields, key("has"), &one.Has, true)
		quantity := func() {
			p.numberCriteria(fields, key("quantity"), i18n.Text("Quantity"), "", &one.QuantityCriteria, 0,
				fxp.FromInteger(maxPrereqQuantity), true)
		}
		// A count of colleges follows the match, as in "spells from at least 2 colleges".
		colleges := one.SubType == spellcmp.CollegeCount
		if !colleges {
			quantity()
		}
		p.typePopup(fields, path, pr)
		addCentered(fields, compactPopup(p, key("match"), i18n.Text("Spell Match"), spellcmp.Types, one.SubType,
			func(t spellcmp.Type) string {
				if t == spellcmp.CollegeCount {
					return i18n.Text("from")
				}
				return t.String()
			}, func(t spellcmp.Type) { one.SubType = t }))
		if colleges {
			quantity()
			addJoiningWords(fields, i18n.Text("college(s)"))
		}
		if one.SubType.UsesStringCriteria() {
			p.textCriteria(fields, key("qualifier"), i18n.Text("Spell"), "", "", "", &one.QualifierCriteria, true)
		}
		p.powerSourceChip(chips, path, one)
	case *gurps.AttributePrereq:
		p.hasPopup(fields, key("has"), &one.Has, false)
		p.typePopup(fields, path, pr)
		flags := gurps.SizeFlag | gurps.DodgeFlag | gurps.ParryFlag | gurps.BlockFlag
		p.attributePopup(fields, key("which"), i18n.Text("Attribute"), "", &one.Which, flags)
		// Named apart from the attribute popup before it.
		p.numberCriteria(fields, key("value"), i18n.Text("Value"), i18n.Text("which"), &one.QualifierCriteria, fxp.Min,
			fxp.Max, false)
		p.optional(chips, path, "combined", one.CombinedWith != "",
			func() { one.CombinedWith = gurps.AttributeIDFor(p.entity, gurps.DexterityID) },
			func() { one.CombinedWith = "" },
			func(chip *unison.Panel) {
				p.attributePopup(chip, key("combined"), i18n.Text("Combined With"), prereqCriteria("combined").prefix,
					&one.CombinedWith, flags)
			})
	case *gurps.EquippedEquipmentPrereq:
		p.typePopup(fields, path, pr)
		p.textCriteria(fields, key("name"), i18n.Text("Name"), i18n.Text("Item name"), whose, whose, &one.NameCriteria,
			true)
		p.textChip(chips, path, "tag", &one.TagsCriteria)
	case *gurps.ContainedQuantityPrereq:
		p.hasPopup(fields, key("has"), &one.Has, false)
		p.typePopup(fields, path, pr)
		p.numberCriteria(fields, key("quantity"), i18n.Text("Quantity"), "", &one.QualifierCriteria, 0,
			fxp.FromInteger(maxPrereqQuantity), true)
	case *gurps.ContainedWeightPrereq:
		p.hasPopup(fields, key("has"), &one.Has, false)
		p.typePopup(fields, path, pr)
		title := i18n.Text("Weight")
		comparison, _ := criteriaTitles(title)
		p.numberCompare(fields, key("weightcmp"), comparison, i18n.Text("which"), &one.WeightCriteria.Compare)
		p.addCompact(fields, NewWeightField(p.targetMgr, key("weight"), title, p.entity,
			func() fxp.Weight { return one.WeightCriteria.Qualifier },
			func(w fxp.Weight) { p.edit(title, key("weight"), "", func() { one.WeightCriteria.Qualifier = w }) },
			0, fxp.Weight(fxp.Max), false).withoutUndo())
	case *gurps.ScriptPrereq:
		p.typePopup(fields, path, pr)
		addJoiningWords(fields, i18n.Text("described as"))
		e := newScriptEditor(func() string { return one.Script },
			func(script string) { p.edit(i18n.Text("Script"), key("script"), "", func() { one.Script = script }) },
			p.scriptOptions(one))
		title := i18n.Text("Description")
		name := NewStringField(p.targetMgr, key("name"), title, func() string { return one.Name },
			func(s string) {
				p.edit(title, key("name"), "", func() { one.Name = s })
				e.refresh()
			})
		name.Watermark = i18n.Text(`Describe this requirement, like "DX + Per totals at least 26"`)
		name.SetMinimumTextWidthUsing(name.Watermark)
		p.addCompact(fields, name.withoutUndo())
		e.field.RefKey = key("script")
		e.field.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
		e.field.withoutUndo()
		box.AddChild(e)
	default:
	}
	if len(chips.Children()) != 0 {
		// The buttons that add unused criteria go after the chips in use, so an added one takes its place among them.
		for _, child := range slices.Clone(chips.Children()) {
			if _, ok := child.Self.(*unison.Button); ok {
				chips.AddChild(child)
			}
		}
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

// scriptOptions returns the script editor's options for a script prerequisite.
func (p *prereqPanel) scriptOptions(pr *gurps.ScriptPrereq) *scriptEditorOptions {
	opts := &scriptEditorOptions{
		Title:  i18n.Text("Script"),
		Hint:   i18n.Text("Tab indents. Esc closes."),
		Footer: i18n.Text("The script's last value decides: true or empty text means met; false means not met; any other text means not met, with that text as the reason."),
	}
	if p.entity != nil {
		opts.Evaluate = func(script string) (checkStatus, string) {
			one := *pr
			one.Script = script
			status, reason := p.evaluateScript(&one)
			reason = strings.TrimSpace(reason)
			switch {
			case status == checkMet:
				return status, i18n.Text("Passed")
			case status == checkFailed:
				return status, fmt.Sprintf(i18n.Text("Couldn't run: %s"), reason)
			case reason == "":
				return status, i18n.Text("Failed")
			default:
				return status, fmt.Sprintf(i18n.Text("Failed: %s"), reason)
			}
		}
	}
	return opts
}

// addEntries returns the entries of a list's Add menu.
func (p *prereqPanel) addEntries(list *gurps.PrereqList, path string) []menuEntry {
	add := func(title string, created gurps.Prereq) {
		at := childPath(path, len(list.Prereqs))
		opens := created.PrereqType() != prereq.List
		focus := at + keyMore
		if opens {
			focus = at + keyFirst
		}
		p.edit(title, path+keyAdd, focus, func() {
			list.Prereqs = append(list.Prereqs, created)
			if opens {
				p.open = at
			}
		})
	}
	entries := []menuEntry{{Label: i18n.Text("Requirement")}}
	for _, t := range p.permittedChoices {
		entries = append(entries, menuEntry{Label: t.AltString(), Act: func() {
			add(i18n.Text("Add Prerequisite"), p.createPrereqForType(t, list))
		}})
	}
	entries = append(entries, menuEntry{Label: i18n.Text("Structure")})
	for i, label := range []string{i18n.Text("All of Group"), i18n.Text("Any of Group")} {
		entries = append(entries, menuEntry{Label: label, Act: func() {
			if path != prereqRootPath || len(list.Prereqs) != 0 {
				add(i18n.Text("Add Group"), &gurps.PrereqList{Type: prereq.List, Parent: list, All: i == 0})
				return
			}
			// An empty root takes the group type itself, rather than holding a group of that type.
			p.edit(label, path+keyPill, path+keyPill, func() { list.All, p.headed = i == 0, true })
		}})
	}
	if list.WhenTL.Compare == criteria.AnyNumber {
		entries = append(entries, menuEntry{Label: i18n.Text("Only When TL…"), Act: func() {
			p.edit(i18n.Text("Add Tech Level Condition"), path+keyAdd, path+":tl"+keyChip, func() {
				list.WhenTL = criteria.Number{Compare: criteria.AtMostNumber, Qualifier: fxp.FromInteger(defaultWhenTL)}
				p.headed = p.headed || path == prereqRootPath
			})
		}})
	}
	return entries
}

// moreButton adds the button for the node's more menu.
func (p *prereqPanel) moreButton(parent *unison.Panel, node gurps.Prereq, path string) {
	b := newPrereqIconButton(path+keyMore, svg.CircledVerticalEllipsis, i18n.Text("More actions"))
	b.ClickCallback = func() { showMenu(b.AsPanel(), p.moreEntries(node, path)) }
	addCentered(parent, b)
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
			p.restructure(title, from, "", func() gurps.Prereq { return relocate(list, i, to, at) })
		}})
	}
	entries = append(entries, menuEntry{Label: i18n.Text("Wrap in Group"), Act: func() {
		p.restructure(i18n.Text("Wrap in Group"), from, "", func() gurps.Prereq {
			group := &gurps.PrereqList{Type: prereq.List, Parent: list, All: !list.All, Prereqs: gurps.Prereqs{node}}
			node.SetParentList(group)
			list.Prereqs[i] = group
			return node
		})
	}})
	if g, ok := node.(*gurps.PrereqList); ok && len(g.Prereqs) != 0 && g.WhenTL.Compare == criteria.AnyNumber &&
		(g.All == list.All || len(g.Prereqs) == 1) {
		entries = append(entries, menuEntry{Label: i18n.Text("Ungroup"), Act: func() {
			p.restructure(i18n.Text("Ungroup"), from, "", func() gurps.Prereq {
				for _, child := range g.Prereqs {
					child.SetParentList(list)
				}
				list.Prereqs = slices.Replace(list.Prereqs, i, i+1, g.Prereqs...)
				return g.Prereqs[0]
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

// dragOver keeps the pointer in view and marks where the prerequisite would be dropped (see dropAt).
func (p *prereqPanel) dragOver(where geom.Point, data any) bool {
	p.ScrollRectIntoView(geom.NewRect(where.X, where.Y-16, 1, 1))
	p.ScrollRectIntoView(geom.NewRect(where.X, where.Y+16, 1, 1))
	target, _, at := p.dropAt(where, data)
	if target != p.dropTarget || at != p.dropWhere {
		p.dropTarget = target
		p.dropWhere = at
		p.MarkForRedraw()
	}
	return true
}

// dropAt returns the panel a prerequisite dragged to where would be dropped on, the path of its node and where it would
// go: before or after a row, before a group over the top of its head and into it below that, after a group beside or
// below its last child, or into an empty group. The panel is nil where nothing can go, such as into itself.
func (p *prereqPanel) dropAt(where geom.Point, data any) (target *unison.Panel, path string, at int) {
	dd, ok := data.(*prereqDrag)
	if !ok || dd.panel != p || p.node(dd.path) == nil {
		return nil, "", 0
	}
	for target = p.PanelAt(where); target != nil && target != p.AsPanel(); target = target.Parent() {
		spot, isTarget := target.ClientData()[prereqDropKey].(prereqDropSpot)
		if !isTarget {
			continue
		}
		if spot.path == dd.path || strings.HasPrefix(spot.path, dd.path+".") {
			return nil, "", 0
		}
		y := target.PointFromRoot(p.PointToRoot(where)).Y
		height := target.FrameRect().Height
		switch spot.part {
		case dropOnEmpty:
			return target, spot.path, dropInto
		case dropOnGroup:
			// Only the rail's margins reach the group itself; beside or below its last child is after it.
			rows := target.Children()[len(target.Children())-1].Children()
			last := rows[len(rows)-1]
			if where.Y > p.RectFromRoot(last.RectToRoot(last.ContentRect(true))).CenterY() {
				return target, spot.path, dropAfter
			}
			return nil, "", 0
		case dropOnHead:
			if spot.path != prereqRootPath && y < height*0.3 {
				return target, spot.path, dropBefore
			}
			return target, spot.path, dropInto
		default:
			if y < height/2 {
				return target, spot.path, dropBefore
			}
			return target, spot.path, dropAfter
		}
	}
	return nil, "", 0
}

func (p *prereqPanel) dragExit() {
	p.dropTarget = nil
	p.MarkForRedraw()
}

// drop moves the dragged prerequisite to where it would go.
func (p *prereqPanel) drop(where geom.Point, data any) {
	target, path, at := p.dropAt(where, data)
	p.dragExit()
	dd, ok := data.(*prereqDrag)
	if target == nil || !ok {
		return
	}
	to, index := p.locate(path)
	if group, isList := p.node(path).(*gurps.PrereqList); isList && at == dropInto {
		to, index = group, len(group.Prereqs)
	} else if at == dropAfter {
		index++
	}
	from, i := p.locate(dd.path)
	if from == to && i < index {
		// Taking the node out moves what comes after it up by one.
		index--
	}
	p.restructure(i18n.Text("Move Prerequisite"), dd.path+keyMore, "", func() gurps.Prereq {
		if from == to && i == index {
			return nil
		}
		return relocate(from, i, to, index)
	})
}

// drawDrop dims the prerequisite being dragged, and marks where it would go, in the ink of GCS's other drop markers: a
// line before or after a row or group, or a tint over a group it would go into.
func (p *prereqPanel) drawDrop(gc *unison.Canvas, _ geom.Rect) {
	if dd, ok := panelDragData.(*prereqDrag); ok && dd.panel == p {
		if more := p.FindRefKey(dd.path + keyMore); more != nil {
			r := p.RectFromRoot(more.Parent().RectToRoot(more.Parent().ContentRect(true)))
			gc.DrawRect(r, faintInk(unison.ThemeSurface).Paint(gc, r, paintstyle.Fill))
		}
	}
	if p.dropTarget == nil {
		return
	}
	r := p.RectFromRoot(p.dropTarget.RectToRoot(p.dropTarget.ContentRect(true)))
	r.Width -= rowEndGap
	if p.dropWhere == dropInto {
		gc.DrawRoundedRect(r, geom.NewUniformSize(6), faintInk(unison.ThemeWarning).Paint(gc, r, paintstyle.Fill))
		return
	}
	y := r.Y
	if p.dropWhere == dropAfter {
		y = r.Bottom()
	}
	paint := unison.ThemeWarning.Paint(gc, r, paintstyle.Stroke)
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

// relocate moves the node at index i of the list from to index at of the list to, as that list stands once the node is
// out, returning the node.
func relocate(from *gurps.PrereqList, i int, to *gurps.PrereqList, at int) gurps.Prereq {
	node := from.Prereqs[i]
	from.Prereqs = slices.Delete(from.Prereqs, i, i+1)
	node.SetParentList(to)
	to.Prereqs = slices.Insert(to.Prereqs, at, node)
	return node
}

// typePopup adds the popup that switches a prerequisite to another of the permitted types. A type that isn't permitted,
// which only a file edited by hand can hold, is shown but not offered for others.
func (p *prereqPanel) typePopup(parent *unison.Panel, path string, pr gurps.Prereq) {
	current := pr.PrereqType()
	items := p.permittedChoices
	if !slices.Contains(items, current) {
		items = append(slices.Clone(items), current)
	}
	var render func(prereq.Type) string
	if current == prereq.EquippedEquipment || current == prereq.Script {
		// With no "has" popup ahead of it, it starts the sentence.
		render = func(t prereq.Type) string { return xstrings.FirstToUpper(t.String()) }
	}
	addCentered(parent, compactPopup(p, path+":type", i18n.Text("Prerequisite Type"), items, current, render,
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

// createPrereqForType returns a new prerequisite of the type for the parent list, or nil for a type that can't be made.
func (p *prereqPanel) createPrereqForType(t prereq.Type, parent *gurps.PrereqList) gurps.Prereq {
	var one gurps.Prereq
	switch t {
	case prereq.Trait:
		trait := gurps.NewTraitPrereq()
		// New ones start without the level criterion, which the editor offers as a chip.
		trait.LevelCriteria = criteria.Number{}
		one = trait
	case prereq.Attribute:
		one = gurps.NewAttributePrereq(p.entity)
	case prereq.ContainedQuantity:
		one = gurps.NewContainedQuantityPrereq()
	case prereq.ContainedWeight:
		one = gurps.NewContainedWeightPrereq(p.entity)
	case prereq.EquippedEquipment:
		one = gurps.NewEquippedEquipmentPrereq()
	case prereq.Skill:
		skill := gurps.NewSkillPrereq()
		skill.LevelCriteria = criteria.Number{}
		one = skill
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
	addCentered(parent, compactPopup(p, key, i18n.Text("Has"), []bool{true, false}, *has, func(v bool) string {
		if v {
			return yes
		}
		return no
	}, func(v bool) { *has = v }))
}

// attributePopup adds a popup of attributes, each after the prefix. A key that isn't one of them is shown as such, and
// kept until another is chosen.
func (p *prereqPanel) attributePopup(parent *unison.Panel, key, name, prefix string, value *string, flags gurps.AttributeFlags) {
	choices, current := gurps.AttributeChoices(p.entity, prefix, flags, *value)
	addCentered(parent, compactPopup(p, key, name, choices, current,
		func(c *gurps.AttributeChoice) string { return c.Title }, func(c *gurps.AttributeChoice) { *value = c.Key }))
}

// powerSourceChip adds the optional power source criterion of a spell prerequisite. "Same as this spell's" is offered
// only to a spell's own prerequisites, or shown when a file edited by hand holds it elsewhere.
func (p *prereqPanel) powerSourceChip(chips *unison.Panel, path string, one *gurps.SpellPrereq) {
	p.optional(chips, path, "power", one.SamePowerSource || one.PowerSourceCriteria.Compare != criteria.AnyText,
		func() {
			one.SamePowerSource = p.ownerIsSpell
			if !p.ownerIsSpell {
				one.PowerSourceCriteria.Compare = criteria.IsText
			}
		},
		func() { one.SamePowerSource, one.PowerSourceCriteria = false, criteria.Text{} },
		func(chip *unison.Panel) {
			prefix := prereqCriteria("power").prefix
			choices := criteria.PrefixedStringComparisonChoices(prefix, prefix)
			same := i18n.Text("and whose power source is the same as this spell's")
			var items []string
			if p.ownerIsSpell || one.SamePowerSource {
				items = append(items, same)
			}
			items = append(items, choices[1:]...)
			current := choices[one.PowerSourceCriteria.Compare.EnsureValid()]
			if one.SamePowerSource {
				current = same
			}
			title := i18n.Text("Power Source")
			comparison, _ := criteriaTitles(title)
			addCentered(chip, compactPopup(p, path+":powercmp", comparison, items, current, nil, func(choice string) {
				one.SamePowerSource = choice == same
				one.PowerSourceCriteria.Compare = criteria.StringComparisons[max(slices.Index(choices, choice), 0)]
			}))
			if !one.SamePowerSource {
				p.textField(chip, path+":power", title, "", &one.PowerSourceCriteria.Qualifier)
			}
		})
}

// prereqCriterion describes an optional criterion of a prerequisite: the subject its controls are named for, the
// button that adds it, and the words its chip's first popup starts each choice with or, when notPrefix is set, each
// "not" choice (see criteria.PrefixedStringComparisonChoices). addTitle and removeTitle, when set, replace the titles
// of adding and removing it that are made from the subject.
type prereqCriterion struct {
	subject, add, prefix, notPrefix, addTitle, removeTitle string
}

// prereqCriteria returns the optional criterion with the key its widgets' reference keys are made from.
func prereqCriteria(key string) prereqCriterion {
	switch key {
	case "notes":
		return prereqCriterion{subject: i18n.Text("Notes"), add: i18n.Text("+ notes"), prefix: i18n.Text("and whose notes")}
	case "tag":
		return prereqCriterion{
			subject: i18n.Text("Tag"), add: i18n.Text("+ tag"),
			prefix: i18n.Text("and at least one tag"), notPrefix: i18n.Text("and all tags"),
		}
	case "specialization":
		return prereqCriterion{
			subject: i18n.Text("Specialization"), add: i18n.Text("+ specialization"),
			prefix: i18n.Text("and whose specialization"),
		}
	case "optspecialization":
		return prereqCriterion{
			subject: i18n.Text("Optional Specialization"), add: i18n.Text("+ optional specialization"),
			prefix: i18n.Text("and whose optional specialization"),
		}
	case "level":
		return prereqCriterion{subject: i18n.Text("Level"), add: i18n.Text("+ level"), prefix: i18n.Text("and whose level")}
	case "power":
		return prereqCriterion{
			subject: i18n.Text("Power Source"), add: i18n.Text("+ power source"),
			prefix: i18n.Text("and whose power source"),
		}
	case "combined":
		return prereqCriterion{
			subject: i18n.Text("Combined With"), add: i18n.Text("+ combined with"),
			prefix: i18n.Text("combined with"), addTitle: i18n.Text("Add Combined Attribute"),
			removeTitle: i18n.Text("Remove Combined Attribute"),
		}
	default:
		return prereqCriterion{}
	}
}

// titles returns the titles of adding and removing the criterion.
func (c *prereqCriterion) titles() (add, remove string) {
	return cmp.Or(c.addTitle, fmt.Sprintf(i18n.Text("Add %s"), c.subject)),
		cmp.Or(c.removeTitle, fmt.Sprintf(i18n.Text("Remove %s"), c.subject))
}

// textChip adds the optional text criterion with the key, which is in use while its comparison isn't "is anything".
func (p *prereqPanel) textChip(chips *unison.Panel, path, key string, c *criteria.Text) {
	p.optional(chips, path, key, c.Compare != criteria.AnyText,
		func() { *c = criteria.Text{Compare: criteria.IsText} },
		func() { *c = criteria.Text{} },
		func(chip *unison.Panel) {
			one := prereqCriteria(key)
			p.textCriteria(chip, path+":"+key, one.subject, "", one.prefix, cmp.Or(one.notPrefix, one.prefix), c, false)
		})
}

// levelChip adds an optional level criterion.
func (p *prereqPanel) levelChip(chips *unison.Panel, path string, level *criteria.Number, on bool) {
	p.optional(chips, path, "level", on,
		func() { *level = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.One} },
		func() { *level = criteria.Number{} },
		func(chip *unison.Panel) {
			p.numberCriteria(chip, path+":level", i18n.Text("Level"), prereqCriteria("level").prefix, level, 0,
				fxp.Thousand, false)
		})
}

// optional adds the optional criterion with the key to chips: as a chip of the controls populate adds while on, else as
// a dashed button that adds it.
func (p *prereqPanel) optional(chips *unison.Panel, path, key string, on bool, add, remove func(), populate func(chip *unison.Panel)) {
	c := prereqCriteria(key)
	addTitle, removeTitle := c.titles()
	addKey := path + ":add " + key
	if on {
		p.chip(chips, path+":"+key, removeTitle, addKey, remove, populate)
		return
	}
	b := newDashedButton(c.add, func() { p.edit(addTitle, addKey, path+":"+key+keyChip, add) })
	b.RefKey = addKey
	fitLine(b)
	addCentered(chips, b)
}

// prereqChip is the panel of an optional criterion in use, which holds its controls.
type prereqChip struct {
	unison.Panel
}

// chip adds a rounded chip keyed key+keyChip to the parent, holding the controls populate adds and a button titled
// title that removes the criterion, giving the focus to the widget with the reference key after.
func (p *prereqPanel) chip(parent *unison.Panel, key, title, after string, remove func(), populate func(chip *unison.Panel)) {
	chip := &prereqChip{}
	chip.Self = chip
	chip.RefKey = key + keyChip
	chip.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: chipPadding, Left: 10, Bottom: chipPadding, Right: 3}))
	chip.DrawCallback = func(gc *unison.Canvas, _ geom.Rect) {
		r := chip.ContentRect(true)
		unison.DrawRoundedRectBase(gc, r, geom.NewUniformSize(r.Height/2), 1, unison.ThemeSurface,
			unison.ThemeSurfaceEdge)
	}
	populate(chip.AsPanel())
	x := newPrereqIconButton("", unison.CircledXSVG, title)
	x.ClickCallback = func() { p.edit(title, chip.RefKey, after, remove) }
	addCentered(chip.AsPanel(), x)
	addCentered(parent, hbox(chip.AsPanel(), 5))
}

// textCriteria adds the comparison of a text criterion, each choice after the prefix or, for a "not" comparison, the
// notPrefix, and, unless it is "is anything", the field for its qualifier, which shows the hint while empty. withAny
// offers "is anything", which a chip leaves to its remove button.
func (p *prereqPanel) textCriteria(parent *unison.Panel, key, subject, hint, prefix, notPrefix string, c *criteria.Text, withAny bool) {
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
func (p *prereqPanel) numberCriteria(parent *unison.Panel, key, subject, prefix string, c *criteria.Number, minValue, maxValue fxp.Int, integer bool) {
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
func (p *prereqPanel) numberCompare(parent *unison.Panel, key, name, prefix string, c *criteria.NumericComparison) {
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
func (p *prereqPanel) textField(parent *unison.Panel, key, title, hint string, value *string) *StringField {
	field := NewStringField(p.targetMgr, key, title, func() string { return *value },
		func(s string) { p.edit(title, key, "", func() { *value = s }) })
	field.Watermark = hint
	field.SetMinimumTextWidthUsing(i18n.Text("Weapon Master (Sword)"))
	p.addCompact(parent, field.withoutUndo())
	return field
}

// addCompact adds the field, which leaves undo to edit, to the parent with rounded borders.
func (p *prereqPanel) addCompact(parent *unison.Panel, field *unison.Field) {
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

// chipPadding is the room a chip leaves above and below the controls it holds.
const chipPadding = 2

// controlHeight returns the height of a control in the parent: within a chip, that of a standard button; anywhere else,
// that of a chip holding one, so that everything on a line, chips and group pills included, is as tall.
func controlHeight(parent *unison.Panel) float32 {
	theme := &unison.DefaultButtonTheme
	height := xmath.Ceil(theme.Font.LineHeight()) + 2*(theme.VMargin+1)
	if parent != nil {
		if _, inChip := parent.Self.(*prereqChip); inChip {
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
		prefSize.Height = controlHeight(popup.Parent())
		text := unison.NewText(popup.Text(), &unison.TextDecoration{Font: popup.Font})
		prefSize.Width = xmath.Ceil(text.Width() + popup.HMargin*2 + 4 + prefSize.Height*0.75)
		return prefSize, prefSize, prefSize
	})
	installPopupSelection(popup, current, func(v T) {
		p.edit(name, key, "", func() { set(v) })
		p.rebuild("")
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
	if _, ok := parent.Layout().(*unison.FlowLayout); ok {
		child.AsPanel().SetLayoutData(align.Middle)
	} else {
		child.AsPanel().SetLayoutData(&unison.FlexLayoutData{VAlign: align.Middle})
	}
	parent.AddChild(child)
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
