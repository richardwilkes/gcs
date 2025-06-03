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
	"fmt"
	"hash"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/progression"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
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
	DefID               string              `json:"id"`
	Type                attribute.Type      `json:"type"`
	Placement           attribute.Placement `json:"placement,omitempty"`
	Name                string              `json:"name"`
	FullName            string              `json:"full_name,omitempty"`
	Base                string              `json:"base,omitempty"`
	CostPerPoint        fxp.Int             `json:"cost_per_point,omitempty"`
	CostAdjPercentPerSM fxp.Int             `json:"cost_adj_percent_per_sm,omitempty"`
	Thresholds          []*PoolThreshold    `json:"thresholds,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (a *AttributeDef) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	e := json.NewEncoder(&buffer)
	e.SetEscapeHTML(false)
	err := e.Encode(&a.AttributeDefData)
	return buffer.Bytes(), err
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *AttributeDef) UnmarshalJSON(data []byte) error {
	var legacy struct {
		AttributeDefData
		// Old data fields
		AttributeBase string `json:"attribute_base"`
	}
	if err := json.Unmarshal(data, &legacy); err != nil {
		return err
	}
	a.AttributeDefData = legacy.AttributeDefData
	if a.Base == "" && legacy.AttributeBase != "" {
		a.Base = ExprToScript(legacy.AttributeBase)
	}
	for _, threshold := range a.Thresholds {
		threshold.Value = strings.ReplaceAll(threshold.Value, "$self", "$"+a.DefID)
	}
	return nil
}

// Clone a copy of this.
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
		a.Thresholds = nil
	}
	return &clone
}

// ID returns the ID.
func (a *AttributeDef) ID() string {
	return a.DefID
}

// SetID sets the ID, sanitizing it in the process (i.e. it may be changed from what you set -- read it back if you want
// to be sure of what it gets set to.
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

// Kind returns the kind of attribute this is.
func (a *AttributeDef) Kind() int {
	switch {
	case a.Pool():
		return PoolAttrKind
	case a.Primary():
		return PrimaryAttrKind
	case a.Secondary():
		return SecondaryAttrKind
	default:
		return -1
	}
}

// Relevant returns true if the attribute is relevant to the given kind.
func (a *AttributeDef) Relevant(kind int) bool {
	if a.Placement == attribute.Hidden {
		return false
	}
	return a.Kind() == kind
}

// Primary returns true if the base value is a non-derived value.
func (a *AttributeDef) Primary() bool {
	if a.Type == attribute.PrimarySeparator {
		return true
	}
	if a.Type == attribute.Pool || a.Type == attribute.PoolRef || a.Placement == attribute.Secondary || a.IsSeparator() {
		return false
	}
	if a.Placement == attribute.Primary {
		return true
	}
	_, err := fxp.FromString(strings.TrimSpace(a.Base))
	return err == nil
}

// Secondary returns true if the base value is a derived value.
func (a *AttributeDef) Secondary() bool {
	if a.Type == attribute.SecondarySeparator {
		return true
	}
	if a.Type == attribute.Pool || a.Type == attribute.PoolRef || a.Placement == attribute.Primary || a.IsSeparator() {
		return false
	}
	if a.Placement == attribute.Secondary {
		return true
	}
	_, err := fxp.FromString(strings.TrimSpace(a.Base))
	return err != nil
}

// Pool returns true if the base value is a pool value.
func (a *AttributeDef) Pool() bool {
	return a.Type == attribute.PoolSeparator || a.Type == attribute.Pool || a.Type == attribute.PoolRef
}

// AllowsDecimal returns true if the value can have a decimal point in it.
func (a *AttributeDef) AllowsDecimal() bool {
	return a.Type == attribute.Decimal || a.Type == attribute.DecimalRef
}

// BaseValue returns the resolved base value.
func (a *AttributeDef) BaseValue(entity *Entity) fxp.Int {
	if a.IsSeparator() {
		return 0
	}
	return ResolveToNumber(entity, ScriptSelfProvider{}, a.Base)
}

// ComputeCost returns the value adjusted for a cost reduction.
func (a *AttributeDef) ComputeCost(entity *Entity, value, costReduction fxp.Int, sizeModifier int) fxp.Int {
	if a.IsSeparator() {
		return 0
	}
	cost := value.Mul(a.CostPerPoint)
	if sizeModifier > 0 && a.CostAdjPercentPerSM > 0 &&
		(a.DefID != "hp" || entity.SheetSettings.DamageProgression != progression.KnowingYourOwnStrength) {
		costReduction += fxp.From(sizeModifier).Mul(a.CostAdjPercentPerSM)
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
	hashhelper.String(h, a.DefID)
	hashhelper.Num8(h, a.Type)
	hashhelper.Num8(h, a.Placement)
	hashhelper.String(h, a.Name)
	hashhelper.String(h, a.FullName)
	hashhelper.String(h, a.Base)
	hashhelper.Num64(h, a.CostPerPoint)
	hashhelper.Num64(h, a.CostAdjPercentPerSM)
	hashhelper.Num64(h, len(a.Thresholds))
	for _, one := range a.Thresholds {
		one.Hash(h)
	}
}

// IsOpen returns true if this attribute is a separator and it is open.
func (a *AttributeDef) IsOpen(entity *Entity, sepCount int) bool {
	return a.IsSeparator() && !IsClosed(a.openKey(entity, a.Kind(), sepCount))
}

// SetOpen sets the open state of this attribute. Does nothing if this attribute is not a separator.
func (a *AttributeDef) SetOpen(entity *Entity, sepCount int, open bool) {
	if a.IsSeparator() {
		SetClosedState(a.openKey(entity, a.Kind(), sepCount), !open)
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
