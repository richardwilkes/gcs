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
	"hash"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/threshold"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
)

var _ Hashable = &PoolThreshold{}

// PoolThreshold holds a point within an attribute pool where changes in state occur.
type PoolThreshold struct {
	PoolThresholdData
	KeyPrefix string
}

// PoolThresholdData holds the data that will be serialized for the PoolThreshold.
type PoolThresholdData struct {
	State       string         `json:"state"`
	Value       string         `json:"value"`
	Explanation string         `json:"explanation,omitempty"`
	Ops         []threshold.Op `json:"ops,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (p *PoolThreshold) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	e := json.NewEncoder(&buffer)
	e.SetEscapeHTML(false)
	err := e.Encode(&p.PoolThresholdData)
	return buffer.Bytes(), err
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *PoolThreshold) UnmarshalJSON(data []byte) error {
	var legacy struct {
		PoolThresholdData
		Expression string  `json:"expression"`
		Multiplier fxp.Int `json:"multiplier"`
		Divisor    fxp.Int `json:"divisor"`
		Addition   fxp.Int `json:"addition"`
	}
	if err := json.Unmarshal(data, &legacy); err != nil {
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
	slices.Sort(legacy.Ops)
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

// Clone a copy of this.
func (p *PoolThreshold) Clone() *PoolThreshold {
	clone := *p
	if p.Ops != nil {
		clone.Ops = make([]threshold.Op, len(p.Ops))
		copy(clone.Ops, p.Ops)
	}
	return &clone
}

// Threshold returns the threshold value for the given maximum.
func (p *PoolThreshold) Threshold(entity *Entity) fxp.Int {
	return ResolveToNumber(entity, ScriptSelfProvider{}, p.Value)
}

// ContainsOp returns true if this PoolThreshold contains the specified ThresholdOp.
func (p *PoolThreshold) ContainsOp(op threshold.Op) bool {
	return slices.Contains(p.Ops, op)
}

// AddOp adds the specified ThresholdOp.
func (p *PoolThreshold) AddOp(op threshold.Op) {
	if !slices.Contains(p.Ops, op) {
		p.Ops = append(p.Ops, op)
		slices.Sort(p.Ops)
	}
}

// RemoveOp removes the specified ThresholdOp.
func (p *PoolThreshold) RemoveOp(op threshold.Op) {
	if i := slices.Index(p.Ops, op); i != -1 {
		p.Ops = slices.Delete(p.Ops, i, i+1)
	}
}

// Hash writes this object's contents into the hasher.
func (p *PoolThreshold) Hash(h hash.Hash) {
	hashhelper.String(h, p.State)
	hashhelper.String(h, p.Value)
	hashhelper.String(h, p.Explanation)
	hashhelper.Num64(h, len(p.Ops))
	for _, one := range p.Ops {
		hashhelper.Num8(h, one)
	}
}

func (p *PoolThreshold) String() string {
	return p.State
}
