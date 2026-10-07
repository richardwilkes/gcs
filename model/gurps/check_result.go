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
