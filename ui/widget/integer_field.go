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

package widget

import (
	"fmt"
	"math"
	"strconv"
)

// IntegerField is field that holds an integer.
type IntegerField = NumericField[int]

// NewIntegerField creates a new field that holds an int.
func NewIntegerField(undoTitle string, get func() int, set func(int), min, max int, forceSign, noMinWidth bool) *IntegerField {
	var getPrototype func(min, max int) []int
	if !noMinWidth {
		getPrototype = func(min, max int) []int {
			if min == math.MinInt {
				min = -99
			}
			if max == math.MaxInt {
				max = 99
			}
			return []int{min, 99, max}
		}
	}
	format := func(value int) string {
		if forceSign {
			return fmt.Sprintf("%+d", value)
		}
		return strconv.Itoa(value)
	}
	return NewNumericField[int](undoTitle, getPrototype, get, set, format, strconv.Atoi, min, max)
}
