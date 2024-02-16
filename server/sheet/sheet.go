/*
 * Copyright ©1998-2024 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package sheet

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

type Identity struct {
	Name         string
	Title        string
	Organization string
}

func createIdentity(entity *gurps.Entity) Identity {
	return Identity{
		Name:         entity.Profile.Name,
		Title:        entity.Profile.Title,
		Organization: entity.Profile.Organization,
	}
}

type Misc struct {
	Created  jio.Time
	Modified jio.Time
	Player   string
}

func createMisc(entity *gurps.Entity) Misc {
	return Misc{
		Created:  entity.CreatedOn,
		Modified: entity.ModifiedOn,
		Player:   entity.Profile.PlayerName,
	}
}

type Description struct {
	Gender       string
	Age          string
	Birthday     string
	Religion     string
	Height       fxp.Length
	Weight       fxp.Weight
	SizeModifier int
	TechLevel    string
	Hair         string
	Eyes         string
	Skin         string
	Hand         string
}

func createDescription(entity *gurps.Entity) Description {
	return Description{
		Gender:       entity.Profile.Gender,
		Age:          entity.Profile.Age,
		Birthday:     entity.Profile.Birthday,
		Religion:     entity.Profile.Religion,
		Height:       entity.Profile.Height,
		Weight:       entity.Profile.Weight,
		SizeModifier: entity.Profile.SizeModifier,
		TechLevel:    entity.Profile.TechLevel,
		Hair:         entity.Profile.Hair,
		Eyes:         entity.Profile.Eyes,
		Skin:         entity.Profile.Skin,
		Hand:         entity.Profile.Handedness,
	}
}

type Points struct {
	Total         fxp.Int
	Unspent       fxp.Int
	Ancestry      fxp.Int
	Attributes    fxp.Int
	Advantages    fxp.Int
	Disadvantages fxp.Int
	Quirks        fxp.Int
	Skills        fxp.Int
	Spells        fxp.Int
}

func createPoints(entity *gurps.Entity) Points {
	var totalPoints fxp.Int
	pointsBreakdown := entity.PointsBreakdown()
	if entity.SheetSettings.ExcludeUnspentPointsFromTotal {
		totalPoints = pointsBreakdown.Total()
	} else {
		totalPoints = entity.TotalPoints
	}
	return Points{
		Total:         totalPoints,
		Unspent:       entity.UnspentPoints(),
		Ancestry:      pointsBreakdown.Ancestry,
		Attributes:    pointsBreakdown.Attributes,
		Advantages:    pointsBreakdown.Advantages,
		Disadvantages: pointsBreakdown.Disadvantages,
		Quirks:        pointsBreakdown.Quirks,
		Skills:        pointsBreakdown.Skills,
		Spells:        pointsBreakdown.Spells,
	}
}

type Attribute struct {
	Type   attribute.Type
	Key    string
	Name   string
	Value  fxp.Int
	Points fxp.Int
}

func createPrimaryAttributes(entity *gurps.Entity) []Attribute {
	var list []Attribute
	for _, def := range gurps.SheetSettingsFor(entity).Attributes.List(false) {
		if def.Primary() {
			var value, points fxp.Int
			if def.Type != attribute.PrimarySeparator {
				attr, ok := entity.Attributes.Set[def.ID()]
				if !ok {
					continue
				}
				value = attr.Maximum()
				points = attr.PointCost()
			}
			list = append(list, Attribute{
				Type:   def.Type,
				Key:    def.ID(),
				Name:   def.CombinedName(),
				Value:  value,
				Points: points,
			})
		}
	}
	return list
}

func createSecondaryAttributes(entity *gurps.Entity) []Attribute {
	var list []Attribute
	for _, def := range gurps.SheetSettingsFor(entity).Attributes.List(false) {
		if def.Secondary() {
			var value, points fxp.Int
			if def.Type != attribute.SecondarySeparator {
				attr, ok := entity.Attributes.Set[def.ID()]
				if !ok {
					continue
				}
				value = attr.Maximum()
				points = attr.PointCost()
			}
			list = append(list, Attribute{
				Type:   def.Type,
				Key:    def.ID(),
				Name:   def.CombinedName(),
				Value:  value,
				Points: points,
			})
		}
	}
	return list
}

type PointPool struct {
	Type   attribute.Type
	Key    string
	Name   string
	Value  fxp.Int
	Max    fxp.Int
	Points fxp.Int
	State  string
	Detail string
}

func createPointPools(entity *gurps.Entity) []PointPool {
	var list []PointPool
	for _, def := range gurps.SheetSettingsFor(entity).Attributes.List(false) {
		if def.Secondary() {
			var value, maximum, points fxp.Int
			var state, detail string
			if def.Type != attribute.PoolSeparator {
				attr, ok := entity.Attributes.Set[def.ID()]
				if !ok {
					continue
				}
				value = attr.Current()
				maximum = attr.Maximum()
				points = attr.PointCost()
				if threshold := attr.CurrentThreshold(); threshold != nil {
					state = threshold.State
					detail = threshold.Explanation
				}
			}
			list = append(list, PointPool{
				Type:   def.Type,
				Key:    def.ID(),
				Name:   def.CombinedName(),
				Value:  value,
				Max:    maximum,
				Points: points,
				State:  state,
				Detail: detail,
			})
		}
	}
	return list
}

type BasicDamage struct {
	Thrust string
	Swing  string
}

func createBasicDamage(entity *gurps.Entity) BasicDamage {
	return BasicDamage{
		Thrust: entity.Thrust().String(),
		Swing:  entity.Swing().String(),
	}
}

type HitLocation struct {
	Roll           string
	Location       string
	LocationDetail string
	HitPenalty     int
	DR             string
	DRDetail       string
	SubLocations   []HitLocation
}

func createHitLocations(entity *gurps.Entity, locations *gurps.Body) []HitLocation {
	if locations == nil {
		return nil
	}
	var list []HitLocation
	for _, loc := range locations.Locations {
		rollRange := loc.RollRange
		if rollRange == "-" {
			rollRange = ""
		}
		var detail xio.ByteBuffer
		dr := loc.DisplayDR(entity, &detail)
		list = append(list, HitLocation{
			Roll:           rollRange,
			Location:       loc.TableName,
			LocationDetail: loc.Description,
			HitPenalty:     loc.HitPenalty,
			DR:             dr,
			DRDetail:       fmt.Sprintf(i18n.Text("The DR covering the %s hit location%s"), loc.TableName, detail.String()),
			SubLocations:   createHitLocations(entity, loc.SubTable),
		})
	}
	return list
}

type Body struct {
	Name      string
	Locations []HitLocation
}

func createBody(entity *gurps.Entity) Body {
	locations := gurps.SheetSettingsFor(entity).BodyType
	return Body{
		Name:      locations.Name,
		Locations: createHitLocations(entity, locations),
	}
}

type Encumbrance struct {
	Current    int
	MaxLoad    [encumbrance.LastLevel + 1]fxp.Weight
	Move       [encumbrance.LastLevel + 1]int
	Dodge      [encumbrance.LastLevel + 1]int
	Overloaded bool
}

func createEncumbrance(entity *gurps.Entity) Encumbrance {
	e := Encumbrance{
		Current:    int(entity.EncumbranceLevel(false)),
		Overloaded: entity.WeightCarried(false) > entity.MaximumCarry(encumbrance.ExtraHeavy),
	}
	for i, enc := range encumbrance.Levels {
		e.MaxLoad[i] = entity.MaximumCarry(enc)
		e.Move[i] = entity.Move(enc)
		e.Dodge[i] = entity.Dodge(enc)
	}
	return e
}

type LiftingAndMovingThings struct {
	BasicLift                fxp.Weight
	OneHandedLift            fxp.Weight
	TwoHandedLift            fxp.Weight
	ShoveAndKnockOver        fxp.Weight
	RunningShoveAndKnockOver fxp.Weight
	CarryOnBack              fxp.Weight
	ShiftSlightly            fxp.Weight
}

func createLiftingAndMovingThings(entity *gurps.Entity) LiftingAndMovingThings {
	return LiftingAndMovingThings{
		BasicLift:                entity.BasicLift(),
		OneHandedLift:            entity.OneHandedLift(),
		TwoHandedLift:            entity.TwoHandedLift(),
		ShoveAndKnockOver:        entity.ShoveAndKnockOver(),
		RunningShoveAndKnockOver: entity.RunningShoveAndKnockOver(),
		CarryOnBack:              entity.CarryOnBack(),
		ShiftSlightly:            entity.ShiftSlightly(),
	}
}

type Column struct {
	Title            string
	TitleIsImageKey  bool
	RightAlignedData bool
	IsLink           bool
	Indentable       bool
}

type Cell struct {
	Primary   string
	Secondary string
	Detail    string
}

type Row struct {
	ID    string
	Depth int
	Cells []Cell
}

type Table struct {
	Columns []Column
	Rows    []Row
}

func createReactions(entity *gurps.Entity) *Table {
	// TODO: Build methods on the primary data types for retrieving header information. Cell info can be obtained with
	//		 calls to .CellData()
	return nil
}

func createConditionalModifiers(entity *gurps.Entity) *Table {
	return nil
}

func createMeleeWeapons(entity *gurps.Entity) *Table {
	return nil
}

func createRangedWeapons(entity *gurps.Entity) *Table {
	return nil
}

func createTraits(entity *gurps.Entity) *Table {
	return nil
}

func createSkills(entity *gurps.Entity) *Table {
	return nil
}

func createSpells(entity *gurps.Entity) *Table {
	return nil
}

func createCarriedEquipment(entity *gurps.Entity) *Table {
	return nil
}

func createOtherEquipment(entity *gurps.Entity) *Table {
	return nil
}

func createNotes(entity *gurps.Entity) *Table {
	return nil
}

type Sheet struct {
	Identity               Identity
	Misc                   Misc
	Description            Description
	Points                 Points
	PrimaryAttributes      []Attribute
	SecondaryAttributes    []Attribute
	PointPools             []PointPool
	BasicDamage            BasicDamage
	Body                   Body
	Encumbrance            Encumbrance
	LiftingAndMovingThings LiftingAndMovingThings
	Reactions              *Table
	ConditionalModifiers   *Table
	MeleeWeapons           *Table
	RangedWeapons          *Table
	Traits                 *Table
	Skills                 *Table
	Spells                 *Table
	CarriedEquipment       *Table
	OtherEquipment         *Table
	Notes                  *Table
	Portrait               []byte
}

func NewSheetFromEntity(entity *gurps.Entity) *Sheet {
	return &Sheet{
		Identity:               createIdentity(entity),
		Misc:                   createMisc(entity),
		Description:            createDescription(entity),
		Points:                 createPoints(entity),
		PrimaryAttributes:      createPrimaryAttributes(entity),
		SecondaryAttributes:    createSecondaryAttributes(entity),
		PointPools:             createPointPools(entity),
		BasicDamage:            createBasicDamage(entity),
		Body:                   createBody(entity),
		Encumbrance:            createEncumbrance(entity),
		LiftingAndMovingThings: createLiftingAndMovingThings(entity),
		Reactions:              createReactions(entity),
		ConditionalModifiers:   createConditionalModifiers(entity),
		MeleeWeapons:           createMeleeWeapons(entity),
		RangedWeapons:          createRangedWeapons(entity),
		Traits:                 createTraits(entity),
		Skills:                 createSkills(entity),
		Spells:                 createSpells(entity),
		CarriedEquipment:       createCarriedEquipment(entity),
		OtherEquipment:         createOtherEquipment(entity),
		Notes:                  createNotes(entity),
		Portrait:               entity.Profile.PortraitData,
	}
}
