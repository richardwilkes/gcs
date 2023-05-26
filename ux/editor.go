/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
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
	"reflect"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/unison"
)

var (
	_ unison.Dockable            = &editor[*gurps.Note, *gurps.NoteEditData]{}
	_ unison.TabCloser           = &editor[*gurps.Note, *gurps.NoteEditData]{}
	_ ModifiableRoot             = &editor[*gurps.Note, *gurps.NoteEditData]{}
	_ unison.UndoManagerProvider = &editor[*gurps.Note, *gurps.NoteEditData]{}
	_ GroupedCloser              = &editor[*gurps.Note, *gurps.NoteEditData]{}
	_ Rebuildable                = &editor[*gurps.Note, *gurps.NoteEditData]{}
)

type editor[N gurps.NodeTypes, D gurps.EditorData[N]] struct {
	unison.Panel
	owner                Rebuildable
	target               N
	previousDockable     unison.Dockable
	previousFocusKey     string
	svg                  *unison.SVG
	undoMgr              *unison.UndoManager
	scroll               *unison.ScrollPanel
	applyButton          *unison.Button
	cancelButton         *unison.Button
	beforeData           D
	editorData           D
	modificationCallback func()
	scale                int
	promptForSave        bool
}

func displayEditor[N gurps.NodeTypes, D gurps.EditorData[N]](owner Rebuildable, target N, svg *unison.SVG, helpMD string, initToolbar func(*editor[N, D], *unison.Panel), initContent func(*editor[N, D], *unison.Panel) func()) {
	lookFor := gurps.AsNode(target).UUID()
	if Activate(func(d unison.Dockable) bool {
		if e, ok := d.AsPanel().Self.(*editor[N, D]); ok {
			return e.owner == owner && gurps.AsNode(e.target).UUID() == lookFor
		}
		return false
	}) {
		return
	}
	e := &editor[N, D]{
		owner:  owner,
		target: target,
		svg:    svg,
		scale:  gurps.GlobalSettings().General.InitialEditorUIScale,
	}
	e.Self = e

	if defDC := DefaultDockContainer(); defDC != nil {
		if e.previousDockable = defDC.CurrentDockable(); !toolbox.IsNil(e.previousDockable) {
			if focus := e.previousDockable.AsPanel().Window().Focus(); focus != nil {
				if unison.Ancestor[unison.Dockable](focus) == e.previousDockable {
					e.previousFocusKey = focus.RefKey
				}
			}
		}
	}

	reflect.ValueOf(&e.beforeData).Elem().Set(reflect.New(reflect.TypeOf(e.beforeData).Elem()))
	e.beforeData.CopyFrom(target)

	reflect.ValueOf(&e.editorData).Elem().Set(reflect.New(reflect.TypeOf(e.editorData).Elem()))
	e.editorData.CopyFrom(target)

	e.undoMgr = unison.NewUndoManager(100, func(err error) { jot.Error(err) })
	e.SetLayout(&unison.FlexLayout{Columns: 1})

	content := unison.NewPanel()
	content.SetBorder(unison.NewEmptyBorder(unison.NewUniformInsets(unison.StdHSpacing * 2)))
	content.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	content.KeyDownCallback = func(keyCode unison.KeyCode, mod unison.Modifiers, repeat bool) bool {
		switch {
		case mod.OSMenuCmdModifierDown() && (keyCode == unison.KeyReturn || keyCode == unison.KeyNumPadEnter):
			if e.applyButton.Enabled() {
				e.applyButton.Click()
			}
			return true
		case mod == 0 && keyCode == unison.KeyEscape:
			if e.cancelButton.Enabled() {
				e.cancelButton.Click()
			}
			return true
		default:
			return false
		}
	}

	e.scroll = unison.NewScrollPanel()
	e.scroll.SetContent(content, unison.HintedFillBehavior, unison.FillBehavior)
	e.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
		VGrab:  true,
	})

	e.AddChild(e.createToolbar(helpMD, initToolbar))
	e.modificationCallback = initContent(e, content)
	e.AddChild(e.scroll)
	e.ClientData()[AssociatedUUIDKey] = gurps.AsNode(target).UUID()
	e.promptForSave = true
	e.scroll.Content().AsPanel().ValidateScrollRoot()
	group := gurps.EditorsDockableGroup
	p := owner.AsPanel()
	for p != nil {
		if _, exists := p.ClientData()[AssociatedUUIDKey]; exists {
			group = gurps.SubEditorsDockableGroup
			break
		}
		p = p.Parent()
	}
	PlaceInDock(e, group, false)
	content.RequestFocus()
}

func (e *editor[N, D]) createToolbar(helpMD string, initToolbar func(*editor[N, D], *unison.Panel)) unison.Paneler {
	toolbar := unison.NewPanel()
	toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.Insets{Bottom: 1},
		false), unison.NewEmptyBorder(unison.StdInsets())))

	toolbar.AddChild(NewDefaultInfoPop())

	if helpMD != "" {
		helpButton := unison.NewSVGButton(svg.Help)
		helpButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Help"))
		helpButton.ClickCallback = func() { HandleLink(nil, helpMD) }
		toolbar.AddChild(helpButton)
	}

	toolbar.AddChild(
		NewScaleField(
			gurps.InitialUIScaleMin,
			gurps.InitialUIScaleMax,
			func() int { return gurps.GlobalSettings().General.InitialEditorUIScale },
			func() int { return e.scale },
			func(scale int) { e.scale = scale },
			nil,
			false,
			e.scroll,
		),
	)

	e.applyButton = unison.NewSVGButton(unison.CheckmarkSVG)
	e.applyButton.Tooltip = unison.NewTooltipWithSecondaryText(i18n.Text("Apply Changes"),
		fmt.Sprintf(i18n.Text("%v%v or %v%v"), unison.OSMenuCmdModifier(), unison.KeyReturn, unison.OSMenuCmdModifier(),
			unison.KeyNumPadEnter))
	e.applyButton.SetEnabled(false)
	e.applyButton.ClickCallback = func() {
		e.apply()
		e.promptForSave = false
		e.AttemptClose()
	}
	toolbar.AddChild(e.applyButton)

	e.cancelButton = unison.NewSVGButton(svg.Not)
	e.cancelButton.Tooltip = unison.NewTooltipWithSecondaryText(i18n.Text("Discard Changes"), unison.KeyEscape.String())
	e.cancelButton.SetEnabled(false)
	e.cancelButton.ClickCallback = func() {
		e.promptForSave = false
		e.AttemptClose()
	}
	toolbar.AddChild(e.cancelButton)

	if initToolbar != nil {
		initToolbar(e, toolbar)
	}

	toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	toolbar.SetLayout(&unison.FlexLayout{
		Columns:  len(toolbar.Children()),
		HSpacing: unison.StdHSpacing,
	})
	return toolbar
}

func (e *editor[N, D]) TitleIcon(suggestedSize unison.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  e.svg,
		Size: suggestedSize,
	}
}

func (e *editor[N, D]) Title() string {
	return fmt.Sprintf(i18n.Text("%s Editor for %s"), gurps.AsNode(e.target).Kind(), e.owner.String())
}

func (e *editor[N, D]) String() string {
	return e.Title()
}

func (e *editor[N, D]) Tooltip() string {
	return ""
}

func (e *editor[N, D]) Modified() bool {
	modified := !reflect.DeepEqual(e.beforeData, e.editorData)
	e.applyButton.SetEnabled(modified)
	e.cancelButton.SetEnabled(modified)
	return modified
}

func (e *editor[N, D]) MarkModified(_ unison.Paneler) {
	UpdateTitleForDockable(e)
	DeepSync(e)
	if e.modificationCallback != nil {
		e.modificationCallback()
	}
}

func (e *editor[N, D]) Rebuild(_ bool) {
	e.MarkModified(nil)
	e.MarkForLayoutRecursively()
	e.MarkForRedraw()
}

func (e *editor[N, D]) CloseWithGroup(other unison.Paneler) bool {
	return e.owner != nil && e.owner == other
}

func (e *editor[N, D]) MayAttemptClose() bool {
	return MayAttemptCloseOfGroup(e)
}

func (e *editor[N, D]) AttemptClose() bool {
	if !CloseGroup(e) {
		return false
	}
	if e.promptForSave && !reflect.DeepEqual(e.beforeData, e.editorData) {
		switch unison.YesNoCancelDialog(fmt.Sprintf(i18n.Text("Save changes made to\n%s?"), e.Title()), "") {
		case unison.ModalResponseDiscard:
		case unison.ModalResponseOK:
			e.apply()
		case unison.ModalResponseCancel:
			return false
		}
	}
	if !toolbox.IsNil(e.previousDockable) {
		if pdc := unison.Ancestor[*unison.DockContainer](e.previousDockable); pdc != nil {
			pdc.SetCurrentDockable(e.previousDockable)
			if e.previousFocusKey != "" {
				if p := e.previousDockable.AsPanel().FindRefKey(e.previousFocusKey); p != nil {
					p.RequestFocus()
				}
			}
		}
	}
	return AttemptCloseForDockable(e)
}

func (e *editor[N, D]) UndoManager() *unison.UndoManager {
	return e.undoMgr
}

func (e *editor[N, D]) apply() {
	e.Window().FocusNext() // Intentionally move the focus to ensure any pending edits are flushed
	if mgr := unison.UndoManagerFor(e.owner); mgr != nil {
		owner := e.owner
		target := e.target
		mgr.Add(&unison.UndoEdit[D]{
			ID:       unison.NextUndoID(),
			EditName: fmt.Sprintf(i18n.Text("%s Changes"), gurps.AsNode(target).Kind()),
			UndoFunc: func(edit *unison.UndoEdit[D]) {
				edit.BeforeData.ApplyTo(target)
				owner.Rebuild(true)
			},
			RedoFunc: func(edit *unison.UndoEdit[D]) {
				edit.AfterData.ApplyTo(target)
				owner.Rebuild(true)
			},
			BeforeData: e.beforeData,
			AfterData:  e.editorData,
		})
	}
	e.editorData.ApplyTo(e.target)
	e.owner.Rebuild(true)
}
