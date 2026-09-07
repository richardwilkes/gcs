// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

// FeatureSwitch is embedded in every persisted feature type. A switchable feature only takes effect while the switch of
// the primary item that owns it (a trait, skill, spell, or piece of equipment) is on; other features always apply. The
// on/off state lives on that item (see FeatureSwitcher), not here, so a modifier's switchable features follow the
// switch of the trait or equipment the modifier belongs to.
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
// spell, or piece of equipment). It holds the on/off state of the switch controlling those features, including the ones
// its modifiers contribute. The state is a local choice made on the sheet rather than part of the item's source data,
// so it is deliberately left out of the item's hash.
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
	IsSwitchedOn() bool
	SetSwitchedOn(on bool)
}

// anyModifierSwitchable returns true if any of the given modifiers has a switchable feature. Only enabled,
// non-container modifiers are considered, matching what is collected for a character (see Entity.processFeatures), so
// this always agrees with what will actually be applied.
func anyModifierSwitchable[T Node[T]](modifiers []T, features func(T) Features) bool {
	return anyEnabledNonContainerModifier(modifiers, func(mod T) bool { return features(mod).AnySwitchable() })
}

// visitEnabledModifiers calls visit for each enabled, non-container modifier among the given ones, at any depth, in
// Traverse order, handing it the modifier's features that take effect for the given state of the switch on the item the
// modifiers belong to (see Features.Active).
func visitEnabledModifiers[T Node[T]](modifiers []T, switchedOn bool, features func(T) Features, visit func(mod T, active Features)) {
	Traverse(func(mod T) bool {
		visit(mod, features(mod).Active(switchedOn))
		return false
	}, true, true, modifiers...)
}

// anyEnabledNonContainerModifier returns true if the given predicate holds for any enabled, non-container modifier
// among the given ones, at any depth, descending only through enabled containers -- exactly the set Traverse(f, true,
// true, modifiers...) visits. It is a plain recursion rather than a Traverse call, since Traverse clones the children
// of every container it descends into and this runs from CellData for every row on every sort and every keystroke of a
// search.
func anyEnabledNonContainerModifier[T Node[T]](modifiers []T, predicate func(T) bool) bool {
	for _, mod := range modifiers {
		if !mod.Enabled() {
			continue
		}
		if !mod.Container() && predicate(mod) {
			return true
		}
		if mod.HasChildren() && anyEnabledNonContainerModifier(mod.NodeChildren(), predicate) {
			return true
		}
	}
	return false
}
