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
	"fmt"
	"hash"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// Feature holds data that affects another object.
type Feature interface {
	nameable.Filler
	FeatureType() feature.Type
	Clone() Feature
	Hash(hash.Hash)
	// IsSwitchable returns true if this feature only takes effect while the owning item's switch is on.
	IsSwitchable() bool
	// SetSwitchable sets whether this feature only takes effect while the owning item's switch is on.
	SetSwitchable(switchable bool)
	// Describe returns a plain-language description of what this Feature does, naming attributes and hit locations as
	// the entity, which may be nil, defines them. Nameable markers take their values from replacements, and those
	// without one are left as they are. Amounts, names and qualifiers are passed through em, which may wrap them for
	// emphasis; pass an identity func for plain text.
	Describe(entity *Entity, replacements map[string]string, em func(string) string) string
}

// Bonus is an extension of a Feature, which provides a numerical bonus or penalty.
type Bonus interface {
	Feature
	Owner() fmt.Stringer
	SetOwner(owner fmt.Stringer)
	SubOwner() fmt.Stringer
	SetSubOwner(owner fmt.Stringer)
	SetLeveledOwner(provider LeveledOwner)
	// AdjustedAmount returns the amount, adjusted for the owner's level when the bonus is per-level.
	AdjustedAmount() fxp.Int
	// AddToTooltip adds this Bonus's details to the tooltip. 'buffer' may be nil.
	AddToTooltip(buffer *xbytes.InsertBuffer)
}

// FeaturesForSelfControlRoll returns the set of features to apply for the given self control roll.
func FeaturesForSelfControlRoll(cr selfctrl.Roll, adj selfctrl.Adjustment) Features {
	if adj.EnsureValid() != selfctrl.MajorCostOfLivingIncrease {
		return nil
	}
	f := NewSkillBonus()
	f.NameCriteria.Qualifier = "Merchant"
	f.Amount = fxp.FromInteger(cr.Penalty())
	return Features{f}
}

// describeAmount returns the amount passed through em, followed by " per level" when it is given per level.
func describeAmount(amount string, perLevel bool, em func(string) string) string {
	if perLevel {
		return em(amount) + i18n.Text(" per level")
	}
	return em(amount)
}

// describePoints returns the amount of points passed through em, followed by " per level" when it is given per level.
func describePoints(amount fxp.Int, perLevel bool, em func(string) string) string {
	text := em(amount.StringWithSign())
	if amount == fxp.One || amount == -fxp.One {
		text += i18n.Text(" point")
	} else {
		text += i18n.Text(" points")
	}
	if perLevel {
		text += i18n.Text(" per level")
	}
	return text
}

// describeSwitchable returns the description of a feature, noting when it only takes effect while its item is
// switched on.
func describeSwitchable(switchable bool, description string) string {
	if switchable {
		return description + i18n.Text(", only while switched on")
	}
	return description
}

// describeTarget returns how a feature names what it applies to: the name in the format one, such as "skill %s", for
// "is", "all" followed by many when any name will do, and many followed by a clause starting with whose, such as
// "whose name", otherwise.
func describeTarget(one, many, whose string, t criteria.Text, replacements map[string]string, em func(string) string) string {
	switch {
	case t.Compare == criteria.AnyText:
		return fmt.Sprintf(i18n.Text("all %s"), many)
	case t.Compare == criteria.IsText && t.Qualifier != "":
		return fmt.Sprintf(one, em(nameable.Apply(t.Qualifier, replacements)))
	default:
		return many + " " + whose + " " + describeText(t, replacements, em)
	}
}

// describeWhose returns a clause such as ` whose usage is Thrown`, or nothing when anything will do.
func describeWhose(whose string, t criteria.Text, replacements map[string]string, em func(string) string) string {
	if t.Compare == criteria.AnyText {
		return ""
	}
	return " " + whose + " " + describeText(t, replacements, em)
}
