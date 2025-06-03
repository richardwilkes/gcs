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
	"cmp"
	"fmt"
	"hash"
	"strings"
	"unsafe"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/skillsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stdmg"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/tid"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/zeebo/xxh3"
)

var _ Node[*Weapon] = &Weapon{}

// Columns that can be used with the weapon method .CellData()
const (
	WeaponHideColumn = iota
	WeaponDescriptionColumn
	WeaponUsageColumn
	WeaponSLColumn
	WeaponParryColumn
	WeaponBlockColumn
	WeaponDamageColumn
	WeaponReachColumn
	WeaponSTColumn
	WeaponAccColumn
	WeaponRangeColumn
	WeaponRoFColumn
	WeaponShotsColumn
	WeaponBulkColumn
	WeaponRecoilColumn
)

// WeaponOwner defines the methods required of a Weapon owner.
type WeaponOwner interface {
	fmt.Stringer
	nameable.Accesser
	DataOwner() DataOwner
	Description() string
	ResolveLocalNotes() string
	FeatureList() Features
	TagList() []string
	RatedStrength() fxp.Int
}

// WeaponData holds the Weapon data that is written to disk.
type WeaponData struct {
	TID        tid.TID         `json:"id"`
	Damage     WeaponDamage    `json:"damage"`
	Strength   WeaponStrength  `json:"strength,omitempty"`
	Usage      string          `json:"usage,omitempty"`
	UsageNotes string          `json:"usage_notes,omitempty"`
	Reach      WeaponReach     `json:"reach,omitempty"`
	Parry      WeaponParry     `json:"parry,omitempty"`
	Block      WeaponBlock     `json:"block,omitempty"`
	Accuracy   WeaponAccuracy  `json:"accuracy,omitempty"`
	Range      WeaponRange     `json:"range,omitempty"`
	RateOfFire WeaponRoF       `json:"rate_of_fire,omitempty"`
	Shots      WeaponShots     `json:"shots,omitempty"`
	Bulk       WeaponBulk      `json:"bulk,omitempty"`
	Recoil     WeaponRecoil    `json:"recoil,omitempty"`
	Defaults   []*SkillDefault `json:"defaults,omitempty"`
	Hide       bool            `json:"hide,omitempty"`
}

// Weapon holds the stats for a weapon.
type Weapon struct {
	WeaponData
	Owner WeaponOwner
}

// ExtractWeaponsOfType filters the input list down to only those weapons of the given type.
func ExtractWeaponsOfType(melee, excludeHidden bool, list []*Weapon) []*Weapon {
	var result []*Weapon
	for _, w := range list {
		if w.IsMelee() == melee && (!excludeHidden || !w.Hide) {
			result = append(result, w)
		}
	}
	return result
}

// SeparateWeapons returns separate lists for melee and ranged weapons found in the input list.
func SeparateWeapons(excludeHidden bool, list []*Weapon) (melee, ranged []*Weapon) {
	for _, w := range list {
		if !excludeHidden || !w.Hide {
			if w.IsMelee() {
				melee = append(melee, w)
			} else {
				ranged = append(ranged, w)
			}
		}
	}
	return melee, ranged
}

// NewWeapon creates a new weapon of the given type.
func NewWeapon(owner WeaponOwner, melee bool) *Weapon {
	var w Weapon
	w.TID = tid.MustNewTID(weaponKind(melee))
	w.Damage = WeaponDamage{
		WeaponDamageData: WeaponDamageData{
			Type:                      "cr",
			StrengthMultiplier:        fxp.One,
			ArmorDivisor:              fxp.One,
			FragmentationArmorDivisor: fxp.One,
		},
	}
	w.Owner = owner
	if melee {
		w.Reach.Min = fxp.One
		w.Reach.Max = fxp.One
		w.Damage.StrengthType = stdmg.Thrust
	} else {
		w.RateOfFire.Mode1.ShotsPerAttack = fxp.One
		w.Damage.Base = dice.New("1d")
	}
	return &w
}

func weaponKind(melee bool) byte {
	if melee {
		return kinds.WeaponMelee
	}
	return kinds.WeaponRanged
}

// IsMelee returns true if this is a melee weapon.
func (w *Weapon) IsMelee() bool {
	return tid.IsKind(w.TID, kinds.WeaponMelee)
}

// IsRanged returns true if this is a ranged weapon.
func (w *Weapon) IsRanged() bool {
	return tid.IsKind(w.TID, kinds.WeaponRanged)
}

// CloneWeapons clones the input list of weapons.
func CloneWeapons(list []*Weapon, preserveIDs bool) []*Weapon {
	if len(list) == 0 {
		return nil
	}
	weapons := make([]*Weapon, len(list))
	for i, w := range list {
		weapons[i] = w.Clone(LibraryFile{}, nil, nil, preserveIDs)
	}
	return weapons
}

// Clone implements Node.
func (w *Weapon) Clone(_ LibraryFile, _ DataOwner, _ *Weapon, preserveID bool) *Weapon {
	other := *w
	if !preserveID {
		other.TID = tid.MustNewTID(w.TID[0])
	}
	other.Damage = *w.Damage.Clone(&other)
	other.Defaults = nil
	if len(w.Defaults) != 0 {
		other.Defaults = make([]*SkillDefault, len(w.Defaults))
		for i, one := range w.Defaults {
			d := *one
			other.Defaults[i] = &d
		}
	}
	return &other
}

// Compare returns an integer indicating the sort order of this weapon compared to the other weapon.
func (w *Weapon) Compare(other *Weapon) int {
	result := txt.NaturalCmp(w.String(), other.String(), true)
	if result == 0 {
		if result = txt.NaturalCmp(w.UsageWithReplacements(), other.UsageWithReplacements(), true); result == 0 {
			if result = txt.NaturalCmp(w.UsageNotesWithReplacements(), other.UsageNotesWithReplacements(), true); result == 0 {
				//nolint:gosec // Just need a tie-breaker
				result = cmp.Compare(uintptr(unsafe.Pointer(w)), uintptr(unsafe.Pointer(other)))
			}
		}
	}
	return result
}

// HashResolved returns a hash value for this weapon's resolved state.
func (w *Weapon) HashResolved() uint64 {
	h := xxh3.New()
	w.Hash(h)
	hashhelper.String(h, w.String())
	hashhelper.Num64(h, w.SkillLevel(nil))
	hashhelper.String(h, w.Damage.ResolvedDamage(nil))
	hashhelper.Bool(h, w.Hide)
	return h.Sum64()
}

// Hash writes this object's contents into the hasher. Note that this only hashes the data that is considered to be
// "source" data, i.e. not expected to be modified by the user after copying from a library.
func (w *WeaponData) Hash(h hash.Hash) {
	if w == nil {
		hashhelper.Num8(h, uint8(255))
		return
	}
	w.Damage.Hash(h)
	w.Strength.Hash(h)
	hashhelper.String(h, w.Usage)
	hashhelper.String(h, w.UsageNotes)
	w.Reach.Hash(h)
	w.Parry.Hash(h)
	w.Block.Hash(h)
	w.Accuracy.Hash(h)
	w.Range.Hash(h)
	w.RateOfFire.Hash(h)
	w.Shots.Hash(h)
	w.Bulk.Hash(h)
	w.Recoil.Hash(h)
	hashhelper.Num64(h, len(w.Defaults))
	for _, one := range w.Defaults {
		one.Hash(h)
	}
}

// MarshalJSON implements json.Marshaler.
func (w *Weapon) MarshalJSON() ([]byte, error) {
	type calc struct {
		Level      fxp.Int `json:"level,omitempty"`
		Damage     string  `json:"damage,omitempty"`
		Parry      string  `json:"parry,omitempty"`
		Block      string  `json:"block,omitempty"`
		Accuracy   string  `json:"accuracy,omitempty"`
		Reach      string  `json:"reach,omitempty"`
		Range      string  `json:"range,omitempty"`
		RateOfFire string  `json:"rate_of_fire,omitempty"`
		Shots      string  `json:"shots,omitempty"`
		Bulk       string  `json:"bulk,omitempty"`
		Recoil     string  `json:"recoil,omitempty"`
		Strength   string  `json:"strength,omitempty"`
	}
	data := struct {
		WeaponData
		Calc *calc `json:"calc,omitempty"`
	}{
		WeaponData: w.WeaponData,
		Calc: &calc{
			Level:  w.SkillLevel(nil).Max(0),
			Damage: w.Damage.ResolvedDamage(nil),
		},
	}
	if data.Calc.Strength = w.Strength.Resolve(w, nil).String(); data.Calc.Strength == w.Strength.String() {
		data.Calc.Strength = ""
	}
	musclePowerIsResolved := w.Entity() != nil
	if w.IsMelee() {
		data.Accuracy = WeaponAccuracy{}
		data.Range = WeaponRange{}
		data.RateOfFire = WeaponRoF{}
		data.Shots = WeaponShots{}
		data.Bulk = WeaponBulk{}
		data.Recoil = WeaponRecoil{}
		if data.Calc.Parry = w.Parry.Resolve(w, nil).String(); data.Calc.Parry == w.Parry.String() {
			data.Calc.Parry = ""
		}
		if data.Calc.Block = w.Block.Resolve(w, nil).String(); data.Calc.Block == w.Block.String() {
			data.Calc.Block = ""
		}
		if data.Calc.Reach = w.Reach.Resolve(w, nil).String(); data.Calc.Reach == w.Reach.String() {
			data.Calc.Reach = ""
		}
	} else {
		data.Parry = WeaponParry{}
		data.Block = WeaponBlock{}
		data.Reach = WeaponReach{}
		if data.Calc.Accuracy = w.Accuracy.Resolve(w, nil).String(); data.Calc.Accuracy == w.Accuracy.String() {
			data.Calc.Accuracy = ""
		}
		if data.Calc.Range = w.Range.Resolve(w, nil).String(musclePowerIsResolved); data.Calc.Range == w.Range.String(false) {
			data.Calc.Range = ""
		}
		if data.Calc.RateOfFire = w.RateOfFire.Resolve(w, nil).String(); data.Calc.RateOfFire == w.RateOfFire.String() {
			data.Calc.RateOfFire = ""
		}
		if data.Calc.Shots = w.Shots.Resolve(w, nil).String(); data.Calc.Shots == w.Shots.String() {
			data.Calc.Shots = ""
		}
		if data.Calc.Bulk = w.Bulk.Resolve(w, nil).String(); data.Calc.Bulk == w.Bulk.String() {
			data.Calc.Bulk = ""
		}
		if data.Calc.Recoil = w.Recoil.Resolve(w, nil).String(); data.Calc.Recoil == w.Recoil.String() {
			data.Calc.Recoil = ""
		}
	}
	if *data.Calc == (calc{}) {
		data.Calc = nil
	}
	return json.Marshal(&data)
}

// UnmarshalJSON implements json.Unmarshaler.
func (w *Weapon) UnmarshalJSON(data []byte) error {
	var localData struct {
		WeaponData
		// Old data fields
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &localData); err != nil {
		return err
	}
	if !tid.IsValid(localData.TID) {
		// Fixup old data that used UUIDs instead of TIDs
		localData.TID = tid.MustNewTID(weaponKind(localData.Type == "melee_weapon"))
	}
	w.WeaponData = localData.WeaponData
	w.Validate()
	return nil
}

// GetSource returns the source of this data.
func (w *Weapon) GetSource() Source {
	return Source{}
}

// ClearSource clears the source of this data.
func (w *Weapon) ClearSource() {
}

// SyncWithSource synchronizes this data with the source.
func (w *Weapon) SyncWithSource() {
}

// ID returns the local ID of this data.
func (w *Weapon) ID() tid.TID {
	return w.TID
}

// Kind returns the kind of data.
func (w *Weapon) Kind() string {
	if w.IsMelee() {
		return i18n.Text("Melee Weapon")
	}
	return i18n.Text("Ranged Weapon")
}

func (w *Weapon) String() string {
	if toolbox.IsNil(w.Owner) {
		return ""
	}
	return w.Owner.Description()
}

// Notes returns the notes for this weapon.
func (w *Weapon) Notes() string {
	var buffer strings.Builder
	if w.Owner != nil {
		buffer.WriteString(w.Owner.ResolveLocalNotes())
		switch owner := w.Owner.(type) {
		case *Equipment:
			Traverse(func(mod *EquipmentModifier) bool {
				if mod.ShowNotesOnWeapon {
					AppendStringOntoNewLine(&buffer, strings.TrimSpace(mod.ResolveLocalNotes()))
				}
				return false
			}, true, true, owner.Modifiers...)
		case *Trait:
			Traverse(func(mod *TraitModifier) bool {
				if mod.ShowNotesOnWeapon {
					AppendStringOntoNewLine(&buffer, strings.TrimSpace(mod.ResolveLocalNotes()))
				}
				return false
			}, true, true, owner.Modifiers...)
		}
	}
	AppendStringOntoNewLine(&buffer, strings.TrimSpace(w.UsageNotesWithReplacements()))
	return buffer.String()
}

// SetOwner sets the owner and ensures sub-components have their owners set.
func (w *Weapon) SetOwner(owner WeaponOwner) {
	w.Owner = owner
	w.Damage.Owner = w
}

// DataOwner returns the weapon owner's data owner.
func (w *Weapon) DataOwner() DataOwner {
	if toolbox.IsNil(w.Owner) {
		return nil
	}
	return w.Owner.DataOwner()
}

// SetDataOwner does nothing.
func (w *Weapon) SetDataOwner(_ DataOwner) {
}

// Entity returns the owning entity, if any.
func (w *Weapon) Entity() *Entity {
	owner := w.DataOwner()
	if owner == nil {
		return nil
	}
	return owner.OwningEntity()
}

// SkillLevel returns the resolved skill level.
func (w *Weapon) SkillLevel(tooltip *xio.ByteBuffer) fxp.Int {
	entity := w.Entity()
	if entity == nil {
		return 0
	}
	var primaryTooltip *xio.ByteBuffer
	if tooltip != nil {
		primaryTooltip = &xio.ByteBuffer{}
	}
	adj := w.skillLevelBaseAdjustment(entity, primaryTooltip) + w.skillLevelPostAdjustment(entity, primaryTooltip)
	best := fxp.Min
	replacements := w.NameableReplacements()
	for _, def := range w.Defaults {
		if level := def.SkillLevelFast(entity, replacements, false, nil, true); level != fxp.Min {
			level += adj
			if best < level {
				best = level
			}
		}
	}
	if best == fxp.Min {
		return 0
	}
	AppendBufferOntoNewLine(tooltip, primaryTooltip)
	if best < 0 {
		best = 0
	}
	return best
}

func (w *Weapon) usesCrossbowSkill() bool {
	replacements := w.NameableReplacements()
	for _, def := range w.Defaults {
		if def.NameWithReplacements(replacements) == "Crossbow" {
			return true
		}
	}
	return false
}

func (w *Weapon) skillLevelBaseAdjustment(e *Entity, tooltip *xio.ByteBuffer) fxp.Int {
	var adj fxp.Int
	minST := w.Strength.Resolve(w, nil).Min
	if !w.IsRanged() || (w.Range.MusclePowered && !w.usesCrossbowSkill()) {
		minST -= e.StrikingStrength()
	} else {
		minST -= e.LiftingStrength()
	}
	if minST > 0 {
		adj -= minST
		if tooltip != nil {
			tooltip.WriteByte('\n')
			tooltip.WriteString(w.String())
			tooltip.WriteString(" [")
			tooltip.WriteString((-minST).String())
			tooltip.WriteString(i18n.Text(" to skill level due to minimum ST requirement"))
			tooltip.WriteByte(']')
		}
	}
	nameQualifier := w.String()
	for _, bonus := range e.NamedWeaponSkillBonusesFor(nameQualifier, w.UsageWithReplacements(), w.Owner.TagList(), tooltip) {
		adj += bonus.AdjustedAmount()
	}
	for _, f := range w.Owner.FeatureList() {
		adj += w.extractSkillBonusForThisWeapon(f, tooltip)
	}
	if t, ok := w.Owner.(*Trait); ok {
		Traverse(func(mod *TraitModifier) bool {
			for _, f := range mod.Features {
				adj += w.extractSkillBonusForThisWeapon(f, tooltip)
			}
			return false
		}, true, true, t.Modifiers...)
	}
	if eqp, ok := w.Owner.(*Equipment); ok {
		Traverse(func(mod *EquipmentModifier) bool {
			for _, f := range mod.Features {
				adj += w.extractSkillBonusForThisWeapon(f, tooltip)
			}
			return false
		}, true, true, eqp.Modifiers...)
	}
	return adj
}

func (w *Weapon) skillLevelPostAdjustment(e *Entity, tooltip *xio.ByteBuffer) fxp.Int {
	if w.IsMelee() &&
		// Cannot use w.ParryParts.Resolve() here, because that calls this
		w.ResolveBoolFlag(wswitch.CanParry, w.Parry.CanParry) &&
		w.ResolveBoolFlag(wswitch.Fencing, w.Parry.Fencing) {
		return w.EncumbrancePenalty(e, tooltip)
	}
	return 0
}

// EncumbrancePenalty returns the current encumbrance penalty.
func (w *Weapon) EncumbrancePenalty(e *Entity, tooltip *xio.ByteBuffer) fxp.Int {
	if e == nil {
		return 0
	}
	penalty := e.EncumbranceLevel(true).Penalty()
	if penalty != 0 && tooltip != nil {
		tooltip.WriteByte('\n')
		tooltip.WriteString(i18n.Text("Encumbrance"))
		tooltip.WriteString(" [")
		tooltip.WriteString(penalty.StringWithSign())
		tooltip.WriteByte(']')
	}
	return penalty
}

func (w *Weapon) extractSkillBonusForThisWeapon(f Feature, tooltip *xio.ByteBuffer) fxp.Int {
	if sb, ok := f.(*SkillBonus); ok {
		if sb.SelectionType.EnsureValid() == skillsel.ThisWeapon {
			if sb.SpecializationCriteria.Matches(w.NameableReplacements(), w.UsageWithReplacements()) {
				sb.AddToTooltip(tooltip)
				return sb.AdjustedAmount()
			}
		}
	}
	return 0
}

// ResolveBoolFlag returns the resolved value of the given bool flag.
func (w *Weapon) ResolveBoolFlag(switchType wswitch.Type, initial bool) bool {
	entity := w.Entity()
	if entity == nil {
		return initial
	}
	t := 0
	f := 0
	for _, bonus := range w.collectWeaponBonuses(1, nil, feature.WeaponSwitch) {
		if bonus.SwitchType == switchType {
			if bonus.SwitchTypeValue {
				t++
			} else {
				f++
			}
		}
	}
	if t > f {
		return true
	}
	if f > t {
		return false
	}
	return initial
}

func (w *Weapon) collectWeaponBonuses(dieCount int, tooltip *xio.ByteBuffer, allowedFeatureTypes ...feature.Type) []*WeaponBonus {
	entity := w.Entity()
	if entity == nil {
		return nil
	}
	allowed := make(map[feature.Type]bool, len(allowedFeatureTypes))
	for _, one := range allowedFeatureTypes {
		allowed[one] = true
	}
	var bestDef *SkillDefault
	best := fxp.Min
	replacements := w.NameableReplacements()
	for _, one := range w.Defaults {
		if one.SkillBased() {
			if level := one.SkillLevelFast(entity, replacements, false, nil, true); best < level {
				best = level
				bestDef = one
			}
		}
	}
	bonusSet := make(map[*WeaponBonus]bool)
	tags := w.Owner.TagList()
	var name, specialization string
	if bestDef != nil {
		name = bestDef.NameWithReplacements(replacements)
		specialization = bestDef.SpecializationWithReplacements(replacements)
	}
	entity.AddWeaponWithSkillBonusesFor(name, specialization, w.UsageWithReplacements(), tags, dieCount, tooltip,
		bonusSet, allowed)
	nameQualifier := w.String()
	entity.AddNamedWeaponBonusesFor(nameQualifier, w.UsageWithReplacements(), tags, dieCount, tooltip, bonusSet,
		allowed)
	for _, f := range w.Owner.FeatureList() {
		w.extractWeaponBonus(f, bonusSet, allowed, fxp.From(dieCount), tooltip)
	}
	if t, ok := w.Owner.(*Trait); ok {
		Traverse(func(mod *TraitModifier) bool {
			var bonus Bonus
			for _, f := range mod.Features {
				if bonus, ok = f.(Bonus); ok {
					bonus.SetSubOwner(mod)
				}
				w.extractWeaponBonus(f, bonusSet, allowed, fxp.From(dieCount), tooltip)
			}
			return false
		}, true, true, t.Modifiers...)
	}
	if eqp, ok := w.Owner.(*Equipment); ok {
		Traverse(func(mod *EquipmentModifier) bool {
			var bonus Bonus
			for _, f := range mod.Features {
				if bonus, ok = f.(Bonus); ok {
					bonus.SetSubOwner(mod)
				}
				w.extractWeaponBonus(f, bonusSet, allowed, fxp.From(dieCount), tooltip)
			}
			return false
		}, true, true, eqp.Modifiers...)
	}
	if len(bonusSet) == 0 {
		return nil
	}
	result := make([]*WeaponBonus, 0, len(bonusSet))
	for bonus := range bonusSet {
		result = append(result, bonus)
	}
	return result
}

func (w *Weapon) extractWeaponBonus(f Feature, set map[*WeaponBonus]bool, allowedFeatureTypes map[feature.Type]bool, dieCount fxp.Int, tooltip *xio.ByteBuffer) {
	if allowedFeatureTypes[f.FeatureType()] {
		if bonus, ok := f.(*WeaponBonus); ok {
			savedLevel := bonus.Level
			savedDieCount := bonus.DieCount
			bonus.Level = bonus.DerivedLevel()
			bonus.DieCount = dieCount
			replacements := w.NameableReplacements()
			switch bonus.SelectionType {
			case wsel.WithRequiredSkill:
			case wsel.ThisWeapon:
				if bonus.SpecializationCriteria.Matches(replacements, w.UsageWithReplacements()) {
					if _, exists := set[bonus]; !exists {
						set[bonus] = true
						bonus.AddToTooltip(tooltip)
					}
				}
			case wsel.WithName:
				if bonus.NameCriteria.Matches(replacements, w.String()) &&
					bonus.SpecializationCriteria.Matches(replacements, w.UsageWithReplacements()) &&
					bonus.TagsCriteria.MatchesList(replacements, w.Owner.TagList()...) {
					if _, exists := set[bonus]; !exists {
						set[bonus] = true
						bonus.AddToTooltip(tooltip)
					}
				}
			default:
				errs.Log(errs.New("unknown selection type"), "type", int(bonus.SelectionType))
			}
			bonus.Level = savedLevel
			bonus.DieCount = savedDieCount
		}
	}
}

// UsageWithReplacements returns the usage of the weapon with any nameable keys replaced.
func (w *Weapon) UsageWithReplacements() string {
	return nameable.Apply(w.Usage, w.NameableReplacements())
}

// UsageNotesWithReplacements returns the usage notes of the weapon with any nameable keys replaced.
func (w *Weapon) UsageNotesWithReplacements() string {
	return nameable.Apply(w.UsageNotes, w.NameableReplacements())
}

// NameableReplacements returns the replacements to be used with this weapon.
func (w *Weapon) NameableReplacements() map[string]string {
	if toolbox.IsNil(w.Owner) {
		return nil
	}
	return w.Owner.NameableReplacements()
}

// FillWithNameableKeys adds any nameable keys found in this Weapon to the provided map.
func (w *Weapon) FillWithNameableKeys(m, existing map[string]string) {
	nameable.Extract(w.Usage, m, existing)
	nameable.Extract(w.UsageNotes, m, existing)
	for _, one := range w.Defaults {
		one.FillWithNameableKeys(m, existing)
	}
}

// ApplyNameableKeys replaces any nameable keys found in this Weapon with the corresponding values in the provided map.
func (w *Weapon) ApplyNameableKeys(_ map[string]string) {
}

// Container returns true if this is a container.
func (w *Weapon) Container() bool {
	return false
}

// IsOpen returns true if this node is currently open.
func (w *Weapon) IsOpen() bool {
	return false
}

// SetOpen sets the current open state for this node.
func (w *Weapon) SetOpen(_ bool) {
}

// Enabled returns true if this node is enabled.
func (w *Weapon) Enabled() bool {
	return true
}

// Parent returns the parent.
func (w *Weapon) Parent() *Weapon {
	return nil
}

// SetParent sets the parent.
func (w *Weapon) SetParent(_ *Weapon) {
}

// HasChildren returns true if this node has children.
func (w *Weapon) HasChildren() bool {
	return false
}

// NodeChildren returns the children of this node, if any.
func (w *Weapon) NodeChildren() []*Weapon {
	return nil
}

// SetChildren sets the children of this node.
func (w *Weapon) SetChildren(_ []*Weapon) {
}

// WeaponHeaderData returns the header data information for the given weapon column.
func WeaponHeaderData(columnID int, melee, forPage bool) HeaderData {
	var data HeaderData
	switch columnID {
	case WeaponHideColumn:
		data.Title = i18n.Text("Hide")
	case WeaponDescriptionColumn:
		if melee {
			data.Title = i18n.Text("Melee Weapon")
		} else {
			data.Title = i18n.Text("Ranged Weapon")
		}
		data.Primary = true
	case WeaponUsageColumn:
		switch {
		case forPage:
			data.Title = i18n.Text("Usage")
		case melee:
			data.Title = i18n.Text("Melee Weapon Usage")
		default:
			data.Title = i18n.Text("Ranged Weapon Usage")
		}
	case WeaponSLColumn:
		data.Title = i18n.Text("SL")
		data.Detail = i18n.Text("Skill Level")
	case WeaponParryColumn:
		data.Title = i18n.Text("Parry")
	case WeaponBlockColumn:
		data.Title = i18n.Text("Block")
	case WeaponDamageColumn:
		data.Title = i18n.Text("Damage")
	case WeaponReachColumn:
		data.Title = i18n.Text("Reach")
	case WeaponSTColumn:
		data.Title = i18n.Text("ST")
		data.Detail = i18n.Text("Minimum Strength")
	case WeaponAccColumn:
		data.Title = i18n.Text("Acc")
		data.Detail = i18n.Text("Accuracy Bonus")
	case WeaponRangeColumn:
		data.Title = i18n.Text("Range")
	case WeaponRoFColumn:
		data.Title = i18n.Text("RoF")
		data.Detail = i18n.Text("Rate of Fire")
	case WeaponShotsColumn:
		data.Title = i18n.Text("Shots")
	case WeaponBulkColumn:
		data.Title = i18n.Text("Bulk")
	case WeaponRecoilColumn:
		data.Title = i18n.Text("Recoil")
	}
	return data
}

// CellData returns the cell data information for the given column.
func (w *Weapon) CellData(columnID int, data *CellData) {
	var buffer xio.ByteBuffer
	data.Type = cell.Text
	switch columnID {
	case WeaponHideColumn:
		data.Type = cell.Toggle
		data.Checked = w.Hide
		data.Alignment = align.Middle
	case WeaponDescriptionColumn:
		data.Primary = w.String()
		data.Secondary = w.Notes()
	case WeaponUsageColumn:
		data.Primary = w.UsageWithReplacements()
	case WeaponSLColumn:
		data.Primary = w.SkillLevel(&buffer).String()
	case WeaponParryColumn:
		parry := w.Parry.Resolve(w, &buffer)
		data.Primary = parry.String()
		data.Tooltip = parry.Tooltip()
	case WeaponBlockColumn:
		data.Primary = w.Block.Resolve(w, &buffer).String()
	case WeaponDamageColumn:
		data.Primary = w.Damage.ResolvedDamage(&buffer)
	case WeaponReachColumn:
		reach := w.Reach.Resolve(w, &buffer)
		data.Primary = reach.String()
		data.Tooltip = reach.Tooltip()
	case WeaponSTColumn:
		weaponST := w.Strength.Resolve(w, &buffer)
		data.Primary = weaponST.String()
		data.Tooltip = weaponST.Tooltip(w)
	case WeaponAccColumn:
		data.Primary = w.Accuracy.Resolve(w, &buffer).String()
	case WeaponRangeColumn:
		data.Primary = w.Range.Resolve(w, &buffer).String(w.Entity() != nil)
	case WeaponRoFColumn:
		rof := w.RateOfFire.Resolve(w, &buffer)
		data.Primary = rof.String()
		data.Tooltip = rof.Tooltip()
	case WeaponShotsColumn:
		shots := w.Shots.Resolve(w, &buffer)
		data.Primary = shots.String()
		data.Tooltip = shots.Tooltip()
	case WeaponBulkColumn:
		bulk := w.Bulk.Resolve(w, &buffer)
		data.Primary = bulk.String()
		data.Tooltip = bulk.Tooltip(w)
	case WeaponRecoilColumn:
		recoil := w.Recoil.Resolve(w, &buffer)
		data.Primary = recoil.String()
		data.Tooltip = recoil.Tooltip()
	case PageRefCellAlias:
		data.Type = cell.PageRef
	}
	if buffer.Len() > 0 {
		if data.Tooltip != "" {
			data.Tooltip += "\n\n"
		}
		data.Tooltip += i18n.Text("Includes modifiers from:") + buffer.String()
	}
}

// CopyFrom implements node.EditorData.
func (w *Weapon) CopyFrom(t *Weapon) {
	*w = *t.Clone(LibraryFile{}, t.DataOwner(), nil, false)
}

// ApplyTo implements node.EditorData.
func (w *Weapon) ApplyTo(t *Weapon) {
	*t = *w.Clone(LibraryFile{}, t.DataOwner(), nil, true)
}

// Validate ensures the weapon data is valid.
func (w *Weapon) Validate() {
	w.Strength.Validate()
	if w.IsMelee() {
		w.Parry.Validate()
		w.Block.Validate()
		w.Reach.Validate()
		w.Accuracy = WeaponAccuracy{}
		w.Range = WeaponRange{}
		w.RateOfFire = WeaponRoF{}
		w.Shots = WeaponShots{}
		w.Bulk = WeaponBulk{}
		w.Recoil = WeaponRecoil{}
	} else {
		if w.Accuracy.Jet || w.RateOfFire.Jet {
			w.Accuracy.Jet = true
			w.RateOfFire.Jet = true
		}
		w.Accuracy.Validate()
		w.Range.Validate()
		w.RateOfFire.Validate()
		w.Shots.Validate()
		w.Bulk.Validate()
		w.Recoil.Validate()
		w.Parry = WeaponParry{}
		w.Block = WeaponBlock{}
		w.Reach = WeaponReach{}
	}
}

// WeaponSVG returns the SVG that should be used for the weapon type.
func WeaponSVG(melee bool) *unison.SVG {
	if melee {
		return svg.MeleeWeapon
	}
	return svg.RangedWeapon
}
