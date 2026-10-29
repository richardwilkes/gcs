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
	"maps"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/drag"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/zeebo/xxh3"
)

const (
	nameablesAddKey = "nameables:add"
	// nameablesCodeKey is the code field of the open row, the same whatever its code, so that a rename keeps its focus.
	nameablesCodeKey = "nameables:code"
	// nameableOptionKey marks the chip of an option, holding its index.
	nameableOptionKey = "nameable.option"
)

// nameableOptionDragKey is the drag data type of an option being dragged to another place among its nameable's.
var nameableOptionDragKey = unison.CreatePrivateDataType("gcs.nameable.option")

// nameableModifier is one of an item's modifiers, as the nameables panel shows it.
type nameableModifier struct {
	id       tid.TID
	location string
	name     string
	disabled bool
	defs     map[string]nameable.Detail
	uses     map[string]int
	// granted holds the definitions of the items the modifier's grants give (see grantedNameableDefinitions).
	granted map[string]nameable.Detail
}

// nameableOwner is the trait or piece of equipment a modifier belongs to, as the nameables panel of the modifier's
// editor shows it: the prompt for the modifier takes definitions from its parts as it does for the owner, the modifier
// taking its place among the owner's modifiers (see nameableDefinitions).
type nameableOwner struct {
	defs map[string]nameable.Detail
	// granted holds the definitions of the items the owner's grants give (see grantedNameableDefinitions).
	granted map[string]nameable.Detail
	// links map the owner's modifiers' codes to the owner's codes that answer them (see gurps.Trait.NameableLinks).
	links     map[tid.TID]map[string]string
	modifiers []nameableModifier
	// self is the ID of the modifier being edited.
	self tid.TID
}

// nameablesState is what the panel edits: the item's nameables and, for an item with modifiers, its links for them.
type nameablesState struct {
	defs  map[string]nameable.Detail
	links map[tid.TID]map[string]string
}

func (s *nameablesState) clone() *nameablesState {
	other := &nameablesState{defs: nameable.CloneDetails(s.defs)}
	if s.links != nil {
		other.links = make(map[tid.TID]map[string]string, len(s.links))
		for id, one := range s.links {
			other.links[id] = maps.Clone(one)
		}
	}
	return other
}

// nameablesPanel edits the nameables an item's text and features refer to as "@code@". Each one is a row that reads as
// a sentence until it is opened to edit it. A code that is referred to but has no nameable gets a row that offers to
// define it, unless its marker, an enabled modifier or an item a grant gives defines it, or, for a modifier, the trait
// or equipment it belongs to does, as the prompt then asks for it (see nameablesPanel.origins). For an item with
// modifiers, each code its modifiers refer to gets a row of its own, which says whether it shares the item's answer,
// takes the answer to one of the item's codes, or is asked for separately. Clicking the title collapses the panel to a
// paragraph of the item's rows. It starts out collapsed when there are nameables.
type nameablesPanel struct {
	sentenceRows[*nameablesState]
	defs      *map[string]nameable.Detail
	links     *map[tid.TID]map[string]string
	uses      func() map[string]int
	granted   func() map[string]nameable.Detail
	modifiers func() []nameableModifier
	owner     func() *nameableOwner
	signature uint64
	// pending is the key of a chip added to the open row that holds nothing yet, such as a new option still being
	// typed. It shows until the row closes, since nothing in the data says it is there.
	pending   string
	dropChip  *unison.Panel
	dropAfter bool
}

// newNameablesPanel returns the nameables section of an editor. uses returns how often the item refers to each code.
// granted returns the definitions of the items the item's grants give (see grantedNameableDefinitions), and is nil
// for an item that has no features. links and modifiers are nil for an item that has no modifiers. owner returns the
// trait or equipment the modifier being edited belongs to, or nil when it belongs to none, and is nil for an item that
// is not a modifier. defs and links point at maps rather than being them, since undo replaces the maps themselves.
//
//nolint:gocritic // See above
func newNameablesPanel(defs *map[string]nameable.Detail, uses func() map[string]int,
	granted func() map[string]nameable.Detail, links *map[tid.TID]map[string]string,
	modifiers func() []nameableModifier, owner func() *nameableOwner,
) *nameablesPanel {
	p := &nameablesPanel{
		defs:      defs,
		links:     links,
		uses:      uses,
		granted:   granted,
		modifiers: modifiers,
		owner:     owner,
	}
	border := initTitledEditorSection(p, i18n.Text("Nameables"))
	p.initRows(nil, p.build, p.state, p.restore, p.dataHash)
	p.initCollapse(border, len(*defs) != 0)
	installPanelDragDrop(p.AsPanel(), nameableOptionDragKey, p.optionDragOver, p.optionDragExit, p.optionDrop)
	p.DrawOverCallback = p.drawOptionDrop
	p.build()
	return p
}

func (p *nameablesPanel) current() *nameablesState {
	s := &nameablesState{defs: *p.defs}
	if p.links != nil {
		s.links = *p.links
	}
	return s
}

func (p *nameablesPanel) state() *nameablesState {
	return p.current().clone()
}

func (p *nameablesPanel) restore(s *nameablesState) {
	s = s.clone()
	*p.defs = s.defs
	if p.links != nil {
		*p.links = s.links
	}
}

func (p *nameablesPanel) dataHash() uint64 {
	return p.current().hash()
}

// hash returns a hash of the state, by which an edit is told from no change. Each string goes in with its length, and
// each map and list with its count, so that no two states run together the same way, as the text of options whose
// words repeat would.
func (s *nameablesState) hash() uint64 {
	h := xxh3.New()
	xhash.Num64(h, uint64(len(s.defs)))
	for _, code := range slices.Sorted(maps.Keys(s.defs)) {
		d := s.defs[code]
		xhash.StringWithLen(h, code)
		xhash.StringWithLen(h, d.Label)
		xhash.StringWithLen(h, d.Tooltip)
		xhash.Num64(h, uint64(len(d.Options)))
		for _, option := range d.Options {
			xhash.StringWithLen(h, option)
		}
		xhash.Bool(h, d.FreeForm)
		xhash.Bool(h, d.NotBlank)
	}
	xhash.Num64(h, uint64(len(s.links)))
	for _, id := range slices.Sorted(maps.Keys(s.links)) {
		xhash.StringWithLen(h, string(id))
		links := s.links[id]
		xhash.Num64(h, uint64(len(links)))
		for _, code := range slices.Sorted(maps.Keys(links)) {
			xhash.StringWithLen(h, code)
			xhash.StringWithLen(h, links[code])
		}
	}
	return h.Sum64()
}

// Sync implements Syncer. The rows show how often the item refers to each code, what the items its grants give define,
// what its modifiers are and what the item a modifier belongs to is, so they are rebuilt whenever any of these changes.
func (p *nameablesPanel) Sync() {
	uses, granted, mods, owner := p.inputs()
	if signature(uses, granted, mods, owner) != p.signature {
		p.rebuild("")
	}
}

// inputs returns what the rows show besides the panel's own data: how often the item refers to each code, the
// definitions of the items its grants give, its modifiers, and the item a modifier belongs to.
func (p *nameablesPanel) inputs() (uses map[string]int, granted map[string]nameable.Detail, mods []nameableModifier,
	owner *nameableOwner,
) {
	if p.modifiers != nil {
		mods = p.modifiers()
	}
	if p.owner != nil {
		owner = p.owner()
	}
	return p.uses(), p.grantedDefinitions(), mods, owner
}

// grantedDefinitions returns the definitions of the items the item's grants give, or nil when it has no features.
func (p *nameablesPanel) grantedDefinitions() map[string]nameable.Detail {
	if p.granted == nil {
		return nil
	}
	return p.granted()
}

func signature(uses map[string]int, granted map[string]nameable.Detail, mods []nameableModifier,
	owner *nameableOwner,
) uint64 {
	// fmt writes maps in key order and the fields of the modifiers and of the owner too, so the text is the same for
	// the same rows.
	return xxh3.HashString(fmt.Sprintf("%v|%v|%v|%v", uses, granted, mods, owner))
}

// codes returns the item's codes, in the order their rows are shown.
func (p *nameablesPanel) codes() []string {
	return slices.Sorted(maps.Keys(*p.defs))
}

// pathOf returns the path of the row for the code's nameable.
func pathOf(code string) string {
	return "n:" + code
}

// markerText returns the code as it appears in text.
func markerText(code string) string {
	return string(nameable.MarkerDelimiter) + code + string(nameable.MarkerDelimiter)
}

func (p *nameablesPanel) build() {
	uses, granted, mods, owner := p.inputs()
	p.signature = signature(uses, granted, mods, owner)
	origins := p.origins(granted, mods, owner)
	p.staticTip = func(path string) string { return p.staticTipOf(path, uses, origins) }
	if _, ok := (*p.defs)[strings.TrimPrefix(p.open, "n:")]; !ok && !strings.HasPrefix(p.open, "m:") {
		p.open = ""
	}
	if p.open == "" || !strings.HasPrefix(p.pending, p.open+":") {
		p.pending = ""
	}
	// Collapsed, the nameables read as one paragraph.
	if p.addTitleBar(p.summary) != nil {
		return
	}
	byModifiers := p.modifierUses(uses, mods)
	defs := *p.defs
	for _, code := range p.codes() {
		detail := defs[code]
		p.AddChild(p.row(code, &detail, uses[code], byModifiers[code]))
	}
	for _, code := range slices.Sorted(maps.Keys(uses)) {
		if _, ok := defs[code]; !ok {
			p.AddChild(p.undefinedRow(code, uses[code], byModifiers[code], origins[code]))
		}
	}
	p.addModifierRows(uses, mods)
	add := newSectionAddButton(p, i18n.Text("Add a nameable"), func() bool {
		p.add()
		return false
	})
	add.RefKey = nameablesAddKey
	// The add button sits under the last row, in line with the more buttons, or beside the placeholder of an empty
	// section, as the features section's does.
	foot := unison.NewPanel()
	foot.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: 4, Right: 8}))
	if len(p.Children()) == 1 {
		// With nothing but the title bar, a placeholder that adds a nameable as the add button does stands before it.
		foot.AddChild(newEmptyPlaceholder("nameables:empty", i18n.Text("No nameables. Click here to add one."), p.add))
	}
	foot.AddChild(add)
	hbox(foot, unison.StdHSpacing)
	add.SetLayoutData(&unison.FlexLayoutData{HAlign: align.End, VAlign: align.Middle, HGrab: len(foot.Children()) == 1})
	p.AddChild(foot)
}

// summary returns the paragraph a collapsed panel shows: the sentence of each nameable, then of each code that is used
// but not defined, each ended with a period.
func (p *nameablesPanel) summary() string {
	uses, granted, mods, owner := p.inputs()
	byModifiers := p.modifierUses(uses, mods)
	origins := p.origins(granted, mods, owner)
	defs := *p.defs
	var sentences []string
	for _, code := range p.codes() {
		detail := defs[code]
		sentences = append(sentences, i18n.Text("%s.", usesText(detail.Describe(code, emphasize), uses[code],
			byModifiers[code])))
	}
	for _, code := range slices.Sorted(maps.Keys(uses)) {
		if _, ok := defs[code]; !ok {
			sentences = append(sentences, i18n.Text("%s.", undefinedText(code, uses[code], byModifiers[code],
				origins[code])))
		}
	}
	if len(sentences) == 0 {
		return i18n.Text("No nameables.")
	}
	return strings.Join(sentences, " ")
}

// modifierUses returns how often the enabled modifiers, whose keys alone are asked about, refer to each code the item
// has (see has), counting both the codes they share with the item and those linked to one of its codes. uses is how
// often the item refers to each code.
func (p *nameablesPanel) modifierUses(uses map[string]int, mods []nameableModifier) map[string]int {
	counts := make(map[string]int)
	for _, one := range mods {
		if one.disabled {
			continue
		}
		for code, n := range one.uses {
			if target := p.linkOf(one.id, code); target != "" {
				code = target
			}
			if p.has(uses, code) {
				counts[code] += n
			}
		}
	}
	return counts
}

// has reports whether the item has the code, defining it or referring to it, uses being how often it refers to each.
// The model treats both alike, keeping the answer in the item's replacements under the code.
func (p *nameablesPanel) has(uses map[string]int, code string) bool {
	if _, defined := (*p.defs)[code]; defined {
		return true
	}
	return uses[code] != 0
}

// nameableOrigin is where the prompt takes the definition of a code the item doesn't define from. The zero value is for
// a code nothing defines.
type nameableOrigin struct {
	// modifier is the enabled modifier that defines the code, itself or by one of its grants, or nil when something
	// else does.
	modifier *nameableModifier
	// owner is whether the item the modifier belongs to defines the code, itself or by one of its grants.
	owner bool
	// granted is whether the definition is that of an item a grant gives.
	granted bool
}

// defined reports whether something defines the code.
func (o nameableOrigin) defined() bool {
	return o.modifier != nil || o.owner || o.granted
}

// origins returns where the prompt takes the definition of each code the item doesn't define from, taking them as
// nameableDefinitions does for the item: from its enabled modifiers, then from the items its grants, and theirs, give.
// For a modifier, they are taken as for the item it belongs to, the modifier among that item's modifiers, and keyed by
// the modifier's own codes. granted holds the definitions of the items the item's grants give.
func (p *nameablesPanel) origins(granted map[string]nameable.Detail, mods []nameableModifier,
	owner *nameableOwner,
) map[string]nameableOrigin {
	item := nameableDefinitionSource{defs: *p.defs, granted: granted}
	// parts holds what each source is, as the origin of a definition it holds itself; the item's is the zero value.
	sources := []nameableDefinitionSource{item}
	parts := []nameableOrigin{{}}
	var links map[tid.TID]map[string]string
	if p.links != nil {
		links = *p.links
	}
	placed := true
	if owner != nil {
		item.links = owner.links[owner.self]
		sources = []nameableDefinitionSource{{defs: owner.defs, granted: owner.granted}}
		parts = []nameableOrigin{{owner: true}}
		mods, links, placed = owner.modifiers, owner.links, false
	}
	for i := range mods {
		one := &mods[i]
		switch {
		case owner != nil && one.id == owner.self:
			if !placed {
				sources = append(sources, item)
				parts = append(parts, nameableOrigin{})
				placed = true
			}
		case !one.disabled:
			sources = append(sources, nameableDefinitionSource{
				defs:    one.defs,
				granted: one.granted,
				links:   links[one.id],
			})
			parts = append(parts, nameableOrigin{modifier: one})
		}
	}
	if !placed {
		sources = append(sources, item)
		parts = append(parts, nameableOrigin{})
	}
	origins := make(map[string]nameableOrigin)
	for code, one := range mergeNameableDefinitions(sources) {
		if origin := parts[one.source]; origin.defined() || one.granted {
			origin.granted = one.granted
			origins[code] = origin
		}
	}
	if len(item.links) == 0 {
		return origins
	}
	// The merged definitions are keyed by the owner's codes, which the modifier's linked codes are asked about as.
	keyed := make(map[string]nameableOrigin, len(origins))
	for code, origin := range origins {
		if _, linked := item.links[code]; !linked {
			keyed[code] = origin
		}
	}
	for code, target := range item.links {
		if origin, ok := origins[target]; ok {
			keyed[code] = origin
		}
	}
	return keyed
}

// linkOf returns the item's code the modifier's code is linked to, or "" when it has no link.
func (p *nameablesPanel) linkOf(id tid.TID, code string) string {
	return (*p.links)[id][code]
}

// shares reports whether the modifier's code shares the item's answer: the item has the code too (see has), as the
// modifier's unlinked code is answered from the item's replacements under the code, and the modifier's is not linked
// to another of the item's codes, which the model would apply first (see gurps.linkedReplacements). uses is how often
// the item refers to each code.
func (p *nameablesPanel) shares(uses map[string]int, id tid.TID, code string) bool {
	if !p.has(uses, code) {
		return false
	}
	target := p.linkOf(id, code)
	return target == "" || target == code
}

// usesText returns the sentence with how often its code is referred to added: by its owner, and by the modifiers that
// share or are linked to it.
func usesText(sentence string, uses, byModifiers int) string {
	switch {
	case byModifiers == 0 && uses == 0:
		return i18n.Text("%s; not used yet", sentence)
	case byModifiers == 0 && uses == 1:
		return i18n.Text("%s; used once", sentence)
	case byModifiers == 0:
		return i18n.Text("%s; used %d times", sentence, uses)
	case byModifiers == 1 && uses == 0:
		return i18n.Text("%s; used once in a modifier", sentence)
	case uses == 0:
		return i18n.Text("%s; used %d times in modifiers", sentence, byModifiers)
	case byModifiers == 1 && uses == 1:
		return i18n.Text("%s; used once, and once in a modifier", sentence)
	case byModifiers == 1:
		return i18n.Text("%s; used %d times, and once in a modifier", sentence, uses)
	case uses == 1:
		return i18n.Text("%s; used once, and %d times in modifiers", sentence, byModifiers)
	default:
		return i18n.Text("%s; used %d times, and %d times in modifiers", sentence, uses, byModifiers)
	}
}

// row returns the panel for a nameable: its sentence, or while it is open its editor, beside a button for more actions.
func (p *nameablesPanel) row(code string, detail *nameable.Detail, uses, byModifiers int) *unison.Panel {
	path := pathOf(code)
	return p.sentenceRow(path, func() string { return usesText(detail.Describe(code, emphasize), uses, byModifiers) },
		true, func() *unison.Panel { return p.editor(path, code, detail) },
		func() []menuEntry {
			return []menuEntry{{Label: i18n.Text("Delete"), Act: func() {
				p.restructure(i18n.Text("Delete Nameable"), path+keyMore, nameablesAddKey, func() string {
					delete(*p.defs, code)
					p.unlink(code)
					if p.open == path {
						p.open = ""
					}
					return ""
				})
			}}}
		}, nil, nil)
}

// unlink removes the links to the item's code.
func (p *nameablesPanel) unlink(code string) {
	if p.links == nil {
		return
	}
	for id, one := range *p.links {
		for local, target := range one {
			if target == code {
				delete(one, local)
			}
		}
		if len(one) == 0 {
			delete(*p.links, id)
		}
	}
}

// inlineMarker reports whether the code the item's text uses is a marker that holds its own definition, written in the
// older forms, such as "@Style: Judo, Karate, etc.@", rather than a bare code.
func inlineMarker(code string) bool {
	m, ok := nameable.NewMarker(code)
	return ok && m.Key() != m.Label
}

// undefinedText returns the sentence for a code the item uses that has no nameable, with how often the item and the
// modifiers that share it refer to it, and where the prompt takes its definition from instead.
func undefinedText(code string, uses, byModifiers int, origin nameableOrigin) string {
	marker := emphasize(markerText(code))
	var text string
	switch {
	case origin.modifier != nil && origin.granted:
		text = i18n.Text("%s is defined by an item the modifier %s grants", marker,
			emphasize(origin.modifier.ownerText()))
	case origin.modifier != nil:
		text = i18n.Text("%s is defined by the modifier %s", marker, emphasize(origin.modifier.ownerText()))
	case origin.owner && origin.granted:
		text = i18n.Text("%s is defined by an item granted by the item the modifier belongs to", marker)
	case origin.owner:
		text = i18n.Text("%s is defined by the item the modifier belongs to", marker)
	case origin.granted:
		text = i18n.Text("%s is defined by a granted item", marker)
	case inlineMarker(code):
		text = i18n.Text("%s defines itself", marker)
	default:
		text = i18n.Text("%s isn't defined", marker)
	}
	return usesText(text, uses, byModifiers)
}

// undefinedRow returns the panel for a code the item refers to that has no nameable, with how often the item and the
// modifiers that share it refer to it, and where the prompt takes its definition from instead. A bare code that nothing
// else defines gets a button that defines one.
func (p *nameablesPanel) undefinedRow(code string, uses, byModifiers int, origin nameableOrigin) *unison.Panel {
	var trail func(row *unison.Panel, sentence *sentenceButton)
	if !origin.defined() && !inlineMarker(code) {
		trail = func(row *unison.Panel, _ *sentenceButton) {
			key := "u:" + code + ":define"
			define := newDashedButton(i18n.Text("Define"), func() {
				p.edit(i18n.Text("Define Nameable"), key, pathOf(code)+keyFirst, func() {
					if *p.defs == nil {
						*p.defs = make(map[string]nameable.Detail)
					}
					(*p.defs)[code] = nameable.Detail{Label: code, FreeForm: true}
					p.open = pathOf(code)
				})
			})
			define.RefKey = key
			define.Accessibility.Name = i18n.Text("Define %s", markerText(code))
			define.OnBackgroundInk = unison.ThemeAlert
			fitLine(define)
			addCentered(row, define)
		}
	}
	return p.sentenceRow("u:"+code, func() string { return undefinedText(code, uses, byModifiers, origin) }, false,
		nil, nil, nil, trail)
}

// staticTipOf returns the tooltip of the row at the path, which is static: that of a code the item refers to that has
// no nameable, or that of a modifier's code that shares the item's answer or has none of the item's codes to take the
// answer to. uses is how often the item refers to each code, and origins where the prompt takes the definition of each
// code the item doesn't define from (see nameablesPanel.origins).
func (p *nameablesPanel) staticTipOf(path string, uses map[string]int, origins map[string]nameableOrigin) string {
	if code, undefined := strings.CutPrefix(path, "u:"); undefined {
		switch origin := origins[code]; {
		case origin.granted:
			return i18n.Text("An item a grant gives defines this, so the player is asked for it as that item asks.")
		case origin.modifier != nil:
			return i18n.Text("A modifier defines this, so the player is asked for it as that modifier asks.")
		case origin.owner:
			return i18n.Text("The item the modifier belongs to defines this, so the player is asked for it as that item asks.")
		case inlineMarker(code):
			return i18n.Text("This marker holds its own definition, written in the text.")
		default:
			return i18n.Text("Until it is defined, the player is asked for it as any text.")
		}
	}
	id, code, _ := strings.Cut(strings.TrimPrefix(path, "m:"), ":")
	if p.shares(uses, tid.TID(id), code) {
		return i18n.Text("The item has this code too, so this always shares the item's answer.")
	}
	return i18n.Text("The item has no nameables of its own for this to take the answer to.")
}

// add adds a nameable with a fresh code and opens it.
func (p *nameablesPanel) add() {
	code := "new"
	for i := 2; ; i++ {
		if _, exists := (*p.defs)[code]; !exists {
			break
		}
		code = fmt.Sprintf("new%d", i)
	}
	p.edit(i18n.Text("Add Nameable"), nameablesAddKey, nameablesCodeKey, func() {
		if *p.defs == nil {
			*p.defs = make(map[string]nameable.Detail)
		}
		(*p.defs)[code] = nameable.Detail{FreeForm: true}
		p.open = pathOf(code)
	})
}

// update replaces the nameable with the code with the result of change, recorded as an edit.
func (p *nameablesPanel) update(code, title, key string, change func(d *nameable.Detail)) {
	p.edit(title, key, "", func() { p.change(code, change) })
}

// change replaces the nameable with the code with the result of change, within an edit already being recorded.
func (p *nameablesPanel) change(code string, change func(d *nameable.Detail)) {
	detail := (*p.defs)[code]
	detail.Options = slices.Clone(detail.Options)
	change(&detail)
	(*p.defs)[code] = detail
}

// editor returns the controls of an open row: its code and name, flowing as a sentence, then its options and the
// rest as chips. detail is the row's copy of the nameable, which the fields read. Typing changes the nameable without
// rebuilding the row, so the fields' setters change the copy too: a sync of a field whose copy had gone stale would
// hand its old text back to the setter, undoing what was typed into it.
func (p *nameablesPanel) editor(path, code string, detail *nameable.Detail) *unison.Panel {
	box := newColumn()
	box.RefKey = path + keyFirst
	key := func(name string) string { return path + ":" + name }
	set := func(title, fieldKey string, change func(d *nameable.Detail)) {
		change(detail)
		p.update(code, title, fieldKey, change)
	}

	// The first row is a sentence of three fixed parts, so it is a flex row rather than a flow, which lets the name take
	// the width left after the code and the words between them.
	fields := unison.NewPanel()
	box.AddChild(fields)
	codeField := NewStringField(p.targetMgr, nameablesCodeKey, i18n.Text("Code"), func() string { return code },
		func(s string) {
			// The code only changes once it is valid.
			if s == code || !validNameableCode(s) {
				return
			}
			if _, exists := (*p.defs)[s]; exists {
				return
			}
			old := code
			p.edit(i18n.Text("Code"), nameablesCodeKey, "", func() {
				(*p.defs)[s] = (*p.defs)[old]
				delete(*p.defs, old)
				if p.links != nil {
					for _, one := range *p.links {
						for local, target := range one {
							if target == old {
								one[local] = s
							}
						}
					}
				}
				p.open = pathOf(s)
			})
			code = s
			// The row's keys are made from its code, so it is rebuilt, keeping the focus in this field. A rebuild ends
			// a run of typing, but the rename carries on as one edit.
			p.rebuild("")
			p.editKey = nameablesCodeKey
		})
	codeField.Font = unison.MonospacedFont
	codeField.Watermark = i18n.Text("code")
	codeField.SetMinimumTextWidthUsing("MMMMMMMM")
	codeField.ValidateCallback = func() bool {
		s := codeField.Text()
		if s == code {
			return true
		}
		_, exists := (*p.defs)[s]
		return validNameableCode(s) && !exists
	}
	p.addCompact(fields, codeField.withoutUndo())
	addJoiningWords(fields, i18n.Text("asks for"))
	name := NewStringField(p.targetMgr, key("name"), i18n.Text("Name"), func() string { return detail.Label },
		func(s string) { set(i18n.Text("Name"), key("name"), func(d *nameable.Detail) { d.Label = s }) })
	name.Watermark = i18n.Text("what the player is asked for")
	name.SetMinimumTextWidthUsing(i18n.Text("Martial arts style"))
	p.addCompact(fields, name.withoutUndo())
	hbox(fields, 6)
	name.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, VAlign: align.Middle, HGrab: true})

	chips := newFlow()
	for i, option := range detail.Options {
		p.chip(chips, fmt.Sprintf("%s:option%d", path, i), i18n.Text("Remove %s", option),
			key("add option"),
			func() { p.change(code, func(d *nameable.Detail) { d.Options = slices.Delete(d.Options, i, i+1) }) },
			func(chip *unison.Panel) {
				addJoiningWords(chip, option)
				chip.ClientData()[nameableOptionKey] = i
				p.dragOption(chip, code, i)
			})
	}
	// A new option is typed into a chip of its own, which holds nothing until Return adds what it holds.
	p.optional(chips, path, "option", i18n.Text("+ option"), i18n.Text("Add Option"), i18n.Text("Remove Option"),
		p.pending == key("option"), func() { p.pending = key("option") }, func() { p.pending = "" },
		func(chip *unison.Panel) { p.newOptionField(chip, path, code) })
	p.optional(chips, path, "free", i18n.Text("+ any text"), i18n.Text("Allow Any Text"),
		i18n.Text("Disallow Any Text"), detail.FreeForm,
		func() { p.change(code, func(d *nameable.Detail) { d.FreeForm = true }) },
		func() { p.change(code, func(d *nameable.Detail) { d.FreeForm = false }) },
		func(chip *unison.Panel) { addJoiningWords(chip, i18n.Text("or any text")) })
	p.optional(chips, path, "notblank", i18n.Text("+ not blank"), i18n.Text("Disallow a Blank Answer"),
		i18n.Text("Allow a Blank Answer"), detail.NotBlank,
		func() { p.change(code, func(d *nameable.Detail) { d.NotBlank = true }) },
		func() { p.change(code, func(d *nameable.Detail) { d.NotBlank = false }) },
		func(chip *unison.Panel) { addJoiningWords(chip, i18n.Text("which can't be blank")) })
	p.optional(chips, path, "desc", i18n.Text("+ description"), i18n.Text("Add Description"),
		i18n.Text("Remove Description"), detail.Tooltip != "" || p.pending == key("desc"),
		func() { p.pending = key("desc") },
		func() {
			p.pending = ""
			p.change(code, func(d *nameable.Detail) { d.Tooltip = "" })
		},
		func(chip *unison.Panel) {
			addJoiningWords(chip, i18n.Text("described as"))
			field := NewStringField(p.targetMgr, key("descfield"), i18n.Text("Description"),
				func() string { return detail.Tooltip },
				func(s string) {
					set(i18n.Text("Description"), key("descfield"), func(d *nameable.Detail) { d.Tooltip = s })
				})
			field.Watermark = i18n.Text("help shown with the question")
			field.SetMinimumTextWidthUsing(i18n.Text("Which style does this trait represent?"))
			p.addCompact(chip, field.withoutUndo())
		})
	// The buttons that add what isn't in use go after the chips in use, so an added one takes its place among them.
	for _, child := range slices.Clone(chips.Children()) {
		if _, ok := child.Self.(*unison.Button); ok {
			chips.AddChild(child)
		}
	}
	box.AddChild(chips)
	return box
}

// newOptionField adds the field that adds an option to the nameable with the code each time Return is pressed in it.
func (p *nameablesPanel) newOptionField(chip *unison.Panel, path, code string) {
	key := path + ":newoption"
	field := NewStringField(p.targetMgr, key, i18n.Text("Option"), func() string { return "" }, func(string) {})
	field.Watermark = i18n.Text("an option, then Return")
	field.SetMinimumTextWidthUsing(field.Watermark, i18n.Text("Martial arts style"))
	field.KeyDownCallback = func(keyCode unison.KeyCode, mods mod.Modifiers, repeat bool) bool {
		if keyCode != unison.KeyReturn && keyCode != unison.KeyNumPadEnter {
			return field.DefaultKeyDown(keyCode, mods, repeat)
		}
		option := strings.TrimSpace(field.Text())
		if option != "" && !slices.Contains((*p.defs)[code].Options, option) {
			p.update(code, i18n.Text("Add Option"), "", func(d *nameable.Detail) { d.Options = append(d.Options, option) })
			p.rebuild(key)
		}
		return true
	}
	p.addCompact(chip, field.withoutUndo())
}

// validNameableCode reports whether s may be used as a nameable's code: letters, digits and underscores only.
func validNameableCode(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		switch {
		case r == '_', r >= '0' && r <= '9', r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
		default:
			return false
		}
	}
	return true
}

// nameableOptionDrag is the payload of an option being dragged: the panel it belongs to, its nameable and its index.
type nameableOptionDrag struct {
	panel *nameablesPanel
	code  string
	index int
}

// dragOption lets a press on an option's chip that moves far enough to be a drag start dragging the option.
func (p *nameablesPanel) dragOption(chip *unison.Panel, code string, index int) {
	var start geom.Point
	var dragged bool
	chip.MouseDownCallback = func(where geom.Point, _, _ int, _ mod.Modifiers) bool {
		start = where
		dragged = false
		return true
	}
	chip.MouseDragCallback = func(where geom.Point, button int, _ mod.Modifiers) bool {
		_, drift := unison.DragGestureParameters()
		if dragged || button != unison.ButtonLeft ||
			(xmath.Abs(where.X-start.X) <= drift && xmath.Abs(where.Y-start.Y) <= drift) {
			return true
		}
		dragged = true
		startPanelDrag(chip, nameableOptionDragKey, &nameableOptionDrag{panel: p, code: code, index: index},
			chip.FrameRect().Size, geom.Point{}, chip.Draw, p.MarkForRedraw)
		return true
	}
}

// optionDropAt returns the chip of the option a dragged one would be dropped beside, and whether it would go after it.
func (p *nameablesPanel) optionDropAt(where geom.Point, data any) (chip *unison.Panel, after bool) {
	if d, ok := data.(*nameableOptionDrag); !ok || d.panel != p {
		return nil, false
	}
	for chip = p.PanelAt(where); chip != nil && chip != p.AsPanel(); chip = chip.Parent() {
		if _, isOption := chip.ClientData()[nameableOptionKey].(int); isOption {
			return chip, chip.PointFromRoot(p.PointToRoot(where)).X >= chip.FrameRect().Width/2
		}
	}
	return nil, false
}

func (p *nameablesPanel) optionDragOver(where geom.Point, data any) drag.Op {
	p.dropChip, p.dropAfter = p.optionDropAt(where, data)
	p.MarkForRedraw()
	if p.dropChip != nil {
		return drag.Move
	}
	return drag.None
}

func (p *nameablesPanel) optionDragExit() {
	p.dropChip = nil
	p.MarkForRedraw()
}

func (p *nameablesPanel) optionDrop(where geom.Point, data any) {
	chip, after := p.optionDropAt(where, data)
	p.optionDragExit()
	d, ok := data.(*nameableOptionDrag)
	if chip == nil || !ok {
		return
	}
	at, _ := chip.ClientData()[nameableOptionKey].(int) //nolint:errcheck // optionDropAt only returns option chips
	if after {
		at++
	}
	if at > d.index {
		at--
	}
	if at == d.index {
		return
	}
	p.update(d.code, i18n.Text("Reorder Options"), "", func(det *nameable.Detail) {
		option := det.Options[d.index]
		det.Options = slices.Insert(slices.Delete(det.Options, d.index, d.index+1), at, option)
	})
	p.rebuild("")
}

// drawOptionDrop marks where a dragged option would go, in the ink of GCS's other drop markers: a line before or after
// the chip it would go beside.
func (p *nameablesPanel) drawOptionDrop(gc *unison.Canvas, _ geom.Rect) {
	if p.dropChip == nil {
		return
	}
	r := p.RectFromRoot(p.dropChip.RectToRoot(p.dropChip.ContentRect(true)))
	x := r.X - 2
	if p.dropAfter {
		x = r.Right() + 2
	}
	paint := unison.ThemeWarning.Paint(gc, r, paintstyle.Stroke)
	paint.SetStrokeWidth(2)
	gc.DrawLine(geom.NewPoint(x, r.Y), geom.NewPoint(x, r.Bottom()), paint)
}

// addModifierRows adds a row for each code a modifier refers to or defines. uses is how often the item refers to each
// code.
func (p *nameablesPanel) addModifierRows(uses map[string]int, mods []nameableModifier) {
	defs := *p.defs
	for _, one := range mods {
		codes := slices.AppendSeq(slices.Collect(maps.Keys(one.defs)), maps.Keys(one.uses))
		slices.Sort(codes)
		for _, code := range slices.Compact(codes) {
			path := "m:" + string(one.id) + ":" + code
			shared := p.shares(uses, one.id, code)
			has := p.has(uses, code)
			p.AddChild(p.sentenceRow(path, func() string { return p.describeModifierCode(&one, code, shared) },
				!shared && len(defs) != 0, func() *unison.Panel { return p.linkEditor(path, &one, code, has) }, nil,
				nil, nil))
		}
	}
}

// ownerText returns how the rows for a modifier's codes name it.
func (one *nameableModifier) ownerText() string {
	switch {
	case one.location != "" && one.disabled:
		return i18n.Text("%s (%s, disabled)", one.name, one.location)
	case one.location != "":
		return i18n.Text("%s (%s)", one.name, one.location)
	case one.disabled:
		return i18n.Text("%s (disabled)", one.name)
	default:
		return one.name
	}
}

// describeModifierCode returns the sentence of the row for a code the modifier refers to or defines. The modifier
// counts as defining a code an item its grants give defines, or whose marker holds its own definition.
func (p *nameablesPanel) describeModifierCode(one *nameableModifier, code string, shared bool) string {
	marker := emphasize(markerText(code))
	owner := emphasize(one.ownerText())
	_, defined := one.defs[code]
	if _, byGrant := one.granted[code]; byGrant || inlineMarker(code) {
		defined = true
	}
	var text string
	switch target := (*p.links)[one.id][code]; {
	case shared && defined:
		text = i18n.Text("%s on %s shares the item's answer", marker, owner)
	case shared:
		text = i18n.Text("%s on %s shares the item's answer, though the modifier doesn't define it", marker, owner)
	case target != "" && defined:
		text = i18n.Text("%s on %s takes the answer to %s", marker, owner, emphasize(markerText(target)))
	case target != "":
		text = i18n.Text("%s on %s takes the answer to %s, though the modifier doesn't define it", marker, owner,
			emphasize(markerText(target)))
	case defined:
		text = i18n.Text("%s on %s is asked for under its own code", marker, owner)
	default:
		text = i18n.Text("%s on %s is asked for under its own code, though the modifier doesn't define it", marker,
			owner)
	}
	return usesText(text, one.uses[code], 0)
}

// linkEditor returns the controls of an open row for a modifier's code: the popup that links it to one of the item's
// codes, or leaves it unlinked, which shares the item's answer when the item has the code too (has; see
// nameablesPanel.has) and otherwise asks for it separately. The code itself is not offered, since a link to it would
// be the same as none.
func (p *nameablesPanel) linkEditor(path string, one *nameableModifier, code string, has bool) *unison.Panel {
	box := newColumn()
	box.RefKey = path + keyFirst
	fields := newFlow()
	box.AddChild(fields)
	addJoiningWords(fields, i18n.Text("%s on %s", markerText(code), one.ownerText()))
	choices := []string{""}
	for _, target := range p.codes() {
		if target != code {
			choices = append(choices, target)
		}
	}
	addCentered(fields, compactPopup(&p.sentenceRows, path+":link", i18n.Text("Link for %s", markerText(code)),
		choices, p.linkOf(one.id, code), func(target string) string { return linkChoice(target, has) },
		func(target string) { p.setLink(one.id, code, target) }))
	return box
}

// linkChoice names a choice of the link popup. has is whether the item has the code too (see nameablesPanel.has).
func linkChoice(target string, has bool) string {
	switch {
	case target != "":
		return i18n.Text("takes the answer to %s", markerText(target))
	case has:
		return i18n.Text("shares the item's answer")
	default:
		return i18n.Text("is asked for under its own code")
	}
}

// setLink links the modifier's code to the item's target code, or unlinks it when target is empty.
func (p *nameablesPanel) setLink(id tid.TID, code, target string) {
	if target == "" {
		delete((*p.links)[id], code)
		if len((*p.links)[id]) == 0 {
			delete(*p.links, id)
		}
		return
	}
	if *p.links == nil {
		*p.links = make(map[tid.TID]map[string]string)
	}
	if (*p.links)[id] == nil {
		(*p.links)[id] = make(map[string]string)
	}
	(*p.links)[id][code] = target
}

// nameableUses returns how many times the texts refer to each nameable code, each prerequisite, feature, weapon and
// other filler that refers to one counting once.
func nameableUses(texts []string, prereq *gurps.PrereqList, features gurps.Features, weapons []*gurps.Weapon,
	more ...nameable.Filler,
) map[string]int {
	uses := nameable.Uses(texts...)
	if prereq != nil {
		for _, one := range prereq.Prereqs {
			more = append(more, one)
		}
	}
	for _, one := range features {
		more = append(more, one)
	}
	for _, one := range weapons {
		more = append(more, one)
	}
	for _, one := range more {
		if xreflect.IsNil(one) {
			continue
		}
		m := make(map[string]string)
		one.FillWithNameableKeys(m, nil)
		for code := range m {
			uses[code]++
		}
	}
	return uses
}

// nameableModifiers returns the modifiers, containers included, for the nameables panel, refs returning the texts and
// features of each, by which it refers to codes.
func nameableModifiers[T interface {
	gurps.Node[T]
	Enabled() bool
	nameable.Definer
}](modifiers []T, refs func(T) (texts []string, features gurps.Features)) []nameableModifier {
	var result []nameableModifier
	gurps.Traverse(func(mod T) bool {
		texts, features := refs(mod)
		result = append(result, nameableModifier{
			id:       mod.ID(),
			location: rowLocation(mod),
			name:     mod.String(),
			disabled: !mod.Enabled(),
			defs:     mod.NameableDefinitions(),
			uses:     nameableUses(texts, nil, features, nil),
			granted:  grantedNameableDefinitions(features),
		})
		return false
	}, false, false, modifiers...)
	return result
}

// traitNameableModifiers returns the trait's modifiers for the nameables panel (see nameableModifiers).
func traitNameableModifiers(modifiers []*gurps.TraitModifier) []nameableModifier {
	return nameableModifiers(modifiers, func(m *gurps.TraitModifier) ([]string, gurps.Features) {
		return []string{m.Name, m.LocalNotes}, m.Features
	})
}

// equipmentNameableModifiers returns the equipment's modifiers for the nameables panel (see nameableModifiers).
func equipmentNameableModifiers(modifiers []*gurps.EquipmentModifier) []nameableModifier {
	return nameableModifiers(modifiers, func(m *gurps.EquipmentModifier) ([]string, gurps.Features) {
		return []string{m.Name, m.LocalNotes}, m.Features
	})
}
