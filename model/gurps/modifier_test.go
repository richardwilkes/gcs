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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
)

func TestMergeReplacements(t *testing.T) {
	c := check.New(t)

	// A nil destination takes the source map as-is.
	src := map[string]string{"Element": "Fire"}
	c.Equal(map[string]string{"Element": "Fire"}, mergeReplacements(nil, src))

	// Values the destination already holds win; new keys are added.
	dst := map[string]string{"Element": "Water"}
	c.Equal(map[string]string{"Element": "Water", "Target": "Orcs"},
		mergeReplacements(dst, map[string]string{"Element": "Fire", "Target": "Orcs"}))

	// An empty source leaves a nil destination nil rather than handing back an empty map.
	c.Nil(mergeReplacements(nil, map[string]string{}))
}

// TestModifierWithoutOwnerKeepsMarkersIntact verifies that a modifier not attached to a trait or piece of equipment
// displays its name and notes exactly as written, full marker syntax included, while an attached one resolves the
// markers against its owner's replacements.
func TestModifierWithoutOwnerKeepsMarkersIntact(t *testing.T) {
	c := check.New(t)
	const raw = "A @Element|Fire|Water@ spell"

	tm := NewTraitModifier(nil, nil, false)
	tm.Name = raw
	tm.LocalNotes = raw
	c.Equal(raw, tm.NameWithReplacements(), "an unattached trait modifier keeps its markers as written")
	c.Equal(raw, tm.LocalNotesWithReplacements(), "an unattached trait modifier keeps its markers as written")

	em := NewEquipmentModifier(nil, nil, false)
	em.Name = raw
	em.LocalNotes = raw
	c.Equal(raw, em.NameWithReplacements(), "an unattached equipment modifier keeps its markers as written")
	c.Equal(raw, em.LocalNotesWithReplacements(), "an unattached equipment modifier keeps its markers as written")

	// Once attached, the owner's replacements are applied and anything still unresolved is shown in compact form.
	trait := NewTrait(nil, nil, false)
	tm.setTrait(trait)
	c.Equal("A @Element@ spell", tm.NameWithReplacements(), "an owner without replacements yields the compact form")
	trait.Replacements = map[string]string{"Element|Fire|Water": "Fire"}
	c.Equal("A Fire spell", tm.NameWithReplacements())
	c.Equal("A Fire spell", tm.LocalNotesWithReplacements())

	equipment := NewEquipment(nil, nil, false)
	em.setEquipment(equipment)
	c.Equal("A @Element@ spell", em.NameWithReplacements(), "an owner without replacements yields the compact form")
	equipment.Replacements = map[string]string{"Element|Fire|Water": "Water"}
	c.Equal("A Water spell", em.NameWithReplacements())
	c.Equal("A Water spell", em.LocalNotesWithReplacements())
}

// TestModifierContainersContributeNameableKeys verifies that container modifiers take part in nameable key
// collection, since their names and notes are displayed with replacements applied, while disabled leaf modifiers
// still do not.
func TestModifierContainersContributeNameableKeys(t *testing.T) {
	c := check.New(t)

	trait := NewTrait(nil, nil, false)
	trait.Name = "Ally"
	group := NewTraitModifier(nil, nil, true)
	group.Name = "@Group@ Options"
	group.LocalNotes = "Applies to @Target@"
	leaf := NewTraitModifier(nil, group, false)
	leaf.Name = "@Leaf@"
	off := NewTraitModifier(nil, group, false)
	off.Name = "@Off@"
	off.Disabled = true
	group.Children = []*TraitModifier{leaf, off}
	trait.Modifiers = []*TraitModifier{group}
	keys := make(map[string]string)
	trait.FillWithNameableKeys(keys, nil)
	c.Equal(map[string]string{
		"Group":  nameable.Unset,
		"Target": nameable.Unset,
		"Leaf":   nameable.Unset,
	}, keys)

	equipment := NewEquipment(nil, nil, false)
	equipment.Name = "Sword"
	egroup := NewEquipmentModifier(nil, nil, true)
	egroup.Name = "@Group@ Options"
	egroup.LocalNotes = "Applies to @Target@"
	eleaf := NewEquipmentModifier(nil, egroup, false)
	eleaf.Name = "@Leaf@"
	eoff := NewEquipmentModifier(nil, egroup, false)
	eoff.Name = "@Off@"
	eoff.Disabled = true
	egroup.Children = []*EquipmentModifier{eleaf, eoff}
	equipment.Modifiers = []*EquipmentModifier{egroup}
	keys = make(map[string]string)
	equipment.FillWithNameableKeys(keys, nil)
	c.Equal(map[string]string{
		"Group":  nameable.Unset,
		"Target": nameable.Unset,
		"Leaf":   nameable.Unset,
	}, keys)
}

// TestModifierHashDistinguishesContainers verifies that a container and a leaf modifier that share all of their
// common sync data still hash differently for both modifier types.
func TestModifierHashDistinguishesContainers(t *testing.T) {
	c := check.New(t)

	tmContainer := NewTraitModifier(nil, nil, true)
	tmLeaf := NewTraitModifier(nil, nil, false)
	tmContainer.Name = "Same"
	tmLeaf.Name = "Same"
	c.NotEqual(Hash64(tmContainer), Hash64(tmLeaf))

	emContainer := NewEquipmentModifier(nil, nil, true)
	emLeaf := NewEquipmentModifier(nil, nil, false)
	emContainer.Name = "Same"
	emLeaf.Name = "Same"
	c.NotEqual(Hash64(emContainer), Hash64(emLeaf))
}

// TestModifierSyncWithSource verifies that a modifier which has drifted from its library source pulls the synced
// fields back across, that one without a data owner is left alone, and that containers do not pick up leaf-only data.
func TestModifierSyncWithSource(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	libFile := LibraryFile{Library: "Test Library", Path: "Test" + TraitModifiersExt}
	install := func(source SrcProvider, id tid.TID) {
		e.SourceMatcher().libHashes = map[LibraryFile]libSrcData{
			libFile: {dataHashes: map[tid.TID]HashAndData{id: {Hash: Hash64(source), Data: source}}},
		}
	}

	// A leaf trait modifier whose local copy has drifted in name and cost.
	source := NewTraitModifier(nil, nil, false)
	source.Name = "Reduced Time"
	source.CostAdj = "+20%"
	source.Tags = []string{"Enhancement"}
	local := NewTraitModifier(e, nil, false)
	local.Name = "Reduced Time (old)"
	local.CostAdj = "+10%"
	local.Source = Source{LibraryFile: libFile, TID: source.TID}
	install(source, source.TID)
	state, _ := e.SourceMatcher().Match(local)
	c.Equal(srcstate.Mismatched, state, "precondition: the local copy differs from its source")
	local.SyncWithSource()
	c.Equal("Reduced Time", local.Name)
	c.Equal("+20%", local.CostAdj)
	c.Equal([]string{"Enhancement"}, local.Tags)
	source.Tags[0] = "changed"
	c.Equal([]string{"Enhancement"}, local.Tags, "synced tags must not alias the source's slice")
	state, _ = e.SourceMatcher().Match(local)
	c.Equal(srcstate.Matched, state, "after syncing, the local copy matches its source")

	// Without a data owner there is nothing to look the source up in, so nothing changes.
	orphan := NewTraitModifier(nil, nil, false)
	orphan.Name = "Reduced Time (old)"
	orphan.Source = Source{LibraryFile: libFile, TID: source.TID}
	orphan.SyncWithSource()
	c.Equal("Reduced Time (old)", orphan.Name)

	// A container only picks up the common data.
	sourceContainer := NewEquipmentModifier(nil, nil, true)
	sourceContainer.Name = "Options"
	sourceContainer.PageRef = "B1"
	localContainer := NewEquipmentModifier(e, nil, true)
	localContainer.Name = "Options (old)"
	localContainer.Source = Source{LibraryFile: libFile, TID: sourceContainer.TID}
	install(sourceContainer, sourceContainer.TID)
	state, _ = e.SourceMatcher().Match(localContainer)
	c.Equal(srcstate.Mismatched, state, "precondition: the local container differs from its source")
	localContainer.SyncWithSource()
	c.Equal("Options", localContainer.Name)
	c.Equal("B1", localContainer.PageRef)
	c.Equal("", localContainer.CostAmount, "a container carries no leaf-only data")
}
