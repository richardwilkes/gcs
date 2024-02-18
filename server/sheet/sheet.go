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

	"github.com/google/uuid"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wpn"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/ux"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

// Identity holds the data needed by the frontend to display the identity block.
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

// Misc holds the data needed by the frontend to display the misc block.
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

// Description holds the data needed by the frontend to display the description block.
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

// Points holds the data needed by the frontend to display the points block.
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

// Attribute holds the data needed by the frontend to display an attribute.
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

// PointPool holds the data needed by the frontend to display a point pool.
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

// BasicDamage holds the data needed by the frontend to display the basic damage.
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

// HitLocation holds the data needed by the frontend to display the hit locations.
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
	list := make([]HitLocation, 0, len(locations.Locations))
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

// Body holds the data needed by the frontend to display the Body section of a sheet.
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

// Encumbrance holds the data needed by the frontend to display the Encumbrance section of a sheet.
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

// LiftingAndMovingThings holds the data needed by the frontend to display the Lifting and Moving Things section of a
// sheet.
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

// Cell holds the data needed by the frontend to display a table cell.
type Cell struct {
	Primary   string
	Secondary string
	Detail    string
}

// Row holds the data needed by the frontend to display a table row.
type Row struct {
	ID    uuid.UUID
	Depth int
	Cells []gurps.CellData
}

// Table holds the data needed by the frontend to display a table.
type Table struct {
	Columns []gurps.HeaderData
	Rows    []Row
}

func createReactions(entity *gurps.Entity) *Table {
	provider := ux.NewReactionModifiersProvider(entity)
	ids := provider.ColumnIDs()
	root := provider.RootData()
	table := &Table{
		Columns: make([]gurps.HeaderData, len(ids)),
		Rows:    make([]Row, 0, len(root)),
	}
	for i, id := range ids {
		table.Columns[i] = gurps.ReactionModifiersHeaderData(id)
	}
	for _, one := range root {
		collectRowData(one, 0, table, provider, ids)
	}
	return table
}

func createConditionalModifiers(entity *gurps.Entity) *Table {
	provider := ux.NewConditionalModifiersProvider(entity)
	ids := provider.ColumnIDs()
	root := provider.RootData()
	table := &Table{
		Columns: make([]gurps.HeaderData, len(ids)),
		Rows:    make([]Row, 0, len(root)),
	}
	for i, id := range ids {
		table.Columns[i] = gurps.ConditionalModifiersHeaderData(id)
	}
	for _, one := range root {
		collectRowData(one, 0, table, provider, ids)
	}
	return table
}

func createMeleeWeapons(entity *gurps.Entity) *Table {
	return createWeapons(entity, wpn.Melee)
}

func createRangedWeapons(entity *gurps.Entity) *Table {
	return createWeapons(entity, wpn.Ranged)
}

func createWeapons(entity *gurps.Entity, weaponType wpn.Type) *Table {
	provider := ux.NewWeaponsProvider(entity, weaponType, true)
	ids := provider.ColumnIDs()
	root := provider.RootData()
	table := &Table{
		Columns: make([]gurps.HeaderData, len(ids)),
		Rows:    make([]Row, 0, len(root)),
	}
	for i, id := range ids {
		table.Columns[i] = gurps.WeaponHeaderData(id, weaponType, true)
	}
	for _, one := range root {
		collectRowData(one, 0, table, provider, ids)
	}
	return table
}

func createTraits(entity *gurps.Entity) *Table {
	provider := ux.NewTraitsProvider(entity, true)
	ids := provider.ColumnIDs()
	root := provider.RootData()
	table := &Table{
		Columns: make([]gurps.HeaderData, len(ids)),
		Rows:    make([]Row, 0, len(root)),
	}
	for i, id := range ids {
		table.Columns[i] = gurps.TraitsHeaderData(id)
	}
	for _, one := range root {
		collectRowData(one, 0, table, provider, ids)
	}
	return table
}

func createSkills(entity *gurps.Entity) *Table {
	provider := ux.NewSkillsProvider(entity, true)
	ids := provider.ColumnIDs()
	root := provider.RootData()
	table := &Table{
		Columns: make([]gurps.HeaderData, len(ids)),
		Rows:    make([]Row, 0, len(root)),
	}
	for i, id := range ids {
		table.Columns[i] = gurps.SkillsHeaderData(id)
	}
	for _, one := range root {
		collectRowData(one, 0, table, provider, ids)
	}
	return table
}

func createSpells(entity *gurps.Entity) *Table {
	provider := ux.NewSpellsProvider(entity, true)
	ids := provider.ColumnIDs()
	root := provider.RootData()
	table := &Table{
		Columns: make([]gurps.HeaderData, len(ids)),
		Rows:    make([]Row, 0, len(root)),
	}
	for i, id := range ids {
		table.Columns[i] = gurps.SpellsHeaderData(id)
	}
	for _, one := range root {
		collectRowData(one, 0, table, provider, ids)
	}
	return table
}

func createCarriedEquipment(entity *gurps.Entity) *Table {
	return createEquipment(entity, true)
}

func createOtherEquipment(entity *gurps.Entity) *Table {
	return createEquipment(entity, false)
}

func createEquipment(entity *gurps.Entity, carried bool) *Table {
	provider := ux.NewEquipmentProvider(entity, carried, true)
	ids := provider.ColumnIDs()
	root := provider.RootData()
	table := &Table{
		Columns: make([]gurps.HeaderData, len(ids)),
		Rows:    make([]Row, 0, len(root)),
	}
	for i, id := range ids {
		table.Columns[i] = gurps.EquipmentHeaderData(id, entity, carried, true)
	}
	for _, one := range root {
		collectRowData(one, 0, table, provider, ids)
	}
	return table
}

func createNotes(entity *gurps.Entity) *Table {
	provider := ux.NewNotesProvider(entity, true)
	ids := provider.ColumnIDs()
	root := provider.RootData()
	table := &Table{
		Columns: make([]gurps.HeaderData, len(ids)),
		Rows:    make([]Row, 0, len(root)),
	}
	for i, id := range ids {
		table.Columns[i] = gurps.NotesHeaderData(id)
	}
	for _, one := range root {
		collectRowData(one, 0, table, provider, ids)
	}
	return table
}

func collectRowData[T gurps.NodeTypes](node gurps.Node[T], depth int, table *Table, provider ux.TableProvider[T], ids []int) {
	row := Row{
		ID:    node.UUID(),
		Depth: depth,
		Cells: make([]gurps.CellData, len(ids)),
	}
	for i, id := range ids {
		node.CellData(id, &row.Cells[i])
	}
	table.Rows = append(table.Rows, row)
	if node.HasChildren() {
		depth++
		for _, child := range node.NodeChildren() {
			collectRowData[T](gurps.AsNode(child), depth, table, provider, ids)
		}
	}
}

// Sheet holds the data needed by the frontend to display a GURPS character sheet.
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

// NewSheetFromEntity creates a new Sheet from the given entity.
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
