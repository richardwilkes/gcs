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
	"fmt"
	"io/fs"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

var _ GroupedCloser = &bodySettingsDockable{}

type bodySettingsDockable struct {
	SettingsDockable
	owner          BodySettingsOwner
	targetMgr      *TargetMgr
	undoMgr        *unison.UndoManager
	body           *gurps.Body
	originalHash   uint64
	toolbar        *unison.Panel
	content        *unison.Panel
	applyButton    *unison.Button
	cancelButton   *unison.Button
	dragTarget     *unison.Panel
	dragTargetBody *gurps.Body
	dragInsert     int
	promptForSave  bool
	inDragOver     bool
}

// ShowBodySettings the Body Settings. Pass in nil to edit the defaults or a sheet to edit the sheet's.
func ShowBodySettings(owner BodySettingsOwner) {
	if Activate(func(d unison.Dockable) bool {
		if s, ok := d.AsPanel().Self.(*bodySettingsDockable); ok && owner == s.owner {
			return true
		}
		return false
	}) {
		return
	}
	d := &bodySettingsDockable{
		owner:         owner,
		promptForSave: true,
	}
	d.Self = d
	d.targetMgr = NewTargetMgr(d)
	d.body = owner.BodySettings(false).Clone(d.owner.Entity(), nil)
	d.TabTitle = owner.BodySettingsTitle()
	d.TabIcon = svg.BodyType
	d.body.ResetTargetKeyPrefixes(d.targetMgr.NextPrefix)
	d.originalHash = gurps.Hash64(d.body)
	d.Extensions = []string{gurps.BodyExt, gurps.BodyExtAlt}
	d.undoMgr = unison.NewUndoManager(100, func(err error) { errs.Log(err) })
	d.Loader = d.load
	d.Saver = d.save
	d.Resetter = d.reset
	d.ModifiedCallback = d.modified
	d.WillCloseCallback = d.willClose
	d.Setup(d.addToStartToolbar, nil, d.initContent)
}

func (d *bodySettingsDockable) UndoManager() *unison.UndoManager {
	return d.undoMgr
}

func (d *bodySettingsDockable) modified() bool {
	modified := d.originalHash != gurps.Hash64(d.body)
	d.applyButton.SetEnabled(modified)
	d.cancelButton.SetEnabled(modified)
	return modified
}

func (d *bodySettingsDockable) willClose() bool {
	if d.promptForSave && d.originalHash != gurps.Hash64(d.body) {
		switch unison.YesNoCancelDialog(fmt.Sprintf(i18n.Text("Apply changes made to\n%s?"), d.Title()), "") {
		case unison.ModalResponseDiscard:
		case unison.ModalResponseOK:
			d.apply()
		case unison.ModalResponseCancel:
			return false
		}
	}
	return true
}

func (d *bodySettingsDockable) CloseWithGroup(other unison.Paneler) bool {
	return d.owner == other
}

func (d *bodySettingsDockable) addToStartToolbar(toolbar *unison.Panel) {
	d.toolbar = toolbar

	helpButton := unison.NewSVGButton(svg.Help)
	helpButton.Tooltip = newWrappedTooltip(i18n.Text("Help"))
	helpButton.ClickCallback = func() { HandleLink(nil, "md:Help/Interface/Body Type") }
	toolbar.AddChild(helpButton)

	d.applyButton = unison.NewSVGButton(unison.CheckmarkSVG)
	d.applyButton.Tooltip = newWrappedTooltip(i18n.Text("Apply Changes"))
	d.applyButton.SetEnabled(false)
	d.applyButton.ClickCallback = func() {
		d.apply()
		d.promptForSave = false
		d.AttemptClose()
	}
	toolbar.AddChild(d.applyButton)

	d.cancelButton = unison.NewSVGButton(svg.Not)
	d.cancelButton.Tooltip = newWrappedTooltip(i18n.Text("Discard Changes"))
	d.cancelButton.SetEnabled(false)
	d.cancelButton.ClickCallback = func() {
		d.promptForSave = false
		d.AttemptClose()
	}
	toolbar.AddChild(d.cancelButton)
}

func (d *bodySettingsDockable) initContent(content *unison.Panel) {
	d.content = content
	d.content.DataDragOverCallback = d.dataDragOver
	d.content.DataDragExitCallback = d.dataDragExit
	d.content.DataDragDropCallback = d.dataDragDrop
	d.content.DrawOverCallback = d.drawOver
	content.SetBorder(nil)
	content.SetLayout(&unison.FlexLayout{Columns: 1})
	content.AddChild(newBodySettingsPanel(d))
}

func (d *bodySettingsDockable) Entity() *gurps.Entity {
	return d.owner.Entity()
}

func (d *bodySettingsDockable) prepareUndo(title string) *unison.UndoEdit[*gurps.Body] {
	return &unison.UndoEdit[*gurps.Body]{
		ID:         unison.NextUndoID(),
		EditName:   title,
		UndoFunc:   func(e *unison.UndoEdit[*gurps.Body]) { d.applyBodyType(e.BeforeData) },
		RedoFunc:   func(e *unison.UndoEdit[*gurps.Body]) { d.applyBodyType(e.AfterData) },
		AbsorbFunc: func(_ *unison.UndoEdit[*gurps.Body], _ unison.Undoable) bool { return false },
		BeforeData: d.body.Clone(d.Entity(), nil),
	}
}

func (d *bodySettingsDockable) finishAndPostUndo(undo *unison.UndoEdit[*gurps.Body]) {
	undo.AfterData = d.body.Clone(d.Entity(), nil)
	d.UndoManager().Add(undo)
}

func (d *bodySettingsDockable) applyBodyType(bodyType *gurps.Body) {
	d.body = bodyType.Clone(d.Entity(), nil)
	d.sync()
}

func (d *bodySettingsDockable) reset() {
	undo := d.prepareUndo(i18n.Text("Reset Body Type"))
	d.body = d.owner.BodySettings(true).Clone(d.owner.Entity(), nil)
	d.body.ResetTargetKeyPrefixes(d.targetMgr.NextPrefix)
	d.finishAndPostUndo(undo)
	d.sync()
}

func (d *bodySettingsDockable) sync() {
	focusRefKey := d.targetMgr.CurrentFocusRef()
	scrollRoot := d.content.ScrollRoot()
	h, v := scrollRoot.Position()
	d.content.RemoveAllChildren()
	d.content.AddChild(newBodySettingsPanel(d))
	d.MarkForLayoutRecursively()
	d.MarkForRedraw()
	d.ValidateLayout()
	d.MarkModified(nil)
	d.targetMgr.ReacquireFocus(focusRefKey, d.toolbar, d.content)
	scrollRoot.SetPosition(h, v)
}

func (d *bodySettingsDockable) load(fileSystem fs.FS, filePath string) error {
	bodyType, err := gurps.NewBodyFromFile(fileSystem, filePath)
	if err != nil {
		return err
	}
	bodyType.ResetTargetKeyPrefixes(d.targetMgr.NextPrefix)
	undo := d.prepareUndo(i18n.Text("Load Body Type"))
	d.body = bodyType
	d.finishAndPostUndo(undo)
	d.sync()
	return nil
}

func (d *bodySettingsDockable) save(filePath string) error {
	return d.body.Save(filePath)
}

func (d *bodySettingsDockable) apply() {
	d.Window().FocusNext() // Intentionally move the focus to ensure any pending edits are flushed
	d.owner.SetBodySettings(d.body.Clone(d.owner.Entity(), nil))
}

func (d *bodySettingsDockable) dataDragOver(where unison.Point, data map[string]any) bool {
	prevInDragOver := d.inDragOver
	dragInsert := d.dragInsert
	dragTarget := d.dragTarget
	d.inDragOver = false
	d.dragInsert = -1
	d.dragTargetBody = nil
	d.dragTarget = nil
	if dragData, ok := data[hitLocationDragDataKey]; ok {
		var dd *hitLocationSettingsPanel
		if dd, ok = dragData.(*hitLocationSettingsPanel); ok && dd.dockable == d {
			parent := dd.Parent()
			where = parent.PointFromRoot(d.content.PointToRoot(where))
			for i, child := range parent.Children() {
				rect := child.FrameRect()
				if where.In(rect) {
					d.dragTarget = parent
					if rect.CenterY() <= where.Y {
						d.dragInsert = i + 1
					} else {
						d.dragInsert = i
					}
					d.inDragOver = true
					break
				}
			}
		}
	}
	if prevInDragOver != d.inDragOver || dragInsert != d.dragInsert || dragTarget != d.dragTarget {
		d.MarkForRedraw()
	}
	return true
}

func (d *bodySettingsDockable) dataDragExit() {
	d.inDragOver = false
	d.dragInsert = -1
	d.dragTargetBody = nil
	d.dragTarget = nil
	d.MarkForRedraw()
}

func (d *bodySettingsDockable) dataDragDrop(_ unison.Point, data map[string]any) {
	if d.inDragOver && d.dragInsert != -1 {
		if dragData, ok := data[hitLocationDragDataKey]; ok {
			var dd *hitLocationSettingsPanel
			if dd, ok = dragData.(*hitLocationSettingsPanel); ok && dd.dockable == d && d.dragInsert != -1 {
				undo := d.prepareUndo(i18n.Text("Hit Location Drag"))
				table := dd.loc.OwningTable()
				i := slices.Index(table.Locations, dd.loc)
				table.Locations = slices.Delete(table.Locations, i, i+1)
				if i < d.dragInsert {
					d.dragInsert--
				}
				table.Locations = slices.Insert(table.Locations, d.dragInsert, dd.loc)
				table.Update(d.Entity())
				d.finishAndPostUndo(undo)
				d.sync()
			}
		}
	}
	d.dataDragExit()
}

func (d *bodySettingsDockable) drawOver(gc *unison.Canvas, rect unison.Rect) {
	if d.inDragOver && d.dragInsert != -1 {
		children := d.dragTarget.Children()
		var y float32
		if d.dragInsert < len(children) {
			y = children[d.dragInsert].FrameRect().Y
		} else {
			y = children[len(children)-1].FrameRect().Bottom()
		}
		pt := d.content.PointFromRoot(d.dragTarget.PointToRoot(unison.Point{Y: y}))
		paint := unison.ThemeWarning.Paint(gc, rect, paintstyle.Stroke)
		paint.SetStrokeWidth(2)
		r := d.content.RectFromRoot(d.dragTarget.RectToRoot(d.dragTarget.ContentRect(false)))
		gc.DrawLine(r.X, pt.Y, r.Right(), pt.Y, paint)
	}
}
