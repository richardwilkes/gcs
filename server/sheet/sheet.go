// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package sheet

import (
	"fmt"
	"path"
	"strconv"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
	"github.com/richardwilkes/gcs/v5/ux"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/tid"
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
	Created  string
	Modified string
	Player   string
}

func createMisc(entity *gurps.Entity) Misc {
	return Misc{
		Created:  entity.CreatedOn.String(),
		Modified: entity.ModifiedOn.String(),
		Player:   entity.Profile.PlayerName,
	}
}

// Description holds the data needed by the frontend to display the description block.
type Description struct {
	Gender       string
	Age          string
	Birthday     string
	Religion     string
	Height       string
	Weight       string
	SizeModifier string
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
		Height:       entity.SheetSettings.DefaultLengthUnits.Format(entity.Profile.Height),
		Weight:       entity.SheetSettings.DefaultWeightUnits.Format(entity.Profile.Weight),
		SizeModifier: fmt.Sprintf("%+d", entity.Profile.AdjustedSizeModifier()),
		TechLevel:    entity.Profile.TechLevel,
		Hair:         entity.Profile.Hair,
		Eyes:         entity.Profile.Eyes,
		Skin:         entity.Profile.Skin,
		Hand:         entity.Profile.Handedness,
	}
}

// Points holds the data needed by the frontend to display the points block.
type Points struct {
	Total         string
	Unspent       string
	Ancestry      string
	Attributes    string
	Advantages    string
	Disadvantages string
	Quirks        string
	Skills        string
	Spells        string
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
		Total:         totalPoints.Comma(),
		Unspent:       entity.UnspentPoints().Comma(),
		Ancestry:      pointsBreakdown.Ancestry.Comma(),
		Attributes:    pointsBreakdown.Attributes.Comma(),
		Advantages:    pointsBreakdown.Advantages.Comma(),
		Disadvantages: pointsBreakdown.Disadvantages.Comma(),
		Quirks:        pointsBreakdown.Quirks.Comma(),
		Skills:        pointsBreakdown.Skills.Comma(),
		Spells:        pointsBreakdown.Spells.Comma(),
	}
}

// Attribute holds the data needed by the frontend to display an attribute.
type Attribute struct {
	Type   string
	Key    string
	Name   string
	Value  string
	Points string
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
				Type:   def.Type.Key(),
				Key:    def.ID(),
				Name:   def.CombinedName(),
				Value:  value.Comma(),
				Points: points.Comma(),
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
				Type:   def.Type.Key(),
				Key:    def.ID(),
				Name:   def.CombinedName(),
				Value:  value.Comma(),
				Points: points.Comma(),
			})
		}
	}
	return list
}

// PointPool holds the data needed by the frontend to display a point pool.
type PointPool struct {
	Type   string
	Key    string
	Name   string
	Value  string
	Max    string
	Points string
	State  string
	Detail string
}

func createPointPools(entity *gurps.Entity) []PointPool {
	var list []PointPool
	for _, def := range gurps.SheetSettingsFor(entity).Attributes.List(false) {
		if def.Pool() {
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
				Type:   def.Type.Key(),
				Key:    def.ID(),
				Name:   def.CombinedName(),
				Value:  value.Comma(),
				Max:    maximum.Comma(),
				Points: points.Comma(),
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
	ID             int
	Roll           string
	Location       string
	LocationDetail string
	HitPenalty     string
	DR             string
	DRDetail       string
	Notes          string
	SubLocations   []HitLocation
}

func createHitLocations(entity *gurps.Entity, locations *gurps.Body, startID int) (list []HitLocation, endID int) {
	if locations == nil {
		return nil, startID
	}
	list = make([]HitLocation, 0, len(locations.Locations))
	for _, loc := range locations.Locations {
		rollRange := loc.RollRange
		if rollRange == "-" {
			rollRange = ""
		}
		var detail xio.ByteBuffer
		dr := loc.DisplayDR(entity, &detail)
		one := HitLocation{
			ID:             startID,
			Roll:           rollRange,
			Location:       loc.TableName,
			LocationDetail: loc.Description,
			HitPenalty:     strconv.Itoa(loc.HitPenalty),
			DR:             dr,
			DRDetail:       fmt.Sprintf(i18n.Text("The DR covering the %s hit location%s"), loc.TableName, detail.String()),
			Notes:          loc.Notes,
		}
		startID++
		one.SubLocations, startID = createHitLocations(entity, loc.SubTable, startID)
		list = append(list, one)
	}
	return list, startID
}

// FindHitLocationByID returns the HitLocation with the given ID.
func FindHitLocationByID(entity *gurps.Entity, locations *gurps.Body, id int) *gurps.HitLocation {
	match, _ := _findHitLocationByID(entity, locations, id, 1)
	return match
}

func _findHitLocationByID(entity *gurps.Entity, locations *gurps.Body, id, startID int) (match *gurps.HitLocation, endID int) {
	if locations == nil {
		return nil, startID
	}
	for _, loc := range locations.Locations {
		if startID == id {
			return loc, startID
		}
		startID++
		if match, startID = _findHitLocationByID(entity, loc.SubTable, id, startID); match != nil {
			return match, startID
		}
	}
	return nil, startID
}

// Body holds the data needed by the frontend to display the Body section of a sheet.
type Body struct {
	Name      string
	Locations []HitLocation
}

func createBody(entity *gurps.Entity) Body {
	locations := gurps.SheetSettingsFor(entity).BodyType
	list, _ := createHitLocations(entity, locations, 1)
	return Body{
		Name:      locations.Name,
		Locations: list,
	}
}

// Encumbrance holds the data needed by the frontend to display the Encumbrance section of a sheet.
type Encumbrance struct {
	Current    int
	MaxLoad    [encumbrance.LastLevel + 1]string
	Move       [encumbrance.LastLevel + 1]string
	Dodge      [encumbrance.LastLevel + 1]string
	Overloaded bool
}

func createEncumbrance(entity *gurps.Entity) Encumbrance {
	e := Encumbrance{
		Current:    int(entity.EncumbranceLevel(false)),
		Overloaded: entity.WeightCarried(false) > entity.MaximumCarry(encumbrance.ExtraHeavy),
	}
	for i, enc := range encumbrance.Levels {
		e.MaxLoad[i] = entity.MaximumCarry(enc).String()
		e.Move[i] = strconv.Itoa(entity.Move(enc))
		e.Dodge[i] = strconv.Itoa(entity.Dodge(enc))
	}
	return e
}

// LiftingAndMovingThings holds the data needed by the frontend to display the Lifting and Moving Things section of a
// sheet.
type LiftingAndMovingThings struct {
	BasicLift                string
	OneHandedLift            string
	TwoHandedLift            string
	ShoveAndKnockOver        string
	RunningShoveAndKnockOver string
	CarryOnBack              string
	ShiftSlightly            string
}

func createLiftingAndMovingThings(entity *gurps.Entity) LiftingAndMovingThings {
	return LiftingAndMovingThings{
		BasicLift:                entity.SheetSettings.DefaultWeightUnits.Format(entity.BasicLift()),
		OneHandedLift:            entity.SheetSettings.DefaultWeightUnits.Format(entity.OneHandedLift()),
		TwoHandedLift:            entity.SheetSettings.DefaultWeightUnits.Format(entity.TwoHandedLift()),
		ShoveAndKnockOver:        entity.SheetSettings.DefaultWeightUnits.Format(entity.ShoveAndKnockOver()),
		RunningShoveAndKnockOver: entity.SheetSettings.DefaultWeightUnits.Format(entity.RunningShoveAndKnockOver()),
		CarryOnBack:              entity.SheetSettings.DefaultWeightUnits.Format(entity.CarryOnBack()),
		ShiftSlightly:            entity.SheetSettings.DefaultWeightUnits.Format(entity.ShiftSlightly()),
	}
}

// Row holds the data needed by the frontend to display a table row.
type Row struct {
	ID    tid.TID
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
	return createWeapons(entity, true)
}

func createRangedWeapons(entity *gurps.Entity) *Table {
	return createWeapons(entity, false)
}

func createWeapons(entity *gurps.Entity, melee bool) *Table {
	provider := ux.NewWeaponsProvider(entity, melee, true)
	ids := provider.ColumnIDs()
	root := provider.RootData()
	table := &Table{
		Columns: make([]gurps.HeaderData, len(ids)),
		Rows:    make([]Row, 0, len(root)),
	}
	for i, id := range ids {
		table.Columns[i] = gurps.WeaponHeaderData(id, melee, true)
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
		ID:    node.ID(),
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

// PageRef holds the data needed by the frontend to display and use a page reference.
type PageRef struct {
	Name   string
	Offset int
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
	PageRefs               map[string]PageRef
	Modified               bool
	ReadOnly               bool
}

// NewSheetFromEntity creates a new Sheet from the given entity.
func NewSheetFromEntity(entity *gurps.Entity, modified, readOnly bool) *Sheet {
	refs := make(map[string]PageRef)
	for _, one := range gurps.GlobalSettings().PageRefs.List() {
		refs[one.ID] = PageRef{
			Name:   path.Base(one.Path),
			Offset: one.Offset,
		}
	}
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
		PageRefs:               refs,
		Modified:               modified,
		ReadOnly:               readOnly,
	}
}
