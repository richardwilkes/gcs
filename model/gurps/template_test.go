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
	"encoding/json/v2"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestTemplateDetailsRoundTrip verifies that the values describing the template as a whole survive a save and a load,
// and that an open end of the target range stays open rather than coming back as zero.
func TestTemplateDetailsRoundTrip(t *testing.T) {
	c := check.New(t)
	template := NewTemplate()
	template.Name = "Knight"
	template.PageRef = "DF1:5"
	template.PageRefHighlight = "Knight"
	template.LocalNotes = "Heavily armored fighter."
	template.TargetPoints = numericRangeAtLeast(fxp.FromInteger(250))

	data, err := json.Marshal(template)
	c.NoError(err)
	text := string(data)
	c.True(strings.Contains(text, `"target_points":{"min":250}`), "an open end must be left out, not written as zero")
	c.True(strings.Contains(text, `"local_notes":"Heavily armored fighter."`))

	var loaded Template
	c.NoError(json.Unmarshal(data, &loaded))
	c.Equal("Knight", loaded.Name)
	c.Equal("DF1:5", loaded.PageRef)
	c.Equal("Knight", loaded.PageRefHighlight)
	c.Equal("Heavily armored fighter.", loaded.LocalNotes)
	c.Equal("250+", loaded.TargetPoints.String())
	c.Nil(loaded.TargetPoints.Max, "the open end must load as open")
}

// TestTemplateWithoutDetailsWritesNone verifies that a template with none of the new values set is written just as it
// was before they existed, so that existing files don't change merely by being saved again.
func TestTemplateWithoutDetailsWritesNone(t *testing.T) {
	c := check.New(t)
	data, err := json.Marshal(NewTemplate())
	c.NoError(err)
	text := string(data)
	for _, key := range []string{"name", "reference", "reference_highlight", "local_notes", "vtt_notes", "target_points"} {
		c.False(strings.Contains(text, `"`+key+`"`), "an unset %s must not be written", key)
	}
}

// TestTemplatePointsRange verifies that what a template costs is the total of its traits, skills and spells, with a
// choice still to be made leaving that total a range and a disabled item counting for nothing.
func TestTemplatePointsRange(t *testing.T) {
	c := check.New(t)
	template := NewTemplate()
	c.Equal("0", template.PointsRange().String(), "an empty template costs nothing")

	trait := NewTrait(nil, nil, false)
	trait.BasePoints = fxp.FromInteger(10)
	choice := newTemplateChoiceTrait("Pick One", "Cheap", "Dear")
	choice.Children[0].BasePoints = fxp.FromInteger(5)
	choice.Children[1].BasePoints = fxp.FromInteger(15)
	skill := NewSkill(nil, nil, false)
	skill.Points = fxp.FromInteger(4)
	spell := NewSpell(nil, nil, false)
	spell.Points = fxp.One
	disabledSkill := NewSkill(nil, nil, false)
	disabledSkill.Points = fxp.FromInteger(8)
	disabledSkill.Disabled = true
	template.Traits = []*Trait{trait, choice}
	template.Skills = []*Skill{skill, disabledSkill}
	template.Spells = []*Spell{spell}
	template.EnsureAttachments()

	c.Equal("20~30", template.PointsRange().String())
}
