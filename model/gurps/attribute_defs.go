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
	"cmp"
	"context"
	"hash"
	"io/fs"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/fatal"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
)

var _ Hashable = &AttributeDefs{}

// AttributeDefs holds a set of AttributeDef objects.
type AttributeDefs struct {
	Set map[string]*AttributeDef
}

// ResolveAttributeName returns the name of the attribute, if possible.
func ResolveAttributeName(entity *Entity, attribute string) string {
	if def := AttributeDefsFor(entity).Set[attribute]; def != nil {
		return def.Name
	}
	return attribute
}

// AttributeDefsFor returns the AttributeDefs for the given Entity, or the global settings if the Entity is nil.
func AttributeDefsFor(entity *Entity) *AttributeDefs {
	return SheetSettingsFor(entity).Attributes
}

// DefaultAttributeIDFor returns the default attribute ID to use for the given Entity, which may be nil.
func DefaultAttributeIDFor(entity *Entity) string {
	list := AttributeDefsFor(entity).List(true)
	if len(list) != 0 {
		return list[0].ID()
	}
	return StrengthID
}

// AttributeIDFor looks up the preferred ID and if it cannot be found, falls back to a default. 'entity' may be nil.
func AttributeIDFor(entity *Entity, preferred string) string {
	defs := AttributeDefsFor(entity)
	if _, exists := defs.Set[preferred]; exists {
		return preferred
	}
	if list := defs.List(true); len(list) != 0 {
		return list[0].ID()
	}
	return StrengthID
}

// FactoryAttributeDefs returns the factory AttributeDef set.
func FactoryAttributeDefs() *AttributeDefs {
	defs, err := NewAttributeDefsFromFile(embeddedFS, "embedded_data/Standard.attr")
	fatal.IfErr(err)
	return defs
}

type attributeDefsData struct {
	Version int            `json:"version"`
	Rows    *AttributeDefs `json:"rows,alt=attributes"`
}

// NewAttributeDefsFromFile loads an AttributeDef set from a file.
func NewAttributeDefsFromFile(fileSystem fs.FS, filePath string) (*AttributeDefs, error) {
	var data struct {
		attributeDefsData
		OldestKey *AttributeDefs `json:"attribute_settings"`
	}
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &data); err != nil {
		return nil, errs.NewWithCause(InvalidFileData(), err)
	}
	if err := jio.CheckVersion(data.Version); err != nil {
		return nil, err
	}
	var defs *AttributeDefs
	if data.Rows != nil {
		defs = data.Rows
	}
	if defs == nil && data.OldestKey != nil {
		defs = data.OldestKey
	}
	if defs == nil {
		defs = FactoryAttributeDefs()
	}
	return defs, nil
}

// Save writes the AttributeDefs to the file as JSON.
func (a *AttributeDefs) Save(filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, &attributeDefsData{
		Version: jio.CurrentDataVersion,
		Rows:    a,
	})
}

// MarshalJSON implements json.Marshaler.
func (a *AttributeDefs) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	e := json.NewEncoder(&buffer)
	e.SetEscapeHTML(false)
	err := e.Encode(a.List(false))
	return buffer.Bytes(), err
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *AttributeDefs) UnmarshalJSON(data []byte) error {
	var list []*AttributeDef
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	a.Set = make(map[string]*AttributeDef, len(list))
	for i, one := range list {
		one.Order = i + 1
		a.Set[one.ID()] = one
	}
	return nil
}

// Clone a copy of this.
func (a *AttributeDefs) Clone() *AttributeDefs {
	clone := &AttributeDefs{Set: make(map[string]*AttributeDef)}
	for k, v := range a.Set {
		clone.Set[k] = v.Clone()
	}
	return clone
}

// List returns the map of AttributeDef objects as an ordered list.
func (a *AttributeDefs) List(omitSeparators bool) []*AttributeDef {
	list := make([]*AttributeDef, 0, len(a.Set))
	for _, v := range a.Set {
		if omitSeparators && v.IsSeparator() {
			continue
		}
		list = append(list, v)
	}
	slices.SortFunc(list, func(a, b *AttributeDef) int { return cmp.Compare(a.Order, b.Order) })
	return list
}

// Hash writes this object's contents into the hasher.
func (a *AttributeDefs) Hash(h hash.Hash) {
	hashhelper.Num64(h, len(a.Set))
	for _, one := range a.List(false) {
		one.Hash(h)
	}
}

// ResetTargetKeyPrefixes assigns new key prefixes for all data within these AttributeDefs.
func (a *AttributeDefs) ResetTargetKeyPrefixes(prefixProvider func() string) {
	for _, one := range a.Set {
		one.KeyPrefix = prefixProvider()
		for _, threshold := range one.Thresholds {
			threshold.KeyPrefix = prefixProvider()
		}
	}
}
