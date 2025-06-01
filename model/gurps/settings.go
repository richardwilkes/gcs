// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"context"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/collection/dict"
	"github.com/richardwilkes/toolbox/tid"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/thememode"
)

const maxRecentFiles = 20

// DefaultNavigatorDividerPosition is the default position for the navigator divider.
const DefaultNavigatorDividerPosition = 330

// Last directory keys
const (
	DefaultLastDirKey  = "default"
	ImagesLastDirKey   = "images"
	SettingsLastDirKey = "settings"
)

var (
	// SettingsPath holds the path to our settings file.
	SettingsPath   string
	globalOnce     sync.Once
	globalSettings Settings
)

// NavigatorSettings holds settings for the navigator view.
type NavigatorSettings struct {
	Nodes map[string]*NavNodeInfo `json:"nodes,omitempty"`
}

// NavNodeInfo holds the ID and last used timestamp for a navigator node.
type NavNodeInfo struct {
	ID       tid.TID `json:"id"`
	LastUsed int64   `json:"last"`
}

// PDFInfo holds IDs and last opened timestamp for a PDF's table of contents.
type PDFInfo struct {
	TOC        map[string]map[int]tid.TID `json:"toc,omitempty"`
	LastOpened int64                      `json:"last"`
}

// Settings holds the application settings.
type Settings struct {
	LastSeenGCSVersion string                     `json:"last_seen_gcs_version,omitempty"`
	General            *GeneralSettings           `json:"general,omitempty"`
	LibrarySet         Libraries                  `json:"libraries,omitempty"`
	LibraryExplorer    NavigatorSettings          `json:"library_explorer"`
	ThemeMode          thememode.Enum             `json:"theme_mode,alt=color_mode"`
	RecentFiles        []string                   `json:"recent_files,omitempty"`
	DeepSearch         []string                   `json:"deep_search,omitempty"`
	LastDirs           map[string]string          `json:"last_dirs,omitempty"`
	ColumnSizing       map[string]map[int]float32 `json:"column_sizing,omitempty"`
	PageRefs           PageRefs                   `json:"page_refs,omitempty"`
	KeyBindings        KeyBindings                `json:"key_bindings,omitempty"`
	WorkspaceFrame     *unison.Rect               `json:"workspace_frame,omitempty"`
	TopDockState       *unison.DockState          `json:"top_dock_state,omitempty"`
	DocDockState       *unison.DockState          `json:"doc_dock_state,omitempty"`
	Colors             colors.Colors              `json:"theme_colors"`
	Fonts              fonts.Fonts                `json:"fonts"`
	Sheet              *SheetSettings             `json:"sheet_settings,omitempty"`
	OpenInWindow       []dgroup.Group             `json:"open_in_window,omitempty"`
	Closed             map[string]int64           `json:"closed,omitempty"`
	PDFs               map[string]*PDFInfo        `json:"pdfs,omitempty"`
}

// IDer defines the methods required of objects that have an ID.
type IDer interface {
	ID() tid.TID
}

// Openable defines the methods required of openable nodes.
type Openable interface {
	IDer
	Container() bool
	IsOpen() bool
	SetOpen(open bool)
}

// GlobalSettings returns the global settings.
func GlobalSettings() *Settings {
	globalOnce.Do(func() {
		dice.GURPSFormat = true
		if err := jio.LoadFromFile(context.Background(), SettingsPath, &globalSettings); err != nil {
			globalSettings = Settings{
				LastSeenGCSVersion: cmdline.AppVersion,
				General:            NewGeneralSettings(),
				LibrarySet:         NewLibraries(),
				Sheet:              FactorySheetSettings(),
			}
		}
		globalSettings.EnsureValidity()
		unison.SetThemeMode(globalSettings.ThemeMode)
		globalSettings.Colors.MakeCurrent()
		globalSettings.Fonts.MakeCurrent()
		unison.DefaultScrollPanelTheme.MouseWheelMultiplier = func() float32 {
			return fxp.As[float32](globalSettings.General.ScrollWheelMultiplier)
		}
		unison.DefaultFieldTheme.InitialClickSelectsAll = func(_ *unison.Field) bool {
			return globalSettings.General.InitialFieldClickSelectsAll
		}
	})
	return &globalSettings
}

// Save to the standard path.
func (s *Settings) Save() error {
	cutoff := time.Now().Add(-time.Hour * 24 * 120).Unix()
	for k, v := range s.LibraryExplorer.Nodes {
		if v.LastUsed < cutoff ||
			// Also prune out old keys before we had dirs in the favorites list
			(!strings.HasPrefix(k, "F@") && !strings.HasPrefix(k, "_@")) {
			delete(s.LibraryExplorer.Nodes, k)
		}
	}
	for k, v := range s.Closed {
		if v < cutoff {
			delete(s.Closed, k)
		}
	}
	columnCutoff := ToColumnCutoff(cutoff)
	for k, v := range s.ColumnSizing {
		if last, ok := v[-1]; !ok {
			// Fixup missing last-used for old data.
			v[-1] = ToColumnCutoff(time.Now().Unix())
		} else if last < columnCutoff {
			delete(s.ColumnSizing, k)
		}
	}
	for k, v := range s.PDFs {
		if v.LastOpened < cutoff {
			delete(s.PDFs, k)
		}
	}
	return jio.SaveToFile(context.Background(), SettingsPath, s)
}

// ToColumnCutoff converts a unix timestamp (in seconds) to a column cutoff value.
func ToColumnCutoff(in int64) float32 {
	return float32(in / (60 * 60 * 24))
}

// EnsureValidity checks the current settings for validity and if they aren't valid, makes them so.
func (s *Settings) EnsureValidity() {
	if s.General == nil {
		s.General = NewGeneralSettings()
	} else {
		s.General.EnsureValidity()
	}
	if len(s.LibrarySet) == 0 {
		s.LibrarySet = NewLibraries()
	}
	if s.LastDirs == nil {
		s.LastDirs = make(map[string]string)
	}
	if s.ColumnSizing == nil {
		s.ColumnSizing = make(map[string]map[int]float32)
	}
	if s.Closed == nil {
		s.Closed = make(map[string]int64)
	}
	if s.PDFs == nil {
		s.PDFs = make(map[string]*PDFInfo)
	}
	if s.Sheet == nil {
		s.Sheet = FactorySheetSettings()
	} else {
		s.Sheet.EnsureValidity()
	}
	s.OpenInWindow = SanitizeDockableGroups(s.OpenInWindow)
}

// SanitizeDockableGroups returns the list of valid dockable groups from the passed-in list, in sorted order.
func SanitizeDockableGroups(groups []dgroup.Group) []dgroup.Group {
	m := make(map[dgroup.Group]bool)
	for _, k := range groups {
		m[k.EnsureValid()] = true
	}
	groups = dict.Keys(m)
	slices.Sort(groups)
	return groups
}

// LastDir returns the last directory used for the given key.
func (s *Settings) LastDir(key string) string {
	if last, ok := s.LastDirs[key]; ok {
		return last
	}
	var home string
	if u, err := user.Current(); err != nil {
		home = os.Getenv("HOME")
	} else {
		home = u.HomeDir
	}
	return home
}

// SetLastDir sets the last directory used for the given key. Ignores attempts to set it to an empty string.
func (s *Settings) SetLastDir(key, dir string) {
	if strings.TrimSpace(dir) != "" {
		s.LastDirs[key] = dir
	}
}

// ListRecentFiles returns the current list of recently opened files. Files that are no longer readable for any reason
// are omitted.
func (s *Settings) ListRecentFiles() []string {
	list := make([]string, 0, len(s.RecentFiles))
	for _, one := range s.RecentFiles {
		if fs.FileIsReadable(one) {
			list = append(list, one)
		}
	}
	if len(list) != len(s.RecentFiles) {
		s.RecentFiles = make([]string, len(list))
		copy(s.RecentFiles, list)
	}
	return list
}

// AddRecentFile adds a file path to the list of recently opened files.
func (s *Settings) AddRecentFile(filePath string) {
	ext := strings.ToLower(path.Ext(filePath))
	for _, one := range AcceptableExtensions() {
		if one == ext {
			full, err := filepath.Abs(filePath)
			if err != nil {
				return
			}
			if fs.FileIsReadable(full) {
				for i, f := range s.RecentFiles {
					if f == full {
						copy(s.RecentFiles[i:], s.RecentFiles[i+1:])
						s.RecentFiles[len(s.RecentFiles)-1] = ""
						s.RecentFiles = s.RecentFiles[:len(s.RecentFiles)-1]
						break
					}
				}
				s.RecentFiles = append(s.RecentFiles, "")
				copy(s.RecentFiles[1:], s.RecentFiles)
				s.RecentFiles[0] = full
				if len(s.RecentFiles) > maxRecentFiles {
					s.RecentFiles = s.RecentFiles[:maxRecentFiles]
				}
			}
			return
		}
	}
}

// GeneralSettings implements gurps.SettingsProvider.
func (s *Settings) GeneralSettings() *GeneralSettings {
	return s.General
}

// SheetSettings implements gurps.SettingsProvider.
func (s *Settings) SheetSettings() *SheetSettings {
	return s.Sheet
}

// Libraries implements gurps.SettingsProvider.
func (s *Settings) Libraries() Libraries {
	return s.LibrarySet
}

// IsNodeOpen returns true if the node is currently open.
func IsNodeOpen(node Openable) bool {
	if !node.Container() {
		return false
	}
	return !IsClosed("n:" + string(node.ID()))
}

// SetNodeOpen sets the current open state for a node. Returns true if a change was made.
func SetNodeOpen(node Openable, open bool) bool {
	if !node.Container() {
		return false
	}
	return SetClosedState("n:"+string(node.ID()), !open)
}

// IsClosed returns true if the specified key is closed.
func IsClosed(key string) bool {
	settings := GlobalSettings()
	_, closed := settings.Closed[key]
	if closed {
		settings.Closed[key] = time.Now().Unix()
	}
	return closed
}

// SetClosedState sets the current closed state for a key. Returns true if a change was made.
func SetClosedState(key string, closed bool) bool {
	settings := GlobalSettings()
	_, wasClosed := settings.Closed[key]
	if wasClosed == closed {
		if closed {
			settings.Closed[key] = time.Now().Unix()
		}
		return false
	}
	if closed {
		settings.Closed[key] = time.Now().Unix()
	} else {
		delete(settings.Closed, key)
	}
	return true
}

// IDForPDFTOC returns the ID for the specified PDF TOC entry.
func IDForPDFTOC(pdfPath, title string, pageNum int) tid.TID {
	settings := GlobalSettings()
	pi, ok := settings.PDFs[pdfPath]
	if !ok {
		pi = &PDFInfo{}
		settings.PDFs[pdfPath] = pi
	}
	pi.LastOpened = time.Now().Unix()
	if pi.TOC == nil {
		pi.TOC = make(map[string]map[int]tid.TID)
	}
	titleMap, ok := pi.TOC[title]
	if !ok {
		titleMap = make(map[int]tid.TID)
		pi.TOC[title] = titleMap
	}
	id, ok := titleMap[pageNum]
	if !ok {
		id = tid.MustNewTID(kinds.TableOfContents)
		titleMap[pageNum] = id
	}
	return id
}

// IDForNavNode returns the ID for the specified navigator node.
func IDForNavNode(fullPath string, kind byte) tid.TID {
	settings := GlobalSettings()
	if settings.LibraryExplorer.Nodes == nil {
		settings.LibraryExplorer.Nodes = make(map[string]*NavNodeInfo)
	}
	info, ok := settings.LibraryExplorer.Nodes[fullPath]
	if !ok {
		info = &NavNodeInfo{
			ID: tid.MustNewTID(kind),
		}
		settings.LibraryExplorer.Nodes[fullPath] = info
	}
	info.LastUsed = time.Now().Unix()
	return info.ID
}
