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
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
)

// filterFieldInfo is the part of a gurps.FilterField that the filter editor needs: what the field is called and
// whether that names something plural, the key it is stored under, the kind of criteria it takes, and a function
// returning the values to suggest for it, which is nil when it offers none. Reducing the fields to this lets the popup,
// the dialog and the editor be written once rather than once per kind of node.
type filterFieldInfo struct {
	key         string
	title       string
	kind        gurps.FilterFieldKind
	plural      bool
	suggestions func() []string
}

// filterFieldInfos reduces a list type's filter fields to what the editor needs, in the order they were given. data
// returns the nodes of the list being filtered, which the suggestions are drawn from each time they are asked for; nil
// leaves the fields that draw them from the list with nothing to suggest.
func filterFieldInfos[T gurps.Node[T]](fields []*gurps.FilterField[T], data func() []T) []filterFieldInfo {
	infos := make([]filterFieldInfo, len(fields))
	for i, field := range fields {
		infos[i] = filterFieldInfo{key: field.Key, title: field.Title, kind: field.Kind, plural: field.Plural}
		if field.Suggestions != nil {
			infos[i].suggestions = func() []string {
				var list []T
				if data != nil {
					list = data()
				}
				return field.Suggestions(list)
			}
		}
	}
	return infos
}

// saveGlobalSettings writes the global settings out, reporting a failure the way the rest of the UI does. The
// workspace may not have been set up yet -- a test may never start one -- in which case the failure is only logged.
func saveGlobalSettings() {
	if err := gurps.GlobalSettings().Save(); err != nil {
		reportUIError(i18n.Text("Unable to save global settings"), err)
	}
}

// reportUIError reports a failure through the workspace's error handler when there is one, and otherwise logs it.
func reportUIError(msg string, err error) {
	if Workspace.ErrorHandler != nil {
		Workspace.ErrorHandler(msg, err)
		return
	}
	errs.Log(errs.NewWithCause(msg, err))
}
