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

// LeveledAmount holds an amount that can be either a fixed amount, or an amount per level.
type LeveledAmount struct {
	Level    fxp.Int `json:"-"`
	Amount   fxp.Int `json:"amount"`
	PerLevel bool    `json:"per_level,omitempty"`
}

// AdjustedAmount returns the amount, adjusted for level, if requested.
func (l *LeveledAmount) AdjustedAmount() fxp.Int {
	if l.PerLevel {
		if l.Level < 0 {
			return 0
		}
		return l.Amount.Mul(l.Level)
	}
	return l.Amount
}

// Format the value.
func (l *LeveledAmount) Format() string {
	amt := l.Amount.StringWithSign()
	if !l.PerLevel {
		return amt
	}
	return fmt.Sprintf(i18n.Text("%s (%s per level)"), l.AdjustedAmount().StringWithSign(), amt)
}

// Hash writes this object's contents into the hasher.
func (l *LeveledAmount) Hash(h hash.Hash) {
	if l == nil {
		hashhelper.Num8(h, uint8(255))
		return
	}
	hashhelper.Num64(h, l.Amount)
	hashhelper.Bool(h, l.PerLevel)
}
