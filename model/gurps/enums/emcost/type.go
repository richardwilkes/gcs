// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package emcost

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/fxp"
)

// StringWithExample returns an example along with the normal String() content.
func (enum Type) StringWithExample() string {
	return fmt.Sprintf("%s (e.g. %s)", enum.String(), enum.AltString())
}

// Permitted returns the permitted values.
func (enum Type) Permitted() []Value {
	if enum.EnsureValid() == Base {
		return []Value{CostFactor, Multiplier}
	}
	return []Value{Addition, Percentage, Multiplier}
}

// FromString examines a string to determine what Value it is, but restricts the result to those allowed for this
// Type.
func (enum Type) FromString(s string) Value {
	cvt := Addition.FromString(s)
	permitted := enum.Permitted()
	for _, one := range permitted {
		if one == cvt {
			return cvt
		}
	}
	return permitted[0]
}

// ExtractValue from the string.
func (enum Type) ExtractValue(s string) fxp.Int {
	return enum.FromString(s).ExtractValue(s)
}

// Format returns a formatted version of the value.
func (enum Type) Format(s string) string {
	cvt := enum.FromString(s)
	return cvt.Format(cvt.ExtractValue(s))
}
