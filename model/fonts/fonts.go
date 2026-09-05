// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package fonts

import (
	"io/fs"
	"sync"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/themeset"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/slant"
	"github.com/richardwilkes/unison/enums/spacing"
	"github.com/richardwilkes/unison/enums/weight"
)

var (
	_       themeset.Entry[unison.FontDescriptor]                 = &ThemedFont{}
	_       themeset.Provider[unison.FontDescriptor, *ThemedFont] = provider{}
	once    sync.Once
	current []*ThemedFont
	factory []*ThemedFont
)

// Additional fonts over and above what unison provides by default.
var (
	FieldSecondary   = &unison.IndirectFont{Font: unison.FieldFont.Face().Font(unison.FieldFont.Size() - 1)}
	PageFieldPrimary = &unison.IndirectFont{Font: unison.MatchFontFace(unison.DefaultSystemFamilyName,
		weight.Medium, spacing.Standard, slant.Upright).Font(7)}
	PageFieldSecondary = &unison.IndirectFont{Font: unison.MatchFontFace(unison.DefaultSystemFamilyName,
		weight.Regular, spacing.Standard, slant.Upright).Font(6)}
	PageLabelPrimary = &unison.IndirectFont{Font: unison.MatchFontFace(unison.DefaultSystemFamilyName,
		weight.Regular, spacing.Standard, slant.Upright).Font(7)}
	PageLabelSecondary = &unison.IndirectFont{Font: unison.MatchFontFace(unison.DefaultSystemFamilyName,
		weight.Regular, spacing.Standard, slant.Upright).Font(6)}
	PageFooterPrimary = &unison.IndirectFont{Font: unison.MatchFontFace(unison.DefaultSystemFamilyName,
		weight.Medium, spacing.Standard, slant.Upright).Font(6)}
	PageFooterSecondary = &unison.IndirectFont{Font: unison.MatchFontFace(unison.DefaultSystemFamilyName,
		weight.Regular, spacing.Standard, slant.Upright).Font(5)}
	BaseMarkdown = &unison.IndirectFont{Font: unison.LabelFont.Face().Font(unison.LabelFont.Size())}
)

// ThemedFont holds a themed font.
type ThemedFont struct {
	ID    string
	Title string
	Font  *unison.IndirectFont
}

// Key implements themeset.Entry.
func (t *ThemedFont) Key() string {
	return t.ID
}

// Value implements themeset.Entry.
func (t *ThemedFont) Value() unison.FontDescriptor {
	return t.Font.Descriptor()
}

// SetValue implements themeset.Entry.
func (t *ThemedFont) SetValue(v unison.FontDescriptor) {
	t.Font.Font = v.Font()
}

// Fonts holds a set of themed fonts.
type Fonts struct {
	themeset.Set[unison.FontDescriptor, *ThemedFont, provider]
}

// provider connects a Fonts to the live and factory theme fonts.
type provider struct{}

func (provider) Current() []*ThemedFont {
	return CurrentFonts()
}

func (provider) Factory() []*ThemedFont {
	return FactoryFonts()
}

// Applied re-lays out every window, since a font change alters the size of nearly everything.
func (provider) Applied() {
	unison.ThemeChanged()
	for _, wnd := range unison.Windows() {
		wnd.Content().MarkForLayoutRecursively()
		wnd.ValidateLayout()
	}
}

type fileData struct {
	Version int   `json:"version"`
	Fonts   Fonts `json:"fonts"`
}

// CurrentFonts returns the current theme fonts.
func CurrentFonts() []*ThemedFont {
	once.Do(initialize)
	return current
}

// FactoryFonts returns the original theme before any modifications.
func FactoryFonts() []*ThemedFont {
	once.Do(initialize)
	return factory
}

func initialize() {
	current = []*ThemedFont{
		{ID: "system", Title: i18n.Text("System"), Font: unison.SystemFont},
		{ID: "system.emphasized", Title: i18n.Text("System (Emphasized)"), Font: unison.EmphasizedSystemFont},
		{ID: "label", Title: i18n.Text("Label"), Font: unison.LabelFont},
		{ID: "field", Title: i18n.Text("Field"), Font: unison.FieldFont},
		{ID: "field.secondary", Title: i18n.Text("Secondary Fields"), Font: FieldSecondary},
		{ID: "keyboard", Title: i18n.Text("Keyboard"), Font: unison.KeyboardFont},
		{ID: "page.field.primary", Title: i18n.Text("Page Primary Fields"), Font: PageFieldPrimary},
		{ID: "page.field.secondary", Title: i18n.Text("Page Secondary Fields"), Font: PageFieldSecondary},
		{ID: "page.label.primary", Title: i18n.Text("Page Primary Labels"), Font: PageLabelPrimary},
		{ID: "page.label.secondary", Title: i18n.Text("Page Secondary Labels"), Font: PageLabelSecondary},
		{ID: "page.footer.primary", Title: i18n.Text("Page Primary Footer"), Font: PageFooterPrimary},
		{ID: "page.footer.secondary", Title: i18n.Text("Page Secondary Footer"), Font: PageFooterSecondary},
		{ID: "markdown.base", Title: i18n.Text("Base Markdown"), Font: BaseMarkdown},
		{ID: "monospaced", Title: i18n.Text("Monospaced"), Font: unison.MonospacedFont},
	}
	factory = make([]*ThemedFont, len(current))
	for i, c := range current {
		if c.Font.Font == nil {
			errs.Log(errs.New("nil font"), "id", c.ID)
			xos.Exit(1)
		}
		factory[i] = &ThemedFont{
			ID:    c.ID,
			Title: c.Title,
			Font: &unison.IndirectFont{
				Font: c.Font.Font,
			},
		}
	}
	unison.DefaultMarkdownTheme.Font = BaseMarkdown
}

// NewFromFS creates a new set of fonts from a file. Any missing values will be filled in with defaults.
func NewFromFS(fileSystem fs.FS, filePath string) (*Fonts, error) {
	var data fileData
	if err := jio.LoadFromFile(fileSystem, filePath, &data); err != nil {
		return nil, errs.Wrap(err)
	}
	if err := jio.CheckVersion(data.Version); err != nil {
		return nil, err
	}
	return &data.Fonts, nil
}

// Save writes the Fonts to the file as JSON.
func (f *Fonts) Save(filePath string) error {
	return jio.SaveToFile(filePath, &fileData{
		Version: jio.CurrentDataVersion,
		Fonts:   *f,
	})
}
