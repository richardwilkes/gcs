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
	"maps"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// The nameables prompt is held in a variable so that tests can substitute a non-interactive implementation.
var promptForNameables = ShowNameablesDialog

// ProcessNameables processes the rows and their children for any nameables. Returns false if the user canceled the
// prompt, in which case the caller is expected to abandon the whole operation the prompt was part of.
func ProcessNameables[T gurps.Node[T]](rows []T) bool {
	return ProcessNameableGroups([]NameableGroup[T]{{Rows: rows}})
}

// NameableGroup is a set of rows whose entries in the nameables prompt share a label. An entry is normally titled with
// the row's own name; a non-empty label makes the title "label: name", which is what tells apart otherwise identical
// rows -- the copies of one modifier that a single drop attached to several traits, say, each needing an answer of
// their own.
type NameableGroup[T gurps.Node[T]] struct {
	Label string
	Rows  []T
	// SharedReplacements indicates the rows keep their replacements in one place, as the modifiers of one trait or
	// piece of equipment do, so a key used by more than one row is asked about once, under the first of them, and that
	// answer is applied to all of them.
	SharedReplacements bool
}

// sharedNameableKey records a key used by the entry at 'entry' but asked about under the entry at 'from', whose answer
// it takes (see NameableGroup.SharedReplacements).
type sharedNameableKey struct {
	entry int
	from  int
	key   string
}

// ProcessNameableGroups processes the rows of each group and their children for any nameables, putting up one prompt
// that covers all of the groups. Nothing is rebuilt or reported here; the caller does that once the answers are in.
// Returns false if the user canceled the prompt, in which case the caller is expected to abandon the whole operation
// the prompt was part of.
func ProcessNameableGroups[T gurps.Node[T]](groups []NameableGroup[T]) bool {
	var data []T
	var titles []string
	var nameables []map[string]string
	var visibleKeys [][]string
	var shared []sharedNameableKey
	for _, group := range groups {
		// For a group with shared replacements, the entry each key was first asked about under.
		askedUnder := make(map[string]int)
		for _, row := range group.Rows {
			gurps.Traverse(func(row T) bool {
				m := make(map[string]string)
				row.FillWithNameableKeys(m, nil)
				if len(m) == 0 {
					return false
				}
				var keys []string
				if gurps.IsNodePreconfigured(row) {
					// Only prompt for keys with no replacement recorded; the rest were already resolved.
					if keys = missingNameableKeys(row, m); len(keys) == 0 {
						return false
					}
				}
				if group.SharedReplacements {
					// A key an earlier row already asks about is left out of this row's fields and takes that row's
					// answer instead, since an untouched second field would put its starting value back over the first
					// answer. A row with nothing left to ask about is left out of the prompt.
					if keys == nil {
						keys = slices.Sorted(maps.Keys(m))
					}
					own := make([]string, 0, len(keys))
					var taken []sharedNameableKey
					for _, k := range keys {
						if from, asked := askedUnder[k]; asked {
							taken = append(taken, sharedNameableKey{entry: len(data), from: from, key: k})
						} else {
							own = append(own, k)
						}
					}
					if len(own) == 0 {
						return false
					}
					shared = append(shared, taken...)
					for _, k := range own {
						askedUnder[k] = len(data)
					}
					keys = own
				}
				title := row.String()
				if group.Label != "" {
					title = group.Label + ": " + title
				}
				data = append(data, row)
				titles = append(titles, title)
				nameables = append(nameables, m)
				visibleKeys = append(visibleKeys, keys) // nil means "show all keys"
				return false
			}, false, false, row)
		}
	}
	if len(data) > 0 {
		if !promptForNameables(titles, nameables, visibleKeys) {
			return false
		}
		for _, one := range shared {
			if v, ok := nameables[one.from][one.key]; ok {
				nameables[one.entry][one.key] = v
			} else {
				delete(nameables[one.entry], one.key) // Cleared under the entry it was asked about, so cleared here too.
			}
		}
		for i, row := range data {
			row.ApplyNameableKeys(nameables[i])
		}
	}
	return true
}

// missingNameableKeys returns the keys of nameables that have no explicit replacement recorded on row.
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
		if headerTitle != one {
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
			if title != marker.Label {
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
	return showListQuestionDialog(i18n.Text("Provide substitutions:"), list)
}

// createNameableField builds the widget used to edit the replacement value for the marker, which comes from
// nameable.NewMarker with its ok result already checked. A FreeForm marker -- which every synthesized old-format
// marker is, matching such markers' traditional unrestricted typing -- gets an editable combo field, while any other
// gets a popup menu limited to the marker's options. A "not set" entry is always offered as well, since it is the only
// way to clear a substitution, and a nameable whose value in m is nameable.Unset starts out on it, regardless of the
// marker's Options and AllowEmpty settings.
func createNameableField(marker *nameable.Marker, m map[string]string) unison.Paneler {
	var initial *string
	if v, ok := m[marker.Key()]; ok && v != nameable.Unset {
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

		field.SetLayoutData(&unison.FlexLayoutData{
			HAlign: align.Fill,
			VAlign: align.Middle,
			HGrab:  true,
		})
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
	if initial != nil && selected == nil {
		// The stored value doesn't match any of the marker's current options -- a legacy value, or the options
		// changed since it was set. Show it rather than displaying "«not set»" while silently leaving the old
		// value in place if the user presses OK without touching this field.
		popup.AddItem(initial)
		selected = initial
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
	popup.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
		HGrab:  true,
	})
	return popup
}
