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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestAttrPanelHashReflectsTraitDrivenVisibility verifies that the panel's hash changes when a trait is added, removed,
// enabled, or disabled in a way that reveals or hides an attribute. Without folding the trait-driven placement into the
// hash, Sync would not rebuild the panel and the revealed/hidden attribute would not update on the sheet.
func TestAttrPanelHashReflectsTraitDrivenVisibility(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()

	// A hidden secondary attribute that becomes visible when the character has the "Magery" trait.
	defs := &gurps.AttributeDefs{Set: map[string]*gurps.AttributeDef{
		"sm": {AttributeDefData: gurps.AttributeDefData{
			DefID:                "sm",
			Type:                 attribute.Integer,
			Base:                 "10",
			Placement:            attribute.Hidden,
			PlacementTrait:       "Magery",
			PlacementWhenPresent: attribute.Secondary,
		}},
	}}

	panel := &AttrPanel{entity: e, kind: gurps.SecondaryAttrKind}
	hidden := panel.computeHash(defs)

	// Adding the matching trait must change the hash so the panel rebuilds and reveals the attribute.
	trait := gurps.NewTrait(e, nil, false)
	trait.Name = "Magery"
	e.Traits = append(e.Traits, trait)
	revealed := panel.computeHash(defs)
	c.NotEqual(hidden, revealed, "adding the trait must change the hash")

	// Disabling the trait must return the hash to the hidden state.
	trait.Disabled = true
	c.Equal(hidden, panel.computeHash(defs), "disabling the trait must restore the hidden hash")

	// Re-enabling it reveals the attribute again.
	trait.Disabled = false
	c.Equal(revealed, panel.computeHash(defs), "re-enabling the trait must restore the revealed hash")

	// Removing the trait returns to the hidden state.
	e.Traits = nil
	c.Equal(hidden, panel.computeHash(defs), "removing the trait must restore the hidden hash")
}

// TestAttrPanelNameTooltipTracksBonuses verifies that the tooltip on an attribute's name label lists the sources of
// its bonuses and is refreshed by Sync, since a change to a trait or piece of equipment alters the bonuses without
// changing the attribute definitions, so the panel is not rebuilt.
func TestAttrPanelNameTooltipTracksBonuses(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	e.Recalculate()
	panel := NewPrimaryAttrPanel(e, NewTargetMgr(unison.NewPanel()))
	label := panel.nameLabels[gurps.StrengthID]
	c.NotNil(label, "the ST name label must be tracked")
	c.Nil(label.Tooltip, "no bonuses means no tooltip")

	trait := gurps.NewTrait(e, nil, false)
	trait.Name = "Increased Strength"
	bonus := gurps.NewAttributeBonus(gurps.StrengthID)
	bonus.Amount = fxp.Two
	trait.Features = gurps.Features{bonus}
	e.Traits = append(e.Traits, trait)
	e.Recalculate()
	hash := panel.hash
	panel.Sync()
	c.Equal(hash, panel.hash, "a change to the bonuses alone must not rebuild the panel")
	c.True(label == panel.nameLabels[gurps.StrengthID], "the existing label must be refreshed in place")
	c.NotNil(label.Tooltip, "a bonus must produce a tooltip")
	c.Equal("Includes modifiers from:\nIncreased Strength [+2]", tooltipText(label.Tooltip),
		"the tooltip lists the bonus and its source")

	e.Traits = nil
	e.Recalculate()
	panel.Sync()
	c.Nil(label.Tooltip, "removing the bonus removes the tooltip")
}

// TestAttrPanelPoolNameTooltipKeepsFullName verifies that a point pool's name label keeps showing the pool's full name
// in its tooltip, with the bonus sources following it when there are any.
func TestAttrPanelPoolNameTooltipKeepsFullName(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	e.Recalculate()
	panel := NewPointPoolsPanel(e, NewTargetMgr(unison.NewPanel()))
	label := panel.nameLabels["fp"]
	c.NotNil(label, "the FP name label must be tracked")
	c.NotNil(label.Tooltip, "a pool with a full name always has a tooltip")
	c.Equal("Fatigue Points", tooltipText(label.Tooltip), "without bonuses, only the full name is shown")

	trait := gurps.NewTrait(e, nil, false)
	trait.Name = "Fit"
	bonus := gurps.NewAttributeBonus("fp")
	bonus.Amount = fxp.Two
	trait.Features = gurps.Features{bonus}
	e.Traits = append(e.Traits, trait)
	e.Recalculate()
	panel.Sync()
	c.Equal("Fatigue Points\nIncludes modifiers from:\nFit [+2]", tooltipText(label.Tooltip),
		"the full name leads the bonus sources")

	e.Traits = nil
	e.Recalculate()
	panel.Sync()
	c.Equal("Fatigue Points", tooltipText(label.Tooltip),
		"removing the bonus leaves just the full name behind")
}

// TestAttrPanelSecondaryNameTooltipTracksBonuses verifies that the secondary attributes panel installs and refreshes
// the bonus tooltip the same way the primary one does, since it is built by the same code but for a different set of
// attributes and had no coverage of its own.
func TestAttrPanelSecondaryNameTooltipTracksBonuses(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	e.Recalculate()
	panel := NewSecondaryAttrPanel(e, NewTargetMgr(unison.NewPanel()))
	label := panel.nameLabels["will"]
	c.NotNil(label, "the Will name label must be tracked")
	c.Nil(label.Tooltip, "no bonuses means no tooltip")

	trait := gurps.NewTrait(e, nil, false)
	trait.Name = "Strong Will"
	bonus := gurps.NewAttributeBonus("will")
	bonus.Amount = fxp.Two
	trait.Features = gurps.Features{bonus}
	e.Traits = append(e.Traits, trait)
	e.Recalculate()
	hash := panel.hash
	panel.Sync()
	c.Equal(hash, panel.hash, "a change to the bonuses alone must not rebuild the panel")
	c.True(label == panel.nameLabels["will"], "the existing label must be refreshed in place")
	c.Equal("Includes modifiers from:\nStrong Will [+2]", tooltipText(label.Tooltip),
		"the tooltip lists the bonus and its source")

	e.Traits = nil
	e.Recalculate()
	panel.Sync()
	c.Nil(label.Tooltip, "removing the bonus removes the tooltip")
}

// TestAttrPanelValueFieldTooltipMatchesNameLabel verifies that the value field of an attribute's row carries the same
// bonus tooltip as its name label, in all three kinds of panel, and that both are kept up to date by Sync. Users hover
// over the number at least as often as the name, so the sources of the bonuses must be reachable from there too.
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
				c.Nil(field.AsPanel().Tooltip, "no bonuses means no tooltip on the value field")
			} else {
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
			panel.Sync()
			c.True(field == panel.valueFields[one.attrID], "the existing value field must be refreshed in place")
			c.Equal(one.withBonus, tooltipText(field.AsPanel().Tooltip),
				"the value field lists the bonus and its source")
			c.Equal(tooltipText(label.Tooltip), tooltipText(field.AsPanel().Tooltip),
				"the value field and the name label show the same tooltip")

			e.Traits = nil
			e.Recalculate()
			panel.Sync()
			if one.withoutTip == "" {
				c.Nil(field.AsPanel().Tooltip, "removing the bonus removes the value field's tooltip")
			} else {
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
