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
	"fmt"
	"io/fs"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

type menuKeySettingsDockable struct {
	SettingsDockable
	content *unison.Panel
}

// ShowMenuKeySettings shows the Menu Key settings.
func ShowMenuKeySettings() {
	if activateDockable[*menuKeySettingsDockable](nil) {
		return
	}
	d := &menuKeySettingsDockable{}
	d.initSettings(d, &settingsSpec{
		title:       i18n.Text("Menu Keys"),
		ext:         gurps.KeySettingsExt,
		loader:      d.load,
		saver:       d.save,
		resetter:    d.reset,
		initContent: d.initContent,
	})
}

func (d *menuKeySettingsDockable) initContent(content *unison.Panel) {
	d.content = initSettingsContent(content, 3)
	d.fill()
}

func (d *menuKeySettingsDockable) reset() {
	g := gurps.GlobalSettings()
	g.KeyBindings.Reset()
	g.KeyBindings.MakeCurrent()
	d.sync()
}

func (d *menuKeySettingsDockable) sync() {
	d.content.RemoveAllChildren()
	d.fill()
	d.MarkForRedraw()
}

func (d *menuKeySettingsDockable) fill() {
	for _, b := range gurps.CurrentBindings() {
		d.createBindingButton(b)
		d.content.AddChild(NewFieldTrailingLabel(b.Action.Title, false))
		d.createResetField(b)
	}
}

func (d *menuKeySettingsDockable) createBindingButton(binding *gurps.Binding) {
	b := unison.NewButton()
	b.Font = unison.KeyboardFont
	setBindingButtonKey(b, binding)
	b.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	b.ClickCallback = func() {
		localBinding := binding.KeyBinding
		capturePanel := unison.NewLabel()
		capturePanel.Font = unison.KeyboardFont
		capturePanel.SetTitle(binding.KeyBinding.String())
		// The label shows only the key combination, which is empty when there is none, so it has to say what it is for
		// itself. The combination captured so far goes in its description and is announced as each one is pressed,
		// since nothing else would tell someone who cannot see the label that their key press registered.
		capturePanel.Accessibility.Name = fmt.Sprintf(i18n.Text("Press the new key combination for %s"),
			binding.Action.Title)
		capturePanel.Accessibility.Description = describeKeyBinding(localBinding)
		capturePanel.HAlign = align.Middle
		unison.InstallDefaultFieldBorder(capturePanel, capturePanel)
		capturePanel.DrawCallback = func(gc *unison.Canvas, rect geom.Rect) {
			gc.DrawRect(rect, unison.DefaultFieldTheme.BackgroundInk.Paint(gc, rect, paintstyle.Fill))
			capturePanel.DefaultDraw(gc, rect)
		}
		capturePanel.KeyDownCallback = func(keyCode unison.KeyCode, mods mod.Modifiers, _ bool) bool {
			localBinding.KeyCode = keyCode
			localBinding.Modifiers = mods
			capturePanel.SetTitle(localBinding.String())
			capturePanel.Accessibility.Description = describeKeyBinding(localBinding)
			capturePanel.MarkForRedraw()
			unison.AnnounceForAccessibility(capturePanel.Accessibility.Description)
			return true
		}
		capturePanel.SetFocusable(true)
		wrapper := unison.NewPanel()
		wrapper.SetLayout(&unison.FlexLayout{
			Columns: 1,
			HAlign:  align.Middle,
			VAlign:  align.Middle,
		})
		capturePanel.SetLayoutData(&unison.FlexLayoutData{
			MinSize: geom.Size{Width: 100, Height: 50},
			HAlign:  align.Fill,
			VAlign:  align.Fill,
			HGrab:   true,
			VGrab:   true,
		})
		wrapper.AddChild(capturePanel)
		if dialog, err := unison.NewDialog(nil, nil, wrapper,
			[]*unison.DialogButtonInfo{
				{
					Title:        i18n.Text("Clear"),
					ResponseCode: unison.ModalResponseUserBase,
				},
				unison.NewCancelButtonInfo(),
				unison.NewOKButtonInfoWithTitle(i18n.Text("Set")),
			}); err != nil {
			errs.Log(err)
		} else {
			unison.DisableMenus = true
			defer func() { unison.DisableMenus = false }()
			switch dialog.RunModal() {
			case unison.ModalResponseUserBase:
				localBinding = unison.KeyBinding{}
				fallthrough
			case unison.ModalResponseOK:
				binding.KeyBinding = localBinding
				g := gurps.GlobalSettings()
				g.KeyBindings.Set(binding.ID, localBinding)
				g.KeyBindings.MakeCurrent()
				setBindingButtonKey(b, binding)
			default:
			}
		}
	}
	d.content.AddChild(b)
}

// setBindingButtonKey shows the binding's key combination on the button that displays it. The button is named after the
// action as well, since the combination alone would not say which action it belongs to, and an action with no key would
// leave the button with no name at all.
func setBindingButtonKey(b *unison.Button, binding *gurps.Binding) {
	b.SetTitle(binding.KeyBinding.String())
	b.Accessibility.Name = fmt.Sprintf(i18n.Text("%s: %s"), binding.Action.Title, describeKeyBinding(binding.KeyBinding))
	b.MarkForRedraw()
}

// describeKeyBinding returns the key combination as it should be spoken, which is how it is drawn unless there is none.
func describeKeyBinding(keyBinding unison.KeyBinding) string {
	if s := keyBinding.String(); s != "" {
		return s
	}
	return i18n.Text("no key")
}

func (d *menuKeySettingsDockable) createResetField(binding *gurps.Binding) {
	b := unison.NewSVGButton(svg.Reset)
	b.Tooltip = newWrappedTooltip(i18n.Text("Reset this key binding"))
	b.ClickCallback = func() {
		if unison.QuestionDialog(fmt.Sprintf(i18n.Text("Are you sure you want to reset '%s'?"), binding.Action.Title), "") == unison.ModalResponseOK {
			d.resetBinding(binding, b)
		}
	}
	b.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Middle,
		VAlign: align.Middle,
	})
	d.content.AddChild(b)
}

// resetBinding restores the given binding to its factory default and updates the button that displays it.
func (d *menuKeySettingsDockable) resetBinding(binding *gurps.Binding, resetButton *unison.Button) {
	g := gurps.GlobalSettings()
	g.KeyBindings.ResetOne(binding.ID)
	g.KeyBindings.MakeCurrent()
	binding.KeyBinding = g.KeyBindings.Current(binding.ID)
	if other := bindingButtonForResetButton(resetButton); other != nil {
		setBindingButtonKey(other, binding)
	}
}

// bindingButtonForResetButton returns the button that displays the key binding for the row the given reset button
// belongs to, or nil if it can't be located. fill() adds three children per binding, in the order [key button, title
// label, reset button], so the key button sits two positions before its reset button.
func bindingButtonForResetButton(resetButton *unison.Button) *unison.Button {
	parent := resetButton.Parent()
	if parent == nil {
		return nil
	}
	i := parent.IndexOfChild(resetButton) - 2
	if i < 0 {
		return nil
	}
	other, ok := parent.Children()[i].Self.(*unison.Button)
	if !ok {
		return nil
	}
	return other
}

func (d *menuKeySettingsDockable) load(fileSystem fs.FS, filePath string) error {
	b, err := gurps.NewKeyBindingsFromFS(fileSystem, filePath)
	if err != nil {
		return err
	}
	g := gurps.GlobalSettings()
	g.KeyBindings = *b
	g.KeyBindings.MakeCurrent()
	d.sync()
	return nil
}

func (d *menuKeySettingsDockable) save(filePath string) error {
	return gurps.GlobalSettings().KeyBindings.Save(filePath)
}
