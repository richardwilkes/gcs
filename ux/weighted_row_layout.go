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
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
)

var _ unison.Layout = &weightedRowLayout{}

// layoutMinBlockWidth is the narrowest a block is ever made, whether by the widths a row divides itself into or by
// dragging the divider between two blocks. It is deliberately far below anything a block's content would ask for: what
// the content wants has no say in how wide the block is, since a block whose width was floored at what it needed
// couldn't be made any narrower, and the tables the lists are drawn with re-fit their columns to whatever width they
// are given and so always need exactly the width they already have.
const layoutMinBlockWidth float32 = 24

// weightedRowLayout lays the children of its target out side by side, dividing the available width among them in
// proportion to their weights and giving each of them the full height of the row. The widths follow the weights alone,
// save that no child is made narrower than layoutMinBlockWidth; the width that pinning a child there costs comes out of
// the children that still have room to give. Content that no longer fits the width its block was given is clipped by
// that block: that is what the user asked for by making it narrower, and widening it again brings the content back.
// The height of the row is that of its tallest child, but never less than the row's own minimum height, and a child
// that carries a minimum height of its own in its unison.FlexLayoutData counts at that height. That is the same place a
// minimum height is read from when the parent is a column laid out by a unison.FlexLayout, so a block's minimum height
// lives in exactly one place no matter which kind of container holds it.
//
// A child marked square is the exception to the weights: its width is taken from the height of the row instead, so that
// its content -- its frame less its border insets -- comes out square whatever the page size and the fonts make of the
// rest of the row. See arrange for how the two are reconciled.
type weightedRowLayout struct {
	weights   []fxp.Int
	square    []bool
	hSpacing  float32
	minHeight float32
}

// weightOf returns the weight of the child at the given index. Anything that isn't a usable weight counts as 1.
func (r *weightedRowLayout) weightOf(index int) float32 {
	if index < len(r.weights) && r.weights[index] > 0 {
		return r.weights[index].AsFloat[float32]()
	}
	return 1
}

// squareOf returns whether the child at the given index takes its width from the height of the row rather than from its
// weight.
func (r *weightedRowLayout) squareOf(index int) bool {
	return index < len(r.square) && r.square[index]
}

// setSquare records whether the child at the given index takes its width from the height of the row. A child the layout
// wasn't told about is left alone, since it has no slot to record it in.
func (r *weightedRowLayout) setSquare(index int, square bool) {
	if index >= 0 && index < len(r.square) {
		r.square[index] = square
	}
}

// widths returns the width to give each of the given number of children when there is the given amount of horizontal
// space to divide among them, weights alone. Nothing is measured, which is what lets a divider be dragged anywhere
// across the row no matter what the blocks either side of it hold. A square child is not honored here; see arrange.
func (r *weightedRowLayout) widths(count int, avail float32) []float32 {
	widths := make([]float32, count)
	if count == 0 {
		return widths
	}
	indexes := make([]int, count)
	for i := range indexes {
		indexes[i] = i
	}
	r.distribute(widths, indexes, avail-r.hSpacing*float32(count-1))
	return widths
}

// distribute hands the given pool of width out among the children at the given indexes, in proportion to their weights,
// writing what each of them gets into widths. No child is made narrower than layoutMinBlockWidth; what pinning one
// there costs comes out of the children that still have room to give, and the widths add up to exactly the pool.
func (r *weightedRowLayout) distribute(widths []float32, indexes []int, pool float32) {
	count := len(indexes)
	if count == 0 {
		return
	}
	if pool <= layoutMinBlockWidth*float32(count) {
		// There isn't room for even the narrowest each child is allowed to be, so the row runs off the end of the space
		// it was given rather than squeezing its children into nothing.
		for _, i := range indexes {
			widths[i] = layoutMinBlockWidth
		}
		return
	}
	pinned := make([]bool, count)
	// Each pass can only pin children that fell below the floor, so no more passes than there are children are ever
	// needed, plus the one that finds nothing left to pin.
	for range count + 1 {
		remaining := pool
		var totalWeight float32
		for at, i := range indexes {
			if pinned[at] {
				remaining -= widths[i]
			} else {
				totalWeight += r.weightOf(i)
			}
		}
		if totalWeight <= 0 {
			break
		}
		last := -1
		var used float32
		for at, i := range indexes {
			if pinned[at] {
				continue
			}
			widths[i] = remaining * r.weightOf(i) / totalWeight
			used += widths[i]
			last = i
		}
		// The rounding residual goes to the last child that hasn't been pinned, so that the widths add up to exactly
		// the space that was available.
		widths[last] += remaining - used
		pinnedOne := false
		for at, i := range indexes {
			if !pinned[at] && widths[i] < layoutMinBlockWidth {
				widths[i] = layoutMinBlockWidth
				pinned[at] = true
				pinnedOne = true
			}
		}
		if !pinnedOne {
			break
		}
	}
}

// heightOf returns the height the row comes to with its children at the given widths: the tallest of them, measured at
// the width each is actually given, with a child that carries a minimum height of its own counted at that height, and
// never less than the row's own minimum height. A square child counts here like any other: what it asks for -- the
// portrait's fixed natural height, or a minimum height of its own -- is a floor the row can't go below, even though
// what the row settles on is then what decides how wide that child is.
func (r *weightedRowLayout) heightOf(children []*unison.Panel, widths []float32) float32 {
	height := r.minHeight
	for i, child := range children {
		_, childPref, _ := child.Sizes(geom.Size{Width: widths[i]})
		childHeight := childPref.Height
		if data, ok := child.LayoutData().(*unison.FlexLayoutData); ok {
			childHeight = max(childHeight, data.MinSize.Height)
		}
		height = max(height, childHeight)
	}
	return height
}

// split returns the indexes of the children that take their width from the height of the row and the indexes of the
// ones that divide up what is left, or nil for both when the row has nothing to reconcile: no square child at all, or
// nothing but square children. Telling that apart costs nothing, which is what lets a row that has no square child in
// it hand out its widths without measuring anything.
func (r *weightedRowLayout) split(count int) (squares, others []int) {
	found := false
	for i := range count {
		if r.squareOf(i) {
			found = true
			break
		}
	}
	if !found {
		return nil, nil
	}
	squares = make([]int, 0, count)
	others = make([]int, 0, count)
	for i := range count {
		if r.squareOf(i) {
			squares = append(squares, i)
		} else {
			others = append(others, i)
		}
	}
	if len(others) == 0 {
		return nil, nil
	}
	return squares, others
}

// widthsAt returns the width to give each of the given children when the row has the given amount of space to divide
// among them and stands at the given height. That is the weights alone unless a square child has to have its width
// worked out from the height, which is the only case where anything has to be measured to place the children. The
// arrangement the row was measured with is what decides it when the row stands at the height that measurement arrived
// at, so that measuring and placing can't disagree. A row standing at any other height -- one a column has stretched
// to the bottom of a taller sibling, say -- squares its square children at the height it actually has instead, since
// that is the height every child of it is about to be given, and the other children divide up what is then left.
func (r *weightedRowLayout) widthsAt(children []*unison.Panel, avail, height float32) []float32 {
	count := len(children)
	squares, others := r.split(count)
	if squares == nil {
		return r.widths(count, avail)
	}
	widths, arranged := r.arrange(children, avail)
	if height != arranged {
		r.fitSquares(children, widths, squares, others, avail-r.hSpacing*float32(count-1), height)
	}
	return widths
}

// squareWidth returns the width that makes the given child's content square when the row is the given height: the
// height less the insets that stand above and below the content, plus the ones that stand either side of it. A child
// with no border has no insets and simply takes the row's height.
func squareWidth(child *unison.Panel, height float32) float32 {
	var insets geom.Insets
	if b := child.Border(); b != nil {
		insets = b.Insets()
	}
	return height - insets.Height() + insets.Width()
}

// arrange works out the width to give each child and the height the row comes to at those widths.
//
// With no square child among them -- or with nothing but square children, since then there is nothing left to derive a
// height from beyond what the square children themselves ask for -- this is simply the weights and the height of the
// tallest child. Otherwise the two have to be reconciled, since a square child's width comes from the height and the
// height comes from the widths:
//
//  1. Every child is given the width its weight calls for, and the row is measured at those widths.
//  2. Each square child is given the width that squares its content at that height, floored at layoutMinBlockWidth and
//     capped so that the rest of the row keeps its floors, and what is left over is handed to the other children in
//     proportion to their weights. The row is then measured again, since the children that just changed width may wrap
//     their content differently and so want a different height.
//  3. If the height moved, the square widths are worked out once more from the height that was actually arrived at, so
//     that the height the row reports and the widths it hands out agree about the square children, and the row is
//     measured again. This is repeated, up to maxSquareFitPasses fits in all, until the height stops moving.
//
// Each pass chases the wrap the one before it caused, and the height only ever moves the same way from one pass to the
// next, since a child that wraps only grows as it narrows and a square child only widens as the row grows. The layout
// has to be a pure function of what it is given so that measuring and placing can't disagree, so rather than iterating
// to a fixed point it stops after a set number of passes. A row whose height was still moving then reports the greater
// of the height its square children were fitted at and the height its children ask for at the widths it is handing
// out, so that nothing is ever clipped. What is left is that the square children come out short of square by however
// much the last pass would still have moved the height: a little squareness is traded for a bounded cost, and a row
// that is still moving after that many passes is one whose wrapping is chasing its own tail.
func (r *weightedRowLayout) arrange(children []*unison.Panel, avail float32) (widths []float32, height float32) {
	count := len(children)
	widths = r.widths(count, avail)
	if count == 0 {
		return widths, r.minHeight
	}
	height = r.heightOf(children, widths)
	squares, others := r.split(count)
	if squares == nil {
		return widths, height
	}
	pool := avail - r.hSpacing*float32(count-1)
	r.fitSquares(children, widths, squares, others, pool, height)
	for pass := 1; ; pass++ {
		next := r.heightOf(children, widths)
		if next == height {
			break
		}
		if pass == maxSquareFitPasses {
			height = max(height, next)
			break
		}
		// The children that aren't square have just changed width and one of them now wants a different height, so the
		// square children are worked out again from the height the row actually arrived at.
		height = next
		r.fitSquares(children, widths, squares, others, pool, height)
	}
	return widths, height
}

// maxSquareFitPasses is how many times arrange will fit the square children to the height of the row before settling
// for the height it has; see arrange for what stopping there leaves behind.
const maxSquareFitPasses = 4

// fitSquares gives each of the square children the width that squares its content at the given height and hands
// whatever of the pool that leaves to the other children in proportion to their weights, writing the results into
// widths. Nothing is measured here, so this costs nothing to repeat.
func (r *weightedRowLayout) fitSquares(children []*unison.Panel, widths []float32, squares, others []int,
	pool, height float32,
) {
	r.applySquareWidths(children, widths, squares, len(others), pool, height)
	var used float32
	for _, i := range squares {
		used += widths[i]
	}
	r.distribute(widths, others, pool-used)
}

// applySquareWidths gives each of the square children the width that squares its content at the given height, leaving
// room for the other children to keep the narrowest they may be. The square children are served in order, so an earlier
// one gets what it asks for and a later one gets whatever is still going.
func (r *weightedRowLayout) applySquareWidths(children []*unison.Panel, widths []float32, squares []int,
	otherCount int, pool, height float32,
) {
	room := pool - layoutMinBlockWidth*float32(otherCount)
	for at, i := range squares {
		// Whatever this child takes has to leave the floor for each of the square children still to come as well.
		available := room - layoutMinBlockWidth*float32(len(squares)-at-1)
		widths[i] = max(min(squareWidth(children[i], height), available), layoutMinBlockWidth)
		room -= widths[i]
	}
}

// LayoutSizes implements unison.Layout.
func (r *weightedRowLayout) LayoutSizes(target *unison.Panel, hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
	var insets geom.Size
	if b := target.Border(); b != nil {
		insets = b.Insets().Size()
		hint = hint.Sub(insets).Max(geom.Size{})
	}
	children := target.Children()
	count := len(children)
	prefSize.Height = r.minHeight
	if count == 0 {
		prefSize = prefSize.Add(insets)
		return prefSize, prefSize, prefSize
	}
	var widths []float32
	if hint.Width > 0 {
		widths, prefSize.Height = r.arrange(children, hint.Width)
	} else {
		// Without a width to divide there is nothing for a square child's width to come from either, so every child is
		// asked what it would like to be.
		widths = make([]float32, count)
		for i, child := range children {
			_, childPref, _ := child.Sizes(geom.Size{})
			widths[i] = childPref.Width
		}
		prefSize.Height = r.heightOf(children, widths)
	}
	for _, width := range widths {
		prefSize.Width += width
	}
	spacing := r.hSpacing * float32(count-1)
	prefSize.Width += spacing
	// The row asks for no more than the floor of each of its children, so that shrinking it moves the divider between
	// them rather than being refused.
	minSize.Width = layoutMinBlockWidth*float32(count) + spacing
	minSize.Height = prefSize.Height
	prefSize = prefSize.Add(insets)
	minSize = minSize.Add(insets)
	return minSize, prefSize, prefSize
}

// PerformLayout implements unison.Layout.
func (r *weightedRowLayout) PerformLayout(target *unison.Panel) {
	children := target.Children()
	if len(children) == 0 {
		return
	}
	rect := target.ContentRect(false)
	// Every child is given the full height of the row, so that is the height a square child's width has to come from;
	// see widthsAt for how that squares with the arrangement the row was measured with.
	widths := r.widthsAt(children, rect.Width, rect.Height)
	x := rect.X
	for i, child := range children {
		child.SetFrameRect(geom.NewRect(x, rect.Y, widths[i], rect.Height))
		x += widths[i] + r.hSpacing
	}
}
