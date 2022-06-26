// Code generated from "enum.go.tmpl" - DO NOT EDIT.

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

package gurps

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	NoCRAdj SelfControlRollAdj = iota
	ActionPenalty
	ReactionPenalty
	FrightCheckPenalty
	FrightCheckBonus
	MinorCostOfLivingIncrease
	MajorCostOfLivingIncrease
	LastSelfControlRollAdj = MajorCostOfLivingIncrease
)

var (
	// AllSelfControlRollAdj holds all possible values.
	AllSelfControlRollAdj = []SelfControlRollAdj{
		NoCRAdj,
		ActionPenalty,
		ReactionPenalty,
		FrightCheckPenalty,
		FrightCheckBonus,
		MinorCostOfLivingIncrease,
		MajorCostOfLivingIncrease,
	}
	selfControlRollAdjData = []struct {
		key    string
		string string
		alt    string
	}{
		{
			key:    "none",
			string: i18n.Text("None"),
			alt:    i18n.Text("None"),
		},
		{
			key:    "action_penalty",
			string: i18n.Text("Includes an Action Penalty for Failure"),
			alt:    i18n.Text("%d Action Penalty"),
		},
		{
			key:    "reaction_penalty",
			string: i18n.Text("Includes a Reaction Penalty for Failure"),
			alt:    i18n.Text("%d Reaction Penalty"),
		},
		{
			key:    "fright_check_penalty",
			string: i18n.Text("Includes Fright Check Penalty"),
			alt:    i18n.Text("%d Fright Check Penalty"),
		},
		{
			key:    "fright_check_bonus",
			string: i18n.Text("Includes Fright Check Bonus"),
			alt:    i18n.Text("+%d Fright Check Bonus"),
		},
		{
			key:    "minor_cost_of_living_increase",
			string: i18n.Text("Includes a Minor Cost of Living Increase"),
			alt:    i18n.Text("+%d%% Cost of Living Increase"),
		},
		{
			key:    "major_cost_of_living_increase",
			string: i18n.Text("Includes a Major Cost of Living Increase and Merchant Skill Penalty"),
			alt:    i18n.Text("+%d%% Cost of Living Increase"),
		},
	}
)

// SelfControlRollAdj holds an Adjustment for a self-control roll.
type SelfControlRollAdj byte

// EnsureValid ensures this is of a known value.
func (enum SelfControlRollAdj) EnsureValid() SelfControlRollAdj {
	if enum <= LastSelfControlRollAdj {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum SelfControlRollAdj) Key() string {
	return selfControlRollAdjData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum SelfControlRollAdj) String() string {
	return selfControlRollAdjData[enum.EnsureValid()].string
}

// AltString returns the alternate string.
func (enum SelfControlRollAdj) AltString() string {
	return selfControlRollAdjData[enum.EnsureValid()].alt
}

// ExtractSelfControlRollAdj extracts the value from a string.
func ExtractSelfControlRollAdj(str string) SelfControlRollAdj {
	for i, one := range selfControlRollAdjData {
		if strings.EqualFold(one.key, str) {
			return SelfControlRollAdj(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum SelfControlRollAdj) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *SelfControlRollAdj) UnmarshalText(text []byte) error {
	*enum = ExtractSelfControlRollAdj(string(text))
	return nil
}
