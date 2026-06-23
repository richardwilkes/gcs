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
	"encoding/json/v2"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestWeaponParryAndBlockStorage(t *testing.T) {
	c := check.New(t)
	const (
		parryKey = "parry"
		blockKey = "block"
	)

	w := gurps.NewWeapon(nil, true)
	data, err := json.Marshal(w)
	c.NoError(err)
	s := string(data)
	c.NotContains(s, parryKey)
	c.NotContains(s, blockKey)

	var loadedWeapon gurps.Weapon
	c.NoError(json.Unmarshal(data, &loadedWeapon))
	c.Equal(gurps.WeaponParry{}, loadedWeapon.Parry)
	c.Equal(gurps.WeaponBlock{}, loadedWeapon.Block)

	w.Parry.CanParry = true
	w.Block.CanBlock = true
	data, err = json.Marshal(&w)
	c.NoError(err)
	s = string(data)
	c.Contains(s, parryKey)
	c.Contains(s, blockKey)

	w = gurps.NewWeapon(nil, false)
	data, err = json.Marshal(w)
	c.NoError(err)
	s = string(data)
	c.NotContains(s, parryKey)
	c.NotContains(s, blockKey)

	loadedWeapon = gurps.Weapon{}
	c.NoError(json.Unmarshal(data, &loadedWeapon))
	c.Equal(gurps.WeaponParry{}, loadedWeapon.Parry)
	c.Equal(gurps.WeaponBlock{}, loadedWeapon.Block)
}

// TestWeaponDamageWithNilOwner verifies that formatting and marshaling a weapon whose damage has no back-reference to
// its owning weapon does not panic. This can happen transiently on the table undo/redo deserialize path, where freshly
// deserialized weapons have not yet had their owners wired up. A non-dice (script) damage base is required to force
// resolution through the scripting path, which previously dereferenced the nil owner and panicked. See issue #1015.
func TestWeaponDamageWithNilOwner(t *testing.T) {
	c := check.New(t)

	w := gurps.NewWeapon(nil, true)
	w.Damage.Base = "$self.skillLevel" // A non-dice base forces resolution through the scripting path.
	w.Damage.Owner = nil               // Explicit: no back-reference to the owning weapon.

	c.NotPanics(func() {
		_ = w.Damage.String()
	})
	c.NotPanics(func() {
		_, err := json.Marshal(w)
		c.NoError(err)
	})
}
