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

	"github.com/dop251/goja"
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
