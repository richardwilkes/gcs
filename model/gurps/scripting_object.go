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
	"slices"
	"strings"

	"github.com/dop251/goja"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/toolbox/v2/tid"
)

var _ goja.DynamicObject = &ScriptObject{}

// ScriptObject is a generic dynamic object for scripting.
type ScriptObject struct {
	cache map[string]goja.Value
	keys  []string
	kvMap map[string]func() goja.Value
}

// NewScriptObject creates a new ScriptObject for the given data and key/value map.
func NewScriptObject(r *goja.Runtime, kvMap map[string]func() goja.Value) *ScriptObject {
	addScriptObjectToString(r, kvMap)
	return &ScriptObject{
		cache: make(map[string]goja.Value),
		keys:  slices.Sorted(maps.Keys(kvMap)),
		kvMap: kvMap,
	}
}

// addScriptObjectToString gives an object that doesn't define one a toString that reports whatever its valueOf reports,
// or "[object Object]" if it has no valueOf either. The inherited Object.prototype.toString will not do for the first
// case: turning an object into a string consults toString before valueOf, and the inherited one always answers
// "[object Object]", so an object that describes itself as a value — an attribute, whose valueOf yields its maximum —
// would convert to "[object Object]" rather than to that value. Since a script's result reaches the rest of GCS by
// being converted to a string, that is the difference between `<script>$st</script>` producing a number and producing
// "[object Object]".
func addScriptObjectToString(r *goja.Runtime, kvMap map[string]func() goja.Value) {
	if _, exists := kvMap["toString"]; exists {
		return
	}
	valueOf, hasValueOf := kvMap["valueOf"]
	kvMap["toString"] = func() goja.Value {
		return r.ToValue(func(_ goja.FunctionCall) goja.Value {
			if hasValueOf {
				if f, ok := goja.AssertFunction(valueOf()); ok {
					v, err := f(goja.Undefined())
					if err != nil {
						// A script exception or a timeout interrupt raised by valueOf belongs to the script that
						// triggered this conversion, so let it keep unwinding instead of reporting a value for it.
						panic(err)
					}
					return r.ToValue(v.String())
				}
			}
			return r.ToValue("[object Object]")
		})
	}
}

// Get implements goja.DynamicObject. A key this object does not provide must yield nil rather than goja.Undefined():
// goja only consults the prototype chain when Get returns nil, so answering Undefined would shadow Object.prototype in
// its entirety, leaving hasOwnProperty, toString, valueOf, constructor and the rest of it undefined — and a TypeError
// when called. Note that a key the object does provide may still legitimately evaluate to Undefined; only keys that are
// absent from kvMap are reported as missing here.
func (s *ScriptObject) Get(key string) goja.Value {
	if v, ok := s.cache[key]; ok {
		return v
	}
	if f, ok := s.kvMap[key]; ok {
		v := f()
		s.cache[key] = v
		return v
	}
	return nil
}

// Set implements goja.DynamicObject.
func (s *ScriptObject) Set(_ string, _ goja.Value) bool {
	return false // We're read-only
}

// Has implements goja.DynamicObject.
func (s *ScriptObject) Has(key string) bool {
	_, ok := s.kvMap[key]
	return ok
}

// Delete implements goja.DynamicObject.
func (s *ScriptObject) Delete(_ string) bool {
	return false // We're read-only
}

// Keys implements goja.DynamicObject.
func (s *ScriptObject) Keys() []string {
	return s.keys
}

// scriptNode is the part of a node's behavior that the identity properties shared by every node script wrapper rely
// on. It is a constraint of its own rather than Node because Node's method set has neither ID nor Container, and the
// wrappers need nothing else from it.
type scriptNode[T any] interface {
	comparable
	ID() tid.TID
	Parent() T
	Container() bool
}

// addScriptNodeIdentity installs the properties every node script wrapper shares: id, parentID, parent, container and
// tags. parentID and parent are undefined for a top-level node. parent is built on demand with ctor, which is the
// wrapper's own constructor, so an ancestor chain is only materialized as far as a script actually walks it.
func addScriptNodeIdentity[T scriptNode[T]](r *goja.Runtime, m map[string]func() goja.Value, node T, tags []string,
	ctor func(*goja.Runtime, T) *goja.Object,
) {
	var zero T
	m["id"] = func() goja.Value { return r.ToValue(string(node.ID())) }
	m["parentID"] = func() goja.Value {
		if parent := node.Parent(); parent != zero {
			return r.ToValue(string(parent.ID()))
		}
		return goja.Undefined()
	}
	m["parent"] = func() goja.Value {
		if parent := node.Parent(); parent != zero {
			return ctor(r, parent)
		}
		return goja.Undefined()
	}
	m["container"] = func() goja.Value { return r.ToValue(node.Container()) }
	m["tags"] = func() goja.Value { return r.ToValue(slices.Clone(tags)) }
}

// secondaryTextProvider is implemented by the nodes whose script wrappers expose their secondary text as notes.
type secondaryTextProvider interface {
	SecondaryText(optionChecker func(display.Option) bool) string
}

// scriptNotes returns the property a script sees as a node's notes: its secondary text with every display option
// enabled, so nothing the sheet might be configured to hide is withheld from the script.
func scriptNotes(r *goja.Runtime, node secondaryTextProvider) func() goja.Value {
	return func() goja.Value {
		return r.ToValue(node.SecondaryText(func(_ display.Option) bool { return true }))
	}
}

// scriptObjects wraps every item that keep accepts -- every item, when keep is nil -- with ctor and returns the
// wrappers as an array for a script. The predicate is deliberately supplied by the caller rather than defaulting to
// Enabled(), so that what a script is and is not shown stays visible at the call site.
func scriptObjects[T any](r *goja.Runtime, items []T, keep func(T) bool,
	ctor func(*goja.Runtime, T) *goja.Object,
) goja.Value {
	objects := make([]*goja.Object, 0, len(items))
	for _, item := range items {
		if keep == nil || keep(item) {
			objects = append(objects, ctor(r, item))
		}
	}
	return r.ToValue(objects)
}

// traversedScriptObjects is scriptObjects over a node tree: it wraps every enabled node that Traverse reaches under
// top and that keep accepts, skipping containers when excludeContainers is set.
func traversedScriptObjects[T Node[T]](r *goja.Runtime, keep func(T) bool, excludeContainers bool,
	ctor func(*goja.Runtime, T) *goja.Object, top ...T,
) goja.Value {
	var objects []*goja.Object
	Traverse(func(node T) bool {
		if keep == nil || keep(node) {
			objects = append(objects, ctor(r, node))
		}
		return false
	}, true, excludeContainers, top...)
	return r.ToValue(objects)
}

// scriptNodeKind is what findScriptNodes needs to know about a node type: the script wrapper's constructor and how to
// read the name and tags it matches against. Tags is a field rather than a method on every node type, hence the
// accessor.
type scriptNodeKind[T Node[T]] struct {
	ctor   func(*goja.Runtime, T) *goja.Object
	nameOf func(T) string
	tagsOf func(T) []string
}

// findScriptNodes collects the wrappers of every enabled node under top, containers included, whose name matches name
// and whose tags include tag -- an empty name or tag matches anything -- and which keep accepts, when keep is given.
func findScriptNodes[T Node[T]](r *goja.Runtime, name, tag string, kind scriptNodeKind[T], keep func(T) bool,
	top ...T,
) goja.Value {
	return traversedScriptObjects(r, func(node T) bool {
		return (keep == nil || keep(node)) && (name == "" || strings.EqualFold(kind.nameOf(node), name)) &&
			matchTag(tag, kind.tagsOf(node))
	}, false, kind.ctor, top...)
}

// matchTag reports whether tags includes tag, ignoring case. An empty tag matches anything.
func matchTag(tag string, tags []string) bool {
	if tag == "" {
		return true
	}
	for _, t := range tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}

// scriptNameTagFinder returns the property for a script function taking a name and a tag, either of which may be
// omitted, that answers with whatever find locates for them.
func scriptNameTagFinder(r *goja.Runtime, find func(name, tag string) goja.Value) func() goja.Value {
	return func() goja.Value {
		return r.ToValue(func(call goja.FunctionCall) goja.Value {
			return find(callArgAsString(call, 0), callArgAsString(call, 1))
		})
	}
}

// addScriptActiveModifiers installs the findActiveModifier and activeModifiers properties on the wrapper of a node
// that carries modifiers: find looks an active modifier up by name and all yields the node's modifiers, of which only
// the active, non-container ones are listed. The name a script passes is trimmed, as it is for the entity's hasTrait,
// traitLevel and skillLevel.
func addScriptActiveModifiers[M Node[M]](r *goja.Runtime, m map[string]func() goja.Value, find func(string) M,
	all func() []M, ctor func(*goja.Runtime, M) *goja.Object,
) {
	var zero M
	m["findActiveModifier"] = func() goja.Value {
		return r.ToValue(func(call goja.FunctionCall) goja.Value {
			if mod := find(callArgAsTrimmedString(call, 0)); mod != zero {
				return ctor(r, mod)
			}
			return goja.Null()
		})
	}
	m["activeModifiers"] = func() goja.Value { return traversedScriptObjects(r, nil, true, ctor, all()...) }
}
