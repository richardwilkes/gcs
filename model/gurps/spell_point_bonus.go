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
	"encoding/json/jsontext"
	"encoding/json/v2"
	"hash"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Bonus = &SpellPointBonus{}

// SpellPointBonus holds an adjustment to a spell's points.
type SpellPointBonus struct {
	SpellBonusData
}

// NewSpellPointBonus creates a new SpellPointBonus.
func NewSpellPointBonus() *SpellPointBonus {
	return &SpellPointBonus{SpellBonusData: newSpellBonusData(feature.SpellPointBonus)}
}

// Clone implements Feature.
func (s *SpellPointBonus) Clone() Feature {
	return clonePtr(s)
}

// Hash writes this object's contents into the hasher.
func (s *SpellPointBonus) Hash(h hash.Hash) {
	if s == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	s.SpellBonusData.Hash(h)
}

// MarshalJSONTo implements json.MarshalerTo.
func (s *SpellPointBonus) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, &s.SpellBonusData)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (s *SpellPointBonus) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return unmarshalWithLegacyTags(dec, &s.SpellBonusData, &s.TagsCriteria)
}
