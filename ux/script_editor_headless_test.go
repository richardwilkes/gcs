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
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// waitForEvaluation waits out scriptEvaluationDelay, then for the work it put off.
func waitForEvaluation(screen *unison.HeadlessScreen) {
	time.Sleep(2 * scriptEvaluationDelay)
	screen.Sync()
}

func TestScriptEditor(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	script := "a\nb"
	var editor *scriptEditor
	var evaluations int
	screen.Do(func() {
		editor = newScriptEditor(func() string { return script }, func(s string) { script = s }, &scriptEditorOptions{
			Title:               "Script",
			KeepFirstLinePrefix: "// prereq count:",
			Inserts:             []scriptMenuEntry{{Label: "Fn", Text: "fn()", CaretFromEnd: 1}},
			Snippets:            []scriptMenuEntry{{Label: "Pick one", Heading: true}, {Label: "True", Text: "true"}},
			Evaluate: func(s string) (gurps.PrereqResult, string) {
				evaluations++
				if strings.Contains(s, "true") {
					return gurps.PrereqMet, "Met"
				}
				return gurps.PrereqUnmet, "Not met"
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
	c.Equal("a\n  b", script, "Shift+Tab changes nothing")
	screen.Do(func() { c.Equal(editor.field.AsPanel(), editor.field.Window().Focus(), "nor moves the focus") })

	screen.Do(func() {
		editor.field.SetSelection(1, 1)
		editor.insertEntry(1, 1, editor.opts.Inserts[0])
	})
	c.Equal("afn()\n  b", script, "an insert goes in at the caret")
	start, end := selection()
	c.Equal(4, start, "the caret goes where the entry asks")
	c.Equal(4, end)
	screen.Do(func() { c.Equal(editor.field.AsPanel(), editor.field.Window().Focus(), "the field takes the focus") })

	screen.Do(func() { editor.insertEntry(0, len(editor.field.Text()), editor.opts.Snippets[1]) })
	c.Equal("true", script, "a snippet replaces the script")
	c.Equal("Not met", result(), "the result waits for the script to go unchanged")
	waitForEvaluation(screen)
	c.Equal("Met", result(), "then follows the change")
	var before, after int
	screen.Do(func() {
		before = evaluations
		editor.field.SetText("1")
		editor.field.SetText("true")
	})
	waitForEvaluation(screen)
	screen.Do(func() { after = evaluations })
	c.Equal(before+1, after, "changes in quick succession are evaluated once")

	var snippets *unison.Button
	screen.Do(func() {
		for _, one := range panelsOfType[*unison.Button](editor.AsPanel()) {
			if one.Text != nil && one.Text.String() == "Snippets" {
				snippets = one
			}
		}
		editor.field.SetText("«x»") // Longer in bytes than in runes.
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
		editor.insertEntry(0, len(editor.field.Text()), editor.opts.Snippets[1])
	})
	c.Equal("// prereq count: 3\ntrue", script, "a snippet keeps the prereq count line")
	screen.Do(func() {
		editor.field.SetText("// Prereq Count: 2")
		editor.insertEntry(0, 0, editor.opts.Inserts[0])
	})
	c.Equal("// Prereq Count: 2\nfn()", script, "nothing goes in above the prereq count line")
}

// TestShowCheckIconMakesRoom checks that a status icon shown in a label that had none, after its row was laid out, is
// given room at the next layout.
func TestShowCheckIconMakesRoom(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	var row *unison.Panel
	var icon *unison.Label
	screen.Do(func() {
		row = unison.NewPanel()
		icon = unison.NewLabel()
		row.AddChild(icon)
		hbox(row, 0)
	})
	w := showInTestWindow(t, screen, 300, row)
	screen.Do(func() {
		showCheckIcon(icon, gurps.PrereqMet)
		w.ValidateLayout()
		c.True(icon.FrameRect().Width > 0)
	})
}
