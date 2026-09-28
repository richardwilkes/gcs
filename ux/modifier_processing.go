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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/promptstep"
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
	// op is the operation the prompt is part of (see promptOperation).
	op promptOperation
	// name is the row's name.
	name string
	// location is the row's kind and the containers above it (see rowLocation). It is empty for a top-level row.
	location string
	// step is the prompt's place among the rows being asked about, counting from 1, and steps is how many rows there
	// are. A steps of 1 or less shows no count.
	step, steps int
}

// processModifiers prompts for which modifiers to enable on each of the rows that modifierTargets picks out of the given
// rows. Nothing is rebuilt here; the caller reports the change once the answers are in (see applyTransfer). Returns
// false if the user canceled a prompt, in which case no further prompts are shown and the caller is expected to abandon
// the whole operation the prompts were part of.
func processModifiers[T gurps.Node[T]](op promptOperation, rows []T) bool {
	targets := modifierTargets(rows)
	return promptForModifierTargets(op, targets, 0, len(targets))
}

// modifierTargets returns the rows the modifier prompt asks about: each of the given rows, and every row below them,
// that holds modifiers (traits and equipment), leaving out preconfigured ones. Other rows, the modifiers themselves
// included, are left out too.
func modifierTargets[T gurps.Node[T]](rows []T) []T {
	var targets []T
	for _, row := range rows {
		gurps.Traverse(func(row T) bool {
			if !gurps.IsNodePreconfigured(row) && hasModifiers(row) {
				targets = append(targets, row)
			}
			return false
		}, false, false, row)
	}
	return targets
}

// promptForModifierTargets puts up the modifier prompt for each of the targets (see modifierTargets). The prompts are
// counted as following the given number already done, out of total, since one transfer may ask about the rows of
// several lists. Answering a prompt only toggles modifiers, which adds and takes away no rows, so a count made
// beforehand holds throughout. Returns false if the user canceled a prompt, in which case no further prompts are shown.
func promptForModifierTargets[T gurps.Node[T]](op promptOperation, targets []T, done, total int) bool {
	for i, row := range targets {
		info := modifierPromptInfo{
			op:       op,
			name:     row.String(),
			location: rowLocation(row),
			step:     done + i + 1,
			steps:    total,
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
	if !showListQuestionDialog(info.op.at(promptstep.Modifiers), header, list, extraHeaders...) {
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
