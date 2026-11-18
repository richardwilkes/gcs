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
	"strconv"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantadjust"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantmerge"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// popupItems returns the items of the popup, in order.
func popupItems[T comparable](popup *unison.PopupMenu[T]) []T {
	items := make([]T, 0, popup.ItemCount())
	for i := range popup.ItemCount() {
		if item, ok := popup.ItemAt(i); ok {
			items = append(items, item)
		}
	}
	return items
}

// TestGrantRowControls checks the controls of an open grant row for each kind of grant: a popup of the kinds of
// adjustment its item allows, showing the phrase of the grant's kind; the amount, signed when it adds levels or points,
// named and explained for its kind, and whole for a skill's or spell's levels, where a trait's levels and any points
// may be fractional; and, only for a grant that sets its item's level, the collision popup, showing the grant's mode.
// A trait without levels has none of them.
func TestGrantRowControls(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	cases := []struct {
		name      string
		grantType feature.Type
		item      any
		kind      grantadjust.Kind
		merge     grantmerge.Mode
		phrase    string // the kind popup's text, empty for no kind popup and no amount field
		text      string // the amount field's text
		title     string
		fractions bool
		collision bool // whether the collision popup shows
	}{
		{name: "unleveled trait", grantType: feature.GrantTrait, item: newGrantTestTrait("Night Vision", 0)},
		{
			name: "trait setting its level", grantType: feature.GrantTrait, item: newGrantTestTrait("Magery", fxp.One),
			kind: grantadjust.SetLevel, merge: grantmerge.KeepHighest, phrase: "setting its level to", text: "2",
			title: "Level", fractions: true, collision: true,
		},
		{
			// A collision mode left from a level set means nothing for levels added, so no popup shows it.
			name: "trait adding levels", grantType: feature.GrantTrait, item: newGrantTestTrait("Fit", fxp.One),
			kind: grantadjust.AddLevels, merge: grantmerge.KeepHighest, phrase: "adjusting its level by", text: "+2",
			title: "Level Adjustment", fractions: true,
		},
		{
			name: "skill adding levels", grantType: feature.GrantSkill, item: newTestSkill("Brawling", fxp.One, nil),
			kind: grantadjust.AddLevels, merge: grantmerge.Ignore, phrase: "adjusting its level by", text: "+2",
			title: "Level Adjustment",
		},
		{
			name: "skill adding points", grantType: feature.GrantSkill, item: newTestSkill("Stealth", fxp.One, nil),
			kind: grantadjust.AddPoints, merge: grantmerge.KeepLowest, phrase: "adjusting its points by", text: "+2",
			title: "Point Adjustment", fractions: true,
		},
		{
			name: "spell setting its level", grantType: feature.GrantSpell, item: newTestSpell("Light", fxp.One, nil),
			kind: grantadjust.SetLevel, merge: grantmerge.KeepLowest, phrase: "setting its level to", text: "2",
			title: "Level", collision: true,
		},
	}
	features := make(gurps.Features, 0, len(cases))
	for _, one := range cases {
		g := gurps.NewGrant(one.grantType)
		g.SetItem(one.item, gurps.LibraryFile{})
		g.Amount = fxp.Two
		g.Adjust = one.kind
		g.Merge = one.merge
		features = append(features, g)
	}
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	for i, one := range cases {
		path := strconv.Itoa(i)
		key := func(name string) string { return path + ":" + name }
		screen.Do(func() { p.toggle(path) })
		var wnd *unison.Window
		var field *DecimalField
		screen.Do(func() {
			wnd = p.Window()
			if !one.collision {
				c.Nil(p.FindRefKey(key("merge")), "%s: no collision popup", one.name)
			} else if popup := featuresPopup[grantmerge.Mode](c, p, key("merge")); popup != nil {
				c.Equal([]grantmerge.Mode{grantmerge.Ignore, grantmerge.KeepHighest, grantmerge.KeepLowest},
					popupItems(popup), one.name)
				selected, _ := popup.Selected()
				c.Equal(one.merge, selected, "%s: the popup shows the grant's collision mode", one.name)
			}
			if one.phrase == "" {
				c.Nil(p.FindRefKey(key("adjust")), "%s: no kind popup", one.name)
				c.Nil(p.FindRefKey(key("amount")), "%s: no amount field", one.name)
				return
			}
			if popup := featuresPopup[grantadjust.Kind](c, p, key("adjust")); popup != nil {
				c.Equal(grantAt(t, p, i).AllowedAdjusts(), popupItems(popup), one.name)
				c.Equal(one.phrase, popup.Text(), one.name)
			}
			var ok bool
			field, ok = refKeySelf(p.AsPanel(), key("amount")).(*DecimalField)
			c.True(ok, "%s: an amount field", one.name)
			if field != nil {
				c.Equal(one.text, field.Text(), one.name)
				c.Equal(wrapTextForTooltip(grantAmountTooltip(grantAt(t, p, i))), tooltipText(field.Tooltip), one.name)
			}
		})
		if field == nil {
			continue
		}
		c.NotNil(screen.AccessibilityTree(wnd))
		node := screen.AccessibilityNodeFor(field)
		c.NotNil(node, "%s: the amount field is described", one.name)
		if node != nil {
			c.Equal(one.title, node.Name, one.name)
		}
		screen.Do(func() {
			field.SetText("1.5")
			if one.fractions {
				c.Equal(fxp.OneAndAHalf, grantAt(t, p, i).Amount, "%s: a fractional amount is taken", one.name)
			} else {
				c.NotEqual("", field.tooltipTextForValidation(), "%s: a fractional amount is invalid", one.name)
			}
		})
	}
}

// TestGrantKindChoiceKeepsTheAmount checks that choosing how a skill grant's amount adjusts the item keeps the amount
// as typed, less a fraction its levels can't hold, shows it with a sign only when it adds levels or points, and is a
// single undoable edit; and that choosing a kind that adds drops the collision mode of a level set along with its
// popup, which a level set chosen again shows with nothing kept.
func TestGrantKindChoiceKeepsTheAmount(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	g := gurps.NewGrant(feature.GrantSkill)
	g.SetItem(newTestSkill("Brawling", fxp.One, nil), gurps.LibraryFile{})
	// Not the amount typed below, so that undoing the typing in place of the choice of kind would show.
	g.Amount = fxp.Five
	features := gurps.Features{g}
	p, host := showFeaturesPanel(t, screen, entity, owner, &features, false)
	amountText := func() string {
		if field, ok := refKeySelf(p.AsPanel(), "0:amount").(*DecimalField); ok {
			return field.Text()
		}
		return "<none>"
	}
	choose := func(kind grantadjust.Kind) {
		if popup := featuresPopup[grantadjust.Kind](c, p, "0:adjust"); popup != nil {
			popup.Select(kind)
		}
	}
	screen.Do(func() { p.toggle("0") })
	screen.Do(func() {
		field, ok := refKeySelf(p.AsPanel(), "0:amount").(*DecimalField)
		if !ok {
			t.Fatal("the open skill grant row has no amount field")
		}
		field.SetText("1")
		choose(grantadjust.SetLevel)
	})
	screen.Do(func() {
		g = grantAt(t, p, 0)
		c.Equal(grantadjust.SetLevel, g.Adjust)
		c.Equal(fxp.One, g.Amount, "the amount is kept as typed")
		c.Equal("1", amountText(), "a level set has no sign")
		host.mgr.Undo()
	})
	screen.Do(func() {
		g = grantAt(t, p, 0)
		c.Equal(grantadjust.AddLevels, g.Adjust, "undo restores the kind")
		c.Equal(fxp.One, g.Amount, "and only the kind")
		choose(grantadjust.AddPoints)
	})
	screen.Do(func() {
		c.Equal("+1", amountText(), "points added are signed")
		choose(grantadjust.SetLevel)
	})
	screen.Do(func() {
		if popup := featuresPopup[grantmerge.Mode](c, p, "0:merge"); popup != nil {
			popup.Select(grantmerge.KeepHighest)
		}
	})
	screen.Do(func() {
		c.Equal(grantmerge.KeepHighest, grantAt(t, p, 0).Merge, "precondition: the level set keeps the highest")
		choose(grantadjust.AddLevels)
	})
	screen.Do(func() {
		c.Equal(grantmerge.Ignore, grantAt(t, p, 0).Merge, "levels added have no use for a collision mode")
		c.Nil(p.FindRefKey("0:merge"), "so its popup is hidden")
		choose(grantadjust.SetLevel)
	})
	screen.Do(func() {
		if popup := featuresPopup[grantmerge.Mode](c, p, "0:merge"); popup != nil {
			selected, _ := popup.Selected()
			c.Equal(grantmerge.Ignore, selected, "a level set chosen again shows the popup, with nothing kept")
		}
		choose(grantadjust.AddPoints)
	})
	screen.Do(func() {
		if field, ok := refKeySelf(p.AsPanel(), "0:amount").(*DecimalField); ok {
			field.SetText("1.5")
		}
		c.Equal(fxp.OneAndAHalf, grantAt(t, p, 0).Amount, "precondition: points may be fractional")
		choose(grantadjust.AddLevels)
	})
	screen.Do(func() {
		c.Equal(fxp.One, grantAt(t, p, 0).Amount, "a skill's levels drop the fraction, as the grant does")
		c.Equal("+1", amountText(), "so the field shows a valid amount")
	})
}

// TestGrantAmountTooltip checks what the tooltip of a grant's amount says for each kind of grant, and for a technique
// or Ritual Magic spell, whose levels count from its default.
func TestGrantAmountTooltip(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		grantType feature.Type
		item      any // the grant's item, nil for none
		kind      grantadjust.Kind
		want      string
	}{
		{
			grantType: feature.GrantTrait, kind: grantadjust.SetLevel,
			want: "The level the grant gives the trait. With Keep Highest or Keep Lowest, a trait already present is " +
				"held to at least or at most that level. A level below 1 does neither.",
		},
		{
			grantType: feature.GrantTrait, kind: grantadjust.AddLevels,
			want: "The levels added to a trait already present, or taken away when negative, never below level 0. On " +
				"its own, the grant gives the trait that level when it is 1 or more. A positive level below 1 does " +
				"nothing at all.",
		},
		{
			grantType: feature.GrantSkill, kind: grantadjust.AddLevels,
			want: "The levels added to a skill already present, or taken away when negative, never below level 0. On " +
				"its own, the grant gives the skill that level when it is 1 or more, counted in steps of points: " +
				"level 1 is what 1 point buys, level 2 what 2 points buy, level 3 what 4 points buy, and each level " +
				"after that 4 points more, or three times those points for a wildcard.",
		},
		{
			grantType: feature.GrantSpell, kind: grantadjust.AddPoints,
			want: "The points added to a spell already present, or taken away when negative, never below 0. On its " +
				"own, the grant gives the spell that many points when they are enough to buy a level: at least 1, " +
				"or 3 for a wildcard.",
		},
		{
			grantType: feature.GrantSkill, kind: grantadjust.SetLevel,
			want: "The skill level the grant gives the skill, in place of the level its points and difficulty would " +
				"give it; bonuses still apply on top. With Keep Highest or Keep Lowest, one already present is held " +
				"to at least or at most that level. A level below 1, or below what 1 point buys (3 for a wildcard), " +
				"does neither.",
		},
		{
			grantType: feature.GrantSkill, item: gurps.NewTechnique(nil, nil, "Karate"), kind: grantadjust.AddLevels,
			want: "The levels added to a technique already present, or taken away when negative, never below level 0. " +
				"On its own, the grant gives the technique that many levels above its default when it is 1 or more, " +
				"at 1 point a level, plus 1 point for a Hard technique.",
		},
		{
			grantType: feature.GrantSkill, item: gurps.NewTechnique(nil, nil, "Karate"), kind: grantadjust.AddPoints,
			want: "The points added to a technique already present, or taken away when negative, never below 0. On " +
				"its own, the grant gives the technique that many points when they are enough to buy a level: at " +
				"least 1, or 2 for a Hard technique.",
		},
		{
			grantType: feature.GrantSkill, item: gurps.NewTechnique(nil, nil, "Karate"), kind: grantadjust.SetLevel,
			want: "The skill level the grant gives the technique, in place of the level its points and difficulty " +
				"would give it; bonuses still apply on top, up to its limit, if any. With Keep Highest or Keep Lowest, " +
				"one already present is held to at least or at most that level. A level below 1, or below what 1 " +
				"point buys (2 for a Hard technique), does neither.",
		},
		{
			grantType: feature.GrantSpell, item: gurps.NewRitualMagicSpell(nil, nil, false), kind: grantadjust.AddLevels,
			want: "The levels added to a Ritual Magic spell already present, or taken away when negative, never below " +
				"level 0. On its own, the grant gives the Ritual Magic spell that many levels above its default when " +
				"it is 1 or more, at 1 point a level, plus 1 point for a Hard Ritual Magic spell.",
		},
	} {
		g := gurps.NewGrant(one.grantType)
		if one.item != nil {
			g.SetItem(one.item, gurps.LibraryFile{})
		}
		g.Adjust = one.kind
		c.Equal(one.want, grantAmountTooltip(g), "%s %s of %v", one.grantType, one.kind, one.item)
	}
}

// TestGrantLibraryItemRefusesWhatGrants checks that an item that grants something itself is not one a grant can give,
// whether the grant is among its own features or those of one of its modifiers, as SetItem strips grants from both,
// while a technique is.
func TestGrantLibraryItemRefusesWhatGrants(t *testing.T) {
	c := check.New(t)
	grant := func() *gurps.Grant {
		g := gurps.NewGrant(feature.GrantTrait)
		g.SetItem(newGrantTestTrait("Night Vision", 0), gurps.LibraryFile{})
		return g
	}
	plain := newGrantTestTrait("Plain", 0)
	plain.Modifiers = []*gurps.TraitModifier{gurps.NewTraitModifier(nil, nil, false)}
	c.NotNil(newGrantLibraryItem(plain, gurps.LibraryFile{}), "a trait whose modifier grants nothing is offered")
	direct := newGrantTestTrait("Direct", 0)
	direct.Features = gurps.Features{grant()}
	c.Nil(newGrantLibraryItem(direct, gurps.LibraryFile{}), "a trait that grants something is not")
	viaModifier := newGrantTestTrait("Via Modifier", 0)
	modifier := gurps.NewTraitModifier(nil, nil, false)
	modifier.Features = gurps.Features{grant()}
	viaModifier.Modifiers = []*gurps.TraitModifier{modifier}
	c.Nil(newGrantLibraryItem(viaModifier, gurps.LibraryFile{}), "nor is one whose modifier grants something")
	c.NotNil(newGrantLibraryItem(gurps.NewTechnique(nil, nil, "Karate"), gurps.LibraryFile{}), "a technique is offered")
}

// TestGrantRowOpensItsItemEditor checks that a click on the drop zone of a grant with no item gives it a blank one and
// opens the editor on that, a skill or spell grant first asking by menu whether a technique or Ritual Magic spell is
// wanted, and that a click on the zone once it names an item opens the editor on that item.
func TestGrantRowOpensItsItemEditor(t *testing.T) {
	for _, tc := range []struct {
		name      string
		grantType feature.Type
		item      any      // the item chosen, nil for none
		text      string   // the zone's text
		menu      []string // the titles of the menu a click on the zone opens, nil for none
		choose    string   // the title of the menu's item chosen
		kind      string   // the kind of item the editor opened is on, which starts its title
	}{
		{
			name: "blank", grantType: feature.GrantTrait,
			text: "Drop a trait here from a library or sheet, or click to start with a blank one.", kind: "Trait",
		},
		{
			name: "chosen", grantType: feature.GrantTrait, item: newGrantTestTrait("Night Vision", 0),
			text: "Night Vision", kind: "Trait",
		},
		{
			name: "blank technique", grantType: feature.GrantSkill,
			text: "Drop a skill or technique here from a library or sheet, or click to start with a blank one.",
			menu: []string{"New Skill", "New Technique"}, choose: "New Technique", kind: "Technique",
		},
		{
			name: "blank ritual magic spell", grantType: feature.GrantSpell,
			text: "Drop a spell here from a library or sheet, or click to start with a blank one.",
			menu: []string{"New Spell", "New Ritual Magic Spell"}, choose: "New Ritual Magic Spell",
			kind: "Ritual Magic Spell",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
			screen.EnableAccessibility()
			sheet, isSheet := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
			if !isSheet {
				t.Fatal("New Character Sheet must open a character sheet")
			}
			var p *featuresPanel
			screen.Do(func() {
				entity := sheet.Entity()
				training := gurps.NewTrait(entity, nil, false)
				training.Name = "Training"
				g := gurps.NewGrant(tc.grantType)
				if tc.item != nil {
					g.SetItem(tc.item, gurps.LibraryFile{})
				}
				training.Features = gurps.Features{g}
				entity.Traits = append(entity.Traits, training)
				sheet.Rebuild(true)
				p = uxtest.PanelsOfType[*featuresPanel](EditTrait(sheet, training).content)[0]
				if p.collapse.collapsed {
					p.collapse.toggle()
				}
				p.toggle("0")
			})
			var zone *unison.Button
			screen.Do(func() {
				if found := p.FindRefKey("0:drop"); found != nil {
					if b, isButton := found.Self.(*unison.Button); isButton {
						zone = b
					}
				}
			})
			if zone == nil {
				t.Fatal("the open grant row has no drop zone")
			}
			screen.Do(func() {
				c.Equal(tc.text, zone.Text.String())
				zone.ScrollIntoView()
			})
			screen.Click(screen.PanelCenter(zone))
			c.Equal(tc.menu, contextMenuTitles(t, screen, wnd), "the menu the click opens")
			if tc.choose != "" {
				chooseOpenMenuItem(t, screen, wnd, tc.choose)
			}
			var editors []unison.Dockable
			screen.Do(func() {
				editors = AllMatchingDockables(func(d unison.Dockable) bool {
					owned, ok := d.AsPanel().Self.(Owned)
					if !ok {
						return false
					}
					_, isGrant := owned.Owner().(*grantItemOwner)
					return isGrant
				})
			})
			if len(editors) != 1 {
				t.Fatalf("expected one editor on a grant's item, found %d", len(editors))
			}
			screen.Do(func() {
				o, ok := editors[0].AsPanel().Self.(Owned).Owner().(*grantItemOwner)
				c.True(ok)
				c.Equal(grantAt(t, p, 0).Key(), o.key, "the editor is on the item the grant holds")
				title := editors[0].Title()
				c.True(strings.HasPrefix(title, tc.kind+" Editor for "), "the editor is on a %s: %s", tc.kind, title)
				if tc.item == nil {
					undoTitle := "<no undo manager>"
					if mgr := unison.UndoManagerFor(p); mgr != nil {
						undoTitle = mgr.UndoTitle()
					}
					c.Equal("Undo New Blank Granted Item", undoTitle, "the blank item is one undoable edit")
				}
			})
		})
	}
}

// TestGrantedItemEditLandsOnTheGrant checks that an edit of a grant's item, made in its editor on a copy, goes to the
// grant the panel holds once its own undo has replaced the grant with a copy.
func TestGrantedItemEditLandsOnTheGrant(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	g := gurps.NewGrant(feature.GrantTrait)
	g.SetItem(newGrantTestTrait("Night Vision", fxp.One), gurps.LibraryFile{})
	features := gurps.Features{g}
	p, host := showFeaturesPanel(t, screen, entity, owner, &features, false)
	screen.Do(func() {
		item := g.Trait.Clone(gurps.LibraryFile{}, nil, nil, gurps.Copy)
		o := &grantItemOwner{panel: p, key: g.Key(), item: item, commit: func(g *gurps.Grant) {
			g.Trait = item.Clone(gurps.LibraryFile{}, nil, nil, gurps.Copy)
		}}
		p.edit("Level", "0:amount", "", func() { g.Amount = fxp.Two })
		host.mgr.Undo()
		item.Name = "Dark Vision"
		o.Rebuild(true)
		current, ok := features[0].(*gurps.Grant)
		c.True(ok)
		c.False(current == g, "undo put a copy of the grant in its place")
		c.Equal("Dark Vision", current.ItemName())
		c.Equal("Night Vision", g.ItemName(), "the grant undo replaced is left alone")
	})
}

// TestGrantItemEditorFollowsTheGrantsType checks that the editor opened on a grant's item, and the hash by which the
// panel tells that the item changed, go by the grant's type, as the grant itself does, when a loaded grant also holds
// an item of another kind.
func TestGrantItemEditorFollowsTheGrantsType(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	owner := gurps.NewTrait(entity, nil, false)
	owner.Name = "Training"
	g := gurps.NewGrant(feature.GrantSkill)
	g.SetItem(newTestSkill("Brawling", fxp.One, nil), gurps.LibraryFile{})
	skillOnly := grantItemHash(g)
	g.Trait = newGrantTestTrait("Night Vision", 0)
	c.Equal(skillOnly, grantItemHash(g), "a stray trait doesn't change which item the grant gives")
	owner.Features = gurps.Features{g}
	entity.Traits = []*gurps.Trait{owner}
	sheet.Rebuild(true)
	var skillEditors, traitEditors int
	screen.Do(func() {
		DisplayNewDockable(sheet)
		outer := EditTrait(sheet, owner)
		p := uxtest.PanelsOfType[*featuresPanel](outer.content)[0]
		p.editGrantedItem(grantAt(t, p, 0))
		skillEditors = len(AllMatchingDockables(func(d unison.Dockable) bool {
			_, ok := d.AsPanel().Self.(*editor[*gurps.Skill, *gurps.SkillEditData])
			return ok
		}))
		traitEditors = len(AllMatchingDockables(func(d unison.Dockable) bool {
			_, ok := d.AsPanel().Self.(*editor[*gurps.Trait, *gurps.TraitEditData])
			return ok
		}))
	})
	c.Equal(1, skillEditors, "the skill the grant gives is edited")
	c.Equal(1, traitEditors, "beside the editor of what carries the grant")
}

// TestGrantItemEditorClosesWhenItsGrantChanges checks that the editor open on a grant's item closes, without asking,
// once the outer editor's panel drops the grant or gives it another item, since what the editor would save has nowhere
// to go, or would overwrite the new item.
func TestGrantItemEditorClosesWhenItsGrantChanges(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(p *featuresPanel, g *gurps.Grant)
	}{
		{name: "deleted", change: func(p *featuresPanel, _ *gurps.Grant) {
			for _, entry := range p.moreEntries("0") {
				if entry.Label == "Delete" {
					entry.Act()
				}
			}
		}},
		{name: "another item", change: func(p *featuresPanel, g *gurps.Grant) {
			// As a drop on the grant's zone does.
			p.sentenceRows.restructure("Replace Granted Item", "0:more", "", func() string {
				g.SetItem(newGrantTestTrait("Dark Vision", 0), gurps.LibraryFile{})
				return grantFocusKey("0", g)
			})
		}},
		{name: "item changed", change: func(p *featuresPanel, g *gurps.Grant) {
			// As Sync Granted Item with Source does.
			p.edit("Sync Granted Item with Source", "0:more", "0:more", func() { g.Trait.Name = "Dark Vision" })
		}},
	} {
		// With deep set, a modified editor is open on one of the item's modifiers too, which closes along with the
		// item's editor, since what it would save has nowhere to go either.
		for _, deep := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s deep=%t", tc.name, deep), func(t *testing.T) {
				c := check.New(t)
				screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
				sheet := newTestSheetForTemplate(t)
				entity := sheet.Entity()
				training := gurps.NewTrait(entity, nil, false)
				training.Name = "Training"
				g := gurps.NewGrant(feature.GrantTrait)
				nightVision := newGrantTestTrait("Night Vision", 0)
				modifier := gurps.NewTraitModifier(nil, nil, false)
				modifier.Name = "Long"
				nightVision.Modifiers = []*gurps.TraitModifier{modifier}
				g.SetItem(nightVision, gurps.LibraryFile{})
				training.Features = gurps.Features{g}
				entity.Traits = []*gurps.Trait{training}
				sheet.Rebuild(true)
				isTraitEditor := func(d unison.Dockable) bool {
					_, ok := d.AsPanel().Self.(*editor[*gurps.Trait, *gurps.TraitEditData])
					return ok
				}
				isModifierEditor := func(d unison.Dockable) bool {
					_, ok := d.AsPanel().Self.(*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData])
					return ok
				}
				var outer *editor[*gurps.Trait, *gurps.TraitEditData]
				var p *featuresPanel
				screen.Do(func() {
					DisplayNewDockable(sheet)
					outer = EditTrait(sheet, training)
					p = uxtest.PanelsOfType[*featuresPanel](outer.content)[0]
					p.editGrantedItem(grantAt(t, p, 0))
				})
				nested := uxtest.OtherEditor(t, screen, outer, isTraitEditor)
				screen.Do(func() {
					nested.editorData.Name = "Changed"
					nested.MarkModified(nil)
				})
				if deep {
					screen.Do(func() { EditTraitModifier(nested, nested.editorData.Modifiers[0]) })
					modifierEditor := uxtest.SoleEditor[*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData]](t,
						screen, isModifierEditor)
					screen.Do(func() {
						modifierEditor.editorData.Name = "Longer"
						modifierEditor.MarkModified(nil)
					})
				}
				screen.Do(func() { tc.change(p, grantAt(t, p, 0)) })
				var open, modifierEditors, windows int
				screen.Do(func() {
					open = len(AllMatchingDockables(isTraitEditor))
					modifierEditors = len(AllMatchingDockables(isModifierEditor))
					windows = len(unison.Windows())
				})
				c.Equal(1, open, "the grant item's editor closes without asking, leaving the outer editor")
				c.Equal(0, modifierEditors, "as does the editor open on one of the item's modifiers")
				c.Equal(1, windows, "nothing asks about the changes: %v", wnd.Title())
			})
		}
	}
}
