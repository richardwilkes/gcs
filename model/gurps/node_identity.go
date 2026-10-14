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
	"reflect"

	"github.com/richardwilkes/toolbox/v2/tid"
)

// NodeIdentities maps node IDs to the objects that carry them, so that a list put back from its serialized form can
// reuse the objects the data was taken from (see RecycleNodes). It holds the modifiers and weapons nested in each node
// along with the nodes themselves. An ID that more than one node carries maps to sharedID instead of an object, since
// there is no telling which of those nodes an object came from.
type NodeIdentities map[tid.TID]any

// sharedID is the entry for an ID that more than one node carries, which recycling leaves alone.
type sharedID struct{}

// CollectNodeIdentities adds the nodes in the list, their descendants, and the modifiers and weapons nested in them to
// the identities. A node whose ID the identities already hold from an earlier call replaces that object, so that
// collecting a list over a stale set of identities brings it up to date. An ID that more than one node in the list
// carries is marked shared, and stays shared through later calls, since nothing could tell which object each of those
// nodes came from.
func CollectNodeIdentities[T Node[T]](identities NodeIdentities, list []T) {
	collectNodeIdentities(identities, make(map[tid.TID]bool), list)
}

// collectNodeIdentities is CollectNodeIdentities with the IDs the call has already visited, which spans the nested
// lists so that a modifier or weapon sharing a node's ID is caught too.
func collectNodeIdentities[T Node[T]](identities NodeIdentities, seen map[tid.TID]bool, list []T) {
	Traverse(func(node T) bool {
		id := node.ID()
		if _, shared := identities[id].(sharedID); !shared {
			if seen[id] {
				identities[id] = sharedID{}
			} else {
				identities[id] = node
			}
		}
		seen[id] = true
		nestedNodeLists(node,
			func(_ *Equipment, modifiers *[]*EquipmentModifier) {
				collectNodeIdentities(identities, seen, *modifiers)
			},
			func(_ *Trait, modifiers *[]*TraitModifier) {
				collectNodeIdentities(identities, seen, *modifiers)
			},
			func(_ WeaponOwner, weapons *[]*Weapon) {
				collectNodeIdentities(identities, seen, *weapons)
			})
		return false
	}, false, false, list...)
}

// RecycleNodes returns the list with each node the identities hold an object for replaced by that object, which takes
// on the node's data. Deserializing a list makes a fresh object for every node, which leaves anything still holding the
// old one -- an undo edit, an open editor -- changing an object no list contains; recycling keeps the object a node
// has always had, so those keep working. Nodes the identities have no object for, or whose ID they mark as shared, are
// kept as they are. Each object is used at most once: its entry is removed from the identities as it is used, so that a
// later node with the same ID keeps the object it came with instead of collapsing into the one already recycled. The
// parents of the children, the targets of the modifiers and the owners of the weapons are set to the objects that hold
// them. The data owners of the nodes come from the deserialized nodes, since their data is copied whole, so they are
// whatever the caller attached when it put the list in place, and the modifiers take theirs from the nodes that hold
// them; a template's and a loot sheet's list setters attach nothing, which the rebuild that follows a restore makes
// good (see EnsureAttachments).
func RecycleNodes[T Node[T]](identities NodeIdentities, list []T) []T {
	for i, node := range list {
		list[i] = recycleNode(identities, node)
	}
	return list
}

func recycleNode[T Node[T]](identities NodeIdentities, node T) T {
	id := node.ID()
	if existing, ok := identities[id].(T); ok {
		delete(identities, id)
		if existing != node {
			reflect.ValueOf(existing).Elem().Set(reflect.ValueOf(node).Elem())
			node = existing
		}
	}
	if children := node.NodeChildren(); len(children) != 0 {
		children = RecycleNodes(identities, children)
		for _, child := range children {
			child.SetParent(node)
		}
		node.SetChildren(children)
	}
	// The nested lists are attached again whether or not the node itself was recycled: a recycled modifier or weapon
	// takes on the deserialized one's data, back-pointers included, so it would otherwise target or belong to the
	// deserialized node, and a recycled weapon's damage would belong to the deserialized weapon.
	nestedNodeLists(node,
		func(target *Equipment, modifiers *[]*EquipmentModifier) {
			*modifiers = RecycleNodes(identities, *modifiers)
			AttachModifiers(target, *modifiers)
		},
		func(target *Trait, modifiers *[]*TraitModifier) {
			*modifiers = RecycleNodes(identities, *modifiers)
			AttachModifiers(target, *modifiers)
		},
		func(owner WeaponOwner, weapons *[]*Weapon) {
			*weapons = RecycleNodes(identities, *weapons)
			for _, w := range *weapons {
				w.SetOwner(owner)
			}
		})
	return node
}

// nestedNodeLists calls the function for each list of nodes nested in the node, along with the node that holds it: the
// modifiers of a trait or a piece of equipment, and the weapons of a trait, skill, spell or piece of equipment.
func nestedNodeLists(node any, equipmentModifiers func(*Equipment, *[]*EquipmentModifier),
	traitModifiers func(*Trait, *[]*TraitModifier), weapons func(WeaponOwner, *[]*Weapon),
) {
	switch n := node.(type) {
	case *Equipment:
		equipmentModifiers(n, &n.Modifiers)
		weapons(n, &n.Weapons)
	case *Trait:
		traitModifiers(n, &n.Modifiers)
		weapons(n, &n.Weapons)
	case *Skill:
		weapons(n, &n.Weapons)
	case *Spell:
		weapons(n, &n.Weapons)
	}
}
