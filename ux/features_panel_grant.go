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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantadjust"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantmerge"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
	"github.com/zeebo/xxh3"
)

// grantLibraryItem is an item a grant can give, from a library or a table: a trait, skill or spell.
type grantLibraryItem struct {
	node any
	from gurps.LibraryFile
}

// newGrantLibraryItem returns the node, from the library file, as an item a grant can give, or nil for a technique or
// an item that grants something itself, among its own features or those of a trait's modifiers, which no grant can
// give.
func newGrantLibraryItem(node any, from gurps.LibraryFile) *grantLibraryItem {
	var features []gurps.Features
	switch one := node.(type) {
	case *gurps.Trait:
		features = []gurps.Features{one.Features}
		gurps.Traverse(func(mod *gurps.TraitModifier) bool {
			features = append(features, mod.Features)
			return false
		}, false, false, one.Modifiers...)
	case *gurps.Skill:
		if one.IsTechnique() {
			return nil
		}
		features = []gurps.Features{one.Features}
	case *gurps.Spell:
		features = []gurps.Features{one.Features}
	}
	for _, list := range features {
		if slices.ContainsFunc(list, func(f gurps.Feature) bool { return gurps.IsGrant(f.FeatureType()) }) {
			return nil
		}
	}
	return &grantLibraryItem{node: node, from: from}
}

// collectGrantItems returns the items a grant can give among the nodes and their descendants, containers left out.
func collectGrantItems[T gurps.Node[T]](from gurps.LibraryFile, nodes []T) []*grantLibraryItem {
	var items []*grantLibraryItem
	gurps.Traverse(func(one T) bool {
		if item := newGrantLibraryItem(one, from); item != nil {
			items = append(items, item)
		}
		return false
	}, false, true, nodes...)
	return items
}

// grantDropZoneMinWidth keeps the drop zone of a grant wide enough to drop on when the name of its item is short.
const grantDropZoneMinWidth = 150

// grant adds the controls of an open grant row: a zone to drop the item on, which names the item chosen, if any, and
// when clicked opens the editor on it, giving the grant a blank item first when it has none; once an item is chosen,
// how its amount adjusts it and the amount (see grantAmount), then, for a grant that sets its level, what happens when
// the sheet already has it; and, as a chip, whether its prerequisites are checked.
func (p *featuresPanel) grant(fields, chips *unison.Panel, path string, g *gurps.Grant) {
	key := func(name string) string { return path + ":" + name }
	chosen := g.Key() != ""
	zone := newEmptyPlaceholder(key("drop"), grantDropText(g.Type, false), func() {
		p.sentenceRows.restructure(i18n.Text("New Blank Granted Item"), key("drop"), "", func() string {
			g.SetItem(nil, gurps.LibraryFile{})
			return grantFocusKey(path, g)
		})
		// The editor opens on the blank item once the panel has been rebuilt around it, as a click on the zone then
		// would, so that the rebuild's focus doesn't take the editor from the front.
		unison.InvokeTask(func() { p.editGrantedItem(g) })
	})
	if chosen {
		zone.SetTitle(g.ItemName())
		zone.ClickCallback = func() { p.editGrantedItem(g) }
		replace := grantDropText(g.Type, true)
		zone.Tooltip = newWrappedTooltip(replace)
		zone.Accessibility.Description = replace
	}
	keepMinWidth(zone.AsPanel(), grantDropZoneMinWidth, zone.DefaultSizes)
	addCentered(fields, zone)
	if chosen && g.Leveled() {
		p.grantAmount(fields, path, g)
	}
	if merges := g.AllowedMerges(); chosen && len(merges) > 1 {
		addJoiningWords(fields, i18n.Text("and if already present"))
		addCentered(fields, compactPopup(&p.sentenceRows, key("merge"), i18n.Text("When Already Present"), merges,
			g.EffectiveMerge(), nil, func(v grantmerge.Mode) { g.Merge = v }))
	}
	p.optional(chips, path, "checkprereqs", i18n.Text("+ check prerequisites"), i18n.Text("Check Prerequisites"),
		i18n.Text("Stop Checking Prerequisites"), g.CheckPrereqs,
		func() { g.CheckPrereqs = true },
		func() { g.CheckPrereqs = false },
		func(chip *unison.Panel) { addJoiningWords(chip, i18n.Text("checking its prerequisites")) })
}

// grantAmount adds the controls of the amount of a grant whose item it means something for: a popup of how the amount
// adjusts the item, then the amount, signed when it adds levels or points. A choice of adjustment rebuilds the row,
// keeps the amount as typed, less a fraction the new kind's whole levels drop as the grant does, and drops a collision
// mode the new kind has no use for.
func (p *featuresPanel) grantAmount(fields *unison.Panel, path string, g *gurps.Grant) {
	key := func(name string) string { return path + ":" + name }
	kind := g.EffectiveAdjust()
	addCentered(fields, compactPopup(&p.sentenceRows, key("adjust"), i18n.Text("Adjustment"), g.AllowedAdjusts(),
		kind, grantadjust.Kind.AltString, func(v grantadjust.Kind) {
			g.Adjust = v
			g.Merge = g.EffectiveMerge()
			if grantAmountWhole(g.Type, v) {
				g.Amount = g.Amount.Floor()
			}
		}))
	title := grantAmountTitle(kind)
	field := NewDecimalField(p.targetMgr, key("amount"), title, func() fxp.Int { return g.Amount },
		func(v fxp.Int) { p.edit(title, key("amount"), "", func() { g.Amount = v }) },
		fxp.Min, fxp.Max, kind != grantadjust.SetLevel, false)
	field.Accessibility.Name = title
	field.Tooltip = newWrappedTooltip(grantAmountTooltip(g.Type, kind))
	if grantAmountWhole(g.Type, kind) {
		refuseFractions(field)
	}
	p.addCompact(fields, field.withoutUndo())
}

// grantAmountWhole returns true if the amount of a grant of the given type, which adjusts its item as kind says, is
// whole: a skill's or spell's levels, where a trait's levels and any points may be fractional.
func grantAmountWhole(t feature.Type, kind grantadjust.Kind) bool {
	return t != feature.GrantTrait && kind != grantadjust.AddPoints
}

// grantItemNoun returns the word for the kind of item a grant of the given type gives.
func grantItemNoun(t feature.Type) string {
	switch t {
	case feature.GrantSkill:
		return i18n.Text("skill")
	case feature.GrantSpell:
		return i18n.Text("spell")
	default:
		return i18n.Text("trait")
	}
}

// grantDropText invites dropping an item on a grant of the given type, to replace the item chosen, which a click edits,
// or to fill the grant, which a click fills with a blank item.
func grantDropText(t feature.Type, chosen bool) string {
	if chosen {
		return i18n.Text("Drop a %s here from a library or sheet to replace the one chosen, or click to edit it.",
			grantItemNoun(t))
	}
	return i18n.Text("Drop a %s here from a library or sheet, or click to start with a blank one.", grantItemNoun(t))
}

// grantFocusKey returns the reference key of the control of the grant row at the path that takes the focus once the
// grant has its item: its amount field, or its first control when the amount means nothing for the item.
func grantFocusKey(path string, g *gurps.Grant) string {
	if g.Leveled() {
		return path + ":amount"
	}
	return path + keyFirst
}

// grantAmountTitle names the amount of a grant that adjusts its item as kind says.
func grantAmountTitle(kind grantadjust.Kind) string {
	switch kind {
	case grantadjust.SetLevel:
		return i18n.Text("Level")
	case grantadjust.AddPoints:
		return i18n.Text("Point Adjustment")
	default:
		return i18n.Text("Level Adjustment")
	}
}

// grantAmountTooltip explains what the amount of a grant of the given type means, which adjusts its item as kind says.
func grantAmountTooltip(t feature.Type, kind grantadjust.Kind) string {
	if t == feature.GrantTrait {
		if kind == grantadjust.SetLevel {
			return i18n.Text(`The level the grant gives the trait. With Keep Highest or Keep Lowest, a trait already present is held to at least or at most that level. A level below 1 does neither.`)
		}
		return i18n.Text(`The levels added to a trait already present, or taken away when negative, never below level 0. On its own, the grant gives the trait that level when it is 1 or more. A positive level below 1 does nothing at all.`)
	}
	noun := grantItemNoun(t)
	switch kind {
	case grantadjust.AddPoints:
		return i18n.Text(`The points added to a %s already present, or taken away when negative, never below 0. On its own, the grant gives the %s that many points when they are enough to buy a level: at least 1, or 3 for a wildcard.`, noun, noun)
	case grantadjust.SetLevel:
		return i18n.Text(`The skill level the grant gives the %s, in place of the level its points and difficulty would give it; bonuses still apply on top. With Keep Highest or Keep Lowest, one already present is held to at least or at most that level. A level below 1, or below what 1 point buys (3 for a wildcard), does neither.`, noun)
	default:
		return i18n.Text(`The levels added to a %s already present, or taken away when negative, never below level 0. On its own, the grant gives the %s that level when it is 1 or more, counted in steps of points: level 1 is what 1 point buys, level 2 what 2 points buy, level 3 what 4 points buy, and each level after that 4 points more, or three times those points for a wildcard.`, noun, noun)
	}
}

// grantItemOwner owns the editor of a grant's item. The editor works on a copy, which each apply, undo and redo hands
// to the grant with the item's key, since undoing the panel's own edits replaces its grants with copies. hash is the
// hash of the grant's item as the editor last saw it, by which the panel tells that the item changed under the editor;
// see closeStaleGrantItemEditors. The panel is known by the features it edits (see editsFeaturesOf), since the editor
// holding it may build a new panel for the same features.
type grantItemOwner struct {
	Rebuildable // the editor holding the panel
	panel       *featuresPanel
	key         tid.TID
	hash        uint64
	item        fmt.Stringer
	commit      func(g *gurps.Grant)
}

// String names the item, for the title of its editor.
func (o *grantItemOwner) String() string {
	return o.item.String()
}

// Rebuild hands the copy to the grant.
func (o *grantItemOwner) Rebuild(_ bool) {
	if g := o.grant(); g != nil {
		o.commit(g)
		o.hash = grantItemHash(g)
		MarkModified(o.panel)
		o.panel.rebuild("")
	}
}

// editsFeaturesOf returns true if the owner's panel edits the same features as p, taking p for its panel from here on
// when it does, since the editor holding the panel builds a new one when its content is rebuilt.
func (o *grantItemOwner) editsFeaturesOf(p *featuresPanel) bool {
	if o.panel.features != p.features {
		return false
	}
	o.panel = p
	return true
}

// grant returns the panel's grant with the item's key, or nil once there is none.
func (o *grantItemOwner) grant() *gurps.Grant {
	for _, f := range *o.panel.features {
		if g, ok := f.(*gurps.Grant); ok && g.Key() == o.key {
			return g
		}
	}
	return nil
}

// current returns true if the panel's grant still holds the item the editor was opened on.
func (o *grantItemOwner) current() bool {
	g := o.grant()
	return g != nil && grantItemHash(g) == o.hash
}

// grantItemHash returns a hash of the grant's copy of its item, modifiers and all. The item is the one of the grant's
// type, as for the grant itself, whatever else a loaded grant holds.
func grantItemHash(g *gurps.Grant) uint64 {
	h := xxh3.New()
	switch g.Type {
	case feature.GrantSkill:
		if g.Skill != nil {
			gurps.HashJSON(h, g.Skill)
		}
	case feature.GrantSpell:
		if g.Spell != nil {
			gurps.HashJSON(h, g.Spell)
		}
	default:
		if g.Trait != nil {
			gurps.HashJSON(h, g.Trait)
		}
	}
	return h.Sum64()
}

// closeStaleGrantItemEditors closes, without asking, the editors open on the items of the panel's grants whose grant
// is gone, gives another item or had its item changed, as by a sync, since what such an editor would save has nowhere
// to go or would overwrite the new item. The editors opened from one, on the item's modifiers and weapons, close with
// it the same way, since what they would save goes into the same data; see editorShell.discardAndCloseAll.
func (p *featuresPanel) closeStaleGrantItemEditors() {
	for _, d := range AllDockables() {
		owned, ok := d.AsPanel().Self.(Owned)
		if !ok {
			continue
		}
		if o, isGrant := owned.Owner().(*grantItemOwner); isGrant && o.editsFeaturesOf(p) && !o.current() {
			if closer, canDiscard := d.(interface{ discardAndCloseAll() }); canDiscard {
				closer.discardAndCloseAll()
			}
		}
	}
}

// editGrantedItem opens the normal editor on a copy of the grant's item, or brings forward the one already open. A
// grant with no item chosen yet has nothing to edit.
func (p *featuresPanel) editGrantedItem(g *gurps.Grant) {
	outer := p.AncestorOrSelf[Rebuildable]()
	if g.Key() == "" || outer == nil || Activate(func(d unison.Dockable) bool {
		owned, ok := d.AsPanel().Self.(Owned)
		if !ok {
			return false
		}
		o, isGrant := owned.Owner().(*grantItemOwner)
		return isGrant && o.editsFeaturesOf(p) && o.key == g.Key()
	}) {
		return
	}
	o := &grantItemOwner{Rebuildable: outer, panel: p, key: g.Key(), hash: grantItemHash(g)}
	// The item is the one of the grant's type, as for the grant itself; see grantItemHash.
	switch g.Type {
	case feature.GrantSkill:
		item := g.Skill.Clone(gurps.LibraryFile{}, nil, nil, gurps.Copy)
		o.item, o.commit = item, func(g *gurps.Grant) { g.Skill = item.Clone(gurps.LibraryFile{}, nil, nil, gurps.Copy) }
		EditSkill(o, item)
	case feature.GrantSpell:
		item := g.Spell.Clone(gurps.LibraryFile{}, nil, nil, gurps.Copy)
		o.item, o.commit = item, func(g *gurps.Grant) { g.Spell = item.Clone(gurps.LibraryFile{}, nil, nil, gurps.Copy) }
		EditSpell(o, item)
	default:
		item := g.Trait.Clone(gurps.LibraryFile{}, nil, nil, gurps.Copy)
		o.item, o.commit = item, func(g *gurps.Grant) { g.Trait = item.Clone(gurps.LibraryFile{}, nil, nil, gurps.Copy) }
		EditTrait(o, item)
	}
}
