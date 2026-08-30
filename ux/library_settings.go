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
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/updatecheck"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
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
	config        gurps.LibraryConfig
	path          string
	special       bool
	isUser        bool
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
	isUser := lib.IsUser()
	d := &librarySettingsDockable{
		library: lib,
		config:  lib.Config(),
		path:    lib.Data().PathOnDisk,
		special: isUser || lib.IsMaster(),
		isUser:  isUser,
	}
	d.Self = d
	d.TabTitle = librarySettingsTitle(d.config.Title)
	d.TabIcon = svg.Settings
	d.WillCloseCallback = d.willClose
	d.Setup(d.addToStartToolbar, nil, d.initContent)
	d.updateToolbar()
	d.nameField.RequestFocus()
	d.promptForSave = true
}

// librarySettingsTitle returns the title of the settings view for a library with the given name. A library created via
// Navigator.addLibrary has no name until one is typed into the view, so an empty name gets a placeholder rather than
// leaving the title to trail off after the colon.
func librarySettingsTitle(name string) string {
	if name == "" {
		name = i18n.Text("Untitled Library")
	}
	return fmt.Sprintf(i18n.Text("Library Settings: %s"), name)
}

func (d *librarySettingsDockable) addToStartToolbar(toolbar *unison.Panel) {
	d.toolbar = toolbar
	d.applyButton = unison.NewSVGButton(unison.CheckmarkSVG)
	d.applyButton.Tooltip = newWrappedTooltip(i18n.Text("Apply Changes"))
	d.applyButton.SetEnabled(false)
	d.applyButton.ClickCallback = func() {
		if d.apply() {
			d.promptForSave = false
			d.AttemptClose()
		}
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
		func() string { return d.config.Title },
		func(s string) {
			d.config.Title = strings.TrimSpace(s)
			d.updateToolbar()
		})
	d.nameField.SetEnabled(!d.special)
	if !d.special {
		d.nameField.ValidateCallback = func() bool { return d.config.Title != "" }
	}
	content.AddChild(d.nameField)

	title = i18n.Text("GitHub Account")
	content.AddChild(NewFieldLeadingLabel(title, false))
	d.githubField = NewStringField(nil, "", title,
		func() string { return d.config.GitHubAccountName },
		func(s string) {
			// Trimmed, since Library.ConfigureForKey trims when the settings are reloaded; an untrimmed name here would
			// let the collision checks pass for a key that collides once it is trimmed on the next launch.
			d.config.GitHubAccountName = strings.TrimSpace(s)
			d.updateToolbar()
		})
	d.githubField.SetEnabled(!d.special)
	if !d.special {
		d.githubField.ValidateCallback = func() bool { return !d.checkForSpecial() && !d.keyInUse() }
	}
	content.AddChild(d.githubField)

	d.addNote(content, i18n.Text("Leave the GitHub Account blank for local directories not on GitHub"))

	title = i18n.Text("GitHub Access Token")
	content.AddChild(NewFieldLeadingLabel(title, false))
	d.tokenField = NewStringField(nil, "", title,
		func() string { return d.config.AccessToken },
		func(s string) {
			d.config.AccessToken = s
			d.updateToolbar()
		})
	d.tokenField.SetEnabled(!d.special)
	content.AddChild(d.tokenField)

	d.addNote(content, i18n.Text(`The GitHub Access Token is only needed for private repositories and only needs the read-only "Content" permission for access to this repo`))

	title = i18n.Text("Repository")
	content.AddChild(NewFieldLeadingLabel(title, false))
	d.repoField = NewStringField(nil, "", title,
		func() string { return d.config.RepoName },
		func(s string) {
			d.config.RepoName = strings.TrimSpace(s) // Trimmed for the same reason as the GitHub account name above
			d.updateToolbar()
		})
	d.repoField.SetEnabled(!d.special)
	if !d.special {
		d.repoField.ValidateCallback = func() bool {
			return d.config.RepoName != "" && !d.checkForSpecial() && !d.keyInUse()
		}
	}
	content.AddChild(d.repoField)

	content.AddChild(unison.NewPanel())
	checkbox := unison.NewCheckBox()
	checkbox.SetTitle(i18n.Text("Use the most recent commit (possibly unreleased) of this repository"))
	checkbox.State = check.FromBool(d.config.UseLatest)
	checkbox.ClickCallback = func() {
		d.config.UseLatest = !d.config.UseLatest
		checkbox.State = check.FromBool(d.config.UseLatest)
		checkbox.MarkForRedraw()
		d.updateToolbar()
	}
	checkbox.SetEnabled(!d.isUser)
	content.AddChild(checkbox)

	title = i18n.Text("Path")
	content.AddChild(NewFieldLeadingLabel(title, false))
	d.pathField = NewStringField(nil, "", title,
		func() string { return d.path },
		func(s string) {
			// Trimmed like the account and repository names: a trailing space survives filepath.IsAbs and
			// filepath.Abs, so it would be recorded as-is and Library.Path would then create an empty directory
			// beside the intended one.
			d.path = strings.TrimSpace(s)
			d.updateToolbar()
		})
	d.pathField.ValidateCallback = func() bool {
		return len(d.path) > 1 && filepath.IsAbs(d.path)
	}

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
	return gurps.IsMasterLibraryKey(d.config.GitHubAccountName, d.config.RepoName) ||
		gurps.IsUserLibraryKey(d.config.GitHubAccountName, d.config.RepoName)
}

// keyInUse returns true if the account/repo pair currently in the fields already belongs to a different library in the
// global set, which the store in apply() would otherwise silently replace.
func (d *librarySettingsDockable) keyInUse() bool {
	return libraryKeyTakenByOther(gurps.GlobalSettings().Libraries, d.config.GitHubAccountName, d.config.RepoName,
		d.library)
}

// libraryKeyTakenByOther returns true if the library key formed from the given account/repo pair belongs to a library
// in the set other than the given one.
func libraryKeyTakenByOther(libs *gurps.Libraries, gitHubAccountName, repoName string, lib *gurps.Library) bool {
	existing := libs.Lookup(gitHubAccountName + "/" + repoName)
	return existing != nil && existing != lib
}

func (d *librarySettingsDockable) choosePath() {
	dlg := unison.NewOpenDialog()
	dlg.SetAllowsMultipleSelection(false)
	dlg.SetResolvesAliases(true)
	dlg.SetCanChooseDirectories(true)
	dlg.SetCanChooseFiles(false)
	usedLastDir := false
	if xos.IsDir(d.path) {
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

func (d *librarySettingsDockable) modified() bool {
	return d.library.Config() != d.config || d.library.Data().PathOnDisk != d.path
}

func (d *librarySettingsDockable) updateToolbar() {
	d.nameField.Validate()
	d.githubField.Validate()
	d.repoField.Validate()
	d.pathField.Validate()
	modified := d.modified()
	d.applyButton.SetEnabled(modified && !d.nameField.Invalid() && !d.githubField.Invalid() &&
		!d.repoField.Invalid() && !d.pathField.Invalid())
	d.cancelButton.SetEnabled(modified)
}

// willClose prompts to save any pending edits before the dockable closes, mirroring the behavior of the editors.
func (d *librarySettingsDockable) willClose() bool {
	if !d.promptForSave || !d.modified() {
		return true
	}
	// The name as edited is what the prompt refers to, since that is what would be saved: the tab title still carries
	// the name the library had when the view was opened, which is empty for a library that has just been added.
	switch unison.YesNoCancelDialog(fmt.Sprintf(i18n.Text("Save changes made to\n%s?"),
		librarySettingsTitle(d.config.Title)), "") {
	case unison.ModalResponseDiscard:
	case unison.ModalResponseOK:
		if !d.applyButton.Enabled() {
			unison.ErrorDialogWithMessage(i18n.Text("Unable to apply the changes"),
				i18n.Text("One or more fields contain invalid values."))
			return false
		}
		if !d.apply() {
			return false
		}
	default:
		return false
	}
	return true
}

// closeLibrarySettings closes any open Library Settings dockable for the given library without prompting to save. For
// use when the library has been removed from the global set, since a later apply from the dockable would put it back.
func closeLibrarySettings(lib *gurps.Library) {
	for _, dockable := range AllDockables() {
		if d, ok := dockable.AsPanel().Self.(*librarySettingsDockable); ok && d.library == lib {
			d.promptForSave = false
			d.AttemptClose()
			return
		}
	}
}

func (d *librarySettingsDockable) apply() bool {
	wnd := d.Window()
	wnd.FocusNext() // Intentionally move the focus to ensure any pending edits are flushed
	libs := gurps.GlobalSettings().Libraries
	// The validation callbacks disable the apply button when the key is taken, but the set may have changed since they
	// last ran, so check again here rather than silently replacing another library.
	if libraryKeyTakenByOther(libs, d.config.GitHubAccountName, d.config.RepoName, d.library) {
		unison.ErrorDialogWithMessage(i18n.Text("Unable to update library"),
			fmt.Sprintf(i18n.Text("Another library is already using %s/%s."), d.config.GitHubAccountName,
				d.config.RepoName))
		return false
	}
	// The deep search content loaders read the library set from background goroutines, so the re-keying goes through
	// Rekey, which swaps the keys as one locked operation rather than a Remove followed by a Store that would leave the
	// library absent from the set in between (see Libraries.Rekey).
	oldKey := d.library.Key()
	d.library.Configure(d.config)
	libs.Rekey(oldKey, d.library)
	if err := d.library.SetPath(d.path); err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to update library location"), err)
	}
	// The library may just have been named or renamed, and the tab should say so.
	d.TabTitle = librarySettingsTitle(d.config.Title)
	UpdateTitleForDockable(d)
	Workspace.Navigator.Reload()
	// A library that has just been pointed at a different repository has no releases to show until it has been checked
	// (see Library.Configure). With the periodic checks on, that is done now, in the background, so that the Library
	// Explorer's indicator is right as soon as it can be. With them off, the Library Explorer asks when its update
	// buttons are clicked, as it does for any unchecked library, rather than a check being made behind a setting that
	// says not to. A library that has already been checked and whose repository didn't change has nothing to ask.
	if libraryCheckWantedAfterApply(d.library, gurps.GlobalSettings().General.LibraryUpdateCheck) {
		go checkForLibraryUpgrade(d.library)
	}
	return true
}

// libraryCheckWantedAfterApply returns true if applying the library's settings should be followed by a background check
// of its releases.
func libraryCheckWantedAfterApply(lib *gurps.Library, option updatecheck.Option) bool {
	return option != updatecheck.Never && lib.NeedsUpgradeCheck()
}

func checkForLibraryUpgrade(lib *gurps.Library) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()
	lib.CheckForAvailableUpgrade(ctx, &http.Client{})
}
