// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"bytes"
	"context"
	"io/fs"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/gcs/v5/model/message"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/tid"
	"github.com/richardwilkes/toolbox/xmath/crc"
)

var _ ListProvider = &Template{}

// Template holds the GURPS Template data that is written to disk.
type Template struct {
	Version   int          `json:"version"`
	ID        tid.TID      `json:"id"`
	Traits    []*Trait     `json:"traits,alt=advantages,omitempty"`
	Skills    []*Skill     `json:"skills,omitempty"`
	Spells    []*Spell     `json:"spells,omitempty"`
	Equipment []*Equipment `json:"equipment,omitempty"`
	Notes     []*Note      `json:"notes,omitempty"`
}

// NewTemplateFromFile loads a Template from a file.
func NewTemplateFromFile(fileSystem fs.FS, filePath string) (*Template, error) {
	var t Template
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &t); err != nil {
		return nil, errs.NewWithCause(message.InvalidFileData(), err)
	}
	if !tid.IsKindAndValid(t.ID, kinds.Template) {
		t.ID = tid.MustNewTID(kinds.Template)
	}
	if err := jio.CheckVersion(t.Version); err != nil {
		return nil, err
	}
	return &t, nil
}

// NewTemplate creates a new Template.
func NewTemplate() *Template {
	template := &Template{
		ID: tid.MustNewTID(kinds.Template),
	}
	return template
}

// Entity implements EntityProvider.
func (t *Template) Entity() *Entity {
	return nil
}

// Save the Template to a file as JSON.
func (t *Template) Save(filePath string) error {
	t.Version = jio.CurrentDataVersion
	return jio.SaveToFile(context.Background(), filePath, t)
}

// TraitList implements ListProvider
func (t *Template) TraitList() []*Trait {
	return t.Traits
}

// SetTraitList implements ListProvider
func (t *Template) SetTraitList(list []*Trait) {
	t.Traits = list
}

// CarriedEquipmentList implements ListProvider
func (t *Template) CarriedEquipmentList() []*Equipment {
	return t.Equipment
}

// SetCarriedEquipmentList implements ListProvider
func (t *Template) SetCarriedEquipmentList(list []*Equipment) {
	t.Equipment = list
}

// OtherEquipmentList implements ListProvider
func (t *Template) OtherEquipmentList() []*Equipment {
	return nil
}

// SetOtherEquipmentList implements ListProvider
func (t *Template) SetOtherEquipmentList(_ []*Equipment) {
}

// SkillList implements ListProvider
func (t *Template) SkillList() []*Skill {
	return t.Skills
}

// SetSkillList implements ListProvider
func (t *Template) SetSkillList(list []*Skill) {
	t.Skills = list
}

// SpellList implements ListProvider
func (t *Template) SpellList() []*Spell {
	return t.Spells
}

// SetSpellList implements ListProvider
func (t *Template) SetSpellList(list []*Spell) {
	t.Spells = list
}

// NoteList implements ListProvider
func (t *Template) NoteList() []*Note {
	return t.Notes
}

// SetNoteList implements ListProvider
func (t *Template) SetNoteList(list []*Note) {
	t.Notes = list
}

// CRC64 computes a CRC-64 value for the canonical disk format of the data.
func (t *Template) CRC64() uint64 {
	var buffer bytes.Buffer
	if err := jio.Save(context.Background(), &buffer, t); err != nil {
		return 0
	}
	return crc.Bytes(0, buffer.Bytes())
}
