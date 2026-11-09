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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/levelcap"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/unison"
	"github.com/zeebo/xxh3"
)

// levelCapMatcherKey starts the reference keys of the controls that say what a cap applies to. Unlike a node's path,
// it doesn't start with treeRootPath, so it is never taken for one.
const levelCapMatcherKey = "m"

// levelCapValueTypes are the kinds of cap value the editor offers, in the order it offers them.
var levelCapValueTypes = []levelcap.Type{
	levelcap.Fixed,
	levelcap.Skill,
	levelcap.Spell,
	levelcap.Trait,
	levelcap.Attribute,
}

// skillLevelCapPanel edits a skill level cap in place: at the top, the controls that say which skills and spells it
// applies to, and below them the tree of values it caps them at. Each value is a row that reads as a sentence until it
// is opened, one at a time, to edit it. Each group's head says whether it takes the lowest or the highest of its
// children, which hang from a rail in the head's color. Every change, typing included, records a snapshot of the whole
// cap with the undo manager of the dialog that holds the panel.
type skillLevelCapPanel struct {
	sentenceTree[skillLevelCapState, gurps.LevelCap]
	// entity, which may be nil, defines the attributes the values can rest on.
	entity   *gurps.Entity
	levelCap *gurps.SkillLevelCap
	// headed has an empty root show its head, once a group type has been chosen for it.
	headed bool
	// rebuilt, if set, is called after each build, which follows every change to the tree.
	rebuilt func()
}

// skillLevelCapState is the data a snapshot of the panel holds: the cap, and whether an empty root shows its head.
type skillLevelCapState struct {
	levelCap *gurps.SkillLevelCap
	headed   bool
}

// newSkillLevelCapPanel creates the editor for the cap, which it edits in place. entity, which may be nil, defines the
// attributes the cap's values can rest on.
func newSkillLevelCapPanel(entity *gurps.Entity, levelCap *gurps.SkillLevelCap) *skillLevelCapPanel {
	p := &skillLevelCapPanel{entity: entity, levelCap: levelCap}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	// No titled border: this panel is the whole of a dialog, which supplies the title.
	p.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(4)))
	p.initRows(levelCapDragKey, p.build, p.state, p.setState, p.stateHash)
	// An Escape that closes no row cancels the dialog.
	p.passEscape = true
	p.staticTip = func(string) string { return unknownLevelCapTooltip() }
	p.initTree(p)
	p.build()
	return p
}

func (p *skillLevelCapPanel) state() skillLevelCapState {
	return skillLevelCapState{levelCap: p.levelCap.Clone(), headed: p.headed}
}

// setState copies the state into the cap being edited rather than replacing it, since the dialog's caller holds it.
func (p *skillLevelCapPanel) setState(s skillLevelCapState) {
	*p.levelCap = *s.levelCap.Clone()
	p.headed = s.headed
}

func (p *skillLevelCapPanel) stateHash() uint64 {
	h := xxh3.New()
	p.levelCap.Hash(h)
	xhash.Bool(h, p.headed)
	return h.Sum64()
}

func (p *skillLevelCapPanel) build() {
	if !levelCapEditable(p.node(p.open)) {
		p.open = ""
	}
	p.AddChild(p.matcher())
	p.AddChild(p.group(p.levelCap.Cap, treeRootPath))
	if p.rebuilt != nil {
		p.rebuilt()
	}
}

// levelCapEditable reports whether the node is a value of a kind the editor can open.
func levelCapEditable(node gurps.LevelCap) bool {
	switch node.(type) {
	case *gurps.FixedLevelCap, *gurps.ReferenceLevelCap:
		return true
	default:
		return false
	}
}

// matcher returns the controls that say which skills and spells the cap applies to, flowing as a sentence would, then
// its optional criteria as chips, offering only those its target consults, then the words that lead into the values.
func (p *skillLevelCapPanel) matcher() *unison.Panel {
	lc := p.levelCap
	box := newColumn()
	// In line with the root's pill below it.
	box.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: 8, Right: 8}))
	fields := newFlow()
	box.AddChild(fields)
	key := levelCapMatcherKey + ":"
	addCentered(fields, compactPopup(&p.sentenceRows, key+"target", i18n.Text("Applies To"), levelcap.Targets,
		lc.Target, levelcap.Target.String, func(t levelcap.Target) { setLevelCapTarget(lc, t) }))
	whose := i18n.Text("whose name")
	p.textCriteria(fields, key+"name", i18n.Text("Name"), levelCapNameHint(lc.Target), whose, whose, &lc.NameCriteria,
		true)
	chips := newFlow()
	if lc.Target.MatchesSkills() {
		p.textChip(chips, levelCapMatcherKey, "specialization", &lc.SpecializationCriteria)
	}
	if lc.Target.MatchesSpells() {
		p.collegesChip(chips, levelCapMatcherKey, &lc.CollegeCriteria)
		p.textChip(chips, levelCapMatcherKey, "power", &lc.PowerSourceCriteria)
	}
	p.tagsChip(chips, levelCapMatcherKey, &lc.TagsCriteria)
	// The buttons that add unused criteria go after the chips in use, so an added one takes its place among them.
	for _, child := range slices.Clone(chips.Children()) {
		if _, ok := child.Self.(*unison.Button); ok {
			chips.AddChild(child)
		}
	}
	box.AddChild(chips)
	addJoiningWords(box, i18n.Text("are capped at"))
	return box
}

// setLevelCapTarget sets what the cap applies to, clearing the criteria the target no longer consults, which nothing
// would show, yet which would still be written out.
func setLevelCapTarget(lc *gurps.SkillLevelCap, target levelcap.Target) {
	lc.Target = target
	if !target.MatchesSkills() {
		lc.SpecializationCriteria = criteria.Text{}
	}
	if !target.MatchesSpells() {
		lc.CollegeCriteria = criteria.Text{}
		lc.PowerSourceCriteria = criteria.Text{}
	}
}

// levelCapNameHint returns the hint the matcher's name field shows while empty.
func levelCapNameHint(target levelcap.Target) string {
	switch target {
	case levelcap.Spells:
		return i18n.Text("Spell name")
	case levelcap.SkillsAndSpells:
		return i18n.Text("Skill or spell name")
	default:
		return i18n.Text("Skill name")
	}
}

// levelCapReferenceNames returns the name a screen reader gives the name field of a value resting on a skill, spell or
// trait of the type, apart from the matcher's, and the hint the field shows while empty.
func levelCapReferenceNames(capType levelcap.Type) (subject, hint string) {
	switch capType {
	case levelcap.Spell:
		return i18n.Text("Spell Name"), i18n.Text("Spell name")
	case levelcap.Trait:
		return i18n.Text("Trait Name"), i18n.Text("Trait name")
	default:
		return i18n.Text("Skill Name"), i18n.Text("Skill name")
	}
}

// row returns the panel for a value: its sentence, or while it is open its editor, beside a button for more actions. A
// value of a kind this version of GCS doesn't know is static text.
func (p *skillLevelCapPanel) row(node gurps.LevelCap, path string) *unison.Panel {
	return p.sentenceRow(path, func() string { return node.Describe(p.entity, emphasize) }, levelCapEditable(node),
		func() *unison.Panel { return p.editor(node, path) }, func() []menuEntry { return p.moreEntries(node, path) },
		nil, nil)
}

// editor returns the controls for an open value, flowing as its sentence does: its kind, then the level of a fixed
// value, or what picks the skill, spell, trait or attribute it rests on and the adjustment to that, then a skill's
// specialization as a chip.
func (p *skillLevelCapPanel) editor(node gurps.LevelCap, path string) *unison.Panel {
	box := newColumn()
	box.RefKey = path + keyFirst
	fields := newFlow()
	box.AddChild(fields)
	key := func(name string) string { return path + ":" + name }
	p.typePopup(fields, path, node)
	switch one := node.(type) {
	case *gurps.FixedLevelCap:
		title := i18n.Text("Level")
		p.addCompact(fields, NewDecimalField(p.targetMgr, key("level"), title, func() fxp.Int { return one.Level },
			func(v fxp.Int) { p.edit(title, key("level"), "", func() { one.Level = v }) }, -fxp.Thousand, fxp.Thousand,
			false, false).withoutUndo())
	case *gurps.ReferenceLevelCap:
		if one.Type == levelcap.Attribute {
			attributePopup(&p.sentenceRows, fields, key("attribute"), i18n.Text("Attribute"), "", p.entity,
				&one.Attribute, gurps.SizeFlag|gurps.DodgeFlag)
		} else {
			subject, hint := levelCapReferenceNames(one.Type)
			whose := i18n.Text("whose name")
			p.textCriteria(fields, key("name"), subject, hint, whose, whose, &one.NameCriteria, true)
		}
		title := i18n.Text("Modifier")
		p.addCompact(fields, NewDecimalField(p.targetMgr, key("modifier"), title, func() fxp.Int { return one.Modifier },
			func(v fxp.Int) { p.edit(title, key("modifier"), "", func() { one.Modifier = v }) }, -fxp.Thousand,
			fxp.Thousand, true, false).withoutUndo())
		if one.Type == levelcap.Skill {
			chips := newFlow()
			p.textChip(chips, path, "skillspecialization", &one.SpecializationCriteria)
			box.AddChild(chips)
		}
	default:
	}
	return box
}

// typePopup adds the popup that switches a value to another kind, which replaces it with a new value of that kind. A
// value resting on a skill, spell or trait hands its name and adjustment on to another of those.
func (p *skillLevelCapPanel) typePopup(parent *unison.Panel, path string, node gurps.LevelCap) {
	addCentered(parent, compactPopup(&p.sentenceRows, path+":type", i18n.Text("Cap Value Type"), levelCapValueTypes,
		node.CapType(), levelcap.Type.String, func(t levelcap.Type) {
			owner, i := p.locate(path)
			group := asLevelCapGroup(owner)
			if group == nil || i < 0 || i >= len(group.Children) || t == node.CapType() {
				return
			}
			created := p.newLevelCap(t, group)
			levelCapHandOnOf(created).take(levelCapHandOnOf(node))
			group.Children[i] = created
		}))
}

// newLevelCap returns a new value of the type for the group: a fixed level of 12, a skill, spell or trait yet to be
// named, or the entity's DX.
func (p *skillLevelCapPanel) newLevelCap(capType levelcap.Type, group *gurps.LevelCapGroup) gurps.LevelCap {
	if capType == levelcap.Fixed {
		return gurps.NewFixedLevelCap(group, fxp.Twelve)
	}
	return gurps.NewReferenceLevelCap(p.entity, group, capType)
}

// levelCapHandOn holds what of a value survives a change of its kind, each part nil when it has none: the name and the
// adjustment of a value resting on a skill, spell or trait. A fixed level or an attribute hands nothing on and takes
// nothing.
type levelCapHandOn struct {
	name     *criteria.Text
	modifier *fxp.Int
}

func levelCapHandOnOf(node gurps.LevelCap) levelCapHandOn {
	if one, ok := node.(*gurps.ReferenceLevelCap); ok && one.Type != levelcap.Attribute {
		return levelCapHandOn{name: &one.NameCriteria, modifier: &one.Modifier}
	}
	return levelCapHandOn{}
}

// take copies into h each part that it and from both have.
func (h levelCapHandOn) take(from levelCapHandOn) {
	if h.name != nil && from.name != nil {
		*h.name = *from.name
	}
	if h.modifier != nil && from.modifier != nil {
		*h.modifier = *from.modifier
	}
}

// addEntries returns what can be added to a group: under the heading value, a value of each kind, which opens at once,
// and under the heading structure, groups.
func (p *skillLevelCapPanel) addEntries(group *gurps.LevelCapGroup, path, value, structure string) []menuEntry {
	entries := make([]menuEntry, 0, len(levelCapValueTypes)+4)
	entries = append(entries, menuEntry{Label: value})
	for _, t := range levelCapValueTypes {
		entries = append(entries, menuEntry{Label: t.AltString(), Act: func() {
			at := childPath(path, len(group.Children))
			p.edit(i18n.Text("Add Cap Value"), p.addKey(group, path), at+keyFirst, func() {
				group.Children = append(group.Children, p.newLevelCap(t, group))
				p.open = at
			})
		}})
	}
	entries = append(entries, menuEntry{Label: structure})
	return append(entries, p.groupEntries(group, path, func(all bool) {
		group.Highest = !all
		p.headed = true
	})...)
}

// unknownLevelCapTooltip returns the tooltip that explains the row of a value of a kind the editor doesn't know, which
// it shows but can't edit. It is looked up when needed rather than held in a variable, since the localization isn't in
// place when the package initializes.
func unknownLevelCapTooltip() string {
	return i18n.Text("This was most likely created by a newer version of GCS, which is needed to make use of it. This version caps nothing with it, but its original data will be written back out unchanged when this cap is saved.")
}
