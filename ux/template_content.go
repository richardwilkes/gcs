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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
)

type templateContent struct {
	unison.Panel
	flex *unison.FlexLayout
}

// newTemplateContent creates a new page.
func newTemplateContent() *templateContent {
	p := &templateContent{
		flex: &unison.FlexLayout{
			Columns:  1,
			VSpacing: 1,
		},
	}
	p.Self = p
	p.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(4)))
	p.SetLayout(p)
	return p
}

func (p *templateContent) LayoutSizes(_ *unison.Panel, _ geom.Size) (minSize, prefSize, maxSize geom.Size) {
	s := gurps.GlobalSettings().Sheet
	w, _ := s.Page.Orientation.Dimensions(gurps.MustParsePageSize(s.Page.Size))
	_, size, _ := p.flex.LayoutSizes(p.AsPanel(), geom.Size{Width: w.Pixels()})
	prefSize.Width = w.Pixels()
	prefSize.Height = size.Height
	return prefSize, prefSize, prefSize
}

func (p *templateContent) PerformLayout(_ *unison.Panel) {
	p.flex.PerformLayout(p.AsPanel())
}

// ApplyPreferredSize to this panel.
func (p *templateContent) ApplyPreferredSize() {
	r := p.FrameRect()
	_, pref, _ := p.Sizes(geom.Size{})
	r.Size = pref
	p.SetFrameRect(r)
	p.ValidateLayout()
}
