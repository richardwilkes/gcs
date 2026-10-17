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
	"os"
	"path/filepath"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/eqcontainer"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// includingDisabled returns the condition with its IncludeDisabled set.
func includingDisabled(cond *gurps.FilterCondition) *gurps.FilterCondition {
	cond.IncludeDisabled = true
	return cond
}

// TestFilterModifierFieldsOnTraits checks the fields a trait filter draws from the trait's modifiers: the enabled ones
// only, unless a condition includes disabled ones, with containers and blank values left out either way. Notes are
// compared with the whole of a condition's text, commas and all.
func TestFilterModifierFieldsOnTraits(t *testing.T) {
	c := check.New(t)
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Innate Attack"
	accurate := gurps.NewTraitModifier(nil, nil, false)
	accurate.Name = "Accurate"
	accurate.LocalNotes = "Sharp"
	accurate.Tags = []string{"Enhancement", " "}
	cheap := gurps.NewTraitModifier(nil, nil, false)
	cheap.Name = "Cheap"
	cheap.LocalNotes = "Brittle"
	cheap.Tags = []string{"Limitation", ""}
	cheap.SetDisabled(true)
	group := gurps.NewTraitModifier(nil, nil, true)
	group.Name = "Group"
	nested := gurps.NewTraitModifier(nil, group, false)
	nested.Name = "Nested"
	nested.LocalNotes = "Melee, Ranged"
	group.Children = []*gurps.TraitModifier{nested}
	trait.Modifiers = []*gurps.TraitModifier{accurate, cheap, group}
	c.True(accurate.Enabled(), "precondition: the accurate modifier is enabled")
	c.False(cheap.Enabled(), "precondition: the cheap modifier is disabled")
	c.True(group.Container(), "precondition: the group modifier is a container")
	c.True(nested.Enabled(), "precondition: the nested modifier is enabled")

	fields := gurps.TraitFilterFields()
	for _, one := range []struct {
		name string
		cond *gurps.FilterCondition
		want bool
	}{
		{"enabled name", newTextFilterCondition("modifier_names", criteria.IsText, "Accurate"), true},
		{"disabled name", newTextFilterCondition("modifier_names", criteria.IsText, "Cheap"), false},
		{
			"disabled name, including disabled",
			includingDisabled(newTextFilterCondition("modifier_names", criteria.IsText, "Cheap")), true,
		},
		{"nested name", newTextFilterCondition("modifier_names", criteria.IsText, "Nested"), true},
		{"container name", newTextFilterCondition("modifier_names", criteria.IsText, "Group"), false},
		{
			"container name, including disabled",
			includingDisabled(newTextFilterCondition("modifier_names", criteria.IsText, "Group")), false,
		},
		{"enabled notes", newTextFilterCondition("modifier_notes", criteria.ContainsText, "sharp"), true},
		{"notes with a comma", newTextFilterCondition("modifier_notes", criteria.IsText, "Melee, Ranged"), true},
		{
			"notes contain a comma", newTextFilterCondition("modifier_notes", criteria.ContainsText, "Melee, Ranged"),
			true,
		},
		{
			"notes contain a part only", newTextFilterCondition("modifier_notes", criteria.ContainsText, "Melee, Thrown"),
			false,
		},
		{
			"notes don't contain a comma",
			newTextFilterCondition("modifier_notes", criteria.DoesNotContainText, "Melee, Ranged"), false,
		},
		{"disabled notes", newTextFilterCondition("modifier_notes", criteria.ContainsText, "brittle"), false},
		{
			"disabled notes, including disabled",
			includingDisabled(newTextFilterCondition("modifier_notes", criteria.ContainsText, "brittle")), true,
		},
		{"enabled tag", newTextFilterCondition("modifier_tags", criteria.IsText, "Enhancement"), true},
		{"disabled tag", newTextFilterCondition("modifier_tags", criteria.IsText, "Limitation"), false},
		{
			"disabled tag, including disabled",
			includingDisabled(newTextFilterCondition("modifier_tags", criteria.IsText, "Limitation")), true,
		},
		{"blank tag", newTextFilterCondition("modifier_tags", criteria.IsText, ""), false},
		{
			"blank tag, including disabled",
			includingDisabled(newTextFilterCondition("modifier_tags", criteria.IsText, "")), false,
		},
		{"none is disabled tag", newTextFilterCondition("modifier_tags", criteria.IsNotText, "Limitation"), true},
		{
			"none is disabled tag, including disabled",
			includingDisabled(newTextFilterCondition("modifier_tags", criteria.IsNotText, "Limitation")), false,
		},
	} {
		c.Equal(one.want, matchesListFilter(newTestListFilter(one.cond), fields, trait), "%s", one.name)
	}
	names := findFilterField(fields, "modifier_names")
	tags := findFilterField(fields, "modifier_tags")
	list := []*gurps.Trait{trait}
	c.Equal([]string{"Accurate", "Nested"}, names.Suggestions(list, false),
		"the names of enabled modifiers are offered, but not those of containers")
	c.Equal([]string{"Accurate", "Cheap", "Nested"}, names.Suggestions(list, true),
		"and those of disabled ones too for a condition that includes them")
	c.Equal([]string{"Enhancement"}, tags.Suggestions(list, false), "the tags of enabled modifiers are offered")
	c.Equal([]string{"Enhancement", "Limitation"}, tags.Suggestions(list, true),
		"and those of disabled ones too for a condition that includes them")
}

// TestFilterModifierFieldsSeeContainerModifiers checks that a trait's modifier fields hold the modifiers of the
// containers holding it, which apply to it, enabled ones only unless a condition includes disabled ones.
func TestFilterModifierFieldsSeeContainerModifiers(t *testing.T) {
	c := check.New(t)
	container := gurps.NewTrait(nil, nil, true)
	container.Name = "Breath Weapons"
	cone := gurps.NewTraitModifier(nil, nil, false)
	cone.Name = "Cone"
	jet := gurps.NewTraitModifier(nil, nil, false)
	jet.Name = "Jet"
	jet.SetDisabled(true)
	container.Modifiers = []*gurps.TraitModifier{cone, jet}
	breath := gurps.NewTrait(nil, container, false)
	breath.Name = "Innate Attack"
	container.Children = []*gurps.Trait{breath}
	c.True(cone.Enabled(), "precondition: Cone is enabled")
	c.False(jet.Enabled(), "precondition: Jet is disabled")
	c.Equal(0, len(breath.Modifiers), "precondition: the trait has no modifiers of its own")

	fields := gurps.TraitFilterFields()
	c.True(matchesListFilter(newTestListFilter(newTextFilterCondition("modifier_names", criteria.IsText, "Cone")),
		fields, breath), "the container's enabled modifier applies to the trait")
	c.False(matchesListFilter(newTestListFilter(newTextFilterCondition("modifier_names", criteria.IsText, "Jet")),
		fields, breath), "its disabled one doesn't")
	c.True(matchesListFilter(newTestListFilter(includingDisabled(newTextFilterCondition("modifier_names",
		criteria.IsText, "Jet"))), fields, breath), "unless the condition includes disabled modifiers")
}

// TestFilterModifierFieldsOnEquipment checks that equipment offers the same fields over its own modifiers.
func TestFilterModifierFieldsOnEquipment(t *testing.T) {
	c := check.New(t)
	equipment := gurps.NewEquipment(nil, nil, false)
	accurate := gurps.NewEquipmentModifier(nil, nil, false)
	accurate.Name = "Accurate"
	accurate.LocalNotes = "Sharp"
	accurate.Tags = []string{"Enhancement"}
	cheap := gurps.NewEquipmentModifier(nil, nil, false)
	cheap.Name = "Cheap"
	cheap.SetDisabled(true)
	equipment.Modifiers = []*gurps.EquipmentModifier{accurate, cheap}
	c.True(accurate.Enabled(), "precondition: the accurate modifier is enabled")
	c.False(cheap.Enabled(), "precondition: the cheap modifier is disabled")

	fields := gurps.EquipmentFilterFields()
	for _, one := range []struct {
		name string
		cond *gurps.FilterCondition
		want bool
	}{
		{"name", newTextFilterCondition("modifier_names", criteria.IsText, "Accurate"), true},
		{"notes", newTextFilterCondition("modifier_notes", criteria.IsText, "Sharp"), true},
		{"tag", newTextFilterCondition("modifier_tags", criteria.IsText, "Enhancement"), true},
		{"disabled name", newTextFilterCondition("modifier_names", criteria.IsText, "Cheap"), false},
		{
			"disabled name, including disabled",
			includingDisabled(newTextFilterCondition("modifier_names", criteria.IsText, "Cheap")), true,
		},
	} {
		c.Equal(one.want, matchesListFilter(newTestListFilter(one.cond), fields, equipment), "%s", one.name)
	}
	names := findFilterField(fields, "modifier_names")
	list := []*gurps.Equipment{equipment}
	c.Equal([]string{"Accurate"}, names.Suggestions(list, false), "the names of enabled modifiers are offered")
	c.Equal([]string{"Accurate", "Cheap"}, names.Suggestions(list, true),
		"and those of disabled ones too for a condition that includes them")
}

// TestFilterIncludeDisabledIgnoredElsewhere checks that a field not drawn from modifiers ignores IncludeDisabled.
func TestFilterIncludeDisabledIgnoredElsewhere(t *testing.T) {
	c := check.New(t)
	trait := gurps.NewTrait(nil, nil, false)
	trait.Tags = []string{"Mental"}
	c.Equal([]string{"Mental"}, trait.TagList(), "precondition: the trait has a tag")
	fields := gurps.TraitFilterFields()
	c.True(matchesListFilter(newTestListFilter(includingDisabled(newTextFilterCondition("tags", criteria.IsText,
		"Mental"))), fields, trait), "the tags field matches as it would without it")
}

// TestFilterEquipmentQuantityFields checks quantity, which every piece of equipment but a group has, and the contained
// quantity and weight, which only a container has, leaving out what a disabled item holds.
func TestFilterEquipmentQuantityFields(t *testing.T) {
	c := check.New(t)
	arrows := gurps.NewEquipment(nil, nil, false)
	arrows.Name = "Arrows"
	arrows.Quantity = fxp.FromInteger(12)
	c.False(arrows.Container(), "precondition: the arrows are not a container")

	quiver := gurps.NewEquipment(nil, nil, true)
	quiver.Name = "Quiver"
	bolts := gurps.NewEquipment(nil, quiver, false)
	bolts.Quantity = fxp.FromInteger(3)
	bolts.BaseWeight = "1 lb"
	spare := gurps.NewEquipment(nil, quiver, false)
	spare.Quantity = fxp.FromInteger(5)
	spare.BaseWeight = "2 lb"
	spare.SetDisabled(true)
	quiver.Children = []*gurps.Equipment{bolts, spare}
	c.True(quiver.Container(), "precondition: the quiver is a container")
	c.Equal(fxp.One, quiver.Quantity, "precondition: the quiver is a single item")
	c.False(bolts.IsDisabled(), "precondition: the bolts are enabled")
	c.True(spare.IsDisabled(), "precondition: the spare bolts are disabled")

	group := gurps.NewEquipment(nil, nil, true)
	group.Name = "Kit"
	group.ContainerType = eqcontainer.Group
	c.True(group.IsGroup(), "precondition: the kit is a group")

	fields := gurps.EquipmentFilterFields()
	threePounds := fxp.WeightFromInteger(3, fxp.Pound)
	for _, one := range []struct {
		name      string
		equipment *gurps.Equipment
		cond      *gurps.FilterCondition
		want      bool
	}{
		{"quantity", arrows, newNumberFilterCondition("quantity", criteria.EqualsNumber, fxp.FromInteger(12)), true},
		{"quantity, container", quiver, newNumberFilterCondition("quantity", criteria.EqualsNumber, fxp.One), true},
		{"quantity, group", group, gurps.NewFilterCondition(nil, "quantity"), false},
		{"not quantity, group", group, negated(gurps.NewFilterCondition(nil, "quantity")), true},
		{
			"contained quantity", quiver,
			newNumberFilterCondition("contained_quantity", criteria.EqualsNumber, fxp.FromInteger(3)), true,
		},
		{"contained quantity, not a container", arrows, gurps.NewFilterCondition(nil, "contained_quantity"), false},
		{
			"not contained quantity, not a container", arrows,
			negated(gurps.NewFilterCondition(nil, "contained_quantity")), true,
		},
		{
			"contained weight", quiver,
			newWeightFilterCondition("contained_weight", criteria.EqualsNumber, threePounds), true,
		},
		{"contained weight, not a container", arrows, gurps.NewFilterCondition(nil, "contained_weight"), false},
		{
			"not contained weight, not a container", arrows,
			negated(gurps.NewFilterCondition(nil, "contained_weight")), true,
		},
	} {
		c.Equal(one.want, matchesListFilter(newTestListFilter(one.cond), fields, one.equipment), "%s", one.name)
	}
}

// TestFilterSkillAndSpellFields checks the optional specialization of a skill, and whether a skill or spell requires a
// tech level, which it does whenever it has one, even an empty one, as library entries do.
func TestFilterSkillAndSpellFields(t *testing.T) {
	c := check.New(t)
	empty := ""
	tlSkill := gurps.NewSkill(nil, nil, false)
	tlSkill.Name = "Electronics Operation"
	tlSkill.TechLevel = &empty
	tlSkill.OptionalSpecialization = "Sensors"
	c.NotNil(tlSkill.TechLevel, "precondition: the TL skill has a tech level")
	plainSkill := gurps.NewSkill(nil, nil, false)
	plainSkill.Name = "Climbing"
	c.Nil(plainSkill.TechLevel, "precondition: the plain skill has no tech level")
	c.Equal("", plainSkill.OptionalSpecialization, "precondition: the plain skill has no optional specialization")
	tlSpell := gurps.NewSpell(nil, nil, false)
	tlSpell.Name = "Seek Machine"
	tlSpell.TechLevel = &empty
	c.NotNil(tlSpell.TechLevel, "precondition: the TL spell has a tech level")
	plainSpell := gurps.NewSpell(nil, nil, false)
	plainSpell.Name = "Light"
	c.Nil(plainSpell.TechLevel, "precondition: the plain spell has no tech level")

	skillFields := gurps.SkillFilterFields()
	optional := newTextFilterCondition("optional_specialization", criteria.IsText, "sensors")
	c.True(matchesListFilter(newTestListFilter(optional), skillFields, tlSkill), "optional specialization")
	c.False(matchesListFilter(newTestListFilter(optional), skillFields, plainSkill), "no optional specialization")
	tl := newTestListFilter(gurps.NewFilterCondition(nil, "tech_level"))
	c.True(matchesListFilter(tl, skillFields, tlSkill), "a skill with an empty tech level requires one")
	c.False(matchesListFilter(tl, skillFields, plainSkill), "a skill without one doesn't")
	spellFields := gurps.SpellFilterFields()
	c.True(matchesListFilter(tl, spellFields, tlSpell), "a spell with an empty tech level requires one")
	c.False(matchesListFilter(tl, spellFields, plainSpell), "a spell without one doesn't")
}

// TestFilterIncludeDisabledOnDisk checks that IncludeDisabled is written only when set, and survives a round trip.
func TestFilterIncludeDisabledOnDisk(t *testing.T) {
	c := check.New(t)
	plain := gurps.NewFilterCondition(nil, "modifier_names")
	c.False(plain.IncludeDisabled, "precondition: the plain condition doesn't include disabled modifiers")
	data, err := jio.Marshal(plain)
	c.NoError(err, "the plain condition should marshal")
	c.Equal(`{"type":"condition","field":"modifier_names"}`, string(data), "an absent IncludeDisabled is omitted")

	including := includingDisabled(gurps.NewFilterCondition(nil, "modifier_names"))
	c.True(including.IncludeDisabled, "precondition: the other condition includes disabled modifiers")
	data, err = jio.Marshal(including)
	c.NoError(err, "the including condition should marshal")
	c.Equal(`{"type":"condition","field":"modifier_names","include_disabled":true}`, string(data),
		"a set IncludeDisabled is written")
	var restored gurps.FilterCondition
	c.NoError(jio.Unmarshal(data, &restored), "the including condition should unmarshal")
	c.True(restored.IncludeDisabled, "IncludeDisabled survives a round trip")
}

// TestFilterIncludeDisabledDroppedWhereUnused checks that loading saved filters clears IncludeDisabled on a condition
// whose field can't include disabled modifiers, at any depth, while keeping it where it can, on a field this version
// doesn't know, and under a list type this version doesn't know.
func TestFilterIncludeDisabledDroppedWhereUnused(t *testing.T) {
	c := check.New(t)
	tags := includingDisabled(gurps.NewFilterCondition(nil, "tags"))
	names := includingDisabled(gurps.NewFilterCondition(nil, "modifier_names"))
	future := includingDisabled(gurps.NewFilterCondition(nil, "future_field"))
	nestedTags := includingDisabled(gurps.NewFilterCondition(nil, "tags"))
	nested := gurps.NewFilterGroup(nil)
	nested.Children = gurps.FilterNodes{nestedTags}
	traits := newTestListFilter(tags, names, future, nested)
	futureListTags := includingDisabled(gurps.NewFilterCondition(nil, "tags"))
	futureList := newTestListFilter(futureListTags)
	for _, one := range []*gurps.FilterCondition{tags, names, future, nestedTags, futureListTags} {
		c.True(one.IncludeDisabled, "precondition: %s starts out including disabled modifiers", one.Field)
	}

	dir := t.TempDir()
	c.NoError(gurps.SaveListFilters(map[string][]*gurps.ListFilter{"adq": {traits}, "zzz": {futureList}},
		filepath.Join(dir, "x"+gurps.ListFiltersExt)), "the filters should save")
	loaded, err := gurps.NewListFiltersFromFile(os.DirFS(dir), "x"+gurps.ListFiltersExt)
	c.NoError(err, "the filters should load")
	c.Equal(1, len(loaded["adq"]), "the trait filter loads")
	c.Equal(1, len(loaded["zzz"]), "the filter of an unknown list type loads")
	if len(loaded["adq"]) != 1 || len(loaded["zzz"]) != 1 {
		return
	}
	root := loaded["adq"][0].Root
	includes := func(node gurps.FilterNode) bool {
		cond, ok := node.(*gurps.FilterCondition)
		return ok && cond.IncludeDisabled
	}
	c.False(includes(root.Children[0]), "a tags condition drops it")
	c.True(includes(root.Children[1]), "a modifier names condition keeps it")
	c.True(includes(root.Children[2]), "a condition on an unknown field keeps it")
	group, ok := root.Children[3].(*gurps.FilterGroup)
	c.True(ok, "the nested group loads")
	if ok {
		c.False(includes(group.Children[0]), "a nested tags condition drops it")
	}
	c.True(includes(loaded["zzz"][0].Root.Children[0]), "a condition under an unknown list type keeps it")
}
