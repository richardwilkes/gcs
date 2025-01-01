// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"os"
	"path/filepath"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/unison"
)

type skillListProvider struct {
	skills []*gurps.Skill
}

func (p *skillListProvider) DataOwner() gurps.DataOwner {
	return nil
}

func (p *skillListProvider) SkillList() []*gurps.Skill {
	return p.skills
}

func (p *skillListProvider) SetSkillList(list []*gurps.Skill) {
	p.skills = list
}

// NewSkillTableDockableFromFile loads a list of skills from a file and creates a new unison.Dockable for them.
func NewSkillTableDockableFromFile(filePath string) (unison.Dockable, error) {
	skills, err := gurps.NewSkillsFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	d := NewSkillTableDockable(filePath, skills)
	d.needsSaveAsPrompt = false
	return d, nil
}

// NewSkillTableDockable creates a new unison.Dockable for skill list files.
func NewSkillTableDockable(filePath string, skills []*gurps.Skill) *TableDockable[*gurps.Skill] {
	provider := &skillListProvider{skills: skills}
	return NewTableDockable(filePath, gurps.SkillsExt, NewSkillsProvider(provider, false),
		func(path string) error { return gurps.SaveSkills(provider.SkillList(), path) },
		NewSkillItemID, NewSkillContainerItemID, NewTechniqueItemID)
}
