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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/zeebo/xxh3"
)

var lastDefaultTypeUsed = gurps.DexterityID

// defaultAddKey is the reference key of the section's add button. A default's row is found by its path, which is its
// index in the list.
const defaultAddKey = "add"

// defaultEmptyKey is the reference key of the placeholder shown in place of the rows while there are none.
const defaultEmptyKey = "empty"

// defaultDropKey marks a row a dragged default can be dropped on, holding its path.
const defaultDropKey = "default.drop"

// defaultsPanel edits a list of skill defaults. Each one is a row that reads as a sentence until it is opened, one at a
// time, to edit it. Every change, typing included, records a snapshot of the whole list with the editor's undo
// manager. Clicking the title collapses the panel to one sentence listing every row's description. It starts out
// collapsed when there are defaults, and open to add one when there are none.
//
// The list is never changed in place, by slices.Delete, slices.Insert or a swap, but replaced with a new one: code still
// holding the old slice, because its owner was not properly updated, would otherwise find it changed under it, with a
// nil where slices.Delete zeroed its vacated end, and panic. The list holds no nils, which loading drops.
type defaultsPanel struct {
	sentenceRows[[]*gurps.SkillDefault]
	entity   *gurps.Entity
	owner    nameable.Accesser
	defaults *[]*gurps.SkillDefault
	// otherTypes holds the types the defaults had when the panel was made that the type popup doesn't otherwise
	// offer, such as Dodge from a file, so that a default changed from one can be changed back.
	otherTypes []string
}

func newDefaultsPanel(entity *gurps.Entity, owner nameable.Accesser, defaults *[]*gurps.SkillDefault) *defaultsPanel {
	p := &defaultsPanel{
		entity:   entity,
		owner:    owner,
		defaults: defaults,
	}
	border := initTitledEditorSection(p, i18n.Text("Defaults"))
	p.initRows(defaultDragKey, p.build, p.state, p.setState, p.dataHash)
	p.initCollapse(border, len(*defaults) != 0)
	p.initDrop(p.dropAt, p.drop)
	choices := p.typeChoices()
	for _, one := range *defaults {
		t := one.Type()
		if !slices.Contains(p.otherTypes, t) &&
			!slices.ContainsFunc(choices, func(c *gurps.AttributeChoice) bool { return c.Key == t }) {
			p.otherTypes = append(p.otherTypes, t)
		}
	}
	p.build()
	return p
}

// copyDefaults returns a copy of the list holding copies of its defaults.
func copyDefaults(list []*gurps.SkillDefault) []*gurps.SkillDefault {
	if len(list) == 0 {
		return nil
	}
	clone := make([]*gurps.SkillDefault, len(list))
	for i, one := range list {
		c := *one
		clone[i] = &c
	}
	return clone
}

// state returns a copy of the panel's data, for a snapshot.
func (p *defaultsPanel) state() []*gurps.SkillDefault {
	return copyDefaults(*p.defaults)
}

// setState installs a copy of the data of a snapshot, as a new list.
func (p *defaultsPanel) setState(state []*gurps.SkillDefault) {
	*p.defaults = copyDefaults(state)
}

// dataHash returns a hash of the defaults.
func (p *defaultsPanel) dataHash() uint64 {
	h := xxh3.New()
	xhash.Num64(h, len(*p.defaults))
	for _, one := range *p.defaults {
		one.Hash(h)
	}
	return h.Sum64()
}

// replace installs a new list, which change makes from a copy of the current one; see defaultsPanel.
func (p *defaultsPanel) replace(change func(list []*gurps.SkillDefault) []*gurps.SkillDefault) {
	*p.defaults = change(slices.Clone(*p.defaults))
}

// index returns the index of the default at the path, or -1 if there is none.
func (p *defaultsPanel) index(path string) int {
	if i, err := strconv.Atoi(path); err == nil && i >= 0 && i < len(*p.defaults) {
		return i
	}
	return -1
}

// replacements returns the values the owning item gives the nameable markers in its defaults.
func (p *defaultsPanel) replacements() map[string]string {
	if p.owner != nil {
		return p.owner.NameableReplacements()
	}
	return nil
}

func (p *defaultsPanel) build() {
	if p.index(p.open) < 0 {
		p.open = ""
	}
	// Collapsed, the defaults read as one paragraph.
	if p.addTitleBar(p.summary) != nil {
		return
	}
	add := newSectionAddButton(p, i18n.Text("Add a default"), func() bool {
		def := &gurps.SkillDefault{DefaultType: p.addType()}
		if def.SkillBased() {
			def.Name = criteria.Text{Compare: criteria.IsText}
		}
		path := strconv.Itoa(len(*p.defaults))
		p.edit(i18n.Text("Add Default"), defaultAddKey, path+keyFirst, func() {
			p.replace(func(list []*gurps.SkillDefault) []*gurps.SkillDefault { return append(list, def) })
			p.open = path
		})
		return true
	})
	add.RefKey = defaultAddKey
	for i, one := range *p.defaults {
		p.AddChild(p.row(one, strconv.Itoa(i)))
	}
	// New defaults go at the end, so the add button sits under the last row, in line with the more buttons.
	foot := unison.NewPanel()
	foot.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: 4, Right: 8}))
	empty := len(*p.defaults) == 0
	if empty {
		// With no defaults, a placeholder that adds one as the add button does stands before it.
		foot.AddChild(newEmptyPlaceholder(defaultEmptyKey, i18n.Text("No defaults. Click here to add one."),
			add.ClickCallback))
	}
	foot.AddChild(add)
	hbox(foot, unison.StdHSpacing)
	add.SetLayoutData(&unison.FlexLayoutData{HAlign: align.End, VAlign: align.Middle, HGrab: !empty})
	p.AddChild(foot)
}

// summary returns the paragraph a collapsed panel shows: the description of each row, joined by semicolons and ended
// with a period.
func (p *defaultsPanel) summary() string {
	if len(*p.defaults) == 0 {
		return i18n.Text("No defaults.")
	}
	replacements := p.replacements()
	descriptions := make([]string, 0, len(*p.defaults))
	for _, one := range *p.defaults {
		descriptions = append(descriptions, one.Describe(p.entity, replacements, emphasize))
	}
	return fmt.Sprintf(i18n.Text("%s."), strings.Join(descriptions, i18n.Text("; ")))
}

// row returns the panel for a default: its sentence, or while it is open its editor, beside a button for more actions.
func (p *defaultsPanel) row(def *gurps.SkillDefault, path string) *unison.Panel {
	row := p.sentenceRow(path, func() string { return def.Describe(p.entity, p.replacements(), emphasize) }, true,
		func() *unison.Panel { return p.editor(def, path) }, func() []menuEntry { return p.moreEntries(path) }, nil, nil)
	row.ClientData()[defaultDropKey] = path
	return row
}

// moreEntries returns the entries of the more menu of the default at the path: Duplicate, Move up and down, and Delete.
func (p *defaultsPanel) moreEntries(path string) []menuEntry {
	i := p.index(path)
	if i < 0 {
		return nil
	}
	from := path + keyMore
	entries := []menuEntry{{Label: i18n.Text("Duplicate"), Act: func() {
		p.restructure(i18n.Text("Duplicate Default"), from, "",
			func(list []*gurps.SkillDefault) ([]*gurps.SkillDefault, int) {
				c := *list[i]
				return slices.Insert(list, i+1, &c), i + 1
			})
	}}}
	titles := []string{i18n.Text("Move Up"), i18n.Text("Move Down")}
	for k, j := range []int{i - 1, i + 1} {
		if j < 0 || j >= len(*p.defaults) {
			continue
		}
		title := titles[k]
		entries = append(entries, menuEntry{Label: title, Act: func() {
			p.restructure(title, from, "", func(list []*gurps.SkillDefault) ([]*gurps.SkillDefault, int) {
				list[i], list[j] = list[j], list[i]
				return list, j
			})
		}})
	}
	return append(entries, menuEntry{}, menuEntry{Label: i18n.Text("Delete"), Act: func() {
		p.restructure(i18n.Text("Delete Default"), from, defaultAddKey,
			func(list []*gurps.SkillDefault) ([]*gurps.SkillDefault, int) {
				list = slices.Delete(list, i, i+1)
				if i < len(list) {
					return list, i
				}
				return list, -1
			})
	}})
}

// dropAt returns the row a default dragged to where would be dropped on and whether it would go before or after it, by
// which half of the row where is in. The row is nil where the default can't go, such as onto itself.
func (p *defaultsPanel) dropAt(where geom.Point, data any) (target *unison.Panel, at int) {
	from := p.dragPath(data)
	if p.index(from) < 0 {
		return nil, 0
	}
	for target = p.PanelAt(where); target != nil && target != p.AsPanel(); target = target.Parent() {
		path, isRow := target.ClientData()[defaultDropKey].(string)
		if !isRow {
			continue
		}
		if path == from {
			return nil, 0
		}
		if target.PointFromRoot(p.PointToRoot(where)).Y < target.FrameRect().Height/2 {
			return target, dropBefore
		}
		return target, dropAfter
	}
	return nil, 0
}

// drop moves the dragged default to where it would go.
func (p *defaultsPanel) drop(where geom.Point, data any) {
	target, at := p.dropAt(where, data)
	p.dragExit()
	if target == nil {
		return
	}
	path, isRow := target.ClientData()[defaultDropKey].(string)
	if !isRow {
		return
	}
	from := p.dragPath(data)
	i := p.index(from)
	index := p.index(path)
	if at == dropAfter {
		index++
	}
	if i < index {
		// Taking the default out moves what comes after it up by one.
		index--
	}
	if i == index {
		return
	}
	p.restructure(i18n.Text("Move Default"), from+keyMore, "",
		func(list []*gurps.SkillDefault) ([]*gurps.SkillDefault, int) {
			one := list[i]
			return slices.Insert(slices.Delete(list, i, i+1), index, one), index
		})
}

// restructure moves, adds or removes defaults through the widget with the reference key from, then rebuilds. change is
// handed a copy of the list to make the new one from, and returns it with the index of the default whose more button
// takes the focus; with -1, the widget with the reference key fallback does. The open row stays open wherever it ends
// up, and closes if it is removed.
func (p *defaultsPanel) restructure(title, from, fallback string, change func(list []*gurps.SkillDefault) ([]*gurps.SkillDefault, int)) {
	var open *gurps.SkillDefault
	if i := p.index(p.open); i >= 0 {
		open = (*p.defaults)[i]
	}
	p.sentenceRows.restructure(title, from, fallback, func() string {
		var dst int
		p.replace(func(list []*gurps.SkillDefault) []*gurps.SkillDefault {
			list, dst = change(list)
			return list
		})
		p.open = ""
		if i := slices.Index(*p.defaults, open); open != nil && i >= 0 {
			p.open = strconv.Itoa(i)
		}
		if dst < 0 {
			return ""
		}
		return strconv.Itoa(dst) + keyMore
	})
}

// defaultTypeFlags are the choices the type popup offers beside the attributes.
const defaultTypeFlags = gurps.TenFlag | gurps.ParryFlag | gurps.BlockFlag | gurps.SkillFlag

// defaultIndent is how far the lines of an open row after its first are indented, to show they belong to it.
const defaultIndent = 12

// editor returns the controls for an open row: its type, the criteria that pick the skill of a skill-based type and the
// modifier, flowing as a sentence would, then its optional criteria as chips. Every line after the first is indented a
// little.
func (p *defaultsPanel) editor(def *gurps.SkillDefault, path string) *unison.Panel {
	key := func(name string) string { return path + ":" + name }
	editor := newColumn()
	editor.RefKey = path + keyFirst
	fields := newHangingFlow(defaultIndent)
	editor.AddChild(fields)
	p.typePopup(fields, path, def)
	chips := newFlow()
	if def.SkillBased() {
		if def.Type() != gurps.SkillID {
			addJoiningWords(fields, i18n.Text("of skill"))
		}
		whose := i18n.Text("whose name")
		p.textCriteria(fields, key("name"), i18n.Text("Name"), "", whose, whose, &def.Name, true)
		p.textChip(chips, path, "specialization", &def.Specialization)
		p.tagsChip(chips, path, &def.Tags)
	}
	title := i18n.Text("Modifier")
	p.addCompact(fields, NewDecimalField(p.targetMgr, key("modifier"), title, func() fxp.Int { return def.Modifier },
		func(v fxp.Int) { p.edit(title, key("modifier"), "", func() { def.Modifier = v }) }, -fxp.Thousand,
		fxp.Thousand, true, false).withoutUndo())
	p.optional(chips, path, "tl", i18n.Text("+ tech level"), i18n.Text("Add Tech Level"), i18n.Text("Remove Tech Level"),
		def.WhenTL.Compare != criteria.AnyNumber,
		func() { def.WhenTL = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: p.techLevel()} },
		func() { def.WhenTL = criteria.Number{} },
		func(chip *unison.Panel) {
			p.numberCriteria(chip, key("tl"), i18n.Text("Tech Level"), i18n.Text("when the tech level"), &def.WhenTL, 0,
				fxp.Twelve, true)
		})
	// The buttons that add unused criteria go after the chips in use, so an added one takes its place among them.
	for _, child := range slices.Clone(chips.Children()) {
		if _, ok := child.Self.(*unison.Button); ok {
			chips.AddChild(child)
		}
	}
	chips.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: defaultIndent}))
	editor.AddChild(chips)
	return editor
}

// techLevel returns the whole tech level of the entity, which a new tech level criterion starts from, or 0 without one.
func (p *defaultsPanel) techLevel() fxp.Int {
	if p.entity == nil {
		return 0
	}
	tl, _, _ := gurps.ExtractTechLevel(p.entity.Profile.TechLevel)
	return tl.Trunc()
}

// typeChoices returns the types a default can be given: an attribute, 10, Parry, Block or Skill.
func (p *defaultsPanel) typeChoices() []*gurps.AttributeChoice {
	// Asked with a current type that is among them, so that none is added for it.
	choices, _ := gurps.AttributeChoices(p.entity, "", defaultTypeFlags, gurps.SkillID)
	return choices
}

// addType returns the type a new default takes: the one last chosen, unless the entity has no such attribute.
func (p *defaultsPanel) addType() string {
	for _, one := range p.typeChoices() {
		if one.Key == lastDefaultTypeUsed {
			return lastDefaultTypeUsed
		}
	}
	return gurps.DexterityID
}

// typePopup adds the popup that switches a default to another type: an attribute, 10, Parry, Block or Skill. A type that
// isn't one of them, which a file can still hold, is named as the row's sentence names it, and offered for as long as
// the panel is open, as are the others of its kind the defaults had when the panel was made.
func (p *defaultsPanel) typePopup(parent *unison.Panel, path string, def *gurps.SkillDefault) {
	choices := p.typeChoices()
	for _, t := range append(slices.Clone(p.otherTypes), def.Type()) {
		if !slices.ContainsFunc(choices, func(c *gurps.AttributeChoice) bool { return c.Key == t }) {
			choices = append(choices, &gurps.AttributeChoice{Key: t, Title: gurps.DefaultTypeTitle(p.entity, t)})
		}
	}
	var current *gurps.AttributeChoice
	for _, one := range choices {
		if one.Key == def.Type() {
			current = one
		}
	}
	addCentered(parent, compactPopup(&p.sentenceRows, path+":type", i18n.Text("Default Type"), choices, current,
		func(c *gurps.AttributeChoice) string { return c.Title }, func(c *gurps.AttributeChoice) {
			lastDefaultTypeUsed = c.Key
			wasSkillBased := def.SkillBased()
			def.DefaultType = c.Key
			// The skill criteria go with the skill-based types, and what they held would otherwise linger unseen on any
			// other, to come back should the default later be made skill-based again. They are dropped within the edit,
			// so that a default set back to what it was leaves the editor unmodified.
			def.Normalize()
			// One that becomes skill-based names its skill, so that the field for the name shows.
			if !wasSkillBased && def.SkillBased() && def.Name.IsZero() {
				def.Name = criteria.Text{Compare: criteria.IsText}
			}
		}))
}
