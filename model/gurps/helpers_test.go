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
	"sync/atomic"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
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

// newSkillBonusTo returns a flat bonus to the skill with the given name.
func newSkillBonusTo(name string, amount fxp.Int) *SkillBonus {
	bonus := NewSkillBonus()
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

// Handle takes the record by value because the slog.Handler interface requires that signature.
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
