/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ux

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// EditNote displays the editor for a note.
func EditNote(owner Rebuildable, note *gurps.Note) {
	displayEditor[*gurps.Note, *gurps.NoteEditData](owner, note, svg.GCSNotes, "md:Help/Interface/Note",
		initNoteToolbar, initNoteEditor)
}

func adjustMarkdownThemeForPage(markdown *unison.Markdown) {
	markdown.Font = gurps.PageFieldPrimaryFont
	markdown.Foreground = &unison.IndirectInk{Target: gurps.OnPageColor}
	markdown.HeadingFont[0] = &unison.DynamicFont{Resolver: func() unison.FontDescriptor { return unison.DeriveMarkdownHeadingFont(markdown.Font, 1) }}
	markdown.HeadingFont[1] = &unison.DynamicFont{Resolver: func() unison.FontDescriptor { return unison.DeriveMarkdownHeadingFont(markdown.Font, 2) }}
	markdown.HeadingFont[2] = &unison.DynamicFont{Resolver: func() unison.FontDescriptor { return unison.DeriveMarkdownHeadingFont(markdown.Font, 3) }}
	markdown.HeadingFont[3] = &unison.DynamicFont{Resolver: func() unison.FontDescriptor { return unison.DeriveMarkdownHeadingFont(markdown.Font, 4) }}
	markdown.HeadingFont[4] = &unison.DynamicFont{Resolver: func() unison.FontDescriptor { return unison.DeriveMarkdownHeadingFont(markdown.Font, 5) }}
	markdown.HeadingFont[5] = &unison.DynamicFont{Resolver: func() unison.FontDescriptor { return unison.DeriveMarkdownHeadingFont(markdown.Font, 6) }}
	markdown.CodeBlockFont = &unison.DynamicFont{
		Resolver: func() unison.FontDescriptor {
			fd := unison.MonospacedFont.Font.Descriptor()
			fd.Size = markdown.Font.Size()
			return fd
		},
	}
	markdown.CodeBackground = gurps.PageStandoutColor
	markdown.OnCodeBackground = gurps.OnPageStandoutColor
}

func initNoteToolbar(_ *editor[*gurps.Note, *gurps.NoteEditData], toolbar *unison.Panel) {
	filler := unison.NewPanel()
	filler.SetLayoutData(&unison.FlexLayoutData{HGrab: true})
	toolbar.AddChild(filler)
	toolbar.AddChild(unison.NewLink(i18n.Text("Markdown Guide"), "", "md:Help/Markdown Guide", unison.DefaultLinkTheme, HandleLink))
}

func initNoteEditor(e *editor[*gurps.Note, *gurps.NoteEditData], content *unison.Panel) func() {
	markdown := unison.NewMarkdown(true)
	markdown.ClientData()[WorkingDirKey] = WorkingDirProvider(e.owner)
	markdown.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	adjustMarkdownThemeForPage(markdown)

	labelText := i18n.Text("Notes")
	label := NewFieldLeadingLabel(labelText)
	label.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.EndAlignment,
		VAlign: unison.StartAlignment,
	})
	label.SetBorder(unison.NewEmptyBorder(unison.Insets{Top: 3}))
	content.AddChild(label)
	field := NewMultiLineStringField(nil, "", labelText,
		func() string { return e.editorData.Text },
		func(value string) {
			e.editorData.Text = value
			markdown.SetContent(value, 0)
			content.MarkForLayoutAndRedraw()
			MarkModified(content)
		})
	field.AutoScroll = false
	field.Font = &unison.DynamicFont{
		Resolver: func() unison.FontDescriptor {
			fd := unison.MonospacedFont.Font.Descriptor()
			fd.Size = unison.DefaultFieldTheme.Font.Size()
			return fd
		},
	}
	content.AddChild(field)

	addPageRefLabelAndField(content, &e.editorData.PageRef)

	label = unison.NewLabel()
	label.Text = i18n.Text("Markdown Preview")
	label.HAlign = unison.MiddleAlignment
	label.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	label.SetBorder(
		unison.NewCompoundBorder(
			unison.NewLineBorder(unison.DividerColor, 0, unison.Insets{Bottom: 1}, false),
			unison.NewEmptyBorder(unison.Insets{Top: unison.StdVSpacing * 3}),
		),
	)
	content.AddChild(label)

	markdown.SetContent(e.editorData.Text, 0)

	markdownWrapper := unison.NewPanel()
	markdownWrapper.SetScale(1.33)
	markdownWrapper.SetLayout(&unison.FlexLayout{Columns: 1})
	markdownWrapper.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	content.AddChild(markdownWrapper)

	markdownWrapper.AddChild(markdown)

	return nil
}
