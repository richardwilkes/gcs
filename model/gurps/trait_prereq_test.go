// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

// isText returns a criterion that is met by text that is the qualifier.
func isText(qualifier string) criteria.Text {
	return criteria.Text{Compare: criteria.IsText, Qualifier: qualifier}
}

// atLeast returns a criterion that is met by a number of at least the qualifier.
func atLeast(qualifier int) criteria.Number {
	return criteria.Number{Compare: criteria.AtLeastNumber, Qualifier: fxp.FromInteger(qualifier)}
}

// TestTraitPrereqCriteria checks the points and tags of a trait prerequisite, and its modifier criteria, which a single
// enabled modifier has to meet together, containers left out.
func TestTraitPrereqCriteria(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	magery := gurps.NewTrait(e, nil, false)
	magery.Name = "Magery"
	magery.BasePoints = fxp.FromInteger(15)
	magery.Tags = []string{"Mental"}
	oneCollege := gurps.NewTraitModifier(e, nil, false)
	oneCollege.Name = "One College"
	oneCollege.LocalNotes = "Fire"
	oneCollege.Tags = []string{"Limitation"}
	water := gurps.NewTraitModifier(e, nil, false)
	water.Name = "Water Only"
	water.LocalNotes = "Water"
	solar := gurps.NewTraitModifier(e, nil, false)
	solar.Name = "Solar"
	solar.Tags = []string{"Limitation"}
	solar.SetDisabled(true)
	group := gurps.NewTraitModifier(e, nil, true)
	group.Name = "Group"
	magery.Modifiers = []*gurps.TraitModifier{oneCollege, water, solar, group}
	e.Traits = []*gurps.Trait{magery}
	c.Equal(fxp.FromInteger(15), magery.AdjustedPoints(nil), "precondition: Magery is worth 15 points")
	c.True(oneCollege.Enabled(), "precondition: One College is enabled")
	c.True(water.Enabled(), "precondition: Water Only is enabled")
	c.False(solar.Enabled(), "precondition: Solar is disabled")
	c.True(group.Container(), "precondition: Group is a container")

	for _, one := range []struct {
		name  string
		setup func(p *gurps.TraitPrereq)
		want  bool
	}{
		{"points met", func(p *gurps.TraitPrereq) { p.PointsCriteria = atLeast(15) }, true},
		{"points unmet", func(p *gurps.TraitPrereq) { p.PointsCriteria = atLeast(16) }, false},
		{"tag met", func(p *gurps.TraitPrereq) { p.TagsCriteria = isText("Mental") }, true},
		{"tag unmet", func(p *gurps.TraitPrereq) { p.TagsCriteria = isText("Physical") }, false},
		{"modifier name", func(p *gurps.TraitPrereq) { p.ModifierNameCriteria = isText("One College") }, true},
		{"disabled modifier name", func(p *gurps.TraitPrereq) { p.ModifierNameCriteria = isText("Solar") }, false},
		{"container modifier name", func(p *gurps.TraitPrereq) { p.ModifierNameCriteria = isText("Group") }, false},
		{"modifier notes", func(p *gurps.TraitPrereq) { p.ModifierNotesCriteria = isText("Fire") }, true},
		{"modifier tag", func(p *gurps.TraitPrereq) { p.ModifierTagsCriteria = isText("Limitation") }, true},
		{"one modifier meets all", func(p *gurps.TraitPrereq) {
			p.ModifierNameCriteria = isText("One College")
			p.ModifierTagsCriteria = isText("Limitation")
		}, true},
		{"no one modifier meets all", func(p *gurps.TraitPrereq) {
			p.ModifierNameCriteria = isText("One College")
			p.ModifierNotesCriteria = isText("Water")
		}, false},
	} {
		p := namedTrait("Magery")
		one.setup(p)
		c.Equal(one.want, p.Satisfied(e, nil, nil, "", nil), one.name)
	}
}

// TestTraitPrereqContainerModifiers checks that a trait's modifier criteria see the enabled modifiers of the containers
// holding it, which apply to it.
func TestTraitPrereqContainerModifiers(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	container := gurps.NewTrait(e, nil, true)
	container.Name = "Breath Weapons"
	cone := gurps.NewTraitModifier(e, nil, false)
	cone.Name = "Cone"
	jet := gurps.NewTraitModifier(e, nil, false)
	jet.Name = "Jet"
	jet.SetDisabled(true)
	container.Modifiers = []*gurps.TraitModifier{cone, jet}
	breath := gurps.NewTrait(e, container, false)
	breath.Name = "Innate Attack"
	container.Children = []*gurps.Trait{breath}
	e.Traits = []*gurps.Trait{container}
	c.True(cone.Enabled(), "precondition: Cone is enabled")
	c.False(jet.Enabled(), "precondition: Jet is disabled")
	c.Equal(0, len(breath.Modifiers), "precondition: the trait has no modifiers of its own")

	p := namedTrait("Innate Attack")
	p.ModifierNameCriteria = isText("Cone")
	c.True(p.Satisfied(e, nil, nil, "", nil), "the container's enabled modifier applies to the trait")
	p.ModifierNameCriteria = isText("Jet")
	c.False(p.Satisfied(e, nil, nil, "", nil), "its disabled one doesn't")
}
