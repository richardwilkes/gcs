/*
 * Copyright ©1998-2024 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package gurps_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wpn"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/check"
)

func TestWeaponParryAndBlockStorage(t *testing.T) {
	const (
		parryKey = "parry"
		blockKey = "block"
	)

	w := gurps.Weapon{WeaponData: gurps.WeaponData{Type: wpn.Melee}}
	data, err := json.Marshal(&w)
	check.NoError(t, err)
	s := string(data)
	check.NotContains(t, s, parryKey)
	check.NotContains(t, s, blockKey)

	var loadedWeapon gurps.Weapon
	check.NoError(t, json.Unmarshal(data, &loadedWeapon))
	check.Equal(t, gurps.WeaponParry{}, loadedWeapon.Parry)
	check.Equal(t, gurps.WeaponBlock{}, loadedWeapon.Block)

	w.Parry.CanParry = true
	w.Block.CanBlock = true
	data, err = json.Marshal(&w)
	check.NoError(t, err)
	s = string(data)
	check.Contains(t, s, parryKey)
	check.Contains(t, s, blockKey)

	w.Type = wpn.Ranged
	data, err = json.Marshal(&w)
	check.NoError(t, err)
	s = string(data)
	check.NotContains(t, s, parryKey)
	check.NotContains(t, s, blockKey)

	loadedWeapon = gurps.Weapon{}
	check.NoError(t, json.Unmarshal(data, &loadedWeapon))
	check.Equal(t, gurps.WeaponParry{}, loadedWeapon.Parry)
	check.Equal(t, gurps.WeaponBlock{}, loadedWeapon.Block)
}
