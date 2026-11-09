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
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/pathop"
	"github.com/richardwilkes/unison/enums/weight"
)

// The banner's colors are fixed rather than taken from the theme, since it is meant to look like hazard tape in both
// light and dark mode. The amber is well short of pure yellow so that the banner does not glare, and the stripes are a
// charcoal rather than black so that they do not shout against it.
var (
	hazardBannerAmber = unison.RGB(214, 168, 0)
	hazardBannerDark  = unison.RGB(40, 40, 40)
)

// hazardBanner is a hazard-striped warning shown across the top of a dockable. A list shows one while a filter is
// hiding part of it, to say why the commands that add, remove or rearrange rows are turned off, and any dockable opened
// from a file written by a newer version of GCS shows one to say that not all of it can be shown or edited. The text
// sits on a solid plate in the middle of the stripes, so that it is never read across them. It is drawn entirely by
// hand, since nothing else in the app looks like this.
type hazardBanner struct {
	unison.Panel
	text string
}

// newHazardBanner returns a banner saying text, with tooltip, if not empty, as the longer explanation behind it.
func newHazardBanner(text, tooltip string) *hazardBanner {
	b := &hazardBanner{text: text}
	b.Self = b
	if tooltip != "" {
		b.Tooltip = unison.NewTooltipWithText(tooltip)
	}
	b.SetSizer(b.sizes)
	b.DrawCallback = b.draw
	b.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	return b
}

// filterBannerText is what the banner over a filtered list says. The commands it refers to are the ones
// installStandardTableCmdHandlers and NewTableDockable turn off while the table is filtered: creating, duplicating,
// deleting, moving and dropping rows.
func filterBannerText() string {
	return i18n.Text("Adding, removing & rearranging items is disabled while filtering")
}

// newerVersionBannerText is what the banner over a file written by a newer version of GCS says (see newerFormatMarker).
// It is kept short enough to fit a narrow list; newerVersionBannerTooltip holds the explanation.
func newerVersionBannerText() string {
	return i18n.Text("Written by a newer version of GCS. Its new data is kept, but cannot be shown or edited.")
}

// newerVersionBannerTooltip explains the newer version banner.
func newerVersionBannerTooltip() string {
	return i18n.Text("This file was written by a newer version of GCS. Anything this version does not understand is kept as-is and written back out unchanged when the file is saved, but it cannot be shown or edited here. Update GCS to make use of it.")
}

// hazardBannerFont is a bold, slightly reduced form of the label font, scaled the same way as the secondary text
// elsewhere in the app so that the banner stays a notice rather than a heading. It is looked up each time rather than
// held, so that a change to the label font in the settings is picked up by the next draw.
func hazardBannerFont() unison.Font {
	desc := unison.LabelFont.Descriptor()
	desc.Weight = weight.Bold
	desc.Size *= 0.8
	return desc.Font()
}

// sizes asks for a single line of the text, its plate and a band of stripes above and below it. The banner is kept
// short, with only enough stripe showing around the plate to read as hazard tape, so that it takes as little from the
// list as it can. The minimum width is left at zero so that the banner never makes the list wider than its table and
// toolbar would be on their own; when narrower than the plate, the banner clips it evenly on both sides.
func (b *hazardBanner) sizes(_ geom.Size) (minSize, prefSize, maxSize geom.Size) {
	plate := b.plateSize()
	prefSize.Width = plate.Width + 2*unison.StdHSpacing
	// Half a spacing above and below the plate, so that the stripes are visible
	prefSize.Height = plate.Height + unison.StdVSpacing
	if border := b.Border(); border != nil {
		prefSize = prefSize.Add(border.Insets().Size())
	}
	prefSize = prefSize.Ceil()
	minSize = prefSize
	minSize.Width = 0
	return minSize, prefSize, unison.MaxSize(prefSize)
}

func (b *hazardBanner) draw(gc *unison.Canvas, _ geom.Rect) {
	rect := b.ContentRect(false)
	gc.Save()
	defer gc.Restore()
	gc.ClipRect(rect, pathop.Intersect, false)
	gc.DrawRect(rect, hazardBannerAmber.Paint(gc, rect, paintstyle.Fill))
	gc.DrawPath(hazardBannerStripes(rect), hazardBannerDark.Paint(gc, rect, paintstyle.Fill))
	plate := b.plateRect(rect)
	gc.DrawRoundedRect(plate, geom.NewSize(unison.StdVSpacing, unison.StdVSpacing),
		unison.Black.Paint(gc, plate, paintstyle.Fill))
	f := hazardBannerFont()
	pt := geom.NewPoint(plate.CenterX()-f.SimpleWidth(b.text)/2, plate.CenterY()-f.LineHeight()/2+f.Baseline())
	gc.DrawSimpleString(b.text, pt, f, unison.White.Paint(gc, plate, paintstyle.Fill))
}

// plateSize returns the size of the plate the text sits on: just large enough to hold a line of the text with a full
// spacing to either side of it and half of one above and below.
func (b *hazardBanner) plateSize() geom.Size {
	f := hazardBannerFont()
	return geom.NewSize(f.SimpleWidth(b.text)+2*unison.StdHSpacing, f.LineHeight()+unison.StdVSpacing)
}

// plateRect returns the rect of the plate, centered in the banner's rect.
func (b *hazardBanner) plateRect(rect geom.Rect) geom.Rect {
	size := b.plateSize()
	return geom.NewRect(rect.CenterX()-size.Width/2, rect.CenterY()-size.Height/2, size.Width, size.Height)
}

// hazardBannerStripes returns the dark stripes as one path, each a parallelogram leaning to the right at 45° that runs
// from the bottom of the rect to its top. The first one starts a rect's height to the left of the rect, since that is
// how far its top edge is shifted from its bottom edge, so that the stripes are already in full swing at the left edge.
// The path spills past the right edge and is expected to be clipped by the caller.
func hazardBannerStripes(rect geom.Rect) *unison.Path {
	const stripeWidth = 24
	path := unison.NewPath()
	for x := rect.X - rect.Height; x < rect.Right(); x += 2 * stripeWidth {
		path.MoveTo(geom.NewPoint(x, rect.Bottom()))
		path.LineTo(geom.NewPoint(x+stripeWidth, rect.Bottom()))
		path.LineTo(geom.NewPoint(x+stripeWidth+rect.Height, rect.Y))
		path.LineTo(geom.NewPoint(x+rect.Height, rect.Y))
		path.Close()
	}
	return path
}
