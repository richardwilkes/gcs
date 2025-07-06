package gio

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/richardwilkes/gcs/v5/model/fxp"
)

// Using simplified structs for now, matching the provided JSON structure.
// We can integrate with the existing model/gurps structs later if needed.

type CharacterSheet struct {
	Version     int            `json:"version"`
	ID          string         `json:"id"`
	TotalPoints int            `json:"total_points"`
	PointsRecord []PointsRecord `json:"points_record"`
	Profile     ProfileData    `json:"profile"`
	Settings    SettingsData   `json:"settings"`
	Attributes  []AttributeValue `json:"attributes"`
	Traits      []Trait        `json:"traits"`
	Skills      []Skill        `json:"skills"`
	CreatedDate time.Time      `json:"created_date"`
	ModifiedDate time.Time     `json:"modified_date"`
	Calc        CalcData       `json:"calc"`
}

type PointsRecord struct {
	When   time.Time `json:"when"`
	Points int       `json:"points"`
	Reason string    `json:"reason"`
}

type ProfileData struct {
	Name         string `json:"name"`
	Age          string `json:"age"`
	Eyes         string `json:"eyes"`
	Hair         string `json:"hair"`
	Skin         string `json:"skin"`
	Handedness   string `json:"handedness"`
	Gender       string `json:"gender"`
	Height       string `json:"height"` // Keeping as string for now
	Weight       string `json:"weight"` // Keeping as string for now
	PlayerName   string `json:"player_name"`
	TechLevel    string `json:"tech_level"`
	Portrait     string `json:"portrait"` // Base64 encoded image data
}

type SettingsData struct {
	Page               PageSettings      `json:"page"`
	BlockLayout        []string          `json:"block_layout"`
	Attributes         []AttributeDef    `json:"attributes"` // These are definitions, not values
	BodyType           BodyType          `json:"body_type"`
	DamageProgression  string            `json:"damage_progression"`
	DefaultLengthUnits string            `json:"default_length_units"`
	DefaultWeightUnits string            `json:"default_weight_units"`
	UserDescriptionDisplay string        `json:"user_description_display"`
	ModifiersDisplay   string            `json:"modifiers_display"`
	NotesDisplay       string            `json:"notes_display"`
	SkillLevelAdjDisplay string          `json:"skill_level_adj_display"`
	ShowSpellAdj       bool              `json:"show_spell_adj"`
}

type PageSettings struct {
	PaperSize   string `json:"paper_size"`
	Orientation string `json:"orientation"`
	TopMargin   string `json:"top_margin"`
	LeftMargin  string `json:"left_margin"`
	BottomMargin string `json:"bottom_margin"`
	RightMargin string `json:"right_margin"`
}

type AttributeDef struct {
	ID                      string          `json:"id"`
	Type                    string          `json:"type"`
	Name                    string          `json:"name"`
	FullName                string          `json:"full_name,omitempty"`
	Base                    string          `json:"base"`
	CostPerPoint            float64         `json:"cost_per_point,omitempty"` // Can be int or float
	CostAdjPercentPerSM     int             `json:"cost_adj_percent_per_sm,omitempty"`
	Thresholds              []PoolThreshold `json:"thresholds,omitempty"`
}


type PoolThreshold struct {
	State       string   `json:"state"`
	Value       string   `json:"value"` // Can be formula
	Explanation string   `json:"explanation,omitempty"`
	Ops         []string `json:"ops,omitempty"`
}

type BodyType struct {
	Name      string         `json:"name"`
	Roll      string         `json:"roll"`
	Locations []HitLocation `json:"locations"`
}

type HitLocation struct {
	ID          string         `json:"id"`
	ChoiceName  string         `json:"choice_name"`
	TableName   string         `json:"table_name"`
	Slots       int            `json:"slots,omitempty"`
	HitPenalty  int            `json:"hit_penalty,omitempty"`
	DRBonus     int            `json:"dr_bonus,omitempty"`
	Description string         `json:"description,omitempty"`
	Calc        HitLocationCalc `json:"calc"`
}

type HitLocationCalc struct {
	RollRange string         `json:"roll_range"`
	DR        map[string]int `json:"dr"`
}

type AttributeValue struct {
	AttrID string      `json:"attr_id"`
	Adj    fxp.Int     `json:"adj"` // Using fxp.Int from existing model
	Calc   AttributeCalc `json:"calc"`
}

type AttributeCalc struct {
	Value   fxp.Int `json:"value"`
	Current fxp.Int `json:"current,omitempty"` // Only for pools
	Points  fxp.Int `json:"points"`
}

type Trait struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Reference  string      `json:"reference,omitempty"`
	Tags       []string    `json:"tags,omitempty"`
	Modifiers  []Modifier  `json:"modifiers,omitempty"`
	BasePoints fxp.Int     `json:"base_points,omitempty"`
	Features   []Feature   `json:"features,omitempty"`
	Calc       PointsCalc  `json:"calc"`
	Children   []Trait     `json:"children,omitempty"` // For nested traits/advantages
	Weapons    []Weapon    `json:"weapons,omitempty"` // For natural attacks
}

type Modifier struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Reference  string  `json:"reference,omitempty"`
	Cost       fxp.Int `json:"cost,omitempty"`
	CostType   string  `json:"cost_type,omitempty"` // "points", "percentage"
	Disabled   bool    `json:"disabled,omitempty"`
	LocalNotes string  `json:"local_notes,omitempty"`
	Levels     fxp.Int `json:"levels,omitempty"`
	Features   []Feature `json:"features,omitempty"`
}

type Feature struct {
	Type          string      `json:"type"`
	SelectionType string      `json:"selection_type,omitempty"`
	Name          FeatureName `json:"name,omitempty"`
	Amount        fxp.Int     `json:"amount,omitempty"`
	Situation     string      `json:"situation,omitempty"`
	Specialization FeatureName `json:"specialization,omitempty"`
}

type FeatureName struct {
	Compare   string `json:"compare"`
	Qualifier string `json:"qualifier"`
}

type PointsCalc struct {
	Points fxp.Int `json:"points"`
}

type Weapon struct {
	ID       string       `json:"id"`
	Damage   DamageSpec   `json:"damage"`
	Usage    string       `json:"usage"`
	Reach    string       `json:"reach,omitempty"`
	Parry    string       `json:"parry,omitempty"` // Can be string like "0", "No"
	Defaults []SkillDefault `json:"defaults"`
	Calc     WeaponCalc   `json:"calc"`
}

type DamageSpec struct {
	Type string `json:"type"`
	St   string `json:"st"` // e.g. "thr", "sw"
	Base string `json:"base,omitempty"`
}

type SkillDefault struct {
	Type      string  `json:"type"` // "dx", "skill", "iq", etc.
	Modifier  fxp.Int `json:"modifier,omitempty"`
	Name      string  `json:"name,omitempty"` // For skill defaults
}

type WeaponCalc struct {
	Level  fxp.Int `json:"level"`
	Damage string  `json:"damage"` // e.g., "1d-4 cr"
	Parry  string  `json:"parry,omitempty"`
}

type Skill struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Reference     string        `json:"reference,omitempty"`
	Tags          []string      `json:"tags,omitempty"`
	Difficulty    string        `json:"difficulty"` // e.g., "dx/h", "iq/a"
	Points        fxp.Int       `json:"points"`
	Calc          SkillCalc     `json:"calc"`
	Specialization string       `json:"specialization,omitempty"`
	TechLevel     string        `json:"tech_level,omitempty"`
	DefaultedFrom *SkillDefault `json:"defaulted_from,omitempty"`
	EncumbrancePenaltyMultiplier int `json:"encumbrance_penalty_multiplier,omitempty"`

}

type SkillCalc struct {
	Level fxp.Int `json:"level"`
	RSL   string  `json:"rsl"` // Relative Skill Level, e.g. "DX+0"
}

type CalcData struct {
	Swing    string    `json:"swing"`
	Thrust   string    `json:"thrust"`
	BasicLift string   `json:"basic_lift"`
	Move     []fxp.Int `json:"move"`
	Dodge    []fxp.Int `json:"dodge"`
}


// LoadCharacterSheet loads a GURPS character sheet from a JSON file.
func LoadCharacterSheet(filePath string) (*CharacterSheet, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var sheet CharacterSheet
	err = json.Unmarshal(data, &sheet)
	if err != nil {
		return nil, err
	}

	return &sheet, nil
}
