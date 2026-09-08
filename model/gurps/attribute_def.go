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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/progression"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// Possible attribute kinds.
const (
	PrimaryAttrKind = iota
	SecondaryAttrKind
	PoolAttrKind
)

var _ Hashable = &AttributeDef{}

// ReservedIDs holds a list of IDs that are reserved for internal use.
var ReservedIDs = []string{SkillID, ParryID, BlockID, SizeModifierID, "10"}

// AttributeDef holds the definition of an attribute.
type AttributeDef struct {
	AttributeDefData
	Order     int
	KeyPrefix string
}

// AttributeDefData holds the data that will be serialized for the AttributeDef.
type AttributeDefData struct {
	DefID                string              `json:"id"`
	Type                 attribute.Type      `json:"type"`
	Placement            attribute.Placement `json:"placement,omitzero"`
	PlacementWhenPresent attribute.Placement `json:"placement_when_present,omitzero"`
	PlacementTrait       string              `json:"placement_trait,omitzero"`
	Name                 string              `json:"name"`
	FullName             string              `json:"full_name,omitzero"`
	Base                 string              `json:"base,omitzero"`
	CostPerPoint         fxp.Int             `json:"cost_per_point,omitzero"`
	CostAdjPercentPerSM  fxp.Int             `json:"cost_adj_percent_per_sm,omitzero"`
	Thresholds           []*PoolThreshold    `json:"thresholds,omitempty"`
}

// MarshalJSONTo implements json.MarshalerTo.
func (a *AttributeDef) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, &a.AttributeDefData)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (a *AttributeDef) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var legacy struct {
		AttributeDefData
		// Old data fields
		AttributeBase string `json:"attribute_base"`
	}
	if err := json.UnmarshalDecode(dec, &legacy); err != nil {
		return err
	}
	a.AttributeDefData = legacy.AttributeDefData
	if a.Base == "" && legacy.AttributeBase != "" {
		a.Base = ExprToScript(legacy.AttributeBase)
	}
	a.Thresholds = slices.DeleteFunc(a.Thresholds, func(one *PoolThreshold) bool { return one == nil })
	for _, threshold := range a.Thresholds {
		threshold.Value = strings.ReplaceAll(threshold.Value, "$self", "$"+a.DefID)
	}
	return nil
}

// Clone returns a copy of this AttributeDef. Thresholds are only carried over for the pool types.
func (a *AttributeDef) Clone() *AttributeDef {
	clone := *a
	if a.Type == attribute.Pool || a.Type == attribute.PoolRef {
		if a.Thresholds != nil {
			clone.Thresholds = make([]*PoolThreshold, len(a.Thresholds))
			for i, one := range a.Thresholds {
				clone.Thresholds[i] = one.Clone()
			}
		}
	} else {
		clone.Thresholds = nil
	}
	return &clone
}

// ID returns the ID.
func (a *AttributeDef) ID() string {
	return a.DefID
}

// SetID sets the ID, sanitizing it in the process, so the stored value may differ from what was passed in.
func (a *AttributeDef) SetID(value string) {
	a.DefID = SanitizeID(value, false, ReservedIDs...)
}

// ResolveFullName returns the full name, using the short name if full name is empty.
func (a *AttributeDef) ResolveFullName() string {
	if a.FullName == "" {
		return a.Name
	}
	return a.FullName
}

// CombinedName returns the combined FullName and Name, as appropriate.
func (a *AttributeDef) CombinedName() string {
	if a.FullName == "" {
		return a.Name
	}
	if a.Name == "" || a.Name == a.FullName {
		return a.FullName
	}
	return a.FullName + " (" + a.Name + ")"
}

// IsSeparator returns true if this is actually just a separator.
func (a *AttributeDef) IsSeparator() bool {
	return a.Type == attribute.PrimarySeparator || a.Type == attribute.SecondarySeparator || a.Type == attribute.PoolSeparator
}

// EffectivePlacement returns the placement to use for this attribute. When the Placement is Hidden, a PlacementTrait
// has been specified, and the entity has an enabled trait with that name, PlacementWhenPresent is returned instead. A
// nil entity or an empty PlacementTrait yields the unmodified Placement.
func (a *AttributeDef) EffectivePlacement(entity *Entity) attribute.Placement {
	if a.Placement == attribute.Hidden && a.PlacementTrait != "" && entity.HasTraitNamed(a.PlacementTrait) {
		return a.PlacementWhenPresent
	}
	return a.Placement
}

// Kind returns the kind of attribute this is, resolved against the given entity: PrimaryAttrKind, SecondaryAttrKind,
// PoolAttrKind, or -1 for a pool whose effective placement is not automatic.
func (a *AttributeDef) Kind(entity *Entity) int {
	switch a.Type {
	case attribute.PrimarySeparator:
		return PrimaryAttrKind
	case attribute.SecondarySeparator:
		return SecondaryAttrKind
	case attribute.PoolSeparator:
		return PoolAttrKind
	case attribute.Pool, attribute.PoolRef:
		if a.EffectivePlacement(entity) == attribute.Automatic {
			return PoolAttrKind
		}
		return -1
	default:
		switch a.EffectivePlacement(entity) {
		case attribute.Primary:
			return PrimaryAttrKind
		case attribute.Secondary:
			return SecondaryAttrKind
		default:
			// Automatic (or hidden) placement: a base that is a plain number is a primary attribute; a base that is an
			// expression is derived from other attributes, so it is a secondary one.
			if _, err := fxp.FromString(strings.TrimSpace(a.Base)); err == nil {
				return PrimaryAttrKind
			}
			return SecondaryAttrKind
		}
	}
}

// Relevant returns true if the attribute is relevant to the given kind, resolved against the given entity.
func (a *AttributeDef) Relevant(entity *Entity, kind int) bool {
	return a.EffectivePlacement(entity) != attribute.Hidden && a.Kind(entity) == kind
}

// Primary returns true if the base value is a non-derived value, resolved against the given entity.
func (a *AttributeDef) Primary(entity *Entity) bool {
	return a.Kind(entity) == PrimaryAttrKind
}

// Secondary returns true if the base value is a derived value, resolved against the given entity.
func (a *AttributeDef) Secondary(entity *Entity) bool {
	return a.Kind(entity) == SecondaryAttrKind
}

// Pool returns true if the base value is a pool value, resolved against the given entity.
func (a *AttributeDef) Pool(entity *Entity) bool {
	return a.Kind(entity) == PoolAttrKind
}

// AllowsDecimal returns true if the value can have a decimal point in it.
func (a *AttributeDef) AllowsDecimal() bool {
	return a.Type == attribute.Decimal || a.Type == attribute.DecimalRef
}

// BaseValue returns the resolved base value.
func (a *AttributeDef) BaseValue(attr *Attribute) fxp.Int {
	if a.IsSeparator() {
		return 0
	}
	return ResolveToNumber(attr.Entity, deferredNewScriptAttribute(attr), a.Base)
}

// ComputeCost returns the value adjusted for a cost reduction.
func (a *AttributeDef) ComputeCost(entity *Entity, value, costReduction fxp.Int, sizeModifier int) fxp.Int {
	if a.IsSeparator() {
		return 0
	}
	cost := value.Mul(a.CostPerPoint)
	if sizeModifier > 0 && a.CostAdjPercentPerSM > 0 &&
		(a.DefID != HitPointsID || entity.SheetSettings.DamageProgression != progression.KnowingYourOwnStrength) {
		costReduction += fxp.FromInteger(sizeModifier).Mul(a.CostAdjPercentPerSM)
	}
	if costReduction > 0 {
		if costReduction > fxp.Eighty {
			costReduction = fxp.Eighty
		}
		cost = cost.Mul(fxp.Hundred - costReduction).Div(fxp.Hundred)
	}
	return fxp.ApplyRounding(cost, false)
}

// Hash writes this object's contents into the hasher.
func (a *AttributeDef) Hash(h hash.Hash) {
	xhash.StringWithLen(h, a.DefID)
	xhash.Num8(h, a.Type)
	xhash.Num8(h, a.Placement)
	xhash.StringWithLen(h, a.PlacementTrait)
	xhash.Num8(h, a.PlacementWhenPresent)
	xhash.StringWithLen(h, a.Name)
	xhash.StringWithLen(h, a.FullName)
	xhash.StringWithLen(h, a.Base)
	xhash.Num64(h, a.CostPerPoint)
	xhash.Num64(h, a.CostAdjPercentPerSM)
	hashList(h, a.Thresholds)
}

// IsOpen returns true if this attribute is a separator and it is open.
func (a *AttributeDef) IsOpen(entity *Entity, sepCount int) bool {
	return a.IsSeparator() && !IsClosed(a.openKey(entity, a.Kind(entity), sepCount))
}

// SetOpen sets the open state of this attribute. Does nothing if this attribute is not a separator.
func (a *AttributeDef) SetOpen(entity *Entity, sepCount int, open bool) {
	if a.IsSeparator() {
		SetClosedState(a.openKey(entity, a.Kind(entity), sepCount), !open)
	}
}

func (a *AttributeDef) openKey(entity *Entity, kind, sepCount int) string {
	var buffer strings.Builder
	buffer.WriteString("a:")
	if entity != nil {
		buffer.WriteString(string(entity.ID))
	} else {
		buffer.WriteString("-")
	}
	fmt.Fprintf(&buffer, ":%d:%d", kind, sepCount)
	return buffer.String()
}
