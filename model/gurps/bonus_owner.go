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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xreflect"
)

// BonusOwner provides a convenience for implementing the owner & sub-owner methods of Bonus.
type BonusOwner struct {
	owner    fmt.Stringer
	subOwner fmt.Stringer
}

// LeveledOwner is an owner that has a level.
type LeveledOwner interface {
	IsLeveled() bool
	CurrentLevel() fxp.Int
}

// Owner returns the owner that is currently set.
func (b *BonusOwner) Owner() fmt.Stringer {
	return b.owner
}

// SetOwner sets the owner to use.
func (b *BonusOwner) SetOwner(owner fmt.Stringer) {
	b.owner = owner
}

// SubOwner returns the sub-owner that is currently set.
func (b *BonusOwner) SubOwner() fmt.Stringer {
	return b.subOwner
}

// SetSubOwner sets the sub-owner to use.
func (b *BonusOwner) SetSubOwner(subOwner fmt.Stringer) {
	b.subOwner = subOwner
}

func (b *BonusOwner) basicAddToTooltip(amt *LeveledAmount, buffer *xbytes.InsertBuffer) {
	if buffer != nil {
		buffer.WriteByte('\n')
		buffer.WriteString(b.parentName())
		buffer.WriteString(" [")
		buffer.WriteString(amt.Format())
		buffer.WriteByte(']')
	}
}

func (b *BonusOwner) parentName() string {
	if xreflect.IsNil(b.owner) {
		return i18n.Text("Unknown")
	}
	owner := b.owner.String()
	if xreflect.IsNil(b.subOwner) {
		return owner
	}
	return fmt.Sprintf("%s (%v)", owner, b.subOwner)
}

// DerivedLeveledOwner returns the sub-owner or owner, if they are LeveledOwners.
func (b *BonusOwner) DerivedLeveledOwner() LeveledOwner {
	if !xreflect.IsNil(b.subOwner) {
		if lo, ok := b.subOwner.(LeveledOwner); ok && lo.IsLeveled() {
			return lo
		}
	}
	if !xreflect.IsNil(b.owner) {
		if lo, ok := b.owner.(LeveledOwner); ok && lo.IsLeveled() {
			return lo
		}
	}
	return zeroLeveledOwner{}
}

type zeroLeveledOwner struct{}

func (z zeroLeveledOwner) IsLeveled() bool       { return false }
func (z zeroLeveledOwner) CurrentLevel() fxp.Int { return 0 }
