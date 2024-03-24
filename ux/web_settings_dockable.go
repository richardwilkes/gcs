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
	"io/fs"
	"net"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/server/state"
	"github.com/richardwilkes/gcs/v5/server/websettings"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	xfs "github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

// These need to be initialized by whatever instantiates the ux package, typically main.go. They are here to break the
// circular reference that would otherwise occur.
var (
	StartServer func(func(error))
	StopServer  func()
)

type webSettingsDockable struct {
	SettingsDockable
	errorMsg                 *unison.Label
	enabledCheckbox          *CheckBox
	addressField             *StringField
	certFileField            *StringField
	certFileButton           *unison.Button
	keyFileField             *StringField
	keyFileButton            *unison.Button
	shutdownGracePeriodField *DecimalField
	readTimeoutField         *DecimalField
	writeTimeoutField        *DecimalField
	idleTimeoutField         *DecimalField
	userList                 *unison.List[*websettings.User]
	userAddButton            *unison.Button
	userNameField            *StringField
	passwordField            *StringField
	originalName             string
	accessList               *unison.List[*websettings.AccessWithKey]
	accessListUser           *websettings.User
	accessOriginal           websettings.AccessWithKey
	accessKeyField           *StringField
	accessDirField           *StringField
	accessDialog             *unison.Dialog
	userDialog               *unison.Dialog
	waitingForSync           atomic.Bool
}

// ShowWebSettings the Web Settings window.
func ShowWebSettings() {
	if Activate(func(d unison.Dockable) bool {
		_, ok := d.AsPanel().Self.(*webSettingsDockable)
		return ok
	}) {
		return
	}
	d := &webSettingsDockable{}
	d.Self = d
	d.TabTitle = i18n.Text("Web Settings")
	d.TabIcon = svg.Settings
	d.Extensions = []string{gurps.WebSettingsExt}
	d.Loader = d.load
	d.Saver = d.save
	d.Resetter = d.reset
	d.Setup(d.addToStartToolbar, nil, d.initContent)
	d.addressField.RequestFocus()
}

func (d *webSettingsDockable) addToStartToolbar(toolbar *unison.Panel) {
	helpButton := unison.NewSVGButton(svg.Help)
	helpButton.Tooltip = newWrappedTooltip(i18n.Text("Help"))
	helpButton.ClickCallback = func() { HandleLink(nil, "md:Help/Interface/Web Settings") }
	toolbar.AddChild(helpButton)
}

func (d *webSettingsDockable) initContent(content *unison.Panel) {
	content.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	d.createEnabledCheckbox(content)
	d.createAddressField(content)
	d.createCertFileField(content)
	d.createKeyFileField(content)
	d.createShutdownGracePeriodField(content)
	d.createReadTimeoutField(content)
	d.createWriteTimeoutField(content)
	d.createIdleTimeoutField(content)
	d.createUsersBlock(content)
	d.syncEnablementToServer(nil)
}

func (d *webSettingsDockable) syncEnablementToServer(callback func()) {
	if !d.waitingForSync.Swap(true) {
		d.setWebServerControlEnablement(false)
		go d.finishSync(callback)
	}
}

func (d *webSettingsDockable) finishSync(callback func()) {
	state.WaitUntil(state.Running, state.Stopped)
	unison.InvokeTask(func() {
		settings := gurps.GlobalSettings().WebServer
		switch state.Current() {
		case state.Stopped:
			settings.Enabled = false
			SetCheckBoxState(d.enabledCheckbox, false)
			d.setWebServerControlEnablement(true)
		case state.Running:
			settings.Enabled = true
			d.enabledCheckbox.SetEnabled(true)
			SetCheckBoxState(d.enabledCheckbox, true)
		default:
			go d.finishSync(callback)
			return
		}
		d.waitingForSync.Store(false)
		if callback != nil {
			callback()
		}
	})
}

func (d *webSettingsDockable) setWebServerControlEnablement(enabled bool) {
	d.enabledCheckbox.SetEnabled(enabled)
	d.addressField.SetEnabled(enabled)
	d.certFileField.SetEnabled(enabled)
	d.certFileButton.SetEnabled(enabled)
	d.keyFileField.SetEnabled(enabled)
	d.keyFileButton.SetEnabled(enabled)
	d.shutdownGracePeriodField.SetEnabled(enabled)
	d.readTimeoutField.SetEnabled(enabled)
	d.writeTimeoutField.SetEnabled(enabled)
	d.idleTimeoutField.SetEnabled(enabled)
	d.userList.SetEnabled(enabled)
	d.userAddButton.SetEnabled(enabled)
}

func (d *webSettingsDockable) updateErrorMsg(err error) {
	parent := d.enabledCheckbox.Parent()
	if parent == nil {
		return
	}
	if err == nil {
		if d.errorMsg != nil {
			d.errorMsg.RemoveFromParent()
			d.errorMsg = nil
		}
	} else {
		if d.errorMsg == nil {
			d.errorMsg = unison.NewLabel()
			d.errorMsg.SetBorder(unison.NewEmptyBorder(unison.NewUniformInsets(2)))
			d.errorMsg.HAlign = align.Middle
			d.errorMsg.OnBackgroundInk = unison.OnErrorColor
			d.errorMsg.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
				gc.DrawRect(rect, unison.ErrorColor.Paint(gc, rect, paintstyle.Fill))
				d.errorMsg.DefaultDraw(gc, rect)
			}
			d.errorMsg.SetLayoutData(&unison.FlexLayoutData{
				HSpan:  2,
				HAlign: align.Fill,
				HGrab:  true,
			})
			parent.AddChildAtIndex(d.errorMsg, 1+parent.IndexOfChild(d.enabledCheckbox))
		}
		d.errorMsg.Text = strings.SplitN(err.Error(), "\n", 2)[0]
	}
	parent.MarkForLayoutRecursivelyUpward()
	parent.MarkForRedraw()
	if err != nil {
		unison.Beep()
	}
}

func (d *webSettingsDockable) createEnabledCheckbox(content *unison.Panel) {
	d.enabledCheckbox = NewCheckBox(nil, "", i18n.Text("Enable Web Server"),
		func() check.Enum { return check.FromBool(gurps.GlobalSettings().WebServer.Enabled) },
		func(state check.Enum) { d.applyServerEnabled(state == check.On) })
	d.enabledCheckbox.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Middle,
	})
	d.enabledCheckbox.SetBorder(unison.NewEmptyBorder(unison.Insets{Bottom: unison.StdVSpacing * 2}))
	content.AddChild(d.enabledCheckbox)
}

func (d *webSettingsDockable) applyServerEnabled(on bool) {
	settings := gurps.GlobalSettings().WebServer
	if on && !settings.Valid() || (settings.CertFile == "") != (settings.KeyFile == "") {
		on = false
		d.updateErrorMsg(errs.New("Invalid web settings"))
	}
	if settings.Enabled != on {
		settings.Enabled = on
		if on {
			d.updateErrorMsg(nil)
			StartServer(d.updateErrorMsg)
		} else {
			StopServer()
		}
	}
	d.syncEnablementToServer(nil)
}

func (d *webSettingsDockable) createAddressField(content *unison.Panel) {
	title := i18n.Text("Server Address")
	content.AddChild(NewFieldLeadingLabel(title, false))
	d.addressField = NewStringField(nil, "", title,
		func() string { return gurps.GlobalSettings().WebServer.Address },
		func(s string) { gurps.GlobalSettings().WebServer.Address = s })
	d.addressField.ValidateCallback = func() bool {
		return isAddressValid(d.addressField.Text())
	}
	d.addressField.Validate()
	content.AddChild(d.addressField)
	content.AddChild(unison.NewPanel())
	note := unison.NewLabel()
	note.Text = i18n.Text(`Provide just a colon followed by a port number (e.g. ":8422") to listen on all available addresses.`)
	desc := note.Font.Descriptor()
	desc.Size -= 2
	note.Font = desc.Font()
	content.AddChild(note)
}

func isAddressValid(address string) bool {
	_, _, err := net.SplitHostPort(address)
	return err == nil
}

func (d *webSettingsDockable) createCertFileField(content *unison.Panel) {
	d.certFileField, d.certFileButton = createFilePathField(content, i18n.Text("Certificate File"),
		func() string { return gurps.GlobalSettings().WebServer.CertFile },
		func(s string) { gurps.GlobalSettings().WebServer.CertFile = s }, false)
}

func (d *webSettingsDockable) createKeyFileField(content *unison.Panel) {
	d.keyFileField, d.keyFileButton = createFilePathField(content, i18n.Text("Key File"),
		func() string { return gurps.GlobalSettings().WebServer.KeyFile },
		func(s string) { gurps.GlobalSettings().WebServer.KeyFile = s }, false)
}

func createFilePathField(content *unison.Panel, title string, get func() string, set func(string), forDirs bool) (*StringField, *unison.Button) {
	content.AddChild(NewFieldLeadingLabel(title, false))
	fileField := NewStringField(nil, "", title, get, set)
	fileField.ValidateCallback = func() bool {
		p := get()
		if p == "" {
			return true
		}
		return filepath.IsAbs(p) && xfs.FileIsReadable(p)
	}
	var icon *unison.SVG
	if forDirs {
		icon = svg.ClosedFolder
	} else {
		icon = svg.GenericFile
	}
	locateButton := unison.NewSVGButton(icon)
	locateButton.ClickCallback = func() { chooseFilePath(fileField, forDirs) }
	content.AddChild(WrapWithSpan(1, fileField, locateButton))
	fileField.Validate()
	return fileField, locateButton
}

func chooseFilePath(field *StringField, forDirs bool) {
	dlg := unison.NewOpenDialog()
	dlg.SetAllowsMultipleSelection(false)
	dlg.SetResolvesAliases(true)
	dlg.SetCanChooseDirectories(forDirs)
	dlg.SetCanChooseFiles(!forDirs)
	usedLastDir := false
	currentPath := field.Text()
	if xfs.FileExists(currentPath) {
		dlg.SetInitialDirectory(filepath.Dir(currentPath))
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
			field.SetText(p)
		}
		field.SelectAll()
		field.RequestFocus()
	}
}

func (d *webSettingsDockable) createShutdownGracePeriodField(content *unison.Panel) {
	d.shutdownGracePeriodField = createSecondsField(content, i18n.Text("Shutdown Grace Period"),
		func() fxp.Int { return gurps.GlobalSettings().WebServer.ShutdownGracePeriod },
		func(v fxp.Int) { gurps.GlobalSettings().WebServer.ShutdownGracePeriod = v },
		websettings.MinimumTimeout, websettings.MaximumTimeout)
}

func (d *webSettingsDockable) createReadTimeoutField(content *unison.Panel) {
	d.readTimeoutField = createSecondsField(content, i18n.Text("Read Timeout"),
		func() fxp.Int { return gurps.GlobalSettings().WebServer.ReadTimeout },
		func(v fxp.Int) { gurps.GlobalSettings().WebServer.ReadTimeout = v },
		websettings.MinimumTimeout, websettings.MaximumTimeout)
}

func (d *webSettingsDockable) createWriteTimeoutField(content *unison.Panel) {
	d.writeTimeoutField = createSecondsField(content, i18n.Text("Write Timeout"),
		func() fxp.Int { return gurps.GlobalSettings().WebServer.WriteTimeout },
		func(v fxp.Int) { gurps.GlobalSettings().WebServer.WriteTimeout = v },
		websettings.MinimumTimeout, websettings.MaximumTimeout)
}

func (d *webSettingsDockable) createIdleTimeoutField(content *unison.Panel) {
	d.idleTimeoutField = createSecondsField(content, i18n.Text("Idle Timeout"),
		func() fxp.Int { return gurps.GlobalSettings().WebServer.IdleTimeout },
		func(v fxp.Int) { gurps.GlobalSettings().WebServer.IdleTimeout = v },
		websettings.MinimumTimeout, websettings.MaximumTimeout)
}

func createSecondsField(content *unison.Panel, title string, get func() fxp.Int, set func(fxp.Int), miniumum, maximum fxp.Int) *DecimalField {
	content.AddChild(NewFieldLeadingLabel(title, false))
	field := NewDecimalField(nil, "", title, get, set, miniumum, maximum, false, false)
	content.AddChild(WrapWithSpan(1, field, NewFieldTrailingLabel(i18n.Text("seconds"), false)))
	return field
}

func (d *webSettingsDockable) createUsersBlock(content *unison.Panel) {
	header := unison.NewPanel()
	header.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Fill,
		HGrab:  true,
	})
	header.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing / 2,
	})
	content.AddChild(header)
	d.userAddButton = unison.NewSVGButton(svg.CircledAdd)
	d.userAddButton.Tooltip = newWrappedTooltip(i18n.Text("Add User"))
	d.userAddButton.ClickCallback = d.addUser
	header.AddChild(d.userAddButton)
	title := unison.NewLabel()
	title.Text = i18n.Text("Users")
	header.AddChild(title)
	d.userList = unison.NewList[*websettings.User]()
	d.userList.BackgroundInk = unison.ContentColor
	d.userList.SetBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.NewUniformInsets(1), false))
	d.userList.SetLayoutData(&unison.FlexLayoutData{
		MinSize: unison.NewSize(300, 64),
		HSpan:   2,
		HAlign:  align.Fill,
		VAlign:  align.Fill,
		HGrab:   true,
		VGrab:   true,
	})
	d.userList.Append(gurps.GlobalSettings().WebServer.Users()...)
	d.userList.DoubleClickCallback = d.editUser
	d.userList.KeyDownCallback = d.handleUserListKey
	content.AddChild(d.userList)
}

func (d *webSettingsDockable) handleUserListKey(keyCode unison.KeyCode, mod unison.Modifiers, repeat bool) bool {
	if gurps.GlobalSettings().WebServer.Enabled {
		return false
	}
	switch keyCode {
	case unison.KeyDelete, unison.KeyBackspace:
		d.deleteUser()
		return true
	case unison.KeyReturn, unison.KeyNumPadEnter:
		d.editUser()
		return true
	}
	return d.userList.DefaultKeyDown(keyCode, mod, repeat)
}

func (d *webSettingsDockable) deleteUser() {
	settings := gurps.GlobalSettings().WebServer
	if settings.Enabled || d.userList.Selection.FirstSet() == -1 {
		return
	}
	for {
		i := d.userList.Selection.LastSet()
		if i == -1 {
			d.userList.Pack()
			d.userList.MarkForLayoutRecursivelyUpward()
			d.ValidateLayout()
			break
		}
		settings.RemoveUser(d.userList.DataAtIndex(i).Name)
		d.userList.Remove(i)
	}
}

func (d *webSettingsDockable) editUser() {
	settings := gurps.GlobalSettings().WebServer
	if settings.Enabled || d.userList.Selection.Count() != 1 {
		return
	}
	u := d.userList.DataAtIndex(d.userList.Selection.FirstSet())
	name := u.Name
	var err error
	panel := d.createUserPanel(u, false)
	d.userDialog, err = unison.NewDialog(nil, nil, panel, []*unison.DialogButtonInfo{
		unison.NewCancelButtonInfo(),
		unison.NewOKButtonInfoWithTitle(i18n.Text("Update")),
	})
	if err != nil {
		errs.Log(err)
		return
	}
	if flex, ok := panel.LayoutData().(*unison.FlexLayoutData); ok {
		flex.MinSize = unison.NewSize(500, 0)
	}
	d.userNameField.Validate()
	defer func() { d.userDialog = nil }()
	if d.userDialog.RunModal() == unison.ModalResponseOK {
		if name != u.Name {
			if !settings.RenameUser(name, u.Name) {
				unison.ErrorDialogWithMessage(i18n.Text("Unable to rename user"),
					i18n.Text("A user with that name already exists."))
				return
			}
		}
		if u.HashedPassword != "" {
			if !settings.SetUserPassword(u.Name, u.HashedPassword) {
				unison.ErrorDialogWithMessage(i18n.Text("Unable to set user's password"),
					i18n.Text("A user with that name cannot be found."))
				return
			}
		}
		settings.SetAccessList(u.Name, u.AccessList)
		all := settings.Users()
		i := slices.IndexFunc(all, func(one *websettings.User) bool { return one.Key() == u.Key() })
		d.userList.Select(false, i)
		d.userList.Pack()
		d.userList.MarkForLayoutRecursivelyUpward()
		d.ValidateLayout()
		d.userList.ScrollRectIntoView(d.userList.RowRect(i).Inset(unison.NewUniformInsets(-2)))
		d.userList.MarkForRedraw()
	}
}

func (d *webSettingsDockable) addUser() {
	settings := gurps.GlobalSettings()
	webSettings := settings.WebServer
	if webSettings.Enabled {
		return
	}
	libraries := settings.LibrarySet
	u := &websettings.User{
		AccessList: map[string]websettings.Access{
			"master": {Dir: libraries.Master().Path(), ReadOnly: true},
			"user":   {Dir: libraries.User().Path(), ReadOnly: false},
		},
	}
	var err error
	panel := d.createUserPanel(u, true)
	d.userDialog, err = unison.NewDialog(nil, nil, panel, []*unison.DialogButtonInfo{
		unison.NewCancelButtonInfo(),
		unison.NewOKButtonInfoWithTitle(i18n.Text("Create")),
	})
	if err != nil {
		errs.Log(err)
		return
	}
	if flex, ok := panel.LayoutData().(*unison.FlexLayoutData); ok {
		flex.MinSize = unison.NewSize(500, 0)
	}
	d.userNameField.Validate()
	d.passwordField.Validate()
	if d.userDialog.RunModal() == unison.ModalResponseOK {
		if webSettings.CreateUser(u.Name, u.HashedPassword) {
			webSettings.SetAccessList(u.Name, u.AccessList)
			all := webSettings.Users()
			i := slices.IndexFunc(all, func(one *websettings.User) bool { return one.Key() == u.Key() })
			d.userList.Insert(i, all[i])
			d.userList.Select(false, i)
			d.userList.Pack()
			d.userList.MarkForLayoutRecursivelyUpward()
			d.ValidateLayout()
			d.userList.ScrollRectIntoView(d.userList.RowRect(i).Inset(unison.NewUniformInsets(-2)))
			d.userList.MarkForRedraw()
		} else {
			unison.ErrorDialogWithMessage(i18n.Text("Unable to create user"),
				i18n.Text("A user with that name already exists."))
		}
	}
	d.userDialog = nil
}

func (d *webSettingsDockable) createUserPanel(u *websettings.User, newUser bool) *unison.ScrollPanel {
	d.accessListUser = u
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	d.createUserNameField(panel, u, newUser)
	d.createPasswordField(panel, u, newUser)
	d.createAccessBlock(panel, u)
	scroller := unison.NewScrollPanel()
	scroller.SetContent(panel, behavior.Follow, behavior.Fill)
	return scroller
}

func (d *webSettingsDockable) createUserNameField(content *unison.Panel, u *websettings.User, newUser bool) {
	title := i18n.Text("Name")
	content.AddChild(NewFieldLeadingLabel(title, false))
	d.userNameField = NewStringField(nil, "", title, func() string { return u.Name }, func(s string) { u.Name = s })
	d.originalName = u.Key()
	d.userNameField.ValidateCallback = func() bool {
		valid := d.isNameFieldValid()
		d.adjustUserDialogOKButton(valid, !newUser || d.isPasswordFieldValid())
		return valid
	}
	content.AddChild(d.userNameField)
}

func (d *webSettingsDockable) createPasswordField(content *unison.Panel, u *websettings.User, newUser bool) {
	var title string
	if newUser {
		title = i18n.Text("Password")
	} else {
		title = i18n.Text("New Password")
	}
	content.AddChild(NewFieldLeadingLabel(title, false))
	u.HashedPassword = ""
	d.passwordField = NewStringField(nil, "", title, func() string { return u.HashedPassword },
		func(s string) { u.HashedPassword = s })
	d.passwordField.ObscurementRune = '•'
	if newUser {
		d.passwordField.ValidateCallback = func() bool {
			valid := d.isPasswordFieldValid()
			d.adjustUserDialogOKButton(d.isNameFieldValid(), valid)
			return valid
		}
	} else {
		d.passwordField.Watermark = i18n.Text("Leave blank to keep the current password")
	}
	content.AddChild(d.passwordField)
}

func (d *webSettingsDockable) createAccessBlock(content *unison.Panel, u *websettings.User) {
	header := unison.NewPanel()
	header.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Fill,
		HGrab:  true,
	})
	header.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing / 2,
	})
	content.AddChild(header)
	addButton := unison.NewSVGButton(svg.CircledAdd)
	addButton.Tooltip = newWrappedTooltip(i18n.Text("Add Access"))
	addButton.ClickCallback = d.addAccess
	header.AddChild(addButton)
	title := unison.NewLabel()
	title.Text = i18n.Text("Access")
	header.AddChild(title)
	d.accessList = unison.NewList[*websettings.AccessWithKey]()
	d.accessList.BackgroundInk = unison.ContentColor
	d.accessList.SetBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.NewUniformInsets(1), false))
	d.accessList.SetLayoutData(&unison.FlexLayoutData{
		MinSize: unison.NewSize(300, 64),
		HSpan:   2,
		HAlign:  align.Fill,
		VAlign:  align.Fill,
		HGrab:   true,
		VGrab:   true,
	})
	d.accessList.Append(u.AccessListWithKeys()...)
	d.accessList.DoubleClickCallback = d.editAccess
	d.accessList.KeyDownCallback = d.handleAccessListKey
	content.AddChild(d.accessList)
}

func (d *webSettingsDockable) handleAccessListKey(keyCode unison.KeyCode, mod unison.Modifiers, repeat bool) bool {
	if gurps.GlobalSettings().WebServer.Enabled {
		return false
	}
	switch keyCode {
	case unison.KeyDelete, unison.KeyBackspace:
		d.deleteAccess()
		return true
	case unison.KeyReturn, unison.KeyNumPadEnter:
		d.editAccess()
		return true
	}
	return d.accessList.DefaultKeyDown(keyCode, mod, repeat)
}

func (d *webSettingsDockable) deleteAccess() {
	settings := gurps.GlobalSettings().WebServer
	if settings.Enabled || d.accessList.Selection.FirstSet() == -1 {
		return
	}
	for {
		i := d.accessList.Selection.LastSet()
		if i == -1 {
			d.accessList.Pack()
			d.accessList.MarkForLayoutRecursivelyUpward()
			d.ValidateLayout()
			break
		}
		delete(d.accessListUser.AccessList, d.accessList.DataAtIndex(i).Key)
		d.accessList.Remove(i)
	}
}

func (d *webSettingsDockable) editAccess() {
	if d.accessList.Selection.Count() != 1 {
		return
	}
	access := d.accessList.DataAtIndex(d.accessList.Selection.FirstSet())
	var err error
	panel := d.createAccessPanel(access)
	d.accessDialog, err = unison.NewDialog(nil, nil, panel, []*unison.DialogButtonInfo{
		unison.NewCancelButtonInfo(),
		unison.NewOKButtonInfoWithTitle(i18n.Text("Update")),
	})
	if err != nil {
		errs.Log(err)
		return
	}
	if flex, ok := panel.LayoutData().(*unison.FlexLayoutData); ok {
		flex.MinSize = unison.NewSize(500, 0)
	}
	d.accessKeyField.Validate()
	d.accessDirField.Validate()
	defer func() { d.accessDialog = nil }()
	if d.accessDialog.RunModal() == unison.ModalResponseOK {
		d.accessListUser.AccessList[access.Key] = access.Access
		all := d.accessListUser.AccessListWithKeys()
		i := slices.IndexFunc(all, func(one *websettings.AccessWithKey) bool { return one.Key == access.Key })
		d.accessList.Select(false, i)
		d.accessList.Pack()
		d.accessList.MarkForLayoutRecursivelyUpward()
		d.ValidateLayout()
		d.accessList.ScrollRectIntoView(d.accessList.RowRect(i).Inset(unison.NewUniformInsets(-2)))
		d.accessList.MarkForRedraw()
	}
}

func (d *webSettingsDockable) addAccess() {
	access := &websettings.AccessWithKey{}
	var err error
	panel := d.createAccessPanel(access)
	d.accessDialog, err = unison.NewDialog(nil, nil, panel, []*unison.DialogButtonInfo{
		unison.NewCancelButtonInfo(),
		unison.NewOKButtonInfoWithTitle(i18n.Text("Create")),
	})
	if err != nil {
		errs.Log(err)
		return
	}
	if flex, ok := panel.LayoutData().(*unison.FlexLayoutData); ok {
		flex.MinSize = unison.NewSize(500, 0)
	}
	d.accessKeyField.Validate()
	d.accessDirField.Validate()
	if d.accessDialog.RunModal() == unison.ModalResponseOK {
		if d.accessListUser.AccessList == nil {
			d.accessListUser.AccessList = make(map[string]websettings.Access)
		}
		d.accessListUser.AccessList[access.Key] = access.Access
		all := d.accessListUser.AccessListWithKeys()
		i := slices.IndexFunc(all, func(one *websettings.AccessWithKey) bool { return one.Key == access.Key })
		d.accessList.Insert(i, all[i])
		d.accessList.Select(false, i)
		d.accessList.Pack()
		d.accessList.MarkForLayoutRecursivelyUpward()
		d.ValidateLayout()
		d.accessList.ScrollRectIntoView(d.accessList.RowRect(i).Inset(unison.NewUniformInsets(-2)))
		d.accessList.MarkForRedraw()
	}
	d.accessDialog = nil
}

func (d *webSettingsDockable) createAccessPanel(access *websettings.AccessWithKey) *unison.Panel {
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{MinSize: unison.NewSize(300, 0)})
	d.accessOriginal = *access
	d.createAccessKeyField(panel, access)
	d.createAccessDirField(panel, access)
	d.createAccessReadOnlyCheckbox(panel, access)
	return panel
}

func (d *webSettingsDockable) createAccessKeyField(content *unison.Panel, access *websettings.AccessWithKey) {
	title := i18n.Text("Key")
	content.AddChild(NewFieldLeadingLabel(title, false))
	d.accessKeyField = NewStringField(nil, "", title, func() string { return access.Key },
		func(s string) { access.Key = s })
	d.accessKeyField.ValidateCallback = func() bool {
		valid := d.isAccessKeyFieldValid()
		d.adjustAccessDialogOKButton(valid, d.isAccessDirFieldValid())
		return valid
	}
	content.AddChild(d.accessKeyField)
}

func (d *webSettingsDockable) createAccessDirField(content *unison.Panel, access *websettings.AccessWithKey) {
	d.accessDirField, _ = createFilePathField(content, i18n.Text("Directory"), func() string { return access.Dir },
		func(s string) { access.Dir = s }, true)
	d.accessDirField.ValidateCallback = func() bool {
		valid := d.isAccessDirFieldValid()
		d.adjustAccessDialogOKButton(d.isAccessKeyFieldValid(), valid)
		return valid
	}
}

func (d *webSettingsDockable) createAccessReadOnlyCheckbox(content *unison.Panel, access *websettings.AccessWithKey) {
	checkbox := NewCheckBox(nil, "", i18n.Text("Read Only"),
		func() check.Enum { return check.FromBool(access.ReadOnly) },
		func(state check.Enum) { access.ReadOnly = state == check.On })
	checkbox.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Middle,
	})
	content.AddChild(checkbox)
}

func (d *webSettingsDockable) isNameFieldValid() bool {
	s := websettings.UserNameToKey(d.userNameField.Text())
	if s == "" {
		return false
	}
	if d.originalName == s {
		return true
	}
	for _, one := range gurps.GlobalSettings().WebServer.Users() {
		if s == one.Key() {
			return false
		}
	}
	return true
}

func (d *webSettingsDockable) isPasswordFieldValid() bool {
	return d.passwordField.Text() != ""
}

func (d *webSettingsDockable) isAccessKeyFieldValid() bool {
	s := strings.TrimSpace(d.accessKeyField.Text())
	if s == "" {
		return false
	}
	if d.accessOriginal.Key == s {
		return true
	}
	for _, one := range d.accessListUser.AccessListWithKeys() {
		if s == one.Key {
			return false
		}
	}
	return true
}

func (d *webSettingsDockable) isAccessDirFieldValid() bool {
	s := strings.TrimSpace(d.accessDirField.Text())
	if s == "" {
		return false
	}
	s = filepath.Clean(s)
	if !filepath.IsAbs(s) {
		return false
	}
	if d.accessOriginal.Dir == s {
		return true
	}
	for _, one := range d.accessListUser.AccessListWithKeys() {
		if s == one.Dir {
			return false
		}
	}
	return true
}

func (d *webSettingsDockable) adjustUserDialogOKButton(nameFieldValid, passwordFieldValid bool) {
	d.userDialog.Button(unison.ModalResponseOK).SetEnabled(nameFieldValid && passwordFieldValid)
}

func (d *webSettingsDockable) adjustAccessDialogOKButton(accessKeyValid, accessDirValid bool) {
	d.accessDialog.Button(unison.ModalResponseOK).SetEnabled(accessKeyValid && accessDirValid)
}

func (d *webSettingsDockable) reset() {
	if state.Current() != state.Stopped {
		if unison.YesNoDialog(i18n.Text("Stop the server?"), i18n.Text("The web server must be stopped before the settings can be reset.")) == unison.ModalResponseOK {
			StopServer()
			d.syncEnablementToServer(d.reset)
		}
		return
	}
	d.sync(nil)
}

func (d *webSettingsDockable) sync(other *websettings.Settings) {
	settings := gurps.GlobalSettings().WebServer
	settings.CopyFrom(other)
	on := settings.Enabled
	settings.Enabled = false
	SetCheckBoxState(d.enabledCheckbox, settings.Enabled)
	SetFieldValue(d.addressField.Field, settings.Address)
	SetFieldValue(d.certFileField.Field, settings.CertFile)
	SetFieldValue(d.keyFileField.Field, settings.KeyFile)
	SetFieldValue(d.shutdownGracePeriodField.Field, settings.ShutdownGracePeriod.String())
	SetFieldValue(d.readTimeoutField.Field, settings.ReadTimeout.String())
	SetFieldValue(d.writeTimeoutField.Field, settings.WriteTimeout.String())
	SetFieldValue(d.idleTimeoutField.Field, settings.IdleTimeout.String())
	d.userList.Clear()
	d.userList.Append(settings.Users()...)
	d.updateErrorMsg(nil)
	d.MarkForRedraw()
	d.applyServerEnabled(on)
}

func (d *webSettingsDockable) load(fileSystem fs.FS, filePath string) error {
	if state.Current() != state.Stopped {
		if unison.YesNoDialog(i18n.Text("Stop the server?"), i18n.Text("The web server must be stopped before new settings can be loaded.")) == unison.ModalResponseOK {
			StopServer()
			d.syncEnablementToServer(func() {
				if err := d.load(fileSystem, filePath); err != nil {
					unison.ErrorDialogWithError(i18n.Text("Unable to load ")+d.TabTitle, err)
				}
			})
		}
		return nil
	}
	settings, err := websettings.NewSettingsFromFile(fileSystem, filePath)
	if err != nil {
		return err
	}
	d.sync(settings)
	return nil
}

func (d *webSettingsDockable) save(filePath string) error {
	return gurps.GlobalSettings().WebServer.Save(filePath)
}
