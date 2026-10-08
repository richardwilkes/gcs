// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package calculator

import "github.com/richardwilkes/toolbox/v2/i18n"

// ExplosionEnvironmentChoice names a medium a blast can spread through, which decides how fast the collateral damage
// falls off with distance (BX414-BX415).
type ExplosionEnvironmentChoice struct {
	Name        string
	Environment ExplosionEnvironment
}

// String implements fmt.Stringer.
func (e ExplosionEnvironmentChoice) String() string {
	return e.Name
}

// ExplosionEnvironmentChoices returns the media a blast can spread through (BX414-BX415).
func ExplosionEnvironmentChoices() []ExplosionEnvironmentChoice {
	return []ExplosionEnvironmentChoice{
		{Name: i18n.Text("Air"), Environment: ExplosionInAir},
		{Name: i18n.Text("Underwater"), Environment: ExplosionUnderwater},
		{Name: i18n.Text("Vacuum or trace atmosphere"), Environment: ExplosionInVacuum},
	}
}

// TargetPostureChoice names a posture the target can be in when the fragments arrive, which sets the penalty they
// take to hit it (BX551). It is the only posture modifier the fragmentation roll takes, and an airburst ignores it
// (BX415).
type TargetPostureChoice struct {
	Name    string
	Posture TargetPosture
}

// String implements fmt.Stringer.
func (p TargetPostureChoice) String() string {
	return p.Name
}

// TargetPostureChoices returns the postures the target can be in when the fragments arrive (BX551).
func TargetPostureChoices() []TargetPostureChoice {
	return []TargetPostureChoice{
		{Name: i18n.Text("Standing"), Posture: StandingTarget},
		{Name: i18n.Text("Crouching, kneeling or sitting (-2)"), Posture: CrouchingTarget},
		{Name: i18n.Text("Crawling or lying down (-2; -4 from a ground burst)"), Posture: ProneTarget},
	}
}

// BlastSituationChoice names where the target can be when the explosion goes off.
type BlastSituationChoice struct {
	Name      string
	Situation BlastSituation
}

// String implements fmt.Stringer.
func (s BlastSituationChoice) String() string {
	return s.Name
}

// BlastSituationChoices returns where the target can be when the explosion goes off, starting with CaughtInBlast.
func BlastSituationChoices() []BlastSituationChoice {
	return []BlastSituationChoice{
		{Name: i18n.Text("Caught in the blast"), Situation: CaughtInBlast},
		{Name: i18n.Text("Struck directly"), Situation: StruckDirectly},
		{Name: i18n.Text("Threw himself on the explosive"), Situation: ThrewSelfOnExplosive},
		{Name: i18n.Text("The explosive went off inside him"), Situation: ExplosiveInsideTarget},
	}
}

// ScatterCause is why an attack missed, which decides whether it scatters by the margin or by its square (BX414). A
// dodge never squares the margin.
type ScatterCause struct {
	Name    string
	Squared bool
}

// String implements fmt.Stringer.
func (s ScatterCause) String() string {
	return s.Name
}

// ScatterCauses returns the reasons an attack can miss (BX414).
func ScatterCauses() []ScatterCause {
	return []ScatterCause{
		{Name: i18n.Text("Failed attack roll")},
		{Name: i18n.Text("Failed attack roll, squared miss"), Squared: true},
		{Name: i18n.Text("Target dodged")},
	}
}
