// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package condition

// State is the outcome of evaluating a node. It is never persisted. The zero value is Failed, so a zero Result never
// reads as Met or passes.
type State byte

// Possible State values.
const (
	// Failed means the node could not be evaluated.
	Failed State = iota
	// Met means the node holds.
	Met
	// Unmet means the node does not hold.
	Unmet
	// Skipped means the node does not apply, such as a group whose only_when guard is unmet.
	Skipped
)

// String implements fmt.Stringer.
func (s State) String() string {
	switch s {
	case Failed:
		return "failed"
	case Met:
		return "met"
	case Unmet:
		return "unmet"
	case Skipped:
		return "skipped"
	default:
		return "?"
	}
}

// Passes returns true for Met and Skipped, which is what "satisfied" means for prereqs and filters.
func (s State) Passes() bool {
	return s == Met || s == Skipped
}

// Mode selects how much an evaluation reports. It is never persisted.
type Mode byte

// Possible Mode values.
const (
	// YesNo decides the state only.
	YesNo Mode = iota
	// Collect also fills in the children, hits and specificity of each result.
	Collect
	// Explain decides the state as YesNo does, then collects as Collect does when the state doesn't pass, to say why.
	Explain
)
