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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/toolbox/v2/check"
)

// runLegacyExport runs the legacy (non-Go-template) exporter over the given template text and returns its output.
func runLegacyExport(t *testing.T, c check.Checker, entity *Entity, tmpl string) string {
	t.Helper()
	entity.Recalculate()
	outPath := filepath.Join(t.TempDir(), "out.txt")
	c.NoError(legacyTextExport(entity, []byte(tmpl), outPath))
	data, err := os.ReadFile(outPath)
	c.NoError(err)
	return string(data)
}

// TestLegacyExportPerceptionPoints verifies that @PERCEPTION_POINTS resolves against the "per" attribute ID rather than
// a non-existent "perception" ID, which would always yield 0.
func TestLegacyExportPerceptionPoints(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	e.Attributes.Set["per"].Adjustment = fxp.Two // "per" costs 5/point, so 2 levels == 10 points
	c.Equal("10", runLegacyExport(t, c, e, "@PERCEPTION_POINTS"))
	c.Equal(e.Attributes.Cost("per").String(), runLegacyExport(t, c, e, "@PERCEPTION_POINTS"))
}

// TestLegacyExportOptionalParensEncoding verifies that the DESCRIPTION_NOTES* family encodes its text like every other
// text-emitting key rather than dumping it raw into the output.
func TestLegacyExportOptionalParensEncoding(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	s := NewSkill(e, nil, false)
	s.Name = "Brawling"
	s.LocalNotes = "a<b & \"c\"\nd"
	e.Skills = append(e.Skills, s)
	c.Equal("[ (a&lt;b &amp; &quot;c&quot;<br>d)]",
		runLegacyExport(t, c, e, "@SKILLS_LOOP_START[@DESCRIPTION_NOTES_PAREN]@SKILLS_LOOP_END"))

	// With encoding turned off, the text passes through untouched.
	c.Equal("[ [a<b & \"c\"\nd]]",
		runLegacyExport(t, c, e, "@ENCODING_OFF@SKILLS_LOOP_START[@DESCRIPTION_NOTES_BRACKET]@SKILLS_LOOP_END"))
}

// TestLegacyExportSkillsLoopCount verifies that @SKILLS_LOOP_COUNT counts the same nodes that @SKILLS_LOOP_START
// iterates over, i.e. containers are included in both.
func TestLegacyExportSkillsLoopCount(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	group := NewSkill(e, nil, true)
	group.Name = "Group"
	child := NewSkill(e, group, false)
	child.Name = "Child"
	group.Children = append(group.Children, child)
	e.Skills = append(e.Skills, group)
	c.Equal("2|<GROUP><ITEM>", runLegacyExport(t, c, e, "@SKILLS_LOOP_COUNT|@SKILLS_LOOP_START<@TYPE>@SKILLS_LOOP_END"))
}

// TestLegacyExportConditionalModifierGroups verifies that the legacy exporter writes the reactions and conditional
// modifiers out flat -- the loop counts and IDs cover the members, not the group containers -- and that a member can
// name its group with @GROUP.
func TestLegacyExportConditionalModifierGroups(t *testing.T) {
	withGroupContainersOnSort(t, true)
	c := check.New(t)
	e := NewEntity()
	addReaction(e, "A", "from everyone", "", fxp.One)
	addReaction(e, "B", "from foes", "Combat", fxp.Two)
	addReaction(e, "C", "from allies", "Combat", fxp.Three)
	addGroupedConditionalModifier(e, "D", "to hit", "Combat", fxp.One)
	addGroupedConditionalModifier(e, "E", "to dodge", "", fxp.Two)
	c.Equal("3|<0:from allies|Combat|+3><1:from foes|Combat|+2><2:from everyone||+1>",
		runLegacyExport(t, c, e, "@REACTION_LOOP_COUNT|@REACTION_LOOP_START<@ID:@SITUATION|@GROUP|@MODIFIER>@REACTION_LOOP_END"))
	c.Equal("2|<0:to hit|Combat|+1><1:to dodge||+2>",
		runLegacyExport(t, c, e, "@CONDITIONAL_MODIFIERS_LOOP_COUNT|@CONDITIONAL_MODIFIERS_LOOP_START<@ID:@SITUATION|@GROUP|@MODIFIER>@CONDITIONAL_MODIFIERS_LOOP_END"))
}

// TestLegacyExportLoopBodyKeyDetection verifies that the decision to treat '@' as the start of a key inside a loop body
// is made from the loop body itself, not from a fixed byte in the outer template.
func TestLegacyExportLoopBodyKeyDetection(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	n := NewNote(e, nil, false)
	n.MarkDown = "Note1"
	e.Notes = append(e.Notes, n)

	// "@7" is literal text at the top level, so it must be literal text inside a loop body too.
	c.Equal("[@7]", runLegacyExport(t, c, e, "@NOTES_LOOP_START[@7]@NOTES_LOOP_END"))

	// A digit following the loop-end marker must not turn every key in the loop body into literal text.
	c.Equal("[Note1]7", runLegacyExport(t, c, e, "@NOTES_LOOP_START[@NOTE]@NOTES_LOOP_END7"))
}

// TestLegacyExportNotesSeparator verifies that the separator between top-level notes is encoded along with the notes
// themselves, so that HTML exports keep the blank line between them.
func TestLegacyExportNotesSeparator(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	first := NewNote(e, nil, false)
	first.MarkDown = "First"
	second := NewNote(e, nil, false)
	second.MarkDown = "Second"
	e.Notes = append(e.Notes, first, second)
	c.Equal("First<br><br>Second", runLegacyExport(t, c, e, "@NOTES"))
	c.Equal("First\n\nSecond", runLegacyExport(t, c, e, "@ENCODING_OFF@NOTES"))
}

// TestLegacyExportOtherEquipmentLoopMarker verifies that the other-equipment loop terminates on its own loop-end marker
// rather than on the carried-equipment one, which is a substring of it.
func TestLegacyExportOtherEquipmentLoopMarker(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	eqp := NewEquipment(e, nil, false)
	eqp.Name = "Rock"
	e.OtherEquipment = append(e.OtherEquipment, eqp)
	c.Equal("|EQUIPMENT_LOOP_END|Rock|", runLegacyExport(t, c, e,
		"@OTHER_EQUIPMENT_LOOP_START|EQUIPMENT_LOOP_END|@DESCRIPTION|@OTHER_EQUIPMENT_LOOP_END"))
}

// TestLegacyExportAttributeNameSuffixes verifies that the more specific attribute key suffixes are matched before the
// less specific ones, and that a suffix match that doesn't resolve to an attribute falls through to the next candidate.
func TestLegacyExportAttributeNameSuffixes(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	c.Equal("ST|Strength|Strength (ST)|10", runLegacyExport(t, c, e,
		"@ST_NAME|@ST_FULL_NAME|@ST_COMBINED_NAME|@ST_CURRENT"))

	// An attribute whose own ID ends with "_full" must still resolve via the "_name" suffix.
	e.SheetSettings.Attributes.Set["hit_full"] = &AttributeDef{
		DefID: "hit_full",
		Type:  attribute.Integer,
		Name:  "Hit Full",
		Base:  "3",
		Order: len(e.SheetSettings.Attributes.Set),
	}
	e.Attributes.Set["hit_full"] = NewAttribute(e, "hit_full", len(e.Attributes.Set))
	c.Equal("Hit Full|3", runLegacyExport(t, c, e, "@HIT_FULL_NAME|@HIT_FULL_CURRENT"))
}

// TestLegacyExportModifierNotesLineBreaks verifies that a trait's modifier notes reach the legacy exporter as plain
// newline-separated text, so writeEncodedText turns the separator into a real <br> rather than escaping markup that
// the model had already embedded.
func TestLegacyExportModifierNotesLineBreaks(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	trait := NewTrait(e, nil, false)
	trait.Name = "Greed"
	trait.SelfControl = selfctrl.CR12
	mod := NewTraitModifier(e, nil, false)
	mod.Name = "Mitigator"
	trait.Modifiers = append(trait.Modifiers, mod)
	e.Traits = append(e.Traits, trait)
	out := runLegacyExport(t, c, e, "@ADVANTAGES_LOOP_START[@DESCRIPTION_MODIFIER_NOTES_BRACKET]@ADVANTAGES_LOOP_END")
	c.Contains(out, "Self-Control Roll (CR): 12 or less (Resist quite often)<br>Mitigator")
	c.Equal(0, strings.Count(out, "&lt;br&gt;"), "no escaped line-break markup is emitted")
}

// TestLegacyExportHierarchicalWeaponLoopCount verifies that the hierarchical loop-count keys report the number of rows
// their loop actually produces -- one per distinct weapon -- rather than the total number of attack modes that the
// flat loop-count keys report.
func TestLegacyExportHierarchicalWeaponLoopCount(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	// "Karate" contributes two melee attack modes and two ranged ones; "Innate Attack" contributes one of each. The
	// hierarchical loops collapse each trait's modes into a single row. A new entity already carries the three
	// unarmed melee modes of "Natural Attacks", so melee has 6 modes across 3 weapons and ranged has 3 across 2.
	karate := NewTrait(e, nil, false)
	karate.Name = "Karate"
	karate.Weapons = append(karate.Weapons, newTestWeapon(karate, true, "Punch"), newTestWeapon(karate, true, "Kick"),
		newTestWeapon(karate, false, "Spit"), newTestWeapon(karate, false, "Sneeze"))
	innate := NewTrait(e, nil, false)
	innate.Name = "Innate Attack"
	innate.Weapons = append(innate.Weapons, newTestWeapon(innate, true, "Slam"),
		newTestWeapon(innate, false, "Bolt"))
	e.Traits = append(e.Traits, karate, innate)

	c.Equal("6|3|(Innate Attack:1)(Karate:2)(Natural Attacks:3)",
		runLegacyExport(t, c, e, "@MELEE_LOOP_COUNT|@HIERARCHICAL_MELEE_LOOP_COUNT|"+
			"@HIERARCHICAL_MELEE_LOOP_START(@DESCRIPTION_PRIMARY:@ATTACK_MODES_LOOP_COUNT)@HIERARCHICAL_MELEE_LOOP_END"))
	c.Equal("3|2|(Innate Attack:1)(Karate:2)",
		runLegacyExport(t, c, e, "@RANGED_LOOP_COUNT|@HIERARCHICAL_RANGED_LOOP_COUNT|"+
			"@HIERARCHICAL_RANGED_LOOP_START(@DESCRIPTION_PRIMARY:@ATTACK_MODES_LOOP_COUNT)@HIERARCHICAL_RANGED_LOOP_END"))
}

func newTestWeapon(owner WeaponOwner, melee bool, usage string) *Weapon {
	w := NewWeapon(owner, melee)
	w.Usage = usage
	return w
}

// TestLegacyExportWeaponLoops verifies that the flat and hierarchical weapon loops behave the same way for melee and
// ranged weapons: the flat loops visit every attack mode as its own weapon and have no attack modes of their own, while
// the hierarchical loops visit each distinct weapon once and run the ATTACK_MODES body for each of its modes using the
// same melee or ranged key set as the enclosing loop.
func TestLegacyExportWeaponLoops(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	karate := NewTrait(e, nil, false)
	karate.Name = "Karate"
	karate.Weapons = append(karate.Weapons, newTestWeapon(karate, true, "Punch"), newTestWeapon(karate, true, "Kick"),
		newTestWeapon(karate, false, "Spit"), newTestWeapon(karate, false, "Sneeze"))
	innate := NewTrait(e, nil, false)
	innate.Name = "Innate Attack"
	innate.Weapons = append(innate.Weapons, newTestWeapon(innate, false, "Bolt"))
	e.Traits = append(e.Traits, karate, innate)
	reach := karate.Weapons[0].Reach.Resolve(karate.Weapons[0], nil).String()
	rof := karate.Weapons[2].RateOfFire.Resolve(karate.Weapons[2], nil).String()

	// Flat loops number every attack mode in sorted order and reject the attack modes keys.
	c.Equal("(0:Bolt)(1:Sneeze)(2:Spit)", runLegacyExport(t, c, e, "@RANGED_LOOP_START(@ID:@USAGE)@RANGED_LOOP_END"))
	c.Equal(strings.Repeat("(0:Unidentified key: &quot;ATTACK_MODES_LOOP_START&quot;)", 3),
		runLegacyExport(t, c, e, "@RANGED_LOOP_START(@ATTACK_MODES_LOOP_COUNT:@ATTACK_MODES_LOOP_START)@RANGED_LOOP_END"))

	// Hierarchical loops number the modes within each weapon, and the modes of a ranged weapon see the ranged keys
	// (ROF) but not the melee ones (REACH), and vice versa.
	c.Equal("[Innate Attack:<0:Bolt:"+rof+":Unidentified key: &quot;REACH&quot;>]"+
		"[Karate:<0:Sneeze:"+rof+":Unidentified key: &quot;REACH&quot;><1:Spit:"+rof+":Unidentified key: &quot;REACH&quot;>]",
		runLegacyExport(t, c, e, "@HIERARCHICAL_RANGED_LOOP_START[@DESCRIPTION_PRIMARY:"+
			"@ATTACK_MODES_LOOP_START<@ID:@USAGE:@ROF:@REACH>@ATTACK_MODES_LOOP_END]@HIERARCHICAL_RANGED_LOOP_END"))
	c.Contains(runLegacyExport(t, c, e, "@HIERARCHICAL_MELEE_LOOP_START[@DESCRIPTION_PRIMARY:"+
		"@ATTACK_MODES_LOOP_START<@ID:@USAGE:@REACH:@ROF>@ATTACK_MODES_LOOP_END]@HIERARCHICAL_MELEE_LOOP_END"),
		"[Karate:<0:Kick:"+reach+":Unidentified key: &quot;ROF&quot;><1:Punch:"+reach+":Unidentified key: &quot;ROF&quot;>]")
}

// TestLegacyExportKeyScanner verifies that the top-level template and loop bodies are scanned by the same rules -- a
// key ends at the first byte outside [A-Za-z0-9_], which is read again as text unless enhanced key parsing is on and it
// is a closing '@' -- and that a key running up to the very end of the template is still emitted.
func TestLegacyExportKeyScanner(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	n := NewNote(e, nil, false)
	n.MarkDown = "Note1"
	e.Notes = append(e.Notes, n)
	c.Equal("10 x10", runLegacyExport(t, c, e, "@ST x@DX"))
	c.Equal("\n10xNote1yz|10",
		runLegacyExport(t, c, e, "@ENHANCED_KEY_PARSING\n@ST@x@NOTES_LOOP_START@@NOTE@y@NOTES_LOOP_END@z|@DX"))
}

// TestLegacyExportHitLocationEquipmentFromModifier verifies that armor whose DR for a location comes from one of its
// modifiers is listed for that location, since the DR printed for the location already includes the modifier's
// contribution.
func TestLegacyExportHitLocationEquipmentFromModifier(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	helm := NewEquipment(e, nil, false)
	helm.Name = "Helmet"
	helm.Modifiers = append(helm.Modifiers, newTestDRBonusModifier(e, "Face Guard", newTestDRBonus(fxp.Three, AllID,
		"skull")))
	e.CarriedEquipment = append(e.CarriedEquipment, helm)
	e.Recalculate()

	ex := &legacyExporter{entity: e}
	skull := e.SheetSettings.BodyType.LookupLocationByID(e, "skull")
	c.NotNil(skull, "the default body has a skull location")
	c.Equal("5", skull.DisplayDR(e, nil), "the modifier's DR reaches the location, on top of the skull's own 2")
	c.Equal([]string{"Helmet"}, ex.hitLocationEquipment(skull), "the armor providing that DR is listed")

	torso := e.SheetSettings.BodyType.LookupLocationByID(e, TorsoID)
	c.NotNil(torso, "the default body has a torso location")
	c.Equal("0", torso.DisplayDR(e, nil), "no DR reaches a location the modifier doesn't name")
	c.Equal(0, len(ex.hitLocationEquipment(torso)), "and nothing is listed for it")
}

// TestLegacyExportHitLocationEquipment verifies which pieces of equipment are listed as providing DR to a hit location:
// only carried, really-equipped ones, counting both their own DR bonuses and those of their enabled, non-container
// modifiers, with switchable bonuses honoring the owning item's switch, and each item named at most once no matter how
// many of its bonuses reach the location.
func TestLegacyExportHitLocationEquipment(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	// Two DR bonuses that both reach the torso; the item must still only be listed once.
	addCarriedEquipmentWithFeatures(e, "Plate Armor",
		newTestDRBonus(fxp.Two, AllID, TorsoID),
		newTestDRBonus(fxp.One, AllID, "Torso"), // the same location, with a different case
	)

	// DR that only reaches the skull through an enabled modifier.
	helm := addCarriedEquipmentWithFeatures(e, "Helmet")
	helm.Modifiers = append(helm.Modifiers, newTestDRBonusModifier(e, "Face Guard", newTestDRBonus(fxp.Three, AllID,
		"skull")))

	// A disabled modifier contributes nothing, exactly as it does when features are collected.
	cloak := addCarriedEquipmentWithFeatures(e, "Cloak")
	disabled := newTestDRBonusModifier(e, "Hood", newTestDRBonus(fxp.Five, AllID, "skull"))
	disabled.Disabled = true
	cloak.Modifiers = append(cloak.Modifiers, disabled)

	// Switchable bonuses, on the item and on one of its modifiers, only count while the item's switch is on.
	switchable := newTestDRBonus(fxp.Seven, AllID, "skull")
	switchable.Switchable = true
	cape := addCarriedEquipmentWithFeatures(e, "Cape", switchable)
	switchableOnMod := newTestDRBonus(fxp.Nine, AllID, "skull")
	switchableOnMod.Switchable = true
	cape.Modifiers = append(cape.Modifiers, newTestDRBonusModifier(e, "Lining", switchableOnMod))

	// A "this armor" bonus needs no examination of its own: it covers the locations the item's and its modifiers'
	// located bonuses name, and those are what get scanned, so the modifier's skull bonus is what lists the robe.
	robe := addCarriedEquipmentWithFeatures(e, "Robe", newTestDRBonus(fxp.Six, AllID)) // no locations, i.e. "this armor"
	robe.Modifiers = append(robe.Modifiers, newTestDRBonusModifier(e, "Cowl", newTestDRBonus(fxp.One, AllID, "skull")))

	// Neither an unequipped carried item nor an item in the other equipment list contributes DR to the character.
	addCarriedEquipmentWithFeatures(e, "Stowed Helm", newTestDRBonus(fxp.Eight, AllID, "skull")).Equipped = false
	spare := NewEquipment(e, nil, false)
	spare.Name = "Spare Helm"
	spare.Features = Features{newTestDRBonus(fxp.Eight, AllID, "skull")}
	e.OtherEquipment = append(e.OtherEquipment, spare)
	e.Recalculate()

	ex := &legacyExporter{entity: e}
	skull := e.SheetSettings.BodyType.LookupLocationByID(e, "skull")
	c.NotNil(skull, "the default body has a skull location")
	c.Equal([]string{"Helmet", "Robe"}, ex.hitLocationEquipment(skull),
		"only the armor actually granting DR to the skull is listed")

	torso := e.SheetSettings.BodyType.LookupLocationByID(e, TorsoID)
	c.NotNil(torso, "the default body has a torso location")
	c.Equal([]string{"Plate Armor"}, ex.hitLocationEquipment(torso),
		"an item with two DR bonuses reaching the same location is listed once")

	// Throwing the cape's switch brings both of its switchable bonuses into play.
	cape.SwitchedOn = true
	c.Equal([]string{"Helmet", "Cape", "Robe"}, ex.hitLocationEquipment(skull),
		"switching the cape on adds it to the skull's list")

	// The same list reaches the template through the @EQUIPMENT key of the hit location loop.
	c.Contains(runLegacyExport(t, c, e, "@HIT_LOCATION_LOOP_START[@WHERE:@EQUIPMENT]@HIT_LOCATION_LOOP_END"),
		"["+skull.TableName+":Helmet, Cape, Robe]")
}

func newTestDRBonusModifier(owner DataOwner, name string, bonus *DRBonus) *EquipmentModifier {
	mod := NewEquipmentModifier(owner, nil, false)
	mod.Name = name
	mod.Features = Features{bonus}
	return mod
}

// TestLegacyExportSharedNodeKeys verifies that the keys shared by the trait, skill, spell, equipment and note loops
// behave identically across those loops while each loop's deliberate differences are preserved.
func TestLegacyExportSharedNodeKeys(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	group := NewTrait(e, nil, true)
	group.Name = "Group"
	group.ContainerType = container.AlternativeAbilities
	group.PageRef = "B1"
	trait := newTraitNeedingMissingTrait(e, "Greed")
	trait.SetParent(group)
	group.Children = append(group.Children, trait)
	mod := NewTraitModifier(e, nil, false)
	mod.Name = "Mitigator"
	mod.LocalNotes = "mod note"
	trait.Modifiers = append(trait.Modifiers, mod)
	e.Traits = []*Trait{group}

	skill := NewSkill(e, nil, false)
	skill.Name = "Brawling"
	skill.PageRef = "B2"
	e.Skills = append(e.Skills, skill)

	spell := NewSpell(e, nil, false)
	spell.Name = "Fireball"
	spell.LocalNotes = "spell note"
	e.Spells = append(e.Spells, spell)

	eqp := NewEquipment(e, nil, false)
	eqp.Name = "Rock"
	eqpMod := NewEquipmentModifier(e, nil, false)
	eqpMod.Name = "Sharp"
	eqpMod.LocalNotes = "eqp mod note"
	eqp.Modifiers = append(eqp.Modifiers, eqpMod)
	e.CarriedEquipment = append(e.CarriedEquipment, eqp)

	noteGroup := NewNote(e, nil, true)
	noteGroup.MarkDown = "Notes"
	note := NewNote(e, noteGroup, false)
	note.MarkDown = "Note1"
	note.PageRef = "B3"
	noteGroup.Children = append(noteGroup.Children, note)
	e.Notes = append(e.Notes, noteGroup)

	// Traits: TYPE emits the container type, PARENT_ID is empty for top-level nodes, the prereq failure shows up in
	// both SATISFIED and STYLE_INDENT_WARNING, and MODIFIER_NOTES_FOR_ resolves the active modifier.
	c.Equal("ALTERNATIVE_ABILITIES|"+string(group.ID())+"||B1|Y||0|\n"+
		"ITEM|"+string(trait.ID())+"|"+string(group.ID())+"||N| style=\"padding-left: 12px;color: red;\" |3|mod note\n",
		runLegacyExport(t, c, e, "@ADVANTAGES_LOOP_START@TYPE|@ID|@PARENT_ID|@REF|@SATISFIED|"+
			"@STYLE_INDENT_WARNING|@DEPTHx3|@MODIFIER_NOTES_FOR_Mitigator\n@ADVANTAGES_LOOP_END"))

	// Skills: the same keys use GROUP for containers, but skills have no MODIFIER_NOTES_FOR_ key.
	c.Equal("ITEM|"+string(skill.ID())+"|B2|Y|Brawling|Unidentified key: &quot;MODIFIER_NOTES_FOR_X&quot;",
		runLegacyExport(t, c, e, "@SKILLS_LOOP_START@TYPE|@ID|@REF|@SATISFIED|@DESCRIPTION_PRIMARY|"+
			"@MODIFIER_NOTES_FOR_X@SKILLS_LOOP_END"))

	// Spells: DESCRIPTION writes the notes and rituals as separate notes, DESCRIPTION_NOTES merges them, and
	// DESCRIPTION_MODIFIER_NOTES is silent rather than unidentified.
	rituals := spell.Rituals()
	c.NotEqual("", rituals)
	c.Equal("Fireball<div class=\"note\">spell note</div><div class=\"note\">"+rituals+"</div>| (spell note; "+
		rituals+")||Fireball",
		runLegacyExport(t, c, e, "@SPELLS_LOOP_START@DESCRIPTION|@DESCRIPTION_NOTES_PAREN|"+
			"@DESCRIPTION_MODIFIER_NOTES_PAREN|@DESCRIPTION_PRIMARY@SPELLS_LOOP_END"))

	// Equipment: MODIFIER_NOTES_FOR_ resolves the active modifier and DESCRIPTION includes the modifier notes.
	c.Equal("eqp mod note|Rock<div class=\"note\">Sharp (eqp mod note)</div>",
		runLegacyExport(t, c, e, "@EQUIPMENT_LOOP_START@MODIFIER_NOTES_FOR_Sharp|@DESCRIPTION@EQUIPMENT_LOOP_END"))

	// Notes: the identity and depth keys work, but there are no description or prerequisite keys and the indent
	// warning is never red.
	c.Equal("GROUP|"+string(noteGroup.ID())+"||||0\n"+
		"ITEM|"+string(note.ID())+"|"+string(noteGroup.ID())+"|B3| style=\"padding-left: 12px;\" |2\n",
		runLegacyExport(t, c, e, "@NOTES_LOOP_START@TYPE|@ID|@PARENT_ID|@REF|@STYLE_INDENT_WARNING|@DEPTHx2\n"+
			"@NOTES_LOOP_END"))
	c.Equal(strings.Repeat("Unidentified key: &quot;DESCRIPTION&quot;|Unidentified key: &quot;SATISFIED&quot;|"+
		"Unidentified key: &quot;DESCRIPTION_NOTES&quot;\n", 2),
		runLegacyExport(t, c, e, "@NOTES_LOOP_START@DESCRIPTION|@SATISFIED|@DESCRIPTION_NOTES\n@NOTES_LOOP_END"))
}

// TestLegacyExportAttributeLoops verifies that the primary, secondary and point pool loops (and their counts) classify
// attributes by their resolved kind: an explicit placement overrides the base-derived classification, and a hidden pool
// is omitted just as it is from the Go-template export.
func TestLegacyExportAttributeLoops(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addDef := func(def *AttributeDef) {
		def.Order = len(e.SheetSettings.Attributes.Set)
		e.SheetSettings.Attributes.Set[def.DefID] = def
		e.Attributes.Set[def.DefID] = NewAttribute(e, def.DefID, len(e.Attributes.Set))
	}
	addDef(&AttributeDef{DefID: "forced", Type: attribute.Integer, Name: "Forced", Base: "$iq", Placement: attribute.Primary})
	addDef(&AttributeDef{DefID: "mana", Type: attribute.Pool, Name: "Mana", Base: "10", Placement: attribute.Hidden})
	c.Equal("5|<st=10><dx=10><iq=10><ht=10><forced=10>|"+
		"9|<will><fright_check><per><vision><hearing><taste_smell><touch><basic_speed><basic_move>|"+
		"2|<fp:10/10><hp:10/10>",
		runLegacyExport(t, c, e, "@PRIMARY_ATTRIBUTE_LOOP_COUNT|"+
			"@PRIMARY_ATTRIBUTE_LOOP_START<@ID=@VALUE>@PRIMARY_ATTRIBUTE_LOOP_END|"+
			"@SECONDARY_ATTRIBUTE_LOOP_COUNT|@SECONDARY_ATTRIBUTE_LOOP_START<@ID>@SECONDARY_ATTRIBUTE_LOOP_END|"+
			"@POINT_POOL_LOOP_COUNT|@POINT_POOL_LOOP_START<@ID:@CURRENT/@MAXIMUM>@POINT_POOL_LOOP_END"))
}
