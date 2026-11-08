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
	"context"
	"log/slog"
	"maps"
	"slices"
	"sync/atomic"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// checkFeaturesJSONRoundTrip verifies that the features survive a JSON round-trip: they marshal, unmarshal back into
// the same number of features, and re-marshal to byte-identical JSON.
func checkFeaturesJSONRoundTrip(c check.Checker, original Features) {
	c.Helper()
	data, err := jio.Marshal(original)
	c.NoError(err)
	var restored Features
	c.NoError(jio.Unmarshal(data, &restored))
	c.Equal(len(original), len(restored), "feature count preserved")
	again, err := jio.Marshal(restored)
	c.NoError(err)
	c.Equal(string(data), string(again), "re-marshaled JSON is stable across a round-trip")
}

// addTraitWithFeatures creates a non-container trait with the given name and features, appends it to the entity's
// traits, and returns it. The entity is not recalculated here.
func addTraitWithFeatures(e *Entity, name string, features ...Feature) *Trait {
	trait := NewTrait(e, nil, false)
	trait.Name = name
	trait.Features = features
	e.Traits = append(e.Traits, trait)
	return trait
}

// addCarriedEquipmentWithFeatures creates a non-container piece of equipment with the given name and features, appends
// it to the entity's carried equipment, and returns it. The entity is not recalculated here.
func addCarriedEquipmentWithFeatures(e *Entity, name string, features ...Feature) *Equipment {
	eqp := NewEquipment(e, nil, false)
	eqp.Name = name
	eqp.Features = features
	e.CarriedEquipment = append(e.CarriedEquipment, eqp)
	return eqp
}

// newSwitchableItemSet creates one non-container item of each kind that can carry a feature switch -- a "Gadget" trait,
// a "Brawling" skill, a "Fireball" spell and an "Amulet" piece of carried equipment -- gives each the supplied
// features, appends them to the entity's lists and returns them. The four items share the feature instances handed in,
// which suits the switch tests, since none of them alter a feature. The entity is not recalculated here.
func newSwitchableItemSet(e *Entity, features ...Feature) (*Trait, *Skill, *Spell, *Equipment) {
	trait := addTraitWithFeatures(e, "Gadget", features...)
	skill := NewSkill(e, nil, false)
	skill.Name = "Brawling"
	skill.Features = features
	e.Skills = append(e.Skills, skill)
	spell := NewSpell(e, nil, false)
	spell.Name = "Fireball"
	spell.Features = features
	e.Spells = append(e.Spells, spell)
	return trait, skill, spell, addCarriedEquipmentWithFeatures(e, "Amulet", features...)
}

// addSkillPrereq appends to the list a prerequisite for a skill with the given name at or above the given level, and
// returns it.
func addSkillPrereq(list *PrereqList, name string, level fxp.Int) *SkillPrereq {
	p := NewSkillPrereq()
	p.Parent = list
	p.NameCriteria.Qualifier = name
	p.LevelCriteria.Qualifier = level
	list.Prereqs = append(list.Prereqs, p)
	return p
}

// addScriptPrereq appends to the list a prerequisite met when the script's result is, and returns it.
func addScriptPrereq(list *PrereqList, script string) *ScriptPrereq {
	p := NewScriptPrereq()
	p.Parent = list
	p.Script = script
	list.Prereqs = append(list.Prereqs, p)
	return p
}

// newSkillBonusTo returns a flat bonus to the skill with the given name.
func newSkillBonusTo(name string, amount fxp.Int) *SkillBonus {
	bonus := NewSkillBonus()
	bonus.NameCriteria.Qualifier = name
	bonus.Amount = amount
	return bonus
}

// newTraitBonusTo returns a bonus of the given levels to the trait with the given name.
func newTraitBonusTo(name string, amount fxp.Int) *TraitBonus {
	bonus := NewTraitBonus()
	bonus.NameCriteria.Qualifier = name
	bonus.Amount = amount
	return bonus
}

// logCountingHandler counts the records at or above its level and prints nothing, so a test can observe whether
// something was logged without depending on the log's textual format, and can silence what it expects to be logged.
type logCountingHandler struct {
	count *atomic.Int32
	level slog.Level
}

func (h logCountingHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h logCountingHandler) Handle(_ context.Context, record slog.Record) error { //nolint:gocritic // interface-mandated signature
	if record.Level >= h.level {
		h.count.Add(1)
	}
	return nil
}

func (h logCountingHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }

func (h logCountingHandler) WithGroup(_ string) slog.Handler { return h }

// countLogs replaces the default logger for the rest of the test with one that only counts the records at or above
// the given level, and returns the count.
func countLogs(t *testing.T, level slog.Level) *atomic.Int32 {
	t.Helper()
	var count atomic.Int32
	prev := slog.Default()
	slog.SetDefault(slog.New(logCountingHandler{count: &count, level: level}))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &count
}

// reloadInto returns a loader that unmarshals the JSON it is given into a freshly zeroed value of T and hands back the
// part of that value the caller cares about, as chosen by project. The zero value is created inside the loader, so the
// same loader can be called repeatedly without one load seeing the leftovers of another.
func reloadInto[T, R any](project func(*T) R) func(data []byte) (R, error) {
	return func(data []byte) (R, error) {
		var v T
		err := jio.Unmarshal(data, &v)
		return project(&v), err
	}
}

// bonusTagsCriteria returns the tags criteria of one of the bonus types that migrated from "category" to "tags".
func bonusTagsCriteria(c check.Checker, f Feature) criteria.Text {
	switch b := f.(type) {
	case *SkillBonus:
		return b.TagsCriteria
	case *SkillPointBonus:
		return b.TagsCriteria
	case *SpellBonus:
		return b.TagsCriteria
	case *SpellPointBonus:
		return b.TagsCriteria
	case *WeaponBonus:
		return b.TagsCriteria
	default:
		c.Errorf("unexpected feature type %T", f)
		return criteria.Text{}
	}
}

// loadSingleFeature loads a feature list holding a single feature from its JSON and returns that feature.
func loadSingleFeature(c check.Checker, data string) Feature {
	var loaded Features
	c.NoError(jio.Unmarshal([]byte(data), &loaded), "%s: should load", data)
	c.Equal(1, len(loaded), "%s: a single feature should load", data)
	if len(loaded) != 1 {
		return nil
	}
	return loaded[0]
}

// nameableKeysOf returns the nameable keys the feature contributes, sorted.
func nameableKeysOf(f Feature) []string {
	m := make(map[string]string)
	f.FillWithNameableKeys(m, nil)
	return slices.Sorted(maps.Keys(m))
}

// stepLevelUpAndDown drives IncrementSkillLevel and DecrementSkillLevel through the expected point ladders, checking
// after each step that exactly one level was gained or lost and that the points landed on the expected value.
func stepLevelUpAndDown(t *testing.T, p SkillAdjustmentProvider, level func() Level, up, down []fxp.Int) {
	t.Helper()
	c := check.New(t)
	for _, expected := range up {
		before := level().Level
		p.IncrementSkillLevel()
		c.Equal(expected, p.RawPoints(), "points after incrementing to %v", expected)
		c.Equal(before+fxp.One, level().Level, "level after incrementing to %v points", expected)
	}
	for _, expected := range down {
		before := level().Level
		p.DecrementSkillLevel()
		c.Equal(expected, p.RawPoints(), "points after decrementing to %v", expected)
		c.Equal(before-fxp.One, level().Level, "level after decrementing to %v points", expected)
	}
}

// recalculatedLevel returns what reads the item's level once the sheet is recalculated, as it is after every step the
// user takes.
func recalculatedLevel(e *Entity, level func() Level) func() Level {
	return func() Level {
		e.Recalculate()
		return level()
	}
}

// newSkillDefaultTo builds a skill-based default. When anySpecialization is true, the default carries no specialization
// criteria (matching any specialization); otherwise it requires an exact ("is") match of the given specialization,
// including the empty string for an unspecialized skill.
func newSkillDefaultTo(name, specialization string, anySpecialization bool, modifier fxp.Int) *SkillDefault {
	def := &SkillDefault{
		DefaultType: SkillID,
		Name:        criteria.Text{TextData: criteria.TextData{Compare: criteria.IsText, Qualifier: name}},
		Modifier:    modifier,
	}
	if !anySpecialization {
		def.Specialization = criteria.Text{Compare: criteria.IsText, Qualifier: specialization}
	}
	return def
}

func textCriteria(compare criteria.StringComparison, qualifier string) criteria.Text {
	return criteria.Text{Compare: compare, Qualifier: qualifier}
}

// addRitualMagic creates a Ritual Magic skill with the given specialization and points. When defaultMode is non-empty
// it also installs a default: "base" defaults to the unspecialized Ritual Magic (an exact, empty-specialization match);
// "any" defaults to any Ritual Magic skill.
func addRitualMagic(e *Entity, specialization string, points fxp.Int, defaultMode string) *Skill {
	sk := addTestSkill(e, "Ritual Magic", specialization, "", points)
	switch defaultMode {
	case "base":
		sk.Defaults = []*SkillDefault{newSkillDefaultTo("Ritual Magic", "", false, -fxp.Six)}
	case "any":
		sk.Defaults = []*SkillDefault{newSkillDefaultTo("Ritual Magic", "", true, -fxp.Six)}
	}
	return sk
}

func newEquipmentItem(name, value, weight string) *Equipment {
	e := NewEquipment(nil, nil, false)
	e.Name = name
	e.BaseValue = value
	e.BaseWeight = weight
	return e
}

// newEquipmentChoice returns an equipment choice container holding an option for each name.
func newEquipmentChoice(names ...string) *Equipment {
	choice := NewEquipmentChoiceContainer(nil, nil)
	for _, name := range names {
		option := newEquipmentItem(name, "10", "1 lb")
		option.SetParent(choice)
		choice.Children = append(choice.Children, option)
	}
	return choice
}

// newRangeTestChoice returns an equipment choice picked as given from two options, one worth $10 and weighing 1 lb, the
// other worth $30 and weighing 3 lb.
func newRangeTestChoice(pickerType picker.Type, compare criteria.NumericComparison, qualifier fxp.Int) *Equipment {
	choice := NewEquipmentChoiceContainer(nil, nil)
	choice.TemplatePicker.Type = pickerType
	choice.TemplatePicker.Qualifier.Compare = compare
	choice.TemplatePicker.Qualifier.Qualifier = qualifier
	for _, one := range []*Equipment{newEquipmentItem("Cheap", "10", "1 lb"), newEquipmentItem("Dear", "30", "3 lb")} {
		one.SetParent(choice)
		choice.Children = append(choice.Children, one)
	}
	return choice
}

// item adds a piece of equipment with the given name, value and weight to the parent.
func item(parent *Equipment, name, value, weight string) *Equipment {
	one := newEquipmentItem(name, value, weight)
	one.SetParent(parent)
	parent.Children = append(parent.Children, one)
	return one
}

// poolThreshold returns the threshold of the entity's pool attribute with the given ID whose state has the given name,
// or nil if there is none.
func poolThreshold(e *Entity, attrID, state string) *PoolThreshold {
	for _, t := range e.SheetSettings.Attributes.Set[attrID].Thresholds {
		if t.State == state {
			return t
		}
	}
	return nil
}

// newMultiplier returns an attribute bonus that multiplies the given attribute by the given factor.
func newMultiplier(attrID string, factor fxp.Int) *AttributeBonus {
	bonus := NewAttributeBonus(attrID)
	bonus.Multiply = true
	bonus.Amount = factor
	return bonus
}
