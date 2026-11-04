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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestSkillPrereqCriteria checks the points and tags of a skill prerequisite.
func TestSkillPrereqCriteria(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	broadsword := gurps.NewSkill(e, nil, false)
	broadsword.Name = "Broadsword"
	broadsword.Points = fxp.Four
	broadsword.Tags = []string{"Combat"}
	e.Skills = []*gurps.Skill{broadsword}
	c.Equal(fxp.Four, broadsword.AdjustedPoints(nil), "precondition: Broadsword has 4 points")

	for _, one := range []struct {
		name  string
		setup func(p *gurps.SkillPrereq)
		want  bool
	}{
		{"points met", func(p *gurps.SkillPrereq) { p.PointsCriteria = atLeast(4) }, true},
		{"points unmet", func(p *gurps.SkillPrereq) { p.PointsCriteria = atLeast(8) }, false},
		{"tag met", func(p *gurps.SkillPrereq) { p.TagsCriteria = isText("Combat") }, true},
		{"tag unmet", func(p *gurps.SkillPrereq) { p.TagsCriteria = isText("Social") }, false},
	} {
		p := gurps.NewSkillPrereq()
		p.NameCriteria = isText("Broadsword")
		one.setup(p)
		c.Equal(one.want, p.Satisfied(e, nil, nil, "", nil), one.name)
	}
}
