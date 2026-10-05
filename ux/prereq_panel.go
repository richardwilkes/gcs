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
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
	"github.com/richardwilkes/unison/enums/weight"
	"github.com/zeebo/xxh3"
)

// A node of the tree is found by its path: prereqRootPath for the root, then the index of each child on the way down,
// separated by dots. A widget's reference key is the path of its node followed by one of the suffixes below, one of
// those of sentenceRows, or a colon and a name of its own.
const (
	prereqRootPath = "r"
	keyAdd         = ":add"
	keyPill        = ":pill"
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

const (
	maxPrereqQuantity = 9999
	defaultWhenTL     = 3
)

// prereqPanel edits a tree of prerequisites. Each one is a row that reads as a sentence until it is opened, one at a
// time, to edit it. Each list is a group, whose head says whether all or any of its children must be met and whose
// children hang from a rail in the head's color. Every change, typing included, records a snapshot of the whole tree
// with the editor's undo manager, and changes to the tree's shape rebuild the panel's content. Clicking the title
// collapses the panel to a paragraph describing the whole tree. It starts out collapsed when there are prerequisites,
// and open to add one when there are none.
type prereqPanel struct {
	sentenceRows[prereqState]
	entity           *gurps.Entity
	root             **gurps.PrereqList
	placeholder      *gurps.PrereqList
	permittedChoices []prereq.Type
	summary          *sentenceButton
	views            []prereqView
	target           any
	hash             uint64
	ownerIsSpell     bool
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
	border := initTitledEditorSection(p, i18n.Text("Prerequisites"))
	p.initRows(prereqDragKey, p.build, p.state, p.setState, p.stateHash)
	p.initCollapse(border, len(p.tree().Prereqs) != 0)
	p.initDrop(func(where geom.Point, data any) (*unison.Panel, int) {
		target, _, at := p.dropAt(where, data)
		return target, at
	}, p.drop)
	p.build()
	// The item being edited, which evaluation leaves out, can only be found once the panel is in its editor.
	unison.InvokeTask(p.refresh)
	return p
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

// prereqState is the data a snapshot of the panel holds: the tree, and whether an empty root shows its head.
type prereqState struct {
	tree   *gurps.PrereqList
	headed bool
}

func cloneTree(list *gurps.PrereqList) *gurps.PrereqList {
	if list == nil {
		return nil
	}
	return list.CloneAsPrereqList(nil)
}

// state returns a copy of the panel's data. Since edit takes one ahead of each change, it then makes a stand-in for a
// missing tree the real one, so that the change lands in it; the copy keeps the tree missing, for undo.
func (p *prereqPanel) state() prereqState {
	s := prereqState{tree: cloneTree(*p.root), headed: p.headed}
	*p.root = p.tree()
	return s
}

// setState returns the panel's data to a copy of the state, for undo and redo.
func (p *prereqPanel) setState(s prereqState) {
	*p.root = cloneTree(s.tree)
	p.placeholder = nil
	p.headed = s.headed
}

// stateHash returns a hash of the panel's data.
func (p *prereqPanel) stateHash() uint64 {
	h := xxh3.New()
	p.tree().Hash(h)
	xhash.Bool(h, p.headed)
	return h.Sum64()
}

// restructure changes the shape of the tree through the widget with the reference key from, then rebuilds. change
// returns the node that is the result, whose more button takes the focus; with none, the widget with the reference key
// fallback does. The open row stays open wherever it ends up, and closes if it is removed.
func (p *prereqPanel) restructure(title, from, fallback string, change func() (dst gurps.Prereq)) {
	open := p.node(p.open)
	p.sentenceRows.restructure(title, from, fallback, func() string {
		dst := change()
		p.open = p.pathOf(open)
		if dst == nil {
			return ""
		}
		return p.pathOf(dst) + keyMore
	})
}

func (p *prereqPanel) build() {
	if node := p.node(p.open); node == nil || node.PrereqType() == prereq.List || node.PrereqType() == prereq.Unknown {
		p.open = ""
	}
	p.views = p.views[:0]
	// Collapsed, the tree reads as one paragraph, which refresh keeps current; expanded, there is none.
	if p.summary = p.addTitleBar(p.summaryText); p.summary == nil {
		p.AddChild(p.group(p.tree(), prereqRootPath))
	}
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

// summaryText returns the paragraph a collapsed panel shows: the whole tree as a sentence, ended with a period.
func (p *prereqPanel) summaryText() string {
	if text := p.tree().Describe(p.entity, nil, emphasize); text != "" {
		return fmt.Sprintf(i18n.Text("%s."), text)
	}
	return i18n.Text("No prerequisites.")
}

// refresh updates the summary, the sentences and the status icons from the tree, in place. The tree's scripts run only
// when there are rows to show their status.
func (p *prereqPanel) refresh() {
	p.hash = gurps.Hash64(p.tree())
	if p.summary != nil {
		p.summary.setText(p.summaryText(), "")
	}
	var checks map[gurps.Prereq]prereqCheck
	if p.entity != nil && p.Parent() != nil && len(p.views) != 0 {
		checks = p.checks()
	}
	for _, v := range p.views {
		var suffix string
		if checks != nil {
			var status gurps.PrereqResult
			var tip string
			status, tip, suffix = p.status(v.node, checks)
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

// prereqCheck is the outcome of checking a node against the sheet, with the error of a script that couldn't run.
type prereqCheck struct {
	status gurps.PrereqResult
	reason string
}

// checks returns the outcome of checking each node of the tree against the sheet, running each script once.
func (p *prereqPanel) checks() map[gurps.Prereq]prereqCheck {
	checks := make(map[gurps.Prereq]prereqCheck)
	gurps.SuppressScriptResolveErrorLogging(func() {
		p.tree().Evaluate(p.entity, p.exclude(), func(one gurps.Prereq, result gurps.PrereqResult, reason string) {
			checks[one] = prereqCheck{status: result, reason: reason}
		})
	})
	return checks
}

// status returns the node's status from the checks, the tooltip of its icon and what a screen reader hears after its
// sentence.
func (p *prereqPanel) status(node gurps.Prereq, checks map[gurps.Prereq]prereqCheck) (status gurps.PrereqResult, tip, suffix string) {
	result := checks[node]
	if result.status == gurps.PrereqSkipped && node == p.tree() {
		// The sheet counts a top level with nothing left to check as met.
		result.status = gurps.PrereqMet
	}
	switch result.status {
	case gurps.PrereqMet:
		return gurps.PrereqMet, i18n.Text("Met"), i18n.Text("met")
	case gurps.PrereqSkipped:
		if list, ok := node.(*gurps.PrereqList); ok && list.AppliesWithParentsAt(p.entity) {
			switch {
			case len(list.Prereqs) == 0:
				return gurps.PrereqSkipped, i18n.Text("Empty group, left out of the check"),
					i18n.Text("empty group, left out of the check")
			case list.HasNothingToCheck():
				return gurps.PrereqSkipped, i18n.Text("Holds nothing to check, left out of the check"),
					i18n.Text("holds nothing to check, left out of the check")
			default:
				return gurps.PrereqSkipped,
					i18n.Text("Nothing in it applies at this tech level, left out of the check"),
					i18n.Text("nothing in it applies at this tech level, left out of the check")
			}
		}
		return gurps.PrereqSkipped, i18n.Text("Doesn't apply at this tech level"),
			i18n.Text("doesn't apply at this tech level")
	default:
	}
	var buffer xbytes.InsertBuffer
	gurps.SuppressScriptResolveErrorLogging(func() { node.Satisfied(p.entity, p.exclude(), &buffer, "\n- ", nil) })
	// One unmet item reads as a sentence; more are a list.
	reason := buffer.String()
	if strings.Count(reason, "\n") == 1 {
		tip = fmt.Sprintf(i18n.Text("Not met: %s"), strings.TrimPrefix(reason, "\n- "))
	} else {
		tip = i18n.Text("Not met:") + reason
	}
	if result.status == gurps.PrereqUnmet {
		return gurps.PrereqUnmet, tip, i18n.Text("not met")
	}
	if _, isScript := node.(*gurps.ScriptPrereq); isScript {
		return gurps.PrereqFailed, tip, fmt.Sprintf(i18n.Text("couldn't run: %s"), result.reason)
	}
	return gurps.PrereqFailed, tip, i18n.Text("couldn't be checked")
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
	box := newColumn()
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
	pill := compactPopup(&p.sentenceRows, path+keyPill, i18n.Text("Requirement"), []bool{true, false}, list.All, groupWord,
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
		empty = newEmptyPlaceholder(path+":empty", text, nil)
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
		empty.ClientData()[prereqDropKey] = prereqDropSpot{path: path, part: dropOnEmpty}
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
	add := newIconButton(path+keyAdd, unison.CircledAddSVG, i18n.Text("Add to this group"))
	add.ClickCallback = func() { showMenu(add.AsPanel(), p.addEntries(list, path)) }
	fitLine(add)
	add.SetLayoutData(&unison.FlexLayoutData{HAlign: align.End, VAlign: align.Middle, HGrab: !emptyRoot})
	head.AddChild(add)
	if path != prereqRootPath {
		p.moreButton(head, list, path)
	} else {
		// The root keeps room for the more button it doesn't have, so that its add button lines up with the others.
		room := newIconButton("", svg.CircledVerticalEllipsis, "")
		room.Hidden = true
		addCentered(head, room)
	}
	// As tall as the add button, since centering a shorter one could put its icon on a half pixel, blurring it.
	fitLine(head.Children()[len(head.Children())-1])
	box.AddChild(hbox(head, unison.StdHSpacing))

	rail := newColumn()
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
	var icon *unison.Label
	row := p.sentenceRow(path, func() string { return pr.Describe(p.entity, nil, emphasize) },
		pr.PrereqType() != prereq.Unknown, func() *unison.Panel { return p.editor(pr, path) },
		func() []menuEntry { return p.moreEntries(pr, path) },
		func(row *unison.Panel) (*unison.Panel, float32) {
			if icon = p.statusIcon(row); icon == nil {
				return nil, 0
			}
			if path == p.open {
				// An open row shows no status, but keeps the room for it, so that its editor lines up with the sentences.
				showCheckIcon(icon, gurps.PrereqMet)
				icon.OnBackgroundInk = unison.Transparent
			}
			return icon.AsPanel(), checkIconSize()
		},
		func(row *unison.Panel, sentence *sentenceButton) {
			if script, ok := pr.(*gurps.ScriptPrereq); ok && script.ResolvedName(nil) == "" {
				describe := newDashedButton(i18n.Text("Add a description"), func() {
					p.open = path
					p.rebuild(path + ":name")
				})
				describe.OnBackgroundInk = unison.ThemeAlert
				addCentered(row, describe)
			}
			p.views = append(p.views, prereqView{node: pr, icon: icon, sentence: sentence})
		})
	row.ClientData()[prereqDropKey] = prereqDropSpot{path: path, part: dropOnRow}
	return row
}

// editor returns the controls for an open row: those that say what the prerequisite requires, flowing as a sentence
// would, then its optional criteria as chips, or the script editor for a script.
func (p *prereqPanel) editor(pr gurps.Prereq, path string) *unison.Panel {
	box := newColumn()
	box.RefKey = path + keyFirst
	fields := newFlow()
	box.AddChild(fields)
	chips := newFlow()
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
		addCentered(fields, compactPopup(&p.sentenceRows, key("match"), i18n.Text("Spell Match"), spellcmp.Types, one.SubType,
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
		p.optionalCriterion(chips, path, "combined", one.CombinedWith != "",
			func() { one.CombinedWith = gurps.AttributeIDFor(p.entity, gurps.DexterityID) },
			func() { one.CombinedWith = "" },
			func(chip *unison.Panel) {
				p.attributePopup(chip, key("combined"), i18n.Text("Combined With"), rowCriteria("combined").prefix,
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
		opts.Evaluate = func(script string) (gurps.PrereqResult, string) {
			one := *pr
			one.Script = script
			status, reason := one.Evaluate(p.entity, p.exclude())
			reason = strings.TrimSpace(reason)
			switch {
			case status == gurps.PrereqMet:
				return status, i18n.Text("Passed")
			case status == gurps.PrereqFailed:
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
	addMoreButton(parent, path, func() []menuEntry { return p.moreEntries(node, path) })
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

// dropAt returns the panel a prerequisite dragged to where would be dropped on, the path of its node and where it would
// go: before or after a row, before a group over the top of its head and into it below that, after a group beside or
// below its last child, or into an empty group. The panel is nil where nothing can go, such as into itself.
func (p *prereqPanel) dropAt(where geom.Point, data any) (target *unison.Panel, path string, at int) {
	from := p.dragPath(data)
	if p.node(from) == nil {
		return nil, "", 0
	}
	for target = p.PanelAt(where); target != nil && target != p.AsPanel(); target = target.Parent() {
		spot, isTarget := target.ClientData()[prereqDropKey].(prereqDropSpot)
		if !isTarget {
			continue
		}
		if spot.path == from || strings.HasPrefix(spot.path, from+".") {
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

// drop moves the dragged prerequisite to where it would go.
func (p *prereqPanel) drop(where geom.Point, data any) {
	target, path, at := p.dropAt(where, data)
	p.dragExit()
	if target == nil {
		return
	}
	to, index := p.locate(path)
	if group, isList := p.node(path).(*gurps.PrereqList); isList && at == dropInto {
		to, index = group, len(group.Prereqs)
	} else if at == dropAfter {
		index++
	}
	dragged := p.dragPath(data)
	from, i := p.locate(dragged)
	if from == to && i < index {
		// Taking the node out moves what comes after it up by one.
		index--
	}
	p.restructure(i18n.Text("Move Prerequisite"), dragged+keyMore, "", func() gurps.Prereq {
		if from == to && i == index {
			return nil
		}
		return relocate(from, i, to, index)
	})
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
	addCentered(parent, compactPopup(&p.sentenceRows, path+":type", i18n.Text("Prerequisite Type"), items, current, render,
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
	addCentered(parent, compactPopup(&p.sentenceRows, key, i18n.Text("Has"), []bool{true, false}, *has, func(v bool) string {
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
	addCentered(parent, compactPopup(&p.sentenceRows, key, name, choices, current,
		func(c *gurps.AttributeChoice) string { return c.Title }, func(c *gurps.AttributeChoice) { *value = c.Key }))
}

// powerSourceChip adds the optional power source criterion of a spell prerequisite. "Same as this spell's" is offered
// only to a spell's own prerequisites, or shown when a file edited by hand holds it elsewhere.
func (p *prereqPanel) powerSourceChip(chips *unison.Panel, path string, one *gurps.SpellPrereq) {
	p.optionalCriterion(chips, path, "power", one.SamePowerSource || one.PowerSourceCriteria.Compare != criteria.AnyText,
		func() {
			one.SamePowerSource = p.ownerIsSpell
			if !p.ownerIsSpell {
				one.PowerSourceCriteria.Compare = criteria.IsText
			}
		},
		func() { one.SamePowerSource, one.PowerSourceCriteria = false, criteria.Text{} },
		func(chip *unison.Panel) {
			prefix := rowCriteria("power").prefix
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
			addCentered(chip, compactPopup(&p.sentenceRows, path+":powercmp", comparison, items, current, nil, func(choice string) {
				one.SamePowerSource = choice == same
				one.PowerSourceCriteria.Compare = criteria.StringComparisons[max(slices.Index(choices, choice), 0)]
			}))
			if !one.SamePowerSource {
				p.textField(chip, path+":power", title, "", &one.PowerSourceCriteria.Qualifier)
			}
		})
}

// levelChip adds an optional level criterion.
func (p *prereqPanel) levelChip(chips *unison.Panel, path string, level *criteria.Number, on bool) {
	p.optionalCriterion(chips, path, "level", on,
		func() { *level = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.One} },
		func() { *level = criteria.Number{} },
		func(chip *unison.Panel) {
			p.numberCriteria(chip, path+":level", i18n.Text("Level"), rowCriteria("level").prefix, level, 0,
				fxp.Thousand, false)
		})
}
