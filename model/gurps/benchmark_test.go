// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/gurpstest"
)

// BenchmarkFileEnvVar names the environment variable that, when set to the path of a character sheet, has the
// benchmarks measure that sheet instead of the one gurpstest.NewLargeEntity builds.
const BenchmarkFileEnvVar = "GCS_BENCH_FILE"

// The benchmarks below measure the model-side work a sheet does when it is updated, on a large character. Run them
// with:
//
//	GOEXPERIMENT=simd go test ./model/gurps -run '^$' -bench . -benchmem
//
// gurpstest.LargeEntitySizeEnvVar scales the generated character; BenchmarkFileEnvVar substitutes a real sheet.

// meleePageColumns and rangedPageColumns are the columns the sheet's attack tables show.
var (
	meleePageColumns = []int{
		gurps.WeaponDescriptionColumn, gurps.WeaponUsageColumn, gurps.WeaponSLColumn, gurps.WeaponParryColumn,
		gurps.WeaponBlockColumn, gurps.WeaponDamageColumn, gurps.WeaponReachColumn, gurps.WeaponSTColumn,
	}
	rangedPageColumns = []int{
		gurps.WeaponDescriptionColumn, gurps.WeaponUsageColumn, gurps.WeaponSLColumn, gurps.WeaponAccColumn,
		gurps.WeaponDamageColumn, gurps.WeaponRangeColumn, gurps.WeaponRoFColumn, gurps.WeaponShotsColumn,
		gurps.WeaponBulkColumn, gurps.WeaponRecoilColumn, gurps.WeaponSTColumn,
	}
	spellPageColumns = []int{
		gurps.SpellDescriptionForPageColumn, gurps.SpellLevelColumn, gurps.SpellRelativeLevelColumn,
		gurps.SpellPointsColumn, gurps.SpellReferenceColumn,
	}
	skillPageColumns = []int{
		gurps.SkillDescriptionColumn, gurps.SkillLevelColumn, gurps.SkillRelativeLevelColumn, gurps.SkillPointsColumn,
		gurps.SkillReferenceColumn,
	}
	traitPageColumns     = []int{gurps.TraitDescriptionColumn, gurps.TraitPointsColumn, gurps.TraitReferenceColumn}
	equipmentPageColumns = []int{
		gurps.EquipmentEquippedColumn, gurps.EquipmentQuantityColumn, gurps.EquipmentDescriptionColumn,
		gurps.EquipmentCostColumn, gurps.EquipmentWeightColumn,
		gurps.EquipmentExtendedCostColumn, gurps.EquipmentExtendedWeightColumn, gurps.EquipmentReferenceColumn,
	}
)

// benchEntity returns the character the benchmarks measure: the sheet named by BenchmarkFileEnvVar, or a generated
// one of the size gurpstest.LargeEntitySizeFromEnv reports.
func benchEntity(b *testing.B) *gurps.Entity {
	b.Helper()
	if path := os.Getenv(BenchmarkFileEnvVar); path != "" {
		e, err := gurps.NewEntityFromFile(os.DirFS(filepath.Dir(path)), filepath.Base(path))
		if err != nil {
			b.Fatalf("unable to load %s: %v", path, err)
		}
		return e
	}
	size, err := gurpstest.LargeEntitySizeFromEnv()
	if err != nil {
		b.Fatal(err)
	}
	return gurpstest.NewLargeEntity(&size)
}

// cellDataFor computes the cell data of every row for each column, which is what a table does when it syncs.
func cellDataFor[T gurps.Node[T]](rows []T, columns []int) {
	for _, row := range rows {
		for _, col := range columns {
			data := gurps.CellData{ForPage: true}
			row.CellData(col, &data)
		}
	}
}

// leaves returns the non-container rows of the list, in order.
func leaves[T gurps.Node[T]](list []T) []T {
	var result []T
	gurps.Traverse(func(one T) bool {
		result = append(result, one)
		return false
	}, false, true, list...)
	return result
}

// coldAndWarm runs fn as two sub-benchmarks: "cold" discards the entity's caches before every iteration, as a
// recalculation does, and "warm" keeps whatever fn left behind.
func coldAndWarm(b *testing.B, e *gurps.Entity, fn func()) {
	b.Helper()
	b.Run("cold", func(b *testing.B) {
		for b.Loop() {
			e.DiscardCaches()
			fn()
		}
	})
	b.Run("warm", func(b *testing.B) {
		for b.Loop() {
			fn()
		}
	})
}

func BenchmarkEntityRecalculate(b *testing.B) {
	e := benchEntity(b)
	for b.Loop() {
		e.Recalculate()
	}
}

func BenchmarkEntityWeapons(b *testing.B) {
	e := benchEntity(b)
	coldAndWarm(b, e, func() {
		e.Weapons(true, false, true)
		e.Weapons(false, false, true)
	})
}

func BenchmarkWeaponCellData(b *testing.B) {
	e := benchEntity(b)
	melee := e.Weapons(true, false, true)
	ranged := e.Weapons(false, false, true)
	coldAndWarm(b, e, func() {
		cellDataFor(melee, meleePageColumns)
		cellDataFor(ranged, rangedPageColumns)
	})
}

func BenchmarkSpellCellData(b *testing.B) {
	e := benchEntity(b)
	spells := leaves(e.Spells)
	coldAndWarm(b, e, func() { cellDataFor(spells, spellPageColumns) })
}

func BenchmarkSkillCellData(b *testing.B) {
	e := benchEntity(b)
	skills := leaves(e.Skills)
	coldAndWarm(b, e, func() { cellDataFor(skills, skillPageColumns) })
}

func BenchmarkTraitCellData(b *testing.B) {
	e := benchEntity(b)
	traits := leaves(e.Traits)
	coldAndWarm(b, e, func() { cellDataFor(traits, traitPageColumns) })
}

func BenchmarkEquipmentCellData(b *testing.B) {
	e := benchEntity(b)
	carried := leaves(e.CarriedEquipment)
	other := leaves(e.OtherEquipment)
	coldAndWarm(b, e, func() {
		cellDataFor(carried, equipmentPageColumns)
		cellDataFor(other, equipmentPageColumns)
	})
}

func BenchmarkSkillNamed(b *testing.B) {
	e := benchEntity(b)
	lookups := []struct {
		name           string
		specialization string
	}{
		{"Broadsword", ""},
		{"Ritual Magic", "Fire"},
		{"Guns", "Pistol"},
		{"Thaumatology", ""},
		{"No Such Skill", ""},
	}
	b.Run("named", func(b *testing.B) {
		for b.Loop() {
			for _, one := range lookups {
				e.SkillNamed(one.name, one.specialization, false, nil)
			}
		}
	})
	b.Run("matching", func(b *testing.B) {
		for b.Loop() {
			for _, one := range lookups {
				name := criteria.Text{Compare: criteria.IsText, Qualifier: one.name}
				var spec criteria.Text
				if one.specialization != "" {
					spec = criteria.Text{Compare: criteria.IsText, Qualifier: one.specialization}
				}
				e.SkillMatching(name, spec, nil, false, nil)
			}
		}
	})
}

// BenchmarkModelSync is the model-side equivalent of a sheet update: a recalculation followed by everything the
// tables ask the model for when they sync.
func BenchmarkModelSync(b *testing.B) {
	e := benchEntity(b)
	traits := leaves(e.Traits)
	skills := leaves(e.Skills)
	spells := leaves(e.Spells)
	carried := leaves(e.CarriedEquipment)
	other := leaves(e.OtherEquipment)
	for b.Loop() {
		e.Recalculate()
		cellDataFor(e.Weapons(true, false, true), meleePageColumns)
		cellDataFor(e.Weapons(false, false, true), rangedPageColumns)
		cellDataFor(traits, traitPageColumns)
		cellDataFor(skills, skillPageColumns)
		cellDataFor(spells, spellPageColumns)
		cellDataFor(carried, equipmentPageColumns)
		cellDataFor(other, equipmentPageColumns)
	}
}
