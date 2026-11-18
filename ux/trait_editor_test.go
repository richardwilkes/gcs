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
	"maps"
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantadjust"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// TestTraitEditorShowsTheRangeOfAnOpenChoice verifies that a trait editor opened outside a sheet counts a mandatory
// modifier choice yet to be made in its Point Cost as the list does, as a range once its options cost different
// amounts, and as a single cost once the trait is marked preconfigured with its pick made, the editor's own pending
// state counting throughout.
func TestTraitEditorShowsTheRangeOfAnOpenChoice(t *testing.T) {
	c := check.New(t)
	trait := gurps.NewTrait(nil, nil, false)
	trait.BasePoints = fxp.FromInteger(10)
	choice := newTraitModifierChoiceFor(nil, true, []string{"Small", "Large"})
	choice.Children[0].CostAdj = "+5"
	choice.Children[1].CostAdj = "+5"
	trait.AddModifiers(choice)
	c.Equal(fxp.FromInteger(15), trait.AdjustedPoints(nil), "the list counts the choice")
	e, content := buildEditorContent(nil, trait, initTraitEditor)
	// The Point Cost field is the first of the editor's non-editable fields.
	pointCost := uxtest.PanelsOfType[*NonEditableField](content)[0]
	c.Equal("15", pointCost.String())
	options := e.editorData.Modifiers[0].Children
	options[1].CostAdj = "+10"
	DeepSync(e)
	c.Equal("15~20", pointCost.String())
	e.editorData.Preconfigured = true
	DeepSync(e)
	c.Equal("15~20", pointCost.String(), "a preconfigured trait still has a choice with no pick to make")
	options[1].SetDisabled(false)
	DeepSync(e)
	c.Equal("20", pointCost.String(), "a preconfigured trait takes the pick already made")
}

// TestTraitEditorLeavesOutInheritedModifiers verifies that the Point Cost of a trait with a modifier choice leaves out
// the modifiers of its container whether the choice is open or made.
func TestTraitEditorLeavesOutInheritedModifiers(t *testing.T) {
	c := check.New(t)
	parent := gurps.NewTrait(nil, nil, true)
	half := gurps.NewTraitModifier(nil, nil, false)
	half.CostAdj = "-50%"
	parent.AddModifiers(half)
	trait := gurps.NewTrait(nil, parent, false)
	trait.BasePoints = fxp.FromInteger(20)
	parent.Children = []*gurps.Trait{trait}
	choice := newTraitModifierChoiceFor(nil, true, []string{"+0%", "+100%"})
	for _, option := range choice.Children {
		option.CostAdj = option.Name
	}
	trait.AddModifiers(choice)
	e, content := buildEditorContent(nil, trait, initTraitEditor)
	pointCost := uxtest.PanelsOfType[*NonEditableField](content)[0]
	c.Equal("20~40", pointCost.String())
	e.editorData.Preconfigured = true
	e.editorData.Modifiers[0].Children[1].SetDisabled(false)
	DeepSync(e)
	c.Equal("40", pointCost.String())
	e.editorData.Modifiers[0].Children[1].SetDisabled(true)
	e.editorData.Modifiers[0].Children[0].SetDisabled(false)
	DeepSync(e)
	c.Equal("20", pointCost.String())
}

// TestTraitEditorLevelFieldsFollowTheLevelsCheckBox verifies that the trait editor's level, cost per level and maximum
// level fields can only be used while the trait has levels.
func TestTraitEditorLevelFieldsFollowTheLevelsCheckBox(t *testing.T) {
	c := check.New(t)
	_, content := buildEditorContent(nil, gurps.NewTrait(nil, nil, false), initTraitEditor)
	box := findCheckBoxTitled(content, "Levels")
	if box == nil {
		t.Fatal("the trait editor has no Levels check box")
	}
	fields := make(map[string]*unison.Panel)
	for _, field := range uxtest.PanelsOfType[*DecimalField](content) {
		if field.undoTitle == "Level" || field.undoTitle == "Cost Per Level" {
			fields[field.undoTitle] = field.AsPanel()
		}
	}
	for _, field := range uxtest.PanelsOfType[*StringField](content) {
		if field.undoTitle == "Maximum Level" {
			fields[field.undoTitle] = field.AsPanel()
		}
	}
	c.Equal(3, len(fields), "precondition: the editor has the three level fields")
	for title, field := range fields {
		c.False(field.Enabled(), "%s starts out unusable", title)
	}
	for _, levels := range []bool{true, false} {
		clickCheckBox(box, levels)
		for title, field := range fields {
			c.Equal(levels, field.Enabled(), "%s with levels %v", title, levels)
		}
	}
}

// TestTraitEditorShowsTheAdjustedMaximumLevel verifies that the trait editor's Adjusted Maximum Level shows nothing
// while there is no maximum, and follows the maximum as it is typed.
func TestTraitEditorShowsTheAdjustedMaximumLevel(t *testing.T) {
	c := check.New(t)
	trait := gurps.NewTrait(nil, nil, false)
	trait.CanLevel = true
	e, content := buildEditorContent(nil, trait, initTraitEditor)
	var maximum *StringField
	for _, field := range uxtest.PanelsOfType[*StringField](content) {
		if field.undoTitle == "Maximum Level" {
			maximum = field
		}
	}
	if maximum == nil {
		t.Fatal("the trait editor has no Maximum Level field")
	}
	// Point Cost is the first of the editor's non-editable fields, and Adjusted Maximum Level the second.
	shown := uxtest.PanelsOfType[*NonEditableField](content)
	if len(shown) < 2 {
		t.Fatal("the trait editor has no Adjusted Maximum Level field")
	}
	adjusted := shown[1]
	c.Equal("", adjusted.String(), "no maximum")
	maximum.SetText("3")
	c.Equal("3", e.editorData.MaxLevels, "precondition: the field edits the maximum")
	c.Equal("3", adjusted.String(), "the typed maximum")
}

// Both container and non-container traits offer the "Switched On" checkbox, since a container's modifiers can hold
// switchable features even though the container itself has none of its own.
func TestTraitEditorHasSwitchedOnCheckBox(t *testing.T) {
	for _, isContainer := range []bool{false, true} {
		name := "trait"
		if isContainer {
			name = "trait container"
		}
		t.Run(name, func(t *testing.T) {
			c := check.New(t)
			sheet := newTestSheetForTemplate(t)
			trait := gurps.NewTrait(sheet.Entity(), nil, isContainer)
			e, content := buildEditorContent(sheet, trait, initTraitEditor)
			box := findCheckBoxTitled(content, i18n.Text("Switched On"))
			c.NotNil(box, "expected a Switched On checkbox in the editor")
			c.False(e.editorData.SwitchedOn, "the switch starts out off")

			clickCheckBox(box, true)
			c.True(e.editorData.SwitchedOn, "checking the box must turn the switch on in the editor's data")
			c.False(trait.SwitchedOn, "the target must not be touched until the edit is applied")

			clickCheckBox(box, false)
			c.False(e.editorData.SwitchedOn, "clearing the box must turn the switch back off")
		})
	}
}

// TestTraitEditorOffersUserDescOnlyWhereItIsKept verifies that the trait editor offers the user description field for
// a trait on a character sheet and for the copy of a trait a grant holds, which goes to the sheet with it, and leaves
// it out for a trait anywhere else, which has none.
func TestTraitEditorOffersUserDescOnlyWhereItIsKept(t *testing.T) {
	c := check.New(t)
	offers := func(owner Rebuildable, trait *gurps.Trait) bool {
		_, content := buildEditorContent(owner, trait, initTraitEditor)
		return slices.Contains(labelTitles(content), "User Description")
	}
	sheet := newTestSheetForTemplate(t)
	c.True(offers(sheet, gurps.NewTrait(sheet.Entity(), nil, false)), "a trait on a sheet has one")
	c.False(offers(nil, gurps.NewTrait(nil, nil, false)), "a trait in a library has none")
	c.False(offers(nil, gurps.NewTrait(gurps.NewTemplate(), nil, false)), "nor does one in a template")
	g := gurps.NewGrant(feature.GrantTrait)
	g.SetItem(newGrantTestTrait("Night Vision", fxp.One), gurps.LibraryFile{})
	features := gurps.Features{g}
	owner := &grantItemOwner{
		Rebuildable: sheet,
		panel:       &featuresPanel{features: &features},
		key:         g.Key(),
		item:        g.Trait,
	}
	c.True(offers(owner, g.Trait.Clone(gurps.LibraryFile{}, nil, nil, gurps.Copy)),
		"the copy of a trait a grant holds has one")
}

// findIntegerFieldLabeled returns the integer field that follows the label bearing the given text among the panel's
// direct children.
func findIntegerFieldLabeled(p *unison.Panel, labelText string) *IntegerField {
	children := p.Children()
	for i, child := range children {
		if label, ok := child.Self.(*unison.Label); ok && label.String() == labelText && i+1 < len(children) {
			if field, ok2 := children[i+1].Self.(*IntegerField); ok2 {
				return field
			}
		}
	}
	return nil
}

// A new trait container holds zero alternative slots, since only an alternative abilities container uses them, so the
// slots field's minimum must follow the container type. A fixed minimum of one showed the invalid zero's error color
// through the blanked field's border insets, and its clamping bumped the editor's copy to one, so the editor opened
// already modified.
func TestTraitEditorAlternativeSlotsFollowContainerType(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	trait := gurps.NewTrait(sheet.Entity(), nil, true)
	e, content := buildEditorContent(sheet, trait, initTraitEditor)
	field := findIntegerFieldLabeled(content, i18n.Text("Alternative Slots"))
	c.NotNil(field, "expected an Alternative Slots field in the trait container editor")
	c.Equal(0, e.editorData.AlternativeSlots, "a new container holds no slots")
	c.False(field.Invalid(), "zero slots is valid for a container that is not an alternative abilities one")
	c.False(field.Enabled(), "the field is blanked for a container that is not an alternative abilities one")
	c.Equal("", e.editorData.Ancestry, "a container that is not an ancestry one holds no ancestry")
	c.False(e.isModified(), "opening the editor must not modify the data")

	e.editorData.ContainerType = container.AlternativeAbilities
	e.MarkModified(nil)
	c.Equal(1, e.editorData.AlternativeSlots, "an alternative abilities container needs at least one slot")
	c.Equal("1", field.Text(), "the field must show the slot it was given")
	c.False(field.Invalid(), "one slot is valid")
	c.True(field.Enabled(), "the field is usable for an alternative abilities container")
	c.Equal(1, field.Min(), "an alternative abilities container may not have fewer than one slot")

	field.SetText("0")
	c.True(field.Invalid(), "zero slots is invalid for an alternative abilities container")
	field.SetText("3")
	c.False(field.Invalid(), "three slots is valid")
	c.Equal(3, e.editorData.AlternativeSlots, "the slot count must reach the editor's data")

	e.editorData.ContainerType = container.Group
	e.MarkModified(nil)
	c.False(field.Invalid(), "the slots are ignored for a group, so nothing in the field is invalid")
	c.False(field.Enabled(), "the field is blanked again for a group")
	c.Equal(0, field.Min(), "a group holds zero slots")
}

// Switching a container to an ancestry one gives it the default ancestry, since the empty ancestry it held until then
// is not something the popup can show.
func TestTraitEditorAncestryFollowsContainerType(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	trait := gurps.NewTrait(sheet.Entity(), nil, true)
	e, _ := buildEditorContent(sheet, trait, initTraitEditor)
	c.Equal("", e.editorData.Ancestry, "a group holds no ancestry")

	e.editorData.ContainerType = container.Ancestry
	e.MarkModified(nil)
	c.Equal(gurps.DefaultAncestry, e.editorData.Ancestry, "an ancestry container starts with the default ancestry")
}

// TestGrantItemEditorLeavesLevelAndEnabledToTheGrant checks that the editor of a grant's trait offers neither a Level
// field nor an Enabled check box, since the grant's row sets the level of the trait it puts on the sheet and that trait
// is enabled whenever the grant is in effect, so nothing would read either.
func TestGrantItemEditorLeavesLevelAndEnabledToTheGrant(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	training := gurps.NewTrait(entity, nil, false)
	training.Name = "Training"
	g := gurps.NewGrant(feature.GrantTrait)
	g.SetItem(newGrantTestTrait("Acute Vision", fxp.Two), gurps.LibraryFile{})
	training.Features = gurps.Features{g}
	entity.Traits = []*gurps.Trait{training}
	sheet.Rebuild(true)
	isTraitEditor := func(d unison.Dockable) bool {
		_, ok := d.AsPanel().Self.(*editor[*gurps.Trait, *gurps.TraitEditData])
		return ok
	}
	var outer *editor[*gurps.Trait, *gurps.TraitEditData]
	screen.Do(func() {
		DisplayNewDockable(sheet)
		outer = EditTrait(sheet, training)
		panels := uxtest.PanelsOfType[*featuresPanel](outer.content)
		c.Equal(1, len(panels))
		panels[0].editGrantedItem(grantAt(t, panels[0], 0))
	})
	nested := uxtest.OtherEditor(t, screen, outer, isTraitEditor)
	hasLevelField := func(content *unison.Panel) bool {
		return slices.ContainsFunc(uxtest.PanelsOfType[*DecimalField](content), func(f *DecimalField) bool {
			return f.undoTitle == "Level"
		})
	}
	hasEnabledCheckBox := func(content *unison.Panel) bool {
		return slices.ContainsFunc(uxtest.PanelsOfType[*CheckBox](content), func(cb *CheckBox) bool {
			return cb.Text.String() == "Enabled"
		})
	}
	screen.Do(func() {
		_, isNested := nested.owner.(*grantItemOwner)
		c.True(isNested, "precondition: the other trait editor is the grant item's")
		c.True(hasLevelField(outer.content), "precondition: a trait's own editor has a Level field")
		c.True(hasEnabledCheckBox(outer.content), "and an Enabled check box")
		c.False(hasLevelField(nested.content), "the grant's row sets the level")
		c.False(hasEnabledCheckBox(nested.content), "and the granted trait is enabled while the grant is in effect")
	})
}

// TestGrantItemEditorShowsTheCostAtTheGrantsLevel checks that the editor of a grant's trait, which leaves the level to
// the grant, shows the point cost at the grant's level rather than at the level the grant's copy of the trait holds,
// which the grant's row never changes, and a cost of 0 for a level below 1, at which the grant puts no trait on the
// sheet.
func TestGrantItemEditorShowsTheCostAtTheGrantsLevel(t *testing.T) {
	for _, tc := range []struct {
		name   string
		kind   grantadjust.Kind
		amount fxp.Int
		want   string
	}{
		{name: "level set", kind: grantadjust.SetLevel, amount: fxp.Three, want: "8"},
		{name: "level set to 1", kind: grantadjust.SetLevel, amount: fxp.One, want: "6"},
		{name: "levels taken away", kind: grantadjust.AddLevels, amount: -fxp.Two, want: "0"},
		{name: "level set below 1", kind: grantadjust.SetLevel, amount: fxp.Half, want: "0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			entity := gurps.NewEntity()
			carrier := gurps.NewTrait(entity, nil, false)
			carrier.Name = "Training"
			trait := newGrantTestTrait("Acute Vision", fxp.One)
			// A base cost, so that the cost at level 0 isn't 0.
			trait.BasePoints = fxp.Five
			g := gurps.NewGrant(feature.GrantTrait)
			g.SetItem(trait, gurps.LibraryFile{})
			carrier.Features = gurps.Features{g}
			g.Adjust = tc.kind
			g.Amount = tc.amount
			root := &stubRebuildable{}
			root.Self = root
			item := g.Trait.Clone(gurps.LibraryFile{}, nil, nil, gurps.Copy)
			owner := &grantItemOwner{
				Rebuildable: root,
				panel:       newFeaturesPanel(entity, carrier, &carrier.Features, false),
				key:         g.Key(),
				hash:        grantItemHash(g),
				item:        item,
			}
			_, content := buildEditorContent(owner, item, initTraitEditor)
			fields := uxtest.PanelsOfType[*NonEditableField](content)
			if len(fields) == 0 {
				t.Fatal("the editor must show the point cost")
			}
			c.Equal(tc.want, fields[0].String())
		})
	}
}

// The trait editor's section edits the editor's copy, so the trait only gets its nameables and the links for its
// modifiers' codes when the edit is applied.
func TestTraitEditorAppliesNameablesAndLinks(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	trait := gurps.NewTrait(sheet.Entity(), nil, false)
	trait.Name = "Trained in @style@"
	mod := gurps.NewTraitModifier(sheet.Entity(), nil, false)
	mod.Name = "@art@ focus"
	trait.SetModifiers([]*gurps.TraitModifier{mod})
	e, content := buildEditorContent(sheet, trait, initTraitEditor)
	p, ok := uxtest.FirstPanelOfType[*nameablesPanel](content)
	c.True(ok, "expected a nameables panel in the trait editor")
	if !ok {
		return
	}
	style := map[string]nameable.Detail{"style": {Label: "Style", FreeForm: true}}
	*p.defs = maps.Clone(style)
	p.setLink(mod.TID, "art", "style")
	c.Nil(trait.Nameables, "the trait is untouched until the edit is applied")
	c.Nil(trait.NameableLinks)

	e.editorData.ApplyTo(trait)
	c.Equal(style, trait.Nameables)
	c.Equal(map[tid.TID]map[string]string{mod.TID: {"art": "style"}}, trait.NameableLinks)
	trait.Replacements = map[string]string{"style": "Judo"}
	c.Equal("Judo focus", trait.Modifiers[0].NameWithReplacements(), "the applied link resolves")
}
