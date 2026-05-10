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
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"hash"
	"strconv"
	"strings"

	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xstrings"
)

var _ Hashable = &HitLocation{}

// HitLocationData holds the Hitlocation data that gets written to disk.
type HitLocationData struct {
	LocID       string `json:"id"`
	ChoiceName  string `json:"choice_name"`
	TableName   string `json:"table_name"`
	Slots       int    `json:"slots,omitzero"`
	HitPenalty  int    `json:"hit_penalty,omitzero"`
	DRBonus     int    `json:"dr_bonus,omitzero"`
	Description string `json:"description,omitzero"`
	Notes       string `json:"notes,omitzero"`
	SubTable    *Body  `json:"sub_table,omitzero"`
}

// HitLocation holds a single hit location.
type HitLocation struct {
	HitLocationData
	Entity      *Entity
	RollRange   string
	KeyPrefix   string
	owningTable *Body
}

// NewHitLocation creates a new hit location.
func NewHitLocation(entity *Entity, keyPrefix string) *HitLocation {
	return &HitLocation{
		HitLocationData: HitLocationData{
			LocID:      "id",
			ChoiceName: i18n.Text("untitled choice"),
			TableName:  i18n.Text("untitled location"),
		},
		Entity:    entity,
		KeyPrefix: keyPrefix,
	}
}

// Clone a copy of this.
func (h *HitLocation) Clone(entity *Entity, owningTable *Body) *HitLocation {
	clone := *h
	clone.Entity = entity
	clone.owningTable = owningTable
	if h.SubTable != nil {
		clone.SubTable = h.SubTable.Clone(entity, &clone)
	}
	return &clone
}

// MarshalJSONTo implements json.MarshalerTo.
func (h *HitLocation) MarshalJSONTo(enc *jsontext.Encoder) error {
	type calc struct {
		RollRange string         `json:"roll_range"`
		DR        map[string]int `json:"dr,omitzero"`
	}
	data := struct {
		HitLocationData
		Calc calc `json:"calc"`
	}{
		HitLocationData: h.HitLocationData,
		Calc: calc{
			RollRange: h.RollRange,
		},
	}
	if h.Entity != nil {
		data.Calc.DR = h.DR(h.Entity, nil, nil)
		if _, exists := data.Calc.DR[AllID]; !exists {
			data.Calc.DR[AllID] = 0
		}
	}
	return json.MarshalEncode(enc, &data)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (h *HitLocation) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	h.HitLocationData = HitLocationData{}
	if err := json.UnmarshalDecode(dec, &h.HitLocationData); err != nil {
		return err
	}
	if h.SubTable != nil {
		h.SubTable.SetOwningLocation(h)
	}
	return nil
}

// ID returns the ID.
func (h *HitLocation) ID() string {
	return h.LocID
}

// SetID sets the ID, sanitizing it in the process (i.e. it may be changed from what you set -- read it back if you want
// to be sure of what it gets set to.
func (h *HitLocation) SetID(value string) {
	h.LocID = SanitizeID(value, false, ReservedIDs...)
}

// OwningTable returns the owning table.
func (h *HitLocation) OwningTable() *Body {
	return h.owningTable
}

// DR computes the DR coverage for this HitLocation. If 'tooltip' isn't nil, the buffer will be updated with details on
// how the DR was calculated. If 'drMap' isn't nil, it will be returned.
func (h *HitLocation) DR(entity *Entity, tooltip *xbytes.InsertBuffer, drMap map[string]int) map[string]int {
	if drMap == nil {
		drMap = make(map[string]int)
	}
	if h.DRBonus != 0 {
		drMap[AllID] += h.DRBonus
		if tooltip != nil {
			fmt.Fprintf(tooltip, i18n.Text("\n- %s [%+d against %s attacks]"), h.ChoiceName, h.DRBonus, AllID)
		}
	}
	drMap = entity.AddDRBonusesFor(h.LocID, tooltip, drMap)
	if h.owningTable != nil && h.owningTable.owningLocation != nil {
		drMap = h.owningTable.owningLocation.DR(entity, tooltip, drMap)
	}
	if tooltip != nil && len(drMap) != 0 {
		keys := make([]string, 0, len(drMap))
		for k := range drMap {
			keys = append(keys, k)
		}
		xstrings.SortStringsNaturalAscending(keys)
		base := drMap[AllID]
		var buffer bytes.Buffer
		buffer.WriteByte('\n')
		for _, k := range keys {
			value := drMap[k]
			if !strings.EqualFold(AllID, k) {
				value += base
			}
			fmt.Fprintf(&buffer, i18n.Text("\n- **DR %d** against **%s** attacks"), value, k)
		}
		buffer.WriteString("\n---\n")
		_ = tooltip.Insert(0, buffer.Bytes())
	}
	return drMap
}

// PD computes the PD (Passive Defense) coverage for this HitLocation. If 'tooltip' isn't nil, the buffer will be
// updated with details on how the PD was calculated. If 'pdMap' isn't nil, it will be returned.
// PD only applies if UsePassiveDefense is enabled in sheet settings (GURPS 3e optional rule).
func (h *HitLocation) PD(entity *Entity, tooltip *xbytes.InsertBuffer, pdMap map[string]int) map[string]int {
	if pdMap == nil {
		pdMap = make(map[string]int)
	}
	// PD is an optional rule - only calculate if enabled
	if entity.SheetSettings == nil || !entity.SheetSettings.UsePassiveDefense {
		return pdMap
	}
	pdMap = entity.AddPDBonusesFor(h.LocID, tooltip, pdMap)
	if h.owningTable != nil && h.owningTable.owningLocation != nil {
		pdMap = h.owningTable.owningLocation.PD(entity, tooltip, pdMap)
	}
	if tooltip != nil && len(pdMap) != 0 {
		// PD is simpler than DR - it doesn't have specializations, just a total value
		// stored under PDSpecializationKey. Show the total PD value.
		pdValue := pdMap[PDSpecializationKey]
		if pdValue > 0 {
			var buffer bytes.Buffer
			buffer.WriteByte('\n')
			fmt.Fprintf(&buffer, i18n.Text("\n- **PD %d**"), pdValue)
			buffer.WriteString("\n---\n")
			_ = tooltip.Insert(0, buffer.Bytes())
		}
	}
	return pdMap
}

// DisplayPD returns the PD for this location, formatted as a string.
// Returns just the total PD value (not broken down by specialization like DR).
func (h *HitLocation) DisplayPD(entity *Entity, tooltip *xbytes.InsertBuffer) string {
	if entity.SheetSettings == nil || !entity.SheetSettings.UsePassiveDefense {
		return "0"
	}
	pdMap := h.PD(entity, tooltip, nil)
	// PD bonuses are stored under the PDSpecializationKey ("pd") from AddPDBonusesFor.
	// All PD bonuses for a location are accumulated under this single key.
	// For display, we just want the total PD value, not a ratio like DR.
	pdValue := pdMap[PDSpecializationKey]
	if pdValue == 0 {
		return "0"
	}
	return strconv.Itoa(pdValue)
}

// DisplayDR returns the DR for this location, formatted as a string.
func (h *HitLocation) DisplayDR(entity *Entity, tooltip *xbytes.InsertBuffer) string {
	drMap := h.DR(entity, tooltip, nil)
	all, exists := drMap[AllID]
	if !exists {
		drMap[AllID] = 0
	}
	keys := make([]string, 0, len(drMap))
	keys = append(keys, AllID)
	for k := range drMap {
		if k != AllID {
			keys = append(keys, k)
		}
	}
	xstrings.SortStringsNaturalAscending(keys[1:])
	var buffer strings.Builder
	for _, k := range keys {
		dr := drMap[k]
		if k != AllID {
			dr += all
		}
		if buffer.Len() != 0 {
			buffer.WriteByte('/')
		}
		buffer.WriteString(strconv.Itoa(dr))
	}
	return buffer.String()
}

// SetSubTable sets the Body as a sub-table.
func (h *HitLocation) SetSubTable(bodyType *Body) {
	if bodyType == nil && h.SubTable != nil {
		h.SubTable.SetOwningLocation(nil)
	}
	if h.SubTable = bodyType; h.SubTable != nil {
		h.SubTable.SetOwningLocation(h)
	}
}

func (h *HitLocation) populateMap(entity *Entity, m map[string]*HitLocation) {
	h.Entity = entity
	m[h.LocID] = h
	if h.SubTable != nil {
		h.SubTable.populateMap(entity, m)
	}
}

func (h *HitLocation) updateRollRange(start int) int {
	switch h.Slots {
	case 0:
		h.RollRange = "-"
	case 1:
		h.RollRange = strconv.Itoa(start)
	default:
		h.RollRange = fmt.Sprintf("%d-%d", start, start+h.Slots-1)
	}
	if h.SubTable != nil {
		h.SubTable.updateRollRanges()
	}
	return start + h.Slots
}

// Hash writes this object's contents into the hasher.
func (h *HitLocation) Hash(hasher hash.Hash) {
	xhash.StringWithLen(hasher, h.LocID)
	xhash.StringWithLen(hasher, h.ChoiceName)
	xhash.StringWithLen(hasher, h.TableName)
	xhash.Num64(hasher, h.Slots)
	xhash.Num64(hasher, h.HitPenalty)
	xhash.Num64(hasher, h.DRBonus)
	xhash.StringWithLen(hasher, h.Description)
	xhash.StringWithLen(hasher, h.Notes)
	if h.SubTable != nil {
		h.SubTable.Hash(hasher)
	} else {
		xhash.Num8(hasher, uint8(255))
	}
}

// ResetTargetKeyPrefixes assigns new key prefixes for all data within this HitLocation.
func (h *HitLocation) ResetTargetKeyPrefixes(prefixProvider func() string) {
	h.KeyPrefix = prefixProvider()
	if h.SubTable != nil {
		h.SubTable.ResetTargetKeyPrefixes(prefixProvider)
	}
}

func (h *HitLocation) rewrap() {
	h.Description = xstrings.Wrap("", h.Description, 60)
	if h.SubTable != nil {
		for _, loc := range h.SubTable.Locations {
			loc.rewrap()
		}
	}
}

// IsOpen returns true if this location contains a sub-table and it is open.
func (h *HitLocation) IsOpen(indexes []int) bool {
	return h.SubTable != nil && !IsClosed(h.openKey(indexes))
}

// SetOpen sets the open state of this location. Does nothing if this location does not contain a sub-table.
func (h *HitLocation) SetOpen(indexes []int, open bool) {
	if h.SubTable != nil {
		SetClosedState(h.openKey(indexes), !open)
	}
}

func (h *HitLocation) openKey(indexes []int) string {
	var buffer strings.Builder
	buffer.WriteString("b:")
	if h.Entity != nil {
		buffer.WriteString(string(h.Entity.ID))
	} else {
		buffer.WriteString("-")
	}
	for _, i := range indexes {
		fmt.Fprintf(&buffer, ":%d", i)
	}
	return buffer.String()
}
