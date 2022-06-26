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

package fxp

import (
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/eval"
	"github.com/richardwilkes/toolbox/log/jot"
)

// The evaluator operators and functions that will be used when calling NewEvaluator().
var (
	EvalOperators = eval.FixedOperators[DP](true)
	EvalFuncs     = eval.FixedFunctions[DP]()
)

// NewEvaluator creates a new evaluator whose number type is an Int.
func NewEvaluator(resolver eval.VariableResolver) *eval.Evaluator {
	return &eval.Evaluator{
		Resolver:  resolver,
		Operators: EvalOperators,
		Functions: EvalFuncs,
	}
}

// EvaluateToNumber evaluates the provided expression and returns a number.
func EvaluateToNumber(expression string, resolver eval.VariableResolver) Int {
	result, err := NewEvaluator(resolver).Evaluate(expression)
	if err != nil {
		jot.Warn(errs.NewWithCausef(err, "unable to resolve '%s'", expression))
		return 0
	}
	if value, ok := result.(Int); ok {
		return value
	}
	if str, ok := result.(string); ok {
		var value Int
		if value, err = FromString(str); err == nil {
			return value
		}
	}
	jot.Warn(errs.Newf("unable to resolve '%s' to a number", expression))
	return 0
}
