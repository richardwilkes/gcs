// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/layoutedge"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/layoutnode"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/paper"
	"github.com/richardwilkes/toolbox/v2/check"
)

// factoryLayoutTree is the tree FactorySheetLayout produces, in the form layoutTreeString renders.
const factoryLayoutTree = "column[" +
	"row[portrait column[row[identity miscellaneous] description] points] " +
	"row[column[row[column[primary_attributes damage] secondary_attributes] point_pools] " +
	"body column[encumbrance lifting]] " +
	"row[reactions conditional_modifiers] melee ranged row[traits skills] spells equipment other_equipment notes]"

// oldDefaultGridTemplate is the HTML grid template the block layout setting this replaced produced for its defaults.
// User-authored export templates depend on it, so the factory layout has to reproduce it byte for byte.
const oldDefaultGridTemplate = `"reactions conditional_modifiers"
"melee melee"
"ranged ranged"
"traits skills"
"spells spells"
"equipment equipment"
"other_equipment other_equipment"
"notes notes"
`

// layoutTreeString renders a layout node and everything below it as a compact string, so that a test can state the
// shape it expects in one readable line.
func layoutTreeString(node *SheetLayoutNode) string {
	if node == nil {
		return "<nil>"
	}
	if node.Type == layoutnode.Block {
		return node.Key
	}
	parts := make([]string, 0, len(node.Children))
	for _, child := range node.Children {
		parts = append(parts, layoutTreeString(child))
	}
	return node.Type.Key() + "[" + strings.Join(parts, " ") + "]"
}

// newTestLayout builds a validated layout from the given root bands, with every block the bands don't mention marked as
// hidden, so that validation doesn't append the other twenty-odd blocks and drown the assertions.
func newTestLayout(bands ...*SheetLayoutNode) *SheetLayout {
	layout := &SheetLayout{Root: containerNode(layoutnode.Column, fxp.One, bands...)}
	mentioned := make(map[string]bool)
	for _, key := range appendBlockKeys(nil, layout.Root) {
		mentioned[key] = true
	}
	for _, key := range AllBlockKeys {
		if !mentioned[key] {
			layout.Hidden = append(layout.Hidden, key)
		}
	}
	layout.EnsureValidity()
	return layout
}

// weightsOf returns the weights of the given node's children.
func weightsOf(node *SheetLayoutNode) []fxp.Int {
	weights := make([]fxp.Int, 0, len(node.Children))
	for _, child := range node.Children {
		weights = append(weights, child.Weight)
	}
	return weights
}

// TestFactorySheetLayout verifies that the factory layout has the shape the sheet has always drawn, that it holds every
// block exactly once, and that validating it changes nothing.
func TestFactorySheetLayout(t *testing.T) {
	c := check.New(t)
	layout := FactorySheetLayout()
	c.Equal(factoryLayoutTree, layoutTreeString(layout.Root))
	c.Equal(AllBlockKeys, sortedKeys(layout.VisibleKeys()))
	c.Equal(0, len(layout.HiddenKeys()))
	layout.EnsureValidity()
	c.Equal(factoryLayoutTree, layoutTreeString(layout.Root), "validating the factory layout must not change it")

	portrait, _, _ := layout.Find(BlockPortraitKey)
	c.NotNil(portrait)
	c.True(portrait.Square, "the factory portrait must take its width from the height of the row it is in")
	c.Equal(fxp.One, portrait.Weight, "the factory portrait keeps a plain weight, which its square flag overrides")
	for _, key := range AllBlockKeys {
		if key == BlockPortraitKey {
			continue
		}
		node, _, _ := layout.Find(key)
		c.NotNil(node)
		c.False(node.Square, "only the portrait is square in the factory layout, but %q was too", key)
	}
}

// sortedKeys returns the given keys in canonical order.
func sortedKeys(keys []string) []string {
	result := make([]string, 0, len(keys))
	for _, key := range AllBlockKeys {
		if slices.Contains(keys, key) {
			result = append(result, key)
		}
	}
	return result
}

// TestSheetLayoutClone verifies that a clone shares nothing with the layout it was made from.
func TestSheetLayoutClone(t *testing.T) {
	c := check.New(t)
	layout := FactorySheetLayout()
	c.True(layout.Hide(BlockNotesKey))
	clone := layout.Clone()
	c.Equal(layoutTreeString(layout.Root), layoutTreeString(clone.Root))
	c.Equal(layout.HiddenKeys(), clone.HiddenKeys())
	c.True(clone.Hide(BlockSpellsKey))
	c.True(layout.Contains(BlockSpellsKey), "altering the clone must not alter the original")
	c.True((*SheetLayout)(nil).Clone() == nil)
}

// TestSheetLayoutJSONRoundTrip verifies that a layout survives being written out and read back in, minimum heights
// included.
func TestSheetLayoutJSONRoundTrip(t *testing.T) {
	c := check.New(t)
	layout := FactorySheetLayout()
	c.True(layout.SetMinHeight(BlockNotesKey, paper.Length{Length: 1.5, Units: paper.Inch}))
	c.True(layout.Hide(BlockRangedKey))
	data, err := jio.Marshal(layout)
	c.NoError(err)
	c.Contains(string(data), `"min_height":"1.5 in"`)
	c.Contains(string(data), `"hidden":["ranged"]`)
	c.Contains(string(data), `"square":true`)
	var restored SheetLayout
	c.NoError(jio.Unmarshal(data, &restored))
	c.Equal(layoutTreeString(layout.Root), layoutTreeString(restored.Root))
	c.Equal(layout.HiddenKeys(), restored.HiddenKeys())
	node, _, _ := restored.Find(BlockNotesKey)
	c.NotNil(node)
	c.Equal(paper.Length{Length: 1.5, Units: paper.Inch}, node.MinHeight)
	c.False(node.Square, "a block that isn't square must not come back square")
	portrait, _, _ := restored.Find(BlockPortraitKey)
	c.NotNil(portrait)
	c.True(portrait.Square, "the square flag must survive the round trip")
	identity, _, _ := restored.Find(BlockIdentityKey)
	c.NotNil(identity)
	c.Equal(factoryIdentityWeight, identity.Weight, "weights must survive the round trip")
}

// TestNewSheetLayoutFromLegacyRows verifies that the rows of the block layout setting older files carry become the
// bands that follow the factory layout's non-list bands.
func TestNewSheetLayoutFromLegacyRows(t *testing.T) {
	t.Run("default rows reproduce the factory layout", func(t *testing.T) {
		c := check.New(t)
		layout := NewSheetLayoutFromLegacyRows([]string{
			"reactions conditional_modifiers",
			"melee",
			"ranged",
			"traits skills",
			"spells",
			"equipment",
			"other_equipment",
			"notes",
		})
		c.Equal(factoryLayoutTree, layoutTreeString(layout.Root))
	})
	t.Run("advantages is mapped onto traits", func(t *testing.T) {
		c := check.New(t)
		layout := NewSheetLayoutFromLegacyRows([]string{"advantages skills"})
		c.Equal([][]string{
			{BlockTraitsKey, BlockSkillsKey},
			{BlockReactionsKey},
			{BlockConditionalModifiersKey},
			{BlockMeleeKey},
			{BlockRangedKey},
			{BlockSpellsKey},
			{BlockEquipmentKey},
			{BlockOtherEquipmentKey},
			{BlockNotesKey},
		}, layout.ListBands())
	})
	t.Run("unmentioned blocks are appended in canonical order", func(t *testing.T) {
		c := check.New(t)
		layout := NewSheetLayoutFromLegacyRows([]string{"notes", "bogus", "  traits   skills  "})
		c.Equal([][]string{
			{BlockNotesKey},
			{BlockTraitsKey, BlockSkillsKey},
			{BlockReactionsKey},
			{BlockConditionalModifiersKey},
			{BlockMeleeKey},
			{BlockRangedKey},
			{BlockSpellsKey},
			{BlockEquipmentKey},
			{BlockOtherEquipmentKey},
		}, layout.ListBands())
		c.Equal(AllBlockKeys, sortedKeys(layout.VisibleKeys()))
	})
}

// TestSheetLayoutEnsureValidityRoot verifies that the root is always a non-nil Column.
func TestSheetLayoutEnsureValidityRoot(t *testing.T) {
	t.Run("nil root", func(t *testing.T) {
		c := check.New(t)
		layout := &SheetLayout{}
		layout.EnsureValidity()
		c.NotNil(layout.Root)
		c.Equal(layoutnode.Column, layout.Root.Type)
		c.Equal(AllBlockKeys, layout.VisibleKeys(), "every block must be appended as a band")
	})
	t.Run("root is a Row", func(t *testing.T) {
		c := check.New(t)
		layout := &SheetLayout{
			Root:   containerNode(layoutnode.Row, fxp.One, blockNode(BlockTraitsKey), blockNode(BlockSkillsKey)),
			Hidden: allKeysExcept(BlockTraitsKey, BlockSkillsKey),
		}
		layout.EnsureValidity()
		c.Equal("column[row[traits skills]]", layoutTreeString(layout.Root))
	})
	t.Run("root is a Block", func(t *testing.T) {
		c := check.New(t)
		layout := &SheetLayout{
			Root:   blockNode(BlockNotesKey),
			Hidden: allKeysExcept(BlockNotesKey),
		}
		layout.EnsureValidity()
		c.Equal("column[notes]", layoutTreeString(layout.Root))
	})
	t.Run("root keeps its identity as a Column", func(t *testing.T) {
		c := check.New(t)
		layout := &SheetLayout{
			Root:   containerNode(layoutnode.Column, fxp.One, blockNode(BlockNotesKey), blockNode(BlockSpellsKey)),
			Hidden: allKeysExcept(BlockNotesKey, BlockSpellsKey),
		}
		layout.Root.Key = BlockTraitsKey
		layout.EnsureValidity()
		c.Equal("column[notes spells]", layoutTreeString(layout.Root))
		c.Equal("", layout.Root.Key)
	})
}

// allKeysExcept returns every block key other than the ones given.
func allKeysExcept(keys ...string) []string {
	result := make([]string, 0, len(AllBlockKeys))
	for _, key := range AllBlockKeys {
		if !slices.Contains(keys, key) {
			result = append(result, key)
		}
	}
	return result
}

// TestSheetLayoutEnsureValidityNodes verifies that malformed nodes are repaired or dropped.
func TestSheetLayoutEnsureValidityNodes(t *testing.T) {
	t.Run("unknown keys are dropped", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(blockNode("bogus"), blockNode(BlockNotesKey))
		c.Equal("column[notes]", layoutTreeString(layout.Root))
	})
	t.Run("duplicate keys keep the first", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			containerNode(layoutnode.Row, fxp.One, blockNode(BlockNotesKey), blockNode(BlockSpellsKey)),
			blockNode(BlockNotesKey),
		)
		c.Equal("column[row[notes spells]]", layoutTreeString(layout.Root))
	})
	t.Run("a Block drops its children", func(t *testing.T) {
		c := check.New(t)
		node := blockNode(BlockNotesKey)
		node.Children = []*SheetLayoutNode{blockNode(BlockSpellsKey)}
		layout := newTestLayout(node, blockNode(BlockSpellsKey))
		c.Equal("column[notes spells]", layoutTreeString(layout.Root))
		found, _, _ := layout.Find(BlockNotesKey)
		c.Equal(0, len(found.Children))
	})
	t.Run("a container drops its key", func(t *testing.T) {
		c := check.New(t)
		row := containerNode(layoutnode.Row, fxp.One, blockNode(BlockNotesKey), blockNode(BlockSpellsKey))
		row.Key = BlockTraitsKey
		layout := newTestLayout(row)
		c.Equal("column[row[notes spells]]", layoutTreeString(layout.Root))
		c.Equal("", layout.Root.Children[0].Key)
	})
	t.Run("empty containers are dropped", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			containerNode(layoutnode.Row, fxp.One),
			blockNode(BlockNotesKey),
			containerNode(layoutnode.Column, fxp.One, containerNode(layoutnode.Row, fxp.One)),
		)
		c.Equal("column[notes]", layoutTreeString(layout.Root))
	})
	t.Run("a single-child container is replaced by its child", func(t *testing.T) {
		c := check.New(t)
		row := containerNode(layoutnode.Row, fxp.Four, blockNode(BlockNotesKey))
		row.MinHeight = paper.Length{Length: 2, Units: paper.Inch}
		layout := newTestLayout(
			containerNode(layoutnode.Row, fxp.One, row, blockNode(BlockSpellsKey)),
		)
		c.Equal("column[row[notes spells]]", layoutTreeString(layout.Root))
		notes, _, _ := layout.Find(BlockNotesKey)
		c.Equal(fxp.Four, notes.Weight, "the child must inherit the container's weight")
		c.Equal(paper.Length{Length: 2, Units: paper.Inch}, notes.MinHeight,
			"the child must inherit the container's minimum height when it has none of its own")
	})
	t.Run("a single-child container keeps the child's own minimum height", func(t *testing.T) {
		c := check.New(t)
		notesNode := blockNode(BlockNotesKey)
		notesNode.MinHeight = paper.Length{Length: 1, Units: paper.Inch}
		row := containerNode(layoutnode.Row, fxp.One, notesNode)
		row.MinHeight = paper.Length{Length: 2, Units: paper.Inch}
		layout := newTestLayout(row, blockNode(BlockSpellsKey))
		notes, _, _ := layout.Find(BlockNotesKey)
		c.Equal(paper.Length{Length: 1, Units: paper.Inch}, notes.MinHeight)
	})
	t.Run("a container of the same type as its parent is spliced in", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(containerNode(layoutnode.Row, fxp.One,
			blockNode(BlockNotesKey),
			containerNode(layoutnode.Row, fxp.One, blockNode(BlockSpellsKey), blockNode(BlockTraitsKey)),
		))
		c.Equal("column[row[notes spells traits]]", layoutTreeString(layout.Root))
	})
	t.Run("a container drops the square flag", func(t *testing.T) {
		c := check.New(t)
		row := containerNode(layoutnode.Row, fxp.One, blockNode(BlockNotesKey), squareBlockNode(BlockSpellsKey))
		row.Square = true
		column := containerNode(layoutnode.Column, fxp.One, row, blockNode(BlockTraitsKey))
		column.Square = true
		layout := newTestLayout(column)
		c.Equal("column[column[row[notes spells] traits]]", layoutTreeString(layout.Root))
		c.False(layout.Root.Square, "the root is a container and can't be square")
		c.False(layout.Root.Children[0].Square, "a Column can't be square")
		c.False(layout.Root.Children[0].Children[0].Square, "a Row can't be square")
		spells, _, _ := layout.Find(BlockSpellsKey)
		c.True(spells.Square, "a Block keeps its square flag")
	})
	t.Run("weights and minimum heights are made usable", func(t *testing.T) {
		c := check.New(t)
		node := blockNode(BlockNotesKey)
		node.Weight = 0
		node.MinHeight = paper.Length{Length: -3, Units: paper.Unit(200)}
		layout := newTestLayout(node)
		notes, _, _ := layout.Find(BlockNotesKey)
		c.Equal(fxp.One, notes.Weight)
		c.Equal(paper.Length{Length: 0, Units: paper.Inch}, notes.MinHeight)
	})
	t.Run("missing blocks are appended in canonical order", func(t *testing.T) {
		c := check.New(t)
		layout := &SheetLayout{Root: containerNode(layoutnode.Column, fxp.One, blockNode(BlockNotesKey))}
		layout.EnsureValidity()
		c.Equal(append([]string{BlockNotesKey}, allKeysExcept(BlockNotesKey)...), layout.VisibleKeys())
	})
}

// TestSheetLayoutBandGroup verifies that a band of the root may be a Column of its own -- a band group -- and that such
// a band survives validation and a JSON round trip, while the shapes that add nothing to it are still cleaned up.
func TestSheetLayoutBandGroup(t *testing.T) {
	t.Run("a Column band is kept", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			containerNode(layoutnode.Column, fxp.One, blockNode(BlockTraitsKey), blockNode(BlockSkillsKey)),
			blockNode(BlockNotesKey),
		)
		c.Equal("column[column[traits skills] notes]", layoutTreeString(layout.Root),
			"a Column band must stay a band of its own rather than being spliced into the root")
	})
	t.Run("a Column band holding a single block is replaced by it", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			containerNode(layoutnode.Column, fxp.Four, blockNode(BlockTraitsKey)),
			blockNode(BlockNotesKey),
		)
		c.Equal("column[traits notes]", layoutTreeString(layout.Root))
		traits, _, _ := layout.Find(BlockTraitsKey)
		c.Equal(fxp.Four, traits.Weight, "the block must inherit the band's weight")
	})
	t.Run("a Column within a band group is spliced into it", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(containerNode(layoutnode.Column, fxp.One,
			blockNode(BlockTraitsKey),
			containerNode(layoutnode.Column, fxp.One, blockNode(BlockSkillsKey), blockNode(BlockNotesKey)),
		))
		c.Equal("column[column[traits skills notes]]", layoutTreeString(layout.Root),
			"only a band of the root may be a Column within a Column")
	})
	t.Run("a Row within a band group stays", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(containerNode(layoutnode.Column, fxp.One,
			blockNode(BlockNotesKey),
			containerNode(layoutnode.Row, fxp.One, blockNode(BlockTraitsKey), blockNode(BlockSkillsKey)),
		))
		c.Equal("column[column[notes row[traits skills]]]", layoutTreeString(layout.Root))
	})
	t.Run("a band group survives a JSON round trip", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			containerNode(layoutnode.Column, fxp.One, blockNode(BlockTraitsKey), blockNode(BlockSkillsKey)),
			blockNode(BlockNotesKey),
		)
		data, err := jio.Marshal(layout)
		c.NoError(err)
		var restored SheetLayout
		c.NoError(jio.Unmarshal(data, &restored))
		c.Equal("column[column[traits skills] notes]", layoutTreeString(restored.Root))
		c.Equal(layout.HiddenKeys(), restored.HiddenKeys())
	})
	t.Run("a group of two lists is two lines of the grid template", func(t *testing.T) {
		c := check.New(t)
		group := newTestLayout(containerNode(layoutnode.Column, fxp.One, blockNode(BlockTraitsKey),
			blockNode(BlockSkillsKey)))
		c.Equal([][]string{{BlockTraitsKey, BlockSkillsKey}}, group.ListBands(),
			"a band group is one band, whatever it holds")
		c.True(strings.HasPrefix(group.HTMLGridTemplate(), "\"traits traits\"\n\"skills skills\"\n"),
			"blocks stacked in a band group each take a line of their own")
		row := newTestLayout(containerNode(layoutnode.Row, fxp.One, blockNode(BlockTraitsKey),
			blockNode(BlockSkillsKey)))
		c.True(strings.HasPrefix(row.HTMLGridTemplate(), "\"traits skills\"\n"),
			"the same two blocks side by side share a line")
	})
}

// TestSheetLayoutEnsureValidityHidden verifies that the hidden list is honored, cleaned up and never allowed to
// contradict the tree.
func TestSheetLayoutEnsureValidityHidden(t *testing.T) {
	t.Run("hidden blocks are not appended", func(t *testing.T) {
		c := check.New(t)
		layout := &SheetLayout{
			Root:   containerNode(layoutnode.Column, fxp.One, blockNode(BlockNotesKey)),
			Hidden: []string{BlockSpellsKey},
		}
		layout.EnsureValidity()
		c.False(layout.Contains(BlockSpellsKey))
		c.Equal([]string{BlockSpellsKey}, layout.HiddenKeys())
	})
	t.Run("hidden is cleaned up", func(t *testing.T) {
		c := check.New(t)
		layout := &SheetLayout{
			Root:   containerNode(layoutnode.Column, fxp.One, blockNode(BlockNotesKey)),
			Hidden: []string{"NOTES", "bogus", "SPELLS", "advantages", " spells ", BlockMeleeKey},
		}
		layout.EnsureValidity()
		c.Equal([]string{BlockMeleeKey, BlockTraitsKey, BlockSpellsKey}, layout.HiddenKeys(),
			"hidden must be lower-cased, mapped, de-duplicated, cleared of anything in the tree and sorted")
		c.True(layout.Contains(BlockNotesKey))
	})
	t.Run("all blocks hidden yields an empty root", func(t *testing.T) {
		c := check.New(t)
		layout := &SheetLayout{Hidden: append([]string{}, AllBlockKeys...)}
		layout.EnsureValidity()
		c.NotNil(layout.Root)
		c.Equal(layoutnode.Column, layout.Root.Type)
		c.Equal(0, len(layout.Root.Children))
		c.Equal(AllBlockKeys, layout.HiddenKeys())
	})
}

// TestSheetLayoutHideShow verifies hiding and showing blocks.
func TestSheetLayoutHideShow(t *testing.T) {
	c := check.New(t)
	layout := FactorySheetLayout()
	c.True(layout.Hide(BlockMeleeKey))
	c.False(layout.Contains(BlockMeleeKey))
	c.Equal([]string{BlockMeleeKey}, layout.HiddenKeys())
	c.False(layout.Hide(BlockMeleeKey), "hiding an already hidden block changes nothing")
	c.False(layout.Hide("bogus"))
	c.True(layout.Show(BlockMeleeKey))
	c.Equal(0, len(layout.HiddenKeys()))
	keys := layout.VisibleKeys()
	c.Equal(BlockMeleeKey, keys[len(keys)-1], "a shown block becomes the last band")
	c.False(layout.Show(BlockMeleeKey), "showing a visible block changes nothing")

	layout = FactorySheetLayout()
	c.Equal("row[traits skills]", layoutTreeString(layout.Root.Children[5]))
	c.True(layout.Hide(BlockTraitsKey))
	c.Equal(BlockSkillsKey, layout.Root.Children[5].Key,
		"the row the hidden block was in must collapse onto its remaining child")
}

// TestSheetLayoutMove verifies each of the ways a block can be moved.
func TestSheetLayoutMove(t *testing.T) {
	t.Run("into an existing row", func(t *testing.T) {
		c := check.New(t)
		traits := weightedBlockNode(BlockTraitsKey, fxp.Three)
		layout := newTestLayout(
			containerNode(layoutnode.Row, fxp.One, traits, blockNode(BlockSkillsKey)),
			blockNode(BlockNotesKey),
		)
		c.True(layout.Move(BlockNotesKey, BlockSkillsKey, layoutedge.Left))
		c.Equal("column[row[traits notes skills]]", layoutTreeString(layout.Root))
		c.Equal([]fxp.Int{fxp.Three, fxp.Two, fxp.One}, weightsOf(layout.Root.Children[0]),
			"the inserted block's weight must be the mean of the row's weights")
	})
	t.Run("to the right within an existing row", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			containerNode(layoutnode.Row, fxp.One, blockNode(BlockTraitsKey), blockNode(BlockSkillsKey)),
			blockNode(BlockNotesKey),
		)
		c.True(layout.Move(BlockNotesKey, BlockTraitsKey, layoutedge.Right))
		c.Equal("column[row[traits notes skills]]", layoutTreeString(layout.Root))
	})
	t.Run("wrapping a lone target into a row", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			weightedBlockNode(BlockSpellsKey, fxp.Four),
			blockNode(BlockTraitsKey),
			blockNode(BlockNotesKey),
		)
		c.True(layout.Move(BlockNotesKey, BlockSpellsKey, layoutedge.Right))
		c.Equal("column[row[spells notes] traits]", layoutTreeString(layout.Root))
		c.Equal(fxp.Four, layout.Root.Children[0].Weight, "the new row must inherit the target's weight")
		c.Equal([]fxp.Int{fxp.One, fxp.One}, weightsOf(layout.Root.Children[0]))
	})
	t.Run("wrapping a lone target into a column", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			containerNode(layoutnode.Row, fxp.One, blockNode(BlockTraitsKey), blockNode(BlockSkillsKey)),
			blockNode(BlockNotesKey),
		)
		c.True(layout.Move(BlockNotesKey, BlockSkillsKey, layoutedge.Bottom))
		c.Equal("column[row[traits column[skills notes]]]", layoutTreeString(layout.Root))
	})
	t.Run("above a root band groups the two of them into one band", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			blockNode(BlockTraitsKey),
			blockNode(BlockSkillsKey),
			blockNode(BlockNotesKey),
		)
		c.True(layout.Move(BlockNotesKey, BlockSkillsKey, layoutedge.Top))
		c.Equal("column[traits column[notes skills]]", layoutTreeString(layout.Root),
			"stacking onto a band groups the two of them rather than adding a band to the page")
	})
	t.Run("below a root band groups the two of them into one band", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			weightedBlockNode(BlockSkillsKey, fxp.Four),
			blockNode(BlockTraitsKey),
			blockNode(BlockNotesKey),
		)
		c.True(layout.Move(BlockTraitsKey, BlockSkillsKey, layoutedge.Bottom))
		c.Equal("column[column[skills traits] notes]", layoutTreeString(layout.Root))
		c.Equal(fxp.Four, layout.Root.Children[0].Weight, "the group must inherit the target's weight")
		c.Equal([]fxp.Int{fxp.One, fxp.One}, weightsOf(layout.Root.Children[0]))
	})
	t.Run("beside a root band still shares a row with it", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			blockNode(BlockTraitsKey),
			blockNode(BlockSkillsKey),
			blockNode(BlockNotesKey),
		)
		c.True(layout.Move(BlockNotesKey, BlockSkillsKey, layoutedge.Left))
		c.Equal("column[traits row[notes skills]]", layoutTreeString(layout.Root))
	})
	t.Run("stacking onto a block within a band group joins that group", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			containerNode(layoutnode.Column, fxp.One, blockNode(BlockTraitsKey), blockNode(BlockSkillsKey)),
			blockNode(BlockNotesKey),
		)
		c.Equal("column[column[traits skills] notes]", layoutTreeString(layout.Root))
		c.True(layout.Move(BlockNotesKey, BlockTraitsKey, layoutedge.Bottom))
		c.Equal("column[column[traits notes skills]]", layoutTreeString(layout.Root),
			"the block joins the group the target is in rather than wrapping the target again")
	})
	t.Run("moving the only remaining child collapses its container", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			containerNode(layoutnode.Row, fxp.One, blockNode(BlockTraitsKey), blockNode(BlockSkillsKey)),
			blockNode(BlockNotesKey),
		)
		c.True(layout.Move(BlockTraitsKey, BlockNotesKey, layoutedge.Right))
		c.Equal("column[skills row[notes traits]]", layoutTreeString(layout.Root))
	})
	t.Run("onto itself is a no-op", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(blockNode(BlockTraitsKey), blockNode(BlockNotesKey))
		c.False(layout.Move(BlockNotesKey, BlockNotesKey, layoutedge.Left))
		c.Equal("column[traits notes]", layoutTreeString(layout.Root))
	})
	t.Run("unknown and hidden blocks are refused", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(blockNode(BlockTraitsKey), blockNode(BlockNotesKey))
		c.False(layout.Move("bogus", BlockNotesKey, layoutedge.Left))
		c.False(layout.Move(BlockNotesKey, "bogus", layoutedge.Left))
		c.False(layout.Move(BlockSpellsKey, BlockNotesKey, layoutedge.Left), "spells is hidden")
		c.False(layout.Move(BlockNotesKey, BlockSpellsKey, layoutedge.Left), "spells is hidden")
		c.Equal("column[traits notes]", layoutTreeString(layout.Root))
	})
}

// TestSheetLayoutMoveBeside verifies that a block can be placed against the edge of a container as well as against the
// edge of another block, which is how it is made to span everything that container holds.
func TestSheetLayoutMoveBeside(t *testing.T) {
	t.Run("beside a row band joins that row", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			containerNode(layoutnode.Row, fxp.One, blockNode(BlockTraitsKey), blockNode(BlockSkillsKey)),
			blockNode(BlockNotesKey),
		)
		row := layout.Root.Children[0]
		c.True(layout.MoveBeside(BlockNotesKey, row, layoutedge.Left))
		c.Equal("column[row[notes traits skills]]", layoutTreeString(layout.Root),
			"the row wrapped around the row is spliced back into one row, so the block stands beside the whole of it")
	})
	t.Run("below a row band groups the two of them into one band", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			containerNode(layoutnode.Row, fxp.One, blockNode(BlockTraitsKey), blockNode(BlockSkillsKey)),
			blockNode(BlockNotesKey),
			blockNode(BlockSpellsKey),
		)
		row := layout.Root.Children[0]
		c.True(layout.MoveBeside(BlockSpellsKey, row, layoutedge.Bottom))
		c.Equal("column[column[row[traits skills] spells] notes]", layoutTreeString(layout.Root),
			"the block spans the full width below the row, and the two of them become one band")
	})
	t.Run("beside a column within a row joins that row", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			containerNode(layoutnode.Row, fxp.One,
				containerNode(layoutnode.Column, fxp.Three, blockNode(BlockTraitsKey), blockNode(BlockSkillsKey)),
				blockNode(BlockNotesKey),
			),
			blockNode(BlockSpellsKey),
		)
		column := layout.Root.Children[0].Children[0]
		c.True(layout.MoveBeside(BlockSpellsKey, column, layoutedge.Right))
		c.Equal("column[row[column[traits skills] spells notes]]", layoutTreeString(layout.Root),
			"the block stands beside the whole column")
		c.Equal([]fxp.Int{fxp.Three, fxp.Two, fxp.One}, weightsOf(layout.Root.Children[0]),
			"the inserted block's weight must be the mean of the row's weights")
	})
	t.Run("beside its own ancestor drops it out of that ancestor", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			containerNode(layoutnode.Row, fxp.One,
				containerNode(layoutnode.Column, fxp.One, blockNode(BlockTraitsKey), blockNode(BlockSkillsKey)),
				blockNode(BlockNotesKey),
			),
		)
		column := layout.Root.Children[0].Children[0]
		c.True(layout.MoveBeside(BlockTraitsKey, column, layoutedge.Left))
		c.Equal("column[row[traits skills notes]]", layoutTreeString(layout.Root),
			"the column the block was in is left with one child and collapses onto it")
	})
	t.Run("above a column within a row wraps that column", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			containerNode(layoutnode.Row, fxp.One,
				containerNode(layoutnode.Column, fxp.One, blockNode(BlockTraitsKey), blockNode(BlockSkillsKey)),
				blockNode(BlockNotesKey),
			),
			blockNode(BlockSpellsKey),
		)
		column := layout.Root.Children[0].Children[0]
		c.True(layout.MoveBeside(BlockSpellsKey, column, layoutedge.Top))
		c.Equal("column[row[column[spells traits skills] notes]]", layoutTreeString(layout.Root),
			"the column wrapped around the column is spliced back into one column")
	})
	t.Run("a target that can't be used is refused", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			containerNode(layoutnode.Row, fxp.One, blockNode(BlockTraitsKey), blockNode(BlockSkillsKey)),
			blockNode(BlockNotesKey),
		)
		notes, _, _ := layout.Find(BlockNotesKey)
		c.False(layout.MoveBeside(BlockNotesKey, notes, layoutedge.Left), "a block can't be moved beside itself")
		c.False(layout.MoveBeside(BlockNotesKey, nil, layoutedge.Left), "there is no such thing as no target")
		c.False(layout.MoveBeside(BlockNotesKey, layout.Root, layoutedge.Left), "the root can't be moved beside")
		c.False(layout.MoveBeside(BlockNotesKey, blockNode(BlockMeleeKey), layoutedge.Left),
			"a node that isn't in the tree can't be moved beside")
		c.False(layout.MoveBeside("bogus", layout.Root.Children[0], layoutedge.Left), "there is no such block")
		c.False(layout.MoveBeside(BlockSpellsKey, layout.Root.Children[0], layoutedge.Left), "spells is hidden")
		c.Equal("column[row[traits skills] notes]", layoutTreeString(layout.Root), "nothing may have changed")
	})
}

// TestSheetLayoutStraddle verifies that a block can be placed against the edge of a pair of things that share a
// parent, spanning both of them, which gathers the two of them into a group of their own first.
func TestSheetLayoutStraddle(t *testing.T) {
	t.Run("beside two stacked bands", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(blockNode(BlockTraitsKey), blockNode(BlockSkillsKey), blockNode(BlockNotesKey))
		traits, _, _ := layout.Find(BlockTraitsKey)
		skills, _, _ := layout.Find(BlockSkillsKey)
		c.True(layout.Straddle(BlockNotesKey, traits, skills, layoutedge.Left))
		c.Equal("column[row[notes column[traits skills]]]", layoutTreeString(layout.Root),
			"the two bands must have become one group with the block beside it")
	})
	t.Run("beside two stacked bands on the right", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(blockNode(BlockTraitsKey), blockNode(BlockSkillsKey), blockNode(BlockNotesKey))
		traits, _, _ := layout.Find(BlockTraitsKey)
		skills, _, _ := layout.Find(BlockSkillsKey)
		c.True(layout.Straddle(BlockNotesKey, skills, traits, layoutedge.Right),
			"the two nodes may be given in either order")
		c.Equal("column[row[column[traits skills] notes]]", layoutTreeString(layout.Root))
	})
	t.Run("above two blocks side by side", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			containerNode(layoutnode.Row, fxp.One,
				weightedBlockNode(BlockTraitsKey, fxp.Two),
				blockNode(BlockSkillsKey),
				blockNode(BlockSpellsKey),
			),
			blockNode(BlockNotesKey),
		)
		traits, _, _ := layout.Find(BlockTraitsKey)
		skills, _, _ := layout.Find(BlockSkillsKey)
		c.True(layout.Straddle(BlockNotesKey, traits, skills, layoutedge.Top))
		c.Equal("column[row[column[notes row[traits skills]] spells]]", layoutTreeString(layout.Root))
		band := layout.Root.Children[0]
		c.Equal([]fxp.Int{fxp.Three, fxp.One}, weightsOf(band),
			"the group must keep the share of the width the pair had between them")
		c.Equal([]fxp.Int{fxp.Two, fxp.One}, weightsOf(band.Children[0].Children[1]),
			"the pair must keep the weights they had relative to each other")
	})
	t.Run("below two blocks side by side", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			containerNode(layoutnode.Row, fxp.One,
				weightedBlockNode(BlockTraitsKey, fxp.Two),
				blockNode(BlockSkillsKey),
				blockNode(BlockSpellsKey),
			),
			blockNode(BlockNotesKey),
		)
		traits, _, _ := layout.Find(BlockTraitsKey)
		skills, _, _ := layout.Find(BlockSkillsKey)
		c.True(layout.Straddle(BlockNotesKey, traits, skills, layoutedge.Bottom))
		c.Equal("column[row[column[row[traits skills] notes] spells]]", layoutTreeString(layout.Root))
		c.Equal([]fxp.Int{fxp.Three, fxp.One}, weightsOf(layout.Root.Children[0]))
	})
	t.Run("a band lying between the two is swept in with them", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(blockNode(BlockTraitsKey), blockNode(BlockMeleeKey), blockNode(BlockSkillsKey),
			blockNode(BlockNotesKey))
		traits, _, _ := layout.Find(BlockTraitsKey)
		skills, _, _ := layout.Find(BlockSkillsKey)
		c.True(layout.Straddle(BlockNotesKey, traits, skills, layoutedge.Left))
		c.Equal("column[row[notes column[traits melee skills]]]", layoutTreeString(layout.Root),
			"the whole range from the first to the second must be grouped")
	})
	t.Run("the block being moved may come from inside the range", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			containerNode(layoutnode.Row, fxp.One, blockNode(BlockTraitsKey), blockNode(BlockSkillsKey)),
			blockNode(BlockNotesKey),
		)
		band := layout.Root.Children[0]
		notes, _, _ := layout.Find(BlockNotesKey)
		c.True(layout.Straddle(BlockSkillsKey, band, notes, layoutedge.Left))
		c.Equal("column[row[skills column[traits notes]]]", layoutTreeString(layout.Root),
			"what the block left behind is what gets grouped")
	})
	t.Run("beside a range within a row inserts the block into it", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			containerNode(layoutnode.Row, fxp.One, blockNode(BlockTraitsKey), blockNode(BlockSkillsKey),
				blockNode(BlockSpellsKey)),
			blockNode(BlockNotesKey),
		)
		traits, _, _ := layout.Find(BlockTraitsKey)
		skills, _, _ := layout.Find(BlockSkillsKey)
		c.True(layout.Straddle(BlockNotesKey, traits, skills, layoutedge.Left))
		c.Equal("column[row[notes traits skills spells]]", layoutTreeString(layout.Root),
			"a Row within a Row is spliced away, so the block simply lands beside the range")
	})
	t.Run("after a range within a row inserts the block into it", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			containerNode(layoutnode.Row, fxp.One, blockNode(BlockTraitsKey), blockNode(BlockSkillsKey),
				blockNode(BlockSpellsKey)),
			blockNode(BlockNotesKey),
		)
		traits, _, _ := layout.Find(BlockTraitsKey)
		skills, _, _ := layout.Find(BlockSkillsKey)
		c.True(layout.Straddle(BlockNotesKey, traits, skills, layoutedge.Right))
		c.Equal("column[row[traits skills notes spells]]", layoutTreeString(layout.Root))
	})
	t.Run("a pair that can't be used is refused", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			containerNode(layoutnode.Row, fxp.One, blockNode(BlockTraitsKey), blockNode(BlockSkillsKey)),
			blockNode(BlockNotesKey),
		)
		band := layout.Root.Children[0]
		traits, _, _ := layout.Find(BlockTraitsKey)
		skills, _, _ := layout.Find(BlockSkillsKey)
		notes, _, _ := layout.Find(BlockNotesKey)
		c.False(layout.Straddle(BlockTraitsKey, traits, skills, layoutedge.Left),
			"a block can't straddle a pair it is one of")
		c.False(layout.Straddle(BlockNotesKey, traits, notes, layoutedge.Left),
			"a block can't straddle a pair it is one of")
		c.False(layout.Straddle(BlockNotesKey, traits, band, layoutedge.Left),
			"the two nodes must share a parent")
		c.False(layout.Straddle(BlockNotesKey, traits, traits, layoutedge.Left), "the two nodes must differ")
		c.False(layout.Straddle(BlockNotesKey, nil, skills, layoutedge.Left), "there is no such thing as no node")
		c.False(layout.Straddle(BlockNotesKey, traits, nil, layoutedge.Left), "there is no such thing as no node")
		c.False(layout.Straddle(BlockNotesKey, traits, blockNode(BlockMeleeKey), layoutedge.Left),
			"a node that isn't in the tree can't be straddled")
		c.False(layout.Straddle("bogus", traits, skills, layoutedge.Left), "there is no such block")
		c.False(layout.Straddle(BlockSpellsKey, traits, skills, layoutedge.Left), "spells is hidden")
		c.Equal("column[row[traits skills] notes]", layoutTreeString(layout.Root), "nothing may have changed")
	})
}

// TestSheetLayoutMoveToBand verifies that a block can be made a band of its own at a chosen position.
func TestSheetLayoutMoveToBand(t *testing.T) {
	t.Run("index is adjusted when the source band vanishes", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			blockNode(BlockMeleeKey),
			blockNode(BlockRangedKey),
			blockNode(BlockNotesKey),
			blockNode(BlockSpellsKey),
		)
		c.True(layout.MoveToBand(BlockMeleeKey, 2))
		c.Equal("column[ranged melee notes spells]", layoutTreeString(layout.Root))
	})
	t.Run("moving down within the bands", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(blockNode(BlockMeleeKey), blockNode(BlockRangedKey), blockNode(BlockNotesKey))
		c.True(layout.MoveToBand(BlockMeleeKey, 3))
		c.Equal("column[ranged notes melee]", layoutTreeString(layout.Root))
	})
	t.Run("out of a container collapses it", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(
			containerNode(layoutnode.Row, fxp.One, blockNode(BlockTraitsKey), blockNode(BlockSkillsKey)),
			blockNode(BlockNotesKey),
		)
		c.True(layout.MoveToBand(BlockTraitsKey, 2))
		c.Equal("column[skills notes traits]", layoutTreeString(layout.Root))
	})
	t.Run("out of range indexes are clamped", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(blockNode(BlockMeleeKey), blockNode(BlockRangedKey))
		c.True(layout.MoveToBand(BlockRangedKey, -5))
		c.Equal("column[ranged melee]", layoutTreeString(layout.Root))
		c.True(layout.MoveToBand(BlockRangedKey, 99))
		c.Equal("column[melee ranged]", layoutTreeString(layout.Root))
	})
	t.Run("unknown blocks are refused", func(t *testing.T) {
		c := check.New(t)
		layout := newTestLayout(blockNode(BlockMeleeKey))
		c.False(layout.MoveToBand("bogus", 0))
		c.False(layout.MoveToBand(BlockNotesKey, 0), "notes is hidden")
	})
}

// TestSheetLayoutSetWeightsAndMinHeight verifies the sizing operations.
func TestSheetLayoutSetWeightsAndMinHeight(t *testing.T) {
	c := check.New(t)
	layout := newTestLayout(containerNode(layoutnode.Row, fxp.One, blockNode(BlockTraitsKey),
		blockNode(BlockSkillsKey)))
	layout.SetWeights(layout.Root.Children[0], []fxp.Int{fxp.Three, fxp.Five, fxp.Ten})
	c.Equal([]fxp.Int{fxp.Three, fxp.Five}, weightsOf(layout.Root.Children[0]))
	layout.SetWeights(layout.Root.Children[0], []fxp.Int{0, -fxp.One})
	c.Equal([]fxp.Int{fxp.Three, fxp.Five}, weightsOf(layout.Root.Children[0]),
		"weights that aren't greater than zero must be ignored")
	layout.SetWeights(nil, []fxp.Int{fxp.One})
	height := paper.Length{Length: 3, Units: paper.Centimeter}
	c.True(layout.SetMinHeight(BlockTraitsKey, height))
	node, _, _ := layout.Find(BlockTraitsKey)
	c.Equal(height, node.MinHeight)
	c.False(layout.SetMinHeight(BlockTraitsKey, height), "setting the same height changes nothing")
	c.False(layout.SetMinHeight(BlockNotesKey, height), "notes is hidden")
}

// TestSheetLayoutFiltered verifies that a filtered copy holds only the blocks that were kept and doesn't grow the rest
// back when it is validated.
func TestSheetLayoutFiltered(t *testing.T) {
	c := check.New(t)
	filtered := FactorySheetLayout().Filtered(IsTemplateBlockKey)
	c.Equal("column[row[traits skills] spells equipment notes]", layoutTreeString(filtered.Root))
	c.Equal(allKeysExcept(BlockTraitsKey, BlockSkillsKey, BlockSpellsKey, BlockEquipmentKey, BlockNotesKey),
		filtered.HiddenKeys())
	filtered.EnsureValidity()
	c.Equal("column[row[traits skills] spells equipment notes]", layoutTreeString(filtered.Root))
}

// TestSheetLayoutListBands verifies the projection the sheet, the template and the exporter build their content from.
func TestSheetLayoutListBands(t *testing.T) {
	c := check.New(t)
	c.Equal([][]string{
		{BlockReactionsKey, BlockConditionalModifiersKey},
		{BlockMeleeKey},
		{BlockRangedKey},
		{BlockTraitsKey, BlockSkillsKey},
		{BlockSpellsKey},
		{BlockEquipmentKey},
		{BlockOtherEquipmentKey},
		{BlockNotesKey},
	}, FactorySheetLayout().ListBands(), "bands holding no list block must be omitted")
	layout := newTestLayout(containerNode(layoutnode.Row, fxp.One,
		blockNode(BlockBodyKey),
		containerNode(layoutnode.Column, fxp.One, blockNode(BlockNotesKey), blockNode(BlockSpellsKey)),
	))
	c.Equal([][]string{{BlockNotesKey, BlockSpellsKey}}, layout.ListBands(),
		"a band's list blocks must be collected depth-first, ignoring the blocks that aren't lists")
}

// TestSheetLayoutHTMLGridTemplate verifies that the CSS grid template user-authored export templates rely on is
// unchanged for the factory layout, and that blocks that aren't on the sheet are still named.
func TestSheetLayoutHTMLGridTemplate(t *testing.T) {
	t.Run("factory layout", func(t *testing.T) {
		c := check.New(t)
		c.Equal(oldDefaultGridTemplate, FactorySheetLayout().HTMLGridTemplate())
	})
	t.Run("hidden blocks are appended", func(t *testing.T) {
		c := check.New(t)
		layout := FactorySheetLayout()
		c.True(layout.Hide(BlockMeleeKey))
		c.Equal(`"reactions conditional_modifiers"
"ranged ranged"
"traits skills"
"spells spells"
"equipment equipment"
"other_equipment other_equipment"
"notes notes"
"melee melee"
`, layout.HTMLGridTemplate())
	})
	t.Run("only a row of exactly two shares a line", func(t *testing.T) {
		// The first band is a band group, so the two blocks stacked in it take a line each, just as they would have
		// done as two bands of their own.
		c := check.New(t)
		layout := newTestLayout(
			containerNode(layoutnode.Column, fxp.One, blockNode(BlockNotesKey), blockNode(BlockSpellsKey)),
			containerNode(layoutnode.Row, fxp.One, blockNode(BlockMeleeKey), blockNode(BlockRangedKey),
				blockNode(BlockTraitsKey)),
		)
		c.Equal(`"notes notes"
"spells spells"
"melee melee"
"ranged ranged"
"traits traits"
"reactions reactions"
"conditional_modifiers conditional_modifiers"
"skills skills"
"equipment equipment"
"other_equipment other_equipment"
`, layout.HTMLGridTemplate())
	})
}
