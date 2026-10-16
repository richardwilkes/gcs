// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package condition

import (
	"fmt"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
)

// libraryCharacter returns a host whose entity has traits and spells, and the entity's owner of a prereq (self).
func libraryCharacter(spells []string, fireExtra int, traits ...*fakeItem) (h *fakeHost, self *fakeItem) {
	h = newHost()
	self = newItem("Owner", map[string]Value{"level": numberValue(9)})
	list := make([]*fakeItem, 0, len(spells)+fireExtra)
	for _, name := range spells {
		list = append(list, newItem(name, map[string]Value{"college": listValue("Fire")}))
	}
	for i := range fireExtra {
		list = append(list, newItem(fmt.Sprintf("Fire %d", i), map[string]Value{"college": listValue("Fire")}))
	}
	h.entity["spells"] = list
	h.entity["traits"] = append(traits, self)
	return h, self
}

func trait(name string, level int64, tags ...string) *fakeItem {
	return newItem(name, map[string]Value{"level": numberValue(level), "tags": listValue(tags...)})
}

func named(of, name string, quantity *criteria.Number) *Count {
	return &Count{Of: of, Where: Ref{Node: checkText("name", criteria.IsText, name)}, Quantity: quantity}
}

func magery(level int64) *Count {
	return &Count{Of: "traits", Where: Ref{Node: &Group{Children: Nodes{
		checkText("name", criteria.IsText, "Magery"),
		&Check{Field: "level", Number: numberCrit(criteria.AtLeastNumber, level)},
	}}}}
}

func attrAtLeast(attr string, n int64) *Compare {
	return &Compare{Left: Operand{Attribute: attr}, Op: OpAtLeast, Right: value(n)}
}

// TestFireSwarm checks the non-exclusive shape: the named spells count toward the college total.
func TestFireSwarm(t *testing.T) {
	c := check.New(t)
	tree := &Group{Children: Nodes{
		named("spells", "Fireball", nil),
		named("spells", "Flame Jet", nil),
		&Count{
			Of: "spells", Where: Ref{Node: checkText("college", criteria.IsText, "Fire")},
			Quantity: atLeast(10),
		},
	}}
	h, self := libraryCharacter([]string{"Fireball", "Flame Jet"}, 8)
	r := eval(t, tree, h, self, collect)
	c.Equal(Met, r.State)
	c.Equal(3, r.Specificity)
	c.Equal(fxp.FromInteger(10), r.Children[2].Have)
	h, self = libraryCharacter([]string{"Fireball", "Flame Jet"}, 7)
	r = eval(t, tree, h, self, collect)
	c.Equal(Unmet, r.State)
	c.Equal(fxp.FromInteger(9), r.Children[2].Have)
	c.Equal(fxp.FromInteger(10), r.Children[2].Need)
}

// TestPsiCap checks a cap relative to another trait: level at most the summed levels of same-named traits plus 6.
func TestPsiCap(t *testing.T) {
	c := check.New(t)
	tree := &Compare{
		Left: Operand{Self: "level"},
		Op:   OpAtMost,
		Right: Operand{
			Aggregate: AggregateSum,
			Members:   Members{Of: "traits", Where: Ref{Node: checkText("name", criteria.IsText, "Psi Talent")}},
			Field:     "level",
			Offset:    fxp.FromInteger(6),
		},
	}
	h, self := libraryCharacter(nil, 0, trait("Psi Talent", 2), trait("Psi Talent", 1))
	c.Equal(Met, eval(t, tree, h, self, yesNo).State)
	h, self = libraryCharacter(nil, 0, trait("Psi Talent", 2))
	r := eval(t, tree, h, self, collect)
	c.Equal(Unmet, r.State)
	c.Equal(fxp.FromInteger(9), r.Have)
	c.Equal(fxp.FromInteger(8), r.Need)
}

// TestWillfulCasting checks an attribute against an attribute.
func TestWillfulCasting(t *testing.T) {
	c := check.New(t)
	tree := &Compare{Left: Operand{Attribute: "will"}, Op: OpMoreThan, Right: Operand{Attribute: "iq"}}
	h, self := libraryCharacter(nil, 0)
	h.attributes["will"] = fxp.FromInteger(13)
	h.attributes["iq"] = fxp.FromInteger(12)
	c.Equal(Met, eval(t, tree, h, self, yesNo).State)
	h.attributes["iq"] = fxp.FromInteger(13)
	c.Equal(Unmet, eval(t, tree, h, self, yesNo).State)
}

// TestDoomWish checks the trade-off between Magery level and attributes, an ordinary tree.
func TestDoomWish(t *testing.T) {
	tree := &Group{Children: Nodes{
		&Group{Mode: GroupAny, Children: Nodes{
			&Group{Children: Nodes{magery(6), attrAtLeast("st", 17)}},
			&Group{Children: Nodes{magery(5), attrAtLeast("st", 18)}},
			&Group{Children: Nodes{magery(4), attrAtLeast("iq", 19)}},
			attrAtLeast("iq", 20),
		}},
		named("spells", "Wish", nil),
		magery(3),
	}}
	for _, tc := range []struct {
		magery, st, iq int64
		want           State
	}{
		{6, 17, 10, Met},
		{5, 17, 10, Unmet},
		{5, 18, 10, Met},
		{4, 10, 19, Met},
		{3, 10, 20, Met},
		{3, 10, 19, Unmet},
		{2, 10, 20, Unmet},
	} {
		h, self := libraryCharacter([]string{"Wish"}, 0, trait("Magery", tc.magery))
		h.attributes["st"] = fxp.FromInteger(tc.st)
		h.attributes["iq"] = fxp.FromInteger(tc.iq)
		check.New(t).Equal(tc.want, eval(t, tree, h, self, collect).State, fmt.Sprint(tc))
	}
}

// TestWealth checks a "none of a set" rule by tag, where the count leaves the owner itself out unless told to keep it.
func TestWealth(t *testing.T) {
	c := check.New(t)
	tree := &Count{
		Of: "traits", Where: Ref{Node: checkText("tags", criteria.ContainsText, "Wealth")},
		Not: true,
	}
	h, self := libraryCharacter(nil, 0, trait("Alertness", 1, "Mental"))
	self.fields["tags"] = listValue("Wealth")
	c.Equal(Met, eval(t, tree, h, self, yesNo).State)
	tree.IncludeSelf = true
	c.Equal(Unmet, eval(t, tree, h, self, yesNo).State)
	tree.IncludeSelf = false
	h, self = libraryCharacter(nil, 0, trait("Filthy Rich", 1, "Wealth"))
	c.Equal(Unmet, eval(t, tree, h, self, yesNo).State)
}

// TestContainedWeight checks the converted contained weight shape on containers and non-containers.
func TestContainedWeight(t *testing.T) {
	c := check.New(t)
	lb := fxp.Weight(fxp.FromStringForced("10.79"))
	tree := &Group{Mode: GroupAny, Children: Nodes{
		&Check{Field: "container", Not: true},
		&Compare{Left: Operand{Self: "contained_weight"}, Op: OpAtMost, Right: Operand{Weight: &lb}},
	}}
	h := newHost()
	for _, tc := range []struct {
		container bool
		weight    int64
		want      State
	}{
		{false, 50, Met},
		{true, 10, Met},
		{true, 11, Unmet},
	} {
		self := newItem("Sack", map[string]Value{"container": boolValue(tc.container), "contained_weight": weightValue(tc.weight)})
		c.Equal(tc.want, eval(t, tree, h, self, collect).State, fmt.Sprint(tc))
	}
}

// TestTechLevelRule checks the skill and spell tech level rule as the converter writes it for an owner with a tech
// level: a candidate without one passes, and one with one passes when it matches the owner's, "" included.
func TestTechLevelRule(t *testing.T) {
	c := check.New(t)
	tree := &Count{Of: "skills", Where: Ref{Node: &Group{Children: Nodes{
		checkText("name", criteria.IsText, "Guns"),
		&Group{Mode: GroupAny, Children: Nodes{
			&Check{Field: "tech_level", Not: true},
			checkText("tech_level", criteria.IsText, "{self.tech_level}"),
		}},
	}}}}
	tl := func(s *string) Value {
		if s == nil {
			return absentValue()
		}
		return textValue(*s)
	}
	empty, three, four := "", "3", "4"
	for _, tc := range []struct {
		self, skill *string
		want        State
	}{
		{&three, nil, Met},
		{&empty, &empty, Met},
		{&empty, &three, Unmet},
		{&three, &empty, Unmet},
		{&three, &three, Met},
		{&three, &four, Unmet},
	} {
		h := newHost()
		self := newItem("Owner", map[string]Value{"tech_level": tl(tc.self)})
		h.entity["skills"] = []*fakeItem{newItem("Guns", map[string]Value{"tech_level": tl(tc.skill)})}
		c.Equal(tc.want, eval(t, tree, h, self, yesNo).State, fmt.Sprint(*tc.self, tc.skill))
	}
}
