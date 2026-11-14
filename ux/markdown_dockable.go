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
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
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
	content  string
	undoMgr  *unison.UndoManager
	scroller *unison.ScrollPanel
	markdown *unison.Markdown
	editor   *StringField
	pan      scrollPanDrag
	scale    int
	// markdownLeftAt and editorLeftAt are where the markdown and the editor were last left for the other, kept as
	// places in their text rather than as scroll offsets, which a change of scale or of the view's width made while the
	// other is shown would move: what was at the top of the markdown's view, and the index of the rune that begins the
	// line at the top of the editor's. markdownLeftX is the markdown's scroll across then, in its own coordinates.
	markdownLeftAt markdownSpot
	markdownLeftX  float32
	editorLeftAt   int
	allowEditing   bool
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

	d.markdown = unison.NewMarkdown(false)
	insets := geom.NewUniformInsets(20)
	d.markdown.SetBorder(unison.NewEmptyBorder(insets))
	d.markdown.SetFocusable(true)
	d.markdown.AnchorLinkHandler = d.ScrollToAnchor
	// The content is normalized before it is hashed, since the editor produces LF line endings and that hash is what
	// determines whether the dockable has been modified. Without this, a file stored with CRLF (or CR) line endings
	// would be reported as modified the moment it was opened.
	d.content = xstrings.NormalizeLineEndings(content)
	if allowEditing {
		d.initFileEditor(d, filePath, "md", d.saveData, d)
	} else {
		// Content that can't be edited can't have unsaved changes either, so the dockable is only a viewer of it.
		d.initFileViewer(d, filePath)
	}
	// The working directory that relative links and images resolve against comes from the backing file path, so it
	// must be captured after that path is set, and before the content is set, since that is when the images are first
	// retrieved and the markdown is not yet inside the dockable for the provider to find.
	d.markdown.ClientData()[WorkingDirKey] = WorkingDirProvider(d)
	d.markdown.SetContent(d.content, 0)

	d.editor = NewMultiLineStringField(nil, "", "",
		func() string { return d.content },
		func(value string) {
			d.content = value
			setMarkdownContentInView(d.scroller, d.markdown, []byte(value))
			d.editor.MarkForLayoutAndRedraw()
			MarkModified(d.editor)
		})
	// The editor fills the view with no label beside it.
	d.editor.Accessibility.Name = i18n.Text("Markdown")
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
	wrapMarkdownToView(d.scroller, d.markdown)
	// The scroll panel rewraps the editor to a new width when it is laid out, so it is laid out here, once the line at the
	// top of the view has been found, to put that line back there.
	d.scroller.FrameChangeCallback = func() {
		if d.editorShown() {
			width := d.editor.FrameRect().Width
			at := d.editorLineAtTopOfView()
			d.scroller.ValidateLayout()
			if d.editor.FrameRect().Width != width {
				d.scrollEditorLineToTopOfView(at)
			}
		}
	}
	d.pan.install(d.scroller, d.markdown.AsPanel(), d.markdown)

	toolbar := NewToolbar()
	toolbar.AddChild(NewDefaultInfoPop())
	toolbar.AddChild(NewMarkdownGuideButton())
	// editorTopAt is the index of the rune that began the line at the top of the editor's view before a change of scale,
	// which rewraps the editor to the view's width at the new scale and so moves its lines.
	var editorTopAt int
	toolbar.AddChild(
		NewScaleField(
			minPDFDockableScale,
			maxPDFDockableScale,
			func() int { return gurps.GlobalSettings().General.InitialMarkdownUIScale },
			func() int { return d.scale },
			func(scale int) {
				if d.editorShown() {
					editorTopAt = d.editorLineAtTopOfView()
				}
				d.scale = scale
			},
			func() {
				if d.editorShown() {
					d.scrollEditorLineToTopOfView(editorTopAt)
				} else {
					rewrapMarkdownToView(d.scroller, d.markdown)
				}
			},
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
			if editToggle.Sticky {
				x, _ := d.scroller.Position()
				d.markdownLeftX = x / d.markdown.Scale().X
				d.markdownLeftAt = markdownSpotAtTopOfView(d.scroller, d.markdown)
				d.editor.SetScale(d.markdown.Scale())
				d.scroller.SetContent(d.editor, behavior.Follow, behavior.Fill)
				d.editor.RequestFocus()
				// The editor is wrapped to the view's width before the line left at the top is found.
				d.scroller.ValidateLayout()
				d.scrollEditorLineToTopOfView(d.editorLeftAt)
			} else {
				d.editorLeftAt = d.editorLineAtTopOfView()
				d.markdown.SetScale(d.editor.Scale())
				d.scroller.SetContent(d.markdown, behavior.Fill, behavior.Fill)
				// The edits and any change of scale or of the view's width were made while the markdown was not in the
				// view.
				rewrapMarkdownToView(d.scroller, d.markdown)
				d.markdown.RequestFocus()
				scale := d.markdown.Scale()
				d.scroller.SetPosition(d.markdownLeftX*scale.X, d.markdownLeftAt.y(d.markdown)*scale.Y)
			}
		}
		toolbar.AddChild(editToggle)
	}

	FinishToolbarLayout(toolbar)

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

// ScrollToAnchor scrolls the heading associated with the given anchor into view, moving the reading caret to what is
// then shown. Has no effect if the markdown is currently being edited rather than displayed, or if no matching anchor
// exists.
func (d *MarkdownDockable) ScrollToAnchor(anchor string) {
	d.ValidateLayout()
	if d.markdown.ScrollToAnchor(anchor) {
		placeReadingCaretAtTopOfView(d.scroller, d.markdown)
	}
}

func (d *MarkdownDockable) editorShown() bool {
	content := d.scroller.Content()
	return content != nil && content.AsPanel() == d.editor.AsPanel()
}

// editorLineAtTopOfView returns the index of the rune that begins the line at the top of the editor's view.
func (d *MarkdownDockable) editorLineAtTopOfView() int {
	// A unit of slack keeps a line whose top is at the top of the view from being taken for the one ending there.
	_, y := d.scroller.Position()
	return d.editor.ToSelectionIndex(geom.NewPoint(0, y/d.editor.Scale().Y+1))
}

// scrollEditorLineToTopOfView scrolls the editor's view to put the line holding the rune at index at its top, showing
// the first line with the border above it.
func (d *MarkdownDockable) scrollEditorLineToTopOfView(index int) {
	var y float32
	if index > 0 {
		y = d.editor.FromSelectionIndex(index).Y * d.editor.Scale().Y
	}
	d.scroller.SetPosition(0, y)
}

// wrapMarkdownToView keeps the markdown wrapped to fill the scroll panel's view as the view's width changes, with
// rewrapMarkdownToView. It stands in for unison.NewMarkdown(true), which wraps to the width of the markdown's parent
// without allowing for the markdown's scale, so the markdown must not be created with that. A change that leaves the
// view's width as it was, such as one of scale, calls rewrapMarkdownToView itself.
func wrapMarkdownToView(scroll *unison.ScrollPanel, markdown *unison.Markdown) {
	scroll.ContentView().FrameChangeCallback = func() { rewrapMarkdownToView(scroll, markdown) }
}

// rewrapMarkdownToView wraps the markdown afresh to fill the scroll panel's view (see markdownWidthInView) and lays the
// scroll panel out again to fit, keeping what was at the top of the view there (see markdownSpot).
func rewrapMarkdownToView(scroll *unison.ScrollPanel, markdown *unison.Markdown) {
	if width := markdownWidthInView(scroll, markdown); width > 0 {
		x, _ := scroll.Position()
		spot := markdownSpotAtTopOfView(scroll, markdown)
		markdown.SetContentBytes(markdown.ContentBytes(), width)
		scroll.MarkForLayoutAndRedraw()
		scroll.ValidateLayout()
		scroll.SetPosition(x, spot.y(markdown)*markdown.Scale().Y)
	}
}

// markdownSpotAtTopOfView returns what is at the top of the scroll panel's view of the markdown.
func markdownSpotAtTopOfView(scroll *unison.ScrollPanel, markdown *unison.Markdown) markdownSpot {
	// The top needs no spot, and a document yet to be laid out, as when the view is first given a width, has none.
	if _, y := scroll.Position(); y > 0 {
		return markdownSpotAt(markdown, y/markdown.Scale().Y)
	}
	return markdownSpot{}
}

// setMarkdownContentInView sets the markdown's content, wrapped to fill the scroll panel's view, or to
// unison.DefaultMarkdownWidth while the markdown is not shown there or the view has yet to be given a width, which a
// window of its own is then sized to fit.
func setMarkdownContentInView(scroll *unison.ScrollPanel, markdown *unison.Markdown, content []byte) {
	width := markdownWidthInView(scroll, markdown)
	if width == 0 {
		width = unison.DefaultMarkdownWidth
	}
	markdown.SetContentBytes(content, width)
}

// markdownWidthInView returns the width to wrap the markdown to so that, at its scale, it fills the width of the scroll
// panel's view, less the slop unison.Markdown leaves when it wraps to its parent. Returns 0 when the markdown is not
// the scroll panel's content or the view has yet to be given a width.
func markdownWidthInView(scroll *unison.ScrollPanel, markdown *unison.Markdown) float32 {
	if content := scroll.Content(); content == nil || content.AsPanel() != markdown.AsPanel() {
		return 0
	}
	width := scroll.ContentView().ContentRect(false).Width
	if width <= 0 {
		return 0
	}
	width = width/markdown.Scale().X - markdown.Slop
	if border := markdown.Border(); border != nil {
		width -= border.Insets().Width()
	}
	// SetContentBytes takes a width below 1 to mean the parent's, which leaves the scale out.
	return max(width, 1)
}

// placeReadingCaretAtTopOfView puts the markdown's reading caret at the start of the first block of text whose top the
// scroll panel shows, without moving the view, so that the keys that move the caret read on from what is shown rather
// than scrolling back to wherever the caret was. The caret stays where it is when the markdown is not the scroll
// panel's content or no block's top is shown.
func placeReadingCaretAtTopOfView(scroll *unison.ScrollPanel, markdown *unison.Markdown) {
	if content := scroll.Content(); content == nil || content.AsPanel() != markdown.AsPanel() {
		return
	}
	view := scroll.ContentView()
	shown := view.RectTo(view.ContentRect(false), markdown.AsPanel())
	x, y := scroll.Position()
	placeReadingCaretInFirstBlockShown(markdown.AsPanel(), markdown.AsPanel(), shown)
	// Placing the caret brings its line wholly into view, which may scroll a block whose top was allowed slack below.
	scroll.SetPosition(x, y)
}

// placeReadingCaretInFirstBlockShown puts the reading caret at the start of the first block within parent, in document
// order, whose top lies within shown, a rect in the markdown's coordinates. Returns true once it has.
func placeReadingCaretInFirstBlockShown(parent, markdown *unison.Panel, shown geom.Rect) bool {
	for _, child := range parent.Children() {
		// A block of text takes a selection within its own text from an assistive technology, so an empty one at its
		// start puts the caret there; nothing else within a markdown takes one. A unit of slack allows for rounding in
		// the scroll that brought a heading to the top of the view.
		if actor, ok := child.Self.(unison.AccessibilityActor); ok {
			top := child.RectTo(child.ContentRect(true), markdown).Y
			if top >= shown.Y-1 && top < shown.Bottom() &&
				actor.PerformAccessibilityAction(accessibility.ActionRequest{Action: accessibility.SetTextSelection}) {
				return true
			}
		}
		if placeReadingCaretInFirstBlockShown(child, markdown, shown) {
			return true
		}
	}
	return false
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
	invalidateDisplayData()
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
