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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison/accessibility"
)

// TestOptionalDecimalFieldBlankMeansNone verifies that no value shows as a blank field and that zero is a value of its
// own. Blank meaning none is covered through the template's target fields.
func TestOptionalDecimalFieldBlankMeansNone(t *testing.T) {
	c := check.New(t)
	var value *fxp.Int
	f := NewOptionalDecimalField(nil, "", "Value",
		func() *fxp.Int { return value },
		func(v *fxp.Int) { value = v })
	c.Equal("", f.Text(), "no value shows as a blank field")

	f.SetText("0")
	c.NotNil(value, "zero is a value of its own")
	c.Equal(fxp.Int(0), *value)
}

// TestOptionalDecimalFieldKeepsItsValueForTextItCannotParse verifies that text a focused field can't parse, such as
// pasted text or the start of a number, a sign or a decimal point by itself, which is typed as it would be on the way
// to a number and which fxp.FromString takes for zero, shows as invalid without changing the value, rather than
// clearing it to none or committing the zero; that undoing an edit made over such text restores the value from before
// it; that leaving the field puts the value's text back; and that an assistive technology handing such text to the
// field without the focus is refused. Text that is not the start of a number is refused when typed.
func TestOptionalDecimalFieldKeepsItsValueForTextItCannotParse(t *testing.T) {
	ten := optionalDecimal{value: fxp.Ten, present: true}
	for _, one := range []struct {
		text  string
		typed bool // The text is the start of a number, which is typed rather than pasted.
	}{
		{text: "abc"},
		{text: "-", typed: true},
		{text: "+", typed: true},
		{text: ".", typed: true},
		{text: "-.", typed: true},
		{text: "+.", typed: true},
	} {
		text := one.text
		t.Run(text, func(t *testing.T) {
			c := check.New(t)
			value := new(fxp.Ten)
			f := NewOptionalDecimalField(nil, "", "Value", func() *fxp.Int { return value },
				func(v *fxp.Int) { value = v })
			root := newPageUndoRoot()
			root.AddChild(f)
			f.gainedFocus()
			if one.typed {
				f.SelectAll()
				for _, ch := range text {
					c.True(f.RuneTypedCallback(ch), "typing %q is accepted", ch)
				}
			} else {
				c.False(f.RuneTypedCallback(rune(text[0])), "typing %q is refused", text[:1])
				c.Equal("10", f.Text(), "which leaves the text as it was")
				f.SetText(text)
			}
			c.Equal(text, f.Text(), "the text shows")
			c.True(f.Invalid(), "as invalid")
			c.Equal(ten, optionalDecimalOf(value), "without changing the value")

			f.SelectAll() // Has the next edit recorded on its own.
			f.SetText("7")
			c.Equal(optionalDecimal{value: fxp.Seven, present: true}, optionalDecimalOf(value))
			root.mgr.Undo()
			c.Equal(ten, optionalDecimalOf(value), "undoing an edit made over the text restores the value from before it")

			f.SetText(text)
			f.lostFocus()
			c.Equal("10", f.Text(), "leaving the field puts the value's text back")
			c.Equal(ten, optionalDecimalOf(value))

			for _, req := range []accessibility.ActionRequest{
				{Action: accessibility.SetValue, Value: text},
				{Action: accessibility.ReplaceText, End: len([]rune(f.Text())), Value: text},
			} {
				c.False(f.PerformAccessibilityAction(req), "%v: the text is refused without the focus", req.Action)
			}
			c.Equal("10", f.Text(), "which leaves the value's text showing")
			c.Equal(ten, optionalDecimalOf(value))
		})
	}
}

// TestOptionalDecimalFieldUnfocusedAccessibilityEditsWithoutANewValue verifies that an assistive technology handing a
// field without the focus the value it holds, however written, leaves it as it was: the value's text still shows, as
// nothing syncs the field after an edit that sets no new value, and no undo is recorded for what changed nothing.
func TestOptionalDecimalFieldUnfocusedAccessibilityEditsWithoutANewValue(t *testing.T) {
	for _, one := range []struct {
		name string
		req  accessibility.ActionRequest
	}{
		{
			name: "setting the value written differently",
			req:  accessibility.ActionRequest{Action: accessibility.SetValue, Value: "10.0"},
		},
		{
			name: "inserting text that leaves the value as it is",
			req:  accessibility.ActionRequest{Action: accessibility.ReplaceText, Start: 2, End: 2, Value: ".0"},
		},
	} {
		t.Run(one.name, func(t *testing.T) {
			c := check.New(t)
			value := new(fxp.Ten)
			f := NewOptionalDecimalField(nil, "", "Value", func() *fxp.Int { return value },
				func(v *fxp.Int) { value = v })
			root := newPageUndoRoot()
			root.AddChild(f)
			c.True(f.PerformAccessibilityAction(one.req), "the request is carried out")
			c.Equal(optionalDecimal{value: fxp.Ten, present: true}, optionalDecimalOf(value), "the value is unchanged")
			c.Equal("10", f.Text(), "the value's text still shows")
			c.False(root.mgr.CanUndo(), "nothing is recorded")
		})
	}
}

// TestOptionalDecimalFieldUnfocusedAccessibilityEditsSetANewValue verifies that an assistive technology handing a field
// without the focus a new value, or blank text for none, sets it as an edit that can be undone.
func TestOptionalDecimalFieldUnfocusedAccessibilityEditsSetANewValue(t *testing.T) {
	ten := optionalDecimal{value: fxp.Ten, present: true}
	for _, one := range []struct {
		name string
		req  accessibility.ActionRequest
		want optionalDecimal
	}{
		{
			name: "setting a new value",
			req:  accessibility.ActionRequest{Action: accessibility.SetValue, Value: "12"},
			want: optionalDecimal{value: fxp.Twelve, present: true},
		},
		{
			name: "replacing part of the text",
			req:  accessibility.ActionRequest{Action: accessibility.ReplaceText, Start: 0, End: 1, Value: "3"},
			want: optionalDecimal{value: fxp.FromInteger(30), present: true},
		},
		{
			name: "setting blank text",
			req:  accessibility.ActionRequest{Action: accessibility.SetValue},
		},
	} {
		t.Run(one.name, func(t *testing.T) {
			c := check.New(t)
			value := new(fxp.Ten)
			f := NewOptionalDecimalField(nil, "", "Value", func() *fxp.Int { return value },
				func(v *fxp.Int) { value = v })
			root := newPageUndoRoot()
			root.AddChild(f)
			c.True(f.PerformAccessibilityAction(one.req), "the request is carried out")
			c.Equal(one.want, optionalDecimalOf(value), "the new value is set")
			c.Equal(formatOptionalDecimal(one.want), f.Text(), "and its text shows")
			c.True(root.mgr.CanUndo(), "the change is recorded")
			root.mgr.Undo()
			c.Equal(ten, optionalDecimalOf(value), "undoing it restores the value from before it")
		})
	}
}
