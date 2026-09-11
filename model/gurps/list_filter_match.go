// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

// NewListFilterMatcher returns a function that reports whether a node passes the filter, with the fields the filter
// refers to looked up once rather than for every node. A nil filter, or one without a root, passes everything.
func NewListFilterMatcher[T Node[T]](f *ListFilter, fields []*FilterField[T]) func(T) bool {
	if f == nil || f.Root == nil {
		return func(T) bool { return true }
	}
	byKey := make(map[string]*FilterField[T], len(fields))
	for _, field := range fields {
		byKey[field.Key] = field
	}
	root := f.Root
	return func(node T) bool { return matchesFilterNode(root, byKey, node) }
}

// matchesFilterNode returns true if the node passes the filter node. A node kind this version of GCS doesn't
// understand never passes, whatever its negation says, since its intent can't be known.
func matchesFilterNode[T Node[T]](n FilterNode, fields map[string]*FilterField[T], node T) bool {
	switch one := n.(type) {
	case *FilterGroup:
		return matchesFilterGroup(one, fields, node)
	case *FilterCondition:
		return matchesFilterCondition(one, fields, node)
	default:
		return false
	}
}

// matchesFilterGroup returns true if the node passes the group: all of its children, or any one of them, depending
// on how the group combines them. A group with no children passes everything either way. The result is then inverted
// if the group is negated.
func matchesFilterGroup[T Node[T]](g *FilterGroup, fields map[string]*FilterField[T], node T) bool {
	result := true
	if g.All {
		for _, child := range g.Children {
			if !matchesFilterNode(child, fields, node) {
				result = false
				break
			}
		}
	} else if len(g.Children) != 0 {
		result = false
		for _, child := range g.Children {
			if matchesFilterNode(child, fields, node) {
				result = true
				break
			}
		}
	}
	return result != g.Not
}

// matchesFilterCondition returns true if the node's field satisfies the condition's criteria, inverted if the
// condition is negated. A condition on a field this version of GCS doesn't know never passes, whatever its negation
// says, for the same reason an unknown node never does.
func matchesFilterCondition[T Node[T]](c *FilterCondition, fields map[string]*FilterField[T], node T) bool {
	field, ok := fields[c.Field]
	if !ok {
		return false
	}
	var result bool
	switch field.Kind {
	case FilterFieldText:
		result = c.Text.Matches(nil, field.Text(node))
	case FilterFieldList:
		result = c.Text.MatchesList(nil, field.List(node)...)
	case FilterFieldNumber:
		result = c.Number.Matches(field.Number(node))
	case FilterFieldWeight:
		result = c.Weight.Matches(field.Weight(node))
	case FilterFieldBool:
		result = field.Bool(node)
	default:
		return false
	}
	return result != c.Not
}
