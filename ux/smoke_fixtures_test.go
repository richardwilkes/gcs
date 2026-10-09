// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/affects"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emcost"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emweight"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/eqcontainer"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/equipmentsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/frequency"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/layoutnode"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/maxusesmod"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/namegen"
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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/threshold"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/traitsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
	"github.com/richardwilkes/unison"
)

// The fixtures in testdata/smoke are meant to hold at least one of every kind of data GCS reads, so that the smoke
// tests exercise all of it. TestSmokeFixtureCoverage keeps that true: it loads every fixture with the loader GCS uses
// for its file type, then checks that every value of the enums that shape the data, and every structural case listed
// in fixtureCases, turns up somewhere in what was loaded. Every enum package in model/gurps/enums must be either
// required here or listed in fixtureExemptEnums with the reason it isn't, so a new enum fails the test until one or the
// other is done.
//
// The fixtures are also kept in the form GCS itself saves them in, so that they exercise the current file format
// rather than a migration from an older one: loading and saving each one through GCS must change nothing. Edit them by
// hand, or open them in GCS and save them, then run the test with -update-fixtures to put them back in that form, which
// is also how they are brought up to date after a change to the file format.
//
// It needs neither the smoke tag nor a headless session, so it runs with the rest of the tests.

// smokeFixtureDir is where the smoke tests' fixtures are kept, relative to this package.
const smokeFixtureDir = "testdata/smoke"

var updateFixtures = flag.Bool("update-fixtures", false, "rewrite the smoke fixtures in the form GCS saves them in")

// fixtureEnums lists the enum packages in model/gurps/enums that the fixtures must hold every value of. They are the
// ones TestSmokeFixtureCoverage requires.
var fixtureEnums = []string{
	"affects", "attribute", "container", "difficulty", "display", "emcost", "emweight", "eqcontainer", "equipmentsel",
	"feature", "frequency", "layoutnode", "maxusesmod", "namegen", "picker", "prereq", "selector", "selfctrl",
	"skillsel", "spellcmp", "spellmatch", "stdmg", "stlimit", "study", "threshold", "traitsel", "wsel", "wswitch",
}

// fixtureExemptEnums lists the enum packages in model/gurps/enums the fixtures need not hold every value of, and why.
var fixtureExemptEnums = map[string]string{
	"autoscale":   "how the PDF viewer scales pages, an application setting rather than data",
	"cell":        "how a list column is drawn, which nothing stores",
	"dgroup":      "which dockables share a window, kept in the application's own settings",
	"encumbrance": "worked out from what a character carries, never stored",
	"filternode":  "list filters, kept in the application's own settings",
	"layoutedge":  "where a dragged block lands in the layout editor, which nothing stores",
	"progression": "one sheet setting whose every value is stored the same way",
	"promptstep":  "the order of the prompts when rows arrive in a document, which nothing stores",
	"srcstate":    "worked out by comparing a row with its library source, never stored",
	"updatecheck": "how often to check for updates, an application setting rather than data",
}

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
	"output template legacy", "key bindings", "page references",
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
	// Loading general settings applies some of them to unison, as starting GCS does.
	preserveGeneralSettingsEffects(t)
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
	c.requireEveryEnumPackage()
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
	c.requireEnum("attribute type", attribute.Types)
	c.requireEnum("attribute placement", attribute.Placements)
	c.requireEnum("pool threshold op", without(threshold.Ops, threshold.Unknown))
	c.requireEnum("display option", display.Options)
	c.requireEnum("layout node", layoutnode.Types)
	c.requireEnum("name generator", namegen.Types)
}

// fixtureCoverage records what the fixtures were found to hold.
type fixtureCoverage struct {
	t    *testing.T
	seen map[string]map[string]bool
}

// preserveGeneralSettingsEffects puts back, when the test ends, the unison state that loading general settings changes
// (see gurps.GeneralSettings.EnsureValidity).
func preserveGeneralSettingsEffects(t *testing.T) {
	t.Helper()
	tooltipDelay := unison.DefaultTooltipTheme.Delay
	tooltipDismissal := unison.DefaultTooltipTheme.Dismissal
	cursorSize := unison.CursorSize()
	focusForReading := unison.FocusForReading()
	t.Cleanup(func() {
		unison.DefaultTooltipTheme.Delay = tooltipDelay
		unison.DefaultTooltipTheme.Dismissal = tooltipDismissal
		unison.SetCursorSize(cursorSize)
		unison.SetFocusForReading(focusForReading)
	})
}

// requireEveryEnumPackage checks that every enum package in model/gurps/enums is either required by this test or
// exempt from it, so that a new one is not overlooked.
func (c *fixtureCoverage) requireEveryEnumPackage() {
	c.t.Helper()
	entries, err := os.ReadDir(filepath.Join("..", "model", "gurps", "enums"))
	if err != nil {
		c.t.Fatal(err)
	}
	for _, one := range entries {
		if !one.IsDir() {
			continue
		}
		name := one.Name()
		_, exempt := fixtureExemptEnums[name]
		if required := slices.Contains(fixtureEnums, name); required == exempt {
			if required {
				c.t.Errorf("enum package %q is both required and exempt", name)
			} else {
				c.t.Errorf("enum package %q is neither required by the fixture coverage nor listed as exempt", name)
			}
		}
	}
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
	case []attribute.Type:
		requireKeys(c, category, list)
	case []attribute.Placement:
		requireKeys(c, category, list)
	case []threshold.Op:
		requireKeys(c, category, list)
	case []display.Option:
		requireKeys(c, category, list)
	case []layoutnode.Type:
		requireKeys(c, category, list)
	case []namegen.Type:
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
// records what it holds. Then it saves what it loaded through GCS and checks that the result matches the fixture.
func (c *fixtureCoverage) load(fsys fs.FS, p string) {
	c.t.Helper()
	ext := strings.ToLower(path.Ext(p))
	c.mark("file type", ext)
	save, err := c.loadAndWalk(fsys, p, ext)
	if err != nil {
		c.t.Errorf("%s: %v", p, err)
		return
	}
	if save != nil {
		c.roundTrip(fsys, p, save)
	}
}

// loadAndWalk loads one fixture and records what it holds. It returns the function that saves what was loaded the way
// GCS saves that file type, or nil for a type GCS doesn't save (or, for key bindings, saves only partly; see
// gurps.KeyBindings.MarshalJSONTo).
func (c *fixtureCoverage) loadAndWalk(fsys fs.FS, p, ext string) (func(string) error, error) {
	switch ext {
	case gurps.SheetExt:
		e, err := gurps.NewEntityFromFile(fsys, p)
		if err != nil {
			return nil, err
		}
		c.entity(e)
		return e.Save, nil
	case gurps.TemplatesExt:
		tmpl, err := gurps.NewTemplateFromFile(fsys, p)
		if err != nil {
			return nil, err
		}
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
		return tmpl.Save, nil
	case gurps.LootExt:
		l, err := gurps.NewLootFromFile(fsys, p)
		if err != nil {
			return nil, err
		}
		c.mark("case", "loot")
		c.equipment(l.Equipment)
		c.notes(l.Notes)
		return l.Save, nil
	case gurps.TraitsExt:
		list, err := gurps.NewTraitsFromFile(fsys, p)
		if err != nil {
			return nil, err
		}
		c.traits(list)
		return func(dst string) error { return gurps.SaveTraits(list, dst) }, nil
	case gurps.TraitModifiersExt:
		list, err := gurps.NewTraitModifiersFromFile(fsys, p)
		if err != nil {
			return nil, err
		}
		c.traitModifiers(list)
		return func(dst string) error { return gurps.SaveTraitModifiers(list, dst) }, nil
	case gurps.SkillsExt:
		list, err := gurps.NewSkillsFromFile(fsys, p)
		if err != nil {
			return nil, err
		}
		c.skills(list)
		return func(dst string) error { return gurps.SaveSkills(list, dst) }, nil
	case gurps.SpellsExt:
		list, err := gurps.NewSpellsFromFile(fsys, p)
		if err != nil {
			return nil, err
		}
		c.spells(list)
		return func(dst string) error { return gurps.SaveSpells(list, dst) }, nil
	case gurps.EquipmentExt:
		list, err := gurps.NewEquipmentFromFile(fsys, p)
		if err != nil {
			return nil, err
		}
		c.equipment(list)
		return func(dst string) error { return gurps.SaveEquipment(list, dst) }, nil
	case gurps.EquipmentModifiersExt:
		list, err := gurps.NewEquipmentModifiersFromFile(fsys, p)
		if err != nil {
			return nil, err
		}
		c.equipmentModifiers(list)
		return func(dst string) error { return gurps.SaveEquipmentModifiers(list, dst) }, nil
	case gurps.NotesExt:
		list, err := gurps.NewNotesFromFile(fsys, p)
		if err != nil {
			return nil, err
		}
		c.notes(list)
		return func(dst string) error { return gurps.SaveNotes(list, dst) }, nil
	case gurps.AncestryExt:
		a, err := gurps.NewAncestryFromFile(fsys, p)
		if err != nil {
			return nil, err
		}
		return a.Save, nil
	case gurps.NamesExt:
		// Loaded once ready to generate names, which checks the definition is complete, and once as an editor would,
		// which keeps it exactly as written for saving.
		n, err := gurps.NewNameGeneratorFromFS(fsys, p)
		if err != nil {
			return nil, err
		}
		c.markKey("name generator", n.Type)
		for _, one := range n.Compound {
			c.markKey("name generator", one.Type)
		}
		if n, err = gurps.ReadNameGeneratorFromFS(fsys, p); err != nil {
			return nil, err
		}
		return n.Save, nil
	case gurps.AttributesExt:
		defs, err := gurps.NewAttributeDefsFromFile(fsys, p)
		if err != nil {
			return nil, err
		}
		c.attributeDefs(defs)
		return defs.Save, nil
	case gurps.BodyExt:
		b, err := gurps.NewBodyFromFile(fsys, p)
		if err != nil {
			return nil, err
		}
		return b.Save, nil
	case gurps.CalendarExt:
		_, err := gurps.NewCalendarRefFromFS(fsys, p)
		return nil, err
	case gurps.ColorSettingsExt:
		cs, err := colors.NewFromFS(fsys, p)
		if err != nil {
			return nil, err
		}
		return cs.Save, nil
	case gurps.FontSettingsExt:
		f, err := fonts.NewFromFS(fsys, p)
		if err != nil {
			return nil, err
		}
		return f.Save, nil
	case gurps.GeneralSettingsExt:
		g, err := gurps.NewGeneralSettingsFromFile(fsys, p)
		if err != nil {
			return nil, err
		}
		return g.Save, nil
	case gurps.KeySettingsExt:
		if _, err := gurps.NewKeyBindingsFromFS(fsys, p); err != nil {
			return nil, err
		}
		var raw map[string]string
		if err := readJSON(fsys, p, &raw); err != nil {
			return nil, err
		}
		if len(raw) != 0 {
			c.mark("case", "key bindings")
		}
		return nil, nil
	case gurps.PageRefSettingsExt:
		refs, err := gurps.NewPageRefsFromFS(fsys, p)
		if err != nil {
			return nil, err
		}
		if len(refs.List()) != 0 {
			c.mark("case", "page references")
		}
		return refs.Save, nil
	case gurps.SheetSettingsExt:
		ss, err := gurps.NewSheetSettingsFromFile(fsys, p)
		if err != nil {
			return nil, err
		}
		c.sheetSettings(ss)
		return ss.Save, nil
	case ".md":
		return nil, nil
	case ".html", ".txt":
		return nil, c.outputTemplate(fsys, p)
	default:
		return nil, fmt.Errorf("no loader for %s files", ext)
	}
}

func readJSON(fsys fs.FS, p string, v any) error {
	data, err := fs.ReadFile(fsys, p)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// roundTrip saves what was loaded from the fixture at p through GCS and checks that the result is the fixture itself,
// apart from line endings, which a checkout may have changed. With -update-fixtures, it writes the result over the
// fixture instead.
func (c *fixtureCoverage) roundTrip(fsys fs.FS, p string, save func(string) error) {
	c.t.Helper()
	dst := filepath.Join(c.t.TempDir(), path.Base(p))
	if err := save(dst); err != nil {
		c.t.Errorf("%s: unable to save: %v", p, err)
		return
	}
	saved, err := os.ReadFile(dst)
	if err != nil {
		c.t.Fatal(err)
	}
	original, err := fs.ReadFile(fsys, p)
	if err != nil {
		c.t.Fatal(err)
	}
	if bytes.Equal(bytes.ReplaceAll(original, []byte("\r\n"), []byte("\n")), saved) {
		return
	}
	if *updateFixtures {
		// G703: p is one of the fixtures the walk found, and rewriting it is what -update-fixtures asks for.
		target := filepath.Join(smokeFixtureDir, filepath.FromSlash(p))
		if err = os.WriteFile(target, saved, 0o640); err != nil { //nolint:gosec // See above.
			c.t.Fatal(err)
		}
		return
	}
	c.t.Errorf("%s is not in the form GCS saves it in; run with -update-fixtures to rewrite it", p)
}

func (c *fixtureCoverage) outputTemplate(fsys fs.FS, p string) error {
	if path.Base(path.Dir(p)) != "Output Templates" {
		return fmt.Errorf("not an output template")
	}
	data, err := fs.ReadFile(fsys, p)
	if err != nil {
		return err
	}
	// The first line is read as gurps.Export reads it, which leaves out a carriage return.
	var first []byte
	if _, first, err = bufio.ScanLines(data, true); err != nil {
		return err
	}
	switch string(first) {
	case "GCS HTML Template v1":
		c.mark("case", "output template html")
	case "GCS Text Template v1":
		c.mark("case", "output template text")
	default:
		c.mark("case", "output template legacy")
	}
	return nil
}

func (c *fixtureCoverage) sheetSettings(ss *gurps.SheetSettings) {
	for _, one := range []display.Option{
		ss.UserDescriptionDisplay, ss.ModifiersDisplay, ss.NotesDisplay,
		ss.SkillLevelAdjDisplay,
	} {
		c.markKey("display option", one)
	}
	if ss.Attributes != nil {
		c.attributeDefs(ss.Attributes)
	}
	if ss.Layout != nil {
		c.layoutNode(ss.Layout.Root)
	}
}

func (c *fixtureCoverage) layoutNode(n *gurps.SheetLayoutNode) {
	if n == nil {
		return
	}
	c.markKey("layout node", n.Type)
	for _, one := range n.Children {
		c.layoutNode(one)
	}
}

func (c *fixtureCoverage) attributeDefs(defs *gurps.AttributeDefs) {
	for _, def := range defs.List(false) {
		c.markKey("attribute type", def.Type)
		c.markKey("attribute placement", def.Placement)
		for _, t := range def.Thresholds {
			for _, op := range t.Ops {
				c.markKey("pool threshold op", op)
			}
		}
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
	c.sheetSettings(e.SheetSettings)
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
			c.modifierContainer("modifier", one.Choice.IsZero(), one.IsMandatoryChoice())
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

func (c *fixtureCoverage) modifierContainer(kind string, group, mandatory bool) {
	switch {
	case group:
		c.mark("case", kind+" group")
	case mandatory:
		c.mark("case", kind+" choice mandatory")
	default:
		c.mark("case", kind+" choice optional")
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
			for _, s := range one.Study {
				c.markKey("study", s.Type)
			}
			c.features(one.Features)
			c.weapons(one.Weapons)
			c.prereqs(one.Prereq)
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
			c.features(one.Features)
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
			c.modifierContainer("equipment modifier", one.Choice.IsZero(), one.IsMandatoryChoice())
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

// prereqs records what a row's prerequisites hold. Every row has a top-level list, so only the lists nested inside it
// count as examples of prereq_list.
func (c *fixtureCoverage) prereqs(list *gurps.PrereqList) {
	if list == nil {
		return
	}
	if !list.All {
		c.mark("case", "prereq list any")
	}
	if list.WhenTL.Compare != criteria.AnyNumber {
		c.mark("case", "prereq when tl")
	}
	for _, one := range list.Prereqs {
		switch p := one.(type) {
		case *gurps.PrereqList:
			c.markKey("prereq", p.PrereqType())
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
