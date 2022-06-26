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

package editors

// Matcher defines the methods that rows that can participate in searching must implement.
type Matcher interface {
	// Match should look for the text in the object and return true if it is present. Note that calls to this method
	// should always pass in text that has already been run through strings.ToLower().
	Match(text string) bool
}
