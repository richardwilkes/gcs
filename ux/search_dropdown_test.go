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

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// A row of the dropdown shows the match's name followed by the title of its list, and is named for both to an
// assistive technology. A match with an icon, as the Library Explorer's have, shows it, and one without a title, as a
// root row of the Library Explorer is, is named by its name alone.
func TestSearchMatchCellShowsTheTitleAndIcon(t *testing.T) {
	c := check.New(t)
	list := unison.NewList[*searchMatch]()
	match := &searchMatch{name: "Findable", title: "Traits", show: func() {}}
	label, ok := (&searchMatchCellFactory{}).CreateCell(list, match, 0, unison.Black, unison.White, false, false).(*unison.Label)
	if !ok {
		t.Fatal("the cell must be a label")
	}
	c.Equal("Findable  Traits", label.String())
	c.Equal("Findable, Traits", label.Accessibility.Name)
	c.Nil(label.Drawable, "a match without an icon shows none")

	icon := &unison.DrawableSVG{SVG: unison.DocumentSVG}
	match = &searchMatch{name: "Library", icon: icon, show: func() {}}
	label, ok = (&searchMatchCellFactory{}).CreateCell(list, match, 0, unison.Black, unison.White, false, false).(*unison.Label)
	if !ok {
		t.Fatal("the cell must be a label")
	}
	c.Equal("Library", label.String())
	c.Equal("Library", label.Accessibility.Name)
	c.True(label.Drawable == icon, "a match with an icon shows it")
}

// A row that only carries a note, as the one saying the matches were cut short is, is set apart from the matches: it
// is drawn dimmed in the italic secondary font, under a line, and shows no icon.
func TestSearchMatchCellSetsANoteApart(t *testing.T) {
	c := check.New(t)
	list := unison.NewList[*searchMatch]()
	factory := &searchMatchCellFactory{}
	match, ok := factory.CreateCell(list, &searchMatch{name: "Findable", show: func() {}}, 0, unison.Black, unison.White,
		false, false).(*unison.Label)
	if !ok {
		t.Fatal("the cell must be a label")
	}
	note, ok := factory.CreateCell(list, &searchMatch{name: "Narrow the search"}, 0, unison.Black, unison.White, false,
		false).(*unison.Label)
	if !ok {
		t.Fatal("the cell must be a label")
	}
	c.Equal("Narrow the search", note.String())
	c.Equal("Narrow the search", note.Accessibility.Name)
	c.Nil(note.Drawable)
	want := noteFont()
	c.True(note.Font.Face() == want.Face() && note.Font.Size() == want.Size(), "a note is drawn in the note font")
	var isDimmed bool
	note.Text.AdjustDecorations(func(d *unison.TextDecoration) { _, isDimmed = d.OnBackgroundInk.(*unison.ColorFilteredInk) })
	c.True(isDimmed, "a note is dimmed")
	c.Equal(match.Border().Insets().Top+1, note.Border().Insets().Top, "a note sits under a line")
	c.Equal(match.Border().Insets().Bottom, note.Border().Insets().Bottom)
}

// The search field switches to the focus border while it holds the focus, as other fields do, and still searches for
// the text it holds when it gains the focus.
func TestSearchFieldShowsTheFocusBorder(t *testing.T) {
	c := check.New(t)
	walks := 0
	s := newSearchDropdown(func(string) []*searchMatch {
		walks++
		return nil
	}, func(*searchMatch) {})
	unfocused := s.field.Border()
	s.field.SetText("findable")
	s.field.GainedFocusCallback()
	c.True(s.field.Border() != unfocused, "gaining the focus shows the focus border")
	c.Equal(1, walks, "and searches for the text")
	s.field.LostFocusCallback()
	c.True(s.field.Border() == unfocused, "losing the focus puts the unfocused border back")
}
