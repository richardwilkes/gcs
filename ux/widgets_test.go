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
	"bytes"
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

// renderTooltipMarkdown renders the text with the same Markdown configuration unison uses for tooltips (goldmark with
// the GFM extension), so the tests exercise the real rendering behavior rather than an approximation.
func renderTooltipMarkdown(c check.Checker, text string) string {
	var buffer bytes.Buffer
	c.NoError(goldmark.New(goldmark.WithExtensions(extension.GFM)).Convert([]byte(text), &buffer))
	return buffer.String()
}

// TestMarkdownHardLineBreaks verifies that markdownHardLineBreaks turns each newline into a hard line break, so that
// newline-separated lines of a tooltip, such as the bonuses of the conditional modifiers amount column, render one per
// line rather than collapsing onto one as soft breaks would, while a blank line still renders as a paragraph break with
// no hard line break, leaving tooltips that already rely on Markdown undisturbed.
func TestMarkdownHardLineBreaks(t *testing.T) {
	c := check.New(t)
	for _, d := range []struct {
		name       string
		in         string
		want       string
		paragraphs int // How many <p> elements the conversion renders
		breaks     int // How many <br> elements the conversion renders
	}{
		{name: "empty", in: "", want: ""},
		{name: "single line", in: "+2 Foo", want: "+2 Foo", paragraphs: 1},
		{name: "two lines", in: "+2 Foo\n+3 Bar", want: "+2 Foo  \n+3 Bar", paragraphs: 1, breaks: 1},
		{name: "three lines", in: "+2 Foo\n+3 Bar\n+1 Baz", want: "+2 Foo  \n+3 Bar  \n+1 Baz", paragraphs: 1, breaks: 2},
		{name: "paragraph break", in: "a\n\nb", want: "a  \n  \nb", paragraphs: 2},
	} {
		converted := markdownHardLineBreaks(d.in)
		c.Equal(d.want, converted, d.name)
		rendered := renderTooltipMarkdown(c, converted)
		c.Equal(d.paragraphs, strings.Count(rendered, "<p>"), d.name)
		c.Equal(d.breaks, strings.Count(rendered, "<br>"), d.name)
	}
}

// TestNewApplyCancelButtons verifies the behavior every editor with pending changes relies on: both buttons start out
// disabled, apply closes the editor only when it reports success, cancel closes it without applying, and the keyboard
// shortcuts are shown only when asked for.
func TestNewApplyCancelButtons(t *testing.T) {
	c := check.New(t)
	toolbar := unison.NewPanel()
	applied, closed := 0, 0
	succeed := true
	applyButton, cancelButton := newApplyCancelButtons(toolbar, false,
		func() bool {
			applied++
			return succeed
		},
		func() { closed++ })
	c.Equal([]*unison.Panel{applyButton.AsPanel(), cancelButton.AsPanel()}, toolbar.Children(),
		"apply then cancel are added to the toolbar")
	c.False(applyButton.Enabled(), "apply starts out disabled")
	c.False(cancelButton.Enabled(), "cancel starts out disabled")

	applyButton.ClickCallback()
	c.Equal(1, applied)
	c.Equal(1, closed, "a successful apply closes the editor")

	succeed = false
	applyButton.ClickCallback()
	c.Equal(2, applied)
	c.Equal(1, closed, "an apply that fails leaves the editor open")

	cancelButton.ClickCallback()
	c.Equal(2, applied, "cancel applies nothing")
	c.Equal(2, closed, "and closes the editor")

	c.Equal(1, len(applyButton.Tooltip.Children()), "without shortcuts, the tooltip is a single line")
	applyButton, cancelButton = newApplyCancelButtons(unison.NewPanel(), true, func() bool { return true }, func() {})
	c.Equal(2, len(applyButton.Tooltip.Children()), "with shortcuts, the tooltip gains a second line")
	c.Equal(2, len(cancelButton.Tooltip.Children()))
}

// TestNewPopupMenuOffersItemsAndReportsChoices verifies that newPopupMenu offers the items in order, starts out showing
// the current value, or nothing when the current value isn't among the items, without reporting it as a choice, and
// hands each later selection to onSelect. Every settings popup is built through this helper, so a callback that fired
// during construction would clobber settings on open.
func TestNewPopupMenuOffersItemsAndReportsChoices(t *testing.T) {
	c := check.New(t)
	var chosen []string
	popup := newPopupMenu([]string{"one", "two", "three"}, "two", func(item string) { chosen = append(chosen, item) })
	c.Equal(3, popup.ItemCount(), "all items should be offered")
	for i, want := range []string{"one", "two", "three"} {
		item, ok := popup.ItemAt(i)
		c.True(ok, "item %d should exist", i)
		c.Equal(want, item, "items should keep their order")
	}
	selected, ok := popup.Selected()
	c.True(ok, "the current value should be selected")
	c.Equal("two", selected, "the current value should be selected")
	c.Nil(chosen, "building the popup must not report the initial selection")

	popup.Select("three")
	c.Equal([]string{"three"}, chosen, "selecting an item should report it")
	popup.Select("three")
	c.Equal([]string{"three"}, chosen, "re-selecting the same item must not report it again")
	popup.Select("one")
	c.Equal([]string{"three", "one"}, chosen, "each new selection should be reported")

	calls := 0
	unknown := newPopupMenu([]int{1, 2, 3}, 42, func(int) { calls++ })
	_, ok = unknown.Selected()
	c.False(ok, "an unknown current value should leave nothing selected")
	c.Equal(0, calls, "an unknown current value must not be reported")
	unknown.SelectIndex(1)
	c.Equal(1, calls, "choosing an item should be reported")
}

// TestNewCriteriaPanel verifies the container every criteria editor shares: a two-column panel added to the parent,
// spanning the requested columns and growing to fill them, behind an empty filler only when asked for one.
func TestNewCriteriaPanel(t *testing.T) {
	c := check.New(t)
	parent := unison.NewPanel()
	panel := newCriteriaPanel(parent, 3, true)
	c.Equal(2, len(parent.Children()), "a filler and the panel are added")
	c.Equal(0, len(parent.Children()[0].Children()), "the filler is empty")
	c.Equal(panel, parent.Children()[1], "the panel follows the filler")
	layout, ok := panel.Layout().(*unison.FlexLayout)
	c.True(ok, "the panel uses a flex layout")
	c.Equal(2, layout.Columns, "one column for the popup and one for the field")
	data, ok := panel.LayoutData().(*unison.FlexLayoutData)
	c.True(ok, "the panel has flex layout data")
	c.Equal(3, data.HSpan, "the panel spans the requested columns")
	c.True(data.HGrab, "the panel grows to fill them")

	parent = unison.NewPanel()
	panel = newCriteriaPanel(parent, 1, false)
	c.Equal(1, len(parent.Children()), "no filler is added unless asked for")
	c.Equal(panel, parent.Children()[0])
}

// TestNewComparisonPopup verifies that newComparisonPopup offers the choices in order with the requested one selected
// and leaves the selection callback for the caller to install, so that installing it cannot report the initial
// selection as a change.
func TestNewComparisonPopup(t *testing.T) {
	c := check.New(t)
	choices := []string{"is anything", "is", "starts with"}
	popup := newComparisonPopup("Test Comparison", choices, 2)
	c.Equal(len(choices), popup.ItemCount(), "all choices should be offered")
	for i, want := range choices {
		item, ok := popup.ItemAt(i)
		c.True(ok, "choice %d should exist", i)
		c.Equal(want, item, "choices should keep their order")
	}
	c.Equal(2, popup.SelectedIndex(), "the requested choice should be selected")
	c.Nil(popup.SelectionChangedCallback, "no selection callback should be installed")
}

// lastLabeledPair checks that the last two children of parent are a tooltip-less label showing labelText and then
// whatever was added after it, and returns the latter.
func lastLabeledPair(t *testing.T, parent *unison.Panel, labelText string) *unison.Panel {
	t.Helper()
	c := check.New(t)
	children := parent.Children()
	c.True(len(children) >= 2, "a label and a field must have been added")
	label, ok := children[len(children)-2].Self.(*unison.Label)
	c.True(ok, "the child before the field must be its label")
	c.Equal(labelText, label.String())
	c.Nil(label.Tooltip, "the label carries no tooltip of its own")
	return children[len(children)-1]
}

// TestAddLabelAndTargetedStringField verifies that the helper adds the label and then the field, registers the field
// with the target manager under its key, widens it to the prototype, installs the tooltip and edits the value the
// accessors reach.
func TestAddLabelAndTargetedStringField(t *testing.T) {
	c := check.New(t)
	parent := unison.NewPanel()
	mgr := NewTargetMgr(parent)
	value := "start"
	get := func() string { return value }
	set := func(s string) { value = s }
	field := addLabelAndTargetedStringField(parent, mgr, "k:name", "Name", "The name", prototypeMinNameWidth, get, set)
	c.True(lastLabeledPair(t, parent, "Name").Is(field))
	c.True(mgr.Find("k:name").Is(field), "the field is reachable through the target manager")
	c.Equal("The name", tooltipText(field.Tooltip))
	c.Equal("start", field.Text())
	unsized := NewStringField(nil, "", "Name", get, func(string) {})
	c.True(field.MinimumTextWidth > unsized.MinimumTextWidth, "the prototype widens the field")
	field.SetText("changed")
	c.Equal("changed", value)

	multi := addLabelAndTargetedMultiLineStringField(parent, mgr, "k:desc", "Description", "", "", get, set)
	c.True(lastLabeledPair(t, parent, "Description").Is(multi))
	c.True(mgr.Find("k:desc").Is(multi))
	c.Nil(multi.Tooltip, "an empty tooltip installs none")
	c.Equal(unsized.MinimumTextWidth, multi.MinimumTextWidth, "an empty prototype leaves the width alone")
	multi.SetText("two\nlines")
	c.Equal("two\nlines", value)
}

// TestAddLabelAndTargetedIntegerField verifies that the integer helper honors forceSign and installs its tooltip as
// the base tooltip, so that a starting value outside the field's range keeps showing the explanation of what is wrong
// with it until it is fixed, at which point the tooltip the caller asked for takes over.
func TestAddLabelAndTargetedIntegerField(t *testing.T) {
	c := check.New(t)
	parent := unison.NewPanel()
	mgr := NewTargetMgr(parent)
	value := 9
	field := addLabelAndTargetedIntegerField(parent, mgr, "k:depth", "Depth", "How deep",
		func() int { return value }, func(v int) { value = v }, 0, 5, true)
	c.True(lastLabeledPair(t, parent, "Depth").Is(field))
	c.True(mgr.Find("k:depth").Is(field), "the field is reachable through the target manager")
	c.Equal("+9", field.Text(), "forceSign is passed through")
	c.Equal("Value must be no more than +5", tooltipText(field.Tooltip),
		"the validation message must not be displaced by the caller's tooltip")
	field.SetText("3")
	c.Equal(3, value)
	c.Equal("How deep", tooltipText(field.Tooltip), "the caller's tooltip takes over once the value is valid")
}

// TestAddLabelAndTargetedPopup verifies that the popup helper adds the label and then the popup, registers it with
// the target manager, offers the items with the current value selected, installs the tooltip and edits the value the
// accessors reach.
func TestAddLabelAndTargetedPopup(t *testing.T) {
	c := check.New(t)
	parent := unison.NewPanel()
	mgr := NewTargetMgr(parent)
	value := "two"
	get := func() string { return value }
	set := func(s string) { value = s }
	popup := addLabelAndTargetedPopup(parent, mgr, "k:choice", "Choice", "Pick one", get, set, "one", "two", "three")
	c.True(lastLabeledPair(t, parent, "Choice").Is(popup))
	c.True(mgr.Find("k:choice").Is(popup), "the popup is reachable through the target manager")
	c.Equal("Pick one", tooltipText(popup.Tooltip))
	c.Equal(3, popup.ItemCount())
	selected, ok := popup.Selected()
	c.True(ok)
	c.Equal("two", selected, "the current value is selected")
	popup.Select("three")
	c.Equal("three", value)

	bare := addLabelAndTargetedPopup(parent, mgr, "k:bare", "Bare", "", get, set, "three")
	c.True(lastLabeledPair(t, parent, "Bare").Is(bare))
	c.Nil(bare.Tooltip, "an empty tooltip installs none")
}

// TestAddLabelAndScriptField verifies that the script helper adds the label and then the wrapper that holds the field
// and its guide buttons, registers the field with the target manager, gives it the tooltip and edits the value the
// accessors reach.
func TestAddLabelAndScriptField(t *testing.T) {
	c := check.New(t)
	parent := unison.NewPanel()
	mgr := NewTargetMgr(parent)
	value := "$st"
	get := func() string { return value }
	set := func(s string) { value = s }
	field := addLabelAndScriptField(parent, mgr, "k:script", "Base", "A script", get, set, true)
	wrapper := lastLabeledPair(t, parent, "Base")
	c.True(wrapper.Is(field.Parent()), "the field sits inside a wrapper after the label")
	c.Equal(3, len(wrapper.Children()), "the field, the scripting guide button and the markdown guide button")
	c.True(wrapper.Children()[0].Is(field))
	c.True(mgr.Find("k:script").Is(field), "the field is reachable through the target manager")
	c.Equal("A script", tooltipText(field.Tooltip))
	field.SetText("$dx")
	c.Equal("$dx", value)

	plain := addLabelAndScriptField(parent, mgr, "k:plain", "Plain", "Another", get, set, false)
	c.Equal(2, len(lastLabeledPair(t, parent, "Plain").Children()), "no markdown guide button was asked for")
	c.True(mgr.Find("k:plain").Is(plain))
}

// TestWidgetHelpersAcceptPointerOrAccessor verifies that a widget helper edits its value the same way whether it is
// given a pointer to the value or a fieldAccessor for it.
func TestWidgetHelpersAcceptPointerOrAccessor(t *testing.T) {
	c := check.New(t)
	parent := unison.NewPanel()

	byPointer := "start"
	field := addStringField(parent, "Name", "", &byPointer)
	c.Equal("start", field.Text())
	field.SetText("changed")
	c.Equal("changed", byPointer, "a pointer is written through")

	var stored string
	reads := 0
	access := newFieldAccessor(func() string {
		reads++
		return stored
	}, func(value string) { stored = "set:" + value })
	field = addStringField(parent, "Name", "", access)
	c.True(reads > 0, "the accessor's getter supplies the starting text")
	field.SetText("typed")
	c.Equal("set:typed", stored, "the accessor's setter receives the edit")

	level := 2
	popup := addPopup(parent, []int{1, 3}, newFieldAccessor(func() int { return level }, func(v int) { level = v }))
	c.Equal(1, level, "a value missing from the choices is replaced by the first one through the setter")
	popup.Select(3)
	c.Equal(3, level, "a choice is stored through the setter")
}
