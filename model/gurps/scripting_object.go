// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	if _, ok := kvMap["toString"]; !ok {
		if _, ok = kvMap["valueOf"]; !ok {
			kvMap["toString"] = func() goja.Value {
				return r.ToValue(func(_ goja.FunctionCall) goja.Value {
					return r.ToValue("[object Object]")
				})
			}
		}
	}
	return &ScriptObject{
		cache: make(map[string]goja.Value),
		keys:  slices.Sorted(maps.Keys(kvMap)),
		kvMap: kvMap,
	}
}

// Get implements goja.DynamicObject.
func (s *ScriptObject) Get(key string) goja.Value {
	if v, ok := s.cache[key]; ok {
		return v
	}
	if f, ok := s.kvMap[key]; ok {
		v := f()
		s.cache[key] = v
		return v
	}
	return goja.Undefined()
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
