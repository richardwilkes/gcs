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
	"encoding/json/jsontext"
	"encoding/json/v2"
	"io/fs"
	"slices"
	"sync/atomic"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/progression"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/tid"
)

// SheetSettingsParts is a set of the parts of the SheetSettings, which a change reports to say what it altered, so
// that a responder showing only some of the settings can ignore a change to the rest.
type SheetSettingsParts uint8

// The parts of the SheetSettings a change can alter.
const (
	// SheetSettingsGeneralPart is everything but the parts below, which is what the sheet settings editor changes
	// field by field: the page, display, unit, number format and rule settings.
	SheetSettingsGeneralPart SheetSettingsParts = 1 << iota
	// SheetSettingsLayoutPart is the block layout.
	SheetSettingsLayoutPart
	// SheetSettingsAttributesPart is the attribute definitions.
	SheetSettingsAttributesPart
	// SheetSettingsBodyTypePart is the body type.
	SheetSettingsBodyTypePart
	// SheetSettingsAllParts is every part, which is what replacing the settings wholesale alters.
	SheetSettingsAllParts = SheetSettingsGeneralPart | SheetSettingsLayoutPart | SheetSettingsAttributesPart |
		SheetSettingsBodyTypePart
)

// SheetSettingsResponder defines the method required to be notified of updates to the SheetSettings.
type SheetSettingsResponder interface {
	// SheetSettingsUpdated will be called when the SheetSettings have been updated. The provided Entity will be nil if
	// it was the default SheetSettings that was updated rather than one attached to a specific entity. parts says
	// which of the settings were changed. fullRebuild will be true if the change requires the content to be rebuilt
	// from scratch rather than merely synchronized.
	SheetSettingsUpdated(entity *Entity, parts SheetSettingsParts, fullRebuild bool)
}

// SheetSettingsData holds the SheetSettings data that is written to disk.
type SheetSettingsData struct {
	Page                          *PageSettings      `json:"page,omitzero"`
	Layout                        *SheetLayout       `json:"layout,omitzero"`
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
	HideUnusedWeaponColumns       bool               `json:"hide_unused_weapon_columns,omitzero"`
	ShowSpellAdj                  bool               `json:"show_spell_adj,omitzero"`
	HideSourceMismatch            bool               `json:"hide_source_mismatch,omitzero"`
	HideTLColumn                  bool               `json:"hide_tl_column,omitzero"`
	HideLCColumn                  bool               `json:"hide_lc_column,omitzero"`
	HidePageRefColumn             bool               `json:"hide_page_ref_column,omitzero"`
	UseTitleInFooter              bool               `json:"use_title_in_footer,omitzero"`
	ExcludeUnspentPointsFromTotal bool               `json:"exclude_unspent_points_from_total,omitzero"`
	ShowLiftingSTDamage           bool               `json:"show_lifting_st_damage,omitzero"`
	ShowIQBasedDamage             bool               `json:"show_iq_based_damage,omitzero"`
	HideZeroValueConditionalMods  bool               `json:"hide_zero_value_conditional_modifiers,omitzero"`
	EnforceTraitPrereqs           bool               `json:"enforce_trait_prereqs,omitzero"`
	HeightFormat                  fxp.NumberFormat   `json:"height_format,omitzero"`
	BodyWeightFormat              fxp.NumberFormat   `json:"body_weight_format,omitzero"`
	EquipmentWeightFormat         fxp.NumberFormat   `json:"equipment_weight_format,omitzero"`
	EquipmentValueFormat          fxp.NumberFormat   `json:"equipment_value_format,omitzero"`
	BuyPercent                    int                `json:"buy_percent,omitzero"`
	SellPercent                   int                `json:"sell_percent,omitzero"`
	// SellMoney holds the equipment selling pays out in: copies, each tagged as money (see MoneyTag), that are never
	// containers and have no data owner or parent. Clones of the settings share the copies, which are never edited,
	// though writing them out and working out their values write back what they already hold, so, like the settings
	// themselves, they are for the UI thread only. Entries are only added (see AddSellMoney) or removed (see
	// RemoveSellMoney).
	SellMoney SellMoneyList `json:"sell_money,omitempty"`
}

// SellMoneyList holds the sell money of a SheetSettings (see SheetSettingsData.SellMoney). It marshals without the
// "calc" objects, which are derived values that mean nothing for equipment held by no sheet.
type SellMoneyList []*Equipment

// MarshalJSONTo implements json.MarshalerTo.
func (l SellMoneyList) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, []*Equipment(l), json.WithMarshalers(omitCalcMarker))
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom. An entry that can't be read is logged and left out rather than
// failing the whole document, which would set a settings file aside, or keep a sheet from opening, over one coin.
func (l *SellMoneyList) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	*l = nil
	if dec.PeekKind() != '[' {
		value, err := dec.ReadValue()
		if err != nil {
			return err
		}
		if value.Kind() != 'n' {
			errs.Log(errs.New("ignoring the sell money, which isn't a list"))
		}
		return nil
	}
	if _, err := dec.ReadToken(); err != nil {
		return err
	}
	for dec.PeekKind() != ']' {
		value, err := dec.ReadValue()
		if err != nil {
			return err
		}
		var eqp *Equipment
		if err = jio.Unmarshal(value, &eqp); err != nil {
			errs.Log(errs.NewWithCause("ignoring a sell money entry that can't be read", err))
			continue
		}
		*l = append(*l, eqp)
	}
	_, err := dec.ReadToken()
	return err
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
		Page:                    NewPageSettings(),
		Layout:                  FactorySheetLayout(),
		Attributes:              FactoryAttributeDefs(),
		BodyType:                FactoryBody(),
		DamageProgression:       progression.BasicSet,
		DefaultLengthUnits:      fxp.FeetAndInches,
		DefaultWeightUnits:      fxp.Pound,
		UserDescriptionDisplay:  display.Tooltip,
		ModifiersDisplay:        display.Inline,
		NotesDisplay:            display.Inline,
		SkillLevelAdjDisplay:    display.Tooltip,
		ShowSpellAdj:            true,
		HideUnusedWeaponColumns: false,
		BuyPercent:              100,
		SellPercent:             50,
	}
}

// NewSheetSettingsFromFile loads new settings from a file.
func NewSheetSettingsFromFile(fileSystem fs.FS, filePath string) (*SheetSettings, error) {
	// The raw document is captured first because SheetSettings has a custom unmarshaler. Embedding SheetSettings in a
	// wrapper struct would promote that unmarshaler onto the wrapper, causing it to consume the entire document and
	// leaving the legacy "sheet_settings" field permanently unpopulated.
	var raw jsontext.Value
	if err := jio.LoadFromFile(fileSystem, filePath, &raw); err != nil {
		return nil, err
	}
	var data struct {
		OldLocation *SheetSettings `json:"sheet_settings"`
	}
	if err := jio.Unmarshal(raw, &data); err != nil {
		return nil, errs.Wrap(err)
	}
	s := data.OldLocation
	if s == nil {
		s = &SheetSettings{}
		if err := jio.Unmarshal(raw, s); err != nil {
			return nil, errs.Wrap(err)
		}
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
	if s.Layout == nil {
		s.Layout = FactorySheetLayout()
	} else {
		s.Layout.EnsureValidity()
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
	s.HeightFormat = s.HeightFormat.EnsureValid()
	s.BodyWeightFormat = s.BodyWeightFormat.EnsureValid()
	s.EquipmentWeightFormat = s.EquipmentWeightFormat.EnsureValid()
	s.EquipmentValueFormat = s.EquipmentValueFormat.EnsureValid()
	if s.BuyPercent <= 0 {
		s.BuyPercent = 100
	}
	if s.SellPercent <= 0 {
		s.SellPercent = 50
	}
	s.BuyPercent = min(max(s.BuyPercent, MinTradePercent), MaxTradePercent)
	s.SellPercent = min(max(s.SellPercent, MinTradePercent), MaxTradePercent)
	// Nothing here may reach GlobalSettings(): Settings.Save runs this while holding closedLock, and loading the global
	// settings runs it inside GlobalSettings' own initializer. Checking the kind, the tags and the IDs is safe; cloning
	// an item or working out its value is not.
	seen := make(map[tid.TID]bool, len(s.SellMoney))
	s.SellMoney = slices.DeleteFunc(s.SellMoney, func(one *Equipment) bool {
		if one == nil || one.Container() || !HasTag(MoneyTag, one.Tags) || seen[one.TID] {
			return true
		}
		seen[one.TID] = true
		return false
	})
}

// MinTradePercent and MaxTradePercent bound the percentage of its full price that equipment is bought or sold at.
var (
	MinTradePercent = 1
	MaxTradePercent = 1000
)

// CanAddSellMoney returns true if a copy of the equipment can be added to the sell money list (see AddSellMoney): it
// isn't a container, it is tagged as money (see MoneyTag), since buying spends only money so tagged, and no entry
// already listed MergeMatches the copy.
func (s *SheetSettings) CanAddSellMoney(eqp *Equipment) bool {
	return s.newSellMoney(eqp, LibraryFile{}) != nil
}

// newSellMoney returns the copy of the equipment that AddSellMoney would add, or nil if it can't be added (see
// CanAddSellMoney). The copy is made as a drop into a list makes one (see Reference), with no data owner or parent and
// a quantity of one. Its preconfigured mark is cleared, since carried equipment never has one and MergeMatches compares
// it, so change paid out from a copy that kept it would never merge into what is carried.
func (s *SheetSettings) newSellMoney(eqp *Equipment, from LibraryFile) *Equipment {
	if eqp == nil || eqp.Container() || !HasTag(MoneyTag, eqp.Tags) {
		return nil
	}
	money := eqp.Clone(from, nil, nil, Reference)
	forgetClosedStates(money)
	money.Quantity = fxp.One
	ClearPreconfigured(money)
	if slices.ContainsFunc(s.SellMoney, money.MergeMatches) {
		return nil
	}
	return money
}

// forgetClosedStates removes the closed states recorded for the equipment and its modifiers, which a clone records for
// its copy (see Equipment.Clone). Sell money, and the copies of it that are paid out, is never shown with notes or
// modifiers to collapse, so what its copies record would only pile up in the settings.
func forgetClosedStates(eqp *Equipment) {
	Traverse(func(one *Equipment) bool {
		forgetNodeClosedStates(one)
		Traverse(forgetNodeClosedStates[*EquipmentModifier], false, false, one.Modifiers...)
		return false
	}, false, false, eqp)
}

// forgetNodeClosedStates removes the closed states recorded for the node's notes and, for a container, its children.
// It never stops a Traverse.
func forgetNodeClosedStates[T Node[T]](node T) bool {
	id := string(node.ID())
	SetClosedState("N:"+id, false)
	SetClosedState("n:"+id, false)
	return false
}

// AddSellMoney adds a copy of the equipment to the sell money list, if it can be added (see CanAddSellMoney). from is
// the library file the equipment is shown from, if any, which a copy of an item with no Source of its own takes as its
// Source (see Reference). Returns true if it was added.
func (s *SheetSettings) AddSellMoney(eqp *Equipment, from LibraryFile) bool {
	money := s.newSellMoney(eqp, from)
	if money == nil {
		return false
	}
	s.SellMoney = append(s.SellMoney, money)
	return true
}

// RemoveSellMoney removes the entry with the ID from the sell money list. Returns true if it was there.
func (s *SheetSettings) RemoveSellMoney(id tid.TID) bool {
	before := len(s.SellMoney)
	s.SellMoney = slices.DeleteFunc(s.SellMoney, func(one *Equipment) bool { return one != nil && one.TID == id })
	return len(s.SellMoney) != before
}

// FormatHeight formats a height for display on the sheet, using the default length units and the height display
// format. The result is for display only and must not be parsed back into a stored value.
func (s *SheetSettings) FormatHeight(height fxp.Length) string {
	return s.DefaultLengthUnits.FormatWith(height, s.HeightFormat)
}

// FormatBodyWeight formats a character's weight for display on the sheet, using the default weight units and the body
// weight display format. The result is for display only and must not be parsed back into a stored value.
func (s *SheetSettings) FormatBodyWeight(weight fxp.Weight) string {
	return s.DefaultWeightUnits.FormatWith(weight, s.BodyWeightFormat)
}

// FormatEquipmentWeight formats an equipment weight for display on the sheet, using the default weight units and the
// equipment weight display format. The result is for display only and must not be parsed back into a stored value.
func (s *SheetSettings) FormatEquipmentWeight(weight fxp.Weight) string {
	return s.DefaultWeightUnits.FormatWith(weight, s.EquipmentWeightFormat)
}

// FormatEquipmentValue formats an equipment value for display on the sheet, using the equipment value display format.
// The result is for display only and must not be parsed back into a stored value.
func (s *SheetSettings) FormatEquipmentValue(value fxp.Int) string {
	return s.EquipmentValueFormat.Format(value)
}

// MarshalJSONTo implements json.MarshalerTo.
func (s *SheetSettings) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, &s.SheetSettingsData)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (s *SheetSettings) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var content struct {
		SheetSettingsData
		OldBodyType             *Body    `json:"hit_locations"`
		OldBlockLayout          []string `json:"block_layout"`
		OldShowTraitModifierAdj bool     `json:"show_advantage_modifier_adj"`
	}
	if err := json.UnmarshalDecode(dec, &content); err != nil {
		return err
	}
	s.SheetSettingsData = content.SheetSettingsData
	if s.BodyType == nil && content.OldBodyType != nil {
		s.BodyType = content.OldBodyType
	}
	if s.Layout == nil && len(content.OldBlockLayout) != 0 {
		s.Layout = NewSheetLayoutFromLegacyRows(content.OldBlockLayout)
	}
	if !s.ShowTraitModifierAdj && content.OldShowTraitModifierAdj {
		s.ShowTraitModifierAdj = true
	}
	s.EnsureValidity()
	// The sell money copies left are fresh objects, shared with no one yet. Like library rows, they have no data owner,
	// and attaching them to that points their weapons and modifiers back at them. The preconfigured mark and disabled
	// state mean nothing for them, so any a file holds is cleared.
	SetDataOwnerAll(nil, s.SellMoney)
	ClearPreconfigured(s.SellMoney...)
	ClearDisabled(s.SellMoney...)
	return nil
}

// Clone creates a copy of this. The sell money entries are shared rather than cloned, which is safe since they are
// never edited (see SheetSettingsData.SellMoney) and keeps the clone cheap, as one is made for every entity loaded.
func (s *SheetSettings) Clone(entity *Entity) *SheetSettings {
	clone := *s
	clone.Entity = entity
	clone.Page = s.Page.Clone()
	clone.Layout = s.Layout.Clone()
	clone.Attributes = s.Attributes.Clone()
	clone.BodyType = s.BodyType.Clone(entity, nil)
	clone.SellMoney = slices.Clone(s.SellMoney)
	return &clone
}

// globalSheetSettingsSnapshot holds a clone of the global sheet settings for readers on goroutines other than the UI
// thread. Entities are unmarshaled on background goroutines as well as the UI thread (the deep search content cache
// parses sheets and templates in the background), and one without embedded settings clones the global defaults, but the
// settings dockables mutate the live SheetSettings in place on the UI thread, so reading it from another goroutine is a
// data race. The UI thread publishes a clone here instead; see SyncGlobalSheetSettings.
var globalSheetSettingsSnapshot atomic.Pointer[SheetSettings]

// SyncGlobalSheetSettings publishes a clone of the current global sheet settings for use on any goroutine.
// GlobalSettings() calls this when the settings are first loaded; any code that changes the global sheet settings
// afterward -- whether by replacing them or by mutating them in place -- must call it again.
func SyncGlobalSheetSettings() {
	syncGlobalSheetSettings(GlobalSettings().Sheet)
}

func syncGlobalSheetSettings(s *SheetSettings) {
	globalSheetSettingsSnapshot.Store(s.Clone(nil))
}

// globalSheetSettingsClone returns a clone of the global sheet settings for the given entity, taken from the snapshot
// the UI thread publishes rather than from the live settings, so it is safe to call from any goroutine.
func globalSheetSettingsClone(entity *Entity) *SheetSettings {
	GlobalSettings() // Makes sure the snapshot has been initialized.
	return globalSheetSettingsSnapshot.Load().Clone(entity)
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
