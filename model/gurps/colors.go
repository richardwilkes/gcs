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
	"log/slog"
	"sync"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/errs"
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

// Additional theme colors
var (
	ThemeHeader   = &unison.ThemeColor{Light: unison.RGB(43, 43, 43), Dark: unison.RGB(64, 64, 64)}
	OnThemeHeader = ThemeHeader.DeriveOn()
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
		{ID: "surface", Title: "Surface", Color: unison.ThemeSurface},
		{ID: "header", Title: "Header", Color: ThemeHeader},
		{ID: "focus", Title: "Focus", Color: unison.ThemeFocus},
		{ID: "tooltip", Title: "Tooltip", Color: unison.ThemeTooltip},
		{ID: "error", Title: "Error", Color: unison.ThemeError},
		{ID: "warning", Title: "Warning", Color: unison.ThemeWarning},
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
	current := CurrentColors()
	c.data = make(map[string]*unison.ThemeColor, len(current))
	for _, one := range current {
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
	factory := FactoryColors()
	if c.data == nil {
		c.data = make(map[string]*unison.ThemeColor, len(factory))
	}
	for _, one := range factory {
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
	factory := FactoryColors()
	if c.data == nil {
		c.data = make(map[string]*unison.ThemeColor, len(factory))
	}
	for _, one := range factory {
		if v, ok := c.data[one.ID]; ok {
			*v = *one.Color
		} else {
			clr := *one.Color
			c.data[one.ID] = &clr
		}
	}
}

// ResetOne resets one color by ID to factory defaults.
func (c *Colors) ResetOne(id string) {
	factory := FactoryColors()
	if c.data == nil {
		c.data = make(map[string]*unison.ThemeColor, len(factory))
	}
	for _, one := range factory {
		if one.ID == id {
			if v, ok := c.data[id]; ok {
				*v = *one.Color
			} else {
				clr := *one.Color
				c.data[id] = &clr
			}
			break
		}
	}
}
