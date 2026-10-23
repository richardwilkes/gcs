// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// roundTripEquipment returns the list as deserializing its serialized form yields it: the same data in fresh objects.
func roundTripEquipment(c check.Checker, list []*gurps.Equipment) []*gurps.Equipment {
	c.Helper()
	data, err := jio.SerializeAndCompress(list)
	c.NoError(err)
	var fresh []*gurps.Equipment
	c.NoError(jio.DecompressAndDeserialize(data, &fresh))
	return fresh
}

func TestRecycleNodesReusesTheObjectsTheDataCameFrom(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	backpack := gurps.NewEquipment(entity, nil, true)
	backpack.Name = "Backpack"
	rope := gurps.NewEquipment(entity, backpack, false)
	rope.Name = "Rope"
	whip := gurps.NewWeapon(rope, true)
	whip.Usage = "Whip"
	rope.Weapons = []*gurps.Weapon{whip}
	fine := gurps.NewEquipmentModifier(entity, nil, true)
	fine.Name = "Quality"
	veryFine := gurps.NewEquipmentModifier(entity, fine, false)
	veryFine.Name = "Very Fine"
	fine.Children = []*gurps.EquipmentModifier{veryFine}
	rope.Modifiers = []*gurps.EquipmentModifier{fine}
	gurps.AttachModifiers(rope, rope.Modifiers)
	backpack.Children = []*gurps.Equipment{rope}
	list := []*gurps.Equipment{backpack}

	identities := make(gurps.NodeIdentities)
	gurps.CollectNodeIdentities(identities, list)
	c.Equal(5, len(identities), "the container, its child, the child's weapon and both modifiers")

	fresh := roundTripEquipment(c, list)
	c.True(fresh[0] != backpack, "deserializing makes fresh objects")
	fresh[0].Name = "Rucksack"
	fresh[0].Children[0].Name = "Cord"
	fresh[0].Children[0].Weapons[0].Usage = "Lash"
	fresh[0].Children[0].Modifiers[0].Children[0].Name = "Fine"

	recycled := gurps.RecycleNodes(identities, fresh)
	c.True(recycled[0] == backpack, "the container must come back as the object it was")
	c.Equal("Rucksack", backpack.Name, "carrying the deserialized data")
	c.True(backpack.Children[0] == rope, "the child must come back as the object it was")
	c.Equal("Cord", rope.Name)
	c.True(rope.Parent() == backpack, "the child's parent must be the recycled container")
	c.True(rope.Weapons[0] == whip, "the weapon must come back as the object it was")
	c.Equal("Lash", whip.Usage)
	c.True(rope.Modifiers[0] == fine, "the modifier must come back as the object it was")
	c.True(fine.Children[0] == veryFine, "the nested modifier must come back as the object it was")
	c.Equal("Fine", veryFine.Name)
	c.True(veryFine.Parent() == fine, "the nested modifier's parent must be the recycled modifier")
}

// TestRecycleNodesKeepsNodesSharingAnIDDistinct verifies that nodes whose ID the identities hold more than once are
// kept as deserialized, there being no telling which object each came from, whether a collection holding the ID only
// once comes before or after the one that finds it shared.
func TestRecycleNodesKeepsNodesSharingAnIDDistinct(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	first := gurps.NewEquipment(entity, nil, false)
	first.Name = "First"
	second := gurps.NewEquipment(entity, nil, false)
	second.Name = "Second"
	second.TID = first.TID
	list := []*gurps.Equipment{first, second}
	identities := make(gurps.NodeIdentities)
	gurps.CollectNodeIdentities(identities, list)

	fresh := roundTripEquipment(c, list)
	freshFirst, freshSecond := fresh[0], fresh[1]
	recycled := gurps.RecycleNodes(identities, fresh)
	c.Equal(2, len(recycled))
	c.True(recycled[0] == freshFirst,
		"a node sharing its ID must be left as it is, there being no telling which object it came from")
	c.True(recycled[1] == freshSecond, "and so must the other")
	c.Equal("First", recycled[0].Name, "the first node must keep its data")
	c.Equal("Second", recycled[1].Name, "the second node must keep its data")
	c.Equal("First", first.Name, "the objects that could not be told apart must be left untouched")
	c.Equal("Second", second.Name)

	identities = make(gurps.NodeIdentities)
	gurps.CollectNodeIdentities(identities, list)
	only := gurps.NewEquipment(entity, nil, false)
	only.Name = "Only"
	only.TID = first.TID
	gurps.CollectNodeIdentities(identities, []*gurps.Equipment{only})
	fresh = roundTripEquipment(c, list)
	freshFirst, freshSecond = fresh[0], fresh[1]
	recycled = gurps.RecycleNodes(identities, fresh)
	c.True(recycled[0] == freshFirst, "an ID shared when collected must stay shared when a later collection holds it once")
	c.True(recycled[1] == freshSecond)
	c.Equal("Only", only.Name, "the object the later collection held must be left untouched")

	identities = make(gurps.NodeIdentities)
	gurps.CollectNodeIdentities(identities, []*gurps.Equipment{only})
	gurps.CollectNodeIdentities(identities, list)
	fresh = roundTripEquipment(c, []*gurps.Equipment{only})
	freshOnly := fresh[0]
	recycled = gurps.RecycleNodes(identities, fresh)
	c.True(recycled[0] == freshOnly, "an ID a later collection finds shared must not be recycled either")
	c.Equal("Only", only.Name)
}

func TestRecycleNodesPointsTheNestedModifiersAndWeaponsAtTheObjectsThatHoldThem(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	rope := gurps.NewEquipment(entity, nil, false)
	rope.Name = "Rope"
	whip := gurps.NewWeapon(rope, true)
	whip.Usage = "Whip"
	rope.Weapons = []*gurps.Weapon{whip}
	fine := gurps.NewEquipmentModifier(entity, nil, true)
	fine.Name = "Quality"
	veryFine := gurps.NewEquipmentModifier(entity, fine, false)
	veryFine.Name = "Very Fine"
	fine.Children = []*gurps.EquipmentModifier{veryFine}
	rope.Modifiers = []*gurps.EquipmentModifier{fine}
	gurps.AttachModifiers(rope, rope.Modifiers)
	list := []*gurps.Equipment{rope}

	// Nothing attached to the fresh list, as a template's or a loot sheet's list setter leaves it.
	identities := make(gurps.NodeIdentities)
	gurps.CollectNodeIdentities(identities, list)
	recycled := gurps.RecycleNodes(identities, roundTripEquipment(c, list))
	c.True(recycled[0] == rope)
	c.True(whip.Owner == rope, "the weapon must belong to the object holding it, not the deserialized node")
	c.True(whip.Damage.Owner == whip, "the damage must belong to the recycled weapon, not the deserialized one")
	c.True(fine.Target() == rope, "the modifier must target the object holding it, not the deserialized node")
	c.True(veryFine.Target() == rope, "and so must its child")

	// The fresh list attached to an entity before recycling, as an entity's list setter leaves it.
	identities = make(gurps.NodeIdentities)
	gurps.CollectNodeIdentities(identities, list)
	fresh := roundTripEquipment(c, list)
	entity.SetOtherEquipmentList(fresh)
	recycled = gurps.RecycleNodes(identities, fresh)
	c.True(recycled[0] == rope)
	c.True(rope.DataOwner() == entity, "the node's data owner is the one the caller attached")
	c.True(whip.Owner == rope, "the weapon must belong to the recycled node, not the discarded deserialized one")
	c.True(whip.Damage.Owner == whip)
	c.True(fine.Target() == rope, "the modifier must target the recycled node, not the discarded deserialized one")
	c.True(veryFine.Target() == rope)
	c.True(fine.DataOwner() == entity, "the modifier's data owner is the node's")
	c.True(veryFine.DataOwner() == entity)

	// A node that is kept as it is, with a weapon and a modifier that are recycled into it.
	identities = make(gurps.NodeIdentities)
	gurps.CollectNodeIdentities(identities, list)
	fresh = roundTripEquipment(c, list)
	freshRope := fresh[0]
	freshRope.TID = gurps.NewEquipment(entity, nil, false).TID
	recycled = gurps.RecycleNodes(identities, fresh)
	c.True(recycled[0] == freshRope, "a node the identities have no object for must be kept as it is")
	c.True(freshRope.Weapons[0] == whip, "while its weapon is recycled")
	c.True(whip.Owner == freshRope, "and belongs to the node that holds it")
	c.True(whip.Damage.Owner == whip, "with its damage belonging to it, not to the deserialized weapon")
	c.True(freshRope.Modifiers[0] == fine, "and its modifier is recycled")
	c.True(fine.Target() == freshRope, "and targets the node that holds it")
	c.True(veryFine.Target() == freshRope)
}
