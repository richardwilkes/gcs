// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"encoding/binary"
	"hash"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

var _ Bonus = &DRBonus{}

// DRBonusData is split out so that it can be adjusted before and after being serialized.
type DRBonusData struct {
	Type           feature.Type `json:"type"`
	Location       string       `json:"location"`
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
			Location:       "torso",
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
	return &other
}

// Normalize adjusts the data to it preferred representation.
func (d *DRBonus) Normalize() {
	s := strings.TrimSpace(d.Location)
	if strings.EqualFold(s, AllID) {
		s = AllID
	}
	d.Location = s
	s = strings.TrimSpace(d.Specialization)
	if s == "" || strings.EqualFold(s, AllID) {
		s = AllID
	}
	d.Specialization = s
}

// FillWithNameableKeys implements Feature.
func (d *DRBonus) FillWithNameableKeys(_ map[string]string) {
}

// SetLevel implements Bonus.
func (d *DRBonus) SetLevel(level fxp.Int) {
	d.Level = level
}

// AddToTooltip implements Bonus.
func (d *DRBonus) AddToTooltip(buffer *xio.ByteBuffer) {
	if buffer != nil {
		d.Normalize()
		buffer.WriteByte('\n')
		buffer.WriteString(d.parentName())
		buffer.WriteString(" [")
		buffer.WriteString(d.LeveledAmount.Format(false))
		buffer.WriteString(i18n.Text(" against "))
		buffer.WriteString(d.Specialization)
		buffer.WriteString(i18n.Text(" attacks]"))
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
	d.DRBonusData = DRBonusData{}
	if err := json.Unmarshal(data, &d.DRBonusData); err != nil {
		return err
	}
	d.Normalize()
	return nil
}

// Hash writes this object's contents into the hasher.
func (d *DRBonus) Hash(h hash.Hash) {
	if d == nil {
		return
	}
	_ = binary.Write(h, binary.LittleEndian, d.Type)
	_, _ = h.Write([]byte(d.Location))
	_, _ = h.Write([]byte(d.Specialization))
	d.LeveledAmount.Hash(h)
}
