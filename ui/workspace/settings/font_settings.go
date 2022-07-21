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

package settings

import (
	"fmt"
	"io/fs"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/gcs/v5/model/theme"
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/gcs/v5/ui/workspace"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

type fontSettingsDockable struct {
	Dockable
	content  *unison.Panel
	noUpdate bool
}

// ShowFontSettings shows the Font settings.
func ShowFontSettings() {
	ws, dc, found := workspace.Activate(func(d unison.Dockable) bool {
		_, ok := d.(*fontSettingsDockable)
		return ok
	})
	if !found && ws != nil {
		d := &fontSettingsDockable{}
		d.Self = d
		d.TabTitle = i18n.Text("Fonts")
		d.Extensions = []string{".fonts"}
		d.Loader = d.load
		d.Saver = d.save
		d.Resetter = d.reset
		d.Setup(ws, dc, nil, nil, d.initContent)
	}
}

func (d *fontSettingsDockable) initContent(content *unison.Panel) {
	d.content = content
	d.content.SetLayout(&unison.FlexLayout{
		Columns:  7,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	d.fill()
}

func (d *fontSettingsDockable) reset() {
	g := settings.Global()
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
	for i, one := range theme.CurrentFonts {
		if i%2 == 0 {
			d.content.AddChild(widget.NewFieldLeadingLabel(one.Title))
		} else {
			d.content.AddChild(widget.NewFieldInteriorLeadingLabel(one.Title))
		}
		d.createFamilyField(i)
		d.createSizeField(i)
		d.createWeightField(i)
		d.createSpacingField(i)
		d.createSlantField(i)
		d.createResetField(i)
	}
	notice := unison.NewLabel()
	notice.Text = "Changing fonts usually requires restarting the app to see content laid out correctly."
	notice.Font = unison.SystemFont
	notice.SetBorder(unison.NewEmptyBorder(unison.Insets{Top: unison.StdVSpacing * 2}))
	notice.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  7,
		VSpan:  1,
		HAlign: unison.MiddleAlignment,
		VAlign: unison.MiddleAlignment,
	})
	d.content.AddChild(notice)
}

func (d *fontSettingsDockable) createFamilyField(index int) {
	p := unison.NewPopupMenu[string]()
	for _, family := range unison.FontFamilies() {
		p.AddItem(family)
	}
	p.Select(theme.CurrentFonts[index].Font.Descriptor().Family)
	p.SelectionCallback = func(_ int, family string) {
		if d.noUpdate {
			return
		}
		fd := theme.CurrentFonts[index].Font.Descriptor()
		fd.Family = family
		d.applyFont(index, fd)
	}
	d.content.AddChild(p)
}

func (d *fontSettingsDockable) createSizeField(index int) {
	field := widget.NewDecimalField(nil, "", i18n.Text("Font Size"),
		func() fxp.Int { return fxp.From(theme.CurrentFonts[index].Font.Size()) },
		func(v fxp.Int) {
			if !d.noUpdate {
				fd := theme.CurrentFonts[index].Font.Descriptor()
				fd.Size = fxp.As[float32](v)
				d.applyFont(index, fd)
			}
		}, fxp.One, fxp.From(999), false, false)
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
	})
	d.content.AddChild(field)
}

func (d *fontSettingsDockable) createWeightField(index int) {
	p := unison.NewPopupMenu[unison.FontWeight]()
	for _, s := range unison.FontWeights {
		p.AddItem(s)
	}
	p.Select(theme.CurrentFonts[index].Font.Descriptor().Weight)
	p.SelectionCallback = func(_ int, item unison.FontWeight) {
		if d.noUpdate {
			return
		}
		fd := theme.CurrentFonts[index].Font.Descriptor()
		fd.Weight = item
		d.applyFont(index, fd)
	}
	d.content.AddChild(p)
}

func (d *fontSettingsDockable) createSpacingField(index int) {
	p := unison.NewPopupMenu[unison.FontSpacing]()
	for _, s := range unison.Spacings {
		p.AddItem(s)
	}
	p.Select(theme.CurrentFonts[index].Font.Descriptor().Spacing)
	p.SelectionCallback = func(_ int, item unison.FontSpacing) {
		if d.noUpdate {
			return
		}
		fd := theme.CurrentFonts[index].Font.Descriptor()
		fd.Spacing = item
		d.applyFont(index, fd)
	}
	d.content.AddChild(p)
}

func (d *fontSettingsDockable) createSlantField(index int) {
	p := unison.NewPopupMenu[unison.FontSlant]()
	for _, s := range unison.Slants {
		p.AddItem(s)
	}
	p.Select(theme.CurrentFonts[index].Font.Descriptor().Slant)
	p.SelectionCallback = func(_ int, item unison.FontSlant) {
		if d.noUpdate {
			return
		}
		fd := theme.CurrentFonts[index].Font.Descriptor()
		fd.Slant = item
		d.applyFont(index, fd)
	}
	d.content.AddChild(p)
}

func (d *fontSettingsDockable) createResetField(index int) {
	b := unison.NewSVGButton(res.ResetSVG)
	b.Tooltip = unison.NewTooltipWithText("Reset this font")
	b.ClickCallback = func() {
		if unison.QuestionDialog(fmt.Sprintf(i18n.Text("Are you sure you want to reset %s?"),
			theme.CurrentFonts[index].Title), "") == unison.ModalResponseOK {
			for _, v := range theme.FactoryFonts {
				if v.ID != theme.CurrentFonts[index].ID {
					continue
				}
				d.applyFont(index, v.Font.Descriptor())
				break
			}
		}
	}
	b.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.MiddleAlignment,
		VAlign: unison.MiddleAlignment,
	})
	d.content.AddChild(b)
}

func (d *fontSettingsDockable) applyFont(index int, fd unison.FontDescriptor) {
	theme.CurrentFonts[index].Font.Font = fd.Font()
	children := d.content.Children()
	i := index * 7
	fd = theme.CurrentFonts[index].Font.Descriptor()
	d.noUpdate = true
	if p, ok := children[i+1].Self.(*unison.PopupMenu[string]); ok {
		p.Select(fd.Family)
	}
	if nf, ok := children[i+2].Self.(*widget.DecimalField); ok {
		nf.SetText(fxp.From(fd.Size).String())
	}
	if p, ok := children[i+3].Self.(*unison.PopupMenu[unison.FontWeight]); ok {
		p.Select(fd.Weight)
	}
	if p, ok := children[i+4].Self.(*unison.PopupMenu[unison.FontSpacing]); ok {
		p.Select(fd.Spacing)
	}
	if p, ok := children[i+5].Self.(*unison.PopupMenu[unison.FontSlant]); ok {
		p.Select(fd.Slant)
	}
	d.noUpdate = false
	unison.ThemeChanged()
}

func (d *fontSettingsDockable) load(fileSystem fs.FS, filePath string) error {
	s, err := theme.NewFontsFromFS(fileSystem, filePath)
	if err != nil {
		return err
	}
	g := settings.Global()
	g.Fonts = *s
	g.Fonts.MakeCurrent()
	d.sync()
	return nil
}

func (d *fontSettingsDockable) save(filePath string) error {
	return settings.Global().Fonts.Save(filePath)
}
