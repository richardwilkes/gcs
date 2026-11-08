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
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
)

// TestAttrPanelHashReflectsTraitDrivenVisibility verifies that the panel's hash changes when a trait is added, removed,
// enabled, or disabled in a way that reveals or hides an attribute. Without folding the trait-driven placement into the
// hash, Sync would not rebuild the panel and the revealed/hidden attribute would not update on the sheet.
func TestAttrPanelHashReflectsTraitDrivenVisibility(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()

	// A hidden secondary attribute that becomes visible when the character has the "Magery" trait.
	defs := &gurps.AttributeDefs{Set: map[string]*gurps.AttributeDef{
		"mana": {AttributeDefData: gurps.AttributeDefData{
			DefID:                "mana",
			Type:                 attribute.Integer,
			Base:                 "10",
			Placement:            attribute.Hidden,
			PlacementTrait:       "Magery",
			PlacementWhenPresent: attribute.Secondary,
		}},
	}}

	panel := &AttrPanel{entity: e, kind: gurps.SecondaryAttrKind}
	hidden := panel.computeHash(defs)

	trait := gurps.NewTrait(e, nil, false)
	trait.Name = "Magery"
	e.Traits = append(e.Traits, trait)
	revealed := panel.computeHash(defs)
	c.NotEqual(hidden, revealed, "adding the trait must change the hash")

	trait.SetDisabled(true)
	c.Equal(hidden, panel.computeHash(defs), "disabling the trait must restore the hidden hash")

	trait.SetDisabled(false)
	c.Equal(revealed, panel.computeHash(defs), "re-enabling the trait must restore the revealed hash")

	e.Traits = nil
	c.Equal(hidden, panel.computeHash(defs), "removing the trait must restore the hidden hash")
}

// TestAttrPanelValueFieldTooltipMatchesNameLabel verifies that the tooltip on an attribute's name label lists the
// sources of its bonuses, after the full name for a point pool, and that the value field of the row carries the same
// tooltip, in all three kinds of panel. Sync must refresh both in place without rebuilding the panel, since a change to
// a trait or piece of equipment alters the bonuses without changing the attribute definitions. Users hover over the
// number at least as often as the name, so the sources of the bonuses must be reachable from there too.
func TestAttrPanelValueFieldTooltipMatchesNameLabel(t *testing.T) {
	for _, one := range []struct {
		name       string
		newPanel   func(*gurps.Entity, *TargetMgr) *AttrPanel
		attrID     string
		traitName  string
		withBonus  string
		withoutTip string
	}{
		{
			name:      "primary",
			newPanel:  NewPrimaryAttrPanel,
			attrID:    gurps.StrengthID,
			traitName: "Increased Strength",
			withBonus: "Includes modifiers from:\nIncreased Strength [+2]",
		},
		{
			name:      "secondary",
			newPanel:  NewSecondaryAttrPanel,
			attrID:    "will",
			traitName: "Strong Will",
			withBonus: "Includes modifiers from:\nStrong Will [+2]",
		},
		{
			name:       "pool",
			newPanel:   NewPointPoolsPanel,
			attrID:     "fp",
			traitName:  "Fit",
			withBonus:  "Fatigue Points\nIncludes modifiers from:\nFit [+2]",
			withoutTip: "Fatigue Points",
		},
	} {
		t.Run(one.name, func(t *testing.T) {
			c := check.New(t)
			e := gurps.NewEntity()
			e.Recalculate()
			panel := one.newPanel(e, NewTargetMgr(unison.NewPanel()))
			label := panel.nameLabels[one.attrID]
			c.NotNil(label, "the name label must be tracked")
			field := panel.valueFields[one.attrID]
			c.NotNil(field, "the value field must be tracked")
			if one.withoutTip == "" {
				c.Nil(label.Tooltip, "no bonuses means no tooltip on the name label")
				c.Nil(field.AsPanel().Tooltip, "no bonuses means no tooltip on the value field")
			} else {
				c.Equal(one.withoutTip, tooltipText(label.Tooltip), "without bonuses, only the full name is shown")
				c.Equal(one.withoutTip, tooltipText(field.AsPanel().Tooltip),
					"without bonuses, the value field shows what the name label shows")
			}

			trait := gurps.NewTrait(e, nil, false)
			trait.Name = one.traitName
			bonus := gurps.NewAttributeBonus(one.attrID)
			bonus.Amount = fxp.Two
			trait.Features = gurps.Features{bonus}
			e.Traits = append(e.Traits, trait)
			e.Recalculate()
			hash := panel.hash
			panel.Sync()
			c.Equal(hash, panel.hash, "a change to the bonuses alone must not rebuild the panel")
			c.True(label == panel.nameLabels[one.attrID], "the existing name label must be refreshed in place")
			c.True(field == panel.valueFields[one.attrID], "the existing value field must be refreshed in place")
			c.Equal(one.withBonus, tooltipText(field.AsPanel().Tooltip),
				"the value field lists the bonus and its source")
			c.Equal(tooltipText(label.Tooltip), tooltipText(field.AsPanel().Tooltip),
				"the value field and the name label show the same tooltip")

			e.Traits = nil
			e.Recalculate()
			panel.Sync()
			if one.withoutTip == "" {
				c.Nil(label.Tooltip, "removing the bonus removes the name label's tooltip")
				c.Nil(field.AsPanel().Tooltip, "removing the bonus removes the value field's tooltip")
			} else {
				c.Equal(one.withoutTip, tooltipText(label.Tooltip),
					"removing the bonus leaves just the full name behind")
				c.Equal(one.withoutTip, tooltipText(field.AsPanel().Tooltip),
					"removing the bonus leaves just the full name behind")
			}
		})
	}
}

// TestAttrPanelValueFieldKeepsValidationTooltip verifies that a Sync landing while the user has typed something invalid
// into an attribute's value field leaves the message explaining the problem in place, and that the bonus tooltip
// installed by that Sync takes over once the content is valid again. Sync runs on nearly every change to the sheet, so
// without this the explanation of why the typing was rejected would disappear while the user still needed it.
func TestAttrPanelValueFieldKeepsValidationTooltip(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	e.Recalculate()
	panel := NewPrimaryAttrPanel(e, NewTargetMgr(unison.NewPanel()))
	field, ok := panel.valueFields[gurps.StrengthID].(*IntegerField)
	c.True(ok, "the ST value field must be an integer field")

	field.SetText("not a number")
	c.Equal("Invalid value", tooltipText(field.Tooltip), "an invalid value explains itself via the tooltip")

	trait := gurps.NewTrait(e, nil, false)
	trait.Name = "Increased Strength"
	bonus := gurps.NewAttributeBonus(gurps.StrengthID)
	bonus.Amount = fxp.Two
	trait.Features = gurps.Features{bonus}
	e.Traits = append(e.Traits, trait)
	e.Recalculate()
	panel.Sync()
	c.Equal("Invalid value", tooltipText(field.Tooltip),
		"the tooltip installed by Sync must not displace the validation message")

	field.SetText("10")
	c.Equal("Includes modifiers from:\nIncreased Strength [+2]", tooltipText(field.Tooltip),
		"the bonus tooltip takes over once the value is valid again")
}

// tooltipText returns the text of a tooltip built by newWrappedTooltip, one line per label. A nil tooltip yields a
// sentinel rather than panicking, so that a missing tooltip fails the comparison readably.
func tooltipText(tip *unison.Panel) string {
	if tip == nil {
		return "<nil>"
	}
	var lines []string
	for _, child := range tip.Children() {
		if label, ok := child.Self.(*unison.Label); ok {
			lines = append(lines, label.String())
		}
	}
	return strings.Join(lines, "\n")
}

// markdownText returns the text of the labels a rendered Markdown panel is drawn with, each ended by a line feed.
func markdownText(md *unison.Markdown) string {
	var b strings.Builder
	md.HasInSelfOrDescendants(func(p *unison.Panel) bool {
		if label, ok := p.Self.(*unison.Label); ok {
			b.WriteString(label.String())
			b.WriteByte('\n')
		}
		return false
	})
	return b.String()
}

// The drawn text is withheld from the node, since macOS would speak both the name and the text.
func TestAttrPanelPointsFieldsAreSpokenAsPoints(t *testing.T) {
	c := check.New(t)
	c.Equal("1 point", spokenPoints(fxp.One))
	c.Equal("-1 point", spokenPoints(fxp.NegOne))
	c.Equal("0 points", spokenPoints(0))
	c.Equal("40 points", spokenPoints(fxp.FromInteger(40)))
	c.Equal("2.5 points", spokenPoints(fxp.FromStringForced("2.5")))

	e := gurps.NewEntity()
	e.Recalculate()
	panel := NewPrimaryAttrPanel(e, NewTargetMgr(unison.NewPanel()))
	st := e.Attributes.Set[gurps.StrengthID]
	c.NotNil(st)
	children := panel.Children()
	i := slices.Index(children, panel.valueFields[gurps.StrengthID].AsPanel())
	if i < 2 {
		t.Fatalf("the points field must be laid out just before the value field, after the heading, but the value "+
			"field is child %d", i)
	}
	points, ok := children[i-1].Self.(*NonEditablePageField)
	if !ok {
		t.Fatalf("the panel before the value field must be the points field, not a %T", children[i-1].Self)
	}
	c.Equal("[0]", points.Text.String())
	c.Equal("0 points", points.Accessibility.Name)

	st.SetMaximum(st.Maximum() + fxp.Two)
	e.Recalculate()
	points.Sync()
	c.Equal("[20]", points.Text.String())
	c.Equal("20 points", points.Accessibility.Name, "the spoken points must follow the attribute")

	node := &accessibility.Node{Name: points.Accessibility.Name, Text: &accessibility.TextInfo{Text: "[20]"}}
	points.Accessibility.Callback(node)
	c.Nil(node.Text, "the drawn text must be withheld from the node")
	c.Equal("20 points", node.Name)
}

// tiredEntity returns a new entity whose FP are damaged to 7 of 10, which puts them in the Tired threshold, with the
// given features added to those the factory definitions give that threshold.
func tiredEntity(t *testing.T, features ...gurps.Feature) *gurps.Entity {
	t.Helper()
	e := gurps.NewEntity()
	var tired *gurps.PoolThreshold
	for _, one := range e.SheetSettings.Attributes.Set[gurps.FatiguePointsID].Thresholds {
		if one.State == "Tired" {
			tired = one
		}
	}
	if tired == nil {
		t.Fatal("the factory FP definition must have a Tired threshold")
	}
	tired.Features = append(tired.Features, features...)
	e.Attributes.Set[gurps.FatiguePointsID].Damage = fxp.Seven
	e.Recalculate()
	return e
}

// TestAttrPanelShowsAdjustedCurrentValue verifies that the value field of an attribute held away from its maximum by a
// pool threshold shows the current value in adjustedInk, names the maximum in its tooltip and spoken name, and puts
// the maximum in place for editing, with the current value spoken in its name instead, also for an assistive
// technology replacing the value without focusing the field, whose edit undo then reverts to the maximum, and that all
// of that goes once the threshold no longer applies.
func TestAttrPanelShowsAdjustedCurrentValue(t *testing.T) {
	c := check.New(t)
	halveST := gurps.NewAttributeBonus(gurps.StrengthID)
	halveST.Multiply = true
	halveST.Amount = fxp.Half
	e := tiredEntity(t, halveST)
	panel := NewPrimaryAttrPanel(e, NewTargetMgr(unison.NewPanel()))
	root := newPageUndoRoot()
	root.AddChild(panel)
	st := e.Attributes.Set[gurps.StrengthID]
	field, ok := panel.valueFields[gurps.StrengthID].(*IntegerField)
	c.True(ok, "the ST value field must be an integer field")
	dx, ok := panel.valueFields[gurps.DexterityID].(*IntegerField)
	c.True(ok, "the DX value field must be an integer field")
	normalInk := dx.OnEditableInk
	c.NotEqual(adjustedInk, normalInk, "an attribute left alone is drawn in the normal ink")
	spokenName := func() string {
		node := &accessibility.Node{Name: "ST"}
		field.Accessibility.Callback(node)
		return node.Name
	}

	c.Equal("5", field.Text(), "the current value shows while the field is not being edited")
	c.Equal(adjustedInk, field.OnEditableInk, "the adjusted value is drawn in adjustedInk")
	tip := tooltipText(field.Tooltip)
	c.Contains(tip, "Adjusted from 10", "the tooltip names the maximum")
	c.Contains(tip, "Tired (FP) [×0.5]", "the tooltip names the threshold")
	c.Equal("ST, adjusted from 10", spokenName())

	field.gainedFocus()
	c.Equal("10", field.Text(), "the maximum is what is edited")
	c.Equal("ST, adjusted to 5", spokenName(), "while the maximum is edited, the current value is spoken")
	field.lostFocus()
	c.Equal("5", field.Text(), "the current value returns once the focus is lost")
	c.Equal("ST, adjusted from 10", spokenName(), "and so does the spoken maximum")

	for _, req := range []accessibility.ActionRequest{
		{Action: accessibility.SetValue, Value: "12"},
		{Action: accessibility.ReplaceText, Start: 0, End: 1, Value: "12"},
	} {
		c.True(field.PerformAccessibilityAction(req), "%v: an assistive technology may replace the value", req.Action)
		c.Equal(fxp.Twelve, st.Maximum(), "%v: an assistive technology's value replaces the maximum", req.Action)
		root.mgr.Undo()
		c.Equal(fxp.Ten, st.Maximum(), "%v: undo restores the maximum rather than the current value shown",
			req.Action)
		// Undo focuses the field, which can't take the focus without a window.
		field.gainedFocus()
		c.Equal("10", field.Text(), "%v: undo puts the maximum in place for editing", req.Action)
		field.lostFocus()
		c.Equal("5", field.Text(), "%v: the current value returns once the focus is lost", req.Action)
	}
	c.False(field.PerformAccessibilityAction(accessibility.ActionRequest{
		Action: accessibility.ReplaceText, Start: 1, End: 1, Value: "0",
	}), "a range of the current value shown means nothing in the maximum, so replacing it is refused")
	c.Equal(fxp.Ten, st.Maximum(), "a refused replacement leaves the maximum alone")
	c.Equal("5", field.Text(), "a refused replacement leaves the current value showing")

	e.Attributes.Set[gurps.FatiguePointsID].Damage = 0
	e.Recalculate()
	DeepSync(panel)
	c.Equal("10", field.Text(), "rested, the current value is the maximum")
	c.Equal(normalInk, field.OnEditableInk, "rested, the value is drawn in the normal ink")
	c.NotContains(tooltipText(field.Tooltip), "Tired", "rested, the tooltip no longer names the threshold")
	c.Equal("ST", spokenName(), "rested, the spoken name is just the attribute's")
	c.True(field.PerformAccessibilityAction(accessibility.ActionRequest{
		Action: accessibility.ReplaceText, Start: 1, End: 2, Value: "1",
	}), "rested, the value shown is the maximum, so a range of it may be replaced")
	c.Equal(fxp.Eleven, st.Maximum(), "rested, replacing a range of the value shown edits the maximum")
}

// TestAttrPanelStandInLimitedBonusesAreNotAdjustments verifies that the value field of an attribute standing in for
// Lifting ST, whose current value takes the lifting-only ST bonuses as part of the value, as the lifting ST the sheet
// works out without it does, is drawn in the normal ink and neither named nor spoken as adjusted while no threshold or
// multiplier alters it, even when it allows decimals and its current value drops the fraction its maximum has, and
// that it is once the factory Tired threshold halves lifting ST.
func TestAttrPanelStandInLimitedBonusesAreNotAdjustments(t *testing.T) {
	for _, one := range []struct {
		name    string
		typ     attribute.Type
		base    string
		maximum string
	}{
		{name: "integer", typ: attribute.Integer, base: "$st", maximum: "10"},
		{name: "decimal", typ: attribute.Decimal, base: "$st + 0.5", maximum: "10.5"},
	} {
		t.Run(one.name, func(t *testing.T) {
			c := check.New(t)
			e := gurps.NewEntity()
			e.SheetSettings.Attributes.Set[gurps.LiftingStrengthID] = &gurps.AttributeDef{
				DefID: gurps.LiftingStrengthID,
				Type:  one.typ,
				Name:  "Lifting ST",
				Base:  one.base,
			}
			e.Attributes.Set[gurps.LiftingStrengthID] = gurps.NewAttribute(e, gurps.LiftingStrengthID,
				len(e.Attributes.Set))
			trait := gurps.NewTrait(e, nil, false)
			trait.Name = "Lifting Harness"
			bonus := gurps.NewAttributeBonus(gurps.StrengthID)
			bonus.Amount = fxp.Two
			bonus.Limitation = stlimit.LiftingOnly
			trait.Features = gurps.Features{bonus}
			e.Traits = append(e.Traits, trait)
			e.Recalculate()
			panel := NewSecondaryAttrPanel(e, NewTargetMgr(unison.NewPanel()))
			var field *unison.Field
			switch f := panel.valueFields[gurps.LiftingStrengthID].(type) {
			case *IntegerField:
				field = f.Field
			case *DecimalField:
				field = f.Field
			default:
				t.Fatalf("the lifting ST value field must be a numeric field, not a %T", f)
			}
			spokenName := func() string {
				node := &accessibility.Node{Name: "Lifting ST"}
				field.Accessibility.Callback(node)
				return node.Name
			}
			normalInk := field.OnEditableInk

			c.Equal("12", field.Text(), "the current value takes the lifting-only ST bonus")
			tip := tooltipText(field.Tooltip)
			c.Contains(tip, "Lifting Harness [+2]", "the tooltip names the lifting-only ST bonus")
			c.NotContains(tip, "Adjusted from", "a lifting-only ST bonus is not an adjustment")
			c.Equal("Lifting ST", spokenName(), "a lifting-only ST bonus is not spoken as an adjustment")
			c.NotEqual(adjustedInk, field.OnEditableInk, "a lifting-only ST bonus is not drawn in adjustedInk")

			e.Attributes.Set[gurps.FatiguePointsID].Damage = fxp.Seven
			e.Recalculate()
			DeepSync(panel)
			c.Equal("6", field.Text(), "tired, the current value is halved")
			c.Contains(tooltipText(field.Tooltip), "Adjusted from "+one.maximum, "tired, the tooltip names the maximum")
			c.Equal("Lifting ST, adjusted from "+one.maximum, spokenName(), "tired, the adjustment is spoken")
			c.Equal(adjustedInk, field.OnEditableInk, "tired, the value is drawn in adjustedInk")

			e.Attributes.Set[gurps.FatiguePointsID].Damage = 0
			e.Recalculate()
			DeepSync(panel)
			c.Equal(normalInk, field.OnEditableInk, "rested, the value is drawn in the normal ink again")
		})
	}
}

// TestAttrPanelPoolStateListsThresholdFeatures verifies that the state of a pool whose current threshold has features
// is drawn in adjustedInk, with those features listed in its tooltip after the explanation, a switchable one without
// saying it is, since a threshold has no switch, and each as written rather than as Markdown, and that the features a
// threshold ignores are neither listed nor enough to draw the state in adjustedInk.
func TestAttrPanelPoolStateListsThresholdFeatures(t *testing.T) {
	c := check.New(t)
	switchable := gurps.NewAttributeBonus(gurps.DexterityID)
	switchable.SetSwitchable(true)
	lore := gurps.NewTraitBonus()
	lore.NameCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Hidden _Lore_ of `Old`"}
	e := tiredEntity(t, switchable, lore)
	perLevel := gurps.NewAttributeBonus(gurps.StrengthID)
	perLevel.Amount = fxp.Two
	perLevel.PerLevel = true
	thisWeapon := gurps.NewWeaponBonus(feature.WeaponBonus)
	thisWeapon.SelectionType = wsel.ThisWeapon
	for _, one := range e.SheetSettings.Attributes.Set[gurps.FatiguePointsID].Thresholds {
		if one.State == "Tiring" {
			one.Features = gurps.Features{perLevel, gurps.NewAttributeBonus(gurps.FatiguePointsID), thisWeapon}
		}
	}
	panel := NewPointPoolsPanel(e, NewTargetMgr(unison.NewPanel()))
	label := panel.stateLabels[gurps.FatiguePointsID]
	c.NotNil(label, "the FP state label must be tracked")
	stateInk := func() unison.Ink {
		var ink unison.Ink
		label.Text.AdjustDecorations(func(d *unison.TextDecoration) { ink = d.OnBackgroundInk })
		return ink
	}
	c.Equal("[TIRED]", strings.ToUpper(label.Text.String()))
	c.Equal(adjustedInk, stateInk(), "a threshold with features is drawn in adjustedInk")
	md := findMarkdown(label.Tooltip)
	c.NotNil(md, "the tooltip must hold the explanation")
	text := markdownText(md)
	c.True(strings.HasPrefix(text, "Move, Dodge and ST are halved"), "the explanation leads the tooltip: %s", text)
	c.Contains(text, "\n•\n×0.5 to ", "the features are listed after it")
	c.Contains(text, "\n•\n+1 to DX\n", "a switchable one among them")
	c.NotContains(text, "switched on", "without a switch")
	c.Contains(text, "\n•\n+1 to the level of trait Hidden _Lore_ of `Old`\n", "a description shows as written")

	e.Attributes.Set[gurps.FatiguePointsID].Damage = fxp.Two
	e.Recalculate()
	panel.Sync()
	c.Equal("[TIRING]", strings.ToUpper(label.Text.String()))
	c.NotEqual(adjustedInk, stateInk(), "a threshold whose features all do nothing is drawn in the normal ink")
	c.Nil(label.Tooltip, "and lists none of them, there being no explanation either")
}
