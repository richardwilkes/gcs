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
	"encoding/json/v2"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/drag"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
)

// showDefaultsPanel shows a defaultsPanel for the defaults in a window, within a host, so that its rebuilds run and its
// edits can be undone. The panel is that of a spell when the owner is one. It is expanded, as most tests look at its
// rows; see TestDefaultsPanelStartingState for how it starts out.
func showDefaultsPanel(t *testing.T, screen *unison.HeadlessScreen, entity *gurps.Entity, owner nameable.Accesser, defaults *[]*gurps.SkillDefault) (*defaultsPanel, *sentenceUndoHost) {
	var p *defaultsPanel
	host := &sentenceUndoHost{mgr: unison.NewUndoManager(100, func(error) {})}
	screen.Do(func() {
		host.Self = host
		host.SetLayout(&unison.FlexLayout{Columns: 1})
		host.KeyDownCallback = func(keyCode unison.KeyCode, _ mod.Modifiers, _ bool) bool {
			if keyCode == unison.KeyEscape {
				host.escapes++
			}
			return true
		}
		_, ownerIsSpell := owner.(*gurps.Spell)
		p = newDefaultsPanel(entity, owner, defaults, ownerIsSpell)
		// Expanded ahead of showing, so that the window is sized for the rows.
		if p.collapse.collapsed {
			p.collapse.toggle()
		}
		host.AddChild(p)
	})
	showInTestWindow(t, screen, 900, host)
	return p, host
}

// newTestDefaults returns a DX default, a Skill default naming a specialized skill, and a Parry default.
func newTestDefaults() []*gurps.SkillDefault {
	return []*gurps.SkillDefault{
		{DefaultType: gurps.DexterityID, Modifier: -fxp.Five},
		{
			DefaultType:    gurps.SkillID,
			Name:           criteria.Text{Compare: criteria.IsText, Qualifier: "Broadsword"},
			Specialization: criteria.Text{Compare: criteria.IsText, Qualifier: "Fencing"},
			Modifier:       -fxp.Two,
		},
		{
			DefaultType: gurps.ParryID,
			Name:        criteria.Text{Compare: criteria.IsText, Qualifier: "Shortsword"},
		},
	}
}

// defaultTypes returns the types of the defaults.
func defaultTypes(list []*gurps.SkillDefault) []string {
	types := make([]string, 0, len(list))
	for _, one := range list {
		types = append(types, one.Type())
	}
	return types
}

// defaultsJSON returns the defaults as a file holds them, which is what tells an editor whether it is modified.
func defaultsJSON(c check.Checker, list []*gurps.SkillDefault) string {
	data, err := json.Marshal(list)
	c.NoError(err)
	return string(data)
}

// defaultTypePopup returns the type popup of the open row at the path, failing the test and returning nil if there is
// none.
func defaultTypePopup(c check.Checker, p *defaultsPanel, path string) *unison.PopupMenu[*gurps.AttributeChoice] {
	popup, ok := refKeySelf(p.AsPanel(), path+":type").(*unison.PopupMenu[*gurps.AttributeChoice])
	c.True(ok, "expected a type popup in row %s", path)
	if !ok {
		return nil
	}
	return popup
}

// chooseDefaultType picks the entry for the type in the type popup of the open row at the path, as a user's choice
// does.
func chooseDefaultType(c check.Checker, p *defaultsPanel, path, key string) {
	popup := defaultTypePopup(c, p, path)
	if popup == nil {
		return
	}
	for i := range popup.ItemCount() {
		if item, ok := popup.ItemAt(i); ok && item != nil && item.Key == key {
			popup.SelectIndex(i)
			return
		}
	}
	c.True(false, "the type popup has no %s entry", key)
}

// refKeySelf returns the Self of the panel within root that has the reference key, or nil if there is none.
func refKeySelf(root *unison.Panel, key string) any {
	if target := root.FindRefKey(key); target != nil {
		return target.Self
	}
	return nil
}

// actOnDefault runs the entry of the more menu of the default at the path with the label, failing the test if there is
// none.
func actOnDefault(c check.Checker, p *defaultsPanel, path, label string) {
	action := menuAction(p.moreEntries(path), label)
	c.NotNil(action, "%s offers %s", path, label)
	if action != nil {
		action()
	}
}

// rowSentence returns the text of the closed row at the path, or "" if it shows none.
func rowSentence(p *defaultsPanel, path string) string {
	if b, ok := refKeySelf(p.AsPanel(), path+keySentence).(*sentenceButton); ok {
		return b.plainText()
	}
	return ""
}

// defaultNameField returns the name field of the open row at the path, failing the test and returning nil if there is
// none.
func defaultNameField(c check.Checker, p *defaultsPanel, path string) *StringField {
	field, ok := refKeySelf(p.AsPanel(), path+":name").(*StringField)
	c.True(ok, "expected a name field in row %s", path)
	if !ok {
		return nil
	}
	return field
}

// TestDefaultsPanelSentences checks that each closed row reads as its default's description, which a screen reader
// hears as a disclosure, with a nameable marker showing the value its owner gives it, as it does in the paragraph of
// the collapsed panel.
func TestDefaultsPanelSentences(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewSkill(entity, nil, false)
	owner.Replacements = map[string]string{"Weapon": "Rapier"}
	defaults := append(newTestDefaults(), &gurps.SkillDefault{
		DefaultType: gurps.SkillID,
		Name:        criteria.Text{Compare: criteria.IsText, Qualifier: "@Weapon@"},
		WhenTL:      criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.Four},
	})
	p, _ := showDefaultsPanel(t, screen, entity, owner, &defaults)
	screen.Do(func() {
		for i, want := range []string{
			"DX at -5",
			"Skill Broadsword (Fencing) at -2",
			"Parry of skill Shortsword at +0",
			"Skill Rapier at +0, when the tech level is at least 4",
		} {
			sentence, ok := p.FindRefKey(strconv.Itoa(i) + keySentence).Self.(*sentenceButton)
			c.True(ok, "row %d is a sentence", i)
			if !ok {
				continue
			}
			c.Equal(want, sentence.plainText())
			c.Equal(want, sentence.Accessibility.Name)
			c.Equal(role.DisclosureTriangle, sentence.Accessibility.Role)
		}
		p.collapse.toggle()
	})
	screen.Do(func() {
		c.Equal("DX at -5; Skill Broadsword (Fencing) at -2; Parry of skill Shortsword at +0; Skill Rapier at +0, when "+
			"the tech level is at least 4.", collapsedSummary(p.AsPanel()))
	})
}

// TestDefaultsPanelOpenAndClose checks that a row opens to its editor, focusing its first field, that only one row is
// open at a time, and that Done and Escape close it, the latter without reaching the editor.
func TestDefaultsPanelOpenAndClose(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	defaults := newTestDefaults()
	p, host := showDefaultsPanel(t, screen, entity, nil, &defaults)
	screen.Do(func() { p.toggle("0") })
	screen.Do(func() {
		c.NotNil(p.FindRefKey("0" + keyFirst))
		c.Equal("0:modifier", p.Window().Focus().RefKey, "an attribute default's first field is its modifier")
		c.Nil(p.FindRefKey("0:name"), "and it has no name")
		p.toggle("1")
	})
	screen.Do(func() {
		c.Nil(p.FindRefKey("0"+keyFirst), "opening another row closes the first")
		c.Equal("1:name", p.Window().Focus().RefKey, "a skill default's first field is its name")
		done, ok := p.FindRefKey("1:done").Self.(*unison.Button)
		c.True(ok)
		done.ClickCallback()
	})
	screen.Do(func() {
		c.Equal("", p.open)
		c.Equal(p.FindRefKey("1"+keySentence), p.Window().Focus(), "closing gives the focus to the sentence")
		p.toggle("1")
	})
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal("", p.open, "Escape closes the open row")
	c.Equal(0, host.escapes, "and doesn't reach the editor")
}

// TestDefaultsPanelBuildingChangesNothing opens every row in turn, holding values a file can hold but the editor
// doesn't write, and checks that none of it changes the data, since opening an editor must not mark it modified, and
// that a type written as "Skill" is shown as a skill default.
func TestDefaultsPanelBuildingChangesNothing(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	defaults := append(newTestDefaults(), &gurps.SkillDefault{DefaultType: "Skill"},
		&gurps.SkillDefault{DefaultType: "unknown"})
	start := defaultsJSON(c, defaults)
	p, _ := showDefaultsPanel(t, screen, entity, nil, &defaults)
	for i := range defaults {
		path := strconv.Itoa(i)
		screen.Do(func() { p.toggle(path) })
		screen.Do(func() { c.NotNil(p.FindRefKey(path + keyFirst)) })
		c.Equal(start, defaultsJSON(c, defaults), "opening %s changes nothing", path)
	}
	screen.Do(func() { p.toggle("3") })
	screen.Do(func() {
		if popup := defaultTypePopup(c, p, "3"); popup != nil {
			c.Equal("Skill", popup.Text(), "a type written as \"Skill\" is a skill")
		}
		c.NotNil(p.FindRefKey("3:namecmp"), "with the criteria of one")
		c.NotNil(p.FindRefKey("3:add tag"))
	})
	c.Equal("Skill", defaults[3].DefaultType)
}

// TestDefaultsPanelAdd checks that the add button adds a default of the type last chosen at the end of the list, in a
// new list, opened with the focus in its first field, that a skill-based one names its skill, and that the new type
// becomes the one added next, unless the entity has no such attribute, when DX takes its place, or without DX, the
// entity's first attribute. A spell default added in a spell's panel names its spell and selects a spell of the same
// power source as that spell.
func TestDefaultsPanelAdd(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	// With room to grow, so that appending to it in place would show through the list held before.
	defaults := slices.Grow(newTestDefaults(), 4)
	held := defaults
	p, host := showDefaultsPanel(t, screen, entity, nil, &defaults)
	defer func(last string) { lastDefaultTypeUsed = last }(lastDefaultTypeUsed)
	lastDefaultTypeUsed = gurps.HealthID
	add := func() {
		screen.Do(func() {
			button, ok := p.FindRefKey(defaultAddKey).Self.(*unison.Button)
			c.True(ok, "the section has an add button")
			button.ClickCallback()
		})
	}
	add()
	c.Equal([]string{gurps.DexterityID, gurps.SkillID, gurps.ParryID, gurps.HealthID}, defaultTypes(defaults),
		"a default is added at the end")
	c.Nil(held[:len(held)+1][3], "of a new list, leaving the room the list held before had untouched")
	c.Equal("3", p.open, "and opens")
	screen.Do(func() { c.Equal("3:modifier", p.Window().Focus().RefKey) })
	c.Equal("Undo Add Default", host.mgr.UndoTitle())

	screen.Do(func() { chooseDefaultType(c, p, "3", gurps.SkillID) })
	c.Equal(gurps.SkillID, lastDefaultTypeUsed)
	add()
	c.Equal(gurps.SkillID, defaults[4].Type(), "the type chosen last is the one added")
	c.Equal(criteria.IsText, defaults[4].Name.Compare, "naming its skill")
	screen.Do(func() { c.Equal("4:name", p.Window().Focus().RefKey, "with the focus in the name") })
	screen.Do(host.mgr.Undo)
	c.Equal(4, len(defaults), "adding is undone in one step")
	c.Equal("3", p.open, "undo opens the row that was open")

	lastDefaultTypeUsed = "custom"
	add()
	c.Equal(gurps.DexterityID, defaults[4].Type(), "a type the entity has no attribute for gives way to DX")

	delete(entity.SheetSettings.Attributes.Set, gurps.DexterityID)
	c.Nil(gurps.AttributeDefsFor(entity).Set[gurps.DexterityID])
	c.NotNil(gurps.AttributeDefsFor(entity).Set[gurps.StrengthID])
	lastDefaultTypeUsed = gurps.DexterityID
	add()
	c.Equal(gurps.StrengthID, defaults[5].Type(), "the entity's first attribute takes the place of DX")

	var ofSpell []*gurps.SkillDefault
	q, _ := showDefaultsPanel(t, screen, entity, gurps.NewSpell(entity, nil, false), &ofSpell)
	lastDefaultTypeUsed = gurps.SpellID
	screen.Do(func() {
		button, ok := q.FindRefKey(defaultAddKey).Self.(*unison.Button)
		c.True(ok, "the spell's section has an add button")
		if ok {
			button.ClickCallback()
		}
	})
	c.True(len(ofSpell) == 1 && ofSpell[0].Type() == gurps.SpellID, "a spell default is added to a spell's panel")
	if len(ofSpell) == 1 {
		c.Equal(criteria.IsText, ofSpell[0].Name.Compare, "naming its spell")
		c.True(ofSpell[0].SamePowerSource, "and selecting a spell of the same power source")
	}
}

// TestDefaultsPanelTypeChange checks that a skill-based default keeps its criteria between the skill-based types, loses
// them on becoming an attribute default, so that a default set back to what it was is unchanged, that one becoming
// skill-based names its skill, so that its name field shows, and that a change of type is undone in one step. A skill
// default made a spell default keeps all but its specialization, and shows a power source in its place, one becoming
// spell-based names its spell, and in a spell's panel selects a spell of the same power source as that spell.
func TestDefaultsPanelTypeChange(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	defaults := append(newTestDefaults(), &gurps.SkillDefault{
		DefaultType:    gurps.SkillID,
		Name:           criteria.Text{Compare: criteria.IsText, Qualifier: "Fireball"},
		Specialization: criteria.Text{Compare: criteria.IsText, Qualifier: "Arcane"},
		Tags:           criteria.Text{Compare: criteria.IsText, Qualifier: "Fire"},
	})
	defaults[1].Tags = criteria.Text{Compare: criteria.IsText, Qualifier: "Melee"}
	p, host := showDefaultsPanel(t, screen, entity, nil, &defaults)
	defer func(last string) { lastDefaultTypeUsed = last }(lastDefaultTypeUsed)
	def := defaults[1]
	screen.Do(func() { p.toggle("1") })
	screen.Do(func() { chooseDefaultType(c, p, "1", gurps.BlockID) })
	c.Equal(gurps.BlockID, def.Type())
	c.Equal("Broadsword", def.Name.Qualifier, "a block default keeps the criteria")
	c.Equal("Melee", def.Tags.Qualifier)
	screen.Do(func() {
		c.NotNil(p.FindRefKey("1:name"), "and shows them")
		c.NotNil(p.FindRefKey("1:tag" + keyChip))
		chooseDefaultType(c, p, "1", gurps.DexterityID)
	})
	c.True(def.Name.IsZero() && def.Specialization.IsZero() && def.Tags.IsZero(),
		"an attribute default drops the criteria")
	screen.Do(func() {
		c.Nil(p.FindRefKey("1:name"), "and their controls")
		c.Nil(p.FindRefKey("1:add specialization"))
		c.Nil(p.FindRefKey("1:tag" + keyChip))
		c.NotNil(p.FindRefKey("1:add tl"), "but keeps the tech level")
		chooseDefaultType(c, p, "1", gurps.SkillID)
	})
	c.Equal(criteria.IsText, def.Name.Compare, "one that becomes skill-based names its skill")
	screen.Do(func() { c.NotNil(p.FindRefKey("1:name"), "so its name field shows") })

	parry := defaults[2]
	screen.Do(func() { p.toggle("2") })
	screen.Do(func() {
		if field := defaultNameField(c, p, "2"); field != nil {
			c.Equal("Shortsword", field.Text(), "a parry default shows the skill it names")
			field.SetText("Smallsword")
		}
	})
	c.Equal("Smallsword", parry.Name.Qualifier, "and edits it")
	screen.Do(func() { chooseDefaultType(c, p, "2", gurps.SkillID) })
	screen.Do(func() { chooseDefaultType(c, p, "2", gurps.ParryID) })
	c.Equal(gurps.ParryID, parry.Type())
	c.Equal(criteria.Text{Compare: criteria.IsText, Qualifier: "Smallsword"}, parry.Name,
		"a parry default made a skill default and back keeps its criteria")

	start := defaultsJSON(c, defaults)
	screen.Do(func() { p.toggle("0") })
	screen.Do(func() { chooseDefaultType(c, p, "0", gurps.SkillID) })
	screen.Do(func() {
		if field := defaultNameField(c, p, "0"); field != nil {
			field.SetText("Judo")
		}
	})
	c.Equal("Judo", defaults[0].Name.Qualifier, "precondition: the default names a skill")
	screen.Do(func() { chooseDefaultType(c, p, "0", gurps.DexterityID) })
	c.Equal(start, defaultsJSON(c, defaults), "switching back leaves nothing of the skill behind")
	c.Equal("Undo Default Type", host.mgr.UndoTitle())
	screen.Do(host.mgr.Undo)
	c.Equal(gurps.SkillID, defaults[0].Type(), "undo brings back the type it had")
	c.Equal("Judo", defaults[0].Name.Qualifier, "with its criteria")
	screen.Do(host.mgr.Redo)
	c.Equal(start, defaultsJSON(c, defaults), "and redo takes them away again")

	spell := defaults[3]
	screen.Do(func() { p.toggle("3") })
	screen.Do(func() { chooseDefaultType(c, p, "3", gurps.SpellID) })
	c.Equal(criteria.Text{Compare: criteria.IsText, Qualifier: "Fireball"}, spell.Name, "a spell default keeps the name")
	c.Equal("Fire", spell.Tags.Qualifier, "and the tags")
	c.True(spell.Specialization.IsZero(), "but not the specialization, which a spell lacks")
	c.False(spell.SamePowerSource, "and outside a spell's panel selects no power source")
	screen.Do(func() {
		c.NotNil(p.FindRefKey("3:name"), "its name shows")
		for _, label := range uxtest.PanelsOfType[*unison.Label](p.FindRefKey("3" + keyFirst)) {
			c.NotEqual("of skill", label.String(), "with no words joining it to a skill")
		}
		c.Nil(p.FindRefKey("3:add specialization"), "and a power source in place of the specialization")
		c.NotNil(p.FindRefKey("3:add power"))
		p.toggle("0")
	})
	screen.Do(func() { chooseDefaultType(c, p, "0", gurps.SpellID) })
	c.Equal(criteria.IsText, defaults[0].Name.Compare, "one that becomes spell-based names its spell")

	ofSpell := []*gurps.SkillDefault{{DefaultType: gurps.DexterityID}}
	q, _ := showDefaultsPanel(t, screen, entity, gurps.NewSpell(entity, nil, false), &ofSpell)
	screen.Do(func() { q.toggle("0") })
	screen.Do(func() { chooseDefaultType(c, q, "0", gurps.SpellID) })
	c.True(ofSpell[0].SamePowerSource, "in a spell's panel, it selects a spell of the same power source")
}

// TestDefaultsPanelChips checks that the specialization, tags and tech level criteria are added as chips and removed
// again, each a step to undo, that a tech level starts at the character's whole tech level, and that an attribute
// default offers only the tech level. A spell default's power source starts as an "is" comparison, with no offer of
// the same power source as the spell's but in a spell's panel, where it starts as that.
func TestDefaultsPanelChips(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	// A fraction, which the whole-number field would show cut off while the sentence showed it whole.
	entity.Profile.TechLevel = "3.5"
	lore := criteria.Text{Compare: criteria.IsText, Qualifier: "Lore"}
	defaults := []*gurps.SkillDefault{
		{DefaultType: gurps.SkillID, Name: criteria.Text{Compare: criteria.IsText, Qualifier: "Brawling"}},
		{DefaultType: gurps.DexterityID},
		{DefaultType: gurps.SpellID, Name: lore},
	}
	p, host := showDefaultsPanel(t, screen, entity, nil, &defaults)
	click := func(key string) {
		screen.Do(func() {
			var buttons []*unison.Button
			if target := p.FindRefKey(key); target != nil {
				buttons = uxtest.PanelsOfType[*unison.Button](target)
			}
			c.NotEqual(0, len(buttons), "expected a button within %s", key)
			if len(buttons) != 0 {
				buttons[len(buttons)-1].ClickCallback()
			}
		})
	}
	screen.Do(func() { p.toggle("0") })
	click("0:add specialization")
	c.Equal(criteria.IsText, defaults[0].Specialization.Compare, "adding the chip adds the criterion")
	screen.Do(func() { c.Equal("0:specializationcmp", p.Window().Focus().RefKey, "and focuses its comparison") })
	c.Equal("Undo Add Specialization", host.mgr.UndoTitle(), "as a step to undo")
	screen.Do(host.mgr.Undo)
	c.True(defaults[0].Specialization.IsZero(), "which takes the criterion away again")
	screen.Do(host.mgr.Redo)
	screen.Do(func() {
		field, ok := refKeySelf(p.AsPanel(), "0:specialization").(*StringField)
		c.True(ok, "the chip has a field")
		if ok {
			field.SetText("Boxing")
		}
		p.toggle("0")
	})
	screen.Do(func() {
		c.Equal("Skill Brawling (Boxing) at +0", rowSentence(p, "0"), "the row reads as the criterion says")
		p.toggle("0")
	})
	click("0:specialization" + keyChip)
	c.True(defaults[0].Specialization.IsZero(), "removing the chip removes it")
	c.Equal("Undo Remove Specialization", host.mgr.UndoTitle())
	screen.Do(host.mgr.Undo)
	c.Equal("Boxing", defaults[0].Specialization.Qualifier, "undo brings it back")
	screen.Do(host.mgr.Redo)
	click("0:add tag")
	c.Equal(criteria.IsText, defaults[0].Tags.Compare)
	click("0:tag" + keyChip)
	c.True(defaults[0].Tags.IsZero())
	click("0:add tl")
	c.Equal(criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.Three}, defaults[0].WhenTL,
		"a tech level starts at the character's, as a whole number")
	screen.Do(func() { c.NotNil(p.FindRefKey("0:tlcmp"), "and shows its comparison") })
	click("0:tl" + keyChip)
	c.Equal(criteria.AnyNumber, defaults[0].WhenTL.Compare)

	screen.Do(func() { p.toggle("1") })
	screen.Do(func() {
		c.Nil(p.FindRefKey("1:add specialization"), "an attribute default has no specialization")
		c.Nil(p.FindRefKey("1:add tag"), "nor tags")
		c.NotNil(p.FindRefKey("1:add tl"), "but has a tech level")
	})

	same := "and whose power source is the same as this spell's"
	screen.Do(func() { p.toggle("2") })
	click("2:add power")
	c.Equal(criteria.Text{Compare: criteria.IsText}, defaults[2].PowerSource,
		"a spell default's power source starts as an \"is\" comparison")
	c.False(defaults[2].SamePowerSource)
	screen.Do(func() {
		popup, ok := refKeySelf(p.AsPanel(), "2:powercmp").(*unison.PopupMenu[string])
		c.True(ok, "the power source chip has a comparison")
		if ok {
			c.Equal(-1, popup.IndexOfItem(same), "which doesn't offer the spell's own outside a spell's panel")
		}
	})
	click("2:power" + keyChip)
	c.True(defaults[2].PowerSource.IsZero(), "removing the chip removes the power source")

	ofSpell := []*gurps.SkillDefault{{DefaultType: gurps.SpellID, Name: lore}}
	q, _ := showDefaultsPanel(t, screen, entity, gurps.NewSpell(entity, nil, false), &ofSpell)
	screen.Do(func() { q.toggle("0") })
	screen.Do(func() {
		var buttons []*unison.Button
		if target := q.FindRefKey("0:add power"); target != nil {
			buttons = uxtest.PanelsOfType[*unison.Button](target)
		}
		c.NotEqual(0, len(buttons), "a spell's panel offers a power source")
		if len(buttons) != 0 {
			buttons[len(buttons)-1].ClickCallback()
		}
	})
	c.True(ofSpell[0].SamePowerSource, "which starts as the same as the spell's")
}

// TestDefaultsPanelMoreMenu checks that Duplicate, Move up, Move down and Delete, which a row's more button offers,
// change the list as they say, each installing a new list rather than changing the one held before and each a step to
// undo, that the open row stays open wherever it goes, that the focus goes to the more button of the row the default
// ends up in, and that the moves are offered only where there is room.
func TestDefaultsPanelMoreMenu(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	defaults := newTestDefaults()
	p, host := showDefaultsPanel(t, screen, entity, nil, &defaults)
	var held []*gurps.SkillDefault
	act := func(path, label string) {
		held = defaults
		screen.Do(func() { actOnDefault(c, p, path, label) })
	}
	focus := func() string {
		var key string
		screen.Do(func() { key = p.Window().Focus().RefKey })
		return key
	}
	undoRedo := func(before, after []string) {
		t.Helper()
		screen.Do(host.mgr.Undo)
		c.Equal(before, defaultTypes(defaults), "undo puts the list back")
		screen.Do(host.mgr.Redo)
		c.Equal(after, defaultTypes(defaults), "and redo makes the change again")
	}
	screen.Do(func() {
		c.Nil(menuAction(p.moreEntries("0"), "Move Up"), "nothing above the top")
		c.Nil(menuAction(p.moreEntries("2"), "Move Down"), "nothing below the bottom")
		p.toggle("1")
	})
	start := []string{gurps.DexterityID, gurps.SkillID, gurps.ParryID}
	// The first is chosen from the menu the row's more button opens, as a user would.
	held = defaults
	var wnd *unison.Window
	var more *unison.Panel
	screen.Do(func() {
		wnd = p.Window()
		more = p.FindRefKey("1" + keyMore)
	})
	screen.Click(screen.PanelCenter(more))
	chooseOpenMenuItem(t, screen, wnd, "Move Up")
	moved := []string{gurps.SkillID, gurps.DexterityID, gurps.ParryID}
	c.Equal(moved, defaultTypes(defaults))
	c.Equal(start, defaultTypes(held), "the list held before is left as it was")
	c.Equal("0", p.open, "the open row moves with its default")
	c.Equal("0"+keyMore, focus(), "the moved row's more button takes the focus")
	c.Equal("Undo Move Up", host.mgr.UndoTitle())
	undoRedo(start, moved)
	act("0", "Move Down")
	c.Equal(start, defaultTypes(defaults))
	c.Equal(moved, defaultTypes(held), "the list held before is left as it was")
	c.Equal("1", p.open)
	c.Equal("1"+keyMore, focus())
	c.Equal("Undo Move Down", host.mgr.UndoTitle())
	undoRedo(moved, start)
	skill := defaults[1]
	act("1", "Duplicate")
	duplicated := []string{gurps.DexterityID, gurps.SkillID, gurps.SkillID, gurps.ParryID}
	c.Equal(duplicated, defaultTypes(defaults))
	c.Equal(start, defaultTypes(held), "the list held before is left as it was")
	c.True(defaults[1] == skill && defaults[2] != skill, "the copy follows the original")
	c.Equal(defaultsJSON(c, defaults[1:2]), defaultsJSON(c, defaults[2:3]), "and is the same")
	c.Equal("2"+keyMore, focus(), "the copy's more button takes the focus")
	c.Equal("1", p.open, "and the original stays open")
	c.Equal("Undo Duplicate Default", host.mgr.UndoTitle())
	undoRedo(start, duplicated)
	act("0", "Delete")
	c.Equal([]string{gurps.SkillID, gurps.SkillID, gurps.ParryID}, defaultTypes(defaults))
	c.Equal(duplicated, defaultTypes(held), "the list held before keeps every default, with no nil left at its end")
	c.Equal("0", p.open, "the open row keeps its place as rows above it go")
	c.Equal("Undo Delete Default", host.mgr.UndoTitle())
	act("0", "Delete")
	c.Equal("", p.open, "deleting the open row leaves none open")
	act("1", "Delete")
	act("0", "Delete")
	c.Equal(0, len(defaults))
	c.Equal(defaultAddKey, focus(), "deleting the last gives the add button the focus")
}

// TestDefaultsPanelUndo checks that typing in a field is recorded as one snapshot of the list, that a change to the
// list's shape is another, and that undo and redo install copies of their snapshots, which later changes to the list
// don't reach.
func TestDefaultsPanelUndo(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	defaults := newTestDefaults()
	p, host := showDefaultsPanel(t, screen, entity, nil, &defaults)
	selectAll := func(key string) {
		screen.Do(func() {
			field, ok := refKeySelf(p.AsPanel(), key).(Selectable)
			c.True(ok, "expected a field at %s", key)
			if ok {
				p.FindRefKey(key).RequestFocus()
				field.SetSelection(0, 1000)
			}
		})
	}
	screen.Do(func() { p.toggle("1") })
	selectAll("1:name")
	screen.Type("Saber")
	c.Equal("Saber", defaults[1].Name.Qualifier)
	c.Equal("Undo Name", host.mgr.UndoTitle())
	screen.Do(func() { actOnDefault(c, p, "2", "Delete") })
	c.Equal(2, len(defaults))
	c.Equal("Undo Delete Default", host.mgr.UndoTitle())
	screen.Do(host.mgr.Undo)
	c.Equal([]string{gurps.DexterityID, gurps.SkillID, gurps.ParryID}, defaultTypes(defaults))
	c.Equal("Saber", defaults[1].Name.Qualifier)
	// A change made to the installed defaults outside of an edit, which a snapshot sharing them would take on.
	defaults[1].Name.Qualifier = "Changed"
	screen.Do(host.mgr.Redo)
	screen.Do(host.mgr.Undo)
	c.Equal("Saber", defaults[1].Name.Qualifier, "undo installs a copy of its snapshot, which keeps its own values")
	screen.Do(host.mgr.Undo)
	c.Equal("Broadsword", defaults[1].Name.Qualifier, "the typing was one edit")
	c.False(host.mgr.CanUndo())
	screen.Do(func() {
		field, ok := refKeySelf(p.AsPanel(), "1:name").(*StringField)
		c.True(ok, "the row the typing was done in is open")
		if ok {
			c.Equal("Broadsword", field.Text())
			c.Equal(field.AsPanel(), field.Window().Focus(), "and the field has the focus")
		}
	})
	screen.Do(host.mgr.Redo)
	c.Equal("Saber", defaults[1].Name.Qualifier)
	screen.Do(host.mgr.Redo)
	c.Equal([]string{gurps.DexterityID, gurps.SkillID}, defaultTypes(defaults))

	screen.Do(func() { p.toggle("0") })
	selectAll("0:modifier")
	screen.Type("-3")
	c.Equal(-fxp.Three, defaults[0].Modifier)
	c.Equal("Undo Modifier", host.mgr.UndoTitle())
	screen.Do(host.mgr.Undo)
	c.Equal(-fxp.Five, defaults[0].Modifier, "a modifier typed is one step to undo")
	screen.Do(host.mgr.Redo)
	c.Equal(-fxp.Three, defaults[0].Modifier, "and to redo")
}

// TestDefaultsPanelDragAndDrop checks that a dragged default goes before or after a row by which half of the row it is
// over, into a new list, that a drop is one step to undo and redo, that the open row follows its default, that a drop
// onto or beside itself, or of a row dragged from a panel of another kind, changes nothing, and that a row dragged from
// another defaults panel is copied in where it is dropped, or at the end of a collapsed panel, which expands, unless
// the panel is read-only.
func TestDefaultsPanelDragAndDrop(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	defaults := newTestDefaults()
	p, host := showDefaultsPanel(t, screen, entity, nil, &defaults)
	dragFrom := func(source *unison.Panel, from, onto string, fraction float32) (accepted bool) {
		screen.Do(func() {
			target := p.FindRefKey(onto + keyMore).Parent()
			r := p.RectFromRoot(target.RectToRoot(target.ContentRect(true)))
			where := geom.NewPoint(r.X+r.Width/3, r.Y+r.Height*fraction)
			data := &rowDrag{panel: source, path: from}
			p.dragOver(where, data)
			accepted = p.dropTarget != nil
			p.drop(where, data)
		})
		return accepted
	}
	dragRow := func(from, onto string, fraction float32) bool { return dragFrom(p.AsPanel(), from, onto, fraction) }
	c.False(dragFrom(unison.NewPanel(), "0", "2", 0.8), "a row from a panel of another kind can't be dropped")
	c.False(dragRow("0", "0", 0.2), "a row can't be dropped onto itself")
	c.True(dragRow("1", "0", 0.8), "after the row above it")
	c.True(dragRow("1", "2", 0.2), "and before the row below it")
	c.Equal([]string{gurps.DexterityID, gurps.SkillID, gurps.ParryID}, defaultTypes(defaults), "but neither moves it")
	c.False(host.mgr.CanUndo(), "nor records a step")

	screen.Do(func() { p.toggle("1") })
	held := defaults
	c.True(dragRow("0", "2", 0.8))
	c.Equal([]string{gurps.SkillID, gurps.ParryID, gurps.DexterityID}, defaultTypes(defaults))
	c.Equal([]string{gurps.DexterityID, gurps.SkillID, gurps.ParryID}, defaultTypes(held),
		"the list held before is left as it was")
	c.Equal("0", p.open, "the open row moves with its default")
	c.Equal("Undo Move Default", host.mgr.UndoTitle())
	screen.Do(host.mgr.Undo)
	c.Equal([]string{gurps.DexterityID, gurps.SkillID, gurps.ParryID}, defaultTypes(defaults),
		"a drop is one step to undo")
	c.Equal("1", p.open)
	screen.Do(host.mgr.Redo)
	c.Equal([]string{gurps.SkillID, gurps.ParryID, gurps.DexterityID}, defaultTypes(defaults), "and to redo")

	otherDefaults := []*gurps.SkillDefault{{DefaultType: gurps.IntelligenceID, Modifier: -fxp.Three}}
	var other *defaultsPanel
	screen.Do(func() { other = newDefaultsPanel(entity, nil, &otherDefaults, false) })
	otherJSON := defaultsJSON(c, otherDefaults)
	modified := host.modified
	c.True(dragFrom(other.AsPanel(), "0", "0", 0.2), "a row from another defaults panel can be dropped")
	c.Equal([]string{gurps.IntelligenceID, gurps.SkillID, gurps.ParryID, gurps.DexterityID}, defaultTypes(defaults),
		"and goes where it is dropped")
	c.True(len(defaults) == 4 && defaults[0] != otherDefaults[0], "as a copy")
	c.Equal(otherJSON, defaultsJSON(c, otherDefaults), "leaving the source as it was")
	c.Equal("1", p.open, "the open row follows its default")
	c.Equal("Undo Copy Default", host.mgr.UndoTitle())
	c.Equal(modified+1, host.modified, "a copy marks the editor modified once")
	screen.Do(host.mgr.Undo)
	c.Equal([]string{gurps.SkillID, gurps.ParryID, gurps.DexterityID}, defaultTypes(defaults),
		"a copy is one step to undo")
	screen.Do(host.mgr.Redo)
	c.Equal([]string{gurps.IntelligenceID, gurps.SkillID, gurps.ParryID, gurps.DexterityID}, defaultTypes(defaults),
		"and to redo")
	c.True(dragFrom(other.AsPanel(), "0", "1", 0.8))
	c.Equal([]string{gurps.IntelligenceID, gurps.SkillID, gurps.IntelligenceID, gurps.ParryID, gurps.DexterityID},
		defaultTypes(defaults), "a copy dropped on the lower half of a row goes after it")

	screen.Do(p.collapse.toggle)
	accepted := false
	screen.Do(func() {
		summary := p.FindRefKey(sectionSummaryKey)
		c.NotNil(summary, "precondition: the collapsed panel shows its paragraph")
		if summary == nil {
			return
		}
		where := p.RectFromRoot(summary.RectToRoot(summary.ContentRect(true))).Center()
		data := &rowDrag{panel: other.AsPanel(), path: "0"}
		p.dragOver(where, data)
		accepted = p.dropTarget != nil
		p.drop(where, data)
	})
	c.True(accepted, "a copy can be dropped on a collapsed panel")
	c.Equal([]string{
		gurps.IntelligenceID, gurps.SkillID, gurps.IntelligenceID, gurps.ParryID, gurps.DexterityID,
		gurps.IntelligenceID,
	}, defaultTypes(defaults), "where it goes at the end")
	c.False(p.collapse.collapsed, "expanding the panel")

	screen.Do(p.makeReadOnly)
	c.False(dragFrom(other.AsPanel(), "0", "0", 0.2), "a read-only panel takes no copies")
	c.Equal(6, len(defaults))
}

// TestDefaultsPanelDragFromRow checks that a row can be dragged through the window by its sentence without also opening
// it, that a click on a sentence still opens its row, that a row dragged by its grip into the defaults panel of another
// window is copied there, with the drag reported as a copy, even from a read-only panel, whose grips offer only that
// and whose own rows don't move, and that a read-only panel refuses such a drag.
func TestDefaultsPanelDragFromRow(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	defaults := newTestDefaults()
	p, host := showDefaultsPanel(t, screen, entity, nil, &defaults)
	var sentence, row *unison.Panel
	var below geom.Point
	screen.Do(func() {
		RegisterWindowDragTypes(p.Window())
		sentence = p.FindRefKey("0" + keySentence)
		row = p.FindRefKey("2" + keyMore).Parent()
		below = geom.NewPoint(40, row.FrameRect().Height*0.8)
	})
	screen.Drag(screen.PanelCenter(sentence), screen.PanelPoint(row, below), 10)
	c.Equal([]string{gurps.SkillID, gurps.ParryID, gurps.DexterityID}, defaultTypes(defaults))
	c.Equal("", p.open, "the drag didn't also click")
	screen.Do(func() { sentence = p.FindRefKey("0" + keySentence) })
	screen.Click(screen.PanelCenter(sentence))
	c.Equal("0", p.open)

	// The second window sits below the first, so that neither hides the other from the drag.
	var others []*gurps.SkillDefault
	q, qHost := showDefaultsPanel(t, screen, entity, nil, &others)
	var pw, qw *unison.Window
	var grip *DragHandle
	var target *unison.Panel
	placeBelow := func() {
		frame := qw.FrameRect()
		frame.X, frame.Y = pw.FrameRect().X, pw.FrameRect().Bottom()+40
		qw.SetFrameRect(frame)
		c.False(pw.FrameRect().Intersects(qw.FrameRect()), "precondition: the windows don't overlap")
		c.True(qw.FrameRect().Bottom() <= 900, "precondition: the second window is on the screen")
	}
	gripOf := func(path string) *DragHandle {
		grips := uxtest.PanelsOfType[*DragHandle](p.FindRefKey(path + keyMore).Parent())
		c.Equal(1, len(grips), "precondition: the row has a grip")
		return grips[0]
	}
	inView := func(grip *DragHandle) {
		c.True(screen.PanelCenter(grip).In(pw.FrameRect()), "precondition: the grip is in view")
	}
	screen.Do(func() {
		pw, qw = p.Window(), q.Window()
		// The open row made the first window's content taller than the window.
		pw.Pack()
		placeBelow()
		RegisterWindowDragTypes(qw)
		grip = gripOf("1")
		target = q.FindRefKey(defaultEmptyKey)
	})
	inView(grip)
	c.True(strings.Contains(tooltipText(grip.Tooltip), "copy"), "the grip says a row can be copied")
	before := defaultsJSON(c, defaults)
	modified := host.modified
	screen.Drag(screen.PanelCenter(grip), screen.PanelCenter(target), 10)
	result := screen.LastDrag()
	c.Equal(drag.Copy, result.Op, "a row over the defaults of another window is copied")
	c.True(result.Dropped)
	c.True(result.Source == pw)
	c.True(result.Target == qw)
	c.Equal([]string{gurps.ParryID}, defaultTypes(others), "the copy lands in the other panel")
	c.Equal(defaultsJSON(c, defaults[1:2]), defaultsJSON(c, others), "equal to the dragged default")
	c.True(len(others) == 1 && others[0] != defaults[1], "but not the same one")
	c.Equal(before, defaultsJSON(c, defaults), "the source is left as it was")
	c.Equal(modified, host.modified, "and its editor isn't marked modified")
	c.Equal("Undo Copy Default", qHost.mgr.UndoTitle())
	c.Equal(1, qHost.modified)

	screen.Do(func() {
		p.makeReadOnly()
		qw.Pack()
		placeBelow()
		grip = gripOf("2")
		target = q.FindRefKey("0" + keyMore).Parent()
	})
	inView(grip)
	tip := tooltipText(grip.Tooltip)
	c.True(strings.Contains(tip, "copy") && !strings.Contains(tip, "rearrange"),
		"the grip of a read-only panel offers a copy but no rearranging: %s", tip)
	screen.Drag(screen.PanelCenter(grip), screen.PanelPoint(target, geom.NewPoint(40, 4)), 10)
	c.Equal([]string{gurps.DexterityID, gurps.ParryID}, defaultTypes(others), "a row is copied out of a read-only panel")
	c.Equal(before, defaultsJSON(c, defaults), "which is left as it was")
	screen.Do(func() {
		grip = gripOf("1")
		target = p.FindRefKey("2" + keyMore).Parent()
	})
	screen.Drag(screen.PanelCenter(grip), screen.PanelPoint(target, below), 10)
	result = screen.LastDrag()
	c.Equal(drag.None, result.Op, "a read-only panel's own rows can't be rearranged")
	c.False(result.Dropped)
	c.Equal(before, defaultsJSON(c, defaults))

	screen.Do(func() {
		q.makeReadOnly()
		qw.Pack()
		placeBelow()
		grip = gripOf("2")
		target = q.FindRefKey("0" + keyMore).Parent()
	})
	inView(grip)
	screen.Drag(screen.PanelCenter(grip), screen.PanelPoint(target, geom.NewPoint(40, 4)), 10)
	result = screen.LastDrag()
	c.Equal(drag.None, result.Op, "a read-only panel refuses the drag")
	c.False(result.Dropped)
	c.Equal([]string{gurps.DexterityID, gurps.ParryID}, defaultTypes(others))
}

// TestDefaultsPanelEmpty checks that a list with no defaults shows a placeholder before the add button, that clicking
// it adds a default of the type last chosen and opens it, as the add button does, and that undo brings it back.
func TestDefaultsPanelEmpty(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	var defaults []*gurps.SkillDefault
	p, host := showDefaultsPanel(t, screen, entity, nil, &defaults)
	defer func(last string) { lastDefaultTypeUsed = last }(lastDefaultTypeUsed)
	lastDefaultTypeUsed = gurps.IntelligenceID
	var empty *unison.Button
	screen.Do(func() {
		var ok bool
		empty, ok = p.FindRefKey(defaultEmptyKey).Self.(*unison.Button)
		c.True(ok, "an empty list shows a placeholder")
		if !ok {
			return
		}
		c.Equal("No defaults. Click here to add one.", empty.Text.String())
		add := p.FindRefKey(defaultAddKey)
		c.NotNil(add)
		r, a := empty.RectToRoot(empty.ContentRect(true)), add.RectToRoot(add.ContentRect(true))
		c.True(a.X >= r.Right(), "the add button follows it")
		c.Equal(r.CenterY(), a.CenterY(), "on its line")
	})
	if empty == nil {
		return
	}
	screen.Do(empty.ClickCallback)
	c.Equal([]string{gurps.IntelligenceID}, defaultTypes(defaults), "clicking it adds a default of the last type")
	c.Equal("0", p.open, "and opens it")
	screen.Do(func() {
		c.Equal("0:modifier", p.Window().Focus().RefKey, "with the focus in its first field")
		c.Nil(p.FindRefKey(defaultEmptyKey), "and the placeholder goes")
	})
	screen.Do(host.mgr.Undo)
	c.Equal(0, len(defaults))
	screen.Do(func() { c.NotNil(p.FindRefKey(defaultEmptyKey), "undo brings the placeholder back") })
}

// TestDefaultsPanelCollapse checks that the title bar collapses the panel to a paragraph listing every row's
// description, joined by semicolons, and expands it again, as does the paragraph, from a click or the keyboard, that a
// screen reader hears whether it is expanded, and that it is no edit and keeps the open row open.
func TestDefaultsPanelCollapse(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	defaults := newTestDefaults()
	p, host := showDefaultsPanel(t, screen, entity, nil, &defaults)
	start := defaultsJSON(c, defaults)
	expanded := func() bool {
		node := &accessibility.Node{}
		p.collapse.Accessibility.Callback(node)
		c.True(node.Expandable)
		return node.Expanded
	}
	screen.Do(func() {
		c.Equal("Defaults", p.collapse.Accessibility.Name, "the title bar is named for the title")
		c.True(expanded(), "precondition: the panel is expanded")
		p.toggle("1")
	})
	screen.Do(func() { p.collapse.MouseUpCallback(geom.Point{X: 1, Y: 1}, 0, mod.None) })
	screen.Do(func() {
		c.False(expanded(), "clicking the title bar collapses the panel")
		for _, key := range []string{"1" + keyFirst, "0" + keySentence, defaultAddKey} {
			c.Nil(p.FindRefKey(key), "collapsing hides %s", key)
		}
		paragraph, ok := p.FindRefKey(sectionSummaryKey).Self.(*sentenceButton)
		c.True(ok, "a collapsed panel shows a paragraph")
		if ok {
			c.Equal("DX at -5; Skill Broadsword (Fencing) at -2; Parry of skill Shortsword at +0.", paragraph.plainText())
		}
		c.Equal(p.collapse.AsPanel(), p.Window().Focus(), "the focus moves from the rows to the title bar")
	})
	screen.KeyPress(unison.KeySpace, mod.None)
	screen.Do(func() {
		c.True(expanded(), "Space expands the panel")
		c.NotNil(p.FindRefKey("1"+keyFirst), "with the open row still open")
	})
	screen.KeyPress(unison.KeyReturn, mod.None)
	screen.Do(func() {
		c.False(expanded(), "Return collapses it")
		p.FindRefKey(sectionSummaryKey).RequestFocus()
	})
	screen.KeyPress(unison.KeySpace, mod.None)
	screen.Do(func() { c.True(expanded(), "the paragraph expands the panel") })
	c.Equal(start, defaultsJSON(c, defaults), "collapsing changes nothing")
	c.False(host.mgr.CanUndo(), "and is not an edit")
	c.Equal(0, host.modified, "nor marks the editor modified")
}

// TestDefaultsPanelStartingState checks that a panel with defaults starts out collapsed and one without starts out open
// with its placeholder, and that a collapsed panel with no defaults says so.
func TestDefaultsPanelStartingState(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	defaults := newTestDefaults()
	var empty []*gurps.SkillDefault
	var full, open *defaultsPanel
	screen.Do(func() {
		full = newDefaultsPanel(entity, nil, &defaults, false)
		open = newDefaultsPanel(entity, nil, &empty, false)
	})
	showInTestWindow(t, screen, 900, full, open)
	screen.Do(func() {
		c.True(full.collapse.collapsed, "a panel with defaults starts out collapsed")
		c.NotNil(full.FindRefKey(sectionSummaryKey), "showing the paragraph")
		c.Nil(full.FindRefKey("0"+keySentence), "in place of the rows")
		c.False(open.collapse.collapsed, "a panel without defaults starts out open")
		c.NotNil(open.FindRefKey(defaultEmptyKey), "showing the placeholder")
		open.collapse.toggle()
	})
	screen.Do(func() {
		b, ok := open.FindRefKey(sectionSummaryKey).Self.(*sentenceButton)
		c.True(ok, "a collapsed panel shows a paragraph")
		if ok {
			c.Equal("No defaults.", b.plainText())
		}
		c.Nil(open.FindRefKey(defaultEmptyKey), "in place of the placeholder")
	})
}

// TestDefaultsPanelWeapon checks that a weapon editor's defaults panel reads the weapon's defaults, with the nameable
// values of the weapon's owner, and edits them.
func TestDefaultsPanelWeapon(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var p *defaultsPanel
	screen.Do(func() {
		entity := sheet.Entity()
		trait := gurps.NewTrait(entity, nil, false)
		trait.Replacements = map[string]string{"Weapon": "Spear"}
		weapon := gurps.NewWeapon(trait, true)
		weapon.Defaults = []*gurps.SkillDefault{
			{DefaultType: gurps.DexterityID, Modifier: -fxp.Four},
			{DefaultType: gurps.SkillID, Name: criteria.Text{Compare: criteria.IsText, Qualifier: "@Weapon@"}},
		}
		trait.Weapons = []*gurps.Weapon{weapon}
		entity.Traits = append(entity.Traits, trait)
		sheet.Rebuild(true)
		before := AllDockables()
		EditWeapon(sheet, weapon)
		for _, d := range AllDockables() {
			if !slices.Contains(before, d) {
				if panels := uxtest.PanelsOfType[*defaultsPanel](d.AsPanel()); len(panels) == 1 {
					p = panels[0]
				}
			}
		}
	})
	if p == nil {
		t.Fatal("the weapon editor must have a defaults panel")
	}
	screen.Do(func() {
		c.Equal("DX at -4; Skill Spear at +0.", collapsedSummary(p.AsPanel()))
		actOnDefault(c, p, "0", "Delete")
	})
	c.Equal([]string{gurps.SkillID}, defaultTypes(*p.defaults), "the weapon's defaults are edited")
}

// collapsedSummary returns the paragraph a collapsed panel of sentence rows shows, or "" if it shows none.
func collapsedSummary(p *unison.Panel) string {
	if b, ok := refKeySelf(p, sectionSummaryKey).(*sentenceButton); ok {
		return b.plainText()
	}
	return ""
}

// TestDefaultsPanelTechLevelWithoutEntity checks that a tech level condition added where there is no character starts
// at 0.
func TestDefaultsPanelTechLevelWithoutEntity(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	defaults := []*gurps.SkillDefault{{DefaultType: gurps.DexterityID}}
	p, _ := showDefaultsPanel(t, screen, nil, nil, &defaults)
	screen.Do(func() { p.toggle("0") })
	screen.Do(func() {
		var buttons []*unison.Button
		if target := p.FindRefKey("0:add tl"); target != nil {
			buttons = uxtest.PanelsOfType[*unison.Button](target)
		}
		c.NotEqual(0, len(buttons), "an attribute default offers a tech level")
		if len(buttons) != 0 {
			buttons[len(buttons)-1].ClickCallback()
		}
	})
	c.Equal(criteria.Number{Compare: criteria.AtLeastNumber}, defaults[0].WhenTL)
}

// TestDefaultsPanelControlNamesDiffer checks that no two controls of an open row are announced by the same name, for
// a skill default and a spell default with every optional criterion added, a parry default, and an attribute default
// with a tech level.
func TestDefaultsPanelControlNamesDiffer(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	defaults := []*gurps.SkillDefault{
		{
			DefaultType:    gurps.SkillID,
			Name:           criteria.Text{Compare: criteria.IsText, Qualifier: "Broadsword"},
			Specialization: criteria.Text{Compare: criteria.IsText, Qualifier: "Fencing"},
			Tags:           criteria.Text{Compare: criteria.IsText, Qualifier: "Melee"},
			WhenTL:         criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.Three},
		},
		{DefaultType: gurps.ParryID, Name: criteria.Text{Compare: criteria.ContainsText, Qualifier: "sword"}},
		{
			DefaultType: gurps.DexterityID,
			WhenTL:      criteria.Number{Compare: criteria.AtMostNumber, Qualifier: fxp.Four},
		},
		{
			DefaultType: gurps.SpellID,
			Name:        criteria.Text{Compare: criteria.IsText, Qualifier: "Fireball"},
			PowerSource: criteria.Text{Compare: criteria.IsText, Qualifier: "Arcane"},
			Tags:        criteria.Text{Compare: criteria.IsText, Qualifier: "Fire"},
			WhenTL:      criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.Three},
		},
	}
	p, _ := showDefaultsPanel(t, screen, entity, nil, &defaults)
	for i := range defaults {
		path := strconv.Itoa(i)
		screen.Do(func() { p.toggle(path) })
		var controls []*unison.Panel
		screen.Do(func() {
			if editor := p.FindRefKey(path + keyFirst); editor != nil {
				editor.HasInSelfOrDescendants(func(one *unison.Panel) bool {
					if one.Focusable() {
						controls = append(controls, one)
					}
					return false
				})
			}
		})
		c.NotNil(screen.AccessibilityTree(p.Window()))
		c.True(len(controls) > 2, "row %s has controls", path)
		names := make(map[string]bool)
		for _, one := range controls {
			node := screen.AccessibilityNodeFor(one)
			c.True(node != nil && node.Name != "", "row %s: every control has a name", path)
			if node != nil && node.Name != "" {
				c.False(names[node.Name], "row %s: %s is shared", path, node.Name)
				names[node.Name] = true
			}
		}
	}
}

// TestSkillEditorDefaultTypeRoundTrip checks that a skill editor whose default is made a skill default, given a name,
// and made an attribute default again is unmodified, as it is after another type is chosen and then the original one,
// which the file wrote in another case.
func TestSkillEditorDefaultTypeRoundTrip(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var loaded gurps.SkillDefault
	c.NoError(json.Unmarshal([]byte(`{"type":"DX","modifier":-5}`), &loaded))
	var e *editor[*gurps.Skill, *gurps.SkillEditData]
	var p *defaultsPanel
	screen.Do(func() {
		entity := sheet.Entity()
		skill := gurps.NewSkill(entity, nil, false)
		skill.Defaults = []*gurps.SkillDefault{&loaded}
		entity.Skills = append(entity.Skills, skill)
		sheet.Rebuild(true)
		e = EditSkill(sheet, skill)
		if panels := uxtest.PanelsOfType[*defaultsPanel](e.content); len(panels) == 1 {
			p = panels[0]
			p.collapse.toggle()
		}
	})
	if p == nil {
		t.Fatal("the skill editor must have a defaults panel")
	}
	screen.Do(func() { p.toggle("0") })
	screen.Do(func() { chooseDefaultType(c, p, "0", gurps.SkillID) })
	screen.Do(func() {
		if field := defaultNameField(c, p, "0"); field != nil {
			field.SetText("Judo")
		}
	})
	screen.Do(func() { c.True(e.isModified(), "precondition: a skill default named Judo is a change") })
	screen.Do(func() { chooseDefaultType(c, p, "0", gurps.DexterityID) })
	screen.Do(func() { c.False(e.isModified(), "made a DX default again, the editor is unmodified") })
	screen.Do(func() { chooseDefaultType(c, p, "0", gurps.IntelligenceID) })
	screen.Do(func() { chooseDefaultType(c, p, "0", gurps.DexterityID) })
	screen.Do(func() { c.False(e.isModified(), "as it is after choosing another type and then DX") })
}

// TestDefaultsPanelBlankSpecialization checks that a specialization chip with no specialization yet shows that it picks
// a skill without one, as the row's sentence says once closed.
func TestDefaultsPanelBlankSpecialization(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	defaults := []*gurps.SkillDefault{{
		DefaultType:    gurps.SkillID,
		Name:           criteria.Text{Compare: criteria.IsText, Qualifier: "Guns"},
		Specialization: criteria.Text{Compare: criteria.IsText},
	}}
	p, _ := showDefaultsPanel(t, screen, entity, nil, &defaults)
	screen.Do(func() { c.Equal("Skill Guns without a specialization at +0", rowSentence(p, "0")) })
	screen.Do(func() { p.toggle("0") })
	screen.Do(func() {
		field, ok := refKeySelf(p.AsPanel(), "0:specialization").(*StringField)
		c.True(ok, "the chip has a field")
		if ok {
			c.Equal("none", field.Watermark, "whose hint says the default picks a skill without one")
		}
	})
}

// TestDefaultsPanelMoreButtonNames checks that a screen reader hears each row's more button named for the row's
// sentence, so that no two are alike, and that the name follows the sentence as it changes.
func TestDefaultsPanelMoreButtonNames(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	defaults := newTestDefaults()
	p, _ := showDefaultsPanel(t, screen, entity, nil, &defaults)
	name := func(path string) string {
		var button *unison.Panel
		screen.Do(func() { button = p.FindRefKey(path + keyMore) })
		c.NotNil(screen.AccessibilityTree(p.Window()))
		if node := screen.AccessibilityNodeFor(button); node != nil {
			return node.Name
		}
		return ""
	}
	c.Equal("More actions for DX at -5", name("0"))
	c.Equal("More actions for Skill Broadsword (Fencing) at -2", name("1"))
	c.Equal("More actions for Parry of skill Shortsword at +0", name("2"))
	screen.Do(func() { p.toggle("1") })
	screen.Do(func() {
		if field := defaultNameField(c, p, "1"); field != nil {
			field.SetText("Saber")
		}
	})
	c.Equal("More actions for Skill Saber (Fencing) at -2", name("1"), "the name follows an edit to the open row")
}

// TestDefaultsPanelUnusualTypes checks that a type the popup doesn't offer for new defaults, such as Dodge, a number,
// one GCS doesn't know or none at all from a file, shows in the popup as the row's sentence names it, and can be chosen
// again after another type has been.
func TestDefaultsPanelUnusualTypes(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	defaults := []*gurps.SkillDefault{
		{DefaultType: gurps.DodgeID, Modifier: -fxp.Two},
		{DefaultType: "12"},
		{DefaultType: gurps.SizeModifierID},
		{},
	}
	p, _ := showDefaultsPanel(t, screen, entity, nil, &defaults)
	for i, want := range []string{"Dodge", "12", `Unknown type "sm"`, "No type"} {
		path := strconv.Itoa(i)
		screen.Do(func() { c.Equal(want+" at "+defaults[i].Modifier.StringWithSign(), rowSentence(p, path)) })
		screen.Do(func() { p.toggle(path) })
		screen.Do(func() {
			if popup := defaultTypePopup(c, p, path); popup != nil {
				c.Equal(want, popup.Text(), "the popup names the type as the sentence does")
			}
		})
	}
	screen.Do(func() { p.toggle("0") })
	screen.Do(func() { chooseDefaultType(c, p, "0", gurps.DexterityID) })
	c.Equal(gurps.DexterityID, defaults[0].Type())
	screen.Do(func() { chooseDefaultType(c, p, "0", gurps.DodgeID) })
	c.Equal(gurps.DodgeID, defaults[0].Type(), "the type it had can be chosen again")
}
