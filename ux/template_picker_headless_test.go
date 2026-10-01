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

	"github.com/richardwilkes/gcs/v5/model/criteria"
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

// The organizing groups kept with what was picked from them reach the table, where one goes only once all of that
// merged into rows already there.
func TestPickerGroupsReachTheTable(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	brawling, karate := newTestSkill("Brawling", fxp.Four, nil), newTestSkill("Karate", fxp.Four, nil)
	group := func(name string, children ...*gurps.Skill) *gurps.Skill {
		one := gurps.NewSkill(nil, nil, true)
		one.Name = name
		one.PickSeparately = true
		one.Children = children
		SetParents(children, one)
		return one
	}
	unarmed := group("Unarmed", newTestSkill("Brawling", fxp.Two, nil), newTestSkill("Judo", fxp.Two, nil))
	striking := group("Striking", newTestSkill("Karate", fxp.Two, nil))
	choice := gurps.NewSkillChoiceContainer(nil, nil)
	choice.TemplatePicker.Qualifier.Compare = criteria.EqualsNumber
	choice.TemplatePicker.Qualifier.Qualifier = fxp.Three
	choice.Children = []*gurps.Skill{unarmed, striking}
	SetParents(choice.Children, choice)
	var sheet *Sheet
	screen.Do(func() {
		sheet = newTestSheetForTemplate(t)
		sheet.Entity().Skills = []*gurps.Skill{brawling, karate}
		sheet.Rebuild(true)
	})
	part := &applyPart[*gurps.Skill]{table: sheet.Skills.Table, rows: []*gurps.Skill{choice}, index: -1}
	resolved := false
	c.True(screen.Post(func() { resolved = part.resolvePickers(promptOperation{}, false) }))
	screen.Sync()
	dialogWnd, dialog := modalDialog(t, screen, wnd)
	var boxes []*unison.CheckBox
	screen.Do(func() { boxes = panelsOfType[*unison.CheckBox](dialogWnd.Content()) })
	for _, box := range boxes {
		screen.Click(screen.PanelCenter(box))
	}
	screen.Click(screen.PanelCenter(dialogButton(t, screen, dialog, unison.ModalResponseOK)))
	c.True(resolved, "every option is picked")
	c.Equal([]*gurps.Skill{unarmed, striking}, part.groups, "the groups kept are handed on")
	screen.Do(func() { part.place(true) })
	c.Equal(fxp.FromInteger(6), brawling.Points)
	c.Equal(fxp.FromInteger(6), karate.Points)
	c.Equal([]*gurps.Skill{brawling, karate, unarmed}, sheet.Entity().Skills,
		"a group with a pick left is kept, one with none left goes")
	c.Equal("Judo", unarmed.Children[0].Name)
}

// The picker dialog is sized with its rows' text in place, wraps its hint to the list's width, and widens when the text
// does.
func TestPickerDialogFitsItsContent(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	s, n := newKnightSession()
	// Names long enough that the rows, rather than the buttons, set the dialog's width.
	n["order"].Name = "Knightly Order of the Realm"
	n["lion"].Name = "Order of the Lion, Sworn to the Crown and to the Defense of the Realm Against All Its Foes"
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
		width := wnd.ContentRect().Width
		choose(s, n, "lion", "honors", "cr", "wm", "fear2")
		refresh()
		_, pref, _ := wnd.Content().Sizes(geom.Size{})
		c.True(pref.Width <= wnd.ContentRect().Width, "the dialog widens to fit longer text")
		c.True(wnd.ContentRect().Width > width, "the dialog widens")
	})
}

// Text a row gains once the dialog is up, as what was picked from a choice, is shown whole.
func TestPickerRowTextIsNeverCut(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	s, n := newKnightSession()
	n["fit2"].Name = "Very Fit"
	screen.Do(func() {
		dialog, refresh := s.newPickerDialog(n["root"], 0)
		c.NotNil(dialog, "the dialog must be made")
		if dialog == nil {
			return
		}
		wnd := dialog.Window()
		defer wnd.Dispose()
		wnd.ValidateLayout()
		choose(s, n, "ea", "ep", "fit", "fit2", "luck")
		s.pickerAnswered[n["fit"]] = true
		refresh()
		wnd.ValidateLayout()
		for _, label := range panelsOfType[*unison.Label](wnd.Content()) {
			_, pref, _ := label.Sizes(geom.Size{})
			c.True(label.FrameRect().Width >= pref.Width, "%q is cut: %v of %v", label.String(), label.FrameRect().Width,
				pref.Width)
		}
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

// The line under the picker's list is a notice for an error or a warning, in a box of the theme's color for it with
// the text in the color drawn on that, while it is plain text otherwise: dimmed while picks are still open, and green
// once all is well.
func TestPickerDialogHintColor(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	for _, tc := range []struct {
		row   string
		picks []string
		state pickerState
		ink   unison.Ink
		boxed bool
	}{
		{"root", []string{"ea", "ep", "fit", "order", "lion", "honors", "cr", "wm", "fear2"}, pickerError, unison.ThemeOnError, true},
		{"order", []string{"order", "lion", "honors", "cr", "wm", "fear2"}, pickerWarning, unison.ThemeOnWarning, true},
		{"root", []string{"ea", "ep", "fit", "order"}, pickerOpen, dimmedTextColor, false},
		{"fit", []string{"fit", "fit1"}, pickerOK, unison.Green, false},
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
			c.Equal(tc.boxed, hints[0].Border() != nil, "only a warning or an error is set out in a box: %s", tc.row)
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
		checkPickerRowPlaces(c, list, map[string]pickerPlace{
			"ea": {}, "ep": {}, "fit": {}, "order": {}, "luck": {}, "shield": {},
		})
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
	// Picks that meet the rule enable OK, so Return would accept the dialog wherever it isn't used otherwise.
	choose(s, n, "fear", "status")
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
		// Only the rows at the top and within social hold groups, so only they leave room for a chevron. Each group's
		// rows are set in one chevron width past its name.
		checkPickerRowPlaces(c, list, map[string]pickerPlace{
			"martial": {slot: true}, "fear": {under: "martial"}, "honors": {under: "martial"},
			"social": {slot: true}, "rank": {under: "social", slot: true}, "inner": {under: "social", slot: true},
			"status": {under: "inner"},
			"luck":   {slot: true},
		})
	})
	c.Equal(3, len(chevrons), "each organizing group is a header")
	c.Equal(5, len(boxes), "fear, honors, rank, status and luck are options")
	screen.Do(func() { c.Equal(boxes[0].AsPanel(), dialogWnd.CurrentFocus(), "the focus starts on the first option") })
	expanded := func(name string) (expandable, expanded bool) {
		screen.AccessibilityTree(dialogWnd)
		node := screen.AccessibilityNodeFor(chevrons[name])
		if node == nil {
			t.Fatalf("the chevron of %s must be described", name)
		}
		c.Equal(role.DisclosureTriangle, node.Role)
		c.Equal("Show or hide "+name, node.Name)
		c.Equal(node.Expanded, node.Pressed, "a chevron is pressed while open")
		c.True(node.Actions.Has(accessibility.Expand) && node.Actions.Has(accessibility.Collapse))
		return node.Expandable, node.Expanded
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
	screen.KeyPress(unison.KeySpace, mod.None)
	_, open = expanded("martial")
	c.True(open, "Space expands it again")
	c.False(done, "without closing the dialog")
	screen.Do(func() { c.Equal(cells, list.Children(), "every row is back in its place") })
	perform := func(name string, action accessibility.Action) {
		screen.AccessibilityTree(dialogWnd)
		c.True(screen.PerformAccessibilityAction(accessibility.ActionRequest{
			Node:   screen.AccessibilityNodeFor(chevrons[name]).ID,
			Action: action,
		}))
	}
	perform("martial", accessibility.Collapse)
	_, open = expanded("martial")
	c.False(open, "a screen reader can collapse it")
	perform("martial", accessibility.Collapse)
	_, open = expanded("martial")
	c.False(open, "and collapsing it again changes nothing")
	perform("martial", accessibility.Expand)
	_, open = expanded("martial")
	c.True(open, "and expand it")

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

	response := unison.ModalResponseCancel
	c.True(screen.Post(func() { response = s.showPicker(n["root"], 0) }))
	screen.Sync()
	modalDialog(t, screen, wnd)
	screen.KeyPress(unison.KeyReturn, mod.None)
	c.Equal(unison.ModalResponseOK, response, "Return first thing accepts the dialog")

	// Return on a chevron is the dialog's too, rather than the chevron's.
	response = unison.ModalResponseCancel
	c.True(screen.Post(func() { response = s.showPicker(n["root"], 0) }))
	screen.Sync()
	dialogWnd, _ = modalDialog(t, screen, wnd)
	var chevron *unison.Button
	screen.Do(func() {
		for _, p := range panelsMatching(dialogWnd.Content(), func(p *unison.Panel) bool {
			label, ok := p.Self.(*unison.Label)
			return ok && label.String() == "martial"
		}) {
			if one, ok := p.Parent().Children()[0].Self.(*unison.Button); ok {
				chevron = one
			}
		}
	})
	if chevron == nil {
		t.Fatal("martial must have a chevron")
	}
	screen.Do(chevron.RequestFocus)
	// The chevron is turned down while its group is open.
	isOpen := func() (open bool) {
		screen.Do(func() {
			if svg, ok := chevron.Drawable.(*unison.DrawableSVG); ok {
				open = svg.RotationDegrees == 90
			}
		})
		return open
	}
	c.True(isOpen(), "martial starts open")
	screen.KeyPress(unison.KeyReturn, mod.None)
	c.Equal(unison.ModalResponseOK, response, "Return on a chevron accepts the dialog")
	c.True(isOpen(), "without closing the group")
}

// An organizing group with nothing in it has no chevron, but its name keeps the place of one.
func TestPickerEmptyGroupHasNoChevron(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	choice := gurps.NewTrait(nil, nil, true)
	choice.TemplatePicker.Type = picker.Count
	choice.TemplatePicker.Qualifier.Qualifier = fxp.One
	empty := gurps.NewTrait(nil, choice, true)
	empty.Name = "empty"
	empty.PickSeparately = true
	luck := gurps.NewTrait(nil, choice, false)
	luck.Name = "luck"
	choice.Children = []*gurps.Trait{empty, luck}
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
		list := panelsOfType[*unison.CheckBox](wnd.Content())[0].Parent().Parent()
		c.Equal(0, len(panelsOfType[*unison.Button](list)), "there is nothing to show or hide")
		checkPickerRowPlaces(c, list, map[string]pickerPlace{"empty": {slot: true}, "luck": {slot: true}})
	})
}

// Closing a group by mouse while the focus is inside it moves the focus to its chevron, so the keyboard still works.
func TestPickerClosingGroupKeepsFocus(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	t.Cleanup(func() { unison.SetAccessibilityEnabled(true) })
	screen.Do(func() { unison.SetAccessibilityEnabled(false) })
	s, n := newOrganizedSession()
	done := false
	c.True(screen.Post(func() {
		s.showPicker(n["root"], 0)
		done = true
	}))
	screen.Sync()
	dialogWnd, _ := modalDialog(t, screen, wnd)
	var chevron *unison.Button
	screen.Do(func() {
		c.Equal("fear", labelTexts(dialogWnd.CurrentFocus().Parent())[0], "the focus starts inside martial")
		for _, p := range panelsMatching(dialogWnd.Content(), func(p *unison.Panel) bool {
			label, ok := p.Self.(*unison.Label)
			return ok && label.String() == "martial"
		}) {
			if one, ok := p.Parent().Children()[0].Self.(*unison.Button); ok {
				chevron = one
			}
		}
	})
	if chevron == nil {
		t.Fatal("martial must have a chevron")
	}
	screen.Click(screen.PanelCenter(chevron))
	screen.Do(func() { c.Equal(chevron.AsPanel(), dialogWnd.CurrentFocus(), "the focus moves to the chevron") })
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.True(done, "Escape still closes the dialog")
}

// A container picked as a unit keeps its checkbox and cost, and its chevron shows what it holds as rows with nothing to
// pick, a choice among them by its rule, widening the dialog to fit them, their page references lined up with those of
// the options.
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
	trinket := trait("Trinket", 0, trait("Ring", 1), trait("Amulet", 2))
	trinket.TemplatePicker.Type = picker.Count
	trinket.TemplatePicker.Qualifier.Compare = criteria.EqualsNumber
	trinket.TemplatePicker.Qualifier.Qualifier = fxp.One
	kit := trait("Kit", 0, trait(long, 5), trait("Pack", 0, trait("Rope", 1)), trinket)
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
		checkPickerRowsAligned(c, row.Parent(), 6)
		// What the unit holds is set in one level past its name.
		checkPickerRowPlaces(c, row.Parent(), map[string]pickerPlace{
			"Kit": {slot: true}, "Luck": {slot: true}, long: {under: "Kit"}, "Pack": {under: "Kit"}, "Rope": {under: "Pack"},
			"Trinket (pick 1)": {under: "Kit"},
		})
		texts := labelTexts(wnd.Content())
		c.True(slices.Contains(texts, long) && slices.Contains(texts, "Pack") && slices.Contains(texts, "Rope"),
			"opening it shows what it holds, however deep")
		c.False(slices.Contains(texts, "Ring"), "but a choice within it shows only its rule, not its options")
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

// The list has room for ten rows however few it holds, so opening a group in a short list shows all that it holds
// without the dialog growing, and the dialog can't be made smaller than that.
func TestPickerListHoldsTenRows(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	choice := gurps.NewTrait(nil, nil, true)
	choice.Name = "Choice"
	choice.TemplatePicker.Type = picker.Count
	choice.TemplatePicker.Qualifier.Qualifier = fxp.One
	kit := gurps.NewTrait(nil, choice, true)
	kit.Name = "Kit"
	for i := range 8 {
		item := gurps.NewTrait(nil, kit, false)
		item.Name = fmt.Sprintf("Item %d", i)
		kit.Children = append(kit.Children, item)
	}
	luck := gurps.NewTrait(nil, choice, false)
	luck.Name = "Luck"
	choice.Children = []*gurps.Trait{kit, luck}
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
		list := panelsOfType[*unison.CheckBox](wnd.Content())[0].Parent().Parent()
		scroll := panelsOfType[*unison.ScrollPanel](wnd.Content())[0]
		chevrons := panelsOfType[*unison.Button](list)
		c.Equal(1, len(chevrons), "the unit has a chevron")
		if len(chevrons) != 1 {
			return
		}
		c.False(slices.Contains(labelTexts(list), "Item 7"), "what it holds starts hidden")
		minimum := pickerListMinSize()
		view := scroll.ContentView().ContentRect(false).Size
		c.True(view.Height >= minimum.Height && view.Width >= minimum.Width,
			"the list has its least room: wants %v, has %v", minimum, view)
		size := wnd.ContentRect().Size
		chevrons[0].Click()
		wnd.ValidateLayout()
		c.True(slices.Contains(labelTexts(list), "Item 7"), "opening it shows what it holds")
		c.Equal(size, wnd.ContentRect().Size, "the dialog keeps its size")
		c.True(list.FrameRect().Height <= scroll.ContentView().ContentRect(false).Height,
			"the list shows every row without scrolling")
		// Asked to be smaller, the dialog keeps the list's room.
		r := wnd.ContentRect()
		r.Width /= 2
		r.Height /= 2
		wnd.SetContentRect(r)
		c.Equal(size, wnd.ContentRect().Size, "the dialog is no smaller than its content's least size")
	})
}

// A chevron in the picker, on a header or a unit's row, gets all the room it asks for, inside its row, with room around
// its icon for the outline it draws when it has the focus, so neither is cut off.
func TestPickerChevronsAreNotClipped(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	choice := gurps.NewTrait(nil, nil, true)
	choice.Name = "Choice"
	choice.TemplatePicker.Type = picker.Count
	choice.TemplatePicker.Qualifier.Qualifier = fxp.One
	group := gurps.NewTrait(nil, choice, true)
	group.Name = "Group"
	group.PickSeparately = true
	fear := gurps.NewTrait(nil, group, false)
	fear.Name = "Fear"
	group.Children = []*gurps.Trait{fear}
	kit := gurps.NewTrait(nil, choice, true)
	kit.Name = "Kit"
	rope := gurps.NewTrait(nil, kit, false)
	rope.Name = "Rope"
	kit.Children = []*gurps.Trait{rope}
	choice.Children = []*gurps.Trait{group, kit}
	s := newPickerSession(promptOperation{}, []*gurps.Trait{choice}, false)
	done := false
	c.True(screen.Post(func() {
		s.showPicker(choice, 0)
		done = true
	}))
	screen.Sync()
	dialogWnd, _ := modalDialog(t, screen, wnd)
	var chevrons []*unison.Button
	screen.Do(func() {
		chevrons = panelsOfType[*unison.Button](panelsOfType[*unison.CheckBox](dialogWnd.Content())[0].Parent().Parent())
	})
	c.Equal(2, len(chevrons), "the group and the unit each have a chevron")
	checkRoom := func(focused bool) {
		for _, chevron := range chevrons {
			if focused {
				screen.Do(chevron.RequestFocus)
			}
			screen.Do(func() {
				dialogWnd.ValidateLayout()
				c.Equal(focused, chevron.Focused())
				_, pref, _ := chevron.Sizes(geom.Size{})
				frame := chevron.FrameRect()
				c.True(frame.Width >= pref.Width && frame.Height >= pref.Height,
					"a chevron gets the room it asks for: %v of %v", frame.Size, pref)
				drawable := chevron.Drawable.LogicalSize()
				// Focused, it draws a 2 point outline just inside its edges, so its icon needs that much room around it.
				c.True(frame.Width-drawable.Width >= 4 && frame.Height-drawable.Height >= 4,
					"a chevron leaves room around its icon for its focus outline: %v in %v", drawable, frame.Size)
				for p := chevron.AsPanel(); p.Parent() != nil && p.Parent() != dialogWnd.Content(); p = p.Parent() {
					within := p.Parent().ContentRect(true)
					rect := p.FrameRect()
					c.True(rect.X >= within.X && rect.Y >= within.Y && rect.Right() <= within.Right() &&
						rect.Bottom() <= within.Bottom(), "a chevron is not cut off by what holds it: %v in %v", rect, within)
				}
			})
		}
	}
	checkRoom(false)
	checkRoom(true)
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.True(done, "Escape closes the dialog")
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

// pickerPlace tells where a row of the picker's list is set: under the row whose name it gives, or the top of the list
// if none, and whether the row's level leaves room for a chevron before its name.
type pickerPlace struct {
	under string
	slot  bool
}

// checkPickerRowPlaces checks that the rows of the picker's list lead with their checkboxes, or room for one, all in one
// column, and that the name of each row given is set one chevron width past the name of the row it is under, or at the
// start of the rows if it is under none, and one more past that when its level leaves room for a chevron, just after
// its chevron if it has one.
func checkPickerRowPlaces(c check.Checker, list *unison.Panel, places map[string]pickerPlace) {
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
	step := pickerDisclosureSize().Width + unison.StdHSpacing
	label := func(name string) *unison.Panel {
		labels := panelsMatching(list, func(p *unison.Panel) bool {
			one, isLabel := p.Self.(*unison.Label)
			return isLabel && one.String() == name
		})
		c.Equal(1, len(labels), "%s is shown once", name)
		if len(labels) != 1 {
			return nil
		}
		return labels[0]
	}
	for name, place := range places {
		one := label(name)
		if one == nil {
			continue
		}
		want := left + pickerCheckBoxSize().Width + unison.StdHSpacing
		if place.under != "" {
			parent := label(place.under)
			if parent == nil {
				continue
			}
			want = x(parent) + step
		}
		if place.slot {
			want += step
		}
		c.Equal(want, x(one), "%s is set in from %q as %+v", name, place.under, place)
		rest := one
		for rest.Parent() != nil && rest.Parent().Parent() != list {
			rest = rest.Parent()
		}
		if chevron, isButton := rest.Children()[0].Self.(*unison.Button); isButton {
			frame := chevron.RectTo(chevron.ContentRect(false), list)
			c.Equal(pickerDisclosureSize().Width, frame.Width, "the chevron of %s takes the room left for one", name)
			c.Equal(x(one), frame.Right()+unison.StdHSpacing, "the chevron of %s is just before its name", name)
		}
	}
}
