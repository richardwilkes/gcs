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
	"slices"
	"strings"

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
// the filter editor shows for it (phrased to follow "must" or "must not"), whether that title names something plural,
// such as "have notes", its kind, and the accessor for its kind, which is the only one of the five that is set.
type FilterField[T Node[T]] struct {
	Key    string
	Title  string
	Plural bool
	Kind   FilterFieldKind
	Text   func(T) string
	List   func(T) []string
	Number func(T) fxp.Int
	Weight func(T) fxp.Weight
	Bool   func(T) bool
	// Has reports whether the node has the field at all, when it can lack one, such as the levels of a trait that
	// can't be leveled. A node without the field satisfies no criteria on it.
	Has func(T) bool
}

// WithPluralTitle marks the field's title as naming something plural, so that what follows it agrees, as in "have notes
// that contain", and returns the field.
func (f *FilterField[T]) WithPluralTitle() *FilterField[T] {
	f.Plural = true
	return f
}

// WithPresence sets what reports whether a node has the field at all, and returns the field.
func (f *FilterField[T]) WithPresence(has func(T) bool) *FilterField[T] {
	f.Has = has
	return f
}

// NewTextFilterField creates a field holding a single text value.
func NewTextFilterField[T Node[T]](key, title string, f func(T) string) *FilterField[T] {
	return &FilterField[T]{Key: key, Title: title, Kind: FilterFieldText, Text: f}
}

// NewListFilterField creates a field holding a list of text values. Its accessor leaves out the values f returns that
// are empty or only space, so they never match, and a list of nothing else holds nothing.
func NewListFilterField[T Node[T]](key, title string, f func(T) []string) *FilterField[T] {
	isBlank := func(value string) bool { return strings.TrimSpace(value) == "" }
	return &FilterField[T]{Key: key, Title: title, Kind: FilterFieldList, List: func(node T) []string {
		values := f(node)
		if !slices.ContainsFunc(values, isBlank) {
			return values
		}
		return slices.DeleteFunc(slices.Clone(values), isBlank)
	}}
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
		NewTextFilterField(filterFieldKeyNotes, i18n.Text("have notes"),
			(*Trait).LocalNotesWithReplacements).WithPluralTitle(),
		NewTextFilterField("user_desc", i18n.Text("have a user description"), (*Trait).UserDescWithReplacements),
		NewListFilterField(filterFieldKeyTags, i18n.Text("have tags"), (*Trait).TagList).WithPluralTitle(),
		NewTextFilterField(filterFieldKeyReference, i18n.Text("have a page reference"),
			func(t *Trait) string { return t.PageRef }),
		NewNumberFilterField(filterFieldKeyPoints, i18n.Text("have points"),
			func(t *Trait) fxp.Int { return t.AdjustedPoints(nil) }).WithPluralTitle(),
		NewNumberFilterField("levels", i18n.Text("have levels"), (*Trait).CurrentLevel).WithPluralTitle().
			WithPresence((*Trait).IsLeveled),
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
		NewTextFilterField(filterFieldKeyNotes, i18n.Text("have notes"),
			(*TraitModifier).LocalNotesWithReplacements).WithPluralTitle(),
		NewListFilterField(filterFieldKeyTags, i18n.Text("have tags"), (*TraitModifier).TagList).WithPluralTitle(),
		NewTextFilterField(filterFieldKeyReference, i18n.Text("have a page reference"),
			func(t *TraitModifier) string { return t.PageRef }),
		NewTextFilterField(filterFieldKeyCost, i18n.Text("have a cost"), func(t *TraitModifier) string {
			if t.Container() {
				return ""
			}
			return t.CostModifierType().Format(t.CostModifier().Simplify())
		}),
		NewTextFilterField("affects", i18n.Text("have a cost application"), func(t *TraitModifier) string {
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
		NewTextFilterField(filterFieldKeyNotes, i18n.Text("have notes"),
			(*Skill).LocalNotesWithReplacements).WithPluralTitle(),
		NewListFilterField(filterFieldKeyTags, i18n.Text("have tags"), (*Skill).TagList).WithPluralTitle(),
		NewTextFilterField(filterFieldKeyReference, i18n.Text("have a page reference"),
			func(s *Skill) string { return s.PageRef }),
		NewTextFilterField("difficulty", i18n.Text("have a difficulty"), func(s *Skill) string {
			if s.Container() {
				return ""
			}
			return s.Difficulty.Description(EntityFromNode(s))
		}),
		NewNumberFilterField(filterFieldKeyPoints, i18n.Text("have points"), (*Skill).RawPoints).WithPluralTitle(),
		NewBoolFilterField("technique", i18n.Text("be a technique"), (*Skill).IsTechnique),
		NewBoolFilterField(filterFieldKeyContainer, i18n.Text("be a container"), (*Skill).Container),
	}
}

// SpellFilterFields returns the fields a saved filter for a spell list may test.
func SpellFilterFields() []*FilterField[*Spell] {
	return []*FilterField[*Spell]{
		NewTextFilterField(filterFieldKeyName, i18n.Text("have a name"), (*Spell).NameWithReplacements),
		NewTextFilterField(filterFieldKeyNotes, i18n.Text("have notes"),
			(*Spell).LocalNotesWithReplacements).WithPluralTitle(),
		NewListFilterField(filterFieldKeyTags, i18n.Text("have tags"), (*Spell).TagList).WithPluralTitle(),
		NewTextFilterField(filterFieldKeyReference, i18n.Text("have a page reference"),
			func(s *Spell) string { return s.PageRef }),
		NewTextFilterField("difficulty", i18n.Text("have a difficulty"), func(s *Spell) string {
			if s.Container() {
				return ""
			}
			return s.Difficulty.Description(EntityFromNode(s))
		}),
		NewListFilterField("college", i18n.Text("have colleges"), (*Spell).CollegeWithReplacements).WithPluralTitle(),
		NewTextFilterField("power_source", i18n.Text("have a power source"), (*Spell).PowerSourceWithReplacements),
		NewTextFilterField("class", i18n.Text("have a class"), (*Spell).ClassWithReplacements),
		NewTextFilterField("resist", i18n.Text("have a resistance"), (*Spell).ResistWithReplacements),
		NewTextFilterField("casting_cost", i18n.Text("have a casting cost"), (*Spell).CastingCostWithReplacements),
		NewTextFilterField("maintenance_cost", i18n.Text("have a maintenance cost"),
			(*Spell).MaintenanceCostWithReplacements),
		NewTextFilterField("casting_time", i18n.Text("have a casting time"), (*Spell).CastingTimeWithReplacements),
		NewTextFilterField("duration", i18n.Text("have a duration"), (*Spell).DurationWithReplacements),
		NewNumberFilterField(filterFieldKeyPoints, i18n.Text("have points"), (*Spell).RawPoints).WithPluralTitle(),
		NewBoolFilterField("ritual_magic", i18n.Text("be ritual magic"), (*Spell).IsRitualMagic),
		NewBoolFilterField(filterFieldKeyContainer, i18n.Text("be a container"), (*Spell).Container),
	}
}

// EquipmentFilterFields returns the fields a saved filter for an equipment list may test.
func EquipmentFilterFields() []*FilterField[*Equipment] {
	return []*FilterField[*Equipment]{
		NewTextFilterField(filterFieldKeyName, i18n.Text("have a name"), (*Equipment).NameWithReplacements),
		NewTextFilterField(filterFieldKeyNotes, i18n.Text("have notes"),
			(*Equipment).LocalNotesWithReplacements).WithPluralTitle(),
		NewListFilterField(filterFieldKeyTags, i18n.Text("have tags"), (*Equipment).TagList).WithPluralTitle(),
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
		NewTextFilterField("container_type", i18n.Text("have a container type"), func(e *Equipment) string {
			if !e.Container() {
				return ""
			}
			return e.ContainerType.String()
		}),
		NewBoolFilterField(filterFieldKeyContainer, i18n.Text("be a container"), (*Equipment).Container),
	}
}

// EquipmentModifierFilterFields returns the fields a saved filter for an equipment modifier list may test.
func EquipmentModifierFilterFields() []*FilterField[*EquipmentModifier] {
	return []*FilterField[*EquipmentModifier]{
		NewTextFilterField(filterFieldKeyName, i18n.Text("have a name"), (*EquipmentModifier).NameWithReplacements),
		NewTextFilterField(filterFieldKeyNotes, i18n.Text("have notes"),
			(*EquipmentModifier).LocalNotesWithReplacements).WithPluralTitle(),
		NewListFilterField(filterFieldKeyTags, i18n.Text("have tags"), (*EquipmentModifier).TagList).WithPluralTitle(),
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
		NewListFilterField(filterFieldKeyTags, i18n.Text("have tags"), (*Note).TagList).WithPluralTitle(),
		NewTextFilterField(filterFieldKeyReference, i18n.Text("have a page reference"),
			func(n *Note) string { return n.PageRef }),
		NewBoolFilterField(filterFieldKeyContainer, i18n.Text("be a container"), (*Note).Container),
	}
}
