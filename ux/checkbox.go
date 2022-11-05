/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/unison"
)

// CheckBox provides a checkbox that works with undo.
type CheckBox struct {
	*unison.CheckBox
	undoID    int64
	targetMgr *TargetMgr
	targetKey string
	get       func() unison.CheckState
	set       func(state unison.CheckState)
	last      unison.CheckState
}

// NewCheckBox creates a new check box.
func NewCheckBox(targetMgr *TargetMgr, targetKey, title string, get func() unison.CheckState, set func(unison.CheckState)) *CheckBox {
	c := &CheckBox{
		CheckBox:  unison.NewCheckBox(),
		undoID:    unison.NextUndoID(),
		targetMgr: targetMgr,
		targetKey: targetKey,
		get:       get,
		set:       set,
	}
	c.Self = c
	c.Text = title
	c.Sync()
	c.ClickCallback = func() {
		if c.last != c.State {
			c.last = c.State
			if mgr := unison.UndoManagerFor(c); mgr != nil {
				undo := NewTargetUndo(c.targetMgr, c.targetKey, c.Text, c.undoID, func(target *unison.Panel, data unison.CheckState) {
					self := c
					if target != nil {
						if field, ok := target.Self.(*CheckBox); ok {
							self = field
						}
					}
					self.State = data
					self.set(data)
					self.MarkForRedraw()
					MarkModified(self)
				}, c.get())
				undo.AfterData = c.State
				mgr.Add(undo)
			}
			c.set(c.State)
			MarkModified(c)
		}
	}
	if targetMgr != nil && targetKey != "" {
		c.RefKey = targetKey
	}
	return c
}

// Sync the checkbox to the current value.
func (c *CheckBox) Sync() {
	c.State = c.get()
	c.last = c.State
	c.MarkForRedraw()
}
