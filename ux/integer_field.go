// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"fmt"
	"math"
	"strconv"
)

// IntegerField is field that holds an integer.
type IntegerField = NumericField[int]

// NewIntegerField creates a new field that holds an int.
func NewIntegerField(targetMgr *TargetMgr, targetKey, undoTitle string, get func() int, set func(int), minValue, maxValue int, forceSign, noMinWidth bool) *IntegerField {
	var getPrototype func(minValue, maxValue int) []int
	if !noMinWidth {
		getPrototype = func(minValue, maxValue int) []int {
			if minValue == math.MinInt {
				minValue = -99
			}
			if maxValue == math.MaxInt {
				maxValue = 99
			}
			return []int{minValue, 99, maxValue}
		}
	}
	format := func(value int) string {
		if forceSign {
			return fmt.Sprintf("%+d", value)
		}
		return strconv.Itoa(value)
	}
	return NewNumericField(targetMgr, targetKey, undoTitle, getPrototype, get, set, format, strconv.Atoi,
		minValue, maxValue)
}
