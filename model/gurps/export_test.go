// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

// SetWeaponMemosEnabled turns the memoizing of weapon resolution, and the keeping of the entity's weapon lists, on or
// off, and returns a function that restores the previous setting.
func SetWeaponMemosEnabled(enabled bool) (restore func()) {
	previous := weaponMemosEnabled
	weaponMemosEnabled = enabled
	return func() { weaponMemosEnabled = previous }
}
