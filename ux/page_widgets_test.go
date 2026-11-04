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

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/role"
)

// testPageBlock stands in for one of the sheet's block panels.
type testPageBlock struct {
	unison.Panel
}

// TestInitTitledPagePanel verifies the shared setup of a titled sheet page block: the panel becomes its own Self, gets
// the titled border with the standard insets inside it, a grid of the requested columns, layout data that fills its
// cell, the banded background and the tint only when asked for, and hands back the layout and layout data installed.
func TestInitTitledPagePanel(t *testing.T) {
	c := check.New(t)
	block := &testPageBlock{}
	layout, layoutData := initTitledPagePanel(block, "Title", 3, true, colors.TintIdentity)
	c.Equal(block, block.Self, "the block becomes its own Self")
	c.Equal(newTitledBlockBorder("Title").Insets().Add(titledPagePanelInsets), block.Border().Insets(),
		"the standard insets sit inside the titled border")
	installed, ok := block.Layout().(*blockLayout)
	c.True(ok, "a titled block's layout stretches its heading across the block")
	c.Equal(installed.FlexLayout, layout, "the layout handed back is the one installed")
	c.Equal(3, layout.Columns)
	c.Equal(float32(4), layout.HSpacing)
	c.Equal(block.LayoutData(), layoutData, "the layout data handed back is the one installed")
	c.Equal(align.Fill, layoutData.HAlign)
	c.Equal(align.Fill, layoutData.VAlign)
	c.False(layoutData.HGrab)
	c.NotNil(block.DrawCallback, "a banded block draws its rows")
	c.NotNil(block.DrawOverCallback, "a tinted block draws its tint")

	plain := &testPageBlock{}
	initTitledPagePanel(plain, "Plain", 2, false, nil)
	c.Nil(plain.DrawCallback, "a block that isn't banded draws nothing of its own")
	c.Nil(plain.DrawOverCallback, "a block without a tint draws no tint")
}

// A titled block's heading covers the painted title strip without moving the content, and follows the strip's height
// when the title's font changes.
func TestTitledPagePanelHeadingCoversTheTitle(t *testing.T) {
	c := check.New(t)
	block := &testPageBlock{}
	initTitledPagePanel(block, "Identity", 2, true, nil)
	block.AddChild(NewPageLabel("Name"))
	block.AddChild(NewPageLabel("Value"))
	children := block.Children()
	c.Equal(3, len(children), "the heading is a child of the block")
	heading := children[0]
	c.Equal(role.Heading, heading.Accessibility.Role)
	c.Equal(1, heading.Accessibility.Level)
	c.Equal("Identity", heading.Accessibility.Name)
	c.Equal("Identity", block.Accessibility.Name, "the block is still named for the focus arriving in it")
	c.Equal(heading, blockHeading(block.AsPanel()))
	c.Equal(1, blockRowsStart(block.AsPanel()), "the rows start after the heading")
	c.Equal(children[1:], blockRows(block.AsPanel()))

	layout, ok := block.Layout().(*blockLayout)
	if !ok {
		t.Fatalf("a titled block's layout must be a blockLayout, not a %T", block.Layout())
	}
	border := layout.border
	plain := &TitledBorder{Title: "Identity"}
	c.Equal(float32(1), border.Insets().Top, "the block's border leaves the title strip to the content")
	c.Equal(plain.TitleHeight()+1, plain.Insets().Top, "a border without a heading keeps the title strip")
	c.Equal(plain.Insets(), border.titleInsets(), "both paint the same area")

	_, pref, _ := block.Sizes(geom.Size{})
	block.SetFrameRect(geom.NewRect(0, 0, pref.Width, pref.Height))
	block.ValidateLayout()
	strip := geom.NewRect(1, 1, pref.Width-2, border.TitleHeight())
	c.Equal(strip, border.TitleStrip(pref), "the strip lies inside the lines at the top and sides of the border")
	c.Equal(strip, heading.FrameRect(), "the heading covers the painted title strip")
	insets := block.Border().Insets()
	c.True(heading.FrameRect().X < insets.Left && heading.FrameRect().Y < insets.Top,
		"the heading reaches outside the insets the content sits within")
	c.Equal(plain.Insets().Top+titledPagePanelInsets.Top, children[1].FrameRect().Y,
		"the content starts where it did when the border kept the title strip")
	c.Equal(children[1].FrameRect().Y, children[2].FrameRect().Y, "the labels share the first content row")

	before := border.TitleHeight()
	border.Font = border.font().Face().Font(border.font().Size() * 2)
	c.True(border.TitleHeight() > before, "a larger font makes the strip taller")
	block.MarkForLayoutRecursively()
	block.ValidateLayout()
	c.Equal(border.TitleStrip(block.FrameRect().Size), heading.FrameRect(), "the heading follows the strip's height")
	c.True(children[1].FrameRect().Y >= heading.FrameRect().Y+heading.FrameRect().Height,
		"the first row is laid out below the taller strip")

	// A block without a title has no heading.
	untitled := &testPageBlock{}
	initTitledPagePanel(untitled, "", 2, true, nil)
	c.Equal(0, len(untitled.Children()))
	c.Equal(plain.TitleHeight()+1+titledPagePanelInsets.Top, untitled.Border().Insets().Top,
		"the border keeps the title strip")
	_, ok = untitled.Layout().(*unison.FlexLayout)
	c.True(ok, "the layout is the plain FlexLayout, not a %T", untitled.Layout())
	c.Nil(blockHeading(untitled.AsPanel()))
	c.Equal(0, blockRowsStart(untitled.AsPanel()), "the banded rows start with the first child")
	untitled.AddChild(NewPageLabel("Name"))
	c.Equal(untitled.Children(), blockRows(untitled.AsPanel()))
}

// The attribute and basic damage blocks rebuild their rows, which must leave the heading first, or the first row would
// be laid out under the title.
func TestRebuildingBlocksKeepTheirHeading(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	entity.Recalculate()
	targetMgr := NewTargetMgr(unison.NewPanel())
	attrs := gurps.SheetSettingsFor(entity).Attributes
	primary := NewPrimaryAttrPanel(entity, targetMgr)
	secondary := NewSecondaryAttrPanel(entity, targetMgr)
	pools := NewPointPoolsPanel(entity, targetMgr)
	damage := NewDamagePanel(entity, targetMgr)
	for _, one := range []struct {
		name    string
		block   unison.Paneler
		rebuild func()
	}{
		{"primary attributes", primary, func() { primary.rebuild(attrs) }},
		{"secondary attributes", secondary, func() { secondary.rebuild(attrs) }},
		{"point pools", pools, func() { pools.rebuild(attrs) }},
		{"basic damage", damage, damage.rebuild},
	} {
		for _, pass := range []string{"built", "rebuilt"} {
			if pass == "rebuilt" {
				one.rebuild()
			}
			panel := one.block.AsPanel()
			children := panel.Children()
			if len(children) < 2 {
				t.Fatalf("%s %s: the block has %d children; expected the heading and its rows", one.name, pass,
					len(children))
			}
			heading := children[0]
			c.Equal(role.Heading, heading.Accessibility.Role, "%s %s: the heading comes first", one.name, pass)
			c.Equal(panel.Accessibility.Name, heading.Accessibility.Name, "%s %s: the heading carries the title", one.name, pass)
			// The sheet may give a block more width than its columns need; the heading must still span the block.
			_, pref, _ := panel.Sizes(geom.Size{})
			panel.SetFrameRect(geom.NewRect(0, 0, pref.Width*2, pref.Height))
			panel.ValidateLayout()
			content := panel.ContentRect(false)
			c.Equal(float32(1), heading.FrameRect().X, "%s %s: the heading starts inside the border's left line", one.name, pass)
			c.Equal(pref.Width*2-2, heading.FrameRect().Width, "%s %s: the heading spans the block", one.name, pass)
			bottom := heading.FrameRect().Y + heading.FrameRect().Height
			for _, child := range children[1:] {
				c.True(child.FrameRect().Y >= bottom, "%s %s: a %T at %v is laid out under the title strip, which ends at %v",
					one.name, pass, child.Self, child.FrameRect(), bottom)
				// A separator row would grab the spare width, but the factory attributes have none.
				c.True(child.FrameRect().Width < content.Width, "%s %s: the rows keep their own widths", one.name, pass)
			}
		}
	}
}

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
	for _, button := range uxtest.PanelsOfType[*unison.Button](root) {
		if drawable, ok := button.Drawable.(*unison.DrawableSVG); !ok || drawable.SVG != svg.Randomize {
			continue
		}
		wrapper := button.Parent()
		label, ok := uxtest.FirstPanelOfType[*unison.Label](wrapper)
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

// rowsAfterHeading is blockRows for a titled block, failing the test if the block does not start with a heading.
func rowsAfterHeading(t *testing.T, block unison.Paneler) []*unison.Panel {
	t.Helper()
	children := block.AsPanel().Children()
	if len(children) == 0 || children[0].Accessibility.Role != role.Heading {
		t.Fatalf("a titled block starts with its heading, but a %T has %d children and starts with %T", block,
			len(children), func() any {
				if len(children) == 0 {
					return nil
				}
				return children[0].Self
			}())
	}
	return children[1:]
}

// TestPageFieldRowsAlternateLabelsAndFields verifies that the Description, Identity and Miscellaneous blocks lay out
// their rows, after the heading, as a label followed by its field. The Description block's banding is drawn from that
// alternation (see TestDescriptionPanelBandsSkipTheHeading), so a helper that added the two in the wrong order would
// misplace every band below it.
func TestPageFieldRowsAlternateLabelsAndFields(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	targetMgr := NewTargetMgr(newPageUndoRoot())
	description := NewDescriptionPanel(entity, targetMgr)
	columns := rowsAfterHeading(t, description)
	c.Equal(3, len(columns), "the Description block has three columns")
	rows := map[string][]*unison.Panel{
		"identity":      rowsAfterHeading(t, NewIdentityPanel(entity, targetMgr)),
		"miscellaneous": rowsAfterHeading(t, NewMiscPanel(entity, targetMgr)),
	}
	for i, column := range columns {
		rows["description column "+string(rune('1'+i))] = column.Children()
	}
	for name, children := range rows {
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

// TestRandomizedPageFieldsAreNamedByTheirLabels verifies that an assistive technology names every randomized field on
// the Description and Identity blocks after the label in the randomizer wrapper before it, and reports that label as
// what names the field. The sibling-label convention cannot do this on its own, since the field's preceding sibling is
// the wrapper rather than the label. It also verifies that each randomizer button is still described, rather than being
// folded away into a label.
func TestRandomizedPageFieldsAreNamedByTheirLabels(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	var wnd *unison.Window
	var rows []randomizedRow
	screen.Do(func() {
		entity := gurps.NewEntity()
		root := newPageUndoRoot()
		root.SetLayout(&unison.FlexLayout{Columns: 1})
		targetMgr := NewTargetMgr(root)
		root.AddChild(NewDescriptionPanel(entity, targetMgr))
		root.AddChild(NewIdentityPanel(entity, targetMgr))
		rows = randomizedRows(t, root.AsPanel())
		w, err := unison.NewWindow("Randomized Fields")
		if err != nil {
			t.Errorf("unable to create the window: %v", err)
			return
		}
		w.Content().SetLayout(&unison.FlexLayout{Columns: 1})
		w.Content().AddChild(root)
		w.Pack()
		w.ToFront()
		wnd = w
	})
	if wnd == nil {
		t.Fatal("the window was not created")
	}
	c.Equal(randomizedFieldCount, len(rows), "every randomizable field has a randomizer button")
	c.NotNil(screen.AccessibilityTree(wnd), "the window is described")
	for _, row := range rows {
		fieldNode := screen.AccessibilityNodeFor(row.field)
		if fieldNode == nil {
			t.Fatalf("the %q field is not described", row.title)
		}
		c.Equal(row.title, fieldNode.Name, "the %q field is named after its label", row.title)
		var label *unison.Label
		screen.Do(func() { label, _ = uxtest.FirstPanelOfType[*unison.Label](row.button.Parent()) })
		labelNode := screen.AccessibilityNodeFor(label)
		if labelNode == nil {
			t.Fatalf("the %q label is not described", row.title)
		}
		c.Equal([]accessibility.NodeID{labelNode.ID}, fieldNode.LabeledBy, "the %q field is labeled by its label",
			row.title)
		c.NotNil(screen.AccessibilityNodeFor(row.button), "the %q randomizer button is described", row.title)
	}
	screen.Do(func() { wnd.Dispose() })
}
