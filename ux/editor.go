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
	"reflect"
	"regexp"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/gcs/v5/ux/svg"
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

// The data of every kind of item whose editor gets the Set Substitutions button, which relies on it.
var (
	_ nameable.Setter = &gurps.EquipmentEditData{}
	_ nameable.Setter = &gurps.NoteEditData{}
	_ nameable.Setter = &gurps.SkillEditData{}
	_ nameable.Setter = &gurps.SpellEditData{}
	_ nameable.Setter = &gurps.TraitEditData{}
)

type editor[N gurps.Node[N], D gurps.EditorData[N]] struct {
	editorShell
	target               N
	nameablesButton      *unison.Button
	meleeWeapons         *weaponsPanel
	rangedWeapons        *weaponsPanel
	content              *unison.Panel
	initContent          func(*editor[N, D], *unison.Panel) func()
	beforeData           D
	editorData           D
	nameablesScratch     N
	nameablesKeys        map[string]string
	modificationCallback func()
	preApplyCallback     func(D)
	scale                int
	// syncedChoice is the choice picker a pending "Sync with Source" last changed the target's to, or nil if none has
	// changed it.
	syncedChoice *gurps.TemplatePicker
	// sourceCleared is a pending "Clear Source", since the source isn't part of the editor's data.
	sourceCleared bool
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

	e.beforeData = newEditorData[N, D](target)
	e.editorData = newEditorData[N, D](target)

	content := e.setUp(2)
	e.AddChild(e.createToolbar(helpMD, initToolbar))
	e.AddChild(e.scroll)
	e.fillContent(content, initContent)
	e.placeInDock(target.ID())
	content.RequestFocus()
	return e
}

func (e *editor[N, D]) createToolbar(helpMD string, initToolbar func(*editor[N, D], *unison.Panel)) unison.Paneler {
	toolbar := NewToolbar()
	toolbar.AddChild(NewDefaultInfoPop())

	if helpMD != "" {
		addHelpButton(toolbar, helpMD)
	}

	AddUIScaleField(toolbar, func() int { return gurps.GlobalSettings().General.InitialEditorUIScale },
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
						if !promptForNameables(promptOperation{}, []nameablesSection{{Title: tmp.String(), Nameables: m}}) {
							return
						}
						tmp.ApplyNameableKeys(m)
						// Applying nameable keys only alters the replacements map, so copy just that back, which the
						// data of every kind of item that gets this button can do. CopyFrom would replace the entire
						// object graph with fresh sub-objects, orphaning the widgets already bound to the existing ones
						// (e.g. the modifiers table) and silently losing their later edits.
						if setter, ok2 := any(e.editorData).(nameable.Setter); ok2 {
							setter.SetNameableReplacements(tmp.NameableReplacements())
							e.Rebuild(false)
							e.showSubstitutions()
						}
					}
				}
				toolbar.AddChild(e.nameablesButton)
			}
		}
		toolbar.AddChild(e.newSourceButton())
	}

	if initToolbar != nil {
		initToolbar(e, toolbar)
	}

	FinishToolbarLayout(toolbar)
	return toolbar
}

// fillContent fills the content panel using initContent, keeping both so that rebuildContent can do it again.
func (e *editor[N, D]) fillContent(content *unison.Panel, initContent func(*editor[N, D], *unison.Panel) func()) {
	e.content = content
	e.initContent = initContent
	e.modificationCallback = initContent(e, content)
}

// rebuildContent refills the content panel after the editor's data has been replaced wholesale, which leaves the old
// widgets bound to objects it no longer holds. Sections of rows are shown as they were, collapsed or not and with the
// same row open. The undo history is cleared, since its edits refer to those widgets and objects, and only once the new
// content has settled, since settling may record edits of its own, as the equipment editor does when it clamps the uses
// left to a lower maximum. Not for the weapon editor, whose initContent keeps state across calls.
func (e *editor[N, D]) rebuildContent() {
	sections := rowSections(e.content)
	views := make([]rowsView, len(sections))
	for i, one := range sections {
		views[i] = one.view()
	}
	e.content.RemoveAllChildren()
	e.meleeWeapons = nil
	e.rangedWeapons = nil
	e.modificationCallback = e.initContent(e, e.content)
	// Which sections there are depends on the target, which is the same, so they come in the same order.
	if sections = rowSections(e.content); len(sections) == len(views) {
		for i, one := range sections {
			one.setView(views[i])
		}
	}
	// Copying the data from another node leaves its modifiers and weapons pointed at that node.
	for _, child := range e.content.Children() {
		if list, ok := child.Self.(interface{ reattach() }); ok {
			list.reattach()
		}
	}
	e.Rebuild(false)
	e.content.ValidateScrollRoot()
	e.undoMgr.Clear()
}

// showSubstitutions makes the rows of the editor's prerequisites, defaults and features again, and those of the editors
// open on its weapons, which take their substitutions from it, so that they show the substitutions set since they were
// made.
func (e *editor[N, D]) showSubstitutions() {
	for _, section := range rowSections(e.content) {
		section.rebuild("")
	}
	traverseGroup(e, func(target GroupedCloser) bool {
		if weaponEditor, ok := target.(*editor[*gurps.Weapon, *gurps.Weapon]); ok && weaponEditor.content != nil {
			for _, section := range rowSections(weaponEditor.content) {
				section.rebuild("")
			}
		}
		return false
	})
}

// pendingNameables returns the editor's data as the source of the substitutions it holds, which Set Substitutions
// changes ahead of their being applied, or nil if its data can't hold substitutions.
func (e *editor[N, D]) pendingNameables() nameable.Accesser {
	if source, ok := any(e.editorData).(nameable.Accesser); ok {
		return source
	}
	return nil
}

// rowSection is a section of an editor's content made of sentence rows, whose view rebuildContent keeps, and which
// Set Substitutions rebuilds.
type rowSection interface {
	view() rowsView
	setView(view rowsView)
	rebuild(focus string)
}

// rowSections returns the sections of the content made of sentence rows, in order.
func rowSections(content *unison.Panel) []rowSection {
	var sections []rowSection
	for _, child := range content.Children() {
		if section, ok := child.Self.(rowSection); ok {
			sections = append(sections, section)
		}
	}
	return sections
}

// editedClone returns a copy of the target with the editor's pending changes applied. It has no parent, so that nothing
// done to it can reach into the target's tree.
func (e *editor[N, D]) editedClone() N {
	clone := e.target.Clone(e.target.GetSource().LibraryFile, e.target.DataOwner(), nil, gurps.Copy)
	e.editorData.ApplyTo(clone)
	if e.sourceCleared {
		clone.ClearSource()
	}
	return clone
}

func (e *editor[N, D]) prepareForSubstitutions() (tmpNode N, m map[string]string) {
	tmpNode = e.editedClone()
	m = make(map[string]string)
	tmpNode.FillWithNameableKeys(m, nil)
	return tmpNode, m
}

func (e *editor[N, D]) Title() string {
	return i18n.Text("%s Editor for %s", e.target.Kind(), e.owner.String())
}

func (e *editor[N, D]) Owner() Rebuildable {
	return e.owner
}

func (e *editor[N, D]) Target() N {
	return e.target
}

var pruneIDFields = regexp.MustCompile(`\s*"id":\s*"[^"]+",?\s*`)

func (e *editor[N, D]) isModified() bool {
	if e.sourceCleared {
		return true
	}
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
	// Syncing refreshes the editor's live previews (extended value/weight, markdown, etc.), which resolve the often
	// incomplete script expressions the user is still typing, so don't log their resolution failures.
	gurps.SuppressScriptResolveErrorLogging(func() {
		UpdateTitleForDockable(e)
		DeepSync(e)
		if e.modificationCallback != nil {
			e.modificationCallback()
		}
	})
	// The editor's tables show row state -- a modifier's enabled checkmark, a weapon's Hide checkmark -- that DeepSync
	// misses, since neither unison.Table nor the panels wrapping the editor's tables is a Syncer. A cell click flips
	// its own drawable, but toggling by command, undo or redo would otherwise leave a stale checkmark. Node.ColumnCell
	// re-derives the cell data on draw, so a redraw is enough.
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
	e.applyEdits()
}

// applyEdits commits the editor's data to its target and records the change as an undoable edit.
func (e *editor[N, D]) applyEdits() {
	if e.preApplyCallback != nil {
		e.preApplyCallback(e.editorData)
	}
	owner := e.owner
	target := e.target
	// The source isn't part of the editor's data, so the undo edit has to carry it itself should applying the edit
	// have cleared it, whether by a pending "Clear Source" or by clearSourceOfTemplatePicker.
	sourceBefore := target.GetSource()
	// Nor are the other modifiers in the target's tree, which the choice rules may change along with it.
	optionsBefore := modifierEnabledStates([]N{target})
	wasEnabled := target.Enabled()
	changesEnabled := false
	if _, ok := gurps.ModifierChoiceFor(target); ok {
		changesEnabled = e.changesEnabled()
	}
	pickerBefore := modifierChoicePicker(target)
	wasChoice := gurps.IsModifierChoice(target)
	e.editorData.ApplyTo(target)
	if e.sourceCleared {
		target.ClearSource()
	}
	pickerAfter := modifierChoicePicker(target)
	applyModifierChoiceRulesAfterEdit(target, wasEnabled, changesEnabled, wasChoice, pickerAfter != pickerBefore,
		e.syncedChoice != nil && *e.syncedChoice == pickerAfter)
	clearSourceOfTemplatePicker(target)
	sourceAfter := target.GetSource()
	optionsAfter := modifierEnabledStates([]N{target})
	if mgr := unison.UndoManagerFor(owner); mgr != nil {
		mgr.Add(&unison.UndoEdit[D]{
			ID:       unison.NextUndoID(),
			EditName: i18n.Text("%s Changes", target.Kind()),
			UndoFunc: func(edit *unison.UndoEdit[D]) {
				edit.BeforeData.ApplyTo(target)
				restoreSource(target, sourceBefore, sourceAfter)
				restoreModifierEnabledStates(optionsBefore)
				rebuildAsModified(owner, true)
			},
			RedoFunc: func(edit *unison.UndoEdit[D]) {
				edit.AfterData.ApplyTo(target)
				restoreSource(target, sourceAfter, sourceBefore)
				restoreModifierEnabledStates(optionsAfter)
				rebuildAsModified(owner, true)
			},
			BeforeData: e.beforeData,
			AfterData:  e.editorData,
		})
	}
	rebuildAsModified(owner, true)
}

// changesEnabled returns true if the editor's data turns the target on or off.
func (e *editor[N, D]) changesEnabled() bool {
	scratch := e.target.Clone(gurps.LibraryFile{}, e.target.DataOwner(), e.target.Parent(), gurps.Copy)
	e.beforeData.ApplyTo(scratch)
	before := scratch.Enabled()
	e.editorData.ApplyTo(scratch)
	return scratch.Enabled() != before
}

// modifierChoicePicker returns the picker of a modifier container, the zero value for a group or anything else.
func modifierChoicePicker[N gurps.Node[N]](node N) gurps.TemplatePicker {
	if provider, ok := any(node).(gurps.ModifierChoiceProvider); ok && !xreflect.IsNil(node) && node.Container() {
		return provider.ModifierChoiceData().Choice
	}
	return gurps.TemplatePicker{}
}

// applyModifierChoiceRulesAfterEdit applies the choice rules after an editor's data has been applied to the target. An
// option keeps its current state unless the editor changed it, since another may have been picked since the editor
// opened; a container, which wasChoice says was a choice before the edit, is settled only when the editor changed what
// it asks for. When what it now asks for is what a pending sync with its source changed it to (choiceSynced), no pick
// is made for it, as when synced from its list; otherwise a choice the user made mandatory is given one.
func applyModifierChoiceRulesAfterEdit[N gurps.Node[N]](target N, wasEnabled, changesEnabled, wasChoice, choiceChanged, choiceSynced bool) {
	if wasChoice || gurps.IsModifierChoice(target) {
		switch {
		case !choiceChanged:
		case choiceSynced:
			gurps.SettleModifierChoicesAround(target)
		default:
			gurps.EnsureModifierChoiceRules(target)
		}
		return
	}
	if _, ok := gurps.ModifierChoiceFor(target); !ok {
		return
	}
	on := target.Enabled()
	gurps.SetModifierEnabled(target, wasEnabled)
	if changesEnabled {
		targets, enabled := gurps.ModifierEnabledChanges([]N{target}, func(N) bool { return on })
		for _, one := range targets {
			gurps.SetModifierEnabled(one, enabled[one])
		}
	}
}

// restoreSource sets the target's source to want when applying an edit changed it from other, leaving it alone
// otherwise, so that undoing or redoing an edit that didn't touch the source can't disturb it.
func restoreSource[N gurps.Node[N]](target N, want, other gurps.Source) {
	if want == other {
		return
	}
	if setter, ok := any(target).(interface{ SetSource(src gurps.Source) }); ok {
		setter.SetSource(want)
	}
}

// newEditorData returns new editor data holding a copy of the target's data.
func newEditorData[N gurps.Node[N], D gurps.EditorData[N]](target N) D {
	var data D
	reflect.ValueOf(&data).Elem().Set(reflect.New(reflect.TypeFor[D]().Elem()))
	data.CopyFrom(target)
	return data
}

// clearSourceOfTemplatePicker clears the source of a container that carries template picker data. Only a template may
// hold picker data and a template is never a source, so the source such a container points at can't have it, and
// syncing with that source would quietly take the choices away.
func clearSourceOfTemplatePicker[N gurps.Node[N]](target N) {
	if !xreflect.IsNil(target) && gurps.IsTemplateChoiceContainer(target) {
		target.ClearSource()
	}
}
