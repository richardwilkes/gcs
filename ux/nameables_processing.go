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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
)

// ProcessNameablesForSelection processes the selected rows and their children for any nameables.
func ProcessNameablesForSelection[T gurps.NodeTypes](table *unison.Table[*Node[T]]) {
	rows := table.SelectedRows(true)
	data := make([]T, 0, len(rows))
	for _, row := range rows {
		data = append(data, row.Data())
	}
	ProcessNameables(table, data)
}

// ProcessNameables processes the rows and their children for any nameables.
func ProcessNameables[T gurps.NodeTypes](owner unison.Paneler, rows []T) {
	var data []T
	var nameables []map[string]string
	for _, row := range rows {
		gurps.Traverse(func(row T) bool {
			m := make(map[string]string)
			gurps.AsNode(row).FillWithNameableKeys(m)
			if len(m) > 0 {
				data = append(data, row)
				nameables = append(nameables, m)
			}
			return false
		}, false, false, row)
	}
	if len(data) > 0 {
		list := unison.NewPanel()
		list.SetBorder(unison.NewEmptyBorder(unison.NewUniformInsets(unison.StdHSpacing)))
		list.SetLayout(&unison.FlexLayout{
			Columns:  2,
			HSpacing: unison.StdHSpacing,
			VSpacing: unison.StdVSpacing,
		})
		for i, one := range data {
			keys := make([]string, 0, len(nameables[i]))
			for k := range nameables[i] {
				keys = append(keys, k)
			}
			txt.SortStringsNaturalAscending(keys)
			if i != 0 {
				sep := unison.NewSeparator()
				sep.SetLayoutData(&unison.FlexLayoutData{
					HSpan:  2,
					HAlign: align.Fill,
					VAlign: align.Middle,
					HGrab:  true,
				})
				list.AddChild(sep)
			}
			header := unison.NewLabel()
			header.Text = txt.Truncate(gurps.AsNode(one).String(), 40, true)
			header.Font = unison.SystemFont
			header.SetLayoutData(&unison.FlexLayoutData{
				HSpan:  2,
				HAlign: align.Fill,
				VAlign: align.Middle,
				HGrab:  true,
			})
			list.AddChild(header)
			for _, k := range keys {
				label := unison.NewLabel()
				label.Text = k
				label.SetLayoutData(&unison.FlexLayoutData{
					HAlign: align.End,
					VAlign: align.Middle,
				})
				list.AddChild(label)
				list.AddChild(createNameableField(k, nameables[i]))
			}
		}
		scroll := unison.NewScrollPanel()
		scroll.SetBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.NewUniformInsets(1), false))
		scroll.SetContent(list, behavior.Fill, behavior.Fill)
		scroll.BackgroundInk = unison.ContentColor
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
		label.Text = i18n.Text("Provide substitutions:")
		panel.AddChild(label)
		panel.AddChild(scroll)
		if unison.QuestionDialogWithPanel(panel) == unison.ModalResponseOK {
			for i, row := range data {
				gurps.AsNode(row).ApplyNameableKeys(nameables[i])
			}
			unison.Ancestor[Rebuildable](owner).Rebuild(true)
		}
	}
}

func createNameableField(key string, m map[string]string) *unison.Field {
	field := unison.NewField()
	field.SetMinimumTextWidthUsing("Something reasonable")
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
		HGrab:  true,
	})
	m[key] = ""
	field.ModifiedCallback = func(_, after *unison.FieldState) {
		m[key] = after.Text
	}
	return field
}
