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
	"fmt"
	"hash"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// TemplatePickerProvider provides access to the valid picker types and the picker data.
type TemplatePickerProvider interface {
	// TemplatePickerData returns the valid picker types and a non-nil pointer to the TemplatePicker.
	TemplatePickerData() ([]picker.Type, *TemplatePicker)
}

// TemplatePickerNode is a constraint for a Node that implements the TemplatePickerProvider interface
type TemplatePickerNode[T TemplatePickerNode[T]] interface {
	Node[T]
	TemplatePickerProvider
}

// assertTemplatePickerNode causes a compile-time constraint validation
func assertTemplatePickerNode[T TemplatePickerNode[T]]() {}

// ModifiableTemplatePickerNode is a constraint for a TemplatePickerNode that can also hold modifiers, and so can have
// modifiers on a container presenting a template choice.
type ModifiableTemplatePickerNode[T ModifiableTemplatePickerNode[T, M], M ModifierNode[M, T]] interface {
	TemplatePickerNode[T]
	ModifiableNode[T, M]
	nameable.Setter
}

// assertModifiableTemplatePickerNode causes a compile-time constraint validation
func assertModifiableTemplatePickerNode[T ModifiableTemplatePickerNode[T, M], M ModifierNode[M, T]]() {
}

// TemplatePicker holds the data necessary to allow a template choice to be made.
type TemplatePicker struct {
	Type      picker.Type     `json:"type"`
	Qualifier criteria.Number `json:"qualifier,omitzero"`
}

// IsZero implements json.isZero.
func (t TemplatePicker) IsZero() bool {
	return t.Type == picker.NotApplicable
}

func (t TemplatePicker) String() string {
	if t.IsZero() {
		return ""
	}
	switch t.Type {
	case picker.Count:
		return fmt.Sprintf(i18n.Text("Pick %s"), t.Qualifier.AltString())
	case picker.Points:
		points := i18n.Text("points")
		if t.Qualifier.Qualifier == fxp.One {
			points = i18n.Text("point")
		}
		return fmt.Sprintf(i18n.Text("Pick %s %s worth"), t.Qualifier.AltString(), points)
	default:
		return ""
	}
}

// Hash writes this object's contents into the hasher.
func (t TemplatePicker) Hash(h hash.Hash) {
	xhash.Num8(h, t.Type)
	if t.Type != picker.NotApplicable {
		t.Qualifier.Hash(h)
	}
}

// HasTemplatePickerData returns true if any node or their child has non-zero template picker data
func HasTemplatePickerData[T Node[T]](nodes ...T) bool {
	var hasPickerData bool
	Traverse(func(node T) bool {
		if node.Container() {
			if tpp, ok := any(node).(TemplatePickerProvider); ok {
				if _, data := tpp.TemplatePickerData(); !data.IsZero() {
					hasPickerData = true
					return true
				}
			}
		}
		return false
	}, false, false, nodes...)
	return hasPickerData
}

// ClearTemplatePickerData removes the template picker data from the nodes and their children. A node that loses its
// picker data also loses its source, since only a template may hold picker data and a template is never a source.
func ClearTemplatePickerData[T Node[T]](nodes ...T) {
	Traverse(func(node T) bool {
		if node.Container() {
			if tpp, ok := any(node).(TemplatePickerProvider); ok {
				if _, data := tpp.TemplatePickerData(); !data.IsZero() {
					*data = TemplatePicker{}
					node.ClearSource()
				}
			}
		}
		return false
	}, false, false, nodes...)
}

// projectPickerModifiersDownward moves the enabled modifiers held by each container that presents a template choice
// (one with template picker data) down onto the rows beneath it, anywhere in the given rows, and removes its disabled
// ones. It is run when a template is loaded, so that a modifier authored on a choice container keeps applying to
// whatever is chosen from it.
//
// A modifier on an ordinary container applies to everything beneath it, since a row's modifiers include those of its
// parents (see Trait.AllModifiers). A choice container, though, is dissolved when the template is applied: only the
// chosen children are kept, re-parented to the container's own parent, and anything held by the container itself is
// left behind. Projecting the modifiers down ahead of time is what lets them survive that.
//
// A disabled modifier on a choice container can never take effect, since it is off while the template is open and is
// left behind with the container when the template is applied, so it is removed rather than projected. A modifier
// container is projected with only the enabled modifiers within it, and one with none is removed as well.
//
// Each child receives its own copy of the container's modifiers, appended after any it already has, with a fresh ID
// and the original's Source. A child that is an ordinary container is a fine place for the copy to rest, since it
// survives the dissolve intact and passes the copy on to everything beneath it. A child that is itself a choice
// container is not, since it will be dissolved in turn, so the copies pass through it to its own children. Nested
// choice containers are handled innermost first, which leaves each chosen row's inherited modifiers in the same order
// they were found in before.
//
// The container's own modifiers are then removed. This is required, not tidiness: the modifier would otherwise be
// found both on the child and on its parent, so a -10% modifier would count as -20% for as long as the template is
// open. It also makes the projection idempotent, as a second pass finds nothing left to move.
//
// A modifier's nameable replacements are held by the row it is attached to, so the values the container supplies for
// the markers its modifiers use are carried along to each row that receives the copies. A row that already has a value
// for one of those markers keeps its own.
//
// A row that had no modifiers of its own is marked as preconfigured, since the ones it receives are all enabled and so
// there is nothing for the player to decide about them when the template is applied: the author's choice is already
// made. A row that had modifiers of its own is left as it was, so that the player is still asked about them. This is
// judged when the first copies arrive, so a row beneath nested choice containers, which receives copies from each of
// them, keeps the mark it got from the innermost one. A row the author already marked keeps its mark.
//
// When there is no row for the copies to land on (the container has no children, or only choice containers with none
// of their own) the enabled modifiers are left where they are, rather than being quietly thrown away.
func projectPickerModifiersDownward[T ModifiableTemplatePickerNode[T, M], M ModifierNode[M, T]](rows []T) {
	for _, row := range rows {
		if !row.Container() {
			continue
		}
		projectPickerModifiersDownward(row.NodeChildren())
		if _, tp := row.TemplatePickerData(); tp.IsZero() || len(row.ModifierList()) == 0 {
			continue
		}
		modifiers := enabledModifiers(row.ModifierList())
		row.SetModifiers(modifiers)
		if len(modifiers) == 0 {
			continue
		}
		keys := make(map[string]string)
		fillWithModifierNameableKeys(modifiers, keys, nil)
		replacements := nameable.Reduce(keys, row.NameableReplacements())
		if projectModifiersOnto(row.NodeChildren(), modifiers, replacements) {
			row.SetModifiers(nil)
		}
	}
}

// projectModifiersOnto appends a copy of the modifiers to each of the rows, passing through any that are themselves
// choice containers to the rows beneath them, and returns true if at least one row received a copy.
func projectModifiersOnto[T ModifiableTemplatePickerNode[T, M], M ModifierNode[M, T]](rows []T, modifiers []M,
	replacements map[string]string,
) bool {
	landed := false
	for _, row := range rows {
		if row.Container() {
			if _, tp := row.TemplatePickerData(); !tp.IsZero() {
				if projectModifiersOnto(row.NodeChildren(), modifiers, replacements) {
					landed = true
				}
				continue
			}
		}
		if len(row.ModifierList()) == 0 && IsNodePreconfigurable(row) {
			if p, ok := any(row).(Preconfigurable); ok {
				p.SetPreconfigured(true)
			}
		}
		row.SetNameableReplacements(mergeReplacements(row.NameableReplacements(), replacements))
		var noParent M
		for _, m := range modifiers {
			// Duplicate gives each copy its own ID but keeps the original's Source, so every copy still syncs with the
			// library modifier the original came from. It has no use for the library file being copied from.
			row.AddModifiers(m.Clone(LibraryFile{}, row.DataOwner(), noParent, Duplicate))
		}
		landed = true
	}
	return landed
}

// enabledModifiers returns the modifiers with the disabled ones taken out, including those within modifier
// containers. A modifier container left with nothing in it is taken out too, as is one that was empty to begin with,
// since it has no effect either way.
func enabledModifiers[M ModifierNode[M, T], T ModifiableNode[T, M]](modifiers []M) []M {
	var kept []M
	for _, m := range modifiers {
		if m.Container() {
			children := enabledModifiers(m.NodeChildren())
			m.SetChildren(children)
			if len(children) == 0 {
				continue
			}
		} else if !m.Enabled() {
			continue
		}
		kept = append(kept, m)
	}
	return kept
}
