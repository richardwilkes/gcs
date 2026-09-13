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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

func TestTraitEditorContainerTypes(t *testing.T) {
	for _, tc := range []struct {
		name  string
		owner gurps.DataOwner
		fixed bool
	}{
		{name: "sheet", owner: gurps.NewEntity()},
		{name: "template", owner: gurps.NewTemplate(), fixed: true},
		{name: "library", fixed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			trait := gurps.NewTrait(tc.owner, nil, true)
			e, content := buildEditorContent(nil, trait, initTraitEditor)
			popups := panelsOfType[*unison.PopupMenu[container.Type]](content)
			if len(popups) != 1 {
				t.Fatalf("got %d container popups", len(popups))
			}
			popup := popups[0]
			c.Equal(tc.fixed, popup.IndexOfItem(container.FixedCost) >= 0)
			for _, kind := range container.Types {
				if kind != container.FixedCost {
					c.True(popup.IndexOfItem(kind) >= 0)
				}
			}
			if tc.fixed {
				popup.Select(container.FixedCost)
				c.Equal(container.FixedCost, e.editorData.ContainerType)
			}
		})
	}
}

func TestTraitEditorFixedPoints(t *testing.T) {
	c := check.New(t)
	trait := gurps.NewTrait(gurps.NewTemplate(), nil, true)
	trait.ContainerType = container.FixedCost
	trait.FixedPoints = new(fxp.Five)
	var update func()
	e, content := buildEditorContent(nil, trait, func(e *editor[*gurps.Trait, *gurps.TraitEditData], p *unison.Panel) func() {
		update = initTraitEditor(e, p)
		return update
	})
	fields := panelsOfType[*DecimalField](content)
	if len(fields) != 1 {
		t.Fatalf("got %d decimal fields", len(fields))
	}
	field := fields[0]
	c.True(field.Enabled())
	c.Equal("5", field.Text())
	field.SetText("12.5")
	c.Equal(new(fxp.FromStringForced("12.5")), e.editorData.FixedPoints)
	c.Equal(new(fxp.Five), trait.FixedPoints, "editing must not mutate the source")
	field.SetText("")
	c.True(e.editorData.FixedPoints == nil)
	c.Equal("", field.tooltipTextForValidation())
	field.SetText("0")
	c.Equal(new(fxp.Int(0)), e.editorData.FixedPoints)
	field.lostFocus()
	c.Equal("0", field.Text())
	e.editorData.ApplyTo(trait)
	c.Equal(new(fxp.Int(0)), trait.FixedPoints)
	field.SetText("-5")
	c.Equal(new(-fxp.Five), e.editorData.FixedPoints)
	popup := panelsOfType[*unison.PopupMenu[container.Type]](content)[0]
	popup.Select(container.Group)
	update()
	c.False(field.Enabled())
	c.True(e.editorData.FixedPoints == nil)
	c.Equal("", field.Text())
	popup.Select(container.FixedCost)
	update()
	c.True(field.Enabled())
	c.True(e.editorData.FixedPoints == nil)
	field.SetText("0")
	c.Equal(new(fxp.Int(0)), e.editorData.FixedPoints)
	field.SetText("")
	field.lostFocus()
	c.Equal("", field.Text())
	c.True(e.editorData.FixedPoints == nil)
}

func TestTraitEditorClearFixedPoints(t *testing.T) {
	for _, text := range []string{"5", "0"} {
		t.Run(text, func(t *testing.T) {
			c := check.New(t)
			trait := gurps.NewTrait(gurps.NewTemplate(), nil, true)
			trait.ContainerType = container.FixedCost
			var update func()
			e, content := buildEditorContent(nil, trait, func(e *editor[*gurps.Trait, *gurps.TraitEditData], p *unison.Panel) func() {
				update = initTraitEditor(e, p)
				return update
			})
			field := panelsOfType[*DecimalField](content)[0]
			var clearButton *unison.Button
			for _, button := range panelsOfType[*unison.Button](content) {
				if button.Text.String() == "Clear" {
					clearButton = button
					break
				}
			}
			if clearButton == nil {
				t.Fatal("missing clear action")
			}
			c.True(clearButton.Enabled())
			field.SetText(text)
			c.True(e.editorData.FixedPoints != nil)
			clearButton.ClickCallback()
			c.Equal("", field.Text())
			c.True(e.editorData.FixedPoints == nil)
			field.lostFocus()
			e.editorData.ApplyTo(trait)
			c.True(trait.FixedPoints == nil)
			field.SetText("0")
			c.Equal(new(fxp.Int(0)), e.editorData.FixedPoints)
			field.SetText("")
			c.True(e.editorData.FixedPoints == nil)
			field.SetText("0")
			c.Equal(new(fxp.Int(0)), e.editorData.FixedPoints)
			popup := panelsOfType[*unison.PopupMenu[container.Type]](content)[0]
			popup.Select(container.Group)
			update()
			c.False(clearButton.Enabled())
			c.True(e.editorData.FixedPoints == nil)
			popup.Select(container.FixedCost)
			update()
			c.True(clearButton.Enabled())
		})
	}
}
