// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package paper_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/paper"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestRealLengthConversion(t *testing.T) {
	c := check.New(t)
	c.Equal(`0.25 in`, paper.Length{Length: 0.25, Units: paper.Inch}.String())
	c.Equal(float32(18), paper.Length{Length: 0.25, Units: paper.Inch}.Pixels())
	c.Equal(`1 in`, paper.Length{Length: 1, Units: paper.Inch}.String())
	c.Equal(float32(72), paper.Length{Length: 1, Units: paper.Inch}.Pixels())
	c.Equal(`15 in`, paper.Length{Length: 15, Units: paper.Inch}.String())
	c.Equal(float32(1080), paper.Length{Length: 15, Units: paper.Inch}.Pixels())
	c.Equal("1 cm", paper.Length{Length: 1, Units: paper.Centimeter}.String())
	c.Equal(float32(28.3464566929), paper.Length{Length: 1, Units: paper.Centimeter}.Pixels())
	c.Equal("1 mm", paper.Length{Length: 1, Units: paper.Millimeter}.String())
	c.Equal(float32(2.8346456693), paper.Length{Length: 1, Units: paper.Millimeter}.Pixels())
}
