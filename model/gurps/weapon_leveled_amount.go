/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package gurps

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/i18n"
)

// WeaponLeveledAmount holds an amount that can be either a fixed amount, or an amount per level and/or per die.
type WeaponLeveledAmount struct {
	Level    fxp.Int
	DieCount fxp.Int
	Amount   fxp.Int
	PerLevel bool
	PerDie   bool
}

// MarshalJSON marshals to JSON.
func (l *WeaponLeveledAmount) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Amount   fxp.Int `json:"amount"`
		PerLevel bool    `json:"leveled,omitempty"`
		PerDie   bool    `json:"per_die,omitempty"`
	}{
		Amount:   l.Amount,
		PerLevel: l.PerLevel,
		PerDie:   l.PerDie,
	})
}

// UnmarshalJSON unmarshals from JSON.
func (l *WeaponLeveledAmount) UnmarshalJSON(data []byte) error {
	var input struct {
		Amount    fxp.Int `json:"amount"`
		PerLevel  bool    `json:"leveled"`
		PerDie    bool    `json:"per_die"`
		OldPerDie bool    `json:"per_level"`
	}
	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}
	l.Amount = input.Amount
	l.PerLevel = input.PerLevel
	l.PerDie = input.PerDie || input.OldPerDie
	return nil
}

// AdjustedAmount returns the amount, adjusted for level, if requested.
func (l *WeaponLeveledAmount) AdjustedAmount() fxp.Int {
	amt := l.Amount
	if l.PerDie {
		if l.DieCount < 0 {
			return 0
		}
		amt = amt.Mul(l.DieCount)
	}
	if l.PerLevel {
		if l.Level < 0 {
			return 0
		}
		amt = amt.Mul(l.Level)
	}
	return amt
}

// Format the value.
func (l *WeaponLeveledAmount) Format(asPercentage bool) string {
	amt := l.Amount.StringWithSign()
	adjustedAmt := l.AdjustedAmount().StringWithSign()
	if asPercentage {
		amt += "%"
		adjustedAmt += "%"
	}
	switch {
	case l.PerDie && l.PerLevel:
		return fmt.Sprintf(i18n.Text("%s (%s per die, per level)"), adjustedAmt, amt)
	case l.PerDie:
		return fmt.Sprintf(i18n.Text("%s (%s per die)"), adjustedAmt, amt)
	case l.PerLevel:
		return fmt.Sprintf(i18n.Text("%s (%s per level)"), adjustedAmt, amt)
	default:
		return amt
	}
}
