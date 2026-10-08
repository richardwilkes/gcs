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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	traitcontainer "github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/eqcontainer"
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

// matchesListFilter reports whether the node passes the filter, building a matcher for just that node. The tests
// apply each filter to one node, so the matcher's once-per-filter field lookup buys them nothing.
func matchesListFilter[T gurps.Node[T]](f *gurps.ListFilter, fields []*gurps.FilterField[T], node T) bool {
	return gurps.NewListFilterMatcher(f, fields)(node)
}

// findFilterField returns the field with the given key, or nil if there is none.
func findFilterField[T gurps.Node[T]](fields []*gurps.FilterField[T], key string) *gurps.FilterField[T] {
	for _, field := range fields {
		if field.Key == key {
			return field
		}
	}
	return nil
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
// both in what it writes and in the parent links that the JSON doesn't carry and the load has to restore, and that a
// padded name is trimmed on the way in.
func TestListFilterJSONRoundTrip(t *testing.T) {
	c := check.New(t)

	f := gurps.NewListFilter("Round Trip")
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
	c.Equal("Round Trip", restored.Name, "the name should survive the round trip")
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

	// A name padded with whitespace on disk, which NewListFilter would already have trimmed, is trimmed as it loads.
	var padded gurps.ListFilter
	c.NoError(jio.Unmarshal([]byte(`{"name": "  Padded  ", "root": {"type": "group", "all": true}}`), &padded),
		"a filter with a padded name should unmarshal")
	c.Equal("Padded", padded.Name, "loading trims the name")
}

// TestListFilterUnknownNodePreserved verifies that a filter node whose type this build doesn't recognize loads as an
// UnknownFilterNode, is written back out unchanged so a newer version of GCS can still use it, and participates in the
// filter's hash rather than being invisible to it.
func TestListFilterUnknownNodePreserved(t *testing.T) {
	c := check.New(t)

	const knownChild = `{"type": "condition", "field": "name", "text": {"compare": "is", "qualifier": "Alertness"}}`
	const unknownChild = `{"type": "future_node", "weird": [1, 2]}`
	filterWith := func(children ...string) string {
		return `{"name": "Filter", "root": {"type": "group", "all": true, "children": [` +
			strings.Join(children, `, `) + `]}}`
	}
	withUnknown := filterWith(knownChild, unknownChild)

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

	// The hash has to see inside the unknown node, so the comparison filters differ from f only in the unknown node's
	// data, or only in its kind; a filter with one child fewer would hash differently on the count alone.
	var otherData gurps.ListFilter
	c.NoError(jio.Unmarshal([]byte(filterWith(knownChild, `{"type": "future_node", "weird": [1, 3]}`)), &otherData),
		"the filter with other data in its unknown node should load")
	c.NotEqual(gurps.Hash64(&otherData), gurps.Hash64(&f), "an unknown node's data must contribute to the filter's hash")
	var otherKind gurps.ListFilter
	c.NoError(jio.Unmarshal([]byte(filterWith(knownChild, `{"type": "other_future_node", "weird": [1, 2]}`)),
		&otherKind), "the filter with another kind of unknown node should load")
	c.NotEqual(gurps.Hash64(&otherKind), gurps.Hash64(&f), "an unknown node's kind must contribute to the filter's hash")
	var same gurps.ListFilter
	c.NoError(jio.Unmarshal([]byte(withUnknown), &same), "the filter should load a second time")
	c.Equal(gurps.Hash64(&same), gurps.Hash64(&f), "the same unknown node hashes the same")

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
			// NodeType is a constant on each concrete type, so the decoded type is read from the struct itself.
			switch node := nodes[0].(type) {
			case *gurps.FilterGroup:
				c.Equal(filternode.Group, one, "only a group decodes to a group")
				c.Equal(one, node.Type, "the decoded type is kept on the group")
			case *gurps.FilterCondition:
				c.Equal(filternode.Condition, one, "only a condition decodes to a condition")
				c.Equal(one, node.Type, "the decoded type is kept on the condition")
			case *gurps.UnknownFilterNode:
				c.Equal(filternode.Unknown, one, "only the reserved type decodes to the unknown wrapper")
				c.Equal(one.Key(), node.Kind, "the unknown wrapper keeps the type it found")
			default:
				c.Errorf("unexpected node type %T", node)
			}
			c.Equal(one, nodes[0].NodeType(), "the node reports the type it was decoded as")
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
	trait.LocalNotes = "Sharp"
	c.Equal(fxp.FromInteger(5), trait.AdjustedPoints(nil), "the trait should be worth the points it was given")
	c.Equal("Sharp", trait.LocalNotesWithReplacements(), "the trait should have the notes it was given")

	untagged := gurps.NewTrait(nil, nil, false)
	untagged.Name = "Untagged"
	c.Equal(0, len(untagged.TagList()), "the untagged trait should have no tags")
	c.Equal("", untagged.LocalNotesWithReplacements(), "the untagged trait should have no notes")

	spaced := gurps.NewTrait(nil, nil, false)
	spaced.Name = "Spaced"
	spaced.LocalNotes = "  "
	c.Equal("  ", spaced.LocalNotesWithReplacements(), "the spaced trait's notes should be only space")

	blankTagged := gurps.NewTrait(nil, nil, false)
	blankTagged.Name = "Blank Tagged"
	blankTagged.Tags = []string{"", "  "}
	c.Equal(2, len(blankTagged.TagList()), "the blank tagged trait should have only blank tags")

	shield := gurps.NewTrait(nil, nil, false)
	shield.Name = "Shield"
	shield.Tags = []string{"", "Shield"}
	c.Equal(2, len(shield.TagList()), "the shield trait should have a blank tag and a tag")

	container := gurps.NewTrait(nil, nil, true)
	container.Name = "Group"
	c.True(container.Container(), "the container trait should be a container")

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
		// An item with no tags doesn't have tags, so no comparison holds for it, though a negated one does.
		{"list is, empty tag list", untagged, newTextFilterCondition("tags", criteria.IsText, "Mental"), false},
		{"list is not, empty tag list", untagged, newTextFilterCondition("tags", criteria.IsNotText, "Mental"), false},
		{"list is empty, empty tag list", untagged, newTextFilterCondition("tags", criteria.IsText, ""), false},
		{"list anything, empty tag list", untagged, gurps.NewFilterCondition(nil, "tags"), false},
		{"list anything, tagged", trait, gurps.NewFilterCondition(nil, "tags"), true},
		{"not list anything, empty tag list", untagged, negated(gurps.NewFilterCondition(nil, "tags")), true},
		{"not list anything, tagged", trait, negated(gurps.NewFilterCondition(nil, "tags")), false},

		// A tag that is empty or only space doesn't count, so it never matches, and one holding only those has none.
		{"list anything, only blank tags", blankTagged, gurps.NewFilterCondition(nil, "tags"), false},
		{"list is empty, only blank tags", blankTagged, newTextFilterCondition("tags", criteria.IsText, ""), false},
		{
			"list is not, only blank tags", blankTagged,
			newTextFilterCondition("tags", criteria.IsNotText, "Mental"), false,
		},
		{"not list anything, only blank tags", blankTagged, negated(gurps.NewFilterCondition(nil, "tags")), true},
		{"list is, a blank tag and a tag", shield, newTextFilterCondition("tags", criteria.IsText, "Shield"), true},
		{"list is empty, a blank tag and a tag", shield, newTextFilterCondition("tags", criteria.IsText, ""), false},
		{"list is not, a blank tag and a tag", shield, newTextFilterCondition("tags", criteria.IsNotText, "Mental"), true},
		{"list anything, a blank tag and a tag", shield, gurps.NewFilterCondition(nil, "tags"), true},

		// Text that is empty, or only space, is not had either.
		{"text anything, no notes", untagged, gurps.NewFilterCondition(nil, "notes"), false},
		{"text is empty, no notes", untagged, newTextFilterCondition("notes", criteria.IsText, ""), false},
		{
			"text does not contain, no notes", untagged,
			newTextFilterCondition("notes", criteria.DoesNotContainText, "x"), false,
		},
		{"not text anything, no notes", untagged, negated(gurps.NewFilterCondition(nil, "notes")), true},
		{"not text anything, space only", spaced, negated(gurps.NewFilterCondition(nil, "notes")), true},
		{"text anything, notes", trait, gurps.NewFilterCondition(nil, "notes"), true},

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
		c.Equal(one.want, matchesListFilter(newTestListFilter(one.cond), gurps.TraitFilterFields(), one.trait),
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
			matchesListFilter(newTestListFilter(one.cond), gurps.EquipmentFilterFields(), equipment),
			"%s", one.name)
	}
}

// TestListFilterTraitLevels checks that a trait that can't be leveled, or a container, has no levels to a filter, so it
// satisfies no criteria on them and a negated condition finds it, while a leveled trait at 0 levels has levels of 0.
// A disabled leveled trait has levels, read as 0 as its points are.
func TestListFilterTraitLevels(t *testing.T) {
	c := check.New(t)
	fields := gurps.TraitFilterFields()

	leveled := gurps.NewTrait(nil, nil, false)
	leveled.Name = "Leveled"
	leveled.CanLevel = true
	leveled.Levels = fxp.Two
	c.True(leveled.IsLeveled(), "the leveled trait should be leveled")
	c.Equal(fxp.Two, leveled.CurrentLevel(), "the leveled trait should have two levels")

	atZero := gurps.NewTrait(nil, nil, false)
	atZero.Name = "At Zero"
	atZero.CanLevel = true
	c.True(atZero.IsLeveled(), "the trait at zero should be leveled")
	c.Equal(fxp.Int(0), atZero.CurrentLevel(), "the trait at zero should have no levels bought")

	plain := gurps.NewTrait(nil, nil, false)
	plain.Name = "Plain"
	c.False(plain.IsLeveled(), "the plain trait should not be leveled")

	container := gurps.NewTrait(nil, nil, true)
	container.Name = "Group"
	container.CanLevel = true
	c.False(container.IsLeveled(), "a container should not be leveled, even when it can level")

	disabled := gurps.NewTrait(nil, nil, false)
	disabled.Name = "Disabled"
	disabled.CanLevel = true
	disabled.Levels = fxp.Two
	disabled.Disabled = true
	c.True(disabled.IsLeveled(), "the disabled trait should be leveled")
	c.Equal(fxp.Int(0), disabled.CurrentLevel(), "the disabled trait should read zero levels")

	anything := newNumberFilterCondition("levels", criteria.AnyNumber, 0)
	notAnything := newNumberFilterCondition("levels", criteria.AnyNumber, 0)
	notAnything.Not = true
	isZero := newNumberFilterCondition("levels", criteria.EqualsNumber, 0)
	atMostOne := newNumberFilterCondition("levels", criteria.AtMostNumber, fxp.One)
	for _, one := range []struct {
		name  string
		cond  *gurps.FilterCondition
		shown map[*gurps.Trait]bool
	}{
		{"levels that are anything", anything, map[*gurps.Trait]bool{
			leveled: true, atZero: true, plain: false, container: false, disabled: true,
		}},
		{"not levels that are anything", notAnything, map[*gurps.Trait]bool{
			leveled: false, atZero: false, plain: true, container: true, disabled: false,
		}},
		{"levels that are 0", isZero, map[*gurps.Trait]bool{
			leveled: false, atZero: true, plain: false, container: false, disabled: true,
		}},
		{"levels that are at most 1", atMostOne, map[*gurps.Trait]bool{
			leveled: false, atZero: true, plain: false, container: false, disabled: true,
		}},
	} {
		f := newTestListFilter(one.cond)
		for trait, want := range one.shown {
			c.Equal(want, matchesListFilter(f, fields, trait), "%s: %s", one.name, trait.Name)
		}
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
	c.False(matchesListFilter(newTestListFilter(matching), fields, trait),
		"negating a condition that matches rejects the node")
	notMatching := newTextFilterCondition("name", criteria.IsText, "Acute Vision")
	notMatching.Not = true
	c.True(matchesListFilter(newTestListFilter(notMatching), fields, trait),
		"negating a condition that doesn't match accepts the node")

	// On a group.
	sub := gurps.NewFilterGroup(nil)
	sub.Not = true
	sub.Children = gurps.FilterNodes{newTextFilterCondition("name", criteria.IsText, "Alertness")}
	c.False(matchesListFilter(newTestListFilter(sub), fields, trait),
		"negating a group whose children match rejects the node")
	sub.Children = gurps.FilterNodes{newTextFilterCondition("name", criteria.IsText, "Acute Vision")}
	c.True(matchesListFilter(newTestListFilter(sub), fields, trait),
		"negating a group whose children don't match accepts the node")

	// A negated "is not" is just an "is".
	negatedIsNot := newTextFilterCondition("name", criteria.IsNotText, "Alertness")
	negatedIsNot.Not = true
	plainIs := newTestListFilter(newTextFilterCondition("name", criteria.IsText, "Alertness"))
	for _, name := range []string{"Alertness", "Acute Vision"} {
		trait.Name = name
		c.Equal(matchesListFilter(plainIs, fields, trait),
			matchesListFilter(newTestListFilter(negatedIsNot), fields, trait),
			`a negated "is not" behaves as a plain "is" for %q`, name)
	}
	trait.Name = "Alertness"

	// And on a list condition, where the criteria's own negation has list semantics of its own.
	absent := newTextFilterCondition("tags", criteria.DoesNotContainText, "Social")
	c.True(matchesListFilter(newTestListFilter(absent), fields, trait),
		`"does not contain" passes when no tag contains the qualifier`)
	absent.Not = true
	c.False(matchesListFilter(newTestListFilter(absent), fields, trait),
		`negating a passing "does not contain" rejects the node`)
	present := newTextFilterCondition("tags", criteria.DoesNotContainText, "ment")
	c.False(matchesListFilter(newTestListFilter(present), fields, trait),
		`"does not contain" fails when a tag contains the qualifier`)
	present.Not = true
	c.True(matchesListFilter(newTestListFilter(present), fields, trait),
		`negating a failing "does not contain" accepts the node`)
}

// TestListFilterCheckResults verifies that a filter is checked as a list of prerequisites is: a condition is met or
// unmet, which its negation swaps, a condition on an unknown field or a node of an unknown kind fails, and a group
// leaves out the children that are skipped and is skipped when nothing is left. Of the rest, an unmet child decides an
// "all of" group and a met one an "any of" group, even beside a failed one; otherwise the group has failed when one of
// the rest has, and is met or unmet as they all are. Negating a group then swaps met and unmet only. A node is shown
// when the filter is met or skipped.
func TestListFilterCheckResults(t *testing.T) {
	c := check.New(t)
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Alertness"
	fields := gurps.TraitFilterFields()
	met := func() *gurps.FilterCondition { return newTextFilterCondition("name", criteria.IsText, "Alertness") }
	unmet := func() *gurps.FilterCondition { return newTextFilterCondition("name", criteria.IsText, "Acute Vision") }
	unknownField := func() *gurps.FilterCondition { return gurps.NewFilterCondition(nil, "field_from_a_newer_gcs") }
	notCond := func(cond *gurps.FilterCondition) *gurps.FilterCondition {
		cond.Not = true
		return cond
	}
	group := func(all, not bool, children ...gurps.FilterNode) *gurps.FilterGroup {
		g := gurps.NewFilterGroup(nil)
		g.All, g.Not = all, not
		g.Children = children
		return g
	}
	empty := func() *gurps.FilterGroup { return group(true, false) }
	emptyNoneOf := func() *gurps.FilterGroup { return group(false, true) }
	for _, one := range []struct {
		name  string
		root  *gurps.FilterGroup
		want  gurps.CheckResult
		shown bool
	}{
		{"empty root", empty(), gurps.CheckSkipped, true},
		{"empty negated root", emptyNoneOf(), gurps.CheckSkipped, true},
		{"met", group(true, false, met()), gurps.CheckMet, true},
		{"unmet", group(true, false, unmet()), gurps.CheckUnmet, false},
		{"negated met", group(true, false, notCond(met())), gurps.CheckUnmet, false},
		{"negated unmet", group(true, false, notCond(unmet())), gurps.CheckMet, true},
		{"unknown field", group(true, false, unknownField()), gurps.CheckFailed, false},
		{"negated unknown field", group(true, false, notCond(unknownField())), gurps.CheckFailed, false},
		{"unknown node", group(true, false, newTestUnknownFilterNode()), gurps.CheckFailed, false},

		{"all of met and met", group(true, false, met(), met()), gurps.CheckMet, true},
		{"all of met and unmet", group(true, false, met(), unmet()), gurps.CheckUnmet, false},
		{"all of met and unknown field", group(true, false, met(), unknownField()), gurps.CheckFailed, false},
		{"all of unmet and unknown field", group(true, false, unmet(), unknownField()), gurps.CheckUnmet, false},
		{
			"all of unmet and unknown node", group(true, false, unmet(), newTestUnknownFilterNode()),
			gurps.CheckUnmet, false,
		},
		{"all of met and empty", group(true, false, met(), empty()), gurps.CheckMet, true},
		{"all of unmet and empty", group(true, false, unmet(), empty()), gurps.CheckUnmet, false},
		{"any of met and unmet", group(false, false, met(), unmet()), gurps.CheckMet, true},
		{"any of unmet and unmet", group(false, false, unmet(), unmet()), gurps.CheckUnmet, false},
		{"any of met and unknown field", group(false, false, met(), unknownField()), gurps.CheckMet, true},
		{"any of unmet and unknown field", group(false, false, unmet(), unknownField()), gurps.CheckFailed, false},
		{"any of met and unknown node", group(false, false, met(), newTestUnknownFilterNode()), gurps.CheckMet, true},
		{
			"any of unmet and unknown node", group(false, false, unmet(), newTestUnknownFilterNode()),
			gurps.CheckFailed, false,
		},
		{"any of unmet and empty", group(false, false, unmet(), empty()), gurps.CheckUnmet, false},

		{"not all of met and met", group(true, true, met(), met()), gurps.CheckUnmet, false},
		{"not all of met and unmet", group(true, true, met(), unmet()), gurps.CheckMet, true},
		{"not all of met and unknown field", group(true, true, met(), unknownField()), gurps.CheckFailed, false},
		{"not all of unmet and unknown field", group(true, true, unmet(), unknownField()), gurps.CheckMet, true},
		{"none of met and unknown field", group(false, true, met(), unknownField()), gurps.CheckUnmet, false},
		{"none of unmet and unknown field", group(false, true, unmet(), unknownField()), gurps.CheckFailed, false},
		{"none of met and unmet", group(false, true, met(), unmet()), gurps.CheckUnmet, false},
		{"none of unmet and unmet", group(false, true, unmet(), unmet()), gurps.CheckMet, true},
		{
			"none of met and unknown node", group(false, true, met(), newTestUnknownFilterNode()),
			gurps.CheckUnmet, false,
		},
		{
			"none of unmet and unknown node", group(false, true, unmet(), newTestUnknownFilterNode()),
			gurps.CheckFailed, false,
		},

		{"nested empty group", group(true, false, empty()), gurps.CheckSkipped, true},
		{"nested negated empty group", group(true, false, emptyNoneOf()), gurps.CheckSkipped, true},
		{
			"group of empty groups", group(false, false, group(true, false, empty(), emptyNoneOf())),
			gurps.CheckSkipped, true,
		},
		{
			"met beside a group of empty groups", group(true, false, met(), group(true, true, empty())),
			gurps.CheckMet, true,
		},
		{
			"nested failed group in any of", group(false, false, group(true, false, newTestUnknownFilterNode()), met()),
			gurps.CheckMet, true,
		},
		{
			"nested failed group in all of", group(true, false, group(false, true, unknownField()), met()),
			gurps.CheckFailed, false,
		},
	} {
		f := gurps.NewListFilter("Test")
		f.Root = one.root
		f.EnsureValidity()
		c.Equal(one.want, gurps.NewListFilterChecker(f, fields)(trait), "%s", one.name)
		c.Equal(one.shown, matchesListFilter(f, fields, trait), "%s shown", one.name)
	}
	c.Equal(gurps.CheckSkipped, gurps.NewListFilterChecker(nil, fields)(trait), "no filter is skipped")
	c.True(matchesListFilter(nil, fields, trait), "so it shows everything")
	c.Nil(findFilterField(fields, "field_from_a_newer_gcs"), "the field really is unknown")
}

// checkFilterFields verifies that a filter field table is well formed: every field has a unique, storable key and a
// title, exactly one accessor, which is the one its kind names, and an accessor that can actually be run against a
// node of the type the table is for. A field that offers suggestions holds text and can make them with or without
// nodes.
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

		if field.Suggestions != nil {
			c.True(field.Kind == gurps.FilterFieldText || field.Kind == gurps.FilterFieldList,
				"%s: field %q offers suggestions, so it should hold text", name, field.Key)
			c.NotPanics(func() { _ = field.Suggestions(samples) },
				"%s: field %q should make suggestions from the samples", name, field.Key)
			c.NotPanics(func() { _ = field.Suggestions(nil) },
				"%s: field %q should make suggestions from no nodes", name, field.Key)
		}
	}
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

// suggestingKeys returns the keys of the fields that offer suggestions, in table order.
func suggestingKeys[T gurps.Node[T]](fields []*gurps.FilterField[T]) []string {
	keys := make([]string, 0, len(fields))
	for _, field := range fields {
		if field.Suggestions != nil {
			keys = append(keys, field.Key)
		}
	}
	return keys
}

// TestFilterFieldSuggestionSources pins the fields that offer suggestions for a condition's value, so that one gaining
// or losing them is noticed.
func TestFilterFieldSuggestionSources(t *testing.T) {
	c := check.New(t)
	c.Equal([]string{"tags", "container_type"}, suggestingKeys(gurps.TraitFilterFields()), "traits")
	c.Equal([]string{"tags"}, suggestingKeys(gurps.TraitModifierFilterFields()), "trait modifiers")
	c.Equal([]string{"tags", "difficulty"}, suggestingKeys(gurps.SkillFilterFields()), "skills")
	c.Equal([]string{"tags", "difficulty", "college", "power_source", "class", "resist"},
		suggestingKeys(gurps.SpellFilterFields()), "spells")
	c.Equal([]string{"tags", "tech_level", "legality_class", "container_type"},
		suggestingKeys(gurps.EquipmentFilterFields()), "equipment")
	c.Equal([]string{"tags", "tech_level"}, suggestingKeys(gurps.EquipmentModifierFilterFields()),
		"equipment modifiers")
	c.Equal([]string{"tags"}, suggestingKeys(gurps.NoteFilterFields()), "notes")
}

// TestFilterFieldListedValues verifies that a field offering the values found in the list being filtered looks at
// every node a filter checks, nested, container and disabled ones included, but not those that lack the field, and
// offers each value once, trimmed and in its first spelling, sorted naturally without regard to case. Blank values
// are left out, as are those of a list field that hold a comma, since the qualifier is split on commas.
func TestFilterFieldListedValues(t *testing.T) {
	c := check.New(t)

	root := gurps.NewTrait(nil, nil, true)
	root.Name = "Root"
	root.Tags = []string{"Magic"}
	child := gurps.NewTrait(nil, root, false)
	child.Name = "Bow, Long"
	child.Tags = []string{"magic", " Zebra ", "", "Odd, Tag"}
	nested := gurps.NewTrait(nil, root, true)
	nested.Name = "Nested"
	grandchild := gurps.NewTrait(nil, nested, false)
	grandchild.Name = "Arrow"
	grandchild.Tags = []string{"apple"}
	grandchild.Disabled = true
	nested.Children = []*gurps.Trait{grandchild}
	root.Children = []*gurps.Trait{child, nested}
	list := []*gurps.Trait{root}

	tags := findFilterField(gurps.TraitFilterFields(), "tags")
	c.Equal([]string{"apple", "Magic", "Zebra"}, tags.Suggestions(list), "the tags in the list are offered")
	c.Equal(0, len(tags.Suggestions(nil)), "no list offers no tags")

	named := gurps.NewTextFilterField("name", "have a name", func(trait *gurps.Trait) string { return trait.Name }).
		WithPresence(func(trait *gurps.Trait) bool { return !trait.Container() }).WithListedValues()
	c.Equal([]string{"Arrow", "Bow, Long"}, named.Suggestions(list),
		"a text field keeps a value holding a comma, and the nodes that lack the field are left out")

	equipment := make([]*gurps.Equipment, 0, 5)
	for _, techLevel := range []string{"10", "9", "2", "", "^"} {
		e := gurps.NewEquipment(nil, nil, false)
		e.TechLevel = techLevel
		equipment = append(equipment, e)
	}
	techLevels := findFilterField(gurps.EquipmentFilterFields(), "tech_level")
	c.Equal([]string{"2", "9", "10", "^"}, techLevels.Suggestions(equipment), "tech levels are sorted naturally")
	c.Equal(0, len(techLevels.Suggestions(nil)), "no list offers no tech levels")
}

// checkContainerTypeChoices verifies that the field offers the String() of each of the types, in order, whatever the
// list being filtered holds, and that what it reports for a container of each type is among them.
func checkContainerTypeChoices[T gurps.Node[T], E fmt.Stringer](c check.Checker, name string,
	field *gurps.FilterField[T], types []E, newContainer func(E) T,
) {
	want := make([]string, 0, len(types))
	containers := make([]T, 0, len(types))
	for _, one := range types {
		want = append(want, one.String())
		containers = append(containers, newContainer(one))
	}
	choices := field.Suggestions(nil)
	c.Equal(want, choices, "%s: every container type is offered", name)
	c.Equal(want, field.Suggestions(containers), "%s: the list being filtered changes nothing", name)
	for i, one := range containers {
		c.True(slices.Contains(choices, field.Text(one)), "%s: the type of a %s container is offered", name, types[i])
	}
}

// TestFilterFieldContainerTypeChoices verifies that the container type fields offer every container type, spelled as
// the fields report a container's type.
func TestFilterFieldContainerTypeChoices(t *testing.T) {
	c := check.New(t)
	checkContainerTypeChoices(c, "traits", findFilterField(gurps.TraitFilterFields(), "container_type"),
		traitcontainer.Types, func(one traitcontainer.Type) *gurps.Trait {
			trait := gurps.NewTrait(nil, nil, true)
			trait.ContainerType = one
			return trait
		})
	checkContainerTypeChoices(c, "equipment", findFilterField(gurps.EquipmentFilterFields(), "container_type"),
		eqcontainer.Types, func(one eqcontainer.Type) *gurps.Equipment {
			e := gurps.NewEquipment(nil, nil, true)
			e.ContainerType = one
			return e
		})
}

// TestFilterFieldDifficultyChoices verifies that the difficulty fields offer each attribute, and 10, at each difficulty
// level, followed by the bare levels of a technique, spelled as the fields report a skill's or spell's difficulty.
func TestFilterFieldDifficultyChoices(t *testing.T) {
	c := check.New(t)
	// The tests load the user's settings file, whose attributes may be anything.
	sheet := gurps.GlobalSettings().Sheet
	saved := sheet.Attributes
	t.Cleanup(func() { sheet.Attributes = saved })
	sheet.Attributes = gurps.FactoryAttributeDefs()

	skillField := findFilterField(gurps.SkillFilterFields(), "difficulty")
	spellField := findFilterField(gurps.SpellFilterFields(), "difficulty")
	choices := skillField.Suggestions(nil)
	for _, want := range []string{"10/E", "DX/A", "IQ/H", "HT/VH", "ST/W"} {
		c.True(slices.Contains(choices, want), "%q is offered", want)
	}
	c.True(len(choices) >= 2, "there are at least the technique levels")
	if len(choices) >= 2 {
		c.Equal([]string{"A", "H"}, choices[len(choices)-2:], "the technique levels come last")
	}
	attributes := 1 // 10
	for _, def := range sheet.Attributes.List(true) {
		if def.DefID != gurps.DodgeID {
			attributes++
		}
	}
	c.Equal(attributes*len(difficulty.Levels)+len(difficulty.TechniqueLevels), len(choices),
		"each attribute at each level and the technique levels are offered, and nothing else")
	c.Equal(choices, spellField.Suggestions(nil), "skills and spells offer the same difficulties")

	tenEasy := gurps.NewSkill(nil, nil, false)
	tenEasy.Difficulty.Attribute = "10"
	tenEasy.Difficulty.Difficulty = difficulty.Easy
	hardTechnique := gurps.NewTechnique(nil, nil, "")
	hardTechnique.Difficulty.Difficulty = difficulty.Hard
	veryHardSpell := gurps.NewSpell(nil, nil, false)
	veryHardSpell.Difficulty.Difficulty = difficulty.VeryHard
	bareRitual := gurps.NewRitualMagicSpell(nil, nil, false)
	bareRitual.Difficulty.Attribute = ""
	for _, one := range []struct {
		want  string
		value string
	}{
		{"DX/A", skillField.Text(gurps.NewSkill(nil, nil, false))},
		{"10/E", skillField.Text(tenEasy)},
		{"A", skillField.Text(gurps.NewTechnique(nil, nil, ""))},
		{"H", skillField.Text(hardTechnique)},
		{"IQ/VH", spellField.Text(veryHardSpell)},
		{"IQ/H", spellField.Text(gurps.NewRitualMagicSpell(nil, nil, false))},
		{"H", spellField.Text(bareRitual)},
	} {
		c.Equal(one.want, one.value, "the field reports the difficulty")
		c.True(slices.Contains(choices, one.value), "%q is offered", one.value)
	}
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

	// Filters are ordered by name, ignoring case, and numbers within a name compare as numbers. The names are chosen
	// so that a byte-wise sort, which puts every upper-case letter ahead of every lower-case one, would order them
	// differently.
	for _, name := range []string{"b", "A", "C"} {
		s.AddListFilter(key, gurps.NewListFilter(name))
	}
	c.Equal([]string{"A", "b", "C"}, listFilterNames(s.ListFiltersFor(key)),
		"saved filters are sorted without regard to case")
	s.AddListFilter(key, gurps.NewListFilter("item10"))
	s.AddListFilter(key, gurps.NewListFilter("item9"))
	c.Equal([]string{"A", "b", "C", "item9", "item10"}, listFilterNames(s.ListFiltersFor(key)),
		"saved filters are sorted naturally, so item9 precedes item10")

	// The returned slice is a copy, so a caller can't reorder or shorten what is stored.
	list := s.ListFiltersFor(key)
	list = append(list, gurps.NewListFilter("intruder"))
	list[0] = nil
	c.Equal([]string{"A", "b", "C", "item9", "item10"}, listFilterNames(s.ListFiltersFor(key)),
		"the returned slice is a copy of the stored one")

	// A filter renamed in place stays where it was until the list is sorted again.
	s.ListFiltersFor(key)[1].Name = "zzz"
	c.Equal([]string{"A", "zzz", "C", "item9", "item10"}, listFilterNames(s.ListFiltersFor(key)),
		"renaming in place doesn't move the filter by itself")
	s.ResortListFilters(key)
	c.Equal([]string{"A", "C", "item9", "item10", "zzz"}, listFilterNames(s.ListFiltersFor(key)),
		"re-sorting puts the renamed filter in its place")
	s.ListFiltersFor(key)[4].Name = "b"
	s.ResortListFilters(key)
	c.Equal([]string{"A", "b", "C", "item9", "item10"}, listFilterNames(s.ListFiltersFor(key)),
		"re-sorting puts it back again")
	s.ResortListFilters("never stored")
	_, exists := s.ListFilters["never stored"]
	c.False(exists, "re-sorting a key that isn't stored doesn't create it")

	// Removal is by pointer identity.
	first := s.ListFiltersFor(key)[0]
	s.RemoveListFilter(key, first)
	c.Equal([]string{"b", "C", "item9", "item10"}, listFilterNames(s.ListFiltersFor(key)),
		"removing drops just that filter")
	c.Equal(-1, slices.Index(s.ListFiltersFor(key), first), "the removed filter is gone")
	s.RemoveListFilter(key, gurps.NewListFilter("b"))
	c.Equal([]string{"b", "C", "item9", "item10"}, listFilterNames(s.ListFiltersFor(key)),
		"removing a filter that isn't stored changes nothing, even when it bears a stored filter's name")

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
	_, exists = s.ListFilters[key]
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

// TestListFilterMatchesEquipmentContainerType verifies that a filter on equipment can tell a group from a physical
// container, as one on traits can tell their kinds of container apart, and that a piece of equipment that isn't a
// container has no container type.
func TestListFilterMatchesEquipmentContainerType(t *testing.T) {
	c := check.New(t)
	group := gurps.NewEquipmentGroup(nil, nil)
	backpack := gurps.NewEquipment(nil, nil, true)
	item := gurps.NewEquipment(nil, nil, false)
	isGroup := newTextFilterCondition("container_type", criteria.IsText, "group")
	isContainer := newTextFilterCondition("container_type", criteria.IsText, "container")
	fields := gurps.EquipmentFilterFields()
	c.True(matchesListFilter(newTestListFilter(isGroup), fields, group))
	c.False(matchesListFilter(newTestListFilter(isGroup), fields, backpack))
	c.True(matchesListFilter(newTestListFilter(isContainer), fields, backpack))
	c.False(matchesListFilter(newTestListFilter(isContainer), fields, item), "an item is no kind of container")
}
