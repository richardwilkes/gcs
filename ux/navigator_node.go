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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/tid"
	"github.com/richardwilkes/toolbox/txt"
	xfs "github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
)

var _ unison.TableRowData[*NavigatorNode] = &NavigatorNode{}

// NavigatorNode holds a library, directory or file.
type NavigatorNode struct {
	id                       tid.TID
	path                     string
	nav                      *Navigator
	library                  *gurps.Library
	parent                   *NavigatorNode
	children                 []*NavigatorNode
	updateCellReleaseVersion string
	updateCellCache          *updatableLibraryCell
}

// NewFavoritesNode creates the Favorites node.
func NewFavoritesNode(nav *Navigator) *NavigatorNode {
	n := &NavigatorNode{
		id:  "00000000000000000",
		nav: nav,
	}
	n.Refresh()
	return n
}

// NewLibraryNode creates a new library node.
func NewLibraryNode(nav *Navigator, lib *gurps.Library) *NavigatorNode {
	var id tid.TID
	switch {
	case lib.IsMaster():
		id = "10000000000000000"
	case lib.IsUser():
		id = "10000000000000001"
	default:
		if lib.ID == "" {
			lib.ID = gurps.IDForNavNode(lib.Path(), kinds.NavigatorLibrary)
		}
		id = lib.ID
	}
	n := &NavigatorNode{
		id:      id,
		nav:     nav,
		library: lib,
	}
	n.Refresh()
	return n
}

// NewDirectoryNode creates a new DirectoryNode.
func NewDirectoryNode(nav *Navigator, lib *gurps.Library, dirPath string, parent *NavigatorNode) *NavigatorNode {
	pathForID := "@" + filepath.Join(lib.Path(), dirPath)
	root := parent
	for root.parent != nil {
		root = root.parent
	}
	if root.IsFavorites() {
		pathForID = "F" + pathForID
	} else {
		pathForID = "_" + pathForID
	}
	n := &NavigatorNode{
		id:      gurps.IDForNavNode(pathForID, kinds.NavigatorDirectory),
		path:    dirPath,
		nav:     nav,
		library: lib,
		parent:  parent,
	}
	n.Refresh()
	return n
}

// NewFileNode creates a new FileNode.
func NewFileNode(lib *gurps.Library, filePath string, parent *NavigatorNode) *NavigatorNode {
	return &NavigatorNode{
		id:      tid.MustNewTID(kinds.NavigatorFile),
		path:    filePath,
		library: lib,
		parent:  parent,
	}
}

// CloneForTarget implements unison.TableRowData. Not permitted at the moment.
func (n *NavigatorNode) CloneForTarget(_ unison.Paneler, _ *NavigatorNode) *NavigatorNode {
	return nil
}

// ID implements unison.TableRowData.
func (n *NavigatorNode) ID() tid.TID {
	return n.id
}

// Parent implements unison.TableRowData.
func (n *NavigatorNode) Parent() *NavigatorNode {
	return n.parent
}

// SetParent implements unison.TableRowData.
func (n *NavigatorNode) SetParent(_ *NavigatorNode) {
}

// CanHaveChildren implements unison.TableRowData.
func (n *NavigatorNode) CanHaveChildren() bool {
	return !n.IsFile()
}

// Children implements unison.TableRowData.
func (n *NavigatorNode) Children() []*NavigatorNode {
	return n.children
}

// SetChildren implements unison.TableRowData.
func (n *NavigatorNode) SetChildren(_ []*NavigatorNode) {
}

// IsFile returns true if this is a file node.
func (n *NavigatorNode) IsFile() bool {
	return tid.IsKind(n.id, kinds.NavigatorFile)
}

// IsDirectory returns true if this is a directory node.
func (n *NavigatorNode) IsDirectory() bool {
	return tid.IsKind(n.id, kinds.NavigatorDirectory)
}

// IsLibrary returns true if this is a library node.
func (n *NavigatorNode) IsLibrary() bool {
	return tid.IsKind(n.id, kinds.NavigatorLibrary)
}

// IsFavorites returns true if this is a favorites node.
func (n *NavigatorNode) IsFavorites() bool {
	return tid.IsKind(n.id, kinds.NavigatorFavorites)
}

// CellDataForSort implements unison.TableRowData.
func (n *NavigatorNode) CellDataForSort(col int) string {
	if col != 0 {
		return ""
	}
	text := n.primaryColumnText()
	switch {
	case n.IsFavorites():
		return "0/" + text
	case n.IsLibrary():
		if n.library.IsUser() {
			return "1/" + text
		}
		if n.library.IsMaster() {
			return "3/" + text
		}
		return "2/" + text
	default:
		return text
	}
}

func filterVersion(version string) string {
	if strings.Index(version, ".") == strings.LastIndex(version, ".") {
		return version
	}
	if strings.HasSuffix(version, ".0") {
		return version[:len(version)-2]
	}
	return version
}

func (n *NavigatorNode) primaryColumnText() string {
	switch {
	case n.IsFavorites():
		return i18n.Text("Favorites")
	case n.IsLibrary():
		if n.library.IsUser() {
			return n.library.Title
		}
		current, _ := n.library.AvailableReleases()
		if current == "" || current == "0" {
			return n.library.Title
		}
		return fmt.Sprintf("%s v%s", n.library.Title, filterVersion(current))
	default:
		return xfs.TrimExtension(path.Base(n.path))
	}
}

// Match looks for the text in the node and return true if it is present. Note that calls to this method should always
// pass in text that has already been run through strings.ToLower().
func (n *NavigatorNode) Match(text string) bool {
	if text == "" {
		return false
	}
	return strings.Contains(strings.ToLower(n.primaryColumnText()), text)
}

// ColumnCell implements unison.TableRowData.
func (n *NavigatorNode) ColumnCell(_, col int, foreground, _ unison.Ink, _, _, _ bool) unison.Paneler {
	if col != 0 {
		return unison.NewLabel()
	}
	title := n.primaryColumnText()
	var ext string
	switch {
	case n.IsFile():
		ext = strings.ToLower(path.Ext(n.path))
	case n.IsOpen():
		ext = gurps.OpenFolder
	default:
		ext = gurps.ClosedFolder
	}
	size := unison.LabelFont.Size() + 5
	fi := gurps.FileInfoFor(ext)
	label := unison.NewLabel()
	label.OnBackgroundInk = foreground
	label.SetTitle(title)
	label.Drawable = &unison.DrawableSVG{
		SVG:  fi.SVG,
		Size: unison.NewSize(size, size),
	}
	if n.IsLibrary() && !n.library.IsUser() {
		if current, releases := n.library.AvailableReleases(); len(releases) != 0 && releases[0].HasUpdate() {
			if relVersion := filterVersion(releases[0].Version); filterVersion(current) != relVersion {
				if n.updateCellReleaseVersion != relVersion || n.updateCellCache == nil {
					n.updateCellReleaseVersion = relVersion
					n.updateCellCache = newUpdatableLibraryCell(n.library, label, releases[0])
				} else {
					n.updateCellCache.updateForeground(foreground)
				}
				return n.updateCellCache
			}
		}
	}
	return label
}

// IsOpen implements unison.TableRowData.
func (n *NavigatorNode) IsOpen() bool {
	return gurps.IsNodeOpen(n)
}

// SetOpen implements unison.TableRowData.
func (n *NavigatorNode) SetOpen(open bool) {
	if gurps.SetNodeOpen(n, open) {
		n.nav.adjustTableSizeEventually()
	}
}

// Container returns true if this node can have children.
func (n *NavigatorNode) Container() bool {
	return n.CanHaveChildren()
}

// Path returns the full path on disk for this node.
func (n *NavigatorNode) Path() string {
	switch {
	case n.IsFavorites():
		return ""
	case n.IsLibrary():
		return n.library.Path()
	default:
		return filepath.Join(n.library.Path(), n.path)
	}
}

// Refresh the contents of this node.
func (n *NavigatorNode) Refresh() {
	switch {
	case n.IsFavorites():
		type fav struct {
			path    string
			library *gurps.Library
		}
		var favs []*fav
		for _, lib := range gurps.GlobalSettings().LibrarySet {
			lib.CleanupFavorites()
			if len(lib.Favorites) != 0 {
				for _, one := range lib.Favorites {
					favs = append(favs, &fav{
						path:    one,
						library: lib,
					})
				}
			}
		}
		groupContainers := gurps.GlobalSettings().General.GroupContainersOnSort
		slices.SortFunc(favs, func(a, b *fav) int {
			if groupContainers {
				aIsDir := xfs.IsDir(filepath.Join(a.library.Path(), a.path))
				if aIsDir != xfs.IsDir(filepath.Join(b.library.Path(), b.path)) {
					if aIsDir {
						return -1 // Directories before files
					}
					return 1 // Files after directories
				}
			}
			result := txt.NaturalCmp(xfs.TrimExtension(a.path), xfs.TrimExtension(b.path), true)
			if result == 0 {
				result = txt.NaturalCmp(a.path, b.path, true)
			}
			return result
		})
		for _, one := range favs {
			p := filepath.Join(one.library.Path(), one.path)
			if xfs.IsDir(p) {
				n.children = append(n.children, NewDirectoryNode(n.nav, one.library, one.path, n))
			} else {
				n.children = append(n.children, NewFileNode(one.library, one.path, n))
			}
		}
	case n.IsLibrary():
		n.children = n.refreshChildren(".", n)
	case n.IsDirectory():
		n.children = n.refreshChildren(n.path, n)
	default:
	}
}

// OpenNodeContent opens the node's content.
func (n *NavigatorNode) OpenNodeContent() (dockable unison.Dockable, wasOpen bool) {
	if n.IsFile() {
		return OpenFile(n.Path(), 0)
	}
	return nil, false
}

func (n *NavigatorNode) refreshChildren(dirPath string, parent *NavigatorNode) []*NavigatorNode {
	libPath := n.library.Path()
	entries, err := os.ReadDir(filepath.Join(libPath, dirPath))
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			// Only log the error if it wasn't due to a missing dir, since that happens during filesystem updates
			errs.Log(errs.NewWithCause("unable to read the directory", err), "dir", dirPath)
		}
		return nil
	}
	groupContainers := gurps.GlobalSettings().General.GroupContainersOnSort
	slices.SortFunc(entries, func(a, b fs.DirEntry) int {
		if groupContainers {
			if aIsDir := a.IsDir(); aIsDir != b.IsDir() {
				if aIsDir {
					return -1 // Directories before files
				}
				return 1 // Files after directories
			}
		}
		aName := a.Name()
		bName := b.Name()
		result := txt.NaturalCmp(xfs.TrimExtension(aName), xfs.TrimExtension(bName), true)
		if result == 0 {
			result = txt.NaturalCmp(aName, bName, true)
		}
		return result
	})
	children := make([]*NavigatorNode, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, ".") {
			p := path.Join(dirPath, name)
			isDir := entry.IsDir()
			if entry.Type() == fs.ModeSymlink {
				var sub []fs.DirEntry
				if sub, err = os.ReadDir(filepath.Join(libPath, p)); err == nil && len(sub) > 0 {
					isDir = true
					for _, token := range n.nav.tokens {
						if token.Library() == n.library {
							token.AddSubPath(p)
						}
					}
				}
			}
			if isDir {
				if !strings.EqualFold(p, "Settings") && !strings.EqualFold(p, "Output Templates") {
					dirNode := NewDirectoryNode(n.nav, n.library, p, parent)
					children = append(children, dirNode)
				}
			} else if !gurps.FileInfoFor(name).IsSpecial {
				children = append(children, NewFileNode(n.library, p, parent))
			}
		}
	}
	return children
}
