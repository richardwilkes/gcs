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
	"encoding/json"
	"fmt"
	"hash"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Bonus = &DRBonus{}

// DRBonusData is split out so that it can be adjusted before and after being serialized.
type DRBonusData struct {
	Type           feature.Type `json:"type"`
	Locations      []string     `json:"locations,omitempty"`
	Specialization string       `json:"specialization,omitempty"`
	LeveledAmount
}

// DRBonus holds the data for a DR adjustment.
type DRBonus struct {
	DRBonusData
	BonusOwner
}

// NewDRBonus creates a new DRBonus.
func NewDRBonus() *DRBonus {
	return &DRBonus{
		DRBonusData: DRBonusData{
			Type:           feature.DRBonus,
			Locations:      []string{TorsoID},
			Specialization: AllID,
			LeveledAmount:  LeveledAmount{Amount: fxp.One},
		},
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
	s := strings.TrimSpace(d.Specialization)
	if s == "" || strings.EqualFold(s, AllID) {
		s = AllID
	}
	d.Specialization = s
}

// FillWithNameableKeys implements Feature.
func (d *DRBonus) FillWithNameableKeys(_, _ map[string]string) {
}

// SetLeveledOwner implements Bonus.
func (d *DRBonus) SetLeveledOwner(owner LeveledOwner) {
	d.LeveledOwner = owner
}

// AddToTooltip implements Bonus.
func (d *DRBonus) AddToTooltip(buffer *xbytes.InsertBuffer) {
	if buffer != nil {
		d.Normalize()
		fmt.Fprintf(buffer, i18n.Text("\n- %s [%s against %s attacks]"), d.parentName(), d.Format(), d.Specialization)
	}
}

// MarshalJSON implements json.Marshaler.
func (d *DRBonus) MarshalJSON() ([]byte, error) {
	d.Normalize()
	if d.Specialization == AllID {
		d.Specialization = ""
	}
	data, err := json.Marshal(&d.DRBonusData)
	d.Normalize()
	return data, err
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *DRBonus) UnmarshalJSON(data []byte) error {
	var dataWithOld struct {
		DRBonusData
		Location string `json:"location"`
	}
	if err := json.Unmarshal(data, &dataWithOld); err != nil {
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
	xhash.Num64(h, len(d.Locations))
	for _, loc := range d.Locations {
		xhash.StringWithLen(h, loc)
	}
	xhash.StringWithLen(h, d.Specialization)
	d.LeveledAmount.Hash(h)
}
