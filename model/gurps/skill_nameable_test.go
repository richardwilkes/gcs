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

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestSkillFillWithNameableKeys verifies that nameable substitution keys are extracted from all of a skill's text
// fields, including the optional specialization (see issue #1054).
func TestSkillFillWithNameableKeys(t *testing.T) {
	c := check.New(t)
	var s Skill
	s.Name = "Skill @who@"
	s.Specialization = "@spec@"
	s.OptionalSpecialization = "@opt@"
	s.LocalNotes = "@note@"
	m := make(map[string]string)
	s.FillWithNameableKeys(m, nil)
	c.Equal(map[string]string{
		"who":  "who",
		"spec": "spec",
		"opt":  "opt",
		"note": "note",
	}, m)
}

// TestSkillOptionalSpecializationOnlyNameableKey verifies that a substitution appearing only in the optional
// specialization is still detected (the specific regression reported in issue #1054).
func TestSkillOptionalSpecializationOnlyNameableKey(t *testing.T) {
	c := check.New(t)
	var s Skill
	s.Name = "Naturalist"
	s.Specialization = "Plants"
	s.OptionalSpecialization = "@region@"
	m := make(map[string]string)
	s.FillWithNameableKeys(m, nil)
	c.Equal(map[string]string{"region": "region"}, m)
}
