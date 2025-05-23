package gurps

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/goblin"
	"github.com/richardwilkes/goblin/ast"
)

const scriptPrefix = "^^^"

var (
	evalEmbeddedRegex = regexp.MustCompile(`\|\|[^|]+\|\|`)
	topScope          = createTopScope()
)

type entityScope struct {
	entity *Entity
}

func createTopScope() ast.Scope {
	scope := goblin.NewScope()

	scope.Define("sprintf", fmt.Sprintf)

	scope.DefineType("FixedInt", fxp.One)
	scope.Define("FixedIntFromString", fxp.FromStringForced)
	scope.Define("FixedIntFromInt", fxp.From[int])
	scope.Define("FixedIntFromFloat", fxp.From[float64])
	scope.Define("FixedIntAsInt", fxp.As[int])
	scope.Define("FixedIntAsFloat", fxp.As[float64])
	scope.Define("ApplyRounding", fxp.ApplyRounding)

	scope.DefineType("Length", fxp.Length(0))
	scope.DefineType("LengthUnit", fxp.Yard)
	scope.DefineGlobal("FeetAndInches", fxp.FeetAndInches)
	scope.DefineGlobal("Inch", fxp.Inch)
	scope.DefineGlobal("Feet", fxp.Feet)
	scope.DefineGlobal("Yard", fxp.Yard)
	scope.DefineGlobal("Mile", fxp.Mile)
	scope.DefineGlobal("Centimeter", fxp.Centimeter)
	scope.DefineGlobal("Kilometer", fxp.Kilometer)
	scope.DefineGlobal("Meter", fxp.Meter)
	scope.Define("ExtractLengthUnit", fxp.ExtractLengthUnit)
	scope.Define("LengthFromInteger", fxp.LengthFromInteger[int])
	scope.Define("LengthFromString", fxp.LengthFromStringForced)

	scope.Define("SSRT", SSRT)
	scope.Define("YardsFromSSRT", YardsFromSSRT)

	return scope
}

// ResolveText will process the text as a script if it starts with ^^^. If it does not, it will look for embedded
// expressions inside || pairs inside the text and evaluate them.
func ResolveText(entity *Entity, text string) string {
	if !strings.HasPrefix(text, scriptPrefix) {
		return evalEmbeddedRegex.ReplaceAllStringFunc(text, entity.EmbeddedEval)
	}

	es := &entityScope{entity: entity}
	scope := topScope.NewScope()
	defer scope.Destroy()

	scope.Define("HasTrait", es.HasTrait)
	scope.Define("TraitLevel", es.TraitLevel)

	scope.Define("SkillLevel", es.SkillLevel)

	scope.Define("AttributeIDs", es.AttributeIDs)
	scope.Define("CurrentAttributeValue", es.CurrentAttributeValue)

	v, err := scope.ParseAndRunWithTimeout(GlobalSettings().PermittedPerScriptExecTime, text[len(scriptPrefix):])
	if err != nil {
		return err.Error()
	}
	return v.String()
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
