package widget

import (
	"strings"

	"github.com/richardwilkes/unison"
)

// DefaultMarkdownWidth is the default maximum width to use, roughly equivalent to a page at 100dpi.
const DefaultMarkdownWidth = 8 * 100

// Markdown provides a simple markdown display widget. It is currently *very* limited in what it can display.
type Markdown struct {
	unison.Panel
	content  string
	maxWidth int
}

// NewMarkdown creates a new markdown widget.
func NewMarkdown() *Markdown {
	m := &Markdown{maxWidth: DefaultMarkdownWidth}
	m.Self = &m
	m.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
	})
	m.SetBorder(unison.NewEmptyBorder(unison.Insets{
		Top:    unison.StdVSpacing,
		Left:   unison.StdHSpacing,
		Bottom: unison.StdVSpacing,
		Right:  unison.StdHSpacing,
	}))
	return m
}

// SetContent replaces the current markdown content.
func (m *Markdown) SetContent(content string, maxWidth int) {
	if maxWidth < 1 {
		maxWidth = DefaultMarkdownWidth
	}
	if m.maxWidth == maxWidth && m.content == content {
		return
	}
	m.RemoveAllChildren()
	m.maxWidth = maxWidth
	m.content = content
	hdrFD := unison.DefaultLabelTheme.Font.Descriptor()
	hdrFD.Weight = unison.BoldFontWeight
	sizes := []float32{hdrFD.Size * 2, hdrFD.Size * 3 / 2, hdrFD.Size * 5 / 4, hdrFD.Size}
	otherDecoration := &unison.TextDecoration{Font: unison.DefaultLabelTheme.Font}
	bulletWidth := float32(maxWidth) - (otherDecoration.Font.SimpleWidth("•") + unison.StdHSpacing)
	for _, line := range strings.Split(content, "\n") {
		found := false
		for i := range sizes {
			if strings.HasPrefix(line, strings.Repeat("#", i+1)+" ") {
				hdrFD.Size = sizes[i]
				font := hdrFD.Font()
				block := unison.NewTextWrappedLines(line[i+2:], &unison.TextDecoration{Font: font}, float32(maxWidth))
				for j, chunk := range block {
					label := unison.NewLabel()
					label.Font = font
					label.Text = chunk.String()
					if j == len(block)-1 {
						label.SetBorder(unison.NewEmptyBorder(unison.Insets{Bottom: sizes[i] / 2}))
					}
					label.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
					m.AddChild(label)
				}
				found = true
				break
			}
		}
		if found {
			continue
		}
		switch {
		case line == "---":
			hr := unison.NewSeparator()
			hr.SetBorder(unison.NewEmptyBorder(unison.NewVerticalInsets(otherDecoration.Font.Size())))
			hr.SetLayoutData(&unison.FlexLayoutData{
				HSpan:  2,
				HAlign: unison.FillAlignment,
				HGrab:  true,
			})
			m.AddChild(hr)
		case strings.HasPrefix(line, "- "):
			block := unison.NewTextWrappedLines(line[2:], otherDecoration, bulletWidth)
			for j, chunk := range block {
				label := unison.NewLabel()
				label.Font = otherDecoration.Font
				if j == 0 {
					label.Text = "•"
				}
				m.AddChild(label)

				label = unison.NewLabel()
				label.Font = otherDecoration.Font
				label.Text = chunk.String()
				m.AddChild(label)
			}
		default:
			block := unison.NewTextWrappedLines(line, otherDecoration, float32(maxWidth))
			for _, chunk := range block {
				label := unison.NewLabel()
				label.Font = otherDecoration.Font
				label.Text = chunk.String()
				label.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
				m.AddChild(label)
			}
		}
	}
	m.MarkForLayoutAndRedraw()
}
