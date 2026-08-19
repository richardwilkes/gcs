// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

// FeatureSwitch is embedded in every persisted feature type. It records whether the feature is "switchable": a
// switchable feature only takes effect while the switch of the primary item that owns it (a trait, skill, spell, or
// piece of equipment) is on. Features that are not switchable always apply. Note that the on/off state itself is not
// stored here, but on the owning item (see FeatureSwitcher), so a modifier's switchable features follow the switch of
// the trait or equipment the modifier belongs to.
type FeatureSwitch struct {
	Switchable bool `json:"switchable,omitzero"`
}

// IsSwitchable implements Feature.
func (s *FeatureSwitch) IsSwitchable() bool {
	return s.Switchable
}

// SetSwitchable implements Feature.
func (s *FeatureSwitch) SetSwitchable(switchable bool) {
	s.Switchable = switchable
}

// ItemSwitch is embedded in the edit data of every primary item type that can own switchable features (a trait, skill,
// spell, or piece of equipment). It holds the on/off state of the switch that controls those features, including the
// ones contributed by the item's modifiers. The state is a local choice made on the sheet rather than part of the
// item's source data, so it is deliberately left out of the item's hash.
type ItemSwitch struct {
	SwitchedOn bool `json:"switched_on,omitzero"`
}

// IsSwitchedOn implements FeatureSwitcher.
func (s *ItemSwitch) IsSwitchedOn() bool {
	return s.SwitchedOn
}

// SetSwitchedOn implements FeatureSwitcher.
func (s *ItemSwitch) SetSwitchedOn(on bool) {
	s.SwitchedOn = on
}

// FeatureSwitcher is implemented by the primary data types that own features and therefore carry the switch that
// controls their switchable features.
type FeatureSwitcher interface {
	// HasSwitchableFeatures returns true if any of the features this item currently contributes could be switched,
	// i.e. any of its own features or those of its enabled modifiers are marked as switchable.
	HasSwitchableFeatures() bool
	// IsSwitchedOn returns true if the switch for this item's switchable features is on.
	IsSwitchedOn() bool
	// SetSwitchedOn sets whether the switch for this item's switchable features is on.
	SetSwitchedOn(on bool)
}

// anyModifierSwitchable returns true if any of the given modifiers has a switchable feature. The traversal matches the
// one used when features are collected for a character (see Entity.processFeatures), i.e. only enabled, non-container
// modifiers are considered, so that this answer always agrees with what will actually be applied.
func anyModifierSwitchable[T Node[T]](modifiers []T, features func(T) Features) bool {
	found := false
	Traverse(func(one T) bool {
		found = found || features(one).AnySwitchable()
		return found
	}, true, true, modifiers...)
	return found
}
