// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"fmt"
	"io/fs"

	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

type fontSettingsDockable struct {
	SettingsDockable
	content    *unison.Panel
	fontPanels []*unison.FontPanel
}

// ShowFontSettings shows the Font settings.
func ShowFontSettings() {
	if Activate(func(d unison.Dockable) bool {
		_, ok := d.AsPanel().Self.(*fontSettingsDockable)
		return ok
	}) {
		return
	}
	d := &fontSettingsDockable{}
	d.Self = d
	d.TabTitle = i18n.Text("Fonts")
	d.TabIcon = svg.Settings
	d.Extensions = []string{gurps.FontSettingsExt}
	d.Loader = d.load
	d.Saver = d.save
	d.Resetter = d.reset
	d.Setup(nil, nil, d.initContent)
}

func (d *fontSettingsDockable) initContent(content *unison.Panel) {
	d.content = content
	d.content.SetLayout(&unison.FlexLayout{
		Columns:  3,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	for i, one := range fonts.CurrentFonts() {
		d.content.AddChild(NewFieldTrailingLabel(one.Title, false))
		fp := d.createFontPanel(i)
		d.fontPanels = append(d.fontPanels, fp)
		d.createResetField(i, fp)
	}
	notice := unison.NewLabel()
	notice.Font = unison.SystemFont
	notice.SetTitle(i18n.Text("Changing fonts usually requires restarting the app to see content laid out correctly."))
	notice.SetBorder(unison.NewEmptyBorder(unison.Insets{Top: unison.StdVSpacing * 2}))
	notice.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  3,
		VSpan:  1,
		HAlign: align.Middle,
		VAlign: align.Middle,
	})
	d.content.AddChild(notice)
}

func (d *fontSettingsDockable) createFontPanel(index int) *unison.FontPanel {
	fp := unison.NewFontPanel()
	fp.SetFontDescriptor(fonts.CurrentFonts()[index].Font.Descriptor())
	fp.FontModifiedCallback = func(fd unison.FontDescriptor) {
		d.applyFont(index, fd)
	}
	d.content.AddChild(fp)
	return fp
}

func (d *fontSettingsDockable) createResetField(index int, fp *unison.FontPanel) {
	b := unison.NewSVGButton(svg.Reset)
	b.Tooltip = newWrappedTooltip("Reset this font")
	b.ClickCallback = func() {
		if unison.QuestionDialog(fmt.Sprintf(i18n.Text("Are you sure you want to reset %s?"),
			fonts.CurrentFonts()[index].Title), "") == unison.ModalResponseOK {
			for _, v := range fonts.FactoryFonts() {
				if v.ID != fonts.CurrentFonts()[index].ID {
					continue
				}
				fp.SetFontDescriptor(v.Font.Descriptor())
				break
			}
		}
	}
	b.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Middle,
		VAlign: align.Middle,
	})
	d.content.AddChild(b)
}

func (d *fontSettingsDockable) reset() {
	g := gurps.GlobalSettings()
	g.Fonts.Reset()
	g.Fonts.MakeCurrent()
	d.sync()
}

func (d *fontSettingsDockable) sync() {
	changed := false
	for i, fp := range d.fontPanels {
		saved := fp.FontModifiedCallback
		fp.FontModifiedCallback = func(_ unison.FontDescriptor) { changed = true }
		fp.SetFontDescriptor(fonts.CurrentFonts()[i].Font.Descriptor())
		fp.FontModifiedCallback = saved
	}
	if changed {
		unison.ThemeChanged()
	}
}

func (d *fontSettingsDockable) applyFont(index int, fd unison.FontDescriptor) {
	f := fonts.CurrentFonts()[index].Font
	if f.Descriptor() != fd {
		f.Font = fd.Font()
		unison.ThemeChanged()
	}
}

func (d *fontSettingsDockable) load(fileSystem fs.FS, filePath string) error {
	s, err := fonts.NewFromFS(fileSystem, filePath)
	if err != nil {
		return err
	}
	g := gurps.GlobalSettings()
	g.Fonts = *s
	g.Fonts.MakeCurrent()
	d.sync()
	return nil
}

func (d *fontSettingsDockable) save(filePath string) error {
	return gurps.GlobalSettings().Fonts.Save(filePath)
}
