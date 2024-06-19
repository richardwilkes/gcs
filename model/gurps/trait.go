// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"context"
	"io/fs"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/affects"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/study"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/tmcost"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison/enums/align"
)

var (
	_ WeaponOwner            = &Trait{}
	_ Node[*Trait]           = &Trait{}
	_ TemplatePickerProvider = &Trait{}
	_ LeveledOwner           = &Trait{}
	_ EditorData[*Trait]     = &TraitEditData{}
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

// TraitData holds the Trait data that is written to disk.
type TraitData struct {
	ContainerBase[*Trait]
	TraitEditData
}

// TraitEditData holds the Trait data that can be edited by the UI detail editor.
type TraitEditData struct {
	Name             string              `json:"name,omitempty"`
	PageRef          string              `json:"reference,omitempty"`
	PageRefHighlight string              `json:"reference_highlight,omitempty"`
	LocalNotes       string              `json:"notes,omitempty"`
	VTTNotes         string              `json:"vtt_notes,omitempty"`
	UserDesc         string              `json:"userdesc,omitempty"`
	Tags             []string            `json:"tags,omitempty"`
	Modifiers        []*TraitModifier    `json:"modifiers,omitempty"`
	CR               selfctrl.Roll       `json:"cr,omitempty"`
	CRAdj            selfctrl.Adjustment `json:"cr_adj,omitempty"`
	Disabled         bool                `json:"disabled,omitempty"`
	TraitNonContainerOnlyEditData
	TraitContainerOnlyEditData
}

// TraitNonContainerOnlyEditData holds the Trait data that is only applicable to traits that aren't containers.
type TraitNonContainerOnlyEditData struct {
	BasePoints       fxp.Int     `json:"base_points,omitempty"`
	Levels           fxp.Int     `json:"levels,omitempty"`
	PointsPerLevel   fxp.Int     `json:"points_per_level,omitempty"`
	Prereq           *PrereqList `json:"prereqs,omitempty"`
	Weapons          []*Weapon   `json:"weapons,omitempty"`
	Features         Features    `json:"features,omitempty"`
	Study            []*Study    `json:"study,omitempty"`
	StudyHoursNeeded study.Level `json:"study_hours_needed,omitempty"`
	RoundCostDown    bool        `json:"round_down,omitempty"`
	CanLevel         bool        `json:"can_level,omitempty"`
}

// TraitContainerOnlyEditData holds the Trait data that is only applicable to traits that are containers.
type TraitContainerOnlyEditData struct {
	Ancestry       string          `json:"ancestry,omitempty"`
	TemplatePicker *TemplatePicker `json:"template_picker,omitempty"`
	ContainerType  container.Type  `json:"container_type,omitempty"`
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
		return nil, errs.NewWithCause(InvalidFileDataMsg(), err)
	}
	if data.Type == "advantage_list" {
		data.Type = traitListTypeKey
	}
	if data.Type != traitListTypeKey {
		return nil, errs.New(UnexpectedFileDataMsg())
	}
	if err := CheckVersion(data.Version); err != nil {
		return nil, err
	}
	return data.Rows, nil
}

// SaveTraits writes the Trait list to the file as JSON.
func SaveTraits(traits []*Trait, filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, &traitListData{
		Type:    traitListTypeKey,
		Version: CurrentDataVersion,
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
	if a.Container() {
		a.TemplatePicker = &TemplatePicker{}
	}
	return a
}

// Clone implements Node.
func (a *Trait) Clone(entity *Entity, parent *Trait, preserveID bool) *Trait {
	other := NewTrait(entity, parent, a.Container())
	if preserveID {
		other.ID = a.ID
	}
	other.IsOpen = a.IsOpen
	other.ThirdParty = a.ThirdParty
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
		Points            fxp.Int `json:"points"`
		UnsatisfiedReason string  `json:"unsatisfied_reason,omitempty"`
		ResolvedNotes     string  `json:"resolved_notes,omitempty"`
	}
	a.ClearUnusedFieldsForType()
	data := struct {
		TraitData
		Calc calc `json:"calc"`
	}{
		TraitData: a.TraitData,
		Calc: calc{
			Points:            a.AdjustedPoints(),
			UnsatisfiedReason: a.UnsatisfiedReason,
		},
	}
	notes := a.resolveLocalNotes()
	if notes != a.LocalNotes {
		data.Calc.ResolvedNotes = notes
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

	// Force the CanLevel flag, if needed
	if !localData.Container() && (localData.Levels != 0 || localData.PointsPerLevel != 0) {
		localData.CanLevel = true
	}

	localData.ClearUnusedFieldsForType()
	a.TraitData = localData.TraitData
	a.transferOldTypeFlagToTags(i18n.Text("Mental"), localData.Mental)
	a.transferOldTypeFlagToTags(i18n.Text("Physical"), localData.Physical)
	a.transferOldTypeFlagToTags(i18n.Text("Social"), localData.Social)
	a.transferOldTypeFlagToTags(i18n.Text("Exotic"), localData.Exotic)
	a.transferOldTypeFlagToTags(i18n.Text("Supernatural"), localData.Supernatural)
	a.Tags = ConvertOldCategoriesToTags(a.Tags, localData.Categories)
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

// TemplatePickerData returns the TemplatePicker data, if any.
func (a *Trait) TemplatePickerData() *TemplatePicker {
	return a.TemplatePicker
}

// TraitsHeaderData returns the header data information for the given trait column.
func TraitsHeaderData(columnID int) HeaderData {
	var data HeaderData
	switch columnID {
	case TraitDescriptionColumn:
		data.Title = i18n.Text("Trait")
		data.Primary = true
	case TraitPointsColumn:
		data.Title = i18n.Text("Pts")
		data.Detail = i18n.Text("Points")
	case TraitTagsColumn:
		data.Title = i18n.Text("Tags")
	case TraitReferenceColumn:
		data.Title = HeaderBookmark
		data.TitleIsImageKey = true
		data.Detail = PageRefTooltipText()
	}
	return data
}

// CellData returns the cell data information for the given column.
func (a *Trait) CellData(columnID int, data *CellData) {
	data.Dim = !a.Enabled()
	switch columnID {
	case TraitDescriptionColumn:
		data.Type = cell.Text
		data.Primary = a.String()
		data.Secondary = a.SecondaryText(func(option display.Option) bool { return option.Inline() })
		data.Disabled = a.EffectivelyDisabled()
		data.UnsatisfiedReason = a.UnsatisfiedReason
		data.Tooltip = a.SecondaryText(func(option display.Option) bool { return option.Tooltip() })
		data.TemplateInfo = a.TemplatePicker.Description()
		if a.Container() {
			switch a.ContainerType {
			case container.AlternativeAbilities:
				data.InlineTag = i18n.Text("Alternate")
			case container.Ancestry:
				data.InlineTag = i18n.Text("Ancestry")
			case container.Attributes:
				data.InlineTag = i18n.Text("Attribute")
			case container.MetaTrait:
				data.InlineTag = i18n.Text("Meta")
			default:
			}
		}
	case TraitPointsColumn:
		data.Type = cell.Text
		data.Primary = a.AdjustedPoints().String()
		data.Alignment = align.End
	case TraitTagsColumn:
		data.Type = cell.Tags
		data.Primary = CombineTags(a.Tags)
	case TraitReferenceColumn, PageRefCellAlias:
		data.Type = cell.PageRef
		data.Primary = a.PageRef
		if a.PageRefHighlight != "" {
			data.Secondary = a.PageRefHighlight
		} else {
			data.Secondary = a.Name
		}
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
	return a.resolveLocalNotes()
}

// IsLeveled returns true if the Trait is capable of having levels.
func (a *Trait) IsLeveled() bool {
	return a.CanLevel && !a.Container()
}

// CurrentLevel returns the current level of the trait or zero if it is not leveled.
func (a *Trait) CurrentLevel() fxp.Int {
	if a.Enabled() && a.IsLeveled() {
		return a.Levels
	}
	return 0
}

// AdjustedPoints returns the total points, taking levels and modifiers into account.
func (a *Trait) AdjustedPoints() fxp.Int {
	if a.EffectivelyDisabled() {
		return 0
	}
	if !a.Container() {
		return AdjustedPoints(a.Entity, a.CanLevel, a.BasePoints, a.Levels, a.PointsPerLevel, a.CR, a.AllModifiers(), a.RoundCostDown)
	}
	var points fxp.Int
	if a.ContainerType == container.AlternativeAbilities {
		values := make([]fxp.Int, len(a.Children))
		for i, one := range a.Children {
			values[i] = one.AdjustedPoints()
			if values[i] > points {
				points = values[i]
			}
		}
		maximum := points
		found := false
		for _, v := range values {
			if !found && maximum == v {
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

func (a *Trait) resolveLocalNotes() string {
	return EvalEmbeddedRegex.ReplaceAllStringFunc(a.LocalNotes, a.Entity.EmbeddedEval)
}

// FeatureList returns the list of Features.
func (a *Trait) FeatureList() Features {
	return a.Features
}

// TagList returns the list of tags.
func (a *Trait) TagList() []string {
	return a.Tags
}

// RatedStrength always return 0 for traits.
func (a *Trait) RatedStrength() fxp.Int {
	return 0
}

// FillWithNameableKeys adds any nameable keys found to the provided map.
func (a *Trait) FillWithNameableKeys(m map[string]string) {
	Extract(a.Name, m)
	Extract(a.LocalNotes, m)
	Extract(a.UserDesc, m)
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
	a.Name = Apply(a.Name, m)
	a.LocalNotes = Apply(a.LocalNotes, m)
	a.UserDesc = Apply(a.UserDesc, m)
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
	if a.CR != selfctrl.NoCR {
		buffer.WriteString(a.CR.String())
		if a.CRAdj != selfctrl.NoCRAdj {
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
		AppendStringOntoNewLine(&buffer, a.ModifierNotes())
	}
	if optionChecker(settings.NotesDisplay) {
		AppendStringOntoNewLine(&buffer, strings.TrimSpace(a.Notes()))
		AppendStringOntoNewLine(&buffer, StudyHoursProgressText(ResolveStudyHours(a.Study), a.StudyHoursNeeded, false))
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
func AdjustedPoints(entity *Entity, canLevel bool, basePoints, levels, pointsPerLevel fxp.Int, cr selfctrl.Roll, modifiers []*TraitModifier, roundCostDown bool) fxp.Int {
	if !canLevel {
		levels = 0
		pointsPerLevel = 0
	}
	var baseEnh, levelEnh, baseLim, levelLim fxp.Int
	multiplier := cr.Multiplier()
	Traverse(func(mod *TraitModifier) bool {
		modifier := mod.CostModifier()
		switch mod.CostType {
		case tmcost.Percentage:
			switch mod.Affects {
			case affects.Total:
				if modifier < 0 {
					baseLim += modifier
					levelLim += modifier
				} else {
					baseEnh += modifier
					levelEnh += modifier
				}
			case affects.BaseOnly:
				if modifier < 0 {
					baseLim += modifier
				} else {
					baseEnh += modifier
				}
			case affects.LevelsOnly:
				if modifier < 0 {
					levelLim += modifier
				} else {
					levelEnh += modifier
				}
			}
		case tmcost.Points:
			if mod.Affects == affects.LevelsOnly {
				if canLevel {
					pointsPerLevel += modifier
				}
			} else {
				basePoints += modifier
			}
		case tmcost.Multiplier:
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

// Kind returns the kind of data.
func (d *TraitData) Kind() string {
	return d.kind(i18n.Text("Trait"))
}

// ClearUnusedFieldsForType zeroes out the fields that are not applicable to this type (container vs not-container).
func (d *TraitData) ClearUnusedFieldsForType() {
	d.clearUnusedFields()
	if d.Container() {
		d.BasePoints = 0
		d.Levels = 0
		d.PointsPerLevel = 0
		d.CanLevel = false
		d.Prereq = nil
		d.Weapons = nil
		d.Features = nil
		d.RoundCostDown = false
		d.StudyHoursNeeded = study.Standard
		if d.TemplatePicker == nil {
			d.TemplatePicker = &TemplatePicker{}
		}
	} else {
		d.ContainerType = 0
		d.TemplatePicker = nil
		d.Ancestry = ""
		if !d.CanLevel {
			d.Levels = 0
			d.PointsPerLevel = 0
		}
	}
}

// CopyFrom implements node.EditorData.
func (d *TraitEditData) CopyFrom(t *Trait) {
	d.copyFrom(t.Entity, &t.TraitEditData, t.Container(), false)
}

// ApplyTo implements node.EditorData.
func (d *TraitEditData) ApplyTo(t *Trait) {
	t.TraitEditData.copyFrom(t.Entity, d, t.Container(), true)
}

func (d *TraitEditData) copyFrom(entity *Entity, other *TraitEditData, isContainer, isApply bool) {
	*d = *other
	d.Tags = txt.CloneStringSlice(d.Tags)
	d.Modifiers = nil
	if len(other.Modifiers) != 0 {
		d.Modifiers = make([]*TraitModifier, 0, len(other.Modifiers))
		for _, one := range other.Modifiers {
			d.Modifiers = append(d.Modifiers, one.Clone(entity, nil, true))
		}
	}
	d.Prereq = d.Prereq.CloneResolvingEmpty(isContainer, isApply)
	d.Weapons = nil
	if len(other.Weapons) != 0 {
		d.Weapons = make([]*Weapon, len(other.Weapons))
		for i := range other.Weapons {
			d.Weapons[i] = other.Weapons[i].Clone(entity, nil, true)
		}
	}
	d.Features = other.Features.Clone()
	if len(other.Study) != 0 {
		d.Study = make([]*Study, len(other.Study))
		for i := range other.Study {
			d.Study[i] = other.Study[i].Clone()
		}
	}
	d.TemplatePicker = d.TemplatePicker.Clone()
}
