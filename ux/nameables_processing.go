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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
)

// ProcessNameablesForSelection processes the selected rows and their children for any nameables.
func ProcessNameablesForSelection[T gurps.Node[T]](table *unison.Table[*Node[T]]) {
	rows := table.SelectedRows(true)
	data := make([]T, 0, len(rows))
	for _, row := range rows {
		data = append(data, row.Data())
	}
	ProcessNameables(table, data)
}

// The nameables prompt is held in a variable so that tests can substitute a non-interactive implementation.
var promptForNameables = ShowNameablesDialog

// ProcessNameables processes the rows and their children for any nameables.
func ProcessNameables[T gurps.Node[T]](owner unison.Paneler, rows []T) {
	var data []T
	var titles []string
	var nameables []map[string]string
	var visibleKeys [][]string
	for _, row := range rows {
		gurps.Traverse(func(row T) bool {
			m := make(map[string]string)
			row.FillWithNameableKeys(m, nil)
			if len(m) == 0 {
				return false
			}
			var keys []string
			if gurps.IsNodePreconfigured(row) {
				// Only prompt for keys that don't already have a replacement recorded; the rest were already
				// resolved and shouldn't be asked about again.
				if keys = missingNameableKeys(row, m); len(keys) == 0 {
					return false
				}
			}
			data = append(data, row)
			titles = append(titles, row.String())
			nameables = append(nameables, m)
			visibleKeys = append(visibleKeys, keys) // nil means "show all keys"
			return false
		}, false, false, row)
	}
	if len(data) > 0 {
		if promptForNameables(titles, nameables, visibleKeys) {
			for i, row := range data {
				row.ApplyNameableKeys(nameables[i])
			}
			// The owner is normally the table the rows live in, and that table may have been replaced by a rebuild
			// before this was called -- the alternate drop path rebuilds before prompting, and the modifier prompt
			// that precedes this one rebuilds when its answer changes something, either of which can add or take away
			// the switch column, and a list can only change its columns by building a new table. An orphaned table
			// has no Rebuildable above it, so the rebuild would silently be skipped, leaving the substitutions in the
			// model while the list the user is looking at goes on showing the raw keys and the values derived from
			// them. The live table has to be looked up without regard for T, since on the very path this is here for
			// the rows are the modifiers that were dropped and T is therefore not the row type of the table they
			// landed in.
			rebuildAsModified(unison.AncestorOrSelf[Rebuildable](liveOwner(owner)), true)
		}
	}
}

// missingNameableKeys returns the keys in m that don't already have an explicit replacement recorded on row.
func missingNameableKeys[T gurps.Node[T]](row T, nameables map[string]string) []string {
	var replacements map[string]string
	if accessor, ok := any(row).(nameable.Accesser); ok && !xreflect.IsNil(accessor) {
		replacements = accessor.NameableReplacements()
	}
	return nameable.Missing(nameables, replacements)
}

// ShowNameablesDialog shows a dialog for editing nameables. For each row, visibleKeys restricts which keys of the
// corresponding nameables map are shown/editable; a nil entry shows all of that row's keys.
func ShowNameablesDialog(titles []string, nameables []map[string]string, visibleKeys [][]string) bool {
	list := unison.NewPanel()
	list.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(unison.StdHSpacing)))
	list.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	for i, one := range titles {
		var keys []string
		if visibleKeys != nil {
			keys = visibleKeys[i]
		}
		if keys == nil {
			keys = make([]string, 0, len(nameables[i]))
			for k := range nameables[i] {
				keys = append(keys, k)
			}
		}
		xstrings.SortStringsNaturalAscending(keys)
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
		headerTitle := xstrings.Truncate(one, 50, true)
		header.SetTitle(headerTitle)
		if strings.HasSuffix(headerTitle, "…") {
			header.Tooltip = newWrappedTooltip(one)
		}
		header.SetLayoutData(&unison.FlexLayoutData{
			HSpan:  2,
			HAlign: align.Fill,
			VAlign: align.Middle,
			HGrab:  true,
		})
		list.AddChild(header)
		for _, k := range keys {
			marker, ok := nameable.NewMarker(k)
			if !ok {
				continue
			}
			label := unison.NewLabel()
			title := xstrings.Truncate(marker.Label, 60, true)
			tooltip := marker.Tooltip
			if strings.HasSuffix(title, "…") {
				tooltip = strings.TrimSpace(marker.Label + "\n\n" + tooltip)
			}
			label.SetTitle(title)
			if tooltip != "" {
				label.Tooltip = newWrappedTooltip(tooltip)
			}
			label.SetLayoutData(&unison.FlexLayoutData{
				HAlign: align.End,
				VAlign: align.Middle,
			})
			label.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: 20}))
			list.AddChild(label)
			list.AddChild(createNameableField(&marker, nameables[i]))
		}
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
	label.SetTitle(i18n.Text("Provide substitutions:"))
	panel.AddChild(label)
	panel.AddChild(scroll)
	return unison.QuestionDialogWithPanel(panel) == unison.ModalResponseOK
}

// createNameableField builds the widget used to edit the replacement value for the marker.
//
// marker comes from nameable.NewMarker, called by our caller with its ok result already checked -- an old-format,
// no-pipe key gets a synthesized marker (see NewMarker) rather than a plain text field, so every nameable, old
// format or new, gets the same NewComboField shell: its options (if any) listed, an "empty" entry when AllowEmpty is
// set (which every synthesized legacy marker has), and free typing enabled when FreeForm is set (also always true
// for a synthesized marker, matching old-format markers' traditional unrestricted typing). A "not set" entry is
// always included too -- for now; that's this function's call, not NewComboField's, since NewComboField itself has
// no opinion on whether "not set" should be offered (see its docs) -- and it's now the only way to clear a
// substitution, there's no separate "clear" button in the dialog.
func createNameableField(marker *nameable.Marker, m map[string]string) unison.Paneler {
	var initial *string
	if v, ok := m[marker.Key()]; ok {
		initial = &v
	}
	options := make([]*string, 0, len(marker.Options)+2)
	options = append(options, nil)
	if marker.AllowEmpty {
		options = append(options, new(string))
	}
	for _, one := range marker.Options {
		options = append(options, &one)
	}

	apply := func(value *string) {
		if value == nil {
			delete(m, marker.Key())
			return
		}
		m[marker.Key()] = *value
	}

	if marker.FreeForm {
		field := unison.NewComboField(options, initial, apply)

		// Use a fixed string to set the lower bounds on the ComboField width
		minWidth := field.MinimumTextWidth
		field.SetMinimumTextWidthUsing("A reasonably wide string")
		field.MinimumTextWidth = max(field.MinimumTextWidth, minWidth)

		return field
	}

	popup := unison.NewPopupMenu[*string]()
	var selected *string
	for _, one := range options {
		popup.AddItem(one)
		if one != nil && initial != nil && *one == *initial {
			selected = one
		}
	}
	popup.Select(selected)
	popup.ItemRendererCallback = func(item *string) string {
		switch {
		case item == nil:
			return i18n.Text("«not set»")
		case *item == "":
			return i18n.Text("«empty»")
		default:
			return *item
		}
	}
	popup.SelectionChangedCallback = func(p *unison.PopupMenu[*string]) {
		if item, ok := p.Selected(); ok {
			apply(item)
		}
	}
	return popup
}
