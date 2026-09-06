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
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
)

// showListQuestionDialog puts up a modal question dialog whose content is list, wrapped in a bordered scroll panel
// that takes whatever room the dialog offers, beneath a label showing header and then any extraHeaders, in order. It
// reports whether the dialog was accepted.
func showListQuestionDialog(header string, list unison.Paneler, extraHeaders ...*unison.Label) bool {
	return unison.QuestionDialogWithPanel(newListQuestionPanel(header, list, extraHeaders...)) == unison.ModalResponseOK
}

// newListQuestionPanel builds the content panel showListQuestionDialog shows: the header labels stacked over the list's
// scroll panel in a single column, with the scroll panel the only child that grows to fill the dialog.
func newListQuestionPanel(header string, list unison.Paneler, extraHeaders ...*unison.Label) *unison.Panel {
	scroll := unison.NewScrollPanel()
	scroll.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.NewUniformInsets(1), false))
	scroll.SetContent(list, behavior.Fill, behavior.Fill)
	scroll.BackgroundInk = unison.ThemeSurface
	scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
	})
	label := unison.NewLabel()
	label.SetTitle(header)
	panel.AddChild(label)
	for _, extra := range extraHeaders {
		panel.AddChild(extra)
	}
	panel.AddChild(scroll)
	return panel
}
