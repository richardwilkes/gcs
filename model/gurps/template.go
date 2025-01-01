// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"hash"
	"io/fs"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/tid"
)

var (
	_ ListProvider = &Template{}
	_ DataOwner    = &Template{}
	_ Hashable     = &Template{}
)

// Template holds the GURPS Template data that is written to disk.
type Template struct {
	TemplateData
	srcMatcher *SrcMatcher
}

// TemplateData holds the GURPS Template data that is written to disk.
type TemplateData struct {
	Version   int          `json:"version"`
	ID        tid.TID      `json:"id"`
	Traits    []*Trait     `json:"traits,alt=advantages,omitempty"`
	Skills    []*Skill     `json:"skills,omitempty"`
	Spells    []*Spell     `json:"spells,omitempty"`
	Equipment []*Equipment `json:"equipment,omitempty"`
	Notes     []*Note      `json:"notes,omitempty"`
	BodyType  *Body        `json:"body_type,omitempty"`
}

// NewTemplateFromFile loads a Template from a file.
func NewTemplateFromFile(fileSystem fs.FS, filePath string) (*Template, error) {
	var t Template
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &t); err != nil {
		return nil, errs.NewWithCause(InvalidFileData(), err)
	}
	if err := jio.CheckVersion(t.Version); err != nil {
		return nil, err
	}
	return &t, nil
}

// NewTemplate creates a new Template.
func NewTemplate() *Template {
	var t Template
	t.ID = tid.MustNewTID(kinds.Template)
	return &t
}

// MarshalJSON implements json.Marshaler.
func (t *Template) MarshalJSON() ([]byte, error) {
	t.EnsureAttachments()
	t.Version = jio.CurrentDataVersion
	return json.Marshal(&t.TemplateData)
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Template) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &t.TemplateData); err != nil {
		return err
	}
	if !tid.IsKindAndValid(t.ID, kinds.Template) {
		t.ID = tid.MustNewTID(kinds.Template)
	}
	t.EnsureAttachments()
	return nil
}

// DataOwner returns the data owner.
func (t *Template) DataOwner() DataOwner {
	return t
}

// WeightUnit returns the weight unit to use for display.
func (t *Template) WeightUnit() fxp.WeightUnit {
	return GlobalSettings().SheetSettings().DefaultWeightUnits
}

// OwningEntity returns nil.
func (t *Template) OwningEntity() *Entity {
	return nil
}

// SourceMatcher returns the SourceMatcher.
func (t *Template) SourceMatcher() *SrcMatcher {
	if t.srcMatcher == nil {
		t.srcMatcher = &SrcMatcher{}
	}
	return t.srcMatcher
}

// Save the Template to a file as JSON.
func (t *Template) Save(filePath string) error {
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

// Hash writes this object's contents into the hasher.
func (t *Template) Hash(h hash.Hash) {
	var buffer bytes.Buffer
	if err := jio.Save(context.Background(), &buffer, t); err != nil {
		errs.Log(err)
		return
	}
	_, _ = h.Write(buffer.Bytes())
}

// EnsureAttachments ensures that all attachments have their data owner set to the Template.
func (t *Template) EnsureAttachments() {
	for _, one := range t.Traits {
		one.SetDataOwner(t)
	}
	for _, one := range t.Skills {
		one.SetDataOwner(t)
	}
	for _, one := range t.Spells {
		one.SetDataOwner(t)
	}
	for _, one := range t.Equipment {
		one.SetDataOwner(t)
	}
	for _, one := range t.Notes {
		one.SetDataOwner(t)
	}
}

// SyncWithLibrarySources syncs the template with the library sources.
func (t *Template) SyncWithLibrarySources() {
	Traverse(func(trait *Trait) bool {
		trait.SyncWithSource()
		Traverse(func(traitModifier *TraitModifier) bool {
			traitModifier.SyncWithSource()
			return false
		}, false, false, trait.Modifiers...)
		return false
	}, false, false, t.Traits...)
	Traverse(func(skill *Skill) bool {
		skill.SyncWithSource()
		return false
	}, false, false, t.Skills...)
	Traverse(func(spell *Spell) bool {
		spell.SyncWithSource()
		return false
	}, false, false, t.Spells...)
	Traverse(func(equipment *Equipment) bool {
		equipment.SyncWithSource()
		Traverse(func(equipmentModifier *EquipmentModifier) bool {
			equipmentModifier.SyncWithSource()
			return false
		}, false, false, equipment.Modifiers...)
		return false
	}, false, false, t.Equipment...)
	Traverse(func(note *Note) bool {
		note.SyncWithSource()
		return false
	}, false, false, t.Notes...)
}
