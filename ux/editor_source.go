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
	"math"
	"reflect"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
)

func (e *editor[N, D]) newSourceButton() *unison.Button {
	b := unison.NewSVGButton(svg.Database)
	b.Tooltip = newWrappedTooltip(i18n.Text("ID & Library Source"))
	b.ClickCallback = func() { showMenu(b.AsPanel(), e.sourceMenuEntries()) }
	return b
}

// sourceMenuEntries returns the entries of the source menu, which describe the target as it would be with the editor's
// pending changes applied. A template choice container gets only its ID, since applying the editor's changes clears any
// source it has (see clearSourceOfTemplatePicker).
func (e *editor[N, D]) sourceMenuEntries() []menuEntry {
	entries := []menuEntry{copyEntry(i18n.Text("ID"), string(e.target.ID()))}
	edited := e.editedClone()
	if gurps.IsTemplateChoiceContainer(edited) {
		return entries
	}
	src := edited.GetSource()
	for _, part := range []struct{ label, value string }{
		{i18n.Text("Source ID"), string(src.TID)},
		{i18n.Text("Source Library"), src.Library},
		{i18n.Text("Source Path"), src.Path},
	} {
		if part.value != "" {
			entries = append(entries, copyEntry(part.label, part.value))
		}
	}
	state, data := gurps.MatchSource(edited)
	return append(
		entries,
		menuEntry{Label: state.String()},
		menuEntry{
			Label:    i18n.Text("Sync with Source"),
			Act:      e.syncWithSource,
			Disabled: state != srcstate.Mismatched || syncChangesKind(edited, data),
		},
		menuEntry{
			Label: i18n.Text("Clear Source"),
			Act:   e.clearSource,
			// Not src.IsZero(), which is also true of a partial source, and that can still be cleared.
			Disabled: src == gurps.Source{},
		},
	)
}

func copyEntry(label, value string) menuEntry {
	return menuEntry{
		Label: i18n.Text("%s: %s", label, value),
		Act:   func() { unison.ClipboardSetText(value) },
	}
}

// syncChangesKind returns true if target and source, the library's copy of it, are equipment of different kinds: one a
// container and the other not, or two kinds of container. An editor can't stage such a sync, since its data doesn't
// carry the kind.
func syncChangesKind(target, source any) bool {
	eqp, ok := target.(*gurps.Equipment)
	if !ok {
		return false
	}
	var other *gurps.Equipment
	if other, ok = source.(*gurps.Equipment); !ok {
		return false
	}
	return eqp.Container() != other.Container() || (eqp.Container() && other.ContainerType != eqp.ContainerType)
}

// syncWithSource syncs the editor's data with the library's copy as a pending change, keeping the pending changes to
// fields a sync doesn't cover, and rebuilds the content, which clears the editor's undo history. Editors open on the
// data's modifiers and weapons are closed first (see closeSubEditorsForSync), since the sync replaces what they edit:
// the modifiers with copies of themselves and the weapons with the library's.
func (e *editor[N, D]) syncWithSource() {
	// Noted before the sub-editors are closed, since closing one gives the focus to the list it was opened from and
	// scrolls the editor to that list.
	focus := e.noteFocus()
	closed := e.closeSubEditorsForSync()
	if focus.gone() {
		// The focus was within a sub-editor that is now closed, so it stays where closing that one left it if that is
		// within the content, as the list it was opened from is, and goes to the content otherwise.
		if focus = e.noteFocus(); focus == nil || len(focus.path) == 0 {
			e.content.RequestFocus()
			focus = e.noteFocus()
		}
	}
	if closed {
		// Released before the sync, so that what a field does on losing the focus is done while it is still part of the
		// editor and before the data is replaced.
		e.releaseContentFocus()
		edited := e.editedClone()
		if state, data := gurps.MatchSource(edited); state == srcstate.Mismatched && !syncChangesKind(edited, data) {
			before := modifierChoicePicker(edited)
			edited.SyncWithSource()
			e.editorData.CopyFrom(edited)
			if choice := modifierChoicePicker(edited); choice != before {
				e.syncedChoice = &choice
			}
			e.rebuildContent()
		}
	}
	focus.restore(e.content, e.scroll)
}

// closeSubEditorsForSync closes the editors open on the data's modifiers and weapons, as CloseGroup does, returning
// false if one of them stays open. A sync replaces every weapon with the library's, so saving the changes pending in a
// weapon's editor would only have them thrown away, and such an editor asks whether to discard them instead. A sync
// leaves the modifiers alone, so what is saved from a modifier's editor is kept. Closing either leaves the sync still
// to be done: of what they edit, only the weapons count toward whether the data matches its source, and their pending
// changes are not saved.
func (e *editor[N, D]) closeSubEditorsForSync() bool {
	var weaponEditors []*editor[*gurps.Weapon, *gurps.Weapon]
	traverseGroup(e, func(target GroupedCloser) bool {
		if weaponEditor, ok := target.(*editor[*gurps.Weapon, *gurps.Weapon]); ok {
			weaponEditor.discardReason = i18n.Text("Syncing with the source replaces this weapon.")
			weaponEditors = append(weaponEditors, weaponEditor)
		}
		return false
	})
	closed := CloseGroup(e)
	for _, weaponEditor := range weaponEditors {
		weaponEditor.discardReason = ""
	}
	return closed
}

// contentFocus records where the keyboard focus was and how the editor was scrolled, so that both can be put back once
// the editor's sub-editors have been closed and its content rebuilt, in which a control may have moved among its
// siblings, or gone.
type contentFocus struct {
	panel *unison.Panel
	// path leads from the content down to panel, and is empty for a panel outside the content.
	path []focusStep
	// field is the text and selection of a panel that is a field, including which end of the selection stays put as
	// it is extended, and nil for any other panel.
	field *unison.FieldState
	// selection is the selected rows of a panel that is a table, by ID, and rows the same by index. Both are empty for
	// any other panel.
	selection map[tid.TID]bool
	rows      []int
	// origin is where the top left corner of a panel within the content was, in the content's coordinates.
	origin geom.Point
	// h and v are the editor's scroll position.
	h, v float32
}

// focusStep is one step of a contentFocus path: a child of the panel the step before it led to, or of the content for
// the first.
type focusStep struct {
	// kind is the type of the child and label the text of the label naming it, if any.
	kind  reflect.Type
	label string
	// index is where the child was among its parent's children.
	index int
}

// statefulField is what a field offers that restoring the focus to it needs.
type statefulField interface {
	Text() string
	GetFieldState() *unison.FieldState
	ApplyFieldState(state *unison.FieldState)
}

// selectableTable is what a table offers that restoring its selection needs.
type selectableTable interface {
	CopySelectionMap() map[tid.TID]bool
	SetSelectionMap(selMap map[tid.TID]bool)
	SelectionCount() int
	IsRowSelected(index int) bool
	LastRowIndex() int
	SelectByIndex(indexes ...int)
}

// noteFocus returns a record of where the keyboard focus is and how the editor is scrolled, or nil if nothing holds
// the focus.
func (e *editor[N, D]) noteFocus() *contentFocus {
	focus := e.Window().CurrentFocus()
	if focus == nil {
		return nil
	}
	cf := newContentFocus(e.content, focus)
	cf.h, cf.v = e.scroll.Position()
	return cf
}

func (e *editor[N, D]) releaseContentFocus() {
	if wnd := e.Window(); unison.AncestorIsOrSelf(wnd.CurrentFocus(), e.content) {
		wnd.SetFocus(nil)
	}
}

// newContentFocus returns a record of where focus is, which may be within content or outside of it. The record has no
// scroll position.
func newContentFocus(content, focus *unison.Panel) *contentFocus {
	cf := &contentFocus{panel: focus}
	switch p := focus.Self.(type) {
	case statefulField:
		cf.field = p.GetFieldState()
	case selectableTable:
		cf.selection = p.CopySelectionMap()
		for i := range p.LastRowIndex() + 1 {
			if p.IsRowSelected(i) {
				cf.rows = append(cf.rows, i)
			}
		}
	}
	if unison.AncestorIsOrSelf(focus, content) {
		for p := focus; p != content; p = p.Parent() {
			cf.path = append(cf.path, focusStep{
				kind:  reflect.TypeOf(p.Self),
				label: panelLabel(p),
				index: p.Parent().IndexOfChild(p),
			})
		}
		slices.Reverse(cf.path)
		cf.origin = originWithin(focus, content)
	}
	return cf
}

// originWithin returns where the top left corner of p is in the coordinates of ancestor. Unlike Panel.PointTo, it
// doesn't go by way of the root, so the result is the same however ancestor has been scrolled, to the last bit.
func originWithin(p, ancestor *unison.Panel) geom.Point {
	var pt geom.Point
	for ; p != ancestor; p = p.Parent() {
		pt = pt.MulPt(p.Scale()).Add(p.FrameRect().Point)
	}
	return pt
}

// gone returns true if what held the focus is no longer in a window, as is so once the dockable it was in has closed.
func (cf *contentFocus) gone() bool {
	return cf != nil && cf.panel.Window() == nil
}

// panelLabel returns the text of the label that names the panel, or "" if it has none: the label the panel points at
// or, failing that, the one just before it among its parent's children.
func panelLabel(p *unison.Panel) string {
	labeler := p.Accessibility.LabeledBy
	if xreflect.IsNil(labeler) {
		parent := p.Parent()
		if parent == nil {
			return ""
		}
		i := parent.IndexOfChild(p)
		if i < 1 {
			return ""
		}
		labeler = parent.Children()[i-1]
	}
	if label, ok := labeler.AsPanel().Self.(*unison.Label); ok {
		return label.String()
	}
	return ""
}

// childOf returns the child of parent the step leads to, or nil if there is none: the child at the same index if it has
// the same kind and label, or else, since controls ahead of it may have come or gone, the only child with that kind and
// label. A child without a label can't be told from its like, so it isn't looked for elsewhere.
func (s focusStep) childOf(parent *unison.Panel) *unison.Panel {
	children := parent.Children()
	// A panel that isn't among its parent's children, as the cell a table hands the focus to isn't, has an index of -1.
	if s.index >= 0 && s.index < len(children) && s.leadsTo(children[s.index]) {
		return children[s.index]
	}
	if s.label == "" {
		return nil
	}
	var found *unison.Panel
	for _, child := range children {
		if s.leadsTo(child) {
			if found != nil {
				return nil
			}
			found = child
		}
	}
	return found
}

func (s focusStep) leadsTo(p *unison.Panel) bool {
	return reflect.TypeOf(p.Self) == s.kind && panelLabel(p) == s.label
}

// find retraces the path within content, which may have been rebuilt since, returning the panel it leads to and true.
// If a step leads nowhere, or that panel can't take the focus, it returns false with the nearest panel that can,
// looking around where the path broke off and then back up it, or nil if there is none.
func (cf *contentFocus) find(content *unison.Panel) (target *unison.Panel, same bool) {
	// trail[i] is the panel step i sets out from.
	trail := make([]*unison.Panel, 0, len(cf.path))
	target = content
	for _, step := range cf.path {
		child := step.childOf(target)
		if child == nil {
			break
		}
		trail = append(trail, target)
		target = child
	}
	if len(trail) == len(cf.path) {
		if takesFocus(target) {
			return target, true
		}
	} else {
		trail = append(trail, target)
	}
	for i, from := range slices.Backward(trail) {
		if nearest := nearestTakingFocus(from.Children(), cf.path[i].index); nearest != nil {
			return nearest, false
		}
	}
	return nil, false
}

func takesFocus(p *unison.Panel) bool {
	return p.Focusable() || p.FirstFocusableChild() != nil
}

// nearestTakingFocus returns the panel nearest to index that can take the focus, the earlier on a tie, or nil if none
// can.
func nearestTakingFocus(panels []*unison.Panel, index int) *unison.Panel {
	var nearest *unison.Panel
	best := math.MaxInt
	for i, p := range panels {
		if distance := max(i-index, index-i); distance < best && takesFocus(p) {
			nearest = p
			best = distance
		}
	}
	return nearest
}

// restore puts the focus back where it was, along with the editor's scrolling, both of which closing a sub-editor may
// have changed. If the focus was within content, which may have been rebuilt since, the editor is instead scrolled by
// as much as what held the focus has moved, which leaves it where it was within the view. A field gets its selection
// back only if its text is unchanged; otherwise it is left with all of its text selected. A table that has been rebuilt
// gets its selected rows back (see restoreSelection). If what held the focus is gone, the nearest panel that can take
// it (see find) gets it and is scrolled into view. Does nothing if cf is nil.
func (cf *contentFocus) restore(content *unison.Panel, scroll *unison.ScrollPanel) {
	if cf == nil {
		return
	}
	target := cf.panel
	same := true
	if len(cf.path) != 0 && !unison.AncestorIsOrSelf(target, content) {
		target, same = cf.find(content)
	}
	if target != nil {
		target.RequestFocus()
	}
	h, v := cf.h, cf.v
	if same {
		switch p := target.Self.(type) {
		case statefulField:
			if cf.field != nil && p.Text() == cf.field.Text {
				p.ApplyFieldState(cf.field)
			}
		case selectableTable:
			if target != cf.panel {
				cf.restoreSelection(p)
			}
		}
		if len(cf.path) != 0 {
			moved := originWithin(target, content).Sub(cf.origin).MulPt(content.Scale())
			h += moved.X
			v += moved.Y
		}
	}
	scroll.SetPosition(h, v)
	if !same && target != nil {
		// Not every panel brings itself into view on taking the focus.
		if focus := content.Window().CurrentFocus(); focus != nil {
			focus.ScrollIntoView()
		}
	}
}

// restoreSelection selects the rows of a rebuilt table that were selected in the table it replaces: those with the same
// IDs or, when it has none of them, as when a sync has replaced the weapons with the library's, those in the same
// places.
func (cf *contentFocus) restoreSelection(table selectableTable) {
	if len(cf.selection) == 0 {
		return
	}
	table.SetSelectionMap(cf.selection)
	if table.SelectionCount() == 0 {
		table.SelectByIndex(cf.rows...)
	}
}

func (e *editor[N, D]) clearSource() {
	e.undoMgr.Add(&unison.UndoEdit[bool]{
		ID:         unison.NextUndoID(),
		EditName:   i18n.Text("Clear Source"),
		UndoFunc:   func(edit *unison.UndoEdit[bool]) { e.setSourceCleared(edit.BeforeData) },
		RedoFunc:   func(edit *unison.UndoEdit[bool]) { e.setSourceCleared(edit.AfterData) },
		BeforeData: e.sourceCleared,
		AfterData:  true,
	})
	e.setSourceCleared(true)
}

func (e *editor[N, D]) setSourceCleared(cleared bool) {
	e.sourceCleared = cleared
	e.MarkModified(nil)
}
