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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
)

// criterionKind identifies which criterion of a check is set.
type criterionKind byte

const (
	criterionNone criterionKind = iota
	criterionText
	criterionNumber
	criterionWeight
)

func (c *Check) criterion() criterionKind {
	switch {
	case !c.Text.IsZero():
		return criterionText
	case !c.Number.IsZero():
		return criterionNumber
	case !c.Weight.IsZero():
		return criterionWeight
	default:
		return criterionNone
	}
}

// negative returns true when the check's criterion is a negative comparison, such as "is not", which holds exactly
// when its positive form doesn't, so it holds for a subject that lacks the field.
func (c *Check) negative() bool {
	switch c.criterion() {
	case criterionText:
		return c.Text.Compare.IsNotType()
	case criterionNumber:
		return c.Number.Compare == criteria.NotEqualsNumber
	case criterionWeight:
		return c.Weight.Compare == criteria.NotEqualsNumber
	default:
		return false
	}
}

// decision is what a check decided before Not. When failed is set, fail is the outcome to report instead.
type decision struct {
	have           fxp.Int // the number or weight compared
	haveKind       ValueKind
	fail           outcome
	met            bool
	testsSomething bool // an "any" check scores 0 and gives no hit
	failed         bool
}

// decide evaluates the check up to its state before Not, filling in d, which the caller owns so that it is never
// copied. A field the subject's kind doesn't have fails the check, and so does a criterion that doesn't suit the
// field's kind of value. A subject that lacks the field meets only the negative criteria, which hold whenever their
// positive form doesn't, and so a check without a criterion asks whether the subject has the field.
func (c *Check) decide(e *evaluator, f frame, d *decision) {
	v, ok := e.checkField(c, f)
	if !ok {
		d.fail, d.failed = outcome{reason: fmt.Sprintf(i18n.Text("no field %q"), c.Field), state: Failed}, true
		return
	}
	d.testsSomething = true
	if v.Kind == ValueAbsent {
		d.met = c.negative()
		return
	}
	d.have, d.haveKind = v.Number, v.Kind
	switch c.criterion() {
	case criterionText:
		if v.Kind != ValueText && v.Kind != ValueList {
			d.fail, d.failed = c.mismatch(v), true
			return
		}
		var fail *outcome
		if d.met, fail = e.matchText(c, v, f); fail != nil {
			d.fail, d.failed = *fail, true
		}
	case criterionNumber:
		if v.Kind != ValueNumber {
			d.fail, d.failed = c.mismatch(v), true
			return
		}
		d.met = c.Number.Matches(v.Number)
	case criterionWeight:
		if v.Kind != ValueWeight {
			d.fail, d.failed = c.mismatch(v), true
			return
		}
		d.met = c.Weight.Matches(fxp.Weight(v.Number))
	default:
		d.met = v.Kind != ValueBool || v.Bool
		d.testsSomething = v.Kind == ValueBool
	}
}

// test is the YesNo evaluation of a check.
func (c *Check) test(e *evaluator, f frame) outcome {
	var d decision
	c.decide(e, f, &d)
	if d.failed {
		return d.fail
	}
	return outcome{state: metState(d.met, c.Not)}
}

// eval implements Node. In Collect mode a Met check that tests something scores 1 and, unless negated, is a hit.
func (c *Check) eval(e *evaluator, f frame) Result {
	var d decision
	c.decide(e, f, &d)
	if d.failed {
		return d.fail.result()
	}
	r := Result{State: metState(d.met, c.Not)}
	if d.haveKind == ValueNumber || d.haveKind == ValueWeight {
		switch c.criterion() {
		case criterionNumber:
			r.Have, r.HaveKind, r.Need = d.have, d.haveKind, c.Number.Qualifier
		case criterionWeight:
			r.Have, r.HaveKind, r.Need = d.have, d.haveKind, fxp.Int(c.Weight.Qualifier)
		default:
		}
	}
	if e.collect && r.State == Met && d.testsSomething {
		r.Specificity = 1
		if !c.Not {
			r.hits = []Hit{{Node: c, Item: f.subject}}
		}
	}
	return r
}

// mismatch returns the Failed outcome of a criterion that does not suit the kind of value the subject reported.
func (c *Check) mismatch(v Value) outcome {
	return outcome{
		reason: fmt.Sprintf(i18n.Text("the criterion does not suit the value of field %q (%s)"), c.Field,
			v.Kind.String()),
		state: Failed,
	}
}

// matchText matches a Text or List value against the check's text criterion, expanding any self tokens. A non-nil
// outcome is the Failed outcome to report instead.
func (e *evaluator) matchText(c *Check, v Value, f frame) (bool, *outcome) {
	if sq := e.selfQualifier(c); sq != nil {
		return e.matchTokens(c, sq, v, f)
	}
	var repl map[string]string
	if strings.IndexByte(c.Text.Qualifier, byte(nameable.MarkerDelimiter)) >= 0 {
		repl = e.replacementsFor(f)
	}
	if v.Kind == ValueList {
		return c.Text.MatchesList(repl, v.List...), nil
	}
	return c.Text.Matches(repl, v.Text), nil
}
