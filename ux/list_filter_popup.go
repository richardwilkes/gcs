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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

// Seams so the commands can be exercised without a window.
var (
	showFilterEditor      = showListFilterDialog
	confirmFilterDeletion = func(name string) bool {
		return unison.QuestionDialog(i18n.Text("Delete the filter '%s'?", name),
			i18n.Text("This cannot be undone.")) == unison.ModalResponseOK
	}
)

// listFilterObserver is implemented by a dockable that shows a saved filter popup, so that it can be told when the
// saved filters of a list type were changed through the popup of another dockable.
type listFilterObserver interface {
	// listFiltersChanged is called after the saved filters for the list type with the given key were changed through
	// source, which is nil when the change didn't come from a popup.
	listFiltersChanged(key string, source *listFilterPopup)
}

// notifyListFiltersChanged tells every open dockable that the saved filters for the list type with the given key have
// changed, so that the ones showing that list type can bring their popups into line. The popup the change was made
// through is passed as source, since it has already done so.
func notifyListFiltersChanged(key string, source *listFilterPopup) {
	for _, d := range AllDockables() {
		if observer, ok := d.(listFilterObserver); ok {
			observer.listFiltersChanged(key, source)
		}
	}
}

// listFilterPopupSpec is what the saved filter popup needs from the dockable that shows it.
type listFilterPopupSpec struct {
	// key is the list type key the saved filters are stored under, such as "adq".
	key string
	// fields are the fields a filter for this list type may test.
	fields []filterFieldInfo
	// current returns the filter in force, which is nil when there is none.
	current func() *gurps.ListFilter
	// choose puts a filter in force. A nil filter drops the one that was, leaving the list to the quick filter alone.
	choose func(f *gurps.ListFilter)
}

// listFilterPopup is the toolbar popup that chooses among a list type's saved filters and creates, edits and deletes
// them. The popup's items are rebuilt from the global settings each time it is about to be shown, since another
// dockable showing the same list type may have changed them in the meantime.
type listFilterPopup struct {
	popup  *unison.PopupMenu[string]
	button *filterPopupButton
	spec   listFilterPopupSpec
	// filters runs parallel to the popup's items, holding the filter each item stands for. It is nil for the No Filter
	// entry, the separators and the three commands, since separators occupy an index of their own.
	filters []*gurps.ListFilter
	// noneIndex is the index of the "No Filter" entry, which stands for no saved filter. It follows the three commands,
	// which lead the popup so that they stay in reach of a long list of filters, and the separator after them.
	noneIndex   int
	newIndex    int
	editIndex   int
	deleteIndex int
	// rebuilding suppresses the selection callback while the items are being replaced, since emptying the popup and
	// filling it again moves the selection without the user having chosen anything.
	rebuilding bool
}

// newListFilterPopup creates the saved filter popup for the list type the spec names.
func newListFilterPopup(spec listFilterPopupSpec) *listFilterPopup {
	p := &listFilterPopup{
		popup: unison.NewPopupMenu[string](),
		spec:  spec,
	}
	p.popup.WillShowMenuCallback = func(_ *unison.PopupMenu[string]) {
		if p.rebuildItems() {
			// The filter that was in force is gone, so the list falls back to the quick filter alone.
			p.spec.choose(nil)
		}
	}
	p.popup.ChoiceMadeCallback = func(popup *unison.PopupMenu[string], index int, _ string) {
		// The commands act and leave the selection where it was; only a filter, or No Filter, becomes the selection.
		switch index {
		case p.newIndex:
			p.newFilter(gurps.NewListFilter(""))
		case p.editIndex:
			p.editFilter()
		case p.deleteIndex:
			p.deleteFilter()
		default:
			popup.SelectIndex(index)
		}
	}
	p.popup.SelectionChangedCallback = func(popup *unison.PopupMenu[string]) {
		p.syncTooltip()
		if p.rebuilding {
			return
		}
		i := popup.SelectedIndex()
		p.enableCommands(i != p.noneIndex)
		p.spec.choose(p.filterAt(i))
	}
	p.button = newFilterPopupButton(p.popup, svg.Filter, func() bool { return p.spec.current() != nil })
	// The popup sits at the end of the toolbar, with no label of its own to name it by.
	p.popup.Accessibility.Name = i18n.Text("Saved Filters")
	p.popup.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	// Fill the popup in now so that it shows "No Filter" before it has ever been opened.
	p.rebuildItems()
	return p
}

// rebuildItems replaces the popup's items with the saved filters as they now stand, keeping the filter in force
// selected. It reports whether that filter is no longer among them, which happens when another dockable showing the
// same list type deleted it, in which case the selection has fallen back to No Filter.
func (p *listFilterPopup) rebuildItems() (lostCurrent bool) {
	defer func(saved bool) { p.rebuilding = saved }(p.rebuilding)
	p.rebuilding = true
	current := p.spec.current()
	p.popup.RemoveAllItems()
	p.filters = nil
	addItem := func(title string, f *gurps.ListFilter) int {
		index := len(p.filters)
		p.popup.AddItem(title)
		p.filters = append(p.filters, f)
		return index
	}
	addSeparator := func() {
		p.popup.AddSeparator()
		p.filters = append(p.filters, nil)
	}
	p.newIndex = addItem(i18n.Text("New Filter…"), nil)
	p.editIndex = addItem(i18n.Text("Edit Filter…"), nil)
	p.deleteIndex = addItem(i18n.Text("Delete Filter…"), nil)
	addSeparator()
	p.noneIndex = addItem(i18n.Text("No Filter"), nil)
	index := p.noneIndex
	if saved := gurps.GlobalSettings().ListFiltersFor(p.spec.key); len(saved) != 0 {
		addSeparator()
		for _, f := range saved {
			// The filter in force is found by identity rather than by name, since a saved filter may bear the same
			// name as one of the popup's own entries.
			if i := addItem(f.Name, f); f == current {
				index = i
			}
		}
	}
	p.popup.SelectIndex(index)
	p.enableCommands(index != p.noneIndex)
	p.syncTooltip() // A rename keeps the selection where it was, yet changes the name shown.
	p.popup.MarkForRedraw()
	return current != nil && index == p.noneIndex
}

// syncTooltip puts a filter name the button cuts short at the head of the tooltip, so that it can still be read in
// full.
func (p *listFilterPopup) syncTooltip() {
	detail := i18n.Text("Choose a saved filter, or create, edit and delete them")
	if name, cut := p.button.textCut(); cut {
		detail = name + "\n\n" + detail
	}
	p.popup.Tooltip = newWrappedTooltipWithSecondaryText(p.popup.Accessibility.Name, detail)
}

// filterAt returns the filter the item at the given index stands for, or nil when the item is No Filter, a separator, a
// command or out of range.
func (p *listFilterPopup) filterAt(index int) *gurps.ListFilter {
	if index < 0 || index >= len(p.filters) {
		return nil
	}
	return p.filters[index]
}

// enableCommands turns the Edit and Delete commands on or off. They only apply to a saved filter, so they are off
// while none is in force.
func (p *listFilterPopup) enableCommands(enabled bool) {
	p.popup.SetItemEnabledAt(p.editIndex, enabled)
	p.popup.SetItemEnabledAt(p.deleteIndex, enabled)
}

// selectFilter rebuilds the popup around the given filter, selects it, and puts it in force. Pass nil to drop the
// saved filter, leaving the list to the quick filter alone.
func (p *listFilterPopup) selectFilter(f *gurps.ListFilter) {
	func() {
		defer func(saved bool) { p.rebuilding = saved }(p.rebuilding)
		p.rebuilding = true
		p.rebuildItems()
		index := p.noneIndex
		if f != nil {
			if i := slices.Index(p.filters, f); i != -1 {
				index = i
			}
		}
		p.popup.SelectIndex(index)
		p.enableCommands(index != p.noneIndex)
	}()
	// The filter is put in force unconditionally: editing a filter without renaming it doesn't move the selection, so
	// the selection callback would not fire, yet the contents changed and the list has to be filtered again.
	p.spec.choose(f)
}

// refresh brings the popup into line with the saved filters after they were changed elsewhere: the items are rebuilt,
// and the filter in force is either put through again, since its contents may have changed, or dropped when it no
// longer exists.
func (p *listFilterPopup) refresh() {
	if p.rebuildItems() {
		p.spec.choose(nil)
	} else if current := p.spec.current(); current != nil {
		p.spec.choose(current)
	}
}

// newFilter puts up the editor for f as a new filter and, if it is accepted, saves it and puts it in force.
func (p *listFilterPopup) newFilter(f *gurps.ListFilter) {
	if showFilterEditor(i18n.Text("New Filter"), p.spec.key, f, p.spec.fields, nil) {
		gurps.GlobalSettings().AddListFilter(p.spec.key, f)
		p.filtersChanged(f)
	}
}

// editFilter puts up the editor for the filter in force. A clone is edited so that a cancel leaves the saved filter
// untouched. When the edit is accepted, the clone's contents are copied into the saved filter rather than replacing
// it, so that every other dockable with the same filter in force keeps it, and sees the change.
func (p *listFilterPopup) editFilter() {
	current := p.spec.current()
	if current == nil {
		return
	}
	clone := current.Clone()
	if showFilterEditor(i18n.Text("Edit Filter"), p.spec.key, clone, p.spec.fields, current) {
		*current = *clone
		gurps.GlobalSettings().ResortListFilters(p.spec.key) // The name may have changed.
		p.filtersChanged(current)
	}
}

// deleteFilter removes the filter in force, once the user has confirmed it, and hands the list back to the quick
// filter.
func (p *listFilterPopup) deleteFilter() {
	current := p.spec.current()
	if current == nil {
		return
	}
	if confirmFilterDeletion(current.Name) {
		gurps.GlobalSettings().RemoveListFilter(p.spec.key, current)
		p.filtersChanged(nil)
	}
}

// filtersChanged finishes a change made to the saved filters through this popup: the settings are saved, the given
// filter is put in force here, and every other dockable showing the same list type is told to catch up.
func (p *listFilterPopup) filtersChanged(f *gurps.ListFilter) {
	saveGlobalSettings()
	p.selectFilter(f)
	notifyListFiltersChanged(p.spec.key, p)
}

// popupWidthSample is the text whose width, in a popup's font, sets how wide a filter popup's button is and how much of
// the chosen item it shows.
const popupWidthSample = "MMMMMMMMMM"

// filterPopupButton draws a filter popup's button. It is normally just wide enough to show popupWidthSample, whatever
// the popup's items, cutting the chosen item short with an ellipsis past that, but shows only its icon while compact.
// The menu it opens still shows every item in full.
type filterPopupButton struct {
	popup *unison.PopupMenu[string]
	icon  *unison.SVG
	// inEffect reports whether the popup narrows the list, which is when the compact icon takes the focus color.
	inEffect func() bool
	compact  bool
}

// newFilterPopupButton takes over the sizing and drawing of the popup's button.
func newFilterPopupButton(popup *unison.PopupMenu[string], icon *unison.SVG, inEffect func() bool) *filterPopupButton {
	b := &filterPopupButton{popup: popup, icon: icon, inEffect: inEffect}
	popup.SetSizer(b.sizes)
	popup.DrawCallback = b.draw
	return b
}

func (b *filterPopupButton) sizes(hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
	_, prefSize, _ = b.popup.DefaultSizes(hint)
	if b.compact {
		prefSize.Width = prefSize.Height
	} else {
		// The space around the text is added as unison.PopupMenu.DefaultSizes adds it.
		width := popupTextWidth(b.popup, popupWidthSample) + b.popup.HMargin*2 + 2 + prefSize.Height*0.75
		if border := b.popup.Border(); border != nil {
			width += border.Insets().Width()
		}
		prefSize.Width = xmath.Ceil(width)
	}
	return prefSize, prefSize, prefSize
}

func (b *filterPopupButton) draw(canvas *unison.Canvas, dirty geom.Rect) {
	if b.compact {
		// Drawn as unison.Button draws an icon with no base.
		r := b.popup.ContentRect(false)
		if b.popup.Focused() {
			unison.DrawRoundedRectBase(canvas, r, b.popup.CornerRadius, 2, unison.Transparent, b.popup.SelectionInk)
		}
		ink := b.popup.OnBackgroundInk
		if b.inEffect() {
			ink = unison.ThemeFocus
		}
		size := b.popup.Font.Baseline()
		r = geom.NewRect(r.X+(r.Width-size)/2, r.Y+(r.Height-size)/2, size, size)
		b.icon.DrawInRectPreservingAspectRatio(canvas, r, nil, ink.Paint(canvas, r, paintstyle.Fill))
		return
	}
	// The button draws the chosen item through the renderer the menu uses too, so it is swapped only while drawing.
	render := b.popup.ItemRendererCallback
	defer func() { b.popup.ItemRendererCallback = render }()
	b.popup.ItemRendererCallback = func(item string) string {
		if render != nil {
			item = render(item)
		}
		return popupFittedText(b.popup, item)
	}
	b.popup.DefaultDraw(canvas, dirty)
}

// textCut returns the text of the popup's chosen item, and whether its button cuts that text short, which a compact
// one always does, showing none of it.
func (b *filterPopupButton) textCut() (text string, cut bool) {
	text = b.popup.Text()
	return text, b.compact || popupFittedText(b.popup, text) != text
}

// compactingToolbarLayout lays a toolbar out in one row, as FinishToolbarLayout does, but turns its filter popups into
// icons while the toolbar is too narrow to show them in full beside its other controls at their least widths.
type compactingToolbarLayout struct {
	*unison.FlexLayout
	buttons []*filterPopupButton
	// changed is called after the buttons turn compact or back.
	changed func()
}

// PerformLayout implements unison.Layout.
func (l *compactingToolbarLayout) PerformLayout(target *unison.Panel) {
	was := l.buttons[0].compact
	l.setCompact(false)
	minSize, _, _ := l.LayoutSizes(target, geom.Size{})
	compact := minSize.Width > target.FrameRect().Width
	l.setCompact(compact)
	if compact != was {
		for _, b := range l.buttons {
			b.popup.MarkForRedraw()
		}
		l.changed()
	}
	l.FlexLayout.PerformLayout(target)
}

func (l *compactingToolbarLayout) setCompact(compact bool) {
	for _, b := range l.buttons {
		b.compact = compact
	}
}

// popupFittedText returns text as the popup's button shows it. Text wider than popupWidthSample is cut to as much of
// its start as fits that width with an ellipsis after it, the ellipsis counted in the width and no space left before
// it.
func popupFittedText(popup *unison.PopupMenu[string], text string) string {
	t := unison.NewText(text, &unison.TextDecoration{Font: popup.Font})
	room := popupTextWidth(popup, popupWidthSample)
	if t.Width() <= room {
		return text
	}
	room -= popupTextWidth(popup, "…")
	runes := t.Runes()
	n := len(runes)
	for n > 0 && t.PositionForRuneIndex(n) > room {
		n--
	}
	return strings.TrimRightFunc(string(runes[:n]), unicode.IsSpace) + "…"
}

// popupTextWidth returns the width of text in the popup's font.
func popupTextWidth(popup *unison.PopupMenu[string], text string) float32 {
	return unison.NewText(text, &unison.TextDecoration{Font: popup.Font}).Width()
}
