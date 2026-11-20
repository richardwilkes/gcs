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
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// Seams so that the source set commands can be exercised without a modal dialog.
var (
	showLibrarySourceSetEditor      = showLibrarySourceSetDialog
	confirmLibrarySourceSetDeletion = func(name string) bool {
		return unison.QuestionDialog(i18n.Text("Delete the set '%s'?", name), "") == unison.ModalResponseOK
	}
)

// librarySourceSetChoice is what an item of the source set popup stands for: a list type's set of every file (a nil
// set), the view's unsaved set, or one of the saved sets. Commands, headers and separators have no key.
type librarySourceSetChoice struct {
	set   *gurps.LibrarySourceSet
	key   string
	adHoc bool
}

// librarySourceSetPopup is the toolbar popup of a library filter view that chooses the source set the view shows, and
// with it the list type, and creates, edits, saves and deletes the saved sets. After the commands, it lists every list
// type under a header of its own: its set of every file, the view's unsaved set when it is of that type, and the saved
// sets of that type. The items are rebuilt each time the popup is about to be shown, since the saved sets may have been
// changed in the meantime. A choice never moves the popup's selection itself: a set is selected only when the items are
// rebuilt for the set the view then shows.
type librarySourceSetPopup struct {
	popup       *unison.PopupMenu[string]
	button      *filterPopupButton
	dockable    *LibraryFilterDockable
	choices     []librarySourceSetChoice // Parallel to the popup's items
	newIndex    int
	editIndex   int
	saveAsIndex int
	deleteIndex int
}

// newLibrarySourceSetPopup creates the source set popup of the view, which compacts to the given icon.
func newLibrarySourceSetPopup(d *LibraryFilterDockable, icon *unison.SVG) *librarySourceSetPopup {
	p := &librarySourceSetPopup{popup: unison.NewPopupMenu[string](), dockable: d}
	p.popup.WillShowMenuCallback = func(_ *unison.PopupMenu[string]) { p.rebuildItems() }
	p.popup.ChoiceMadeCallback = func(_ *unison.PopupMenu[string], index int, _ string) {
		switch index {
		case p.newIndex:
			p.newSet()
		case p.editIndex:
			p.editSet()
		case p.saveAsIndex:
			p.saveSetAs()
		case p.deleteIndex:
			p.deleteSet()
		default:
			p.choose(index)
		}
	}
	p.button = newFilterPopupButton(p.popup, icon, func() bool { return d.set != nil })
	// The popup sits among the toolbar's icons, with no label of its own to name it by.
	p.popup.Accessibility.Name = i18n.Text("Source Set")
	p.popup.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	p.rebuildItems()
	return p
}

// rebuildItems replaces the popup's items with the source sets as they now stand, selecting the view's. Edit and
// Delete are only offered for a saved set.
func (p *librarySourceSetPopup) rebuildItems() {
	d := p.dockable
	p.popup.RemoveAllItems()
	p.choices = nil
	p.newIndex = p.addCommand(i18n.Text("New Set…"))
	p.editIndex = p.addCommand(i18n.Text("Edit Set…"))
	p.saveAsIndex = p.addCommand(i18n.Text("Save Set As…"))
	p.deleteIndex = p.addCommand(i18n.Text("Delete Set…"))
	saved := d.set != nil && !d.adHoc
	p.popup.SetItemEnabledAt(p.editIndex, saved)
	p.popup.SetItemEnabledAt(p.deleteIndex, saved)
	selected := -1
	for _, kind := range libraryFilterKinds() {
		p.popup.AddSeparator()
		p.choices = append(p.choices, librarySourceSetChoice{})
		p.popup.AddDisabledItem(kind.title)
		p.choices = append(p.choices, librarySourceSetChoice{})
		current := kind.key == d.kind.key
		if i := p.addChoice(libraryFilterSetTitle(kind, nil, false),
			librarySourceSetChoice{key: kind.key}); current && d.set == nil {
			selected = i
		}
		if current && d.adHoc && d.set != nil {
			selected = p.addChoice(libraryFilterSetTitle(kind, d.set, true),
				librarySourceSetChoice{set: d.set, key: kind.key, adHoc: true})
		}
		for _, set := range gurps.GlobalSettings().LibrarySourceSetsFor(kind.key) {
			// The view's set is found by identity rather than by name, since a set may bear the name of another item.
			if i := p.addChoice(set.Name, librarySourceSetChoice{set: set, key: kind.key}); current && saved &&
				set == d.set {
				selected = i
			}
		}
	}
	if selected != -1 {
		p.popup.SelectIndex(selected)
	}
	p.syncTooltip()
	p.popup.MarkForRedraw()
}

func (p *librarySourceSetPopup) addCommand(title string) int {
	return p.addChoice(title, librarySourceSetChoice{})
}

func (p *librarySourceSetPopup) addChoice(title string, choice librarySourceSetChoice) int {
	p.popup.AddItem(title)
	p.choices = append(p.choices, choice)
	return len(p.choices) - 1
}

// choose has the view show the set the item at the index stands for. The view's content is replaced when the set is of
// another list type, popup and all, so that is left until the popup's menu has finished with the choice.
func (p *librarySourceSetPopup) choose(index int) {
	if index < 0 || index >= len(p.choices) || p.choices[index].key == "" {
		return
	}
	choice := p.choices[index]
	d := p.dockable
	unison.InvokeTask(func() {
		if !d.closed {
			d.chooseSet(choice.key, choice.set, choice.adHoc)
		}
	})
}

// refresh brings the popup, and the view, into line with the source sets after they were changed through another
// view's popup: a view whose saved set was deleted keeps showing the same sources as an unsaved set, and one whose set
// was changed shows it as it now is, switching list type with it if need be, re-reading its files if they are no longer
// the same ones, and retitling itself. Every other view only has its items rebuilt.
func (p *librarySourceSetPopup) refresh() {
	d := p.dockable
	if d.set == nil || d.adHoc {
		p.rebuildItems()
		return
	}
	if key, ok := gurps.GlobalSettings().LibrarySourceSetKey(d.set); ok {
		d.showSet(key, d.set, false)
	} else {
		d.showSet(d.kind.key, d.set.Clone(), true)
	}
}

// newSet puts up the editor for a new set of the view's list type and, if it is accepted, saves the set and has the
// view show it.
func (p *librarySourceSetPopup) newSet() {
	key := p.dockable.kind.key
	set := &gurps.LibrarySourceSet{}
	if showLibrarySourceSetEditor(i18n.Text("New Set"), &key, set, nil) {
		gurps.GlobalSettings().AddLibrarySourceSet(key, set)
		p.setsChanged(key, set)
	}
}

// editSet puts up the editor for the view's saved set. A clone is edited so that a cancel leaves the set untouched.
// When the edit is accepted, the clone's contents are copied into the saved set rather than replacing it, so that every
// other view showing the set keeps it and sees the change. A set whose list type was changed moves to that type's sets,
// and is no longer recorded as the one last used for its old type.
func (p *librarySourceSetPopup) editSet() {
	d := p.dockable
	current := d.set
	if current == nil || d.adHoc {
		return
	}
	settings := gurps.GlobalSettings()
	oldKey, ok := settings.LibrarySourceSetKey(current)
	if !ok {
		return
	}
	key := oldKey
	oldName := current.Name
	clone := current.Clone()
	if !showLibrarySourceSetEditor(i18n.Text("Edit Set"), &key, clone, current) {
		return
	}
	*current = *clone
	if key != oldKey {
		settings.RemoveLibrarySourceSet(oldKey, current)
		settings.AddLibrarySourceSet(key, current)
		forgetLastLibrarySourceSet(oldKey, oldName)
	} else {
		settings.ResortLibrarySourceSets(key) // The name may have changed.
	}
	// The view records the set, under its new name, as the one last used for its type.
	p.setsChanged(key, current)
}

// saveSetAs puts up the editor for a new set holding the sources the view shows, under a name no set of its list type
// has yet, built-in ones included, and, if it is accepted, saves the set and has the view show it in place of what it
// showed.
func (p *librarySourceSetPopup) saveSetAs() {
	d := p.dockable
	key := d.kind.key
	set := &gurps.LibrarySourceSet{
		Name: gurps.GlobalSettings().UniqueLibrarySourceSetName(key, libraryFilterSetTitle(d.kind, d.set, d.adHoc),
			builtInLibrarySourceSetTitles(d.kind)...),
		Entries: d.shownEntries(),
	}
	if showLibrarySourceSetEditor(i18n.Text("Save Set As"), &key, set, nil) {
		gurps.GlobalSettings().AddLibrarySourceSet(key, set)
		p.setsChanged(key, set)
	}
}

// deleteSet removes the view's saved set, once the user has confirmed it, and has the view show every file of its list
// type instead.
func (p *librarySourceSetPopup) deleteSet() {
	d := p.dockable
	current := d.set
	if current == nil || d.adHoc {
		return
	}
	settings := gurps.GlobalSettings()
	key, ok := settings.LibrarySourceSetKey(current)
	if !ok || !confirmLibrarySourceSetDeletion(current.Name) {
		return
	}
	settings.RemoveLibrarySourceSet(key, current)
	// The view records every file of the type as what was last used for it.
	p.setsChanged(key, nil)
}

// setsChanged finishes a change made to the saved sets through this popup: the view shows the set of the list type with
// the key, recording it as the one last used, the settings are saved, and every other view is told to catch up. The
// view's content, this popup with it, is replaced when the list type changes, which leaves the popup only its view.
func (p *librarySourceSetPopup) setsChanged(key string, set *gurps.LibrarySourceSet) {
	p.dockable.chooseSet(key, set, false)
	saveGlobalSettings()
	notifyLibrarySourceSetsChanged(p)
}

// notifyLibrarySourceSetsChanged tells every library filter view that the saved source sets have changed, so that each
// can bring itself into line (see librarySourceSetPopup.refresh). The view of the popup the change was made through is
// passed as source, since it has already done so.
func notifyLibrarySourceSetsChanged(source *librarySourceSetPopup) {
	for _, d := range allLibraryFilters() {
		if source == nil || d != source.dockable {
			d.content.sourceSetPopup().refresh()
		}
	}
}

// forgetLastLibrarySourceSet records every file of the list type with the key as what was last used for it, in place of
// the set with the name, ignoring case and surrounding whitespace, which is no longer one of that type's sets.
func forgetLastLibrarySourceSet(key, name string) {
	last := gurps.GlobalSettings().LibraryFilter.LastSets
	if recorded, ok := last[key]; ok && strings.EqualFold(strings.TrimSpace(recorded), strings.TrimSpace(name)) {
		delete(last, key)
	}
}

// syncTooltip puts a set name the button cuts short at the head of the tooltip, so that it can still be read in full.
func (p *librarySourceSetPopup) syncTooltip() {
	detail := i18n.Text("Choose the library files to show, or create, edit, save and delete sets of them")
	if name, cut := p.button.textCut(); cut {
		detail = name + "\n\n" + detail
	}
	p.popup.Tooltip = newWrappedTooltipWithSecondaryText(p.popup.Accessibility.Name, detail)
}
