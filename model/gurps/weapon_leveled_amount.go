// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"fmt"
	"hash"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
)

// WeaponLeveledAmount holds an amount that can be either a fixed amount, or an amount per level and/or per die.
type WeaponLeveledAmount struct {
	Level    fxp.Int `json:"-"`
	DieCount fxp.Int `json:"-"`
	Amount   fxp.Int `json:"amount"`
	PerLevel bool    `json:"leveled,omitempty"`
	PerDie   bool    `json:"per_die,alt=per_level,omitempty"`
}

// AdjustedAmount returns the amount, adjusted for level, if requested.
func (w *WeaponLeveledAmount) AdjustedAmount() fxp.Int {
	amt := w.Amount
	if w.PerDie {
		if w.DieCount < 0 {
			return 0
		}
		amt = amt.Mul(w.DieCount)
	}
	if w.PerLevel {
		if w.Level < 0 {
			return 0
		}
		amt = amt.Mul(w.Level)
	}
	return amt
}

// Format the value.
func (w *WeaponLeveledAmount) Format(asPercentage bool) string {
	amt := w.Amount.StringWithSign()
	adjustedAmt := w.AdjustedAmount().StringWithSign()
	if asPercentage {
		amt += "%"
		adjustedAmt += "%"
	}
	switch {
	case w.PerDie && w.PerLevel:
		return fmt.Sprintf(i18n.Text("%s (%s per die, per level)"), adjustedAmt, amt)
	case w.PerDie:
		return fmt.Sprintf(i18n.Text("%s (%s per die)"), adjustedAmt, amt)
	case w.PerLevel:
		return fmt.Sprintf(i18n.Text("%s (%s per level)"), adjustedAmt, amt)
	default:
		return amt
	}
}

// Hash writes this object's contents into the hasher. Note that this only hashes the data that is considered to be
// "source" data, i.e. not expected to be modified by the user after copying from a library.
func (w *WeaponLeveledAmount) Hash(h hash.Hash) {
	if w == nil {
		hashhelper.Num8(h, uint8(255))
		return
	}
	hashhelper.Num64(h, w.Amount)
	hashhelper.Bool(h, w.PerLevel)
	hashhelper.Bool(h, w.PerDie)
}
