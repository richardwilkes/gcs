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
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"testing/fstest"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison/enums/align"
)

// TestNewTraitsFromFileAttachesContainerModifiers verifies that loading a trait list attaches the modifiers of container
// rows as well as those of leaf rows. The loader used to skip containers, so a container's modifiers never learned
// which trait they belonged to and could not resolve its nameable placeholders.
func TestNewTraitsFromFileAttachesContainerModifiers(t *testing.T) {
	c := check.New(t)
	container := NewTrait(nil, nil, true)
	container.Name = "Container"
	container.Replacements = map[string]string{"element": "Fire"}
	containerMod := NewTraitModifier(nil, nil, false)
	containerMod.Name = "@element@ Only"
	container.Modifiers = []*TraitModifier{containerMod}
	child := NewTrait(nil, container, false)
	child.Name = "Child"
	child.Replacements = map[string]string{"element": "Ice"}
	childMod := NewTraitModifier(nil, nil, false)
	childMod.Name = "@element@ Only"
	child.Modifiers = []*TraitModifier{childMod}
	container.Children = []*Trait{child}

	dir := t.TempDir()
	c.NoError(SaveTraits([]*Trait{container}, filepath.Join(dir, "Test"+TraitsExt)))
	rows, err := NewTraitsFromFile(os.DirFS(dir), "Test"+TraitsExt)
	c.NoError(err)
	c.Equal(1, len(rows), "the list has a single top-level row")
	if len(rows) != 1 || len(rows[0].Modifiers) != 1 || len(rows[0].Children) != 1 ||
		len(rows[0].Children[0].Modifiers) != 1 {
		c.Fatalf("the loaded rows do not have the expected shape: %#v", rows)
	}
	loadedContainer := rows[0]
	loadedChild := loadedContainer.Children[0]
	c.True(loadedContainer.Container(), "the top-level row is a container")
	c.True(loadedChild.Parent() == loadedContainer, "the child points back at its parent")
	c.True(loadedContainer.Modifiers[0].OwningTrait() == loadedContainer,
		"the container's modifier belongs to the container")
	c.True(loadedChild.Modifiers[0].OwningTrait() == loadedChild, "the child's modifier belongs to the child")
	c.Equal("Fire Only", loadedContainer.Modifiers[0].NameWithReplacements(),
		"the container's modifier resolves the container's replacements")
	c.Equal("Ice Only", loadedChild.Modifiers[0].NameWithReplacements(),
		"the child's modifier resolves the child's replacements")
}

// TestLoadRowsRejectsFutureVersion verifies that the shared loader still refuses a list written by a newer release.
func TestLoadRowsRejectsFutureVersion(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	c.NoError(jio.SaveToFile(filepath.Join(dir, "Test"+NotesExt),
		&listData[*Note]{Version: jio.CurrentDataVersion + 1}))
	_, err := NewNotesFromFile(os.DirFS(dir), "Test"+NotesExt)
	c.HasError(err, "a list from a newer release is rejected")
}

// TestLegacyNodeUnmarshal verifies the fix-ups applied to nodes written before TIDs and tags existed: the kind is taken
// from the old type string, the old open flag is honored, categories are folded into the sorted tags, the old notes
// are migrated, and the children point back at their parent.
func TestLegacyNodeUnmarshal(t *testing.T) {
	c := check.New(t)
	data := `{
	"version": ` + strconv.Itoa(jio.CurrentDataVersion) + `,
	"rows": [
		{
			"id": "not-a-tid",
			"type": "advantage_container",
			"open": true,
			"name": "Parent",
			"notes": "plain notes",
			"tags": ["Zeta"],
			"categories": ["Alpha/Beta", "zeta"],
			"children": [
				{
					"type": "advantage",
					"name": "Child",
					"mental": true,
					"categories": ["Gamma"]
				}
			]
		}
	]
}`
	traits, err := NewTraitsFromFile(fstest.MapFS{"Test.adq": {Data: []byte(data)}}, "Test.adq")
	c.NoError(err)
	c.Equal(1, len(traits))
	if len(traits) != 1 || len(traits[0].Children) != 1 {
		c.Fatalf("the loaded rows do not have the expected shape: %#v", traits)
	}
	parent := traits[0]
	child := parent.Children[0]
	c.True(tid.IsValid(parent.TID), "the parent was given a valid TID")
	c.True(tid.IsKind(parent.TID, kinds.TraitContainer), "the parent's kind comes from the old type string")
	c.True(parent.IsOpen(), "the old open flag was honored")
	c.Equal("plain notes", parent.LocalNotes, "the old notes were migrated")
	c.Equal([]string{"Alpha", "Beta", "Zeta"}, parent.Tags,
		"categories are folded into the tags, without duplicates, and sorted")
	c.True(tid.IsValid(child.TID), "the child was given a valid TID")
	c.True(tid.IsKind(child.TID, kinds.Trait), "the child's kind comes from the old type string")
	c.True(child.Parent() == parent, "the child points back at its parent")
	c.Equal([]string{"Gamma", "Mental"}, child.Tags, "the old type flags become tags")
}

// TestFixupLegacyTIDKeepsValidIDs verifies that a node that already has a TID is left alone, and that the open flag
// is therefore not carried over for it, since its open state is tracked by TID.
func TestFixupLegacyTIDKeepsValidIDs(t *testing.T) {
	c := check.New(t)
	id := tid.MustNewTID(kinds.Skill)
	original := id
	c.False(fixupLegacyTID(&id, "skill_container", skillKind), "a valid TID is not replaced")
	c.Equal(original, id)
	id = ""
	c.True(fixupLegacyTID(&id, "skill_container", skillKind), "a missing TID is replaced")
	c.True(tid.IsKind(id, kinds.SkillContainer), "the container suffix selects the container kind")
	id = "legacy-uuid"
	c.True(fixupLegacyTID(&id, "skill", skillKind), "a UUID is replaced")
	c.True(tid.IsKind(id, kinds.Skill), "no suffix selects the plain kind")
}

// TestSharedCellAndHeaderData verifies the cell and header helpers shared by every node type.
func TestSharedCellAndHeaderData(t *testing.T) {
	c := check.New(t)

	header := pageRefHeaderData()
	c.Equal(HeaderBookmark, header.Title)
	c.True(header.TitleIsImageKey)
	c.Equal(PageRefTooltip(), header.Detail)
	header = libSrcHeaderData()
	c.Equal(HeaderDatabase, header.Title)
	c.True(header.TitleIsImageKey)
	c.Equal(LibSrcTooltip(), header.Detail)
	header = switchHeaderData()
	c.Equal(HeaderSwitch, header.Title)
	c.True(header.TitleIsImageKey)
	c.Equal(SwitchHeaderTooltip(), header.Detail)
	header = enabledHeaderData()
	c.Equal(HeaderCheckmark, header.Title)
	c.True(header.TitleIsImageKey)
	c.Equal(ModifierEnabledTooltip(), header.Detail)
	header = tagsHeaderData()
	c.Equal("Tags", header.Title)
	c.False(header.TitleIsImageKey)

	var data CellData
	fillTagsCell(&data, []string{"b", "a"})
	c.Equal(cell.Tags, data.Type)
	c.Equal(CombineTags([]string{"b", "a"}), data.Primary)

	calls := 0
	fallback := func() string {
		calls++
		return "fallback"
	}
	data = CellData{}
	fillPageRefCell(&data, "B12", "highlight", fallback)
	c.Equal(cell.PageRef, data.Type)
	c.Equal("B12", data.Primary)
	c.Equal("highlight", data.Secondary, "the highlight wins when there is one")
	c.Equal(0, calls, "the fallback is not evaluated when the highlight is used")
	data = CellData{}
	fillPageRefCell(&data, "B12", "", fallback)
	c.Equal("fallback", data.Secondary, "the fallback is used when there is no highlight")
	c.Equal(1, calls)

	// Without an owner there is nobody to ask about the source, so only the type and alignment are filled in.
	trait := NewTrait(nil, nil, false)
	data = CellData{}
	fillLibSrcCell(&data, nil, trait)
	c.Equal(cell.Text, data.Type)
	c.Equal(align.Middle, data.Alignment)
	c.Equal("", data.Primary)
	c.Equal("", data.Tooltip)

	// With an owner, a trait that has no source is reported as custom, with no source details in the tooltip.
	entity := NewEntity()
	data = CellData{}
	fillLibSrcCell(&data, entity, trait)
	c.Equal(cell.Text, data.Type)
	c.NotEqual("", data.Primary, "the source state is shown")
	c.NotContains(data.Tooltip, "\n", "custom data has no source details to show")
}
