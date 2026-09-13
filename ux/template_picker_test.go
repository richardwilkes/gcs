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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestFixedCostPickerSelectionState(t *testing.T) {
	for _, tc := range []struct {
		name    string
		compare criteria.NumericComparison
		limit   int
		want    bool
	}{
		{name: "exact equal", compare: criteria.EqualsNumber, limit: 30, want: true},
		{name: "exact different", compare: criteria.EqualsNumber, limit: 10},
		{name: "at most below", compare: criteria.AtMostNumber, limit: 10},
		{name: "at most equal", compare: criteria.AtMostNumber, limit: 30, want: true},
		{name: "at most above", compare: criteria.AtMostNumber, limit: 40, want: true},
		{name: "at least above", compare: criteria.AtLeastNumber, limit: 40},
		{name: "at least equal", compare: criteria.AtLeastNumber, limit: 30, want: true},
		{name: "at least below", compare: criteria.AtLeastNumber, limit: 10, want: true},
		{name: "not equals equal", compare: criteria.NotEqualsNumber, limit: 30},
		{name: "not equals different", compare: criteria.NotEqualsNumber, limit: 10, want: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			parent := gurps.NewTrait(gurps.NewTemplate(), nil, true)
			parent.ContainerType = container.FixedCost
			parent.FixedPoints = new(fxp.FromInteger(30))
			parent.TemplatePicker.Type = picker.Points
			parent.TemplatePicker.Qualifier.Compare = tc.compare
			parent.TemplatePicker.Qualifier.Qualifier = fxp.FromInteger(tc.limit)
			for _, cost := range []int{10, 20, 99} {
				child := gurps.NewTrait(parent.DataOwner(), parent, false)
				child.BasePoints = fxp.FromInteger(cost)
				parent.Children = append(parent.Children, child)
			}
			selected := parent.Children[:2]
			total, matches := pickerSelectionState(&parent.TemplatePicker, selected)
			c.Equal(*parent.FixedPoints, total)
			c.Equal(tc.want, matches)
			// A different parent override must not change either the selected total or its validity.
			parent.FixedPoints = new(fxp.FromInteger(999))
			revisedTotal, revisedMatches := pickerSelectionState(&parent.TemplatePicker, selected)
			c.Equal(total, revisedTotal)
			c.Equal(matches, revisedMatches)
		})
	}
}

func TestPickerSelectionCountAndEmpty(t *testing.T) {
	c := check.New(t)
	tp := &gurps.TemplatePicker{Type: picker.Count}
	tp.Qualifier.Compare = criteria.EqualsNumber
	tp.Qualifier.Qualifier = fxp.One
	child := gurps.NewTrait(nil, nil, false)
	child.BasePoints = fxp.Ten
	total, matches := pickerSelectionState(tp, []*gurps.Trait{child})
	c.Equal(fxp.One, total)
	c.True(matches)
	total, matches = pickerSelectionState(tp, []*gurps.Trait(nil))
	c.Equal(fxp.Int(0), total)
	c.False(matches)
}

func TestNestedFixedCostPickerSelection(t *testing.T) {
	for _, points := range []fxp.Int{fxp.Five, 0, -fxp.Five} {
		t.Run(points.String(), func(t *testing.T) {
			c := check.New(t)
			child := gurps.NewTrait(gurps.NewTemplate(), nil, true)
			child.ContainerType = container.FixedCost
			child.FixedPoints = new(points)
			child.TemplatePicker.Type = picker.Points
			child.TemplatePicker.Qualifier.Compare = criteria.EqualsNumber
			child.TemplatePicker.Qualifier.Qualifier = fxp.Ten
			parentPicker := child.TemplatePicker
			parentPicker.Qualifier.Qualifier = points
			total, matches := pickerSelectionState(&parentPicker, []*gurps.Trait{child})
			c.Equal(points, total)
			c.True(matches)
			child.FixedPoints = nil
			c.Equal(fxp.Ten, rawPoints(child))
		})
	}
}
