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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
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

// findFeaturesPanel returns the first features panel found anywhere beneath the given panel.
func findFeaturesPanel(p *unison.Panel) *featuresPanel {
	panel, _ := uxtest.FirstPanelOfType[*featuresPanel](p)
	return panel
}

// Both container and non-container traits offer the "Switched On" checkbox, since a container's modifiers can hold
// switchable features even though the container itself has none of its own.
func TestTraitEditorHasSwitchedOnCheckBox(t *testing.T) {
	for _, isContainer := range []bool{false, true} {
		name := "trait"
		if isContainer {
			name = "trait container"
		}
		t.Run(name, func(t *testing.T) {
			c := check.New(t)
			sheet := newTestSheetForTemplate(t)
			trait := gurps.NewTrait(sheet.Entity(), nil, isContainer)
			e, content := buildEditorContent(sheet, trait, initTraitEditor)
			box := findCheckBoxTitled(content, i18n.Text("Switched On"))
			c.NotNil(box, "expected a Switched On checkbox in the editor")
			c.False(e.editorData.SwitchedOn, "the switch starts out off")

			clickCheckBox(box, true)
			c.True(e.editorData.SwitchedOn, "checking the box must turn the switch on in the editor's data")
			c.False(trait.SwitchedOn, "the target must not be touched until the edit is applied")

			clickCheckBox(box, false)
			c.False(e.editorData.SwitchedOn, "clearing the box must turn the switch back off")
		})
	}
}

// findIntegerFieldLabeled returns the integer field that follows the label bearing the given text among the panel's
// direct children.
func findIntegerFieldLabeled(p *unison.Panel, labelText string) *IntegerField {
	children := p.Children()
	for i, child := range children {
		if label, ok := child.Self.(*unison.Label); ok && label.String() == labelText && i+1 < len(children) {
			if field, ok2 := children[i+1].Self.(*IntegerField); ok2 {
				return field
			}
		}
	}
	return nil
}

// A new trait container holds zero alternative slots, since only an alternative abilities container uses them, so the
// slots field's minimum must follow the container type. A fixed minimum of one showed the invalid zero's error color
// through the blanked field's border insets, and its clamping bumped the editor's copy to one, so the editor opened
// already modified.
func TestTraitEditorAlternativeSlotsFollowContainerType(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	trait := gurps.NewTrait(sheet.Entity(), nil, true)
	e, content := buildEditorContent(sheet, trait, initTraitEditor)
	field := findIntegerFieldLabeled(content, i18n.Text("Alternative Slots"))
	c.NotNil(field, "expected an Alternative Slots field in the trait container editor")
	c.Equal(0, e.editorData.AlternativeSlots, "a new container holds no slots")
	c.False(field.Invalid(), "zero slots is valid for a container that is not an alternative abilities one")
	c.False(field.Enabled(), "the field is blanked for a container that is not an alternative abilities one")
	c.Equal("", e.editorData.Ancestry, "a container that is not an ancestry one holds no ancestry")
	c.False(e.isModified(), "opening the editor must not modify the data")

	e.editorData.ContainerType = container.AlternativeAbilities
	e.MarkModified(nil)
	c.Equal(1, e.editorData.AlternativeSlots, "an alternative abilities container needs at least one slot")
	c.Equal("1", field.Text(), "the field must show the slot it was given")
	c.False(field.Invalid(), "one slot is valid")
	c.True(field.Enabled(), "the field is usable for an alternative abilities container")
	c.Equal(1, field.Min(), "an alternative abilities container may not have fewer than one slot")

	field.SetText("0")
	c.True(field.Invalid(), "zero slots is invalid for an alternative abilities container")
	field.SetText("3")
	c.False(field.Invalid(), "three slots is valid")
	c.Equal(3, e.editorData.AlternativeSlots, "the slot count must reach the editor's data")

	e.editorData.ContainerType = container.Group
	e.MarkModified(nil)
	c.False(field.Invalid(), "the slots are ignored for a group, so nothing in the field is invalid")
	c.False(field.Enabled(), "the field is blanked again for a group")
	c.Equal(0, field.Min(), "a group holds zero slots")
}

// Switching a container to an ancestry one gives it the default ancestry, since the empty ancestry it held until then
// is not something the popup can show.
func TestTraitEditorAncestryFollowsContainerType(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	trait := gurps.NewTrait(sheet.Entity(), nil, true)
	e, _ := buildEditorContent(sheet, trait, initTraitEditor)
	c.Equal("", e.editorData.Ancestry, "a group holds no ancestry")

	e.editorData.ContainerType = container.Ancestry
	e.MarkModified(nil)
	c.Equal(gurps.DefaultAncestry, e.editorData.Ancestry, "an ancestry container starts with the default ancestry")
}

// Both container and non-container equipment offer the "Switched On" checkbox -- a container holds features of its own
// as well as modifiers that can contribute switchable ones -- and it must sit with the editor's other checkbox rather
// than being stranded in the middle of the numeric fields.
func TestEquipmentEditorHasSwitchedOnCheckBox(t *testing.T) {
	for _, isContainer := range []bool{false, true} {
		name := "equipment"
		if isContainer {
			name = "equipment container"
		}
		t.Run(name, func(t *testing.T) {
			c := check.New(t)
			sheet := newTestSheetForTemplate(t)
			equipment := gurps.NewEquipment(sheet.Entity(), nil, isContainer)
			e, content := buildEditorContent(sheet, equipment, initEquipmentEditor(true))
			box := findCheckBoxTitled(content, i18n.Text("Switched On"))
			c.NotNil(box, "expected a Switched On checkbox in the editor")
			c.False(e.editorData.SwitchedOn, "the switch starts out off")

			clickCheckBox(box, true)
			c.True(e.editorData.SwitchedOn, "checking the box must turn the switch on in the editor's data")
			c.False(equipment.SwitchedOn, "the target must not be touched until the edit is applied")

			clickCheckBox(box, false)
			c.False(e.editorData.SwitchedOn, "clearing the box must turn the switch back off")

			// The checkbox and the empty panel that keeps it in the field column are added as a pair, so the widget two
			// places ahead of it is the editor's other checkbox.
			weightBox := findCheckBoxTitled(content, i18n.Text("Ignore weight for skills"))
			c.NotNil(weightBox, "expected an Ignore weight for skills checkbox in the editor")
			children := content.Children()
			i := slices.Index(children, box.AsPanel())
			c.True(i >= 2 && children[i-2] == weightBox.AsPanel(),
				"the Switched On checkbox must sit with the editor's other checkbox")
		})
	}
}

// A non-container spell can be given features, along with the switch that governs the switchable ones; a spell
// container, which has no features of its own, gets neither.
func TestSpellEditorHasFeaturesPanelAndSwitch(t *testing.T) {
	t.Run("spell", func(t *testing.T) {
		c := check.New(t)
		sheet := newTestSheetForTemplate(t)
		spell := gurps.NewSpell(sheet.Entity(), nil, false)
		e, content := buildEditorContent(sheet, spell, initSpellEditor)
		c.NotNil(findFeaturesPanel(content), "expected a features panel in the spell editor")
		box := findCheckBoxTitled(content, i18n.Text("Switched On"))
		c.NotNil(box, "expected a Switched On checkbox in the spell editor")

		clickCheckBox(box, true)
		c.True(e.editorData.SwitchedOn, "checking the box must turn the switch on in the editor's data")
	})

	t.Run("spell container", func(t *testing.T) {
		c := check.New(t)
		sheet := newTestSheetForTemplate(t)
		spell := gurps.NewSpell(sheet.Entity(), nil, true)
		_, content := buildEditorContent(sheet, spell, initSpellEditor)
		c.Nil(findFeaturesPanel(content), "a spell container must not offer a features panel")
		c.Nil(findCheckBoxTitled(content, i18n.Text("Switched On")),
			"a spell container has nothing to switch, so it must not offer the checkbox")
	})
}

// Only a skill that can hold features offers the switch.
func TestSkillEditorSwitchedOnCheckBoxOnlyForNonContainers(t *testing.T) {
	t.Run("skill", func(t *testing.T) {
		c := check.New(t)
		sheet := newTestSheetForTemplate(t)
		skill := gurps.NewSkill(sheet.Entity(), nil, false)
		e, content := buildEditorContent(sheet, skill, initSkillEditor)
		box := findCheckBoxTitled(content, i18n.Text("Switched On"))
		c.NotNil(box, "expected a Switched On checkbox in the skill editor")

		clickCheckBox(box, true)
		c.True(e.editorData.SwitchedOn, "checking the box must turn the switch on in the editor's data")
	})

	t.Run("skill container", func(t *testing.T) {
		c := check.New(t)
		sheet := newTestSheetForTemplate(t)
		skill := gurps.NewSkill(sheet.Entity(), nil, true)
		_, content := buildEditorContent(sheet, skill, initSkillEditor)
		c.Nil(findCheckBoxTitled(content, i18n.Text("Switched On")),
			"a skill container has nothing to switch, so it must not offer the checkbox")
	})
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

// focusRecordingPanel records whether it was asked to take the focus without scrolling.
type focusRecordingPanel struct {
	unison.Panel
	withoutScroll int
}

func (p *focusRecordingPanel) RequestFocusWithoutScroll() { p.withoutScroll++ }

// Giving the focus back to a table after an editor closes must not go through the table's default focus handling, which
// scrolls the whole table into view and so moved the sheet even when the edited row was already on screen.
func TestRestoreFocusSkipsTableScrolling(t *testing.T) {
	c := check.New(t)
	var _ focusWithoutScroller = (*unison.Table[*Node[*gurps.Trait]])(nil)

	p := &focusRecordingPanel{}
	p.Self = p
	restoreFocus(p.AsPanel())
	c.Equal(1, p.withoutScroll, "a panel that can take the focus without scrolling must be asked to do so")

	plain := unison.NewPanel()
	plain.SetFocusable(true)
	restoreFocus(plain) // No window, so this must simply do nothing rather than panic.
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

// TestReadOnlyEditor checks that an editor opened from a read-only one is read-only too, and that it leaves on only what
// reads or opens its rows, through rebuilds, offers no substitutions, no change to its source and no new items, has a
// modifier table whose rows open but don't change, and never reports a change to apply.
func TestReadOnlyEditor(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	screen.Do(func() {
		sheet := newTestSheetForTemplate(t)
		entity := sheet.Entity()
		host := gurps.NewTrait(entity, nil, false)
		trait := gurps.NewTrait(entity, nil, false)
		trait.Name = "Ally (@Who@)"
		trait.CanLevel = true
		trait.Levels = fxp.One
		trait.Source = testSource
		modifier := gurps.NewTraitModifier(entity, nil, false)
		trait.Modifiers = []*gurps.TraitModifier{modifier}
		entity.Traits = []*gurps.Trait{host, trait}
		sheet.Rebuild(true)
		DisplayNewDockable(sheet)
		hostEditor := EditTrait(sheet, host)
		hostEditor.makeReadOnly()
		for _, e := range []*editor[*gurps.Trait, *gurps.TraitEditData]{EditTrait(sheet, trait), EditTrait(hostEditor, trait)} {
			c.Equal(e.owner == hostEditor, e.readOnly)
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
			e.editorData.Name = trait.Name
		}
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
