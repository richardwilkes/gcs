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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/xio"
)

// Feature holds data that affects another object.
type Feature interface {
	nameable.Filler
	FeatureType() feature.Type
	Clone() Feature
	Hash(hash.Hash)
}

// Bonus is an extension of a Feature, which provides a numerical bonus or penalty.
type Bonus interface {
	Feature
	// Owner returns the owner that is currently set.
	Owner() fmt.Stringer
	// SetOwner sets the owner to use.
	SetOwner(owner fmt.Stringer)
	// SubOwner returns the sub-owner that is currently set.
	SubOwner() fmt.Stringer
	// SetSubOwner sets the sub-owner to use.
	SetSubOwner(owner fmt.Stringer)
	// SetLevel sets the level.
	SetLevel(level fxp.Int)
	// AdjustedAmount returns the amount, adjusted for level, if requested.
	AdjustedAmount() fxp.Int
	// AddToTooltip adds this Bonus's details to the tooltip. 'buffer' may be nil.
	AddToTooltip(buffer *xio.ByteBuffer)
}

// FeaturesForSelfControlRoll returns the set of features to apply for the given self control roll.
func FeaturesForSelfControlRoll(cr selfctrl.Roll, adj selfctrl.Adjustment) Features {
	if adj.EnsureValid() != selfctrl.MajorCostOfLivingIncrease {
		return nil
	}
	f := NewSkillBonus()
	f.NameCriteria.Qualifier = "Merchant"
	f.Amount = fxp.From(cr.Penalty())
	return Features{f}
}
