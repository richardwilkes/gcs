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
		list := &pickerList{panel: holder, pt: picker.Count, refresh: func() {}}
		newPickerSession(promptOperation{}, []*gurps.Trait{trait}, false).addPickerRow(list, trait, nil, false, 0)
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

// The line under the picker's list is in the warning color for an error or a warning, and green once all is well.
func TestPickerDialogHintColor(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	for _, tc := range []struct {
		row   string
		picks []string
		state pickerState
		ink   unison.Ink
	}{
		{"root", []string{"ea", "ep", "fit", "order", "lion", "honors", "cr", "wm", "fear2"}, pickerError, unison.ThemeWarning},
		{"order", []string{"order", "lion", "honors", "cr", "wm", "fear2"}, pickerWarning, unison.ThemeWarning},
		{"fit", []string{"fit", "fit1"}, pickerOK, unison.Green},
	} {
		s, n := newKnightSession()
		choose(s, n, tc.picks...)
		c.Equal(tc.state, s.state(n[tc.row]), tc.row)
		screen.Do(func() {
			dialog, _ := s.newPickerDialog(n[tc.row], 0)
			c.NotNil(dialog, "the dialog must be made")
			if dialog == nil {
				return
			}
			defer dialog.Window().Dispose()
			hints := panelsOfType[*textLabel](dialog.Window().Content())
			c.Equal(1, len(hints))
			c.Equal(tc.ink, hints[0].ink, tc.row)
		})
	}
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

// With no chevron among them, the picker's options leave no room for one, so their names follow their checkboxes.
func TestPickerRowsWithoutGroupsLeaveNoRoomForChevrons(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	s, n := newKnightSession()
	screen.Do(func() {
		dialog, _ := s.newPickerDialog(n["root"], 0)
		c.NotNil(dialog, "the dialog must be made")
		if dialog == nil {
			return
		}
		wnd := dialog.Window()
		defer wnd.Dispose()
		wnd.ValidateLayout()
		list := panelsOfType[*unison.CheckBox](wnd.Content())[0].Parent().Parent()
		checkPickerRowPlaces(c, list, map[string]int{"ea": 0, "ep": 0, "fit": 0, "order": 0, "luck": 0, "shield": 0})
	})
}

// An organizing group in the picker is a heading with a chevron that shows or hides its options, and no checkbox. Its
// page reference lines up with theirs, and its chevron and name with those of the rows beside it.
func TestPickerOrganizingGroupHeaders(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	s, n := newOrganizedSession()
	for _, name := range []string{"martial", "inner", "fear", "status", "luck"} {
		n[name].PageRef = "B10"
	}
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
		checkPickerRowsAligned(c, list, 5)
		// Only the groups at the top and within social hold groups, so only their rows leave room for a chevron.
		checkPickerRowPlaces(c, list, map[string]int{
			"martial": 1, "fear": 1, "honors": 1,
			"social": 1, "rank": 2, "inner": 2, "status": 2,
			"luck": 1,
		})
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
	c.Equal(role.Link, focus.Accessibility.Role, "Tab goes from the chevron to the header's page reference")
	tab()
	c.Equal(boxes[0].AsPanel(), focus, "and then to the first option beneath it")

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

// A container picked as a unit keeps its checkbox and cost, and its chevron shows what it holds as rows with nothing to
// pick, widening the dialog to fit them, their page references lined up with those of the options.
func TestPickerUnitContainerInformationRows(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	trait := func(name string, points int, children ...*gurps.Trait) *gurps.Trait {
		one := gurps.NewTrait(nil, nil, len(children) != 0)
		one.Name = name
		one.BasePoints = fxp.FromInteger(points)
		one.Children = children
		SetParents(children, one)
		return one
	}
	const long = "Sword of the Realm, Forged in the Fires Beneath the Mountain for the Order's First Champion"
	kit := trait("Kit", 0, trait(long, 5), trait("Pack", 0, trait("Rope", 1)))
	choice := trait("Choice", 0, kit, trait("Luck", 15))
	gurps.Traverse(func(one *gurps.Trait) bool {
		one.PageRef = "B10"
		return false
	}, false, false, kit, choice.Children[1])
	choice.TemplatePicker.Type = picker.Count
	choice.TemplatePicker.Qualifier.Qualifier = fxp.One
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
		boxes := panelsOfType[*unison.CheckBox](wnd.Content())
		c.Equal(2, len(boxes), "only the options have checkboxes")
		row := boxes[0].Parent()
		c.Equal(boxes[0].AsPanel(), row.Children()[0], "the row starts with its checkbox")
		chevron, ok := row.Children()[1].Children()[0].Self.(*unison.Button)
		c.True(ok, "and then its chevron")
		if !ok {
			return
		}
		c.NotEqual("", labelTexts(row)[2], "the unit keeps its cost")
		c.False(slices.Contains(labelTexts(wnd.Content()), "Rope"), "what it holds starts hidden")
		width := wnd.ContentRect().Width
		chevron.Click()
		wnd.ValidateLayout()
		checkPickerRowsAligned(c, row.Parent(), 5)
		// What the unit holds is set in one level past its name.
		checkPickerRowPlaces(c, row.Parent(), map[string]int{"Kit": 1, "Luck": 1, long: 2, "Pack": 2, "Rope": 3})
		texts := labelTexts(wnd.Content())
		c.True(slices.Contains(texts, long) && slices.Contains(texts, "Pack") && slices.Contains(texts, "Rope"),
			"opening it shows what it holds, however deep")
		c.Equal(2, len(panelsOfType[*unison.CheckBox](wnd.Content())), "with nothing to pick")
		_, pref, _ := wnd.Content().Sizes(geom.Size{})
		c.True(wnd.ContentRect().Width > width, "the dialog widens")
		c.True(pref.Width <= wnd.ContentRect().Width, "to fit them")
		cells := slices.Clone(row.Parent().Children())
		chevron.Click()
		c.False(slices.Contains(labelTexts(wnd.Content()), "Rope"), "closing it hides them again")
		chevron.Click()
		c.Equal(cells, row.Parent().Children(), "and opening it puts every cell back in its place")
	})
}

// checkPickerRowsAligned checks that every row of the picker's list has a cell in each of its columns, none spanning
// more, and that it shows want page references, all alike, lined up with one another.
func checkPickerRowsAligned(c check.Checker, list *unison.Panel, want int) {
	layout, ok := list.Layout().(*unison.FlexLayout)
	c.True(ok, "the list is laid out in columns")
	if !ok {
		return
	}
	cells := list.Children()
	c.Equal(0, len(cells)%layout.Columns, "every row has a cell in each column")
	for _, cell := range cells {
		if data, isFlex := cell.LayoutData().(*unison.FlexLayoutData); isFlex {
			c.True(data.HSpan <= 1, "no cell spans columns")
		}
	}
	list.ValidateLayout()
	links := panelsMatching(list, func(p *unison.Panel) bool { return p.Accessibility.Role == role.Link })
	c.Equal(want, len(links), "each row shown has its page reference")
	for _, link := range links[1:] {
		c.Equal(links[0].RectTo(links[0].ContentRect(false), list).X, link.RectTo(link.ContentRect(false), list).X,
			"the page references line up")
	}
}

// checkPickerRowPlaces checks that the rows of the picker's list lead with their checkboxes, or room for one, all in one
// column, and that the name of each row given is set that many chevron widths past it, just after its chevron if it has
// one.
func checkPickerRowPlaces(c check.Checker, list *unison.Panel, places map[string]int) {
	layout, ok := list.Layout().(*unison.FlexLayout)
	c.True(ok, "the list is laid out in columns")
	if !ok {
		return
	}
	list.ValidateLayout()
	x := func(p *unison.Panel) float32 { return p.RectTo(p.ContentRect(false), list).X }
	boxes := panelsOfType[*unison.CheckBox](list)
	c.NotEqual(0, len(boxes), "there are options")
	if len(boxes) == 0 {
		return
	}
	left := x(boxes[0].AsPanel())
	cells := list.Children()
	for i := 0; i < len(cells); i += layout.Columns {
		if len(cells[i].Children()) == 2 {
			c.Equal(left, x(cells[i].Children()[0]), "every row leads with a checkbox, or room for one, in one column")
		}
	}
	base := left + pickerCheckBoxSize().Width + unison.StdHSpacing
	step := pickerDisclosureSize() + unison.StdHSpacing
	for name, level := range places {
		labels := panelsMatching(list, func(p *unison.Panel) bool {
			label, isLabel := p.Self.(*unison.Label)
			return isLabel && label.String() == name
		})
		c.Equal(1, len(labels), "%s is shown once", name)
		if len(labels) != 1 {
			continue
		}
		c.Equal(base+float32(level)*step, x(labels[0]), "%s is set %d chevron widths in", name, level)
		rest := labels[0]
		for rest.Parent() != nil && rest.Parent().Parent() != list {
			rest = rest.Parent()
		}
		if chevron, isButton := rest.Children()[0].Self.(*unison.Button); isButton {
			frame := chevron.RectTo(chevron.ContentRect(false), list)
			c.Equal(pickerDisclosureSize(), frame.Width, "the chevron of %s takes the room left for one", name)
			c.Equal(x(labels[0]), frame.Right()+unison.StdHSpacing, "the chevron of %s is just before its name", name)
		}
	}
}
