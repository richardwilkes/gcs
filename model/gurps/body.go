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
	"context"
	"embed"
	"hash"
	"io/fs"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/fatal"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
)

var _ Hashable = &Body{}

const noNeedForRewrapVersion = 4

//go:embed embedded_data
var embeddedFS embed.FS

// BodyData holds the Body data that gets written to disk.
type BodyData struct {
	Name      string         `json:"name,omitempty"`
	Roll      *dice.Dice     `json:"roll"`
	Locations []*HitLocation `json:"locations,omitempty"`
}

// Body holds a set of hit locations.
type Body struct {
	BodyData
	KeyPrefix      string
	owningLocation *HitLocation
	locationLookup map[string]*HitLocation
}

type standaloneBodyData struct {
	Version int `json:"version"`
	BodyData
	OldHitLocations *Body `json:"hit_locations,omitempty"`
}

// BodyFor returns the Body for the given Entity, or the global settings if the Entity is nil.
func BodyFor(entity *Entity) *Body {
	return SheetSettingsFor(entity).BodyType
}

// FactoryBody returns a new copy of the default factory Body.
func FactoryBody() *Body {
	bodyType, err := NewBodyFromFile(embeddedFS, "embedded_data/Humanoid.body")
	fatal.IfErr(err)
	return bodyType
}

// NewBodyFromFile loads a Body from a file.
func NewBodyFromFile(fileSystem fs.FS, filePath string) (*Body, error) {
	var data standaloneBodyData
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &data); err != nil {
		return nil, errs.NewWithCause(InvalidFileData(), err)
	}
	if err := jio.CheckVersion(data.Version); err != nil {
		return nil, err
	}
	var body Body
	if data.OldHitLocations == nil {
		body.BodyData = data.BodyData
	} else {
		body.BodyData = data.OldHitLocations.BodyData
	}
	body.Update(nil)
	if data.Version < noNeedForRewrapVersion {
		body.Rewrap()
	}
	return &body, nil
}

// MarshalJSON implements json.Marshaler.
func (b *Body) MarshalJSON() ([]byte, error) {
	return json.Marshal(&b.BodyData)
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Body) UnmarshalJSON(data []byte) error {
	b.BodyData = BodyData{}
	if err := json.Unmarshal(data, &b.BodyData); err != nil {
		return err
	}
	b.Update(nil)
	return nil
}

// Rewrap the description field. Should only be called for older data (prior to noNeedForRewrapVersion)
func (b *Body) Rewrap() {
	for _, loc := range b.Locations {
		loc.rewrap()
	}
}

// Clone a copy of this.
func (b *Body) Clone(entity *Entity, owningLocation *HitLocation) *Body {
	clone := &Body{
		BodyData: BodyData{
			Name:      b.Name,
			Roll:      dice.New(b.Roll.String()),
			Locations: make([]*HitLocation, len(b.Locations)),
		},
		owningLocation: owningLocation,
	}
	for i, one := range b.Locations {
		clone.Locations[i] = one.Clone(entity, clone)
	}
	clone.Update(entity)
	return clone
}

// Save writes the Body to the file as JSON.
func (b *Body) Save(filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, &standaloneBodyData{
		Version:  jio.CurrentDataVersion,
		BodyData: b.BodyData,
	})
}

// Update the role ranges and populate the lookup map.
func (b *Body) Update(entity *Entity) {
	for _, one := range b.Locations {
		one.owningTable = b
	}
	b.updateRollRanges()
	b.locationLookup = make(map[string]*HitLocation)
	b.populateMap(entity, b.locationLookup)
}

// OwningLocation returns the owning hit location, or nil if this is the top-level body.
func (b *Body) OwningLocation() *HitLocation {
	return b.owningLocation
}

// SetOwningLocation sets the owning HitLocation.
func (b *Body) SetOwningLocation(loc *HitLocation) {
	b.owningLocation = loc
	if loc != nil {
		b.Name = ""
	}
}

func (b *Body) updateRollRanges() {
	start := b.Roll.Minimum(false)
	for _, location := range b.Locations {
		start = location.updateRollRange(start)
	}
}

func (b *Body) populateMap(entity *Entity, m map[string]*HitLocation) {
	for _, location := range b.Locations {
		location.populateMap(entity, m)
	}
}

// AddLocation adds a HitLocation to the end of list.
func (b *Body) AddLocation(loc *HitLocation) {
	b.Locations = append(b.Locations, loc)
	loc.owningTable = b
}

// RemoveLocation removes a HitLocation.
func (b *Body) RemoveLocation(loc *HitLocation) {
	for i, one := range b.Locations {
		if one == loc {
			copy(b.Locations[i:], b.Locations[i+1:])
			b.Locations[len(b.Locations)-1] = nil
			b.Locations = b.Locations[:len(b.Locations)-1]
			loc.owningTable = nil
		}
	}
}

// UniqueHitLocations returns the list of unique hit locations.
func (b *Body) UniqueHitLocations(entity *Entity) []*HitLocation {
	if len(b.locationLookup) == 0 {
		b.Update(entity)
	}
	locations := make([]*HitLocation, 0, len(b.locationLookup))
	for _, v := range b.locationLookup {
		locations = append(locations, v)
	}
	slices.SortFunc(locations, func(a, b *HitLocation) int {
		result := txt.NaturalCmp(a.ChoiceName, b.ChoiceName, true)
		if result == 0 {
			result = txt.NaturalCmp(a.ID(), b.ID(), true)
		}
		return result
	})
	return locations
}

// LookupLocationByID returns the HitLocation that matches the given ID.
func (b *Body) LookupLocationByID(entity *Entity, idStr string) *HitLocation {
	if len(b.locationLookup) == 0 {
		b.Update(entity)
	}
	return b.locationLookup[idStr]
}

// Hash writes this object's contents into the hasher.
func (b *Body) Hash(h hash.Hash) {
	hashhelper.String(h, b.Name)
	b.Roll.Hash(h)
	hashhelper.Num64(h, len(b.Locations))
	for _, one := range b.Locations {
		one.Hash(h)
	}
}

// ResetTargetKeyPrefixes assigns new key prefixes for all data within this Body.
func (b *Body) ResetTargetKeyPrefixes(prefixProvider func() string) {
	b.KeyPrefix = prefixProvider()
	for _, one := range b.Locations {
		one.ResetTargetKeyPrefixes(prefixProvider)
	}
}

// ParentIndexes returns the indexes of the owning location and table.
func (b *Body) ParentIndexes() []int {
	if b.owningLocation == nil {
		return nil
	}
	indexes := b.owningLocation.owningTable.ParentIndexes()
	for i, loc := range b.owningLocation.owningTable.Locations {
		if loc == b.owningLocation {
			return append(indexes, i)
		}
	}
	return nil
}

// FirstDisclosureState returns the open state of the first row that can be opened.
func (b *Body) FirstDisclosureState() (open, exists bool) {
	for i, one := range b.Locations {
		if one.SubTable != nil {
			return one.IsOpen(append(b.ParentIndexes(), i)), true
		}
	}
	return false, false
}

// SetDisclosureState sets the open state of all rows that can be opened.
func (b *Body) SetDisclosureState(open bool) {
	b.setDisclosureState(b.ParentIndexes(), open)
}

func (b *Body) setDisclosureState(parents []int, open bool) {
	for i, one := range b.Locations {
		if one.SubTable == nil {
			continue
		}
		indexes := make([]int, len(parents)+1)
		copy(indexes, parents)
		indexes[len(parents)] = i
		one.SetOpen(indexes, open)
		one.SubTable.setDisclosureState(indexes, open)
	}
}
