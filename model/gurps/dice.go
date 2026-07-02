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
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/v2/xrand"
)

// Pre-configured dice rollers.
var (
	Roller      *dice.Roller
	RollerExtra *dice.Roller
)

// InitRollers initializes the global Roller and RollerExtra variables. It must be called before either of those
// variables is used.
func InitRollers() error {
	cfg := &dice.Config{
		Randomizer:             xrand.New(),
		MaxCount:               999_999,
		MaxSides:               999_999,
		MaxModifier:            999_999,
		MaxMultiplier:          999_999,
		GURPSFormat:            true,
		ExtraDiceFromModifiers: false,
	}
	dice.SetDefaultConfig(cfg)
	var err error
	if Roller, err = dice.NewRoller(cfg); err != nil {
		return err
	}
	cfg.ExtraDiceFromModifiers = true
	if RollerExtra, err = dice.NewRoller(cfg); err != nil {
		return err
	}
	return nil
}

// Roll the dice using the RollerExtra if withExtra is true, otherwise use the Roller.
func Roll(d dice.Dice, withExtra bool) int {
	if withExtra {
		return RollerExtra.Roll(d)
	}
	return Roller.Roll(d)
}

// FormatDice formats the provided dice using the RollerExtra if withExtra is true, otherwise use the Roller.
func FormatDice(d dice.Dice, withExtra bool) string {
	if withExtra {
		return RollerExtra.Format(d)
	}
	return Roller.Format(d)
}
