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
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/drag"
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

// checkStatus is the outcome of checking a requirement, such as a script, against a sheet.
type checkStatus uint8

// Possible checkStatus values.
const (
	checkMet checkStatus = iota
	checkUnmet
	checkFailed
	checkSkipped
)

// showCheckIcon has the label show the icon of the status.
func showCheckIcon(label *unison.Label, status checkStatus) {
	icon, ink := unison.CheckmarkSVG, unison.Ink(colors.Success)
	switch status {
	case checkUnmet:
		icon, ink = svg.Not, colors.Failure
	case checkFailed:
		icon, ink = unison.TriangleExclamationSVG, unison.ThemeWarning
	case checkSkipped:
		icon, ink = unison.DashSVG, faint(unison.ThemeOnSurface)
	default:
	}
	if label.Drawable == nil {
		// Without an icon the label took no room, so the panels around it must be laid out again.
		label.MarkForLayoutRecursivelyUpward()
	}
	size := unison.DefaultLabelTheme.Font.Baseline()
	label.Drawable = &unison.DrawableSVG{SVG: icon, Size: geom.NewSize(size, size).Ceil()}
	label.OnBackgroundInk = ink
	label.MarkForRedraw()
}

// faint returns the ink at 30% opacity.
func faint(ink unison.Ink) unison.Ink {
	return &unison.ColorFilteredInk{OriginalInk: ink, ColorFilter: unison.Alpha30Filter()}
}

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
	// A change made in a row's editor is shown by opening that row.
	if path, name, ok := strings.Cut(snapshot.focus, ":"); ok && ":"+name != keyMore && ":"+name != keySentence {
		if node := p.node(path); node != nil && node.PrereqType() != prereq.List {
			p.open = path
		}
	}
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
		first := target.FirstFocusableChild()
		// A row's editor gives the focus to its first text field, where typing goes.
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
	if text := tree.Describe(p.entity, nil, em); text != "" {
		summary = fmt.Sprintf(i18n.Text("%s."), text)
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
			v.sentence.setText(v.node.Describe(p.entity, nil, em), suffix)
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
// sentence. Everything within a list that doesn't apply at the sheet's tech level is skipped, as is an empty group, which
// is always met.
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
	var met, failed bool
	var reason string
	if script, isScript := node.(*gurps.ScriptPrereq); isScript {
		gurps.SuppressScriptResolveErrorLogging(func() { met, reason, failed = script.Evaluate(p.entity, p.exclude()) })
	} else {
		var buffer xbytes.InsertBuffer
		met = node.Satisfied(p.entity, p.exclude(), &buffer, "\n- ", nil)
		// One unmet item reads as a sentence; more are a list.
		if reason = buffer.String(); strings.Count(reason, "\n") == 1 {
			reason = strings.TrimPrefix(reason, "\n- ")
		}
	}
	switch {
	case failed:
		return checkFailed, reason, fmt.Sprintf(i18n.Text("script error: %s"), reason)
	case met:
		return checkMet, i18n.Text("Met"), i18n.Text("met")
	case strings.HasPrefix(reason, "\n"):
		return checkUnmet, i18n.Text("Not met:") + reason, i18n.Text("not met")
	default:
		return checkUnmet, fmt.Sprintf(i18n.Text("Not met: %s"), reason), i18n.Text("not met")
	}
}

// statusIcon adds the icon for a node's status to the parent, which shows once there is a sheet. Without one its room
// is kept, so that everything lines up as it would with one. Screen readers skip it, since the accessible name of the
// node's sentence says the status.
func (p *prereqPanel) statusIcon(parent *unison.Panel) *unison.Label {
	icon := unison.NewLabel()
	icon.Accessibility.Role = role.None
	insets := geom.Insets{Left: 4}
	if p.entity == nil {
		insets.Left += xmath.Ceil(unison.DefaultLabelTheme.Font.Baseline())
	}
	icon.SetBorder(unison.NewEmptyBorder(insets))
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

// tintedOn returns the color of text drawn on the color: white or black, as On would pick, with some of the color mixed
// in so that it doesn't read as hard as pure white or black.
func tintedOn(c unison.ThemeColor) unison.ThemeColor {
	on := func(c unison.Color) unison.Color { return c.OnCustom(unison.Black, unison.White).Blend(c, 0.15) }
	return unison.ThemeColor{Light: on(c.Light), Dark: on(c.Dark)}
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
	// The same insets on the sides as a row's, so that the grips and the buttons on the right line up down the panel.
	head.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: 2, Left: 4, Bottom: 2, Right: 8}))
	pill := compactPopup(p, path+keyPill, i18n.Text("Requirement"), []bool{true, false}, list.All, groupWord,
		func(all bool) { list.All = all })
	pill.HMargin = 10
	pill.CornerRadius = geom.NewUniformSize(100)
	pill.BackgroundInk = color
	pill.OnBackgroundInk = color.Derive(tintedOn)
	pill.EdgeInk = unison.Transparent
	desc := pill.Font.Descriptor()
	desc.Weight = weight.Bold
	pill.Font = desc.Font()
	head.ClientData()[prereqDropKey] = path
	if path != prereqRootPath {
		p.grip(head, path)
	}
	// An empty root has nothing for its pill or status to speak of.
	if path != prereqRootPath || len(list.Prereqs) != 0 {
		p.views = append(p.views, prereqView{node: list, icon: p.statusIcon(head), group: box})
		put(head, pill)
	}
	if list.WhenTL.Compare != criteria.AnyNumber {
		p.chip(head, path+":tl", i18n.Text("When TL"), i18n.Text("Remove Tech Level Condition"), path+keyAdd,
			func() { list.WhenTL = criteria.Number{} },
			func(chip *unison.Panel) {
				p.numberCriteria(chip, path+":tl", i18n.Text("Tech Level"), &list.WhenTL, true, 0, fxp.Twelve, true)
			})
	}
	add := newPrereqIconButton(path+keyAdd, unison.CircledAddSVG, i18n.Text("Add to this group"))
	add.ClickCallback = func() { showMenu(add.AsPanel(), p.addEntries(list, path)) }
	fitLine(add)
	add.SetLayoutData(&unison.FlexLayoutData{HAlign: align.End, VAlign: align.Middle, HGrab: true})
	head.AddChild(add)
	if path != prereqRootPath {
		p.moreButton(head, list, path)
	} else {
		// The root's add button is spelled out, as a cue to what the bare ones below do, and keeps room for the more
		// button the root doesn't have, so that it lines up with them.
		add.SetTitle(i18n.Text("Add"))
		add.HideBase = false
		room := newPrereqIconButton("", svg.CircledVerticalEllipsis, "")
		room.Hidden = true
		put(head, room)
	}
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
	p.grip(row, path)
	view := prereqView{node: pr, icon: p.statusIcon(row)}
	var main *unison.Panel
	if open {
		row.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: 6, Left: 4, Bottom: 8, Right: 8}))
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
		row.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: 3, Left: 4, Bottom: 3, Right: 8}))
		var click func()
		if pr.PrereqType() != prereq.Unknown {
			click = func() { p.toggle(path) }
		}
		view.sentence = newSentenceButton(pr.Describe(p.entity, nil, em), click, func() bool { return p.open == path })
		view.sentence.RefKey = path + keySentence
		p.dragBy(view.sentence.AsPanel(), row, path)
		if click == nil {
			// One this version of GCS doesn't understand can't be edited, so its sentence is static text.
			view.sentence.Tooltip = newWrappedTooltip(i18n.Text("This was most likely created by a newer version of GCS. Its original data will be written back out unchanged when this file is saved."))
		}
		main = view.sentence.AsPanel()
		row.AddChild(main)
		if script, ok := pr.(*gurps.ScriptPrereq); ok && script.ResolvedName(nil) == "" {
			describe := newDashedButton(i18n.Text("Add a description"), func() {
				p.open = path
				p.rebuild(path + ":name")
			})
			describe.OnBackgroundInk = unison.ThemeAlert
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
			if _, ok := child.Self.(*unison.Button); ok {
				fitLine(child)
			}
			child.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Start})
		}
	}
	row.DrawCallback = func(gc *unison.Canvas, _ geom.Rect) {
		// Short of the right edge, so that an open row's background doesn't run into the edge of the panel.
		r := row.ContentRect(true)
		r.Width -= 4
		if open {
			gc.DrawRoundedRect(r, geom.NewUniformSize(8), unison.ThemeBelowSurface.Paint(gc, r, paintstyle.Fill))
			return
		}
		gc.DrawLine(geom.NewPoint(r.X, r.Bottom()-0.5), geom.NewPoint(r.Right(), r.Bottom()-0.5),
			faint(unison.ThemeSurfaceEdge).Paint(gc, r, paintstyle.Stroke))
	}
	return row
}

// grip adds a drag handle to the row, which may be a group's head, and lets either be dragged to move the node at the
// path.
func (p *prereqPanel) grip(row *unison.Panel, path string) {
	handle := NewDragHandle(prereqDragKey, nil)
	put(row, handle)
	p.dragBy(handle.AsPanel(), row, path)
	p.dragBy(row, row, path)
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
		size := row.FrameRect().Size
		img, err := unison.NewImageFromDrawing(int(xmath.Ceil(size.Width)), int(xmath.Ceil(size.Height)), 144,
			func(gc *unison.Canvas) {
				r := geom.Rect{Size: size}
				gc.DrawRect(r, unison.ThemeBelowSurface.Paint(gc, r, paintstyle.Fill))
				row.Draw(gc, r)
			})
		if err != nil {
			errs.Log(err)
			return true
		}
		panelDragData = &prereqDrag{panel: p, path: path}
		row.StartDrag(img, geom.Point{}, func() {
			panelDragData = nil
			p.MarkForRedraw()
		}, drag.Move,
			drag.Data{Type: prereqDragKey, Data: []byte{0}})
		return true
	}
	target.MouseUpCallback = func(where geom.Point, button int, mods mod.Modifiers) bool {
		return dragged || up == nil || up(where, button, mods)
	}
}

// toggle opens the row at the path, closing any other, or closes it if it is the open one. The focus goes into the row
// that opens (see keyFirst), or to the sentence of the one that closes.
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
		words(fields, i18n.Text("whose name"))
		p.textCriteria(fields, key("name"), i18n.Text("Name"), i18n.Text("Trait name"), &one.NameCriteria, true)
		level := &one.LevelCriteria
		// Every trait has a level of at least 0, so that is the same as having no level criteria.
		p.levelChip(chips, path, level,
			level.Compare != criteria.AnyNumber && (level.Compare != criteria.AtLeastNumber || level.Qualifier > 0))
		p.textChip(chips, path, "notes", &one.NotesCriteria)
	case *gurps.SkillPrereq:
		p.hasPopup(fields, key("has"), &one.Has, false)
		p.typePopup(fields, path, pr)
		words(fields, i18n.Text("whose name"))
		p.textCriteria(fields, key("name"), i18n.Text("Name"), i18n.Text("Skill name"), &one.NameCriteria, true)
		p.textChip(chips, path, "specialization", &one.SpecializationCriteria)
		p.textChip(chips, path, "optspecialization", &one.OptionalSpecializationCriteria)
		// Unlike a trait's, "at least 0" is a real criterion here, leaving out skills with no usable level.
		p.levelChip(chips, path, &one.LevelCriteria, one.LevelCriteria.Compare != criteria.AnyNumber)
	case *gurps.SpellPrereq:
		p.hasPopup(fields, key("has"), &one.Has, true)
		quantity := func() {
			p.numberCriteria(fields, key("quantity"), i18n.Text("Quantity"), &one.QuantityCriteria, false, 0,
				fxp.FromInteger(9999), true)
		}
		// A count of colleges follows the match, as in "spells from at least 2 colleges".
		colleges := one.SubType == spellcmp.CollegeCount
		if !colleges {
			quantity()
		}
		p.typePopup(fields, path, pr)
		put(fields, compactPopup(p, key("match"), i18n.Text("Spell Match"), spellcmp.Types, one.SubType,
			func(t spellcmp.Type) string {
				if t == spellcmp.CollegeCount {
					return i18n.Text("from")
				}
				return t.String()
			}, func(t spellcmp.Type) { one.SubType = t }))
		if colleges {
			quantity()
			words(fields, i18n.Text("college(s)"))
		}
		if one.SubType.UsesStringCriteria() {
			p.textCriteria(fields, key("qualifier"), i18n.Text("Spell"), "", &one.QualifierCriteria, true)
		}
		p.powerSourceChip(chips, path, one)
	case *gurps.AttributePrereq:
		p.hasPopup(fields, key("has"), &one.Has, false)
		p.typePopup(fields, path, pr)
		flags := gurps.SizeFlag | gurps.DodgeFlag | gurps.ParryFlag | gurps.BlockFlag
		p.attributePopup(fields, key("which"), i18n.Text("Attribute"), &one.Which, flags)
		words(fields, i18n.Text("which"))
		// Named apart from the attribute popup before it.
		p.numberCriteria(fields, key("value"), i18n.Text("Value"), &one.QualifierCriteria, true, fxp.Min, fxp.Max,
			false)
		p.optional(chips, path, "combined", one.CombinedWith != "",
			func() { one.CombinedWith = gurps.AttributeIDFor(p.entity, gurps.DexterityID) },
			func() { one.CombinedWith = "" },
			func(chip *unison.Panel) {
				p.attributePopup(chip, key("combined"), i18n.Text("Combined With"), &one.CombinedWith, flags)
			})
	case *gurps.EquippedEquipmentPrereq:
		p.typePopup(fields, path, pr)
		words(fields, i18n.Text("whose name"))
		p.textCriteria(fields, key("name"), i18n.Text("Name"), i18n.Text("Item name"), &one.NameCriteria, true)
		p.textChip(chips, path, "tag", &one.TagsCriteria)
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
		words(fields, i18n.Text("of"))
		p.numberCompare(fields, key("weightcmp"), comparison, &one.WeightCriteria.Compare, false)
		p.addCompact(fields, NewWeightField(p.targetMgr, key("weight"), title, p.entity,
			func() fxp.Weight { return one.WeightCriteria.Qualifier },
			func(w fxp.Weight) { p.edit(title, key("weight"), func() { one.WeightCriteria.Qualifier = w }) },
			0, fxp.Weight(fxp.Max), false).Field)
	case *gurps.ScriptPrereq:
		p.typePopup(fields, path, pr)
		words(fields, i18n.Text("described as"))
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

// scriptOptions returns the script editor's options for a script prerequisite. Each snippet ends in an expression that
// is true when met and otherwise the reason it isn't. Evaluate, when there is a sheet, runs the script being edited
// against it.
func (p *prereqPanel) scriptOptions(pr *gurps.ScriptPrereq) *scriptEditorOptions {
	opts := &scriptEditorOptions{
		Title:               i18n.Text("Script"),
		Hint:                i18n.Text("Tab indents. Shift+Tab leaves. Esc closes."),
		KeepFirstLinePrefix: gurps.ScriptPrereqCountPrefix,
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
		entries = append(entries, menuEntry{Label: prereqTypeName(t), Act: func() {
			add(i18n.Text("Add Prerequisite"), p.createPrereqForType(t, list))
		}})
	}
	entries = append(entries, menuEntry{Label: i18n.Text("Structure")})
	for i, label := range []string{i18n.Text("All of Group"), i18n.Text("Any of Group")} {
		entries = append(entries, menuEntry{Label: label, Act: func() {
			add(i18n.Text("Add Group"), &gurps.PrereqList{Type: prereq.List, Parent: list, All: i == 0})
		}})
	}
	if list.WhenTL.Compare == criteria.AnyNumber {
		entries = append(entries, menuEntry{Label: i18n.Text("Only When TL…"), Act: func() {
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
	b := newPrereqIconButton(path+keyMore, svg.CircledVerticalEllipsis, i18n.Text("More actions"))
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

// drawDrop dims the prerequisite being dragged, and marks where it would go, in the ink of GCS's other drop markers: a
// line before or after a row or group, or a tint over a group it would go into.
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
	// Short of the right edge, as an open row's background is.
	r.Width -= 4
	if p.dropWhere == dropInto {
		gc.DrawRoundedRect(r, geom.NewUniformSize(6), faint(unison.ThemeWarning).Paint(gc, r, paintstyle.Fill))
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
	var render func(prereq.Type) string
	if current == prereq.EquippedEquipment || current == prereq.Script {
		// With no "has" popup ahead of it, it starts the sentence.
		render = func(t prereq.Type) string { return xstrings.FirstToUpper(t.String()) }
	}
	put(parent, compactPopup(p, path+":type", i18n.Text("Prerequisite Type"), items, current, render,
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

// prereqTypeName returns the name of the type in the Add menu.
func prereqTypeName(t prereq.Type) string {
	switch t {
	case prereq.Trait:
		return i18n.Text("Trait")
	case prereq.Attribute:
		return i18n.Text("Attribute")
	case prereq.ContainedQuantity:
		return i18n.Text("Contained Quantity")
	case prereq.ContainedWeight:
		return i18n.Text("Contained Weight")
	case prereq.EquippedEquipment:
		return i18n.Text("Equipped Equipment")
	case prereq.Skill:
		return i18n.Text("Skill")
	case prereq.Spell:
		return i18n.Text("Spell")
	case prereq.Script:
		return i18n.Text("Script")
	default:
		return t.String()
	}
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
	p.optional(chips, path, "power", one.SamePowerSource || one.PowerSourceCriteria.Compare != criteria.AnyText,
		func() {
			one.SamePowerSource = p.ownerIsSpell
			if !p.ownerIsSpell {
				one.PowerSourceCriteria.Compare = criteria.IsText
			}
		},
		func() { one.SamePowerSource, one.PowerSourceCriteria = false, criteria.Text{} },
		func(chip *unison.Panel) {
			// The choices are the comparisons, as strings, after "is the same as this spell's" when that is offered.
			same := i18n.Text("is the same as this spell's")
			var items []string
			if p.ownerIsSpell || one.SamePowerSource {
				items = append(items, same)
			}
			for _, cmp := range criteria.StringComparisons[1:] {
				items = append(items, cmp.String())
			}
			current := one.PowerSourceCriteria.Compare.String()
			if one.SamePowerSource {
				current = same
			}
			title := i18n.Text("Power Source")
			comparison, _ := criteriaTitles(title)
			put(chip, compactPopup(p, path+":powercmp", comparison, items, current, nil, func(choice string) {
				one.SamePowerSource = choice == same
				one.PowerSourceCriteria.Compare = criteria.AnyText
				for _, cmp := range criteria.StringComparisons[1:] {
					if cmp.String() == choice {
						one.PowerSourceCriteria.Compare = cmp
					}
				}
			}))
			if !one.SamePowerSource {
				p.textField(chip, path+":power", title, "", &one.PowerSourceCriteria.Qualifier)
			}
		})
}

// prereqCriterion describes an optional criterion of a prerequisite: the subject its controls are named for, the
// button that adds it, the words its chip starts with, and the titles of adding and removing it.
type prereqCriterion struct {
	subject, add, words, addTitle, removeTitle string
}

// prereqCriteria returns the optional criteria, by the key their widgets' reference keys are made from.
func prereqCriteria() map[string]*prereqCriterion {
	return map[string]*prereqCriterion{
		"notes": {
			i18n.Text("Notes"), i18n.Text("+ notes"), i18n.Text("and whose notes"), i18n.Text("Add Notes"),
			i18n.Text("Remove Notes"),
		},
		"tag": {
			i18n.Text("Tag"), i18n.Text("+ tag"), i18n.Text("and whose tag"), i18n.Text("Add Tag"),
			i18n.Text("Remove Tag"),
		},
		"specialization": {
			i18n.Text("Specialization"), i18n.Text("+ specialization"), i18n.Text("and whose specialization"),
			i18n.Text("Add Specialization"), i18n.Text("Remove Specialization"),
		},
		"optspecialization": {
			i18n.Text("Optional Specialization"), i18n.Text("+ optional specialization"),
			i18n.Text("and whose optional specialization"), i18n.Text("Add Optional Specialization"),
			i18n.Text("Remove Optional Specialization"),
		},
		"level": {
			i18n.Text("Level"), i18n.Text("+ level"), i18n.Text("and whose level"), i18n.Text("Add Level"),
			i18n.Text("Remove Level"),
		},
		"power": {
			i18n.Text("Power Source"), i18n.Text("+ power source"), i18n.Text("and whose power source"),
			i18n.Text("Add Power Source"), i18n.Text("Remove Power Source"),
		},
		"combined": {
			i18n.Text("Combined With"), i18n.Text("+ combined with"), i18n.Text("combined with"),
			i18n.Text("Add Combined Attribute"), i18n.Text("Remove Combined Attribute"),
		},
	}
}

// textChip adds the optional text criterion with the key, which is in use while its comparison isn't "is anything".
func (p *prereqPanel) textChip(chips *unison.Panel, path, key string, c *criteria.Text) {
	p.optional(chips, path, key, c.Compare != criteria.AnyText,
		func() { *c = criteria.Text{Compare: criteria.IsText} },
		func() { *c = criteria.Text{} },
		func(chip *unison.Panel) {
			p.textCriteria(chip, path+":"+key, prereqCriteria()[key].subject, "", c, false)
		})
}

// levelChip adds an optional level criterion.
func (p *prereqPanel) levelChip(chips *unison.Panel, path string, level *criteria.Number, on bool) {
	p.optional(chips, path, "level", on,
		func() { *level = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.One} },
		func() { *level = criteria.Number{} },
		func(chip *unison.Panel) {
			p.numberCriteria(chip, path+":level", i18n.Text("Level"), level, true, 0, fxp.Thousand, false)
		})
}

// optional adds the optional criterion with the key to chips: when on, as a chip holding the controls populate adds;
// otherwise as a dashed chip that adds it.
func (p *prereqPanel) optional(chips *unison.Panel, path, key string, on bool, add, remove func(), populate func(chip *unison.Panel)) {
	c := prereqCriteria()[key]
	addKey := path + ":add " + key
	chipKey := path + ":" + key + keyChip
	if on {
		p.chip(chips, path+":"+key, c.words, c.removeTitle, addKey, remove, populate)
		return
	}
	b := newDashedButton(c.add, func() {
		if after := p.edit(c.addTitle, addKey, add); after != nil {
			after.focus = chipKey
		}
		p.rebuild(chipKey)
	})
	b.RefKey = addKey
	fitLine(b)
	put(chips, b)
}

// chip adds a rounded chip, whose reference key is key followed by keyChip, to the parent with the text, the controls
// populate adds, and a button titled title that calls remove and then gives the focus to the widget with the reference
// key after.
func (p *prereqPanel) chip(parent *unison.Panel, key, text, title, after string, remove func(), populate func(chip *unison.Panel)) {
	chip := unison.NewPanel()
	chip.RefKey = key + keyChip
	chip.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: chipPadding, Left: 10, Bottom: chipPadding, Right: 3}))
	chip.DrawCallback = func(gc *unison.Canvas, _ geom.Rect) {
		r := chip.ContentRect(true)
		unison.DrawRoundedRectBase(gc, r, geom.NewUniformSize(r.Height/2), 1, unison.ThemeSurface,
			unison.ThemeSurfaceEdge)
	}
	words(chip, text)
	populate(chip)
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

// numberCriteria adds the comparison of a numeric criterion, worded as numberCompare words it, and, unless it is
// "anything", the field for its qualifier, which takes whole numbers when integer is true.
func (p *prereqPanel) numberCriteria(parent *unison.Panel, key, subject string, c *criteria.Number, is bool, minValue, maxValue fxp.Int, integer bool) {
	comparison, _ := criteriaTitles(subject)
	p.numberCompare(parent, key+"cmp", comparison, &c.Compare, is)
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

// numberCompare adds the popup for a numeric comparison: with is, worded to follow a subject, as in "whose level is at
// least"; otherwise in the short words that come before a quantity, as in "at least 2 spells".
func (p *prereqPanel) numberCompare(parent *unison.Panel, key, name string, c *criteria.NumericComparison, is bool) {
	items := criteria.NumericComparisons
	// "anything" is left to a chip's remove button, but is shown when a file edited by hand holds it.
	if *c != criteria.AnyNumber {
		items = items[1:]
	}
	put(parent, compactPopup(p, key, name, items, *c, func(v criteria.NumericComparison) string {
		if is {
			return v.String()
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
	// Padded to the height of the controls around it, less the height of its text.
	pad := (controlHeight(parent) - field.Font.LineHeight()) / 2
	unison.InstallFocusBorders(field, field, compactFieldBorder(true, pad), compactFieldBorder(false, pad))
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

// compactFieldBorder returns the border of a compact field, which leaves pad above and below its text.
func compactFieldBorder(focused bool, pad float32) unison.Border {
	ink := unison.Ink(unison.ThemeSurfaceEdge)
	var w float32 = 1
	if focused {
		ink = unison.ThemeFocus
		w = 2
	}
	return unison.NewCompoundBorder(unison.NewLineBorder(ink, geom.NewUniformSize(4), geom.NewUniformInsets(w), false),
		unison.NewEmptyBorder(geom.Insets{Top: pad - w, Left: 6 - w, Bottom: pad - w, Right: 6 - w}))
}

// chipPadding is the room a chip leaves above and below the controls it holds.
const chipPadding = 2

// controlHeight returns the height of a control in the parent: within a chip, that of a standard button; anywhere else,
// that of a chip holding one, so that everything on a line, chips and group pills included, is as tall.
func controlHeight(parent *unison.Panel) float32 {
	theme := &unison.DefaultButtonTheme
	height := xmath.Ceil(theme.Font.LineHeight()) + 2*(theme.VMargin+1)
	if parent == nil || !strings.HasSuffix(parent.RefKey, keyChip) {
		height += 2 * chipPadding
	}
	return height
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
		p.edit(name, key, func() { set(v) })
		p.rebuild("")
	})
	return popup
}

// words adds text that joins the controls of an open row into a sentence.
func words(parent *unison.Panel, text string) {
	label := unison.NewLabel()
	label.SetTitle(text)
	put(parent, label)
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

// newDashedButton returns a button drawn as a dashed outline, which adds something. Under the pointer it is filled, and
// its outline is solid and in the focus color.
func newDashedButton(title string, click func()) *unison.Button {
	b := unison.NewButton()
	b.HideBase = true
	b.OnBackgroundInk = unison.ThemeOnSurface
	b.CornerRadius = geom.NewUniformSize(100)
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
