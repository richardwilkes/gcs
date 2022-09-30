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
	"context"
	"io/fs"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/gid"
	"github.com/richardwilkes/gcs/v5/model/gurps/nameables"
	"github.com/richardwilkes/gcs/v5/model/gurps/trait"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/settings/display"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/slices"
)

var (
	_ WeaponOwner  = &Trait{}
	_ Node[*Trait] = &Trait{}
)

// Columns that can be used with the trait method .CellData()
const (
	TraitDescriptionColumn = iota
	TraitPointsColumn
	TraitTagsColumn
	TraitReferenceColumn
)

const (
	traitListTypeKey = "trait_list"
	traitTypeKey     = "trait"
)

// Trait holds an advantage, disadvantage, quirk, or perk.
type Trait struct {
	TraitData
	Entity            *Entity
	UnsatisfiedReason string
}

type traitListData struct {
	Type    string   `json:"type"`
	Version int      `json:"version"`
	Rows    []*Trait `json:"rows"`
}

// NewTraitsFromFile loads an Trait list from a file.
func NewTraitsFromFile(fileSystem fs.FS, filePath string) ([]*Trait, error) {
	var data traitListData
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &data); err != nil {
		return nil, errs.NewWithCause(gid.InvalidFileDataMsg, err)
	}
	if data.Type == "advantage_list" {
		data.Type = traitListTypeKey
	}
	if data.Type != traitListTypeKey {
		return nil, errs.New(gid.UnexpectedFileDataMsg)
	}
	if err := gid.CheckVersion(data.Version); err != nil {
		return nil, err
	}
	return data.Rows, nil
}

// SaveTraits writes the Trait list to the file as JSON.
func SaveTraits(traits []*Trait, filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, &traitListData{
		Type:    traitListTypeKey,
		Version: gid.CurrentDataVersion,
		Rows:    traits,
	})
}

// NewTrait creates a new Trait.
func NewTrait(entity *Entity, parent *Trait, container bool) *Trait {
	a := &Trait{
		TraitData: TraitData{
			ContainerBase: newContainerBase[*Trait](traitTypeKey, container),
		},
		Entity: entity,
	}
	a.Name = a.Kind()
	a.parent = parent
	return a
}

// Clone implements Node.
func (a *Trait) Clone(entity *Entity, parent *Trait, preserveID bool) *Trait {
	other := NewTrait(entity, parent, a.Container())
	if preserveID {
		other.ID = a.ID
	}
	other.IsOpen = a.IsOpen
	other.TraitEditData.CopyFrom(a)
	if a.HasChildren() {
		other.Children = make([]*Trait, 0, len(a.Children))
		for _, child := range a.Children {
			other.Children = append(other.Children, child.Clone(entity, other, preserveID))
		}
	}
	return other
}

// MarshalJSON implements json.Marshaler.
func (a *Trait) MarshalJSON() ([]byte, error) {
	type calc struct {
		Points fxp.Int `json:"points"`
	}
	a.ClearUnusedFieldsForType()
	data := struct {
		TraitData
		Calc calc `json:"calc"`
	}{
		TraitData: a.TraitData,
		Calc: calc{
			Points: a.AdjustedPoints(),
		},
	}
	return json.Marshal(&data)
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *Trait) UnmarshalJSON(data []byte) error {
	var localData struct {
		TraitData
		// Old data fields
		Categories   []string `json:"categories"`
		Mental       bool     `json:"mental"`
		Physical     bool     `json:"physical"`
		Social       bool     `json:"social"`
		Exotic       bool     `json:"exotic"`
		Supernatural bool     `json:"supernatural"`
	}
	if err := json.Unmarshal(data, &localData); err != nil {
		return err
	}

	// Swap out old type keys
	switch localData.Type {
	case "advantage":
		localData.Type = traitTypeKey
	case "advantage_container":
		localData.Type = traitTypeKey + ContainerKeyPostfix
	}

	localData.ClearUnusedFieldsForType()
	a.TraitData = localData.TraitData
	a.transferOldTypeFlagToTags(i18n.Text("Mental"), localData.Mental)
	a.transferOldTypeFlagToTags(i18n.Text("Physical"), localData.Physical)
	a.transferOldTypeFlagToTags(i18n.Text("Social"), localData.Social)
	a.transferOldTypeFlagToTags(i18n.Text("Exotic"), localData.Exotic)
	a.transferOldTypeFlagToTags(i18n.Text("Supernatural"), localData.Supernatural)
	a.Tags = convertOldCategoriesToTags(a.Tags, localData.Categories)
	slices.Sort(a.Tags)
	if a.Container() {
		for _, one := range a.Children {
			one.parent = a
		}
	}
	return nil
}

func (a *Trait) transferOldTypeFlagToTags(name string, flag bool) {
	if flag && !slices.Contains(a.Tags, name) {
		a.Tags = append(a.Tags, name)
	}
}

// EffectivelyDisabled returns true if this node or a parent is disabled.
func (a *Trait) EffectivelyDisabled() bool {
	if a.Disabled {
		return true
	}
	p := a.Parent()
	for p != nil {
		if p.Disabled {
			return true
		}
		p = p.Parent()
	}
	return false
}

// CellData returns the cell data information for the given column.
func (a *Trait) CellData(column int, data *CellData) {
	data.Dim = !a.Enabled()
	switch column {
	case TraitDescriptionColumn:
		data.Type = Text
		data.Primary = a.String()
		data.Secondary = a.SecondaryText(func(option display.Option) bool { return option.Inline() })
		data.Disabled = a.EffectivelyDisabled()
		data.UnsatisfiedReason = a.UnsatisfiedReason
		data.Tooltip = a.SecondaryText(func(option display.Option) bool { return option.Tooltip() })
	case TraitPointsColumn:
		data.Type = Text
		data.Primary = a.AdjustedPoints().String()
		data.Alignment = unison.EndAlignment
	case TraitTagsColumn:
		data.Type = Text
		data.Primary = CombineTags(a.Tags)
	case TraitReferenceColumn, PageRefCellAlias:
		data.Type = PageRef
		data.Primary = a.PageRef
		data.Secondary = a.Name
	}
}

// Depth returns the number of parents this node has.
func (a *Trait) Depth() int {
	count := 0
	p := a.parent
	for p != nil {
		count++
		p = p.parent
	}
	return count
}

// OwningEntity returns the owning Entity.
func (a *Trait) OwningEntity() *Entity {
	return a.Entity
}

// SetOwningEntity sets the owning entity and configures any sub-components as needed.
func (a *Trait) SetOwningEntity(entity *Entity) {
	a.Entity = entity
	if a.Container() {
		for _, child := range a.Children {
			child.SetOwningEntity(entity)
		}
	} else {
		for _, w := range a.Weapons {
			w.SetOwner(a)
		}
	}
	for _, m := range a.Modifiers {
		m.SetOwningEntity(entity)
	}
}

// Notes returns the local notes.
func (a *Trait) Notes() string {
	return a.LocalNotes
}

// IsLeveled returns true if the Trait is capable of having levels.
func (a *Trait) IsLeveled() bool {
	return !a.Container() && a.PointsPerLevel != 0
}

// AdjustedPoints returns the total points, taking levels and modifiers into account.
func (a *Trait) AdjustedPoints() fxp.Int {
	if a.EffectivelyDisabled() {
		return 0
	}
	if !a.Container() {
		return AdjustedPoints(a.Entity, a.BasePoints, a.Levels, a.PointsPerLevel, a.CR, a.AllModifiers(), a.RoundCostDown)
	}
	var points fxp.Int
	if a.ContainerType == trait.AlternativeAbilities {
		values := make([]fxp.Int, len(a.Children))
		for i, one := range a.Children {
			values[i] = one.AdjustedPoints()
			if values[i] > points {
				points = values[i]
			}
		}
		max := points
		found := false
		for _, v := range values {
			if !found && max == v {
				found = true
			} else {
				points += fxp.ApplyRounding(calculateModifierPoints(v, fxp.Twenty), a.RoundCostDown)
			}
		}
	} else {
		for _, one := range a.Children {
			points += one.AdjustedPoints()
		}
	}
	return points
}

// AllModifiers returns the modifiers plus any inherited from parents.
func (a *Trait) AllModifiers() []*TraitModifier {
	all := make([]*TraitModifier, len(a.Modifiers))
	copy(all, a.Modifiers)
	p := a.parent
	for p != nil {
		all = append(all, p.Modifiers...)
		p = p.parent
	}
	return all
}

// Enabled returns true if this Trait and all of its parents are enabled.
func (a *Trait) Enabled() bool {
	if a.Disabled {
		return false
	}
	p := a.parent
	for p != nil {
		if p.Disabled {
			return false
		}
		p = p.parent
	}
	return true
}

// Description returns a description, which doesn't include any levels.
func (a *Trait) Description() string {
	return a.Name
}

// String implements fmt.Stringer.
func (a *Trait) String() string {
	var buffer strings.Builder
	buffer.WriteString(a.Name)
	if a.IsLeveled() {
		buffer.WriteByte(' ')
		buffer.WriteString(a.Levels.String())
	}
	return buffer.String()
}

// FeatureList returns the list of Features.
func (a *Trait) FeatureList() feature.Features {
	return a.Features
}

// TagList returns the list of tags.
func (a *Trait) TagList() []string {
	return a.Tags
}

// FillWithNameableKeys adds any nameable keys found to the provided map.
func (a *Trait) FillWithNameableKeys(m map[string]string) {
	nameables.Extract(a.Name, m)
	nameables.Extract(a.LocalNotes, m)
	nameables.Extract(a.VTTNotes, m)
	nameables.Extract(a.UserDesc, m)
	if a.Prereq != nil {
		a.Prereq.FillWithNameableKeys(m)
	}
	for _, one := range a.Features {
		one.FillWithNameableKeys(m)
	}
	for _, one := range a.Weapons {
		one.FillWithNameableKeys(m)
	}
	Traverse(func(mod *TraitModifier) bool {
		mod.FillWithNameableKeys(m)
		return false
	}, true, true, a.Modifiers...)
}

// ApplyNameableKeys replaces any nameable keys found with the corresponding values in the provided map.
func (a *Trait) ApplyNameableKeys(m map[string]string) {
	a.Name = nameables.Apply(a.Name, m)
	a.LocalNotes = nameables.Apply(a.LocalNotes, m)
	a.VTTNotes = nameables.Apply(a.VTTNotes, m)
	a.UserDesc = nameables.Apply(a.UserDesc, m)
	if a.Prereq != nil {
		a.Prereq.ApplyNameableKeys(m)
	}
	for _, one := range a.Features {
		one.ApplyNameableKeys(m)
	}
	for _, one := range a.Weapons {
		one.ApplyNameableKeys(m)
	}
	Traverse(func(mod *TraitModifier) bool {
		mod.ApplyNameableKeys(m)
		return false
	}, true, true, a.Modifiers...)
}

// ActiveModifierFor returns the first modifier that matches the name (case-insensitive).
func (a *Trait) ActiveModifierFor(name string) *TraitModifier {
	var found *TraitModifier
	Traverse(func(mod *TraitModifier) bool {
		if strings.EqualFold(mod.Name, name) {
			found = mod
			return true
		}
		return false
	}, true, true, a.Modifiers...)
	return found
}

// ModifierNotes returns the notes due to modifiers.
func (a *Trait) ModifierNotes() string {
	var buffer strings.Builder
	if a.CR != trait.None {
		buffer.WriteString(a.CR.String())
		if a.CRAdj != NoCRAdj {
			buffer.WriteString(", ")
			buffer.WriteString(a.CRAdj.Description(a.CR))
		}
	}
	Traverse(func(mod *TraitModifier) bool {
		if buffer.Len() != 0 {
			buffer.WriteString("; ")
		}
		buffer.WriteString(mod.FullDescription())
		return false
	}, true, true, a.Modifiers...)
	return buffer.String()
}

// SecondaryText returns the "secondary" text: the text display below an Trait.
func (a *Trait) SecondaryText(optionChecker func(display.Option) bool) string {
	var buffer strings.Builder
	settings := SheetSettingsFor(a.Entity)
	if a.UserDesc != "" && optionChecker(settings.UserDescriptionDisplay) {
		buffer.WriteString(a.UserDesc)
	}
	if optionChecker(settings.ModifiersDisplay) {
		if notes := a.ModifierNotes(); notes != "" {
			if buffer.Len() != 0 {
				buffer.WriteByte('\n')
			}
			buffer.WriteString(notes)
		}
	}
	if a.LocalNotes != "" && optionChecker(settings.NotesDisplay) {
		if buffer.Len() != 0 {
			buffer.WriteByte('\n')
		}
		buffer.WriteString(a.LocalNotes)
	}
	return buffer.String()
}

// HasTag returns true if 'tag' is present in 'tags'. This check both ignores case and can check for subsets that are
// colon-separated.
func HasTag(tag string, tags []string) bool {
	tag = strings.TrimSpace(tag)
	for _, one := range tags {
		for _, part := range strings.Split(one, ":") {
			if strings.EqualFold(tag, strings.TrimSpace(part)) {
				return true
			}
		}
	}
	return false
}

// CombineTags combines multiple tags into a single string.
func CombineTags(tags []string) string {
	return strings.Join(tags, ", ")
}

// ExtractTags from a combined tags string.
func ExtractTags(tags string) []string {
	var list []string
	for _, one := range strings.Split(tags, ",") {
		if one = strings.TrimSpace(one); one != "" {
			list = append(list, one)
		}
	}
	return list
}

// AdjustedPoints returns the total points, taking levels and modifiers into account. 'entity' may be nil.
func AdjustedPoints(entity *Entity, basePoints, levels, pointsPerLevel fxp.Int, cr trait.SelfControlRoll, modifiers []*TraitModifier, roundCostDown bool) fxp.Int {
	var baseEnh, levelEnh, baseLim, levelLim fxp.Int
	multiplier := cr.Multiplier()
	Traverse(func(mod *TraitModifier) bool {
		modifier := mod.CostModifier()
		switch mod.CostType {
		case trait.Percentage:
			switch mod.Affects {
			case trait.Total:
				if modifier < 0 {
					baseLim += modifier
					levelLim += modifier
				} else {
					baseEnh += modifier
					levelEnh += modifier
				}
			case trait.BaseOnly:
				if modifier < 0 {
					baseLim += modifier
				} else {
					baseEnh += modifier
				}
			case trait.LevelsOnly:
				if modifier < 0 {
					levelLim += modifier
				} else {
					levelEnh += modifier
				}
			}
		case trait.Points:
			if mod.Affects == trait.LevelsOnly {
				pointsPerLevel += modifier
			} else {
				basePoints += modifier
			}
		case trait.Multiplier:
			multiplier = multiplier.Mul(modifier)
		}
		return false
	}, true, true, modifiers...)
	modifiedBasePoints := basePoints
	leveledPoints := pointsPerLevel.Mul(levels)
	if baseEnh != 0 || baseLim != 0 || levelEnh != 0 || levelLim != 0 {
		if SheetSettingsFor(entity).UseMultiplicativeModifiers {
			if baseEnh == levelEnh && baseLim == levelLim {
				modifiedBasePoints = modifyPoints(modifyPoints(modifiedBasePoints+leveledPoints, baseEnh), (-fxp.Eighty).Max(baseLim))
			} else {
				modifiedBasePoints = modifyPoints(modifyPoints(modifiedBasePoints, baseEnh), (-fxp.Eighty).Max(baseLim)) +
					modifyPoints(modifyPoints(leveledPoints, levelEnh), (-fxp.Eighty).Max(levelLim))
			}
		} else {
			baseMod := (-fxp.Eighty).Max(baseEnh + baseLim)
			levelMod := (-fxp.Eighty).Max(levelEnh + levelLim)
			if baseMod == levelMod {
				modifiedBasePoints = modifyPoints(modifiedBasePoints+leveledPoints, baseMod)
			} else {
				modifiedBasePoints = modifyPoints(modifiedBasePoints, baseMod) + modifyPoints(leveledPoints, levelMod)
			}
		}
	} else {
		modifiedBasePoints += leveledPoints
	}
	return fxp.ApplyRounding(modifiedBasePoints.Mul(multiplier), roundCostDown)
}

func modifyPoints(points, modifier fxp.Int) fxp.Int {
	return points + calculateModifierPoints(points, modifier)
}

func calculateModifierPoints(points, modifier fxp.Int) fxp.Int {
	return points.Mul(modifier).Div(fxp.Hundred)
}
