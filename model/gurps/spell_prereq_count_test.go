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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestCountPrereqsForSpellWithNoPrereqs verifies that a spell with no "prereqs" key -- which leaves Prereq as a nil
// *PrereqList and therefore arrives as a typed-nil Prereq interface -- counts as zero rather than panicking.
func TestCountPrereqsForSpellWithNoPrereqs(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	light := addTestSpell(e, "Light", fxp.One)
	c.Nil(light.Prereq, "a freshly created spell must have no prerequisite list")
	count := -1
	c.NotPanics(func() { count = CountPrereqsForSpell(light, e.Spells) },
		"a spell without prerequisites must not panic")
	c.Equal(0, count)
}

// TestCountPrereqsForSpellTraversesIntoSpellWithNoPrereqs verifies that the recursive traversal into a required spell
// tolerates that spell having no prerequisites of its own.
func TestCountPrereqsForSpellTraversesIntoSpellWithNoPrereqs(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addTestSpell(e, "Light", fxp.One)
	continualLight := addTestSpell(e, "Continual Light", fxp.One)
	addSpellNamePrereq(continualLight, "Light")
	count := -1
	c.NotPanics(func() { count = CountPrereqsForSpell(continualLight, e.Spells) },
		"traversal into a spell without prerequisites must not panic")
	c.Equal(1, count, "only the required spell itself should be counted")
}

// TestCountPrereqsForSpellChainedPrereqs verifies that the nil guard doesn't short-circuit the normal recursive count.
func TestCountPrereqsForSpellChainedPrereqs(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addTestSpell(e, "Ignite Fire", fxp.One)
	light := addTestSpell(e, "Light", fxp.One)
	addSpellNamePrereq(light, "Ignite Fire")
	continualLight := addTestSpell(e, "Continual Light", fxp.One)
	addSpellNamePrereq(continualLight, "Light")
	c.Equal(2, CountPrereqsForSpell(continualLight, e.Spells),
		"both Light and its own Ignite Fire prerequisite should be counted")
}

// TestCountPrereqsForSpellWithNilEntryInList verifies that a nil entry inside a prerequisite list is skipped rather
// than panicking.
func TestCountPrereqsForSpellWithNilEntryInList(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addTestSpell(e, "Light", fxp.One)
	continualLight := addTestSpell(e, "Continual Light", fxp.One)
	addSpellNamePrereq(continualLight, "Light")
	continualLight.Prereq.Prereqs = append(continualLight.Prereq.Prereqs, nil)
	count := -1
	c.NotPanics(func() { count = CountPrereqsForSpell(continualLight, e.Spells) },
		"a nil entry within a prerequisite list must not panic")
	c.Equal(1, count)
}
