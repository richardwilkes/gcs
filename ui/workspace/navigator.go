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

package workspace

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/gcs/v5/setup/trampolines"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/toolbox/desktop"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/xmath/geom"
	"github.com/richardwilkes/unison"
	"github.com/rjeczalik/notify"
)

var _ unison.Dockable = &Navigator{}

// FileBackedDockable defines methods a Dockable that is based on a file should implement.
type FileBackedDockable interface {
	unison.Dockable
	BackingFilePath() string
}

// Navigator holds the workspace navigation panel.
type Navigator struct {
	unison.Panel
	scroll     *unison.ScrollPanel
	table      *unison.Table[*NavigatorNode]
	tokens     []*library.MonitorToken
	needReload bool
}

// RegisterFileTypes registers special navigator file types.
func RegisterFileTypes() {
	registerSpecialFileInfo(library.ClosedFolder, res.ClosedFolderSVG)
	registerSpecialFileInfo(library.OpenFolder, res.OpenFolderSVG)
	registerSpecialFileInfo(library.GenericFile, res.GenericFileSVG)
}

func registerSpecialFileInfo(key string, svg *unison.SVG) {
	library.FileInfo{
		Extension: key,
		SVG:       svg,
		IsSpecial: true,
	}.Register()
}

func newNavigator() *Navigator {
	n := &Navigator{
		scroll: unison.NewScrollPanel(),
		table:  unison.NewTable[*NavigatorNode](&unison.SimpleTableModel[*NavigatorNode]{}),
	}
	n.Self = n

	n.table.ColumnSizes = make([]unison.ColumnSize, 1)
	globalSettings := settings.Global()
	libs := globalSettings.LibrarySet.List()
	rows := make([]*NavigatorNode, 0, len(libs))
	n.needReload = true
	for _, lib := range libs {
		n.tokens = append(n.tokens, lib.Watch(n.watchCallback, true))
		rows = append(rows, NewLibraryNode(n, lib))
	}
	n.needReload = false
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
	n.AddChild(n.scroll)

	n.table.DoubleClickCallback = n.handleSelectionDoubleClick
	trampolines.SetLibraryUpdatesAvailable(func() { n.table.EventuallySizeColumnsToFit(true) })
	n.table.MouseDownCallback = n.mouseDown
	return n
}

func (n *Navigator) mouseDown(where unison.Point, button, clickCount int, mod unison.Modifiers) bool {
	stop := n.table.DefaultMouseDown(where, button, clickCount, mod)
	if button == unison.ButtonRight && clickCount == 1 {
		if sel := n.table.SelectedRows(false); len(sel) != 0 {
			f := unison.DefaultMenuFactory()
			cm := f.NewMenu(unison.PopupMenuTemporaryBaseID|unison.ContextMenuIDFlag, "", nil)
			id := 1
			cm.InsertItem(-1, newShowNodeOnDiskMenuItem(f, &id, sel))
			cm.InsertItem(-1, newLibraryReleaseNotesMenuItem(f, &id, sel))
			cm.InsertItem(-1, newUpdateLibraryMenuItem(f, &id, sel))
			if cm.Count() > 0 {
				n.FlushDrawing()
				cm.Popup(geom.Rect[float32]{
					Point: n.PointToRoot(where),
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

func filterLibraries(sel []*NavigatorNode, f func(*library.Release) bool) []*library.Library {
	var libs []*library.Library
	for _, node := range sel {
		if node.nodeType == libraryNode {
			if rel := node.library.AvailableUpdate(); rel != nil && f(rel) {
				libs = append(libs, node.library)
			}
		}
	}
	return libs
}

func newLibraryReleaseNotesMenuItem(f unison.MenuFactory, id *int, sel []*NavigatorNode) unison.MenuItem {
	libs := filterLibraries(sel, func(rel *library.Release) bool { return rel.HasReleaseNotes() })
	if len(libs) == 0 {
		return nil
	}
	useID := *id
	*id++
	return f.NewItem(unison.PopupMenuTemporaryBaseID+useID, i18n.Text("Show Library Release Notes"),
		unison.KeyBinding{}, nil, func(item unison.MenuItem) {
			for _, lib := range libs {
				rel := lib.AvailableUpdate()
				trampolines.CallShowReleaseNotesMarkdown(fmt.Sprintf("%s v%s Release Notes", lib.Title,
					filterVersion(rel.Version)), fmt.Sprintf("## Version %s\n%s", rel.Version, rel.Notes))
			}
		})
}

func newUpdateLibraryMenuItem(f unison.MenuFactory, id *int, sel []*NavigatorNode) unison.MenuItem {
	libs := filterLibraries(sel, func(rel *library.Release) bool { return rel.HasUpdate() })
	var title string
	switch len(libs) {
	case 0:
		return nil
	case 1:
		title = i18n.Text("Update Library")
	default:
		title = i18n.Text("Update Libraries")
	}
	useID := *id
	*id++
	return f.NewItem(unison.PopupMenuTemporaryBaseID+useID, title, unison.KeyBinding{}, nil,
		func(item unison.MenuItem) {
			for _, lib := range libs {
				// TODO: Implement
				jot.Infof("Initiate update for %s", lib.Title)
			}
		})
}

func (n *Navigator) watchCallback(_ *library.Library, _ string, _ notify.Event) {
	n.eventuallyReload()
}

func (n *Navigator) eventuallyReload() {
	if !n.needReload {
		n.needReload = true
		unison.InvokeTaskAfter(n.reload, time.Millisecond*100)
	}
}

func (n *Navigator) reload() {
	n.needReload = false
	disclosed := n.DisclosedPaths()
	selection := n.SelectedPaths()
	libs := settings.Global().LibrarySet.List()
	rows := make([]*NavigatorNode, 0, len(libs))
	for _, one := range libs {
		rows = append(rows, NewLibraryNode(n, one))
	}
	n.table.SetRootRows(rows)
	n.ApplyDisclosedPaths(disclosed)
	n.ApplySelectedPaths(selection)
}

func (n *Navigator) adjustTableSize() {
	n.table.SyncToModel()
	n.table.SizeColumnsToFit(true)
}

// TitleIcon implements unison.Dockable
func (n *Navigator) TitleIcon(suggestedSize unison.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  unison.DocumentSVG(),
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

func (n *Navigator) handleSelectionDoubleClick() {
	window := n.Window()
	selection := n.table.SelectedRows(false)
	if len(selection) > 4 {
		if unison.QuestionDialog(i18n.Text("Are you sure you want to open all of these?"),
			fmt.Sprintf(i18n.Text("%d files will be opened."), len(selection))) != unison.ModalResponseOK {
			return
		}
	}
	for _, row := range selection {
		row.Open(window)
	}
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

// OpenFiles attempts to open the given file paths.
func OpenFiles(filePaths []string) {
	for _, wnd := range unison.Windows() {
		if ws := FromWindow(wnd); ws != nil {
			for _, one := range filePaths {
				if p, err := filepath.Abs(one); err != nil {
					unison.ErrorDialogWithError(i18n.Text("Unable to open ")+one, err)
				} else {
					OpenFile(wnd, p)
				}
			}
		}
	}
}

// DisplayNewDockable adds the Dockable to the dock and gives it the focus.
func DisplayNewDockable(wnd *unison.Window, dockable unison.Dockable) {
	ws := FromWindowOrAny(wnd)
	if ws == nil {
		ShowUnableToLocateWorkspaceError()
		return
	}
	defer func() {
		if children := dockable.AsPanel().Children(); len(children) > 1 {
			widget.FocusFirstContent(children[0], children[1])
		}
	}()
	if fbd, ok := dockable.(FileBackedDockable); ok {
		fi := library.FileInfoFor(fbd.BackingFilePath())
		if dc := ws.CurrentlyFocusedDockContainer(); dc != nil && DockContainerHoldsExtension(dc, fi.ExtensionsToGroupWith...) {
			dc.Stack(dockable, -1)
			return
		} else if dc = ws.LocateDockContainerForExtension(fi.ExtensionsToGroupWith...); dc != nil {
			dc.Stack(dockable, -1)
			return
		}
	}
	ws.DocumentDock.DockTo(dockable, nil, unison.RightSide)
}

// OpenFile attempts to open the given file path in the given window, which should contain a workspace. May pass nil for
// wnd to let it pick the first such window it discovers.
func OpenFile(wnd *unison.Window, filePath string) (dockable unison.Dockable, wasOpen bool) {
	ws := FromWindowOrAny(wnd)
	if ws == nil {
		ShowUnableToLocateWorkspaceError()
		return nil, false
	}
	var err error
	if filePath, err = filepath.Abs(filePath); err != nil {
		unison.ErrorDialogWithError(i18n.Text("Unable to resolve path"), err)
		return nil, false
	}
	if d := ws.LocateFileBackedDockable(filePath); d != nil {
		dc := unison.Ancestor[*unison.DockContainer](d)
		dc.SetCurrentDockable(d)
		dc.AcquireFocus()
		return d, true
	}
	fi := library.FileInfoFor(filePath)
	if fi.IsSpecial {
		return nil, false
	}
	var d unison.Dockable
	if d, err = fi.Load(filePath); err != nil {
		unison.ErrorDialogWithError(i18n.Text("Unable to open file"), err)
		return nil, false
	}
	settings.Global().AddRecentFile(filePath)
	DisplayNewDockable(wnd, d)
	return d, false
}
