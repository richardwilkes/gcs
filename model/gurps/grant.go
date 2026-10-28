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
	"maps"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantadjust"
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
	// Amount is, as Adjust says, the level the grant sets, the levels it adds or the points it adds. A trait's levels
	// may be fractional. Levels added to a skill or spell not yet on the sheet are priced in points: level 1 is what 1
	// point buys, level 2 what 2 points buy and level n above that what 4×(n-2) points buy, three times as many for a
	// wildcard, while a technique or Ritual Magic spell, whose levels count from its default, takes 1 point a level, plus
	// 1 for a Hard one. A negative amount added only takes levels or points away from an item already present, and a
	// level set below 1 does nothing.
	Amount fxp.Int `json:"amount,omitzero"`
	// Adjust is how Amount adjusts the item; see EffectiveAdjust.
	Adjust grantadjust.Kind `json:"adjust,omitzero"`
	// Merge is what a level set does to an item already present; see EffectiveMerge. Levels or points are always added
	// to one.
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
	// judged is the state the item has as the prerequisites of judgedFor, what carries the grant that gives the item
	// its details this pass, see it, when another grant in effect would give them in that one's stead; see viewFor.
	// detail is, in a judged state, the grant giving the item its details.
	judged    *grantState
	judgedFor any
	detail    *Grant
	// levelNotes and pointNotes are what each grant gave the item: the levels merged into it or kept for it, and a
	// granted skill's or spell's own points plus any merged into it, of which points is the total. ownPoints is what of
	// those a skill or spell counts as its own when the other grants merge into it: for a granted one, the points the
	// grant in effect gives it standing alone (see grantPlan.standalonePoints and Grant.pointsBeyondDefault), and for
	// any, those the grants merged into it while it had no level of its own gave it (see Entity.mergeGrant).
	levelNotes []grantNote
	pointNotes []grantNote
	points     fxp.Int
	ownPoints  fxp.Int
	// unreached is set when the grant that put a skill or spell on the sheet is in effect but gives it nothing, since
	// the level it sets is below what the points of its first level buy (see grantPlan.reaches and firstLevelText),
	// which leaves the item out of effect unless another grant stands in for that one, and stays set while one does.
	// unresolved is set along with it when the reason is rather that the sheet can't resolve the attribute or skill the
	// item's level rests on.
	unreached  bool
	unresolved bool
	// unreachedPoints, set along with unreached while no grant stands in, prices what the grant that put the item on
	// the sheet gives it as the prerequisites being judged see the sheet, which may give it something; see
	// inEffectAsJudged.
	unreachedPoints *grantReprice
}

// grantLeveler is what a grant needs of its skill or spell to price the amount it gives.
type grantLeveler interface {
	// pointsForLevels returns the points that buy the item the given number of levels standing alone, or zero for fewer
	// than 1: the tier the levels make (see grantTierPoints), or for a technique or Ritual Magic spell, whose levels
	// count from its default, those levels above it (see techniquePoints).
	pointsForLevels(levels fxp.Int) fxp.Int
	// pointsForBaseLevel returns the fewest points that buy the item the level, before bonuses, as the sheet sees it,
	// or zero when none do.
	pointsForBaseLevel(e *Entity, level fxp.Int) fxp.Int
	// baseUnresolved returns true if the sheet can't resolve what the item's level rests on, the attribute or, when
	// restsOnSkill says so, the skill, which leaves pointsForBaseLevel no level to price from.
	baseUnresolved(e *Entity) bool
	// restsOnSkill returns true if the item's level rests on a skill: it is a technique whose default is a skill, or
	// the parry or block of one, or a Ritual Magic spell.
	restsOnSkill() bool
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
// a keep note, which a Keep Highest or Keep Lowest grant leaves, is the level the grant keeps the item at rather than
// an addition, or, with points set, the points it keeps a skill or spell at, whose level is read along with the item's
// own so that both see the same bonuses. lower is set for a Keep Lowest grant, which keeps the lower of the two levels
// rather than the higher; see grantState.applyLevelNotes. reprice is set for the points a level set gave a skill or
// spell; see amountFor.
type grantNote struct {
	owner   any
	source  string
	amount  fxp.Int
	keep    bool
	points  bool
	lower   bool
	reprice *grantReprice
}

// amountFor returns the note's amount as the prerequisites of exclude see the sheet, which for the points a level set
// gave are priced afresh; see grantReprice.pointsFor.
func (n *grantNote) amountFor(exclude any) fxp.Int {
	if n.reprice == nil {
		return n.amount
	}
	return n.reprice.pointsFor(exclude)
}

// grantReprice prices afresh the points a level set gives a skill or spell, which applyGrants prices against the sheet
// as a whole, as the prerequisites of an item carrying a grant see the sheet, without what that item's grants give the
// attribute or skill the level rests on (see Entity.JudgingPrereqsOf), so that they see the item at the level set,
// which the grant setting it gives whether or not those grants are in effect. owner carries the grant, points are the
// points as priced against the sheet as a whole, and busy is set while price runs.
type grantReprice struct {
	entity *Entity
	owner  any
	price  func() fxp.Int
	points fxp.Int
	busy   bool
}

// pointsFor returns the points as the prerequisites of exclude, or when it is nil those being judged, see the sheet:
// priced afresh when exclude carries a grant, but not the one setting the level, and otherwise, or while they are
// priced afresh, as for a level set on an item resting on itself, as priced against the sheet as a whole. They are
// priced once for each time those prerequisites are judged; see Entity.judgedPoints.
func (r *grantReprice) pointsFor(exclude any) fxp.Int {
	e := r.entity
	if exclude = e.judgingOr(exclude); r.busy || exclude == r.owner || !e.carriesGrant(exclude) {
		return r.points
	}
	points := r.points
	e.JudgingPrereqsOf(exclude, func() {
		if cached, ok := e.judgedPoints[r]; ok {
			points = cached
			return
		}
		r.busy = true
		defer func() { r.busy = false }()
		points = r.price()
		if e.judgedPoints == nil {
			e.judgedPoints = make(map[*grantReprice]fxp.Int)
		}
		e.judgedPoints[r] = points
	})
	return points
}

// itemLabel returns the name of the item the grant gives, with the amount it gives. replacements are those of what
// carries the grant; see itemNameWith.
func (g *Grant) itemLabel(replacements map[string]string) string {
	name := g.itemNameWith(replacements)
	amount := g.amount()
	switch {
	case !g.Leveled():
		return name
	case g.Type == feature.GrantTrait:
		return name + " " + amount.String()
	}
	switch g.EffectiveAdjust() {
	case grantadjust.AddPoints:
		return i18n.Text("%s with %s", name, pointsText(amount, nil))
	case grantadjust.SetLevel:
		return i18n.Text("%s at skill level %s", name, amount.String())
	default:
		return i18n.Text("%s at level %s", name, amount.String())
	}
}

// amount returns the amount the grant gives: as set for a trait, whose levels may be fractional, and for points, and
// whole otherwise.
func (g *Grant) amount() fxp.Int {
	if g.Type == feature.GrantTrait || g.EffectiveAdjust() == grantadjust.AddPoints {
		return g.Amount
	}
	return g.Amount.Floor()
}

// amountPhrase returns how a sentence names the amount the grant gives its item, "at level N", "with N points" or "at
// skill level N", with the number emphasized by em when that is not nil.
func (g *Grant) amountPhrase(em func(string) string) string {
	amount := g.amount()
	switch {
	case g.EffectiveAdjust() == grantadjust.AddPoints:
		return i18n.Text("with %s", pointsText(amount, em))
	case g.setsSkillLevel():
		return i18n.Text("at skill level %s", emphasized(amount.String(), em))
	default:
		return i18n.Text("at level %s", emphasized(amount.String(), em))
	}
}

// adds returns true if the grant adds its amount, levels or points, to its item when already present, rather than
// setting the item's level; see EffectiveAdjust.
func (g *Grant) adds() bool {
	return g.Leveled() && g.EffectiveAdjust() != grantadjust.SetLevel
}

// setsSkillLevel returns true if the grant sets the level of a skill or spell, which the sheet prices in points.
func (g *Grant) setsSkillLevel() bool {
	return g.Type != feature.GrantTrait && g.EffectiveAdjust() == grantadjust.SetLevel
}

// emphasized returns the text emphasized by em, or as it is when em is nil.
func emphasized(text string, em func(string) string) string {
	if em == nil {
		return text
	}
	return em(text)
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
	adds := g.adds()
	if (adds || merge != grantmerge.Ignore) && g.mergesNothing() {
		switch {
		case g.EffectiveAdjust() == grantadjust.AddPoints:
			return i18n.Text("Does not change %s, already present, since points too few to buy a level grant nothing",
				name)
		case g.setsSkillLevel():
			return i18n.Text("Does not change %s, already present, since a skill level below 1 grants nothing", name)
		default:
			return i18n.Text("Does not change %s, already present, since a level below 1 grants nothing", name)
		}
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
	amount := g.amount()
	switch {
	case adds && amount < 0:
		count := g.countText(amount.Abs(), nil)
		return pick(i18n.Text("Removes %s from %s, already present", count, name),
			i18n.Text("Would remove %s from %s, already present, but the grant is not in effect", count, name),
			i18n.Text(
				"Would remove %s from %s, already present, but the prerequisites of what carries the grant are not met",
				count, name,
			))
	case adds:
		count := g.countText(amount, nil)
		return pick(i18n.Text("Adds %s to %s, already present", count, name),
			i18n.Text("Would add %s to %s, already present, but the grant is not in effect", count, name),
			i18n.Text(
				"Would add %s to %s, already present, but the prerequisites of what carries the grant are not met",
				count, name,
			))
	case merge == grantmerge.KeepHighest:
		phrase := g.amountPhrase(nil)
		return pick(i18n.Text("Keeps the higher level of %s, as it is or %s", name, phrase),
			i18n.Text("Would keep the higher level of %s, as it is or %s, but the grant is not in effect", name, phrase),
			i18n.Text(
				"Would keep the higher level of %s, as it is or %s, but the prerequisites of what carries the grant are not met",
				name, phrase,
			))
	case merge == grantmerge.KeepLowest:
		phrase := g.amountPhrase(nil)
		return pick(i18n.Text("Keeps the lower level of %s, as it is or %s", name, phrase),
			i18n.Text("Would keep the lower level of %s, as it is or %s, but the grant is not in effect", name, phrase),
			i18n.Text(
				"Would keep the lower level of %s, as it is or %s, but the prerequisites of what carries the grant are not met",
				name, phrase,
			))
	default:
		return i18n.Text("Does not grant %s, since it is already present", label)
	}
}

// idleSummary returns why the grant, whose plan has no item (see grantPlan), does nothing, for the tooltip of what
// carries it: its item can't be granted, or is absent while the grant can't put it on the sheet. replacements are
// those of what carries the grant; see itemNameWith.
func (g *Grant) idleSummary(replacements map[string]string) string {
	name := g.itemNameWith(replacements)
	switch {
	case !g.itemGrantable():
		return i18n.Text("Does not grant %s, since a container can't be granted", name)
	case g.removes():
		// A negative amount would take the levels or points it names away from the item, were it present.
		return i18n.Text("Removes nothing from %s, since it is not present", name)
	}
	label := g.itemLabel(replacements)
	switch {
	case g.EffectiveAdjust() == grantadjust.AddPoints:
		return i18n.Text("Does not grant %s, since points too few to buy a level grant nothing", label)
	case g.setsSkillLevel():
		return i18n.Text("Does not grant %s, since a skill level below 1 grants nothing", label)
	default:
		return i18n.Text("Does not grant %s, since a level below 1 grants nothing", label)
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
// item of the grant's kind when item is nil. A blank technique or Ritual Magic spell, as NewTechnique and
// NewRitualMagicSpell make one, is passed in as item, with an empty library file. A kind of adjustment or a merge mode
// that means nothing for the item falls back as EffectiveAdjust and EffectiveMerge say. A grant that sets a leveled
// trait's level sets it to the trait's level, while any other keeps its amount. Any grants the item or its modifiers
// carry are left out of the copy, as SyncItemWithSource leaves them out, since a granted item can't grant anything.
func (g *Grant) SetItem(item any, from LibraryFile) {
	g.Trait, g.Skill, g.Spell = nil, nil, nil
	switch one := item.(type) {
	case *Trait:
		g.Trait = one.Clone(from, nil, nil, Reference)
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
	g.Adjust = g.EffectiveAdjust()
	g.Merge = g.EffectiveMerge()
	if g.Adjust == grantadjust.SetLevel && g.Trait != nil && g.Trait.IsLeveled() && g.Trait.Levels > 0 {
		g.Amount = g.Trait.Levels
	}
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

// CarriesGrants returns true if the trait, skill, spell or equipment, or a modifier of the trait or equipment, carries
// a grant with an item chosen, and so may have put rows on the sheet, in its own list or another.
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

// NewGrant creates a new Grant of the given type, which must be one of the grant feature types: one that sets a trait's
// level, or adds levels to a skill or spell.
func NewGrant(featureType feature.Type) *Grant {
	g := &Grant{
		Type:   featureType,
		Amount: fxp.One,
	}
	if featureType == feature.GrantTrait {
		g.Adjust = grantadjust.SetLevel
	}
	return g
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

// RenewKey gives the grant's copy of its item a new ID of the kind the item it makes has (see madeKind), and so the
// grant a new key, as a duplicate of a grant needs.
func (g *Grant) RenewKey() {
	if g.Trait != nil {
		g.Trait.TID = tid.MustNewTID(madeKind(g.Trait))
	}
	if g.Skill != nil {
		g.Skill.TID = tid.MustNewTID(madeKind(g.Skill))
	}
	if g.Spell != nil {
		g.Spell.TID = tid.MustNewTID(madeKind(g.Spell))
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

// Leveled returns true if the amount means anything for the granted item: always for a skill or spell, and for a trait
// only when it has levels.
func (g *Grant) Leveled() bool {
	if g.Type == feature.GrantTrait {
		return g.Trait != nil && g.Trait.IsLeveled()
	}
	return true
}

// itemGrantable returns true if the grant's item is one a grant can put on the sheet: anything but a container, which
// can't be dropped on a grant or added as one, so that a grant of one from a hand-edited file does nothing.
func (g *Grant) itemGrantable() bool {
	switch g.Type {
	case feature.GrantTrait:
		return g.Trait != nil && !g.Trait.Container()
	case feature.GrantSkill:
		return g.Skill != nil && !g.Skill.Container()
	case feature.GrantSpell:
		return g.Spell != nil && !g.Spell.Container()
	default:
		return false
	}
}

// putsItemOnSheet returns true if the grant puts its item on the sheet when the item is absent: a trait without levels
// always does, a grant of points only when they buy a level, and anything else only at level 1 or more, since a
// negative amount only takes away from an item already present. Whether a level set is high enough to give the item
// anything depends on the sheet, which this leaves to grantPlan.reaches, so that the rows on a sheet are decided by its
// data alone.
func (g *Grant) putsItemOnSheet() bool {
	switch {
	case !g.Leveled():
		return true
	case g.EffectiveAdjust() == grantadjust.AddPoints:
		return g.pointsOn(nil, g.leveler()) > 0
	default:
		return g.amount() >= fxp.One
	}
}

// removes returns true if the grant only takes away from its item when already present: it adds a negative amount of
// levels or points (see adds). A level set below 1 takes nothing away.
func (g *Grant) removes() bool {
	return g.adds() && g.amount() < 0
}

// mergesNothing returns true if the grant, finding its item already present, changes nothing about it: a leveled
// grant that can't put the item on the sheet gives nothing, as Describe says of it, unless it takes away the levels or
// points it names (see removes) or adds points too few to buy a level, which still add to those the item has.
func (g *Grant) mergesNothing() bool {
	if !g.Leveled() || g.putsItemOnSheet() || g.removes() {
		return false
	}
	return g.EffectiveAdjust() != grantadjust.AddPoints || g.amount() <= 0
}

// leveler returns the grant's copy of its skill or spell, or nil for a trait grant or one with no item chosen. The
// placeholders the copy leaves to what carries the grant are left unanswered; see levelerWith.
func (g *Grant) leveler() grantLeveler {
	switch {
	case g.Type == feature.GrantSkill && g.Skill != nil:
		return g.Skill
	case g.Type == feature.GrantSpell && g.Spell != nil:
		return g.Spell
	default:
		return nil
	}
}

// levelerWith is leveler with the placeholders the copy leaves to what carries the grant answered by answers, those of
// the carrier (see answers), as the item the grant makes answers them. Only an item whose level rests on a skill (see
// grantLeveler.restsOnSkill) needs them, since that skill may be named by one, so only it is copied to hold them.
func (g *Grant) levelerWith(answers map[string]string) grantLeveler {
	item := g.leveler()
	if len(answers) == 0 || item == nil || !item.restsOnSkill() {
		return item
	}
	switch one := item.(type) {
	case *Skill:
		shallow := *one
		shallow.Replacements = mergeReplacements(maps.Clone(one.Replacements), answers)
		return &shallow
	case *Spell:
		shallow := *one
		shallow.Replacements = mergeReplacements(maps.Clone(one.Replacements), answers)
		return &shallow
	default:
		return item
	}
}

// pointsOn returns the points the grant gives the skill or spell standing alone, or merged into one with no level of
// its own: those its levels cost (see grantLeveler.pointsForLevels), those it adds when they buy the first level, or
// the fewest that buy the level it sets as the sheet sees the item, of which a nil sheet or a level below 1 gets none.
// A nil item gets none.
func (g *Grant) pointsOn(e *Entity, item grantLeveler) fxp.Int {
	if item == nil {
		return 0
	}
	amount := g.amount()
	switch g.EffectiveAdjust() {
	case grantadjust.AddPoints:
		if amount >= item.pointsForLevels(fxp.One) {
			return amount
		}
		return 0
	case grantadjust.SetLevel:
		if e == nil || amount < fxp.One {
			return 0
		}
		return item.pointsForBaseLevel(e, amount)
	default:
		return item.pointsForLevels(amount)
	}
}

// baseUnresolved returns true if the grant sets the level of the skill or spell and the sheet can't resolve what that
// level rests on (see grantLeveler.baseUnresolved), which leaves the level set no points. A nil item has nothing to
// resolve.
func (g *Grant) baseUnresolved(e *Entity, item grantLeveler) bool {
	return item != nil && g.setsSkillLevel() && item.baseUnresolved(e)
}

// pointsBeyondDefault returns the points the grant gives the skill or spell on the sheet that holder is, which for a
// level set on a skill are less those the default it uses stands for, since its level adds them (see
// CalculateSkillLevel), so that the skill has the level set, or the default's when that is higher, rather than one
// raised by both.
func (g *Grant) pointsBeyondDefault(points fxp.Int, holder grantHolder) fxp.Int {
	s, ok := holder.(*Skill)
	if !ok || g.EffectiveAdjust() != grantadjust.SetLevel || s.WildcardDifficulty() || s.DefaultedFrom == nil ||
		s.DefaultedFrom.Points <= 0 {
		return points
	}
	return (points - s.DefaultedFrom.Points).Max(0)
}

// firstLevelText names the lowest level points buy the skill or spell, the one the points of its first level buy (see
// grantLeveler.pointsForLevels), below which a level set gives it nothing. A nil item is taken as one whose first level
// costs 1 point.
func firstLevelText(item grantLeveler) string {
	points := fxp.One
	if item != nil {
		points = item.pointsForLevels(fxp.One)
	}
	if points == fxp.One {
		return i18n.Text("what 1 point buys")
	}
	return i18n.Text("what %s points buy", points.String())
}

// unresolvedText names what the sheet lacks when it can't resolve what the level of the skill or spell rests on (see
// grantLeveler.baseUnresolved): the skill for one whose level rests on a skill (see grantLeveler.restsOnSkill), and the
// attribute otherwise. A nil item is taken as one resting on an attribute.
func unresolvedText(item grantLeveler) string {
	if item != nil && item.restsOnSkill() {
		return i18n.Text("the sheet lacks the skill its level rests on")
	}
	return i18n.Text("the sheet lacks the attribute its level rests on")
}

// AllowedAdjusts returns the kinds of adjustment that mean something for the granted item, the first of them being the
// one EffectiveAdjust falls back to: for a trait, setting its level or adding levels, since its points follow from its
// level, and for a skill or spell, all of them.
func (g *Grant) AllowedAdjusts() []grantadjust.Kind {
	switch g.Type {
	case feature.GrantTrait:
		return []grantadjust.Kind{grantadjust.SetLevel, grantadjust.AddLevels}
	case feature.GrantSkill, feature.GrantSpell:
		return grantadjust.Kinds
	default:
		return nil
	}
}

// EffectiveAdjust returns the kind of adjustment, treating one that means nothing for the granted item as the first of
// those that do (see AllowedAdjusts): setting a trait's level, or adding levels to a skill or spell.
func (g *Grant) EffectiveAdjust() grantadjust.Kind {
	allowed := g.AllowedAdjusts()
	if adjust := g.Adjust.EnsureValid(); slices.Contains(allowed, adjust) {
		return adjust
	}
	if len(allowed) != 0 {
		return allowed[0]
	}
	return grantadjust.AddLevels
}

// AllowedMerges returns the merge modes that mean something for the granted item: all of them for a grant that sets
// its item's level, and Ignore alone otherwise, since levels or points added are always added to an item already
// present, and a trait without levels has none to set.
func (g *Grant) AllowedMerges() []grantmerge.Mode {
	if !g.Leveled() || g.EffectiveAdjust() != grantadjust.SetLevel {
		return []grantmerge.Mode{grantmerge.Ignore}
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
// put on the sheet. item is the skill or spell, or nil for a trait.
func grantedByText(link *GrantedBy, state *grantState, item grantLeveler) string {
	if link == nil {
		return ""
	}
	if !xreflect.IsNil(state.standIn.owner) {
		if state.unreached {
			return i18n.Text("Granted by %s, standing in for %s, whose grant gives it nothing",
				grantSourceName(state.standIn), grantSourceName(state.source))
		}
		return i18n.Text("Granted by %s, standing in for %s, whose grant is not in effect", grantSourceName(state.standIn),
			grantSourceName(state.source))
	}
	text := i18n.Text("Granted by %s", grantSourceName(state.source))
	switch {
	case state.unresolved:
		text += "\n" + i18n.Text("The grant gives it nothing, since %s.", unresolvedText(item))
	case state.unreached:
		text += "\n" + i18n.Text("The grant gives it nothing, since the level it sets is below %s.",
			firstLevelText(item))
	case !state.inEffect:
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
		if g.Skill.IsTechnique() {
			item = i18n.Text("the technique %s", em(name))
		} else {
			item = i18n.Text("the skill %s", em(name))
		}
	case feature.GrantSpell:
		item = i18n.Text("the spell %s", em(name))
	default:
		item = i18n.Text("the trait %s", em(name))
	}
	amount := g.amount()
	if g.Leveled() && !g.putsItemOnSheet() {
		// The grant puts nothing on the sheet, and a negative amount only takes away from an item present.
		var text string
		switch adjust := g.EffectiveAdjust(); {
		case g.removes():
			text = i18n.Text("Removes %s from %s when already present", g.countText(amount.Abs(), em), item)
		case !g.mergesNothing():
			text = i18n.Text("Adds %s to %s when already present, though too few to grant it", pointsText(amount, em),
				item)
		case adjust == grantadjust.AddPoints:
			text = i18n.Text("Would grant %s with %s, but points too few to buy a level grant nothing", item,
				pointsText(amount, em))
		case g.setsSkillLevel():
			text = i18n.Text("Would grant %s at skill level %s, but a skill level below 1 grants nothing", item,
				em(amount.String()))
		default:
			text = i18n.Text("Would grant %s at level %s, but a level below 1 grants nothing", item, em(amount.String()))
		}
		return describeSwitchable(g.Switchable, text)
	}
	text := i18n.Text("Grants %s", item)
	if g.Leveled() {
		text += " " + g.amountPhrase(em)
	}
	switch merge := g.EffectiveMerge(); {
	case g.adds():
		text += i18n.Text(", or adds %s to it when already present", g.countText(amount, em))
	case merge == grantmerge.KeepHighest:
		text += i18n.Text(", or keeps the higher of the two levels when already present")
	case merge == grantmerge.KeepLowest:
		text += i18n.Text(", or keeps the lower of the two levels when already present")
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
	text := emphasized(levels.String(), em)
	if levels == fxp.One {
		return i18n.Text("%s level", text)
	}
	return i18n.Text("%s levels", text)
}

// pointsText returns the count of points as words, "1 point" or "N points", with the number emphasized by em when that
// is not nil.
func pointsText(points fxp.Int, em func(string) string) string {
	text := emphasized(points.String(), em)
	if points == fxp.One {
		return i18n.Text("%s point", text)
	}
	return i18n.Text("%s points", text)
}

// countText returns the count of what the grant's amount adjusts its item by, points for a grant of points and levels
// otherwise; see pointsText and levelsText.
func (g *Grant) countText(count fxp.Int, em func(string) string) string {
	if g.EffectiveAdjust() == grantadjust.AddPoints {
		return pointsText(count, em)
	}
	return levelsText(count, em)
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
	xhash.Num64(h, g.Amount)
	xhash.Num8(h, g.Adjust)
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

func (s *grantState) reset() {
	*s = grantState{source: s.source}
}

// giver returns what carries the grant giving the item its details this pass: the one standing in, or else the one
// that put the item on the sheet, or nil while neither is in effect.
func (s *grantState) giver() any {
	switch {
	case !xreflect.IsNil(s.standIn.owner):
		return s.standIn.owner
	case s.inEffect:
		return s.source.owner
	default:
		return nil
	}
}

// viewFor returns the state as the prerequisites of exclude see it: the judged one when exclude carries the grant
// giving the item its details and another grant would give them in its stead, so that those prerequisites are never
// met through the item its own grant makes, and this one otherwise.
func (s *grantState) viewFor(exclude any) *grantState {
	if s.judged != nil && exclude != nil && exclude == s.judgedFor {
		return s.judged
	}
	return s
}

// prereqsChecked returns true if a grant giving the item its details checks its prerequisites, as the item stands or as
// the prerequisites of the carrier it has a judged state for see it (see viewFor).
func (s *grantState) prereqsChecked() bool {
	return !s.bypass || (s.judged != nil && !s.judged.bypass)
}

func (s *grantState) addLevels(source grantSource, amount fxp.Int) {
	if amount != 0 {
		s.levelNotes = append(s.levelNotes, newGrantNote(source, amount))
	}
}

// keepLevel records that a Keep Highest grant keeps the item at the level when it is higher than the item's own, or,
// with lower set, that a Keep Lowest grant keeps it there when it is lower.
func (s *grantState) keepLevel(source grantSource, level fxp.Int, lower bool) {
	note := newGrantNote(source, level)
	note.keep = true
	note.lower = lower
	s.levelNotes = append(s.levelNotes, note)
}

// keepPoints is keepLevel for a skill or spell kept at the level the points give it, which is read along with the
// item's own level; see applyLevelNotes. Zero points keep it at no level. reprice, set for a level set, prices the
// points afresh for the prerequisites being judged.
func (s *grantState) keepPoints(source grantSource, points fxp.Int, lower bool, reprice *grantReprice) {
	note := newGrantNote(source, points)
	note.keep = true
	note.points = true
	note.lower = lower
	note.reprice = reprice
	s.levelNotes = append(s.levelNotes, note)
}

// addPoints records that a grant gave the skill or spell the points. reprice, set for a level set, prices them afresh
// for the prerequisites being judged, and with it the note is kept though the points are zero.
func (s *grantState) addPoints(source grantSource, amount fxp.Int, reprice *grantReprice) {
	if amount != 0 || reprice != nil {
		s.points += amount
		note := newGrantNote(source, amount)
		note.reprice = reprice
		s.pointNotes = append(s.pointNotes, note)
	}
}

// inEffectAsJudged returns true if the item is in effect, or, as the prerequisites being judged see the sheet, the
// grant whose level set puts it there gives it something, though nothing priced against the sheet as a whole; see
// unreachedPoints.
func (s *grantState) inEffectAsJudged() bool {
	return s.inEffect || (s.unreachedPoints != nil && s.unreachedPoints.pointsFor(nil) > 0)
}

func newGrantNote(source grantSource, amount fxp.Int) grantNote {
	return grantNote{owner: source.owner, source: grantSourceName(source), amount: amount}
}

// pointsExcept returns the points grants gave the item, as the prerequisites of exclude see it (see viewFor), leaving
// out those of the grants carried by exclude, which may be nil to leave out none. The prerequisites of a carrier are
// judged this way, so that an item never meets its own prerequisites with what it grants, and see the points of a level
// set priced afresh, including those of one that gives nothing priced against the sheet as a whole (see
// unreachedPoints).
func (s *grantState) pointsExcept(exclude any) fxp.Int {
	s = s.viewFor(exclude)
	if exclude == nil {
		return s.points
	}
	var total fxp.Int
	for i := range s.pointNotes {
		if note := &s.pointNotes[i]; note.owner != exclude {
			total += note.amountFor(exclude)
		}
	}
	if s.unreachedPoints != nil {
		total += s.unreachedPoints.pointsFor(exclude)
	}
	return total
}

// applyLevelNotes returns the level with what the grants made of it, leaving out what the grants carried by exclude
// gave (see pointsExcept): the levels merged into it are added, and then, in the order the grants were merged, each
// level a Keep Highest grant keeps it at replaces it when higher, and each a Keep Lowest grant keeps it at when lower,
// so that a Keep Lowest grant, merged after any Keep Highest one (see comparePlanting), caps what that raised. The
// level a skill or spell is kept at is read here from the points kept by levelAt, which a trait, whose grants keep
// levels outright, leaves nil, so that it sees the bonuses and penalties the item's own level does, which may not have
// been in place when the grants were applied. The notes of the grants that changed the level go into the tooltip,
// which may be nil.
func (s *grantState) applyLevelNotes(exclude any, level fxp.Int, tooltip *xbytes.InsertBuffer, levelAt func(points fxp.Int) fxp.Int) fxp.Int {
	s = s.viewFor(exclude)
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
		kept := note.amountFor(exclude)
		if note.points {
			// No points keep the item at no level.
			if levelAt == nil || kept == 0 {
				continue
			}
			if kept = levelAt(kept); kept == fxp.Min {
				continue
			}
		}
		diff := kept - level
		if note.lower {
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
	xhash.Bool(h, s.unreached)
	xhash.Bool(h, s.unresolved)
	hashGrantNotes(h, s.levelNotes)
	hashGrantNotes(h, s.pointNotes)
}

func hashGrantNotes(h hash.Hash, notes []grantNote) {
	xhash.Num64(h, uint64(len(notes)))
	for _, note := range notes {
		xhash.Num64(h, note.amount)
		xhash.Bool(h, note.keep)
		xhash.Bool(h, note.points)
		xhash.Bool(h, note.lower)
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
	s = s.viewFor(exclude)
	granted := s.pointsExcept(exclude)
	if granted == 0 {
		return points
	}
	for _, note := range s.pointNotes {
		if amount := note.amountFor(exclude); amount != 0 && (exclude == nil || note.owner != exclude) {
			addGrantNoteToTooltip(note, amount, tooltip)
		}
	}
	return (points + granted).Max(0)
}

// applyLevels returns the level with what the grants made of a skill's or spell's level (see applyLevelNotes),
// leaving out what the grants carried by exclude gave. Levels taken away leave it no lower than 0, or than it was, when
// it was already below that. levelAt returns the level the item would have with the given points in place of its own,
// which the level a Keep Highest or Keep Lowest grant keeps is read with. An item with no level, as the points other
// grants take away may leave one a grant keeps a level for, has none to add levels to or take them from, but a Keep
// Highest grant still keeps it at its level, as it does any lower one, which then reads as the level kept, bonuses
// and all; see keepLevelOfNone.
func (s *grantState) applyLevels(exclude any, level Level, levelAt func(points fxp.Int) Level) Level {
	if len(s.viewFor(exclude).levelNotes) == 0 {
		return level
	}
	if level.Level == fxp.Min {
		return s.keepLevelOfNone(exclude, level, levelAt)
	}
	var tooltip xbytes.InsertBuffer
	tooltip.WriteString(level.Tooltip)
	result := s.applyLevelNotes(exclude, level.Level, &tooltip, func(points fxp.Int) fxp.Int {
		return levelAt(points).Level
	}).Max(level.Level.Min(0))
	level.RelativeLevel += result - level.Level
	level.Level = result
	level.Tooltip = tooltip.String()
	return level
}

// applyLevelsWithinLimit is applyLevels for a technique with a limit or a Ritual Magic spell, which no points raise
// above its limit: a level the grants raise above the given one is held to the highest level points give it (see
// levelAt), which the limit caps. A level a grant keeps is read with levelAt, so only levels added can overshoot.
func (s *grantState) applyLevelsWithinLimit(exclude any, level Level, levelAt func(points fxp.Int) Level) Level {
	result := s.applyLevels(exclude, level, levelAt)
	if result.Level <= level.Level {
		return result
	}
	if most := levelAt(fxp.MaxBasePoints).Level; most != fxp.Min && result.Level > most {
		result.RelativeLevel -= result.Level - most
		result.Level = most
	}
	return result
}

// keepLevelOfNone is applyLevels for an item with no level: the levels added or taken away are passed over, and the
// keep notes are applied in order, as applyLevelNotes applies them, with the first Keep Highest one whose level exists
// giving the item that level, which those after it compare with as usual.
func (s *grantState) keepLevelOfNone(exclude any, level Level, levelAt func(points fxp.Int) Level) Level {
	var tooltip xbytes.InsertBuffer
	tooltip.WriteString(level.Tooltip)
	for _, note := range s.viewFor(exclude).levelNotes {
		if !note.keep || !note.points || note.owner == exclude || (note.lower && level.Level == fxp.Min) {
			continue
		}
		points := note.amountFor(exclude)
		if points == 0 {
			continue
		}
		kept := levelAt(points)
		if kept.Level == fxp.Min {
			continue
		}
		if level.Level == fxp.Min {
			level = kept
			tooltip.Reset()
			tooltip.WriteString(kept.Tooltip)
			continue
		}
		diff := kept.Level - level.Level
		if note.lower {
			diff = diff.Min(0)
		} else {
			diff = diff.Max(0)
		}
		if diff != 0 {
			level.Level += diff
			level.RelativeLevel += diff
			addGrantNoteToTooltip(note, diff, &tooltip)
		}
	}
	level.Tooltip = tooltip.String()
	return level
}
