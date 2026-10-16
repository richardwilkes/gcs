// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package condition

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/i18n"
)

// operandValue is a resolved operand. Its kind is ValueNumber or ValueWeight, or ValueAbsent for a sum or maximum over
// no members, which takes the kind of whatever it is combined with.
type operandValue struct {
	value fxp.Int
	kind  ValueKind
}

// contribution is a member an aggregate operand drew on, kept in Collect mode for hits.
type contribution struct {
	item  Item
	where Result
}

// operandEval resolves the operands of one compare.
type operandEval struct {
	e       *evaluator
	contrib []contribution
	collect bool
}

// resolve resolves an operand, then applies its scale and offset. A non-nil result is the Failed result to
// report instead.
func (oe *operandEval) resolve(o *Operand, f frame) (operandValue, *Result) {
	var v operandValue
	var fail *Result
	switch o.Source() {
	case SourceValue:
		v = operandValue{value: *o.Value, kind: ValueNumber}
	case SourceWeight:
		v = operandValue{value: fxp.Int(*o.Weight), kind: ValueWeight}
	case SourceAttribute:
		v, fail = oe.number(o.Attribute, oe.e.host.Attribute, i18n.Text("unknown attribute %q"))
	case SourceEntity:
		v, fail = oe.number(o.Entity, oe.e.host.Entity, i18n.Text("unknown entity value %q"))
	case SourceSelf:
		v, fail = oe.self(o.Self)
	case SourceAggregate:
		v, fail = oe.aggregate(o, f)
	default:
		v, fail = oe.sum(o, f)
	}
	if fail != nil {
		return v, fail
	}
	if o.Scale != nil {
		v.value = v.value.Mul(*o.Scale)
	}
	v.value += o.Offset
	return v, nil
}

func (oe *operandEval) number(key string, lookup func(string) (fxp.Int, bool), format string) (operandValue, *Result) {
	value, ok := lookup(key)
	if !ok {
		r := failure(fmt.Sprintf(format, key))
		return operandValue{}, &r
	}
	return operandValue{value: value, kind: ValueNumber}, nil
}

// self resolves a Number or Weight field of self. Anything else is unresolved.
func (oe *operandEval) self(key string) (operandValue, *Result) {
	v, ok := oe.e.host.Field(oe.e.self, key)
	if !ok || (v.Kind != ValueNumber && v.Kind != ValueWeight) {
		r := failure(fmt.Sprintf(i18n.Text("self has no number or weight field %q"), key))
		return operandValue{}, &r
	}
	return operandValue{value: v.Number, kind: v.Kind}, nil
}

// aggregate counts the matching members, or sums or finds the maximum of their Field, skipping members whose field is
// Absent. A field the members' kind lacks, or that is neither a Number nor a Weight, fails, as do members whose fields
// mix the two. The maximum of nothing is 0.
func (oe *operandEval) aggregate(o *Operand, f frame) (operandValue, *Result) {
	var v operandValue
	var extreme []contribution
	var bad *Result
	fail := oe.e.eachMember(&o.Members, f, func(item Item, where *Result) bool {
		if o.Aggregate == AggregateCount {
			v.value += fxp.One
			oe.add(item, where)
			return true
		}
		field, ok := oe.e.host.Field(item, o.Field)
		switch {
		case ok && field.Kind == ValueAbsent:
			return true
		case !ok || (field.Kind != ValueNumber && field.Kind != ValueWeight) ||
			(v.kind != ValueAbsent && field.Kind != v.kind):
			r := failure(fmt.Sprintf(i18n.Text("members have no number or weight field %q"), o.Field))
			bad = &r
			return false
		case o.Aggregate == AggregateSum:
			v.value += field.Number
			oe.add(item, where)
		case v.kind == ValueAbsent || field.Number > v.value:
			v.value = field.Number
			extreme = oe.keep(extreme[:0], item, where)
		case field.Number == v.value:
			extreme = oe.keep(extreme, item, where)
		default:
		}
		v.kind = field.Kind
		return true
	})
	switch {
	case fail != nil:
		return v, fail
	case bad != nil:
		return v, bad
	case o.Aggregate == AggregateCount:
		v.kind = ValueNumber
	default:
	}
	oe.contrib = append(oe.contrib, extreme...)
	return v, nil
}

func (oe *operandEval) add(item Item, where *Result) {
	if oe.collect {
		oe.contrib = append(oe.contrib, contribution{item: item, where: *where})
	}
}

func (oe *operandEval) keep(list []contribution, item Item, where *Result) []contribution {
	if oe.collect {
		list = append(list, contribution{item: item, where: *where})
	}
	return list
}

// sum adds the operands of a sum. Numbers and Weights may not mix.
func (oe *operandEval) sum(o *Operand, f frame) (operandValue, *Result) {
	var total operandValue
	for i := range o.Sum {
		v, fail := oe.resolve(&o.Sum[i], f)
		if fail != nil {
			return total, fail
		}
		if total.kind, fail = combineKinds(total.kind, v.kind); fail != nil {
			return total, fail
		}
		total.value += v.value
	}
	return total, nil
}

// combineKinds returns the kind of two values used together, failing when a Number meets a Weight.
func combineKinds(a, b ValueKind) (ValueKind, *Result) {
	switch {
	case a == ValueAbsent:
		return b, nil
	case b == ValueAbsent || a == b:
		return a, nil
	default:
		r := failure(i18n.Text("a number is used with a weight"))
		return a, &r
	}
}

// Compare returns true when left and right satisfy the operator. An operator outside the enum compares as OpIs.
func (enum CompareOp) Compare(left, right fxp.Int) bool {
	switch enum {
	case OpIsNot:
		return left != right
	case OpAtLeast:
		return left >= right
	case OpAtMost:
		return left <= right
	case OpMoreThan:
		return left > right
	case OpLessThan:
		return left < right
	default:
		return left == right
	}
}
