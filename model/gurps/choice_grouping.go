// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"slices"

	"github.com/richardwilkes/toolbox/v2/xreflect"
)

// choiceGrouping holds the template-only flag that makes a group inside a template choice only organize its options,
// which are then picked one by one, rather than being picked as a unit.
type choiceGrouping struct {
	PickSeparately bool `json:"pick_separately,omitzero"`
}

func (g *choiceGrouping) pickSeparately() *bool {
	return &g.PickSeparately
}

type choiceGrouper interface {
	pickSeparately() *bool
	canPickSeparately() bool
}

// CanPickSeparately returns true if the node is a kind of group that may be picked from separately: a trait group, a
// skill or spell container or an equipment group, none of them a choice itself.
func CanPickSeparately[T Node[T]](node T) bool {
	g, ok := any(node).(choiceGrouper)
	return ok && !xreflect.IsNil(node) && g.canPickSeparately()
}

// PickSeparatelyHasEffect returns true if the node can be picked from separately and sits in a template choice,
// directly or through groups that are themselves picked from separately.
func PickSeparatelyHasEffect[T Node[T]](node T) bool {
	if !CanPickSeparately(node) {
		return false
	}
	for parent := node.Parent(); !xreflect.IsNil(parent); parent = parent.Parent() {
		if IsTemplateChoiceContainer(parent) {
			return true
		}
		if !pickedSeparately(parent) {
			return false
		}
	}
	return false
}

// IsOrganizingGroup returns true if the node is a group whose options are picked one by one in the template choice it
// sits in (see PickSeparatelyHasEffect).
func IsOrganizingGroup[T Node[T]](node T) bool {
	return !xreflect.IsNil(node) && pickedSeparately(node) && PickSeparatelyHasEffect(node)
}

// TemplateChoiceOptions returns what the template choice container offers: its children, with each organizing group
// among them replaced by the options it holds. Anything else returns its children.
func TemplateChoiceOptions[T Node[T]](container T) []T {
	children := container.NodeChildren()
	if !IsTemplateChoiceContainer(container) {
		return children
	}
	return organizedOptions(children)
}

// organizedOptions returns the children with each group picked from separately replaced by its options, or the
// children themselves, without allocating, when there is none.
func organizedOptions[T Node[T]](children []T) []T {
	for i, child := range children {
		if !pickedSeparately(child) {
			continue
		}
		options := slices.Clip(children[:i])
		for _, one := range children[i:] {
			if pickedSeparately(one) {
				options = append(options, organizedOptions(one.NodeChildren())...)
			} else {
				options = append(options, one)
			}
		}
		return options
	}
	return children
}

// pickedSeparately returns true if the node is flagged to be picked from separately and can be.
func pickedSeparately[T Node[T]](node T) bool {
	g, ok := any(node).(choiceGrouper)
	return ok && *g.pickSeparately() && g.canPickSeparately()
}

// clearPickSeparately clears the node's flag to be picked from separately, if it has one.
func clearPickSeparately[T Node[T]](node T) {
	if g, ok := any(node).(choiceGrouper); ok {
		*g.pickSeparately() = false
	}
}
