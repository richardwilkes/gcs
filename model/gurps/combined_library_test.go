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
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
)

// buildCombinedLibraryFixture writes two small "book" folders into a temporary library, covering every data type that
// participates in combining, and returns the library along with a snapshot of every file written, so that tests can
// verify the sources are never altered. Book A is the higher priority source in most tests; each type has an item the
// two books share and at least one the second book alone provides.
func buildCombinedLibraryFixture(t *testing.T) (lib *Library, snapshot map[string][]byte) {
	t.Helper()
	c := check.New(t)
	lib = NewLibrary("Test Library", "", "", "test_library", t.TempDir())
	root := lib.Path()

	newTrait := func(name, ref string, points int) *Trait {
		tr := NewTrait(nil, nil, false)
		tr.Name = name
		tr.PageRef = ref
		tr.BasePoints = fxp.FromInteger(points)
		return tr
	}
	c.NoError(SaveTraits([]*Trait{newTrait("Combat Reflexes", "A43", 15), newTrait("Luck", "A66", 15)},
		filepath.Join(root, "Book A", "Book A Traits.adq")))
	c.NoError(SaveTraits([]*Trait{newTrait("Combat Reflexes", "B12", 20), newTrait("Danger Sense", "B34", 15)},
		filepath.Join(root, "Book B", "Book B Traits.adq")))

	newTraitMod := func(name, ref, notes string) *TraitModifier {
		m := NewTraitModifier(nil, nil, false)
		m.Name = name
		m.PageRef = ref
		m.LocalNotes = notes
		return m
	}
	c.NoError(SaveTraitModifiers([]*TraitModifier{newTraitMod("Cosmic", "A101", "from A")},
		filepath.Join(root, "Book A", "Book A Trait Modifiers.adm")))
	c.NoError(SaveTraitModifiers([]*TraitModifier{
		newTraitMod("Cosmic", "B101", "from B"),
		newTraitMod("Nuisance Effect", "B102", "from B"),
	}, filepath.Join(root, "Book B", "Book B Trait Modifiers.adm")))

	newSkill := func(name, ref string) *Skill {
		s := NewSkill(nil, nil, false)
		s.Name = name
		s.PageRef = ref
		return s
	}
	c.NoError(SaveSkills([]*Skill{newSkill("Karate", "A10")},
		filepath.Join(root, "Book A", "Book A Skills.skl")))
	c.NoError(SaveSkills([]*Skill{newSkill("Karate", "B20"), newSkill("Judo", "B21")},
		filepath.Join(root, "Book B", "Book B Skills.skl")))

	newSpell := func(name, ref string, points int) *Spell {
		s := NewSpell(nil, nil, false)
		s.Name = name
		s.PageRef = ref
		s.Points = fxp.FromInteger(points)
		return s
	}
	c.NoError(SaveSpells([]*Spell{newSpell("Light", "A110", 1)},
		filepath.Join(root, "Book A", "Book A Spells.spl")))
	c.NoError(SaveSpells([]*Spell{newSpell("Light", "B110", 2), newSpell("Darkness", "B111", 4)},
		filepath.Join(root, "Book B", "Book B Spells.spl")))

	newEqp := func(name, tl, ref, value string) *Equipment {
		eqp := NewEquipment(nil, nil, false)
		eqp.Name = name
		eqp.TechLevel = tl
		eqp.PageRef = ref
		eqp.BaseValue = value
		return eqp
	}
	c.NoError(SaveEquipment([]*Equipment{
		newEqp("Broad Sword", "3", "A100", "500"),
		newEqp("Medical Supplies", "9", "A200", "10"),
	}, filepath.Join(root, "Book A", "Book A Equipment.eqp")))
	c.NoError(SaveEquipment([]*Equipment{
		newEqp("Broad Sword", "3", "B55", "750"),
		newEqp("Medical Supplies", "10", "B60", "20"),
		newEqp("Unique Thing", "8", "B70", "99"),
	}, filepath.Join(root, "Book B", "Book B Equipment.eqp")))

	newEqpMod := func(name, ref, notes string) *EquipmentModifier {
		m := NewEquipmentModifier(nil, nil, false)
		m.Name = name
		m.PageRef = ref
		m.LocalNotes = notes
		return m
	}
	c.NoError(SaveEquipmentModifiers([]*EquipmentModifier{newEqpMod("Fine", "A74", "from A")},
		filepath.Join(root, "Book A", "Book A Equipment Modifiers.eqm")))
	c.NoError(SaveEquipmentModifiers([]*EquipmentModifier{
		newEqpMod("Fine", "B74", "from B"),
		newEqpMod("Cheap", "B75", "from B"),
	}, filepath.Join(root, "Book B", "Book B Equipment Modifiers.eqm")))

	combatA := newFixtureNote(nil, "Combat", true)
	newFixtureNote(combatA, "Rule X", false)
	c.NoError(SaveNotes([]*Note{combatA}, filepath.Join(root, "Book A", "Book A Rules.not")))
	combatB := newFixtureNote(nil, "Combat", true)
	newFixtureNote(combatB, "Rule Y", false)
	c.NoError(SaveNotes([]*Note{combatB}, filepath.Join(root, "Book B", "Book B Rules.not")))

	snapshot = snapshotDirFiles(t, root)
	return lib, snapshot
}

func newFixtureNote(parent *Note, markdown string, container bool) *Note {
	n := NewNote(nil, parent, container)
	n.MarkDown = markdown
	if parent != nil {
		parent.Children = append(parent.Children, n)
	}
	return n
}

func snapshotDirFiles(t *testing.T, dir string) map[string][]byte {
	t.Helper()
	c := check.New(t)
	root, err := os.OpenRoot(dir)
	c.NoError(err)
	defer func() { c.NoError(root.Close()) }()
	snapshot := make(map[string][]byte)
	c.NoError(fs.WalkDir(root.FS(), ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			data, readErr := root.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			snapshot[path] = data
		}
		return nil
	}))
	return snapshot
}

// combineFixture runs a combination of the fixture's two books, in the given folder order, into a fresh directory.
func combineFixture(t *testing.T, lib *Library, destDir string, folders ...string) []string {
	t.Helper()
	c := check.New(t)
	sources := make([]CombinedLibrarySource, 0, len(folders))
	for _, folder := range folders {
		sources = append(sources, CombinedLibrarySource{Library: lib, Folder: folder})
	}
	created, err := CreateCombinedLibrary(CombinedLibraryOptions{Name: "Combo", Sources: sources}, destDir)
	c.NoError(err)
	return created
}

func TestCreateCombinedLibrary(t *testing.T) {
	c := check.New(t)
	lib, before := buildCombinedLibraryFixture(t)
	destDir := filepath.Join(t.TempDir(), "Combo")
	created := combineFixture(t, lib, destDir, "Book A", "Book B")

	// Every data type is present among the sources, so all seven files are produced, in the documented order.
	c.Equal([]string{
		filepath.Join(destDir, "Combo Traits.adq"),
		filepath.Join(destDir, "Combo Trait Modifiers.adm"),
		filepath.Join(destDir, "Combo Skills.skl"),
		filepath.Join(destDir, "Combo Spells.spl"),
		filepath.Join(destDir, "Combo Equipment.eqp"),
		filepath.Join(destDir, "Combo Equipment Modifiers.eqm"),
		filepath.Join(destDir, "Combo Rules.not"),
	}, created)

	// The sources must not have been altered in any way.
	c.Equal(before, snapshotDirFiles(t, lib.Path()))

	traits, err := NewTraitsFromFile(os.DirFS(destDir), "Combo Traits.adq")
	c.NoError(err)
	c.Equal(3, len(traits))
	c.Equal("Combat Reflexes", traits[0].Name)
	c.Equal(fxp.FromInteger(15), traits[0].BasePoints, "the higher priority book's stats win")
	c.Equal("A43,B12", traits[0].PageRef)
	c.Equal("Danger Sense", traits[2].Name, "an item only the lower priority book has is still carried over")

	traitMods, err := NewTraitModifiersFromFile(os.DirFS(destDir), "Combo Trait Modifiers.adm")
	c.NoError(err)
	c.Equal(2, len(traitMods))
	c.Equal("Cosmic", traitMods[0].Name)
	c.Equal("from A", traitMods[0].LocalNotes)
	c.Equal("A101,B101", traitMods[0].PageRef)
	c.Equal("Nuisance Effect", traitMods[1].Name)

	skills, err := NewSkillsFromFile(os.DirFS(destDir), "Combo Skills.skl")
	c.NoError(err)
	c.Equal(2, len(skills))
	c.Equal("Karate", skills[0].Name)
	c.Equal("A10,B20", skills[0].PageRef)
	c.Equal("Judo", skills[1].Name)

	spells, err := NewSpellsFromFile(os.DirFS(destDir), "Combo Spells.spl")
	c.NoError(err)
	c.Equal(2, len(spells))
	c.Equal("Light", spells[0].Name)
	c.Equal(fxp.FromInteger(1), spells[0].Points)
	c.Equal("A110,B110", spells[0].PageRef)
	c.Equal("Darkness", spells[1].Name)

	eqp, err := NewEquipmentFromFile(os.DirFS(destDir), "Combo Equipment.eqp")
	c.NoError(err)
	c.Equal(4, len(eqp))
	byName := make(map[string]*Equipment)
	for _, one := range eqp {
		byName[one.Name+"/TL"+one.TechLevel] = one
	}
	// A collision keeps the stats of the highest priority book but gathers both books' page references.
	broadSword := byName["Broad Sword/TL3"]
	c.NotNil(broadSword)
	c.Equal("500", broadSword.BaseValue)
	c.Equal("A100,B55", broadSword.PageRef)
	// The name alone is not the identity: differing tech levels remain distinct items.
	c.NotNil(byName["Medical Supplies/TL9"])
	c.NotNil(byName["Medical Supplies/TL10"])
	c.NotNil(byName["Unique Thing/TL8"])

	eqpMods, err := NewEquipmentModifiersFromFile(os.DirFS(destDir), "Combo Equipment Modifiers.eqm")
	c.NoError(err)
	c.Equal(2, len(eqpMods))
	c.Equal("Fine", eqpMods[0].Name)
	c.Equal("from A", eqpMods[0].LocalNotes)
	c.Equal("A74,B74", eqpMods[0].PageRef)
	c.Equal("Cheap", eqpMods[1].Name)

	notes, err := NewNotesFromFile(os.DirFS(destDir), "Combo Rules.not")
	c.NoError(err)
	c.Equal(1, len(notes))
	c.True(notes[0].Container())
	c.Equal("Combat", notes[0].MarkDown)
	c.Equal(2, len(notes[0].Children))
	c.Equal("Rule X", notes[0].Children[0].MarkDown)
	c.Equal("Rule Y", notes[0].Children[1].MarkDown)
	c.Equal(notes[0], notes[0].Children[1].Parent(), "merged children must be reparented to the surviving container")
}

// TestCombinedRowsAreRootData pins down the source-of-truth decision: combined rows carry no Source of their own, so a
// row dragged from a combined file onto a sheet anchors to the combined file rather than to the book it came from. Were
// the rows to reference the books instead, "Sync with Source" would restore the book's stats and throw away the merged
// page references that are the whole point of combining.
func TestCombinedRowsAreRootData(t *testing.T) {
	c := check.New(t)
	lib, _ := buildCombinedLibraryFixture(t)
	destDir := filepath.Join(t.TempDir(), "Combo")
	combineFixture(t, lib, destDir, "Book A", "Book B")

	traits, err := NewTraitsFromFile(os.DirFS(destDir), "Combo Traits.adq")
	c.NoError(err)
	for _, one := range traits {
		c.True(one.Source.IsZero(), "trait %q must not reference a book", one.Name)
	}
	eqp, err := NewEquipmentFromFile(os.DirFS(destDir), "Combo Equipment.eqp")
	c.NoError(err)
	for _, one := range eqp {
		c.True(one.Source.IsZero(), "equipment %q must not reference a book", one.Name)
	}
	// Nested rows must be cleared too, not just the top level.
	notes, err := NewNotesFromFile(os.DirFS(destDir), "Combo Rules.not")
	c.NoError(err)
	c.Equal(1, len(notes))
	c.True(notes[0].Source.IsZero(), "the container must not reference a book")
	for _, child := range notes[0].Children {
		c.True(child.Source.IsZero(), "child note %q must not reference a book", child.MarkDown)
	}
}

// TestCreateCombinedLibraryReusesIDs verifies that regenerating a combined file keeps the IDs of the rows that are
// still present, so that sheets referring to the combined file survive a re-run.
func TestCreateCombinedLibraryReusesIDs(t *testing.T) {
	c := check.New(t)
	lib, _ := buildCombinedLibraryFixture(t)
	destDir := filepath.Join(t.TempDir(), "Combo")

	combineFixture(t, lib, destDir, "Book A", "Book B")
	firstTraits, err := NewTraitsFromFile(os.DirFS(destDir), "Combo Traits.adq")
	c.NoError(err)
	firstNotes, err := NewNotesFromFile(os.DirFS(destDir), "Combo Rules.not")
	c.NoError(err)

	combineFixture(t, lib, destDir, "Book A", "Book B")
	secondTraits, err := NewTraitsFromFile(os.DirFS(destDir), "Combo Traits.adq")
	c.NoError(err)
	secondNotes, err := NewNotesFromFile(os.DirFS(destDir), "Combo Rules.not")
	c.NoError(err)

	c.Equal(len(firstTraits), len(secondTraits))
	for i, one := range firstTraits {
		c.Equal(one.ID(), secondTraits[i].ID(), "trait %q must keep its ID across a regeneration", one.Name)
	}
	// Nested rows keep their IDs too.
	c.Equal(1, len(secondNotes))
	c.Equal(firstNotes[0].ID(), secondNotes[0].ID(), "the container must keep its ID")
	c.Equal(len(firstNotes[0].Children), len(secondNotes[0].Children))
	for i, child := range firstNotes[0].Children {
		c.Equal(child.ID(), secondNotes[0].Children[i].ID(), "child note %q must keep its ID", child.MarkDown)
	}

	// A row that appears only after the first run gets an ID of its own rather than stealing another row's.
	c.NoError(SaveTraits([]*Trait{func() *Trait {
		tr := NewTrait(nil, nil, false)
		tr.Name = "Ambidexterity"
		tr.PageRef = "A39"
		return tr
	}()}, filepath.Join(lib.Path(), "Book A", "Book A Extra Traits.adq")))
	combineFixture(t, lib, destDir, "Book A", "Book B")
	thirdTraits, err := NewTraitsFromFile(os.DirFS(destDir), "Combo Traits.adq")
	c.NoError(err)
	c.Equal(len(firstTraits)+1, len(thirdTraits))
	seen := make(map[tid.TID]bool, len(thirdTraits))
	for _, one := range thirdTraits {
		c.False(seen[one.ID()], "trait %q must not share an ID with another row", one.Name)
		seen[one.ID()] = true
	}
	byName := make(map[string]tid.TID, len(thirdTraits))
	for _, one := range thirdTraits {
		byName[one.Name] = one.ID()
	}
	for _, one := range firstTraits {
		c.Equal(one.ID(), byName[one.Name], "trait %q must still keep its ID", one.Name)
	}
}

// TestCreateCombinedLibraryReusesIDsForDuplicateSiblings covers the case real books actually hit: a container with
// two identically named children (Ultra Tech has several, such as two "Damage Resistance" rows under one implant).
// Pairing same-identity siblings up in document order is what keeps the second and later ones from being handed a
// fresh ID on every regeneration.
func TestCreateCombinedLibraryReusesIDsForDuplicateSiblings(t *testing.T) {
	c := check.New(t)
	lib := NewLibrary("Test Library", "", "", "test_library", t.TempDir())
	root := lib.Path()
	implant := newFixtureNote(nil, "Implant", true)
	newFixtureNote(implant, "Damage Resistance", false)
	newFixtureNote(implant, "Damage Resistance", false)
	newFixtureNote(implant, "Damage Resistance", false)
	c.NoError(SaveNotes([]*Note{implant}, filepath.Join(root, "Book A", "Book A Rules.not")))

	destDir := filepath.Join(t.TempDir(), "Combo")
	combineFixture(t, lib, destDir, "Book A")
	first, err := NewNotesFromFile(os.DirFS(destDir), "Combo Rules.not")
	c.NoError(err)
	c.Equal(1, len(first))
	c.Equal(3, len(first[0].Children), "all three same-named children survive")

	combineFixture(t, lib, destDir, "Book A")
	second, err := NewNotesFromFile(os.DirFS(destDir), "Combo Rules.not")
	c.NoError(err)
	c.Equal(1, len(second))
	c.Equal(first[0].ID(), second[0].ID())
	c.Equal(len(first[0].Children), len(second[0].Children))
	seen := make(map[tid.TID]bool, len(second[0].Children))
	for i, child := range first[0].Children {
		c.Equal(child.ID(), second[0].Children[i].ID(),
			"same-identity sibling %d must keep its own ID, not just the first of them", i)
		c.False(seen[second[0].Children[i].ID()], "sibling %d must not share an ID with another row", i)
		seen[second[0].Children[i].ID()] = true
	}
}

// TestCreateCombinedLibraryKeepsDuplicatesWithinASource checks that merging happens between sources, not inside one. A
// book that lists the same item twice is entitled to do so, and combining must not silently drop one of them.
func TestCreateCombinedLibraryKeepsDuplicatesWithinASource(t *testing.T) {
	c := check.New(t)
	lib := NewLibrary("Test Library", "", "", "test_library", t.TempDir())
	root := lib.Path()
	newSkill := func(name, ref string) *Skill {
		s := NewSkill(nil, nil, false)
		s.Name = name
		s.PageRef = ref
		return s
	}
	// Twice within a single file, and again in a second file of the same source folder that belongs to the same
	// table ("Extra Skills" ends with "Skills"). Files within a source are read in natural filename order, so the
	// "Extra" file is read first.
	c.NoError(SaveSkills([]*Skill{newSkill("Karate", "A10"), newSkill("Karate", "A11")},
		filepath.Join(root, "Book A", "Book A Skills.skl")))
	c.NoError(SaveSkills([]*Skill{newSkill("Karate", "A12")},
		filepath.Join(root, "Book A", "Book A Extra Skills.skl")))
	c.NoError(SaveSkills([]*Skill{newSkill("Karate", "B20")},
		filepath.Join(root, "Book B", "Book B Skills.skl")))

	destDir := filepath.Join(t.TempDir(), "Combo")
	combineFixture(t, lib, destDir, "Book A", "Book B")
	skills, err := NewSkillsFromFile(os.DirFS(destDir), "Combo Skills.skl")
	c.NoError(err)
	c.Equal(3, len(skills), "all three of Book A's entries survive")
	// Book B's matching row merges into the first of them, and only that one.
	c.Equal("A12,B20", skills[0].PageRef)
	c.Equal("A10", skills[1].PageRef)
	c.Equal("A11", skills[2].PageRef)
}

// TestCreateCombinedLibrarySubfolders exercises the "include subfolders" path: subfolder files are only reached with
// it on, and the subfolder structure is mirrored in the output rather than flattened into it.
func TestCreateCombinedLibrarySubfolders(t *testing.T) {
	c := check.New(t)
	lib := NewLibrary("Test Library", "", "", "test_library", t.TempDir())
	root := lib.Path()
	newSkill := func(name, ref string) *Skill {
		s := NewSkill(nil, nil, false)
		s.Name = name
		s.PageRef = ref
		return s
	}
	c.NoError(SaveSkills([]*Skill{newSkill("Karate", "A10")},
		filepath.Join(root, "Book A", "Book A Skills.skl")))
	c.NoError(SaveSkills([]*Skill{newSkill("Judo", "A30"), newSkill("Karate", "A31")},
		filepath.Join(root, "Book A", "Nested", "Nested Skills.skl")))
	c.NoError(SaveSkills([]*Skill{newSkill("Karate", "B10")},
		filepath.Join(root, "Book B", "Book B Skills.skl")))
	c.NoError(SaveSkills([]*Skill{newSkill("Judo", "B30")},
		filepath.Join(root, "Book B", "Nested", "Nested Skills.skl")))
	// A hidden folder is skipped whether or not subfolders are included.
	c.NoError(SaveSkills([]*Skill{newSkill("Sumo", "A90")},
		filepath.Join(root, "Book A", ".hidden", "Hidden Skills.skl")))

	withoutSubfolders := filepath.Join(t.TempDir(), "Flat")
	_, err := CreateCombinedLibrary(CombinedLibraryOptions{
		Name:    "Combo",
		Sources: []CombinedLibrarySource{{Library: lib, Folder: "Book A"}},
	}, withoutSubfolders)
	c.NoError(err)
	flat, err := NewSkillsFromFile(os.DirFS(withoutSubfolders), "Combo Skills.skl")
	c.NoError(err)
	c.Equal(1, len(flat), "only the folder's own files are read")
	c.Equal("Karate", flat[0].Name)

	withSubfolders := filepath.Join(t.TempDir(), "Deep")
	created, err := CreateCombinedLibrary(CombinedLibraryOptions{
		Name: "Combo",
		Sources: []CombinedLibrarySource{
			{Library: lib, Folder: "Book A"},
			{Library: lib, Folder: "Book B"},
		},
		IncludeSubfolders: true,
	}, withSubfolders)
	c.NoError(err)
	c.Equal([]string{
		filepath.Join(withSubfolders, "Combo Skills.skl"),
		filepath.Join(withSubfolders, "Nested", "Combo Skills.skl"),
	}, created)

	top, err := NewSkillsFromFile(os.DirFS(withSubfolders), "Combo Skills.skl")
	c.NoError(err)
	c.Equal(1, len(top), "the root files merge among themselves")
	c.Equal("Karate", top[0].Name)
	c.Equal("A10,B10", top[0].PageRef)

	nested, err := NewSkillsFromFile(os.DirFS(filepath.Join(withSubfolders, "Nested")), "Combo Skills.skl")
	c.NoError(err)
	c.Equal(2, len(nested), "the subfolder keeps its own combined file")
	c.Equal("Judo", nested[0].Name)
	c.Equal("A30,B30", nested[0].PageRef, "matching subfolders merge across sources")
	c.Equal("Karate", nested[1].Name)
	c.Equal("A31", nested[1].PageRef, "the root Karate does not bleed into the subfolder")
}

// TestCreateCombinedLibraryGroupsByTableName pins down the name-based grouping: specialized lists such as armor
// design tables keep combined files of their own instead of being folded into the general equipment list, while a
// file from a companion volume whose name ends with a standard table name joins that table despite its prefix.
func TestCreateCombinedLibraryGroupsByTableName(t *testing.T) {
	c := check.New(t)
	lib := NewLibrary("Test Library", "", "", "test_library", t.TempDir())
	root := lib.Path()
	newEqp := func(name, tl, ref string) *Equipment {
		eqp := NewEquipment(nil, nil, false)
		eqp.Name = name
		eqp.TechLevel = tl
		eqp.PageRef = ref
		return eqp
	}
	newEqpMod := func(name, ref string) *EquipmentModifier {
		m := NewEquipmentModifier(nil, nil, false)
		m.Name = name
		m.PageRef = ref
		return m
	}
	c.NoError(SaveEquipment([]*Equipment{newEqp("Broad Sword", "3", "B100")},
		filepath.Join(root, "Basic Set", "Basic Set Equipment.eqp")))
	c.NoError(SaveEquipmentModifiers([]*EquipmentModifier{newEqpMod("Fine", "B90")},
		filepath.Join(root, "Basic Set", "Basic Set Equipment Modifiers.eqm")))
	c.NoError(SaveEquipment([]*Equipment{newEqp("Broad Sword", "3", "LT50")},
		filepath.Join(root, "Low Tech", "Low Tech Equipment.eqp")))
	c.NoError(SaveEquipment([]*Equipment{newEqp("Plate Harness", "4", "LT110")},
		filepath.Join(root, "Low Tech", "Low Tech Armor (by location).eqp")))
	c.NoError(SaveEquipmentModifiers([]*EquipmentModifier{newEqpMod("Fine", "LTC20")},
		filepath.Join(root, "Low Tech", "Low Tech Companion 2 Equipment Modifiers.eqm")))

	destDir := filepath.Join(t.TempDir(), "Combo")
	created, err := CreateCombinedLibrary(CombinedLibraryOptions{
		Name: "Combo",
		Sources: []CombinedLibrarySource{
			{Library: lib, Folder: "Basic Set"},
			{Library: lib, Folder: "Low Tech"},
		},
	}, destDir)
	c.NoError(err)
	c.Equal([]string{
		filepath.Join(destDir, "Combo Armor (by location).eqp"),
		filepath.Join(destDir, "Combo Equipment.eqp"),
		filepath.Join(destDir, "Combo Equipment Modifiers.eqm"),
	}, created)

	// The armor design table stays its own file rather than polluting the general equipment list.
	armor, err := NewEquipmentFromFile(os.DirFS(destDir), "Combo Armor (by location).eqp")
	c.NoError(err)
	c.Equal(1, len(armor))
	c.Equal("Plate Harness", armor[0].Name)

	eqp, err := NewEquipmentFromFile(os.DirFS(destDir), "Combo Equipment.eqp")
	c.NoError(err)
	c.Equal(1, len(eqp), "the armor table's rows must not appear here")
	c.Equal("B100,LT50", eqp[0].PageRef)

	// The companion volume's modifiers merge into Equipment Modifiers despite the extra words in its file name.
	mods, err := NewEquipmentModifiersFromFile(os.DirFS(destDir), "Combo Equipment Modifiers.eqm")
	c.NoError(err)
	c.Equal(1, len(mods))
	c.Equal("Fine", mods[0].Name)
	c.Equal("B90,LTC20", mods[0].PageRef)
}

// TestCombinedTableName exercises the table-name heuristic against the naming conventions found in real libraries.
func TestCombinedTableName(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		base     string
		ext      string
		prefixes []string
		want     string
	}{
		// Standard tables, with any book prefix.
		{"Basic Set Equipment", EquipmentExt, []string{"Basic Set"}, "Equipment"},
		{"Low Tech Companion 2 Equipment Modifiers", EquipmentModifiersExt, []string{"Low Tech"}, "Equipment Modifiers"},
		{"Basic Set TL Cost Modifiers", EquipmentModifiersExt, []string{"Basic Set"}, "TL Cost Modifiers"},
		{"Power Ups 4 Enhancements Enhancement Modifiers", TraitModifiersExt, []string{"Power Ups"}, "Enhancement Modifiers"},
		{"Elvish Skills", SkillsExt, []string{"Fantasy Folk", "Elves", "Traits and Skills"}, "Skills"},
		{"Discworld Roleplaying Game Notes", NotesExt, []string{"Discworld Roleplaying Game"}, "Notes"},
		// Specialized lists keep their own tables, named without the folder prefix.
		{"Low Tech Armor (by location)", EquipmentExt, []string{"Low Tech"}, "Armor (by location)"},
		{"Low Tech Instant Armor", EquipmentExt, []string{"Low Tech"}, "Instant Armor"},
		{"Martial Arts Technical Grappling Techniques", SkillsExt, []string{"Martial Arts"}, "Technical Grappling Techniques"},
		// The space boundary keeps "Meta-Traits" from matching "Traits", and the longest prefix wins so a
		// subfolder's name beats its parent's.
		{"Ultra Tech Meta-Traits", TraitsExt, []string{"Ultra Tech"}, "Meta-Traits"},
		{"Dungeon Fantasy 1 Equipment Enchantments", EquipmentModifiersExt, []string{"Dungeon Fantasy", "Dungeon Fantasy 1"}, "Equipment Enchantments"},
		// A separator after the prefix belongs to the prefix, not the table.
		{
			"Dungeon Fantasy Adventure 1 - Mirror of the Fire Demons Contents", NotesExt,
			[]string{"Dungeon Fantasy", "Dungeon Fantasy Adventure 1"},
			"Mirror of the Fire Demons Contents",
		},
		// No standard suffix and no matching prefix: the whole name is the table.
		{"Conditions", TraitsExt, []string{"Basic Set"}, "Conditions"},
		{"CDRU TL9 Missile Weapons", EquipmentExt, []string{"Home Brew"}, "CDRU TL9 Missile Weapons"},
		// A file named exactly for a standard table, or exactly for its folder, still resolves sanely.
		{"Equipment", EquipmentExt, []string{"Book"}, "Equipment"},
		{"Book", EquipmentExt, []string{"Book"}, "Book"},
	} {
		c.Equal(one.want, combinedTableName(one.base, one.ext, one.prefixes), "%s (%s)", one.base, one.ext)
	}
}

// TestCreateCombinedLibraryLeavesNoPartialOutput checks that a failure part way through writes nothing, rather than
// leaving a half-populated folder behind.
func TestCreateCombinedLibraryLeavesNoPartialOutput(t *testing.T) {
	c := check.New(t)
	lib, _ := buildCombinedLibraryFixture(t)
	// Traits are combined first and Rules last, so corrupting the notes file fails only after the traits have been
	// merged, which is exactly the case that used to leave output behind.
	c.NoError(os.WriteFile(filepath.Join(lib.Path(), "Book A", "Book A Rules.not"),
		[]byte("this is not valid data"), 0o600))

	destDir := filepath.Join(t.TempDir(), "Combo")
	_, err := CreateCombinedLibrary(CombinedLibraryOptions{
		Name: "Combo",
		Sources: []CombinedLibrarySource{
			{Library: lib, Folder: "Book A"},
			{Library: lib, Folder: "Book B"},
		},
	}, destDir)
	c.HasError(err)
	_, statErr := os.Stat(destDir)
	c.True(os.IsNotExist(statErr), "a failed run must not create the destination folder")
}

// TestCreateCombinedLibraryKeepsExistingOutputOnFailure checks that a failed re-run leaves a previously generated
// combined folder untouched rather than partially overwritten.
func TestCreateCombinedLibraryKeepsExistingOutputOnFailure(t *testing.T) {
	c := check.New(t)
	lib, _ := buildCombinedLibraryFixture(t)
	destDir := filepath.Join(t.TempDir(), "Combo")
	combineFixture(t, lib, destDir, "Book A", "Book B")
	good := snapshotDirFiles(t, destDir)

	c.NoError(os.WriteFile(filepath.Join(lib.Path(), "Book A", "Book A Rules.not"),
		[]byte("this is not valid data"), 0o600))
	_, err := CreateCombinedLibrary(CombinedLibraryOptions{
		Name: "Combo",
		Sources: []CombinedLibrarySource{
			{Library: lib, Folder: "Book A"},
			{Library: lib, Folder: "Book B"},
		},
	}, destDir)
	c.HasError(err)
	c.Equal(good, snapshotDirFiles(t, destDir), "the previous output must survive a failed re-run intact")
}

func TestCreateCombinedLibraryPriorityOrder(t *testing.T) {
	c := check.New(t)
	lib, _ := buildCombinedLibraryFixture(t)
	destDir := filepath.Join(t.TempDir(), "Combo")
	combineFixture(t, lib, destDir, "Book B", "Book A")

	eqp, err := NewEquipmentFromFile(os.DirFS(destDir), "Combo Equipment.eqp")
	c.NoError(err)
	var found bool
	for _, one := range eqp {
		if one.Name == "Broad Sword" {
			// With Book B as the highest priority, its stats win and its reference comes first.
			c.Equal("750", one.BaseValue)
			c.Equal("B55,A100", one.PageRef)
			found = true
		}
	}
	c.True(found, "Broad Sword not found")

	traits, err := NewTraitsFromFile(os.DirFS(destDir), "Combo Traits.adq")
	c.NoError(err)
	c.Equal("Combat Reflexes", traits[0].Name)
	c.Equal(fxp.FromInteger(20), traits[0].BasePoints, "Book B's stats win when it is the higher priority")
	c.Equal("B12,A43", traits[0].PageRef)
}

func TestCreateCombinedLibraryValidation(t *testing.T) {
	c := check.New(t)
	lib, _ := buildCombinedLibraryFixture(t)
	destDir := filepath.Join(t.TempDir(), "Combo")
	_, err := CreateCombinedLibrary(CombinedLibraryOptions{
		Name:    "  ",
		Sources: []CombinedLibrarySource{{Library: lib, Folder: "Book A"}},
	}, destDir)
	c.HasError(err)
	_, err = CreateCombinedLibrary(CombinedLibraryOptions{Name: "Combo"}, destDir)
	c.HasError(err)
	_, err = CreateCombinedLibrary(CombinedLibraryOptions{
		Name:    "Combo",
		Sources: []CombinedLibrarySource{{Library: lib, Folder: "No Such Folder"}},
	}, destDir)
	c.HasError(err)
}

func TestCombinePageRefs(t *testing.T) {
	c := check.New(t)
	c.Equal("A1", combinePageRefs("A1", ""))
	c.Equal("B2", combinePageRefs("", "B2"))
	c.Equal("A1,B2", combinePageRefs("A1", "B2"))
	c.Equal("A1", combinePageRefs("A1", "a1"), "duplicates are dropped case-insensitively")
	c.Equal("A1,B2", combinePageRefs("A1, B2", "B2"))
	c.Equal("A1,B2,C3", combinePageRefs("A1,B2", "B2, C3"))
}

// planFixtureSources returns the fixture's two books as sources, highest priority first.
func planFixtureSources(lib *Library) []CombinedLibrarySource {
	return []CombinedLibrarySource{
		{Library: lib, Folder: "Book A"},
		{Library: lib, Folder: "Book B"},
	}
}

// TestCombinedPlanDefaultsMatchCreateCombinedLibrary verifies the promise the staging GUI is built on: adding source
// folders and creating without touching anything produces exactly what CreateCombinedLibrary produces.
func TestCombinedPlanDefaultsMatchCreateCombinedLibrary(t *testing.T) {
	c := check.New(t)
	lib, _ := buildCombinedLibraryFixture(t)

	viaOpts := filepath.Join(t.TempDir(), "Combo")
	createdOpts, err := CreateCombinedLibrary(CombinedLibraryOptions{
		Name:    "Combo",
		Sources: planFixtureSources(lib),
	}, viaOpts)
	c.NoError(err)

	viaPlan := filepath.Join(t.TempDir(), "Combo")
	var plan CombinedPlan
	for _, src := range planFixtureSources(lib) {
		c.NoError(plan.AddSource(src, false))
	}
	c.Equal(plan.Targets("Combo", viaPlan), func() []string {
		targets := make([]string, 0, len(createdOpts))
		for _, one := range createdOpts {
			rel, relErr := filepath.Rel(viaOpts, one)
			c.NoError(relErr)
			targets = append(targets, filepath.Join(viaPlan, rel))
		}
		return targets
	}(), "Targets must predict exactly what Create writes")
	createdPlan, err := plan.Create("Combo", viaPlan)
	c.NoError(err)
	c.Equal(len(createdOpts), len(createdPlan))

	// The file sets must correspond name for name. Contents differ only in freshly minted row IDs, so compare row
	// names per file rather than bytes.
	for i, one := range createdOpts {
		c.Equal(filepath.Base(one), filepath.Base(createdPlan[i]))
	}
	skillsA, err := NewSkillsFromFile(os.DirFS(viaOpts), "Combo Skills.skl")
	c.NoError(err)
	skillsB, err := NewSkillsFromFile(os.DirFS(viaPlan), "Combo Skills.skl")
	c.NoError(err)
	c.Equal(len(skillsA), len(skillsB))
	for i := range skillsA {
		c.Equal(skillsA[i].Name, skillsB[i].Name)
		c.Equal(skillsA[i].PageRef, skillsB[i].PageRef)
	}
}

// TestCombinedPlanOmitComponent checks that removing a staged component keeps its rows out of the output.
func TestCombinedPlanOmitComponent(t *testing.T) {
	c := check.New(t)
	lib, _ := buildCombinedLibraryFixture(t)
	var plan CombinedPlan
	for _, src := range planFixtureSources(lib) {
		c.NoError(plan.AddSource(src, false))
	}
	for _, f := range plan.Files {
		if f.Ext == EquipmentExt {
			f.Components = slices.DeleteFunc(f.Components, func(comp CombinedComponent) bool {
				return strings.HasSuffix(comp.Path, "Book B Equipment.eqp")
			})
		}
	}
	destDir := filepath.Join(t.TempDir(), "Combo")
	_, err := plan.Create("Combo", destDir)
	c.NoError(err)
	eqp, err := NewEquipmentFromFile(os.DirFS(destDir), "Combo Equipment.eqp")
	c.NoError(err)
	c.Equal(2, len(eqp), "only Book A's items remain")
	for _, one := range eqp {
		c.NotEqual("Unique Thing", one.Name)
		c.False(strings.Contains(one.PageRef, "B"), "no Book B page references without its component: %s", one.PageRef)
	}
}

// TestCombinedPlanNewFileAndRename checks user-created files, custom names, and that default names track the plan
// name while custom names pin it.
func TestCombinedPlanNewFileAndRename(t *testing.T) {
	c := check.New(t)
	lib, _ := buildCombinedLibraryFixture(t)
	var plan CombinedPlan
	for _, src := range planFixtureSources(lib) {
		c.NoError(plan.AddSource(src, false))
	}
	// Move Book B's equipment into a user-created file, as one would to keep a specialized list separate.
	var eqpFile *CombinedFile
	for _, f := range plan.Files {
		if f.Ext == EquipmentExt {
			eqpFile = f
		}
	}
	c.NotNil(eqpFile)
	special := &CombinedFile{Subpath: "", CustomName: "My Game Special", Ext: EquipmentExt}
	plan.Files = append(plan.Files, special)
	moved := slices.DeleteFunc(slices.Clone(eqpFile.Components), func(comp CombinedComponent) bool {
		return !strings.HasSuffix(comp.Path, "Book B Equipment.eqp")
	})
	eqpFile.Components = slices.DeleteFunc(eqpFile.Components, func(comp CombinedComponent) bool {
		return strings.HasSuffix(comp.Path, "Book B Equipment.eqp")
	})
	special.Components = append(special.Components, moved...)

	// Rename the rules file; the rest keep tracking the plan name.
	for _, f := range plan.Files {
		if f.Ext == NotesExt {
			f.CustomName = "House Rules"
		}
	}

	destDir := filepath.Join(t.TempDir(), "Combo")
	created, err := plan.Create("My Game", destDir)
	c.NoError(err)
	c.Equal([]string{
		filepath.Join(destDir, "My Game Traits.adq"),
		filepath.Join(destDir, "My Game Trait Modifiers.adm"),
		filepath.Join(destDir, "My Game Skills.skl"),
		filepath.Join(destDir, "My Game Spells.spl"),
		filepath.Join(destDir, "My Game Equipment.eqp"),
		filepath.Join(destDir, "My Game Special.eqp"),
		filepath.Join(destDir, "My Game Equipment Modifiers.eqm"),
		filepath.Join(destDir, "House Rules.not"),
	}, created)

	eqp, err := NewEquipmentFromFile(os.DirFS(destDir), "My Game Equipment.eqp")
	c.NoError(err)
	c.Equal(2, len(eqp), "Book A's equipment only")
	specialRows, err := NewEquipmentFromFile(os.DirFS(destDir), "My Game Special.eqp")
	c.NoError(err)
	c.Equal(3, len(specialRows), "Book B's equipment moved wholesale")
}

// TestCombinedPlanReorderChangesPriority checks that component order is priority: after moving Book B's skills above
// Book A's, Book B's stats win and its references come first.
func TestCombinedPlanReorderChangesPriority(t *testing.T) {
	c := check.New(t)
	lib, _ := buildCombinedLibraryFixture(t)
	var plan CombinedPlan
	for _, src := range planFixtureSources(lib) {
		c.NoError(plan.AddSource(src, false))
	}
	for _, f := range plan.Files {
		if f.Ext == SkillsExt {
			c.Equal(2, len(f.Components))
			f.Components[0], f.Components[1] = f.Components[1], f.Components[0]
		}
	}
	destDir := filepath.Join(t.TempDir(), "Combo")
	_, err := plan.Create("Combo", destDir)
	c.NoError(err)
	skills, err := NewSkillsFromFile(os.DirFS(destDir), "Combo Skills.skl")
	c.NoError(err)
	c.Equal(2, len(skills))
	c.Equal("Karate", skills[0].Name)
	c.Equal("B20,A10", skills[0].PageRef, "Book B's reference leads once it has priority")
}

// TestCombinedPlanInterleavingSplitsABook checks the merge unit: a book's consecutive components keep their
// deliberate duplicates, but a component from another book placed between them splits the run, and the later part
// merges like any lower-priority contribution.
func TestCombinedPlanInterleavingSplitsABook(t *testing.T) {
	c := check.New(t)
	lib := NewLibrary("Test Library", "", "", "test_library", t.TempDir())
	root := lib.Path()
	newSkill := func(name, ref string) *Skill {
		s := NewSkill(nil, nil, false)
		s.Name = name
		s.PageRef = ref
		return s
	}
	c.NoError(SaveSkills([]*Skill{newSkill("Karate", "A10"), newSkill("Karate", "A11")},
		filepath.Join(root, "Book A", "Book A Skills.skl")))
	c.NoError(SaveSkills([]*Skill{newSkill("Karate", "A12")},
		filepath.Join(root, "Book A", "Extra Skills.skl")))
	c.NoError(SaveSkills([]*Skill{newSkill("Karate", "B20")},
		filepath.Join(root, "Book B", "Book B Skills.skl")))
	var plan CombinedPlan
	c.NoError(plan.AddSource(CombinedLibrarySource{Library: lib, Folder: "Book A"}, false))
	c.NoError(plan.AddSource(CombinedLibrarySource{Library: lib, Folder: "Book B"}, false))
	for _, f := range plan.Files {
		if f.Ext == SkillsExt {
			c.Equal(3, len(f.Components))
			// Default order: Book A Skills, Extra Skills (both Book A, natural order), then Book B. Move Book
			// B's component between Book A's two.
			f.Components[1], f.Components[2] = f.Components[2], f.Components[1]
		}
	}
	destDir := filepath.Join(t.TempDir(), "Combo")
	_, err := plan.Create("Combo", destDir)
	c.NoError(err)
	skills, err := NewSkillsFromFile(os.DirFS(destDir), "Combo Skills.skl")
	c.NoError(err)
	c.Equal(2, len(skills), "the split run merges its Karate into the leader; the same-file duplicate survives")
	c.Equal("A10,B20,A12", skills[0].PageRef)
	c.Equal("A11", skills[1].PageRef)
}

// TestCombinedPlanDuplicateTargets checks that two staged files resolving to the same output path are refused rather
// than silently overwriting one another.
func TestCombinedPlanDuplicateTargets(t *testing.T) {
	c := check.New(t)
	lib, _ := buildCombinedLibraryFixture(t)
	var plan CombinedPlan
	for _, src := range planFixtureSources(lib) {
		c.NoError(plan.AddSource(src, false))
	}
	for _, f := range plan.Files {
		if f.Ext == SkillsExt {
			f.CustomName = "Combo Equipment Modifiers" // Same name as another staged file, differing only in case...
		}
	}
	destDir := filepath.Join(t.TempDir(), "Combo")
	_, err := plan.Create("Combo", destDir)
	c.NoError(err, "same name with a different extension is fine")

	for _, f := range plan.Files {
		if f.Ext == SkillsExt {
			f.CustomName = ""
		}
		if f.Ext == TraitsExt {
			f.CustomName = "combo skills"
			f.Ext = SkillsExt // Force the collision; not reachable through the UI, which fixes extensions.
		}
	}
	_, err = plan.Create("Combo", destDir)
	c.HasError(err)
}

// TestCombinedPlanAddFile checks single-file staging: default placement matches folder staging, duplicates are
// refused, and non-combinable files are ignored.
func TestCombinedPlanAddFile(t *testing.T) {
	c := check.New(t)
	lib, _ := buildCombinedLibraryFixture(t)
	src := CombinedLibrarySource{Library: lib, Folder: "Book A"}
	var plan CombinedPlan
	c.True(plan.AddFile(src, "Book A Skills.skl"))
	c.False(plan.AddFile(src, "Book A Skills.skl"), "already staged")
	c.False(plan.AddFile(src, "No Such.txt"), "not a combinable type")
	c.Equal(1, len(plan.Files))
	c.Equal("Skills", plan.Files[0].Table)
	c.Equal("", plan.Files[0].Subpath)

	// Adding the folder afterwards fills in the rest without duplicating the staged file.
	c.NoError(plan.AddSource(src, false))
	for _, f := range plan.Files {
		if f.Ext == SkillsExt {
			c.Equal(1, len(f.Components))
		}
	}
}

// TestCombinedPlanEmptyAndInvalid checks Create's validation.
func TestCombinedPlanEmptyAndInvalid(t *testing.T) {
	c := check.New(t)
	var plan CombinedPlan
	destDir := filepath.Join(t.TempDir(), "Combo")
	_, err := plan.Create("Combo", destDir)
	c.HasError(err, "nothing staged")
	lib, _ := buildCombinedLibraryFixture(t)
	c.NoError(plan.AddSource(CombinedLibrarySource{Library: lib, Folder: "Book A"}, false))
	_, err = plan.Create("  ", destDir)
	c.HasError(err, "no name")
	// An empty user-created file is skipped rather than written.
	plan.Files = append(plan.Files, &CombinedFile{CustomName: "Empty", Ext: SkillsExt})
	created, err := plan.Create("Combo", destDir)
	c.NoError(err)
	for _, one := range created {
		c.NotEqual("Empty.skl", filepath.Base(one))
	}
}
