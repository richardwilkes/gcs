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
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

func TestScriptEditor(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	script := "a\nb"
	var editor *scriptEditor
	screen.Do(func() {
		editor = newScriptEditor(func() string { return script }, func(s string) { script = s }, scriptEditorOptions{
			Title:    "Script",
			Inserts:  []scriptMenuEntry{{Label: "Fn", Text: "fn()", CaretFromEnd: 1}},
			Snippets: []scriptMenuEntry{{Label: "Pick one", Heading: true}, {Label: "True", Text: "true"}},
			Evaluate: func(s string) (scriptStatus, string) {
				if strings.Contains(s, "true") {
					return scriptMet, "Met"
				}
				return scriptUnmet, "Not met"
			},
		})
	})
	showInTestWindow(t, screen, 500, editor)
	result := func() (text string) {
		screen.Do(func() { text = editor.result.plainText() })
		return text
	}
	selection := func() (start, end int) {
		screen.Do(func() { start, end = editor.field.Selection() })
		return start, end
	}
	c.Equal("Not met", result())

	screen.Do(func() {
		editor.field.RequestFocus()
		editor.field.SetSelection(2, 2)
	})
	screen.KeyPress(unison.KeyTab, mod.None)
	c.Equal("a\n  b", script, "Tab inserts two spaces at the caret")
	screen.Do(func() { c.Equal(editor.field.AsPanel(), editor.field.Window().Focus(), "and keeps the focus") })
	screen.KeyPress(unison.KeyTab, mod.Shift)
	screen.Do(func() { c.NotEqual(editor.field.AsPanel(), editor.field.Window().Focus(), "Shift+Tab leaves") })

	screen.Do(func() {
		editor.field.SetSelection(1, 1)
		editor.put(1, 1, editor.opts.Inserts[0])
	})
	c.Equal("afn()\n  b", script, "an insert goes in at the caret")
	start, end := selection()
	c.Equal(4, start, "the caret goes where the entry asks")
	c.Equal(4, end)
	screen.Do(func() { c.Equal(editor.field.AsPanel(), editor.field.Window().Focus(), "the field takes the focus") })

	screen.Do(func() { editor.put(0, len(editor.field.Text()), editor.opts.Snippets[1]) })
	c.Equal("true", script, "a snippet replaces the script")
	c.Equal("Met", result(), "the result follows each change")

	var snippets *unison.Button
	screen.Do(func() {
		for _, one := range panelsOfType[*unison.Button](editor.AsPanel()) {
			if one.Text != nil && one.Text.String() == "Snippets" {
				snippets = one
			}
		}
		editor.field.SetText("x")
	})
	if snippets == nil {
		t.Fatal("the editor must have a Snippets button")
	}
	screen.Click(screen.PanelCenter(snippets))
	screen.KeyPress(unison.KeyDown, mod.None)
	screen.KeyPress(unison.KeyReturn, mod.None)
	c.Equal("true", script, "the Snippets menu skips its heading and applies the snippet chosen")

	screen.Do(func() {
		editor.field.SetText("// prereq count: 3\nold")
		editor.put(0, len(editor.field.Text()), editor.opts.Snippets[1])
	})
	c.Equal("// prereq count: 3\ntrue", script, "a snippet keeps the prereq count line")
	screen.Do(func() {
		editor.field.SetText("// Prereq Count: 2")
		editor.put(0, 0, editor.opts.Inserts[0])
	})
	c.Equal("// Prereq Count: 2\nfn()", script, "nothing goes in above the prereq count line")
}
