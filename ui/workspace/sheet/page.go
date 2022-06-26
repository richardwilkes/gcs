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

package sheet

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/constants"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/theme"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xmath"
	"github.com/richardwilkes/unison"
)

// Page holds a logical page worth of content.
type Page struct {
	unison.Panel
	flex       *unison.FlexLayout
	entity     *gurps.Entity
	lastInsets unison.Insets
}

// NewPage creates a new page.
func NewPage(entity *gurps.Entity) *Page {
	p := &Page{
		entity: entity,
		flex: &unison.FlexLayout{
			Columns:  1,
			HSpacing: 1,
			VSpacing: 1,
		},
	}
	p.Self = p
	p.lastInsets = p.insets()
	p.SetBorder(unison.NewEmptyBorder(p.lastInsets))
	p.SetLayout(p)
	p.DrawCallback = p.drawSelf
	return p
}

// LayoutSizes implements unison.Layout
func (p *Page) LayoutSizes(_ *unison.Panel, _ unison.Size) (min, pref, max unison.Size) {
	s := gurps.SheetSettingsFor(p.entity)
	w, _ := s.Page.Orientation.Dimensions(s.Page.Size.Dimensions())
	if insets := p.insets(); insets != p.lastInsets {
		p.lastInsets = insets
		p.SetBorder(unison.NewEmptyBorder(insets))
	}
	_, size, _ := p.flex.LayoutSizes(p.AsPanel(), unison.Size{Width: w.Pixels()})
	pref.Width = w.Pixels()
	pref.Height = size.Height
	return pref, pref, pref
}

// PerformLayout implements unison.Layout
func (p *Page) PerformLayout(_ *unison.Panel) {
	p.flex.PerformLayout(p.AsPanel())
}

// ApplyPreferredSize to this panel.
func (p *Page) ApplyPreferredSize() {
	r := p.FrameRect()
	_, pref, _ := p.Sizes(unison.Size{})
	r.Size = pref
	p.SetFrameRect(r)
	p.ValidateLayout()
}

func (p *Page) insets() unison.Insets {
	sheetSettings := gurps.SheetSettingsFor(p.entity)
	insets := unison.Insets{
		Top:    sheetSettings.Page.TopMargin.Pixels(),
		Left:   sheetSettings.Page.LeftMargin.Pixels(),
		Bottom: sheetSettings.Page.BottomMargin.Pixels(),
		Right:  sheetSettings.Page.RightMargin.Pixels(),
	}
	height := theme.PageFooterSecondaryFont.LineHeight()
	insets.Bottom += xmath.Max(theme.PageFooterPrimaryFont.LineHeight(), height) + height
	return insets
}

func (p *Page) drawSelf(gc *unison.Canvas, _ unison.Rect) {
	insets := p.insets()
	_, prefSize, _ := p.LayoutSizes(nil, unison.Size{})
	r := unison.Rect{Size: prefSize}
	gc.DrawRect(r, theme.PageColor.Paint(gc, r, unison.Fill))
	r.X += insets.Left
	r.Width -= insets.Left + insets.Right
	r.Y = r.Bottom() - insets.Bottom
	r.Height = insets.Bottom
	parent := p.Parent()
	pageNumber := parent.IndexOfChild(p) + 1

	primaryDecorations := &unison.TextDecoration{
		Font:  theme.PageFooterPrimaryFont,
		Paint: theme.OnPageColor.Paint(gc, r, unison.Fill),
	}
	secondaryDecorations := &unison.TextDecoration{
		Font:  theme.PageFooterSecondaryFont,
		Paint: primaryDecorations.Paint,
	}

	var title string
	if gurps.SheetSettingsFor(p.entity).UseTitleInFooter {
		title = p.entity.Profile.Title
	} else {
		title = p.entity.Profile.Name
	}
	center := unison.NewText(title, primaryDecorations)
	left := unison.NewText(fmt.Sprintf(i18n.Text("%s is copyrighted ©%s by %s"), cmdline.AppName,
		cmdline.ResolveCopyrightYears(), cmdline.CopyrightHolder), secondaryDecorations)
	right := unison.NewText(fmt.Sprintf(i18n.Text("Modified %s"), p.entity.ModifiedOn), secondaryDecorations)
	if pageNumber&1 == 0 {
		left, right = right, left
	}
	y := r.Y + xmath.Max(xmath.Max(left.Baseline(), right.Baseline()), center.Baseline())
	left.Draw(gc, r.X, y)
	center.Draw(gc, r.X+(r.Width-center.Width())/2, y)
	right.Draw(gc, r.Right()-right.Width(), y)
	y = r.Y + xmath.Max(xmath.Max(left.Height(), right.Height()), center.Height())

	center = unison.NewText(constants.WebSiteDomain, secondaryDecorations)
	left = unison.NewText(i18n.Text("All rights reserved"), secondaryDecorations)
	right = unison.NewText(fmt.Sprintf(i18n.Text("Page %d of %d"), pageNumber, len(parent.Children())), secondaryDecorations)
	if pageNumber&1 == 0 {
		left, right = right, left
	}
	y += xmath.Max(xmath.Max(left.Baseline(), right.Baseline()), center.Baseline())
	left.Draw(gc, r.X, y)
	center.Draw(gc, r.X+(r.Width-center.Width())/2, y)
	right.Draw(gc, r.Right()-right.Width(), y)
}
