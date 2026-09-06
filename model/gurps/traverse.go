// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import "slices"

type traversalData[T Node[T]] struct {
	list  []T
	index int
}

// Traverse calls the function 'f' for each node and its children in the input list, recursively. Return true from the
// function to abort early. If excludeContainers is true, then nodes that are containers will not be passed to 'f',
// although their children will still be processed as usual.
func Traverse[T Node[T]](f func(T) bool, onlyEnabled, excludeContainers bool, in ...T) {
	tracking := []*traversalData[T]{
		{
			list:  in,
			index: 0,
		},
	}
	for len(tracking) != 0 {
		current := tracking[len(tracking)-1]
		if current.index >= len(current.list) {
			tracking = tracking[:len(tracking)-1]
			continue
		}
		node := current.list[current.index]
		current.index++
		if !onlyEnabled || node.Enabled() {
			if (!excludeContainers || !node.Container()) && f(node) {
				return
			}
			if node.HasChildren() {
				tracking = append(tracking, &traversalData[T]{list: slices.Clone(node.NodeChildren())})
			}
		}
	}
}

// SetDataOwnerAll sets the data owner of each node in the list. Each node's SetDataOwner is responsible for reaching
// its own children, modifiers and weapons, so the list is not traversed.
func SetDataOwnerAll[T Node[T]](owner DataOwner, list []T) {
	for _, one := range list {
		one.SetDataOwner(owner)
	}
}

// forEachNode calls visit for every node in each of the lists, recursively. Disabled nodes and containers are
// included.
func forEachNode[T Node[T]](visit func(T), lists ...[]T) {
	for _, list := range lists {
		Traverse(func(node T) bool {
			visit(node)
			return false
		}, false, false, list...)
	}
}

// sourcedNode is the subset of Node that source syncing and hashing need, so that one walk can serve every node type,
// including the modifiers that hang off traits and equipment.
type sourcedNode interface {
	GetSource() Source
	SyncWithSource()
}

// forEachSourcedNode calls visit for every node in the provider's lists, recursively, along with the modifiers of each
// trait and piece of equipment. Disabled nodes and containers are included. Lists the provider doesn't have are nil
// and contribute nothing.
func forEachSourcedNode(provider ListProvider, visit func(sourcedNode)) {
	forEachNode(func(t *Trait) {
		visit(t)
		forEachNode(func(mod *TraitModifier) { visit(mod) }, t.Modifiers)
	}, provider.TraitList())
	forEachNode(func(s *Skill) { visit(s) }, provider.SkillList())
	forEachNode(func(s *Spell) { visit(s) }, provider.SpellList())
	forEachNode(func(eqp *Equipment) {
		visit(eqp)
		forEachNode(func(mod *EquipmentModifier) { visit(mod) }, eqp.Modifiers)
	}, provider.CarriedEquipmentList(), provider.OtherEquipmentList())
	forEachNode(func(n *Note) { visit(n) }, provider.NoteList())
}

// syncWithLibrarySources syncs every node the provider holds with its library source.
func syncWithLibrarySources(provider ListProvider) {
	forEachSourcedNode(provider, sourcedNode.SyncWithSource)
}

// countNodes returns the number of nodes in the list, descending into containers, that include accepts. A nil include
// accepts every node. When onlyEnabled is true, disabled nodes and their descendants are skipped, as they are by
// Traverse.
func countNodes[T Node[T]](list []T, onlyEnabled bool, include func(T) bool) int {
	count := 0
	Traverse(func(node T) bool {
		if include == nil || include(node) {
			count++
		}
		return false
	}, onlyEnabled, false, list...)
	return count
}
