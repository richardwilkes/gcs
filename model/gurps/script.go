package gurps

import (
	"context"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja/parser"
	"github.com/richardwilkes/toolbox/errs"
)

const scriptPrefix = "^^^"

var (
	evalEmbeddedRegex  = regexp.MustCompile(`\|\|[^|]+\|\|`)
	scriptCacheLock    sync.RWMutex
	scriptCache        = make(map[string]*goja.Program)
	globalResolveCache = NewScriptResolveCache()
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
		return evalEmbeddedRegex.ReplaceAllStringFunc(text, entity.EmbeddedEval)
	}
	var resolveCache *ScriptResolveCache
	if entity == nil {
		resolveCache = globalResolveCache
	} else {
		resolveCache = entity.resolveCache
	}
	if cached, exists := resolveCache.Get(text); exists {
		return cached
	}
	var result string
	if v, err := RunScript(GlobalSettings().PermittedPerScriptExecTime, text[len(scriptPrefix):],
		ScriptArg{Name: "entity", Value: &scriptEntity{entity: entity}},
	); err != nil {
		result = err.Error()
	} else {
		result = v.String()
	}
	resolveCache.Set(text, result)
	return result
}

// RunScript compiles and runs a script with the provided arguments. A timeout of 0 or less means no timeout.
// The script should be a valid JavaScript function body, and it will be wrapped in an anonymous function to avoid
// polluting the global scope. The arguments will be set as global variables in the script's context. The return value
// is the result of the script execution, or an error if it fails.
func RunScript(timeout time.Duration, text string, args ...ScriptArg) (goja.Value, error) {
	scriptCacheLock.RLock()
	program, exists := scriptCache[text]
	scriptCacheLock.RUnlock()
	if !exists {
		var err error
		program, err = goja.Compile("", "(function() {"+text+"})();", true)
		if err != nil {
			return nil, errs.New("failed to compile script: " + err.Error())
		}
		scriptCacheLock.Lock()
		scriptCache[text] = program
		scriptCacheLock.Unlock()
	}
	vm, ok := vmPool.Get().(*goja.Runtime)
	if !ok {
		return nil, errs.New("failed to get VM from pool")
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
			return nil, errs.Newf("failed to set argument %s: %s", arg.Name, err.Error())
		}
	}
	if timeout > 0 {
		defer time.AfterFunc(timeout, func() { vm.Interrupt("timeout") }).Stop()
	}
	return vm.RunProgram(program)
}
