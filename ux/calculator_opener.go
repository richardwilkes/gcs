// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.
package ux

// OpenCalculators brings the calculators forward, opening them if they are not already open, with preselect, which may
// be nil, as the sheet they start out taking their numbers from. The ux/calculators package sets it when it is linked
// in, which main does with a blank import; until then the Calculators menu item and the sheets' calculator buttons do
// nothing.
var OpenCalculators func(preselect *Sheet)

// DisplayCalculator runs OpenCalculators, if it has been set.
func DisplayCalculator(preselect *Sheet) {
	if OpenCalculators != nil {
		OpenCalculators(preselect)
	}
}
