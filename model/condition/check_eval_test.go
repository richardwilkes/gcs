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

// TestCheckValueKinds checks every value kind against every criterion kind, including the Failed of a criterion that
// does not suit the value and of a field the subject's kind lacks. TestCheckAbsent covers Absent values.
func TestCheckValueKinds(t *testing.T) {
	h := newHost()
	subject := newItem("Fireball", map[string]Value{
		"tags":    listValue("Fire", "Missile"),
		"none":    listValue(),
		"level":   numberValue(3),
		"weight":  weightValue(5),
		"on":      boolValue(true),
		"off":     boolValue(false),
		"tl":      absentValue(),
		"college": textValue("Fire"),
	})
	for _, tc := range []struct {
		name  string
		node  *Check
		state State
		spec  int
		hit   bool
	}{
		{"text is", checkText("name", criteria.IsText, "fireball"), Met, 1, true},
		{"text is not", checkText("name", criteria.IsNotText, "fireball"), Unmet, 0, false},
		{"text no comma split", checkText("name", criteria.IsText, "Fireball, Water"), Unmet, 0, false},
		{"list is", checkText("tags", criteria.IsText, "missile"), Met, 1, true},
		{"list comma", checkText("tags", criteria.IsText, "Water, Fire"), Met, 1, true},
		{"list is not", checkText("tags", criteria.IsNotText, "Water, Fire"), Unmet, 0, false},
		{"empty list is not", checkText("none", criteria.IsNotText, "x"), Met, 1, true},
		{"empty list is", checkText("none", criteria.IsText, "x"), Unmet, 0, false},
		{"number", &Check{Field: "level", Number: numberCrit(criteria.AtLeastNumber, 3)}, Met, 1, true},
		{"weight", &Check{Field: "weight", Weight: weightCrit(criteria.AtMostNumber, 4)}, Unmet, 0, false},
		{"bool true", &Check{Field: "on"}, Met, 1, true},
		{"bool false", &Check{Field: "off"}, Unmet, 0, false},
		{"not bool false", &Check{Field: "off", Not: true}, Met, 1, false},
		{"any", &Check{Field: "name"}, Met, 0, false},
		{"any list", &Check{Field: "none"}, Met, 0, false},
		{"text on number", checkText("level", criteria.IsText, "3"), Failed, 0, false},
		{"number on text", &Check{Field: "name", Number: numberCrit(criteria.AtLeastNumber, 3)}, Failed, 0, false},
		{"weight on number", &Check{Field: "level", Weight: weightCrit(criteria.AtLeastNumber, 3)}, Failed, 0, false},
		{"text on bool", checkText("on", criteria.IsText, "true"), Failed, 0, false},
		{"missing field", checkText("missing", criteria.IsText, "x"), Failed, 0, false},
		{"not missing field", &Check{Field: "missing", Not: true}, Failed, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			r := evalFor(t, tc.node, h, nil, subject, collect)
			c.Equal(tc.state, r.State)
			c.Equal(tc.spec, r.Specificity)
			c.Equal(tc.hit, len(hits(&r, tc.node)) == 1)
			c.Equal(r.State, evalFor(t, tc.node, h, nil, subject, yesNo).State)
		})
	}
}

// TestCheckAbsent is the truth table of an Absent value: a subject that lacks the field meets only the negative
// criteria, which hold exactly when their positive form doesn't, so "not (tags is X)" and "tags is not X" agree. A
// Met check that tests something scores 1, and is a hit unless negated.
func TestCheckAbsent(t *testing.T) {
	h := newHost()
	subject := newItem("Fireball", map[string]Value{"tl": absentValue()})
	for _, crit := range []struct {
		name     string
		check    Check
		negative bool
	}{
		{"any", Check{Field: "tl"}, false},
		{"text", Check{Field: "tl", Text: textCrit(criteria.IsText, "")}, false},
		{"negative text", Check{Field: "tl", Text: textCrit(criteria.IsNotText, "3")}, true},
		{"does not contain", Check{Field: "tl", Text: textCrit(criteria.DoesNotContainText, "3")}, true},
		{"number", Check{Field: "tl", Number: numberCrit(criteria.EqualsNumber, 0)}, false},
		{"negative number", Check{Field: "tl", Number: numberCrit(criteria.NotEqualsNumber, 0)}, true},
		{"weight", Check{Field: "tl", Weight: weightCrit(criteria.AtMostNumber, 0)}, false},
		{"negative weight", Check{Field: "tl", Weight: weightCrit(criteria.NotEqualsNumber, 0)}, true},
	} {
		for _, not := range []bool{false, true} {
			node := crit.check
			node.Not = not
			t.Run(fmt.Sprintf("%s not=%v", crit.name, not), func(t *testing.T) {
				c := check.New(t)
				r := evalFor(t, &node, h, nil, subject, collect)
				met := crit.negative != not
				c.Equal(metState(met, false), r.State)
				c.Equal(met, r.Specificity == 1)
				c.Equal(met && !not, len(hits(&r, &node)) == 1)
				c.Equal(r.State, evalFor(t, &node, h, nil, subject, yesNo).State)
			})
		}
	}
}

// TestCheckHaveNeed checks the values a Number or Weight check reports, and the reasons it fails.
func TestCheckHaveNeed(t *testing.T) {
	c := check.New(t)
	h := newHost()
	subject := newItem("x", map[string]Value{"level": numberValue(3), "weight": weightValue(5)})
	r := evalFor(t, &Check{Field: "level", Number: numberCrit(criteria.AtLeastNumber, 4)}, h, nil, subject, yesNo)
	c.Equal(fxp.FromInteger(3), r.Have)
	c.Equal(ValueNumber, r.HaveKind)
	c.Equal(fxp.FromInteger(4), r.Need)
	r = evalFor(t, &Check{Field: "weight", Weight: weightCrit(criteria.AtLeastNumber, 2)}, h, nil, subject, yesNo)
	c.Equal(fxp.FromInteger(5), r.Have)
	c.Equal(ValueWeight, r.HaveKind)
	c.Equal(fxp.FromInteger(2), r.Need)
	r = evalFor(t, checkText("level", criteria.IsText, "3"), h, nil, subject, yesNo)
	c.Equal(Failed, r.State)
	c.Equal(`the criterion does not suit the value of field "level" (number)`, r.Reason)
	r = evalFor(t, &Check{Field: "gone"}, h, nil, subject, collect)
	c.Equal(`no field "gone"`, r.Reason)
	r = evalFor(t, &Check{Field: "level", Number: numberCrit(criteria.AtLeastNumber, 1)}, h, nil, newItem("x", map[string]Value{"level": absentValue()}), collect)
	c.Equal(Unmet, r.State)
	c.Equal(ValueAbsent, r.HaveKind)
}

// TestReplacements checks that qualifiers use the host's replacements, fetched once per (self, subject).
func TestReplacements(t *testing.T) {
	c := check.New(t)
	h := newHost()
	h.replacements = map[string]string{"Spell": "Fireball"}
	subject := newItem("Fireball", nil)
	g := &Group{Children: Nodes{
		checkText("name", criteria.IsText, "@Spell@"),
		checkText("name", criteria.StartsWithText, "@Spell@"),
		checkText("name", criteria.IsText, "Fireball"),
	}}
	c.Equal(Met, evalFor(t, g, h, nil, subject, collect).State)
	c.Equal(int64(1), h.replaced.Load())
}
