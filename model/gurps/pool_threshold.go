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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/zeebo/xxh3"
)

var _ Hashable = &PoolThreshold{}

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
// attribute it halved and, for ST, the limitation it halved it under.
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
// only, ST. Keys are matched ignoring case, as they were, and unknown ones are dropped.
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
// what the operations convert to (see legacyOps), so that older versions of GCS, which read only the operations, load
// a file saved by this one without losing what its thresholds do. Once the features are changed by anything other
// than that conversion, the operations no longer describe them and are left out.
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
// featuresForLegacyOps), or nil when they aren't exactly that.
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
	if len(ops) == 0 || featuresHash(featuresForLegacyOps(ops)) != featuresHash(p.Features) {
		return nil
	}
	return ops
}

// featuresHash returns a hash of the features.
func featuresHash(list Features) uint64 {
	h := xxh3.New()
	hashList(h, list)
	return h.Sum64()
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

// String implements fmt.Stringer, naming the threshold by its state, as tooltips name it as the source of its features.
func (p *PoolThreshold) String() string {
	return p.State
}
