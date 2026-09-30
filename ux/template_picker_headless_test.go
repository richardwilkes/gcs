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
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
)

// The picker's page reference link is outside any list, so it is an ordinary tab stop, with or without a screen reader.
func TestPickerRowPageReferenceIsFollowedFromTheKeyboard(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	const ref = "https://example.com/ref"
	var holder, link *unison.Panel
	var box *unison.CheckBox
	var linkCount int
	screen.Do(func() {
		trait := gurps.NewTrait(nil, nil, false)
		trait.Name = "Alpha"
		trait.PageRef = ref
		holder = unison.NewPanel()
		newPickerSession(promptOperation{}, []*gurps.Trait{trait}, false).addPickerRow(holder, trait, picker.Count, 0, 0, false, func() {})
		if boxes := panelsOfType[*unison.CheckBox](holder); len(boxes) == 1 {
			box = boxes[0]
		}
		links := panelsMatching(holder, func(p *unison.Panel) bool { return p.Accessibility.Role == role.Link })
		linkCount = len(links)
		if linkCount == 1 {
			link = links[0]
		}
		wnd.Content().AddChild(holder)
	})
	t.Cleanup(func() { screen.Do(holder.RemoveFromParent) })
	if box == nil {
		t.Fatal("the option must have a checkbox")
	}
	if link == nil {
		t.Fatalf("the option must end with its page reference as a link, but holds %d links", linkCount)
	}
	focus := func() (focus *unison.Panel) {
		screen.Do(func() { focus = wnd.Focus() })
		return focus
	}
	screen.Do(box.RequestFocus)
	c.Equal(box.AsPanel(), focus())
	screen.KeyPress(unison.KeyTab, mod.None)
	c.Equal(link, focus(), "Tab goes from the checkbox to the page reference")
	c.Nil(screen.OpenedURLs(), "nothing has asked for the browser yet")
	screen.KeyPress(unison.KeyReturn, mod.None)
	c.Equal([]string{ref}, screen.OpenedURLs(), "Return follows the page reference")
	screen.KeyPress(unison.KeySpace, mod.None)
	c.Equal([]string{ref}, screen.OpenedURLs(), "and so does Space")
	c.Equal(link, focus(), "the focus stays on the page reference")
}

// The picker dialog is sized with its rows' text in place, wraps its hint to the list's width, and widens when the text
// does.
func TestPickerDialogFitsItsContent(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	s, n := newKnightSession()
	// Names long enough that the rows, rather than the buttons, set the dialog's width.
	n["order"].Name = "Knightly Order of the Realm"
	n["lion"].Name = "Order of the Lion, Sworn to the Crown"
	choose(s, n, "ea", "ep", "fit", "order")
	screen.Do(func() {
		dialog, refresh := s.newPickerDialog(n["root"], 0)
		c.NotNil(dialog, "the dialog must be made")
		if dialog == nil {
			return
		}
		wnd := dialog.Window()
		defer wnd.Dispose()
		fits := func(msg string) {
			_, pref, _ := wnd.Content().Sizes(geom.Size{})
			size := wnd.ContentRect().Size
			c.True(pref.Width <= size.Width && pref.Height <= size.Height, "%s: wants %v, has %v", msg, pref, size)
		}
		fits("the rows' text is in place when the dialog is sized")
		wnd.ValidateLayout()
		hints := panelsOfType[*textLabel](wnd.Content())
		scrolls := panelsOfType[*unison.ScrollPanel](wnd.Content())
		c.Equal(1, len(hints))
		c.Equal(1, len(scrolls))
		c.True(len(hints[0].lines(hints[0].ContentRect(false).Width)) > 1, "the hint wraps")
		c.True(hints[0].FrameRect().Width <= scrolls[0].FrameRect().Width, "the hint is no wider than the list")
		choose(s, n, "lion", "honors", "cr", "wm", "fear2")
		refresh()
		fits("the dialog grows to fit longer text")
	})
}

// A picker too tall for the display keeps its place in the list when a click refreshes it.
func TestPickerDialogKeepsItsScrollPosition(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	choice := gurps.NewTrait(nil, nil, true)
	choice.TemplatePicker.Type = picker.Count
	choice.TemplatePicker.Qualifier.Qualifier = fxp.One
	for i := range 120 {
		option := gurps.NewTrait(nil, choice, false)
		option.Name = fmt.Sprintf("Option %d", i)
		choice.Children = append(choice.Children, option)
	}
	s := newPickerSession(promptOperation{}, []*gurps.Trait{choice}, false)
	screen.Do(func() {
		dialog, _ := s.newPickerDialog(choice, 0)
		c.NotNil(dialog, "the dialog must be made")
		if dialog == nil {
			return
		}
		wnd := dialog.Window()
		defer wnd.Dispose()
		wnd.ValidateLayout()
		scroll := panelsOfType[*unison.ScrollPanel](wnd.Content())[0]
		scroll.SetPosition(0, 500)
		wnd.ValidateLayout()
		_, before := scroll.Position()
		c.Equal(float32(500), before, "the list scrolls")
		size := wnd.ContentRect().Size
		panelsOfType[*unison.CheckBox](scroll.AsPanel())[60].Click()
		wnd.ValidateLayout()
		_, after := scroll.Position()
		c.Equal(before, after, "the list keeps its place")
		c.Equal(size, wnd.ContentRect().Size, "the dialog keeps its size")
		c.True(s.chosen[choice.Children[60]])
	})
}

// Override in a modifier prompt put up from the picker backs out when it changed nothing, and keeps a partial answer
// when it did.
func TestPickerModifierPromptOverride(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	s, n := newKnightSession()
	res := n["res"]
	override := func(pick bool) {
		done := false
		c.True(screen.Post(func() {
			s.chooseModifiers(res)
			done = true
		}))
		screen.Sync()
		dialogWnd, dialog := modalDialog(t, screen, wnd)
		if pick {
			var radios []*unison.RadioButton
			screen.Do(func() { radios = panelsOfType[*unison.RadioButton](dialogWnd.Content()) })
			screen.Click(screen.PanelCenter(radios[1]))
		}
		screen.Click(screen.PanelCenter(dialogButton(t, screen, dialog, unison.ModalResponseUserBase)))
		c.True(done, "the prompt has returned")
	}
	override(false)
	c.False(s.chosen[res], "Override with nothing changed backs out")
	c.False(s.modsAnswered[res])
	override(true)
	c.True(s.chosen[res], "Override keeps a partial answer")
	c.True(s.modsAnswered[res])
	c.True(res.Modifiers[0].Children[1].Enabled())
}

// An organizing group in the picker is a heading with a chevron that shows or hides its options, and no checkbox.
func TestPickerOrganizingGroupHeaders(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	s, n := newOrganizedSession()
	done := false
	c.True(screen.Post(func() {
		s.showPicker(n["root"], 0)
		done = true
	}))
	screen.Sync()
	dialogWnd, _ := modalDialog(t, screen, wnd)
	chevrons := make(map[string]*unison.Button)
	var boxes []*unison.CheckBox
	var list *unison.Panel
	var cells []*unison.Panel
	screen.Do(func() {
		for _, p := range panelsMatching(dialogWnd.Content(), func(p *unison.Panel) bool {
			return p.Accessibility.Role == role.Heading
		}) {
			label, ok := p.Self.(*unison.Label)
			c.True(ok)
			chevron, isButton := p.Parent().Children()[0].Self.(*unison.Button)
			c.True(isButton, "a header starts with its chevron")
			chevrons[label.String()] = chevron
			c.Equal(map[string]int{"martial": 1, "social": 1, "inner": 2}[label.String()], p.Accessibility.Level,
				"%s is a heading at its depth", label.String())
			c.Equal(0, len(panelsOfType[*unison.CheckBox](p.Parent())), "a header has no checkbox")
		}
		boxes = panelsOfType[*unison.CheckBox](dialogWnd.Content())
		list = boxes[0].Parent().Parent()
		cells = slices.Clone(list.Children())
	})
	c.Equal(3, len(chevrons), "each organizing group is a header")
	c.Equal(5, len(boxes), "fear, honors, rank, status and luck are options")
	expanded := func(name string) (expandable, expanded bool) {
		tree := screen.AccessibilityTree(dialogWnd)
		tree.Walk(func(node *accessibility.Node) bool {
			if node.Name == "Show or hide "+name {
				expandable, expanded = node.Expandable, node.Expanded
			}
			return true
		})
		return expandable, expanded
	}
	shownBoxes := func() (count int) {
		screen.Do(func() { count = len(panelsOfType[*unison.CheckBox](list)) })
		return count
	}
	expandable, open := expanded("martial")
	c.True(expandable && open, "a header starts open")
	screen.Do(chevrons["martial"].RequestFocus)
	var focus *unison.Panel
	tab := func() {
		screen.KeyPress(unison.KeyTab, mod.None)
		screen.Do(func() { focus = dialogWnd.Focus() })
	}
	// Where headings take the focus for a screen reader, Tab stops at the name on the way.
	tab()
	if focus.Accessibility.Role == role.Heading {
		tab()
	}
	c.Equal(boxes[0].AsPanel(), focus, "Tab goes from the chevron to the first option beneath it")

	screen.Do(chevrons["martial"].RequestFocus)
	screen.KeyPress(unison.KeySpace, mod.None)
	_, open = expanded("martial")
	c.False(open, "Space collapses the header")
	c.Equal(3, shownBoxes(), "its options are hidden")
	screen.KeyPress(unison.KeyReturn, mod.None)
	_, open = expanded("martial")
	c.True(open, "Return expands it again")
	c.False(done, "without closing the dialog")
	screen.Do(func() { c.Equal(cells, list.Children(), "every row is back in its place") })

	screen.Do(func() {
		chevrons["inner"].Click()
		chevrons["social"].Click()
	})
	c.Equal(3, shownBoxes(), "a collapsed group hides the groups within it")
	screen.Do(chevrons["social"].Click)
	c.Equal(4, shownBoxes(), "which stay as they were when it opens")
	screen.Do(func() {
		_, pref, _ := dialogWnd.Content().Sizes(geom.Size{})
		c.True(pref.Width <= dialogWnd.ContentRect().Width, "the dialog still fits")
	})
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.True(done, "the dialog has closed")
}
