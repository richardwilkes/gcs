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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// randomizedFieldCount is the number of randomizer buttons the Description and Identity blocks hold between them:
// gender, age, birthday, height, weight, hair, eyes, skin and handedness on the one, and the name on the other.
const randomizedFieldCount = 10

// randomizedRow is a randomizer button on a sheet block, together with the title of its row and the field it fills in.
type randomizedRow struct {
	title  string
	button *unison.Button
	field  SelectableTextField
}

// randomizedRows returns every randomizer row within root. The helpers that build these rows add a wrapper holding the
// randomizer button and the label, then the field, so the field is always the wrapper's next sibling.
func randomizedRows(t *testing.T, root *unison.Panel) []randomizedRow {
	t.Helper()
	var rows []randomizedRow
	for _, button := range panelsOfType[*unison.Button](root) {
		if drawable, ok := button.Drawable.(*unison.DrawableSVG); !ok || drawable.SVG != svg.Randomize {
			continue
		}
		wrapper := button.Parent()
		label, ok := firstPanelOfType[*unison.Label](wrapper)
		if !ok {
			t.Fatal("a randomizer wrapper holds no label")
		}
		siblings := wrapper.Parent().Children()
		var field SelectableTextField
		for i, one := range siblings {
			if one == wrapper {
				if i+1 >= len(siblings) {
					t.Fatalf("no field follows the %q randomizer", label.String())
				}
				if field, ok = siblings[i+1].Self.(SelectableTextField); !ok {
					t.Fatalf("the child after the %q randomizer is a %T, not a field", label.String(), siblings[i+1].Self)
				}
				break
			}
		}
		rows = append(rows, randomizedRow{title: label.String(), button: button, field: field})
	}
	return rows
}

// expectedFieldText returns the text a field should show for the value its getter currently returns.
func expectedFieldText(t *testing.T, field SelectableTextField) string {
	t.Helper()
	switch f := field.(type) {
	case *StringField:
		return f.get()
	case *LengthField:
		return f.Format(f.get())
	case *WeightField:
		return f.Format(f.get())
	default:
		t.Fatalf("unexpected field type %T", field)
		return ""
	}
}

// TestRandomizedPageFieldsShowWhatTheyStore verifies that every randomizer on the Description and Identity blocks
// stores a new value on the character, shows that value in the field beside it, and marks the sheet modified with the
// field as the source, and that each of those fields is flagged SkipDeepSync so the edit does not trigger a deep sync.
func TestRandomizedPageFieldsShowWhatTheyStore(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	root := newPageUndoRoot()
	targetMgr := NewTargetMgr(root)
	root.AddChild(NewDescriptionPanel(entity, targetMgr))
	root.AddChild(NewIdentityPanel(entity, targetMgr))

	rows := randomizedRows(t, root.AsPanel())
	c.Equal(randomizedFieldCount, len(rows), "every randomizable field has a randomizer button")

	for _, row := range rows {
		panel := row.field.AsPanel()
		_, skip := panel.ClientData()[SkipDeepSync]
		c.True(skip, "the %q field is flagged SkipDeepSync", row.title)
		textOf, ok := panel.Self.(interface{ Text() string })
		if !ok {
			t.Fatalf("the %q field has no text: %T", row.title, panel.Self)
		}
		c.Equal(expectedFieldText(t, row.field), textOf.Text(), "the %q field shows the stored value before randomizing",
			row.title)

		root.modifiedBy = nil
		row.button.ClickCallback()
		text := textOf.Text()
		c.NotEqual("", text, "randomizing the %q field yields a value", row.title)
		c.Equal(expectedFieldText(t, row.field), text, "the %q field shows the value randomizing stored", row.title)
		c.True(len(root.modifiedBy) > 0, "randomizing the %q field marks the sheet modified", row.title)
		for _, src := range root.modifiedBy {
			c.True(src.AsPanel() == panel, "the %q field is the source of each modification, but got %T", row.title,
				src)
		}
	}
}

// TestPageFieldRowsAlternateLabelsAndFields verifies that the Description, Identity and Miscellaneous blocks lay out
// their rows as a label followed by its field. The Description block's banding is drawn from that alternation, so a
// helper that added the two in the wrong order would misplace every band below it.
func TestPageFieldRowsAlternateLabelsAndFields(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	targetMgr := NewTargetMgr(newPageUndoRoot())
	description := NewDescriptionPanel(entity, targetMgr)
	columns := description.Children()
	c.Equal(3, len(columns), "the Description block has three columns")
	rows := map[string]*unison.Panel{
		"identity":      NewIdentityPanel(entity, targetMgr).AsPanel(),
		"miscellaneous": NewMiscPanel(entity, targetMgr).AsPanel(),
	}
	for i, column := range columns {
		rows["description column "+string(rune('1'+i))] = column
	}
	for name, panel := range rows {
		children := panel.Children()
		c.True(len(children) > 0 && len(children)%2 == 0, "%s holds whole label/field pairs, but has %d children",
			name, len(children))
		for i, child := range children {
			if i%2 == 0 {
				_, isLabel := child.Self.(*unison.Label)
				isWrapper := buttonWithSVG(child, svg.Randomize) != nil
				c.True(isLabel || isWrapper, "%s child %d is a label or a randomizer wrapper, but is a %T", name, i,
					child.Self)
			} else {
				_, isLabel := child.Self.(*unison.Label)
				c.False(isLabel, "%s child %d is a field, but is a label", name, i)
			}
		}
	}
}
