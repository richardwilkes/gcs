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
	"strings"

	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

// NewDefaultInfoPop creates a new InfoPop with the message about mouse wheel scaling.
func NewDefaultInfoPop() *unison.Label {
	infoPop := NewInfoPop()
	AddScalingHelpToInfoPop(infoPop)
	return infoPop
}

// NewInfoPop creates a new InfoPop.
func NewInfoPop() *unison.Label {
	infoPop := unison.NewLabel()
	infoPop.OnBackgroundInk = unison.DefaultButtonTheme.OnBackgroundInk
	baseline := unison.DefaultButtonTheme.Font.Baseline()
	infoPop.Drawable = &unison.DrawableSVG{
		SVG:  svg.Info,
		Size: unison.NewSize(baseline, baseline).Ceil(),
	}
	return infoPop
}

// ClearInfoPop clears the InfoPop data.
func ClearInfoPop(target unison.Paneler) {
	panel := target.AsPanel()
	panel.Tooltip = nil
	panel.TooltipImmediate = false
}

// AddHelpToInfoPop adds one or more lines of help text to an InfoPop.
func AddHelpToInfoPop(target unison.Paneler, text string) {
	tip := prepareInfoPop(target)
	for _, str := range strings.Split(text, "\n") {
		if str == "" && len(tip.Children()) == 0 {
			continue
		}
		label := unison.NewLabel()
		label.LabelTheme = unison.DefaultTooltipTheme.Label
		label.SetTitle(str)
		label.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
		tip.AddChild(label)
	}
}

// AddScalingHelpToInfoPop adds the help info about scaling to an InfoPop.
func AddScalingHelpToInfoPop(target unison.Paneler) {
	AddHelpToInfoPop(target, fmt.Sprintf(i18n.Text(`
Holding down the %s key while using
the mouse wheel will change the scale.`), unison.OptionModifier.String()))
}

// AddKeyBindingInfoToInfoPop adds information about a key binding to an InfoPop.
func AddKeyBindingInfoToInfoPop(target unison.Paneler, keyBinding unison.KeyBinding, text string) {
	keyLabel := unison.NewLabel()
	keyLabel.LabelTheme = unison.DefaultTooltipTheme.Label
	keyLabel.OnBackgroundInk = unison.DefaultTooltipTheme.BackgroundInk
	keyLabel.Font = unison.DefaultMenuItemTheme.KeyFont
	keyLabel.SetTitle(keyBinding.String())
	keyLabel.HAlign = align.Middle
	keyLabel.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill})
	keyLabel.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, unison.DefaultTooltipTheme.Label.OnBackgroundInk.Paint(gc, rect, paintstyle.Fill))
		keyLabel.DefaultDraw(gc, rect)
	}
	keyLabel.SetBorder(unison.NewEmptyBorder(unison.NewHorizontalInsets(4)))
	tip := prepareInfoPop(target)
	tip.AddChild(keyLabel)

	descLabel := unison.NewLabel()
	descLabel.LabelTheme = unison.DefaultTooltipTheme.Label
	descLabel.SetTitle(text)
	tip.AddChild(descLabel)
}

func prepareInfoPop(target unison.Paneler) *unison.Panel {
	panel := target.AsPanel()
	panel.TooltipImmediate = true
	if panel.Tooltip == nil {
		panel.Tooltip = unison.NewTooltipBase()
		panel.Tooltip.SetLayout(&unison.FlexLayout{
			Columns:  2,
			HSpacing: unison.StdHSpacing,
			VSpacing: unison.StdVSpacing,
		})
	}
	return panel.Tooltip
}
