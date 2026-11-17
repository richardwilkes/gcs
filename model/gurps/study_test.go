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
	"testing"
	"testing/fstest"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/study"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestCloneStudyList verifies the shared study deep-copy: an empty list clones to nil and each session is copied so the
// clone can be edited without touching the source.
func TestCloneStudyList(t *testing.T) {
	c := check.New(t)
	c.Nil(cloneStudyList(nil), "an absent list clones to nil")
	c.Nil(cloneStudyList([]*Study{}), "an empty list clones to nil")

	list := []*Study{{Type: study.Self, Hours: fxp.Four, Note: "Reading"}, {Type: study.Job, Hours: fxp.Two}}
	clone := cloneStudyList(list)
	c.Equal(len(list), len(clone), "every session is carried over")
	for i := range list {
		c.True(list[i] != clone[i], "session %d is a distinct object", i)
		c.Equal(*list[i], *clone[i], "session %d holds the same data", i)
	}
	clone[0].Hours = fxp.One
	c.Equal(fxp.Four, list[0].Hours, "editing the clone leaves the source alone")
}

// TestResolveStudyHours verifies that each session's hours count toward the total by what its type of study is worth:
// half for self-study, a quarter on the job, in full with a teacher and double for intensive training.
func TestResolveStudyHours(t *testing.T) {
	c := check.New(t)
	// No other assignment of the four multipliers to these hours gives the same total.
	c.Equal(fxp.Eight, ResolveStudyHours([]*Study{
		{Type: study.Self, Hours: fxp.Four},
		{Type: study.Job, Hours: fxp.Eight},
		{Type: study.Teacher, Hours: fxp.Two},
		{Type: study.Intensive, Hours: fxp.One},
	}))
}

// TestStudyHoursProgressText verifies the study progress text: nothing without hours unless forced, when negative hours
// show as 0, the hours the study level needs, and emphasis unless forced.
func TestStudyHoursProgressText(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		hours  fxp.Int
		needed study.Level
		force  bool
		want   string
	}{
		{0, study.Standard, false, ""},
		{-fxp.One, study.Standard, false, ""},
		{-fxp.One, study.Standard, true, "Studied 0 of 200 hours"},
		{fxp.Fifty, study.Standard, false, "_Studied 50 of 200 hours._"},
		{fxp.Fifty, study.Level2, false, "_Studied 50 of 160 hours._"},
		{fxp.Fifty, study.Level2, true, "Studied 50 of 160 hours"},
	} {
		c.Equal(tc.want, StudyHoursProgressText(tc.hours, tc.needed, tc.force), "%v hours, %v, forced %v", tc.hours,
			tc.needed, tc.force)
	}
}

// TestLoadingOffASheetClearsStudy verifies that a template and a trait, skill or spell list have any study and study
// hours needed removed when loaded, while a character sheet keeps both.
func TestLoadingOffASheetClearsStudy(t *testing.T) {
	c := check.New(t)
	studied := func() ([]*Study, study.Level) {
		return []*Study{{Type: study.Teacher, Hours: fxp.Ten}}, study.Level2
	}
	trait := NewTrait(nil, nil, false)
	trait.Study, trait.StudyHoursNeeded = studied()
	skill := NewSkill(nil, nil, false)
	skill.Study, skill.StudyHoursNeeded = studied()
	spell := NewSpell(nil, nil, false)
	spell.Study, spell.StudyHoursNeeded = studied()

	entity := NewEntity()
	entity.Traits = []*Trait{trait}
	entity.Skills = []*Skill{skill}
	entity.Spells = []*Spell{spell}
	data, err := json.Marshal(entity)
	c.NoError(err)
	var loadedEntity Entity
	c.NoError(json.Unmarshal(data, &loadedEntity))
	c.Equal(1, len(loadedEntity.Traits[0].Study), "a character sheet must keep a trait's study")
	c.Equal(study.Level2, loadedEntity.Skills[0].StudyHoursNeeded, "a character sheet must keep a skill's hours needed")
	c.Equal(1, len(loadedEntity.Spells[0].Study), "a character sheet must keep a spell's study")

	template := NewTemplate()
	template.Traits = []*Trait{trait}
	template.Skills = []*Skill{skill}
	template.Spells = []*Spell{spell}
	data, err = json.Marshal(template)
	c.NoError(err)
	var loadedTemplate Template
	c.NoError(json.Unmarshal(data, &loadedTemplate))
	c.Equal(0, len(loadedTemplate.Traits[0].Study), "a template must not keep a trait's study")
	c.Equal(study.Standard, loadedTemplate.Traits[0].StudyHoursNeeded, "a template must not keep a trait's hours needed")
	c.Equal(0, len(loadedTemplate.Skills[0].Study), "a template must not keep a skill's study")
	c.Equal(study.Standard, loadedTemplate.Skills[0].StudyHoursNeeded, "a template must not keep a skill's hours needed")
	c.Equal(0, len(loadedTemplate.Spells[0].Study), "a template must not keep a spell's study")
	c.Equal(study.Standard, loadedTemplate.Spells[0].StudyHoursNeeded, "a template must not keep a spell's hours needed")

	list, err := json.Marshal(&listData[*Skill]{Header: jio.NewHeader(SkillListDataKind), Rows: []*Skill{skill}})
	c.NoError(err)
	rows, err := NewSkillsFromFile(fstest.MapFS{"list.skl": &fstest.MapFile{Data: list}}, "list.skl")
	c.NoError(err)
	c.Equal(0, len(rows[0].Study), "a skill list must not keep study")
	c.Equal(study.Standard, rows[0].StudyHoursNeeded, "a skill list must not keep study hours needed")
}
