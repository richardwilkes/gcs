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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/frequency"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestTemplatePickerLoadsUnknownTypeAsNotApplicable verifies what lets the rest of the code read a picker's type
// without checking it first: a type this version doesn't know, as a newer or hand-edited file might hold, loads as
// NotApplicable, leaving the container with no picker at all rather than one with a type nothing can handle.
func TestTemplatePickerLoadsUnknownTypeAsNotApplicable(t *testing.T) {
	c := check.New(t)
	var tp TemplatePicker
	c.NoError(json.Unmarshal([]byte(`{"type":"not_a_real_picker_type"}`), &tp))
	c.Equal(picker.NotApplicable, tp.Type)
	c.True(tp.IsZero())
}

// TestHasTemplatePickerData verifies that a node carrying template choices is spotted
func TestHasTemplatePickerData(t *testing.T) {
	c := check.New(t)
	plain := NewTrait(nil, nil, false)
	plain.Name = "Claws"
	c.False(HasTemplatePickerData(plain), "a plain trait carries no choices")

	emptyContainer := NewTrait(nil, nil, true)
	emptyContainer.Name = "Advantages"
	c.False(HasTemplatePickerData(emptyContainer), "a container without choices carries none")

	choices := newTemplateChoiceTrait("Pick One", "First", "Second")
	c.True(HasTemplatePickerData(plain, choices))

	emptyContainer.Children = []*Trait{choices}
	choices.SetParent(emptyContainer)
	c.True(HasTemplatePickerData(emptyContainer), "choices nested deeper must be found as well")

	skillContainer := NewSkill(nil, nil, true)
	skillContainer.Name = "Techniques"
	c.False(HasTemplatePickerData(skillContainer))

	skillContainer.TemplatePicker.Type = picker.Points
	c.True(HasTemplatePickerData(skillContainer), "skills carry choices too")
}

// TestClearTemplatePickerData verifies that the choices are removed from every container beneath the rows as well, since
// a copied container brings its whole subtree with it.
func TestClearTemplatePickerData(t *testing.T) {
	c := check.New(t)
	outer := NewTrait(nil, nil, true)
	outer.Name = "Advantages"
	inner := newTemplateChoiceTrait("Pick One", "First", "Second")
	inner.SetParent(outer)
	outer.Children = []*Trait{inner}

	// Verify the test data has picker data
	c.True(HasTemplatePickerData(outer))

	// Clear any template picker data
	ClearTemplatePickerData(outer)

	// Verify the test data no longer carries picker data
	c.False(HasTemplatePickerData(outer), "no picker data may be left")

	// Verify we still have the same data otherwise (this could use more checks)
	c.Equal(2, len(inner.Children), "clearing must not disturb anything else")
}

// TestClearTemplatePickerDataClearsSource verifies that a container losing its choices also loses its source, while
// the rest of the subtree keeps theirs. Only a template may hold choices and a template is never a source, so the source
// a container with choices points at can't have them, and syncing with it would quietly take them away.
func TestClearTemplatePickerDataClearsSource(t *testing.T) {
	c := check.New(t)
	outer := NewTrait(nil, nil, true)
	outer.Name = "Advantages"
	outer.Source = Source{Library: "lib", Path: "outer.adq", TID: outer.ID()}
	inner := newTemplateChoiceTrait("Pick One", "First", "Second")
	inner.Source = Source{Library: "lib", Path: "inner.adq", TID: inner.ID()}
	inner.SetParent(outer)
	outer.Children = []*Trait{inner}
	child := inner.Children[0]
	child.Source = Source{Library: "lib", Path: "child.adq", TID: child.ID()}

	ClearTemplatePickerData(outer)

	c.Equal(Source{}, inner.Source, "the container that lost its choices must lose its source as well")
	c.NotEqual(Source{}, outer.Source, "a container that had no choices must keep its source")
	c.NotEqual(Source{}, child.Source, "an option must keep its source")
}

// newTemplateChoiceTrait returns a trait container carrying template choices, holding a child for each of the given
// names.
func newTemplateChoiceTrait(name string, childNames ...string) *Trait {
	group := NewTrait(nil, nil, true)
	group.Name = name
	group.TemplatePicker.Type = picker.Count
	group.TemplatePicker.Qualifier.Compare = criteria.EqualsNumber
	group.TemplatePicker.Qualifier.Qualifier = fxp.One
	children := make([]*Trait, 0, len(childNames))
	for _, childName := range childNames {
		child := NewTrait(nil, group, false)
		child.Name = childName
		children = append(children, child)
	}
	group.Children = children
	return group
}

// TestNewChoiceContainers verifies that each kind of choice container is created as a container that holds choices,
// asking for exactly one of its children, and named after what it is.
func TestNewChoiceContainers(t *testing.T) {
	c := check.New(t)
	trait := NewTraitChoiceContainer(nil, nil)
	c.True(IsTemplateChoiceContainer(trait))
	c.Equal("Pick 1", trait.TemplatePicker.String())
	c.Equal("Trait Choice", trait.Name)
	c.Equal(container.Group, trait.ContainerType)

	skill := NewSkillChoiceContainer(nil, nil)
	c.True(IsTemplateChoiceContainer(skill))
	c.Equal("Skill Choice", skill.Name)

	spell := NewSpellChoiceContainer(nil, nil)
	c.True(IsTemplateChoiceContainer(spell))
	c.Equal("Spell Choice", spell.Name)

	c.False(IsTemplateChoiceContainer(NewTrait(nil, nil, true)), "a plain container holds no choices")
	plain := NewTrait(nil, nil, false)
	plain.TemplatePicker.Type = picker.Count
	c.False(IsTemplateChoiceContainer(plain), "only a container may hold choices")
}

// TestTemplateLoadNormalizesChoiceContainers verifies that loading a template turns a choice container of any other
// type into a plain group, dropping what only that type used, and removes its modifiers, while leaving every other
// container's type and modifiers alone.
func TestTemplateLoadNormalizesChoiceContainers(t *testing.T) {
	c := check.New(t)
	outer := NewTrait(nil, nil, true)
	outer.Name = "Lens"
	outer.ContainerType = container.MetaTrait
	choice := newTemplateChoiceTrait("Pick One", "First", "Second")
	choice.ContainerType = container.Ancestry
	choice.Ancestry = "Human"
	choice.Modifiers = []*TraitModifier{NewTraitModifier(nil, nil, false)}
	choice.VTTNotes = "vtt"
	choice.UserDesc = "desc"
	choice.Tags = []string{"Advantage"}
	choice.SelfControl = selfctrl.CR12
	choice.Frequency = frequency.FR9
	choice.Disabled = true
	choice.Prereq = NewPrereqList()
	choice.Prereq.Prereqs = append(choice.Prereq.Prereqs, NewTraitPrereq())
	choice.Source = Source{Library: "lib", Path: "choice.adq", TID: choice.ID()}
	choice.SetParent(outer)
	outer.Modifiers = []*TraitModifier{NewTraitModifier(nil, nil, false)}
	outer.Tags = []string{"Lens"}
	outer.Children = []*Trait{choice}
	alternatives := newTemplateChoiceTrait("Pick Another", "Third", "Fourth")
	alternatives.ContainerType = container.AlternativeAbilities
	alternatives.AlternativeSlots = 2
	abilities := NewTrait(nil, nil, true)
	abilities.Name = "Abilities"
	abilities.ContainerType = container.AlternativeAbilities
	abilities.AlternativeSlots = 2
	tmpl := NewTemplate()
	tmpl.Traits = []*Trait{outer, alternatives, abilities}

	data, err := json.Marshal(tmpl)
	c.NoError(err)
	var loaded Template
	c.NoError(json.Unmarshal(data, &loaded))
	c.Equal(3, len(loaded.Traits))

	c.Equal(container.MetaTrait, loaded.Traits[0].ContainerType, "a container without choices keeps its type")
	c.Equal(1, len(loaded.Traits[0].Modifiers), "a container without choices keeps its modifiers")
	c.Equal([]string{"Lens"}, loaded.Traits[0].Tags, "a container without choices keeps its tags")
	loadedChoice := loaded.Traits[0].Children[0]
	c.Equal(0, len(loadedChoice.Modifiers), "a choice container must not keep modifiers")
	c.Equal("", loadedChoice.VTTNotes, "a choice container must not keep VTT notes")
	c.Equal("", loadedChoice.UserDesc, "a choice container must not keep a user description")
	c.Equal(0, len(loadedChoice.Tags), "a choice container must not keep tags")
	c.Equal(selfctrl.None, loadedChoice.SelfControl, "a choice container must not keep a self-control roll")
	c.Equal(frequency.None, loadedChoice.Frequency, "a choice container must not keep a frequency")
	c.False(loadedChoice.Disabled, "a choice container must not stay disabled")
	c.True(loadedChoice.Prereq.IsZero(), "a choice container must not keep prerequisites")
	c.True(loadedChoice.Source.IsZero(), "a choice container must not keep a source")
	c.Equal(container.Group, loadedChoice.ContainerType, "a nested choice container must become a group")
	c.Equal("", loadedChoice.Ancestry, "a choice container no longer an ancestry must not keep one")
	c.True(IsTemplateChoiceContainer(loadedChoice), "the choices themselves must survive")

	c.Equal(container.Group, loaded.Traits[1].ContainerType, "a choice container must become a group")
	c.Equal(0, loaded.Traits[1].AlternativeSlots, "a choice container no longer holding alternatives must not keep slots")

	c.Equal(container.AlternativeAbilities, loaded.Traits[2].ContainerType, "a container without choices keeps its type")
	c.Equal(2, loaded.Traits[2].AlternativeSlots)
}

// TestTemplateChoiceConversion verifies which containers may become choice containers, what the conversion reports it
// will lose, and that converting back only removes the choice.
func TestTemplateChoiceConversion(t *testing.T) {
	c := check.New(t)
	meta := NewTrait(nil, nil, true)
	meta.ContainerType = container.MetaTrait
	c.False(CanConvertToTemplateChoiceContainer(meta), "only a group may become a choice container")
	c.False(CanConvertToTemplateChoiceContainer(NewTrait(nil, nil, false)), "only a container may become one")
	c.False(CanConvertToTemplateChoiceContainer(NewTraitChoiceContainer(nil, nil)), "a choice container already is one")
	c.True(CanConvertToTemplateChoiceContainer(NewSkill(nil, nil, true)), "any skill container may become one")

	group := NewTrait(nil, nil, true)
	c.Equal(0, len(TemplateChoiceConversionLosses(group)), "a bare group loses nothing")
	group.Frequency = frequency.FR9
	group.Preconfigured = true
	group.Disabled = true
	group.SwitchedOn = true
	group.Prereq = NewPrereqList()
	group.Prereq.Prereqs = append(group.Prereq.Prereqs, NewTraitPrereq())
	group.Source = Source{Library: "lib", Path: "group.adq", TID: group.ID()}
	c.Equal(6, len(TemplateChoiceConversionLosses(group)))

	ConvertToTemplateChoiceContainer(group)
	c.True(IsTemplateChoiceContainer(group))
	c.Equal("Pick 1", group.TemplatePicker.String())
	c.Equal(frequency.None, group.Frequency)
	c.False(group.Preconfigured)
	c.False(group.Disabled)
	c.False(group.SwitchedOn)
	c.True(group.Prereq.IsZero())
	c.Equal(Source{}, group.Source)

	ConvertFromTemplateChoiceContainer(group)
	c.False(IsTemplateChoiceContainer(group))
	c.True(group.Container(), "it must still be a container")
	c.Equal(container.Group, group.ContainerType)
}

// TestSkillChoiceConversionLosses verifies that a skill choice container can't keep VTT notes or tags either.
func TestSkillChoiceConversionLosses(t *testing.T) {
	c := check.New(t)
	group := NewSkill(nil, nil, true)
	c.Equal(0, len(TemplateChoiceConversionLosses(group)), "a bare skill container loses nothing")
	group.VTTNotes = "vtt"
	group.Tags = []string{"Combat"}
	c.Equal(2, len(TemplateChoiceConversionLosses(group)))
	ConvertToTemplateChoiceContainer(group)
	c.True(IsTemplateChoiceContainer(group))
	c.Equal("", group.VTTNotes)
	c.Equal(0, len(group.Tags))
}

// TestLoadingOutsideATemplateClearsPickerData verifies that only a template keeps template picker data when loaded: a
// character sheet and a standalone list both have it removed, along with the source of each container that had it.
func TestLoadingOutsideATemplateClearsPickerData(t *testing.T) {
	c := check.New(t)
	choices := newTemplateChoiceTrait("Pick One", "First", "Second")
	choices.Source = Source{Library: "lib", Path: "choices.adq", TID: choices.ID()}

	entity := NewEntity()
	entity.Traits = []*Trait{choices}
	data, err := json.Marshal(entity)
	c.NoError(err)
	var loadedEntity Entity
	c.NoError(json.Unmarshal(data, &loadedEntity))
	c.Equal(1, len(loadedEntity.Traits))
	c.False(HasTemplatePickerData(loadedEntity.Traits...), "a character sheet must not keep picker data")
	c.True(loadedEntity.Traits[0].Source.IsZero(), "the container that lost its picker data must lose its source")
	c.Equal(2, len(loadedEntity.Traits[0].Children), "the options must be left alone")

	list, err := json.Marshal(&listData[*Trait]{Version: jio.CurrentDataVersion, Rows: []*Trait{choices}})
	c.NoError(err)
	rows, err := NewTraitsFromFile(fstest.MapFS{"list.adq": &fstest.MapFile{Data: list}}, "list.adq")
	c.NoError(err)
	c.Equal(1, len(rows))
	c.False(HasTemplatePickerData(rows...), "a standalone list must not keep picker data")
}
