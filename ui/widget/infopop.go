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

package widget

import (
	"fmt"
	"strings"

	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xmath"
	"github.com/richardwilkes/unison"
)

// InfoPop holds the data necessary for the info pop control.
type InfoPop struct {
	unison.Panel
	Drawable      *unison.DrawableSVG
	Target        unison.Paneler
	popup         *unison.Panel
	savedOverdraw func(*unison.Canvas, unison.Rect)
	needRestore   bool
}

// NewDefaultInfoPop creates a new InfoPop with the message about mouse wheel scaling.
func NewDefaultInfoPop(target unison.Paneler) *InfoPop {
	p := NewInfoPop()
	p.Target = target
	p.AddScalingHelp()
	return p
}

// NewInfoPop creates a new InfoPop, which is an icon which shows a forced tooltip while the mouse is over it.
func NewInfoPop() *InfoPop {
	p := &InfoPop{}
	p.Self = p
	height := unison.DefaultLabelTheme.Font.Baseline()
	size := unison.NewSize(height, height)
	size.GrowToInteger()
	p.popup = unison.NewTooltipBase()
	p.popup.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	p.SetSizer(func(hint unison.Size) (min, pref, max unison.Size) {
		pref = p.Drawable.LogicalSize()
		pref.GrowToInteger()
		pref.ConstrainForHint(hint)
		return pref, pref, pref
	})
	p.Drawable = &unison.DrawableSVG{
		SVG:  res.InfoSVG,
		Size: size,
	}
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		r := p.ContentRect(false)
		s := p.Drawable.LogicalSize()
		r.X = xmath.Floor(r.X + (r.Width-s.Width)/2)
		r.Y = xmath.Floor(r.Y + (r.Height-s.Height)/2)
		r.Size = s
		gc.Save()
		gc.ClipRect(r, unison.IntersectClipOp, false)
		p.Drawable.DrawInRect(gc, r, nil, unison.DefaultLabelTheme.OnBackgroundInk.Paint(gc, r, unison.Fill))
		gc.Restore()
	}
	p.MouseEnterCallback = func(where unison.Point, mod unison.Modifiers) bool {
		if p.Target != nil && len(p.popup.Children()) != 0 {
			panel := p.Target.AsPanel()
			p.savedOverdraw = panel.DrawOverCallback
			panel.DrawOverCallback = func(gc *unison.Canvas, rect unison.Rect) {
				if p.savedOverdraw != nil {
					gc.Save()
					p.savedOverdraw(gc, rect)
					gc.Restore()
				}
				panel.AddChild(p.popup)
				p.popup.Draw(gc, rect)
				panel.RemoveChild(p.popup)
			}
			p.popup.Pack()
			p.popup.ValidateLayout()
			panel.MarkForRedraw()
			p.needRestore = true
		}
		return true
	}
	p.MouseExitCallback = func() bool {
		if p.needRestore {
			p.needRestore = false
			panel := p.Target.AsPanel()
			panel.DrawOverCallback = p.savedOverdraw
			p.savedOverdraw = nil
			panel.MarkForRedraw()
		}
		return true
	}
	return p
}

// AddScalingHelp adds the help info about scaling.
func (p *InfoPop) AddScalingHelp() {
	p.AddHelpInfo(fmt.Sprintf(i18n.Text(`Holding down the %s key while using
the mouse wheel will change the scale.`), unison.OptionModifier.String()))
}

// AddHelpInfo adds one or more lines of help text.
func (p *InfoPop) AddHelpInfo(text string) {
	for _, str := range strings.Split(text, "\n") {
		label := unison.NewLabel()
		label.LabelTheme = unison.DefaultTooltipTheme.Label
		label.Text = str
		label.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
		p.popup.AddChild(label)
	}
}

// AddKeyBindingInfo adds information about a key binding.
func (p *InfoPop) AddKeyBindingInfo(keyBinding unison.KeyBinding, text string) {
	keyLabel := unison.NewLabel()
	keyLabel.LabelTheme = unison.DefaultTooltipTheme.Label
	keyLabel.OnBackgroundInk = unison.DefaultTooltipTheme.BackgroundInk
	keyLabel.Font = unison.DefaultMenuItemTheme.KeyFont
	keyLabel.Text = keyBinding.String()
	keyLabel.HAlign = unison.MiddleAlignment
	keyLabel.SetLayoutData(&unison.FlexLayoutData{HAlign: unison.FillAlignment})
	keyLabel.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, unison.DefaultTooltipTheme.Label.OnBackgroundInk.Paint(gc, rect, unison.Fill))
		keyLabel.DefaultDraw(gc, rect)
	}
	keyLabel.SetBorder(unison.NewEmptyBorder(unison.NewHorizontalInsets(4)))
	p.popup.AddChild(keyLabel)

	descLabel := unison.NewLabel()
	descLabel.LabelTheme = unison.DefaultTooltipTheme.Label
	descLabel.Text = text
	p.popup.AddChild(descLabel)
}
