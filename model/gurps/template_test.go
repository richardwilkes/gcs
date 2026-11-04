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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/eqcontainer"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestTemplateDetailsRoundTrip verifies that the values describing the template as a whole survive a save and a load,
// that an open end of the target range stays open rather than coming back as zero, and that a template with none of
// them set is written just as it was before they existed, so that existing files don't change merely by being saved
// again.
func TestTemplateDetailsRoundTrip(t *testing.T) {
	c := check.New(t)
	template := NewTemplate()
	data, err := json.Marshal(template)
	c.NoError(err)
	text := string(data)
	for _, key := range []string{"name", "reference", "reference_highlight", "local_notes", "vtt_notes", "target_points"} {
		c.False(strings.Contains(text, `"`+key+`"`), "an unset %s must not be written", key)
	}

	template.Name = "Knight"
	template.PageRef = "DF1:5"
	template.PageRefHighlight = "Knight"
	template.LocalNotes = "Heavily armored fighter."
	template.TargetPoints = numericRangeAtLeast(fxp.FromInteger(250))

	data, err = json.Marshal(template)
	c.NoError(err)
	text = string(data)
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

// TestTemplateLoadGroupsEquipmentChoiceContainers verifies that loading a template turns an equipment choice container
// that isn't a group into one, clearing what only a physical container holds.
func TestTemplateLoadGroupsEquipmentChoiceContainers(t *testing.T) {
	c := check.New(t)
	choice := newEquipmentChoice("Sword", "Axe")
	choice.ContainerType = eqcontainer.Container
	choice.BaseValue = "50"
	choice.VTTNotes = "vtt"
	choice.Tags = []string{"Weapon"}
	choice.Source = Source{Library: "lib", Path: "choice.eqp", TID: choice.ID()}
	backpack := NewEquipment(nil, nil, true)
	backpack.BaseValue = "60"
	tmpl := NewTemplate()
	tmpl.Equipment = []*Equipment{choice, backpack}

	data, err := json.Marshal(tmpl)
	c.NoError(err)
	var loaded Template
	c.NoError(json.Unmarshal(data, &loaded))
	c.Equal(2, len(loaded.Equipment))
	c.True(loaded.Equipment[0].IsGroup(), "a choice container must become a group")
	c.Equal("", loaded.Equipment[0].BaseValue, "a choice container must not keep a value of its own")
	c.Equal("", loaded.Equipment[0].VTTNotes, "a choice container must not keep VTT notes")
	c.Equal(0, len(loaded.Equipment[0].Tags), "a choice container must not keep tags")
	c.True(loaded.Equipment[0].Source.IsZero(), "a choice container must not keep a source")
	c.True(IsTemplateChoiceContainer(loaded.Equipment[0]), "the choices themselves must survive")
	c.True(loaded.Equipment[1].IsPhysicalContainer(), "a container without choices keeps its type")
	c.Equal("60", loaded.Equipment[1].BaseValue)
}
