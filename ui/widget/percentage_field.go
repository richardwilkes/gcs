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
	"strings"
)

// PercentageField is field that holds a percentage.
type PercentageField = NumericField[int]

// NewPercentageField creates a new field that holds a percentage.
func NewPercentageField(undoTitle string, get func() int, set func(int), min, max int, forceSign, noMinWidth bool) *PercentageField {
	var getPrototype func(min, max int) []int
	if !noMinWidth {
		getPrototype = func(min, max int) []int {
			if min == math.MinInt {
				min = -100
			}
			if max == math.MaxInt {
				max = 100
			}
			return []int{min, 100, max}
		}
	}
	format := func(value int) string {
		if forceSign {
			return fmt.Sprintf("%+d%%", value)
		}
		return strconv.Itoa(value) + "%"
	}
	extract := func(s string) (int, error) {
		return strconv.Atoi(strings.TrimSpace(strings.TrimSuffix(s, "%")))
	}
	return NewNumericField[int](undoTitle, getPrototype, get, set, format, extract, min, max)
}
