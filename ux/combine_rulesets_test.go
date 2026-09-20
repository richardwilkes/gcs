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
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestValidateCombinedRulesetName checks that names are accepted or rejected rather than quietly rewritten, and that
// the Windows rules are applied on every platform so a combined library stays portable.
func TestValidateCombinedRulesetName(t *testing.T) {
	c := check.New(t)

	for _, name := range []string{
		"Combo",
		"My Campaign",
		"Basic Set + Martial Arts",
		"D&D-ish",
		"Combo.v2",
		"COM10",   // Only COM1-COM9 are reserved
		"console", // Merely starts with a reserved name
		"  Combo  ",
	} {
		c.Equal("", validateCombinedRulesetName(name), "%q should be accepted", name)
	}

	// An empty name reports no problem: the buttons gated on the name are already disabled for it.
	c.Equal("", validateCombinedRulesetName(""))
	c.Equal("", validateCombinedRulesetName("   "))

	for _, name := range []string{
		`Basic Set: Characters`, // The colon used to be silently rewritten to "@6"
		`Book/Chapter`,          // As was the slash
		`Book\Chapter`,          // And the backslash
		`Say "What"`,
		"Who?",
		"Star*",
		"A<B",
		"A>B",
		"A|B",
		"Combo\t2",  // Control characters
		"Combo\x00", // ...including NUL
		".",
		"..",
		"Combo.", // A trailing period is dropped by Windows
		// Reserved device names, with or without an extension, in any case, and once trimmed.
		"con",
		"CON",
		"nul.txt",
		"LPT9",
		"  aux  ",
	} {
		c.NotEqual("", validateCombinedRulesetName(name), "%q should be rejected", name)
	}
}

// TestValidateCombinedRulesetNameLeavesTheNameAlone documents the behavior change: a name that is accepted is used
// exactly as typed, rather than being run through a sanitizer that substitutes characters behind the user's back.
func TestValidateCombinedRulesetNameLeavesTheNameAlone(t *testing.T) {
	c := check.New(t)
	// "@" was previously rewritten to "@3" by xfilepath.SanitizeName even though it is a perfectly legal filename
	// character on every platform GCS supports.
	c.Equal("", validateCombinedRulesetName("Email@Home"))
}

// buildCombineTreeFixture creates a small library on disk with a mix of combinable data files, other files, nested
// folders, and things that must be pruned or skipped.
func buildCombineTreeFixture(t *testing.T) *gurps.Library {
	t.Helper()
	c := check.New(t)
	lib := gurps.NewLibrary("Test Library", "", "", "test_library", t.TempDir())
	root := lib.Path()
	newSkill := func(name string) *gurps.Skill {
		s := gurps.NewSkill(nil, nil, false)
		s.Name = name
		return s
	}
	c.NoError(gurps.SaveSkills([]*gurps.Skill{newSkill("Karate")},
		filepath.Join(root, "Book A", "Book A Skills.skl")))
	c.NoError(gurps.SaveSkills([]*gurps.Skill{newSkill("Judo")},
		filepath.Join(root, "Book A", "Nested", "Nested Skills.skl")))
	c.NoError(gurps.SaveSkills([]*gurps.Skill{newSkill("Boxing")},
		filepath.Join(root, "Book B", "Book B Skills.skl")))
	// A root-level data file has no source folder and must not be offered.
	c.NoError(gurps.SaveSkills([]*gurps.Skill{newSkill("Sumo")}, filepath.Join(root, "Root Skills.skl")))
	// Non-combinable files and empty folders must be pruned.
	c.NoError(os.WriteFile(filepath.Join(root, "Book A", "readme.txt"), []byte("hi"), 0o600))
	c.NoError(os.MkdirAll(filepath.Join(root, "Empty Folder"), 0o750))
	// Hidden folders are skipped entirely.
	c.NoError(gurps.SaveSkills([]*gurps.Skill{newSkill("Hidden")},
		filepath.Join(root, ".hidden", "Hidden Skills.skl")))
	return lib
}

// treeTitles flattens a subtree into indented titles for compact assertions.
func treeTitles(nodes []*combineNode, indent string) []string {
	out := make([]string, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, indent+n.title)
		out = append(out, treeTitles(n.children, indent+"  ")...)
	}
	return out
}

// TestBuildCombineSourceTree checks the left tree's shape: folders and combinable files only, pruned of everything
// with nothing to offer, with root-level data files excluded.
func TestBuildCombineSourceTree(t *testing.T) {
	c := check.New(t)
	lib := buildCombineTreeFixture(t)
	openMap := make(map[string]bool)

	root := newCombineNode(combineRoleLibrary, "L:"+lib.Key(), "Test Library", openMap, nil)
	root.lib = lib
	buildCombineSourceFolder(root, lib, "", "")
	c.Equal([]string{
		"Book A",
		"  Book A Skills.skl",
		"  Nested",
		"    Nested Skills.skl",
		"Book B",
		"  Book B Skills.skl",
	}, treeTitles(root.children, ""))

	// A filter narrows the tree to matches and their ancestors...
	filtered := newCombineNode(combineRoleLibrary, "L:"+lib.Key(), "Test Library", openMap, nil)
	filtered.lib = lib
	buildCombineSourceFolder(filtered, lib, "", "nested")
	c.Equal([]string{
		"Book A",
		"  Nested",
		"    Nested Skills.skl",
	}, treeTitles(filtered.children, ""))

	// ...while a filter matching a folder keeps everything beneath it.
	byFolder := newCombineNode(combineRoleLibrary, "L:"+lib.Key(), "Test Library", openMap, nil)
	byFolder.lib = lib
	buildCombineSourceFolder(byFolder, lib, "", "book a")
	c.Equal([]string{
		"Book A",
		"  Book A Skills.skl",
		"  Nested",
		"    Nested Skills.skl",
	}, treeTitles(byFolder.children, ""))
}

// TestTopFolderOf checks the split of a library-relative path into its source folder and remainder.
func TestTopFolderOf(t *testing.T) {
	c := check.New(t)
	folder, rest := topFolderOf("Book A/Book A Skills.skl")
	c.Equal("Book A", folder)
	c.Equal("Book A Skills.skl", rest)
	folder, rest = topFolderOf("Book A/Nested/Deep Skills.skl")
	c.Equal("Book A", folder)
	c.Equal("Nested/Deep Skills.skl", rest)
	folder, rest = topFolderOf("Book A")
	c.Equal("Book A", folder)
	c.Equal("", rest)
}

// buildRightTree renders a plan into staging nodes the way refreshRight does, without needing a table.
func buildRightTree(plan *gurps.CombinedPlan, name string) []*combineNode {
	openMap := make(map[string]bool)
	folders := make(map[string]*combineNode)
	var roots []*combineNode
	var folderFor func(subpath string) *combineNode
	folderFor = func(subpath string) *combineNode {
		if subpath == "" {
			return nil
		}
		if n, ok := folders[subpath]; ok {
			return n
		}
		var parent *combineNode
		if i := strings.LastIndex(subpath, "/"); i != -1 {
			parent = folderFor(subpath[:i])
		}
		n := newCombineNode(combineRoleOutFolder, "D:"+strings.ToLower(subpath), path.Base(subpath), openMap, parent)
		if parent == nil {
			roots = append(roots, n)
		}
		folders[subpath] = n
		return n
	}
	for _, f := range plan.Files {
		fileName := f.FileName(name)
		parent := folderFor(f.Subpath)
		fileNode := newCombineNode(combineRoleOutFile, "F:"+strings.ToLower(f.Subpath+"/"+fileName), fileName,
			openMap, parent)
		fileNode.outFile = f
		if parent == nil {
			roots = append(roots, fileNode)
		}
		for i, comp := range f.Components {
			compNode := newCombineNode(combineRoleComponent, componentKey(comp), path.Base(comp.Path), openMap,
				fileNode)
			compNode.outFile = f
			compNode.compIndex = i
			compNode.comp = comp
		}
	}
	return roots
}

// findTreeNode returns the first node in the forest whose title matches.
func findTreeNode(nodes []*combineNode, title string) *combineNode {
	for _, n := range nodes {
		if n.title == title {
			return n
		}
		if found := findTreeNode(n.children, title); found != nil {
			return found
		}
	}
	return nil
}

// stageDnDFixture builds a plan holding the fixture's two books.
func stageDnDFixture(t *testing.T) (*gurps.Library, *gurps.CombinedPlan) {
	t.Helper()
	c := check.New(t)
	lib := buildCombineTreeFixture(t)
	plan := &gurps.CombinedPlan{}
	c.NoError(plan.AddSource(gurps.CombinedLibrarySource{Library: lib, Folder: "Book A"}, false))
	c.NoError(plan.AddSource(gurps.CombinedLibrarySource{Library: lib, Folder: "Book B"}, false))
	return lib, plan
}

// TestRebuildCombinePlanMovesComponents simulates a drop that moved one component into another file: the rebuilt
// plan must show it there, in the dropped position.
func TestRebuildCombinePlanMovesComponents(t *testing.T) {
	c := check.New(t)
	_, plan := stageDnDFixture(t)
	special := &gurps.CombinedFile{CustomName: "Special", Ext: gurps.SkillsExt}
	plan.Files = append(plan.Files, special)
	roots := buildRightTree(plan, "Combo")

	// Drag "Book B Skills.skl" from the default skills file into Special, as unison's drop machinery would: the
	// node is re-parented and the trees' child lists updated.
	comp := findTreeNode(roots, "Book B Skills.skl")
	c.NotNil(comp)
	origin := comp.parent
	origin.children = slices.DeleteFunc(slices.Clone(origin.children), func(n *combineNode) bool { return n == comp })
	target := findTreeNode(roots, "Special.skl")
	c.NotNil(target)
	comp.parent = target
	target.children = append(target.children, comp)

	rebuildCombinePlan(plan, roots, false)
	c.Equal(1, len(special.Components))
	c.Equal("Book B/Book B Skills.skl", special.Components[0].Path)
	for _, f := range plan.Files {
		if f.Table == "Skills" {
			c.Equal(1, len(f.Components), "the origin file keeps only Book A's component")
			c.Equal("Book A/Book A Skills.skl", f.Components[0].Path)
		}
	}
}

// TestRebuildCombinePlanReorders simulates a drop that reordered components within a file: the new order is the new
// priority.
func TestRebuildCombinePlanReorders(t *testing.T) {
	c := check.New(t)
	_, plan := stageDnDFixture(t)
	roots := buildRightTree(plan, "Combo")
	skills := findTreeNode(roots, "Combo Skills.skl")
	c.NotNil(skills)
	c.Equal(2, len(skills.children))
	skills.children[0], skills.children[1] = skills.children[1], skills.children[0]

	rebuildCombinePlan(plan, roots, false)
	for _, f := range plan.Files {
		if f.Table == "Skills" {
			c.Equal("Book B/Book B Skills.skl", f.Components[0].Path, "the dropped order is the priority order")
			c.Equal("Book A/Book A Skills.skl", f.Components[1].Path)
		}
	}
}

// TestRebuildCombinePlanSnapsBackIllegalDrops checks that a component dropped outside any file, or into a file of
// another type, returns to the file it came from.
func TestRebuildCombinePlanSnapsBackIllegalDrops(t *testing.T) {
	c := check.New(t)
	_, plan := stageDnDFixture(t)
	wrongType := &gurps.CombinedFile{CustomName: "Wrong", Ext: gurps.EquipmentExt}
	plan.Files = append(plan.Files, wrongType)
	roots := buildRightTree(plan, "Combo")

	comp := findTreeNode(roots, "Book B Skills.skl")
	origin := comp.parent
	origin.children = slices.DeleteFunc(slices.Clone(origin.children), func(n *combineNode) bool { return n == comp })
	target := findTreeNode(roots, "Wrong.eqp")
	comp.parent = target
	target.children = append(target.children, comp)

	rebuildCombinePlan(plan, roots, false)
	c.Equal(0, len(wrongType.Components), "a skills component cannot live in an equipment file")
	found := false
	for _, f := range plan.Files {
		for _, one := range f.Components {
			if one.Path == "Book B/Book B Skills.skl" {
				c.Equal("Skills", f.Table, "the component snapped back to the file it came from")
				found = true
			}
		}
	}
	c.True(found, "the component must not be lost")
}

// TestRebuildCombinePlanImportsSourceDrops checks drags from the source tree: a source file dropped inside a matching
// file joins it exactly there, while a source folder dropped anywhere stages at its defaults, and duplicates are not
// created.
func TestRebuildCombinePlanImportsSourceDrops(t *testing.T) {
	c := check.New(t)
	lib := buildCombineTreeFixture(t)
	plan := &gurps.CombinedPlan{}
	c.NoError(plan.AddSource(gurps.CombinedLibrarySource{Library: lib, Folder: "Book A"}, false))
	roots := buildRightTree(plan, "Combo")

	// A cloned source-file node dropped into the matching skills file, ahead of the existing component.
	skills := findTreeNode(roots, "Combo Skills.skl")
	c.NotNil(skills)
	dropped := newCombineNode(combineRoleSourceFile, "L:x", "Book B Skills.skl", make(map[string]bool), nil)
	dropped.lib = lib
	dropped.relPath = "Book B/Book B Skills.skl"
	dropped.parent = skills
	skills.children = append([]*combineNode{dropped}, skills.children...)
	rebuildCombinePlan(plan, roots, false)
	for _, f := range plan.Files {
		if f.Table == "Skills" {
			c.Equal(2, len(f.Components))
			c.Equal("Book B/Book B Skills.skl", f.Components[0].Path, "dropped at the head of the file")
		}
	}

	// Dropping the same file again changes nothing.
	roots = buildRightTree(plan, "Combo")
	again := newCombineNode(combineRoleSourceFile, "L:y", "Book B Skills.skl", make(map[string]bool), nil)
	again.lib = lib
	again.relPath = "Book B/Book B Skills.skl"
	roots = append(roots, again)
	rebuildCombinePlan(plan, roots, false)
	for _, f := range plan.Files {
		if f.Table == "Skills" {
			c.Equal(2, len(f.Components), "no duplicate staging from a second drop")
		}
	}

	// A source folder dropped loose stages at its defaults.
	roots = buildRightTree(plan, "Combo")
	folderNode := newCombineNode(combineRoleSourceFolder, "L:z", "Book A", make(map[string]bool), nil)
	folderNode.lib = lib
	folderNode.relPath = "Book A"
	roots = append(roots, folderNode)
	rebuildCombinePlan(plan, roots, true)
	c.True(plan.Contains(gurps.CombinedComponent{Library: lib, Path: "Book A/Nested/Nested Skills.skl"}),
		"the folder's subfolder files staged at their defaults")
}

// TestRebuildCombinePlanRehomesFiles checks that a file sitting under a different folder after a drop adopts that
// folder's position, and that renaming a subpath rehomes everything beneath it.
func TestRebuildCombinePlanRehomesFiles(t *testing.T) {
	c := check.New(t)
	lib := buildCombineTreeFixture(t)
	plan := &gurps.CombinedPlan{}
	c.NoError(plan.AddSource(gurps.CombinedLibrarySource{Library: lib, Folder: "Book A"}, true))
	roots := buildRightTree(plan, "Combo")

	// Simulate dragging the nested combined file up to the top level.
	nested := findTreeNode(roots, "Nested")
	c.NotNil(nested)
	c.Equal(1, len(nested.children))
	moved := nested.children[0]
	nested.children = nil
	moved.parent = nil
	roots = append(roots, moved)
	rebuildCombinePlan(plan, roots, true)
	for _, f := range plan.Files {
		c.Equal("", f.Subpath, "every staged file now sits at the top level")
	}

	// renameCombineSubpath rehomes by prefix.
	plan2 := &gurps.CombinedPlan{}
	c.NoError(plan2.AddSource(gurps.CombinedLibrarySource{Library: lib, Folder: "Book A"}, true))
	renameCombineSubpath(plan2, "Nested", "Deeper")
	found := false
	for _, f := range plan2.Files {
		c.NotEqual("Nested", f.Subpath)
		if f.Subpath == "Deeper" {
			found = true
		}
	}
	c.True(found)
}
