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
	"encoding/json/v2"
	"testing"
	"testing/fstest"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/eqcontainer"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
)

// newTraitGroup returns a trait group holding the children, picked from separately when separately is true.
func newTraitGroup(name string, separately bool, children ...*Trait) *Trait {
	group := NewTrait(nil, nil, true)
	group.Name = name
	group.PickSeparately = separately
	group.Children = children
	for _, child := range children {
		child.SetParent(group)
	}
	return group
}

// newOrganizedChoice returns "Pick 1" of a, then b, c and d in organizing groups, the inner one nested, then e.
func newOrganizedChoice() (choice, outer, inner *Trait) {
	choice = newTemplateChoiceTrait("Choice", "a")
	inner = newTraitGroup("inner", true, NewTrait(nil, nil, false), NewTrait(nil, nil, false))
	inner.Children[0].Name, inner.Children[1].Name = "c", "d"
	b := NewTrait(nil, nil, false)
	b.Name = "b"
	outer = newTraitGroup("outer", true, b, inner)
	e := NewTrait(nil, choice, false)
	e.Name = "e"
	outer.SetParent(choice)
	choice.Children = append(choice.Children, outer, e)
	return choice, outer, inner
}

func TestPickSeparatelyPersistence(t *testing.T) {
	c := check.New(t)
	group := newTraitGroup("Group", false)
	data, err := json.Marshal(group)
	c.NoError(err)
	c.NotContains(string(data), "pick_separately", "the key is left out when off")
	hash := Hash64(group)

	group.PickSeparately = true
	c.Equal(hash, Hash64(group), "the flag isn't source data")
	data, err = json.Marshal(group)
	c.NoError(err)
	c.Contains(string(data), `"pick_separately":true`)
	var loaded Trait
	c.NoError(json.Unmarshal(data, &loaded))
	c.True(loaded.PickSeparately)
	c.True(group.Clone(LibraryFile{}, nil, nil, Copy).PickSeparately, "a clone keeps it")

	e := NewEntity()
	libFile := LibraryFile{Library: "Test Library", Path: "Test" + TraitsExt}
	source := newTraitGroup("Source", false)
	e.SourceMatcher().libHashes = map[LibraryFile]libSrcData{
		libFile: {dataHashes: map[tid.TID]HashAndData{source.TID: {Hash: Hash64(source), Data: source}}},
	}
	local := newTraitGroup("Local", true)
	local.SetDataOwner(e)
	local.Source = Source{LibraryFile: libFile, TID: source.TID}
	state, _ := e.SourceMatcher().Match(local)
	c.Equal(srcstate.Mismatched, state)
	local.SyncWithSource()
	c.Equal("Source", local.Name)
	c.True(local.PickSeparately, "syncing leaves it alone")
}

func TestPickSeparatelyClearing(t *testing.T) {
	c := check.New(t)
	choice, outer, inner := newOrganizedChoice()
	ClearTemplatePickerData(choice)
	c.False(outer.PickSeparately)
	c.False(inner.PickSeparately)

	choice, outer, inner = newOrganizedChoice()
	list, err := json.Marshal(&listData[*Trait]{Version: jio.CurrentDataVersion, Rows: []*Trait{choice}})
	c.NoError(err)
	rows, err := NewTraitsFromFile(fstest.MapFS{"list.adq": &fstest.MapFile{Data: list}}, "list.adq")
	c.NoError(err)
	c.False(rows[0].Children[1].PickSeparately, "a list doesn't keep it")

	outer.TemplatePicker = newTemplateChoicePicker()
	normalizeTemplateChoiceContainer(outer)
	c.False(outer.PickSeparately, "a choice is never picked from separately")

	inner.ContainerType = container.MetaTrait
	inner.ClearUnusedFieldsForType()
	c.False(inner.PickSeparately, "only a group can be")
	leaf := NewTrait(nil, nil, false)
	leaf.PickSeparately = true
	leaf.ClearUnusedFieldsForType()
	c.False(leaf.PickSeparately)
	skill := NewSkill(nil, nil, false)
	skill.PickSeparately = true
	skill.ClearUnusedFieldsForType()
	c.False(skill.PickSeparately)
	eqp := NewEquipment(nil, nil, true)
	eqp.PickSeparately = true
	eqp.ClearUnusedFieldsForType()
	c.False(eqp.PickSeparately, "a physical container is picked as a unit")
	group := NewEquipmentGroup(nil, nil)
	group.PickSeparately = true
	group.ClearUnusedFieldsForType()
	c.True(group.PickSeparately)
}

func TestPickSeparatelyHasEffect(t *testing.T) {
	c := check.New(t)
	choice, outer, inner := newOrganizedChoice()
	c.False(CanPickSeparately(choice))
	c.False(CanPickSeparately(choice.Children[0]))
	c.True(PickSeparatelyHasEffect(outer))
	c.True(IsOrganizingGroup(outer))
	c.True(IsOrganizingGroup(inner), "through an organizing group")

	outer.PickSeparately = false
	c.True(PickSeparatelyHasEffect(outer))
	c.False(IsOrganizingGroup(outer))
	c.False(PickSeparatelyHasEffect(inner), "not inside a unit group")
	c.False(IsOrganizingGroup(inner))

	inner.SetParent(nil)
	c.False(PickSeparatelyHasEffect(inner), "not outside a choice")
	c.False(IsOrganizingGroup(inner))

	c.True(CanPickSeparately(NewSkill(nil, nil, true)))
	c.False(CanPickSeparately(NewSkillChoiceContainer(nil, nil)))
	c.True(CanPickSeparately(NewSpell(nil, nil, true)))
	c.True(CanPickSeparately(NewEquipmentGroup(nil, nil)))
	c.False(CanPickSeparately(NewEquipment(nil, nil, true)))
	c.False(CanPickSeparately(NewNote(nil, nil, true)))
	var none *Trait
	c.False(IsOrganizingGroup(none))
	c.Equal(eqcontainer.Group, NewEquipmentChoiceContainer(nil, nil).ContainerType)
	c.False(CanPickSeparately(NewEquipmentChoiceContainer(nil, nil)))
}

func TestTemplateChoiceOptions(t *testing.T) {
	c := check.New(t)
	choice, outer, _ := newOrganizedChoice()
	c.Equal([]string{"a", "b", "c", "d", "e"}, traitNames(TemplateChoiceOptions(choice)))
	c.Equal([]string{"a", "outer", "e"}, traitNames(choice.Children), "the children are left alone")
	c.Equal([]string{"b", "inner"}, traitNames(TemplateChoiceOptions(outer)), "only a choice offers options")

	outer.PickSeparately = false
	options := TemplateChoiceOptions(choice)
	c.Equal([]string{"a", "outer", "e"}, traitNames(options))
	c.True(&options[0] == &choice.Children[0], "the children themselves come back")
	c.Equal(0.0, testing.AllocsPerRun(10, func() { TemplateChoiceOptions(choice) }))
}

func traitNames(traits []*Trait) []string {
	names := make([]string, len(traits))
	for i, one := range traits {
		names[i] = one.Name
	}
	return names
}
