package main

//go:generate go run enumgen.go

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/txt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	rootDir   = ".."
	genSuffix = "_gen.go"
)

type enumValue struct {
	Name          string
	Key           string
	OldKeys       []string
	String        string
	Alt           string
	NoLocalize    bool
	EmptyStringOK bool
}

type enumInfo struct {
	Pkg        string
	Name       string
	Desc       string
	StandAlone bool
	Values     []enumValue
}

func main() {
	const (
		enumTmpl = "enum.go.tmpl"
	)
	removeExistingGenFiles()
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/paper",
		Name:       "orientation",
		Desc:       "holds the orientation of the page",
		StandAlone: true,
		Values: []enumValue{
			{
				Key: "portrait",
			},
			{
				Key: "landscape",
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/paper",
		Name:       "units",
		Desc:       "holds the real-world length unit type",
		StandAlone: true,
		Values: []enumValue{
			{
				Name:       "Inch",
				Key:        "in",
				String:     "in",
				NoLocalize: true,
			},
			{
				Name:       "Centimeter",
				Key:        "cm",
				String:     "cm",
				NoLocalize: true,
			},
			{
				Name:       "Millimeter",
				Key:        "mm",
				String:     "mm",
				NoLocalize: true,
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/paper",
		Name:       "size",
		Desc:       "holds a standard paper dimension",
		StandAlone: true,
		Values: []enumValue{
			{
				Key: "letter",
			},
			{
				Key: "legal",
			},
			{
				Key: "tabloid",
			},
			{
				Key: "a0",
			},
			{
				Key: "a1",
			},
			{
				Key: "a2",
			},
			{
				Key: "a3",
			},
			{
				Key: "a4",
			},
			{
				Key: "a5",
			},
			{
				Key: "a6",
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/settings/display",
		Name:       "option",
		Desc:       "holds a display option",
		StandAlone: true,
		Values: []enumValue{
			{
				Key: "not_shown",
			},
			{
				Key: "inline",
			},
			{
				Key: "tooltip",
			},
			{
				Key:    "inline_and_tooltip",
				String: "Inline & Tooltip",
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps/trait",
		Name:       "affects",
		Desc:       "describes how a TraitModifier affects the point cost",
		StandAlone: true,
		Values: []enumValue{
			{
				Key:    "total",
				String: "to cost",
			},
			{
				Key:    "base_only",
				String: "to base cost only",
				Alt:    "(base only)",
			},
			{
				Key:    "levels_only",
				String: "to leveled cost only",
				Alt:    "(levels only)",
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps/trait",
		Name:       "container_type",
		Desc:       "holds the type of a trait container",
		StandAlone: true,
		Values: []enumValue{
			{
				Key: "group",
			},
			{
				Key:    "meta_trait",
				String: "Meta-Trait",
			},
			{
				Key: "race",
			},
			{
				Key: "alternative_abilities",
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps/trait",
		Name:       "modifier_cost_type",
		Desc:       "describes how a TraitModifier's point cost is applied",
		StandAlone: true,
		Values: []enumValue{
			{
				Key:    "percentage",
				String: "%",
			},
			{
				Key:    "points",
				String: "points",
			},
			{
				Key:    "multiplier",
				String: "×",
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps/attribute",
		Name:       "bonus_limitation",
		Desc:       "holds a limitation for an AttributeBonus",
		StandAlone: true,
		Values: []enumValue{
			{
				Key:           "none",
				EmptyStringOK: true,
				NoLocalize:    true,
			},
			{
				Key:    "striking_only",
				String: "for striking only",
			},
			{
				Key:    "lifting_only",
				String: "for lifting only",
			},
			{
				Key:    "throwing_only",
				String: "for throwing only",
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps/attribute",
		Name:       "damage_progression",
		Desc:       "controls how Thrust and Swing are calculated",
		StandAlone: true,
		Values: []enumValue{
			{
				Key: "basic_set",
			},
			{
				Key: "knowing_your_own_strength",
				Alt: "Pyramid 3-83, pages 16-19",
			},
			{
				Key: "no_school_grognard_damage",
				Alt: "https://noschoolgrognard.blogspot.com/2013/04/adjusting-swing-damage-in-dungeon.html",
			},
			{
				Key:    "thrust_equals_swing_minus_2",
				String: "Thrust = Swing-2",
				Alt:    "https://github.com/richardwilkes/gcs/issues/97",
			},
			{
				Key:    "swing_equals_thrust_plus_2",
				String: "Swing = Thrust+2",
				Alt:    "Houserule originating with Kevin Smyth. See https://gamingballistic.com/2020/12/04/df-eastmarch-boss-fight-and-house-rules/",
			},
			{
				Key: "phoenix_flame_d3",
				Alt: "Houserule that use d3s instead of d6s for Damage. See: https://github.com/richardwilkes/gcs/pull/393",
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps/attribute",
		Name:       "threshold_op",
		Desc:       "holds an operation to apply when a pool threshold is hit",
		StandAlone: true,
		Values: []enumValue{
			{
				Key: "unknown",
				Alt: "Unknown",
			},
			{
				Key: "halve_move",
				Alt: "Halve Move (round up)",
			},
			{
				Key: "halve_dodge",
				Alt: "Halve Dodge (round up)",
			},
			{
				Name:   "HalveST",
				Key:    "halve_st",
				String: "Halve Strength",
				Alt:    "Halve Strength (round up; does not affect HP and damage)",
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps/attribute",
		Name:       "type",
		Desc:       "holds the type of an attribute definition",
		StandAlone: true,
		Values: []enumValue{
			{
				Key: "integer",
			},
			{
				Key: "decimal",
			},
			{
				Key: "pool",
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps/datafile",
		Name:       "encumbrance",
		Desc:       "holds the encumbrance level",
		StandAlone: true,
		Values: []enumValue{
			{
				Key: "none",
			},
			{
				Key: "light",
			},
			{
				Key: "medium",
			},
			{
				Key: "heavy",
			},
			{
				Key:    "extra_heavy",
				String: "X-Heavy",
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps/datafile",
		Name:       "type",
		Desc:       "holds the type of an Entity",
		StandAlone: true,
		Values: []enumValue{
			{
				Name:   "PC",
				Key:    "character",
				String: "PC",
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps/equipment",
		Name:       "modifier_cost_type",
		Desc:       "describes how an EquipmentModifier's cost is applied",
		StandAlone: true,
		Values: []enumValue{
			{
				Name:   "OriginalCost",
				Key:    "to_original_cost",
				String: "to original cost",
				Alt:    `"+5", "-5", "+10%", "-10%", "x3.2"`,
			},
			{
				Name:   "BaseCost",
				Key:    "to_base_cost",
				String: "to base cost",
				Alt:    `"x2", "+2 CF", "-0.2 CF"`,
			},
			{
				Name:   "FinalBaseCost",
				Key:    "to_final_base_cost",
				String: "to final base cost",
				Alt:    `"+5", "-5", "+10%", "-10%", "x3.2"`,
			},
			{
				Name:   "FinalCost",
				Key:    "to_final_cost",
				String: "to final cost",
				Alt:    `"+5", "-5", "+10%", "-10%", "x3.2"`,
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps/equipment",
		Name:       "modifier_cost_value_type",
		Desc:       "describes how an EquipmentModifier's cost is applied",
		StandAlone: true,
		Values: []enumValue{
			{
				Name: "Addition",
				Key:  "+",
			},
			{
				Name: "Percentage",
				Key:  "%",
			},
			{
				Name:   "Multiplier",
				Key:    "x",
				String: "x",
			},
			{
				Name:   "CostFactor",
				Key:    "cf",
				String: "CF",
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps/equipment",
		Name:       "modifier_weight_type",
		Desc:       "describes how an EquipmentModifier's weight is applied",
		StandAlone: true,
		Values: []enumValue{
			{
				Name:   "OriginalWeight",
				Key:    "to_original_weight",
				String: "to original weight",
				Alt:    `"+5 lb", "-5 lb", "+10%", "-10%"`,
			},
			{
				Name:   "BaseWeight",
				Key:    "to_base_weight",
				String: "to base weight",
				Alt:    `"+5 lb", "-5 lb", "x10%", "x3", "x2/3"`,
			},
			{
				Name:   "FinalBaseWeight",
				Key:    "to_final_base_weight",
				String: "to final base weight",
				Alt:    `"+5 lb", "-5 lb", "x10%", "x3", "x2/3"`,
			},
			{
				Name:   "FinalWeight",
				Key:    "to_final_weight",
				String: "to final weight",
				Alt:    `"+5 lb", "-5 lb", "x10%", "x3", "x2/3"`,
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps/equipment",
		Name:       "modifier_weight_value_type",
		Desc:       "describes how an EquipmentModifier's weight is applied",
		StandAlone: true,
		Values: []enumValue{
			{
				Name:       "WeightAddition",
				Key:        "+",
				NoLocalize: true,
			},
			{
				Name:       "WeightPercentageAdder",
				Key:        "%",
				NoLocalize: true,
			},
			{
				Name:       "WeightPercentageMultiplier",
				Key:        "x%",
				String:     "x%",
				NoLocalize: true,
			},
			{
				Name:       "WeightMultiplier",
				Key:        "x",
				String:     "x",
				NoLocalize: true,
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps/feature",
		Name:       "type",
		Desc:       "holds the type of a Feature",
		StandAlone: false,
		Values: []enumValue{
			{
				Key:    "attribute_bonus",
				String: "Gives an attribute modifier of",
			},
			{
				Key:    "conditional_modifier",
				String: "Gives a conditional modifier of",
			},
			{
				Name:   "DRBonus",
				Key:    "dr_bonus",
				String: "Gives a DR bonus of",
			},
			{
				Key:    "reaction_bonus",
				String: "Gives a reaction modifier of",
			},
			{
				Key:    "skill_bonus",
				String: "Gives a skill level modifier of",
			},
			{
				Key:    "skill_point_bonus",
				String: "Gives a skill point modifier of",
			},
			{
				Key:    "spell_bonus",
				String: "Gives a spell level modifier of",
			},
			{
				Key:    "spell_point_bonus",
				String: "Gives a spell point modifier of",
			},
			{
				Key:    "weapon_bonus",
				String: "Gives a weapon damage modifier of",
			},
			{
				Key:    "cost_reduction",
				String: "Reduces the attribute cost of",
			},
			{
				Key:    "contained_weight_reduction",
				String: "Reduces the contained weight by",
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps/prereq",
		Name:       "type",
		Desc:       "holds the type of a Prereq",
		StandAlone: true,
		Values: []enumValue{
			{
				Name:   "List",
				Key:    "prereq_list",
				String: "a list",
			},
			{
				Name:    "Trait",
				Key:     "trait_prereq",
				OldKeys: []string{"advantage_prereq"},
				String:  "a trait",
			},
			{
				Name:   "Attribute",
				Key:    "attribute_prereq",
				String: "the attribute",
			},
			{
				Name:   "ContainedQuantity",
				Key:    "contained_quantity_prereq",
				String: "a contained quantity of",
			},
			{
				Name:   "ContainedWeight",
				Key:    "contained_weight_prereq",
				String: "a contained weight",
			},
			{
				Name:   "Skill",
				Key:    "skill_prereq",
				String: "a skill",
			},
			{
				Name:   "Spell",
				Key:    "spell_prereq",
				String: "spell(s)",
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps/skill",
		Name:       "difficulty",
		Desc:       "holds the difficulty level of a skill",
		StandAlone: true,
		Values: []enumValue{
			{
				Name:       "Easy",
				Key:        "e",
				NoLocalize: true,
			},
			{
				Name:       "Average",
				Key:        "a",
				NoLocalize: true,
			},
			{
				Name:       "Hard",
				Key:        "h",
				NoLocalize: true,
			},
			{
				Name:       "VeryHard",
				Key:        "vh",
				String:     "VH",
				NoLocalize: true,
			},
			{
				Name:       "Wildcard",
				Key:        "w",
				NoLocalize: true,
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps/skill",
		Name:       "selection_type",
		Desc:       "holds the type of a selection",
		StandAlone: true,
		Values: []enumValue{
			{
				Key:    "skills_with_name",
				String: "to skills whose name",
			},
			{
				Key:    "this_weapon",
				String: "to this weapon",
			},
			{
				Key:    "weapons_with_name",
				String: "to weapons whose name",
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps/spell",
		Name:       "comparison_type",
		Desc:       "holds the type of a comparison",
		StandAlone: true,
		Values: []enumValue{
			{
				Key:    "name",
				String: "whose name",
			},
			{
				Key:     "tag",
				OldKeys: []string{"category"},
				String:  "with a tag which",
			},
			{
				Key:    "college",
				String: "whose college name",
			},
			{
				Key:    "college_count",
				String: "from different colleges",
			},
			{
				Key:    "any",
				String: "of any kind",
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps/spell",
		Name:       "match_type",
		Desc:       "holds the type of a match",
		StandAlone: true,
		Values: []enumValue{
			{
				Key:    "all_colleges",
				String: "to all colleges",
			},
			{
				Key:    "college_name",
				String: "to the college whose name",
			},
			{
				Name:   "PowerSource",
				Key:    "power_source_name",
				String: "to the power source whose name",
			},
			{
				Name:   "Spell",
				Key:    "spell_name",
				String: "to the spell whose name",
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps/weapon",
		Name:       "selection_type",
		Desc:       "holds the type of a selection",
		StandAlone: true,
		Values: []enumValue{
			{
				Name:   "WithRequiredSkill",
				Key:    "weapons_with_required_skill",
				String: "to weapons whose required skill name",
			},
			{
				Key:    "this_weapon",
				String: "to this weapon",
			},
			{
				Name:   "WithName",
				Key:    "weapons_with_name",
				String: "to weapons whose name",
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps/weapon",
		Name:       "strength_damage",
		Desc:       "holds the type of strength dice to add to damage",
		StandAlone: true,
		Values: []enumValue{
			{
				Key:    "none",
				String: "None",
			},
			{
				Name:       "Thrust",
				Key:        "thr",
				String:     "thr",
				NoLocalize: true,
			},
			{
				Name:   "LeveledThrust",
				Key:    "thr_leveled",
				String: "thr (leveled)",
			},
			{
				Name:       "Swing",
				Key:        "sw",
				String:     "sw",
				NoLocalize: true,
			},
			{
				Name:   "LeveledSwing",
				Key:    "sw_leveled",
				String: "sw (leveled)",
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps/weapon",
		Name:       "type",
		Desc:       "holds the type of an weapon definition",
		StandAlone: true,
		Values: []enumValue{
			{
				Name: "Melee",
				Key:  "melee_weapon",
				Alt:  i18n.Text("Melee Weapons"),
			},
			{
				Name: "Ranged",
				Key:  "ranged_weapon",
				Alt:  i18n.Text("Ranged Weapons"),
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps",
		Name:       "self_control_roll_adj",
		Desc:       "holds an Adjustment for a self-control roll",
		StandAlone: true,
		Values: []enumValue{
			{
				Name: "NoCRAdj",
				Key:  "none",
				Alt:  "None",
			},
			{
				Key:    "action_penalty",
				String: "Includes an Action Penalty for Failure",
				Alt:    "%d Action Penalty",
			},
			{
				Key:    "reaction_penalty",
				String: "Includes a Reaction Penalty for Failure",
				Alt:    "%d Reaction Penalty",
			},
			{
				Key:    "fright_check_penalty",
				String: "Includes Fright Check Penalty",
				Alt:    "%d Fright Check Penalty",
			},
			{
				Key:    "fright_check_bonus",
				String: "Includes Fright Check Bonus",
				Alt:    "+%d Fright Check Bonus",
			},
			{
				Key:    "minor_cost_of_living_increase",
				String: "Includes a Minor Cost of Living Increase",
				Alt:    "+%d%% Cost of Living Increase",
			},
			{
				Key:    "major_cost_of_living_increase",
				String: "Includes a Major Cost of Living Increase and Merchant Skill Penalty",
				Alt:    "+%d%% Cost of Living Increase",
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps/measure",
		Name:       "length_units",
		Desc:       "holds the length unit type. Note that conversions to/from metric are done using the simplified GURPS metric conversion of 1 yd = 1 meter. For consistency, all metric lengths are converted to meters, then to yards, rather than the variations at different lengths that the GURPS rules suggest",
		StandAlone: true,
		Values: []enumValue{
			{
				Name:   "FeetAndInches",
				Key:    "ft_in",
				String: "Feet & Inches",
			},
			{
				Name:       "Inch",
				Key:        "in",
				String:     "in",
				NoLocalize: true,
			},
			{
				Name:       "Feet",
				Key:        "ft",
				String:     "ft",
				NoLocalize: true,
			},
			{
				Name:       "Yard",
				Key:        "yd",
				String:     "yd",
				NoLocalize: true,
			},
			{
				Name:       "Mile",
				Key:        "mi",
				String:     "mi",
				NoLocalize: true,
			},
			{
				Name:       "Centimeter",
				Key:        "cm",
				String:     "cm",
				NoLocalize: true,
			},
			{
				Name:       "Kilometer",
				Key:        "km",
				String:     "km",
				NoLocalize: true,
			},
			{
				Name:       "Meter",
				Key:        "m",
				String:     "m",
				NoLocalize: true,
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps/measure",
		Name:       "weight_units",
		Desc:       "holds the weight unit type. Note that conversions to/from metric are done using the simplified GURPS metric conversion of 1 lb = 0.5kg. For consistency, all metric weights are converted to kilograms, then to pounds, rather than the variations at different weights that the GURPS rules suggest",
		StandAlone: true,
		Values: []enumValue{
			{
				Name:       "Pound",
				Key:        "lb",
				String:     "lb",
				NoLocalize: true,
			},
			{
				Name:       "PoundAlt",
				Key:        "#",
				String:     "#",
				NoLocalize: true,
			},
			{
				Name:       "Ounce",
				Key:        "oz",
				String:     "oz",
				NoLocalize: true,
			},
			{
				Name:       "Ton",
				Key:        "tn",
				String:     "tn",
				NoLocalize: true,
			},
			{
				Name:       "TonAlt",
				Key:        "t",
				String:     "t",
				NoLocalize: true,
			},
			{
				Name:       "Kilogram",
				Key:        "kg",
				String:     "kg",
				NoLocalize: true,
			},
			{
				Name:       "Gram",
				Key:        "g",
				String:     "g",
				NoLocalize: true,
			},
		},
	})
	processSourceTemplate(enumTmpl, &enumInfo{
		Pkg:        "model/gurps",
		Name:       "cell_type",
		Desc:       "holds the type of table cell",
		StandAlone: true,
		Values: []enumValue{
			{
				Key: "text",
			},
			{
				Key: "toggle",
			},
			{
				Key: "page_ref",
			},
		},
	})
}

func removeExistingGenFiles() {
	root, err := filepath.Abs(rootDir)
	jot.FatalIfErr(err)
	jot.FatalIfErr(filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		name := info.Name()
		if info.IsDir() {
			if name == ".git" {
				return filepath.SkipDir
			}
		} else {
			if strings.HasSuffix(name, genSuffix) {
				jot.FatalIfErr(os.Remove(path))
			}
		}
		return nil
	}))
}

func processSourceTemplate(tmplName string, info *enumInfo) {
	tmpl, err := template.New(tmplName).Funcs(template.FuncMap{
		"add":          add,
		"emptyIfTrue":  emptyIfTrue,
		"fileLeaf":     filepath.Base,
		"firstToLower": txt.FirstToLower,
		"join":         join,
		"last":         last,
		"toCamelCase":  txt.ToCamelCase,
		"toIdentifier": toIdentifier,
		"wrapComment":  wrapComment,
	}).ParseFiles(tmplName)
	jot.FatalIfErr(err)
	var buffer bytes.Buffer
	writeGeneratedFromComment(&buffer, tmplName)
	jot.FatalIfErr(tmpl.Execute(&buffer, info))
	var data []byte
	if data, err = format.Source(buffer.Bytes()); err != nil {
		fmt.Println("unable to format source file: " + filepath.Join(info.Pkg, info.Name+genSuffix))
		data = buffer.Bytes()
	}
	dir := filepath.Join(rootDir, info.Pkg)
	jot.FatalIfErr(os.MkdirAll(dir, 0o750))
	jot.FatalIfErr(os.WriteFile(filepath.Join(dir, info.Name+genSuffix), data, 0o640))
}

func writeGeneratedFromComment(w io.Writer, tmplName string) {
	_, err := fmt.Fprintf(w, "// Code generated from \"%s\" - DO NOT EDIT.\n\n", tmplName)
	jot.FatalIfErr(err)
}

func add(a, b int) int {
	return a + b
}

func join(values []string) string {
	var buffer strings.Builder
	for i, one := range values {
		if i != 0 {
			buffer.WriteString(", ")
		}
		fmt.Fprintf(&buffer, "%q", one)
	}
	return buffer.String()
}

func (e *enumInfo) LocalType() string {
	return txt.FirstToLower(toIdentifier(e.Name)) + "Data"
}

func (e *enumInfo) IDFor(v enumValue) string {
	id := v.Name
	if id == "" {
		id = toIdentifier(v.Key)
	}
	if !e.StandAlone {
		id += toIdentifier(e.Name)
	}
	return id
}

func (e *enumInfo) HasAlt() bool {
	for _, one := range e.Values {
		if one.Alt != "" {
			return true
		}
	}
	return false
}

func (e *enumInfo) HasOldKeys() bool {
	for _, one := range e.Values {
		if len(one.OldKeys) != 0 {
			return true
		}
	}
	return false
}

func (e *enumInfo) NeedI18N() bool {
	for _, one := range e.Values {
		if !one.NoLocalize || one.Alt != "" {
			return true
		}
	}
	return false
}

func (e *enumValue) StringValue() string {
	if e.String == "" && !e.EmptyStringOK {
		return cases.Title(language.AmericanEnglish).String(strings.ReplaceAll(e.Key, "_", " "))
	}
	return e.String
}

func last(in []enumValue) enumValue {
	return in[len(in)-1]
}

func emptyIfTrue(str string, test bool) string {
	if test {
		return ""
	}
	return str
}

func toIdentifier(in string) string {
	var buffer strings.Builder
	useUpper := true
	for i, ch := range in {
		isUpper := ch >= 'A' && ch <= 'Z'
		isLower := ch >= 'a' && ch <= 'z'
		isDigit := ch >= '0' && ch <= '9'
		isAlpha := isUpper || isLower
		if i == 0 && !isAlpha {
			if !isDigit {
				continue
			}
			buffer.WriteString("_")
		}
		if isAlpha {
			if useUpper {
				buffer.WriteRune(unicode.ToUpper(ch))
			} else {
				buffer.WriteRune(unicode.ToLower(ch))
			}
			useUpper = false
		} else {
			if isDigit {
				buffer.WriteRune(ch)
			}
			useUpper = true
		}
	}
	return buffer.String()
}

func wrapComment(in string, cols int) string {
	return txt.Wrap("// ", in, cols)
}
