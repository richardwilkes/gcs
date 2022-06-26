/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package feature

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/nameables"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

// Feature holds data that affects another object.
type Feature interface {
	nameables.Nameables
	FeatureType() Type
	// FeatureMapKey returns the key used for matching within the feature map.
	FeatureMapKey() string
	Clone() Feature
}

// Bonus is an extension of a Feature, which provides a numerical bonus or penalty.
type Bonus interface {
	Feature
	// SetParent sets the parent to use.
	SetParent(parent fmt.Stringer)
	// SetLevel sets the level.
	SetLevel(level fxp.Int)
	// AdjustedAmount returns the amount, adjusted for level, if requested.
	AdjustedAmount() fxp.Int
	// AddToTooltip adds this Bonus's details to the tooltip. 'buffer' may be nil.
	AddToTooltip(buffer *xio.ByteBuffer)
}

func basicAddToTooltip(parent fmt.Stringer, amt *LeveledAmount, buffer *xio.ByteBuffer) {
	if buffer != nil {
		buffer.WriteByte('\n')
		buffer.WriteString(parentName(parent))
		buffer.WriteString(" [")
		buffer.WriteString(amt.FormatWithLevel())
		buffer.WriteByte(']')
	}
}

func parentName(parent fmt.Stringer) string {
	if parent == nil {
		return i18n.Text("Unknown")
	}
	return parent.String()
}
