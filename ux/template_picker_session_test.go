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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// newKnightSession returns a session over the class advantages of the DeepNesting mock-up, and its rows by name.
func newKnightSession() (s *pickerSession[*gurps.Trait], n map[string]*gurps.Trait) {
	n = make(map[string]*gurps.Trait)
	leaf := func(name string, points int) *gurps.Trait {
		trait := gurps.NewTrait(nil, nil, false)
		trait.Name = name
		trait.BasePoints = fxp.FromInteger(points)
		n[name] = trait
		return trait
	}
	pick := func(name string, pt picker.Type, qualifier int, children ...*gurps.Trait) *gurps.Trait {
		trait := gurps.NewTrait(nil, nil, true)
		trait.Name = name
		trait.TemplatePicker.Type = pt
		trait.TemplatePicker.Qualifier.Compare = criteria.EqualsNumber
		trait.TemplatePicker.Qualifier.Qualifier = fxp.FromInteger(qualifier)
		trait.Children = children
		SetParents(children, trait)
		n[name] = trait
		return trait
	}
	// Each option's name is its cost.
	mods := func(name string, choices ...[]string) *gurps.Trait {
		trait := leaf(name, 0)
		for _, costs := range choices {
			choice := newTraitModifierChoiceFor(nil, true, costs)
			for _, option := range choice.Children {
				option.CostAdj = option.Name
			}
			trait.AddModifiers(choice)
		}
		return trait
	}
	resPart := mods("resPart", []string{"+30", "+15", "+10", "+5"}, []string{"x1", "x0.5"})
	resPart.Preconfigured = true
	resPart.Modifiers[0].Children[2].SetEnabled(true)
	root := pick("root", picker.Points, 60,
		leaf("ea", 25), leaf("ep", 5),
		pick("fit", picker.Count, 1, leaf("fit1", 5), leaf("fit2", 15)),
		pick("order", picker.Count, 1,
			pick("rose", picker.Points, 20,
				pick("graces", picker.Count, 1, leaf("voice", 10), leaf("charisma", 5)),
				mods("res", []string{"+30", "+15", "+10", "+5"}, []string{"x1", "x0.5"}),
				leaf("honest", 1)),
			pick("lion", picker.Points, 25,
				pick("honors", picker.Count, 1, leaf("cr", 15), leaf("hpt", 10), leaf("htk", 4)),
				mods("wm", []string{"+20", "+25", "+30", "+35", "+40", "+45"}),
				leaf("fear2", 4)),
			leaf("tower", 5)),
		leaf("luck", 15), leaf("shield", 15), resPart)
	return newPickerSession(promptOperation{}, []*gurps.Trait{root}, true), n
}

func choose(s *pickerSession[*gurps.Trait], n map[string]*gurps.Trait, names ...string) {
	for _, name := range names {
		s.chosen[n[name]] = true
	}
}

// TestPickerSessionExpectsByTheRules verifies that a choice is expected to cost what its rules say, whatever is
// picked below it, and that what is picked rolls up to the top.
func TestPickerSessionExpectsByTheRules(t *testing.T) {
	c := check.New(t)
	s, n := newKnightSession()
	choose(s, n, "ea", "ep", "fit", "order")
	c.Equal("40~70 / 60", s.pillText(n["root"]), "an unpicked choice counts as its rules expect")
	c.Equal(pickerOpen, s.state(n["root"]))

	choose(s, n, "lion", "honors", "cr", "wm", "fear2")
	c.Equal("39~64 / 25", s.pillText(n["lion"]))
	c.Equal(pickerError, s.state(n["lion"]))
	c.Equal("5~25", s.expected[pickerMeasureKey[*gurps.Trait]{n["order"], picker.Points}].String())
	c.Equal("39~64", s.actual(n["order"], picker.Points).String(), "the overage shows on the line it is on")
	c.Equal("1 / 1", s.pillText(n["order"]))
	c.Equal(pickerWarning, s.state(n["order"]), "a child's error is a warning to its parent")
	c.Equal("74~109 / 60", s.pillText(n["root"]))
	c.Equal(pickerError, s.state(n["root"]))

	n["wm"].Modifiers[0].Children[4].SetEnabled(true)
	c.Equal("39~64", s.actual(n["lion"], picker.Points).String(), "a pick is only a default until answered")
	s.modsAnswered[n["wm"]] = true
	c.Equal("59", s.actual(n["lion"], picker.Points).String())
	c.Equal("94~104 / 60", s.pillText(n["root"]))
}

// TestPickerSessionStates verifies each of the four states a choice can be in.
func TestPickerSessionStates(t *testing.T) {
	c := check.New(t)
	s, n := newKnightSession()
	choose(s, n, "ea", "ep")
	c.Equal("30 / 60", s.pillText(n["root"]))
	c.Equal(pickerError, s.state(n["root"]))

	s, n = newKnightSession()
	root := n["root"]
	choose(s, n, "ea", "fit", "fit2", "order", "tower", "luck")
	c.Equal("60 / 60", s.pillText(root))
	c.Equal(pickerOK, s.state(root))
	root.TemplatePicker.Qualifier.Compare = criteria.AtLeastNumber
	choose(s, n, "ep")
	c.Equal("65 / ≥60", s.pillText(root))
	c.Equal(pickerOK, s.state(root))

	s, n = newKnightSession()
	choose(s, n, "ea", "fit", "fit1", "fit2", "luck")
	c.Equal("2 / 1", s.pillText(n["fit"]))
	c.Equal(" [15 points]", s.cost(n["fit2"], picker.Points).text, "a count choice shows what its options cost too")
	c.Equal("20", s.actual(n["fit"], picker.Points).String(), "a count overridden past its number costs every pick")
	c.Equal("60 / 60", s.pillText(n["root"]))
	c.Equal(pickerWarning, s.state(n["root"]))

	s, n = newKnightSession()
	choose(s, n, "ea", "order", "rose", "luck")
	c.Equal("60 / 60", s.pillText(n["root"]), "an unpicked points choice counts as its target")
	c.Equal(pickerOpen, s.state(n["order"]), "but is still to be answered")
	c.Equal(pickerOpen, s.state(n["root"]))

	c.False(s.resolved(n["resPart"]), "a preconfigured row with a mandatory pick missing")
	n["resPart"].Modifiers[1].Children[0].SetEnabled(true)
	c.True(s.resolved(n["resPart"]))
	n["wm"].Modifiers[0].Children[0].SetEnabled(true)
	c.False(s.resolved(n["wm"]))
	s.modsAnswered[n["wm"]] = true
	c.True(s.resolved(n["wm"]))
}

// TestPickerSessionPillTextByValue verifies that a choice made by value reads as money.
func TestPickerSessionPillTextByValue(t *testing.T) {
	c := check.New(t)
	choice := gurps.NewEquipmentChoiceContainer(nil, nil)
	choice.TemplatePicker.Type = picker.Value
	choice.TemplatePicker.Qualifier.Qualifier = fxp.FromInteger(50)
	for _, value := range []string{"30", "20"} {
		option := gurps.NewEquipment(nil, choice, false)
		option.BaseValue = value
		choice.Children = append(choice.Children, option)
	}
	s := newPickerSession(promptOperation{}, []*gurps.Equipment{choice}, true)
	s.chosen[choice.Children[0]] = true
	c.Equal("$30 / $50", s.pillText(choice))
	c.Equal(pickerError, s.ownState(choice), "a settled total short of its target can't meet it")
}

// TestPickerSessionProcessesRows verifies that choices already answered are applied without their dialog, and that the
// rest are asked once what holds them has been.
func TestPickerSessionProcessesRows(t *testing.T) {
	c := check.New(t)
	s, n := newKnightSession()
	var asked []string
	s.runPicker = func(row *gurps.Trait, _ int) int {
		asked = append(asked, row.Name)
		switch row.Name {
		case "root":
			choose(s, n, "ea", "fit", "order")
		case "order":
			choose(s, n, "tower")
		}
		return unison.ModalResponseOK
	}
	s.pickerAnswered[n["fit"]] = true
	choose(s, n, "fit2")
	rows, abort := s.processRows([]*gurps.Trait{n["root"]})
	c.False(abort)
	c.Equal([]string{"root", "order"}, asked)
	names := make([]string, 0, len(rows))
	for _, row := range rows {
		names = append(names, row.Name)
		c.Nil(row.Parent())
	}
	c.Equal([]string{"ea", "fit2", "tower"}, names)

	s, n = newKnightSession()
	s.runPicker = func(*gurps.Trait, int) int { return unison.ModalResponseCancel }
	_, abort = s.processRows([]*gurps.Trait{n["root"]})
	c.True(abort)
}

// TestPickerSessionChoosesModifiers verifies that modifiers chosen from the picker count as answered and check the
// row, that canceling puts everything back, and that the later prompt asks only about what is still open.
func TestPickerSessionChoosesModifiers(t *testing.T) {
	c := check.New(t)
	var answer func(mods []*gurps.TraitModifier) (canceled bool)
	var asked []string
	swapForTest(t, &promptForTraitModifiers, func(info *modifierPromptInfo, mods []*gurps.TraitModifier) (changed, canceled bool) {
		text := fmt.Sprintf("%s (%d of %d) %d", info.name, info.step, info.steps, len(mods))
		if info.early != nil {
			text += " " + info.early.cost()
		}
		asked = append(asked, text)
		return false, answer(mods)
	})
	s, n := newKnightSession()
	wm, res := n["wm"], n["res"]
	answer = func(mods []*gurps.TraitModifier) bool {
		mods[0].Children[1].SetEnabled(true)
		return false
	}
	s.chooseModifiers(wm)
	c.True(s.chosen[wm], "confirming checks the row")
	c.True(s.resolved(wm))
	c.Equal("25", s.actual(wm, picker.Points).String())

	s.chosen[wm] = false
	answer = func(mods []*gurps.TraitModifier) bool {
		mods[0].Children[1].SetEnabled(false)
		mods[0].Children[2].SetEnabled(true)
		return true
	}
	s.chooseModifiers(wm)
	c.False(s.chosen[wm], "canceling leaves the row as it was")
	c.True(wm.Modifiers[0].Children[1].Enabled(), "and puts the picks back")
	c.False(wm.Modifiers[0].Children[2].Enabled())
	c.True(s.modsAnswered[wm])

	// A partial answer narrows the row, which is asked about the rest later.
	answer = func(mods []*gurps.TraitModifier) bool {
		mods[0].Children[3].SetEnabled(true)
		return false
	}
	s.chooseModifiers(res)
	c.False(s.resolved(res))
	c.Equal([]string{"wm (1 of 1) 1 wm: 20~45 points", "wm (1 of 1) 1 wm: 25 points", "res (1 of 1) 2 res: 3~30 points"},
		asked)
	c.Equal("3~5", s.actual(res, picker.Points).String())
	rows := []*gurps.Trait{wm, res, n["resPart"]}
	targets := modifierTargets(rows, true, s.modsAnswered)
	c.Equal([]*gurps.Trait{res, n["resPart"]}, targets, "an answered row isn't asked about again")
	asked = nil
	answer = func([]*gurps.TraitModifier) bool { return false }
	c.True(promptForModifierTargets(promptOperation{}, targets, 0, len(targets), true, s.modsAnswered))
	c.Equal([]string{"res (1 of 2) 1", "resPart (2 of 2) 1"}, asked, "only the choice left open is asked about")
}

// TestPickerSessionPreconfiguredModifiers verifies that a preconfigured row gets the modifier prompt from the picker
// only while a mandatory pick is missing, and then only for that choice.
func TestPickerSessionPreconfiguredModifiers(t *testing.T) {
	c := check.New(t)
	var early *earlyModifierPrompt
	var asked [][]*gurps.TraitModifier
	swapForTest(t, &promptForTraitModifiers, func(info *modifierPromptInfo, mods []*gurps.TraitModifier) (changed, canceled bool) {
		early = info.early
		asked = append(asked, mods)
		return false, false
	})
	s, n := newKnightSession()
	resPart := n["resPart"]
	c.Equal([]*gurps.Trait{resPart}, s.modTargets(resPart))
	s.chooseModifiers(resPart)
	c.NotNil(early)
	c.True(early.preconfigured)
	c.Equal([][]*gurps.TraitModifier{{resPart.Modifiers[1]}}, asked)

	resPart.Modifiers[1].Children[0].SetEnabled(true)
	s = newPickerSession(promptOperation{}, []*gurps.Trait{n["root"]}, true)
	c.Equal(0, len(s.modTargets(resPart)), "fully picked, it has nothing to ask")
	c.True(s.resolved(resPart))
}

// TestPickerSessionInheritedModifierChoice verifies that a choice inherited from a container above the choice counts
// as it stands until the later prompt, which still asks about it.
func TestPickerSessionInheritedModifierChoice(t *testing.T) {
	c := check.New(t)
	outer := gurps.NewTrait(nil, nil, true)
	outer.AddModifiers(newTraitModifierChoiceFor(nil, true, []string{"+1", "+3"}))
	choice := gurps.NewTrait(nil, outer, true)
	choice.TemplatePicker.Type = picker.Count
	choice.TemplatePicker.Qualifier.Compare = criteria.EqualsNumber
	choice.TemplatePicker.Qualifier.Qualifier = fxp.One
	outer.Children = []*gurps.Trait{choice}
	option := gurps.NewTrait(nil, choice, false)
	option.BasePoints = fxp.FromInteger(10)
	choice.Children = []*gurps.Trait{option}
	s := newPickerSession(promptOperation{}, []*gurps.Trait{outer}, true)
	s.chosen[option] = true
	c.True(s.resolved(option))
	c.Equal("10", s.actual(option, picker.Points).String())
	c.Equal(pickerOK, s.state(choice))
	c.Equal(0, len(s.modTargets(option)))
	c.Equal([]*gurps.Trait{outer}, modifierTargets([]*gurps.Trait{outer}, true, s.modsAnswered))
}

// TestPickerSessionPlainContainerModifiers verifies that a plain container picked from a choice is open while a row
// below it has a mandatory modifier choice to make, and that its prompt leaves out a choice below it.
func TestPickerSessionPlainContainerModifiers(t *testing.T) {
	c := check.New(t)
	root := gurps.NewTrait(nil, nil, true)
	root.TemplatePicker.Type = picker.Count
	root.TemplatePicker.Qualifier.Qualifier = fxp.One
	pack := gurps.NewTrait(nil, root, true)
	root.Children = []*gurps.Trait{pack}
	open := gurps.NewTrait(nil, pack, false)
	open.AddModifiers(newTraitModifierChoiceFor(nil, true, []string{"+1", "+3"}))
	sub := gurps.NewTrait(nil, pack, true)
	sub.TemplatePicker.Type = picker.Count
	sub.TemplatePicker.Qualifier.Qualifier = fxp.One
	pack.Children = []*gurps.Trait{open, sub}
	later := gurps.NewTrait(nil, sub, false)
	later.AddModifiers(newTraitModifierChoiceFor(nil, true, []string{"+1", "+3"}))
	sub.Children = []*gurps.Trait{later}
	s := newPickerSession(promptOperation{}, []*gurps.Trait{root}, true)
	s.chosen[pack] = true
	c.Equal([]*gurps.Trait{open}, s.modTargets(pack), "a choice below is answered in its own dialog")
	c.False(s.resolved(pack), "a mandatory pick is still to be made below it")
	c.Equal(pickerOpen, s.state(root))
}

// TestPickerSessionPlainContainerHoldingAChoice verifies that a plain container picked from a choice costs what the
// picks of a choice inside it come to, is open until that choice is made, and can have it made from its row.
func TestPickerSessionPlainContainerHoldingAChoice(t *testing.T) {
	c := check.New(t)
	newChoice := func(parent *gurps.Trait, pt picker.Type, qualifier int) *gurps.Trait {
		choice := gurps.NewTrait(nil, parent, true)
		choice.TemplatePicker.Type = pt
		choice.TemplatePicker.Qualifier.Compare = criteria.EqualsNumber
		choice.TemplatePicker.Qualifier.Qualifier = fxp.FromInteger(qualifier)
		return choice
	}
	newLeaf := func(parent *gurps.Trait, points int) *gurps.Trait {
		leaf := gurps.NewTrait(nil, parent, false)
		leaf.BasePoints = fxp.FromInteger(points)
		parent.Children = append(parent.Children, leaf)
		return leaf
	}
	root := newChoice(nil, picker.Points, 20)
	pack := gurps.NewTrait(nil, root, true)
	sub := newChoice(pack, picker.Count, 1)
	pack.Children = []*gurps.Trait{sub}
	newLeaf(sub, 5)
	ten := newLeaf(sub, 10)
	root.Children = append(root.Children, pack)
	x := newLeaf(root, 10)
	s := newPickerSession(promptOperation{}, []*gurps.Trait{root}, true)
	s.chosen[pack], s.chosen[x] = true, true
	c.Equal("15~20 / 20", s.pillText(root))
	c.False(s.resolved(pack), "the choice inside it is still to be made")
	var asked []*gurps.Trait
	s.runPicker = func(row *gurps.Trait, _ int) int {
		asked = append(asked, row)
		s.chosen[ten] = true
		return unison.ModalResponseOK
	}
	s.chosen[pack] = false
	s.chooseWithin(pack, 0)
	c.Equal([]*gurps.Trait{sub}, asked)
	c.True(s.chosen[pack], "confirming checks the row")
	c.True(s.resolved(pack))
	c.Equal("20 / 20", s.pillText(root))
	c.Equal(pickerOK, s.state(root))
}

// TestPickerSessionChoosesEquipmentModifiers verifies that equipment modifiers can be chosen from the picker too, the
// prompt showing the value and weight.
func TestPickerSessionChoosesEquipmentModifiers(t *testing.T) {
	c := check.New(t)
	var cost string
	swapForTest(t, &promptForEquipmentModifiers, func(info *modifierPromptInfo, mods []*gurps.EquipmentModifier) (changed, canceled bool) {
		cost = info.early.cost()
		mods[0].Children[1].SetEnabled(true)
		return false, false
	})
	choice := gurps.NewEquipmentChoiceContainer(nil, nil)
	choice.TemplatePicker.Type = picker.Value
	sword, _ := newEditorEquipmentWithChoice([2]string{"+50", "+1 lb"}, [2]string{"+100", "+2 lb"})
	sword.Name = "Sword"
	sword.SetParent(choice)
	choice.Children = []*gurps.Equipment{sword}
	s := newPickerSession(promptOperation{}, []*gurps.Equipment{choice}, true)
	c.Equal("$150~200", formatPickerTotal(sword, picker.Value, s.actual(sword, picker.Value)))
	s.chooseModifiers(sword)
	c.Equal("Sword: $150~200, 3~4 lb", cost)
	c.True(s.chosen[sword])
	c.Equal("$200", formatPickerTotal(sword, picker.Value, s.actual(sword, picker.Value)))
}

// TestPickerStatePill verifies that each state has its own look, a warning's text being dark on yellow.
func TestPickerStatePill(t *testing.T) {
	c := check.New(t)
	pill, update := newPickerStatePill(0)
	for state, img := range map[pickerState]*unison.SVG{
		pickerOK: unison.CheckmarkSVG, pickerOpen: unison.CircledQuestionSVG,
		pickerWarning: unison.TriangleExclamationSVG, pickerError: svg.Not,
	} {
		update(state, "1 / 1")
		drawable, ok := pill.Drawable.(*unison.DrawableSVG)
		c.True(ok)
		c.Equal(img, drawable.SVG)
	}
	update(pickerWarning, "1 / 1")
	c.Equal(unison.OnLight, pill.OnBackgroundInk)
}

// TestPickerSessionChoosesPicks verifies that a choice picked from another can be answered from its row, however deep,
// and is then applied without being asked again, while those left unanswered are still asked.
func TestPickerSessionChoosesPicks(t *testing.T) {
	c := check.New(t)
	s, n := newKnightSession()
	var asked []string
	answers := map[string]func(depth int) int{
		"order": func(depth int) int {
			choose(s, n, "rose")
			s.choosePicks(n["rose"], depth)
			return unison.ModalResponseOK
		},
		"rose": func(depth int) int {
			choose(s, n, "honest")
			s.choosePicks(n["graces"], depth)
			return unison.ModalResponseUserBase
		},
		"graces": func(int) int {
			choose(s, n, "voice")
			return unison.ModalResponseOK
		},
		"root": func(int) int {
			choose(s, n, "ea", "fit", "order")
			return unison.ModalResponseOK
		},
		"fit": func(int) int {
			choose(s, n, "fit1")
			return unison.ModalResponseOK
		},
	}
	s.runPicker = func(row *gurps.Trait, depth int) int {
		asked = append(asked, fmt.Sprintf("%s %d", row.Name, depth))
		return answers[row.Name](depth)
	}
	s.choosePicks(n["order"], 0)
	c.Equal([]string{"order 1", "rose 2", "graces 3"}, asked)
	for _, name := range []string{"order", "rose", "graces"} {
		c.True(s.chosen[n[name]], "confirming checks the row")
		c.True(s.pickerAnswered[n[name]])
	}
	c.Equal("11 / 20", s.pillText(n["rose"]), "Override keeps picks that miss the rule")

	asked = nil
	rows, abort := s.processRows([]*gurps.Trait{n["root"]})
	c.False(abort)
	c.Equal([]string{"root 0", "fit 0"}, asked, "only the choice left unanswered is asked")
	names := make([]string, 0, len(rows))
	for _, row := range rows {
		names = append(names, row.Name)
	}
	c.Equal([]string{"ea", "fit1", "voice", "honest"}, names)
}

// TestPickerSessionChoosesPicksBacksOut verifies that canceling a choice put up from its row puts back everything
// changed under it, and that Override with nothing picked leaves it to be asked later.
func TestPickerSessionChoosesPicksBacksOut(t *testing.T) {
	c := check.New(t)
	swapForTest(t, &promptForTraitModifiers, func(_ *modifierPromptInfo, mods []*gurps.TraitModifier) (changed, canceled bool) {
		mods[0].Children[4].SetEnabled(true)
		return false, false
	})
	s, n := newKnightSession()
	wm := n["wm"]
	s.runPicker = func(row *gurps.Trait, depth int) int {
		switch row.Name {
		case "order":
			choose(s, n, "lion")
			s.choosePicks(n["lion"], depth)
			return unison.ModalResponseCancel
		case "lion":
			s.chooseModifiers(wm)
			n["fear2"].Levels = fxp.Two
			choose(s, n, "honors", "cr")
			s.pickerAnswered[n["honors"]] = true
		}
		return unison.ModalResponseOK
	}
	s.choosePicks(n["order"], 0)
	for _, name := range []string{"order", "lion", "honors", "cr", "wm"} {
		c.False(s.chosen[n[name]], name)
	}
	c.Equal(0, len(s.pickerAnswered))
	c.False(s.modsAnswered[wm])
	c.False(wm.Modifiers[0].Children[4].Enabled())
	c.Equal(fxp.Int(0), n["fear2"].Levels)

	s.runPicker = func(*gurps.Trait, int) int { return unison.ModalResponseUserBase }
	s.choosePicks(n["fit"], 0)
	c.False(s.chosen[n["fit"]], "Override with nothing picked backs out")
	c.False(s.pickerAnswered[n["fit"]])

	// Taking back every pick of an answered choice leaves it to be asked later too.
	s, n = newKnightSession()
	fit := n["fit"]
	var asked []string
	answer := func(row *gurps.Trait, _ int) int {
		asked = append(asked, row.Name)
		switch row.Name {
		case "root":
			choose(s, n, "ea", "fit")
		case "fit":
			choose(s, n, "fit1")
		}
		return unison.ModalResponseOK
	}
	s.runPicker = answer
	s.choosePicks(fit, 0)
	c.True(s.pickerAnswered[fit])
	s.runPicker = func(*gurps.Trait, int) int {
		s.clear(fit)
		return unison.ModalResponseUserBase
	}
	s.choosePicks(fit, 0)
	c.True(s.chosen[fit], "the row stays as it was")
	c.False(s.pickerAnswered[fit])
	s.chosen[fit] = false
	s.runPicker, asked = answer, nil
	_, abort := s.processRows([]*gurps.Trait{n["root"]})
	c.False(abort)
	c.Equal([]string{"root", "fit"}, asked)
}

// TestPickerSessionClear verifies that clearing a choice takes back only its own picks.
func TestPickerSessionClear(t *testing.T) {
	c := check.New(t)
	s, n := newKnightSession()
	choose(s, n, "rose", "graces", "voice", "honest")
	s.pickerAnswered[n["graces"]] = true
	s.clear(n["rose"])
	c.False(s.chosen[n["graces"]])
	c.False(s.chosen[n["honest"]])
	c.True(s.chosen[n["rose"]])
	c.True(s.chosen[n["voice"]], "the answers below are kept")
	c.True(s.pickerAnswered[n["graces"]])
}

// TestPickerSessionRowText verifies what follows each row's name, what it costs, and the line under the list, with the
// state that colors them and the tooltip saying why.
func TestPickerSessionRowText(t *testing.T) {
	c := check.New(t)
	type session = pickerSession[*gurps.Trait]
	detail := (*session).detail
	cost := func(s *session, row *gurps.Trait) pickerText { return s.cost(row, picker.Points) }
	hint := (*session).hint
	start := []string{"ea", "ep", "fit", "order"}
	over := []string{"ea", "ep", "fit", "order", "lion", "honors", "cr", "wm", "fear2"}
	misses := []string{"fit1", "fit2", "rose", "honest", "graces", "voice"}
	for _, tc := range []struct {
		picks []string
		row   string
		text  func(*session, *gurps.Trait) pickerText
		want  pickerText
	}{
		{start, "fit", detail, pickerText{text: " (pick 1)"}},
		{start, "fit", cost, pickerText{text: " [5~15 points]", state: pickerOpen}},
		{start, "ea", detail, pickerText{}},
		{start, "ea", cost, pickerText{text: " [25 points]"}},
		{start, "resPart", detail, pickerText{text: " [+10]", tip: "Preconfigured"}},
		{start, "root", hint, pickerText{text: "2 picks still depend on choices below. Choose them now to fix the " +
			"cost, or later when the template is applied.", state: pickerOpen}},
		{over, "lion", detail, pickerText{
			text: " (pick 25 points worth)", state: pickerError,
			tip: "The picks come to 39~64 points, but this asks for 25 points.",
		}},
		{over, "lion", cost, pickerText{text: " [39~64 points]", tip: "Over: expected 25 points.", state: pickerError}},
		{over, "order", detail, pickerText{
			text: ": lion", state: pickerWarning,
			tip: "Something picked below needs attention: lion.",
		}},
		{over, "order", cost, pickerText{text: " [39~64 points]", tip: "Over: expected 5~25 points.", state: pickerError}},
		{over, "honors", detail, pickerText{text: ": cr"}},
		{over, "root", hint, pickerText{text: "Over by 14~49 points. Needs attention below: order. Override to keep " +
			"it anyway.", state: pickerError}},
		{over, "order", hint, pickerText{
			text:  "Needs attention below: lion. Override to keep it anyway.",
			state: pickerWarning,
		}},
		{misses, "fit", detail, pickerText{text: " (pick 1)", tip: "2 picked, but this asks for 1.", state: pickerError}},
		{misses, "fit", hint, pickerText{
			text:  "2 picked, but this asks for 1. Override to keep it anyway.",
			state: pickerError,
		}},
		{misses, "rose", cost, pickerText{text: " [11 points]", tip: "Under: expected 20 points.", state: pickerError}},
		{misses, "rose", hint, pickerText{text: "9 points short. Override to keep it anyway.", state: pickerError}},
		{misses, "graces", hint, pickerText{text: "Every pick has a fixed cost."}},
	} {
		s, n := newKnightSession()
		choose(s, n, tc.picks...)
		c.Equal(tc.want, tc.text(s, n[tc.row]), tc.row)
	}

	s, n := newKnightSession()
	wm := n["wm"]
	wm.Modifiers[0].Children[1].SetEnabled(true)
	c.Equal(pickerText{}, s.detail(wm), "a pick is only a default until answered")
	s.modsAnswered[wm] = true
	c.Equal(pickerText{text: " [+25]"}, s.detail(wm))

	// A rule asking for anything but its number says so when the picks come to it.
	s, n = newKnightSession()
	n["root"].TemplatePicker.Qualifier.Compare = criteria.NotEqualsNumber
	n["fit"].TemplatePicker.Qualifier.Compare = criteria.NotEqualsNumber
	choose(s, n, "ea", "ep", "luck", "shield", "fit1")
	c.Equal(pickerText{
		text:  "The picks come to 60 points, but this asks for anything but 60 points. Override to keep it anyway.",
		state: pickerError,
	}, s.hint(n["root"]))
	c.Equal(pickerText{
		text:  "1 picked, but this asks for anything but 1. Override to keep it anyway.",
		state: pickerError,
	}, s.hint(n["fit"]))
}
