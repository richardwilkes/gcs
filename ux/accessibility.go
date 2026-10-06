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
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
)

// speakAs gives static text a name for a screen reader to speak in place of its drawn text; an empty name restores the
// drawn text. macOS speaks a label whose name differs from its text as both, so the text is withheld from the node
// while a name is set. It replaces the panel's Accessibility.Callback.
func speakAs(p unison.Paneler, name string) {
	panel := p.AsPanel()
	panel.Accessibility.Name = name
	panel.Accessibility.Callback = func(node *accessibility.Node) {
		if panel.Accessibility.Name != "" {
			node.Text = nil
		}
	}
}

// newLink returns a unison link that also opens on an assistive technology's Press. Unison answers a Press on a link
// with a synthesized click, but the table's space key at cell level presses a cell's content without one (a click on a
// field would start an unwanted edit), and a link in a cell never takes the focus its own keys need. Handling the Press
// directly makes space on a cell holding a link open the link rather than fall back to the table's row shortcut, which
// opens the row's editor.
func newLink(title, tooltip, target string, theme *unison.LinkTheme, clickHandler func(unison.Paneler, string)) *unison.Label {
	link := unison.NewLink(title, tooltip, target, theme, clickHandler)
	link.Accessibility.ActionCallback = func(req accessibility.ActionRequest) bool {
		if req.Action != accessibility.Press || clickHandler == nil {
			return false
		}
		clickHandler(link, target)
		return true
	}
	return link
}

// addAccessibilityCallback runs adjust after the panel's existing Accessibility.Callback. Anything that later replaces
// the callback, such as speakAs, discards adjust, so call this last.
func addAccessibilityCallback(p unison.Paneler, adjust func(node *accessibility.Node)) {
	panel := p.AsPanel()
	prior := panel.Accessibility.Callback
	if prior == nil {
		panel.Accessibility.Callback = adjust
		return
	}
	panel.Accessibility.Callback = func(node *accessibility.Node) {
		prior(node)
		adjust(node)
	}
}

// tabStopForReading keeps a control out of the Tab order except while a screen reader is in use and the general setting
// that puts static text and disabled controls in the Tab order is on. That setting drives unison.SetFocusForReading,
// which does not add a control made unfocusable.
//
// Unison does not ask the application whether a panel takes the focus, so the control's focusability is updated as the
// window is described, which happens only while a screen reader is listening and again when the setting changes. Once
// the screen reader is gone nothing describes the window, so the control leaves the Tab order when it next loses the
// focus. Its Accessibility.Callback and LostFocusCallback are wrapped, so call this after setting them.
func tabStopForReading(p unison.Paneler) {
	panel := p.AsPanel()
	panel.SetFocusable(false)
	reading := false
	addAccessibilityCallback(panel, func(node *accessibility.Node) {
		if want := unison.FocusForReading(); want != reading {
			reading = want
			panel.SetFocusable(want)
		}
		// The node was filled in before any SetFocusable call above.
		if node.Focusable = panel.Focusable(); node.Focusable {
			node.Actions = node.Actions.With(accessibility.Focus)
		} else {
			node.Actions = node.Actions.Without(accessibility.Focus)
		}
	})
	lost := panel.LostFocusCallback
	panel.LostFocusCallback = func() {
		if reading && (!unison.IsAccessibilityActive() || !unison.FocusForReading()) {
			reading = false
			panel.SetFocusable(false)
		}
		if lost != nil {
			lost()
		}
	}
}

// describeWithHint puts the text of the label that follows a control, such as its units, at the front of the control's
// accessible description, so a screen reader speaks it with the control rather than as separate static text. hint is
// called each time the control is described.
func describeWithHint(control unison.Paneler, hint func() string) {
	addAccessibilityCallback(control, func(node *accessibility.Node) {
		text := strings.TrimSpace(hint())
		switch {
		case text == "" || text == node.Name || text == node.Description:
		case node.Description == "":
			node.Description = text
		default:
			node.Description = text + ". " + node.Description
		}
	})
}

// ignoreEmptyStaticText hides static text from a screen reader when it draws nothing, has no name and labels no
// control, as unison does for a label whose role it derives itself. text is what the panel draws.
func ignoreEmptyStaticText(b *unison.AccessibilityBuilder, text string) {
	node := b.Node()
	if text == "" && node.Name == "" && xreflect.IsNil(b.Panel().Accessibility.LabeledBy) {
		node.Ignored = true
	}
}

// announceColumnsOnFocus has a table announce its column titles when the keyboard focus arrives on it at row level. The
// header is never a tab stop, so nothing else reads the titles on the way in.
//
// The row-level check waits for the focus change to settle (see announceAfterFocusSettles), since a mouse press
// requests the focus before it returns the cursor to row level. Arriving at cell level means returning from editing a
// cell, returning to a cell the person had already moved to, or a screen reader focusing a cell it named itself, so
// nothing is announced then.
func announceColumnsOnFocus[T unison.TableRowConstraint[T]](table *unison.Table[T], header *unison.TableHeader[T]) {
	gained := table.GainedFocusCallback
	table.GainedFocusCallback = func() {
		if gained != nil {
			gained()
		}
		if !unison.IsAccessibilityActive() {
			return
		}
		announceAfterFocusSettles(func() string {
			if !table.Focused() || table.LeadColumnIndex() >= 0 {
				return ""
			}
			return columnsAnnouncement(header)
		})
	}
}

// announceAfterFocusSettles announces what text returns, if anything, once the pending focus change has been published.
// A focus change reaches the screen reader when the window next draws, but an announcement goes out at once, so one
// made from a GainedFocusCallback would precede the focus change and could be cut off by it. Each pass of the event
// loop runs one queued task and then draws, so a task queued from within a queued task runs after that draw.
func announceAfterFocusSettles(text func() string) {
	unison.InvokeTask(func() {
		unison.InvokeTask(func() {
			if s := text(); s != "" {
				unison.AnnounceForAccessibility(s)
			}
		})
	})
}

func columnsAnnouncement[T unison.TableRowConstraint[T]](header *unison.TableHeader[T]) string {
	titles := make([]string, 0, len(header.ColumnHeaders))
	for _, one := range header.ColumnHeaders {
		if title := columnHeaderTitle[T](one); title != "" {
			titles = append(titles, title)
		}
	}
	if len(titles) == 0 {
		return ""
	}
	return i18n.Text("Columns: %s", strings.Join(titles, i18n.Text(", ")))
}

// columnHeaderTitle returns a column header's title, chosen as unison's unexported axColumnHeaderName names the header
// for a screen reader.
func columnHeaderTitle[T unison.TableRowConstraint[T]](header unison.TableColumnHeader[T]) string {
	panel := header.AsPanel()
	if panel.Accessibility.Name != "" {
		return panel.Accessibility.Name
	}
	if label := headerLabel[T](header); label != nil {
		if text := label.String(); text != "" {
			return text
		}
	} else if text := labelText(panel); text != "" {
		return text
	}
	return tooltipTextOf(panel)
}

// labelBackedHeader is implemented by this package's column headers that embed a unison.Label.
type labelBackedHeader interface {
	headerLabel() *unison.Label
}

// headerLabel returns the label a column header is built around, or nil. Unison recognizes such a header by an
// unexported method, so the header types GCS uses are matched here instead. Checking for String or SetTitle would not
// work, since every panel has a String that returns its type name.
func headerLabel[T unison.TableRowConstraint[T]](header unison.TableColumnHeader[T]) *unison.Label {
	switch h := header.(type) {
	case *unison.DefaultTableColumnHeader[T]:
		return h.Label
	case labelBackedHeader:
		return h.headerLabel()
	default:
		return nil
	}
}

// labelText returns the text a panel is named by, as unison's unexported axLabelText does: its accessible name, else
// its text if it is a Label, else its children's text gathered the same way.
func labelText(p *unison.Panel) string {
	var buffer strings.Builder
	appendLabelText(&buffer, p)
	return buffer.String()
}

func appendLabelText(buffer *strings.Builder, p *unison.Panel) {
	if p == nil || p.Hidden {
		return
	}
	text := p.Accessibility.Name
	if text == "" {
		if label, ok := p.Self.(*unison.Label); ok {
			text = label.String()
		}
	}
	if text != "" {
		if needsSeparatingSpace(buffer.String(), text) {
			buffer.WriteByte(' ')
		}
		buffer.WriteString(text)
		return
	}
	for _, child := range p.Children() {
		appendLabelText(buffer, child)
	}
}

func needsSeparatingSpace(before, after string) bool {
	if before == "" || after == "" {
		return false
	}
	last, _ := utf8.DecodeLastRuneInString(before)
	first, _ := utf8.DecodeRuneInString(after)
	return !unicode.IsSpace(last) && !unicode.IsSpace(first)
}

// tooltipTextOf returns a panel's tooltip text, as unison's unexported axTooltipText does.
func tooltipTextOf(p *unison.Panel) string {
	tip := p.Tooltip
	if tip == nil {
		return ""
	}
	if tip.Accessibility.Name != "" {
		return tip.Accessibility.Name
	}
	return labelText(tip)
}
