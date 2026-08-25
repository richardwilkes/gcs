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
	"hash"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/layoutedge"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/layoutnode"
	"github.com/richardwilkes/gcs/v5/model/paper"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xstrings"
)

// Block keys. Each one identifies a single top-level block on the sheet. The ten list keys keep the exact string values
// the old block layout setting used, since they double as the unison RefKeys the page lists are found by and as the CSS
// grid-area names user-authored HTML export templates place their blocks with.
const (
	BlockPortraitKey             = "portrait"
	BlockIdentityKey             = "identity"
	BlockMiscellaneousKey        = "miscellaneous"
	BlockPointsKey               = "points"
	BlockDescriptionKey          = "description"
	BlockPrimaryAttributesKey    = "primary_attributes"
	BlockSecondaryAttributesKey  = "secondary_attributes"
	BlockPointPoolsKey           = "point_pools"
	BlockBodyKey                 = "body"
	BlockDamageKey               = "damage"
	BlockEncumbranceKey          = "encumbrance"
	BlockLiftingKey              = "lifting"
	BlockReactionsKey            = "reactions"
	BlockConditionalModifiersKey = "conditional_modifiers"
	BlockMeleeKey                = "melee"
	BlockRangedKey               = "ranged"
	BlockTraitsKey               = "traits"
	BlockSkillsKey               = "skills"
	BlockSpellsKey               = "spells"
	BlockEquipmentKey            = "equipment"
	BlockOtherEquipmentKey       = "other_equipment"
	BlockNotesKey                = "notes"
)

// AllBlockKeys holds every block key, in the canonical order. This is the order blocks that a layout fails to mention
// are appended to it in, as well as the order hidden blocks are listed in.
var AllBlockKeys = []string{
	BlockPortraitKey,
	BlockIdentityKey,
	BlockMiscellaneousKey,
	BlockPointsKey,
	BlockDescriptionKey,
	BlockPrimaryAttributesKey,
	BlockSecondaryAttributesKey,
	BlockPointPoolsKey,
	BlockBodyKey,
	BlockDamageKey,
	BlockEncumbranceKey,
	BlockLiftingKey,
	BlockReactionsKey,
	BlockConditionalModifiersKey,
	BlockMeleeKey,
	BlockRangedKey,
	BlockTraitsKey,
	BlockSkillsKey,
	BlockSpellsKey,
	BlockEquipmentKey,
	BlockOtherEquipmentKey,
	BlockNotesKey,
}

// blockKeyOrder maps a block key to its position in AllBlockKeys. A key that isn't present isn't a block key at all,
// which is how validity is checked.
var blockKeyOrder = func() map[string]int {
	m := make(map[string]int, len(AllBlockKeys))
	for i, key := range AllBlockKeys {
		m[key] = i
	}
	return m
}()

// listBlockKeys holds the block keys of the ten list blocks, i.e. the ones backed by a table of rows and therefore the
// only ones permitted to split across a page boundary.
var listBlockKeys = map[string]bool{
	BlockReactionsKey:            true,
	BlockConditionalModifiersKey: true,
	BlockMeleeKey:                true,
	BlockRangedKey:               true,
	BlockTraitsKey:               true,
	BlockSkillsKey:               true,
	BlockSpellsKey:               true,
	BlockEquipmentKey:            true,
	BlockOtherEquipmentKey:       true,
	BlockNotesKey:                true,
}

// templateBlockKeys holds the block keys a template file can show.
var templateBlockKeys = map[string]bool{
	BlockTraitsKey:    true,
	BlockSkillsKey:    true,
	BlockSpellsKey:    true,
	BlockEquipmentKey: true,
	BlockNotesKey:     true,
}

// BlockTitle returns the title of the block with the given key, as shown on the sheet. An unknown key yields an empty
// string.
func BlockTitle(key string) string {
	switch key {
	case BlockPortraitKey:
		return i18n.Text("Portrait")
	case BlockIdentityKey:
		return i18n.Text("Identity")
	case BlockMiscellaneousKey:
		return i18n.Text("Miscellaneous")
	case BlockPointsKey:
		return i18n.Text("Points")
	case BlockDescriptionKey:
		return i18n.Text("Description")
	case BlockPrimaryAttributesKey:
		return i18n.Text("Primary Attributes")
	case BlockSecondaryAttributesKey:
		return i18n.Text("Secondary Attributes")
	case BlockPointPoolsKey:
		return i18n.Text("Point Pools")
	case BlockBodyKey:
		return i18n.Text("Body Type")
	case BlockDamageKey:
		return i18n.Text("Basic Damage")
	case BlockEncumbranceKey:
		return i18n.Text("Encumbrance, Move & Dodge")
	case BlockLiftingKey:
		return i18n.Text("Lifting & Moving Things")
	case BlockReactionsKey:
		return i18n.Text("Reaction Modifiers")
	case BlockConditionalModifiersKey:
		return i18n.Text("Conditional Modifiers")
	case BlockMeleeKey:
		return i18n.Text("Melee Weapons")
	case BlockRangedKey:
		return i18n.Text("Ranged Weapons")
	case BlockTraitsKey:
		return i18n.Text("Traits")
	case BlockSkillsKey:
		return i18n.Text("Skills")
	case BlockSpellsKey:
		return i18n.Text("Spells")
	case BlockEquipmentKey:
		return i18n.Text("Carried Equipment")
	case BlockOtherEquipmentKey:
		return i18n.Text("Other Equipment")
	case BlockNotesKey:
		return i18n.Text("Notes")
	default:
		return ""
	}
}

// IsBlockKey returns true if the given key identifies a block.
func IsBlockKey(key string) bool {
	_, exists := blockKeyOrder[key]
	return exists
}

// IsListBlockKey returns true if the given key identifies one of the ten list blocks. Only those may be split across a
// page boundary when printing or exporting.
func IsListBlockKey(key string) bool {
	return listBlockKeys[key]
}

// IsTemplateBlockKey returns true if the given key identifies a block a template file can show.
func IsTemplateBlockKey(key string) bool {
	return templateBlockKeys[key]
}

// mapOldBlockKeys converts a block key that has been renamed since it was written into its current form.
func mapOldBlockKeys(key string) string {
	if key == "advantages" {
		return BlockTraitsKey
	}
	return key
}

// normalizeBlockKey puts a block key that came from a file into the form the rest of the code expects.
func normalizeBlockKey(key string) string {
	return mapOldBlockKeys(strings.ToLower(strings.TrimSpace(key)))
}

// SheetLayoutNode is one node in the tree of blocks, rows and columns that makes up a sheet layout. A Block node names
// the block it shows and has no children; a Row lays its children out side by side, dividing the available width among
// them in proportion to their weights; a Column stacks its children.
type SheetLayoutNode struct {
	Type      layoutnode.Type    `json:"type"`
	Key       string             `json:"key,omitzero"`        // Block only
	Weight    fxp.Int            `json:"weight,omitzero"`     // Share of the width when the parent is a Row; <= 0 is 1
	MinHeight paper.Length       `json:"min_height,omitzero"` // Block only; 0 means use the natural height
	Children  []*SheetLayoutNode `json:"children,omitzero"`   // Row and Column only
	// Square is Block only. In a Row, the block's width is taken from the row's height so that its content is square,
	// instead of from its weight; ignored elsewhere.
	Square bool `json:"square,omitzero"`
}

// Clone creates a deep copy of this node.
func (n *SheetLayoutNode) Clone() *SheetLayoutNode {
	if n == nil {
		return nil
	}
	clone := *n
	if len(n.Children) != 0 {
		clone.Children = make([]*SheetLayoutNode, len(n.Children))
		for i, child := range n.Children {
			clone.Children[i] = child.Clone()
		}
	}
	return &clone
}

// Hash writes this object's contents into the hasher.
func (n *SheetLayoutNode) Hash(h hash.Hash) {
	if n == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, n.Type)
	xhash.StringWithLen(h, n.Key)
	xhash.Num64(h, n.Weight)
	xhash.Float64(h, n.MinHeight.Length)
	xhash.Num8(h, n.MinHeight.Units)
	xhash.Bool(h, n.Square)
	xhash.Num64(h, len(n.Children))
	for _, child := range n.Children {
		child.Hash(h)
	}
}

// SheetLayout holds the arrangement of the blocks on a sheet. Root is always a non-nil Column whose children are the
// full-width bands the sheet is made of, top to bottom. A band may be a Column itself -- a band group, which stacks
// what it holds the way separate bands would while remaining one thing that a block can be placed beside so as to span
// all of them. Hidden holds the keys of the blocks that have deliberately been removed from the sheet; any block that
// appears in neither is appended as a new band by EnsureValidity, so a layout written by an older version of GCS gains
// the blocks it doesn't know about.
type SheetLayout struct {
	Root   *SheetLayoutNode `json:"root"`
	Hidden []string         `json:"hidden,omitzero"`
}

// blockNode creates a new Block node for the given key.
func blockNode(key string) *SheetLayoutNode {
	return &SheetLayoutNode{
		Type:   layoutnode.Block,
		Key:    key,
		Weight: fxp.One,
	}
}

// weightedBlockNode creates a new Block node for the given key with the given weight.
func weightedBlockNode(key string, nodeWeight fxp.Int) *SheetLayoutNode {
	return &SheetLayoutNode{
		Type:   layoutnode.Block,
		Key:    key,
		Weight: nodeWeight,
	}
}

// squareBlockNode creates a new Block node for the given key whose width, when it is in a Row, is taken from the height
// of that row rather than from its weight, so that its content comes out square.
func squareBlockNode(key string) *SheetLayoutNode {
	node := blockNode(key)
	node.Square = true
	return node
}

// containerNode creates a new Row or Column node with the given weight and children.
func containerNode(nodeType layoutnode.Type, nodeWeight fxp.Int, children ...*SheetLayoutNode) *SheetLayoutNode {
	return &SheetLayoutNode{
		Type:     nodeType,
		Weight:   nodeWeight,
		Children: children,
	}
}

// The weights of the factory layout's top bands. They are the proportions the blocks came out at when the sheet drew
// its top area from a pair of fixed grids, measured at the default page width, so that a sheet that has never had its
// layout touched looks the way it always has.
//
// The portrait has no weight among them: it is marked square instead, taking its width from the height of the row it is
// in, which is what its hand-tuned weight was only able to imitate at one page size and one set of fonts.
var (
	factoryIdentityColumnWeight    = fxp.FromStringForced("4.3")
	factoryIdentityWeight          = fxp.FromStringForced("1.6")
	factoryAttributesColumnWeight  = fxp.FromStringForced("1.5")
	factoryEncumbranceColumnWeight = fxp.FromStringForced("1.6")
)

// factoryTopBands returns the bands that hold the twelve non-list blocks, arranged the way the sheet has always drawn
// them.
func factoryTopBands() []*SheetLayoutNode {
	return []*SheetLayoutNode{
		// Points stands beside the identity/description column rather than inside its first row: it is far taller
		// than the identity block, and putting it in that row would stretch the row to its height, whereas as a column
		// of its own it spans both rows, the way the old grid let it.
		containerNode(
			layoutnode.Row, fxp.One,
			squareBlockNode(BlockPortraitKey),
			containerNode(
				layoutnode.Column, factoryIdentityColumnWeight,
				containerNode(
					layoutnode.Row, fxp.One,
					weightedBlockNode(BlockIdentityKey, factoryIdentityWeight),
					blockNode(BlockMiscellaneousKey),
				),
				blockNode(BlockDescriptionKey),
			),
			blockNode(BlockPointsKey),
		),
		containerNode(
			layoutnode.Row, fxp.One,
			containerNode(
				layoutnode.Column, factoryAttributesColumnWeight,
				containerNode(
					layoutnode.Row, fxp.One,
					containerNode(
						layoutnode.Column, fxp.One,
						blockNode(BlockPrimaryAttributesKey),
						blockNode(BlockDamageKey),
					),
					blockNode(BlockSecondaryAttributesKey),
				),
				blockNode(BlockPointPoolsKey),
			),
			blockNode(BlockBodyKey),
			containerNode(
				layoutnode.Column, factoryEncumbranceColumnWeight,
				blockNode(BlockEncumbranceKey),
				blockNode(BlockLiftingKey),
			),
		),
	}
}

// FactorySheetLayout returns a new SheetLayout with the factory defaults, which reproduce the arrangement the sheet has
// always used.
func FactorySheetLayout() *SheetLayout {
	children := factoryTopBands()
	children = append(
		children,
		containerNode(
			layoutnode.Row, fxp.One,
			blockNode(BlockReactionsKey),
			blockNode(BlockConditionalModifiersKey),
		),
		blockNode(BlockMeleeKey),
		blockNode(BlockRangedKey),
		containerNode(
			layoutnode.Row, fxp.One,
			blockNode(BlockTraitsKey),
			blockNode(BlockSkillsKey),
		),
		blockNode(BlockSpellsKey),
		blockNode(BlockEquipmentKey),
		blockNode(BlockOtherEquipmentKey),
		blockNode(BlockNotesKey),
	)
	return &SheetLayout{Root: containerNode(layoutnode.Column, fxp.One, children...)}
}

// NewSheetLayoutFromLegacyRows creates a new SheetLayout from the rows of the "block_layout" setting that older files
// carry. Each row names one or two blocks that were shown side by side; they become the bands that follow the factory
// layout's non-list bands.
func NewSheetLayoutFromLegacyRows(rows []string) *SheetLayout {
	children := factoryTopBands()
	for _, row := range rows {
		var keys []string
		for part := range strings.SplitSeq(xstrings.CollapseSpaces(strings.ToLower(strings.TrimSpace(row))), " ") {
			if part = mapOldBlockKeys(part); IsBlockKey(part) {
				keys = append(keys, part)
			}
		}
		switch len(keys) {
		case 0:
		case 1:
			children = append(children, blockNode(keys[0]))
		default:
			band := containerNode(layoutnode.Row, fxp.One)
			for _, key := range keys {
				band.Children = append(band.Children, blockNode(key))
			}
			children = append(children, band)
		}
	}
	layout := &SheetLayout{Root: containerNode(layoutnode.Column, fxp.One, children...)}
	layout.EnsureValidity()
	return layout
}

// Clone creates a deep copy of this layout.
func (l *SheetLayout) Clone() *SheetLayout {
	if l == nil {
		return nil
	}
	clone := &SheetLayout{Root: l.Root.Clone()}
	if len(l.Hidden) != 0 {
		clone.Hidden = slices.Clone(l.Hidden)
	}
	return clone
}

// Hash writes this object's contents into the hasher. This is what tells one arrangement of the blocks from another,
// which is how an edit that turned out to change nothing -- a block dragged back to where it started, say -- is
// recognized and left off the undo stack.
func (l *SheetLayout) Hash(h hash.Hash) {
	if l == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	l.Root.Hash(h)
	xhash.Num64(h, len(l.Hidden))
	for _, key := range l.Hidden {
		xhash.StringWithLen(h, key)
	}
}

// Reset returns the layout to the factory defaults.
func (l *SheetLayout) Reset() {
	*l = *FactorySheetLayout()
}

// EnsureValidity checks the current layout for validity and if it isn't valid, makes it so.
//
// This rewrites the layout into its canonical form: the root is a Column; every block appears exactly once, either
// in the tree or in Hidden; containers hold at least two children; and every weight and minimum height is usable.
//
// A band of the root may be a Column of its own -- a band group, which is how two bands are made to stack as one thing
// that something else can be placed beside so as to span both of them. Only the root's own children are spared the
// splicing of a container into a same-typed parent: a Column anywhere below a band group is still spliced into it, so
// nothing deeper than a band of the root is ever a Column within a Column.
func (l *SheetLayout) EnsureValidity() {
	if l.Root == nil {
		l.Root = containerNode(layoutnode.Column, fxp.One)
	} else if l.Root.Type = l.Root.Type.EnsureValid(); l.Root.Type != layoutnode.Column {
		l.Root = containerNode(layoutnode.Column, fxp.One, l.Root)
	}
	l.Root.Key = ""
	l.Root.Weight = fxp.One
	l.Root.MinHeight = paper.Length{}
	l.Root.Square = false
	seen := make(map[string]bool, len(AllBlockKeys))
	bands := make([]*SheetLayoutNode, 0, len(l.Root.Children))
	for _, child := range l.Root.Children {
		if child = normalizeLayoutNode(child, seen); child != nil {
			bands = append(bands, child)
		}
	}
	l.Root.Children = bands
	hiddenSeen := make(map[string]bool, len(l.Hidden))
	hidden := make([]string, 0, len(l.Hidden))
	for _, key := range l.Hidden {
		if key = normalizeBlockKey(key); IsBlockKey(key) && !seen[key] && !hiddenSeen[key] {
			hiddenSeen[key] = true
			hidden = append(hidden, key)
		}
	}
	slices.SortFunc(hidden, func(a, b string) int { return blockKeyOrder[a] - blockKeyOrder[b] })
	if len(hidden) == 0 {
		hidden = nil
	}
	l.Hidden = hidden
	for _, key := range AllBlockKeys {
		if !seen[key] && !hiddenSeen[key] {
			l.Root.Children = append(l.Root.Children, blockNode(key))
		}
	}
}

// normalizeLayoutNode normalizes a single node and everything below it, returning the node that should take its place,
// or nil if it should be dropped altogether. Blocks with an unknown or already-seen key are dropped, as are containers
// that end up with no children; a container left holding a single child is replaced by that child, which inherits the
// container's weight and, if it has none of its own, its minimum height. A container never carries the square flag,
// since only a block has content to be made square.
func normalizeLayoutNode(node *SheetLayoutNode, seen map[string]bool) *SheetLayoutNode {
	if node == nil {
		return nil
	}
	node.Type = node.Type.EnsureValid()
	if node.Weight <= 0 {
		node.Weight = fxp.One
	}
	node.MinHeight.EnsureValidity()
	if node.Type == layoutnode.Block {
		node.Children = nil
		if node.Key = normalizeBlockKey(node.Key); !IsBlockKey(node.Key) || seen[node.Key] {
			return nil
		}
		seen[node.Key] = true
		return node
	}
	node.Key = ""
	node.Square = false
	children := make([]*SheetLayoutNode, 0, len(node.Children))
	for _, child := range node.Children {
		if child = normalizeLayoutNode(child, seen); child == nil {
			continue
		}
		if child.Type == node.Type {
			// A container of the same type as its parent adds nothing, so its children are spliced in directly.
			children = append(children, child.Children...)
			continue
		}
		children = append(children, child)
	}
	node.Children = children
	switch len(children) {
	case 0:
		return nil
	case 1:
		only := children[0]
		only.Weight = node.Weight
		if only.MinHeight.Length == 0 {
			only.MinHeight = node.MinHeight
		}
		return only
	default:
		return node
	}
}

// Contains returns true if the block with the given key is present in the tree.
func (l *SheetLayout) Contains(key string) bool {
	node, _, _ := l.Find(key)
	return node != nil
}

// Find locates the block with the given key, returning it along with its parent and its index within that parent. If
// the key isn't in the tree, (nil, nil, -1) is returned.
func (l *SheetLayout) Find(key string) (node, parent *SheetLayoutNode, index int) {
	return findBlockNode(l.Root, normalizeBlockKey(key))
}

func findBlockNode(within *SheetLayoutNode, key string) (node, parent *SheetLayoutNode, index int) {
	if within == nil {
		return nil, nil, -1
	}
	for i, child := range within.Children {
		if child.Type == layoutnode.Block {
			if child.Key == key {
				return child, within, i
			}
			continue
		}
		if found, foundParent, foundIndex := findBlockNode(child, key); found != nil {
			return found, foundParent, foundIndex
		}
	}
	return nil, nil, -1
}

// VisibleKeys returns the keys of the blocks in the tree, in depth-first order, i.e. the order they are drawn in.
func (l *SheetLayout) VisibleKeys() []string {
	return appendBlockKeys(make([]string, 0, len(AllBlockKeys)), l.Root)
}

// HiddenKeys returns the keys of the blocks that have been removed from the sheet, in canonical order.
func (l *SheetLayout) HiddenKeys() []string {
	return slices.Clone(l.Hidden)
}

func appendBlockKeys(keys []string, node *SheetLayoutNode) []string {
	if node == nil {
		return keys
	}
	for _, child := range node.Children {
		if child.Type == layoutnode.Block {
			keys = append(keys, child.Key)
		} else {
			keys = appendBlockKeys(keys, child)
		}
	}
	return keys
}

// Hide removes the block with the given key from the tree and records it as hidden. Returns true if the layout was
// altered.
func (l *SheetLayout) Hide(key string) bool {
	key = normalizeBlockKey(key)
	if !IsBlockKey(key) {
		return false
	}
	node, parent, index := l.Find(key)
	if node == nil {
		return false
	}
	parent.Children = slices.Delete(parent.Children, index, index+1)
	l.Hidden = append(l.Hidden, key)
	l.EnsureValidity()
	return true
}

// Show restores a hidden block by appending it as a new band at the bottom of the sheet. Returns true if the layout was
// altered.
func (l *SheetLayout) Show(key string) bool {
	key = normalizeBlockKey(key)
	if !IsBlockKey(key) || l.Contains(key) {
		return false
	}
	l.Hidden = slices.DeleteFunc(l.Hidden, func(one string) bool { return one == key })
	l.Root.Children = append(l.Root.Children, blockNode(key))
	l.EnsureValidity()
	return true
}

// Move relocates the block with the given key so that it sits against the given edge of the target block. Returns true
// if the layout was altered.
func (l *SheetLayout) Move(key, target string, edge layoutedge.Enum) bool {
	targetNode, _, _ := l.Find(target)
	return l.MoveBeside(key, targetNode, edge)
}

// MoveBeside relocates the block with the given key so that it sits against the given edge of the target node, which
// may be any node in the tree other than the root: a block, or a row or column with any number of blocks below it.
// Placing a block beside a container is how it is made to span everything in that container, which the edge of a single
// block within it can't ask for.
//
// The block is detached from wherever it was first, so a target that is one of its own ancestors simply loses it. If
// the target's parent is already the kind of container the edge calls for -- a Row for Left and Right, a Column for Top
// and Bottom -- the block becomes a sibling of the target there, with the mean of that parent's weights; if it isn't,
// the target is replaced by a new container of that kind, inheriting the target's weight and holding the two of them in
// the order the edge calls for. Returns true if the layout was altered, and false for an unknown or absent key, a
// target that is nil, the root, the block itself, or not in the tree.
//
// A band of the root is the one exception to the sibling rule: Top and Bottom there wrap the target in a band group of
// its own rather than making the block a band of the page, since MoveToBand is what asks for a band and this is the
// only way to ask for the two of them to stack as one thing that a third block can then be placed beside.
func (l *SheetLayout) MoveBeside(key string, target *SheetLayoutNode, edge layoutedge.Enum) bool {
	key = normalizeBlockKey(key)
	if !IsBlockKey(key) || target == nil || target == l.Root {
		return false
	}
	node, parent, index := l.Find(key)
	if node == nil || node == target {
		return false
	}
	if targetParent, _ := findLayoutNodeParent(l.Root, target); targetParent == nil {
		return false
	}
	parent.Children = slices.Delete(parent.Children, index, index+1)
	targetParent, targetIndex := findLayoutNodeParent(l.Root, target)
	if targetParent == nil {
		// Can't happen, since the target was found a moment ago and isn't the block that was just detached, but if it
		// somehow did, putting the block back is better than losing it.
		parent.Children = slices.Insert(parent.Children, index, node)
		return false
	}
	desired := layoutnode.Column
	if edge = edge.EnsureValid(); edge == layoutedge.Left || edge == layoutedge.Right {
		desired = layoutnode.Row
	}
	// Stacking something against a band of the page groups the two of them into a band of their own instead of adding a
	// band to the page, which is MoveToBand's business.
	stackingOntoABand := desired == layoutnode.Column && targetParent == l.Root
	if targetParent.Type == desired && !stackingOntoABand {
		node.Weight = meanWeight(targetParent.Children)
		if edge == layoutedge.Right || edge == layoutedge.Bottom {
			targetIndex++
		}
		targetParent.Children = slices.Insert(targetParent.Children, targetIndex, node)
	} else {
		container := containerNode(desired, target.Weight)
		node.Weight = fxp.One
		target.Weight = fxp.One
		if edge == layoutedge.Left || edge == layoutedge.Top {
			container.Children = []*SheetLayoutNode{node, target}
		} else {
			container.Children = []*SheetLayoutNode{target, node}
		}
		targetParent.Children[targetIndex] = container
	}
	l.EnsureValidity()
	return true
}

// Parent locates the given node within the tree by identity rather than by key, since a container has no key to be
// found by, and returns the node's parent along with its index within that parent. (nil, -1) is returned for a node
// that is nil, is the root, or isn't in the tree.
func (l *SheetLayout) Parent(node *SheetLayoutNode) (parent *SheetLayoutNode, index int) {
	if node == nil || node == l.Root {
		return nil, -1
	}
	return findLayoutNodeParent(l.Root, node)
}

// findLayoutNodeParent locates the given node within the tree by identity rather than by key, since a container has no
// key to be found by. Returns the node's parent and its index within it, or (nil, -1) if it isn't in the tree.
func findLayoutNodeParent(within, target *SheetLayoutNode) (parent *SheetLayoutNode, index int) {
	if within == nil {
		return nil, -1
	}
	for i, child := range within.Children {
		if child == target {
			return within, i
		}
		if found, foundIndex := findLayoutNodeParent(child, target); found != nil {
			return found, foundIndex
		}
	}
	return nil, -1
}

// Straddle relocates the block with the given key so that it sits against the given edge of the pair of nodes given,
// spanning both of them and everything between them. The two nodes must be children of one and the same parent -- any
// container, the root included -- and may be given in either order; everything from the earlier of them to the later,
// inclusive, is gathered into a group of its own, so that something lying between them that has nothing to show is
// swept in with them rather than being left behind. The block then stands against the given edge of that group: Left
// and Right put the two of them in a Row, Top and Bottom in a Column.
//
// This is the only way to ask for a block to span two things that aren't already gathered into a container of their
// own, since the edge of a single one of them can't ask for it and MoveBeside can only place a block against something
// the tree already holds as one thing.
//
// The block is detached from wherever it was before the group is formed, so straddling a pair it was itself inside
// works: whatever it leaves behind is what gets grouped. The group keeps the weights of the things it gathers, and the
// container that takes the range's place inherits the sum of them, so that a pair within a Row keeps the share of the
// width the two of them had between them.
//
// An edge that calls for the kind of container the parent already is -- Left or Right within a Row, Top or Bottom
// within a Column -- is not a special case: validation splices such a container away again, which degenerates into
// inserting the block beside the range rather than around it, which is what that gesture means.
//
// Returns true if the layout was altered, and false for an unknown or absent key, a node that is nil, isn't in the
// tree, is the block being moved, or doesn't share its parent with the other.
func (l *SheetLayout) Straddle(key string, first, second *SheetLayoutNode, edge layoutedge.Enum) bool {
	key = normalizeBlockKey(key)
	if !IsBlockKey(key) || first == nil || second == nil || first == second {
		return false
	}
	node, parent, index := l.Find(key)
	if node == nil || node == first || node == second {
		return false
	}
	rangeParent, _ := findLayoutNodeParent(l.Root, first)
	if rangeParent == nil {
		return false
	}
	if secondParent, _ := findLayoutNodeParent(l.Root, second); secondParent != rangeParent {
		return false
	}
	parent.Children = slices.Delete(parent.Children, index, index+1)
	// Taking the block out may have shifted the parent's indexes, so where the range lies is worked out again rather
	// than being remembered from before.
	firstIndex := slices.Index(rangeParent.Children, first)
	secondIndex := slices.Index(rangeParent.Children, second)
	if firstIndex < 0 || secondIndex < 0 {
		// Can't happen, since both were found a moment ago and neither is the block that was just detached, but if it
		// somehow did, putting the block back is better than losing it.
		parent.Children = slices.Insert(parent.Children, index, node)
		return false
	}
	if firstIndex > secondIndex {
		firstIndex, secondIndex = secondIndex, firstIndex
	}
	group := containerNode(rangeParent.Type, fxp.One,
		slices.Clone(rangeParent.Children[firstIndex:secondIndex+1])...)
	var total fxp.Int
	for _, child := range group.Children {
		total += child.Weight
	}
	if total <= 0 {
		total = fxp.One
	}
	desired := layoutnode.Column
	if edge = edge.EnsureValid(); edge == layoutedge.Left || edge == layoutedge.Right {
		desired = layoutnode.Row
	}
	// The wrapper takes the place of the range, so it is the wrapper that inherits the width the range had; the two
	// things it holds share it out evenly.
	node.Weight = fxp.One
	wrapper := containerNode(desired, total)
	if edge == layoutedge.Left || edge == layoutedge.Top {
		wrapper.Children = []*SheetLayoutNode{node, group}
	} else {
		wrapper.Children = []*SheetLayoutNode{group, node}
	}
	rangeParent.Children = slices.Replace(rangeParent.Children, firstIndex, secondIndex+1, wrapper)
	l.EnsureValidity()
	return true
}

// MoveToBand relocates the block with the given key so that it becomes a band of its own at the given index among the
// root's bands. Returns true if the layout was altered.
func (l *SheetLayout) MoveToBand(key string, index int) bool {
	key = normalizeBlockKey(key)
	if !IsBlockKey(key) {
		return false
	}
	node, parent, at := l.Find(key)
	if node == nil {
		return false
	}
	parent.Children = slices.Delete(parent.Children, at, at+1)
	if parent == l.Root && index > at {
		// The band the block was in is gone, so everything after it shifted up by one.
		index--
	}
	index = min(max(index, 0), len(l.Root.Children))
	node.Weight = fxp.One
	l.Root.Children = slices.Insert(l.Root.Children, index, node)
	l.EnsureValidity()
	return true
}

// meanWeight returns the average of the weights of the given nodes.
func meanWeight(nodes []*SheetLayoutNode) fxp.Int {
	if len(nodes) == 0 {
		return fxp.One
	}
	var total fxp.Int
	for _, node := range nodes {
		total += node.Weight
	}
	if total = total.Div(fxp.FromInteger(len(nodes))); total <= 0 {
		return fxp.One
	}
	return total
}

// SetWeights assigns the weights of the given container's children, ignoring any that aren't greater than zero. Extra
// weights are ignored, as are children the weights don't reach.
func (l *SheetLayout) SetWeights(row *SheetLayoutNode, weights []fxp.Int) {
	if row == nil {
		return
	}
	for i, child := range row.Children {
		if i >= len(weights) {
			break
		}
		if weights[i] > 0 {
			child.Weight = weights[i]
		}
	}
	l.EnsureValidity()
}

// SetMinHeight sets the minimum height of the block with the given key. A height of zero restores the block's natural
// height. Returns true if the layout was altered.
func (l *SheetLayout) SetMinHeight(key string, h paper.Length) bool {
	node, _, _ := l.Find(key)
	if node == nil {
		return false
	}
	h.EnsureValidity()
	if node.MinHeight == h {
		return false
	}
	node.MinHeight = h
	return true
}

// Filtered returns a copy of this layout holding only the blocks the keep function accepts. The rejected blocks are
// recorded as hidden, so that validating the copy doesn't restore them.
func (l *SheetLayout) Filtered(keep func(key string) bool) *SheetLayout {
	clone := l.Clone()
	clone.Root = filterLayoutNode(clone.Root, keep)
	for _, key := range AllBlockKeys {
		if !keep(key) {
			clone.Hidden = append(clone.Hidden, key)
		}
	}
	clone.EnsureValidity()
	return clone
}

func filterLayoutNode(node *SheetLayoutNode, keep func(key string) bool) *SheetLayoutNode {
	if node == nil {
		return nil
	}
	if node.Type == layoutnode.Block {
		if !keep(node.Key) {
			return nil
		}
		return node
	}
	children := make([]*SheetLayoutNode, 0, len(node.Children))
	for _, child := range node.Children {
		if child = filterLayoutNode(child, keep); child != nil {
			children = append(children, child)
		}
	}
	node.Children = children
	return node
}

// layoutListBand pairs one of the root's bands with the list blocks it holds.
type layoutListBand struct {
	node *SheetLayoutNode
	keys []string
}

// listBands returns each of the root's bands that holds at least one list block, paired with the keys of those blocks
// in depth-first order.
func (l *SheetLayout) listBands() []layoutListBand {
	if l.Root == nil {
		return nil
	}
	bands := make([]layoutListBand, 0, len(l.Root.Children))
	for _, child := range l.Root.Children {
		keys := appendListBlockKeys(nil, child)
		if len(keys) != 0 {
			bands = append(bands, layoutListBand{node: child, keys: keys})
		}
	}
	return bands
}

func appendListBlockKeys(keys []string, node *SheetLayoutNode) []string {
	if node == nil {
		return keys
	}
	if node.Type == layoutnode.Block {
		if IsListBlockKey(node.Key) {
			keys = append(keys, node.Key)
		}
		return keys
	}
	for _, child := range node.Children {
		keys = appendListBlockKeys(keys, child)
	}
	return keys
}

// ListBands returns, for each of the root's bands that holds at least one list block, the keys of those blocks in
// depth-first order. Bands that hold no list block are omitted.
//
// This is the projection of the tree that matches the rows the old block layout setting described, and the one
// HTMLGridTemplate is built on. The sheet, the template dockable and the page exporter all lay their content out from
// the tree itself.
func (l *SheetLayout) ListBands() [][]string {
	bands := l.listBands()
	result := make([][]string, 0, len(bands))
	for _, band := range bands {
		result = append(result, band.keys)
	}
	return result
}

// HTMLGridTemplate returns the CSS grid template the HTML export templates lay their list blocks out with. Each band
// that holds exactly two list blocks side by side contributes a row naming both of them; every other list block gets a
// row of its own. Blocks that aren't in the tree at all are named at the end, so that a template that positions them
// still works.
func (l *SheetLayout) HTMLGridTemplate() string {
	var buffer strings.Builder
	emitted := make(map[string]bool, len(listBlockKeys))
	for _, band := range l.listBands() {
		if band.node.Type == layoutnode.Row && len(band.keys) == 2 {
			appendToGridTemplate(&buffer, band.keys[0], band.keys[1])
			emitted[band.keys[0]] = true
			emitted[band.keys[1]] = true
			continue
		}
		for _, key := range band.keys {
			appendToGridTemplate(&buffer, key, key)
			emitted[key] = true
		}
	}
	for _, key := range AllBlockKeys {
		if IsListBlockKey(key) && !emitted[key] {
			appendToGridTemplate(&buffer, key, key)
		}
	}
	return buffer.String()
}

func appendToGridTemplate(buffer *strings.Builder, left, right string) {
	buffer.WriteByte('"')
	buffer.WriteString(left)
	buffer.WriteByte(' ')
	buffer.WriteString(right)
	buffer.WriteByte('"')
	buffer.WriteByte('\n')
}
