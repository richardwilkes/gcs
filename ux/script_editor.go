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
	"time"
	"unicode/utf8"

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

// scriptEvaluationDelay is how long a script must go unchanged before it is evaluated, so that one that runs away
// doesn't stall every keystroke.
const scriptEvaluationDelay = 250 * time.Millisecond

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
	// Title is the field's accessible name and its undo title, and Hint is shown beside the menus.
	Title string
	Hint  string
	// KeepFirstLinePrefix, when set, starts a first line, matched without regard to case, that inserts and snippets
	// leave in place.
	KeepFirstLinePrefix string
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
	opts   *scriptEditorOptions
	icon   *unison.Label
	result *sentenceButton
}

// newScriptEditor returns a script editor for the script the accessors reach.
func newScriptEditor(get func() string, set func(string), opts *scriptEditorOptions) *scriptEditor {
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
	hint.SetTitle(opts.Hint)
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
	frame := unison.NewPanel()
	frame.SetLayout(&unison.FlexLayout{Columns: 1})
	frame.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
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
		row.AddChild(e.icon)
		e.result = newSentenceButton("", nil, nil)
		putOnLine(e.icon.AsPanel(), e.result.lineHeight(), checkIconSize())
		e.result.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
		row.AddChild(e.result)
		e.AddChild(hbox(row, unison.StdIconGap))
		e.evaluate(e.field.CurrentValue())
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

// put replaces the runes from start to end with the entry's text, below a first line starting with the
// KeepFirstLinePrefix, and leaves the caret where the entry asks. The order of the calls matters: gaining the focus
// selects everything, so the selection is set last.
func (e *scriptEditor) put(start, end int, entry scriptMenuEntry) {
	text := []rune(e.field.Text())
	if keep := e.opts.KeepFirstLinePrefix; keep != "" && strings.HasPrefix(strings.ToLower(string(text)), strings.ToLower(keep)) {
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

// refresh evaluates the script again and shows the result, once it has gone unchanged for scriptEvaluationDelay. It is
// called for each change to the script; call it when whatever the script reads changes.
func (e *scriptEditor) refresh() {
	if e.opts.Evaluate == nil {
		return
	}
	script := e.field.CurrentValue()
	unison.InvokeTaskAfter(func() {
		if script == e.field.CurrentValue() {
			e.evaluate(script)
		}
	}, scriptEvaluationDelay)
}

// evaluate shows the result of evaluating the script.
func (e *scriptEditor) evaluate(script string) {
	var status checkStatus
	var text string
	gurps.SuppressScriptResolveErrorLogging(func() { status, text = e.opts.Evaluate(script) })
	showCheckIcon(e.icon, status)
	e.result.setText(text, "")
}
