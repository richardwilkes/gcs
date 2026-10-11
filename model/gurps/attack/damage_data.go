// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package attack

import (
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stdmg"
)

// DamageData holds the weapon damage data that is written to disk.
type DamageData struct {
	Type                      string       `json:"type"`
	StrengthType              stdmg.Option `json:"st,omitzero"`
	Leveled                   bool         `json:"leveled,omitzero"`
	StrengthMultiplier        fxp.Int      `json:"st_mul,omitzero"`
	Base                      string       `json:"base,omitzero"`
	BaseLeveled               string       `json:"base_leveled,omitzero"`
	ArmorDivisor              fxp.Int      `json:"armor_divisor,omitzero"`
	Fragmentation             string       `json:"fragmentation,omitzero"`
	FragmentationArmorDivisor fxp.Int      `json:"fragmentation_armor_divisor,omitzero"`
	FragmentationType         string       `json:"fragmentation_type,omitzero"`
	ModifierPerDie            fxp.Int      `json:"modifier_per_die,omitzero"`
}

// Normalize puts the data into the one form used for it in memory, which is the form it has once loaded from a file,
// so that a weapon hashes the same and shows the same damage whether or not it has been saved and loaded since it was
// created or edited. A ST multiplier or armor divisor of 0 is not valid and becomes 1. Without any fragmentation, its
// armor divisor and type have nothing to apply to and aren't written to disk, so they become 1 and empty. Fragmentation
// dice supplied by a selector.WeaponFragmentationDice override don't count as the weapon's own: resolving the damage
// then starts from an armor divisor of 1 and no type, so such an override has to bring its own through
// selector.WeaponFragmentationArmorDivisor and selector.WeaponFragmentationType.
func (w *DamageData) Normalize() {
	if w.StrengthMultiplier == 0 {
		w.StrengthMultiplier = fxp.One
	}
	if w.ArmorDivisor == 0 {
		w.ArmorDivisor = fxp.One
	}
	w.Fragmentation = strings.TrimSpace(w.Fragmentation)
	if w.Fragmentation == "" {
		w.FragmentationType = ""
	}
	if w.Fragmentation == "" || w.FragmentationArmorDivisor == 0 {
		w.FragmentationArmorDivisor = fxp.One
	}
}
