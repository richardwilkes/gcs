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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/server"
	"github.com/richardwilkes/gcs/v5/server/websettings"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	xfs "github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
)

type webSettingsDockable struct {
	SettingsDockable
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
	d.syncEnablementToServer(nil)
}

func (d *webSettingsDockable) syncEnablementToServer(callback func()) {
	d.setWebServerControlEnablement(false)
	go d.finishSync(callback)
}

func (d *webSettingsDockable) finishSync(callback func()) {
	server.WaitUntilState(server.Running, server.Stopped)
	unison.InvokeTask(func() {
		settings := gurps.GlobalSettings().WebServer
		switch server.CurrentState() {
		case server.Stopped:
			settings.Enabled = false
			d.enabledCheckbox.State = check.Off
			d.setWebServerControlEnablement(true)
		case server.Running:
			settings.Enabled = true
			d.enabledCheckbox.State = check.On
			d.enabledCheckbox.SetEnabled(true)
		default:
			go d.finishSync(callback)
		}
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
	if on && !settings.Valid() {
		on = false
	}
	if settings.Enabled != on {
		settings.Enabled = on
		if on {
			server.Start()
		} else {
			server.Stop()
		}
		d.syncEnablementToServer(nil)
	}
}

func (d *webSettingsDockable) createAddressField(content *unison.Panel) {
	title := i18n.Text("Server Address")
	content.AddChild(NewFieldLeadingLabel(title, false))
	d.addressField = NewStringField(nil, "", title,
		func() string { return gurps.GlobalSettings().WebServer.Address },
		func(s string) {
			if isAddressValid(s) {
				gurps.GlobalSettings().WebServer.Address = s
			}
		})
	d.addressField.ValidateCallback = func() bool {
		return isAddressValid(d.addressField.Text())
	}
	content.AddChild(d.addressField)
}

func isAddressValid(address string) bool {
	if address == "" {
		return true
	}
	if _, _, err := net.SplitHostPort(address); err != nil {
		return false
	}
	return true
}

func (d *webSettingsDockable) createCertFileField(content *unison.Panel) {
	d.certFileField, d.certFileButton = createFilePathField(content, i18n.Text("Certificate File"),
		func() string { return gurps.GlobalSettings().WebServer.CertFile },
		func(s string) { gurps.GlobalSettings().WebServer.CertFile = s })
}

func (d *webSettingsDockable) createKeyFileField(content *unison.Panel) {
	d.keyFileField, d.keyFileButton = createFilePathField(content, i18n.Text("Key File"),
		func() string { return gurps.GlobalSettings().WebServer.KeyFile },
		func(s string) { gurps.GlobalSettings().WebServer.KeyFile = s })
}

func createFilePathField(content *unison.Panel, title string, get func() string, set func(string)) (*StringField, *unison.Button) {
	content.AddChild(NewFieldLeadingLabel(title, false))
	fileField := NewStringField(nil, "", title, get, set)
	fileField.ValidateCallback = func() bool {
		p := get()
		if p == "" {
			return true
		}
		return filepath.IsAbs(p) && xfs.FileIsReadable(p)
	}
	locateButton := unison.NewSVGButton(svg.GenericFile)
	locateButton.ClickCallback = func() { chooseFilePath(fileField) }
	content.AddChild(WrapWithSpan(1, fileField, locateButton))
	return fileField, locateButton
}

func chooseFilePath(field *StringField) {
	dlg := unison.NewOpenDialog()
	dlg.SetAllowsMultipleSelection(false)
	dlg.SetResolvesAliases(true)
	dlg.SetCanChooseDirectories(false)
	dlg.SetCanChooseFiles(true)
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

func (d *webSettingsDockable) reset() {
	if server.CurrentState() != server.Stopped {
		if unison.YesNoDialog(i18n.Text("Stop the server?"), i18n.Text("The web server must be stopped before the settings can be reset.")) == unison.ModalResponseOK {
			server.Stop()
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
	d.MarkForRedraw()
	d.applyServerEnabled(on)
}

func (d *webSettingsDockable) load(fileSystem fs.FS, filePath string) error {
	if server.CurrentState() != server.Stopped {
		if unison.YesNoDialog(i18n.Text("Stop the server?"), i18n.Text("The web server must be stopped before new settings can be loaded.")) == unison.ModalResponseOK {
			server.Stop()
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
