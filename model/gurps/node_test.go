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
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison/enums/align"
)

// TestNewTraitsFromFileAttachesContainerModifiers verifies that loading a trait list attaches the modifiers of
// container rows as well as those of leaf rows. The loader used to skip containers, so a container's modifiers never
// learned which trait they belonged to and could not resolve its nameable placeholders.
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
	c.True(loadedContainer.Modifiers[0].Target() == loadedContainer,
		"the container's modifier belongs to the container")
	c.True(loadedChild.Modifiers[0].Target() == loadedChild, "the child's modifier belongs to the child")
	c.Equal("Fire Only", loadedContainer.Modifiers[0].NameWithReplacements(),
		"the container's modifier resolves the container's replacements")
	c.Equal("Ice Only", loadedChild.Modifiers[0].NameWithReplacements(),
		"the child's modifier resolves the child's replacements")
}

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

// TestFixupLegacyTIDKeepsValidIDs verifies that a node that already has a TID of a kind its type can have is left
// alone, and that the open flag is therefore not carried over for it, since its open state is tracked by TID.
func TestFixupLegacyTIDKeepsValidIDs(t *testing.T) {
	c := check.New(t)
	own := []byte{kinds.Skill, kinds.SkillContainer, kinds.Technique}
	id := tid.MustNewTID(kinds.Skill)
	original := id
	c.False(fixupLegacyTID(&id, "skill_container", skillKind, own...), "a valid TID is not replaced")
	c.Equal(original, id)
	id = ""
	c.True(fixupLegacyTID(&id, "skill_container", skillKind, own...), "a missing TID is replaced")
	c.True(tid.IsKind(id, kinds.SkillContainer), "the container suffix selects the container kind")
	id = "legacy-uuid"
	c.True(fixupLegacyTID(&id, "skill", skillKind, own...), "a UUID is replaced")
	c.True(tid.IsKind(id, kinds.Skill), "no suffix selects the plain kind")

	// A hand-edited file can give a node an ID of a kind its type can't have, which is replaced as a missing one is,
	// while an ID of any kind the type can have is kept.
	types := []struct {
		load func(c check.Checker, id tid.TID) byte
		own  []byte // The kinds the type can have, starting with the one an ID of another kind is replaced with.
	}{
		{load: kindOnLoad[Trait], own: []byte{kinds.Trait, kinds.TraitContainer}},
		{load: kindOnLoad[TraitModifier], own: []byte{kinds.TraitModifier, kinds.TraitModifierContainer}},
		{load: kindOnLoad[Skill], own: []byte{kinds.Skill, kinds.SkillContainer, kinds.Technique}},
		{load: kindOnLoad[Spell], own: []byte{kinds.Spell, kinds.SpellContainer, kinds.RitualMagicSpell}},
		{load: kindOnLoad[Equipment], own: []byte{kinds.Equipment, kinds.EquipmentContainer}},
		{load: kindOnLoad[EquipmentModifier], own: []byte{kinds.EquipmentModifier, kinds.EquipmentModifierContainer}},
		{load: kindOnLoad[Note], own: []byte{kinds.Note, kinds.NoteContainer}},
	}
	for _, one := range types {
		for _, other := range types {
			for _, kind := range other.own {
				want := one.own[0]
				if slices.Contains(one.own, kind) {
					want = kind
				}
				c.Equal(string(want), string(one.load(c, tid.MustNewTID(kind))),
					"a node of a type with kinds %q loaded with an ID of kind %c", one.own, kind)
			}
		}
	}
}

// kindOnLoad returns the kind of the ID a T has once loaded from data that gives it only the ID.
func kindOnLoad[T any, P interface {
	*T
	ID() tid.TID
}](c check.Checker, id tid.TID) byte {
	c.Helper()
	var node T
	c.NoError(json.Unmarshal([]byte(`{"id":"`+string(id)+`"}`), P(&node)))
	return tidKind(P(&node).ID())
}

func TestSharedCellAndHeaderData(t *testing.T) {
	c := check.New(t)

	header := pageRefHeaderData()
	c.Equal(HeaderBookmark, header.Title)
	c.True(header.TitleIsImageKey)
	c.Equal("Page Reference", header.Name)
	c.Equal(PageRefTooltip(), header.Detail)
	header = libSrcHeaderData()
	c.Equal(HeaderDatabase, header.Title)
	c.True(header.TitleIsImageKey)
	c.Equal("Library Source", header.Name)
	c.Equal(LibSrcTooltip(), header.Detail)
	header = switchHeaderData()
	c.Equal(HeaderSwitch, header.Title)
	c.True(header.TitleIsImageKey)
	c.Equal("Switched On", header.Name)
	c.Equal(SwitchHeaderTooltip(), header.Detail)
	header = enabledHeaderData()
	c.Equal(HeaderCheckmark, header.Title)
	c.True(header.TitleIsImageKey)
	c.Equal("Enabled", header.Name)
	c.Equal(ModifierEnabledTooltip(), header.Detail)
	header = tagsHeaderData()
	c.Equal("Tags", header.Title)
	c.False(header.TitleIsImageKey)
	c.Equal("", header.Name, "a title that is a word is read as it is")
	header = abbreviatedHeaderData("Pts", "Points")
	c.Equal("Pts", header.Title)
	c.False(header.TitleIsImageKey)
	c.Equal("Points", header.Name)
	c.Equal("Points", header.Detail, "what an abbreviation stands for is its tooltip as well")

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

func TestClonePtr(t *testing.T) {
	c := check.New(t)
	c.Nil(clonePtr[string](nil), "nil clones to nil")
	tl := "3"
	clone := clonePtr(&tl)
	c.True(&tl != clone, "the clone is a distinct object")
	c.Equal("3", *clone, "the clone holds the same value")
	*clone = "4"
	c.Equal("3", tl, "editing the clone leaves the source alone")
}

// TestHeaderDataNamesWhatIsNotReadable walks every header data function, so that a new icon column, or one titled with
// an abbreviation of up to three characters, is caught if it lacks a name.
func TestHeaderDataNamesWhatIsNotReadable(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	functions := map[string]func(columnID int) HeaderData{
		"conditional modifiers": ConditionalModifiersHeaderData,
		"reaction modifiers":    ReactionModifiersHeaderData,
		"equipment modifiers":   EquipmentModifierHeaderData,
		"notes":                 NotesHeaderData,
		"skills":                SkillsHeaderData,
		"trait modifiers":       TraitModifierHeaderData,
		"spells":                SpellsHeaderData,
		"traits":                TraitsHeaderData,
		"melee weapons":         func(columnID int) HeaderData { return WeaponHeaderData(columnID, true, false) },
		"ranged weapons":        func(columnID int) HeaderData { return WeaponHeaderData(columnID, false, false) },
		"page melee weapons":    func(columnID int) HeaderData { return WeaponHeaderData(columnID, true, true) },
		"carried equipment":     func(columnID int) HeaderData { return EquipmentHeaderData(columnID, entity, true, false) },
		"other equipment":       func(columnID int) HeaderData { return EquipmentHeaderData(columnID, entity, false, true) },
	}
	abbreviations := map[string]string{
		"±":    "Modifier",
		"#":    "Quantity",
		"TL":   "Tech Level",
		"LC":   "Legality Class",
		"Diff": "Difficulty",
		"SL":   "Skill Level",
		"RSL":  "Relative Skill Level",
		"Pts":  "Points",
		"P#":   "Prerequisite Count",
		"ST":   "Minimum Strength",
		"Acc":  "Accuracy Bonus",
		"RoF":  "Rate of Fire",
	}
	seen := make(map[string]bool)
	for name, f := range functions {
		columns := 0
		// IDs that name no column, such as a retired one, have no title. 64 is well past the last ID of any kind.
		for columnID := range 64 {
			data := f(columnID)
			if data.Title == "" {
				continue
			}
			columns++
			switch {
			case data.TitleIsImageKey:
				c.NotEqual("", data.Name, "%s column %d, an icon, must carry a name", name, columnID)
			case len([]rune(data.Title)) <= 3:
				c.NotEqual("", data.Name, "%s column %d, titled %q, must carry the name it stands for", name, columnID,
					data.Title)
			}
			if expected, ok := abbreviations[data.Title]; ok {
				seen[data.Title] = true
				c.Equal(expected, data.Name, "%s column %d, titled %q, must be named for what it stands for", name,
					columnID, data.Title)
				c.Equal(expected, data.Detail, "%s column %d, titled %q, must explain itself in its tooltip", name,
					columnID, data.Title)
			}
			if data.Name != "" {
				c.True(data.TitleIsImageKey || abbreviations[data.Title] != "",
					"%s column %d, titled %q, is a word and needs no name", name, columnID, data.Title)
			}
		}
		c.True(columns > 1, "%s must have columns", name)
	}
	for title := range abbreviations {
		c.True(seen[title], "the abbreviation %q must be in use somewhere", title)
	}
}

// TestBonusLegacyCategoryMigratesToTags verifies that each bonus type that once stored its tags criteria under
// "category" adopts a legacy "category" when it has no "tags", and prefers "tags" when both are present.
func TestBonusLegacyCategoryMigratesToTags(t *testing.T) {
	c := check.New(t)
	for _, typ := range []feature.Type{
		feature.SkillBonus,
		feature.SkillPointBonus,
		feature.SpellBonus,
		feature.SpellPointBonus,
		feature.WeaponBonus,
	} {
		key := typ.Key()
		f := loadSingleFeature(c, `[{"type":"`+key+`","category":{"compare":"is","qualifier":"Old"}}]`)
		if f == nil {
			continue
		}
		c.Equal(typ, f.FeatureType(), "%s: type survives the load", key)
		c.Equal(textCriteria(criteria.IsText, "Old"), bonusTagsCriteria(c, f),
			"%s: a legacy category becomes the tags criteria", key)

		f = loadSingleFeature(c, `[{"type":"`+key+`","tags":{"compare":"contains","qualifier":"New"},"category":{"compare":"is","qualifier":"Old"}}]`)
		if f == nil {
			continue
		}
		c.Equal(textCriteria(criteria.ContainsText, "New"), bonusTagsCriteria(c, f),
			"%s: tags win over a legacy category", key)

		f = loadSingleFeature(c, `[{"type":"`+key+`"}]`)
		if f == nil {
			continue
		}
		c.True(bonusTagsCriteria(c, f).IsZero(), "%s: no tags and no category leaves the tags criteria empty", key)
	}
}

func TestDisabledIsClearedOffASheet(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	skill := NewSkill(e, nil, false)
	skill.SetDisabled(true)
	spell := NewSpell(e, nil, false)
	spell.SetDisabled(true)
	eqp := NewEquipment(e, nil, false)
	eqp.SetDisabled(true)
	c.True(skill.Clone(LibraryFile{}, e, nil, Copy).IsDisabled())
	c.True(spell.Clone(LibraryFile{}, e, nil, Copy).IsDisabled())
	c.True(eqp.Clone(LibraryFile{}, e, nil, Copy).IsDisabled())
	c.False(skill.Clone(LibraryFile{}, nil, nil, Copy).IsDisabled())
	c.False(spell.Clone(LibraryFile{}, nil, nil, Copy).IsDisabled())
	c.False(eqp.Clone(LibraryFile{}, nil, nil, Copy).IsDisabled())
}

func TestLoadingOffASheetClearsDisabled(t *testing.T) {
	c := check.New(t)
	trait := NewTrait(nil, nil, false)
	trait.SetDisabled(true)
	skill := NewSkill(nil, nil, false)
	skill.SetDisabled(true)
	spell := NewSpell(nil, nil, false)
	spell.SetDisabled(true)
	eqp := NewEquipment(nil, nil, false)
	eqp.SetDisabled(true)

	entity := NewEntity()
	entity.Skills = []*Skill{skill}
	entity.Spells = []*Spell{spell}
	entity.CarriedEquipment = []*Equipment{eqp}
	data, err := json.Marshal(entity)
	c.NoError(err)
	var loadedEntity Entity
	c.NoError(json.Unmarshal(data, &loadedEntity))
	c.True(loadedEntity.Skills[0].IsDisabled(), "a character sheet keeps a disabled skill")
	c.True(loadedEntity.Spells[0].IsDisabled(), "a character sheet keeps a disabled spell")
	c.True(loadedEntity.CarriedEquipment[0].IsDisabled(), "a character sheet keeps disabled equipment")

	template := NewTemplate()
	template.Traits = []*Trait{trait}
	template.Skills = []*Skill{skill}
	template.Spells = []*Spell{spell}
	template.Equipment = []*Equipment{eqp}
	data, err = json.Marshal(template)
	c.NoError(err)
	var loadedTemplate Template
	c.NoError(json.Unmarshal(data, &loadedTemplate))
	c.True(loadedTemplate.Traits[0].IsDisabled(), "a template keeps a disabled trait")
	c.False(loadedTemplate.Skills[0].IsDisabled(), "a template must not keep a disabled skill")
	c.False(loadedTemplate.Spells[0].IsDisabled(), "a template must not keep a disabled spell")
	c.False(loadedTemplate.Equipment[0].IsDisabled(), "a template must not keep disabled equipment")

	loot := NewLoot()
	loot.Equipment = []*Equipment{eqp}
	data, err = json.Marshal(loot)
	c.NoError(err)
	var loadedLoot Loot
	c.NoError(json.Unmarshal(data, &loadedLoot))
	c.False(loadedLoot.Equipment[0].IsDisabled(), "a loot sheet must not keep disabled equipment")

	list, err := json.Marshal(&listData[*Skill]{Version: jio.CurrentDataVersion, Rows: []*Skill{skill}})
	c.NoError(err)
	skills, err := NewSkillsFromFile(fstest.MapFS{"list.skl": &fstest.MapFile{Data: list}}, "list.skl")
	c.NoError(err)
	c.False(skills[0].IsDisabled(), "a skill list must not keep a disabled skill")
	list, err = json.Marshal(&listData[*Trait]{Version: jio.CurrentDataVersion, Rows: []*Trait{trait}})
	c.NoError(err)
	traits, err := NewTraitsFromFile(fstest.MapFS{"list.adq": &fstest.MapFile{Data: list}}, "list.adq")
	c.NoError(err)
	c.True(traits[0].IsDisabled(), "a trait list keeps a disabled trait")
}

// TestLoadingDropsNullNodes verifies that a null in any list of nodes a file holds -- the lists of a sheet, template,
// loot sheet or list file, and the children, modifiers and weapons of a node -- is dropped on load rather than becoming
// a nil node that panics the first thing to walk the list.
func TestLoadingDropsNullNodes(t *testing.T) {
	c := check.New(t)
	id := func(node interface{ ID() tid.TID }) string { return string(node.ID()) }
	// Each list of modifiers holds a null, a container with a null child and a leaf with a null weapon.
	modifiers := `[null,{"id":%q,"children":[null]},{"id":%q,"weapons":[null]}]`
	traitModifiers := fmt.Sprintf(modifiers, id(NewTraitModifier(nil, nil, true)),
		id(NewTraitModifier(nil, nil, false)))
	equipmentModifiers := fmt.Sprintf(modifiers, id(NewEquipmentModifier(nil, nil, true)),
		id(NewEquipmentModifier(nil, nil, false)))
	traits := fmt.Sprintf(`[null,{"id":%q,"children":[null,{"id":%q,"modifiers":%s,"weapons":[null]}]}]`,
		id(NewTrait(nil, nil, true)), id(NewTrait(nil, nil, false)), traitModifiers)
	skills := fmt.Sprintf(`[null,{"id":%q,"children":[null,{"id":%q,"weapons":[null]}]}]`,
		id(NewSkill(nil, nil, true)), id(NewSkill(nil, nil, false)))
	spells := fmt.Sprintf(`[null,{"id":%q,"children":[null,{"id":%q,"weapons":[null]}]}]`,
		id(NewSpell(nil, nil, true)), id(NewSpell(nil, nil, false)))
	equipment := fmt.Sprintf(`[null,{"id":%q,"children":[null,{"id":%q,"modifiers":%s,"weapons":[null]}]}]`,
		id(NewEquipment(nil, nil, true)), id(NewEquipment(nil, nil, false)), equipmentModifiers)
	notes := fmt.Sprintf(`[null,{"id":%q,"children":[null]}]`, id(NewNote(nil, nil, true)))

	checkTraitModifiers := func(list []*TraitModifier) {
		c.Equal(2, len(list), "trait modifiers")
		c.Equal(0, len(list[0].Children), "trait modifier children")
		c.Equal(0, len(list[1].Weapons), "trait modifier weapons")
	}
	checkEquipmentModifiers := func(list []*EquipmentModifier) {
		c.Equal(2, len(list), "equipment modifiers")
		c.Equal(0, len(list[0].Children), "equipment modifier children")
		c.Equal(0, len(list[1].Weapons), "equipment modifier weapons")
	}
	checkTraits := func(list []*Trait) {
		c.Equal(1, len(list), "traits")
		c.Equal(1, len(list[0].Children), "trait children")
		checkTraitModifiers(list[0].Children[0].Modifiers)
		c.Equal(0, len(list[0].Children[0].Weapons), "trait weapons")
	}
	checkSkills := func(list []*Skill) {
		c.Equal(1, len(list), "skills")
		c.Equal(1, len(list[0].Children), "skill children")
		c.Equal(0, len(list[0].Children[0].Weapons), "skill weapons")
	}
	checkSpells := func(list []*Spell) {
		c.Equal(1, len(list), "spells")
		c.Equal(1, len(list[0].Children), "spell children")
		c.Equal(0, len(list[0].Children[0].Weapons), "spell weapons")
	}
	checkEquipment := func(list []*Equipment) {
		c.Equal(1, len(list), "equipment")
		c.Equal(1, len(list[0].Children), "equipment children")
		checkEquipmentModifiers(list[0].Children[0].Modifiers)
		c.Equal(0, len(list[0].Children[0].Weapons), "equipment weapons")
	}
	checkNotes := func(list []*Note) {
		c.Equal(1, len(list), "notes")
		c.Equal(0, len(list[0].Children), "note children")
	}

	var entity Entity
	c.NoError(json.Unmarshal([]byte(`{"traits":`+traits+`,"skills":`+skills+`,"spells":`+spells+`,"equipment":`+
		equipment+`,"other_equipment":`+equipment+`,"notes":`+notes+`}`), &entity))
	checkTraits(entity.Traits)
	checkSkills(entity.Skills)
	checkSpells(entity.Spells)
	checkEquipment(entity.CarriedEquipment)
	checkEquipment(entity.OtherEquipment)
	checkNotes(entity.Notes)
	var legacyEntity Entity
	c.NoError(json.Unmarshal([]byte(`{"advantages":`+traits+`}`), &legacyEntity))
	checkTraits(legacyEntity.Traits)

	var tmpl Template
	c.NoError(json.Unmarshal([]byte(`{"traits":`+traits+`,"skills":`+skills+`,"spells":`+spells+`,"equipment":`+
		equipment+`,"notes":`+notes+`}`), &tmpl))
	checkTraits(tmpl.Traits)
	checkSkills(tmpl.Skills)
	checkSpells(tmpl.Spells)
	checkEquipment(tmpl.Equipment)
	checkNotes(tmpl.Notes)
	var legacyTemplate Template
	c.NoError(json.Unmarshal([]byte(`{"advantages":`+traits+`}`), &legacyTemplate))
	checkTraits(legacyTemplate.Traits)

	var loot Loot
	c.NoError(json.Unmarshal([]byte(`{"equipment":`+equipment+`,"notes":`+notes+`}`), &loot))
	checkEquipment(loot.Equipment)
	checkNotes(loot.Notes)

	listFile := func(rows string) fstest.MapFS {
		data := fmt.Sprintf(`{"version":%d,"rows":%s}`, jio.CurrentDataVersion, rows)
		return fstest.MapFS{"list": &fstest.MapFile{Data: []byte(data)}}
	}
	loadedTraits, err := NewTraitsFromFile(listFile(traits), "list")
	c.NoError(err)
	checkTraits(loadedTraits)
	loadedSkills, err := NewSkillsFromFile(listFile(skills), "list")
	c.NoError(err)
	checkSkills(loadedSkills)
	loadedSpells, err := NewSpellsFromFile(listFile(spells), "list")
	c.NoError(err)
	checkSpells(loadedSpells)
	loadedEquipment, err := NewEquipmentFromFile(listFile(equipment), "list")
	c.NoError(err)
	checkEquipment(loadedEquipment)
	loadedNotes, err := NewNotesFromFile(listFile(notes), "list")
	c.NoError(err)
	checkNotes(loadedNotes)
	loadedTraitModifiers, err := NewTraitModifiersFromFile(listFile(traitModifiers), "list")
	c.NoError(err)
	checkTraitModifiers(loadedTraitModifiers)
	loadedEquipmentModifiers, err := NewEquipmentModifiersFromFile(listFile(equipmentModifiers), "list")
	c.NoError(err)
	checkEquipmentModifiers(loadedEquipmentModifiers)
}

// TestMarshalNodeDataEnvelope verifies the envelope every node type writes through marshalNodeData: the data fields
// come first and "calc" last, a nil calc is left out while an empty one that is always wanted is still written, the
// hashed form has no calc at all, and each type keeps its own calc key.
func TestMarshalNodeDataEnvelope(t *testing.T) {
	c := check.New(t)
	marshaled := func(in any) string {
		data, err := jio.Marshal(in)
		c.NoError(err)
		return string(data)
	}

	// A note records its resolved text under its own key, and only when the text resolves to something else.
	note := NewNote(nil, nil, false)
	note.MarkDown = "Owes @who@ a favor"
	c.False(strings.Contains(marshaled(note), `"calc"`), "a note whose text resolves to itself has no calc")
	note.Replacements = map[string]string{"who": "Baron Hale"}
	out := marshaled(note)
	c.True(strings.Contains(out, `"calc":{"resolved_text":"Owes Baron Hale a favor"}`), "resolved text: %s", out)
	c.True(strings.Index(out, `"markdown"`) < strings.Index(out, `"calc"`), "the data fields precede calc: %s", out)
	c.False(strings.Contains(hashedJSON(t, note), `"calc"`), "the hashed form has no calc")
	var restored Note
	c.NoError(jio.Unmarshal([]byte(out), &restored))
	c.Equal(note.MarkDown, restored.MarkDown, "the data fields survive a round-trip")
	c.Equal(note.Replacements, restored.Replacements, "the replacements survive a round-trip")

	// The modifiers share the resolved-notes calc, and likewise drop it when there is nothing to say. Their
	// replacements are those of the item they belong to.
	tm := NewTraitModifier(nil, nil, false)
	tm.LocalNotes = "Only while @when@"
	c.False(strings.Contains(marshaled(tm), `"calc"`), "a trait modifier with unchanged notes has no calc")
	trait := NewTrait(nil, nil, false)
	trait.Replacements = map[string]string{"when": "raining"}
	tm.SetTarget(trait)
	out = marshaled(tm)
	c.True(strings.Contains(out, `"calc":{"resolved_notes":"Only while raining"}`), "trait modifier notes: %s", out)
	em := NewEquipmentModifier(nil, nil, false)
	em.LocalNotes = "Only while @when@"
	eqp := NewEquipment(nil, nil, false)
	eqp.Replacements = map[string]string{"when": "raining"}
	em.SetTarget(eqp)
	out = marshaled(em)
	c.True(strings.Contains(out, `"calc":{"resolved_notes":"Only while raining"}`), "equipment modifier notes: %s", out)

	// A trait and a piece of equipment always publish their calc, since it always has values in it.
	trait = NewTrait(nil, nil, false)
	out = marshaled(trait)
	c.True(strings.Contains(out, `"calc":{"points":0}`), "a trait always writes its points: %s", out)
	c.False(strings.Contains(out, `"resolved_notes"`), "unchanged notes are not recorded: %s", out)
	c.False(strings.Contains(hashedJSON(t, trait), `"calc"`), "the hashed trait has no calc")
	eqp = NewEquipment(nil, nil, false)
	out = marshaled(eqp)
	c.True(strings.Contains(out, `"calc":{"value":0,"extended_value":0,"weight":"0 lb","extended_weight":"0 lb"}`),
		"equipment always writes its totals: %s", out)
	c.False(strings.Contains(hashedJSON(t, eqp), `"calc"`), "the hashed equipment has no calc")
}

// TestNodeSyncDataIsInlined verifies that embedding NodeSyncData leaves the on-disk format of each node type as it was
// when the fields were declared directly: the shared members are top-level keys of the node, in their original order,
// rather than nested under an object of their own, and they survive a round trip.
func TestNodeSyncDataIsInlined(t *testing.T) {
	c := check.New(t)
	fill := func(n *NodeSyncData) {
		n.Name = "Name"
		n.PageRef = "B1"
		n.PageRefHighlight = "highlight"
		n.LocalNotes = "notes"
		n.Tags = []string{"one", "two"}
	}
	skill := NewSkill(nil, nil, false)
	fill(&skill.SkillSyncData)
	spell := NewSpell(nil, nil, false)
	fill(&spell.SpellSyncData)
	trait := NewTrait(nil, nil, false)
	fill(&trait.NodeSyncData)
	traitMod := NewTraitModifier(nil, nil, false)
	fill(&traitMod.TraitModifierSyncData)
	eqpMod := NewEquipmentModifier(nil, nil, false)
	fill(&eqpMod.EquipmentModifierSyncData)
	expected := []string{"name", "reference", "reference_highlight", "local_notes", "tags"}
	for _, tc := range []struct {
		name    string
		node    any
		restore func([]byte) (*NodeSyncData, error)
	}{
		{"skill", skill, reloadInto(func(s *Skill) *NodeSyncData { return &s.SkillSyncData })},
		{"spell", spell, reloadInto(func(s *Spell) *NodeSyncData { return &s.SpellSyncData })},
		{"trait", trait, reloadInto(func(tr *Trait) *NodeSyncData { return &tr.NodeSyncData })},
		{"trait modifier", traitMod, reloadInto(func(m *TraitModifier) *NodeSyncData {
			return &m.TraitModifierSyncData
		})},
		{"equipment modifier", eqpMod, reloadInto(func(m *EquipmentModifier) *NodeSyncData {
			return &m.EquipmentModifierSyncData
		})},
	} {
		data, err := jio.Marshal(tc.node)
		c.NoError(err, tc.name)
		keys := topLevelKeys(t, data)
		c.False(slices.ContainsFunc(keys, func(k string) bool { return strings.HasSuffix(k, "SyncData") }),
			"%s: the shared fields must not be nested", tc.name)
		var shared []string
		for _, key := range keys {
			if slices.Contains(expected, key) {
				shared = append(shared, key)
			}
		}
		c.Equal(expected, shared, "%s: the shared fields are top-level keys in declaration order", tc.name)
		restored, err := tc.restore(data)
		c.NoError(err, tc.name)
		c.Equal(NodeSyncData{
			Name:             "Name",
			PageRef:          "B1",
			PageRefHighlight: "highlight",
			LocalNotes:       "notes",
			Tags:             []string{"one", "two"},
		}, *restored, "%s: the shared fields round-trip", tc.name)
	}
}

// topLevelKeys returns the member names of the JSON object in data, in the order they appear.
func topLevelKeys(t *testing.T, data []byte) []string {
	t.Helper()
	c := check.New(t)
	dec := jsontext.NewDecoder(bytes.NewReader(data))
	tok, err := dec.ReadToken()
	c.NoError(err)
	c.Equal(jsontext.BeginObject, tok, "a JSON object")
	var keys []string
	for dec.PeekKind() != '}' {
		tok, err = dec.ReadToken()
		c.NoError(err)
		keys = append(keys, tok.String())
		c.NoError(dec.SkipValue())
	}
	return keys
}

// unsortedTags is the order a user might enter tags in; sortedTags is the order they are kept in.
var (
	unsortedTags = []string{"Zeta", "alpha", "Mid"}
	sortedTags   = []string{"Mid", "Zeta", "alpha"}
)

// savedLists is the part of a saved node the tag sorting tests look at.
type savedLists struct {
	Tags    []string `json:"tags"`
	College []string `json:"college"`
}

// savedListsOf saves the node and returns the tags and colleges in the order they were written.
func savedListsOf(t *testing.T, node any) savedLists {
	t.Helper()
	data, err := jio.Marshal(node)
	if err != nil {
		t.Fatal(err)
	}
	var lists savedLists
	if err = json.Unmarshal(data, &lists); err != nil {
		t.Fatal(err)
	}
	return lists
}

// TestApplyToSortsTags verifies that committing an editor's data sorts the tags, whatever order they were entered in,
// so that the order shows neither in the item nor in the file it is saved to (#1217).
func TestApplyToSortsTags(t *testing.T) {
	c := check.New(t)

	trait := NewTrait(nil, nil, false)
	var traitData TraitEditData
	traitData.CopyFrom(trait)
	traitData.Tags = unsortedTags
	traitData.ApplyTo(trait)
	c.Equal(sortedTags, trait.Tags, "trait")

	traitModifier := NewTraitModifier(nil, nil, false)
	var traitModifierData TraitModifierEditData
	traitModifierData.CopyFrom(traitModifier)
	traitModifierData.Tags = unsortedTags
	traitModifierData.ApplyTo(traitModifier)
	c.Equal(sortedTags, traitModifier.Tags, "trait modifier")

	skill := NewSkill(nil, nil, false)
	var skillData SkillEditData
	skillData.CopyFrom(skill)
	skillData.Tags = unsortedTags
	skillData.ApplyTo(skill)
	c.Equal(sortedTags, skill.Tags, "skill")

	spell := NewSpell(nil, nil, false)
	var spellData SpellEditData
	spellData.CopyFrom(spell)
	spellData.Tags = unsortedTags
	spellData.College = CollegeList{"Water", "Air", "Fire"}
	spellData.ApplyTo(spell)
	c.Equal(sortedTags, spell.Tags, "spell")
	c.Equal(CollegeList{"Air", "Fire", "Water"}, spell.College, "spell colleges")

	equipment := NewEquipment(nil, nil, false)
	var equipmentData EquipmentEditData
	equipmentData.CopyFrom(equipment)
	equipmentData.Tags = unsortedTags
	equipmentData.ApplyTo(equipment)
	c.Equal(sortedTags, equipment.Tags, "equipment")

	equipmentModifier := NewEquipmentModifier(nil, nil, false)
	var equipmentModifierData EquipmentModifierEditData
	equipmentModifierData.CopyFrom(equipmentModifier)
	equipmentModifierData.Tags = unsortedTags
	equipmentModifierData.ApplyTo(equipmentModifier)
	c.Equal(sortedTags, equipmentModifier.Tags, "equipment modifier")

	note := NewNote(nil, nil, false)
	var noteData NoteEditData
	noteData.CopyFrom(note)
	noteData.Tags = unsortedTags
	noteData.ApplyTo(note)
	c.Equal(sortedTags, note.Tags, "note")
}

// TestMarshalSortsTags verifies that saving a node writes its tags sorted even when it holds them in another order, as
// one built by code rather than through an editor might, and that a spell's colleges are written the same way.
func TestMarshalSortsTags(t *testing.T) {
	c := check.New(t)

	trait := NewTrait(nil, nil, false)
	trait.Tags = append([]string(nil), unsortedTags...)
	c.Equal(sortedTags, savedListsOf(t, trait).Tags, "trait")

	traitModifier := NewTraitModifier(nil, nil, false)
	traitModifier.Tags = append([]string(nil), unsortedTags...)
	c.Equal(sortedTags, savedListsOf(t, traitModifier).Tags, "trait modifier")

	skill := NewSkill(nil, nil, false)
	skill.Tags = append([]string(nil), unsortedTags...)
	c.Equal(sortedTags, savedListsOf(t, skill).Tags, "skill")

	spell := NewSpell(nil, nil, false)
	spell.Tags = append([]string(nil), unsortedTags...)
	spell.College = CollegeList{"Water", "Air", "Fire"}
	lists := savedListsOf(t, spell)
	c.Equal(sortedTags, lists.Tags, "spell")
	c.Equal([]string{"Air", "Fire", "Water"}, lists.College, "spell colleges")

	equipment := NewEquipment(nil, nil, false)
	equipment.Tags = append([]string(nil), unsortedTags...)
	c.Equal(sortedTags, savedListsOf(t, equipment).Tags, "equipment")

	equipmentModifier := NewEquipmentModifier(nil, nil, false)
	equipmentModifier.Tags = append([]string(nil), unsortedTags...)
	c.Equal(sortedTags, savedListsOf(t, equipmentModifier).Tags, "equipment modifier")

	note := NewNote(nil, nil, false)
	note.Tags = append([]string(nil), unsortedTags...)
	c.Equal(sortedTags, savedListsOf(t, note).Tags, "note")
}
