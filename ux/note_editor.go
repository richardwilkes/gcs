// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// EditNote displays the editor for a note.
func EditNote(owner Rebuildable, note *gurps.Note) {
	displayEditor(owner, note, svg.GCSNotes, "md:Help/Interface/Note",
		nil, initNoteEditor, nil)
}

func adjustMarkdownThemeForPage(markdown *unison.Markdown, baseFont unison.Font) {
	markdown.Font = baseFont
	markdown.HeadingFont[0] = &unison.DynamicFont{
		Resolver: func() unison.FontDescriptor { return unison.DeriveMarkdownHeadingFont(markdown.Font, 1) },
	}
	markdown.HeadingFont[1] = &unison.DynamicFont{
		Resolver: func() unison.FontDescriptor { return unison.DeriveMarkdownHeadingFont(markdown.Font, 2) },
	}
	markdown.HeadingFont[2] = &unison.DynamicFont{
		Resolver: func() unison.FontDescriptor { return unison.DeriveMarkdownHeadingFont(markdown.Font, 3) },
	}
	markdown.HeadingFont[3] = &unison.DynamicFont{
		Resolver: func() unison.FontDescriptor { return unison.DeriveMarkdownHeadingFont(markdown.Font, 4) },
	}
	markdown.HeadingFont[4] = &unison.DynamicFont{
		Resolver: func() unison.FontDescriptor { return unison.DeriveMarkdownHeadingFont(markdown.Font, 5) },
	}
	markdown.HeadingFont[5] = &unison.DynamicFont{
		Resolver: func() unison.FontDescriptor { return unison.DeriveMarkdownHeadingFont(markdown.Font, 6) },
	}
	markdown.CodeBlockFont = &unison.DynamicFont{
		Resolver: func() unison.FontDescriptor {
			fd := unison.MonospacedFont.Font.Descriptor()
			fd.Size = markdown.Font.Size()
			return fd
		},
	}
}

func initNoteEditor(e *editor[*gurps.Note, *gurps.NoteEditData], content *unison.Panel) func() {
	markdown := unison.NewMarkdown(true)
	markdown.ClientData()[WorkingDirKey] = WorkingDirProvider(e.owner)
	markdown.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Fill,
		HGrab:  true,
	})
	adjustMarkdownThemeForPage(markdown, &unison.DynamicFont{
		Resolver: func() unison.FontDescriptor {
			fd := fonts.PageFieldPrimary.Descriptor()
			fd.Size = fonts.BaseMarkdown.Size()
			return fd
		},
	})
	markdown.SetVSpacing(xmath.Floor(markdown.Font.LineHeight() / 2))

	labelText := i18n.Text("Notes")
	label := NewFieldLeadingLabel(labelText, false)
	label.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.End,
		VAlign: align.Start,
	})
	label.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: 3}))
	content.AddChild(label)
	field := addScriptField(content, nil, "", labelText,
		i18n.Text("These notes will be interpreted as markdown and may also have scripts embedded in them by wrapping each script in <script>your script goes here</script> tags."),
		func() string { return e.editorData.MarkDown },
		func(value string) {
			e.editorData.MarkDown = value
			markdown.SetContent(gurps.ResolveText(gurps.EntityFromNode(e.target), gurps.ScriptSelfProvider{}, value), 0)
			content.MarkForLayoutAndRedraw()
			MarkModified(content)
		})
	field.Font = &unison.DynamicFont{
		Resolver: func() unison.FontDescriptor {
			fd := unison.MonospacedFont.Font.Descriptor()
			fd.Size = unison.DefaultFieldTheme.Font.Size()
			return fd
		},
	}
	parent := field.Parent()
	if layout, ok := parent.Layout().(*unison.FlexLayout); ok {
		layout.Columns++
	}
	markdownHelpButton := unison.NewSVGButton(svg.MarkdownFile)
	markdownHelpButton.ClickCallback = func() { HandleLink(nil, "md:Help/Markdown Guide") }
	markdownHelpButton.Tooltip = newWrappedTooltip(i18n.Text("Markdown Guide"))
	parent.AddChild(markdownHelpButton)

	addPageRefLabelAndField(content, &e.editorData.PageRef)
	addPageRefHighlightLabelAndField(content, &e.editorData.PageRefHighlight)
	addSourceFields(content, &e.target.SourcedID)

	label = unison.NewLabel()
	label.SetTitle(i18n.Text("Markdown Preview"))
	label.HAlign = align.Middle
	label.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Fill,
		HGrab:  true,
	})
	label.SetBorder(
		unison.NewCompoundBorder(
			unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.Insets{Bottom: 1}, false),
			unison.NewEmptyBorder(geom.Insets{Top: unison.StdVSpacing * 3}),
		),
	)
	content.AddChild(label)

	markdown.SetContent(gurps.ResolveText(gurps.EntityFromNode(e.target), gurps.ScriptSelfProvider{},
		e.editorData.MarkDown), 0)
	content.AddChild(markdown)
	return nil
}
