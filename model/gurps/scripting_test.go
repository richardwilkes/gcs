// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestScriptMathExp2 verifies that Math.exp2 is exposed to scripts as a real member of the built-in Math object (rather
// than an unreachable global with a dotted name), while leaving the standard Math members intact.
func TestScriptMathExp2(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		script string
		want   string
	}{
		{script: "typeof Math.exp2", want: "function"},
		{script: "Math.exp2(0)", want: "1"},
		{script: "Math.exp2(3)", want: "8"},
		{script: "Math.exp2(10)", want: "1024"},
		{script: "typeof Math.pow", want: "function"}, // ensure the built-in Math members are still present
	} {
		v, err := runScript(0, tc.script)
		c.NoError(err, "script %q", tc.script)
		c.Equal(tc.want, v.String(), "script %q", tc.script)
	}
}
