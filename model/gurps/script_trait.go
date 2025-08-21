// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/toolbox/v2/tid"
)

type scriptTrait struct {
	trait          *Trait
	ID             tid.TID
	ParentID       tid.TID
	Name           string
	Notes          string
	Kind           string
	Levels         *float64
	children       []*scriptTrait
	Tags           []string
	Container      bool
	HasChildren    bool
	cachedChildren bool
}

// DeferredNewScriptTrait creates a deferred ScriptSelfProvider for the given Equipment.
func DeferredNewScriptTrait(trait *Trait) ScriptSelfProvider {
	if trait == nil {
		return ScriptSelfProvider{}
	}
	return ScriptSelfProvider{
		ID:       trait.TID,
		Provider: func() any { return newScriptTrait(trait) },
	}
}

func newScriptTrait(trait *Trait) *scriptTrait {
	var parentID tid.TID
	if trait.parent != nil {
		parentID = trait.parent.TID
	}
	t := scriptTrait{
		trait:       trait,
		ID:          trait.TID,
		ParentID:    parentID,
		Name:        trait.NameWithReplacements(),
		Notes:       trait.SecondaryText(func(_ display.Option) bool { return true }),
		Tags:        slices.Clone(trait.Tags),
		Container:   trait.Container(),
		HasChildren: trait.HasChildren(),
	}
	if trait.Container() {
		t.Kind = strings.ReplaceAll(trait.ContainerType.Key(), "_", " ")
	} else if trait.CanLevel {
		levels := fxp.AsFloat[float64](trait.Levels)
		t.Levels = &levels
	}
	return &t
}

func (t *scriptTrait) Children() []*scriptTrait {
	if !t.cachedChildren {
		t.cachedChildren = true
		if len(t.trait.Children) != 0 {
			t.children = make([]*scriptTrait, 0, len(t.trait.Children))
			for _, child := range t.trait.Children {
				if child.Enabled() {
					t.children = append(t.children, newScriptTrait(child))
				}
			}
		}
	}
	return t.children
}

func (t *scriptTrait) Find(name, tag string) []*scriptTrait {
	if !t.trait.Container() {
		return nil
	}
	return findScriptTraits(name, tag, t.trait.Children...)
}

func findScriptTraits(name, tag string, topLevelTraits ...*Trait) []*scriptTrait {
	var traits []*scriptTrait
	Traverse(func(trait *Trait) bool {
		if (name == "" || strings.EqualFold(trait.NameWithReplacements(), name)) && matchTag(tag, trait.Tags) {
			traits = append(traits, newScriptTrait(trait))
		}
		return false
	}, true, false, topLevelTraits...)
	return traits
}
