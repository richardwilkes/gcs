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
	"hash"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xstrings"
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
	_ gurps.Hashable     = &MarkdownDockable{}
)

// MarkdownDockable holds the view for a markdown file.
type MarkdownDockable struct {
	fileBackedPanel
	content      string
	undoMgr      *unison.UndoManager
	scroller     *unison.ScrollPanel
	markdown     *unison.Markdown
	editor       *StringField
	pan          scrollPanDrag
	scale        int
	savedScrollX float32
	savedScrollY float32
	allowEditing bool
}

// ShowReadOnlyMarkdown attempts to show the given markdown content in a dockable.
func ShowReadOnlyMarkdown(title, content string) {
	if d := LocateFileBackedDockable(markdownContentOnlyPrefix + title); d != nil {
		ActivateDockable(d)
		return
	}
	DisplayNewDockable(NewMarkdownDockableWithContent(title, content, false, false))
}

// NewMarkdownDockable creates a new unison.Dockable for markdown files.
func NewMarkdownDockable(filePath string, allowEditing, startInEditMode bool) (unison.Dockable, error) {
	return openDockableFromFile(filePath, readMarkdownFile, func(filePath, content string) *MarkdownDockable {
		return newMarkdownDockable(filePath, content, allowEditing, startInEditMode)
	})
}

func readMarkdownFile(fileSystem fs.FS, filePath string) (string, error) {
	data, err := fs.ReadFile(fileSystem, filePath)
	if err != nil {
		return "", errs.Wrap(err)
	}
	return string(data), nil
}

// NewMarkdownDockableWithContent creates a new unison.Dockable for markdown content that is not in a file.
func NewMarkdownDockableWithContent(title, content string, allowEditing, startInEditMode bool) unison.Dockable {
	return newMarkdownDockable(markdownContentOnlyPrefix+title, content, allowEditing, startInEditMode)
}

func newMarkdownDockable(filePath, content string, allowEditing, startInEditMode bool) *MarkdownDockable {
	d := &MarkdownDockable{
		undoMgr:      unison.NewUndoManager(200, func(err error) { errs.Log(err) }),
		scale:        gurps.GlobalSettings().General.InitialMarkdownUIScale,
		allowEditing: allowEditing,
	}
	d.Self = d
	d.SetLayout(&unison.FlexLayout{Columns: 1})

	d.markdown = unison.NewMarkdown(true)
	d.markdown.ClientData()[WorkingDirKey] = WorkingDirProvider(d)
	insets := geom.NewUniformInsets(20)
	d.markdown.SetBorder(unison.NewEmptyBorder(insets))
	d.markdown.SetFocusable(true)
	// The content is normalized before it is hashed, since the editor produces LF line endings and that hash is what
	// determines whether the dockable has been modified. Without this, a file stored with CRLF (or CR) line endings
	// would be reported as modified the moment it was opened.
	d.content = xstrings.NormalizeLineEndings(content)
	d.markdown.SetContent(d.content, 0)
	if allowEditing {
		d.initFileEditor(d, filePath, "md", d.saveData, d)
	} else {
		// Content that can't be edited can't have unsaved changes either, so the dockable is only a viewer of it.
		d.initFileViewer(d, filePath)
	}

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
	d.pan.install(d.scroller, d.markdown.AsPanel(), d.markdown)

	toolbar := newToolbar()
	toolbar.AddChild(NewDefaultInfoPop())
	toolbar.AddChild(NewMarkdownGuideButton())
	toolbar.AddChild(
		NewScaleField(
			minPDFDockableScale,
			maxPDFDockableScale,
			func() int { return gurps.GlobalSettings().General.InitialMarkdownUIScale },
			func() int { return d.scale },
			func(scale int) { d.scale = scale },
			nil,
			false,
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

	finishToolbarLayout(toolbar)

	d.AddChild(toolbar)
	d.AddChild(d.scroller)

	d.InstallCmdHandlers(SaveItemID, func(_ any) bool { return d.Modified() }, func(_ any) { d.save(false) })
	d.InstallCmdHandlers(SaveAsItemID, func(_ any) bool { return d.allowEditing }, func(_ any) { d.save(true) })

	return d
}

// Hash implements gurps.Hashable, so that the dockable can tell whether its content has changed since it was opened or
// last saved.
func (d *MarkdownDockable) Hash(h hash.Hash) {
	xhash.StringWithLen(h, d.content)
}

// ScrollToAnchor scrolls the heading associated with the given anchor into view. Has no effect if the markdown is
// currently being edited rather than displayed, or if no matching anchor exists.
func (d *MarkdownDockable) ScrollToAnchor(anchor string) {
	d.ValidateLayout()
	d.markdown.ScrollToAnchor(anchor)
}

// UndoManager implements unison.UndoManagerProvider.
func (d *MarkdownDockable) UndoManager() *unison.UndoManager {
	return d.undoMgr
}

// TitleIcon implements unison.Dockable. Content that is not in a file has no file type to take an icon from.
func (d *MarkdownDockable) TitleIcon(suggestedSize geom.Size) unison.Drawable {
	if strings.HasPrefix(d.path, markdownContentOnlyPrefix) {
		return &unison.DrawableSVG{
			SVG:  svg.MarkdownFile,
			Size: suggestedSize,
		}
	}
	return d.fileBackedPanel.TitleIcon(suggestedSize)
}

// Tooltip implements unison.Dockable. Content that is not in a file has no path to show.
func (d *MarkdownDockable) Tooltip() string {
	if strings.HasPrefix(d.path, markdownContentOnlyPrefix) {
		return ""
	}
	return d.fileBackedPanel.Tooltip()
}

// MarkModified implements ModifiableRoot.
func (d *MarkdownDockable) MarkModified(_ unison.Paneler) {
	UpdateTitleForDockable(d)
}

func (d *MarkdownDockable) saveData(filePath string) error {
	dirPath := filepath.Dir(filePath)
	if err := os.MkdirAll(dirPath, 0o750); err != nil {
		return errs.NewWithCause(dirPath, err)
	}
	if err := xos.WriteSafeFile(filePath, func(w io.Writer) error {
		if _, innerErr := w.Write([]byte(d.content)); innerErr != nil {
			return errs.NewWithCause(filePath, innerErr)
		}
		return nil
	}); err != nil {
		return errs.NewWithCause(filePath, err)
	}
	return nil
}
