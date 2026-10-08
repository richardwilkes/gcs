// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

// Package gurpstest provides helpers for tests outside the gurps package that build gurps data. The gurps package's own
// tests cannot import it, since it imports gurps, so they keep private copies of these helpers.
package gurpstest

import (
	"context"
	"log/slog"
	"sync/atomic"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
)

// AddTraitWithFeatures creates a non-container trait with the given name and features, appends it to the entity's
// traits, and returns it. The entity is not recalculated here.
func AddTraitWithFeatures(e *gurps.Entity, name string, features ...gurps.Feature) *gurps.Trait {
	trait := gurps.NewTrait(e, nil, false)
	trait.Name = name
	trait.Features = features
	e.Traits = append(e.Traits, trait)
	return trait
}

// AddCarriedEquipmentWithFeatures creates a non-container piece of equipment with the given name and features, appends
// it to the entity's carried equipment, and returns it. The entity is not recalculated here.
func AddCarriedEquipmentWithFeatures(e *gurps.Entity, name string, features ...gurps.Feature) *gurps.Equipment {
	eqp := gurps.NewEquipment(e, nil, false)
	eqp.Name = name
	eqp.Features = features
	e.CarriedEquipment = append(e.CarriedEquipment, eqp)
	return eqp
}

// AddReaction adds a trait carrying one reaction bonus with the given situation, group and amount to the entity.
func AddReaction(e *gurps.Entity, name, situation, group string, amt fxp.Int) *gurps.Trait {
	bonus := gurps.NewReactionBonus()
	bonus.Situation = situation
	bonus.Group = group
	bonus.Amount = amt
	return AddTraitWithFeatures(e, name, bonus)
}

// AddGroupedConditionalModifier adds a trait carrying one conditional modifier bonus with the given situation, group
// and amount to the entity.
func AddGroupedConditionalModifier(e *gurps.Entity, name, situation, group string, amt fxp.Int) *gurps.Trait {
	bonus := gurps.NewConditionalModifierBonus()
	bonus.Situation = situation
	bonus.Group = group
	bonus.Amount = amt
	return AddTraitWithFeatures(e, name, bonus)
}

// NewDRBonus returns a DR bonus with the given amount, specialization and locations.
func NewDRBonus(amount fxp.Int, specialization string, locations ...string) *gurps.DRBonus {
	bonus := gurps.NewDRBonus()
	bonus.Locations = locations
	bonus.Specialization = specialization
	bonus.Amount = amount
	return bonus
}

// NewTraitNeedingMissingTrait creates a trait whose prerequisite requires another trait the entity doesn't have, so
// that recalculating the entity marks it unsatisfied. It is not added to the entity.
func NewTraitNeedingMissingTrait(e *gurps.Entity, name string) *gurps.Trait {
	return NewTraitRequiring(e, name, "Combat Reflexes")
}

// NewTraitRequiring creates a trait with the given name whose prerequisite requires a trait with the required name.
// It is not added to the entity.
func NewTraitRequiring(e *gurps.Entity, name, required string) *gurps.Trait {
	t := gurps.NewTrait(e, nil, false)
	t.Name = name
	t.Prereq = NewPrereqListRequiringTrait(required)
	return t
}

// NewPrereqListRequiringTrait returns a prerequisite list satisfied only by a trait with the given name.
func NewPrereqListRequiringTrait(name string) *gurps.PrereqList {
	return newPrereqListForTrait(name, true)
}

// NewPrereqListForbiddingTrait returns a prerequisite list satisfied only by the absence of a trait with the given
// name.
func NewPrereqListForbiddingTrait(name string) *gurps.PrereqList {
	return newPrereqListForTrait(name, false)
}

func newPrereqListForTrait(name string, has bool) *gurps.PrereqList {
	list := gurps.NewPrereqList()
	p := gurps.NewTraitPrereq()
	p.Parent = list
	p.Has = has
	p.NameCriteria.Qualifier = name
	list.Prereqs = append(list.Prereqs, p)
	return list
}

// WithGroupContainersOnSort sets the general "group containers when sorting" setting for the duration of the test.
func WithGroupContainersOnSort(t *testing.T, on bool) {
	t.Helper()
	general := gurps.GlobalSettings().General
	was := general.GroupContainersOnSort
	general.GroupContainersOnSort = on
	t.Cleanup(func() { general.GroupContainersOnSort = was })
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

// CountLogs replaces the default logger for the rest of the test with one that only counts the records at or above
// the given level, and returns the count.
func CountLogs(t *testing.T, level slog.Level) *atomic.Int32 {
	t.Helper()
	var count atomic.Int32
	prev := slog.Default()
	slog.SetDefault(slog.New(logCountingHandler{count: &count, level: level}))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &count
}
