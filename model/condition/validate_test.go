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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
)

func problemStrings(problems []Problem) []string {
	if len(problems) == 0 {
		return nil
	}
	list := make([]string, 0, len(problems))
	for _, p := range problems {
		list = append(list, p.Error())
	}
	return list
}

// nestedNode returns levels groups nested one inside the next, with leaf as the child of the innermost.
func nestedNode(levels int, leaf Node) Node {
	n := leaf
	for range levels {
		n = &Group{Children: Nodes{n}}
	}
	return n
}

func TestValidate(t *testing.T) {
	spells := func(q *criteria.Number) *Count { return &Count{Of: "spells", Quantity: q} }
	depth := strings.Repeat("children[0].", MaxDepth)
	for _, tc := range []struct {
		name string
		node Node
		want []string
	}{
		{"valid", &Group{Mode: GroupAtLeast, N: 1, Children: Nodes{
			checkText("name", criteria.IsText, "{self.name} {{x}"),
			&Script{Script: "true"},
			&Count{Of: "spells", GroupBy: "college", PerGroup: atLeast(2), Exclusive: true},
			&Compare{Left: Operand{Aggregate: AggregateSum, Members: Members{Of: "traits"}, Field: "level"}, Right: value(1)},
			&Group{Exclusive: true, Children: Nodes{spells(nil), spells(atLeast(12))}},
		}}, nil},
		{"nil", nil, nil},
		{"empty group", &Group{}, nil},
		{"at_least zero", &Group{Mode: GroupAtLeast, Children: Nodes{&Const{}}}, []string{
			"the number of children needed is outside 1 to the number of children",
		}},
		{"at_least too many", &Group{Mode: GroupAtLeast, N: 2, Children: Nodes{&Const{}}}, []string{
			"the number of children needed is outside 1 to the number of children",
		}},
		{"n on all", &Group{N: 1, Children: Nodes{&Const{}}}, []string{
			`a number of children needed is set, but the mode is not "at least"`,
		}},
		{"unknown mode", &Group{Mode: 9}, []string{"unknown group mode"}},
		{"nil child", &Group{Children: Nodes{&Const{}, nil}}, []string{"children[1]: missing child"}},
		{"exclusive any", &Group{Mode: GroupAny, Exclusive: true, Children: Nodes{spells(nil)}}, []string{
			`an exclusive group's mode is not "all"`,
		}},
		{"exclusive non-count", &Group{Exclusive: true, Children: Nodes{nil, &Const{}}}, []string{
			"an exclusive group holds something other than a count", "children[0]: missing child",
		}},
		{"exclusive negated", &Group{Exclusive: true, Children: Nodes{&Count{Of: "spells", Not: true}}}, []string{
			"an exclusive group holds a count that is negated or grouped",
		}},
		{"exclusive grouped", &Group{Exclusive: true, Children: Nodes{&Count{Of: "spells", GroupBy: "college"}}}, []string{
			"an exclusive group holds a count that is negated or grouped",
		}},
		{"exclusive at most", &Group{Exclusive: true, Children: Nodes{spells(numberCritPtr(criteria.AtMostNumber, 2))}}, []string{
			`an exclusive group holds a count whose quantity is not "at least" a whole number`,
		}},
		{"exclusive fraction", &Group{Exclusive: true, Children: Nodes{spells(&criteria.Number{
			Compare: criteria.AtLeastNumber, Qualifier: fxp.FromStringForced("1.5"),
		})}}, []string{`an exclusive group holds a count whose quantity is not "at least" a whole number`}},
		{"exclusive collections", &Group{Exclusive: true, Children: Nodes{spells(nil), &Count{Of: "skills"}}}, []string{
			"an exclusive group holds counts over different collections",
		}},
		{"check without field", &Check{}, []string{"no field"}},
		{"two criteria", &Check{Field: "f", Text: textCrit(criteria.IsText, "a"), Number: numberCrit(criteria.AtLeastNumber, 1)}, []string{
			"more than one criterion",
		}},
		{"unknown text comparison", &Check{Field: "f", Text: textCrit(99, "a")}, []string{"a criterion has an unknown comparison"}},
		{"unknown number comparison", &Check{Field: "f", Number: numberCrit(99, 1)}, []string{"a criterion has an unknown comparison"}},
		{"unknown weight comparison", &Check{Field: "f", Weight: weightCrit(99, 1)}, []string{"a criterion has an unknown comparison"}},
		{"lone brace", checkText("f", criteria.IsText, "a{b"), []string{
			`a "{" that neither starts a {self.<field>} token nor is doubled`,
		}},
		{"bad token", checkText("f", criteria.IsText, "{self.a b}"), []string{
			`a "{" that neither starts a {self.<field>} token nor is doubled`,
		}},
		{"empty token", checkText("f", criteria.IsText, "{self.}"), []string{
			`a "{" that neither starts a {self.<field>} token nor is doubled`,
		}},
		{"count without of", &Count{}, []string{"no collection"}},
		{"per_group without group_by", &Count{Of: "x", PerGroup: atLeast(2)}, []string{
			"a per-group quantity without a group-by field",
		}},
		{"exclusive without group_by", &Count{Of: "x", Exclusive: true}, []string{"exclusive without a group-by field"}},
		{"exclusive per_group", &Count{Of: "x", GroupBy: "y", Exclusive: true, PerGroup: &criteria.Number{}}, []string{
			`an exclusive count's per-group quantity is not "at least" a whole number`,
		}},
		{"count where", &Count{Of: "x", Where: Ref{Node: &Check{}}}, []string{"where: no field"}},
		{"operands without source", &Compare{}, []string{"left: an operand has no value", "right: an operand has no value"}},
		{"operand with two sources", &Compare{Left: Operand{Attribute: "st", Entity: "sm"}, Right: Operand{Self: "level"}}, []string{
			"left: an operand has more than one value",
		}},
		{"unknown op", &Compare{Op: 9, Left: value(1), Right: value(1)}, []string{"unknown operator"}},
		{"mixed constants", &Compare{Left: Operand{Sum: []Operand{value(1), {Attribute: "st"}}}, Right: Operand{Weight: weightPtr(1)}}, []string{
			"number and weight constants are mixed",
		}},
		{"mixed constants on one side", &Compare{Left: Operand{Sum: []Operand{value(1), {Weight: weightPtr(1)}}}, Right: Operand{Entity: "sm"}}, []string{
			"number and weight constants are mixed",
		}},
		{"aggregate problems", &Compare{
			Left:  Operand{Aggregate: AggregateMax, Members: Members{Where: Ref{Node: &Script{Lang: "lua"}}}},
			Right: Operand{Aggregate: AggregateCount, Members: Members{Of: "x"}, Field: "level"},
		}, []string{
			"left: an aggregate has no field",
			"left: an aggregate has no collection",
			"left.where: unknown script language",
			"right: a count has a field",
		}},
		{"members without aggregate", &Compare{Left: Operand{Value: intPtr(1), Members: Members{Of: "x"}}, Right: value(1)}, []string{
			"left: an operand has a collection or field but no aggregate",
		}},
		{"unknown aggregate", &Compare{Left: Operand{Aggregate: 77, Members: Members{Of: "x"}}, Right: value(1)}, []string{
			"left: unknown aggregate",
		}},
		{"sum problems", &Compare{Left: Operand{Sum: []Operand{{}, value(1)}}, Right: value(1)}, []string{
			"left.sum[0]: an operand has no value",
		}},
		{"script language", &Script{Script: "x", Lang: "js"}, []string{"unknown script language"}},
		{"pre-order", &Group{
			N:        1,
			OnlyWhen: Ref{Node: &Check{}},
			Children: Nodes{&Count{}, &Group{Children: Nodes{nil}}},
		}, []string{
			`a number of children needed is set, but the mode is not "at least"`,
			"only_when: no field",
			"children[0]: no collection",
			"children[1].children[0]: missing child",
		}},
		{"deepest", nestedNode(MaxDepth-1, &Const{}), nil},
		{"too deep", nestedNode(MaxDepth, &Check{}), []string{depth[:len(depth)-1] + ": nested more than 32 deep"}},
		{"too deep operand", nestedNode(MaxDepth-1, &Compare{Left: value(1), Right: value(1)}), []string{
			depth[:len(depth)-len("children[0].")] + "left: nested more than 32 deep",
			depth[:len(depth)-len("children[0].")] + "right: nested more than 32 deep",
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			check.New(t).Equal(tc.want, problemStrings(Validate(tc.node)))
		})
	}
}

// TestProblemsFailTheTree checks that a tree with a problem anywhere is Failed as a whole, with the first problem as
// the reason, even where the rest would decide it.
func TestProblemsFailTheTree(t *testing.T) {
	c := check.New(t)
	tree := &Group{Mode: GroupAny, Children: Nodes{&Const{Value: true}, &Group{Children: Nodes{&Compare{}, &Count{}}}}}
	for _, opt := range []Options{yesNo, collect, explain} {
		r := eval(t, tree, newHost(), nil, opt)
		c.Equal(Failed, r.State)
		c.Equal("children[1].children[0].left: an operand has no value", r.Reason)
		c.Equal(Node(tree), r.Node)
	}
	state, err := NewMatcher(tree, newHost(), Options{}).State(t.Context(), nil)
	c.NoError(err)
	c.Equal(Failed, state)
}
