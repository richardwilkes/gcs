package gurps

import (
	"context"
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
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/tid"
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
		return vm
	}}
)

// ScriptSelfProvider is a provider for the "self" variable in scripts.
type ScriptSelfProvider struct {
	ID       tid.TID
	Provider func() any
}

// ResolveID returns the ID of the provider. If the the underlying Provider is nil, an empty string is returned.
func (s ScriptSelfProvider) ResolveID() tid.TID {
	if s.Provider == nil {
		return ""
	}
	return s.ID
}

type scriptResolveKey struct {
	id   tid.TID
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

func scriptIff(condition bool, trueValue, falseValue any) any {
	if condition {
		return trueValue
	}
	return falseValue
}

func scriptSigned(value float64) string {
	return fxp.From(value).StringWithSign()
}

// ResolveText will process the text as a script if it starts with ^^^. If it does not, it will look for embedded
// expressions inside || pairs inside the text and evaluate them.
func ResolveText(entity *Entity, selfProvider ScriptSelfProvider, text string) string {
	return embeddedScriptRegex.ReplaceAllStringFunc(text, func(s string) string {
		return resolveScript(entity, selfProvider, s[len(scriptStart):len(s)-len(scriptEnd)])
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
	result := resolveScript(entity, selfProvider, text)
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
	result := resolveScript(entity, selfProvider, text)
	w, err := fxp.WeightFromString(result, defUnits)
	if err != nil {
		slog.Error("unable to resolve script result to a weight", "result", result, "script", text)
		return 0
	}
	return w
}

func resolveScript(entity *Entity, selfProvider ScriptSelfProvider, text string) string {
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
	args := []ScriptArg{{Name: "entity", Value: newScriptEntity(entity)}}
	if selfProvider.Provider != nil {
		args = append(args, ScriptArg{Name: "self", Value: selfProvider.Provider})
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
					Value: func() any { return newScriptAttribute(attr) },
				})
			}
		}
	}
	if v, err := RunScript(fxp.SecondsToDuration(maxTime), text, args...); err != nil {
		var interruptedErr *goja.InterruptedError
		if errors.As(err, &interruptedErr) {
			result = fmt.Sprintf(i18n.Text("script execution timed out (limited to %v seconds)"), maxTime)
		} else {
			result = err.Error()
		}
	} else {
		if attr, ok := v.Export().(*scriptAttribute); ok {
			result = fmt.Sprintf("%v", attr.ValueOf())
		} else {
			result = v.String()
		}
	}
	resolveCache[key] = result
	return result
}

// RunScript compiles and runs a script with the provided arguments. A timeout of 0 or less means no timeout.
// The script should be a valid JavaScript function body, and it will be wrapped in an anonymous function to avoid
// polluting the global scope. The arguments will be set as global variables in the script's context. The return value
// is the result of the script execution, or an error if it fails.
func RunScript(timeout time.Duration, text string, args ...ScriptArg) (goja.Value, error) {
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
		if valueProvider, ok2 := arg.Value.(func() any); ok2 {
			var cachedResult goja.Value
			if err := globals.DefineAccessorProperty(arg.Name, vm.ToValue(func(_ goja.FunctionCall) goja.Value {
				if cachedResult == nil {
					cachedResult = vm.ToValue(valueProvider())
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
