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
	"encoding/json/jsontext"
	"encoding/json/v2"
	"io/fs"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/progression"
	"github.com/richardwilkes/gcs/v5/model/jio"
)

// SheetSettingsResponder defines the method required to be notified of updates to the SheetSettings.
type SheetSettingsResponder interface {
	// SheetSettingsUpdated will be called when the SheetSettings have been updated. The provided Entity will be nil if
	// it was the default SheetSettings that was updated rather than one attached to a specific entity. blockLayout will
	// be true if the BlockLayout was altered, which usually requires a full rebuild.
	SheetSettingsUpdated(entity *Entity, blockLayout bool)
}

// SheetSettingsData holds the SheetSettings data that is written to disk.
type SheetSettingsData struct {
	Page                          *PageSettings      `json:"page,omitzero"`
	BlockLayout                   *BlockLayout       `json:"block_layout,omitzero"`
	Attributes                    *AttributeDefs     `json:"attributes,omitzero"`
	BodyType                      *Body              `json:"body_type,omitzero"`
	DamageProgression             progression.Option `json:"damage_progression"`
	DefaultLengthUnits            fxp.LengthUnit     `json:"default_length_units"`
	DefaultWeightUnits            fxp.WeightUnit     `json:"default_weight_units"`
	UserDescriptionDisplay        display.Option     `json:"user_description_display"`
	ModifiersDisplay              display.Option     `json:"modifiers_display"`
	NotesDisplay                  display.Option     `json:"notes_display"`
	SkillLevelAdjDisplay          display.Option     `json:"skill_level_adj_display"`
	UseMultiplicativeModifiers    bool               `json:"use_multiplicative_modifiers,omitzero"`
	UseModifyingDicePlusAdds      bool               `json:"use_modifying_dice_plus_adds,omitzero"`
	UseHalfStatDefaults           bool               `json:"use_half_stat_defaults,omitzero"`
	ShowTraitModifierAdj          bool               `json:"show_trait_modifier_adj,omitzero"`
	ShowEquipmentModifierAdj      bool               `json:"show_equipment_modifier_adj,omitzero"`
	ShowAllWeapons                bool               `json:"show_all_weapons,omitzero"`
	ShowSpellAdj                  bool               `json:"show_spell_adj,omitzero"`
	HideSourceMismatch            bool               `json:"hide_source_mismatch,omitzero"`
	HideTLColumn                  bool               `json:"hide_tl_column,omitzero"`
	HideLCColumn                  bool               `json:"hide_lc_column,omitzero"`
	HidePageRefColumn             bool               `json:"hide_page_ref_column,omitzero"`
	UseTitleInFooter              bool               `json:"use_title_in_footer,omitzero"`
	ExcludeUnspentPointsFromTotal bool               `json:"exclude_unspent_points_from_total,omitzero"`
	ShowLiftingSTDamage           bool               `json:"show_lifting_st_damage,omitzero"`
	ShowIQBasedDamage             bool               `json:"show_iq_based_damage,omitzero"`
	UseSkillModifierAdjustments   bool               `json:"use_skill_modifier_adjustments,omitzero"`
	EasySkillModifierOverride             fxp.Int            `json:"easy_skill_modifier_override,omitzero"`
	AverageSkillModifierOverride          fxp.Int            `json:"average_skill_modifier_override,omitzero"`
	HardSkillModifierOverride             fxp.Int            `json:"hard_skill_modifier_override,omitzero"`
	VeryHardSkillModifierOverride         fxp.Int            `json:"very_hard_skill_modifier_override,omitzero"`
	EasySkillModifierAdjustment          fxp.Int            `json:"easy_skill_modifier_adjustment,omitzero"`
	AverageSkillModifierAdjustment       fxp.Int            `json:"average_skill_modifier_adjustment,omitzero"`
	HardSkillModifierAdjustment          fxp.Int            `json:"hard_skill_modifier_adjustment,omitzero"`
	VeryHardSkillModifierAdjustment      fxp.Int            `json:"very_hard_skill_modifier_adjustment,omitzero"`
	UseBasicMoveForDodge                 bool               `json:"use_basic_move_for_dodge,omitzero"`
	IncludeDodgeFlatBonus                bool               `json:"include_dodge_flat_bonus,omitzero"`
	IncludePDArmor                       bool               `json:"include_pd_armor,omitzero"`
	IncludePDShields                     bool               `json:"include_pd_shields,omitzero"`
	UsePassiveDefense                    bool               `json:"use_passive_defense,omitzero"` // GURPS 3e optional rule: PD applies when active defense fails (also shows PD column)
	ShowPDColumn                         bool               `json:"show_pd_column,omitzero"`      // DEPRECATED: Automatically synced with UsePassiveDefense in EnsureValidity(). Kept for backward compatibility with old character sheets.
	DodgeOverride                        fxp.Int            `json:"dodge_override,omitzero"`
}

// SheetSettings holds sheet settings.
type SheetSettings struct {
	SheetSettingsData
	Entity *Entity `json:"-"`
}

// SheetSettingsFor returns the SheetSettings for the given Entity, or the global settings if the Entity is nil.
func SheetSettingsFor(entity *Entity) *SheetSettings {
	if entity == nil {
		return GlobalSettings().SheetSettings()
	}
	return entity.SheetSettings
}

// FactorySheetSettings returns a new SheetSettings with factory defaults.
func FactorySheetSettings() *SheetSettings {
	return &SheetSettings{
		SheetSettingsData: SheetSettingsData{
			Page:                   NewPageSettings(),
			BlockLayout:            NewBlockLayout(),
			Attributes:             FactoryAttributeDefs(),
			BodyType:               FactoryBody(),
			DamageProgression:      progression.BasicSet,
			DefaultLengthUnits:     fxp.FeetAndInches,
			DefaultWeightUnits:     fxp.Pound,
			UserDescriptionDisplay: display.Tooltip,
			ModifiersDisplay:       display.Inline,
			NotesDisplay:           display.Inline,
			SkillLevelAdjDisplay:   display.Tooltip,
			ShowSpellAdj:           true,
			// GURPS 4E defaults: Use Basic Speed, include flat +3, no PD
			UseBasicMoveForDodge:  false,
			IncludeDodgeFlatBonus: true,
			IncludePDArmor:        false,
			IncludePDShields:       false,
			UsePassiveDefense:      false, // PD is a GURPS 3e optional rule, disabled by default (automatically shows PD column when enabled)
		},
	}
}

// NewSheetSettingsFromFile loads new settings from a file.
func NewSheetSettingsFromFile(fileSystem fs.FS, filePath string) (*SheetSettings, error) {
	var data struct {
		SheetSettings
		OldLocation *SheetSettings `json:"sheet_settings"`
	}
	if err := jio.Load(fileSystem, filePath, &data); err != nil {
		return nil, err
	}
	var s *SheetSettings
	if data.OldLocation != nil {
		s = data.OldLocation
	} else {
		ss := data.SheetSettings
		s = &ss
	}
	s.EnsureValidity()
	return s, nil
}

// EnsureValidity checks the current settings for validity and if they aren't valid, makes them so.
func (s *SheetSettings) EnsureValidity() {
	if s.Page == nil {
		s.Page = NewPageSettings()
	} else {
		s.Page.EnsureValidity()
	}
	if s.BlockLayout == nil {
		s.BlockLayout = NewBlockLayout()
	} else {
		s.BlockLayout.EnsureValidity()
	}
	if s.Attributes == nil {
		s.Attributes = FactoryAttributeDefs()
	}
	if s.BodyType == nil {
		s.BodyType = FactoryBody()
	}
	s.DamageProgression = s.DamageProgression.EnsureValid()
	s.DefaultLengthUnits = s.DefaultLengthUnits.EnsureValid()
	s.DefaultWeightUnits = s.DefaultWeightUnits.EnsureValid()
	s.UserDescriptionDisplay = s.UserDescriptionDisplay.EnsureValid()
	s.ModifiersDisplay = s.ModifiersDisplay.EnsureValid()
	s.NotesDisplay = s.NotesDisplay.EnsureValid()
	s.SkillLevelAdjDisplay = s.SkillLevelAdjDisplay.EnsureValid()
	// Ensure GURPS 4E defaults for dodge calculation fields
	// This handles backward compatibility for character sheets created before dodge customization was added.
	// We use a conservative heuristic: only set defaults if BOTH dodge fields AND skill modifier fields
	// are at their zero values, which strongly indicates an old character sheet where these fields
	// were never present in the JSON (and thus defaulted to zero values).
	// This avoids incorrectly setting defaults if a user explicitly sets all dodge fields to false
	// in a new character sheet (which would be very unusual anyway).
	// NOTE: PD (Passive Defense) fields are not checked here as PD does not affect base Dodge.
	// PD is a separate mechanic that applies during combat resolution when an active defense fails.
	dodgeFieldsAtDefaults := !s.IncludeDodgeFlatBonus && !s.UseBasicMoveForDodge
	skillModifierFieldsAtDefaults := !s.UseSkillModifierAdjustments &&
		s.EasySkillModifierOverride == 0 && s.AverageSkillModifierOverride == 0 &&
		s.HardSkillModifierOverride == 0 && s.VeryHardSkillModifierOverride == 0 &&
		s.EasySkillModifierAdjustment == 0 && s.AverageSkillModifierAdjustment == 0 &&
		s.HardSkillModifierAdjustment == 0 && s.VeryHardSkillModifierAdjustment == 0
	if dodgeFieldsAtDefaults && skillModifierFieldsAtDefaults {
		// Both feature sets at zero values - very likely an old character sheet, set GURPS 4E defaults
		s.IncludeDodgeFlatBonus = true // GURPS 4E includes flat +3 bonus
		// Other fields are already false, which matches GURPS 4E defaults
	}
	// Ensure ShowPDColumn is always synced with UsePassiveDefense
	s.ShowPDColumn = s.UsePassiveDefense
}

// MarshalJSONTo implements json.MarshalerTo.
func (s *SheetSettings) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, &s.SheetSettingsData)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (s *SheetSettings) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var content struct {
		SheetSettingsData
		OldBodyType             *Body `json:"hit_locations"`
		OldShowTraitModifierAdj bool  `json:"show_advantage_modifier_adj"`
	}
	if err := json.UnmarshalDecode(dec, &content); err != nil {
		return err
	}
	s.SheetSettingsData = content.SheetSettingsData
	if s.BodyType == nil && content.OldBodyType != nil {
		s.BodyType = content.OldBodyType
	}
	if !s.ShowTraitModifierAdj && content.OldShowTraitModifierAdj {
		s.ShowTraitModifierAdj = true
	}
	s.EnsureValidity()
	return nil
}

// Clone creates a copy of this.
func (s *SheetSettings) Clone(entity *Entity) *SheetSettings {
	clone := *s
	clone.Page = s.Page.Clone()
	clone.BlockLayout = s.BlockLayout.Clone()
	clone.Attributes = s.Attributes.Clone()
	clone.BodyType = s.BodyType.Clone(entity, nil)
	return &clone
}

// SetOwningEntity sets the owning entity and configures any sub-components as needed.
func (s *SheetSettings) SetOwningEntity(entity *Entity) {
	s.Entity = entity
	s.BodyType.Update(entity)
}

// Save writes the settings to the file as JSON.
func (s *SheetSettings) Save(filePath string) error {
	return jio.SaveToFile(filePath, s)
}
