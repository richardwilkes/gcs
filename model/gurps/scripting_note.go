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
	"slices"
	"strings"

	"github.com/dop251/goja"
)

func deferredNewScriptNote(note *Note) ScriptSelfProvider {
	if note == nil {
		return ScriptSelfProvider{}
	}
	return ScriptSelfProvider{
		ID:       string(note.TID),
		Provider: func(r *goja.Runtime) any { return newScriptNote(r, note) },
	}
}

func newScriptNote(r *goja.Runtime, note *Note) *goja.Object {
	m := make(map[string]func() goja.Value)
	m["id"] = func() goja.Value { return r.ToValue(note.TID) }
	m["parentID"] = func() goja.Value {
		if note.parent == nil {
			return goja.Undefined()
		}
		return r.ToValue(note.parent.TID)
	}
	m["parent"] = func() goja.Value {
		if note.parent == nil {
			return goja.Undefined()
		}
		return newScriptNote(r, note.parent)
	}
	m["description"] = func() goja.Value { return r.ToValue(note.TextWithReplacements()) }
	m["tags"] = func() goja.Value { return r.ToValue(slices.Clone(note.Tags)) }
	m["container"] = func() goja.Value { return r.ToValue(note.Container()) }
	if note.Container() {
		m["children"] = func() goja.Value {
			children := make([]*goja.Object, 0, len(note.Children))
			for _, child := range note.Children {
				children = append(children, newScriptNote(r, child))
			}
			return r.ToValue(children)
		}
		m["find"] = func() goja.Value {
			return r.ToValue(func(call goja.FunctionCall) goja.Value {
				name := callArgAsString(call, 0)
				tag := callArgAsString(call, 1)
				return findScriptNotes(r, name, tag, note.Children...)
			})
		}
	}
	return r.NewDynamicObject(NewScriptObject(r, m))
}

func findScriptNotes(r *goja.Runtime, name, tag string, topLevelNotes ...*Note) goja.Value {
	var notes []*goja.Object
	Traverse(func(note *Note) bool {
		if (name == "" || strings.EqualFold(note.TextWithReplacements(), name)) && matchTag(tag, note.Tags) {
			notes = append(notes, newScriptNote(r, note))
		}
		return false
	}, true, false, topLevelNotes...)
	return r.ToValue(notes)
}
