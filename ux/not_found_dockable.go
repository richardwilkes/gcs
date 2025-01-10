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
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

var (
	_ unison.Dockable  = &notFoundDockable{}
	_ unison.TabCloser = &notFoundDockable{}
)

type notFoundDockable struct {
	unison.Panel
	msg string
}

func newNotFoundDockable(msg string) unison.Dockable {
	d := &notFoundDockable{
		msg: msg,
	}
	d.Self = d
	d.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   align.Middle,
		VAlign:   align.Middle,
	})
	label := unison.NewLabel()
	label.SetTitle(msg)
	d.AddChild(label)
	return d
}

func (d *notFoundDockable) Title() string {
	return i18n.Text("Not Found")
}

func (d *notFoundDockable) TitleIcon(suggestedSize unison.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  unison.DocumentSVG,
		Size: suggestedSize,
	}
}

func (d *notFoundDockable) Tooltip() string {
	return ""
}

func (d *notFoundDockable) Modified() bool {
	return false
}

func (d *notFoundDockable) MayAttemptClose() bool {
	return true
}

func (d *notFoundDockable) AttemptClose() bool {
	return AttemptCloseForDockable(d)
}
