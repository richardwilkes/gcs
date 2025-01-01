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
	_ "embed"
	"fmt"

	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/slant"
	"github.com/richardwilkes/unison/enums/spacing"
	"github.com/richardwilkes/unison/enums/weight"
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
			errs.Log(err)
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
	content.SetSizer(func(_ unison.Size) (minSize, prefSize, maxSize unison.Size) {
		prefSize = w.img.LogicalSize()
		return prefSize, prefSize, prefSize
	})
	content.SetLayout(nil)
	content.DrawCallback = w.drawContentBackground
	w.WillCloseCallback = func() {
		aboutWnd.Window = nil
	}
	w.Pack()
	r := w.FrameRect()
	primary := unison.PrimaryDisplay()
	usable := primary.Usable
	r.X = usable.X + (usable.Width-r.Width)/2
	r.Y = usable.Y + (usable.Height-r.Height)/3
	r = r.Align()
	w.SetFrameRect(primary.FitRectOnto(r))
	return nil
}

func (w *aboutWindow) drawContentBackground(gc *unison.Canvas, _ unison.Rect) {
	r := w.Content().ContentRect(true)
	gc.DrawImageInRect(w.img, r, nil, nil)
	gc.DrawRect(r, unison.NewEvenlySpacedGradient(unison.Point{Y: 0.25}, unison.Point{Y: 1}, 0, 0,
		unison.Transparent, unison.Black).Paint(gc, r, paintstyle.Fill))

	face := unison.MatchFontFace(unison.DefaultSystemFamilyName, weight.Regular, spacing.Standard, slant.Upright)
	boldFace := unison.MatchFontFace(unison.DefaultSystemFamilyName, weight.Black, spacing.Standard, slant.Upright)
	dec := &unison.TextDecoration{
		Font:            face.Font(7),
		OnBackgroundInk: unison.Gray,
	}
	text := unison.NewText(i18n.Text("This product includes copyrighted material from the "), dec)
	text.AddString("GURPS", &unison.TextDecoration{
		Font:            boldFace.Font(7),
		OnBackgroundInk: unison.Gray,
	})
	text.AddString(i18n.Text(" game, which is used by permission of Steve Jackson Games."), dec)
	y := r.Height - 10
	text.Draw(gc, (r.Width-text.Width())/2, y)
	y -= text.Height()

	font := face.Font(8)
	dec = &unison.TextDecoration{
		Font:            font,
		OnBackgroundInk: unison.Gray,
	}
	text = unison.NewText("GURPS", &unison.TextDecoration{
		Font:            boldFace.Font(8),
		OnBackgroundInk: unison.Gray,
	})
	text.AddString(i18n.Text(" is a trademark of Steve Jackson Games, used by permission. All rights reserved."), dec)
	text.Draw(gc, (r.Width-text.Width())/2, y)
	lineHeight := text.Height()
	y -= lineHeight * 1.5

	fg := unison.RGB(204, 204, 204)
	text = unison.NewText(cmdline.Copyright(), &unison.TextDecoration{
		Font:            font,
		OnBackgroundInk: fg,
	})
	text.Draw(gc, (r.Width-text.Width())/2, y)

	buildText := unison.NewText(i18n.Text("Build ")+cmdline.BuildNumber, &unison.TextDecoration{
		Font:            font,
		OnBackgroundInk: fg,
	})
	var t string
	if cmdline.AppVersion != "" && cmdline.AppVersion != "0.0" {
		t = i18n.Text("Version ") + cmdline.AppVersion
	} else {
		t = i18n.Text("Development")
	}
	versionText := unison.NewText(t, &unison.TextDecoration{
		Font:            unison.MatchFontFace(unison.DefaultSystemFamilyName, weight.Black, spacing.Standard, slant.Upright).Font(10),
		OnBackgroundInk: unison.White,
	})

	const (
		hMargin = 8
		vMargin = 4
	)
	width := max(versionText.Width(), buildText.Width()) + hMargin*2
	backing := unison.NewRect((r.Width-width)/2, 65, width, versionText.Height()+buildText.Height()+vMargin*2)
	gc.DrawRoundedRect(backing, 8, 8, unison.Black.SetAlphaIntensity(0.7).Paint(gc, backing, paintstyle.Fill))
	p := unison.Black.Paint(gc, backing, paintstyle.Stroke)
	p.SetStrokeWidth(2)
	gc.DrawRoundedRect(backing, 8, 8, p)

	versionText.Draw(gc, (r.Width-versionText.Width())/2, backing.Y+vMargin+versionText.Baseline())
	buildText.Draw(gc, (r.Width-buildText.Width())/2, backing.Y+vMargin+versionText.Height()+buildText.Baseline())
}
