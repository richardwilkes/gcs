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
	"math"
	"reflect"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/i18n"
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
		Label: fmt.Sprintf(i18n.Text("%s: %s"), label, value),
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
// data's modifiers and weapons are closed first, since the sync replaces what they edit.
func (e *editor[N, D]) syncWithSource() {
	// Noted before the sub-editors are closed, since closing one gives the focus to the list it was opened from.
	focus := e.noteFocus()
	if CloseGroup(e) {
		// Released before the sync, so that what a field does on losing the focus is done while it is still part of the
		// editor and before the data is replaced.
		e.releaseContentFocus()
		edited := e.editedClone()
		if state, data := gurps.MatchSource(edited); state == srcstate.Mismatched && !syncChangesKind(edited, data) {
			edited.SyncWithSource()
			e.editorData.CopyFrom(edited)
			choice := modifierChoicePicker(edited)
			e.syncedChoice = &choice
			e.rebuildContent()
		}
	}
	focus.restore(e.content, e.scroll)
}

// contentFocus records where the keyboard focus was, so that it can be put back once the editor's content has been
// rebuilt, in which a control may have moved among its siblings, or gone.
type contentFocus struct {
	panel *unison.Panel
	// path leads from the content down to panel, and is empty for a panel outside the content.
	path     []focusStep
	text     string
	selStart int
	selEnd   int
	isField  bool
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

// selectableText is what a field offers that restoring the focus to it needs.
type selectableText interface {
	Text() string
	Selection() (start, end int)
	SetSelection(start, end int)
}

// noteFocus returns a record of where the keyboard focus is, or nil if nothing holds it.
func (e *editor[N, D]) noteFocus() *contentFocus {
	focus := e.Window().CurrentFocus()
	switch {
	case focus == nil:
		return nil
	case unison.AncestorIsOrSelf(focus, e.content):
		return newContentFocus(e.content, focus)
	default:
		return &contentFocus{panel: focus}
	}
}

func (e *editor[N, D]) releaseContentFocus() {
	if wnd := e.Window(); unison.AncestorIsOrSelf(wnd.CurrentFocus(), e.content) {
		wnd.SetFocus(nil)
	}
}

// newContentFocus returns a record of where focus, a panel within content, is.
func newContentFocus(content, focus *unison.Panel) *contentFocus {
	cf := &contentFocus{panel: focus}
	for p := focus; p != content; p = p.Parent() {
		cf.path = append(cf.path, focusStep{
			kind:  reflect.TypeOf(p.Self),
			label: panelLabel(p),
			index: p.Parent().IndexOfChild(p),
		})
	}
	slices.Reverse(cf.path)
	if field, ok := focus.Self.(selectableText); ok {
		cf.isField = true
		cf.text = field.Text()
		cf.selStart, cf.selEnd = field.Selection()
	}
	return cf
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
	if s.index < len(children) && s.leadsTo(children[s.index]) {
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

// restore puts the focus back where it was, without scrolling if that was within content, which may have been rebuilt
// since. A field gets its selection back only if its text is unchanged; otherwise it is left with all of its text
// selected. If what held the focus is gone, the nearest panel that can take it (see find) gets it and is scrolled into
// view. Does nothing if cf is nil.
func (cf *contentFocus) restore(content *unison.Panel, scroll *unison.ScrollPanel) {
	if cf == nil {
		return
	}
	target := cf.panel
	same := true
	if !unison.AncestorIsOrSelf(target, content) {
		if len(cf.path) == 0 {
			target.RequestFocus()
			return
		}
		if target, same = cf.find(content); target == nil {
			return
		}
	}
	h, v := scroll.Position()
	target.RequestFocus()
	if !same {
		// Not every panel brings itself into view on taking the focus.
		if focus := content.Window().CurrentFocus(); focus != nil {
			focus.ScrollIntoView()
		}
		return
	}
	if field, ok := target.Self.(selectableText); ok && cf.isField && field.Text() == cf.text {
		field.SetSelection(cf.selStart, cf.selEnd)
	}
	scroll.SetPosition(h, v)
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
