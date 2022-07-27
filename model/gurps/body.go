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
	"context"
	"embed"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/crc"
	"github.com/richardwilkes/gcs/v5/model/gurps/gid"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/txt"
)

const (
	bodyTypeListTypeKey    = "body_type"
	oldBodyTypeListTypeKey = "hit_locations"
)

const noNeedForRewrapVersion = 4

//go:embed data
var embeddedFS embed.FS

// Body holds a set of hit locations.
type Body struct {
	Name           string         `json:"name,omitempty"`
	Roll           *dice.Dice     `json:"roll"`
	Locations      []*HitLocation `json:"locations,omitempty"`
	KeyPrefix      string         `json:"-"`
	owningLocation *HitLocation
	locationLookup map[string]*HitLocation
}

type bodyData struct {
	Type    string `json:"type"`
	Version int    `json:"version"`
	*Body
}

// BodyFor returns the Body for the given Entity, or the global settings if the Entity is nil.
func BodyFor(entity *Entity) *Body {
	return SheetSettingsFor(entity).BodyType
}

// FactoryBody returns a new copy of the default factory Body.
func FactoryBody() *Body {
	bodyType, err := NewBodyFromFile(embeddedFS, "data/body/Humanoid.body")
	jot.FatalIfErr(err)
	return bodyType
}

// FactoryBodies returns the list of the known factory Body types.
func FactoryBodies() []*Body {
	entries, err := embeddedFS.ReadDir("data/body")
	jot.FatalIfErr(err)
	list := make([]*Body, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if path.Ext(name) == ".body" {
			var bodyType *Body
			bodyType, err = NewBodyFromFile(embeddedFS, "data/body/"+name)
			jot.FatalIfErr(err)
			list = append(list, bodyType)
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return txt.NaturalLess(list[i].Name, list[j].Name, true)
	})
	return list
}

// NewBodyFromFile loads a Body from a file.
func NewBodyFromFile(fileSystem fs.FS, filePath string) (*Body, error) {
	var data struct {
		bodyData
		OldHitLocations *Body `json:"hit_locations"`
	}
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &data); err != nil {
		return nil, errs.NewWithCause(gid.InvalidFileDataMsg, err)
	}
	if data.Type != bodyTypeListTypeKey {
		if data.Type == oldBodyTypeListTypeKey {
			data.Body = data.OldHitLocations
		} else {
			return nil, errs.New(gid.UnexpectedFileDataMsg)
		}
	}
	if err := gid.CheckVersion(data.Version); err != nil {
		return nil, err
	}
	data.Body.EnsureValidity()
	if data.Version < noNeedForRewrapVersion {
		data.Body.Rewrap()
	}
	data.Body.Update(nil)
	return data.Body, nil
}

// EnsureValidity checks the current settings for validity and if they aren't valid, makes them so.
func (b *Body) EnsureValidity() {
	// TODO: Implement validity check
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
		Name:           b.Name,
		Roll:           dice.New(b.Roll.String()),
		Locations:      make([]*HitLocation, len(b.Locations)),
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
	return jio.SaveToFile(context.Background(), filePath, &bodyData{
		Type:    bodyTypeListTypeKey,
		Version: gid.CurrentDataVersion,
		Body:    b,
	})
}

// Update the role ranges and populate the lookup map.
func (b *Body) Update(entity *Entity) {
	b.updateRollRanges()
	b.locationLookup = make(map[string]*HitLocation)
	b.populateMap(entity, b.locationLookup)
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
	sort.Slice(locations, func(i, j int) bool {
		if txt.NaturalLess(locations[i].ChoiceName, locations[j].ChoiceName, false) {
			return true
		}
		if strings.EqualFold(locations[i].ChoiceName, locations[j].ChoiceName) {
			return txt.NaturalLess(locations[i].ID(), locations[j].ID(), false)
		}
		return false
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

// CRC64 calculates a CRC-64 for this data.
func (b *Body) CRC64() uint64 {
	return b.crc64(0)
}

func (b *Body) crc64(c uint64) uint64 {
	c = crc.String(c, b.Name)
	c = crc.String(c, b.Roll.String())
	c = crc.Number(c, len(b.Locations))
	for _, loc := range b.Locations {
		c = loc.crc64(c)
	}
	return c
}

// ResetTargetKeyPrefixes assigns new key prefixes for all data within this Body.
func (b *Body) ResetTargetKeyPrefixes(prefixProvider func() string) {
	b.KeyPrefix = prefixProvider()
	for _, one := range b.Locations {
		one.ResetTargetKeyPrefixes(prefixProvider)
	}
}
