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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
	xfs "github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/toolbox/xio/fs/safe"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
)

const markdownContentOnlyPrefix = "//////////"

var (
	_ FileBackedDockable = &MarkdownDockable{}
	_ unison.TabCloser   = &MarkdownDockable{}
	_ ModifiableRoot     = &MarkdownDockable{}
	_ KeyedDockable      = &MarkdownDockable{}
)

// MarkdownDockable holds the view for an image file.
type MarkdownDockable struct {
	unison.Panel
	path              string
	original          string
	content           string
	undoMgr           *unison.UndoManager
	scroller          *unison.ScrollPanel
	markdown          *unison.Markdown
	editor            *StringField
	scale             int
	dragStart         unison.Point
	dragOrigin        unison.Point
	savedScrollX      float32
	savedScrollY      float32
	inDrag            bool
	allowEditing      bool
	needsSaveAsPrompt bool
}

// ShowReadOnlyMarkdown attempts to show the given markdown content in a dockable.
func ShowReadOnlyMarkdown(title, content string) {
	if d := LocateFileBackedDockable(markdownContentOnlyPrefix + title); d != nil {
		ActivateDockable(d)
		return
	}
	d, err := NewMarkdownDockableWithContent(title, content, false, false)
	if err != nil {
		Workspace.ErrorHandler(fmt.Sprintf(i18n.Text("Unable to open %s"), title), err)
		return
	}
	DisplayNewDockable(d)
}

// NewMarkdownDockable creates a new unison.Dockable for markdown files.
func NewMarkdownDockable(filePath string, allowEditing, startInEditMode bool) (unison.Dockable, error) {
	d, err := newMarkdownDockable(filePath, "", allowEditing, startInEditMode)
	if err != nil {
		return nil, err
	}
	d.needsSaveAsPrompt = false
	return d, nil
}

// NewMarkdownDockableWithContent creates a new unison.Dockable for markdown content.
func NewMarkdownDockableWithContent(title, content string, allowEditing, startInEditMode bool) (unison.Dockable, error) {
	return newMarkdownDockable(markdownContentOnlyPrefix+title, content, allowEditing, startInEditMode)
}

func newMarkdownDockable(filePath, content string, allowEditing, startInEditMode bool) (*MarkdownDockable, error) {
	d := &MarkdownDockable{
		path:              filePath,
		undoMgr:           unison.NewUndoManager(200, func(err error) { errs.Log(err) }),
		scale:             gurps.GlobalSettings().General.InitialMarkdownUIScale,
		allowEditing:      allowEditing,
		needsSaveAsPrompt: true,
	}
	d.Self = d
	d.SetLayout(&unison.FlexLayout{Columns: 1})

	d.markdown = unison.NewMarkdown(true)
	d.markdown.ClientData()[WorkingDirKey] = WorkingDirProvider(d)
	insets := unison.NewUniformInsets(20)
	d.markdown.SetBorder(unison.NewEmptyBorder(insets))
	d.markdown.MouseDownCallback = d.mouseDown
	d.markdown.MouseDragCallback = d.mouseDrag
	d.markdown.MouseUpCallback = d.mouseUp
	d.markdown.UpdateCursorCallback = d.updateCursor
	d.markdown.SetFocusable(true)
	d.original = content
	if !strings.HasPrefix(d.path, markdownContentOnlyPrefix) {
		data, err := os.ReadFile(d.BackingFilePath())
		if err != nil {
			return nil, err
		}
		d.original = string(data)
	}
	d.content = txt.NormalizeLineEndings(d.original)
	d.markdown.SetContent(d.content, 0)

	d.editor = NewMultiLineStringField(nil, "", "",
		func() string { return d.content },
		func(value string) {
			d.content = value
			d.markdown.SetContent(value, 0)
			d.editor.MarkForLayoutAndRedraw()
			MarkModified(d.editor)
		})
	unison.UninstallFocusBorders(d.editor, d.editor)
	d.editor.SetBorder(unison.NewEmptyBorder(insets))
	d.editor.NoSelectAllOnFocus = true
	d.editor.AutoScroll = false
	d.editor.Font = &unison.DynamicFont{
		Resolver: func() unison.FontDescriptor {
			fd := unison.MonospacedFont.Font.Descriptor()
			fd.Size = unison.DefaultFieldTheme.Font.Size()
			return fd
		},
	}

	d.scroller = unison.NewScrollPanel()
	d.scroller.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	if allowEditing && startInEditMode {
		d.scroller.SetContent(d.editor, behavior.Follow, behavior.Fill)
	} else {
		d.scroller.SetContent(d.markdown, behavior.Fill, behavior.Fill)
	}

	toolbar := unison.NewPanel()
	toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, 0, unison.Insets{Bottom: 1},
		false), unison.NewEmptyBorder(unison.StdInsets())))
	toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	toolbar.AddChild(NewDefaultInfoPop())
	toolbar.AddChild(
		NewScaleField(
			minPDFDockableScale,
			maxPDFDockableScale,
			func() int { return gurps.GlobalSettings().General.InitialMarkdownUIScale },
			func() int { return d.scale },
			func(scale int) { d.scale = scale },
			nil,
			false,
			d.scroller,
		),
	)

	if allowEditing {
		editToggle := unison.NewSVGButton(svg.Edit)
		editToggle.Sticky = startInEditMode
		editToggle.Tooltip = newWrappedTooltip(i18n.Text("Toggle Edit Mode"))
		editToggle.ClickCallback = func() {
			editToggle.Sticky = !editToggle.Sticky
			editToggle.MarkForRedraw()
			x, y := d.scroller.Position()
			if editToggle.Sticky {
				d.editor.SetScale(d.markdown.Scale())
				d.scroller.SetContent(d.editor, behavior.Follow, behavior.Fill)
				d.editor.RequestFocus()
			} else {
				d.markdown.SetScale(d.editor.Scale())
				d.scroller.SetContent(d.markdown, behavior.Fill, behavior.Fill)
				d.markdown.RequestFocus()
			}
			d.scroller.SetPosition(d.savedScrollX, d.savedScrollY)
			d.savedScrollX, d.savedScrollY = x, y
		}
		toolbar.AddChild(editToggle)
	}

	toolbar.SetLayout(&unison.FlexLayout{
		Columns:  len(toolbar.Children()),
		HSpacing: unison.StdHSpacing,
	})

	d.AddChild(toolbar)
	d.AddChild(d.scroller)

	d.InstallCmdHandlers(SaveItemID, func(_ any) bool { return d.Modified() }, func(_ any) { d.save(false) })
	d.InstallCmdHandlers(SaveAsItemID, func(_ any) bool { return d.allowEditing }, func(_ any) { d.save(true) })

	return d, nil
}

// DockKey implements KeyedDockable.
func (d *MarkdownDockable) DockKey() string {
	return filePrefix + d.path
}

func (d *MarkdownDockable) updateCursor(_ unison.Point) *unison.Cursor {
	if d.inDrag {
		return unison.MoveCursor()
	}
	return unison.ArrowCursor()
}

func (d *MarkdownDockable) mouseDown(where unison.Point, _, _ int, _ unison.Modifiers) bool {
	d.dragStart = d.markdown.PointToRoot(where)
	d.dragOrigin.X, d.dragOrigin.Y = d.scroller.Position()
	d.inDrag = true
	d.markdown.RequestFocus()
	d.UpdateCursorNow()
	return true
}

func (d *MarkdownDockable) mouseDrag(where unison.Point, _ int, _ unison.Modifiers) bool {
	pt := d.dragStart.Sub(d.markdown.PointToRoot(where)).Add(d.dragOrigin)
	d.scroller.SetPosition(pt.X, pt.Y)
	return true
}

func (d *MarkdownDockable) mouseUp(_ unison.Point, _ int, _ unison.Modifiers) bool {
	d.inDrag = false
	d.UpdateCursorNow()
	return true
}

// UndoManager implements undo.Provider
func (d *MarkdownDockable) UndoManager() *unison.UndoManager {
	return d.undoMgr
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
		SVG:  gurps.FileInfoFor(d.BackingFilePath()).SVG,
		Size: suggestedSize,
	}
}

// Title implements workspace.FileBackedDockable
func (d *MarkdownDockable) Title() string {
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
	UpdateTitleForDockable(d)
}

// Modified implements workspace.FileBackedDockable
func (d *MarkdownDockable) Modified() bool {
	return d.allowEditing && d.original != d.content
}

// MarkModified implements ModifiableRoot.
func (d *MarkdownDockable) MarkModified(_ unison.Paneler) {
	UpdateTitleForDockable(d)
}

// MayAttemptClose implements unison.TabCloser
func (d *MarkdownDockable) MayAttemptClose() bool {
	return true
}

// AttemptClose implements unison.TabCloser
func (d *MarkdownDockable) AttemptClose() bool {
	if AttemptSaveForDockable(d) {
		return AttemptCloseForDockable(d)
	}
	return false
}

func (d *MarkdownDockable) save(forceSaveAs bool) bool {
	success := false
	if forceSaveAs || d.needsSaveAsPrompt {
		success = SaveDockableAs(d, "md", d.saveData, func(path string) {
			d.path = path
			d.original = d.content
		})
	} else {
		success = SaveDockable(d, d.saveData, func() { d.original = d.content })
	}
	if success {
		d.needsSaveAsPrompt = false
	}
	return success
}

func (d *MarkdownDockable) saveData(filePath string) error {
	dirPath := filepath.Dir(filePath)
	if err := os.MkdirAll(dirPath, 0o750); err != nil {
		return errs.NewWithCause(dirPath, err)
	}
	if err := safe.WriteFileWithMode(filePath, func(w io.Writer) error {
		if _, innerErr := w.Write([]byte(d.content)); innerErr != nil {
			return errs.NewWithCause(filePath, innerErr)
		}
		return nil
	}, 0o640); err != nil {
		return errs.NewWithCause(filePath, err)
	}
	return nil
}
