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
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/role"
)

// scriptKeepLinePrefix starts a first line that inserts and snippets leave in place, since it is read separately from
// the script (see the spell prereq count).
const scriptKeepLinePrefix = "// prereq count:"

// checkStatus is the outcome of checking a requirement, such as a script, against a sheet.
type checkStatus uint8

// Possible checkStatus values.
const (
	checkMet checkStatus = iota
	checkUnmet
	checkFailed
	checkSkipped
)

// showCheckIcon has the label show the icon of the status.
func showCheckIcon(label *unison.Label, status checkStatus) {
	icon, ink := unison.CheckmarkSVG, unison.Ink(colors.Success)
	switch status {
	case checkUnmet:
		icon, ink = svg.Not, colors.Failure
	case checkFailed:
		icon, ink = unison.TriangleExclamationSVG, unison.ThemeWarning
	case checkSkipped:
		icon, ink = unison.DashSVG, faint(unison.ThemeOnSurface)
	default:
	}
	if label.Drawable == nil {
		// Without an icon the label took no room, so the panels around it must be laid out again.
		label.MarkForLayoutRecursivelyUpward()
	}
	size := unison.DefaultLabelTheme.Font.Baseline()
	label.Drawable = &unison.DrawableSVG{SVG: icon, Size: geom.NewSize(size, size).Ceil()}
	label.OnBackgroundInk = ink
	label.MarkForRedraw()
}

// faint returns the ink at 30% opacity.
func faint(ink unison.Ink) unison.Ink {
	return &unison.ColorFilteredInk{OriginalInk: ink, ColorFilter: unison.Alpha30Filter()}
}

// scriptMenuEntry is one item of the script editor's Insert or Snippets menu.
type scriptMenuEntry struct {
	// Label is what the menu shows. A heading is shown disabled, after a separator unless it comes first.
	Label   string
	Heading bool
	// Text is what the entry inserts, and CaretFromEnd is how many runes before its end the caret is left.
	Text         string
	CaretFromEnd int
}

// scriptEditorOptions configures a script editor.
type scriptEditorOptions struct {
	// Title is the field's accessible name and its undo title.
	Title string
	// Inserts go in at the caret; Snippets replace the script. Either menu is left out when it has no entries.
	Inserts  []scriptMenuEntry
	Snippets []scriptMenuEntry
	// Evaluate, when set, runs the script for the result line, whose text should say the outcome.
	Evaluate func(script string) (status checkStatus, text string)
}

// scriptEditor edits a script in a monospaced field under a toolbar of Insert and Snippets menus, with an optional line
// showing the result of evaluating it. Changes go through the setter; undo is left to the field's undo manager.
type scriptEditor struct {
	unison.Panel
	field  *StringField
	opts   scriptEditorOptions
	icon   *unison.Label
	result *sentenceButton
}

// newScriptEditor returns a script editor for the script the accessors reach.
func newScriptEditor(get func() string, set func(string), opts scriptEditorOptions) *scriptEditor {
	e := &scriptEditor{opts: opts}
	e.Self = e
	e.SetLayout(&unison.FlexLayout{Columns: 1, VSpacing: unison.StdVSpacing})
	e.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})

	bar := unison.NewPanel()
	e.addMenuButton(bar, i18n.Text("Insert"), opts.Inserts, func(entry scriptMenuEntry) {
		start, end := e.field.Selection()
		e.put(start, end, entry)
	})
	e.addMenuButton(bar, i18n.Text("Snippets"), opts.Snippets, func(entry scriptMenuEntry) {
		e.put(0, utf8.RuneCountInString(e.field.Text()), entry)
	})
	hint := unison.NewLabel()
	hint.SetTitle(i18n.Text("Tab indents. Esc closes."))
	hint.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, VAlign: align.Middle, HGrab: true})
	bar.AddChild(hint)
	guide := unison.NewSVGButton(svg.Script)
	guide.ClickCallback = func() { HandleLink(nil, "md:User%20Guide/Scripting%20Guide") }
	guide.Tooltip = newWrappedTooltip(i18n.Text("Scripting Guide"))
	bar.AddChild(guide)
	bar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{},
		geom.Insets{Bottom: 1}, false), unison.NewEmptyBorder(geom.NewUniformInsets(4))))
	bar.DrawCallback = func(gc *unison.Canvas, r geom.Rect) {
		gc.DrawRect(r, unison.ThemeSurface.Paint(gc, r, paintstyle.Fill))
	}
	frame := newPrereqColumn()
	frame.SetLayout(&unison.FlexLayout{Columns: 1})
	frame.AddChild(hbox(bar, unison.StdHSpacing))
	e.AddChild(frame)

	e.field = NewMultiLineStringField(nil, "", opts.Title, get, func(script string) {
		set(script)
		e.refresh()
	})
	e.field.AutoScroll = false
	e.field.SetMinimumTextWidthUsing("floor($basic_speed)")
	e.field.Font = &unison.DynamicFont{
		Resolver: func() unison.FontDescriptor {
			fd := unison.MonospacedFont.Font.Descriptor()
			fd.Size = unison.DefaultFieldTheme.Font.Size()
			return fd
		},
	}
	keyDown := e.field.KeyDownCallback
	e.field.KeyDownCallback = func(keyCode unison.KeyCode, mods mod.Modifiers, repeat bool) bool {
		if keyCode == unison.KeyTab && mods&mod.NonSticky == 0 {
			e.field.DefaultRuneTyped(' ')
			e.field.DefaultRuneTyped(' ')
			return true
		}
		return keyDown(keyCode, mods, repeat)
	}
	// The field fills a box under the toolbar, which shows the field's focus, and is never shorter than 5 lines.
	unison.UninstallFocusBorders(e.field, e.field)
	e.field.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: 4, Left: 6, Bottom: 4, Right: 6}))
	frameBorder := func(ink unison.Ink) unison.Border {
		return unison.NewLineBorder(ink, geom.NewUniformSize(4), geom.NewUniformInsets(1), false)
	}
	unison.InstallFocusBorders(e.field, frame, frameBorder(unison.ThemeFocus), frameBorder(unison.ThemeSurfaceEdge))
	e.field.SetSizer(func(hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
		minSize, prefSize, maxSize = e.field.DefaultSizes(hint)
		height := 5*e.field.Font.LineHeight() + e.field.Border().Insets().Height()
		minSize.Height, prefSize.Height = max(minSize.Height, height), max(prefSize.Height, height)
		return minSize, prefSize, maxSize
	})
	frame.AddChild(e.field)

	if opts.Evaluate != nil {
		row := unison.NewPanel()
		e.icon = unison.NewLabel()
		e.icon.Accessibility.Role = role.None // The text says the outcome.
		e.icon.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: 2}))
		e.icon.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Start})
		row.AddChild(e.icon)
		e.result = newSentenceButton("", nil, nil)
		e.result.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
		row.AddChild(e.result)
		e.AddChild(hbox(row, unison.StdIconGap))
		e.refresh()
	}
	return e
}

func (e *scriptEditor) addMenuButton(bar *unison.Panel, title string, entries []scriptMenuEntry, act func(scriptMenuEntry)) {
	if len(entries) == 0 {
		return
	}
	button := unison.NewButton()
	button.SetTitle(title)
	button.ClickCallback = func() {
		menu := make([]menuEntry, len(entries))
		for i, entry := range entries {
			menu[i].Label = entry.Label
			if !entry.Heading {
				menu[i].Act = func() { act(entry) }
			}
		}
		showMenu(button.AsPanel(), menu)
	}
	bar.AddChild(button)
}

// menuEntry is one item of a menu that showMenu builds. One with no action is a heading, shown disabled after a
// separator unless it comes first; with no label as well, it is just the separator.
type menuEntry struct {
	Label string
	Act   func()
}

// showMenu pops up a menu of the entries below the anchor.
func showMenu(anchor *unison.Panel, entries []menuEntry) {
	f := unison.DefaultMenuFactory()
	id := unison.ContextMenuIDFlag
	m := f.NewMenu(id, "", nil)
	for i, entry := range entries {
		id++
		if entry.Act != nil {
			m.InsertItem(-1, f.NewItem(id, entry.Label, unison.KeyBinding{}, nil, func(unison.MenuItem) { entry.Act() }))
			continue
		}
		if i != 0 {
			m.InsertSeparator(-1, false)
		}
		if entry.Label != "" {
			m.InsertItem(-1, f.NewItem(id, entry.Label, unison.KeyBinding{}, func(unison.MenuItem) bool { return false }, nil))
		}
	}
	m.Popup(anchor.RectToRoot(anchor.ContentRect(true)), 0)
}

// put replaces the runes from start to end with the entry's text, below a first line starting with
// scriptKeepLinePrefix, and leaves the caret where the entry asks. The order of the calls matters: gaining the focus
// selects everything, so the selection is set last.
func (e *scriptEditor) put(start, end int, entry scriptMenuEntry) {
	text := []rune(e.field.Text())
	if strings.HasPrefix(strings.ToLower(string(text)), scriptKeepLinePrefix) {
		keep := slices.Index(text, '\n') + 1
		if keep == 0 {
			text = append(text, '\n')
			keep = len(text)
		}
		start, end = max(start, keep), max(end, keep)
	}
	insert := []rune(entry.Text)
	e.field.SetText(string(text[:start]) + string(insert) + string(text[end:]))
	e.field.RequestFocus()
	caret := start + len(insert) - entry.CaretFromEnd
	e.field.SetSelection(caret, caret)
}

// refresh evaluates the script again and shows the result. It is called for each change to the script; call it when
// whatever the script reads changes.
func (e *scriptEditor) refresh() {
	if e.opts.Evaluate == nil {
		return
	}
	var status checkStatus
	var text string
	gurps.SuppressScriptResolveErrorLogging(func() { status, text = e.opts.Evaluate(e.field.CurrentValue()) })
	showCheckIcon(e.icon, status)
	e.result.setText(text, "")
}
