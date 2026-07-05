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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selector"
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
