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
	"strconv"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/frequency"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selector"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stdmg"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// makeOverride builds a SelectorOverride (which implements gurps.Override) with the given priority and a number of
// active criteria, so it exposes a known specificity to the resolver.
func makeOverride(value string, priority, specificity int) *gurps.SelectorOverride {
	o := gurps.NewSelectorOverride(selector.WeaponDamageType)
	o.Value = value
	o.Priority = priority
	// Start from a fully-open matcher (specificity 0), then pin criteria to raise it.
	o.NameCriteria.Compare = criteria.AnyText
	pins := []*criteria.Text{&o.NameCriteria, &o.UsageCriteria, &o.TagsCriteria}
	for i := 0; i < specificity && i < len(pins); i++ {
		pins[i].Compare = criteria.IsText
		pins[i].Qualifier = "x"
	}
	return o
}

func candidate(o *gurps.SelectorOverride) gurps.OverrideCandidate[string] {
	return gurps.OverrideCandidate[string]{Value: o.Value, Override: o}
}

func TestResolveOverride(t *testing.T) {
	c := check.New(t)
	identity := func(s string) string { return s }

	// No candidates: the base value survives.
	c.Equal("cr", gurps.ResolveOverride("cr", nil, identity, nil), "no override keeps base")

	// A single ordinary (priority 0) override still replaces the base.
	c.Equal("burn", gurps.ResolveOverride("cr",
		[]gurps.OverrideCandidate[string]{candidate(makeOverride("burn", 0, 0))}, identity, nil),
		"priority-0 override still beats base")

	// Highest priority wins regardless of listing order.
	c.Equal("burn", gurps.ResolveOverride("cr", []gurps.OverrideCandidate[string]{
		candidate(makeOverride("imp", 0, 0)),
		candidate(makeOverride("burn", 10, 0)),
	}, identity, nil), "highest priority wins")

	// Equal priority: the more specific match wins.
	c.Equal("burn", gurps.ResolveOverride("cr", []gurps.OverrideCandidate[string]{
		candidate(makeOverride("imp", 5, 1)),
		candidate(makeOverride("burn", 5, 3)),
	}, identity, nil), "more specific wins the tie")

	// Equal priority and specificity but different values: a real conflict. Resolution is deterministic (first by
	// rendered value, so "burn" < "imp") and the tooltip flags it, independent of input order.
	var tooltip xbytes.InsertBuffer
	got := gurps.ResolveOverride("cr", []gurps.OverrideCandidate[string]{
		candidate(makeOverride("imp", 5, 2)),
		candidate(makeOverride("burn", 5, 2)),
	}, identity, &tooltip)
	c.Equal("burn", got, "conflict resolves deterministically")
	c.True(strings.Contains(tooltip.String(), "conflict"), "conflict is flagged in tooltip")
}

// TestSelectorFieldDescriptors guards the field wiring: every selector.Field must have a descriptor, and a constrained
// field's suggested states must all be valid stored values with a non-empty picker label. This catches enum-key typos
// in the descriptor table.
func TestSelectorFieldDescriptors(t *testing.T) {
	c := check.New(t)
	for _, field := range selector.Fields {
		d := gurps.SelectorFieldDescriptorFor(field)
		c.Equal(field, d.Field, "descriptor is registered for %v", field)
		if d.FreeForm || d.StateTitle == nil {
			continue
		}
		for _, state := range d.SuggestedStates {
			// A constrained state must produce a non-empty picker label.
			c.True(d.StateTitle(state) != "", "state %q of %v has a label", state, field)
		}
	}

	// The strength-basis states must map to distinct, valid stdmg options (a duplicated or misspelled key would collapse
	// two picker entries onto the same option).
	d := gurps.SelectorFieldDescriptorFor(selector.WeaponDamageStrengthBasis)
	seen := make(map[stdmg.Option]bool)
	for _, state := range d.SuggestedStates {
		opt := stdmg.ExtractOption(state)
		c.Equal(state, opt.Key(), "state %q round-trips through stdmg", state)
		c.False(seen[opt], "state %q is not a duplicate", state)
		seen[opt] = true
	}

	// The numeric damage fields validate fixed-point input; the dice-spec fields are free-form with no validator.
	for _, field := range []selector.Field{
		selector.WeaponArmorDivisor, selector.WeaponFragmentationArmorDivisor,
		selector.WeaponDamageStrengthMultiplier, selector.WeaponDamagePerDieModifier,
	} {
		nd := gurps.SelectorFieldDescriptorFor(field)
		c.True(nd.Validate != nil, "%v has a validator", field)
		if nd.Validate != nil {
			c.True(nd.Validate("2"), "%v accepts a number", field)
			c.False(nd.Validate("abc"), "%v rejects a non-number", field)
		}
	}
	for _, field := range []selector.Field{
		selector.WeaponBaseDamageDice, selector.WeaponBaseDamageDicePerLevel, selector.WeaponFragmentationDice,
	} {
		c.True(gurps.SelectorFieldDescriptorFor(field).Validate == nil, "%v is free-form (no validator)", field)
	}

	// Scope: weapon fields default to the weapon scope; the trait frequency field is trait-scoped.
	c.Equal(gurps.SelectorScopeWeapon, gurps.SelectorFieldDescriptorFor(selector.WeaponDamageType).Scope, "damage type is weapon-scoped")
	freq := gurps.SelectorFieldDescriptorFor(selector.TraitFrequency)
	c.Equal(gurps.SelectorScopeTrait, freq.Scope, "trait frequency is trait-scoped")

	// Every frequency state must be a valid roll whose label round-trips (a bad key would mislabel or drop a choice).
	c.Equal(len(frequency.Rolls), len(freq.SuggestedStates), "one state per frequency roll")
	for _, state := range freq.SuggestedStates {
		n, err := strconv.Atoi(state)
		c.True(err == nil, "state %q is numeric", state)
		roll := frequency.Roll(n)
		c.Equal(roll, roll.EnsureValid(), "state %q is a valid frequency roll", state)
		c.Equal(roll.String(), freq.StateTitle(state), "state %q label matches the roll", state)
	}

	// The self-control fields are trait-scoped constrained enums whose states round-trip through their enums.
	for _, field := range []selector.Field{selector.TraitSelfControlRoll, selector.TraitSelfControlAdjustment} {
		c.Equal(gurps.SelectorScopeTrait, gurps.SelectorFieldDescriptorFor(field).Scope, "%v is trait-scoped", field)
	}
	cr := gurps.SelectorFieldDescriptorFor(selector.TraitSelfControlRoll)
	c.Equal(len(selfctrl.Rolls), len(cr.SuggestedStates), "one state per self-control roll")
	for _, state := range cr.SuggestedStates {
		n, err := strconv.Atoi(state)
		c.True(err == nil, "state %q is numeric", state)
		roll := selfctrl.Roll(n)
		c.Equal(roll, roll.EnsureValid(), "state %q is a valid self-control roll", state)
		c.Equal(roll.String(), cr.StateTitle(state), "state %q label matches the roll", state)
	}
	adj := gurps.SelectorFieldDescriptorFor(selector.TraitSelfControlAdjustment)
	c.Equal(len(selfctrl.Adjustments), len(adj.SuggestedStates), "one state per adjustment")
	for _, state := range adj.SuggestedStates {
		option := selfctrl.ExtractAdjustment(state)
		c.Equal(state, option.Key(), "state %q round-trips through the adjustment key", state)
		c.Equal(option.String(), adj.StateTitle(state), "state %q label matches the adjustment", state)
	}
}
