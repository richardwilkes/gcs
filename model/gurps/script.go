package gurps

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja/parser"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/eval"
	"github.com/richardwilkes/toolbox/i18n"
)

const scriptPrefix = "^^^"

// The evaluator operators and functions that will be used when calling newEvaluator().
var (
	EvalOperators = eval.FixedOperators[fxp.DP](true)
	EvalFuncs     = eval.FixedFunctions[fxp.DP]()
)

var (
	evalEmbeddedRegex  = regexp.MustCompile(`\|\|[^|]+\|\|`)
	scriptCache        = make(map[string]*goja.Program)
	globalResolveCache = make(map[string]string)
	vmPool             = sync.Pool{New: func() any {
		vm := goja.New()
		vm.SetFieldNameMapper(scriptNameMapper{})
		vm.SetParserOptions(parser.WithDisableSourceMaps)
		mustSet(vm, "console", &scriptConsole{})
		mustSet(vm, "fixed", newScriptFixed())
		mustSet(vm, "length", newScriptLength())
		mustSet(vm, "weight", newScriptWeight())
		mustSet(vm, "ssrt", &scriptSSRT{})
		return vm
	}}
)

// ScriptArg is a named argument to be passed to RunString.
type ScriptArg struct {
	Name  string
	Value any
}

func mustSet(vm *goja.Runtime, name string, value any) {
	if err := vm.Set(name, value); err != nil {
		panic(errs.Newf("failed to set %s: %s", name, err.Error()))
	}
}

// ResolveText will process the text as a script if it starts with ^^^. If it does not, it will look for embedded
// expressions inside || pairs inside the text and evaluate them.
func ResolveText(entity *Entity, text string) string {
	if !strings.HasPrefix(text, scriptPrefix) {
		return evalEmbeddedRegex.ReplaceAllStringFunc(text, func(s string) string {
			if entity == nil {
				return s
			}
			exp := s[2 : len(s)-2]
			result, err := newEvaluator(entity).Evaluate(exp)
			if err != nil {
				slog.Error("expression evaluation failed", "expression", exp, "error", err)
			}
			return fmt.Sprintf("%v", result)
		})
	}
	return resolveScript(entity, text[len(scriptPrefix):])
}

// ResolveToNumber evaluates the text as a script if it starts with ^^^, or evaluates it as an expression otherwise.
func ResolveToNumber(entity *Entity, text string) fxp.Int {
	if !strings.HasPrefix(text, scriptPrefix) {
		result, err := newEvaluator(entity).Evaluate(text)
		if err != nil {
			slog.Error("expression evaluation failed", "expression", text, "error", err)
			return 0
		}
		if value, ok := result.(fxp.Int); ok {
			return value
		}
		if str, ok := result.(string); ok {
			var value fxp.Int
			if value, err = fxp.FromString(str); err == nil {
				return value
			}
		}
		slog.Error("unable to resolve expression to a number", "expression", text, "result", result, "error", err)
		return 0
	}
	text = resolveScript(entity, text[len(scriptPrefix):])
	value, err := fxp.FromString(text)
	if err != nil {
		slog.Error("unable to resolve script result to a number", "result", text)
		return 0
	}
	return value
}

func resolveScript(entity *Entity, text string) string {
	var resolveCache map[string]string
	if entity == nil {
		resolveCache = globalResolveCache
	} else {
		resolveCache = entity.scriptCache
	}
	if cached, exists := resolveCache[text]; exists {
		return cached
	}
	var result string
	maxTime := GlobalSettings().General.PermittedPerScriptExecTime
	if v, err := RunScript(fxp.SecondsToDuration(maxTime), text,
		ScriptArg{Name: "entity", Value: &scriptEntity{entity: entity}}); err != nil {
		var interruptedErr *goja.InterruptedError
		if errors.As(err, &interruptedErr) {
			result = fmt.Sprintf(i18n.Text("script execution timed out (limited to %v seconds)"), maxTime)
		} else {
			result = err.Error()
		}
	} else {
		result = v.String()
	}
	resolveCache[text] = result
	return result
}

// RunScript compiles and runs a script with the provided arguments. A timeout of 0 or less means no timeout.
// The script should be a valid JavaScript function body, and it will be wrapped in an anonymous function to avoid
// polluting the global scope. The arguments will be set as global variables in the script's context. The return value
// is the result of the script execution, or an error if it fails.
func RunScript(timeout time.Duration, text string, args ...ScriptArg) (goja.Value, error) {
	program, exists := scriptCache[text]
	if !exists {
		var err error
		program, err = goja.Compile("", "(function() {"+text+"})();", true)
		if err != nil {
			return nil, fmt.Errorf("failed to compile script: %w", err)
		}
		scriptCache[text] = program
	}
	vm, ok := vmPool.Get().(*goja.Runtime)
	if !ok {
		return nil, errors.New("failed to get VM from pool")
	}
	defer func() {
		globals := vm.GlobalObject()
		for _, arg := range args {
			if err := globals.Delete(arg.Name); err != nil {
				errs.LogWithLevel(context.Background(), slog.LevelWarn, nil, err, "name", arg.Name)
			}
		}
		vm.ClearInterrupt()
		vmPool.Put(vm)
	}()
	for _, arg := range args {
		if err := vm.Set(arg.Name, arg.Value); err != nil {
			return nil, fmt.Errorf("failed to set argument %q: %w", arg.Name, err)
		}
	}
	if timeout > 0 {
		defer time.AfterFunc(timeout, func() { vm.Interrupt("timeout") }).Stop()
	}
	return vm.RunProgram(program)
}

func newEvaluator(resolver eval.VariableResolver) *eval.Evaluator {
	return &eval.Evaluator{
		Resolver:  resolver,
		Operators: EvalOperators,
		Functions: EvalFuncs,
	}
}
