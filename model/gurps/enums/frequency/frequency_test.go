// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package frequency_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/frequency"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestRollNumber(t *testing.T) {
	c := check.New(t)
	for _, data := range []struct {
		roll frequency.Roll
		want int
	}{
		{roll: frequency.None, want: 0},
		{roll: frequency.FR6, want: 6},
		{roll: frequency.FR9, want: 9},
		{roll: frequency.FR15, want: 15},
		{roll: frequency.Constant, want: 0},
	} {
		c.Equal(data.want, data.roll.Number(), "Number() for %v", data.roll)
	}
}
