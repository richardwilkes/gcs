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

package model

import (
	"context"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/toolbox/xio/fs/paths"
	"github.com/richardwilkes/unison"
)

const maxRecentFiles = 20

// Dir keys
const (
	DefaultLastDirKey = "default"
	ImagesDirKey      = "images"
)

var global *Settings

// NavigatorSettings holds settings for the navigator view.
type NavigatorSettings struct {
	DividerPosition float32  `json:"divider_position"`
	OpenRowKeys     []string `json:"open_row_keys,omitempty"`
}

// Settings holds the application settings.
type Settings struct {
	LastSeenGCSVersion string                `json:"last_seen_gcs_version,omitempty"`
	General            *GeneralSheetSettings `json:"general,omitempty"`
	LibrarySet         Libraries             `json:"libraries,omitempty"`
	LibraryExplorer    NavigatorSettings     `json:"library_explorer"`
	RecentFiles        []string              `json:"recent_files,omitempty"`
	LastDirs           map[string]string     `json:"last_dirs,omitempty"`
	PageRefs           PageRefs              `json:"page_refs,omitempty"`
	KeyBindings        KeyBindings           `json:"key_bindings,omitempty"`
	WorkspaceFrame     *unison.Rect          `json:"workspace_frame,omitempty"`
	Colors             Colors                `json:"colors"`
	Fonts              Fonts                 `json:"fonts"`
	QuickExports       *QuickExports         `json:"quick_exports,omitempty"`
	Sheet              *SheetSettings        `json:"sheet_settings,omitempty"`
	ColorMode          unison.ColorMode      `json:"color_mode"`
}

// DefaultSettings returns new default settings.
func DefaultSettings() *Settings {
	return &Settings{
		LastSeenGCSVersion: cmdline.AppVersion,
		General:            NewGeneralSheetSettings(),
		LibrarySet:         NewLibraries(),
		LibraryExplorer:    NavigatorSettings{DividerPosition: 330},
		LastDirs:           make(map[string]string),
		QuickExports:       NewQuickExports(),
		Sheet:              FactorySheetSettings(),
	}
}

// GlobalSettings returns the global settings.
func GlobalSettings() *Settings {
	if global == nil {
		dice.GURPSFormat = true
		if err := jio.LoadFromFile(context.Background(), SettingsPath(), &global); err != nil {
			global = DefaultSettings()
		}
		global.EnsureValidity()
		InstallEvaluatorFunctions(fxp.EvalFuncs)
		unison.SetColorMode(global.ColorMode)
		global.Colors.MakeCurrent()
		global.Fonts.MakeCurrent()
	}
	return global
}

// Save to the standard path.
func (s *Settings) Save() error {
	return jio.SaveToFile(context.Background(), SettingsPath(), s)
}

// EnsureValidity checks the current settings for validity and if they aren't valid, makes them so.
func (s *Settings) EnsureValidity() {
	if s.General == nil {
		s.General = NewGeneralSheetSettings()
	} else {
		s.General.EnsureValidity()
	}
	if len(s.LibrarySet) == 0 {
		s.LibrarySet = NewLibraries()
	}
	if s.LastDirs == nil {
		s.LastDirs = make(map[string]string)
	}
	if s.QuickExports == nil {
		s.QuickExports = NewQuickExports()
	}
	if s.Sheet == nil {
		s.Sheet = FactorySheetSettings()
	} else {
		s.Sheet.EnsureValidity()
	}
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
func (s *Settings) GeneralSettings() *GeneralSheetSettings {
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

// SettingsPath returns the path to our settings file.
func SettingsPath() string {
	return filepath.Join(paths.AppDataDir(), cmdline.AppCmdName+"_prefs.json")
}
