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
	"strconv"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/autoscale"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/desktop"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/toolbox/xmath"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/blendmode"
	"github.com/richardwilkes/unison/enums/filtermode"
	"github.com/richardwilkes/unison/enums/mipmapmode"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

const (
	minPDFDockableScale                = 50
	maxPDFDockableScale                = 200
	scaleCompensation                  = float32(100) / maxPDFDockableScale
	maxElapsedRenderTimeWithoutOverlay = time.Second / 2
	renderTimeSlop                     = time.Millisecond * 10
)

var (
	_ FileBackedDockable = &PDFDockable{}
	_ unison.TabCloser   = &PDFDockable{}
)

// PDFDockable holds the view for a PDFRenderer file.
type PDFDockable struct {
	unison.Panel
	path                   string
	pdf                    *PDFRenderer
	content                *unison.Panel
	docScroll              *unison.ScrollPanel
	docPanel               *unison.Panel
	tocScroll              *unison.ScrollPanel
	tocPanel               *unison.Table[*tocNode]
	divider                *unison.Panel
	tocScrollLayoutData    *unison.FlexLayoutData
	pageNumberField        *unison.Field
	scaleField             *PercentageField
	autoScalingPopup       *unison.PopupMenu[autoscale.Option]
	searchField            *unison.Field
	matchesLabel           *unison.Label
	sideBarButton          *unison.Button
	backButton             *unison.Button
	forwardButton          *unison.Button
	firstPageButton        *unison.Button
	previousPageButton     *unison.Button
	nextPageButton         *unison.Button
	lastPageButton         *unison.Button
	page                   *PDFPage
	link                   *PDFLink
	rolloverRect           unison.Rect
	scale                  int
	historyPos             int
	history                []int
	dragStart              unison.Point
	dragOrigin             unison.Point
	autoScaling            autoscale.Option
	inDrag                 bool
	noUpdate               bool
	adjustTableSizePending bool
	needDockableResize     bool
}

// NewPDFDockable creates a new unison.Dockable for PDFRenderer files.
func NewPDFDockable(filePath string, initialPage int) (unison.Dockable, error) {
	generalSettings := gurps.GlobalSettings().General
	d := &PDFDockable{
		path:               filePath,
		scale:              generalSettings.InitialPDFUIScale,
		autoScaling:        generalSettings.PDFAutoScaling,
		noUpdate:           true,
		needDockableResize: true,
	}
	d.Self = d
	var err error
	if d.pdf, err = NewPDFRenderer(filePath, func() {
		unison.InvokeTask(d.pageLoaded)
	}); err != nil {
		return nil, err
	}
	d.KeyDownCallback = d.keyDown
	d.FocusChangeInHierarchyCallback = d.focusChangeInHierarchy
	d.GainedFocusCallback = d.pdf.RequestRenderPriority
	d.SetLayout(&unison.FlexLayout{Columns: 1})

	d.createTOC()
	d.createContent()

	d.AddChild(d.createToolbar()) // Creation of the toolbar has to be after content creation
	d.AddChild(d.content)
	d.content.AddChild(d.docScroll)

	d.noUpdate = false
	d.scaleField.SetEnabled(d.autoScaling != autoscale.No)
	d.LoadPage(initialPage)

	return d, nil
}

// DockKey implements KeyedDockable.
func (d *PDFDockable) DockKey() string {
	return filePrefix + d.path
}

func (d *PDFDockable) createToolbar() *unison.Panel {
	outer := unison.NewPanel()
	outer.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, 0, unison.Insets{Bottom: 1},
		false), unison.NewEmptyBorder(unison.StdInsets())))
	outer.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})

	first := unison.NewPanel()
	first.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})

	info := NewInfoPop()
	AddHelpToInfoPop(info, i18n.Text("Within this view, these keys have the following effects:\n"))
	AddKeyBindingInfoToInfoPop(info, unison.KeyBinding{KeyCode: unison.KeyHome}, i18n.Text("Go to first page"))
	AddKeyBindingInfoToInfoPop(info, unison.KeyBinding{KeyCode: unison.KeyEnd}, i18n.Text("Go to last page"))
	AddKeyBindingInfoToInfoPop(info, unison.KeyBinding{KeyCode: unison.KeyLeft}, i18n.Text("Go to previous page"))
	AddKeyBindingInfoToInfoPop(info, unison.KeyBinding{KeyCode: unison.KeyUp}, i18n.Text("Go to previous page"))
	AddKeyBindingInfoToInfoPop(info, unison.KeyBinding{KeyCode: unison.KeyRight}, i18n.Text("Go to next page"))
	AddKeyBindingInfoToInfoPop(info, unison.KeyBinding{KeyCode: unison.KeyDown}, i18n.Text("Go to next page"))
	AddScalingHelpToInfoPop(info)
	first.AddChild(info)

	d.scaleField = NewScaleField(
		minPDFDockableScale,
		maxPDFDockableScale,
		func() int { return gurps.GlobalSettings().General.InitialPDFUIScale },
		func() int { return d.scale },
		func(scale int) { d.scale = scale },
		d.MarkForRedraw,
		false,
		d.docScroll,
	)
	d.scaleField.SetEnabled(false)
	first.AddChild(d.scaleField)

	d.autoScalingPopup = unison.NewPopupMenu[autoscale.Option]()
	for _, mode := range autoscale.Options {
		d.autoScalingPopup.AddItem(mode)
	}
	d.autoScalingPopup.Select(d.autoScaling)
	d.autoScalingPopup.SelectionChangedCallback = func(popup *unison.PopupMenu[autoscale.Option]) {
		if mode, ok := popup.Selected(); ok {
			d.autoScaling = mode
			d.scaleField.SetEnabled(d.autoScaling == autoscale.No)
			d.docScroll.MarkForRedraw()
		}
	}
	first.AddChild(d.autoScalingPopup)

	pageLabel := unison.NewLabel()
	pageLabel.Font = unison.DefaultFieldTheme.Font
	pageLabel.SetTitle(i18n.Text("Page"))
	first.AddChild(pageLabel)

	d.pageNumberField = unison.NewField()
	d.pageNumberField.SetMinimumTextWidthUsing(strconv.Itoa(d.pdf.PageCount() * 10))
	d.pageNumberField.ModifiedCallback = func(_, after *unison.FieldState) {
		if d.noUpdate {
			return
		}
		if pageNum, e := strconv.Atoi(after.Text); e == nil && pageNum > 0 && pageNum <= d.pdf.PageCount() {
			d.LoadPage(pageNum - 1)
		}
	}
	d.pageNumberField.ValidateCallback = func() bool {
		pageNum, e := strconv.Atoi(d.pageNumberField.Text())
		if e != nil || pageNum < 1 || pageNum > d.pdf.PageCount() {
			return false
		}
		return true
	}
	first.AddChild(d.pageNumberField)

	ofLabel := unison.NewLabel()
	ofLabel.Font = unison.DefaultFieldTheme.Font
	ofLabel.SetTitle(fmt.Sprintf(i18n.Text("of %d"), d.pdf.PageCount()))
	first.AddChild(ofLabel)

	first.AddChild(NewToolbarSeparator())

	d.backButton = unison.NewSVGButton(svg.Back)
	d.backButton.Tooltip = newWrappedTooltip(i18n.Text("Back"))
	d.backButton.ClickCallback = d.Back
	first.AddChild(d.backButton)

	d.forwardButton = unison.NewSVGButton(svg.Forward)
	d.forwardButton.Tooltip = newWrappedTooltip(i18n.Text("Forward"))
	d.forwardButton.ClickCallback = d.Forward
	first.AddChild(d.forwardButton)

	first.AddChild(NewToolbarSeparator())

	d.firstPageButton = unison.NewSVGButton(svg.First)
	d.firstPageButton.Tooltip = newWrappedTooltip(i18n.Text("First Page"))
	d.firstPageButton.ClickCallback = func() { d.LoadPage(0) }
	first.AddChild(d.firstPageButton)

	d.previousPageButton = unison.NewSVGButton(svg.Previous)
	d.previousPageButton.Tooltip = newWrappedTooltip(i18n.Text("Previous Page"))
	d.previousPageButton.ClickCallback = func() { d.LoadPage(d.pdf.MostRecentPageNumber() - 1) }
	first.AddChild(d.previousPageButton)

	d.nextPageButton = unison.NewSVGButton(svg.Next)
	d.nextPageButton.Tooltip = newWrappedTooltip(i18n.Text("Next Page"))
	d.nextPageButton.ClickCallback = func() { d.LoadPage(d.pdf.MostRecentPageNumber() + 1) }
	first.AddChild(d.nextPageButton)

	d.lastPageButton = unison.NewSVGButton(svg.Last)
	d.lastPageButton.Tooltip = newWrappedTooltip(i18n.Text("Last Page"))
	d.lastPageButton.ClickCallback = func() { d.LoadPage(d.pdf.PageCount() - 1) }
	first.AddChild(d.lastPageButton)

	first.SetLayout(&unison.FlexLayout{
		Columns:  len(first.Children()),
		HSpacing: unison.StdHSpacing,
	})
	outer.AddChild(first)

	second := unison.NewPanel()
	second.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})

	d.sideBarButton = unison.NewSVGButton(svg.SideBar)
	d.sideBarButton.Tooltip = newWrappedTooltip(i18n.Text("Toggle the Sidebar"))
	d.sideBarButton.ClickCallback = d.toggleSideBar
	d.sideBarButton.SetEnabled(false)
	second.AddChild(d.sideBarButton)

	d.searchField = NewSearchField(i18n.Text("Page Search"), func(_, _ *unison.FieldState) {
		if d.noUpdate {
			return
		}
		d.LoadPage(d.pdf.MostRecentPageNumber())
	})
	second.AddChild(d.searchField)

	d.matchesLabel = unison.NewLabel()
	d.matchesLabel.SetTitle("-")
	d.matchesLabel.Tooltip = newWrappedTooltip(i18n.Text("Number of matches found"))
	second.AddChild(d.matchesLabel)

	second.SetLayout(&unison.FlexLayout{
		Columns:  len(second.Children()),
		HSpacing: unison.StdHSpacing,
	})
	outer.AddChild(second)

	outer.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	return outer
}

func (d *PDFDockable) createTOC() {
	d.tocPanel = unison.NewTable(&unison.SimpleTableModel[*tocNode]{})
	d.tocPanel.Columns = make([]unison.ColumnInfo, 1)
	d.tocPanel.DoubleClickCallback = d.tocDoubleClick
	d.tocPanel.SelectionChangedCallback = d.tocSelectionChanged

	d.tocScroll = unison.NewScrollPanel()
	d.tocScrollLayoutData = &unison.FlexLayoutData{
		SizeHint: unison.Size{Width: 200},
		HAlign:   align.Fill,
		VAlign:   align.Fill,
		VGrab:    true,
	}
	d.tocScroll.SetLayoutData(d.tocScrollLayoutData)
	d.tocScroll.SetContent(d.tocPanel, behavior.Fill, behavior.Fill)

	d.divider = unison.NewPanel()
	d.divider.SetLayoutData(&unison.FlexLayoutData{
		SizeHint: unison.Size{Width: unison.DefaultDockTheme.DockDividerSize()},
		HAlign:   align.Fill,
		VAlign:   align.Fill,
		VGrab:    true,
	})
	d.divider.UpdateCursorCallback = func(_ unison.Point) *unison.Cursor {
		return unison.ResizeHorizontalCursor()
	}
	d.divider.DrawCallback = func(gc *unison.Canvas, _ unison.Rect) {
		unison.DefaultDockTheme.DrawHorizontalGripper(gc, d.divider.ContentRect(true))
	}
	var initialPosition float32
	var eventPosition float32
	d.divider.MouseDownCallback = func(where unison.Point, _, _ int, _ unison.Modifiers) bool {
		initialPosition = d.tocScrollLayoutData.SizeHint.Width
		eventPosition = d.divider.Parent().PointFromRoot(d.divider.PointToRoot(where)).X
		return true
	}
	d.divider.MouseDragCallback = func(where unison.Point, _ int, _ unison.Modifiers) bool {
		pos := eventPosition - d.divider.Parent().PointFromRoot(d.divider.PointToRoot(where)).X
		old := d.tocScrollLayoutData.SizeHint.Width
		d.tocScrollLayoutData.SizeHint.Width = max(initialPosition-pos, 1)
		if old != d.tocScrollLayoutData.SizeHint.Width {
			d.divider.Parent().MarkForLayoutAndRedraw()
		}
		return true
	}
}

func (d *PDFDockable) createContent() {
	d.docPanel = unison.NewPanel()
	d.docPanel.SetSizer(d.docSizer)
	d.docPanel.DrawCallback = d.draw
	d.docPanel.MouseDownCallback = d.mouseDown
	d.docPanel.MouseMoveCallback = d.mouseMove
	d.docPanel.MouseDragCallback = d.mouseDrag
	d.docPanel.MouseUpCallback = d.mouseUp
	d.docPanel.UpdateCursorCallback = d.updateCursor
	d.docPanel.SetFocusable(true)

	d.docScroll = unison.NewScrollPanel()
	d.docScroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	d.docScroll.SetContent(d.docPanel, behavior.Fill, behavior.Fill)
	d.docScroll.ContentView().DrawOverCallback = d.drawOverlay

	d.content = unison.NewPanel()
	d.content.SetLayout(&unison.FlexLayout{
		Columns: 1,
		HAlign:  align.Fill,
		VAlign:  align.Fill,
	})
	d.content.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
}

// ClearHistory clears the existing history.
func (d *PDFDockable) ClearHistory() {
	d.history = nil
	d.historyPos = 0
	d.backButton.SetEnabled(false)
	d.forwardButton.SetEnabled(false)
}

// SetSearchText sets the search text and updates the display.
func (d *PDFDockable) SetSearchText(text string) {
	d.searchField.SetText(text)
}

func (d *PDFDockable) toggleSideBar() {
	if layout, ok := d.content.Layout().(*unison.FlexLayout); ok {
		if layout.Columns == 1 {
			layout.Columns = 3
			d.content.AddChildAtIndex(d.tocScroll, 0)
			d.content.AddChildAtIndex(d.divider, 1)
		} else {
			layout.Columns = 1
			d.divider.RemoveFromParent()
			d.tocScroll.RemoveFromParent()
		}
		d.content.MarkForLayoutAndRedraw()
	}
}

// Back moves back in history one step.
func (d *PDFDockable) Back() {
	if d.historyPos > 0 {
		d.historyPos--
		d.LoadPage(d.history[d.historyPos])
	}
}

// Forward moves forward in history one step.
func (d *PDFDockable) Forward() {
	if d.historyPos < len(d.history)-1 {
		d.historyPos++
		d.LoadPage(d.history[d.historyPos])
	}
}

// LoadPage loads the specified page.
func (d *PDFDockable) LoadPage(pageNumber int) {
	d.pdf.LoadPage(pageNumber, d.searchField.Text())
	d.MarkForRedraw()
}

func (d *PDFDockable) pageLoaded() {
	d.noUpdate = true
	d.scaleField.SetEnabled(false)
	defer func() {
		d.noUpdate = false
		d.scaleField.SetEnabled(d.autoScaling == autoscale.No)
	}()

	d.page = d.pdf.CurrentPage()

	if d.tocPanel.RootRowCount() == 0 {
		if toc := d.pdf.CurrentPage().TOC; len(toc) != 0 {
			d.tocPanel.SetRootRows(newTOC(d, nil, toc))
			d.tocPanel.SizeColumnsToFit(true)
			d.tocScrollLayoutData.SizeHint.Width = max(min(d.tocPanel.Columns[0].Current, 300), d.tocScrollLayoutData.SizeHint.Width)
			d.sideBarButton.SetEnabled(true)
		}
	}

	pageText := ""
	if d.page.PageNumber >= 0 {
		pageText = strconv.Itoa(d.page.PageNumber + 1)
	}
	if pageText != d.pageNumberField.Text() {
		d.pageNumberField.SetText(pageText)
		d.pageNumberField.Parent().MarkForLayoutAndRedraw()
	}

	matchText := "-"
	if d.searchField.Text() != "" {
		matchText = strconv.Itoa(len(d.page.Matches))
	}
	if matchText != d.matchesLabel.Text.String() {
		d.matchesLabel.SetTitle(matchText)
		d.matchesLabel.Parent().MarkForLayoutAndRedraw()
	}

	pageNumber := d.page.PageNumber
	if d.history == nil {
		d.history = append(d.history, pageNumber)
		d.historyPos = 0
	} else if d.history[d.historyPos] != pageNumber {
		d.historyPos++
		if d.historyPos < len(d.history) {
			if d.history[d.historyPos] != pageNumber {
				d.history[d.historyPos] = pageNumber
				d.history = d.history[:d.historyPos+1]
			}
		} else {
			d.history = append(d.history, pageNumber)
		}
	}
	lastPageNumber := d.pdf.PageCount() - 1
	d.backButton.SetEnabled(d.historyPos > 0)
	d.forwardButton.SetEnabled(d.historyPos < len(d.history)-1)
	d.firstPageButton.SetEnabled(pageNumber != 0)
	d.previousPageButton.SetEnabled(pageNumber > 0)
	d.nextPageButton.SetEnabled(pageNumber < lastPageNumber)
	d.lastPageButton.SetEnabled(pageNumber != lastPageNumber)

	d.docPanel.MarkForLayoutAndRedraw()
	d.docScroll.MarkForLayoutAndRedraw()
	d.link = nil

	if d.needDockableResize {
		d.needDockableResize = false
		if !IsDockableInWorkspace(d) {
			wnd := d.Window()
			wnd.Pack()
			frame := wnd.FrameRect()
			wnd.SetFrameRect(unison.BestDisplayForRect(frame).FitRectOnto(frame))
			d.searchField.SetScrollOffset(unison.Point{})
		}
	}
}

func (d *PDFDockable) overLink(where unison.Point) (rect unison.Rect, link *PDFLink) {
	if d.page != nil && d.page.Links != nil {
		where.X /= scaleCompensation
		where.Y /= scaleCompensation
		for _, link = range d.page.Links {
			if where.In(link.Bounds) {
				return link.Bounds, link
			}
		}
	}
	return rect, nil
}

func (d *PDFDockable) checkForLinkAt(where unison.Point) bool {
	r, link := d.overLink(where)
	if r != d.rolloverRect || link != d.link {
		d.rolloverRect = r
		d.link = link
		d.MarkForRedraw()
	}
	return link != nil
}

func (d *PDFDockable) updateCursor(pt unison.Point) *unison.Cursor {
	if d.inDrag {
		return unison.MoveCursor()
	}
	if _, link := d.overLink(pt); link != nil {
		return unison.PointingCursor()
	}
	return unison.ArrowCursor()
}

func (d *PDFDockable) mouseDown(where unison.Point, _, _ int, _ unison.Modifiers) bool {
	d.dragStart = d.docPanel.PointToRoot(where)
	d.dragOrigin.X, d.dragOrigin.Y = d.docScroll.Position()
	d.inDrag = !d.checkForLinkAt(where)
	d.docPanel.RequestFocus()
	d.UpdateCursorNow()
	return true
}

func (d *PDFDockable) mouseDrag(where unison.Point, _ int, _ unison.Modifiers) bool {
	if d.inDrag {
		pt := d.dragStart.Sub(d.docPanel.PointToRoot(where)).Add(d.dragOrigin)
		d.docScroll.SetPosition(pt.X, pt.Y)
	} else {
		d.checkForLinkAt(where)
	}
	return true
}

func (d *PDFDockable) mouseMove(where unison.Point, _ unison.Modifiers) bool {
	d.checkForLinkAt(where)
	return true
}

func (d *PDFDockable) mouseUp(where unison.Point, button int, _ unison.Modifiers) bool {
	if d.inDrag {
		d.inDrag = false
		d.UpdateCursorNow()
	} else {
		d.checkForLinkAt(where)
		if button == unison.ButtonLeft && d.link != nil {
			if d.link.PageNumber >= 0 {
				d.LoadPage(d.link.PageNumber)
			} else if err := desktop.Open(d.link.URI); err != nil {
				Workspace.ErrorHandler(i18n.Text("Unable to open link"), err)
			}
		}
	}
	return true
}

func (d *PDFDockable) focusChangeInHierarchy(_, _ *unison.Panel) {
	d.pdf.RequestRenderPriority()
}

func (d *PDFDockable) keyDown(keyCode unison.KeyCode, _ unison.Modifiers, _ bool) bool {
	switch keyCode {
	case unison.KeyHome:
		d.LoadPage(0)
	case unison.KeyEnd:
		d.LoadPage(d.pdf.PageCount() - 1)
	case unison.KeyLeft, unison.KeyUp:
		d.LoadPage(d.pdf.MostRecentPageNumber() - 1)
	case unison.KeyRight, unison.KeyDown:
		d.LoadPage(d.pdf.MostRecentPageNumber() + 1)
	default:
		return false
	}
	return true
}

func (d *PDFDockable) docSizer(_ unison.Size) (minSize, prefSize, maxSize unison.Size) {
	if d.page == nil || d.page.Error != nil {
		prefSize.Width = 400
		prefSize.Height = 300
	} else {
		prefSize = d.page.Image.LogicalSize()
		prefSize.Width *= scaleCompensation
		prefSize.Height *= scaleCompensation
	}
	return unison.NewSize(50, 50), prefSize, unison.MaxSize(prefSize)
}

func (d *PDFDockable) draw(gc *unison.Canvas, dirty unison.Rect) {
	gc.DrawRect(dirty, unison.ThemeSurface.Paint(gc, dirty, paintstyle.Fill))
	if d.page != nil && d.page.Image != nil {
		switch d.autoScaling {
		case autoscale.FitWidth:
			size := d.page.Image.LogicalSize()
			cSize := d.docScroll.ContentView().ContentRect(false).Size
			desiredScale := xmath.Floor((cSize.Width / size.Width / scaleCompensation) * 100)
			desiredScaleInt := int(min(max(desiredScale, minPDFDockableScale), maxPDFDockableScale))
			if d.scaleField.CurrentValue() != desiredScaleInt {
				d.scaleField.SetEnabled(true)
				d.scaleField.SetText(d.scaleField.Format(desiredScaleInt))
				d.scaleField.SetEnabled(false)
			}
		case autoscale.FitPage:
			size := d.page.Image.LogicalSize()
			cSize := d.docScroll.ContentView().ContentRect(false).Size
			desiredScale := xmath.Floor((min(cSize.Width/size.Width, cSize.Height/size.Height) / scaleCompensation) * 100)
			desiredScaleInt := int(min(max(desiredScale, minPDFDockableScale), maxPDFDockableScale))
			if d.scaleField.CurrentValue() != desiredScaleInt {
				d.scaleField.SetEnabled(true)
				d.scaleField.SetText(d.scaleField.Format(desiredScaleInt))
				d.scaleField.SetEnabled(false)
			}
		default:
		}
		gc.Save()
		gc.Scale(scaleCompensation, scaleCompensation)
		r := unison.Rect{Size: d.page.Image.LogicalSize()}
		gc.DrawRect(r, unison.White.Paint(gc, r, paintstyle.Fill))
		gc.DrawImageInRect(d.page.Image, r, &unison.SamplingOptions{
			UseCubic:       true,
			CubicResampler: unison.MitchellResampler(),
			FilterMode:     filtermode.Linear,
			MipMapMode:     mipmapmode.Linear,
		}, nil)
		if len(d.page.Matches) != 0 {
			p := unison.NewPaint()
			p.SetStyle(paintstyle.Fill)
			p.SetBlendMode(blendmode.Modulate)
			p.SetColor(adjustForModulate(unison.ThemeFocus.GetColor()))
			for _, match := range d.page.Matches {
				gc.DrawRect(match, p)
			}
		}
		if d.link != nil {
			p := unison.NewPaint()
			p.SetStyle(paintstyle.Fill)
			p.SetBlendMode(blendmode.Modulate)
			p.SetColor(adjustForModulate(unison.ThemeFocus.GetColor()))
			gc.DrawRect(d.rolloverRect, p)
		}
		gc.Restore()
	}
}

func adjustForModulate(c unison.Color) unison.Color {
	saturation := c.Saturation()
	if saturation > 0.5 {
		c = c.AdjustSaturation(-(saturation - 0.5))
	}
	lightness := c.PerceivedLightness()
	if lightness < 0.6 {
		c = c.AdjustPerceivedLightness(max(0.6-lightness, 0.2))
	}
	return c
}

func (d *PDFDockable) drawOverlay(gc *unison.Canvas, dirty unison.Rect) {
	if d.page != nil && d.page.Error != nil {
		d.drawOverlayMsg(gc, dirty, fmt.Sprintf("%s", d.page.Error), true) //nolint:gocritic // I want the extra processing %s does in this case
	}
	if finished, pageNumber, requested := d.pdf.RenderingFinished(); !finished {
		if waitFor := maxElapsedRenderTimeWithoutOverlay - time.Since(requested); waitFor > renderTimeSlop {
			unison.InvokeTaskAfter(d.MarkForRedraw, waitFor)
		} else {
			d.drawOverlayMsg(gc, dirty, fmt.Sprintf(i18n.Text("Rendering page %d…"), pageNumber+1), false)
		}
	}
}

func (d *PDFDockable) drawOverlayMsg(gc *unison.Canvas, dirty unison.Rect, msg string, forError bool) {
	var fgInk, bgInk unison.Ink
	var icon unison.Drawable
	font := unison.SystemFont.Face().Font(24)
	baseline := font.Baseline()
	if forError {
		fgInk = unison.ThemeOnError
		bgInk = unison.ThemeError.GetColor().SetAlphaIntensity(0.7)
		icon = &unison.DrawableSVG{
			SVG:  unison.CircledExclamationSVG,
			Size: unison.NewSize(baseline, baseline),
		}
	} else {
		fgInk = unison.ThemeOnSurface
		bgInk = unison.ThemeSurface.GetColor().SetAlphaIntensity(0.7)
	}
	decoration := &unison.TextDecoration{
		Font:            font,
		OnBackgroundInk: fgInk,
	}
	text := unison.NewText(msg, decoration)
	r := d.docScroll.ContentView().ContentRect(false)
	cy := r.CenterY()
	width := text.Width()
	height := text.Height()
	var iconSize unison.Size
	if icon != nil {
		iconSize = icon.LogicalSize()
		width += iconSize.Width + unison.StdHSpacing
		if height < iconSize.Height {
			height = iconSize.Height
		}
	}
	backWidth := width + 40
	backHeight := height + 40
	r.X += (r.Width - backWidth) / 2
	if forError {
		r.Y = cy - (backHeight + unison.StdVSpacing)
	} else {
		r.Y = cy + unison.StdVSpacing
	}
	r.Width = backWidth
	r.Height = backHeight
	gc.DrawRoundedRect(r, 10, 10, bgInk.Paint(gc, dirty, paintstyle.Fill))
	x := r.X + (r.Width-width)/2
	if icon != nil {
		icon.DrawInRect(gc, unison.NewRect(x, r.Y+(r.Height-iconSize.Height)/2, iconSize.Width, iconSize.Height), nil,
			decoration.OnBackgroundInk.Paint(gc, r, paintstyle.Fill))
		x += iconSize.Width + unison.StdHSpacing
	}
	text.Draw(gc, x, r.Y+(r.Height-height)/2+baseline)
}

// TitleIcon implements workspace.FileBackedDockable
func (d *PDFDockable) TitleIcon(suggestedSize unison.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  gurps.FileInfoFor(d.path).SVG,
		Size: suggestedSize,
	}
}

// Title implements workspace.FileBackedDockable
func (d *PDFDockable) Title() string {
	return fs.BaseName(d.path)
}

// Tooltip implements workspace.FileBackedDockable
func (d *PDFDockable) Tooltip() string {
	return d.path
}

// BackingFilePath implements workspace.FileBackedDockable
func (d *PDFDockable) BackingFilePath() string {
	return d.path
}

// SetBackingFilePath implements workspace.FileBackedDockable
func (d *PDFDockable) SetBackingFilePath(p string) {
	d.path = p
	UpdateTitleForDockable(d)
}

// Modified implements workspace.FileBackedDockable
func (d *PDFDockable) Modified() bool {
	return false
}

// MayAttemptClose implements unison.TabCloser
func (d *PDFDockable) MayAttemptClose() bool {
	return true
}

// AttemptClose implements unison.TabCloser
func (d *PDFDockable) AttemptClose() bool {
	return AttemptCloseForDockable(d)
}

func (d *PDFDockable) tocDoubleClick() {
	altered := false
	for _, row := range d.tocPanel.SelectedRows(false) {
		if row.CanHaveChildren() {
			altered = true
			row.SetOpen(!row.IsOpen())
		}
	}
	if altered {
		d.tocPanel.SyncToModel()
	}
}

func (d *PDFDockable) tocSelectionChanged() {
	if d.tocPanel.HasSelection() {
		d.LoadPage(d.tocPanel.SelectedRows(true)[0].pageNumber)
	}
}

func (d *PDFDockable) adjustTableSizeEventually() {
	if !d.adjustTableSizePending {
		d.adjustTableSizePending = true
		unison.InvokeTaskAfter(d.adjustTableSize, time.Millisecond)
	}
}

func (d *PDFDockable) adjustTableSize() {
	d.adjustTableSizePending = false
	d.tocPanel.SyncToModel()
	d.tocPanel.SizeColumnsToFit(true)
}
