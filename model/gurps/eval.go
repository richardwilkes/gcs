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
	"fmt"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/eval"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xmath/rand"
)

// InstallEvaluatorFunctions installs additional functions for the evaluator.
func InstallEvaluatorFunctions(m map[string]eval.Function) {
	m["add_dice"] = evalAddDice
	m["advantage_level"] = evalTraitLevel // For older files
	m["dice"] = evalDice
	m["dice_count"] = evalDiceCount
	m["dice_modifier"] = evalDiceModifier
	m["dice_multiplier"] = evalDiceMultiplier
	m["dice_sides"] = evalDiceSides
	m["enc"] = evalEncumbrance
	m["has_trait"] = evalHasTrait
	m["random_height"] = evalRandomHeight
	m["random_weight"] = evalRandomWeight
	m["roll"] = evalRoll
	m["signed"] = evalSigned
	m["skill_level"] = evalSkillLevel
	m["ssrt"] = evalSSRT
	m["ssrt_to_yards"] = evalSSRTYards
	m["subtract_dice"] = evalSubtractDice
	m["trait_level"] = evalTraitLevel
	m["weapon_damage"] = evalWeaponDamage
}

func evalToBool(ev *eval.Evaluator, arguments string) (bool, error) {
	evaluated, err := ev.EvaluateNew(arguments)
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

func evalToNumber(ev *eval.Evaluator, arguments string) (fxp.Int, error) {
	evaluated, err := ev.EvaluateNew(arguments)
	if err != nil {
		return 0, err
	}
	return eval.FixedFrom[fxp.DP](evaluated)
}

func evalToString(ev *eval.Evaluator, arguments string) (string, error) {
	v, err := ev.EvaluateNew(arguments)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", v), nil
}

// evalEncumbrance takes up to 2 arguments: forSkills (bool, required) and returnFactor (bool, optional)
func evalEncumbrance(ev *eval.Evaluator, arguments string) (any, error) {
	arg, remaining := eval.NextArg(arguments)
	forSkills, err := evalToBool(ev, arg)
	if err != nil {
		return nil, err
	}
	var returnFactor bool
	arg, _ = eval.NextArg(remaining)
	if arg = strings.TrimSpace(arg); arg != "" {
		if returnFactor, err = evalToBool(ev, remaining); err != nil {
			return nil, err
		}
	}
	e, ok := ev.Resolver.(*Entity)
	if !ok {
		return fxp.Int(0), nil
	}
	level := fxp.From(int(e.EncumbranceLevel(forSkills)))
	if returnFactor {
		return fxp.One - level.Mul(fxp.Two).Div(fxp.Ten), nil
	}
	return level, nil
}

// evalSkillLevel takes up to 3 arguments: name (string, required), specialization (string, optional), relative (bool, optional)
func evalSkillLevel(ev *eval.Evaluator, arguments string) (any, error) {
	e, ok := ev.Resolver.(*Entity)
	if !ok {
		return fxp.Int(0), nil
	}
	name, remaining := eval.NextArg(arguments)
	var err error
	if name, err = evalToString(ev, name); err != nil {
		return fxp.Int(0), err
	}
	name = strings.Trim(name, `"`)
	var specialization string
	specialization, remaining = eval.NextArg(remaining)
	if specialization = strings.TrimSpace(specialization); specialization != "" {
		if specialization, err = evalToString(ev, specialization); err != nil {
			return fxp.Int(0), err
		}
		specialization = strings.Trim(specialization, `"`)
	}
	var relative bool
	arg, _ := eval.NextArg(remaining)
	if arg = strings.TrimSpace(arg); arg != "" {
		if relative, err = evalToBool(ev, arg); err != nil {
			return fxp.Int(0), err
		}
	}
	if e.isSkillLevelResolutionExcluded(name, specialization) {
		return fxp.Int(0), nil
	}
	e.registerSkillLevelResolutionExclusion(name, specialization)
	defer e.unregisterSkillLevelResolutionExclusion(name, specialization)
	var level fxp.Int
	Traverse(func(s *Skill) bool {
		if strings.EqualFold(s.NameWithReplacements(), name) &&
			strings.EqualFold(s.SpecializationWithReplacements(), specialization) {
			s.UpdateLevel()
			if relative {
				level = s.LevelData.RelativeLevel
			} else {
				level = s.LevelData.Level
			}
			return true
		}
		return false
	}, true, true, e.Skills...)
	return level, nil
}

func evalHasTrait(ev *eval.Evaluator, arguments string) (any, error) {
	e, ok := ev.Resolver.(*Entity)
	if !ok {
		return false, nil
	}
	arguments = strings.Trim(arguments, `"`)
	found := false
	Traverse(func(t *Trait) bool {
		if strings.EqualFold(t.NameWithReplacements(), arguments) {
			found = true
			return true
		}
		return false
	}, true, false, e.Traits...)
	return found, nil
}

func evalTraitLevel(ev *eval.Evaluator, arguments string) (any, error) {
	e, ok := ev.Resolver.(*Entity)
	if !ok {
		return -fxp.One, nil
	}
	arguments = strings.Trim(arguments, `"`)
	levels := -fxp.One
	Traverse(func(t *Trait) bool {
		if strings.EqualFold(t.NameWithReplacements(), arguments) {
			if t.IsLeveled() {
				if levels == -fxp.One {
					levels = t.Levels
				} else {
					levels += t.Levels
				}
			}
		}
		return false
	}, true, false, e.Traits...)
	return levels, nil
}

func evalWeaponDamage(ev *eval.Evaluator, arguments string) (any, error) {
	e, ok := ev.Resolver.(*Entity)
	if !ok {
		return "", nil
	}
	var arg string
	arg, arguments = eval.NextArg(arguments)
	name, err := evalToString(ev, strings.Trim(arg, `"`))
	if err != nil {
		return nil, err
	}
	arg, _ = eval.NextArg(arguments)
	var usage string
	if usage, err = evalToString(ev, strings.Trim(arg, `"`)); err != nil {
		return nil, err
	}
	for _, w := range e.Weapons(true, false) {
		if strings.EqualFold(w.String(), name) && strings.EqualFold(w.UsageWithReplacements(), usage) {
			return w.Damage.ResolvedDamage(nil), nil
		}
	}
	for _, w := range e.Weapons(false, false) {
		if strings.EqualFold(w.String(), name) && strings.EqualFold(w.UsageWithReplacements(), usage) {
			return w.Damage.ResolvedDamage(nil), nil
		}
	}
	return "", nil
}

func evalDice(ev *eval.Evaluator, arguments string) (any, error) {
	var argList []int
	for arguments != "" {
		var arg string
		arg, arguments = eval.NextArg(arguments)
		n, err := evalToNumber(ev, arg)
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

func evalAddDice(ev *eval.Evaluator, arguments string) (any, error) {
	var first string
	first, arguments = eval.NextArg(arguments)
	if strings.IndexByte(first, '(') != -1 {
		var err error
		if first, err = evalToString(ev, first); err != nil {
			return nil, err
		}
	}
	var second string
	second, _ = eval.NextArg(arguments)
	if strings.IndexByte(second, '(') != -1 {
		var err error
		if second, err = evalToString(ev, second); err != nil {
			return nil, err
		}
	}
	d1 := dice.New(first)
	d2 := dice.New(second)
	if d1.Sides != d2.Sides {
		return nil, errs.New("dice sides must match")
	}
	if d1.Multiplier > 0 {
		d1.Count *= d1.Multiplier
		d1.Modifier *= d1.Multiplier
		d1.Multiplier = 1
	}
	if d2.Multiplier > 0 {
		d2.Count *= d2.Multiplier
		d2.Modifier *= d2.Multiplier
		d2.Multiplier = 1
	}
	d1.Count += d2.Count
	d1.Modifier += d2.Modifier
	return d1.String(), nil
}

func evalSubtractDice(ev *eval.Evaluator, arguments string) (any, error) {
	var first string
	first, arguments = eval.NextArg(arguments)
	if strings.IndexByte(first, '(') != -1 {
		var err error
		if first, err = evalToString(ev, first); err != nil {
			return nil, err
		}
	}
	var second string
	second, _ = eval.NextArg(arguments)
	if strings.IndexByte(second, '(') != -1 {
		var err error
		if second, err = evalToString(ev, second); err != nil {
			return nil, err
		}
	}
	d1 := dice.New(first)
	d2 := dice.New(second)
	if d1.Sides != d2.Sides {
		return nil, errs.New("dice sides must match")
	}
	if d1.Multiplier > 0 {
		d1.Count *= d1.Multiplier
		d1.Modifier *= d1.Multiplier
		d1.Multiplier = 1
	}
	if d2.Multiplier > 0 {
		d2.Count *= d2.Multiplier
		d2.Modifier *= d2.Multiplier
		d2.Multiplier = 1
	}
	d1.Count -= d2.Count
	d1.Count = max(d1.Count, 0)
	d1.Modifier -= d2.Modifier
	return d1.String(), nil
}

func evalDiceCount(ev *eval.Evaluator, arguments string) (any, error) {
	d, err := convertToDice(ev, arguments)
	if err != nil {
		return nil, err
	}
	return d.Count, nil
}

func evalDiceSides(ev *eval.Evaluator, arguments string) (any, error) {
	d, err := convertToDice(ev, arguments)
	if err != nil {
		return nil, err
	}
	return d.Sides, nil
}

func evalDiceModifier(ev *eval.Evaluator, arguments string) (any, error) {
	d, err := convertToDice(ev, arguments)
	if err != nil {
		return nil, err
	}
	return d.Modifier, nil
}

func evalDiceMultiplier(ev *eval.Evaluator, arguments string) (any, error) {
	d, err := convertToDice(ev, arguments)
	if err != nil {
		return nil, err
	}
	return d.Multiplier, nil
}

func evalRoll(ev *eval.Evaluator, arguments string) (any, error) {
	d, err := convertToDice(ev, arguments)
	if err != nil {
		return nil, err
	}
	return fxp.From(d.Roll(false)), nil
}

func convertToDice(ev *eval.Evaluator, arguments string) (*dice.Dice, error) {
	if strings.IndexByte(arguments, '(') != -1 {
		var err error
		if arguments, err = evalToString(ev, arguments); err != nil {
			return nil, err
		}
	}
	return dice.New(arguments), nil
}

func evalSigned(ev *eval.Evaluator, arguments string) (any, error) {
	n, err := evalToNumber(ev, arguments)
	if err != nil {
		return nil, err
	}
	return n.StringWithSign(), nil
}

func evalSSRT(ev *eval.Evaluator, arguments string) (any, error) {
	// Takes 3 args: length (number), units (string), flag (bool) indicating for size (true) or speed/range (false)
	var arg string
	arg, arguments = eval.NextArg(arguments)
	n, err := evalToString(ev, arg)
	if err != nil {
		return nil, err
	}
	arg, arguments = eval.NextArg(arguments)
	var units string
	if units, err = evalToString(ev, arg); err != nil {
		return nil, err
	}
	arg, _ = eval.NextArg(arguments)
	var wantSize bool
	if wantSize, err = evalToBool(ev, arg); err != nil {
		return nil, err
	}
	var length fxp.Length
	if length, err = fxp.LengthFromString(n+" "+units, fxp.Yard); err != nil {
		return nil, err
	}
	result := yardsToValue(length, wantSize)
	if !wantSize {
		result = -result
	}
	return fxp.From(result), nil
}

func evalSSRTYards(ev *eval.Evaluator, arguments string) (any, error) {
	v, err := evalToNumber(ev, arguments)
	if err != nil {
		return nil, err
	}
	return valueToYards(fxp.As[int](v)), nil
}

// evalRandomHeight generates a random height in inches based on the chart from B18.
func evalRandomHeight(ev *eval.Evaluator, arguments string) (any, error) {
	if _, ok := ev.Resolver.(*Entity); !ok {
		return -fxp.One, nil
	}
	stDecimal, err := evalToNumber(ev, arguments)
	if err != nil {
		return nil, err
	}
	st := fxp.As[int](stDecimal)
	r := rand.NewCryptoRand()
	return fxp.From(68 + (st-10)*2 + (r.Intn(6) + 1) - (r.Intn(6) + 1)), nil
}

// evalRandomWeight generates a random weight in pounds based on the chart from B18.
func evalRandomWeight(ev *eval.Evaluator, arguments string) (any, error) {
	e, ok := ev.Resolver.(*Entity)
	if !ok {
		return -fxp.One, nil
	}
	var arg string
	arg, arguments = eval.NextArg(arguments)
	stDecimal, err := evalToNumber(ev, arg)
	if err != nil {
		return nil, err
	}
	st := fxp.As[int](stDecimal)
	var adj int
	if arguments != "" {
		var adjDecimal fxp.Int
		if adjDecimal, err = evalToNumber(ev, arguments); err != nil {
			return nil, err
		}
		adj = fxp.As[int](adjDecimal)
	}
	adj += 3 // Average
	skinny := false
	overweight := false
	fat := false
	veryFat := false
	Traverse(func(t *Trait) bool {
		switch strings.ToLower(t.NameWithReplacements()) {
		case "skinny":
			skinny = true
		case "overweight":
			overweight = true
		case "fat":
			fat = true
		case "very Fat":
			veryFat = true
		}
		return false
	}, true, false, e.Traits...)
	switch {
	case skinny:
		adj--
	case veryFat:
		adj += 3
	case fat:
		adj += 2
	case overweight:
		adj++
	}
	r := rand.NewCryptoRand()
	mid := 145 + (st-10)*15
	deviation := mid/5 + 2
	return fxp.From(((mid + r.Intn(deviation) - r.Intn(deviation)) * adj) / 3), nil
}
