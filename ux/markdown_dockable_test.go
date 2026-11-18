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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
)

// TestMarkdownDockableNotModifiedWhenOpened verifies that opening markdown, whether read from a file or given as
// content, never reports the dockable as modified, whatever its line endings, while an edit is still reported. The
// content is normalized to LF when loaded, so the copy retained for comparison must be too, or a file with CRLF (or CR)
// line endings would prompt to save on close though nothing was edited.
func TestMarkdownDockableNotModifiedWhenOpened(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name string
		data string
	}{
		{name: "LF", data: "# Title\n\nSome text.\n"},
		{name: "CRLF", data: "# Title\r\n\r\nSome text.\r\n"},
		{name: "CR", data: "# Title\r\rSome text.\r"},
		{name: "mixed", data: "# Title\r\n\nSome text.\r"},
		{name: "empty", data: ""},
	} {
		p := filepath.Join(t.TempDir(), "notes.md")
		c.NoError(os.WriteFile(p, []byte(tc.data), 0o600), tc.name)
		fromFile, err := NewMarkdownDockable(p, true, false)
		c.NoError(err, tc.name)
		for _, opened := range []struct {
			how      string
			dockable unison.Dockable
		}{
			{how: "file", dockable: fromFile},
			{how: "content", dockable: NewMarkdownDockableWithContent(tc.name, tc.data, true, false)},
		} {
			d, ok := opened.dockable.(*MarkdownDockable)
			c.True(ok, "%s %s: expected a markdown dockable", tc.name, opened.how)
			if !ok {
				continue
			}
			c.False(d.Modified(), "%s %s: a freshly opened markdown dockable must not be reported as modified", tc.name,
				opened.how)
			d.content += "More text.\n"
			c.True(d.Modified(), "%s %s: an edited markdown dockable must be reported as modified", tc.name, opened.how)
		}
	}
}

// Saving an edited markdown file writes the edited content back to it and clears the modified state, which is what the
// hash over its content is for.
func TestMarkdownDockableSaves(t *testing.T) {
	c := check.New(t)
	p := filepath.Join(t.TempDir(), "notes.md")
	c.NoError(os.WriteFile(p, []byte("# Title\n"), 0o600))
	dockable, err := NewMarkdownDockable(p, true, false)
	c.NoError(err)
	d, ok := dockable.(*MarkdownDockable)
	c.True(ok)
	d.content += "\nMore text.\n"
	c.True(d.Modified())
	c.True(d.save(false))
	c.False(d.Modified(), "the content must not be reported as modified once it has been saved")
	data, err := os.ReadFile(p)
	c.NoError(err)
	c.Equal("# Title\n\nMore text.\n", string(data))
}

// TestMarkdownDockableCaretReadsOnFromAnAnchor verifies that the keys moving the reading caret go on from the heading
// ScrollToAnchor, or a click on a link to it, shows rather than scrolling the view back to the top of the document or
// to the link.
func TestMarkdownDockableCaretReadsOnFromAnAnchor(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	filler := strings.Repeat("Some text.\n\n", 100)
	content := "# Top\n\n[Far away](#far)\n\n" + filler + "## Far\n\n" + filler
	for _, tc := range []struct {
		name   string
		arrive func(d *MarkdownDockable)
	}{
		{name: "ScrollToAnchor", arrive: func(d *MarkdownDockable) { screen.Do(func() { d.ScrollToAnchor("far") }) }},
		{
			name: "link to the heading",
			arrive: func(d *MarkdownDockable) {
				var link *unison.Label
				screen.Do(func() {
					for _, one := range uxtest.PanelsOfType[*unison.Label](d.markdown.AsPanel()) {
						if one.Accessibility.URL == "#far" {
							link = one
						}
					}
				})
				c.True(link != nil, "the document must hold the link")
				if link != nil {
					screen.Click(screen.PanelCenter(link))
				}
			},
		},
	} {
		d, ok := NewMarkdownDockableWithContent("notes", content, false, false).(*MarkdownDockable)
		c.True(ok, "%s: expected a markdown dockable", tc.name)
		var before, after float32
		screen.Do(func() {
			DisplayNewDockable(d)
			d.markdown.RequestFocus()
		})
		screen.Sync()
		tc.arrive(d)
		screen.Do(func() { _, before = d.scroller.Position() })
		c.True(before > 0, "%s: the heading must be below the top of the document", tc.name)
		screen.KeyPress(unison.KeyDown, 0)
		screen.Do(func() { _, after = d.scroller.Position() })
		c.Equal(before, after, "%s: Down reads on from the heading", tc.name)

		uxtest.CloseEditorWithoutPrompt(t, screen, d)
	}
}

// TestMarkdownViewsWrapToTheirViewAtAnyScale verifies that a markdown file and the User Guide wrap their document to
// the width of their view at the scale it is shown at, so that a larger scale needs no scrolling sideways.
func TestMarkdownViewsWrapToTheirViewAtAnyScale(t *testing.T) {
	c := check.New(t)
	content := "# Top\n\n" + strings.Repeat("Some text that runs on across the width of the view. ", 50) + "\n"
	for _, tc := range []struct {
		name string
		open func() (d unison.Dockable, scroll *unison.ScrollPanel, markdown *unison.Markdown)
	}{
		{
			name: "markdown file",
			open: func() (unison.Dockable, *unison.ScrollPanel, *unison.Markdown) {
				d := newMarkdownDockable(markdownContentOnlyPrefix+"notes", content, false, false)
				return d, d.scroller, d.markdown
			},
		},
		{
			name: "User Guide",
			open: func() (unison.Dockable, *unison.ScrollPanel, *unison.Markdown) {
				d := newUserGuideDockable()
				d.navigate("Equipment.md", "", true)
				return d, d.docScroll, d.markdown
			},
		},
	} {
		d, scroll, markdown := tc.open()
		d.AsPanel().SetFrameRect(geom.NewRect(0, 0, 1200, 800))
		d.AsPanel().ValidateLayout()
		d.AsPanel().PerformCmd(nil, Scale200ItemID)
		c.Equal(geom.NewPoint(2, 2), markdown.Scale(), "%s: the document is shown at 200%%", tc.name)
		view := scroll.ContentView().ContentRect(false).Width
		c.True(view > 0, "%s: the view must have been laid out", tc.name)
		c.True(markdown.FrameRect().Width <= view, "%s: the document is %v wide in a view %v wide", tc.name,
			markdown.FrameRect().Width, view)
	}
}

// TestMarkdownViewsKeepTheirPlaceWhenRewrapped verifies that a markdown file and the User Guide keep the heading at the
// top of their view there when they wrap their document afresh: at another scale, in a narrower view or, for the guide,
// with its side bar toggled. A markdown file keeps it too when the scale or the view changes while it is being edited.
func TestMarkdownViewsKeepTheirPlaceWhenRewrapped(t *testing.T) {
	c := check.New(t)
	filler := strings.Repeat("Some text that runs on across the width of the view, and then some more of it. ", 10) +
		"\n\n"
	content := "# Top\n\n" + strings.Repeat(filler, 20) + "## Middle\n\n" + strings.Repeat(filler, 40)
	type view struct {
		d        unison.Dockable
		scroll   *unison.ScrollPanel
		markdown *unison.Markdown
		heading  string
		side     func() // toggles the side bar; nil for a view with none
		edit     func() // toggles edit mode; nil for a view with none
	}
	for _, open := range []struct {
		name string
		open func() view
	}{
		{
			name: "markdown file",
			open: func() view {
				d := newMarkdownDockable(markdownContentOnlyPrefix+"notes", content, true, false)
				return view{
					d:        d,
					scroll:   d.scroller,
					markdown: d.markdown,
					heading:  "middle",
					edit:     func() { clickMarkdownEditToggle(c, d) },
				}
			},
		},
		{
			name: "User Guide",
			open: func() view {
				d := newUserGuideDockable()
				d.navigate("Traits.md", "", true)
				return view{
					d:        d,
					scroll:   d.docScroll,
					markdown: d.markdown,
					heading:  "managing-traits",
					side:     func() { d.side.toggle(d.content, d.markdown) },
				}
			},
		},
	} {
		for _, change := range []struct {
			name   string
			change func(v view) bool // returns false when the view has no such change
		}{
			{
				name: "scale changed",
				change: func(v view) bool {
					v.d.AsPanel().PerformCmd(nil, Scale200ItemID)
					return true
				},
			},
			{
				name: "view narrowed",
				change: func(v view) bool {
					v.d.AsPanel().SetFrameRect(geom.NewRect(0, 0, 700, 800))
					return true
				},
			},
			{
				name: "side bar toggled",
				change: func(v view) bool {
					if v.side == nil {
						return false
					}
					v.side()
					return true
				},
			},
			{
				name: "scale changed while editing",
				change: func(v view) bool {
					if v.edit == nil {
						return false
					}
					v.edit()
					v.d.AsPanel().PerformCmd(nil, Scale200ItemID)
					v.edit()
					return true
				},
			},
			{
				name: "view narrowed while editing",
				change: func(v view) bool {
					if v.edit == nil {
						return false
					}
					v.edit()
					v.d.AsPanel().SetFrameRect(geom.NewRect(0, 0, 700, 800))
					v.d.AsPanel().ValidateLayout()
					v.edit()
					return true
				},
			},
		} {
			name := open.name + ", " + change.name
			v := open.open()
			v.d.AsPanel().SetFrameRect(geom.NewRect(0, 0, 1200, 800))
			v.d.AsPanel().ValidateLayout()
			c.True(v.markdown.ScrollToAnchor(v.heading), "%s: the heading must be found", name)
			_, y := v.scroll.Position()
			c.True(y > 0, "%s: the heading must be below the top of the document", name)
			if !change.change(v) {
				continue
			}
			v.d.AsPanel().ValidateLayout()
			_, y = v.scroll.Position()
			v.markdown.ScrollToAnchor(v.heading)
			_, want := v.scroll.Position()
			c.True(y-want > -1 && y-want < 1, "%s: the heading is %v from the top of the view", name, want-y)
		}
	}
}

// TestMarkdownEditorKeepsItsPlace verifies that the editor of a markdown file keeps the line at the top of its view
// there when the scale or the view's width changes, and comes back to it when it was left, though the scale or the view
// changed while the document was being read.
func TestMarkdownEditorKeepsItsPlace(t *testing.T) {
	c := check.New(t)
	filler := strings.Repeat("Some text that runs on across the width of the view, and then some more of it. ", 10) +
		"\n\n"
	above := "# Top\n\n" + strings.Repeat(filler, 20)
	content := above + "## Middle\n\n" + strings.Repeat(filler, 40)
	// index is where the line under test begins.
	index := len([]rune(above))
	for _, tc := range []struct {
		name   string
		change func(d *MarkdownDockable)
	}{
		{
			name: "scale changed while reading",
			change: func(d *MarkdownDockable) {
				clickMarkdownEditToggle(c, d)
				d.PerformCmd(nil, Scale200ItemID)
				clickMarkdownEditToggle(c, d)
			},
		},
		{
			name: "view narrowed while reading",
			change: func(d *MarkdownDockable) {
				clickMarkdownEditToggle(c, d)
				d.SetFrameRect(geom.NewRect(0, 0, 700, 800))
				d.ValidateLayout()
				clickMarkdownEditToggle(c, d)
			},
		},
		{name: "scale changed while editing", change: func(d *MarkdownDockable) { d.PerformCmd(nil, Scale200ItemID) }},
		{
			name:   "view narrowed while editing",
			change: func(d *MarkdownDockable) { d.SetFrameRect(geom.NewRect(0, 0, 700, 800)) },
		},
	} {
		d := newMarkdownDockable(markdownContentOnlyPrefix+"notes", content, true, true)
		d.SetFrameRect(geom.NewRect(0, 0, 1200, 800))
		d.ValidateLayout()
		// lineBelowTop returns how far the line that starts at index lies below the top of the view.
		lineBelowTop := func() float32 {
			view := d.scroller.ContentView()
			return d.editor.PointTo(d.editor.FromSelectionIndex(index), view.AsPanel()).Y - view.ContentRect(false).Y
		}
		_, y := d.scroller.Position()
		d.scroller.SetPosition(0, y+lineBelowTop())
		c.True(lineBelowTop() > -1 && lineBelowTop() < 1, "%s: the line must be at the top of the view", tc.name)
		tc.change(d)
		d.ValidateLayout()
		below := lineBelowTop()
		c.True(below > -1 && below < 1, "%s: the line is %v below the top of the view", tc.name, below)
	}
}

// clickMarkdownEditToggle clicks the button that toggles the markdown dockable's edit mode.
func clickMarkdownEditToggle(c check.Checker, d *MarkdownDockable) {
	for _, b := range uxtest.PanelsOfType[*unison.Button](d.AsPanel()) {
		if drawable, ok := b.Drawable.(*unison.DrawableSVG); ok && drawable.SVG == svg.Edit {
			b.ClickCallback()
			return
		}
	}
	c.True(false, "the dockable must have an edit mode toggle")
}

// TestMarkdownDockableResolvesRelativePathsAgainstItsFile verifies that a markdown dockable opened from a file resolves
// the relative links and images in its content against the file's own folder, while content that is not in a file
// resolves them against the current directory. The working directory is captured when the dockable is built, before
// its content is set and its images are first retrieved, so capturing it before the backing file path is known sends
// every relative image to the wrong folder.
func TestMarkdownDockableResolvesRelativePathsAgainstItsFile(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "notes.md")
	c.NoError(os.WriteFile(p, []byte("![icon](images/icon.svg)\n"), 0o600))
	for _, allowEditing := range []bool{true, false} {
		opened, err := NewMarkdownDockable(p, allowEditing, false)
		c.NoError(err, "allowEditing=%v", allowEditing)
		d, ok := opened.(*MarkdownDockable)
		c.True(ok, "allowEditing=%v: expected a markdown dockable", allowEditing)
		if !ok {
			continue
		}
		c.Equal(dir, d.markdown.ClientData()[WorkingDirKey],
			"allowEditing=%v: a markdown file must resolve relative paths against its own folder", allowEditing)
	}
	d, ok := NewMarkdownDockableWithContent("Notes", "![icon](images/icon.svg)\n", false, false).(*MarkdownDockable)
	c.True(ok, "expected a markdown dockable")
	if ok {
		c.Equal(".", d.markdown.ClientData()[WorkingDirKey],
			"markdown content that is not in a file must resolve relative paths against the current directory")
	}
}
