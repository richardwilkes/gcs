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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
)

// newSizedPanel returns a panel that claims the given sizes, so that a layout can be exercised without any real
// content. The height is what the panel asks for no matter how wide it is made.
func newSizedPanel(minWidth, prefWidth, height float32) *unison.Panel {
	return newMeasuredPanel(minWidth, prefWidth, func(_ float32) float32 { return height })
}

// newMeasuredPanel returns a panel whose height depends on the width it is measured at, the way a panel that wraps its
// content does.
func newMeasuredPanel(minWidth, prefWidth float32, heightAt func(width float32) float32) *unison.Panel {
	p := unison.NewPanel()
	p.SetSizer(func(hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
		width := hint.Width
		if width <= 0 {
			width = prefWidth
		}
		minSize = geom.NewSize(minWidth, heightAt(width))
		prefSize = geom.NewSize(width, heightAt(width))
		return minSize, prefSize, prefSize
	})
	return p
}

// newTestRow returns a panel laid out by a weightedRowLayout holding the given children.
func newTestRow(weights []fxp.Int, minHeight float32, children ...*unison.Panel) (*unison.Panel, *weightedRowLayout) {
	row := unison.NewPanel()
	layout := &weightedRowLayout{weights: weights, hSpacing: 1, minHeight: minHeight}
	row.SetLayout(layout)
	for _, child := range children {
		row.AddChild(child)
	}
	return row, layout
}

// newTestSquareRow returns a panel laid out by a weightedRowLayout holding the given children, with the ones the square
// flags name taking their widths from the height of the row rather than from their weights.
func newTestSquareRow(weights []fxp.Int, square []bool, children ...*unison.Panel) (*unison.Panel, *weightedRowLayout) {
	row, layout := newTestRow(weights, 0, children...)
	layout.square = square
	return row, layout
}

// squareTestInsets are the insets of the bordered panel the square tests use. They are deliberately lopsided, so that a
// width worked out from the height by any route other than the right one comes out wrong.
var squareTestInsets = geom.NewInsets(10, 2, 3, 4)

// newSquareTestPanel returns a panel with a border around it whose natural height is fixed however wide it is made, the
// way the portrait's is, so that the height of the row it is in has to come from what stands beside it.
func newSquareTestPanel(height float32) *unison.Panel {
	p := newSizedPanel(0, 10, height)
	p.SetBorder(unison.NewEmptyBorder(squareTestInsets))
	return p
}

// TestWeightedRowLayoutDividesByWeight verifies that the available width, less the spacing between the children, is
// handed out in proportion to the weights, and that the children are placed one after the other across the row.
func TestWeightedRowLayoutDividesByWeight(t *testing.T) {
	c := check.New(t)
	row, layout := newTestRow([]fxp.Int{fxp.One, fxp.Two, fxp.One}, 0,
		newSizedPanel(0, 10, 30),
		newSizedPanel(0, 10, 30),
		newSizedPanel(0, 10, 30))
	widths := layout.widths(len(row.Children()), 100)
	c.Equal([]float32{24.5, 49, 24.5}, widths, "the width less the spacing must be split 1:2:1")

	row.SetFrameRect(geom.NewRect(0, 0, 100, 40))
	layout.PerformLayout(row)
	children := row.Children()
	c.Equal(geom.NewRect(0, 0, 24.5, 40), children[0].FrameRect())
	c.Equal(geom.NewRect(25.5, 0, 49, 40), children[1].FrameRect(), "each child follows the last, plus the spacing")
	c.Equal(geom.NewRect(75.5, 0, 24.5, 40), children[2].FrameRect())
}

// TestWeightedRowLayoutPinsToTheFloor verifies that a child whose share of the width would leave it narrower than the
// one width no block may go below is pinned there, and that what pinning it costs comes out of the children that still
// have room to give, with the widths still adding up to exactly what was available.
func TestWeightedRowLayoutPinsToTheFloor(t *testing.T) {
	c := check.New(t)
	row, layout := newTestRow([]fxp.Int{fxp.FromInteger(10), fxp.One}, 0,
		newSizedPanel(0, 10, 30),
		newSizedPanel(0, 10, 30))
	widths := layout.widths(len(row.Children()), 100)
	c.Equal([]float32{99 - layoutMinBlockWidth, layoutMinBlockWidth}, widths,
		"the second child is pinned at the floor and the first gets the rest")
	c.Equal(float32(99), widths[0]+widths[1], "the widths must add up to the space less the spacing")
}

// TestWeightedRowLayoutIgnoresTheWidthItsContentWants verifies that a child is given the share of the row its weight
// calls for however wide its content would like to be, since a block that could not be made narrower than its content
// could not be resized at all. The page lists are exactly that case: their tables re-fit their columns to whatever
// width they are given, so what they ask for is always the width they already have.
func TestWeightedRowLayoutIgnoresTheWidthItsContentWants(t *testing.T) {
	c := check.New(t)
	row, layout := newTestRow([]fxp.Int{fxp.One, fxp.One}, 0,
		newSizedPanel(50, 50, 30),
		newSizedPanel(50, 50, 30))
	widths := layout.widths(len(row.Children()), 101)
	c.Equal([]float32{50, 50}, widths, "an even split gives each child exactly the width its content asked for")

	layout.weights = []fxp.Int{fxp.Three, fxp.One}
	widths = layout.widths(len(row.Children()), 101)
	c.Equal([]float32{75, 25}, widths, "changing the weights must re-proportion the row regardless of the content")
}

// TestWeightedRowLayoutOverflowsWhenTheFloorsDoNotFit verifies that a row that cannot fit even the narrowest each of
// its children may be hands out that width anyway, running off the end rather than squeezing them into nothing.
func TestWeightedRowLayoutOverflowsWhenTheFloorsDoNotFit(t *testing.T) {
	c := check.New(t)
	row, layout := newTestRow([]fxp.Int{fxp.One, fxp.One}, 0,
		newSizedPanel(80, 80, 30),
		newSizedPanel(80, 80, 30))
	widths := layout.widths(len(row.Children()), 40)
	c.Equal([]float32{layoutMinBlockWidth, layoutMinBlockWidth}, widths, "both children get the narrowest they may be")

	minSize, _, _ := layout.LayoutSizes(row, geom.NewSize(100, 0))
	c.Equal(2*layoutMinBlockWidth+1, minSize.Width,
		"the row's minimum width is the floor of each of its children, plus the spacing")
}

// TestWeightedRowLayoutHeightIsTheTallestChild verifies that the row is as tall as its tallest child, measured at the
// width that child is actually given, but never shorter than a minimum height carried by a child or by the row itself.
func TestWeightedRowLayoutHeightIsTheTallestChild(t *testing.T) {
	c := check.New(t)
	tall := newSizedPanel(0, 10, 50)
	short := newSizedPanel(0, 10, 30)
	row, layout := newTestRow([]fxp.Int{fxp.One, fxp.One}, 0, short, tall)
	_, prefSize, _ := layout.LayoutSizes(row, geom.NewSize(100, 0))
	c.Equal(float32(50), prefSize.Height, "the row is as tall as its tallest child")

	tall.SetLayoutData(&unison.FlexLayoutData{MinSize: geom.Size{Height: 70}})
	_, prefSize, _ = layout.LayoutSizes(row, geom.NewSize(100, 0))
	c.Equal(float32(70), prefSize.Height, "a child's own minimum height counts toward the row's height")

	layout.minHeight = 90
	_, prefSize, _ = layout.LayoutSizes(row, geom.NewSize(100, 0))
	c.Equal(float32(90), prefSize.Height, "the row's own minimum height counts too")

	layout.minHeight = 0
	tall.SetLayoutData(&unison.FlexLayoutData{})
	_, prefSize, _ = layout.LayoutSizes(row, geom.NewSize(100, 0))
	c.Equal(float32(50), prefSize.Height, "the row is back to the height of its tallest child")
}

// TestWeightedRowLayoutMeasuresHeightAtTheGivenWidth verifies that a child that wraps its content is asked how tall it
// is at the width the row gives it, rather than at whatever width it would have chosen for itself.
func TestWeightedRowLayoutMeasuresHeightAtTheGivenWidth(t *testing.T) {
	c := check.New(t)
	wrapping := newMeasuredPanel(0, 200, func(width float32) float32 { return 4000 / width })
	row, layout := newTestRow([]fxp.Int{fxp.One, fxp.One}, 0, wrapping, newSizedPanel(0, 10, 10))
	_, prefSize, _ := layout.LayoutSizes(row, geom.NewSize(101, 0))
	c.Equal(float32(80), prefSize.Height, "the wrapping child is measured at the 50 pixels it is given")
}

// TestWeightedRowLayoutWithoutAWidthUsesPreferredWidths verifies that a row asked for its size without a width to work
// from reports what its children would like to be, rather than dividing a width it doesn't have.
func TestWeightedRowLayoutWithoutAWidthUsesPreferredWidths(t *testing.T) {
	c := check.New(t)
	row, layout := newTestRow([]fxp.Int{fxp.One, fxp.Three}, 0,
		newSizedPanel(10, 40, 30),
		newSizedPanel(20, 60, 20))
	minSize, prefSize, maxSize := layout.LayoutSizes(row, geom.Size{})
	c.Equal(float32(101), prefSize.Width, "the preferred widths of the children, plus the spacing")
	c.Equal(2*layoutMinBlockWidth+1, minSize.Width, "the floor of each of the children, plus the spacing")
	c.Equal(float32(30), prefSize.Height)
	c.Equal(prefSize, maxSize, "a row is never given more than it asked for")
}

// TestWeightedRowLayoutSquareChildTakesItsWidthFromTheHeight verifies that a square child is given the width that makes
// its content -- its frame less its border insets -- as wide as the row is tall less those insets, that the row is as
// tall as the child beside it, and that what is left over goes to that child.
func TestWeightedRowLayoutSquareChildTakesItsWidthFromTheHeight(t *testing.T) {
	c := check.New(t)
	square := newSquareTestPanel(30)
	tall := newSizedPanel(0, 10, 80)
	row, layout := newTestSquareRow([]fxp.Int{fxp.One, fxp.One}, []bool{true, false}, square, tall)
	widths, height := layout.arrange(row.Children(), 200)
	c.Equal(float32(80), height, "the row is as tall as the child that isn't square")
	c.Equal(80-squareTestInsets.Height()+squareTestInsets.Width(), widths[0],
		"the square child's width is the row's height, less the insets above and below its content and plus the ones "+
			"either side of it")
	c.Equal(199-widths[0], widths[1], "everything the square child didn't take goes to the other child")

	_, prefSize, _ := layout.LayoutSizes(row, geom.NewSize(200, 0))
	c.Equal(float32(80), prefSize.Height, "the row reports the height its square child was measured against")
	c.Equal(float32(200), prefSize.Width, "the widths still add up to the space the row was given")

	row.SetFrameRect(geom.NewRect(0, 0, 200, prefSize.Height))
	layout.PerformLayout(row)
	c.Equal(geom.NewRect(0, 0, widths[0], 80), square.FrameRect(), "placing the children must agree with measuring them")
	content := square.ContentRect(false).Size
	c.Equal(content.Height, content.Width, "the square child's content must come out square")
}

// TestWeightedRowLayoutSquareChildDrivesTheHeight verifies that a square child's own minimum height counts toward the
// height of the row like any other child's, and that its width then follows that height.
func TestWeightedRowLayoutSquareChildDrivesTheHeight(t *testing.T) {
	c := check.New(t)
	square := newSquareTestPanel(30)
	square.SetLayoutData(&unison.FlexLayoutData{MinSize: geom.Size{Height: 120}})
	short := newSizedPanel(0, 10, 80)
	row, layout := newTestSquareRow([]fxp.Int{fxp.One, fxp.One}, []bool{true, false}, square, short)
	widths, height := layout.arrange(row.Children(), 200)
	c.Equal(float32(120), height, "a square child's minimum height raises the row like any other child's")
	c.Equal(120-squareTestInsets.Height()+squareTestInsets.Width(), widths[0],
		"and the width it is then given comes from that height")
	c.Equal(199-widths[0], widths[1])
}

// TestWeightedRowLayoutOtherChildrenShareWhatIsLeft verifies that the children that aren't square divide up whatever the
// square ones didn't take, in proportion to their own weights.
func TestWeightedRowLayoutOtherChildrenShareWhatIsLeft(t *testing.T) {
	c := check.New(t)
	square := newSquareTestPanel(30)
	row, layout := newTestSquareRow([]fxp.Int{fxp.One, fxp.One, fxp.Three}, []bool{true, false, false},
		square, newSizedPanel(0, 10, 80), newSizedPanel(0, 10, 40))
	widths, height := layout.arrange(row.Children(), 200)
	c.Equal(float32(80), height)
	squareWidth := 80 - squareTestInsets.Height() + squareTestInsets.Width()
	c.Equal(squareWidth, widths[0])
	remainder := 198 - squareWidth
	c.Equal(remainder/4, widths[1], "the rest of the row is split 1:3")
	c.Equal(3*remainder/4, widths[2])
}

// TestWeightedRowLayoutSquareChildIgnoresThePageWidth verifies that changing how much width the row has to divide up
// leaves a square child exactly as it was, since its width comes from the height instead. This is what makes the
// portrait's picture area square on a page of any size rather than only on the one its weight was tuned for.
func TestWeightedRowLayoutSquareChildIgnoresThePageWidth(t *testing.T) {
	c := check.New(t)
	square := newSquareTestPanel(30)
	row, layout := newTestSquareRow([]fxp.Int{fxp.One, fxp.One}, []bool{true, false},
		square, newSizedPanel(0, 10, 80))
	letter, letterHeight := layout.arrange(row.Children(), 576)
	a4, a4Height := layout.arrange(row.Children(), 559)
	c.Equal(letterHeight, a4Height, "the height doesn't depend on the width here")
	c.Equal(letter[0], a4[0], "the square child must be the same width on either page")
	c.Equal(letter[1]-17, a4[1], "the whole of the difference comes out of the child that isn't square")
}

// TestWeightedRowLayoutAllSquareFallsBackToWeights verifies that a row whose children are all square divides itself up
// by weight after all: there is nothing beside them for the height to come from, so there is nothing to take a width
// from either.
func TestWeightedRowLayoutAllSquareFallsBackToWeights(t *testing.T) {
	c := check.New(t)
	row, layout := newTestSquareRow([]fxp.Int{fxp.One, fxp.Three}, []bool{true, true},
		newSquareTestPanel(30), newSquareTestPanel(30))
	widths, height := layout.arrange(row.Children(), 200)
	c.Equal([]float32{49.75, 149.25}, widths, "the width less the spacing must be split 1:3")
	c.Equal(float32(30), height, "the row is as tall as the tallest of them")
}

// TestWeightedRowLayoutSquareChildAgainstWrappingContent verifies that a square child beside a child that wraps its
// content ends up with a width that matches the height the row reports, even though giving it that width changed how
// the other child wraps, and that measuring the row twice gives the same answer both times.
func TestWeightedRowLayoutSquareChildAgainstWrappingContent(t *testing.T) {
	c := check.New(t)
	square := newSizedPanel(0, 10, 30)
	wrapping := newMeasuredPanel(0, 200, func(width float32) float32 { return 4000 / width })
	row, layout := newTestSquareRow([]fxp.Int{fxp.One, fxp.One}, []bool{true, false}, square, wrapping)
	widths, height := layout.arrange(row.Children(), 200)
	c.Equal(height, widths[0], "with no border to allow for, the square child's width is the height of the row")
	c.True(height < 4000/99.5, "the second pass must have caught up with the wrapping the first one caused")
	again, againHeight := layout.arrange(row.Children(), 200)
	c.Equal(widths, again, "the arrangement must be the same every time it is worked out")
	c.Equal(height, againHeight)
}

// TestWeightedRowLayoutWithoutChildren verifies that an empty row is harmless: it asks for nothing but its own minimum
// height and lays nothing out.
func TestWeightedRowLayoutWithoutChildren(t *testing.T) {
	c := check.New(t)
	row, layout := newTestRow(nil, 25)
	minSize, prefSize, _ := layout.LayoutSizes(row, geom.NewSize(100, 0))
	c.Equal(geom.NewSize(0, 25), prefSize)
	c.Equal(geom.NewSize(0, 25), minSize)
	row.SetFrameRect(geom.NewRect(0, 0, 100, 25))
	layout.PerformLayout(row)
	c.Equal(0, len(row.Children()))
}

// TestWeightedRowLayoutSquareChildFollowsTheHeightItIsGiven verifies that a row placed taller than it asked for -- as
// one nested in a column beside something taller is, since a column stretches its children to its bottom -- squares its
// square child at the height it actually has rather than at the height it was measured at, and hands what that leaves
// to the other child.
func TestWeightedRowLayoutSquareChildFollowsTheHeightItIsGiven(t *testing.T) {
	c := check.New(t)
	square := newSquareTestPanel(30)
	other := newSizedPanel(0, 10, 80)
	row, layout := newTestSquareRow([]fxp.Int{fxp.One, fxp.One}, []bool{true, false}, square, other)
	_, prefSize, _ := layout.LayoutSizes(row, geom.NewSize(400, 0))
	c.Equal(float32(80), prefSize.Height, "the row asks for the height of the child that isn't square")

	row.SetFrameRect(geom.NewRect(0, 0, 400, 233))
	layout.PerformLayout(row)
	c.Equal(float32(233), square.FrameRect().Height, "every child is given the full height of the row")
	c.Equal(233-squareTestInsets.Height()+squareTestInsets.Width(), square.FrameRect().Width,
		"so the square child's width must come from that height")
	content := square.ContentRect(false).Size
	c.Equal(content.Height, content.Width, "and its content must come out square")
	c.Equal(399-square.FrameRect().Width, other.FrameRect().Width, "the other child gets what is left")

	row.SetFrameRect(geom.NewRect(0, 0, 400, 80))
	layout.PerformLayout(row)
	c.Equal(80-squareTestInsets.Height()+squareTestInsets.Width(), square.FrameRect().Width,
		"at the height it asked for, the row places its children exactly as it measured them")
}

// TestWeightedRowLayoutNeverReportsLessThanItsChildrenNeed verifies that the height the row reports is never less than
// what a child asks for at the width the row hands it, however many passes it would take the wrapping to settle, so
// that nothing is ever clipped: when the passes run out, the square child comes out short of square instead.
func TestWeightedRowLayoutNeverReportsLessThanItsChildrenNeed(t *testing.T) {
	c := check.New(t)
	square := newSizedPanel(0, 10, 10)
	// The other child's height rises as steeply as its width falls, and the row is just wide enough that the two of
	// them settle only slowly: each pass widens the square child a little, which narrows this one and raises the row
	// a little, and so on, well past the number of passes the layout allows itself.
	wrapping := newMeasuredPanel(0, 200, func(width float32) float32 { return 10000 / width })
	row, layout := newTestSquareRow([]fxp.Int{fxp.One, fxp.Three}, []bool{true, false}, square, wrapping)
	widths, height := layout.arrange(row.Children(), 200)
	needed := layout.heightOf(row.Children(), widths)
	c.True(height >= needed, "the row reports %v but its children need %v at the widths it hands out", height, needed)
	c.True(widths[0] <= height, "the square child may be short of square, but never wider than the row is tall")
	c.True(widths[0] > 49.75, "the passes must have moved the square child from where the weights alone put it")

	_, prefSize, _ := layout.LayoutSizes(row, geom.NewSize(200, 0))
	c.Equal(height, prefSize.Height, "the row must report the height it arranged itself at")
	row.SetFrameRect(geom.NewRect(0, 0, 200, prefSize.Height))
	layout.PerformLayout(row)
	c.Equal(widths[0], square.FrameRect().Width, "placing must hand out the widths measuring arrived at")
	c.Equal(widths[1], wrapping.FrameRect().Width)
	_, wrappingPref, _ := wrapping.Sizes(geom.Size{Width: wrapping.FrameRect().Width})
	c.True(wrappingPref.Height <= wrapping.FrameRect().Height, "nothing may be clipped")
}
