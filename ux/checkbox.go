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
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/check"
)

// CheckBox provides a checkbox that works with undo.
type CheckBox struct {
	*unison.CheckBox
	undoID    int64
	targetMgr *TargetMgr
	targetKey string
	get       func() check.Enum
	set       func(state check.Enum)
	OnSet     func()
	last      check.Enum
}

// NewCheckBox creates a new check box.
func NewCheckBox(targetMgr *TargetMgr, targetKey, title string, get func() check.Enum, set func(check.Enum)) *CheckBox {
	c := &CheckBox{
		CheckBox:  unison.NewCheckBox(),
		undoID:    unison.NextUndoID(),
		targetMgr: targetMgr,
		targetKey: targetKey,
		get:       get,
		set:       set,
	}
	c.Self = c
	c.SetTitle(title)
	c.Sync()
	c.ClickCallback = func() {
		if c.last != c.State {
			c.last = c.State
			if mgr := unison.UndoManagerFor(c); mgr != nil {
				undo := NewTargetUndo(c.targetMgr, c.targetKey, c.Text.String(), c.undoID, func(target *unison.Panel, data check.Enum) {
					self := c
					if target != nil {
						if field, ok := target.Self.(*CheckBox); ok {
							self = field
						}
					}
					self.State = data
					self.set(data)
					self.MarkForRedraw()
					if self.OnSet != nil {
						self.OnSet()
					}
					MarkModified(self)
				}, c.get())
				undo.AfterData = c.State
				mgr.Add(undo)
			}
			c.set(c.State)
			if c.OnSet != nil {
				c.OnSet()
			}
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
	prevState := c.State
	c.State = c.get()
	c.last = c.State
	c.MarkForRedraw()
	if prevState != c.State {
		if c.OnSet != nil {
			c.OnSet()
		}
	}
}
