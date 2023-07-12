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
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// ScaleDelta is the delta used when adjusting the view scale incrementally.
const ScaleDelta = 10

// NewScaleField creates a new scale field and hooks it into the target.
func NewScaleField(min, max int, def, get func() int, set func(int), afterApply func(), attemptCenter bool, scroller *unison.ScrollPanel) *PercentageField {
	applyFunc := func() {
		scale := float32(get()) / 100
		if header := scroller.ColumnHeader(); !toolbox.IsNil(header) {
			header.AsPanel().SetScale(scale)
		}
		scroller.Content().AsPanel().SetScale(scale)
	}
	scaleTitle := i18n.Text("Scale")
	overrideCenter := false
	var override unison.Point
	var scaleField *PercentageField
	scaleField = NewPercentageField(nil, "", scaleTitle, get,
		func(scale int) {
			if scaleField != nil && !scaleField.Enabled() {
				SetFieldValue(scaleField.Field, scaleField.Format(get()))
				return
			}
			var view, content *unison.Panel
			var viewRect unison.Rect
			var center unison.Point
			var dX, dY float32
			if attemptCenter {
				view = scroller.ContentView()
				viewRect = view.ContentRect(false)
				content = scroller.Content().AsPanel()
				if overrideCenter {
					center = override
				} else {
					center = viewRect.Center()
				}
				dX = center.X - viewRect.X
				dY = center.Y - viewRect.Y
				center = content.PointFromRoot(view.PointToRoot(center))
			}

			set(scale)
			applyFunc()
			scroller.MarkForLayoutRecursively()
			scroller.ValidateLayout()

			if content != nil {
				fScale := float32(scale) / 100
				viewRect.Width /= fScale
				viewRect.Height /= fScale
				viewRect.X = center.X - dX/fScale
				viewRect.Y = center.Y - dY/fScale
				content.ScrollRectIntoView(viewRect)
			}

			if afterApply != nil {
				afterApply()
			}
		}, min, max, false, false)
	scaleField.SetMarksModified(false)
	scaleField.Tooltip = unison.NewTooltipWithText(scaleTitle)
	scroller.ContentView().MouseWheelCallback = func(where, delta unison.Point, mod unison.Modifiers) bool {
		if !mod.OptionDown() || !scaleField.Enabled() {
			return false
		}
		current := get()
		scale := current + int(delta.Y*ScaleDelta)
		if scale < min {
			scale = min
		} else if scale > max {
			scale = max
		}
		if current != scale {
			overrideCenter = true
			override = where
			SetFieldValue(scaleField.Field, scaleField.Format(scale))
			overrideCenter = false
		}
		return true
	}
	installFunc := func() {
		scroller.ParentChangedCallback = nil
		installViewScaleHandlers(unison.Ancestor[unison.Dockable](scroller), def, min, max, get, func(scale int) {
			if get() != scale {
				SetFieldValue(scaleField.Field, scaleField.Format(scale))
			}
		})
		applyFunc()
		if afterApply != nil {
			afterApply()
		}
	}
	if dockable := unison.Ancestor[unison.Dockable](scroller); toolbox.IsNil(dockable) {
		scroller.ParentChangedCallback = installFunc
	} else {
		installFunc()
	}
	return scaleField
}

func installViewScaleHandlers(paneler unison.Paneler, def func() int, min, max int, current func() int, adjuster func(scale int)) {
	p := paneler.AsPanel()
	installViewScaleHandler(p, ScaleDefaultItemID, def(), min, max, current, adjuster)
	installViewDeltaScaleHandler(p, ScaleUpItemID, ScaleDelta, min, max, current, adjuster)
	installViewDeltaScaleHandler(p, ScaleDownItemID, -ScaleDelta, min, max, current, adjuster)
	installViewScaleHandler(p, Scale25ItemID, 25, min, max, current, adjuster)
	installViewScaleHandler(p, Scale50ItemID, 50, min, max, current, adjuster)
	installViewScaleHandler(p, Scale75ItemID, 75, min, max, current, adjuster)
	installViewScaleHandler(p, Scale100ItemID, 100, min, max, current, adjuster)
	installViewScaleHandler(p, Scale200ItemID, 200, min, max, current, adjuster)
	installViewScaleHandler(p, Scale300ItemID, 300, min, max, current, adjuster)
	installViewScaleHandler(p, Scale400ItemID, 400, min, max, current, adjuster)
	installViewScaleHandler(p, Scale500ItemID, 500, min, max, current, adjuster)
	installViewScaleHandler(p, Scale600ItemID, 600, min, max, current, adjuster)
}

func installViewScaleHandler(p *unison.Panel, itemID, scale, min, max int, current func() int, adjuster func(int)) {
	if min <= scale && max >= scale {
		p.InstallCmdHandlers(itemID, func(_ any) bool { return current() != scale }, func(_ any) { adjuster(scale) })
	}
}

func installViewDeltaScaleHandler(p *unison.Panel, itemID, delta, min, max int, current func() int, adjuster func(int)) {
	calc := func() int {
		c := current()
		adjusted := c + delta
		if adjusted < min {
			adjusted = min
		} else if adjusted > max {
			adjusted = max
		}
		return adjusted
	}
	p.InstallCmdHandlers(itemID, func(_ any) bool { return calc() != current() }, func(_ any) { adjuster(calc()) })
}
