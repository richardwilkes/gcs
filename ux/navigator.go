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
	"bytes"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/side"
	"github.com/rjeczalik/notify"
)

const (
	// NavigatorDockKey is the key used to store the Navigator in the top Dock.
	NavigatorDockKey      = "navigator"
	minTextWidthCandidate = "Abcdefghijklmnopqrstuvwxyz0123456789"
)

var (
	_ unison.Dockable = &Navigator{}
	_ KeyedDockable   = &Navigator{}
)

// FileBackedDockable defines methods a Dockable that is based on a file should implement.
type FileBackedDockable interface {
	unison.Dockable
	unison.TabCloser
	BackingFilePath() string
	SetBackingFilePath(p string)
}

// Navigator holds the workspace navigation panel.
type Navigator struct {
	unison.Panel
	toolbar                   *unison.Panel
	buttonRow                 *unison.Panel // The toolbar row the update button is added to when there is an update
	backButton                *unison.Button
	forwardButton             *unison.Button
	searchField               *unison.Field
	matchesLabel              *unison.Label
	deleteButton              *unison.Button
	renameButton              *unison.Button
	newFolderButton           *unison.Button
	downloadLibraryButton     *unison.Button
	libraryReleaseNotesButton *unison.Button
	configLibraryButton       *unison.Button
	favoriteButton            *unison.Button
	appUpdateButton           *unison.Button
	scroll                    *unison.ScrollPanel
	table                     *unison.Table[*NavigatorNode]
	tokens                    []*gurps.MonitorToken
	searchResult              []*NavigatorNode
	deepSearch                map[string]bool
	contentCache              map[string]*contentCacheEntry
	lastBuild                 *contentCacheBuild
	invalidatedPaths          map[string]bool
	appUpdatePulse            appUpdatePulse
	cacheGeneration           atomic.Int64
	searchIndex               int
	prewarmSuspensions        int
	needReload                bool
	prewarmPending            bool
	adjustTableSizePending    bool
}

func newNavigator() *Navigator {
	n := &Navigator{
		toolbar:     unison.NewPanel(),
		scroll:      unison.NewScrollPanel(),
		table:       unison.NewTable(&unison.SimpleTableModel[*NavigatorNode]{}),
		deepSearch:  make(map[string]bool),
		searchIndex: -1,
	}
	n.Self = n

	globalSettings := gurps.GlobalSettings()
	n.mapDeepSearch()

	n.setupToolBar()

	n.table.PreventUserColumnResize = true
	n.table.ShowFirstColumnDivider = false
	n.table.ShowLastColumnDivider = false
	n.table.Columns = make([]unison.ColumnInfo, 1)
	n.needReload = true
	rows := n.populateRows()
	n.needReload = false
	scale := float32(globalSettings.General.NavigatorUIScale) / 100
	n.table.SetScale(geom.NewPoint(scale, scale))
	n.table.SetRootRows(rows)
	n.table.SizeColumnsToFit(true)

	n.scroll.SetContent(n.table, behavior.Fill, behavior.Fill)
	n.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})

	n.SetLayout(&unison.FlexLayout{
		Columns: 1,
		HAlign:  align.Fill,
		VAlign:  align.Fill,
	})
	n.AddChild(n.toolbar)
	n.AddChild(n.scroll)

	n.table.DoubleClickCallback = n.handleSelectionDoubleClick
	gurps.SetNotifyOfLibraryChangeFunc(n.EventuallyReload)
	n.table.MouseDownCallback = n.mouseDown
	n.table.SelectionChangedCallback = n.selectionChanged
	n.table.KeyDownCallback = n.tableKeyDown

	n.InstallCmdHandlers(JumpToSearchFilterItemID,
		func(any) bool { return !n.searchField.Focused() },
		func(any) { n.searchField.RequestFocus() })

	n.selectionChanged()
	// The launch-time update check can finish before the Library Explorer exists, in which case the notification that
	// would have revealed the button has already come and gone.
	n.syncAppUpdateButton()
	n.EventuallyReload() // Without this, the version for libraries is sometimes truncated at initial load
	return n
}

// DockKey implements KeyedDockable.
func (n *Navigator) DockKey() string {
	return NavigatorDockKey
}

func (n *Navigator) mapDeepSearch() {
	n.deepSearch = make(map[string]bool)
	for _, one := range gurps.GlobalSettings().DeepSearch {
		for _, ext := range gurps.FileInfoFor(one).UTI.Extensions {
			n.deepSearch[ext] = true
		}
	}
	n.prewarmContentCache()
}

func (n *Navigator) setupToolBar() {
	helpButton := unison.NewSVGButton(svg.Help)
	helpButton.Tooltip = newWrappedTooltip(i18n.Text("Help"))
	helpButton.ClickCallback = func() { HandleLink(nil, "md:User%20Guide/Library%20Explorer") }

	hierarchyButton := unison.NewSVGButton(svg.Hierarchy)
	hierarchyButton.Tooltip = newWrappedTooltip(i18n.Text("Opens/closes all hierarchical rows"))
	hierarchyButton.ClickCallback = n.toggleHierarchy

	n.deleteButton = unison.NewSVGButton(unison.TrashSVG)
	n.deleteButton.Tooltip = newWrappedTooltip(i18n.Text("Delete"))
	n.deleteButton.ClickCallback = n.deleteSelection

	n.renameButton = unison.NewSVGButton(svg.SignPost)
	n.renameButton.Tooltip = newWrappedTooltip(i18n.Text("Rename"))
	n.renameButton.ClickCallback = n.renameSelection

	n.newFolderButton = unison.NewSVGButton(svg.NewFolder)
	n.newFolderButton.Tooltip = newWrappedTooltip(i18n.Text("New Folder"))
	n.newFolderButton.ClickCallback = n.newFolder

	addLibraryButton := unison.NewSVGButton(unison.CircledAddSVG)
	addLibraryButton.Tooltip = newWrappedTooltip(i18n.Text("Add Library"))
	addLibraryButton.ClickCallback = n.addLibrary

	n.downloadLibraryButton = unison.NewSVGButton(svg.Download)
	n.downloadLibraryButton.Tooltip = newWrappedTooltip(i18n.Text("Update"))
	n.downloadLibraryButton.ClickCallback = n.updateLibrarySelection

	n.libraryReleaseNotesButton = unison.NewSVGButton(svg.ReleaseNotes)
	n.libraryReleaseNotesButton.Tooltip = newWrappedTooltip(i18n.Text("Release Notes"))
	n.libraryReleaseNotesButton.ClickCallback = n.showSelectionReleaseNotes

	n.configLibraryButton = unison.NewSVGButton(svg.Gears)
	n.configLibraryButton.Tooltip = newWrappedTooltip(i18n.Text("Configure"))
	n.configLibraryButton.ClickCallback = n.configureSelection

	n.favoriteButton = unison.NewSVGButton(svg.Star)
	n.favoriteButton.Tooltip = newWrappedTooltip(i18n.Text("Toggle Favorite"))
	n.favoriteButton.ClickCallback = n.favoriteSelection

	n.appUpdateButton = newAppUpdateButton(NotifyOfAppUpdate)
	n.appUpdatePulse.apply = n.applyAppUpdatePulse

	first := unison.NewPanel()
	first.AddChild(NewDefaultInfoPop())
	first.AddChild(helpButton)
	first.AddChild(
		NewScaleField(
			gurps.InitialUIScaleMin,
			gurps.InitialUIScaleMax,
			func() int { return gurps.InitialNavigatorUIScaleDef },
			func() int { return gurps.GlobalSettings().General.NavigatorUIScale },
			func(scale int) { gurps.GlobalSettings().General.NavigatorUIScale = scale },
			nil,
			false,
			false,
			n.scroll,
		),
	)
	first.AddChild(hierarchyButton)
	first.AddChild(NewToolbarSeparator())
	first.AddChild(addLibraryButton)
	first.AddChild(n.downloadLibraryButton)
	first.AddChild(n.libraryReleaseNotesButton)
	first.AddChild(n.configLibraryButton)
	first.AddChild(NewToolbarSeparator())
	first.AddChild(n.newFolderButton)
	first.AddChild(n.renameButton)
	first.AddChild(n.deleteButton)
	first.AddChild(n.favoriteButton)
	n.buttonRow = first // n.appUpdateButton joins this row only while there is an update to announce
	for _, child := range first.Children() {
		child.SetLayoutData(align.Middle)
	}
	first.SetLayout(&unison.FlowLayout{
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	first.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})

	n.backButton = unison.NewSVGButton(svg.Back)
	n.backButton.Tooltip = newWrappedTooltip(i18n.Text("Previous Match"))
	n.backButton.ClickCallback = n.previousMatch
	n.backButton.SetEnabled(false)

	n.forwardButton = unison.NewSVGButton(svg.Forward)
	n.forwardButton.Tooltip = newWrappedTooltip(i18n.Text("Next Match"))
	n.forwardButton.ClickCallback = n.nextMatch
	n.forwardButton.SetEnabled(false)

	n.searchField = NewSearchField(i18n.Text("Search"), n.searchModified)
	n.searchField.KeyDownCallback = n.searchKeydown

	n.matchesLabel = unison.NewLabel()
	n.matchesLabel.SetTitle("-")
	n.matchesLabel.Tooltip = newWrappedTooltip(i18n.Text("Number of matches found"))

	second := unison.NewPanel()
	second.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	second.AddChild(n.backButton)
	second.AddChild(n.forwardButton)
	second.AddChild(n.searchField)
	second.AddChild(n.matchesLabel)
	second.SetLayout(&unison.FlexLayout{
		Columns:  len(second.Children()),
		HSpacing: unison.StdHSpacing,
	})

	n.toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{},
		geom.Insets{Bottom: 1}, false), unison.NewEmptyBorder(unison.StdInsets())))
	n.toolbar.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: unison.StdVSpacing,
	})
	n.toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	n.toolbar.AddChild(first)
	n.toolbar.AddChild(second)
}

// InitialFocus causes the navigator to focus its initial component.
func (n *Navigator) InitialFocus() {
	FocusFirstContent(n.toolbar, n.table.AsPanel())
}

func (n *Navigator) addLibrary() {
	ShowLibrarySettings(&gurps.Library{})
}

func (n *Navigator) favoriteSelection() {
	if n.table.HasSelection() && toggleFavorites(n.table.SelectedRows(true)) {
		n.Reload()
	}
}

// toggleFavorites toggles the favorite state of each of the given rows, skipping library and favorites nodes as well as
// any row that resolves to a file already toggled by an earlier row. Returns true if at least one favorite was toggled.
func toggleFavorites(rows []*NavigatorNode) bool {
	changed := false
	seen := make(map[string]bool)
	for _, row := range rows {
		if row.IsLibrary() || row.IsFavorites() {
			continue
		}
		// Key on the full path on disk rather than the library-relative path, since the same relative path can exist in
		// more than one library and favorites are tracked per library.
		p := row.Path()
		if seen[p] {
			continue
		}
		changed = true
		seen[p] = true
		row.library.ToggleFavorite(row.path)
	}
	return changed
}

func (n *Navigator) deleteSelection() {
	if n.table.HasSelection() {
		selection := n.table.SelectedRows(true)
		hasLibs := false
		hasOther := false
		title := ""
		for _, row := range selection {
			if row.IsLibrary() {
				if row.library.IsMaster() || row.library.IsUser() {
					return
				}
				if title == "" {
					title = row.library.Data().Title
				} else {
					title = i18n.Text("these libraries")
				}
				hasLibs = true
			} else {
				hasOther = true
				if title == "" {
					title = row.primaryColumnText()
				} else {
					title = i18n.Text("these entries")
				}
			}
		}
		switch {
		case hasLibs && hasOther:
			return
		case hasLibs:
			header := xstrings.Wrap("", fmt.Sprintf(i18n.Text("Are you sure you want to remove %s?"), title), 100)
			if unison.QuestionDialog(header,
				i18n.Text("Note: This action will NOT remove any files from disk.")) == unison.ModalResponseOK {
				libs := gurps.GlobalSettings().Libraries
				for _, row := range selection {
					// Removal must go through Remove rather than delete(), since the deep search content loaders
					// read the library set from background goroutines.
					libs.Remove(row.library.Key())
					row.library.StopAllWatches()
					// Close any settings dockable still open for the library, since an apply from it would put the
					// library back into the set.
					closeLibrarySettings(row.library)
				}
				n.Reload()
			}
		case hasOther:
			header := xstrings.Wrap("", fmt.Sprintf(i18n.Text("Are you sure you want to remove %s?"), title), 100)
			note := xstrings.Wrap("", fmt.Sprintf(i18n.Text("Note: This action cannot be undone and will remove %s from disk."), title), 100)
			if unison.QuestionDialog(header, note) == unison.ModalResponseOK {
				if n.closeSelection(selection) {
					defer n.Reload()
					for _, row := range selection {
						p := row.Path()
						if row.IsDirectory() {
							if err := os.RemoveAll(p); err != nil {
								Workspace.ErrorHandler(fmt.Sprintf(i18n.Text("Unable to remove directory:\n%s"), p), err)
								return
							}
						} else {
							if err := os.Remove(p); err != nil {
								Workspace.ErrorHandler(fmt.Sprintf(i18n.Text("Unable to remove file:\n%s"), p), err)
								return
							}
						}
					}
				}
			}
		}
	}
}

func (n *Navigator) closeSelection(selection []*NavigatorNode) bool {
	for _, row := range selection {
		p := row.Path()
		if row.IsDirectory() {
			if len(row.children) != 0 {
				if !n.closeSelection(row.children) {
					return false
				}
			}
		} else {
			if dockable := LocateFileBackedDockable(p); dockable != nil {
				if !dockable.MayAttemptClose() {
					return false
				}
				if !dockable.AttemptClose() {
					return false
				}
			}
		}
	}
	return true
}

var disallowedWindowsFileNames = map[string]bool{
	"con":  true,
	"prn":  true,
	"aux":  true,
	"nul":  true,
	"com0": true,
	"com1": true,
	"com2": true,
	"com3": true,
	"com4": true,
	"com5": true,
	"com6": true,
	"com7": true,
	"com8": true,
	"com9": true,
	"lpt0": true,
	"lpt1": true,
	"lpt2": true,
	"lpt3": true,
	"lpt4": true,
	"lpt5": true,
	"lpt6": true,
	"lpt7": true,
	"lpt8": true,
	"lpt9": true,
}

// trimmedNameIsValid returns the given name with leading and trailing whitespace removed, along with whether that
// trimmed name is usable as a file or directory name. Callers must build their paths from the trimmed name, since that
// is the name that was validated.
func trimmedNameIsValid(name string) (trimmed string, valid bool) {
	trimmed = strings.TrimSpace(name)
	return trimmed, trimmed != "" && !strings.HasPrefix(trimmed, ".") && !strings.ContainsAny(trimmed, `/\:`) &&
		!disallowedWindowsFileNames[strings.ToLower(trimmed)]
}

// renamedPath returns the path that renaming the file or directory at oldPath to the given new name produces. The name
// is trimmed, matching what trimmedNameIsValid validates, so that the path checked for collisions is the same one the
// rename targets.
func renamedPath(oldPath, newName string) string {
	return filepath.Join(filepath.Dir(oldPath), strings.TrimSpace(newName)+filepath.Ext(oldPath))
}

// newFolderPath returns the path that creating a folder with the given name inside parentDir produces. The name is
// trimmed, matching what trimmedNameIsValid validates, so that the path checked for collisions is the same one that
// gets created.
func newFolderPath(parentDir, name string) string {
	return filepath.Join(parentDir, strings.TrimSpace(name))
}

func (n *Navigator) renameSelection() {
	if n.table.SelectionCount() == 1 {
		row := n.table.SelectedRows(false)[0]
		if row.IsLibrary() {
			return
		}

		oldName := row.primaryColumnText()
		newName := oldName

		oldField := NewStringField(nil, "", "", func() string { return oldName }, func(_ string) {})
		oldField.SetEnabled(false)

		newField := NewStringField(nil, "", "", func() string { return newName }, func(s string) { newName = s })
		newField.SetMinimumTextWidthUsing(minTextWidthCandidate)

		panel := unison.NewPanel()
		panel.SetLayout(&unison.FlexLayout{
			Columns:  2,
			HSpacing: unison.StdHSpacing,
			VSpacing: unison.StdVSpacing,
		})
		panel.AddChild(NewFieldLeadingLabel(i18n.Text("Current Name"), false))
		panel.AddChild(oldField)
		panel.AddChild(NewFieldLeadingLabel(i18n.Text("New Name"), false))
		panel.AddChild(newField)

		dialog, err := unison.NewDialog(unison.DefaultDialogTheme.QuestionIcon,
			unison.DefaultDialogTheme.QuestionIconInk, panel,
			[]*unison.DialogButtonInfo{unison.NewCancelButtonInfo(), unison.NewOKButtonInfo()})
		if err != nil {
			Workspace.ErrorHandler(i18n.Text("Unable to create rename dialog"), err)
			return
		}
		newField.ValidateCallback = func() bool {
			_, valid := trimmedNameIsValid(newName)
			if valid {
				if _, err = os.Stat(renamedPath(row.Path(), newName)); err == nil {
					valid = false
				}
			}
			dialog.Button(unison.ModalResponseOK).SetEnabled(valid)
			return valid
		}
		newField.Validate() // Here to update the OK button
		if dialog.RunModal() == unison.ModalResponseOK {
			oldPath := row.Path()
			newPath := renamedPath(oldPath, newName)
			if err = os.Rename(oldPath, newPath); err != nil {
				Workspace.ErrorHandler(fmt.Sprintf(i18n.Text("Unable to rename:\n%s"), oldPath), err)
			} else {
				n.fixupFavoritePath(row, oldPath, newPath)
				n.adjustBackingFilePath(row, oldPath, newPath)
				n.Reload()
				n.ApplySelectedPaths([]string{newPath})
				n.MarkForRedraw()
			}
		}
	}
}

func (n *Navigator) fixupFavoritePath(row *NavigatorNode, oldPath, newPath string) {
	if row.IsFile() || row.IsDirectory() {
		prefix := row.library.Data().PathOnDisk + string([]rune{filepath.Separator})
		row.library.RenameFavorite(strings.TrimPrefix(oldPath, prefix), strings.TrimPrefix(newPath, prefix))
	}
}

func (n *Navigator) adjustBackingFilePath(row *NavigatorNode, oldPath, newPath string) {
	switch {
	case row.IsDirectory():
		if !strings.HasSuffix(oldPath, string(os.PathSeparator)) {
			oldPath += string(os.PathSeparator)
		}
		for _, one := range AllDockables() {
			if fbd, ok := one.(FileBackedDockable); ok {
				p := fbd.BackingFilePath()
				if after, ok2 := strings.CutPrefix(p, oldPath); ok2 {
					fbd.SetBackingFilePath(filepath.Join(newPath, after))
				}
			}
		}
	case row.IsFile():
		if dockable := LocateFileBackedDockable(oldPath); dockable != nil {
			dockable.SetBackingFilePath(newPath)
		}
	}
}

// selectedLibraries returns the libraries among the selected rows, in selection order.
func (n *Navigator) selectedLibraries() []*gurps.Library {
	var libs []*gurps.Library
	for _, row := range n.table.SelectedRows(true) {
		if row.IsLibrary() {
			libs = append(libs, row.library)
		}
	}
	return libs
}

// updateLibrarySelection updates each selected library to its newest release, checking first for any whose releases
// aren't known yet.
func (n *Navigator) updateLibrarySelection() {
	libs := n.selectedLibraries()
	if !n.checkLibraryReleases(libs) {
		return
	}
	for _, lib := range libs {
		_, releases := lib.AvailableReleases()
		if len(releases) == 0 || !releases[0].HasUpdate() {
			reportNoLibraryReleases(lib)
			return
		}
		if !initiateLibraryUpdate(lib, &releases[0]) {
			return
		}
	}
}

// showSelectionReleaseNotes shows the release notes of each selected library, checking first for any whose releases
// aren't known yet.
func (n *Navigator) showSelectionReleaseNotes() {
	libs := n.selectedLibraries()
	if !n.checkLibraryReleases(libs) {
		return
	}
	for _, lib := range libs {
		current, releases := lib.AvailableReleases()
		if len(releases) == 0 || !releases[0].HasUpdate() {
			reportNoLibraryReleases(lib)
			return
		}
		var content strings.Builder
		for i, release := range releases {
			if i != 0 {
				content.WriteString("\n\n")
			}
			content.WriteString(i18n.Text("### Version "))
			content.WriteString(filterVersion(release.Version))
			content.WriteString("\n\n")
			if release.Version == current {
				content.WriteString(i18n.Text("> This version is what you currently have on disk."))
				content.WriteString("\n\n")
			}
			content.WriteString(release.Notes)
		}
		ShowReadOnlyMarkdown(fmt.Sprintf(i18n.Text("%s Release Notes"), lib.Data().Title), content.String())
	}
}

// checkLibraryReleases is checkLibraryReleases with the toolbar brought back into line afterwards: a check that found
// nothing to offer leaves the buttons with nothing to do, and one that found an update reloads the tree on its own.
func (n *Navigator) checkLibraryReleases(libs []*gurps.Library) bool {
	ok := checkLibraryReleases(libs)
	n.selectionChanged()
	return ok
}

func (n *Navigator) configureSelection() {
	for _, row := range n.table.SelectedRows(true) {
		if row.IsLibrary() {
			ShowLibrarySettings(row.library)
		}
	}
}

func (n *Navigator) searchKeydown(keyCode unison.KeyCode, mods mod.Modifiers, repeat bool) bool {
	if keyCode == unison.KeyReturn || keyCode == unison.KeyNumPadEnter {
		if mods.ShiftDown() {
			n.previousMatch()
		} else {
			n.nextMatch()
		}
		return true
	}
	return n.searchField.DefaultKeyDown(keyCode, mods, repeat)
}

func (n *Navigator) tableKeyDown(keyCode unison.KeyCode, mods mod.Modifiers, repeat bool) bool {
	if unison.IsControlAction(keyCode, mods) {
		return n.table.DefaultKeyDown(keyCode, mods, repeat)
	}
	switch keyCode {
	case unison.KeyBackspace, unison.KeyDelete:
		if n.deleteButton.Enabled() {
			n.deleteButton.Click()
		}
		return true
	default:
		return n.table.DefaultKeyDown(keyCode, mods, repeat)
	}
}

func (n *Navigator) mouseDown(where geom.Point, button, clickCount int, mods mod.Modifiers) bool {
	stop := n.table.DefaultMouseDown(where, button, clickCount, mods)
	if button == unison.ButtonRight && clickCount == 1 {
		if sel := n.table.SelectedRows(false); len(sel) != 0 {
			f := unison.DefaultMenuFactory()
			cm := f.NewMenu(unison.PopupMenuTemporaryBaseID|unison.ContextMenuIDFlag, "", nil)
			id := 1
			for _, one := range sel {
				if one.IsFile() || one.IsDirectory() {
					cm.InsertItem(-1, newContextMenuItemFromButton(f, &id, n.favoriteButton))
					cm.InsertSeparator(-1, true)
					break
				}
			}
			if len(sel) == 1 && sel[0].IsFile() {
				p := sel[0].Path()
				switch filepath.Ext(p) {
				case gurps.SheetExt:
					cm.InsertItem(-1, cloneSheetMenuItem(f, &id, p))
					cm.InsertSeparator(-1, true)
				case gurps.TemplatesExt:
					cm.InsertItem(-1, newSheetFromTemplateMenuItem(f, &id, p))
					if CanApplyTemplate() {
						cm.InsertItem(-1, newApplyTemplateMenuItem(f, &id, p))
					}
					cm.InsertSeparator(-1, true)
				}
			}
			cm.InsertItem(-1, newShowNodeOnDiskMenuItem(f, &id, sel))
			cm.InsertSeparator(-1, true)
			cm.InsertItem(-1, newContextMenuItemFromButton(f, &id, n.libraryReleaseNotesButton))
			cm.InsertItem(-1, newContextMenuItemFromButton(f, &id, n.configLibraryButton))
			cm.InsertItem(-1, newContextMenuItemFromButton(f, &id, n.downloadLibraryButton))
			cm.InsertSeparator(-1, true)
			cm.InsertItem(-1, newContextMenuItemFromButton(f, &id, n.newFolderButton))
			cm.InsertItem(-1, newContextMenuItemFromButton(f, &id, n.renameButton))
			cm.InsertItem(-1, newContextMenuItemFromButton(f, &id, n.deleteButton))
			count := cm.Count()
			if count > 0 {
				count--
				if cm.ItemAtIndex(count).IsSeparator() {
					cm.RemoveItem(count)
				}
				n.FlushDrawing()
				cm.Popup(geom.Rect{
					Point:  n.table.PointToRoot(where),
					Width:  1,
					Height: 1,
				}, 0)
			}
			cm.Dispose()
			stop = true
		}
	}
	return stop
}

func cloneSheetMenuItem(f unison.MenuFactory, id *int, sheetPath string) unison.MenuItem {
	useID := *id
	*id++
	return f.NewItem(unison.PopupMenuTemporaryBaseID+useID, cloneSheetAction.Title,
		unison.KeyBinding{}, nil, func(_ unison.MenuItem) {
			CloneSheet(sheetPath)
		})
}

func newSheetFromTemplateMenuItem(f unison.MenuFactory, id *int, templatePath string) unison.MenuItem {
	useID := *id
	*id++
	return f.NewItem(unison.PopupMenuTemporaryBaseID+useID, newSheetFromTemplateAction.Title,
		unison.KeyBinding{}, nil, func(_ unison.MenuItem) {
			NewSheetFromTemplate(templatePath)
		})
}

func newApplyTemplateMenuItem(f unison.MenuFactory, id *int, templatePath string) unison.MenuItem {
	useID := *id
	*id++
	return f.NewItem(unison.PopupMenuTemporaryBaseID+useID, applyTemplateAction.Title,
		unison.KeyBinding{}, nil, func(_ unison.MenuItem) {
			ApplyTemplate(templatePath)
		})
}

func newContextMenuItemFromButton(f unison.MenuFactory, id *int, button *unison.Button) unison.MenuItem {
	if button.Enabled() {
		useID := *id
		*id++
		var title string
		if label, ok := button.Tooltip.Children()[0].Self.(*unison.Label); ok {
			title = label.String()
		}
		return f.NewItem(unison.PopupMenuTemporaryBaseID+useID, title, unison.KeyBinding{}, nil,
			func(_ unison.MenuItem) { button.ClickCallback() })
	}
	return nil
}

func newShowNodeOnDiskMenuItem(f unison.MenuFactory, id *int, sel []*NavigatorNode) unison.MenuItem {
	useID := *id
	*id++
	return f.NewItem(unison.PopupMenuTemporaryBaseID+useID, i18n.Text("Show on Disk"), unison.KeyBinding{}, nil,
		func(_ unison.MenuItem) {
			m := make(map[string]struct{})
			for _, node := range sel {
				p := node.Path()
				if node.IsFile() {
					p = filepath.Dir(p)
				}
				m[p] = struct{}{}
			}
			for p := range m {
				if err := xos.OpenBrowser(p); err != nil {
					Workspace.ErrorHandler(i18n.Text("Unable to show location on disk"), err)
				}
			}
		})
}

// watchCallback is what the filesystem watches on the libraries report each change to, on the UI thread. Whatever the
// change, the deep search content cache entry for the path is dropped so that a search re-reads the file rather than
// matching on its previous contents. The path arrives named the way the library names it (see Library.Watch), which
// is also how NavigatorNode.Path() names it, so it is the cache key as is. A change that may have altered the tree is
// followed by a reload; a plain write to a file's contents does not need one, and a file arriving in pieces would
// otherwise restart the reload's cache rebuild on every piece.
func (n *Navigator) watchCallback(_ *gurps.Library, fullPath string, what notify.Event) {
	n.invalidateContentCacheEntry(fullPath)
	if what&^notify.Write != 0 {
		n.EventuallyReload()
	}
}

// EventuallyReload calls Reload() after a small delay, collapsing intervening requests to do the same. May be called
// from any goroutine: the library update checks and the filesystem watches both report from background goroutines, so
// the needReload bookkeeping is pushed onto the UI thread rather than being touched directly.
func (n *Navigator) EventuallyReload() {
	unison.InvokeTask(func() {
		if !n.needReload {
			n.needReload = true
			unison.InvokeTaskAfter(n.Reload, time.Millisecond*100)
		}
	})
}

// Reload the content of the navigator view.
func (n *Navigator) Reload() {
	n.needReload = false
	for _, token := range n.tokens {
		token.Stop()
	}
	n.tokens = nil
	disclosed := n.DisclosedPaths()
	selection := n.SelectedPaths()
	n.table.SetRootRows(n.populateRows())
	n.ApplyDisclosedPaths(disclosed)
	n.table.SyncToModel()
	n.ApplySelectedPaths(selection)
	n.table.SizeColumnsToFit(true)
	n.prewarmContentCache()
}

func (n *Navigator) populateRows() []*NavigatorNode {
	libs := gurps.GlobalSettings().Libraries.List()
	rows := make([]*NavigatorNode, 0, 1+len(libs))
	rows = append(rows, NewFavoritesNode(n))
	for _, lib := range libs {
		n.tokens = append(n.tokens, lib.Watch(n.watchCallback, true))
		rows = append(rows, NewLibraryNode(n, lib))
	}
	return rows
}

func (n *Navigator) adjustTableSizeEventually() {
	if !n.adjustTableSizePending {
		n.adjustTableSizePending = true
		unison.InvokeTaskAfter(n.adjustTableSize, time.Millisecond)
	}
}

func (n *Navigator) adjustTableSize() {
	n.adjustTableSizePending = false
	n.table.SyncToModel()
	n.table.SizeColumnsToFit(true)
}

// TitleIcon implements unison.Dockable
func (n *Navigator) TitleIcon(suggestedSize geom.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  unison.DocumentSVG,
		Size: suggestedSize,
	}
}

// Title implements unison.Dockable
func (n *Navigator) Title() string {
	return i18n.Text("Library Explorer")
}

// Tooltip implements unison.Dockable
func (n *Navigator) Tooltip() string {
	return ""
}

// Modified implements unison.Dockable
func (n *Navigator) Modified() bool {
	return false
}

func (n *Navigator) selectionChanged() {
	deleteEnabled := false
	renameEnabled := false
	newFolderEnabled := false
	downloadEnabled := false
	configEnabled := false
	favoriteEnabled := false
	if n.table.HasSelection() {
		deleteEnabled = true
		downloadEnabled = true
		configEnabled = true
		renameEnabled = n.table.SelectionCount() == 1
		newFolderEnabled = renameEnabled
		hasLibs := false
		hasOther := false
		for _, row := range n.table.SelectedRows(true) {
			if row.IsLibrary() {
				renameEnabled = false
				hasLibs = true
				if row.library.IsMaster() || row.library.IsUser() {
					deleteEnabled = false
				}
				if downloadEnabled {
					downloadEnabled = libraryUpdateButtonsEnabled(row.library)
				}
			} else {
				hasOther = true
				configEnabled = false
				downloadEnabled = false
				favoriteEnabled = true
				if row.IsFavorites() {
					renameEnabled = false
					deleteEnabled = false
					favoriteEnabled = false
				}
			}
		}
		if hasLibs && hasOther {
			deleteEnabled = false
		}
	}
	n.favoriteButton.SetEnabled(favoriteEnabled)
	n.deleteButton.SetEnabled(deleteEnabled)
	n.renameButton.SetEnabled(renameEnabled)
	n.newFolderButton.SetEnabled(newFolderEnabled)
	n.downloadLibraryButton.SetEnabled(downloadEnabled)
	n.libraryReleaseNotesButton.SetEnabled(downloadEnabled)
	n.configLibraryButton.SetEnabled(configEnabled)
}

func (n *Navigator) handleSelectionDoubleClick() {
	selection := n.table.SelectedRows(false)
	if len(selection) > 4 {
		if unison.QuestionDialog(i18n.Text("Are you sure you want to open all of these?"),
			fmt.Sprintf(i18n.Text("%d files will be opened."), len(selection))) != unison.ModalResponseOK {
			return
		}
	}
	altered := false
	for _, row := range selection {
		if row.CanHaveChildren() {
			altered = true
			row.SetOpen(!row.IsOpen())
		} else {
			if d, _ := row.OpenNodeContent(); !xreflect.IsNil(d) {
				if slices.Contains(n.searchResult, row) {
					// If we didn't match on the file name, copy the search text into the newly opened dockable's search
					// field
					if !row.Match(strings.ToLower(n.searchField.Text())) {
						if f := findSearchFieldInSelfOrDescendants(d.AsPanel()); f != nil {
							f.SetText(n.searchField.Text())
						}
					}
				}
			}
		}
	}
	if altered {
		n.table.SyncToModel()
	}
}

func findSearchFieldInSelfOrDescendants(p *unison.Panel) *unison.Field {
	if f, ok := p.Self.(*unison.Field); ok {
		if _, ok = f.ClientData()[searchFieldClientDataKey]; ok {
			return f
		}
	}
	for _, child := range p.Children() {
		if f := findSearchFieldInSelfOrDescendants(child); f != nil {
			return f
		}
	}
	return nil
}

func (n *Navigator) toggleHierarchy() {
	first := true
	open := false
	for _, row := range n.table.RootRows() {
		if row.CanHaveChildren() {
			if first {
				first = false
				open = !row.IsOpen()
			}
			setNavigatorRowOpen(row, open)
		}
	}
	n.table.SyncToModel()
	n.table.PruneSelectionOfUndisclosedNodes()
}

func setNavigatorRowOpen(row *NavigatorNode, open bool) {
	row.SetOpen(open)
	for _, child := range row.Children() {
		if child.CanHaveChildren() {
			setNavigatorRowOpen(child, open)
		}
	}
}

func (n *Navigator) searchModified(_, _ *unison.FieldState) {
	n.searchIndex = -1
	n.searchResult = nil
	n.search(strings.ToLower(n.searchField.Text()), n.table.RootRows())
	n.adjustForMatch()
}

// rerunSearch rebuilds the search results for the given text, preserving the user's place in the match list when the
// row they were on is still among the new results. A completed background cache prewarm re-runs any active search, and
// without this the completion would silently reset the position while the user is walking the matches.
func (n *Navigator) rerunSearch(text string, rows []*NavigatorNode) {
	var current *NavigatorNode
	if n.searchIndex >= 0 && n.searchIndex < len(n.searchResult) {
		current = n.searchResult[n.searchIndex]
	}
	n.searchResult = nil
	n.search(text, rows)
	n.searchIndex = -1
	if current != nil {
		n.searchIndex = slices.Index(n.searchResult, current)
		if n.searchIndex == -1 {
			// A Reload rebuilds the tree with fresh node objects, so pointer identity fails even though the row the
			// user was on is still among the results. Fall back to the path, the same identity the navigator uses to
			// carry selection and disclosure across a reload. A favorited file appears twice with the same path; the
			// first copy is close enough.
			p := current.Path()
			n.searchIndex = slices.IndexFunc(n.searchResult, func(row *NavigatorNode) bool { return row.Path() == p })
		}
	}
}

// rerunActiveSearch re-runs the search currently in the search field, if there is one, keeping the user's place in the
// match list, and refreshes the search toolbar. Only the toolbar is refreshed: this runs from asynchronous completions,
// and re-selecting the current match would yank the selection and scroll position away from whatever the user has
// clicked on since.
func (n *Navigator) rerunActiveSearch() {
	if n.searchField != nil && n.searchField.Text() != "" {
		n.rerunSearch(strings.ToLower(n.searchField.Text()), n.table.RootRows())
		n.updateMatchControls()
	}
}

func (n *Navigator) search(text string, rows []*NavigatorNode) {
	if text == "" {
		return
	}
	for _, row := range rows {
		if row.Match(text) {
			n.searchResult = append(n.searchResult, row)
		} else if row.IsFile() {
			p := row.Path()
			// The deep search check must come before the cache lookup, so that disabling a type in the settings takes
			// effect immediately rather than only after the asynchronous cache rebuild completes.
			if n.deepSearch[gurps.FileInfoFor(p).UTI.Extensions[0]] {
				// A cache hit is served as is, without checking the file on disk: the filesystem watches drop the
				// entry for a file as soon as its change is reported (see watchCallback), which is what keeps this
				// from being a stat per deep-searchable file on every keystroke.
				entry, ok := n.contentCache[p]
				if !ok || entry == nil {
					entry = loadContentCacheEntry(p)
					n.addToContentCache(p, entry)
				}
				if entry != nil && entry.content != "" && strings.Contains(entry.content, text) {
					n.searchResult = append(n.searchResult, row)
				}
			}
		}
		if row.CanHaveChildren() {
			n.search(text, row.Children())
		}
	}
}

// prepareProfileForContentCache returns the profile's text fields in the form the deep search content cache expects.
// The search text is lowercased before comparison, so the content must be lowercased as well.
func prepareProfileForContentCache(profile *gurps.Profile) string {
	return strings.ToLower(strings.Join([]string{
		profile.Name,
		profile.Age,
		profile.Birthday,
		profile.Eyes,
		profile.Hair,
		profile.Skin,
		profile.Handedness,
		profile.Gender,
		profile.PlayerName,
		profile.Title,
		profile.Organization,
		profile.Religion,
	}, "\n"))
}

func prepareForContentCache[T gurps.Node[T]](data []T) string {
	var buffer strings.Builder
	gurps.Traverse(func(one T) bool {
		buffer.WriteString(strings.ToLower(one.String()))
		buffer.WriteByte('\n')
		return false
	}, false, false, data...)
	return buffer.String()
}

// contentCacheEntry holds the extracted, lowercased and trimmed text of one file for the deep search, along with the
// file metadata used to decide whether the entry can be reused when the cache is revalidated after a reload. Size and
// modification time are deliberately used instead of a content hash: hashing would require reading every file just to
// decide whether it needs to be re-read, and the case it would catch — a content change that alters neither size nor
// modification time — is unlikely to occur in real use.
type contentCacheEntry struct {
	modTime time.Time
	content string
	size    int64
}

// isCurrent returns true if the file at the given path still has the size and modification time captured in the entry.
// Only the background cache builds check this; a search serves cache hits without touching the disk, relying on the
// filesystem watches to drop the entries of changed files (see invalidateContentCacheEntry).
func (e *contentCacheEntry) isCurrent(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.Size() == e.size && fi.ModTime().Equal(e.modTime)
}

// contentCacheBuild is one background build of the deep search content cache, kept so that the build which supersedes
// it can inherit the entries it completed instead of starting over from the live cache alone. Reloads arrive
// back-to-back at startup — the navigator's own, then one for each launch-time library check that finds a change — and
// each cancels the build before it, so without the hand-off the first full parse of a large library would be restarted
// from zero several times over. The inherited entries are revalidated against the files' size and modification time
// like any other, so a canceled build's partial work is as safe to reuse as a finished one's.
type contentCacheBuild struct {
	done    chan struct{}                 // Closed once entries may be read
	entries map[string]*contentCacheEntry // What the build started from, overlaid with what it completed
}

// loadContentCacheEntry reads the file at the given path and extracts its searchable text. The search text is
// lowercased before comparison, so the content is lowercased as well. A file that fails to load or parse — including
// one whose parse panics — produces an entry with empty content, so it isn't pointlessly re-read (or re-panicked on)
// on every keystroke. The file metadata is captured before the read so that a concurrent modification can only make
// the entry look older than it is, never newer.
func loadContentCacheEntry(p string) *contentCacheEntry {
	entry := &contentCacheEntry{}
	if fi, err := os.Stat(p); err == nil {
		entry.modTime = fi.ModTime()
		entry.size = fi.Size()
	}
	xos.SafeCall(func() { entry.content = extractContentForCache(p) }, nil)
	return entry
}

// extractContentForCache extracts the searchable text of the file at the given path.
func extractContentForCache(p string) string {
	dir := os.DirFS(filepath.Dir(p))
	fileName := filepath.Base(p)
	var content string
	switch gurps.FileInfoFor(p).UTI.Extensions[0] {
	case gurps.EquipmentExt:
		if data, err := gurps.NewEquipmentFromFile(dir, fileName); err == nil {
			content = prepareForContentCache(data)
		}
	case gurps.EquipmentModifiersExt:
		if data, err := gurps.NewEquipmentModifiersFromFile(dir, fileName); err == nil {
			content = prepareForContentCache(data)
		}
	case gurps.NotesExt:
		if data, err := gurps.NewNotesFromFile(dir, fileName); err == nil {
			content = prepareForContentCache(data)
		}
	case gurps.SheetExt:
		if data, err := gurps.NewEntityFromFile(dir, fileName); err == nil {
			for _, one := range data.Skills {
				one.TechLevel = nil
			}
			for _, one := range data.Spells {
				one.TechLevel = nil
			}
			content = strings.Join([]string{
				prepareProfileForContentCache(&data.Profile),
				prepareForContentCache(data.Traits),
				prepareForContentCache(data.Skills),
				prepareForContentCache(data.Spells),
				prepareForContentCache(data.CarriedEquipment),
				prepareForContentCache(data.OtherEquipment),
				prepareForContentCache(data.Notes),
			}, "\n")
		}
	case gurps.SkillsExt:
		if data, err := gurps.NewSkillsFromFile(dir, fileName); err == nil {
			for _, one := range data {
				one.TechLevel = nil
			}
			content = prepareForContentCache(data)
		}
	case gurps.SpellsExt:
		if data, err := gurps.NewSpellsFromFile(dir, fileName); err == nil {
			for _, one := range data {
				one.TechLevel = nil
			}
			content = prepareForContentCache(data)
		}
	case gurps.TemplatesExt:
		if data, err := gurps.NewTemplateFromFile(dir, fileName); err == nil {
			for _, one := range data.Skills {
				one.TechLevel = nil
			}
			for _, one := range data.Spells {
				one.TechLevel = nil
			}
			content = strings.Join([]string{
				prepareForContentCache(data.Traits),
				prepareForContentCache(data.Skills),
				prepareForContentCache(data.Spells),
				prepareForContentCache(data.Equipment),
				prepareForContentCache(data.Notes),
			}, "\n")
		}
	case gurps.LootExt:
		if data, err := gurps.NewLootFromFile(dir, fileName); err == nil {
			content = strings.Join([]string{ //nolint:gocritic // Fine as-is
				prepareForContentCache(data.Equipment),
				prepareForContentCache(data.Notes),
			}, "\n")
		}
	// TODO: Re-enable Campaign files
	// case gurps.CampaignExt:
	// TODO: Implement
	case gurps.TraitModifiersExt:
		if data, err := gurps.NewTraitModifiersFromFile(dir, fileName); err == nil {
			content = prepareForContentCache(data)
		}
	case gurps.TraitsExt:
		if data, err := gurps.NewTraitsFromFile(dir, fileName); err == nil {
			content = prepareForContentCache(data)
		}
	case uti.Markdown.Extensions[0]:
		if data, err := os.ReadFile(p); err == nil {
			content = string(bytes.ToLower(data))
		}
	}
	return strings.TrimSpace(content)
}

func (n *Navigator) addToContentCache(p string, entry *contentCacheEntry) {
	if n.contentCache == nil {
		n.contentCache = make(map[string]*contentCacheEntry)
	}
	n.contentCache[p] = entry
}

// invalidateContentCacheEntry drops the deep search content cache entry for the given path, so that the next search
// re-reads the file rather than matching on its previous contents. The filesystem watches call this as each change is
// reported, which is what lets a search serve cache hits without checking the files on disk: revalidating every hit
// against the file's size and modification time instead would be a stat per deep-searchable file on each keystroke,
// thousands of syscalls per character typed with the master library enabled. A background build that is in flight may
// already have read the file's old contents, so the path is also noted for applyPrewarmedContentCache to drop from that
// build's result. Must be called on the UI thread, like the searches and the cache builds.
func (n *Navigator) invalidateContentCacheEntry(p string) {
	delete(n.contentCache, p)
	if n.invalidatedPaths == nil {
		n.invalidatedPaths = make(map[string]bool)
	}
	n.invalidatedPaths[p] = true
}

// collectDeepSearchPaths adds the path of each file row whose type is enabled for deep search to the given map. The
// favorites node repeats paths that also appear under their libraries, so a map is used to de-duplicate.
func (n *Navigator) collectDeepSearchPaths(rows []*NavigatorNode, paths map[string]bool) {
	for _, row := range rows {
		if row.IsFile() {
			if p := row.Path(); n.deepSearch[gurps.FileInfoFor(p).UTI.Extensions[0]] {
				paths[p] = true
			}
		}
		if row.CanHaveChildren() {
			n.collectDeepSearchPaths(row.Children(), paths)
		}
	}
}

// prewarmContentCache rebuilds the deep search content cache on background goroutines so the first keystroke in the
// search field doesn't have to read and parse every deep-searchable file inline on the UI thread. Entries from the
// previous cache are reused when the file's size and modification time are unchanged, so a reload triggered by a
// single file change only pays to re-parse that file rather than entire libraries. Must be called on the UI thread;
// the completed cache is swapped in on the UI thread as well. A newer call cancels an older one that is still running
// and inherits the entries it had completed, so back-to-back reloads make a large build incremental rather than
// restarting it from zero.
func (n *Navigator) prewarmContentCache() {
	if n.prewarmSuspensions > 0 {
		n.prewarmPending = true
		return
	}
	if n.table == nil {
		return // Not fully constructed, as happens in tests
	}
	gen := n.cacheGeneration.Add(1)
	// The build below starts from the cache as it is now, which already lacks the entries invalidated so far, so only
	// invalidations that arrive while it runs need to be applied to its result. The build it inherits from may have
	// read those files before their changes were reported, though, so its entries for them are not inherited.
	invalidated := n.invalidatedPaths
	n.invalidatedPaths = nil
	paths := make(map[string]bool)
	n.collectDeepSearchPaths(n.table.RootRows(), paths)
	if len(paths) == 0 {
		// The user just disabled the last deep-search type. There is nothing to build, so no completion will run the
		// usual re-search; re-run any active search here so the now-invalid deep-search matches don't linger in the
		// results until the next keystroke.
		n.contentCache = nil
		n.lastBuild = nil
		n.rerunActiveSearch()
		return
	}
	// Republish the settings the background parses read, so they are as fresh as this build and the workers never
	// touch the live, UI-thread-mutated structures even if some settings editor missed publishing after a change.
	gurps.SyncScriptExecTimeLimit()
	gurps.SyncGlobalSheetSettings()
	// The workers read this snapshot, since searches on the UI thread may add to the live map while they run. It is
	// allocated even when the live cache is empty, as it is at startup, since the build overlays its results onto it.
	previous := make(map[string]*contentCacheEntry, len(paths))
	maps.Copy(previous, n.contentCache)
	prior := n.lastBuild
	build := &contentCacheBuild{done: make(chan struct{})}
	n.lastBuild = build
	go func() {
		if prior != nil {
			// The generation bump above canceled the prior build, if it was still running, and it stops at the next
			// file boundary, so this wait is brief. Its entries fill in what the live cache lacks; anything the live
			// cache holds is at least as new.
			<-prior.done
			for p, entry := range prior.entries {
				if _, ok := previous[p]; !ok && !invalidated[p] {
					previous[p] = entry
				}
			}
		}
		fresh := buildContentCache(paths, previous, func() bool { return n.cacheGeneration.Load() != gen })
		// Hand everything this build knows to the one that supersedes it, if any: the entries it completed and the
		// ones it inherited but was canceled before it reached. The snapshot is private to this build and its
		// workers have all exited, so it can be overlaid in place.
		maps.Copy(previous, fresh)
		build.entries = previous
		close(build.done)
		unison.InvokeTask(func() {
			// A search made while the cache was still being built may have produced incomplete results, so re-run any
			// active search now that the full content is available, keeping the user's place in the match list.
			if n.applyPrewarmedContentCache(gen, fresh) {
				n.lastBuild = nil // The live cache now holds everything this build knew, so there is nothing to inherit
				n.rerunActiveSearch()
			}
		})
	}()
}

// applyPrewarmedContentCache installs a content cache produced by a background build and reports whether it did so. A
// build whose generation is no longer current has been superseded — a newer prewarm or a suspension started after it
// began — so its result is not installed, where it would clobber the cache the newer build will install; the newer
// build inherits its entries instead (see contentCacheBuild). Entries for files whose changes were reported while the
// build ran are left out, since the build may have read them before the change; the next search re-reads those inline.
// Must be called on the UI thread.
func (n *Navigator) applyPrewarmedContentCache(gen int64, fresh map[string]*contentCacheEntry) bool {
	if gen != n.cacheGeneration.Load() {
		return false
	}
	for p := range n.invalidatedPaths {
		delete(fresh, p)
	}
	n.invalidatedPaths = nil
	n.contentCache = fresh
	return true
}

// suspendContentCachePrewarm holds off background rebuilds of the deep search content cache until the matching call to
// resumeContentCachePrewarm. A library update writes hundreds of files, and each batch of filesystem watch events
// would otherwise kick off another rebuild that the next batch immediately cancels, so a caller that is about to churn
// the libraries suspends first. Any build already in flight is abandoned as well, since it is reading files that are
// about to be replaced; the rebuild on resume inherits what it completed and revalidates each entry against the file on
// disk, so the replaced files are re-read and the rest are not. Suspensions nest; must be called on the UI thread, like
// the prewarm itself.
func (n *Navigator) suspendContentCachePrewarm() {
	n.prewarmSuspensions++
	if n.prewarmSuspensions == 1 {
		n.cacheGeneration.Add(1)
		// A build may just have been abandoned, and there is no way to tell, so a rebuild on resume is always owed.
		n.prewarmPending = true
	}
}

// resumeContentCachePrewarm lifts the hold placed by the matching call to suspendContentCachePrewarm. When the last
// suspension is lifted, the single rebuild that the held-off requests collapsed into is owed. It is handed to a Reload
// rather than started directly, since Reload runs a prewarm of its own and the work a suspension covers ends with a
// reload anyway (a library update schedules one via EventuallyReload): starting the rebuild directly would only have
// that reload's prewarm cancel and repeat it. The EventuallyReload here collapses with the caller's into a single
// Reload, leaving one rebuild. Must be called on the UI thread.
func (n *Navigator) resumeContentCachePrewarm() {
	if n.liftContentCachePrewarmSuspension() {
		n.EventuallyReload()
	}
}

// liftContentCachePrewarmSuspension is the bookkeeping half of resumeContentCachePrewarm: it lifts one suspension and
// reports whether that was the last one with a rebuild owed, taking ownership of the owed rebuild -- clearing the
// pending flag -- when so. An unmatched call is harmless.
func (n *Navigator) liftContentCachePrewarmSuspension() bool {
	if n.prewarmSuspensions == 0 {
		return false
	}
	n.prewarmSuspensions--
	if n.prewarmSuspensions != 0 || !n.prewarmPending {
		return false
	}
	n.prewarmPending = false
	return true
}

// buildContentCache produces a content cache entry for each of the given paths, spreading the work across the
// available CPUs. Entries from the previous cache are reused when the file's size and modification time are
// unchanged. The canceled function, if non-nil, is polled between files so an obsolete build can stop early.
func buildContentCache(paths map[string]bool, previous map[string]*contentCacheEntry, canceled func() bool) map[string]*contentCacheEntry {
	var wg sync.WaitGroup
	var lock sync.Mutex
	fresh := make(map[string]*contentCacheEntry, len(paths))
	in := make(chan string, len(paths))
	for p := range paths {
		in <- p
	}
	close(in)
	for range min(runtime.NumCPU(), len(paths)) {
		wg.Go(func() {
			for p := range in {
				if canceled != nil && canceled() {
					return
				}
				entry := previous[p]
				if entry == nil || !entry.isCurrent(p) {
					entry = loadContentCacheEntry(p)
				}
				lock.Lock()
				fresh[p] = entry
				lock.Unlock()
			}
		})
	}
	wg.Wait()
	return fresh
}

func (n *Navigator) previousMatch() {
	if n.searchIndex > 0 {
		n.searchIndex--
		n.adjustForMatch()
	}
}

func (n *Navigator) nextMatch() {
	if n.searchIndex < len(n.searchResult)-1 {
		n.searchIndex++
		n.adjustForMatch()
	}
}

// adjustForMatch updates the search toolbar and, when a match is current, selects it and scrolls it into view. Only
// direct user actions (typing in the search field, stepping between matches) should call this; asynchronous paths that
// merely refresh the results must use updateMatchControls instead, since grabbing the selection while the user may
// have moved on to other rows would discard their place.
func (n *Navigator) adjustForMatch() {
	n.updateMatchControls()
	if n.searchIndex >= 0 && n.searchIndex < len(n.searchResult) {
		row := n.searchResult[n.searchIndex]
		n.table.DiscloseRow(row, false)
		n.table.ClearSelection()
		i := n.table.RowToIndex(row)
		n.table.SelectByIndex(i)
		n.ValidateLayout()
		n.table.ScrollRowIntoView(i)
	}
}

// updateMatchControls updates the back/forward buttons and the matches label for the current search results without
// touching the table's selection or scroll position.
func (n *Navigator) updateMatchControls() {
	n.backButton.SetEnabled(n.searchIndex > 0)
	n.forwardButton.SetEnabled(len(n.searchResult) != 0 && n.searchIndex != len(n.searchResult)-1)
	if len(n.searchResult) != 0 {
		if n.searchIndex < 0 {
			n.matchesLabel.SetTitle(fmt.Sprintf(i18n.Text("- of %d"), len(n.searchResult)))
		} else {
			n.matchesLabel.SetTitle(fmt.Sprintf(i18n.Text("%d of %d"), n.searchIndex+1, len(n.searchResult)))
		}
	} else {
		n.matchesLabel.SetTitle("-")
	}
	n.matchesLabel.Parent().MarkForLayoutAndRedraw()
}

// DisclosedPaths returns a list of paths that are currently disclosed.
func (n *Navigator) DisclosedPaths() []string {
	return n.accumulateDisclosedPaths(n.table.RootRows(), nil)
}

func (n *Navigator) accumulateDisclosedPaths(rows []*NavigatorNode, disclosedPaths []string) []string {
	for _, row := range rows {
		if row.IsOpen() {
			disclosedPaths = append(disclosedPaths, row.Path())
		}
		disclosedPaths = n.accumulateDisclosedPaths(row.Children(), disclosedPaths)
	}
	return disclosedPaths
}

// ApplyDisclosedPaths closes all nodes except the ones provided, which are explicitly opened.
func (n *Navigator) ApplyDisclosedPaths(paths []string) {
	m := make(map[string]bool, len(paths))
	for _, one := range paths {
		m[one] = true
	}
	n.applyDisclosedPaths(n.table.RootRows(), m)
}

func (n *Navigator) applyDisclosedPaths(rows []*NavigatorNode, paths map[string]bool) {
	for _, row := range rows {
		open := paths[row.Path()]
		if row.IsOpen() != open {
			row.SetOpen(open)
		}
		n.applyDisclosedPaths(row.Children(), paths)
	}
}

// SelectedPaths returns a list of paths that are currently selected.
func (n *Navigator) SelectedPaths() []string {
	sel := n.table.SelectedRows(false)
	paths := make([]string, 0, len(sel))
	for _, row := range sel {
		paths = append(paths, row.Path())
	}
	return paths
}

// ApplySelectedPaths replaces the selection with the nodes that match the given paths.
func (n *Navigator) ApplySelectedPaths(paths []string) {
	m := make(map[string]bool, len(paths))
	for _, p := range paths {
		m[p] = true
	}
	selMap := make(map[tid.TID]bool)
	count := n.table.LastRowIndex()
	for i := 0; i <= count; i++ {
		row := n.table.RowFromIndex(i)
		if m[row.Path()] {
			selMap[row.ID()] = true
		}
	}
	n.table.SetSelectionMap(selMap)
}

// OpenFiles attempts to open the given file paths.
func OpenFiles(filePaths []string) {
	for _, one := range filePaths {
		if p, err := filepath.Abs(one); err != nil {
			Workspace.ErrorHandler(i18n.Text("Unable to open ")+one, err)
		} else {
			Workspace.Window.ToFront()
			OpenFile(p, gurps.PageInfo{})
		}
	}
}

// DisplayNewDockable adds the Dockable to the dock and gives it the focus. A Dockable that is already on display -- the
// ancestry and name generator editors place themselves in the dock as they are shown, so opening one of their files
// hands back an editor already in the dock -- is activated where it is.
func DisplayNewDockable(dockable unison.Dockable) {
	if p := dockable.AsPanel(); p.Parent() != nil || p.Window() != nil {
		ActivateDockable(dockable)
		return
	}
	InstallDockUndockCmd(dockable)
	defer func() {
		if children := dockable.AsPanel().Children(); len(children) > 1 {
			FocusFirstContent(children[0], children[1])
		}
	}()
	if fbd, ok := dockable.(FileBackedDockable); ok {
		var group *dgroup.Group
		fi := gurps.FileInfoFor(fbd.BackingFilePath())
		switch {
		case fi.IsImage:
			g := dgroup.Images
			group = &g
		case fi.IsPDF:
			g := dgroup.PDFs
			group = &g
		case fi.UTI.Extensions[0] == gurps.SheetExt:
			g := dgroup.CharacterSheets
			group = &g
		case fi.UTI.Extensions[0] == gurps.TemplatesExt:
			g := dgroup.CharacterTemplates
			group = &g
		case fi.UTI.Extensions[0] == gurps.LootExt:
			g := dgroup.LootSheets
			group = &g
		// TODO: Re-enable Campaign files
		// case fi.UTI.Extensions[0] == gurps.CampaignExt:
		// 	g := dgroup.Campaigns
		// 	group = &g
		case fi.UTI.Extensions[0] == gurps.TraitsExt,
			fi.UTI.Extensions[0] == gurps.TraitModifiersExt,
			fi.UTI.Extensions[0] == gurps.EquipmentExt,
			fi.UTI.Extensions[0] == gurps.EquipmentModifiersExt,
			fi.UTI.Extensions[0] == gurps.SkillsExt,
			fi.UTI.Extensions[0] == gurps.SpellsExt,
			fi.UTI.Extensions[0] == gurps.NotesExt:
			g := dgroup.Libraries
			group = &g
		case fi.UTI == uti.Markdown:
			g := dgroup.Markdown
			group = &g
		}
		if group != nil {
			if slices.Contains(gurps.GlobalSettings().OpenInWindow, *group) {
				// The window may not come into existence until later, in which case the focus the deferred call above
				// asks for goes nowhere and is made again once the window has been created.
				if _, err := placeInWindow(dockable, *group); err != nil {
					errs.Log(err)
				}
				return
			}
			dockable.AsPanel().ClientData()[dockGroupClientDataKey] = *group
		}
		if dc := CurrentlyFocusedDockContainer(); dc != nil && DockContainerHoldsExtension(dc, fi.GroupWith...) {
			dc.Stack(dockable, -1)
			return
		} else if dc = LocateDockContainerForExtension(fi.GroupWith...); dc != nil {
			dc.Stack(dockable, -1)
			return
		}
	}
	Workspace.DocumentDock.DockTo(dockable, nil, side.Right)
}

// OpenFile attempts to open the given file path.
func OpenFile(filePath string, initialPage gurps.PageInfo) (dockable unison.Dockable, wasOpen bool) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		Workspace.ErrorHandler(fmt.Sprintf(i18n.Text("Unable to resolve path:\n%s"), filePath), err)
		return nil, false
	}
	if d := LocateFileBackedDockable(absPath); d != nil {
		ActivateDockable(d)
		return d, true
	}
	fi := gurps.FileInfoFor(absPath)
	if fi.IsSpecial {
		return nil, false
	}
	if fi.IsPDF && strings.TrimSpace(gurps.GlobalSettings().General.ExternalPDFCmdLine) != "" {
		openExternalPDF(absPath, "", initialPage)
		return nil, false
	}
	var d unison.Dockable
	if d, err = fi.Load(absPath, initialPage); err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to open file:\n")+absPath, err)
		return nil, false
	}
	gurps.GlobalSettings().AddRecentFile(absPath)
	DisplayNewDockable(d)
	return d, false
}

func (n *Navigator) newFolder() {
	if n.table.SelectionCount() == 1 {
		row := n.table.SelectedRows(false)[0]
		parentDir := row.Path()
		if row.IsFile() {
			parentDir = filepath.Dir(parentDir)
		}
		name := ""
		field := NewStringField(nil, "", "", func() string { return name }, func(s string) { name = s })
		field.SetMinimumTextWidthUsing(minTextWidthCandidate)

		panel := unison.NewPanel()
		panel.SetLayout(&unison.FlexLayout{
			Columns:  2,
			HSpacing: unison.StdHSpacing,
			VSpacing: unison.StdVSpacing,
		})
		panel.AddChild(NewFieldLeadingLabel(i18n.Text("Folder Name"), false))
		panel.AddChild(field)

		dialog, err := unison.NewDialog(unison.DefaultDialogTheme.QuestionIcon,
			unison.DefaultDialogTheme.QuestionIconInk, panel,
			[]*unison.DialogButtonInfo{unison.NewCancelButtonInfo(), unison.NewOKButtonInfo()})
		if err != nil {
			Workspace.ErrorHandler(i18n.Text("Unable to create new folder dialog"), err)
			return
		}
		field.ValidateCallback = func() bool {
			_, valid := trimmedNameIsValid(name)
			if valid {
				if _, err = os.Stat(newFolderPath(parentDir, name)); err == nil {
					valid = false
				}
			}
			dialog.Button(unison.ModalResponseOK).SetEnabled(valid)
			return valid
		}
		field.Validate() // Here to update the OK button
		if dialog.RunModal() == unison.ModalResponseOK {
			dirPath := newFolderPath(parentDir, name)
			if err = os.Mkdir(dirPath, 0o750); err != nil {
				Workspace.ErrorHandler(fmt.Sprintf(i18n.Text("Unable to create:\n%s"), dirPath), err)
			} else {
				if !row.IsFile() && !row.IsOpen() {
					row.SetOpen(true)
				}
				n.Reload()
				n.ApplySelectedPaths([]string{dirPath})
				n.MarkForRedraw()
			}
		}
	}
}
