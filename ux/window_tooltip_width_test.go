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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
)

// TestWidestTooltipWidth verifies that the widest tooltip anywhere in a panel's subtree is what sets the width, and
// that a subtree with no tooltips asks for nothing.
func TestWidestTooltipWidth(t *testing.T) {
	c := check.New(t)
	root := unison.NewPanel()
	c.Equal(float32(0), widestTooltipWidth(root), "no tooltips, no width")

	child := unison.NewPanel()
	child.Tooltip = newWrappedTooltip("short")
	root.AddChild(child)
	_, short, _ := child.Tooltip.Sizes(geom.Size{})
	c.Equal(short.Width, widestTooltipWidth(root), "a child's tooltip counts")

	grandchild := unison.NewPanel()
	grandchild.Tooltip = newWrappedTooltip("a considerably longer tooltip than the short one")
	child.AddChild(grandchild)
	_, long, _ := grandchild.Tooltip.Sizes(geom.Size{})
	c.True(long.Width > short.Width, "the fixture's long tooltip must be the wider one")
	c.Equal(long.Width, widestTooltipWidth(root), "the widest tooltip in the subtree wins, however deep it is")
}

// TestEditorWindowsFitTheirTooltips opens the ancestry and name generator editors in windows of their own and checks
// that each window is at least as wide as the widest tooltip the editor holds. Packed around their fields alone, those
// windows were narrower than their tooltips, which a window squeezes into its own width.
func TestEditorWindowsFitTheirTooltips(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	swapForTest(t, &gurps.GlobalSettings().OpenInWindow, []dgroup.Group{dgroup.Settings})

	chooseMenuBarItem(t, screen, wnd, "File", "New Ancestry")
	ancestry := soleEditor[*ancestryEditorDockable](t, screen, isAncestryEditor)
	checkWindowFitsTooltips(t, c, screen, wnd, ancestry)
	closeEditorWithoutPrompt(t, screen, ancestry)

	chooseMenuBarItem(t, screen, wnd, "File", "New Name Generator")
	names := soleEditor[*nameGeneratorEditorDockable](t, screen, isNameGeneratorEditor)
	checkWindowFitsTooltips(t, c, screen, wnd, names)
	closeEditorWithoutPrompt(t, screen, names)
}

// checkWindowFitsTooltips checks that the dockable was given a window of its own, other than the workspace window, and
// that the window's content is at least as wide as the dockable's widest tooltip. It also checks that the dockable
// alone would have packed narrower than that, since otherwise the window's width would prove nothing.
func checkWindowFitsTooltips(t *testing.T, c check.Checker, screen *unison.HeadlessScreen, workspace *unison.Window, d unison.Dockable) {
	t.Helper()
	var editorWnd *unison.Window
	var contentWidth, tooltipWidth, packedWidth float32
	screen.Do(func() {
		editorWnd = d.AsPanel().Window()
		if editorWnd == nil {
			return
		}
		contentWidth = editorWnd.ContentRect().Width
		tooltipWidth = widestTooltipWidth(d.AsPanel())
		_, pref, _ := d.AsPanel().Sizes(geom.Size{})
		packedWidth = pref.Width
	})
	if editorWnd == nil {
		t.Fatalf("%s has no window", d.Title())
	}
	c.True(editorWnd != workspace, "%s opens in a window of its own", d.Title())
	c.True(tooltipWidth > 0, "%s has tooltips", d.Title())
	c.True(packedWidth < tooltipWidth, "%s packs narrower (%v) than its widest tooltip (%v); the check below would prove nothing otherwise",
		d.Title(), packedWidth, tooltipWidth)
	c.True(contentWidth >= tooltipWidth, "%s's window content (%v wide) fits its widest tooltip (%v wide)",
		d.Title(), contentWidth, tooltipWidth)
}
