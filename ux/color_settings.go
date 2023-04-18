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

package ux

import (
	"fmt"
	"io/fs"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
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
		Columns:  8,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	d.fill()
}

func (d *colorSettingsDockable) addToStartToolbar(toolbar *unison.Panel) {
	label := unison.NewLabel()
	label.Text = i18n.Text("Color Mode")
	toolbar.AddChild(label)
	p := unison.NewPopupMenu[unison.ColorMode]()
	for _, mode := range unison.AllColorModes {
		p.AddItem(mode)
	}
	p.Select(gurps.GlobalSettings().ColorMode)
	p.SelectionChangedCallback = func(popup *unison.PopupMenu[unison.ColorMode]) {
		if mode, ok := popup.Selected(); ok {
			gurps.GlobalSettings().ColorMode = mode
			unison.SetColorMode(mode)
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
	for i, one := range gurps.CurrentColors() {
		if i%2 == 0 {
			d.content.AddChild(NewFieldLeadingLabel(one.Title))
		} else {
			d.content.AddChild(NewFieldInteriorLeadingLabel(one.Title))
		}
		d.createColorWellField(one, true)
		d.createColorWellField(one, false)
		d.createResetField(one)
	}
}

func (d *colorSettingsDockable) createColorWellField(c *gurps.ThemedColor, light bool) {
	w := unison.NewWell()
	w.Mask = unison.ColorWellMask
	if light {
		w.SetInk(c.Color.Light)
		w.Tooltip = unison.NewTooltipWithText(i18n.Text("The color to use when light mode is enabled"))
		w.InkChangedCallback = func() {
			if clr, ok := w.Ink().(unison.Color); ok {
				c.Color.Light = clr
				unison.ThemeChanged()
			}
		}
	} else {
		w.SetInk(c.Color.Dark)
		w.Tooltip = unison.NewTooltipWithText(i18n.Text("The color to use when dark mode is enabled"))
		w.InkChangedCallback = func() {
			if clr, ok := w.Ink().(unison.Color); ok {
				c.Color.Dark = clr
				unison.ThemeChanged()
			}
		}
	}
	d.content.AddChild(w)
}

func (d *colorSettingsDockable) createResetField(c *gurps.ThemedColor) {
	b := unison.NewSVGButton(svg.Reset)
	b.Tooltip = unison.NewTooltipWithText("Reset this color")
	b.ClickCallback = func() {
		if unison.QuestionDialog(fmt.Sprintf(i18n.Text("Are you sure you want to reset %s?"), c.Title), "") == unison.ModalResponseOK {
			for _, v := range gurps.FactoryColors() {
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
		HAlign: unison.MiddleAlignment,
		VAlign: unison.MiddleAlignment,
	})
	d.content.AddChild(b)
}

func (d *colorSettingsDockable) load(fileSystem fs.FS, filePath string) error {
	s, err := gurps.NewColorsFromFS(fileSystem, filePath)
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
