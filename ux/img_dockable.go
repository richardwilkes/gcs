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
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/uti"
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
	_ KeyedDockable      = &ImageDockable{}
)

// ImageDockable holds the view for an image file.
type ImageDockable struct {
	fileBackedPanel
	drawable      unison.Drawable
	drawablePanel *unison.Panel
	scroll        *unison.ScrollPanel
	pan           scrollPanDrag
	scale         int
}

// NewImageDockable creates a new unison.Dockable for image files.
func NewImageDockable(filePath string) (unison.Dockable, error) {
	var drawable unison.Drawable
	var size geom.Size
	var kind string
	if isSVGPath(filePath) {
		vector, err := loadSVGFromFile(filePath)
		if err != nil {
			return nil, err
		}
		drawable = &unison.DrawableSVG{
			SVG:  vector,
			Size: vector.SuggestedSize(),
		}
		size = vector.Size()
		kind = "SVG"
	} else {
		img, err := unison.NewImageFromFilePathOrURL(context.Background(), nil, filePath,
			geom.NewPoint(1, 1).DivPt(primaryDisplayScale()), 0)
		if err != nil {
			return nil, err
		}
		drawable = img
		size = img.Size()
		kind = imgfmt.ForPath(filePath).String()
	}
	d := &ImageDockable{
		drawable: drawable,
		scale:    gurps.GlobalSettings().General.InitialImageUIScale,
	}
	d.Self = d
	d.initFileViewer(d, filePath)
	d.SetLayout(&unison.FlexLayout{Columns: 1})

	d.drawablePanel = unison.NewPanel()
	d.drawablePanel.SetSizer(d.imageSizer)
	d.drawablePanel.DrawCallback = d.draw
	d.drawablePanel.SetFocusable(true)

	d.scroll = unison.NewScrollPanel()
	d.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	d.scroll.SetContent(d.drawablePanel, behavior.Fill, behavior.Fill)
	d.pan.install(d.scroll, d.drawablePanel, d)

	typeLabel := unison.NewLabel()
	typeLabel.Font = unison.DefaultFieldTheme.Font
	typeLabel.SetTitle(kind)

	sizeLabel := unison.NewLabel()
	sizeLabel.Font = unison.DefaultFieldTheme.Font
	sizeLabel.SetTitle(fmt.Sprintf("%d x %d pixels", int(size.Width), int(size.Height)))

	toolbar := newToolbar()
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
	finishToolbarLayout(toolbar)

	d.AddChild(toolbar)
	d.AddChild(d.scroll)

	return d, nil
}

// isSVGPath returns true if the path uses one of the extensions registered for SVG content. Note that this covers both
// ".svg" and ".svgz", since both are registered by uti.SVG and therefore advertised as openable.
func isSVGPath(filePath string) bool {
	return slices.Contains(uti.SVG.Extensions, strings.ToLower(filepath.Ext(filePath)))
}

// loadSVGFromFile loads the SVG found at the given path. ".svgz" files hold gzip-compressed SVG data, which the SVG
// parser can't consume directly, so the content is decompressed first. The gzip header is checked rather than the
// extension, since compressed content is sometimes stored with the plain ".svg" extension as well.
func loadSVGFromFile(filePath string) (*unison.SVG, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, errs.Wrap(err)
	}
	defer xio.CloseIgnoringErrors(f)
	buffered := bufio.NewReader(f)
	var r io.Reader = buffered
	if header, peekErr := buffered.Peek(2); peekErr == nil && header[0] == 0x1f && header[1] == 0x8b {
		var gz *gzip.Reader
		if gz, err = gzip.NewReader(buffered); err != nil {
			return nil, errs.Wrap(err)
		}
		defer xio.CloseIgnoringErrors(gz)
		r = gz
	}
	return unison.NewSVGFromReader(r)
}

func (d *ImageDockable) imageSizer(_ geom.Size) (minSize, prefSize, maxSize geom.Size) {
	prefSize = d.drawable.LogicalSize()
	return geom.NewSize(50, 50), prefSize, unison.MaxSize(prefSize)
}

func (d *ImageDockable) draw(gc *unison.Canvas, dirty geom.Rect) {
	gc.DrawRect(dirty, unison.ThemeSurface.Paint(gc, dirty, paintstyle.Fill))
	d.drawable.DrawInRect(gc, geom.Rect{Size: d.drawable.LogicalSize()}, nil, nil)
}
