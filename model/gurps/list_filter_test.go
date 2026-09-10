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
	"encoding/json/jsontext"
	"fmt"
	"regexp"
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/filternode"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// newTextFilterCondition creates an unowned condition on the field with the given key using a text comparison. The
// same criteria is what a list-valued field is tested with, so this serves both kinds.
func newTextFilterCondition(field string, compare criteria.StringComparison, qualifier string) *gurps.FilterCondition {
	cond := gurps.NewFilterCondition(nil, field)
	cond.Text = criteria.Text{Compare: compare, Qualifier: qualifier}
	return cond
}

// newNumberFilterCondition creates an unowned condition on the field with the given key using a number comparison.
func newNumberFilterCondition(field string, compare criteria.NumericComparison, qualifier fxp.Int,
) *gurps.FilterCondition {
	cond := gurps.NewFilterCondition(nil, field)
	cond.Number = criteria.Number{Compare: compare, Qualifier: qualifier}
	return cond
}

// newWeightFilterCondition creates an unowned condition on the field with the given key using a weight comparison.
func newWeightFilterCondition(field string, compare criteria.NumericComparison, qualifier fxp.Weight,
) *gurps.FilterCondition {
	cond := gurps.NewFilterCondition(nil, field)
	cond.Weight = criteria.Weight{Compare: compare, Qualifier: qualifier}
	return cond
}

// newTestListFilter creates a filter whose root group requires all of the given nodes to match.
func newTestListFilter(nodes ...gurps.FilterNode) *gurps.ListFilter {
	f := gurps.NewListFilter("Test")
	f.Root.Children = nodes
	f.EnsureValidity()
	return f
}

// newTestUnknownFilterNode creates an unknown filter node of the kind a newer version of GCS might write.
func newTestUnknownFilterNode() *gurps.UnknownFilterNode {
	return gurps.NewUnknownFilterNode("future_node", jsontext.Value(`{"type":"future_node","weird":[1,2]}`))
}

// checkFilterParents verifies that every node below group reports group as its parent, recursing into sub-groups.
func checkFilterParents(c check.Checker, group *gurps.FilterGroup, path string) {
	for i, child := range group.Children {
		childPath := fmt.Sprintf("%s/%d", path, i)
		c.True(child.ParentGroup() == group, "%s should report the group holding it as its parent", childPath)
		if sub, ok := child.(*gurps.FilterGroup); ok {
			checkFilterParents(c, sub, childPath)
		}
	}
}

// TestListFilterJSONRoundTrip verifies that a filter holding every kind of node survives a save and load unchanged,
// both in what it writes and in the parent links that the JSON doesn't carry and the load has to restore.
func TestListFilterJSONRoundTrip(t *testing.T) {
	c := check.New(t)

	// The name is deliberately padded, since loading trims it.
	f := gurps.NewListFilter("  Round Trip  ")
	textCond := newTextFilterCondition("name", criteria.ContainsText, "Alert")
	listCond := newTextFilterCondition("tags", criteria.IsText, "Mental, Physical")
	numberCond := newNumberFilterCondition("points", criteria.AtLeastNumber, fxp.FromInteger(5))
	boolCond := gurps.NewFilterCondition(nil, "container")
	boolCond.Not = true
	anythingCond := gurps.NewFilterCondition(nil, "reference")
	sub := gurps.NewFilterGroup(nil)
	sub.All = false
	sub.Not = true
	sub.Children = gurps.FilterNodes{newWeightFilterCondition("weight", criteria.AtMostNumber,
		fxp.WeightFromInteger(5, fxp.Pound))}
	f.Root.Children = gurps.FilterNodes{textCond, listCond, numberCond, boolCond, anythingCond, sub}
	f.EnsureValidity()

	data, err := jio.Marshal(f)
	c.NoError(err, "a filter should marshal")

	var restored gurps.ListFilter
	c.NoError(jio.Unmarshal(data, &restored), "a filter should unmarshal")
	c.Equal("Round Trip", restored.Name, "the name should survive the round trip, trimmed")
	c.True(restored.Root.All, "the root group's combining mode should survive the round trip")
	c.Equal(6, len(restored.Root.Children), "every child should survive the round trip")

	again, err := jio.Marshal(&restored)
	c.NoError(err, "a restored filter should marshal")
	c.Equal(string(data), string(again), "re-saving a loaded filter must reproduce the original bytes")

	// The parent links are rebuilt by ListFilter.UnmarshalJSONFrom by way of EnsureValidity.
	c.Nil(restored.Root.ParentGroup(), "the root group has no parent")
	checkFilterParents(c, restored.Root, "root")
	restoredSub, ok := restored.Root.Children[5].(*gurps.FilterGroup)
	c.True(ok, "the nested group should load as a group")
	if ok {
		c.False(restoredSub.All, "the nested group's combining mode should survive the round trip")
		c.True(restoredSub.Not, "the nested group's negation should survive the round trip")
		c.Equal(1, len(restoredSub.Children), "the nested group's child should survive the round trip")
	}
	restoredBool, ok := restored.Root.Children[3].(*gurps.FilterCondition)
	c.True(ok, "the bool condition should load as a condition")
	if ok {
		c.True(restoredBool.Not, "a condition's negation should survive the round trip")
	}

	// The on-disk shape: a group names its type, always writes how it combines, and leaves out a negation it doesn't
	// have.
	groupData, err := jio.Marshal(gurps.NewFilterGroup(nil))
	c.NoError(err, "a group should marshal")
	c.Equal(`{"type":"group","all":true}`, string(groupData),
		"a group writes its type and combining mode, but not an absent negation")

	// A condition whose criteria all accept anything writes none of them.
	anythingData, err := jio.Marshal(anythingCond)
	c.NoError(err, "a condition should marshal")
	c.Equal(`{"type":"condition","field":"reference"}`, string(anythingData),
		`the "text", "number" and "weight" criteria are omitted when they accept anything`)

	whole := string(data)
	c.Contains(whole, `"type":"group"`, "the groups name their type")
	c.Contains(whole, `"type":"condition"`, "the conditions name their type")
	c.Contains(whole, `"all":true`, "an all-of group writes its combining mode")
	c.Contains(whole, `"all":false`, "an any-of group writes its combining mode")
	c.Contains(whole, `"not":true`, "a negation is written")
	c.NotContains(whole, `"not":false`, "an absent negation is omitted")
	c.Contains(whole, `"text":{"compare":"contains","qualifier":"Alert"}`, "a text criteria is written")
	c.Contains(whole, `"number":{"compare":"at_least"`, "a number criteria is written")
	c.Contains(whole, `"weight":{"compare":"at_most"`, "a weight criteria is written")
}

// TestListFilterUnknownNodePreserved verifies that a filter node whose type this build doesn't recognize loads as an
// UnknownFilterNode, is written back out unchanged so a newer version of GCS can still use it, and participates in the
// filter's hash rather than being invisible to it.
func TestListFilterUnknownNodePreserved(t *testing.T) {
	c := check.New(t)

	const knownChild = `{"type": "condition", "field": "name", "text": {"compare": "is", "qualifier": "Alertness"}}`
	const unknownChild = `{"type": "future_node", "weird": [1, 2]}`
	withUnknown := `{"name": "Filter", "root": {"type": "group", "all": true, "children": [` +
		knownChild + `, ` + unknownChild + `]}}`
	withoutUnknown := `{"name": "Filter", "root": {"type": "group", "all": true, "children": [` + knownChild + `]}}`

	var f gurps.ListFilter
	c.NoError(jio.Unmarshal([]byte(withUnknown), &f), "an unknown node type must not prevent loading")
	c.Equal(2, len(f.Root.Children), "both nodes should be present")

	_, ok := f.Root.Children[0].(*gurps.FilterCondition)
	c.True(ok, "the known node should still load normally")

	unknown, ok := f.Root.Children[1].(*gurps.UnknownFilterNode)
	c.True(ok, "an unrecognized node should load as an UnknownFilterNode, not be coerced to another type")
	if ok {
		c.Equal("future_node", unknown.Kind, "the original type should be retained")
		c.Equal(filternode.Unknown, unknown.NodeType(), "an UnknownFilterNode reports the Unknown type")
		c.True(unknown.ParentGroup() == f.Root, "an UnknownFilterNode is linked to its parent like any other node")
	}

	out, err := jio.Marshal(f.Root.Children)
	c.NoError(err, "the children should marshal")
	c.Equal(compactJSON(c, unknownChild), jsonArrayElement(c, string(out), 1),
		"saving must reproduce the unknown node exactly")

	var without gurps.ListFilter
	c.NoError(jio.Unmarshal([]byte(withoutUnknown), &without), "the comparison filter should load")
	c.NotEqual(gurps.Hash64(&without), gurps.Hash64(&f), "an unknown node must contribute to the filter's hash")

	// A clone must carry the data along, since editors work on clones.
	clone := f.Clone()
	clonedUnknown, ok := clone.Root.Children[1].(*gurps.UnknownFilterNode)
	c.True(ok, "a cloned UnknownFilterNode is still an UnknownFilterNode")
	if ok {
		c.Equal(unknown.Kind, clonedUnknown.Kind, "the clone retains the original type")
		c.Equal(unknown.Data.String(), clonedUnknown.Data.String(), "the clone retains the original data")
	}
}

// TestEveryFilterNodeTypeDecodesToItsConcreteType verifies that the polymorphic filter node decoder has a concrete type
// for every node type this build knows about, so that none of them is quietly preserved as an UnknownFilterNode, and
// that the reserved Unknown type itself, which has no concrete representation, is.
func TestEveryFilterNodeTypeDecodesToItsConcreteType(t *testing.T) {
	for _, one := range append(slices.Clone(filternode.Types), filternode.Unknown) {
		t.Run(one.Key(), func(t *testing.T) {
			c := check.New(t)
			var nodes gurps.FilterNodes
			c.NoError(jio.Unmarshal(fmt.Appendf(nil, `[{"type":%q}]`, one.Key()), &nodes), "the node should load")
			c.Equal(1, len(nodes), "the node should be present")
			if len(nodes) != 1 {
				return
			}
			_, isUnknown := nodes[0].(*gurps.UnknownFilterNode)
			c.Equal(one == filternode.Unknown, isUnknown, "unknown wrapper used")
			c.Equal(one, nodes[0].NodeType(), "the node type round-trips")
		})
	}
}

// TestListFilterMatchesByKind verifies that a condition on each kind of field is evaluated with the criteria that kind
// calls for, against the value the field's accessor produces.
func TestListFilterMatchesByKind(t *testing.T) {
	c := check.New(t)

	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Alertness"
	trait.Tags = []string{"Mental", "Physical"}
	trait.BasePoints = fxp.FromInteger(5)
	c.Equal(fxp.FromInteger(5), trait.AdjustedPoints(), "the trait should be worth the points it was given")

	untagged := gurps.NewTrait(nil, nil, false)
	untagged.Name = "Untagged"

	container := gurps.NewTrait(nil, nil, true)
	container.Name = "Group"

	for _, one := range []struct {
		name  string
		trait *gurps.Trait
		cond  *gurps.FilterCondition
		want  bool
	}{
		// Text, which ignores case throughout.
		{"text is", trait, newTextFilterCondition("name", criteria.IsText, "alertness"), true},
		{"text is, no match", trait, newTextFilterCondition("name", criteria.IsText, "Acute Vision"), false},
		{"text is not", trait, newTextFilterCondition("name", criteria.IsNotText, "Acute Vision"), true},
		{"text is not, no match", trait, newTextFilterCondition("name", criteria.IsNotText, "alertness"), false},
		{"text contains", trait, newTextFilterCondition("name", criteria.ContainsText, "LERT"), true},
		{"text contains, no match", trait, newTextFilterCondition("name", criteria.ContainsText, "vision"), false},
		{"text does not contain", trait, newTextFilterCondition("name", criteria.DoesNotContainText, "vision"), true},
		{
			"text does not contain, no match", trait,
			newTextFilterCondition("name", criteria.DoesNotContainText, "LERT"), false,
		},
		{"text starts with", trait, newTextFilterCondition("name", criteria.StartsWithText, "ALER"), true},
		{"text starts with, no match", trait, newTextFilterCondition("name", criteria.StartsWithText, "ness"), false},
		{"text ends with", trait, newTextFilterCondition("name", criteria.EndsWithText, "NESS"), true},
		{"text ends with, no match", trait, newTextFilterCondition("name", criteria.EndsWithText, "aler"), false},
		{"text anything", trait, gurps.NewFilterCondition(nil, "name"), true},

		// List, where any one of several values may satisfy a positive comparison.
		{"list is, first value", trait, newTextFilterCondition("tags", criteria.IsText, "mental"), true},
		{"list is, second value", trait, newTextFilterCondition("tags", criteria.IsText, "physical"), true},
		{"list is, no match", trait, newTextFilterCondition("tags", criteria.IsText, "Social"), false},
		{
			"list is, comma-separated qualifier", trait,
			newTextFilterCondition("tags", criteria.IsText, "Mental, Physical"), true,
		},
		{
			"list is, comma-separated qualifier, no match", trait,
			newTextFilterCondition("tags", criteria.IsText, "Social, Exotic"), false,
		},
		{"list contains", trait, newTextFilterCondition("tags", criteria.ContainsText, "ment"), true},
		{"list does not contain", trait, newTextFilterCondition("tags", criteria.DoesNotContainText, "Social"), true},
		{
			"list does not contain, no match", trait,
			newTextFilterCondition("tags", criteria.DoesNotContainText, "ment"), false,
		},
		{"list is, empty tag list", untagged, newTextFilterCondition("tags", criteria.IsText, "Mental"), false},
		{"list is not, empty tag list", untagged, newTextFilterCondition("tags", criteria.IsNotText, "Mental"), true},
		{"list anything, empty tag list", untagged, gurps.NewFilterCondition(nil, "tags"), true},

		// Number.
		{"number is", trait, newNumberFilterCondition("points", criteria.EqualsNumber, fxp.FromInteger(5)), true},
		{
			"number is, no match", trait,
			newNumberFilterCondition("points", criteria.EqualsNumber, fxp.FromInteger(6)), false,
		},
		{"number at least", trait, newNumberFilterCondition("points", criteria.AtLeastNumber, fxp.FromInteger(5)), true},
		{
			"number at least, no match", trait,
			newNumberFilterCondition("points", criteria.AtLeastNumber, fxp.FromInteger(6)), false,
		},
		{"number at most", trait, newNumberFilterCondition("points", criteria.AtMostNumber, fxp.FromInteger(5)), true},
		{
			"number at most, no match", trait,
			newNumberFilterCondition("points", criteria.AtMostNumber, fxp.FromInteger(4)), false,
		},

		// Bool, which needs no criteria at all.
		{"bool, true", container, gurps.NewFilterCondition(nil, "container"), true},
		{"bool, false", trait, gurps.NewFilterCondition(nil, "container"), false},
	} {
		c.Equal(one.want, gurps.MatchesListFilter(newTestListFilter(one.cond), gurps.TraitFilterFields(), one.trait),
			"%s", one.name)
	}

	equipment := gurps.NewEquipment(nil, nil, false)
	equipment.Name = "Sack"
	equipment.BaseWeight = "5 lb"
	fivePounds := fxp.WeightFromInteger(5, fxp.Pound)
	for _, one := range []struct {
		name string
		cond *gurps.FilterCondition
		want bool
	}{
		{"weight is", newWeightFilterCondition("weight", criteria.EqualsNumber, fivePounds), true},
		{"weight is, no match", newWeightFilterCondition("weight", criteria.EqualsNumber,
			fxp.WeightFromInteger(6, fxp.Pound)), false},
		{"weight at least", newWeightFilterCondition("weight", criteria.AtLeastNumber, fivePounds), true},
		{"weight at least, no match", newWeightFilterCondition("weight", criteria.AtLeastNumber,
			fxp.WeightFromInteger(6, fxp.Pound)), false},
		{"weight at most", newWeightFilterCondition("weight", criteria.AtMostNumber, fivePounds), true},
		{"weight at most, no match", newWeightFilterCondition("weight", criteria.AtMostNumber,
			fxp.WeightFromInteger(4, fxp.Pound)), false},
		{"weight anything", gurps.NewFilterCondition(nil, "weight"), true},
	} {
		c.Equal(one.want,
			gurps.MatchesListFilter(newTestListFilter(one.cond), gurps.EquipmentFilterFields(), equipment),
			"%s", one.name)
	}
}

// TestListFilterEmptyGroupMatchesEverything verifies that a group with no children passes everything, whichever way it
// combines its children, and that negating such a group rejects everything.
func TestListFilterEmptyGroupMatchesEverything(t *testing.T) {
	c := check.New(t)
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Alertness"
	fields := gurps.TraitFilterFields()
	for _, all := range []bool{true, false} {
		f := gurps.NewListFilter("Empty")
		f.Root.All = all
		c.True(gurps.MatchesListFilter(f, fields, trait), "an empty group with all=%v matches everything", all)
		f.Root.Not = true
		c.False(gurps.MatchesListFilter(f, fields, trait), "a negated empty group with all=%v matches nothing", all)
	}
}

// TestListFilterNegation verifies that negation inverts a node's result wherever it appears, and that it composes with
// the criteria's own "not" forms rather than being confused by them.
func TestListFilterNegation(t *testing.T) {
	c := check.New(t)
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Alertness"
	trait.Tags = []string{"Mental", "Physical"}
	fields := gurps.TraitFilterFields()

	// On a condition.
	matching := newTextFilterCondition("name", criteria.IsText, "Alertness")
	matching.Not = true
	c.False(gurps.MatchesListFilter(newTestListFilter(matching), fields, trait),
		"negating a condition that matches rejects the node")
	notMatching := newTextFilterCondition("name", criteria.IsText, "Acute Vision")
	notMatching.Not = true
	c.True(gurps.MatchesListFilter(newTestListFilter(notMatching), fields, trait),
		"negating a condition that doesn't match accepts the node")

	// On a group.
	sub := gurps.NewFilterGroup(nil)
	sub.Not = true
	sub.Children = gurps.FilterNodes{newTextFilterCondition("name", criteria.IsText, "Alertness")}
	c.False(gurps.MatchesListFilter(newTestListFilter(sub), fields, trait),
		"negating a group whose children match rejects the node")
	sub.Children = gurps.FilterNodes{newTextFilterCondition("name", criteria.IsText, "Acute Vision")}
	c.True(gurps.MatchesListFilter(newTestListFilter(sub), fields, trait),
		"negating a group whose children don't match accepts the node")

	// A negated "is not" is just an "is".
	negatedIsNot := newTextFilterCondition("name", criteria.IsNotText, "Alertness")
	negatedIsNot.Not = true
	plainIs := newTestListFilter(newTextFilterCondition("name", criteria.IsText, "Alertness"))
	for _, name := range []string{"Alertness", "Acute Vision"} {
		trait.Name = name
		c.Equal(gurps.MatchesListFilter(plainIs, fields, trait),
			gurps.MatchesListFilter(newTestListFilter(negatedIsNot), fields, trait),
			`a negated "is not" behaves as a plain "is" for %q`, name)
	}
	trait.Name = "Alertness"

	// And on a list condition, where the criteria's own negation has list semantics of its own.
	absent := newTextFilterCondition("tags", criteria.DoesNotContainText, "Social")
	c.True(gurps.MatchesListFilter(newTestListFilter(absent), fields, trait),
		`"does not contain" passes when no tag contains the qualifier`)
	absent.Not = true
	c.False(gurps.MatchesListFilter(newTestListFilter(absent), fields, trait),
		`negating a passing "does not contain" rejects the node`)
	present := newTextFilterCondition("tags", criteria.DoesNotContainText, "ment")
	c.False(gurps.MatchesListFilter(newTestListFilter(present), fields, trait),
		`"does not contain" fails when a tag contains the qualifier`)
	present.Not = true
	c.True(gurps.MatchesListFilter(newTestListFilter(present), fields, trait),
		`negating a failing "does not contain" accepts the node`)
}

// TestListFilterUnknownFieldNeverMatches verifies that a condition naming a field this build doesn't know about never
// passes, whatever its negation says. This is deliberate: the condition's intent can't be known, so it fails closed
// rather than quietly turning into "everything" for whichever half of the negation it landed on.
func TestListFilterUnknownFieldNeverMatches(t *testing.T) {
	c := check.New(t)
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Alertness"
	fields := gurps.TraitFilterFields()

	cond := newTextFilterCondition("field_from_a_newer_gcs", criteria.IsText, "Alertness")
	c.False(gurps.MatchesListFilter(newTestListFilter(cond), fields, trait),
		"a condition on an unknown field never matches")
	cond.Not = true
	c.False(gurps.MatchesListFilter(newTestListFilter(cond), fields, trait),
		"negating a condition on an unknown field still never matches")

	anything := gurps.NewFilterCondition(nil, "field_from_a_newer_gcs")
	c.False(gurps.MatchesListFilter(newTestListFilter(anything), fields, trait),
		"a condition on an unknown field never matches, even when its criteria accept anything")
	c.Nil(gurps.FindFilterField(fields, "field_from_a_newer_gcs"), "the field really is unknown")
}

// TestListFilterUnknownNodeNeverMatches verifies that a node type this build doesn't understand never passes, so that
// an all-of group holding one fails and an any-of group holding one gains nothing from it, and that a negation applied
// to the group inverts that result rather than the unknown node's.
func TestListFilterUnknownNodeNeverMatches(t *testing.T) {
	c := check.New(t)
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Alertness"
	fields := gurps.TraitFilterFields()
	matches := func() *gurps.FilterCondition {
		return newTextFilterCondition("name", criteria.IsText, "Alertness")
	}
	doesNotMatch := func() *gurps.FilterCondition {
		return newTextFilterCondition("name", criteria.IsText, "Acute Vision")
	}

	for _, one := range []struct {
		name     string
		all      bool
		not      bool
		children gurps.FilterNodes
		want     bool
	}{
		{"all-of with an unknown node alone", true, false, gurps.FilterNodes{newTestUnknownFilterNode()}, false},
		{
			"all-of with an unknown node beside a match", true, false,
			gurps.FilterNodes{newTestUnknownFilterNode(), matches()},
			false,
		},
		{"any-of with an unknown node alone", false, false, gurps.FilterNodes{newTestUnknownFilterNode()}, false},
		{
			"any-of with an unknown node beside a match", false, false,
			gurps.FilterNodes{newTestUnknownFilterNode(), matches()},
			true,
		},
		{
			"any-of with an unknown node beside a non-match", false, false,
			gurps.FilterNodes{newTestUnknownFilterNode(), doesNotMatch()},
			false,
		},
		{
			"negated all-of with an unknown node alone", true, true,
			gurps.FilterNodes{newTestUnknownFilterNode()},
			true,
		},
		{
			"negated all-of with an unknown node beside a match", true, true,
			gurps.FilterNodes{newTestUnknownFilterNode(), matches()},
			true,
		},
		{
			"negated any-of with an unknown node beside a match", false, true,
			gurps.FilterNodes{newTestUnknownFilterNode(), matches()},
			false,
		},
		{
			"negated any-of with an unknown node alone", false, true,
			gurps.FilterNodes{newTestUnknownFilterNode()},
			true,
		},
	} {
		f := gurps.NewListFilter("Test")
		f.Root.All = one.all
		f.Root.Not = one.not
		f.Root.Children = one.children
		f.EnsureValidity()
		c.Equal(one.want, gurps.MatchesListFilter(f, fields, trait), "%s", one.name)
	}
}

// checkFilterFields verifies that a filter field table is well formed: every field has a unique, storable key and a
// title, exactly one accessor, which is the one its kind names, and an accessor that can actually be run against a
// node of the type the table is for.
func checkFilterFields[T gurps.Node[T]](c check.Checker, name string, fields []*gurps.FilterField[T], samples ...T) {
	c.NotEqual(0, len(fields), "%s: the field table should not be empty", name)
	keyPattern := regexp.MustCompile(`^[a-z0-9_]+$`)
	seen := make(map[string]bool, len(fields))
	for _, field := range fields {
		c.NotEqual("", field.Key, "%s: every field should have a key", name)
		c.True(keyPattern.MatchString(field.Key), "%s: key %q should match %v", name, field.Key, keyPattern)
		c.False(seen[field.Key], "%s: key %q should appear only once", name, field.Key)
		seen[field.Key] = true
		c.NotEqual("", field.Title, "%s: field %q should have a title", name, field.Key)

		present := map[gurps.FilterFieldKind]bool{
			gurps.FilterFieldText:   field.Text != nil,
			gurps.FilterFieldList:   field.List != nil,
			gurps.FilterFieldNumber: field.Number != nil,
			gurps.FilterFieldWeight: field.Weight != nil,
			gurps.FilterFieldBool:   field.Bool != nil,
		}
		count := 0
		for _, set := range present {
			if set {
				count++
			}
		}
		c.Equal(1, count, "%s: field %q should have exactly one accessor", name, field.Key)
		c.True(present[field.Kind], "%s: field %q should have the accessor its kind names", name, field.Key)

		c.True(gurps.FindFilterField(fields, field.Key) == field, "%s: field %q should be findable by key", name,
			field.Key)

		for i, sample := range samples {
			c.NotPanics(func() {
				switch field.Kind {
				case gurps.FilterFieldText:
					_ = field.Text(sample)
				case gurps.FilterFieldList:
					_ = field.List(sample)
				case gurps.FilterFieldNumber:
					_ = field.Number(sample)
				case gurps.FilterFieldWeight:
					_ = field.Weight(sample)
				case gurps.FilterFieldBool:
					_ = field.Bool(sample)
				}
			}, "%s: field %q should evaluate against sample %d", name, field.Key, i)
		}
	}
	c.Nil(gurps.FindFilterField(fields, "no_such_field"), "%s: an absent key should not be found", name)
}

// TestFilterFieldTablesAreWellFormed verifies that every list type's filter field table can be stored, shown and
// evaluated, since a malformed entry would only be discovered when a user built a filter that used it.
func TestFilterFieldTablesAreWellFormed(t *testing.T) {
	c := check.New(t)
	checkFilterFields(c, "traits", gurps.TraitFilterFields(), gurps.NewTrait(nil, nil, false),
		gurps.NewTrait(nil, nil, true))
	checkFilterFields(c, "trait modifiers", gurps.TraitModifierFilterFields(),
		gurps.NewTraitModifier(nil, nil, false), gurps.NewTraitModifier(nil, nil, true))
	checkFilterFields(c, "skills", gurps.SkillFilterFields(), gurps.NewSkill(nil, nil, false),
		gurps.NewSkill(nil, nil, true))
	checkFilterFields(c, "spells", gurps.SpellFilterFields(), gurps.NewSpell(nil, nil, false),
		gurps.NewSpell(nil, nil, true))
	checkFilterFields(c, "equipment", gurps.EquipmentFilterFields(), gurps.NewEquipment(nil, nil, false),
		gurps.NewEquipment(nil, nil, true))
	checkFilterFields(c, "equipment modifiers", gurps.EquipmentModifierFilterFields(),
		gurps.NewEquipmentModifier(nil, nil, false), gurps.NewEquipmentModifier(nil, nil, true))
	checkFilterFields(c, "notes", gurps.NoteFilterFields(), gurps.NewNote(nil, nil, false),
		gurps.NewNote(nil, nil, true))
}

// TestListFilterCloneIsDeep verifies that a clone shares nothing with the filter it came from, since the filter editor
// works on a clone and must not alter the saved filter until the user accepts the changes.
func TestListFilterCloneIsDeep(t *testing.T) {
	c := check.New(t)

	f := gurps.NewListFilter("Deep")
	cond := newTextFilterCondition("name", criteria.IsText, "Alertness")
	sub := gurps.NewFilterGroup(nil)
	sub.Children = gurps.FilterNodes{newTextFilterCondition("tags", criteria.ContainsText, "Mental")}
	f.Root.Children = gurps.FilterNodes{cond, sub}
	f.EnsureValidity()
	before := gurps.Hash64(f)

	clone := f.Clone()
	c.Equal(before, gurps.Hash64(clone), "a clone has the same content as the original")
	c.True(clone != f, "the clone is a different filter")
	c.True(clone.Root != f.Root, "the clone has its own root group")
	c.Nil(clone.Root.ParentGroup(), "the clone's root group has no parent")
	checkFilterParents(c, clone.Root, "clone root")
	for i := range f.Root.Children {
		c.True(clone.Root.Children[i] != f.Root.Children[i], "the clone has its own child %d", i)
	}

	clonedSub, ok := clone.Root.Children[1].(*gurps.FilterGroup)
	c.True(ok, "the clone's nested group is still a group")
	if !ok {
		return
	}
	c.True(clonedSub.Children[0] != sub.Children[0], "the clone has its own nested child")
	clonedSub.Children = append(clonedSub.Children, gurps.NewFilterCondition(clonedSub, "notes"))
	clonedSub.Not = true
	clonedCond, ok := clone.Root.Children[0].(*gurps.FilterCondition)
	c.True(ok, "the clone's condition is still a condition")
	if ok {
		clonedCond.Text.Qualifier = "Changed"
	}

	c.Equal(before, gurps.Hash64(f), "mutating the clone leaves the original alone")
	c.Equal(1, len(sub.Children), "the original's nested group did not gain a child")
	c.False(sub.Not, "the original's nested group was not negated")
	c.Equal("Alertness", cond.Text.Qualifier, "the original's qualifier was not changed")
	c.NotEqual(gurps.Hash64(f), gurps.Hash64(clone), "the mutated clone no longer matches the original")
	checkFilterParents(c, clone.Root, "mutated clone root")
}

// TestListFilterHash verifies that the hash covers everything a filter is made of, since it is what decides whether a
// filter has unsaved changes, and that it copes with the nils a partially built filter may hold.
func TestListFilterHash(t *testing.T) {
	c := check.New(t)
	build := func() *gurps.ListFilter {
		f := gurps.NewListFilter("Hashed")
		f.Root.Children = gurps.FilterNodes{
			newTextFilterCondition("name", criteria.IsText, "Alertness"),
			newNumberFilterCondition("points", criteria.AtLeastNumber, fxp.FromInteger(5)),
			newWeightFilterCondition("weight", criteria.AtMostNumber, fxp.WeightFromInteger(5, fxp.Pound)),
		}
		f.EnsureValidity()
		return f
	}
	base := gurps.Hash64(build())
	c.Equal(base, gurps.Hash64(build()), "filters with the same content hash the same")

	// The filters are built by build(), so anything but a condition at these positions is a bug in this test.
	condAt := func(f *gurps.ListFilter, index int) *gurps.FilterCondition {
		cond, ok := f.Root.Children[index].(*gurps.FilterCondition)
		c.True(ok, "child %d of the built filter should be a condition", index)
		if !ok {
			return gurps.NewFilterCondition(nil, "unused")
		}
		return cond
	}

	for _, one := range []struct {
		name   string
		mutate func(f *gurps.ListFilter)
	}{
		{"name", func(f *gurps.ListFilter) { f.Name = "Other" }},
		{"group negation", func(f *gurps.ListFilter) { f.Root.Not = true }},
		{"group combining mode", func(f *gurps.ListFilter) { f.Root.All = false }},
		{"condition negation", func(f *gurps.ListFilter) { condAt(f, 0).Not = true }},
		{"condition field", func(f *gurps.ListFilter) { condAt(f, 0).Field = "notes" }},
		{"text comparison", func(f *gurps.ListFilter) { condAt(f, 0).Text.Compare = criteria.ContainsText }},
		{"text qualifier", func(f *gurps.ListFilter) { condAt(f, 0).Text.Qualifier = "Acute Vision" }},
		{"number comparison", func(f *gurps.ListFilter) { condAt(f, 1).Number.Compare = criteria.AtMostNumber }},
		{"number qualifier", func(f *gurps.ListFilter) { condAt(f, 1).Number.Qualifier = fxp.FromInteger(6) }},
		{"weight comparison", func(f *gurps.ListFilter) { condAt(f, 2).Weight.Compare = criteria.AtLeastNumber }},
		{"weight qualifier", func(f *gurps.ListFilter) {
			condAt(f, 2).Weight.Qualifier = fxp.WeightFromInteger(6, fxp.Pound)
		}},
		{"child count", func(f *gurps.ListFilter) { f.Root.Children = f.Root.Children[:2] }},
	} {
		edited := build()
		one.mutate(edited)
		c.NotEqual(base, gurps.Hash64(edited), "changing the %s changes the hash", one.name)
	}

	var nilFilter *gurps.ListFilter
	c.NotPanics(func() { _ = gurps.Hash64(nilFilter) }, "hashing a nil filter must not panic")
	var nilGroup *gurps.FilterGroup
	c.NotPanics(func() { _ = gurps.Hash64(nilGroup) }, "hashing a nil group must not panic")
	c.NotPanics(func() { _ = gurps.Hash64(&gurps.ListFilter{Name: "Rootless"}) },
		"hashing a filter without a root must not panic")
	c.NotEqual(base, gurps.Hash64(&gurps.ListFilter{Name: "Rootless"}),
		"a filter without a root does not hash like a populated one")
}

// TestSettingsListFilterHelpers verifies the settings-level bookkeeping for saved filters: the ordering a user sees,
// the copies that keep callers from reaching into the stored list, and the pruning that EnsureValidity does.
func TestSettingsListFilterHelpers(t *testing.T) {
	c := check.New(t)
	const key = "adq"
	s := &gurps.Settings{}

	c.Nil(s.ListFiltersFor(key), "an unknown key has no filters")
	c.Nil(s.ListFilters, "asking about an unknown key does not create the map")

	// Filters are ordered by name, ignoring case, and numbers within a name compare as numbers.
	for _, name := range []string{"b", "A", "c"} {
		s.AddListFilter(key, gurps.NewListFilter(name))
	}
	c.Equal([]string{"A", "b", "c"}, listFilterNames(s.ListFiltersFor(key)),
		"saved filters are sorted without regard to case")
	s.AddListFilter(key, gurps.NewListFilter("item10"))
	s.AddListFilter(key, gurps.NewListFilter("item9"))
	c.Equal([]string{"A", "b", "c", "item9", "item10"}, listFilterNames(s.ListFiltersFor(key)),
		"saved filters are sorted naturally, so item9 precedes item10")

	// The returned slice is a copy, so a caller can't reorder or shorten what is stored.
	list := s.ListFiltersFor(key)
	list = append(list, gurps.NewListFilter("intruder"))
	list[0] = nil
	c.Equal([]string{"A", "b", "c", "item9", "item10"}, listFilterNames(s.ListFiltersFor(key)),
		"the returned slice is a copy of the stored one")

	// Replacement is by pointer identity, and adds when the filter to replace isn't there.
	old := s.ListFiltersFor(key)[0]
	replacement := gurps.NewListFilter("A2")
	s.ReplaceListFilter(key, old, replacement)
	c.Equal([]string{"A2", "b", "c", "item9", "item10"}, listFilterNames(s.ListFiltersFor(key)),
		"replacing swaps the filter in place")
	c.Equal(-1, slices.Index(s.ListFiltersFor(key), old), "the replaced filter is gone")
	s.ReplaceListFilter(key, gurps.NewListFilter("never stored"), gurps.NewListFilter("d"))
	c.Equal([]string{"A2", "b", "c", "d", "item9", "item10"}, listFilterNames(s.ListFiltersFor(key)),
		"replacing a filter that isn't stored adds the replacement")

	// A filter renamed in place stays where it was until the list is sorted again.
	s.ListFiltersFor(key)[1].Name = "zzz"
	c.Equal([]string{"A2", "zzz", "c", "d", "item9", "item10"}, listFilterNames(s.ListFiltersFor(key)),
		"renaming in place doesn't move the filter by itself")
	s.ResortListFilters(key)
	c.Equal([]string{"A2", "c", "d", "item9", "item10", "zzz"}, listFilterNames(s.ListFiltersFor(key)),
		"re-sorting puts the renamed filter in its place")
	s.ListFiltersFor(key)[5].Name = "b"
	s.ResortListFilters(key)
	c.Equal([]string{"A2", "b", "c", "d", "item9", "item10"}, listFilterNames(s.ListFiltersFor(key)),
		"re-sorting puts it back again")
	s.ResortListFilters("never stored")
	c.Nil(s.ListFilters["never stored"], "re-sorting a key that isn't stored doesn't create it")

	s.RemoveListFilter(key, replacement)
	c.Equal([]string{"b", "c", "d", "item9", "item10"}, listFilterNames(s.ListFiltersFor(key)),
		"removing drops just that filter")
	s.RemoveListFilter(key, gurps.NewListFilter("never stored"))
	c.Equal([]string{"b", "c", "d", "item9", "item10"}, listFilterNames(s.ListFiltersFor(key)),
		"removing a filter that isn't stored changes nothing")

	// Name collisions ignore case and surrounding whitespace, and the filter being edited doesn't collide with itself.
	kept := s.ListFiltersFor(key)[0]
	c.True(s.ListFilterNameInUse(key, "  B  ", nil), "a name in use is detected regardless of case and whitespace")
	c.False(s.ListFilterNameInUse(key, "B", kept), "the filter named as the exception doesn't count")
	c.False(s.ListFilterNameInUse(key, "unused", nil), "a name that isn't in use is not reported")
	c.False(s.ListFilterNameInUse("spl", "b", nil), "names are tracked per list type")

	// Emptying a key removes it rather than leaving an empty list behind.
	for _, f := range s.ListFiltersFor(key) {
		s.RemoveListFilter(key, f)
	}
	_, exists := s.ListFilters[key]
	c.False(exists, "removing the last filter deletes the key")

	// EnsureValidity prunes what can't be used and keeps what it doesn't recognize.
	s.ListFilters = map[string][]*gurps.ListFilter{
		key:      {nil, gurps.NewListFilter("Keep"), gurps.NewListFilter("  "), gurps.NewListFilter("keep")},
		"spl":    {nil, gurps.NewListFilter("")},
		"future": {gurps.NewListFilter("From a newer GCS")},
	}
	s.EnsureValidity()
	c.Equal([]string{"Keep"}, listFilterNames(s.ListFiltersFor(key)),
		"nil, unnamed and duplicate-named filters are dropped")
	_, exists = s.ListFilters["spl"]
	c.False(exists, "a key left with nothing usable is deleted")
	c.Equal([]string{"From a newer GCS"}, listFilterNames(s.ListFiltersFor("future")),
		"an unrecognized key is kept, since a newer version of GCS may have written it")

	// A filter that arrives as part of a settings load has its parent links restored.
	source := &gurps.Settings{}
	nested := gurps.NewListFilter("Nested")
	sub := gurps.NewFilterGroup(nil)
	sub.All = false
	sub.Children = gurps.FilterNodes{newTextFilterCondition("name", criteria.IsText, "Alertness")}
	nested.Root.Children = gurps.FilterNodes{newTextFilterCondition("notes", criteria.ContainsText, "x"), sub}
	nested.EnsureValidity()
	source.AddListFilter(key, nested)
	data, err := jio.Marshal(source)
	c.NoError(err, "settings holding a filter should marshal")
	var loaded gurps.Settings
	c.NoError(jio.Unmarshal(data, &loaded), "settings holding a filter should unmarshal")
	loaded.EnsureValidity()
	loadedFilters := loaded.ListFiltersFor(key)
	c.Equal(1, len(loadedFilters), "the saved filter should load")
	if len(loadedFilters) != 1 {
		return
	}
	loadedFilter := loadedFilters[0]
	c.Equal("Nested", loadedFilter.Name, "the saved filter's name should load")
	c.NotNil(loadedFilter.Root, "a loaded filter always has a root group")
	c.Nil(loadedFilter.Root.ParentGroup(), "a loaded filter's root group has no parent")
	c.Equal(2, len(loadedFilter.Root.Children), "the saved filter's children should load")
	checkFilterParents(c, loadedFilter.Root, "loaded root")
}

// listFilterNames returns the names of the given filters, so that ordering can be compared directly.
func listFilterNames(filters []*gurps.ListFilter) []string {
	names := make([]string, len(filters))
	for i, f := range filters {
		names[i] = f.Name
	}
	return names
}

// TestListFilterKeyForExtension verifies that the key saved filters are stored under is derived from a file extension
// the way the rest of GCS spells one, and that only the library list types have one.
func TestListFilterKeyForExtension(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		in   string
		want string
	}{
		{".adq", "adq"},
		{".ADQ", "adq"},
		{"adq", "adq"},
		{" .adq ", "adq"},
		{".Adm", "adm"},
		{".gcs", ""},
		{"", ""},
		{".", ""},
		{"nonsense", ""},
	} {
		c.Equal(one.want, gurps.ListFilterKeyForExtension(one.in), "key for %q", one.in)
	}

	keys := gurps.ListFilterKeys()
	c.Equal(7, len(keys), "there is one key per library list type")
	c.True(slices.IsSorted(keys), "the keys are sorted")
	c.Equal(len(keys), len(slices.Compact(slices.Clone(keys))), "the keys are unique")
	for _, want := range []string{"adq", "adm", "skl", "spl", "eqp", "eqm", "not"} {
		c.True(slices.Contains(keys, want), "the keys include %q", want)
	}
}
