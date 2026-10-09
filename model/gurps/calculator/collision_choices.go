// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package calculator

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/i18n"
)

// CollisionShape is the shape of an object in a collision, which decides the type of damage it inflicts and whether it
// does half damage: a bullet-shaped, sharp or spiked object does half damage, but of its own type rather than crushing
// (BX430).
type CollisionShape struct {
	Name       string
	DamageType string
	HalfDamage bool
}

// String implements fmt.Stringer.
func (s CollisionShape) String() string {
	return s.Name
}

// CollisionShapes returns the shapes an object in a collision can have (BX430).
func CollisionShapes() []CollisionShape {
	return []CollisionShape{
		{Name: i18n.Text("Blunt"), DamageType: "cr"},
		{Name: i18n.Text("Bullet-shaped"), DamageType: "pi", HalfDamage: true},
		{Name: i18n.Text("Sharp"), DamageType: "cut", HalfDamage: true},
		{Name: i18n.Text("Spiked"), DamageType: "imp", HalfDamage: true},
	}
}

// CollisionSurface is a kind of immovable object (BX431): a hard one is hit as if the mover had twice its HP, an
// elastic one gives extra DR, and water can be dived into cleanly.
type CollisionSurface struct {
	Name    string
	Hard    bool
	Elastic bool
	Water   bool
}

// String implements fmt.Stringer.
func (s CollisionSurface) String() string {
	return s.Name
}

// CollisionSurfaces returns the kinds of immovable object something can fall onto or hit (BX431). The first is the
// hard surface, which a sudden stop is worked out against.
func CollisionSurfaces() []CollisionSurface {
	return []CollisionSurface{
		{Name: i18n.Text("Hard (ground, concrete, a wall)"), Hard: true},
		{Name: i18n.Text("Soft (forest litter, hay, swamp)")},
		{Name: i18n.Text("Soft and elastic (mattress, net, airbag)"), Elastic: true},
		{Name: i18n.Text("Water or another fluid"), Water: true},
	}
}

// TerminalVelocityChoice is a terminal velocity at one gravity in one atmosphere (BX431); BaseVelocity is zero for the
// choice that applies no limit, and Custom marks the one whose value is typed in.
type TerminalVelocityChoice struct {
	Name         string
	BaseVelocity fxp.Int
	Custom       bool
}

// String implements fmt.Stringer.
func (t TerminalVelocityChoice) String() string {
	return t.Name
}

// TerminalVelocityChoices returns the terminal velocities a fall can be limited to (BX431), ending with the custom one.
func TerminalVelocityChoices() []TerminalVelocityChoice {
	return []TerminalVelocityChoice{
		{Name: i18n.Text("Human, spread-eagled (60)"), BaseVelocity: fxp.Sixty},
		{Name: i18n.Text("Human, swan dive (100)"), BaseVelocity: fxp.Hundred},
		{Name: i18n.Text("Dense or streamlined object (200)"), BaseVelocity: fxp.FromInteger(200)},
		{Name: i18n.Text("None")},
		{Name: i18n.Text("Custom"), Custom: true},
	}
}

// CollisionAngleChoice names one of the angles two objects can collide at (BX432).
type CollisionAngleChoice struct {
	Name  string
	Angle CollisionAngle
}

// String implements fmt.Stringer.
func (a CollisionAngleChoice) String() string {
	return a.Name
}

// CollisionAngleChoices returns the angles two objects can collide at (BX432).
func CollisionAngleChoices() []CollisionAngleChoice {
	return []CollisionAngleChoice{
		{Name: i18n.Text("Head-on"), Angle: HeadOnCollision},
		{Name: i18n.Text("Rear-end"), Angle: RearEndCollision},
		{Name: i18n.Text("Side-on, or the struck object is stationary"), Angle: SideOnCollision},
	}
}

// CollisionRestraint is what holds an occupant in place during a sudden stop, and the DR it gives against the damage
// (BX432).
type CollisionRestraint struct {
	Name string
	DR   int
}

// String implements fmt.Stringer.
func (r CollisionRestraint) String() string {
	return r.Name
}

// CollisionRestraints returns what can hold an occupant in place during a sudden stop (BX432).
func CollisionRestraints() []CollisionRestraint {
	return []CollisionRestraint{
		{Name: i18n.Text("None")},
		{Name: i18n.Text("Seatbelt or straps (DR 5)"), DR: 5},
		{Name: i18n.Text("Airbag (DR 10)"), DR: 10},
	}
}
