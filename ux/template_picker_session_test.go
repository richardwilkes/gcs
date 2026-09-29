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
	box := unison.NewCheckBox()
	updatePickerCheckBoxTitle(box, n["fit2"], picker.Count, true)
	c.Equal("fit2 [15 points]", box.Text.String(), "a count choice shows what its options cost too")
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
