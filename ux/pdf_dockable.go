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
	"fmt"
	"strconv"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/desktop"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/toolbox/xmath"
	"github.com/richardwilkes/unison"
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
	toolbar                *unison.Panel
	content                *unison.Panel
	docScroll              *unison.ScrollPanel
	docPanel               *unison.Panel
	tocScroll              *unison.ScrollPanel
	tocPanel               *unison.Table[*tocNode]
	divider                *unison.Panel
	tocScrollLayoutData    *unison.FlexLayoutData
	pageNumberField        *unison.Field
	scaleField             *PercentageField
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
	inDrag                 bool
	noUpdate               bool
	adjustTableSizePending bool
}

// NewPDFDockable creates a new unison.Dockable for PDFRenderer files.
func NewPDFDockable(filePath string) (unison.Dockable, error) {
	d := &PDFDockable{
		path:     filePath,
		scale:    100,
		noUpdate: true,
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
	d.createToolbar() // Has to be after content creation

	d.AddChild(d.toolbar)
	d.AddChild(d.content)
	d.content.AddChild(d.docScroll)

	d.noUpdate = false
	d.scaleField.SetEnabled(true)
	d.LoadPage(0)

	return d, nil
}

func (d *PDFDockable) createToolbar() {
	d.toolbar = unison.NewPanel()
	d.toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.Insets{Bottom: 1},
		false), unison.NewEmptyBorder(unison.StdInsets())))
	d.toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})

	d.sideBarButton = unison.NewSVGButton(svg.SideBar)
	d.sideBarButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Toggle the Sidebar"))
	d.sideBarButton.ClickCallback = d.toggleSideBar
	d.sideBarButton.SetEnabled(false)
	d.toolbar.AddChild(d.sideBarButton)

	info := NewInfoPop()
	AddHelpToInfoPop(info, i18n.Text("Within this view, these keys have the following effects:\n"))
	AddKeyBindingInfoToInfoPop(info, unison.KeyBinding{KeyCode: unison.KeyHome}, i18n.Text("Go to first page"))
	AddKeyBindingInfoToInfoPop(info, unison.KeyBinding{KeyCode: unison.KeyEnd}, i18n.Text("Go to last page"))
	AddKeyBindingInfoToInfoPop(info, unison.KeyBinding{KeyCode: unison.KeyLeft}, i18n.Text("Go to previous page"))
	AddKeyBindingInfoToInfoPop(info, unison.KeyBinding{KeyCode: unison.KeyUp}, i18n.Text("Go to previous page"))
	AddKeyBindingInfoToInfoPop(info, unison.KeyBinding{KeyCode: unison.KeyRight}, i18n.Text("Go to next page"))
	AddKeyBindingInfoToInfoPop(info, unison.KeyBinding{KeyCode: unison.KeyDown}, i18n.Text("Go to next page"))
	AddScalingHelpToInfoPop(info)
	d.toolbar.AddChild(info)

	d.scaleField = NewScaleField(
		minPDFDockableScale,
		maxPDFDockableScale,
		func() int { return 100 },
		func() int { return d.scale },
		func(scale int) { d.scale = scale },
		d.MarkForRedraw,
		false,
		d.docScroll,
	)
	d.scaleField.SetEnabled(false)
	d.toolbar.AddChild(d.scaleField)

	pageLabel := unison.NewLabel()
	pageLabel.Font = unison.DefaultFieldTheme.Font
	pageLabel.Text = i18n.Text("Page")
	d.toolbar.AddChild(pageLabel)

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
	d.toolbar.AddChild(d.pageNumberField)

	ofLabel := unison.NewLabel()
	ofLabel.Font = unison.DefaultFieldTheme.Font
	ofLabel.Text = fmt.Sprintf(i18n.Text("of %d"), d.pdf.PageCount())
	d.toolbar.AddChild(ofLabel)

	d.toolbar.AddChild(NewToolbarSeparator())

	d.backButton = unison.NewSVGButton(svg.Back)
	d.backButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Back"))
	d.backButton.ClickCallback = d.Back
	d.toolbar.AddChild(d.backButton)

	d.forwardButton = unison.NewSVGButton(svg.Forward)
	d.forwardButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Forward"))
	d.forwardButton.ClickCallback = d.Forward
	d.toolbar.AddChild(d.forwardButton)

	d.toolbar.AddChild(NewToolbarSeparator())

	d.firstPageButton = unison.NewSVGButton(svg.First)
	d.firstPageButton.Tooltip = unison.NewTooltipWithText(i18n.Text("First Page"))
	d.firstPageButton.ClickCallback = func() { d.LoadPage(0) }
	d.toolbar.AddChild(d.firstPageButton)

	d.previousPageButton = unison.NewSVGButton(svg.Previous)
	d.previousPageButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Previous Page"))
	d.previousPageButton.ClickCallback = func() { d.LoadPage(d.pdf.MostRecentPageNumber() - 1) }
	d.toolbar.AddChild(d.previousPageButton)

	d.nextPageButton = unison.NewSVGButton(svg.Next)
	d.nextPageButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Next Page"))
	d.nextPageButton.ClickCallback = func() { d.LoadPage(d.pdf.MostRecentPageNumber() + 1) }
	d.toolbar.AddChild(d.nextPageButton)

	d.lastPageButton = unison.NewSVGButton(svg.Last)
	d.lastPageButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Last Page"))
	d.lastPageButton.ClickCallback = func() { d.LoadPage(d.pdf.PageCount() - 1) }
	d.toolbar.AddChild(d.lastPageButton)

	d.searchField = NewSearchField()
	pageSearch := i18n.Text("Page Search")
	d.searchField.Watermark = pageSearch
	d.searchField.Tooltip = unison.NewTooltipWithText(pageSearch)
	d.searchField.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
		HGrab:  true,
	})
	d.searchField.ModifiedCallback = func(_, _ *unison.FieldState) {
		if d.noUpdate {
			return
		}
		d.LoadPage(d.pdf.MostRecentPageNumber())
	}
	d.toolbar.AddChild(d.searchField)

	d.matchesLabel = unison.NewLabel()
	d.matchesLabel.Text = "-"
	d.matchesLabel.Tooltip = unison.NewTooltipWithText(i18n.Text("Number of matches found"))
	d.toolbar.AddChild(d.matchesLabel)

	d.toolbar.SetLayout(&unison.FlexLayout{
		Columns:  len(d.toolbar.Children()),
		HSpacing: unison.StdHSpacing,
	})
}

func (d *PDFDockable) createTOC() {
	d.tocPanel = unison.NewTable[*tocNode](&unison.SimpleTableModel[*tocNode]{})
	d.tocPanel.Columns = make([]unison.ColumnInfo, 1)
	d.tocPanel.DoubleClickCallback = d.tocDoubleClick
	d.tocPanel.SelectionChangedCallback = d.tocSelectionChanged

	d.tocScroll = unison.NewScrollPanel()
	d.tocScrollLayoutData = &unison.FlexLayoutData{
		SizeHint: unison.Size{Width: 200},
		HAlign:   unison.FillAlignment,
		VAlign:   unison.FillAlignment,
		VGrab:    true,
	}
	d.tocScroll.SetLayoutData(d.tocScrollLayoutData)
	d.tocScroll.SetContent(d.tocPanel, unison.FillBehavior, unison.FillBehavior)

	d.divider = unison.NewPanel()
	d.divider.SetLayoutData(&unison.FlexLayoutData{
		SizeHint: unison.Size{Width: unison.DefaultDockTheme.DockDividerSize()},
		HAlign:   unison.FillAlignment,
		VAlign:   unison.FillAlignment,
		VGrab:    true,
	})
	d.divider.UpdateCursorCallback = func(_ unison.Point) *unison.Cursor {
		return unison.ResizeHorizontalCursor()
	}
	d.divider.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		r := d.divider.ContentRect(true)
		unison.DefaultDockTheme.DrawHorizontalGripper(gc, r)
	}
	var initialPosition float32
	var eventPosition float32
	d.divider.MouseDownCallback = func(where unison.Point, button, _ int, _ unison.Modifiers) bool {
		initialPosition = d.tocScrollLayoutData.SizeHint.Width
		eventPosition = d.divider.Parent().PointFromRoot(d.divider.PointToRoot(where)).X
		return true
	}
	d.divider.MouseDragCallback = func(where unison.Point, _ int, _ unison.Modifiers) bool {
		pos := eventPosition - d.divider.Parent().PointFromRoot(d.divider.PointToRoot(where)).X
		old := d.tocScrollLayoutData.SizeHint.Width
		d.tocScrollLayoutData.SizeHint.Width = xmath.Max(initialPosition-pos, 1)
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
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
		VGrab:  true,
	})
	d.docScroll.SetContent(d.docPanel, unison.FillBehavior, unison.FillBehavior)
	d.docScroll.ContentView().DrawOverCallback = d.drawOverlay

	d.content = unison.NewPanel()
	d.content.SetLayout(&unison.FlexLayout{
		Columns: 1,
		HAlign:  unison.FillAlignment,
		VAlign:  unison.FillAlignment,
	})
	d.content.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
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
		d.scaleField.SetEnabled(true)
	}()

	d.page = d.pdf.CurrentPage()

	if d.tocPanel.RootRowCount() == 0 {
		if toc := d.pdf.CurrentPage().TOC; len(toc) != 0 {
			d.tocPanel.SetRootRows(newTOC(d, nil, toc))
			d.tocPanel.SizeColumnsToFit(true)
			d.tocScrollLayoutData.SizeHint.Width = xmath.Max(xmath.Min(d.tocPanel.Columns[0].Current, 300), d.tocScrollLayoutData.SizeHint.Width)
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
	if matchText != d.matchesLabel.Text {
		d.matchesLabel.Text = matchText
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
}

func (d *PDFDockable) overLink(where unison.Point) (rect unison.Rect, link *PDFLink) {
	if d.page != nil && d.page.Links != nil {
		where.X /= scaleCompensation
		where.Y /= scaleCompensation
		for _, link = range d.page.Links {
			if link.Bounds.ContainsPoint(where) {
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

func (d *PDFDockable) updateCursor(_ unison.Point) *unison.Cursor {
	if d.inDrag {
		return unison.MoveCursor()
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
		pt := d.dragStart
		pt.Subtract(d.docPanel.PointToRoot(where))
		d.docScroll.SetPosition(d.dragOrigin.X+pt.X, d.dragOrigin.Y+pt.Y)
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
				unison.ErrorDialogWithError(i18n.Text("Unable to open link"), err)
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

func (d *PDFDockable) docSizer(_ unison.Size) (min, pref, max unison.Size) {
	if d.page == nil || d.page.Error != nil {
		pref.Width = 400
		pref.Height = 300
	} else {
		pref = d.page.Image.LogicalSize()
		pref.Width *= scaleCompensation
		pref.Height *= scaleCompensation
	}
	return unison.NewSize(50, 50), pref, unison.MaxSize(pref)
}

func (d *PDFDockable) draw(gc *unison.Canvas, dirty unison.Rect) {
	gc.DrawRect(dirty, unison.ContentColor.Paint(gc, dirty, unison.Fill))
	if d.page != nil && d.page.Image != nil {
		gc.Save()
		gc.Scale(scaleCompensation, scaleCompensation)
		r := unison.Rect{Size: d.page.Image.LogicalSize()}
		gc.DrawRect(r, unison.White.Paint(gc, r, unison.Fill))
		gc.DrawImageInRect(d.page.Image, r, &unison.SamplingOptions{
			UseCubic:       true,
			CubicResampler: unison.MitchellResampler(),
			FilterMode:     unison.FilterModeLinear,
			MipMapMode:     unison.MipMapModeLinear,
		}, nil)
		if len(d.page.Matches) != 0 {
			p := unison.NewPaint()
			p.SetStyle(unison.Fill)
			p.SetBlendMode(unison.ModulateBlendMode)
			p.SetColor(gurps.PDFMarkerHighlightColor.GetColor())
			for _, match := range d.page.Matches {
				gc.DrawRect(match, p)
			}
		}
		if d.link != nil {
			p := unison.NewPaint()
			p.SetStyle(unison.Fill)
			p.SetBlendMode(unison.ModulateBlendMode)
			p.SetColor(gurps.PDFLinkHighlightColor.GetColor())
			gc.DrawRect(d.rolloverRect, p)
		}
		gc.Restore()
	}
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
		fgInk = unison.OnErrorColor
		bgInk = unison.ErrorColor.GetColor().SetAlphaIntensity(0.7)
		icon = &unison.DrawableSVG{
			SVG:  unison.CircledExclamationSVG,
			Size: unison.NewSize(baseline, baseline),
		}
	} else {
		fgInk = unison.OnContentColor
		bgInk = unison.ContentColor.GetColor().SetAlphaIntensity(0.7)
	}
	decoration := &unison.TextDecoration{
		Font:       font,
		Foreground: fgInk,
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
	gc.DrawRoundedRect(r, 10, 10, bgInk.Paint(gc, dirty, unison.Fill))
	x := r.X + (r.Width-width)/2
	if icon != nil {
		icon.DrawInRect(gc, unison.NewRect(x, r.Y+(r.Height-iconSize.Height)/2, iconSize.Width, iconSize.Height), nil,
			decoration.Foreground.Paint(gc, r, unison.Fill))
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
	if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
		dc.UpdateTitle(d)
	}
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
	if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
		dc.Close(d)
	}
	return true
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
