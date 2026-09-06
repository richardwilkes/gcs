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
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// situationBonus is the shared portion of a bonus that is keyed by a free-form situation description rather than by a
// selector. ConditionalModifierBonus and ReactionBonus embed it; they remain distinct types because the entity and
// the feature editor dispatch on them, and each keeps its own Clone and nil-guarded Hash.
type situationBonus struct {
	Situation string `json:"situation,omitzero"`
	LeveledAmount
	BonusOwner `json:"-"`
}

// situationKeyed is implemented by the bonuses that embed situationBonus. It lets the entity collect them generically.
type situationKeyed interface {
	Bonus
	situation() string
}

func (s *situationBonus) situation() string {
	return s.Situation
}

// FillWithNameableKeys implements Feature.
func (s *situationBonus) FillWithNameableKeys(m, existing map[string]string) {
	nameable.Extract(m, existing, s.Situation)
}

// SetLeveledOwner implements Bonus.
func (s *situationBonus) SetLeveledOwner(owner LeveledOwner) {
	s.LeveledOwner = owner
}

// AddToTooltip implements Bonus.
func (s *situationBonus) AddToTooltip(buffer *xbytes.InsertBuffer) {
	s.basicAddToTooltip(&s.LeveledAmount, buffer)
}

// hashSituation writes the situation and amount into the hasher. The concrete types write their own type and switch
// first, so that the nil guard and the type discriminator stay with them.
func (s *situationBonus) hashSituation(h hash.Hash) {
	xhash.StringWithLen(h, s.Situation)
	s.Hash(h) // the embedded LeveledAmount
}

// situationModifiersFromFeatureList adds the amount of each feature in the list that is a T to the conditional
// modifier for its situation in m, creating the modifier if this is the first contribution to that situation. The
// situation has the owner's nameable replacements applied first, so that two features reading "@Target@" resolve to
// separate entries when their owners fill the key in differently.
func situationModifiersFromFeatureList[T situationKeyed](source string, features Features, m map[string]*ConditionalModifier) {
	for _, f := range features {
		if bonus, ok := f.(T); ok {
			addSituationModifier(m, source, nameable.Apply(bonus.situation(), bonusReplacements(bonus)), bonus.AdjustedAmount())
		}
	}
}

// addSituationModifier adds amt from source to the conditional modifier for situation in m, creating it if necessary.
func addSituationModifier(m map[string]*ConditionalModifier, source, situation string, amt fxp.Int) {
	if r, exists := m[situation]; exists {
		r.Add(source, amt)
	} else {
		m[situation] = NewConditionalModifier(source, situation, amt)
	}
}
