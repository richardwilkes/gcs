package gurps

import (
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
)

type scriptTrait struct {
	Name     string         `json:"name"`
	Kind     string         `json:"kind"`
	Levels   *float64       `json:"levels,omitempty"`
	Children []*scriptTrait `json:"children,omitempty"`
	Tags     []string       `json:"tags,omitempty"`
}

func deferredNewScriptTrait(trait *Trait) ScriptSelfProvider {
	if trait == nil {
		return ScriptSelfProvider{}
	}
	return ScriptSelfProvider{
		ID:       trait.TID,
		Provider: func() any { return newScriptTrait(trait, true) },
	}
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
