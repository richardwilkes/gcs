// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
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
	nameablesButton      *unison.Button
	beforeData           D
	editorData           D
	modificationCallback func()
	preApplyCallback     func(D)
	scale                int
	promptForSave        bool
}

func displayEditor[N gurps.NodeTypes, D gurps.EditorData[N]](owner Rebuildable, target N, icon *unison.SVG, helpMD string, initToolbar func(*editor[N, D], *unison.Panel), initContent func(*editor[N, D], *unison.Panel) func(), preApplyCallback func(D)) {
	lookFor := gurps.AsNode(target).ID()
	if Activate(func(d unison.Dockable) bool {
		if e, ok := d.AsPanel().Self.(*editor[N, D]); ok {
			return e.owner == owner && gurps.AsNode(e.target).ID() == lookFor
		}
		return false
	}) {
		return
	}
	e := &editor[N, D]{
		owner:            owner,
		target:           target,
		svg:              icon,
		scale:            gurps.GlobalSettings().General.InitialEditorUIScale,
		preApplyCallback: preApplyCallback,
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

	e.undoMgr = unison.NewUndoManager(100, func(err error) { errs.Log(err) })
	e.SetLayout(&unison.FlexLayout{Columns: 1})

	content := unison.NewPanel()
	content.SetBorder(unison.NewEmptyBorder(unison.NewUniformInsets(unison.StdHSpacing * 2)))
	content.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	content.KeyDownCallback = func(keyCode unison.KeyCode, mod unison.Modifiers, _ bool) bool {
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
	e.scroll.SetContent(content, behavior.HintedFill, behavior.Fill)
	e.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})

	e.AddChild(e.createToolbar(helpMD, initToolbar))
	e.modificationCallback = initContent(e, content)
	e.AddChild(e.scroll)
	e.ClientData()[AssociatedIDKey] = gurps.AsNode(target).ID()
	e.promptForSave = true
	e.scroll.Content().AsPanel().ValidateScrollRoot()
	group := dgroup.Editors
	p := owner.AsPanel()
	for p != nil {
		if _, exists := p.ClientData()[AssociatedIDKey]; exists {
			group = dgroup.SubEditors
			break
		}
		p = p.Parent()
	}
	PlaceInDock(e, group, false)
	content.RequestFocus()
}

func (e *editor[N, D]) createToolbar(helpMD string, initToolbar func(*editor[N, D], *unison.Panel)) unison.Paneler {
	toolbar := unison.NewPanel()
	toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, 0, unison.Insets{Bottom: 1},
		false), unison.NewEmptyBorder(unison.StdInsets())))

	toolbar.AddChild(NewDefaultInfoPop())

	if helpMD != "" {
		helpButton := unison.NewSVGButton(svg.Help)
		helpButton.Tooltip = newWrappedTooltip(i18n.Text("Help"))
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
	e.applyButton.Tooltip = newWrappedTooltipWithSecondaryText(i18n.Text("Apply Changes"),
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
	e.cancelButton.Tooltip = newWrappedTooltipWithSecondaryText(i18n.Text("Discard Changes"), unison.KeyEscape.String())
	e.cancelButton.SetEnabled(false)
	e.cancelButton.ClickCallback = func() {
		e.promptForSave = false
		e.AttemptClose()
	}
	toolbar.AddChild(e.cancelButton)

	target := any(e.target)
	if _, ok := target.(*gurps.Weapon); !ok {
		if _, ok = target.(*gurps.EquipmentModifier); !ok {
			if _, ok = target.(*gurps.TraitModifier); !ok {
				e.nameablesButton = unison.NewSVGButton(svg.Naming)
				e.nameablesButton.Tooltip = newWrappedTooltip(i18n.Text("Set Substitutions"))
				e.nameablesButton.ClickCallback = func() {
					if tmp, m := e.prepareForSubstitutions(); len(m) > 0 {
						ShowNameablesDialog([]string{tmp.String()}, []map[string]string{m})
						tmp.ApplyNameableKeys(m)
						e.editorData.CopyFrom(tmp)
						e.Rebuild(false)
					}
				}
				toolbar.AddChild(e.nameablesButton)
			}
		}
	}

	if initToolbar != nil {
		initToolbar(e, toolbar)
	}

	toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	toolbar.SetLayout(&unison.FlexLayout{
		Columns:  len(toolbar.Children()),
		HSpacing: unison.StdHSpacing,
	})
	return toolbar
}

func (e *editor[N, D]) prepareForSubstitutions() (tmpNode N, m map[string]string) {
	node := gurps.AsNode(e.target)
	tmpNode = node.Clone(node.GetSource().LibraryFile, node.DataOwner(), nil, true)
	e.editorData.ApplyTo(tmpNode)
	m = make(map[string]string)
	tmpNode.FillWithNameableKeys(m, nil)
	return tmpNode, m
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

var pruneIDFields = regexp.MustCompile(`\s*"id":\s*"[^"]+",?\s*`)

func (e *editor[N, D]) isModified() bool {
	d1, err := json.Marshal(e.beforeData)
	if err != nil {
		errs.Log(errs.Wrap(err))
		return false
	}
	var d2 []byte
	d2, err = json.Marshal(e.editorData)
	if err != nil {
		errs.Log(errs.Wrap(err))
		return false
	}
	none := []byte{}
	d1 = pruneIDFields.ReplaceAll(d1, none)
	d2 = pruneIDFields.ReplaceAll(d2, none)
	return !bytes.Equal(d1, d2)
}

func (e *editor[N, D]) Modified() bool {
	modified := e.isModified()
	e.applyButton.SetEnabled(modified)
	e.cancelButton.SetEnabled(modified)
	if e.nameablesButton != nil {
		_, m := e.prepareForSubstitutions()
		e.nameablesButton.SetEnabled(len(m) > 0)
	}
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
	gurps.DiscardGlobalResolveCache()
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
	if e.promptForSave && e.isModified() {
		switch unison.YesNoCancelDialog(fmt.Sprintf(i18n.Text("Save changes made to\n%s?"), e.Title()), "") {
		case unison.ModalResponseDiscard:
		case unison.ModalResponseOK:
			e.apply()
		default:
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
	if e.preApplyCallback != nil {
		e.preApplyCallback(e.editorData)
	}
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
