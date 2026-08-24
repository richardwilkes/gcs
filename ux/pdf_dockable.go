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
	pdfDragThreshold                   = float32(3)
	pdfAutoScrollInterval              = 30 * time.Millisecond
	pdfAutoScrollMax                   = float32(24)
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
	pendingAnchor          *pdfPendingSelPoint
	pendingExtent          *pdfPendingSelPoint
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
	autoScrollGen          int
	loadStarted            time.Time
	docSize                geom.Size
	provisionalDocSize     geom.Size
	rolloverRect           geom.Rect
	dragStart              geom.Point
	dragOrigin             geom.Point
	pressRoot              geom.Point
	lastDragRoot           geom.Point
	autoScrollDelta        geom.Point
	selAnchor              pdfTextPos
	selExtent              pdfTextPos
	selUnitLo              pdfTextPos
	selUnitHi              pdfTextPos
	autoScaling            autoscale.Option
	lastMods               mod.Modifiers
	dragMode               pdfSelectMode
	inDrag                 bool
	hasSelection           bool
	selAnchorSet           bool
	selExtentSet           bool
	selecting              bool
	maybeSelecting         bool
	copyPending            bool
	autoScrollOn           bool
	noUpdate               bool
	adjustTableSizePending bool
	viewSyncPending        bool
	closed                 bool
}

type pdfTextPos struct {
	page  int
	index int
}

func (p pdfTextPos) before(other pdfTextPos) bool {
	if p.page != other.page {
		return p.page < other.page
	}
	return p.index < other.index
}

type pdfSelectMode uint8

const (
	pdfSelectPoint pdfSelectMode = iota // The character nearest the point
	pdfSelectWord                       // The whole word the point lands in
	pdfSelectLine                       // The whole line the point lands in
)

type pdfPendingSelPoint struct {
	pt   geom.Point
	page int
	mode pdfSelectMode
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
	d.FocusChangeInHierarchyCallback = d.focusChangeInHierarchy
	d.GainedFocusCallback = d.requestRenderPriority
	d.SetLayout(&unison.FlexLayout{Columns: 1})

	d.createTOC()
	d.createContent()

	d.AddChild(d.createToolbar())
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
	}, func(pageNumber int) {
		unison.InvokeTask(func() { d.textExtracted(pageNumber) })
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
	d.notifyBoundsKnown()
	d.LoadPage(d.initialPageInfo)
	if d.HasInSelfOrDescendants(func(p *unison.Panel) bool { return p.Focused() }) {
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
		unison.InvokeTaskAfter(d.notifyBoundsKnown,
			max(maxElapsedRenderTimeWithoutOverlay-time.Since(d.loadStarted), 0))
	}
}

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

func (d *PDFDockable) requestRenderPriority() {
	if d.pdf != nil {
		d.pdf.RequestRenderPriority()
	}
}

func (d *PDFDockable) pageCount() int {
	if d.pdf == nil {
		return 0
	}
	return d.pdf.PageCount()
}

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
	AddKeyBindingInfoToInfoPop(info, unison.KeyBinding{KeyCode: unison.KeyEscape},
		i18n.Text("Clear the text selection"))
	AddHelpToInfoPop(info, fmt.Sprintf(i18n.Text(`
Dragging with the mouse selects text, which
can then be copied with the Copy command.
Holding down the %s key while dragging, or
dragging with the right mouse button, pans
the page instead.`), mod.Option.String()))
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
	d.docPanel.InstallCmdHandlers(unison.CopyItemID,
		func(_ any) bool { return d.hasSelectionRange() },
		func(_ any) { d.copySelection() })

	d.docScroll = unison.NewScrollPanel()
	d.docScroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	d.docScroll.SetContent(d.docPanel, behavior.Fill, behavior.Fill)
	d.docScroll.SetLayout(&pdfScrollLayout{d: d})
	cv := d.docScroll.ContentView()
	cv.DrawOverCallback = d.drawOverlay
	cv.FrameChangeCallback = d.scheduleViewSync

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

func (d *PDFDockable) scheduleViewSync() {
	if !d.viewSyncPending {
		d.viewSyncPending = true
		unison.InvokeTask(d.syncViewState)
	}
}

func (d *PDFDockable) syncViewState() {
	d.viewSyncPending = false
	if len(d.pageRects) == 0 {
		return
	}
	if !d.flushPendingScroll() {
		d.pdf.SetWantedPages([]int{d.pendingScrollPage})
		return
	}

	first, limit := d.visiblePages()
	d.currentPage = first
	if d.history == nil {
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

	// The count comes from the page's text rather than from anything that has been rendered, so it appears as soon as
	// the text of the current page has been extracted -- which searching asks for -- rather than waiting on an image.
	// A dash stands in until then, and for the case of nothing being searched for at all.
	matchText := "-"
	if search := d.searchField.Text(); search != "" {
		if matches, ok := d.pdf.SearchMatches(d.currentPage, search); ok {
			matchText = strconv.Itoa(len(matches))
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
	d.pdf.SetWantedPages(want)
}

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

func (d *PDFDockable) panelScale() float32 {
	return float32(d.scale) / 100
}

func (d *PDFDockable) horizontalOffset() float32 {
	return max((d.docPanel.ContentRect(false).Width-d.docSize.Width)/2, 0)
}

func (d *PDFDockable) visiblePages() (first, limit int) {
	panelScale := d.panelScale()
	top := d.docScroll.Bar(false).Value() / panelScale
	bottom := top + d.docScroll.ContentView().ContentRect(false).Height/panelScale
	// A page whose bottom edge is only just below the top of the viewport isn't meaningfully visible, so bias past it.
	return d.pageRange(top+1, bottom)
}

func (d *PDFDockable) pageRange(top, bottom float32) (first, limit int) {
	count := len(d.pageRects)
	first = min(sort.Search(count, func(i int) bool { return d.pageRects[i].Bottom() > top }), count-1)
	limit = max(sort.Search(count, func(i int) bool { return d.pageRects[i].Y >= bottom }), first+1)
	return first, limit
}

func (d *PDFDockable) pageAt(y float32) int {
	i := sort.Search(len(d.pageRects), func(i int) bool { return d.pageRects[i].Bottom() > y })
	if i >= len(d.pageRects) {
		return -1
	}
	return i
}

func (d *PDFDockable) pageLocalPoint(pageNumber int, where geom.Point) geom.Point {
	return where.Sub(d.pageRects[pageNumber].Point).Sub(geom.NewPoint(d.horizontalOffset(), 0)).Div(scaleAdj)
}

func (d *PDFDockable) pageLocation(where geom.Point) (pageNumber int, local geom.Point, inside bool) {
	pageNumber = d.pageAt(where.Y)
	if pageNumber < 0 {
		return -1, geom.Point{}, false
	}
	local = d.pageLocalPoint(pageNumber, where)
	size := d.pdf.PageLogicalSize(pageNumber)
	inside = local.X >= 0 && local.Y >= 0 && local.X < size.Width && local.Y < size.Height
	return pageNumber, local, inside
}

func (d *PDFDockable) pageLocationClamped(where geom.Point) (pageNumber int, local geom.Point, ok bool) {
	if len(d.pageRects) == 0 {
		return -1, geom.Point{}, false
	}
	pageNumber, local, inside := d.pageLocation(where)
	if inside {
		return pageNumber, local, true
	}
	if pageNumber < 0 {
		pageNumber = len(d.pageRects) - 1
		local = d.pageLocalPoint(pageNumber, where)
	}
	size := d.pdf.PageLogicalSize(pageNumber)
	local.X = min(max(local.X, 0), size.Width)
	local.Y = min(max(local.Y, 0), size.Height)
	return pageNumber, local, true
}

func (d *PDFDockable) overLink(where geom.Point) (rect geom.Rect, link *PDFLink, pageNumber int) {
	pageNumber, pt, _ := d.pageLocation(where)
	if pageNumber < 0 {
		return rect, nil, -1
	}
	page := d.pdf.CachedPage(pageNumber)
	if page == nil {
		return rect, nil, -1
	}
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
	if d.lastMods.OptionDown() {
		return unison.OpenHandCursor()
	}
	if _, link, _ := d.overLink(pt); link != nil {
		return unison.PointingCursor()
	}
	if _, _, inside := d.pageLocation(pt); inside {
		return unison.TextCursor()
	}
	return unison.ArrowCursor()
}

func (d *PDFDockable) mouseDown(where geom.Point, button, clickCount int, mods mod.Modifiers) bool {
	d.docPanel.RequestFocus()
	d.dragStart = d.docPanel.PointToRoot(where)
	d.dragOrigin.X, d.dragOrigin.Y = d.docScroll.Position()
	d.lastMods = mods
	d.inDrag = false
	d.selecting = false
	d.maybeSelecting = false
	if button != unison.ButtonLeft || mods.OptionDown() {
		d.inDrag = true
		d.UpdateCursorNow()
		return true
	}
	overLink := d.checkForLinkAt(where)
	d.clearSelection()
	if !overLink {
		d.pressRoot = d.dragStart
		d.maybeSelecting = true
		switch {
		case clickCount == 2:
			d.selecting = true
			d.dragMode = pdfSelectWord
		case clickCount >= 3:
			d.selecting = true
			d.dragMode = pdfSelectLine
		default:
			d.dragMode = pdfSelectPoint
		}
		d.setSelectionPoint(where, true, d.dragMode)
	}
	d.UpdateCursorNow()
	return true
}

func (d *PDFDockable) mouseDrag(where geom.Point, _ int, _ mod.Modifiers) bool {
	if d.inDrag {
		pt := d.dragStart.Sub(d.docPanel.PointToRoot(where)).Add(d.dragOrigin)
		d.docScroll.SetPosition(pt.X, pt.Y)
		return true
	}
	if !d.maybeSelecting {
		d.checkForLinkAt(where)
		return true
	}
	root := d.docPanel.PointToRoot(where)
	if !d.selecting {
		delta := root.Sub(d.pressRoot)
		if max(xmath.Abs(delta.X), xmath.Abs(delta.Y)) <= pdfDragThreshold {
			return true
		}
		d.selecting = true
	}
	d.lastDragRoot = root
	d.setSelectionPoint(where, false, d.dragMode)
	d.updateAutoScroll(where)
	d.MarkForRedraw()
	return true
}

func (d *PDFDockable) mouseMove(where geom.Point, mods mod.Modifiers) bool {
	optionChanged := mods.OptionDown() != d.lastMods.OptionDown()
	d.lastMods = mods
	if optionChanged {
		d.UpdateCursorNow()
	}
	d.checkForLinkAt(where)
	return true
}

func (d *PDFDockable) mouseUp(where geom.Point, button int, _ mod.Modifiers) bool {
	switch {
	case d.inDrag:
		d.inDrag = false
		d.UpdateCursorNow()
	case d.selecting:
		d.selecting = false
		d.maybeSelecting = false
		d.stopAutoScroll()
		d.UpdateCursorNow()
	default:
		d.maybeSelecting = false
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

func (d *PDFDockable) clearSelection() {
	if !d.hasSelection && !d.selAnchorSet && !d.selExtentSet && d.pendingAnchor == nil && d.pendingExtent == nil &&
		!d.copyPending {
		return
	}
	d.hasSelection = false
	d.selAnchorSet = false
	d.selExtentSet = false
	d.selAnchor = pdfTextPos{}
	d.selExtent = pdfTextPos{}
	d.selUnitLo = pdfTextPos{}
	d.selUnitHi = pdfTextPos{}
	d.pendingAnchor = nil
	d.pendingExtent = nil
	d.copyPending = false
	d.MarkForRedraw()
}

func (d *PDFDockable) hasSelectionRange() bool {
	return d.hasSelection && d.selAnchor != d.selExtent
}

func (d *PDFDockable) orderedSelection() (lo, hi pdfTextPos) {
	if d.selExtent.before(d.selAnchor) {
		return d.selExtent, d.selAnchor
	}
	return d.selAnchor, d.selExtent
}

func (d *PDFDockable) setSelectionPoint(where geom.Point, anchor bool, mode pdfSelectMode) {
	pageNumber, local, ok := d.pageLocationClamped(where)
	if !ok {
		return
	}
	d.resolveSelectionPoint(pageNumber, local, anchor, mode)
}

func (d *PDFDockable) resolveSelectionPoint(pageNumber int, local geom.Point, anchor bool, mode pdfSelectMode) {
	index, ok := d.pdf.TextIndexAt(pageNumber, local)
	if !ok {
		pending := &pdfPendingSelPoint{pt: local, page: pageNumber, mode: mode}
		if anchor {
			d.pendingAnchor = pending
		} else {
			d.pendingExtent = pending
		}
		return
	}
	unitLo := pdfTextPos{page: pageNumber, index: index}
	unitHi := unitLo
	switch mode {
	case pdfSelectWord:
		unitLo.index, unitHi.index, _ = d.pdf.TextWordAt(pageNumber, index)
	case pdfSelectLine:
		unitLo.index, unitHi.index, _ = d.pdf.TextLineAt(pageNumber, index)
	}
	if anchor {
		d.selUnitLo = unitLo
		d.selUnitHi = unitHi
		d.selAnchorSet = true
		d.pendingAnchor = nil
		switch {
		case d.selExtentSet:
			d.selAnchor, d.selExtent = pdfTextPosSpan(unitLo, unitHi, d.selExtent, d.selExtent)
		case d.pendingExtent == nil:
			d.selAnchor = unitLo
			d.selExtent = unitHi
			d.selExtentSet = true
		default:
			d.selAnchor = unitLo
		}
	} else {
		d.pendingExtent = nil
		if d.selAnchorSet {
			d.selAnchor, d.selExtent = pdfTextPosSpan(d.selUnitLo, d.selUnitHi, unitLo, unitHi)
		} else {
			d.selExtent = unitLo
		}
		d.selExtentSet = true
	}
	d.hasSelection = d.selAnchorSet && d.selExtentSet
	d.MarkForRedraw()
}

func pdfTextPosSpan(aLo, aHi, bLo, bHi pdfTextPos) (lo, hi pdfTextPos) {
	lo, hi = aLo, aHi
	if bLo.before(lo) {
		lo = bLo
	}
	if hi.before(bHi) {
		hi = bHi
	}
	return lo, hi
}

func (d *PDFDockable) textExtracted(pageNumber int) {
	if d.pdf == nil || d.closed {
		return
	}
	if pending := d.pendingAnchor; pending != nil && pending.page == pageNumber {
		d.pendingAnchor = nil
		d.resolveSelectionPoint(pending.page, pending.pt, true, pending.mode)
	}
	if pending := d.pendingExtent; pending != nil && pending.page == pageNumber {
		d.pendingExtent = nil
		d.resolveSelectionPoint(pending.page, pending.pt, false, pending.mode)
	}
	if d.copyPending {
		d.copySelection()
	}
	d.MarkForRedraw()
	// The matches label is computed from the current page's text, so it has an answer to show now that this page's
	// text has landed.
	d.scheduleViewSync()
}

func pdfSelectionRangeFor(lo, hi pdfTextPos, pageNumber int, lengthOf func(int) (int, bool)) (start, end int, ok bool) {
	if pageNumber < lo.page || pageNumber > hi.page {
		return 0, 0, true
	}
	if pageNumber == lo.page {
		start = lo.index
	}
	if pageNumber == hi.page {
		return start, hi.index, true
	}
	length, known := lengthOf(pageNumber)
	if !known {
		return 0, 0, false
	}
	return start, length, true
}

func (d *PDFDockable) selectionRange(pageNumber int) (start, end int, ok bool) {
	if !d.hasSelection {
		return 0, 0, true
	}
	lo, hi := d.orderedSelection()
	return pdfSelectionRangeFor(lo, hi, pageNumber, d.pdf.TextLength)
}

func (d *PDFDockable) selectionText() (text string, ok bool) {
	lo, hi := d.orderedSelection()
	parts := make([]string, 0, (hi.page-lo.page)+1)
	for pageNumber := lo.page; pageNumber <= hi.page; pageNumber++ {
		start, end, known := d.selectionRange(pageNumber)
		if !known {
			return "", false
		}
		if start >= end {
			continue
		}
		if part := d.pdf.TextRange(pageNumber, start, end); part != "" {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, "\n"), true
}

func (d *PDFDockable) copySelection() {
	if !d.hasSelectionRange() {
		d.copyPending = false
		return
	}
	text, ok := d.selectionText()
	if !ok {
		d.copyPending = true
		return
	}
	d.copyPending = false
	unison.ClipboardSetText(text)
}

func (d *PDFDockable) updateAutoScroll(where geom.Point) {
	view := d.docScroll.ContentView()
	viewRect := view.ContentRect(false)
	pt := view.PointFromRoot(d.docPanel.PointToRoot(where))
	d.autoScrollDelta = geom.NewPoint(pdfAutoScrollAmount(pt.X, viewRect.X, viewRect.Right()),
		pdfAutoScrollAmount(pt.Y, viewRect.Y, viewRect.Bottom()))
	if d.autoScrollDelta == (geom.Point{}) || d.autoScrollOn {
		return
	}
	d.autoScrollOn = true
	d.scheduleAutoScrollTick()
}

func (d *PDFDockable) scheduleAutoScrollTick() {
	generation := d.autoScrollGen
	unison.InvokeTaskAfter(func() { d.autoScrollTick(generation) }, pdfAutoScrollInterval)
}

func (d *PDFDockable) autoScrollTick(generation int) {
	if !d.autoScrollOn || generation != d.autoScrollGen {
		return
	}
	if !d.selecting || d.closed || d.pdf == nil || d.autoScrollDelta == (geom.Point{}) {
		d.stopAutoScroll()
		return
	}
	h, v := d.docScroll.Position()
	d.docScroll.SetPosition(h+d.autoScrollDelta.X, v+d.autoScrollDelta.Y)
	where := d.docPanel.PointFromRoot(d.lastDragRoot)
	d.setSelectionPoint(where, false, d.dragMode)
	d.updateAutoScroll(where)
	d.MarkForRedraw()
	d.scheduleAutoScrollTick()
}

func (d *PDFDockable) stopAutoScroll() {
	d.autoScrollOn = false
	d.autoScrollGen++
	d.autoScrollDelta = geom.Point{}
}

func pdfAutoScrollAmount(pos, lo, hi float32) float32 {
	if pos < lo {
		return -min(lo-pos, pdfAutoScrollMax)
	}
	if pos > hi {
		return min(pos-hi, pdfAutoScrollMax)
	}
	return 0
}

func (d *PDFDockable) focusChangeInHierarchy(_, _ *unison.Panel) {
	d.requestRenderPriority()
}

func (d *PDFDockable) keyDown(keyCode unison.KeyCode, _ mod.Modifiers, _ bool) bool {
	if len(d.pageRects) == 0 {
		return false
	}
	if keyCode == unison.KeyEscape {
		if d.hasSelectionRange() {
			d.clearSelection()
			return true
		}
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

func (d *PDFDockable) docSizer(_ geom.Size) (minSize, prefSize, maxSize geom.Size) {
	prefSize = d.docSize
	if d.pdf == nil {
		prefSize = d.provisionalDocSize
	}
	return geom.NewSize(50, 50), prefSize, unison.MaxSize(prefSize)
}

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
	// The search text is read once here rather than once per page: it can't change in the middle of a draw, and the
	// pages below ask the renderer for their hits with it.
	search := d.searchField.Text()
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
			// The hits are marked only on top of a page that has actually been rendered. Bars floating on the blank
			// white slot a page occupies until its image arrives would read as marks on an empty page, and nobody can
			// see what they are supposed to be marking there anyway.
			if search != "" {
				if matches, _ := d.pdf.SearchMatches(i, search); len(matches) != 0 {
					p := pdfHighlightPaint(unison.ThemeWarning.GetColor())
					for _, match := range matches {
						gc.DrawRect(pdfSearchHitBar(match), p)
					}
				}
			}
		}
		if d.hasSelection {
			if start, end, ok := d.selectionRange(i); ok && start < end {
				if rects := d.pdf.TextHighlights(i, start, end); len(rects) != 0 {
					p := pdfHighlightPaint(adjustForModulate(unison.ThemeFocus.GetColor()))
					for _, sel := range rects {
						gc.DrawRect(sel, p)
					}
				}
			}
		}
		if d.link != nil && d.linkPage == i {
			gc.DrawRect(d.rolloverRect, pdfHighlightPaint(adjustForModulate(unison.ThemeFocus.GetColor())))
		}
		gc.Restore()
	}
}

const pdfSearchHitBarFraction = float32(0.25)

func pdfHighlightPaint(color unison.Color) *unison.Paint {
	p := unison.NewPaint()
	p.SetStyle(paintstyle.Fill)
	p.SetBlendMode(blendmode.Modulate)
	p.SetColor(color)
	return p
}

func pdfSearchHitBar(hit geom.Rect) geom.Rect {
	height := hit.Height * pdfSearchHitBarFraction
	return geom.NewRect(hit.X, hit.Bottom()-height, hit.Width, height)
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
		//nolint:gocritic // I want the extra processing %s does in this case
		d.drawOverlayMsg(gc, dirty, fmt.Sprintf("%s", d.loadErr), true)
		return
	}
	if d.pdf == nil {
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
		//nolint:gocritic // I want the extra processing %s does in this case
		d.drawOverlayMsg(gc, dirty, fmt.Sprintf("%s", page.Error), true)
		return
	}
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
	// The selection points into text the renderer is about to release, and the auto-scroll loop would go on driving it
	// for another tick or two, so both are let go of before the document is.
	d.clearSelection()
	d.stopAutoScroll()
	d.selecting = false
	d.maybeSelecting = false
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
