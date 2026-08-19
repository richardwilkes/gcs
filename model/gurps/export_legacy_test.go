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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/toolbox/v2/check"
)

// runLegacyExport runs the legacy (non-Go-template) exporter over the given template text and returns the resulting
// output.
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
		AttributeDefData: AttributeDefData{
			DefID: "hit_full",
			Type:  attribute.Integer,
			Name:  "Hit Full",
			Base:  "3",
		},
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
	plate := NewEquipment(e, nil, false)
	plate.Name = "Plate Armor"
	plate.Features = Features{
		newTestDRBonus(fxp.Two, AllID, TorsoID),
		newTestDRBonus(fxp.One, AllID, "Torso"), // the same location, with a different case
	}

	// DR that only reaches the skull through an enabled modifier.
	helm := NewEquipment(e, nil, false)
	helm.Name = "Helmet"
	helm.Modifiers = append(helm.Modifiers, newTestDRBonusModifier(e, "Face Guard", newTestDRBonus(fxp.Three, AllID,
		"skull")))

	// A disabled modifier contributes nothing, exactly as it does when features are collected.
	cloak := NewEquipment(e, nil, false)
	cloak.Name = "Cloak"
	disabled := newTestDRBonusModifier(e, "Hood", newTestDRBonus(fxp.Five, AllID, "skull"))
	disabled.Disabled = true
	cloak.Modifiers = append(cloak.Modifiers, disabled)

	// Switchable bonuses, on the item and on one of its modifiers, only count while the item's switch is on.
	cape := NewEquipment(e, nil, false)
	cape.Name = "Cape"
	switchable := newTestDRBonus(fxp.Seven, AllID, "skull")
	switchable.Switchable = true
	cape.Features = Features{switchable}
	switchableOnMod := newTestDRBonus(fxp.Nine, AllID, "skull")
	switchableOnMod.Switchable = true
	cape.Modifiers = append(cape.Modifiers, newTestDRBonusModifier(e, "Lining", switchableOnMod))

	// A "this armor" bonus needs no examination of its own: it covers the locations the item's and its modifiers'
	// located bonuses name, and those are what get scanned, so the modifier's skull bonus is what lists the robe.
	robe := NewEquipment(e, nil, false)
	robe.Name = "Robe"
	robe.Features = Features{newTestDRBonus(fxp.Six, AllID)} // no locations, i.e. "this armor"
	robe.Modifiers = append(robe.Modifiers, newTestDRBonusModifier(e, "Cowl", newTestDRBonus(fxp.One, AllID, "skull")))

	// Neither an unequipped carried item nor an item in the other equipment list contributes DR to the character.
	stowed := NewEquipment(e, nil, false)
	stowed.Name = "Stowed Helm"
	stowed.Equipped = false
	stowed.Features = Features{newTestDRBonus(fxp.Eight, AllID, "skull")}
	spare := NewEquipment(e, nil, false)
	spare.Name = "Spare Helm"
	spare.Features = Features{newTestDRBonus(fxp.Eight, AllID, "skull")}

	e.CarriedEquipment = append(e.CarriedEquipment, plate, helm, cloak, cape, robe, stowed)
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
