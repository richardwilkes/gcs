// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package calculators

import (
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
)

func TestTextLabelLinksReachScreenReadersAndTheKeyboard(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet := openNewCharacterSheet(t, screen)
	screen.Do(func() { Display(sheet) })
	calc := uxtest.SoleEditor[*Dockable](t, screen, func(d unison.Dockable) bool {
		_, isCalculator := d.AsPanel().Self.(*Dockable)
		return isCalculator
	})
	demolition := calc.demolition
	selectCalculatorTab(t, screen, calc, demolition)
	if screen.AccessibilityTree(wnd) == nil {
		t.Fatal("the window must be described")
	}
	current := func(l *ux.TextLabel) (index int) {
		screen.Do(func() { index = l.CurrentLink() })
		return index
	}
	focus := func() (focus *unison.Panel) {
		screen.Do(func() { focus = wnd.Focus() })
		return focus
	}

	var multi *ux.TextLabel
	var refs []string
	var followed []string
	record := func(ref string) { followed = append(followed, ref) }
	screen.Do(func() {
		for _, one := range uxtest.PanelsOfType[*ux.TextLabel](demolition.content) {
			if links := one.Links(); len(links) > 1 {
				multi = one
				for _, link := range links {
					refs = append(refs, link.Ref)
				}
				break
			}
		}
		if multi != nil {
			multi.LinkHandler = record
		}
	})
	if multi == nil {
		t.Fatal("the demolition calculator must have a note citing more than one page")
	}

	tree := screen.AccessibilityTree(wnd)
	node := screen.AccessibilityNodeFor(multi)
	if node == nil {
		t.Fatal("the note must be described")
	}
	c.Equal(role.Label, node.Role)
	c.Equal(len(refs), len(node.Children), "each link is a child of the note")
	for i, id := range node.Children {
		link := tree.Node(id)
		if link == nil {
			t.Fatalf("link %d of the note must be described", i)
		}
		c.Equal(role.Link, link.Role)
		c.Equal(refs[i], link.Name, "link %d is named for the page it cites", i)
		c.True(link.Actions.Has(accessibility.Press), "link %d offers the press that follows it", i)
		c.True(link.Bounds.Width > 0 && link.Bounds.Height > 0, "link %d lies somewhere: %v", i, link.Bounds)
	}
	c.True(screen.PerformAccessibilityAction(accessibility.ActionRequest{Node: node.Children[1], Action: accessibility.Press}),
		"a screen reader can follow a link")
	c.Equal([]string{refs[1]}, followed)
	followed = nil

	last := len(refs) - 1
	screen.Do(multi.RequestFocus)
	c.Equal(multi.AsPanel(), focus(), "a label holding links takes the focus")
	c.Equal(0, current(multi), "the focus arrives on the first link")
	c.Equal(node.Children[0], screen.AccessibilityTree(wnd).Focus, "the screen reader is told the link holds the focus")
	screen.KeyPress(unison.KeyRight, mod.None)
	c.Equal(1, current(multi), "the right arrow puts the keyboard on the next link")
	c.Equal(node.Children[1], screen.AccessibilityTree(wnd).Focus)
	screen.KeyPress(unison.KeyEnd, mod.None)
	c.Equal(last, current(multi), "End puts the keyboard on the last link")
	screen.KeyPress(unison.KeyRight, mod.None)
	c.Equal(last, current(multi), "the keyboard stays on the last link")
	screen.KeyPress(unison.KeyHome, mod.None)
	c.Equal(0, current(multi), "Home puts the keyboard on the first link")
	screen.KeyPress(unison.KeyLeft, mod.None)
	c.Equal(0, current(multi), "the keyboard stays on the first link, the text as a whole being no stop")
	screen.KeyPress(unison.KeyRight, mod.None)
	screen.KeyPress(unison.KeyReturn, mod.None)
	c.Equal([]string{refs[1]}, followed, "Enter follows the link the keyboard is on")
	screen.KeyPress(unison.KeySpace, mod.None)
	c.Equal([]string{refs[1], refs[1]}, followed, "as Space does")
	followed = nil

	screen.KeyPress(unison.KeyHome, mod.None)
	for i := 1; i <= last; i++ {
		screen.KeyPress(unison.KeyTab, mod.None)
		c.Equal(multi.AsPanel(), focus(), "Tab stays within the label while it has a link to go to")
		c.Equal(i, current(multi), "Tab goes to the next link")
	}
	screen.KeyPress(unison.KeyTab, mod.None)
	after := focus()
	c.NotEqual(multi.AsPanel(), after, "Tab from the last link leaves the label")
	screen.KeyPress(unison.KeyTab, mod.Shift)
	c.Equal(multi.AsPanel(), focus(), "Shift-Tab comes back to the label")
	c.Equal(last, current(multi), "on its last link")
	for i := last - 1; i >= 0; i-- {
		screen.KeyPress(unison.KeyTab, mod.Shift)
		c.Equal(multi.AsPanel(), focus())
		c.Equal(i, current(multi), "Shift-Tab goes to the link before")
	}
	screen.KeyPress(unison.KeyTab, mod.Shift)
	c.NotEqual(multi.AsPanel(), focus(), "Shift-Tab from the first link leaves the label")
	c.NotEqual(after, focus(), "by the other side")
	screen.KeyPress(unison.KeyTab, mod.None)
	c.Equal(multi.AsPanel(), focus())
	c.Equal(0, current(multi), "Tab comes back to the label on its first link")

	c.True(screen.PerformAccessibilityAction(accessibility.ActionRequest{Node: node.Children[last], Action: accessibility.Focus}),
		"a screen reader can put the focus on a link")
	c.Equal(last, current(multi))
	c.Equal(node.Children[last], screen.AccessibilityTree(wnd).Focus)

	setFocusForReading := uxtest.FocusForReadingSetter(t, screen, wnd)
	setFocusForReading(true)
	screen.KeyPress(unison.KeyTab, mod.None)
	c.NotEqual(multi.AsPanel(), focus())
	screen.KeyPress(unison.KeyTab, mod.Shift)
	c.Equal(multi.AsPanel(), focus())
	c.Equal(last, current(multi), "Shift-Tab still arrives on the last link")
	screen.KeyPress(unison.KeyHome, mod.None)
	screen.KeyPress(unison.KeyTab, mod.Shift)
	c.Equal(multi.AsPanel(), focus(), "the text as a whole is the stop ahead of the first link")
	c.Equal(-1, current(multi))
	c.Equal(node.ID, screen.AccessibilityTree(wnd).Focus, "the text as a whole is what holds the focus")
	screen.KeyPress(unison.KeyReturn, mod.None)
	c.Equal(0, len(followed), "with several links and the keyboard on none of them, Enter follows nothing")
	screen.KeyPress(unison.KeyTab, mod.None)
	c.Equal(0, current(multi), "Tab goes from the text to its first link")
	screen.KeyPress(unison.KeyLeft, mod.None)
	c.Equal(-1, current(multi), "the left arrow from the first link puts the keyboard back on the text")
	screen.Do(func() { wnd.SetFocus(nil) })
	screen.Do(multi.RequestFocus)
	c.Equal(-1, current(multi), "the focus arrives on the text as a whole")
	setFocusForReading(false)

	hiking := calc.hiking
	selectCalculatorTab(t, screen, calc, hiking)
	single := hiking.hikingRollPageLabel
	screen.Do(func() { single.LinkHandler = record })
	c.Equal(1, len(single.Links()), "the hiking roll's page label cites one page")
	screen.Do(single.RequestFocus)
	c.Equal(single.AsPanel(), focus(), "a label holding a link takes the focus")
	c.Equal(0, current(single), "with the keyboard on the link")
	screen.KeyPress(unison.KeySpace, mod.None)
	c.Equal([]string{single.Links()[0].Ref}, followed, "Space follows the link")
}

// Each of the header's links opens its page with the words it stands for highlighted.
func TestCalculatorHeaderIsOneHeadingHoldingItsLinks(t *testing.T) {
	c := check.New(t)
	type opened struct{ pageRef, highlight string }
	var followed []opened
	uxtest.SwapForTest(t, &headerPageRefOpener, func(pageRef, highlight string) {
		followed = append(followed, opened{pageRef: pageRef, highlight: highlight})
	})
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	calc := openCalculator(t, screen)
	explosion := calc.explosion
	selectCalculatorTab(t, screen, calc, explosion)
	var header *ux.TextLabel
	var texts []string
	screen.Do(func() {
		if children := explosion.content.Children(); len(children) > 0 {
			if label, isLabel := children[0].Self.(*ux.TextLabel); isLabel {
				header = label
			}
		}
		texts = uxtest.LabelTexts(explosion.content)
	})
	if header == nil {
		t.Fatal("the calculator must start with its header")
	}
	const title = "Explosions & Area Attacks (BX413, BX414)"
	c.Equal(title, header.String())
	for _, piece := range []string{"Explosions & Area Attacks (", ", ", ")"} {
		c.False(slices.Contains(texts, piece), "the header is no longer made of pieces, but %q is one", piece)
	}

	tree := screen.AccessibilityTree(wnd)
	if tree == nil {
		t.Fatal("the window must be described")
	}
	node := screen.AccessibilityNodeFor(header)
	if node == nil {
		t.Fatal("the header must be described")
	}
	c.Equal(role.Heading, node.Role)
	c.Equal(1, node.Level)
	c.Equal(title, node.Name)
	if len(node.Children) != 2 {
		t.Fatalf("the header must hold its two links, but holds %d children", len(node.Children))
	}
	for i, ref := range []string{"BX413", "BX414"} {
		link := tree.Node(node.Children[i])
		if link == nil {
			t.Fatalf("link %d of the header must be described", i)
		}
		c.Equal(role.Link, link.Role)
		c.Equal(ref, link.Name)
		c.True(link.Actions.Has(accessibility.Press), "%s offers the press that follows it", ref)
	}
	c.True(screen.PerformAccessibilityAction(accessibility.ActionRequest{Node: node.Children[1], Action: accessibility.Press}),
		"a screen reader can follow a link")
	c.Equal([]opened{{pageRef: "BX414", highlight: "Explosions"}}, followed)
	followed = nil

	screen.Do(header.RequestFocus)
	var focused bool
	screen.Do(func() { focused = wnd.Focus() == header.AsPanel() })
	c.True(focused, "the header holds links, so it takes the focus")
	c.Equal(node.Children[0], screen.AccessibilityTree(wnd).Focus, "the screen reader is told the link holds the focus")
	screen.KeyPress(unison.KeyReturn, mod.None)
	c.Equal([]opened{{pageRef: "BX413", highlight: "Area and Spreading Attacks"}}, followed)
	followed = nil
	screen.KeyPress(unison.KeyTab, mod.None)
	screen.Do(func() { focused = wnd.Focus() == header.AsPanel() })
	c.True(focused, "Tab goes on to the header's second link")
	c.Equal(node.Children[1], screen.AccessibilityTree(wnd).Focus)
	screen.KeyPress(unison.KeyReturn, mod.None)
	c.Equal([]opened{{pageRef: "BX414", highlight: "Explosions"}}, followed)
	screen.KeyPress(unison.KeyTab, mod.None)
	screen.Do(func() { focused = wnd.Focus() == header.AsPanel() })
	c.False(focused, "Tab from the last link leaves the header")

	// Every page a header cites is a link, whatever the book.
	for _, tab := range calc.tabs {
		var text string
		var refs []string
		screen.Do(func() {
			children := tab.panel().Children()
			if len(children) == 0 {
				return
			}
			if label, isLabel := children[0].Self.(*ux.TextLabel); isLabel {
				text = label.String()
				_, pref, _ := label.Sizes(geom.Size{})
				label.SetFrameRect(geom.NewRect(0, 0, pref.Width, pref.Height))
				for _, link := range label.Links() {
					refs = append(refs, link.Ref)
				}
			}
		})
		open := strings.LastIndex(text, "(")
		if open < 0 || !strings.HasSuffix(text, ")") {
			t.Fatalf("the header of %s must end with its page references, but is %q", tab.title(), text)
		}
		c.Equal(strings.Split(text[open+1:len(text)-1], ", "), refs, "each page %q cites is a link", text)
	}
}

func TestTextLabelLinksAreTabStopsWithoutAScreenReader(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	// Focus for reading is on, but must change nothing while no screen reader is listening.
	gs := gurps.GlobalSettings().General
	uxtest.SwapForTest(t, &gs.FocusForReading, true)
	saved := unison.FocusForReading()
	t.Cleanup(func() { screen.Do(func() { unison.SetFocusForReading(saved) }) })
	screen.Do(gs.UpdateFocusForReading)
	calc := openCalculator(t, screen)
	explosion := calc.explosion
	selectCalculatorTab(t, screen, calc, explosion)
	var active bool
	screen.Do(func() { active = unison.IsAccessibilityActive() })
	if active {
		t.Fatal("no screen reader may be listening")
	}
	var header *ux.TextLabel
	var plain []*ux.TextLabel
	var stops []*unison.Panel
	var followed []string
	screen.Do(func() {
		for _, one := range uxtest.PanelsOfType[*ux.TextLabel](explosion.content) {
			if one.String() != "" && len(one.Links()) == 0 {
				plain = append(plain, one)
			}
		}
		if children := explosion.content.Children(); len(children) > 0 {
			if label, isLabel := children[0].Self.(*ux.TextLabel); isLabel {
				header = label
				header.LinkHandler = func(ref string) { followed = append(followed, ref) }
			}
		}
		stops = uxtest.PanelsMatching(explosion.content, (*unison.Panel).Focusable)
	})
	if header == nil {
		t.Fatal("the calculator must start with its header")
	}
	c.Equal("Explosions & Area Attacks (BX413, BX414)", header.String())
	c.True(len(plain) > 0, "the calculator must have text citing no page")
	for _, one := range plain {
		var focusable bool
		screen.Do(func() { focusable = one.Focusable() })
		c.False(focusable, "%q holds no link, so it is no tab stop", one.String())
	}
	for _, one := range stops {
		if label, isLabel := one.Self.(*unison.Label); isLabel {
			t.Errorf("%q is static text, so it may be no tab stop", label.String())
		}
	}
	if len(stops) < 2 || stops[0] != header.AsPanel() {
		t.Fatal("the header must be the first of the calculator's tab stops")
	}
	focus := func() (focus *unison.Panel) {
		screen.Do(func() { focus = wnd.Focus() })
		return focus
	}
	current := func() (index int) {
		screen.Do(func() { index = header.CurrentLink() })
		return index
	}

	screen.Do(stops[1].RequestFocus)
	c.Equal(stops[1], focus())
	screen.KeyPress(unison.KeyTab, mod.Shift)
	c.Equal(header.AsPanel(), focus(), "Shift-Tab reaches the header")
	c.Equal(1, current(), "on its last link")
	screen.KeyPress(unison.KeyReturn, mod.None)
	c.Equal([]string{"BX414"}, followed, "Enter follows the link the keyboard is on")
	screen.KeyPress(unison.KeyTab, mod.Shift)
	c.Equal(header.AsPanel(), focus())
	c.Equal(0, current(), "Shift-Tab goes to the link before")
	screen.KeyPress(unison.KeySpace, mod.None)
	c.Equal([]string{"BX414", "BX413"}, followed, "Space follows the link the keyboard is on")
	screen.KeyPress(unison.KeyLeft, mod.None)
	c.Equal(0, current(), "the text as a whole is no stop")
	screen.KeyPress(unison.KeyTab, mod.None)
	c.Equal(header.AsPanel(), focus())
	c.Equal(1, current(), "Tab goes to the next link")
	screen.KeyPress(unison.KeyTab, mod.None)
	c.Equal(stops[1], focus(), "Tab from the last link goes on to the control after the header")

	// The Target subheader between these two is static text, so Tab passes over it.
	screen.Do(explosion.hotFragmentsBox.RequestFocus)
	screen.KeyPress(unison.KeyTab, mod.None)
	c.Equal(explosion.target.popup.AsPanel(), focus(), "Tab goes from Hot fragments to the target's Source")
	screen.Do(func() { active = unison.IsAccessibilityActive() })
	c.False(active, "nothing here may have started a screen reader listening")
}
