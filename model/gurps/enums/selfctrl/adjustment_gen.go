// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package selfctrl

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	NoCRAdj Adjustment = iota
	ActionPenalty
	ReactionPenalty
	FrightCheckPenalty
	FrightCheckBonus
	MinorCostOfLivingIncrease
	MajorCostOfLivingIncrease
)

// LastAdjustment is the last valid value.
const LastAdjustment Adjustment = MajorCostOfLivingIncrease

// Adjustments holds all possible values.
var Adjustments = []Adjustment{
	NoCRAdj,
	ActionPenalty,
	ReactionPenalty,
	FrightCheckPenalty,
	FrightCheckBonus,
	MinorCostOfLivingIncrease,
	MajorCostOfLivingIncrease,
}

// Adjustment holds an Adjustment for a self-control roll.
type Adjustment byte

// EnsureValid ensures this is of a known value.
func (enum Adjustment) EnsureValid() Adjustment {
	if enum <= MajorCostOfLivingIncrease {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Adjustment) Key() string {
	switch enum {
	case NoCRAdj:
		return "none"
	case ActionPenalty:
		return "action_penalty"
	case ReactionPenalty:
		return "reaction_penalty"
	case FrightCheckPenalty:
		return "fright_check_penalty"
	case FrightCheckBonus:
		return "fright_check_bonus"
	case MinorCostOfLivingIncrease:
		return "minor_cost_of_living_increase"
	case MajorCostOfLivingIncrease:
		return "major_cost_of_living_increase"
	default:
		return Adjustment(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum Adjustment) String() string {
	switch enum {
	case NoCRAdj:
		return i18n.Text("None")
	case ActionPenalty:
		return i18n.Text("Includes an Action Penalty for Failure")
	case ReactionPenalty:
		return i18n.Text("Includes a Reaction Penalty for Failure")
	case FrightCheckPenalty:
		return i18n.Text("Includes Fright Check Penalty")
	case FrightCheckBonus:
		return i18n.Text("Includes Fright Check Bonus")
	case MinorCostOfLivingIncrease:
		return i18n.Text("Includes a Minor Cost of Living Increase")
	case MajorCostOfLivingIncrease:
		return i18n.Text("Includes a Major Cost of Living Increase and Merchant Skill Penalty")
	default:
		return Adjustment(0).String()
	}
}

// AltString returns the alternate string.
func (enum Adjustment) AltString() string {
	switch enum {
	case NoCRAdj:
		return i18n.Text("None")
	case ActionPenalty:
		return i18n.Text("%d Action Penalty")
	case ReactionPenalty:
		return i18n.Text("%d Reaction Penalty")
	case FrightCheckPenalty:
		return i18n.Text("%d Fright Check Penalty")
	case FrightCheckBonus:
		return i18n.Text("+%d Fright Check Bonus")
	case MinorCostOfLivingIncrease:
		return i18n.Text("+%d%% Cost of Living Increase")
	case MajorCostOfLivingIncrease:
		return i18n.Text("+%d%% Cost of Living Increase")
	default:
		return Adjustment(0).AltString()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Adjustment) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Adjustment) UnmarshalText(text []byte) error {
	*enum = ExtractAdjustment(string(text))
	return nil
}

// ExtractAdjustment extracts the value from a string.
func ExtractAdjustment(str string) Adjustment {
	for _, enum := range Adjustments {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
