// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurpstest

import (
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/attack"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellmatch"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stdmg"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/tid"
)

// LargeEntitySizeEnvVar names the environment variable LargeEntitySizeFromEnv reads: a multiplier applied to
// DefaultLargeEntitySize, so "0.25" builds a quarter-sized character and "2" one twice the size.
const LargeEntitySizeEnvVar = "GCS_BENCH_SIZE"

// LargeEntitySize is how many rows NewLargeEntity puts into each of a character sheet's lists. The counts are of the
// rows themselves; the containers they are grouped into come on top. Two lists hold rows beyond their count: the
// skills always include the fixed skills the weapons and ritual magic spells default to and the three with nameable
// markers, so a Skills count below their number still yields them all, and the traits also get the natural attacks
// trait, three traits with unmet prerequisites and, when a list that would own weapons or carry bonuses is empty, a
// trait to hold them. Reactions and ConditionalModifiers are how many of each bonus the traits carry, which is also
// the number of rows the two derived lists show; they go one to a trait until the traits run out, then around again.
// The weapons are spread over the equipment, traits, skills and spells, so they are not extra rows of their own.
// PerContainer is how many rows a container holds before the next one is started.
type LargeEntitySize struct {
	Traits               int
	Skills               int
	Spells               int
	RitualMagicSpells    int
	CarriedEquipment     int
	OtherEquipment       int
	Notes                int
	Reactions            int
	ConditionalModifiers int
	MeleeWeapons         int
	RangedWeapons        int
	PerContainer         int
}

// DefaultLargeEntitySize is the size the benchmarks use unless LargeEntitySizeEnvVar says otherwise. It is a little
// larger than the biggest real sheets seen so far, which had around 100 skills, 550 spells and 45 attacks.
var DefaultLargeEntitySize = LargeEntitySize{
	Traits:               200,
	Skills:               150,
	Spells:               500,
	RitualMagicSpells:    100,
	CarriedEquipment:     300,
	OtherEquipment:       150,
	Notes:                60,
	Reactions:            40,
	ConditionalModifiers: 40,
	MeleeWeapons:         60,
	RangedWeapons:        50,
	PerContainer:         8,
}

// Scaled returns a copy of the size with every count multiplied by the factor and rounded, but never below one.
// PerContainer is kept as it is.
func (s *LargeEntitySize) Scaled(factor float64) LargeEntitySize {
	scale := func(n int) int { return max(1, int(math.Round(float64(n)*factor))) }
	return LargeEntitySize{
		Traits:               scale(s.Traits),
		Skills:               scale(s.Skills),
		Spells:               scale(s.Spells),
		RitualMagicSpells:    scale(s.RitualMagicSpells),
		CarriedEquipment:     scale(s.CarriedEquipment),
		OtherEquipment:       scale(s.OtherEquipment),
		Notes:                scale(s.Notes),
		Reactions:            scale(s.Reactions),
		ConditionalModifiers: scale(s.ConditionalModifiers),
		MeleeWeapons:         scale(s.MeleeWeapons),
		RangedWeapons:        scale(s.RangedWeapons),
		PerContainer:         s.PerContainer,
	}
}

// LargeEntitySizeFromEnv returns DefaultLargeEntitySize scaled by the multiplier in LargeEntitySizeEnvVar, or the
// default itself when the variable is unset. A value that isn't a positive number yields the default and an error.
func LargeEntitySizeFromEnv() (LargeEntitySize, error) {
	v := os.Getenv(LargeEntitySizeEnvVar)
	if v == "" {
		return DefaultLargeEntitySize, nil
	}
	factor, err := strconv.ParseFloat(v, 64)
	if err != nil || factor <= 0 || math.IsInf(factor, 0) || math.IsNaN(factor) {
		return DefaultLargeEntitySize, fmt.Errorf("%s must be a positive number, not %q", LargeEntitySizeEnvVar, v)
	}
	return DefaultLargeEntitySize.Scaled(factor), nil
}

// The names and usages the generated rows share.
const (
	ritualMagicSkillName = "Ritual Magic"
	usageThrust          = "Thrust"
	usageShoot           = "Shoot"
	damageImpaling       = "imp"
)

// largeEntityColleges are the spell colleges a large entity's spells and ritual magic skills are spread over.
var largeEntityColleges = []string{
	"Air", "Animal", "Body Control", "Communication & Empathy", "Earth", "Fire", "Healing", "Illusion & Creation",
	"Light & Darkness", "Water",
}

// largeEntitySkill describes one of the skills a large entity always has, since its weapons default to them.
type largeEntitySkill struct {
	name           string
	specialization string
	attribute      string
	difficulty     difficulty.Level
	points         int
}

// largeEntityWeaponSkills are the skills the weapons default to. Their order matters: the weapon specs below refer to
// them by index.
var largeEntityWeaponSkills = []largeEntitySkill{
	{name: "Broadsword", attribute: gurps.DexterityID, difficulty: difficulty.Average, points: 8},
	{name: "Knife", attribute: gurps.DexterityID, difficulty: difficulty.Easy, points: 2},
	{name: "Spear", attribute: gurps.DexterityID, difficulty: difficulty.Average, points: 4},
	{name: "Brawling", attribute: gurps.DexterityID, difficulty: difficulty.Easy, points: 4},
	{name: "Karate", attribute: gurps.DexterityID, difficulty: difficulty.Hard, points: 8},
	{name: "Bow", attribute: gurps.DexterityID, difficulty: difficulty.Average, points: 8},
	{name: "Crossbow", attribute: gurps.DexterityID, difficulty: difficulty.Easy, points: 2},
	{name: "Guns", specialization: "Pistol", attribute: gurps.DexterityID, difficulty: difficulty.Easy, points: 4},
	{name: "Guns", specialization: "Rifle", attribute: gurps.DexterityID, difficulty: difficulty.Easy, points: 4},
	{name: "Shield", attribute: gurps.DexterityID, difficulty: difficulty.Easy, points: 2},
	{name: "Thrown Weapon", specialization: "Knife", attribute: gurps.DexterityID, difficulty: difficulty.Easy, points: 2},
	{name: "Innate Attack", specialization: "Projectile", attribute: gurps.DexterityID, difficulty: difficulty.Easy, points: 4},
	{name: "Thaumatology", attribute: gurps.IntelligenceID, difficulty: difficulty.VeryHard, points: 4},
	{name: ritualMagicSkillName, attribute: gurps.IntelligenceID, difficulty: difficulty.VeryHard, points: 2},
}

// largeEntityMeleeSpec describes one kind of melee weapon. skill and alt index largeEntityWeaponSkills.
type largeEntityMeleeSpec struct {
	usage        string
	strength     string
	strengthType stdmg.Option
	base         string
	damageType   string
	reach        string
	parry        string
	block        string
	skill        int
	alt          int
}

var largeEntityMeleeSpecs = []largeEntityMeleeSpec{
	{"Swing", "10", stdmg.Swing, "+1", "cut", "1", "0", "No", 0, 1},
	{usageThrust, "10", stdmg.Thrust, "+2", damageImpaling, "1", "0", "No", 0, 2},
	{usageThrust, "6", stdmg.Thrust, "", damageImpaling, "C,1", "-1", "No", 1, 3},
	{usageThrust, "9†", stdmg.Thrust, "+2", damageImpaling, "1,2*", "0U", "No", 2, 0},
	{"Punch", "", stdmg.Thrust, "", "cr", "C", "0", "No", 3, 4},
	{"Kick", "", stdmg.Thrust, "+1", "cr", "C,1", "No", "No", 4, 3},
	{"Bash", "10", stdmg.Thrust, "", "cr", "1", "No", "0", 9, 3},
	{"Swing", "11", stdmg.Swing, "+2", "cut", "1,2", "0", "No", 0, 2},
}

// largeEntityRangedSpec describes one kind of ranged weapon. skill indexes largeEntityWeaponSkills.
type largeEntityRangedSpec struct {
	usage      string
	strength   string
	base       string
	damageType string
	accuracy   string
	rng        string
	rof        string
	shots      string
	bulk       string
	recoil     string
	skill      int
	muscle     bool
}

var largeEntityRangedSpecs = []largeEntityRangedSpec{
	{usageShoot, "10†", "", damageImpaling, "3", "x15/x20", "1", "1(2)", "-7", "", 5, true},
	{usageShoot, "7†", "+4", damageImpaling, "4", "x20/x25", "1", "1(4)", "-6", "", 6, true},
	{usageShoot, "9", "2d+2", "pi", "2", "150/1850", "3", "15+1(3)", "-2", "2", 7, false},
	{usageShoot, "10†", "7d", "pi", "5", "1000/4200", "10", "30+1(3)", "-5*", "2", 8, false},
	{"Throw", "6", "-1", damageImpaling, "0", "x0.8/x1.5", "1", "T(1)", "-2", "", 10, true},
	{"Fire", "", "2d", "burn", "3", "50/100", "1", "", "", "1", 11, false},
}

// largeEntityBuilder puts a large entity together one list at a time, remembering the rows that can own weapons so
// that the weapons can be handed out last, once every list exists.
type largeEntityBuilder struct {
	entity     *gurps.Entity
	size       LargeEntitySize
	rows       int
	equipped   []*gurps.Equipment
	unequipped []*gurps.Equipment
	other      []*gurps.Equipment
	traits     []*gurps.Trait
	skills     []*gurps.Skill
	spells     []*gurps.Spell
}

// NewLargeEntity returns a recalculated character whose sheet has the number of rows the size gives in every list
// (see LargeEntitySize for the rows that come on top), with every container open. The character is deterministic
// apart from its IDs and timestamps: the same size always yields the same names, points, features and weapons. It is
// meant for benchmarks and for tests that need every kind of row at once, so each
// list exercises the costly paths a real sheet does: the skills include techniques, skills with defaults and skills
// whose names carry nameable markers (every college has a "Ritual Magic (@Specialty@)" skill the ritual magic spells
// resolve through); the traits carry modifiers and the features behind the reactions, conditional modifiers and
// weapon bonuses; the equipment is nested three deep with modifiers, some of it unequipped; the weapons default to
// the weapon skills and sit on equipment, traits, skills and spells; every few rows carry page references and local
// notes; and every fourth trait, piece of equipment and note defines a nameable its text uses, some answered and some
// not, which the traits' and equipment's modifiers share or are linked to, beside a marker that defines itself.
func NewLargeEntity(size *LargeEntitySize) *gurps.Entity {
	b := &largeEntityBuilder{entity: gurps.NewEntity(), size: *size}
	b.size.PerContainer = max(1, b.size.PerContainer)
	b.entity.Profile.Name = "Large Benchmark Character"
	// A new entity's profile is filled in at random, so it is set by hand.
	b.entity.Profile.Gender = "Male"
	b.entity.Profile.Age = "30"
	b.entity.Profile.Birthday = "January 1"
	b.entity.Profile.Eyes = "Brown"
	b.entity.Profile.Hair = "Black"
	b.entity.Profile.Skin = "Tan"
	b.entity.Profile.Handedness = "Right"
	b.entity.Profile.Height = fxp.LengthFromInteger(70, fxp.Inch)
	b.entity.Profile.Weight = fxp.WeightFromInteger(160, fxp.Pound)
	b.entity.SheetSettings.ShowSpellAdj = true
	b.setAttribute(gurps.StrengthID, 12)
	b.setAttribute(gurps.DexterityID, 13)
	b.setAttribute(gurps.IntelligenceID, 14)
	b.setAttribute(gurps.HealthID, 11)
	b.addTraits()
	b.addSkills()
	b.addSpells()
	b.addEquipment(b.size.CarriedEquipment, true)
	b.addEquipment(b.size.OtherEquipment, false)
	b.addNotes()
	b.addWeapons()
	b.entity.Recalculate()
	return b.entity
}

func (b *largeEntityBuilder) setAttribute(id string, value int) {
	if attr, ok := b.entity.Attributes.Set[id]; ok {
		attr.SetMaximum(fxp.FromInteger(value))
	}
}

// nextRow counts a row and returns the page reference and local notes it gets, if any: every third row has page
// references and every fifth has notes, so that the page reference columns and the notes disclosures both have work
// to do.
func (b *largeEntityBuilder) nextRow() (pageRef, notes string) {
	b.rows++
	if b.rows%3 == 0 {
		pageRef = []string{"B123,MA45", "B271", "LT12,B404,MA100", "DF1:23"}[b.rows/3%4]
	}
	if b.rows%5 == 0 {
		notes = fmt.Sprintf("Local note for row %d, with *emphasis* and a [link](https://example.com).", b.rows)
	}
	return pageRef, notes
}

func (b *largeEntityBuilder) addTraits() {
	var container *gurps.Trait
	for i := range b.size.Traits {
		if i%b.size.PerContainer == 0 {
			container = gurps.NewTrait(b.entity, nil, true)
			container.Name = fmt.Sprintf("Trait Group %d", i/b.size.PerContainer+1)
			b.entity.Traits = append(b.entity.Traits, container)
		}
		t := gurps.NewTrait(b.entity, container, false)
		t.Name = fmt.Sprintf("Trait %d", i+1)
		t.PageRef, t.LocalNotes = b.nextRow()
		t.BasePoints = fxp.FromInteger(i%7 + 1)
		if i%2 == 1 {
			t.BasePoints = -t.BasePoints
			t.Tags = []string{"Disadvantage"}
		} else {
			t.Tags = []string{"Advantage"}
		}
		if i%3 == 0 {
			t.CanLevel = true
			t.Levels = fxp.FromInteger(i%4 + 1)
			t.PointsPerLevel = fxp.Five
		}
		t.Modifiers = b.traitModifiers(i)
		if i%4 == 0 {
			b.addTraitNameables(t, i)
		}
		if bonus := b.derivedBonus(i); bonus != nil {
			t.Features = append(t.Features, bonus)
		}
		t.Features = append(t.Features, largeEntityTraitFeature(i)...)
		container.Children = append(container.Children, t)
		b.traits = append(b.traits, t)
	}
	// A size with fewer traits than reactions and conditional modifiers puts the rest on the traits it has, going
	// around again, or on one made for the purpose when it has none, so that the derived lists still show the rows
	// asked for.
	for i := b.size.Traits; i < b.size.Reactions+b.size.ConditionalModifiers; i++ {
		if len(b.traits) == 0 {
			t := gurps.NewTrait(b.entity, nil, false)
			t.Name = "Bonus Holder"
			b.entity.Traits = append(b.entity.Traits, t)
			b.traits = append(b.traits, t)
		}
		t := b.traits[i%len(b.traits)]
		t.Features = append(t.Features, b.derivedBonus(i))
	}
	b.entity.Traits = append(b.entity.Traits, gurps.NewNaturalAttacks(b.entity, nil))
	for i := range 3 {
		b.entity.Traits = append(b.entity.Traits,
			NewTraitNeedingMissingTrait(b.entity, fmt.Sprintf("Needs Combat Reflexes %d", i+1)))
	}
}

// addTraitNameables gives the i-th trait a nameable its name and its first modifier use, answered on every other such
// trait, a modifier whose code is linked to it, and a marker in its notes that defines itself.
func (b *largeEntityBuilder) addTraitNameables(t *gurps.Trait, i int) {
	const style = "style"
	t.Name += " (@style@)"
	t.LocalNotes += " Trained against @Rival school: Kodokan, Shotokan, etc.@."
	t.Nameables = map[string]nameable.Detail{style: {
		Label: "Style", Options: []string{"Judo", "Karate", "Boxing"}, FreeForm: true,
	}}
	if i%8 == 0 {
		t.Replacements = map[string]string{style: "Judo"}
	}
	t.Modifiers[0].Name += " (@style@)"
	linked := gurps.NewTraitModifier(b.entity, nil, false)
	linked.Name = "Focus on @art@"
	linked.Nameables = map[string]nameable.Detail{"art": {Label: "Art", FreeForm: true}}
	t.SetModifiers(append(t.Modifiers, linked))
	t.NameableLinks = map[tid.TID]map[string]string{linked.TID: {"art": style}}
}

// derivedBonus returns the i-th bonus behind the derived lists: a reaction bonus for the first Reactions values of i,
// a conditional modifier bonus for the next ConditionalModifiers, and nil beyond. Each has a situation of its own, so
// that every bonus is a row of its list.
func (b *largeEntityBuilder) derivedBonus(i int) gurps.Feature {
	switch {
	case i < b.size.Reactions:
		bonus := gurps.NewReactionBonus()
		bonus.Situation = fmt.Sprintf("from those who know of trait %d", i+1)
		bonus.Group = largeEntityGroup(i)
		bonus.Amount = fxp.FromInteger(i%3 + 1)
		return bonus
	case i < b.size.Reactions+b.size.ConditionalModifiers:
		bonus := gurps.NewConditionalModifierBonus()
		bonus.Situation = fmt.Sprintf("when trait %d applies", i+1)
		bonus.Group = largeEntityGroup(i)
		bonus.Amount = -fxp.FromInteger(i%3 + 1)
		return bonus
	default:
		return nil
	}
}

// largeEntityGroup returns the group every fourth reaction or conditional modifier is filed under, so that the two
// derived lists have containers as well as plain rows.
func largeEntityGroup(i int) string {
	if i%4 != 0 {
		return ""
	}
	return []string{"Social", "Combat", "Magical"}[i/4%3]
}

// largeEntityTraitFeature returns the features that make a trait matter to the rest of the sheet: every tenth trait
// has a different kind, chosen so that the skills, the spells, the attributes and the weapons each have bonuses to
// collect.
func largeEntityTraitFeature(i int) []gurps.Feature {
	switch i % 10 {
	case 2:
		bonus := gurps.NewSkillBonus()
		bonus.NameCriteria.Qualifier = largeEntityWeaponSkills[i/10%len(largeEntityWeaponSkills)].name
		bonus.Amount = fxp.One
		return []gurps.Feature{bonus}
	case 3:
		bonus := gurps.NewAttributeBonus(gurps.StrengthID)
		bonus.Limitation = stlimit.StrikingOnly
		bonus.Amount = fxp.One
		bonus.PerLevel = true
		return []gurps.Feature{bonus}
	case 4:
		bonus := gurps.NewSpellBonus()
		bonus.SpellMatchType = spellmatch.CollegeName
		bonus.NameCriteria.Compare = criteria.IsText
		bonus.NameCriteria.Qualifier = largeEntityColleges[i/10%len(largeEntityColleges)]
		bonus.Amount = fxp.One
		return []gurps.Feature{bonus}
	case 5:
		sw := gurps.NewWeaponBonus(feature.WeaponSwitch)
		sw.NameCriteria.Qualifier = largeEntityWeaponSkills[i/10%4].name
		sw.SwitchType = wswitch.CanParry
		sw.SwitchTypeValue = true
		return []gurps.Feature{sw}
	case 6:
		bonus := gurps.NewWeaponBonus(feature.WeaponBonus)
		bonus.NameCriteria.Qualifier = largeEntityWeaponSkills[i/10%4].name
		bonus.RelativeLevelCriteria.Compare = criteria.AtLeastNumber
		bonus.RelativeLevelCriteria.Qualifier = fxp.One
		bonus.PerDie = true
		return []gurps.Feature{bonus}
	default:
		return nil
	}
}

// traitModifiers returns the two modifiers every trait carries, the second of which is disabled. The modifier of
// every tenth trait is what carries that trait's weapon bonus, so that the bonuses reached through modifiers are
// collected too.
func (b *largeEntityBuilder) traitModifiers(i int) []*gurps.TraitModifier {
	first := gurps.NewTraitModifier(b.entity, nil, false)
	first.Name = fmt.Sprintf("Modifier %d", i%5+1)
	first.CostAdj = []string{"+10%", "-20%", "+50%", "-10%", "+100%"}[i%5]
	if i%10 == 1 {
		bonus := gurps.NewWeaponBonus(feature.WeaponBonus)
		bonus.NameCriteria.Qualifier = largeEntityWeaponSkills[i/10%4].name
		bonus.Amount = fxp.Two
		first.Features = append(first.Features, bonus)
	}
	second := gurps.NewTraitModifier(b.entity, nil, false)
	second.Name = "Unused Modifier"
	second.CostAdj = "-50%"
	second.Disabled = true
	return []*gurps.TraitModifier{first, second}
}

func (b *largeEntityBuilder) addSkills() {
	var container *gurps.Skill
	place := func(sk *gurps.Skill) {
		if len(b.skills)%b.size.PerContainer == 0 {
			container = gurps.NewSkill(b.entity, nil, true)
			container.Name = fmt.Sprintf("Skill Group %d", len(b.skills)/b.size.PerContainer+1)
			b.entity.Skills = append(b.entity.Skills, container)
		}
		sk.PageRef, sk.LocalNotes = b.nextRow()
		sk.SetParent(container)
		container.Children = append(container.Children, sk)
		b.skills = append(b.skills, sk)
	}
	for _, spec := range largeEntityWeaponSkills {
		sk := gurps.NewSkill(b.entity, nil, false)
		sk.Name = spec.name
		sk.Specialization = spec.specialization
		sk.Difficulty.Attribute = spec.attribute
		sk.Difficulty.Difficulty = spec.difficulty
		sk.Points = fxp.FromInteger(spec.points)
		place(sk)
	}
	for i, college := range largeEntityColleges {
		sk := gurps.NewSkill(b.entity, nil, false)
		sk.Name = ritualMagicSkillName
		sk.Specialization = "@Specialty@"
		sk.Replacements = map[string]string{"Specialty": college}
		sk.Nameables = map[string]nameable.Detail{"Specialty": {Label: "College", Options: largeEntityColleges}}
		sk.Difficulty.Attribute = gurps.IntelligenceID
		sk.Difficulty.Difficulty = difficulty.VeryHard
		sk.Points = fxp.FromInteger(4 * (i%3 + 1))
		sk.Defaults = []*gurps.SkillDefault{{
			DefaultType: gurps.SkillID,
			Name:        criteria.Text{Compare: criteria.IsText, Qualifier: ritualMagicSkillName},
			Modifier:    -fxp.Six,
		}}
		place(sk)
	}
	for i, name := range []string{"Fast-Draw (@Item@)", "Hobby Skill (@Hobby@)", "Expert Skill (@Field@)"} {
		sk := gurps.NewSkill(b.entity, nil, false)
		sk.Name = name
		if i < 2 {
			// The last one is left without a replacement, so that the fallback text for an unfilled marker is shown.
			sk.Replacements = map[string]string{"Item": "Sword", "Hobby": "Chess"}
		}
		sk.Difficulty.Attribute = gurps.DexterityID
		sk.Difficulty.Difficulty = difficulty.Easy
		sk.Points = fxp.One
		place(sk)
	}
	attributes := []string{gurps.DexterityID, gurps.IntelligenceID, gurps.HealthID}
	levels := []difficulty.Level{difficulty.Easy, difficulty.Average, difficulty.Hard, difficulty.VeryHard}
	for i := len(b.skills); i < b.size.Skills; i++ {
		var sk *gurps.Skill
		if i%15 == 0 {
			sk = gurps.NewTechnique(b.entity, nil, largeEntityWeaponSkills[i/15%5].name)
			sk.Name = fmt.Sprintf("Technique %d", i+1)
			if i%30 == 0 {
				sk.Difficulty.Difficulty = difficulty.Hard
			}
		} else {
			sk = gurps.NewSkill(b.entity, nil, false)
			sk.Name = fmt.Sprintf("Skill %d", i+1)
			sk.Difficulty.Attribute = attributes[i%len(attributes)]
			sk.Difficulty.Difficulty = levels[i%len(levels)]
			if i%4 == 0 {
				tl := "3"
				sk.TechLevel = &tl
			}
			if i%6 == 0 {
				sk.Defaults = []*gurps.SkillDefault{
					{DefaultType: gurps.DexterityID, Modifier: -fxp.Five},
					{
						DefaultType: gurps.SkillID,
						Name: criteria.Text{
							Compare:   criteria.IsText,
							Qualifier: largeEntityWeaponSkills[i/6%len(largeEntityWeaponSkills)].name,
						},
						Modifier: -fxp.Four,
					},
				}
			}
		}
		sk.Points = fxp.FromInteger([]int{0, 1, 2, 4, 8}[i%5])
		place(sk)
	}
}

func (b *largeEntityBuilder) addSpells() {
	colleges := make([]*gurps.Spell, len(largeEntityColleges))
	for i, college := range largeEntityColleges {
		colleges[i] = gurps.NewSpell(b.entity, nil, true)
		colleges[i].Name = college
		b.entity.Spells = append(b.entity.Spells, colleges[i])
	}
	subs := make([]*gurps.Spell, len(largeEntityColleges))
	place := func(s *gurps.Spell, college int) {
		parent := colleges[college]
		if len(parent.Children) == 0 || len(subs[college].Children) == b.size.PerContainer {
			subs[college] = gurps.NewSpell(b.entity, parent, true)
			subs[college].Name = fmt.Sprintf("%s Spells %d", largeEntityColleges[college], len(parent.Children)+1)
			parent.Children = append(parent.Children, subs[college])
		}
		s.PageRef, s.LocalNotes = b.nextRow()
		s.SetParent(subs[college])
		subs[college].Children = append(subs[college].Children, s)
		b.spells = append(b.spells, s)
	}
	classes := []string{"Regular", "Area", "Missile", "Blocking", "Regular; Resisted by HT"}
	for i := range b.size.Spells {
		college := i % len(largeEntityColleges)
		s := gurps.NewSpell(b.entity, nil, false)
		s.Name = fmt.Sprintf("%s Spell %d", largeEntityColleges[college], i/len(largeEntityColleges)+1)
		s.College = gurps.CollegeList{largeEntityColleges[college]}
		s.Class = classes[i%len(classes)]
		if i%3 == 0 {
			s.Resist = "HT"
		}
		s.CastingCost = strconv.Itoa(i%4 + 1)
		s.MaintenanceCost = []string{"", "Same", "Half", "1"}[i%4]
		s.CastingTime = []string{"1 second", "2 seconds", "5 seconds", "1 minute"}[i%4]
		s.Duration = []string{"Instant", "1 minute", "10 minutes", "1 hour"}[i%4]
		if i%7 == 0 {
			s.Difficulty.Difficulty = difficulty.VeryHard
		}
		s.Points = fxp.FromInteger([]int{1, 1, 2, 4, 8}[i%5])
		s.Tags = []string{"Spell", largeEntityColleges[college]}
		place(s, college)
	}
	for i := range b.size.RitualMagicSpells {
		college := i % len(largeEntityColleges)
		s := gurps.NewRitualMagicSpell(b.entity, nil, false)
		s.Name = fmt.Sprintf("%s Ritual %d", largeEntityColleges[college], i/len(largeEntityColleges)+1)
		s.College = gurps.CollegeList{largeEntityColleges[college]}
		if i%4 == 0 {
			s.College = append(s.College, largeEntityColleges[(college+1)%len(largeEntityColleges)])
		}
		s.PrereqCount = i % 6
		s.Points = fxp.FromInteger(i % 3)
		s.Class = classes[i%len(classes)]
		s.Tags = []string{"Ritual", largeEntityColleges[college]}
		place(s, college)
	}
}

func (b *largeEntityBuilder) addEquipment(count int, carried bool) {
	per := b.size.PerContainer
	var top, mid *gurps.Equipment
	for i := range count {
		if i%(per*per) == 0 {
			top = gurps.NewEquipment(b.entity, nil, true)
			top.Name = fmt.Sprintf("Pack %d", i/(per*per)+1)
			top.BaseWeight = "2 lb"
			top.BaseValue = "50"
			if carried {
				b.entity.CarriedEquipment = append(b.entity.CarriedEquipment, top)
			} else {
				b.entity.OtherEquipment = append(b.entity.OtherEquipment, top)
			}
		}
		if i%per == 0 {
			mid = gurps.NewEquipment(b.entity, top, true)
			mid.Name = fmt.Sprintf("Pouch %d", i/per+1)
			mid.BaseWeight = "0.5 lb"
			mid.BaseValue = "5"
			top.Children = append(top.Children, mid)
		}
		item := gurps.NewEquipment(b.entity, mid, false)
		item.Name = fmt.Sprintf("Item %d", i+1)
		item.PageRef, item.LocalNotes = b.nextRow()
		item.Quantity = fxp.FromInteger(i%3 + 1)
		item.BaseValue = strconv.Itoa(10 * (i%20 + 1))
		item.BaseWeight = fmt.Sprintf("%g lb", 0.25*float64(i%8+1))
		item.TechLevel = strconv.Itoa(i%5 + 3)
		item.LegalityClass = strconv.Itoa(i%3 + 2)
		if i%5 == 0 {
			item.MaxUses = 10
			item.Uses = i % 10
		}
		item.Tags = []string{[]string{"Melee Weapon", "Ranged Weapon", "Armor", "Tool", "Consumable"}[i%5]}
		if i%2 == 0 {
			item.Modifiers = b.equipmentModifiers(i)
		}
		if i%4 == 0 {
			b.addEquipmentNameables(item, i)
		}
		mid.Children = append(mid.Children, item)
		switch {
		case !carried:
			b.other = append(b.other, item)
		case i%7 == 0:
			item.Equipped = false
			b.unequipped = append(b.unequipped, item)
		default:
			b.equipped = append(b.equipped, item)
		}
	}
}

// addEquipmentNameables gives the i-th piece of equipment, which has modifiers, a nameable its name and its first
// modifier use, answered on every other such piece, and a modifier whose code is linked to it.
func (b *largeEntityBuilder) addEquipmentNameables(item *gurps.Equipment, i int) {
	const metal = "metal"
	item.Name += " of @metal@"
	item.Nameables = map[string]nameable.Detail{metal: {Label: "Metal", Options: []string{"Bronze", "Iron", "Steel"}}}
	if i%8 == 0 {
		item.Replacements = map[string]string{metal: "Steel"}
	}
	item.Modifiers[0].Name += " @metal@"
	linked := gurps.NewEquipmentModifier(b.entity, nil, false)
	linked.Name = "@finish@ inlay"
	linked.Nameables = map[string]nameable.Detail{"finish": {Label: "Finish", FreeForm: true}}
	item.SetModifiers(append(item.Modifiers, linked))
	item.NameableLinks = map[tid.TID]map[string]string{linked.TID: {"finish": metal}}
}

// equipmentModifiers returns the two modifiers every other piece of equipment carries: one that alters its cost and
// weight, and one that alters its weapons through a feature.
func (b *largeEntityBuilder) equipmentModifiers(i int) []*gurps.EquipmentModifier {
	first := gurps.NewEquipmentModifier(b.entity, nil, false)
	first.Name = []string{"Fine", "Cheap", "Good", "Ornate"}[i%4]
	first.CostAmount = []string{"+300%", "-60%", "+100%", "+1000%"}[i%4]
	if i%4 == 0 {
		first.WeightAmount = "-10%"
	}
	second := gurps.NewEquipmentModifier(b.entity, nil, false)
	switch i % 3 {
	case 0:
		second.Name = "Balanced"
		second.CostAmount = "+400%"
		bonus := gurps.NewWeaponBonus(feature.WeaponBonus)
		bonus.SelectionType = wsel.ThisWeapon
		bonus.Amount = fxp.One
		second.Features = append(second.Features, bonus)
	case 1:
		second.Name = "Fencing Grip"
		second.CostAmount = "+50%"
		sw := gurps.NewWeaponBonus(feature.WeaponSwitch)
		sw.SelectionType = wsel.ThisWeapon
		sw.SwitchType = wswitch.Fencing
		sw.SwitchTypeValue = true
		second.Features = append(second.Features, sw)
	default:
		second.Name = "Heavy"
		second.WeightAmount = "+1 lb"
		bonus := gurps.NewWeaponBonus(feature.WeaponMinSTBonus)
		bonus.SelectionType = wsel.ThisWeapon
		bonus.Amount = fxp.Two
		second.Features = append(second.Features, bonus)
	}
	if i%8 == 0 {
		second.Disabled = true
	}
	return []*gurps.EquipmentModifier{first, second}
}

func (b *largeEntityBuilder) addNotes() {
	var container *gurps.Note
	for i := range b.size.Notes {
		if i%b.size.PerContainer == 0 {
			container = gurps.NewNote(b.entity, nil, true)
			container.MarkDown = fmt.Sprintf("Notes %d", i/b.size.PerContainer+1)
			b.entity.Notes = append(b.entity.Notes, container)
		}
		n := gurps.NewNote(b.entity, container, false)
		n.MarkDown = fmt.Sprintf("## Note %d\n\nSome **markdown** with a [link](https://example.com) and a list:\n\n"+
			"- the first point\n- the second point, which is a little longer than the first\n", i+1)
		n.PageRef, _ = b.nextRow()
		if i%4 == 0 {
			n.MarkDown += "\nAbout @topic@.\n"
			n.Nameables = map[string]nameable.Detail{"topic": {Label: "Topic", FreeForm: true}}
			if i%8 == 0 {
				n.Replacements = map[string]string{"topic": "dragons"}
			}
		}
		container.Children = append(container.Children, n)
	}
}

// addWeapons hands out the weapons: most go to equipped carried equipment, the rest to traits, skills, spells and,
// to exercise the paths that keep them off the page, to unequipped and uncarried equipment.
func (b *largeEntityBuilder) addWeapons() {
	for i := range b.size.MeleeWeapons {
		var owner gurps.WeaponOwner
		switch {
		case i%10 < 6:
			owner = b.equipmentOwner(i)
		case i%10 < 8:
			owner = pickOwner(b, b.traits, i)
		case i%10 == 8:
			owner = pickOwner(b, b.skills, i)
		default:
			owner = pickOwner(b, b.spells, i)
		}
		b.attachWeapon(owner, b.newMeleeWeapon(owner, i))
	}
	for i := range b.size.RangedWeapons {
		var owner gurps.WeaponOwner
		switch {
		case i%10 < 6:
			owner = b.equipmentOwner(i)
		case i%10 < 8:
			owner = pickOwner(b, b.spells, i)
		default:
			owner = pickOwner(b, b.traits, i)
		}
		b.attachWeapon(owner, b.newRangedWeapon(owner, i))
	}
}

// equipmentOwner returns the equipment the i-th weapon of a kind goes on: usually an equipped item, with every twelfth
// going to an unequipped one and every thirteenth to uncarried equipment.
func (b *largeEntityBuilder) equipmentOwner(i int) gurps.WeaponOwner {
	switch {
	case i%12 == 11 && len(b.unequipped) > 0:
		return pickOwner(b, b.unequipped, i)
	case i%13 == 12 && len(b.other) > 0:
		return pickOwner(b, b.other, i)
	default:
		return pickOwner(b, b.equipped, i)
	}
}

// pickOwner returns the i-th element of the list, wrapping around, or a trait made for the purpose when the list is
// empty, which only a size that leaves that list empty brings about.
func pickOwner[T gurps.WeaponOwner](b *largeEntityBuilder, list []T, i int) gurps.WeaponOwner {
	if len(list) == 0 {
		t := gurps.NewTrait(b.entity, nil, false)
		t.Name = fmt.Sprintf("Weapon Holder %d", i+1)
		b.entity.Traits = append(b.entity.Traits, t)
		return t
	}
	return list[i%len(list)]
}

func (b *largeEntityBuilder) attachWeapon(owner gurps.WeaponOwner, w *gurps.Weapon) {
	switch o := owner.(type) {
	case *gurps.Equipment:
		o.Weapons = append(o.Weapons, w)
	case *gurps.Trait:
		o.Weapons = append(o.Weapons, w)
	case *gurps.Skill:
		o.Weapons = append(o.Weapons, w)
	case *gurps.Spell:
		o.Weapons = append(o.Weapons, w)
	}
}

func (b *largeEntityBuilder) newMeleeWeapon(owner gurps.WeaponOwner, i int) *gurps.Weapon {
	spec := &largeEntityMeleeSpecs[i%len(largeEntityMeleeSpecs)]
	w := gurps.NewWeapon(owner, true)
	w.Damage.Owner = w
	w.Usage = spec.usage
	w.Strength = attack.ParseStrength(spec.strength)
	w.Damage.StrengthType = spec.strengthType
	w.Damage.Base = spec.base
	w.Damage.Type = spec.damageType
	w.Reach = attack.ParseReach(spec.reach)
	w.Parry = attack.ParseParry(spec.parry)
	w.Block = attack.ParseBlock(spec.block)
	w.Defaults = []*gurps.SkillDefault{
		{DefaultType: gurps.DexterityID, Modifier: -fxp.Four},
		largeEntitySkillDefault(spec.skill, 0),
		largeEntitySkillDefault(spec.alt, -2),
	}
	if i%5 == 0 {
		w.Defaults = append(w.Defaults, &gurps.SkillDefault{
			DefaultType: gurps.SkillID,
			Name:        criteria.Text{Compare: criteria.IsText, Qualifier: "Shield"},
			Modifier:    -fxp.Three,
		})
	}
	return w
}

func (b *largeEntityBuilder) newRangedWeapon(owner gurps.WeaponOwner, i int) *gurps.Weapon {
	spec := &largeEntityRangedSpecs[i%len(largeEntityRangedSpecs)]
	w := gurps.NewWeapon(owner, false)
	w.Damage.Owner = w
	w.Usage = spec.usage
	w.Strength = attack.ParseStrength(spec.strength)
	w.Damage.Base = spec.base
	w.Damage.Type = spec.damageType
	if spec.muscle {
		w.Damage.StrengthType = stdmg.Thrust
	} else {
		w.Damage.StrengthType = stdmg.None
	}
	w.Accuracy = attack.ParseAccuracy(spec.accuracy)
	w.Range = attack.ParseRange(spec.rng)
	w.RateOfFire = attack.ParseRoF(spec.rof)
	w.Shots = attack.ParseShots(spec.shots)
	w.Bulk = attack.ParseBulk(spec.bulk)
	w.Recoil = attack.ParseRecoil(spec.recoil)
	w.Defaults = []*gurps.SkillDefault{
		{DefaultType: gurps.DexterityID, Modifier: -fxp.Four},
		largeEntitySkillDefault(spec.skill, 0),
	}
	return w
}

// largeEntitySkillDefault returns a default to the weapon skill at the index in largeEntityWeaponSkills.
func largeEntitySkillDefault(index, modifier int) *gurps.SkillDefault {
	spec := &largeEntityWeaponSkills[index]
	def := &gurps.SkillDefault{
		DefaultType: gurps.SkillID,
		Name:        criteria.Text{Compare: criteria.IsText, Qualifier: spec.name},
		Modifier:    fxp.FromInteger(modifier),
	}
	if spec.specialization != "" {
		def.Specialization = criteria.Text{Compare: criteria.IsText, Qualifier: spec.specialization}
	}
	return def
}
