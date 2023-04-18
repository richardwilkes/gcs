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

package ux

import (
	"fmt"
	"io/fs"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/unison"
)

type menuKeySettingsDockable struct {
	SettingsDockable
	content *unison.Panel
}

// ShowMenuKeySettings shows the Menu Key settings.
func ShowMenuKeySettings() {
	if Activate(func(d unison.Dockable) bool {
		_, ok := d.AsPanel().Self.(*menuKeySettingsDockable)
		return ok
	}) {
		return
	}
	d := &menuKeySettingsDockable{}
	d.Self = d
	d.TabTitle = i18n.Text("Menu Keys")
	d.TabIcon = svg.Settings
	d.Extensions = []string{gurps.KeySettingsExt}
	d.Loader = d.load
	d.Saver = d.save
	d.Resetter = d.reset
	d.Setup(nil, nil, d.initContent)
}

func (d *menuKeySettingsDockable) initContent(content *unison.Panel) {
	d.content = content
	d.content.SetLayout(&unison.FlexLayout{
		Columns:  3,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
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
		d.content.AddChild(NewFieldLeadingLabel(b.Action.Title))
		d.createBindingButton(b)
		d.createResetField(b)
	}
}

func (d *menuKeySettingsDockable) createBindingButton(binding *gurps.Binding) {
	b := unison.NewButton()
	b.Font = unison.KeyboardFont
	b.Text = binding.Action.KeyBinding.String()
	b.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
	})
	b.ClickCallback = func() {
		localBinding := binding.KeyBinding
		capturePanel := unison.NewLabel()
		capturePanel.Font = unison.KeyboardFont
		capturePanel.Text = binding.KeyBinding.String()
		capturePanel.HAlign = unison.MiddleAlignment
		capturePanel.SetBorder(unison.DefaultFieldTheme.FocusedBorder)
		capturePanel.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
			gc.DrawRect(rect, unison.DefaultFieldTheme.BackgroundInk.Paint(gc, rect, unison.Fill))
			capturePanel.DefaultDraw(gc, rect)
		}
		capturePanel.KeyDownCallback = func(keyCode unison.KeyCode, mod unison.Modifiers, repeat bool) bool {
			localBinding.KeyCode = keyCode
			localBinding.Modifiers = mod
			capturePanel.Text = localBinding.String()
			capturePanel.MarkForRedraw()
			return true
		}
		capturePanel.SetFocusable(true)
		wrapper := unison.NewPanel()
		wrapper.SetLayout(&unison.FlexLayout{
			Columns: 1,
			HAlign:  unison.MiddleAlignment,
			VAlign:  unison.MiddleAlignment,
		})
		capturePanel.SetLayoutData(&unison.FlexLayoutData{
			MinSize: unison.Size{Width: 100, Height: 50},
			HAlign:  unison.FillAlignment,
			VAlign:  unison.FillAlignment,
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
			jot.Error(err)
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
				b.Text = localBinding.String()
				b.MarkForRedraw()
			default:
			}
		}
	}
	d.content.AddChild(b)
}

func (d *menuKeySettingsDockable) createResetField(binding *gurps.Binding) {
	b := unison.NewSVGButton(svg.Reset)
	b.Tooltip = unison.NewTooltipWithText("Reset this key binding")
	b.ClickCallback = func() {
		if unison.QuestionDialog(fmt.Sprintf(i18n.Text("Are you sure you want to reset '%s'?"), binding.Action.Title), "") == unison.ModalResponseOK {
			g := gurps.GlobalSettings()
			g.KeyBindings.ResetOne(binding.ID)
			g.KeyBindings.MakeCurrent()
			binding.KeyBinding = g.KeyBindings.Current(binding.ID)
			parent := b.Parent()
			if other, ok := parent.Children()[parent.IndexOfChild(b)-1].Self.(*unison.Button); ok {
				other.Text = binding.KeyBinding.String()
			}
		}
	}
	b.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.MiddleAlignment,
		VAlign: unison.MiddleAlignment,
	})
	d.content.AddChild(b)
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
