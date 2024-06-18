package gurps

import (
	"context"
	"io/fs"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/tid"
	"github.com/richardwilkes/toolbox/txt"
)

const equipmentListTypeKey = "equipment_list"

// Equipment holds a piece of equipment.
type Equipment struct {
	EquipmentData
	Entity            *gurps.Entity
	UnsatisfiedReason string
	Parent            *Equipment
}

// EquipmentData holds the equipment data that is written to disk.
type EquipmentData struct {
	LocalID      tid.TID `json:"local_id"`
	DataSource   string  `json:"data_source,omitempty"`
	DataSourceID tid.TID `json:"data_source_id,omitempty"`
	EquipmentBase
	EquipmentMutable
}

// EquipmentBase holds the base data for equipment, which consists of fields that are not mutable unless a new piece of
// equipment is created.
type EquipmentBase struct {
	Name                   string            `json:"name,alt=description,omitempty"`
	Tags                   []string          `json:"tags,omitempty"`
	PageRef                string            `json:"reference,omitempty"`
	PageRefHighlight       string            `json:"reference_highlight,omitempty"`
	Notes                  string            `json:"notes,omitempty"`
	VTTNotes               string            `json:"vtt_notes,omitempty"`
	TechLevel              string            `json:"tech_level,omitempty"`
	LegalityClass          string            `json:"legality_class,omitempty"`
	Value                  fxp.Int           `json:"value,omitempty"`
	Weight                 fxp.Weight        `json:"weight,omitempty"`
	RatedST                fxp.Int           `json:"rated_strength,omitempty"`
	MaxUses                int               `json:"max_uses,omitempty"`
	Prereq                 *gurps.PrereqList `json:"prereqs,omitempty"`
	Features               gurps.Features    `json:"features,omitempty"`
	Weapons                []*gurps.Weapon   `json:"weapons,omitempty"`
	WeightIgnoredForSkills bool              `json:"ignore_weight_for_skills,omitempty"`
}

// EquipmentMutable holds the mutable data for equipment.
type EquipmentMutable struct {
	Modifiers  []*gurps.EquipmentModifier `json:"modifiers,omitempty"`
	Quantity   fxp.Int                    `json:"quantity,omitempty"`
	Uses       int                        `json:"uses,omitempty"`
	Equipped   bool                       `json:"equipped,omitempty"`
	ThirdParty map[string]any             `json:"third_party,omitempty"`
	EquipmentMutableContainerOnly
}

// EquipmentMutableContainerOnly holds the mutable data for equipment that is only applicable to containers.
type EquipmentMutableContainerOnly struct {
	Children []*Equipment `json:"children,omitempty"`
}

type equipmentListData struct {
	Type    string       `json:"type"`
	Version int          `json:"version"`
	Rows    []*Equipment `json:"rows"`
}

// NewEquipmentFromFile loads an Equipment list from a file.
func NewEquipmentFromFile(fileSystem fs.FS, filePath string) ([]*Equipment, error) {
	var data equipmentListData
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &data); err != nil {
		return nil, errs.NewWithCause(gurps.InvalidFileDataMsg(), err)
	}
	if data.Type != equipmentListTypeKey {
		return nil, errs.New(gurps.UnexpectedFileDataMsg())
	}
	if err := gurps.CheckVersion(data.Version); err != nil {
		return nil, err
	}
	return data.Rows, nil
}

// SaveEquipment writes the Equipment list to the file as JSON.
func SaveEquipment(equipment []*Equipment, filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, &equipmentListData{
		Type:    equipmentListTypeKey,
		Version: gurps.CurrentDataVersion,
		Rows:    equipment,
	})
}

// NewEquipment creates a new Equipment.
func NewEquipment(entity *gurps.Entity, parent *Equipment, container bool) *Equipment {
	var kind byte
	if container {
		kind = 'E'
	} else {
		kind = 'e'
	}
	e := Equipment{
		EquipmentData: EquipmentData{
			LocalID: tid.MustNewTID(kind),
			EquipmentBase: EquipmentBase{
				LegalityClass: "4",
			},
			EquipmentMutable: EquipmentMutable{
				Quantity: fxp.One,
				Equipped: true,
			},
		},
		Entity: entity,
		Parent: parent,
	}
	e.Name = e.Kind()
	return &e
}

// Kind returns the kind of data.
func (e *Equipment) Kind() string {
	if e.Container() {
		return i18n.Text("Equipment Container")
	}
	return i18n.Text("Equipment")
}

// GetParent returns the parent.
func (e Equipment) GetParent() *Equipment {
	return e.Parent
}

// SetParent sets the parent.
func (e *Equipment) SetParent(parent *Equipment) {
	e.Parent = parent
}

// Container returns true if this is a container.
func (e Equipment) Container() bool {
	return tid.IsKindAndValid(e.LocalID, 'E')
}

// HasChildren returns true if this node has children.
func (e Equipment) HasChildren() bool {
	return e.Container() && len(e.Children) > 0
}

// NodeChildren returns the children of this node, if any.
func (e Equipment) NodeChildren() []*Equipment {
	return e.Children
}

// SetChildren sets the children of this node.
func (e *Equipment) SetChildren(children []*Equipment) {
	e.Children = children
}

// AdjustedValue returns the value after adjustments for any modifiers. Does not include the value of children.
func (e *Equipment) AdjustedValue() fxp.Int {
	return gurps.ValueAdjustedForModifiers(e.Value, e.Modifiers)
}

// ExtendedValue returns the extended value.
func (e *Equipment) ExtendedValue() fxp.Int {
	if e.Quantity <= 0 {
		return 0
	}
	value := e.AdjustedValue()
	if e.Container() {
		for _, one := range e.Children {
			value += one.ExtendedValue()
		}
	}
	return value.Mul(e.Quantity)
}

// ExtendedWeight returns the extended weight.
func (e *Equipment) ExtendedWeight(forSkills bool, defUnits fxp.WeightUnit) fxp.Weight {
	return ExtendedWeightAdjustedForModifiers(defUnits, e.Quantity, e.Weight, e.Modifiers, e.Features, e.Children, forSkills, e.WeightIgnoredForSkills && e.Equipped)
}

// ExtendedWeightAdjustedForModifiers calculates the extended weight.
func ExtendedWeightAdjustedForModifiers(defUnits fxp.WeightUnit, qty fxp.Int, baseWeight fxp.Weight, modifiers []*gurps.EquipmentModifier, features gurps.Features, children []*Equipment, forSkills, weightIgnoredForSkills bool) fxp.Weight {
	if qty <= 0 {
		return 0
	}
	var base fxp.Int
	if !forSkills || !weightIgnoredForSkills {
		base = fxp.Int(gurps.WeightAdjustedForModifiers(baseWeight, modifiers, defUnits))
	}
	if len(children) != 0 {
		var contained fxp.Int
		for _, one := range children {
			contained += fxp.Int(one.ExtendedWeight(forSkills, defUnits))
		}
		var percentage, reduction fxp.Int
		for _, one := range features {
			if cwr, ok := one.(*gurps.ContainedWeightReduction); ok {
				if cwr.IsPercentageReduction() {
					percentage += cwr.PercentageReduction()
				} else {
					reduction += fxp.Int(cwr.FixedReduction(defUnits))
				}
			}
		}
		gurps.Traverse(func(mod *gurps.EquipmentModifier) bool {
			for _, f := range mod.Features {
				if cwr, ok := f.(*gurps.ContainedWeightReduction); ok {
					if cwr.IsPercentageReduction() {
						percentage += cwr.PercentageReduction()
					} else {
						reduction += fxp.Int(cwr.FixedReduction(defUnits))
					}
				}
			}
			return false
		}, true, true, modifiers...)
		if percentage >= fxp.Hundred {
			contained = 0
		} else if percentage > 0 {
			contained -= contained.Mul(percentage).Div(fxp.Hundred)
		}
		base += (contained - reduction).Max(0)
	}
	return fxp.Weight(base.Mul(qty))
}

func (e *Equipment) resolveLocalNotes() string {
	return gurps.EvalEmbeddedRegex.ReplaceAllStringFunc(e.Notes, e.Entity.EmbeddedEval)
}

// ClearUnusedFieldsForType zeroes out the fields that are not applicable to this type (container vs not-container).
func (e *Equipment) ClearUnusedFieldsForType() {
	if !e.Container() {
		e.EquipmentMutableContainerOnly = EquipmentMutableContainerOnly{}
	}
}

// Clone implements Node.
func (e *Equipment) Clone(entity *gurps.Entity, parent *Equipment, preserveID bool) *Equipment {
	// TODO: This can probably be improved by not creating a new Equipment and then copying everything.
	other := NewEquipment(entity, parent, e.Container())
	other.CopyFrom(e)
	if preserveID {
		other.LocalID = e.LocalID
	}
	if e.HasChildren() {
		other.Children = make([]*Equipment, 0, len(e.Children))
		for _, child := range e.Children {
			other.Children = append(other.Children, child.Clone(entity, other, preserveID))
		}
	}
	return other
}

// CopyFrom implements node.EditorData. Generates a new local ID in this object.
func (e *Equipment) CopyFrom(other *Equipment) {
	e.copyFrom(other.Entity, other, false)
	e.LocalID = tid.MustNewTID(e.LocalID[0])
}

// ApplyTo implements node.EditorData. Preserves the existing local ID in 'other'.
func (e *Equipment) ApplyTo(other *Equipment) {
	id := other.LocalID
	other.copyFrom(other.Entity, e, true)
	other.LocalID = id
}

func (e *Equipment) copyFrom(entity *gurps.Entity, other *Equipment, isApply bool) {
	e.EquipmentData = other.EquipmentData
	e.Tags = txt.CloneStringSlice(e.Tags)
	e.Modifiers = nil
	if len(other.Modifiers) != 0 {
		e.Modifiers = make([]*gurps.EquipmentModifier, 0, len(other.Modifiers))
		for _, one := range other.Modifiers {
			e.Modifiers = append(e.Modifiers, one.Clone(entity, nil, true))
		}
	}
	e.Prereq = e.Prereq.CloneResolvingEmpty(false, isApply)
	e.Weapons = nil
	if len(other.Weapons) != 0 {
		e.Weapons = make([]*gurps.Weapon, 0, len(other.Weapons))
		for _, one := range other.Weapons {
			e.Weapons = append(e.Weapons, one.Clone(entity, nil, true))
		}
	}
	e.Features = other.Features.Clone()
	// TODO: Is it correct to not clone the children here?
}

// MarshalJSON implements json.Marshaler.
func (e *Equipment) MarshalJSON() ([]byte, error) {
	e.ClearUnusedFieldsForType()
	defUnits := gurps.SheetSettingsFor(e.Entity).DefaultWeightUnits
	type calc struct {
		ExtendedValue           fxp.Int     `json:"extended_value"`
		ExtendedWeight          fxp.Weight  `json:"extended_weight"`
		ExtendedWeightForSkills *fxp.Weight `json:"extended_weight_for_skills,omitempty"`
		ResolvedNotes           string      `json:"resolved_notes,omitempty"`
		UnsatisfiedReason       string      `json:"unsatisfied_reason,omitempty"`
	}
	data := struct {
		EquipmentData
		Calc calc `json:"calc"`
	}{
		EquipmentData: e.EquipmentData,
		Calc: calc{
			ExtendedValue:           e.ExtendedValue(),
			ExtendedWeight:          e.ExtendedWeight(false, defUnits),
			ExtendedWeightForSkills: nil,
			UnsatisfiedReason:       e.UnsatisfiedReason,
		},
	}
	notes := e.resolveLocalNotes()
	if notes != e.Notes {
		data.Calc.ResolvedNotes = notes
	}
	if e.WeightIgnoredForSkills && e.Equipped {
		w := e.ExtendedWeight(true, defUnits)
		data.Calc.ExtendedWeightForSkills = &w
	}
	return json.Marshal(&data)
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *Equipment) UnmarshalJSON(data []byte) error {
	var localData struct {
		EquipmentData
		// Old data fields
		Type       string   `json:"type"`
		Categories []string `json:"categories"`
	}
	if err := json.Unmarshal(data, &localData); err != nil {
		return err
	}
	if !tid.IsKindAndValid(localData.LocalID, 'E') && !tid.IsKindAndValid(localData.LocalID, 'e') {
		switch localData.Type {
		case "equipment":
			localData.LocalID = tid.MustNewTID('e')
		case "equipment_container":
			localData.LocalID = tid.MustNewTID('E')
		default:
			return errs.New("invalid data type")
		}
	}
	e.EquipmentData = localData.EquipmentData
	e.ClearUnusedFieldsForType()
	e.Tags = gurps.ConvertOldCategoriesToTags(e.Tags, localData.Categories)
	slices.Sort(e.Tags)
	if e.Container() {
		if e.Quantity == 0 {
			// Old formats omitted the quantity for containers. Try to see if it was omitted or if it was explicitly
			// set to zero.
			m := make(map[string]any)
			if err := json.Unmarshal(data, &m); err == nil {
				if _, exists := m["quantity"]; !exists {
					e.Quantity = fxp.One
				}
			}
		}
		for _, one := range e.Children {
			one.Parent = e
		}
	}
	return nil
}
