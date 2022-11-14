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
	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/toolbox/xmath"
	"github.com/richardwilkes/unison"
)

var _ unison.Layout = &portraitLayout{}

type portraitLayout struct {
	portrait *PortraitPanel
	rest     *unison.Panel
}

func (p *portraitLayout) LayoutSizes(_ *unison.Panel, hint unison.Size) (min, pref, max unison.Size) {
	var width, height float32
	insets := p.portrait.Border().Insets()
	_, pref, _ = p.rest.Sizes(unison.Size{})
	if height -= insets.Top + insets.Bottom; height > 0 {
		width = height * 0.75
	} else {
		width = model.PortraitWidth
	}
	width += insets.Left + insets.Right
	pref.Width += width + 1
	if hint.Width > 0 {
		pref.Width = hint.Width
	}
	return pref, pref, pref
}

func (p *portraitLayout) PerformLayout(target *unison.Panel) {
	r := target.ContentRect(false)
	frame := r
	insets := p.portrait.Border().Insets()
	frame.Width = xmath.Ceil(((frame.Height - (insets.Top + insets.Bottom)) * 0.75) + insets.Left + insets.Right)
	p.portrait.SetFrameRect(frame)
	r.X += frame.Width + 1
	r.Width -= frame.Width + 1
	p.rest.SetFrameRect(r)
}
