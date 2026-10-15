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
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/check"
)

// showMergeListFiltersDialog asks for the name of the filter that merges the given ones, which are for the list type
// with the key, and whether it requires all of them or any one. It reports whether the dialog was accepted.
func showMergeListFiltersDialog(key string, filters []*gurps.ListFilter) (name string, all, ok bool) {
	names := make([]string, len(filters))
	for i, f := range filters {
		names[i] = f.Name
	}
	name = gurps.GlobalSettings().UniqueListFilterName(key, strings.Join(names, " + "))
	all = true

	content := unison.NewPanel()
	content.SetLayout(&unison.FlexLayout{Columns: 2, HSpacing: unison.StdHSpacing, VSpacing: unison.StdVSpacing})
	content.AddChild(NewFieldLeadingLabel(i18n.Text("Name"), false))
	nameField := NewStringField(nil, "", i18n.Text("Name"), func() string { return name },
		func(s string) { name = s })
	nameField.SetMinimumTextWidthUsing(strings.Repeat("M", 20))
	nameField.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	content.AddChild(nameField)

	content.AddChild(NewFieldLeadingLabel(i18n.Text("Match"), false))
	modes := unison.NewPopupMenu[string]()
	modes.AddItem(filterAllOf.String(), filterAnyOf.String())
	modes.SelectIndex(0)
	modes.SelectionChangedCallback = func(popup *unison.PopupMenu[string]) { all = popup.SelectedIndex() == 0 }
	content.AddChild(modes)

	content.AddChild(unison.NewPanel())
	sources := NewWrappingLabel()
	sources.SetTitle(strings.Join(names, i18n.Text(", ")))
	content.AddChild(sources)

	dialog, err := unison.NewDialog(nil, nil, content,
		[]*unison.DialogButtonInfo{unison.NewCancelButtonInfo(), unison.NewOKButtonInfoWithTitle(i18n.Text("Merge"))})
	if err != nil {
		reportUIError(i18n.Text("Unable to create the merge dialog"), err)
		return "", false, false
	}
	dialog.Window().SetTitle(i18n.Text("Merge Filters"))
	nameField.ValidateCallback = func() bool {
		trimmed := strings.TrimSpace(name)
		valid := trimmed != "" && !gurps.GlobalSettings().ListFilterNameInUse(key, trimmed, nil)
		dialog.Button(unison.ModalResponseOK).SetEnabled(valid)
		return valid
	}
	nameField.Validate()
	nameField.RequestFocus()
	nameField.SelectAll()
	if dialog.RunModal() != unison.ModalResponseOK {
		return "", false, false
	}
	return strings.TrimSpace(name), all, true
}

// importedListFilter is a filter chosen for import, along with the list type it is for.
type importedListFilter struct {
	key    string
	filter *gurps.ListFilter
}

// showImportListFiltersDialog shows the filters in the set, read from source, grouped by list type, and returns the
// ones the user leaves checked. Every filter starts checked but those already saved with the same name and conditions.
// Filters with no conditions, and those for list types this version of GCS doesn't know, are left out.
func showImportListFiltersDialog(source string, set map[string][]*gurps.ListFilter) []importedListFilter {
	s := gurps.GlobalSettings()
	list := unison.NewPanel()
	list.SetLayout(&unison.FlexLayout{Columns: 1, VSpacing: unison.StdVSpacing})
	list.SetBorder(unison.NewEmptyBorder(unison.StdInsets()))
	var boxes []*unison.CheckBox
	var offered []importedListFilter
	for _, kind := range listFilterKinds() {
		filters := slices.DeleteFunc(slices.Clone(set[kind.key]), func(f *gurps.ListFilter) bool { return f.Root.Empty() })
		if len(filters) == 0 {
			continue
		}
		heading := unison.NewLabel()
		heading.Font = unison.EmphasizedSystemFont
		heading.SetTitle(kind.title)
		list.AddChild(heading)
		for _, f := range filters {
			box := unison.NewCheckBox()
			box.SetTitle(f.Name)
			box.State = check.On
			list.AddChild(box)
			var note string
			switch {
			case s.HasIdenticalListFilter(kind.key, f):
				box.State = check.Off
				note = i18n.Text("Already saved with the same conditions")
			case s.ListFilterNameInUse(kind.key, f.Name, nil):
				note = i18n.Text("A filter with this name is already saved, so a number will be added to this one")
			}
			if note != "" {
				hint := NewFieldTrailingHint(box, note, true)
				hint.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: unison.StdHSpacing * 4}))
				list.AddChild(hint)
			}
			boxes = append(boxes, box)
			offered = append(offered, importedListFilter{key: kind.key, filter: f})
		}
	}
	if len(offered) == 0 {
		unison.ErrorDialogWithMessage(i18n.Text("Nothing to import"),
			i18n.Text("%s holds no saved filters this version of GCS can use.", source))
		return nil
	}

	content := unison.NewPanel()
	content.SetLayout(&unison.FlexLayout{Columns: 1, VSpacing: unison.StdVSpacing})
	from := NewWrappingLabel()
	from.SetTitle(i18n.Text("From %s", source))
	content.AddChild(from)
	scroll := unison.NewScrollPanel()
	scroll.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.NewUniformInsets(1), false))
	scroll.SetContent(list, behavior.Fill, behavior.Fill)
	scroll.SetLayoutData(&unison.FlexLayoutData{
		MinSize: geom.NewSize(400, 300),
		HAlign:  align.Fill,
		VAlign:  align.Fill,
		HGrab:   true,
		VGrab:   true,
	})
	content.AddChild(scroll)

	dialog, err := unison.NewDialog(nil, nil, content,
		[]*unison.DialogButtonInfo{unison.NewCancelButtonInfo(), unison.NewOKButtonInfoWithTitle(i18n.Text("Import"))})
	if err != nil {
		reportUIError(i18n.Text("Unable to create the import dialog"), err)
		return nil
	}
	dialog.Window().SetTitle(i18n.Text("Import Filters"))
	if dialog.RunModal() != unison.ModalResponseOK {
		return nil
	}
	var chosen []importedListFilter
	for i, box := range boxes {
		if box.State == check.On {
			chosen = append(chosen, offered[i])
		}
	}
	return chosen
}
