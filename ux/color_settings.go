// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"io/fs"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/thememode"
)

type colorSettingsDockable struct {
	SettingsDockable
	content *unison.Panel
}

// ShowColorSettings shows the Color settings.
func ShowColorSettings() {
	if ActivateDockableOfType[*colorSettingsDockable](nil) {
		return
	}
	d := &colorSettingsDockable{}
	// The sections wrap to the view's width, but a view too narrow for even one of them scrolls sideways.
	d.ContentWidthBehavior = behavior.HintedFill
	d.initSettings(d, &settingsSpec{
		title:             i18n.Text("Colors"),
		ext:               gurps.ColorSettingsExt,
		loader:            d.load,
		saver:             d.save,
		resetter:          d.reset,
		addToStartToolbar: d.addToStartToolbar,
		initContent:       d.initContent,
	})
}

// initContent lays the sections out in a flow, so that they sit side by side when the view is wide enough for both and
// stack when it is not.
func (d *colorSettingsDockable) initContent(content *unison.Panel) {
	content.SetLayout(&unison.FlowLayout{
		HSpacing: unison.StdHSpacing * 4,
		VSpacing: unison.DefaultLabelTheme.Font.LineHeight(),
	})
	d.content = content
	d.fill()
}

func (d *colorSettingsDockable) addToStartToolbar(toolbar *unison.Panel) {
	label := unison.NewLabel()
	label.SetTitle(i18n.Text("Color Mode"))
	toolbar.AddChild(label)
	toolbar.AddChild(newPopupMenu(thememode.All, gurps.GlobalSettings().ThemeMode, func(mode thememode.Enum) {
		gurps.GlobalSettings().ThemeMode = mode
		unison.SetThemeMode(mode)
	}))
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
	section := d.createSection(i18n.Text("Theme Colors"), "")
	inTints := false
	for _, one := range colors.Current() {
		if !inTints && strings.HasPrefix(one.ID, "tint_") {
			inTints = true
			section = d.createSection(i18n.Text("Sheet Block Tints"),
				i18n.Text("Alpha channel > 0 enables the tint, while 0 turns it off"))
		}
		section.AddChild(NewFieldLeadingLabel(one.Title, false))
		light := d.createColorWellField(section, one, true)
		dark := d.createColorWellField(section, one, false)
		d.createResetField(section, one, light, dark)
	}
}

// createSection adds a section to the content for rows of colors, each a name, a well apiece for the color's light and
// dark mode variants and a reset button, with the mode names over the wells and, when note is not empty, the note
// under the title.
func (d *colorSettingsDockable) createSection(title, note string) *unison.Panel {
	section := newSettingsSection(d.content, title, 4)
	if note != "" {
		label := unison.NewLabel()
		label.Font = unison.SystemFont
		label.SetTitle(note)
		label.SetLayoutData(&unison.FlexLayoutData{HSpan: 4})
		section.AddChild(label)
	}
	section.AddChild(unison.NewPanel()) // Nothing heads the name column...
	for _, mode := range []thememode.Enum{thememode.Light, thememode.Dark} {
		label := unison.NewLabel()
		label.SetTitle(mode.String())
		label.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Middle})
		section.AddChild(label)
	}
	section.AddChild(unison.NewPanel()) // ...or the reset column.
	return section
}

// createColorWellField adds the well that edits the light or dark mode variant of the color. Each well is named for
// both the color and the mode, since only the first of the two follows the label that names the row, and it would take
// the label's text as its name while the second went without one.
func (d *colorSettingsDockable) createColorWellField(section *unison.Panel, c *colors.ThemedColor, light bool) *unison.Well {
	w := unison.NewWell()
	w.Mask = unison.ColorWellMask
	if light {
		w.SetInk(c.Color.Light)
		w.Tooltip = newWrappedTooltip(i18n.Text("Light Mode Color"))
		w.Accessibility.Name = i18n.Text("%s (Light Mode)", c.Title)
		w.InkChangedCallback = func() {
			if clr, ok := w.Ink().(unison.Color); ok {
				c.Color.Light = clr
				unison.ThemeChanged()
			}
		}
	} else {
		w.SetInk(c.Color.Dark)
		w.Tooltip = newWrappedTooltip(i18n.Text("Dark Mode Color"))
		w.Accessibility.Name = i18n.Text("%s (Dark Mode)", c.Title)
		w.InkChangedCallback = func() {
			if clr, ok := w.Ink().(unison.Color); ok {
				c.Color.Dark = clr
				unison.ThemeChanged()
			}
		}
	}
	w.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Middle,
		VAlign: align.Middle,
	})
	section.AddChild(w)
	return w
}

func (d *colorSettingsDockable) createResetField(section *unison.Panel, c *colors.ThemedColor, light, dark *unison.Well) {
	b := unison.NewSVGButton(svg.Reset)
	b.Tooltip = newWrappedTooltip(i18n.Text("Reset this color"))
	b.ClickCallback = func() {
		if unison.QuestionDialog(i18n.Text("Are you sure you want to reset %s?", c.Title), "") == unison.ModalResponseOK {
			for _, v := range colors.Factory() {
				if v.ID != c.ID {
					continue
				}
				*c.Color = *v.Color
				light.SetInk(c.Color.Light)
				dark.SetInk(c.Color.Dark)
				unison.ThemeChanged()
				break
			}
		}
	}
	b.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Start,
		VAlign: align.Middle,
	})
	section.AddChild(b)
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
	g := gurps.GlobalSettings()
	g.Colors.CaptureCurrent() // The wells edit the live colors, so pull those edits in before exporting them.
	return g.Colors.Save(filePath)
}

// InstallTintFunc installs a tint function for the given panel and theme color.
func InstallTintFunc(panel unison.Paneler, themeColor *unison.ThemeColor) {
	panel.AsPanel().DrawOverCallback = func(gc *unison.Canvas, rect geom.Rect) {
		c := themeColor.GetColor()
		if c.Invisible() {
			return
		}
		gc.DrawRect(rect, c.SetAlphaIntensity(0.1).Paint(gc, rect, paintstyle.Fill))
	}
}
