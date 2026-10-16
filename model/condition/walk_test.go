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
	"slices"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

func kindName(n Node) string {
	switch n.(type) {
	case *Group:
		return "group"
	case *Check:
		return "check"
	case *Const:
		return "const"
	case *Count:
		return "count"
	case *Compare:
		return "compare"
	default:
		return "script"
	}
}

func walkPaths(n Node, fn func(path Path, n Node) bool) []string {
	var list []string
	Walk(n, func(path Path, n Node) bool {
		list = append(list, path.String()+" "+kindName(n))
		return fn(path, n)
	})
	return list
}

func TestWalk(t *testing.T) {
	c := check.New(t)
	tree := fullTree()
	tree.Children = append(tree.Children, nil, &Group{OnlyWhen: Ref{Node: &Group{}}})
	all := func(Path, Node) bool { return true }
	c.Equal([]string{
		" group",
		"only_when check",
		"children[0] const",
		"children[1] count",
		"children[1].where check",
		"children[2] compare",
		"children[2].left.where check",
		"children[2].left.sum[1].where check",
		"children[3] script",
		"children[4] group",
		"children[4].children[0] const",
		"children[6] group",
	}, walkPaths(tree, all))

	// Returning false skips the nodes beneath.
	c.Equal([]string{
		" group",
		"only_when check",
		"children[0] const",
		"children[1] count",
		"children[2] compare",
		"children[3] script",
		"children[4] group",
		"children[6] group",
	}, walkPaths(tree, func(path Path, _ Node) bool { return len(path) == 0 }))

	// A kept path is not overwritten by later steps once cloned, and an appended one never clobbers the walk.
	var kept []Path
	Walk(tree, func(path Path, _ Node) bool {
		kept = append(kept, slices.Clone(path))
		_ = append(path, Slot{Kind: SlotRight})
		return true
	})
	c.Equal("children[2].left.sum[1].where", kept[7].String())

	c.Equal([]string(nil), walkPaths(nil, all))
}

func TestSlotStrings(t *testing.T) {
	c := check.New(t)
	c.Equal("children[3].only_when.where.left.right.sum[1].?", Path{
		{Kind: SlotChild, Index: 3},
		{Kind: SlotOnlyWhen},
		{Kind: SlotWhere},
		{Kind: SlotLeft},
		{Kind: SlotRight},
		{Kind: SlotSum, Index: 1},
		{Kind: 99},
	}.String())
	c.Equal("", Path(nil).String())
}

func TestOperandSource(t *testing.T) {
	c := check.New(t)
	for want, o := range map[OperandSource]Operand{
		SourceNone:      {},
		SourceValue:     {Value: intPtr(0)},
		SourceWeight:    {Weight: weightPtr(0)},
		SourceAttribute: {Attribute: "st"},
		SourceEntity:    {Entity: "sm"},
		SourceSelf:      {Self: "level"},
		SourceAggregate: {Aggregate: AggregateCount},
		SourceSum:       {Sum: []Operand{{}}},
		SourceMany:      {Value: intPtr(1), Sum: []Operand{{}}},
	} {
		c.Equal(want, o.Source())
	}
}
