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
	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
)

// ProcessModifiersForSelection processes the selected rows for modifiers that can be toggled on or off.
func ProcessModifiersForSelection[T model.NodeTypes](table *unison.Table[*Node[T]]) {
	rows := table.SelectedRows(true)
	data := make([]T, 0, len(rows))
	for _, row := range rows {
		data = append(data, row.Data())
	}
	ProcessModifiers(table, data)
}

// ProcessModifiers processes the rows for modifiers that can be toggled on or off.
func ProcessModifiers[T model.NodeTypes](owner unison.Paneler, rows []T) {
	for _, row := range rows {
		model.Traverse(func(row T) bool {
			switch t := (any(row)).(type) {
			case *model.Trait:
				if processModifiers(txt.Truncate(model.AsNode(row).String(), 40, true), t.Modifiers) {
					unison.Ancestor[Rebuildable](owner).Rebuild(true)
				}
			case *model.Equipment:
				if processModifiers(txt.Truncate(model.AsNode(row).String(), 40, true), t.Modifiers) {
					unison.Ancestor[Rebuildable](owner).Rebuild(true)
				}
			}
			return false
		}, false, false, row)
	}
}

func processModifiers[T *model.TraitModifier | *model.EquipmentModifier](title string, modifiers []T) bool {
	if len(modifiers) == 0 {
		return false
	}
	list := unison.NewPanel()
	list.SetBorder(unison.NewEmptyBorder(unison.NewUniformInsets(unison.StdHSpacing)))
	list.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	tracker := make(map[*unison.CheckBox]model.GeneralModifier)
	model.Traverse[T](func(m T) bool {
		if mod, ok := any(m).(model.GeneralModifier); ok {
			var p *unison.Panel
			if mod.Container() {
				label := unison.NewLabel()
				label.Text = mod.FullDescription()
				if cost := mod.FullCostDescription(); cost != "" {
					label.Text += " (" + cost + ")"
				}
				p = label.AsPanel()
			} else {
				cb := unison.NewCheckBox()
				cb.Text = mod.FullDescription()
				if cost := mod.FullCostDescription(); cost != "" {
					cb.Text += " (" + cost + ")"
				}
				cb.State = unison.CheckStateFromBool(mod.Enabled())
				tracker[cb] = mod
				p = cb.AsPanel()
			}
			p.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: float32(mod.Depth()) * 16}))
			list.AddChild(p)
		}
		return false
	}, false, false, modifiers...)
	scroll := unison.NewScrollPanel()
	scroll.SetBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.NewUniformInsets(1), false))
	scroll.SetContent(list, unison.FillBehavior, unison.FillBehavior)
	scroll.BackgroundInk = unison.ContentColor
	scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
		VGrab:  true,
	})
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   unison.FillAlignment,
		VAlign:   unison.FillAlignment,
	})
	label := unison.NewLabel()
	label.Text = i18n.Text("Select Modifiers for")
	panel.AddChild(label)
	label = unison.NewLabel()
	label.Text = title
	label.Font = unison.SystemFont
	panel.AddChild(label)
	panel.AddChild(scroll)
	if unison.QuestionDialogWithPanel(panel) == unison.ModalResponseOK {
		changed := false
		for _, row := range list.Children() {
			if cb, ok := row.Self.(*unison.CheckBox); ok {
				var mod model.GeneralModifier
				if mod, ok = tracker[cb]; ok {
					if on := cb.State == unison.OnCheckState; mod.Enabled() != on {
						mod.SetEnabled(on)
						changed = true
					}
				}
			}
		}
		return changed
	}
	return false
}
