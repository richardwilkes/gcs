package gurps

import (
	"slices"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/json"
)

type scriptEquipment struct {
	Name           string             `json:"name"`
	Children       []*scriptEquipment `json:"children,omitempty"`
	Tags           []string           `json:"tags,omitempty"`
	Quantity       fxp.Int            `json:"quantity"`
	Value          fxp.Int            `json:"value"`
	ExtendedValue  fxp.Int            `json:"extendedValue"`
	Weight         fxp.Weight         `json:"weight"`
	ExtendedWeight fxp.Weight         `json:"extendedWeight"`
	Level          fxp.Int            `json:"level,omitempty"`
	Uses           int                `json:"uses,omitempty"`
	Equipped       bool               `json:"equipped,omitempty"`
}

func newScriptEquipment(entity *Entity, equipment *Equipment, includeChildren bool) *scriptEquipment {
	defUnits := SheetSettingsFor(entity).DefaultWeightUnits
	e := scriptEquipment{
		Name:           equipment.NameWithReplacements(),
		Tags:           slices.Clone(equipment.Tags),
		Quantity:       equipment.Quantity,
		Value:          equipment.AdjustedValue(),
		ExtendedValue:  equipment.ExtendedValue(),
		Weight:         equipment.AdjustedWeight(false, defUnits),
		ExtendedWeight: equipment.ExtendedWeight(false, defUnits),
		Level:          equipment.Level,
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

func (t *scriptEquipment) String() string {
	data, err := json.Marshal(t)
	if err != nil {
		return err.Error()
	}
	return string(data)
}
