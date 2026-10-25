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
	"cmp"
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
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/zeebo/xxh3"
)

// grantPlan is what reconcileGrants decided for one grant on the sheet: the item it put there, or the item already on
// the sheet that it merges into, or neither. seq is the grant's place in sheet order, which the plans are not kept in
// (see reconcileGrants), and standsIn is set by a pass in which the grant stands in for the one that put its target on
// the sheet; see applyGrants.
type grantPlan struct {
	grant    *Grant
	source   grantSource
	node     grantHolder
	target   grantHolder
	seq      int
	standsIn bool
}

// sheetGrant is a grant found on the sheet, along with the ID of the item or modifier carrying it, what carries it, its
// place in sheet order and whether a row it put on the sheet is there; see Entity.forEachGrant.
type sheetGrant struct {
	grantor tid.TID
	source  grantSource
	grant   *Grant
	seq     int
	holds   bool
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

// GrantingItem returns the trait, skill, spell or equipment carrying the grant that put the node on the sheet, or the
// one carrying the grant standing in for it (see applyGrants), or nil for a node no grant put there.
func GrantingItem(node any) any {
	if h, ok := node.(grantHolder); ok && IsGranted(node) {
		state := h.grants()
		if !xreflect.IsNil(state.standIn.owner) {
			return state.standIn.owner
		}
		if !xreflect.IsNil(state.source.owner) {
			return state.source.owner
		}
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
// does not. When several grants would put the same item on the sheet, the grants are taken in the order
// comparePlanting gives rather than sheet order, so that which of them puts it there, and which stands in for that
// one (see applyGrants), is decided by the grants rather than by where their carriers are listed; among grants that
// compare equal, one whose row is already on the sheet comes first, so that the row stays as it is when their
// carriers are reordered. A grant whose own row is absent takes over a row of the same item that no grant on the sheet
// claims, so that the row stays where the user put it when the carrier whose grant put it there is disabled or removed
// while an equal grant of another carrier remains.
func (e *Entity) reconcileGrants() {
	e.grantPlans = e.grantPlans[:0]
	e.grantCarriersJudged = false
	traits := newGrantReconciler(e, &e.Traits)
	skills := newGrantReconciler(e, &e.Skills)
	spells := newGrantReconciler(e, &e.Spells)
	var grants []sheetGrant
	e.forEachGrant(func(grantor tid.TID, source grantSource, g *Grant) {
		k := grantKey{grantor: grantor, key: g.Key()}
		var holds bool
		switch g.Type {
		case feature.GrantTrait:
			holds = traits.claim(k)
		case feature.GrantSkill:
			holds = skills.claim(k)
		case feature.GrantSpell:
			holds = spells.claim(k)
		}
		grants = append(grants, sheetGrant{grantor: grantor, source: source, grant: g, seq: len(grants), holds: holds})
	})
	slices.SortStableFunc(grants, func(a, b sheetGrant) int {
		if c := comparePlanting(a.grant, b.grant); c != 0 {
			return c
		}
		switch {
		case a.holds == b.holds:
			return 0
		case a.holds:
			return -1
		default:
			return 1
		}
	})
	// A container or technique can't be granted, as the pickers leave them out, so a grant of one from a hand-edited
	// file does nothing.
	for _, one := range grants {
		g := one.grant
		planned := len(e.grantPlans)
		switch g.Type {
		case feature.GrantTrait:
			if !g.Trait.Container() {
				traits.consider(one.grantor, one.source, g, g.Trait)
			}
		case feature.GrantSkill:
			if !g.Skill.IsTechnique() && !g.Skill.Container() {
				skills.consider(one.grantor, one.source, g, g.Skill)
			}
		case feature.GrantSpell:
			if !g.Spell.Container() {
				spells.consider(one.grantor, one.source, g, g.Spell)
			}
		}
		if planned < len(e.grantPlans) {
			e.grantPlans[planned].seq = one.seq
		}
	}
	traits.finish()
	skills.finish()
	spells.finish()
}

// comparePlanting orders two grants for reconcileGrants, which takes them in this order when they would put the same
// item on the sheet: a grant that stands for the item on its own, Ignore and then Keep Highest, comes before one that
// adds to it, Merge Levels and then Merge Points, so that an Ignore grant gives the item its highest level and a merge
// adds to that rather than being what is added to; within a mode the higher level comes first, then one that treats
// the item's prerequisites as met before one that checks them, and then the items are compared by what tells apart
// those the reconciler counts as the same item (see comparePlantingItems). Grants that compare equal give the same
// result whichever of them puts the item on the sheet.
func comparePlanting(a, b *Grant) int {
	if c := cmp.Compare(plantingRank(a.EffectiveMerge()), plantingRank(b.EffectiveMerge())); c != 0 {
		return c
	}
	if c := cmp.Compare(b.level(), a.level()); c != 0 {
		return c
	}
	if a.CheckPrereqs != b.CheckPrereqs {
		if a.CheckPrereqs {
			return 1
		}
		return -1
	}
	return comparePlantingItems(a, b)
}

// comparePlantingItems orders two grants by what tells their items apart among those the reconciler counts as the same
// item (see grantReconciler.consider), which the items' hashes leave out: a trait whose self-control roll and frequency
// make it worth more comes first, as a higher level does, and then the stricter roll and frequency, a skill's optional
// specialization, and the user description, VTT notes and switch of any item, so that only grants whose items the
// sheet can't tell apart compare equal.
func comparePlantingItems(a, b *Grant) int {
	if c := cmp.Compare(a.Type, b.Type); c != 0 {
		return c
	}
	var aNotes, bNotes string
	var aSwitch, bSwitch bool
	switch a.Type {
	case feature.GrantTrait:
		ta, tb := a.Trait, b.Trait
		if c := cmp.Compare(tb.SelfControl.Multiplier().Mul(tb.Frequency.Multiplier()),
			ta.SelfControl.Multiplier().Mul(ta.Frequency.Multiplier())); c != 0 {
			return c
		}
		if c := cmp.Compare(ta.SelfControl, tb.SelfControl); c != 0 {
			return c
		}
		if c := cmp.Compare(ta.Frequency, tb.Frequency); c != 0 {
			return c
		}
		if c := cmp.Compare(ta.UserDesc, tb.UserDesc); c != 0 {
			return c
		}
		aNotes, bNotes = ta.VTTNotes, tb.VTTNotes
		aSwitch, bSwitch = ta.SwitchedOn, tb.SwitchedOn
	case feature.GrantSkill:
		sa, sb := a.Skill, b.Skill
		if c := cmp.Compare(sa.OptionalSpecialization, sb.OptionalSpecialization); c != 0 {
			return c
		}
		aNotes, bNotes = sa.VTTNotes, sb.VTTNotes
		aSwitch, bSwitch = sa.SwitchedOn, sb.SwitchedOn
	case feature.GrantSpell:
		sa, sb := a.Spell, b.Spell
		aNotes, bNotes = sa.VTTNotes, sb.VTTNotes
		aSwitch, bSwitch = sa.SwitchedOn, sb.SwitchedOn
	}
	if c := cmp.Compare(aNotes, bNotes); c != 0 {
		return c
	}
	switch {
	case aSwitch == bSwitch:
		return 0
	case aSwitch:
		return -1
	default:
		return 1
	}
}

// plantingRank returns the place of a merge mode in comparePlanting's order.
func plantingRank(merge grantmerge.Mode) int {
	switch merge {
	case grantmerge.Ignore:
		return 0
	case grantmerge.KeepHighest:
		return 1
	case grantmerge.MergeLevels:
		return 2
	default:
		return 3
	}
}

type grantReconciler[T grantNode[T]] struct {
	entity   *Entity
	list     *[]T
	existing map[grantKey][]T
	// granted holds the rows grants put on the sheet, in sheet order, and claimed the keys of the grants on the sheet;
	// a row under any other key is an orphan a grant may take over, see orphan.
	granted []T
	claimed map[grantKey]bool
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
		claimed:  make(map[grantKey]bool),
		present:  make(map[string][]T),
		kept:     make(map[T]bool),
	}
	r.file(*list)
	return r
}

// file sorts the nodes into those a grant put on the sheet, by the grant's key, and those a grant can merge into. A
// granted row goes with its grant, children and all, so nothing within one is filed on its own.
func (r *grantReconciler[T]) file(nodes []T) {
	for _, node := range nodes {
		switch link := *grantLink(node); {
		case link != nil:
			k := grantKey{grantor: link.Grantor, key: link.Key}
			r.existing[k] = append(r.existing[k], node)
			r.granted = append(r.granted, node)
		case node.Container():
			r.file(node.NodeChildren())
		case !userDisabled(node):
			r.unfiled = append(r.unfiled, node)
		}
	}
}

// claim records that a grant on the sheet has the key and returns true if a row the grant put on the sheet is there.
func (r *grantReconciler[T]) claim(k grantKey) bool {
	r.claimed[k] = true
	return len(r.existing[k]) > 0
}

// orphan returns the first row in sheet order that is the same item as the probe and that no grant on the sheet claims,
// taking it out of the running, or false when there is none.
func (r *grantReconciler[T]) orphan(probe T) (T, bool) {
	h := identityHash(probe)
	for i, candidate := range r.granted {
		link := *grantLink(candidate)
		if r.claimed[grantKey{grantor: link.Grantor, key: link.Key}] {
			continue
		}
		if identityHash(candidate) == h && SameItem(candidate, probe) && sameModifiers(candidate, probe) {
			r.granted = slices.Delete(r.granted, i, i+1)
			return candidate, true
		}
	}
	var none T
	return none, false
}

// presentNamed returns the items present with the same name as the probe.
func (r *grantReconciler[T]) presentNamed(probe T) []T {
	for _, node := range r.unfiled {
		r.present[node.rawName()] = append(r.present[node.rawName()], node)
	}
	r.unfiled = r.unfiled[:0]
	return r.present[probe.rawName()]
}

// identityHash returns the item's hash (see Hash64) with any grants among its features left out, as they are left out
// of the item a grant gives (see Grant.SetItem), so that an item carrying grants and a grant's copy of it count as the
// same item.
func identityHash[T grantNode[T]](node T) uint64 {
	switch n := any(node).(type) {
	case *Trait:
		if features, dropped := withoutGrants(n.Features); dropped {
			shallow := *n
			shallow.Features = features
			return Hash64(&shallow)
		}
	case *Skill:
		if features, dropped := withoutGrants(n.Features); dropped {
			shallow := *n
			shallow.Features = features
			return Hash64(&shallow)
		}
	case *Spell:
		if features, dropped := withoutGrants(n.Features); dropped {
			shallow := *n
			shallow.Features = features
			return Hash64(&shallow)
		}
	}
	return Hash64(node)
}

// withoutGrants returns the features with any grants left out, and whether there were any; the list itself comes back
// when there were none.
func withoutGrants(list Features) (features Features, dropped bool) {
	isGrant := func(f Feature) bool {
		_, ok := f.(*Grant)
		return ok
	}
	if !slices.ContainsFunc(list, isGrant) {
		return list, false
	}
	return slices.DeleteFunc(slices.Clone(list), isGrant), true
}

// sameModifiers returns true if the items' own modifiers in effect match, which SameItem leaves out, since a trait's
// hash does: a trait with a modifier is another item than the plain trait, so a grant of one never merges into the
// other.
func sameModifiers[T grantNode[T]](a, b T) bool {
	ta, ok := any(a).(*Trait)
	if !ok {
		return true
	}
	tb, ok := any(b).(*Trait)
	return ok && ta.modifierSignature() == tb.modifierSignature()
}

// modifierSignature returns a hash of the trait's own modifiers in effect, by their sync data and levels, in any order,
// with any grants among their features left out as identityHash leaves out the trait's own.
func (t *Trait) modifierSignature() uint64 {
	var hashes []uint64
	h := xxh3.New()
	Traverse(func(m *TraitModifier) bool {
		h.Reset()
		if features, dropped := withoutGrants(m.Features); dropped {
			shallow := *m
			shallow.Features = features
			shallow.Hash(h)
		} else {
			m.Hash(h)
		}
		xhash.Num64(h, m.Levels)
		hashes = append(hashes, h.Sum64())
		return false
	}, true, true, t.Modifiers...)
	slices.Sort(hashes)
	h.Reset()
	for _, one := range hashes {
		xhash.Num64(h, one)
	}
	return h.Sum64()
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
	h := identityHash(probe)
	for _, candidate := range r.presentNamed(probe) {
		if identityHash(candidate) == h && SameItem(candidate, probe) && sameModifiers(candidate, probe) {
			r.entity.grantPlans = append(r.entity.grantPlans, grantPlan{grant: g, source: source, target: candidate})
			return
		}
	}
	if !g.putsItemOnSheet() {
		return
	}
	if !has {
		node, has = r.orphan(fresh)
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
	// The user may have moved the item into a container and disabled that, in which case it counts as absent.
	if !userDisabled(node) {
		r.unfiled = append(r.unfiled, node)
	}
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

// replace puts fresh where old is, keeping the old item's ID and switch, which the user may have flipped. Old is out of
// the running from then on, or another grant with the same key would claim it (see consider) once it is in no list.
func (r *grantReconciler[T]) replace(old, fresh T) {
	r.kept[old] = true
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
// itself, or an item one of exclude's grants put on the sheet (see grantedFor), so that an item never meets its own
// prerequisites with what it grants. What exclude's grants merged into an item taking part is left out of its level and
// points as well (see grantState.applyLevelNotes), as are the bonuses of the items left out and the skill defaults
// resting on them (see Entity.JudgingPrereqsOf).
func excludedFromPrereqs(exclude, node any) bool {
	return node == exclude || grantedFor(exclude, node)
}

// grantedFor returns true if the node is an item on the sheet only because of the grants carried by owner: every grant
// in effect that puts it on the sheet, or would have had another not put it there first (see applyGrants), is owner's,
// or, while none is in effect, the grant that put it there is. An item another carrier's grant in effect would put on
// the sheet is not granted for owner even while owner's grant is the one in effect, since counting it only while the
// other's stands in would bring owner's grant into effect and so take the stand-in away again. owner may be nil, which
// nothing is granted for.
func grantedFor(owner, node any) bool {
	h, ok := node.(grantHolder)
	if !ok || owner == nil || !IsGranted(node) {
		return false
	}
	state := h.grants()
	if len(state.providers) == 0 {
		return state.source.owner == owner
	}
	for _, provider := range state.providers {
		if provider != owner {
			return false
		}
	}
	return true
}

// grantCarrierUnmet returns true if the item carrying a grant has prerequisites that are not met, which leaves its
// grants out of effect, as a switch turned off does, with the items they put on the sheet staying there dimmed. Until a
// pass of this recalculation has judged them, they count as met. A trait above its maximum level, which records that
// along with any unmet prerequisite, keeps its grants in effect, as it does its other features.
func (e *Entity) grantCarrierUnmet(owner fmt.Stringer) bool {
	if !e.grantCarriersJudged {
		return false
	}
	switch one := owner.(type) {
	case *Trait:
		return one.prereqsUnmet
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

// applyGrants decides, for the current recalculation pass, which granted items are in effect and what the grants in
// effect merge into the items already on the sheet. A grant that would have put its item on the sheet, had another
// grant not put it there first, stands in for that grant while it is not in effect, giving the item its own level, so
// that the item is in effect whichever of them is listed first. The merges follow, so that they find the items in
// effect whatever the order.
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
	for i := range e.grantPlans {
		plan := &e.grantPlans[i]
		plan.standsIn = false
		if plan.node != nil && e.grantInEffect(plan) {
			e.grantItem(plan, plan.node)
		}
		if plan.grant.putsItemOnSheet() && e.grantInEffect(plan) {
			holder := plan.node
			if holder == nil {
				holder = plan.target
			}
			if IsGranted(holder) {
				holder.grants().providers = append(holder.grants().providers, plan.source.owner)
			}
		}
	}
	for i := range e.grantPlans {
		plan := &e.grantPlans[i]
		if plan.node == nil && IsGranted(plan.target) && !plan.target.grants().inEffect && e.grantInEffect(plan) &&
			plan.grant.putsItemOnSheet() {
			plan.standsIn = true
			plan.target.grants().standIn = plan.source
			e.grantItem(plan, plan.target)
		}
	}
	for i := range e.grantPlans {
		plan := &e.grantPlans[i]
		if plan.node == nil && !plan.standsIn && e.grantInEffect(plan) {
			e.mergeGrant(plan)
		}
	}
	e.grantCarriersJudged = true
}

// grantInEffect returns true if the plan's grant is in effect in the current pass: what carries it contributes its
// features and has its prerequisites met.
func (e *Entity) grantInEffect(plan *grantPlan) bool {
	return e.features.grants[plan.grant] && !e.grantCarrierUnmet(plan.source.owner)
}

// grantItem puts the holder in effect as the item the plan's grant gives: the item the grant put on the sheet, or one
// another grant put there that this one stands in for, which is brought to this grant's level.
func (e *Entity) grantItem(plan *grantPlan, holder grantHolder) {
	state := holder.grants()
	state.inEffect = true
	state.bypass = !plan.grant.CheckPrereqs
	switch n := holder.(type) {
	case *Trait:
		// The item a grant put on the sheet has that grant's level, so this changes only the level of one stood in for.
		if n.IsLeveled() {
			state.addLevels(plan.source, plan.grant.level()-n.Levels)
		}
	case SkillLevelStepper:
		state.tierPoints = grantTierPoints(plan.grant.Level, n.WildcardDifficulty())
		state.addPoints(plan.source, state.tierPoints)
	}
}

// mergeGrant merges a grant in effect into the item already on the sheet that it matches, as its merge mode says. A
// skill or spell with no points has no level to add to or compare with, so a grant whose level would have put the item
// on the sheet gives it the points of its tier instead, as if it had. A Keep Highest grant leaves the level it keeps,
// or for a skill or spell the points of the tier it keeps, for the item's level to be compared with when it is read,
// with the other merges in (see grantState.applyLevelNotes).
func (e *Entity) mergeGrant(plan *grantPlan) {
	g := plan.grant
	if g.mergesNothing() {
		return
	}
	state := plan.target.grants()
	source := plan.source
	level := g.level()
	merge := g.EffectiveMerge()
	// The item's own level and points are read without what other grants gave it, so that what each grant gives
	// doesn't depend on the order of the grants. A granted item's own points are those of the tier the grant in effect
	// gives it.
	var own, ownPoints fxp.Int
	switch t := plan.target.(type) {
	case *Skill:
		raw := t.Points
		if t.Granted != nil {
			raw = state.tierPoints
		}
		ownPoints = t.pointsWithoutGrants(raw)
		own = t.levelAt(e, t.DefaultedFrom, raw).Level
	case *Spell:
		raw := t.Points
		if t.Granted != nil {
			raw = state.tierPoints
		}
		ownPoints = t.pointsWithoutGrants(raw)
		own = t.levelAt(e, raw).Level
	}
	if stepper, ok := plan.target.(SkillLevelStepper); ok && own == fxp.Min &&
		(merge == grantmerge.MergeLevels || merge == grantmerge.KeepHighest) {
		if g.putsItemOnSheet() {
			state.addPoints(source, grantTierPoints(level, stepper.WildcardDifficulty()))
		}
		return
	}
	switch merge {
	case grantmerge.MergeLevels:
		state.addLevels(source, level)
	case grantmerge.MergePoints:
		if stepper, ok := plan.target.(SkillLevelStepper); ok {
			state.addPoints(source, grantPointsDelta(level, ownPoints, state.tiersTaken, stepper.WildcardDifficulty()))
			if level < 0 {
				state.tiersTaken += level
			}
		}
	case grantmerge.KeepHighest:
		switch t := plan.target.(type) {
		case *Trait:
			// A disadvantage, whose levels cost less than nothing, is better with fewer of them, whatever it costs at
			// the level it has.
			state.keepLevel(source, level, t.PointsPerLevel < 0)
		case SkillLevelStepper:
			// The level the grant gives is that of the item on the sheet with the points of the grant's tier, so that
			// the placeholder answers and bonuses the item has count for both.
			state.keepTier(source, grantTierPoints(level, t.WildcardDifficulty()))
		}
	}
}

// carriesGrant returns true if owner, which may be nil, carries a grant that puts an item on the sheet or merges into
// one, whether or not the grant is in effect this pass.
func (e *Entity) carriesGrant(owner any) bool {
	if e == nil || owner == nil {
		return false
	}
	for _, plan := range e.grantPlans {
		if plan.source.owner == owner {
			return true
		}
	}
	return false
}

// grantTierPointsFor returns the points of the tier the grant that put the skill or spell on the sheet gives it, whether
// or not the grant is in effect, or zero for one no grant put there.
func (e *Entity) grantTierPointsFor(holder grantHolder) fxp.Int {
	stepper, ok := holder.(SkillLevelStepper)
	if !ok || e == nil {
		return 0
	}
	for _, plan := range e.grantPlans {
		if plan.node == holder {
			return grantTierPoints(plan.grant.Level, stepper.WildcardDifficulty())
		}
	}
	return 0
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

// levelAt returns the level the skill would have with the given points in place of its own, adjusted for the point
// bonuses as its own are, leaving out what grants gave it, and the given default in place of the one it recorded.
func (s *Skill) levelAt(e *Entity, def *SkillDefault, points fxp.Int) Level {
	name := s.NameWithReplacements()
	specialization := s.SpecializationWithReplacements()
	optionalSpecialization := s.OptionalSpecializationWithReplacements()
	points = AdjustedPointsForNonContainerSkillOrTechnique(e, points, name, specialization, optionalSpecialization,
		s.Tags, nil)
	return CalculateSkillLevel(e, name, specialization, optionalSpecialization, s.Tags, def, s.AdjustedDifficulty(),
		points, s.EncumbrancePenaltyMultiplier)
}

// levelAt returns the level the spell would have with the given points in place of its own, adjusted for the point
// bonuses as its own are, leaving out what grants gave it.
func (s *Spell) levelAt(e *Entity, points fxp.Int) Level {
	return s.levelFromAdjustedPoints(AdjustedPointsForNonContainerSpell(e, points, s.NameWithReplacements(),
		s.PowerSourceWithReplacements(), s.CollegeWithReplacements(), s.Tags, nil))
}

// grantSummary returns what the grants carried by the item, or by its modifiers, did at the last recalculation, a line
// for each in the order the item carries them, or an empty string when it carries none.
func (e *Entity) grantSummary(item any) string {
	if e == nil {
		return ""
	}
	var plans []*grantPlan
	for i := range e.grantPlans {
		if e.grantPlans[i].source.owner == item {
			plans = append(plans, &e.grantPlans[i])
		}
	}
	slices.SortFunc(plans, func(a, b *grantPlan) int { return cmp.Compare(a.seq, b.seq) })
	var lines []string
	for _, plan := range plans {
		active := e.features.grants[plan.grant]
		answers := plan.grant.answers(plan.source)
		name := plan.grant.itemLabel(answers)
		var line string
		switch {
		case e.grantCarrierUnmet(plan.source.owner):
			line = i18n.Text("Does not grant %s, since the prerequisites of what carries the grant are not met", name)
		case plan.node != nil && active:
			line = i18n.Text("Grants %s", name)
		case plan.node != nil:
			line = i18n.Text("Would grant %s, but the grant is not in effect", name)
		case plan.standsIn:
			line = i18n.Text("Grants %s, standing in for a grant of it that is not in effect", name)
		case !active:
			line = i18n.Text("Would add to %s, but the grant is not in effect", plan.grant.itemNameWith(answers))
		default:
			line = plan.grant.mergeSummary(name, answers)
		}
		if !xreflect.IsNil(plan.source.subOwner) {
			line = i18n.Text("%s (from %v)", line, plan.source.subOwner)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
