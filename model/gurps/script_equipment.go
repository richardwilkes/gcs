package gurps

import (
	"slices"

	"github.com/richardwilkes/gcs/v5/model/fxp"
)

type scriptEquipment struct {
	Name           string             `json:"name"`
	Children       []*scriptEquipment `json:"children,omitempty"`
	Tags           []string           `json:"tags,omitempty"`
	Quantity       float64            `json:"quantity"`
	Value          float64            `json:"value"`
	ExtendedValue  float64            `json:"extendedValue"`
	Weight         float64            `json:"weight"`
	ExtendedWeight float64            `json:"extendedWeight"`
	Level          float64            `json:"level,omitempty"`
	Uses           int                `json:"uses,omitempty"`
	Equipped       bool               `json:"equipped,omitempty"`
}

func deferredNewScriptEquipment(entity *Entity, equipment *Equipment) func() any {
	if equipment == nil {
		return nil
	}
	return func() any {
		return newScriptEquipment(entity, equipment, true)
	}
}

func newScriptEquipment(entity *Entity, equipment *Equipment, includeChildren bool) *scriptEquipment {
	defUnits := SheetSettingsFor(entity).DefaultWeightUnits
	e := scriptEquipment{
		Name:           equipment.NameWithReplacements(),
		Tags:           slices.Clone(equipment.Tags),
		Quantity:       fxp.As[float64](equipment.Quantity),
		Value:          fxp.As[float64](equipment.AdjustedValue()),
		ExtendedValue:  fxp.As[float64](equipment.ExtendedValue()),
		Weight:         fxp.As[float64](fxp.Int(equipment.AdjustedWeight(false, defUnits))),
		ExtendedWeight: fxp.As[float64](fxp.Int(equipment.ExtendedWeight(false, defUnits))),
		Level:          fxp.As[float64](equipment.Level),
		Uses:           equipment.Uses,
		Equipped:       equipment.Equipped,
	}
	if includeChildren {
		children := equipment.NodeChildren()
		e.Children = make([]*scriptEquipment, 0, len(children))
		for _, child := range children {
			e.Children = append(e.Children, newScriptEquipment(entity, child, true))
		}
	}
	return &e
}
