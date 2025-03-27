// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package main

//go:generate go run srcgen.go

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

	"github.com/richardwilkes/toolbox/fatal"
	"github.com/richardwilkes/toolbox/txt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	rootDir     = ".."
	genGoSuffix = "_gen.go"
)

type enumValue struct {
	Name          string
	Key           string
	OldKeys       []string
	String        string
	Alt           string
	NoLocalize    bool
	NoLocalizeAlt bool
	EmptyStringOK bool
}

type enumInfo struct {
	Pkg    string
	Name   string
	Desc   string
	Values []*enumValue
}

func main() {
	removeExistingGenFiles()
	for _, one := range allEnums {
		processEnumTemplate("enum.go.tmpl", one)
	}
}

func removeExistingGenFiles() {
	root, err := filepath.Abs(rootDir)
	fatal.IfErr(err)
	fatal.IfErr(filepath.Walk(root, func(path string, info os.FileInfo, _ error) error {
		name := info.Name()
		switch {
		case info.IsDir():
			if name == ".git" {
				return filepath.SkipDir
			}
		case strings.HasSuffix(name, genGoSuffix):
			fatal.IfErr(os.Remove(path))
		}
		return nil
	}))
}

func processEnumTemplate(tmplName string, info *enumInfo) {
	tmpl, err := template.New(tmplName).Funcs(template.FuncMap{
		"fileLeaf":     filepath.Base,
		"join":         join,
		"last":         last,
		"toIdentifier": toIdentifier,
		"wrapComment":  wrapComment,
	}).ParseFiles(tmplName)
	fatal.IfErr(err)
	var buffer bytes.Buffer
	writeGeneratedFromComment(&buffer, tmplName)
	fatal.IfErr(tmpl.Execute(&buffer, info))
	var data []byte
	if data, err = format.Source(buffer.Bytes()); err != nil {
		fmt.Println("unable to format source file: " + filepath.Join(info.Pkg, info.Name+genGoSuffix))
		data = buffer.Bytes()
	}
	dir := filepath.Join(rootDir, info.Pkg)
	fatal.IfErr(os.MkdirAll(dir, 0o750))
	fatal.IfErr(os.WriteFile(filepath.Join(dir, info.Name+genGoSuffix), data, 0o640))
}

func writeGeneratedFromComment(w io.Writer, tmplName string) {
	_, err := fmt.Fprintf(w, "// Code generated from \"%s\" - DO NOT EDIT.\n\n", tmplName)
	fatal.IfErr(err)
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

func (e *enumInfo) IDFor(v *enumValue) string {
	id := v.Name
	if id == "" {
		id = toIdentifier(v.Key)
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
		if !one.NoLocalize || (one.Alt != "" && !one.NoLocalizeAlt) {
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

func last(in []*enumValue) *enumValue {
	return in[len(in)-1]
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

var allEnums = []*enumInfo{
	{
		Pkg:  "model/fxp",
		Name: "length_unit",
		Desc: "holds the length unit type. Note that conversions to/from metric are done using the simplified GURPS metric conversion of 1 yd = 1 meter. For consistency, all metric lengths are converted to meters, then to yards, rather than the variations at different lengths that the GURPS rules suggest",
		Values: []*enumValue{
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
	},
	{
		Pkg:  "model/fxp",
		Name: "weight_unit",
		Desc: "holds the weight unit type. Note that conversions to/from metric are done using the simplified GURPS metric conversion of 1 lb = 0.5kg. For consistency, all metric weights are converted to kilograms, then to pounds, rather than the variations at different weights that the GURPS rules suggest",
		Values: []*enumValue{
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
	},
	{
		Pkg:  "model/gurps/enums/affects",
		Name: "option",
		Desc: "describes how a TraitModifier affects the point cost",
		Values: []*enumValue{
			{
				Key:           "total",
				String:        "to cost",
				NoLocalizeAlt: true,
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
	},
	{
		Pkg:  "model/gurps/enums/attribute",
		Name: "placement",
		Desc: "determines the placement of the attribute on the sheet",
		Values: []*enumValue{
			{Key: "automatic"},
			{Key: "primary"},
			{Key: "secondary"},
			{Key: "hidden"},
		},
	},
	{
		Pkg:  "model/gurps/enums/attribute",
		Name: "type",
		Desc: "holds the type of an attribute definition",
		Values: []*enumValue{
			{Key: "integer"},
			{Key: "integer_ref", String: "Integer (Display Only)"},
			{Key: "decimal"},
			{Key: "decimal_ref", String: "Decimal (Display Only)"},
			{Key: "pool"},
			{Key: "pool_ref", String: "Pool (Display Only for Maximum)"},
			{Key: "primary_separator"},
			{Key: "secondary_separator"},
			{Key: "pool_separator"},
		},
	},
	{
		Pkg:  "model/gurps/enums/autoscale",
		Name: "option",
		Desc: "holds the possible auto-scaling options",
		Values: []*enumValue{
			{Key: "no", String: "No Auto-Scaling"},
			{Key: "fit_width"},
			{Key: "fit_page"},
		},
	},
	{
		Pkg:  "model/gurps/enums/cell",
		Name: "type",
		Desc: "holds the type of table cell",
		Values: []*enumValue{
			{Key: "text"},
			{Key: "tags"},
			{Key: "toggle"},
			{Key: "page_ref"},
			{Key: "markdown"},
		},
	},
	{
		Pkg:  "model/gurps/enums/container",
		Name: "type",
		Desc: "holds the type of a trait container",
		Values: []*enumValue{
			{Key: "group"},
			{Key: "alternative_abilities"},
			{
				Key:     "ancestry",
				OldKeys: []string{"race"},
			},
			{Key: "attributes"},
			{
				Key:    "meta_trait",
				String: "Meta-Trait",
			},
		},
	},
	{
		Pkg:  "model/gurps/enums/dgroup",
		Name: "group",
		Desc: "holds the set of dockable groupings",
		Values: []*enumValue{
			{Key: "character_sheets"},
			{Key: "character_templates"},
			{Key: "loot_sheets"},
			// {Key: "campaigns"}, // TODO: Re-enable Campaign files
			{Key: "editors"},
			{Key: "images"},
			{Key: "libraries"},
			{Key: "markdown"},
			{Key: "pdfs", Name: "PDFs", String: "PDFs"},
			{Key: "settings"},
			{Key: "sub-editors"},
		},
	},
	{
		Pkg:  "model/gurps/enums/difficulty",
		Name: "level",
		Desc: "holds the difficulty level of a skill",
		Values: []*enumValue{
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
	},
	{
		Pkg:  "model/gurps/enums/display",
		Name: "option",
		Desc: "holds a display option",
		Values: []*enumValue{
			{Key: "not_shown"},
			{Key: "inline"},
			{Key: "tooltip"},
			{
				Key:    "inline_and_tooltip",
				String: "Inline & Tooltip",
			},
		},
	},
	{
		Pkg:  "model/gurps/enums/emcost",
		Name: "type",
		Desc: "describes how an Equipment Modifier's cost is applied",
		Values: []*enumValue{
			{
				Name:   "Original",
				Key:    "to_original_cost",
				String: "to original cost",
				Alt:    `"+5", "-5", "+10%", "-10%", "x3.2"`,
			},
			{
				Name:   "Base",
				Key:    "to_base_cost",
				String: "to base cost",
				Alt:    `"x2", "+2 CF", "-0.2 CF"`,
			},
			{
				Name:   "FinalBase",
				Key:    "to_final_base_cost",
				String: "to final base cost",
				Alt:    `"+5", "-5", "+10%", "-10%", "x3.2"`,
			},
			{
				Name:   "Final",
				Key:    "to_final_cost",
				String: "to final cost",
				Alt:    `"+5", "-5", "+10%", "-10%", "x3.2"`,
			},
		},
	},
	{
		Pkg:  "model/gurps/enums/emcost",
		Name: "value",
		Desc: "describes how an Equipment Modifier's cost value is applied",
		Values: []*enumValue{
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
	},
	{
		Pkg:  "model/gurps/enums/emweight",
		Name: "type",
		Desc: "describes how an Equipment Modifier's weight is applied",
		Values: []*enumValue{
			{
				Name:   "Original",
				Key:    "to_original_weight",
				String: "to original weight",
				Alt:    `"+5 lb", "-5 lb", "+10%", "-10%"`,
			},
			{
				Name:   "Base",
				Key:    "to_base_weight",
				String: "to base weight",
				Alt:    `"+5 lb", "-5 lb", "x10%", "x3", "x2/3"`,
			},
			{
				Name:   "FinalBase",
				Key:    "to_final_base_weight",
				String: "to final base weight",
				Alt:    `"+5 lb", "-5 lb", "x10%", "x3", "x2/3"`,
			},
			{
				Name:   "Final",
				Key:    "to_final_weight",
				String: "to final weight",
				Alt:    `"+5 lb", "-5 lb", "x10%", "x3", "x2/3"`,
			},
		},
	},
	{
		Pkg:  "model/gurps/enums/emweight",
		Name: "value",
		Desc: "describes how an Equipment Modifier's weight value is applied",
		Values: []*enumValue{
			{
				Name:       "Addition",
				Key:        "+",
				NoLocalize: true,
			},
			{
				Name:       "PercentageAdder",
				Key:        "%",
				NoLocalize: true,
			},
			{
				Name:       "PercentageMultiplier",
				Key:        "x%",
				String:     "x%",
				NoLocalize: true,
			},
			{
				Name:       "Multiplier",
				Key:        "x",
				String:     "x",
				NoLocalize: true,
			},
		},
	},
	{
		Pkg:  "model/gurps/enums/encumbrance",
		Name: "level",
		Desc: "holds the encumbrance level",
		Values: []*enumValue{
			{
				Name: "No",
				Key:  "none",
			},
			{Key: "light"},
			{Key: "medium"},
			{Key: "heavy"},
			{
				Key:    "extra_heavy",
				String: "X-Heavy",
			},
		},
	},
	{
		Pkg:  "model/gurps/enums/feature",
		Name: "type",
		Desc: "holds the type of a Feature",
		Values: []*enumValue{
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
				Key:    "weapon_acc_bonus",
				String: "Gives a weapon accuracy modifier of",
			},
			{
				Key:    "weapon_scope_acc_bonus",
				String: "Gives a weapon scope accuracy modifier of",
			},
			{
				Name:   "WeaponDRDivisorBonus",
				Key:    "weapon_dr_divisor_bonus",
				String: "Gives a weapon DR divisor modifier of",
			},
			{
				Name:   "WeaponEffectiveSTBonus",
				Key:    "weapon_effective_st_bonus",
				String: "Gives a weapon effective ST modifier of",
			},
			{
				Name:   "WeaponMinSTBonus",
				Key:    "weapon_min_st_bonus",
				String: "Gives a weapon minimum ST modifier of",
			},
			{
				Key:    "weapon_min_reach_bonus",
				String: "Gives a weapon minimum reach modifier of",
			},
			{
				Key:    "weapon_max_reach_bonus",
				String: "Gives a weapon maximum reach modifier of",
			},
			{
				Key:    "weapon_half_damage_range_bonus",
				String: "Gives a weapon half-damage range modifier of",
			},
			{
				Key:    "weapon_min_range_bonus",
				String: "Gives a weapon minimum range modifier of",
			},
			{
				Key:    "weapon_max_range_bonus",
				String: "Gives a weapon maximum range modifier of",
			},
			{
				Key:    "weapon_recoil_bonus",
				String: "Gives a weapon recoil modifier of",
			},
			{
				Key:    "weapon_bulk_bonus",
				String: "Gives a weapon bulk modifier of",
			},
			{
				Key:    "weapon_parry_bonus",
				String: "Gives a weapon parry modifier of",
			},
			{
				Key:    "weapon_block_bonus",
				String: "Gives a weapon block modifier of",
			},
			{
				Key:    "weapon_rof_mode_1_shots_bonus",
				String: "Gives a weapon shots per attack (mode 1) modifier of",
			},
			{
				Key:    "weapon_rof_mode_1_secondary_bonus",
				String: "Gives a weapon secondary projectiles (mode 1) modifier of",
			},
			{
				Key:    "weapon_rof_mode_2_shots_bonus",
				String: "Gives a weapon shots per attack (mode 2) modifier of",
			},
			{
				Key:    "weapon_rof_mode_2_secondary_bonus",
				String: "Gives a weapon secondary projectiles (mode 2) modifier of",
			},
			{
				Key:    "weapon_non_chamber_shots_bonus",
				String: "Gives a weapon non-chamber shots modifier of",
			},
			{
				Key:    "weapon_chamber_shots_bonus",
				String: "Gives a weapon chamber shots modifier of",
			},
			{
				Key:    "weapon_shot_duration_bonus",
				String: "Gives a weapon shot duration modifier of",
			},
			{
				Key:    "weapon_reload_time_bonus",
				String: "Gives a weapon reload time modifier of",
			},
			{
				Key:    "weapon_switch",
				String: "Set the weapon flag",
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
	},
	{
		Pkg:  "model/gurps/enums/namegen",
		Name: "builtin",
		Desc: "holds a built-in name data type",
		Values: []*enumValue{
			{Key: "none"},
			{Key: "american_male"},
			{Key: "american_female"},
			{Key: "american_last"},
			{Key: "unweighted_american_male"},
			{Key: "unweighted_american_female"},
			{Key: "unweighted_american_last"},
		},
	},
	{
		Pkg:  "model/gurps/enums/namegen",
		Name: "type",
		Desc: "holds a name generation type",
		Values: []*enumValue{
			{Key: "simple"},
			{Key: "markov_letter", OldKeys: []string{"markov_chain"}},
			{Key: "markov_run"},
			{Key: "compound"},
		},
	},
	{
		Pkg:  "model/gurps/enums/picker",
		Name: "type",
		Desc: "holds the type of template picker",
		Values: []*enumValue{
			{Key: "not_applicable"},
			{Key: "count"},
			{Key: "points"},
		},
	},
	{
		Pkg:  "model/gurps/enums/prereq",
		Name: "type",
		Desc: "holds the type of a Prereq",
		Values: []*enumValue{
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
				Name:   "EquippedEquipment",
				Key:    "equipped_equipment",
				String: "has equipped equipment",
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
	},
	{
		Pkg:  "model/gurps/enums/progression",
		Name: "option",
		Desc: "controls how Thrust and Swing are calculated",
		Values: []*enumValue{
			{
				Key: "basic_set",
				Alt: "*The standard damage progression*",
			},
			{
				Key: "knowing_your_own_strength",
				Alt: "*From [Pyramid 3-83, pages 16-19](PY83:16)*",
			},
			{
				Key: "no_school_grognard_damage",
				Alt: "*From [Adjusting Swing Damage in Dungeon Fantasy](https://noschoolgrognard.blogspot.com/2013/04/adjusting-swing-damage-in-dungeon.html)*",
			},
			{
				Key:    "thrust_equals_swing_minus_2",
				String: "Thrust = Swing-2",
				Alt:    "*From [Alternate Damage Scheme (Thr = Sw-2)](https://github.com/richardwilkes/gcs/issues/97)*",
			},
			{
				Key:    "swing_equals_thrust_plus_2",
				String: "Swing = Thrust+2",
				Alt:    "*From a [house rule](https://gamingballistic.com/2020/12/04/df-eastmarch-boss-fight-and-house-rules/) originating with Kevin Smyth*",
			},
			{
				Key:    "tbone_1",
				String: "T Bone's New Damage for ST (option 1)",
				Alt:    "*From [T Bone's Games Diner](https://www.gamesdiner.com/rules-nugget-gurps-new-damage-for-st/)*",
			},
			{
				Key:    "tbone_1_clean",
				String: "T Bone's New Damage for ST (option 1, cleaned)",
				Alt:    "*From [T Bone's Games Diner](https://www.gamesdiner.com/rules-nugget-gurps-new-damage-for-st/)*",
			},
			{
				Key:    "tbone_2",
				String: "T Bone's New Damage for ST (option 2)",
				Alt:    "*From [T Bone's Games Diner](https://www.gamesdiner.com/rules-nugget-gurps-new-damage-for-st/)*",
			},
			{
				Key:    "tbone_2_clean",
				String: "T Bone's New Damage for ST (option 2, cleaned)",
				Alt:    "*From [T Bone's Games Diner](https://www.gamesdiner.com/rules-nugget-gurps-new-damage-for-st/)*",
			},
			{
				Key: "phoenix_flame_d3",
				Alt: "*From a [house rule](https://github.com/richardwilkes/gcs/pull/393) that uses d3s instead of d6s for damage*",
			},
		},
	},
	{
		Pkg:  "model/gurps/enums/selfctrl",
		Name: "adjustment",
		Desc: "holds an Adjustment for a self-control roll",
		Values: []*enumValue{
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
	},
	{
		Pkg:  "model/gurps/enums/skillsel",
		Name: "type",
		Desc: "holds the type of a selection",
		Values: []*enumValue{
			{
				Name:   "Name",
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
	},
	{
		Pkg:  "model/gurps/enums/spellcmp",
		Name: "type",
		Desc: "holds the type of a comparison",
		Values: []*enumValue{
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
	},
	{
		Pkg:  "model/gurps/enums/spellmatch",
		Name: "type",
		Desc: "holds the type of a match",
		Values: []*enumValue{
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
				Name:   "Name",
				Key:    "spell_name",
				String: "to the spell whose name",
			},
		},
	},
	{
		Pkg:  "model/gurps/enums/srcstate",
		Name: "value",
		Desc: "describes the state of a source compared to a piece of data",
		Values: []*enumValue{
			{
				Key:           "custom",
				String:        "Custom data which did not come from a library source",
				Alt:           "—",
				NoLocalizeAlt: true,
			},
			{
				Key:           "matched",
				String:        "Data matches library source data",
				NoLocalizeAlt: true,
			},
			{
				Key:           "mismatched",
				String:        "Data does NOT match library source data",
				Alt:           "!",
				NoLocalizeAlt: true,
			},
			{
				Key:           "missing",
				String:        "Unable to locate the library source data to compare against",
				Alt:           "?",
				NoLocalizeAlt: true,
			},
		},
	},
	{
		Pkg:  "model/gurps/enums/stdmg",
		Name: "option",
		Desc: "holds the type of strength dice to add to damage",
		Values: []*enumValue{
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
				Name:   "LiftingThrust",
				Key:    "lift_thr",
				String: "lifting thr",
			},
			{
				Name:   "TelekineticThrust",
				Key:    "tk_thr",
				String: "telekinetic thr",
			},
			{
				Name:   "IQThrust",
				Key:    "iq_thr",
				String: "iq thr",
			},
			{
				Name:       "Swing",
				Key:        "sw",
				String:     "sw",
				NoLocalize: true,
			},
			{
				Name:   "LiftingSwing",
				Key:    "lift_sw",
				String: "lifting sw",
			},
			{
				Name:   "TelekineticSwing",
				Key:    "tk_sw",
				String: "telekinetic sw",
			},
			{
				Name:   "IQSwing",
				Key:    "iq_sw",
				String: "iq sw",
			},
			// The following entries are only here to allow loading of old data and will be converted at load
			{
				Name:   "OldLeveledThrust",
				Key:    "thr_leveled",
				String: "thr (leveled)",
			},
			{
				Name:   "OldLeveledSwing",
				Key:    "sw_leveled",
				String: "sw (leveled)",
			},
		},
	},
	{
		Pkg:  "model/gurps/enums/stlimit",
		Name: "option",
		Desc: "holds a limitation for a Strength AttributeBonus",
		Values: []*enumValue{
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
	},
	{
		Pkg:  "model/gurps/enums/study",
		Name: "level",
		Desc: "holds the number of study hours required per point",
		Values: []*enumValue{
			{
				Name:   "Standard",
				Key:    "",
				String: "Standard",
			},
			{
				Name:   "Level1",
				Key:    "180",
				String: "Reduction for Talent level 1",
			},
			{
				Name:   "Level2",
				Key:    "160",
				String: "Reduction for Talent level 2",
			},
			{
				Name:   "Level3",
				Key:    "140",
				String: "Reduction for Talent level 3",
			},
			{
				Name:   "Level4",
				Key:    "120",
				String: "Reduction for Talent level 4",
			},
		},
	},
	{
		Pkg:  "model/gurps/enums/study",
		Name: "type",
		Desc: "holds the type of study",
		Values: []*enumValue{
			{
				Key:    "self",
				String: "Self-Taught",
			},
			{
				Key:    "job",
				String: "On-the-Job Training",
			},
			{
				Key:    "teacher",
				String: "Professional Teacher",
			},
			{
				Key:    "intensive",
				String: "Intensive Training",
			},
		},
	},
	{
		Pkg:  "model/gurps/enums/threshold",
		Name: "op",
		Desc: "holds an operation to apply when a pool threshold is hit",
		Values: []*enumValue{
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
	},
	{
		Pkg:  "model/gurps/enums/tmcost",
		Name: "type",
		Desc: "describes how a TraitModifier's point cost is applied",
		Values: []*enumValue{
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
	},
	{
		Pkg:  "model/gurps/enums/wsel",
		Name: "type",
		Desc: "holds the type of a weapon selection",
		Values: []*enumValue{
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
	},
	{
		Pkg:  "model/gurps/enums/wswitch",
		Name: "type",
		Desc: "holds the type of a weapon switch",
		Values: []*enumValue{
			{
				Key:           "not_switched",
				EmptyStringOK: true,
				NoLocalize:    true,
			},
			{Key: "can_block"},
			{Key: "can_parry"},
			{Key: "close_combat"},
			{Key: "fencing"},
			{
				Key:    "full_auto_1",
				String: "Fully Automatic (Mode 1)",
			},
			{
				Key:    "full_auto_2",
				String: "Fully Automatic (Mode 2)",
			},
			{
				Key:    "bipod",
				String: "Has Bipod",
			},
			{
				Key:    "controlled_bursts_1",
				String: "High-cyclic Controlled Bursts (Mode 1)",
			},
			{
				Key:    "controlled_bursts_2",
				String: "High-cyclic Controlled Bursts (Mode 2)",
			},
			{Key: "jet"},
			{Key: "mounted"},
			{Key: "muscle_powered"},
			{Key: "range_in_miles"},
			{Key: "reach_change_requires_ready"},
			{Key: "reload_time_is_per_shot"},
			{Key: "retracting_stock"},
			{
				Key:    "two-handed",
				String: "Two-handed",
			},
			{Key: "thrown"},
			{Key: "unbalanced"},
			{
				Key:    "two-handed_and_unready_after_attack",
				String: "Two-handed and Unready After Attack",
			},
			{
				Key:    "musket_rest",
				String: "Uses a Musket Rest",
			},
		},
	},
	{
		Pkg:  "model/paper",
		Name: "orientation",
		Desc: "holds the orientation of the page",
		Values: []*enumValue{
			{Key: "portrait"},
			{Key: "landscape"},
		},
	},
	{
		Pkg:  "model/paper",
		Name: "unit",
		Desc: "holds the real-world length unit type",
		Values: []*enumValue{
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
	},
}
