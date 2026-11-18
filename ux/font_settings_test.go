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
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestFontChangeRewrapsMarkdownViews verifies that a change of font lays the document of every open markdown view,
// including the About box's, out again with it, rather than leaving its lines broken and measured with the font it was
// built with.
func TestFontChangeRewrapsMarkdownViews(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	saved := fonts.BaseMarkdown.Font
	t.Cleanup(func() { fonts.BaseMarkdown.Font = saved })
	index := slices.IndexFunc(fonts.CurrentFonts(),
		func(f *fonts.ThemedFont) bool { return f.Font == fonts.BaseMarkdown })
	c.True(index >= 0, "the markdown font must be one of the fonts set in the Font settings")
	content := "# Top\n\n" + strings.Repeat("Some text that runs on across the width of the view. ", 200) + "\n"
	file, ok := NewMarkdownDockableWithContent("notes", content, false, false).(*MarkdownDockable)
	c.True(ok, "expected a markdown dockable")
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	c.True(ok, "expected a sheet")
	var about *unison.Window
	screen.Do(func() {
		DisplayNewDockable(file)
		showUserGuide("Equipment.md", "")
		note := gurps.NewNote(sheet.Entity(), nil, false)
		note.MarkDown = content
		EditNote(sheet, note)
		ShowSheetSettings(nil)
		ShowAbout(nil)
		about = aboutWnd
	})
	screen.Sync()
	guide := uxtest.SoleEditor[*userGuideDockable](t, screen, func(unison.Dockable) bool { return true })
	noteEditor := uxtest.SoleEditor[*editor[*gurps.Note, *gurps.NoteEditData]](t, screen,
		func(unison.Dockable) bool { return true })
	settings := uxtest.SoleEditor[*sheetSettingsDockable](t, screen, func(unison.Dockable) bool { return true })
	// soleMarkdown returns the one markdown within the panel.
	soleMarkdown := func(name string, root func() *unison.Panel) (markdown *unison.Markdown) {
		screen.Do(func() {
			all := uxtest.PanelsOfType[*unison.Markdown](root())
			c.Equal(1, len(all), "%s: the markdown views", name)
			if len(all) != 0 {
				markdown = all[0]
			}
		})
		return markdown
	}
	views := []struct {
		name     string
		markdown *unison.Markdown
	}{
		{name: "markdown file", markdown: file.markdown},
		{name: "User Guide", markdown: guide.markdown},
		{name: "note editor's preview", markdown: soleMarkdown("note editor", noteEditor.AsPanel)},
		{name: "damage progression description", markdown: soleMarkdown("sheet settings", settings.AsPanel)},
		{name: "About box", markdown: soleMarkdown("About box", func() *unison.Panel { return about.Content() })},
	}
	heights := func() []float32 {
		var all []float32
		screen.Do(func() {
			for _, view := range views {
				all = append(all, view.markdown.FrameRect().Height)
			}
		})
		return all
	}
	before := heights()
	screen.Do(func() {
		fd := saved.Descriptor()
		fd.Size *= 2
		(&fontSettingsDockable{}).applyFont(index, fd)
	})
	screen.Sync()
	after := heights()
	var aboutFits bool
	screen.Do(func() {
		markdown := views[len(views)-1].markdown
		content := about.Content()
		aboutFits = content.RectToRoot(content.ContentRect(true)).Contains(markdown.RectToRoot(markdown.ContentRect(true)))
		(&fontSettingsDockable{}).applyFont(index, saved.Descriptor())
	})
	for i, view := range views {
		c.True(after[i] > before[i]*1.5, "%s: doubling the size of its font took the document from %v to %v high",
			view.name, before[i], after[i])
	}
	c.True(aboutFits, "the About box, which cannot be resized, must be fitted to its larger document")

	screen.Do(func() { about.AttemptClose() })
	uxtest.CloseEditorWithoutPrompt(t, screen, settings)
	uxtest.CloseEditorWithoutPrompt(t, screen, noteEditor)
	uxtest.CloseEditorWithoutPrompt(t, screen, guide)
	uxtest.CloseEditorWithoutPrompt(t, screen, file)
	uxtest.CloseEditorWithoutPrompt(t, screen, sheet)
}

// TestFontSettingsResetShowsTheFactoryFonts verifies that resetting the Font settings puts the factory fonts back in
// use and shows them in the font panels, which edit the fonts in use.
func TestFontSettingsResetShowsTheFactoryFonts(t *testing.T) {
	c := check.New(t)
	g := gurps.GlobalSettings()
	uxtest.SwapForTest(t, &g.Fonts, g.Fonts)
	live := fonts.CurrentFonts()[0]
	saved := live.Font.Font
	t.Cleanup(func() { live.Font.Font = saved })
	var factory unison.FontDescriptor
	for _, one := range fonts.FactoryFonts() {
		if one.ID == live.ID {
			factory = one.Font.Descriptor()
		}
	}
	d := &fontSettingsDockable{}
	d.initContent(unison.NewPanel())
	changed := factory
	changed.Size *= 2
	d.fontPanels[0].SetFontDescriptor(changed)
	c.Equal(changed, live.Font.Descriptor(), "precondition: the font panel edits the font in use")

	d.reset()
	c.Equal(factory, live.Font.Descriptor(), "the reset puts the factory font back in use")
	c.Equal(factory, d.fontPanels[0].FontDescriptor(), "and the font panel shows it")
}
