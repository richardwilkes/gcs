// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/json"
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
