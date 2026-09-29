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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison/accessibility"
)

// The spoken names must be updated by Sync, not only by drawing, since a block scrolled out of view is not drawn.
func TestEncumbrancePanelSpeaksTheCurrentLevel(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	entity.Recalculate()
	panel := NewEncumbrancePanel(entity)
	c.Equal(len(encumbrance.Levels), len(panel.row))
	c.Equal(len(encumbrance.Levels), len(panel.levels))
	spoken := func() (levels, names []string) {
		for i, name := range panel.row {
			levels = append(levels, panel.levels[i].Accessibility.Name)
			names = append(names, name.AsPanel().Accessibility.Name)
		}
		return levels, names
	}
	var syncer Syncer = panel
	syncer.Sync()
	c.Equal(0, panel.current, "a new character carries nothing")
	levels, names := spoken()
	c.Equal("Current Encumbrance Level 0", levels[0])
	c.Equal("Current Encumbrance Level None", names[0])
	for i := 1; i < len(levels); i++ {
		c.Equal("", levels[i], "a row that is not current speaks its own text")
		c.Equal("", names[i], "a row that is not current speaks its own text")
	}

	// Basic lift at ST 10 is 20 lb, so 30 lb is Light.
	eq := gurps.NewEquipment(entity, nil, false)
	eq.BaseWeight = "30 lb"
	entity.CarriedEquipment = append(entity.CarriedEquipment, eq)
	entity.Recalculate()
	panel.Sync()
	c.Equal(1, panel.current)
	c.False(panel.overloaded)
	levels, names = spoken()
	c.Equal("Current Encumbrance Level 1", levels[1])
	c.Equal("Current Encumbrance Level Light", names[1])
	c.Equal("", levels[0], "the row that was current speaks its own text again")
	c.Equal("", names[0], "the row that was current speaks its own text again")

	// The extra-heavy limit at ST 10 is 200 lb.
	eq.BaseWeight = "500 lb"
	entity.Recalculate()
	panel.Sync()
	last := len(encumbrance.Levels) - 1
	c.Equal(last, panel.current)
	c.True(panel.overloaded)
	levels, names = spoken()
	c.Equal("Current Encumbrance Level 4, carrying more than the maximum load", levels[last])
	c.Equal("Current Encumbrance Level "+encumbrance.Levels[last].String()+", carrying more than the maximum load",
		names[last])
	c.Equal("", levels[1])
	c.Equal("", names[1])
	eq.BaseWeight = "30 lb"
	entity.Recalculate()
	panel.Sync()
	levels, names = spoken()
	c.Equal("Current Encumbrance Level 1", levels[1], "back within the levels, nothing more is said")
	c.Equal("Current Encumbrance Level Light", names[1])
	c.Equal("", levels[last])
	c.Equal("", names[last])

	node := &accessibility.Node{Text: &accessibility.TextInfo{Text: "1"}}
	panel.levels[1].Accessibility.Callback(node)
	c.Nil(node.Text, "the current row's drawn text is withheld in favor of its name")
	node = &accessibility.Node{Text: &accessibility.TextInfo{Text: "0"}}
	panel.levels[0].Accessibility.Callback(node)
	c.NotNil(node.Text, "a row that is not current keeps its drawn text")
}
