// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestRefuseFractions checks that a field refusing fractions still takes text that spells a whole number with a
// decimal point, such as pasted "3.0", and refuses only text with a fractional part and a typed decimal point.
func TestRefuseFractions(t *testing.T) {
	c := check.New(t)
	value := fxp.Three
	field := NewDecimalField(nil, "", "", func() fxp.Int { return value }, func(v fxp.Int) { value = v }, 0,
		fxp.Ten, false, false)
	refuseFractions(field)

	v, err := field.extract("3.0")
	c.NoError(err, "a whole number written with a decimal point is valid")
	c.Equal(fxp.Three, v)
	c.Equal("", field.tooltipTextForValidation(), "the field's own text is valid")
	field.SetText("3.0")
	c.Equal("", field.tooltipTextForValidation(), "pasted 3.0 is valid")
	c.Equal(fxp.Three, field.mustExtract(field.Text()), "and hands the setter 3, not the minimum")

	_, err = field.extract("3.5")
	c.HasError(err, "a fractional part is invalid")
	field.SetText("3.5")
	c.NotEqual("", field.tooltipTextForValidation(), "pasted 3.5 is invalid")

	field.SetText("3")
	field.SetSelectionToEnd()
	c.False(field.RuneTypedCallback('.'), "a typed decimal point is refused")
	c.Equal("3", field.Text())
	c.True(field.RuneTypedCallback('1'), "a typed digit is taken")
	c.Equal("31", field.Text())
}
