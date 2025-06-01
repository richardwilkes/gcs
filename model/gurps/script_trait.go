package gurps

import (
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/json"
)

type scriptTrait struct {
	Name     string         `json:"name"`
	Kind     string         `json:"kind"`
	Levels   *float64       `json:"levels,omitempty"`
	Children []*scriptTrait `json:"children,omitempty"`
	Tags     []string       `json:"tags,omitempty"`
}

func newScriptTrait(trait *Trait, includeEnabledChildren bool) *scriptTrait {
	t := scriptTrait{
		Name: trait.NameWithReplacements(),
		Tags: slices.Clone(trait.Tags),
	}
	if trait.Container() {
		t.Kind = strings.ReplaceAll(trait.ContainerType.Key(), "_", " ")
		if includeEnabledChildren {
			children := trait.NodeChildren()
			t.Children = make([]*scriptTrait, 0, len(children))
			for _, child := range children {
				if child.Enabled() {
					t.Children = append(t.Children, newScriptTrait(child, true))
				}
			}
		}
	} else {
		t.Kind = "trait"
		if trait.CanLevel {
			levels := fxp.As[float64](trait.Levels)
			t.Levels = &levels
		}
	}
	return &t
}

func (t *scriptTrait) String() string {
	data, err := json.Marshal(t)
	if err != nil {
		return err.Error()
	}
	return string(data)
}
