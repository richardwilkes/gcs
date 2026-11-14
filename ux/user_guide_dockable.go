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
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/gcs/v5/userguide"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
)

var (
	_ unison.Dockable  = &userGuideDockable{}
	_ unison.TabCloser = &userGuideDockable{}
)

// userGuideDockable shows the User Guide built into GCS, a page at a time beside its table of contents, with a history
// to go back and forward through and a search across every page. Only one is ever open; see showUserGuide.
type userGuideDockable struct {
	unison.Panel
	toolbar       *unison.Panel
	content       *unison.Panel // the side bar, when shown, then docScroll
	side          *sideBar
	toc           *unison.Table[*userGuideTOCNode]
	docScroll     *unison.ScrollPanel
	markdown      *unison.Markdown
	search        *searchDropdown
	sideBarButton *unison.Button
	backButton    *unison.Button
	forwardButton *unison.Button
	pendingScroll func() // the scroll navigate could not make before the view was laid out in a window
	pan           scrollPanDrag
	page          string // the page shown, a path within userguide.FS()
	history       []userGuideLocation
	historyPos    int
	scale         int
	// syncingTOC is set while the table of contents' selection is made to follow navigation or narrowed to one row, so
	// that the change is not taken as a choice of page.
	syncingTOC             bool
	adjustTableSizePending bool
}

// userGuideLocation is a place in the User Guide's history.
type userGuideLocation struct {
	page   string
	anchor string       // the heading ID scrolled to on arrival, "" for the top
	left   markdownSpot // what was at the top of the view when this place was last left; Back and Forward return to it
}

// markdownSpot is a place in a markdown document that stays put when the document is wrapped to another width or shown
// at another scale, unlike a scroll offset: one of the document's top-level blocks, and a fraction of the way down the
// stretch of the document that block begins, which runs to the top of the next block, or to the bottom of the
// document for the last. The first block's stretch begins at the top of the document, so the zero value is that top.
type markdownSpot struct {
	block    int
	fraction float32
}

// markdownSpotAt returns the place y, in the markdown's own coordinates, is at.
func markdownSpotAt(markdown *unison.Markdown, y float32) markdownSpot {
	count := len(markdown.Children())
	for i := range count {
		top, bottom := markdownBlockStretch(markdown, i)
		if y < bottom || i == count-1 {
			if bottom <= top {
				return markdownSpot{block: i}
			}
			return markdownSpot{block: i, fraction: min(max((y-top)/(bottom-top), 0), 1)}
		}
	}
	return markdownSpot{}
}

// y returns where the place is in the markdown's own coordinates.
func (s markdownSpot) y(markdown *unison.Markdown) float32 {
	count := len(markdown.Children())
	if count == 0 {
		return 0
	}
	top, bottom := markdownBlockStretch(markdown, min(s.block, count-1))
	return top + s.fraction*(bottom-top)
}

// markdownBlockStretch returns the top and bottom, in the markdown's own coordinates, of the stretch of the document
// that its top-level block at index i begins; see markdownSpot.
func markdownBlockStretch(markdown *unison.Markdown, i int) (top, bottom float32) {
	blocks := markdown.Children()
	if i > 0 {
		top = blocks[i].FrameRect().Y
	}
	if i < len(blocks)-1 {
		bottom = blocks[i+1].FrameRect().Y
	} else {
		bottom = markdown.ContentRect(true).Height
	}
	return top, bottom
}

// showUserGuide brings the User Guide forward, opening it if needed, at the page and anchor (Home when page is ""),
// with the focus on the page. The anchor is a heading ID, unescaped.
func showUserGuide(page, anchor string) {
	if page == "" {
		page = userguide.HomePage
	}
	var d *userGuideDockable
	if ActivateDockableOfType(func(one *userGuideDockable) bool {
		d = one
		return true
	}) {
		// Activation leaves the focus where it was within the view, or else puts it on the view's first control.
		d.markdown.RequestFocus()
		d.navigate(page, anchor, true)
		return
	}
	d = newUserGuideDockable()
	// Navigating before the view is placed leaves the scroll pending, since the view has not been laid out; it is made
	// once the dock has laid the view out.
	d.navigate(page, anchor, true)
	PlaceInDock(d, dgroup.Markdown, false)
	d.markdown.RequestFocus()
	unison.InvokeTask(d.flushPendingScroll)
}

func newUserGuideDockable() *userGuideDockable {
	d := &userGuideDockable{scale: gurps.GlobalSettings().General.InitialMarkdownUIScale}
	d.Self = d
	d.SetLayout(&unison.FlexLayout{Columns: 1})

	d.markdown = unison.NewMarkdown(false)
	d.markdown.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(20)))
	d.markdown.SetFocusable(true)
	d.markdown.FileSystem = userguide.FS()
	d.markdown.WorkingDirProvider = func(_ unison.Paneler) string {
		if dir := path.Dir(d.page); dir != "." {
			return dir
		}
		return ""
	}
	d.markdown.LinkHandler = d.handleLink
	d.markdown.AnchorLinkHandler = func(anchor string) { d.navigate(d.page, headingID(anchor), true) }

	d.docScroll = unison.NewScrollPanel()
	d.docScroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	d.docScroll.SetContent(d.markdown, behavior.Fill, behavior.Fill)
	wrapMarkdownToView(d.docScroll, d.markdown)
	d.pan.install(d.docScroll, d.markdown.AsPanel(), d.markdown)

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

	d.createTOC()
	d.toolbar = d.createToolbar()
	d.AddChild(d.toolbar)
	d.AddChild(d.content)
	// The scale field only finishes setting itself up once the scroll panel is within the dockable.
	d.content.AddChild(d.docScroll)
	d.side.toggle(d.content, d.markdown)
	return d
}

func (d *userGuideDockable) createToolbar() *unison.Panel {
	toolbar := NewToolbar()
	toolbar.AddChild(NewDefaultInfoPop())

	d.sideBarButton = unison.NewSVGButton(svg.SideBar)
	d.sideBarButton.Tooltip = newWrappedTooltip(i18n.Text("Toggle the Sidebar"))
	d.sideBarButton.ClickCallback = func() { d.side.toggle(d.content, d.markdown) }
	toolbar.AddChild(d.sideBarButton)

	d.backButton = unison.NewSVGButton(svg.Back)
	d.backButton.Tooltip = newWrappedTooltip(i18n.Text("Back"))
	d.backButton.ClickCallback = d.back
	d.backButton.SetEnabled(false)
	toolbar.AddChild(d.backButton)

	d.forwardButton = unison.NewSVGButton(svg.Forward)
	d.forwardButton.Tooltip = newWrappedTooltip(i18n.Text("Forward"))
	d.forwardButton.ClickCallback = d.forward
	d.forwardButton.SetEnabled(false)
	toolbar.AddChild(d.forwardButton)

	toolbar.AddChild(NewScaleField(
		minPDFDockableScale,
		maxPDFDockableScale,
		func() int { return gurps.GlobalSettings().General.InitialMarkdownUIScale },
		func() int { return d.scale },
		func(scale int) { d.scale = scale },
		func() { rewrapMarkdownToView(d.docScroll, d.markdown) },
		false,
		false,
		d.docScroll,
	))

	d.search = newSearchDropdown(d.findMatches, func(match *searchMatch) { match.show() })
	d.search.field.Tooltip = newWrappedTooltipWithSecondaryText(i18n.Text("Search"),
		i18n.Text("Lists the matching pages and sections as you type\nPress RETURN to go to the highlighted one"))
	toolbar.AddChild(d.search.field)

	finishFlowToolbarLayout(toolbar, d.search.field)
	installJumpToField(d.AsPanel(), d.search.field)
	return toolbar
}

func (d *userGuideDockable) createTOC() {
	d.toc = unison.NewTable(&unison.SimpleTableModel[*userGuideTOCNode]{})
	d.toc.Accessibility.Name = i18n.Text("Table of Contents")
	d.toc.Columns = make([]unison.ColumnInfo, 1)
	d.toc.ShowFirstColumnDivider = false
	d.toc.ShowLastColumnDivider = false
	d.toc.PreventUserColumnResize = true
	d.toc.DoubleClickCallback = d.tocDoubleClick
	d.toc.SelectionChangedCallback = d.tocSelectionChanged
	d.side = newSideBar(d.toc)
	entries, err := userguide.TOC()
	if err != nil {
		errs.Log(err)
		return
	}
	d.toc.SetRootRows(newUserGuideTOC(d, nil, entries, ""))
	d.toc.SizeColumnsToFit(true)
	d.side.fitTo(d.toc.Columns[0].Current)
}

// navigate shows the page, scrolled to the anchor (the top when ""), recording the move in the history when
// recordHistory is true; Back and Forward pass false and are returned to the scroll position the place was left at.
func (d *userGuideDockable) navigate(page, anchor string, recordHistory bool) {
	var data []byte
	pageChanged := page != d.page
	if pageChanged {
		var err error
		if data, err = userguide.Content(page); err != nil {
			errs.Log(err, "page", page)
			return
		}
	}
	if recordHistory {
		d.saveScrollPosition()
		d.recordHistory(userGuideLocation{page: page, anchor: anchor})
	}
	if pageChanged {
		// The markdown asks for the page's folder while the content is set, to find its images.
		d.page = page
		setMarkdownContentInView(d.docScroll, d.markdown, data)
		d.markdown.Accessibility.Name = userguide.Title(page)
		UpdateTitleForDockable(d)
	}
	restore := !recordHistory
	var spot markdownSpot
	if restore {
		spot = d.history[d.historyPos].left
	}
	d.pendingScroll = func() { d.scrollTo(anchor, spot, restore) }
	d.applyPendingScroll()
	d.syncTOCSelection()
}

// scrollTo lays the page out afresh and scrolls to bring the spot to the top of the view when restore is set, else the
// anchor, or the top when it is "", then moves the reading caret to what is shown.
func (d *userGuideDockable) scrollTo(anchor string, spot markdownSpot, restore bool) {
	d.docScroll.SetPosition(0, 0)
	d.markdown.MarkForLayoutRecursivelyUpward()
	d.markdown.ValidateScrollRoot()
	switch {
	case restore:
		d.docScroll.SetPosition(0, spot.y(d.markdown)*d.markdown.Scale().Y)
	case anchor != "":
		d.markdown.ScrollToAnchor(anchor)
	}
	placeReadingCaretAtTopOfView(d.docScroll, d.markdown)
}

// laidOut returns true if the view is in a window and has been given room there, so that a scroll can be made.
func (d *userGuideDockable) laidOut() bool {
	return d.Window() != nil && d.docScroll.ContentView().ContentRect(false).Height > 0
}

func (d *userGuideDockable) applyPendingScroll() {
	if scroll := d.pendingScroll; scroll != nil && d.laidOut() {
		d.pendingScroll = nil
		scroll()
	}
}

// flushPendingScroll lays out the window the view has been placed in and makes the scroll that navigate left pending,
// then scrolls the table of contents to the page's row, which could not be done before the table was in a window
// either.
func (d *userGuideDockable) flushPendingScroll() {
	if d.pendingScroll == nil {
		return
	}
	if wnd := d.Window(); wnd != nil {
		wnd.ValidateLayout()
	}
	d.applyPendingScroll()
	d.syncTOCSelection()
}

// saveScrollPosition records what is at the top of the view in the current history entry. A view still waiting on a
// scroll has not been shown at the place, so what was recorded when it was last left stays.
func (d *userGuideDockable) saveScrollPosition() {
	if d.pendingScroll == nil && d.historyPos < len(d.history) {
		_, y := d.docScroll.Position()
		d.history[d.historyPos].left = markdownSpotAt(d.markdown, y/d.markdown.Scale().Y)
	}
}

// recordHistory makes the place the current history entry. A place that is already current adds nothing, nor does one
// that is the next entry, which keeps the entries after it; any other drops the entries after the current one.
func (d *userGuideDockable) recordHistory(loc userGuideLocation) {
	switch {
	case len(d.history) == 0:
		d.history = []userGuideLocation{loc}
		d.historyPos = 0
	case !d.history[d.historyPos].samePlace(loc):
		d.historyPos++
		if d.historyPos < len(d.history) {
			if !d.history[d.historyPos].samePlace(loc) {
				d.history[d.historyPos] = loc
				d.history = d.history[:d.historyPos+1]
			}
		} else {
			d.history = append(d.history, loc)
		}
	}
	d.syncHistoryButtons()
}

func (loc userGuideLocation) samePlace(other userGuideLocation) bool {
	return loc.page == other.page && loc.anchor == other.anchor
}

func (d *userGuideDockable) syncHistoryButtons() {
	d.backButton.SetEnabled(d.historyPos > 0)
	d.forwardButton.SetEnabled(d.historyPos < len(d.history)-1)
}

func (d *userGuideDockable) back() {
	if d.historyPos > 0 {
		d.moveInHistory(-1)
	}
}

func (d *userGuideDockable) forward() {
	if d.historyPos < len(d.history)-1 {
		d.moveInHistory(1)
	}
}

func (d *userGuideDockable) moveInHistory(delta int) {
	d.saveScrollPosition()
	d.historyPos += delta
	d.syncHistoryButtons()
	loc := d.history[d.historyPos]
	d.navigate(loc.page, loc.anchor, false)
}

// syncTOCSelection selects the first row of the table of contents that opens the page shown, alone, disclosing it and
// scrolling it into view, or clears the selection when there is none. A person already on that row keeps the cell
// cursor in the column it is in.
func (d *userGuideDockable) syncTOCSelection() {
	d.syncingTOC = true
	defer func() { d.syncingTOC = false }()
	node := findUserGuideTOCNode(d.toc.RootRows(), d.page)
	if node == nil {
		d.toc.ClearSelection()
		return
	}
	d.toc.DiscloseRow(node, false)
	i := d.toc.RowToIndex(node)
	if i < 0 {
		d.toc.ClearSelection()
		return
	}
	col := -1
	if d.toc.LeadRowIndex() == i {
		col = d.toc.LeadColumnIndex()
	}
	d.toc.ValidateScrollRoot()
	d.toc.SetLeadCell(i, col)
}

func findUserGuideTOCNode(nodes []*userGuideTOCNode, page string) *userGuideTOCNode {
	for _, node := range nodes {
		if node.page == page {
			return node
		}
		if found := findUserGuideTOCNode(node.children, page); found != nil {
			return found
		}
	}
	return nil
}

// tocSelectionChanged shows the page of the row the person moved to in the table of contents (see currentRowIndex). A
// group has no page to show, and moving to the row of the page shown leaves the view where it is; either way, that row
// is left selected alone, with the cell cursor in the column it is in, as navigate leaves the page's row.
func (d *userGuideDockable) tocSelectionChanged() {
	if d.syncingTOC {
		return
	}
	if i := currentRowIndex(d.toc); i >= 0 {
		if row := d.toc.RowFromIndex(i); row.page != "" && row.page != d.page {
			d.navigate(row.page, "", true)
		} else if d.toc.SelectionCount() > 1 {
			d.syncingTOC = true
			defer func() { d.syncingTOC = false }()
			d.toc.SetLeadCell(i, d.toc.LeadColumnIndex())
		}
	}
}

func (d *userGuideDockable) tocDoubleClick() {
	altered := false
	for _, row := range d.toc.SelectedRows(false) {
		if row.CanHaveChildren() {
			altered = true
			row.SetOpen(!row.IsOpen())
		}
	}
	if altered {
		d.toc.SyncToModel()
	}
}

func (d *userGuideDockable) adjustTableSizeEventually() {
	if !d.adjustTableSizePending {
		d.adjustTableSizePending = true
		unison.InvokeTaskAfter(d.adjustTableSize, time.Millisecond)
	}
}

func (d *userGuideDockable) adjustTableSize() {
	d.adjustTableSizePending = false
	d.toc.SyncToModel()
	d.toc.SizeColumnsToFit(true)
}

// handleLink follows a link in the page shown: one to a place in the guide is shown in this view, and anything else is
// handed to OpenPageReference, which sends an http, https or mailto: URL to the system and an md: reference to the User
// Guide or the library file it names, and takes the rest for a page reference. HandleLink is not used, since it would
// look for a relative target in the process's working directory.
func (d *userGuideDockable) handleLink(_ unison.Paneler, target string) {
	if strings.HasPrefix(target, "#") {
		// The markdown hands over an anchor link only when the page has no heading it names.
		return
	}
	if page, anchor, ok := userguide.ResolveLink(d.page, target); ok {
		d.navigate(page, anchor, true)
		return
	}
	OpenPageReference(target, "", nil)
}

// headingID returns the heading ID an anchor in a link names, which may be URL-escaped.
func headingID(anchor string) string {
	if unescaped, err := url.PathUnescape(anchor); err == nil {
		return unescaped
	}
	return anchor
}

// findMatches returns the places in the guide matching the lower-cased text, up to one more than maxSearchMatches of
// them (see searchMatchRows).
func (d *userGuideDockable) findMatches(text string) []*searchMatch {
	size := unison.LabelFont.Size() + 5
	pageIcon := &unison.DrawableSVG{SVG: svg.MarkdownFile, Size: geom.NewSize(size, size)}
	sectionIcon := &unison.DrawableSVG{SVG: svg.Bookmark, Size: geom.NewSize(size, size)}
	matches := userguide.Search(text, maxSearchMatches+1)
	found := make([]*searchMatch, 0, len(matches))
	for _, match := range matches {
		row := &searchMatch{
			show: func() { d.navigate(match.Page, match.Anchor, true) },
			icon: sectionIcon,
			name: match.Text,
			key:  match.Page + "#" + match.Anchor,
		}
		pageTitle := userguide.Title(match.Page)
		switch {
		case match.Kind == userguide.TitleMatch:
			row.icon = pageIcon
		case match.Kind == userguide.BodyMatch && match.Section != "":
			row.title = i18n.Text("%s › %s", pageTitle, match.Section)
		default:
			row.title = pageTitle
		}
		found = append(found, row)
	}
	return found
}

// TitleIcon implements unison.Dockable.
func (d *userGuideDockable) TitleIcon(suggestedSize geom.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  svg.Help,
		Size: suggestedSize,
	}
}

// Title implements unison.Dockable.
func (d *userGuideDockable) Title() string {
	return i18n.Text("User Guide")
}

// Tooltip implements unison.Dockable. It is the title of the page shown.
func (d *userGuideDockable) Tooltip() string {
	if d.page == "" {
		return ""
	}
	return userguide.Title(d.page)
}

// Modified implements unison.Dockable. The guide is only read.
func (d *userGuideDockable) Modified() bool {
	return false
}

// MayAttemptClose implements unison.TabCloser.
func (d *userGuideDockable) MayAttemptClose() bool {
	return true
}

// AttemptClose implements unison.TabCloser.
func (d *userGuideDockable) AttemptClose() bool {
	return AttemptCloseForDockable(d)
}
