package gurps

import (
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/tid"
)

type scriptEquipment struct {
	equipment              *Equipment
	ID                     tid.TID
	ParentID               tid.TID
	Name                   string
	TechLevel              string
	LegalityClass          string
	children               []*scriptEquipment
	Tags                   []string
	Quantity               float64
	value                  float64
	extendedValue          float64
	weight                 float64
	extendedWeight         float64
	Level                  float64
	Uses                   int
	MaxUses                int
	WeightIgnoredForSkills bool
	Equipped               bool
	Container              bool
	HasChildren            bool
	cachedChildren         bool
	cachedValue            bool
	cachedExtendedValue    bool
	cachedWeight           bool
	cachedExtendedWeight   bool
}

// DeferredNewScriptEquipment creates a deferred ScriptSelfProvider for the given Equipment.
func DeferredNewScriptEquipment(equipment *Equipment) ScriptSelfProvider {
	if equipment == nil {
		return ScriptSelfProvider{}
	}
	return ScriptSelfProvider{
		ID:       equipment.TID,
		Provider: func() any { return newScriptEquipment(equipment) },
	}
}

func newScriptEquipment(equipment *Equipment) *scriptEquipment {
	var parentID tid.TID
	if equipment.parent != nil {
		parentID = equipment.parent.TID
	}
	return &scriptEquipment{
		equipment:              equipment,
		ID:                     equipment.TID,
		ParentID:               parentID,
		Name:                   equipment.NameWithReplacements(),
		TechLevel:              equipment.TechLevel,
		LegalityClass:          equipment.LegalityClass,
		Tags:                   slices.Clone(equipment.Tags),
		Quantity:               fxp.As[float64](equipment.Quantity),
		Level:                  fxp.As[float64](equipment.Level),
		Uses:                   equipment.Uses,
		MaxUses:                equipment.MaxUses,
		WeightIgnoredForSkills: equipment.WeightIgnoredForSkills,
		Equipped:               equipment.Equipped,
		Container:              equipment.Container(),
		HasChildren:            equipment.HasChildren(),
	}
}

func (e *scriptEquipment) Value() float64 {
	if !e.cachedValue {
		e.cachedValue = true
		e.value = fxp.As[float64](e.equipment.AdjustedValue())
	}
	return e.value
}

func (e *scriptEquipment) ExtendedValue() float64 {
	if !e.cachedExtendedValue {
		e.cachedExtendedValue = true
		e.extendedValue = fxp.As[float64](e.equipment.ExtendedValue())
	}
	return e.extendedValue
}

func (e *scriptEquipment) Weight() float64 {
	if !e.cachedWeight {
		e.cachedWeight = true
		e.weight = fxp.As[float64](fxp.Int(e.equipment.AdjustedWeight(false, fxp.Pound)))
	}
	return e.weight
}

func (e *scriptEquipment) ExtendedWeight() float64 {
	if !e.cachedExtendedWeight {
		e.cachedExtendedWeight = true
		e.extendedWeight = fxp.As[float64](fxp.Int(e.equipment.ExtendedWeight(false, fxp.Pound)))
	}
	return e.extendedWeight
}

func (e *scriptEquipment) Children() []*scriptEquipment {
	if !e.cachedChildren {
		e.cachedChildren = true
		if len(e.equipment.Children) != 0 {
			e.children = make([]*scriptEquipment, 0, len(e.equipment.Children))
			for _, child := range e.equipment.Children {
				if child.Quantity > 0 {
					e.children = append(e.children, newScriptEquipment(child))
				}
			}
		}
	}
	return e.children
}

func (e *scriptEquipment) Find(name, tag string) []*scriptEquipment {
	if !e.equipment.Container() {
		return nil
	}
	return findScriptEquipment(name, tag, e.equipment.Children...)
}

func findScriptEquipment(name, tag string, topLevelItems ...*Equipment) []*scriptEquipment {
	var items []*scriptEquipment
	Traverse(func(item *Equipment) bool {
		if item.Quantity > 0 {
			parent := item.parent
			for parent != nil {
				if parent.Quantity <= 0 {
					return false
				}
				parent = parent.parent
			}
			if (name == "" || strings.EqualFold(item.NameWithReplacements(), name)) && matchTag(tag, item.Tags) {
				items = append(items, newScriptEquipment(item))
			}
		}
		return false
	}, true, false, topLevelItems...)
	return items
}

func matchTag(tag string, tags []string) bool {
	if tag == "" {
		return true
	}
	for _, t := range tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}
