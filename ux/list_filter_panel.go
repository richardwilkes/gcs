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
	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/unison"
	"github.com/zeebo/xxh3"
)

// lastFilterFieldKeyUsed remembers, per list type key, the field the user last chose for a condition, so the next
// condition starts out testing the same field. Session only.
var lastFilterFieldKeyUsed = make(map[string]string)

// listFilterPanel edits a saved filter's tree of nodes. Each condition is a row that reads as a sentence until it is
// opened, one at a time, to edit it. Each group's head says whether all, any, none or not all of its children must
// match, and its children hang from a rail in the head's color. Every change, typing included, records a snapshot of
// the whole tree with the undo manager of the dialog that holds the panel.
type listFilterPanel struct {
	sentenceTree[filterTreeState, gurps.FilterNode]
	// listKey is the key of the list type the filter belongs to.
	listKey string
	filter  *gurps.ListFilter
	// fields are the fields the list type offers, in the order they are shown, fieldKeys their keys in that order, and
	// byKey the same fields by key.
	fields    []filterFieldInfo
	fieldKeys []string
	byKey     map[string]filterFieldInfo
	// weightUnits are the units weights are described in, looked up on each build.
	weightUnits fxp.WeightUnit
	// headed has an empty root show its head, once a group type has been chosen for it.
	headed bool
}

// filterTreeState is the data a snapshot of the panel holds: the tree, and whether an empty root shows its head.
type filterTreeState struct {
	root   *gurps.FilterGroup
	headed bool
}

// newListFilterPanel creates the editor for the given filter. listKey is the list type the filter belongs to, used to
// remember the last field chosen, and fields are the fields that list type offers, in the order they are shown.
func newListFilterPanel(listKey string, filter *gurps.ListFilter, fields []filterFieldInfo) *listFilterPanel {
	p := &listFilterPanel{
		listKey:   listKey,
		filter:    filter,
		fields:    fields,
		fieldKeys: make([]string, len(fields)),
		byKey:     make(map[string]filterFieldInfo, len(fields)),
	}
	for i, info := range fields {
		p.fieldKeys[i] = info.key
		p.byKey[info.key] = info
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	// No titled border here, unlike the prerequisites and features sections of an editor: this panel is the whole of a
	// dialog, which supplies the title, so all it needs is a little breathing room.
	p.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(4)))
	p.initRows(listFilterDragKey, p.build, p.state, p.setState, p.stateHash)
	// An Escape that closes no row cancels the dialog.
	p.passEscape = true
	p.staticTip = func(path string) string {
		if _, ok := p.node(path).(*gurps.FilterCondition); ok {
			return unknownFilterFieldTooltip()
		}
		return preservedFilterNodeTooltip()
	}
	p.initTree(p)
	p.build()
	return p
}

func (p *listFilterPanel) state() filterTreeState {
	return filterTreeState{root: p.filter.Root.CloneAsFilterGroup(nil), headed: p.headed}
}

func (p *listFilterPanel) setState(s filterTreeState) {
	p.filter.Root = s.root.CloneAsFilterGroup(nil)
	p.headed = s.headed
}

func (p *listFilterPanel) stateHash() uint64 {
	h := xxh3.New()
	p.filter.Root.Hash(h)
	xhash.Bool(h, p.headed)
	return h.Sum64()
}

func (p *listFilterPanel) build() {
	if cond, ok := p.node(p.open).(*gurps.FilterCondition); !ok || !p.knows(cond.Field) {
		p.open = ""
	}
	p.weightUnits = gurps.SheetSettingsFor(nil).DefaultWeightUnits
	p.AddChild(p.group(p.filter.Root, treeRootPath))
}

// row returns the panel for a node other than a group: its sentence, or while it is open its editor, beside a button
// for more actions. A node this version of GCS can't edit, of a kind or on a field it doesn't know, is static text.
func (p *listFilterPanel) row(node gurps.FilterNode, path string) *unison.Panel {
	cond, editable := node.(*gurps.FilterCondition)
	if editable {
		editable = p.knows(cond.Field)
	}
	return p.sentenceRow(path, func() string { return p.describe(node) }, editable,
		func() *unison.Panel { return p.editor(cond, path) },
		func() []menuEntry { return p.moreEntries(node, path) }, nil, nil)
}

// describe returns the sentence of a node other than a group.
func (p *listFilterPanel) describe(node gurps.FilterNode) string {
	switch one := node.(type) {
	case *gurps.FilterCondition:
		return one.Describe(p.lookup, p.weightUnits, emphasize)
	case *gurps.UnknownFilterNode:
		return one.Describe()
	}
	return ""
}

// lookup implements gurps.FilterFieldLookup over the fields of the list type.
func (p *listFilterPanel) lookup(key string) (title string, kind gurps.FilterFieldKind, plural, ok bool) {
	info, ok := p.byKey[key]
	return info.title, info.kind, info.plural, ok
}

// knows reports whether the list type has a field with the key.
func (p *listFilterPanel) knows(key string) bool {
	_, ok := p.byKey[key]
	return ok
}

// editor returns the controls for an open condition, flowing as its sentence does: whether it must or must not match,
// the field it tests, and the criteria that go with the field's kind, worded to agree with a plural field and offering
// the field's suggestions, if it has any, for the value it is compared with.
func (p *listFilterPanel) editor(cond *gurps.FilterCondition, path string) *unison.Panel {
	box := newColumn()
	box.RefKey = path + keyFirst
	flow := newFlow()
	box.AddChild(flow)
	key := func(name string) string { return path + ":" + name }
	addCentered(flow, compactPopup(&p.sentenceRows, key("not"), i18n.Text("Must"), []bool{false, true},
		cond.Not, mustWord, func(not bool) { cond.Not = not }))
	addCentered(flow, compactPopup(&p.sentenceRows, key("field"), i18n.Text("Field"), p.fieldKeys, cond.Field,
		func(k string) string { return p.byKey[k].title },
		func(k string) {
			cond.Field = k
			lastFilterFieldKeyUsed[p.listKey] = k
			resetCriteria(cond)
		}))
	info := p.byKey[cond.Field]
	switch info.kind {
	case gurps.FilterFieldText:
		render := criteria.StringComparison.PluralClause
		if !info.plural {
			that := i18n.Text("that")
			render = textWordsAfter(that, that)
		}
		p.textCriteriaWith(flow, key("text"), i18n.Text("Text"), "", render, &cond.Text, true, info.suggestions)
	case gurps.FilterFieldList:
		if field := p.textCriteriaWith(flow, key("text"), i18n.Text("List"), "", criteria.StringComparison.ListClause,
			&cond.Text, true, info.suggestions); field != nil {
			field.Tooltip = newWrappedTooltip(i18n.Text(`Separate multiple values with commas to match any one of them, e.g. "Sword, Axe"`))
		}
	case gurps.FilterFieldNumber:
		p.numberCriteria(flow, key("number"), i18n.Text("Number"), filterNumericWords(info), &cond.Number, fxp.Min,
			fxp.Max, false, true)
	case gurps.FilterFieldWeight:
		p.weightCriteria(flow, key("weight"), i18n.Text("Weight"), filterNumericWords(info), nil, &cond.Weight, true)
	case gurps.FilterFieldBool:
		// Nothing to compare against; Must or Must not covers it.
	}
	return box
}

// filterNumericWords returns the words of the numeric comparison of a condition on the field: after "that", or in
// agreement with a plural title.
func filterNumericWords(info filterFieldInfo) numericWords {
	if info.plural {
		return criteria.NumericComparison.PluralClause
	}
	return numericWordsAfter(i18n.Text("that"))
}

func mustWord(not bool) string {
	if not {
		return i18n.Text("Must not")
	}
	return i18n.Text("Must")
}

// resetCriteria puts the condition's criteria back to where a new condition starts. Each zero criterion accepts
// anything, and a yes/no field, which has none, then requires the value to be true. Only the criterion that goes with
// the field's kind is ever consulted, so the others are cleared rather than kept around where nothing shows or uses
// them, yet would still be written out to disk.
func resetCriteria(cond *gurps.FilterCondition) {
	cond.Text = criteria.Text{}
	cond.Number = criteria.Number{}
	cond.Weight = criteria.Weight{}
}

// addEntries returns what can be added to a group: under the heading condition, a condition that starts out testing
// the field last chosen, and under the heading structure, groups.
func (p *listFilterPanel) addEntries(group *gurps.FilterGroup, path, condition, structure string) []menuEntry {
	at := childPath(path, len(group.Children))
	entries := make([]menuEntry, 0, 5)
	entries = append(entries, menuEntry{Label: condition}, menuEntry{Label: i18n.Text("New Condition"), Act: func() {
		cond := gurps.NewFilterCondition(group, p.defaultFieldKey())
		p.edit(i18n.Text("Add Condition"), p.addKey(group, path), at+":field", func() {
			group.Children = append(group.Children, cond)
			p.open = at
		})
	}}, menuEntry{Label: structure})
	return append(entries, p.groupEntries(group, path, func(all bool) {
		if all {
			filterAllOf.apply(group)
		} else {
			filterAnyOf.apply(group)
		}
		p.headed = true
	})...)
}

// defaultFieldKey returns the key of the field a newly added condition should start out testing: the one last chosen
// for this list type, if it is still among the fields, and otherwise the first field.
func (p *listFilterPanel) defaultFieldKey() string {
	if len(p.fields) == 0 {
		return ""
	}
	if key, ok := lastFilterFieldKeyUsed[p.listKey]; ok && p.knows(key) {
		return key
	}
	return p.fields[0].key
}

// preservedFilterNodeTooltip returns the tooltip that explains the row of a node of a kind the editor doesn't know,
// which it shows but can't edit. It is looked up when needed rather than held in a variable, since the localization
// isn't in place when the package initializes.
func preservedFilterNodeTooltip() string {
	return i18n.Text("This was most likely created by a newer version of GCS. This version can't check it, which hides an item unless the rest of the filter decides without it. Its original data will be written back out unchanged when this filter is saved.")
}

// unknownFilterFieldTooltip returns the tooltip that explains the row of a condition on a field the editor doesn't
// know. Such a condition loads as an ordinary one, so whatever else a newer version of GCS gave it may not survive a
// save.
func unknownFilterFieldTooltip() string {
	return i18n.Text("This was most likely created by a newer version of GCS. This version can't tell what it asks for, so it can't be checked, which hides an item unless the rest of the filter decides without it. Saving the filter may not keep all of it.")
}
