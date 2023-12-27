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
	"cmp"
	"context"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"strings"
	"unsafe"

	"github.com/google/uuid"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xio"
)

var (
	_ Node[*Weapon] = &Weapon{}
	// WeaponCtxKey is the context key used to store the weapon in the context.
	WeaponCtxKey = weaponCtxKey(1)
)

// Columns that can be used with the weapon method .CellData()
const (
	WeaponDescriptionColumn = iota
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

type weaponCtxKey int

// WeaponOwner defines the methods required of a Weapon owner.
type WeaponOwner interface {
	fmt.Stringer
	OwningEntity() *Entity
	Description() string
	Notes() string
	FeatureList() Features
	TagList() []string
	RatedStrength() fxp.Int
}

// WeaponData holds the Weapon data that is written to disk.
type WeaponData struct {
	ID         uuid.UUID       `json:"id"`
	Type       WeaponType      `json:"type"`
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
}

// Weapon holds the stats for a weapon.
type Weapon struct {
	WeaponData
	Owner WeaponOwner
}

// ExtractWeaponsOfType filters the input list down to only those weapons of the given type.
func ExtractWeaponsOfType(desiredType WeaponType, list []*Weapon) []*Weapon {
	var result []*Weapon
	for _, w := range list {
		if w.Type == desiredType {
			result = append(result, w)
		}
	}
	return result
}

// SeparateWeapons returns separate lists for melee and ranged weapons found in the input list.
func SeparateWeapons(list []*Weapon) (melee, ranged []*Weapon) {
	for _, w := range list {
		switch w.Type {
		case MeleeWeaponType:
			melee = append(melee, w)
		case RangedWeaponType:
			ranged = append(ranged, w)
		default:
		}
	}
	return melee, ranged
}

// NewWeapon creates a new weapon of the given type.
func NewWeapon(owner WeaponOwner, weaponType WeaponType) *Weapon {
	w := &Weapon{
		WeaponData: WeaponData{
			ID:   NewUUID(),
			Type: weaponType,
			Damage: WeaponDamage{
				WeaponDamageData: WeaponDamageData{
					Type:                      "cr",
					ArmorDivisor:              fxp.One,
					FragmentationArmorDivisor: fxp.One,
				},
			},
		},
		Owner: owner,
	}
	switch weaponType {
	case MeleeWeaponType:
		w.Reach.Min = fxp.One
		w.Reach.Max = fxp.One
		w.Damage.StrengthType = ThrustStrengthDamage
	case RangedWeaponType:
		w.RateOfFire.Mode1.ShotsPerAttack = fxp.One
		w.Damage.Base = dice.New("1d")
	default:
	}
	return w
}

// Clone implements Node.
func (w *Weapon) Clone(_ *Entity, _ *Weapon, preserveID bool) *Weapon {
	other := *w
	if !preserveID {
		other.ID = uuid.New()
	}
	other.Damage = *other.Damage.Clone(&other)
	if other.Defaults != nil {
		other.Defaults = make([]*SkillDefault, 0, len(w.Defaults))
		for _, one := range w.Defaults {
			d := *one
			other.Defaults = append(other.Defaults, &d)
		}
	}
	return &other
}

// Compare returns an integer indicating the sort order of this weapon compared to the other weapon.
func (w *Weapon) Compare(other *Weapon) int {
	result := txt.NaturalCmp(w.String(), other.String(), true)
	if result == 0 {
		if result = txt.NaturalCmp(w.Usage, other.Usage, true); result == 0 {
			if result = txt.NaturalCmp(w.UsageNotes, other.UsageNotes, true); result == 0 {
				result = cmp.Compare(uintptr(unsafe.Pointer(w)), uintptr(unsafe.Pointer(other))) //nolint:gosec // Just need a tie-breaker
			}
		}
	}
	return result
}

// HashCode returns a hash value for this weapon's resolved state.
// nolint:errcheck // Not checking errors on writes to a bytes.Buffer
func (w *Weapon) HashCode() uint32 {
	h := fnv.New32()
	_, _ = h.Write(w.ID[:])
	_, _ = h.Write([]byte{byte(w.Type)})
	_, _ = h.Write([]byte(w.String()))
	_, _ = h.Write([]byte(w.UsageNotes))
	_, _ = h.Write([]byte(w.Usage))
	_ = binary.Write(h, binary.LittleEndian, w.SkillLevel(nil))
	_, _ = h.Write([]byte(w.Damage.ResolvedDamage(nil)))
	w.Parry.hash(h)
	w.Block.hash(h)
	w.Accuracy.hash(h)
	w.Reach.hash(h)
	w.Range.hash(h)
	w.RateOfFire.hash(h)
	w.Shots.hash(h)
	w.Bulk.hash(h)
	w.Recoil.hash(h)
	w.Strength.hash(h)
	return h.Sum32()
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
	musclePowerIsResolved := w.PC() != nil
	switch w.Type {
	case MeleeWeaponType:
		if data.Calc.Parry = w.Parry.Resolve(w, nil).String(); data.Calc.Parry == w.Parry.String() {
			data.Calc.Parry = ""
		}
		if data.Calc.Block = w.Block.Resolve(w, nil).String(); data.Calc.Block == w.Block.String() {
			data.Calc.Block = ""
		}
		if data.Calc.Reach = w.Reach.Resolve(w, nil).String(); data.Calc.Reach == w.Reach.String() {
			data.Calc.Reach = ""
		}
	case RangedWeaponType:
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
	default:
	}
	if *data.Calc == (calc{}) {
		data.Calc = nil
	}
	return json.MarshalWithContext(context.WithValue(context.Background(), WeaponCtxKey, w), &data)
}

// UnmarshalJSON implements json.Unmarshaler.
func (w *Weapon) UnmarshalJSON(data []byte) error {
	w.WeaponData = WeaponData{}
	if err := json.Unmarshal(data, &w.WeaponData); err != nil {
		return err
	}
	w.Validate()
	return nil
}

// UUID returns the UUID of this data.
func (w *Weapon) UUID() uuid.UUID {
	return w.ID
}

// Kind returns the kind of data.
func (w *Weapon) Kind() string {
	return w.Type.String()
}

func (w *Weapon) String() string {
	if w.Owner == nil {
		return ""
	}
	return w.Owner.Description()
}

// Notes returns the notes for this weapon.
func (w *Weapon) Notes() string {
	var buffer strings.Builder
	if w.Owner != nil {
		buffer.WriteString(w.Owner.Notes())
	}
	AppendStringOntoNewLine(&buffer, strings.TrimSpace(w.UsageNotes))
	return buffer.String()
}

// SetOwner sets the owner and ensures sub-components have their owners set.
func (w *Weapon) SetOwner(owner WeaponOwner) {
	w.Owner = owner
	w.Damage.Owner = w
}

// Entity returns the owning entity, if any.
func (w *Weapon) Entity() *Entity {
	if w.Owner == nil {
		return nil
	}
	entity := w.Owner.OwningEntity()
	if entity == nil {
		return nil
	}
	return entity
}

// PC returns the owning PC, if any.
func (w *Weapon) PC() *Entity {
	if entity := w.Entity(); entity != nil && entity.Type == PC {
		return entity
	}
	return nil
}

// SkillLevel returns the resolved skill level.
func (w *Weapon) SkillLevel(tooltip *xio.ByteBuffer) fxp.Int {
	pc := w.PC()
	if pc == nil {
		return 0
	}
	var primaryTooltip *xio.ByteBuffer
	if tooltip != nil {
		primaryTooltip = &xio.ByteBuffer{}
	}
	adj := w.skillLevelBaseAdjustment(pc, primaryTooltip) + w.skillLevelPostAdjustment(pc, primaryTooltip)
	best := fxp.Min
	for _, def := range w.Defaults {
		if level := def.SkillLevelFast(pc, false, nil, true); level != fxp.Min {
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

func (w *Weapon) skillLevelBaseAdjustment(entity *Entity, tooltip *xio.ByteBuffer) fxp.Int {
	var adj fxp.Int
	if minST := w.Strength.Resolve(w, nil).Minimum - entity.StrikingStrength(); minST > 0 {
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
	for _, bonus := range entity.NamedWeaponSkillBonusesFor(nameQualifier, w.Usage, w.Owner.TagList(), tooltip) {
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

func (w *Weapon) skillLevelPostAdjustment(entity *Entity, tooltip *xio.ByteBuffer) fxp.Int {
	if w.Type.EnsureValid() == MeleeWeaponType &&
		// Cannot use w.ParryParts.Resolve() here, because that calls this
		w.ResolveBoolFlag(CanParryWeaponSwitchType, !w.Parry.No) &&
		w.ResolveBoolFlag(FencingWeaponSwitchType, w.Parry.Fencing) {
		return w.EncumbrancePenalty(entity, tooltip)
	}
	return 0
}

// EncumbrancePenalty returns the current encumbrance penalty.
func (w *Weapon) EncumbrancePenalty(entity *Entity, tooltip *xio.ByteBuffer) fxp.Int {
	if entity == nil {
		return 0
	}
	penalty := entity.EncumbranceLevel(true).Penalty()
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
		if sb.SelectionType.EnsureValid() == ThisWeaponSkillSelectionType {
			if sb.SpecializationCriteria.Matches(w.Usage) {
				sb.AddToTooltip(tooltip)
				return sb.AdjustedAmount()
			}
		}
	}
	return 0
}

// ResolveBoolFlag returns the resolved value of the given bool flag.
func (w *Weapon) ResolveBoolFlag(switchType WeaponSwitchType, initial bool) bool {
	pc := w.PC()
	if pc == nil {
		return initial
	}
	t := 0
	f := 0
	for _, bonus := range w.collectWeaponBonuses(1, nil, WeaponSwitchFeatureType) {
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

func (w *Weapon) collectWeaponBonuses(dieCount int, tooltip *xio.ByteBuffer, allowedFeatureTypes ...FeatureType) []*WeaponBonus {
	pc := w.PC()
	if pc == nil {
		return nil
	}
	allowed := make(map[FeatureType]bool, len(allowedFeatureTypes))
	for _, one := range allowedFeatureTypes {
		allowed[one] = true
	}
	var bestDef *SkillDefault
	best := fxp.Min
	for _, one := range w.Defaults {
		if one.SkillBased() {
			if level := one.SkillLevelFast(pc, false, nil, true); best < level {
				best = level
				bestDef = one
			}
		}
	}
	bonusSet := make(map[*WeaponBonus]bool)
	tags := w.Owner.TagList()
	if bestDef != nil {
		pc.AddWeaponWithSkillBonusesFor(bestDef.Name, bestDef.Specialization, tags, dieCount, tooltip, bonusSet, allowed)
	}
	nameQualifier := w.String()
	pc.AddNamedWeaponBonusesFor(nameQualifier, w.Usage, tags, dieCount, tooltip, bonusSet, allowed)
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

func (w *Weapon) extractWeaponBonus(f Feature, set map[*WeaponBonus]bool, allowedFeatureTypes map[FeatureType]bool, dieCount fxp.Int, tooltip *xio.ByteBuffer) {
	if allowedFeatureTypes[f.FeatureType()] {
		if bonus, ok := f.(*WeaponBonus); ok {
			level := bonus.LeveledAmount.Level
			if bonus.Type == WeaponBonusFeatureType {
				bonus.LeveledAmount.Level = dieCount
			} else {
				bonus.LeveledAmount.Level = bonus.DerivedLevel()
			}
			switch bonus.SelectionType {
			case WithRequiredSkillWeaponSelectionType:
			case ThisWeaponWeaponSelectionType:
				if bonus.SpecializationCriteria.Matches(w.Usage) {
					if _, exists := set[bonus]; !exists {
						set[bonus] = true
						bonus.AddToTooltip(tooltip)
					}
				}
			case WithNameWeaponSelectionType:
				if bonus.NameCriteria.Matches(w.String()) && bonus.SpecializationCriteria.Matches(w.Usage) &&
					bonus.TagsCriteria.MatchesList(w.Owner.TagList()...) {
					if _, exists := set[bonus]; !exists {
						set[bonus] = true
						bonus.AddToTooltip(tooltip)
					}
				}
			default:
				errs.Log(errs.New("unknown selection type"), "type", int(bonus.SelectionType))
			}
			bonus.LeveledAmount.Level = level
		}
	}
}

// FillWithNameableKeys adds any nameable keys found in this Weapon to the provided map.
func (w *Weapon) FillWithNameableKeys(m map[string]string) {
	for _, one := range w.Defaults {
		one.FillWithNameableKeys(m)
	}
}

// ApplyNameableKeys replaces any nameable keys found in this Weapon with the corresponding values in the provided map.
func (w *Weapon) ApplyNameableKeys(m map[string]string) {
	for _, one := range w.Defaults {
		one.ApplyNameableKeys(m)
	}
}

// Container returns true if this is a container.
func (w *Weapon) Container() bool {
	return false
}

// Open returns true if this node is currently open.
func (w *Weapon) Open() bool {
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

// CellData returns the cell data information for the given column.
func (w *Weapon) CellData(columnID int, data *CellData) {
	var buffer xio.ByteBuffer
	data.Type = TextCellType
	switch columnID {
	case WeaponDescriptionColumn:
		data.Primary = w.String()
		data.Secondary = w.Notes()
	case WeaponUsageColumn:
		data.Primary = w.Usage
	case WeaponSLColumn:
		data.Primary = w.SkillLevel(&buffer).String()
	case WeaponParryColumn:
		parry := w.Parry.Resolve(w, &buffer)
		data.Primary = parry.String()
		data.Tooltip = parry.Tooltip(w)
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
		data.Primary = w.Range.Resolve(w, &buffer).String(w.PC() != nil)
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
		data.Type = PageRefCellType
	}
	if buffer.Len() > 0 {
		if data.Tooltip != "" {
			data.Tooltip += "\n\n"
		}
		data.Tooltip += i18n.Text("Includes modifiers from:") + buffer.String()
	}
}

// OwningEntity returns the owning Entity.
func (w *Weapon) OwningEntity() *Entity {
	return w.Entity()
}

// SetOwningEntity sets the owning entity and configures any sub-components as needed.
func (w *Weapon) SetOwningEntity(_ *Entity) {
}

// CopyFrom implements node.EditorData.
func (w *Weapon) CopyFrom(t *Weapon) {
	*w = *t.Clone(t.Entity(), nil, true)
}

// ApplyTo implements node.EditorData.
func (w *Weapon) ApplyTo(t *Weapon) {
	*t = *w.Clone(t.Entity(), nil, true)
}

// Validate ensures the weapon data is valid.
func (w *Weapon) Validate() {
	var zero uuid.UUID
	if w.WeaponData.ID == zero {
		w.WeaponData.ID = NewUUID()
	}
	w.Strength.Validate()
	switch w.Type {
	case MeleeWeaponType:
		w.Parry.Validate()
		w.Block.Validate()
		w.Reach.Validate()
		w.Accuracy = WeaponAccuracy{}
		w.Range = WeaponRange{}
		w.RateOfFire = WeaponRoF{}
		w.Shots = WeaponShots{}
		w.Bulk = WeaponBulk{}
		w.Recoil = WeaponRecoil{}
	case RangedWeaponType:
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
	default:
	}
}
