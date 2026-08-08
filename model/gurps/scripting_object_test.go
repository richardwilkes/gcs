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
	"testing"

	"github.com/dop251/goja"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestScriptObjectPrototypeIsReachable verifies that the members Object.prototype supplies are reachable from the
// objects handed to scripts. goja consults a dynamic object's prototype only when its Get returns nil, so a Get that
// answered goja.Undefined() for keys the object does not provide shadowed Object.prototype in its entirety:
// hasOwnProperty, valueOf and constructor all read as undefined, and calling any of them threw a TypeError.
func TestScriptObjectPrototypeIsReachable(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	trait := NewTrait(e, nil, false)
	trait.Name = "Sample"
	e.Traits = []*Trait{trait}
	e.Recalculate()
	selfArg := ScriptArg{Name: "self", Value: func(r *goja.Runtime) any { return newScriptTrait(r, trait) }}
	for _, tc := range []struct {
		name   string
		script string
		want   string
	}{
		{name: "hasOwnProperty exists", script: "typeof self.hasOwnProperty", want: "function"},
		{name: "hasOwnProperty for a provided key", script: "String(self.hasOwnProperty('name'))", want: "true"},
		{name: "hasOwnProperty for a missing key", script: "String(self.hasOwnProperty('notAKey'))", want: "false"},
		{name: "constructor", script: "String(self.constructor === Object)", want: "true"},
		{name: "isPrototypeOf", script: "String(Object.prototype.isPrototypeOf(self))", want: "true"},
		{name: "prototype toString", script: "Object.prototype.toString.call(self)", want: "[object Object]"},
		{name: "in for a provided key", script: "String('name' in self)", want: "true"},
		{name: "in for a missing key", script: "String('notAKey' in self)", want: "false"},
		// Reading a key the object does not provide must still come out as undefined rather than picking up something
		// unexpected from the prototype.
		{name: "missing key", script: "typeof self.notAKey", want: "undefined"},
		{name: "provided key", script: "self.name", want: "Sample"},
	} {
		v, err := runScript(0, tc.script, selfArg)
		c.NoError(err, "case %q", tc.name)
		c.Equal(tc.want, v, "case %q", tc.name)
	}
}

// TestScriptObjectWithValueOfConvertsToItsValue verifies that an object which describes itself as a value through
// valueOf — the attributes bound as $st, $dx and the rest — both answers toString and still turns into that value when
// converted to a string, which is how a script's result reaches the rest of GCS. The two pull against each other:
// JavaScript consults toString before valueOf, so an object with no toString of its own would inherit
// Object.prototype.toString and convert to "[object Object]", while an object with no toString at all leaves
// `$st.toString()` undefined and a TypeError when called.
func TestScriptObjectWithValueOfConvertsToItsValue(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	e.Recalculate()
	attr := e.Attributes.Find("st")
	c.NotNil(attr)
	attrArg := ScriptArg{Name: "$st", Value: func(r *goja.Runtime) any { return newScriptAttribute(r, attr) }}

	// A new entity's ST is 10, so this is what every conversion below has to produce.
	value, err := runScript(0, "String($st.valueOf())", attrArg)
	c.NoError(err)
	c.Equal("10", value)

	for _, tc := range []struct {
		name   string
		script string
		want   string
	}{
		{name: "toString exists", script: "typeof $st.toString", want: "function"},
		{name: "toString reports the value", script: "$st.toString()", want: value},
		{name: "explicit string conversion", script: "String($st)", want: value},
		{name: "concatenation", script: "'' + $st", want: value},
		// The bare attribute is the shape the sheet actually resolves; runScript delivers its result by converting it
		// to a string, so this is what ResolveToNumber has to be able to parse.
		{name: "result conversion", script: "$st", want: value},
		{name: "hasOwnProperty for a provided key", script: "String($st.hasOwnProperty('current'))", want: "true"},
		{name: "hasOwnProperty for a missing key", script: "String($st.hasOwnProperty('notAKey'))", want: "false"},
	} {
		var v string
		v, err = runScript(0, tc.script, attrArg)
		c.NoError(err, "case %q", tc.name)
		c.Equal(tc.want, v, "case %q", tc.name)
	}
}
