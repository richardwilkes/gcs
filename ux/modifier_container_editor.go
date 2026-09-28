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

// initModifierContainerEditor fills in the content of the editor for a modifier group or choice. Neither is a modifier
// in its own right: a group only organizes the modifiers it holds, and a choice also asks for one of them to be picked.
// So the editor has only what describes the container and, for a choice, what it asks for. The rest of a modifier's
// fields (cost, level, enabled state, features and so on) mean nothing to a container, and aren't kept for one.
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

// addModifierChoiceField adds the popup that says what a modifier choice asks for: exactly one of its options, which
// makes it mandatory, or at most one. The choice can't be taken out of use here, since that would leave a group
// behind; that is the job of the "Convert to Group" command.
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
