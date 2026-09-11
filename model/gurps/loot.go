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
	"encoding/json/jsontext"
	"encoding/json/v2"
	"hash"
	"io/fs"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox/v2/tid"
)

var (
	_ ListProvider     = &Loot{}
	_ DataOwner        = &Loot{}
	_ Hashable         = &Loot{}
	_ PageInfoProvider = &Loot{}
)

// Loot holds the base information for loot.
type Loot struct {
	LootData
	srcMatcher *SrcMatcher
}

// LootData holds the Loot data that is written to disk.
type LootData struct {
	Version    int          `json:"version"`
	ID         tid.TID      `json:"id"`
	Name       string       `json:"name,omitzero"`
	Location   string       `json:"location,omitzero"`
	Session    string       `json:"session,omitzero"`
	ModifiedOn jio.Time     `json:"modified_date"`
	Equipment  []*Equipment `json:"equipment,omitempty"`
	Notes      []*Note      `json:"notes,omitempty"`
}

// NewLootFromFile loads Loot from a file.
func NewLootFromFile(fileSystem fs.FS, filePath string) (*Loot, error) {
	return newLootFromFile(fileSystem, filePath)
}

// NewLootFromFileWithSavedCalc loads a Loot from a file, keeping the resolved text that the file's "calc" objects
// record for its notes (see Note.StringWithSavedCalc), so that a reader which only needs the text of the loot sheet
// need not run the scripts embedded in its notes. It is the counterpart of NewEntityFromFileWithSavedCalc for loot
// sheets.
func NewLootFromFileWithSavedCalc(fileSystem fs.FS, filePath string) (*Loot, error) {
	return newLootFromFile(fileSystem, filePath, json.WithUnmarshalers(savedCalcMarker))
}

func newLootFromFile(fileSystem fs.FS, filePath string, opts ...json.Options) (*Loot, error) {
	var l Loot
	if err := jio.LoadVersionedFile(fileSystem, filePath, &l, &l.Version, opts...); err != nil {
		return nil, err
	}
	return &l, nil
}

// NewLoot creates a new Loot.
func NewLoot() *Loot {
	var l Loot
	l.ID = tid.MustNewTID(kinds.Loot)
	l.ModifiedOn = jio.Now()
	return &l
}

// Save the Loot to a file as JSON.
func (l *Loot) Save(filePath string) error {
	AdjustEquipmentUsesForSave(l.Equipment)
	return jio.SaveToFile(filePath, l)
}

// MarshalJSONTo implements json.MarshalerTo.
func (l *Loot) MarshalJSONTo(enc *jsontext.Encoder) error {
	l.EnsureAttachments()
	l.Version = jio.CurrentDataVersion
	return json.MarshalEncode(enc, &l.LootData)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (l *Loot) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	l.LootData = LootData{}
	if err := json.UnmarshalDecode(dec, &l.LootData); err != nil {
		return err
	}
	if !tid.IsKindAndValid(l.ID, kinds.Loot) {
		l.ID = tid.MustNewTID(kinds.Loot)
	}
	l.EnsureAttachments()
	return nil
}

// OwningEntity implements DataOwner.
func (l *Loot) OwningEntity() *Entity {
	return nil
}

// SourceMatcher implements DataOwner.
func (l *Loot) SourceMatcher() *SrcMatcher {
	if l.srcMatcher == nil {
		l.srcMatcher = &SrcMatcher{}
	}
	return l.srcMatcher
}

// Hash implements Hashable.
func (l *Loot) Hash(h hash.Hash) {
	saved := l.ModifiedOn
	l.ModifiedOn = jio.Time{}
	defer func() { l.ModifiedOn = saved }()
	HashJSON(h, l)
}

// EnsureAttachments ensures that all attachments have their data owner set properly.
func (l *Loot) EnsureAttachments() {
	SetDataOwnerAll(l, l.Equipment)
	SetDataOwnerAll(l, l.Notes)
}

// SyncWithLibrarySources syncs the loot sheet with the library sources.
func (l *Loot) SyncWithLibrarySources() {
	syncWithLibrarySources(l)
}

// DataOwner implements ListProvider.
func (l *Loot) DataOwner() DataOwner {
	return l
}

// WeightUnit returns the weight unit to use for display.
func (l *Loot) WeightUnit() fxp.WeightUnit {
	return GlobalSettings().SheetSettings().DefaultWeightUnits
}

// OtherEquipmentList implements ListProvider.
func (l *Loot) OtherEquipmentList() []*Equipment {
	return l.Equipment
}

// SetOtherEquipmentList implements ListProvider.
func (l *Loot) SetOtherEquipmentList(list []*Equipment) {
	l.Equipment = list
}

// NoteList implements ListProvider.
func (l *Loot) NoteList() []*Note {
	return l.Notes
}

// SetNoteList implements ListProvider.
func (l *Loot) SetNoteList(list []*Note) {
	l.Notes = list
}

// CarriedEquipmentList implements ListProvider.
func (l *Loot) CarriedEquipmentList() []*Equipment {
	return nil
}

// SetCarriedEquipmentList implements ListProvider.
func (l *Loot) SetCarriedEquipmentList(_ []*Equipment) {
}

// SkillList implements ListProvider.
func (l *Loot) SkillList() []*Skill {
	return nil
}

// SetSkillList implements ListProvider.
func (l *Loot) SetSkillList(_ []*Skill) {
}

// SpellList implements ListProvider.
func (l *Loot) SpellList() []*Spell {
	return nil
}

// SetSpellList implements ListProvider.
func (l *Loot) SetSpellList(_ []*Spell) {
}

// TraitList implements ListProvider.
func (l *Loot) TraitList() []*Trait {
	return nil
}

// SetTraitList implements ListProvider.
func (l *Loot) SetTraitList(_ []*Trait) {
}

// PageTitle implements PageInfoProvider.
func (l *Loot) PageTitle() string {
	return l.Name
}

// PageKeywords implements PageInfoProvider.
func (l *Loot) PageKeywords() string {
	return "GCS Loot Sheet"
}

// ModifiedOnString implements PageInfoProvider.
func (l *Loot) ModifiedOnString() string {
	return l.ModifiedOn.String()
}

// PageSettings implements PageInfoProvider.
func (l *Loot) PageSettings() *PageSettings {
	return GlobalSettings().SheetSettings().Page
}
