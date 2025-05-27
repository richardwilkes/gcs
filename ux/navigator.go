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
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/desktop"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/tid"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
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
	scroll                    *unison.ScrollPanel
	table                     *unison.Table[*NavigatorNode]
	tokens                    []*gurps.MonitorToken
	searchResult              []*NavigatorNode
	deepSearch                map[string]bool
	contentCache              map[string]string
	searchIndex               int
	needReload                bool
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

	n.table.Columns = make([]unison.ColumnInfo, 1)
	n.needReload = true
	rows := n.populateRows()
	n.needReload = false
	n.table.SetScale(float32(globalSettings.General.NavigatorUIScale) / 100)
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
	gurps.NotifyOfLibraryChangeFunc = n.EventuallyReload
	n.table.MouseDownCallback = n.mouseDown
	n.table.SelectionChangedCallback = n.selectionChanged
	n.table.KeyDownCallback = n.tableKeyDown

	n.InstallCmdHandlers(JumpToSearchFilterItemID,
		func(any) bool { return !n.searchField.Focused() },
		func(any) { n.searchField.RequestFocus() })

	n.selectionChanged()
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
		for _, ext := range gurps.FileInfoFor(one).Extensions {
			n.deepSearch[ext] = true
		}
	}
}

func (n *Navigator) setupToolBar() {
	helpButton := unison.NewSVGButton(svg.Help)
	helpButton.Tooltip = newWrappedTooltip(i18n.Text("Help"))
	helpButton.ClickCallback = func() { HandleLink(nil, "md:Help/Interface/Library Tree") }

	hierarchyButton := unison.NewSVGButton(svg.Hierarchy)
	hierarchyButton.Tooltip = newWrappedTooltip(i18n.Text("Opens/closes all hierarchical rows"))
	hierarchyButton.ClickCallback = n.toggleHierarchy

	n.deleteButton = unison.NewSVGButton(svg.Trash)
	n.deleteButton.Tooltip = newWrappedTooltip(i18n.Text("Delete"))
	n.deleteButton.ClickCallback = n.deleteSelection

	n.renameButton = unison.NewSVGButton(svg.SignPost)
	n.renameButton.Tooltip = newWrappedTooltip(i18n.Text("Rename"))
	n.renameButton.ClickCallback = n.renameSelection

	n.newFolderButton = unison.NewSVGButton(svg.NewFolder)
	n.newFolderButton.Tooltip = newWrappedTooltip(i18n.Text("New Folder"))
	n.newFolderButton.ClickCallback = n.newFolder

	addLibraryButton := unison.NewSVGButton(svg.CircledAdd)
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

	n.toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, 0, unison.Insets{Bottom: 1},
		false), unison.NewEmptyBorder(unison.StdInsets())))
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
	if n.table.HasSelection() {
		changed := false
		selection := n.table.SelectedRows(true)
		seen := make(map[string]bool)
		for _, row := range selection {
			if seen[row.path] || row.IsLibrary() || row.IsFavorites() {
				continue
			}
			changed = true
			seen[row.path] = true
			if i := slices.Index(row.library.Favorites, row.path); i != -1 {
				row.library.Favorites = slices.Delete(row.library.Favorites, i, i+1)
			} else {
				row.library.Favorites = append(row.library.Favorites, row.path)
			}
		}
		if changed {
			n.Reload()
		}
	}
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
					title = row.library.Title
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
		if hasLibs && hasOther {
			return
		}
		switch {
		case hasLibs && hasOther:
			return
		case hasLibs:
			header := txt.Wrap("", fmt.Sprintf(i18n.Text("Are you sure you want to remove %s?"), title), 100)
			if unison.QuestionDialog(header,
				i18n.Text("Note: This action will NOT remove any files from disk.")) == unison.ModalResponseOK {
				libs := gurps.GlobalSettings().LibrarySet
				for _, row := range selection {
					delete(libs, row.library.Key())
					row.library.StopAllWatches()
				}
				n.Reload()
			}
		case hasOther:
			header := txt.Wrap("", fmt.Sprintf(i18n.Text("Are you sure you want to remove %s?"), title), 100)
			note := txt.Wrap("", fmt.Sprintf(i18n.Text("Note: This action cannot be undone and will remove %s from disk."), title), 100)
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
			trimmed := strings.TrimSpace(newName)
			valid := trimmed != "" && !strings.HasPrefix(trimmed, ".") && !strings.ContainsAny(newName, `/\:`) &&
				!disallowedWindowsFileNames[strings.ToLower(newName)]
			dialog.Button(unison.ModalResponseOK).SetEnabled(valid)
			return valid
		}
		if dialog.RunModal() == unison.ModalResponseOK {
			oldPath := row.Path()
			newPath := filepath.Join(filepath.Dir(oldPath), newName+filepath.Ext(oldPath))
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
		prefix := row.library.PathOnDisk + string([]rune{filepath.Separator})
		oldPath = strings.TrimPrefix(oldPath, prefix)
		if i := slices.Index(row.library.Favorites, oldPath); i != -1 {
			row.library.Favorites = slices.Delete(row.library.Favorites, i, i+1)
			row.library.Favorites = append(row.library.Favorites, strings.TrimPrefix(newPath, prefix))
		}
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
				if strings.HasPrefix(p, oldPath) {
					fbd.SetBackingFilePath(filepath.Join(newPath, strings.TrimPrefix(p, oldPath)))
				}
			}
		}
	case row.IsFile():
		if dockable := LocateFileBackedDockable(oldPath); dockable != nil {
			dockable.SetBackingFilePath(newPath)
		}
	}
}

func (n *Navigator) updateLibrarySelection() {
	for _, row := range n.table.SelectedRows(true) {
		if row.IsLibrary() {
			_, releases := row.library.AvailableReleases()
			if len(releases) == 0 || !releases[0].HasUpdate() || !initiateLibraryUpdate(row.library, releases[0]) {
				return
			}
		}
	}
}

func (n *Navigator) showSelectionReleaseNotes() {
	for _, row := range n.table.SelectedRows(true) {
		if !row.IsLibrary() {
			continue
		}
		current, releases := row.library.AvailableReleases()
		if len(releases) == 0 || !releases[0].HasUpdate() {
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
		ShowReadOnlyMarkdown(fmt.Sprintf(i18n.Text("%s Release Notes"), row.library.Title), content.String())
	}
}

func (n *Navigator) configureSelection() {
	for _, row := range n.table.SelectedRows(true) {
		if row.IsLibrary() {
			ShowLibrarySettings(row.library)
		}
	}
}

func (n *Navigator) searchKeydown(keyCode unison.KeyCode, mod unison.Modifiers, repeat bool) bool {
	if keyCode == unison.KeyReturn || keyCode == unison.KeyNumPadEnter {
		if mod.ShiftDown() {
			n.previousMatch()
		} else {
			n.nextMatch()
		}
		return true
	}
	return n.searchField.DefaultKeyDown(keyCode, mod, repeat)
}

func (n *Navigator) tableKeyDown(keyCode unison.KeyCode, mod unison.Modifiers, repeat bool) bool {
	if unison.IsControlAction(keyCode, mod) {
		return n.table.DefaultKeyDown(keyCode, mod, repeat)
	}
	switch keyCode {
	case unison.KeyBackspace, unison.KeyDelete:
		if n.deleteButton.Enabled() {
			n.deleteButton.Click()
		}
		return true
	default:
		return n.table.DefaultKeyDown(keyCode, mod, repeat)
	}
}

func (n *Navigator) mouseDown(where unison.Point, button, clickCount int, mod unison.Modifiers) bool {
	stop := n.table.DefaultMouseDown(where, button, clickCount, mod)
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
				cm.Popup(unison.Rect{
					Point: n.table.PointToRoot(where),
					Size: unison.Size{
						Width:  1,
						Height: 1,
					},
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
				if err := desktop.Open(p); err != nil {
					Workspace.ErrorHandler(i18n.Text("Unable to show location on disk"), err)
				}
			}
		})
}

func (n *Navigator) watchCallback(_ *gurps.Library, _ string, _ notify.Event) {
	n.EventuallyReload()
}

// EventuallyReload calls Reload() after a small delay, collapsing intervening requests to do the same.
func (n *Navigator) EventuallyReload() {
	if !n.needReload {
		n.needReload = true
		unison.InvokeTaskAfter(n.Reload, time.Millisecond*100)
	}
}

// Reload the content of the navigator view.
func (n *Navigator) Reload() {
	n.contentCache = nil
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
}

func (n *Navigator) populateRows() []*NavigatorNode {
	libs := gurps.GlobalSettings().LibrarySet.List()
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
func (n *Navigator) TitleIcon(suggestedSize unison.Size) unison.Drawable {
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
					_, releases := row.library.AvailableReleases()
					downloadEnabled = len(releases) != 0 && releases[0].HasUpdate()
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
			if d, _ := row.OpenNodeContent(); !toolbox.IsNil(d) {
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

func (n *Navigator) search(text string, rows []*NavigatorNode) {
	if text == "" {
		return
	}
	for _, row := range rows {
		if row.Match(text) {
			n.searchResult = append(n.searchResult, row)
		} else if row.IsFile() {
			p := row.Path()
			content, ok := n.contentCache[p]
			if !ok {
				fi := gurps.FileInfoFor(p)
				if n.deepSearch[fi.Extensions[0]] {
					dir := os.DirFS(filepath.Dir(p))
					fileName := filepath.Base(p)
					switch fi.Extensions[0] {
					case gurps.EquipmentExt:
						if data, err := gurps.NewEquipmentFromFile(dir, fileName); err == nil {
							content = n.addToContentCache(p, prepareForContentCache(data))
						}
					case gurps.EquipmentModifiersExt:
						if data, err := gurps.NewEquipmentModifiersFromFile(dir, fileName); err == nil {
							content = n.addToContentCache(p, prepareForContentCache(data))
						}
					case gurps.NotesExt:
						if data, err := gurps.NewNotesFromFile(dir, fileName); err == nil {
							content = n.addToContentCache(p, prepareForContentCache(data))
						}
					case gurps.SheetExt:
						if data, err := gurps.NewEntityFromFile(dir, fileName); err == nil {
							for _, one := range data.Skills {
								one.TechLevel = nil
							}
							for _, one := range data.Spells {
								one.TechLevel = nil
							}
							content = n.addToContentCache(p, strings.Join([]string{
								data.Profile.Name,
								data.Profile.Age,
								data.Profile.Birthday,
								data.Profile.Eyes,
								data.Profile.Hair,
								data.Profile.Skin,
								data.Profile.Handedness,
								data.Profile.Gender,
								data.Profile.PlayerName,
								data.Profile.Title,
								data.Profile.Organization,
								data.Profile.Religion,
								prepareForContentCache(data.Traits),
								prepareForContentCache(data.Skills),
								prepareForContentCache(data.Spells),
								prepareForContentCache(data.CarriedEquipment),
								prepareForContentCache(data.OtherEquipment),
								prepareForContentCache(data.Notes),
							}, "\n"))
						}
					case gurps.SkillsExt:
						if data, err := gurps.NewSkillsFromFile(dir, fileName); err == nil {
							for _, one := range data {
								one.TechLevel = nil
							}
							content = n.addToContentCache(p, prepareForContentCache(data))
						}
					case gurps.SpellsExt:
						if data, err := gurps.NewSpellsFromFile(dir, fileName); err == nil {
							for _, one := range data {
								one.TechLevel = nil
							}
							content = n.addToContentCache(p, prepareForContentCache(data))
						}
					case gurps.TemplatesExt:
						if data, err := gurps.NewTemplateFromFile(dir, fileName); err == nil {
							for _, one := range data.Skills {
								one.TechLevel = nil
							}
							for _, one := range data.Spells {
								one.TechLevel = nil
							}
							content = n.addToContentCache(p, strings.Join([]string{
								prepareForContentCache(data.Traits),
								prepareForContentCache(data.Skills),
								prepareForContentCache(data.Spells),
								prepareForContentCache(data.Equipment),
								prepareForContentCache(data.Notes),
							}, "\n"))
						}
					case gurps.LootExt:
						if data, err := gurps.NewLootFromFile(dir, fileName); err == nil {
							content = n.addToContentCache(p, strings.Join([]string{ //nolint:gocritic // Fine as-is
								prepareForContentCache(data.Equipment),
								prepareForContentCache(data.Notes),
							}, "\n"))
						}
					// TODO: Re-enable Campaign files
					// case gurps.CampaignExt:
					// TODO: Implement
					case gurps.TraitModifiersExt:
						if data, err := gurps.NewTraitModifiersFromFile(dir, fileName); err == nil {
							content = n.addToContentCache(p, prepareForContentCache(data))
						}
					case gurps.TraitsExt:
						if data, err := gurps.NewTraitsFromFile(dir, fileName); err == nil {
							content = n.addToContentCache(p, prepareForContentCache(data))
						}
					case gurps.MarkdownExt:
						if data, err := os.ReadFile(p); err == nil {
							content = string(bytes.ToLower(data))
						}
					}
				}
			}
			content = strings.TrimSpace(content)
			if content != "" && strings.Contains(content, text) {
				n.searchResult = append(n.searchResult, row)
			}
		}
		if row.CanHaveChildren() {
			n.search(text, row.Children())
		}
	}
}

func prepareForContentCache[T gurps.NodeTypes](data []T) string {
	var buffer strings.Builder
	gurps.Traverse(func(one T) bool {
		buffer.WriteString(strings.ToLower(one.String()))
		buffer.WriteByte('\n')
		return false
	}, false, false, data...)
	return buffer.String()
}

func (n *Navigator) addToContentCache(p, content string) string {
	if n.contentCache == nil {
		n.contentCache = make(map[string]string)
	}
	n.contentCache[p] = content
	return content
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

func (n *Navigator) adjustForMatch() {
	n.backButton.SetEnabled(n.searchIndex > 0)
	n.forwardButton.SetEnabled(len(n.searchResult) != 0 && n.searchIndex != len(n.searchResult)-1)
	if len(n.searchResult) != 0 {
		if n.searchIndex < 0 {
			n.matchesLabel.SetTitle(fmt.Sprintf(i18n.Text("- of %d"), len(n.searchResult)))
		} else {
			n.matchesLabel.SetTitle(fmt.Sprintf(i18n.Text("%d of %d"), n.searchIndex+1, len(n.searchResult)))
		}
		if n.searchIndex >= 0 {
			row := n.searchResult[n.searchIndex]
			n.table.DiscloseRow(row, false)
			n.table.ClearSelection()
			i := n.table.RowToIndex(row)
			n.table.SelectByIndex(i)
			n.ValidateLayout()
			n.table.ScrollRowIntoView(i)
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
			OpenFile(p, 0)
		}
	}
}

// DisplayNewDockable adds the Dockable to the dock and gives it the focus.
func DisplayNewDockable(dockable unison.Dockable) {
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
		case fi.Extensions[0] == gurps.SheetExt:
			g := dgroup.CharacterSheets
			group = &g
		case fi.Extensions[0] == gurps.TemplatesExt:
			g := dgroup.CharacterTemplates
			group = &g
		case fi.Extensions[0] == gurps.LootExt:
			g := dgroup.LootSheets
			group = &g
		// TODO: Re-enable Campaign files
		// case fi.Extensions[0] == gurps.CampaignExt:
		// 	g := dgroup.Campaigns
		// 	group = &g
		case fi.Extensions[0] == gurps.TraitsExt,
			fi.Extensions[0] == gurps.TraitModifiersExt,
			fi.Extensions[0] == gurps.EquipmentExt,
			fi.Extensions[0] == gurps.EquipmentModifiersExt,
			fi.Extensions[0] == gurps.SkillsExt,
			fi.Extensions[0] == gurps.SpellsExt,
			fi.Extensions[0] == gurps.NotesExt:
			g := dgroup.Libraries
			group = &g
		case fi.Extensions[0] == gurps.MarkdownExt:
			g := dgroup.Markdown
			group = &g
		}
		if group != nil {
			if slices.Contains(gurps.GlobalSettings().OpenInWindow, *group) {
				if _, err := NewWindowForDockable(dockable, *group); err != nil {
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
func OpenFile(filePath string, initialPage int) (dockable unison.Dockable, wasOpen bool) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to resolve path:\n"+filePath), err)
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
		openExternalPDF(absPath, 1)
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
			trimmed := strings.TrimSpace(name)
			valid := trimmed != "" && !strings.HasPrefix(trimmed, ".") && !strings.ContainsAny(name, `/\:`) &&
				!disallowedWindowsFileNames[strings.ToLower(name)]
			if valid {
				if _, err = os.Stat(filepath.Join(parentDir, trimmed)); err == nil {
					valid = false
				}
			}
			dialog.Button(unison.ModalResponseOK).SetEnabled(valid)
			return valid
		}
		if dialog.RunModal() == unison.ModalResponseOK {
			dirPath := filepath.Join(parentDir, name)
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
