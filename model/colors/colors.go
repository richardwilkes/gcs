// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package colors

import (
	"io/fs"
	"sync"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/themeset"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

// minimumVersion is the oldest theme color data that can be loaded. Files from before the theme rework hold a different
// set of colors, so they are refused rather than silently loaded as a mostly factory theme.
const minimumVersion = 5

var (
	_       themeset.Entry[unison.ThemeColor]                  = &ThemedColor{}
	_       themeset.Provider[unison.ThemeColor, *ThemedColor] = provider{}
	once    sync.Once
	current []*ThemedColor
	factory []*ThemedColor
)

// Additional theme colors
var (
	Header                  = &unison.ThemeColor{Light: unison.RGB(80, 80, 80), Dark: unison.RGB(64, 64, 64)}
	OnHeader                = Header.DeriveOn()
	TintPortrait            = &unison.ThemeColor{Light: unison.Transparent, Dark: unison.Transparent}
	TintIdentity            = &unison.ThemeColor{Light: unison.Transparent, Dark: unison.Transparent}
	TintMisc                = &unison.ThemeColor{Light: unison.Transparent, Dark: unison.Transparent}
	TintPoints              = &unison.ThemeColor{Light: unison.Transparent, Dark: unison.Transparent}
	TintDescription         = &unison.ThemeColor{Light: unison.Transparent, Dark: unison.Transparent}
	TintPrimaryAttributes   = &unison.ThemeColor{Light: unison.Transparent, Dark: unison.Transparent}
	TintSecondaryAttributes = &unison.ThemeColor{Light: unison.Transparent, Dark: unison.Transparent}
	TintBody                = &unison.ThemeColor{Light: unison.Transparent, Dark: unison.Transparent}
	TintEncumbrance         = &unison.ThemeColor{Light: unison.Transparent, Dark: unison.Transparent}
	TintLifting             = &unison.ThemeColor{Light: unison.Transparent, Dark: unison.Transparent}
	TintDamage              = &unison.ThemeColor{Light: unison.Transparent, Dark: unison.Transparent}
	TintPools               = &unison.ThemeColor{Light: unison.Transparent, Dark: unison.Transparent}
	TintReactions           = &unison.ThemeColor{Light: unison.Transparent, Dark: unison.Transparent}
	TintConditions          = &unison.ThemeColor{Light: unison.Transparent, Dark: unison.Transparent}
	TintMelee               = &unison.ThemeColor{Light: unison.Transparent, Dark: unison.Transparent}
	TintRanged              = &unison.ThemeColor{Light: unison.Transparent, Dark: unison.Transparent}
	TintTraits              = &unison.ThemeColor{Light: unison.Transparent, Dark: unison.Transparent}
	TintSkills              = &unison.ThemeColor{Light: unison.Transparent, Dark: unison.Transparent}
	TintSpells              = &unison.ThemeColor{Light: unison.Transparent, Dark: unison.Transparent}
	TintCarriedEquipment    = &unison.ThemeColor{Light: unison.Transparent, Dark: unison.Transparent}
	TintOtherEquipment      = &unison.ThemeColor{Light: unison.Transparent, Dark: unison.Transparent}
	TintNotes               = &unison.ThemeColor{Light: unison.Transparent, Dark: unison.Transparent}
)

// ThemedColor holds a themed color.
type ThemedColor struct {
	ID    string
	Title string
	Color *unison.ThemeColor
}

// Key implements themeset.Entry.
func (t *ThemedColor) Key() string {
	return t.ID
}

// Value implements themeset.Entry.
func (t *ThemedColor) Value() unison.ThemeColor {
	return *t.Color
}

// SetValue implements themeset.Entry.
func (t *ThemedColor) SetValue(v unison.ThemeColor) {
	*t.Color = v
}

// Colors holds a set of themed colors.
type Colors struct {
	themeset.Set[unison.ThemeColor, *ThemedColor, provider]
}

// provider connects a Colors to the live and factory theme colors.
type provider struct{}

func (provider) Current() []*ThemedColor {
	return Current()
}

func (provider) Factory() []*ThemedColor {
	return Factory()
}

func (provider) Applied() {
	unison.ThemeChanged()
}

type fileData struct {
	Version int    `json:"version"`
	Colors  Colors `json:"colors"`
}

// Current returns the current theme.
func Current() []*ThemedColor {
	once.Do(initialize)
	return current
}

// Factory returns the original theme before any modifications.
func Factory() []*ThemedColor {
	once.Do(initialize)
	return factory
}

func initialize() {
	current = []*ThemedColor{
		{ID: "surface", Title: i18n.Text("Surface"), Color: unison.ThemeSurface},
		{ID: "header", Title: i18n.Text("Header"), Color: Header},
		{ID: "banding", Title: i18n.Text("Banding"), Color: unison.ThemeBanding},
		{ID: "focus", Title: i18n.Text("Focus"), Color: unison.ThemeFocus},
		{ID: "tooltip", Title: i18n.Text("Tooltip"), Color: unison.ThemeTooltip},
		{ID: "error", Title: i18n.Text("Error"), Color: unison.ThemeError},
		{ID: "warning", Title: i18n.Text("Warning"), Color: unison.ThemeWarning},
		{ID: "cursor_fg", Title: i18n.Text("Cursor Foreground"), Color: unison.ThemeCursorForeground},
		{ID: "cursor_bg", Title: i18n.Text("Cursor Background"), Color: unison.ThemeCursorBackground},
		{ID: "tint_portrait", Title: i18n.Text("Portrait"), Color: TintPortrait},
		{ID: "tint_identity", Title: i18n.Text("Identity"), Color: TintIdentity},
		{ID: "tint_misc", Title: i18n.Text("Miscellaneous"), Color: TintMisc},
		{ID: "tint_points", Title: i18n.Text("Points"), Color: TintPoints},
		{ID: "tint_description", Title: i18n.Text("Description"), Color: TintDescription},
		{ID: "tint_primary_attributes", Title: i18n.Text("Primary Attributes"), Color: TintPrimaryAttributes},
		{ID: "tint_secondary_attributes", Title: i18n.Text("Secondary Attributes"), Color: TintSecondaryAttributes},
		{ID: "tint_body", Title: i18n.Text("Body Type"), Color: TintBody},
		{ID: "tint_encumbrance", Title: i18n.Text("Encumbrance, Move & Dodge"), Color: TintEncumbrance},
		{ID: "tint_lifting", Title: i18n.Text("Lifting & Moving Things"), Color: TintLifting},
		{ID: "tint_damage", Title: i18n.Text("Basic Damage"), Color: TintDamage},
		{ID: "tint_pools", Title: i18n.Text("Point Pools"), Color: TintPools},
		{ID: "tint_reactions", Title: i18n.Text("Reactions"), Color: TintReactions},
		{ID: "tint_conditions", Title: i18n.Text("Conditions"), Color: TintConditions},
		{ID: "tint_melee", Title: i18n.Text("Melee Weapons"), Color: TintMelee},
		{ID: "tint_ranged", Title: i18n.Text("Ranged Weapons"), Color: TintRanged},
		{ID: "tint_traits", Title: i18n.Text("Traits"), Color: TintTraits},
		{ID: "tint_skills", Title: i18n.Text("Skills"), Color: TintSkills},
		{ID: "tint_spells", Title: i18n.Text("Spells"), Color: TintSpells},
		{ID: "tint_carried_equipment", Title: i18n.Text("Carried Equipment"), Color: TintCarriedEquipment},
		{ID: "tint_other_equipment", Title: i18n.Text("Other Equipment"), Color: TintOtherEquipment},
		{ID: "tint_notes", Title: i18n.Text("Notes"), Color: TintNotes},
	}
	factory = make([]*ThemedColor, len(current))
	for i, c := range current {
		factory[i] = &ThemedColor{
			ID:    c.ID,
			Title: c.Title,
			Color: &unison.ThemeColor{
				Light: c.Color.Light,
				Dark:  c.Color.Dark,
			},
		}
	}
}

// NewFromFS creates a new set of colors from a file. Any missing values will be filled in with defaults.
func NewFromFS(fileSystem fs.FS, filePath string) (*Colors, error) {
	var data fileData
	if err := jio.LoadFromFile(fileSystem, filePath, &data); err != nil {
		return nil, errs.Wrap(err)
	}
	if err := jio.CheckVersionWithMinimum(data.Version, minimumVersion); err != nil {
		return nil, err
	}
	return &data.Colors, nil
}

// Save writes the Colors to the file as JSON.
func (c *Colors) Save(filePath string) error {
	return jio.SaveToFile(filePath, &fileData{
		Version: jio.CurrentDataVersion,
		Colors:  *c,
	})
}
