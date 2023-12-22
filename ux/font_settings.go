/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ux

import (
	"fmt"
	"io/fs"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

type fontSettingsDockable struct {
	SettingsDockable
	content         *unison.Panel
	allFaces        []unison.FontFaceDescriptor
	monospacedFaces []unison.FontFaceDescriptor
	noUpdate        bool
}

// ShowFontSettings shows the Font settings.
func ShowFontSettings() {
	if Activate(func(d unison.Dockable) bool {
		_, ok := d.AsPanel().Self.(*fontSettingsDockable)
		return ok
	}) {
		return
	}
	all, monospaced := unison.AllFontFaces()
	d := &fontSettingsDockable{allFaces: all, monospacedFaces: monospaced}
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
		Columns:  4,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	d.fill()
}

func (d *fontSettingsDockable) reset() {
	g := gurps.GlobalSettings()
	g.Fonts.Reset()
	g.Fonts.MakeCurrent()
	d.sync()
}

func (d *fontSettingsDockable) sync() {
	d.content.RemoveAllChildren()
	d.fill()
	d.MarkForRedraw()
}

func (d *fontSettingsDockable) fill() {
	for i, one := range gurps.CurrentFonts() {
		if i%2 == 0 {
			d.content.AddChild(NewFieldLeadingLabel(one.Title, false))
		} else {
			d.content.AddChild(NewFieldInteriorLeadingLabel(one.Title, false))
		}
		d.createFaceField(i)
		d.createSizeField(i)
		d.createResetField(i)
	}
	notice := unison.NewLabel()
	notice.Text = "Changing fonts usually requires restarting the app to see content laid out correctly."
	notice.Font = unison.SystemFont
	notice.SetBorder(unison.NewEmptyBorder(unison.Insets{Top: unison.StdVSpacing * 2}))
	notice.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  4,
		VSpan:  1,
		HAlign: align.Middle,
		VAlign: align.Middle,
	})
	d.content.AddChild(notice)
}

func (d *fontSettingsDockable) createFaceField(index int) {
	p := unison.NewPopupMenu[unison.FontFaceDescriptor]()
	var list []unison.FontFaceDescriptor
	if gurps.CurrentFonts()[index].ID == "monospaced" {
		list = d.monospacedFaces
	} else {
		list = d.allFaces
	}
	for _, ffd := range list {
		p.AddItem(ffd)
	}
	p.Select(gurps.CurrentFonts()[index].Font.Descriptor().FontFaceDescriptor)
	p.SelectionChangedCallback = func(popup *unison.PopupMenu[unison.FontFaceDescriptor]) {
		if d.noUpdate {
			return
		}
		if ffd, ok := popup.Selected(); ok {
			fd2 := gurps.CurrentFonts()[index].Font.Descriptor()
			fd2.FontFaceDescriptor = ffd
			d.applyFont(index, fd2)
		}
	}
	d.content.AddChild(p)
}

func (d *fontSettingsDockable) createSizeField(index int) {
	field := NewDecimalField(nil, "", i18n.Text("Font Size"),
		func() fxp.Int { return fxp.From(gurps.CurrentFonts()[index].Font.Size()) },
		func(v fxp.Int) {
			if !d.noUpdate {
				fd := gurps.CurrentFonts()[index].Font.Descriptor()
				fd.Size = fxp.As[float32](v)
				d.applyFont(index, fd)
			}
		}, fxp.One, fxp.From(999), false, false)
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	d.content.AddChild(field)
}

func (d *fontSettingsDockable) createResetField(index int) {
	b := unison.NewSVGButton(svg.Reset)
	b.Tooltip = newWrappedTooltip("Reset this font")
	b.ClickCallback = func() {
		if unison.QuestionDialog(fmt.Sprintf(i18n.Text("Are you sure you want to reset %s?"),
			gurps.CurrentFonts()[index].Title), "") == unison.ModalResponseOK {
			for _, v := range gurps.FactoryFonts() {
				if v.ID != gurps.CurrentFonts()[index].ID {
					continue
				}
				d.applyFont(index, v.Font.Descriptor())
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

func (d *fontSettingsDockable) applyFont(index int, fd unison.FontDescriptor) {
	gurps.CurrentFonts()[index].Font.Font = fd.Font()
	children := d.content.Children()
	i := index * 4
	fd = gurps.CurrentFonts()[index].Font.Descriptor()
	d.noUpdate = true
	if p, ok := children[i+1].Self.(*unison.PopupMenu[unison.FontFaceDescriptor]); ok {
		p.Select(fd.FontFaceDescriptor)
	}
	if nf, ok := children[i+2].Self.(*DecimalField); ok {
		nf.SetText(fxp.From(fd.Size).String())
	}
	d.noUpdate = false
	unison.ThemeChanged()
}

func (d *fontSettingsDockable) load(fileSystem fs.FS, filePath string) error {
	s, err := gurps.NewFontsFromFS(fileSystem, filePath)
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
