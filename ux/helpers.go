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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/mod"
)

const dataOwnerProviderKey = "data_owner_provider"

// noModifiersDown returns true if none of the non-sticky modifier keys (Shift, Control, Option, Command) were down when
// the event occurred. The sticky lock keys (CapsLock, NumLock) are ignored, since a latched one must not prevent a
// plain keypress from being recognized.
func noModifiersDown(mods mod.Modifiers) bool {
	return mods&mod.NonSticky == 0
}

// SetFieldValue sets the text of the field, marking it and its parents up to the enclosing Dockable as needing layout,
// if the value differs from what the field currently holds.
func SetFieldValue(field *unison.Field, value string) {
	if value != field.Text() {
		field.SetText(value)
		MarkForLayoutWithinDockable(field)
	}
}

// MarkForLayoutWithinDockable sets the NeedsLayout flag on the provided panel and its parents, up to and including the
// first Dockable.
func MarkForLayoutWithinDockable(panel unison.Paneler) {
	p := panel.AsPanel()
	for p != nil {
		p.NeedsLayout = true
		if _, ok := p.Self.(unison.Dockable); ok {
			break
		}
		p = p.Parent()
	}
}

// SetCheckBoxState sets the checkbox state based on the value of checked.
func SetCheckBoxState(checkbox *CheckBox, checked bool) {
	checkbox.State = check.FromBool(checked)
	checkbox.Sync()
}

// FocusFirstContent focuses the first non-button widget in the content, failing that the first focusable widget in the
// content, and failing that the first focusable widget in the toolbar.
//
// Only the content and toolbar subtrees are scanned. Cycling the window's focus (via FocusNext) instead would roam the
// entire window and depend on wherever the focus happened to be beforehand, which could intermittently leave the focus
// outside the just-opened dockable -- most visibly in the navigator.
func FocusFirstContent(toolbar, content unison.Paneler) {
	if target := firstContentFocusTarget(toolbar.AsPanel(), content.AsPanel()); target != nil {
		target.RequestFocus()
	}
}

// firstContentFocusTarget returns the panel that FocusFirstContent should focus, or nil if neither the content nor the
// toolbar contains a focusable widget.
func firstContentFocusTarget(toolbar, content *unison.Panel) *unison.Panel {
	if target := firstFocusableInSubtree(content, true); target != nil {
		return target
	}
	if target := firstFocusableInSubtree(content, false); target != nil {
		return target
	}
	return firstFocusableInSubtree(toolbar, false)
}

// firstFocusableInSubtree returns the first focusable panel found in a pre-order traversal of the subtree rooted at p,
// matching the order the window uses when cycling focus, or nil if there is none. When skipButtons is true, buttons are
// not treated as valid targets.
func firstFocusableInSubtree(p *unison.Panel, skipButtons bool) *unison.Panel {
	if p == nil {
		return nil
	}
	if p.Focusable() {
		if _, isButton := p.Self.(*unison.Button); !skipButtons || !isButton {
			return p
		}
	}
	for _, child := range p.Children() {
		if target := firstFocusableInSubtree(child, skipButtons); target != nil {
			return target
		}
	}
	return nil
}

// SetDataOwnerProvider sets the DataOwnerProvider into the client data of the target, removing it when provider is nil.
func SetDataOwnerProvider(target unison.Paneler, provider gurps.DataOwnerProvider) {
	data := target.AsPanel().ClientData()
	if xreflect.IsNil(provider) {
		delete(data, dataOwnerProviderKey)
	} else {
		data[dataOwnerProviderKey] = provider
	}
}

// DetermineDataOwnerProvider returns the DataOwnerProvider for the given target, preferring the nearest one in its
// ancestry and falling back to the one in its client data, or nil if there is none.
func DetermineDataOwnerProvider(target unison.Paneler) gurps.DataOwnerProvider {
	if provider := unison.AncestorOrSelf[gurps.DataOwnerProvider](target); provider != nil {
		return provider
	}
	if data, ok := target.AsPanel().ClientData()[dataOwnerProviderKey]; ok {
		if provider, ok2 := data.(gurps.DataOwnerProvider); ok2 {
			return provider
		}
	}
	return nil
}
