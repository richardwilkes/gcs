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

package gurps

import (
	"fmt"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/eval"
	"github.com/richardwilkes/toolbox/txt"
)

// InstallEvaluatorFunctions installs additional functions for the evaluator.
func InstallEvaluatorFunctions(m map[string]eval.Function) {
	m["advantage_level"] = evalTraitLevel // For older files
	m["dice"] = evalDice
	m["enc"] = evalEncumbrance
	m["roll"] = evalRoll
	m["signed"] = evalSigned
	m["skill_level"] = evalSkillLevel
	m["ssrt"] = evalSSRT
	m["ssrt_to_yards"] = evalSSRTYards
	m["trait_level"] = evalTraitLevel
}

func evalToBool(e *eval.Evaluator, arguments string) (bool, error) {
	evaluated, err := e.EvaluateNew(arguments)
	if err != nil {
		return false, err
	}
	switch a := evaluated.(type) {
	case bool:
		return a, nil
	case fxp.Int:
		return a != 0, nil
	case string:
		return txt.IsTruthy(a), nil
	default:
		return false, nil
	}
}

func evalToNumber(e *eval.Evaluator, arguments string) (fxp.Int, error) {
	evaluated, err := e.EvaluateNew(arguments)
	if err != nil {
		return 0, err
	}
	return eval.FixedFrom[fxp.DP](evaluated)
}

func evalToString(e *eval.Evaluator, arguments string) (string, error) {
	v, err := e.EvaluateNew(arguments)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", v), nil
}

// evalEncumbrance takes up to 2 arguments: forSkills (bool, required) and returnFactor (bool, optional)
func evalEncumbrance(e *eval.Evaluator, arguments string) (any, error) {
	arg, remaining := eval.NextArg(arguments)
	forSkills, err := evalToBool(e, arg)
	if err != nil {
		return nil, err
	}
	var returnFactor bool
	arg, _ = eval.NextArg(remaining)
	if arg = strings.TrimSpace(arg); arg != "" {
		if returnFactor, err = evalToBool(e, remaining); err != nil {
			return nil, err
		}
	}
	entity, ok := e.Resolver.(*Entity)
	if !ok || entity.Type != PC {
		return fxp.Int(0), nil
	}
	level := fxp.From(int(entity.EncumbranceLevel(forSkills)))
	if returnFactor {
		return fxp.One - level.Mul(fxp.Two).Div(fxp.Ten), nil
	}
	return level, nil
}

// evalSkillLevel takes up to 3 arguments: name (string, required), specialization (string, optional), relative (bool, optional)
func evalSkillLevel(e *eval.Evaluator, arguments string) (any, error) {
	entity, ok := e.Resolver.(*Entity)
	if !ok || entity.Type != PC {
		return fxp.Int(0), nil
	}
	name, remaining := eval.NextArg(arguments)
	var err error
	if name, err = evalToString(e, name); err != nil {
		return fxp.Int(0), err
	}
	name = strings.Trim(name, `"`)
	var specialization string
	specialization, remaining = eval.NextArg(remaining)
	if specialization = strings.TrimSpace(specialization); specialization != "" {
		if specialization, err = evalToString(e, specialization); err != nil {
			return fxp.Int(0), err
		}
		specialization = strings.Trim(specialization, `"`)
	}
	var relative bool
	arg, _ := eval.NextArg(remaining)
	if arg = strings.TrimSpace(arg); arg != "" {
		if relative, err = evalToBool(e, arg); err != nil {
			return fxp.Int(0), err
		}
	}
	if entity.isSkillLevelResolutionExcluded(name, specialization) {
		return fxp.Int(0), nil
	}
	entity.registerSkillLevelResolutionExclusion(name, specialization)
	defer entity.unregisterSkillLevelResolutionExclusion(name, specialization)
	var level fxp.Int
	Traverse(func(s *Skill) bool {
		if strings.EqualFold(s.Name, name) && strings.EqualFold(s.Specialization, specialization) {
			s.UpdateLevel()
			if relative {
				level = s.LevelData.RelativeLevel
			} else {
				level = s.LevelData.Level
			}
			return true
		}
		return false
	}, true, true, entity.Skills...)
	return level, nil
}

func evalTraitLevel(e *eval.Evaluator, arguments string) (any, error) {
	entity, ok := e.Resolver.(*Entity)
	if !ok || entity.Type != PC {
		return -fxp.One, nil
	}
	arguments = strings.Trim(arguments, `"`)
	levels := -fxp.One
	Traverse(func(t *Trait) bool {
		if strings.EqualFold(t.Name, arguments) {
			if t.IsLeveled() {
				if levels == -fxp.One {
					levels = t.Levels
				} else {
					levels += t.Levels
				}
			}
		}
		return false
	}, true, false, entity.Traits...)
	return levels, nil
}

func evalDice(e *eval.Evaluator, arguments string) (any, error) {
	var argList []int
	for arguments != "" {
		var arg string
		arg, arguments = eval.NextArg(arguments)
		n, err := evalToNumber(e, arg)
		if err != nil {
			return nil, err
		}
		argList = append(argList, fxp.As[int](n))
	}
	var d *dice.Dice
	switch len(argList) {
	case 1:
		d = &dice.Dice{
			Count:      1,
			Sides:      argList[0],
			Multiplier: 1,
		}
	case 2:
		d = &dice.Dice{
			Count:      argList[0],
			Sides:      argList[1],
			Multiplier: 1,
		}
	case 3:
		d = &dice.Dice{
			Count:      argList[0],
			Sides:      argList[1],
			Modifier:   argList[2],
			Multiplier: 1,
		}
	case 4:
		d = &dice.Dice{
			Count:      argList[0],
			Sides:      argList[1],
			Modifier:   argList[2],
			Multiplier: argList[3],
		}
	default:
		return nil, errs.New("invalid dice specification")
	}
	return d.String(), nil
}

func evalRoll(e *eval.Evaluator, arguments string) (any, error) {
	if strings.IndexByte(arguments, '(') != -1 {
		var err error
		if arguments, err = evalToString(e, arguments); err != nil {
			return nil, err
		}
	}
	return fxp.From(dice.New(arguments).Roll(false)), nil
}

func evalSigned(e *eval.Evaluator, arguments string) (any, error) {
	n, err := evalToNumber(e, arguments)
	if err != nil {
		return nil, err
	}
	return n.StringWithSign(), nil
}

func evalSSRT(e *eval.Evaluator, arguments string) (any, error) {
	// Takes 3 args: length (number), units (string), flag (bool) indicating for size (true) or speed/range (false)
	var arg string
	arg, arguments = eval.NextArg(arguments)
	n, err := evalToString(e, arg)
	if err != nil {
		return nil, err
	}
	arg, arguments = eval.NextArg(arguments)
	var units string
	if units, err = evalToString(e, arg); err != nil {
		return nil, err
	}
	arg, _ = eval.NextArg(arguments)
	var wantSize bool
	if wantSize, err = evalToBool(e, arg); err != nil {
		return nil, err
	}
	var length Length
	if length, err = LengthFromString(n+" "+units, Yard); err != nil {
		return nil, err
	}
	result := yardsToValue(length, wantSize)
	if !wantSize {
		result = -result
	}
	return fxp.From(result), nil
}

func evalSSRTYards(e *eval.Evaluator, arguments string) (any, error) {
	v, err := evalToNumber(e, arguments)
	if err != nil {
		return nil, err
	}
	return valueToYards(fxp.As[int](v)), nil
}

func yardsToValue(length Length, allowNegative bool) int {
	inches := fxp.Int(length)
	yards := inches.Div(fxp.ThirtySix)
	if allowNegative {
		switch {
		case inches <= fxp.One.Div(fxp.Five):
			return -15
		case inches <= fxp.One.Div(fxp.Three):
			return -14
		case inches <= fxp.Half:
			return -13
		case inches <= fxp.Two.Div(fxp.Three):
			return -12
		case inches <= fxp.One:
			return -11
		case inches <= fxp.OneAndAHalf:
			return -10
		case inches <= fxp.Two:
			return -9
		case inches <= fxp.Three:
			return -8
		case inches <= fxp.Five:
			return -7
		case inches <= fxp.Eight:
			return -6
		}
		feet := inches.Div(fxp.Twelve)
		switch {
		case feet <= fxp.One:
			return -5
		case feet <= fxp.OneAndAHalf:
			return -4
		case feet <= fxp.Two:
			return -3
		case yards <= fxp.One:
			return -2
		case yards <= fxp.OneAndAHalf:
			return -1
		}
	}
	if yards <= fxp.Two {
		return 0
	}
	amt := 0
	for yards > fxp.Ten {
		yards = yards.Div(fxp.Ten)
		amt += 6
	}
	switch {
	case yards > fxp.Seven:
		return amt + 4
	case yards > fxp.Five:
		return amt + 3
	case yards > fxp.Three:
		return amt + 2
	case yards > fxp.Two:
		return amt + 1
	case yards > fxp.OneAndAHalf:
		return amt
	default:
		return amt - 1
	}
}

func valueToYards(value int) fxp.Int {
	if value < -15 {
		value = -15
	}
	switch value {
	case -15:
		return fxp.One.Div(fxp.Five).Div(fxp.ThirtySix)
	case -14:
		return fxp.One.Div(fxp.Three).Div(fxp.ThirtySix)
	case -13:
		return fxp.Half.Div(fxp.ThirtySix)
	case -12:
		return fxp.Two.Div(fxp.Three).Div(fxp.ThirtySix)
	case -11:
		return fxp.One.Div(fxp.ThirtySix)
	case -10:
		return fxp.OneAndAHalf.Div(fxp.ThirtySix)
	case -9:
		return fxp.Two.Div(fxp.ThirtySix)
	case -8:
		return fxp.Three.Div(fxp.ThirtySix)
	case -7:
		return fxp.Five.Div(fxp.ThirtySix)
	case -6:
		return fxp.Eight.Div(fxp.ThirtySix)
	case -5:
		return fxp.One.Div(fxp.Three)
	case -4:
		return fxp.OneAndAHalf.Div(fxp.Three)
	case -3:
		return fxp.Two.Div(fxp.Three)
	case -2:
		return fxp.One
	case -1:
		return fxp.OneAndAHalf
	case 0:
		return fxp.Two
	case 1:
		return fxp.Three
	case 2:
		return fxp.Five
	case 3:
		return fxp.Seven
	}
	value -= 4
	multiplier := fxp.One
	for i := 0; i < value/6; i++ {
		multiplier = multiplier.Mul(fxp.Ten)
	}
	var v fxp.Int
	switch value % 6 {
	case 0:
		v = fxp.Ten
	case 1:
		v = fxp.Fifteen
	case 2:
		v = fxp.Twenty
	case 3:
		v = fxp.Thirty
	case 4:
		v = fxp.Fifty
	case 5:
		v = fxp.Seventy
	}
	return v.Mul(multiplier)
}
