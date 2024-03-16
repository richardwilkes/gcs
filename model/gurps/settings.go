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

package gurps

import (
	"context"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/server/websettings"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/collection/dict"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/thememode"
)

const maxRecentFiles = 20

// Last directory keys
const (
	DefaultLastDirKey  = "default"
	ImagesLastDirKey   = "images"
	SettingsLastDirKey = "settings"
)

// SettingsPath holds the path to our settings file.
var SettingsPath string

var global *Settings

// NavigatorSettings holds settings for the navigator view.
type NavigatorSettings struct {
	DividerPosition float32  `json:"divider_position"`
	OpenRowKeys     []string `json:"open_row_keys,omitempty"`
}

// Settings holds the application settings.
type Settings struct {
	LastSeenGCSVersion string                     `json:"last_seen_gcs_version,omitempty"`
	General            *GeneralSettings           `json:"general,omitempty"`
	LibrarySet         Libraries                  `json:"libraries,omitempty"`
	LibraryExplorer    NavigatorSettings          `json:"library_explorer"`
	RecentFiles        []string                   `json:"recent_files,omitempty"`
	DeepSearch         []string                   `json:"deep_search,omitempty"`
	LastDirs           map[string]string          `json:"last_dirs,omitempty"`
	ColumnSizing       map[string]map[int]float32 `json:"column_sizing,omitempty"`
	PageRefs           PageRefs                   `json:"page_refs,omitempty"`
	KeyBindings        KeyBindings                `json:"key_bindings,omitempty"`
	WorkspaceFrame     *unison.Rect               `json:"workspace_frame,omitempty"`
	Colors             Colors                     `json:"colors"`
	Fonts              Fonts                      `json:"fonts"`
	Sheet              *SheetSettings             `json:"sheet_settings,omitempty"`
	OpenInWindow       []dgroup.Group             `json:"open_in_window,omitempty"`
	WebServer          *websettings.Settings      `json:"web_server,omitempty"`
	ThemeMode          thememode.Enum             `json:"theme_mode,alt=color_mode"`
}

// DefaultSettings returns new default settings.
func DefaultSettings() *Settings {
	return &Settings{
		LastSeenGCSVersion: cmdline.AppVersion,
		General:            NewGeneralSettings(),
		LibrarySet:         NewLibraries(),
		LibraryExplorer:    NavigatorSettings{DividerPosition: 330},
		LastDirs:           make(map[string]string),
		Sheet:              FactorySheetSettings(),
		WebServer:          websettings.Default(),
	}
}

// GlobalSettings returns the global settings.
func GlobalSettings() *Settings {
	if global == nil {
		dice.GURPSFormat = true
		fixupMovedSettingsFileIfNeeded()
		if err := jio.LoadFromFile(context.Background(), SettingsPath, &global); err != nil {
			global = DefaultSettings()
		}
		global.EnsureValidity()
		InstallEvaluatorFunctions(fxp.EvalFuncs)
		unison.SetThemeMode(global.ThemeMode)
		global.Colors.MakeCurrent()
		global.Fonts.MakeCurrent()
		unison.DefaultScrollPanelTheme.MouseWheelMultiplier = func() float32 {
			return fxp.As[float32](global.General.ScrollWheelMultiplier)
		}
		unison.DefaultFieldTheme.InitialClickSelectsAll = func(_ *unison.Field) bool {
			return global.General.InitialFieldClickSelectsAll
		}
	}
	return global
}

// Save to the standard path.
func (s *Settings) Save() error {
	return jio.SaveToFile(context.Background(), SettingsPath, s)
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
	if s.Sheet == nil {
		s.Sheet = FactorySheetSettings()
	} else {
		s.Sheet.EnsureValidity()
	}
	s.OpenInWindow = SanitizeDockableGroups(s.OpenInWindow)
	if s.WebServer == nil {
		s.WebServer = websettings.Default()
	} else {
		s.WebServer.Validate()
	}
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
