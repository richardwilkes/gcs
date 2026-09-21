// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestHasTemplatePickerData verifies that a node carrying template choices is spotted
func TestHasTemplatePickerData(t *testing.T) {
	c := check.New(t)
	plain := NewTrait(nil, nil, false)
	plain.Name = "Claws"
	c.False(HasTemplatePickerData(plain), "a plain trait carries no choices")

	emptyContainer := NewTrait(nil, nil, true)
	emptyContainer.Name = "Advantages"
	c.False(HasTemplatePickerData(emptyContainer), "a container without choices carries none")

	choices := newTemplateChoiceTrait("Pick One", "First", "Second")
	c.True(HasTemplatePickerData(plain, choices))

	emptyContainer.Children = []*Trait{choices}
	choices.SetParent(emptyContainer)
	c.True(HasTemplatePickerData(emptyContainer), "choices nested deeper must be found as well")

	skillContainer := NewSkill(nil, nil, true)
	skillContainer.Name = "Techniques"
	c.False(HasTemplatePickerData(skillContainer))

	skillContainer.TemplatePicker.Type = picker.Points
	c.True(HasTemplatePickerData(skillContainer), "skills carry choices too")
}

// TestClearTemplatePickerData verifies that the choices are removed from every container beneath the rows as well, since
// a copied container brings its whole subtree with it.
func TestClearTemplatePickerData(t *testing.T) {
	c := check.New(t)
	outer := NewTrait(nil, nil, true)
	outer.Name = "Advantages"
	inner := newTemplateChoiceTrait("Pick One", "First", "Second")
	inner.SetParent(outer)
	outer.Children = []*Trait{inner}

	// Verify the test data has picker data
	c.True(HasTemplatePickerData(outer))

	// Clear any template picker data
	ClearTemplatePickerData(outer)

	// Verify the test data no longer carries picker data
	c.False(HasTemplatePickerData(outer), "no picker data may be left")

	// Verify we still have the same data otherwise (this could use more checks)
	c.Equal(2, len(inner.Children), "clearing must not disturb anything else")
}

// newTemplateChoiceTrait returns a trait container carrying template choices, holding a child for each of the given
// names.
func newTemplateChoiceTrait(name string, childNames ...string) *Trait {
	container := NewTrait(nil, nil, true)
	container.Name = name
	container.TemplatePicker.Type = picker.Count
	container.TemplatePicker.Qualifier.Compare = criteria.EqualsNumber
	container.TemplatePicker.Qualifier.Qualifier = fxp.One
	children := make([]*Trait, 0, len(childNames))
	for _, childName := range childNames {
		child := NewTrait(nil, container, false)
		child.Name = childName
		children = append(children, child)
	}
	container.Children = children
	return container
}
