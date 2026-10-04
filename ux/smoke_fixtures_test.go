// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

//go:build smoke

package ux

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/affects"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emcost"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emweight"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/eqcontainer"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/equipmentsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/frequency"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/maxusesmod"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selector"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/skillsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellcmp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellmatch"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stdmg"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/study"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/traitsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
)

// The fixtures in testdata/smoke are meant to hold at least one of every kind of data GCS reads, so that the smoke
// tests exercise all of it. TestSmokeFixtureCoverage keeps that true: it loads every fixture with the loader GCS uses
// for its file type, then checks that every value of the enums that shape the data, and every structural case listed
// in fixtureCases, turns up somewhere in what was loaded. When GCS gains a new kind of data, the test fails until a
// fixture includes it.
//
// The fixtures are written in the form GCS itself saves them in, so loading and saving one through GCS changes
// nothing. Edit them by hand, or open them in GCS, change them and save.

// fixtureCases lists the structural cases the fixtures must hold that no enum describes.
var fixtureCases = []string{
	"trait container nested", "trait leveled", "trait round down", "trait disabled", "trait switched on",
	"trait preconfigured", "trait nameable filled", "trait nameable open", "trait modifiers", "trait choice nested",
	"trait pick separately", "trait third party data", "modifier group", "modifier choice mandatory",
	"modifier choice optional", "modifier disabled", "skill container nested", "skill specialization",
	"skill tech level", "skill tech level unset", "technique", "technique limit", "skill defaults",
	"spell multiple colleges", "ritual magic spell", "spell container", "equipment container nested",
	"equipment group", "equipment uses", "equipment modifiers", "equipment modifier group",
	"equipment modifier choice mandatory", "equipment modifier choice optional", "equipment not equipped",
	"note container", "note nameable", "weapon melee", "weapon ranged", "weapon hidden", "prereq list any",
	"prereq when tl", "sheet portrait", "sheet points record", "sheet other equipment", "sheet pool damage",
	"template body type", "template ancestry", "loot", "output template html", "output template text",
	"output template legacy", "name generator simple", "name generator markov_letter",
	"name generator markov_run", "name generator compound",
}

// fixtureFileTypes lists the file types the fixtures must include an example of.
var fixtureFileTypes = []string{
	gurps.SheetExt, gurps.TemplatesExt, gurps.LootExt, gurps.TraitsExt, gurps.TraitModifiersExt,
	gurps.SkillsExt, gurps.SpellsExt, gurps.EquipmentExt, gurps.EquipmentModifiersExt, gurps.NotesExt,
	gurps.AncestryExt, gurps.NamesExt, gurps.AttributesExt, gurps.BodyExt, gurps.CalendarExt,
	gurps.ColorSettingsExt, gurps.FontSettingsExt, gurps.GeneralSettingsExt, gurps.KeySettingsExt,
	gurps.PageRefSettingsExt, gurps.SheetSettingsExt, ".md",
}

// TestSmokeFixtureCoverage checks that every fixture loads, and that together they hold every kind of data GCS reads.
func TestSmokeFixtureCoverage(t *testing.T) {
	c := &fixtureCoverage{t: t, seen: make(map[string]map[string]bool)}
	fsys := os.DirFS(smokeFixtureDir)
	for _, dir := range []string{"master_library", "user_library", "files"} {
		if err := fs.WalkDir(fsys, dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || d.Name() == ".gitkeep" {
				return err
			}
			c.load(fsys, p)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	c.require("file type", fixtureFileTypes)
	c.require("case", fixtureCases)
	c.requireEnum("feature", without(feature.Types, feature.Unknown))
	c.requireEnum("prereq", prereq.Types)
	c.requireEnum("trait container type", container.Types)
	c.requireEnum("equipment container type", eqcontainer.Types)
	c.requireEnum("trait choice", without(picker.TypesForTraits, picker.NotApplicable))
	c.requireEnum("skill choice", without(picker.TypesForSkills, picker.NotApplicable))
	c.requireEnum("spell choice", without(picker.TypesForSpells, picker.NotApplicable))
	c.requireEnum("equipment choice", without(picker.TypesForEquipment, picker.NotApplicable))
	c.requireEnum("modifier affects", affects.Options)
	c.requireEnum("equipment modifier cost", emcost.Types)
	c.requireEnum("equipment modifier weight", emweight.Types)
	c.requireEnum("study", study.Types)
	c.requireEnum("study hours", study.Levels)
	c.requireEnum("self-control adjustment", selfctrl.Adjustments)
	c.require("self-control roll", stringsOf(without(selfctrl.Rolls, 0)))
	c.require("frequency", stringsOf(without(frequency.Rolls, 0)))
	c.requireEnum("skill difficulty", difficulty.Levels)
	c.requireEnum("technique difficulty", difficulty.TechniqueLevels)
	c.requireEnum("strength damage", stdmg.Options)
	c.requireEnum("strength limitation", stlimit.Options)
	c.requireEnum("weapon selection", wsel.Types)
	c.requireEnum("weapon switch", wswitch.Types)
	c.requireEnum("skill selection", skillsel.Types)
	c.requireEnum("spell match", spellmatch.Types)
	c.requireEnum("spell comparison", spellcmp.Types)
	c.requireEnum("trait selection", traitsel.Types)
	c.requireEnum("equipment selection", equipmentsel.Types)
	c.requireEnum("selector field", selector.Fields)
	c.requireEnum("max uses adjustment", maxusesmod.Types)
}

// fixtureCoverage records what the fixtures were found to hold.
type fixtureCoverage struct {
	t    *testing.T
	seen map[string]map[string]bool
}

func (c *fixtureCoverage) mark(category, value string) {
	if c.seen[category] == nil {
		c.seen[category] = make(map[string]bool)
	}
	c.seen[category][value] = true
}

func (c *fixtureCoverage) markKey(category string, value interface{ Key() string }) {
	c.mark(category, value.Key())
}

func (c *fixtureCoverage) require(category string, values []string) {
	c.t.Helper()
	for _, one := range values {
		if !c.seen[category][one] {
			c.t.Errorf("no fixture has %s %q", category, one)
		}
	}
}

func requireKeys[T interface{ Key() string }](c *fixtureCoverage, category string, values []T) {
	c.t.Helper()
	keys := make([]string, len(values))
	for i, one := range values {
		keys[i] = one.Key()
	}
	c.require(category, keys)
}

func (c *fixtureCoverage) requireEnum(category string, values any) {
	c.t.Helper()
	switch list := values.(type) {
	case []feature.Type:
		requireKeys(c, category, list)
	case []prereq.Type:
		requireKeys(c, category, list)
	case []container.Type:
		requireKeys(c, category, list)
	case []eqcontainer.Type:
		requireKeys(c, category, list)
	case []picker.Type:
		requireKeys(c, category, list)
	case []affects.Option:
		requireKeys(c, category, list)
	case []emcost.Type:
		requireKeys(c, category, list)
	case []emweight.Type:
		requireKeys(c, category, list)
	case []study.Type:
		requireKeys(c, category, list)
	case []study.Level:
		requireKeys(c, category, list)
	case []selfctrl.Adjustment:
		requireKeys(c, category, list)
	case []difficulty.Level:
		requireKeys(c, category, list)
	case []stdmg.Option:
		requireKeys(c, category, list)
	case []stlimit.Option:
		requireKeys(c, category, list)
	case []wsel.Type:
		requireKeys(c, category, list)
	case []wswitch.Type:
		requireKeys(c, category, list)
	case []skillsel.Type:
		requireKeys(c, category, list)
	case []spellmatch.Type:
		requireKeys(c, category, list)
	case []spellcmp.Type:
		requireKeys(c, category, list)
	case []traitsel.Type:
		requireKeys(c, category, list)
	case []equipmentsel.Type:
		requireKeys(c, category, list)
	case []selector.Field:
		requireKeys(c, category, list)
	case []maxusesmod.Type:
		requireKeys(c, category, list)
	default:
		c.t.Fatalf("unhandled enum list %T", values)
	}
}

func without[T comparable](list []T, omit T) []T {
	return slices.DeleteFunc(slices.Clone(list), func(one T) bool { return one == omit })
}

func stringsOf[T any](list []T) []string {
	out := make([]string, len(list))
	for i, one := range list {
		out[i] = fmt.Sprint(one)
	}
	return out
}

// load loads one fixture with the loader GCS uses for its file type, failing the test if it can't be loaded, and
// records what it holds.
func (c *fixtureCoverage) load(fsys fs.FS, p string) {
	c.t.Helper()
	ext := strings.ToLower(path.Ext(p))
	c.mark("file type", ext)
	var err error
	switch ext {
	case gurps.SheetExt:
		var e *gurps.Entity
		if e, err = gurps.NewEntityFromFile(fsys, p); err == nil {
			c.entity(e)
		}
	case gurps.TemplatesExt:
		var tmpl *gurps.Template
		if tmpl, err = gurps.NewTemplateFromFile(fsys, p); err == nil {
			c.traits(tmpl.Traits)
			c.skills(tmpl.Skills)
			c.spells(tmpl.Spells)
			c.equipment(tmpl.Equipment)
			c.notes(tmpl.Notes)
			if tmpl.BodyType != nil {
				c.mark("case", "template body type")
			}
			gurps.Traverse(func(one *gurps.Trait) bool {
				if one.Container() && one.ContainerType == container.Ancestry {
					c.mark("case", "template ancestry")
				}
				return false
			}, false, false, tmpl.Traits...)
		}
	case gurps.LootExt:
		var l *gurps.Loot
		if l, err = gurps.NewLootFromFile(fsys, p); err == nil {
			c.mark("case", "loot")
			c.equipment(l.Equipment)
			c.notes(l.Notes)
		}
	case gurps.TraitsExt:
		var list []*gurps.Trait
		if list, err = gurps.NewTraitsFromFile(fsys, p); err == nil {
			c.traits(list)
		}
	case gurps.TraitModifiersExt:
		var list []*gurps.TraitModifier
		if list, err = gurps.NewTraitModifiersFromFile(fsys, p); err == nil {
			c.traitModifiers(list)
		}
	case gurps.SkillsExt:
		var list []*gurps.Skill
		if list, err = gurps.NewSkillsFromFile(fsys, p); err == nil {
			c.skills(list)
		}
	case gurps.SpellsExt:
		var list []*gurps.Spell
		if list, err = gurps.NewSpellsFromFile(fsys, p); err == nil {
			c.spells(list)
		}
	case gurps.EquipmentExt:
		var list []*gurps.Equipment
		if list, err = gurps.NewEquipmentFromFile(fsys, p); err == nil {
			c.equipment(list)
		}
	case gurps.EquipmentModifiersExt:
		var list []*gurps.EquipmentModifier
		if list, err = gurps.NewEquipmentModifiersFromFile(fsys, p); err == nil {
			c.equipmentModifiers(list)
		}
	case gurps.NotesExt:
		var list []*gurps.Note
		if list, err = gurps.NewNotesFromFile(fsys, p); err == nil {
			c.notes(list)
		}
	case gurps.AncestryExt:
		_, err = gurps.NewAncestryFromFile(fsys, p)
	case gurps.NamesExt:
		var n *gurps.NameGenerator
		if n, err = gurps.NewNameGeneratorFromFS(fsys, p); err == nil {
			c.mark("case", "name generator "+n.Type.Key())
		}
	case gurps.AttributesExt:
		_, err = gurps.NewAttributeDefsFromFile(fsys, p)
	case gurps.BodyExt:
		_, err = gurps.NewBodyFromFile(fsys, p)
	case gurps.CalendarExt:
		_, err = gurps.NewCalendarRefFromFS(fsys, p)
	case gurps.ColorSettingsExt:
		_, err = colors.NewFromFS(fsys, p)
	case gurps.FontSettingsExt:
		_, err = fonts.NewFromFS(fsys, p)
	case gurps.GeneralSettingsExt:
		_, err = gurps.NewGeneralSettingsFromFile(fsys, p)
	case gurps.KeySettingsExt:
		_, err = gurps.NewKeyBindingsFromFS(fsys, p)
	case gurps.PageRefSettingsExt:
		_, err = gurps.NewPageRefsFromFS(fsys, p)
	case gurps.SheetSettingsExt:
		_, err = gurps.NewSheetSettingsFromFile(fsys, p)
	case ".md":
	case ".html", ".txt":
		c.outputTemplate(fsys, p)
	default:
		c.t.Errorf("%s: no loader for %s files", p, ext)
	}
	if err != nil {
		c.t.Errorf("%s: %v", p, err)
	}
}

func (c *fixtureCoverage) outputTemplate(fsys fs.FS, p string) {
	c.t.Helper()
	if path.Base(path.Dir(p)) != "Output Templates" {
		c.t.Errorf("%s: not an output template", p)
		return
	}
	data, err := fs.ReadFile(fsys, p)
	if err != nil {
		c.t.Errorf("%s: %v", p, err)
		return
	}
	first, _, _ := strings.Cut(string(data), "\n")
	switch first {
	case "GCS HTML Template v1":
		c.mark("case", "output template html")
	case "GCS Text Template v1":
		c.mark("case", "output template text")
	default:
		c.mark("case", "output template legacy")
	}
}

func (c *fixtureCoverage) entity(e *gurps.Entity) {
	if len(e.Profile.PortraitData) != 0 {
		c.mark("case", "sheet portrait")
	}
	if len(e.PointsRecord) > 1 {
		c.mark("case", "sheet points record")
	}
	if len(e.OtherEquipment) != 0 {
		c.mark("case", "sheet other equipment")
	}
	for _, attr := range e.Attributes.Set {
		if attr.Damage != 0 {
			c.mark("case", "sheet pool damage")
		}
	}
	c.traits(e.Traits)
	c.skills(e.Skills)
	c.spells(e.Spells)
	c.equipment(e.CarriedEquipment)
	c.equipment(e.OtherEquipment)
	c.notes(e.Notes)
}

func (c *fixtureCoverage) traits(list []*gurps.Trait) {
	gurps.Traverse(func(one *gurps.Trait) bool {
		if one.Container() {
			if !one.TemplatePicker.IsZero() {
				c.markKey("trait choice", one.TemplatePicker.Type)
				if p := one.Parent(); p != nil && !p.TemplatePicker.IsZero() {
					c.mark("case", "trait choice nested")
				}
			} else {
				c.markKey("trait container type", one.ContainerType)
			}
			if one.PickSeparately {
				c.mark("case", "trait pick separately")
			}
			if one.Parent() != nil && one.Parent().Container() && one.TemplatePicker.IsZero() {
				c.mark("case", "trait container nested")
			}
		} else {
			if one.CanLevel {
				c.mark("case", "trait leveled")
			}
			if one.RoundCostDown {
				c.mark("case", "trait round down")
			}
			c.features(one.Features)
			c.weapons(one.Weapons)
			for _, s := range one.Study {
				c.markKey("study", s.Type)
			}
			c.markKey("study hours", one.StudyHoursNeeded)
		}
		if one.SelfControl != 0 {
			c.mark("self-control roll", one.SelfControl.String())
		}
		c.markKey("self-control adjustment", one.SelfControlAdj)
		if one.Frequency != 0 {
			c.mark("frequency", one.Frequency.String())
		}
		c.flags("trait", one.Disabled, one.SwitchedOn, one.Preconfigured)
		c.nameables("trait", one.Name, one.Replacements)
		if len(one.Modifiers) != 0 {
			c.mark("case", "trait modifiers")
			c.traitModifiers(one.Modifiers)
		}
		if len(one.ThirdParty) != 0 {
			c.mark("case", "trait third party data")
		}
		c.prereqs(one.Prereq)
		return false
	}, false, false, list...)
}

func (c *fixtureCoverage) traitModifiers(list []*gurps.TraitModifier) {
	gurps.Traverse(func(one *gurps.TraitModifier) bool {
		if one.Container() {
			c.modifierContainer("modifier", one.Choice)
		} else {
			c.markKey("modifier affects", one.Affects)
			c.features(one.Features)
			if one.Disabled {
				c.mark("case", "modifier disabled")
			}
		}
		return false
	}, false, false, list...)
}

func (c *fixtureCoverage) modifierContainer(kind string, choice gurps.TemplatePicker) {
	switch {
	case choice.IsZero():
		c.mark("case", kind+" group")
	case choice.Qualifier.Compare == criteria.AtMostNumber:
		c.mark("case", kind+" choice optional")
	default:
		c.mark("case", kind+" choice mandatory")
	}
}

func (c *fixtureCoverage) skills(list []*gurps.Skill) {
	gurps.Traverse(func(one *gurps.Skill) bool {
		switch {
		case one.Container():
			if !one.TemplatePicker.IsZero() {
				c.markKey("skill choice", one.TemplatePicker.Type)
			} else if p := one.Parent(); p != nil {
				c.mark("case", "skill container nested")
			}
		case one.IsTechnique():
			c.mark("case", "technique")
			c.markKey("technique difficulty", one.Difficulty.Difficulty)
			if one.TechniqueLimitModifier != nil {
				c.mark("case", "technique limit")
			}
		default:
			c.markKey("skill difficulty", one.Difficulty.Difficulty)
			if one.Specialization != "" {
				c.mark("case", "skill specialization")
			}
			if one.TechLevel != nil {
				if *one.TechLevel == "" {
					c.mark("case", "skill tech level unset")
				} else {
					c.mark("case", "skill tech level")
				}
			}
			if len(one.Defaults) != 0 {
				c.mark("case", "skill defaults")
			}
			for _, s := range one.Study {
				c.markKey("study", s.Type)
			}
			c.markKey("study hours", one.StudyHoursNeeded)
			c.features(one.Features)
			c.weapons(one.Weapons)
			c.prereqs(one.Prereq)
		}
		return false
	}, false, false, list...)
}

func (c *fixtureCoverage) spells(list []*gurps.Spell) {
	gurps.Traverse(func(one *gurps.Spell) bool {
		switch {
		case one.Container():
			if !one.TemplatePicker.IsZero() {
				c.markKey("spell choice", one.TemplatePicker.Type)
			} else {
				c.mark("case", "spell container")
			}
		default:
			if one.IsRitualMagic() {
				c.mark("case", "ritual magic spell")
			}
			if len(one.College) > 1 {
				c.mark("case", "spell multiple colleges")
			}
			for _, s := range one.Study {
				c.markKey("study", s.Type)
			}
			c.weapons(one.Weapons)
			c.prereqs(one.Prereq)
		}
		return false
	}, false, false, list...)
}

func (c *fixtureCoverage) equipment(list []*gurps.Equipment) {
	gurps.Traverse(func(one *gurps.Equipment) bool {
		if one.Container() {
			if !one.TemplatePicker.IsZero() {
				c.markKey("equipment choice", one.TemplatePicker.Type)
			} else {
				c.markKey("equipment container type", one.ContainerType)
				if one.ContainerType == eqcontainer.Group {
					c.mark("case", "equipment group")
				} else if p := one.Parent(); p != nil {
					c.mark("case", "equipment container nested")
				}
			}
		}
		if one.MaxUses != 0 {
			c.mark("case", "equipment uses")
		}
		if !one.Equipped {
			c.mark("case", "equipment not equipped")
		}
		if len(one.Modifiers) != 0 {
			c.mark("case", "equipment modifiers")
			c.equipmentModifiers(one.Modifiers)
		}
		c.features(one.Features)
		c.weapons(one.Weapons)
		c.prereqs(one.Prereq)
		return false
	}, false, false, list...)
}

func (c *fixtureCoverage) equipmentModifiers(list []*gurps.EquipmentModifier) {
	gurps.Traverse(func(one *gurps.EquipmentModifier) bool {
		if one.Container() {
			c.modifierContainer("equipment modifier", one.Choice)
		} else {
			c.markKey("equipment modifier cost", one.CostType)
			c.markKey("equipment modifier weight", one.WeightType)
			c.features(one.Features)
		}
		return false
	}, false, false, list...)
}

func (c *fixtureCoverage) notes(list []*gurps.Note) {
	gurps.Traverse(func(one *gurps.Note) bool {
		if one.Container() {
			c.mark("case", "note container")
		}
		if strings.ContainsRune(one.MarkDown, '@') {
			c.mark("case", "note nameable")
		}
		return false
	}, false, false, list...)
}

func (c *fixtureCoverage) flags(kind string, disabled, switchedOn, preconfigured bool) {
	if disabled {
		c.mark("case", kind+" disabled")
	}
	if switchedOn {
		c.mark("case", kind+" switched on")
	}
	if preconfigured {
		c.mark("case", kind+" preconfigured")
	}
}

func (c *fixtureCoverage) nameables(kind, text string, replacements map[string]string) {
	if !strings.ContainsRune(text, '@') {
		return
	}
	if len(replacements) != 0 {
		c.mark("case", kind+" nameable filled")
	} else {
		c.mark("case", kind+" nameable open")
	}
}

func (c *fixtureCoverage) features(list gurps.Features) {
	for _, f := range list {
		c.markKey("feature", f.FeatureType())
		switch one := f.(type) {
		case *gurps.AttributeBonus:
			c.markKey("strength limitation", one.Limitation)
		case *gurps.WeaponBonus:
			c.markKey("weapon selection", one.SelectionType)
			if one.Type == feature.WeaponSwitch {
				c.markKey("weapon switch", one.SwitchType)
			}
		case *gurps.SkillBonus:
			c.markKey("skill selection", one.SelectionType)
		case *gurps.SpellBonus:
			c.markKey("spell match", one.SpellMatchType)
		case *gurps.SpellPointBonus:
			c.markKey("spell match", one.SpellMatchType)
		case *gurps.TraitMaxLevelBonus:
			c.markKey("trait selection", one.SelectionType)
			c.markKey("max uses adjustment", one.Operation())
		case *gurps.EquipmentMaxUsesBonus:
			c.markKey("equipment selection", one.SelectionType)
			c.markKey("max uses adjustment", one.Operation())
		case *gurps.SelectorOverride:
			c.markKey("selector field", one.Field)
		}
	}
}

func (c *fixtureCoverage) prereqs(list *gurps.PrereqList) {
	if list == nil {
		return
	}
	c.markKey("prereq", list.PrereqType())
	if !list.All {
		c.mark("case", "prereq list any")
	}
	if list.WhenTL.Compare != criteria.AnyNumber {
		c.mark("case", "prereq when tl")
	}
	for _, one := range list.Prereqs {
		switch p := one.(type) {
		case *gurps.PrereqList:
			c.prereqs(p)
			continue
		case *gurps.SpellPrereq:
			c.markKey("spell comparison", p.SubType)
		}
		c.markKey("prereq", one.PrereqType())
	}
}

func (c *fixtureCoverage) weapons(list []*gurps.Weapon) {
	for _, w := range list {
		switch {
		case w.Hide:
			c.mark("case", "weapon hidden")
		case w.IsMelee():
			c.mark("case", "weapon melee")
		default:
			c.mark("case", "weapon ranged")
		}
		st := w.Damage.StrengthType
		if w.Damage.Leveled {
			switch st {
			case stdmg.Thrust:
				st = stdmg.OldLeveledThrust
			case stdmg.Swing:
				st = stdmg.OldLeveledSwing
			}
		}
		c.markKey("strength damage", st)
	}
}
