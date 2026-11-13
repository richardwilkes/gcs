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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// modifierPrompt records one invocation of a modifier enable/disable prompt.
type modifierPrompt struct {
	title     string
	modifiers []string
}

// captureModifierPrompts substitutes non-interactive modifier prompts that record what they were asked to show,
// returning the accumulator they append to. The real prompts are restored when the test finishes.
func captureModifierPrompts(t *testing.T) *[]modifierPrompt {
	t.Helper()
	var prompts []modifierPrompt
	uxtest.SwapForTest(t, &promptForTraitModifiers, func(info *modifierPromptInfo, modifiers []*gurps.TraitModifier) (changed, canceled bool) {
		title := info.name
		p := modifierPrompt{title: title}
		for _, one := range modifiers {
			p.modifiers = append(p.modifiers, one.Name)
		}
		prompts = append(prompts, p)
		return false, false
	})
	uxtest.SwapForTest(t, &promptForEquipmentModifiers, func(info *modifierPromptInfo, modifiers []*gurps.EquipmentModifier) (changed, canceled bool) {
		title := info.name
		p := modifierPrompt{title: title}
		for _, one := range modifiers {
			p.modifiers = append(p.modifiers, one.Name)
		}
		prompts = append(prompts, p)
		return false, false
	})
	return &prompts
}

// namedTraits fills the entity's traits list with one non-container trait per name and returns them.
func namedTraits(entity *gurps.Entity, names ...string) []*gurps.Trait {
	traits := make([]*gurps.Trait, len(names))
	for i, name := range names {
		traits[i] = gurps.NewTrait(entity, nil, false)
		traits[i].Name = name
	}
	entity.Traits = traits
	return traits
}

// namedEquipment fills the entity's carried equipment list with one non-container item per name and returns them.
func namedEquipment(entity *gurps.Entity, names ...string) []*gurps.Equipment {
	equipment := make([]*gurps.Equipment, len(names))
	for i, name := range names {
		equipment[i] = gurps.NewEquipment(entity, nil, false)
		equipment[i].Name = name
	}
	entity.CarriedEquipment = equipment
	return equipment
}

// entityWithNamedTraits returns a new entity whose traits list holds one non-container trait per name, along with those
// traits.
func entityWithNamedTraits(names ...string) (*gurps.Entity, []*gurps.Trait) {
	entity := gurps.NewEntity()
	return entity, namedTraits(entity, names...)
}

// entityWithNamedEquipment returns a new entity whose carried equipment list holds one non-container item per name,
// along with those items.
func entityWithNamedEquipment(names ...string) (*gurps.Entity, []*gurps.Equipment) {
	entity := gurps.NewEntity()
	return entity, namedEquipment(entity, names...)
}

// tabledProvider gives the provider the table and root rows that NewNodeTable would, then hands the provider back, so
// that a provider a test only needs in order to drive a drop can be had in a single expression.
func tabledProvider[T gurps.Node[T]](provider TableProvider[T]) TableProvider[T] {
	newProviderTable(provider)
	return provider
}

// forbidModifierPrompts substitutes modifier and nameables prompts that fail the test if shown, so that a test which
// expects none fails rather than waits on a real dialog. A test that expects the nameables prompt substitutes its own
// responder after calling this.
func forbidModifierPrompts(t *testing.T) {
	t.Helper()
	uxtest.SwapForTest(t, &promptForTraitModifiers, func(info *modifierPromptInfo, _ []*gurps.TraitModifier) (changed, canceled bool) {
		title := info.name
		t.Errorf("the modifier prompt must not be shown, but was shown for %q", title)
		return false, false
	})
	uxtest.SwapForTest(t, &promptForEquipmentModifiers, func(info *modifierPromptInfo, _ []*gurps.EquipmentModifier) (changed, canceled bool) {
		title := info.name
		t.Errorf("the modifier prompt must not be shown, but was shown for %q", title)
		return false, false
	})
	uxtest.SwapForTest(t, &promptForNameables, slicedNameablesPrompt(func(titles []string, _ []map[string]string, _ [][]string) bool {
		t.Errorf("the nameables prompt must not be shown, but was shown for %v", titles)
		return false
	}))
}

// processModifiers prompts for the modifiers of the rows as applyTransfer does.
func processModifiers[T gurps.Node[T]](rows []T, requirePicks bool) bool {
	targets := modifierTargets(rows, requirePicks, nil)
	return promptForModifierTargets(promptOperation{}, targets, &modifierPromptCount{total: len(targets)}, requirePicks,
		nil, nil)
}

// TestProcessModifiersIgnoresModifierRows documents that processModifiers only has something to do for rows that can
// hold modifiers. Handing it the modifiers themselves matches nothing, which is why its callers pass the rows that
// carry the modifiers (see applyTransfer).
func TestProcessModifiersIgnoresModifierRows(t *testing.T) {
	c := check.New(t)
	prompts := captureModifierPrompts(t)
	entity := gurps.NewEntity()

	traitMod := gurps.NewTraitModifier(entity, nil, false)
	traitMod.Name = "Trait Modifier"
	processModifiers([]*gurps.TraitModifier{traitMod}, true)
	equipmentMod := gurps.NewEquipmentModifier(entity, nil, false)
	equipmentMod.Name = "Equipment Modifier"
	processModifiers([]*gurps.EquipmentModifier{equipmentMod}, true)
	c.Equal(0, len(*prompts), "modifier rows have no modifiers of their own to prompt for")

	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Trait"
	trait.Modifiers = []*gurps.TraitModifier{traitMod}
	processModifiers([]*gurps.Trait{trait}, true)
	c.Equal([]modifierPrompt{{title: "Trait", modifiers: []string{"Trait Modifier"}}}, *prompts,
		"a trait must be prompted for with its own modifiers")
}

// TestModifierPromptsCountOnlyRowsWithModifiers verifies that a row without modifiers is neither prompted for nor
// counted, so the count the prompts show is of the prompts the user will actually see.
func TestModifierPromptsCountOnlyRowsWithModifiers(t *testing.T) {
	c := check.New(t)
	var steps [][2]int
	uxtest.SwapForTest(t, &promptForTraitModifiers, func(info *modifierPromptInfo, _ []*gurps.TraitModifier) (changed, canceled bool) {
		steps = append(steps, [2]int{info.step, info.steps})
		return false, false
	})
	entity := gurps.NewEntity()
	plain := gurps.NewTrait(entity, nil, false)
	first := gurps.NewTrait(entity, nil, false)
	first.AddModifiers(gurps.NewTraitModifier(entity, nil, false))
	second := gurps.NewTrait(entity, nil, false)
	second.AddModifiers(gurps.NewTraitModifier(entity, nil, false))
	c.True(processModifiers([]*gurps.Trait{plain, first, plain, second}, false))
	c.Equal([][2]int{{1, 2}, {2, 2}}, steps)
}

// TestAltDropOnTraitSwitchesTheDroppedModifierOn verifies that dropping a trait modifier onto a trait row, or an
// equipment modifier onto an equipment row, adds an enabled copy of it without prompting, leaving the row's existing
// modifiers and the dragged modifier as they were.
func TestAltDropOnTraitSwitchesTheDroppedModifierOn(t *testing.T) {
	t.Run("traits", func(t *testing.T) {
		c := check.New(t)
		forbidModifierPrompts(t)
		entity, traits := entityWithNamedTraits("Target Trait")
		target := traits[0]
		existing := gurps.NewTraitModifier(entity, nil, false)
		existing.Name = "Existing"
		existing.SetDisabled(true)
		target.Modifiers = []*gurps.TraitModifier{existing}
		dropped := gurps.NewTraitModifier(entity, nil, false)
		dropped.Name = "Dropped"
		dropped.SetDisabled(true)

		altDrop(tabledProvider(NewTraitsProvider(entity, false)).AltDropSupport(), []int{0}, dropped)

		c.Equal(2, len(target.Modifiers), "the dropped modifier must be added to the target trait")
		c.Equal("Dropped", target.Modifiers[1].Name, "the dropped modifier must be added to the target trait")
		c.False(target.Modifiers[1].IsDisabled(), "the copy must be switched on")
		c.True(target.Modifiers[0].IsDisabled(), "the trait's existing modifiers must be left alone")
		c.True(dropped.IsDisabled(), "the dragged modifier itself must be left alone")
	})

	t.Run("equipment", func(t *testing.T) {
		c := check.New(t)
		forbidModifierPrompts(t)
		entity, equipment := entityWithNamedEquipment("Target Equipment")
		target := equipment[0]
		existing := gurps.NewEquipmentModifier(entity, nil, false)
		existing.Name = "Existing"
		existing.SetDisabled(true)
		target.Modifiers = []*gurps.EquipmentModifier{existing}
		dropped := gurps.NewEquipmentModifier(entity, nil, false)
		dropped.Name = "Dropped"
		dropped.SetDisabled(true)

		altDrop(tabledProvider(NewEquipmentProvider(entity, true, false)).AltDropSupport(), []int{0}, dropped)

		c.Equal(2, len(target.Modifiers), "the dropped modifier must be added to the target equipment")
		c.Equal("Dropped", target.Modifiers[1].Name, "the dropped modifier must be added to the target equipment")
		c.False(target.Modifiers[1].IsDisabled(), "the copy must be switched on")
		c.True(target.Modifiers[0].IsDisabled(), "the item's existing modifiers must be left alone")
		c.True(dropped.IsDisabled(), "the dragged modifier itself must be left alone")
	})
}

// TestAltDropOnSeveralTraitsGivesEachItsOwnCopy verifies that dropping a modifier onto several selected traits or
// equipment items attaches a separate copy to each of them, a container and one of its own descendants included, since
// both were pointed at. A single shared modifier would tie the rows together, so that enabling it on one would enable
// it on all and renaming it would rename it everywhere.
func TestAltDropOnSeveralTraitsGivesEachItsOwnCopy(t *testing.T) {
	t.Run("traits", func(t *testing.T) {
		c := check.New(t)
		forbidModifierPrompts(t)
		entity, traits := entityWithNamedTraits("First Trait", "Second Trait")
		first, second := traits[0], traits[1]
		dropped := gurps.NewTraitModifier(entity, nil, false)
		dropped.Name = "Dropped"

		altDrop(tabledProvider(NewTraitsProvider(entity, false)).AltDropSupport(), []int{0, 1}, dropped)

		c.Equal(1, len(first.Modifiers), "the dropped modifier must be added to the first trait")
		c.Equal(1, len(second.Modifiers), "the dropped modifier must be added to the second trait")
		c.Equal("Dropped", first.Modifiers[0].Name, "the first trait must get the dropped modifier")
		c.Equal("Dropped", second.Modifiers[0].Name, "the second trait must get the dropped modifier")
		c.NotEqual(first.Modifiers[0].ID(), second.Modifiers[0].ID(), "each trait must get a copy of its own")
		c.NotEqual(dropped.ID(), first.Modifiers[0].ID(), "the dragged modifier itself must not be attached")
	})

	t.Run("equipment", func(t *testing.T) {
		c := check.New(t)
		forbidModifierPrompts(t)
		entity, equipment := entityWithNamedEquipment("First Item", "Second Item")
		first, second := equipment[0], equipment[1]
		dropped := gurps.NewEquipmentModifier(entity, nil, false)
		dropped.Name = "Dropped"

		altDrop(tabledProvider(NewEquipmentProvider(entity, true, false)).AltDropSupport(), []int{0, 1}, dropped)

		c.Equal(1, len(first.Modifiers), "the dropped modifier must be added to the first item")
		c.Equal(1, len(second.Modifiers), "the dropped modifier must be added to the second item")
		c.Equal("Dropped", first.Modifiers[0].Name, "the first item must get the dropped modifier")
		c.Equal("Dropped", second.Modifiers[0].Name, "the second item must get the dropped modifier")
		c.NotEqual(first.Modifiers[0].ID(), second.Modifiers[0].ID(), "each item must get a copy of its own")
		c.NotEqual(dropped.ID(), first.Modifiers[0].ID(), "the dragged modifier itself must not be attached")
	})

	t.Run("container and child", func(t *testing.T) {
		c := check.New(t)
		forbidModifierPrompts(t)
		entity := gurps.NewEntity()

		container := gurps.NewTrait(entity, nil, true)
		container.Name = "Container Trait"
		child := gurps.NewTrait(entity, container, false)
		child.Name = "Child Trait"
		container.Children = []*gurps.Trait{child}
		entity.Traits = []*gurps.Trait{container}

		provider := NewTraitsProvider(entity, false)
		table := newProviderTable(provider)
		c.Equal(1, table.LastRowIndex(), "the container's child must be disclosed")

		dropped := gurps.NewTraitModifier(entity, nil, false)
		dropped.Name = "Dropped"
		altDrop(provider.AltDropSupport(), []int{0, 1}, dropped)

		c.Equal(1, len(container.Modifiers), "the container must receive the dropped modifier")
		c.Equal(1, len(child.Modifiers), "the child must receive its own copy of the dropped modifier")
		c.NotEqual(container.Modifiers[0].ID(), child.Modifiers[0].ID(), "each must get a copy of its own")
	})
}

// TestAltDropOnAMissingTraitRowIsANoOp verifies that the trait and equipment drop handlers skip a row index which
// resolves to nothing, and that a drop none of whose indexes resolve reports that it changed nothing, so that no edit
// is recorded.
func TestAltDropOnAMissingTraitRowIsANoOp(t *testing.T) {
	t.Run("traits", func(t *testing.T) {
		c := check.New(t)
		forbidModifierPrompts(t)
		entity, traits := entityWithNamedTraits("Trait")
		trait := traits[0]
		provider := tabledProvider(NewTraitsProvider(entity, false))
		dropped := gurps.NewTraitModifier(entity, nil, false)
		dropped.Name = "Dropped"

		c.False(provider.AltDropSupport().Drop([]int{99}, newDragData(dropped)),
			"a drop with no resolvable target changes nothing, so it must say so")
		c.Equal(0, len(trait.Modifiers), "a drop with no resolvable target must attach nothing")

		c.True(provider.AltDropSupport().Drop([]int{99, 0}, newDragData(dropped)))
		c.Equal(1, len(trait.Modifiers), "an index that resolves to nothing must be skipped rather than stop the drop")
	})

	t.Run("equipment", func(t *testing.T) {
		c := check.New(t)
		forbidModifierPrompts(t)
		entity, equipment := entityWithNamedEquipment("Item")
		item := equipment[0]
		provider := tabledProvider(NewEquipmentProvider(entity, true, false))
		dropped := gurps.NewEquipmentModifier(entity, nil, false)
		dropped.Name = "Dropped"

		c.False(provider.AltDropSupport().Drop([]int{99}, newDragData(dropped)),
			"a drop with no resolvable target changes nothing, so it must say so")
		c.Equal(0, len(item.Modifiers), "a drop with no resolvable target must attach nothing")

		c.True(provider.AltDropSupport().Drop([]int{99, 0}, newDragData(dropped)))
		c.Equal(1, len(item.Modifiers), "an index that resolves to nothing must be skipped rather than stop the drop")
	})
}

// TestAltDropOfTheWrongKindOfModifierIsIgnored verifies that each provider's alternate drop handler only acts on drag
// data holding its own kind of modifier, reporting that nothing changed otherwise.
func TestAltDropOfTheWrongKindOfModifierIsIgnored(t *testing.T) {
	c := check.New(t)
	forbidModifierPrompts(t)
	entity := gurps.NewEntity()
	trait := namedTraits(entity, "Trait")[0]
	item := namedEquipment(entity, "Item")[0]

	traitsProv := tabledProvider(NewTraitsProvider(entity, false))
	equipmentProv := tabledProvider(NewEquipmentProvider(entity, true, false))

	traitMod := gurps.NewTraitModifier(entity, nil, false)
	traitMod.Name = "Trait Modifier"
	traitModData := newDragData(traitMod)
	equipmentMod := gurps.NewEquipmentModifier(entity, nil, false)
	equipmentMod.Name = "Equipment Modifier"
	equipmentModData := newDragData(equipmentMod)

	c.False(traitsProv.AltDropSupport().Drop([]int{0}, equipmentModData),
		"a drop of the wrong kind of modifier changes nothing, so it must say so, or an empty edit would be recorded")
	c.False(equipmentProv.AltDropSupport().Drop([]int{0}, traitModData),
		"a drop of the wrong kind of modifier changes nothing, so it must say so, or an empty edit would be recorded")
	c.Equal(0, len(trait.Modifiers), "equipment modifiers must not be attached to a trait")
	c.Equal(0, len(item.Modifiers), "trait modifiers must not be attached to equipment")

	c.True(traitsProv.AltDropSupport().Drop([]int{0}, traitModData))
	c.True(equipmentProv.AltDropSupport().Drop([]int{0}, equipmentModData))
	c.Equal(1, len(trait.Modifiers), "trait modifiers must still be attached to a trait")
	c.Equal(1, len(item.Modifiers), "equipment modifiers must still be attached to equipment")
}

// The modifier prompt's list has the same least room as the picker's, however few modifiers it holds, and the prompt
// can't be made too small to give it that.
func TestModifierPromptListHoldsTenRows(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	s, n := newKnightSession()
	res := n["res"]
	done := false
	c.True(screen.Post(func() {
		s.chooseModifiers(res)
		done = true
	}))
	screen.Sync()
	dialogWnd, dialog := modalDialog(t, screen, wnd)
	uxtest.CaptureScreen(t, c, screen, "modifier_prompt")
	screen.Do(func() {
		scroll := uxtest.PanelsOfType[*unison.ScrollPanel](dialogWnd.Content())[0]
		layout, ok := scroll.Layout().(*minSizeLayout)
		c.True(ok, "the list's scroll panel holds its least size")
		if !ok {
			return
		}
		minimum := layout.minimum
		view := scroll.ContentView().ContentRect(false).Size
		c.True(view.Height >= minimum.Height && view.Width >= minimum.Width,
			"the list has its least room: wants %v, has %v", minimum, view)
		r := dialogWnd.ContentRect()
		r.Width /= 2
		r.Height /= 2
		dialogWnd.SetContentRect(r)
		dialogWnd.ValidateLayout()
		view = scroll.ContentView().ContentRect(false).Size
		c.True(view.Height >= minimum.Height && view.Width >= minimum.Width,
			"asked to be smaller, the prompt keeps the list's room: wants %v, has %v", minimum, view)
	})
	screen.Click(screen.PanelCenter(dialogButton(t, screen, dialog, unison.ModalResponseCancel)))
	c.True(done, "the prompt has returned")
}

// TestPreconfiguredAsksOnlyAboutUnresolvedChoices verifies that a preconfigured trait is asked about the mandatory
// modifier choices it has left without a pick, and about nothing else: not its other modifiers, not its optional
// choices and not the mandatory choices it has already picked for.
func TestPreconfiguredAsksOnlyAboutUnresolvedChoices(t *testing.T) {
	c := check.New(t)
	prompts := captureModifierPrompts(t)
	plain := gurps.NewTraitModifier(nil, nil, false)
	plain.Name = "Plain"
	made := newTraitModifierChoiceFor(nil, true, []string{"Low", "High"}, "High")
	made.Name = "Made"
	open := newTraitModifierChoiceFor(nil, true, []string{"Hot", "Cold"})
	open.Name = "Open"
	optional := newTraitModifierChoiceFor(nil, false, []string{"Loud", "Quiet"})
	optional.Name = "Optional"
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Blast"
	trait.Modifiers = []*gurps.TraitModifier{plain, made, open, optional}
	trait.Preconfigured = true

	c.True(processModifiers([]*gurps.Trait{trait}, true))
	c.Equal([]modifierPrompt{{title: "Blast", modifiers: []string{"Open"}}}, *prompts,
		"only the unresolved mandatory choice is asked about")

	*prompts = nil
	open.Children[0].SetDisabled(false)
	c.True(processModifiers([]*gurps.Trait{trait}, true))
	c.Equal(0, len(*prompts), "with every mandatory choice made, a preconfigured trait isn't asked at all")

	*prompts = nil
	trait.Preconfigured = false
	c.True(processModifiers([]*gurps.Trait{trait}, true))
	c.Equal([]modifierPrompt{{title: "Blast", modifiers: []string{"Plain", "Made", "Open", "Optional"}}}, *prompts,
		"a trait that isn't preconfigured is asked about everything")

	*prompts = nil
	open.Children[0].SetDisabled(true)
	trait.Preconfigured = true
	c.True(processModifiers([]*gurps.Trait{trait}, false))
	c.Equal(0, len(*prompts), "a preconfigured row headed for a template isn't asked about its choices")
}

// TestModifierSelectionTreatsChoicesByKind verifies the prompt asking which modifiers to enable: a modifier outside a
// choice gets a check box, and each choice, reaching through the groups beneath it but not into a choice within it,
// gets radio buttons for its options, an optional one adding "None"; picking an option or clicking a check box tells
// the prompt. On a sheet a mandatory choice shows whether it has its pick and holds the prompt open until it does,
// unless it has no options to pick from; elsewhere it is offered as an optional one is.
func TestModifierSelectionTreatsChoicesByKind(t *testing.T) {
	c := check.New(t)
	named := func(parent *gurps.TraitModifier, name string) *gurps.TraitModifier {
		m := gurps.NewTraitModifier(nil, parent, false)
		m.Name = name
		m.SetDisabled(true)
		parent.Children = append(parent.Children, m)
		return m
	}
	outer := gurps.NewTraitModifier(nil, nil, true)
	plain := named(outer, "Plain")
	plain.SetDisabled(false)
	choice := gurps.NewTraitModifierChoice(nil, outer)
	outer.Children = append(outer.Children, choice)
	direct := named(choice, "Direct")
	group := gurps.NewTraitModifier(nil, choice, true)
	choice.Children = append(choice.Children, group)
	grouped := named(group, "Grouped")
	inner := gurps.NewTraitModifierChoice(nil, choice)
	choice.Children = append(choice.Children, inner)
	low := named(inner, "Low")
	high := named(inner, "High")
	optional := newTraitModifierChoiceFor(nil, false, []string{"Hot", "Cold"}, "Hot", "Cold")

	s := newModifierSelection([]*gurps.TraitModifier{outer, optional}, true, false)
	c.NotNil(s)
	var changes int
	s.onChange = func() { changes++ }
	c.Equal(1, len(s.boxes), "only the modifier outside every choice gets a check box")
	c.Equal(3, len(s.choices), "each choice, nested or not, gets its own group of radio buttons")
	c.Equal(7, len(uxtest.PanelsOfType[*unison.RadioButton](s.list)), "only the optional choice adds one for None")
	optionsOf := func(choice *choiceRadioGroup) map[gurps.GeneralModifier]*unison.RadioButton {
		options := make(map[gurps.GeneralModifier]*unison.RadioButton)
		for rb, gm := range choice.options {
			options[gm] = rb
		}
		return options
	}
	outerOptions := optionsOf(s.choices[0])
	innerOptions := optionsOf(s.choices[1])
	c.Equal(2, len(outerOptions))
	c.NotNil(outerOptions[direct])
	c.NotNil(outerOptions[grouped], "an option in a group beneath the choice is one of its options")
	c.Equal(2, len(innerOptions), "the nested choice's options are its own")
	c.NotNil(innerOptions[low])
	c.NotNil(innerOptions[high])
	c.False(s.choices[2].mandatory)
	c.True(s.choices[2].group.Selected(optionsOf(s.choices[2])[optional.Children[0]]),
		"the first enabled option starts out picked")
	pills := func() []string {
		return slices.DeleteFunc(labelTitles(s.list), func(title string) bool { return title != "Required" && title != "Picked" })
	}
	c.Equal([]string{"Required", "Required"}, pills())
	c.False(s.complete(), "a mandatory choice with nothing picked must hold the prompt open")

	outerOptions[grouped].Click()
	c.Equal(1, changes, "picking an option tells the prompt")
	c.Equal([]string{"Picked", "Required"}, pills())
	c.False(s.complete(), "the nested choice still has to be made")
	innerOptions[high].Click()
	c.Equal(2, changes)
	c.True(s.complete())
	for cb := range s.boxes {
		cb.Click()
	}
	c.Equal(3, changes, "clicking a check box tells the prompt as well")
	c.True(s.apply())
	c.False(plain.Enabled())
	c.False(direct.Enabled())
	c.True(grouped.Enabled())
	c.False(low.Enabled())
	c.True(high.Enabled())
	c.True(optional.Children[0].Enabled())
	c.False(optional.Children[1].Enabled(), "only one option of a choice may be left on")

	s = newModifierSelection([]*gurps.TraitModifier{newTraitModifierChoiceFor(nil, true, []string{"Low", "High"})}, false, false)
	c.True(s.complete(), "off a sheet a mandatory choice may be left without its pick")
	c.Equal(3, len(uxtest.PanelsOfType[*unison.RadioButton](s.list)), "the options and None")
	c.Nil(s.choices[0].updateStatus)

	s = newModifierSelection([]*gurps.TraitModifier{gurps.NewTraitModifierChoice(nil, nil)}, true, false)
	c.NotNil(s)
	c.True(s.complete(), "a mandatory choice with no options to pick from must not hold the prompt open")
	c.Nil(s.choices[0].updateStatus, "nothing is flagged")
	for _, row := range s.list.Children() {
		layout, ok := row.Layout().(*unison.FlexLayout)
		c.True(ok)
		c.Equal(len(row.Children()), layout.Columns, "each row has a column for each of its parts")
	}
}

// TestModifierSelectionClearsAndPreviews verifies that clearing takes back the choices' picks and unchecks the check
// boxes, that a preview sees the answers without keeping them, and that a name still to be filled in is in italics.
func TestModifierSelectionClearsAndPreviews(t *testing.T) {
	c := check.New(t)
	mandatory := newTraitModifierChoiceFor(nil, true, []string{"Low", "@High@"}, "Low")
	optional := newTraitModifierChoiceFor(nil, false, []string{"Hot", "Cold"}, "Hot")
	plain := gurps.NewTraitModifier(nil, nil, false)
	plain.Name = "Plain"
	mods := []*gurps.TraitModifier{mandatory, optional, plain}
	enabled := func() string {
		var names []string
		gurps.Traverse(func(m *gurps.TraitModifier) bool {
			names = append(names, m.Name)
			return false
		}, true, true, mods...)
		return strings.Join(names, ", ")
	}
	s := newModifierSelection(mods, true, true)
	c.True(s.italics)
	for rb, gm := range s.choices[0].options {
		if gm == mandatory.Children[1] {
			rb.Click()
		}
	}
	c.Equal("@High@, Hot, Plain", s.preview(enabled))
	c.Equal("Low, Hot, Plain", enabled(), "a preview keeps nothing")
	s.clear()
	c.False(s.complete())
	c.Equal("", s.preview(enabled))
	c.False(newModifierSelection(mods, true, false).italics)
}

// TestApplyingAChoiceToLootAsksForItsPick verifies that applying a container of modifiers to equipment on a loot sheet
// asks which of its modifiers to enable, a pick required, just as on a character sheet, since both are sheets.
func TestApplyingAChoiceToLootAsksForItsPick(t *testing.T) {
	c := check.New(t)
	sheet := newTestLootSheet(t)
	eqp := gurps.NewEquipment(sheet.loot, nil, false)
	eqp.Name = "Sword"
	sheet.loot.Equipment = []*gurps.Equipment{eqp}
	sheet.Rebuild(true)
	var asked []bool
	uxtest.SwapForTest(t, &promptForEquipmentModifiers,
		func(info *modifierPromptInfo, _ []*gurps.EquipmentModifier) (changed, canceled bool) {
			asked = append(asked, info.requirePicks)
			return false, false
		})
	choice := newEquipmentModifierChoiceFor(nil, []string{"A"}, "A")
	attachModifierClones("", []*unison.Table[*Node[*gurps.Equipment]]{sheet.Equipment.Table}, sheet.loot,
		[]*gurps.Equipment{eqp}, []*gurps.EquipmentModifier{choice}, gurps.LibraryFile{})
	c.Equal([]bool{true}, asked, "the loot sheet asks, requiring the pick")
}

// TestPromptRequiresPicksOnlyForSheets verifies that rows headed for a template are asked about their modifiers
// without a pick being required, and rows headed for a loot sheet with one.
func TestPromptRequiresPicksOnlyForSheets(t *testing.T) {
	c := check.New(t)
	var traitAsked, equipmentAsked []bool
	uxtest.SwapForTest(t, &promptForTraitModifiers,
		func(info *modifierPromptInfo, _ []*gurps.TraitModifier) (changed, canceled bool) {
			traitAsked = append(traitAsked, info.requirePicks)
			return false, false
		})
	uxtest.SwapForTest(t, &promptForEquipmentModifiers,
		func(info *modifierPromptInfo, _ []*gurps.EquipmentModifier) (changed, canceled bool) {
			equipmentAsked = append(equipmentAsked, info.requirePicks)
			return false, false
		})
	trait := gurps.NewTrait(nil, nil, false)
	trait.Modifiers = []*gurps.TraitModifier{newTraitModifierChoiceFor(nil, true, []string{"A"})}
	template := newTestTemplateWithTraits()
	toTemplate := &applyPart[*gurps.Trait]{table: template.Traits.Table, rows: []*gurps.Trait{trait}}
	c.True(toTemplate.promptForModifiers(promptOperation{}, &modifierPromptCount{total: 1}, nil))
	c.Equal([]bool{false}, traitAsked, "a template doesn't require the pick")

	loot := newTestLootSheet(t)
	equipment := gurps.NewEquipment(nil, nil, false)
	equipment.Modifiers = []*gurps.EquipmentModifier{gurps.NewEquipmentModifierChoice(nil, nil)}
	toLoot := &applyPart[*gurps.Equipment]{table: loot.Equipment.Table, rows: []*gurps.Equipment{equipment}}
	c.True(toLoot.promptForModifiers(promptOperation{}, &modifierPromptCount{total: 1}, nil))
	c.Equal([]bool{true}, equipmentAsked, "a loot sheet requires it")
}

// newPickTestModifiers returns a mandatory choice between Low and High with no pick made, and a modifier that is off.
func newPickTestModifiers() (choice, plain *gurps.TraitModifier) {
	plain = gurps.NewTraitModifier(nil, nil, false)
	plain.Name = "Plain"
	plain.SetDisabled(true)
	return newTraitModifierChoiceFor(nil, true, []string{"Low", "High"}), plain
}

// TestModifierPromptPicksRandomly verifies that the modifier prompt's random buttons keep what was set in it and give a
// mandatory choice without a pick one, that only a prompt asking for a pick has them, that Auto Pick is available
// only while a pick is still to be made there, and that one put up from the picker doesn't offer to make the picks
// still to come.
func TestModifierPromptPicksRandomly(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	// show puts up the prompt for the modifiers and returns its window, along with what tells whether it has returned.
	show := func(info *modifierPromptInfo, modifiers ...*gurps.TraitModifier) (dialogWnd *unison.Window, returned func() bool) {
		done := false
		c.True(screen.Post(func() {
			showModifiersDialog(info, modifiers)
			done = true
		}))
		screen.Sync()
		dialogWnd, _ = modalDialog(t, screen, wnd)
		return dialogWnd, func() bool { return done }
	}
	button := func(dialogWnd *unison.Window, title string) (b *unison.Button) {
		screen.Do(func() { b = buttonWithTitle(dialogWnd.Content(), title) })
		return b
	}
	enabled := func(dialogWnd *unison.Window, title string) (on bool) {
		screen.Do(func() {
			b := buttonWithTitle(dialogWnd.Content(), title)
			on = b != nil && b.Enabled()
		})
		return on
	}

	choice, plain := newPickTestModifiers()
	info := &modifierPromptInfo{name: "Blast", requirePicks: true, random: &randomPicks{rnd: firstPick{}}}
	dialogWnd, returned := show(info, choice, plain)
	var boxes []*unison.CheckBox
	screen.Do(func() { boxes = uxtest.PanelsOfType[*unison.CheckBox](dialogWnd.Content()) })
	c.Equal(1, len(boxes))
	screen.Click(screen.PanelCenter(boxes[0]))
	screen.Click(screen.PanelCenter(button(dialogWnd, "Auto Pick")))
	c.True(returned())
	c.True(choice.Children[0].Enabled(), "the choice was given its pick")
	c.False(choice.Children[1].Enabled())
	c.True(plain.Enabled(), "what was set in the prompt is kept")
	c.False(info.random.rest, "only this prompt's pick was asked for")

	choice, plain = newPickTestModifiers()
	dialogWnd, returned = show(info, choice, plain)
	screen.Click(screen.PanelCenter(button(dialogWnd, "Auto Pick All")))
	c.True(returned())
	c.True(gurps.ModifierChoiceIsResolved(choice))
	c.False(plain.Enabled(), "a modifier outside a choice is left as it was")
	c.True(info.random.rest)

	choice, plain = newPickTestModifiers()
	dialogWnd, returned = show(info, choice, plain)
	var radios []*unison.RadioButton
	screen.Do(func() { radios = uxtest.PanelsOfType[*unison.RadioButton](dialogWnd.Content()) })
	c.Equal(2, len(radios))
	screen.Click(screen.PanelCenter(radios[1]))
	c.False(enabled(dialogWnd, "Auto Pick"), "a pick made by hand leaves none to make here")
	c.True(enabled(dialogWnd, "Auto Pick All"), "but still those of the prompts to come")
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.True(returned())

	optional := newTraitModifierChoiceFor(nil, false, []string{"Loud", "Quiet"})
	made := newTraitModifierChoiceFor(nil, true, []string{"Low", "High"}, "High")
	dialogWnd, returned = show(info, optional, made, plain)
	c.Nil(button(dialogWnd, "Auto Pick"), "a prompt that asks for no pick has none to make")
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.True(returned())

	choice, plain = newPickTestModifiers()
	info.early = &earlyModifierPrompt{cost: func() string { return "" }}
	dialogWnd, returned = show(info, choice, plain)
	c.NotNil(button(dialogWnd, "Auto Pick"))
	c.Nil(button(dialogWnd, "Auto Pick All"), "a prompt put up from the picker is one of its picks")
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.True(returned())

	made = newTraitModifierChoiceFor(nil, true, []string{"Low", "High"}, "High")
	dialogWnd, returned = show(info, made)
	c.False(enabled(dialogWnd, "Auto Pick"), "a prompt put up from the picker with its pick made has none to make")
	screen.Click(screen.PanelCenter(button(dialogWnd, "Clear Selections")))
	ok := enabled(dialogWnd, "Auto Pick")
	c.True(ok, "clearing the selections leaves the pick to make")
	if !ok {
		screen.KeyPress(unison.KeyEscape, mod.None)
		return
	}
	screen.Click(screen.PanelCenter(button(dialogWnd, "Auto Pick")))
	c.True(returned())
	c.True(made.Children[0].Enabled(), "the cleared choice was given a pick")
	c.False(made.Children[1].Enabled())
}

// TestRemainingModifierPromptsAreAnsweredRandomly verifies that once a prompt asks for the picks still to come to be
// made at random, a modifier prompt that asks for one is answered rather than put up, with only its unmade picks
// changed, and taken out of the count, while one that asks for none, as with a mandatory choice already made, is still
// put up, as all are when picks aren't required.
func TestRemainingModifierPromptsAreAnsweredRandomly(t *testing.T) {
	c := check.New(t)
	newRows := func() (rows []*gurps.Trait, lastChoice, lastPlain *gurps.TraitModifier) {
		for _, name := range []string{"First", "Optional", "Made", "Last"} {
			trait := gurps.NewTrait(nil, nil, false)
			trait.Name = name
			lastChoice, lastPlain = newPickTestModifiers()
			switch name {
			case "Optional":
				trait.Modifiers = []*gurps.TraitModifier{lastPlain}
			case "Made":
				lastChoice.Children[1].SetDisabled(false)
				fallthrough
			default:
				trait.Modifiers = []*gurps.TraitModifier{lastChoice, lastPlain}
			}
			rows = append(rows, trait)
		}
		return rows, lastChoice, lastPlain
	}
	var asked []string
	// Each prompt presses Auto Pick All.
	uxtest.SwapForTest(t, &promptForTraitModifiers, func(info *modifierPromptInfo, mods []*gurps.TraitModifier) (changed, canceled bool) {
		asked = append(asked, fmt.Sprintf("%s (%d of %d)", info.name, info.step, info.steps))
		info.random.rest = true
		return gurps.PickUnresolvedModifierChoices(info.random.intn, mods...), false
	})
	rows, choice, plain := newRows()
	parts := newApplyParts(applyPart[*gurps.Trait]{table: newTestSheetForTemplate(t).Traits.Table, rows: rows})
	parts.random.rnd = firstPick{}
	c.True(parts.promptForModifiers(promptOperation{}))
	c.Equal([]string{"First (1 of 4)", "Optional (2 of 3)", "Made (3 of 3)"}, asked)
	c.True(choice.Children[0].Enabled(), "the pick of the prompt not put up was made")
	c.False(choice.Children[1].Enabled())
	c.False(plain.Enabled(), "and nothing else about it changed")

	asked = nil
	rows, _, _ = newRows()
	parts = newApplyParts(applyPart[*gurps.Trait]{table: newTestTemplateWithTraits().Traits.Table, rows: rows})
	parts.random.rest = true
	c.True(parts.promptForModifiers(promptOperation{}))
	c.Equal([]string{"First (1 of 4)", "Optional (2 of 4)", "Made (3 of 4)", "Last (4 of 4)"}, asked,
		"where no pick is required there is none to make")
}
