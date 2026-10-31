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
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xfilepath"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
)

var _ unison.TableRowData[*NavigatorNode] = &NavigatorNode{}

// NavigatorNode holds the favorites, a library, a directory or a file.
type NavigatorNode struct {
	id                       tid.TID
	path                     string // Relative to the library and slash-separated, as listFolder builds it
	nav                      *Navigator
	library                  *library.Library
	parent                   *NavigatorNode
	children                 []*NavigatorNode
	updateCellReleaseVersion string
	updateCellCache          *updatableLibraryCell
}

// NewFavoritesNode creates the Favorites node, with a row for each favorite a walk of the libraries listed (see
// listFavorites).
func NewFavoritesNode(nav *Navigator, favorites []*favoriteEntry) *NavigatorNode {
	n := &NavigatorNode{
		id:  "00000000000000000",
		nav: nav,
	}
	for _, one := range favorites {
		if one.entry.isDir {
			n.children = append(n.children, NewDirectoryNode(nav, one.library, one.entry.path, n, one.entry.entries))
		} else {
			n.children = append(n.children, NewFileNode(one.library, one.entry.path, n))
		}
	}
	return n
}

// NewLibraryNode creates a new library node, with rows for the entries a walk of the library listed (see listFolder).
func NewLibraryNode(nav *Navigator, lib *library.Library, entries []*navigatorEntry) *NavigatorNode {
	var id tid.TID
	switch {
	case lib.IsMaster():
		id = "10000000000000000"
	case lib.IsUser():
		id = "10000000000000001"
	default:
		if id = lib.Data().ID; id == "" {
			id = gurps.IDForNavNode(lib.Path(false), kinds.NavigatorLibrary)
			lib.SetID(id)
		}
	}
	n := &NavigatorNode{
		id:      id,
		nav:     nav,
		library: lib,
	}
	n.children = n.childrenFor(entries)
	return n
}

// NewDirectoryNode creates a new directory node. Its ID, on which the table keys the row's selection and the settings
// its open state, is kept in the settings under the folder's path on disk, marked as being under Favorites or a
// library. A folder can show up under Favorites more than once, when a favorite lies within a favorited folder, so a
// copy shown within a favorite other than the nearest one holding the folder is keyed by that favorite's path too,
// leaving a folder shown once with the key, and so the open state, it has always had.
func NewDirectoryNode(nav *Navigator, lib *library.Library, dirPath string, parent *NavigatorNode, entries []*navigatorEntry) *NavigatorNode {
	pathForID := "@" + filepath.Join(lib.Path(false), dirPath)
	root := parent
	var shownWithin *NavigatorNode // The row beneath the root that this one is shown within, unless it is that row
	for root.parent != nil {
		shownWithin = root
		root = root.parent
	}
	if root.IsFavorites() {
		pathForID = "F" + pathForID
		if shownWithin != nil && hasNearerFavorite(lib, shownWithin.path, dirPath) {
			pathForID += "\x00" + filepath.Join(lib.Path(false), shownWithin.path)
		}
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
	n.children = n.childrenFor(entries)
	return n
}

// childrenFor returns a row for each of the entries, which a walk of the libraries listed in the node's folder.
func (n *NavigatorNode) childrenFor(entries []*navigatorEntry) []*NavigatorNode {
	if len(entries) == 0 {
		return nil
	}
	children := make([]*NavigatorNode, 0, len(entries))
	for _, entry := range entries {
		if entry.isDir {
			children = append(children, NewDirectoryNode(n.nav, n.library, entry.path, n, entry.entries))
		} else {
			children = append(children, NewFileNode(n.library, entry.path, n))
		}
	}
	return children
}

// hasNearerFavorite returns true if a favorite of the library lies within the favorite folder and is the folder at
// dirPath or holds it, so that the folder shows up under Favorites within both.
func hasNearerFavorite(lib *library.Library, favorite, dirPath string) bool {
	return slices.ContainsFunc(lib.Favorites(), func(one string) bool {
		return strings.HasPrefix(one, favorite+"/") && (one == dirPath || strings.HasPrefix(dirPath, one+"/"))
	})
}

// NewFileNode creates a new file node.
func NewFileNode(lib *library.Library, filePath string, parent *NavigatorNode) *NavigatorNode {
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

// CellDataForSort implements unison.TableRowData. Nothing sorts the navigator, but a screen reader names each row by
// this text, so it is the text the row shows.
func (n *NavigatorNode) CellDataForSort(col int) string {
	if col != 0 {
		return ""
	}
	return n.primaryColumnText()
}

func filterVersion(version string) string {
	first := strings.Index(version, ".")
	switch {
	case first != -1 && first == strings.LastIndex(version, "."):
	case strings.HasSuffix(version, ".0"):
		version = version[:len(version)-2]
	case first == -1 && len(version) > 7:
		return version[:7] // Shorten long commit hashes to 7 characters to match github's style
	}
	return "v" + version
}

func (n *NavigatorNode) primaryColumnText() string {
	switch {
	case n.IsFavorites():
		return i18n.Text("Favorites")
	case n.IsLibrary():
		data := n.library.Data()
		if n.library.IsUser() {
			return data.Title
		}
		current, _ := n.library.AvailableReleases()
		if current == "" || current == "0" {
			return data.Title
		}
		if data.UseLatest {
			return data.Title + " (" + filterVersion(current) + ")"
		}
		return data.Title + " " + filterVersion(current)
	default:
		return xfilepath.TrimExtension(path.Base(n.path))
	}
}

// Match returns true if the text is present in the node. The text must already have been lowercased by the caller.
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
		Size: geom.NewSize(size, size),
	}
	if n.IsLibrary() && !n.library.IsUser() {
		if current, releases := n.library.AvailableReleases(); len(releases) != 0 && releases[0].HasUpdate() {
			if relVersion := filterVersion(releases[0].Version); filterVersion(current) != relVersion {
				if n.updateCellReleaseVersion != relVersion || n.updateCellCache == nil {
					n.updateCellReleaseVersion = relVersion
					n.updateCellCache = newUpdatableLibraryCell(n.library, label, &releases[0])
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

// placeKey returns a key for the row's place in the tree: its path on disk followed by those of its ancestors, each
// ending in a NUL, which no path holds, so that the Favorites row, whose path is empty, has a key too. The path alone
// would not do, since a file or folder can show up in several places: under its library and under Favorites, both as a
// favorite and within a favorited folder, and under a library rooted within another's folder. Unlike a file row's ID,
// the key survives a Reload, which rebuilds the tree from fresh nodes.
func (n *NavigatorNode) placeKey() string {
	var key strings.Builder
	for p := n; p != nil; p = p.parent {
		key.WriteString(p.Path())
		key.WriteByte(0)
	}
	return key.String()
}

// Path returns the full path on disk for this node.
func (n *NavigatorNode) Path() string {
	switch {
	case n.IsFavorites():
		return ""
	case n.IsLibrary():
		return n.library.Path(false)
	default:
		return filepath.Join(n.library.Path(false), n.path)
	}
}

// OpenNodeContent opens the node's content.
func (n *NavigatorNode) OpenNodeContent() (dockable unison.Dockable, wasOpen bool) {
	if n.IsFile() {
		return OpenFile(n.Path(), gurps.PageInfo{})
	}
	return nil, false
}

// navigatorEntry is one entry of a folder as a walk of the libraries listed it (see listFolder): a subfolder, with the
// entries beneath it, or a file. The walk runs on a background goroutine and reads nothing but the disk; the rows are
// built from its entries on the UI thread (see NewLibraryNode).
type navigatorEntry struct {
	path    string            // Relative to the library and slash-separated
	entries []*navigatorEntry // What a folder holds; nil for a file
	isDir   bool
}

// favoriteEntry is one favorite as a walk of the libraries listed it (see listFavorites).
type favoriteEntry struct {
	library *library.Library
	entry   *navigatorEntry
}

// libraryListing is what a walk of the libraries listed in one of them.
type libraryListing struct {
	library *library.Library
	entries []*navigatorEntry
}

// navigatorListing is what one walk of the libraries listed (see listLibraries): the entries of each library, in the
// order the libraries were given, and the favorites.
type navigatorListing struct {
	libraries []libraryListing
	favorites []*favoriteEntry
}

// librariesLister is the signature of listLibraries, which the navigator holds as a field so that tests can hold a walk
// back.
type librariesLister func(libs []*library.Library, groupContainers bool, symlinkedFolder func(lib *library.Library, dirPath string)) *navigatorListing

// listLibraries lists the given libraries and their favorites on the caller's goroutine, which Navigator.Reload makes a
// background one, so that a library on a slow disk or a network share does not stall the UI for as long as listing its
// folders takes. symlinkedFolder is told of each symlinked folder found, so that the library's watch can be made to
// follow it; see listFolder.
func listLibraries(libs []*library.Library, groupContainers bool, symlinkedFolder func(lib *library.Library, dirPath string)) *navigatorListing {
	listing := &navigatorListing{libraries: make([]libraryListing, 0, len(libs))}
	for _, lib := range libs {
		listing.libraries = append(listing.libraries, libraryListing{
			library: lib,
			entries: listFolder(lib, ".", groupContainers, symlinkedFolder),
		})
	}
	listing.favorites = listFavorites(libs, groupContainers, symlinkedFolder)
	return listing
}

// bareListing returns a listing of the given libraries with nothing listed in any of them, which is what the navigator
// shows until its first walk of the libraries returns; see Navigator.Reload.
func bareListing(libs []*library.Library) *navigatorListing {
	listing := &navigatorListing{libraries: make([]libraryListing, 0, len(libs))}
	for _, lib := range libs {
		listing.libraries = append(listing.libraries, libraryListing{library: lib})
	}
	return listing
}

// listFolder lists the folder at dirPath within the library, and the folders beneath it, in the order the tree shows
// them: folders ahead of files when groupContainers is set, within that in natural order of the name without its
// extension and then of the whole name. Entries whose names begin with a dot are left out, as are the library's
// settings and output template folders and the files of the types the file type registry marks as special, which the
// type that stands in for every unknown one is. A symlink counts as a folder when it leads to one that holds something,
// and symlinkedFolder is told of it, since the native filesystem monitoring does not follow symlinks on its own (see
// MonitorToken.AddSubPath); any other symlink is listed as a file. One that leads back to a folder it lies within is
// listed as a folder with nothing beneath it, since that folder would otherwise be listed within itself over and over
// until the path grew too long to resolve, and symlinkedFolder is not told of it, since that folder is watched already.
// A folder that cannot be read lists nothing, which is logged unless the folder is missing, as it is during filesystem
// updates.
func listFolder(lib *library.Library, dirPath string, groupContainers bool, symlinkedFolder func(lib *library.Library, dirPath string)) []*navigatorEntry {
	libPath := lib.Path(false)
	entries, err := os.ReadDir(filepath.Join(libPath, dirPath))
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			errs.Log(errs.NewWithCause("unable to read the directory", err), "dir", dirPath)
		}
		return nil
	}
	slices.SortFunc(entries, func(a, b fs.DirEntry) int {
		if groupContainers {
			if aIsDir := a.IsDir(); aIsDir != b.IsDir() {
				if aIsDir {
					return -1 // Directories before files
				}
				return 1
			}
		}
		aName := a.Name()
		bName := b.Name()
		result := xstrings.NaturalCmp(xfilepath.TrimExtension(aName), xfilepath.TrimExtension(bName), true)
		if result == 0 {
			result = xstrings.NaturalCmp(aName, bName, true)
		}
		return result
	})
	listed := make([]*navigatorEntry, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		p := path.Join(dirPath, name)
		isDir := entry.IsDir()
		leadsBack := false
		if entry.Type() == fs.ModeSymlink {
			if sub, readErr := os.ReadDir(filepath.Join(libPath, p)); readErr == nil && len(sub) > 0 {
				isDir = true
				if leadsBack = symlinkLeadsBack(libPath, dirPath, p); !leadsBack {
					symlinkedFolder(lib, p)
				}
			}
		}
		switch {
		case isDir:
			if !strings.EqualFold(p, library.SettingsDirName) && !strings.EqualFold(p, library.OutputTemplatesDirName) {
				var beneath []*navigatorEntry
				if !leadsBack {
					beneath = listFolder(lib, p, groupContainers, symlinkedFolder)
				}
				listed = append(listed, &navigatorEntry{path: p, entries: beneath, isDir: true})
			}
		case !gurps.FileInfoFor(name).IsSpecial:
			listed = append(listed, &navigatorEntry{path: p})
		}
	}
	return listed
}

// symlinkLeadsBack returns true if the symlink at p within the library at libPath leads to the folder at dirPath, which
// holds it, or to one above that, up to the library's own folder.
func symlinkLeadsBack(libPath, dirPath, p string) bool {
	target, err := os.Stat(filepath.Join(libPath, p))
	if err != nil {
		return false
	}
	for folder := dirPath; ; {
		if fi, statErr := os.Stat(filepath.Join(libPath, folder)); statErr == nil && os.SameFile(fi, target) {
			return true
		}
		parent := path.Dir(folder)
		if parent == folder {
			return false
		}
		folder = parent
	}
}

// listFavorites lists the favorites of the given libraries, after dropping from each library the favorites that can no
// longer be found (see Library.CleanupFavorites), in the order the Favorites row shows them: folders ahead of files
// when groupContainers is set, within that in natural order of the path without its extension and then of the whole
// path. A folder among them is listed as listFolder lists it.
func listFavorites(libs []*library.Library, groupContainers bool, symlinkedFolder func(lib *library.Library, dirPath string)) []*favoriteEntry {
	var favs []*favoriteEntry
	for _, lib := range libs {
		lib.CleanupFavorites()
		for _, one := range lib.Favorites() {
			favs = append(favs, &favoriteEntry{
				library: lib,
				entry:   &navigatorEntry{path: one, isDir: xos.IsDir(filepath.Join(lib.Path(false), one))},
			})
		}
	}
	slices.SortFunc(favs, func(a, b *favoriteEntry) int {
		if groupContainers && a.entry.isDir != b.entry.isDir {
			if a.entry.isDir {
				return -1 // Directories before files
			}
			return 1
		}
		result := xstrings.NaturalCmp(xfilepath.TrimExtension(a.entry.path), xfilepath.TrimExtension(b.entry.path), true)
		if result == 0 {
			result = xstrings.NaturalCmp(a.entry.path, b.entry.path, true)
		}
		return result
	})
	for _, one := range favs {
		if one.entry.isDir {
			one.entry.entries = listFolder(one.library, one.entry.path, groupContainers, symlinkedFolder)
		}
	}
	return favs
}
