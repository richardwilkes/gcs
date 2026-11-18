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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/study"
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

// TestStudyPanelEditsTheStudyList verifies that the study panel's add button puts an entry of the type last chosen at
// the top of the list, that the hours typed into an entry count toward the total shown, and that an entry's delete
// button takes that entry out of the list and the total.
func TestStudyPanelEditsTheStudyList(t *testing.T) {
	c := check.New(t)
	previous := lastStudyTypeUsed
	t.Cleanup(func() { lastStudyTypeUsed = previous })
	lastStudyTypeUsed = study.Teacher
	needed := study.Standard
	first := &gurps.Study{Type: study.Teacher, Hours: fxp.Ten}
	studies := []*gurps.Study{first}
	p := newStudyPanel(&needed, &studies)
	c.Equal("Studied 10 of 200 hours (", p.total.Text.String())
	buttonOf := func(panel *unison.Panel, icon *unison.SVG) *unison.Button {
		for _, b := range uxtest.PanelsOfType[*unison.Button](panel) {
			if drawable, ok := b.Drawable.(*unison.DrawableSVG); ok && drawable.SVG == icon {
				return b
			}
		}
		t.Fatal("no button shows the icon")
		return nil
	}
	buttonOf(p.Children()[0], unison.CircledAddSVG).ClickCallback()
	rows := p.Children()[1:]
	c.Equal(2, len(studies), "the add button adds an entry")
	c.Equal(2, len(rows), "with a row of its own")
	if len(studies) != 2 || len(rows) != 2 {
		return
	}
	c.True(studies[1] == first, "the new entry goes first")
	added := studies[0]
	c.Equal(study.Teacher, added.Type, "of the type last chosen")
	uxtest.PanelsOfType[*DecimalField](rows[0])[0].SetText("4")
	c.Equal(fxp.Four, added.Hours)
	c.Equal("Studied 14 of 200 hours (", p.total.Text.String(), "the hours count toward the total")
	buttonOf(rows[1], unison.TrashSVG).ClickCallback()
	c.Equal([]*gurps.Study{added}, studies, "deleting an entry takes it out of the list")
	c.Equal(2, len(p.Children()), "along with its row")
	c.Equal("Studied 4 of 200 hours (", p.total.Text.String(), "and the total")
}
