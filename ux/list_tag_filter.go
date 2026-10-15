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

	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/mod"
)

// tagFilterModifiers returns the modifier keys held while a tag is chosen. Tests replace it.
var tagFilterModifiers = func(popup *unison.PopupMenu[string]) mod.Modifiers {
	if wnd := popup.Window(); wnd != nil {
		return wnd.CurrentKeyModifiers()
	}
	return 0
}

// listTagFilterPopupSpec is what the tag filter popup needs from the dockable that shows it.
type listTagFilterPopupSpec struct {
	// tags returns the tags of the items the saved filter in force lets through.
	tags func() []string
	// changed is called after the chosen tags change.
	changed func()
	// save turns the chosen tags into a saved filter.
	save func()
}

// The values of the popup's own entries. A tag's value is the tag in lower case, and a tag holding a NUL is never
// offered, so no tag can be taken for one of them.
const (
	anyTagValue     = "\x00any tag\x00"
	saveFilterValue = "\x00save filter\x00"
)

// listTagFilterPopup is the toolbar popup that narrows a list to the items carrying every tag chosen in it. The choice
// belongs to the dockable alone and is never saved, though its last entry turns it into a saved filter.
type listTagFilterPopup struct {
	popup *unison.PopupMenu[string]
	spec  listTagFilterPopupSpec
	// selected holds the values of the chosen tags, in the order chosen.
	selected []string
	// titles maps the value of each tag on offer to the tag as the list spells it, which is what its item shows.
	titles map[string]string
}

// newListTagFilterPopup creates the tag filter popup.
func newListTagFilterPopup(spec listTagFilterPopupSpec) *listTagFilterPopup {
	p := &listTagFilterPopup{
		popup: unison.NewPopupMenu[string](),
		spec:  spec,
	}
	p.popup.ItemRendererCallback = p.title
	p.popup.WillShowMenuCallback = func(_ *unison.PopupMenu[string]) {
		if p.rebuildItems() {
			p.spec.changed()
		}
	}
	p.popup.ChoiceMadeCallback = func(popup *unison.PopupMenu[string], _ int, value string) {
		switch value {
		case saveFilterValue:
			p.spec.save()
		case anyTagValue:
			p.choose(nil)
		default:
			// A plain click chooses just the one tag, while Shift or the menu command key adds it or takes it off.
			values := []string{value}
			if mods := tagFilterModifiers(popup); mods.ShiftDown() || mods.OSMenuCommandDown() {
				if i := slices.Index(p.selected, value); i != -1 {
					values = slices.Delete(slices.Clone(p.selected), i, i+1)
				} else {
					values = append(slices.Clone(p.selected), value)
				}
			}
			p.choose(values)
		}
	}
	p.popup.Accessibility.Name = i18n.Text("Tag Filter")
	// A tag the button cuts short is already listed in full in the tooltip, among the chosen tags.
	fixPopupWidth(p.popup)
	p.popup.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	p.rebuildItems()
	return p
}

// title returns what the item with the given value shows.
func (p *listTagFilterPopup) title(value string) string {
	switch value {
	case anyTagValue:
		return i18n.Text("Any Tag")
	case saveFilterValue:
		return i18n.Text("Save Filter…")
	default:
		return p.titles[value]
	}
}

// rebuildItems replaces the popup's items with Any Tag, the tags of the items the saved filter lets through, in the
// order ListFilterTags gives them, and the Save Filter command, and selects the chosen tags. A chosen tag no longer on
// offer is dropped, and it reports whether any was, in which case the caller has to filter the list again.
func (p *listTagFilterPopup) rebuildItems() (dropped bool) {
	titles := make(map[string]string)
	var values []string
	for _, tag := range p.spec.tags() {
		if !strings.ContainsRune(tag, 0) {
			value := strings.ToLower(tag)
			titles[value] = tag
			values = append(values, value)
		}
	}
	p.titles = titles
	if kept := slices.DeleteFunc(slices.Clone(p.selected), func(value string) bool {
		_, ok := titles[value]
		return !ok
	}); len(kept) != len(p.selected) {
		p.selected = kept
		dropped = true
	}
	p.popup.RemoveAllItems()
	p.popup.AddItem(anyTagValue)
	if len(values) != 0 {
		p.popup.AddSeparator()
		p.popup.AddItem(values...)
	}
	p.popup.AddSeparator()
	p.popup.AddItem(saveFilterValue)
	p.syncSelection()
	return dropped
}

// chosenTags returns the chosen tags spelled as the list spells them, which is how the filter is given them.
func (p *listTagFilterPopup) chosenTags() []string {
	tags := make([]string, 0, len(p.selected))
	for _, value := range p.selected {
		tags = append(tags, p.titles[value])
	}
	return tags
}

// clear drops the chosen tags without filtering the list again, which the caller does.
func (p *listTagFilterPopup) clear() {
	p.selected = nil
	p.syncSelection()
}

// choose makes the tags with the given values the chosen tags and has the list filtered by them, unless they are the
// ones already chosen.
func (p *listTagFilterPopup) choose(values []string) {
	if slices.Equal(p.selected, values) {
		return
	}
	p.selected = values
	p.syncSelection()
	p.spec.changed()
}

// syncSelection brings the popup's selection, its Save Filter command and its tooltip into line with the chosen tags.
func (p *listTagFilterPopup) syncSelection() {
	if len(p.selected) == 0 {
		p.popup.Select(anyTagValue)
	} else {
		p.popup.Select(p.selected...)
	}
	p.popup.SetItemEnabledAt(p.popup.IndexOfItem(saveFilterValue), len(p.selected) != 0)
	detail := i18n.Text("Shift-Click or %s-Click to select more than one",
		strings.TrimSuffix(mod.OSMenuCommand().String(), "+"))
	if len(p.selected) != 0 {
		detail += "\n\n" + i18n.Text("Requires these tags:") + "\n● " + strings.Join(p.chosenTags(), "\n● ")
	}
	p.popup.Tooltip = newWrappedTooltipWithSecondaryText(p.popup.Accessibility.Name, detail)
}
