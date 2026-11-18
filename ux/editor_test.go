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
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
	uncheck "github.com/richardwilkes/unison/enums/check"
)

// newEditorForTrait returns an editor holding the two copies of a trait's edit data that displayEditor would give it,
// without building any of the editor's UI. The trait carries weapons and a script-bearing modifier, so the derived
// values both copies would otherwise publish come from running the scripts in the data.
func newEditorForTrait() *editor[*gurps.Trait, *gurps.TraitEditData] {
	entity := gurps.NewEntity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Claws"
	trait.LocalNotes = "Sharp. <script>1 + 1</script>"
	modifier := gurps.NewTraitModifier(entity, nil, false)
	modifier.Name = "Long"
	modifier.LocalNotes = "Longer. <script>2 + 2</script>"
	trait.Modifiers = []*gurps.TraitModifier{modifier}
	trait.Weapons = []*gurps.Weapon{gurps.NewWeapon(trait, true), gurps.NewWeapon(trait, false)}
	entity.Traits = append(entity.Traits, trait)
	entity.Recalculate()

	e := &editor[*gurps.Trait, *gurps.TraitEditData]{target: trait}
	e.beforeData = &gurps.TraitEditData{}
	e.beforeData.CopyFrom(trait)
	e.editorData = &gurps.TraitEditData{}
	e.editorData.CopyFrom(trait)
	return e
}

// An editor must report unsaved changes only when its data has been edited. The two copies it compares have separate
// script caches, so when their derived values were compared, a script that timed out while marshaling one copy but not
// the other made an untouched editor prompt to save.
func TestEditorIsModifiedFollowsTheDataAlone(t *testing.T) {
	c := check.New(t)
	e := newEditorForTrait()
	c.False(e.isModified(), "an editor whose two copies of the data agree must report no changes")

	e.editorData.Name = "Fangs"
	c.True(e.isModified(), "editing a field must be reported as a change")

	e.editorData.Name = "Claws"
	c.False(e.isModified(), "putting the original value back must clear the change")
}

// buildEditorContent builds an editor's content as displayEditor would, without the toolbar and the docking machinery
// that needs a window, and installs the init function's callback as the real editor does, so a change to one widget
// adjusts the others.
func buildEditorContent[N gurps.Node[N], D gurps.EditorData[N]](owner Rebuildable, target N,
	initContent func(*editor[N, D], *unison.Panel) func(),
) (*editor[N, D], *unison.Panel) {
	e := &editor[N, D]{owner: owner, target: target}
	e.Self = e
	e.undoMgr = unison.NewUndoManager(100, func(_ error) {})
	e.SetLayout(&unison.FlexLayout{Columns: 1})
	reflect.ValueOf(&e.beforeData).Elem().Set(reflect.New(reflect.TypeOf(e.beforeData).Elem()))
	e.beforeData.CopyFrom(target)
	reflect.ValueOf(&e.editorData).Elem().Set(reflect.New(reflect.TypeOf(e.editorData).Elem()))
	e.editorData.CopyFrom(target)
	content := e.newContentPanel(2)
	e.AddChild(e.scroll)
	e.fillContent(content, initContent)
	return e, content
}

// checkBoxesTitled returns every checkbox bearing the given title found anywhere beneath the given panel.
func checkBoxesTitled(p *unison.Panel, title string) []*CheckBox {
	var boxes []*CheckBox
	for _, box := range uxtest.PanelsOfType[*CheckBox](p) {
		if box.Text.String() == title {
			boxes = append(boxes, box)
		}
	}
	return boxes
}

// findCheckBoxTitled returns the first checkbox bearing the given title found anywhere beneath the given panel.
func findCheckBoxTitled(p *unison.Panel, title string) *CheckBox {
	if boxes := checkBoxesTitled(p, title); len(boxes) != 0 {
		return boxes[0]
	}
	return nil
}

// clickCheckBox puts the checkbox into the given state and runs its click callback, as a user's click does.
func clickCheckBox(box *CheckBox, on bool) {
	box.State = uncheck.FromBool(on)
	box.ClickCallback()
}

// scrollRecordingPanel stands in for a scroll area above a table, recording the rects it is asked to bring into view.
type scrollRecordingPanel struct {
	unison.Panel
	rects []geom.Rect
}

func newScrollRecordingPanel() *scrollRecordingPanel {
	p := &scrollRecordingPanel{}
	p.Self = p
	p.ScrollRectIntoViewCallback = func(rect geom.Rect) bool {
		p.rects = append(p.rects, rect)
		return true
	}
	return p
}

func TestRevealRowForDataScrollsOnlyTheEditedRow(t *testing.T) {
	c := check.New(t)
	scroller := newScrollRecordingPanel()
	provider := &fakeAltDropProvider{}
	table := unison.NewTable(provider)
	scroller.AddChild(table)
	traits := []*gurps.Trait{
		gurps.NewTrait(nil, nil, false),
		gurps.NewTrait(nil, nil, false),
		gurps.NewTrait(nil, nil, false),
	}
	rows := make([]*Node[*gurps.Trait], len(traits))
	for i, trait := range traits {
		rows[i] = NewNode(table, nil, trait, false)
	}
	table.SetRootRows(rows)
	c.True(table.FrameRect().Height > table.RowFrame(2).Height, "the table must be taller than a single row")

	revealRowForData(table, traits[2])
	c.Equal(1, len(scroller.rects), "revealing a displayed row must ask for exactly one scroll")
	c.Equal(table.RowFrame(2), scroller.rects[0], "only the edited row, not the whole table, may be scrolled into view")

	scroller.rects = nil
	revealRowForData(table, gurps.NewTrait(nil, nil, false))
	c.Equal(0, len(scroller.rects), "an item the table does not display must not cause any scrolling")
}

// TestReadOnlyEditor checks that an editor opened from a read-only one or from a locked list is read-only too, and that
// it leaves on only what reads or opens its rows, through rebuilds, offers no substitutions, no change to its source
// and no new items, has a modifier table whose rows open but don't change, and never reports a change to apply.
func TestReadOnlyEditor(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	screen.Do(func() {
		newAlly := func(owner gurps.DataOwner) *gurps.Trait {
			trait := gurps.NewTrait(owner, nil, false)
			trait.Name = "Ally (@Who@)"
			trait.CanLevel = true
			trait.Levels = fxp.One
			trait.Source = testSource
			trait.Modifiers = []*gurps.TraitModifier{gurps.NewTraitModifier(owner, nil, false)}
			return trait
		}
		sheet := newTestSheetForTemplate(t)
		entity := sheet.Entity()
		host := gurps.NewTrait(entity, nil, false)
		trait := newAlly(entity)
		entity.Traits = []*gurps.Trait{host, trait}
		sheet.Rebuild(true)
		DisplayNewDockable(sheet)
		hostEditor := EditTrait(sheet, host)
		hostEditor.makeReadOnly()
		listTrait := newAlly(nil)
		list := NewTraitTableDockable("test"+gurps.TraitsExt, []*gurps.Trait{listTrait})
		DisplayNewDockable(list)
		list.setLocked(true)
		// Each editor is opened only once the one before it has been checked, since an editor that shares its slice of
		// the dock with a later one is hidden behind it, which disables it.
		for _, one := range []struct {
			owner  Rebuildable
			target *gurps.Trait
		}{{sheet, trait}, {hostEditor, trait}, {list, listTrait}} {
			e := EditTrait(one.owner, one.target)
			c.Equal(e.owner != sheet, e.readOnly)
			live := func() int {
				return len(uxtest.PanelsMatching(e.content, func(p *unison.Panel) bool {
					_, toggle := p.Self.(*sectionToggle)
					_, inList := p.Parent().Self.(interface{ reattach() })
					return p.Enabled() && p.Focusable() && !toggle && !inList
				}))
			}
			c.Equal(e.readOnly, live() == 0, "only what reads or opens rows is left on")
			e.Rebuild(false)
			c.Equal(e.readOnly, live() == 0, "after a rebuild too")
			entries := e.sourceMenuEntries()
			c.Equal(e.readOnly, entries[len(entries)-1].Disabled, "Clear Source")
			c.Equal(!e.readOnly, e.nameablesButton.Enabled(), "Set Substitutions")
			c.Equal(!e.readOnly, e.CanPerformCmd(nil, NewTraitModifierItemID), "New Trait Modifier")
			table, ok := uxtest.FirstPanelOfType[*unison.Table[*Node[*gurps.TraitModifier]]](e.content)
			c.True(ok)
			table.SelectByIndex(0)
			c.True(table.CanPerformCmd(nil, OpenEditorItemID), "a modifier opens")
			c.Equal(!e.readOnly, table.CanPerformCmd(nil, ToggleStateItemID), "Toggle State")
			c.Equal(!e.readOnly, table.CanPerformCmd(nil, unison.DeleteItemID), "Delete")
			e.editorData.Name = "Enemy"
			c.Equal(!e.readOnly, e.Modified(), "a change to apply")
			e.editorData.Name = one.target.Name
		}
	})
}

// An editor open on a granted row follows the row through the recalculations that reconcile the sheet's grants: one
// that remakes the row from a changed grant, as a new object under the same ID, has the editor show the new object,
// closing the editors opened from it on the old object's modifiers, one that takes the row off the sheet closes the
// editor, whether the sheet is rebuilt or only marked modified, and one that gives the row the details of another grant
// in place has the editor show them, leaving the editors opened from it open.
func TestGrantedRowEditorFollowsItsRow(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	var sheet *Sheet
	var training *gurps.Trait
	var g *gurps.Grant
	granted := func() *gurps.Trait {
		for _, one := range sheet.Entity().Traits {
			if gurps.IsGranted(one) {
				return one
			}
		}
		return nil
	}
	isOpen := func(d unison.Dockable) bool {
		return slices.ContainsFunc(AllDockables(), func(open unison.Dockable) bool { return open.AsPanel().Self == d })
	}
	var e *editor[*gurps.Trait, *gurps.TraitEditData]
	var modifierEditor unison.Dockable
	screen.Do(func() {
		sheet = newTestSheetForTemplate(t)
		entity := sheet.Entity()
		training = gurps.NewTrait(entity, nil, false)
		training.Name = "Training"
		item := newGrantTestTrait("Night Vision", fxp.Two)
		item.Modifiers = []*gurps.TraitModifier{gurps.NewTraitModifier(nil, nil, false)}
		g = gurps.NewGrant(feature.GrantTrait)
		g.SetItem(item, gurps.LibraryFile{})
		g.Amount = fxp.Two
		training.Features = gurps.Features{g}
		entity.Traits = []*gurps.Trait{training}
		sheet.Rebuild(true)
		DisplayNewDockable(sheet)
		row := granted()
		if row == nil {
			t.Fatal("precondition: the grant put its trait on the sheet")
		}
		e = EditTrait(sheet, row)
		c.True(e.readOnly, "precondition: a granted row's editor only shows it")
		c.Equal(fxp.Two, e.editorData.Levels)
		EditTraitModifier(e, row.Modifiers[0])
		for _, d := range AllDockables() {
			if _, ok := d.AsPanel().Self.(*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData]); ok {
				modifierEditor = d
			}
		}
		if modifierEditor == nil {
			t.Fatal("precondition: the modifier's editor opened")
		}
	})

	// Change the grant's level, as applying the carrier's editor does, and rebuild the sheet as that does too.
	screen.Do(func() {
		g.Amount = fxp.Five
		sheet.Rebuild(true)
		row := granted()
		c.NotNil(row, "the grant still puts its trait on the sheet")
		c.True(isOpen(e), "the editor stays open")
		c.True(e.target == row, "on the row's new object")
		c.Equal(fxp.Five, e.editorData.Levels, "showing it as remade")
		c.False(isOpen(modifierEditor), "the editor opened from it on the old object's modifier closes")
		c.True(EditTrait(sheet, row) == e, "and opening the row again brings the editor forward instead of another")
	})

	// Disable the carrier by its checkmark, which only marks the sheet modified, taking the row off the sheet.
	screen.Do(func() {
		training.Disabled = true
		sheet.MarkModified(nil)
		c.Nil(granted(), "precondition: the disabled carrier's grant takes its trait off the sheet")
		c.False(isOpen(e), "the editor closes with the row")
	})

	// Open the editor again, then take the grant away altogether.
	screen.Do(func() {
		training.Disabled = false
		sheet.Rebuild(true)
		row := granted()
		if row == nil {
			t.Fatal("precondition: the grant put its trait back on the sheet")
		}
		e = EditTrait(sheet, row)
		training.Features = nil
		sheet.Rebuild(true)
		c.Nil(granted(), "precondition: the trait leaves the sheet with its grant")
		c.False(isOpen(e), "the editor closes with the row")
	})

	// Carry worn goggles granting Night Vision 1, which stand in for the unequipped spare goggles granting Night Vision
	// 3, and open the editor on the row.
	var spare *gurps.Equipment
	var row *gurps.Trait
	screen.Do(func() {
		entity := sheet.Entity()
		goggles := func(name string, level fxp.Int) *gurps.Equipment {
			eqp := gurps.NewEquipment(entity, nil, false)
			eqp.Name = name
			item := newGrantTestTrait("Night Vision", level)
			item.Modifiers = []*gurps.TraitModifier{gurps.NewTraitModifier(nil, nil, false)}
			grant := gurps.NewGrant(feature.GrantTrait)
			grant.SetItem(item, gurps.LibraryFile{})
			eqp.Features = gurps.Features{grant}
			return eqp
		}
		spare = goggles("Spare Goggles", fxp.Three)
		spare.Equipped = false
		entity.CarriedEquipment = []*gurps.Equipment{spare, goggles("Worn Goggles", fxp.One)}
		sheet.Rebuild(true)
		if row = granted(); row == nil {
			t.Fatal("precondition: the worn goggles put Night Vision on the sheet")
		}
		e = EditTrait(sheet, row)
		c.Equal(fxp.One, e.editorData.Levels, "precondition: at the worn goggles' level")
		EditTraitModifier(e, row.Modifiers[0])
		modifierEditor = nil
		for _, d := range AllDockables() {
			if _, ok := d.AsPanel().Self.(*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData]); ok {
				modifierEditor = d
			}
		}
		if modifierEditor == nil {
			t.Fatal("precondition: the modifier's editor opened")
		}
	})

	// Equip the spare goggles, which gives the row their level in place.
	screen.Do(func() {
		spare.Equipped = true
		sheet.Rebuild(true)
		c.True(granted() == row, "precondition: the row stays the object it was")
		c.Equal(fxp.Three, row.Levels, "precondition: at the spare goggles' level")
		c.True(isOpen(e), "the editor stays open")
		c.Equal(fxp.Three, e.editorData.Levels, "showing the level the row took")
		c.True(isOpen(modifierEditor), "as does the editor opened from it, whose object is the same")
	})
}

// Tags are kept sorted whatever order they are typed in, so that a file never records the order the user entered them
// in (#1217). A spell's colleges are kept the same way.
func TestEditorApplySortsTagsAndColleges(t *testing.T) {
	c := check.New(t)
	data := gurps.NewTemplate()
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Trait"
	data.Traits = []*gurps.Trait{trait}
	spell := gurps.NewSpell(nil, nil, false)
	spell.Name = "Spell"
	data.Spells = []*gurps.Spell{spell}
	template := newTestTemplateDockable("Source", data)

	traitEditor, content := buildEditorContent(template, trait, initTraitEditor)
	tagsField := stringFieldLabeled(content, i18n.Text("Tags"))
	c.NotNil(tagsField, "the trait editor must have a Tags field")
	tagsField.SetText("BBB, AAA")
	c.Equal([]string{"BBB", "AAA"}, traitEditor.editorData.Tags, "the editor holds the tags as typed")
	traitEditor.applyEdits()
	c.Equal([]string{"AAA", "BBB"}, trait.Tags, "applying the edit sorts the trait's tags")

	spellEditor, content := buildEditorContent(template, spell, initSpellEditor)
	tagsField = stringFieldLabeled(content, i18n.Text("Tags"))
	c.NotNil(tagsField, "the spell editor must have a Tags field")
	tagsField.SetText("BBB, AAA")
	collegeField := stringFieldLabeled(content, i18n.Text("College"))
	c.NotNil(collegeField, "the spell editor must have a College field")
	collegeField.SetText("Water, Air")
	c.Equal(gurps.CollegeList{"Water", "Air"}, spellEditor.editorData.College, "the editor holds the colleges as typed")
	spellEditor.applyEdits()
	c.Equal([]string{"AAA", "BBB"}, spell.Tags, "applying the edit sorts the spell's tags")
	c.Equal(gurps.CollegeList{"Air", "Water"}, spell.College, "applying the edit sorts the spell's colleges")
}

// TestModifierEditorFollowsTheChoiceRules verifies that enabling an option in its own editor makes it the pick,
// turning the one that was picked off, and that undo puts both back. On a sheet, the pick of a mandatory choice can't
// be turned off there either, though it can be elsewhere.
func TestModifierEditorFollowsTheChoiceRules(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	traitEditor, _, _ := traitEditorOnSheet(t, sheet, newTraitModifierChoiceFor(sheet.Entity(), true, []string{"A", "B"}, "A"))
	options := traitEditor.editorData.Modifiers[0].Children

	e, _ := buildEditorContent(traitEditor, options[1], initTraitModifierEditor)
	e.editorData.Disabled = false
	e.applyEdits()
	c.False(options[0].Enabled(), "enabling B in its editor turns A off")
	c.True(options[1].Enabled())
	mgr := unison.UndoManagerFor(traitEditor)
	mgr.Undo()
	c.True(options[0].Enabled(), "undo turns A back on")
	c.False(options[1].Enabled(), "and B back off")
	mgr.Redo()
	c.False(options[0].Enabled(), "redo turns A off again")
	c.True(options[1].Enabled(), "and B back on")
	mgr.Undo()

	e, _ = buildEditorContent(traitEditor, options[0], initTraitModifierEditor)
	e.editorData.Disabled = true
	e.applyEdits()
	c.True(options[0].Enabled(), "the pick of a mandatory choice on a sheet can't be turned off in its editor")

	offSheet := newTraitModifierChoiceFor(nil, true, []string{"A"}, "A")
	e, _ = buildEditorContent(nil, offSheet.Children[0], initTraitModifierEditor)
	e.editorData.Disabled = true
	e.applyEdits()
	c.False(offSheet.Children[0].Enabled(), "off a sheet a mandatory choice may be left without its pick")
}

// TestEditorKeepsAPickMadeSinceItOpened verifies that applying an option's editor, opened while it was the pick, after
// another option has been picked in the list leaves that later pick alone, since the editor didn't change whether its
// option is enabled.
func TestEditorKeepsAPickMadeSinceItOpened(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	traitEditor, _, table := traitEditorOnSheet(t, sheet, newTraitModifierChoiceFor(sheet.Entity(), true, []string{"A", "B"}, "A"))
	options := traitEditor.editorData.Modifiers[0].Children
	aEditor, _ := buildEditorContent(traitEditor, options[0], initTraitModifierEditor)
	adjustModifierEnabled(traitEditor, table, options[1], true)
	c.True(options[1].Enabled())
	aEditor.editorData.Name = "Renamed"
	aEditor.applyEdits()
	c.Equal("Renamed", options[0].Name)
	c.False(options[0].Enabled(), "the editor didn't change whether A is enabled, so A stays off")
	c.True(options[1].Enabled(), "and B stays the pick")
}

// TestChoiceMadeMandatoryInItsEditorOnASheet verifies that applying a choice's editor on a sheet without changing what
// the choice asks for leaves its options as they are, so a mandatory choice flagged for want of a pick isn't given one,
// while a choice made mandatory there gets its first option picked, as Convert to Choice does, undo and redo covering
// it.
func TestChoiceMadeMandatoryInItsEditorOnASheet(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	traitEditor, _, _ := traitEditorOnSheet(t, sheet, newTraitModifierChoiceFor(sheet.Entity(), true, []string{"A", "B"}))
	choice := traitEditor.editorData.Modifiers[0]
	choiceEditor, _ := buildEditorContent(traitEditor, choice, initTraitModifierEditor)
	choiceEditor.editorData.Name = "Renamed"
	choiceEditor.applyEdits()
	c.Equal("Renamed", choice.Name)
	c.False(choice.Children[0].Enabled(), "the edit doesn't pick the first option")
	c.False(choice.Children[1].Enabled())

	choice.SetMandatoryChoice(false)
	choiceEditor, _ = buildEditorContent(traitEditor, choice, initTraitModifierEditor)
	choiceEditor.editorData.SetMandatoryChoice(true)
	choiceEditor.applyEdits()
	c.True(gurps.IsMandatoryModifierChoice(choice))
	c.True(choice.Children[0].Enabled(), "the first option is picked")
	mgr := unison.UndoManagerFor(traitEditor)
	mgr.Undo()
	c.False(gurps.IsMandatoryModifierChoice(choice))
	c.False(choice.Children[0].Enabled(), "undo takes the pick back")
	mgr.Redo()
	c.True(choice.Children[0].Enabled(), "redo picks it again")
}

// pageRefButtonAndField returns the button that opens the page reference the content edits and the field that holds the
// reference, or nil when no field is labeled by it, both located through the Page Reference label, which the button
// follows and the field is labeled by.
func pageRefButtonAndField(c check.Checker, content *unison.Panel) (*unison.Button, *StringField) {
	var label *unison.Label
	for _, one := range uxtest.PanelsOfType[*unison.Label](content) {
		if one.String() == "Page Reference" {
			label = one
			break
		}
	}
	c.NotNil(label, "the content must have a page reference label")
	siblings := label.Parent().Children()
	index := slices.Index(siblings, label.AsPanel())
	c.True(index >= 0 && index+1 < len(siblings), "the page reference label must have a sibling after it")
	button, ok := siblings[index+1].Self.(*unison.Button)
	c.True(ok, "the page reference label must be followed by its button")
	for _, field := range uxtest.PanelsOfType[*StringField](content) {
		if field.Accessibility.LabeledBy == label {
			return button, field
		}
	}
	return button, nil
}

// TestEditorPageRefButtonFollowsTheField verifies that every editor puts a button beside its page reference that is
// enabled only while the field holds a reference to open, including after an undo and in a read-only editor.
func TestEditorPageRefButtonFollowsTheField(t *testing.T) {
	c := check.New(t)
	for name, content := range everyEditorContent(t) {
		button, field := pageRefButtonAndField(c, content)
		c.NotNil(field, name+" editor must label its page reference field")
		c.False(button.Enabled(), name+" editor has no reference, so there is nothing to open")
	}

	sheet := newTestSheetForTemplate(t)
	trait := gurps.NewTrait(sheet.Entity(), nil, false)
	trait.PageRef = "B42"
	e, content := buildEditorContent(sheet, trait, initTraitEditor)
	button, field := pageRefButtonAndField(c, content)
	c.NotNil(field, "the trait editor must label its page reference field")
	c.True(button.Enabled(), "a reference loaded with the editor enables the button")

	field.SetText("")
	c.Equal("", e.editorData.PageRef)
	c.False(button.Enabled(), "clearing the reference disables the button")
	e.UndoManager().Undo()
	c.Equal("B42", e.editorData.PageRef)
	c.True(button.Enabled(), "undoing the clear enables the button again")
	field.SetText(" ; ")
	c.False(button.Enabled(), "separators alone are no reference")

	field.SetText("B42")
	e.makeReadOnly()
	c.False(field.Enabled(), "a read-only editor's reference can't be changed")
	c.True(button.Enabled(), "a read-only editor's reference can still be opened, since that changes nothing")
}

// TestEditorPageRefHighlightFallsBackToTheName verifies that the page reference button looks for the name on the page
// while the highlight field is blank, as the page reference cell of a table does, and follows the name as edited.
func TestEditorPageRefHighlightFallsBackToTheName(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	trait := gurps.NewTrait(sheet.Entity(), nil, false)
	trait.Name = "Acute Senses"
	trait.PageRef = "B35"
	e, _ := buildEditorContent(sheet, trait, initTraitEditor)
	c.Equal("Acute Senses", e.pageRefHighlight(), "a blank highlight looks for the name")
	e.editorData.Name = "Acute Vision"
	c.Equal("Acute Vision", e.pageRefHighlight(), "the name as edited, not as saved")
	e.editorData.PageRefHighlight = "Vision"
	c.Equal("Vision", e.pageRefHighlight(), "a highlight is looked for as given")
}

// TestSkillEditorRowsFollowSubstitutions checks that the prerequisites, defaults and features of a skill editor read
// their nameable markers with the values the editor's data holds, and show the new ones once Set Substitutions changes
// them, but not when it is canceled.
func TestSkillEditorRowsFollowSubstitutions(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
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
			for _, p := range uxtest.PanelsOfType[*prereqPanel](e.content) {
				text = collapsedSummary(p.AsPanel())
			}
		})
		return text
	}
	summaries := func() (defaults, features string) {
		screen.Do(func() {
			for _, p := range uxtest.PanelsOfType[*defaultsPanel](e.content) {
				defaults = collapsedSummary(p.AsPanel())
			}
			for _, p := range uxtest.PanelsOfType[*featuresPanel](e.content) {
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
		uxtest.SwapForTest(t, &promptForNameables, func(_ promptOperation, sections []nameablesSection) bool {
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

// TestWeaponEditorTakesItemSubstitutions checks that the defaults of a weapon edited from its item's editor, or from
// the editor of the item editor's copy of the modifier holding it, show the substitutions the item's editor holds,
// including those Set Substitutions makes while the weapon's editor is open.
func TestWeaponEditorTakesItemSubstitutions(t *testing.T) {
	for _, tc := range []struct {
		name string
		// edit puts an item on the sheet whose substitutions give the weapon's @Blade@ the value Spear, with the weapon
		// as one of the item's own or one its modifier holds, opens the weapon's editor from the item's editor or the
		// modifier's, and returns the item editor's Set Substitutions button.
		edit func(sheet *Sheet, weapon *gurps.Weapon) *unison.Button
	}{
		{
			name: "skill",
			edit: func(sheet *Sheet, weapon *gurps.Weapon) *unison.Button {
				entity := sheet.Entity()
				skill := gurps.NewSkill(entity, nil, false)
				skill.Replacements = map[string]string{"Blade": "Spear"}
				weapon.SetOwner(skill)
				skill.Weapons = []*gurps.Weapon{weapon}
				entity.Skills = append(entity.Skills, skill)
				sheet.Rebuild(true)
				e := EditSkill(sheet, skill)
				EditWeapon(e, e.editorData.Weapons[0])
				return e.nameablesButton
			},
		},
		{
			// The modifier's code is linked to one of the trait's, which answers it.
			name: "trait modifier",
			edit: func(sheet *Sheet, weapon *gurps.Weapon) *unison.Button {
				entity := sheet.Entity()
				modifier := gurps.NewTraitModifier(entity, nil, false)
				modifier.Weapons = []*gurps.Weapon{weapon}
				trait := gurps.NewTrait(entity, nil, false)
				trait.SetModifiers([]*gurps.TraitModifier{modifier})
				trait.NameableLinks = map[tid.TID]map[string]string{modifier.TID: {"Blade": "Weapon"}}
				trait.Replacements = map[string]string{"Weapon": "Spear"}
				entity.Traits = append(entity.Traits, trait)
				sheet.Rebuild(true)
				e := EditTrait(sheet, trait)
				modEditor := EditTraitModifier(e, e.editorData.Modifiers[0])
				EditWeapon(modEditor, modEditor.editorData.Weapons[0])
				return e.nameablesButton
			},
		},
		{
			name: "equipment modifier",
			edit: func(sheet *Sheet, weapon *gurps.Weapon) *unison.Button {
				entity := sheet.Entity()
				modifier := gurps.NewEquipmentModifier(entity, nil, false)
				modifier.Weapons = []*gurps.Weapon{weapon}
				eqp := gurps.NewEquipment(entity, nil, false)
				eqp.SetModifiers([]*gurps.EquipmentModifier{modifier})
				eqp.Replacements = map[string]string{"Blade": "Spear"}
				entity.CarriedEquipment = append(entity.CarriedEquipment, eqp)
				sheet.Rebuild(true)
				e := EditEquipment(sheet, eqp, true)
				modEditor := EditEquipmentModifier(e, e.editorData.Modifiers[0])
				EditWeapon(modEditor, modEditor.editorData.Weapons[0])
				return e.nameablesButton
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			screen, _ := uxtest.StartHeadlessWorkspace(t, c)
			sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
			if !ok {
				t.Fatal("New Character Sheet must open a character sheet")
			}
			var button *unison.Button
			screen.Do(func() {
				weapon := gurps.NewWeapon(nil, true)
				weapon.Defaults = []*gurps.SkillDefault{
					{DefaultType: gurps.SkillID, Name: criteria.Text{Compare: criteria.IsText, Qualifier: "@Blade@"}},
				}
				button = tc.edit(sheet, weapon)
			})
			weaponEditor := soleEditorOfType[*editor[*gurps.Weapon, *gurps.Weapon]](t, screen)
			var p *defaultsPanel
			screen.Do(func() {
				if panels := uxtest.PanelsOfType[*defaultsPanel](weaponEditor.content); len(panels) == 1 {
					p = panels[0]
				}
			})
			if p == nil {
				t.Fatal("the weapon editor must have a defaults panel")
			}
			screen.Do(func() { c.Equal("Skill Spear at +0.", collapsedSummary(p.AsPanel())) })
			uxtest.SwapForTest(t, &promptForNameables, func(_ promptOperation, sections []nameablesSection) bool {
				for _, section := range sections {
					for k := range section.Nameables {
						section.Nameables[k] = "Rapier"
					}
				}
				return true
			})
			screen.Do(button.ClickCallback)
			screen.Do(func() {
				c.Equal("Skill Rapier at +0.", collapsedSummary(p.AsPanel()), "the open weapon editor shows the new value")
			})
		})
	}
}

// TestModifierEditorFollowsItemEditor checks that the editor of a modifier edited from its item's editor follows what
// the item's editor holds: the modifier's features show the substitutions it holds, both those Set Substitutions made
// before the modifier's editor opened and those it makes while it is open, and follow its link for the modifier's code,
// and the modifier's nameables what defines its codes, as the item's nameables or its other modifiers change, or that
// is undone, and as a grant is added to the item's features, while the modifier's editor is open.
func TestModifierEditorFollowsItemEditor(t *testing.T) {
	for _, tc := range []struct {
		name string
		// edit puts an item on the sheet whose substitutions give the modifier's @Blade@ the value Spear, with the
		// modifier holding the features, the item using and defining @Arm@ and both using @Grip@, which neither
		// defines, and a second modifier, Pommel, which defines @Grip@ but is turned off, opens the item's editor, and
		// returns its Set Substitutions button, its content and a function that opens the editor of the item editor's
		// copy of the modifier, returning its content.
		edit func(sheet *Sheet, features gurps.Features) (*unison.Button, *unison.Panel, func() *unison.Panel)
		// toggle turns Pommel on or off from the modifiers list of the item's editor, given its content.
		toggle func(content *unison.Panel)
	}{
		{
			// The modifier's code is linked to one of the trait's, which answers it.
			name: "trait modifier",
			edit: func(sheet *Sheet, features gurps.Features) (*unison.Button, *unison.Panel, func() *unison.Panel) {
				entity := sheet.Entity()
				modifier := gurps.NewTraitModifier(entity, nil, false)
				modifier.Name = "@Grip@ Hilt"
				modifier.Features = features
				pommel := gurps.NewTraitModifier(entity, nil, false)
				pommel.Name = "Pommel"
				pommel.Nameables = map[string]nameable.Detail{"Grip": {FreeForm: true}}
				pommel.SetDisabled(true)
				trait := gurps.NewTrait(entity, nil, false)
				trait.Name = "@Arm@ @Grip@ Training"
				trait.Nameables = map[string]nameable.Detail{"Arm": {FreeForm: true}, "Weapon": {FreeForm: true}}
				trait.SetModifiers([]*gurps.TraitModifier{modifier, pommel})
				trait.NameableLinks = map[tid.TID]map[string]string{modifier.TID: {"Blade": "Weapon"}}
				trait.Replacements = map[string]string{"Weapon": "Spear"}
				entity.Traits = append(entity.Traits, trait)
				sheet.Rebuild(true)
				e := EditTrait(sheet, trait)
				return e.nameablesButton, e.content, func() *unison.Panel {
					return EditTraitModifier(e, e.editorData.Modifiers[0]).content
				}
			},
			toggle: func(content *unison.Panel) {
				if p, ok := uxtest.FirstPanelOfType[*traitModifiersPanel](content); ok {
					p.table.SetSelectionMap(map[tid.TID]bool{(*p.list)[1].ID(): true})
					p.table.PerformCmd(nil, ToggleStateItemID)
				}
			},
		},
		{
			name: "equipment modifier",
			edit: func(sheet *Sheet, features gurps.Features) (*unison.Button, *unison.Panel, func() *unison.Panel) {
				entity := sheet.Entity()
				modifier := gurps.NewEquipmentModifier(entity, nil, false)
				modifier.Name = "@Grip@ Hilt"
				modifier.Features = features
				pommel := gurps.NewEquipmentModifier(entity, nil, false)
				pommel.Name = "Pommel"
				pommel.Nameables = map[string]nameable.Detail{"Grip": {FreeForm: true}}
				pommel.SetDisabled(true)
				eqp := gurps.NewEquipment(entity, nil, false)
				eqp.Name = "@Arm@ @Grip@ Kit"
				eqp.Nameables = map[string]nameable.Detail{"Arm": {FreeForm: true}}
				eqp.SetModifiers([]*gurps.EquipmentModifier{modifier, pommel})
				eqp.Replacements = map[string]string{"Blade": "Spear"}
				entity.CarriedEquipment = append(entity.CarriedEquipment, eqp)
				sheet.Rebuild(true)
				e := EditEquipment(sheet, eqp, true)
				return e.nameablesButton, e.content, func() *unison.Panel {
					return EditEquipmentModifier(e, e.editorData.Modifiers[0]).content
				}
			},
			toggle: func(content *unison.Panel) {
				if p, ok := uxtest.FirstPanelOfType[*equipmentModifiersPanel](content); ok {
					p.table.SetSelectionMap(map[tid.TID]bool{(*p.list)[1].ID(): true})
					p.table.PerformCmd(nil, ToggleStateItemID)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			screen, _ := uxtest.StartHeadlessWorkspace(t, c)
			sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
			if !ok {
				t.Fatal("New Character Sheet must open a character sheet")
			}
			var button *unison.Button
			var itemContent *unison.Panel
			var editModifier func() *unison.Panel
			screen.Do(func() {
				bonus := gurps.NewSkillBonus()
				bonus.NameCriteria.Qualifier = "@Blade@"
				button, itemContent, editModifier = tc.edit(sheet, gurps.Features{bonus})
			})
			// Each code is given its own answer, so that the one a link leads to can be told from the others.
			answer := func(value string) {
				uxtest.SwapForTest(t, &promptForNameables, func(_ promptOperation, sections []nameablesSection) bool {
					for _, section := range sections {
						for k := range section.Nameables {
							section.Nameables[k] = value + " " + k
						}
					}
					return true
				})
				screen.Do(button.ClickCallback)
			}
			answer("Rapier")
			var p *featuresPanel
			var modifierNameables *nameablesPanel
			screen.Do(func() {
				content := editModifier()
				if panels := uxtest.PanelsOfType[*featuresPanel](content); len(panels) == 1 {
					p = panels[0]
				}
				modifierNameables, _ = uxtest.FirstPanelOfType[*nameablesPanel](content)
			})
			if p == nil || modifierNameables == nil {
				t.Fatal("the modifier editor must have a features panel and a nameables panel")
			}
			summary := func() string {
				var text string
				screen.Do(func() { text = collapsedSummary(p.AsPanel()) })
				return text
			}
			text := summary()
			c.True(strings.Contains(text, "Rapier"), "the features show the value set before the editor opened: %s", text)
			answer("Sabre")
			before := summary()
			c.True(strings.Contains(before, "Sabre"), "and the one set while it is open: %s", before)

			var nameables *nameablesPanel
			var path string
			screen.Do(func() {
				nameables, ok = uxtest.FirstPanelOfType[*nameablesPanel](itemContent)
				if !ok {
					return
				}
				if mods := nameables.modifiers(); len(mods) == 2 {
					path = "m:" + string(mods[0].id) + ":Blade"
				}
				if nameables.collapse.collapsed {
					nameables.collapse.toggle()
				}
			})
			if nameables == nil || path == "" {
				t.Fatal("the item editor must have a nameables panel with a row for the modifier's code")
			}
			screen.Do(func() { nameables.toggle(path) })
			screen.Do(func() {
				popup, isPopup := nameables.FindRefKey(path + ":link").Self.(*unison.PopupMenu[string])
				c.True(isPopup, "the modifier's code can be linked")
				if isPopup {
					popup.Select("Arm")
				}
			})
			text = summary()
			c.True(strings.Contains(text, "Sabre Arm"), "the features follow the new link: %s", text)
			screen.Do(unison.UndoManagerFor(nameables).Undo)
			c.Equal(before, summary(), "and the undo of the new link")

			grip := func() string {
				var sentence string
				screen.Do(func() { sentence, _ = nameableRowSentence(modifierNameables, "u:Grip") })
				return sentence
			}
			undefined := grip()
			c.Equal("@Grip@ isn't defined; used once", undefined)
			screen.Do(func() { clickNameablesButton(c, nameables, "u:Grip:define") })
			c.Equal("@Grip@ is defined by the item the modifier belongs to; used once", grip(),
				"the modifier's nameables follow the item's new definition")
			screen.Do(unison.UndoManagerFor(nameables).Undo)
			c.Equal(undefined, grip(), "and the undo of the definition")

			screen.Do(func() { tc.toggle(itemContent) })
			c.Equal("@Grip@ is defined by the modifier Pommel; used once", grip(),
				"the modifier's nameables follow another modifier being turned on")
			screen.Do(unison.UndoManagerFor(nameables).Undo)
			c.Equal(undefined, grip(), "and the undo of turning it on")

			screen.Do(func() {
				features, found := uxtest.FirstPanelOfType[*featuresPanel](itemContent)
				c.True(found, "the item editor has a features panel")
				if !found {
					return
				}
				granted := gurps.NewTrait(nil, nil, false)
				granted.Nameables = map[string]nameable.Detail{"Grip": {FreeForm: true}}
				grant := gurps.NewGrant(feature.GrantTrait)
				grant.SetItem(granted, gurps.LibraryFile{})
				features.edit("Add Feature", "", "", func() { *features.features = append(*features.features, grant) })
			})
			c.Equal("@Grip@ is defined by an item granted by the item the modifier belongs to; used once", grip(),
				"the modifier's nameables follow a grant added to the item")
		})
	}
}

// TestEditorUndoRestoresSourceClearedByTemplatePicker verifies that applying an edit to a choice container that still
// holds a source clears the source, and that undo and redo put it back and take it away again along with everything
// else. A choice is stripped of its source whenever it is created, converted, loaded or transferred, so this is only a
// backstop, but the editor must not be the way one keeps it. An edit to a plain group must leave its source alone when
// applied, undone and redone.
func TestEditorUndoRestoresSourceClearedByTemplatePicker(t *testing.T) {
	c := check.New(t)
	choice := gurps.NewTraitChoiceContainer(nil, nil)
	choice.Name = "Pick One"
	choiceSource := gurps.Source{Library: "lib", Path: "choice.adq", TID: choice.ID()}
	choice.Source = choiceSource
	group := gurps.NewTrait(nil, nil, true)
	group.Name = "Advantages"
	groupSource := gurps.Source{Library: "lib", Path: "group.adq", TID: group.ID()}
	group.Source = groupSource
	data := gurps.NewTemplate()
	data.Traits = []*gurps.Trait{choice, group}
	template := newTestTemplateDockable("Source", data)
	mgr := unison.UndoManagerFor(template)
	c.NotNil(mgr, "the template must have an undo manager")

	e, _ := buildEditorContent(template, choice, initTraitEditor)
	e.editorData.Name = "Pick Another"
	e.applyEdits()
	c.Equal("Pick Another", choice.Name)
	c.True(choice.Source.IsZero(), "a choice container must not keep a source through an edit")

	mgr.Undo()
	c.Equal("Pick One", choice.Name)
	c.Equal(choiceSource, choice.Source, "undo must give the source back")

	mgr.Redo()
	c.True(choice.Source.IsZero(), "redo must take the source away again")

	e, _ = buildEditorContent(template, group, initTraitEditor)
	e.editorData.Name = "Renamed"
	e.applyEdits()
	c.Equal(groupSource, group.Source, "an edit to a plain group must leave its source alone")
	mgr.Undo()
	c.Equal("Advantages", group.Name)
	c.Equal(groupSource, group.Source, "undoing an edit that left the source alone must leave it alone")
	mgr.Redo()
	c.Equal(groupSource, group.Source, "redoing an edit that left the source alone must leave it alone")
}

// TestGrantedItemEditorIsReadOnly checks that a granted item's editor only shows it.
func TestGrantedItemEditorIsReadOnly(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	screen.Do(func() {
		sheet, _, granted, stealth := newTestSheetWithGrant(t)
		DisplayNewDockable(sheet)
		c.True(EditSkill(sheet, granted).readOnly)
		c.False(EditSkill(sheet, stealth).readOnly)
	})
}

// TestNestedEditorTitle checks that an editor opened from a document is named for the document, while one opened from
// another editor, for a grant's item or for one of the item's modifiers, is named for its item.
func TestNestedEditorTitle(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	outer := newGrantTestTrait("Training", 0)
	inner := newGrantTestTrait("Night Vision", 0)
	modifier := gurps.NewTraitModifier(nil, nil, false)
	modifier.Name = "Mod"
	inner.Modifiers = []*gurps.TraitModifier{modifier}
	g := gurps.NewGrant(feature.GrantTrait)
	g.SetItem(inner, gurps.LibraryFile{})
	outer.Features = gurps.Features{g}
	data := gurps.NewTemplate()
	data.Traits = []*gurps.Trait{outer}
	isTraitEditor := func(d unison.Dockable) bool {
		_, ok := d.AsPanel().Self.(*editor[*gurps.Trait, *gurps.TraitEditData])
		return ok
	}
	var e *editor[*gurps.Trait, *gurps.TraitEditData]
	var outerTitle string
	screen.Do(func() {
		template := newTestTemplateDockable("Template", data)
		DisplayNewDockable(template)
		e = EditTrait(template, outer)
		outerTitle = e.Title()
		p := uxtest.PanelsOfType[*featuresPanel](e.content)[0]
		p.editGrantedItem(grantAt(t, p, 0))
	})
	nested := uxtest.OtherEditor(t, screen, e, isTraitEditor)
	var innerTitle, modTitle string
	screen.Do(func() {
		innerTitle = nested.Title()
		EditTraitModifier(nested, nested.editorData.Modifiers[0])
	})
	modEditor := uxtest.SoleEditor[*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData]](t, screen,
		func(d unison.Dockable) bool {
			_, ok := d.AsPanel().Self.(*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData])
			return ok
		})
	screen.Do(func() { modTitle = modEditor.Title() })
	c.Equal("Trait Editor for Template", outerTitle)
	c.Equal("Trait Editor for Night Vision", innerTitle)
	c.Equal("Trait Modifier Editor for Mod", modTitle)
}
