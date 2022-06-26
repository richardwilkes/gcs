/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package spell

// UsesStringCriteria returns true if the comparison uses a string value.
func (enum ComparisonType) UsesStringCriteria() bool {
	v := enum.EnsureValid()
	return v == Name || v == Tag || v == College
}
