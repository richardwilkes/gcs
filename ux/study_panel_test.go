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
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestStudyPanelOnlyOnSheets verifies that the trait, skill and spell editors show the study panel for a row on a
// character sheet and leave it out for a row in a template or a library.
func TestStudyPanelOnlyOnSheets(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	template := newTestTemplateDockable("Study", gurps.NewTemplate())
	owners := []struct {
		name      string
		owner     gurps.DataOwner
		rebuilder Rebuildable
		want      int
	}{
		{"sheet", sheet.Entity(), sheet, 1},
		{"template", template.template, template, 0},
		{"library", nil, template, 0},
	}
	for _, one := range owners {
		contents := make(map[string]*unison.Panel)
		_, contents["trait"] = buildEditorContent(one.rebuilder, gurps.NewTrait(one.owner, nil, false), initTraitEditor)
		_, contents["skill"] = buildEditorContent(one.rebuilder, gurps.NewSkill(one.owner, nil, false), initSkillEditor)
		_, contents["spell"] = buildEditorContent(one.rebuilder, gurps.NewSpell(one.owner, nil, false), initSpellEditor)
		for kind, content := range contents {
			c.Equal(one.want, len(uxtest.PanelsOfType[*studyPanel](content)), "%s in a %s", kind, one.name)
		}
	}
}
