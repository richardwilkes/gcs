// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"fmt"
	"hash"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantmerge"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/zeebo/xxh3"
)

var _ Feature = &Grant{}

// Grant is a feature that puts a trait, skill or spell on the sheet while the feature is in effect. The item it grants
// is held in full by the feature, and the ID of that copy is the key that ties the item on the sheet to the grant. A
// library copies features verbatim, so the key stays the same when the item carrying the grant is synced.
type Grant struct {
	Type feature.Type `json:"type"`
	FeatureSwitch
	// Level is a trait's level, or the tier of a skill or spell: tier 1 is what 1 point buys, tier 2 what 2 points buy
	// and tier n above that what 4×(n-2) points buy. A negative level only takes away from an item already present.
	Level fxp.Int         `json:"level,omitzero"`
	Merge grantmerge.Mode `json:"merge,omitzero"`
	Trait *Trait          `json:"trait,omitempty"`
	Skill *Skill          `json:"skill,omitempty"`
	Spell *Spell          `json:"spell,omitempty"`
	// CheckPrereqs is whether the prerequisites of the item the grant puts on the sheet are checked as for any other
	// item. Otherwise, the item is treated as having them met.
	CheckPrereqs bool `json:"check_prereqs,omitzero"`
}

// GrantedBy ties a trait, skill or spell on a sheet to the grant that put it there.
type GrantedBy struct {
	// Grantor is the ID of the item, or modifier, carrying the grant.
	Grantor tid.TID `json:"grantor"`
	// Key is the ID of the grant's copy of the item.
	Key tid.TID `json:"key"`
	// Hash is the hash of the grant, the answers of what carries it and, for an item with an empty tech level, the
	// sheet's tech level, as they were when the item was last made from them, so that a change to any is noticed.
	Hash string `json:"hash"`
}

// grantState is what a recalculation decided about the grants touching a trait, skill or spell. It is never saved.
type grantState struct {
	// source is what carries the grant that put the item on the sheet. It is set when the grants are reconciled, while
	// the rest is set by each pass.
	source grantSource
	// standIn is what carries the grant standing in for the one that put the item on the sheet, while that one is not
	// in effect; see Entity.applyGrants.
	standIn grantSource
	// inEffect and bypass are whether the grant that put the item on the sheet, or one standing in for it, is in
	// effect, and treats the item's prerequisites as met.
	inEffect bool
	bypass   bool
	// providers are the carriers of the grants in effect that put the item on the sheet, or would have, had another
	// not put it there first; see grantedFor.
	providers []any
	// levelNotes and pointNotes are what each grant gave the item: the levels merged into it or kept for it, and a
	// granted skill's or spell's own points plus any merged into it, of which points is the total. tierPoints is the
	// first of those, the points of the tier the grant in effect gives a granted skill or spell, which stand for its
	// own points when the other grants merge into it. tiersTaken is the tiers the Merge Points grants with a negative
	// level have taken from a skill or spell so far, so that they take their tiers off together (see
	// grantPointsDelta). keepLower is set when the item is a disadvantage, for which the level a Keep Highest grant
	// keeps is the lower one.
	levelNotes []grantNote
	pointNotes []grantNote
	points     fxp.Int
	tierPoints fxp.Int
	tiersTaken fxp.Int
	keepLower  bool
}

// grantItem is what a grant needs of its item, whatever its kind.
type grantItem interface {
	Hashable
	nameable.Filler
	ID() tid.TID
	GetSource() Source
	SyncWithSource()
}

// grantNote is what one grant gave an item: the amount, what carries the grant and that carrier's name. The amount of
// a keep note, which a Keep Highest grant leaves, is the level the grant keeps the item at rather than an addition, or,
// with tier set, the points of the tier it keeps a skill or spell at, whose level is read along with the item's own so
// that both see the same bonuses; see grantState.applyLevelNotes.
type grantNote struct {
	owner  any
	source string
	amount fxp.Int
	keep   bool
	tier   bool
}

// itemLabel returns the name of the item the grant gives, with the level or tier it gives. replacements are those of
// what carries the grant; see itemNameWith.
func (g *Grant) itemLabel(replacements map[string]string) string {
	name := g.itemNameWith(replacements)
	switch {
	case !g.Leveled():
		return name
	case g.Type == feature.GrantTrait:
		return name + " " + g.level().String()
	default:
		return i18n.Text("%s at level %s", name, g.level().String())
	}
}

// level returns the level the grant gives: as set for a trait, whose levels may be fractional, and whole otherwise.
func (g *Grant) level() fxp.Int {
	if g.Type == feature.GrantTrait {
		return g.Level
	}
	return g.Level.Floor()
}

// grantStatus is whether a grant is in effect, or why it is not, for the line Entity.grantSummary gives it.
type grantStatus uint8

const (
	grantActive   grantStatus = iota // In effect.
	grantInactive                    // Not in effect, since what carries it does not contribute its features.
	grantUnmet                       // Not in effect, since the prerequisites of what carries it are not met.
)

// mergeSummary returns what the grant does to the item already on the sheet, for the tooltip of what carries it, or
// what it would do when status says it is not in effect, and why; one that changes nothing about the item says so
// whatever its status. label is the grant's itemLabel, and replacements are those of what carries the grant; see
// itemNameWith.
func (g *Grant) mergeSummary(label string, replacements map[string]string, status grantStatus) string {
	name := g.itemNameWith(replacements)
	merge := g.EffectiveMerge()
	if merge != grantmerge.Ignore && g.mergesNothing() {
		return i18n.Text("Does not change %s, already present, since a level below 1 grants nothing", name)
	}
	pick := func(does, inactive, unmet string) string {
		switch status {
		case grantInactive:
			return inactive
		case grantUnmet:
			return unmet
		default:
			return does
		}
	}
	levels := g.level()
	switch merge {
	case grantmerge.MergeLevels:
		switch {
		case levels == -fxp.One:
			return pick(i18n.Text("Removes 1 level from %s, already present", name),
				i18n.Text("Would remove 1 level from %s, already present, but the grant is not in effect", name),
				i18n.Text(
					"Would remove 1 level from %s, already present, but the prerequisites of what carries the grant are not met",
					name,
				))
		case levels < 0:
			count := levels.Abs().String()
			return pick(i18n.Text("Removes %s levels from %s, already present", count, name),
				i18n.Text("Would remove %s levels from %s, already present, but the grant is not in effect", count, name),
				i18n.Text(
					"Would remove %s levels from %s, already present, but the prerequisites of what carries the grant are not met",
					count, name,
				))
		case levels == fxp.One:
			return pick(i18n.Text("Adds 1 level to %s, already present", name),
				i18n.Text("Would add 1 level to %s, already present, but the grant is not in effect", name),
				i18n.Text(
					"Would add 1 level to %s, already present, but the prerequisites of what carries the grant are not met",
					name,
				))
		default:
			count := levels.String()
			return pick(i18n.Text("Adds %s levels to %s, already present", count, name),
				i18n.Text("Would add %s levels to %s, already present, but the grant is not in effect", count, name),
				i18n.Text(
					"Would add %s levels to %s, already present, but the prerequisites of what carries the grant are not met",
					count, name,
				))
		}
	case grantmerge.MergePoints:
		if levels < 0 {
			count := levelsText(levels.Abs(), nil)
			return pick(i18n.Text("Removes the points of %s from %s, already present", count, name),
				i18n.Text("Would remove the points of %s from %s, already present, but the grant is not in effect",
					count, name),
				i18n.Text(
					"Would remove the points of %s from %s, already present, but the prerequisites of what carries the grant are not met",
					count, name,
				))
		}
		return pick(i18n.Text("Adds the points of level %s to %s, already present", levels.String(), name),
			i18n.Text("Would add the points of level %s to %s, already present, but the grant is not in effect",
				levels.String(), name),
			i18n.Text(
				"Would add the points of level %s to %s, already present, but the prerequisites of what carries the grant are not met",
				levels.String(), name,
			))
	case grantmerge.KeepHighest:
		return pick(i18n.Text("Keeps the better of %s as it is and at level %s", name, levels.String()),
			i18n.Text("Would keep the better of %s as it is and at level %s, but the grant is not in effect", name,
				levels.String()),
			i18n.Text(
				"Would keep the better of %s as it is and at level %s, but the prerequisites of what carries the grant are not met",
				name, levels.String(),
			))
	default:
		return i18n.Text("Does not grant %s, since it is already present", label)
	}
}

// idleSummary returns why the grant, whose plan has no item (see grantPlan), does nothing, for the tooltip of what
// carries it: its item can't be granted, or is absent while the grant is below level 1. replacements are those of what
// carries the grant; see itemNameWith.
func (g *Grant) idleSummary(replacements map[string]string) string {
	name := g.itemNameWith(replacements)
	switch {
	case !g.itemGrantable():
		if skill, ok := g.Item().(*Skill); ok && skill.IsTechnique() {
			return i18n.Text("Does not grant %s, since a technique can't be granted", name)
		}
		return i18n.Text("Does not grant %s, since a container can't be granted", name)
	case g.mergesNothing():
		return i18n.Text("Does not grant %s, since a level below 1 grants nothing", g.itemLabel(replacements))
	default:
		// A negative level would take the levels or points it names away from the item, were it present.
		return i18n.Text("Removes nothing from %s, since it is not present", name)
	}
}

// Item returns the grant's copy of its Item, a trait, skill or spell, or nil if nothing has been chosen yet.
func (g *Grant) Item() grantItem {
	switch {
	case g.Type == feature.GrantTrait && g.Trait != nil:
		return g.Trait
	case g.Type == feature.GrantSkill && g.Skill != nil:
		return g.Skill
	case g.Type == feature.GrantSpell && g.Spell != nil:
		return g.Spell
	default:
		return nil
	}
}

// SetItem makes the grant give a copy of the item, a trait, skill or spell found in the library file, or a new blank
// item of the grant's kind when item is nil. A leveled trait is granted at its level. Any grants the item or its
// modifiers carry are left out of the copy, as SyncItemWithSource leaves them out, since a granted item can't grant
// anything.
func (g *Grant) SetItem(item any, from LibraryFile) {
	g.Trait, g.Skill, g.Spell = nil, nil, nil
	switch one := item.(type) {
	case *Trait:
		g.Trait = one.Clone(from, nil, nil, Reference)
		if g.Trait.IsLeveled() && g.Trait.Levels > 0 {
			g.Level = g.Trait.Levels
		}
	case *Skill:
		g.Skill = one.Clone(from, nil, nil, Reference)
	case *Spell:
		g.Spell = one.Clone(from, nil, nil, Reference)
	default:
		switch g.Type {
		case feature.GrantSkill:
			g.Skill = NewSkill(nil, nil, false)
		case feature.GrantSpell:
			g.Spell = NewSpell(nil, nil, false)
		default:
			g.Trait = NewTrait(nil, nil, false)
		}
	}
	DropGrants(g.Item())
	g.Merge = g.EffectiveMerge()
}

// ItemHasSource returns true if the grant's copy of its item came from a library, and so can be synced with it.
func (g *Grant) ItemHasSource() bool {
	item := g.Item()
	return item != nil && !item.GetSource().IsZero()
}

// SyncItemWithSource brings the grant's copy of its item in line with the library data it came from. The copy keeps
// its ID, which ties the grant to the item it put on the sheet, so that item is refreshed rather than replaced. Any
// grants the library's item has gained are dropped, as SetItem drops them, since a granted item can't grant anything.
func (g *Grant) SyncItemWithSource() {
	item := g.Item()
	if item == nil {
		return
	}
	item.SyncWithSource()
	DropGrants(item)
}

// DropGrants removes the grants among the features of a trait, skill, spell or modifier, and of a trait's modifiers,
// for an item a grant gives or a modifier of one, since a granted item grants nothing. Anything else is left alone.
func DropGrants(item any) {
	var list *Features
	switch one := item.(type) {
	case *Trait:
		list = &one.Features
		Traverse(func(mod *TraitModifier) bool {
			DropGrants(mod)
			return false
		}, false, false, one.Modifiers...)
	case *Skill:
		list = &one.Features
	case *Spell:
		list = &one.Features
	case *TraitModifier:
		list = &one.Features
	case *EquipmentModifier:
		list = &one.Features
	default:
		return
	}
	*list = slices.DeleteFunc(*list, func(f Feature) bool {
		_, isGrant := f.(*Grant)
		return isGrant
	})
}

// CarriesGrants returns true if the trait, skill, spell or equipment, or a modifier of the trait or equipment, carries a
// grant with an item chosen, and so may have put rows on the sheet, in its own list or another.
func CarriesGrants(item any) bool {
	hasGrant := func(list Features) bool {
		return slices.ContainsFunc(list, func(f Feature) bool {
			g, ok := f.(*Grant)
			return ok && g.Key() != ""
		})
	}
	switch one := item.(type) {
	case *Trait:
		if !one.Container() && hasGrant(one.Features) {
			return true
		}
		found := false
		Traverse(func(mod *TraitModifier) bool {
			found = hasGrant(mod.Features)
			return found
		}, false, true, one.Modifiers...)
		return found
	case *Skill:
		return hasGrant(one.Features)
	case *Spell:
		return hasGrant(one.Features)
	case *Equipment:
		if hasGrant(one.Features) {
			return true
		}
		found := false
		Traverse(func(mod *EquipmentModifier) bool {
			found = hasGrant(mod.Features)
			return found
		}, false, true, one.Modifiers...)
		return found
	default:
		return false
	}
}

// NewGrant creates a new Grant of the given type, which must be one of the grant feature types.
func NewGrant(featureType feature.Type) *Grant {
	return &Grant{
		Type:  featureType,
		Level: fxp.One,
	}
}

// IsGrant returns true if the feature type is one of the grant types.
func IsGrant(featureType feature.Type) bool {
	switch featureType {
	case feature.GrantTrait, feature.GrantSkill, feature.GrantSpell:
		return true
	default:
		return false
	}
}

// FeatureType implements Feature.
func (g *Grant) FeatureType() feature.Type {
	return g.Type
}

// Clone implements Feature. The granted item's copy keeps its ID, since that is the grant's key.
func (g *Grant) Clone() Feature {
	other := *g
	if g.Trait != nil {
		other.Trait = g.Trait.Clone(LibraryFile{}, nil, nil, Copy)
	}
	if g.Skill != nil {
		other.Skill = g.Skill.Clone(LibraryFile{}, nil, nil, Copy)
	}
	if g.Spell != nil {
		other.Spell = g.Spell.Clone(LibraryFile{}, nil, nil, Copy)
	}
	return &other
}

// RenewKey gives the grant's copy of its item a new ID, and so the grant a new key, as a duplicate of a grant needs.
func (g *Grant) RenewKey() {
	if g.Trait != nil {
		g.Trait.TID = tid.MustNewTID(traitKind(false))
	}
	if g.Skill != nil {
		g.Skill.TID = tid.MustNewTID(skillKind(false))
	}
	if g.Spell != nil {
		g.Spell.TID = tid.MustNewTID(spellKind(false))
	}
}

// FillWithNameableKeys implements Feature. The placeholders of the granted item that it doesn't answer itself are
// answered by what carries the grant.
func (g *Grant) FillWithNameableKeys(m, existing map[string]string) {
	item := g.Item()
	if item == nil {
		return
	}
	keys := make(map[string]string)
	item.FillWithNameableKeys(keys, nil)
	for k, v := range keys {
		if v == nameable.Unset {
			if answer, ok := existing[k]; ok {
				v = answer
			}
			m[k] = v
		}
	}
}

// answers returns the answers of what carries the grant for the placeholders of the granted item it leaves to them:
// those of the modifier carrying it, which sees its owner's through its nameable links, or else those of the owner, as
// bonusReplacements resolves a bonus.
func (g *Grant) answers(source grantSource) map[string]string {
	var replacements map[string]string
	if na, ok := source.subOwner.(nameable.Accesser); ok {
		replacements = na.NameableReplacements()
	} else if na, ok = source.owner.(nameable.Accesser); ok {
		replacements = na.NameableReplacements()
	} else {
		return nil
	}
	m := make(map[string]string)
	g.FillWithNameableKeys(m, replacements)
	return nameable.Reduce(m, m) // Leaves out the ones not yet answered.
}

// Key returns the key that ties an item on a sheet to this grant, or an empty one if nothing has been chosen yet.
func (g *Grant) Key() tid.TID {
	if item := g.Item(); item != nil {
		return item.ID()
	}
	return ""
}

// ItemName returns the name of the granted item, or an empty string if nothing has been chosen yet. The placeholders
// the item leaves to what carries the grant are left as they are; see itemNameWith.
func (g *Grant) ItemName() string {
	switch item := g.Item().(type) {
	case *Trait:
		return item.NameWithReplacements()
	case fmt.Stringer:
		return item.String()
	default:
		return ""
	}
}

// itemNameWith is ItemName with the placeholders the item leaves to what carries the grant answered by the given
// replacements, those of the carrier, as the item on the sheet answers them (see answers).
func (g *Grant) itemNameWith(replacements map[string]string) string {
	return nameable.Apply(g.ItemName(), replacements)
}

// Leveled returns true if the level means anything for the granted item: always for a skill or spell, and for a trait
// only when it has levels.
func (g *Grant) Leveled() bool {
	if g.Type == feature.GrantTrait {
		return g.Trait != nil && g.Trait.IsLeveled()
	}
	return true
}

// itemGrantable returns true if the grant's item is one a grant can put on the sheet: anything but a container or a
// technique, which the pickers leave out, so that a grant of one from a hand-edited file does nothing.
func (g *Grant) itemGrantable() bool {
	switch g.Type {
	case feature.GrantTrait:
		return g.Trait != nil && !g.Trait.Container()
	case feature.GrantSkill:
		return g.Skill != nil && !g.Skill.IsTechnique() && !g.Skill.Container()
	case feature.GrantSpell:
		return g.Spell != nil && !g.Spell.Container()
	default:
		return false
	}
}

// putsItemOnSheet returns true if the grant puts its item on the sheet when the item is absent: a trait without levels
// always does, and anything else only at level 1 or more, since a negative level only takes away from an item already
// present.
func (g *Grant) putsItemOnSheet() bool {
	return !g.Leveled() || g.Level.Floor() >= fxp.One
}

// mergesNothing returns true if the grant, finding its item already present, changes nothing about it: below level 1
// a leveled grant gives nothing, as Describe says of it, except that a negative level takes away the levels or points
// it names, which Keep Highest never does.
func (g *Grant) mergesNothing() bool {
	if !g.Leveled() || g.putsItemOnSheet() {
		return false
	}
	merge := g.EffectiveMerge()
	return g.level() >= 0 || (merge != grantmerge.MergeLevels && merge != grantmerge.MergePoints)
}

// AllowedMerges returns the merge modes that mean something for the granted item: Ignore alone for a trait without
// levels, everything but Merge Points for a trait with them, and everything for a skill or spell.
func (g *Grant) AllowedMerges() []grantmerge.Mode {
	if !g.Leveled() {
		return []grantmerge.Mode{grantmerge.Ignore}
	}
	if g.Type == feature.GrantTrait {
		return []grantmerge.Mode{grantmerge.Ignore, grantmerge.MergeLevels, grantmerge.KeepHighest}
	}
	return grantmerge.Modes
}

// EffectiveMerge returns the merge mode, treating one that means nothing for the granted item as Ignore.
func (g *Grant) EffectiveMerge() grantmerge.Mode {
	if merge := g.Merge.EnsureValid(); slices.Contains(g.AllowedMerges(), merge) {
		return merge
	}
	return grantmerge.Ignore
}

// contentHash returns the hash recorded in GrantedBy, covering everything about the grant that shapes the item it
// puts on a sheet, answers included, along with the sheet's tech level when it fills in the item's empty one.
func (g *Grant) contentHash(answers map[string]string, techLevel string) string {
	h := xxh3.New()
	HashJSON(h, g)
	if len(answers) != 0 {
		HashJSON(h, answers)
	}
	if item, ok := g.Item().(TechLevelProvider); ok && item.RequiresTL() && item.TL() == "" {
		xhash.StringWithLen(h, techLevel)
	}
	return fmt.Sprintf("%016x", h.Sum64())
}

// cloneGrantedBy returns the link a clone of a granted item carries: a copy of the original's when the clone is for the
// same sheet, whose grants decide whether the clone stays, and none for another sheet, a template or a library, where
// no grant claims it and the clone stands on its own.
func cloneGrantedBy[T Node[T]](link *GrantedBy, original, clone T) *GrantedBy {
	if link == nil {
		return nil
	}
	if e := EntityFromNode(clone); e == nil || e != EntityFromNode(original) {
		return nil
	}
	return clonePtr(link)
}

// grantedByText returns how a cell describes where a granted item came from, or an empty string for an item no grant
// put on the sheet.
func grantedByText(link *GrantedBy, state *grantState) string {
	if link == nil {
		return ""
	}
	if !xreflect.IsNil(state.standIn.owner) {
		return i18n.Text("Granted by %s, standing in for %s, whose grant is not in effect", grantSourceName(state.standIn),
			grantSourceName(state.source))
	}
	text := i18n.Text("Granted by %s", grantSourceName(state.source))
	if !state.inEffect {
		text += "\n" + i18n.Text("The grant is not in effect.")
	}
	return text
}

// Describe implements Feature. The replacements, those of what carries the grant, answer the placeholders the granted
// item leaves to it.
func (g *Grant) Describe(_ *Entity, replacements map[string]string, em func(string) string) string {
	if g.Key() == "" {
		switch g.Type {
		case feature.GrantSkill:
			return describeSwitchable(g.Switchable, i18n.Text("Grants a skill, not yet chosen"))
		case feature.GrantSpell:
			return describeSwitchable(g.Switchable, i18n.Text("Grants a spell, not yet chosen"))
		default:
			return describeSwitchable(g.Switchable, i18n.Text("Grants a trait, not yet chosen"))
		}
	}
	var item string
	switch name := g.itemNameWith(replacements); g.Type {
	case feature.GrantSkill:
		item = i18n.Text("the skill %s", em(name))
	case feature.GrantSpell:
		item = i18n.Text("the spell %s", em(name))
	default:
		item = i18n.Text("the trait %s", em(name))
	}
	level := g.level()
	if g.Leveled() && !g.putsItemOnSheet() {
		// Below level 1, the grant puts nothing on the sheet, and a negative level only takes away from an item present.
		var text string
		switch merge := g.EffectiveMerge(); {
		case level < 0 && merge == grantmerge.MergeLevels:
			text = i18n.Text("Removes %s from %s when already present", levelsText(level.Abs(), em), item)
		case level < 0 && merge == grantmerge.MergePoints:
			text = i18n.Text("Removes the points of %s from %s when already present", levelsText(level.Abs(), em), item)
		default:
			text = i18n.Text("Would grant %s at level %s, but a level below 1 grants nothing", item, em(level.String()))
		}
		return describeSwitchable(g.Switchable, text)
	}
	text := i18n.Text("Grants %s", item)
	if g.Leveled() {
		text += i18n.Text(" at level %s", em(level.String()))
	}
	switch g.EffectiveMerge() {
	case grantmerge.MergeLevels:
		text += i18n.Text(", or adds %s levels to it when already present", em(g.level().StringWithSign()))
	case grantmerge.MergePoints:
		text += i18n.Text(", or adds the points of that level to it when already present")
	case grantmerge.KeepHighest:
		text += i18n.Text(", or keeps the better of the two levels when already present")
	default:
		text += i18n.Text(", unless already present")
	}
	if g.CheckPrereqs {
		text += i18n.Text(", checking its prerequisites")
	}
	return describeSwitchable(g.Switchable, text)
}

// levelsText returns the count of levels as words, "1 level" or "N levels", with the number emphasized by em when that
// is not nil.
func levelsText(levels fxp.Int, em func(string) string) string {
	text := levels.String()
	if em != nil {
		text = em(text)
	}
	if levels == fxp.One {
		return i18n.Text("%s level", text)
	}
	return i18n.Text("%s levels", text)
}

// Hash writes this object's contents into the hasher. The grant's copy of its item is hashed whole, modifiers and edit
// data included, since a sync of what carries the grant replaces the whole copy.
func (g *Grant) Hash(h hash.Hash) {
	if g == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, g.Type)
	xhash.Bool(h, g.Switchable)
	xhash.Num64(h, g.Level)
	xhash.Num8(h, g.Merge)
	xhash.Bool(h, g.CheckPrereqs)
	xhash.StringWithLen(h, string(g.Key()))
	if item := g.Item(); item != nil {
		HashJSON(h, item)
	}
}

// grantTierPoints returns the points that buy the given tier of a skill or spell, or zero below the first.
func grantTierPoints(tier fxp.Int, wildcard bool) fxp.Int {
	var points fxp.Int
	switch tier = tier.Floor(); {
	case tier < fxp.One:
		return 0
	case tier == fxp.One:
		points = fxp.One
	case tier == fxp.Two:
		points = fxp.Two
	default:
		points = fxp.Four.Mul(tier - fxp.Two)
	}
	if wildcard {
		points = points.Mul(fxp.Three)
	}
	return points
}

// grantTierOf returns the tier the given points buy for a skill or spell, or zero when they buy none.
func grantTierOf(points fxp.Int, wildcard bool) fxp.Int {
	if wildcard {
		points = points.Div(fxp.Three)
	}
	switch points = points.Floor(); {
	case points < fxp.One:
		return 0
	case points < fxp.Two:
		return fxp.One
	case points < fxp.Four:
		return fxp.Two
	default:
		return fxp.Two + points.Div(fxp.Four).Floor()
	}
}

// grantPointsDelta returns the change to an item's points a grant of the given tier makes when merged as points: the
// points of that tier, or, for a negative tier, what takes the item down to the tier that many below its own, points
// beyond its tier included, since leaving them would leave the item at its tier when they are as many as the step
// down, and never below zero. taken is the tiers other negative grants have already taken from the item (see
// grantState.tiersTaken), which this one's come off after, so that they stack whatever their order.
func grantPointsDelta(tier, current, taken fxp.Int, wildcard bool) fxp.Int {
	if tier = tier.Floor(); tier >= 0 {
		return grantTierPoints(tier, wildcard)
	}
	have := grantTierOf(current, wildcard)
	from := current
	if taken != 0 {
		from = grantTierPoints((have + taken).Max(0), wildcard)
	}
	return grantTierPoints((have+taken+tier).Max(0), wildcard) - from
}

func (s *grantState) reset() {
	*s = grantState{source: s.source}
}

func (s *grantState) addLevels(source grantSource, amount fxp.Int) {
	if amount != 0 {
		s.levelNotes = append(s.levelNotes, newGrantNote(source, amount))
	}
}

// keepLevel records that a Keep Highest grant keeps the item at the level, the lower of it and the item's own being the
// better when lower is set, as it is for a disadvantage.
func (s *grantState) keepLevel(source grantSource, level fxp.Int, lower bool) {
	note := newGrantNote(source, level)
	note.keep = true
	s.levelNotes = append(s.levelNotes, note)
	s.keepLower = lower
}

// keepTier records that a Keep Highest grant keeps a skill or spell at the level the points of its tier give it, which
// is read along with the item's own level; see applyLevelNotes.
func (s *grantState) keepTier(source grantSource, points fxp.Int) {
	note := newGrantNote(source, points)
	note.keep = true
	note.tier = true
	s.levelNotes = append(s.levelNotes, note)
}

func (s *grantState) addPoints(source grantSource, amount fxp.Int) {
	if amount != 0 {
		s.points += amount
		s.pointNotes = append(s.pointNotes, newGrantNote(source, amount))
	}
}

func newGrantNote(source grantSource, amount fxp.Int) grantNote {
	return grantNote{owner: source.owner, source: grantSourceName(source), amount: amount}
}

// pointsExcept returns the points grants gave the item, leaving out those of the grants carried by exclude, which may
// be nil to leave out none. The prerequisites of a carrier are judged this way, so that an item never meets its own
// prerequisites with what it grants.
func (s *grantState) pointsExcept(exclude any) fxp.Int {
	total := s.points
	if exclude != nil {
		for _, note := range s.pointNotes {
			if note.owner == exclude {
				total -= note.amount
			}
		}
	}
	return total
}

// applyLevelNotes returns the level with what the grants made of it, leaving out what the grants carried by exclude
// gave (see pointsExcept): the levels merged into it are added, and then each level a Keep Highest grant keeps it at
// replaces it when better, which for a disadvantage is lower. The level a skill or spell is kept at is read here from
// the points of its tier by levelAt, which a trait, whose grants keep levels outright, leaves nil, so that it sees the
// bonuses and penalties the item's own level does, which may not have been in place when the grants were applied. The
// notes of the grants that changed the level go into the tooltip, which may be nil.
func (s *grantState) applyLevelNotes(exclude any, level fxp.Int, tooltip *xbytes.InsertBuffer, levelAt func(points fxp.Int) fxp.Int) fxp.Int {
	for _, note := range s.levelNotes {
		if !note.keep && note.owner != exclude {
			level += note.amount
			addGrantNoteToTooltip(note, note.amount, tooltip)
		}
	}
	for _, note := range s.levelNotes {
		if !note.keep || note.owner == exclude {
			continue
		}
		kept := note.amount
		if note.tier {
			if levelAt == nil {
				continue
			}
			if kept = levelAt(note.amount); kept == fxp.Min {
				continue
			}
		}
		diff := kept - level
		if s.keepLower {
			diff = diff.Min(0)
		} else {
			diff = diff.Max(0)
		}
		if diff != 0 {
			level += diff
			addGrantNoteToTooltip(note, diff, tooltip)
		}
	}
	return level
}

func (s *grantState) hash(h hash.Hash) {
	xhash.Bool(h, s.inEffect)
	xhash.Bool(h, s.keepLower)
	hashGrantNotes(h, s.levelNotes)
	hashGrantNotes(h, s.pointNotes)
}

func hashGrantNotes(h hash.Hash, notes []grantNote) {
	xhash.Num64(h, uint64(len(notes)))
	for _, note := range notes {
		xhash.Num64(h, note.amount)
		xhash.Bool(h, note.keep)
		xhash.Bool(h, note.tier)
	}
}

// addGrantNoteToTooltip adds the note to the tooltip, which may be nil, with the amount it changed the item by.
func addGrantNoteToTooltip(note grantNote, amount fxp.Int, buffer *xbytes.InsertBuffer) {
	if buffer == nil {
		return
	}
	buffer.WriteByte('\n')
	buffer.WriteString(note.source)
	buffer.WriteString(" [")
	buffer.WriteString(amount.StringWithSign())
	buffer.WriteByte(']')
}

// applyPoints returns the points with those grants gave a skill or spell added, leaving out what the grants carried by
// exclude gave (see pointsExcept), from the tooltip, which may be nil, as from the total.
func (s *grantState) applyPoints(exclude any, points fxp.Int, tooltip *xbytes.InsertBuffer) fxp.Int {
	granted := s.pointsExcept(exclude)
	if granted == 0 {
		return points
	}
	for _, note := range s.pointNotes {
		if exclude == nil || note.owner != exclude {
			addGrantNoteToTooltip(note, note.amount, tooltip)
		}
	}
	return (points + granted).Max(0)
}

// applyLevels returns the level with what the grants made of a skill's or spell's level (see applyLevelNotes), unless
// it has none, leaving out what the grants carried by exclude gave. levelAt returns the level the item would have with
// the given points in place of its own, which a Keep Highest grant's level is read with.
func (s *grantState) applyLevels(exclude any, level Level, levelAt func(points fxp.Int) fxp.Int) Level {
	if len(s.levelNotes) == 0 || level.Level == fxp.Min {
		return level
	}
	var tooltip xbytes.InsertBuffer
	tooltip.WriteString(level.Tooltip)
	result := s.applyLevelNotes(exclude, level.Level, &tooltip, levelAt)
	level.RelativeLevel += result - level.Level
	level.Level = result
	level.Tooltip = tooltip.String()
	return level
}
