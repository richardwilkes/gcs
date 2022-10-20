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

package external

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/constants"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/gcs/v5/ui/workspace"
	"github.com/richardwilkes/toolbox/i18n"
	xfs "github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
)

const (
	minImageDockableScale = 10
	maxImageDockableScale = 1000
)

var (
	_ workspace.FileBackedDockable = &ImageDockable{}
	_ unison.TabCloser             = &ImageDockable{}
)

// ImageDockable holds the view for an image file.
type ImageDockable struct {
	unison.Panel
	path       string
	img        *unison.Image
	imgPanel   *unison.Panel
	scroll     *unison.ScrollPanel
	scaleField *widget.PercentageField
	dragStart  unison.Point
	dragOrigin unison.Point
	scale      int
	inDrag     bool
}

// NewImageDockable creates a new unison.Dockable for image files.
func NewImageDockable(filePath string) (unison.Dockable, error) {
	img, err := unison.NewImageFromFilePathOrURL(filePath, 1)
	if err != nil {
		return nil, err
	}
	d := &ImageDockable{
		path:  filePath,
		img:   img,
		scale: 100,
	}
	d.Self = d
	d.SetLayout(&unison.FlexLayout{Columns: 1})

	d.imgPanel = unison.NewPanel()
	d.imgPanel.SetSizer(d.imageSizer)
	d.imgPanel.DrawCallback = d.draw
	d.imgPanel.MouseDownCallback = d.mouseDown
	d.imgPanel.MouseDragCallback = d.mouseDrag
	d.imgPanel.MouseUpCallback = d.mouseUp
	d.imgPanel.UpdateCursorCallback = d.updateCursor
	d.imgPanel.MouseWheelCallback = d.mouseWheel
	d.imgPanel.SetFocusable(true)

	d.scroll = unison.NewScrollPanel()
	d.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
		VGrab:  true,
	})
	d.scroll.SetContent(d.imgPanel, unison.FillBehavior, unison.FillBehavior)

	info := widget.NewInfoPop()
	info.Target = d.scroll
	info.AddHelpInfo(fmt.Sprintf(i18n.Text(`Holding down the %s key while using
the mouse wheel will change the scale.`), unison.OptionModifier.String()))

	scaleTitle := i18n.Text("Scale")
	d.scaleField = widget.NewPercentageField(nil, "", scaleTitle,
		func() int { return d.scale },
		func(v int) {
			viewRect := d.scroll.ContentView().ContentRect(false)
			center := d.imgPanel.PointFromRoot(d.scroll.ContentView().PointToRoot(viewRect.Center()))
			center.X /= float32(d.scale) / 100
			center.X *= float32(v) / 100
			center.Y /= float32(d.scale) / 100
			center.Y *= float32(v) / 100
			d.scale = v
			d.scroll.MarkForLayoutAndRedraw()
			d.scroll.ValidateLayout()
			viewRect.X = center.X - viewRect.Width/2
			viewRect.Y = center.Y - viewRect.Height/2
			d.imgPanel.ScrollRectIntoView(viewRect)
		}, minImageDockableScale, maxImageDockableScale, false, false)
	d.scaleField.Tooltip = unison.NewTooltipWithText(scaleTitle)

	typeLabel := unison.NewLabel()
	typeLabel.Text = unison.EncodedImageFormatForPath(filePath).String()
	typeLabel.Font = unison.DefaultFieldTheme.Font

	sizeLabel := unison.NewLabel()
	size := img.Size()
	sizeLabel.Text = fmt.Sprintf("%d x %d pixels", int(size.Width), int(size.Height))
	sizeLabel.Font = unison.DefaultFieldTheme.Font

	toolbar := unison.NewPanel()
	toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.Insets{Bottom: 1},
		false), unison.NewEmptyBorder(unison.StdInsets())))
	toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	toolbar.AddChild(info)
	toolbar.AddChild(d.scaleField)
	toolbar.AddChild(typeLabel)
	toolbar.AddChild(sizeLabel)
	toolbar.SetLayout(&unison.FlexLayout{
		Columns:  len(toolbar.Children()),
		HSpacing: unison.StdHSpacing,
	})

	d.AddChild(toolbar)
	d.AddChild(d.scroll)

	widget.InstallViewScaleHandlers(d, func() int { return 100 }, minImageDockableScale, maxImageDockableScale,
		func() int { return d.scale }, d.adjustScale)

	return d, nil
}

func (d *ImageDockable) updateCursor(_ unison.Point) *unison.Cursor {
	if d.inDrag {
		return unison.MoveCursor()
	}
	return unison.ArrowCursor()
}

func (d *ImageDockable) mouseDown(where unison.Point, _, _ int, _ unison.Modifiers) bool {
	d.dragStart = d.imgPanel.PointToRoot(where)
	d.dragOrigin.X, d.dragOrigin.Y = d.scroll.Position()
	d.inDrag = true
	d.RequestFocus()
	d.UpdateCursorNow()
	return true
}

func (d *ImageDockable) mouseDrag(where unison.Point, _ int, _ unison.Modifiers) bool {
	pt := d.dragStart
	pt.Subtract(d.imgPanel.PointToRoot(where))
	d.scroll.SetPosition(d.dragOrigin.X+pt.X, d.dragOrigin.Y+pt.Y)
	return true
}

func (d *ImageDockable) mouseUp(_ unison.Point, _ int, _ unison.Modifiers) bool {
	d.inDrag = false
	d.UpdateCursorNow()
	return true
}

func (d *ImageDockable) mouseWheel(_, delta unison.Point, mod unison.Modifiers) bool {
	if !mod.OptionDown() {
		return false
	}
	scale := d.scale + int(delta.Y*constants.ScaleDelta)
	if scale < minImageDockableScale {
		scale = minImageDockableScale
	} else if scale > maxImageDockableScale {
		scale = maxImageDockableScale
	}
	widget.SetFieldValue(d.scaleField.Field, d.scaleField.Format(scale))
	return true
}

func (d *ImageDockable) adjustScale(scale int) {
	if d.scale != scale {
		widget.SetFieldValue(d.scaleField.Field, d.scaleField.Format(scale))
	}
}

func (d *ImageDockable) imageSizer(_ unison.Size) (min, pref, max unison.Size) {
	pref = d.img.Size()
	pref.Width *= float32(d.scale) / 100
	pref.Height *= float32(d.scale) / 100
	return unison.NewSize(50, 50), pref, unison.MaxSize(pref)
}

func (d *ImageDockable) draw(gc *unison.Canvas, dirty unison.Rect) {
	gc.DrawRect(dirty, unison.ContentColor.Paint(gc, dirty, unison.Fill))
	size := d.img.Size()
	gc.DrawImageInRect(d.img, unison.NewRect(0, 0, size.Width*float32(d.scale)/100, size.Height*float32(d.scale)/100), nil, nil)
}

// TitleIcon implements workspace.FileBackedDockable
func (d *ImageDockable) TitleIcon(suggestedSize unison.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  library.FileInfoFor(d.path).SVG,
		Size: suggestedSize,
	}
}

// Title implements workspace.FileBackedDockable
func (d *ImageDockable) Title() string {
	return xfs.BaseName(d.path)
}

// Tooltip implements workspace.FileBackedDockable
func (d *ImageDockable) Tooltip() string {
	return d.path
}

// BackingFilePath implements workspace.FileBackedDockable
func (d *ImageDockable) BackingFilePath() string {
	return d.path
}

// SetBackingFilePath implements workspace.FileBackedDockable
func (d *ImageDockable) SetBackingFilePath(p string) {
	d.path = p
	if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
		dc.UpdateTitle(d)
	}
}

// Modified implements workspace.FileBackedDockable
func (d *ImageDockable) Modified() bool {
	return false
}

// MayAttemptClose implements unison.TabCloser
func (d *ImageDockable) MayAttemptClose() bool {
	return true
}

// AttemptClose implements unison.TabCloser
func (d *ImageDockable) AttemptClose() bool {
	if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
		dc.Close(d)
	}
	return true
}
