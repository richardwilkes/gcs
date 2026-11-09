// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package difficulty_test

import (
	"fmt"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestLevelUnmarshalTextAcceptsOnlyKnownKeys verifies that a key loads its level whatever its case, and that a key
// matching no level is an error naming the key rather than the default level.
func TestLevelUnmarshalTextAcceptsOnlyKnownKeys(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		text string
		want difficulty.Level
	}{
		{"h", difficulty.Hard},
		{"VH", difficulty.VeryHard},
	} {
		var level difficulty.Level
		c.NoError(level.UnmarshalText([]byte(one.text)), one.text)
		c.Equal(one.want, level, one.text)
	}
	var level difficulty.Level
	err := level.UnmarshalText([]byte("x"))
	c.HasError(err, "an unknown key is an error")
	c.Contains(fmt.Sprint(err), `"x"`, "the error names the key")
}
