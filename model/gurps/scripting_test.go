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
	"fmt"
	"sync"
	"testing"

	"github.com/dop251/goja"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestScriptMathExp2 verifies that Math.exp2 is exposed to scripts as a real member of the built-in Math object (rather
// than an unreachable global with a dotted name), while leaving the standard Math members intact.
func TestScriptMathExp2(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		script string
		want   string
	}{
		{script: "typeof Math.exp2", want: "function"},
		{script: "Math.exp2(0)", want: "1"},
		{script: "Math.exp2(3)", want: "8"},
		{script: "Math.exp2(10)", want: "1024"},
		{script: "typeof Math.pow", want: "function"}, // ensure the built-in Math members are still present
	} {
		v, err := runScript(0, tc.script)
		c.NoError(err, "script %q", tc.script)
		c.Equal(tc.want, v.String(), "script %q", tc.script)
	}
}

// TestScriptRandomWeightInPounds verifies that entity.randomWeightInPounds returns a numeric result without panicking
// for very low (even negative) strength values. At such ST the internal deviation used as the bound for rnd.Intn would
// otherwise go non-positive; the clamp keeps the call safe regardless of how rnd.Intn treats a non-positive bound.
func TestScriptRandomWeightInPounds(t *testing.T) {
	c := check.New(t)
	entityArg := ScriptArg{Name: entityScriptArgName, Value: func(r *goja.Runtime) any { return newScriptEntity(r, nil) }}
	for _, st := range []int{-100, -20, -1, 0, 1, 10, 20} {
		v, err := runScript(0, fmt.Sprintf("entity.randomWeightInPounds(%d, 0)", st), entityArg)
		c.NoError(err, "st %d", st)
		c.NotNil(v, "st %d", st)
	}
}

// TestScriptEntityPoints verifies that reading the fields of entity.points reports values consistent with
// entity.PointsBreakdown(). The implementation computes the breakdown once per access to entity.points and reads each
// field (total, unspent, skills, spells, …) from that single result, so this also guards against the fields drifting
// apart from the canonical breakdown.
func TestScriptEntityPoints(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	sk := NewSkill(e, nil, false)
	sk.Points = fxp.FromInteger(2)
	e.Skills = append(e.Skills, sk)
	sp := NewSpell(e, nil, false)
	sp.Points = fxp.FromInteger(3)
	e.Spells = append(e.Spells, sp)
	e.Recalculate()
	e.TotalPoints = fxp.FromInteger(10)

	pb := e.PointsBreakdown()
	entityArg := ScriptArg{Name: entityScriptArgName, Value: func(r *goja.Runtime) any { return newScriptEntity(r, e) }}
	for _, tc := range []struct {
		field string
		want  int
	}{
		{field: "total", want: fxp.AsInteger[int](pb.Total())},
		{field: "unspent", want: fxp.AsInteger[int](e.TotalPoints - pb.Total())},
		{field: "skills", want: fxp.AsInteger[int](pb.Skills)},
		{field: "spells", want: fxp.AsInteger[int](pb.Spells)},
		{field: "attributes", want: fxp.AsInteger[int](pb.Attributes)},
	} {
		v, err := runScript(0, "entity.points."+tc.field, entityArg)
		c.NoError(err, "field %q", tc.field)
		c.Equal(int64(tc.want), v.ToInteger(), "field %q", tc.field)
	}
}

// TestScriptThrustSwingFor exercises the entity.thrustFor and entity.swingFor script bindings, which format their dice
// using the UseModifyingDicePlusAdds flag read through the entity's sheet settings (via SheetSettingsFor(entity)). The
// GURPS damage progression keeps thrust/swing modifiers within the -3..+2 range, so the plain and modifying-dice
// formats never actually diverge for these dice; the flag is therefore behaviorally invisible here. The test guards the
// wiring itself: the bindings must exist, resolve without error, and return exactly what FormatDice produces for the
// entity's current setting, regardless of how that flag is toggled.
func TestScriptThrustSwingFor(t *testing.T) {
	c := check.New(t)
	for _, fn := range []string{"thrustFor", "swingFor"} {
		roll := func(e *Entity, st int) dice.Dice {
			if fn == "thrustFor" {
				return e.ThrustFor(st)
			}
			return e.SwingFor(st)
		}
		for _, extra := range []bool{false, true} {
			e := NewEntity()
			e.SheetSettings.UseModifyingDicePlusAdds = extra
			const st = 20 // Yields a multi-die result (thrust 2d-1, swing 3d+2) rather than a trivial 1d.
			want := FormatDice(roll(e, st), extra)
			entityArg := ScriptArg{Name: entityScriptArgName, Value: func(r *goja.Runtime) any { return newScriptEntity(r, e) }}
			v, err := runScript(0, fmt.Sprintf("entity.%s(%d)", fn, st), entityArg)
			c.NoError(err, "%s extra=%v", fn, extra)
			c.Equal(want, v.String(), "%s extra=%v", fn, extra)
		}
	}
}

// TestScriptResolutionConcurrency hammers the package-global script state (the compiled-program cache, the entity-less
// resolve cache and its discard path, and the entity-less recursion-depth counter) from several goroutines at once. Run
// under the race detector (go test -race) it guards against reintroducing unsynchronized access to that shared state.
// The goroutine count is kept below maximumAllowedResolvingDepth so concurrency alone cannot trip the depth limit.
func TestScriptResolutionConcurrency(t *testing.T) {
	DiscardGlobalResolveCache()
	const goroutines = 8
	const iterations = 200
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := range goroutines {
		go func() {
			defer wg.Done()
			for i := range iterations {
				// A shared script (a cache hit once warmed) exercises concurrent reads of the same key.
				if got := ResolveScript(nil, ScriptSelfProvider{}, "3 + 4"); got != "7" {
					t.Errorf("shared script = %q, want %q", got, "7")
				}
				// A unique script per (goroutine, iteration) exercises concurrent compile + store.
				_ = ResolveScript(nil, ScriptSelfProvider{}, fmt.Sprintf("%d + %d", g, i))
				if i%25 == 0 {
					DiscardGlobalResolveCache()
				}
			}
		}()
	}
	wg.Wait()
}
