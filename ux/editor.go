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
	"bytes"
	"fmt"
	"reflect"
	"regexp"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
)

var (
	_ unison.Dockable            = &editor[*gurps.Note, *gurps.NoteEditData]{}
	_ unison.TabCloser           = &editor[*gurps.Note, *gurps.NoteEditData]{}
	_ ModifiableRoot             = &editor[*gurps.Note, *gurps.NoteEditData]{}
	_ unison.UndoManagerProvider = &editor[*gurps.Note, *gurps.NoteEditData]{}
	_ GroupedCloser              = &editor[*gurps.Note, *gurps.NoteEditData]{}
	_ Rebuildable                = &editor[*gurps.Note, *gurps.NoteEditData]{}
	_ Owned                      = &editor[*gurps.Note, *gurps.NoteEditData]{}
)

type editor[N gurps.Node[N], D gurps.EditorData[N]] struct {
	editorShell
	target               N
	nameablesButton      *unison.Button
	meleeWeapons         *weaponsPanel
	rangedWeapons        *weaponsPanel
	beforeData           D
	editorData           D
	nameablesScratch     N
	nameablesKeys        map[string]string
	modificationCallback func()
	preApplyCallback     func(D)
	scale                int
}

func displayEditor[N gurps.Node[N], D gurps.EditorData[N]](owner Rebuildable, target N, icon *unison.SVG, helpMD string, initToolbar func(*editor[N, D], *unison.Panel), initContent func(*editor[N, D], *unison.Panel) func(), preApplyCallback func(D)) *editor[N, D] {
	var found *editor[N, D]
	lookFor := target.ID()
	if Activate(func(d unison.Dockable) bool {
		if e, ok := d.AsPanel().Self.(*editor[N, D]); ok {
			if e.owner == owner && e.target.ID() == lookFor {
				found = e
				return true
			}
		}
		return false
	}) {
		return found
	}
	e := &editor[N, D]{
		owner:            owner,
		icon:             icon,
		target:           target,
		scale:            gurps.GlobalSettings().General.InitialEditorUIScale,
		preApplyCallback: preApplyCallback,
	}
	e.Self = e

	reflect.ValueOf(&e.beforeData).Elem().Set(reflect.New(reflect.TypeOf(e.beforeData).Elem()))
	e.beforeData.CopyFrom(target)

	reflect.ValueOf(&e.editorData).Elem().Set(reflect.New(reflect.TypeOf(e.editorData).Elem()))
	e.editorData.CopyFrom(target)

	content := e.setUp(2)
	e.AddChild(e.createToolbar(helpMD, initToolbar))
	e.AddChild(e.scroll)
	e.modificationCallback = initContent(e, content)
	e.placeInDock(target.ID())
	content.RequestFocus()
	return e
}

func (e *editor[N, D]) createToolbar(helpMD string, initToolbar func(*editor[N, D], *unison.Panel)) unison.Paneler {
	toolbar := newToolbar()
	toolbar.AddChild(NewDefaultInfoPop())

	if helpMD != "" {
		addHelpButton(toolbar, helpMD)
	}

	addUIScaleField(toolbar, func() int { return gurps.GlobalSettings().General.InitialEditorUIScale },
		func() int { return e.scale }, func(scale int) { e.scale = scale }, false, e.scroll)

	e.addApplyAndCancelButtons(toolbar, e.apply)

	target := any(e.target)
	if _, ok := target.(*gurps.Weapon); !ok {
		if _, ok = target.(*gurps.EquipmentModifier); !ok {
			if _, ok = target.(*gurps.TraitModifier); !ok {
				e.nameablesButton = unison.NewSVGButton(svg.Naming)
				e.nameablesButton.Tooltip = newWrappedTooltip(i18n.Text("Set Substitutions"))
				e.nameablesButton.ClickCallback = func() {
					if tmp, m := e.prepareForSubstitutions(); len(m) > 0 {
						ShowNameablesDialog([]string{tmp.String()}, []map[string]string{m}, [][]string{nil})
						tmp.ApplyNameableKeys(m)
						// Applying nameable keys only alters the replacements map, so copy just that back. CopyFrom
						// would replace the entire object graph with fresh sub-objects, orphaning the widgets already
						// bound to the existing ones (e.g. the modifiers table) and silently losing their later edits.
						if setter, ok2 := any(e.editorData).(nameable.Setter); ok2 {
							setter.SetNameableReplacements(tmp.NameableReplacements())
						} else {
							e.editorData.CopyFrom(tmp)
						}
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

	finishToolbarLayout(toolbar)
	return toolbar
}

func (e *editor[N, D]) prepareForSubstitutions() (tmpNode N, m map[string]string) {
	tmpNode = e.target.Clone(e.target.GetSource().LibraryFile, e.target.DataOwner(), nil, gurps.Copy)
	e.editorData.ApplyTo(tmpNode)
	m = make(map[string]string)
	tmpNode.FillWithNameableKeys(m, nil)
	return tmpNode, m
}

func (e *editor[N, D]) Title() string {
	return fmt.Sprintf(i18n.Text("%s Editor for %s"), e.target.Kind(), e.owner.String())
}

func (e *editor[N, D]) Owner() Rebuildable {
	return e.owner
}

func (e *editor[N, D]) Target() N {
	return e.target
}

var pruneIDFields = regexp.MustCompile(`\s*"id":\s*"[^"]+",?\s*`)

func (e *editor[N, D]) isModified() bool {
	d1, err := gurps.MarshalWithoutCalc(e.beforeData)
	if err != nil {
		errs.Log(err)
		return false
	}
	var d2 []byte
	d2, err = gurps.MarshalWithoutCalc(e.editorData)
	if err != nil {
		errs.Log(err)
		return false
	}
	none := []byte{}
	d1 = pruneIDFields.ReplaceAll(d1, none)
	d2 = pruneIDFields.ReplaceAll(d2, none)
	return !bytes.Equal(d1, d2)
}

func (e *editor[N, D]) Modified() bool {
	modified := e.enableApplyAndCancel(e.isModified())
	if e.nameablesButton != nil {
		e.nameablesButton.SetEnabled(e.hasNameableKeys())
	}
	return modified
}

// hasNameableKeys reports whether the current editor data contains any nameable keys. It reuses a single scratch node
// to avoid cloning the target on every call, since Modified is invoked frequently during layout and redraw.
func (e *editor[N, D]) hasNameableKeys() bool {
	if xreflect.IsNil(e.nameablesScratch) {
		e.nameablesScratch = e.target.Clone(e.target.GetSource().LibraryFile, e.target.DataOwner(), nil, gurps.Copy)
		e.nameablesKeys = make(map[string]string)
	} else {
		clear(e.nameablesKeys)
	}
	e.editorData.ApplyTo(e.nameablesScratch)
	e.nameablesScratch.FillWithNameableKeys(e.nameablesKeys, nil)
	return len(e.nameablesKeys) > 0
}

func (e *editor[N, D]) MarkModified(_ unison.Paneler) {
	// Editing a field re-syncs the editor's live previews (extended value/weight, markdown, etc.), which resolve the
	// in-progress, often incomplete script expressions the user is still typing. Suppress the error logging those
	// failed resolutions would otherwise produce; resolutions performed anywhere else continue to log normally.
	gurps.SuppressScriptResolveErrorLogging(func() {
		UpdateTitleForDockable(e)
		DeepSync(e)
		if e.modificationCallback != nil {
			e.modificationCallback()
		}
	})
	// The editor's tables show row state -- a modifier's enabled checkmark, a weapon's Hide checkmark -- that nothing
	// above has marked for redraw, since DeepSync only reaches Syncers and neither unison.Table nor the panels wrapping
	// the editor's tables is one. A cell click flips its own drawable by hand, but the command path and the undo or
	// redo of either has nothing to flip, so without this a toggle could leave a stale checkmark on screen.
	// Node.ColumnCell re-derives the cell data on draw, so a redraw is all that is needed.
	e.MarkForRedraw()
}

func (e *editor[N, D]) Rebuild(_ bool) {
	if entity := gurps.EntityFromNode(e.target); entity != nil {
		entity.DiscardCaches()
	} else {
		gurps.DiscardGlobalResolveCache()
	}
	e.MarkModified(nil)
	e.MarkForLayoutRecursively()
}

func (e *editor[N, D]) AttemptClose() bool {
	if !CloseGroup(e) || !e.confirmClose(e.isModified, e.apply) {
		return false
	}
	if p := e.returnToPrevious(); p != nil {
		if table, ok := p.Self.(*unison.Table[*Node[N]]); ok {
			revealRowForData(table, e.target)
		}
	}
	return AttemptCloseForDockable(e)
}

// revealRowForData scrolls the table's row holding data into view, but only if the table currently displays such a row
// and it is not already visible.
func revealRowForData[T gurps.Node[T]](table *unison.Table[*Node[T]], data T) {
	if index := rowIndexForData(table, data); index != -1 {
		table.ScrollRowIntoView(index)
	}
}

// rowIndexForData returns the index of the displayed row holding data, or -1 if the table does not currently display
// one.
func rowIndexForData[T gurps.Node[T]](table *unison.Table[*Node[T]], data T) int {
	id := data.ID()
	for i := table.LastRowIndex(); i >= 0; i-- {
		if row := table.RowFromIndex(i); row != nil && row.ID() == id {
			return i
		}
	}
	return -1
}

func (e *editor[N, D]) apply() {
	e.Window().FocusNext() // Move the focus to flush any pending edits
	if e.preApplyCallback != nil {
		e.preApplyCallback(e.editorData)
	}
	if mgr := unison.UndoManagerFor(e.owner); mgr != nil {
		owner := e.owner
		target := e.target
		mgr.Add(&unison.UndoEdit[D]{
			ID:       unison.NextUndoID(),
			EditName: fmt.Sprintf(i18n.Text("%s Changes"), target.Kind()),
			UndoFunc: func(edit *unison.UndoEdit[D]) {
				edit.BeforeData.ApplyTo(target)
				rebuildAsModified(owner, true)
			},
			RedoFunc: func(edit *unison.UndoEdit[D]) {
				edit.AfterData.ApplyTo(target)
				rebuildAsModified(owner, true)
			},
			BeforeData: e.beforeData,
			AfterData:  e.editorData,
		})
	}
	e.editorData.ApplyTo(e.target)
	rebuildAsModified(e.owner, true)
}
