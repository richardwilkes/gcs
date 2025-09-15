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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/check"
)

type modifiersOnly interface {
	*gurps.TraitModifier | *gurps.EquipmentModifier
	fmt.Stringer
	nameable.Applier
}

// ProcessModifiersForSelection processes the selected rows for modifiers that can be toggled on or off.
func ProcessModifiersForSelection[T gurps.NodeTypes](table *unison.Table[*Node[T]]) {
	rows := table.SelectedRows(true)
	data := make([]T, 0, len(rows))
	for _, row := range rows {
		data = append(data, row.Data())
	}
	ProcessModifiers(table, data)
}

// ProcessModifiers processes the rows for modifiers that can be toggled on or off.
func ProcessModifiers[T gurps.NodeTypes](owner unison.Paneler, rows []T) {
	for _, row := range rows {
		gurps.Traverse(func(row T) bool {
			switch t := (any(row)).(type) {
			case *gurps.Trait:
				if processModifiers(xstrings.Truncate(gurps.AsNode(row).String(), 40, true), t.Modifiers) {
					unison.Ancestor[Rebuildable](owner).Rebuild(true)
				}
			case *gurps.Equipment:
				if processModifiers(xstrings.Truncate(gurps.AsNode(row).String(), 40, true), t.Modifiers) {
					unison.Ancestor[Rebuildable](owner).Rebuild(true)
				}
			}
			return false
		}, false, false, row)
	}
}

func processModifiers[T modifiersOnly](title string, modifiers []T) bool {
	if len(modifiers) == 0 {
		return false
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
		if mod, ok := any(m).(gurps.GeneralModifier); ok {
			text := mod.FullDescription()
			if cost := mod.FullCostDescription(); cost != "" {
				text += "; **" + cost + "**"
			}
			wrapper := unison.NewPanel()
			indent := float32(mod.Depth()) * indentIncrement
			md := unison.NewMarkdown(false)
			md.SetLayoutData(&unison.FlexLayoutData{
				HAlign: align.Fill,
				VAlign: align.Start,
				HGrab:  true,
			})
			md.SetContent(text, 800-indent)
			md.MouseDownCallback = func(_ geom.Point, _, _ int, _ unison.Modifiers) bool {
				return true
			}
			md.UpdateCursorCallback = func(_ geom.Point) *unison.Cursor {
				return unison.PointingCursor()
			}
			if !mod.Container() {
				cb := unison.NewCheckBox()
				cb.Font = unison.DefaultMarkdownTheme.Font
				cb.State = check.FromBool(mod.Enabled())
				tracker[cb] = mod
				cb.SetLayoutData(&unison.FlexLayoutData{
					VAlign: align.Start,
				})
				cb.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: 2}))
				md.MouseUpCallback = func(where geom.Point, _ int, _ unison.Modifiers) bool {
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
	if border, ok := children[len(children)-1].Border().(*unison.EmptyBorder); ok {
		insets := border.Insets()
		insets.Bottom = 0
		children[len(children)-1].SetBorder(unison.NewEmptyBorder(insets))
	}
	scroll := unison.NewScrollPanel()
	scroll.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.NewUniformInsets(1), false))
	scroll.SetContent(list, behavior.Fill, behavior.Fill)
	scroll.BackgroundInk = unison.ThemeSurface
	scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
	})
	label := unison.NewLabel()
	label.SetTitle(i18n.Text("Select Modifiers for:"))
	panel.AddChild(label)
	label = unison.NewLabel()
	label.Font = unison.SystemFont
	label.SetTitle(title)
	panel.AddChild(label)
	panel.AddChild(scroll)
	if unison.QuestionDialogWithPanel(panel) == unison.ModalResponseOK {
		changed := false
		for cb, mod := range tracker {
			if on := cb.State == check.On; mod.Enabled() != on {
				mod.SetEnabled(on)
				changed = true
			}
		}
		return changed
	}
	return false
}
