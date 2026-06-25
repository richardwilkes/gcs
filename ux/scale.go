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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// ScaleDelta is the delta used when adjusting the view scale incrementally.
const ScaleDelta = 10

// NewScaleField creates a new scale field and hooks it into the target.
func NewScaleField(minValue, maxValue int, defValue, get func() int, set func(int), afterApply func(), attemptCenter, adjustForDisplayPPI bool, scroller *unison.ScrollPanel) *PercentageField {
	applyFunc := func() {
		scale := float32(get()) / 100
		if adjustForDisplayPPI {
			scale *= float32(gurps.GlobalSettings().General.MonitorPPI()) / 72
		}
		scalePt := geom.NewPoint(scale, scale)
		if header := scroller.ColumnHeader(); !xreflect.IsNil(header) {
			header.AsPanel().SetScale(scalePt)
		}
		scroller.Content().AsPanel().SetScale(scalePt)
	}
	scaleTitle := i18n.Text("Scale")
	overrideCenter := false
	var override geom.Point
	var scaleField *PercentageField
	scaleField = NewPercentageField(nil, "", scaleTitle, get,
		func(scale int) {
			if scaleField != nil && !scaleField.Enabled() {
				SetFieldValue(scaleField.Field, scaleField.Format(get()))
				return
			}
			var view, content *unison.Panel
			var viewRect geom.Rect
			var center geom.Point
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
		}, minValue, maxValue, false, false)
	scaleField.SetMarksModified(false)
	scaleField.Tooltip = newWrappedTooltip(scaleTitle)
	scroller.ContentView().MouseWheelCallback = func(where, delta geom.Point, mods mod.Modifiers) bool {
		if !mods.OptionDown() || !scaleField.Enabled() {
			return false
		}
		current := get()
		scale := current + int(delta.Y*ScaleDelta)
		if scale < minValue {
			scale = minValue
		} else if scale > maxValue {
			scale = maxValue
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
		installViewScaleHandlers(unison.Ancestor[unison.Dockable](scroller), minValue, maxValue, get, defValue,
			func(scale int) {
				if get() != scale {
					SetFieldValue(scaleField.Field, scaleField.Format(scale))
				}
			})
		applyFunc()
		if afterApply != nil {
			afterApply()
		}
	}
	if dockable := unison.Ancestor[unison.Dockable](scroller); xreflect.IsNil(dockable) {
		scroller.ParentChangedCallback = installFunc
	} else {
		installFunc()
	}
	return scaleField
}

func installViewScaleHandlers(paneler unison.Paneler, minValue, maxValue int, current, def func() int, adjuster func(scale int)) {
	p := paneler.AsPanel()
	installViewScaleHandler(p, ScaleDefaultItemID, minValue, maxValue, current, def, adjuster)
	installViewDeltaScaleHandler(p, ScaleUpItemID, ScaleDelta, minValue, maxValue, current, adjuster)
	installViewDeltaScaleHandler(p, ScaleDownItemID, -ScaleDelta, minValue, maxValue, current, adjuster)
	installViewScaleHandler(p, Scale25ItemID, minValue, maxValue, current, func() int { return 25 }, adjuster)
	installViewScaleHandler(p, Scale50ItemID, minValue, maxValue, current, func() int { return 50 }, adjuster)
	installViewScaleHandler(p, Scale75ItemID, minValue, maxValue, current, func() int { return 75 }, adjuster)
	installViewScaleHandler(p, Scale100ItemID, minValue, maxValue, current, func() int { return 100 }, adjuster)
	installViewScaleHandler(p, Scale200ItemID, minValue, maxValue, current, func() int { return 200 }, adjuster)
	installViewScaleHandler(p, Scale300ItemID, minValue, maxValue, current, func() int { return 300 }, adjuster)
	installViewScaleHandler(p, Scale400ItemID, minValue, maxValue, current, func() int { return 400 }, adjuster)
	installViewScaleHandler(p, Scale500ItemID, minValue, maxValue, current, func() int { return 500 }, adjuster)
	installViewScaleHandler(p, Scale600ItemID, minValue, maxValue, current, func() int { return 600 }, adjuster)
}

func installViewScaleHandler(p *unison.Panel, itemID, minValue, maxValue int, current, desired func() int, adjuster func(int)) {
	scale := desired()
	if minValue <= scale && maxValue >= scale {
		p.InstallCmdHandlers(itemID,
			func(_ any) bool {
				d := desired()
				return minValue <= d && maxValue >= d && current() != d
			},
			func(_ any) { adjuster(desired()) })
	}
}

func installViewDeltaScaleHandler(p *unison.Panel, itemID, delta, minValue, maxValue int, current func() int, adjuster func(int)) {
	calc := func() int {
		c := current()
		adjusted := c + delta
		if adjusted < minValue {
			adjusted = minValue
		} else if adjusted > maxValue {
			adjusted = maxValue
		}
		return adjusted
	}
	p.InstallCmdHandlers(itemID, func(_ any) bool { return calc() != current() }, func(_ any) { adjuster(calc()) })
}
