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
	"fmt"
	"hash"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Bonus = &DRBonus{}

// DRBonusData is split out so that it can be adjusted before and after being serialized.
type DRBonusData struct {
	Type feature.Type `json:"type"`
	FeatureSwitch
	Locations      []string `json:"locations,omitempty"`
	Specialization string   `json:"specialization,omitzero"`
	LeveledAmount
}

// DRBonus holds the data for a DR adjustment.
type DRBonus struct {
	DRBonusData
	BonusOwner `json:"-"`
}

// NewDRBonus creates a new DRBonus.
func NewDRBonus() *DRBonus {
	return &DRBonus{
		Type:           feature.DRBonus,
		Locations:      []string{TorsoID},
		Specialization: AllID,
		Amount:         fxp.One,
	}
}

// FeatureType implements Feature.
func (d *DRBonus) FeatureType() feature.Type {
	return d.Type
}

// Clone implements Feature.
func (d *DRBonus) Clone() Feature {
	other := *d
	other.Locations = slices.Clone(d.Locations)
	return &other
}

// Normalize adjusts the data to it preferred representation.
func (d *DRBonus) Normalize() {
	for i, loc := range d.Locations {
		loc = strings.TrimSpace(loc)
		if strings.EqualFold(loc, AllID) {
			d.Locations = []string{AllID}
			break
		}
		d.Locations[i] = loc
	}
	d.Specialization = normalizeDRSpecialization(d.Specialization)
}

// normalizeDRSpecialization returns the preferred representation of a DR bonus specialization: surrounding whitespace
// is removed and an empty value or any casing of "all" becomes AllID.
func normalizeDRSpecialization(specialization string) string {
	s := strings.TrimSpace(specialization)
	if s == "" || strings.EqualFold(s, AllID) {
		return AllID
	}
	return s
}

// FillWithNameableKeys implements Feature.
func (d *DRBonus) FillWithNameableKeys(m, existing map[string]string) {
	nameable.Extract(m, existing, d.Specialization)
}

// SpecializationWithReplacements returns the damage type this bonus applies against, with the owning item's nameable
// replacements applied and normalized the same way the stored value is. An unresolved marker is left standing (as
// "@Label@") so that an unanswered substitution shows up visibly rather than silently becoming DR against everything.
func (d *DRBonus) SpecializationWithReplacements() string {
	return normalizeDRSpecialization(nameable.Apply(d.Specialization, bonusReplacements(d)))
}

// SetLeveledOwner implements Bonus.
func (d *DRBonus) SetLeveledOwner(owner LeveledOwner) {
	d.LeveledOwner = owner
}

// AddToTooltip implements Bonus.
func (d *DRBonus) AddToTooltip(buffer *xbytes.InsertBuffer) {
	if buffer != nil {
		fmt.Fprintf(buffer, i18n.Text("\n- %s [%s against %s attacks]"), d.parentName(), d.Format(),
			d.SpecializationWithReplacements())
	}
}

// MarshalJSONTo implements json.MarshalerTo.
func (d *DRBonus) MarshalJSONTo(enc *jsontext.Encoder) error {
	d.Normalize()
	if d.Specialization == AllID {
		d.Specialization = ""
	}
	err := json.MarshalEncode(enc, &d.DRBonusData)
	d.Normalize()
	return err
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (d *DRBonus) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var dataWithOld struct {
		DRBonusData
		Location string `json:"location"`
	}
	if err := json.UnmarshalDecode(dec, &dataWithOld); err != nil {
		return err
	}
	if dataWithOld.Location != "" {
		dataWithOld.Locations = append(dataWithOld.Locations, dataWithOld.Location)
	}
	d.DRBonusData = dataWithOld.DRBonusData
	d.Normalize()
	return nil
}

// Hash writes this object's contents into the hasher.
func (d *DRBonus) Hash(h hash.Hash) {
	if d == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, d.Type)
	xhash.Bool(h, d.Switchable)
	xhash.Num64(h, len(d.Locations))
	for _, loc := range d.Locations {
		xhash.StringWithLen(h, loc)
	}
	xhash.StringWithLen(h, d.Specialization)
	d.LeveledAmount.Hash(h)
}
