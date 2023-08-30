/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/gcs/v5/model/dbg"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/eval"
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
		if dbg.VariableResolver {
			errs.Log(errs.NewWithCause("unable to resolve expression", err), "expression", expression)
		}
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
	if dbg.VariableResolver {
		errs.Log(errs.NewWithCause("unable to resolve expression to a number", err), "expression", expression)
	}
	return 0
}
