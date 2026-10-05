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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellcmp"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/role"
	"github.com/zeebo/xxh3"
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
	sentenceTree[prereqState, gurps.Prereq]
	entity           *gurps.Entity
	root             **gurps.PrereqList
	placeholder      *gurps.PrereqList
	permittedChoices []prereq.Type
	paragraph        *sentenceButton
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
	p.initTree(p)
	p.build()
	// The item being edited, which evaluation leaves out, can only be found once the panel is in its editor. A collapsed
	// panel shows no statuses, so it is left as built.
	unison.InvokeTask(func() {
		if len(p.views) != 0 {
			p.refresh()
		}
	})
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

func (p *prereqPanel) build() {
	if node := p.node(p.open); node == nil || node.PrereqType() == prereq.List || node.PrereqType() == prereq.Unknown {
		p.open = ""
	}
	p.views = p.views[:0]
	// Collapsed, the tree reads as one paragraph, which refresh keeps current; expanded, there is none.
	if p.paragraph = p.addTitleBar(p.summary); p.paragraph != nil {
		// The paragraph already describes the tree as it is, and there are no statuses to work out.
		p.hash = gurps.Hash64(p.tree())
		return
	}
	p.AddChild(p.group(p.tree(), treeRootPath))
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

// summary returns the paragraph a collapsed panel shows: the whole tree as a sentence, ended with a period. A tree with
// nothing to check says so, since a tech level condition is all it would otherwise describe.
func (p *prereqPanel) summary() string {
	tree := p.tree()
	if tree.HasNothingToCheck() {
		return i18n.Text("No prerequisites.")
	}
	return i18n.Text("%s.", tree.Describe(p.entity, nil, emphasize))
}

// refresh updates the paragraph, the sentences and the status icons from the tree, in place. The tree's scripts run only
// when there are rows to show their status.
func (p *prereqPanel) refresh() {
	p.hash = gurps.Hash64(p.tree())
	if p.paragraph != nil {
		p.paragraph.setText(p.summary(), "")
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
		tip = i18n.Text("Not met: %s", strings.TrimPrefix(reason, "\n- "))
	} else {
		tip = i18n.Text("Not met:") + reason
	}
	if result.status == gurps.PrereqUnmet {
		return gurps.PrereqUnmet, tip, i18n.Text("not met")
	}
	if _, isScript := node.(*gurps.ScriptPrereq); isScript {
		return gurps.PrereqFailed, tip, i18n.Text("couldn't run: %s", result.reason)
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

// groupName returns the accessible name of a group.
func groupName(list *gurps.PrereqList) string {
	name := groupWord(list.All)
	if list.WhenTL.Compare != criteria.AnyNumber {
		name += i18n.Text(", only when TL %s", list.WhenTL.AltString())
	}
	return name
}

func groupWord(all bool) string {
	if all {
		return i18n.Text("All of")
	}
	return i18n.Text("Any of")
}

// asPrereqList returns the node as a list, or nil if it isn't one.
func asPrereqList(node gurps.Prereq) *gurps.PrereqList {
	if list, ok := node.(*gurps.PrereqList); ok {
		return list
	}
	return nil
}

func (p *prereqPanel) treeRoot() gurps.Prereq {
	return p.tree()
}

func (p *prereqPanel) treeChildren(node gurps.Prereq) (*[]gurps.Prereq, bool) {
	if list := asPrereqList(node); list != nil {
		return (*[]gurps.Prereq)(&list.Prereqs), true
	}
	return nil, false
}

func (p *prereqPanel) treeSetParent(node, group gurps.Prereq) {
	node.SetParentList(asPrereqList(group))
}

func (p *prereqPanel) treeClone(node, parent gurps.Prereq) gurps.Prereq {
	return node.Clone(asPrereqList(parent))
}

func (p *prereqPanel) treeAll(group gurps.Prereq) bool {
	list := asPrereqList(group)
	return list != nil && list.All
}

func (p *prereqPanel) treeNewGroup(parent gurps.Prereq, all bool) gurps.Prereq {
	return &gurps.PrereqList{Type: prereq.List, Parent: asPrereqList(parent), All: all}
}

// treeCanUngroup implements treeNodes. A group with a tech level condition can't be ungrouped, since its children
// would lose the condition.
func (p *prereqPanel) treeCanUngroup(group gurps.Prereq) bool {
	list := asPrereqList(group)
	return list != nil && list.WhenTL.Compare == criteria.AnyNumber
}

func (p *prereqPanel) treeTitles(_ gurps.Prereq) treeEditTitles {
	return treeEditTitles{
		duplicate: i18n.Text("Duplicate Prerequisite"),
		move:      i18n.Text("Move Prerequisite"),
		delete:    i18n.Text("Delete Prerequisite"),
	}
}

// treeHasLead implements treeNodes. With a sheet, each head starts with its status icon.
func (p *prereqPanel) treeHasLead() bool {
	return p.entity != nil
}

func (p *prereqPanel) treeGroupName(group gurps.Prereq) string {
	return groupName(asPrereqList(group))
}

// treeGroupHead implements treeNodes: the group's status icon, its pill and, when it has one, the chip of its tech
// level condition.
func (p *prereqPanel) treeGroupHead(group gurps.Prereq, path string, head, box *unison.Panel) *unison.ThemeColor {
	list := asPrereqList(group)
	color := groupColor(list.All)
	p.views = append(p.views, prereqView{node: list, icon: p.statusIcon(head), group: box})
	pill := compactPopup(&p.sentenceRows, path+keyPill, i18n.Text("Requirement"), []bool{true, false}, list.All, groupWord,
		func(all bool) { list.All = all })
	stylePill(pill, color)
	addCentered(head, pill)
	if list.WhenTL.Compare != criteria.AnyNumber {
		p.chip(head, path+":tl", i18n.Text("Remove Tech Level Condition"), p.addKey(list, path),
			func() { list.WhenTL = criteria.Number{} },
			func(chip *unison.Panel) {
				p.numberCriteria(chip, path+":tl", i18n.Text("Tech Level"), numericWordsAfter(i18n.Text("When TL")),
					&list.WhenTL, 0, fxp.Twelve, true, false)
			})
	}
	return color
}

// treeEmpty implements treeNodes. An empty root shows its placeholder alone until a group type or tech level has been
// chosen for it.
func (p *prereqPanel) treeEmpty(_ gurps.Prereq, path string) (text string, bare bool) {
	if path == treeRootPath && !p.headed {
		return i18n.Text("No prerequisites. Click here to add one."), true
	}
	return i18n.Text("Empty group. Add a requirement or drag one here."), false
}

func (p *prereqPanel) treeRow(node gurps.Prereq, path string) *unison.Panel {
	return p.row(node, path)
}

func (p *prereqPanel) treeAddEntries(group gurps.Prereq, path string) []menuEntry {
	return p.addEntries(asPrereqList(group), path)
}

func (p *prereqPanel) treeGroupAdds(group gurps.Prereq, path string) []menuEntry {
	return p.addEntriesUnder(asPrereqList(group), path, i18n.Text("Add Requirement"), i18n.Text("Add Structure"))
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
			p.numberCriteria(fields, key("quantity"), i18n.Text("Quantity"), numericWordsAfter(""),
				&one.QuantityCriteria, 0, fxp.FromInteger(maxPrereqQuantity), true, false)
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
		p.numberCriteria(fields, key("value"), i18n.Text("Value"), numericWordsAfter(i18n.Text("which")),
			&one.QualifierCriteria, fxp.Min, fxp.Max, false, false)
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
		p.numberCriteria(fields, key("quantity"), i18n.Text("Quantity"), numericWordsAfter(""),
			&one.QualifierCriteria, 0, fxp.FromInteger(maxPrereqQuantity), true, false)
	case *gurps.ContainedWeightPrereq:
		p.hasPopup(fields, key("has"), &one.Has, false)
		p.typePopup(fields, path, pr)
		p.weightCriteria(fields, key("weight"), i18n.Text("Weight"), numericWordsAfter(i18n.Text("which")), p.entity,
			&one.WeightCriteria, false)
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
				return status, i18n.Text("Couldn't run: %s", reason)
			case reason == "":
				return status, i18n.Text("Failed")
			default:
				return status, i18n.Text("Failed: %s", reason)
			}
		}
	}
	return opts
}

// addEntries returns the menu of an add button: what can be added to a list, under headings.
func (p *prereqPanel) addEntries(list *gurps.PrereqList, path string) []menuEntry {
	return p.addEntriesUnder(list, path, i18n.Text("Requirement"), i18n.Text("Structure"))
}

// addEntriesUnder returns what can be added to a list, the prerequisites under the heading requirement and groups and
// conditions under the heading structure.
func (p *prereqPanel) addEntriesUnder(list *gurps.PrereqList, path, requirement, structure string) []menuEntry {
	// What is added opens at once; groups are added by groupEntries.
	add := func(title string, created gurps.Prereq) {
		at := childPath(path, len(list.Prereqs))
		p.edit(title, p.addKey(list, path), at+keyFirst, func() {
			list.Prereqs = append(list.Prereqs, created)
			p.open = at
		})
	}
	entries := []menuEntry{{Label: requirement}}
	for _, t := range p.permittedChoices {
		entries = append(entries, menuEntry{Label: t.AltString(), Act: func() {
			add(i18n.Text("Add Prerequisite"), p.createPrereqForType(t, list))
		}})
	}
	entries = append(entries, menuEntry{Label: structure})
	entries = append(entries, p.groupEntries(list, path, func(all bool) { list.All, p.headed = all, true })...)
	if list.WhenTL.Compare == criteria.AnyNumber {
		entries = append(entries, menuEntry{Label: i18n.Text("Only When TL…"), Act: func() {
			p.edit(i18n.Text("Add Tech Level Condition"), p.addKey(list, path), path+":tl"+keyChip, func() {
				list.WhenTL = criteria.Number{Compare: criteria.AtMostNumber, Qualifier: fxp.FromInteger(defaultWhenTL)}
				p.headed = p.headed || path == treeRootPath
			})
		}})
	}
	return entries
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
			owner, i := p.locate(path)
			list := asPrereqList(owner)
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
			p.numberCriteria(chip, path+":level", i18n.Text("Level"), numericWordsAfter(rowCriteria("level").prefix),
				level, 0, fxp.Thousand, false, false)
		})
}
