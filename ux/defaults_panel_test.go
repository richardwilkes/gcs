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
	"github.com/zeebo/xxh3"
)

// selectPopupIndex selects the given index and then runs the popup's selection callback once, mirroring what happens
// when the user picks that entry. The callback is invoked directly rather than through SelectIndex, which wraps it in
// unison.SafeCall and would swallow a failure inside it, so it is taken off the popup while the selection is made.
func selectPopupIndex[T comparable](popup *unison.PopupMenu[T], index int) {
	callback := popup.SelectionChangedCallback
	popup.SelectionChangedCallback = nil
	popup.SelectIndex(index)
	popup.SelectionChangedCallback = callback
	callback(popup)
}

// showDefaultsPanel shows a defaultsPanel for the defaults in a window, within a host, so that its rebuilds run and its
// edits can be undone. The panel is expanded, as most tests look at its rows; see TestDefaultsPanelStartingState for
// how it starts out.
func showDefaultsPanel(t *testing.T, screen *unison.HeadlessScreen, entity *gurps.Entity, owner nameable.Accesser, defaults *[]*gurps.SkillDefault) (*defaultsPanel, *prereqUndoHost) {
	var p *defaultsPanel
	host := &prereqUndoHost{mgr: unison.NewUndoManager(100, func(error) {})}
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

// defaultsHash returns a hash of the defaults.
func defaultsHash(list []*gurps.SkillDefault) uint64 {
	h := xxh3.New()
	for _, one := range list {
		one.Hash(h)
	}
	return h.Sum64()
}

// defaultTypePopup returns the type popup of the open row at the path, failing the test if there is none.
func defaultTypePopup(c check.Checker, p *defaultsPanel, path string) *unison.PopupMenu[*gurps.AttributeChoice] {
	popup, ok := p.FindRefKey(path + ":type").Self.(*unison.PopupMenu[*gurps.AttributeChoice])
	c.True(ok, "expected a type popup in row %s", path)
	return popup
}

// chooseDefaultType picks the entry for the type in the type popup of the open row at the path, as a user's choice
// does.
func chooseDefaultType(c check.Checker, p *defaultsPanel, path, key string) {
	popup := defaultTypePopup(c, p, path)
	for i := range popup.ItemCount() {
		if item, ok := popup.ItemAt(i); ok && item != nil && item.Key == key {
			popup.SelectIndex(i)
			return
		}
	}
	c.True(false, "the type popup has no %s entry", key)
}

// plainDescription returns the description of the default without emphasis.
func plainDescription(entity *gurps.Entity, def *gurps.SkillDefault) string {
	return def.Describe(entity, nil, func(s string) string { return s })
}

// TestDefaultsPanelSentences checks that each closed row reads as its default's description, which a screen reader
// hears as a disclosure, with a nameable marker showing the value its owner gives it.
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
// doesn't write, and checks that none of it changes the data, since opening an editor must not mark it modified, and
// that a type written as "Skill" is shown as a skill default.
func TestDefaultsPanelBuildingChangesNothing(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	defaults := append(newTestDefaults(), &gurps.SkillDefault{DefaultType: "Skill"},
		&gurps.SkillDefault{DefaultType: "unknown"})
	hash := defaultsHash(defaults)
	p, _ := showDefaultsPanel(t, screen, entity, nil, &defaults)
	for i := range defaults {
		path := strconv.Itoa(i)
		screen.Do(func() { p.toggle(path) })
		screen.Do(func() { c.NotNil(p.FindRefKey(path + keyFirst)) })
		c.Equal(hash, defaultsHash(defaults), "opening %s changes nothing", path)
	}
	screen.Do(func() { p.toggle("3") })
	screen.Do(func() {
		c.Equal("Skill", defaultTypePopup(c, p, "3").Text(), "a type written as \"Skill\" is a skill")
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
	defaults := newTestDefaults()
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
	c.Equal(3, len(held), "of a new list")
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

// TestDefaultsPanelTypeChange checks that a skill-based default keeps its criteria between the skill-based types, loses
// them on becoming an attribute default, so that a default set back to what it was is unchanged, and that one becoming
// skill-based names its skill, so that its name field shows.
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

	start := defaultsHash(defaults)
	screen.Do(func() { p.toggle("0") })
	screen.Do(func() { chooseDefaultType(c, p, "0", gurps.SkillID) })
	screen.Do(func() {
		field, ok := p.FindRefKey("0:name").Self.(*StringField)
		c.True(ok, "precondition: a skill default has a name field")
		if ok {
			field.SetText("Judo")
		}
	})
	c.Equal("Judo", defaults[0].Name.Qualifier, "precondition: the default names a skill")
	screen.Do(func() { chooseDefaultType(c, p, "0", gurps.DexterityID) })
	c.Equal(start, defaultsHash(defaults), "switching back leaves nothing of the skill behind")
	c.True(host.mgr.CanUndo())
}

// TestDefaultsPanelChips checks that the specialization, tags and tech level criteria are added as chips and removed
// again, that a tech level starts at the character's whole tech level, and that an attribute default offers only the
// tech level.
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
	p, _ := showDefaultsPanel(t, screen, entity, nil, &defaults)
	click := func(key string) {
		screen.Do(func() {
			buttons := panelsOfType[*unison.Button](p.FindRefKey(key))
			c.NotEqual(0, len(buttons), "expected a button within %s", key)
			if len(buttons) != 0 {
				buttons[len(buttons)-1].ClickCallback()
			}
		})
	}
	def := defaults[0]
	screen.Do(func() { p.toggle("0") })
	click("0:add specialization")
	c.Equal(criteria.IsText, def.Specialization.Compare, "adding the chip adds the criterion")
	screen.Do(func() { c.Equal("0:specializationcmp", p.Window().Focus().RefKey, "and focuses its comparison") })
	screen.Do(func() {
		field, ok := p.FindRefKey("0:specialization").Self.(*StringField)
		c.True(ok, "the chip has a field")
		if ok {
			field.SetText("Boxing")
		}
	})
	c.Equal("Skill Brawling (Boxing) at +0", plainDescription(entity, def))
	click("0:specialization" + keyChip)
	c.True(def.Specialization.IsZero(), "removing the chip removes it")
	click("0:add tag")
	c.Equal(criteria.IsText, def.Tags.Compare)
	click("0:tag" + keyChip)
	c.True(def.Tags.IsZero())
	click("0:add tl")
	c.Equal(criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.Three}, def.WhenTL,
		"a tech level starts at the character's, as a whole number")
	screen.Do(func() { c.NotNil(p.FindRefKey("0:tlcmp"), "and shows its comparison") })
	click("0:tl" + keyChip)
	c.Equal(criteria.AnyNumber, def.WhenTL.Compare)

	screen.Do(func() { p.toggle("1") })
	screen.Do(func() {
		c.Nil(p.FindRefKey("1:add specialization"), "an attribute default has no specialization")
		c.Nil(p.FindRefKey("1:add tag"), "nor tags")
		c.NotNil(p.FindRefKey("1:add tl"), "but has a tech level")
	})
}

// TestDefaultsPanelMoreMenu checks that Duplicate, Move up, Move down and Delete change the list as they say, each
// installing a new list rather than changing the one held before, that the open row stays open wherever it goes, and
// that the moves are offered only where there is room.
func TestDefaultsPanelMoreMenu(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	defaults := newTestDefaults()
	p, _ := showDefaultsPanel(t, screen, entity, nil, &defaults)
	act := func(path, label string) {
		screen.Do(func() {
			action := prereqMenuAction(p.moreEntries(path), label)
			c.NotNil(action, "%s offers %s", path, label)
			if action != nil {
				action()
			}
		})
	}
	screen.Do(func() {
		c.Nil(prereqMenuAction(p.moreEntries("0"), "Move Up"), "nothing above the top")
		c.Nil(prereqMenuAction(p.moreEntries("2"), "Move Down"), "nothing below the bottom")
		p.toggle("1")
	})
	skill := defaults[1]
	held := defaults
	act("1", "Move Up")
	c.Equal([]string{gurps.SkillID, gurps.DexterityID, gurps.ParryID}, defaultTypes(defaults))
	c.Equal([]string{gurps.DexterityID, gurps.SkillID, gurps.ParryID}, defaultTypes(held),
		"the list held before is left as it was")
	c.Equal("0", p.open, "the open row moves with its default")
	screen.Do(func() { c.Equal("0"+keyMore, p.Window().Focus().RefKey, "the moved row's more button takes the focus") })
	act("0", "Move Down")
	c.Equal([]string{gurps.DexterityID, gurps.SkillID, gurps.ParryID}, defaultTypes(defaults))
	c.Equal("1", p.open)
	act("1", "Duplicate")
	c.Equal([]string{gurps.DexterityID, gurps.SkillID, gurps.SkillID, gurps.ParryID}, defaultTypes(defaults))
	c.True(defaults[1] == skill && defaults[2] != skill, "the copy follows the original")
	c.Equal(defaultsHash(defaults[1:2]), defaultsHash(defaults[2:3]), "and is the same")
	held = defaults
	act("0", "Delete")
	c.Equal([]string{gurps.SkillID, gurps.SkillID, gurps.ParryID}, defaultTypes(defaults))
	c.Equal([]string{gurps.DexterityID, gurps.SkillID, gurps.SkillID, gurps.ParryID}, defaultTypes(held),
		"the list held before keeps every default, with no nil left at its end")
	c.Equal("0", p.open, "the open row keeps its place as rows above it go")
	act("0", "Delete")
	c.Equal("", p.open, "deleting the open row leaves none open")
	act("1", "Delete")
	act("0", "Delete")
	c.Equal(0, len(defaults))
	screen.Do(func() {
		c.Equal(defaultAddKey, p.Window().Focus().RefKey, "deleting the last gives the add button the focus")
	})
}

// TestDefaultsPanelUndo checks that typing in a field is recorded as one snapshot of the list, that a change to the
// list's shape is another, and that undo and redo install copies of their snapshots in new lists.
func TestDefaultsPanelUndo(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	defaults := newTestDefaults()
	p, host := showDefaultsPanel(t, screen, entity, nil, &defaults)
	screen.Do(func() { p.toggle("1") })
	screen.Do(func() {
		field, ok := p.FindRefKey("1:name").Self.(*StringField)
		c.True(ok, "the open skill default has a name field")
		field.RequestFocus()
		field.SetSelection(0, len(field.Text()))
	})
	screen.Type("Saber")
	c.Equal("Saber", defaults[1].Name.Qualifier)
	c.Equal("Undo Name", host.mgr.UndoTitle())
	screen.Do(func() { prereqMenuAction(p.moreEntries("2"), "Delete")() })
	c.Equal(2, len(defaults))
	c.Equal("Undo Delete Default", host.mgr.UndoTitle())
	installed := defaults[1]
	held := defaults
	screen.Do(host.mgr.Undo)
	c.Equal([]string{gurps.DexterityID, gurps.SkillID, gurps.ParryID}, defaultTypes(defaults))
	c.True(installed != defaults[1], "undo installs a copy")
	c.Equal(2, len(held), "in a new list")
	c.Equal("Saber", defaults[1].Name.Qualifier)
	screen.Do(host.mgr.Undo)
	c.Equal("Broadsword", defaults[1].Name.Qualifier, "the typing was one edit")
	c.False(host.mgr.CanUndo())
	screen.Do(func() {
		field, ok := p.FindRefKey("1:name").Self.(*StringField)
		c.True(ok, "the row the typing was done in is open")
		c.Equal("Broadsword", field.Text())
		c.Equal(field.AsPanel(), field.Window().Focus(), "and the field has the focus")
	})
	screen.Do(host.mgr.Redo)
	c.Equal("Saber", defaults[1].Name.Qualifier)
	screen.Do(host.mgr.Redo)
	c.Equal([]string{gurps.DexterityID, gurps.SkillID}, defaultTypes(defaults))
}

// TestDefaultsPanelDragAndDrop checks that a dragged default goes before or after a row by which half of the row it is
// over, into a new list, that a drop is one step to undo and redo, that the open row follows its default, and that a
// drop onto or beside itself changes nothing.
func TestDefaultsPanelDragAndDrop(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	defaults := newTestDefaults()
	p, host := showDefaultsPanel(t, screen, entity, nil, &defaults)
	drag := func(from, onto string, fraction float32) (accepted bool) {
		screen.Do(func() {
			target := p.FindRefKey(onto + keyMore).Parent()
			r := p.RectFromRoot(target.RectToRoot(target.ContentRect(true)))
			where := geom.NewPoint(r.X+r.Width/3, r.Y+r.Height*fraction)
			data := &rowDrag{panel: p.AsPanel(), path: from}
			p.dragOver(where, data)
			accepted = p.dropTarget != nil
			p.drop(where, data)
		})
		return accepted
	}
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

// TestDefaultsPanelEmpty checks that a list with no defaults shows a placeholder before the add button, that clicking
// it adds a default of the type last chosen and opens it, as the add button does, and that undo brings it back.
func TestDefaultsPanelEmpty(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	var defaults []*gurps.SkillDefault
	c.Equal(0, len(defaults), "precondition: the list has no defaults")
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
	hash := defaultsHash(defaults)
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
	c.Equal(hash, defaultsHash(defaults), "collapsing changes nothing")
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
	c.Equal(0, len(empty), "precondition: the list has no defaults")
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

// TestDefaultsPanelWeapon checks that a weapon's defaults panel reads the weapon's defaults, with their owner's
// nameable values, and edits them.
func TestDefaultsPanelWeapon(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Replacements = map[string]string{"Weapon": "Spear"}
	weapon := gurps.NewWeapon(trait, true)
	weapon.Defaults = []*gurps.SkillDefault{
		{DefaultType: gurps.DexterityID, Modifier: -fxp.Four},
		{DefaultType: gurps.SkillID, Name: criteria.Text{Compare: criteria.IsText, Qualifier: "@Weapon@"}},
	}
	p, _ := showDefaultsPanel(t, screen, gurps.EntityFromNode(weapon), weapon, &weapon.Defaults)
	screen.Do(func() {
		var sentences []string
		for i := range weapon.Defaults {
			if b, ok := p.FindRefKey(strconv.Itoa(i) + keySentence).Self.(*sentenceButton); ok {
				sentences = append(sentences, b.plainText())
			}
		}
		c.Equal("DX at -4|Skill Spear at +0", strings.Join(sentences, "|"))
		prereqMenuAction(p.moreEntries("0"), "Delete")()
	})
	c.Equal([]string{gurps.SkillID}, defaultTypes(weapon.Defaults), "the weapon's defaults are edited")
}

// collapsedSummary returns the paragraph a collapsed panel of sentence rows shows, or "" if it shows none.
func collapsedSummary(p *unison.Panel) string {
	if b, ok := p.FindRefKey(sectionSummaryKey).Self.(*sentenceButton); ok {
		return b.plainText()
	}
	return ""
}

// TestSkillEditorRowsFollowSubstitutions checks that the defaults and features of a skill editor read their nameable
// markers with the values the editor's data holds, and show the new ones once Set Substitutions changes them.
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
		entity.Skills = append(entity.Skills, skill)
		sheet.Rebuild(true)
		e = EditSkill(sheet, skill)
	})
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

	swapForTest(t, &promptForNameables, func(_ promptOperation, sections []nameablesSection) bool {
		for _, section := range sections {
			for k := range section.Nameables {
				section.Nameables[k] = "Rapier"
			}
		}
		return true
	})
	screen.Do(e.nameablesButton.ClickCallback)
	c.Equal("Rapier", e.editorData.Replacements["Weapon"], "precondition: the editor's data takes the new value")
	defaults, features = summaries()
	c.Equal("Skill Rapier at +0.", defaults, "Set Substitutions shows the new value in the defaults")
	c.True(strings.Contains(features, "Rapier"), "and in the features: %s", features)
}
