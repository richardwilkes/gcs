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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/role"
)

type traitRow = *Node[*gurps.Trait]

// plainHeader is a column header not built around a label, which GCS never builds but unison allows.
type plainHeader struct {
	unison.Panel
	state unison.SortState
}

func newPlainHeader() *plainHeader {
	h := &plainHeader{}
	h.Self = h
	return h
}

func (h *plainHeader) SortState() unison.SortState         { return h.state }
func (h *plainHeader) SetSortState(state unison.SortState) { h.state = state }
func (h *plainHeader) Less() func(a, b string) bool        { return nil }

// columnHeaderTitle must name a header as unison's axColumnHeaderName does.
func TestColumnHeaderTitle(t *testing.T) {
	c := check.New(t)
	title := columnHeaderTitle[traitRow]
	c.Equal("Points", title(NewTableColumnHeader[*gurps.Trait]("Points", "The points", nil)))
	c.Equal("Trait", title(NewPageTableColumnHeader[*gurps.Trait]("Trait", "", nil)))
	c.Equal("Enabled", title(NewEditorListSVGHeader[*gurps.Trait](svg.Weight, "Enabled", nil, false)))
	c.Equal("Enabled", title(NewEditorListSVGHeader[*gurps.Trait](svg.Weight, "Enabled", nil, true)))
	c.Equal("", title(NewTableColumnHeader[*gurps.Trait]("", "", nil)), "a header with nothing to say has no title")
	named := NewTableColumnHeader[*gurps.Trait]("Pts", "", nil)
	named.Accessibility.Name = "Points"
	c.Equal("Points", title(named), "the name wins over the text")

	plain := newPlainHeader()
	c.Equal("", title(plain))
	plain.Tooltip = unison.NewTooltipWithText("The tip")
	c.Equal("The tip", title(plain))
	plain.Tooltip.Accessibility.Name = ""
	c.Equal("The tip", title(plain), "a tooltip without a name is read from its labels")
	within := unison.NewLabel()
	within.SetTitle("Within")
	plain.AddChild(within)
	c.Equal("Within", title(plain), "the labels within the header win over its tooltip")
	second := unison.NewLabel()
	second.SetTitle("the header")
	plain.AddChild(second)
	c.Equal("Within the header", title(plain), "the labels are run together with a space between them")
	second.Hidden = true
	c.Equal("Within", title(plain), "a hidden label says nothing")
	plain.Accessibility.Name = "Named"
	c.Equal("Named", title(plain))

	data := gurps.HeaderData{
		Title: gurps.HeaderBookmark, Name: "Page Reference", Detail: gurps.PageRefTooltip(),
		TitleIsImageKey: true,
	}
	c.Equal("Page Reference", title(headerFromData[*gurps.Trait](data, false)))
	c.Equal("Page Reference", title(headerFromData[*gurps.Trait](data, true)))
	data = gurps.HeaderData{
		Title: gurps.HeaderStackedCoins, Name: "Extended Value", Detail: "The value of all",
		TitleIsImageKey: true,
	}
	c.Equal("Extended Value", columnHeaderTitle[*Node[*gurps.Equipment]](headerFromData[*gurps.Equipment](data, false)))
	data = gurps.HeaderData{Title: "Pts", Name: "Points", Detail: "Points"}
	c.Equal("Points", title(headerFromData[*gurps.Trait](data, false)))
	c.Equal("Points", title(headerFromData[*gurps.Trait](data, true)))
	data = gurps.HeaderData{Title: "Trait", Detail: "The trait"}
	c.Equal("Trait", title(headerFromData[*gurps.Trait](data, false)), "a title that is a word is read as drawn")
}

// The header row is never a tab stop, so the columns are announced as the focus arrives on a list at row level.
// Arriving at cell level is a return from a cell, so nothing is said. A click puts the cursor at row level, so it
// announces the columns whatever level the list was left at.
func TestTableAnnouncesItsColumnsOnEntry(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var table *unison.Table[traitRow]
	var header *unison.TableHeader[traitRow]
	var field *StringField
	screen.Do(func() {
		entity := sheet.Entity()
		tr := gurps.NewTrait(entity, nil, false)
		tr.Name = "Alpha"
		tr.PageRef = "B34"
		entity.Traits = append(entity.Traits, tr)
		entity.Recalculate()
		sheet.Rebuild(true)
		table, _ = uxtest.FirstPanelOfType[*unison.Table[traitRow]](sheet.AsPanel())
		header, _ = uxtest.FirstPanelOfType[*unison.TableHeader[traitRow]](sheet.AsPanel())
		field, _ = uxtest.FirstPanelOfType[*StringField](sheet.AsPanel())
	})
	if table == nil || header == nil || field == nil {
		t.Fatal("the character sheet must show the traits list, its header and a text field")
	}
	const expected = "Columns: Trait, Points, Page Reference, Library Source"
	c.Equal(expected, columnsAnnouncement(header),
		"the columns are named by their titles, the abbreviation by what it stands for and the icons by their short names")
	leadColumn := func() (col int) {
		screen.Do(func() { col = table.LeadColumnIndex() })
		return col
	}

	// Asking for the tree turns accessibility on, without which nothing is announced.
	tree := screen.AccessibilityTree(wnd)
	if tree == nil {
		t.Fatal("the window must be described")
	}
	tableNode := screen.AccessibilityNodeFor(table)
	if tableNode == nil {
		t.Fatal("the traits list must be described")
	}
	var alpha *accessibility.Node
	for _, id := range tableNode.Children {
		if n := tree.Node(id); n != nil && n.Role == role.Row && strings.HasPrefix(n.Name, "Alpha") {
			alpha = n
		}
	}
	if alpha == nil {
		t.Fatal("the Alpha row must be described")
	}
	c.True(strings.Contains(alpha.Name, "Page Reference B34"), "the row reads the reference after its column's name: %q", alpha.Name)
	c.False(strings.Contains(alpha.Name, gurps.PageRefTooltip()), "the row never reads the header's tooltip: %q", alpha.Name)
	screen.Announcements()
	focus := func(p unison.Paneler) { screen.Do(func() { p.AsPanel().RequestFocus() }) }
	focus(field)
	c.Equal(0, len(screen.Announcements()), "nothing is announced for a text field")
	focus(table)
	c.Equal([]string{expected}, screen.Announcements(), "the columns are announced as the focus arrives on the list")

	screen.Do(func() { table.SetLeadCell(0, 1) })
	screen.Announcements()
	focus(field)
	focus(table)
	c.Equal(0, len(screen.Announcements()), "nothing is announced when the focus returns at cell level")
	c.Equal(1, leadColumn(), "the cursor stays on the cell")

	focus(field)
	screen.Announcements()
	screen.Click(screen.PanelPoint(table, geom.NewPoint(4, 4)))
	c.Equal(-1, leadColumn(), "a click puts the cursor at row level")
	c.Equal([]string{expected}, screen.Announcements(), "the columns are announced as a click brings the focus in")

	screen.Do(func() { table.SetLeadCell(0, -1) })
	screen.Announcements()
	focus(field)
	focus(table)
	c.Equal([]string{expected}, screen.Announcements())
}
