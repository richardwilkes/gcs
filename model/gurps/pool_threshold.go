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
	"hash"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/skillsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/zeebo/xxh3"
)

var _ Hashable = &PoolThreshold{}

// ThresholdAttributeFlags has AttributeChoicesFor give, for the definitions a pool threshold belongs to, the attributes
// the threshold's attribute bonuses alter: those defined other than the pools, along with Move, Dodge, Parry and Block,
// but not the size modifier.
const ThresholdAttributeFlags = DodgeFlag | ParryFlag | BlockFlag | MoveFlag | NoPoolsFlag

// PoolThreshold holds a point within an attribute pool where changes in state occur.
type PoolThreshold struct {
	PoolThresholdData
	KeyPrefix string
}

// PoolThresholdData holds the data that will be serialized for the PoolThreshold.
type PoolThresholdData struct {
	// TID identifies the threshold, as the carrier of its grants, to the items they put on the sheet; see GrantedBy.
	TID         tid.TID  `json:"id"`
	State       string   `json:"state"`
	Value       string   `json:"value"`
	Explanation string   `json:"explanation,omitzero"`
	Features    Features `json:"features,omitempty"`
}

// NewPoolThreshold returns an empty threshold with an ID of its own.
func NewPoolThreshold() *PoolThreshold {
	return &PoolThreshold{TID: tid.MustNewTID(kinds.PoolThreshold)}
}

// legacyThresholdOps lists the keys of the operations a threshold held before it held features, each with the
// attribute it halved and, for ST, the limitation it halved it under. The lifting-only multiplier that halve_st
// becomes reaches everything that reads the lifting ST, not only the Basic Lift the flag halved; that is intended,
// see LiftingStrength.
var legacyThresholdOps = []struct {
	key        string
	attrID     string
	limitation stlimit.Option
}{
	{key: "halve_move", attrID: MoveID},
	{key: "halve_dodge", attrID: DodgeID},
	{key: "halve_st", attrID: StrengthID, limitation: stlimit.LiftingOnly},
}

// featuresForLegacyOps returns the features that do what the given operations did: halve Move, Dodge or, for lifting
// only, ST. Keys are matched ignoring case, as they were, and unknown ones are dropped. By design, the multipliers
// stack, one for each pool at such a threshold, whereas older versions of GCS, which read the operations written back
// for them (see MarshalJSONTo), halve Move and Dodge at most twice and ST at most once, so a sheet with more pools than
// that at such thresholds gives different values there.
func featuresForLegacyOps(ops []string) Features {
	var list Features
	for _, op := range legacyThresholdOps {
		if !slices.ContainsFunc(ops, func(one string) bool { return strings.EqualFold(one, op.key) }) {
			continue
		}
		bonus := NewAttributeBonus(op.attrID)
		bonus.Limitation = op.limitation
		bonus.Multiply = true
		bonus.Amount = fxp.Half
		list = append(list, bonus)
	}
	return list
}

// MarshalJSONTo implements json.MarshalerTo.
//
// TODO: Stop writing the legacy operations once older versions of GCS no longer need to load files saved by this one.
//
// The operations a threshold held before it held features are written alongside the features while those are exactly
// what the operations convert to, in any order (see legacyOps), so that older versions of GCS, which read only the
// operations, load a file saved by this one without losing what its thresholds do. Once the features are changed by
// anything other than that conversion or a reordering of it, the operations no longer describe them and are left out.
func (p *PoolThreshold) MarshalJSONTo(enc *jsontext.Encoder) error {
	if ops := p.legacyOps(); len(ops) != 0 {
		return json.MarshalEncode(enc, &struct {
			PoolThresholdData
			Ops []string `json:"ops"`
		}{PoolThresholdData: p.PoolThresholdData, Ops: ops})
	}
	return json.MarshalEncode(enc, &p.PoolThresholdData)
}

// legacyOps returns the keys of the operations the threshold's features are the conversion of (see
// featuresForLegacyOps), in any order, since the order of those features changes nothing they do, or nil when they
// aren't exactly that.
func (p *PoolThreshold) legacyOps() []string {
	if len(p.Features) == 0 || len(p.Features) > len(legacyThresholdOps) {
		return nil
	}
	var ops []string
	for _, op := range legacyThresholdOps {
		if slices.ContainsFunc(p.Features, func(f Feature) bool {
			bonus, ok := f.(*AttributeBonus)
			return ok && bonus.Attribute == op.attrID && bonus.ActualLimitation() == op.limitation
		}) {
			ops = append(ops, op.key)
		}
	}
	if len(ops) == 0 || !slices.Equal(featureHashes(featuresForLegacyOps(ops)), featureHashes(p.Features)) {
		return nil
	}
	return ops
}

// featureHashes returns the hash of each of the features, sorted, so that the same features in any order give the
// same hashes.
func featureHashes(list Features) []uint64 {
	hashes := make([]uint64, 0, len(list))
	for _, f := range list {
		h := xxh3.New()
		f.Hash(h)
		hashes = append(hashes, h.Sum64())
	}
	slices.Sort(hashes)
	return hashes
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom. Data written before thresholds held features has its operations
// converted to the features that do the same (see featuresForLegacyOps), and a threshold written before they had IDs
// is given one.
func (p *PoolThreshold) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var legacy struct {
		PoolThresholdData
		Expression string   `json:"expression"`
		Multiplier fxp.Int  `json:"multiplier"`
		Divisor    fxp.Int  `json:"divisor"`
		Addition   fxp.Int  `json:"addition"`
		Ops        []string `json:"ops"`
	}
	if err := json.UnmarshalDecode(dec, &legacy); err != nil {
		return err
	}
	if legacy.Value == "" {
		if legacy.Expression == "" {
			legacy.Expression = convertToExpression(legacy.Multiplier, legacy.Divisor, legacy.Addition)
		}
		if legacy.Expression != "" {
			legacy.Value = ExprToScript(legacy.Expression)
		}
	}
	if len(legacy.Features) == 0 {
		legacy.Features = featuresForLegacyOps(legacy.Ops)
	}
	if !tid.IsKindAndValid(legacy.TID, kinds.PoolThreshold) {
		legacy.TID = tid.MustNewTID(kinds.PoolThreshold)
	}
	p.PoolThresholdData = legacy.PoolThresholdData
	return nil
}

func convertToExpression(multiplier, divisor, addition fxp.Int) string {
	const self = "$self"
	const minusSelf = "-$self"
	if multiplier == 0 {
		return addition.String()
	}
	if multiplier == -fxp.One && (divisor == fxp.One || divisor == 0) {
		if addition != 0 {
			return minusSelf + addition.StringWithSign()
		}
		return minusSelf
	}
	if multiplier == fxp.One && (divisor == fxp.One || divisor == 0) {
		if addition != 0 {
			return self + addition.StringWithSign()
		}
		return self
	}
	ex := "round("
	switch multiplier {
	case fxp.One:
		ex += self
	case -fxp.One:
		ex += minusSelf
	default:
		if multiplier < 0 {
			ex += "-$self*" + (-multiplier).String()
		} else {
			ex += "$self*" + multiplier.String()
		}
	}
	if divisor != fxp.One && divisor != 0 {
		ex += "/" + divisor.String()
	}
	if addition != 0 {
		ex += addition.StringWithSign()
	}
	return ex + ")"
}

// Clone returns a copy of this, with copies of its features, and the same ID, so that the items the grants of the copy
// put on a sheet are those of the original.
func (p *PoolThreshold) Clone() *PoolThreshold {
	clone := *p
	clone.Features = p.Features.Clone()
	return &clone
}

// Threshold returns the threshold value for the given attribute.
func (p *PoolThreshold) Threshold(attr *Attribute) fxp.Int {
	return ResolveToNumber(attr.Entity, deferredNewScriptAttribute(attr), p.Value)
}

// ResolveExplanation returns the explanation with any embedded scripts resolved.
func (p *PoolThreshold) ResolveExplanation(attr *Attribute) string {
	return ResolveText(attr.Entity, deferredNewScriptAttribute(attr), p.Explanation)
}

// Hash writes this object's contents into the hasher. The ID is left out, as it tells thresholds apart rather than
// describing one.
func (p *PoolThreshold) Hash(h hash.Hash) {
	xhash.StringWithLen(h, p.State)
	xhash.StringWithLen(h, p.Value)
	xhash.StringWithLen(h, p.Explanation)
	hashList(h, p.Features)
}

// ThresholdIgnores returns true if the feature does nothing as one of the features of a pool threshold belonging to the
// definitions: one of a type a threshold may not hold; an amount per level, since a threshold has no level; an
// attribute bonus to an attribute ThresholdAttributeFlags doesn't give, such as a pool or the size modifier; or a
// bonus to this weapon or this armor, since a threshold is no item. A threshold has no switch either, but applies a
// switchable feature as it does any other.
func ThresholdIgnores(defs *AttributeDefs, f Feature) bool {
	if !f.FeatureType().AllowedInThreshold() {
		return true
	}
	switch one := f.(type) {
	case *AttributeBonus:
		// A factor isn't given per level, whatever a file says.
		if one.PerLevel && !one.Multiply {
			return true
		}
		// With no current key, only the choice standing in for an unrecognized one has none.
		choices, _ := AttributeChoicesFor(defs, "", ThresholdAttributeFlags, "")
		return !slices.ContainsFunc(choices, func(c *AttributeChoice) bool {
			return c.Key != "" && c.Key == one.Attribute
		})
	case *ConditionalModifierBonus:
		return one.PerLevel
	case *ReactionBonus:
		return one.PerLevel
	case *DRBonus:
		return one.PerLevel || len(one.Locations) == 0
	case *SkillBonus:
		return one.PerLevel || one.SelectionType == skillsel.ThisWeapon
	case *SkillPointBonus:
		return one.PerLevel
	case *SpellBonus:
		return one.PerLevel
	case *SpellPointBonus:
		return one.PerLevel
	case *TraitBonus:
		return one.PerLevel
	case *WeaponBonus:
		return one.PerLevel || one.SelectionType == wsel.ThisWeapon
	default:
		return false
	}
}

// DescribeThresholdFeature returns the description of one of a pool threshold's features, as Feature.Describe gives it,
// except that it never says the feature takes effect only while switched on: a threshold has no switch, and applies a
// switchable feature as it does any other.
func DescribeThresholdFeature(f Feature, entity *Entity, replacements map[string]string, em func(string) string) string {
	if f.IsSwitchable() {
		f = f.Clone()
		f.SetSwitchable(false)
	}
	return f.Describe(entity, replacements, em)
}

// String implements fmt.Stringer, naming the threshold by its state, as tooltips name it as the source of its features.
func (p *PoolThreshold) String() string {
	return p.State
}
