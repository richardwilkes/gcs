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
	"os"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/xfilepath"
	"github.com/richardwilkes/toolbox/v2/xio"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/imgfmt"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

const (
	minImageDockableScale = 10
	maxImageDockableScale = 1000
)

var (
	_ FileBackedDockable = &ImageDockable{}
	_ unison.TabCloser   = &ImageDockable{}
)

// ImageDockable holds the view for an image file.
type ImageDockable struct {
	unison.Panel
	path          string
	drawable      unison.Drawable
	drawablePanel *unison.Panel
	scroll        *unison.ScrollPanel
	scale         int
	dragStart     geom.Point
	dragOrigin    geom.Point
	inDrag        bool
}

// NewImageDockable creates a new unison.Dockable for image files.
func NewImageDockable(filePath string) (unison.Dockable, error) {
	var drawable unison.Drawable
	var size geom.Size
	if strings.HasSuffix(strings.ToLower(filePath), ".svg") {
		r, err := os.Open(filePath)
		if err != nil {
			return nil, errs.Wrap(err)
		}
		defer xio.CloseIgnoringErrors(r)
		var svg *unison.SVG
		if svg, err = unison.NewSVGFromReader(r, unison.SVGOptionIgnoreUnsupported(),
			unison.SVGOptionWarnParseErrors()); err != nil {
			return nil, err
		}
		drawable = &unison.DrawableSVG{
			SVG:  svg,
			Size: svg.SuggestedSize(),
		}
		size = svg.Size()
	} else {
		img, err := unison.NewImageFromFilePathOrURL(filePath, geom.NewPoint(1, 1).DivPt(unison.PrimaryDisplay().Scale))
		if err != nil {
			return nil, err
		}
		drawable = img
		size = img.Size()
	}
	d := &ImageDockable{
		path:     filePath,
		drawable: drawable,
		scale:    gurps.GlobalSettings().General.InitialImageUIScale,
	}
	d.Self = d
	d.SetLayout(&unison.FlexLayout{Columns: 1})

	d.drawablePanel = unison.NewPanel()
	d.drawablePanel.SetSizer(d.imageSizer)
	d.drawablePanel.DrawCallback = d.draw
	d.drawablePanel.MouseDownCallback = d.mouseDown
	d.drawablePanel.MouseDragCallback = d.mouseDrag
	d.drawablePanel.MouseUpCallback = d.mouseUp
	d.drawablePanel.UpdateCursorCallback = d.updateCursor
	d.drawablePanel.SetFocusable(true)

	d.scroll = unison.NewScrollPanel()
	d.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	d.scroll.SetContent(d.drawablePanel, behavior.Fill, behavior.Fill)

	typeLabel := unison.NewLabel()
	typeLabel.Font = unison.DefaultFieldTheme.Font
	typeLabel.SetTitle(imgfmt.ForPath(filePath).String())

	sizeLabel := unison.NewLabel()
	sizeLabel.Font = unison.DefaultFieldTheme.Font
	sizeLabel.SetTitle(fmt.Sprintf("%d x %d pixels", int(size.Width), int(size.Height)))

	toolbar := unison.NewPanel()
	toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{},
		geom.Insets{Bottom: 1}, false), unison.NewEmptyBorder(unison.StdInsets())))
	toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	toolbar.AddChild(NewDefaultInfoPop())
	toolbar.AddChild(
		NewScaleField(
			minImageDockableScale,
			maxImageDockableScale,
			func() int { return gurps.GlobalSettings().General.InitialImageUIScale },
			func() int { return d.scale },
			func(scale int) { d.scale = scale },
			nil,
			true,
			false,
			d.scroll,
		),
	)
	toolbar.AddChild(typeLabel)
	toolbar.AddChild(sizeLabel)
	toolbar.SetLayout(&unison.FlexLayout{
		Columns:  len(toolbar.Children()),
		HSpacing: unison.StdHSpacing,
	})

	d.AddChild(toolbar)
	d.AddChild(d.scroll)

	return d, nil
}

func (d *ImageDockable) updateCursor(_ geom.Point) *unison.Cursor {
	if d.inDrag {
		return unison.MoveCursor()
	}
	return unison.ArrowCursor()
}

func (d *ImageDockable) mouseDown(where geom.Point, _, _ int, _ unison.Modifiers) bool {
	d.dragStart = d.drawablePanel.PointToRoot(where)
	d.dragOrigin.X, d.dragOrigin.Y = d.scroll.Position()
	d.inDrag = true
	d.RequestFocus()
	d.UpdateCursorNow()
	return true
}

func (d *ImageDockable) mouseDrag(where geom.Point, _ int, _ unison.Modifiers) bool {
	pt := d.dragStart.Sub(d.drawablePanel.PointToRoot(where)).Add(d.dragOrigin)
	d.scroll.SetPosition(pt.X, pt.Y)
	return true
}

func (d *ImageDockable) mouseUp(_ geom.Point, _ int, _ unison.Modifiers) bool {
	d.inDrag = false
	d.UpdateCursorNow()
	return true
}

func (d *ImageDockable) imageSizer(_ geom.Size) (minSize, prefSize, maxSize geom.Size) {
	prefSize = d.drawable.LogicalSize()
	return geom.NewSize(50, 50), prefSize, unison.MaxSize(prefSize)
}

func (d *ImageDockable) draw(gc *unison.Canvas, dirty geom.Rect) {
	gc.DrawRect(dirty, unison.ThemeSurface.Paint(gc, dirty, paintstyle.Fill))
	d.drawable.DrawInRect(gc, geom.Rect{Size: d.drawable.LogicalSize()}, nil, nil)
}

// TitleIcon implements ux.FileBackedDockable
func (d *ImageDockable) TitleIcon(suggestedSize geom.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  gurps.FileInfoFor(d.path).SVG,
		Size: suggestedSize,
	}
}

// Title implements ux.FileBackedDockable
func (d *ImageDockable) Title() string {
	return xfilepath.BaseName(d.path)
}

// Tooltip implements ux.FileBackedDockable
func (d *ImageDockable) Tooltip() string {
	return d.path
}

// BackingFilePath implements ux.FileBackedDockable
func (d *ImageDockable) BackingFilePath() string {
	return d.path
}

// SetBackingFilePath implements ux.FileBackedDockable
func (d *ImageDockable) SetBackingFilePath(p string) {
	d.path = p
	UpdateTitleForDockable(d)
}

// Modified implements ux.FileBackedDockable
func (d *ImageDockable) Modified() bool {
	return false
}

// MayAttemptClose implements unison.TabCloser
func (d *ImageDockable) MayAttemptClose() bool {
	return true
}

// AttemptClose implements unison.TabCloser
func (d *ImageDockable) AttemptClose() bool {
	return AttemptCloseForDockable(d)
}
