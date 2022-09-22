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

package theme

import (
	"context"
	"io/fs"

	"github.com/richardwilkes/gcs/v5/model/gurps/gid"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/unison"
)

const fontsTypeKey = "theme_fonts"

// Additional fonts over and above what unison provides by default.
var (
	FieldSecondaryFont      = &unison.IndirectFont{Font: unison.FieldFont.Face().Font(unison.FieldFont.Size() - 1)}
	PageFieldPrimaryFont    = &unison.IndirectFont{Font: unison.MatchFontFace(unison.DefaultSystemFamilyName, unison.MediumFontWeight, unison.StandardSpacing, unison.NoSlant).Font(7)}
	PageFieldSecondaryFont  = &unison.IndirectFont{Font: unison.MatchFontFace(unison.DefaultSystemFamilyName, unison.NormalFontWeight, unison.StandardSpacing, unison.NoSlant).Font(6)}
	PageLabelPrimaryFont    = &unison.IndirectFont{Font: unison.MatchFontFace(unison.DefaultSystemFamilyName, unison.NormalFontWeight, unison.StandardSpacing, unison.NoSlant).Font(7)}
	PageLabelSecondaryFont  = &unison.IndirectFont{Font: unison.MatchFontFace(unison.DefaultSystemFamilyName, unison.NormalFontWeight, unison.StandardSpacing, unison.NoSlant).Font(6)}
	PageFooterPrimaryFont   = &unison.IndirectFont{Font: unison.MatchFontFace(unison.DefaultSystemFamilyName, unison.MediumFontWeight, unison.StandardSpacing, unison.NoSlant).Font(6)}
	PageFooterSecondaryFont = &unison.IndirectFont{Font: unison.MatchFontFace(unison.DefaultSystemFamilyName, unison.NormalFontWeight, unison.StandardSpacing, unison.NoSlant).Font(5)}
)

var (
	// CurrentFonts holds the current theme fonts.
	CurrentFonts = []*ThemedFont{
		{ID: "system", Title: i18n.Text("System"), Font: unison.SystemFont},
		{ID: "system.emphasized", Title: i18n.Text("System (Emphasized)"), Font: unison.EmphasizedSystemFont},
		{ID: "system.small", Title: i18n.Text("System (Small)"), Font: unison.SmallSystemFont},
		{ID: "system.small.emphasized", Title: i18n.Text("System (Small, Emphasized)"), Font: unison.EmphasizedSmallSystemFont},
		{ID: "label", Title: i18n.Text("Label"), Font: unison.LabelFont},
		{ID: "field", Title: i18n.Text("Field"), Font: unison.FieldFont},
		{ID: "field.secondary", Title: i18n.Text("Secondary Fields"), Font: FieldSecondaryFont},
		{ID: "keyboard", Title: i18n.Text("Keyboard"), Font: unison.KeyboardFont},
		{ID: "page.field.primary", Title: i18n.Text("Page Primary Fields"), Font: PageFieldPrimaryFont},
		{ID: "page.field.secondary", Title: i18n.Text("Page Secondary Fields"), Font: PageFieldSecondaryFont},
		{ID: "page.label.primary", Title: i18n.Text("Page Primary Labels"), Font: PageLabelPrimaryFont},
		{ID: "page.label.secondary", Title: i18n.Text("Page Secondary Labels"), Font: PageLabelSecondaryFont},
		{ID: "page.footer.primary", Title: i18n.Text("Page Primary Footer"), Font: PageFooterPrimaryFont},
		{ID: "page.footer.secondary", Title: i18n.Text("Page Secondary Footer"), Font: PageFooterSecondaryFont},
	}
	// FactoryFonts holds the original theme before any modifications.
	FactoryFonts []*ThemedFont
)

func init() {
	FactoryFonts = make([]*ThemedFont, len(CurrentFonts))
	for i, c := range CurrentFonts {
		if c.Font.Font == nil {
			jot.Fatal(1, i, c)
		}
		FactoryFonts[i] = &ThemedFont{
			ID:    c.ID,
			Title: c.Title,
			Font: &unison.IndirectFont{
				Font: c.Font.Font,
			},
		}
	}
}

// ThemedFont holds a themed font.
type ThemedFont struct {
	ID    string
	Title string
	Font  *unison.IndirectFont
}

// Fonts holds a set of themed fonts.
type Fonts struct {
	data map[string]unison.FontDescriptor // Just here for serialization
}

type fontsData struct {
	Type    string `json:"type"`
	Version int    `json:"version"`
	Fonts
}

// NewFontsFromFS creates a new set of fonts from a file. Any missing values will be filled in with defaults.
func NewFontsFromFS(fileSystem fs.FS, filePath string) (*Fonts, error) {
	var current fontsData
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &current); err != nil {
		type oldFont struct {
			Name  string `json:"name"`
			Style string `json:"style"`
			Size  int    `json:"size"`
		}
		var old struct {
			Version  int                 `json:"version"`
			OldFonts map[string]*oldFont `json:"fonts"`
		}
		if err2 := jio.LoadFromFS(context.Background(), fileSystem, filePath, &old); err2 != nil {
			return nil, err
		}
		current.Type = fontsTypeKey
		current.Version = old.Version
		if old.Version == 1 {
			current.data = make(map[string]unison.FontDescriptor, len(FactoryFonts))
			for _, ff := range FactoryFonts {
				if f, ok := old.OldFonts[ff.ID]; ok {
					current.data[ff.ID] = unison.FontDescriptor{
						Family:  f.Name,
						Size:    float32(f.Size),
						Weight:  unison.WeightFromString(f.Style),
						Spacing: unison.SpacingFromString(f.Style),
						Slant:   unison.SlantFromString(f.Style),
					}
				} else {
					current.data[ff.ID] = ff.Font.Descriptor()
				}
			}
			current.Version = gid.CurrentDataVersion
		}
	}
	// During development of v5, forgot to add the type & version initially, so try and fix that up
	if current.Type == "" && current.Version == 0 {
		current.Type = fontsTypeKey
		current.Version = gid.CurrentDataVersion
	}
	if current.Type != fontsTypeKey {
		return nil, errs.New(gid.UnexpectedFileDataMsg)
	}
	if err := gid.CheckVersion(current.Version); err != nil {
		return nil, err
	}
	return &current.Fonts, nil
}

// Save writes the Fonts to the file as JSON.
func (f *Fonts) Save(filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, &fontsData{
		Type:    fontsTypeKey,
		Version: gid.CurrentDataVersion,
		Fonts:   *f,
	})
}

// MarshalJSON implements json.Marshaler.
func (f *Fonts) MarshalJSON() ([]byte, error) {
	m := make(map[string]bool)
	for _, one := range FactoryFonts {
		m[one.ID] = true
	}
	for k := range f.data {
		if _, ok := m[k]; !ok {
			delete(f.data, k)
		}
	}
	return json.Marshal(&f.data)
}

// UnmarshalJSON implements json.Unmarshaler.
func (f *Fonts) UnmarshalJSON(data []byte) error {
	f.data = make(map[string]unison.FontDescriptor, len(FactoryFonts))
	if err := json.Unmarshal(data, &f.data); err != nil {
		return err
	}
	for _, one := range FactoryFonts {
		if _, ok := f.data[one.ID]; !ok {
			f.data[one.ID] = one.Font.Descriptor()
		}
	}
	return nil
}

// MakeCurrent applies these fonts to the current theme font set and updates all windows.
func (f *Fonts) MakeCurrent() {
	for _, one := range CurrentFonts {
		if v, ok := f.data[one.ID]; ok {
			one.Font.Font = v.Font()
		}
	}
	unison.ThemeChanged()
}

// Reset to factory defaults.
func (f *Fonts) Reset() {
	f.data = make(map[string]unison.FontDescriptor, len(CurrentFonts))
	for _, one := range FactoryFonts {
		f.data[one.ID] = one.Font.Descriptor()
	}
}

// ResetOne resets one font by ID to factory defaults.
func (f *Fonts) ResetOne(id string) {
	for _, v := range FactoryFonts {
		if v.ID == id {
			f.data[id] = v.Font.Descriptor()
			break
		}
	}
}
