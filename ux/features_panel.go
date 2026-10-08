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
	"strconv"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/equipmentsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/maxusesmod"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selector"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/skillsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellmatch"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/traitsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xslices"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
	"github.com/zeebo/xxh3"
)

var (
	lastFeatureTypeUsed   = feature.AttributeBonus
	lastAttributeIDUsed   = gurps.StrengthID
	lastSelectorFieldUsed = selector.WeaponDamageType
)

// featureAddKey is the reference key of the section's add button. A feature's row is found by its path, which is its
// index in the list.
const featureAddKey = "add"

// featureEmptyKey is the reference key of the placeholder shown in place of the rows while there are none.
const featureEmptyKey = "empty"

// featureDropKey marks a row a dragged feature can be dropped on, holding its path.
const featureDropKey = "feature.drop"

// featuresPanel edits a list of features. Each one is a row that reads as a sentence until it is opened, one at a time,
// to edit it. Every change, typing included, records a snapshot of the whole list with the editor's undo manager.
// Clicking the title collapses the panel to a paragraph of every row's sentence. It starts out collapsed when there
// are features, and open to add one when there are none.
type featuresPanel struct {
	sentenceRows[featuresState]
	entity   *gurps.Entity
	owner    fmt.Stringer
	features *gurps.Features
	// names, when set, gives the nameable replacements in place of the owner (see withReplacementsFrom).
	names nameable.Accesser
	// pending is the key of an optional criterion added to the open row that holds nothing yet. It shows until the row
	// closes, since nothing in the data says it is there.
	pending string
	// percentSuspended holds the weapon damage bonuses whose "as a %" is suspended while they hold dice (see
	// damageField). It is kept here, rather than with the controls, so that it lasts while the panel is rebuilt.
	percentSuspended     map[*gurps.WeaponBonus]bool
	forEquipmentModifier bool
}

// featuresState is the data a snapshot of the panel holds: the features, and the indexes of those whose "as a %" is
// suspended, so that undo and redo suspend it for the copies they install.
type featuresState struct {
	features  gurps.Features
	suspended []int
}

func newFeaturesPanel(entity *gurps.Entity, owner fmt.Stringer, features *gurps.Features, forEquipmentModifier bool) *featuresPanel {
	p := &featuresPanel{
		entity:               entity,
		owner:                owner,
		features:             features,
		percentSuspended:     make(map[*gurps.WeaponBonus]bool),
		forEquipmentModifier: forEquipmentModifier,
	}
	border := initTitledEditorSection(p, i18n.Text("Features"))
	p.initRows(featureDragKey, p.build, p.state, p.setState, p.dataHash)
	p.initCollapse(border, len(*features) != 0)
	p.initDrop(p.dropAt, p.drop)
	p.build()
	return p
}

// state returns a copy of the panel's data, for a snapshot.
func (p *featuresPanel) state() featuresState {
	state := featuresState{features: p.features.Clone()}
	for i, one := range *p.features {
		if bonus, ok := one.(*gurps.WeaponBonus); ok && p.percentSuspended[bonus] {
			state.suspended = append(state.suspended, i)
		}
	}
	return state
}

// setState installs a copy of the data of a snapshot, suspending the percentage of the copies it held suspended.
func (p *featuresPanel) setState(state featuresState) {
	*p.features = state.features.Clone()
	clear(p.percentSuspended)
	for _, i := range state.suspended {
		if bonus, ok := (*p.features)[i].(*gurps.WeaponBonus); ok {
			p.percentSuspended[bonus] = true
		}
	}
}

// dataHash returns a hash of the features.
func (p *featuresPanel) dataHash() uint64 {
	h := xxh3.New()
	xhash.Num64(h, len(*p.features))
	for _, one := range *p.features {
		one.Hash(h)
	}
	return h.Sum64()
}

// index returns the index of the feature at the path, or -1 if there is none.
func (p *featuresPanel) index(path string) int {
	if i, err := strconv.Atoi(path); err == nil && i >= 0 && i < len(*p.features) {
		return i
	}
	return -1
}

// withReplacementsFrom has the panel take the values of the nameable markers in its features from source rather than
// from the owner, such as from an editor's data, which Set Substitutions changes, and fills it again with them.
func (p *featuresPanel) withReplacementsFrom(source nameable.Accesser) *featuresPanel {
	p.names = source
	p.RemoveAllChildren()
	p.build()
	return p
}

// replacements returns the values of the nameable markers in the features: those of the source withReplacementsFrom
// set, or else those the owning item gives them.
func (p *featuresPanel) replacements() map[string]string {
	if p.names != nil {
		return p.names.NameableReplacements()
	}
	if owner, ok := p.owner.(nameable.Accesser); ok {
		return owner.NameableReplacements()
	}
	return nil
}

func (p *featuresPanel) build() {
	if i := p.index(p.open); i < 0 || (*p.features)[i].FeatureType() == feature.Unknown {
		p.open = ""
	}
	if p.open == "" || !strings.HasPrefix(p.pending, p.open+":") {
		p.pending = ""
	}
	// Collapsed, the features read as one paragraph.
	if p.addTitleBar(p.summary) != nil {
		return
	}
	add := newSectionAddButton(p, i18n.Text("Add a feature"), func() bool {
		t := lastFeatureTypeUsed
		if !slices.Contains(p.featureTypesList(), t) {
			t = feature.AttributeBonus
		}
		created := p.createFeatureForType(t)
		if created == nil {
			return false
		}
		path := strconv.Itoa(len(*p.features))
		p.edit(i18n.Text("Add Feature"), featureAddKey, path+keyFirst, func() {
			*p.features = append(*p.features, created)
			p.open = path
		})
		return true
	})
	add.RefKey = featureAddKey
	for i, one := range *p.features {
		p.AddChild(p.row(one, strconv.Itoa(i)))
	}
	// New features go at the end, so the add button sits under the last row, in line with the more buttons.
	foot := unison.NewPanel()
	foot.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: 4, Right: 8}))
	empty := len(*p.features) == 0
	if empty {
		// With no features, a placeholder that adds one as the add button does stands before it.
		foot.AddChild(newEmptyPlaceholder(featureEmptyKey, i18n.Text("No features. Click here to add one."),
			add.ClickCallback))
	}
	foot.AddChild(add)
	hbox(foot, unison.StdHSpacing)
	add.SetLayoutData(&unison.FlexLayoutData{HAlign: align.End, VAlign: align.Middle, HGrab: !empty})
	p.AddChild(foot)
}

// summary returns the paragraph a collapsed panel shows: the sentence of each row, ended with a period.
func (p *featuresPanel) summary() string {
	if len(*p.features) == 0 {
		return i18n.Text("No features.")
	}
	replacements := p.replacements()
	sentences := make([]string, 0, len(*p.features))
	for _, one := range *p.features {
		sentences = append(sentences, i18n.Text("%s.", one.Describe(p.entity, replacements, emphasize)))
	}
	return strings.Join(sentences, " ")
}

// row returns the panel for a feature: its sentence, or while it is open its editor, beside a button for more actions.
func (p *featuresPanel) row(f gurps.Feature, path string) *unison.Panel {
	row := p.sentenceRow(path, func() string { return f.Describe(p.entity, p.replacements(), emphasize) },
		f.FeatureType() != feature.Unknown, func() *unison.Panel { return p.editor(f, path) },
		func() []menuEntry { return p.moreEntries(path) }, nil, nil)
	row.ClientData()[featureDropKey] = path
	return row
}

// moreEntries returns the entries of the more menu of the feature at the path: Duplicate, Move up and down, and Delete.
func (p *featuresPanel) moreEntries(path string) []menuEntry {
	i := p.index(path)
	if i < 0 {
		return nil
	}
	from := path + keyMore
	entries := []menuEntry{{Label: i18n.Text("Duplicate"), Act: func() {
		p.restructure(i18n.Text("Duplicate Feature"), from, "", func() int {
			*p.features = slices.Insert(*p.features, i+1, (*p.features)[i].Clone())
			return i + 1
		})
	}}}
	titles := []string{i18n.Text("Move Up"), i18n.Text("Move Down")}
	for k, j := range []int{i - 1, i + 1} {
		if j < 0 || j >= len(*p.features) {
			continue
		}
		title := titles[k]
		entries = append(entries, menuEntry{Label: title, Act: func() {
			p.restructure(title, from, "", func() int {
				list := *p.features
				list[i], list[j] = list[j], list[i]
				return j
			})
		}})
	}
	return append(entries, menuEntry{}, menuEntry{Label: i18n.Text("Delete"), Act: func() {
		p.restructure(i18n.Text("Delete Feature"), from, featureAddKey, func() int {
			*p.features = slices.Delete(*p.features, i, i+1)
			if i < len(*p.features) {
				return i
			}
			return -1
		})
	}})
}

// dropAt returns the row a feature dragged to where would be dropped on and whether it would go before or after it, by
// which half of the row where is in. The row is nil where the feature can't go, such as onto itself.
func (p *featuresPanel) dropAt(where geom.Point, data any) (target *unison.Panel, at int) {
	from := p.dragPath(data)
	if p.index(from) < 0 {
		return nil, 0
	}
	for target = p.PanelAt(where); target != nil && target != p.AsPanel(); target = target.Parent() {
		path, isRow := target.ClientData()[featureDropKey].(string)
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

// drop moves the dragged feature to where it would go.
func (p *featuresPanel) drop(where geom.Point, data any) {
	target, at := p.dropAt(where, data)
	p.dragExit()
	if target == nil {
		return
	}
	path, isRow := target.ClientData()[featureDropKey].(string)
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
		// Taking the feature out moves what comes after it up by one.
		index--
	}
	if i == index {
		return
	}
	p.restructure(i18n.Text("Move Feature"), from+keyMore, "", func() int {
		one := (*p.features)[i]
		*p.features = slices.Insert(slices.Delete(*p.features, i, i+1), index, one)
		return index
	})
}

// restructure moves, adds or removes features through the widget with the reference key from, then rebuilds. change
// returns the index of the feature whose more button takes the focus; with -1, the widget with the reference key
// fallback does. The open row stays open wherever it ends up, and closes if it is removed.
func (p *featuresPanel) restructure(title, from, fallback string, change func() int) {
	var open gurps.Feature
	if i := p.index(p.open); i >= 0 {
		open = (*p.features)[i]
	}
	p.sentenceRows.restructure(title, from, fallback, func() string {
		dst := change()
		p.open = ""
		if i := slices.Index(*p.features, open); open != nil && i >= 0 {
			p.open = strconv.Itoa(i)
		}
		if dst < 0 {
			return ""
		}
		return strconv.Itoa(dst) + keyMore
	})
}

// featureIndent is how far the lines of an open row after its first are indented, to show they belong to it.
const featureIndent = 12

// editor returns the controls for an open row: its type, followed by those that say what the feature does, flowing as
// a sentence would, then its optional criteria as chips. Every line after the first is indented a little.
func (p *featuresPanel) editor(f gurps.Feature, path string) *unison.Panel {
	editor := newColumn()
	editor.RefKey = path + keyFirst
	fields := newHangingFlow(featureIndent)
	editor.AddChild(fields)
	p.typePopup(fields, path, f)
	box := newColumn()
	box.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: featureIndent}))
	chips := newFlow()
	key := func(name string) string { return path + ":" + name }
	var note *unison.Label
	switch one := f.(type) {
	case *gurps.AttributeBonus:
		p.amount(fields, path, &one.Amount, &one.PerLevel)
		p.attributePopup(fields, key("attribute"), i18n.Text("to"), &one.Attribute, func(id string) { lastAttributeIDUsed = id })
		if one.Attribute == gurps.StrengthID {
			p.optional(chips, path, "limitation", i18n.Text("+ limitation"), i18n.Text("Add Limitation"),
				i18n.Text("Remove Limitation"), one.Limitation != stlimit.None,
				func() { one.Limitation = stlimit.StrikingOnly },
				func() { one.Limitation = stlimit.None },
				func(chip *unison.Panel) {
					addCentered(chip, compactPopup(&p.sentenceRows, key("limitation"), i18n.Text("Limitation"),
						stlimit.Options[1:], one.Limitation, nil, func(v stlimit.Option) { one.Limitation = v }))
				})
		}
	case *gurps.ConditionalModifierBonus:
		p.amount(fields, path, &one.Amount, &one.PerLevel)
		p.situation(box, chips, path, &one.Situation, &one.Group, i18n.Text("Triggering Condition"))
	case *gurps.ReactionBonus:
		p.amount(fields, path, &one.Amount, &one.PerLevel)
		p.situation(box, chips, path, &one.Situation, &one.Group, i18n.Text("from/to target"))
	case *gurps.DRBonus:
		p.amount(fields, path, &one.Amount, &one.PerLevel)
		note = p.locations(fields, path, one)
		spec := strings.TrimSpace(one.Specialization)
		// After the locations, which it qualifies.
		p.optional(fields, path, "against", i18n.Text("+ against"), i18n.Text("Add Attack Specialization"),
			i18n.Text("Remove Attack Specialization"),
			(spec != "" && !strings.EqualFold(spec, gurps.AllID)) || p.pending == key("against"),
			func() { p.pending = key("against") },
			func() { one.Specialization, p.pending = gurps.AllID, "" },
			func(chip *unison.Panel) {
				addJoiningWords(chip, i18n.Text("against"))
				// Named outright, since the words before the field would otherwise be taken as its name.
				p.textField(chip, key("against"), i18n.Text("Attack Specialization"), gurps.AllID,
					&one.Specialization, nil).Accessibility.Name = i18n.Text("Attack Specialization")
				addJoiningWords(chip, i18n.Text("attacks"))
			})
	case *gurps.SkillBonus:
		p.amount(fields, path, &one.Amount, &one.PerLevel)
		selectionPopup(p, fields, path, skillsel.Types, &one.SelectionType, skillsel.ThisWeapon, &one.NameCriteria)
		if one.SelectionType == skillsel.Name {
			p.textChip(chips, path, "specialization", &one.SpecializationCriteria)
			p.textChip(chips, path, "optspecialization", &one.OptionalSpecializationCriteria)
		} else {
			p.textChip(chips, path, "usage", &one.SpecializationCriteria)
		}
		if one.SelectionType != skillsel.ThisWeapon {
			p.tagsChip(chips, path, &one.TagsCriteria)
		}
	case *gurps.SkillPointBonus:
		p.amount(fields, path, &one.Amount, &one.PerLevel)
		whose := i18n.Text("to skills whose name")
		p.textCriteria(fields, key("name"), i18n.Text("Name"), "", whose, whose, &one.NameCriteria, true)
		p.textChip(chips, path, "specialization", &one.SpecializationCriteria)
		p.textChip(chips, path, "optspecialization", &one.OptionalSpecializationCriteria)
		p.tagsChip(chips, path, &one.TagsCriteria)
	case *gurps.SpellBonus:
		p.amount(fields, path, &one.Amount, &one.PerLevel)
		selectionPopup(p, fields, path, spellmatch.Types, &one.SpellMatchType, spellmatch.AllColleges, &one.NameCriteria)
		p.tagsChip(chips, path, &one.TagsCriteria)
	case *gurps.SpellPointBonus:
		p.amount(fields, path, &one.Amount, &one.PerLevel)
		selectionPopup(p, fields, path, spellmatch.Types, &one.SpellMatchType, spellmatch.AllColleges, &one.NameCriteria)
		p.tagsChip(chips, path, &one.TagsCriteria)
	case *gurps.TraitBonus:
		p.amount(fields, path, &one.Amount, &one.PerLevel)
		whose := i18n.Text("to traits whose name")
		p.textCriteria(fields, key("name"), i18n.Text("Name"), "", whose, whose, &one.NameCriteria, true)
		p.tagsChip(chips, path, &one.TagsCriteria)
	case *gurps.EquipmentMaxUsesBonus:
		p.maxAdjustment(fields, path, &one.MaxUsesModAmount, i18n.Text("Maximum Uses Adjustment"))
		selectionPopup(p, fields, path, equipmentsel.Types, &one.SelectionType, equipmentsel.ThisEquipment,
			&one.NameCriteria)
		if one.SelectionType == equipmentsel.EquipmentWithName {
			p.tagsChip(chips, path, &one.TagsCriteria)
		}
	case *gurps.TraitMaxLevelBonus:
		p.maxAdjustment(fields, path, &one.MaxUsesModAmount, i18n.Text("Maximum Level Adjustment"))
		selectionPopup(p, fields, path, traitsel.Types, &one.SelectionType, traitsel.ThisTrait, &one.NameCriteria)
		if one.SelectionType == traitsel.TraitWithName {
			p.tagsChip(chips, path, &one.TagsCriteria)
		}
	case *gurps.WeaponBonus:
		p.weaponBonus(fields, chips, path, one)
	case *gurps.CostReduction:
		p.attributePopup(fields, key("attribute"), "", &one.Attribute, nil)
		items := make([]fxp.Int, 0, 17)
		for i := 5; i <= 80; i += 5 {
			items = append(items, fxp.FromInteger(i))
		}
		if !slices.Contains(items, one.Percentage) {
			items = append(items, one.Percentage)
		}
		addCentered(fields, compactPopup(&p.sentenceRows, key("reduction"), i18n.Text("Reduction"), items,
			one.Percentage, func(v fxp.Int) string { return i18n.Text("by %s%%", v.String()) },
			func(v fxp.Int) { one.Percentage = v }))
	case *gurps.ContainedWeightReduction:
		title := i18n.Text("Contained Weight Reduction")
		units := gurps.SheetSettingsFor(p.entity).DefaultWeightUnits
		field := NewStringField(p.targetMgr, key("reduction"), title, func() string { return one.Reduction },
			func(s string) {
				p.edit(title, key("reduction"), "", func() {
					//nolint:errcheck // A valid value is always returned
					one.Reduction, _ = gurps.ExtractContainedWeightReduction(s, units)
				})
			})
		field.SetMinimumTextWidthUsing("1,000 lb")
		field.Tooltip = newWrappedTooltip(i18n.Text(`Enter a weight or percentage, e.g. "2 lb" or "5%"`))
		field.ValidateCallback = func() bool {
			_, err := gurps.ExtractContainedWeightReduction(field.Text(), units)
			return err == nil
		}
		p.addCompact(fields, field.withoutUndo())
	case *gurps.SelectorOverride:
		p.selectorOverride(fields, chips, path, one)
	default:
		errs.Log(errs.New("unknown feature type"), "type", f.FeatureType().Key())
	}
	if len(chips.Children()) != 0 {
		// The buttons that add unused criteria go after the chips in use, so an added one takes its place among them.
		for _, child := range slices.Clone(chips.Children()) {
			if _, ok := child.Self.(*unison.Button); ok {
				chips.AddChild(child)
			}
		}
		p.switchable(chips, path, f)
		box.AddChild(chips)
	} else {
		p.switchable(fields, path, f)
	}
	if note != nil {
		box.AddChild(note)
	}
	if len(box.Children()) != 0 {
		editor.AddChild(box)
	}
	return editor
}

// featureTypeEntry is an entry of the feature type popup: a type, or the heading of a group of them.
type featureTypeEntry struct {
	heading     string
	featureType feature.Type
}

// featureTypeGroups returns the headings of the groups the feature type popup files the types under, in order.
func featureTypeGroups() []string {
	return []string{
		i18n.Text("Attributes and Traits"), i18n.Text("Skills and Spells"), i18n.Text("Defense"),
		i18n.Text("Weapons"), i18n.Text("Equipment"),
	}
}

// featureTypeGroup returns the index of the group the feature type popup files the type under.
func featureTypeGroup(t feature.Type) int {
	if t.IsWeaponBonus() {
		return 3
	}
	switch t {
	case feature.SkillBonus, feature.SkillPointBonus, feature.SpellBonus, feature.SpellPointBonus:
		return 1
	case feature.DRBonus:
		return 2
	case feature.SelectorOverride:
		return 3
	case feature.EquipmentMaxUsesBonus, feature.ContainedWeightReduction:
		return 4
	default:
		return 0
	}
}

// typePopup adds the popup that switches a feature to another type, keeping whether it is switchable, with the types
// under headings of what they apply to. A type that isn't offered here, which a file can still hold, is shown but not
// offered for others.
func (p *featuresPanel) typePopup(parent *unison.Panel, path string, f gurps.Feature) {
	current := featureTypeEntry{featureType: f.FeatureType()}
	types := p.featureTypesList()
	if !slices.Contains(types, current.featureType) {
		types = append(slices.Clone(types), current.featureType)
	}
	popup := compactPopup(&p.sentenceRows, path+":type", i18n.Text("Feature Type"), nil, current,
		func(e featureTypeEntry) string {
			if e.heading != "" {
				return e.heading
			}
			return e.featureType.String()
		},
		func(e featureTypeEntry) {
			i := p.index(path)
			if i < 0 {
				return
			}
			if created := p.createFeatureForType(e.featureType); created != nil {
				lastFeatureTypeUsed = e.featureType
				created.SetSwitchable(f.IsSwitchable())
				(*p.features)[i] = created
			}
		})
	// Filled in afterward, since compactPopup takes no headings, with the callback held back so that selecting the
	// current type isn't taken as a choice.
	choose := popup.SelectionChangedCallback
	popup.SelectionChangedCallback = nil
	for i, heading := range featureTypeGroups() {
		var group []featureTypeEntry
		for _, t := range types {
			if featureTypeGroup(t) == i {
				group = append(group, featureTypeEntry{featureType: t})
			}
		}
		if len(group) == 0 {
			continue
		}
		if popup.ItemCount() != 0 {
			popup.AddSeparator()
		}
		popup.AddDisabledItem(featureTypeEntry{heading: heading})
		popup.AddItem(group...)
	}
	popup.Select(current)
	popup.SelectionChangedCallback = choose
	addCentered(parent, popup)
}

// switchable adds the pill that makes the feature switchable, which comes last: after the chips, or at the end of the
// sentence when there are none.
func (p *featuresPanel) switchable(parent *unison.Panel, path string, f gurps.Feature) {
	words := i18n.Text("only while switched on")
	p.optional(parent, path, "switchable", i18n.Text("+ switchable"), i18n.Text("Add Switchable"),
		i18n.Text("Remove Switchable"), f.IsSwitchable(),
		func() { f.SetSwitchable(true) },
		func() { f.SetSwitchable(false) },
		func(chip *unison.Panel) { addJoiningWords(chip, words) })
	tip := newWrappedTooltip(gurps.SwitchableTooltip())
	if one := parent.FindRefKey(path + ":add switchable"); one != nil {
		one.Tooltip = tip
	} else if one = parent.FindRefKey(path + ":switchable" + keyChip); one != nil {
		one.Tooltip = tip
	}
}

// checkBox returns a checkbox titled title, checked while on, which hands a change to set within an edit.
func (p *featuresPanel) checkBox(key, title string, on bool, set func(bool)) *unison.CheckBox {
	box := unison.NewCheckBox()
	box.SetTitle(title)
	box.RefKey = key
	box.State = check.FromBool(on)
	box.ClickCallback = func() {
		p.edit(title, key, "", func() { set(box.State == check.On) })
		// Each click is a step of its own to undo, unlike a run of typing.
		p.editKey = ""
	}
	return box
}

// amount adds the field for an amount and the checkbox that gives it per level.
func (p *featuresPanel) amount(parent *unison.Panel, path string, amount *fxp.Int, perLevel *bool) {
	key := path + ":amount"
	title := i18n.Text("Amount")
	p.addCompact(parent, NewDecimalField(p.targetMgr, key, title, func() fxp.Int { return *amount },
		func(v fxp.Int) { p.edit(title, key, "", func() { *amount = v }) }, fxp.Min, fxp.Max, true, false).withoutUndo())
	p.perLevel(parent, path, perLevel)
}

func (p *featuresPanel) perLevel(parent *unison.Panel, path string, perLevel *bool) {
	addCentered(parent, p.checkBox(path+":perlevel", i18n.Text("per level"), *perLevel,
		func(on bool) { *perLevel = on }))
}

// attributePopup adds a popup of attributes, each after the prefix, handing a choice to chosen as well when it is set.
// A key that isn't one of them is shown as such, and kept until another is chosen.
func (p *featuresPanel) attributePopup(parent *unison.Panel, key, prefix string, value *string, chosen func(string)) {
	choices, current := gurps.AttributeChoices(p.entity, prefix, gurps.SizeFlag|gurps.DodgeFlag|gurps.ParryFlag|
		gurps.BlockFlag, *value)
	addCentered(parent, compactPopup(&p.sentenceRows, key, i18n.Text("Attribute"), choices, current,
		func(c *gurps.AttributeChoice) string { return c.Title }, func(c *gurps.AttributeChoice) {
			*value = c.Key
			if chosen != nil {
				chosen(c.Key)
			}
		}))
}

// selectionPopup adds the popup that picks what a bonus applies to, followed by the criteria for their name unless
// this, which needs none, is picked. It is a plain function because methods cannot have type parameters.
func selectionPopup[E comparable](p *featuresPanel, parent *unison.Panel, path string, items []E, selection *E, this E, name *criteria.Text) {
	addCentered(parent, compactPopup(&p.sentenceRows, path+":selection", i18n.Text("Selection Type"), items,
		*selection, nil, func(v E) { *selection = v }))
	if *selection != this {
		p.textCriteria(parent, path+":name", i18n.Text("Name"), "", "", "", name, true)
	}
}

// situation adds the field for the situation of a conditional modifier or reaction bonus, which shows the hint while
// empty, on a line of its own where a long situation can wrap, and the chip for the group it is filed under in the
// table that displays it.
func (p *featuresPanel) situation(box, chips *unison.Panel, path string, situation, group *string, hint string) {
	key := path + ":group"
	title := i18n.Text("Situation")
	field := NewMultiLineStringField(p.targetMgr, path+":situation", title, func() string { return *situation },
		func(s string) { p.edit(title, path+":situation", "", func() { *situation = s }) })
	field.Watermark = hint
	// Named for the hint, as it always has been.
	field.Accessibility.Name = hint
	// It grows to fit its text, so scrolling within it would only shift the text out of place.
	field.AutoScroll = false
	p.addCompact(box, field.withoutUndo())
	field.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	p.optional(chips, path, "group", i18n.Text("+ group"), i18n.Text("Add Group"), i18n.Text("Remove Group"),
		*group != "" || p.pending == key,
		func() { p.pending = key },
		func() { *group, p.pending = "", "" },
		func(chip *unison.Panel) {
			addJoiningWords(chip, i18n.Text("in group"))
			// Named outright, since the words before the field would otherwise be taken as its name.
			p.textField(chip, key, i18n.Text("Group"), "", group, nil).Accessibility.Name = i18n.Text("Group")
		})
}

// locations adds the popup of where a DR bonus applies and, while that is a list of locations, a checkbox for each
// after it, continuing the sentence. "To this armor" is offered only by an equipment modifier's bonus, or shown when a
// file holds it elsewhere. It returns the note that some of the locations aren't in the body, if they aren't, which
// goes on a line of its own after the sentence.
func (p *featuresPanel) locations(fields *unison.Panel, path string, one *gurps.DRBonus) *unison.Label {
	thisArmor, all, chosen := i18n.Text("to this armor"), i18n.Text("to all locations"), i18n.Text("to these locations:")
	current := chosen
	switch {
	case len(one.Locations) == 0:
		current = thisArmor
	case slices.Contains(one.Locations, gurps.AllID):
		current = all
	}
	var items []string
	if p.forEquipmentModifier || current == thisArmor {
		items = append(items, thisArmor)
	}
	items = append(items, all, chosen)
	addCentered(fields, compactPopup(&p.sentenceRows, path+":locations", i18n.Text("Locations"), items, current, nil,
		func(choice string) {
			switch choice {
			case thisArmor:
				one.Locations = nil
			case all:
				one.Locations = []string{gurps.AllID}
			default:
				one.Locations = slices.DeleteFunc(slices.Clone(one.Locations),
					func(loc string) bool { return loc == gurps.AllID })
				if len(one.Locations) == 0 {
					one.Locations = []string{gurps.TorsoID}
				}
			}
		}))
	if current != chosen {
		return nil
	}
	return p.locationBoxes(fields, path, one)
}

// locationBoxes adds a checkbox for each hit location of the body to the parent, followed by one for each location of
// the bonus the body doesn't have, which are marked as such, and returns the note saying what the mark means when
// there are any. Ticking more boxes has the bonus cover more locations. Unticking the last one is refused, since a
// bonus with no locations applies to the armor it is attached to instead.
func (p *featuresPanel) locationBoxes(parent *unison.Panel, path string, one *gurps.DRBonus) *unison.Label {
	title := i18n.Text("Locations")
	newBox := func(id, name string) *unison.CheckBox {
		key := path + ":loc " + id
		box := unison.NewCheckBox()
		box.SetTitle(name)
		box.RefKey = key
		box.State = check.FromBool(slices.Contains(one.Locations, id))
		// Rebuilt after each change, which shows a refused one as it was.
		box.ClickCallback = func() {
			p.edit(title, key, key, func() {
				switch {
				case box.State == check.On:
					one.Locations = append(slices.Clone(one.Locations), id)
					slices.Sort(one.Locations)
				case len(one.Locations) > 1:
					one.Locations = slices.DeleteFunc(slices.Clone(one.Locations), func(in string) bool { return in == id })
				}
			})
		}
		return box
	}
	missing := xslices.MapFromKeys(one.Locations, func(in string) string { return in })
	locs := gurps.BodyFor(p.entity).UniqueHitLocations(p.entity)
	boxes := make([]*unison.CheckBox, 0, len(locs)+len(missing))
	for _, loc := range locs {
		boxes = append(boxes, newBox(loc.LocID, loc.ChoiceName))
		delete(missing, loc.LocID)
	}
	const missingKey = "missing"
	for id := range maps.Keys(missing) {
		box := newBox(id, id+"*")
		box.ClientData()[missingKey] = true
		boxes = append(boxes, box)
	}
	slices.SortFunc(boxes, func(a, b *unison.CheckBox) int {
		_, am := a.ClientData()[missingKey]
		_, bm := b.ClientData()[missingKey]
		if am != bm {
			if am {
				return 1
			}
			return -1
		}
		return xstrings.NaturalCmp(a.Text.String(), b.Text.String(), true)
	})
	for _, box := range boxes {
		addCentered(parent, box)
	}
	if len(missing) == 0 {
		return nil
	}
	label := unison.NewLabel()
	fd := label.Font.Descriptor()
	fd.Size *= 0.8
	label.Font = fd.Font()
	label.SetTitle(i18n.Text("* Locations not present in current body type"))
	return label
}

// maxAdjustment adds the field for the amount of a maximum uses or maximum level adjustment, titled title, and the
// checkbox that gives it per level.
func (p *featuresPanel) maxAdjustment(parent *unison.Panel, path string, amount *gurps.MaxUsesModAmount, title string) {
	key := path + ":amount"
	field := NewStringField(p.targetMgr, key, title, func() string { return amount.Amount },
		func(s string) { p.edit(title, key, "", func() { amount.Amount = maxusesmod.Normalize(s) }) })
	field.SetMinimumTextWidthUsing("-1,000,000")
	field.Tooltip = newWrappedTooltip(i18n.Text(`Enter a number, percentage or multiplier, e.g. "-1", "10%" or "x2"`))
	p.addCompact(parent, field.withoutUndo())
	p.perLevel(parent, path, &amount.PerLevel)
}

// weaponBonus adds the controls of a weapon bonus: the flag a weapon switch sets, or the amount of any other.
func (p *featuresPanel) weaponBonus(fields, chips *unison.Panel, path string, one *gurps.WeaponBonus) {
	key := func(name string) string { return path + ":" + name }
	if one.Type == feature.WeaponSwitch {
		items := wswitch.Types[1:]
		if !slices.Contains(items, one.SwitchType) {
			items = append(slices.Clone(items), one.SwitchType)
		}
		addCentered(fields, compactPopup(&p.sentenceRows, key("switch"), i18n.Text("Switch"), items, one.SwitchType,
			nil, func(v wswitch.Type) { one.SwitchType = v }))
		addCentered(fields, compactPopup(&p.sentenceRows, key("switchvalue"), i18n.Text("Switch Value"),
			[]bool{true, false}, one.SwitchTypeValue, func(v bool) string {
				if v {
					return i18n.Text("to true")
				}
				return i18n.Text("to false")
			}, func(v bool) { one.SwitchTypeValue = v }))
	} else {
		var percent *unison.CheckBox
		if one.Type == feature.WeaponBonus {
			p.damageField(fields, path, one, func() {
				percent.State = check.FromBool(one.Percent)
				percent.SetEnabled(one.Dice.IsZero())
			})
			p.perLevel(fields, path, &one.PerLevel)
		} else {
			p.amount(fields, path, &one.Amount, &one.PerLevel)
		}
		if one.Type != feature.WeaponMinSTBonus && one.Type != feature.WeaponEffectiveSTBonus {
			// No per-die for MinST or effective ST bonuses, since that would cause an infinite loop on resolution.
			addCentered(fields, p.checkBox(key("perdie"), i18n.Text("per die"), one.PerDie,
				func(on bool) { one.PerDie = on }))
		}
		percent = p.checkBox(key("percent"), i18n.Text("as a %"), one.Percent, func(on bool) { one.Percent = on })
		// A bonus that adds dice cannot also be a percentage.
		percent.SetEnabled(one.Dice.IsZero())
		addCentered(fields, percent)
	}
	selectionPopup(p, fields, path, wsel.Types, &one.SelectionType, wsel.ThisWeapon, &one.NameCriteria)
	switch one.SelectionType {
	case wsel.WithRequiredSkill:
		p.textChip(chips, path, "specialization", &one.SpecializationCriteria)
		p.textChip(chips, path, "usage", &one.UsageCriteria)
	default:
		p.textChip(chips, path, "usage", &one.SpecializationCriteria)
	}
	if one.SelectionType != wsel.ThisWeapon {
		p.tagsChip(chips, path, &one.TagsCriteria)
	}
	if one.SelectionType == wsel.WithRequiredSkill {
		level := &one.RelativeLevelCriteria
		p.optional(chips, path, "level", i18n.Text("+ relative skill level"), i18n.Text("Add Relative Skill Level"),
			i18n.Text("Remove Relative Skill Level"), level.Compare != criteria.AnyNumber,
			func() { *level = criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.One} },
			func() { *level = criteria.Number{Compare: criteria.AnyNumber} },
			func(chip *unison.Panel) {
				p.numberCriteria(chip, key("level"), i18n.Text("Level"),
					numericWordsAfter(i18n.Text("and whose relative skill level")), level, -fxp.Thousand, fxp.Thousand,
					true, false)
			})
	}
}

// damageField adds the amount field of a weapon damage bonus, which accepts a dice specification as well as a flat
// number, e.g. "+2", "-1d" or "+2d+1x3". Text in any other form is flagged and leaves the bonus unchanged. changed is
// called after each change the field makes to the bonus.
func (p *featuresPanel) damageField(parent *unison.Panel, path string, one *gurps.WeaponBonus, changed func()) {
	key := path + ":amount"
	title := i18n.Text("Amount")
	// Dice cannot be a percentage, so "as a %" is suspended, not cleared, while the field holds dice, and restored once
	// they go away, so the user never ends up with a choice they did not make.
	field := NewStringField(p.targetMgr, key, title,
		func() string { return gurps.FormatWeaponDamageBonus(one.Dice, one.Amount) },
		func(s string) {
			d, amount, ok := gurps.ParseWeaponDamageBonus(s)
			if !ok {
				return
			}
			p.edit(title, key, "", func() {
				one.Dice = d
				one.Amount = amount
				switch {
				case d.IsZero():
					if p.percentSuspended[one] {
						one.Percent = true
						delete(p.percentSuspended, one)
					}
				case one.Percent:
					one.Percent = false
					p.percentSuspended[one] = true
				}
			})
			changed()
		})
	field.ValidateCallback = func() bool {
		_, _, ok := gurps.ParseWeaponDamageBonus(field.Text())
		return ok
	}
	field.SetMinimumTextWidthUsing("+99d+99.99")
	field.Tooltip = newWrappedTooltip(i18n.Text(`Enter a number or a dice specification, e.g. "+2", "-1d", "+1d+2" or "2dx3"`))
	p.addCompact(parent, field.withoutUndo())
}

// selectorOverride adds the controls of a selector override: "[field] to [value] with priority [n]", followed by the
// items it applies to. Changing the field resets the value, since another field may take other values, and clears the
// usage criteria of a field that applies to traits, which have no usage.
func (p *featuresPanel) selectorOverride(fields, chips *unison.Panel, path string, one *gurps.SelectorOverride) {
	key := func(name string) string { return path + ":" + name }
	addCentered(fields, compactPopup(&p.sentenceRows, key("field"), i18n.Text("Field"), selector.Fields, one.Field, nil,
		func(v selector.Field) {
			one.Field = v
			lastSelectorFieldUsed = v
			d := gurps.SelectorFieldDescriptorFor(v)
			one.Value = ""
			if len(d.SuggestedStates) != 0 {
				one.Value = d.SuggestedStates[0]
			}
			if d.Scope == gurps.SelectorScopeTrait {
				one.UsageCriteria = criteria.Text{Compare: criteria.AnyText}
			}
		}))
	addJoiningWords(fields, i18n.Text("to"))
	title := i18n.Text("Value")
	d := gurps.SelectorFieldDescriptorFor(one.Field)
	if len(d.SuggestedStates) != 0 && !d.FreeForm {
		// A constrained field shows a popup of human labels while storing the canonical value behind each one. A value
		// that isn't one of them is shown as it is.
		items := d.SuggestedStates
		if !slices.Contains(items, one.Value) {
			items = append(slices.Clone(items), one.Value)
		}
		var render func(string) string
		if d.StateTitle != nil {
			render = d.StateTitle
		}
		addCentered(fields, compactPopup(&p.sentenceRows, key("value"), title, items, one.Value, render,
			func(v string) { one.Value = v }))
	} else {
		field := p.textField(fields, key("value"), title, "", &one.Value, nil)
		// Named outright, since the words before the field would otherwise be taken as its name.
		field.Accessibility.Name = title
		if len(d.SuggestedStates) != 0 {
			field.Tooltip = newWrappedTooltip(i18n.Text("Suggested values: %s", strings.Join(d.SuggestedStates, ", ")))
		}
		if d.Validate != nil {
			field.ValidateCallback = func() bool { return d.Validate(field.Text()) }
		}
	}
	addJoiningWords(fields, i18n.Text("with priority"))
	priorityTitle := i18n.Text("Priority")
	priority := NewIntegerField(p.targetMgr, key("priority"), priorityTitle, func() int { return one.Priority },
		func(v int) { p.edit(priorityTitle, key("priority"), "", func() { one.Priority = v }) }, -99, 99, false, false)
	// Named outright, since the words before the field would otherwise be taken as its name.
	priority.Accessibility.Name = priorityTitle
	p.addCompact(fields, priority.withoutUndo())
	whose := i18n.Text("on weapons whose name")
	if d.Scope == gurps.SelectorScopeTrait {
		whose = i18n.Text("on traits whose name")
	} else {
		p.textChip(chips, path, "usage", &one.UsageCriteria)
	}
	p.textCriteria(fields, key("name"), i18n.Text("Name"), "", whose, whose, &one.NameCriteria, true)
	p.tagsChip(chips, path, &one.TagsCriteria)
}

func (p *featuresPanel) featureTypesList() []feature.Type {
	if e, ok := p.owner.(*gurps.Equipment); ok && e.Container() {
		return feature.SelectableTypes
	}
	return feature.SelectableTypesWithoutContainedWeightReduction
}

func (p *featuresPanel) createFeatureForType(featureType feature.Type) gurps.Feature {
	if featureType.IsWeaponBonus() {
		bonus := gurps.NewWeaponBonus(featureType)
		if featureType == feature.WeaponSwitch {
			bonus.SwitchType = wswitch.Types[1]
		}
		bonus.SetOwner(p.owner)
		return bonus
	}
	var bonus gurps.Bonus
	switch featureType {
	case feature.AttributeBonus:
		bonus = gurps.NewAttributeBonus(lastAttributeIDUsed)
	case feature.ConditionalModifier:
		bonus = gurps.NewConditionalModifierBonus()
	case feature.ContainedWeightReduction:
		return gurps.NewContainedWeightReduction()
	case feature.CostReduction:
		return gurps.NewCostReduction(lastAttributeIDUsed)
	case feature.EquipmentMaxUsesBonus:
		bonus = gurps.NewEquipmentMaxUsesBonus()
	case feature.DRBonus:
		bonus = gurps.NewDRBonus()
	case feature.ReactionBonus:
		bonus = gurps.NewReactionBonus()
	case feature.SkillBonus:
		bonus = gurps.NewSkillBonus()
	case feature.SkillPointBonus:
		bonus = gurps.NewSkillPointBonus()
	case feature.SpellBonus:
		bonus = gurps.NewSpellBonus()
	case feature.SpellPointBonus:
		bonus = gurps.NewSpellPointBonus()
	case feature.TraitBonus:
		bonus = gurps.NewTraitBonus()
	case feature.TraitMaxLevelBonus:
		bonus = gurps.NewTraitMaxLevelBonus()
	case feature.SelectorOverride:
		override := gurps.NewSelectorOverride(lastSelectorFieldUsed)
		override.SetOwner(p.owner)
		return override
	default:
		errs.Log(errs.New("unknown feature type"), "type", featureType.Key())
		return nil
	}
	bonus.SetOwner(p.owner)
	return bonus
}
