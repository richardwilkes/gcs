/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
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
	"fmt"
	"os"
	"strings"

	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	xfs "github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
)

const markdownContentOnlyPrefix = "//////////"

var (
	_ FileBackedDockable = &MarkdownDockable{}
	_ unison.TabCloser   = &MarkdownDockable{}
)

// MarkdownDockable holds the view for an image file.
type MarkdownDockable struct {
	unison.Panel
	path       string
	title      string
	scroll     *unison.ScrollPanel
	markdown   *Markdown
	scale      int
	dragStart  unison.Point
	dragOrigin unison.Point
	inDrag     bool
}

// ShowReleaseNotesMarkdown attempts to show the given markdown content in a dockable.
func ShowReleaseNotesMarkdown(title, content string) {
	ws := WorkspaceFromWindowOrAny(nil)
	if d := ws.LocateFileBackedDockable(markdownContentOnlyPrefix + title); d != nil {
		dc := unison.Ancestor[*unison.DockContainer](d)
		dc.SetCurrentDockable(d)
		dc.AcquireFocus()
		return
	}
	d, err := NewMarkdownDockableWithContent(title, content)
	if err != nil {
		unison.ErrorDialogWithError(fmt.Sprintf(i18n.Text("Unable to open %s"), title), err)
		return
	}
	DisplayNewDockable(nil, d)
}

// NewMarkdownDockable creates a new unison.Dockable for markdown files.
func NewMarkdownDockable(filePath string) (unison.Dockable, error) {
	return newMarkdownDockable(filePath, "", "")
}

// NewMarkdownDockableWithContent creates a new unison.Dockable for markdown content.
func NewMarkdownDockableWithContent(title, content string) (unison.Dockable, error) {
	return newMarkdownDockable(markdownContentOnlyPrefix+title, title, content)
}

func newMarkdownDockable(filePath, title, content string) (unison.Dockable, error) {
	d := &MarkdownDockable{
		path:  filePath,
		title: title,
		scale: 100,
	}
	d.Self = d
	d.SetLayout(&unison.FlexLayout{Columns: 1})

	d.markdown = NewMarkdown()
	d.markdown.MouseDownCallback = d.mouseDown
	d.markdown.MouseDragCallback = d.mouseDrag
	d.markdown.MouseUpCallback = d.mouseUp
	d.markdown.UpdateCursorCallback = d.updateCursor
	d.markdown.SetFocusable(true)
	if !strings.HasPrefix(d.path, markdownContentOnlyPrefix) {
		data, err := os.ReadFile(d.BackingFilePath())
		if err != nil {
			return nil, err
		}
		content = string(data)
	}
	d.markdown.SetContent(content, 0)

	d.scroll = unison.NewScrollPanel()
	d.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
		VGrab:  true,
	})
	d.scroll.SetContent(d.markdown, unison.FillBehavior, unison.FillBehavior)

	toolbar := unison.NewPanel()
	toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.Insets{Bottom: 1},
		false), unison.NewEmptyBorder(unison.StdInsets())))
	toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	toolbar.AddChild(NewDefaultInfoPop())
	toolbar.AddChild(NewScaleField(minPDFDockableScale, maxPDFDockableScale, func() int { return 100 },
		func() int { return d.scale }, func(scale int) { d.scale = scale }, d.scroll, nil, false))
	toolbar.SetLayout(&unison.FlexLayout{
		Columns:  len(toolbar.Children()),
		HSpacing: unison.StdHSpacing,
	})

	d.AddChild(toolbar)
	d.AddChild(d.scroll)

	return d, nil
}

func (d *MarkdownDockable) updateCursor(_ unison.Point) *unison.Cursor {
	if d.inDrag {
		return unison.MoveCursor()
	}
	return unison.ArrowCursor()
}

func (d *MarkdownDockable) mouseDown(where unison.Point, _, _ int, _ unison.Modifiers) bool {
	d.dragStart = d.markdown.PointToRoot(where)
	d.dragOrigin.X, d.dragOrigin.Y = d.scroll.Position()
	d.inDrag = true
	d.markdown.RequestFocus()
	d.UpdateCursorNow()
	return true
}

func (d *MarkdownDockable) mouseDrag(where unison.Point, _ int, _ unison.Modifiers) bool {
	pt := d.dragStart
	pt.Subtract(d.markdown.PointToRoot(where))
	d.scroll.SetPosition(d.dragOrigin.X+pt.X, d.dragOrigin.Y+pt.Y)
	return true
}

func (d *MarkdownDockable) mouseUp(_ unison.Point, _ int, _ unison.Modifiers) bool {
	d.inDrag = false
	d.UpdateCursorNow()
	return true
}

// TitleIcon implements workspace.FileBackedDockable
func (d *MarkdownDockable) TitleIcon(suggestedSize unison.Size) unison.Drawable {
	if strings.HasPrefix(d.path, markdownContentOnlyPrefix) {
		return &unison.DrawableSVG{
			SVG:  svg.MarkdownFile,
			Size: suggestedSize,
		}
	}
	return &unison.DrawableSVG{
		SVG:  model.FileInfoFor(d.BackingFilePath()).SVG,
		Size: suggestedSize,
	}
}

// Title implements workspace.FileBackedDockable
func (d *MarkdownDockable) Title() string {
	if d.title != "" {
		return d.title
	}
	return xfs.BaseName(d.path)
}

// Tooltip implements workspace.FileBackedDockable
func (d *MarkdownDockable) Tooltip() string {
	if strings.HasPrefix(d.path, markdownContentOnlyPrefix) {
		return ""
	}
	return d.BackingFilePath()
}

// BackingFilePath implements workspace.FileBackedDockable
func (d *MarkdownDockable) BackingFilePath() string {
	return d.path
}

// SetBackingFilePath implements workspace.FileBackedDockable
func (d *MarkdownDockable) SetBackingFilePath(p string) {
	d.path = p
	if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
		dc.UpdateTitle(d)
	}
}

// Modified implements workspace.FileBackedDockable
func (d *MarkdownDockable) Modified() bool {
	return false
}

// MayAttemptClose implements unison.TabCloser
func (d *MarkdownDockable) MayAttemptClose() bool {
	return true
}

// AttemptClose implements unison.TabCloser
func (d *MarkdownDockable) AttemptClose() bool {
	if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
		dc.Close(d)
	}
	return true
}
