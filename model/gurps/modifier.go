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
	"maps"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/xreflect"
)

// assertModifiableNode is used at compile time to check a *constraint*
func assertModifiableNode[T ModifiableNode[T, M], M ModifierNode[M, T]]() {}

// ModifiableNode is a Node constraint, narrowed for the Modifiable interface
type ModifiableNode[T ModifiableNode[T, M], M ModifierNode[M, T]] interface {
	Node[T]
	Modifiable[T, M]
}

// Modifiable is an interface for a type designed to have a matching Modifier
type Modifiable[T Modifiable[T, M], M Modifier[M, T]] interface {
	ModifierList() []M
	SetModifiers([]M)
	AddModifiers(...M)
}

// GeneralModifier is used for common access to modifiers.
type GeneralModifier interface {
	Container() bool
	Depth() int
	NameWithReplacements() string
	FullDescription() string
	FullCostDescription() string
	Enabled() bool
	SetEnabled(enabled bool)
}

// assertModifierNode is used at compile time to check a *constraint*
func assertModifierNode[M ModifierNode[M, T], T ModifiableNode[T, M]]() {}

// ModifierNode is Node constraint, narrowed for the Modifier interface
type ModifierNode[M ModifierNode[M, T], T ModifiableNode[T, M]] interface {
	Node[M]
	Modifier[M, T]
}

// Modifier is an interface for a type designed to have a matching Modifiable
type Modifier[M Modifier[M, T], T Modifiable[T, M]] interface {
	Target() T
	SetTarget(T) M
	GeneralModifier
}

// attachModifiers points each of the modifiers at the target and gives them the target's data owner.
func attachModifiers[T ModifiableNode[T, M], M ModifierNode[M, T], S ~[]M](target T, modifiers S) {
	owner := target.DataOwner()
	for _, m := range modifiers {
		m.SetDataOwner(owner)
		m.SetTarget(target)
	}
}

// activeModifierFor returns the first enabled, non-container modifier whose name matches (case-insensitive), or the
// zero value if there is none.
func activeModifierFor[M ModifierNode[M, T], T ModifiableNode[T, M], S ~[]M](modifiers S, name string) M {
	var found M
	Traverse(func(mod M) bool {
		if strings.EqualFold(mod.NameWithReplacements(), name) {
			found = mod
			return true
		}
		return false
	}, true, true, modifiers...)
	return found
}

// modifierDescriptions returns the full descriptions of the enabled, non-container modifiers, separated by "; ".
func modifierDescriptions[M ModifierNode[M, T], T ModifiableNode[T, M], S ~[]M](modifiers S) string {
	var buffer strings.Builder
	Traverse(func(mod M) bool {
		if buffer.Len() != 0 {
			buffer.WriteString("; ")
		}
		buffer.WriteString(mod.FullDescription())
		return false
	}, true, true, modifiers...)
	return buffer.String()
}

// fillWithModifierNameableKeys adds the nameable keys of the enabled modifiers, containers included, to m.
func fillWithModifierNameableKeys[M ModifierNode[M, T], T ModifiableNode[T, M], S ~[]M](modifiers S, m, existing map[string]string) {
	Traverse(func(mod M) bool {
		mod.FillWithNameableKeys(m, existing)
		return false
	}, true, false, modifiers...)
}

// mergeReplacements folds src into dst, keeping whatever value dst already holds for a key, and returns the result. A
// nil dst takes a copy of src rather than src itself, so that the result never shares storage with src.
func mergeReplacements[M ~map[string]string](dst, src M) M {
	if len(src) == 0 {
		return dst
	}
	if dst == nil {
		return maps.Clone(src)
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
