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
	"maps"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantmerge"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xreflect"
)

// grantPlan is what reconcileGrants decided for one grant on the sheet: the item it put there, or the item already on
// the sheet that it merges into, or neither.
type grantPlan struct {
	grant  *Grant
	source grantSource
	node   grantHolder
	target grantHolder
}

// grantSource is what carries a grant: the item, and the modifier of it when the grant is the modifier's.
type grantSource struct {
	owner    fmt.Stringer
	subOwner fmt.Stringer
}

// grantHolder is implemented by the kinds of item a grant can put on a sheet.
type grantHolder interface {
	grants() *grantState
	rawName() string
}

// grantNode is the constraint for the kinds of item a grant can put on a sheet.
type grantNode[T Node[T]] interface {
	Node[T]
	Disableable
	grantHolder
	nameable.Accesser
	ApplyNameableKeys(m map[string]string)
}

type grantKey struct {
	grantor tid.TID
	key     tid.TID
}

// IsGranted returns true if the node is a trait, skill or spell a grant put on the sheet, which only changes with its
// grant.
func IsGranted(node any) bool {
	return *grantLink(node) != nil
}

// ReadOnly returns true if a grant put the trait on the sheet, so that it changes only with its grant.
func (t *Trait) ReadOnly() bool { return t.Granted != nil }

// ReadOnly returns true if a grant put the skill on the sheet, so that it changes only with its grant.
func (s *Skill) ReadOnly() bool { return s.Granted != nil }

// ReadOnly returns true if a grant put the spell on the sheet, so that it changes only with its grant.
func (s *Spell) ReadOnly() bool { return s.Granted != nil }

func (t *Trait) grantNotInEffect() bool { return t.Granted != nil && !t.grant.inEffect }
func (s *Skill) grantNotInEffect() bool { return s.Granted != nil && !s.grant.inEffect }
func (s *Spell) grantNotInEffect() bool { return s.Granted != nil && !s.grant.inEffect }

// GrantingItem returns the trait, skill, spell or equipment carrying the grant that put the node on the sheet, or nil
// for a node no grant put there.
func GrantingItem(node any) any {
	if h, ok := node.(grantHolder); ok && IsGranted(node) && !xreflect.IsNil(h.grants().source.owner) {
		return h.grants().source.owner
	}
	return nil
}

func (t *Trait) grants() *grantState { return &t.grant }
func (s *Skill) grants() *grantState { return &s.grant }
func (s *Spell) grants() *grantState { return &s.grant }

func (t *Trait) rawName() string { return t.Name }
func (s *Skill) rawName() string { return s.Name }
func (s *Spell) rawName() string { return s.Name }

// ClassifyingPoints returns the points that decide whether the trait counts as an advantage, perk, disadvantage or
// quirk: its adjusted points, or for a granted trait in effect, which costs nothing, what it would cost if bought.
func (t *Trait) ClassifyingPoints() fxp.Int {
	if t.Granted == nil || t.EffectivelyDisabled() {
		return t.AdjustedPoints(nil)
	}
	return t.pointsWith(t.AllModifiers())
}

// grantLink returns where a trait, skill or spell keeps the link to the grant that put it on the sheet.
func grantLink(node any) **GrantedBy {
	switch n := node.(type) {
	case *Trait:
		return &n.Granted
	case *Skill:
		return &n.Granted
	case *Spell:
		return &n.Granted
	default:
		return new(*GrantedBy)
	}
}

// forEachGrant calls fn with every grant on the sheet, in sheet order, whether in effect or not, along with the ID of
// the item or modifier carrying it and what carries it. Grants carried by an item or modifier the user has disabled,
// or by a granted item, are skipped.
func (e *Entity) forEachGrant(fn func(grantor tid.TID, source grantSource, g *Grant)) {
	each := func(grantor tid.TID, source grantSource, list Features) {
		for _, f := range list {
			if g, ok := f.(*Grant); ok && g.Key() != "" {
				fn(grantor, source, g)
			}
		}
	}
	Traverse(func(t *Trait) bool {
		if t.Granted == nil && !userDisabled(t) {
			if !t.Container() {
				each(t.TID, grantSource{owner: t}, t.Features)
			}
			Traverse(func(mod *TraitModifier) bool {
				if mod.Enabled() {
					each(mod.TID, grantSource{owner: t, subOwner: mod}, mod.Features)
				}
				return false
			}, false, true, t.Modifiers...)
		}
		return false
	}, false, false, e.Traits...)
	Traverse(func(s *Skill) bool {
		if s.Granted == nil && !userDisabled(s) {
			each(s.TID, grantSource{owner: s}, s.Features)
		}
		return false
	}, false, true, e.Skills...)
	Traverse(func(s *Spell) bool {
		if s.Granted == nil && !userDisabled(s) {
			each(s.TID, grantSource{owner: s}, s.Features)
		}
		return false
	}, false, true, e.Spells...)
	Traverse(func(eqp *Equipment) bool {
		if eqp.Enabled() {
			each(eqp.TID, grantSource{owner: eqp}, eqp.Features)
			Traverse(func(mod *EquipmentModifier) bool {
				if mod.Enabled() {
					each(mod.TID, grantSource{owner: eqp, subOwner: mod}, mod.Features)
				}
				return false
			}, false, true, eqp.Modifiers...)
		}
		return false
	}, false, false, e.CarriedEquipment...)
}

// reconcileGrants brings the items grants put on the sheet in line with its grants, before the recalculation passes,
// so that the items on the sheet don't change while they run. A grant whose item is already present merges into it
// instead, as its merge mode says. Whether a grant is in effect is left to each pass. Which item a grant merges into
// is decided by the sheet's data alone: an item the user has disabled counts as absent, but one the sheet disabled
// does not.
func (e *Entity) reconcileGrants() {
	e.grantPlans = e.grantPlans[:0]
	e.grantCarriersJudged = false
	traits := newGrantReconciler(e, &e.Traits)
	skills := newGrantReconciler(e, &e.Skills)
	spells := newGrantReconciler(e, &e.Spells)
	e.forEachGrant(func(grantor tid.TID, source grantSource, g *Grant) {
		switch g.Type {
		case feature.GrantTrait:
			traits.consider(grantor, source, g, g.Trait)
		case feature.GrantSkill:
			if !g.Skill.IsTechnique() {
				skills.consider(grantor, source, g, g.Skill)
			}
		case feature.GrantSpell:
			spells.consider(grantor, source, g, g.Spell)
		}
	})
	traits.finish()
	skills.finish()
	spells.finish()
}

type grantReconciler[T grantNode[T]] struct {
	entity   *Entity
	list     *[]T
	existing map[grantKey][]T
	// present holds the items a grant can merge into by name, and unfiled those not yet filed under it, which only a
	// grant needs. An item can only be the same as one with the same name, so only those need hashing.
	present map[string][]T
	unfiled []T
	kept    map[T]bool
}

func newGrantReconciler[T grantNode[T]](e *Entity, list *[]T) *grantReconciler[T] {
	r := &grantReconciler[T]{
		entity:   e,
		list:     list,
		existing: make(map[grantKey][]T),
		present:  make(map[string][]T),
		kept:     make(map[T]bool),
	}
	Traverse(func(node T) bool {
		if link := *grantLink(node); link != nil {
			k := grantKey{grantor: link.Grantor, key: link.Key}
			r.existing[k] = append(r.existing[k], node)
		} else if !userDisabled(node) {
			r.unfiled = append(r.unfiled, node)
		}
		return false
	}, false, true, *list...)
	return r
}

// presentNamed returns the items present with the same name as the probe.
func (r *grantReconciler[T]) presentNamed(probe T) []T {
	for _, node := range r.unfiled {
		r.present[node.rawName()] = append(r.present[node.rawName()], node)
	}
	r.unfiled = r.unfiled[:0]
	return r.present[probe.rawName()]
}

// userDisabled returns true if the user has disabled the node or one of its parents.
func userDisabled[T grantNode[T]](node T) bool {
	for ; !xreflect.IsNil(node); node = node.Parent() {
		if node.IsDisabled() {
			return true
		}
	}
	return false
}

func (r *grantReconciler[T]) consider(grantor tid.TID, source grantSource, g *Grant, item T) {
	answers := g.answers(source)
	link := &GrantedBy{Grantor: grantor, Key: g.Key(), Hash: g.contentHash(answers)}
	k := grantKey{grantor: link.Grantor, key: link.Key}
	// A copy of a granted item, or a grant duplicated without a new key, can leave several items with the same key.
	// Each grant claims the first one no other grant has.
	var node T
	has := false
	for _, one := range r.existing[k] {
		if !r.kept[one] {
			node = one
			has = true
			break
		}
	}
	probe := node
	var fresh T
	if !has || (*grantLink(node)).Hash != link.Hash {
		fresh = r.makeItem(g, item, link, answers)
		probe = fresh
	}
	h := Hash64(probe)
	for _, candidate := range r.presentNamed(probe) {
		if Hash64(candidate) == h && SameItem(candidate, probe) {
			r.entity.grantPlans = append(r.entity.grantPlans, grantPlan{grant: g, source: source, target: candidate})
			return
		}
	}
	// Only a trait without levels, or a grant of at least level 1, puts its item on the sheet.
	if g.Leveled() && g.Level.Floor() < fxp.One {
		return
	}
	switch {
	case !has:
		*r.list = append(*r.list, fresh)
		node = fresh
	case !xreflect.IsNil(fresh):
		r.replace(node, fresh)
		node = fresh
	}
	r.kept[node] = true
	node.grants().source = source
	r.unfiled = append(r.unfiled, node)
	r.entity.grantPlans = append(r.entity.grantPlans, grantPlan{grant: g, source: source, node: node})
}

// makeItem returns a new item for the sheet made from the grant's copy of it, with the answers of what carries the
// grant for the placeholders it doesn't answer itself.
func (r *grantReconciler[T]) makeItem(g *Grant, item T, link *GrantedBy, answers map[string]string) T {
	var noParent T
	node := item.Clone(LibraryFile{}, r.entity, noParent, Duplicate)
	*grantLink(node) = link
	node.ApplyNameableKeys(mergeReplacements(maps.Clone(node.NameableReplacements()), answers))
	node.SetDisabled(false)
	// The grant is the item's source, so it has no library source to be synced with.
	node.ClearSource()
	switch n := any(node).(type) {
	case *Trait:
		if n.IsLeveled() {
			n.Levels = g.Level
		}
		for _, mod := range n.Modifiers {
			Traverse(func(m *TraitModifier) bool {
				m.ClearSource()
				return false
			}, false, false, mod)
		}
	case *Skill:
		n.Points = 0
		ResolveEmptyTechLevel(n, r.entity.Profile.TechLevel)
	case *Spell:
		n.Points = 0
		ResolveEmptyTechLevel(n, r.entity.Profile.TechLevel)
	}
	return node
}

// replace puts fresh where old is, keeping the old item's ID and switch, which the user may have flipped.
func (r *grantReconciler[T]) replace(old, fresh T) {
	switch n := any(fresh).(type) {
	case *Trait:
		if o, ok := any(old).(*Trait); ok {
			n.TID = o.TID
			n.SwitchedOn = o.SwitchedOn
		}
	case *Skill:
		if o, ok := any(old).(*Skill); ok {
			n.TID = o.TID
			n.SwitchedOn = o.SwitchedOn
		}
	case *Spell:
		if o, ok := any(old).(*Spell); ok {
			n.TID = o.TID
			n.SwitchedOn = o.SwitchedOn
		}
	}
	parent := old.Parent()
	fresh.SetParent(parent)
	if xreflect.IsNil(parent) {
		if i := slices.Index(*r.list, old); i >= 0 {
			(*r.list)[i] = fresh
		}
		return
	}
	children := parent.NodeChildren()
	if i := slices.Index(children, old); i >= 0 {
		children[i] = fresh
	}
}

// finish removes the granted items no grant kept.
func (r *grantReconciler[T]) finish() {
	for _, nodes := range r.existing {
		for _, node := range nodes {
			if !r.kept[node] {
				removeGrantedNode(r.list, node)
			}
		}
	}
}

// removeGrantedNode takes the node out of its parent, or out of the list when it has none.
func removeGrantedNode[T Node[T]](list *[]T, node T) {
	isNode := func(other T) bool { return other == node }
	if parent := node.Parent(); !xreflect.IsNil(parent) {
		parent.SetChildren(slices.DeleteFunc(parent.NodeChildren(), isNode))
	} else {
		*list = slices.DeleteFunc(*list, isNode)
	}
}

// excludedFromPrereqs returns true if the node takes no part in judging the prerequisites of exclude: it is exclude
// itself, or an item one of exclude's grants put on the sheet, in effect or not, so that an item never meets its own
// prerequisites with what it grants.
func excludedFromPrereqs(exclude, node any) bool {
	if node == exclude {
		return true
	}
	h, ok := node.(grantHolder)
	return ok && exclude != nil && IsGranted(node) && h.grants().source.owner == exclude
}

// grantCarrierUnmet returns true if the item carrying a grant has prerequisites that are not met, which leaves its
// grants out as if they weren't there. Until a pass of this recalculation has judged them, they count as met.
func (e *Entity) grantCarrierUnmet(owner fmt.Stringer) bool {
	if !e.grantCarriersJudged {
		return false
	}
	switch one := owner.(type) {
	case *Trait:
		return one.UnsatisfiedReason != ""
	case *Skill:
		return one.UnsatisfiedReason != ""
	case *Spell:
		return one.UnsatisfiedReason != ""
	case *Equipment:
		return one.UnsatisfiedReason != ""
	default:
		return false
	}
}

// removeSuppressedItems takes off the sheet, once it has settled, the items put there by grants whose carriers have
// prerequisites that are not met. They stay through the passes, not in effect, so that the items on the sheet don't
// change while they run. The next recalculation puts them back to judge the prerequisites again.
func (e *Entity) removeSuppressedItems() {
	for _, plan := range e.grantPlans {
		if !e.grantCarrierUnmet(plan.source.owner) {
			continue
		}
		switch n := plan.node.(type) {
		case *Trait:
			removeGrantedNode(&e.Traits, n)
		case *Skill:
			removeGrantedNode(&e.Skills, n)
		case *Spell:
			removeGrantedNode(&e.Spells, n)
		}
	}
}

// applyGrants decides, for the current recalculation pass, which granted items are in effect and what the grants in
// effect merge into the items already on the sheet. Keep Highest merges go last, so that they compare with what the
// other grants gave.
func (e *Entity) applyGrants() {
	// What the last pass gave an item is taken back, whether or not the item is still planned for. It is kept until
	// now so that a recalculation starts from the derived state the last one left.
	for _, holder := range e.grantHolders {
		holder.grants().reset()
	}
	e.grantHolders = e.grantHolders[:0]
	for _, plan := range e.grantPlans {
		holder := plan.node
		if holder == nil {
			holder = plan.target
		}
		holder.grants().reset()
		e.grantHolders = append(e.grantHolders, holder)
	}
	var keepHighest []grantPlan
	for _, plan := range e.grantPlans {
		if !e.features.grants[plan.grant] || e.grantCarrierUnmet(plan.source.owner) {
			continue
		}
		switch {
		case plan.node != nil:
			state := plan.node.grants()
			state.inEffect = true
			state.bypass = !plan.grant.CheckPrereqs
			if n, ok := plan.node.(SkillLevelStepper); ok {
				state.addPoints(grantSourceName(plan.source), grantTierPoints(plan.grant.Level, n.WildcardDifficulty()))
			}
		case plan.grant.effectiveMerge() == grantmerge.KeepHighest:
			keepHighest = append(keepHighest, plan)
		default:
			e.mergeGrant(plan)
		}
	}
	for _, plan := range keepHighest {
		e.mergeGrant(plan)
	}
	e.grantCarriersJudged = true
}

// mergeGrant merges a grant in effect into the item already on the sheet that it matches, as its merge mode says.
func (e *Entity) mergeGrant(plan grantPlan) {
	g := plan.grant
	state := plan.target.grants()
	source := grantSourceName(plan.source)
	level := g.level()
	switch g.effectiveMerge() {
	case grantmerge.MergeLevels:
		state.addLevels(source, level)
	case grantmerge.MergePoints:
		switch t := plan.target.(type) {
		case *Skill:
			state.addPoints(source, grantPointsDelta(level, t.AdjustedPoints(nil), t.WildcardDifficulty()))
		case *Spell:
			state.addPoints(source, grantPointsDelta(level, t.AdjustedPoints(nil), t.WildcardDifficulty()))
		}
	case grantmerge.KeepHighest:
		switch t := plan.target.(type) {
		case *Trait:
			diff := level - t.internalCurrentLevel(nil)
			// A disadvantage is better with fewer levels.
			if t.pointsWith(t.AllModifiers()) < 0 {
				diff = diff.Min(0)
			} else {
				diff = diff.Max(0)
			}
			state.addLevels(source, diff)
		case *Skill:
			current := t.CalculateLevel(nil).Level
			granted := g.Skill.levelAt(e, grantTierPoints(level, g.Skill.WildcardDifficulty())).Level
			if current != fxp.Min && granted != fxp.Min && granted > current {
				state.addLevels(source, granted-current)
			}
		case *Spell:
			current := t.CalculateLevel().Level
			granted := g.Spell.levelAt(e, grantTierPoints(level, g.Spell.WildcardDifficulty())).Level
			if current != fxp.Min && granted != fxp.Min && granted > current {
				state.addLevels(source, granted-current)
			}
		}
	}
}

// grantSourceName names what carries a grant, the way bonus tooltips name their sources.
func grantSourceName(active grantSource) string {
	if xreflect.IsNil(active.owner) {
		return i18n.Text("Unknown")
	}
	if xreflect.IsNil(active.subOwner) {
		return active.owner.String()
	}
	return fmt.Sprintf("%s (%v)", active.owner, active.subOwner)
}

// levelAt returns the level the skill, which need not be on the sheet, has on the sheet with the given points.
func (s *Skill) levelAt(e *Entity, points fxp.Int) Level {
	return CalculateSkillLevel(e, s.NameWithReplacements(), s.SpecializationWithReplacements(),
		s.OptionalSpecializationWithReplacements(), s.Tags, nil, s.AdjustedDifficulty(), points,
		s.EncumbrancePenaltyMultiplier)
}

// levelAt returns the level the spell, which need not be on the sheet, has on the sheet with the given points.
func (s *Spell) levelAt(e *Entity, points fxp.Int) Level {
	if s.IsRitualMagic() {
		return CalculateRitualMagicSpellLevel(e, s.NameWithReplacements(), s.PowerSourceWithReplacements(),
			s.RitualSkillNameWithReplacements(), s.PrereqCount, s.CollegeWithReplacements(), s.Tags, s.Difficulty,
			points)
	}
	return CalculateSpellLevel(e, s.NameWithReplacements(), s.PowerSourceWithReplacements(),
		s.CollegeWithReplacements(), s.Tags, s.Difficulty, points)
}

// grantSummary returns what the grants carried by the item, or by its modifiers, did at the last recalculation, a line
// for each, or an empty string when it carries none.
func (e *Entity) grantSummary(item any) string {
	if e == nil {
		return ""
	}
	var lines []string
	for _, plan := range e.grantPlans {
		if plan.source.owner != item {
			continue
		}
		active := e.features.grants[plan.grant]
		name := plan.grant.itemLabel()
		var line string
		switch {
		case e.grantCarrierUnmet(plan.source.owner):
			line = i18n.Text("Does not grant %s, since the prerequisites of what carries the grant are not met", name)
		case plan.node != nil && active:
			line = i18n.Text("Grants %s", name)
		case plan.node != nil:
			line = i18n.Text("Would grant %s, but the grant is not in effect", name)
		case !active:
			line = i18n.Text("Would add to %s, but the grant is not in effect", plan.grant.ItemName())
		default:
			line = plan.grant.mergeSummary(name)
		}
		if !xreflect.IsNil(plan.source.subOwner) {
			line = i18n.Text("%s (from %v)", line, plan.source.subOwner)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
