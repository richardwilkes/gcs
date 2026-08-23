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
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/autoscale"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xfilepath"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/blendmode"
	"github.com/richardwilkes/unison/enums/filtermode"
	"github.com/richardwilkes/unison/enums/mipmapmode"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

const (
	minPDFDockableScale                = 50
	maxPDFDockableScale                = 200
	scaleAdj                           = float32(100) / maxPDFDockableScale
	pdfPageGap                         = float32(8)
	maxElapsedRenderTimeWithoutOverlay = time.Second / 2
	renderTimeSlop                     = time.Millisecond * 10
)

var (
	_ FileBackedDockable     = &PDFDockable{}
	_ unison.TabCloser       = &PDFDockable{}
	_ boundsDeferredDockable = &PDFDockable{}
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
	link                   *PDFLink
	loadErr                error
	labelToPageNumMap      map[string]int
	pageLabels             []string
	pageRects              []geom.Rect
	history                []int
	boundsKnownCallbacks   []func()
	initialPageInfo        gurps.PageInfo
	currentPage            int
	pendingScrollPage      int
	linkPage               int
	scale                  int
	historyPos             int
	loadStarted            time.Time
	docSize                geom.Size
	provisionalDocSize     geom.Size
	rolloverRect           geom.Rect
	dragStart              geom.Point
	dragOrigin             geom.Point
	autoScaling            autoscale.Option
	inDrag                 bool
	noUpdate               bool
	adjustTableSizePending bool
	viewSyncPending        bool
	closed                 bool
}

// NewPDFDockable creates a new unison.Dockable for PDFRenderer files. The document itself is not loaded here: opening
// it can take anywhere from a moment to several minutes -- the worst case being a large file that the OS has to fetch
// back from cloud storage before it can be read -- and doing that on the UI thread makes the whole application
// unresponsive before its window has even appeared. Instead, a fully-formed dockable in a loading state is returned
// immediately and the document is prepared on a background goroutine. As a result, this never fails; a document that
// can't be loaded reports the failure within the dockable rather than by returning an error.
func NewPDFDockable(filePath string, initialPageInfo gurps.PageInfo) (unison.Dockable, error) {
	generalSettings := gurps.GlobalSettings().General
	d := &PDFDockable{
		path:              filePath,
		initialPageInfo:   initialPageInfo,
		scale:             generalSettings.InitialPDFUIScale,
		autoScaling:       generalSettings.PDFAutoScaling,
		noUpdate:          true,
		pendingScrollPage: -1,
		linkPage:          -1,
		loadStarted:       time.Now(),
	}
	d.Self = d
	d.KeyDownCallback = d.keyDown
	// Both of these ask the renderer to claim the render queue's priority, but d.pdf is nil until the load completes,
	// so they have to go through a nil-checked helper. In particular, GainedFocusCallback can't be the method value
	// d.pdf.RequestRenderPriority any longer: a method value binds its receiver where it is taken, which would capture
	// the nil renderer here and then dereference it the first time the view was focused.
	d.FocusChangeInHierarchyCallback = d.focusChangeInHierarchy
	d.GainedFocusCallback = d.requestRenderPriority
	d.SetLayout(&unison.FlexLayout{Columns: 1})

	d.createTOC()
	d.createContent()

	d.AddChild(d.createToolbar()) // Creation of the toolbar has to be after content creation
	d.AddChild(d.content)
	d.content.AddChild(d.docScroll)

	d.noUpdate = false
	d.scaleField.SetEnabled(d.autoScaling == autoscale.No)

	// Both of these ultimately call unison.PrimaryDisplay(), which is only safe on the UI thread, so they are resolved
	// here and handed to the renderer rather than being looked up by the goroutine below.
	ppi := float32(generalSettings.MonitorPPI())
	scaleAdjust := geom.NewPoint(1, 1).DivPt(primaryDisplayScale())
	d.provisionalDocSize = provisionalPDFDocSize(ppi, scaleAdjust)
	go d.load(filePath, ppi, scaleAdjust)

	return d, nil
}

// load prepares the document on a background goroutine and then hands the result back to the UI thread. Nothing in
// here may touch the dockable's fields or any other UI object: the file path and the display-derived values are passed
// in as parameters for that reason, and everything that has to be done with the result happens in loadCompleted.
func (d *PDFDockable) load(filePath string, ppi float32, scaleAdjust geom.Point) {
	pdf, err := NewPDFRenderer(filePath, ppi, scaleAdjust, func(_ int) {
		unison.InvokeTask(d.pageRendered)
	})
	unison.InvokeTask(func() { d.loadCompleted(pdf, err) })
}

// loadCompleted installs the freshly prepared renderer, or records the failure that prevented one from being made. It
// runs on the UI thread, since it touches the dockable's fields and the widgets that were built against a page count
// that wasn't known yet.
func (d *PDFDockable) loadCompleted(pdf *PDFRenderer, err error) {
	if d.closed {
		if pdf != nil {
			pdf.Close()
		}
		d.boundsKnownCallbacks = nil
		return
	}
	if err != nil {
		errs.Log(err, "path", d.path)
		d.loadErr = err
		d.MarkForRedraw()
		d.notifyBoundsKnown()
		return
	}
	d.pdf = pdf
	d.buildPageLayout()
	d.buildPageLabelCache()
	d.syncPageCountDependents()
	d.MarkForLayoutAndRedraw()
	// This has to come before the scroll below: the callbacks are what create the window for a dockable whose placement
	// was deferred until its size was known, and a scroll can only be applied once there is a window to apply it
	// within. In this order the initially requested page is scrolled to right now, rather than having to wait for a
	// later layout pass to notice the pending request.
	d.notifyBoundsKnown()
	// The scroll to the initially requested page couldn't be honored at construction time, since there were no pages to
	// scroll to yet, so it is issued now.
	d.LoadPage(d.initialPageInfo)
	if d.HasInSelfOrDescendants(func(p *unison.Panel) bool { return p.Focused() }) {
		// The focus arrived while the document was still loading, so the claim the focus callbacks would have made was
		// dropped on the floor. Make it now, so a document that was opened and focused normally renders ahead of the
		// ones merely catching up in the background.
		d.pdf.RequestRenderPriority()
	}
}

// BoundsKnown implements boundsDeferredDockable. The size of the document isn't known until it has been loaded, at
// which point the provisional size docSizer reports gives way to the real one. A document that failed to load never
// gets a real size, but it is never going to get one, either, so it counts as known.
func (d *PDFDockable) BoundsKnown() bool {
	return d.pdf != nil || d.loadErr != nil
}

// WhenBoundsKnown implements boundsDeferredDockable. If the bounds are already known, the callback is invoked before
// returning.
func (d *PDFDockable) WhenBoundsKnown(callback func()) {
	if d.BoundsKnown() {
		callback()
		return
	}
	d.boundsKnownCallbacks = append(d.boundsKnownCallbacks, callback)
	if len(d.boundsKnownCallbacks) == 1 {
		// Waiting on a document that turns out to be slow can't be allowed to go on indefinitely, so the wait is capped
		// at the same threshold the loading overlay uses, measured from the same moment. Whatever is waiting therefore
		// proceeds precisely when the overlay would appear: a window created at that point is sized to the provisional
		// page docSizer reports and shows the user that the document is on its way, rather than nothing at all having
		// happened. The load isn't abandoned; it simply stops being waited on. Only the first registration arms this,
		// since the cap is on how long the document may take, not on how long any one caller has been waiting.
		unison.InvokeTaskAfter(d.notifyBoundsKnown,
			max(maxElapsedRenderTimeWithoutOverlay-time.Since(d.loadStarted), 0))
	}
}

// notifyBoundsKnown invokes the callbacks registered through WhenBoundsKnown and discards them. The list is emptied
// before anything is invoked, so each callback runs exactly once no matter how the load completion and the cap on the
// wait interleave. Nothing is invoked once the dockable has been closed: there is no longer a view for a callback to
// act upon, so the registrations are simply dropped.
func (d *PDFDockable) notifyBoundsKnown() {
	callbacks := d.boundsKnownCallbacks
	d.boundsKnownCallbacks = nil
	if d.closed {
		return
	}
	for _, callback := range callbacks {
		callback()
	}
}

// requestRenderPriority claims the render queue's priority for this document, if it has finished loading.
func (d *PDFDockable) requestRenderPriority() {
	if d.pdf != nil {
		d.pdf.RequestRenderPriority()
	}
}

// pageCount returns the number of pages in the document, which is zero until it has finished loading and stays zero if
// it failed to load.
func (d *PDFDockable) pageCount() int {
	if d.pdf == nil {
		return 0
	}
	return d.pdf.PageCount()
}

// buildPageLayout computes the area each page occupies within the document panel. Pages are stacked vertically,
// separated and surrounded by pdfPageGap, with each page horizontally centered within the width of the widest page.
// The sizes come from the renderer's precomputed logical sizes, so this can be done without rendering anything and a
// page's image lands exactly within the slot reserved for it.
func (d *PDFDockable) buildPageLayout() {
	count := d.pdf.PageCount()
	d.pageRects = make([]geom.Rect, count)
	var widest float32
	for i := range count {
		widest = max(widest, d.pdf.PageLogicalSize(i).Width*scaleAdj)
	}
	docWidth := widest + 2*pdfPageGap
	y := pdfPageGap
	for i := range count {
		size := d.pdf.PageLogicalSize(i).Mul(scaleAdj)
		d.pageRects[i] = geom.NewRect(pdfPageGap+(docWidth-2*pdfPageGap-size.Width)/2, y, size.Width, size.Height)
		y += size.Height + pdfPageGap
	}
	d.docSize = geom.NewSize(docWidth, y)
}

func (d *PDFDockable) buildPageLabelCache() {
	d.pageLabels = d.pdf.doc.PageLabels()
	d.labelToPageNumMap = make(map[string]int)
	for pageNum, label := range d.pageLabels {
		label = normalizePageLabel(label)
		if _, exists := d.labelToPageNumMap[label]; !exists {
			d.labelToPageNumMap[label] = pageNum
		}
	}
}

func normalizePageLabel(label string) string {
	return strings.ToLower(strings.TrimSpace(label))
}

func (d *PDFDockable) pageLabel(pageNum int) string {
	if pageNum < 0 || pageNum >= len(d.pageLabels) {
		return ""
	}
	return d.pageLabels[pageNum]
}

// DockKey implements KeyedDockable.
func (d *PDFDockable) DockKey() string {
	return filePrefix + d.path
}

func (d *PDFDockable) createToolbar() *unison.Panel {
	outer := unison.NewPanel()
	outer.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{},
		geom.Insets{Bottom: 1}, false), unison.NewEmptyBorder(unison.StdInsets())))
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
		func() {
			d.MarkForRedraw()
			d.scheduleViewSync()
		},
		true,
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
			d.scheduleViewSync()
		}
	}
	first.AddChild(d.autoScalingPopup)

	pageLabel := unison.NewLabel()
	pageLabel.Font = unison.DefaultFieldTheme.Font
	pageLabel.SetTitle(i18n.Text("Page"))
	first.AddChild(pageLabel)

	// The toolbar is built before the document has been loaded, so everything that depends on the page count starts out
	// reflecting a count of zero and is brought up to date by syncPageCountDependents once the load completes. The
	// navigation widgets are created disabled for the same reason: there is nowhere to navigate to yet, and a document
	// that fails to load leaves them that way. syncViewState gives them their real state on the first sync after the
	// document has been laid out.
	d.pageNumberField = unison.NewField()
	d.pageNumberField.SetEnabled(false)
	d.pageNumberField.SetMinimumTextWidthUsing("1000")
	d.pageNumberField.ModifiedCallback = func(_, after *unison.FieldState) {
		if d.noUpdate || d.pdf == nil {
			return
		}
		if pageNum, exists := d.labelToPageNumMap[normalizePageLabel(after.Text)]; exists {
			d.ScrollToPage(pageNum, true)
		}
	}
	d.pageNumberField.ValidateCallback = func() bool {
		if d.pdf == nil {
			return true
		}
		_, exists := d.labelToPageNumMap[normalizePageLabel(d.pageNumberField.Text())]
		return exists
	}
	first.AddChild(d.pageNumberField)

	d.backButton = unison.NewSVGButton(svg.Back)
	d.backButton.Tooltip = newWrappedTooltip(i18n.Text("Back"))
	d.backButton.ClickCallback = d.Back
	d.backButton.SetEnabled(false)
	first.AddChild(d.backButton)

	d.forwardButton = unison.NewSVGButton(svg.Forward)
	d.forwardButton.Tooltip = newWrappedTooltip(i18n.Text("Forward"))
	d.forwardButton.ClickCallback = d.Forward
	d.forwardButton.SetEnabled(false)
	first.AddChild(d.forwardButton)

	d.firstPageButton = unison.NewSVGButton(svg.First)
	d.firstPageButton.Tooltip = newWrappedTooltip(i18n.Text("First Page"))
	d.firstPageButton.ClickCallback = func() { d.ScrollToPage(0, true) }
	d.firstPageButton.SetEnabled(false)
	first.AddChild(d.firstPageButton)

	d.previousPageButton = unison.NewSVGButton(svg.Previous)
	d.previousPageButton.Tooltip = newWrappedTooltip(i18n.Text("Previous Page"))
	d.previousPageButton.ClickCallback = func() { d.ScrollToPage(d.currentPage-1, true) }
	d.previousPageButton.SetEnabled(false)
	first.AddChild(d.previousPageButton)

	d.nextPageButton = unison.NewSVGButton(svg.Next)
	d.nextPageButton.Tooltip = newWrappedTooltip(i18n.Text("Next Page"))
	d.nextPageButton.ClickCallback = func() { d.ScrollToPage(d.currentPage+1, true) }
	d.nextPageButton.SetEnabled(false)
	first.AddChild(d.nextPageButton)

	d.lastPageButton = unison.NewSVGButton(svg.Last)
	d.lastPageButton.Tooltip = newWrappedTooltip(i18n.Text("Last Page"))
	d.lastPageButton.ClickCallback = func() { d.ScrollToPage(d.pageCount()-1, true) }
	d.lastPageButton.SetEnabled(false)
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
		d.MarkForRedraw()
		d.scheduleViewSync()
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

// syncPageCountDependents brings the toolbar widgets whose appearance depends on the page count into agreement with
// it. They are created while the document is still loading, so the page count they were built against was zero.
func (d *PDFDockable) syncPageCountDependents() {
	d.pageNumberField.SetEnabled(true)
	d.pageNumberField.SetMinimumTextWidthUsing(d.pageLabels...)
	d.pageNumberField.Parent().MarkForLayoutAndRedraw()
}

func (d *PDFDockable) createTOC() {
	d.tocPanel = unison.NewTable(&unison.SimpleTableModel[*tocNode]{})
	d.tocPanel.Columns = make([]unison.ColumnInfo, 1)
	d.tocPanel.ShowFirstColumnDivider = false
	d.tocPanel.ShowLastColumnDivider = false
	d.tocPanel.PreventUserColumnResize = true
	d.tocPanel.DoubleClickCallback = d.tocDoubleClick
	d.tocPanel.SelectionChangedCallback = d.tocSelectionChanged

	d.tocScroll = unison.NewScrollPanel()
	d.tocScrollLayoutData = &unison.FlexLayoutData{
		SizeHint: geom.Size{Width: 200},
		HAlign:   align.Fill,
		VAlign:   align.Fill,
		VGrab:    true,
	}
	d.tocScroll.SetLayoutData(d.tocScrollLayoutData)
	d.tocScroll.SetContent(d.tocPanel, behavior.Fill, behavior.Fill)

	d.divider = unison.NewPanel()
	d.divider.SetLayoutData(&unison.FlexLayoutData{
		SizeHint: geom.Size{Width: unison.DefaultDockTheme.DockDividerSize()},
		HAlign:   align.Fill,
		VAlign:   align.Fill,
		VGrab:    true,
	})
	d.divider.UpdateCursorCallback = func(_ geom.Point) *unison.Cursor {
		return unison.ResizeHorizontalCursor()
	}
	d.divider.DrawCallback = func(gc *unison.Canvas, _ geom.Rect) {
		unison.DefaultDockTheme.DrawHorizontalGripper(gc, d.divider.ContentRect(true))
	}
	var initialPosition float32
	var eventPosition float32
	d.divider.MouseDownCallback = func(where geom.Point, _, _ int, _ mod.Modifiers) bool {
		initialPosition = d.tocScrollLayoutData.SizeHint.Width
		eventPosition = d.divider.Parent().PointFromRoot(d.divider.PointToRoot(where)).X
		return true
	}
	d.divider.MouseDragCallback = func(where geom.Point, _ int, _ mod.Modifiers) bool {
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
	// A panel's layout takes precedence over its sizer, and a ScrollPanel is its own layout, so the one-page preferred
	// size has to be provided by wrapping that layout rather than by setting a sizer.
	d.docScroll.SetLayout(&pdfScrollLayout{d: d})
	d.docScroll.ContentView().DrawOverCallback = d.drawOverlay

	// The viewport's frame changes both when it is first laid out and whenever it is resized, either of which alters
	// which pages are visible.
	d.docScroll.ContentView().FrameChangeCallback = d.scheduleViewSync

	// There is no scroll-changed notification, so chain the vertical scroll bar's callback, which unison has already
	// pointed at the scroll panel's Sync. That has to keep running, so call it first.
	verticalBar := d.docScroll.Bar(false)
	syncScrollPanel := verticalBar.ChangedCallback
	verticalBar.ChangedCallback = func() {
		if syncScrollPanel != nil {
			syncScrollPanel()
		}
		d.scheduleViewSync()
	}

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
		d.ScrollToPage(d.history[d.historyPos], false)
	}
}

// Forward moves forward in history one step.
func (d *PDFDockable) Forward() {
	if d.historyPos < len(d.history)-1 {
		d.historyPos++
		d.ScrollToPage(d.history[d.historyPos], false)
	}
}

// LoadPage scrolls to the specified page, recording the jump in the history.
func (d *PDFDockable) LoadPage(pageInfo gurps.PageInfo) {
	if d.pdf == nil {
		// The document hasn't finished loading, so there is nothing to scroll to yet. Hold onto the request and let the
		// load completion issue it, which is the same path the page the dockable was created for takes. This matters
		// for a page reference that targets a document which is already open but still loading: the jump it asks for
		// must supersede the one the dockable was opened with rather than being dropped.
		d.initialPageInfo = pageInfo
		return
	}
	pageNum, exists := d.labelToPageNumMap[normalizePageLabel(pageInfo.Label)]
	if exists {
		pageNum = min(max(pageNum+pageInfo.Offset, 0), d.pdf.PageCount()-1)
	} else {
		pageNum = 0
	}
	d.ScrollToPage(pageNum, true)
}

// ScrollToPage scrolls the view so that the top of the given 0-based page is at the top of the view. If recordHistory
// is true, the jump is added to the navigation history, which is what distinguishes an explicit jump from ordinary
// scrolling.
func (d *PDFDockable) ScrollToPage(pageNumber int, recordHistory bool) {
	if len(d.pageRects) == 0 {
		return
	}
	pageNumber = min(max(pageNumber, 0), len(d.pageRects)-1)
	if recordHistory {
		d.recordHistory(pageNumber)
	}
	d.pendingScrollPage = pageNumber
	d.flushPendingScroll()
	d.scheduleViewSync()
}

// flushPendingScroll applies a pending scroll request, returning true if it was applied. Setting the scroll position
// before the scroll panel has been laid out silently clamps it to zero, since the scroll bar's extent and maximum are
// only established during layout, so the request is held until a layout has occurred.
func (d *PDFDockable) flushPendingScroll() bool {
	if d.pendingScrollPage < 0 {
		return true
	}
	if d.Window() == nil {
		return false
	}
	d.docScroll.MarkForLayoutRecursively()
	d.docScroll.ValidateLayout()
	if d.docScroll.ContentView().ContentRect(false).Height <= 0 {
		return false
	}
	pageNumber := d.pendingScrollPage
	d.pendingScrollPage = -1
	h, _ := d.docScroll.Position()
	d.docScroll.SetPosition(h, (d.pageRects[pageNumber].Y-pdfPageGap)*d.panelScale())
	return true
}

func (d *PDFDockable) recordHistory(pageNumber int) {
	if d.history == nil {
		d.history = []int{pageNumber}
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
	d.backButton.SetEnabled(d.historyPos > 0)
	d.forwardButton.SetEnabled(d.historyPos < len(d.history)-1)
}

// scheduleViewSync requests a state sync with the current view. Multiple requests made before the sync runs collapse
// into a single one.
func (d *PDFDockable) scheduleViewSync() {
	if !d.viewSyncPending {
		d.viewSyncPending = true
		unison.InvokeTask(d.syncViewState)
	}
}

// syncViewState brings everything that depends on which pages are visible back into agreement with the view. This is
// the only place that state is updated in response to scrolling, resizing and rendering.
func (d *PDFDockable) syncViewState() {
	d.viewSyncPending = false
	if len(d.pageRects) == 0 {
		return
	}
	if !d.flushPendingScroll() {
		// The scroll panel hasn't been laid out yet, so the target page can't be scrolled to. Get it rendering in the
		// meantime; the layout that eventually occurs will trigger another sync.
		d.pdf.SetWantedPages([]int{d.pendingScrollPage}, d.searchField.Text())
		return
	}

	first, limit := d.visiblePages()
	d.currentPage = first
	if d.history == nil {
		// Establish the history root, so that a later jump elsewhere can navigate back to where the user was.
		d.history = []int{d.currentPage}
		d.historyPos = 0
	}

	noUpdate := d.noUpdate
	d.noUpdate = true
	defer func() { d.noUpdate = noUpdate }()

	if d.pdf != nil {
		pageText := d.pageLabel(d.currentPage)
		if pageText != d.pageNumberField.Text() {
			d.pageNumberField.SetText(pageText)
			d.pageNumberField.Parent().MarkForLayoutAndRedraw()
		}
	}

	matchText := "-"
	if d.searchField.Text() != "" {
		// A page that is still pending may nonetheless have something in the cache: the placeholder rendered under the
		// previous search text, whose matches have nothing to do with what is in the search field now. Show nothing
		// for the count until the re-render for the current search has landed.
		if _, pending := d.pdf.PendingSince(d.currentPage); !pending {
			if page := d.pdf.CachedPage(d.currentPage); page != nil {
				matchText = strconv.Itoa(len(page.Matches))
			}
		}
	}
	if matchText != d.matchesLabel.Text.String() {
		d.matchesLabel.SetTitle(matchText)
		d.matchesLabel.Parent().MarkForLayoutAndRedraw()
	}

	lastPageNumber := len(d.pageRects) - 1
	d.backButton.SetEnabled(d.historyPos > 0)
	d.forwardButton.SetEnabled(d.historyPos < len(d.history)-1)
	d.firstPageButton.SetEnabled(d.currentPage != 0)
	d.previousPageButton.SetEnabled(d.currentPage > 0)
	d.nextPageButton.SetEnabled(d.currentPage < lastPageNumber)
	d.lastPageButton.SetEnabled(d.currentPage != lastPageNumber)

	d.applyAutoScaling()

	want := make([]int, 0, (limit-first)+2)
	for i := first; i < limit; i++ {
		want = append(want, i)
	}
	if limit <= lastPageNumber {
		want = append(want, limit)
	}
	if first > 0 {
		want = append(want, first-1)
	}
	d.pdf.SetWantedPages(want, d.searchField.Text())
}

// applyAutoScaling adjusts the scale to fit the current page, if auto-scaling is on. The scale is only altered when it
// actually differs from what is wanted, which is what stops the resize-triggers-rescale-triggers-resize loop.
func (d *PDFDockable) applyAutoScaling() {
	var desiredScale float32
	slotSize := d.pdf.PageLogicalSize(d.currentPage).Mul(scaleAdj)
	viewSize := d.docScroll.ContentView().ContentRect(false).Size
	switch d.autoScaling {
	case autoscale.FitWidth:
		desiredScale = xmath.Floor(100 * viewSize.Width / (slotSize.Width + 2*pdfPageGap))
	case autoscale.FitPage:
		desiredScale = xmath.Floor(100 * min(viewSize.Width/(slotSize.Width+2*pdfPageGap),
			viewSize.Height/(slotSize.Height+2*pdfPageGap)))
	default:
		return
	}
	desiredScaleInt := int(min(max(desiredScale, minPDFDockableScale), maxPDFDockableScale))
	if d.scaleField.CurrentValue() != desiredScaleInt {
		d.scaleField.SetEnabled(true)
		d.scaleField.SetText(d.scaleField.Format(desiredScaleInt))
		d.scaleField.SetEnabled(false)
	}
}

func (d *PDFDockable) pageRendered() {
	if d.pdf == nil {
		// Nothing can have been rendered before the renderer is installed, so this can't happen today, but the callback
		// this arrives through is handed to the renderer while it is still being constructed, so guard anyway.
		return
	}
	if d.tocPanel.RootRowCount() == 0 {
		if toc := d.pdf.TOC(); len(toc) != 0 {
			d.tocPanel.SetRootRows(newTOC(d, nil, toc))
			d.tocPanel.SizeColumnsToFit(true)
			d.tocScrollLayoutData.SizeHint.Width = max(min(d.tocPanel.Columns[0].Current, 300), d.tocScrollLayoutData.SizeHint.Width)
			d.sideBarButton.SetEnabled(true)
		}
	}
	d.MarkForRedraw()
	d.scheduleViewSync()
}

// panelScale returns the scale that has been applied to the document panel.
func (d *PDFDockable) panelScale() float32 {
	return float32(d.scale) / 100
}

// horizontalOffset returns the amount the pages have to be shifted right to stay centered when the document panel has
// been made wider than the document itself. unison stores a panel's frame size in local, unscaled units -- SetFrameRect
// divides the size by the panel's scale and FrameRect multiplies it back -- so ContentRect is already in the same
// coordinate space as the page layout and must not be adjusted by the zoom scale.
func (d *PDFDockable) horizontalOffset() float32 {
	return max((d.docPanel.ContentRect(false).Width-d.docSize.Width)/2, 0)
}

// visiblePages returns the half-open range of pages currently visible in the viewport. The first of them is the page
// considered to be the current one. At least one page is always returned.
func (d *PDFDockable) visiblePages() (first, limit int) {
	panelScale := d.panelScale()
	top := d.docScroll.Bar(false).Value() / panelScale
	bottom := top + d.docScroll.ContentView().ContentRect(false).Height/panelScale
	// A page whose bottom edge is only just below the top of the viewport isn't meaningfully visible, so bias past it.
	return d.pageRange(top+1, bottom)
}

// pageRange returns the half-open range of pages that intersect the given vertical span, which is expressed in
// document panel coordinates. At least one page is always returned.
func (d *PDFDockable) pageRange(top, bottom float32) (first, limit int) {
	count := len(d.pageRects)
	first = min(sort.Search(count, func(i int) bool { return d.pageRects[i].Bottom() > top }), count-1)
	limit = max(sort.Search(count, func(i int) bool { return d.pageRects[i].Y >= bottom }), first+1)
	return first, limit
}

// pageAt returns the index of the page whose vertical band contains the given document panel-local y coordinate, or -1
// if it is below the last page. The gap above a page counts as part of that page.
func (d *PDFDockable) pageAt(y float32) int {
	i := sort.Search(len(d.pageRects), func(i int) bool { return d.pageRects[i].Bottom() > y })
	if i >= len(d.pageRects) {
		return -1
	}
	return i
}

func (d *PDFDockable) overLink(where geom.Point) (rect geom.Rect, link *PDFLink, pageNumber int) {
	pageNumber = d.pageAt(where.Y)
	if pageNumber < 0 {
		return rect, nil, -1
	}
	page := d.pdf.CachedPage(pageNumber)
	if page == nil {
		return rect, nil, -1
	}
	pt := where.Sub(d.pageRects[pageNumber].Point).Sub(geom.NewPoint(d.horizontalOffset(), 0)).Div(scaleAdj)
	for _, one := range page.Links {
		if pt.In(one.Bounds) {
			return one.Bounds, one, pageNumber
		}
	}
	return rect, nil, -1
}

func (d *PDFDockable) checkForLinkAt(where geom.Point) bool {
	r, link, pageNumber := d.overLink(where)
	if r != d.rolloverRect || link != d.link || pageNumber != d.linkPage {
		d.rolloverRect = r
		d.link = link
		d.linkPage = pageNumber
		d.MarkForRedraw()
	}
	return link != nil
}

func (d *PDFDockable) updateCursor(pt geom.Point) *unison.Cursor {
	if d.inDrag {
		return unison.MoveCursor()
	}
	if _, link, _ := d.overLink(pt); link != nil {
		return unison.PointingCursor()
	}
	return unison.ArrowCursor()
}

func (d *PDFDockable) mouseDown(where geom.Point, _, _ int, _ mod.Modifiers) bool {
	d.dragStart = d.docPanel.PointToRoot(where)
	d.dragOrigin.X, d.dragOrigin.Y = d.docScroll.Position()
	d.inDrag = !d.checkForLinkAt(where)
	d.docPanel.RequestFocus()
	d.UpdateCursorNow()
	return true
}

func (d *PDFDockable) mouseDrag(where geom.Point, _ int, _ mod.Modifiers) bool {
	if d.inDrag {
		pt := d.dragStart.Sub(d.docPanel.PointToRoot(where)).Add(d.dragOrigin)
		d.docScroll.SetPosition(pt.X, pt.Y)
	} else {
		d.checkForLinkAt(where)
	}
	return true
}

func (d *PDFDockable) mouseMove(where geom.Point, _ mod.Modifiers) bool {
	d.checkForLinkAt(where)
	return true
}

func (d *PDFDockable) mouseUp(where geom.Point, button int, _ mod.Modifiers) bool {
	if d.inDrag {
		d.inDrag = false
		d.UpdateCursorNow()
	} else {
		d.checkForLinkAt(where)
		if button == unison.ButtonLeft && d.link != nil {
			if d.link.PageNumber >= 0 {
				d.ScrollToPage(d.link.PageNumber, true)
			} else if err := xos.OpenBrowser(d.link.URI); err != nil {
				Workspace.ErrorHandler(i18n.Text("Unable to open link"), err)
			}
		}
	}
	return true
}

func (d *PDFDockable) focusChangeInHierarchy(_, _ *unison.Panel) {
	d.requestRenderPriority()
}

func (d *PDFDockable) keyDown(keyCode unison.KeyCode, _ mod.Modifiers, _ bool) bool {
	if len(d.pageRects) == 0 {
		// There is nothing to navigate within while the document is loading or after it has failed to load, so let the
		// keys go on to whatever else might want them.
		return false
	}
	switch keyCode {
	case unison.KeyHome:
		d.ScrollToPage(0, true)
	case unison.KeyEnd:
		d.ScrollToPage(len(d.pageRects)-1, true)
	case unison.KeyLeft, unison.KeyUp:
		d.ScrollToPage(d.currentPage-1, true)
	case unison.KeyRight, unison.KeyDown:
		d.ScrollToPage(d.currentPage+1, true)
	default:
		return false
	}
	return true
}

// docSizer reports the size of the document panel, which is the size of the document laid out by buildPageLayout. Until
// the document has been loaded there is no layout to report, so a provisional single-page size stands in for it, so
// that anything sizing itself to fit the dockable -- a window opened for it, in particular -- has something
// page-shaped to work with rather than nothing at all. A document that failed to load keeps the provisional size,
// since no real one will ever be known for it.
func (d *PDFDockable) docSizer(_ geom.Size) (minSize, prefSize, maxSize geom.Size) {
	prefSize = d.docSize
	if d.pdf == nil {
		prefSize = d.provisionalDocSize
	}
	return geom.NewSize(50, 50), prefSize, unison.MaxSize(prefSize)
}

// pdfScrollLayout wraps the document ScrollPanel's own layout to change only what the scroll area reports as its
// preferred size: enough to show a single page at the current zoom, rather than its content's preference, which is the
// entire document stacked end to end. A window is packed around the preferred size, and for a many-page document the
// content's preference is tens of thousands of points tall, so such a window is clamped to the display and the title
// bar and toolbar then eat into what remains, cutting the bottom off the page. Preferring one page plus its
// surrounding margins instead makes a packed window come out at exactly a full page plus the toolbar whenever the
// display is tall enough. Only what is asked of the layouts above changes: laying out the scroll panel's children is
// delegated to the ScrollPanel itself, so scrolling still covers the whole document, and the maximum is unbounded, so
// the dock and the user remain free to make the view any size they like. Before the document has been loaded -- and in
// its place if the load fails -- the same US Letter stand-in the document panel reports is used.
type pdfScrollLayout struct {
	d *PDFDockable
}

func (l *pdfScrollLayout) LayoutSizes(_ *unison.Panel, _ geom.Size) (minSize, prefSize, maxSize geom.Size) {
	d := l.d
	if len(d.pageRects) == 0 {
		prefSize = d.provisionalDocSize
	} else {
		prefSize = geom.NewSize(d.docSize.Width, d.pageRects[0].Height+2*pdfPageGap)
	}
	return geom.NewSize(50, 50), prefSize, unison.MaxSize(prefSize)
}

func (l *pdfScrollLayout) PerformLayout(p *unison.Panel) {
	l.d.docScroll.PerformLayout(p)
}

// provisionalPDFDocSize returns the document size to report while the real one is unknown: a single US Letter page,
// sized as the renderer would size one for this display and surrounded by the same margins buildPageLayout puts around
// a real page. A window packed around that has the proportions of a page rather than those of an empty panel, and no
// resize is needed once the document does arrive: the view simply scrolls within the frame the user already has.
func provisionalPDFDocSize(ppi float32, scaleAdjust geom.Point) geom.Size {
	size := pdfUSLetterLogicalSize(ppi, scaleAdjust).Mul(scaleAdj)
	return geom.NewSize(size.Width+2*pdfPageGap, size.Height+2*pdfPageGap)
}

func (d *PDFDockable) draw(gc *unison.Canvas, dirty geom.Rect) {
	gc.DrawRect(dirty, unison.ThemeSurface.Paint(gc, dirty, paintstyle.Fill))
	if len(d.pageRects) == 0 {
		return
	}
	xOff := d.horizontalOffset()
	first, limit := d.pageRange(dirty.Y, dirty.Bottom())
	for i := first; i < limit; i++ {
		rect := d.pageRects[i]
		gc.Save()
		gc.Translate(geom.NewPoint(rect.X+xOff, rect.Y))
		gc.Scale(geom.NewPoint(scaleAdj, scaleAdj))
		r := geom.Rect{Size: d.pdf.PageLogicalSize(i)}
		gc.DrawRect(r, unison.White.Paint(gc, r, paintstyle.Fill))
		if page := d.pdf.CachedPage(i); page != nil && page.Image != nil {
			gc.DrawImageInRect(page.Image, r, &unison.SamplingOptions{
				UseCubic:       true,
				CubicResampler: unison.MitchellResampler(),
				FilterMode:     filtermode.Linear,
				MipMapMode:     mipmapmode.Linear,
			}, nil)
			if len(page.Matches) != 0 {
				p := unison.NewPaint()
				p.SetStyle(paintstyle.Fill)
				p.SetBlendMode(blendmode.Modulate)
				p.SetColor(adjustForModulate(unison.ThemeFocus.GetColor()))
				for _, match := range page.Matches {
					gc.DrawRect(match, p)
				}
			}
		}
		if d.link != nil && d.linkPage == i {
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

func (d *PDFDockable) drawOverlay(gc *unison.Canvas, dirty geom.Rect) {
	if d.loadErr != nil {
		d.drawOverlayMsg(gc, dirty, fmt.Sprintf("%s", d.loadErr), true) //nolint:gocritic // I want the extra processing %s does in this case
		return
	}
	if d.pdf == nil {
		// The document is still being loaded. Hold the message back for the same interval a slow page render does, so
		// that opening a document which comes up promptly doesn't flash an overlay on screen and straight back off
		// again. There is nothing to redraw for on its own, so a redraw is scheduled for the moment the threshold is
		// crossed.
		if waitFor := maxElapsedRenderTimeWithoutOverlay - time.Since(d.loadStarted); waitFor > renderTimeSlop {
			unison.InvokeTaskAfter(d.MarkForRedraw, waitFor)
		} else {
			d.drawOverlayMsg(gc, dirty, fmt.Sprintf(i18n.Text("Loading %s…"), d.Title()), false)
		}
		return
	}
	if len(d.pageRects) == 0 {
		return
	}
	if page := d.pdf.CachedPage(d.currentPage); page != nil && page.Error != nil {
		d.drawOverlayMsg(gc, dirty, fmt.Sprintf("%s", page.Error), true) //nolint:gocritic // I want the extra processing %s does in this case
		return
	}
	// Report on the visible page that has been waiting the longest, since it is the one most likely to have crossed
	// the threshold where an overlay is worthwhile.
	pageNumber := -1
	var requested time.Time
	first, limit := d.visiblePages()
	for i := first; i < limit; i++ {
		if when, pending := d.pdf.PendingSince(i); pending && (pageNumber == -1 || when.Before(requested)) {
			pageNumber = i
			requested = when
		}
	}
	if pageNumber == -1 {
		return
	}
	if waitFor := maxElapsedRenderTimeWithoutOverlay - time.Since(requested); waitFor > renderTimeSlop {
		unison.InvokeTaskAfter(d.MarkForRedraw, waitFor)
	} else {
		label := d.pageLabel(pageNumber)
		if label == "" {
			label = strconv.Itoa(pageNumber + 1)
		}
		d.drawOverlayMsg(gc, dirty, fmt.Sprintf(i18n.Text("Rendering page %s…"), label), false)
	}
}

func (d *PDFDockable) drawOverlayMsg(gc *unison.Canvas, dirty geom.Rect, msg string, forError bool) {
	var fgInk, bgInk unison.Ink
	var icon unison.Drawable
	font := unison.SystemFont.Face().Font(24)
	baseline := font.Baseline()
	if forError {
		fgInk = unison.ThemeOnError
		bgInk = unison.ThemeError.GetColor().SetAlphaIntensity(0.7)
		icon = &unison.DrawableSVG{
			SVG:  unison.CircledExclamationSVG,
			Size: geom.NewSize(baseline, baseline),
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
	var iconSize geom.Size
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
	gc.DrawRoundedRect(r, geom.NewUniformSize(10), bgInk.Paint(gc, dirty, paintstyle.Fill))
	x := r.X + (r.Width-width)/2
	if icon != nil {
		icon.DrawInRect(gc, geom.NewRect(x, r.Y+(r.Height-iconSize.Height)/2, iconSize.Width, iconSize.Height), nil,
			decoration.OnBackgroundInk.Paint(gc, r, paintstyle.Fill))
		x += iconSize.Width + unison.StdHSpacing
	}
	text.Draw(gc, geom.NewPoint(x, r.Y+(r.Height-height)/2+baseline))
}

// TitleIcon implements ux.FileBackedDockable
func (d *PDFDockable) TitleIcon(suggestedSize geom.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  gurps.FileInfoFor(d.path).SVG,
		Size: suggestedSize,
	}
}

// Title implements ux.FileBackedDockable
func (d *PDFDockable) Title() string {
	return xfilepath.BaseName(d.path)
}

// Tooltip implements ux.FileBackedDockable
func (d *PDFDockable) Tooltip() string {
	return d.path
}

// BackingFilePath implements ux.FileBackedDockable
func (d *PDFDockable) BackingFilePath() string {
	return d.path
}

// SetBackingFilePath implements ux.FileBackedDockable
func (d *PDFDockable) SetBackingFilePath(p string) {
	d.path = p
	UpdateTitleForDockable(d)
}

// Modified implements ux.FileBackedDockable
func (d *PDFDockable) Modified() bool {
	return false
}

// MayAttemptClose implements unison.TabCloser
func (d *PDFDockable) MayAttemptClose() bool {
	return true
}

// AttemptClose implements unison.TabCloser
func (d *PDFDockable) AttemptClose() bool {
	if !AttemptCloseForDockable(d) {
		return false
	}
	// This flag needs no synchronization, even though the load runs on a background goroutine: it is only ever written
	// here and only ever read by the load completion task, and both of those run on the UI thread, so they can't
	// interleave. If the load hasn't finished yet, its completion task will see this and dispose of the renderer it
	// produced rather than installing it into a dockable that is already gone.
	d.closed = true
	if d.pdf != nil {
		d.pdf.Close()
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
		d.ScrollToPage(d.tocPanel.SelectedRows(true)[0].pageNumber, true)
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
