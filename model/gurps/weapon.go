/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
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
	"fmt"
	"hash/fnv"
	"strings"
	"unsafe"

	"github.com/google/uuid"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/datafile"
	"github.com/richardwilkes/gcs/v5/model/gurps/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/gid"
	"github.com/richardwilkes/gcs/v5/model/gurps/skill"
	"github.com/richardwilkes/gcs/v5/model/gurps/weapon"
	"github.com/richardwilkes/gcs/v5/model/id"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/rpgtools/dice"
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
	FeatureList() feature.Features
	TagList() []string
}

// WeaponData holds the Weapon data that is written to disk.
type WeaponData struct {
	ID              uuid.UUID       `json:"id"`
	Type            weapon.Type     `json:"type"`
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

// Weapon holds the stats for a weapon.
type Weapon struct {
	WeaponData
	Owner WeaponOwner
}

// ExtractWeaponsOfType filters the input list down to only those weapons of the given type.
func ExtractWeaponsOfType(desiredType weapon.Type, list []*Weapon) []*Weapon {
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
		case weapon.Melee:
			melee = append(melee, w)
		case weapon.Ranged:
			ranged = append(ranged, w)
		}
	}
	return melee, ranged
}

// NewWeapon creates a new weapon of the given type.
func NewWeapon(owner WeaponOwner, weaponType weapon.Type) *Weapon {
	w := &Weapon{
		WeaponData: WeaponData{
			ID:   id.NewUUID(),
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
	case weapon.Melee:
		w.Reach = "1"
		w.Damage.StrengthType = weapon.Thrust
	case weapon.Ranged:
		w.RateOfFire = "1"
		w.Damage.Base = dice.New("1d")
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

// Less returns true if this weapon should be sorted above the other weapon.
func (w *Weapon) Less(other *Weapon) bool {
	s1 := w.String()
	s2 := other.String()
	if txt.NaturalLess(s1, s2, true) {
		return true
	}
	if s1 != s2 {
		return false
	}
	if txt.NaturalLess(w.Usage, other.Usage, true) {
		return true
	}
	if w.Usage != other.Usage {
		return false
	}
	if txt.NaturalLess(w.UsageNotes, other.UsageNotes, true) {
		return true
	}
	if w.UsageNotes != other.UsageNotes {
		return false
	}
	return uintptr(unsafe.Pointer(w)) < uintptr(unsafe.Pointer(other)) //nolint:gosec // Just need a tie-breaker
}

// HashCode returns a hash value for this weapon's resolved state.
func (w *Weapon) HashCode() uint32 {
	h := fnv.New32()
	h.Write([]byte(w.ID.String()))
	h.Write([]byte{byte(w.Type)})
	h.Write([]byte(w.String()))
	h.Write([]byte(w.UsageNotes))
	h.Write([]byte(w.Usage))
	h.Write([]byte(w.SkillLevel(nil).String()))
	h.Write([]byte(w.Accuracy))
	h.Write([]byte(w.Parry))
	h.Write([]byte(w.Block))
	h.Write([]byte(w.Damage.ResolvedDamage(nil)))
	h.Write([]byte(w.Reach))
	h.Write([]byte(w.Range))
	h.Write([]byte(w.RateOfFire))
	h.Write([]byte(w.Shots))
	h.Write([]byte(w.Bulk))
	h.Write([]byte(w.Recoil))
	h.Write([]byte(w.MinimumStrength))
	return h.Sum32()
}

// MarshalJSON implements json.Marshaler.
func (w *Weapon) MarshalJSON() ([]byte, error) {
	type calc struct {
		Level  fxp.Int `json:"level,omitempty"`
		Parry  string  `json:"parry,omitempty"`
		Block  string  `json:"block,omitempty"`
		Range  string  `json:"range,omitempty"`
		Damage string  `json:"damage,omitempty"`
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
	if w.Type == weapon.Melee {
		data.Calc.Parry = w.ResolvedParry(nil)
		data.Calc.Block = w.ResolvedBlock(nil)
	} else {
		data.Calc.Range = w.ResolvedRange()
	}
	return json.Marshal(&data)
}

// UnmarshalJSON implements json.Unmarshaler.
func (w *Weapon) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &w.WeaponData); err != nil {
		return err
	}
	var zero uuid.UUID
	if w.WeaponData.ID == zero {
		w.WeaponData.ID = id.NewUUID()
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
	if strings.TrimSpace(w.UsageNotes) != "" {
		if buffer.Len() != 0 {
			buffer.WriteByte('\n')
		}
		buffer.WriteString(w.UsageNotes)
	}
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
	if entity := w.Entity(); entity != nil && entity.Type == datafile.PC {
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
	if tooltip != nil && primaryTooltip != nil && primaryTooltip.Len() != 0 {
		if tooltip.Len() != 0 {
			tooltip.WriteByte('\n')
		}
		tooltip.WriteString(primaryTooltip.String())
	}
	if best < 0 {
		best = 0
	}
	return best
}

func (w *Weapon) skillLevelBaseAdjustment(entity *Entity, tooltip *xio.ByteBuffer) fxp.Int {
	var adj fxp.Int
	if minST := w.ResolvedMinimumStrength() - (entity.StrengthOrZero() + entity.StrikingStrengthBonus); minST > 0 {
		adj -= minST
	}
	nameQualifier := w.String()
	for _, bonus := range entity.NamedWeaponSkillBonusesFor(feature.WeaponNamedIDPrefix+"*", nameQualifier, w.Usage,
		w.Owner.TagList(), tooltip) {
		adj += bonus.AdjustedAmount()
	}
	for _, bonus := range entity.NamedWeaponSkillBonusesFor(feature.WeaponNamedIDPrefix+"/"+nameQualifier,
		nameQualifier, w.Usage, w.Owner.TagList(), tooltip) {
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
	if w.Type.EnsureValid() == weapon.Melee && strings.Contains(w.Parry, "F") {
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

func (w *Weapon) extractSkillBonusForThisWeapon(f feature.Feature, tooltip *xio.ByteBuffer) fxp.Int {
	if sb, ok := f.(*feature.SkillBonus); ok {
		if sb.SelectionType.EnsureValid() == skill.ThisWeapon {
			if sb.SpecializationCriteria.Matches(w.Usage) {
				sb.AddToTooltip(tooltip)
				return sb.AdjustedAmount()
			}
		}
	}
	return 0
}

// ResolvedParry returns the resolved parry level.
func (w *Weapon) ResolvedParry(tooltip *xio.ByteBuffer) string {
	return w.resolvedValue(w.Parry, gid.Parry, tooltip)
}

// ResolvedBlock returns the resolved block level.
func (w *Weapon) ResolvedBlock(tooltip *xio.ByteBuffer) string {
	return w.resolvedValue(w.Block, gid.Block, tooltip)
}

// ResolvedRange returns the range, fully resolved for the user's ST, if possible.
func (w *Weapon) ResolvedRange() string {
	//nolint:ifshort // No, pc isn't just used on the next line...
	pc := w.PC()
	if pc == nil {
		return w.Range
	}
	st := (pc.StrengthOrZero() + pc.ThrowingStrengthBonus).Trunc()
	var savedRange string
	calcRange := w.Range
	for calcRange != savedRange {
		savedRange = calcRange
		calcRange = w.resolveRange(calcRange, st)
	}
	return calcRange
}

func (w *Weapon) resolvedValue(input, baseDefaultType string, tooltip *xio.ByteBuffer) string {
	pc := w.PC()
	if pc == nil {
		return input
	}
	var buffer strings.Builder
	skillLevel := fxp.Max
	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		line := scanner.Text()
		if buffer.Len() != 0 {
			buffer.WriteByte('\n')
		}
		if line != "" {
			max := len(line)
			i := 0
			for i < max && line[i] == ' ' {
				i++
			}
			if i < max {
				ch := line[i]
				neg := false
				modifier := 0
				found := false
				if ch == '-' || ch == '+' {
					neg = ch == '-'
					i++
					if i < max {
						ch = line[i]
					}
				}
				for i < max && ch >= '0' && ch <= '9' {
					found = true
					modifier *= 10
					modifier += int(ch - '0')
					i++
					if i < max {
						ch = line[i]
					}
				}
				if found {
					if skillLevel == fxp.Max {
						var primaryTooltip, secondaryTooltip *xio.ByteBuffer
						if tooltip != nil {
							primaryTooltip = &xio.ByteBuffer{}
						}
						preAdj := w.skillLevelBaseAdjustment(pc, primaryTooltip)
						postAdj := w.skillLevelPostAdjustment(pc, primaryTooltip)
						adj := fxp.Three
						if baseDefaultType == gid.Parry {
							adj += pc.ParryBonus
						} else {
							adj += pc.BlockBonus
						}
						best := fxp.Min
						for _, def := range w.Defaults {
							level := def.SkillLevelFast(pc, false, nil, true)
							if level == fxp.Min {
								continue
							}
							level += preAdj
							if baseDefaultType != def.Type() {
								level = (level.Div(fxp.Two) + adj).Trunc()
							}
							level += postAdj
							var possibleTooltip *xio.ByteBuffer
							if def.Type() == gid.Skill && def.Name == "Karate" {
								if tooltip != nil {
									possibleTooltip = &xio.ByteBuffer{}
								}
								level += w.EncumbrancePenalty(pc, possibleTooltip)
							}
							if best < level {
								best = level
								secondaryTooltip = possibleTooltip
							}
						}
						if best != fxp.Min && tooltip != nil {
							if primaryTooltip != nil && primaryTooltip.Len() != 0 {
								if tooltip.Len() != 0 {
									tooltip.WriteByte('\n')
								}
								tooltip.WriteString(primaryTooltip.String())
							}
							if secondaryTooltip != nil && secondaryTooltip.Len() != 0 {
								if tooltip.Len() != 0 {
									tooltip.WriteByte('\n')
								}
								tooltip.WriteString(secondaryTooltip.String())
							}
						}
						skillLevel = best.Max(0)
					}
					if neg {
						modifier = -modifier
					}
					num := (skillLevel + fxp.From(modifier)).Trunc().String()
					if i < max {
						buffer.WriteString(num)
						line = line[i:]
					} else {
						line = num
					}
				}
			}
		}
		buffer.WriteString(line)
	}
	return buffer.String()
}

func (w *Weapon) resolveRange(inRange string, st fxp.Int) string {
	where := strings.IndexByte(inRange, 'x')
	if where == -1 {
		return inRange
	}
	last := where + 1
	max := len(inRange)
	if last < max && inRange[last] == ' ' {
		last++
	}
	if last >= max {
		return inRange
	}
	ch := inRange[last]
	found := false
	decimal := false
	started := last
	for (!decimal && ch == '.') || (ch >= '0' && ch <= '9') {
		found = true
		if ch == '.' {
			decimal = true
		}
		last++
		if last >= max {
			break
		}
		ch = inRange[last]
	}
	if !found {
		return inRange
	}
	value, err := fxp.FromString(inRange[started:last])
	if err != nil {
		return inRange
	}
	var buffer strings.Builder
	if where > 0 {
		buffer.WriteString(inRange[:where])
	}
	buffer.WriteString(value.Mul(st).Trunc().String())
	if last < max {
		buffer.WriteString(inRange[last:])
	}
	return buffer.String()
}

// ResolvedMinimumStrength returns the resolved minimum strength required to use this weapon, or 0 if there is none.
func (w *Weapon) ResolvedMinimumStrength() fxp.Int {
	started := false
	value := 0
	for _, ch := range w.MinimumStrength {
		if ch >= '0' && ch <= '9' {
			value *= 10
			value += int(ch - '0')
			started = true
		} else if started {
			break
		}
	}
	return fxp.From(value)
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
func (w *Weapon) CellData(column int, data *CellData) {
	var buffer xio.ByteBuffer
	data.Type = Text
	switch column {
	case WeaponDescriptionColumn:
		data.Primary = w.String()
		data.Secondary = w.Notes()
	case WeaponUsageColumn:
		data.Primary = w.Usage
	case WeaponSLColumn:
		data.Primary = w.SkillLevel(&buffer).String()
	case WeaponParryColumn:
		data.Primary = w.ResolvedParry(&buffer)
	case WeaponBlockColumn:
		data.Primary = w.ResolvedBlock(&buffer)
	case WeaponDamageColumn:
		data.Primary = w.Damage.ResolvedDamage(&buffer)
	case WeaponReachColumn:
		data.Primary = w.Reach
	case WeaponSTColumn:
		data.Primary = w.MinimumStrength
	case WeaponAccColumn:
		data.Primary = w.Accuracy
	case WeaponRangeColumn:
		data.Primary = w.ResolvedRange()
	case WeaponRoFColumn:
		data.Primary = w.RateOfFire
	case WeaponShotsColumn:
		data.Primary = w.Shots
	case WeaponBulkColumn:
		data.Primary = w.Bulk
	case WeaponRecoilColumn:
		data.Primary = w.Recoil
	case PageRefCellAlias:
		data.Type = PageRef
	}
	if buffer.Len() > 0 {
		data.Tooltip = i18n.Text("Includes modifiers from:") + buffer.String()
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
