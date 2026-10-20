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
	"cmp"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"hash"
	"slices"
	"strings"
	"sync"
	"unsafe"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/attack"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selector"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/skillsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stdmg"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/gcs/v5/model/txtutil"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/zeebo/xxh3"
)

var _ = assertNode[*Weapon]

const currentWeaponSubVersion = 1

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
	ResolveLocalNotes() string
	// ActiveFeatures returns the owner's own features that currently take effect (i.e. excluding any switchable
	// features while the owner's switch is off).
	ActiveFeatures() Features
	TagList() []string
	RatedStrength() fxp.Int
}

// WeaponData holds the Weapon data that is written to disk.
type WeaponData struct {
	TID        tid.TID         `json:"id"`
	SubVersion int             `json:"sv"`
	Damage     WeaponDamage    `json:"damage"`
	Strength   attack.Strength `json:"strength,omitzero"`
	Usage      string          `json:"usage,omitzero"`
	UsageNotes string          `json:"usage_notes,omitzero"`
	Reach      attack.Reach    `json:"reach,omitzero"`
	Parry      attack.Parry    `json:"parry,omitzero"`
	Block      attack.Block    `json:"block,omitzero"`
	Accuracy   attack.Accuracy `json:"accuracy,omitzero"`
	Range      attack.Range    `json:"range,omitzero"`
	RateOfFire attack.RoF      `json:"rate_of_fire,omitzero"`
	Shots      attack.Shots    `json:"shots,omitzero"`
	Bulk       attack.Bulk     `json:"bulk,omitzero"`
	Recoil     attack.Recoil   `json:"recoil,omitzero"`
	Defaults   []*SkillDefault `json:"defaults,omitempty"`
	Hide       bool            `json:"hide,omitzero"`
}

// Weapon holds the stats for a weapon. What its methods, and those of its damage, resolve is kept until its entity's
// caches are discarded (see Entity.DiscardCaches), unless the weapon's own data changes or it is given another owner
// first. A change to the owner's data, such as its rated ST or features, or to the rest of the entity is therefore
// seen only after a recalculation or DiscardCaches.
type Weapon struct {
	WeaponData
	Owner         WeaponOwner
	ClonedFromTID tid.TID
	NotEquipped   bool
	NotCarried    bool
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

// NewWeapon creates a new weapon of the given type.
func NewWeapon(owner WeaponOwner, melee bool) *Weapon {
	var w Weapon
	w.TID = tid.MustNewTID(weaponKind(melee))
	w.SubVersion = currentWeaponSubVersion
	w.Damage = WeaponDamage{
		Type:                      "cr",
		StrengthMultiplier:        fxp.One,
		ArmorDivisor:              fxp.One,
		FragmentationArmorDivisor: fxp.One,
	}
	w.Owner = owner
	if melee {
		w.Reach.Min = fxp.One
		w.Reach.Max = fxp.One
		w.Damage.StrengthType = stdmg.Thrust
	} else {
		w.RateOfFire.Mode1.ShotsPerAttack = fxp.One
		w.Damage.Base = "1d"
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

// CloneWeapons clones the input list of weapons, pointing each copy at the given owner. Weapon.Clone() keeps the
// original's owner, and a weapon resolves its nameable replacements, skill defaults and strength requirements through
// its owner, so a copy left pointing at the original's owner would report that node's values rather than its own.
func CloneWeapons(list []*Weapon, owner WeaponOwner, mode CloneMode) []*Weapon {
	if len(list) == 0 {
		return nil
	}
	weapons := make([]*Weapon, len(list))
	for i, w := range list {
		weapons[i] = w.Clone(LibraryFile{}, nil, nil, mode)
		weapons[i].SetOwner(owner)
	}
	return weapons
}

// Clone implements Node.
func (w *Weapon) Clone(_ LibraryFile, _ DataOwner, _ *Weapon, mode CloneMode) *Weapon {
	other := *w
	if mode != Copy {
		other.TID = tid.MustNewTID(w.TID[0])
		other.ClonedFromTID = w.TID
	}
	other.Damage = *w.Damage.Clone(&other)
	other.Defaults = cloneSkillDefaults(w.Defaults)
	return &other
}

// Compare returns an integer indicating the sort order of this weapon compared to the other weapon.
func (w *Weapon) Compare(other *Weapon) int {
	result := xstrings.NaturalCmp(w.String(), other.String(), true)
	if result == 0 {
		if result = xstrings.NaturalCmp(w.UsageWithReplacements(), other.UsageWithReplacements(), true); result == 0 {
			if result = xstrings.NaturalCmp(w.ResolveUsageNotes(), other.ResolveUsageNotes(), true); result == 0 {
				//nolint:gosec // Just need a tie-breaker
				result = cmp.Compare(uintptr(unsafe.Pointer(w)), uintptr(unsafe.Pointer(other)))
			}
		}
	}
	return result
}

// HashDisplayedState returns a hash value for this weapon's displayed state. Useful for detecting unique weapon entries
// to list on the sheet. It reads values kept as Weapon describes.
func (w *Weapon) HashDisplayedState() uint64 {
	h := xxh3.New()
	w.Hash(h)
	xhash.Bool(h, w.Hide)
	xhash.StringWithLen(h, w.String())
	xhash.StringWithLen(h, w.Notes())
	xhash.StringWithLen(h, w.UsageWithReplacements())
	xhash.Num64(h, w.SkillLevel(nil))
	xhash.StringWithLen(h, w.Damage.ResolvedDamage(nil))
	xhash.StringWithLen(h, w.ResolvedStrength(nil).String())
	if w.IsMelee() {
		xhash.StringWithLen(h, w.ResolvedParry(nil).String())
		xhash.StringWithLen(h, w.ResolvedBlock(nil).String())
		xhash.StringWithLen(h, w.ResolvedReach(nil).String())
	} else {
		xhash.StringWithLen(h, w.ResolvedAccuracy(nil).String())
		xhash.StringWithLen(h, w.ResolvedRange(nil).String(w.musclePowerIsResolved()))
		xhash.StringWithLen(h, w.ResolvedRateOfFire(nil).String())
		xhash.StringWithLen(h, w.ResolvedShots(nil).String())
		xhash.StringWithLen(h, w.ResolvedBulk(nil).String())
		xhash.StringWithLen(h, w.ResolvedRecoil(nil).String())
	}
	return h.Sum64()
}

// Hash writes this object's contents into the hasher. Note that this only hashes the data that is considered to be
// "source" data, i.e. not expected to be modified by the user after copying from a library.
func (w *WeaponData) Hash(h hash.Hash) {
	if w == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, syncKind(w))
	w.Damage.Hash(h)
	w.Strength.Hash(h)
	xhash.StringWithLen(h, w.Usage)
	xhash.StringWithLen(h, w.UsageNotes)
	w.Reach.Hash(h)
	w.Parry.Hash(h)
	w.Block.Hash(h)
	w.Accuracy.Hash(h)
	w.Range.Hash(h)
	w.RateOfFire.Hash(h)
	w.Shots.Hash(h)
	w.Bulk.Hash(h)
	w.Recoil.Hash(h)
	hashList(h, w.Defaults)
}

// MarshalJSONTo implements json.MarshalerTo.
func (w *Weapon) MarshalJSONTo(enc *jsontext.Encoder) error {
	type calc struct {
		Level      fxp.Int `json:"level,omitzero"`
		Damage     string  `json:"damage,omitzero"`
		Parry      string  `json:"parry,omitzero"`
		Block      string  `json:"block,omitzero"`
		Accuracy   string  `json:"accuracy,omitzero"`
		Reach      string  `json:"reach,omitzero"`
		Range      string  `json:"range,omitzero"`
		RateOfFire string  `json:"rate_of_fire,omitzero"`
		Shots      string  `json:"shots,omitzero"`
		Bulk       string  `json:"bulk,omitzero"`
		Recoil     string  `json:"recoil,omitzero"`
		Strength   string  `json:"strength,omitzero"`
	}
	w.SubVersion = currentWeaponSubVersion
	data := w.WeaponData
	// Clearing the fields that don't apply to this kind of weapon is part of writing out the data, so it happens
	// whether or not the derived values are wanted.
	melee := w.IsMelee()
	if melee {
		data.Accuracy = attack.Accuracy{}
		data.Range = attack.Range{}
		data.RateOfFire = attack.RoF{}
		data.Shots = attack.Shots{}
		data.Bulk = attack.Bulk{}
		data.Recoil = attack.Recoil{}
	} else {
		data.Parry = attack.Parry{}
		data.Block = attack.Block{}
		data.Reach = attack.Reach{}
	}
	return marshalNodeData(enc, &data, func() *calc {
		c := &calc{
			Level:  w.SkillLevel(nil).Max(0),
			Damage: w.Damage.ResolvedDamage(nil),
		}
		if c.Strength = w.ResolvedStrength(nil).String(); c.Strength == w.Strength.String() {
			c.Strength = ""
		}
		if melee {
			if c.Parry = w.ResolvedParry(nil).String(); c.Parry == w.Parry.String() {
				c.Parry = ""
			}
			if c.Block = w.ResolvedBlock(nil).String(); c.Block == w.Block.String() {
				c.Block = ""
			}
			if c.Reach = w.ResolvedReach(nil).String(); c.Reach == w.Reach.String() {
				c.Reach = ""
			}
		} else {
			if c.Accuracy = w.ResolvedAccuracy(nil).String(); c.Accuracy == w.Accuracy.String() {
				c.Accuracy = ""
			}
			if c.Range = w.ResolvedRange(nil).String(w.musclePowerIsResolved()); c.Range == w.Range.String(false) {
				c.Range = ""
			}
			if c.RateOfFire = w.ResolvedRateOfFire(nil).String(); c.RateOfFire == w.RateOfFire.String() {
				c.RateOfFire = ""
			}
			if c.Shots = w.ResolvedShots(nil).String(); c.Shots == w.Shots.String() {
				c.Shots = ""
			}
			if c.Bulk = w.ResolvedBulk(nil).String(); c.Bulk == w.Bulk.String() {
				c.Bulk = ""
			}
			if c.Recoil = w.ResolvedRecoil(nil).String(); c.Recoil == w.Recoil.String() {
				c.Recoil = ""
			}
		}
		if *c == (calc{}) {
			return nil
		}
		return c
	})
}

func (w *Weapon) musclePowerIsResolved() bool {
	return w.Entity() != nil || w.ratedStrength() > 0
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (w *Weapon) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var localData struct {
		WeaponData
		// Old data fields
		Type string `json:"type"`
	}
	if err := json.UnmarshalDecode(dec, &localData); err != nil {
		return err
	}
	if !tid.IsValid(localData.TID) {
		// Fixup old data that used UUIDs instead of TIDs
		localData.TID = tid.MustNewTID(weaponKind(localData.Type == "melee_weapon"))
	}
	w.WeaponData = localData.WeaponData
	// A JSON null in the defaults array decodes into a nil pointer without error, and every walker over the defaults
	// dereferences it, so drop such entries the same way Skill.UnmarshalJSONFrom does.
	w.Defaults = slices.DeleteFunc(w.Defaults, func(one *SkillDefault) bool { return one == nil })
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
func (w *WeaponData) ID() tid.TID {
	return w.TID
}

// Kind returns the kind of data.
func (w *Weapon) Kind() string {
	if w.IsMelee() {
		return i18n.Text("Melee Attack")
	}
	return i18n.Text("Ranged Attack")
}

func (w *Weapon) String() string {
	if xreflect.IsNil(w.Owner) {
		return ""
	}
	s := w.Owner.String()
	if w.NotCarried {
		s += " (" + i18n.Text("not carried") + ")"
	} else if w.NotEquipped {
		s += " (" + i18n.Text("not equipped") + ")"
	}
	return s
}

// Notes returns the notes for this weapon.
func (w *Weapon) Notes() string {
	var buffer strings.Builder
	w.forEachOwnerModifier(func(mod ownerModifier, _ Features) {
		if mod.ShowsNotesOnWeapon() {
			txtutil.AppendStringOntoNewLine(&buffer, strings.TrimSpace(mod.ResolveLocalNotes()))
		}
	})
	txtutil.AppendStringOntoNewLine(&buffer, strings.TrimSpace(w.ResolveUsageNotes()))
	return buffer.String()
}

// ownerModifier is what a weapon needs from the modifiers of its owner. Both TraitModifier and EquipmentModifier
// satisfy it.
type ownerModifier interface {
	fmt.Stringer
	ShowsNotesOnWeapon() bool
	ResolveLocalNotes() string
}

// forEachOwnerModifier calls visit for each enabled, non-container modifier of the weapon's owner, at any depth,
// together with those of the modifier's features that currently take effect, i.e. excluding its switchable features
// while the owner's switch is off. Only traits and equipment carry modifiers, so a weapon with any other owner, or no
// owner at all, visits nothing.
func (w *Weapon) forEachOwnerModifier(visit func(mod ownerModifier, active Features)) {
	switch owner := w.Owner.(type) {
	case *Trait:
		visitEnabledModifiers(owner.Modifiers, owner.SwitchedOn,
			func(mod *TraitModifier) Features { return mod.Features },
			func(mod *TraitModifier, active Features) { visit(mod, active) })
	case *Equipment:
		visitEnabledModifiers(owner.Modifiers, owner.SwitchedOn,
			func(mod *EquipmentModifier) Features { return mod.Features },
			func(mod *EquipmentModifier, active Features) { visit(mod, active) })
	}
}

// SetOwner sets the owner and ensures sub-components have their owners set.
func (w *Weapon) SetOwner(owner WeaponOwner) {
	w.Owner = owner
	w.Damage.Owner = w
	w.performDataSubVersionFixups()
}

func (w *Weapon) performDataSubVersionFixups() {
	switch w.SubVersion {
	case currentWeaponSubVersion:
	case 0: // Prior to sub-versioning
		if leveler, ok := w.Owner.(interface{ IsLeveled() bool }); ok && leveler.IsLeveled() {
			w.Damage.BaseLeveled = w.Damage.Base
			w.Damage.Base = ""
		}
	}
	w.SubVersion = currentWeaponSubVersion
}

// DataOwner returns the weapon owner's data owner.
func (w *Weapon) DataOwner() DataOwner {
	if xreflect.IsNil(w.Owner) {
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

// effectiveStrength returns the ST the weapon is used at: the rated ST of its owner when the owner has one, otherwise
// the entity's ST as chosen by defaultST (0 when there is no entity). The weapon's effective ST bonuses are then
// applied, explained in tooltip when it is not nil, and the result is capped at three times the weapon's minimum ST.
func (w *Weapon) effectiveStrength(defaultST func(entity *Entity) fxp.Int, tooltip *xbytes.InsertBuffer) fxp.Int {
	st := w.ratedStrength()
	if st == 0 {
		if entity := w.Entity(); entity != nil {
			st = defaultST(entity)
		}
	}
	st = max(w.weaponAdjustment(oneDieCount, tooltip, feature.WeaponEffectiveSTBonus).applyTo(st), 0)
	if maxST := w.ResolvedStrength(nil).Min.Mul(fxp.Three); maxST > 0 && maxST < st {
		st = maxST
	}
	return st
}

// SkillLevel returns the resolved skill level. The result is kept as Weapon describes.
func (w *Weapon) SkillLevel(tooltip *xbytes.InsertBuffer) fxp.Int {
	return memoize(w, func(m *weaponMemo) *memoized[fxp.Int] { return &m.skillLevel }, tooltip, w.skillLevel)
}

func (w *Weapon) skillLevel(tooltip *xbytes.InsertBuffer) fxp.Int {
	entity := w.Entity()
	if entity == nil {
		return 0
	}
	var primaryTooltip *xbytes.InsertBuffer
	if tooltip != nil {
		primaryTooltip = &xbytes.InsertBuffer{}
	}
	adj := w.skillLevelBaseAdjustment(entity, primaryTooltip) + w.skillLevelPostAdjustment(entity, primaryTooltip)
	best := fxp.Min
	replacements := w.NameableReplacements()
	for _, def := range w.Defaults {
		// A parry- or block-type default deliberately resolves to its defense level here, not the named skill's level.
		// Such a default exists to feed the parry or block calculation -- the Tonfa in the High Tech library carries
		// "Brawling Parry" and "Karate Parry" alongside its real attack defaults -- and scoring it by the named skill
		// would let a high Karate become the weapon's attack skill. Halved, it stays out of the way.
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
	txtutil.AppendBufferOntoNewLine(tooltip, primaryTooltip)
	if best < 0 {
		best = 0
	}
	return best
}

// usesCrossbowSkill reports whether the weapon is wielded with the Crossbow skill. A default may name it outright,
// which counts whether or not the character has the skill, or select it by some other criteria (a tag, say), in which
// case the skill the default resolves to decides.
func (w *Weapon) usesCrossbowSkill(e *Entity) bool {
	replacements := w.NameableReplacements()
	for _, def := range w.Defaults {
		if !def.SkillBased() {
			continue
		}
		if def.Name.Compare.EnsureValid() == criteria.IsText && def.NameWithReplacements(replacements) == "Crossbow" {
			return true
		}
		if sk := def.bestFastMatchingSkill(e, replacements, false, nil); sk != nil && sk.NameWithReplacements() == "Crossbow" {
			return true
		}
	}
	return false
}

func (w *Weapon) skillLevelBaseAdjustment(e *Entity, tooltip *xbytes.InsertBuffer) fxp.Int {
	var adj fxp.Int
	minST := w.ResolvedStrength(nil).Min
	switch w.Damage.resolvedStrengthType(nil) {
	case stdmg.TelekineticThrust, stdmg.TelekineticSwing:
		minST -= e.TelekineticStrength()
	case stdmg.IQThrust, stdmg.IQSwing:
		minST -= e.ResolveAttributeCurrent(IntelligenceID).Max(0).Floor()
	default:
		if !w.IsRanged() || (w.ResolveBoolFlag(wswitch.MusclePowered, w.Range.MusclePowered) && !w.usesCrossbowSkill(e)) {
			minST -= e.StrikingStrength()
		} else {
			minST -= e.LiftingStrength()
		}
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
	for _, f := range w.Owner.ActiveFeatures() {
		adj += w.extractSkillBonusForThisWeapon(f, tooltip)
	}
	w.forEachOwnerModifier(func(_ ownerModifier, active Features) {
		for _, f := range active {
			adj += w.extractSkillBonusForThisWeapon(f, tooltip)
		}
	})
	return adj
}

func (w *Weapon) skillLevelPostAdjustment(e *Entity, tooltip *xbytes.InsertBuffer) fxp.Int {
	if w.IsMelee() &&
		// Cannot use w.ResolvedParry() here, because it ends up calling this
		w.ResolveBoolFlag(wswitch.CanParry, w.Parry.CanParry) &&
		w.ResolveBoolFlag(wswitch.Fencing, w.Parry.Fencing) {
		return w.EncumbrancePenalty(e, tooltip)
	}
	return 0
}

// resolveDefenseModifier resolves the modifier of one of this weapon's defenses -- parry or block, as named by
// defenseID -- from the weapon's defaults, starting from the weapon's own modifier in current and applying the
// entity's bonus to that defense (whose explanation is defenseBonusTooltip) and the weapon bonuses of bonusType. It
// returns 0 when no default resolves.
func (w *Weapon) resolveDefenseModifier(entity *Entity, modifiersTooltip *xbytes.InsertBuffer, current fxp.Int, defenseID string, bonusType feature.Type, defenseBonusTooltip string) fxp.Int {
	defenseBonus := entity.defenseBonus(defenseID)
	var primaryTooltip *xbytes.InsertBuffer
	if modifiersTooltip != nil {
		primaryTooltip = &xbytes.InsertBuffer{}
	}
	preAdj := w.skillLevelBaseAdjustment(entity, primaryTooltip)
	postAdj := w.skillLevelPostAdjustment(entity, primaryTooltip)
	replacements := w.NameableReplacements()
	best := fxp.Min
	for _, def := range w.Defaults {
		// A default of the other defense type names a skill, not a basis for this defense, so resolve it as one of
		// this defense's type.
		def = def.asDefense(defenseID)
		var level fxp.Int
		if def.Type() == defenseID {
			// A defense-type default names the skill whose defense is wanted, so the skill-level adjustment has to be
			// folded in before the conversion to a defense level, exactly as it is on the other path. SkillLevelFast()
			// would hand back an already-converted level, and adding the adjustment to that would weight it twice.
			level = def.defenseLevelFast(entity, replacements, preAdj, defenseBonus)
			if level == fxp.Min {
				continue
			}
		} else {
			level = def.SkillLevelFast(entity, replacements, false, nil, true)
			if level == fxp.Min {
				continue
			}
			level = defenseLevelFromSkill(level+preAdj, defenseBonus)
		}
		level += postAdj
		if best < level {
			best = level
		}
	}
	if best == fxp.Min {
		return 0
	}
	txtutil.AppendBufferOntoNewLine(modifiersTooltip, primaryTooltip)
	txtutil.AppendStringOntoNewLine(modifiersTooltip, defenseBonusTooltip)
	adj := w.weaponAdjustment(w.baseDamageDieCount, modifiersTooltip, bonusType)
	return adj.applyTo(current + best).Max(0).Floor()
}

// EncumbrancePenalty returns the current encumbrance penalty.
func (w *Weapon) EncumbrancePenalty(e *Entity, tooltip *xbytes.InsertBuffer) fxp.Int {
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

func (w *Weapon) extractSkillBonusForThisWeapon(f Feature, tooltip *xbytes.InsertBuffer) fxp.Int {
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

// ResolveBoolFlag returns the resolved value of the given bool flag. It reads values kept as Weapon describes.
func (w *Weapon) ResolveBoolFlag(switchType wswitch.Type, initial bool) bool {
	entity := w.Entity()
	if entity == nil {
		return initial
	}
	t := 0
	f := 0
	for _, bonus := range w.applicableWeaponBonuses(entity) {
		if bonus.Type == feature.WeaponSwitch && bonus.SwitchType == switchType {
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

// ResolveSelector resolves a weapon-scoped multi-state field, applying any SelectorOverride features that target the
// given field and select this weapon: by the skill it is used with, by its name, or by being a weapon of the item
// carrying the override. base is the field's intrinsic value. When more than one override matches, the winner is
// chosen by the override-resolution ladder (priority, then specificity), and, if tooltip is non-nil, the full contest
// is reported. It reads values kept as Weapon describes.
func (w *Weapon) ResolveSelector(field selector.Field, base string, tooltip *xbytes.InsertBuffer) string {
	entity := w.Entity()
	if entity == nil {
		return base
	}
	return ResolveOverride(base, w.collectSelectorOverrides(entity, field), func(s string) string { return s }, tooltip)
}

// collectSelectorOverrides returns the candidates for the given field: the entity's overrides selecting weapons by
// required skill or by name that match this weapon, and those of the weapon's owner, and of its enabled modifiers, that
// select the owner's weapons or name this one. The skills the weapon is used with are only resolved when an override
// selects by them, since this runs for every field of every weapon displayed.
func (w *Weapon) collectSelectorOverrides(entity *Entity, field selector.Field) []OverrideCandidate[string] {
	var candidates []OverrideCandidate[string]
	seen := make(map[*SelectorOverride]bool)
	add := func(o *SelectorOverride) {
		if !seen[o] {
			seen[o] = true
			candidates = append(candidates, OverrideCandidate[string]{Value: o.Value, Override: o})
		}
	}
	var bySkill []*SelectorOverride
	for _, o := range entity.features.selectorOverrides {
		if o.Field != field {
			continue
		}
		switch o.SelectionType {
		case wsel.ThisWeapon: // Collected from the owner below
		case wsel.WithName:
			if o.MatchesWeapon(w) {
				add(o)
			}
		default:
			bySkill = append(bySkill, o)
		}
	}
	if len(bySkill) != 0 {
		usage := w.UsageWithReplacements()
		tags := w.Owner.TagList()
		w.forEachRequiredSkill(entity, func(name, specialization string) {
			rsl := sync.OnceValue(func() fxp.Int { return entity.relativeSkillLevel(name, specialization) })
			for _, o := range bySkill {
				if o.matchesRequiredSkill(name, specialization, usage, tags, rsl) {
					add(o)
				}
			}
		})
	}
	for _, f := range w.Owner.ActiveFeatures() {
		w.extractSelectorOverride(f, field, nil, add)
	}
	w.forEachOwnerModifier(func(mod ownerModifier, active Features) {
		for _, f := range active {
			w.extractSelectorOverride(f, field, mod, add)
		}
	})
	return candidates
}

// extractSelectorOverride hands f to add if it is an override of the given field, carried by the weapon's owner or by
// the owner's modifier sub, that selects the owner's weapons or names this one. The owner is set first: an override of
// equipment that isn't equipped never reaches the entity's lists, so nothing else sets it, and it supplies the
// substitutions the criteria are filled from and the name the tooltip reports.
func (w *Weapon) extractSelectorOverride(f Feature, field selector.Field, sub fmt.Stringer, add func(*SelectorOverride)) {
	o, ok := f.(*SelectorOverride)
	if !ok || o.Field != field {
		return
	}
	o.SetOwner(w.Owner)
	o.SetSubOwner(sub)
	switch o.SelectionType {
	case wsel.ThisWeapon:
		if o.matchesThisWeapon(w) {
			add(o)
		}
	case wsel.WithName:
		if o.MatchesWeapon(w) {
			add(o)
		}
	default: // Selected by required skill, which only the entity's lists resolve
	}
}

// dieCountFunc supplies the number of dice that a per-die weapon bonus is multiplied by when reporting what it
// contributes. It is a function rather than a value because resolving the weapon's base damage dice is expensive and is
// only needed when a per-die bonus actually matches the weapon.
type dieCountFunc func() fxp.Int

// oneDieCount reports a single die. Used for the bonuses that participate in resolving the weapon's base damage dice --
// minimum ST, effective ST and the weapon switches -- since consulting those dice here would recurse. This matches what
// AdjustedAmountForWeapon applies for the ST bonuses, so the reported and applied amounts agree.
func oneDieCount() fxp.Int {
	return fxp.One
}

// baseDamageDieCount returns the number of dice in the weapon's base damage. This is the multiplier
// AdjustedAmountForWeapon applies to a per-die bonus, so reporting the same count keeps a tooltip in step with the
// amount actually applied. Must not be used for the bonuses that resolving the base damage dice itself depends upon;
// see oneDieCount.
func (w *Weapon) baseDamageDieCount() fxp.Int {
	return fxp.FromInteger(w.Damage.BaseDamageDice().Count)
}

// collectWeaponBonuses returns the weapon bonuses of the given feature types that apply to this weapon, in the order
// they are found, and, if tooltip isn't nil, adds each to it with the die count the bonus scales by. The bonuses of
// every type are found once per weapon per state of the entity (see applicableWeaponBonuses); this picks out the
// types asked for.
func (w *Weapon) collectWeaponBonuses(dieCount dieCountFunc, tooltip *xbytes.InsertBuffer, allowedFeatureTypes ...feature.Type) []*WeaponBonus {
	entity := w.Entity()
	if entity == nil {
		return nil
	}
	dieCount = sync.OnceValue(dieCount) // At most one resolution per collection, no matter how many bonuses match
	var result []*WeaponBonus
	for _, bonus := range w.applicableWeaponBonuses(entity) {
		if !slices.Contains(allowedFeatureTypes, bonus.Type) {
			continue
		}
		if tooltip != nil {
			bonus.addToTooltip(bonus.resolveDieCount(dieCount), bonus.DerivedLeveledOwner(), tooltip)
		}
		result = append(result, bonus)
	}
	return result
}

// applicableWeaponBonuses returns every weapon bonus, of every type, that applies to this weapon: those of the entity
// that select it by the skill it is used with or by its name, and those of its owner and of its owner's enabled
// modifiers that select the owner's weapons or name this one. They come in the order they are found, without
// duplicates, from the weapon's memo when it has one.
func (w *Weapon) applicableWeaponBonuses(entity *Entity) []*WeaponBonus {
	m := w.memo()
	if m != nil && m.hasBonuses {
		return m.bonuses
	}
	bonuses := w.findWeaponBonuses(entity)
	if m != nil {
		m.bonuses = bonuses
		m.hasBonuses = true
	}
	return bonuses
}

// findWeaponBonuses is applicableWeaponBonuses without the memo.
func (w *Weapon) findWeaponBonuses(entity *Entity) []*WeaponBonus {
	var result []*WeaponBonus
	seen := make(map[*WeaponBonus]bool)
	add := func(bonus *WeaponBonus) {
		if !seen[bonus] {
			seen[bonus] = true
			result = append(result, bonus)
		}
	}
	tags := w.Owner.TagList()
	usage := w.UsageWithReplacements()
	// Finding the skill the weapon is used with takes several walks of the skills, so it is only done when a bonus
	// that selects weapons by that skill exists to collect.
	if entity.hasWeaponBonusesBySkill() {
		w.forEachRequiredSkill(entity, func(name, specialization string) {
			entity.addWeaponWithSkillBonusesFor(name, specialization, usage, tags, add)
		})
	}
	entity.addNamedWeaponBonusesFor(w.String(), usage, tags, add)
	for _, f := range w.Owner.ActiveFeatures() {
		w.extractWeaponBonus(f, add)
	}
	w.forEachOwnerModifier(func(mod ownerModifier, active Features) {
		for _, f := range active {
			if bonus, ok := f.(Bonus); ok {
				bonus.SetSubOwner(mod)
			}
			w.extractWeaponBonus(f, add)
		}
	})
	return result
}

// forEachRequiredSkill calls visit with the name and specialization of the skill the weapon attacks with, by which the
// features selecting weapons by their required skill are qualified, then with those of each skill in the chain that a
// technique it resolves to is based on. The name is empty when no skill-based default resolves, and must then be
// matched with matchesRequiredSkillName. Only the defaults are ranked and skills looked up here; nothing of the weapon
// itself is resolved, since this runs while its fields are. The skills are found once per weapon per state of the
// entity, from the weapon's memo when it has one.
func (w *Weapon) forEachRequiredSkill(entity *Entity, visit func(name, specialization string)) {
	m := w.memo()
	if m == nil || !m.hasRequired {
		skills := w.findRequiredSkills(entity)
		if m != nil {
			m.requiredSkills = skills
			m.hasRequired = true
		}
		for _, one := range skills {
			visit(one.name, one.specialization)
		}
		return
	}
	for _, one := range m.requiredSkills {
		visit(one.name, one.specialization)
	}
}

// findRequiredSkills is forEachRequiredSkill without the memo, returning the skills instead of visiting them.
func (w *Weapon) findRequiredSkills(entity *Entity) []requiredSkill {
	var skills []requiredSkill
	visit := func(name, specialization string) {
		skills = append(skills, requiredSkill{name: name, specialization: specialization})
	}
	var bestDef *SkillDefault
	var bestSkill *Skill
	best := fxp.Min
	replacements := w.NameableReplacements()
	for _, one := range w.Defaults {
		if one.SkillBased() {
			// Ranked exactly the way SkillLevel() ranks them, so that the skill whose bonuses are collected is the one
			// the weapon actually attacks with. That includes leaving a parry- or block-type default scored by its
			// defense level; see the note in SkillLevel(). The skill each default scores by is kept, since it is the
			// one the winner is resolved to below.
			sk := one.bestFastMatchingSkill(entity, replacements, false, nil)
			if level := one.fastLevelFromSkill(entity, sk); best < level {
				best = level
				bestDef = one
				bestSkill = sk
			}
		}
	}
	var name, specialization string
	if bestDef != nil {
		name = bestDef.NameWithReplacements(replacements)
		specialization = bestDef.SpecializationWithReplacements(replacements)
		// The qualifiers only name a skill when the default names one outright. Resolving the default to the skill it
		// actually scored -- the same one bestFast picked when it won the ranking above -- lets a default that selects
		// its skill by tag, or by any other criteria, still collect the weapon bonuses qualified by the skill the
		// weapon really attacks with.
		if bestSkill != nil {
			name = bestSkill.NameWithReplacements()
			specialization = bestSkill.SpecializationWithReplacements()
		}
	}
	visit(name, specialization)
	// When the weapon's best skill default resolves to a technique, also apply weapon bonuses that are qualified by the
	// skill the technique is based on, evaluated against that base skill's relative level rather than the technique's.
	// Per the rules, certain bonuses are determined by the base skill, not the technique - e.g. a Kick that defaults to
	// the Kicking technique still gets the Brawling/Karate damage bonus based on the Brawling/Karate relative skill
	// level (issue #767). The walk starts from the skill the winning default scored by, which honors its tag and TL
	// criteria.
	sk := bestSkill
	seen := make(map[*Skill]bool)
	for sk != nil && sk.IsTechnique() && !seen[sk] {
		seen[sk] = true
		base := sk.DefaultSkill()
		if base == nil {
			break
		}
		visit(base.NameWithReplacements(), base.SpecializationWithReplacements())
		sk = base
	}
	return skills
}

// matchesRequiredSkillName reports whether a feature's required-skill name criterion matches a skill name visited by
// forEachRequiredSkill. An empty skill means none resolved, which a criterion naming a skill outright cannot match: an
// empty "is" qualifier is the unfilled state a new feature starts in, not a skill name. The other comparisons see the
// empty name, so "is anything" and "is not X" select such a weapon, as they do every other.
func matchesRequiredSkillName(name criteria.Text, replacements map[string]string, skill string) bool {
	if skill == "" && name.Compare == criteria.IsText {
		return false
	}
	return name.Matches(replacements, skill)
}

// extractWeaponBonus hands f to add if it is a weapon bonus, carried by the weapon's owner or by one of the owner's
// modifiers, that selects the owner's weapons or names this one.
func (w *Weapon) extractWeaponBonus(f Feature, add func(*WeaponBonus)) {
	bonus, ok := f.(*WeaponBonus)
	if !ok {
		return
	}
	replacements := w.NameableReplacements()
	switch bonus.SelectionType {
	case wsel.WithRequiredSkill: // Selected by required skill, which only the entity's lists resolve
	case wsel.ThisWeapon:
		if bonus.SpecializationCriteria.Matches(replacements, w.UsageWithReplacements()) {
			add(bonus)
		}
	case wsel.WithName:
		if bonus.NameCriteria.Matches(replacements, w.String()) &&
			bonus.SpecializationCriteria.Matches(replacements, w.UsageWithReplacements()) &&
			bonus.TagsCriteria.MatchesList(replacements, w.Owner.TagList()...) {
			add(bonus)
		}
	default:
		errs.Log(errs.New("unknown selection type"), "type", int(bonus.SelectionType))
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

// ResolveUsageNotes resolves the usage notes, running any embedded scripts to get the final result.
func (w *Weapon) ResolveUsageNotes() string {
	return ResolveText(w.Entity(), deferredNewScriptWeapon(w), w.UsageNotesWithReplacements())
}

// NameableReplacements returns the replacements to be used with this weapon.
func (w *Weapon) NameableReplacements() map[string]string {
	if xreflect.IsNil(w.Owner) {
		return nil
	}
	return w.Owner.NameableReplacements()
}

// FillWithNameableKeys adds any nameable keys found in this Weapon to the provided map.
func (w *Weapon) FillWithNameableKeys(m, existing map[string]string) {
	nameable.Extract(
		m, existing,
		w.Usage,
		w.UsageNotes,
	)
	for _, one := range w.Defaults {
		one.FillWithNameableKeys(m, existing)
	}
}

// ApplyNameableKeys does nothing, since a weapon resolves its nameable keys through its owner.
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
			data.Title = i18n.Text("Melee Attack")
		} else {
			data.Title = i18n.Text("Ranged Attack")
		}
		data.Primary = true
	case WeaponUsageColumn:
		switch {
		case forPage:
			data.Title = i18n.Text("Usage")
		case melee:
			data.Title = i18n.Text("Melee Attack Usage")
		default:
			data.Title = i18n.Text("Ranged Attack Usage")
		}
	case WeaponSLColumn:
		data = abbreviatedHeaderData(i18n.Text("SL"), i18n.Text("Skill Level"))
	case WeaponParryColumn:
		data.Title = i18n.Text("Parry")
	case WeaponBlockColumn:
		data.Title = i18n.Text("Block")
	case WeaponDamageColumn:
		data.Title = i18n.Text("Damage")
	case WeaponReachColumn:
		data.Title = i18n.Text("Reach")
	case WeaponSTColumn:
		data = abbreviatedHeaderData(i18n.Text("ST"), i18n.Text("Minimum Strength"))
	case WeaponAccColumn:
		data = abbreviatedHeaderData(i18n.Text("Acc"), i18n.Text("Accuracy Bonus"))
	case WeaponRangeColumn:
		data.Title = i18n.Text("Range")
	case WeaponRoFColumn:
		data = abbreviatedHeaderData(i18n.Text("RoF"), i18n.Text("Rate of Fire"))
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
	var buffer xbytes.InsertBuffer
	data.Self = w
	data.Type = cell.Text
	switch columnID {
	case WeaponHideColumn:
		data.Type = cell.Toggle
		data.Name = i18n.Text("Hidden")
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
		parry := w.ResolvedParry(&buffer)
		data.Primary = parry.String()
		data.Tooltip = parry.Tooltip()
	case WeaponBlockColumn:
		data.Primary = w.ResolvedBlock(&buffer).String()
	case WeaponDamageColumn:
		data.Primary = w.Damage.ResolvedDamage(&buffer)
	case WeaponReachColumn:
		reach := w.ResolvedReach(&buffer)
		data.Primary = reach.String()
		data.Tooltip = reach.Tooltip()
	case WeaponSTColumn:
		weaponST := w.ResolvedStrength(&buffer)
		data.Primary = weaponST.String()
		data.Tooltip = weaponST.Tooltip(w.ratedStrength())
	case WeaponAccColumn:
		data.Primary = w.ResolvedAccuracy(&buffer).String()
	case WeaponRangeColumn:
		data.Primary = w.ResolvedRange(&buffer).String(w.musclePowerIsResolved())
	case WeaponRoFColumn:
		rof := w.ResolvedRateOfFire(&buffer)
		data.Primary = rof.String()
		data.Tooltip = rof.Tooltip()
	case WeaponShotsColumn:
		shots := w.ResolvedShots(&buffer)
		data.Primary = shots.String()
		data.Tooltip = shots.Tooltip()
	case WeaponBulkColumn:
		bulk := w.ResolvedBulk(&buffer)
		data.Primary = bulk.String()
		data.Tooltip = bulk.Tooltip(w.ResolvedAccuracy(nil), w.ResolvedRecoil(nil), w.ResolvedStrength(nil))
	case WeaponRecoilColumn:
		recoil := w.ResolvedRecoil(&buffer)
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

// ColumnHasData returns true if this weapon has meaningful data to display in the given column. Columns that are not
// eligible to be hidden always report true.
func (w *Weapon) ColumnHasData(columnID int) bool {
	var buffer xbytes.InsertBuffer
	switch columnID {
	case WeaponUsageColumn:
		return w.UsageWithReplacements() != ""
	case WeaponParryColumn:
		return w.ResolvedParry(&buffer).CanParry
	case WeaponBlockColumn:
		return w.ResolvedBlock(&buffer).CanBlock
	case WeaponReachColumn:
		return w.ResolvedReach(&buffer).String() != ""
	case WeaponAccColumn:
		acc := w.ResolvedAccuracy(&buffer)
		return acc.Jet || acc.Base != 0 || acc.Scope != 0
	case WeaponRangeColumn:
		return w.ResolvedRange(&buffer).String(w.musclePowerIsResolved()) != ""
	case WeaponRoFColumn:
		return w.ResolvedRateOfFire(&buffer).String() != ""
	case WeaponShotsColumn:
		return w.ResolvedShots(&buffer).String() != ""
	case WeaponBulkColumn:
		return w.ResolvedBulk(&buffer).String() != ""
	case WeaponRecoilColumn:
		return w.ResolvedRecoil(&buffer).String() != ""
	case WeaponSTColumn:
		return w.ResolvedStrength(&buffer).String() != ""
	default:
		return true
	}
}

// CopyFrom implements EditorData.
func (w *Weapon) CopyFrom(t *Weapon) {
	*w = *t.Clone(LibraryFile{}, t.DataOwner(), nil, Duplicate)
	w.Damage.Owner = w // The clone's damage belongs to the clone, which was only a means of making the copy
}

// ApplyTo implements EditorData.
func (w *Weapon) ApplyTo(t *Weapon) {
	// Unlike the other EditorData implementations, a Weapon serves as its own editor data, so it carries the TID minted
	// for it by CopyFrom. Applying an edit must not disturb the target's identity: the TID is persisted as the "id" and
	// keys table selections, and ClonedFromTID records the weapon a duplicate or reference was cloned from.
	savedTID := t.TID
	savedClonedFromTID := t.ClonedFromTID
	*t = *w.Clone(LibraryFile{}, t.DataOwner(), nil, Copy)
	t.Damage.Owner = t
	t.TID = savedTID
	t.ClonedFromTID = savedClonedFromTID
}

// Validate ensures the weapon data is valid.
func (w *Weapon) Validate() {
	w.Strength.Validate()
	if w.IsMelee() {
		w.Parry.Validate()
		w.Block.Validate()
		w.Reach.Validate()
		w.Accuracy = attack.Accuracy{}
		w.Range = attack.Range{}
		w.RateOfFire = attack.RoF{}
		w.Shots = attack.Shots{}
		w.Bulk = attack.Bulk{}
		w.Recoil = attack.Recoil{}
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
		w.Parry = attack.Parry{}
		w.Block = attack.Block{}
		w.Reach = attack.Reach{}
	}
}
