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
	_ "embed"
	"fmt"

	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/xmath"
	"github.com/richardwilkes/unison"
)

var (
	//go:embed "images/about-1200x820.png"
	aboutImageData []byte
	aboutWnd       = &aboutWindow{}
)

type aboutWindow struct {
	*unison.Window
	img *unison.Image
}

// ShowAbout displays the about box.
func ShowAbout(_ unison.MenuItem) {
	if aboutWnd.Window == nil {
		if err := aboutWnd.prepare(); err != nil {
			jot.Error(err)
			return
		}
	}
	aboutWnd.ToFront()
}

func (w *aboutWindow) prepare() error {
	var err error
	if w.img == nil {
		if w.img, err = unison.NewImageFromBytes(aboutImageData, 0.5); err != nil {
			return errs.NewWithCause("unable to load about image", err)
		}
	}
	if w.Window, err = unison.NewWindow(fmt.Sprintf(i18n.Text("About %s"), cmdline.AppName),
		unison.NotResizableWindowOption()); err != nil {
		return errs.NewWithCause("unable to create about window", err)
	}
	SetupMenuBar(w.Window)
	content := w.Content()
	content.SetSizer(func(hint unison.Size) (min, pref, max unison.Size) {
		pref = w.img.LogicalSize()
		return pref, pref, pref
	})
	content.SetLayout(nil)
	content.DrawCallback = w.drawContentBackground
	w.WillCloseCallback = func() {
		aboutWnd.Window = nil
	}
	w.Pack()
	r := w.ContentRect()
	usable := unison.PrimaryDisplay().Usable
	r.X = usable.X + (usable.Width-r.Width)/2
	r.Y = usable.Y + (usable.Height-r.Height)/2
	r.Point.Align()
	w.SetContentRect(r)
	return nil
}

func (w *aboutWindow) drawContentBackground(gc *unison.Canvas, _ unison.Rect) {
	r := w.Content().ContentRect(true)
	gc.DrawImageInRect(w.img, r, nil, nil)
	gc.DrawRect(r, unison.NewEvenlySpacedGradient(unison.Point{Y: 0.25}, unison.Point{Y: 1}, 0, 0,
		unison.Transparent, unison.Black).Paint(gc, r, unison.Fill))

	face := unison.MatchFontFace(unison.DefaultSystemFamilyName, unison.NormalFontWeight, unison.StandardSpacing, unison.NoSlant)
	boldFace := unison.MatchFontFace(unison.DefaultSystemFamilyName, unison.BlackFontWeight, unison.StandardSpacing, unison.NoSlant)
	paint := unison.Gray.Paint(gc, unison.Rect{}, unison.Fill)
	dec := &unison.TextDecoration{
		Font:  face.Font(7),
		Paint: paint,
	}
	text := unison.NewText(i18n.Text("This product includes copyrighted material from the "), dec)
	text.AddString("GURPS", &unison.TextDecoration{
		Font:  boldFace.Font(7),
		Paint: paint,
	})
	text.AddString(i18n.Text(" game, which is used by permission of Steve Jackson Games."), dec)
	y := r.Height - 10
	text.Draw(gc, (r.Width-text.Width())/2, y)
	y -= text.Height()

	font := face.Font(8)
	dec = &unison.TextDecoration{
		Font:  font,
		Paint: paint,
	}
	text = unison.NewText("GURPS", &unison.TextDecoration{
		Font:  boldFace.Font(8),
		Paint: paint,
	})
	text.AddString(i18n.Text(" is a trademark of Steve Jackson Games, used by permission. All rights reserved."), dec)
	text.Draw(gc, (r.Width-text.Width())/2, y)
	lineHeight := text.Height()
	y -= lineHeight * 1.5

	paint = unison.RGB(204, 204, 204).Paint(gc, unison.Rect{}, unison.Fill)
	text = unison.NewText(cmdline.Copyright(), &unison.TextDecoration{
		Font:  font,
		Paint: paint,
	})
	text.Draw(gc, (r.Width-text.Width())/2, y)

	buildText := unison.NewText(i18n.Text("Build ")+cmdline.BuildNumber, &unison.TextDecoration{
		Font:  font,
		Paint: paint,
	})
	var t string
	if cmdline.AppVersion != "" && cmdline.AppVersion != "0.0" {
		t = i18n.Text("Version ") + cmdline.AppVersion
	} else {
		t = i18n.Text("Development")
	}
	versionText := unison.NewText(t, &unison.TextDecoration{
		Font: unison.MatchFontFace(unison.DefaultSystemFamilyName, unison.BlackFontWeight,
			unison.StandardSpacing, unison.NoSlant).Font(10),
		Paint: unison.White.Paint(gc, unison.Rect{}, unison.Fill),
	})

	const (
		hMargin = 8
		vMargin = 4
	)
	var backing unison.Rect
	backing.Width = xmath.Max(versionText.Width(), buildText.Width()) + hMargin*2
	backing.Height = versionText.Height() + buildText.Height() + vMargin*2
	backing.X = (r.Width - backing.Width) / 2
	backing.Y = 65
	gc.DrawRoundedRect(backing, 8, 8, unison.Black.SetAlphaIntensity(0.7).Paint(gc, backing, unison.Fill))
	p := unison.Black.Paint(gc, backing, unison.Stroke)
	p.SetStrokeWidth(2)
	gc.DrawRoundedRect(backing, 8, 8, p)

	versionText.Draw(gc, (r.Width-versionText.Width())/2, backing.Y+vMargin+versionText.Baseline())
	buildText.Draw(gc, (r.Width-buildText.Width())/2, backing.Y+vMargin+versionText.Height()+buildText.Baseline())
}
