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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/thememode"
	"github.com/richardwilkes/unison/enums/weight"
)

type colorSettingsDockable struct {
	SettingsDockable
	content *unison.Panel
}

// ShowColorSettings shows the Color settings.
func ShowColorSettings() {
	if Activate(func(d unison.Dockable) bool {
		_, ok := d.AsPanel().Self.(*colorSettingsDockable)
		return ok
	}) {
		return
	}
	d := &colorSettingsDockable{}
	d.Self = d
	d.TabTitle = i18n.Text("Colors")
	d.TabIcon = svg.Settings
	d.Extensions = []string{gurps.ColorSettingsExt}
	d.Loader = d.load
	d.Saver = d.save
	d.Resetter = d.reset
	d.Setup(d.addToStartToolbar, nil, d.initContent)
}

func (d *colorSettingsDockable) initContent(content *unison.Panel) {
	d.content = content
	d.content.SetLayout(&unison.FlexLayout{
		Columns:  4,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	d.fill()
}

func (d *colorSettingsDockable) addToStartToolbar(toolbar *unison.Panel) {
	label := unison.NewLabel()
	label.SetTitle(i18n.Text("Color Mode"))
	toolbar.AddChild(label)
	p := unison.NewPopupMenu[thememode.Enum]()
	for _, mode := range thememode.All {
		p.AddItem(mode)
	}
	p.Select(gurps.GlobalSettings().ThemeMode)
	p.SelectionChangedCallback = func(popup *unison.PopupMenu[thememode.Enum]) {
		if mode, ok := popup.Selected(); ok {
			gurps.GlobalSettings().ThemeMode = mode
			unison.SetThemeMode(mode)
		}
	}
	toolbar.AddChild(p)
}

func (d *colorSettingsDockable) reset() {
	g := gurps.GlobalSettings()
	g.Colors.Reset()
	g.Colors.MakeCurrent()
	d.sync()
}

func (d *colorSettingsDockable) sync() {
	d.content.RemoveAllChildren()
	d.fill()
	d.MarkForRedraw()
}

func (d *colorSettingsDockable) fill() {
	foundTint := false
	d.createHeader(i18n.Text("Theme Colors"), 0, false)
	for _, one := range colors.Current() {
		if !foundTint && strings.HasPrefix(one.ID, "tint_") {
			foundTint = true
			d.createHeader(i18n.Text("Sheet Block Tints"), unison.StdVSpacing*4, false)
			d.createHeader(i18n.Text("(Alpha channel > 0 enables the tint, while 0 turns it off)"), 0, true)
		}
		d.content.AddChild(NewFieldLeadingLabel(one.Title, false))
		d.createColorWellField(one, true)
		d.createColorWellField(one, false)
		d.createResetField(one)
	}
}

func (d *colorSettingsDockable) createHeader(title string, topMargin float32, small bool) {
	label := unison.NewLabel()
	if topMargin > 0 {
		label.SetBorder(unison.NewEmptyBorder(unison.Insets{Top: topMargin}))
	}
	desc := label.Font.Descriptor()
	if small {
		desc.Size *= 0.8
	} else {
		label.Underline = true
	}
	desc.Weight = weight.Bold
	label.Font = desc.Font()
	label.SetTitle(title)
	label.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  4,
		HAlign: align.Middle,
		VAlign: align.Middle,
	})
	d.content.AddChild(label)
}

func (d *colorSettingsDockable) createColorWellField(c *colors.ThemedColor, light bool) {
	w := unison.NewWell()
	w.Mask = unison.ColorWellMask
	if light {
		w.SetInk(c.Color.Light)
		w.Tooltip = newWrappedTooltip(i18n.Text("Light Mode Color"))
		w.InkChangedCallback = func() {
			if clr, ok := w.Ink().(unison.Color); ok {
				c.Color.Light = clr
				unison.ThemeChanged()
			}
		}
	} else {
		w.SetInk(c.Color.Dark)
		w.Tooltip = newWrappedTooltip(i18n.Text("Dark Mode Color"))
		w.InkChangedCallback = func() {
			if clr, ok := w.Ink().(unison.Color); ok {
				c.Color.Dark = clr
				unison.ThemeChanged()
			}
		}
	}
	d.content.AddChild(w)
}

func (d *colorSettingsDockable) createResetField(c *colors.ThemedColor) {
	b := unison.NewSVGButton(svg.Reset)
	b.Tooltip = newWrappedTooltip("Reset this color")
	b.ClickCallback = func() {
		if unison.QuestionDialog(fmt.Sprintf(i18n.Text("Are you sure you want to reset %s?"), c.Title), "") == unison.ModalResponseOK {
			for _, v := range colors.Factory() {
				if v.ID != c.ID {
					continue
				}
				*c.Color = *v.Color
				i := b.Parent().IndexOfChild(b)
				children := b.Parent().Children()
				if w, ok := children[i-2].Self.(*unison.Well); ok {
					w.SetInk(c.Color.Light)
				}
				if w, ok := children[i-1].Self.(*unison.Well); ok {
					w.SetInk(c.Color.Dark)
				}
				unison.ThemeChanged()
				break
			}
		}
	}
	b.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Start,
		VAlign: align.Middle,
	})
	d.content.AddChild(b)
}

func (d *colorSettingsDockable) load(fileSystem fs.FS, filePath string) error {
	s, err := colors.NewFromFS(fileSystem, filePath)
	if err != nil {
		return err
	}
	g := gurps.GlobalSettings()
	g.Colors = *s
	g.Colors.MakeCurrent()
	d.sync()
	return nil
}

func (d *colorSettingsDockable) save(filePath string) error {
	return gurps.GlobalSettings().Colors.Save(filePath)
}

// InstallTintFunc installs a tint function for the given panel and theme color.
func InstallTintFunc(panel unison.Paneler, themeColor *unison.ThemeColor) {
	panel.AsPanel().DrawOverCallback = func(gc *unison.Canvas, rect unison.Rect) {
		c := themeColor.GetColor()
		if c.Invisible() {
			return
		}
		gc.DrawRect(rect, c.SetAlphaIntensity(0.1).Paint(gc, rect, paintstyle.Fill))
	}
}
