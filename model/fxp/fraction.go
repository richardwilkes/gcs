// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package fxp

import "github.com/richardwilkes/toolbox/xmath/fixed/f64"

// Fraction is an alias for the fixed-point fractional type we are using.
type Fraction = f64.Fraction[DP]

// NewFraction creates a new fractional value from a string.
func NewFraction(str string) Fraction {
	return f64.NewFraction[DP](str)
}
