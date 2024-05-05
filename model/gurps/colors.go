/*
 * Copyright ©1998-2024 by Richard A. Wilkes. All rights reserved.
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
	"log/slog"
	"sync"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

const (
	minimumColorsVersion = 5
	currentColorsVersion = 5
	colorsTypeKey        = "theme_colors"
)

var (
	colorsOnce    sync.Once
	currentColors []*ThemedColor
	factoryColors []*ThemedColor
)

// ThemedColor holds a themed color.
type ThemedColor struct {
	ID    string
	Title string
	Color *unison.ThemeColor
}

// Colors holds a set of themed colors.
type Colors struct {
	data map[string]*unison.ThemeColor // Just here for serialization
}

type colorsData struct {
	Type    string `json:"type"`
	Version int    `json:"version"`
	Colors
}

// CurrentColors returns the current theme.
func CurrentColors() []*ThemedColor {
	colorsOnce.Do(initColors)
	return currentColors
}

// FactoryColors returns the original theme before any modifications.
func FactoryColors() []*ThemedColor {
	colorsOnce.Do(initColors)
	return factoryColors
}

func initColors() {
	currentColors = []*ThemedColor{
		{ID: "primary", Title: i18n.Text("Primary"), Color: &unison.PrimaryTheme.Primary},
		{ID: "on_primary", Title: i18n.Text("On Primary"), Color: &unison.PrimaryTheme.OnPrimary},
		{ID: "primary_variant", Title: i18n.Text("Primary Variant"), Color: &unison.PrimaryTheme.PrimaryVariant},
		{ID: "secondary", Title: i18n.Text("Secondary"), Color: &unison.PrimaryTheme.Secondary},
		{ID: "on_secondary", Title: i18n.Text("On Secondary"), Color: &unison.PrimaryTheme.OnSecondary},
		{ID: "secondary_variant", Title: i18n.Text("Secondary Variant"), Color: &unison.PrimaryTheme.SecondaryVariant},
		{ID: "tertiary", Title: i18n.Text("Tertiary"), Color: &unison.PrimaryTheme.Tertiary},
		{ID: "on_tertiary", Title: i18n.Text("On Tertiary"), Color: &unison.PrimaryTheme.OnTertiary},
		{ID: "tertiary_variant", Title: i18n.Text("Tertiary Variant"), Color: &unison.PrimaryTheme.TertiaryVariant},
		{ID: "surface", Title: i18n.Text("Surface"), Color: &unison.PrimaryTheme.Surface},
		{ID: "on_surface", Title: i18n.Text("On Surface"), Color: &unison.PrimaryTheme.OnSurface},
		{ID: "surface_above", Title: i18n.Text("Surface Above"), Color: &unison.PrimaryTheme.SurfaceAbove},
		{ID: "surface_below", Title: i18n.Text("Surface Below"), Color: &unison.PrimaryTheme.SurfaceBelow},
		{ID: "error", Title: i18n.Text("Error"), Color: &unison.PrimaryTheme.Error},
		{ID: "on_error", Title: i18n.Text("On Error"), Color: &unison.PrimaryTheme.OnError},
		{ID: "warning", Title: i18n.Text("Warning"), Color: &unison.PrimaryTheme.Warning},
		{ID: "on_warning", Title: i18n.Text("On Warning"), Color: &unison.PrimaryTheme.OnWarning},
		{ID: "outline", Title: i18n.Text("Outline"), Color: &unison.PrimaryTheme.Outline},
		{ID: "outline_variant", Title: i18n.Text("Outline Variant"), Color: &unison.PrimaryTheme.OutlineVariant},
		{ID: "shadow", Title: i18n.Text("Shadow"), Color: &unison.PrimaryTheme.Shadow},
	}
	factoryColors = make([]*ThemedColor, len(currentColors))
	for i, c := range currentColors {
		factoryColors[i] = &ThemedColor{
			ID:    c.ID,
			Title: c.Title,
			Color: &unison.ThemeColor{
				Light: c.Color.Light,
				Dark:  c.Color.Dark,
			},
		}
	}
}

// NewColorsFromFS creates a new set of colors from a file. Any missing values will be filled in with defaults.
func NewColorsFromFS(fileSystem fs.FS, filePath string) (*Colors, error) {
	var current colorsData
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &current); err != nil {
		return nil, errs.Wrap(err)
	}
	if current.Type != colorsTypeKey {
		return nil, errs.New(unexpectedFileDataMsg())
	}
	if current.Version < minimumColorsVersion {
		return nil, errs.New("The theme color data is too old to be used")
	}
	if current.Version > currentColorsVersion {
		return nil, errs.New("The theme color data is too new to be used")
	}
	return &current.Colors, nil
}

// Save writes the Colors to the file as JSON.
func (c *Colors) Save(filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, &colorsData{
		Type:    colorsTypeKey,
		Version: currentColorsVersion,
		Colors:  *c,
	})
}

// MarshalJSON implements json.Marshaler.
func (c *Colors) MarshalJSON() ([]byte, error) {
	c.data = make(map[string]*unison.ThemeColor, len(CurrentColors()))
	for _, one := range CurrentColors() {
		c.data[one.ID] = one.Color
	}
	return json.Marshal(&c.data)
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *Colors) UnmarshalJSON(data []byte) error {
	c.data = nil
	var err error
	toolbox.CallWithHandler(func() {
		err = json.Unmarshal(data, &c.data)
	}, func(e error) {
		err = e
	})
	if err != nil {
		c.data = nil
		errs.LogWithLevel(context.Background(), slog.LevelWarn, slog.Default(),
			errs.NewWithCause("Unable to load theme color data", err))
	}
	if c.data == nil {
		c.data = make(map[string]*unison.ThemeColor, len(CurrentColors()))
	}
	for _, one := range FactoryColors() {
		if _, ok := c.data[one.ID]; !ok {
			clr := *one.Color
			c.data[one.ID] = &clr
		}
	}
	return nil
}

// MakeCurrent applies these colors to the current theme color set and updates all windows.
func (c *Colors) MakeCurrent() {
	for _, one := range CurrentColors() {
		if v, ok := c.data[one.ID]; ok {
			*one.Color = *v
		}
	}
	unison.ThemeChanged()
}

// Reset to factory defaults.
func (c *Colors) Reset() {
	for _, one := range FactoryColors() {
		*c.data[one.ID] = *one.Color
	}
}

// ResetOne resets one color by ID to factory defaults.
func (c *Colors) ResetOne(id string) {
	for _, v := range FactoryColors() {
		if v.ID == id {
			*c.data[id] = *v.Color
			break
		}
	}
}
