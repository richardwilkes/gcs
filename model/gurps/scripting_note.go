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
	"github.com/dop251/goja"
)

func deferredNewScriptNote(note *Note) ScriptSelfProvider {
	return deferredScriptSelf(note, newScriptNote)
}

func newScriptNote(r *goja.Runtime, note *Note) *goja.Object {
	m := make(map[string]func() goja.Value)
	addScriptNodeIdentity(r, m, note, note.Tags, newScriptNote)
	m["description"] = func() goja.Value { return r.ToValue(note.TextWithReplacements()) }
	if note.Container() {
		m["children"] = func() goja.Value { return scriptObjects(r, note.Children, nil, newScriptNote) }
		m["find"] = scriptNameTagFinder(r, func(name, tag string) goja.Value {
			return findScriptNotes(r, name, tag, note.Children...)
		})
	}
	return r.NewDynamicObject(NewScriptObject(r, m))
}

func findScriptNotes(r *goja.Runtime, name, tag string, topLevelNotes ...*Note) goja.Value {
	return findScriptNodes(r, name, tag, scriptNodeKind[*Note]{
		ctor:   newScriptNote,
		nameOf: (*Note).TextWithReplacements,
		tagsOf: func(note *Note) []string { return note.Tags },
	}, nil, topLevelNotes...)
}
