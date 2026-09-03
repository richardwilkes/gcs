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
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// buildCombinedLibraryFixture writes two small "book" folders into a temporary library and returns the library along
// with a snapshot of every file written, so that tests can verify the sources are never altered.
func buildCombinedLibraryFixture(t *testing.T) (lib *Library, snapshot map[string][]byte) {
	t.Helper()
	c := check.New(t)
	lib = NewLibrary("Test Library", "", "", "test_library", t.TempDir())
	root := lib.Path()

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

	newNote := func(parent *Note, markdown string, container bool) *Note {
		n := NewNote(nil, parent, container)
		n.MarkDown = markdown
		if parent != nil {
			parent.Children = append(parent.Children, n)
		}
		return n
	}
	combatA := newNote(nil, "Combat", true)
	newNote(combatA, "Rule X", false)
	c.NoError(SaveNotes([]*Note{combatA}, filepath.Join(root, "Book A", "Book A Rules.not")))
	combatB := newNote(nil, "Combat", true)
	newNote(combatB, "Rule Y", false)
	c.NoError(SaveNotes([]*Note{combatB}, filepath.Join(root, "Book B", "Book B Rules.not")))

	snapshot = snapshotDirFiles(t, root)
	return lib, snapshot
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

func TestCreateCombinedLibrary(t *testing.T) {
	c := check.New(t)
	lib, before := buildCombinedLibraryFixture(t)
	destDir := filepath.Join(t.TempDir(), "Combo")
	created, err := CreateCombinedLibrary(CombinedLibraryOptions{
		Name: "Combo",
		Sources: []CombinedLibrarySource{
			{Library: lib, Folder: "Book A"},
			{Library: lib, Folder: "Book B"},
		},
	}, destDir)
	c.NoError(err)

	// Only the data types present among the sources produce files.
	c.Equal(3, len(created))
	c.True(filepath.Join(destDir, "Combo Skills.skl") == created[0], "skills file should be first: %v", created)
	c.True(filepath.Join(destDir, "Combo Equipment.eqp") == created[1], "equipment file should be second: %v", created)
	c.True(filepath.Join(destDir, "Combo Rules.not") == created[2], "rules file should be last: %v", created)

	// The sources must not have been altered in any way.
	c.Equal(before, snapshotDirFiles(t, lib.Path()))

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
	// Each combined row points back at the file it came from, so source syncing keeps working.
	c.Equal(lib.Key(), broadSword.Source.Library)
	c.Equal("Book A/Book A Equipment.eqp", broadSword.Source.Path)
	c.Equal("Book B/Book B Equipment.eqp", byName["Unique Thing/TL8"].Source.Path)

	skills, err := NewSkillsFromFile(os.DirFS(destDir), "Combo Skills.skl")
	c.NoError(err)
	c.Equal(2, len(skills))
	c.Equal("Karate", skills[0].Name)
	c.Equal("A10,B20", skills[0].PageRef)
	c.Equal("Judo", skills[1].Name)

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

func TestCreateCombinedLibraryPriorityOrder(t *testing.T) {
	c := check.New(t)
	lib, _ := buildCombinedLibraryFixture(t)
	destDir := filepath.Join(t.TempDir(), "Combo")
	_, err := CreateCombinedLibrary(CombinedLibraryOptions{
		Name: "Combo",
		Sources: []CombinedLibrarySource{
			{Library: lib, Folder: "Book B"},
			{Library: lib, Folder: "Book A"},
		},
	}, destDir)
	c.NoError(err)
	eqp, err := NewEquipmentFromFile(os.DirFS(destDir), "Combo Equipment.eqp")
	c.NoError(err)
	for _, one := range eqp {
		if one.Name == "Broad Sword" {
			// With Book B as the highest priority, its stats win and its reference comes first.
			c.Equal("750", one.BaseValue)
			c.Equal("B55,A100", one.PageRef)
			c.Equal("Book B/Book B Equipment.eqp", one.Source.Path)
			return
		}
	}
	t.Fatal("Broad Sword not found")
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
