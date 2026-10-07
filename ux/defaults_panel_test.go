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
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
)

// showDefaultsPanel shows a defaultsPanel for the defaults in a window, within a host, so that its rebuilds run and its
// edits can be undone. The panel is expanded, as most tests look at its rows; see TestDefaultsPanelStartingState for
// how it starts out.
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
		p = newDefaultsPanel(entity, owner, defaults)
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
	screen, _ := startHeadlessWorkspace(t, c)
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
	screen, _ := startHeadlessWorkspace(t, c)
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
// doesn't write, and checks that none of it changes the data, since opening an editor must not mark it modified, that
// a type written as "Skill" is shown as a skill default, and that an unknown type is shown and described as such.
func TestDefaultsPanelBuildingChangesNothing(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	defaults := append(newTestDefaults(), &gurps.SkillDefault{DefaultType: "Skill"},
		&gurps.SkillDefault{DefaultType: "unknown"})
	start := defaultsJSON(c, defaults)
	p, _ := showDefaultsPanel(t, screen, entity, nil, &defaults)
	screen.Do(func() { c.Equal(`Unknown type "unknown" at +0`, rowSentence(p, "4"), "an unknown type says so") })
	for i := range defaults {
		path := strconv.Itoa(i)
		screen.Do(func() { p.toggle(path) })
		screen.Do(func() { c.NotNil(p.FindRefKey(path + keyFirst)) })
		c.Equal(start, defaultsJSON(c, defaults), "opening %s changes nothing", path)
	}
	screen.Do(func() {
		if popup := defaultTypePopup(c, p, "4"); popup != nil {
			c.Equal(`Unknown type "unknown"`, popup.Text(), "the type popup shows an unknown type as such")
		}
		p.toggle("3")
	})
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
// becomes the one added next, unless the entity has no such attribute.
func TestDefaultsPanelAdd(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
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
}

// TestDefaultsPanelAddWithoutDX checks that a new default of an entity without DX takes a type the entity has rather
// than DX.
func TestDefaultsPanelAddWithoutDX(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	delete(entity.SheetSettings.Attributes.Set, gurps.DexterityID)
	c.Nil(gurps.AttributeDefsFor(entity).Set[gurps.DexterityID])
	c.NotNil(gurps.AttributeDefsFor(entity).Set[gurps.StrengthID])
	var defaults []*gurps.SkillDefault
	p, _ := showDefaultsPanel(t, screen, entity, nil, &defaults)
	defer func(last string) { lastDefaultTypeUsed = last }(lastDefaultTypeUsed)
	lastDefaultTypeUsed = gurps.DexterityID
	screen.Do(func() {
		button, ok := p.FindRefKey(defaultAddKey).Self.(*unison.Button)
		c.True(ok, "the section has an add button")
		button.ClickCallback()
	})
	c.Equal([]string{gurps.StrengthID}, defaultTypes(defaults), "the entity's first attribute takes the place of DX")
}

// TestDefaultsPanelTypeChange checks that a skill-based default keeps its criteria between the skill-based types, loses
// them on becoming an attribute default, so that a default set back to what it was is unchanged, that one becoming
// skill-based names its skill, so that its name field shows, and that a change of type is undone in one step.
func TestDefaultsPanelTypeChange(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	defaults := newTestDefaults()
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
}

// TestDefaultsPanelChips checks that the specialization, tags and tech level criteria are added as chips and removed
// again, each a step to undo, that a tech level starts at the character's whole tech level, and that an attribute
// default offers only the tech level.
func TestDefaultsPanelChips(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	// A fraction, which the whole-number field would show cut off while the sentence showed it whole.
	entity.Profile.TechLevel = "3.5"
	defaults := []*gurps.SkillDefault{
		{DefaultType: gurps.SkillID, Name: criteria.Text{Compare: criteria.IsText, Qualifier: "Brawling"}},
		{DefaultType: gurps.DexterityID},
	}
	p, host := showDefaultsPanel(t, screen, entity, nil, &defaults)
	click := func(key string) {
		screen.Do(func() {
			var buttons []*unison.Button
			if target := p.FindRefKey(key); target != nil {
				buttons = panelsOfType[*unison.Button](target)
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
}

// TestDefaultsPanelMoreMenu checks that Duplicate, Move up, Move down and Delete change the list as they say, each
// installing a new list rather than changing the one held before and each a step to undo, that the open row stays open
// wherever it goes, that the focus goes to the more button of the row the default ends up in, and that the moves are
// offered only where there is room.
func TestDefaultsPanelMoreMenu(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
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
	act("1", "Move Up")
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
	screen, _ := startHeadlessWorkspace(t, c)
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
// over, into a new list, that a drop is one step to undo and redo, that the open row follows its default, and that a
// drop onto or beside itself, or of a row dragged from another panel, changes nothing.
func TestDefaultsPanelDragAndDrop(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
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
	drag := func(from, onto string, fraction float32) bool { return dragFrom(p.AsPanel(), from, onto, fraction) }
	c.False(dragFrom(unison.NewPanel(), "0", "2", 0.8), "a row from another panel can't be dropped")
	c.False(drag("0", "0", 0.2), "a row can't be dropped onto itself")
	c.True(drag("1", "0", 0.8), "after the row above it")
	c.True(drag("1", "2", 0.2), "and before the row below it")
	c.Equal([]string{gurps.DexterityID, gurps.SkillID, gurps.ParryID}, defaultTypes(defaults), "but neither moves it")
	c.False(host.mgr.CanUndo(), "nor records a step")

	screen.Do(func() { p.toggle("1") })
	held := defaults
	c.True(drag("0", "2", 0.8))
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
}

// TestDefaultsPanelDragFromRow checks that a row can be dragged through the window by its sentence without also opening
// it, and that a click on a sentence still opens its row.
func TestDefaultsPanelDragFromRow(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	defaults := newTestDefaults()
	p, _ := showDefaultsPanel(t, screen, entity, nil, &defaults)
	var sentence, row *unison.Panel
	var below geom.Point
	screen.Do(func() {
		registerWindowDragTypes(p.Window())
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
}

// TestDefaultsPanelEmpty checks that a list with no defaults shows a placeholder before the add button, that clicking
// it adds a default of the type last chosen and opens it, as the add button does, and that undo brings it back.
func TestDefaultsPanelEmpty(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
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
	screen, _ := startHeadlessWorkspace(t, c)
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
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	defaults := newTestDefaults()
	var empty []*gurps.SkillDefault
	var full, open *defaultsPanel
	screen.Do(func() {
		full = newDefaultsPanel(entity, nil, &defaults)
		open = newDefaultsPanel(entity, nil, &empty)
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
	screen, _ := startHeadlessWorkspace(t, c)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
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
				if panels := panelsOfType[*defaultsPanel](d.AsPanel()); len(panels) == 1 {
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

// TestSkillEditorRowsFollowSubstitutions checks that the prerequisites, defaults and features of a skill editor read
// their nameable markers with the values the editor's data holds, and show the new ones once Set Substitutions changes
// them, but not when it is canceled.
func TestSkillEditorRowsFollowSubstitutions(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	bonus := gurps.NewSkillBonus()
	bonus.NameCriteria.Qualifier = "@Weapon@"
	var e *editor[*gurps.Skill, *gurps.SkillEditData]
	screen.Do(func() {
		entity := sheet.Entity()
		skill := gurps.NewSkill(entity, nil, false)
		skill.Replacements = map[string]string{"Weapon": "Spear"}
		skill.Defaults = []*gurps.SkillDefault{
			{DefaultType: gurps.SkillID, Name: criteria.Text{Compare: criteria.IsText, Qualifier: "@Weapon@"}},
		}
		skill.Features = gurps.Features{bonus}
		trait := gurps.NewTraitPrereq()
		trait.NameCriteria.Qualifier = "@Weapon@ Mastery"
		skill.Prereq = gurps.NewPrereqList()
		skill.Prereq.Prereqs = gurps.Prereqs{trait}
		entity.Skills = append(entity.Skills, skill)
		sheet.Rebuild(true)
		e = EditSkill(sheet, skill)
	})
	prereqs := func() string {
		var text string
		screen.Do(func() {
			for _, p := range panelsOfType[*prereqPanel](e.content) {
				text = collapsedSummary(p.AsPanel())
			}
		})
		return text
	}
	summaries := func() (defaults, features string) {
		screen.Do(func() {
			for _, p := range panelsOfType[*defaultsPanel](e.content) {
				defaults = collapsedSummary(p.AsPanel())
			}
			for _, p := range panelsOfType[*featuresPanel](e.content) {
				features = collapsedSummary(p.AsPanel())
			}
		})
		return defaults, features
	}
	defaults, features := summaries()
	c.Equal("Skill Spear at +0.", defaults, "the defaults take the skill's values")
	c.True(strings.Contains(features, "Spear"), "as do the features: %s", features)
	c.Equal("Has trait Spear Mastery.", prereqs(), "and the prerequisites")

	answer := func(value string, accept bool) {
		swapForTest(t, &promptForNameables, func(_ promptOperation, sections []nameablesSection) bool {
			for _, section := range sections {
				for k := range section.Nameables {
					section.Nameables[k] = value
				}
			}
			return accept
		})
	}
	answer("Rapier", false)
	screen.Do(e.nameablesButton.ClickCallback)
	c.Equal("Spear", e.editorData.Replacements["Weapon"], "Cancel leaves the editor's data alone")
	screen.Do(func() { c.False(e.isModified(), "and the editor unmodified") })
	defaults, features = summaries()
	c.Equal("Skill Spear at +0.", defaults, "and the rows as they were")
	c.True(strings.Contains(features, "Spear"), "features included: %s", features)

	answer("Rapier", true)
	screen.Do(e.nameablesButton.ClickCallback)
	c.Equal("Rapier", e.editorData.Replacements["Weapon"], "precondition: the editor's data takes the new value")
	defaults, features = summaries()
	c.Equal("Skill Rapier at +0.", defaults, "Set Substitutions shows the new value in the defaults")
	c.True(strings.Contains(features, "Rapier"), "and in the features: %s", features)
	c.Equal("Has trait Rapier Mastery.", prereqs(), "and in the prerequisites")
}

// TestDefaultsPanelTechLevelWithoutEntity checks that a tech level condition added where there is no character starts
// at 0.
func TestDefaultsPanelTechLevelWithoutEntity(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	defaults := []*gurps.SkillDefault{{DefaultType: gurps.DexterityID}}
	p, _ := showDefaultsPanel(t, screen, nil, nil, &defaults)
	screen.Do(func() { p.toggle("0") })
	screen.Do(func() {
		var buttons []*unison.Button
		if target := p.FindRefKey("0:add tl"); target != nil {
			buttons = panelsOfType[*unison.Button](target)
		}
		c.NotEqual(0, len(buttons), "an attribute default offers a tech level")
		if len(buttons) != 0 {
			buttons[len(buttons)-1].ClickCallback()
		}
	})
	c.Equal(criteria.Number{Compare: criteria.AtLeastNumber}, defaults[0].WhenTL)
}

// TestDefaultsPanelControlNamesDiffer checks that no two controls of an open row are announced by the same name, for
// a skill default with every optional criterion added, a parry default, and an attribute default with a tech level.
func TestDefaultsPanelControlNamesDiffer(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
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
	screen, _ := startHeadlessWorkspace(t, c)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
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
		if panels := panelsOfType[*defaultsPanel](e.content); len(panels) == 1 {
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

// TestSkillEditorSyncKeepsDefaultsAsTheyWere checks that syncing a skill editor with its source shows the defaults
// section as it was before, expanded with the same row open, and gives the focus back to the field in that row that
// held it.
func TestSkillEditorSyncKeepsDefaultsAsTheyWere(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	user := gurps.GlobalSettings().Libraries.User()
	lib := gurps.NewSkill(nil, nil, false)
	lib.Name = "Broadsword"
	lib.Defaults = newTestDefaults()
	libFile := gurps.LibraryFile{Library: user.Key(), Path: "Skills/Test" + gurps.SkillsExt}
	c.NoError(gurps.SaveSkills([]*gurps.Skill{lib}, filepath.Join(user.Path(false), filepath.FromSlash(libFile.Path))))
	var e *editor[*gurps.Skill, *gurps.SkillEditData]
	sections := func() *defaultsPanel {
		if panels := panelsOfType[*defaultsPanel](e.content); len(panels) == 1 {
			return panels[0]
		}
		return nil
	}
	var old *defaultsPanel
	screen.Do(func() {
		entity := sheet.Entity()
		local := lib.Clone(libFile, entity, nil, gurps.Reference)
		local.Name = "Broadsword (old)"
		entity.Skills = append(entity.Skills, local)
		sheet.Rebuild(true)
		e = EditSkill(sheet, local)
		old = sections()
	})
	if old == nil {
		t.Fatal("the skill editor must have a defaults section")
	}
	screen.Do(func() {
		c.True(old.collapse.collapsed, "precondition: the defaults start out collapsed")
		old.collapse.toggle()
	})
	screen.Do(func() { old.toggle("1") })
	var field *unison.Panel
	screen.Do(func() {
		if field = old.FindRefKey("1:name"); field != nil {
			field.RequestFocus()
		}
	})
	if field == nil {
		t.Fatal("the open default must have a name field")
	}

	chooseSyncWithSource(t, screen, wnd, e)
	screen.Do(func() {
		c.Equal("Broadsword", e.editorData.Name, "the skill is synced")
		synced := sections()
		if synced == nil || synced == old {
			t.Error("the sync must rebuild the defaults section")
			return
		}
		c.False(synced.collapse.collapsed, "the defaults stay expanded")
		c.Equal("1", synced.open, "with the same row open")
		rebuilt := synced.FindRefKey("1:name")
		c.True(rebuilt != nil && rebuilt != field && wnd.CurrentFocus() == rebuilt,
			"and the focus on its rebuilt name field")
	})
	screen.Click(screen.PanelCenter(e.cancelButton))
}

// TestDefaultsPanelBlankSpecialization checks that a specialization chip with no specialization yet shows that it picks
// a skill without one, as the row's sentence says once closed.
func TestDefaultsPanelBlankSpecialization(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
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

// TestSkillEditorSyncKeepsSubstitutions checks that substitutions set in a skill editor but not yet applied still show
// in its features and defaults once Sync with Source has made them again.
func TestSkillEditorSyncKeepsSubstitutions(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	user := gurps.GlobalSettings().Libraries.User()
	lib := gurps.NewSkill(nil, nil, false)
	lib.Name = "Fast-Draw (@Weapon@)"
	lib.Defaults = []*gurps.SkillDefault{
		{DefaultType: gurps.SkillID, Name: criteria.Text{Compare: criteria.IsText, Qualifier: "@Weapon@"}},
	}
	bonus := gurps.NewSkillBonus()
	bonus.NameCriteria.Qualifier = "@Weapon@"
	lib.Features = gurps.Features{bonus}
	libFile := gurps.LibraryFile{Library: user.Key(), Path: "Skills/Test" + gurps.SkillsExt}
	c.NoError(gurps.SaveSkills([]*gurps.Skill{lib}, filepath.Join(user.Path(false), filepath.FromSlash(libFile.Path))))
	var e *editor[*gurps.Skill, *gurps.SkillEditData]
	screen.Do(func() {
		entity := sheet.Entity()
		local := lib.Clone(libFile, entity, nil, gurps.Reference)
		local.Replacements = map[string]string{"Weapon": "Spear"}
		// Different from its source, so that the sync has something to do.
		local.Name = "Fast-Draw (old)"
		entity.Skills = append(entity.Skills, local)
		sheet.Rebuild(true)
		e = EditSkill(sheet, local)
	})
	swapForTest(t, &promptForNameables, func(_ promptOperation, sections []nameablesSection) bool {
		for _, section := range sections {
			for k := range section.Nameables {
				section.Nameables[k] = "Rapier"
			}
		}
		return true
	})
	screen.Do(e.nameablesButton.ClickCallback)
	var old *featuresPanel
	screen.Do(func() {
		if panels := panelsOfType[*featuresPanel](e.content); len(panels) == 1 {
			old = panels[0]
		}
	})
	chooseSyncWithSource(t, screen, wnd, e)
	screen.Do(func() {
		c.Equal("Fast-Draw (@Weapon@)", e.editorData.Name, "precondition: the skill is synced")
		c.Equal("Rapier", e.editorData.Replacements["Weapon"], "precondition: the sync keeps the substitutions")
		c.False(slices.Contains(panelsOfType[*featuresPanel](e.content), old), "precondition: the features are made again")
		for _, p := range panelsOfType[*defaultsPanel](e.content) {
			c.Equal("Skill Rapier at +0.", collapsedSummary(p.AsPanel()), "the defaults show them")
		}
		for _, p := range panelsOfType[*featuresPanel](e.content) {
			summary := collapsedSummary(p.AsPanel())
			c.True(strings.Contains(summary, "Rapier"), "as do the features: %s", summary)
		}
	})
}

// TestDefaultsPanelMoreButtonNames checks that a screen reader hears each row's more button named for the row's
// sentence, so that no two are alike, and that the name follows the sentence as it changes.
func TestDefaultsPanelMoreButtonNames(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
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
	screen, _ := startHeadlessWorkspace(t, c)
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

// TestWeaponEditorTakesItemSubstitutions checks that the defaults of a weapon edited from its skill's editor show the
// substitutions that editor holds, including those Set Substitutions makes while the weapon's editor is open.
func TestWeaponEditorTakesItemSubstitutions(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var e *editor[*gurps.Skill, *gurps.SkillEditData]
	var p *defaultsPanel
	screen.Do(func() {
		entity := sheet.Entity()
		skill := gurps.NewSkill(entity, nil, false)
		skill.Replacements = map[string]string{"Weapon": "Spear"}
		weapon := gurps.NewWeapon(skill, true)
		weapon.Defaults = []*gurps.SkillDefault{
			{DefaultType: gurps.SkillID, Name: criteria.Text{Compare: criteria.IsText, Qualifier: "@Weapon@"}},
		}
		skill.Weapons = []*gurps.Weapon{weapon}
		entity.Skills = append(entity.Skills, skill)
		sheet.Rebuild(true)
		e = EditSkill(sheet, skill)
		before := AllDockables()
		EditWeapon(e, e.editorData.Weapons[0])
		for _, d := range AllDockables() {
			if !slices.Contains(before, d) {
				if panels := panelsOfType[*defaultsPanel](d.AsPanel()); len(panels) == 1 {
					p = panels[0]
				}
			}
		}
	})
	if p == nil {
		t.Fatal("the weapon editor must have a defaults panel")
	}
	screen.Do(func() { c.Equal("Skill Spear at +0.", collapsedSummary(p.AsPanel())) })
	swapForTest(t, &promptForNameables, func(_ promptOperation, sections []nameablesSection) bool {
		for _, section := range sections {
			for k := range section.Nameables {
				section.Nameables[k] = "Rapier"
			}
		}
		return true
	})
	screen.Do(e.nameablesButton.ClickCallback)
	screen.Do(func() {
		c.Equal("Skill Rapier at +0.", collapsedSummary(p.AsPanel()), "the open weapon editor shows the new value")
	})
}
