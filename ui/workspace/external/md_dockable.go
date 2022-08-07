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

package external

import (
	"fmt"
	"os"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/gcs/v5/ui/workspace"
	"github.com/richardwilkes/toolbox/i18n"
	xfs "github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
)

const markdownContentOnlyPrefix = "//////////"

var (
	_ workspace.FileBackedDockable = &MarkdownDockable{}
	_ unison.TabCloser             = &MarkdownDockable{}
)

// MarkdownDockable holds the view for an image file.
type MarkdownDockable struct {
	unison.Panel
	path       string
	title      string
	scroll     *unison.ScrollPanel
	markdown   *widget.Markdown
	scaleField *widget.PercentageField
	scale      int
}

// ShowReleaseNotesMarkdown attempts to show the given markdown content in a dockable.
func ShowReleaseNotesMarkdown(title, content string) {
	ws := workspace.FromWindowOrAny(nil)
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
	workspace.DisplayNewDockable(nil, d)
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
	d.KeyDownCallback = d.keyDown
	d.SetLayout(&unison.FlexLayout{Columns: 1})

	d.markdown = widget.NewMarkdown()
	d.markdown.MouseWheelCallback = d.mouseWheel
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

	scaleTitle := i18n.Text("Scale")
	d.scaleField = widget.NewPercentageField(nil, "", scaleTitle,
		func() int { return d.scale },
		func(v int) {
			d.scale = v
			d.markdown.SetScale(float32(d.scale) / 100)
			d.scroll.Sync()
		}, minPDFDockableScale, maxPDFDockableScale, false, false)
	d.scaleField.Tooltip = unison.NewTooltipWithText(scaleTitle)

	toolbar := unison.NewPanel()
	toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.Insets{Bottom: 1},
		false), unison.NewEmptyBorder(unison.StdInsets())))
	toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	toolbar.AddChild(d.scaleField)
	toolbar.SetLayout(&unison.FlexLayout{
		Columns:  len(toolbar.Children()),
		HSpacing: unison.StdHSpacing,
	})

	d.AddChild(toolbar)
	d.AddChild(d.scroll)

	return d, nil
}

func (d *MarkdownDockable) mouseWheel(_, delta unison.Point, mod unison.Modifiers) bool {
	if !mod.OptionDown() {
		return false
	}
	scale := d.scale + int(delta.Y*deltaPDFDockableScale)
	if scale < minPDFDockableScale {
		scale = minPDFDockableScale
	} else if scale > maxPDFDockableScale {
		scale = maxPDFDockableScale
	}
	widget.SetFieldValue(d.scaleField.Field, d.scaleField.Format(scale))
	return true
}

func (d *MarkdownDockable) keyDown(keyCode unison.KeyCode, _ unison.Modifiers, _ bool) bool {
	scale := d.scale
	switch keyCode {
	case unison.KeyQ:
		scale = 25
	case unison.KeyH:
		scale = 50
	case unison.KeyT:
		scale = 75
	case unison.Key1:
		scale = 100
	case unison.Key2:
		scale = 200
	case unison.Key3:
		scale = 300
	case unison.KeyMinus:
		scale -= deltaPDFDockableScale
		if scale < minPDFDockableScale {
			scale = minPDFDockableScale
		}
	case unison.KeyEqual:
		scale += deltaPDFDockableScale
		if scale > maxPDFDockableScale {
			scale = maxPDFDockableScale
		}
	default:
		return false
	}
	if d.scale != scale {
		widget.SetFieldValue(d.scaleField.Field, d.scaleField.Format(scale))
	}
	return true
}

// TitleIcon implements workspace.FileBackedDockable
func (d *MarkdownDockable) TitleIcon(suggestedSize unison.Size) unison.Drawable {
	if strings.HasPrefix(d.path, markdownContentOnlyPrefix) {
		return &unison.DrawableSVG{
			SVG:  res.MarkdownFileSVG,
			Size: suggestedSize,
		}
	}
	return &unison.DrawableSVG{
		SVG:  library.FileInfoFor(d.BackingFilePath()).SVG,
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
