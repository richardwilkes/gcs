// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import "strings"

// NewListFilterChecker returns a function that checks a node against the filter, with the fields the filter refers to
// looked up once rather than for every node. A nil filter, or one without a root, is skipped for every node.
func NewListFilterChecker[T Node[T]](f *ListFilter, fields []*FilterField[T]) func(T) CheckResult {
	if f == nil || f.Root == nil {
		return func(T) CheckResult { return CheckSkipped }
	}
	byKey := make(map[string]*FilterField[T], len(fields))
	for _, field := range fields {
		byKey[field.Key] = field
	}
	root := f.Root
	return func(node T) CheckResult { return checkFilterNode(root, byKey, node) }
}

// NewListFilterMatcher returns a function that reports whether a node passes the filter: whether NewListFilterChecker
// finds the filter met or skipped for it. A filter with nothing in it is skipped, so it passes everything, while one
// that is unmet or can't be checked hides the node.
func NewListFilterMatcher[T Node[T]](f *ListFilter, fields []*FilterField[T]) func(T) bool {
	check := NewListFilterChecker(f, fields)
	return func(node T) bool {
		result := check(node)
		return result == CheckMet || result == CheckSkipped
	}
}

// checkFilterNode checks the node against the filter node. A node kind this version of GCS doesn't understand fails,
// since its intent can't be known.
func checkFilterNode[T Node[T]](n FilterNode, fields map[string]*FilterField[T], node T) CheckResult {
	switch one := n.(type) {
	case *FilterGroup:
		return checkFilterGroup(one, fields, node)
	case *FilterCondition:
		return checkFilterCondition(one, fields, node)
	default:
		return CheckFailed
	}
}

// checkFilterGroup checks the node against the group, as a list of prerequisites is checked: children that are skipped
// are left out, and a group with nothing left is skipped. Of the rest, one unmet child makes an "all of" group unmet,
// and one met child makes an "any of" group met, even when others failed. Otherwise the group has failed when any of
// the rest has, and is met or unmet as they all are. Negating the group then swaps met and unmet.
func checkFilterGroup[T Node[T]](g *FilterGroup, fields map[string]*FilterField[T], node T) CheckResult {
	var tally checkTally
	for _, child := range g.Children {
		tally.add(checkFilterNode(child, fields, node))
	}
	return tally.combine(g.All).negate(g.Not)
}

// checkFilterCondition checks the node's field against the condition's criteria, swapping met and unmet when the
// condition is negated. A node whose text field is empty or only space, whose list field holds nothing, or that the
// field's Has says lacks it, such as a trait that can't be leveled, doesn't have the field, as the condition's title
// puts it, so it satisfies no criteria. A condition on a field this version of GCS doesn't know fails, for the same
// reason an unknown node does.
func checkFilterCondition[T Node[T]](c *FilterCondition, fields map[string]*FilterField[T], node T) CheckResult {
	field, ok := fields[c.Field]
	if !ok {
		return CheckFailed
	}
	var met bool
	if field.Has != nil && !field.Has(node) {
		return CheckUnmet.negate(c.Not)
	}
	switch field.Kind {
	case FilterFieldText:
		if value := field.Text(node); strings.TrimSpace(value) != "" {
			met = c.Text.Matches(nil, value)
		}
	case FilterFieldList:
		if values := field.List(node); len(values) != 0 {
			met = c.Text.MatchesList(nil, values...)
		}
	case FilterFieldNumber:
		met = c.Number.Matches(field.Number(node))
	case FilterFieldWeight:
		met = c.Weight.Matches(field.Weight(node))
	case FilterFieldBool:
		met = field.Bool(node)
	default:
		return CheckFailed
	}
	result := CheckUnmet
	if met {
		result = CheckMet
	}
	return result.negate(c.Not)
}
