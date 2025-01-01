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
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

const attributeSettingsDragDataKey = "drag.attr"

var _ GroupedCloser = &attributeSettingsDockable{}

type attributeSettingsDockable struct {
	SettingsDockable
	owner           EntityPanel
	targetMgr       *TargetMgr
	undoMgr         *unison.UndoManager
	defs            *gurps.AttributeDefs
	originalHash    uint64
	toolbar         *unison.Panel
	content         *unison.Panel
	applyButton     *unison.Button
	cancelButton    *unison.Button
	dragTargetPool  *poolSettingsPanel
	defInsert       int
	thresholdInsert int
	lastDragData    map[string]any
	promptForSave   bool
	inDragOver      bool
	waitingForTimer bool
}

type attributeSettingsDragData struct {
	owner     *gurps.Entity
	def       *gurps.AttributeDef
	threshold *gurps.PoolThreshold
}

// ShowAttributeSettings the Attribute Settings. Pass in nil to edit the defaults or a sheet to edit the sheet's.
func ShowAttributeSettings(owner EntityPanel) {
	if Activate(func(d unison.Dockable) bool {
		if s, ok := d.AsPanel().Self.(*attributeSettingsDockable); ok && owner == s.owner {
			return true
		}
		return false
	}) {
		return
	}
	d := &attributeSettingsDockable{
		owner:         owner,
		promptForSave: true,
	}
	d.Self = d
	d.targetMgr = NewTargetMgr(d)
	if owner != nil {
		d.defs = d.owner.Entity().SheetSettings.Attributes.Clone()
		d.TabTitle = i18n.Text("Attributes: " + owner.Entity().Profile.Name)
	} else {
		d.defs = gurps.GlobalSettings().Sheet.Attributes.Clone()
		d.TabTitle = i18n.Text("Default Attributes")
	}
	d.TabIcon = svg.Attributes
	d.defs.ResetTargetKeyPrefixes(d.targetMgr.NextPrefix)
	d.originalHash = gurps.Hash64(d.defs)
	d.Extensions = []string{gurps.AttributesExt, gurps.AttributesExtAlt1, gurps.AttributesExtAlt2}
	d.undoMgr = unison.NewUndoManager(100, func(err error) { errs.Log(err) })
	d.Loader = d.load
	d.Saver = d.save
	d.Resetter = d.reset
	d.ModifiedCallback = d.modified
	d.WillCloseCallback = d.willClose
	d.Setup(d.addToStartToolbar, nil, d.initContent)
	scroller := unison.Ancestor[*unison.ScrollPanel](d.content)
	scroller.ScrollRectIntoViewCallback = func(rect unison.Rect) bool {
		if scroller.DefaultScrollRectIntoView(rect) {
			if d.inDragOver && !d.waitingForTimer {
				d.waitingForTimer = true
				unison.InvokeTaskAfter(func() {
					d.waitingForTimer = false
					if d.inDragOver {
						w := d.Window()
						pt := d.content.PointFromRoot(w.Content().PointToRoot(w.MouseLocation()))
						d.dataDragOver(pt, d.lastDragData)
					}
				}, 50*time.Millisecond)
			}
			return true
		}
		return false
	}
}

func (d *attributeSettingsDockable) UndoManager() *unison.UndoManager {
	return d.undoMgr
}

func (d *attributeSettingsDockable) modified() bool {
	modified := d.originalHash != gurps.Hash64(d.defs)
	d.applyButton.SetEnabled(modified)
	d.cancelButton.SetEnabled(modified)
	return modified
}

func (d *attributeSettingsDockable) willClose() bool {
	if d.promptForSave && d.originalHash != gurps.Hash64(d.defs) {
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

func (d *attributeSettingsDockable) CloseWithGroup(other unison.Paneler) bool {
	return d.owner != nil && d.owner == other
}

func (d *attributeSettingsDockable) addToStartToolbar(toolbar *unison.Panel) {
	d.toolbar = toolbar

	helpButton := unison.NewSVGButton(svg.Help)
	helpButton.Tooltip = newWrappedTooltip(i18n.Text("Help"))
	helpButton.ClickCallback = func() { HandleLink(nil, "md:Help/Interface/Attributes") }
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

	toolbar.AddChild(NewToolbarSeparator())

	addButton := unison.NewSVGButton(svg.CircledAdd)
	addAttributeText := i18n.Text("Add Attribute")
	addButton.Tooltip = newWrappedTooltip(addAttributeText)
	addButton.ClickCallback = func() {
		undo := &unison.UndoEdit[*gurps.AttributeDefs]{
			ID:         unison.NextUndoID(),
			EditName:   addAttributeText,
			UndoFunc:   func(e *unison.UndoEdit[*gurps.AttributeDefs]) { d.applyAttrDefs(e.BeforeData) },
			RedoFunc:   func(e *unison.UndoEdit[*gurps.AttributeDefs]) { d.applyAttrDefs(e.AfterData) },
			AbsorbFunc: func(_ *unison.UndoEdit[*gurps.AttributeDefs], _ unison.Undoable) bool { return false },
		}
		undo.BeforeData = d.defs.Clone()
		attrDef := &gurps.AttributeDef{}
		base := ""
		for {
			for v := 'a'; v <= 'z'; v++ {
				attempt := fmt.Sprintf("%s%c", base, v)
				if _, exists := d.defs.Set[attempt]; !exists {
					attrDef.DefID = attempt
					break
				}
			}
			if attrDef.DefID != "" {
				break
			}
			base += "a"
		}
		for _, v := range d.defs.Set {
			if attrDef.Order <= v.Order {
				attrDef.Order = v.Order + 1
			}
		}
		d.defs.Set[attrDef.DefID] = attrDef
		p := newAttrDefSettingsPanel(d, attrDef)
		d.content.AddChild(p)
		undo.AfterData = d.defs.Clone()
		d.UndoManager().Add(undo)
		d.MarkModified(nil)
		d.MarkForLayoutAndRedraw()
		d.ValidateLayout()
		FocusFirstContent(d.toolbar, p.AsPanel())
		d.Window().Focus().ScrollIntoView()
	}
	toolbar.AddChild(addButton)
}

func (d *attributeSettingsDockable) initContent(content *unison.Panel) {
	d.content = content
	d.content.DataDragOverCallback = d.dataDragOver
	d.content.DataDragExitCallback = d.dataDragExit
	d.content.DataDragDropCallback = d.dataDragDrop
	d.content.DrawOverCallback = d.drawOver
	content.SetBorder(nil)
	content.SetLayout(&unison.FlexLayout{Columns: 1})
	for _, def := range d.defs.List(false) {
		content.AddChild(newAttrDefSettingsPanel(d, def))
	}
}

func (d *attributeSettingsDockable) Entity() *gurps.Entity {
	if d.owner != nil {
		return d.owner.Entity()
	}
	return nil
}

func (d *attributeSettingsDockable) applyAttrDefs(defs *gurps.AttributeDefs) {
	d.defs = defs.Clone()
	d.sync()
}

func (d *attributeSettingsDockable) reset() {
	undo := &unison.UndoEdit[*gurps.AttributeDefs]{
		ID:         unison.NextUndoID(),
		EditName:   i18n.Text("Reset Attributes"),
		UndoFunc:   func(e *unison.UndoEdit[*gurps.AttributeDefs]) { d.applyAttrDefs(e.BeforeData) },
		RedoFunc:   func(e *unison.UndoEdit[*gurps.AttributeDefs]) { d.applyAttrDefs(e.AfterData) },
		AbsorbFunc: func(_ *unison.UndoEdit[*gurps.AttributeDefs], _ unison.Undoable) bool { return false },
		BeforeData: d.defs.Clone(),
	}
	if d.owner != nil {
		d.defs = gurps.GlobalSettings().Sheet.Attributes.Clone()
	} else {
		d.defs = gurps.FactoryAttributeDefs()
	}
	d.defs.ResetTargetKeyPrefixes(d.targetMgr.NextPrefix)
	undo.AfterData = d.defs.Clone()
	d.UndoManager().Add(undo)
	d.sync()
}

func (d *attributeSettingsDockable) sync() {
	focusRefKey := d.targetMgr.CurrentFocusRef()
	scrollRoot := d.content.ScrollRoot()
	h, v := scrollRoot.Position()
	d.content.RemoveAllChildren()
	for _, def := range d.defs.List(false) {
		d.content.AddChild(newAttrDefSettingsPanel(d, def))
	}
	d.MarkForLayoutAndRedraw()
	d.ValidateLayout()
	d.MarkModified(nil)
	d.targetMgr.ReacquireFocus(focusRefKey, d.toolbar, d.content)
	scrollRoot.SetPosition(h, v)
}

func (d *attributeSettingsDockable) load(fileSystem fs.FS, filePath string) error {
	defs, err := gurps.NewAttributeDefsFromFile(fileSystem, filePath)
	if err != nil {
		return err
	}
	defs.ResetTargetKeyPrefixes(d.targetMgr.NextPrefix)
	undo := &unison.UndoEdit[*gurps.AttributeDefs]{
		ID:         unison.NextUndoID(),
		EditName:   i18n.Text("Load Attributes"),
		UndoFunc:   func(e *unison.UndoEdit[*gurps.AttributeDefs]) { d.applyAttrDefs(e.BeforeData) },
		RedoFunc:   func(e *unison.UndoEdit[*gurps.AttributeDefs]) { d.applyAttrDefs(e.AfterData) },
		AbsorbFunc: func(_ *unison.UndoEdit[*gurps.AttributeDefs], _ unison.Undoable) bool { return false },
		BeforeData: d.defs.Clone(),
	}
	d.defs = defs
	undo.AfterData = d.defs.Clone()
	d.UndoManager().Add(undo)
	d.sync()
	return nil
}

func (d *attributeSettingsDockable) save(filePath string) error {
	return d.defs.Save(filePath)
}

func (d *attributeSettingsDockable) apply() {
	d.Window().FocusNext() // Intentionally move the focus to ensure any pending edits are flushed
	if d.owner == nil {
		gurps.GlobalSettings().Sheet.Attributes = d.defs.Clone()
		return
	}
	entity := d.owner.Entity()
	entity.SheetSettings.Attributes = d.defs.Clone()
	for attrID, def := range entity.SheetSettings.Attributes.Set {
		if attr, exists := entity.Attributes.Set[attrID]; exists {
			attr.Order = def.Order
		} else {
			entity.Attributes.Set[attrID] = gurps.NewAttribute(entity, attrID, def.Order)
		}
	}
	for attrID := range entity.Attributes.Set {
		if _, exists := d.defs.Set[attrID]; !exists {
			delete(entity.Attributes.Set, attrID)
		}
	}
	for _, one := range AllDockables() {
		if s, ok := one.(gurps.SheetSettingsResponder); ok {
			s.SheetSettingsUpdated(entity, true)
		}
	}
}

func (d *attributeSettingsDockable) dataDragOver(where unison.Point, data map[string]any) bool {
	d.lastDragData = data
	d.content.ScrollRectIntoView(unison.NewRect(where.X, where.Y-16, 1, 1))
	d.content.ScrollRectIntoView(unison.NewRect(where.X, where.Y+16, 1, 1))

	prevInDragOver := d.inDragOver
	prevDefInsert := d.defInsert
	prevThresholdInsert := d.thresholdInsert
	d.inDragOver = false
	d.defInsert = -1
	d.thresholdInsert = -1
	d.dragTargetPool = nil
	if dragData, ok := data[attributeSettingsDragDataKey]; ok {
		var dd *attributeSettingsDragData
		if dd, ok = dragData.(*attributeSettingsDragData); ok && dd.owner == d.Entity() {
			children := d.content.Children()
			rootPt := d.content.PointToRoot(where)
			if dd.threshold == nil {
				pt := d.content.PointFromRoot(rootPt)
				for i, child := range children {
					rect := child.FrameRect()
					if pt.In(rect) {
						if rect.CenterY() <= pt.Y {
							d.defInsert = i + 1
						} else {
							d.defInsert = i
						}
						d.inDragOver = true
						break
					}
				}
			} else {
				for i, def := range d.defs.List(false) {
					if def != dd.def || (def.Type != attribute.Pool && def.Type != attribute.PoolRef) {
						continue
					}
					pp, ok2 := children[i].Self.(*attrDefSettingsPanel)
					if !ok2 {
						continue
					}
					p := pp.poolPanel
					pt := p.PointFromRoot(rootPt)
					for j, child := range p.Children() {
						rect := child.FrameRect()
						if !pt.In(rect) {
							continue
						}
						d.dragTargetPool = p
						d.defInsert = i
						if rect.CenterY() <= pt.Y {
							d.thresholdInsert = j + 1
						} else {
							d.thresholdInsert = j
						}
						d.inDragOver = true
						break
					}
					if d.inDragOver {
						break
					}
				}
			}
		}
	}
	if prevInDragOver != d.inDragOver || prevDefInsert != d.defInsert || prevThresholdInsert != d.thresholdInsert {
		d.MarkForRedraw()
	}
	return true
}

func (d *attributeSettingsDockable) dataDragExit() {
	d.inDragOver = false
	d.defInsert = -1
	d.thresholdInsert = -1
	d.dragTargetPool = nil
	d.lastDragData = nil
	d.MarkForRedraw()
}

func (d *attributeSettingsDockable) dataDragDrop(_ unison.Point, data map[string]any) {
	if d.inDragOver && d.defInsert != -1 {
		if dragData, ok := data[attributeSettingsDragDataKey]; ok {
			var dd *attributeSettingsDragData
			if dd, ok = dragData.(*attributeSettingsDragData); ok {
				undo := &unison.UndoEdit[*gurps.AttributeDefs]{
					ID:         unison.NextUndoID(),
					UndoFunc:   func(e *unison.UndoEdit[*gurps.AttributeDefs]) { d.applyAttrDefs(e.BeforeData) },
					RedoFunc:   func(e *unison.UndoEdit[*gurps.AttributeDefs]) { d.applyAttrDefs(e.AfterData) },
					AbsorbFunc: func(_ *unison.UndoEdit[*gurps.AttributeDefs], _ unison.Undoable) bool { return false },
				}
				undo.BeforeData = d.defs.Clone()
				if d.thresholdInsert != -1 {
					undo.EditName = i18n.Text("Pool Threshold Drag")
					i := slices.Index(dd.def.Thresholds, dd.threshold)
					dd.def.Thresholds = slices.Delete(dd.def.Thresholds, i, i+1)
					if i < d.thresholdInsert {
						d.thresholdInsert--
					}
					dd.def.Thresholds = slices.Insert(dd.def.Thresholds, d.thresholdInsert, dd.threshold)
				} else {
					undo.EditName = i18n.Text("Attribute Definition Drag")
					list := d.defs.List(false)
					i := slices.Index(list, dd.def)
					list = slices.Delete(list, i, i+1)
					if i < d.defInsert {
						d.defInsert--
					}
					list = slices.Insert(list, d.defInsert, dd.def)
					for j, def := range list {
						def.Order = j
					}
				}
				undo.AfterData = d.defs.Clone()
				d.applyAttrDefs(undo.AfterData)
				d.UndoManager().Add(undo)
				d.MarkModified(nil)
				d.MarkForLayoutAndRedraw()
			}
		}
	}
	d.dataDragExit()
}

func (d *attributeSettingsDockable) drawOver(gc *unison.Canvas, rect unison.Rect) {
	if d.inDragOver {
		if d.thresholdInsert != -1 {
			children := d.dragTargetPool.Children()
			var y float32
			if d.thresholdInsert < len(children) {
				y = children[d.thresholdInsert].FrameRect().Y
			} else {
				y = children[len(children)-1].FrameRect().Bottom()
			}
			pt := d.content.PointFromRoot(d.dragTargetPool.PointToRoot(unison.Point{Y: y}))
			paint := unison.ThemeWarning.Paint(gc, rect, paintstyle.Stroke)
			paint.SetStrokeWidth(2)
			r := d.content.RectFromRoot(d.dragTargetPool.RectToRoot(d.dragTargetPool.ContentRect(false)))
			gc.DrawLine(r.X, pt.Y, r.Right(), pt.Y, paint)
		} else if d.defInsert != -1 {
			children := d.content.Children()
			var y float32
			if d.defInsert < len(children) {
				y = children[d.defInsert].FrameRect().Y
			} else {
				y = children[len(children)-1].FrameRect().Bottom()
			}
			paint := unison.ThemeWarning.Paint(gc, rect, paintstyle.Stroke)
			paint.SetStrokeWidth(2)
			gc.DrawLine(rect.X, y, rect.Right(), y, paint)
		}
	}
}
