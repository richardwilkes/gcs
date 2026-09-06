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
	"reflect"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

var (
	_ unison.Dockable            = &pointsEditor{}
	_ unison.TabCloser           = &pointsEditor{}
	_ ModifiableRoot             = &pointsEditor{}
	_ unison.UndoManagerProvider = &pointsEditor{}
	_ GroupedCloser              = &pointsEditor{}
	_ Rebuildable                = &pointsEditor{}
)

type pointsEditor struct {
	editorShell
	entity  *gurps.Entity
	content *unison.Panel
	before  []*gurps.PointsRecord
	current []*gurps.PointsRecord
}

// newPointsEditor returns an editor holding the two copies of the entity's points record list, without building any of
// the editor's UI. Both copies are put into the same order, since the editor displays and compares them in that order
// and a list loaded from a file may have been stored in some other order.
func newPointsEditor(owner Rebuildable, entity *gurps.Entity) *pointsEditor {
	e := &pointsEditor{
		owner:   owner,
		icon:    svg.Edit,
		entity:  entity,
		before:  gurps.ClonePointsRecordList(entity.PointsRecord),
		current: gurps.ClonePointsRecordList(entity.PointsRecord),
	}
	e.Self = e
	gurps.SortPointsRecordList(e.before)
	gurps.SortPointsRecordList(e.current)
	return e
}

func displayPointsEditor(owner Rebuildable, entity *gurps.Entity) {
	if Activate(func(d unison.Dockable) bool {
		if e, ok := d.AsPanel().Self.(*pointsEditor); ok {
			return e.owner == owner && entity == e.entity
		}
		return false
	}) {
		return
	}
	e := newPointsEditor(owner, entity)
	e.content = e.setUp(5)
	e.AddChild(e.createToolbar())
	e.AddChild(e.scroll)
	e.initContent()
	e.placeInDock(entity.ID)
	if children := e.content.Children(); len(children) != 0 {
		children[3].RequestFocus()
	}
}

func (e *pointsEditor) createToolbar() unison.Paneler {
	toolbar := newToolbar()
	addHelpButton(toolbar, "md:User%20Guide/Character%20Points")
	e.addApplyAndCancelButtons(toolbar, e.apply)
	toolbar.AddChild(NewToolbarSeparator())

	addButton := unison.NewSVGButton(unison.CircledAddSVG)
	addButton.Tooltip = newWrappedTooltip(i18n.Text("Add Entry"))
	addButton.ClickCallback = e.addEntry
	toolbar.AddChild(addButton)

	finishToolbarLayout(toolbar)
	return toolbar
}

func (e *pointsEditor) initContent() {
	for _, rec := range e.current {
		e.createRow(rec, -1)
	}
}

func (e *pointsEditor) createRow(rec *gurps.PointsRecord, index int) {
	deleteButton := unison.NewSVGButton(unison.TrashSVG)
	deleteButton.Tooltip = newWrappedTooltip(i18n.Text("Remove Entry"))
	deleteButton.ClickCallback = func() { e.removeEntry(rec) }
	e.content.AddChildAtIndex(deleteButton, index)
	if index != -1 {
		index++
	}

	var when *StringField
	whenText := i18n.Text("When")
	when = NewStringField(nil, "", whenText,
		func() string { return rec.When.String() },
		func(value string) {
			t, err := jio.NewTimeFrom(value)
			if err != nil {
				return
			}
			rec.When = t
			MarkModified(e.content)
		})
	when.ValidateCallback = func() bool {
		_, err := jio.NewTimeFrom(when.Text())
		return err == nil
	}
	when.Watermark = whenText
	when.SetMinimumTextWidthUsing(jio.Now().String() + "abcdefg")
	when.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill})
	e.content.AddChildAtIndex(when, index)
	if index != -1 {
		index++
	}

	pts := NewDecimalField(nil, "", i18n.Text("Points"),
		func() fxp.Int { return rec.Points },
		func(value fxp.Int) {
			rec.Points = value
			MarkModified(e.content)
		}, fxp.Min, fxp.Max, true, false)
	e.content.AddChildAtIndex(pts, index)
	if index != -1 {
		index++
	}

	reasonText := i18n.Text("Reason")
	reason := NewStringField(nil, "", reasonText,
		func() string { return rec.Reason },
		func(value string) {
			rec.Reason = value
			MarkModified(e.content)
		})
	reason.Watermark = reasonText
	e.content.AddChildAtIndex(reason, index)
	if index != -1 {
		index++
	}

	copyToOtherSheetButton := unison.NewSVGButton(svg.Stamper)
	copyToOtherSheetButton.Tooltip = newWrappedTooltip(i18n.Text("Copy to Other Character Sheet"))
	copyToOtherSheetButton.ClickCallback = func() { e.copyToOtherSheet(rec) }
	e.content.AddChildAtIndex(copyToOtherSheetButton, index)
}

func (e *pointsEditor) addEntry() {
	rec := &gurps.PointsRecord{When: jio.Now()}
	e.current = slices.Insert(e.current, 0, rec)
	e.createRow(rec, 0)
	e.content.Pack()
	e.content.MarkForRedraw()
	MarkModified(e.content)
	e.content.Children()[2].RequestFocus()
}

func (e *pointsEditor) removeEntry(rec *gurps.PointsRecord) {
	for i, one := range e.current {
		if one != rec {
			continue
		}
		e.current = slices.Delete(e.current, i, i+1)
		i *= 5
		for j := 4; j >= 0; j-- {
			e.content.RemoveChildAtIndex(i + j)
		}
		e.content.Pack()
		MarkForLayoutWithinDockable(e.content)
		e.content.MarkForRedraw()
		MarkModified(e.content)
		break
	}
}

func (e *pointsEditor) copyToOtherSheet(rec *gurps.PointsRecord) {
	availableSheets := OpenSheets(unison.AncestorOrSelf[*Sheet](e.owner))
	if len(availableSheets) == 0 {
		unison.WarningDialogWithMessage(i18n.Text("No other character sheets are open!"),
			i18n.Text("Open one or more other character sheets first."))
		return
	}
	sheets := PromptForDestination(availableSheets)
	if len(sheets) == 0 {
		return
	}
	entities := make(map[*gurps.Entity]bool, len(availableSheets))
	for _, sheet := range sheets {
		entities[sheet.entity] = true
	}
	for _, one := range AllMatchingDockables(func(d unison.Dockable) bool {
		if pe, ok := d.AsPanel().Self.(*pointsEditor); ok {
			return entities[pe.entity]
		}
		return false
	}) {
		if pe, ok := one.AsPanel().Self.(*pointsEditor); ok {
			if !pe.AttemptClose() {
				return
			}
		}
	}
	for _, sheet := range sheets {
		pe := newPointsEditor(sheet, sheet.entity)
		pe.current = slices.Insert(pe.current, 0, rec)
		gurps.SortPointsRecordList(pe.current)
		pe.applyWithoutFocusNext()
	}
}

func (e *pointsEditor) Title() string {
	return fmt.Sprintf(i18n.Text("Points Record for %s"), e.owner.String())
}

func (e *pointsEditor) Modified() bool {
	return e.enableApplyAndCancel(e.isModified())
}

func (e *pointsEditor) isModified() bool {
	return !reflect.DeepEqual(e.before, e.current)
}

func (e *pointsEditor) MarkModified(_ unison.Paneler) {
	UpdateTitleForDockable(e)
	DeepSync(e)
}

func (e *pointsEditor) Rebuild(_ bool) {
	gurps.DiscardGlobalResolveCache()
	e.MarkModified(nil)
	e.MarkForLayoutRecursively()
	e.MarkForRedraw()
}

func (e *pointsEditor) AttemptClose() bool {
	if !CloseGroup(e) || !e.confirmClose(e.isModified, e.apply) {
		return false
	}
	if dc := e.Ancestor[*unison.DockContainer](); dc != nil {
		dc.Close(e)
		e.returnToPrevious()
		return true
	}
	return e.Window().AttemptClose()
}

func (e *pointsEditor) apply() {
	e.Window().FocusNext() // Intentionally move the focus to ensure any pending edits are flushed
	e.applyWithoutFocusNext()
}

func (e *pointsEditor) applyWithoutFocusNext() {
	owner := e.owner
	entity := e.entity
	if mgr := unison.UndoManagerFor(owner); mgr != nil {
		mgr.Add(&unison.UndoEdit[[]*gurps.PointsRecord]{
			ID:       unison.NextUndoID(),
			EditName: i18n.Text("Point Record Changes"),
			UndoFunc: func(edit *unison.UndoEdit[[]*gurps.PointsRecord]) {
				entity.SetPointsRecord(edit.BeforeData)
				rebuildAsModified(owner, false)
			},
			RedoFunc: func(edit *unison.UndoEdit[[]*gurps.PointsRecord]) {
				entity.SetPointsRecord(edit.AfterData)
				rebuildAsModified(owner, false)
			},
			BeforeData: e.before,
			AfterData:  e.current,
		})
	}
	entity.SetPointsRecord(e.current)
	rebuildAsModified(owner, true)
}
