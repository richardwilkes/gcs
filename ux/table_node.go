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
	"slices"
	"strings"
	"unicode"

	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
)

const (
	noInvertColorsMarker = "no_invert"
	markdownCellInksKey  = "markdown_cell_inks"
)

// markdownCellInks holds the colors a markdown cell draws with. The markdown is handed these indirect colors when it is
// created and they are re-pointed whenever its row is selected or deselected, so that the new colors take effect on the
// next draw without rebuilding the markdown. A rebuild replaces every panel within it, which gives each a new
// accessibility node and so tells assistive technologies that the cell changed, pulling a screen reader back to the
// row being left each time the selection moves.
type markdownCellInks struct {
	foreground    unison.Ink
	background    unison.Ink
	onBackground  unison.IndirectColor
	link          unison.IndirectColor
	linkOnPressed unison.IndirectColor
	selected      bool
	initialized   bool
}

var _ unison.TableRowData[*Node[*gurps.Trait]] = &Node[*gurps.Trait]{}

// CellCache holds a panel built for a table row's cell, along with what it was built from, so that the panel can be
// used again rather than built anew each time the cell is drawn or measured.
type CellCache struct {
	Panel unison.Paneler
	// Data is the data the panel was built from. The panel may change it: a check cell flips Checked as it is
	// clicked, and back should the click be refused, so that the panel goes on matching the model and is kept.
	Data        *gurps.CellData
	Width       float32
	NotesClosed bool
	PanelStamp  uint64 // The stamp (see cellPanelEpoch) the panel was built at.
}

// Matches returns true if the panel was built for the given width, data and notes disclosure, and with the fonts in
// use now (see cellPanelEpoch). A width of zero or less asks for the panel's natural size, which no width constrains,
// so any such width matches a panel built for any other. A nil cache never matches.
func (c *CellCache) Matches(width float32, data *gurps.CellData, notesClosed bool) bool {
	return c != nil && c.Panel != nil && c.NotesClosed == notesClosed && c.PanelStamp == currentCellPanelStamp() &&
		(c.Width == width || (c.Width <= 0 && width <= 0)) && *c.Data == *data
}

// cellStamp marks what a row derived from the model for one of its cells. The row derives it afresh once a change has
// been reported to the display (see displayEpoch) and, if its provider keeps its rows across syncs (see rowReuse),
// once the table has synced again, which is when a row built anew would have derived it.
type cellStamp struct {
	display uint64
	sync    uint64
}

// nodeCell is what a row keeps for one of its columns: the cell's data, derived from the model once per change
// reported to the display and per sync (see cellStamp), the panels built from it for the widths the table asks for,
// and the widths the page reference columns are sized by. Each derived value is stamped with the cellStamp it was
// derived at; the panels are kept for as long as they match the data (see CellCache.Matches).
type nodeCell struct {
	display       gurps.CellData // The data for display, with the page's display preferences applied.
	plain         gurps.CellData // The data for sorting and searching, without them; the same as display off a page.
	displayAt     cellStamp
	plainAt       cellStamp
	unconstrained *CellCache    // The panel built at its natural size, which the sizing passes ask for.
	fitted        [4]*CellCache // The panels built for the most recent widths, newest first; see ColumnCell.
	refWidth      float32
	collapsed     float32
	refWidthAt    cellStamp
	collapsedAt   cellStamp
}

// Node represents a row in a table.
type Node[T gurps.Node[T]] struct {
	table       *unison.Table[*Node[T]]
	parent      *Node[T]
	data        T
	children    []*Node[T]
	cells       []nodeCell // One per column, allocated the first time a cell is wanted.
	reuse       *rowReuse  // Set when the provider keeps its rows across syncs.
	childrenGen uint64     // The reuse generation the children were brought up to date in.
	forPage     bool
}

// NewNode creates a new node for a table.
func NewNode[T gurps.Node[T]](table *unison.Table[*Node[T]], parent *Node[T], data T, forPage bool) *Node[T] {
	return &Node[T]{
		table:   table,
		parent:  parent,
		data:    data,
		forPage: forPage,
	}
}

// NewNodeLike creates a new node for a table based on the characteristics of an existing node in that table.
func NewNodeLike[T gurps.Node[T]](like *Node[T], data T) *Node[T] {
	return &Node[T]{
		table:   like.table,
		parent:  like.parent,
		data:    data,
		reuse:   like.reuse,
		forPage: like.forPage,
	}
}

// nodesFor returns a node under the parent for each of the data items, in their order: the one among the previous
// nodes that holds the same item, when there is one, or else a new one sharing the given reuse, which may be nil. Only
// nodes that were under the same parent may be passed as the previous ones, so that a row moved elsewhere gets a new
// node there, with the right parent.
func nodesFor[T gurps.Node[T]](table *unison.Table[*Node[T]], parent *Node[T], data []T, previous []*Node[T], forPage bool, reuse *rowReuse) []*Node[T] {
	nodes := make([]*Node[T], len(data))
	var byData map[T]*Node[T]
	for i, one := range data {
		if byData == nil {
			// The items are usually the ones there were before, in the same order, so they are matched up by position
			// until that fails.
			if i < len(previous) && previous[i].data == one {
				nodes[i] = previous[i]
				continue
			}
			byData = make(map[T]*Node[T], max(len(previous)-i, 0))
			for _, node := range previous[min(i, len(previous)):] {
				byData[node.data] = node
			}
		}
		if node, ok := byData[one]; ok {
			// Taken out, so that a list holding the same item twice doesn't show the same node twice.
			delete(byData, one)
			nodes[i] = node
		} else {
			nodes[i] = &Node[T]{
				table:   table,
				parent:  parent,
				data:    one,
				reuse:   reuse,
				forPage: forPage,
			}
		}
	}
	return nodes
}

// cell returns what the row keeps for the given column.
func (n *Node[T]) cell(col int) *nodeCell {
	if col >= len(n.cells) {
		cells := make([]nodeCell, max(len(n.table.Columns), col+1))
		copy(cells, n.cells)
		n.cells = cells
	}
	return &n.cells[col]
}

// CloneForTarget implements unison.TableRowData.
func (n *Node[T]) CloneForTarget(target unison.Paneler, newParent *Node[T]) *Node[T] {
	table, ok := target.(*unison.Table[*Node[T]])
	if !ok {
		xos.ExitIfErr(errs.New("unable to convert to table"))
	}
	var owner gurps.DataOwner
	if provider := DetermineDataOwnerProvider(target); provider != nil {
		owner = provider.DataOwner()
	}
	return NewNode(table, newParent,
		n.data.Clone(libraryFileFromTable(n.table), owner, newParent.Data(), gurps.Reference), false)
}

// ID implements unison.TableRowData.
func (n *Node[T]) ID() tid.TID {
	return n.data.ID()
}

// Parent implements unison.TableRowData.
func (n *Node[T]) Parent() *Node[T] {
	return n.parent
}

// SetParent implements unison.TableRowData.
func (n *Node[T]) SetParent(parent *Node[T]) {
	n.data.SetParent(parent.Data())
}

// CanHaveChildren implements unison.TableRowData.
func (n *Node[T]) CanHaveChildren() bool {
	return n.data.Container()
}

// Children implements unison.TableRowData. A row whose provider keeps its rows across syncs (see rowReuse) brings its
// children up to date with the model once per sync, keeping the node of each child that is still one of its own.
func (n *Node[T]) Children() []*Node[T] {
	switch {
	case n.reuse == nil:
		if n.data.Container() && n.children == nil {
			n.children = nodesFor(n.table, n, n.data.NodeChildren(), nil, n.forPage, nil)
		}
	case n.childrenGen != n.reuse.gen || (n.children == nil && n.data.Container()):
		if n.data.Container() {
			n.children = nodesFor(n.table, n, n.data.NodeChildren(), n.children, n.forPage, n.reuse)
		} else {
			n.children = nil
		}
		n.childrenGen = n.reuse.gen
	}
	return n.children
}

// SetChildren implements unison.TableRowData.
func (n *Node[T]) SetChildren(children []*Node[T]) {
	if n.data.Container() {
		n.data.SetChildren(ExtractNodeDataFromList(children))
		n.children = nil
	}
}

// RefreshChildren discards the cached child nodes so that the next call to Children rebuilds them from the underlying
// data, without keeping any of them. Call it after altering the data's children directly.
func (n *Node[T]) RefreshChildren() {
	n.children = nil
}

// currentCellStamp returns the stamp that marks what the row derives for its cells as current; see cellStamp.
func (n *Node[T]) currentCellStamp() cellStamp {
	stamp := cellStamp{display: currentDisplayStamp()}
	if n.reuse != nil {
		stamp.sync = n.reuse.gen
	}
	return stamp
}

// cellData returns the cell data for the given column of this row. forDisplay says whether the data is for something
// the user will see, which decides whether the page-only display preferences (such as the decimal places shown for
// equipment weights) are applied; they are, but only for a row on a page. Sorting passes false, since their rounding is
// lossy and would tie rows whose real values differ. Searching asks both ways (see columnContains), so that a row is
// found by the value the user sees as well as by the exact value. The data is derived from the model once per change
// reported to the display and per sync (see cellStamp) rather than on every call, since deriving it can be costly and
// the table asks for it several times per sync and once per draw.
func (n *Node[T]) cellData(col int, forDisplay bool) gurps.CellData {
	c := n.cell(col)
	forPage := n.forPage && forDisplay
	data, at := &c.plain, &c.plainAt
	if forPage {
		data, at = &c.display, &c.displayAt
	}
	if stamp := n.currentCellStamp(); *at != stamp {
		*data = gurps.CellData{ForPage: forPage}
		n.data.CellData(n.table.Columns[col].ID, data)
		*at = stamp
	}
	result := *data
	// A row a filter shows only as context for a match beneath it is dimmed like a disabled row, so the matches stand
	// out. Sorting and matching don't see the dimming, and it isn't kept, since the filter can change without a
	// change being reported.
	if forDisplay && n.table.IsFilterContextRow(n) {
		result.Dim = true
	}
	return result
}

// CellDataForSort implements unison.TableRowData.
func (n *Node[T]) CellDataForSort(index int) string {
	data := n.cellData(index, false)
	s := data.ForSort()
	if gurps.GlobalSettings().General.GroupContainersOnSort && n.data.Container() {
		return containerMarker + s
	}
	return s
}

// CellDataForAccessibility implements unison.TableRowAccessibleText: the cell's text without the marker CellDataForSort
// adds to group containers ahead of the other rows, which only the sort has a use for.
func (n *Node[T]) CellDataForAccessibility(col int) string {
	data := n.cellData(col, true)
	return data.ForSort()
}

// ColumnCell implements unison.TableRowData. A panel built earlier for the same data and width is used again with
// its inks brought up to date; the sizing passes, which ask for every cell at its natural size and then at the width
// settled on, and the draws that follow therefore build each cell once per width rather than once per call, and a row
// kept across syncs (see rowReuse) builds only the cells whose data changed. Four widths are kept, since a sync of a
// page list asks for the cells of the column that takes the excess width at up to four: the table is sized to its
// preferred width, rounded up to whole pixels, before the layout gives it back its own, and the columns are sized
// for each, before and after the page reference columns take their share of the excess (see sizePageTableColumns).
func (n *Node[T]) ColumnCell(row, col int, foreground, background unison.Ink, selected, indirectlySelected, _ bool) unison.Paneler {
	data := n.cellData(col, true)
	width := n.table.CellWidth(row, col)
	// The notes disclosure button is part of the panel, so a panel built while the notes were closed can't serve
	// once they are open, whether or not the data changed.
	notesClosed := data.Type == cell.Text && data.Secondary != "" && gurps.IsClosed(n.noteKey())
	selected = selected || indirectlySelected
	c := n.cell(col)
	if width <= 0 {
		if c.unconstrained.Matches(width, &data, notesClosed) {
			applyInkRecursively(c.unconstrained.Panel.AsPanel(), foreground, background, selected)
			return c.unconstrained.Panel
		}
		c.unconstrained = n.buildCell(&data, width, notesClosed, foreground, background, selected)
		return c.unconstrained.Panel
	}
	for _, cached := range c.fitted {
		if cached.Matches(width, &data, notesClosed) {
			applyInkRecursively(cached.Panel.AsPanel(), foreground, background, selected)
			return cached.Panel
		}
	}
	built := n.buildCell(&data, width, notesClosed, foreground, background, selected)
	copy(c.fitted[1:], c.fitted[:len(c.fitted)-1])
	c.fitted[0] = built
	return built.Panel
}

// buildCell builds the panel for the cell from a copy of the data, which the panel keeps (see CellCache.Data).
func (n *Node[T]) buildCell(data *gurps.CellData, width float32, notesClosed bool, foreground, background unison.Ink, selected bool) *CellCache {
	kept := new(gurps.CellData)
	*kept = *data
	return &CellCache{
		Panel:       n.CellFromCellData(kept, width, foreground, background, selected),
		Data:        kept,
		Width:       width,
		NotesClosed: notesClosed,
		PanelStamp:  currentCellPanelStamp(),
	}
}

func applyInkRecursively(panel *unison.Panel, foreground, background unison.Ink, selected bool) {
	switch part := panel.Self.(type) {
	case *unison.Markdown:
		adjustMarkdownInk(part, foreground, background, selected)
		return
	case *unison.Label:
		if part.OnBackgroundInk != foreground {
			if _, exists := part.ClientData()[noInvertColorsMarker]; !exists {
				part.OnBackgroundInk = foreground
				part.SetTitle(part.String())
			}
		}
	case *unison.Tag:
		if part.OnBackgroundInk != background || part.BackgroundInk != foreground {
			if _, exists := part.ClientData()[noInvertColorsMarker]; !exists {
				part.BackgroundInk = foreground
				part.OnBackgroundInk = background
				part.SetTitle(part.Text.String())
			}
		}
	case *unison.Button:
		if part.OnBackgroundInk != foreground {
			if _, exists := part.ClientData()[noInvertColorsMarker]; !exists {
				part.OnBackgroundInk = foreground
				if part.HideBase {
					part.SelectionInk = foreground
				}
			}
		}
	}
	for _, child := range panel.Children() {
		applyInkRecursively(child, foreground, background, selected)
	}
}

// IsOpen implements unison.TableRowData.
func (n *Node[T]) IsOpen() bool {
	return n.data.IsOpen()
}

// SetOpen implements unison.TableRowData. It only records the state: every caller, in unison and here, syncs the table
// itself once it has set all the rows it means to, so syncing here as well would rebuild the table once per row.
func (n *Node[T]) SetOpen(open bool) {
	n.data.SetOpen(open)
}

// Data returns the underlying data object. A nil node returns the zero value.
func (n *Node[T]) Data() T {
	if n == nil {
		var zero T
		return zero
	}
	return n.data
}

// Match returns true if the text is present in the node. The text must already have been lowercased by the caller.
func (n *Node[T]) Match(text string) bool {
	if text != "" {
		for i := range n.table.Columns {
			if n.columnContains(i, text) {
				return true
			}
		}
	}
	return false
}

// columnContains returns true if the lowercased text is present in the given column's text. On a page, both the exact
// text and the display rendering are checked, so that a user can find a row by the rounded value they can see on the
// page as well as by the exact value behind it, which is what the cell's tooltip shows. Off a page the two are the
// same, so the display rendering is not computed a second time.
func (n *Node[T]) columnContains(col int, text string) bool {
	data := n.cellData(col, false)
	if strings.Contains(strings.ToLower(data.ForSort()), text) {
		return true
	}
	if !n.forPage {
		return false
	}
	data = n.cellData(col, true)
	return strings.Contains(strings.ToLower(data.ForSort()), text)
}

// CellFromCellData creates a new panel for the given cell data.
func (n *Node[T]) CellFromCellData(c *gurps.CellData, width float32, foreground, background unison.Ink, selected bool) unison.Paneler {
	switch c.Type {
	case cell.Text, cell.Tags:
		return n.createLabelCell(c, width, foreground, background, selected)
	case cell.Toggle:
		return n.createToggleCell(c, foreground)
	case cell.Switch:
		return n.createSwitchCell(c, foreground)
	case cell.PageRef:
		return n.createPageRefCell(c, width, foreground)
	case cell.Markdown:
		return n.createMarkdownCell(c.Primary, width, fonts.PageFieldPrimary, foreground, background, selected)
	default:
		return unison.NewPanel()
	}
}

func (n *Node[T]) createMarkdownCell(content string, width float32, font unison.Font, foreground, background unison.Ink, selected bool) unison.Paneler {
	m := unison.NewMarkdown(false)
	m.StripBottomEmptyMargin = true
	if wd, ok := n.table.ClientData()[WorkingDirKey]; ok {
		m.ClientData()[WorkingDirKey] = wd
	}
	if n.forPage {
		adjustMarkdownThemeForPage(m, font)
	}
	inks := &markdownCellInks{}
	m.ClientData()[markdownCellInksKey] = inks
	m.OnBackgroundInk = &inks.onBackground
	m.LinkInk = &inks.link
	m.LinkOnPressedInk = &inks.linkOnPressed
	adjustMarkdownInk(m, foreground, background, selected)
	// Wrapped to a whole-pixel width for the same reason addLabelCell wraps to one: each line becomes a label, which
	// reports a width rounded up to whole pixels.
	m.SetContent(markdownHardLineBreaks(content), xmath.Floor(width))
	return m
}

// adjustMarkdownInk re-points the colors of a markdown cell made by createMarkdownCell to suit its row's selection
// state. See markdownCellInks.
func adjustMarkdownInk(md *unison.Markdown, foreground, background unison.Ink, selected bool) {
	inks, ok := md.ClientData()[markdownCellInksKey].(*markdownCellInks)
	if !ok || (inks.initialized && inks.selected == selected && inks.foreground == foreground &&
		inks.background == background) {
		return
	}
	inks.initialized = true
	inks.selected = selected
	inks.foreground = foreground
	inks.background = background
	if selected {
		var link unison.ColorProvider
		if c, isThemeColor := foreground.(*unison.ThemeColor); isThemeColor {
			link = &unison.ThemeColor{
				Light: c.Light.AdjustHue(180),
				Dark:  c.Dark.AdjustHue(180),
			}
		} else {
			link = unison.Yellow
		}
		inks.onBackground.Target = asColorProvider(foreground, unison.ThemeOnFocus)
		inks.link.Target = link
		inks.linkOnPressed.Target = asColorProvider(background, unison.ThemeFocus)
	} else {
		inks.onBackground.Target = asColorProvider(unison.DefaultMarkdownTheme.OnBackgroundInk, unison.ThemeOnSurface)
		inks.link.Target = asColorProvider(unison.DefaultMarkdownTheme.LinkInk, unison.ThemeFocus)
		inks.linkOnPressed.Target = asColorProvider(unison.DefaultMarkdownTheme.LinkOnPressedInk, unison.ThemeOnFocus)
	}
}

// asColorProvider returns the ink as a color provider, or the fallback if the ink is something else, such as a
// gradient, which an indirect color cannot refer to.
func asColorProvider(ink unison.Ink, fallback unison.ColorProvider) unison.ColorProvider {
	if c, ok := ink.(unison.ColorProvider); ok {
		return c
	}
	return fallback
}

func (n *Node[T]) createLabelCell(c *gurps.CellData, width float32, foreground, background unison.Ink, selected bool) unison.Paneler {
	p := unison.NewPanel()
	p.SetLayout(&unison.FlexLayout{
		Columns: 1,
		HAlign:  c.Alignment,
	})
	if c.Secondary != "" {
		outer := unison.NewPanel()
		outer.SetLayout(&unison.FlexLayout{
			Columns:  2,
			HSpacing: unison.StdHSpacing / 2,
		})
		inner := unison.NewPanel()
		inner.SetLayout(&unison.FlexLayout{
			Columns: 1,
			HAlign:  c.Alignment,
		})
		inner.SetLayoutData(&unison.FlexLayoutData{
			HAlign: align.Fill,
			HGrab:  true,
		})
		outer.AddChild(inner)
		button := unison.NewButton()
		button.HideBase = true
		button.DrawableOnlyHMargin = 0
		button.DrawableOnlyVMargin = 0
		button.Font = n.primaryFieldFont()
		baseline := button.Font.Baseline()
		size := max(baseline-2, 6)
		isClosed := gurps.IsClosed(n.noteKey())
		var s *unison.SVG
		// The button is nothing but its icon, so its tooltip is also what a screen reader calls it. Clicking it
		// changes the notes disclosure, which the cell's panel is kept under (see CellCache), so the cell is built
		// anew, and this button with it, and the tooltip never needs to change once set. The key is looked up on each
		// click, as converting the row to or from a container changes its ID, and so its key, but can keep the panel.
		if isClosed {
			s = svg.NotesExpand
			button.Tooltip = newWrappedTooltip(i18n.Text("Show notes"))
		} else {
			s = svg.NotesCollapse
			button.Tooltip = newWrappedTooltip(i18n.Text("Hide notes"))
		}
		button.Drawable = &unison.DrawableSVG{
			SVG:  s,
			Size: geom.NewSize(size, size).Ceil(),
		}
		button.ClickCallback = func() {
			key := n.noteKey()
			gurps.SetClosedState(key, !gurps.IsClosed(key))
			n.table.SyncToModel()
		}
		if baseline-size > 0 {
			button.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: baseline - size}))
		}
		button.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Start})
		outer.AddChild(button)
		p.AddChild(outer)
		_, prefSize, _ := button.Sizes(geom.Size{})
		n.addLabelCell(c, inner, width-((unison.StdHSpacing/2)+prefSize.Width), c.Primary, c.InlineTag,
			n.primaryFieldFont(), foreground, background, true)
		if !isClosed {
			p.AddChild(n.createMarkdownCell(c.Secondary, width, n.secondaryFieldFont(), foreground, background,
				selected))
		}
	} else {
		n.addLabelCell(c, p, width, c.Primary, c.InlineTag, n.primaryFieldFont(), foreground, background, true)
	}
	switch {
	case c.UnsatisfiedReason != "":
		p.AddChild(n.errorTag(i18n.Text("Unsatisfied prerequisite(s)")))
	case c.PrereqContradiction != "":
		p.AddChild(makeTagForNode(i18n.Text("Contradictory prerequisite(s)"), unison.ThemeWarning,
			unison.ThemeOnWarning, n.secondaryFieldFont(), unison.TriangleExclamationSVG))
	}
	if c.TemplateInfo != "" {
		// Unlike the prerequisite tags, whose colors are fixed, this one is drawn in the row's own colors, so it must
		// follow them as the row is selected and deselected rather than keeping those it was built with (see
		// applyInkRecursively).
		tag := makeTagForNode(c.TemplateInfo, foreground, background, n.secondaryFieldFont(), svg.GCSTemplate)
		delete(tag.ClientData(), noInvertColorsMarker)
		p.AddChild(tag)
	}
	if c.UnresolvedChoice != "" {
		p.AddChild(n.errorTag(i18n.Text("Modifier choice required")))
	}
	if c.ChoiceInfo != "" {
		// Drawn in the row's own colors, as the template choice tag above is, for the same reason.
		tag := makeTagForNode(c.ChoiceInfo, foreground, background, n.secondaryFieldFont(), svg.SignPost)
		delete(tag.ClientData(), noInvertColorsMarker)
		if c.ChoiceRequired {
			// A mandatory choice yet to be made on a sheet says so right beside what it asks for.
			row := unison.NewPanel()
			row.SetLayout(&unison.FlexLayout{
				Columns:  2,
				HSpacing: unison.StdHSpacing / 2,
			})
			row.AddChild(tag)
			row.AddChild(n.errorTag(i18n.Text("Required")))
			p.AddChild(row)
		} else {
			p.AddChild(tag)
		}
	}
	if tooltip := labelCellTooltip(c); tooltip != "" {
		var workingDir string
		if wd, ok := n.table.ClientData()[WorkingDirKey]; ok {
			if wdStr, ok2 := wd.(string); ok2 {
				workingDir = wdStr
			}
		}
		p.Tooltip = newMarkdownTooltip(tooltip, workingDir)
	}
	return p
}

// labelCellTooltip returns the text of the tooltip a label cell shows. The reason for an unsatisfied prerequisite
// replaces the cell's own tooltip, since the row is flagged as broken and the reason is what the user needs to put it
// right. The explanation of a contradiction among the prerequisites is put ahead of the cell's own tooltip instead,
// since the row is enabled and in use, so what its own tooltip says still applies. The explanation of a mandatory
// modifier choice left unresolved on a sheet goes ahead of all of that.
func labelCellTooltip(c *gurps.CellData) string {
	tooltip := c.UnsatisfiedReason
	if tooltip == "" {
		tooltip = gurps.JoinTooltipBlocks(c.PrereqContradiction, c.Tooltip)
	}
	return gurps.JoinTooltipBlocks(c.UnresolvedChoice, tooltip)
}

// errorTag returns a tag that flags a problem with the row.
func (n *Node[T]) errorTag(title string) *unison.Tag {
	return makeTagForNode(title, unison.ThemeError, unison.ThemeOnError, n.secondaryFieldFont(),
		unison.TriangleExclamationSVG)
}

func makeTagForNode(title string, bg, fg unison.Ink, font unison.Font, img *unison.SVG) *unison.Tag {
	tag := unison.NewTag()
	tag.BackgroundInk = bg
	tag.OnBackgroundInk = fg
	tag.Font = font
	tag.SetBorder(unison.NewEmptyBorder(geom.NewVerticalInsets(1)))
	height := font.LineHeight() - 2
	tag.Drawable = &unison.DrawableSVG{
		SVG:  img,
		Size: geom.NewSize(height, height),
	}
	tag.SetTitle(title)
	tag.ClientData()[noInvertColorsMarker] = true
	return tag
}

func (n *Node[T]) addLabelCell(c *gurps.CellData, parent *unison.Panel, width float32, text, inlineTag string, f unison.Font, foreground, background unison.Ink, primary bool) {
	decoration := &unison.TextDecoration{
		Font:          f,
		StrikeThrough: primary && c.Disabled,
	}
	var tag *unison.Tag
	if inlineTag != "" {
		tag = unison.NewTag()
		tag.BackgroundInk = foreground
		tag.OnBackgroundInk = background
		tag.Font = &unison.DynamicFont{
			Resolver: func() unison.FontDescriptor {
				desc := f.Descriptor()
				desc.Size = max(desc.Size-2, 1)
				return desc
			},
		}
		tag.SetTitle(inlineTag)
		tag.SetEnabled(!c.Dim)
	}
	// A label reports its width rounded up to whole pixels, so the text is wrapped to a whole-pixel width. Wrapped to
	// the fractional width a cell usually has, a line that fit could still come out a fraction too wide and push what
	// sits beside it, the notes disclosure button or the inline tag, past the edge of the cell.
	var lines []*unison.Text
	if width > 0 {
		if tag != nil {
			_, size, _ := tag.Sizes(geom.Size{})
			lines = unison.NewTextWrappedLines(text, decoration, xmath.Floor(width-(size.Width+unison.StdHSpacing)))
			if len(lines) > 1 {
				lines = lines[:1]
				lines = append(lines, unison.NewTextWrappedLines(strings.TrimPrefix(text, lines[0].String()),
					decoration, xmath.Floor(width))...)
			}
		} else {
			lines = unison.NewTextWrappedLines(text, decoration, xmath.Floor(width))
		}
	} else {
		lines = unison.NewTextLines(text, decoration)
	}
	for _, line := range lines {
		label := unison.NewLabel()
		label.Font = f
		label.StrikeThrough = primary && c.Disabled
		label.HAlign = c.Alignment
		label.OnBackgroundInk = foreground
		// A line keeps the whitespace it was broken at, which only fit because the wrapping doesn't count it; left
		// on, the label would be measured with it and come out wider than the width the line was wrapped to.
		label.SetTitle(strings.TrimRightFunc(line.String(), unicode.IsSpace))
		label.SetEnabled(!c.Dim)
		if tag != nil {
			wrapper := unison.NewPanel()
			wrapper.SetLayout(&unison.FlexLayout{
				Columns:  2,
				HSpacing: unison.StdHSpacing,
				HAlign:   align.Start,
				VAlign:   align.Middle,
			})
			wrapper.AddChild(label)
			wrapper.AddChild(tag)
			parent.AddChild(wrapper)
			tag = nil
		} else {
			parent.AddChild(label)
		}
	}
}

// newCheckCell creates a cell that toggles its checked state when clicked. svgFor supplies the SVG to draw for a given
// state, or nil if nothing should be drawn for it, and onClick is called with the cell and the modifiers that were
// down at the time of the click, after c.Checked has been updated to its new state. onClick returns whether it took the
// change; if it didn't, the cell is put back the way it was, so that it never shows a state the model didn't take on.
//
// Of mouse presses, only a single click of the primary button toggles the cell. A press of any other button is left
// unconsumed for the table. A right-click never reaches the cell, although these cells sit at the front of the page
// lists where one is likely: unison takes it for the table's context menu and the table selects the row as the press
// lands (see unison.ContextMenuPressHandler); only a right-drag hands the press back. A primary press with any other
// click count is consumed but doesn't toggle, or the second click of a double-click would flip the state straight back
// at the cost of two undo edits. The drag and up callbacks report the same consumption as their press, so a press the
// cell didn't take is the table's from beginning to end.
func (n *Node[T]) newCheckCell(c *gurps.CellData, foreground unison.Ink, svgFor func(on bool) *unison.SVG,
	onClick func(label *unison.Label, mods mod.Modifiers) bool,
) *unison.Label {
	label := unison.NewLabel()
	label.VAlign = align.Start
	font := n.primaryFieldFont()
	fd := font.Descriptor()
	fd.Size -= 2
	label.Font = fd.Font()
	label.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: 1}))
	baseline := font.Baseline()
	setDrawable := func(on bool) {
		if svg := svgFor(on); svg != nil {
			label.Drawable = &unison.DrawableSVG{
				SVG:  svg,
				Size: geom.Size{Width: baseline, Height: baseline},
			}
		} else {
			label.Drawable = nil
		}
	}
	setDrawable(c.Checked)
	label.HAlign = c.Alignment
	label.OnBackgroundInk = foreground
	if c.Tooltip != "" {
		label.Tooltip = newWrappedTooltip(c.Tooltip)
	}
	// The cell is drawn as a label holding a checkmark, or nothing at all, so a screen reader is told outright that it
	// is a check box, what it is called and whether it is checked. c.Checked is read when the cell is described rather
	// than now, since a click changes it in place.
	label.Accessibility.Role = role.CheckBox
	label.Accessibility.Name = c.Name
	label.Accessibility.Callback = func(node *accessibility.Node) {
		node.HasCheck = true
		node.Checked = check.FromBool(c.Checked)
	}
	toggle := func(mods mod.Modifiers) {
		if inReadOnlyEditor(label) {
			return
		}
		c.Checked = !c.Checked
		// The new state is drawn and a layout asked for before the click is reported, since reporting it may build
		// this cell anew -- a page marked as modified re-syncs its lists, a rebuild re-syncs or replaces the table,
		// and a list that isn't on a page builds all of its rows anew when synced -- leaving this label detached from
		// the panel tree, where changing its drawable and marking it for layout and redraw would have no effect at
		// all. An editor or list dockable that is only marked as modified doesn't re-sync its table, so the row and
		// this cell are kept. c is the data the cell's panel is cached under (see CellCache.Data), so once the model
		// has followed the click, a row that is kept finds the panel matching the data derived afresh and keeps it;
		// when the model refused it, the flip below puts the data back.
		setDrawable(c.Checked)
		label.MarkForLayoutAndRedraw()
		if !onClick(label, mods) {
			c.Checked = !c.Checked
			setDrawable(c.Checked)
			label.MarkForLayoutAndRedraw()
		}
	}
	// The table synthesizes no click for Space at cell level (see newLink), so without this, Space would open the row's
	// editor rather than toggle the cell.
	label.Accessibility.ActionCallback = func(req accessibility.ActionRequest) bool {
		if req.Action != accessibility.Press && req.Action != accessibility.Toggle {
			return false
		}
		n.keepingCellCursor(func() { toggle(mod.None) })
		return true
	}
	tookPress := false
	label.MouseDownCallback = func(_ geom.Point, button, clickCount int, mods mod.Modifiers) bool {
		tookPress = button == unison.ButtonLeft
		if !tookPress || clickCount != 1 {
			return tookPress
		}
		toggle(mods)
		return true
	}
	label.MouseDragCallback = func(_ geom.Point, _ int, _ mod.Modifiers) bool {
		return tookPress
	}
	label.MouseUpCallback = func(_ geom.Point, _ int, _ mod.Modifiers) bool {
		return tookPress
	}
	return label
}

// keepingCellCursor runs f and then puts the cell cursor back on the cell it was on, if it was on one and the row still
// exists. A change to a row re-syncs the table and restores its selection, which moves the cursor to row level, so a
// second Space would open the row's editor rather than work the cell again.
func (n *Node[T]) keepingCellCursor(f func()) {
	table := n.table
	if table == nil {
		f()
		return
	}
	row, col := table.LeadRowIndex(), table.LeadColumnIndex()
	if row < 0 || col < 0 {
		f()
		return
	}
	id := table.RowFromIndex(row).ID()
	f()
	for i := range table.LastRowIndex() + 1 {
		if table.RowFromIndex(i).ID() == id {
			if table.LeadRowIndex() != i || table.LeadColumnIndex() != col {
				table.SetLeadCell(i, col)
			}
			return
		}
	}
}

func (n *Node[T]) createToggleCell(c *gurps.CellData, foreground unison.Ink) unison.Paneler {
	toggle := n.newCheckCell(c, foreground,
		func(on bool) *unison.SVG {
			if on {
				return unison.CheckmarkSVG
			}
			return nil
		},
		func(label *unison.Label, _ mod.Modifiers) bool {
			// The pick of a mandatory modifier choice on a sheet can't be turned off (see
			// gurps.IsLockedModifierChoiceSelection), nor can disabled equipment be equipped or unequipped, so the
			// click is refused, which puts the checkmark back.
			if gurps.IsLockedModifierChoiceSelection(n.data) {
				return false
			}
			if eqp, ok := any(n.data).(*gurps.Equipment); ok && !eqp.Enabled() {
				return false
			}
			// The panel may be kept under the flipped data (see newCheckCell), so it has to be dimmed as one built
			// from that data would be, and before the click is reported, for the same reason the drawable is.
			label.SetEnabled(!c.Checked || !c.Dim)
			if !handleCheck(n.data, label, c.Checked) {
				MarkModified(label)
			}
			return true
		})
	// Only a checkmark is dimmed. An unchecked cell, which draws nothing, stays enabled even when its data is dimmed,
	// as it always is for unequipped equipment, so that an assistive technology can still press it.
	toggle.SetEnabled(!c.Checked || !c.Dim)
	return toggle
}

// createSwitchCell creates the cell for a switch column. Unlike a toggle cell, the "off" state is drawn (as a dash), so
// that a row whose switch is off can be told apart from one that has nothing to switch (which gets a blank cell).
func (n *Node[T]) createSwitchCell(c *gurps.CellData, foreground unison.Ink) unison.Paneler {
	label := n.newCheckCell(c, foreground,
		func(on bool) *unison.SVG {
			if on {
				return unison.CheckmarkSVG
			}
			return unison.DashSVG
		},
		func(label *unison.Label, mods mod.Modifiers) bool {
			return toggleFeatureSwitch(n, label, c.Checked, mods.OptionDown())
		})
	if c.Dim {
		// The switch of an item that isn't currently contributing its features -- a piece of equipment that isn't
		// equipped, a trait that is turned off -- is drawn dimmed to say so, which is exactly what a disabled label
		// does. The switch stays as usable as an undimmed one: the table hands mouse events and tooltip lookups to its
		// cells itself, without consulting their enabled state, and the window only ever sees the table. Dimming says
		// the switch has no effect at the moment, not that it can't be thrown.
		label.SetEnabled(false)
	}
	return label
}

// handleCheck applies the new state of a toggle cell to the item it belongs to, registering an undoable edit. Each kind
// of toggle is routed through the same machinery as the command that performs it for a whole selection, so that a click
// and the command behave identically -- same undo title, same choice of recalculating and of rebuilding the owner
// versus merely marking it as modified. It returns true if it has already reported the change to the item's owner, in
// which case the caller must not mark the owner as modified on top of that; false means the caller has to do the
// reporting, which is also what happens when there is no owner to report to.
func handleCheck(data any, toggle unison.Paneler, checked bool) bool {
	owner := unison.AncestorOrSelf[Rebuildable](toggle)
	switch item := data.(type) {
	case *gurps.Equipment:
		adjustEquipped(owner, toggle, item, checked)
	case *gurps.TraitModifier:
		adjustModifierEnabled(owner, toggle, item, checked)
	case *gurps.EquipmentModifier:
		adjustModifierEnabled(owner, toggle, item, checked)
	case *gurps.Weapon:
		adjustHidden(owner, toggle, item, checked)
	default:
		return false
	}
	return !xreflect.IsNil(owner)
}

func convertLinksForPageRef(in string) (string, *unison.SVG) {
	switch {
	case unison.HasURLPrefix(in):
		return i18n.Text("link"), svg.Link
	case strings.HasPrefix(in, "md:"):
		return i18n.Text("md"), svg.MarkdownFile
	default:
		return in, nil
	}
}

// pageRefSeparator is placed between page references when more than one is shown in a cell.
const pageRefSeparator = ", "

func (n *Node[T]) createPageRefCell(c *gurps.CellData, width float32, foreground unison.Ink) unison.Paneler {
	font := n.primaryFieldFont()
	parts := ExtractPageReferences(c.Primary)
	if len(parts) == 0 {
		return n.createPageRefLink(c, "", font, foreground)
	}
	links := make([]*unison.Label, len(parts))
	for i, part := range parts {
		links[i] = n.createPageRefLink(c, part, font, foreground)
	}
	// Determine how many references fit within the available width, always showing at least the first one, with a
	// trailing "+" on the last shown one to indicate there are more. A non-positive width means the preferred size is
	// being asked for (e.g. while the table sizes its columns), so only one reference is claimed and the column doesn't
	// grow to fit them all.
	shown := 1
	if width > 0 {
		sepWidth := pageRefSeparatorWidth(font)
		_, size, _ := links[0].Sizes(geom.Size{})
		cumulative := size.Width
		for i := 1; i < len(parts); i++ {
			_, size, _ = links[i].Sizes(geom.Size{})
			next := cumulative + sepWidth + size.Width
			if next > width {
				break
			}
			cumulative = next
			shown = i + 1
		}
	}
	if shown < len(parts) {
		last := links[shown-1]
		last.SetTitle(last.String() + "+")
		last.Tooltip = newWrappedTooltip(strings.Join(parts, "\n"))
	}
	if shown == 1 {
		return links[0]
	}
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns: shown*2 - 1,
		VAlign:  align.Start,
	})
	for i := range shown {
		if i > 0 {
			separator := unison.NewLabel()
			separator.Font = font
			separator.OnBackgroundInk = foreground
			separator.VAlign = align.Start
			separator.SetTitle(pageRefSeparator)
			panel.AddChild(separator)
		}
		panel.AddChild(links[i])
	}
	return panel
}

// createPageRefLink creates an individually-clickable hyperlink for a single page reference. An empty ref produces a
// disabled, empty link.
func (n *Node[T]) createPageRefLink(c *gurps.CellData, ref string, font unison.Font, foreground unison.Ink) *unison.Label {
	var title, tooltip string
	var icon *unison.DrawableSVG
	if ref != "" {
		var img *unison.SVG
		title, img = convertLinksForPageRef(ref)
		if img != nil {
			title = ""
			height := font.Baseline()
			icon = &unison.DrawableSVG{
				SVG:  img,
				Size: geom.NewSize(height, height).Ceil(),
			}
			tooltip = ref
		}
	}
	theme := unison.DefaultLinkTheme
	theme.OnBackgroundInk = foreground
	theme.Font = font
	link := newLink(title, tooltip, ref, &theme, func(_ unison.Paneler, target string) {
		OpenPageReference(target, c.Secondary, nil)
	})
	// In a table cell the focus stays with the table, so the link needs no tab stop, nor the room unison leaves either
	// side of it for a focus outline, which would widen every page reference column.
	link.SetFocusable(false)
	link.SetBorder(nil)
	link.VAlign = align.Start
	if icon != nil {
		link.Drawable = icon
	}
	if tooltip != "" {
		link.Tooltip = newWrappedTooltip(tooltip)
	}
	link.SetEnabled(!c.Dim && (title != "" || icon != nil))
	if ref == "" {
		// A row with no page reference shows nothing, so there is no link for a screen reader to offer; the cell is
		// announced as blank instead.
		link.Accessibility.Role = role.None
	}
	return link
}

// pageRefSeparatorWidth returns the width of the separator placed between adjacent page references.
func pageRefSeparatorWidth(font unison.Font) float32 {
	return unison.NewText(pageRefSeparator, &unison.TextDecoration{Font: font}).Width()
}

// pageRefColumnFullWidth returns the content width needed to display every page reference in the given column for this
// row. It returns -1 if the column isn't a page reference column. This is used to let page reference columns claim
// excess width (up to the point where all references are visible) before the remainder goes to the excess-width column.
func (n *Node[T]) pageRefColumnFullWidth(col int) float32 {
	memo := n.cell(col)
	stamp := n.currentCellStamp()
	if memo.refWidthAt == stamp {
		return memo.refWidth
	}
	memo.refWidth = n.computePageRefColumnFullWidth(col)
	memo.refWidthAt = stamp
	return memo.refWidth
}

func (n *Node[T]) computePageRefColumnFullWidth(col int) float32 {
	c := n.cellData(col, true)
	if c.Type != cell.PageRef {
		return -1
	}
	parts := ExtractPageReferences(c.Primary)
	if len(parts) == 0 {
		return 0
	}
	font := n.primaryFieldFont()
	sepWidth := pageRefSeparatorWidth(font)
	var width float32
	for i, part := range parts {
		if i > 0 {
			width += sepWidth
		}
		_, size, _ := n.createPageRefLink(&c, part, font, unison.DefaultLinkTheme.OnBackgroundInk).Sizes(geom.Size{})
		width += size.Width
	}
	return width
}

// excessColumnCollapsedWidth returns the column width needed to display the given column's primary content with any
// secondary (note) content collapsed, including padding and any hierarchy indent for this row. It returns -1 if the
// column isn't a text column. This is used to reserve enough room for the primary column before handing leftover space
// to page reference columns.
func (n *Node[T]) excessColumnCollapsedWidth(col int) float32 {
	memo := n.cell(col)
	stamp := n.currentCellStamp()
	if memo.collapsedAt != stamp {
		memo.collapsed = n.computeExcessColumnCollapsedWidth(col)
		memo.collapsedAt = stamp
	}
	width := memo.collapsed
	if width < 0 {
		return width
	}
	// The hierarchy indent is added each time, since it isn't part of the data.
	if n.table.Columns[col].ID == n.table.HierarchyColumnID {
		if indent := n.table.CurrentHierarchyIndent(); indent > 0 {
			depth := 0
			for p := n.parent; p != nil; p = p.parent {
				depth++
			}
			width += n.table.Padding.Left + indent*float32(depth+1)
		}
	}
	return width
}

// computeExcessColumnCollapsedWidth is excessColumnCollapsedWidth without the memo and the hierarchy indent.
func (n *Node[T]) computeExcessColumnCollapsedWidth(col int) float32 {
	c := n.cellData(col, true)
	if c.Type != cell.Text {
		return -1
	}
	hadSecondary := c.Secondary != ""
	c.Secondary = ""
	_, size, _ := n.createLabelCell(&c, 0, unison.ThemeOnSurface, unison.ThemeSurface, false).AsPanel().Sizes(geom.Size{})
	width := size.Width + n.table.Padding.Left + n.table.Padding.Right
	if hadSecondary {
		// Account for the notes disclosure button shown beside the primary text.
		width += n.primaryFieldFont().Baseline() + unison.StdHSpacing/2
	}
	return width
}

func (n *Node[T]) primaryFieldFont() unison.Font {
	if n.forPage {
		return fonts.PageFieldPrimary
	}
	return unison.FieldFont
}

func (n *Node[T]) secondaryFieldFont() unison.Font {
	if n.forPage {
		return fonts.PageFieldSecondary
	}
	return fonts.FieldSecondary
}

// FindRowIndexByID returns the row index of the row with the given ID in the given table.
func FindRowIndexByID[T gurps.Node[T]](table *unison.Table[*Node[T]], id tid.TID) int {
	_, i := rowIndex(id, 0, table.RootRows())
	return i
}

func rowIndex[T gurps.Node[T]](id tid.TID, startIndex int, rows []*Node[T]) (updatedStartIndex, result int) {
	for _, row := range rows {
		if id == row.data.ID() {
			return 0, startIndex
		}
		startIndex++
		if row.IsOpen() {
			if startIndex, result = rowIndex(id, startIndex, row.Children()); result != -1 {
				return 0, result
			}
		}
	}
	return startIndex, -1
}

// InsertItems into a table. A new modifier that goes into a modifier choice arrives turned off, since nothing makes it
// the one to pick.
func InsertItems[T gurps.Node[T]](owner Rebuildable, table *unison.Table[*Node[T]], topList func() []T, setTopList func([]T), rowData func(table *unison.Table[*Node[T]]) []*Node[T], items ...T) {
	if len(items) == 0 {
		return
	}
	undo := beginTableUndo(table, i18n.Text("Insert %s", items[0].Kind()), nil, nil)
	var target, zero T
	i := table.FirstSelectedRowIndex()
	if i != -1 {
		row := table.RowFromIndex(i)
		if target = row.Data(); target != zero {
			if row.CanHaveChildren() {
				SetParents(items, target)
				row.data.SetChildren(append(row.data.NodeChildren(), items...))
			} else {
				// The items go in after the target, within its parent's children or the top-level list.
				parent := row.Parent()
				if parentData := parent.Data(); parentData != zero {
					SetParents(items, parentData)
					children := parent.data.NodeChildren()
					parent.data.SetChildren(slices.Insert(children, slices.Index(children, target)+1, items...))
				} else {
					SetParents(items, zero)
					list := topList()
					setTopList(slices.Insert(list, slices.Index(list, target)+1, items...))
				}
			}
		}
	}
	if target == zero {
		// Nothing usable was selected, so append to the end of the top-level list.
		SetParents(items, zero)
		setTopList(append(topList(), items...))
	}
	for _, item := range items {
		if _, inChoice := gurps.ModifierChoiceFor(item); inChoice {
			gurps.SetModifierEnabled(item, false)
		}
	}
	table.SetRootRows(rowData(table))
	table.ValidateScrollRoot()
	table.RequestFocus()
	selMap := make(map[tid.TID]bool)
	for _, item := range items {
		selMap[item.ID()] = true
	}
	table.SetSelectionMap(selMap)
	table.ScrollRowCellIntoView(table.LastSelectedRowIndex(), 0)
	table.ScrollRowCellIntoView(table.FirstSelectedRowIndex(), 0)
	commitTableUndo(table, undo)
	// The change is reported once, by rebuilding the owner: an inserted item can bring a weapon, reaction or
	// conditional modifier list onto the page, or the switch column into one of the lists, and only a rebuild creates
	// those.
	rebuildAsModified(owner, true)
}

// SetParents of each item.
func SetParents[T gurps.Node[T]](items []T, parent T) {
	for _, item := range items {
		item.SetParent(parent)
	}
}

// ExtractNodeDataFromList returns the underlying node data.
func ExtractNodeDataFromList[T gurps.Node[T]](list []*Node[T]) []T {
	dataList := make([]T, 0, len(list))
	for _, child := range list {
		dataList = append(dataList, child.data)
	}
	return dataList
}
