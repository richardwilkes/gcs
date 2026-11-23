// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package export

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/model/gurps/gurpstest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// runLegacyExport runs the legacy (non-Go-template) exporter over the given template text and returns its output.
func runLegacyExport(t *testing.T, c check.Checker, entity *gurps.Entity, tmpl string) string {
	t.Helper()
	entity.Recalculate()
	outPath := filepath.Join(t.TempDir(), "out.txt")
	c.NoError(legacyTextExport(entity, []byte(tmpl), outPath))
	data, err := os.ReadFile(outPath)
	c.NoError(err)
	return string(data)
}

// TestLegacyExportAttributePointsKeys verifies that each fixed *_POINTS key emits the point cost of its own attribute.
func TestLegacyExportAttributePointsKeys(t *testing.T) {
	// Each adjustment gives its attribute a cost no other one has, so a key reading the wrong attribute fails.
	rows := []struct {
		key        string
		attrID     string
		adjustment fxp.Int
		expected   string
	}{
		{key: "ST_POINTS", attrID: "st", adjustment: fxp.One, expected: "10"},
		{key: "DX_POINTS", attrID: "dx", adjustment: fxp.Two, expected: "40"},
		{key: "IQ_POINTS", attrID: "iq", adjustment: fxp.Three, expected: "60"},
		{key: "HT_POINTS", attrID: "ht", adjustment: fxp.Two, expected: "20"},
		{key: "PERCEPTION_POINTS", attrID: "per", adjustment: fxp.Three, expected: "15"},
		{key: "WILL_POINTS", attrID: "will", adjustment: fxp.Five, expected: "25"},
		{key: "FP_POINTS", attrID: "fp", adjustment: fxp.One, expected: "3"},
		{key: "HP_POINTS", attrID: "hp", adjustment: fxp.Four, expected: "8"},
		{key: "BASIC_SPEED_POINTS", attrID: "basic_speed", adjustment: fxp.Quarter, expected: "5"},
		{key: "BASIC_MOVE_POINTS", attrID: "basic_move", adjustment: fxp.Seven, expected: "35"},
	}
	e := gurps.NewEntity()
	for _, one := range rows {
		e.Attributes.Set[one.attrID].Adjustment = one.adjustment
	}
	for _, one := range rows {
		t.Run(one.key, func(t *testing.T) {
			c := check.New(t)
			c.Equal(one.expected, runLegacyExport(t, c, e, "@"+one.key))
		})
	}
}

// TestLegacyExportDescriptionNotesKeys verifies that the DESCRIPTION_NOTES* and DESCRIPTION_MODIFIER_NOTES* keys wrap
// their text as their suffix asks, encode it like every other text-emitting key unless encoding is off, and emit
// nothing at all for blank text.
func TestLegacyExportDescriptionNotesKeys(t *testing.T) {
	e := gurps.NewEntity()
	skill := gurps.NewSkill(e, nil, false)
	skill.Name = "Brawling"
	skill.LocalNotes = "a<b>c & \"d\" 'e'\nf é"
	e.Skills = append(e.Skills, skill)
	trait := gurps.NewTrait(e, nil, false)
	trait.Name = "Greed"
	trait.LocalNotes = " \n "
	trait.SelfControl = selfctrl.CR12
	mod := gurps.NewTraitModifier(e, nil, false)
	mod.Name = "Mitigator"
	trait.AddModifiers(mod)
	e.Traits = []*gurps.Trait{trait}
	const encoded = "a&lt;b&gt;c &amp; &quot;d&quot; &apos;e&apos;<br>f &#233;"
	for _, one := range []struct {
		name     string
		tmpl     string
		expected string
	}{
		{
			name:     "paren",
			tmpl:     "@SKILLS_LOOP_START[@DESCRIPTION_NOTES_PAREN]@SKILLS_LOOP_END",
			expected: "[ (" + encoded + ")]",
		},
		{
			name:     "curly",
			tmpl:     "@SKILLS_LOOP_START[@DESCRIPTION_NOTES_CURLY]@SKILLS_LOOP_END",
			expected: "[ {" + encoded + "}]",
		},
		{
			name:     "unsuffixed",
			tmpl:     "@SKILLS_LOOP_START[@DESCRIPTION_NOTES]@SKILLS_LOOP_END",
			expected: "[" + encoded + "]",
		},
		{
			name:     "encoding off",
			tmpl:     "@ENCODING_OFF@SKILLS_LOOP_START[@DESCRIPTION_NOTES_BRACKET]@SKILLS_LOOP_END",
			expected: "[ [a<b>c & \"d\" 'e'\nf é]]",
		},
		{
			// Modifier notes arrive as plain newline-separated text, so the line break is encoded, not escaped markup.
			name:     "modifier notes",
			tmpl:     "@ADVANTAGES_LOOP_START[@DESCRIPTION_MODIFIER_NOTES_BRACKET]@ADVANTAGES_LOOP_END",
			expected: "[ [Self-Control Roll (CR): 12 or less (Resist quite often)<br>Mitigator]]",
		},
		{
			name:     "blank",
			tmpl:     "@ADVANTAGES_LOOP_START[@DESCRIPTION_NOTES_PAREN]@ADVANTAGES_LOOP_END",
			expected: "[]",
		},
	} {
		t.Run(one.name, func(t *testing.T) {
			c := check.New(t)
			c.Equal(one.expected, runLegacyExport(t, c, e, one.tmpl))
		})
	}
}

// TestLegacyExportLoopCountsSkipDisabled verifies that the skills, spells and equipment loop counts leave out disabled
// items, as their loops do, and that @SKILLS_LOOP_COUNT counts containers, which @SKILLS_LOOP_START also iterates over.
func TestLegacyExportLoopCountsSkipDisabled(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	for _, name := range []string{"A", "B"} {
		skill := gurps.NewSkill(e, nil, false)
		skill.Name = name
		e.Skills = append(e.Skills, skill)
		spell := gurps.NewSpell(e, nil, false)
		spell.Name = name
		e.Spells = append(e.Spells, spell)
		carried := gurps.NewEquipment(e, nil, false)
		carried.Name = name
		e.CarriedEquipment = append(e.CarriedEquipment, carried)
		other := gurps.NewEquipment(e, nil, false)
		other.Name = name
		e.OtherEquipment = append(e.OtherEquipment, other)
	}
	e.Skills[1].SetDisabled(true)
	e.Spells[1].SetDisabled(true)
	e.CarriedEquipment[1].SetDisabled(true)
	e.OtherEquipment[1].SetDisabled(true)
	c.Equal("1|<A>1|<A>1|<A>1|<A>", runLegacyExport(t, c, e,
		"@SKILLS_LOOP_COUNT|@SKILLS_LOOP_START<@DESCRIPTION>@SKILLS_LOOP_END"+
			"@SPELLS_LOOP_COUNT|@SPELLS_LOOP_START<@DESCRIPTION_PRIMARY>@SPELLS_LOOP_END"+
			"@EQUIPMENT_LOOP_COUNT|@EQUIPMENT_LOOP_START<@DESCRIPTION>@EQUIPMENT_LOOP_END"+
			"@OTHER_EQUIPMENT_LOOP_COUNT|@OTHER_EQUIPMENT_LOOP_START<@DESCRIPTION>@OTHER_EQUIPMENT_LOOP_END"))

	// Containers are counted, since the loop visits them too.
	e = gurps.NewEntity()
	group := gurps.NewSkill(e, nil, true)
	group.Name = "Group"
	child := gurps.NewSkill(e, group, false)
	child.Name = "Child"
	group.Children = append(group.Children, child)
	e.Skills = append(e.Skills, group)
	c.Equal("2|<GROUP><ITEM>", runLegacyExport(t, c, e, "@SKILLS_LOOP_COUNT|@SKILLS_LOOP_START<@TYPE>@SKILLS_LOOP_END"))
}

// TestLegacyExportConditionalModifierGroups verifies that the legacy exporter writes the reactions and conditional
// modifiers out flat -- the loop counts and IDs cover the members, not the group containers -- and that a member can
// name its group with @GROUP.
func TestLegacyExportConditionalModifierGroups(t *testing.T) {
	c := check.New(t)
	e := newGroupedModifiersEntity(t)
	c.Equal("3|<0:from allies|Combat|+3><1:from foes|Combat|+2><2:from everyone||+1>",
		runLegacyExport(t, c, e, "@REACTION_LOOP_COUNT|@REACTION_LOOP_START<@ID:@SITUATION|@GROUP|@MODIFIER>@REACTION_LOOP_END"))
	c.Equal("2|<0:to hit|Combat|+1><1:to dodge||+2>",
		runLegacyExport(t, c, e, "@CONDITIONAL_MODIFIERS_LOOP_COUNT|@CONDITIONAL_MODIFIERS_LOOP_START<@ID:@SITUATION|@GROUP|@MODIFIER>@CONDITIONAL_MODIFIERS_LOOP_END"))
}

// TestLegacyExportNotesSeparator verifies that the separator between top-level notes is encoded along with the notes
// themselves, so that HTML exports keep the blank line between them.
func TestLegacyExportNotesSeparator(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	first := gurps.NewNote(e, nil, false)
	first.MarkDown = "First"
	second := gurps.NewNote(e, nil, false)
	second.MarkDown = "Second"
	e.Notes = append(e.Notes, first, second)
	c.Equal("First<br><br>Second", runLegacyExport(t, c, e, "@NOTES"))
	c.Equal("First\n\nSecond", runLegacyExport(t, c, e, "@ENCODING_OFF@NOTES"))
}

// TestLegacyExportAttributeNameSuffixes verifies what each attribute key suffix emits, that the more specific suffixes
// are matched before the less specific ones, and that a suffix match that doesn't resolve to an attribute falls through
// to the next candidate.
func TestLegacyExportAttributeNameSuffixes(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	c.Equal("ST|Strength|Strength (ST)|10", runLegacyExport(t, c, e,
		"@ST_NAME|@ST_FULL_NAME|@ST_COMBINED_NAME|@ST_CURRENT"))

	// Vision has no fixed *_POINTS key, so this reaches the generic suffix.
	e.Attributes.Set["vision"].Adjustment = fxp.Three
	c.Equal("6", runLegacyExport(t, c, e, "@VISION_POINTS"))

	// An attribute whose own ID ends with "_full" must still resolve via the "_name" suffix.
	e.SheetSettings.Attributes.Set["hit_full"] = &gurps.AttributeDef{
		DefID: "hit_full",
		Type:  attribute.Integer,
		Name:  "Hit Full",
		Base:  "3",
		Order: len(e.SheetSettings.Attributes.Set),
	}
	e.Attributes.Set["hit_full"] = gurps.NewAttribute(e, "hit_full", len(e.Attributes.Set))
	c.Equal("Hit Full|3", runLegacyExport(t, c, e, "@HIT_FULL_NAME|@HIT_FULL_CURRENT"))
}

func newTestWeapon(owner gurps.WeaponOwner, melee bool, usage string) *gurps.Weapon {
	w := gurps.NewWeapon(owner, melee)
	w.Usage = usage
	return w
}

// TestLegacyExportWeaponLoops verifies that the flat and hierarchical weapon loops behave the same way for melee and
// ranged weapons: the flat loops visit every attack mode as its own weapon and have no attack modes of their own, while
// the hierarchical loops visit each distinct weapon once, as their counts report, and run the ATTACK_MODES body for
// each of its modes using the same melee or ranged key set as the enclosing loop.
func TestLegacyExportWeaponLoops(t *testing.T) {
	e := gurps.NewEntity()
	// Melee has 3 modes on 1 weapon and ranged has 4 modes on 2 weapons, so no two of the four loop counts match.
	karate := gurps.NewTrait(e, nil, false)
	karate.Name = "Karate"
	karate.Weapons = append(karate.Weapons, newTestWeapon(karate, true, "Punch"), newTestWeapon(karate, true, "Kick"),
		newTestWeapon(karate, true, "Chop"), newTestWeapon(karate, false, "Spit"), newTestWeapon(karate, false, "Sneeze"))
	innate := gurps.NewTrait(e, nil, false)
	innate.Name = "Innate Attack"
	innate.Weapons = append(innate.Weapons, newTestWeapon(innate, false, "Bolt"), newTestWeapon(innate, false, "Blast"))
	e.Traits = []*gurps.Trait{karate, innate} // Drops the "Natural Attacks" a new entity starts with.
	const (
		noReach = "Unidentified key: &quot;REACH&quot;"
		noROF   = "Unidentified key: &quot;ROF&quot;"
	)
	for _, one := range []struct {
		name     string
		tmpl     string
		expected string
	}{
		{
			name: "melee counts",
			tmpl: "@MELEE_LOOP_COUNT|@HIERARCHICAL_MELEE_LOOP_COUNT|" +
				"@HIERARCHICAL_MELEE_LOOP_START(@DESCRIPTION_PRIMARY:@ATTACK_MODES_LOOP_COUNT)@HIERARCHICAL_MELEE_LOOP_END",
			expected: "3|1|(Karate:3)",
		},
		{
			name: "ranged counts",
			tmpl: "@RANGED_LOOP_COUNT|@HIERARCHICAL_RANGED_LOOP_COUNT|" +
				"@HIERARCHICAL_RANGED_LOOP_START(@DESCRIPTION_PRIMARY:@ATTACK_MODES_LOOP_COUNT)@HIERARCHICAL_RANGED_LOOP_END",
			expected: "4|2|(Innate Attack:2)(Karate:2)",
		},
		{
			name:     "flat loop numbers every mode",
			tmpl:     "@RANGED_LOOP_START(@ID:@USAGE)@RANGED_LOOP_END",
			expected: "(0:Blast)(1:Bolt)(2:Sneeze)(3:Spit)",
		},
		{
			name:     "flat loop rejects attack modes",
			tmpl:     "@RANGED_LOOP_START(@ATTACK_MODES_LOOP_COUNT:@ATTACK_MODES_LOOP_START)@RANGED_LOOP_END",
			expected: strings.Repeat("(0:Unidentified key: &quot;ATTACK_MODES_LOOP_START&quot;)", 4),
		},
		{
			name: "hierarchical ranged modes",
			tmpl: "@HIERARCHICAL_RANGED_LOOP_START[@DESCRIPTION_PRIMARY:" +
				"@ATTACK_MODES_LOOP_START<@ID:@USAGE:@ROF:@REACH>@ATTACK_MODES_LOOP_END]@HIERARCHICAL_RANGED_LOOP_END",
			expected: "[Innate Attack:<0:Blast:1:" + noReach + "><1:Bolt:1:" + noReach + ">]" +
				"[Karate:<0:Sneeze:1:" + noReach + "><1:Spit:1:" + noReach + ">]",
		},
		{
			name: "hierarchical melee modes",
			tmpl: "@HIERARCHICAL_MELEE_LOOP_START[@DESCRIPTION_PRIMARY:" +
				"@ATTACK_MODES_LOOP_START<@ID:@USAGE:@REACH:@ROF>@ATTACK_MODES_LOOP_END]@HIERARCHICAL_MELEE_LOOP_END",
			expected: "[Karate:<0:Chop:1:" + noROF + "><1:Kick:1:" + noROF + "><2:Punch:1:" + noROF + ">]",
		},
	} {
		t.Run(one.name, func(t *testing.T) {
			c := check.New(t)
			c.Equal(one.expected, runLegacyExport(t, c, e, one.tmpl))
		})
	}
}

// TestLegacyExportKeyScanner verifies that the top-level template and loop bodies are scanned by the same rules -- a
// key ends at the first byte outside [A-Za-z0-9_], which is read again as text unless enhanced key parsing is on and it
// is a closing '@' -- and that a key running up to the very end of the template is still emitted. Whether '@' starts a
// key inside a loop body is decided from the loop body itself, not from a fixed byte in the outer template.
func TestLegacyExportKeyScanner(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	n := gurps.NewNote(e, nil, false)
	n.MarkDown = "Note1"
	e.Notes = append(e.Notes, n)
	c.Equal("10 x10", runLegacyExport(t, c, e, "@ST x@DX"))

	// "@7" is literal text at the top level, so it must be literal text inside a loop body too.
	c.Equal("[@7]", runLegacyExport(t, c, e, "@NOTES_LOOP_START[@7]@NOTES_LOOP_END"))

	// A digit following the loop-end marker must not turn every key in the loop body into literal text.
	c.Equal("[Note1]7", runLegacyExport(t, c, e, "@NOTES_LOOP_START[@NOTE]@NOTES_LOOP_END7"))

	c.Equal("\n10xNote1yz|10",
		runLegacyExport(t, c, e, "@ENHANCED_KEY_PARSING\n@ST@x@NOTES_LOOP_START@@NOTE@y@NOTES_LOOP_END@z|@DX"))
}

// TestLegacyExportHitLocationEquipment verifies which pieces of equipment the hit location loop lists as providing DR
// to a location: only carried, really-equipped ones, counting both their own DR bonuses and those of their enabled,
// non-container modifiers, with switchable bonuses honoring the owning item's switch, and each item named at most once
// no matter how many of its bonuses reach the location.
func TestLegacyExportHitLocationEquipment(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()

	// Two DR bonuses that both reach the torso; the item must still only be listed once.
	gurpstest.AddCarriedEquipmentWithFeatures(
		e, "Plate Armor",
		gurpstest.NewDRBonus(fxp.Two, gurps.AllID, gurps.TorsoID),
		gurpstest.NewDRBonus(fxp.One, gurps.AllID, "Torso"), // the same location, with a different case
	)

	// DR that only reaches the skull through an enabled modifier.
	helm := gurpstest.AddCarriedEquipmentWithFeatures(e, "Helmet")
	helm.AddModifiers(newTestDRBonusModifier(e, "Face Guard", gurpstest.NewDRBonus(fxp.Three, gurps.AllID, "skull")))

	// A disabled modifier contributes nothing, exactly as it does when features are collected.
	cloak := gurpstest.AddCarriedEquipmentWithFeatures(e, "Cloak")
	disabled := newTestDRBonusModifier(e, "Hood", gurpstest.NewDRBonus(fxp.Five, gurps.AllID, "skull"))
	disabled.SetDisabled(true)
	cloak.AddModifiers(disabled)

	// Switchable bonuses, on the item and on one of its modifiers, only count while the item's switch is on.
	switchable := gurpstest.NewDRBonus(fxp.Seven, gurps.AllID, "skull")
	switchable.Switchable = true
	cape := gurpstest.AddCarriedEquipmentWithFeatures(e, "Cape", switchable)
	switchableOnMod := gurpstest.NewDRBonus(fxp.Nine, gurps.AllID, "skull")
	switchableOnMod.Switchable = true
	cape.AddModifiers(newTestDRBonusModifier(e, "Lining", switchableOnMod))

	// A "this armor" bonus needs no examination of its own: it covers the locations the item's and its modifiers'
	// located bonuses name, and those are what get scanned, so the modifier's skull bonus is what lists the robe.
	robe := gurpstest.AddCarriedEquipmentWithFeatures(e, "Robe", gurpstest.NewDRBonus(fxp.Six, gurps.AllID)) // no locations, i.e. "this armor"
	robe.AddModifiers(newTestDRBonusModifier(e, "Cowl", gurpstest.NewDRBonus(fxp.One, gurps.AllID, "skull")))

	// Neither an unequipped carried item nor an item in the other equipment list contributes DR to the character.
	gurpstest.AddCarriedEquipmentWithFeatures(e, "Stowed Helm", gurpstest.NewDRBonus(fxp.Eight, gurps.AllID, "skull")).Equipped = false
	spare := gurps.NewEquipment(e, nil, false)
	spare.Name = "Spare Helm"
	spare.Features = gurps.Features{gurpstest.NewDRBonus(fxp.Eight, gurps.AllID, "skull")}
	e.OtherEquipment = append(e.OtherEquipment, spare)

	const tmpl = "@HIT_LOCATION_LOOP_START[@WHERE:@EQUIPMENT]@HIT_LOCATION_LOOP_END"
	out := runLegacyExport(t, c, e, tmpl)
	c.Contains(out, "[Skull:Helmet, Robe]", "only the armor actually granting DR to the skull is listed")
	c.Contains(out, "[Torso:Plate Armor]", "an item with two DR bonuses reaching the same location is listed once")

	// EQUIPMENT_FORMATTED writes the same list as one paragraph per item.
	c.Contains(runLegacyExport(t, c, e,
		"@HIT_LOCATION_LOOP_START[@WHERE:@EQUIPMENT_FORMATTED]@HIT_LOCATION_LOOP_END"),
		"[Skull:<p>Helmet</p>\n<p>Robe</p>\n]")

	// Throwing the cape's switch brings both of its switchable bonuses into play.
	cape.SwitchedOn = true
	c.Contains(runLegacyExport(t, c, e, tmpl), "[Skull:Helmet, Cape, Robe]",
		"switching the cape on adds it to the skull's list")
}

func newTestDRBonusModifier(owner gurps.DataOwner, name string, bonus *gurps.DRBonus) *gurps.EquipmentModifier {
	mod := gurps.NewEquipmentModifier(owner, nil, false)
	mod.Name = name
	mod.Features = gurps.Features{bonus}
	return mod
}

// TestLegacyExportSharedNodeKeys verifies that the keys shared by the trait, skill, spell, equipment and note loops
// behave identically across those loops while each loop's deliberate differences, such as the TYPE each gives its
// containers, are preserved.
func TestLegacyExportSharedNodeKeys(t *testing.T) {
	e := gurps.NewEntity()

	group := gurps.NewTrait(e, nil, true)
	group.Name = "Group"
	group.ContainerType = container.AlternativeAbilities
	group.PageRef = "B1"
	trait := gurpstest.NewTraitNeedingMissingTrait(e, "Greed")
	trait.SetParent(group)
	group.Children = append(group.Children, trait)
	mod := gurps.NewTraitModifier(e, nil, false)
	mod.Name = "Mitigator"
	mod.LocalNotes = "mod note"
	trait.Modifiers = append(trait.Modifiers, mod)
	e.Traits = []*gurps.Trait{group}

	skill := gurps.NewSkill(e, nil, false)
	skill.Name = "Brawling"
	skill.PageRef = "B2"
	e.Skills = append(e.Skills, skill)

	spell := gurps.NewSpell(e, nil, false)
	spell.Name = "Fireball"
	spell.LocalNotes = "spell note"
	e.Spells = append(e.Spells, spell)

	eqp := gurps.NewEquipment(e, nil, false)
	eqp.Name = "Rock"
	eqpMod := gurps.NewEquipmentModifier(e, nil, false)
	eqpMod.Name = "Sharp"
	eqpMod.LocalNotes = "eqp mod note"
	eqp.Modifiers = append(eqp.Modifiers, eqpMod)
	e.CarriedEquipment = append(e.CarriedEquipment, eqp)

	noteGroup := gurps.NewNote(e, nil, true)
	noteGroup.MarkDown = "Notes"
	note := gurps.NewNote(e, noteGroup, false)
	note.MarkDown = "Note1"
	note.PageRef = "B3"
	noteGroup.Children = append(noteGroup.Children, note)
	e.Notes = append(e.Notes, noteGroup)

	// TYPE emits the container type, PARENT_ID is empty for top-level nodes, the prereq failure shows up in both
	// SATISFIED and STYLE_INDENT_WARNING, and MODIFIER_NOTES_FOR_ resolves the active modifier.
	t.Run("traits", func(t *testing.T) {
		c := check.New(t)
		c.Equal("ALTERNATIVE_ABILITIES|"+string(group.ID())+"||B1|Y||0|\n"+
			"ITEM|"+string(trait.ID())+"|"+string(group.ID())+"||N| style=\"padding-left: 12px;color: red;\" |3|mod note\n",
			runLegacyExport(t, c, e, "@ADVANTAGES_LOOP_START@TYPE|@ID|@PARENT_ID|@REF|@SATISFIED|"+
				"@STYLE_INDENT_WARNING|@DEPTHx3|@MODIFIER_NOTES_FOR_Mitigator\n@ADVANTAGES_LOOP_END"))
	})

	// The shared keys work, but skills have no MODIFIER_NOTES_FOR_ key.
	t.Run("skills", func(t *testing.T) {
		c := check.New(t)
		c.Equal("ITEM|"+string(skill.ID())+"|B2|Y|Brawling|Unidentified key: &quot;MODIFIER_NOTES_FOR_X&quot;",
			runLegacyExport(t, c, e, "@SKILLS_LOOP_START@TYPE|@ID|@REF|@SATISFIED|@DESCRIPTION_PRIMARY|"+
				"@MODIFIER_NOTES_FOR_X@SKILLS_LOOP_END"))
	})

	// DESCRIPTION writes the notes and rituals as separate notes, DESCRIPTION_NOTES merges them, and
	// DESCRIPTION_MODIFIER_NOTES is silent rather than unidentified.
	t.Run("spells", func(t *testing.T) {
		c := check.New(t)
		rituals := spell.Rituals()
		c.NotEqual("", rituals)
		c.Equal("Fireball<div class=\"note\">spell note</div><div class=\"note\">"+rituals+"</div>| (spell note; "+
			rituals+")||Fireball",
			runLegacyExport(t, c, e, "@SPELLS_LOOP_START@DESCRIPTION|@DESCRIPTION_NOTES_PAREN|"+
				"@DESCRIPTION_MODIFIER_NOTES_PAREN|@DESCRIPTION_PRIMARY@SPELLS_LOOP_END"))
	})

	// DESCRIPTION writes the note naming what a spell defaults to ahead of its notes, as for a skill, and so does
	// DESCRIPTION_MODIFIER_NOTES.
	t.Run("spell with a default", func(t *testing.T) {
		c := check.New(t)
		entity, lore := newSpellDefaultEntity()
		rituals := lore.Rituals()
		c.NotEqual("", rituals)
		c.Equal("Lore<div class=\"note\">Default: Thaumatology-2</div><div class=\"note\">"+rituals+"</div>"+
			"| (Default: Thaumatology-2)", runLegacyExport(t, c, entity,
			"@SPELLS_LOOP_START@DESCRIPTION|@DESCRIPTION_MODIFIER_NOTES_PAREN@SPELLS_LOOP_END"))
	})

	// MODIFIER_NOTES_FOR_ resolves the active modifier and DESCRIPTION includes the modifier notes.
	t.Run("equipment", func(t *testing.T) {
		c := check.New(t)
		c.Equal("eqp mod note|Rock<div class=\"note\">Sharp (eqp mod note)</div>",
			runLegacyExport(t, c, e, "@EQUIPMENT_LOOP_START@MODIFIER_NOTES_FOR_Sharp|@DESCRIPTION@EQUIPMENT_LOOP_END"))
	})

	// TYPE tells a group from a physical container.
	t.Run("equipment type", func(t *testing.T) {
		c := check.New(t)
		c.Equal("Backpack|CONTAINER\nKit|GROUP\nRope|ITEM\n", runLegacyExport(t, c, newContainerTypesEntity(),
			"@EQUIPMENT_LOOP_START@DESCRIPTION|@TYPE\n@EQUIPMENT_LOOP_END"))
	})

	// The identity and depth keys work, but there are no description or prerequisite keys and the indent warning is
	// never red.
	t.Run("notes", func(t *testing.T) {
		c := check.New(t)
		c.Equal("GROUP|"+string(noteGroup.ID())+"||||0\n"+
			"ITEM|"+string(note.ID())+"|"+string(noteGroup.ID())+"|B3| style=\"padding-left: 12px;\" |2\n",
			runLegacyExport(t, c, e, "@NOTES_LOOP_START@TYPE|@ID|@PARENT_ID|@REF|@STYLE_INDENT_WARNING|@DEPTHx2\n"+
				"@NOTES_LOOP_END"))
		c.Equal(strings.Repeat("Unidentified key: &quot;DESCRIPTION&quot;|Unidentified key: &quot;SATISFIED&quot;|"+
			"Unidentified key: &quot;DESCRIPTION_NOTES&quot;\n", 2),
			runLegacyExport(t, c, e, "@NOTES_LOOP_START@DESCRIPTION|@SATISFIED|@DESCRIPTION_NOTES\n@NOTES_LOOP_END"))
	})
}

// TestLegacyExportAttributeLoops verifies that the primary, secondary and point pool loops (and their counts) classify
// attributes by their resolved kind: an explicit placement overrides the base-derived classification, and a hidden pool
// is omitted just as it is from the Go-template export.
func TestLegacyExportAttributeLoops(t *testing.T) {
	c := check.New(t)
	e := newAttributeKindsEntity()
	c.Equal("5|<st=10><dx=10><iq=10><ht=10><forced=10>|"+
		"9|<will><fright_check><per><vision><hearing><taste_smell><touch><basic_speed><basic_move>|"+
		"2|<fp:10/10><hp:10/10>",
		runLegacyExport(t, c, e, "@PRIMARY_ATTRIBUTE_LOOP_COUNT|"+
			"@PRIMARY_ATTRIBUTE_LOOP_START<@ID=@VALUE>@PRIMARY_ATTRIBUTE_LOOP_END|"+
			"@SECONDARY_ATTRIBUTE_LOOP_COUNT|@SECONDARY_ATTRIBUTE_LOOP_START<@ID>@SECONDARY_ATTRIBUTE_LOOP_END|"+
			"@POINT_POOL_LOOP_COUNT|@POINT_POOL_LOOP_START<@ID:@CURRENT/@MAXIMUM>@POINT_POOL_LOOP_END"))
}

// TestLegacyExportLeavesAGroupsOwnFiguresEmpty verifies that the legacy exporter leaves out the quantity, value and
// weight of a group, which has none of its own, while still giving the totals of what it holds.
func TestLegacyExportLeavesAGroupsOwnFiguresEmpty(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	group := gurps.NewEquipmentGroup(e, nil)
	group.Name = "Kit"
	rope := gurps.NewEquipment(e, group, false)
	rope.Name = "Rope"
	rope.BaseValue = "5"
	rope.BaseWeight = "2 lb"
	group.Children = []*gurps.Equipment{rope}
	e.CarriedEquipment = []*gurps.Equipment{group}
	c.Equal("Kit||||2 lb|5|\nRope|1|5|2 lb|2 lb|5|\n", runLegacyExport(t, c, e,
		"@EQUIPMENT_LOOP_START@DESCRIPTION|@QTY|@COST|@WEIGHT|@WEIGHT_SUMMARY|@COST_SUMMARY|\n@EQUIPMENT_LOOP_END"))
}

// TestLegacyExportLeavesAMissingAttributeOrThresholdEmpty verifies that the fixed attribute and pool threshold keys,
// which name attributes and threshold states of the standard set, emit nothing for a sheet whose attribute set lacks
// them rather than an out-of-range sentinel.
func TestLegacyExportLeavesAMissingAttributeOrThresholdEmpty(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	const tmpl = "@ST|@DX|@IQ|@HT|@FP|@BASIC_FP|@HP|@BASIC_HP|@WILL|@FRIGHT_CHECK|@BASIC_SPEED|@BASIC_MOVE|@PERCEPTION|" +
		"@VISION|@HEARING|@TASTE_SMELL|@TOUCH"
	const thresholdTmpl = "@TIRED|@FP_COLLAPSE|@UNCONSCIOUS|@REELING|@HP_COLLAPSE|@DEATH_CHECK_1|@DEATH_CHECK_2|" +
		"@DEATH_CHECK_3|@DEATH_CHECK_4|@DEAD"
	c.Equal("10|10|10|10|10|10|10|10|10|10|5|5|10|10|10|10|10", runLegacyExport(t, c, e, tmpl))
	c.Equal("3|0|-10|3|0|-10|-20|-30|-40|-50", runLegacyExport(t, c, e, thresholdTmpl))
	for _, id := range []string{fpAttrID, gurps.HitPointsID} {
		for _, one := range e.SheetSettings.Attributes.Set[id].Thresholds {
			one.State = "Renamed " + one.State
		}
	}
	c.Equal("|||||||||", runLegacyExport(t, c, e, thresholdTmpl))
	clear(e.SheetSettings.Attributes.Set)
	clear(e.Attributes.Set)
	c.Equal("||||||||||||||||", runLegacyExport(t, c, e, tmpl))
	c.Equal("|||||||||", runLegacyExport(t, c, e, thresholdTmpl))
}

// TestLegacyExportTraitLoopsSelectByClass verifies that each trait loop and its count select the same enabled traits:
// by classifying points for the advantage, perk, quirk and disadvantage loops, by tag for languages and by name prefix,
// ignoring case, for cultural familiarities, all subject to the tag filters.
func TestLegacyExportTraitLoopsSelectByClass(t *testing.T) {
	e := gurps.NewEntity()
	e.Traits = nil
	addTrait := func(name string, points fxp.Int, tags ...string) *gurps.Trait {
		trait := gurps.NewTrait(e, nil, false)
		trait.Name = name
		trait.BasePoints = points
		tags = append(tags, "Listed")
		trait.Tags = tags
		e.Traits = append(e.Traits, trait)
		return trait
	}
	addTrait("Adv", fxp.Two)
	addTrait("Perk", fxp.One)
	addTrait("Zero", 0)
	addTrait("Quirk", -fxp.One)
	addTrait("Disad", -fxp.Two)
	addTrait("Lang", 0, "Language")
	addTrait("Cultural Familiarity (Elves)", 0)
	addTrait("Cultural Familiarity (Dwarves)", fxp.Ten, "Language").SetDisabled(true)
	for _, tc := range []struct {
		loop string
		want string
	}{
		{loop: "ADVANTAGES", want: "7:[Adv][Perk][Zero][Quirk][Disad][Lang][Cultural Familiarity (Elves)]"},
		{loop: "ADVANTAGES_ALL", want: "2:[Adv][Perk]"},
		{loop: "ADVANTAGES_ONLY", want: "1:[Adv]"},
		{loop: "DISADVANTAGES", want: "1:[Disad]"},
		{loop: "DISADVANTAGES_ALL", want: "2:[Quirk][Disad]"},
		{loop: "QUIRKS", want: "1:[Quirk]"},
		{loop: "PERKS", want: "1:[Perk]"},
		{loop: "LANGUAGES", want: "1:[Lang]"},
		{loop: "CULTURAL_FAMILIARITIES", want: "1:[Cultural Familiarity (Elves)]"},
	} {
		t.Run(tc.loop, func(t *testing.T) {
			c := check.New(t)
			tmpl := "@" + tc.loop + "_LOOP_COUNT:@" + tc.loop + "_LOOP_START[@DESCRIPTION_PRIMARY]@" + tc.loop +
				"_LOOP_END"
			c.Equal(tc.want, runLegacyExport(t, c, e, tmpl))
			c.Equal("0:", runLegacyExport(t, c, e, "@EXCLUDE_TAGS_Listed"+tmpl), "a tag filter applies to the loop")
		})
	}
}

// TestLegacyExportTagFilterKeys verifies that the ONLY_TAGS_ and EXCLUDE_TAGS_ keys, and their ONLY_CATEGORIES_ and
// EXCLUDE_CATEGORIES_ aliases, restrict the trait and equipment loops and counts that follow them, that an "only" tag
// wins over an "exclude" tag, and that a trait or equipment loop clears the filters once it has run.
func TestLegacyExportTagFilterKeys(t *testing.T) {
	e := gurps.NewEntity()
	e.Traits = nil
	for _, one := range []struct {
		name string
		tags []string
	}{
		{name: "Strong", tags: []string{"Physical"}},
		{name: "Smart", tags: []string{"Mental"}},
		{name: "Both", tags: []string{"Physical", "Mental"}},
		{name: "Plain"},
	} {
		trait := gurps.NewTrait(e, nil, false)
		trait.Name = one.name
		trait.Tags = one.tags
		e.Traits = append(e.Traits, trait)
	}
	newEquipment := func(name, tag string) *gurps.Equipment {
		eqp := gurps.NewEquipment(e, nil, false)
		eqp.Name = name
		eqp.Tags = []string{tag}
		return eqp
	}
	e.CarriedEquipment = []*gurps.Equipment{newEquipment("Sword", "Weapon"), newEquipment("Rope", "Gear")}
	e.OtherEquipment = []*gurps.Equipment{newEquipment("Gem", "Treasure"), newEquipment("Coin", "Money")}
	const (
		traits  = "@ADVANTAGES_LOOP_COUNT:@ADVANTAGES_LOOP_START[@DESCRIPTION_PRIMARY]@ADVANTAGES_LOOP_END"
		carried = "@EQUIPMENT_LOOP_COUNT:@EQUIPMENT_LOOP_START[@DESCRIPTION_PRIMARY]@EQUIPMENT_LOOP_END"
		other   = "@OTHER_EQUIPMENT_LOOP_COUNT:@OTHER_EQUIPMENT_LOOP_START[@DESCRIPTION_PRIMARY]@OTHER_EQUIPMENT_LOOP_END"
	)
	for _, tc := range []struct {
		name string
		tmpl string
		want string
	}{
		{name: "only tags", tmpl: "@ONLY_TAGS_Physical" + traits, want: "2:[Strong][Both]"},
		{name: "exclude tags", tmpl: "@EXCLUDE_TAGS_Physical" + traits, want: "2:[Smart][Plain]"},
		{name: "only categories", tmpl: "@ONLY_CATEGORIES_Mental" + traits, want: "2:[Smart][Both]"},
		{name: "exclude categories", tmpl: "@EXCLUDE_CATEGORIES_Mental" + traits, want: "2:[Strong][Plain]"},
		{name: "several tags in one key", tmpl: "@EXCLUDE_TAGS_Physical_Mental" + traits, want: "1:[Plain]"},
		{
			name: "only wins over exclude",
			tmpl: "@EXCLUDE_TAGS_Mental@ONLY_TAGS_Physical" + traits,
			want: "2:[Strong][Both]",
		},
		{name: "carried equipment", tmpl: "@ONLY_TAGS_Weapon" + carried, want: "1:[Sword]"},
		{name: "other equipment", tmpl: "@EXCLUDE_TAGS_Treasure" + other, want: "1:[Coin]"},
		{
			name: "trait loop clears the filters",
			tmpl: "@ONLY_TAGS_Physical" + traits + "|@EXCLUDE_TAGS_Physical" + traits + "|" + traits,
			want: "2:[Strong][Both]|2:[Smart][Plain]|4:[Strong][Smart][Both][Plain]",
		},
		{
			name: "equipment loop clears the filters",
			tmpl: "@ONLY_TAGS_Weapon" + carried + "|@EXCLUDE_TAGS_Weapon" + carried + "|" + carried,
			want: "1:[Sword]|1:[Rope]|2:[Sword][Rope]",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			c.Equal(tc.want, runLegacyExport(t, c, e, tc.tmpl))
		})
	}
}

// TestLegacyExportEncumbranceLoopKeys verifies that the encumbrance loop visits every level, giving each its load, move
// and dodge, and marks only the character's current level.
func TestLegacyExportEncumbranceLoopKeys(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	anvil := gurps.NewEquipment(e, nil, false)
	anvil.Name = "Anvil"
	anvil.BaseWeight = "50 lb" // Basic Lift is 20 lb, so this is Medium encumbrance.
	e.CarriedEquipment = append(e.CarriedEquipment, anvil)
	c.Equal("5|"+
		"[|||None (0)|None (0)|0|20 lb|5|8]"+
		"[|||Light (1)|Light (1)|1|40 lb|4|7]"+
		"[current|1|&#8226;|&#8226; Medium (2)|Medium (2)|2|60 lb|3|6]"+
		"[|||Heavy (3)|Heavy (3)|3|120 lb|2|5]"+
		"[|||X-Heavy (4)|X-Heavy (4)|4|200 lb|1|4]",
		runLegacyExport(t, c, e, "@ENCUMBRANCE_LOOP_COUNT|@ENCUMBRANCE_LOOP_START[@CURRENT_MARKER|@CURRENT_MARKER_1|"+
			"@CURRENT_MARKER_BULLET|@LEVEL|@LEVEL_NO_MARKER|@LEVEL_ONLY|@MAX_LOAD|@MOVE|@DODGE]@ENCUMBRANCE_LOOP_END"))
	c.Equal("[|None (0)][|Light (1)][•|• Medium (2)][|Heavy (3)][|X-Heavy (4)]", runLegacyExport(t, c, e,
		"@ENCODING_OFF@ENCUMBRANCE_LOOP_START[@CURRENT_MARKER_BULLET|@LEVEL]@ENCUMBRANCE_LOOP_END"))
}

// TestLegacyExportAmmoKey verifies that the AMMO key gives, for a weapon owned by equipment, the total quantity of the
// carried, really-equipped equipment whose AmmoType tag names the type the owner's UsesAmmoType tag names, matching the
// tag prefixes without regard to case and the types without regard to spaces; 0 when the owner names no type; and
// nothing for a weapon not owned by equipment.
func TestLegacyExportAmmoKey(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	addAmmoTestEquipment := func(parent *gurps.Equipment, name string, qty fxp.Int, tags ...string) *gurps.Equipment {
		eqp := gurps.NewEquipment(e, parent, false)
		eqp.Name = name
		eqp.Quantity = qty
		eqp.Tags = tags
		if parent == nil {
			e.CarriedEquipment = append(e.CarriedEquipment, eqp)
		} else {
			parent.Children = append(parent.Children, eqp)
		}
		return eqp
	}
	bow := addAmmoTestEquipment(nil, "Bow", fxp.One, "Missile", "UsesAmmoType: Broad Head")
	bow.Weapons = []*gurps.Weapon{newTestWeapon(bow, false, "Shoot")}
	sling := addAmmoTestEquipment(nil, "Sling", fxp.One)
	sling.Weapons = []*gurps.Weapon{newTestWeapon(sling, false, "Throw")}
	addAmmoTestEquipment(nil, "Arrows", fxp.Twelve, "AmmoType:BroadHead")
	quiver := gurps.NewEquipment(e, nil, true)
	quiver.Name = "Quiver"
	e.CarriedEquipment = append(e.CarriedEquipment, quiver)
	addAmmoTestEquipment(quiver, "Spare Arrows", fxp.Five, "AMMOTYPE:Broad Head")
	addAmmoTestEquipment(nil, "Stowed Arrows", fxp.Hundred, "AmmoType:BroadHead").Equipped = false
	addAmmoTestEquipment(nil, "Bolts", fxp.Thirty, "AmmoType:Bolt")
	cached := gurps.NewEquipment(e, nil, false)
	cached.Name = "Cached Arrows"
	cached.Quantity = fxp.FiveHundred
	cached.Tags = []string{"AmmoType:BroadHead"}
	e.OtherEquipment = append(e.OtherEquipment, cached)
	spit := gurps.NewTrait(e, nil, false)
	spit.Name = "Spit"
	spit.Weapons = []*gurps.Weapon{newTestWeapon(spit, false, "Spit")}
	e.Traits = append(e.Traits, spit)
	c.Equal("[Bow:17][Sling:0][Spit:]",
		runLegacyExport(t, c, e, "@RANGED_LOOP_START[@DESCRIPTION_PRIMARY:@AMMO]@RANGED_LOOP_END"))
}

// TestLegacyExportParagraphKeys verifies that the formatted user description and note keys write each line of their
// text, encoded, as its own paragraph and write nothing for blank text, while DESCRIPTION_USER writes the encoded text
// without paragraphs.
func TestLegacyExportParagraphKeys(t *testing.T) {
	const text = "Tall & lean\n<b>Scarred</b>"
	for _, tc := range []struct {
		name string
		text string
		tmpl string
		want string
	}{
		{
			name: "user description",
			text: text,
			tmpl: "@ADVANTAGES_LOOP_START[@DESCRIPTION_USER]@ADVANTAGES_LOOP_END",
			want: "[Tall &amp; lean<br>&lt;b&gt;Scarred&lt;/b&gt;]",
		},
		{
			name: "formatted user description",
			text: text,
			tmpl: "@ADVANTAGES_LOOP_START[@DESCRIPTION_USER_FORMATTED]@ADVANTAGES_LOOP_END",
			want: "[<p>Tall &amp; lean</p>\n<p>&lt;b&gt;Scarred&lt;/b&gt;</p>\n]",
		},
		{
			name: "blank formatted user description",
			text: " \n ",
			tmpl: "@ADVANTAGES_LOOP_START[@DESCRIPTION_USER_FORMATTED]@ADVANTAGES_LOOP_END",
			want: "[]",
		},
		{
			name: "formatted note",
			text: text,
			tmpl: "@NOTES_LOOP_START[@NOTE_FORMATTED]@NOTES_LOOP_END",
			want: "[<p>Tall &amp; lean</p>\n<p>&lt;b&gt;Scarred&lt;/b&gt;</p>\n]",
		},
		{
			name: "blank formatted note",
			text: " \n ",
			tmpl: "@NOTES_LOOP_START[@NOTE_FORMATTED]@NOTES_LOOP_END",
			want: "[]",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			e := gurps.NewEntity()
			trait := gurps.NewTrait(e, nil, false)
			trait.UserDesc = tc.text
			e.Traits = []*gurps.Trait{trait}
			note := gurps.NewNote(e, nil, false)
			note.MarkDown = tc.text
			e.Notes = []*gurps.Note{note}
			c.Equal(tc.want, runLegacyExport(t, c, e, tc.tmpl))
		})
	}
}

// TestLegacyExportColorKeys verifies that a COLOR_ key writes the current theme color whose ID is the rest of the key
// in lowercase, and reports a key naming no theme color as unidentified.
func TestLegacyExportColorKeys(t *testing.T) {
	c := check.New(t)
	prev := *colors.Success
	t.Cleanup(func() { *colors.Success = prev })
	// The same color in both modes keeps the result independent of the host's dark mode setting.
	*colors.Success = unison.ThemeColor{Light: unison.RGB(1, 2, 3), Dark: unison.RGB(1, 2, 3)}
	c.Equal("#010203|Unidentified key: &quot;COLOR_NOPE&quot;",
		runLegacyExport(t, c, gurps.NewEntity(), "@COLOR_SUCCESS|@COLOR_NOPE"))
}

// TestLegacyExportReportsUncreatableOutput verifies that the legacy exporter returns the error from creating its output
// file.
func TestLegacyExportReportsUncreatableOutput(t *testing.T) {
	c := check.New(t)
	err := legacyTextExport(gurps.NewEntity(), []byte("@NAME"), filepath.Join(t.TempDir(), "missing", "out.txt"))
	c.True(errors.Is(err, fs.ErrNotExist), "the error is the one creating the file gave")
}
