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
	"strconv"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestScanSelfTokens(t *testing.T) {
	c := check.New(t)
	for q, want := range map[string][]qualifierPart{
		"{self.college}":        {{text: "college", token: true}},
		"a {self.x}, b{self.y}": {{text: "a "}, {text: "x", token: true}, {text: ", b"}, {text: "y", token: true}},
		"{{self.x}":             {{text: "{self.x}"}},
		"{{{self.x_2}}":         {{text: "{"}, {text: "x_2", token: true}, {text: "}"}},
		"":                      nil,
		"{self.a}{self.b}":      {{text: "a", token: true}, {text: "b", token: true}},
	} {
		sq, problem := scanSelfTokens(q)
		c.Equal("", problem, q)
		c.Equal(want, sq.parts, q)
	}
	for _, q := range []string{"{self.}", "{self.x", "a{b", "{self.a b}", "{other.x}", "{"} {
		sq, problem := scanSelfTokens(q)
		c.Nil(sq, q)
		c.NotEqual("", problem, q)
		c.Equal(problem, tokenProblem(q))
	}
	c.Equal("", tokenProblem("no braces"))
}

// TestSelfTokens checks self token matching: escapes, list splitting, list token values multiplying the alternatives,
// Absent and missing self fields, and the alternative cap.
func TestSelfTokens(t *testing.T) {
	h := newHost()
	h.replacements = map[string]string{"P": "{self.college}", "Q": "Big", "{": ""}
	many := make([]string, 9)
	for i := range many {
		many[i] = strconv.Itoa(i)
	}
	self := newItem("self", map[string]Value{
		"college":  textValue("Fire"),
		"colleges": listValue("Fire", "Air"),
		"empty":    listValue(),
		"commas":   textValue("A, B"),
		"tl":       absentValue(),
		"level":    numberValue(3),
		"flag":     boolValue(true),
		"many":     listValue(many...),
	})
	for _, tc := range []struct {
		name    string
		check   *Check
		subject map[string]Value
		state   State
	}{
		{"text", checkText("v", criteria.IsText, "{self.college}"), map[string]Value{"v": textValue("fire")}, Met},
		{"text no", checkText("v", criteria.IsText, "{self.college}"), map[string]Value{"v": textValue("air")}, Unmet},
		{"escape", checkText("v", criteria.IsText, "{{self.college}"), map[string]Value{"v": textValue("{self.college}")}, Met},
		{"list token", checkText("v", criteria.IsText, "{self.colleges}"), map[string]Value{"v": textValue("Air")}, Met},
		{"list token negative", checkText("v", criteria.IsNotText, "{self.colleges}"), map[string]Value{"v": textValue("Air")}, Unmet},
		{"list token negative met", checkText("v", criteria.IsNotText, "{self.colleges}"), map[string]Value{"v": textValue("Water")}, Met},
		{"empty list token", checkText("v", criteria.IsText, "x{self.empty}"), map[string]Value{"v": textValue("x")}, Met},
		{"list value split", checkText("v", criteria.IsText, "Water, {self.college} ,"), map[string]Value{"v": listValue("Mind", "fire")}, Met},
		{"list value empty", checkText("v", criteria.IsText, " , {self.empty}"), map[string]Value{"v": listValue()}, Met},
		{"text value untrimmed", checkText("v", criteria.StartsWithText, " {self.college} "), map[string]Value{"v": textValue(" fire ball")}, Met},
		{"text value untrimmed no", checkText("v", criteria.StartsWithText, " {self.college} "), map[string]Value{"v": textValue("fire ball")}, Unmet},
		{"list piece unicode trim", checkText("v", criteria.IsText, "\u00a0{self.college}\v, x"), map[string]Value{"v": listValue("fire")}, Met},
		{"token value not split", checkText("v", criteria.IsText, "{self.commas}"), map[string]Value{"v": listValue("A, B")}, Met},
		{"token value not split 2", checkText("v", criteria.IsText, "{self.commas}"), map[string]Value{"v": listValue("A")}, Unmet},
		{"list does not contain", checkText("v", criteria.DoesNotContainText, "{self.colleges}"), map[string]Value{"v": listValue("Earth")}, Met},
		{"starts with", checkText("v", criteria.DoesNotStartWithText, "{self.college}"), map[string]Value{"v": textValue("Fireball")}, Unmet},
		{"ends with", checkText("v", criteria.DoesNotEndWithText, "{self.college}"), map[string]Value{"v": textValue("Fireball")}, Met},
		{"number token", checkText("v", criteria.IsText, "L{self.level}"), map[string]Value{"v": textValue("L3")}, Met},
		{"bool token", checkText("v", criteria.IsText, "{self.flag}"), map[string]Value{"v": textValue("true")}, Met},
		{"absent token empty", checkText("v", criteria.IsText, "{self.tl}"), map[string]Value{"v": textValue("")}, Met},
		{"missing token", checkText("v", criteria.IsText, "{self.missing}"), map[string]Value{"v": textValue("")}, Failed},
		{"cap", checkText("v", criteria.IsText, "{self.many}{self.many}"), map[string]Value{"v": textValue("11")}, Failed},
		{"under cap", checkText("v", criteria.IsText, "{self.many}{self.colleges}"), map[string]Value{"v": textValue("8air")}, Met},
		{"cap across pieces", checkText("v", criteria.IsText, "{self.many}{self.colleges}, {self.many}{self.colleges}, {self.many}{self.colleges}, {self.many}{self.colleges}"), map[string]Value{"v": listValue("x")}, Failed},
		{"replacement is literal", checkText("v", criteria.IsText, "@P@-{self.college}"), map[string]Value{"v": textValue("{self.college}-Fire")}, Met},
		{"replacement in list piece", checkText("v", criteria.IsText, "@Q@ {self.college}, z"), map[string]Value{"v": listValue("big fire")}, Met},
		{"replaced to nothing", checkText("v", criteria.IsText, "@{{@"), map[string]Value{"v": listValue()}, Met},
		{"absent subject", checkText("v", criteria.IsText, "{self.tl}"), map[string]Value{"v": absentValue()}, Unmet},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			subject := newItem("subject", tc.subject)
			c.Equal(tc.state, evalFor(t, tc.check, h, self, subject, yesNo).State)
		})
	}
}
