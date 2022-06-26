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
	nodeType navigatorNodeType
	open     bool
	id       uuid.UUID
	path     string
	nav      *Navigator
	library  *library.Library
	parent   *NavigatorNode
	children []*NavigatorNode
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
	if n.nodeType == libraryNode {
		return n.library.Title
	}
	return path.Base(n.path)
}

// ColumnCell implements unison.TableRowData.
func (n *NavigatorNode) ColumnCell(_, col int, foreground, _ unison.Ink, _, _, _ bool) unison.Paneler {
	if col != 0 {
		return unison.NewLabel()
	}
	title := n.CellDataForSort(col)
	var ext string
	if n.nodeType == fileNode {
		ext = strings.ToLower(path.Ext(title))
	} else if n.open {
		ext = library.OpenFolder
	} else {
		ext = library.ClosedFolder
	}
	return createNodeCell(ext, title, foreground)
}

// IsOpen implements unison.TableRowData.
func (n *NavigatorNode) IsOpen() bool {
	return n.open
}

// SetOpen implements unison.TableRowData.
func (n *NavigatorNode) SetOpen(open bool) {
	if open != n.open && n.nodeType != fileNode {
		n.open = open
		n.nav.adjustTableSize()
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
		jot.Error(errs.NewWithCausef(err, "unable to read the directory: %s", dirPath))
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
				}
			}
			if isDir {
				dirNode := NewDirectoryNode(n.nav, n.library, p, parent)
				if dirNode.recursiveFileCount() > 0 {
					children = append(children, dirNode)
				}
			} else if !library.FileInfoFor(name).IsSpecial {
				children = append(children, NewFileNode(n.library, p, parent))
			}
		}
	}
	return children
}

func (n *NavigatorNode) recursiveFileCount() int {
	count := 0
	for _, child := range n.children {
		switch child.nodeType {
		case directoryNode:
			count += child.recursiveFileCount()
		case fileNode:
			count++
		}
	}
	return count
}
