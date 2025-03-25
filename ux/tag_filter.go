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

	"github.com/richardwilkes/toolbox/collection/dict"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// TagProvider defines the methods required for the tag filter popup.
type TagProvider interface {
	AllTags() []string
	ApplyFilter(tags []string)
}

// NewTagFilterPopup creates a new tag filter popup.
func NewTagFilterPopup(tagProvider TagProvider) *unison.PopupMenu[string] {
	p := unison.NewPopupMenu[string]()
	p.WillShowMenuCallback = func(_ *unison.PopupMenu[string]) {
		var selection []string
		for _, i := range p.SelectedIndexes() {
			if i != 0 {
				if item, ok := p.ItemAt(i); ok {
					selection = append(selection, item)
				}
			}
		}
		p.RemoveAllItems()
		p.AddItem(i18n.Text("Any Tag"))
		tags := tagProvider.AllTags()
		if len(tags) != 0 {
			p.AddSeparator()
			for _, tag := range tags {
				p.AddItem(tag)
			}
		}
		if len(selection) != 0 {
			p.Select(selection...)
		} else {
			p.SelectIndex(0)
		}
	}
	p.ChoiceMadeCallback = func(popup *unison.PopupMenu[string], index int, _ string) {
		simple := index == 0
		if !simple {
			modifiers := popup.Window().CurrentKeyModifiers()
			simple = !modifiers.ShiftDown() && !modifiers.OSMenuCmdModifierDown()
		}
		if simple {
			popup.SelectIndex(index)
		} else {
			m := make(map[int]bool)
			wasSelected := false
			for _, i := range popup.SelectedIndexes() {
				if i != 0 {
					if index == i {
						wasSelected = true
					} else {
						m[i] = true
					}
				}
			}
			if !wasSelected {
				m[index] = true
			}
			if len(m) == 0 {
				popup.SelectIndex(0)
			} else {
				popup.SelectIndex(dict.Keys(m)...)
			}
		}
	}
	tagFilterTooltip := i18n.Text("Tag Filter")
	baseTooltip := fmt.Sprintf(i18n.Text("Shift-Click or %v-Click to select more than one"), unison.OSMenuCmdModifier())
	p.Tooltip = newWrappedTooltipWithSecondaryText(tagFilterTooltip, baseTooltip)
	p.SelectionChangedCallback = func(_ *unison.PopupMenu[string]) {
		tags := SelectedTags(p)
		var extra string
		if len(tags) != 0 {
			extra = i18n.Text("\n\nRequires these tags:\n● ") + strings.Join(tags, "\n● ")
		}
		p.Tooltip = newWrappedTooltipWithSecondaryText(tagFilterTooltip, baseTooltip+extra)
		tagProvider.ApplyFilter(tags)
	}
	p.WillShowMenuCallback(p)
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	return p
}

// SelectedTags returns the set of tags currently selected in a tag filter popup.
func SelectedTags(popup *unison.PopupMenu[string]) []string {
	indexes := popup.SelectedIndexes()
	tags := make([]string, 0, len(indexes))
	set := make(map[string]bool)
	for _, i := range indexes {
		if i != 0 {
			if tag, ok := popup.ItemAt(i); ok && !set[tag] {
				set[tag] = true
				tags = append(tags, tag)
			}
		}
	}
	return tags
}
