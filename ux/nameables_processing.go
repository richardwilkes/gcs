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
	var titles []string
	var nameables []map[string]string
	for _, row := range rows {
		gurps.Traverse(func(row T) bool {
			m := make(map[string]string)
			gurps.AsNode(row).FillWithNameableKeys(m, nil)
			if len(m) > 0 {
				data = append(data, row)
				titles = append(titles, gurps.AsNode(row).String())
				nameables = append(nameables, m)
			}
			return false
		}, false, false, row)
	}
	if len(data) > 0 {
		if ShowNameablesDialog(titles, nameables) {
			for i, row := range data {
				gurps.AsNode(row).ApplyNameableKeys(nameables[i])
			}
			if rebuildable := unison.Ancestor[Rebuildable](owner); rebuildable != nil {
				rebuildable.Rebuild(true)
			}
		}
	}
}

// ShowNameablesDialog shows a dialog for editing nameables.
func ShowNameablesDialog(titles []string, nameables []map[string]string) bool {
	list := unison.NewPanel()
	list.SetBorder(unison.NewEmptyBorder(unison.NewUniformInsets(unison.StdHSpacing)))
	list.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	for i, one := range titles {
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
		header.Font = unison.SystemFont
		header.SetTitle(txt.Truncate(one, 40, true))
		header.SetLayoutData(&unison.FlexLayoutData{
			HSpan:  2,
			HAlign: align.Fill,
			VAlign: align.Middle,
			HGrab:  true,
		})
		list.AddChild(header)
		for _, k := range keys {
			label := unison.NewLabel()
			label.SetTitle(k)
			label.SetLayoutData(&unison.FlexLayoutData{
				HAlign: align.End,
				VAlign: align.Middle,
			})
			list.AddChild(label)
			list.AddChild(createNameableField(k, nameables[i]))
		}
	}
	scroll := unison.NewScrollPanel()
	scroll.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, 0, unison.NewUniformInsets(1), false))
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
	label.SetTitle(i18n.Text("Provide substitutions:"))
	panel.AddChild(label)
	panel.AddChild(scroll)
	return unison.QuestionDialogWithPanel(panel) == unison.ModalResponseOK
}

func createNameableField(key string, m map[string]string) *unison.Field {
	field := unison.NewField()
	field.SetMinimumTextWidthUsing("Something reasonable")
	field.SetText(m[key])
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
		HGrab:  true,
	})
	field.ModifiedCallback = func(_, after *unison.FieldState) {
		m[key] = after.Text
	}
	return field
}
