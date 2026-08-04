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
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
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
		c.Equal(tc.want, v, "script %q", tc.script)
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
		_, err = fxp.FromString(v)
		c.NoError(err, "st %d result %q", st, v)
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
		want  fxp.Int
	}{
		{field: "total", want: pb.Total()},
		{field: "unspent", want: e.TotalPoints - pb.Total()},
		{field: "skills", want: pb.Skills},
		{field: "spells", want: pb.Spells},
		{field: "attributes", want: pb.Attributes},
	} {
		v, err := runScript(0, "entity.points."+tc.field, entityArg)
		c.NoError(err, "field %q", tc.field)
		got, err := fxp.FromString(v)
		c.NoError(err, "field %q result %q", tc.field, v)
		c.Equal(tc.want, got, "field %q", tc.field)
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
			c.Equal(want, v, "%s extra=%v", fn, extra)
		}
	}
}

// TestScriptTraitPoints verifies that reading self.points from a trait script returns the trait's AdjustedPoints(),
// which already accounts for base points, cost-per-level, trait modifiers, and the reduced cost of children within an
// Alternative Abilities container. This addresses GitHub issue #1053 (exposing a trait's total value to scripts).
func TestScriptTraitPoints(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	// A simple, non-leveled trait.
	simple := NewTrait(e, nil, false)
	simple.BasePoints = fxp.FromInteger(10)

	// A leveled trait: 5 points per level, 3 levels.
	leveled := NewTrait(e, nil, false)
	leveled.CanLevel = true
	leveled.PointsPerLevel = fxp.FromInteger(5)
	leveled.Levels = fxp.FromInteger(3)

	// An Alternative Abilities container whose children have differing costs; the container total is reduced per the
	// Alternative Abilities rules rather than being the simple sum.
	altAbilities := NewTrait(e, nil, true)
	altAbilities.ContainerType = container.AlternativeAbilities
	child1 := NewTrait(e, altAbilities, false)
	child1.BasePoints = fxp.FromInteger(20)
	child2 := NewTrait(e, altAbilities, false)
	child2.BasePoints = fxp.FromInteger(10)
	altAbilities.Children = []*Trait{child1, child2}

	for _, trait := range []*Trait{simple, leveled, altAbilities} {
		want := trait.AdjustedPoints()
		selfArg := ScriptArg{Name: "self", Value: func(r *goja.Runtime) any { return newScriptTrait(r, trait) }}
		v, err := runScript(0, "self.points", selfArg)
		c.NoError(err, "trait %q", trait.NameWithReplacements())
		got, err := fxp.FromString(v)
		c.NoError(err, "trait %q result %q", trait.NameWithReplacements(), v)
		c.Equal(want, got, "trait %q", trait.NameWithReplacements())
	}

	// Sanity checks on the expected values so the test is meaningful even if AdjustedPoints changes.
	c.Equal(fxp.FromInteger(10), simple.AdjustedPoints())
	c.Equal(fxp.FromInteger(15), leveled.AdjustedPoints())
	// 20 (the most expensive) + round(20% of 10) = 20 + 2 = 22.
	c.Equal(fxp.FromInteger(22), altAbilities.AdjustedPoints())
}

// TestScriptSkillOptionalSpecializationMatch verifies that the script skill-lookup bindings match a skill by its
// optional specialization as well as its required specialization. This addresses GitHub issue #1062, where
// entity.findSkills(name, specialization), entity.skillLevel(name, specialization), and a skill container's
// find(name, specialization) only checked the required specialization.
func TestScriptSkillOptionalSpecializationMatch(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	// A top-level skill with both a required and an optional specialization.
	sk := NewSkill(e, nil, false)
	sk.Name = "Guns"
	sk.Specialization = "Pistol"
	sk.OptionalSpecialization = "Glock"
	sk.Points = fxp.One

	// A container holding a child skill that likewise carries an optional specialization.
	group := NewSkill(e, nil, true)
	group.Name = "Group"
	child := NewSkill(e, group, false)
	child.Name = "Longbow"
	child.Specialization = "Musket"
	child.OptionalSpecialization = "Brown Bess"
	child.Points = fxp.One
	group.Children = []*Skill{child}

	e.Skills = []*Skill{sk, group}
	e.Recalculate()

	entityArg := ScriptArg{Name: entityScriptArgName, Value: func(r *goja.Runtime) any { return newScriptEntity(r, e) }}
	for _, tc := range []struct {
		name   string
		script string
		want   string
	}{
		// findSkills matches by required, optional, or (empty) any specialization.
		{name: "findSkills required", script: `entity.findSkills("Guns", "Pistol", "").length`, want: "1"},
		{name: "findSkills optional", script: `entity.findSkills("Guns", "Glock", "").length`, want: "1"},
		{name: "findSkills mismatch", script: `entity.findSkills("Guns", "Revolver", "").length`, want: "0"},
		{name: "findSkills any", script: `entity.findSkills("Guns", "", "").length`, want: "1"},
		// skillLevel resolves a level when matched by either specialization, and 0 when unmatched.
		{name: "skillLevel required", script: `entity.skillLevel("Guns", "Pistol", false) > 0 ? 1 : 0`, want: "1"},
		{name: "skillLevel optional", script: `entity.skillLevel("Guns", "Glock", false) > 0 ? 1 : 0`, want: "1"},
		{name: "skillLevel mismatch", script: `entity.skillLevel("Guns", "Revolver", false)`, want: "0"},
		// A container's find() matches its children by required or optional specialization.
		{name: "container find required", script: `entity.findSkills("Group", "", "")[0].find("Longbow", "Musket", "").length`, want: "1"},
		{name: "container find optional", script: `entity.findSkills("Group", "", "")[0].find("Longbow", "Brown Bess", "").length`, want: "1"},
		{name: "container find mismatch", script: `entity.findSkills("Group", "", "")[0].find("Longbow", "Flintlock", "").length`, want: "0"},
	} {
		v, err := runScript(0, tc.script, entityArg)
		c.NoError(err, "case %q", tc.name)
		c.Equal(tc.want, v, "case %q", tc.name)
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

// errorCountingHandler is a slog.Handler that counts the error-level records it is asked to handle. It lets a test
// observe whether a failed script resolution logged an error without depending on the log's textual format.
type errorCountingHandler struct {
	count *atomic.Int32
}

func (h errorCountingHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= slog.LevelError
}

// Handle takes the record by value because the slog.Handler interface requires that signature.
func (h errorCountingHandler) Handle(_ context.Context, record slog.Record) error { //nolint:gocritic // interface-mandated signature
	if record.Level >= slog.LevelError {
		h.count.Add(1)
	}
	return nil
}

func (h errorCountingHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }

func (h errorCountingHandler) WithGroup(_ string) slog.Handler { return h }

// TestSuppressScriptResolveErrorLogging verifies that SuppressScriptResolveErrorLogging silences the error logging that
// a failed script resolution would otherwise produce, that the suppression only applies within the dynamic scope of the
// supplied function (so failures elsewhere still log), and that nested suppression is handled correctly.
func TestSuppressScriptResolveErrorLogging(t *testing.T) {
	c := check.New(t)
	var count atomic.Int32
	prev := slog.Default()
	slog.SetDefault(slog.New(errorCountingHandler{count: &count}))
	defer slog.SetDefault(prev)

	// The script resolved below produces a non-numeric string, so ResolveToNumber fails to parse it and would log an
	// error.
	resolve := func() {
		count.Store(0)
		DiscardGlobalResolveCache()
		c.Equal(fxp.Int(0), ResolveToNumber(nil, ScriptSelfProvider{}, "'not a number'"))
	}

	// Baseline: without suppression, the failed resolution logs exactly one error.
	resolve()
	c.Equal(int32(1), count.Load(), "an error should be logged when not suppressed")

	// Within SuppressScriptResolveErrorLogging, the identical failure logs nothing.
	count.Store(0)
	DiscardGlobalResolveCache()
	SuppressScriptResolveErrorLogging(func() {
		c.Equal(fxp.Int(0), ResolveToNumber(nil, ScriptSelfProvider{}, "'not a number'"))
	})
	c.Equal(int32(0), count.Load(), "no error should be logged while suppressed")

	// Nested suppression stays suppressed until the outermost scope exits.
	count.Store(0)
	DiscardGlobalResolveCache()
	SuppressScriptResolveErrorLogging(func() {
		SuppressScriptResolveErrorLogging(func() {
			c.Equal(fxp.Int(0), ResolveToNumber(nil, ScriptSelfProvider{}, "'not a number'"))
		})
		// Still within the outer scope, so logging remains suppressed.
		c.Equal(fxp.Int(0), ResolveToNumber(nil, ScriptSelfProvider{}, "'not a number'"))
	})
	c.Equal(int32(0), count.Load(), "no error should be logged within nested suppression")

	// Scope ends: once suppression returns, logging resumes for failures produced anywhere else.
	resolve()
	c.Equal(int32(1), count.Load(), "error logging should resume after suppression returns")
}

// TestScriptResultConversionHonorsTimeout verifies that converting a script's result to a string happens while the
// runtime is still checked out and the timeout is still armed. Turning an object into a string runs the object's
// toString, so a toString that never returns must be cut short by the per-script time limit and reported as a timeout
// rather than hanging the application.
func TestScriptResultConversionHonorsTimeout(t *testing.T) {
	c := check.New(t)
	settings := GlobalSettings().General
	prev := settings.PermittedPerScriptExecTime
	settings.PermittedPerScriptExecTime = fxp.Tenth
	defer func() { settings.PermittedPerScriptExecTime = prev }()
	DiscardGlobalResolveCache()
	defer DiscardGlobalResolveCache()

	done := make(chan string, 1)
	go func() {
		done <- ResolveScript(nil, ScriptSelfProvider{}, "({toString: function() { for (;;); }})")
	}()
	select {
	case result := <-done:
		c.HasPrefix(result, "script execution timed out")
	case <-time.After(30 * time.Second):
		t.Fatal("conversion of the script result was not interrupted by the timeout")
	}

	// A well-behaved object result is still converted normally.
	c.Equal("converted", ResolveScript(nil, ScriptSelfProvider{}, "({toString: function() { return 'converted' }})"))
}

// TestScriptObjectResultConcurrency resolves object-valued scripts, whose results can only be turned into strings by
// running JavaScript, from several goroutines at once. Run under the race detector (go test -race) it guards against
// converting a result after its runtime has been handed back to the pool, where another goroutine may already be using
// it. The goroutine count is kept below maximumAllowedResolvingDepth so concurrency alone cannot trip the depth limit.
func TestScriptObjectResultConcurrency(t *testing.T) {
	DiscardGlobalResolveCache()
	defer DiscardGlobalResolveCache()
	const goroutines = 8
	const iterations = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := range goroutines {
		go func() {
			defer wg.Done()
			for i := range iterations {
				// The script text is unique per (goroutine, iteration) so each resolution actually runs rather than
				// being served from the resolve cache.
				want := fmt.Sprintf("g%d-%d", g, i)
				script := fmt.Sprintf("({toString: function() { return %q }})", want)
				if got := ResolveScript(nil, ScriptSelfProvider{}, script); got != want {
					t.Errorf("object result = %q, want %q", got, want)
				}
				if i%25 == 0 {
					DiscardGlobalResolveCache()
				}
			}
		}()
	}
	wg.Wait()
}

// TestScriptGlobalPollutionDoesNotLeak verifies that a script cannot alter the shared state of a pooled runtime for the
// scripts that follow it. Strict mode alone does not provide this: it only rejects assignment to undeclared names, so
// adding globals through globalThis, replacing the predefined bindings, and mutating the built-ins all have to be
// prevented or undone explicitly.
func TestScriptGlobalPollutionDoesNotLeak(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name    string
		script  string
		rejects bool   // whether the script itself is expected to fail
		probe   string // evaluated afterwards; must yield want, proving nothing leaked
		want    string
	}{
		{
			name:  "added global",
			probe: "typeof gcsLeakedGlobal",
			want:  "undefined",
			// A plain assignment would be rejected by strict mode, but going through globalThis is allowed.
			script: "globalThis.gcsLeakedGlobal = 1; 'ok'",
		},
		{
			name:   "added non-configurable global",
			script: "Object.defineProperty(globalThis, 'gcsPinnedGlobal', {value: 1}); 'ok'",
			probe:  "typeof gcsPinnedGlobal",
			want:   "undefined",
		},
		{
			name:   "added symbol-keyed global",
			script: "globalThis[Symbol.for('gcsLeakedSymbol')] = 1; 'ok'",
			probe:  "typeof globalThis[Symbol.for('gcsLeakedSymbol')]",
			want:   "undefined",
		},
		{
			name:    "replaced predefined global",
			script:  "dice = null; 'ok'",
			rejects: true,
			probe:   "typeof dice",
			want:    "object",
		},
		{
			name:    "replaced built-in member",
			script:  "Math.exp2 = null; 'ok'",
			rejects: true,
			probe:   "Math.exp2(3)",
			want:    "8",
		},
		{
			name:    "replaced built-in prototype member",
			script:  "Object.prototype.toString = null; 'ok'",
			rejects: true,
			probe:   "({}).toString()",
			want:    "[object Object]",
		},
		{
			name:    "added built-in prototype member",
			script:  "Array.prototype.gcsLeakedMember = 1; 'ok'",
			rejects: true,
			probe:   "typeof [].gcsLeakedMember",
			want:    "undefined",
		},
	} {
		_, err := runScript(0, tc.script)
		if tc.rejects {
			c.HasError(err, "case %q should have been rejected", tc.name)
		} else {
			c.NoError(err, "case %q", tc.name)
		}
		// Probe repeatedly: a runtime that was polluted rather than restored or discarded is likely, but not certain,
		// to be the one handed out for the very next run.
		for range 3 {
			v, probeErr := runScript(0, tc.probe)
			c.NoError(probeErr, "case %q probe", tc.name)
			c.Equal(tc.want, v, "case %q probe", tc.name)
		}
	}

	// The arguments a run is given are bound as globals too, and must likewise be gone by the time a script that was
	// given none is run.
	entityArg := ScriptArg{Name: entityScriptArgName, Value: func(r *goja.Runtime) any { return newScriptEntity(r, nil) }}
	v, err := runScript(0, "typeof "+entityScriptArgName, entityArg)
	c.NoError(err)
	c.Equal("object", v)
	for range 3 {
		v, err = runScript(0, "typeof "+entityScriptArgName)
		c.NoError(err)
		c.Equal("undefined", v)
	}
}

// TestScriptTimeoutDoesNotAffectNextRun verifies that a script's timeout cannot interrupt a later, unrelated script.
// Stopping the timer is not enough on its own, because time.Timer.Stop does not wait for a timeout that has already
// begun firing; that timeout would otherwise land on the runtime after the next script had picked it up. The window is
// microseconds wide in practice, so it is reproduced here by invoking the timeout's interrupt directly, exactly as a
// timer goroutine that had already started would.
func TestScriptTimeoutDoesNotAffectNextRun(t *testing.T) {
	c := check.New(t)
	vm := goja.New()

	// An hour, so only the explicit calls below can fire the timeout.
	timeout := newScriptTimeout(vm, time.Hour)
	timeout.release()
	timeout.interrupt() // The timer goroutine, already running when release stopped it, fires late.
	v, err := vm.RunString("1 + 1")
	c.NoError(err, "a released timeout must not interrupt the runtime")
	c.Equal(int64(2), v.ToInteger())

	// A timeout that fires while its own run is still in progress does interrupt, and releasing it afterwards leaves
	// the runtime usable again rather than carrying the interrupt into the next run.
	timeout = newScriptTimeout(vm, time.Hour)
	timeout.interrupt()
	_, err = vm.RunString("1 + 1")
	c.HasError(err, "a timeout that fires during its own run must interrupt the runtime")
	var interruptedErr *goja.InterruptedError
	c.True(errors.As(err, &interruptedErr), "expected an interrupt, got %v", err)
	timeout.release()
	v, err = vm.RunString("1 + 1")
	c.NoError(err, "releasing a timeout must clear the interrupt it delivered")
	c.Equal(int64(2), v.ToInteger())

	// End to end: a script that really does time out must not poison the run that follows it.
	_, err = runScript(10*time.Millisecond, "(function() { for (;;); })()")
	c.HasError(err, "the script should have timed out")
	result, err := runScript(0, "1 + 1")
	c.NoError(err)
	c.Equal("2", result)
}

// TestScriptEvalSemantics pins down the contract runScript actually implements: the text is evaluated as an expression
// or sequence of statements whose value is that of its last expression, declarations within it do not escape into the
// runtime, and — because the text is evaluated rather than used as a function body — a top-level return is a syntax
// error.
func TestScriptEvalSemantics(t *testing.T) {
	c := check.New(t)

	v, err := runScript(0, "6 * 7")
	c.NoError(err)
	c.Equal("42", v)

	v, err = runScript(0, "var gcsScoped = 6; gcsScoped * 7")
	c.NoError(err)
	c.Equal("42", v)

	// The declaration above must not have become a global.
	v, err = runScript(0, "typeof gcsScoped")
	c.NoError(err)
	c.Equal("undefined", v)

	_, err = runScript(0, "return 42")
	c.HasError(err)
	c.Contains(err.Error(), "Illegal return statement")
}
