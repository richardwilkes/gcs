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
	"bytes"
	"fmt"
	"strings"

	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/desktop"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xmath"
	"github.com/richardwilkes/unison"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	tableAST "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// DefaultMarkdownWidth is the default maximum width to use, roughly equivalent to a page at 100dpi.
const DefaultMarkdownWidth = 8 * 100

// Markdown provides a simple markdown display widget. It is currently *very* limited in what it can display.
type Markdown struct {
	unison.Panel
	node       ast.Node
	content    []byte
	block      *unison.Panel
	textRow    *unison.Panel
	text       *unison.Text
	decoration *unison.TextDecoration
	imgCache   map[string]*unison.Image
	index      int
	maxWidth   float32
	ordered    bool
	isHeader   bool
}

// NewMarkdown creates a new markdown widget.
func NewMarkdown() *Markdown {
	m := &Markdown{imgCache: make(map[string]*unison.Image)}
	m.Self = &m
	m.ParentChangedCallback = func() {
		// TODO: This is fragile... need to have a better way to achieve dynamic sizing
		if m.Parent() != nil {
			m.Parent().FrameChangeCallback = func() {
				w := m.Parent().ContentRect(false).Width - 20
				if border := m.Border(); border != nil {
					insets := border.Insets()
					w -= insets.Width()
				}
				m.SetContentBytes(m.content, w)
			}
		}
	}
	return m
}

// SetContent replaces the current markdown content.
func (m *Markdown) SetContent(content string, maxWidth float32) {
	m.SetContentBytes([]byte(content), maxWidth)
}

// SetContentBytes replaces the current markdown content.
func (m *Markdown) SetContentBytes(content []byte, maxWidth float32) {
	if maxWidth < 1 {
		if p := m.Parent(); p != nil {
			maxWidth = p.ContentRect(false).Width
		} else {
			maxWidth = DefaultMarkdownWidth
		}
	}
	if m.maxWidth == maxWidth && bytes.Equal(m.content, content) {
		return
	}
	m.RemoveAllChildren()
	fd := unison.LabelFont.Descriptor()
	fd.Weight = unison.NormalFontWeight
	fd.Slant = unison.NoSlant
	fd.Spacing = unison.StandardSpacing
	m.maxWidth = maxWidth
	m.content = content
	m.block = m.AsPanel()
	m.textRow = nil
	m.text = nil
	m.decoration = &unison.TextDecoration{
		Font:       fd.Font(),
		Foreground: unison.OnBackgroundColor,
	}
	m.index = 0
	m.ordered = false
	spacing := xmath.Ceil(fd.Size)
	m.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: spacing,
	})
	m.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HGrab:  true,
		HAlign: unison.FillAlignment,
	})
	m.SetBorder(unison.NewEmptyBorder(unison.NewUniformInsets(spacing)))
	m.node = goldmark.New(goldmark.WithExtensions(extension.GFM)).Parser().Parse(text.NewReader(m.content))
	m.walk(m.node)
	m.MarkForLayoutAndRedraw()
}

func (m *Markdown) walk(node ast.Node) {
	save := m.node
	m.node = node
	switch m.node.Kind() {
	// Block types
	case ast.KindDocument:
		m.processChildren()
	case ast.KindTextBlock, ast.KindParagraph:
		m.processParagraphOrTextBlock()
	case ast.KindHeading:
		m.processHeading()
	case ast.KindThematicBreak:
		m.processThematicBreak()
	case ast.KindCodeBlock, ast.KindFencedCodeBlock:
		m.processCodeBlock()
	case ast.KindBlockquote:
		m.processBlockquote()
	case ast.KindList:
		m.processList()
	case ast.KindListItem:
		m.processListItem()
	case ast.KindHTMLBlock:
		// Ignore
	case tableAST.KindTable:
		m.processTable()
	case tableAST.KindTableHeader:
		m.processTableHeader()
	case tableAST.KindTableRow:
		m.processTableRow()
	case tableAST.KindTableCell:
		m.processTableCell()

	// Inline types
	case ast.KindText:
		m.processText()
	case ast.KindEmphasis:
		m.processEmphasis()
	case ast.KindCodeSpan:
		m.processCodeSpan()
	case ast.KindRawHTML:
		m.processRawHTML()
	case ast.KindString:
		m.processString()
	case ast.KindLink:
		m.processLink()
	case ast.KindImage:
		m.processImage()
	case ast.KindAutoLink:
		m.processAutoLink()

	default:
		jot.Infof("unhandled markdown element: %v", m.node.Kind())
	}
	m.node = save
}

func (m *Markdown) processChildren() {
	for child := m.node.FirstChild(); child != nil; child = child.NextSibling() {
		m.walk(child)
	}
}

func (m *Markdown) processParagraphOrTextBlock() {
	p := unison.NewPanel()
	p.SetLayout(&unison.FlexLayout{Columns: 1})
	save := m.block
	m.block.AddChild(p)
	m.block = p
	m.text = unison.NewText("", m.decoration)
	m.processChildren()
	m.finishTextRow()
	m.block = save
}

func (m *Markdown) processHeading() {
	if heading, ok := m.node.(*ast.Heading); ok {
		saveDec := m.decoration
		saveBlock := m.block
		m.decoration = m.decoration.Clone()
		fd := m.decoration.Font.Descriptor()
		fd.Weight = unison.BoldFontWeight
		fd.Slant = unison.NoSlant
		fd.Spacing = unison.StandardSpacing
		fd.Size = unison.LabelFont.Size()
		switch heading.Level {
		case 1:
			fd.Size *= 2.5
		case 2:
			fd.Size *= 2
		case 3:
			fd.Size *= 1.75
		case 4:
			fd.Size *= 1.5
		case 5:
			fd.Size *= 1.25
		default:
			// Remaining are normal sized
		}
		m.decoration.Font = fd.Font()

		p := unison.NewPanel()
		p.SetLayout(&unison.FlexLayout{Columns: 1})
		m.block.AddChild(p)
		m.block = p
		m.text = unison.NewText("", m.decoration)
		m.processChildren()
		m.finishTextRow()
		m.decoration = saveDec
		m.block = saveBlock
	}
}

func (m *Markdown) processThematicBreak() {
	hr := unison.NewSeparator()
	hr.SetLayoutData(&unison.FlexLayoutData{
		HGrab:  true,
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
	})
	m.block.AddChild(hr)
}

func (m *Markdown) processCodeBlock() {
	saveDec := m.decoration
	saveBlock := m.block
	m.decoration = m.decoration.Clone()
	fd := unison.MonospacedFont.Descriptor()
	fd.Size = unison.LabelFont.Size()
	m.decoration.Font = fd.Font()
	m.decoration.Foreground = unison.OnContentColor

	p := unison.NewPanel()
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, unison.ContentColor.Paint(gc, rect, unison.Fill))
	}
	p.SetLayout(&unison.FlexLayout{Columns: 1})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	p.SetBorder(unison.NewEmptyBorder(unison.NewUniformInsets(xmath.Ceil(fd.Size / 2))))
	m.block.AddChild(p)
	m.block = p
	lines := m.node.Lines()
	count := lines.Len()
	for i := 0; i < count; i++ {
		segment := lines.At(i)
		label := unison.NewRichLabel()
		label.Text = unison.NewText(string(segment.Value(m.content)), m.decoration)
		p.AddChild(label)
	}
	m.text = nil
	m.textRow = nil
	m.decoration = saveDec
	m.block = saveBlock
}

func (m *Markdown) processBlockquote() {
	saveDec := m.decoration
	saveBlock := m.block
	m.decoration = m.decoration.Clone()
	m.decoration.Foreground = unison.OnContentColor

	p := unison.NewPanel()
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, unison.ContentColor.Paint(gc, rect, unison.Fill))
	}
	p.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: xmath.Ceil(unison.LabelFont.Size()),
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	p.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.SelectionColor, 0, unison.Insets{Left: 2}, false),
		unison.NewEmptyBorder(unison.NewUniformInsets(xmath.Ceil(m.decoration.Font.Size()/2)))))
	m.block.AddChild(p)
	m.block = p
	m.text = unison.NewText("", m.decoration)
	m.processChildren()
	m.finishTextRow()
	m.decoration = saveDec
	m.block = saveBlock
}

func (m *Markdown) processList() {
	if list, ok := m.node.(*ast.List); ok {
		saveIndex := m.index
		saveOrdered := m.ordered
		saveBlock := m.block
		m.index = list.Start
		m.ordered = list.IsOrdered()
		p := unison.NewPanel()
		p.SetLayout(&unison.FlexLayout{
			Columns:  2,
			HSpacing: m.decoration.Font.SimpleWidth(" "),
		})
		p.SetLayoutData(&unison.FlexLayoutData{
			HAlign: unison.FillAlignment,
			HGrab:  true,
		})
		m.block.AddChild(p)
		m.block = p
		m.processChildren()
		m.index = saveIndex
		m.ordered = saveOrdered
		m.block = saveBlock
	}
}

func (m *Markdown) processListItem() {
	var bullet string
	if m.ordered {
		bullet = fmt.Sprintf("%d.", m.index)
		m.index++
	} else {
		bullet = "•"
	}
	label := unison.NewRichLabel()
	label.Text = unison.NewText(bullet, m.decoration)
	label.SetLayoutData(&unison.FlexLayoutData{HAlign: unison.EndAlignment})
	m.block.AddChild(label)

	saveBlock := m.block
	p := unison.NewPanel()
	p.SetLayout(&unison.FlexLayout{Columns: 1})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	m.block.AddChild(p)
	m.block = p
	m.processChildren()
	m.block = saveBlock
}

func (m *Markdown) processTable() {
	if table, ok := m.node.(*tableAST.Table); ok {
		if len(table.Alignments) != 0 {
			saveBlock := m.block
			p := unison.NewPanel()
			p.SetBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.NewUniformInsets(1), false))
			p.SetLayout(&unison.FlexLayout{Columns: len(table.Alignments)})
			m.block.AddChild(p)
			m.block = p
			m.processChildren()
			m.block = saveBlock
		}
	}
}

func (m *Markdown) processTableHeader() {
	m.isHeader = true
	m.processChildren()
	m.isHeader = false
}

func (m *Markdown) processTableRow() {
	m.processChildren()
}

func (m *Markdown) processTableCell() {
	if cell, ok := m.node.(*tableAST.TableCell); ok {
		saveDec := m.decoration
		saveBlock := m.block
		align := m.alignment(cell.Alignment)
		m.decoration = m.decoration.Clone()
		if m.isHeader {
			fd := m.decoration.Font.Descriptor()
			fd.Weight = unison.BoldFontWeight
			fd.Slant = unison.NoSlant
			fd.Spacing = unison.StandardSpacing
			m.decoration.Font = fd.Font()
			if align != unison.EndAlignment {
				align = unison.MiddleAlignment
			}
		}
		p := unison.NewPanel()
		p.SetBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.NewUniformInsets(1), false))
		p.SetLayout(&unison.FlexLayout{
			Columns: 1,
			HAlign:  align,
		})
		p.SetLayoutData(&unison.FlexLayoutData{
			HAlign: unison.FillAlignment,
			VAlign: unison.FillAlignment,
			VGrab:  true,
		})
		m.block.AddChild(p)

		inner := unison.NewPanel()
		inner.SetBorder(unison.NewEmptyBorder(unison.StdInsets()))
		inner.SetLayout(&unison.FlexLayout{
			Columns: 1,
			HAlign:  align,
		})
		inner.SetLayoutData(&unison.FlexLayoutData{
			HAlign: align,
		})
		p.AddChild(inner)

		m.block = inner
		m.text = unison.NewText("", m.decoration)
		m.processChildren()
		m.finishTextRow()
		m.decoration = saveDec
		m.block = saveBlock
	}
}

func (m *Markdown) alignment(alignment tableAST.Alignment) unison.Alignment {
	switch alignment {
	case tableAST.AlignLeft:
		return unison.StartAlignment
	case tableAST.AlignRight:
		return unison.EndAlignment
	case tableAST.AlignCenter:
		return unison.MiddleAlignment
	default:
		return unison.StartAlignment
	}
}

func (m *Markdown) processText() {
	if t, ok := m.node.(*ast.Text); ok {
		b := util.UnescapePunctuations(t.Text(m.content))
		b = util.ResolveNumericReferences(b)
		str := string(util.ResolveEntityNames(b))
		if t.SoftLineBreak() {
			str += " "
		}
		m.text.AddString(str, m.decoration)
		if t.HardLineBreak() {
			m.issueLineBreak()
		}
	}
}

func (m *Markdown) processEmphasis() {
	if emphasis, ok := m.node.(*ast.Emphasis); ok {
		save := m.decoration
		m.decoration = save.Clone()
		fd := m.decoration.Font.Descriptor()
		if emphasis.Level == 1 {
			fd.Slant = unison.ItalicSlant
		} else {
			fd.Weight = unison.BoldFontWeight
		}
		m.decoration.Font = fd.Font()
		m.processChildren()
		m.decoration = save
	}
}

func (m *Markdown) processCodeSpan() {
	save := m.decoration
	m.decoration = save.Clone()
	m.decoration.Foreground = unison.OnContentColor
	m.decoration.Background = unison.ContentColor
	fd := unison.MonospacedFont.Descriptor()
	fd.Size = unison.LabelFont.Size()
	m.decoration.Font = fd.Font()
	m.processChildren()
	m.decoration = save
}

func (m *Markdown) processRawHTML() {
	if raw, ok := m.node.(*ast.RawHTML); ok {
		count := raw.Segments.Len()
		for i := 0; i < count; i++ {
			segment := raw.Segments.At(i)
			switch txt.CollapseSpaces(strings.ToLower(string(segment.Value(m.content)))) {
			case "<br>", "<br/>", "<br />":
				m.issueLineBreak()
			case "<hr>", "<hr/>", "<hr />":
				m.issueLineBreak()
				m.processThematicBreak()
				m.issueLineBreak()
			}
		}
	}
}

func (m *Markdown) processString() {
	if t, ok := m.node.(*ast.String); ok {
		b := util.UnescapePunctuations(t.Value)
		b = util.ResolveNumericReferences(b)
		str := string(util.ResolveEntityNames(b))
		m.text.AddString(str, m.decoration)
	}
}

func (m *Markdown) processLink() {
	if link, ok := m.node.(*ast.Link); ok {
		m.flushText()
		p := m.createLink(string(link.Text(m.content)), string(link.Destination), string(link.Title))
		m.addToTextRow(p)
	}
}

func (m *Markdown) createLink(label, target, tooltip string) *unison.RichLabel {
	onFg := model.OnLinkColor
	onBg := model.LinkColor
	dec := m.decoration.Clone()
	dec.Underline = true
	p := unison.NewRichLabel()
	p.Text = unison.NewText(label, dec)
	if target != "" {
		in := false
		p.MouseDownCallback = func(where unison.Point, button, clickCount int, mod unison.Modifiers) bool {
			p.Text.AdjustDecorations(func(decoration *unison.TextDecoration) {
				decoration.Foreground = onFg
				decoration.Background = onBg
			})
			p.MarkForRedraw()
			in = true
			return true
		}
		p.MouseDragCallback = func(where unison.Point, button int, mod unison.Modifiers) bool {
			now := p.ContentRect(true).ContainsPoint(where)
			if now != in {
				in = now
				p.Text.AdjustDecorations(func(decoration *unison.TextDecoration) {
					if in {
						decoration.Foreground = onFg
						decoration.Background = onBg
					} else {
						decoration.Foreground = dec.Foreground
						decoration.Background = dec.Background
					}
				})
				p.MarkForRedraw()
			}
			return true
		}
		p.MouseUpCallback = func(where unison.Point, button int, mod unison.Modifiers) bool {
			p.Text.AdjustDecorations(func(decoration *unison.TextDecoration) {
				decoration.Foreground = dec.Foreground
				decoration.Background = dec.Background
			})
			p.MarkForRedraw()
			if p.ContentRect(true).ContainsPoint(where) {
				if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
					if err := desktop.Open(target); err != nil {
						unison.ErrorDialogWithError(i18n.Text("Opening the link failed"), err)
					}
				}
				// TODO: Support other types
			}
			return true
		}
	}
	if tooltip != "" {
		p.Tooltip = unison.NewTooltipWithText(tooltip)
	}
	return p
}

func (m *Markdown) processImage() {
	if image, ok := m.node.(*ast.Image); ok {
		m.flushText()
		target := string(image.Destination)
		var img *unison.Image
		if img, ok = m.imgCache[target]; !ok {
			var err error
			if img, err = unison.NewImageFromFilePathOrURL(target, 1); err != nil {
				jot.Error(errs.Wrap(err))
			} else {
				m.imgCache[target] = img
			}
		}
		label := unison.NewLabel()
		if img == nil {
			size := xmath.Max(m.decoration.Font.Size(), 24)
			label.Drawable = &unison.DrawableSVG{
				SVG:  svg.BrokenImage,
				Size: unison.NewSize(size, size),
			}
		} else {
			label.Drawable = img
		}
		primary := string(image.Text(m.content))
		secondary := string(image.Title)
		if primary == "" && secondary != "" {
			primary = secondary
			secondary = ""
		}
		if primary != "" {
			if secondary != "" {
				label.Tooltip = unison.NewTooltipWithSecondaryText(primary, secondary)
			} else {
				label.Tooltip = unison.NewTooltipWithText(primary)
			}
		}
		m.addToTextRow(label)
	}
}

func (m *Markdown) processAutoLink() {
	if link, ok := m.node.(*ast.AutoLink); ok {
		m.flushText()
		url := string(link.URL(m.content))
		p := m.createLink(url, url, "")
		m.addToTextRow(p)
	}
}

func (m *Markdown) addToTextRow(p unison.Paneler) {
	if m.textRow == nil {
		m.textRow = unison.NewPanel()
		m.textRow.SetLayout(&unison.FlowLayout{})
		m.textRow.SetLayoutData(&unison.FlexLayoutData{
			HAlign: unison.FillAlignment,
			HGrab:  true,
		})
		m.block.AddChild(m.textRow)
	}
	m.textRow.AddChild(p)
}

func (m *Markdown) addLabelToTextRow(t *unison.Text) {
	label := unison.NewRichLabel()
	label.Text = t
	m.addToTextRow(label)
}

func (m *Markdown) issueLineBreak() {
	m.flushText()
	m.addToTextRow(unison.NewRichLabel())
	m.textRow = nil
}

func (m *Markdown) flushText() {
	if len(m.text.Runes()) != 0 {
		remaining := m.maxWidth
		if m.textRow != nil {
			_, prefSize, _ := m.textRow.Sizes(unison.Size{Width: m.maxWidth})
			remaining -= prefSize.Width
		}
		min := m.decoration.Font.SimpleWidth("W")
		if remaining < min {
			// Remaining space is less than the width of a W, so go to the next line
			m.addToTextRow(unison.NewRichLabel())
			m.textRow = nil
			remaining = m.maxWidth
		}
		if remaining < m.text.Width() {
			// Remaining space isn't large enough for the text we have, so put a chunk that will fit on this line, then
			// go to the next line
			part := m.text.BreakToWidth(remaining)[0]
			m.text = m.text.Slice(len(part.Runes()), len(m.text.Runes()))
			m.addLabelToTextRow(part)
			m.textRow = nil
			// Now break the remaining text up to the max width size and add each line
			for _, part = range m.text.BreakToWidth(m.maxWidth) {
				m.textRow = nil
				m.addLabelToTextRow(part)
			}
		} else {
			m.addLabelToTextRow(m.text)
		}
		m.text = unison.NewText("", m.decoration)
	}
}

func (m *Markdown) finishTextRow() {
	m.flushText()
	m.text = nil
	m.textRow = nil
}
