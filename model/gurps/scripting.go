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
	"encoding/json/v2"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja/parser"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xos"
)

var (
	scriptStart         = "<script>"
	scriptEnd           = "</script>"
	embeddedScriptRegex = regexp.MustCompile(`(?s)` + scriptStart + `.*?` + scriptEnd)
	vmPool              = sync.Pool{New: func() any { return newScriptVM() }}
)

// freezeBuiltInsProgram freezes the built-ins: every object reachable as a global, together with their prototype
// chains and the prototypes they hand to the instances they create. Runtimes are pooled and reused by unrelated
// scripts, and strict mode does nothing to stop a script from mutating the shared state it can reach
// (`Math.exp2 = null`, `Object.prototype.toString = ...`, `Array.prototype.foo = 1`, ...), so without this a single
// script could silently alter the behavior of every script that later ran on the same runtime, including those
// belonging to other documents. What is left writable — a property added to a built-in method's function object, say —
// cannot alter what any built-in does. Host objects (the Go-backed bindings) cannot be frozen, but goja already rejects
// adding to or replacing their fields.
//
// Freezing the objects is not enough on its own, because the global bindings that name them are separate: goja creates
// `Math`, `JSON`, `Date` and the rest as writable, configurable properties of the global object, so `Math = null` or
// `delete globalThis.JSON` would leave a frozen-but-unreachable built-in behind. The second loop therefore pins every
// global binding that exists at this point as non-writable and non-configurable, which turns both of those into a
// TypeError under the strict mode every script runs in. The global object itself is deliberately left extensible, since
// each run defines its arguments on it; those, and anything else a script leaves there, are removed again by
// scriptVM.restoreGlobals.
var freezeBuiltInsProgram = goja.MustCompile("", `(function() {
	'use strict';
	var seen = new Set();
	var freeze = function(obj) {
		while (obj !== null && (typeof obj === 'object' || typeof obj === 'function') && obj !== globalThis &&
			!seen.has(obj)) {
			seen.add(obj);
			try {
				Object.freeze(obj);
			} catch (err) {
				// Host objects cannot be frozen, but they reject mutation on their own.
			}
			if (obj.prototype !== obj) {
				freeze(obj.prototype);
			}
			obj = Object.getPrototypeOf(obj);
		}
	};
	var names = Object.getOwnPropertyNames(globalThis);
	for (var i = 0; i < names.length; i++) {
		var desc = Object.getOwnPropertyDescriptor(globalThis, names[i]);
		// Only data properties are considered; reading an accessor would run arbitrary code.
		if (desc && 'value' in desc) {
			freeze(desc.value);
		}
	}
	freeze(Object.getPrototypeOf(globalThis));
	var pin = function(obj, keys) {
		for (var i = 0; i < keys.length; i++) {
			var desc = Object.getOwnPropertyDescriptor(obj, keys[i]);
			if (!desc) {
				continue;
			}
			try {
				if ('value' in desc) {
					if (desc.writable || desc.configurable) {
						Object.defineProperty(obj, keys[i], {writable: false, configurable: false});
					}
				} else if (desc.configurable) {
					// An accessor has no value to pin, but making it non-configurable stops it being replaced.
					Object.defineProperty(obj, keys[i], {configurable: false});
				}
			} catch (err) {
				// A binding that refuses to be pinned is caught by scriptVM.restoreGlobals instead.
			}
		}
	};
	pin(globalThis, Object.getOwnPropertyNames(globalThis));
	pin(globalThis, Object.getOwnPropertySymbols(globalThis));
})();`, true)

// scriptVM pairs a goja runtime with the global state it was created with. Runtimes are pooled and reused, so
// scriptVM.restoreGlobals uses the recorded baseline to strip anything a script left behind before the runtime is made
// available to the next script.
type scriptVM struct {
	runtime         *goja.Runtime
	baselineNames   map[string]goja.Value
	baselineSymbols map[*goja.Symbol]bool
}

func newScriptVM() *scriptVM {
	vm := goja.New()
	vm.SetFieldNameMapper(scriptNameMapper{})
	vm.SetParserOptions(parser.WithDisableSourceMaps)
	globals := vm.GlobalObject()
	mustSetMember(globals.Get("Math").ToObject(vm), "exp2", math.Exp2)
	mustDefineGlobal(vm, "console", scriptConsole{})
	mustDefineGlobal(vm, "dice", scriptDice{})
	mustDefineGlobal(vm, "iff", scriptIff)
	mustDefineGlobal(vm, "measure", scriptMeasurement{})
	mustDefineGlobal(vm, "signedValue", scriptSigned)
	mustDefineGlobal(vm, "formatNum", scriptFormatNum)
	if _, err := vm.RunProgram(freezeBuiltInsProgram); err != nil {
		panic(errs.NewWithCause("failed to freeze script built-ins", err))
	}
	s := &scriptVM{
		runtime:         vm,
		baselineNames:   make(map[string]goja.Value),
		baselineSymbols: make(map[*goja.Symbol]bool),
	}
	for _, name := range globals.GetOwnPropertyNames() {
		// The value is recorded, not just the name, so that restoreGlobals can tell a binding that still exists from
		// one that still holds what it was created with. Reading it here is safe: nothing has run on this runtime yet.
		s.baselineNames[name] = globals.Get(name)
	}
	for _, sym := range globals.Symbols() {
		s.baselineSymbols[sym] = true
	}
	return s
}

// restoreGlobals removes everything the run that just finished left on the global object: the arguments that were
// defined for it, plus any globals the script created itself, which strict mode does not prevent (`globalThis.foo = 1`
// and `Object.defineProperty(globalThis, ...)` both succeed). It also verifies that the baseline globals themselves
// came through the run untouched, since removing what was added says nothing about what was replaced or deleted:
// freezeBuiltInsProgram pins those bindings so neither should be possible, but a runtime whose `JSON` is missing or
// whose `Math` is no longer the frozen built-in must never be handed to another script regardless of how it got that
// way. It reports whether the runtime was fully restored; if it was not, the caller must discard the runtime rather
// than return it to the pool, since the residue would silently alter unrelated scripts run later.
func (s *scriptVM) restoreGlobals() bool {
	globals := s.runtime.GlobalObject()
	restored := true
	surviving := 0
	for _, name := range globals.GetOwnPropertyNames() {
		if baseline, isBaseline := s.baselineNames[name]; isBaseline {
			surviving++
			if !globals.Get(name).SameAs(baseline) {
				errs.LogWithLevel(context.Background(), slog.LevelWarn, nil,
					errs.New("script replaced a baseline global"), "name", name)
				restored = false
			}
			continue
		}
		if err := globals.Delete(name); err != nil {
			errs.LogWithLevel(context.Background(), slog.LevelWarn, nil, err, "name", name)
			restored = false
		}
	}
	if surviving != len(s.baselineNames) {
		errs.LogWithLevel(context.Background(), slog.LevelWarn, nil, errs.New("script deleted a baseline global"),
			"expected", len(s.baselineNames), "surviving", surviving)
		restored = false
	}
	survivingSymbols := 0
	for _, sym := range globals.Symbols() {
		if s.baselineSymbols[sym] {
			survivingSymbols++
			continue
		}
		if err := globals.DeleteSymbol(sym); err != nil {
			errs.LogWithLevel(context.Background(), slog.LevelWarn, nil, err, "symbol", sym.String())
			restored = false
		}
	}
	if survivingSymbols != len(s.baselineSymbols) {
		errs.LogWithLevel(context.Background(), slog.LevelWarn, nil,
			errs.New("script deleted a baseline symbol global"), "expected", len(s.baselineSymbols), "surviving",
			survivingSymbols)
		restored = false
	}
	return restored
}

// scriptTimeout arms the interrupt that aborts a script which runs longer than it is permitted to. time.Timer.Stop
// does not wait for an AfterFunc that has already begun running, so stopping the timer is not enough on its own: a
// timeout that fires just as its run ends could otherwise land on the runtime after another script has picked it up,
// aborting that unrelated script with a bogus timeout. release closes that window.
type scriptTimeout struct {
	vm       *goja.Runtime
	timer    *time.Timer
	lock     sync.Mutex
	finished bool
}

// newScriptTimeout returns a scriptTimeout that will interrupt the given runtime once the timeout elapses. The caller
// must call release before the runtime is used for anything else.
func newScriptTimeout(vm *goja.Runtime, timeout time.Duration) *scriptTimeout {
	t := &scriptTimeout{vm: vm}
	t.timer = time.AfterFunc(timeout, t.interrupt)
	return t
}

// interrupt aborts the run this timeout was created for, unless that run has already been released.
func (t *scriptTimeout) interrupt() {
	t.lock.Lock()
	defer t.lock.Unlock()
	if !t.finished {
		t.vm.Interrupt("timeout")
	}
}

// release disarms the timeout. Once it returns, no interrupt from this timeout can reach the runtime; one that was
// delivered after the run finished, but before the timeout could be marked as done, is cleared here.
func (t *scriptTimeout) release() {
	t.timer.Stop()
	t.lock.Lock()
	defer t.lock.Unlock()
	t.finished = true
	t.vm.ClearInterrupt()
}

// ScriptSelfProvider is a provider for the "self" variable in scripts.
type ScriptSelfProvider struct {
	ID       string
	Provider func(r *goja.Runtime) any
}

// ResolveID returns the ID of the provider. If the the underlying Provider is nil, an empty string is returned.
func (s ScriptSelfProvider) ResolveID() string {
	if s.Provider == nil {
		return ""
	}
	return s.ID
}

type scriptResolveKey struct {
	id   string
	text string
}

// ScriptArg is a named argument to be passed to RunString.
type ScriptArg struct {
	Name  string
	Value any
}

// scriptCache holds compiled programs keyed by their source text. It is shared by every entity and every goroutine that
// resolves scripts, so all access is guarded by scriptCacheMutex.
var (
	scriptCacheMutex sync.RWMutex
	scriptCache      = make(map[string]*goja.Program)
)

// globalResolveCache holds resolved results for scripts that have no associated entity. Unlike an entity's own
// scriptCache (which is only touched while recalculating that single entity), this is package-global state that may be
// reached from multiple goroutines, so all access is guarded by globalResolveMutex.
var (
	globalResolveMutex sync.Mutex
	globalResolveCache = make(map[scriptResolveKey]string)
)

// DiscardGlobalResolveCache clears the global resolve cache.
func DiscardGlobalResolveCache() {
	globalResolveMutex.Lock()
	defer globalResolveMutex.Unlock()
	clear(globalResolveCache)
}

// scriptResolveErrorSuppression tracks nested requests to suppress the error logging that normally occurs when a script
// fails to resolve to an expected result (e.g. a number or weight). The item editors resolve partially-typed—and thus
// frequently invalid—scripts to build live previews as the user types; logging every intermediate failure would flood
// the log with noise. It is an atomic counter so suppression can be nested and remains correct even if resolution ever
// spans goroutines.
var scriptResolveErrorSuppression atomic.Int32

// SuppressScriptResolveErrorLogging runs f with the error logging that normally accompanies a failed script resolution
// suppressed. Failures that occur outside the dynamic scope of f continue to be logged normally. This is intended for
// contexts such as the item editors, which repeatedly resolve incomplete scripts to produce live previews.
func SuppressScriptResolveErrorLogging(f func()) {
	scriptResolveErrorSuppression.Add(1)
	defer scriptResolveErrorSuppression.Add(-1)
	f()
}

// scriptResolveErrorLoggingSuppressed reports whether failed-resolution error logging is currently suppressed.
func scriptResolveErrorLoggingSuppressed() bool {
	return scriptResolveErrorSuppression.Load() > 0
}

// mustDefineGlobal defines a global binding that scripts may read but not replace. A plain Runtime.Set would create a
// writable global, which a script could then overwrite (`dice = null`) for every script that later reused the pooled
// runtime.
func mustDefineGlobal(vm *goja.Runtime, name string, value any) {
	if err := vm.GlobalObject().DefineDataProperty(name, vm.ToValue(value), goja.FLAG_FALSE, goja.FLAG_FALSE,
		goja.FLAG_TRUE); err != nil {
		panic(errs.Newf("failed to set %s: %s", name, err.Error()))
	}
}

// mustSetMember sets a member on an existing object (e.g. adding a function to the built-in Math object). Unlike
// Runtime.Set, this resolves a property on the object rather than creating a top-level global with a dotted name.
func mustSetMember(obj *goja.Object, name string, value any) {
	if err := obj.Set(name, value); err != nil {
		panic(errs.Newf("failed to set %s: %s", name, err.Error()))
	}
}

// ResolveText will process embedded scripts.
func ResolveText(entity *Entity, selfProvider ScriptSelfProvider, text string) string {
	return embeddedScriptRegex.ReplaceAllStringFunc(text, func(s string) string {
		return ResolveScript(entity, selfProvider, s[len(scriptStart):len(s)-len(scriptEnd)])
	})
}

// ResolveToNumber resolves the text to a fixed-point number. If the text is just a number, that value is returned,
// otherwise, it will be evaluated as Javascript and the result of that will attempt to be processed as a number. If
// this fails, a value of 0 will be returned.
func ResolveToNumber(entity *Entity, selfProvider ScriptSelfProvider, text string) fxp.Int {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return 0
	}
	if v, err := fxp.FromString(trimmed); err == nil {
		return v
	}
	result := ResolveScript(entity, selfProvider, text)
	value, err := fxp.FromString(result)
	if err != nil {
		if !scriptResolveErrorLoggingSuppressed() {
			slog.Error("unable to resolve script result to a number", "result", result, "script", text)
		}
		return 0
	}
	return value
}

// ResolveToWeight resolves the text to a weight. If the text is just a weight, that weight is returned,
// otherwise, it will be evaluated as Javascript and the result of that will attempt to be processed as a weight. If
// this fails, a weight of 0 will be returned.
func ResolveToWeight(entity *Entity, selfProvider ScriptSelfProvider, text string, defUnits fxp.WeightUnit) fxp.Weight {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return 0
	}
	if w, err := fxp.WeightFromString(trimmed, defUnits); err == nil {
		return w
	}
	result := ResolveScript(entity, selfProvider, text)
	w, err := fxp.WeightFromString(result, defUnits)
	if err != nil {
		if !scriptResolveErrorLoggingSuppressed() {
			slog.Error("unable to resolve script result to a weight", "result", result, "script", text)
		}
		return 0
	}
	return w
}

const maximumAllowedResolvingDepth = 20

// entityScriptArgName is the name the entity is bound to within scripts.
const entityScriptArgName = "entity"

// globalScriptResolvingDepth guards entity-less script resolution against runaway or circular references. Entity-scoped
// resolution uses the entity's own scriptResolvingDepth field instead (see enterScriptResolution); only scripts with no
// associated entity (e.g. nodes in a standalone library file) fall back to this package-global counter, which is atomic
// so that concurrent entity-less resolutions stay race-free.
var globalScriptResolvingDepth atomic.Int32

// enterScriptResolution increments the appropriate recursion-depth counter and returns the new depth along with a
// function that restores it. Resolution recurses through the goja boundary on the calling goroutine (resolving one
// script can read a value whose own script must be resolved); tracking the depth per-entity keeps that count accurate
// even when unrelated entities are resolved concurrently on different goroutines.
func enterScriptResolution(entity *Entity) (depth int, leave func()) {
	if entity != nil {
		entity.scriptResolvingDepth++
		return entity.scriptResolvingDepth, func() { entity.scriptResolvingDepth-- }
	}
	return int(globalScriptResolvingDepth.Add(1)), func() { globalScriptResolvingDepth.Add(-1) }
}

// ResolveScript will process a script.
func ResolveScript(entity *Entity, selfProvider ScriptSelfProvider, text string) string {
	depth, leave := enterScriptResolution(entity)
	defer leave()
	if depth > maximumAllowedResolvingDepth {
		return "script resolution exceeded maximum depth (possible circular reference)"
	}
	key := scriptResolveKey{id: selfProvider.ResolveID(), text: text}
	if cached, exists := lookupResolvedScript(entity, key); exists {
		return cached
	}
	var result string
	maxTime := GlobalSettings().General.PermittedPerScriptExecTime
	args := []ScriptArg{{
		Name:  entityScriptArgName,
		Value: func(r *goja.Runtime) any { return newScriptEntity(r, entity) },
	}}
	if selfProvider.Provider != nil {
		args = append(args, ScriptArg{
			Name:  "self",
			Value: selfProvider.Provider,
		})
	}
	if entity != nil {
		list := entity.Attributes.List()
		for _, attr := range list {
			if def := attr.AttributeDef(); def != nil {
				if def.IsSeparator() {
					continue
				}
				args = append(args, ScriptArg{
					Name:  "$" + attr.AttrID,
					Value: func(r *goja.Runtime) any { return newScriptAttribute(r, attr) },
				})
			}
		}
	}
	var err error
	xos.SafeCall(func() { result, err = runScript(fxp.SecondsToDuration(maxTime), text, args...) },
		func(panicErr error) { err = panicErr })
	if err != nil {
		//nolint:errcheck // we don't care about the error value here, just the type
		if _, ok := errors.AsType[*goja.InterruptedError](err); ok {
			result = fmt.Sprintf("script execution timed out (limited to %v seconds)", maxTime)
		} else {
			result = err.Error()
		}
	}
	storeResolvedScript(entity, key, result)
	return result
}

// lookupResolvedScript returns a previously resolved result for the given key. Entity-scoped results live in the
// entity's own cache (only touched while recalculating that entity); entity-less results live in the package-global
// cache and are read under globalResolveMutex.
func lookupResolvedScript(entity *Entity, key scriptResolveKey) (string, bool) {
	if entity != nil {
		cached, exists := entity.scriptCache[key]
		return cached, exists
	}
	globalResolveMutex.Lock()
	defer globalResolveMutex.Unlock()
	cached, exists := globalResolveCache[key]
	return cached, exists
}

// storeResolvedScript records a resolved result for the given key. See lookupResolvedScript for where each is kept.
func storeResolvedScript(entity *Entity, key scriptResolveKey, result string) {
	if entity != nil {
		entity.scriptCache[key] = result
		return
	}
	globalResolveMutex.Lock()
	defer globalResolveMutex.Unlock()
	globalResolveCache[key] = result
}

// compiledProgram returns the compiled program for the given script text, compiling and caching it on first use. The
// text is evaluated by an eval call inside an anonymous strict-mode function, so the script's value is that of its last
// expression and any variables it declares are confined to that function. Note that this does not sandbox the script:
// strict mode only rejects assignment to undeclared names, so a script can still reach shared state via globalThis and
// the built-ins. Keeping the pooled runtimes clean is handled instead by freezeBuiltInsProgram and
// scriptVM.restoreGlobals.
func compiledProgram(text string) (*goja.Program, error) {
	scriptCacheMutex.RLock()
	program, exists := scriptCache[text]
	scriptCacheMutex.RUnlock()
	if exists {
		return program, nil
	}
	jsBytes, err := json.Marshal(text)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal script text: %w", err)
	}
	if program, err = goja.Compile("", "(function() { 'use strict'; return eval("+string(jsBytes)+"); })();", true); err != nil {
		return nil, fmt.Errorf("failed to compile script: %w", err)
	}
	scriptCacheMutex.Lock()
	defer scriptCacheMutex.Unlock()
	// Re-check in case another goroutine compiled and stored the same text while we were compiling, so all callers
	// share a single program instance.
	if existing, ok := scriptCache[text]; ok {
		return existing, nil
	}
	scriptCache[text] = program
	return program, nil
}

// runScript compiles and runs a script with the provided arguments, returning the string form of its result. A timeout
// of 0 or less means no timeout. The script text is evaluated as a JavaScript expression or sequence of statements (see
// compiledProgram), so its value is that of its last expression; a top-level `return` is not permitted. The arguments
// are exposed as globals for the duration of the run and removed again afterwards. The result is converted to a string
// here, rather than by the caller, because that conversion can itself run script code — an object's `toString` or
// `valueOf` — which must happen while this runtime is still checked out of the pool and still covered by the timeout.
func runScript(timeout time.Duration, text string, args ...ScriptArg) (string, error) {
	program, err := compiledProgram(text)
	if err != nil {
		return "", err
	}
	s, ok := vmPool.Get().(*scriptVM)
	if !ok {
		return "", errors.New("failed to get VM from pool")
	}
	vm := s.runtime
	globals := vm.GlobalObject()
	reusable := false
	defer func() {
		// Only return the VM to the pool if the run completed without panicking and every global it touched could be
		// restored. A panic (e.g. from a Go function invoked by the script) can leave the VM in an inconsistent state,
		// and a global that could not be removed would alter later scripts, so in either case we discard the VM and let
		// the pool create a fresh one rather than risk reusing a corrupt or polluted one.
		if reusable && s.restoreGlobals() {
			vmPool.Put(s)
		}
	}()
	for _, arg := range args {
		if valueProvider, ok2 := arg.Value.(func(r *goja.Runtime) any); ok2 {
			var cachedResult goja.Value
			if err = globals.DefineAccessorProperty(arg.Name, vm.ToValue(func(_ goja.FunctionCall) goja.Value {
				if cachedResult == nil {
					cachedResult = vm.ToValue(valueProvider(vm))
				}
				return cachedResult
			}), nil, goja.FLAG_TRUE, goja.FLAG_TRUE); err != nil {
				return "", fmt.Errorf("failed to define accessor for argument %q: %w", arg.Name, err)
			}
			continue
		}
		if err = vm.Set(arg.Name, arg.Value); err != nil {
			return "", fmt.Errorf("failed to set argument %q: %w", arg.Name, err)
		}
	}
	if timeout > 0 {
		defer newScriptTimeout(vm, timeout).release()
	}
	value, err := vm.RunProgram(program)
	if err != nil {
		// A returned error (a timeout interrupt or a script exception) leaves the VM reusable; only a panic, which
		// would prevent reaching this line, marks it as corrupt.
		reusable = true
		return "", err
	}
	// This may panic if the value's conversion throws or is interrupted; the VM is then left out of the pool, since a
	// partially unwound conversion may have left it inconsistent.
	result := value.String()
	reusable = true
	return result, nil
}

// scriptLevel converts a computed level for consumption by a script. An uncomputable level is stored as the fxp.Min
// sentinel, which would otherwise be handed to scripts as a nonsensical -922337203685477, so clamp it to 0, matching
// what the weapon wrapper does.
func scriptLevel(level Level) int {
	return fxp.AsInteger[int](level.Level.Max(0))
}

// scriptRelativeLevel converts a computed relative level for consumption by a script. The relative level is meaningless
// when the level itself couldn't be computed, so report 0 rather than leaking a sentinel value.
func scriptRelativeLevel(level Level) int {
	if level.Level == fxp.Min || level.RelativeLevel == fxp.Min {
		return 0
	}
	return fxp.AsInteger[int](level.RelativeLevel)
}

func callArgAsTrimmedString(call goja.FunctionCall, index int) string {
	return strings.TrimSpace(callArgAsString(call, index))
}

func callArgAsString(call goja.FunctionCall, index int) string {
	if arg := call.Argument(index); !goja.IsUndefined(arg) {
		return arg.String()
	}
	return ""
}

func scriptIff(condition bool, trueValue, falseValue any) any {
	if condition {
		return trueValue
	}
	return falseValue
}

func scriptSigned(value float64) string {
	return fxp.FromFloat(value).StringWithSign()
}

func scriptFormatNum(value float64, withCommas, withSign bool) string {
	v := fxp.FromFloat(value)
	if withSign {
		if withCommas {
			return v.CommaWithSign()
		}
		return v.StringWithSign()
	}
	if withCommas {
		return v.Comma()
	}
	return v.String()
}
