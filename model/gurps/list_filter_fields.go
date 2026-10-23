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
	"fmt"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/eqcontainer"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xstrings"
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
	// FilterFieldTextList is a list of free text values, such as notes, each compared with the whole of a
	// criteria.Text, which isn't split on commas.
	FilterFieldTextList
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

	filterFieldKeyModifierNames = "modifier_names"
	filterFieldKeyModifierNotes = "modifier_notes"
	filterFieldKeyModifierTags  = "modifier_tags"
)

// FilterField describes one field of a list type that a saved filter may test: the key it is stored under, the title
// the filter editor shows for it (phrased to follow "must" or "must not"), whether that title names something plural,
// such as "have notes", its kind, and the accessor for its kind, which is the only one of the five that is set. Both
// list kinds use List.
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
	// ListWithDisabled, when set, returns the values of a list field drawn from modifiers, disabled ones included, for
	// a condition that includes them; List then returns those of the enabled modifiers only.
	ListWithDisabled func(T) []string
	// Suggestions returns the values worth offering for a condition that compares the field with a single value, drawn
	// from list, the nodes of the list being filtered, or fixed. A field drawn from modifiers offers those of the
	// disabled ones too when includeDisabled is set. It is nil for a field that offers none.
	Suggestions func(list []T, includeDisabled bool) []string
}

// WithPluralTitle marks the field's title as naming something plural, so that what follows it agrees, as in "have notes
// that contain", and returns the field.
func (f *FilterField[T]) WithPluralTitle() *FilterField[T] {
	f.Plural = true
	return f
}

// WithWholeValues makes a list field compare each of its values with the whole of a condition's text, rather than with
// each of its comma-separated parts, for values of free text such as notes, and returns the field.
func (f *FilterField[T]) WithWholeValues() *FilterField[T] {
	f.Kind = FilterFieldTextList
	return f
}

// WithPresence sets what reports whether a node has the field at all, and returns the field.
func (f *FilterField[T]) WithPresence(has func(T) bool) *FilterField[T] {
	f.Has = has
	return f
}

// WithListedValues makes the field suggest the distinct values it holds across the list being filtered, and returns
// the field.
func (f *FilterField[T]) WithListedValues() *FilterField[T] {
	f.Suggestions = func(list []T, includeDisabled bool) []string {
		return listedFilterValues(f, list, nil, includeDisabled)
	}
	return f
}

// WithChoices makes the field suggest the values choices returns, whatever the list being filtered holds, and returns
// the field.
func (f *FilterField[T]) WithChoices(choices func() []string) *FilterField[T] {
	f.Suggestions = func([]T, bool) []string { return choices() }
	return f
}

// ListFilterTags returns the tags held by the nodes of list, at any depth, that pass the filter, gathered as the tags
// field's suggestions are. A nil filter passes every node. It returns nil when fields has no tags field.
func ListFilterTags[T Node[T]](f *ListFilter, fields []*FilterField[T], list []T) []string {
	for _, field := range fields {
		if field.Key == filterFieldKeyTags {
			return listedFilterValues(field, list, NewListFilterMatcher(f, fields), false)
		}
	}
	return nil
}

// listedFilterValues returns the values the field holds across list and everything below it, which are the nodes a
// filter checks, made distinct by distinctFilterValues and sorted naturally without regard to case. A field drawn from
// modifiers offers those of the disabled ones too when includeDisabled is set. When keep is not nil, only the nodes it
// accepts contribute. A list value that an "is" condition naming it couldn't match, such as one holding a comma or a
// nameable marker with choices, is left out.
func listedFilterValues[T Node[T]](field *FilterField[T], list []T, keep func(T) bool, includeDisabled bool) []string {
	var values []string
	Traverse(func(node T) bool {
		if (field.Has != nil && !field.Has(node)) || (keep != nil && !keep(node)) {
			return false
		}
		switch field.Kind {
		case FilterFieldText:
			values = append(values, field.Text(node))
		case FilterFieldList:
			get := field.List
			if includeDisabled && field.ListWithDisabled != nil {
				get = field.ListWithDisabled
			}
			for _, value := range get(node) {
				if (criteria.Text{Compare: criteria.IsText, Qualifier: value}).MatchesList(nil, value) {
					values = append(values, value)
				}
			}
		default:
		}
		return false
	}, false, false, list...)
	values = distinctFilterValues(values)
	slices.SortFunc(values, func(a, b string) int { return xstrings.NaturalCmp(a, b, true) })
	return values
}

// distinctFilterValues returns the values trimmed, without the blank ones, and without those that differ from an
// earlier one only in case.
func distinctFilterValues(values []string) []string {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if key := strings.ToLower(value); !seen[key] {
			seen[key] = true
			result = append(result, value)
		}
	}
	return result
}

// enumStrings returns the String() of each of the values, in order.
func enumStrings[E fmt.Stringer](values []E) []string {
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = value.String()
	}
	return result
}

// difficultyChoices returns the difficulties as the difficulty fields show them for a skill or spell that belongs to no
// entity: each attribute its editor offers, which includes 10, at each level, followed by the bare levels of a
// technique.
func difficultyChoices() []string {
	attrs, _ := AttributeChoices(nil, "", TenFlag, "10")
	choices := make([]string, 0, len(attrs)*len(difficulty.Levels)+len(difficulty.TechniqueLevels))
	for _, attr := range attrs {
		for _, level := range difficulty.Levels {
			choices = append(choices, (&AttributeDifficulty{Attribute: attr.Key, Difficulty: level}).Description(nil))
		}
	}
	for _, level := range difficulty.TechniqueLevels {
		choices = append(choices, (&AttributeDifficulty{Difficulty: level}).Description(nil))
	}
	// Two attributes may share a name.
	return distinctFilterValues(choices)
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

// NewModifierFilterField creates a list field holding the values that value returns for the modifiers that mods
// returns, containers left out: those of the enabled modifiers, or of all of them for a condition that includes
// disabled ones. Values that are empty or only space are left out, as NewListFilterField does.
func NewModifierFilterField[T Node[T], M Node[M]](key, title string, mods func(T) []M, value func(M) []string) *FilterField[T] {
	collect := func(onlyEnabled bool) func(T) []string {
		return func(node T) []string {
			var values []string
			Traverse(func(mod M) bool {
				for _, one := range value(mod) {
					if strings.TrimSpace(one) != "" {
						values = append(values, one)
					}
				}
				return false
			}, onlyEnabled, true, mods(node)...)
			return values
		}
	}
	return &FilterField[T]{
		Key:              key,
		Title:            title,
		Plural:           true,
		Kind:             FilterFieldList,
		List:             collect(true),
		ListWithDisabled: collect(false),
	}
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
		NewListFilterField(filterFieldKeyTags, i18n.Text("have tags"), (*Trait).TagList).WithPluralTitle().
			WithListedValues(),
		NewTextFilterField(filterFieldKeyReference, i18n.Text("have a page reference"),
			func(t *Trait) string { return t.PageRef }),
		NewModifierFilterField(filterFieldKeyModifierNames, i18n.Text("have modifier names"), traitModifiers,
			func(m *TraitModifier) []string { return modifierNames(m) }).WithListedValues(),
		NewModifierFilterField(filterFieldKeyModifierNotes, i18n.Text("have modifier notes"), traitModifiers,
			func(m *TraitModifier) []string { return []string{m.LocalNotesWithReplacements()} }).WithWholeValues(),
		NewModifierFilterField(filterFieldKeyModifierTags, i18n.Text("have modifier tags"), traitModifiers,
			(*TraitModifier).TagList).WithListedValues(),
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
		}).WithChoices(func() []string { return enumStrings(container.Types) }),
		NewBoolFilterField(filterFieldKeyContainer, i18n.Text("be a container"), (*Trait).Container),
	}
}

// traitModifiers returns the modifiers of the trait and of the containers holding it, which apply to it as well.
func traitModifiers(t *Trait) []*TraitModifier { return t.AllModifiers() }

// TraitModifierFilterFields returns the fields a saved filter for a trait modifier list may test.
func TraitModifierFilterFields() []*FilterField[*TraitModifier] {
	return []*FilterField[*TraitModifier]{
		NewTextFilterField(filterFieldKeyName, i18n.Text("have a name"), (*TraitModifier).NameWithReplacements),
		NewTextFilterField(filterFieldKeyNotes, i18n.Text("have notes"),
			(*TraitModifier).LocalNotesWithReplacements).WithPluralTitle(),
		NewListFilterField(filterFieldKeyTags, i18n.Text("have tags"), (*TraitModifier).TagList).WithPluralTitle().
			WithListedValues(),
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
		NewTextFilterField("optional_specialization", i18n.Text("have an optional specialization"),
			(*Skill).OptionalSpecializationWithReplacements),
		NewTextFilterField(filterFieldKeyNotes, i18n.Text("have notes"),
			(*Skill).LocalNotesWithReplacements).WithPluralTitle(),
		NewListFilterField(filterFieldKeyTags, i18n.Text("have tags"), (*Skill).TagList).WithPluralTitle().
			WithListedValues(),
		NewTextFilterField(filterFieldKeyReference, i18n.Text("have a page reference"),
			func(s *Skill) string { return s.PageRef }),
		NewTextFilterField("difficulty", i18n.Text("have a difficulty"), func(s *Skill) string {
			if s.Container() {
				return ""
			}
			return s.Difficulty.Description(EntityFromNode(s))
		}).WithChoices(difficultyChoices),
		NewNumberFilterField(filterFieldKeyPoints, i18n.Text("have points"), (*Skill).RawPoints).WithPluralTitle(),
		NewBoolFilterField(filterFieldKeyTechLevel, i18n.Text("require a tech level"), (*Skill).RequiresTL),
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
		NewListFilterField(filterFieldKeyTags, i18n.Text("have tags"), (*Spell).TagList).WithPluralTitle().
			WithListedValues(),
		NewTextFilterField(filterFieldKeyReference, i18n.Text("have a page reference"),
			func(s *Spell) string { return s.PageRef }),
		NewTextFilterField("difficulty", i18n.Text("have a difficulty"), func(s *Spell) string {
			if s.Container() {
				return ""
			}
			return s.Difficulty.Description(EntityFromNode(s))
		}).WithChoices(difficultyChoices),
		NewListFilterField("college", i18n.Text("have colleges"), (*Spell).CollegeWithReplacements).WithPluralTitle().
			WithListedValues(),
		NewTextFilterField("power_source", i18n.Text("have a power source"),
			(*Spell).PowerSourceWithReplacements).WithListedValues(),
		NewTextFilterField("class", i18n.Text("have a class"), (*Spell).ClassWithReplacements).WithListedValues(),
		NewTextFilterField("resist", i18n.Text("have a resistance"),
			(*Spell).ResistWithReplacements).WithListedValues(),
		NewTextFilterField("casting_cost", i18n.Text("have a casting cost"), (*Spell).CastingCostWithReplacements),
		NewTextFilterField("maintenance_cost", i18n.Text("have a maintenance cost"),
			(*Spell).MaintenanceCostWithReplacements),
		NewTextFilterField("casting_time", i18n.Text("have a casting time"), (*Spell).CastingTimeWithReplacements),
		NewTextFilterField("duration", i18n.Text("have a duration"), (*Spell).DurationWithReplacements),
		NewNumberFilterField(filterFieldKeyPoints, i18n.Text("have points"), (*Spell).RawPoints).WithPluralTitle(),
		NewBoolFilterField(filterFieldKeyTechLevel, i18n.Text("require a tech level"), (*Spell).RequiresTL),
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
		NewListFilterField(filterFieldKeyTags, i18n.Text("have tags"), (*Equipment).TagList).WithPluralTitle().
			WithListedValues(),
		NewTextFilterField(filterFieldKeyReference, i18n.Text("have a page reference"),
			func(e *Equipment) string { return e.PageRef }),
		NewTextFilterField(filterFieldKeyTechLevel, i18n.Text("have a tech level"),
			func(e *Equipment) string { return e.TechLevel }).WithListedValues(),
		// The raw legality class is what the LC column shows, so it is what a user will type.
		NewTextFilterField("legality_class", i18n.Text("have a legality class"),
			func(e *Equipment) string { return e.LegalityClass }).WithListedValues(),
		NewModifierFilterField(filterFieldKeyModifierNames, i18n.Text("have modifier names"), equipmentModifiers,
			func(m *EquipmentModifier) []string { return modifierNames(m) }).WithListedValues(),
		NewModifierFilterField(filterFieldKeyModifierNotes, i18n.Text("have modifier notes"), equipmentModifiers,
			func(m *EquipmentModifier) []string { return []string{m.LocalNotesWithReplacements()} }).WithWholeValues(),
		NewModifierFilterField(filterFieldKeyModifierTags, i18n.Text("have modifier tags"), equipmentModifiers,
			(*EquipmentModifier).TagList).WithListedValues(),
		// A group's quantity is always 1, and isn't shown.
		NewNumberFilterField("quantity", i18n.Text("have a quantity"), func(e *Equipment) fxp.Int { return e.Quantity }).
			WithPresence(func(e *Equipment) bool { return !e.IsGroup() }),
		NewNumberFilterField(filterFieldKeyCost, i18n.Text("have a cost"), (*Equipment).AdjustedValue),
		NewWeightFilterField(filterFieldKeyWeight, i18n.Text("have a weight"), func(e *Equipment) fxp.Weight {
			return e.AdjustedWeight(false, SheetSettingsFor(EntityFromNode(e)).DefaultWeightUnits)
		}),
		NewNumberFilterField("contained_quantity", i18n.Text("have a contained quantity"),
			func(e *Equipment) fxp.Int { return containedQuantity(e.Children) }).WithPresence((*Equipment).Container),
		NewWeightFilterField("contained_weight", i18n.Text("have a contained weight"), func(e *Equipment) fxp.Weight {
			return e.ContainedWeight(false, SheetSettingsFor(EntityFromNode(e)).DefaultWeightUnits)
		}).WithPresence((*Equipment).Container),
		NewTextFilterField("container_type", i18n.Text("have a container type"), func(e *Equipment) string {
			if !e.Container() {
				return ""
			}
			return e.ContainerType.String()
		}).WithChoices(func() []string { return enumStrings(eqcontainer.Types) }),
		NewBoolFilterField(filterFieldKeyContainer, i18n.Text("be a container"), (*Equipment).Container),
	}
}

func equipmentModifiers(e *Equipment) []*EquipmentModifier { return e.Modifiers }

// EquipmentModifierFilterFields returns the fields a saved filter for an equipment modifier list may test.
func EquipmentModifierFilterFields() []*FilterField[*EquipmentModifier] {
	return []*FilterField[*EquipmentModifier]{
		NewTextFilterField(filterFieldKeyName, i18n.Text("have a name"), (*EquipmentModifier).NameWithReplacements),
		NewTextFilterField(filterFieldKeyNotes, i18n.Text("have notes"),
			(*EquipmentModifier).LocalNotesWithReplacements).WithPluralTitle(),
		NewListFilterField(filterFieldKeyTags, i18n.Text("have tags"), (*EquipmentModifier).TagList).WithPluralTitle().
			WithListedValues(),
		NewTextFilterField(filterFieldKeyReference, i18n.Text("have a page reference"),
			func(e *EquipmentModifier) string { return e.PageRef }),
		NewTextFilterField(filterFieldKeyTechLevel, i18n.Text("have a tech level"),
			func(e *EquipmentModifier) string { return e.TechLevel }).WithListedValues(),
		NewTextFilterField(filterFieldKeyCost, i18n.Text("have a cost"), (*EquipmentModifier).CostDescription),
		NewTextFilterField(filterFieldKeyWeight, i18n.Text("have a weight"), (*EquipmentModifier).WeightDescription),
		NewBoolFilterField(filterFieldKeyContainer, i18n.Text("be a container"), (*EquipmentModifier).Container),
	}
}

// NoteFilterFields returns the fields a saved filter for a note list may test.
func NoteFilterFields() []*FilterField[*Note] {
	return []*FilterField[*Note]{
		NewTextFilterField("text", i18n.Text("have text"), (*Note).TextWithReplacements),
		NewListFilterField(filterFieldKeyTags, i18n.Text("have tags"), (*Note).TagList).WithPluralTitle().
			WithListedValues(),
		NewTextFilterField(filterFieldKeyReference, i18n.Text("have a page reference"),
			func(n *Note) string { return n.PageRef }),
		NewBoolFilterField(filterFieldKeyContainer, i18n.Text("be a container"), (*Note).Container),
	}
}
