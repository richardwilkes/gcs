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
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/toolbox/xio/fs"
)

const (
	descriptionExportKey        = "DESCRIPTION"
	descriptionPrimaryExportKey = "DESCRIPTION_PRIMARY"
	idExportKey                 = "ID"
	nameExportKey               = "NAME"
	parentIDExportKey           = "PARENT_ID"
	pointsExportKey             = "POINTS"
	refExportKey                = "REF"
	satisfiedExportKey          = "SATISFIED"
	styleIndentWarningExportKey = "STYLE_INDENT_WARNING"
	techLevelExportKey          = "TL"
	typeExportKey               = "TYPE"
	weightExportKey             = "WEIGHT"
	hpAttrID                    = "hp"
	fpAttrID                    = "fp"
)

type legacyExporter struct {
	entity             *Entity
	points             *PointsBreakdown
	template           []byte
	pos                int
	exportPath         string
	onlyTags           map[string]bool
	excludedTags       map[string]bool
	out                *bufio.Writer
	encodeText         bool
	enhancedKeyParsing bool
}

// legacyTextExport performs the text template export function that matches the old Java code base.
func legacyTextExport(entity *Entity, tmpl []byte, exportPath string) (err error) {
	ex := &legacyExporter{
		entity:       entity,
		points:       entity.PointsBreakdown(),
		template:     tmpl,
		exportPath:   exportPath,
		onlyTags:     make(map[string]bool),
		excludedTags: make(map[string]bool),
		encodeText:   true,
	}
	var out *os.File
	if out, err = os.Create(exportPath); err != nil {
		return errs.Wrap(err)
	}
	ex.out = bufio.NewWriter(out)
	defer func() { //nolint:gosec // Yes, this is safe
		if flushErr := ex.out.Flush(); flushErr != nil && err == nil {
			err = errs.Wrap(flushErr)
		}
		if closeErr := out.Close(); closeErr != nil && err == nil {
			err = errs.Wrap(closeErr)
		}
	}()
	lookForKeyMarker := true
	var keyBuffer bytes.Buffer
	for ex.pos < len(ex.template) {
		ch := ex.template[ex.pos]
		ex.pos++
		switch {
		case lookForKeyMarker:
			var next byte
			if ex.pos < len(ex.template) {
				next = ex.template[ex.pos]
			}
			if ch == '@' && (next < '0' || next > '9') {
				lookForKeyMarker = false
			} else {
				ex.out.WriteByte(ch)
			}
		case ch == '_' || (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z'):
			keyBuffer.WriteByte(ch)
		default:
			if !ex.enhancedKeyParsing || ch != '@' {
				ex.pos--
			}
			if err = ex.emitKey(keyBuffer.String()); err != nil {
				return err
			}
			keyBuffer.Reset()
			lookForKeyMarker = true
		}
	}
	if keyBuffer.Len() != 0 {
		if err = ex.emitKey(keyBuffer.String()); err != nil {
			return err
		}
	}
	return nil
}

func (ex *legacyExporter) emitKey(key string) error {
	switch key {
	case "GRID_TEMPLATE":
		ex.out.WriteString(ex.entity.SheetSettings.BlockLayout.HTMLGridTemplate())
	case "ENCODING_OFF":
		ex.encodeText = false
	case "ENHANCED_KEY_PARSING":
		ex.enhancedKeyParsing = true
	case "PORTRAIT":
		if ex.entity.Profile.CanExportPortrait() {
			if ext := ex.entity.Profile.PortraitExtension(); ext != "" {
				leafName := fs.TrimExtension(filepath.Base(ex.exportPath)) + ext
				if err := os.WriteFile(filepath.Join(filepath.Dir(ex.exportPath), leafName), ex.entity.Profile.PortraitData, 0o640); err != nil {
					return errs.Wrap(err)
				}
				ex.out.WriteString(url.PathEscape(leafName))
			}
		}
	case "PORTRAIT_EMBEDDED":
		if len(ex.entity.Profile.PortraitData) != 0 {
			ex.out.WriteString("data:")
			ex.out.WriteString(http.DetectContentType(ex.entity.Profile.PortraitData))
			ex.out.WriteString(";base64,")
			ex.out.WriteString(base64.StdEncoding.EncodeToString(ex.entity.Profile.PortraitData))
		}
	case nameExportKey:
		ex.writeEncodedText(ex.entity.Profile.Name)
	case "TITLE":
		ex.writeEncodedText(ex.entity.Profile.Title)
	case "ORGANIZATION":
		ex.writeEncodedText(ex.entity.Profile.Organization)
	case "RELIGION":
		ex.writeEncodedText(ex.entity.Profile.Religion)
	case "PLAYER":
		ex.writeEncodedText(ex.entity.Profile.PlayerName)
	case "CREATED_ON":
		ex.writeEncodedText(ex.entity.CreatedOn.String())
	case "MODIFIED_ON":
		ex.writeEncodedText(ex.entity.ModifiedOn.String())
	case "TOTAL_POINTS":
		if ex.entity.SheetSettings.ExcludeUnspentPointsFromTotal {
			ex.writeEncodedText(ex.points.Total().String())
		} else {
			ex.writeEncodedText(ex.entity.TotalPoints.String())
		}
	case "ATTRIBUTE_POINTS":
		ex.writeEncodedText(ex.points.Attributes.String())
	case "ST_POINTS":
		ex.writeEncodedText(ex.entity.Attributes.Cost(StrengthID).String())
	case "DX_POINTS":
		ex.writeEncodedText(ex.entity.Attributes.Cost(DexterityID).String())
	case "IQ_POINTS":
		ex.writeEncodedText(ex.entity.Attributes.Cost(IntelligenceID).String())
	case "HT_POINTS":
		ex.writeEncodedText(ex.entity.Attributes.Cost("ht").String())
	case "PERCEPTION_POINTS":
		ex.writeEncodedText(ex.entity.Attributes.Cost("perception").String())
	case "WILL_POINTS":
		ex.writeEncodedText(ex.entity.Attributes.Cost("will").String())
	case "FP_POINTS":
		ex.writeEncodedText(ex.entity.Attributes.Cost(fpAttrID).String())
	case "HP_POINTS":
		ex.writeEncodedText(ex.entity.Attributes.Cost(hpAttrID).String())
	case "BASIC_SPEED_POINTS":
		ex.writeEncodedText(ex.entity.Attributes.Cost(BasicSpeedID).String())
	case "BASIC_MOVE_POINTS":
		ex.writeEncodedText(ex.entity.Attributes.Cost(BasicMoveID).String())
	case "ADVANTAGE_POINTS":
		ex.writeEncodedText(ex.points.Advantages.String())
	case "DISADVANTAGE_POINTS":
		ex.writeEncodedText(ex.points.Disadvantages.String())
	case "QUIRK_POINTS":
		ex.writeEncodedText(ex.points.Quirks.String())
	case "RACE_POINTS", "ANCESTRY_POINTS":
		ex.writeEncodedText(ex.points.Ancestry.String())
	case "SKILL_POINTS":
		ex.writeEncodedText(ex.points.Skills.String())
	case "SPELL_POINTS":
		ex.writeEncodedText(ex.points.Spells.String())
	case "UNSPENT_POINTS", "EARNED_POINTS":
		ex.writeEncodedText(ex.entity.UnspentPoints().String())
	case "HEIGHT":
		ex.writeEncodedText(ex.entity.SheetSettings.DefaultLengthUnits.Format(ex.entity.Profile.Height))
	case weightExportKey:
		ex.writeEncodedText(ex.entity.SheetSettings.DefaultWeightUnits.Format(ex.entity.Profile.Weight))
	case "GENDER":
		ex.writeEncodedText(ex.entity.Profile.Gender)
	case "HAIR":
		ex.writeEncodedText(ex.entity.Profile.Hair)
	case "EYES":
		ex.writeEncodedText(ex.entity.Profile.Eyes)
	case "AGE":
		ex.writeEncodedText(ex.entity.Profile.Age)
	case "SIZE":
		ex.writeEncodedText(fmt.Sprintf("%+d", ex.entity.Profile.AdjustedSizeModifier()))
	case "SKIN":
		ex.writeEncodedText(ex.entity.Profile.Skin)
	case "BIRTHDAY":
		ex.writeEncodedText(ex.entity.Profile.Birthday)
	case techLevelExportKey:
		ex.writeEncodedText(ex.entity.Profile.TechLevel)
	case "HAND":
		ex.writeEncodedText(ex.entity.Profile.Handedness)
	case "ST":
		ex.writeEncodedText(ex.entity.Attributes.Current(StrengthID).String())
	case "DX":
		ex.writeEncodedText(ex.entity.Attributes.Current(DexterityID).String())
	case "IQ":
		ex.writeEncodedText(ex.entity.Attributes.Current(IntelligenceID).String())
	case "HT":
		ex.writeEncodedText(ex.entity.Attributes.Current("ht").String())
	case "FP":
		ex.writeEncodedText(ex.entity.Attributes.Current(fpAttrID).String())
	case "BASIC_FP":
		ex.writeEncodedText(ex.entity.Attributes.Maximum(fpAttrID).String())
	case "HP":
		ex.writeEncodedText(ex.entity.Attributes.Current(hpAttrID).String())
	case "BASIC_HP":
		ex.writeEncodedText(ex.entity.Attributes.Maximum(hpAttrID).String())
	case "WILL":
		ex.writeEncodedText(ex.entity.Attributes.Current("will").String())
	case "FRIGHT_CHECK":
		ex.writeEncodedText(ex.entity.Attributes.Current("fright_check").String())
	case "BASIC_SPEED":
		ex.writeEncodedText(ex.entity.Attributes.Current(BasicSpeedID).String())
	case "BASIC_MOVE":
		ex.writeEncodedText(ex.entity.Attributes.Current(BasicMoveID).String())
	case "PERCEPTION":
		ex.writeEncodedText(ex.entity.Attributes.Current("per").String())
	case "VISION":
		ex.writeEncodedText(ex.entity.Attributes.Current("vision").String())
	case "HEARING":
		ex.writeEncodedText(ex.entity.Attributes.Current("hearing").String())
	case "TASTE_SMELL":
		ex.writeEncodedText(ex.entity.Attributes.Current("taste_smell").String())
	case "TOUCH":
		ex.writeEncodedText(ex.entity.Attributes.Current("touch").String())
	case "THRUST":
		ex.writeEncodedText(ex.entity.Thrust().String())
	case "SWING":
		ex.writeEncodedText(ex.entity.Swing().String())
	case "GENERAL_DR":
		dr := 0
		if torso := ex.entity.SheetSettings.BodyType.LookupLocationByID(ex.entity, TorsoID); torso != nil {
			dr = torso.DR(ex.entity, nil, nil)[AllID]
		}
		ex.writeEncodedText(strconv.Itoa(dr))
	case "CURRENT_DODGE":
		ex.writeEncodedText(strconv.Itoa(ex.entity.Dodge(ex.entity.EncumbranceLevel(false))))
	case "CURRENT_MOVE":
		ex.writeEncodedText(strconv.Itoa(ex.entity.Move(ex.entity.EncumbranceLevel(false))))
	case "BEST_CURRENT_PARRY":
		best := "-"
		bestValue := fxp.Min
		for _, w := range ex.entity.EquippedWeapons(true, true) {
			if parry := w.Parry.Resolve(w, nil); parry.CanParry && parry.Modifier > bestValue {
				best = parry.String()
				bestValue = parry.Modifier
			}
		}
		ex.writeEncodedText(best)
	case "BEST_CURRENT_BLOCK":
		best := "-"
		bestValue := fxp.Min
		for _, w := range ex.entity.EquippedWeapons(true, true) {
			if block := w.Block.Resolve(w, nil); block.CanBlock && block.Modifier > bestValue {
				best = block.String()
				bestValue = block.Modifier
			}
		}
		ex.writeEncodedText(best)
	case "TIRED":
		ex.writeEncodedText(ex.entity.Attributes.PoolThreshold(fpAttrID, "tired").String())
	case "FP_COLLAPSE":
		ex.writeEncodedText(ex.entity.Attributes.PoolThreshold(fpAttrID, "collapse").String())
	case "UNCONSCIOUS":
		ex.writeEncodedText(ex.entity.Attributes.PoolThreshold(fpAttrID, "unconscious").String())
	case "REELING":
		ex.writeEncodedText(ex.entity.Attributes.PoolThreshold(hpAttrID, "reeling").String())
	case "HP_COLLAPSE":
		ex.writeEncodedText(ex.entity.Attributes.PoolThreshold(hpAttrID, "collapse").String())
	case "DEATH_CHECK_1":
		ex.writeEncodedText(ex.entity.Attributes.PoolThreshold(hpAttrID, "dying #1").String())
	case "DEATH_CHECK_2":
		ex.writeEncodedText(ex.entity.Attributes.PoolThreshold(hpAttrID, "dying #2").String())
	case "DEATH_CHECK_3":
		ex.writeEncodedText(ex.entity.Attributes.PoolThreshold(hpAttrID, "dying #3").String())
	case "DEATH_CHECK_4":
		ex.writeEncodedText(ex.entity.Attributes.PoolThreshold(hpAttrID, "dying #4").String())
	case "DEAD":
		ex.writeEncodedText(ex.entity.Attributes.PoolThreshold(hpAttrID, "dead").String())
	case "BASIC_LIFT":
		ex.writeEncodedText(ex.entity.SheetSettings.DefaultWeightUnits.Format(ex.entity.BasicLift()))
	case "ONE_HANDED_LIFT":
		ex.writeEncodedText(ex.entity.SheetSettings.DefaultWeightUnits.Format(ex.entity.OneHandedLift()))
	case "TWO_HANDED_LIFT":
		ex.writeEncodedText(ex.entity.SheetSettings.DefaultWeightUnits.Format(ex.entity.TwoHandedLift()))
	case "SHOVE":
		ex.writeEncodedText(ex.entity.SheetSettings.DefaultWeightUnits.Format(ex.entity.ShoveAndKnockOver()))
	case "RUNNING_SHOVE":
		ex.writeEncodedText(ex.entity.SheetSettings.DefaultWeightUnits.Format(ex.entity.RunningShoveAndKnockOver()))
	case "CARRY_ON_BACK":
		ex.writeEncodedText(ex.entity.SheetSettings.DefaultWeightUnits.Format(ex.entity.CarryOnBack()))
	case "SHIFT_SLIGHTLY":
		ex.writeEncodedText(ex.entity.SheetSettings.DefaultWeightUnits.Format(ex.entity.ShiftSlightly()))
	case "CARRIED_WEIGHT":
		ex.writeEncodedText(ex.entity.SheetSettings.DefaultWeightUnits.Format(ex.entity.WeightCarried(false)))
	case "CARRIED_VALUE":
		ex.writeEncodedText("$" + ex.entity.WealthCarried().String())
	case "OTHER_EQUIPMENT_VALUE":
		ex.writeEncodedText("$" + ex.entity.WealthNotCarried().String())
	case "NOTES":
		needBlanks := false
		Traverse(func(n *Note) bool {
			if needBlanks {
				ex.out.WriteString("\n\n")
			} else {
				needBlanks = true
			}
			ex.writeEncodedText(n.String())
			return false
		}, false, false, ex.entity.Notes...)
	case "RACE", "ANCESTRY":
		ex.writeEncodedText(ex.entity.Ancestry().Name)
	case "BODY_TYPE":
		ex.writeEncodedText(ex.entity.SheetSettings.BodyType.Name)
	case "ENCUMBRANCE_LOOP_COUNT":
		ex.writeEncodedText(strconv.Itoa(len(encumbrance.Levels)))
	case "ENCUMBRANCE_LOOP_START":
		ex.processEncumbranceLoop(ex.extractUpToMarker("ENCUMBRANCE_LOOP_END"))
	case "HIT_LOCATION_LOOP_COUNT":
		ex.writeEncodedText(strconv.Itoa(len(ex.entity.SheetSettings.BodyType.Locations)))
	case "HIT_LOCATION_LOOP_START":
		ex.processHitLocationLoop(ex.extractUpToMarker("HIT_LOCATION_LOOP_END"))
	case "ADVANTAGES_LOOP_COUNT":
		ex.writeTraitLoopCount(ex.includeByTraitTags)
	case "ADVANTAGES_LOOP_START":
		ex.processTraitLoop(ex.extractUpToMarker("ADVANTAGES_LOOP_END"), ex.includeByTraitTags)
	case "ADVANTAGES_ALL_LOOP_COUNT":
		ex.writeTraitLoopCount(ex.includeAdvantagesAndPerks)
	case "ADVANTAGES_ALL_LOOP_START":
		ex.processTraitLoop(ex.extractUpToMarker("ADVANTAGES_ALL_LOOP_END"), ex.includeAdvantagesAndPerks)
	case "ADVANTAGES_ONLY_LOOP_COUNT":
		ex.writeTraitLoopCount(ex.includeAdvantages)
	case "ADVANTAGES_ONLY_LOOP_START":
		ex.processTraitLoop(ex.extractUpToMarker("ADVANTAGES_ONLY_LOOP_END"), ex.includeAdvantages)
	case "DISADVANTAGES_LOOP_COUNT":
		ex.writeTraitLoopCount(ex.includeDisadvantages)
	case "DISADVANTAGES_LOOP_START":
		ex.processTraitLoop(ex.extractUpToMarker("DISADVANTAGES_LOOP_END"), ex.includeDisadvantages)
	case "DISADVANTAGES_ALL_LOOP_COUNT":
		ex.writeTraitLoopCount(ex.includeDisadvantagesAndQuirks)
	case "DISADVANTAGES_ALL_LOOP_START":
		ex.processTraitLoop(ex.extractUpToMarker("DISADVANTAGES_ALL_LOOP_END"), ex.includeDisadvantagesAndQuirks)
	case "QUIRKS_LOOP_COUNT":
		ex.writeTraitLoopCount(ex.includeQuirks)
	case "QUIRKS_LOOP_START":
		ex.processTraitLoop(ex.extractUpToMarker("QUIRKS_LOOP_END"), ex.includeQuirks)
	case "PERKS_LOOP_COUNT":
		ex.writeTraitLoopCount(ex.includePerks)
	case "PERKS_LOOP_START":
		ex.processTraitLoop(ex.extractUpToMarker("PERKS_LOOP_END"), ex.includePerks)
	case "LANGUAGES_LOOP_COUNT":
		ex.writeTraitLoopCount(ex.includeLanguages)
	case "LANGUAGES_LOOP_START":
		ex.processTraitLoop(ex.extractUpToMarker("LANGUAGES_LOOP_END"), ex.includeLanguages)
	case "CULTURAL_FAMILIARITIES_LOOP_COUNT":
		ex.writeTraitLoopCount(ex.includeCulturalFamiliarities)
	case "CULTURAL_FAMILIARITIES_LOOP_START":
		ex.processTraitLoop(ex.extractUpToMarker("CULTURAL_FAMILIARITIES_LOOP_END"), ex.includeCulturalFamiliarities)
	case "SKILLS_LOOP_COUNT":
		count := 0
		Traverse(func(_ *Skill) bool {
			count++
			return false
		}, false, true, ex.entity.Skills...)
		ex.writeEncodedText(strconv.Itoa(count))
	case "SKILLS_LOOP_START":
		ex.processSkillsLoop(ex.extractUpToMarker("SKILLS_LOOP_END"))
	case "SPELLS_LOOP_COUNT":
		count := 0
		Traverse(func(_ *Spell) bool {
			count++
			return false
		}, false, false, ex.entity.Spells...)
		ex.writeEncodedText(strconv.Itoa(count))
	case "SPELLS_LOOP_START":
		ex.processSpellsLoop(ex.extractUpToMarker("SPELLS_LOOP_END"))
	case "MELEE_LOOP_COUNT", "HIERARCHICAL_MELEE_LOOP_COUNT":
		ex.writeEncodedText(strconv.Itoa(len(ex.entity.EquippedWeapons(true, true))))
	case "MELEE_LOOP_START":
		ex.processMeleeLoop(ex.extractUpToMarker("MELEE_LOOP_END"))
	case "HIERARCHICAL_MELEE_LOOP_START":
		ex.processHierarchicalMeleeLoop(ex.extractUpToMarker("HIERARCHICAL_MELEE_LOOP_END"))
	case "RANGED_LOOP_COUNT", "HIERARCHICAL_RANGED_LOOP_COUNT":
		ex.writeEncodedText(strconv.Itoa(len(ex.entity.EquippedWeapons(false, true))))
	case "RANGED_LOOP_START":
		ex.processRangedLoop(ex.extractUpToMarker("RANGED_LOOP_END"))
	case "HIERARCHICAL_RANGED_LOOP_START":
		ex.processHierarchicalRangedLoop(ex.extractUpToMarker("HIERARCHICAL_RANGED_LOOP_END"))
	case "EQUIPMENT_LOOP_COUNT":
		count := 0
		Traverse(func(eqp *Equipment) bool {
			if ex.includeByTags(eqp.Tags) {
				count++
			}
			return false
		}, false, false, ex.entity.CarriedEquipment...)
		ex.writeEncodedText(strconv.Itoa(count))
	case "EQUIPMENT_LOOP_START":
		ex.processEquipmentLoop(ex.extractUpToMarker("EQUIPMENT_LOOP_END"), true)
	case "OTHER_EQUIPMENT_LOOP_COUNT":
		count := 0
		Traverse(func(eqp *Equipment) bool {
			if ex.includeByTags(eqp.Tags) {
				count++
			}
			return false
		}, false, false, ex.entity.OtherEquipment...)
		ex.writeEncodedText(strconv.Itoa(count))
	case "OTHER_EQUIPMENT_LOOP_START":
		ex.processEquipmentLoop(ex.extractUpToMarker("EQUIPMENT_LOOP_END"), false)
	case "NOTES_LOOP_COUNT":
		count := 0
		Traverse(func(_ *Note) bool {
			count++
			return false
		}, false, false, ex.entity.Notes...)
		ex.writeEncodedText(strconv.Itoa(count))
	case "NOTES_LOOP_START":
		ex.processNotesLoop(ex.extractUpToMarker("NOTES_LOOP_END"))
	case "REACTION_LOOP_COUNT":
		ex.writeEncodedText(strconv.Itoa(len(ex.entity.Reactions())))
	case "REACTION_LOOP_START":
		ex.processConditionalModifiersLoop(ex.entity.Reactions(), ex.extractUpToMarker("REACTION_LOOP_END"))
	case "CONDITIONAL_MODIFIERS_LOOP_COUNT":
		ex.writeEncodedText(strconv.Itoa(len(ex.entity.ConditionalModifiers())))
	case "CONDITIONAL_MODIFIERS_LOOP_START":
		ex.processConditionalModifiersLoop(ex.entity.ConditionalModifiers(), ex.extractUpToMarker("CONDITIONAL_MODIFIERS_LOOP_END"))
	case "PRIMARY_ATTRIBUTE_LOOP_COUNT":
		count := 0
		for _, def := range ex.entity.SheetSettings.Attributes.List(true) {
			if (def.Type != attribute.Pool && def.Type != attribute.PoolRef) && def.Primary() {
				if _, exists := ex.entity.Attributes.Set[def.DefID]; exists {
					count++
				}
			}
		}
		ex.writeEncodedText(strconv.Itoa(count))
	case "PRIMARY_ATTRIBUTE_LOOP_START":
		ex.processAttributesLoop(ex.extractUpToMarker("PRIMARY_ATTRIBUTE_LOOP_END"), true)
	case "SECONDARY_ATTRIBUTE_LOOP_COUNT":
		count := 0
		for _, def := range ex.entity.SheetSettings.Attributes.List(true) {
			if (def.Type != attribute.Pool && def.Type != attribute.PoolRef) && !def.Primary() {
				if _, exists := ex.entity.Attributes.Set[def.DefID]; exists {
					count++
				}
			}
		}
		ex.writeEncodedText(strconv.Itoa(count))
	case "SECONDARY_ATTRIBUTE_LOOP_START":
		ex.processAttributesLoop(ex.extractUpToMarker("SECONDARY_ATTRIBUTE_LOOP_END"), false)
	case "POINT_POOL_LOOP_COUNT":
		count := 0
		for _, def := range ex.entity.SheetSettings.Attributes.List(true) {
			if def.Type == attribute.Pool || def.Type == attribute.PoolRef {
				if _, exists := ex.entity.Attributes.Set[def.DefID]; exists {
					count++
				}
			}
		}
		ex.writeEncodedText(strconv.Itoa(count))
	case "POINT_POOL_LOOP_START":
		ex.processPointPoolLoop(ex.extractUpToMarker("POINT_POOL_LOOP_END"))
	case "CONTINUE_ID", "CAMPAIGN", "OPTIONS_CODE":
		// No-op
	default:
		switch {
		case strings.HasPrefix(key, "ONLY_CATEGORIES_"):
			splitIntoMap(key, "ONLY_CATEGORIES_", ex.onlyTags)
		case strings.HasPrefix(key, "ONLY_TAGS_"):
			splitIntoMap(key, "ONLY_TAGS_", ex.onlyTags)
		case strings.HasPrefix(key, "EXCLUDE_CATEGORIES_"):
			splitIntoMap(key, "EXCLUDE_CATEGORIES_", ex.excludedTags)
		case strings.HasPrefix(key, "EXCLUDE_TAGS_"):
			splitIntoMap(key, "EXCLUDE_TAGS_", ex.excludedTags)
		case strings.HasPrefix(key, "COLOR_"):
			ex.handleColor(key)
		default:
			attrKey := strings.ToLower(key)
			if attr, ok := ex.entity.Attributes.Set[attrKey]; ok {
				if def := attr.AttributeDef(); def != nil {
					ex.writeEncodedText(attr.Maximum().String())
					return nil
				}
			}
			switch {
			case strings.HasSuffix(attrKey, "_name"):
				if attr, ok := ex.entity.Attributes.Set[attrKey[:len(attrKey)-len("_name")]]; ok {
					if def := attr.AttributeDef(); def != nil {
						ex.writeEncodedText(def.Name)
						return nil
					}
				}
			case strings.HasSuffix(attrKey, "_full_name"):
				if attr, ok := ex.entity.Attributes.Set[attrKey[:len(attrKey)-len("_full_name")]]; ok {
					if def := attr.AttributeDef(); def != nil {
						ex.writeEncodedText(def.ResolveFullName())
						return nil
					}
				}
			case strings.HasSuffix(attrKey, "_combined_name"):
				if attr, ok := ex.entity.Attributes.Set[attrKey[:len(attrKey)-len("_combined_name")]]; ok {
					if def := attr.AttributeDef(); def != nil {
						ex.writeEncodedText(def.CombinedName())
						return nil
					}
				}
			case strings.HasSuffix(attrKey, "_points"):
				if attr, ok := ex.entity.Attributes.Set[attrKey[:len(attrKey)-len("_points")]]; ok {
					if def := attr.AttributeDef(); def != nil {
						ex.writeEncodedText(attr.PointCost().String())
						return nil
					}
				}
			case strings.HasSuffix(attrKey, "_current"):
				if attr, ok := ex.entity.Attributes.Set[attrKey[:len(attrKey)-len("_current")]]; ok {
					if def := attr.AttributeDef(); def != nil {
						ex.writeEncodedText(attr.Current().String())
						return nil
					}
				}
			}
			ex.unidentifiedKey(key)
		}
	}
	return nil
}

func (ex *legacyExporter) unidentifiedKey(key string) {
	ex.writeEncodedText(fmt.Sprintf(`Unidentified key: %q`, key))
}

func splitIntoMap(in, prefix string, m map[string]bool) {
	for _, one := range strings.Split(in[len(prefix):], "_") {
		if one != "" {
			m[one] = true
		}
	}
}

func (ex *legacyExporter) extractUpToMarker(marker string) []byte {
	remaining := ex.template[ex.pos:]
	i := bytes.Index(remaining, []byte(marker))
	if i == -1 {
		ex.pos = len(ex.template)
		return remaining
	}
	buffer := ex.template[ex.pos : ex.pos+i]
	ex.pos += i + len(marker)
	if ex.enhancedKeyParsing && ex.pos < len(ex.template) && ex.template[ex.pos] == '@' {
		ex.pos++
	}
	return buffer
}

func (ex *legacyExporter) writeEncodedText(text string) {
	if ex.encodeText {
		for _, ch := range text {
			switch ch {
			case '<':
				ex.out.WriteString("&lt;")
			case '>':
				ex.out.WriteString("&gt;")
			case '&':
				ex.out.WriteString("&amp;")
			case '"':
				ex.out.WriteString("&quot;")
			case '\'':
				ex.out.WriteString("&apos;")
			case '\n':
				ex.out.WriteString("<br>")
			default:
				if ch >= ' ' && ch <= '~' {
					ex.out.WriteRune(ch)
				} else {
					ex.out.WriteString("&#")
					ex.out.WriteString(strconv.Itoa(int(ch)))
					ex.out.WriteByte(';')
				}
			}
		}
	} else {
		ex.out.WriteString(text)
	}
}

func (ex *legacyExporter) writeTraitLoopCount(f func(*Trait) bool) {
	count := 0
	Traverse(func(t *Trait) bool {
		if f(t) {
			count++
		}
		return false
	}, true, false, ex.entity.Traits...)
	ex.writeEncodedText(strconv.Itoa(count))
}

func (ex *legacyExporter) includeByTags(tags []string) bool {
	for cat := range ex.onlyTags {
		if HasTag(cat, tags) {
			return true
		}
	}
	if len(ex.onlyTags) != 0 {
		return false
	}
	for cat := range ex.excludedTags {
		if HasTag(cat, tags) {
			return false
		}
	}
	return true
}

func (ex *legacyExporter) includeByTraitTags(t *Trait) bool {
	return ex.includeByTags(t.Tags)
}

func (ex *legacyExporter) includeAdvantages(t *Trait) bool {
	return t.AdjustedPoints() > fxp.One && ex.includeByTraitTags(t)
}

func (ex *legacyExporter) includePerks(t *Trait) bool {
	return t.AdjustedPoints() == fxp.One && ex.includeByTraitTags(t)
}

func (ex *legacyExporter) includeAdvantagesAndPerks(t *Trait) bool {
	return t.AdjustedPoints() > 0 && ex.includeByTraitTags(t)
}

func (ex *legacyExporter) includeDisadvantages(t *Trait) bool {
	return t.AdjustedPoints() < -fxp.One && ex.includeByTraitTags(t)
}

func (ex *legacyExporter) includeQuirks(t *Trait) bool {
	return t.AdjustedPoints() == -fxp.One && ex.includeByTraitTags(t)
}

func (ex *legacyExporter) includeDisadvantagesAndQuirks(t *Trait) bool {
	return t.AdjustedPoints() < 0 && ex.includeByTraitTags(t)
}

func (ex *legacyExporter) includeLanguages(t *Trait) bool {
	return HasTag("Language", t.Tags) && ex.includeByTraitTags(t)
}

func (ex *legacyExporter) includeCulturalFamiliarities(t *Trait) bool {
	return strings.HasPrefix(strings.ToLower(t.NameWithReplacements()), "cultural familiarity (") &&
		ex.includeByTraitTags(t)
}

func (ex *legacyExporter) processEncumbranceLoop(buffer []byte) {
	for _, enc := range encumbrance.Levels {
		ex.processBuffer(buffer, func(key string, _ []byte, index int) int {
			switch key {
			case "CURRENT_MARKER":
				if enc == ex.entity.EncumbranceLevel(false) {
					ex.writeEncodedText("current")
				}
			case "CURRENT_MARKER_1":
				if enc == ex.entity.EncumbranceLevel(false) {
					ex.writeEncodedText("1")
				}
			case "CURRENT_MARKER_BULLET":
				if enc == ex.entity.EncumbranceLevel(false) {
					ex.writeEncodedText("•")
				}
			case "LEVEL":
				if enc == ex.entity.EncumbranceLevel(false) {
					ex.writeEncodedText("• ")
				}
				fallthrough
			case "LEVEL_NO_MARKER":
				ex.writeEncodedText(fmt.Sprintf("%s (%s)", enc.String(), (-enc.Penalty()).String()))
			case "LEVEL_ONLY":
				ex.writeEncodedText((-enc.Penalty()).String())
			case "MAX_LOAD":
				ex.writeEncodedText(ex.entity.SheetSettings.DefaultWeightUnits.Format(ex.entity.MaximumCarry(enc)))
			case "MOVE":
				ex.writeEncodedText(strconv.Itoa(ex.entity.Move(enc)))
			case "DODGE":
				ex.writeEncodedText(strconv.Itoa(ex.entity.Dodge(enc)))
			default:
				ex.unidentifiedKey(key)
			}
			return index
		})
	}
}

func (ex *legacyExporter) processHitLocationLoop(buffer []byte) {
	for i, location := range ex.entity.SheetSettings.BodyType.Locations {
		ex.processBuffer(buffer, func(key string, _ []byte, index int) int {
			switch key {
			case idExportKey:
				ex.writeEncodedText(strconv.Itoa(i))
			case "ROLL":
				ex.writeEncodedText(location.RollRange)
			case "WHERE":
				ex.writeEncodedText(location.TableName)
			case "PENALTY":
				ex.writeEncodedText(strconv.Itoa(location.HitPenalty))
			case "DR":
				ex.writeEncodedText(location.DisplayDR(ex.entity, nil))
			case "DR_TOOLTIP":
				var tooltip xio.ByteBuffer
				location.DisplayDR(ex.entity, &tooltip)
				ex.writeEncodedText(tooltip.String())
			case "EQUIPMENT":
				ex.writeEncodedText(strings.Join(ex.hitLocationEquipment(location), ", "))
			case "EQUIPMENT_FORMATTED":
				for _, one := range ex.hitLocationEquipment(location) {
					ex.out.WriteString("<p>")
					ex.writeEncodedText(one)
					ex.out.WriteString("</p>\n")
				}
			default:
				ex.unidentifiedKey(key)
			}
			return index
		})
	}
}

func (ex *legacyExporter) hitLocationEquipment(location *HitLocation) []string {
	var list []string
	Traverse(func(eqp *Equipment) bool {
		if eqp.Equipped {
			for _, f := range eqp.Features {
				if bonus, ok := f.(*DRBonus); ok {
					for _, loc := range bonus.Locations {
						if loc == AllID || strings.EqualFold(location.LocID, loc) {
							list = append(list, eqp.NameWithReplacements())
							break
						}
					}
				}
			}
		}
		return false
	}, false, false, ex.entity.CarriedEquipment...)
	return list
}

func (ex *legacyExporter) processTraitLoop(buffer []byte, f func(*Trait) bool) {
	Traverse(func(t *Trait) bool {
		if f(t) {
			ex.processBuffer(buffer, func(key string, _ []byte, index int) int {
				switch key {
				case idExportKey:
					ex.writeEncodedText(string(t.TID))
				case parentIDExportKey:
					parent := t.Parent()
					if parent != nil {
						ex.writeEncodedText(string(parent.TID))
					}
				case typeExportKey:
					if t.Container() {
						ex.writeEncodedText(strings.ToUpper(t.ContainerType.Key()))
					} else {
						ex.writeEncodedText("ITEM")
					}
				case pointsExportKey:
					ex.writeEncodedText(t.AdjustedPoints().String())
				case descriptionExportKey:
					ex.writeEncodedText(t.String())
					ex.writeNote(t.ModifierNotes())
					ex.writeNote(t.Notes())
				case descriptionPrimaryExportKey:
					ex.writeEncodedText(t.String())
				case "DESCRIPTION_USER":
					ex.writeEncodedText(t.UserDescWithReplacements())
				case "DESCRIPTION_USER_FORMATTED":
					userDesc := t.UserDescWithReplacements()
					if userDesc != "" {
						for _, one := range strings.Split(userDesc, "\n") {
							ex.out.WriteString("<p>")
							ex.writeEncodedText(one)
							ex.out.WriteString("</p>\n")
						}
					}
				case refExportKey:
					ex.writeEncodedText(t.PageRef)
				case styleIndentWarningExportKey:
					ex.handleStyleIndentWarning(t.Depth(), t.UnsatisfiedReason == "")
				case satisfiedExportKey:
					ex.handleSatisfied(t.UnsatisfiedReason == "")
				default:
					switch {
					case strings.HasPrefix(key, "DESCRIPTION_MODIFIER_NOTES"):
						ex.writeWithOptionalParens(key, t.ModifierNotes())
					case strings.HasPrefix(key, "DESCRIPTION_NOTES"):
						ex.writeWithOptionalParens(key, t.Notes())
					case strings.HasPrefix(key, "MODIFIER_NOTES_FOR_"):
						if mod := t.ActiveModifierFor(key[len("MODIFIER_NOTES_FOR_"):]); mod != nil {
							ex.writeEncodedText(mod.LocalNotesWithReplacements())
						}
					case strings.HasPrefix(key, "DEPTHx"):
						ex.handlePrefixDepth(key, t.Depth())
					default:
						ex.unidentifiedKey(key)
					}
				}
				return index
			})
		}
		return false
	}, true, false, ex.entity.Traits...)
	ex.onlyTags = make(map[string]bool)
	ex.excludedTags = make(map[string]bool)
}

func (ex *legacyExporter) processSkillsLoop(buffer []byte) {
	Traverse(func(s *Skill) bool {
		ex.processBuffer(buffer, func(key string, _ []byte, index int) int {
			switch key {
			case idExportKey:
				ex.writeEncodedText(string(s.TID))
			case parentIDExportKey:
				parent := s.Parent()
				if parent != nil {
					ex.writeEncodedText(string(parent.TID))
				}
			case typeExportKey:
				if s.Container() {
					ex.writeEncodedText("GROUP")
				} else {
					ex.writeEncodedText("ITEM")
				}
			case pointsExportKey:
				ex.writeEncodedText(s.AdjustedPoints(nil).String())
			case descriptionExportKey:
				ex.writeEncodedText(s.String())
				ex.writeNote(s.ModifierNotes())
				ex.writeNote(s.Notes())
			case descriptionPrimaryExportKey:
				ex.writeEncodedText(s.String())
			case "SL":
				ex.writeEncodedText(s.CalculateLevel(nil).LevelAsString(s.Container()))
			case "RSL":
				ex.writeEncodedText(s.RelativeLevel())
			case "DIFFICULTY":
				if !s.Container() {
					ex.writeEncodedText(s.Difficulty.Description(EntityFromNode(s)))
				}
			case refExportKey:
				ex.writeEncodedText(s.PageRef)
			case styleIndentWarningExportKey:
				ex.handleStyleIndentWarning(s.Depth(), s.UnsatisfiedReason == "")
			case satisfiedExportKey:
				ex.handleSatisfied(s.UnsatisfiedReason == "")
			default:
				switch {
				case strings.HasPrefix(key, "DESCRIPTION_MODIFIER_NOTES"):
					ex.writeWithOptionalParens(key, s.ModifierNotes())
				case strings.HasPrefix(key, "DESCRIPTION_NOTES"):
					ex.writeWithOptionalParens(key, s.Notes())
				case strings.HasPrefix(key, "DEPTHx"):
					ex.handlePrefixDepth(key, s.Depth())
				default:
					ex.unidentifiedKey(key)
				}
			}
			return index
		})
		return false
	}, false, false, ex.entity.Skills...)
}

func (ex *legacyExporter) processSpellsLoop(buffer []byte) {
	Traverse(func(s *Spell) bool {
		ex.processBuffer(buffer, func(key string, _ []byte, index int) int {
			switch key {
			case idExportKey:
				ex.writeEncodedText(string(s.TID))
			case parentIDExportKey:
				parent := s.Parent()
				if parent != nil {
					ex.writeEncodedText(string(parent.TID))
				}
			case typeExportKey:
				if s.Container() {
					ex.writeEncodedText("GROUP")
				} else {
					ex.writeEncodedText("ITEM")
				}
			case pointsExportKey:
				ex.writeEncodedText(s.AdjustedPoints(nil).String())
			case descriptionExportKey:
				ex.writeEncodedText(s.String())
				ex.writeNote(s.Notes())
				ex.writeNote(s.Rituals())
			case descriptionPrimaryExportKey:
				ex.writeEncodedText(s.String())
			case "SL":
				ex.writeEncodedText(s.CalculateLevel().LevelAsString(s.Container()))
			case "RSL":
				ex.writeEncodedText(s.RelativeLevel())
			case "DIFFICULTY":
				if !s.Container() {
					ex.writeEncodedText(s.Difficulty.Description(EntityFromNode(s)))
				}
			case "CLASS":
				ex.writeEncodedText(s.ClassWithReplacements())
			case "COLLEGE":
				ex.writeEncodedText(strings.Join(s.CollegeWithReplacements(), ", "))
			case "MANA_CAST":
				ex.writeEncodedText(s.CastingCostWithReplacements())
			case "MANA_MAINTAIN":
				ex.writeEncodedText(s.MaintenanceCostWithReplacements())
			case "TIME_CAST":
				ex.writeEncodedText(s.CastingTimeWithReplacements())
			case "DURATION":
				ex.writeEncodedText(s.DurationWithReplacements())
			case "RESIST":
				ex.writeEncodedText(s.ResistWithReplacements())
			case refExportKey:
				ex.writeEncodedText(s.PageRef)
			case styleIndentWarningExportKey:
				ex.handleStyleIndentWarning(s.Depth(), s.UnsatisfiedReason == "")
			case satisfiedExportKey:
				ex.handleSatisfied(s.UnsatisfiedReason == "")
			default:
				switch {
				case strings.HasPrefix(key, "DESCRIPTION_MODIFIER_NOTES"):
					// Here for legacy reasons. Spells have never had these notes.
				case strings.HasPrefix(key, "DESCRIPTION_NOTES"):
					notes := s.Notes()
					rituals := s.Rituals()
					if rituals != "" {
						if strings.TrimSpace(notes) != "" {
							notes += "; "
						}
						notes += rituals
					}
					ex.writeWithOptionalParens(key, notes)
				case strings.HasPrefix(key, "DEPTHx"):
					ex.handlePrefixDepth(key, s.Depth())
				default:
					ex.unidentifiedKey(key)
				}
			}
			return index
		})
		return false
	}, false, false, ex.entity.Spells...)
}

func (ex *legacyExporter) processEquipmentLoop(buffer []byte, carried bool) {
	var eqpList []*Equipment
	if carried {
		eqpList = ex.entity.CarriedEquipment
	} else {
		eqpList = ex.entity.OtherEquipment
	}
	Traverse(func(eqp *Equipment) bool {
		if ex.includeByTags(eqp.Tags) {
			ex.processBuffer(buffer, func(key string, _ []byte, index int) int {
				switch key {
				case idExportKey:
					ex.writeEncodedText(string(eqp.TID))
				case parentIDExportKey:
					parent := eqp.Parent()
					if parent != nil {
						ex.writeEncodedText(string(parent.TID))
					}
				case typeExportKey:
					if eqp.Container() {
						ex.writeEncodedText("GROUP")
					} else {
						ex.writeEncodedText("ITEM")
					}
				case descriptionExportKey:
					ex.writeEncodedText(eqp.String())
					ex.writeNote(eqp.ModifierNotes())
					ex.writeNote(eqp.Notes())
				case descriptionPrimaryExportKey:
					ex.writeEncodedText(eqp.String())
				case refExportKey:
					ex.writeEncodedText(eqp.PageRef)
				case styleIndentWarningExportKey:
					ex.handleStyleIndentWarning(eqp.Depth(), eqp.UnsatisfiedReason == "")
				case satisfiedExportKey:
					ex.handleSatisfied(eqp.UnsatisfiedReason == "")
				case "STATE":
					switch {
					case !carried:
						ex.writeEncodedText("-")
					case eqp.Equipped:
						ex.writeEncodedText("E")
					default:
						ex.writeEncodedText("C")
					}
				case "EQUIPPED":
					if carried && eqp.Equipped {
						ex.writeEncodedText("✓")
					}
				case "EQUIPPED_FA":
					if carried && eqp.Equipped {
						ex.out.WriteString(`<i class="fas fa-check-circle"></i>`)
					}
				case "EQUIPPED_NUM":
					if carried && eqp.Equipped {
						ex.writeEncodedText("1")
					} else {
						ex.writeEncodedText("0")
					}
				case "CARRIED_STATUS":
					switch {
					case !carried:
						ex.writeEncodedText("0")
					case eqp.Equipped:
						ex.writeEncodedText("2")
					default:
						ex.writeEncodedText("1")
					}
				case "QTY":
					ex.writeEncodedText(eqp.Quantity.String())
				case "COST":
					ex.writeEncodedText(eqp.AdjustedValue().String())
				case weightExportKey:
					ex.writeEncodedText(ex.entity.SheetSettings.DefaultWeightUnits.Format(eqp.AdjustedWeight(false, ex.entity.SheetSettings.DefaultWeightUnits)))
				case "COST_SUMMARY":
					ex.writeEncodedText(eqp.ExtendedValue().String())
				case "WEIGHT_SUMMARY":
					ex.writeEncodedText(ex.entity.SheetSettings.DefaultWeightUnits.Format(eqp.ExtendedWeight(false, ex.entity.SheetSettings.DefaultWeightUnits)))
				case "WEIGHT_RAW":
					ex.writeEncodedText(fxp.Int(eqp.AdjustedWeight(false, ex.entity.SheetSettings.DefaultWeightUnits)).String())
				case techLevelExportKey:
					ex.writeEncodedText(eqp.TechLevel)
				case "LEGALITY_CLASS", "LC":
					ex.writeEncodedText(eqp.LegalityClass)
				case "TAGS", "CATEGORIES":
					ex.writeEncodedText(CombineTags(eqp.Tags))
				case "LOCATION":
					parent := eqp.Parent()
					if parent != nil {
						ex.writeEncodedText(parent.NameWithReplacements())
					}
				case "USES":
					ex.writeEncodedText(strconv.Itoa(eqp.Uses))
				case "MAX_USES":
					ex.writeEncodedText(strconv.Itoa(eqp.MaxUses))
				default:
					switch {
					case strings.HasPrefix(key, "DESCRIPTION_MODIFIER_NOTES"):
						ex.writeWithOptionalParens(key, eqp.ModifierNotes())
					case strings.HasPrefix(key, "DESCRIPTION_NOTES"):
						ex.writeWithOptionalParens(key, eqp.Notes())
					case strings.HasPrefix(key, "MODIFIER_NOTES_FOR_"):
						if mod := eqp.ActiveModifierFor(key[len("MODIFIER_NOTES_FOR_"):]); mod != nil {
							ex.writeEncodedText(mod.LocalNotesWithReplacements())
						}
					case strings.HasPrefix(key, "DEPTHx"):
						ex.handlePrefixDepth(key, eqp.Depth())
					default:
						ex.unidentifiedKey(key)
					}
				}
				return index
			})
		}
		return false
	}, false, false, eqpList...)
	ex.onlyTags = make(map[string]bool)
	ex.excludedTags = make(map[string]bool)
}

func (ex *legacyExporter) processNotesLoop(buffer []byte) {
	Traverse(func(n *Note) bool {
		ex.processBuffer(buffer, func(key string, _ []byte, index int) int {
			switch key {
			case idExportKey:
				ex.writeEncodedText(string(n.TID))
			case parentIDExportKey:
				parent := n.Parent()
				if parent != nil {
					ex.writeEncodedText(string(parent.TID))
				}
			case typeExportKey:
				if n.Container() {
					ex.writeEncodedText("GROUP")
				} else {
					ex.writeEncodedText("ITEM")
				}
			case refExportKey:
				ex.writeEncodedText(n.PageRef)
			case "NOTE":
				ex.writeEncodedText(n.String())
			case "NOTE_FORMATTED":
				s := n.String()
				if strings.TrimSpace(s) != "" {
					for _, one := range strings.Split(s, "\n") {
						ex.out.WriteString("<p>")
						ex.writeEncodedText(one)
						ex.out.WriteString("</p>\n")
					}
				}
			case styleIndentWarningExportKey:
				ex.handleStyleIndentWarning(n.Depth(), true)
			default:
				switch {
				case strings.HasPrefix(key, "DEPTHx"):
					ex.handlePrefixDepth(key, n.Depth())
				default:
					ex.unidentifiedKey(key)
				}
			}
			return index
		})
		return false
	}, false, false, ex.entity.Notes...)
}

func (ex *legacyExporter) processConditionalModifiersLoop(list []*ConditionalModifier, buffer []byte) {
	for i, one := range list {
		ex.processBuffer(buffer, func(key string, _ []byte, index int) int {
			switch key {
			case idExportKey:
				ex.writeEncodedText(strconv.Itoa(i))
			case "MODIFIER":
				ex.writeEncodedText(one.Total().StringWithSign())
			case "SITUATION":
				ex.writeEncodedText(one.From)
			default:
				ex.unidentifiedKey(key)
			}
			return index
		})
	}
}

func (ex *legacyExporter) processAttributesLoop(buffer []byte, primary bool) {
	for _, def := range ex.entity.SheetSettings.Attributes.List(true) {
		if (def.Type != attribute.Pool && def.Type != attribute.PoolRef) && def.Primary() == primary {
			if attr, ok := ex.entity.Attributes.Set[def.DefID]; ok {
				ex.processBuffer(buffer, func(key string, _ []byte, index int) int {
					switch key {
					case idExportKey:
						ex.writeEncodedText(def.DefID)
					case nameExportKey:
						ex.writeEncodedText(def.Name)
					case "FULL_NAME":
						ex.writeEncodedText(def.ResolveFullName())
					case "COMBINED_NAME":
						ex.writeEncodedText(def.CombinedName())
					case "VALUE":
						ex.writeEncodedText(attr.Maximum().String())
					case "POINTS":
						ex.writeEncodedText(attr.PointCost().String())
					default:
						ex.unidentifiedKey(key)
					}
					return index
				})
			}
		}
	}
}

func (ex *legacyExporter) processPointPoolLoop(buffer []byte) {
	for _, def := range ex.entity.SheetSettings.Attributes.List(true) {
		if def.Type == attribute.Pool || def.Type == attribute.PoolRef {
			if attr, ok := ex.entity.Attributes.Set[def.DefID]; ok {
				ex.processBuffer(buffer, func(key string, _ []byte, index int) int {
					switch key {
					case idExportKey:
						ex.writeEncodedText(def.DefID)
					case nameExportKey:
						ex.writeEncodedText(def.Name)
					case "FULL_NAME":
						ex.writeEncodedText(def.ResolveFullName())
					case "COMBINED_NAME":
						ex.writeEncodedText(def.CombinedName())
					case "CURRENT":
						ex.writeEncodedText(attr.Current().String())
					case "MAXIMUM":
						ex.writeEncodedText(attr.Maximum().String())
					case "POINTS":
						ex.writeEncodedText(attr.PointCost().String())
					default:
						ex.unidentifiedKey(key)
					}
					return index
				})
			}
		}
	}
}

func (ex *legacyExporter) processMeleeLoop(buffer []byte) {
	for i, w := range ex.entity.EquippedWeapons(true, true) {
		ex.processBuffer(buffer, func(key string, buf []byte, index int) int {
			return ex.processMeleeKeys(key, i, w, nil, buf, index)
		})
	}
}

func (ex *legacyExporter) processHierarchicalMeleeLoop(buffer []byte) {
	m := make(map[string][]*Weapon)
	for _, w := range ex.entity.EquippedWeapons(true, true) {
		key := w.String()
		m[key] = append(m[key], w)
	}
	list := make([]*Weapon, 0, len(m))
	for _, v := range m {
		list = append(list, v[0])
	}
	slices.SortFunc(list, func(a, b *Weapon) int { return a.Compare(b) })
	for i, w := range list {
		ex.processBuffer(buffer, func(key string, buf []byte, index int) int {
			return ex.processMeleeKeys(key, i, w, m[w.String()], buf, index)
		})
	}
}

func (ex *legacyExporter) processRangedLoop(buffer []byte) {
	for i, w := range ex.entity.EquippedWeapons(false, true) {
		ex.processBuffer(buffer, func(key string, buf []byte, index int) int {
			return ex.processRangedKeys(key, i, w, nil, buf, index)
		})
	}
}

func (ex *legacyExporter) processHierarchicalRangedLoop(buffer []byte) {
	m := make(map[string][]*Weapon)
	for _, w := range ex.entity.EquippedWeapons(false, true) {
		key := w.String()
		m[key] = append(m[key], w)
	}
	list := make([]*Weapon, 0, len(m))
	for _, v := range m {
		list = append(list, v[0])
	}
	slices.SortFunc(list, func(a, b *Weapon) int { return a.Compare(b) })
	for i, w := range list {
		ex.processBuffer(buffer, func(key string, buf []byte, index int) int {
			return ex.processRangedKeys(key, i, w, m[w.String()], buf, index)
		})
	}
}

func (ex *legacyExporter) processMeleeKeys(key string, currentID int, w *Weapon, attackModes []*Weapon, buf []byte, index int) int {
	switch key {
	case "PARRY":
		ex.writeEncodedText(w.Parry.Resolve(w, nil).String())
	case "BLOCK":
		ex.writeEncodedText(w.Block.Resolve(w, nil).String())
	case "REACH":
		ex.writeEncodedText(w.Reach.Resolve(w, nil).String())
	case "ATTACK_MODES_LOOP_COUNT":
		ex.writeEncodedText(strconv.Itoa(len(attackModes)))
	case "ATTACK_MODES_LOOP_START":
		if len(attackModes) != 0 {
			buf, index = ex.subBufferExtractUpToMarker("ATTACK_MODES_LOOP_END", buf, index)
			for i, mode := range attackModes {
				ex.processBuffer(buf, func(key string, innerBuf []byte, innerIndex int) int {
					return ex.processMeleeKeys(key, i, mode, nil, innerBuf, innerIndex)
				})
			}
		} else {
			ex.unidentifiedKey(key)
		}
	default:
		ex.processWeaponKeys(key, currentID, w)
	}
	return index
}

func (ex *legacyExporter) processRangedKeys(key string, currentID int, w *Weapon, attackModes []*Weapon, buf []byte, index int) int {
	switch key {
	case "BULK":
		ex.writeEncodedText(w.Bulk.Resolve(w, nil).String())
	case "ACCURACY":
		ex.writeEncodedText(w.Accuracy.Resolve(w, nil).String())
	case "RANGE":
		ex.writeEncodedText(w.Range.Resolve(w, nil).String(true))
	case "ROF":
		ex.writeEncodedText(w.RateOfFire.Resolve(w, nil).String())
	case "SHOTS":
		ex.writeEncodedText(w.Shots.Resolve(w, nil).String())
	case "RECOIL":
		ex.writeEncodedText(w.Recoil.Resolve(w, nil).String())
	case "ATTACK_MODES_LOOP_COUNT":
		ex.writeEncodedText(strconv.Itoa(len(attackModes)))
	case "ATTACK_MODES_LOOP_START":
		if len(attackModes) != 0 {
			buf, index = ex.subBufferExtractUpToMarker("ATTACK_MODES_LOOP_END", buf, index)
			for i, mode := range attackModes {
				ex.processBuffer(buf, func(key string, innerBuf []byte, innerIndex int) int {
					return ex.processRangedKeys(key, i, mode, nil, innerBuf, innerIndex)
				})
			}
		} else {
			ex.unidentifiedKey(key)
		}
	default:
		ex.processWeaponKeys(key, currentID, w)
	}
	return index
}

func (ex *legacyExporter) processWeaponKeys(key string, currentID int, w *Weapon) {
	switch key {
	case idExportKey:
		ex.writeEncodedText(strconv.Itoa(currentID))
	case descriptionExportKey:
		ex.writeEncodedText(w.String())
		ex.writeNote(w.Notes())
	case descriptionPrimaryExportKey:
		ex.writeEncodedText(w.String())
	case "USAGE":
		ex.writeEncodedText(w.UsageWithReplacements())
	case "LEVEL":
		ex.writeEncodedText(w.SkillLevel(nil).String())
	case "DAMAGE":
		ex.writeEncodedText(w.Damage.ResolvedDamage(nil))
	case "UNMODIFIED_DAMAGE":
		ex.writeEncodedText(w.Damage.String())
	case "STRENGTH":
		ex.writeEncodedText(w.Strength.Resolve(w, nil).String())
	case "WEAPON_STRENGTH":
		st := w.Strength.Resolve(w, nil).Min
		if st > 0 {
			ex.writeEncodedText(st.String())
		}
	case "COST":
		if eqp, ok := w.Owner.(*Equipment); ok {
			ex.writeEncodedText(eqp.AdjustedValue().String())
		}
	case "LEGALITY_CLASS", "LC":
		if eqp, ok := w.Owner.(*Equipment); ok {
			ex.writeEncodedText(eqp.LegalityClass)
		}
	case techLevelExportKey:
		if eqp, ok := w.Owner.(*Equipment); ok {
			ex.writeEncodedText(eqp.TechLevel)
		}
	case weightExportKey:
		if eqp, ok := w.Owner.(*Equipment); ok {
			ex.writeEncodedText(ex.entity.SheetSettings.DefaultWeightUnits.Format(eqp.AdjustedWeight(false, ex.entity.SheetSettings.DefaultWeightUnits)))
		}
	case "AMMO":
		if eqp, ok := w.Owner.(*Equipment); ok {
			ex.writeEncodedText(ex.ammoFor(eqp).String())
		}
	default:
		switch {
		case strings.HasPrefix(key, "DESCRIPTION_NOTES"):
			ex.writeWithOptionalParens(key, w.Notes())
		default:
			ex.unidentifiedKey(key)
		}
	}
}

func (ex *legacyExporter) ammoFor(weaponEqp *Equipment) fxp.Int {
	uses := ""
	for _, cat := range weaponEqp.TagList() {
		if strings.HasPrefix(strings.ToLower(cat), "usesammotype:") {
			uses = strings.ReplaceAll(cat[len("usesammotype:"):], " ", "")
			break
		}
	}
	if uses == "" {
		return 0
	}
	var total fxp.Int
	Traverse(func(eqp *Equipment) bool {
		if eqp.Equipped && eqp.Quantity > 0 {
			for _, cat := range eqp.Tags {
				if strings.HasPrefix(strings.ToLower(cat), "ammotype:") {
					if uses == strings.ReplaceAll(cat[len("ammotype:"):], " ", "") {
						total += eqp.Quantity
						break
					}
				}
			}
		}
		return false
	}, false, false, ex.entity.CarriedEquipment...)
	return total
}

func (ex *legacyExporter) handleStyleIndentWarning(depth int, satisfied bool) {
	var style strings.Builder
	if depth > 0 {
		fmt.Fprintf(&style, "padding-left: %dpx;", depth*12)
	}
	if !satisfied {
		style.WriteString("color: red;")
	}
	if style.Len() != 0 {
		ex.out.WriteString(` style="`)
		ex.out.WriteString(style.String())
		ex.out.WriteString(`" `)
	}
}

func (ex *legacyExporter) handleSatisfied(satisfied bool) {
	if satisfied {
		ex.writeEncodedText("Y")
	} else {
		ex.writeEncodedText("N")
	}
}

func (ex *legacyExporter) handlePrefixDepth(key string, depth int) {
	if amt, err := strconv.Atoi(key[len("DEPTHx"):]); err != nil {
		ex.unidentifiedKey(key)
	} else {
		ex.writeEncodedText(strconv.Itoa(amt * depth))
	}
}

func (ex *legacyExporter) writeNote(note string) {
	if strings.TrimSpace(note) != "" {
		ex.out.WriteString(`<div class="note">`)
		ex.writeEncodedText(note)
		ex.out.WriteString("</div>")
	}
}

func (ex *legacyExporter) writeWithOptionalParens(key, text string) {
	if strings.TrimSpace(text) != "" {
		var pre, post string
		switch {
		case strings.HasSuffix(key, "_PAREN"):
			pre = " ("
			post = ")"
		case strings.HasSuffix(key, "_BRACKET"):
			pre = " ["
			post = "]"
		case strings.HasSuffix(key, "_CURLY"):
			pre = " {"
			post = "}"
		}
		ex.out.WriteString(pre)
		ex.out.WriteString(text)
		ex.out.WriteString(post)
	}
}

func (ex *legacyExporter) processBuffer(buffer []byte, f func(key string, buf []byte, index int) int) {
	var keyBuffer bytes.Buffer
	lookForKeyMarker := true
	i := 0
	for i < len(buffer) {
		ch := buffer[i]
		i++
		switch {
		case lookForKeyMarker:
			var next byte
			if ex.pos < len(ex.template) {
				next = ex.template[ex.pos]
			}
			if ch == '@' && (next < '0' || next > '9') {
				lookForKeyMarker = false
			} else {
				ex.out.WriteByte(ch)
			}
		case ch == '_' || (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z'):
			keyBuffer.WriteByte(ch)
		default:
			if !ex.enhancedKeyParsing || ch != '@' {
				i--
			}
			i = f(keyBuffer.String(), buffer, i)
			keyBuffer.Reset()
			lookForKeyMarker = true
		}
	}
}

func (ex *legacyExporter) subBufferExtractUpToMarker(marker string, buf []byte, start int) (buffer []byte, newStart int) {
	remaining := buf[start:]
	i := bytes.Index(remaining, []byte(marker))
	if i == -1 {
		return remaining, len(buf)
	}
	buffer = buf[start : start+i]
	start += i + len(marker)
	if ex.enhancedKeyParsing && start < len(buf) && buf[start] == '@' {
		start++
	}
	return buffer, start
}

func (ex *legacyExporter) handleColor(key string) {
	id := strings.ToLower(key[len("COLOR_"):])
	for _, c := range colors.Current() {
		if c.ID == id {
			ex.out.WriteString(c.Color.GetColor().String())
			return
		}
	}
	ex.unidentifiedKey(key)
}
