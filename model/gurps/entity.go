// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"context"
	"fmt"
	"hash"
	"io/fs"
	"log/slog"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/progression"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/skillsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/threshold"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/collection/dict"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/eval"
	"github.com/richardwilkes/toolbox/fatal"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/tid"
	"github.com/richardwilkes/toolbox/xio"
)

var (
	_ eval.VariableResolver = &Entity{}
	_ ListProvider          = &Entity{}
	_ DataOwner             = &Entity{}
	_ Hashable              = &Entity{}
	_ PageInfoProvider      = &Entity{}
)

// PointsBreakdown holds the points spent on a character.
type PointsBreakdown struct {
	Ancestry      fxp.Int
	Attributes    fxp.Int
	Advantages    fxp.Int
	Disadvantages fxp.Int
	Quirks        fxp.Int
	Skills        fxp.Int
	Spells        fxp.Int
}

// Total returns the total number of points spent on a character.
func (pb *PointsBreakdown) Total() fxp.Int {
	return pb.Ancestry + pb.Attributes + pb.Advantages + pb.Disadvantages + pb.Quirks + pb.Skills + pb.Spells
}

// EntityData holds the Entity data that is written to disk.
type EntityData struct {
	Version          int             `json:"version"`
	ID               tid.TID         `json:"id"`
	TotalPoints      fxp.Int         `json:"total_points"`
	PointsRecord     []*PointsRecord `json:"points_record,omitempty"`
	Profile          Profile         `json:"profile"`
	SheetSettings    *SheetSettings  `json:"settings,omitempty"`
	Attributes       *Attributes     `json:"attributes,omitempty"`
	Traits           []*Trait        `json:"traits,alt=advantages,omitempty"`
	Skills           []*Skill        `json:"skills,omitempty"`
	Spells           []*Spell        `json:"spells,omitempty"`
	CarriedEquipment []*Equipment    `json:"equipment,omitempty"`
	OtherEquipment   []*Equipment    `json:"other_equipment,omitempty"`
	Notes            []*Note         `json:"notes,omitempty"`
	CreatedOn        jio.Time        `json:"created_date"`
	ModifiedOn       jio.Time        `json:"modified_date"`
	ThirdParty       map[string]any  `json:"third_party,omitempty"`
}

type features struct {
	attributeBonuses  []*AttributeBonus
	costReductions    []*CostReduction
	drBonuses         []*DRBonus
	skillBonuses      []*SkillBonus
	skillPointBonuses []*SkillPointBonus
	spellBonuses      []*SpellBonus
	spellPointBonuses []*SpellPointBonus
	weaponBonuses     []*WeaponBonus
}

// Entity holds the base information for various types of entities: PC, NPC, Creature, etc.
type Entity struct {
	EntityData
	LiftingStrengthBonus           fxp.Int
	StrikingStrengthBonus          fxp.Int
	ThrowingStrengthBonus          fxp.Int
	DodgeBonus                     fxp.Int
	ParryBonus                     fxp.Int
	ParryBonusTooltip              string
	BlockBonus                     fxp.Int
	BlockBonusTooltip              string
	srcMatcher                     *SrcMatcher
	features                       features
	variableResolverExclusions     map[string]bool
	skillResolverExclusions        map[string]bool
	scriptCache                    map[scriptResolveKey]string
	variableCache                  map[string]string
	basicLiftCache                 fxp.Weight
	encumbranceLevelCache          encumbrance.Level
	encumbranceLevelForSkillsCache encumbrance.Level
}

// NewEntityFromFile loads an Entity from a file.
func NewEntityFromFile(fileSystem fs.FS, filePath string) (*Entity, error) {
	var e Entity
	e.DiscardCaches()
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &e); err != nil {
		return nil, errs.NewWithCause(InvalidFileData(), err)
	}
	if err := jio.CheckVersion(e.Version); err != nil {
		return nil, err
	}
	return &e, nil
}

// NewEntity creates a new Entity.
func NewEntity() *Entity {
	settings := GlobalSettings().GeneralSettings()
	var e Entity
	e.DiscardCaches()
	e.ID = tid.MustNewTID(kinds.Entity)
	e.TotalPoints = settings.InitialPoints
	e.PointsRecord = append(e.PointsRecord, &PointsRecord{
		When:   jio.Now(),
		Points: settings.InitialPoints,
		Reason: i18n.Text("Initial points"),
	})
	e.CreatedOn = jio.Now()
	e.SheetSettings = GlobalSettings().SheetSettings().Clone(&e)
	e.Attributes = NewAttributes(&e)
	if settings.AutoFillProfile {
		e.Profile.AutoFill(&e)
	}
	if settings.AutoAddNaturalAttacks {
		e.Traits = append(e.Traits, NewNaturalAttacks(&e, nil))
	}
	e.ModifiedOn = e.CreatedOn
	e.Recalculate()
	return &e
}

// DataOwner returns the data owner.
func (e *Entity) DataOwner() DataOwner {
	return e
}

// OwningEntity returns the Entity.
func (e *Entity) OwningEntity() *Entity {
	return e
}

// SourceMatcher returns the SourceMatcher.
func (e *Entity) SourceMatcher() *SrcMatcher {
	if e.srcMatcher == nil {
		e.srcMatcher = &SrcMatcher{}
	}
	return e.srcMatcher
}

// Entity implements EntityProvider.
func (e *Entity) Entity() *Entity {
	return e
}

// Save the Entity to a file as JSON.
func (e *Entity) Save(filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, e)
}

// MarshalJSON implements json.Marshaler.
func (e *Entity) MarshalJSON() ([]byte, error) {
	e.Recalculate()
	type calc struct {
		Swing                 *dice.Dice `json:"swing"`
		Thrust                *dice.Dice `json:"thrust"`
		BasicLift             fxp.Weight `json:"basic_lift"`
		LiftingStrengthBonus  fxp.Int    `json:"lifting_st_bonus,omitempty"`
		StrikingStrengthBonus fxp.Int    `json:"striking_st_bonus,omitempty"`
		ThrowingStrengthBonus fxp.Int    `json:"throwing_st_bonus,omitempty"`
		DodgeBonus            fxp.Int    `json:"dodge_bonus,omitempty"`
		ParryBonus            fxp.Int    `json:"parry_bonus,omitempty"`
		BlockBonus            fxp.Int    `json:"block_bonus,omitempty"`
		Move                  []int      `json:"move"`
		Dodge                 []int      `json:"dodge"`
	}
	data := struct {
		EntityData
		Calc calc `json:"calc"`
	}{
		EntityData: e.EntityData,
		Calc: calc{
			Swing:                 e.Swing(),
			Thrust:                e.Thrust(),
			BasicLift:             e.BasicLift(),
			LiftingStrengthBonus:  e.LiftingStrengthBonus,
			StrikingStrengthBonus: e.StrikingStrengthBonus,
			ThrowingStrengthBonus: e.ThrowingStrengthBonus,
			DodgeBonus:            e.DodgeBonus,
			ParryBonus:            e.ParryBonus,
			BlockBonus:            e.BlockBonus,
			Move:                  make([]int, len(encumbrance.Levels)),
			Dodge:                 make([]int, len(encumbrance.Levels)),
		},
	}
	data.Version = jio.CurrentDataVersion
	for i, one := range encumbrance.Levels {
		data.Calc.Move[i] = e.Move(one)
		data.Calc.Dodge[i] = e.Dodge(one)
	}
	return json.Marshal(&data)
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *Entity) UnmarshalJSON(data []byte) error {
	e.EntityData = EntityData{}
	if err := json.Unmarshal(data, &e.EntityData); err != nil {
		return err
	}
	if !tid.IsKindAndValid(e.ID, kinds.Entity) {
		e.ID = tid.MustNewTID(kinds.Entity)
	}
	if e.SheetSettings == nil {
		e.SheetSettings = GlobalSettings().SheetSettings().Clone(e)
	}
	if e.Attributes == nil {
		e.Attributes = NewAttributes(e)
	}
	if e.Version < noNeedForRewrapVersion {
		e.SheetSettings.BodyType.Rewrap()
	}
	var total fxp.Int
	for _, rec := range e.PointsRecord {
		total += rec.Points
	}
	if total != e.TotalPoints {
		e.PointsRecord = append(e.PointsRecord, &PointsRecord{
			Points: e.TotalPoints - total,
			When:   jio.Now(),
			Reason: i18n.Text("Reconciliation"),
		})
		slices.SortFunc(e.PointsRecord, func(a, b *PointsRecord) int { return b.When.Compare(a.When) })
	}
	e.Recalculate()
	return nil
}

// DiscardCaches discards the internal caches.
func (e *Entity) DiscardCaches() {
	e.variableResolverExclusions = make(map[string]bool)
	e.skillResolverExclusions = make(map[string]bool)
	e.scriptCache = make(map[scriptResolveKey]string)
	e.variableCache = make(map[string]string)
	e.basicLiftCache = -1
	e.encumbranceLevelCache = encumbrance.LastLevel + 1
	e.encumbranceLevelForSkillsCache = encumbrance.LastLevel + 1
}

// Recalculate the statistics.
func (e *Entity) Recalculate() {
	if e == nil {
		return
	}
	e.ensureAttachments()
	e.DiscardCaches()
	e.SourceMatcher().PrepareHashes(e)
	e.UpdateSkills()
	e.UpdateSpells()
	for range 5 {
		// Some skill & spell levels won't be correct until the features & prerequisites have been processed, and those
		// can't be processed in some cases until the skills & spells have known levels. Due to this circular
		// referencing, we need to update the skills & spells at least twice. Once they no longer change, we can stop
		// processing. To avoid an infinite loop, we limit the number of iterations to 5.
		e.processFeatures()
		e.processPrereqs()
		e.DiscardCaches()
		skillsChanged := e.UpdateSkills()
		spellsChanged := e.UpdateSpells()
		if !skillsChanged && !spellsChanged {
			break
		}
	}
}

func (e *Entity) ensureAttachments() {
	e.SheetSettings.SetOwningEntity(e)
	for _, attr := range e.Attributes.Set {
		attr.Entity = e
	}
	for _, one := range e.Traits {
		one.SetDataOwner(e)
	}
	for _, one := range e.Skills {
		one.SetDataOwner(e)
	}
	for _, one := range e.Spells {
		one.SetDataOwner(e)
	}
	for _, one := range e.CarriedEquipment {
		one.SetDataOwner(e)
	}
	for _, one := range e.OtherEquipment {
		one.SetDataOwner(e)
	}
	for _, one := range e.Notes {
		one.SetDataOwner(e)
	}
}

func (e *Entity) processFeatures() {
	e.features = features{}
	Traverse(func(a *Trait) bool {
		var levels fxp.Int
		if a.IsLeveled() {
			levels = a.Levels.Max(0)
		}
		if !a.Container() {
			for _, f := range a.Features {
				e.processFeature(a, nil, f, levels)
			}
		}
		for _, f := range FeaturesForSelfControlRoll(a.CR, a.CRAdj) {
			e.processFeature(a, nil, f, levels)
		}
		Traverse(func(mod *TraitModifier) bool {
			for _, f := range mod.Features {
				e.processFeature(a, nil, f, mod.CurrentLevel())
			}
			return false
		}, true, true, a.Modifiers...)
		return false
	}, true, false, e.Traits...)
	Traverse(func(s *Skill) bool {
		for _, f := range s.Features {
			e.processFeature(s, nil, f, s.LevelData.Level)
		}
		return false
	}, false, true, e.Skills...)
	Traverse(func(eqp *Equipment) bool {
		if !eqp.Equipped || eqp.Quantity <= 0 {
			return false
		}
		for _, f := range eqp.Features {
			e.processFeature(eqp, nil, f, eqp.Level.Max(0))
		}
		Traverse(func(mod *EquipmentModifier) bool {
			for _, f := range mod.Features {
				e.processFeature(eqp, mod, f, eqp.Level.Max(0))
			}
			return false
		}, true, true, eqp.Modifiers...)
		return false
	}, false, false, e.CarriedEquipment...)
	e.LiftingStrengthBonus = e.AttributeBonusFor(StrengthID, stlimit.LiftingOnly, nil).Trunc()
	e.StrikingStrengthBonus = e.AttributeBonusFor(StrengthID, stlimit.StrikingOnly, nil).Trunc()
	e.ThrowingStrengthBonus = e.AttributeBonusFor(StrengthID, stlimit.ThrowingOnly, nil).Trunc()
	for _, attr := range e.Attributes.Set {
		if def := attr.AttributeDef(); def != nil {
			attr.Bonus = e.AttributeBonusFor(attr.AttrID, stlimit.None, nil)
			if !def.AllowsDecimal() {
				attr.Bonus = attr.Bonus.Trunc()
			}
			attr.CostReduction = e.CostReductionFor(attr.AttrID)
		} else {
			attr.Bonus = 0
			attr.CostReduction = 0
		}
	}
	e.Profile.Update(e)
	if e.ResolveAttribute(DodgeID) == nil {
		e.DodgeBonus = e.AttributeBonusFor(DodgeID, stlimit.None, nil).Trunc()
	} else {
		e.DodgeBonus = 0
	}
	var tooltip xio.ByteBuffer
	e.ParryBonus = e.AttributeBonusFor(ParryID, stlimit.None, &tooltip).Trunc()
	e.ParryBonusTooltip = tooltip.String()
	tooltip.Reset()
	e.BlockBonus = e.AttributeBonusFor(BlockID, stlimit.None, &tooltip).Trunc()
	e.BlockBonusTooltip = tooltip.String()
}

func (e *Entity) processFeature(owner, subOwner fmt.Stringer, f Feature, levels fxp.Int) {
	if bonus, ok := f.(Bonus); ok {
		bonus.SetOwner(owner)
		bonus.SetSubOwner(subOwner)
		bonus.SetLevel(levels)
	}
	switch actual := f.(type) {
	case *AttributeBonus:
		e.features.attributeBonuses = append(e.features.attributeBonuses, actual)
	case *CostReduction:
		e.features.costReductions = append(e.features.costReductions, actual)
	case *DRBonus:
		if len(actual.Locations) == 0 { // "this armor"
			if eqp, ok := owner.(*Equipment); ok {
				allLocations := make(map[string]bool)
				locationsMatched := make(map[string]bool)
				for _, f2 := range eqp.FeatureList() {
					if drBonus, ok2 := f2.(*DRBonus); ok2 && len(drBonus.Locations) != 0 {
						for _, loc := range drBonus.Locations {
							allLocations[loc] = true
						}
						if drBonus.Specialization == actual.Specialization {
							for _, loc := range drBonus.Locations {
								locationsMatched[loc] = true
							}
							additionalDRBonus := DRBonus{
								DRBonusData: DRBonusData{
									Type:           feature.DRBonus,
									Locations:      slices.Clone(drBonus.Locations),
									Specialization: actual.Specialization,
									LeveledAmount:  actual.LeveledAmount,
								},
							}
							additionalDRBonus.SetOwner(owner)
							additionalDRBonus.SetSubOwner(subOwner)
							additionalDRBonus.SetLevel(levels)
							e.features.drBonuses = append(e.features.drBonuses, &additionalDRBonus)
						}
					}
				}
				for k := range locationsMatched {
					delete(allLocations, k)
				}
				if len(allLocations) != 0 {
					locations := dict.Keys(allLocations)
					slices.Sort(locations)
					additionalDRBonus := DRBonus{
						DRBonusData: DRBonusData{
							Type:           feature.DRBonus,
							Locations:      locations,
							Specialization: actual.Specialization,
							LeveledAmount:  actual.LeveledAmount,
						},
					}
					additionalDRBonus.SetOwner(owner)
					additionalDRBonus.SetSubOwner(subOwner)
					additionalDRBonus.SetLevel(levels)
					e.features.drBonuses = append(e.features.drBonuses, &additionalDRBonus)
				}
			}
		} else {
			e.features.drBonuses = append(e.features.drBonuses, actual)
		}
	case *SkillBonus:
		e.features.skillBonuses = append(e.features.skillBonuses, actual)
	case *SkillPointBonus:
		e.features.skillPointBonuses = append(e.features.skillPointBonuses, actual)
	case *SpellBonus:
		e.features.spellBonuses = append(e.features.spellBonuses, actual)
	case *SpellPointBonus:
		e.features.spellPointBonuses = append(e.features.spellPointBonuses, actual)
	case *WeaponBonus:
		e.features.weaponBonuses = append(e.features.weaponBonuses, actual)
	case *ConditionalModifierBonus, *ContainedWeightReduction, *ReactionBonus:
		// Not collected at this stage
	default:
		errs.Log(errs.New("unhandled feature"), "type", f.FeatureType())
	}
}

func (e *Entity) processPrereqs() {
	const prefix = "\n● "
	notMetPrefix := i18n.Text("Prerequisites have not been met:")
	Traverse(func(a *Trait) bool {
		a.UnsatisfiedReason = ""
		if a.Prereq != nil {
			var tooltip xio.ByteBuffer
			var eqpPenalty bool
			if !a.Prereq.Satisfied(e, a, &tooltip, prefix, &eqpPenalty) {
				a.UnsatisfiedReason = notMetPrefix + tooltip.String()
			}
		}
		return false
	}, true, false, e.Traits...)
	Traverse(func(s *Skill) bool {
		s.UnsatisfiedReason = ""
		if !s.Container() {
			var tooltip xio.ByteBuffer
			satisfied := true
			if s.Prereq != nil {
				var eqpPenalty bool
				satisfied = s.Prereq.Satisfied(e, s, &tooltip, prefix, &eqpPenalty)
				if eqpPenalty {
					penalty := NewSkillBonus()
					penalty.NameCriteria.Qualifier = s.NameWithReplacements()
					penalty.SpecializationCriteria.Compare = criteria.IsText
					penalty.SpecializationCriteria.Qualifier = s.SpecializationWithReplacements()
					if s.TechLevel != nil && *s.TechLevel != "" {
						penalty.Amount = -fxp.Ten
					} else {
						penalty.Amount = -fxp.Five
					}
					penalty.SetOwner(s)
					e.features.skillBonuses = append(e.features.skillBonuses, penalty)
				}
			}
			if satisfied && s.IsTechnique() {
				satisfied = s.TechniqueSatisfied(&tooltip, prefix)
			}
			if !satisfied {
				s.UnsatisfiedReason = notMetPrefix + tooltip.String()
			}
		}
		return false
	}, false, false, e.Skills...)
	Traverse(func(s *Spell) bool {
		s.UnsatisfiedReason = ""
		if !s.Container() {
			var tooltip xio.ByteBuffer
			satisfied := true
			if s.Prereq != nil {
				var eqpPenalty bool
				satisfied = s.Prereq.Satisfied(e, s, &tooltip, prefix, &eqpPenalty)
				if eqpPenalty {
					penalty := NewSpellBonus()
					penalty.NameCriteria.Qualifier = s.NameWithReplacements()
					if s.TechLevel != nil && *s.TechLevel != "" {
						penalty.Amount = -fxp.Ten
					} else {
						penalty.Amount = -fxp.Five
					}
					penalty.SetOwner(s)
					e.features.spellBonuses = append(e.features.spellBonuses, penalty)
				}
			}
			if satisfied && s.IsRitualMagic() {
				satisfied = s.RitualMagicSatisfied(&tooltip, prefix)
			}
			if !satisfied {
				s.UnsatisfiedReason = notMetPrefix + tooltip.String()
			}
		}
		return false
	}, false, false, e.Spells...)
	equipmentFunc := func(eqp *Equipment) bool {
		eqp.UnsatisfiedReason = ""
		if eqp.Prereq != nil {
			var tooltip xio.ByteBuffer
			var eqpPenalty bool
			if !eqp.Prereq.Satisfied(e, eqp, &tooltip, prefix, &eqpPenalty) {
				eqp.UnsatisfiedReason = notMetPrefix + tooltip.String()
			}
		}
		return false
	}
	Traverse(equipmentFunc, false, false, e.CarriedEquipment...)
	Traverse(equipmentFunc, false, false, e.OtherEquipment...)
}

// UpdateSkills updates the levels of all skills.
func (e *Entity) UpdateSkills() bool {
	changed := false
	Traverse(func(s *Skill) bool {
		if s.UpdateLevel() {
			changed = true
		}
		return false
	}, false, true, e.Skills...)
	return changed
}

// UpdateSpells updates the levels of all spells.
func (e *Entity) UpdateSpells() bool {
	changed := false
	Traverse(func(s *Spell) bool {
		if s.UpdateLevel() {
			changed = true
		}
		return false
	}, false, true, e.Spells...)
	return changed
}

// UnspentPoints returns the number of unspent points.
func (e *Entity) UnspentPoints() fxp.Int {
	return e.TotalPoints - e.PointsBreakdown().Total()
}

// SetUnspentPoints sets the number of unspent points.
func (e *Entity) SetUnspentPoints(unspent fxp.Int) {
	if unspent != e.UnspentPoints() {
		e.TotalPoints = unspent + e.PointsBreakdown().Total()
	}
}

// PointsBreakdown returns the point breakdown for spent points.
func (e *Entity) PointsBreakdown() *PointsBreakdown {
	var pb PointsBreakdown
	for _, attr := range e.Attributes.Set {
		pb.Attributes += attr.PointCost()
	}
	for _, one := range e.Traits {
		calculateSingleTraitPoints(one, &pb)
	}
	Traverse(func(s *Skill) bool {
		pb.Skills += s.Points
		return false
	}, false, true, e.Skills...)
	Traverse(func(s *Spell) bool {
		pb.Spells += s.Points
		return false
	}, false, true, e.Spells...)
	return &pb
}

func calculateSingleTraitPoints(t *Trait, pb *PointsBreakdown) {
	if t.Disabled {
		return
	}
	if t.Container() {
		switch t.ContainerType {
		case container.Group:
			for _, child := range t.Children {
				calculateSingleTraitPoints(child, pb)
			}
			return
		case container.Ancestry:
			pb.Ancestry += t.AdjustedPoints()
			return
		case container.Attributes:
			pb.Attributes += t.AdjustedPoints()
			return
		default:
		}
	}
	pts := t.AdjustedPoints()
	switch {
	case pts == -fxp.One:
		pb.Quirks += pts
	case pts > 0:
		pb.Advantages += pts
	case pts < 0:
		pb.Disadvantages += pts
	}
}

// WealthCarried returns the current wealth being carried.
func (e *Entity) WealthCarried() fxp.Int {
	var value fxp.Int
	for _, one := range e.CarriedEquipment {
		value += one.ExtendedValue()
	}
	return value
}

// WealthNotCarried returns the current wealth not being carried.
func (e *Entity) WealthNotCarried() fxp.Int {
	var value fxp.Int
	for _, one := range e.OtherEquipment {
		value += one.ExtendedValue()
	}
	return value
}

// StrikingStrength returns the adjusted ST for striking purposes.
func (e *Entity) StrikingStrength() fxp.Int {
	var st fxp.Int
	if e.ResolveAttribute(StrikingStrengthID) != nil {
		st = e.ResolveAttributeCurrent(StrikingStrengthID)
	} else {
		st = e.ResolveAttributeCurrent(StrengthID).Max(0)
	}
	st += e.StrikingStrengthBonus
	return st.Trunc()
}

// LiftingStrength returns the adjusted ST for lifting purposes.
func (e *Entity) LiftingStrength() fxp.Int {
	var st fxp.Int
	if e.ResolveAttribute(LiftingStrengthID) != nil {
		st = e.ResolveAttributeCurrent(LiftingStrengthID)
	} else {
		st = e.ResolveAttributeCurrent(StrengthID).Max(0)
	}
	st += e.LiftingStrengthBonus
	return st.Trunc()
}

// ThrowingStrength returns the adjusted ST for throwing purposes.
func (e *Entity) ThrowingStrength() fxp.Int {
	var st fxp.Int
	if e.ResolveAttribute(ThrowingStrengthID) != nil {
		st = e.ResolveAttributeCurrent(ThrowingStrengthID)
	} else {
		st = e.ResolveAttributeCurrent(StrengthID).Max(0)
	}
	st += e.ThrowingStrengthBonus
	return st.Trunc()
}

// TelekineticStrength returns the total telekinetic strength.
func (e *Entity) TelekineticStrength() fxp.Int {
	var levels fxp.Int
	Traverse(func(a *Trait) bool {
		if !a.Container() && a.IsLeveled() {
			if strings.EqualFold(a.NameWithReplacements(), "telekinesis") {
				levels += a.Levels.Max(0)
			}
		}
		return false
	}, true, false, e.Traits...)
	return levels.Trunc()
}

// Thrust returns the thrust value for the current strength.
func (e *Entity) Thrust() *dice.Dice {
	return e.ThrustFor(fxp.As[int](e.StrikingStrength()))
}

// LiftingThrust returns the lifting thrust value for the current strength.
func (e *Entity) LiftingThrust() *dice.Dice {
	return e.ThrustFor(fxp.As[int](e.LiftingStrength()))
}

// IQThrust returns the IQ thrust value for the current intelligence.
func (e *Entity) IQThrust() *dice.Dice {
	return e.ThrustFor(fxp.As[int](e.ResolveAttributeCurrent(IntelligenceID)))
}

// TelekineticThrust returns the telekinetic thrust value for the current telekinesis level.
func (e *Entity) TelekineticThrust() *dice.Dice {
	return e.ThrustFor(fxp.As[int](e.TelekineticStrength()))
}

// ThrustFor returns the thrust value for the provided strength.
func (e *Entity) ThrustFor(st int) *dice.Dice {
	return e.SheetSettings.DamageProgression.Thrust(st)
}

// Swing returns the swing value for the current strength.
func (e *Entity) Swing() *dice.Dice {
	return e.SwingFor(fxp.As[int](e.StrikingStrength()))
}

// LiftingSwing returns the lifting swing value for the current strength.
func (e *Entity) LiftingSwing() *dice.Dice {
	return e.SwingFor(fxp.As[int](e.LiftingStrength()))
}

// IQSwing returns the IQ swing value for the current intelligence.
func (e *Entity) IQSwing() *dice.Dice {
	return e.SwingFor(fxp.As[int](e.ResolveAttributeCurrent(IntelligenceID)))
}

// TelekineticSwing returns the telekinetic swing value for the current telekinesis level.
func (e *Entity) TelekineticSwing() *dice.Dice {
	return e.SwingFor(fxp.As[int](e.TelekineticStrength()))
}

// SwingFor returns the swing value for the provided strength.
func (e *Entity) SwingFor(st int) *dice.Dice {
	return e.SheetSettings.DamageProgression.Swing(st)
}

// AttributeBonusFor returns the bonus for the given attribute.
func (e *Entity) AttributeBonusFor(attributeID string, limitation stlimit.Option, tooltip *xio.ByteBuffer) fxp.Int {
	var total fxp.Int
	for _, one := range e.features.attributeBonuses {
		if one.ActualLimitation() == limitation && one.Attribute == attributeID {
			total += one.AdjustedAmount()
			one.AddToTooltip(tooltip)
		}
	}
	return total
}

// CostReductionFor returns the total cost reduction for the given ID.
func (e *Entity) CostReductionFor(attributeID string) fxp.Int {
	var total fxp.Int
	for _, one := range e.features.costReductions {
		if one.Attribute == attributeID {
			total += one.Percentage
		}
	}
	if total > fxp.Eighty {
		total = fxp.Eighty
	}
	return total.Max(0)
}

// AddDRBonusesFor locates any active DR bonuses and adds them to the map. If 'drMap' is nil, it will be created. The
// provided map (or the newly created one) will be returned.
func (e *Entity) AddDRBonusesFor(locationID string, tooltip *xio.ByteBuffer, drMap map[string]int) map[string]int {
	if drMap == nil {
		drMap = make(map[string]int)
	}
	isTopLevel := false
	for _, one := range e.SheetSettings.BodyType.Locations {
		if one.LocID == locationID {
			isTopLevel = true
			break
		}
	}
	for _, one := range e.features.drBonuses {
		for _, loc := range one.Locations {
			if (loc == AllID && isTopLevel) || strings.EqualFold(loc, locationID) {
				drMap[strings.ToLower(one.Specialization)] += fxp.As[int](one.AdjustedAmount())
				one.AddToTooltip(tooltip)
				break
			}
		}
	}
	return drMap
}

// SkillBonusFor returns the total bonus for the matching skill bonuses.
func (e *Entity) SkillBonusFor(name, specialization string, tags []string, tooltip *xio.ByteBuffer) fxp.Int {
	var total fxp.Int
	for _, bonus := range e.features.skillBonuses {
		if bonus.SelectionType == skillsel.Name {
			var replacements map[string]string
			if na, ok := bonus.Owner().(nameable.Accesser); ok {
				replacements = na.NameableReplacements()
			}
			if bonus.NameCriteria.Matches(replacements, name) &&
				bonus.SpecializationCriteria.Matches(replacements, specialization) &&
				bonus.TagsCriteria.MatchesList(replacements, tags...) {
				total += bonus.AdjustedAmount()
				bonus.AddToTooltip(tooltip)
			}
		}
	}
	return total
}

// SkillPointBonusFor returns the total point bonus for the matching skill point bonuses.
func (e *Entity) SkillPointBonusFor(name, specialization string, tags []string, tooltip *xio.ByteBuffer) fxp.Int {
	var total fxp.Int
	for _, bonus := range e.features.skillPointBonuses {
		var replacements map[string]string
		if na, ok := bonus.Owner().(nameable.Accesser); ok {
			replacements = na.NameableReplacements()
		}
		if bonus.NameCriteria.Matches(replacements, name) &&
			bonus.SpecializationCriteria.Matches(replacements, specialization) &&
			bonus.TagsCriteria.MatchesList(replacements, tags...) {
			total += bonus.AdjustedAmount()
			bonus.AddToTooltip(tooltip)
		}
	}
	return total
}

// SpellBonusFor returns the total bonus for the matching spell bonuses.
func (e *Entity) SpellBonusFor(name, powerSource string, colleges, tags []string, tooltip *xio.ByteBuffer) fxp.Int {
	var total fxp.Int
	for _, bonus := range e.features.spellBonuses {
		var replacements map[string]string
		if na, ok := bonus.Owner().(nameable.Accesser); ok {
			replacements = na.NameableReplacements()
		}
		if bonus.TagsCriteria.MatchesList(replacements, tags...) &&
			bonus.MatchForType(replacements, name, powerSource, colleges) {
			total += bonus.AdjustedAmount()
			bonus.AddToTooltip(tooltip)
		}
	}
	return total
}

// SpellPointBonusFor returns the total point bonus for the matching spell point bonuses.
func (e *Entity) SpellPointBonusFor(name, powerSource string, colleges, tags []string, tooltip *xio.ByteBuffer) fxp.Int {
	var total fxp.Int
	for _, bonus := range e.features.spellPointBonuses {
		var replacements map[string]string
		if na, ok := bonus.Owner().(nameable.Accesser); ok {
			replacements = na.NameableReplacements()
		}
		if bonus.TagsCriteria.MatchesList(replacements, tags...) &&
			bonus.MatchForType(replacements, name, powerSource, colleges) {
			total += bonus.AdjustedAmount()
			bonus.AddToTooltip(tooltip)
		}
	}
	return total
}

// AddWeaponWithSkillBonusesFor adds the bonuses for matching weapons that match to the map. If 'm' is nil, it will be
// created. The provided map (or the newly created one) will be returned.
func (e *Entity) AddWeaponWithSkillBonusesFor(name, specialization, usage string, tags []string, dieCount int, tooltip *xio.ByteBuffer, m map[*WeaponBonus]bool, allowedFeatureTypes map[feature.Type]bool) map[*WeaponBonus]bool {
	if m == nil {
		m = make(map[*WeaponBonus]bool)
	}
	rsl := fxp.Min
	for _, sk := range e.SkillNamed(name, specialization, true, nil) {
		if rsl < sk.LevelData.RelativeLevel {
			rsl = sk.LevelData.RelativeLevel
		}
	}
	for _, bonus := range e.features.weaponBonuses {
		if allowedFeatureTypes[bonus.Type] &&
			bonus.SelectionType == wsel.WithRequiredSkill &&
			bonus.RelativeLevelCriteria.Matches(rsl) {
			var replacements map[string]string
			if na, ok := bonus.Owner().(nameable.Accesser); ok {
				replacements = na.NameableReplacements()
			}
			if bonus.NameCriteria.Matches(replacements, name) &&
				bonus.SpecializationCriteria.Matches(replacements, specialization) &&
				bonus.UsageCriteria.Matches(replacements, usage) &&
				bonus.TagsCriteria.MatchesList(replacements, tags...) {
				addWeaponBonusToMap(bonus, dieCount, tooltip, m)
			}
		}
	}
	return m
}

// AddNamedWeaponBonusesFor adds the bonuses for matching weapons that match to the map. If 'm' is nil, it will
// be created. The provided map (or the newly created one) will be returned.
func (e *Entity) AddNamedWeaponBonusesFor(nameQualifier, usageQualifier string, tagsQualifier []string, dieCount int, tooltip *xio.ByteBuffer, m map[*WeaponBonus]bool, allowedFeatureTypes map[feature.Type]bool) map[*WeaponBonus]bool {
	if m == nil {
		m = make(map[*WeaponBonus]bool)
	}
	for _, bonus := range e.features.weaponBonuses {
		if allowedFeatureTypes[bonus.Type] &&
			bonus.SelectionType == wsel.WithName {
			var replacements map[string]string
			if na, ok := bonus.Owner().(nameable.Accesser); ok {
				replacements = na.NameableReplacements()
			}
			if bonus.NameCriteria.Matches(replacements, nameQualifier) &&
				bonus.SpecializationCriteria.Matches(replacements, usageQualifier) &&
				bonus.TagsCriteria.MatchesList(replacements, tagsQualifier...) {
				addWeaponBonusToMap(bonus, dieCount, tooltip, m)
			}
		}
	}
	return m
}

func addWeaponBonusToMap(bonus *WeaponBonus, dieCount int, tooltip *xio.ByteBuffer, m map[*WeaponBonus]bool) {
	savedLevel := bonus.Level
	savedDieCount := bonus.DieCount
	bonus.DieCount = fxp.From(dieCount)
	bonus.Level = bonus.DerivedLevel()
	bonus.AddToTooltip(tooltip)
	bonus.Level = savedLevel
	bonus.DieCount = savedDieCount
	m[bonus] = true
}

// NamedWeaponSkillBonusesFor returns the bonuses for matching weapons.
func (e *Entity) NamedWeaponSkillBonusesFor(name, usage string, tags []string, tooltip *xio.ByteBuffer) []*SkillBonus {
	var bonuses []*SkillBonus
	for _, bonus := range e.features.skillBonuses {
		if bonus.SelectionType == skillsel.WeaponsWithName {
			var replacements map[string]string
			if na, ok := bonus.Owner().(nameable.Accesser); ok {
				replacements = na.NameableReplacements()
			}
			if bonus.NameCriteria.Matches(replacements, name) &&
				bonus.SpecializationCriteria.Matches(replacements, usage) &&
				bonus.TagsCriteria.MatchesList(replacements, tags...) {
				bonuses = append(bonuses, bonus)
				bonus.AddToTooltip(tooltip)
			}
		}
	}
	return bonuses
}

// Move returns the current Move value for the given Encumbrance.
func (e *Entity) Move(enc encumbrance.Level) int {
	var initialMove fxp.Int
	if e.ResolveAttribute(MoveID) != nil {
		initialMove = e.ResolveAttributeCurrent(MoveID)
	} else {
		initialMove = e.ResolveAttributeCurrent(BasicMoveID).Max(0)
	}
	if divisor := 2 * min(CountThresholdOpMet(threshold.HalveMove, e.Attributes), 2); divisor > 0 {
		initialMove = initialMove.Div(fxp.From(divisor)).Ceil()
	}
	move := initialMove.Mul(fxp.Ten + fxp.Two.Mul(enc.Penalty())).Div(fxp.Ten).Trunc()
	if move < fxp.One {
		if initialMove > 0 {
			return 1
		}
		return 0
	}
	return fxp.As[int](move)
}

// BestSkillNamed returns the best skill that matches.
func (e *Entity) BestSkillNamed(name, specialization string, requirePoints bool, excludes map[string]bool) *Skill {
	var best *Skill
	level := fxp.Min
	for _, sk := range e.SkillNamed(name, specialization, requirePoints, excludes) {
		skillLevel := sk.CalculateLevel(excludes).Level
		if best == nil || level < skillLevel {
			best = sk
			level = skillLevel
		}
	}
	return best
}

// SkillNamed returns a list of skills that match.
func (e *Entity) SkillNamed(name, specialization string, requirePoints bool, excludes map[string]bool) []*Skill {
	var list []*Skill
	Traverse(func(sk *Skill) bool {
		if !excludes[sk.String()] {
			if !requirePoints || sk.IsTechnique() || sk.AdjustedPoints(nil) > 0 {
				if strings.EqualFold(sk.NameWithReplacements(), name) {
					if specialization == "" || strings.EqualFold(sk.SpecializationWithReplacements(), specialization) {
						list = append(list, sk)
					}
				}
			}
		}
		return false
	}, false, true, e.Skills...)
	return list
}

// Dodge returns the current Dodge value for the given Encumbrance.
func (e *Entity) Dodge(enc encumbrance.Level) int {
	var dodge fxp.Int
	if e.ResolveAttribute(DodgeID) != nil {
		dodge = e.ResolveAttributeCurrent(DodgeID)
	} else {
		dodge = e.ResolveAttributeCurrent(BasicSpeedID).Max(0) + fxp.Three
	}
	dodge += e.DodgeBonus
	divisor := 2 * min(CountThresholdOpMet(threshold.HalveDodge, e.Attributes), 2)
	if divisor > 0 {
		dodge = dodge.Div(fxp.From(divisor)).Ceil()
	}
	return fxp.As[int]((dodge + enc.Penalty()).Max(fxp.One))
}

// EncumbranceLevel returns the current Encumbrance level.
func (e *Entity) EncumbranceLevel(forSkills bool) encumbrance.Level {
	if forSkills {
		if e.encumbranceLevelForSkillsCache != encumbrance.LastLevel+1 {
			return e.encumbranceLevelForSkillsCache
		}
	} else if e.encumbranceLevelCache != encumbrance.LastLevel+1 {
		return e.encumbranceLevelCache
	}
	carried := e.WeightCarried(forSkills)
	for _, one := range encumbrance.Levels {
		if carried <= e.MaximumCarry(one) {
			if forSkills {
				e.encumbranceLevelForSkillsCache = one
			} else {
				e.encumbranceLevelCache = one
			}
			return one
		}
	}
	if forSkills {
		e.encumbranceLevelForSkillsCache = encumbrance.ExtraHeavy
	} else {
		e.encumbranceLevelCache = encumbrance.ExtraHeavy
	}
	return encumbrance.ExtraHeavy
}

// WeightUnit returns the weight unit that should be used for display.
func (e *Entity) WeightUnit() fxp.WeightUnit {
	return e.SheetSettings.DefaultWeightUnits
}

// WeightCarried returns the carried weight.
func (e *Entity) WeightCarried(forSkills bool) fxp.Weight {
	var total fxp.Weight
	for _, one := range e.CarriedEquipment {
		total += one.ExtendedWeight(forSkills, e.SheetSettings.DefaultWeightUnits)
	}
	return total
}

// MaximumCarry returns the maximum amount the Entity can carry for the specified encumbrance level.
func (e *Entity) MaximumCarry(enc encumbrance.Level) fxp.Weight {
	return fxp.Weight(fxp.Int(e.BasicLift()).Mul(enc.WeightMultiplier()))
}

// OneHandedLift returns the one-handed lift value.
func (e *Entity) OneHandedLift() fxp.Weight {
	return fxp.Weight(fxp.Int(e.BasicLift()).Mul(fxp.Two))
}

// TwoHandedLift returns the two-handed lift value.
func (e *Entity) TwoHandedLift() fxp.Weight {
	return fxp.Weight(fxp.Int(e.BasicLift()).Mul(fxp.Eight))
}

// ShoveAndKnockOver returns the shove & knock over value.
func (e *Entity) ShoveAndKnockOver() fxp.Weight {
	return fxp.Weight(fxp.Int(e.BasicLift()).Mul(fxp.Twelve))
}

// RunningShoveAndKnockOver returns the running shove & knock over value.
func (e *Entity) RunningShoveAndKnockOver() fxp.Weight {
	return fxp.Weight(fxp.Int(e.BasicLift()).Mul(fxp.TwentyFour))
}

// CarryOnBack returns the carry on back value.
func (e *Entity) CarryOnBack() fxp.Weight {
	return fxp.Weight(fxp.Int(e.BasicLift()).Mul(fxp.Fifteen))
}

// ShiftSlightly returns the shift slightly value.
func (e *Entity) ShiftSlightly() fxp.Weight {
	return fxp.Weight(fxp.Int(e.BasicLift()).Mul(fxp.Fifty))
}

// BasicLift returns the entity's Basic Lift.
func (e *Entity) BasicLift() fxp.Weight {
	if e.basicLiftCache != -1 {
		return e.basicLiftCache
	}
	e.basicLiftCache = e.BasicLiftForST(e.LiftingStrength())
	return e.basicLiftCache
}

// BasicLiftForST returns the entity's Basic Lift as if their base ST was the given value.
func (e *Entity) BasicLiftForST(st fxp.Int) fxp.Weight {
	st = st.Trunc()
	if IsThresholdOpMet(threshold.HalveST, e.Attributes) {
		st = st.Div(fxp.Two)
		if st != st.Trunc() {
			st = st.Trunc() + fxp.One
		}
	}
	if st < fxp.One {
		return 0
	}
	var v fxp.Int
	if e.SheetSettings.DamageProgression == progression.KnowingYourOwnStrength {
		var diff fxp.Int
		if st > fxp.Nineteen {
			diff = st.Div(fxp.Ten).Trunc() - fxp.One
			st -= diff.Mul(fxp.Ten)
		}
		v = fxp.From(math.Pow(10, fxp.As[float64](st)/10)).Mul(fxp.Two)
		if st <= fxp.Six {
			v = v.Mul(fxp.Ten).Round().Div(fxp.Ten)
		} else {
			v = v.Round()
		}
		v = v.Mul(fxp.From(math.Pow(10, fxp.As[float64](diff))))
	} else {
		v = st.Mul(st).Div(fxp.Five)
	}
	if v >= fxp.Ten {
		v = v.Round()
	}
	return fxp.Weight(v.Mul(fxp.Ten).Trunc().Div(fxp.Ten))
}

func (e *Entity) isSkillLevelResolutionExcluded(name, specialization string) bool {
	if e.skillResolverExclusions[e.skillLevelResolutionKey(name, specialization)] {
		args := []any{"name", name}
		if specialization != "" {
			args = append(args, "specialization", specialization)
		}
		slog.Error("attempt to resolve skill level via itself", args...)
		return true
	}
	return false
}

func (e *Entity) registerSkillLevelResolutionExclusion(name, specialization string) {
	e.skillResolverExclusions[e.skillLevelResolutionKey(name, specialization)] = true
}

func (e *Entity) unregisterSkillLevelResolutionExclusion(name, specialization string) {
	delete(e.skillResolverExclusions, e.skillLevelResolutionKey(name, specialization))
}

func (e *Entity) skillLevelResolutionKey(name, specialization string) string {
	return name + "\u0000" + specialization
}

// ResolveVariable implements eval.VariableResolver.
func (e *Entity) ResolveVariable(variableName string) string {
	if e.variableResolverExclusions[variableName] {
		slog.Error("attempt to resolve variable via itself", "name", variableName)
		return ""
	}
	if v, ok := e.variableCache[variableName]; ok {
		return v
	}
	e.variableResolverExclusions[variableName] = true
	defer func() { delete(e.variableResolverExclusions, variableName) }()
	if SizeModifierID == variableName {
		result := strconv.Itoa(e.Profile.AdjustedSizeModifier())
		e.variableCache[variableName] = result
		return result
	}
	parts := strings.SplitN(variableName, ".", 2)
	attr := e.Attributes.Set[parts[0]]
	if attr == nil {
		slog.Error("no such variable", "name", variableName)
		return ""
	}
	def := attr.AttributeDef()
	if def == nil {
		slog.Error("no such variable definition", "name", variableName)
		return ""
	}
	if (def.Type == attribute.Pool || def.Type == attribute.PoolRef) && len(parts) > 1 && parts[1] == "current" {
		result := attr.Current().String()
		e.variableCache[variableName] = result
		return result
	}
	result := attr.Maximum().String()
	e.variableCache[variableName] = result
	return result
}

// ResolveAttributeDef resolves the given attribute ID to its AttributeDef, or nil.
func (e *Entity) ResolveAttributeDef(attrID string) *AttributeDef {
	if e != nil {
		if a, ok := e.Attributes.Set[attrID]; ok {
			return a.AttributeDef()
		}
	}
	return nil
}

// ResolveAttributeName resolves the given attribute ID to its name, or <unknown>.
func (e *Entity) ResolveAttributeName(attrID string) string {
	if def := e.ResolveAttributeDef(attrID); def != nil {
		return def.Name
	}
	return i18n.Text("<unknown>")
}

// ResolveAttribute resolves the given attribute ID to its Attribute, or nil.
func (e *Entity) ResolveAttribute(attrID string) *Attribute {
	if e != nil {
		if a, ok := e.Attributes.Set[attrID]; ok {
			return a
		}
	}
	return nil
}

// ResolveAttributeCurrent resolves the given attribute ID to its current value, or fxp.Min.
func (e *Entity) ResolveAttributeCurrent(attrID string) fxp.Int {
	if e != nil {
		return e.Attributes.Current(attrID)
	}
	return fxp.Min
}

// PreservesUserDesc returns true if the user description widget should be preserved when written to disk. Normally,
// only character sheets should return true for this.
func (e *Entity) PreservesUserDesc() bool {
	return true
}

// Ancestry returns the current Ancestry.
func (e *Entity) Ancestry() *Ancestry {
	var anc *Ancestry
	Traverse(func(t *Trait) bool {
		if t.Container() && t.ContainerType == container.Ancestry && t.Enabled() {
			if anc = LookupAncestry(t.Ancestry, GlobalSettings().Libraries()); anc != nil {
				return true
			}
		}
		return false
	}, true, false, e.Traits...)
	if anc == nil {
		if anc = LookupAncestry(DefaultAncestry, GlobalSettings().Libraries()); anc == nil {
			fatal.IfErr(errs.New("unable to load default ancestry (Human)"))
		}
	}
	return anc
}

// WeaponOwner implements WeaponListProvider. In the case of an Entity, always returns nil, as entities rely on
// sub-components for their weapons and don't allow them to be created directly.
func (e *Entity) WeaponOwner() WeaponOwner {
	return nil
}

// Weapons implements WeaponListProvider.
func (e *Entity) Weapons(melee, excludeHidden bool) []*Weapon {
	return e.EquippedWeapons(melee, excludeHidden)
}

// SetWeapons implements WeaponListProvider.
func (e *Entity) SetWeapons(_ bool, _ []*Weapon) {
	// Not permitted
}

// EquippedWeapons returns a sorted list of equipped weapons.
func (e *Entity) EquippedWeapons(melee, excludeHidden bool) []*Weapon {
	m := make(map[uint64]*Weapon)
	Traverse(func(a *Trait) bool {
		for _, w := range a.Weapons {
			if w.IsMelee() == melee && (!excludeHidden || !w.Hide) {
				m[w.HashResolved()] = w
			}
		}
		return false
	}, true, true, e.Traits...)
	Traverse(func(eqp *Equipment) bool {
		if eqp.Equipped {
			for _, w := range eqp.Weapons {
				if w.IsMelee() == melee && (!excludeHidden || !w.Hide) {
					m[w.HashResolved()] = w
				}
			}
		}
		return false
	}, false, false, e.CarriedEquipment...)
	Traverse(func(s *Skill) bool {
		for _, w := range s.Weapons {
			if w.IsMelee() == melee && (!excludeHidden || !w.Hide) {
				m[w.HashResolved()] = w
			}
		}
		return false
	}, false, true, e.Skills...)
	Traverse(func(s *Spell) bool {
		for _, w := range s.Weapons {
			if w.IsMelee() == melee && (!excludeHidden || !w.Hide) {
				m[w.HashResolved()] = w
			}
		}
		return false
	}, false, true, e.Spells...)
	list := make([]*Weapon, 0, len(m))
	for _, v := range m {
		list = append(list, v)
	}
	slices.SortFunc(list, func(a, b *Weapon) int { return a.Compare(b) })
	return list
}

// Reactions returns the current set of reactions.
func (e *Entity) Reactions() []*ConditionalModifier {
	m := make(map[string]*ConditionalModifier)
	Traverse(func(a *Trait) bool {
		source := i18n.Text("from trait ") + a.String()
		if !a.Container() {
			e.reactionsFromFeatureList(source, a.Features, m)
		}
		Traverse(func(mod *TraitModifier) bool {
			e.reactionsFromFeatureList(source, mod.Features, m)
			return false
		}, true, true, a.Modifiers...)
		if a.CR != selfctrl.NoCR && a.CRAdj == selfctrl.ReactionPenalty {
			amt := fxp.From(selfctrl.ReactionPenalty.Adjustment(a.CR))
			situation := fmt.Sprintf(i18n.Text("from others when %s is triggered"), a.String())
			if r, exists := m[situation]; exists {
				r.Add(source, amt)
			} else {
				m[situation] = NewConditionalModifier(source, situation, amt)
			}
		}
		return false
	}, true, false, e.Traits...)
	Traverse(func(eqp *Equipment) bool {
		if eqp.Equipped && eqp.Quantity > 0 {
			source := i18n.Text("from equipment ") + eqp.NameWithReplacements()
			e.reactionsFromFeatureList(source, eqp.Features, m)
			Traverse(func(mod *EquipmentModifier) bool {
				e.reactionsFromFeatureList(source, mod.Features, m)
				return false
			}, true, true, eqp.Modifiers...)
		}
		return false
	}, false, false, e.CarriedEquipment...)
	Traverse(func(sk *Skill) bool {
		e.reactionsFromFeatureList(i18n.Text("from skill ")+sk.String(), sk.Features, m)
		return false
	}, false, true, e.Skills...)
	list := make([]*ConditionalModifier, 0, len(m))
	for _, v := range m {
		list = append(list, v)
	}
	slices.SortFunc(list, func(a, b *ConditionalModifier) int { return a.Compare(b) })
	return list
}

func (e *Entity) reactionsFromFeatureList(source string, features Features, m map[string]*ConditionalModifier) {
	for _, f := range features {
		bonus, ok := f.(*ReactionBonus)
		if !ok {
			continue
		}
		amt := bonus.AdjustedAmount()
		var replacements map[string]string
		var na nameable.Accesser
		if na, ok = bonus.Owner().(nameable.Accesser); ok {
			replacements = na.NameableReplacements()
		}
		situation := nameable.Apply(bonus.Situation, replacements)
		if r, exists := m[situation]; exists {
			r.Add(source, amt)
		} else {
			m[situation] = NewConditionalModifier(source, situation, amt)
		}
	}
}

// ConditionalModifiers returns the current set of conditional modifiers.
func (e *Entity) ConditionalModifiers() []*ConditionalModifier {
	m := make(map[string]*ConditionalModifier)
	Traverse(func(a *Trait) bool {
		source := i18n.Text("from trait ") + a.String()
		if !a.Container() {
			e.conditionalModifiersFromFeatureList(source, a.Features, m)
		}
		Traverse(func(mod *TraitModifier) bool {
			e.conditionalModifiersFromFeatureList(source, mod.Features, m)
			return false
		}, true, true, a.Modifiers...)
		return false
	}, true, false, e.Traits...)
	Traverse(func(eqp *Equipment) bool {
		if eqp.Equipped && eqp.Quantity > 0 {
			source := i18n.Text("from equipment ") + eqp.NameWithReplacements()
			e.conditionalModifiersFromFeatureList(source, eqp.Features, m)
			Traverse(func(mod *EquipmentModifier) bool {
				e.conditionalModifiersFromFeatureList(source, mod.Features, m)
				return false
			}, true, true, eqp.Modifiers...)
		}
		return false
	}, false, false, e.CarriedEquipment...)
	Traverse(func(sk *Skill) bool {
		e.conditionalModifiersFromFeatureList(i18n.Text("from skill ")+sk.String(), sk.Features, m)
		return false
	}, false, true, e.Skills...)
	list := make([]*ConditionalModifier, 0, len(m))
	for _, v := range m {
		list = append(list, v)
	}
	slices.SortFunc(list, func(a, b *ConditionalModifier) int { return a.Compare(b) })
	return list
}

func (e *Entity) conditionalModifiersFromFeatureList(source string, features Features, m map[string]*ConditionalModifier) {
	for _, f := range features {
		bonus, ok := f.(*ConditionalModifierBonus)
		if !ok {
			continue
		}
		amt := bonus.AdjustedAmount()
		var replacements map[string]string
		var na nameable.Accesser
		if na, ok = bonus.Owner().(nameable.Accesser); ok {
			replacements = na.NameableReplacements()
		}
		situation := nameable.Apply(bonus.Situation, replacements)
		if r, exists := m[situation]; exists {
			r.Add(source, amt)
		} else {
			m[situation] = NewConditionalModifier(source, situation, amt)
		}
	}
}

// TraitList implements ListProvider
func (e *Entity) TraitList() []*Trait {
	return e.Traits
}

// SetTraitList implements ListProvider
func (e *Entity) SetTraitList(list []*Trait) {
	for _, one := range list {
		one.SetDataOwner(e)
	}
	e.Traits = list
}

// CarriedEquipmentList implements ListProvider
func (e *Entity) CarriedEquipmentList() []*Equipment {
	return e.CarriedEquipment
}

// SetCarriedEquipmentList implements ListProvider
func (e *Entity) SetCarriedEquipmentList(list []*Equipment) {
	for _, one := range list {
		one.SetDataOwner(e)
	}
	e.CarriedEquipment = list
}

// OtherEquipmentList implements ListProvider
func (e *Entity) OtherEquipmentList() []*Equipment {
	return e.OtherEquipment
}

// SetOtherEquipmentList implements ListProvider
func (e *Entity) SetOtherEquipmentList(list []*Equipment) {
	for _, one := range list {
		one.SetDataOwner(e)
	}
	e.OtherEquipment = list
}

// SkillList implements ListProvider
func (e *Entity) SkillList() []*Skill {
	return e.Skills
}

// SetSkillList implements ListProvider
func (e *Entity) SetSkillList(list []*Skill) {
	for _, one := range list {
		one.SetDataOwner(e)
	}
	e.Skills = list
}

// SpellList implements ListProvider
func (e *Entity) SpellList() []*Spell {
	return e.Spells
}

// SetSpellList implements ListProvider
func (e *Entity) SetSpellList(list []*Spell) {
	for _, one := range list {
		one.SetDataOwner(e)
	}
	e.Spells = list
}

// NoteList implements ListProvider
func (e *Entity) NoteList() []*Note {
	return e.Notes
}

// SetNoteList implements ListProvider
func (e *Entity) SetNoteList(list []*Note) {
	for _, one := range list {
		one.SetDataOwner(e)
	}
	e.Notes = list
}

// Hash writes this object's contents into the hasher.
func (e *Entity) Hash(h hash.Hash) {
	var buffer bytes.Buffer
	saved := e.ModifiedOn
	e.ModifiedOn = jio.Time{}
	defer func() { e.ModifiedOn = saved }()
	if err := jio.Save(context.Background(), &buffer, e); err != nil {
		errs.Log(err)
		return
	}
	_, _ = h.Write(buffer.Bytes())
}

// SetPointsRecord sets a new points record list, adjusting the total points.
func (e *Entity) SetPointsRecord(record []*PointsRecord) {
	e.PointsRecord = ClonePointsRecordList(record)
	slices.SortFunc(e.PointsRecord, func(a, b *PointsRecord) int { return b.When.Compare(a.When) })
	e.TotalPoints = 0
	for _, rec := range record {
		e.TotalPoints += rec.Points
	}
}

// SyncWithLibrarySources syncs the entity with the library sources.
func (e *Entity) SyncWithLibrarySources() {
	Traverse(func(trait *Trait) bool {
		trait.SyncWithSource()
		Traverse(func(traitModifier *TraitModifier) bool {
			traitModifier.SyncWithSource()
			return false
		}, false, false, trait.Modifiers...)
		return false
	}, false, false, e.Traits...)
	Traverse(func(skill *Skill) bool {
		skill.SyncWithSource()
		return false
	}, false, false, e.Skills...)
	Traverse(func(spell *Spell) bool {
		spell.SyncWithSource()
		return false
	}, false, false, e.Spells...)
	Traverse(func(equipment *Equipment) bool {
		equipment.SyncWithSource()
		Traverse(func(equipmentModifier *EquipmentModifier) bool {
			equipmentModifier.SyncWithSource()
			return false
		}, false, false, equipment.Modifiers...)
		return false
	}, false, false, e.CarriedEquipment...)
	Traverse(func(equipment *Equipment) bool {
		equipment.SyncWithSource()
		Traverse(func(equipmentModifier *EquipmentModifier) bool {
			equipmentModifier.SyncWithSource()
			return false
		}, false, false, equipment.Modifiers...)
		return false
	}, false, false, e.OtherEquipment...)
	Traverse(func(note *Note) bool {
		note.SyncWithSource()
		return false
	}, false, false, e.Notes...)
}

// PageSettings implements PageInfoProvider.
func (e *Entity) PageSettings() *PageSettings {
	return e.SheetSettings.Page
}

// PageTitle implements PageInfoProvider.
func (e *Entity) PageTitle() string {
	if e.SheetSettings.UseTitleInFooter {
		return e.Profile.Title
	}
	return e.Profile.Name
}

// ModifiedOnString implements PageInfoProvider.
func (e *Entity) ModifiedOnString() string {
	return e.ModifiedOn.String()
}
