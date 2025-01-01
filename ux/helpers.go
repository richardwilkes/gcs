// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/check"
)

const dataOwnerProviderKey = "data_owner_provider"

// SetFieldValue sets the value of this field, marking the field and all of its parents as needing to be laid out again
// if the value is not what is currently in the field.
func SetFieldValue(field *unison.Field, value string) {
	if value != field.Text() {
		field.SetText(value)
		MarkForLayoutWithinDockable(field)
	}
}

// MarkForLayoutWithinDockable sets the NeedsLayout flag on the provided panel and all of its parents up to the first
// Dockable.
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

// FocusFirstContent attempts to focus the first non-button widget in the content. Failing that, tries to focus the
// first focusable widget in the content. Failing that, tries to focus the first focusable widget in the toolbar.
func FocusFirstContent(toolbar, content unison.Paneler) {
	c := content.AsPanel()
	c.RequestFocus()
	wnd := c.Window()
	for {
		actual := wnd.Focus()
		parent := actual
		for parent != nil && !parent.Is(content) {
			parent = parent.Parent()
		}
		if parent == nil {
			// Nothing found within the content that isn't a button, so go back to the first one.
			c.RequestFocus()
			parent = wnd.Focus()
			for parent != nil && !parent.Is(content) {
				parent = parent.Parent()
			}
			if parent == nil {
				// Nothing found within the content that is focusable, so try the toolbar.
				toolbar.AsPanel().RequestFocus()
			}
			break
		}
		if _, ok := actual.Self.(*unison.Button); !ok {
			break
		}
		wnd.FocusNext()
	}
}

// SetDataOwnerProvider sets the DataOwnerProvider into the client data of the target.
func SetDataOwnerProvider(target unison.Paneler, provider gurps.DataOwnerProvider) {
	data := target.AsPanel().ClientData()
	if toolbox.IsNil(provider) {
		delete(data, dataOwnerProviderKey)
	} else {
		data[dataOwnerProviderKey] = provider
	}
}

// DetermineDataOwnerProvider returns the DataOwnerProvider for the given target.
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
