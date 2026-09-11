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
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/i18n"
)

// FilterFieldKind is the kind of value a filter field holds, which decides which criteria a condition on it uses.
// It never reaches disk; the field's key does.
type FilterFieldKind byte

// Possible FilterFieldKind values.
const (
	// FilterFieldText is a single text value, compared with a criteria.Text.
	FilterFieldText FilterFieldKind = iota
	// FilterFieldList is a list of text values, such as tags, compared with a criteria.Text using its list semantics.
	FilterFieldList
	// FilterFieldNumber is a number, compared with a criteria.Number.
	FilterFieldNumber
	// FilterFieldWeight is a weight, compared with a criteria.Weight.
	FilterFieldWeight
	// FilterFieldBool is a yes/no value that needs no criteria; a condition on it is satisfied when the value is true.
	FilterFieldBool
)

// The keys of the fields that several list types share. A key identifies a field within a list type only, so the
// same key may name fields of different kinds on different types.
const (
	filterFieldKeyName      = "name"
	filterFieldKeyNotes     = "notes"
	filterFieldKeyTags      = "tags"
	filterFieldKeyReference = "reference"
	filterFieldKeyPoints    = "points"
	filterFieldKeyTechLevel = "tech_level"
	filterFieldKeyContainer = "container"
	filterFieldKeyCost      = "cost"
	filterFieldKeyWeight    = "weight"
)

// FilterField describes one field of a list type that a saved filter may test: the key it is stored under, the title
// the filter editor shows for it (phrased to follow "must" or "must not"), its kind, and the accessor for its kind,
// which is the only one of the five that is set.
type FilterField[T Node[T]] struct {
	Key    string
	Title  string
	Kind   FilterFieldKind
	Text   func(T) string
	List   func(T) []string
	Number func(T) fxp.Int
	Weight func(T) fxp.Weight
	Bool   func(T) bool
}

// NewTextFilterField creates a field holding a single text value.
func NewTextFilterField[T Node[T]](key, title string, f func(T) string) *FilterField[T] {
	return &FilterField[T]{Key: key, Title: title, Kind: FilterFieldText, Text: f}
}

// NewListFilterField creates a field holding a list of text values.
func NewListFilterField[T Node[T]](key, title string, f func(T) []string) *FilterField[T] {
	return &FilterField[T]{Key: key, Title: title, Kind: FilterFieldList, List: f}
}

// NewNumberFilterField creates a field holding a number.
func NewNumberFilterField[T Node[T]](key, title string, f func(T) fxp.Int) *FilterField[T] {
	return &FilterField[T]{Key: key, Title: title, Kind: FilterFieldNumber, Number: f}
}

// NewWeightFilterField creates a field holding a weight.
func NewWeightFilterField[T Node[T]](key, title string, f func(T) fxp.Weight) *FilterField[T] {
	return &FilterField[T]{Key: key, Title: title, Kind: FilterFieldWeight, Weight: f}
}

// NewBoolFilterField creates a field holding a yes/no value.
func NewBoolFilterField[T Node[T]](key, title string, f func(T) bool) *FilterField[T] {
	return &FilterField[T]{Key: key, Title: title, Kind: FilterFieldBool, Bool: f}
}

// The accessors below use the *WithReplacements forms rather than Notes() and its kin: filtering runs them once per
// row on every change, and the latter resolve embedded scripts, which is far too costly for that.

// TraitFilterFields returns the fields a saved filter for a trait list may test.
func TraitFilterFields() []*FilterField[*Trait] {
	return []*FilterField[*Trait]{
		NewTextFilterField(filterFieldKeyName, i18n.Text("have a name"), (*Trait).NameWithReplacements),
		NewTextFilterField(filterFieldKeyNotes, i18n.Text("have notes"), (*Trait).LocalNotesWithReplacements),
		NewTextFilterField("user_desc", i18n.Text("have a user description"), (*Trait).UserDescWithReplacements),
		NewListFilterField(filterFieldKeyTags, i18n.Text("have tags"), (*Trait).TagList),
		NewTextFilterField(filterFieldKeyReference, i18n.Text("have a page reference"),
			func(t *Trait) string { return t.PageRef }),
		NewNumberFilterField(filterFieldKeyPoints, i18n.Text("have points"), (*Trait).AdjustedPoints),
		NewNumberFilterField("levels", i18n.Text("have levels"), (*Trait).CurrentLevel),
		NewTextFilterField("cr", i18n.Text("have a self-control roll"),
			func(t *Trait) string { return t.SelfControl.ShortString() }),
		NewTextFilterField("container_type", i18n.Text("have a container type"), func(t *Trait) string {
			if !t.Container() {
				return ""
			}
			return t.ContainerType.String()
		}),
		NewBoolFilterField(filterFieldKeyContainer, i18n.Text("be a container"), (*Trait).Container),
	}
}

// TraitModifierFilterFields returns the fields a saved filter for a trait modifier list may test.
func TraitModifierFilterFields() []*FilterField[*TraitModifier] {
	return []*FilterField[*TraitModifier]{
		NewTextFilterField(filterFieldKeyName, i18n.Text("have a name"), (*TraitModifier).NameWithReplacements),
		NewTextFilterField(filterFieldKeyNotes, i18n.Text("have notes"), (*TraitModifier).LocalNotesWithReplacements),
		NewListFilterField(filterFieldKeyTags, i18n.Text("have tags"), (*TraitModifier).TagList),
		NewTextFilterField(filterFieldKeyReference, i18n.Text("have a page reference"),
			func(t *TraitModifier) string { return t.PageRef }),
		NewTextFilterField(filterFieldKeyCost, i18n.Text("have a cost"), func(t *TraitModifier) string {
			if t.Container() {
				return ""
			}
			return t.CostModifierType().Format(t.CostModifier().Simplify())
		}),
		NewTextFilterField("affects", i18n.Text("affect the cost"), func(t *TraitModifier) string {
			if t.Container() {
				return ""
			}
			return t.Affects.String()
		}),
		NewBoolFilterField(filterFieldKeyContainer, i18n.Text("be a container"), (*TraitModifier).Container),
	}
}

// SkillFilterFields returns the fields a saved filter for a skill list may test.
func SkillFilterFields() []*FilterField[*Skill] {
	return []*FilterField[*Skill]{
		NewTextFilterField(filterFieldKeyName, i18n.Text("have a name"), (*Skill).NameWithReplacements),
		NewTextFilterField("specialization", i18n.Text("have a specialization"),
			(*Skill).SpecializationWithReplacements),
		NewTextFilterField(filterFieldKeyNotes, i18n.Text("have notes"), (*Skill).LocalNotesWithReplacements),
		NewListFilterField(filterFieldKeyTags, i18n.Text("have tags"), (*Skill).TagList),
		NewTextFilterField(filterFieldKeyReference, i18n.Text("have a page reference"),
			func(s *Skill) string { return s.PageRef }),
		NewTextFilterField("difficulty", i18n.Text("have a difficulty"), func(s *Skill) string {
			if s.Container() {
				return ""
			}
			return s.Difficulty.Description(EntityFromNode(s))
		}),
		NewTextFilterField(filterFieldKeyTechLevel, i18n.Text("have a tech level"), (*Skill).TL),
		NewNumberFilterField(filterFieldKeyPoints, i18n.Text("have points"), (*Skill).RawPoints),
		NewBoolFilterField("technique", i18n.Text("be a technique"), (*Skill).IsTechnique),
		NewBoolFilterField(filterFieldKeyContainer, i18n.Text("be a container"), (*Skill).Container),
	}
}

// SpellFilterFields returns the fields a saved filter for a spell list may test.
func SpellFilterFields() []*FilterField[*Spell] {
	return []*FilterField[*Spell]{
		NewTextFilterField(filterFieldKeyName, i18n.Text("have a name"), (*Spell).NameWithReplacements),
		NewTextFilterField(filterFieldKeyNotes, i18n.Text("have notes"), (*Spell).LocalNotesWithReplacements),
		NewListFilterField(filterFieldKeyTags, i18n.Text("have tags"), (*Spell).TagList),
		NewTextFilterField(filterFieldKeyReference, i18n.Text("have a page reference"),
			func(s *Spell) string { return s.PageRef }),
		NewTextFilterField("difficulty", i18n.Text("have a difficulty"), func(s *Spell) string {
			if s.Container() {
				return ""
			}
			return s.Difficulty.Description(EntityFromNode(s))
		}),
		NewListFilterField("college", i18n.Text("have colleges"), (*Spell).CollegeWithReplacements),
		NewTextFilterField("power_source", i18n.Text("have a power source"), (*Spell).PowerSourceWithReplacements),
		NewTextFilterField("class", i18n.Text("have a class"), (*Spell).ClassWithReplacements),
		NewTextFilterField("resist", i18n.Text("have a resistance"), (*Spell).ResistWithReplacements),
		NewTextFilterField("casting_cost", i18n.Text("have a casting cost"), (*Spell).CastingCostWithReplacements),
		NewTextFilterField("maintenance_cost", i18n.Text("have a maintenance cost"),
			(*Spell).MaintenanceCostWithReplacements),
		NewTextFilterField("casting_time", i18n.Text("have a casting time"), (*Spell).CastingTimeWithReplacements),
		NewTextFilterField("duration", i18n.Text("have a duration"), (*Spell).DurationWithReplacements),
		NewTextFilterField(filterFieldKeyTechLevel, i18n.Text("have a tech level"), (*Spell).TL),
		NewNumberFilterField(filterFieldKeyPoints, i18n.Text("have points"), (*Spell).RawPoints),
		NewBoolFilterField("ritual_magic", i18n.Text("be ritual magic"), (*Spell).IsRitualMagic),
		NewBoolFilterField(filterFieldKeyContainer, i18n.Text("be a container"), (*Spell).Container),
	}
}

// EquipmentFilterFields returns the fields a saved filter for an equipment list may test.
func EquipmentFilterFields() []*FilterField[*Equipment] {
	return []*FilterField[*Equipment]{
		NewTextFilterField(filterFieldKeyName, i18n.Text("have a name"), (*Equipment).NameWithReplacements),
		NewTextFilterField(filterFieldKeyNotes, i18n.Text("have notes"), (*Equipment).LocalNotesWithReplacements),
		NewListFilterField(filterFieldKeyTags, i18n.Text("have tags"), (*Equipment).TagList),
		NewTextFilterField(filterFieldKeyReference, i18n.Text("have a page reference"),
			func(e *Equipment) string { return e.PageRef }),
		NewTextFilterField(filterFieldKeyTechLevel, i18n.Text("have a tech level"),
			func(e *Equipment) string { return e.TechLevel }),
		// The raw legality class is what the LC column shows, so it is what a user will type.
		NewTextFilterField("legality_class", i18n.Text("have a legality class"),
			func(e *Equipment) string { return e.LegalityClass }),
		NewNumberFilterField(filterFieldKeyCost, i18n.Text("have a cost"), (*Equipment).AdjustedValue),
		NewWeightFilterField(filterFieldKeyWeight, i18n.Text("have a weight"), func(e *Equipment) fxp.Weight {
			return e.AdjustedWeight(false, SheetSettingsFor(EntityFromNode(e)).DefaultWeightUnits)
		}),
		NewBoolFilterField(filterFieldKeyContainer, i18n.Text("be a container"), (*Equipment).Container),
	}
}

// EquipmentModifierFilterFields returns the fields a saved filter for an equipment modifier list may test.
func EquipmentModifierFilterFields() []*FilterField[*EquipmentModifier] {
	return []*FilterField[*EquipmentModifier]{
		NewTextFilterField(filterFieldKeyName, i18n.Text("have a name"), (*EquipmentModifier).NameWithReplacements),
		NewTextFilterField(filterFieldKeyNotes, i18n.Text("have notes"),
			(*EquipmentModifier).LocalNotesWithReplacements),
		NewListFilterField(filterFieldKeyTags, i18n.Text("have tags"), (*EquipmentModifier).TagList),
		NewTextFilterField(filterFieldKeyReference, i18n.Text("have a page reference"),
			func(e *EquipmentModifier) string { return e.PageRef }),
		NewTextFilterField(filterFieldKeyTechLevel, i18n.Text("have a tech level"),
			func(e *EquipmentModifier) string { return e.TechLevel }),
		NewTextFilterField(filterFieldKeyCost, i18n.Text("have a cost"), (*EquipmentModifier).CostDescription),
		NewTextFilterField(filterFieldKeyWeight, i18n.Text("have a weight"), (*EquipmentModifier).WeightDescription),
		NewBoolFilterField(filterFieldKeyContainer, i18n.Text("be a container"), (*EquipmentModifier).Container),
	}
}

// NoteFilterFields returns the fields a saved filter for a note list may test.
func NoteFilterFields() []*FilterField[*Note] {
	return []*FilterField[*Note]{
		NewTextFilterField("text", i18n.Text("have text"), (*Note).TextWithReplacements),
		NewListFilterField(filterFieldKeyTags, i18n.Text("have tags"), (*Note).TagList),
		NewTextFilterField(filterFieldKeyReference, i18n.Text("have a page reference"),
			func(n *Note) string { return n.PageRef }),
		NewBoolFilterField(filterFieldKeyContainer, i18n.Text("be a container"), (*Note).Container),
	}
}
