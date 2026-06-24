// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package selfctrl_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestRollNumber(t *testing.T) {
	c := check.New(t)
	for _, data := range []struct {
		roll selfctrl.Roll
		want int
	}{
		{roll: selfctrl.None, want: 0},
		{roll: selfctrl.Always, want: 0},
		{roll: selfctrl.CR6, want: 6},
		{roll: selfctrl.CR12, want: 12},
		{roll: selfctrl.CR15, want: 15},
	} {
		c.Equal(data.want, data.roll.Number(), "Number() for %v", data.roll)
	}
}
