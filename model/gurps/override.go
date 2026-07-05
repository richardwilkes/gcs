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
	"fmt"
	"sort"

	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// Override is a Feature that *replaces* a field's value rather than adjusting it additively. Unlike a Bonus, overrides
// are never accumulated: when more than one applies to the same field, exactly one must win. The winner is chosen by a
// fixed, user-visible ladder (see ResolveOverride) so the outcome is predictable and never depends on the order rows
// happen to be stored in.
type Override interface {
	Feature
	// Owner returns the owner that is currently set.
	Owner() fmt.Stringer
	// SetOwner sets the owner to use.
	SetOwner(owner fmt.Stringer)
	// SubOwner returns the sub-owner that is currently set.
	SubOwner() fmt.Stringer
	// SetSubOwner sets the sub-owner to use.
	SetSubOwner(owner fmt.Stringer)
	// OverridePriority returns the author-assigned priority. Higher wins. This is the deliberate control knob: an
	// ordinary override leaves it at 0, a special effect raises it to out-rank ordinary ones.
	OverridePriority() int
	// OverrideSpecificity returns how narrowly this override's match criteria target a field. Higher is more specific.
	// It breaks ties between overrides of equal priority: "the more specific rule wins", the same intuition CSS uses.
	OverrideSpecificity() int
}

// OverrideCandidate pairs a resolved value with the Override that produced it, so the resolver can both pick a winner
// and report the full contest in a tooltip.
type OverrideCandidate[T comparable] struct {
	Value    T
	Override Override
}

// ResolveOverride applies the override-resolution ladder and returns the winning value. base is the field's intrinsic
// value, used when nothing applies; it also plays the part of an implicit lowest-priority contender, so an ordinary
// priority-0 override still replaces it, yet the "pick the single winner" logic stays uniform. render turns a value
// into display text (for the tooltip and for the deterministic final tie-break). If tooltip is non-nil, the entire
// contest is written to it, the winner and any genuine conflict flagged, so the user can always see why a value won.
//
// The ladder, applied in order:
//  1. Highest OverridePriority wins.
//  2. Highest OverrideSpecificity wins (more specific criteria beat broader ones).
//  3. Last resort: the value that renders first (stable string order). This is order-independent, so dragging rows
//     around never changes the result; when it is what decides two *different* values, that is a real conflict and is
//     marked as such in the tooltip rather than silently resolved.
func ResolveOverride[T comparable](base T, candidates []OverrideCandidate[T], render func(T) string, tooltip *xbytes.InsertBuffer) T {
	if len(candidates) == 0 {
		return base
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if pa, pb := a.Override.OverridePriority(), b.Override.OverridePriority(); pa != pb {
			return pa > pb
		}
		if sa, sb := a.Override.OverrideSpecificity(), b.Override.OverrideSpecificity(); sa != sb {
			return sa > sb
		}
		return render(a.Value) < render(b.Value)
	})
	winner := candidates[0]
	conflict := len(candidates) > 1 &&
		candidates[1].Override.OverridePriority() == winner.Override.OverridePriority() &&
		candidates[1].Override.OverrideSpecificity() == winner.Override.OverrideSpecificity() &&
		candidates[1].Value != winner.Value
	if tooltip != nil {
		addOverrideContestToTooltip(winner, candidates, render, conflict, tooltip)
	}
	return winner.Value
}

func addOverrideContestToTooltip[T comparable](winner OverrideCandidate[T], candidates []OverrideCandidate[T], render func(T) string, conflict bool, tooltip *xbytes.InsertBuffer) {
	tooltip.WriteByte('\n')
	fmt.Fprintf(tooltip, i18n.Text("Set to **%q** by **%s**"), render(winner.Value),
		overrideSourceName(winner.Override))
	if conflict {
		tooltip.WriteString(i18n.Text(" (conflict — resolved by name)"))
	}
	for _, c := range candidates[1:] {
		tooltip.WriteByte('\n')
		fmt.Fprintf(tooltip, i18n.Text("  overriding **%q** from **%s** (priority **%d**)"), render(c.Value),
			overrideSourceName(c.Override), c.Override.OverridePriority())
	}
}

func overrideSourceName(o Override) string {
	if bo, ok := o.(interface{ parentName() string }); ok {
		return bo.parentName()
	}
	if owner := o.Owner(); owner != nil {
		return owner.String()
	}
	return i18n.Text("Unknown")
}
