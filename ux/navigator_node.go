/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/txt"
	xfs "github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/slices"
)

var _ unison.TableRowData[*NavigatorNode] = &NavigatorNode{}

type navigatorNodeType uint8

const (
	favoritesNode navigatorNodeType = iota
	libraryNode
	directoryNode
	fileNode
)

// NavigatorNode holds a library, directory or file.
type NavigatorNode struct {
	nodeType                 navigatorNodeType
	open                     bool
	id                       uuid.UUID
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
		nodeType: favoritesNode,
		id:       uuid.New(),
		nav:      nav,
	}
	n.Refresh()
	return n
}

// NewLibraryNode creates a new library node.
func NewLibraryNode(nav *Navigator, lib *gurps.Library) *NavigatorNode {
	n := &NavigatorNode{
		nodeType: libraryNode,
		id:       uuid.New(),
		nav:      nav,
		library:  lib,
	}
	n.Refresh()
	return n
}

// NewDirectoryNode creates a new DirectoryNode.
func NewDirectoryNode(nav *Navigator, lib *gurps.Library, dirPath string, parent *NavigatorNode) *NavigatorNode {
	n := &NavigatorNode{
		nodeType: directoryNode,
		id:       uuid.New(),
		path:     dirPath,
		nav:      nav,
		library:  lib,
		parent:   parent,
	}
	n.Refresh()
	return n
}

// NewFileNode creates a new FileNode.
func NewFileNode(lib *gurps.Library, filePath string, parent *NavigatorNode) *NavigatorNode {
	return &NavigatorNode{
		nodeType: fileNode,
		id:       uuid.New(),
		path:     filePath,
		library:  lib,
		parent:   parent,
	}
}

// CloneForTarget implements unison.TableRowData. Not permitted at the moment.
func (n *NavigatorNode) CloneForTarget(_ unison.Paneler, _ *NavigatorNode) *NavigatorNode {
	return nil
}

// UUID implements unison.TableRowData.
func (n *NavigatorNode) UUID() uuid.UUID {
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
	if n.nodeType == libraryNode || n.nodeType == directoryNode || n.nodeType == favoritesNode {
		return true
	}
	return false
}

// Children implements unison.TableRowData.
func (n *NavigatorNode) Children() []*NavigatorNode {
	return n.children
}

// SetChildren implements unison.TableRowData.
func (n *NavigatorNode) SetChildren(_ []*NavigatorNode) {
}

// CellDataForSort implements unison.TableRowData.
func (n *NavigatorNode) CellDataForSort(col int) string {
	if col != 0 {
		return ""
	}
	text := n.primaryColumnText()
	switch n.nodeType {
	case favoritesNode:
		return "0/" + text
	case libraryNode:
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
	switch n.nodeType {
	case favoritesNode:
		return i18n.Text("Favorites")
	case libraryNode:
		if n.library.IsUser() || n.library.CachedVersion == "" || n.library.CachedVersion == "0" {
			return n.library.Title
		}
		return fmt.Sprintf("%s v%s", n.library.Title, filterVersion(n.library.CachedVersion))
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
	if n.nodeType == fileNode {
		ext = strings.ToLower(path.Ext(n.path))
	} else if n.open {
		ext = gurps.OpenFolder
	} else {
		ext = gurps.ClosedFolder
	}
	size := unison.LabelFont.Size() + 5
	fi := gurps.FileInfoFor(ext)
	label := unison.NewLabel()
	label.OnBackgroundInk = foreground
	label.Text = title
	label.Drawable = &unison.DrawableSVG{
		SVG:  fi.SVG,
		Size: unison.NewSize(size, size),
	}
	if n.nodeType == libraryNode && !n.library.IsUser() {
		if rel := n.library.AvailableUpdate(); rel != nil && rel.HasUpdate() {
			if relVersion := filterVersion(rel.Version); filterVersion(n.library.CachedVersion) != relVersion {
				if n.updateCellReleaseVersion != relVersion || n.updateCellCache == nil {
					n.updateCellReleaseVersion = relVersion
					n.updateCellCache = newUpdatableLibraryCell(n.library, label, rel)
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
	return n.open
}

// SetOpen implements unison.TableRowData.
func (n *NavigatorNode) SetOpen(open bool) {
	if open != n.open && n.nodeType != fileNode {
		n.open = open
		n.nav.adjustTableSizeEventually()
	}
}

// Path returns the full path on disk for this node.
func (n *NavigatorNode) Path() string {
	switch n.nodeType {
	case favoritesNode:
		return ""
	case libraryNode:
		return n.library.Path()
	default:
		return filepath.Join(n.library.Path(), n.path)
	}
}

// Refresh the contents of this node.
func (n *NavigatorNode) Refresh() {
	switch n.nodeType {
	case favoritesNode:
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
		slices.SortFunc(favs, func(a, b *fav) int { return txt.NaturalCmp(a.path, b.path, true) })
		for _, one := range favs {
			n.children = append(n.children, NewFileNode(one.library, one.path, n))
		}
	case libraryNode:
		n.children = n.refreshChildren(".", n)
	case directoryNode:
		n.children = n.refreshChildren(n.path, n)
	default:
	}
}

// Open the node.
func (n *NavigatorNode) Open() {
	if n.nodeType == fileNode {
		OpenFile(n.Path(), 0)
	}
}

func (n *NavigatorNode) refreshChildren(dirPath string, parent *NavigatorNode) []*NavigatorNode {
	libPath := n.library.Path()
	entries, err := os.ReadDir(filepath.Join(libPath, dirPath))
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			// Only log the error if it wasn't due to a missing dir, since that happens during filesystem updates
			jot.Error(errs.NewWithCausef(err, "unable to read the directory: %s", dirPath))
		}
		return nil
	}
	sort.Slice(entries, func(i, j int) bool {
		return txt.NaturalLess(entries[i].Name(), entries[j].Name(), true)
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
