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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/mod"
)

// editorDockable is what editorShell needs of the editor embedding it.
type editorDockable interface {
	unison.Dockable
	unison.TabCloser
}

// editorShell is the dockable shell shared by editor and pointsEditor. Their title, their data, how it is applied and
// whether it has changed stay with the embedding type, which the shell reaches through Self.
type editorShell struct {
	unison.Panel
	owner            Rebuildable
	icon             *unison.SVG
	previousDockable unison.Dockable
	previousFocusKey string
	undoMgr          *unison.UndoManager
	scroll           *unison.ScrollPanel
	applyButton      *unison.Button
	cancelButton     *unison.Button
	// promptForSave is cleared by the Apply and Discard buttons, which have already settled what happens to pending
	// changes, so closing then asks nothing.
	promptForSave bool
	// discardReason, when not empty, says why saving the pending changes would be futile, as it is while what the editor
	// edits is about to be replaced. Closing then asks whether to discard them rather than whether to save them.
	discardReason string
}

// editor returns the editor embedding the shell, or nil if Self has not been set to one.
func (s *editorShell) editor() editorDockable {
	if d, ok := s.Self.(editorDockable); ok {
		return d
	}
	return nil
}

// setUp readies the shell once Self has been set: it records where to return to when the editor closes, creates the
// undo manager, and returns the content panel, laid out in the given number of columns. The editor then adds its
// toolbar and the scroll panel itself, since the toolbar may need the scroll panel.
func (s *editorShell) setUp(columns int) *unison.Panel {
	if defDC := DefaultDockContainer(); defDC != nil {
		if s.previousDockable = defDC.CurrentDockable(); !xreflect.IsNil(s.previousDockable) {
			if focus := s.previousDockable.AsPanel().Window().Focus(); focus != nil {
				if focus.Ancestor[unison.Dockable]() == s.previousDockable {
					s.previousFocusKey = focus.RefKey
				}
			}
		}
	}
	s.undoMgr = unison.NewUndoManager(100, func(err error) { errs.Log(err) })
	s.SetLayout(&unison.FlexLayout{Columns: 1})
	return s.newContentPanel(columns)
}

// newContentPanel creates the editor's content panel, laid out in the given number of columns, and the scroll panel
// that holds it. Cmd-Return within the content applies the changes and Escape discards them, by clicking the buttons
// addApplyAndCancelButtons creates.
func (s *editorShell) newContentPanel(columns int) *unison.Panel {
	content := unison.NewPanel()
	content.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(unison.StdHSpacing * 2)))
	content.SetLayout(&unison.FlexLayout{
		Columns:  columns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	content.KeyDownCallback = func(keyCode unison.KeyCode, mods mod.Modifiers, _ bool) bool {
		switch {
		case mods.OSMenuCommandDown() && (keyCode == unison.KeyReturn || keyCode == unison.KeyNumPadEnter):
			if s.applyButton.Enabled() {
				s.applyButton.Click()
			}
			return true
		case noModifiersDown(mods) && keyCode == unison.KeyEscape:
			if s.cancelButton.Enabled() {
				s.cancelButton.Click()
			}
			return true
		default:
			return false
		}
	}
	s.scroll = unison.NewScrollPanel()
	s.scroll.SetContent(content, behavior.HintedFill, behavior.Fill)
	s.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	return content
}

// addApplyAndCancelButtons adds the Apply and Discard buttons to the toolbar. Both close the editor without prompting,
// since the user has just said what to do with the changes; Apply applies them first.
func (s *editorShell) addApplyAndCancelButtons(toolbar *unison.Panel, apply func()) {
	s.applyButton, s.cancelButton = newApplyCancelButtons(toolbar, true,
		func() bool {
			apply()
			return true
		},
		s.discardAndClose)
}

// discardAndClose closes the editor, dropping any pending changes without asking about them.
func (s *editorShell) discardAndClose() {
	s.promptForSave = false
	s.editor().AttemptClose()
}

// discardEditorsFor closes the editors open on any of the given IDs without asking about their pending changes. It is
// for when an editor's item has changed in a way that makes the editor's data wrong, where CloseID's offer to save the
// changes would do harm.
func discardEditorsFor(ids map[tid.TID]bool) {
	for _, d := range AllDockables() {
		if closer, ok := d.(interface{ discardAndClose() }); ok {
			if id, ok2 := d.AsPanel().ClientData()[AssociatedIDKey].(tid.TID); ok2 && ids[id] {
				closer.discardAndClose()
			}
		}
	}
}

// placeInDock records the ID of what the editor edits, so that closing that item's dockable closes the editor too (see
// CloseID), and places the editor in the dock, grouped with the other editors, or with the sub-editors when its owner
// is itself within an editor. Closing the editor prompts to save from here on.
func (s *editorShell) placeInDock(id tid.TID) {
	s.ClientData()[AssociatedIDKey] = id
	s.promptForSave = true
	s.scroll.Content().AsPanel().ValidateScrollRoot()
	group := dgroup.Editors
	if !xreflect.IsNil(s.owner) {
		for p := s.owner.AsPanel(); p != nil; p = p.Parent() {
			if _, exists := p.ClientData()[AssociatedIDKey]; exists {
				group = dgroup.SubEditors
				break
			}
		}
	}
	PlaceInDock(s.editor(), group, false)
}

// enableApplyAndCancel enables the Apply and Discard buttons when there are changes to apply and disables them when
// there are none, returning what it was given so that a Modified method can end with it.
func (s *editorShell) enableApplyAndCancel(modified bool) bool {
	s.applyButton.SetEnabled(modified)
	s.cancelButton.SetEnabled(modified)
	return modified
}

// confirmClose asks whether to save pending changes, when isModified reports some and neither button has already
// settled what to do with them, applying them with apply if the user says to. When saving them would be futile (see
// discardReason), it asks whether to discard them instead. It returns false if the user cancels, in which case the
// editor stays open.
func (s *editorShell) confirmClose(isModified func() bool, apply func()) bool {
	if !s.promptForSave || !isModified() {
		return true
	}
	if s.discardReason != "" {
		return unison.QuestionDialog(i18n.Text("Discard changes made to\n%s?", s.editor().Title()),
			s.discardReason) == unison.ModalResponseOK
	}
	switch unison.YesNoCancelDialog(i18n.Text("Save changes made to\n%s?", s.editor().Title()), "") {
	case unison.ModalResponseDiscard:
		return true
	case unison.ModalResponseOK:
		apply()
		return true
	default:
		return false
	}
}

// returnToPrevious makes the dockable that was current when the editor opened current again and gives the keyboard
// focus back to the panel within it that held it then, returning that panel or nil when there is no such dockable or
// panel.
func (s *editorShell) returnToPrevious() *unison.Panel {
	if xreflect.IsNil(s.previousDockable) {
		return nil
	}
	dc := unison.Ancestor[*unison.DockContainer](s.previousDockable)
	if dc == nil {
		return nil
	}
	dc.SetCurrentDockable(s.previousDockable)
	if s.previousFocusKey == "" {
		return nil
	}
	p := s.previousDockable.AsPanel().FindRefKey(s.previousFocusKey)
	if p != nil {
		restoreFocus(p)
	}
	return p
}

// focusWithoutScroller is implemented by panels, such as tables, that can take the keyboard focus without performing
// their default focus-gained scrolling.
type focusWithoutScroller interface {
	RequestFocusWithoutScroll()
}

// restoreFocus gives the keyboard focus back to a panel that held it before an editor was opened. Tables are focused
// without their default scrolling, which brings the entire table into view even when the user's row is still visible;
// callers that know the row can follow up with revealRowForData.
func restoreFocus(p *unison.Panel) {
	if f, ok := p.Self.(focusWithoutScroller); ok {
		f.RequestFocusWithoutScroll()
		return
	}
	p.RequestFocus()
}

func (s *editorShell) TitleIcon(suggestedSize geom.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  s.icon,
		Size: suggestedSize,
	}
}

func (s *editorShell) String() string {
	return s.editor().Title()
}

func (s *editorShell) Tooltip() string {
	return ""
}

func (s *editorShell) UndoManager() *unison.UndoManager {
	return s.undoMgr
}

func (s *editorShell) CloseWithGroup(other unison.Paneler) bool {
	return s.owner != nil && s.owner == other
}

func (s *editorShell) MayAttemptClose() bool {
	return MayAttemptCloseOfGroup(s.editor())
}
