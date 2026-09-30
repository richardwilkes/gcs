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
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

// initModifierContainerEditor fills in the editor for a modifier group or choice, which has only what describes the
// container and, for a choice, what it asks for; a modifier's other fields mean nothing to a container.
func initModifierContainerEditor(content *unison.Panel, data *gurps.NodeSyncData, container *gurps.ModifierContainerSyncData, source *gurps.SourcedID) {
	addNameLabelAndField(content, &data.Name)
	addLabelAndMultiLineStringField(content, i18n.Text("Notes"), "", &data.LocalNotes)
	if container.IsChoice() {
		addModifierChoiceField(content, container)
	}
	addTagsLabelAndField(content, &data.Tags)
	addPageRefLabelAndField(content, &data.PageRef)
	addPageRefHighlightLabelAndField(content, &data.PageRefHighlight)
	addSourceFields(content, source)
}

// addModifierChoiceField adds the popup that says whether a modifier choice is mandatory or optional. Making it a group
// is left to "Convert to Group".
func addModifierChoiceField(parent *unison.Panel, container *gurps.ModifierContainerSyncData) {
	mandatory := i18n.Text("Mandatory: exactly one must be picked")
	optional := i18n.Text("Optional: at most one may be picked")
	current := optional
	if container.IsMandatoryChoice() {
		current = mandatory
	}
	label := addLabel(parent, i18n.Text("Choice"),
		i18n.Text("A mandatory choice must be made before the modifiers can be used, and is often what sets the price. An optional one makes its modifiers mutually exclusive."))
	parent.AddChild(labelControl(newPopupMenu([]string{mandatory, optional}, current, func(item string) {
		container.SetMandatoryChoice(item == mandatory)
		MarkModified(parent)
	}), label))
}
