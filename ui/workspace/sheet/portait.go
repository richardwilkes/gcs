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

package sheet

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// PortraitPanel holds the contents of the portrait block on the sheet.
type PortraitPanel struct {
	unison.Panel
	sheet *Sheet
}

// NewPortraitPanel creates a new portrait panel.
func NewPortraitPanel(sheet *Sheet) *PortraitPanel {
	p := &PortraitPanel{sheet: sheet}
	p.Self = p
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.StartAlignment,
		VAlign: unison.StartAlignment,
		VSpan:  2,
	})
	p.SetBorder(&widget.TitledBorder{Title: i18n.Text("Portrait")})
	p.Tooltip = unison.NewTooltipWithText(fmt.Sprintf(i18n.Text(`Double-click to set a character portrait, or drag an image onto this block.

The dimensions of the chosen picture should be in a ratio of 3 pixels wide
for every 4 pixels tall to scale without distortion.

Recommended minimum dimensions are %dx%d.`), gurps.PortraitWidth*2, gurps.PortraitHeight*2))
	p.DrawCallback = p.drawSelf
	return p
}

func (p *PortraitPanel) drawSelf(gc *unison.Canvas, _ unison.Rect) {
	r := p.ContentRect(false)
	paint := unison.ContentColor.Paint(gc, r, unison.Fill)
	gc.DrawRect(r, paint)
	if img := p.sheet.entity.Profile.Portrait(); img != nil {
		img.DrawInRect(gc, r, nil, paint)
	}
}

// Sync the panel to the current data.
func (p *PortraitPanel) Sync() {
	// Nothing to do
}
