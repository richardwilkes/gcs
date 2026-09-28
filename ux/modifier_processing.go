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

	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/mod"
)

// The modifier prompts are held in variables so that tests can substitute non-interactive implementations. Each reports
// whether any modifier was changed and whether the prompt was canceled.
var (
	promptForTraitModifiers     = showModifiersDialog[*gurps.TraitModifier]
	promptForEquipmentModifiers = showModifiersDialog[*gurps.EquipmentModifier]
)

// modifierPromptInfo is what the modifier prompt shows about the row whose modifiers it asks about.
type modifierPromptInfo struct {
	// operation describes what the prompt is part of (see newOperationLabel). It may be empty.
	op promptOperation
	// name is the row's name.
	name string
	// location is the row's kind and the containers above it (see rowLocation).
	location string
	// step is the prompt's place among the rows being asked about, counting from 1, and steps is how many rows there
	// are. A steps of 1 or less shows no count.
	step, steps int
}

// ProcessModifiers prompts for which modifiers to enable on each row that can hold them (traits and equipment) and on
// every row below it. Other rows, the modifiers themselves included, preconfigured rows and rows without modifiers are
// skipped. Nothing is rebuilt here; the caller reports the change once the answers are in (see applyTransfer). The
// operation describes what the prompts are part of and may be empty (see newOperationLabel). Returns false if the user
// canceled a prompt, in which case no further prompts are shown and the caller is expected to abandon the whole
// operation the prompts were part of.
func ProcessModifiers[T gurps.Node[T]](op promptOperation, rows []T) bool {
	// The rows to be asked about are gathered first, so that each prompt can say how many there are in all. Answering
	// a prompt only toggles modifiers, which adds and takes away no rows, so the count holds throughout.
	var targets []T
	for _, row := range rows {
		gurps.Traverse(func(row T) bool {
			if !gurps.IsNodePreconfigured(row) && hasModifiers(row) {
				targets = append(targets, row)
			}
			return false
		}, false, false, row)
	}
	for i, row := range targets {
		info := modifierPromptInfo{
			op:       op,
			name:     row.String(),
			location: rowLocation(row),
			step:     i + 1,
			steps:    len(targets),
		}
		var canceled bool
		switch t := any(row).(type) {
		case *gurps.Trait:
			_, canceled = promptForTraitModifiers(&info, t.Modifiers)
		case *gurps.Equipment:
			_, canceled = promptForEquipmentModifiers(&info, t.Modifiers)
		}
		if canceled {
			return false
		}
	}
	return true
}

// hasModifiers returns true if the row is a trait or piece of equipment with modifiers.
func hasModifiers[T gurps.Node[T]](row T) bool {
	switch t := any(row).(type) {
	case *gurps.Trait:
		return len(t.Modifiers) != 0
	case *gurps.Equipment:
		return len(t.Modifiers) != 0
	default:
		return false
	}
}

func showModifiersDialog[T gurps.Node[T]](info *modifierPromptInfo, modifiers []T) (changed, canceled bool) {
	if len(modifiers) == 0 {
		return false, false
	}
	list := unison.NewPanel()
	list.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(unison.StdHSpacing)))
	list.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: 2 * unison.StdVSpacing,
	})
	tracker := make(map[*unison.CheckBox]gurps.GeneralModifier)
	indentIncrement := xmath.Ceil(unison.DefaultMarkdownTheme.Font.Baseline() * 1.5)
	gurps.Traverse(func(m T) bool {
		if gm, ok := any(m).(gurps.GeneralModifier); ok {
			text := gm.FullDescription()
			if cost := gm.FullCostDescription(); cost != "" {
				text += "; **" + cost + "**"
			}
			wrapper := unison.NewPanel()
			indent := float32(gm.Depth()) * indentIncrement
			md := unison.NewMarkdown(false)
			md.SetLayoutData(&unison.FlexLayoutData{
				HAlign: align.Fill,
				VAlign: align.Start,
				HGrab:  true,
			})
			md.SetContent(text, 800-indent)
			md.MouseDownCallback = func(_ geom.Point, _, _ int, _ mod.Modifiers) bool {
				return true
			}
			md.UpdateCursorCallback = func(_ geom.Point) *unison.Cursor {
				return unison.PointingCursor()
			}
			if !gm.Container() {
				cb := unison.NewCheckBox()
				cb.Font = unison.DefaultMarkdownTheme.Font
				cb.State = check.FromBool(gm.Enabled())
				tracker[cb] = gm
				cb.SetLayoutData(&unison.FlexLayoutData{
					VAlign: align.Start,
				})
				cb.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: 2}))
				md.MouseUpCallback = func(where geom.Point, _ int, _ mod.Modifiers) bool {
					if cb != nil && where.In(md.ContentRect(false)) {
						cb.Click()
					}
					return true
				}
				wrapper.AddChild(cb)
			}
			wrapper.AddChild(md)
			wrapper.SetLayout(&unison.FlexLayout{
				Columns:  len(wrapper.Children()),
				HSpacing: unison.StdHSpacing,
				VSpacing: unison.StdVSpacing,
				HAlign:   align.Fill,
				VAlign:   align.Start,
			})
			wrapper.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: indent}))
			list.AddChild(wrapper)
		}
		return false
	}, false, false, modifiers...)
	children := list.Children()
	if len(children) == 0 {
		return false, false
	}
	if border, ok := children[len(children)-1].Border().(*unison.EmptyBorder); ok {
		insets := border.Insets()
		insets.Bottom = 0
		children[len(children)-1].SetBorder(unison.NewEmptyBorder(insets))
	}
	header := i18n.Text("Select Modifiers for:")
	if info.steps > 1 {
		header = fmt.Sprintf(i18n.Text("Select Modifiers (%d of %d) for:"), info.step, info.steps)
	}
	extraHeaders := []*unison.Label{newTruncatedLabel(info.name, 60, unison.SystemFont)}
	if info.location != "" {
		extraHeaders = append(extraHeaders, newTruncatedLabel(info.location, 80, fonts.FieldSecondary))
	}
	if !showListQuestionDialog(info.op.at(i18n.Text("Modifiers")), header, list, extraHeaders...) {
		return false, true
	}
	for cb, gm := range tracker {
		if on := cb.State == check.On; gm.Enabled() != on {
			gm.SetEnabled(on)
			changed = true
		}
	}
	return changed, false
}
