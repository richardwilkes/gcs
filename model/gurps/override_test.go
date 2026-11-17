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
	"fmt"
	"hash"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selector"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// makeOverride builds a SelectorOverride with the given priority and enough pinned criteria to give it the requested
// specificity.
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

// TestResolveOverride verifies the resolution ladder: the base value stands only without overrides, the highest
// priority wins, then the most specific, and only a tie between different values, settled by which renders first, is
// flagged in the tooltip as a conflict.
func TestResolveOverride(t *testing.T) {
	c := check.New(t)
	identity := func(s string) string { return s }

	c.Equal("cr", gurps.ResolveOverride("cr", nil, identity, nil), "no override keeps base")

	c.Equal("burn", gurps.ResolveOverride("cr",
		[]gurps.OverrideCandidate[string]{candidate(makeOverride("burn", 0, 0))}, identity, nil),
		"priority-0 override still beats base")

	c.Equal("burn", gurps.ResolveOverride("cr", []gurps.OverrideCandidate[string]{
		candidate(makeOverride("imp", 0, 0)),
		candidate(makeOverride("burn", 10, 0)),
	}, identity, nil), "highest priority wins")

	c.Equal("burn", gurps.ResolveOverride("cr", []gurps.OverrideCandidate[string]{
		candidate(makeOverride("imp", 5, 1)),
		candidate(makeOverride("burn", 5, 3)),
	}, identity, nil), "more specific wins the tie")

	// Equal priority and specificity but different values is a real conflict: resolved deterministically by rendered
	// value ("burn" < "imp"), independent of input order, and flagged in the tooltip.
	var tooltip xbytes.InsertBuffer
	got := gurps.ResolveOverride("cr", []gurps.OverrideCandidate[string]{
		candidate(makeOverride("imp", 5, 2)),
		candidate(makeOverride("burn", 5, 2)),
	}, identity, &tooltip)
	c.Equal("burn", got, "conflict resolves deterministically")
	c.True(strings.Contains(tooltip.String(), "conflict"), "conflict is flagged in tooltip")

	// Equal priority and specificity with all the same value is not a conflict: nothing was decided by name order.
	tooltip.Reset()
	got = gurps.ResolveOverride("cr", []gurps.OverrideCandidate[string]{
		candidate(makeOverride("burn", 5, 2)),
		candidate(makeOverride("burn", 5, 2)),
	}, identity, &tooltip)
	c.Equal("burn", got, "agreeing candidates resolve to their shared value")
	c.False(strings.Contains(tooltip.String(), "conflict"), "agreement is not flagged as a conflict")

	// Three tied candidates valued {burn, burn, imp}: the winner's immediate runner-up agrees with it, but "imp" lost
	// purely to name order, so this is a genuine conflict. The whole tied group must be scanned, not just index 1.
	tooltip.Reset()
	got = gurps.ResolveOverride("cr", []gurps.OverrideCandidate[string]{
		candidate(makeOverride("imp", 5, 2)),
		candidate(makeOverride("burn", 5, 2)),
		candidate(makeOverride("burn", 5, 2)),
	}, identity, &tooltip)
	c.Equal("burn", got, "conflict past the runner-up still resolves deterministically")
	c.True(strings.Contains(tooltip.String(), "conflict"), "conflict beyond the runner-up is flagged in tooltip")

	// A differing value that lost on priority or specificity is not a conflict -- the ladder decided it, not name
	// order.
	tooltip.Reset()
	got = gurps.ResolveOverride("cr", []gurps.OverrideCandidate[string]{
		candidate(makeOverride("burn", 5, 2)),
		candidate(makeOverride("burn", 5, 2)),
		candidate(makeOverride("imp", 5, 1)),
	}, identity, &tooltip)
	c.Equal("burn", got, "less specific candidate loses")
	c.False(strings.Contains(tooltip.String(), "conflict"), "a ladder-decided loss is not a conflict")
}

// nilOwnerOverride is an Override that does *not* embed BonusOwner, so overrideSourceName falls through to the
// Owner()/String() path. Its owner is a typed nil, which passes a plain `!= nil` interface check yet panics if String()
// is called on it.
type nilOwnerOverride struct {
	owner    fmt.Stringer
	subOwner fmt.Stringer
}

type nilStringer struct{ name string }

// String dereferences the receiver, so calling it on a nil *nilStringer panics.
func (n *nilStringer) String() string { return n.name }

func (n *nilOwnerOverride) FillWithNameableKeys(_, _ map[string]string) {}
func (n *nilOwnerOverride) FeatureType() feature.Type                   { return feature.SelectorOverride }
func (n *nilOwnerOverride) Hash(_ hash.Hash)                            {}
func (n *nilOwnerOverride) Owner() fmt.Stringer                         { return n.owner }
func (n *nilOwnerOverride) SetOwner(owner fmt.Stringer)                 { n.owner = owner }
func (n *nilOwnerOverride) SubOwner() fmt.Stringer                      { return n.subOwner }
func (n *nilOwnerOverride) SetSubOwner(owner fmt.Stringer)              { n.subOwner = owner }
func (n *nilOwnerOverride) OverridePriority() int                       { return 0 }
func (n *nilOwnerOverride) OverrideSpecificity() int                    { return 0 }
func (n *nilOwnerOverride) IsSwitchable() bool                          { return false }
func (n *nilOwnerOverride) SetSwitchable(_ bool)                        {}
func (n *nilOwnerOverride) Describe(_ *gurps.Entity, _ map[string]string, _ func(string) string) string {
	return ""
}

func (n *nilOwnerOverride) Clone() gurps.Feature {
	return &nilOwnerOverride{owner: n.owner, subOwner: n.subOwner}
}

// TestResolveOverrideTooltipTypedNilOwner verifies that an Override whose owner is a nil pointer stored in a
// fmt.Stringer renders as "Unknown" in the tooltip rather than panicking inside String().
func TestResolveOverrideTooltipTypedNilOwner(t *testing.T) {
	c := check.New(t)
	identity := func(s string) string { return s }
	o := &nilOwnerOverride{}
	var typedNil *nilStringer
	o.SetOwner(typedNil) // Non-nil interface holding a nil pointer.
	c.True(o.Owner() != nil, "a typed nil passes a plain interface nil check")
	var tooltip xbytes.InsertBuffer
	c.NotPanics(func() {
		c.Equal("burn", gurps.ResolveOverride("cr",
			[]gurps.OverrideCandidate[string]{{Value: "burn", Override: o}}, identity, &tooltip),
			"the override still wins")
	}, "a typed-nil owner does not panic while building the tooltip")
	c.True(strings.Contains(tooltip.String(), "Unknown"), "a typed-nil owner renders as Unknown")
}
