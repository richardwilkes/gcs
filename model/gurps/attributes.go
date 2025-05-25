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
	"hash"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
)

var _ Hashable = &Attributes{}

// Attributes holds a set of Attribute objects.
type Attributes struct {
	Set map[string]*Attribute
}

// NewAttributes creates a new Attributes.
func NewAttributes(entity *Entity) *Attributes {
	a := &Attributes{Set: make(map[string]*Attribute)}
	for attrID, def := range entity.SheetSettings.Attributes.Set {
		a.Set[attrID] = NewAttribute(entity, attrID, def.Order)
	}
	return a
}

// MarshalJSON implements json.Marshaler.
func (a *Attributes) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	e := json.NewEncoder(&buffer)
	e.SetEscapeHTML(false)
	err := e.Encode(a.List())
	return buffer.Bytes(), err
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *Attributes) UnmarshalJSON(data []byte) error {
	var list []*Attribute
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	a.Set = make(map[string]*Attribute, len(list))
	for i, one := range list {
		one.Order = i
		a.Set[one.ID()] = one
	}
	return nil
}

// Clone a copy of this.
func (a *Attributes) Clone(entity *Entity) *Attributes {
	clone := &Attributes{Set: make(map[string]*Attribute)}
	for k, v := range a.Set {
		clone.Set[k] = v.Clone(entity)
	}
	return clone
}

// List returns the map of Attribute objects as an ordered list.
func (a *Attributes) List() []*Attribute {
	list := make([]*Attribute, 0, len(a.Set))
	for _, v := range a.Set {
		list = append(list, v)
	}
	slices.SortFunc(list, func(a, b *Attribute) int { return cmp.Compare(a.Order, b.Order) })
	return list
}

// Hash writes this object's contents into the hasher.
func (a *Attributes) Hash(h hash.Hash) {
	hashhelper.Num64(h, len(a.Set))
	for _, one := range a.List() {
		one.Hash(h)
	}
}

// Find resolves the given ID or name to an Attribute, or nil if not found.
func (a *Attributes) Find(idOrName string) *Attribute {
	if attr, ok := a.Set[idOrName]; ok {
		return attr
	}
	list := a.List()
	for _, one := range list {
		if one.NameMatches(idOrName) {
			return one
		}
	}
	return nil
}

// Cost returns the points spent for the specified Attribute.
func (a *Attributes) Cost(attrID string) fxp.Int {
	if attr, ok := a.Set[attrID]; ok {
		return attr.PointCost()
	}
	return 0
}

// Current resolves the given attribute ID to its current value, or fxp.Min.
func (a *Attributes) Current(attrID string) fxp.Int {
	if attr, ok := a.Set[attrID]; ok {
		return attr.Current()
	}
	if v, err := fxp.FromString(attrID); err == nil {
		return v
	}
	return fxp.Min
}

// Maximum resolves the given attribute ID to its maximum value, or fxp.Min.
func (a *Attributes) Maximum(attrID string) fxp.Int {
	if attr, ok := a.Set[attrID]; ok {
		return attr.Maximum()
	}
	if v, err := fxp.FromString(attrID); err == nil {
		return v
	}
	return fxp.Min
}

// PoolThreshold resolves the given attribute ID and state to the value for its pool threshold, or fxp.Min.
func (a *Attributes) PoolThreshold(attrID, state string) fxp.Int {
	if attr, ok := a.Set[attrID]; ok {
		if def := attr.AttributeDef(); def != nil {
			for _, one := range def.Thresholds {
				if strings.EqualFold(one.State, state) {
					return one.Threshold(attr.Entity)
				}
			}
		}
	}
	return fxp.Min
}

// FirstDisclosureState returns the open state of the first row that can be opened.
func (a *Attributes) FirstDisclosureState(entity *Entity) (open, exists bool) {
	for _, one := range a.List() {
		def := one.AttributeDef()
		if def != nil && def.IsSeparator() {
			return def.IsOpen(entity, 0), true
		}
	}
	return false, false
}

// SetDisclosureState sets the open state of all rows that can be opened.
func (a *Attributes) SetDisclosureState(entity *Entity, open bool) {
	m := make(map[int]int)
	for _, one := range a.List() {
		def := one.AttributeDef()
		if def != nil && def.IsSeparator() {
			kind := def.Kind()
			sepCount := m[kind]
			m[kind] = sepCount + 1
			def.SetOpen(entity, sepCount, open)
		}
	}
}
