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
	"strings"

	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

var aboutWnd *unison.Window

// ShowAbout displays the about box.
func ShowAbout(_ unison.MenuItem) {
	if aboutWnd == nil {
		var err error
		if aboutWnd, err = unison.NewWindow(fmt.Sprintf(i18n.Text("About %s"), xos.AppName),
			unison.NotResizableWindowOption()); err != nil {
			errs.Log(errs.NewWithCause("unable to create about window", err))
			return
		}
		aboutWnd.WillCloseCallback = func() {
			aboutWnd = nil
		}
		SetupMenuBar(aboutWnd)

		content := aboutWnd.Content()
		content.SetLayout(&unison.FlexLayout{
			Columns:  2,
			HSpacing: 16,
			VSpacing: 16,
			HAlign:   align.Middle,
			VAlign:   align.Middle,
		})
		content.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(16)))

		img := unison.NewPanel()
		img.SetSizer(func(_ geom.Size) (minSize, prefSize, maxSize geom.Size) {
			size := geom.NewSize(256, 256)
			return size, size, size
		})
		img.DrawCallback = func(canvas *unison.Canvas, rect geom.Rect) {
			svg.AppIcon.DrawInRectPreservingAspectRatio(canvas, rect, nil, nil)
		}
		content.AddChild(img)

		var version string
		if xos.AppVersion != "" && xos.AppVersion != "0.0" && !strings.HasSuffix(xos.AppVersion, "+dirty") {
			version = "Version **" + xos.AppVersion + "**"
		} else {
			version = "_**Development Version**_"
		}
		md := unison.NewMarkdown(false)
		md.SetContent(fmt.Sprintf(`# GURPS Character Sheet

> %s<br>Build **%s**

%s

**GURPS** is a trademark of Steve Jackson Games, used by permission. All rights reserved.

This product includes copyrighted material from the **GURPS** game, which is used by permission of Steve Jackson Games.`,
			version, xos.BuildNumber, xos.Copyright()), 400)
		content.AddChild(md)

		aboutWnd.Pack()
		r := aboutWnd.FrameRect()
		primary := unison.PrimaryDisplay()
		usable := primary.Usable
		r.X = usable.X + (usable.Width-r.Width)/2
		r.Y = usable.Y + (usable.Height-r.Height)/3
		r = r.Align()
		aboutWnd.SetFrameRect(primary.FitRectOnto(r))
	}
	aboutWnd.ToFront()
}
