// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/xreflect"
)

// assertModifiable is used at compile time to check a *constraint*
func assertModifiable[N Modifiable[MN, TN], MN ModifierNode[MN, TN], TN Node[TN]]() {}

// Modifiable represents an interface for an object that accepts modifier nodes
type Modifiable[T ModifierNode[T, TN], TN Node[TN]] interface {
	ModifierList() []T
	SetModifiers([]T)
	AddModifier(T)
}

// GeneralModifier is used for common access to modifiers.
type GeneralModifier interface {
	Container() bool
	Depth() int
	FullDescription() string
	FullCostDescription() string
	Enabled() bool
	SetEnabled(enabled bool)
}

// assertModifierNode is used at compile time to check a *constraint*
func assertModifierNode[T ModifierNode[T, TN], TN Node[TN]]() {}

// ModifierNode is a narrowed constraint for a Node
// The narrowed constraint enforces that TargetNode and SetTargetNode functions are satisfied.
// The narrowed constraint enforces the GeneralModifier interface is satisfied.
type ModifierNode[N Node[N], TN Node[TN]] interface {
	Node[N]
	TargetNode() TN
	SetTargetNode(TN)
	GeneralModifier
}

// modifiable is a default implementation that satisfies the Modifiable constraint
type modifiable[T ModifierNode[T, TN], TN Node[TN]] struct {
	Modifiers []T `json:"modifiers,omitempty"`
}

// ModifierList returns the list of modifiers
func (m *modifiable[T, TN]) ModifierList() []T {
	return m.Modifiers
}

// SetModifiers sets the list of modifiers
func (m *modifiable[T, TN]) SetModifiers(all []T) {
	m.Modifiers = all
}

// AddModifier adds a modifier to the list
func (m *modifiable[T, TN]) AddModifier(mod T) {
	m.Modifiers = append(m.Modifiers, mod)
}

// SetTargetNode calls SetTargetNode on each modifier.
func (m *modifiable[T, TN]) SetTargetNode(target TN) {
	for _, one := range m.Modifiers {
		one.SetTargetNode(target)
	}
}

// mergeReplacements folds src into dst, keeping whatever value dst already holds for a key, and returns the result. A
// nil dst simply takes src.
func mergeReplacements(dst, src map[string]string) map[string]string {
	if len(src) == 0 {
		return dst
	}
	if dst == nil {
		return src
	}
	for k, v := range src {
		if _, exists := dst[k]; !exists {
			dst[k] = v
		}
	}
	return dst
}

// applyOwnerReplacements applies the nameable replacements of the trait or equipment that owns a modifier to s. A
// modifier without an owner returns s untouched, markers and all. That is deliberately not the same as calling
// nameable.Apply with a nil map, which would render any unresolved markers in their compact form and so change what
// an unattached modifier (in a library file editor, for example) displays.
func applyOwnerReplacements[T any, PT interface {
	*T
	nameable.Accesser
}](s string, owner PT) string {
	if owner == nil {
		return s
	}
	return nameable.Apply(s, owner.NameableReplacements())
}

// modifierSecondaryText returns the "secondary" text for a modifier: its resolved local notes, provided the sheet
// settings say notes are shown in the way optionChecker asks about (inline or as a tooltip).
func modifierSecondaryText[T interface {
	Node[T]
	ResolveLocalNotes() string
}](node T, optionChecker func(display.Option) bool) string {
	if !optionChecker(SheetSettingsFor(EntityFromNode(node)).NotesDisplay) {
		return ""
	}
	return node.ResolveLocalNotes()
}

// syncFromSource looks the node up in its data owner's source matcher and, when the node has drifted from its library
// source, hands the library's copy to apply so the node can pull the synced fields across. A node with no data owner,
// no source, or a source it already matches is left alone.
func syncFromSource[T Node[T]](node T, apply func(source T)) {
	owner := node.DataOwner()
	if xreflect.IsNil(owner) {
		return
	}
	if state, data := owner.SourceMatcher().Match(node); state == srcstate.Mismatched {
		if source, ok := data.(T); ok {
			apply(source)
		}
	}
}
