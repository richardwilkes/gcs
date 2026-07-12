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
	"net/url"
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/drag"
	"github.com/richardwilkes/unison/enums/mod"
)

// fakeDragInfo is a minimal drag.Info carrying only a set of data types, letting tests drive the drag callbacks
// without a native drag session.
type fakeDragInfo struct {
	types []string
}

func (f *fakeDragInfo) SourceDragOpMask() drag.Op        { return drag.Copy | drag.Move }
func (f *fakeDragInfo) DataTypes() []string              { return f.types }
func (f *fakeDragInfo) HasString() bool                  { return false }
func (f *fakeDragInfo) HasFilePaths() bool               { return false }
func (f *fakeDragInfo) HasURLs() bool                    { return false }
func (f *fakeDragInfo) HasDataType(dataType string) bool { return slices.Contains(f.types, dataType) }
func (f *fakeDragInfo) Text() string                     { return "" }
func (f *fakeDragInfo) FilePaths() []string              { return nil }
func (f *fakeDragInfo) URLs() []*url.URL                 { return nil }
func (f *fakeDragInfo) Data(_ string) []byte             { return nil }

// fakeAltDropProvider is the minimal TableProvider needed to exercise InstallTableDropSupport's alternate drop path
// (dropping a modifier onto a row) in a headless test. Only the methods that path touches do anything.
type fakeAltDropProvider struct {
	unison.SimpleTableModel[*Node[*gurps.Trait]]
	altDrops []int
}

func (p *fakeAltDropProvider) DataOwner() gurps.DataOwner                    { return nil }
func (p *fakeAltDropProvider) SetTable(_ *unison.Table[*Node[*gurps.Trait]]) {}
func (p *fakeAltDropProvider) RootData() []*gurps.Trait                      { return nil }
func (p *fakeAltDropProvider) SetRootData(_ []*gurps.Trait)                  {}
func (p *fakeAltDropProvider) DragKey() *uti.DataType                        { return traitDragKey }
func (p *fakeAltDropProvider) DragSVG() *unison.SVG                          { return nil }
func (p *fakeAltDropProvider) ItemNames() (singular, plural string)          { return "Trait", "Traits" }
func (p *fakeAltDropProvider) ColumnIDs() []int                              { return nil }
func (p *fakeAltDropProvider) HierarchyColumnID() int                        { return -1 }
func (p *fakeAltDropProvider) ExcessWidthColumnID() int                      { return -1 }
func (p *fakeAltDropProvider) ContextMenuItems() []ContextMenuItem           { return nil }
func (p *fakeAltDropProvider) Serialize() ([]byte, error)                    { return nil, nil }
func (p *fakeAltDropProvider) Deserialize(_ []byte) error                    { return nil }
func (p *fakeAltDropProvider) RefKey() string                                { return "" }
func (p *fakeAltDropProvider) AllTags() []string                             { return nil }
func (p *fakeAltDropProvider) Headers() []unison.TableColumnHeader[*Node[*gurps.Trait]] {
	return nil
}
func (p *fakeAltDropProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*gurps.Trait]]) {}
func (p *fakeAltDropProvider) DropShouldMoveData(_, _ *unison.Table[*Node[*gurps.Trait]]) bool {
	return false
}
func (p *fakeAltDropProvider) ProcessDropData(_, _ *unison.Table[*Node[*gurps.Trait]])        {}
func (p *fakeAltDropProvider) OpenEditor(_ Rebuildable, _ *unison.Table[*Node[*gurps.Trait]]) {}
func (p *fakeAltDropProvider) CreateItem(_ Rebuildable, _ *unison.Table[*Node[*gurps.Trait]], _ ItemVariant) {
}

func (p *fakeAltDropProvider) AltDropSupport() *AltDropSupport {
	return &AltDropSupport{
		DragKey: traitModifierDragKey,
		Drop:    func(rowIndex int, _ any) { p.altDrops = append(p.altDrops, rowIndex) },
	}
}

// newAltDropTestTable returns a table with drop support installed and a single root row, whose height is the table's
// MinimumRowHeight since no columns are configured, so points with a small Y are over the row and large Y values miss.
func newAltDropTestTable(provider *fakeAltDropProvider) *unison.Table[*Node[*gurps.Trait]] {
	table := unison.NewTable[*Node[*gurps.Trait]](provider)
	InstallTableDropSupport(table, provider)
	table.SetRootRows([]*Node[*gurps.Trait]{NewNode(table, nil, gurps.NewTrait(nil, nil, false), false)})
	return table
}

// TestAltDropDragFeedbackFlushed verifies that dragging a modifier over a table row immediately flushes the drawing,
// since a native drag has no continuous redraw loop and the row highlight never appears without an explicit flush.
// This covers every table built through InstallTableDropSupport: character sheets, loot sheets, templates and editors.
func TestAltDropDragFeedbackFlushed(t *testing.T) {
	c := check.New(t)
	var flushes []*unison.Panel
	original := flushDragFeedback
	flushDragFeedback = func(panel *unison.Panel) { flushes = append(flushes, panel) }
	defer func() { flushDragFeedback = original }()

	provider := &fakeAltDropProvider{}
	table := newAltDropTestTable(provider)
	overRow := geom.Point{X: 1, Y: table.MinimumRowHeight / 2}
	offRows := geom.Point{X: 1, Y: 10000}
	di := &fakeDragInfo{types: []string{traitModifierDragKey.UTI}}

	// Entering over a row must offer a copy and flush the highlight to the screen.
	c.Equal(drag.Copy, table.DragEnteredCallback(di, overRow, mod.None), "enter over row")
	c.Equal(1, len(flushes), "enter over row must flush")
	c.Equal(table.AsPanel(), flushes[0], "the table itself must be flushed")

	// Each update over a row must flush as well, since the highlight may have moved to a different row.
	c.Equal(drag.Copy, table.DragUpdatedCallback(di, overRow, mod.None), "update over row")
	c.Equal(2, len(flushes), "update over row must flush")

	// Moving off of the rows offers no drop and must flush again to erase the previous highlight.
	c.Equal(drag.None, table.DragUpdatedCallback(di, offRows, mod.None), "update off rows")
	c.Equal(3, len(flushes), "update off rows must flush to erase the highlight")

	// Exiting with no highlight showing has nothing to erase.
	table.DragExitedCallback()
	c.Equal(3, len(flushes), "exit without a highlight showing need not flush")

	// Exiting while a highlight is showing must flush to erase it.
	c.Equal(drag.Copy, table.DragEnteredCallback(di, overRow, mod.None), "re-enter over row")
	c.Equal(4, len(flushes), "re-enter over row must flush")
	table.DragExitedCallback()
	c.Equal(5, len(flushes), "exit with a highlight showing must flush to erase it")

	// A drag of some other type must not trigger the alternate drop feedback.
	flushes = nil
	c.Equal(drag.None, table.DragEnteredCallback(&fakeDragInfo{types: []string{noteDragKey.UTI}}, overRow, mod.None),
		"unrelated drag type")
	c.Equal(0, len(flushes), "unrelated drag type must not flush")
}

// TestAltDropDeliversRowIndex verifies that releasing an alternate drop over a row hands the row index to the
// provider's drop handler and erases the highlight, while a release that isn't over a row does nothing.
func TestAltDropDeliversRowIndex(t *testing.T) {
	c := check.New(t)
	var flushes int
	original := flushDragFeedback
	flushDragFeedback = func(_ *unison.Panel) { flushes++ }
	defer func() { flushDragFeedback = original }()

	provider := &fakeAltDropProvider{}
	table := newAltDropTestTable(provider)
	overRow := geom.Point{X: 1, Y: table.MinimumRowHeight / 2}
	di := &fakeDragInfo{types: []string{traitModifierDragKey.UTI}}

	c.Equal(drag.Copy, table.DragEnteredCallback(di, overRow, mod.None), "enter over row")
	flushes = 0
	c.True(table.DropCallback(di, overRow, mod.None), "drop over row must be handled")
	c.Equal([]int{0}, provider.altDrops, "drop must deliver the row index")
	c.Equal(1, flushes, "drop must flush to erase the highlight")

	// With no row targeted, the drop must be declined and the handler left uncalled.
	c.False(table.DropCallback(di, overRow, mod.None), "drop without a targeted row must be declined")
	c.Equal([]int{0}, provider.altDrops, "declined drop must not invoke the handler")
}
