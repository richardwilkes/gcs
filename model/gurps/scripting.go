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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"regexp"
	"strings"
	"sync"
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
	scriptCache         = make(map[string]*goja.Program)
	globalResolveCache  = make(map[scriptResolveKey]string)
	vmPool              = sync.Pool{New: func() any {
		vm := goja.New()
		vm.SetFieldNameMapper(scriptNameMapper{})
		vm.SetParserOptions(parser.WithDisableSourceMaps)
		mustSet(vm, "console", scriptConsole{})
		mustSet(vm, "dice", scriptDice{})
		mustSet(vm, "iff", scriptIff)
		mustSet(vm, "measure", scriptMeasurement{})
		mustSet(vm, "Math.exp2", math.Exp2)
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

// DiscardGlobalResolveCache clears the global resolve cache.
func DiscardGlobalResolveCache() {
	if len(globalResolveCache) != 0 {
		globalResolveCache = make(map[scriptResolveKey]string)
	}
}

func mustSet(vm *goja.Runtime, name string, value any) {
	if err := vm.Set(name, value); err != nil {
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

var scriptResolvingDepth = 0

// ResolveScript will process a script.
func ResolveScript(entity *Entity, selfProvider ScriptSelfProvider, text string) string {
	scriptResolvingDepth++
	defer func() {
		scriptResolvingDepth--
	}()
	if scriptResolvingDepth > maximumAllowedResolvingDepth {
		return "script resolution exceeded maximum depth (possible circular reference)"
	}
	var resolveCache map[scriptResolveKey]string
	if entity == nil {
		resolveCache = globalResolveCache
	} else {
		resolveCache = entity.scriptCache
	}
	key := scriptResolveKey{id: selfProvider.ResolveID(), text: text}
	if cached, exists := resolveCache[key]; exists {
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
	resolveCache[key] = result
	return result
}

// runScript compiles and runs a script with the provided arguments. A timeout of 0 or less means no timeout.
// The script should be a valid JavaScript function body, and it will be wrapped in an anonymous function to avoid
// polluting the global scope. The arguments will be set as global variables in the script's context. The return value
// is the result of the script execution, or an error if it fails.
func runScript(timeout time.Duration, text string, args ...ScriptArg) (goja.Value, error) {
	program, exists := scriptCache[text]
	if !exists {
		jsBytes, err := json.Marshal(text)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal script text: %w", err)
		}
		program, err = goja.Compile("", "(function() { 'use strict'; return eval("+string(jsBytes)+"); })();", true)
		if err != nil {
			return nil, fmt.Errorf("failed to compile script: %w", err)
		}
		scriptCache[text] = program
	}
	vm, ok := vmPool.Get().(*goja.Runtime)
	if !ok {
		return nil, errors.New("failed to get VM from pool")
	}
	globals := vm.GlobalObject()
	defer func() {
		for _, arg := range args {
			if err := globals.Delete(arg.Name); err != nil {
				errs.LogWithLevel(context.Background(), slog.LevelWarn, nil, err, "name", arg.Name)
			}
		}
		vm.ClearInterrupt()
		vmPool.Put(vm)
	}()
	for _, arg := range args {
		if valueProvider, ok2 := arg.Value.(func(r *goja.Runtime) any); ok2 {
			var cachedResult goja.Value
			if err := globals.DefineAccessorProperty(arg.Name, vm.ToValue(func(_ goja.FunctionCall) goja.Value {
				if cachedResult == nil {
					cachedResult = vm.ToValue(valueProvider(vm))
				}
				return cachedResult
			}), nil, goja.FLAG_TRUE, goja.FLAG_TRUE); err != nil {
				return nil, fmt.Errorf("failed to define accessor for argument %q: %w", arg.Name, err)
			}
			continue
		}
		if err := vm.Set(arg.Name, arg.Value); err != nil {
			return nil, fmt.Errorf("failed to set argument %q: %w", arg.Name, err)
		}
	}
	if timeout > 0 {
		defer time.AfterFunc(timeout, func() { vm.Interrupt("timeout") }).Stop()
	}
	return vm.RunProgram(program)
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
