/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package gurps

import (
	"bufio"
	"bytes"
	"encoding/base64"
	htmltmpl "html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	texttmpl "text/template"

	"github.com/google/uuid"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/toolbox/xio/fs"
	"golang.org/x/exp/slices"
)

type exportedMeleeWeapon struct {
	Description string
	Notes       string
	Usage       string
	Level       fxp.Int
	Parry       string
	Block       string
	Damage      string
	Reach       string
	Strength    string
}

type exportedRangedWeapon struct {
	Description string
	Notes       string
	Usage       string
	Level       fxp.Int
	Accuracy    string
	Range       string
	Damage      string
	RateOfFire  string
	Shots       string
	Bulk        string
	Recoil      string
	Strength    string
}

type exportedHitLocation struct {
	RollRange string
	Where     string
	Penalty   int
	DR        string
	Notes     string
	Depth     int
}

type exportedBodyType struct {
	Name      string
	Locations []*exportedHitLocation
}

type exportedNote struct {
	ID          uuid.UUID
	ParentID    uuid.UUID
	Type        string
	Description string
	PageRef     string
	Depth       int
}

type exportedMana struct {
	Cast     string
	Maintain string
}

type exportedSpell struct {
	ID                uuid.UUID
	ParentID          uuid.UUID
	Type              string
	Points            fxp.Int
	Level             string
	RelativeLevel     string
	Difficulty        string
	Class             string
	Colleges          []string
	Mana              exportedMana
	TimeToCast        string
	Duration          string
	Resist            string
	Description       string
	Notes             string
	Rituals           string
	UnsatisfiedReason string
	PageRef           string
	Tags              []string
	Depth             int
}

type exportedEquipment struct {
	ID                uuid.UUID
	ParentID          uuid.UUID
	Type              string
	Quantity          fxp.Int
	Description       string
	ModifierNotes     string
	Notes             string
	TechLevel         string
	LegalityClass     string
	UnsatisfiedReason string
	PageRef           string
	Tags              []string
	Depth             int
	Uses              int
	MaxUses           int
	Cost              fxp.Int
	ExtendedCost      fxp.Int
	Weight            Weight
	ExtendedWeight    Weight
	Equipped          bool
}

type exportedAllEquipment struct {
	Carried       []*exportedEquipment
	CarriedValue  fxp.Int
	CarriedWeight Weight
	Other         []*exportedEquipment
	OtherValue    fxp.Int
}

type exportedSkill struct {
	ID                uuid.UUID
	ParentID          uuid.UUID
	Type              string
	Points            fxp.Int
	Level             string
	RelativeLevel     string
	Difficulty        string
	Description       string
	ModifierNotes     string
	Notes             string
	UnsatisfiedReason string
	PageRef           string
	Tags              []string
	Depth             int
}

type exportedTrait struct {
	ID                uuid.UUID
	ParentID          uuid.UUID
	Type              string
	Points            fxp.Int
	Description       string
	UserDescription   string
	ModifierNotes     string
	Notes             string
	UnsatisfiedReason string
	PageRef           string
	Tags              []string
	Depth             int
}

type exportedSource struct {
	Source string
	Amount fxp.Int
}

type exportedConditionalModifier struct {
	ID        uuid.UUID
	Total     fxp.Int
	Situation string
	Sources   []*exportedSource
}

type exportedLift struct {
	Basic         Weight
	OneHanded     Weight
	TwoHanded     Weight
	Shove         Weight
	RunningShove  Weight
	CarryOnBack   Weight
	ShiftSlightly Weight
}

type exportedEncumbrance struct {
	Name      string
	Level     int
	Penalty   int
	Move      int
	Dodge     int
	MaxLoad   Weight
	IsCurrent bool
}

type exportedAttribute struct {
	ID           string
	Name         string
	FullName     string
	CombinedName string
	Value        fxp.Int
	Points       fxp.Int
	order        int
}

type exportedPool struct {
	*exportedAttribute
	Current fxp.Int
	Maximum fxp.Int
}

type exportedAttributes struct {
	Primary   []*exportedAttribute
	Secondary []*exportedAttribute
	Pools     []*exportedPool
}

type exportedPoints struct {
	Total   fxp.Int
	Unspent fxp.Int
	PointsBreakdown
}

type exportedEntity struct {
	Name                    string
	EmbeddedPortraitDataURL htmltmpl.URL
	Player                  string
	Title                   string
	Organization            string
	Religion                string
	TechLevel               string
	SizeModifier            int
	Age                     string
	Birthday                string
	Eyes                    string
	Hair                    string
	Skin                    string
	Handedness              string
	Gender                  string
	Height                  Length
	Weight                  Weight
	Thrust                  *dice.Dice
	Swing                   *dice.Dice
	Encumbrance             []*exportedEncumbrance
	Lift                    exportedLift
	Points                  exportedPoints
	Attributes              exportedAttributes
	BodyType                exportedBodyType
	Reactions               []*exportedConditionalModifier
	ConditionalModifiers    []*exportedConditionalModifier
	Traits                  []*exportedTrait
	Skills                  []*exportedSkill
	Spells                  []*exportedSpell
	Equipment               exportedAllEquipment
	Notes                   []*exportedNote
	MeleeWeapons            []*exportedMeleeWeapon
	RangedWeapons           []*exportedRangedWeapon
	GridTemplate            htmltmpl.CSS
	CreatedOn               string
	ModifiedOn              string
}

// ExportSheets exports the files to a text representation.
func ExportSheets(templatePath string, fileList []string) error {
	for _, one := range fileList {
		if FileInfoFor(one).IsExportable {
			// Currently, only one file type supports exporting. Should this change, this will need to be adjusted to
			// call the correct loader.
			entity, err := NewEntityFromFile(os.DirFS(filepath.Dir(one)), filepath.Base(one))
			if err != nil {
				return err
			}
			if err = Export(entity, templatePath, fs.TrimExtension(one)+filepath.Ext(templatePath)); err != nil {
				return err
			}
		} else {
			jot.Warn("not exportable, skipping: " + one)
		}
	}
	return nil
}

// Export an Entity to exportPath using the template found at templatePath.
func Export(entity *Entity, templatePath, exportPath string) error {
	tmpl, err := os.ReadFile(templatePath)
	if err != nil {
		return errs.Wrap(err)
	}
	var advance int
	var line []byte
	if advance, line, err = bufio.ScanLines(tmpl, true); err != nil {
		return errs.Wrap(err)
	}
	entity.Recalculate()
	switch {
	case bytes.HasPrefix(line, []byte("GCS HTML Template v1")):
		var t *htmltmpl.Template
		if t, err = htmltmpl.New("").Funcs(createTemplateFuncs()).Parse(string(tmpl[advance:])); err != nil {
			return errs.Wrap(err)
		}
		return export(entity, t, exportPath)
	case bytes.HasPrefix(line, []byte("GCS Text Template v1")):
		var t *texttmpl.Template
		if t, err = texttmpl.New("").Funcs(createTemplateFuncs()).Parse(string(tmpl[advance:])); err != nil {
			return errs.Wrap(err)
		}
		return export(entity, t, exportPath)
	default: // Legacy text export
		return legacyTextExport(entity, tmpl, exportPath)
	}
}

func createTemplateFuncs() texttmpl.FuncMap {
	return texttmpl.FuncMap{
		"numberToInt":   fxp.As[int],
		"numberToFloat": fxp.As[float64],
		"join":          strings.Join,
	}
}

type exporter interface {
	Execute(wr io.Writer, data any) error
}

func export(entity *Entity, tmpl exporter, exportPath string) (err error) {
	var f *os.File
	f, err = os.Create(exportPath)
	if err != nil {
		return errs.Wrap(err)
	}
	buffer := bufio.NewWriter(f)
	defer func() {
		if flushErr := buffer.Flush(); flushErr != nil && err == nil {
			err = errs.Wrap(flushErr)
		}
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = errs.Wrap(closeErr)
		}
	}()
	pb := entity.PointsBreakdown()
	data := &exportedEntity{
		Name:         entity.Profile.Name,
		Player:       entity.Profile.PlayerName,
		Title:        entity.Profile.Title,
		Organization: entity.Profile.Organization,
		Religion:     entity.Profile.Religion,
		TechLevel:    entity.Profile.TechLevel,
		SizeModifier: entity.Profile.AdjustedSizeModifier(),
		Age:          entity.Profile.Age,
		Birthday:     entity.Profile.Birthday,
		Eyes:         entity.Profile.Eyes,
		Hair:         entity.Profile.Hair,
		Skin:         entity.Profile.Skin,
		Handedness:   entity.Profile.Handedness,
		Gender:       entity.Profile.Gender,
		Height:       entity.Profile.Height,
		Weight:       entity.Profile.Weight,
		Thrust:       entity.Thrust(),
		Swing:        entity.Swing(),
		Lift: exportedLift{
			Basic:         entity.BasicLift(),
			OneHanded:     entity.OneHandedLift(),
			TwoHanded:     entity.TwoHandedLift(),
			Shove:         entity.ShoveAndKnockOver(),
			RunningShove:  entity.RunningShoveAndKnockOver(),
			CarryOnBack:   entity.CarryOnBack(),
			ShiftSlightly: entity.ShiftSlightly(),
		},
		Points: exportedPoints{
			Total:           entity.TotalPoints,
			Unspent:         entity.UnspentPoints(),
			PointsBreakdown: *pb,
		},
		BodyType: exportedBodyType{
			Name:      entity.SheetSettings.BodyType.Name,
			Locations: addToHitLocations(entity, nil, 0, entity.SheetSettings.BodyType.Locations),
		},
		Reactions:            newExportedConditionalModifiers(entity.Reactions()),
		ConditionalModifiers: newExportedConditionalModifiers(entity.ConditionalModifiers()),
		Equipment: exportedAllEquipment{
			Carried:       newExportedEquipment(entity, entity.CarriedEquipment, true),
			CarriedValue:  entity.WealthCarried(),
			CarriedWeight: entity.WeightCarried(false),
			Other:         newExportedEquipment(entity, entity.OtherEquipment, false),
			OtherValue:    entity.WealthNotCarried(),
		},
		GridTemplate: htmltmpl.CSS(entity.SheetSettings.BlockLayout.HTMLGridTemplate()),
		CreatedOn:    entity.CreatedOn.String(),
		ModifiedOn:   entity.ModifiedOn.String(),
	}
	if entity.SheetSettings.ExcludeUnspentPointsFromTotal {
		data.Points.Total = pb.Total()
	}
	if len(entity.Profile.PortraitData) != 0 {
		data.EmbeddedPortraitDataURL = htmltmpl.URL("data:" + http.DetectContentType(entity.Profile.PortraitData) + ";base64," + base64.StdEncoding.EncodeToString(entity.Profile.PortraitData)) //nolint:gosec // This is a valid data URL
	}
	for _, def := range entity.SheetSettings.Attributes.List(true) {
		if attr, ok := entity.Attributes.Set[def.DefID]; ok {
			switch {
			case def.Primary():
				data.Attributes.Primary = append(data.Attributes.Primary, newExportedAttribute(def, attr))
			case def.Secondary():
				data.Attributes.Secondary = append(data.Attributes.Secondary, newExportedAttribute(def, attr))
			case def.Pool():
				data.Attributes.Pools = append(data.Attributes.Pools, &exportedPool{
					exportedAttribute: newExportedAttribute(def, attr),
					Current:           attr.Current(),
					Maximum:           attr.Maximum(),
				})
			}
		}
	}
	slices.SortFunc(data.Attributes.Primary, func(a, b *exportedAttribute) int { return a.order - b.order })
	slices.SortFunc(data.Attributes.Secondary, func(a, b *exportedAttribute) int { return a.order - b.order })
	slices.SortFunc(data.Attributes.Pools, func(a, b *exportedPool) int { return a.order - b.order })
	currentEnc := entity.EncumbranceLevel(false)
	for _, enc := range AllEncumbrance {
		penalty := fxp.As[int](enc.Penalty())
		data.Encumbrance = append(data.Encumbrance, &exportedEncumbrance{
			Name:      enc.String(),
			Level:     -penalty,
			Penalty:   penalty,
			Move:      entity.Move(enc),
			Dodge:     entity.Dodge(enc),
			MaxLoad:   entity.MaximumCarry(enc),
			IsCurrent: enc == currentEnc,
		})
	}
	Traverse(func(t *Trait) bool {
		trait := &exportedTrait{
			ID:                t.ID,
			Points:            t.AdjustedPoints(),
			Description:       t.String(),
			UserDescription:   t.UserDesc,
			ModifierNotes:     t.ModifierNotes(),
			Notes:             t.Notes(),
			UnsatisfiedReason: t.UnsatisfiedReason,
			PageRef:           t.PageRef,
			Tags:              slices.Clone(t.Tags),
			Depth:             t.Depth(),
		}
		if parent := t.Parent(); parent != nil {
			trait.ParentID = parent.ID
		}
		if t.Container() {
			trait.Type = t.ContainerType.Key()
		} else {
			trait.Type = groupOrItem(false)
		}
		data.Traits = append(data.Traits, trait)
		return false
	}, true, false, entity.Traits...)
	Traverse(func(s *Skill) bool {
		skill := &exportedSkill{
			ID:                s.ID,
			Type:              groupOrItem(s.Container()),
			Points:            s.AdjustedPoints(nil),
			Level:             s.CalculateLevel().LevelAsString(s.Container()),
			RelativeLevel:     s.RelativeLevel(),
			Description:       s.String(),
			ModifierNotes:     s.ModifierNotes(),
			Notes:             s.Notes(),
			UnsatisfiedReason: s.UnsatisfiedReason,
			PageRef:           s.PageRef,
			Tags:              slices.Clone(s.Tags),
			Depth:             s.Depth(),
		}
		if parent := s.Parent(); parent != nil {
			skill.ParentID = parent.ID
		}
		if !s.Container() {
			skill.Difficulty = s.Difficulty.Description(entity)
		}
		data.Skills = append(data.Skills, skill)
		return false
	}, true, false, entity.Skills...)
	Traverse(func(s *Spell) bool {
		spell := &exportedSpell{
			ID:            s.ID,
			Type:          groupOrItem(s.Container()),
			Points:        s.AdjustedPoints(nil),
			Level:         s.CalculateLevel().LevelAsString(s.Container()),
			RelativeLevel: s.RelativeLevel(),
			Class:         s.Class,
			Colleges:      slices.Clone(s.College),
			Mana: exportedMana{
				Cast:     s.CastingCost,
				Maintain: s.MaintenanceCost,
			},
			TimeToCast:        s.CastingTime,
			Duration:          s.Duration,
			Resist:            s.Resist,
			Description:       s.String(),
			Notes:             s.Notes(),
			Rituals:           s.Rituals(),
			UnsatisfiedReason: s.UnsatisfiedReason,
			PageRef:           s.PageRef,
			Tags:              slices.Clone(s.Tags),
			Depth:             s.Depth(),
		}
		if parent := s.Parent(); parent != nil {
			spell.ParentID = parent.ID
		}
		if !s.Container() {
			spell.Difficulty = s.Difficulty.Description(entity)
		}
		data.Spells = append(data.Spells, spell)
		return false
	}, true, false, entity.Spells...)
	Traverse(func(n *Note) bool {
		note := &exportedNote{
			ID:          n.ID,
			Type:        groupOrItem(n.Container()),
			Description: n.String(),
			PageRef:     n.PageRef,
			Depth:       n.Depth(),
		}
		if parent := n.Parent(); parent != nil {
			note.ParentID = parent.ID
		}
		data.Notes = append(data.Notes, note)
		return false
	}, true, false, entity.Notes...)
	for _, w := range entity.EquippedWeapons(MeleeWeaponType) {
		data.MeleeWeapons = append(data.MeleeWeapons, &exportedMeleeWeapon{
			Description: w.String(),
			Notes:       w.Notes(),
			Usage:       w.Usage,
			Level:       w.SkillLevel(nil),
			Parry:       w.ResolvedParry(nil),
			Block:       w.ResolvedBlock(nil),
			Damage:      w.Damage.ResolvedDamage(nil),
			Reach:       w.Reach,
			Strength:    w.MinimumStrength,
		})
	}
	for _, w := range entity.EquippedWeapons(RangedWeaponType) {
		data.RangedWeapons = append(data.RangedWeapons, &exportedRangedWeapon{
			Description: w.String(),
			Notes:       w.Notes(),
			Usage:       w.Usage,
			Level:       w.SkillLevel(nil),
			Accuracy:    w.Accuracy,
			Range:       w.ResolvedRange(),
			Damage:      w.Damage.ResolvedDamage(nil),
			RateOfFire:  w.RateOfFire,
			Shots:       w.Shots,
			Bulk:        w.Bulk,
			Recoil:      w.Recoil,
			Strength:    w.MinimumStrength,
		})
	}
	if err = tmpl.Execute(buffer, data); err != nil {
		err = errs.Wrap(err)
		return err
	}
	return nil
}

func newExportedAttribute(def *AttributeDef, attr *Attribute) *exportedAttribute {
	return &exportedAttribute{
		ID:           def.DefID,
		Name:         def.Name,
		FullName:     def.ResolveFullName(),
		CombinedName: def.CombinedName(),
		Value:        attr.Maximum(),
		Points:       attr.PointCost(),
		order:        def.Order,
	}
}

func newExportedConditionalModifiers(list []*ConditionalModifier) []*exportedConditionalModifier {
	result := make([]*exportedConditionalModifier, 0, len(list))
	for _, one := range list {
		r := &exportedConditionalModifier{
			ID:        one.ID,
			Situation: one.From,
			Total:     one.Total(),
		}
		r.Sources = make([]*exportedSource, 0, len(one.Sources))
		for i := range one.Sources {
			r.Sources = append(r.Sources, &exportedSource{
				Source: one.Sources[i],
				Amount: one.Amounts[i],
			})
		}
		result = append(result, r)
	}
	return result
}

func newExportedEquipment(entity *Entity, list []*Equipment, carried bool) []*exportedEquipment {
	var result []*exportedEquipment
	Traverse(func(e *Equipment) bool {
		equipment := &exportedEquipment{
			ID:                e.ID,
			Type:              groupOrItem(e.Container()),
			Quantity:          e.Quantity,
			Description:       e.String(),
			ModifierNotes:     e.ModifierNotes(),
			Notes:             e.Notes(),
			TechLevel:         e.TechLevel,
			LegalityClass:     e.LegalityClass,
			UnsatisfiedReason: e.UnsatisfiedReason,
			PageRef:           e.PageRef,
			Tags:              slices.Clone(e.Tags),
			Depth:             e.Depth(),
			Uses:              e.Uses,
			MaxUses:           e.MaxUses,
			Cost:              e.AdjustedValue(),
			ExtendedCost:      e.ExtendedValue(),
			Weight:            e.AdjustedWeight(false, entity.SheetSettings.DefaultWeightUnits),
			ExtendedWeight:    e.ExtendedWeight(false, entity.SheetSettings.DefaultWeightUnits),
			Equipped:          carried && e.Equipped,
		}
		if parent := e.Parent(); parent != nil {
			equipment.ParentID = parent.ID
		}
		result = append(result, equipment)
		return false
	}, true, false, list...)
	return result
}

func groupOrItem(isContainer bool) string {
	if isContainer {
		return "group"
	}
	return "item"
}

func addToHitLocations(entity *Entity, locations []*exportedHitLocation, depth int, hitLocations []*HitLocation) []*exportedHitLocation {
	for _, location := range hitLocations {
		loc := &exportedHitLocation{
			RollRange: location.RollRange,
			Where:     location.TableName,
			Penalty:   location.HitPenalty,
			Depth:     depth,
		}
		var tooltip xio.ByteBuffer
		loc.DR = location.DisplayDR(entity, &tooltip)
		loc.Notes = tooltip.String()
		locations = append(locations, loc)
		if location.SubTable != nil {
			locations = addToHitLocations(entity, locations, depth+1, location.SubTable.Locations)
		}
	}
	return locations
}
