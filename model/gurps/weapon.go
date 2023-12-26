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

var _ Node[*Weapon] = &Weapon{}

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
	ID   uuid.UUID  `json:"id"`
	Type WeaponType `json:"type"`

	Jet                 bool `json:"jet,omitempty"`
	RetractingStock     bool `json:"retracting_stock,omitempty"`
	Thrown              bool `json:"thrown,omitempty"`
	ReloadTimeIsPerShot bool `json:"reload_time_is_per_shot,omitempty"`

	Usage         string         `json:"usage,omitempty"`
	UsageNotes    string         `json:"usage_notes,omitempty"`
	Damage        WeaponDamage   `json:"damage"`
	Parry         string         `json:"parry,omitempty"`
	ParryParts    WeaponParry    `json:"-"`
	Block         string         `json:"block,omitempty"`
	BlockParts    WeaponBlock    `json:"-"`
	Accuracy      string         `json:"accuracy,omitempty"`
	AccuracyParts WeaponAccuracy `json:"-"`
	Reach         string         `json:"reach,omitempty"`
	ReachParts    WeaponReach    `json:"-"`
	Range         string         `json:"range,omitempty"`
	RangeParts    WeaponRange    `json:"-"`
	Strength      string         `json:"strength,omitempty"`
	StrengthParts WeaponStrength `json:"-"`

	NormalBulk      fxp.Int    `json:"normal_bulk,omitempty"`
	GiantBulk       fxp.Int    `json:"giant_bulk,omitempty"`
	ShotRecoil      fxp.Int    `json:"shot_recoil,omitempty"`
	SlugRecoil      fxp.Int    `json:"slug_recoil,omitempty"`
	NonChamberShots fxp.Int    `json:"non_chamber_shots,omitempty"`
	ChamberShots    fxp.Int    `json:"chamber_shots,omitempty"`
	ShotDuration    fxp.Int    `json:"shot_duration,omitempty"`
	ReloadTime      fxp.Int    `json:"reload_time,omitempty"`
	RateOfFireMode1 RateOfFire `json:"rate_of_fire_mode_1,omitempty"`
	RateOfFireMode2 RateOfFire `json:"rate_of_fire_mode_2,omitempty"`

	Defaults []*SkillDefault `json:"defaults,omitempty"`
}

/*
v5.18 format:

type WeaponData struct {
	ID              uuid.UUID       `json:"id"`
	Type            WeaponType      `json:"type"`
	Damage          WeaponDamage    `json:"damage"`
	MinimumStrength string          `json:"strength,omitempty"`
	Usage           string          `json:"usage,omitempty"`
	UsageNotes      string          `json:"usage_notes,omitempty"`
	Reach           string          `json:"reach,omitempty"`
	Parry           string          `json:"parry,omitempty"`
	Block           string          `json:"block,omitempty"`
	Accuracy        string          `json:"accuracy,omitempty"`
	Range           string          `json:"range,omitempty"`
	RateOfFire      string          `json:"rate_of_fire,omitempty"`
	Shots           string          `json:"shots,omitempty"`
	Bulk            string          `json:"bulk,omitempty"`
	Recoil          string          `json:"recoil,omitempty"`
	Defaults        []*SkillDefault `json:"defaults,omitempty"`
}

Display order:
  Melee: Description, Usage, Skill Level, Parry, Block, Damage, Reach, ST
  Ranged: Description, Usage, Skill Level, Accuracy, Damage, Range, RoF, Shots, Bulk, Recoil, ST
*/

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
		w.ReachParts.Min = fxp.One
		w.ReachParts.Max = fxp.One
		w.Damage.StrengthType = ThrustStrengthDamage
	case RangedWeaponType:
		w.RateOfFireMode1.ShotsPerAttack = fxp.One
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
	_ = binary.Write(h, binary.LittleEndian, w.Jet)
	_ = binary.Write(h, binary.LittleEndian, w.NormalBulk)
	_ = binary.Write(h, binary.LittleEndian, w.GiantBulk)
	_ = binary.Write(h, binary.LittleEndian, w.RetractingStock)
	_ = binary.Write(h, binary.LittleEndian, w.ShotRecoil)
	_ = binary.Write(h, binary.LittleEndian, w.SlugRecoil)
	_ = binary.Write(h, binary.LittleEndian, w.NonChamberShots)
	_ = binary.Write(h, binary.LittleEndian, w.ChamberShots)
	_ = binary.Write(h, binary.LittleEndian, w.ShotDuration)
	_ = binary.Write(h, binary.LittleEndian, w.ReloadTime)
	_ = binary.Write(h, binary.LittleEndian, w.Thrown)
	_ = binary.Write(h, binary.LittleEndian, w.ReloadTimeIsPerShot)
	w.ParryParts.hash(h)
	w.BlockParts.hash(h)
	w.AccuracyParts.hash(h)
	w.ReachParts.hash(h)
	w.RangeParts.hash(h)
	w.StrengthParts.hash(h)
	w.RateOfFireMode1.hash(h)
	w.RateOfFireMode2.hash(h)
	return h.Sum32()
}

// MarshalJSON implements json.Marshaler.
func (w *Weapon) MarshalJSON() ([]byte, error) {
	w.WeaponData.Strength = w.StrengthParts.String()
	musclePowerIsResolved := w.PC() != nil
	switch w.Type {
	case MeleeWeaponType:
		w.WeaponData.Parry = w.ParryParts.String()
		w.WeaponData.Block = w.BlockParts.String()
		w.WeaponData.Reach = w.ReachParts.String()
	case RangedWeaponType:
		w.WeaponData.Accuracy = w.AccuracyParts.String(w)
		w.WeaponData.Range = w.RangeParts.String(musclePowerIsResolved)
	default:
	}
	type calc struct {
		Level      fxp.Int `json:"level,omitempty"`
		RateOfFire string  `json:"rate_of_fire,omitempty"`
		Damage     string  `json:"damage,omitempty"`
		Bulk       string  `json:"bulk,omitempty"`
		Shots      string  `json:"shots,omitempty"`
		// From here down are the new fields
		Parry    string `json:"parry,omitempty"`
		Block    string `json:"block,omitempty"`
		Accuracy string `json:"accuracy,omitempty"`
		Reach    string `json:"reach,omitempty"`
		Range    string `json:"range,omitempty"`
		Strength string `json:"strength,omitempty"`
	}
	data := struct {
		WeaponData
		Calc calc `json:"calc"`
	}{
		WeaponData: w.WeaponData,
		Calc: calc{
			Level:  w.SkillLevel(nil).Max(0),
			Damage: w.Damage.ResolvedDamage(nil),
		},
	}
	if data.Calc.Strength = w.StrengthParts.Resolve(w, nil).String(); data.Calc.Strength == w.WeaponData.Strength {
		data.Calc.Strength = ""
	}
	switch w.Type {
	case MeleeWeaponType:
		if data.Calc.Parry = w.ParryParts.Resolve(w, nil).String(); data.Calc.Parry == w.WeaponData.Parry {
			data.Calc.Parry = ""
		}
		if data.Calc.Block = w.BlockParts.Resolve(w, nil).String(); data.Calc.Block == w.WeaponData.Block {
			data.Calc.Block = ""
		}
		if data.Calc.Reach = w.ReachParts.Resolve(w, nil).String(); data.Calc.Reach == w.WeaponData.Reach {
			data.Calc.Reach = ""
		}
	case RangedWeaponType:
		if data.Calc.Accuracy = w.AccuracyParts.Resolve(w, nil).String(w); data.Calc.Accuracy == w.WeaponData.Accuracy {
			data.Calc.Accuracy = ""
		}
		if data.Calc.Range = w.RangeParts.Resolve(w, nil).String(musclePowerIsResolved); data.Calc.Range == w.WeaponData.Range {
			data.Calc.Range = ""
		}
		data.Calc.Bulk = w.CombinedBulk(nil)
		data.Calc.RateOfFire = w.CombinedRateOfFire(nil)
		data.Calc.Shots = w.CombinedShots(nil)
	default:
	}
	return json.Marshal(&data)
}

// UnmarshalJSON implements json.Unmarshaler.
func (w *Weapon) UnmarshalJSON(data []byte) error {
	type oldWeaponData struct {
		WeaponData
		OldBulk       string `json:"bulk"`
		OldRecoil     string `json:"recoil"`
		OldRateOfFire string `json:"rate_of_fire"`
		OldShots      string `json:"shots"`
	}
	var wdata oldWeaponData
	if err := json.Unmarshal(data, &wdata); err != nil {
		return err
	}
	w.WeaponData = wdata.WeaponData
	if strings.Contains(strings.ToLower(wdata.Accuracy), "jet") ||
		strings.Contains(strings.ToLower(wdata.OldRateOfFire), "jet") {
		w.Jet = true
	}
	w.StrengthParts = ParseWeaponStrength(wdata.Strength)
	switch w.WeaponData.Type {
	case MeleeWeaponType:
		w.ParryParts = ParseWeaponParry(wdata.Parry)
		w.BlockParts = ParseWeaponBlock(wdata.Block)
		w.ReachParts = ParseWeaponReach(wdata.Reach)
	case RangedWeaponType:
		if !w.Jet {
			w.AccuracyParts = ParseWeaponAccuracy(wdata.Accuracy)
			if wdata.OldRateOfFire != "" {
				parts := strings.Split(wdata.OldRateOfFire, "/")
				w.RateOfFireMode1.parseOldRateOfFire(parts[0])
				if len(parts) > 1 {
					w.RateOfFireMode2.parseOldRateOfFire(parts[1])
				}
			}
		}
		w.RangeParts = ParseWeaponRange(wdata.Range)
		if wdata.OldBulk != "" {
			w.RetractingStock = strings.Contains(wdata.OldBulk, "*")
			parts := strings.Split(wdata.OldBulk, "/")
			w.NormalBulk, _ = fxp.Extract(parts[0])
			if len(parts) > 1 {
				w.GiantBulk, _ = fxp.Extract(parts[1])
			}
		}
		if wdata.OldRecoil != "" {
			parts := strings.Split(wdata.OldRecoil, "/")
			w.ShotRecoil, _ = fxp.Extract(parts[0])
			if len(parts) > 1 {
				w.SlugRecoil, _ = fxp.Extract(parts[1])
			}
		}
		if wdata.OldShots != "" {
			lowered := strings.ToLower(wdata.OldShots)
			lowered = strings.ReplaceAll(lowered, " ", "")
			if !strings.Contains(lowered, "fp") &&
				!strings.Contains(lowered, "hrs") &&
				!strings.Contains(lowered, "day") {
				w.Thrown = strings.Contains(lowered, "t")
				if !strings.Contains(lowered, "spec") {
					w.NonChamberShots, lowered = fxp.Extract(lowered)
					if strings.HasPrefix(lowered, "+") {
						w.ChamberShots, lowered = fxp.Extract(lowered)
					}
					if strings.HasPrefix(lowered, "x") {
						w.ShotDuration, lowered = fxp.Extract(lowered[1:])
					}
					if strings.HasPrefix(lowered, "(") {
						w.ReloadTime, _ = fxp.Extract(lowered[1:])
						w.ReloadTimeIsPerShot = strings.Contains(lowered, "i")
					}
				}
			}
		}
	default:
	}
	var zero uuid.UUID
	if w.WeaponData.ID == zero {
		w.WeaponData.ID = NewUUID()
	}
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
	if minST := w.StrengthParts.Resolve(w, nil).Minimum - entity.StrikingStrength(); minST > 0 {
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
		w.ResolveBoolFlag(CanParryWeaponSwitchType, w.ParryParts.Permitted) &&
		w.ResolveBoolFlag(FencingWeaponSwitchType, w.ParryParts.Fencing) {
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

// ResolvedRateOfFire returns the resolved weapon and scope accuracies for this weapon.
func (w *Weapon) ResolvedRateOfFire(tooltip *xio.ByteBuffer) (shots1, secondary1, shots2, secondary2 fxp.Int) {
	if w.ResolveBoolFlag(JetWeaponSwitchType, w.Jet) {
		return 0, 0, 0, 0
	}
	pc := w.PC()
	if pc == nil {
		return w.RateOfFireMode1.ShotsPerAttack, w.RateOfFireMode1.SecondaryProjectiles,
			w.RateOfFireMode2.ShotsPerAttack, w.RateOfFireMode2.SecondaryProjectiles
	}
	var mode1, mode2 xio.ByteBuffer
	shots1, secondary1 = w.collectRateOfFireBonuses(&mode1, &w.RateOfFireMode1, WeaponRofMode1ShotsBonusFeatureType, WeaponRofMode1SecondaryBonusFeatureType)
	shots2, secondary2 = w.collectRateOfFireBonuses(&mode2, &w.RateOfFireMode2, WeaponRofMode2ShotsBonusFeatureType, WeaponRofMode2SecondaryBonusFeatureType)
	switch {
	case mode1.Len() == 0:
		AppendBufferOntoNewLine(tooltip, &mode2)
	case mode2.Len() != 0:
		if mode1.Len() != 0 {
			_ = mode1.InsertString(0, i18n.Text("First mode:\n"))
		}
		if mode2.Len() != 0 {
			_ = mode2.InsertString(0, i18n.Text("Second mode:\n"))
			if mode1.Len() != 0 {
				mode1.WriteString("\n\n")
			}
			mode1.WriteString(mode2.String())
		}
		AppendBufferOntoNewLine(tooltip, &mode1)
	default:
		AppendBufferOntoNewLine(tooltip, &mode1)
	}
	return shots1, secondary1, shots2, secondary2
}

func (w *Weapon) collectRateOfFireBonuses(tooltip *xio.ByteBuffer, rof *RateOfFire, shotsFeature, secondaryFeature FeatureType) (shots, secondary fxp.Int) {
	shots = rof.ShotsPerAttack
	secondary = rof.SecondaryProjectiles
	for _, bonus := range w.collectWeaponBonuses(1, tooltip, shotsFeature, secondaryFeature) {
		if bonus.Type == shotsFeature {
			shots += bonus.AdjustedAmount()
		} else if bonus.Type == secondaryFeature {
			secondary += bonus.AdjustedAmount()
		}
	}
	return shots.Ceil().Max(0), secondary.Ceil().Max(0)
}

// ResolvedBulk returns the resolved bulk for this weapon.
func (w *Weapon) ResolvedBulk(tooltip *xio.ByteBuffer) (normal, giant fxp.Int) {
	normal = w.NormalBulk
	giant = w.GiantBulk
	for _, bonus := range w.collectWeaponBonuses(1, tooltip, WeaponBulkBonusFeatureType) {
		normal += bonus.AdjustedAmount()
		giant += bonus.AdjustedAmount()
	}
	return normal.Min(0), giant.Min(0)
}

// ResolvedRecoil returns the resolved recoil for this weapon.
func (w *Weapon) ResolvedRecoil(tooltip *xio.ByteBuffer) (shot, slug fxp.Int) {
	shot = w.ShotRecoil
	slug = w.SlugRecoil
	if shot <= fxp.One && slug <= fxp.One {
		return shot, slug
	}
	for _, bonus := range w.collectWeaponBonuses(1, tooltip, WeaponRecoilBonusFeatureType) {
		shot += bonus.AdjustedAmount()
		slug += bonus.AdjustedAmount()
	}
	if w.ShotRecoil <= fxp.One {
		shot = w.ShotRecoil
	} else {
		shot = shot.Max(fxp.One)
	}
	if w.SlugRecoil <= fxp.One {
		slug = w.SlugRecoil
	} else {
		slug = slug.Max(fxp.One)
	}
	return shot, slug
}

// ResolvedShot returns the resolved shots for this weapon.
func (w *Weapon) ResolvedShot(tooltip *xio.ByteBuffer) (nonChamberShots, chamberShots, shotDuration, reloadTime fxp.Int) {
	nonChamberShots = w.NonChamberShots
	chamberShots = w.ChamberShots
	shotDuration = w.ShotDuration
	reloadTime = w.ReloadTime
	for _, bonus := range w.collectWeaponBonuses(1, tooltip, WeaponNonChamberShotsBonusFeatureType, WeaponChamberShotsBonusFeatureType, WeaponShotDurationBonusFeatureType, WeaponReloadTimeBonusFeatureType) {
		switch bonus.Type {
		case WeaponNonChamberShotsBonusFeatureType:
			nonChamberShots += bonus.AdjustedAmount()
		case WeaponChamberShotsBonusFeatureType:
			chamberShots += bonus.AdjustedAmount()
		case WeaponShotDurationBonusFeatureType:
			shotDuration += bonus.AdjustedAmount()
		case WeaponReloadTimeBonusFeatureType:
			reloadTime += bonus.AdjustedAmount()
		default:
		}
	}
	return nonChamberShots.Max(0), chamberShots.Max(0), shotDuration.Max(0), reloadTime.Max(0)
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
		parry := w.ParryParts.Resolve(w, &buffer)
		data.Primary = parry.String()
		data.Tooltip = parry.Tooltip(w)
	case WeaponBlockColumn:
		data.Primary = w.BlockParts.Resolve(w, &buffer).String()
	case WeaponDamageColumn:
		data.Primary = w.Damage.ResolvedDamage(&buffer)
	case WeaponReachColumn:
		reach := w.ReachParts.Resolve(w, &buffer)
		data.Primary = reach.String()
		data.Tooltip = reach.Tooltip()
	case WeaponSTColumn:
		weaponST := w.StrengthParts.Resolve(w, &buffer)
		data.Primary = weaponST.String()
		data.Tooltip = weaponST.Tooltip(w)
	case WeaponAccColumn:
		data.Primary = w.AccuracyParts.Resolve(w, &buffer).String(w)
	case WeaponRangeColumn:
		data.Primary = w.RangeParts.Resolve(w, &buffer).String(w.PC() != nil)
	case WeaponRoFColumn:
		data.Primary = w.CombinedRateOfFire(&buffer)
		if w.Type == RangedWeaponType && !w.ResolveBoolFlag(JetWeaponSwitchType, w.Jet) {
			var mode1, mode2 xio.ByteBuffer
			rof1 := w.RateOfFireMode1.Combined(&mode1)
			rof2 := w.RateOfFireMode2.Combined(&mode2)
			switch {
			case rof1 == "":
				data.Tooltip = mode2.String()
			case rof2 != "":
				if mode1.Len() != 0 {
					_ = mode1.InsertString(0, i18n.Text("First mode:\n"))
				}
				if mode2.Len() != 0 {
					_ = mode2.InsertString(0, i18n.Text("Second mode:\n"))
					if mode1.Len() != 0 {
						mode1.WriteString("\n\n")
					}
					mode1.WriteString(mode2.String())
				}
				data.Tooltip = mode1.String()
			default:
				data.Tooltip = mode1.String()
			}
		}
	case WeaponShotsColumn:
		data.Primary = w.CombinedShots(&buffer)
		if w.ResolveBoolFlag(ReloadTimeIsPerShotWeaponSwitchType, w.ReloadTimeIsPerShot) {
			data.Tooltip = i18n.Text("Reload time is per shot")
		}
	case WeaponBulkColumn:
		data.Primary = w.CombinedBulk(&buffer)
		if w.ResolveBoolFlag(RetractingStockWeaponSwitchType, w.RetractingStock) {
			wd := *w
			if wd.NormalBulk < 0 {
				wd.NormalBulk += fxp.One
			}
			if wd.GiantBulk < 0 {
				wd.GiantBulk += fxp.One
			}
			accuracy := w.AccuracyParts.Resolve(w, nil)
			accuracy.Base -= fxp.One
			if wd.ShotRecoil > fxp.One {
				wd.ShotRecoil += fxp.One
			}
			if wd.SlugRecoil > fxp.One {
				wd.SlugRecoil += fxp.One
			}
			data.Tooltip = fmt.Sprintf(i18n.Text("Has a retracting stock. With the stock folded, the weapon's stats change to Bulk %s, Accuracy %s, Recoil %s, and minimum ST %v. Folding or unfolding the stock takes one Ready maneuver."),
				wd.CombinedBulk(nil), accuracy.String(w), wd.CombinedRecoil(nil),
				w.StrengthParts.Resolve(w, nil).Minimum.Mul(fxp.OnePointTwo).Ceil())
		}
	case WeaponRecoilColumn:
		data.Primary = w.CombinedRecoil(&buffer)
		if strings.Contains(data.Primary, "/") {
			data.Tooltip = i18n.Text("First Recoil value is for shot, second is for slugs")
		}
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

// CombinedBulk returns the combined string used in the GURPS weapon tables for bulk.
func (w *Weapon) CombinedBulk(tooltip *xio.ByteBuffer) string {
	if w.Type != RangedWeaponType {
		return ""
	}
	normalBulk, giantBulk := w.ResolvedBulk(tooltip)
	if normalBulk >= 0 && giantBulk >= 0 {
		return ""
	}
	var buffer strings.Builder
	buffer.WriteString(normalBulk.String())
	if giantBulk != 0 && giantBulk != normalBulk {
		buffer.WriteByte('/')
		buffer.WriteString(giantBulk.String())
	}
	if w.ResolveBoolFlag(RetractingStockWeaponSwitchType, w.RetractingStock) {
		buffer.WriteByte('*')
	}
	return buffer.String()
}

// CombinedRecoil returns the combined string used in the GURPS weapon tables for recoil.
func (w *Weapon) CombinedRecoil(tooltip *xio.ByteBuffer) string {
	if w.Type != RangedWeaponType {
		return ""
	}
	shot, slug := w.ResolvedRecoil(tooltip)
	if shot == 0 && slug == 0 {
		return ""
	}
	var buffer strings.Builder
	buffer.WriteString(shot.String())
	if slug != 0 && shot != slug {
		buffer.WriteByte('/')
		buffer.WriteString(slug.String())
	}
	return buffer.String()
}

// CombinedRateOfFire returns the combined string used in the GURPS weapon tables for rate of fire.
func (w *Weapon) CombinedRateOfFire(tooltip *xio.ByteBuffer) string {
	if w.Type != RangedWeaponType {
		return ""
	}
	if w.ResolveBoolFlag(JetWeaponSwitchType, w.Jet) {
		return i18n.Text("Jet")
	}
	rofCopy1 := w.RateOfFireMode1
	rofCopy2 := w.RateOfFireMode2
	rofCopy1.ShotsPerAttack, rofCopy1.SecondaryProjectiles, rofCopy2.ShotsPerAttack, rofCopy2.SecondaryProjectiles = w.ResolvedRateOfFire(tooltip)
	rof1 := rofCopy1.Combined(nil)
	rof2 := rofCopy2.Combined(nil)
	if rof1 == "" {
		return rof2
	}
	if rof2 != "" {
		return rof1 + "/" + rof2
	}
	return rof1
}

// CombinedShots returns the combined string used in the GURPS weapon tables for shots.
func (w *Weapon) CombinedShots(tooltip *xio.ByteBuffer) string {
	if w.Type != RangedWeaponType {
		return ""
	}
	var buffer strings.Builder
	nonChamberShots, chamberShots, shotDuration, reloadTime := w.ResolvedShot(tooltip)
	if w.ResolveBoolFlag(ThrownWeaponSwitchType, w.Thrown) {
		buffer.WriteByte('T')
	} else {
		if nonChamberShots > 0 {
			buffer.WriteString(nonChamberShots.String())
			if chamberShots > 0 {
				buffer.WriteByte('+')
				buffer.WriteString(chamberShots.String())
			}
		}
	}
	if reloadTime > 0 {
		buffer.WriteByte('(')
		buffer.WriteString(reloadTime.String())
		if w.ResolveBoolFlag(ReloadTimeIsPerShotWeaponSwitchType, w.ReloadTimeIsPerShot) {
			buffer.WriteByte('i')
		}
		buffer.WriteByte(')')
	}
	if shotDuration > 0 {
		buffer.WriteByte('x')
		buffer.WriteString(shotDuration.String())
		buffer.WriteByte('s')
	}
	return buffer.String()
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
	w.StrengthParts.Validate()
	switch w.Type {
	case MeleeWeaponType:
		w.ParryParts.Validate()
		w.BlockParts.Validate()
		w.ReachParts.Validate()
	case RangedWeaponType:
		w.AccuracyParts.Validate()
		w.RangeParts.Validate()
	default:
	}
}
