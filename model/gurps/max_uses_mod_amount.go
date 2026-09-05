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
	"hash"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/maxusesmod"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xreflect"
)

// MaxUsesModAmount holds the amount portion of a maximum uses or maximum level adjustment. The Amount field encodes
// both the operation and the value: a plain number (e.g. "+1" or "-1") is an addition, a value with a trailing "%"
// (e.g. "10%") is a percentage adjustment, and a value with a leading or trailing "x" (e.g. "x2") is a multiplier. The
// amount may optionally be scaled by the level of the owning item.
type MaxUsesModAmount struct {
	LeveledOwner LeveledOwner `json:"-"`
	Amount       string       `json:"amount"`
	PerLevel     bool         `json:"per_level,omitzero"`
}

// Operation returns the type of adjustment encoded in the Amount.
func (m *MaxUsesModAmount) Operation() maxusesmod.Type {
	return maxusesmod.FromString(m.Amount)
}

// SetLeveledOwner implements Bonus.
func (m *MaxUsesModAmount) SetLeveledOwner(owner LeveledOwner) {
	m.LeveledOwner = owner
}

// AdjustedAmount implements Bonus. It returns the numeric portion of the Amount, scaled by the leveled owner's current
// level when this is a per-level bonus. A per-level bonus whose owner has no levels contributes nothing.
func (m *MaxUsesModAmount) AdjustedAmount() fxp.Int {
	value := m.Operation().ExtractValue(m.Amount)
	if m.PerLevel {
		if xreflect.IsNil(m.LeveledOwner) || !m.LeveledOwner.IsLeveled() {
			return 0
		}
		level := m.LeveledOwner.CurrentLevel()
		if level <= 0 {
			return 0
		}
		value = value.Mul(level)
	}
	return value
}

// Format returns the adjusted amount formatted according to the operation encoded in the Amount.
func (m *MaxUsesModAmount) Format() string {
	return m.Operation().Format(m.AdjustedAmount())
}

// addToTooltip appends a line to the tooltip describing this adjustment and the bonus it came from.
func (m *MaxUsesModAmount) addToTooltip(parentName string, buffer *xbytes.InsertBuffer) {
	if buffer != nil {
		buffer.WriteByte('\n')
		buffer.WriteString(parentName)
		buffer.WriteString(" [")
		buffer.WriteString(m.Format())
		buffer.WriteByte(']')
	}
}

// Hash writes this object's contents into the hasher.
func (m *MaxUsesModAmount) Hash(h hash.Hash) {
	if m == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.StringWithLen(h, m.Amount)
	xhash.Bool(h, m.PerLevel)
}
