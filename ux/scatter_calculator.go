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
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

var (
	_ calculatorTab = &scatterCalculator{}

	scatterCauses = []scatterCause{
		{name: i18n.Text("Failed attack roll")},
		{name: i18n.Text("Failed attack roll, squared miss"), squared: true},
		{name: i18n.Text("Target dodged")},
	}
)

// scatterCause is why the attack missed, which decides whether it scatters by the margin or by its square (BX414). A
// dodge never squares the margin.
type scatterCause struct {
	name    string
	squared bool
}

func (s scatterCause) String() string {
	return s.name
}

// scatterCalculator works out how far an attack that missed lands from where it was aimed (BX414). Everything it needs
// is typed in, since a miss is a matter of the roll rather than of anything a sheet knows.
type scatterCalculator struct {
	calculatorContent
	marginField   *IntegerField
	distanceField *DecimalField
	result        *unison.Label
	distance      fxp.Int
	causeIndex    int
	margin        int
}

func newScatterCalculator() *scatterCalculator {
	s := &scatterCalculator{
		distance: fxp.Ten,
		margin:   1,
	}
	s.createContent()
	return s
}

// title implements calculatorTab.
func (s *scatterCalculator) title() string {
	return i18n.Text("Scatter")
}

// panel implements calculatorTab.
func (s *scatterCalculator) panel() *unison.Panel {
	return s.content
}

// preselect implements calculatorTab. Nothing here comes from a sheet.
func (s *scatterCalculator) preselect(_ *Sheet) {
}

// sheetChanged implements calculatorTab. Nothing here comes from a sheet.
func (s *scatterCalculator) sheetChanged(_ *Sheet) {
}

func (s *scatterCalculator) createContent() {
	s.initCalculatorContent()
	s.content.AddChild(s.createHeader(i18n.Text("Scatter"), []linkSpec{{pageRef: "BX414", highlight: "Scatter"}}, 0))
	row := s.addRow(2)
	addPlainLabel(row, i18n.Text("Cause of the miss:"))
	addIndexPopup(row, scatterCauses, &s.causeIndex, s.changed)
	s.marginField = sameWidth(NewIntegerField(nil, "", i18n.Text("Margin"),
		func() int { return s.margin },
		func(v int) {
			s.margin = v
			s.changed()
		},
		0, 100, false, false))
	s.addFieldRow(s.marginField, i18n.Text("points of margin"))
	s.distanceField = sameWidth(NewDecimalField(nil, "", i18n.Text("Distance to Target"),
		func() fxp.Int { return s.distance },
		func(v fxp.Int) {
			s.distance = v
			s.changed()
		},
		0, fxp.Max, false, false))
	s.addFieldRow(s.distanceField, i18n.Text("yards to the target"))
	row = s.addResultRow()
	addPlainLabel(row, i18n.Text("Scatter:"))
	s.result = addResultLabel(row)
	s.addNotes(
		i18n.Text("The miss is squared when the target was flying or underwater, or when Artillery or Dropping was used against a target the attacker could not see; a dodge is never squared. When the target dodged, the margin is its margin of success."),
		i18n.Text("Roll 1d for the direction: a 1 is the direction the attacker faces, and each higher number turns 60° further clockwise."),
		fmt.Sprintf(i18n.Text("Deliberately attacking an area rather than a target standing in it is at %+d to hit. The area cannot defend, though anyone in it may dive for cover."),
			gurps.AreaAttackBonus),
	)
}

// changed implements calculatorTab.
func (s *scatterCalculator) changed() {
	yards, capped := gurps.ScatterDistance(s.margin, s.distance, scatterCauses[s.causeIndex].squared)
	text := fmt.Sprintf(i18n.Text("%s yards"), yards.Comma())
	if capped {
		text += i18n.Text(" (limited to half the distance)")
	}
	s.result.SetTitle(text)
	s.result.MarkForLayoutRecursivelyUpward()
	s.content.MarkForRedraw()
}
