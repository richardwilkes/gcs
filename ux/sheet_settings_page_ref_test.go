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

	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
)

func TestSheetSettingsCheckBoxPageRefIsOnePieceOfText(t *testing.T) {
	c := check.New(t)
	d := &sheetSettingsDockable{}
	panel := unison.NewPanel()
	var clicked []bool
	box := d.addCheckBox(panel, "Use the rule", "B269", true, func(checked bool) { clicked = append(clicked, checked) })
	c.Equal("Use the rule", box.Text.String())
	if len(panel.Children()) != 1 {
		t.Fatalf("the checkbox and its page reference must share a wrapper, but the panel holds %d children",
			len(panel.Children()))
	}
	children := panel.Children()[0].Children()
	if len(children) != 2 {
		t.Fatalf("the wrapper must hold the checkbox and its page reference, but holds %d children", len(children))
	}
	c.Equal(box.AsPanel(), children[0])
	label, ok := children[1].Self.(*TextLabel)
	if !ok {
		t.Fatalf("the page reference must be a label holding the link, not a %T", children[1].Self)
	}
	c.Equal("(B269)", label.String())
	c.True(label.Focusable(), "the link is reached by the keyboard focus, screen reader or not")
	_, pref, _ := label.Sizes(geom.Size{})
	label.SetFrameRect(geom.NewRect(0, 0, pref.Width, pref.Height))
	links := label.Links()
	if len(links) != 1 {
		t.Fatalf("the page reference must hold one link, but holds %d", len(links))
	}
	c.Equal("B269", links[0].Ref)
	var followed []string
	label.LinkHandler = func(ref string) { followed = append(followed, ref) }
	c.True(label.keyDown(unison.KeySpace, 0, false), "Space follows the only link")
	c.Equal([]string{"B269"}, followed)

	plain := unison.NewPanel()
	box = d.addCheckBox(plain, "No page", "", false, func(bool) {})
	c.Equal([]*unison.Panel{box.AsPanel()}, plain.Children(), "a checkbox without a page reference stands alone")
}

// LinkPageRefs recognizes only Basic Set references, so this covers those to other books, such as "PY65:30".
func TestSheetSettingsPageRefsAreAllLinks(t *testing.T) {
	c := check.New(t)
	d := &sheetSettingsDockable{}
	var refs []string
	for _, option := range sheetOptions() {
		if option.pageRef == "" {
			continue
		}
		refs = append(refs, option.pageRef)
		panel := unison.NewPanel()
		d.addCheckBox(panel, option.title, option.pageRef, false, func(bool) {})
		labels := uxtest.PanelsOfType[*TextLabel](panel)
		if len(labels) != 1 {
			t.Fatalf("%q must be followed by its page reference, but %d labels follow it", option.title, len(labels))
		}
		label := labels[0]
		c.Equal("("+option.pageRef+")", label.String())
		_, pref, _ := label.Sizes(geom.Size{})
		label.SetFrameRect(geom.NewRect(0, 0, pref.Width, pref.Height))
		links := label.Links()
		if len(links) != 1 {
			t.Fatalf("the page reference of %q must hold one link, but holds %d", option.title, len(links))
		}
		c.Equal(option.pageRef, links[0].Ref, "the link of %q is the whole of its page reference", option.title)
		var followed []string
		label.LinkHandler = func(ref string) { followed = append(followed, ref) }
		c.True(label.keyDown(unison.KeySpace, 0, false))
		c.Equal([]string{option.pageRef}, followed, "the link of %q opens its page", option.title)
	}
	c.Equal([]string{"P102", "PY65:30", "B269", "PY120:7"}, refs, "the sheet settings that cite a page")
}

func TestFindRefs(t *testing.T) {
	c := check.New(t)
	found := func(text string, refs ...string) []string {
		spans := findRefs(text, refs)
		if len(spans) == 0 {
			return nil
		}
		result := make([]string, 0, len(spans))
		for _, one := range spans {
			result = append(result, text[one[0]:one[1]])
		}
		return result
	}
	c.Equal([]string{"PY65:30"}, found("(PY65:30)", "PY65:30"))
	c.Equal([]string{"B351", "HT55"}, found("Hiking (B351, HT55)", "HT55", "B351"), "in the order they appear")
	c.Equal([]string{"B102", "B10"}, found("(B102, B10)", "B10", "B102"), "a reference is found as a whole word")
	c.Equal([]string{"B10"}, found("B102 and B10", "B10"), "and never within another")
	c.Equal([]string{"B10", "B10"}, found("B10 and B10", "B10"), "each time it appears")
	c.Equal([]string{"PY65:30"}, found("PY65:30 PY65", "PY65:30", "PY65:3"), "the longer of two wins")
	c.Nil(found("nothing to find", "B10", ""))
}
