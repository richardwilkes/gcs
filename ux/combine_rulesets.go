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
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xfilepath"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/check"
)

// combinedRulesetFolder is one selectable folder within a library.
type combinedRulesetFolder struct {
	lib    *gurps.Library
	folder string
}

func (c *combinedRulesetFolder) String() string {
	return c.lib.Data().Title + ": " + c.folder
}

// prioritizedRulesetFolder wraps a folder with its position in the priority order, so that the selected list shows the
// priority each folder will combine with.
type prioritizedRulesetFolder struct {
	*combinedRulesetFolder
	position int
}

func (p prioritizedRulesetFolder) String() string {
	return fmt.Sprintf("%d. %s", p.position, p.combinedRulesetFolder)
}

// combineRulesets puts up the dialog for combining the like-for-like data files of several library folders into a
// single set of files within the user library, then performs the combination.
func combineRulesets() {
	available := collectCombinedRulesetFolders()
	if len(available) == 0 {
		unison.WarningDialogWithMessage(i18n.Text("No library folders found"),
			i18n.Text("There are no folders in your libraries to combine."))
		return
	}
	var selected []*combinedRulesetFolder

	nameField := unison.NewField()
	nameField.SetLayoutData(&unison.FlexLayoutData{
		MinSize: geom.Size{Width: 300},
		HAlign:  align.Fill,
		HGrab:   true,
	})

	subfoldersCheckBox := unison.NewCheckBox()
	subfoldersCheckBox.SetTitle(i18n.Text("Include files found in subfolders"))

	availableList := unison.NewList[*combinedRulesetFolder]()
	availableList.SetAllowMultipleSelection(true)
	availableList.Append(available...)
	selectedList := unison.NewList[prioritizedRulesetFolder]()
	selectedList.SetAllowMultipleSelection(false)

	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
	})

	intro := unison.NewLabel()
	intro.SetTitle(i18n.Text("Combine the data files of the chosen folders into a new folder in the User Library."))
	panel.AddChild(intro)

	nameRow := unison.NewPanel()
	nameRow.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	nameRow.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	nameLabel := unison.NewLabel()
	nameLabel.SetTitle(i18n.Text("Name:"))
	nameRow.AddChild(nameLabel)
	nameRow.AddChild(nameField)
	panel.AddChild(nameRow)

	lists := unison.NewPanel()
	lists.SetLayout(&unison.FlexLayout{
		Columns:  3,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
	})
	lists.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	panel.AddChild(lists)

	lists.AddChild(newCombineRulesetsListHeader(i18n.Text("Available Folders")))
	lists.AddChild(unison.NewPanel()) // Spacer above the add/remove buttons
	lists.AddChild(newCombineRulesetsListHeader(i18n.Text("Combine (highest priority first)")))

	lists.AddChild(newCombineRulesetsListScroller(availableList))

	moveButtons := unison.NewPanel()
	moveButtons.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: unison.StdVSpacing,
	})
	moveButtons.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Middle})
	addButton := unison.NewButton()
	addButton.SetTitle(i18n.Text("Add →"))
	addButton.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill})
	moveButtons.AddChild(addButton)
	removeButton := unison.NewButton()
	removeButton.SetTitle(i18n.Text("← Remove"))
	removeButton.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill})
	moveButtons.AddChild(removeButton)
	lists.AddChild(moveButtons)

	selectedColumn := unison.NewPanel()
	selectedColumn.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: unison.StdVSpacing,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
	})
	selectedColumn.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	selectedColumn.AddChild(newCombineRulesetsListScroller(selectedList))
	orderButtons := unison.NewPanel()
	orderButtons.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
	})
	upButton := unison.NewButton()
	upButton.SetTitle(i18n.Text("Raise Priority"))
	orderButtons.AddChild(upButton)
	downButton := unison.NewButton()
	downButton.SetTitle(i18n.Text("Lower Priority"))
	orderButtons.AddChild(downButton)
	selectedColumn.AddChild(orderButtons)
	lists.AddChild(selectedColumn)

	panel.AddChild(subfoldersCheckBox)

	note := unison.NewLabel()
	note.SetTitle(i18n.Text("Items found in more than one folder keep the stats of the highest priority folder and gather every folder's page references."))
	panel.AddChild(note)

	dialog, err := unison.NewDialog(unison.DefaultDialogTheme.QuestionIcon, unison.DefaultDialogTheme.QuestionIconInk,
		panel, []*unison.DialogButtonInfo{
			unison.NewCancelButtonInfo(),
			unison.NewOKButtonInfoWithTitle(i18n.Text("Create")),
		})
	if err != nil {
		errs.Log(err)
		return
	}
	okButton := dialog.Button(unison.ModalResponseOK)

	refreshSelected := func(newSelection int) {
		selectedList.Clear()
		for i, one := range selected {
			selectedList.Append(prioritizedRulesetFolder{combinedRulesetFolder: one, position: i + 1})
		}
		if newSelection >= 0 && newSelection < len(selected) {
			selectedList.Select(false, newSelection)
		}
		selectedList.MarkForRedraw()
		okButton.SetEnabled(strings.TrimSpace(nameField.Text()) != "" && len(selected) != 0)
	}
	nameField.ModifiedCallback = func(_, _ *unison.FieldState) {
		okButton.SetEnabled(strings.TrimSpace(nameField.Text()) != "" && len(selected) != 0)
	}
	addSelection := func() {
		i := availableList.Selection.FirstSet()
		for i != -1 {
			one := availableList.DataAtIndex(i)
			if !slices.Contains(selected, one) {
				selected = append(selected, one)
			}
			i = availableList.Selection.NextSet(i + 1)
		}
		refreshSelected(-1)
	}
	removeSelection := func() {
		if i := selectedList.Selection.FirstSet(); i != -1 && i < len(selected) {
			selected = slices.Delete(selected, i, i+1)
			refreshSelected(-1)
		}
	}
	addButton.ClickCallback = addSelection
	availableList.DoubleClickCallback = addSelection
	removeButton.ClickCallback = removeSelection
	selectedList.DoubleClickCallback = removeSelection
	upButton.ClickCallback = func() {
		if i := selectedList.Selection.FirstSet(); i > 0 && i < len(selected) {
			selected[i-1], selected[i] = selected[i], selected[i-1]
			refreshSelected(i - 1)
		}
	}
	downButton.ClickCallback = func() {
		if i := selectedList.Selection.FirstSet(); i != -1 && i < len(selected)-1 {
			selected[i], selected[i+1] = selected[i+1], selected[i]
			refreshSelected(i + 1)
		}
	}
	okButton.SetEnabled(false)

	if dialog.RunModal() != unison.ModalResponseOK {
		return
	}
	name := xfilepath.SanitizeName(strings.TrimSpace(nameField.Text()))
	if name == "" {
		return
	}
	sources := make([]gurps.CombinedLibrarySource, 0, len(selected))
	for _, one := range selected {
		sources = append(sources, gurps.CombinedLibrarySource{Library: one.lib, Folder: one.folder})
	}
	performRulesetCombination(gurps.CombinedLibraryOptions{
		Name:              name,
		Sources:           sources,
		IncludeSubfolders: subfoldersCheckBox.State == check.On,
	})
}

// performRulesetCombination runs the combination for the already-gathered options, writing the result into a folder in
// the user library named for the combination.
func performRulesetCombination(opts gurps.CombinedLibraryOptions) {
	unableMsg := i18n.Text("Unable to create the combined ruleset")
	destDir := filepath.Join(gurps.GlobalSettings().Libraries().User().Path(), opts.Name)
	if xos.IsDir(destDir) {
		if unison.YesNoDialog(fmt.Sprintf(i18n.Text(`A folder named "%s" already exists in the User Library.`), opts.Name),
			i18n.Text("Any combined data files it contains with the same names will be replaced. Continue?")) != unison.ModalResponseOK {
			return
		}
	}
	// A combined file that is already open would be overwritten out from under its view, so close such views first,
	// the same way the rules lookup download does.
	for _, target := range gurps.CombinedLibraryFilePaths(opts.Name, destDir) {
		target = filepath.Clean(target)
		for _, one := range AllDockables() {
			if tc, ok := one.(unison.TabCloser); ok {
				var fbd FileBackedDockable
				if fbd, ok = one.(FileBackedDockable); ok {
					if filepath.Clean(fbd.BackingFilePath()) == target {
						if !tc.MayAttemptClose() || !tc.AttemptClose() {
							unison.WarningDialogWithMessage(i18n.Text("Combination canceled"),
								i18n.Text("Cannot replace a combined file while it is open."))
							return
						}
						break
					}
				}
			}
		}
	}
	if _, err := gurps.CreateCombinedLibrary(opts, destDir); err != nil {
		Workspace.ErrorHandler(unableMsg, err)
		return
	}
	Workspace.Navigator.EventuallyReload()
}

// collectCombinedRulesetFolders gathers the folders that may be combined: every top-level folder of every library.
func collectCombinedRulesetFolders() []*combinedRulesetFolder {
	var folders []*combinedRulesetFolder
	for _, lib := range gurps.GlobalSettings().Libraries().List() {
		entries, err := os.ReadDir(lib.Path())
		if err != nil {
			errs.Log(err, "dir", lib.Path())
			continue
		}
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				names = append(names, entry.Name())
			}
		}
		xstrings.SortStringsNaturalAscending(names)
		for _, name := range names {
			folders = append(folders, &combinedRulesetFolder{lib: lib, folder: name})
		}
	}
	return folders
}

func newCombineRulesetsListHeader(title string) *unison.Label {
	header := unison.NewLabel()
	header.SetTitle(title)
	return header
}

func newCombineRulesetsListScroller(content unison.Paneler) *unison.ScrollPanel {
	scroll := unison.NewScrollPanel()
	scroll.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.NewUniformInsets(1), false))
	scroll.SetContent(content, behavior.Fill, behavior.Fill)
	scroll.BackgroundInk = unison.ThemeSurface
	scroll.SetLayoutData(&unison.FlexLayoutData{
		MinSize: geom.Size{Width: 320, Height: 320},
		HAlign:  align.Fill,
		VAlign:  align.Fill,
		HGrab:   true,
		VGrab:   true,
	})
	return scroll
}
