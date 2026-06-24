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
	vmPool              = sync.Pool{New: func() any {
		vm := goja.New()
		vm.SetFieldNameMapper(scriptNameMapper{})
		vm.SetParserOptions(parser.WithDisableSourceMaps)
		mustSet(vm, "console", scriptConsole{})
		mustSet(vm, "dice", scriptDice{})
		mustSet(vm, "iff", scriptIff)
		mustSet(vm, "measure", scriptMeasurement{})
		mustSetMember(vm.GlobalObject().Get("Math").ToObject(vm), "exp2", math.Exp2)
		mustSet(vm, "signedValue", scriptSigned)
		mustSet(vm, "formatNum", scriptFormatNum)
		return vm
	}}
)

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

func mustSet(vm *goja.Runtime, name string, value any) {
	if err := vm.Set(name, value); err != nil {
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
		slog.Error("unable to resolve script result to a number", "result", result, "script", text)
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
		slog.Error("unable to resolve script result to a weight", "result", result, "script", text)
		return 0
	}
	return w
}

const maximumAllowedResolvingDepth = 20

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
		Name:  "entity",
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
	var v goja.Value
	var err error
	xos.SafeCall(func() { v, err = runScript(fxp.SecondsToDuration(maxTime), text, args...) },
		func(panicErr error) { err = panicErr })
	if err != nil {
		var interruptedErr *goja.InterruptedError
		if errors.As(err, &interruptedErr) {
			result = fmt.Sprintf("script execution timed out (limited to %v seconds)", maxTime)
		} else {
			result = err.Error()
		}
	} else {
		xos.SafeCall(func() { result = v.String() }, func(panicErr error) { result = panicErr.Error() })
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
// text is wrapped in an anonymous strict-mode function so it cannot pollute the global scope.
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

// runScript compiles and runs a script with the provided arguments. A timeout of 0 or less means no timeout.
// The script should be a valid JavaScript function body, and it will be wrapped in an anonymous function to avoid
// polluting the global scope. The arguments will be set as global variables in the script's context. The return value
// is the result of the script execution, or an error if it fails.
func runScript(timeout time.Duration, text string, args ...ScriptArg) (goja.Value, error) {
	program, err := compiledProgram(text)
	if err != nil {
		return nil, err
	}
	vm, ok := vmPool.Get().(*goja.Runtime)
	if !ok {
		return nil, errors.New("failed to get VM from pool")
	}
	// Clear any interrupt that might still be set from a prior, timed-out execution before running. time.Timer.Stop
	// does not wait for an AfterFunc that has already begun firing, so a timeout interrupt can land on the VM after it
	// was returned to the pool. Clearing it here, on the next use, guarantees it cannot spuriously interrupt this run.
	vm.ClearInterrupt()
	globals := vm.GlobalObject()
	reusable := false
	defer func() {
		// Only return the VM to the pool if the run completed without panicking. A panic (e.g. from a Go function
		// invoked by the script) can leave the VM in an inconsistent state, so in that case we discard it and let the
		// pool create a fresh one rather than risk reusing a corrupt VM.
		if !reusable {
			return
		}
		for _, arg := range args {
			if localErr := globals.Delete(arg.Name); localErr != nil {
				errs.LogWithLevel(context.Background(), slog.LevelWarn, nil, localErr, "name", arg.Name)
			}
		}
		vmPool.Put(vm)
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
				return nil, fmt.Errorf("failed to define accessor for argument %q: %w", arg.Name, err)
			}
			continue
		}
		if err = vm.Set(arg.Name, arg.Value); err != nil {
			return nil, fmt.Errorf("failed to set argument %q: %w", arg.Name, err)
		}
	}
	if timeout > 0 {
		defer time.AfterFunc(timeout, func() { vm.Interrupt("timeout") }).Stop()
	}
	value, err := vm.RunProgram(program)
	// A returned error (including a timeout interrupt or a script exception) leaves the VM reusable; only a panic,
	// which would prevent reaching this line, marks it as corrupt.
	reusable = true
	return value, err
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
