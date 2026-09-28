// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"fmt"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/promptstep"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestPromptOperationTitle verifies that a prompt is titled with its operation's name and its step, and with whichever
// of them it has when it lacks the other, and that setting the step leaves the operation it was set on alone.
func TestPromptOperationTitle(t *testing.T) {
	c := check.New(t)
	op := promptOperation{name: "Apply Template", description: "Applying template Knight to Sir Bob"}
	c.Equal("Apply Template: Modifiers", op.at(promptstep.Modifiers).title())
	c.Equal("Apply Template: Remove Choices", op.at(promptstep.RemoveChoices).title())
	c.Equal("Apply Template", op.title(), "setting the step must not change the operation it was set on")
	c.Equal("Substitutions", promptOperation{}.at(promptstep.Substitutions).title())
	c.Equal("", promptOperation{}.title(), "the zero value must have no step")
}

// TestDescribeRows verifies that a lone row is named and that several are counted by their kind.
func TestDescribeRows(t *testing.T) {
	c := check.New(t)
	one := gurps.NewTrait(nil, nil, false)
	one.Name = "Luck"
	c.Equal("Luck", describeRows([]*gurps.Trait{one}))
	c.Equal("2 traits", describeRows([]*gurps.Trait{one, gurps.NewTrait(nil, nil, false)}))
	c.Equal("2 pieces of equipment",
		describeRows([]*gurps.Equipment{gurps.NewEquipment(nil, nil, false), gurps.NewEquipment(nil, nil, false)}))
	c.Equal("2 modifiers", describeRows([]*gurps.TraitModifier{
		gurps.NewTraitModifier(nil, nil, false),
		gurps.NewTraitModifier(nil, nil, false),
	}))
}

// TestRowLocation verifies that a row is placed by its kind and the containers above it, outermost first, that a
// top-level row isn't placed at all, and that a long run of containers keeps its outermost and innermost ends.
func TestRowLocation(t *testing.T) {
	c := check.New(t)
	top := gurps.NewTrait(nil, nil, false)
	c.Equal("", rowLocation(top), "a top-level row has nothing to say about where it is")

	parent := func(name string, of *gurps.Trait) *gurps.Trait {
		container := gurps.NewTrait(nil, of, true)
		container.Name = name
		return container
	}
	outer := parent("Knight", nil)
	inner := parent("Class Advantages", outer)
	row := gurps.NewTrait(nil, inner, false)
	c.Equal("Trait in Knight › Class Advantages", rowLocation(row))

	container := outer
	for i := range 6 {
		container = parent(fmt.Sprintf("A Fairly Long Container Name %d", i), container)
	}
	deep := gurps.NewTrait(nil, container, false)
	c.Equal("Trait in Knight › … › A Fairly Long Container Name 5", rowLocation(deep),
		"the containers in the middle must be left out, keeping both ends")
}

// TestNameList verifies that up to ten names are listed one per line, that more than that ends in a count of the rest
// in the tenth line, and that each name is cut down.
func TestNameList(t *testing.T) {
	c := check.New(t)
	names := func(count int) []string {
		list := make([]string, count)
		for i := range list {
			list[i] = fmt.Sprintf("Name %d", i+1)
		}
		return list
	}
	c.Equal("Name 1\nName 2", nameList(names(2)))
	lines := strings.Split(nameList(names(10)), "\n")
	c.Equal(10, len(lines), "ten names must all be shown")
	c.Equal("Name 10", lines[9])
	lines = strings.Split(nameList(names(12)), "\n")
	c.Equal(10, len(lines), "more than ten must still take ten lines")
	c.Equal("Name 9", lines[8])
	c.Equal("and 3 more", lines[9], "the last line must count every name not shown")
	long := strings.Repeat("x", maxNameLength+10)
	c.Equal(strings.Repeat("x", maxNameLength)+"…", nameList([]string{long}), "a long name must be cut down")
}

// TestShortNames verifies that names are cut down for use in a description, and that short ones are left alone.
func TestShortNames(t *testing.T) {
	c := check.New(t)
	long := strings.Repeat("x", maxNameLength+10)
	short := shortNames("Knight", long)
	c.Equal("Knight", short[0])
	c.Equal(strings.Repeat("x", maxNameLength)+"…", short[1])
	c.Equal("Knight, Elf", joinNames([]string{"Knight", "Elf"}))
}
