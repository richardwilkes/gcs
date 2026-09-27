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

// CanTakeModifiers returns true if modifiers may be attached to the node. A template choice container can't take them,
// since it dissolves into the options chosen from it, so modifiers attached to it would only change the cost of those
// options on the template, and its editor offers no way to remove them. An equipment group can't either: it only
// organizes what it holds, so it keeps nothing of its own, modifiers included (see Equipment.ClearUnusedFieldsForType).
func CanTakeModifiers[T Node[T]](node T) bool {
	if xreflect.IsNil(node) || IsTemplateChoiceContainer(node) {
		return false
	}
	eqp, isEquipment := any(node).(*Equipment)
	return !isEquipment || !eqp.IsGroup()
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
	// fillWithNameableKeysEvenIfDisabled is FillWithNameableKeys without the enabled check.
	fillWithNameableKeysEvenIfDisabled(m, existing map[string]string)
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

// ownerNameableReplacements returns the replacements a trait or piece of equipment should hold after applying m, the
// answers for the keys it has in use: those answers, reduced to the keys in use, plus whatever it already holds for
// keys only its disabled modifiers use, since those are wanted again when such a modifier is re-enabled. Anything else
// held for a key no longer in use is dropped. Returns nil when there is nothing to hold.
func ownerNameableReplacements[T ModifiableNode[T, M], M ModifierNode[M, T]](owner T, existing, m map[string]string) map[string]string {
	inUse := make(map[string]string)
	owner.FillWithNameableKeys(inUse, nil)
	result := nameable.Reduce(inUse, m)
	if len(existing) == 0 {
		return result
	}
	all := make(map[string]string)
	Traverse(func(mod M) bool {
		mod.fillWithNameableKeysEvenIfDisabled(all, existing)
		return false
	}, false, false, owner.ModifierList()...)
	for k := range all {
		if _, used := inUse[k]; used {
			continue
		}
		if v, held := existing[k]; held {
			if result == nil {
				result = make(map[string]string)
			}
			result[k] = v
		}
	}
	return result
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

// modifierNameableReplacements returns the replacements a modifier's target should hold after applying m, the answers
// for the modifier's own keys. A modifier keeps no replacements of its own, so the answers are merged into a copy of
// what the target holds rather than replacing it: an answer overrides the target's value, a key left at nameable.Unset
// keeps it, and a key of the modifier's missing from m (a substitution the user cleared) is removed. Everything else
// the target holds is kept, including answers for disabled modifiers. Returns nil when there is nothing to hold.
func modifierNameableReplacements(existing map[string]string, modifier nameable.Filler, m map[string]string) map[string]string {
	merged := maps.Clone(existing)
	if merged == nil {
		merged = make(map[string]string, len(m))
	}
	for k, v := range m {
		if v != nameable.Unset {
			merged[k] = v
		}
	}
	own := make(map[string]string)
	modifier.FillWithNameableKeys(own, nil)
	for k := range own {
		if _, answered := m[k]; !answered {
			delete(merged, k)
		}
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
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
