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
	"encoding/json/v2"
	"testing"
	"testing/fstest"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
)

// TestTemplatePickerLoadsUnknownTypeAsNotApplicable verifies what lets the rest of the code read a picker's type
// without checking it first: a type this version doesn't know, as a newer or hand-edited file might hold, loads as
// NotApplicable, leaving the container with no picker at all rather than one with a type nothing can handle.
func TestTemplatePickerLoadsUnknownTypeAsNotApplicable(t *testing.T) {
	c := check.New(t)
	var tp TemplatePicker
	c.NoError(json.Unmarshal([]byte(`{"type":"not_a_real_picker_type"}`), &tp))
	c.Equal(picker.NotApplicable, tp.Type)
	c.True(tp.IsZero())
}

// TestHasTemplatePickerData verifies that a node carrying template choices is spotted
func TestHasTemplatePickerData(t *testing.T) {
	c := check.New(t)
	plain := NewTrait(nil, nil, false)
	plain.Name = "Claws"
	c.False(HasTemplatePickerData(plain), "a plain trait carries no choices")

	emptyContainer := NewTrait(nil, nil, true)
	emptyContainer.Name = "Advantages"
	c.False(HasTemplatePickerData(emptyContainer), "a container without choices carries none")

	choices := newTemplateChoiceTrait("Pick One", "First", "Second")
	c.True(HasTemplatePickerData(plain, choices))

	emptyContainer.Children = []*Trait{choices}
	choices.SetParent(emptyContainer)
	c.True(HasTemplatePickerData(emptyContainer), "choices nested deeper must be found as well")

	skillContainer := NewSkill(nil, nil, true)
	skillContainer.Name = "Techniques"
	c.False(HasTemplatePickerData(skillContainer))

	skillContainer.TemplatePicker.Type = picker.Points
	c.True(HasTemplatePickerData(skillContainer), "skills carry choices too")
}

// TestClearTemplatePickerData verifies that the choices are removed from every container beneath the rows as well, since
// a copied container brings its whole subtree with it.
func TestClearTemplatePickerData(t *testing.T) {
	c := check.New(t)
	outer := NewTrait(nil, nil, true)
	outer.Name = "Advantages"
	inner := newTemplateChoiceTrait("Pick One", "First", "Second")
	inner.SetParent(outer)
	outer.Children = []*Trait{inner}

	// Verify the test data has picker data
	c.True(HasTemplatePickerData(outer))

	// Clear any template picker data
	ClearTemplatePickerData(outer)

	// Verify the test data no longer carries picker data
	c.False(HasTemplatePickerData(outer), "no picker data may be left")

	// Verify we still have the same data otherwise (this could use more checks)
	c.Equal(2, len(inner.Children), "clearing must not disturb anything else")
}

// TestClearTemplatePickerDataClearsSource verifies that a container losing its choices also loses its source, while
// the rest of the subtree keeps theirs. Only a template may hold choices and a template is never a source, so the source
// a container with choices points at can't have them, and syncing with it would quietly take them away.
func TestClearTemplatePickerDataClearsSource(t *testing.T) {
	c := check.New(t)
	outer := NewTrait(nil, nil, true)
	outer.Name = "Advantages"
	outer.Source = Source{Library: "lib", Path: "outer.adq", TID: outer.ID()}
	inner := newTemplateChoiceTrait("Pick One", "First", "Second")
	inner.Source = Source{Library: "lib", Path: "inner.adq", TID: inner.ID()}
	inner.SetParent(outer)
	outer.Children = []*Trait{inner}
	child := inner.Children[0]
	child.Source = Source{Library: "lib", Path: "child.adq", TID: child.ID()}

	ClearTemplatePickerData(outer)

	c.Equal(Source{}, inner.Source, "the container that lost its choices must lose its source as well")
	c.NotEqual(Source{}, outer.Source, "a container that had no choices must keep its source")
	c.NotEqual(Source{}, child.Source, "an option must keep its source")
}

// newTemplateChoiceTrait returns a trait container carrying template choices, holding a child for each of the given
// names.
func newTemplateChoiceTrait(name string, childNames ...string) *Trait {
	container := NewTrait(nil, nil, true)
	container.Name = name
	container.TemplatePicker.Type = picker.Count
	container.TemplatePicker.Qualifier.Compare = criteria.EqualsNumber
	container.TemplatePicker.Qualifier.Qualifier = fxp.One
	children := make([]*Trait, 0, len(childNames))
	for _, childName := range childNames {
		child := NewTrait(nil, container, false)
		child.Name = childName
		children = append(children, child)
	}
	container.Children = children
	return container
}

func newPickerModTestContainer(name string, children ...*Trait) *Trait {
	container := NewTrait(nil, nil, true)
	container.Name = name
	container.Children = children
	for _, child := range children {
		child.SetParent(container)
	}
	return container
}

func newPickerModTestPicker(name string, children ...*Trait) *Trait {
	container := newPickerModTestContainer(name, children...)
	container.TemplatePicker.Type = picker.Points
	container.TemplatePicker.Qualifier.Compare = criteria.EqualsNumber
	container.TemplatePicker.Qualifier.Qualifier = fxp.FromInteger(20)
	return container
}

func newPickerModTestLeaf(name string, basePoints int) *Trait {
	trait := NewTrait(nil, nil, false)
	trait.Name = name
	trait.BasePoints = fxp.FromInteger(basePoints)
	return trait
}

func newPickerModTestLeveled(name string, basePoints, levels, pointsPerLevel int) *Trait {
	trait := newPickerModTestLeaf(name, basePoints)
	trait.CanLevel = true
	trait.Levels = fxp.FromInteger(levels)
	trait.PointsPerLevel = fxp.FromInteger(pointsPerLevel)
	return trait
}

func newPickerModTestModifier(name, costAdj string) *TraitModifier {
	mod := NewTraitModifier(nil, nil, false)
	mod.Name = name
	mod.CostAdj = costAdj
	return mod
}

func modifierNames(mods []*TraitModifier) []string {
	names := make([]string, 0, len(mods))
	for _, one := range mods {
		names = append(names, one.Name)
	}
	return names
}

// checkProjectedCopy verifies that a projected modifier is a fresh copy of the original: attached to the row it landed
// on, with an ID of its own, and otherwise the same as the original.
func checkProjectedCopy(c check.Checker, original, projected *TraitModifier, row *Trait, seen map[tid.TID]bool) {
	c.True(original != projected, "the copy must not be the original itself")
	c.NotEqual(original.TID, projected.TID, "the copy must have its own ID")
	c.True(tid.IsKindAndValid(projected.TID, traitModifierKind(original.Container())))
	c.False(seen[projected.TID], "each copy must have an ID not shared with any other copy")
	seen[projected.TID] = true
	c.True(row == projected.Target(), "the copy must be attached to the row it landed on")
	c.Equal(original.Name, projected.Name)
	c.Equal(original.CostAdj, projected.CostAdj)
	c.Equal(original.Disabled, projected.Disabled)
	c.Equal(original.PageRef, projected.PageRef)
	c.Equal(original.LocalNotes, projected.LocalNotes)
	c.Equal(original.Source, projected.Source, "the copy must keep the original's source")
	c.Equal(len(original.Children), len(projected.Children))
	for i, child := range projected.Children {
		c.True(projected == child.Parent(), "a nested copy must belong to the copy of its container")
		checkProjectedCopy(c, original.Children[i], child, row, seen)
	}
}

// TestProjectPickerModifiersOntoLeaves verifies the plain case: each child of a choice container gets its own copy of
// the container's modifiers, and the container is left with none.
func TestProjectPickerModifiersOntoLeaves(t *testing.T) {
	c := check.New(t)
	own := newPickerModTestModifier("Own", "+5%")
	first := newPickerModTestLeaf("First", 10)
	first.AddModifiers(own)
	second := newPickerModTestLeaf("Second", 5)
	third := newPickerModTestLeaf("Third", 15)
	pickerContainer := newPickerModTestPicker("20 points from", first, second, third)
	necromancy := newPickerModTestModifier("Necromancy", "-10%")
	necromancy.PageRef = "DF9:15"
	necromancy.LocalNotes = "Power modifier"
	necromancy.Source = Source{Library: "lib", Path: "mods.adm", TID: necromancy.TID}
	pickerContainer.AddModifiers(necromancy)

	projectPickerModifiersDownward([]*Trait{pickerContainer})

	c.Equal(0, len(pickerContainer.Modifiers), "the choice container must be left without modifiers")
	seen := make(map[tid.TID]bool)
	for _, child := range pickerContainer.Children {
		c.Equal(1, countNamed(child.AllModifiers(), necromancy.Name), "the modifier must apply exactly once")
		checkProjectedCopy(c, necromancy, child.Modifiers[len(child.Modifiers)-1], child, seen)
	}
	c.Equal([]string{"Own", "Necromancy"}, modifierNames(first.Modifiers), "copies must follow a row's own modifiers")
}

// TestProjectPickerModifiersPassesThroughNestedPickers verifies that a choice container beneath another is not a
// resting place, since it is dissolved in turn: the copies land on the rows beneath it instead, at any depth.
func TestProjectPickerModifiersPassesThroughNestedPickers(t *testing.T) {
	for _, depth := range []int{2, 3} {
		c := check.New(t)
		leaves := []*Trait{newPickerModTestLeaf("First", 10), newPickerModTestLeaf("Second", 5)}
		innermost := newPickerModTestPicker("Level 1", leaves...)
		pickers := []*Trait{innermost}
		outer := innermost
		for i := 2; i <= depth; i++ {
			outer = newPickerModTestPicker("Level", outer)
			pickers = append(pickers, outer)
		}
		mod := newPickerModTestModifier("Power", "-10%")
		outer.AddModifiers(mod)

		projectPickerModifiersDownward([]*Trait{outer})

		for _, one := range pickers {
			c.Equal(0, len(one.Modifiers), "no choice container may hold the modifier, depth %d", depth)
		}
		seen := make(map[tid.TID]bool)
		for _, leaf := range leaves {
			c.Equal(1, len(leaf.Modifiers), "depth %d", depth)
			checkProjectedCopy(c, mod, leaf.Modifiers[0], leaf, seen)
			c.Equal(1, len(leaf.AllModifiers()), "depth %d", depth)
		}
	}
}

// TestProjectPickerModifiersRestsOnOrdinaryContainer verifies that an ordinary container beneath a choice container
// takes the copy itself, since it survives the dissolve and passes the modifier on to everything beneath it.
func TestProjectPickerModifiersRestsOnOrdinaryContainer(t *testing.T) {
	c := check.New(t)
	luck := newPickerModTestLeaf("Luck", 15)
	extraordinary := newPickerModTestLeaf("Extraordinary Luck", 30)
	either := newPickerModTestContainer("Luck or Extraordinary Luck", luck, extraordinary)
	pickerContainer := newPickerModTestPicker("15 points in Shamanic Abilities", newPickerModTestLeaf("Channeling", 10),
		either)
	mod := newPickerModTestModifier("Shamanic Gift", "-10%")
	pickerContainer.AddModifiers(mod)

	projectPickerModifiersDownward([]*Trait{pickerContainer})

	c.Equal(0, len(pickerContainer.Modifiers))
	c.Equal([]string{"Shamanic Gift"}, modifierNames(either.Modifiers), "the ordinary container must take the copy")
	for _, leaf := range either.Children {
		c.Equal(0, len(leaf.Modifiers), "the copy must not be pushed past an ordinary container")
		c.Equal([]string{"Shamanic Gift"}, modifierNames(leaf.AllModifiers()), "the modifier must apply exactly once")
	}
}

// TestProjectPickerModifiersIsIdempotent verifies that projecting a second time changes nothing, since it happens on
// every load and the file isn't rewritten until it is saved.
func TestProjectPickerModifiersIsIdempotent(t *testing.T) {
	c := check.New(t)
	either := newPickerModTestContainer("Blessed or Very Blessed", newPickerModTestLeaf("Blessed", 10),
		newPickerModTestLeaf("Very Blessed", 20))
	inner := newPickerModTestPicker("Inner", newPickerModTestLeaf("Deep", 5))
	inner.AddModifiers(newPickerModTestModifier("Inner Mod", "-5%"))
	outer := newPickerModTestPicker("Outer", newPickerModTestLeaf("First", 10), either, inner)
	outer.AddModifiers(newPickerModTestModifier("Outer Mod", "-10%"))
	rows := []*Trait{outer}

	projectPickerModifiersDownward(rows)
	var before []string
	var ids []tid.TID
	Traverse(func(one *Trait) bool {
		before = append(before, one.Name+":"+joinNames(modifierNames(one.Modifiers)))
		for _, mod := range one.Modifiers {
			ids = append(ids, mod.TID)
		}
		return false
	}, false, false, rows...)

	projectPickerModifiersDownward(rows)
	var after []string
	var idsAfter []tid.TID
	Traverse(func(one *Trait) bool {
		after = append(after, one.Name+":"+joinNames(modifierNames(one.Modifiers)))
		for _, mod := range one.Modifiers {
			idsAfter = append(idsAfter, mod.TID)
		}
		return false
	}, false, false, rows...)

	c.Equal(before, after, "a second projection must not change anything")
	c.Equal(ids, idsAfter, "a second projection must not replace any modifier")
	c.Equal([]string{"Inner Mod", "Outer Mod"}, modifierNames(inner.Children[0].Modifiers),
		"each modifier must be present exactly once")
}

func countNamed(mods []*TraitModifier, name string) int {
	count := 0
	for _, one := range mods {
		if one.Name == name {
			count++
		}
	}
	return count
}

func joinNames(names []string) string {
	var result string
	for i, name := range names {
		if i != 0 {
			result += ","
		}
		result += name
	}
	return result
}

// TestProjectPickerModifiersPreservesCost verifies the property that matters most: a row beneath a choice container
// costs the same after projection as it would beneath an ordinary container holding the same modifier, and the same as
// it did before projection, while the choice container was still being walked through by AllModifiers.
func TestProjectPickerModifiersPreservesCost(t *testing.T) {
	c := check.New(t)
	build := func() *Trait {
		container := newPickerModTestContainer("20 Points of Necromantic Abilities",
			newPickerModTestLeaf("Death Vision", 2),
			newPickerModTestLeaf("Spirit Empathy", 10),
			newPickerModTestLeveled("Magery", 5, 3, 10),
			newPickerModTestLeveled("Resistant to Disease", 0, 1, 7),
			newPickerModTestContainer("Luck or Extraordinary Luck", newPickerModTestLeaf("Luck", 15),
				newPickerModTestLeveled("Extra Life", 0, 2, 25)),
		)
		container.AddModifiers(newPickerModTestModifier("Necromancy", "-10%"))
		return container
	}
	ordinary := build()
	pickerContainer := build()
	pickerContainer.TemplatePicker.Type = picker.Points
	pickerContainer.TemplatePicker.Qualifier.Compare = criteria.EqualsNumber
	pickerContainer.TemplatePicker.Qualifier.Qualifier = fxp.FromInteger(20)

	unmodified := build()
	unmodified.Modifiers = nil

	collect := func(root *Trait) map[string]fxp.Int {
		points := make(map[string]fxp.Int)
		Traverse(func(one *Trait) bool {
			if one != root {
				points[one.Name] = one.AdjustedPoints(nil)
			}
			return false
		}, false, false, root.Children...)
		return points
	}

	expected := collect(ordinary)
	c.Equal(expected, collect(pickerContainer), "before projection, AllModifiers finds the container's modifier")
	c.NotEqual(expected, collect(unmodified), "the modifier must make a difference for the test to mean anything")

	projectPickerModifiersDownward([]*Trait{pickerContainer})

	c.Equal(0, len(pickerContainer.Modifiers))
	c.Equal(expected, collect(pickerContainer), "projection must not change what any row costs")

	// Once the container is dissolved, as applying the template does, the chosen rows must still cost the same.
	for _, child := range pickerContainer.Children {
		child.SetParent(nil)
	}
	c.Equal(expected, collect(pickerContainer), "the chosen rows must keep the modifier once the container is gone")
}

// TestProjectPickerModifiersKeepsInheritedOrder verifies that projecting nested choice containers leaves a row's
// inherited modifiers in the order AllModifiers reported them before projection, innermost container first.
func TestProjectPickerModifiersKeepsInheritedOrder(t *testing.T) {
	c := check.New(t)
	leaf := newPickerModTestLeaf("Leaf", 10)
	leaf.AddModifiers(newPickerModTestModifier("Own", "+5%"))
	inner := newPickerModTestPicker("Inner", leaf)
	inner.AddModifiers(newPickerModTestModifier("Inner", "-5%"))
	outer := newPickerModTestPicker("Outer", inner)
	outer.AddModifiers(newPickerModTestModifier("Outer", "-10%"))
	before := modifierNames(leaf.AllModifiers())

	projectPickerModifiersDownward([]*Trait{outer})

	c.Equal(before, modifierNames(leaf.AllModifiers()))
	c.Equal([]string{"Own", "Inner", "Outer"}, modifierNames(leaf.Modifiers))
}

// TestProjectPickerModifiersWithNowhereToGo verifies that a choice container with no row for its modifiers to land on
// keeps its enabled ones, rather than losing them, though its disabled ones are still removed.
func TestProjectPickerModifiersWithNowhereToGo(t *testing.T) {
	c := check.New(t)
	empty := newPickerModTestPicker("Empty")
	disabled := newPickerModTestModifier("Disabled", "-5%")
	disabled.Disabled = true
	empty.AddModifiers(newPickerModTestModifier("Kept", "-10%"), disabled)
	emptyInner := newPickerModTestPicker("Empty Inner")
	onlyEmptyPickers := newPickerModTestPicker("Only Empty Pickers", emptyInner)
	onlyEmptyPickers.AddModifiers(newPickerModTestModifier("Also Kept", "-10%"))

	projectPickerModifiersDownward([]*Trait{empty, onlyEmptyPickers})

	c.Equal([]string{"Kept"}, modifierNames(empty.Modifiers))
	c.Equal([]string{"Also Kept"}, modifierNames(onlyEmptyPickers.Modifiers))
	c.Equal(0, len(emptyInner.Modifiers))
}

// TestProjectPickerModifiersCopiesModifierContainers verifies that a modifier container is copied along with the
// enabled modifiers in it, each part with its own ID and the original's source, and that the disabled ones, and any
// container left empty without them, are removed.
func TestProjectPickerModifiersCopiesModifierContainers(t *testing.T) {
	c := check.New(t)
	group := NewTraitModifier(nil, nil, true)
	group.Name = "Power Modifiers"
	group.PageRef = "B254"
	first := newPickerModTestModifier("Psionic", "-10%")
	first.SetParent(group)
	first.Source = Source{Library: "lib", Path: "mods.adm", TID: first.TID}
	nested := NewTraitModifier(nil, group, true)
	nested.Name = "Nested"
	second := newPickerModTestModifier("Unconscious Only", "-20%")
	second.SetParent(nested)
	third := newPickerModTestModifier("Disabled", "-5%")
	third.SetParent(nested)
	third.Disabled = true
	nested.Children = []*TraitModifier{second, third}
	onlyDisabled := NewTraitModifier(nil, group, true)
	onlyDisabled.Name = "Only Disabled"
	fourth := newPickerModTestModifier("Also Disabled", "-5%")
	fourth.SetParent(onlyDisabled)
	fourth.Disabled = true
	onlyDisabled.Children = []*TraitModifier{fourth}
	empty := NewTraitModifier(nil, group, true)
	empty.Name = "Empty"
	group.Children = []*TraitModifier{first, nested, onlyDisabled, empty}
	emptyAtTop := NewTraitModifier(nil, nil, true)
	emptyAtTop.Name = "Empty At Top"

	leaves := []*Trait{newPickerModTestLeaf("First", 10), newPickerModTestLeaf("Second", 5)}
	pickerContainer := newPickerModTestPicker("Pick", leaves...)
	pickerContainer.AddModifiers(group, emptyAtTop)

	projectPickerModifiersDownward([]*Trait{pickerContainer})

	c.Equal(0, len(pickerContainer.Modifiers))
	seen := make(map[tid.TID]bool)
	for _, leaf := range leaves {
		c.Equal(1, len(leaf.Modifiers), "an empty modifier container must not be projected")
		projected := leaf.Modifiers[0]
		c.True(tid.IsKind(projected.TID, kinds.TraitModifierContainer))
		c.Equal([]string{"Psionic", "Nested"}, modifierNames(projected.Children))
		c.Equal([]string{"Unconscious Only"}, modifierNames(projected.Children[1].Children))
		checkProjectedCopy(c, group, projected, leaf, seen)
	}
}

// TestProjectPickerModifiersRemovesDisabled verifies that a disabled modifier is neither projected nor left on the
// choice container, including when it is the only modifier the container has.
func TestProjectPickerModifiersRemovesDisabled(t *testing.T) {
	c := check.New(t)
	leaves := []*Trait{newPickerModTestLeaf("First", 10), newPickerModTestLeaf("Second", 5)}
	pickerContainer := newPickerModTestPicker("20 points from", leaves...)
	disabled := newPickerModTestModifier("Unconscious Only", "-20%")
	disabled.Disabled = true
	pickerContainer.AddModifiers(disabled, newPickerModTestModifier("Necromancy", "-10%"))

	onlyDisabledLeaves := []*Trait{newPickerModTestLeaf("Third", 10), newPickerModTestLeaf("Fourth", 5)}
	onlyDisabled := newPickerModTestPicker("30 points from", onlyDisabledLeaves...)
	placeholder := newPickerModTestModifier("Modifier", "")
	placeholder.Disabled = true
	onlyDisabled.AddModifiers(placeholder)

	projectPickerModifiersDownward([]*Trait{pickerContainer, onlyDisabled})

	c.Equal(0, len(pickerContainer.Modifiers))
	for _, leaf := range leaves {
		c.Equal([]string{"Necromancy"}, modifierNames(leaf.Modifiers), "only the enabled modifier may be projected")
	}
	c.Equal(0, len(onlyDisabled.Modifiers), "a disabled modifier must be removed even with nothing to project")
	for i, leaf := range onlyDisabledLeaves {
		c.Equal(0, len(leaf.Modifiers))
		c.False(leaf.Preconfigured, "a row that received nothing must not be marked")
		c.Equal(fxp.FromInteger([]int{10, 5}[i]), leaf.AdjustedPoints(nil))
	}
}

// TestProjectPickerModifiersCarriesReplacements verifies that the values the choice container supplies for the
// markers its projected modifiers use go along with the copies, without overriding a row's own values or bringing
// along any the projected modifiers don't use, such as those of a disabled modifier.
func TestProjectPickerModifiersCarriesReplacements(t *testing.T) {
	c := check.New(t)
	first := newPickerModTestLeaf("First", 10)
	second := newPickerModTestLeaf("Second", 5)
	second.Replacements = map[string]string{"Power": "Divine"}
	pickerContainer := newPickerModTestPicker("Pick @Count@", first, second)
	pickerContainer.Replacements = map[string]string{"Power": "Psionic", "Count": "20"}
	mod := newPickerModTestModifier("@Power@", "-10%")
	disabled := newPickerModTestModifier("@Style@", "-5%")
	disabled.Disabled = true
	pickerContainer.AddModifiers(mod, disabled)
	pickerContainer.Replacements["Style"] = "Chi"

	projectPickerModifiersDownward([]*Trait{pickerContainer})

	c.Equal(map[string]string{"Power": "Psionic"}, first.Replacements)
	c.Equal("Psionic", first.Modifiers[0].NameWithReplacements())
	c.Equal(map[string]string{"Power": "Divine"}, second.Replacements,
		"a row's own value must win")
	c.Equal("Divine", second.Modifiers[0].NameWithReplacements())
}

// TestProjectPickerModifiersOnLoad verifies that templates and trait lists are projected as they are loaded, and that
// choice containers holding skills and spells, which have no modifiers, are left alone.
func TestProjectPickerModifiersOnLoad(t *testing.T) {
	c := check.New(t)
	newTraits := func() []*Trait {
		pickerContainer := newPickerModTestPicker("20 points from", newPickerModTestLeaf("First", 10),
			newPickerModTestLeaf("Second", 5))
		pickerContainer.AddModifiers(newPickerModTestModifier("Necromancy", "-10%"))
		return []*Trait{newPickerModTestContainer("Advantages", pickerContainer)}
	}
	checkTraits := func(traits []*Trait) {
		pickerContainer := traits[0].Children[0]
		c.Equal(0, len(pickerContainer.Modifiers))
		for _, child := range pickerContainer.Children {
			c.Equal([]string{"Necromancy"}, modifierNames(child.Modifiers))
			c.True(child == child.Modifiers[0].Target())
		}
	}

	skillPicker := NewSkill(nil, nil, true)
	skillPicker.Name = "Pick a skill"
	skillPicker.TemplatePicker.Type = picker.Count
	skillPicker.TemplatePicker.Qualifier.Compare = criteria.EqualsNumber
	skillPicker.TemplatePicker.Qualifier.Qualifier = fxp.One
	skill := NewSkill(nil, skillPicker, false)
	skill.Name = "Occultism"
	skillPicker.Children = []*Skill{skill}
	spellPicker := NewSpell(nil, nil, true)
	spellPicker.Name = "Pick a spell"
	spellPicker.TemplatePicker = skillPicker.TemplatePicker
	spell := NewSpell(nil, spellPicker, false)
	spell.Name = "Zombie"
	spellPicker.Children = []*Spell{spell}

	tmpl := NewTemplate()
	tmpl.Traits = newTraits()
	tmpl.Skills = []*Skill{skillPicker}
	tmpl.Spells = []*Spell{spellPicker}
	data, err := json.Marshal(tmpl)
	c.NoError(err)
	loaded, err := NewTemplateFromFile(fstest.MapFS{"Test.gct": {Data: data}}, "Test.gct")
	c.NoError(err)
	checkTraits(loaded.Traits)
	c.Equal(1, len(loaded.Skills))
	c.False(loaded.Skills[0].TemplatePicker.IsZero(), "a skill choice container must be left alone")
	c.Equal(1, len(loaded.Skills[0].Children))
	c.Equal(1, len(loaded.Spells))
	c.False(loaded.Spells[0].TemplatePicker.IsZero(), "a spell choice container must be left alone")
	c.Equal(1, len(loaded.Spells[0].Children))

	data, err = json.Marshal(&listData[*Trait]{Version: jio.CurrentDataVersion, Rows: newTraits()})
	c.NoError(err)
	traits, err := NewTraitsFromFile(fstest.MapFS{"Test.adq": {Data: data}}, "Test.adq")
	c.NoError(err)
	checkTraits(traits)
}

// TestProjectPickerModifiersMarksPreconfigured verifies that a row is marked as preconfigured only when it had no
// modifiers of its own before any choice container above it projected onto it.
func TestProjectPickerModifiersMarksPreconfigured(t *testing.T) {
	c := check.New(t)
	bare := newPickerModTestLeaf("Bare", 10)
	withOwn := newPickerModTestLeaf("With Own", 10)
	withOwn.AddModifiers(newPickerModTestModifier("Own", "+5%"))
	authorMarked := newPickerModTestLeaf("Author Marked", 10)
	authorMarked.AddModifiers(newPickerModTestModifier("Own", "+5%"))
	authorMarked.Preconfigured = true
	either := newPickerModTestContainer("Either", newPickerModTestLeaf("Luck", 15))
	disabled := newPickerModTestModifier("Placeholder", "")
	disabled.Disabled = true
	enabledPicker := newPickerModTestPicker("Enabled", bare, withOwn, authorMarked, either)
	enabledPicker.AddModifiers(newPickerModTestModifier("Power", "-10%"), disabled)

	// The inner container's copies reach the row first, but it had no modifiers of its own before that, so the outer
	// container's copies must not undo the mark.
	twiceTarget := newPickerModTestLeaf("Twice Target", 10)
	twiceWithOwn := newPickerModTestLeaf("Twice With Own", 10)
	twiceWithOwn.AddModifiers(newPickerModTestModifier("Own", "+5%"))
	inner := newPickerModTestPicker("Inner", twiceTarget, twiceWithOwn)
	inner.AddModifiers(newPickerModTestModifier("Inner", "-5%"))
	outer := newPickerModTestPicker("Outer", inner)
	outer.AddModifiers(newPickerModTestModifier("Outer", "-10%"))

	projectPickerModifiersDownward([]*Trait{enabledPicker, outer})

	c.True(bare.Preconfigured, "a bare row must be marked")
	c.True(either.Preconfigured, "a bare ordinary container must be marked")
	c.False(either.Children[0].Preconfigured, "rows beneath the one that received the copies are left alone")
	c.False(withOwn.Preconfigured, "a row with modifiers of its own must still be asked about them")
	c.True(authorMarked.Preconfigured, "a row the author marked must keep its mark")
	c.True(twiceTarget.Preconfigured, "a row that was bare before any projection must be marked")
	c.False(twiceWithOwn.Preconfigured)
	c.False(enabledPicker.Preconfigured)
}
