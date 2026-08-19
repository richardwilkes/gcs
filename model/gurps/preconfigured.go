// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import "github.com/richardwilkes/toolbox/v2/xreflect"

// Preconfigurable interface for objects that can have their selections and substitutions marked as preconfigured.
type Preconfigurable interface {
	IsPreconfigured() bool
	PreconfiguredRef() *bool
	SetPreconfigured(bool)
}

// preconfigurable self contained Preconfigurable implementation that can be embedded in another struct in this package
type preconfigurable struct {
	Preconfigured bool `json:"preconfigured,omitzero"`
}

// IsPreconfigured implements Preconfigurable.
func (tl *preconfigurable) IsPreconfigured() bool {
	return tl.Preconfigured
}

// PreconfiguredRef implements Preconfigurable.
func (tl *preconfigurable) PreconfiguredRef() *bool {
	return &tl.Preconfigured
}

// SetPreconfigured implements Preconfigurable.
func (tl *preconfigurable) SetPreconfigured(preconfigured bool) {
	tl.Preconfigured = preconfigured
}

// IsNodePreconfigured reports whether the given node has been marked as preconfigured, meaning its modifiers and nameable
// replacements are considered already finalized and shouldn't be prompted for again (e.g. when applying a template).
func IsNodePreconfigured[T Node[T]](node T) bool {
	if tl, ok := any(node).(Preconfigurable); ok && !xreflect.IsNil(tl) {
		return tl.IsPreconfigured()
	}
	return false
}
