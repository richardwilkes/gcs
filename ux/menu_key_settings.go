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
	"io/fs"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/side"
)

type menuKeySettingsDockable struct {
	SettingsDockable
	content     *unison.Panel
	filterField *unison.Field
	rows        map[string]*menuKeyRow // By binding ID
}

// menuKeyRow is a binding's row in the view: its label, the button that shows its key and opens the dialog that changes
// it, and the one that resets it, which is enabled only while the binding differs from its factory default.
type menuKeyRow struct {
	binding     *gurps.Binding
	label       *unison.Label
	keyButton   *unison.Button
	resetButton *unison.Button
	indent      float32 // How far the row's grid is inset from the section's
}

// ShowMenuKeySettings shows the Menu Key settings.
func ShowMenuKeySettings() {
	if ActivateDockableOfType[*menuKeySettingsDockable](nil) {
		return
	}
	d := &menuKeySettingsDockable{}
	// The sections wrap to the view's width, but a view too narrow for even one of them scrolls sideways.
	d.ContentWidthBehavior = behavior.HintedFill
	d.initSettings(d, &settingsSpec{
		title:             i18n.Text("Menu Keys"),
		ext:               gurps.KeySettingsExt,
		loader:            d.load,
		saver:             d.save,
		resetter:          d.reset,
		addToStartToolbar: d.addToStartToolbar,
		initContent:       d.initContent,
	})
}

func (d *menuKeySettingsDockable) addToStartToolbar(toolbar *unison.Panel) {
	d.filterField = NewToolbarSearchField(i18n.Text("Filter"), func(_, _ *unison.FieldState) { d.sync() })
	toolbar.AddChild(d.filterField)
}

// initContent lays the sections out in a flow, so that they sit side by side when the view is wide enough for more
// than one and stack when it is not.
func (d *menuKeySettingsDockable) initContent(content *unison.Panel) {
	content.SetLayout(&unison.FlowLayout{
		HSpacing: unison.StdHSpacing * 4,
		VSpacing: unison.DefaultLabelTheme.Font.LineHeight(),
	})
	d.content = content
	d.fill()
}

func (d *menuKeySettingsDockable) reset() {
	g := gurps.GlobalSettings()
	g.KeyBindings.Reset()
	g.KeyBindings.MakeCurrent()
	d.sync()
}

// sync rebuilds the rows from the current bindings.
func (d *menuKeySettingsDockable) sync() {
	d.content.RemoveAllChildren()
	d.fill()
	d.MarkForRedraw()
}

// fill adds a section for each menu that holds key-bindable actions, with only the rows the filter field's text lets
// through, and no section at all for a menu it lets none of.
func (d *menuKeySettingsDockable) fill() {
	d.rows = make(map[string]*menuKeyRow)
	var text string
	if d.filterField != nil {
		text = strings.ToLower(strings.TrimSpace(d.filterField.Text()))
	}
	for _, section := range menuKeySections() {
		if section = section.filtered(text); section != nil {
			var rows []*menuKeyRow
			d.addRows(newSettingsSection(d.content, section.title, 3), section, &rows, 0)
			alignRows(rows)
		}
	}
}

// addRows adds a row to the grid, which is inset from the section's by indent, for each of the section's bindings,
// then a block for each of its submenus, appending the rows to rows.
func (d *menuKeySettingsDockable) addRows(grid *unison.Panel, section *menuKeySection, rows *[]*menuKeyRow, indent float32) {
	for _, binding := range section.bindings {
		row := d.addRow(grid, binding)
		row.indent = indent
		*rows = append(*rows, row)
	}
	for _, sub := range section.submenus {
		d.addSubmenu(grid, sub, rows, indent)
	}
}

// addSubmenu adds the submenu's rows to the grid in a block of their own under a heading that shows its title as the
// menu does, with the arrow that marks a submenu, in a lighter ink. The block is inset, with a rule down its left edge,
// so that the heading is seen to own its rows.
func (d *menuKeySettingsDockable) addSubmenu(grid *unison.Panel, sub *menuKeySection, rows *[]*menuKeyRow, indent float32) {
	heading := unison.NewLabel()
	heading.SetTitle(sub.title)
	heading.OnBackgroundInk = &unison.ColorFilteredInk{OriginalInk: unison.ThemeOnSurface, ColorFilter: unison.Alpha50Filter()}
	baseline := heading.Font.Baseline()
	heading.Drawable = &unison.DrawableSVG{
		SVG:  unison.ChevronRightSVG,
		Size: geom.NewSize(baseline, baseline),
	}
	heading.Side = side.Right
	heading.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: unison.StdVSpacing * 2}))
	heading.SetLayoutData(&unison.FlexLayoutData{HSpan: 3})
	grid.AddChild(heading)
	block := unison.NewPanel()
	block.SetBorder(unison.NewCompoundBorder(
		unison.NewEmptyBorder(geom.Insets{Left: unison.StdHSpacing}),
		unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.Insets{Left: 1}, false),
		unison.NewEmptyBorder(geom.Insets{Left: unison.StdHSpacing - 1}),
	))
	block.SetLayout(&unison.FlexLayout{
		Columns:  3,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	block.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  3,
		HAlign: align.Fill,
	})
	block.Accessibility.Name = sub.title // The heading is static text, which nothing reads as naming the block
	grid.AddChild(block)
	d.addRows(block, sub, rows, indent+block.Border().Insets().Left)
}

// alignRows makes the labels end at the same place and the key buttons at least as wide as the widest, so that the key
// and reset columns line up across the grids a section's rows are split between: the section's own and the blocks of
// its submenus, whose labels are narrower by the blocks' indent.
func alignRows(rows []*menuKeyRow) {
	var labelEnd, keyWidth float32
	for _, row := range rows {
		_, pref, _ := row.label.Sizes(geom.Size{})
		labelEnd = max(labelEnd, row.indent+pref.Width)
		_, pref, _ = row.keyButton.Sizes(geom.Size{})
		keyWidth = max(keyWidth, pref.Width)
	}
	for _, row := range rows {
		setMinWidth(row.label, labelEnd-row.indent)
		setMinWidth(row.keyButton, keyWidth)
	}
}

// setMinWidth sets the least width the panel's FlexLayout gives it.
func setMinWidth(panel unison.Paneler, width float32) {
	if data, ok := panel.AsPanel().LayoutData().(*unison.FlexLayoutData); ok {
		data.MinSize.Width = width
	}
}

func (d *menuKeySettingsDockable) addRow(grid *unison.Panel, binding *gurps.Binding) *menuKeyRow {
	row := &menuKeyRow{binding: binding}
	row.label = NewFieldLeadingLabel(binding.Title, false)
	row.label.HAlign = align.End // The label is widened to the section's widest, so its text must keep to its end
	grid.AddChild(row.label)
	row.keyButton = unison.NewButton()
	row.keyButton.Font = unison.KeyboardFont
	row.keyButton.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	row.keyButton.ClickCallback = func() { d.editBinding(row) }
	grid.AddChild(row.keyButton)
	row.resetButton = unison.NewSVGButton(svg.Reset)
	row.resetButton.Tooltip = newWrappedTooltip(i18n.Text("Reset this key binding"))
	row.resetButton.ClickCallback = func() {
		if unison.QuestionDialog(i18n.Text("Are you sure you want to reset '%s'?", binding.Title),
			"") == unison.ModalResponseOK {
			d.resetBinding(row)
		}
	}
	row.resetButton.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Start,
		VAlign: align.Middle,
	})
	grid.AddChild(row.resetButton)
	row.refresh()
	d.rows[binding.ID] = row
	return row
}

// refresh shows the binding's current key on the key button and enables the reset button only while the key differs
// from the factory default. The key button is named after the action as well, since the key alone would not say which
// action it belongs to, and an action with no key says so rather than leaving a blank button.
func (r *menuKeyRow) refresh() {
	r.binding.KeyBinding = gurps.GlobalSettings().KeyBindings.Current(r.binding.ID)
	ink := unison.DefaultButtonTheme.OnBackgroundInk
	title := r.binding.KeyBinding.String()
	if title == "" {
		title = i18n.Text("None")
		ink = &unison.ColorFilteredInk{OriginalInk: ink, ColorFilter: unison.Alpha50Filter()}
	}
	r.keyButton.OnBackgroundInk = ink
	r.keyButton.SetTitle(title)
	r.keyButton.Accessibility.Name = i18n.Text("%s: %s", r.binding.Title, describeKeyBinding(r.binding.KeyBinding))
	r.keyButton.MarkForLayoutRecursivelyUpward() // The key's width may have changed
	r.keyButton.MarkForRedraw()
	r.resetButton.SetEnabled(r.binding.KeyBinding != gurps.FactoryKeyBinding(r.binding.ID))
}

// describeKeyBinding returns the key combination as it should be spoken, which is how it is drawn unless there is none.
func describeKeyBinding(keyBinding unison.KeyBinding) string {
	if s := keyBinding.String(); s != "" {
		return s
	}
	return i18n.Text("no key")
}

// editBinding runs the dialog that captures a new key combination for the row's binding and applies what it captured.
func (d *menuKeySettingsDockable) editBinding(row *menuKeyRow) {
	binding := row.binding
	captured := binding.KeyBinding
	prompt := unison.NewLabel()
	prompt.SetTitle(i18n.Text("Press the new key combination for %s", binding.Title))
	keyLabel := unison.NewLabel()
	keyLabel.Font = unison.KeyboardFont
	keyLabel.HAlign = align.Middle
	// The label shows only the key combination, which is empty when there is none, so the prompt names it. The
	// combination captured so far goes in its description and is announced as each one is pressed, since nothing else
	// would tell someone who cannot see the label that their key press registered.
	keyLabel.Accessibility.LabeledBy = prompt
	unison.InstallDefaultFieldBorder(keyLabel, keyLabel)
	keyLabel.DrawCallback = func(gc *unison.Canvas, rect geom.Rect) {
		gc.DrawRect(rect, unison.DefaultFieldTheme.BackgroundInk.Paint(gc, rect, paintstyle.Fill))
		keyLabel.DefaultDraw(gc, rect)
	}
	keyLabel.SetFocusable(true)
	keyLabel.SetLayoutData(&unison.FlexLayoutData{
		MinSize: geom.Size{Width: 100, Height: 50},
		HAlign:  align.Fill,
		VAlign:  align.Fill,
		HGrab:   true,
		VGrab:   true,
	})
	warning := unison.NewLabel()
	warning.Font = unison.SystemFont
	content := unison.NewPanel()
	content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	content.AddChild(prompt)
	content.AddChild(keyLabel)
	var dialog *unison.Dialog
	refresh := func() {
		keyLabel.SetTitle(captured.String())
		description := describeKeyBinding(captured)
		text := conflictWarning(binding.ID, captured)
		if text != "" {
			description += ". " + text
		}
		keyLabel.Accessibility.Description = description
		keyLabel.MarkForRedraw()
		if text == warning.String() {
			return
		}
		warning.SetTitle(text)
		switch {
		case text == "":
			warning.RemoveFromParent()
		case warning.Parent() == nil:
			content.AddChild(warning)
		}
		if dialog != nil {
			dialog.Window().Pack() // The warning takes a line of its own, and may be wider than the rest
		}
	}
	keyLabel.KeyDownCallback = func(keyCode unison.KeyCode, mods mod.Modifiers, _ bool) bool {
		switch classifyCapturedKey(keyCode, mods) {
		case keyIsModifier:
			return true // Waiting for the key it modifies
		case keyDrivesDialog:
			return false // Escape cancels, Return sets and Tab moves the focus, as in any dialog
		default:
			captured = unison.KeyBinding{KeyCode: keyCode, Modifiers: mods & mod.NonSticky}
			refresh()
			unison.AnnounceForAccessibility(keyLabel.Accessibility.Description)
			return true
		}
	}
	refresh()
	var err error
	if dialog, err = unison.NewDialog(nil, nil, content, []*unison.DialogButtonInfo{
		{
			Title:        i18n.Text("Clear"),
			ResponseCode: unison.ModalResponseUserBase,
		},
		unison.NewCancelButtonInfo(),
		unison.NewOKButtonInfoWithTitle(i18n.Text("Set")),
	}); err != nil {
		errs.Log(err)
		return
	}
	unison.DisableMenus = true
	defer func() { unison.DisableMenus = false }()
	switch dialog.RunModal() {
	case unison.ModalResponseUserBase:
		captured = unison.KeyBinding{}
		fallthrough
	case unison.ModalResponseOK:
		d.applyBinding(row, captured)
	default:
	}
}

// capturedKeyUse is what the key capture dialog does with a key press.
type capturedKeyUse int

const (
	keyIsCombination capturedKeyUse = iota // The press is the new key combination
	keyIsModifier                          // A modifier on its own, which waits for the key it modifies
	keyDrivesDialog                        // A bare Escape, Return or Tab, which cancels, sets or moves the focus
)

// classifyCapturedKey says what the key capture dialog does with a key press. A modifier on its own is no combination
// yet, and Escape, Return and Tab with no modifier drive the dialog as they would any other, since one of them alone
// would be no use as a menu key. Shift+Tab moves the focus back, so it counts as bare.
func classifyCapturedKey(keyCode unison.KeyCode, mods mod.Modifiers) capturedKeyUse {
	switch keyCode {
	case unison.KeyLShift, unison.KeyLControl, unison.KeyLOption, unison.KeyLCommand, unison.KeyRShift,
		unison.KeyRControl, unison.KeyROption, unison.KeyRCommand, unison.KeyCapsLock:
		return keyIsModifier
	case unison.KeyEscape, unison.KeyReturn, unison.KeyNumPadEnter:
		if mods&mod.NonSticky == 0 {
			return keyDrivesDialog
		}
	case unison.KeyTab:
		if mods&(mod.NonSticky&^mod.Shift) == 0 {
			return keyDrivesDialog
		}
	}
	return keyIsCombination
}

// conflictingBindings returns the bindings other than the one with the ID that are set to the key combination, in title
// order, which is none when the combination is empty.
func conflictingBindings(id string, keyBinding unison.KeyBinding) []*gurps.Binding {
	if keyBinding.IsZero() {
		return nil
	}
	var list []*gurps.Binding
	for _, b := range gurps.CurrentBindings() {
		if b.ID != id && b.KeyBinding == keyBinding {
			list = append(list, b)
		}
	}
	return list
}

// conflictWarning returns what to tell someone giving the binding with the ID the key combination about the other
// bindings that have it, which it will be taken from, or an empty string if none do.
func conflictWarning(id string, keyBinding unison.KeyBinding) string {
	others := conflictingBindings(id, keyBinding)
	if len(others) == 0 {
		return ""
	}
	titles := make([]string, len(others))
	for i, other := range others {
		titles[i] = other.Title
	}
	return i18n.Text("Already used by %s. Set will remove it from there.", strings.Join(titles, ", "))
}

// applyBinding sets the row's binding to the key combination, taking it from any other bindings that have it, and
// updates the rows of all of them.
func (d *menuKeySettingsDockable) applyBinding(row *menuKeyRow, keyBinding unison.KeyBinding) {
	g := gurps.GlobalSettings()
	others := conflictingBindings(row.binding.ID, keyBinding)
	for _, other := range others {
		g.KeyBindings.Set(other.ID, unison.KeyBinding{})
	}
	g.KeyBindings.Set(row.binding.ID, keyBinding)
	g.KeyBindings.MakeCurrent()
	for _, other := range others {
		if otherRow, ok := d.rows[other.ID]; ok { // Only the rows the filter lets through exist
			otherRow.refresh()
		}
	}
	row.refresh()
}

// resetBinding restores the row's binding to its factory default, which it takes from any other binding that has it.
func (d *menuKeySettingsDockable) resetBinding(row *menuKeyRow) {
	d.applyBinding(row, gurps.FactoryKeyBinding(row.binding.ID))
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
