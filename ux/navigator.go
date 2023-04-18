/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ux

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/desktop"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/toolbox/xmath/geom"
	"github.com/richardwilkes/unison"
	"github.com/rjeczalik/notify"
)

const minTextWidthCandidate = "Abcdefghijklmnopqrstuvwxyz0123456789"

var _ unison.Dockable = &Navigator{}

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
	hierarchyButton           *unison.Button
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
	scroll                    *unison.ScrollPanel
	table                     *unison.Table[*NavigatorNode]
	tokens                    []*gurps.MonitorToken
	searchResult              []*NavigatorNode
	searchIndex               int
	needReload                bool
	adjustTableSizePending    bool
}

func newNavigator() *Navigator {
	n := &Navigator{
		toolbar: unison.NewPanel(),
		scroll:  unison.NewScrollPanel(),
		table:   unison.NewTable[*NavigatorNode](&unison.SimpleTableModel[*NavigatorNode]{}),
	}
	n.Self = n
	n.setupToolBar()

	n.table.Columns = make([]unison.ColumnInfo, 1)
	globalSettings := gurps.GlobalSettings()
	libs := globalSettings.LibrarySet.List()
	rows := make([]*NavigatorNode, 0, len(libs))
	n.needReload = true
	for _, lib := range libs {
		n.tokens = append(n.tokens, lib.Watch(n.watchCallback, true))
		rows = append(rows, NewLibraryNode(n, lib))
	}
	n.needReload = false
	n.table.SetScale(float32(gurps.GlobalSettings().General.NavigatorUIScale) / 100)
	n.table.SetRootRows(rows)
	n.ApplyDisclosedPaths(globalSettings.LibraryExplorer.OpenRowKeys)
	n.table.SizeColumnsToFit(true)

	n.scroll.SetContent(n.table, unison.FillBehavior, unison.FillBehavior)
	n.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
		VGrab:  true,
	})

	n.SetLayout(&unison.FlexLayout{
		Columns: 1,
		HAlign:  unison.FillAlignment,
		VAlign:  unison.FillAlignment,
	})
	n.AddChild(n.toolbar)
	n.AddChild(n.scroll)

	n.table.DoubleClickCallback = n.handleSelectionDoubleClick
	gurps.NotifyOfLibraryChangeFunc = n.EventuallyReload
	n.table.MouseDownCallback = n.mouseDown
	n.table.SelectionChangedCallback = n.selectionChanged
	n.table.KeyDownCallback = n.tableKeyDown

	n.selectionChanged()
	return n
}

func (n *Navigator) setupToolBar() {
	helpButton := unison.NewSVGButton(svg.Help)
	helpButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Help"))
	helpButton.ClickCallback = func() { HandleLink(nil, "md:Help/Interface/Library Tree") }

	n.hierarchyButton = unison.NewSVGButton(svg.Hierarchy)
	n.hierarchyButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Opens/closes all hierarchical rows"))
	n.hierarchyButton.ClickCallback = n.toggleHierarchy

	n.deleteButton = unison.NewSVGButton(svg.Trash)
	n.deleteButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Delete"))
	n.deleteButton.ClickCallback = n.deleteSelection

	n.renameButton = unison.NewSVGButton(svg.SignPost)
	n.renameButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Rename"))
	n.renameButton.ClickCallback = n.renameSelection

	n.newFolderButton = unison.NewSVGButton(svg.NewFolder)
	n.newFolderButton.Tooltip = unison.NewTooltipWithText(i18n.Text("New Folder"))
	n.newFolderButton.ClickCallback = n.newFolder

	addLibraryButton := unison.NewSVGButton(svg.CircledAdd)
	addLibraryButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Add Library"))
	addLibraryButton.ClickCallback = n.addLibrary

	n.downloadLibraryButton = unison.NewSVGButton(svg.Download)
	n.downloadLibraryButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Update"))
	n.downloadLibraryButton.ClickCallback = n.updateLibrarySelection

	n.libraryReleaseNotesButton = unison.NewSVGButton(svg.ReleaseNotes)
	n.libraryReleaseNotesButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Release Notes"))
	n.libraryReleaseNotesButton.ClickCallback = n.showSelectionReleaseNotes

	n.configLibraryButton = unison.NewSVGButton(svg.Gears)
	n.configLibraryButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Configure"))
	n.configLibraryButton.ClickCallback = n.configureSelection

	first := unison.NewPanel()
	first.AddChild(NewDefaultInfoPop())
	first.AddChild(helpButton)
	first.AddChild(
		NewScaleField(
			gurps.InitialUIScaleMin,
			gurps.InitialUIScaleMax,
			func() int { return 100 },
			func() int { return gurps.GlobalSettings().General.NavigatorUIScale },
			func(scale int) { gurps.GlobalSettings().General.NavigatorUIScale = scale },
			nil,
			false,
			n.scroll,
		),
	)
	first.AddChild(n.hierarchyButton)
	first.AddChild(NewToolbarSeparator())
	first.AddChild(addLibraryButton)
	first.AddChild(n.downloadLibraryButton)
	first.AddChild(n.libraryReleaseNotesButton)
	first.AddChild(n.configLibraryButton)
	first.AddChild(NewToolbarSeparator())
	first.AddChild(n.newFolderButton)
	first.AddChild(n.renameButton)
	first.AddChild(n.deleteButton)
	for _, child := range first.Children() {
		child.SetLayoutData(unison.MiddleAlignment)
	}
	first.SetLayout(&unison.FlowLayout{
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	first.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})

	n.backButton = unison.NewSVGButton(svg.Back)
	n.backButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Previous Match"))
	n.backButton.ClickCallback = n.previousMatch
	n.backButton.SetEnabled(false)

	n.forwardButton = unison.NewSVGButton(svg.Forward)
	n.forwardButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Next Match"))
	n.forwardButton.ClickCallback = n.nextMatch
	n.forwardButton.SetEnabled(false)

	n.searchField = unison.NewField()
	search := i18n.Text("Search")
	n.searchField.Watermark = search
	n.searchField.Tooltip = unison.NewTooltipWithText(search)
	n.searchField.ModifiedCallback = n.searchModified
	n.searchField.KeyDownCallback = n.searchKeydown
	n.searchField.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
		HGrab:  true,
	})

	n.matchesLabel = unison.NewLabel()
	n.matchesLabel.Text = "-"
	n.matchesLabel.Tooltip = unison.NewTooltipWithText(i18n.Text("Number of matches found"))

	second := unison.NewPanel()
	second.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
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

	n.toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.Insets{Bottom: 1},
		false), unison.NewEmptyBorder(unison.StdInsets())))
	n.toolbar.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: unison.StdVSpacing,
	})
	n.toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
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

func (n *Navigator) deleteSelection() {
	if n.table.HasSelection() {
		selection := n.table.SelectedRows(true)
		hasLibs := false
		hasOther := false
		title := ""
		for _, row := range selection {
			if row.nodeType == libraryNode {
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
						if row.nodeType == directoryNode {
							if err := os.RemoveAll(p); err != nil {
								unison.ErrorDialogWithError(fmt.Sprintf(i18n.Text("Unable to remove directory:\n%s"), p), err)
								return
							}
						} else {
							if err := os.Remove(p); err != nil {
								unison.ErrorDialogWithError(fmt.Sprintf(i18n.Text("Unable to remove file:\n%s"), p), err)
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
		if row.nodeType == directoryNode {
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
		if row.nodeType == libraryNode {
			return
		}

		oldName := row.primaryColumnText()
		newName := oldName

		oldField := NewStringField(nil, "", "", func() string { return oldName }, func(s string) {})
		oldField.SetEnabled(false)

		newField := NewStringField(nil, "", "", func() string { return newName }, func(s string) { newName = s })
		newField.SetMinimumTextWidthUsing(minTextWidthCandidate)

		panel := unison.NewPanel()
		panel.SetLayout(&unison.FlexLayout{
			Columns:  2,
			HSpacing: unison.StdHSpacing,
			VSpacing: unison.StdVSpacing,
		})
		panel.AddChild(NewFieldLeadingLabel(i18n.Text("Current Name")))
		panel.AddChild(oldField)
		panel.AddChild(NewFieldLeadingLabel(i18n.Text("New Name")))
		panel.AddChild(newField)

		dialog, err := unison.NewDialog(unison.DefaultDialogTheme.QuestionIcon,
			unison.DefaultDialogTheme.QuestionIconInk, panel,
			[]*unison.DialogButtonInfo{unison.NewCancelButtonInfo(), unison.NewOKButtonInfo()})
		if err != nil {
			unison.ErrorDialogWithError(i18n.Text("Unable to create rename dialog"), err)
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
				unison.ErrorDialogWithError(fmt.Sprintf(i18n.Text("Unable to rename:\n%s"), oldPath), err)
			} else {
				n.Reload()
				n.ApplySelectedPaths([]string{newPath})
				n.MarkForRedraw()
				n.adjustBackingFilePath(row, oldPath, newPath)
			}
		}
	}
}

func (n *Navigator) adjustBackingFilePath(row *NavigatorNode, oldPath, newPath string) {
	switch row.nodeType {
	case directoryNode:
		if !strings.HasSuffix(oldPath, string(os.PathSeparator)) {
			oldPath += string(os.PathSeparator)
		}
		Workspace.DocumentDock.RootDockLayout().ForEachDockContainer(func(dc *unison.DockContainer) bool {
			for _, one := range dc.Dockables() {
				if fbd, ok := one.(FileBackedDockable); ok {
					p := fbd.BackingFilePath()
					if strings.HasPrefix(p, oldPath) {
						fbd.SetBackingFilePath(filepath.Join(newPath, strings.TrimPrefix(p, oldPath)))
					}
				}
			}
			return false
		})
	case fileNode:
		if dockable := LocateFileBackedDockable(oldPath); dockable != nil {
			dockable.SetBackingFilePath(newPath)
		}
	}
}

func (n *Navigator) updateLibrarySelection() {
	for _, row := range n.table.SelectedRows(true) {
		if row.nodeType == libraryNode {
			rel := row.library.AvailableUpdate()
			if rel == nil || !rel.HasUpdate() || !initiateLibraryUpdate(row.library, *rel) {
				return
			}
		}
	}
}

func (n *Navigator) showSelectionReleaseNotes() {
	for _, row := range n.table.SelectedRows(true) {
		if row.nodeType == libraryNode {
			rel := row.library.AvailableUpdate()
			if rel == nil || !rel.HasUpdate() {
				return
			}
			ShowReadOnlyMarkdown(fmt.Sprintf("%s v%s Release Notes", row.library.Title,
				filterVersion(rel.Version)), fmt.Sprintf("## Version %s\n%s", rel.Version, rel.Notes))
		}
	}
}

func (n *Navigator) configureSelection() {
	for _, row := range n.table.SelectedRows(true) {
		if row.nodeType == libraryNode {
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
			if len(sel) == 1 && sel[0].nodeType == fileNode {
				p := sel[0].Path()
				if filepath.Ext(p) == gurps.TemplatesExt && CanApplyTemplate() {
					cm.InsertItem(-1, newApplyTemplateMenuItem(f, &id, p))
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
				cm.Popup(geom.Rect[float32]{
					Point: n.table.PointToRoot(where),
					Size: geom.Size[float32]{
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

func newApplyTemplateMenuItem(f unison.MenuFactory, id *int, templatePath string) unison.MenuItem {
	useID := *id
	*id++
	return f.NewItem(unison.PopupMenuTemporaryBaseID+useID, i18n.Text("Apply Template to Character Sheet"),
		unison.KeyBinding{}, nil, func(item unison.MenuItem) {
			ApplyTemplate(templatePath)
		})
}

func newContextMenuItemFromButton(f unison.MenuFactory, id *int, button *unison.Button) unison.MenuItem {
	if button.Enabled() {
		useID := *id
		*id++
		return f.NewItem(unison.PopupMenuTemporaryBaseID+useID,
			button.Tooltip.Children()[0].Self.(*unison.Label).Text, unison.KeyBinding{}, nil,
			func(item unison.MenuItem) { button.ClickCallback() })
	}
	return nil
}

func newShowNodeOnDiskMenuItem(f unison.MenuFactory, id *int, sel []*NavigatorNode) unison.MenuItem {
	useID := *id
	*id++
	return f.NewItem(unison.PopupMenuTemporaryBaseID+useID, i18n.Text("Show on Disk"), unison.KeyBinding{}, nil,
		func(item unison.MenuItem) {
			m := make(map[string]struct{})
			for _, node := range sel {
				p := node.Path()
				if node.nodeType == fileNode {
					p = filepath.Dir(p)
				}
				m[p] = struct{}{}
			}
			for p := range m {
				if err := desktop.Open(p); err != nil {
					unison.ErrorDialogWithError(i18n.Text("Unable to show location on disk"), err)
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
	n.needReload = false
	for _, token := range n.tokens {
		token.Stop()
	}
	n.tokens = nil
	disclosed := n.DisclosedPaths()
	selection := n.SelectedPaths()
	libs := gurps.GlobalSettings().LibrarySet.List()
	rows := make([]*NavigatorNode, 0, len(libs))
	for _, lib := range libs {
		n.tokens = append(n.tokens, lib.Watch(n.watchCallback, true))
		rows = append(rows, NewLibraryNode(n, lib))
	}
	n.table.SetRootRows(rows)
	n.ApplyDisclosedPaths(disclosed)
	n.table.SyncToModel()
	n.ApplySelectedPaths(selection)
	n.table.SizeColumnsToFit(true)
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
	if n.table.HasSelection() {
		deleteEnabled = true
		downloadEnabled = true
		configEnabled = true
		renameEnabled = n.table.SelectionCount() == 1
		newFolderEnabled = renameEnabled
		hasLibs := false
		hasOther := false
		for _, row := range n.table.SelectedRows(true) {
			if row.nodeType == libraryNode {
				renameEnabled = false
				hasLibs = true
				if row.library.IsMaster() || row.library.IsUser() {
					deleteEnabled = false
				}
				if downloadEnabled {
					rel := row.library.AvailableUpdate()
					downloadEnabled = rel != nil && rel.HasUpdate()
				}
			} else {
				hasOther = true
				configEnabled = false
				downloadEnabled = false
			}
		}
		if hasLibs && hasOther {
			deleteEnabled = false
		}
	}
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
			row.Open()
		}
	}
	if altered {
		n.table.SyncToModel()
	}
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
	n.searchIndex = 0
	n.searchResult = nil
	text := strings.ToLower(n.searchField.Text())
	for _, row := range n.table.RootRows() {
		n.search(text, row)
	}
	n.adjustForMatch()
}

func (n *Navigator) search(text string, row *NavigatorNode) {
	if row.Match(text) {
		n.searchResult = append(n.searchResult, row)
	}
	if row.CanHaveChildren() {
		for _, child := range row.Children() {
			n.search(text, child)
		}
	}
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
	n.backButton.SetEnabled(n.searchIndex != 0)
	n.forwardButton.SetEnabled(len(n.searchResult) != 0 && n.searchIndex != len(n.searchResult)-1)
	if len(n.searchResult) != 0 {
		n.matchesLabel.Text = fmt.Sprintf(i18n.Text("%d of %d"), n.searchIndex+1, len(n.searchResult))
		row := n.searchResult[n.searchIndex]
		n.table.DiscloseRow(row, false)
		n.table.ClearSelection()
		rowIndex := n.table.RowToIndex(row)
		n.table.SelectByIndex(rowIndex)
		n.table.ScrollRowIntoView(rowIndex)
	} else {
		n.matchesLabel.Text = "-"
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
	selMap := make(map[uuid.UUID]bool)
	count := n.table.LastRowIndex()
	for i := 0; i <= count; i++ {
		row := n.table.RowFromIndex(i)
		if m[row.Path()] {
			selMap[row.UUID()] = true
		}
	}
	n.table.SetSelectionMap(selMap)
}

// HandleLink will try to open http, https, and md links, as well as resolve page references.
func HandleLink(src unison.Paneler, target string) {
	if strings.HasPrefix(strings.ToLower(target), "md:") {
		if revised, err := url.PathUnescape(target); err == nil {
			target = revised
		}
	}
	if strings.HasPrefix(target, "./") || strings.HasPrefix(target, "../") {
		if md, ok := src.AsPanel().Self.(*unison.Markdown); ok && md.WorkingDir != "" {
			p := target
			if revised, err := url.PathUnescape(p); err == nil {
				p = revised
			}
			if p = filepath.Join(md.WorkingDir, p); fs.FileIsReadable(p) {
				OpenFile(p)
				return
			}
		}
		unison.ErrorDialogWithMessage(i18n.Text("Unable to open ")+target,
			i18n.Text("Does the file exist and do you have access to read it?"))
		return
	}
	OpenPageReference(target, "", nil)
}

// OpenFiles attempts to open the given file paths.
func OpenFiles(filePaths []string) {
	for _, one := range filePaths {
		if p, err := filepath.Abs(one); err != nil {
			unison.ErrorDialogWithError(i18n.Text("Unable to open ")+one, err)
		} else {
			Workspace.Window.ToFront()
			OpenFile(p)
		}
	}
}

// DisplayNewDockable adds the Dockable to the dock and gives it the focus.
func DisplayNewDockable(dockable unison.Dockable) {
	defer func() {
		if children := dockable.AsPanel().Children(); len(children) > 1 {
			FocusFirstContent(children[0], children[1])
		}
	}()
	if fbd, ok := dockable.(FileBackedDockable); ok {
		fi := gurps.FileInfoFor(fbd.BackingFilePath())
		if dc := CurrentlyFocusedDockContainer(); dc != nil && DockContainerHoldsExtension(dc, fi.GroupWith...) {
			dc.Stack(dockable, -1)
			return
		} else if dc = LocateDockContainerForExtension(fi.GroupWith...); dc != nil {
			dc.Stack(dockable, -1)
			return
		}
	}
	Workspace.DocumentDock.DockTo(dockable, nil, unison.RightSide)
}

// OpenFile attempts to open the given file path.
func OpenFile(filePath string) (dockable unison.Dockable, wasOpen bool) {
	var err error
	if filePath, err = filepath.Abs(filePath); err != nil {
		unison.ErrorDialogWithError(i18n.Text("Unable to resolve path"), err)
		return nil, false
	}
	if d := LocateFileBackedDockable(filePath); d != nil {
		ActivateDockable(d)
		return d, true
	}
	fi := gurps.FileInfoFor(filePath)
	if fi.IsSpecial {
		return nil, false
	}
	var d unison.Dockable
	if d, err = fi.Load(filePath); err != nil {
		unison.ErrorDialogWithError(i18n.Text("Unable to open file"), err)
		return nil, false
	}
	gurps.GlobalSettings().AddRecentFile(filePath)
	DisplayNewDockable(d)
	return d, false
}

func (n *Navigator) newFolder() {
	if n.table.SelectionCount() == 1 {
		row := n.table.SelectedRows(false)[0]
		parentDir := row.Path()
		if row.nodeType == fileNode {
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
		panel.AddChild(NewFieldLeadingLabel(i18n.Text("Folder Name")))
		panel.AddChild(field)

		dialog, err := unison.NewDialog(unison.DefaultDialogTheme.QuestionIcon,
			unison.DefaultDialogTheme.QuestionIconInk, panel,
			[]*unison.DialogButtonInfo{unison.NewCancelButtonInfo(), unison.NewOKButtonInfo()})
		if err != nil {
			unison.ErrorDialogWithError(i18n.Text("Unable to create new folder dialog"), err)
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
				unison.ErrorDialogWithError(fmt.Sprintf(i18n.Text("Unable to create:\n%s"), dirPath), err)
			} else {
				if row.nodeType != fileNode && !row.IsOpen() {
					row.SetOpen(true)
				}
				n.Reload()
				n.ApplySelectedPaths([]string{dirPath})
				n.MarkForRedraw()
			}
		}
	}
}
