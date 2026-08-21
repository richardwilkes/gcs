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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/mod"
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

// nameableReplacementsSetter is implemented by editor data that can have its nameable replacements set directly.
type nameableReplacementsSetter interface {
	SetNameableReplacements(replacements map[string]string)
}

type editor[N gurps.Node[N], D gurps.EditorData[N]] struct {
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
	meleeWeapons         *weaponsPanel
	rangedWeapons        *weaponsPanel
	beforeData           D
	editorData           D
	nameablesScratch     N
	nameablesKeys        map[string]string
	modificationCallback func()
	preApplyCallback     func(D)
	scale                int
	promptForSave        bool
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
		target:           target,
		svg:              icon,
		scale:            gurps.GlobalSettings().General.InitialEditorUIScale,
		preApplyCallback: preApplyCallback,
	}
	e.Self = e

	if defDC := DefaultDockContainer(); defDC != nil {
		if e.previousDockable = defDC.CurrentDockable(); !xreflect.IsNil(e.previousDockable) {
			if focus := e.previousDockable.AsPanel().Window().Focus(); focus != nil {
				if focus.Ancestor[unison.Dockable]() == e.previousDockable {
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
	content.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(unison.StdHSpacing * 2)))
	content.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	content.KeyDownCallback = func(keyCode unison.KeyCode, mods mod.Modifiers, _ bool) bool {
		switch {
		case mods.OSMenuCommandDown() && (keyCode == unison.KeyReturn || keyCode == unison.KeyNumPadEnter):
			if e.applyButton.Enabled() {
				e.applyButton.Click()
			}
			return true
		case noModifiersDown(mods) && keyCode == unison.KeyEscape:
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
	e.AddChild(e.scroll)
	e.modificationCallback = initContent(e, content)
	e.ClientData()[AssociatedIDKey] = target.ID()
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
	return e
}

func (e *editor[N, D]) createToolbar(helpMD string, initToolbar func(*editor[N, D], *unison.Panel)) unison.Paneler {
	toolbar := unison.NewPanel()
	toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{},
		geom.Insets{Bottom: 1}, false), unison.NewEmptyBorder(unison.StdInsets())))

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
			false,
			e.scroll,
		),
	)

	e.applyButton = unison.NewSVGButton(unison.CheckmarkSVG)
	e.applyButton.Tooltip = newWrappedTooltipWithSecondaryText(i18n.Text("Apply Changes"),
		fmt.Sprintf(i18n.Text("%v%v or %v%v"), mod.OSMenuCommand(), unison.KeyReturn, mod.OSMenuCommand(),
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
						ShowNameablesDialog([]string{tmp.String()}, []map[string]string{m}, [][]string{nil})
						tmp.ApplyNameableKeys(m)
						// Applying nameable keys only alters the replacements map, so copy just that back into the
						// editor data. Using CopyFrom here would replace the entire object graph with fresh
						// sub-objects, orphaning the widgets that are already bound to the existing ones (e.g. the
						// modifiers table), causing their subsequent edits to be silently lost.
						if setter, ok2 := any(e.editorData).(nameableReplacementsSetter); ok2 {
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
	tmpNode = e.target.Clone(e.target.GetSource().LibraryFile, e.target.DataOwner(), nil, gurps.Copy)
	e.editorData.ApplyTo(tmpNode)
	m = make(map[string]string)
	tmpNode.FillWithNameableKeys(m, nil)
	return tmpNode, m
}

func (e *editor[N, D]) TitleIcon(suggestedSize geom.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  e.svg,
		Size: suggestedSize,
	}
}

func (e *editor[N, D]) Title() string {
	return fmt.Sprintf(i18n.Text("%s Editor for %s"), e.target.Kind(), e.owner.String())
}

func (e *editor[N, D]) String() string {
	return e.Title()
}

func (e *editor[N, D]) Tooltip() string {
	return ""
}

func (e *editor[N, D]) Owner() Rebuildable {
	return e.owner
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
	modified := e.isModified()
	e.applyButton.SetEnabled(modified)
	e.cancelButton.SetEnabled(modified)
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
}

func (e *editor[N, D]) Rebuild(_ bool) {
	if entity := gurps.EntityFromNode(e.target); entity != nil {
		entity.DiscardCaches()
	} else {
		gurps.DiscardGlobalResolveCache()
	}
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
	if p := showPreviousDockable(e.previousDockable, e.previousFocusKey); p != nil {
		restoreFocus(p)
		if table, ok := p.Self.(*unison.Table[*Node[N]]); ok {
			revealRowForData(table, e.target)
		}
	}
	return AttemptCloseForDockable(e)
}

// showPreviousDockable makes the dockable that was current when an editor was opened current again and returns the
// panel within it, identified by its RefKey, that held the keyboard focus at that time. It returns nil when there is no
// such dockable or panel.
func showPreviousDockable(previous unison.Dockable, focusKey string) *unison.Panel {
	if xreflect.IsNil(previous) {
		return nil
	}
	dc := unison.Ancestor[*unison.DockContainer](previous)
	if dc == nil {
		return nil
	}
	dc.SetCurrentDockable(previous)
	if focusKey == "" {
		return nil
	}
	return previous.AsPanel().FindRefKey(focusKey)
}

// focusWithoutScroller is implemented by panels, such as tables, that can take the keyboard focus without performing
// their default focus-gained scrolling.
type focusWithoutScroller interface {
	RequestFocusWithoutScroll()
}

// restoreFocus gives the keyboard focus back to a panel that held it before an editor was opened. A table's default
// focus handling scrolls the entire table into view, which moves the surrounding content even when the row the user was
// working with is still visible, so tables are focused without that. Callers that know which row matters can follow up
// with revealRowForData, which only scrolls if that row is actually out of view.
func restoreFocus(p *unison.Panel) {
	if f, ok := p.Self.(focusWithoutScroller); ok {
		f.RequestFocusWithoutScroll()
		return
	}
	p.RequestFocus()
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
