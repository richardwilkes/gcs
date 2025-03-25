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
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	xfs "github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/slant"
)

type librarySettingsDockable struct {
	SettingsDockable
	library       *gurps.Library
	toolbar       *unison.Panel
	applyButton   *unison.Button
	cancelButton  *unison.Button
	nameField     *StringField
	githubField   *StringField
	tokenField    *StringField
	repoField     *StringField
	pathField     *StringField
	name          string
	github        string
	token         string
	repo          string
	path          string
	special       bool
	promptForSave bool
}

// ShowLibrarySettings the Library Settings view for a specific library.
func ShowLibrarySettings(lib *gurps.Library) {
	if Activate(func(d unison.Dockable) bool {
		if settingsDockable, ok := d.AsPanel().Self.(*librarySettingsDockable); ok && settingsDockable.library == lib {
			return true
		}
		return false
	}) {
		return
	}
	d := &librarySettingsDockable{
		library: lib,
		name:    lib.Title,
		github:  lib.GitHubAccountName,
		token:   lib.AccessToken,
		repo:    lib.RepoName,
		path:    lib.PathOnDisk,
		special: lib.IsMaster() || lib.IsUser(),
	}
	d.Self = d
	d.TabTitle = fmt.Sprintf(i18n.Text("Library Settings: %s"), lib.Title)
	d.TabIcon = svg.Settings
	d.Setup(d.addToStartToolbar, nil, d.initContent)
	d.updateToolbar()
	d.nameField.RequestFocus()
}

func (d *librarySettingsDockable) addToStartToolbar(toolbar *unison.Panel) {
	d.toolbar = toolbar
	d.applyButton = unison.NewSVGButton(unison.CheckmarkSVG)
	d.applyButton.Tooltip = newWrappedTooltip(i18n.Text("Apply Changes"))
	d.applyButton.SetEnabled(false)
	d.applyButton.ClickCallback = func() {
		d.apply()
		d.promptForSave = false
		d.AttemptClose()
	}
	toolbar.AddChild(d.applyButton)

	d.cancelButton = unison.NewSVGButton(svg.Not)
	d.cancelButton.Tooltip = newWrappedTooltip(i18n.Text("Discard Changes"))
	d.cancelButton.SetEnabled(false)
	d.cancelButton.ClickCallback = func() {
		d.promptForSave = false
		d.AttemptClose()
	}
	toolbar.AddChild(d.cancelButton)
}

func (d *librarySettingsDockable) initContent(content *unison.Panel) {
	content.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})

	title := i18n.Text("Name")
	content.AddChild(NewFieldLeadingLabel(title, false))
	d.nameField = NewStringField(nil, "", title,
		func() string { return d.name },
		func(s string) {
			d.name = strings.TrimSpace(s)
			d.updateToolbar()
		})
	d.nameField.SetEnabled(!d.special)
	if !d.special {
		d.nameField.ValidateCallback = func() bool { return d.name != "" }
	}
	content.AddChild(d.nameField)

	title = i18n.Text("GitHub Account")
	content.AddChild(NewFieldLeadingLabel(title, false))
	d.githubField = NewStringField(nil, "", title,
		func() string { return d.github },
		func(s string) {
			d.github = s
			d.updateToolbar()
		})
	d.githubField.SetEnabled(!d.special)
	if !d.special {
		d.githubField.ValidateCallback = func() bool { return !d.checkForSpecial() }
	}
	content.AddChild(d.githubField)

	d.addNote(content, i18n.Text("Leave the GitHub Account blank for local directories not on GitHub"))

	title = i18n.Text("GitHub Access Token")
	content.AddChild(NewFieldLeadingLabel(title, false))
	d.tokenField = NewStringField(nil, "", title,
		func() string { return d.token },
		func(s string) {
			d.token = s
			d.updateToolbar()
		})
	d.tokenField.SetEnabled(!d.special)
	content.AddChild(d.tokenField)

	d.addNote(content, i18n.Text(`The GitHub Access Token is only needed for private repositories and only needs the read-only "Content" permission for access to this repo`))

	title = i18n.Text("Repository")
	content.AddChild(NewFieldLeadingLabel(title, false))
	d.repoField = NewStringField(nil, "", title,
		func() string { return d.repo },
		func(s string) {
			d.repo = s
			d.updateToolbar()
		})
	d.repoField.SetEnabled(!d.special)
	if !d.special {
		d.repoField.ValidateCallback = func() bool { return d.repo != "" && !d.checkForSpecial() }
	}
	content.AddChild(d.repoField)

	title = i18n.Text("Path")
	content.AddChild(NewFieldLeadingLabel(title, false))
	d.pathField = NewStringField(nil, "", title,
		func() string { return d.path },
		func(s string) {
			d.path = s
			d.updateToolbar()
		})
	d.pathField.ValidateCallback = func() bool { return len(d.path) > 1 && filepath.IsAbs(d.path) }

	locateButton := unison.NewSVGButton(svg.ClosedFolder)
	locateButton.ClickCallback = d.choosePath

	wrapper := unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
	})
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	wrapper.AddChild(d.pathField)
	wrapper.AddChild(locateButton)

	content.AddChild(wrapper)

	d.addNote(content, fmt.Sprintf(i18n.Text(`Once configured, GitHub repositories will be scanned for release tags in the form "v%d.x.y" through "v%d.x.y", where x and y can be any numeric value`),
		jio.MinimumLibraryVersion, jio.CurrentDataVersion))
}

func (d *librarySettingsDockable) addNote(parent *unison.Panel, note string) {
	fd := unison.DefaultLabelTheme.Font.Descriptor()
	fd.Slant = slant.Italic
	fd.Size--
	font := fd.Font()
	for _, line := range unison.NewTextWrappedLines(note, &unison.TextDecoration{
		Font:            font,
		OnBackgroundInk: unison.DefaultLabelTheme.OnBackgroundInk,
	}, 400) {
		label := unison.NewLabel()
		label.Text = line
		parent.AddChild(unison.NewPanel())
		parent.AddChild(label)
	}
}

func (d *librarySettingsDockable) checkForSpecial() bool {
	lib := &gurps.Library{
		GitHubAccountName: d.github,
		RepoName:          d.repo,
	}
	return lib.IsMaster() || lib.IsUser()
}

func (d *librarySettingsDockable) choosePath() {
	dlg := unison.NewOpenDialog()
	dlg.SetAllowsMultipleSelection(false)
	dlg.SetResolvesAliases(true)
	dlg.SetCanChooseDirectories(true)
	dlg.SetCanChooseFiles(false)
	usedLastDir := false
	if xfs.IsDir(d.path) {
		dlg.SetInitialDirectory(filepath.Dir(d.path))
	} else {
		dlg.SetInitialDirectory(gurps.GlobalSettings().LastDir(gurps.DefaultLastDirKey))
		usedLastDir = true
	}
	if dlg.RunModal() {
		p, err := filepath.Abs(dlg.Path())
		if err != nil {
			unison.ErrorDialogWithMessage(i18n.Text("Unable to resolve absolute path"), dlg.Path())
		} else {
			if usedLastDir {
				gurps.GlobalSettings().SetLastDir(gurps.DefaultLastDirKey, filepath.Dir(p))
			}
			d.pathField.SetText(p)
		}
		d.pathField.SelectAll()
		d.pathField.RequestFocus()
	}
}

func (d *librarySettingsDockable) updateToolbar() {
	d.nameField.Validate()
	d.githubField.Validate()
	d.repoField.Validate()
	d.pathField.Validate()
	modified := d.library.Title != d.name || d.library.GitHubAccountName != d.github ||
		d.library.AccessToken != d.token || d.library.RepoName != d.repo || d.library.PathOnDisk != d.path
	d.applyButton.SetEnabled(modified && !d.nameField.Invalid() && !d.githubField.Invalid() &&
		!d.repoField.Invalid() && !d.pathField.Invalid())
	d.cancelButton.SetEnabled(modified)
}

func (d *librarySettingsDockable) apply() {
	wnd := d.Window()
	wnd.FocusNext() // Intentionally move the focus to ensure any pending edits are flushed
	libs := gurps.GlobalSettings().LibrarySet
	delete(libs, d.library.Key())
	d.library.Title = d.name
	d.library.GitHubAccountName = d.github
	d.library.AccessToken = d.token
	d.library.RepoName = d.repo
	libs[d.library.Key()] = d.library
	if err := d.library.SetPath(d.path); err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to update library location"), err)
	}
	Workspace.Navigator.Reload()
	go checkForLibraryUpgrade(d.library)
}

func checkForLibraryUpgrade(lib *gurps.Library) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()
	lib.CheckForAvailableUpgrade(ctx, &http.Client{})
}
