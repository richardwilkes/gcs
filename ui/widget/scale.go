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

package widget

import (
	"github.com/richardwilkes/gcs/v5/constants"
	"github.com/richardwilkes/unison"
)

// InstallViewScaleHandlers installs the standard view scale handlers.
func InstallViewScaleHandlers(paneler unison.Paneler, def func() int, min, max int, current func() int, adjuster func(scale int)) {
	p := paneler.AsPanel()
	installViewScaleHandler(p, constants.ScaleDefaultItemID, def(), min, max, current, adjuster)
	installViewDeltaScaleHandler(p, constants.ScaleUpItemID, constants.ScaleDelta, min, max, current, adjuster)
	installViewDeltaScaleHandler(p, constants.ScaleDownItemID, -constants.ScaleDelta, min, max, current, adjuster)
	installViewScaleHandler(p, constants.Scale25ItemID, 25, min, max, current, adjuster)
	installViewScaleHandler(p, constants.Scale50ItemID, 50, min, max, current, adjuster)
	installViewScaleHandler(p, constants.Scale75ItemID, 75, min, max, current, adjuster)
	installViewScaleHandler(p, constants.Scale100ItemID, 100, min, max, current, adjuster)
	installViewScaleHandler(p, constants.Scale200ItemID, 200, min, max, current, adjuster)
	installViewScaleHandler(p, constants.Scale300ItemID, 300, min, max, current, adjuster)
	installViewScaleHandler(p, constants.Scale400ItemID, 400, min, max, current, adjuster)
	installViewScaleHandler(p, constants.Scale500ItemID, 500, min, max, current, adjuster)
	installViewScaleHandler(p, constants.Scale600ItemID, 600, min, max, current, adjuster)
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
