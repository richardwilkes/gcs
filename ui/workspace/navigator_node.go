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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/txt"
	xfs "github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
)

var _ unison.TableRowData[*NavigatorNode] = &NavigatorNode{}

type navigatorNodeType uint8

const (
	libraryNode navigatorNodeType = iota
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
	library                  *library.Library
	parent                   *NavigatorNode
	children                 []*NavigatorNode
	updateCellReleaseVersion string
	updateCellCache          *updatableLibraryCell
}

// NewLibraryNode creates a new library node.
func NewLibraryNode(nav *Navigator, lib *library.Library) *NavigatorNode {
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
func NewDirectoryNode(nav *Navigator, lib *library.Library, dirPath string, parent *NavigatorNode) *NavigatorNode {
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
func NewFileNode(lib *library.Library, filePath string, parent *NavigatorNode) *NavigatorNode {
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
	if n.nodeType == libraryNode || n.nodeType == directoryNode {
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
	if n.nodeType == libraryNode {
		if n.library.IsUser() {
			return "0/" + text
		}
		if n.library.IsMaster() {
			return "2/" + text
		}
		return "1/" + text
	}
	return text
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
	if n.nodeType == libraryNode {
		if n.library.IsUser() || n.library.CachedVersion == "" || n.library.CachedVersion == "0" {
			return n.library.Title
		}
		return fmt.Sprintf("%s v%s", n.library.Title, filterVersion(n.library.CachedVersion))
	}
	return xfs.TrimExtension(path.Base(n.path))
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
		ext = library.OpenFolder
	} else {
		ext = library.ClosedFolder
	}
	size := unison.LabelFont.Size() + 5
	fi := library.FileInfoFor(ext)
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
	if n.nodeType == libraryNode {
		return n.library.Path()
	}
	return filepath.Join(n.library.Path(), n.path)
}

// Refresh the contents of this node.
func (n *NavigatorNode) Refresh() {
	switch n.nodeType {
	case libraryNode:
		n.children = n.refreshChildren(".", n)
	case directoryNode:
		n.children = n.refreshChildren(n.path, n)
	default:
	}
}

// Open the node.
func (n *NavigatorNode) Open(wnd *unison.Window) {
	if n.nodeType == fileNode {
		OpenFile(wnd, n.Path())
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
			} else if !library.FileInfoFor(name).IsSpecial {
				children = append(children, NewFileNode(n.library, p, parent))
			}
		}
	}
	return children
}
