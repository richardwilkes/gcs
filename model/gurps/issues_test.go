// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps_test

import (
	"os"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestIssue767KickDamageUsesBaseSkill(t *testing.T) {
	c := check.New(t)
	e, err := gurps.NewEntityFromFile(os.DirFS("testdata"), "issue767.gcs")
	c.NoError(err)
	var kick *gurps.Weapon
	gurps.Traverse(func(tr *gurps.Trait) bool {
		for _, w := range tr.Weapons {
			if w.UsageWithReplacements() == "Kick" {
				kick = w
			}
		}
		return false
	}, false, false, e.Traits...)
	c.NotNil(kick)
	// Brawling is at DX+2 (RSL +2), so its "+1 per die, RSL >= 2" weapon damage bonus must apply even though the
	// best skill default for the Kick is the Kicking technique (RSL -1 vs Brawling). 1d-2 base + 1 = 1d-1.
	c.Equal("1d-1 cr", kick.Damage.ResolvedDamage(nil))
}
