package gurps

import (
	"context"
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
)

const scriptPrefix = "^^^"

var (
	evalEmbeddedRegex = regexp.MustCompile(`\|\|[^|]+\|\|`)
	scriptCache       = make(map[string]*goja.Program)
	scriptCacheLock   sync.RWMutex
	vmPool            = sync.Pool{New: func() any {
		vm := goja.New()
		vm.SetFieldNameMapper(goja.UncapFieldNameMapper())
		vm.SetParserOptions(parser.WithDisableSourceMaps)

		mustSet(vm, "printf", fmt.Printf)

		mustSet(vm, "fixedIntFromString", fxp.FromStringForced)
		mustSet(vm, "fixedIntFromInt", fxp.From[int])
		mustSet(vm, "fixedIntFromFloat", fxp.From[float64])
		mustSet(vm, "fixedIntAsInt", fxp.As[int])
		mustSet(vm, "fixedIntAsFloat", fxp.As[float64])
		mustSet(vm, "fixedIntApplyRounding", fxp.ApplyRounding)
		mustSet(vm, "fixedIntAdd", addFixedInts)
		mustSet(vm, "fixedIntSub", subFixedInts)

		mustSet(vm, "feetAndInches", fxp.FeetAndInches)
		mustSet(vm, "inch", fxp.Inch)
		mustSet(vm, "feet", fxp.Feet)
		mustSet(vm, "yard", fxp.Yard)
		mustSet(vm, "mile", fxp.Mile)
		mustSet(vm, "centimeter", fxp.Centimeter)
		mustSet(vm, "kilometer", fxp.Kilometer)
		mustSet(vm, "meter", fxp.Meter)
		mustSet(vm, "extractLengthUnit", fxp.ExtractLengthUnit)
		mustSet(vm, "lengthFromInteger", fxp.LengthFromInteger[int])
		mustSet(vm, "lengthFromString", fxp.LengthFromStringForced)

		mustSet(vm, "ssrt", SSRT)
		mustSet(vm, "yardsFromSSRT", YardsFromSSRT)

		return vm
	}}
)

// ScriptArg is a named argument to be passed to RunString.
type ScriptArg struct {
	Name  string
	Value any
}

type entityScope struct {
	entity *Entity
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
	es := &entityScope{entity: entity}
	v, err := RunScript(GlobalSettings().PermittedPerScriptExecTime, text[len(scriptPrefix):],
		ScriptArg{Name: "hasTrait", Value: es.HasTrait},
		ScriptArg{Name: "traitLevel", Value: es.TraitLevel},
		ScriptArg{Name: "skillLevel", Value: es.SkillLevel},
		ScriptArg{Name: "attributeIDs", Value: es.AttributeIDs},
		ScriptArg{Name: "currentAttributeValue", Value: es.CurrentAttributeValue},
	)
	if err != nil {
		return err.Error()
	}
	return v.String()
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

func addFixedInts(a, b fxp.Int) fxp.Int {
	return a + b
}

func subFixedInts(a, b fxp.Int) fxp.Int {
	return a - b
}

// HasTrait checks if the entity has a trait with the given name.
func (e *entityScope) HasTrait(traitName string) bool {
	if e.entity == nil {
		return false
	}
	found := false
	Traverse(func(t *Trait) bool {
		if strings.EqualFold(t.NameWithReplacements(), traitName) {
			found = true
			return true
		}
		return false
	}, true, false, e.entity.Traits...)
	return found
}

// TraitLevel returns the level of the trait with the given name, or -1 if not found or not leveled.
func (e *entityScope) TraitLevel(traitName string) fxp.Int {
	if e.entity == nil {
		return -fxp.One
	}
	levels := -fxp.One
	Traverse(func(t *Trait) bool {
		if strings.EqualFold(t.NameWithReplacements(), traitName) {
			if t.IsLeveled() {
				if levels == -fxp.One {
					levels = t.Levels
				} else {
					levels += t.Levels
				}
			}
		}
		return false
	}, true, false, e.entity.Traits...)
	return levels
}

// SkillLevel returns the level of the skill with the given name, or 0 if not found.
func (e *entityScope) SkillLevel(name, specialization string, relative bool) int {
	if e.entity == nil {
		return 0
	}
	if e.entity.isSkillLevelResolutionExcluded(name, specialization) {
		return 0
	}
	e.entity.registerSkillLevelResolutionExclusion(name, specialization)
	defer e.entity.unregisterSkillLevelResolutionExclusion(name, specialization)
	var level int
	Traverse(func(s *Skill) bool {
		if strings.EqualFold(s.NameWithReplacements(), name) &&
			strings.EqualFold(s.SpecializationWithReplacements(), specialization) {
			s.UpdateLevel()
			if relative {
				level = fxp.As[int](s.LevelData.RelativeLevel)
			} else {
				level = fxp.As[int](s.LevelData.Level)
			}
			return true
		}
		return false
	}, true, true, e.entity.Skills...)
	return level
}

// AttributeIDs returns a list of available attribute IDs.
func (e *entityScope) AttributeIDs() []string {
	if e.entity == nil {
		return nil
	}
	attrs := e.entity.Attributes.List()
	ids := make([]string, 0, len(attrs))
	for _, attr := range attrs {
		if def := attr.AttributeDef(); def != nil {
			if def.IsSeparator() {
				continue
			}
			ids = append(ids, def.DefID)
		}
	}
	return ids
}

// CurrentAttributeValue resolves the given attribute ID to its current value.
func (e *entityScope) CurrentAttributeValue(attrID string) (value fxp.Int, exists bool) {
	if e.entity == nil {
		return 0, false
	}
	if value = e.entity.Attributes.Current(attrID); value == fxp.Min {
		return 0, false
	}
	return value, true
}

// SSRT is a function that takes a length and converts it to a value from the Size and Speed/Range table.
func SSRT(length fxp.Length, forSize bool) int {
	result := yardsToValue(length, forSize)
	if !forSize {
		result = -result
	}
	return result
}

// YardsFromSSRT converts a value from the Size and Speed/Range table to a length in yards.
func YardsFromSSRT(ssrtValue int) fxp.Int {
	return valueToYards(ssrtValue)
}
