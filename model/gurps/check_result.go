// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

// CheckResult is the outcome of checking something, such as a prerequisite against a sheet, or a saved filter against
// an item.
type CheckResult uint8

// Possible CheckResult values. A skipped check is left out, as one that doesn't apply, and a failed one couldn't be
// made, such as when a script that decides it couldn't run.
const (
	CheckMet CheckResult = iota
	CheckUnmet
	CheckSkipped
	CheckFailed
)

// negate swaps met and unmet when not is set, leaving skipped and failed alone.
func (r CheckResult) negate(not bool) CheckResult {
	if not {
		switch r {
		case CheckMet:
			return CheckUnmet
		case CheckUnmet:
			return CheckMet
		default:
		}
	}
	return r
}

// checkTally records the results of a group of checks, such as a list of prerequisites, so they can be combined.
type checkTally struct {
	met    bool
	unmet  bool
	failed bool
}

// add records a result. A skipped one is left out.
func (t *checkTally) add(result CheckResult) {
	switch result {
	case CheckMet:
		t.met = true
	case CheckUnmet:
		t.unmet = true
	case CheckFailed:
		t.failed = true
	default:
	}
}

// combine returns the result of the group. One result that decides it does so even when others failed: an unmet one
// when all is set, and a met one otherwise. Without one, the group has failed when any check has, is skipped when
// every check was left out, and is otherwise met when all is set and unmet when it isn't.
func (t checkTally) combine(all bool) CheckResult {
	switch {
	case all && t.unmet:
		return CheckUnmet
	case !all && t.met:
		return CheckMet
	case t.failed:
		return CheckFailed
	case t.met:
		return CheckMet
	case t.unmet:
		return CheckUnmet
	default:
		return CheckSkipped
	}
}
